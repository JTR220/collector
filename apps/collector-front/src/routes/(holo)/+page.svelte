<script lang="ts">
	import { onMount } from 'svelte';
	import {
		fetchArticles,
		fetchCategories,
		articleImage,
		type ArticleAPI,
		type CategoryAPI
	} from '$lib/api/catalog';
	import { eur, pct } from '$lib/utils/format';
	import GSelect from '$lib/components/galerie/GSelect.svelte';

	let articles = $state<ArticleAPI[]>([]);
	let categories = $state<CategoryAPI[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);

	// Collage décoratif du bandeau hero (en attente d'une vraie photo d'ambiance).
	const heroGlyphs = [
		{ icon: '🎴', label: 'Cartes', r: -4 },
		{ icon: '🎮', label: 'Consoles', r: 3 },
		{ icon: '📚', label: 'Comics', r: -2 },
		{ icon: '💿', label: 'Vinyles', r: 4 },
		{ icon: '🧸', label: 'Designer toys', r: -3 },
		{ icon: '⌚', label: 'Horlogerie', r: 2 }
	];

	onMount(async () => {
		try {
			[articles, categories] = await Promise.all([fetchArticles(), fetchCategories()]);
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
	let filterRarity = $state('');
	let filterGrade = $state('');
	let filterMax = $state(0);
	let availableOnly = $state(false);
	let sort = $state<'recent' | 'price-asc' | 'price-desc' | 'name'>('recent');

	const categoryNames = $derived([...categories].map((c) => c.name).sort());
	const rarities = $derived([...new Set(articles.map((a) => a.rarity).filter(Boolean))].sort());
	const grades = $derived([...new Set(articles.map((a) => a.grade).filter(Boolean))].sort());

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
					(!filterRarity || a.rarity === filterRarity) &&
					(!filterGrade || a.grade === filterGrade) &&
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

	const sortLabels: Record<string, string> = {
		recent: 'Plus récents',
		'price-asc': 'Prix croissant',
		'price-desc': 'Prix décroissant',
		name: 'Nom A→Z'
	};

	const hasFilters = $derived(
		!!(search || filterCat || filterRarity || filterGrade || filterMax || availableOnly)
	);
	function resetFilters() {
		search = '';
		filterCat = '';
		filterRarity = '';
		filterGrade = '';
		filterMax = 0;
		availableOnly = false;
		sort = 'recent';
	}
</script>

<svelte:head><title>Collector.shop · Marché</title></svelte:head>

<!-- Hero -->
<section class="hero">
	<div class="hero-text">
		<div class="hero-eyebrow">Marché entre collectionneurs</div>
		<h1 class="hero-title">Chinez la pièce rare, vendez vos trésors.</h1>
		<p class="hero-sub">
			Cartes à jouer, consoles, comics, vinyles, designer toys, horlogerie&nbsp;: chaque pièce est
			authentifiée et vendue en direct entre particuliers vérifiés.
		</p>
		<a class="hero-cta" href="#grille">Découvrir la sélection</a>
	</div>
	<div class="hero-art" aria-hidden="true">
		<div class="hero-collage">
			{#each heroGlyphs as g}
				<div class="hero-tile" style="--r:{g.r}deg">
					<span class="hero-tile-icon">{g.icon}</span>
					<span class="hero-tile-label">{g.label}</span>
				</div>
			{/each}
		</div>
	</div>
</section>

<!-- Filtres catégorie (pills) -->
<div class="pill-row">
	<button class="pill" class:pill-active={!filterCat} onclick={() => (filterCat = '')}>Tout</button>
	{#each categoryNames as cat}
		<button class="pill" class:pill-active={filterCat === cat} onclick={() => (filterCat = cat)}>
			{cat}
		</button>
	{/each}
	<div class="pill-spacer"></div>
	<GSelect
		bind:value={sort}
		ariaLabel="Tri"
		compact
		options={Object.entries(sortLabels).map(([value, label]) => ({ value, label }))}
	/>
</div>

<!-- Recherche & filtres avancés -->
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
		bind:value={filterRarity}
		ariaLabel="Rareté"
		placeholder="Toutes raretés"
		compact
		options={[
			{ value: '', label: 'Toutes raretés' },
			...rarities.map((r) => ({ value: r, label: r }))
		]}
	/>
	<GSelect
		bind:value={filterGrade}
		ariaLabel="Grade"
		placeholder="Tous grades"
		compact
		options={[{ value: '', label: 'Tous grades' }, ...grades.map((g) => ({ value: g, label: g }))]}
	/>
	<GSelect
		bind:value={filterMax}
		ariaLabel="Prix maximum"
		placeholder="Tous prix"
		compact
		options={[
			{ value: 0, label: 'Tous prix' },
			{ value: 100, label: '≤ 100 €' },
			{ value: 500, label: '≤ 500 €' },
			{ value: 1000, label: '≤ 1 000 €' },
			{ value: 5000, label: '≤ 5 000 €' }
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
	<section class="grid-section" id="grille">
		{#each filtered as article (article.ID)}
			{@const img = articleImage(article)}
			<a class="card" href={`/lot/${article.ID}`}>
				<div class="card-art">
					{#if img}
						<img
							class="card-art-img"
							src={img}
							alt={article.name}
							loading="lazy"
							onerror={(e) => ((e.currentTarget as HTMLImageElement).style.display = 'none')}
						/>
					{:else}
						<span class="card-art-label">photo produit</span>
					{/if}
					{#if article.sold}
						<span class="card-sold">vendu</span>
					{/if}
				</div>

				<div class="card-body">
					<div class="card-top-row">
						{#if article.grade}<span class="card-condition">{article.grade}</span>{/if}
						<span class="card-wish" aria-hidden="true"></span>
					</div>
					<p class="card-name">{article.name}</p>
					<span class="card-price">{eur(article.prix)}</span>
					<p class="card-seller">Vendu par {article.seller} · Particulier vérifié</p>
				</div>
			</a>
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
						<span class="ticker-name">{article.name}</span>
						<span class="ticker-price">{eur(article.prix)}</span>
						<span style="color:{article.delta >= 0 ? '#3f7a52' : '#b0432a'}"
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
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 32px;
		margin: 24px 0;
		padding: 48px;
		border-radius: 20px;
		background: linear-gradient(135deg, #1e3b2c, #2a4e3a);
		color: var(--c-bg);
		flex-wrap: wrap;
	}
	.hero-text {
		max-width: 560px;
	}
	.hero-eyebrow {
		font-family: var(--f-body);
		font-size: 12px;
		letter-spacing: 0.15em;
		text-transform: uppercase;
		color: #c9e0ce;
		margin-bottom: 12px;
	}
	.hero-title {
		font-family: var(--f-serif);
		font-weight: 600;
		font-size: clamp(28px, 4vw, 38px);
		line-height: 1.15;
		margin: 0 0 16px;
		text-wrap: balance;
	}
	.hero-sub {
		font-family: var(--f-body);
		font-size: 15px;
		color: #d8e6db;
		line-height: 1.55;
		margin: 0 0 24px;
	}
	.hero-cta {
		display: inline-block;
		padding: 12px 28px;
		background: var(--c-accent);
		color: #fff;
		border-radius: 8px;
		font-size: 14px;
		font-weight: 600;
		text-decoration: none;
		transition: filter 120ms;
	}
	.hero-cta:hover {
		filter: brightness(1.08);
		color: #fff;
	}
	.hero-art {
		flex: none;
		width: 340px;
		height: 220px;
	}
	.hero-collage {
		width: 100%;
		height: 100%;
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 12px;
		padding: 4px;
	}
	.hero-tile {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 6px;
		background: var(--c-bg);
		border-radius: 10px;
		box-shadow: 0 6px 14px rgba(0, 0, 0, 0.18);
		transform: rotate(var(--r));
		transition: transform 150ms;
	}
	.hero-tile:hover {
		transform: rotate(0deg) scale(1.05);
	}
	.hero-tile-icon {
		font-size: 24px;
	}
	.hero-tile-label {
		font-family: var(--f-body);
		font-size: 10px;
		font-weight: 600;
		letter-spacing: 0.02em;
		color: var(--c-ink);
		text-align: center;
	}
	@media (max-width: 900px) {
		.hero-art {
			display: none;
		}
	}

	/* ── Pills catégories ── */
	.pill-row {
		display: flex;
		align-items: center;
		gap: 10px;
		flex-wrap: wrap;
		padding: 8px 0 14px;
	}
	.pill {
		padding: 8px 16px;
		border-radius: var(--r-pill);
		border: 1px solid var(--c-border);
		background: var(--c-surface);
		font-family: var(--f-body);
		font-size: 13px;
		font-weight: 500;
		color: var(--c-text-tertiary);
		cursor: pointer;
		white-space: nowrap;
		transition:
			background 120ms,
			color 120ms,
			border-color 120ms;
	}
	.pill:hover {
		border-color: var(--c-ink);
	}
	.pill-active {
		background: var(--c-ink);
		border-color: var(--c-ink);
		color: var(--c-bg);
	}
	.pill-spacer {
		flex: 1;
	}

	/* ── Barre de recherche & filtres ── */
	.toolbar {
		display: flex;
		gap: 10px;
		flex-wrap: wrap;
		align-items: center;
		padding: 4px 0 18px;
		border-top: 1px solid var(--c-border);
		padding-top: 16px;
	}
	.tb-search {
		flex: 1 1 260px;
		display: flex;
		align-items: center;
		gap: 8px;
		background: var(--c-surface);
		border: 1px solid var(--c-border);
		border-radius: var(--r-pill);
		padding: 0 16px;
		transition: border-color 150ms;
	}
	.tb-search:focus-within {
		border-color: var(--c-ink);
	}
	.tb-ico {
		color: var(--c-text-muted);
		font-size: 15px;
	}
	.tb-input {
		flex: 1;
		background: none;
		border: none;
		outline: none;
		padding: 11px 0;
		color: var(--c-text);
		font-family: var(--f-body);
		font-size: 13px;
	}
	.tb-input::placeholder {
		color: var(--c-text-muted);
	}
	.tb-check {
		display: flex;
		align-items: center;
		gap: 7px;
		font-family: var(--f-body);
		font-size: 13px;
		color: var(--c-text-tertiary);
		cursor: pointer;
		user-select: none;
		white-space: nowrap;
	}
	.tb-check input {
		accent-color: var(--c-ink);
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
		font-family: var(--f-body);
		font-size: 13px;
		color: var(--c-text-muted);
	}
	.result-clear {
		background: none;
		border: none;
		cursor: pointer;
		font-family: var(--f-body);
		font-size: 12px;
		color: var(--c-text-muted);
	}
	.result-clear:hover {
		color: var(--c-ink);
	}

	/* ── États ── */
	.state-msg {
		text-align: center;
		padding: 60px 0;
		font-family: var(--f-serif);
		font-style: italic;
		font-size: 16px;
		color: var(--c-text-muted);
	}
	.state-error {
		padding: 12px 16px;
		border-radius: 7px;
		background: #fbe9e3;
		border: 1px solid rgba(176, 67, 42, 0.3);
		color: var(--c-error);
		font-family: var(--f-body);
		font-size: 13px;
		margin-bottom: 20px;
	}

	/* ── Grille ── */
	.grid-section {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 24px;
		padding: 4px 0 40px;
	}
	@media (max-width: 1100px) {
		.grid-section {
			grid-template-columns: repeat(3, 1fr);
		}
	}
	@media (max-width: 780px) {
		.grid-section {
			grid-template-columns: repeat(2, 1fr);
			gap: 14px;
		}
	}
	@media (max-width: 480px) {
		.grid-section {
			grid-template-columns: 1fr;
		}
	}

	/* ── Carte ── */
	.card {
		position: relative;
		background: var(--c-surface);
		border: 1px solid var(--c-border);
		border-radius: var(--r-card);
		overflow: hidden;
		display: flex;
		flex-direction: column;
		text-decoration: none;
		transition: border-color 120ms;
	}
	.card:hover {
		border-color: var(--c-ink);
	}

	.card-art {
		position: relative;
		height: 180px;
		background: var(--c-bg);
		display: flex;
		align-items: center;
		justify-content: center;
		overflow: hidden;
	}
	.card-art-label {
		font-family: var(--f-body);
		font-size: 11px;
		letter-spacing: 0.08em;
		color: var(--c-icon-muted);
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
		top: 10px;
		right: 10px;
		font-family: var(--f-body);
		font-size: 10px;
		font-weight: 600;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		padding: 3px 9px;
		border-radius: 5px;
		color: #fff;
		background: var(--c-accent);
	}

	.card-body {
		padding: 16px;
		display: flex;
		flex-direction: column;
		gap: 8px;
		flex: 1;
	}
	.card-top-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 6px;
	}
	.card-condition {
		font-family: var(--f-body);
		font-size: 11px;
		font-weight: 600;
		color: var(--c-ink);
		background: var(--c-badge-verified-bg);
		padding: 3px 8px;
		border-radius: 5px;
	}
	.card-wish {
		width: 16px;
		height: 16px;
		border-radius: 50%;
		border: 1.5px solid var(--c-icon-muted);
		flex-shrink: 0;
		margin-left: auto;
	}
	.card-name {
		font-size: 14px;
		font-weight: 600;
		color: var(--c-text);
		line-height: 1.35;
		margin: 0;
		min-height: 38px;
	}
	.card-price {
		font-family: var(--f-serif);
		font-size: 19px;
		font-weight: 600;
		color: var(--c-ink);
	}
	.card-seller {
		font-family: var(--f-body);
		font-size: 12px;
		color: var(--c-text-muted);
		margin: 0;
	}

	/* ── Ticker ── */
	.ticker {
		display: flex;
		align-items: center;
		gap: 18px;
		border-top: 1px solid var(--c-border);
		padding: 14px 0;
		overflow: hidden;
	}
	.ticker-live {
		display: flex;
		align-items: center;
		gap: 6px;
		flex-shrink: 0;
		font-family: var(--f-body);
		font-size: 11px;
		font-weight: 600;
		color: var(--c-text-muted);
		letter-spacing: 0.04em;
	}
	.ticker-dot {
		width: 6px;
		height: 6px;
		border-radius: 50%;
		background: #3f7a52;
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
		font-family: var(--f-body);
		font-size: 12px;
	}
	.ticker-name {
		color: var(--c-text-muted);
	}
	.ticker-price {
		color: var(--c-text);
		font-weight: 600;
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
