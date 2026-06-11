<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth, isAuthenticated } from '$lib/stores/auth';
	import { fetchMe, type MeResponse } from '$lib/api/auth';
	import { fetchMyWishlist, type WishlistItem } from '$lib/api/wishlist';
	import { fetchMyOrders, ORDER_STATUS_LABELS, type Order } from '$lib/api/market';
	import { eur } from '$lib/utils/format';
	import GPanel from '$lib/components/galerie/GPanel.svelte';
	import GAvatar from '$lib/components/galerie/GAvatar.svelte';
	import GChip from '$lib/components/galerie/GChip.svelte';
	import Kicker from '$lib/components/galerie/Kicker.svelte';

	let me = $state<MeResponse | null>(null);
	let loading = $state(true);
	let wishlist = $state<WishlistItem[]>([]);
	let orders = $state<Order[]>([]);

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
		try {
			me = await fetchMe($auth.token);
			[wishlist, orders] = await Promise.all([
				fetchMyWishlist($auth.token),
				fetchMyOrders($auth.token)
			]);
		} catch {
			auth.logout();
			goto('/login');
		} finally {
			loading = false;
		}
	});

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
						Aucune pièce en wishlist — parcourez la <a href="/">vitrine</a> pour en ajouter.
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
</style>
