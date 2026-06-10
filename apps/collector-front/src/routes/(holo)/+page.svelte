<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { fetchArticles, articleImage, type ArticleAPI } from '$lib/api/catalog';
	import { eur, pct, sparkPath } from '$lib/utils/format';
	import GPanel from '$lib/components/galerie/GPanel.svelte';
	import GMeter from '$lib/components/galerie/GMeter.svelte';
	import GChip from '$lib/components/galerie/GChip.svelte';
	import GSpark from '$lib/components/galerie/GSpark.svelte';
	import Kicker from '$lib/components/galerie/Kicker.svelte';

	let articles = $state<ArticleAPI[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);

	onMount(async () => {
		try {
			articles = await fetchArticles();
		} catch (e) {
			error = 'Impossible de charger le catalogue. Vérifiez que le catalog-service est démarré.';
			console.error(e);
		} finally {
			loading = false;
		}
	});

	// Filtres pilotés par l'URL : /?cat=TCG&max=300
	const filterCat = $derived($page.url.searchParams.get('cat'));
	const filterMax = $derived(Number($page.url.searchParams.get('max')) || null);

	const filtered = $derived(
		articles.filter(
			(a) => (!filterCat || a.category.name === filterCat) && (!filterMax || a.prix <= filterMax)
		)
	);

	function setCategoryFilter(cat: string) {
		goto(filterCat === cat ? '/' : `/?cat=${encodeURIComponent(cat)}`, { noScroll: true });
	}

	const meters = $derived(
		(() => {
			const counts: Record<string, number> = {};
			for (const a of articles) counts[a.category.name] = (counts[a.category.name] ?? 0) + 1;
			const total = articles.length || 1;
			return Object.entries(counts).map(([label, val]) => ({
				label,
				value: Math.round((val / total) * 100),
				count: val
			}));
		})()
	);

	// Hue dérivée du nom de catégorie pour le placeholder art
	function categoryHue(name: string): number {
		const map: Record<string, number> = {
			TCG: 18,
			Console: 48,
			Comics: 220,
			Vinyle: 350,
			'Designer Toy': 280,
			Horlogerie: 195
		};
		return map[name] ?? (name.charCodeAt(0) * 47) % 360;
	}

	// Sparkline fictive dérivée du prix (pour la démo)
	function demoSpark(prix: number, delta: number): number[] {
		const base = prix / (1 + delta / 100);
		return [
			base * 0.88,
			base * 0.91,
			base * 0.94,
			base * 0.97,
			base,
			prix * 0.98,
			prix * 0.99,
			prix
		];
	}
</script>

<svelte:head><title>Collector.shop · Vitrine</title></svelte:head>

<!-- Hero -->
<section class="hero">
	<div>
		<Kicker>Saison 03 — sélection de la semaine</Kicker>
		<h1 class="hero-title">Holo rares<br />&amp; scellés.</h1>
		<p class="hero-lede">
			Pièces vérifiées · grading PSA / CGC · livraison tracée sous boîtier antichoc. Chaque lot est
			authentifié avant mise en ligne.
		</p>
		<div class="hero-ctas">
			<a href="/drops" class="btn-primary">Parcourir la sélection</a>
			<a href="/vendre" class="btn-link">Vendre une pièce →</a>
			<a href="/grading" class="btn-link">Comment fonctionne le grading →</a>
		</div>
	</div>

	<GPanel>
		<Kicker>Catégories</Kicker>
		<div class="meters-list">
			{#if meters.length === 0}
				{#each ['TCG', 'Console', 'Comics', 'Vinyles'] as cat, i}
					<div class="meter-row">
						<span class="meter-label">{cat}</span>
						<div style="flex:1"><GMeter value={[42, 28, 18, 12][i]} /></div>
						<span class="meter-count">{[42, 28, 18, 12][i]}</span>
					</div>
				{/each}
			{:else}
				{#each meters as m}
					<button
						class="meter-row meter-btn"
						class:meter-active={filterCat === m.label}
						onclick={() => setCategoryFilter(m.label)}
						title={filterCat === m.label ? 'Retirer le filtre' : `Filtrer : ${m.label}`}
					>
						<span class="meter-label">{m.label}</span>
						<div style="flex:1"><GMeter value={m.value} /></div>
						<span class="meter-count">{m.count}</span>
					</button>
				{/each}
			{/if}
		</div>
	</GPanel>
</section>

<!-- États -->
{#if loading}
	<div class="state-msg">Chargement du catalogue…</div>
{:else if error}
	<div class="state-error">{error}</div>
{:else if articles.length === 0}
	<div class="state-msg">Aucun article disponible pour le moment.</div>
{:else}
	{#if filterCat || filterMax}
		<div class="filter-bar">
			<span class="filter-label">
				Filtre actif :
				{#if filterCat}{filterCat}{/if}
				{#if filterMax}{filterCat ? ' · ' : ''}moins de {eur(filterMax)}{/if}
				— {filtered.length} pièce{filtered.length > 1 ? 's' : ''}
			</span>
			<a href="/" class="filter-clear">× effacer</a>
		</div>
	{/if}

	<!-- Grille de cartes -->
	<section class="grid-section">
		{#each filtered as article (article.ID)}
			{@const hue = categoryHue(article.category.name)}
			{@const spark = demoSpark(article.prix, article.delta)}
			{@const up = article.delta >= 0}
			{@const img = articleImage(article)}
			<article class="card">
				<div class="card-eyebrow">
					<span class="card-cat">{article.category.name} · {article.year}</span>
					<span class="card-id">{article.slug || `#${article.ID}`}</span>
				</div>

				<!-- Photo produit (placeholder dégradé en dessous si l'image manque ou casse) -->
				<div
					class="card-art"
					style="background:linear-gradient(155deg, oklch(0.30 0.045 {hue}) 0%, oklch(0.24 0.045 {hue}) 100%)"
				>
					<div class="card-art-trame" aria-hidden="true"></div>
					<span class="card-art-label">photo produit<br />{article.slug || `#${article.ID}`}</span>
					{#if img}
						<img
							class="card-art-img"
							src={img}
							alt={article.name}
							loading="lazy"
							onerror={(e) => ((e.currentTarget as HTMLImageElement).style.display = 'none')}
						/>
					{/if}
					{#if article.saleType === 'direct'}
						<span class="card-direct">vente directe</span>
					{/if}
					{#if article.sold}
						<span class="card-sold">vendu</span>
					{/if}
					<span class="card-rarity">{article.rarity}</span>
				</div>

				<div class="card-body">
					<div class="card-name-row">
						<div>
							<p class="card-name">{article.name}</p>
							<p class="card-series">{article.series}</p>
						</div>
						<div class="card-chips">
							<GChip>{article.grade}</GChip>
						</div>
					</div>

					<div class="card-divider"></div>

					<div class="card-price-row">
						<div>
							<div class="card-price-label">Cote actuelle</div>
							<div class="card-price">{eur(article.prix)}</div>
						</div>
						<div class="card-spark-col">
							<GSpark values={spark} color={up ? '#86c099' : '#d79c86'} w={80} h={24} dot={false} />
							<span class="card-delta" style="color:{up ? '#86c099' : '#d79c86'}"
								>{pct(article.delta)}</span
							>
						</div>
					</div>

					<a class="card-cta" href={`/lot/${article.ID}`}>Voir le lot</a>
				</div>
			</article>
		{/each}
	</section>

	<!-- Footer ticker -->
	<footer class="ticker">
		<span class="ticker-live">
			<span class="ticker-dot"></span>
			Dernières ventes
		</span>
		<div class="ticker-track">
			<div class="ticker-inner">
				{#each [...articles, ...articles] as article}
					<span class="ticker-item">
						<span class="ticker-id">{article.slug || `#${article.ID}`}</span>
						<span class="ticker-name">{article.name}</span>
						<span class="ticker-price">{eur(article.prix)}</span>
						<span style="color:{article.delta >= 0 ? '#86c099' : '#d79c86'}"
							>{pct(article.delta)}</span
						>
					</span>
				{/each}
			</div>
		</div>
	</footer>
{/if}

<style>
	/* ── Hero ── */
	.hero {
		display: grid;
		grid-template-columns: 1.55fr 1fr;
		gap: 56px;
		padding: 30px 0 26px;
		align-items: start;
	}
	@media (max-width: 768px) {
		.hero {
			grid-template-columns: 1fr;
			gap: 24px;
		}
	}

	.hero-title {
		font-family: 'Newsreader', Georgia, serif;
		font-weight: 500;
		font-size: clamp(36px, 5vw, 50px);
		line-height: 1.05;
		color: #ece5da;
		margin: 10px 0 14px;
		text-wrap: balance;
	}
	.hero-lede {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 15px;
		color: #a39a8c;
		line-height: 1.6;
		margin-bottom: 24px;
		max-width: 480px;
	}
	.hero-ctas {
		display: flex;
		align-items: center;
		gap: 20px;
		flex-wrap: wrap;
	}
	.btn-primary {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13px;
		font-weight: 600;
		padding: 12px 22px;
		border-radius: 7px;
		background: #86b3a4;
		color: #191714;
		text-decoration: none;
		transition: filter 120ms;
	}
	.btn-primary:hover {
		filter: brightness(1.08);
	}
	.btn-link {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13px;
		color: #a39a8c;
		text-decoration: underline;
		text-underline-offset: 3px;
	}
	.btn-link:hover {
		color: #ece5da;
	}

	.meters-list {
		display: flex;
		flex-direction: column;
		gap: 10px;
		margin-top: 14px;
	}
	.meter-row {
		display: flex;
		align-items: center;
		gap: 12px;
	}
	.meter-btn {
		background: none;
		border: none;
		padding: 4px 6px;
		margin: -4px -6px;
		border-radius: 6px;
		cursor: pointer;
		width: 100%;
		text-align: left;
		transition: background 120ms;
	}
	.meter-btn:hover {
		background: rgba(255, 255, 255, 0.04);
	}
	.meter-active {
		background: rgba(134, 179, 164, 0.08);
	}
	.meter-active .meter-label {
		color: #86b3a4;
	}

	.filter-bar {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
		padding: 9px 14px;
		border: 1px solid rgba(134, 179, 164, 0.28);
		border-radius: 7px;
		background: rgba(134, 179, 164, 0.05);
		margin-bottom: 16px;
	}
	.filter-label {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 12.5px;
		color: #86b3a4;
	}
	.filter-clear {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 12px;
		color: #a39a8c;
		text-decoration: none;
	}
	.filter-clear:hover {
		color: #ece5da;
	}
	.meter-label {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 12px;
		color: #a39a8c;
		width: 90px;
		flex-shrink: 0;
	}
	.meter-count {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 12px;
		color: #766d60;
		width: 28px;
		text-align: right;
		flex-shrink: 0;
	}

	/* ── États ── */
	.state-msg {
		text-align: center;
		padding: 60px 0;
		font-family: 'Newsreader', Georgia, serif;
		font-style: italic;
		font-size: 16px;
		color: #a39a8c;
	}
	.state-error {
		padding: 12px 16px;
		border-radius: 7px;
		background: rgba(215, 156, 134, 0.06);
		border: 1px solid rgba(215, 156, 134, 0.3);
		color: #d79c86;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13px;
		margin-bottom: 20px;
	}

	/* ── Grille ── */
	.grid-section {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 18px;
		padding: 4px 0 24px;
	}
	@media (max-width: 900px) {
		.grid-section {
			grid-template-columns: repeat(2, 1fr);
		}
	}
	@media (max-width: 580px) {
		.grid-section {
			grid-template-columns: 1fr;
		}
	}

	/* ── Carte ── */
	.card {
		position: relative;
		background: #221f1b;
		border: 1px solid rgba(236, 229, 218, 0.1);
		border-radius: 9px;
		overflow: hidden;
		display: flex;
		flex-direction: column;
		transition: border-color 120ms;
		cursor: pointer;
	}
	.card:hover {
		border-color: rgba(236, 229, 218, 0.17);
	}

	.card-eyebrow {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 12px 14px 0;
	}
	.card-cat {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 10px;
		letter-spacing: 0.14em;
		text-transform: uppercase;
		color: #766d60;
	}
	.card-id {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 10px;
		color: #766d60;
	}

	/* Art placeholder */
	.card-art {
		position: relative;
		height: 132px;
		margin: 10px 14px 0;
		border-radius: 6px;
		border: 1px solid rgba(236, 229, 218, 0.08);
		display: flex;
		align-items: center;
		justify-content: center;
		overflow: hidden;
	}
	.card-art-trame {
		position: absolute;
		inset: 0;
		opacity: 0.08;
		background: repeating-linear-gradient(45deg, #ece5da 0 1px, transparent 1px 9px);
	}
	.card-art-label {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 10px;
		letter-spacing: 0.08em;
		text-align: center;
		line-height: 1.5;
		color: rgba(236, 229, 218, 0.35);
		position: relative;
	}
	.card-art-img {
		position: absolute;
		inset: 0;
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	.card-direct {
		position: absolute;
		top: 7px;
		left: 8px;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 9px;
		letter-spacing: 0.14em;
		text-transform: uppercase;
		padding: 2px 7px;
		border-radius: 3px;
		color: #191714;
		background: #86b3a4;
		font-weight: 600;
	}
	.card-sold {
		position: absolute;
		top: 7px;
		right: 8px;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 9px;
		letter-spacing: 0.14em;
		text-transform: uppercase;
		padding: 2px 7px;
		border-radius: 3px;
		color: #ece5da;
		background: rgba(215, 156, 134, 0.85);
		font-weight: 600;
	}
	.card-rarity {
		position: absolute;
		bottom: 7px;
		right: 8px;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 9px;
		letter-spacing: 0.14em;
		text-transform: uppercase;
		padding: 2px 7px;
		border: 1px solid rgba(236, 229, 218, 0.2);
		border-radius: 3px;
		color: rgba(236, 229, 218, 0.55);
		background: rgba(0, 0, 0, 0.25);
	}

	.card-body {
		padding: 12px 14px 14px;
		display: flex;
		flex-direction: column;
		gap: 10px;
		flex: 1;
	}

	.card-name-row {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		gap: 8px;
	}
	.card-name {
		font-family: 'Newsreader', Georgia, serif;
		font-size: 19px;
		color: #ece5da;
		margin: 0 0 3px;
		line-height: 1.1;
	}
	.card-series {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 11px;
		color: #766d60;
		margin: 0;
	}
	.card-chips {
		flex-shrink: 0;
	}

	.card-divider {
		border: none;
		border-top: 1px solid rgba(236, 229, 218, 0.08);
	}

	.card-price-row {
		display: flex;
		justify-content: space-between;
		align-items: flex-end;
	}
	.card-price-label {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 10px;
		letter-spacing: 0.14em;
		text-transform: uppercase;
		color: #766d60;
		margin-bottom: 3px;
	}
	.card-price {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 21px;
		color: #ece5da;
		font-weight: 500;
		line-height: 1;
	}
	.card-spark-col {
		display: flex;
		flex-direction: column;
		align-items: flex-end;
		gap: 3px;
	}
	.card-delta {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 11px;
	}

	.card-cta {
		display: block;
		text-align: center;
		text-decoration: none;
		box-sizing: border-box;
		width: 100%;
		padding: 9px;
		border-radius: 6px;
		border: 1px solid rgba(236, 229, 218, 0.12);
		background: transparent;
		color: #a39a8c;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 12px;
		font-weight: 500;
		cursor: pointer;
		transition:
			border-color 120ms,
			color 120ms,
			background 120ms;
		margin-top: auto;
	}
	.card-cta:hover {
		border-color: #86b3a4;
		color: #86b3a4;
		background: rgba(134, 179, 164, 0.04);
	}
	/* Lien étendu : toute la carte est cliquable via le CTA */
	.card-cta::after {
		content: '';
		position: absolute;
		inset: 0;
	}

	/* ── Ticker ── */
	.ticker {
		display: flex;
		align-items: center;
		gap: 18px;
		border-top: 1px solid rgba(236, 229, 218, 0.1);
		padding: 12px 0;
		overflow: hidden;
	}
	.ticker-live {
		display: flex;
		align-items: center;
		gap: 6px;
		flex-shrink: 0;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 11px;
		font-weight: 600;
		color: #a39a8c;
		letter-spacing: 0.04em;
	}
	.ticker-dot {
		width: 6px;
		height: 6px;
		border-radius: 50%;
		background: #86c099;
		display: inline-block;
	}
	.ticker-track {
		flex: 1;
		overflow: hidden;
	}
	.ticker-inner {
		display: flex;
		animation: ticker 30s linear infinite;
		width: max-content;
	}
	.ticker-item {
		display: inline-flex;
		gap: 10px;
		margin-right: 32px;
		align-items: center;
		white-space: nowrap;
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 11px;
	}
	.ticker-id {
		color: #86b3a4;
	}
	.ticker-name {
		color: #766d60;
	}
	.ticker-price {
		color: #ece5da;
	}

	@keyframes ticker {
		from {
			transform: translateX(0);
		}
		to {
			transform: translateX(-50%);
		}
	}
</style>
