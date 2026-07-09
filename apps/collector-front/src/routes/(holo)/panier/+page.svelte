<script lang="ts">
	import { goto } from '$app/navigation';
	import { cart, cartTotal } from '$lib/stores/cart';
	import { articleImage, type ArticleAPI } from '$lib/api/catalog';
	import { eur } from '$lib/utils/format';

	function remove(id: number) {
		cart.remove(id);
	}

	const groups = $derived.by(() => {
		const bySeller = new Map<string, ArticleAPI[]>();
		for (const item of $cart) {
			const list = bySeller.get(item.seller) ?? [];
			list.push(item);
			bySeller.set(item.seller, list);
		}
		return [...bySeller.entries()];
	});

	const shipping = $derived($cart.reduce((sum, i) => sum + i.fraisPort, 0));
	const subtotal = $derived($cart.reduce((sum, i) => sum + i.prix, 0));

	function goToPaiement() {
		goto('/paiement');
	}
</script>

<svelte:head><title>Panier · Collector.shop</title></svelte:head>

<section class="steps">
	<span class="step step-active">1. Panier</span>
	<span class="step">2. Livraison</span>
	<span class="step">3. Paiement</span>
</section>

{#if $cart.length === 0}
	<div class="empty">
		<p>Votre panier est vide. <a href="/">Parcourir le marché</a>.</p>
	</div>
{:else}
	<div class="panier-grid">
		<div class="cart-col">
			<h1 class="cart-title">Votre panier ({$cart.length} article{$cart.length > 1 ? 's' : ''})</h1>

			{#each groups as [seller, items] (seller)}
				<div class="seller-label">Vendu par {seller}</div>
				{#each items as item (item.ID)}
					{@const img = articleImage(item)}
					<div class="cart-row">
						<div class="cart-thumb">
							{#if img}<img src={img} alt={item.name} />{/if}
						</div>
						<div class="cart-info">
							<a class="cart-name" href={`/lot/${item.ID}`}>{item.name}</a>
							{#if item.grade}<span class="cart-condition">{item.grade}</span>{/if}
						</div>
						<span class="cart-price">{eur(item.prix)}</span>
						<button class="cart-remove" onclick={() => remove(item.ID)}>Retirer</button>
					</div>
				{/each}
			{/each}
		</div>

		<div class="summary-card">
			<h2 class="summary-title">Récapitulatif</h2>
			<div class="summary-row">
				<span>Sous-total</span>
				<span>{eur(subtotal)}</span>
			</div>
			<div class="summary-row">
				<span>Frais de port ({groups.length} vendeur{groups.length > 1 ? 's' : ''})</span>
				<span>{eur(shipping)}</span>
			</div>
			<div class="summary-divider"></div>
			<div class="summary-total">
				<span>Total</span>
				<span>{eur($cartTotal)}</span>
			</div>
			<button class="btn-checkout" onclick={goToPaiement}>Passer à la commande</button>
			<p class="summary-trust">
				Paiement 100% sécurisé · Aucun échange de coordonnées entre membres
			</p>
		</div>
	</div>
{/if}

<style>
	.steps {
		display: flex;
		gap: 28px;
		padding: 24px 0 32px;
		font-family: var(--f-body);
		font-size: 13px;
		color: var(--c-text-muted);
	}
	.step-active {
		color: var(--c-ink);
		font-weight: 600;
		border-bottom: 2px solid var(--c-ink);
		padding-bottom: 4px;
	}

	.empty {
		font-family: var(--f-body);
		font-size: 14px;
		color: var(--c-text-muted);
		padding: 40px 0;
	}
	.empty a {
		color: var(--c-ink);
		font-weight: 600;
	}

	.panier-grid {
		display: grid;
		grid-template-columns: 1.5fr 0.9fr;
		gap: 48px;
		align-items: start;
		padding-bottom: 40px;
	}
	@media (max-width: 900px) {
		.panier-grid {
			grid-template-columns: 1fr;
			gap: 24px;
		}
	}

	.cart-title {
		font-family: var(--f-serif);
		font-weight: 600;
		font-size: 24px;
		color: var(--c-text);
		margin: 0 0 16px;
	}
	.seller-label {
		font-family: var(--f-body);
		font-size: 13px;
		font-weight: 600;
		color: var(--c-text-muted);
		margin: 20px 0 10px;
	}
	.seller-label:first-of-type {
		margin-top: 0;
	}
	.cart-row {
		display: flex;
		gap: 16px;
		align-items: center;
		padding: 16px 0;
		border-bottom: 1px solid var(--c-border);
	}
	.cart-thumb {
		width: 88px;
		height: 88px;
		flex-shrink: 0;
		border-radius: 10px;
		background: var(--c-bg);
		overflow: hidden;
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
		gap: 4px;
	}
	.cart-name {
		font-family: var(--f-body);
		font-size: 14px;
		font-weight: 600;
		color: var(--c-text);
		text-decoration: none;
	}
	.cart-name:hover {
		color: var(--c-ink);
	}
	.cart-condition {
		font-family: var(--f-body);
		font-size: 12px;
		color: var(--c-text-muted);
	}
	.cart-price {
		font-family: var(--f-serif);
		font-size: 17px;
		font-weight: 600;
		color: var(--c-ink);
		width: 90px;
		text-align: right;
		flex-shrink: 0;
	}
	.cart-remove {
		flex-shrink: 0;
		background: none;
		border: 1px solid var(--c-border);
		border-radius: 6px;
		padding: 6px 12px;
		color: var(--c-text-muted);
		font-family: var(--f-body);
		font-size: 11.5px;
		cursor: pointer;
	}
	.cart-remove:hover {
		border-color: var(--c-error);
		color: var(--c-error);
	}

	.summary-card {
		background: var(--c-surface);
		border: 1px solid var(--c-border);
		border-radius: 16px;
		padding: 28px;
		display: flex;
		flex-direction: column;
		gap: 16px;
	}
	.summary-title {
		font-family: var(--f-serif);
		font-size: 19px;
		font-weight: 600;
		color: var(--c-text);
		margin: 0;
	}
	.summary-row {
		display: flex;
		justify-content: space-between;
		font-family: var(--f-body);
		font-size: 14px;
		color: var(--c-text-tertiary);
	}
	.summary-divider {
		height: 1px;
		background: var(--c-border);
	}
	.summary-total {
		display: flex;
		justify-content: space-between;
		font-family: var(--f-serif);
		font-size: 20px;
		font-weight: 700;
		color: var(--c-ink);
	}
	.btn-checkout {
		padding: 15px;
		border: none;
		border-radius: 10px;
		background: var(--c-accent);
		color: #fff;
		font-family: var(--f-body);
		font-size: 15px;
		font-weight: 600;
		cursor: pointer;
		transition: filter 150ms;
	}
	.btn-checkout:hover {
		filter: brightness(1.08);
	}
	.summary-trust {
		font-family: var(--f-body);
		font-size: 12px;
		color: var(--c-text-muted);
		text-align: center;
		margin: 0;
	}
</style>
