<script lang="ts">
	import { onMount } from 'svelte';
	import { auth } from '$lib/stores/auth';
	import { fetchArticles, articleImage, type ArticleAPI } from '$lib/api/catalog';
	import { eur, pct } from '$lib/utils/format';
	import GChip from '$lib/components/galerie/GChip.svelte';
	import GSpark from '$lib/components/galerie/GSpark.svelte';
	import GSelect from '$lib/components/galerie/GSelect.svelte';
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

	// Recherche & filtres (état local, orientés marketplace)
	let search = $state('');
	let filterCat = $state('');
	let filterMax = $state(0);
	let availableOnly = $state(false);
	let sort = $state<'recent' | 'price-asc' | 'price-desc' | 'name'>('recent');

	const categories = $derived([...new Set(articles.map((a) => a.category.name))].sort());

	const filtered = $derived(
		articles
			.filter((a) => {
				const q = search.trim().toLowerCase();
				const matchQ =
					!q ||
					a.name.toLowerCase().includes(q) ||
					(a.series ?? '').toLowerCase().includes(q) ||
					(a.slug ?? '').toLowerCase().includes(q) ||
					a.category.name.toLowerCase().includes(q);
				return (
					matchQ &&
					(!filterCat || a.category.name === filterCat) &&
					(!filterMax || a.prix <= filterMax) &&
					(!availableOnly || !a.sold)
				);
			})
			.sort((a, b) => {
				switch (sort) {
					case 'price-asc':
						return a.prix - b.prix;
					case 'price-desc':
						return b.prix - a.prix;
					case 'name':
						return a.name.localeCompare(b.name);
					default:
						return b.ID - a.ID;
				}
			})
	);

	const hasFilters = $derived(!!(search || filterCat || filterMax || availableOnly));
	function resetFilters() {
		search = '';
		filterCat = '';
		filterMax = 0;
		availableOnly = false;
		sort = 'recent';
	}

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

<svelte:head><title>Collector.shop · Marché</title></svelte:head>

<!-- En-tête marché -->
<section class="market-head">
	<div class="mh-text">
		<Kicker>Marché</Kicker>
		<h1 class="mh-title">Le marché des collectionneurs</h1>
		<p class="mh-sub">
			Achat direct entre membres · chaque lot authentifié · grading PSA / CGC. Trouvez la pièce,
			ou mettez la vôtre en vente.
		</p>
	</div>
	{#if $auth.user}
		<a class="mh-sell" href="/vendre">+ Vendre une pièce</a>
	{:else}
		<a class="mh-sell" href="/login">Se connecter pour vendre</a>
	{/if}
</section>

<!-- Barre de recherche & filtres -->
<div class="toolbar">
	<div class="tb-search">
		<span class="tb-ico" aria-hidden="true">⌕</span>
		<input
			class="tb-input"
			type="search"
			placeholder="Rechercher une pièce, une série, une référence…"
			bind:value={search}
		/>
	</div>
	<GSelect
		bind:value={filterCat}
		ariaLabel="Catégorie"
		placeholder="Toutes catégories"
		options={[{ value: '', label: 'Toutes catégories' }, ...categories.map((c) => ({ value: c, label: c }))]}
	/>
	<GSelect
		bind:value={filterMax}
		ariaLabel="Prix maximum"
		placeholder="Tous prix"
		options={[
			{ value: 0, label: 'Tous prix' },
			{ value: 100, label: '≤ 100 €' },
			{ value: 500, label: '≤ 500 €' },
			{ value: 1000, label: '≤ 1 000 €' },
			{ value: 5000, label: '≤ 5 000 €' }
		]}
	/>
	<GSelect
		bind:value={sort}
		ariaLabel="Tri"
		options={[
			{ value: 'recent', label: 'Plus récents' },
			{ value: 'price-asc', label: 'Prix croissant' },
			{ value: 'price-desc', label: 'Prix décroissant' },
			{ value: 'name', label: 'Nom A→Z' }
		]}
	/>
	<label class="tb-check">
		<input type="checkbox" bind:checked={availableOnly} />
		Disponibles
	</label>
</div>

<!-- États -->
{#if loading}
	<div class="state-msg">Chargement du catalogue…</div>
{:else if error}
	<div class="state-error">{error}</div>
{:else if articles.length === 0}
	<div class="state-msg">Aucun article disponible pour le moment.</div>
{:else}
	<div class="result-bar">
		<span class="result-count">
			{filtered.length} pièce{filtered.length > 1 ? 's' : ''}
			{#if hasFilters}sur {articles.length}{/if}
		</span>
		{#if hasFilters}
			<button class="result-clear" onclick={resetFilters}>× réinitialiser</button>
		{/if}
	</div>

	{#if filtered.length === 0}
		<div class="state-msg">Aucune pièce ne correspond à votre recherche.</div>
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
	/* ── En-tête marché ── */
	.market-head {
		display: flex;
		align-items: flex-end;
		justify-content: space-between;
		gap: 24px;
		padding: 28px 0 18px;
		flex-wrap: wrap;
	}
	.mh-title {
		font-family: 'Newsreader', Georgia, serif;
		font-weight: 500;
		font-size: clamp(30px, 4vw, 42px);
		line-height: 1.05;
		color: #ece5da;
		margin: 8px 0 10px;
		text-wrap: balance;
	}
	.mh-sub {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 14px;
		color: #a39a8c;
		line-height: 1.55;
		max-width: 520px;
		margin: 0;
	}
	.mh-sell {
		flex-shrink: 0;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13px;
		font-weight: 600;
		padding: 11px 20px;
		border-radius: 8px;
		background: #86b3a4;
		color: #191714;
		text-decoration: none;
		transition: filter 120ms;
	}
	.mh-sell:hover {
		filter: brightness(1.08);
	}

	/* ── Barre de recherche & filtres ── */
	.toolbar {
		display: flex;
		gap: 10px;
		flex-wrap: wrap;
		align-items: center;
		padding: 12px 0 18px;
		border-top: 1px solid rgba(236, 229, 218, 0.08);
	}
	.tb-search {
		flex: 1 1 260px;
		display: flex;
		align-items: center;
		gap: 8px;
		background: rgba(255, 255, 255, 0.04);
		border: 1px solid rgba(236, 229, 218, 0.12);
		border-radius: 8px;
		padding: 0 12px;
		transition:
			border-color 150ms,
			box-shadow 150ms;
	}
	.tb-search:focus-within {
		border-color: rgba(134, 179, 164, 0.5);
		box-shadow: 0 0 0 3px rgba(134, 179, 164, 0.08);
	}
	.tb-ico {
		color: #766d60;
		font-size: 15px;
	}
	.tb-input {
		flex: 1;
		background: none;
		border: none;
		outline: none;
		padding: 11px 0;
		color: #ece5da;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 14px;
	}
	.tb-input::placeholder {
		color: rgba(236, 229, 218, 0.28);
	}
	.tb-check {
		display: flex;
		align-items: center;
		gap: 7px;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13px;
		color: #a39a8c;
		cursor: pointer;
		user-select: none;
	}
	.tb-check input {
		accent-color: #86b3a4;
		cursor: pointer;
	}

	/* ── Barre de résultats ── */
	.result-bar {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
		margin-bottom: 16px;
	}
	.result-count {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 12px;
		letter-spacing: 0.04em;
		color: #a39a8c;
	}
	.result-clear {
		background: none;
		border: none;
		cursor: pointer;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 12px;
		color: #a39a8c;
	}
	.result-clear:hover {
		color: #ece5da;
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
