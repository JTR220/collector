<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { auth, isAuthenticated } from '$lib/stores/auth';
	import { fetchArticle, articleImage, type ArticleAPI } from '$lib/api/catalog';
	import { addToWishlist, removeFromWishlist, fetchMyWishlist } from '$lib/api/wishlist';
	import { buyArticle } from '$lib/api/market';
	import { fetchPriceHistory } from '$lib/api/priceTracker';
	import { eur, eurC, pct, sparkPath } from '$lib/utils/format';
	import GPanel from '$lib/components/galerie/GPanel.svelte';
	import GChip from '$lib/components/galerie/GChip.svelte';
	import Kicker from '$lib/components/galerie/Kicker.svelte';

	let article = $state<ArticleAPI | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let inWishlist = $state(false);
	let actionMsg = $state<string | null>(null);
	let actionBusy = $state(false);

	let trackedPrices = $state<number[]>([]);

	onMount(async () => {
		try {
			article = await fetchArticle($page.params.id ?? '');
			if ($auth.token) {
				const wishlist = await fetchMyWishlist($auth.token);
				inWishlist = wishlist.some((w) => w.articleId === article?.ID);
			}
		} catch (e) {
			error = 'Lot introuvable ou catalog-service indisponible.';
			console.error(e);
		} finally {
			loading = false;
		}

		// Historique temps reel du price-tracker (echec silencieux si indisponible)
		if (article) {
			try {
				const hist = await fetchPriceHistory(article.ID);
				trackedPrices = hist.map((h) => h.new_price);
			} catch {
				trackedPrices = [];
			}
		}
	});

	function requireAuth(): string | null {
		if (!$isAuthenticated || !$auth.token) {
			goto('/login');
			return null;
		}
		return $auth.token;
	}

	async function toggleWishlist() {
		const token = requireAuth();
		if (!token || !article) return;
		actionBusy = true;
		try {
			if (inWishlist) {
				await removeFromWishlist(token, article.ID);
				inWishlist = false;
				actionMsg = 'Retiré de la wishlist.';
			} else {
				const res = await addToWishlist(token, article.ID);
				inWishlist = true;
				actionMsg = res.already ? 'Déjà dans la wishlist.' : 'Ajouté à la wishlist.';
			}
		} catch (e) {
			actionMsg = e instanceof Error ? e.message : 'Erreur wishlist.';
		} finally {
			actionBusy = false;
		}
	}

	async function buyNow() {
		const token = requireAuth();
		if (!token || !article) return;
		actionBusy = true;
		actionMsg = null;
		try {
			await buyArticle(token, article.ID);
			article.sold = true;
			actionMsg = 'Achat confirmé — retrouvez la commande dans votre profil.';
		} catch (e) {
			actionMsg = e instanceof Error ? e.message : 'Erreur lors de l’achat.';
		} finally {
			actionBusy = false;
		}
	}

	const W = 560;
	const H = 140;
	// Serie du catalogue, prolongee par les changements captes par le price-tracker
	const historySeries = $derived.by(() => {
		const base = article?.priceHistory ?? [];
		if (!trackedPrices.length) return base;
		const merged = [...base];
		for (const p of trackedPrices) {
			if (merged[merged.length - 1] !== p) merged.push(p);
		}
		return merged;
	});
	const historyPath = $derived(article ? sparkPath(historySeries, W, H) : '');
	const up = $derived((article?.delta ?? 0) >= 0);
</script>

<svelte:head><title>{article ? article.name : 'Lot'} · Collector.shop</title></svelte:head>

{#if loading}
	<p class="state-msg">Chargement du lot…</p>
{:else if error || !article}
	<p class="state-msg error">{error}</p>
	<a class="back-link" href="/">← Retour à la vitrine</a>
{:else}
	{@const img = articleImage(article)}
	<a class="back-link" href="/">← Vitrine</a>

	<section class="lot-grid">
		<!-- Photo produit (glyph en repli) -->
		<div class="lot-art">
			<span class="lot-glyph">{article.glyph}</span>
			{#if img}
				<img
					class="lot-art-img"
					src={img}
					alt={article.name}
					onerror={(e) => ((e.currentTarget as HTMLImageElement).style.display = 'none')}
				/>
			{/if}
			{#if article.sold}
				<span class="lot-sold">vendu</span>
			{/if}
			<span class="lot-rarity">{article.rarity}</span>
		</div>

		<!-- Infos -->
		<div>
			<Kicker color="#86b3a4">{article.category.name} · {article.year} · {article.slug}</Kicker>
			<h1 class="lot-title">{article.name}</h1>
			<p class="lot-series">{article.series}</p>

			<div class="lot-chips">
				<GChip>{article.grade}</GChip>
				<GChip>{article.rarity}</GChip>
			</div>

			<p class="lot-desc">{article.description}</p>

			<div class="lot-price-row">
				<div>
					<div class="price-label">Cote actuelle</div>
					<div class="price-val">{eur(article.prix)}</div>
					<div class="price-delta" style="color:{up ? '#86c099' : '#d79c86'}">
						{pct(article.delta)} sur la période
					</div>
				</div>
				<div style="text-align:right">
					<div class="price-label">Frais de port</div>
					<div class="price-sub">{eurC(article.fraisPort)}</div>
					<div class="price-label" style="margin-top:8px">Estim. resell</div>
					<div class="price-sub resell">{eur(article.resellPrice)}</div>
				</div>
			</div>

			<div class="lot-actions">
				{#if article.sold}
					<button class="btn-primary" disabled>Vendu</button>
				{:else}
					<button class="btn-primary" disabled={actionBusy} onclick={buyNow}>
						Acheter maintenant · {eur(article.prix)}
					</button>
				{/if}
				<button class="btn-ghost" disabled={actionBusy} onclick={toggleWishlist}>
					{inWishlist ? '♥ Dans la wishlist' : '♡ Wishlist'}
				</button>
			</div>

			{#if actionMsg}
				<p class="action-msg">{actionMsg}</p>
			{/if}

			<div class="lot-seller">
				<Kicker>Vendeur</Kicker>
				<div class="seller-row">
					<span class="seller-name">@{article.seller}</span>
					<span class="seller-score">★ {article.sellerScore.toFixed(2)} / 5</span>
				</div>
			</div>
		</div>
	</section>

	<!-- Historique de prix -->
	<GPanel style="margin-top:18px">
		<Kicker>Historique de cote · 8 derniers relevés</Kicker>
		<svg
			viewBox="0 0 {W} {H}"
			width="100%"
			height={H}
			preserveAspectRatio="none"
			style="display:block;margin-top:12px"
		>
			{#each [0, 1, 2, 3] as i}
				<line
					x1="0"
					y1={(H * i) / 3}
					x2={W}
					y2={(H * i) / 3}
					stroke="rgba(236,229,218,0.06)"
					stroke-width="0.5"
				/>
			{/each}
			<path d={historyPath} stroke={up ? '#86c099' : '#d79c86'} stroke-width="2" fill="none" />
		</svg>
		<div class="history-vals">
			{#each historySeries as p}
				<span>{eur(p)}</span>
			{/each}
		</div>
	</GPanel>
{/if}

<style>
	.state-msg {
		text-align: center;
		padding: 60px 0;
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 12px;
		color: #766d60;
		letter-spacing: 0.12em;
	}
	.state-msg.error {
		color: #d79c86;
	}

	.back-link {
		display: inline-block;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 12.5px;
		color: #a39a8c;
		text-decoration: none;
		margin-bottom: 16px;
	}
	.back-link:hover {
		color: #ece5da;
	}

	.lot-grid {
		display: grid;
		grid-template-columns: 0.8fr 1.2fr;
		gap: 32px;
		align-items: start;
	}
	@media (max-width: 768px) {
		.lot-grid {
			grid-template-columns: 1fr;
		}
	}

	.lot-art {
		aspect-ratio: 3/4;
		border-radius: 12px;
		position: relative;
		background: radial-gradient(120% 90% at 30% 20%, #4a6a5a 0%, #2a3a32 55%, #191714 100%);
		border: 1px solid rgba(236, 229, 218, 0.1);
		display: flex;
		align-items: center;
		justify-content: center;
	}
	.lot-glyph {
		font-family: 'Newsreader', Georgia, serif;
		font-size: 120px;
		color: rgba(236, 229, 218, 0.85);
	}
	.lot-art-img {
		position: absolute;
		inset: 0;
		width: 100%;
		height: 100%;
		object-fit: cover;
		border-radius: 12px;
	}
	.lot-sold {
		position: absolute;
		top: 12px;
		left: 12px;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 10px;
		letter-spacing: 0.14em;
		text-transform: uppercase;
		padding: 3px 9px;
		border-radius: 4px;
		color: #ece5da;
		background: rgba(215, 156, 134, 0.85);
		font-weight: 600;
	}
	.lot-rarity {
		position: absolute;
		bottom: 12px;
		right: 12px;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 10px;
		letter-spacing: 0.14em;
		text-transform: uppercase;
		padding: 3px 9px;
		border: 1px solid rgba(236, 229, 218, 0.2);
		border-radius: 4px;
		color: rgba(236, 229, 218, 0.6);
		background: rgba(0, 0, 0, 0.25);
	}

	.lot-title {
		font-family: 'Newsreader', Georgia, serif;
		font-weight: 500;
		font-size: 42px;
		line-height: 1.05;
		color: #ece5da;
		margin: 8px 0 4px;
	}
	.lot-series {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13px;
		color: #766d60;
		margin: 0 0 14px;
	}
	.lot-chips {
		display: flex;
		gap: 6px;
		margin-bottom: 16px;
	}
	.lot-desc {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 14px;
		color: #a39a8c;
		line-height: 1.6;
		max-width: 520px;
		margin-bottom: 20px;
	}

	.lot-price-row {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		padding: 16px;
		border: 1px solid rgba(236, 229, 218, 0.1);
		border-radius: 9px;
		background: rgba(255, 255, 255, 0.03);
		margin-bottom: 12px;
	}
	.price-label {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 10px;
		letter-spacing: 0.14em;
		text-transform: uppercase;
		color: #766d60;
	}
	.price-val {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 34px;
		color: #ece5da;
		font-weight: 500;
		line-height: 1.1;
		margin-top: 4px;
	}
	.price-delta {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 12px;
		margin-top: 4px;
	}
	.price-sub {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 16px;
		color: #ece5da;
		margin-top: 3px;
	}
	.price-sub.resell {
		color: #86b3a4;
	}

	.lot-actions {
		display: flex;
		gap: 10px;
		flex-wrap: wrap;
	}
	.btn-primary {
		padding: 12px 22px;
		border-radius: 7px;
		border: none;
		background: #86b3a4;
		color: #191714;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13px;
		font-weight: 600;
		cursor: pointer;
		transition:
			filter 120ms,
			opacity 120ms;
	}
	.btn-primary:hover:not(:disabled) {
		filter: brightness(1.08);
	}
	.btn-primary:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}
	.btn-ghost {
		padding: 12px 18px;
		border-radius: 7px;
		border: 1px solid rgba(236, 229, 218, 0.14);
		background: transparent;
		color: #a39a8c;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13px;
		cursor: pointer;
		transition:
			border-color 120ms,
			color 120ms;
	}
	.btn-ghost:hover:not(:disabled) {
		border-color: #86b3a4;
		color: #86b3a4;
	}

	.action-msg {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 12.5px;
		color: #86b3a4;
		margin-top: 12px;
	}

	.lot-seller {
		margin-top: 20px;
	}
	.seller-row {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-top: 8px;
		padding: 12px 14px;
		border: 1px solid rgba(236, 229, 218, 0.1);
		border-radius: 8px;
		max-width: 360px;
	}
	.seller-name {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 13px;
		color: #86b3a4;
	}
	.seller-score {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 12.5px;
		color: #a39a8c;
	}

	.history-vals {
		display: flex;
		justify-content: space-between;
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 10px;
		color: #766d60;
		margin-top: 6px;
	}
</style>
