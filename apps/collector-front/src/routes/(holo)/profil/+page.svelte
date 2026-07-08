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
		ORDER_STATUS_LABELS,
		type Order,
		type OrderStatus
	} from '$lib/api/market';
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

	const pendingSales = $derived(sales.filter((s) => s.status === 'pending'));

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
		const [w, o, s] = await Promise.allSettled([
			fetchMyWishlist($auth.token),
			fetchMyOrders($auth.token),
			fetchMySales($auth.token)
		]);
		if (w.status === 'fulfilled') wishlist = w.value;
		if (o.status === 'fulfilled') orders = o.value;
		if (s.status === 'fulfilled') sales = s.value;
		loading = false;
	});

	async function decide(order: Order, accept: boolean) {
		if (!$auth.token) return;
		salesBusyId = order.ID;
		salesMsg = null;
		try {
			const { order: updated } = accept
				? await acceptOrder($auth.token, order.ID)
				: await rejectOrder($auth.token, order.ID);
			sales = sales.map((s) => (s.ID === updated.ID ? { ...s, status: updated.status } : s));
			salesMsg = accept ? 'Commande acceptée.' : 'Commande refusée — la pièce redevient disponible.';
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
						<span class="item-name">{s.article?.name ?? `Lot #${s.articleId}`}</span>
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
					<a class="item-row" href={`/lot/${o.articleId}`}>
						<span class="item-date">{fmtDate(o.CreatedAt)}</span>
						<span class="item-name">{o.article?.name ?? `Lot #${o.articleId}`}</span>
						<GChip>{ORDER_STATUS_LABELS[o.status] ?? o.status}</GChip>
						<span class="item-price">{eur(o.price)}</span>
					</a>
				{:else}
					<p class="item-empty">Aucun achat pour l'instant.</p>
				{/each}
			</div>
		</GPanel>
	</div>

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
	}
	.item-row:hover .item-name {
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
