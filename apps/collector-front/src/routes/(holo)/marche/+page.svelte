<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth, isAuthenticated } from '$lib/stores/auth';
	import { articleImage, type ArticleAPI } from '$lib/api/catalog';
	import {
		fetchMyListings,
		fetchMyOrders,
		fetchMySales,
		updateOrderStatus,
		deleteListing,
		ORDER_STATUS_LABELS,
		type Order,
		type OrderStatus
	} from '$lib/api/market';
	import { eur } from '$lib/utils/format';
	import GPanel from '$lib/components/galerie/GPanel.svelte';
	import GChip from '$lib/components/galerie/GChip.svelte';
	import Kicker from '$lib/components/galerie/Kicker.svelte';

	type Tab = 'annonces' | 'ventes' | 'achats';
	let tab = $state<Tab>('annonces');

	let listings = $state<ArticleAPI[]>([]);
	let sales = $state<Order[]>([]);
	let orders = $state<Order[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let busyId = $state<number | null>(null);

	onMount(async () => {
		if (!$isAuthenticated || !$auth.token) {
			goto('/login');
			return;
		}
		try {
			[listings, sales, orders] = await Promise.all([
				fetchMyListings($auth.token),
				fetchMySales($auth.token),
				fetchMyOrders($auth.token)
			]);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Impossible de charger votre marché.';
			console.error(e);
		} finally {
			loading = false;
		}
	});

	const statusColor: Record<OrderStatus, string> = {
		paid: '#86b3a4',
		shipped: '#a39a8c',
		delivered: '#86c099',
		cancelled: '#d79c86'
	};

	async function setStatus(order: Order, status: OrderStatus, list: 'ventes' | 'achats') {
		if (!$auth.token) return;
		busyId = order.ID;
		try {
			const updated = await updateOrderStatus($auth.token, order.ID, status);
			if (list === 'ventes') sales = sales.map((o) => (o.ID === updated.ID ? updated : o));
			else orders = orders.map((o) => (o.ID === updated.ID ? updated : o));
		} catch (e) {
			error = e instanceof Error ? e.message : 'Erreur lors de la mise à jour.';
		} finally {
			busyId = null;
		}
	}

	async function removeListing(articleId: number) {
		if (!$auth.token) return;
		busyId = articleId;
		try {
			await deleteListing($auth.token, articleId);
			listings = listings.filter((a) => a.ID !== articleId);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Erreur lors du retrait.';
		} finally {
			busyId = null;
		}
	}

	const tabs: { id: Tab; label: string }[] = [
		{ id: 'annonces', label: 'Mes annonces' },
		{ id: 'ventes', label: 'Mes ventes' },
		{ id: 'achats', label: 'Mes achats' }
	];

	const counts = $derived<Record<Tab, number>>({
		annonces: listings.length,
		ventes: sales.length,
		achats: orders.length
	});
</script>

<svelte:head><title>Mon marché · Collector.shop</title></svelte:head>

<div class="mk-head">
	<Kicker color="#86b3a4">Marketplace</Kicker>
	<div class="mk-title-row">
		<h1 class="mk-title">Mon marché</h1>
		<a class="btn-primary" href="/vendre">+ Vendre une pièce</a>
	</div>
</div>

<div class="mk-tabs">
	{#each tabs as t}
		<button class="mk-tab" class:active={tab === t.id} onclick={() => (tab = t.id)}>
			{t.label} <span class="mk-count">{counts[t.id]}</span>
		</button>
	{/each}
</div>

{#if error}
	<p class="mk-error">{error}</p>
{/if}

{#if loading}
	<p class="state-msg">Chargement de votre marché…</p>
{:else if tab === 'annonces'}
	{#if listings.length === 0}
		<GPanel>
			<p class="empty">Aucune annonce pour le moment. <a href="/vendre">Mettez votre première pièce en vente →</a></p>
		</GPanel>
	{:else}
		<div class="mk-list">
			{#each listings as article (article.ID)}
				{@const img = articleImage(article)}
				<GPanel>
					<div class="row">
						<a class="thumb" href={`/lot/${article.ID}`}>
							{#if img}<img src={img} alt={article.name} />{:else}<span>{article.glyph || '◈'}</span>{/if}
						</a>
						<div class="row-main">
							<a class="row-name" href={`/lot/${article.ID}`}>{article.name}</a>
							<span class="row-sub">{article.slug} · {article.category?.name ?? '—'}</span>
						</div>
						<span class="row-price">{eur(article.prix)}</span>
						<GChip color={article.sold ? '#d79c86' : '#86b3a4'}>
							{article.sold ? 'Vendu' : 'En vente'}
						</GChip>
						{#if !article.sold}
							<button class="btn-danger" disabled={busyId === article.ID} onclick={() => removeListing(article.ID)}>
								Retirer
							</button>
						{/if}
					</div>
				</GPanel>
			{/each}
		</div>
	{/if}
{:else}
	{@const list = tab === 'ventes' ? sales : orders}
	{#if list.length === 0}
		<GPanel>
			<p class="empty">
				{tab === 'ventes'
					? 'Aucune vente pour le moment — vos pièces attendent leur acheteur.'
					: 'Aucun achat pour le moment.'}
				{#if tab === 'achats'}<a href="/">Parcourir la vitrine →</a>{/if}
			</p>
		</GPanel>
	{:else}
		<div class="mk-list">
			{#each list as order (order.ID)}
				{@const img = articleImage(order.article)}
				<GPanel>
					<div class="row">
						<a class="thumb" href={`/lot/${order.articleId}`}>
							{#if img}<img src={img} alt={order.article?.name} />{:else}<span>{order.article?.glyph || '◈'}</span>{/if}
						</a>
						<div class="row-main">
							<a class="row-name" href={`/lot/${order.articleId}`}>{order.article?.name ?? `Lot #${order.articleId}`}</a>
							<span class="row-sub">
								Commande #{order.ID} ·
								{tab === 'ventes' ? `acheteur #${order.buyerId}` : `vendeur @${order.article?.seller ?? order.sellerId}`}
							</span>
						</div>
						<span class="row-price">{eur(order.price + order.fraisPort)}</span>
						<GChip color={statusColor[order.status]}>{ORDER_STATUS_LABELS[order.status]}</GChip>

						{#if tab === 'ventes' && order.status === 'paid'}
							<button class="btn-act" disabled={busyId === order.ID} onclick={() => setStatus(order, 'shipped', 'ventes')}>
								Marquer expédiée
							</button>
						{:else if tab === 'achats' && order.status === 'shipped'}
							<button class="btn-act" disabled={busyId === order.ID} onclick={() => setStatus(order, 'delivered', 'achats')}>
								Confirmer réception
							</button>
						{:else if tab === 'achats' && order.status === 'paid'}
							<button class="btn-danger" disabled={busyId === order.ID} onclick={() => setStatus(order, 'cancelled', 'achats')}>
								Annuler
							</button>
						{/if}
					</div>
				</GPanel>
			{/each}
		</div>
	{/if}
{/if}

<style>
	.mk-head { margin-bottom: 18px; }
	.mk-title-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 16px;
		flex-wrap: wrap;
	}
	.mk-title {
		font-family: 'Newsreader', Georgia, serif;
		font-weight: 500;
		font-size: 36px;
		color: #ece5da;
		margin: 8px 0 0;
	}
	.btn-primary {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13px;
		font-weight: 600;
		padding: 11px 20px;
		border-radius: 7px;
		background: #86b3a4;
		color: #191714;
		text-decoration: none;
		transition: filter 120ms;
	}
	.btn-primary:hover { filter: brightness(1.08); }

	.mk-tabs {
		display: flex;
		gap: 8px;
		margin-bottom: 16px;
		border-bottom: 1px solid rgba(236, 229, 218, 0.1);
		padding-bottom: 0;
	}
	.mk-tab {
		background: none;
		border: none;
		border-bottom: 2px solid transparent;
		padding: 8px 12px;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13px;
		color: #a39a8c;
		cursor: pointer;
		transition: color 120ms, border-color 120ms;
	}
	.mk-tab:hover { color: #ece5da; }
	.mk-tab.active { color: #ece5da; font-weight: 600; border-bottom-color: #86b3a4; }
	.mk-count {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 11px;
		color: #766d60;
		margin-left: 4px;
	}

	.mk-error {
		padding: 10px 14px;
		border-radius: 7px;
		border: 1px solid rgba(215, 156, 134, 0.3);
		background: rgba(215, 156, 134, 0.06);
		color: #d79c86;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13px;
		margin-bottom: 14px;
	}
	.state-msg {
		text-align: center;
		padding: 60px 0;
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 12px;
		color: #766d60;
		letter-spacing: 0.12em;
	}
	.empty {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13.5px;
		color: #a39a8c;
		margin: 4px 0;
	}
	.empty a { color: #86b3a4; }

	.mk-list { display: flex; flex-direction: column; gap: 10px; }
	.row { display: flex; align-items: center; gap: 14px; flex-wrap: wrap; }
	.thumb {
		width: 56px;
		height: 56px;
		border-radius: 7px;
		border: 1px solid rgba(236, 229, 218, 0.1);
		background: radial-gradient(120% 90% at 30% 20%, #4a6a5a 0%, #2a3a32 55%, #191714 100%);
		display: flex;
		align-items: center;
		justify-content: center;
		overflow: hidden;
		flex-shrink: 0;
		text-decoration: none;
	}
	.thumb img { width: 100%; height: 100%; object-fit: cover; }
	.thumb span {
		font-family: 'Newsreader', Georgia, serif;
		font-size: 22px;
		color: rgba(236, 229, 218, 0.85);
	}
	.row-main { flex: 1; min-width: 160px; display: flex; flex-direction: column; gap: 2px; }
	.row-name {
		font-family: 'Newsreader', Georgia, serif;
		font-size: 17px;
		color: #ece5da;
		text-decoration: none;
	}
	.row-name:hover { color: #86b3a4; }
	.row-sub {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 11px;
		color: #766d60;
	}
	.row-price {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 15px;
		color: #ece5da;
	}

	.btn-act {
		padding: 8px 14px;
		border-radius: 6px;
		border: 1px solid rgba(134, 179, 164, 0.4);
		background: transparent;
		color: #86b3a4;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 12px;
		cursor: pointer;
		transition: background 120ms;
	}
	.btn-act:hover:not(:disabled) { background: rgba(134, 179, 164, 0.08); }
	.btn-danger {
		padding: 8px 14px;
		border-radius: 6px;
		border: 1px solid rgba(215, 156, 134, 0.4);
		background: transparent;
		color: #d79c86;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 12px;
		cursor: pointer;
		transition: background 120ms;
	}
	.btn-danger:hover:not(:disabled) { background: rgba(215, 156, 134, 0.08); }
	.btn-act:disabled,
	.btn-danger:disabled { opacity: 0.5; cursor: not-allowed; }
</style>
