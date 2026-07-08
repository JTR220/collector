<script lang="ts">
	import { goto } from '$app/navigation';
	import { auth, isAuthenticated } from '$lib/stores/auth';
	import { cart, cartTotal } from '$lib/stores/cart';
	import { buyArticle } from '$lib/api/market';
	import { articleImage } from '$lib/api/catalog';
	import { eur } from '$lib/utils/format';
	import GPanel from '$lib/components/galerie/GPanel.svelte';
	import Kicker from '$lib/components/galerie/Kicker.svelte';

	let checkingOut = $state(false);
	let results = $state<Record<number, 'ok' | string>>({});

	function remove(id: number) {
		cart.remove(id);
		const next = { ...results };
		delete next[id];
		results = next;
	}

	async function checkout() {
		if (!$isAuthenticated || !$auth.token) {
			goto('/login');
			return;
		}
		const token = $auth.token;
		checkingOut = true;
		results = {};
		// Sequentiel (et non Promise.all) : deux achats concurrents sur le meme
		// article echoueraient de toute facon en base (compare-and-swap sold),
		// autant garder des messages d'erreur lisibles un par un.
		for (const item of $cart) {
			try {
				await buyArticle(token, item.ID);
				results = { ...results, [item.ID]: 'ok' };
				cart.remove(item.ID);
			} catch (e) {
				results = { ...results, [item.ID]: e instanceof Error ? e.message : 'Achat impossible.' };
			}
		}
		checkingOut = false;
	}
</script>

<svelte:head><title>Panier · Collector.shop</title></svelte:head>

<section class="head">
	<Kicker>Panier</Kicker>
	<h1 class="title">Votre panier</h1>
	<p class="sub">
		Regroupez plusieurs pièces avant de valider — chaque achat reste une commande distincte, soumise
		à la validation de son vendeur.
	</p>
</section>

{#if $cart.length === 0}
	<GPanel>
		<p class="empty">Votre panier est vide. <a href="/">Parcourir le marché</a>.</p>
	</GPanel>
{:else}
	<GPanel>
		<div class="cart-list">
			{#each $cart as item (item.ID)}
				{@const img = articleImage(item)}
				<div class="cart-row">
					<div class="cart-thumb">
						{#if img}<img src={img} alt={item.name} />{/if}
					</div>
					<div class="cart-info">
						<a class="cart-name" href={`/lot/${item.ID}`}>{item.name}</a>
						<span class="cart-seller">@{item.seller}</span>
					</div>
					<span class="cart-price">{eur(item.prix + item.fraisPort)}</span>
					{#if results[item.ID] === 'ok'}
						<span class="cart-result cart-result-ok">✓ Achetée</span>
					{:else if results[item.ID]}
						<span class="cart-result cart-result-err">{results[item.ID]}</span>
					{:else}
						<button class="cart-remove" onclick={() => remove(item.ID)}>Retirer</button>
					{/if}
				</div>
			{/each}
		</div>

		<div class="cart-total-row">
			<span class="cart-total-label">Total ({$cart.length} pièce{$cart.length > 1 ? 's' : ''})</span
			>
			<span class="cart-total-val">{eur($cartTotal)}</span>
		</div>

		<button class="btn-checkout" disabled={checkingOut} onclick={checkout}>
			{checkingOut ? 'Validation…' : 'Valider le panier'}
		</button>
	</GPanel>
{/if}

<style>
	.head {
		padding: 20px 0 18px;
	}
	.title {
		font-family: 'Newsreader', Georgia, serif;
		font-weight: 500;
		font-size: clamp(28px, 4vw, 40px);
		color: #ece5da;
		margin: 8px 0 10px;
	}
	.sub {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 14px;
		color: #a39a8c;
		line-height: 1.55;
		max-width: 560px;
		margin: 0;
	}
	.empty {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13px;
		color: #766d60;
		padding: 12px 0;
	}
	.empty a {
		color: #86b3a4;
	}

	.cart-list {
		display: flex;
		flex-direction: column;
	}
	.cart-row {
		display: flex;
		align-items: center;
		gap: 14px;
		padding: 12px 0;
		border-bottom: 1px solid rgba(236, 229, 218, 0.1);
	}
	.cart-thumb {
		width: 48px;
		height: 48px;
		border-radius: 6px;
		background: rgba(255, 255, 255, 0.04);
		overflow: hidden;
		flex-shrink: 0;
	}
	.cart-thumb img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	.cart-info {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 2px;
	}
	.cart-name {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13.5px;
		color: #ece5da;
		text-decoration: none;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.cart-name:hover {
		color: #86b3a4;
	}
	.cart-seller {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 11px;
		color: #766d60;
	}
	.cart-price {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 13px;
		color: #a39a8c;
		flex-shrink: 0;
	}
	.cart-remove {
		flex-shrink: 0;
		background: none;
		border: 1px solid rgba(236, 229, 218, 0.14);
		border-radius: 6px;
		padding: 6px 12px;
		color: #a39a8c;
		font-size: 11.5px;
		cursor: pointer;
	}
	.cart-remove:hover {
		border-color: rgba(215, 156, 134, 0.4);
		color: #d79c86;
	}
	.cart-result {
		flex-shrink: 0;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 11.5px;
		max-width: 200px;
		text-align: right;
	}
	.cart-result-ok {
		color: #86c099;
	}
	.cart-result-err {
		color: #d79c86;
	}

	.cart-total-row {
		display: flex;
		justify-content: space-between;
		align-items: baseline;
		padding-top: 16px;
		margin-top: 6px;
	}
	.cart-total-label {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13px;
		color: #a39a8c;
	}
	.cart-total-val {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 22px;
		color: #ece5da;
	}
	.btn-checkout {
		width: 100%;
		margin-top: 14px;
		padding: 13px;
		border-radius: 8px;
		border: none;
		background: #86b3a4;
		color: #191714;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13.5px;
		font-weight: 700;
		cursor: pointer;
		transition: filter 150ms;
	}
	.btn-checkout:hover:not(:disabled) {
		filter: brightness(1.08);
	}
	.btn-checkout:disabled {
		opacity: 0.55;
		cursor: not-allowed;
	}
</style>
