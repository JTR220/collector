<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth, isAuthenticated } from '$lib/stores/auth';
	import { cart, cartTotal } from '$lib/stores/cart';
	import { buyArticle } from '$lib/api/market';
	import { eur } from '$lib/utils/format';

	let prenom = $state('');
	let nom = $state('');
	let adresse = $state('');
	let codePostal = $state('');
	let ville = $state('');
	let numeroCarte = $state('');
	let expiration = $state('');
	let cvc = $state('');

	let submitting = $state(false);
	let results = $state<Record<number, 'ok' | string>>({});
	let formError = $state<string | null>(null);

	onMount(() => {
		if ($cart.length === 0) goto('/panier');
	});

	const shipping = $derived($cart.reduce((sum, i) => sum + i.fraisPort, 0));

	function formValid(): boolean {
		return !!(
			prenom.trim() &&
			nom.trim() &&
			adresse.trim() &&
			codePostal.trim() &&
			ville.trim() &&
			numeroCarte.trim() &&
			expiration.trim() &&
			cvc.trim()
		);
	}

	async function payer() {
		if (!$isAuthenticated || !$auth.token) {
			goto('/login');
			return;
		}
		if (!formValid()) {
			formError = 'Merci de renseigner tous les champs requis.';
			return;
		}
		formError = null;
		const token = $auth.token;
		submitting = true;
		results = {};
		// Sequentiel : chaque achat est une commande distincte soumise a la
		// validation de son vendeur, on garde des messages d'erreur lisibles un par un.
		for (const item of $cart) {
			try {
				await buyArticle(token, item.ID);
				results = { ...results, [item.ID]: 'ok' };
				cart.remove(item.ID);
			} catch (e) {
				results = { ...results, [item.ID]: e instanceof Error ? e.message : 'Achat impossible.' };
			}
		}
		submitting = false;
		if ($cart.length === 0) goto('/profil');
	}
</script>

<svelte:head><title>Paiement · Collector.shop</title></svelte:head>

<section class="steps">
	<span class="step">1. Panier</span>
	<span class="step">2. Livraison</span>
	<span class="step step-active">3. Paiement</span>
</section>

<div class="paiement-grid">
	<div class="form-col">
		<div class="block">
			<h2 class="block-title">Adresse de livraison</h2>
			<div class="fields-grid">
				<input class="in" placeholder="Prénom" bind:value={prenom} />
				<input class="in" placeholder="Nom" bind:value={nom} />
				<input class="in span-2" placeholder="Adresse" bind:value={adresse} />
				<input class="in" placeholder="Code postal" bind:value={codePostal} />
				<input class="in" placeholder="Ville" bind:value={ville} />
			</div>
		</div>

		<div class="block">
			<div class="paiement-head">
				<h2 class="block-title">Paiement</h2>
				<span class="secure-badge">🔒 Connexion sécurisée HTTPS</span>
			</div>
			<div class="card-box">
				<div class="card-radio">
					<span class="radio-dot"></span>
					<span class="radio-label">Carte bancaire</span>
				</div>
				<input class="in" placeholder="Numéro de carte" bind:value={numeroCarte} />
				<div class="fields-grid two">
					<input class="in" placeholder="MM / AA" bind:value={expiration} />
					<input class="in" placeholder="CVC" bind:value={cvc} />
				</div>
			</div>
		</div>

		{#if formError}<p class="form-error">{formError}</p>{/if}
	</div>

	<div class="summary-card">
		<h2 class="summary-title">Résumé de commande</h2>
		{#each $cart as item (item.ID)}
			<div class="summary-item">
				<span>{item.name}</span>
				<span class="summary-item-price">{eur(item.prix)}</span>
			</div>
			{#if results[item.ID] === 'ok'}
				<span class="result-ok">✓ Achetée</span>
			{:else if results[item.ID]}
				<span class="result-err">{results[item.ID]}</span>
			{/if}
		{/each}
		<div class="summary-divider"></div>
		<div class="summary-row">
			<span>Frais de port</span>
			<span>{eur(shipping)}</span>
		</div>
		<div class="summary-total">
			<span>Total</span>
			<span>{eur($cartTotal)}</span>
		</div>
		<button class="btn-pay" disabled={submitting || $cart.length === 0} onclick={payer}>
			{submitting ? 'Validation…' : `Payer ${eur($cartTotal)}`}
		</button>
		<p class="summary-trust">En validant, vous acceptez les CGV de Collector.shop</p>
	</div>
</div>

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

	.paiement-grid {
		display: grid;
		grid-template-columns: 1.4fr 0.9fr;
		gap: 48px;
		align-items: start;
		padding-bottom: 40px;
	}
	@media (max-width: 900px) {
		.paiement-grid {
			grid-template-columns: 1fr;
			gap: 24px;
		}
	}

	.form-col {
		display: flex;
		flex-direction: column;
		gap: 28px;
	}
	.block-title {
		font-family: var(--f-serif);
		font-size: 19px;
		font-weight: 600;
		color: var(--c-text);
		margin: 0 0 16px;
	}
	.fields-grid {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 14px;
	}
	.fields-grid.two {
		margin-top: 0;
	}
	.span-2 {
		grid-column: span 2;
	}
	.in {
		width: 100%;
		box-sizing: border-box;
		padding: 13px 14px;
		border-radius: var(--r-input);
		border: 1px solid var(--c-border);
		background: var(--c-surface);
		color: var(--c-text);
		font-family: var(--f-body);
		font-size: 13px;
		outline: none;
		transition: border-color 150ms;
	}
	.in:focus {
		border-color: var(--c-ink);
	}
	.in::placeholder {
		color: var(--c-text-muted);
	}

	.paiement-head {
		display: flex;
		align-items: center;
		gap: 10px;
		margin-bottom: 16px;
	}
	.paiement-head .block-title {
		margin: 0;
	}
	.secure-badge {
		font-family: var(--f-body);
		font-size: 11px;
		font-weight: 600;
		color: var(--c-ink);
		background: var(--c-badge-verified-bg);
		padding: 3px 9px;
		border-radius: 6px;
	}
	.card-box {
		border: 1px solid var(--c-border);
		border-radius: 12px;
		padding: 20px;
		background: var(--c-surface);
		display: flex;
		flex-direction: column;
		gap: 14px;
	}
	.card-radio {
		display: flex;
		align-items: center;
		gap: 10px;
	}
	.radio-dot {
		width: 16px;
		height: 16px;
		border-radius: 50%;
		border: 5px solid var(--c-ink);
		box-sizing: border-box;
	}
	.radio-label {
		font-family: var(--f-body);
		font-size: 14px;
		font-weight: 600;
		color: var(--c-text);
	}

	.form-error {
		font-family: var(--f-body);
		font-size: 13px;
		color: var(--c-error);
		margin: 0;
	}

	.summary-card {
		background: var(--c-surface);
		border: 1px solid var(--c-border);
		border-radius: 16px;
		padding: 28px;
		display: flex;
		flex-direction: column;
		gap: 12px;
	}
	.summary-title {
		font-family: var(--f-serif);
		font-size: 19px;
		font-weight: 600;
		color: var(--c-text);
		margin: 0 0 4px;
	}
	.summary-item {
		display: flex;
		justify-content: space-between;
		font-family: var(--f-body);
		font-size: 13px;
		color: var(--c-text-tertiary);
	}
	.summary-item-price {
		color: var(--c-text);
		font-weight: 600;
	}
	.result-ok {
		font-family: var(--f-body);
		font-size: 11.5px;
		color: #3f7a52;
		margin-top: -6px;
	}
	.result-err {
		font-family: var(--f-body);
		font-size: 11.5px;
		color: var(--c-error);
		margin-top: -6px;
	}
	.summary-divider {
		height: 1px;
		background: var(--c-border);
		margin: 4px 0;
	}
	.summary-row {
		display: flex;
		justify-content: space-between;
		font-family: var(--f-body);
		font-size: 14px;
		color: var(--c-text-tertiary);
	}
	.summary-total {
		display: flex;
		justify-content: space-between;
		font-family: var(--f-serif);
		font-size: 20px;
		font-weight: 700;
		color: var(--c-ink);
	}
	.btn-pay {
		padding: 15px;
		border: none;
		border-radius: 10px;
		background: var(--c-accent);
		color: #fff;
		font-family: var(--f-body);
		font-size: 15px;
		font-weight: 600;
		cursor: pointer;
		transition:
			filter 150ms,
			opacity 150ms;
	}
	.btn-pay:hover:not(:disabled) {
		filter: brightness(1.08);
	}
	.btn-pay:disabled {
		opacity: 0.55;
		cursor: not-allowed;
	}
	.summary-trust {
		font-family: var(--f-body);
		font-size: 12px;
		color: var(--c-text-muted);
		text-align: center;
		margin: 0;
	}
</style>
