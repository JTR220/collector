<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth, isAuthenticated } from '$lib/stores/auth';
	import {
		fetchMe,
		updateProfile,
		exportMyData,
		deleteMyAccount,
		type MeResponse
	} from '$lib/api/auth';
	import { fetchMyWishlist, type WishlistItem } from '$lib/api/wishlist';
	import {
		fetchMyOrders,
		fetchMySales,
		acceptOrder,
		rejectOrder,
		leaveReview,
		ORDER_STATUS_LABELS,
		type Order,
		type OrderStatus
	} from '$lib/api/market';
	import { fetchMyArticles, deleteArticle, type ArticleAPI } from '$lib/api/catalog';
	import {
		fetchReceivedOffers,
		fetchSentOffers,
		acceptOffer,
		rejectOffer,
		payOffer,
		OFFER_STATUS_LABELS,
		type Offer
	} from '$lib/api/offers';
	import { eur } from '$lib/utils/format';
	import GPanel from '$lib/components/galerie/GPanel.svelte';
	import GAvatar from '$lib/components/galerie/GAvatar.svelte';
	import GChip from '$lib/components/galerie/GChip.svelte';
	import Kicker from '$lib/components/galerie/Kicker.svelte';

	let me = $state<MeResponse | null>(null);
	let loading = $state(true);
	let wishlist = $state<WishlistItem[]>([]);
	let orders = $state<Order[]>([]);
	let sales = $state<Order[]>([]);
	let salesBusyId = $state<number | null>(null);
	let salesMsg = $state<string | null>(null);
	let receivedOffers = $state<Offer[]>([]);
	let sentOffers = $state<Offer[]>([]);
	let offersBusyId = $state<number | null>(null);
	let offersMsg = $state<string | null>(null);
	let confirmOffer = $state<Offer | null>(null);
	let myArticles = $state<ArticleAPI[]>([]);
	let articleBusyId = $state<number | null>(null);
	let articlesMsg = $state<string | null>(null);
	let reviewFormOrderId = $state<number | null>(null);
	let reviewRating = $state(5);
	let reviewComment = $state('');
	let reviewBusy = $state(false);
	let reviewMsg = $state<string | null>(null);

	// ── RGPD : modification / export / suppression du compte ──────────────────
	let editingProfile = $state(false);
	let editName = $state('');
	let editEmail = $state('');
	let editPassword = $state('');
	let profileBusy = $state(false);
	let profileMsg = $state<string | null>(null);
	let exportBusy = $state(false);
	let deleteBusy = $state(false);

	const pendingSales = $derived(sales.filter((s) => s.status === 'pending'));
	const reviewableStatuses: OrderStatus[] = ['paid', 'shipped', 'delivered'];
	function canReview(o: Order) {
		return !o.reviewed && reviewableStatuses.includes(o.status);
	}

	const articleStats = $derived({
		listed: myArticles.filter((a) => !a.sold).length,
		sold: myArticles.filter((a) => a.sold).length,
		totalViews: myArticles.reduce((sum, a) => sum + a.views, 0)
	});

	const initials = $derived(
		me?.name
			? me.name
					.split(' ')
					.map((w) => w[0])
					.join('')
					.toUpperCase()
					.slice(0, 2)
			: '??'
	);
	const handle = $derived(me?.name ? me.name.toLowerCase().replace(/\s+/g, '_') : '…');

	onMount(async () => {
		if (!$isAuthenticated || !$auth.user) {
			goto('/login');
			return;
		}
		// Seul un echec d'authentification (/me) deconnecte l'utilisateur.
		try {
			me = await fetchMe();
			// On rafraichit le role dans le store pour garder l'onglet Admin fiable.
			auth.login({
				id: me.id,
				name: me.name,
				email: me.email,
				role: me.role
			});
		} catch {
			auth.logout();
			goto('/login');
			return;
		}
		// Wishlist et commandes viennent du catalog-service : une panne de ce
		// service ne doit PAS deconnecter — on charge ce qui repond.
		const [w, o, s, a, ro, so] = await Promise.allSettled([
			fetchMyWishlist(),
			fetchMyOrders(),
			fetchMySales(),
			fetchMyArticles(),
			fetchReceivedOffers(),
			fetchSentOffers()
		]);
		if (w.status === 'fulfilled') wishlist = w.value;
		if (o.status === 'fulfilled') orders = o.value;
		if (s.status === 'fulfilled') sales = s.value;
		if (a.status === 'fulfilled') myArticles = a.value;
		if (ro.status === 'fulfilled') receivedOffers = ro.value;
		if (so.status === 'fulfilled') sentOffers = so.value;
		loading = false;
	});

	async function removeArticle(article: ArticleAPI) {
		if (!$auth.user) return;
		if (!confirm(`Retirer « ${article.name} » du catalogue ?`)) return;
		articleBusyId = article.ID;
		articlesMsg = null;
		try {
			await deleteArticle(article.ID);
			myArticles = myArticles.filter((a) => a.ID !== article.ID);
		} catch (e) {
			articlesMsg = e instanceof Error ? e.message : "Impossible de retirer l'annonce.";
		} finally {
			articleBusyId = null;
		}
	}

	function toggleReview(order: Order) {
		reviewMsg = null;
		if (reviewFormOrderId === order.ID) {
			reviewFormOrderId = null;
			return;
		}
		reviewFormOrderId = order.ID;
		reviewRating = 5;
		reviewComment = '';
	}

	async function submitReview(order: Order) {
		if (!$auth.user) return;
		reviewBusy = true;
		reviewMsg = null;
		try {
			await leaveReview(order.ID, reviewRating, reviewComment.trim());
			orders = orders.map((o) => (o.ID === order.ID ? { ...o, reviewed: true } : o));
			reviewFormOrderId = null;
		} catch (e) {
			reviewMsg = e instanceof Error ? e.message : "Impossible d'enregistrer l'avis.";
		} finally {
			reviewBusy = false;
		}
	}

	async function decide(order: Order, accept: boolean) {
		if (!$auth.user) return;
		salesBusyId = order.ID;
		salesMsg = null;
		try {
			const { order: updated } = accept ? await acceptOrder(order.ID) : await rejectOrder(order.ID);
			sales = sales.map((s) => (s.ID === updated.ID ? { ...s, status: updated.status } : s));
			salesMsg = accept
				? 'Commande acceptée.'
				: 'Commande refusée — la pièce redevient disponible.';
		} catch (e) {
			salesMsg = e instanceof Error ? e.message : 'Erreur lors du traitement de la commande.';
		} finally {
			salesBusyId = null;
		}
	}

	function openAcceptConfirm(offer: Offer) {
		offersMsg = null;
		confirmOffer = offer;
	}

	function closeAcceptConfirm() {
		confirmOffer = null;
	}

	async function confirmAcceptOffer() {
		if (!confirmOffer) return;
		const offer = confirmOffer;
		offersBusyId = offer.ID;
		offersMsg = null;
		try {
			await acceptOffer(offer.ID);
			receivedOffers = receivedOffers.filter((o) => o.ID !== offer.ID);
			offersMsg = 'Offre acceptée — l’acheteur peut désormais payer au prix négocié.';
		} catch (e) {
			offersMsg = e instanceof Error ? e.message : "Impossible d'accepter l'offre.";
		} finally {
			offersBusyId = null;
			confirmOffer = null;
		}
	}

	async function declineOffer(offer: Offer) {
		offersBusyId = offer.ID;
		offersMsg = null;
		try {
			await rejectOffer(offer.ID);
			receivedOffers = receivedOffers.filter((o) => o.ID !== offer.ID);
			offersMsg = 'Offre refusée.';
		} catch (e) {
			offersMsg = e instanceof Error ? e.message : "Impossible de refuser l'offre.";
		} finally {
			offersBusyId = null;
		}
	}

	async function payAcceptedOffer(offer: Offer) {
		offersBusyId = offer.ID;
		offersMsg = null;
		try {
			const { order } = await payOffer(offer.ID);
			sentOffers = sentOffers.map((o) => (o.ID === offer.ID ? { ...o, status: 'purchased' } : o));
			orders = [order, ...orders];
			offersMsg = 'Paiement effectué — retrouvez la commande dans « Mes achats ».';
		} catch (e) {
			offersMsg = e instanceof Error ? e.message : 'Impossible de payer cette offre.';
		} finally {
			offersBusyId = null;
		}
	}

	// Historique : flux chronologique fusionnant achats et ajouts wishlist.
	type Activity = {
		date: string;
		kind: 'buy' | 'wish';
		name: string;
		articleId: number;
		status?: OrderStatus;
		price?: number;
	};
	const timeline = $derived<Activity[]>(
		[
			...orders.map((o): Activity => ({
				date: o.CreatedAt,
				kind: 'buy',
				name: o.article?.name ?? `Lot #${o.articleId}`,
				articleId: o.articleId,
				status: o.status,
				price: o.price
			})),
			...wishlist.map((w): Activity => ({
				date: w.CreatedAt,
				kind: 'wish',
				name: w.article?.name ?? `Lot #${w.articleId}`,
				articleId: w.articleId
			}))
		].sort((a, b) => new Date(b.date).getTime() - new Date(a.date).getTime())
	);

	function logout() {
		auth.logout();
		goto('/login');
	}

	function toggleEditProfile() {
		profileMsg = null;
		editingProfile = !editingProfile;
		if (editingProfile && me) {
			editName = me.name;
			editEmail = me.email;
			editPassword = '';
		}
	}

	async function saveProfile() {
		if (!me) return;
		profileBusy = true;
		profileMsg = null;
		try {
			const updated = await updateProfile({
				name: editName.trim(),
				email: editEmail.trim(),
				...(editPassword ? { password: editPassword } : {})
			});
			me = { ...me, name: updated.name, email: updated.email };
			auth.login({ id: me.id, name: me.name, email: me.email, role: me.role });
			editingProfile = false;
			profileMsg = 'Profil mis à jour.';
		} catch (e) {
			profileMsg = e instanceof Error ? e.message : 'Impossible de mettre à jour le profil.';
		} finally {
			profileBusy = false;
		}
	}

	async function downloadMyData() {
		exportBusy = true;
		profileMsg = null;
		try {
			const data = await exportMyData();
			const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
			const url = URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = url;
			a.download = 'mes-donnees-collector.json';
			a.click();
			URL.revokeObjectURL(url);
		} catch (e) {
			profileMsg = e instanceof Error ? e.message : "Impossible d'exporter vos données.";
		} finally {
			exportBusy = false;
		}
	}

	async function deleteAccount() {
		if (
			!confirm(
				'Supprimer définitivement votre compte et toutes vos données personnelles ? Cette action est irréversible.'
			)
		)
			return;
		deleteBusy = true;
		profileMsg = null;
		try {
			await deleteMyAccount();
			auth.logout();
			goto('/login');
		} catch (e) {
			profileMsg = e instanceof Error ? e.message : 'Impossible de supprimer le compte.';
			deleteBusy = false;
		}
	}

	const fmtDate = (iso: string) =>
		new Date(iso).toLocaleDateString('fr-FR', {
			day: '2-digit',
			month: '2-digit',
			year: 'numeric'
		});
</script>

<svelte:head><title>PROFIL · Collector.shop</title></svelte:head>

{#if loading}
	<div class="state-msg">Chargement du profil…</div>
{:else if me}
	<!-- Bandeau identité -->
	<section class="identity">
		<GAvatar {initials} size={96} />
		<div class="identity-text">
			<Kicker>Mon compte</Kicker>
			<h1 class="identity-name">{me.name}</h1>
			<div class="identity-meta">
				<span>@{handle}</span>
				<span class="meta-sep">·</span>
				<span>{me.email}</span>
			</div>
		</div>
		<div class="identity-btns">
			<button class="btn-ghost" onclick={logout}>Se déconnecter</button>
		</div>
	</section>

	{#if pendingSales.length > 0}
		<GPanel style="margin-bottom:14px">
			<Kicker>Ventes à valider · {pendingSales.length}</Kicker>
			<div class="item-list">
				{#each pendingSales as s (s.ID)}
					<div class="sale-row">
						<a class="item-name" href={`/lot/${s.articleId}`}
							>{s.article?.name ?? `Lot #${s.articleId}`}</a
						>
						<span class="item-price">{eur(s.price)}</span>
						<div class="sale-actions">
							<button
								class="btn-accept"
								disabled={salesBusyId === s.ID}
								onclick={() => decide(s, true)}>Accepter</button
							>
							<button
								class="btn-reject"
								disabled={salesBusyId === s.ID}
								onclick={() => decide(s, false)}>Refuser</button
							>
						</div>
					</div>
				{/each}
			</div>
			{#if salesMsg}
				<p class="action-msg">{salesMsg}</p>
			{/if}
		</GPanel>
	{/if}

	{#if receivedOffers.length > 0}
		<GPanel style="margin-bottom:14px">
			<Kicker>Offres reçues · {receivedOffers.length}</Kicker>
			<div class="item-list">
				{#each receivedOffers as o (o.ID)}
					<div class="sale-row">
						<a class="item-name" href={`/lot/${o.articleId}`}
							>{o.article?.name ?? `Lot #${o.articleId}`}</a
						>
						{#if o.article}
							<span class="offer-vs-price">{eur(o.article.prix)}</span>
						{/if}
						<span class="item-price">{eur(o.price)}</span>
						<div class="sale-actions">
							<button
								class="btn-accept"
								disabled={offersBusyId === o.ID}
								onclick={() => openAcceptConfirm(o)}>Accepter</button
							>
							<button
								class="btn-reject"
								disabled={offersBusyId === o.ID}
								onclick={() => declineOffer(o)}>Refuser</button
							>
						</div>
					</div>
					{#if o.message}<p class="offer-message">« {o.message} »</p>{/if}
				{/each}
			</div>
			{#if offersMsg}
				<p class="action-msg">{offersMsg}</p>
			{/if}
		</GPanel>
	{/if}

	{#if confirmOffer}
		<div
			class="modal-overlay"
			role="button"
			tabindex="0"
			onclick={closeAcceptConfirm}
			onkeydown={(e) => e.key === 'Escape' && closeAcceptConfirm()}
		>
			<div
				class="modal-box"
				role="dialog"
				aria-modal="true"
				tabindex="-1"
				onclick={(e) => e.stopPropagation()}
			>
				<h3 class="modal-title">Accepter cette offre ?</h3>
				<p class="modal-text">
					Accepter l'offre à <strong>{eur(confirmOffer.price)}</strong> pour «
					{confirmOffer.article?.name ?? `Lot #${confirmOffer.articleId}`} »
					{#if confirmOffer.article}
						(au lieu de {eur(confirmOffer.article.prix)})
					{/if}
					? L'acheteur pourra payer à ce prix réduit.
				</p>
				<div class="modal-actions">
					<button class="btn-ghost-sm" onclick={closeAcceptConfirm}>Annuler</button>
					<button
						class="btn-accept"
						disabled={offersBusyId === confirmOffer.ID}
						onclick={confirmAcceptOffer}>Confirmer l'acceptation</button
					>
				</div>
			</div>
		</div>
	{/if}

	<!-- Confidentialité & compte -->
	<GPanel style="margin-bottom:14px">
		<div class="privacy-head">
			<Kicker>Confidentialité &amp; compte</Kicker>
			<button class="btn-ghost-sm" onclick={toggleEditProfile}>
				{editingProfile ? 'Annuler' : 'Modifier mon profil'}
			</button>
		</div>

		{#if editingProfile}
			<div class="edit-form">
				<label class="field">
					<span>Nom</span>
					<input type="text" bind:value={editName} disabled={profileBusy} />
				</label>
				<label class="field">
					<span>Email</span>
					<input type="email" bind:value={editEmail} disabled={profileBusy} />
				</label>
				<label class="field">
					<span>Nouveau mot de passe (facultatif)</span>
					<input
						type="password"
						bind:value={editPassword}
						placeholder="Laisser vide pour ne pas changer"
						disabled={profileBusy}
					/>
				</label>
				<button class="btn-accept" disabled={profileBusy} onclick={saveProfile}>
					Enregistrer
				</button>
			</div>
		{/if}

		<div class="privacy-actions">
			<button class="btn-ghost-sm" disabled={exportBusy} onclick={downloadMyData}>
				Exporter mes données (JSON)
			</button>
			<button class="btn-reject-sm" disabled={deleteBusy} onclick={deleteAccount}>
				Supprimer mon compte
			</button>
		</div>
		{#if profileMsg}<p class="action-msg">{profileMsg}</p>{/if}
	</GPanel>

	<!-- Statistiques -->
	<GPanel style="margin-bottom:14px">
		<Kicker>Statistiques</Kicker>
		<div class="stats-row">
			<div class="stat-tile">
				<span class="stat-value">{articleStats.listed}</span>
				<span class="stat-label">En vente</span>
			</div>
			<div class="stat-tile">
				<span class="stat-value">{articleStats.sold}</span>
				<span class="stat-label">Vendues</span>
			</div>
			<div class="stat-tile">
				<span class="stat-value">{articleStats.totalViews}</span>
				<span class="stat-label">Vues cumulées</span>
			</div>
		</div>
	</GPanel>

	<div class="two-col">
		<!-- Wishlist -->
		<GPanel>
			<Kicker>Wishlist · {wishlist.length}</Kicker>
			<div class="item-list">
				{#each wishlist as w (w.ID)}
					<a class="item-row" href={`/lot/${w.articleId}`}>
						<span class="item-name">{w.article?.name ?? `Lot #${w.articleId}`}</span>
						{#if w.article}
							<span class="item-price">{eur(w.article.prix)}</span>
						{/if}
					</a>
				{:else}
					<p class="item-empty">
						Aucune pièce en wishlist. Ajoutez-en depuis la <a href="/">vitrine</a>.
					</p>
				{/each}
			</div>
		</GPanel>

		<!-- Mes achats -->
		<GPanel>
			<Kicker>Mes achats · {orders.length}</Kicker>
			<div class="item-list">
				{#each orders as o (o.ID)}
					<div class="item-row">
						<span class="item-date">{fmtDate(o.CreatedAt)}</span>
						<a class="item-name" href={`/lot/${o.articleId}`}
							>{o.article?.name ?? `Lot #${o.articleId}`}</a
						>
						<GChip>{ORDER_STATUS_LABELS[o.status] ?? o.status}</GChip>
						<span class="item-price">{eur(o.price)}</span>
						{#if o.reviewed}
							<span class="review-tag">avis laissé</span>
						{:else if canReview(o)}
							<button class="btn-ghost-sm" onclick={() => toggleReview(o)}>★ Avis</button>
						{/if}
					</div>
					{#if reviewFormOrderId === o.ID}
						<div class="review-form">
							<div class="review-stars-input">
								{#each [1, 2, 3, 4, 5] as n}
									<button
										type="button"
										class="star-btn"
										class:star-on={n <= reviewRating}
										onclick={() => (reviewRating = n)}
										aria-label={`${n} étoile${n > 1 ? 's' : ''}`}>★</button
									>
								{/each}
							</div>
							<textarea
								placeholder="Commentaire (facultatif)…"
								bind:value={reviewComment}
								disabled={reviewBusy}
								rows="2"></textarea>
							<button class="btn-accept" disabled={reviewBusy} onclick={() => submitReview(o)}>
								Envoyer l'avis
							</button>
						</div>
					{/if}
				{:else}
					<p class="item-empty">Aucun achat pour l'instant.</p>
				{/each}
			</div>
			{#if reviewMsg}<p class="action-msg">{reviewMsg}</p>{/if}
		</GPanel>
	</div>

	{#if sentOffers.length > 0}
		<GPanel style="margin-top:14px">
			<Kicker>Mes offres envoyées · {sentOffers.length}</Kicker>
			<div class="item-list">
				{#each sentOffers as o (o.ID)}
					<div class="item-row">
						<span class="item-date">{fmtDate(o.CreatedAt)}</span>
						<a class="item-name" href={`/lot/${o.articleId}`}
							>{o.article?.name ?? `Lot #${o.articleId}`}</a
						>
						<GChip>{OFFER_STATUS_LABELS[o.status] ?? o.status}</GChip>
						<span class="item-price">{eur(o.price)}</span>
						{#if o.status === 'accepted'}
							<button
								class="btn-accept"
								disabled={offersBusyId === o.ID}
								onclick={() => payAcceptedOffer(o)}>Payer {eur(o.price)}</button
							>
						{/if}
					</div>
				{:else}
					<p class="item-empty">Aucune offre envoyée pour l'instant.</p>
				{/each}
			</div>
			{#if offersMsg}<p class="action-msg">{offersMsg}</p>{/if}
		</GPanel>
	{/if}

	<!-- Mes annonces -->
	<GPanel style="margin-top:14px">
		<Kicker>Mes annonces · {myArticles.length}</Kicker>
		<div class="item-list">
			{#each myArticles as a (a.ID)}
				<div class="listing-row">
					<a class="item-name" href={`/lot/${a.ID}`}>{a.name}</a>
					<span class="item-views" title="Vues">👁 {a.views}</span>
					<GChip>{a.sold ? 'Vendue' : 'En vente'}</GChip>
					<span class="item-price">{eur(a.prix)}</span>
					<div class="listing-actions">
						<a class="btn-ghost-sm" href={`/vendre?edit=${a.ID}`}>Modifier</a>
						<button
							class="btn-reject-sm"
							disabled={articleBusyId === a.ID}
							onclick={() => removeArticle(a)}>Supprimer</button
						>
					</div>
				</div>
			{:else}
				<p class="item-empty">
					Aucune annonce pour l'instant. <a href="/vendre">Mettez une pièce en vente</a>.
				</p>
			{/each}
		</div>
		{#if articlesMsg}
			<p class="action-msg">{articlesMsg}</p>
		{/if}
	</GPanel>

	<!-- Historique d'activité -->
	<div class="history-wrap">
		<GPanel>
			<Kicker>Historique · {timeline.length}</Kicker>
			<div class="timeline">
				{#each timeline as ev (ev.kind + ev.articleId + ev.date)}
					<a class="tl-row" href={`/lot/${ev.articleId}`}>
						<span class="tl-dot" class:tl-buy={ev.kind === 'buy'}></span>
						<span class="tl-date">{fmtDate(ev.date)}</span>
						<span class="tl-action">{ev.kind === 'buy' ? 'Achat' : 'Wishlist'}</span>
						<span class="tl-name">{ev.name}</span>
						{#if ev.kind === 'buy'}
							<GChip>{(ev.status && ORDER_STATUS_LABELS[ev.status]) ?? ev.status}</GChip>
							<span class="tl-price">{eur(ev.price ?? 0)}</span>
						{/if}
					</a>
				{:else}
					<p class="item-empty">
						Aucune activité pour l'instant — vos achats et ajouts wishlist apparaîtront ici.
					</p>
				{/each}
			</div>
		</GPanel>
	</div>
{/if}

<style>
	.state-msg {
		text-align: center;
		padding: 80px 0;
		font-family: var(--f-serif);
		font-style: italic;
		font-size: 15px;
		color: var(--c-text-muted);
	}

	/* Identité */
	.identity {
		display: flex;
		align-items: center;
		gap: 24px;
		padding: 10px 0 22px;
		flex-wrap: wrap;
	}
	.identity-text {
		flex: 1;
		min-width: 200px;
	}
	.identity-name {
		font-family: var(--f-serif);
		font-weight: 600;
		font-size: 40px;
		line-height: 1.05;
		margin: 8px 0 0;
		letter-spacing: -0.01em;
		color: var(--c-text);
	}
	.identity-meta {
		display: flex;
		gap: 10px;
		align-items: center;
		margin-top: 8px;
		font-family: var(--f-body);
		font-size: 12.5px;
		color: var(--c-text-muted);
	}
	.meta-sep {
		color: var(--c-border);
	}
	.identity-btns {
		display: flex;
		flex-direction: column;
		gap: 8px;
		flex-shrink: 0;
	}
	.btn-ghost {
		font-family: var(--f-body);
		font-size: 12.5px;
		padding: 9px 22px;
		border-radius: 7px;
		border: 1px solid var(--c-border);
		cursor: pointer;
		background: transparent;
		color: var(--c-text-tertiary);
		transition:
			border-color 120ms,
			color 120ms;
	}
	.btn-ghost:hover {
		border-color: var(--c-ink);
		color: var(--c-ink);
	}

	/* Ventes à valider */
	.sale-row {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 11px 0;
		border-bottom: 1px solid var(--c-border);
	}
	.sale-actions {
		display: flex;
		gap: 8px;
		flex-shrink: 0;
	}
	.btn-accept,
	.btn-reject {
		font-family: var(--f-body);
		font-size: 12px;
		font-weight: 600;
		padding: 7px 14px;
		border-radius: 6px;
		cursor: pointer;
		border: none;
		transition: filter 120ms;
	}
	.btn-accept {
		background: var(--c-ink);
		color: var(--c-bg);
	}
	.btn-reject {
		background: transparent;
		border: 1px solid var(--c-error);
		color: var(--c-error);
	}
	.btn-accept:hover:not(:disabled),
	.btn-reject:hover:not(:disabled) {
		filter: brightness(1.1);
	}
	.btn-accept:disabled,
	.btn-reject:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}
	.action-msg {
		font-family: var(--f-body);
		font-size: 12.5px;
		color: var(--c-ink);
		font-weight: 600;
		margin-top: 10px;
	}

	/* Confidentialité & compte */
	.privacy-head {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
		flex-wrap: wrap;
	}
	.edit-form {
		display: flex;
		flex-direction: column;
		gap: 12px;
		margin-top: 14px;
		padding-top: 14px;
		border-top: 1px solid var(--c-border);
	}
	.field {
		display: flex;
		flex-direction: column;
		gap: 5px;
	}
	.field span {
		font-family: var(--f-body);
		font-size: 11.5px;
		color: var(--c-text-muted);
	}
	.field input {
		background: var(--c-bg);
		border: 1px solid var(--c-border);
		border-radius: 8px;
		padding: 9px 12px;
		color: var(--c-text);
		font-family: var(--f-body);
		font-size: 13px;
	}
	.field input:focus {
		outline: none;
		border-color: var(--c-ink);
	}
	.edit-form .btn-accept {
		align-self: flex-start;
	}
	.privacy-actions {
		display: flex;
		gap: 10px;
		margin-top: 14px;
		flex-wrap: wrap;
	}

	/* Statistiques */
	.stats-row {
		display: flex;
		gap: 14px;
		margin-top: 10px;
		flex-wrap: wrap;
	}
	.stat-tile {
		flex: 1;
		min-width: 120px;
		display: flex;
		flex-direction: column;
		gap: 4px;
		padding: 12px 16px;
		border-radius: 8px;
		background: var(--c-bg);
		border: 1px solid var(--c-border);
	}
	.stat-value {
		font-family: var(--f-serif);
		font-size: 22px;
		font-weight: 600;
		color: var(--c-ink);
	}
	.stat-label {
		font-family: var(--f-body);
		font-size: 11.5px;
		color: var(--c-text-muted);
	}

	/* Mes annonces */
	.listing-row {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 11px 0;
		border-bottom: 1px solid var(--c-border);
	}
	.item-views {
		font-family: var(--f-body);
		font-size: 11.5px;
		color: var(--c-text-muted);
		flex-shrink: 0;
	}
	.listing-actions {
		display: flex;
		gap: 8px;
		flex-shrink: 0;
	}
	.btn-ghost-sm,
	.btn-reject-sm {
		font-family: var(--f-body);
		font-size: 12px;
		font-weight: 600;
		padding: 6px 12px;
		border-radius: 6px;
		cursor: pointer;
		text-decoration: none;
		transition: filter 120ms;
	}
	.btn-ghost-sm {
		background: transparent;
		border: 1px solid var(--c-border);
		color: var(--c-text-tertiary);
	}
	.btn-reject-sm {
		background: transparent;
		border: 1px solid var(--c-error);
		color: var(--c-error);
	}
	.btn-ghost-sm:hover,
	.btn-reject-sm:hover:not(:disabled) {
		filter: brightness(1.15);
	}
	.btn-reject-sm:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	/* Avis */
	.review-tag {
		font-family: var(--f-body);
		font-size: 11.5px;
		color: var(--c-text-muted);
		font-style: italic;
		flex-shrink: 0;
	}
	.review-form {
		display: flex;
		flex-direction: column;
		gap: 8px;
		padding: 12px 0 16px;
		border-bottom: 1px solid var(--c-border);
	}
	.review-stars-input {
		display: flex;
		gap: 4px;
	}
	.star-btn {
		background: none;
		border: none;
		font-size: 18px;
		color: var(--c-border);
		cursor: pointer;
		padding: 0;
		line-height: 1;
	}
	.star-btn.star-on {
		color: var(--c-accent);
	}

	/* Offres */
	.offer-vs-price {
		font-family: var(--f-body);
		font-size: 12px;
		color: var(--c-text-muted);
		text-decoration: line-through;
		flex-shrink: 0;
	}
	.offer-message {
		font-family: var(--f-body);
		font-size: 12px;
		font-style: italic;
		color: var(--c-text-muted);
		margin: -4px 0 8px;
		padding-bottom: 8px;
		border-bottom: 1px solid var(--c-border);
	}

	.modal-overlay {
		position: fixed;
		inset: 0;
		background: rgba(43, 38, 32, 0.45);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 20000;
		padding: 20px;
	}
	.modal-box {
		background: var(--c-surface);
		border: 1px solid var(--c-border);
		border-radius: 14px;
		max-width: 420px;
		width: 100%;
		padding: 24px;
		box-shadow: 0 20px 48px -12px rgba(43, 38, 32, 0.35);
	}
	.modal-title {
		font-family: var(--f-serif);
		font-size: 19px;
		font-weight: 600;
		color: var(--c-text);
		margin: 0 0 12px;
	}
	.modal-text {
		font-family: var(--f-body);
		font-size: 13.5px;
		line-height: 1.6;
		color: var(--c-text-tertiary);
		margin: 0 0 20px;
	}
	.modal-actions {
		display: flex;
		justify-content: flex-end;
		gap: 10px;
	}
	.review-form textarea {
		resize: vertical;
		background: var(--c-bg);
		border: 1px solid var(--c-border);
		border-radius: 8px;
		padding: 9px 12px;
		color: var(--c-text);
		font-family: var(--f-body);
		font-size: 13px;
	}
	.review-form textarea:focus {
		outline: none;
		border-color: var(--c-ink);
	}
	.review-form .btn-accept {
		align-self: flex-start;
	}

	/* 2 colonnes */
	.two-col {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 14px;
	}
	@media (max-width: 768px) {
		.two-col {
			grid-template-columns: 1fr;
		}
	}

	/* Listes */
	.item-list {
		display: flex;
		flex-direction: column;
		margin-top: 6px;
	}
	.item-row {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 11px 0;
		border-bottom: 1px solid var(--c-border);
		text-decoration: none;
	}
	.item-date {
		font-family: var(--f-body);
		font-size: 11px;
		color: var(--c-text-muted);
		flex-shrink: 0;
	}
	.item-name {
		flex: 1;
		font-family: var(--f-body);
		font-size: 13px;
		color: var(--c-text);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		text-decoration: none;
	}
	.item-row:hover .item-name,
	a.item-name:hover {
		color: var(--c-ink);
	}
	.item-price {
		font-family: var(--f-body);
		font-size: 12.5px;
		font-weight: 600;
		color: var(--c-ink);
		flex-shrink: 0;
	}
	.item-empty {
		font-family: var(--f-body);
		font-size: 12.5px;
		color: var(--c-text-muted);
		line-height: 1.5;
		padding: 12px 0;
	}
	.item-empty a {
		color: var(--c-ink);
		font-weight: 600;
	}

	/* Historique */
	.history-wrap {
		margin-top: 14px;
	}
	.timeline {
		display: flex;
		flex-direction: column;
		margin-top: 6px;
	}
	.tl-row {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 11px 0;
		border-bottom: 1px solid var(--c-border);
		text-decoration: none;
	}
	.tl-dot {
		width: 7px;
		height: 7px;
		border-radius: 50%;
		background: var(--c-border);
		flex-shrink: 0;
	}
	.tl-dot.tl-buy {
		background: var(--c-ink);
	}
	.tl-date {
		font-family: var(--f-body);
		font-size: 11px;
		color: var(--c-text-muted);
		flex-shrink: 0;
		width: 82px;
	}
	.tl-action {
		font-family: var(--f-body);
		font-size: 10.5px;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		color: var(--c-text-muted);
		flex-shrink: 0;
		width: 68px;
	}
	.tl-name {
		flex: 1;
		font-family: var(--f-body);
		font-size: 13px;
		color: var(--c-text);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.tl-row:hover .tl-name {
		color: var(--c-ink);
	}
	.tl-price {
		font-family: var(--f-body);
		font-size: 12.5px;
		font-weight: 600;
		color: var(--c-ink);
		flex-shrink: 0;
	}
</style>
