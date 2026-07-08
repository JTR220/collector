<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth, isAuthenticated } from '$lib/stores/auth';
	import { fetchMe, type MeResponse } from '$lib/api/auth';
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
	let myArticles = $state<ArticleAPI[]>([]);
	let articleBusyId = $state<number | null>(null);
	let articlesMsg = $state<string | null>(null);
	let reviewFormOrderId = $state<number | null>(null);
	let reviewRating = $state(5);
	let reviewComment = $state('');
	let reviewBusy = $state(false);
	let reviewMsg = $state<string | null>(null);

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
		if (!$isAuthenticated || !$auth.token) {
			goto('/login');
			return;
		}
		// Seul un echec d'authentification (/me) deconnecte l'utilisateur.
		try {
			me = await fetchMe($auth.token);
			// On rafraichit le role dans le store pour garder l'onglet Admin fiable.
			auth.login($auth.token, {
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
		const [w, o, s, a] = await Promise.allSettled([
			fetchMyWishlist($auth.token),
			fetchMyOrders($auth.token),
			fetchMySales($auth.token),
			fetchMyArticles($auth.token)
		]);
		if (w.status === 'fulfilled') wishlist = w.value;
		if (o.status === 'fulfilled') orders = o.value;
		if (s.status === 'fulfilled') sales = s.value;
		if (a.status === 'fulfilled') myArticles = a.value;
		loading = false;
	});

	async function removeArticle(article: ArticleAPI) {
		if (!$auth.token) return;
		if (!confirm(`Retirer « ${article.name} » du catalogue ?`)) return;
		articleBusyId = article.ID;
		articlesMsg = null;
		try {
			await deleteArticle($auth.token, article.ID);
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
		if (!$auth.token) return;
		reviewBusy = true;
		reviewMsg = null;
		try {
			await leaveReview($auth.token, order.ID, reviewRating, reviewComment.trim());
			orders = orders.map((o) => (o.ID === order.ID ? { ...o, reviewed: true } : o));
			reviewFormOrderId = null;
		} catch (e) {
			reviewMsg = e instanceof Error ? e.message : "Impossible d'enregistrer l'avis.";
		} finally {
			reviewBusy = false;
		}
	}

	async function decide(order: Order, accept: boolean) {
		if (!$auth.token) return;
		salesBusyId = order.ID;
		salesMsg = null;
		try {
			const { order: updated } = accept
				? await acceptOrder($auth.token, order.ID)
				: await rejectOrder($auth.token, order.ID);
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
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 12px;
		color: #a39a8c;
		letter-spacing: 0.18em;
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
		font-family: 'Newsreader', Georgia, serif;
		font-weight: 500;
		font-size: 46px;
		line-height: 1.05;
		margin: 8px 0 0;
		letter-spacing: -0.01em;
		color: #ece5da;
	}
	.identity-meta {
		display: flex;
		gap: 10px;
		align-items: center;
		margin-top: 8px;
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 12.5px;
		color: #a39a8c;
	}
	.meta-sep {
		color: #766d60;
	}
	.identity-btns {
		display: flex;
		flex-direction: column;
		gap: 8px;
		flex-shrink: 0;
	}
	.btn-ghost {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 12.5px;
		padding: 9px 22px;
		border-radius: 7px;
		border: 1px solid rgba(236, 229, 218, 0.1);
		cursor: pointer;
		background: transparent;
		color: #a39a8c;
		transition:
			border-color 120ms,
			color 120ms;
	}
	.btn-ghost:hover {
		border-color: rgba(236, 229, 218, 0.22);
		color: #ece5da;
	}

	/* Ventes à valider */
	.sale-row {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 11px 0;
		border-bottom: 1px solid rgba(236, 229, 218, 0.1);
	}
	.sale-actions {
		display: flex;
		gap: 8px;
		flex-shrink: 0;
	}
	.btn-accept,
	.btn-reject {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 12px;
		font-weight: 600;
		padding: 7px 14px;
		border-radius: 6px;
		cursor: pointer;
		border: none;
		transition: filter 120ms;
	}
	.btn-accept {
		background: #86b3a4;
		color: #191714;
	}
	.btn-reject {
		background: transparent;
		border: 1px solid rgba(215, 156, 134, 0.5);
		color: #d79c86;
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
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 12.5px;
		color: #86b3a4;
		margin-top: 10px;
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
		background: rgba(255, 255, 255, 0.03);
		border: 1px solid rgba(236, 229, 218, 0.08);
	}
	.stat-value {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 22px;
		color: #ece5da;
	}
	.stat-label {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 11.5px;
		color: #a39a8c;
	}

	/* Mes annonces */
	.listing-row {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 11px 0;
		border-bottom: 1px solid rgba(236, 229, 218, 0.1);
	}
	.item-views {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 11.5px;
		color: #766d60;
		flex-shrink: 0;
	}
	.listing-actions {
		display: flex;
		gap: 8px;
		flex-shrink: 0;
	}
	.btn-ghost-sm,
	.btn-reject-sm {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
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
		border: 1px solid rgba(236, 229, 218, 0.16);
		color: #a39a8c;
	}
	.btn-reject-sm {
		background: transparent;
		border: 1px solid rgba(215, 156, 134, 0.5);
		color: #d79c86;
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
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 11.5px;
		color: #766d60;
		font-style: italic;
		flex-shrink: 0;
	}
	.review-form {
		display: flex;
		flex-direction: column;
		gap: 8px;
		padding: 12px 0 16px;
		border-bottom: 1px solid rgba(236, 229, 218, 0.1);
	}
	.review-stars-input {
		display: flex;
		gap: 4px;
	}
	.star-btn {
		background: none;
		border: none;
		font-size: 18px;
		color: rgba(236, 229, 218, 0.25);
		cursor: pointer;
		padding: 0;
		line-height: 1;
	}
	.star-btn.star-on {
		color: #e0b260;
	}
	.review-form textarea {
		resize: vertical;
		background: rgba(255, 255, 255, 0.03);
		border: 1px solid rgba(236, 229, 218, 0.14);
		border-radius: 8px;
		padding: 9px 12px;
		color: #ece5da;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13px;
	}
	.review-form textarea:focus {
		outline: none;
		border-color: #86b3a4;
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
		border-bottom: 1px solid rgba(236, 229, 218, 0.1);
		text-decoration: none;
	}
	.item-date {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 11px;
		color: #766d60;
		flex-shrink: 0;
	}
	.item-name {
		flex: 1;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13px;
		color: #ece5da;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		text-decoration: none;
	}
	.item-row:hover .item-name,
	a.item-name:hover {
		color: #86b3a4;
	}
	.item-price {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 12.5px;
		color: #a39a8c;
		flex-shrink: 0;
	}
	.item-empty {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 12.5px;
		color: #766d60;
		line-height: 1.5;
		padding: 12px 0;
	}
	.item-empty a {
		color: #86b3a4;
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
		border-bottom: 1px solid rgba(236, 229, 218, 0.1);
		text-decoration: none;
	}
	.tl-dot {
		width: 7px;
		height: 7px;
		border-radius: 50%;
		background: #766d60;
		flex-shrink: 0;
	}
	.tl-dot.tl-buy {
		background: #86b3a4;
	}
	.tl-date {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 11px;
		color: #766d60;
		flex-shrink: 0;
		width: 82px;
	}
	.tl-action {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 10.5px;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		color: #a39a8c;
		flex-shrink: 0;
		width: 68px;
	}
	.tl-name {
		flex: 1;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13px;
		color: #ece5da;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.tl-row:hover .tl-name {
		color: #86b3a4;
	}
	.tl-price {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 12.5px;
		color: #a39a8c;
		flex-shrink: 0;
	}
</style>
