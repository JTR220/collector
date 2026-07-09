<script lang="ts">
	import { env } from '$env/dynamic/public';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/stores/auth';
	import type { Article, Category } from '$lib/types/catalog';
	import GSelect from '$lib/components/galerie/GSelect.svelte';
	import GConfirmModal from '$lib/components/galerie/GConfirmModal.svelte';

	const catalogApiBaseUrl = env.PUBLIC_CATALOG_API_BASE_URL ?? 'http://localhost:8081';

	type CategoryFormState = {
		name: string;
		description: string;
	};

	type ArticleFormState = {
		name: string;
		description: string;
		prix: string;
		fraisPort: string;
		categoryId: string;
	};

	const emptyCategoryForm = (): CategoryFormState => ({
		name: '',
		description: ''
	});

	const emptyArticleForm = (): ArticleFormState => ({
		name: '',
		description: '',
		prix: '',
		fraisPort: '',
		categoryId: ''
	});

	type AdminStats = {
		gmv: number;
		totalOrders: number;
		avgOrderValue: number;
		ordersByStatus: { paid: number; shipped: number; delivered: number; cancelled: number };
		totalArticles: number;
		activeListings: number;
		soldArticles: number;
		sellThrough: number;
		avgListing: number;
		categories: number;
		activeSellers: number;
		byCategory: { name: string; count: number }[];
		recentOrders: {
			id: number;
			article: string;
			price: number;
			status: string;
			buyerId: number;
			createdAt: string;
		}[];
	};

	const ORDER_STATUS_FR: Record<string, string> = {
		paid: 'Payée',
		shipped: 'Expédiée',
		delivered: 'Livrée',
		cancelled: 'Annulée'
	};
	const eur = (n: number) => `${n.toLocaleString('fr-FR')} €`;

	let categories: Category[] = [];
	let articles: Article[] = [];
	let stats: AdminStats | null = null;

	// Recherche d'article (titre, réf/ID ou catégorie) pour retrouver vite une pièce.
	let rechercheArticle = '';
	$: articlesAffiches = articles.filter((a) => {
		const q = rechercheArticle.trim().toLowerCase();
		if (!q) return true;
		return (
			a.name.toLowerCase().includes(q) ||
			String(a.ID).includes(q) ||
			(a.category?.name ?? '').toLowerCase().includes(q)
		);
	});

	// Sélection multiple pour la suppression groupée depuis le catalogue.
	let selectionArticles = new Set<number>();
	$: toutSelectionne =
		articlesAffiches.length > 0 && articlesAffiches.every((a) => selectionArticles.has(a.ID));

	function toggleSelectionArticle(id: number) {
		const next = new Set(selectionArticles);
		if (next.has(id)) next.delete(id);
		else next.add(id);
		selectionArticles = next;
	}

	function toggleToutSelectionner() {
		selectionArticles = toutSelectionne ? new Set() : new Set(articlesAffiches.map((a) => a.ID));
	}

	// Modale de confirmation (remplace window.confirm) pour les suppressions.
	let confirmOpen = false;
	let confirmTitle = '';
	let confirmMessage = '';
	let confirmCheckboxLabel = '';
	let confirmCheckboxChecked = false;
	let confirmAction: (() => void) | null = null;

	function demanderSuppressionArticle(id: number, name: string) {
		confirmTitle = 'Retirer cet article ?';
		confirmMessage = `« ${name} » sera retiré du catalogue. Cette action est irréversible.`;
		confirmCheckboxLabel = '';
		confirmCheckboxChecked = false;
		confirmAction = () => supprimerArticle(id);
		confirmOpen = true;
	}

	function demanderSuppressionSelection() {
		const n = selectionArticles.size;
		confirmTitle = `Retirer ${n} article${n > 1 ? 's' : ''} ?`;
		confirmMessage = `${n} article${n > 1 ? 's' : ''} sélectionné${n > 1 ? 's' : ''} seront retirés du catalogue. Cette action est irréversible.`;
		confirmCheckboxLabel = 'Je confirme vouloir supprimer définitivement ces articles';
		confirmCheckboxChecked = false;
		confirmAction = () => supprimerSelection();
		confirmOpen = true;
	}

	async function supprimerSelection() {
		const ids = Array.from(selectionArticles);
		selectionArticles = new Set();
		erreur = '';
		succes = '';
		try {
			for (const id of ids) {
				const response = await fetch(`${catalogApiBaseUrl}/article/${id}`, {
					method: 'DELETE',
					headers: authHeaders(false)
				});
				if (!response.ok) {
					const payload = await response.json().catch(() => null);
					throw new Error(payload?.error ?? 'Suppression impossible.');
				}
			}
			succes = `${ids.length} article${ids.length > 1 ? 's' : ''} retiré${ids.length > 1 ? 's' : ''}.`;
		} catch (error) {
			erreur = error instanceof Error ? error.message : 'Erreur inconnue.';
		} finally {
			await Promise.all([chargerArticles(), chargerStats()]);
		}
	}

	let categoryForm = emptyCategoryForm();
	let articleForm = emptyArticleForm();

	let chargement = true;
	let soumissionCategory = false;
	let soumissionArticle = false;
	let erreur = '';
	let succes = '';

	// En-tetes authentifies : les ecritures et le back-office exigent un token admin.
	function authHeaders(json = true): Record<string, string> {
		const h: Record<string, string> = {};
		if (json) h['Content-Type'] = 'application/json';
		if ($auth.token) h.Authorization = `Bearer ${$auth.token}`;
		return h;
	}

	function normalizeCategory(category: Record<string, unknown>): Category {
		return {
			ID: Number(category.ID ?? 0),
			CreatedAt: typeof category.CreatedAt === 'string' ? category.CreatedAt : undefined,
			UpdatedAt: typeof category.UpdatedAt === 'string' ? category.UpdatedAt : undefined,
			name: String(category.name ?? ''),
			description: String(category.description ?? '')
		};
	}

	function normalizeArticle(article: Record<string, unknown>): Article {
		const category =
			article.category && typeof article.category === 'object'
				? normalizeCategory(article.category as Record<string, unknown>)
				: undefined;

		return {
			ID: Number(article.ID ?? 0),
			CreatedAt: typeof article.CreatedAt === 'string' ? article.CreatedAt : undefined,
			UpdatedAt: typeof article.UpdatedAt === 'string' ? article.UpdatedAt : undefined,
			name: String(article.name ?? ''),
			description: String(article.description ?? ''),
			prix: Number(article.prix ?? 0),
			fraisPort: Number(article.fraisPort ?? 0),
			categoryId: Number(article.categoryId ?? article.CategoryID ?? category?.ID ?? 0),
			category
		};
	}

	async function chargerCategories() {
		const response = await fetch(`${catalogApiBaseUrl}/category`);
		if (!response.ok) {
			throw new Error('Impossible de charger les categories.');
		}
		const payload = (await response.json()) as Record<string, unknown>[];
		categories = payload.map(normalizeCategory);
	}

	// /admin/articles (et non /article) : la moderation doit voir aussi les
	// pieces deja vendues, exclues du catalogue public depuis GetAllArticles.
	async function chargerArticles() {
		const response = await fetch(`${catalogApiBaseUrl}/admin/articles`, {
			headers: authHeaders(false)
		});
		if (!response.ok) {
			throw new Error('Impossible de charger les articles.');
		}
		const payload = (await response.json()) as Record<string, unknown>[];
		articles = payload.map(normalizeArticle);
	}

	async function chargerStats() {
		const response = await fetch(`${catalogApiBaseUrl}/admin/stats`, {
			headers: authHeaders(false)
		});
		if (!response.ok) {
			throw new Error('Impossible de charger les statistiques.');
		}
		stats = (await response.json()) as AdminStats;
	}

	async function chargerCatalogue() {
		chargement = true;
		erreur = '';

		try {
			await Promise.all([chargerCategories(), chargerArticles(), chargerStats()]);
		} catch (error) {
			erreur = error instanceof Error ? error.message : 'Erreur inconnue.';
		} finally {
			chargement = false;
		}
	}

	async function supprimerArticle(id: number) {
		erreur = '';
		succes = '';
		try {
			const response = await fetch(`${catalogApiBaseUrl}/article/${id}`, {
				method: 'DELETE',
				headers: authHeaders(false)
			});
			if (!response.ok) {
				const payload = await response.json().catch(() => null);
				throw new Error(payload?.error ?? 'Suppression impossible.');
			}
			succes = 'Article retiré.';
			await Promise.all([chargerArticles(), chargerStats()]);
		} catch (error) {
			erreur = error instanceof Error ? error.message : 'Erreur inconnue.';
		}
	}

	async function soumettreCategory() {
		soumissionCategory = true;
		erreur = '';
		succes = '';

		try {
			const response = await fetch(`${catalogApiBaseUrl}/category`, {
				method: 'POST',
				headers: authHeaders(),
				body: JSON.stringify(categoryForm)
			});

			if (!response.ok) {
				const payload = await response.json().catch(() => null);
				throw new Error(payload?.error ?? 'Creation de categorie impossible.');
			}

			categoryForm = emptyCategoryForm();
			succes = 'Categorie creee.';
			await Promise.all([chargerCategories(), chargerArticles()]);
		} catch (error) {
			erreur = error instanceof Error ? error.message : 'Erreur inconnue.';
		} finally {
			soumissionCategory = false;
		}
	}

	async function soumettreArticle() {
		if (!articleForm.categoryId) {
			erreur = 'Choisissez une catégorie.';
			return;
		}
		soumissionArticle = true;
		erreur = '';
		succes = '';

		try {
			const response = await fetch(`${catalogApiBaseUrl}/article`, {
				method: 'POST',
				headers: authHeaders(),
				body: JSON.stringify({
					name: articleForm.name,
					description: articleForm.description,
					prix: Number(articleForm.prix),
					fraisPort: Number(articleForm.fraisPort),
					categoryId: Number(articleForm.categoryId)
				})
			});

			if (!response.ok) {
				const payload = await response.json().catch(() => null);
				throw new Error(payload?.error ?? "Creation d'article impossible.");
			}

			articleForm = emptyArticleForm();
			succes = 'Article cree.';
			await Promise.all([chargerArticles(), chargerStats()]);
		} catch (error) {
			erreur = error instanceof Error ? error.message : 'Erreur inconnue.';
		} finally {
			soumissionArticle = false;
		}
	}

	onMount(async () => {
		// Page reservee aux administrateurs.
		if ($auth.user?.role !== 'admin') {
			goto('/login');
			return;
		}
		await chargerCatalogue();
	});
</script>

<svelte:head>
	<title>Collector.shop | Admin Catalogue</title>
	<meta name="description" content="Administration du catalogue collector.shop." />
</svelte:head>

<div class="admin">
	<header class="head">
		<div>
			<div class="eyebrow">Administration</div>
			<h1 class="title">Catalogue collectors</h1>
			<p class="subtitle">Gestion du catalogue et des catégories.</p>
		</div>
		<div class="stats">
			<div class="stat">
				<span class="stat-val">{articles.length}</span>
				<span class="stat-label">articles</span>
			</div>
			<div class="stat">
				<span class="stat-val">{categories.length}</span>
				<span class="stat-label">catégories</span>
			</div>
		</div>
	</header>

	{#if erreur}
		<div class="msg msg-error">{erreur}</div>
	{/if}
	{#if succes}
		<div class="msg msg-success">{succes}</div>
	{/if}

	<!-- Tableau de bord back-office -->
	{#if stats}
		<!-- Rangée KPI business -->
		<section class="kpis">
			<div class="kpi kpi-hero">
				<span class="kpi-label">Volume d'affaires (GMV)</span>
				<span class="kpi-val">{eur(stats.gmv)}</span>
				<span class="kpi-sub">Panier moyen {eur(Math.round(stats.avgOrderValue))}</span>
			</div>
			<div class="kpi">
				<span class="kpi-label">Commandes</span>
				<span class="kpi-val">{stats.totalOrders}</span>
				<span class="kpi-sub">{stats.ordersByStatus.paid} à expédier</span>
			</div>
			<div class="kpi">
				<span class="kpi-label">Taux d'écoulement</span>
				<span class="kpi-val">{stats.sellThrough.toFixed(1)}%</span>
				<span class="kpi-sub">{stats.soldArticles}/{stats.totalArticles} vendues</span>
			</div>
			<div class="kpi">
				<span class="kpi-label">Annonces actives</span>
				<span class="kpi-val">{stats.activeListings}</span>
				<span class="kpi-sub">prix moyen {eur(Math.round(stats.avgListing))}</span>
			</div>
			<div class="kpi">
				<span class="kpi-label">Vendeurs actifs</span>
				<span class="kpi-val">{stats.activeSellers}</span>
				<span class="kpi-sub">{stats.categories} catégories</span>
			</div>
		</section>

		<div class="dash-grid">
			<!-- Entonnoir des commandes -->
			<section class="panel">
				<div class="eyebrow">Suivi</div>
				<h2 class="panel-title">Commandes par statut</h2>
				<div class="funnel">
					{#each ['paid', 'shipped', 'delivered', 'cancelled'] as st}
						{@const n = stats.ordersByStatus[st as keyof typeof stats.ordersByStatus]}
						<div class="funnel-row">
							<span class="funnel-name" class:funnel-cancel={st === 'cancelled'}>
								{ORDER_STATUS_FR[st]}
							</span>
							<span class="funnel-track">
								<span
									class="funnel-fill"
									class:funnel-fill-cancel={st === 'cancelled'}
									style={`width:${stats.totalOrders ? (n / stats.totalOrders) * 100 : 0}%`}
								></span>
							</span>
							<span class="funnel-count">{n}</span>
						</div>
					{/each}
				</div>

				<div class="eyebrow" style="margin-top:20px">Catalogue</div>
				<h2 class="panel-title">Répartition par catégorie</h2>
				<div class="funnel">
					{#each stats.byCategory as row}
						{#if row.name}
							<div class="funnel-row">
								<span class="funnel-name">{row.name}</span>
								<span class="funnel-track">
									<span
										class="funnel-fill"
										style={`width:${stats.totalArticles ? (row.count / stats.totalArticles) * 100 : 0}%`}
									></span>
								</span>
								<span class="funnel-count">{row.count}</span>
							</div>
						{/if}
					{/each}
				</div>
			</section>

			<!-- Activité récente -->
			<section class="panel">
				<div class="eyebrow">Activité</div>
				<h2 class="panel-title">Dernières commandes</h2>
				{#if stats.recentOrders.length === 0}
					<div class="empty">Aucune commande pour l'instant.</div>
				{:else}
					<div class="orders">
						{#each stats.recentOrders as o}
							<div class="order-row">
								<span class="order-date">{o.createdAt}</span>
								<span class="order-name">{o.article || `Lot #${o.id}`}</span>
								<span class="order-status order-{o.status}"
									>{ORDER_STATUS_FR[o.status] ?? o.status}</span
								>
								<span class="order-price">{eur(o.price)}</span>
							</div>
						{/each}
					</div>
				{/if}
			</section>
		</div>
	{/if}

	<div class="grid">
		<!-- Colonne gauche : catalogue existant -->
		<section class="panel">
			<div class="eyebrow">Catalogue</div>
			<h2 class="panel-title">Articles en vente</h2>

			<div class="admin-search">
				<span class="admin-search-ico" aria-hidden="true">⌕</span>
				<input
					class="admin-search-in"
					type="search"
					placeholder="Rechercher un article (titre, #ID ou catégorie)…"
					bind:value={rechercheArticle}
				/>
				{#if rechercheArticle}
					<span class="admin-search-count">{articlesAffiches.length}/{articles.length}</span>
				{/if}
			</div>

			{#if articlesAffiches.length > 0}
				<div class="bulk-bar">
					<label class="bulk-select-all">
						<input
							class="chk"
							type="checkbox"
							checked={toutSelectionne}
							onchange={toggleToutSelectionner}
						/>
						<span>Tout sélectionner</span>
					</label>
					{#if selectionArticles.size > 0}
						<button type="button" class="bulk-delete-btn" onclick={demanderSuppressionSelection}>
							Retirer la sélection ({selectionArticles.size})
						</button>
					{/if}
				</div>
			{/if}

			{#if chargement}
				<div class="empty">Chargement du catalogue...</div>
			{:else if articles.length === 0}
				<div class="empty">
					Aucun article pour le moment. Créez une catégorie puis un article depuis la colonne de
					droite.
				</div>
			{:else if articlesAffiches.length === 0}
				<div class="empty">Aucun article ne correspond à « {rechercheArticle} ».</div>
			{:else}
				<div class="cards">
					{#each articlesAffiches as article}
						<article class="card" class:card-selected={selectionArticles.has(article.ID)}>
							<div class="card-head">
								<div class="card-head-left">
									<label class="card-select">
										<input
											class="chk"
											type="checkbox"
											checked={selectionArticles.has(article.ID)}
											onchange={() => toggleSelectionArticle(article.ID)}
										/>
									</label>
									<div>
										<span class="card-id">Article #{article.ID}</span>
										<h3 class="card-name">{article.name}</h3>
									</div>
								</div>
								{#if article.category}
									<span class="tag">{article.category.name}</span>
								{/if}
							</div>
							<p class="card-desc">{article.description}</p>
							<div class="card-foot">
								<div>
									<span class="foot-label">Port</span>
									<span class="foot-val">{article.fraisPort} €</span>
								</div>
								<div class="foot-right">
									<span class="foot-label">Prix</span>
									<span class="foot-price">{article.prix} €</span>
								</div>
							</div>
							<button
								class="btn-delete"
								type="button"
								onclick={() => demanderSuppressionArticle(article.ID, article.name)}
							>
								Retirer du catalogue
							</button>
						</article>
					{/each}
				</div>
			{/if}
		</section>

		<!-- Colonne droite : catégories + formulaires -->
		<div class="side">
			<section class="panel">
				<div class="eyebrow">Taxonomie</div>
				<h2 class="panel-title">Catégories actuelles</h2>
				<div class="cats">
					{#if categories.length === 0}
						<p class="empty">Aucune catégorie créée.</p>
					{:else}
						{#each categories as category}
							<div class="cat">
								<p class="cat-name">{category.name}</p>
								<p class="cat-desc">{category.description}</p>
							</div>
						{/each}
					{/if}
				</div>
			</section>

			<section class="panel">
				<div class="eyebrow">Création</div>
				<h2 class="panel-title">Nouvelle catégorie</h2>
				<form
					class="form"
					onsubmit={(e) => {
						e.preventDefault();
						soumettreCategory();
					}}
				>
					<input
						class="input"
						type="text"
						bind:value={categoryForm.name}
						placeholder="Jeux vidéo"
						required
					/>
					<textarea
						class="input textarea"
						bind:value={categoryForm.description}
						placeholder="Consoles, cartouches, accessoires..."
						required></textarea>
					<button class="btn" type="submit" disabled={soumissionCategory}>
						{soumissionCategory ? 'Création...' : 'Créer la catégorie'}
					</button>
				</form>
			</section>

			<section class="panel">
				<div class="eyebrow">Création</div>
				<h2 class="panel-title">Nouvel article</h2>
				<form
					class="form"
					onsubmit={(e) => {
						e.preventDefault();
						soumettreArticle();
					}}
				>
					<input
						class="input"
						type="text"
						bind:value={articleForm.name}
						placeholder="Game Boy Color — Édition Pikachu"
						required
					/>
					<textarea
						class="input textarea"
						bind:value={articleForm.description}
						placeholder="Description du collector..."
						required></textarea>
					<div class="form-row">
						<input
							class="input"
							type="number"
							min="0"
							step="0.01"
							bind:value={articleForm.prix}
							placeholder="Prix"
							required
						/>
						<input
							class="input"
							type="number"
							min="0"
							step="0.01"
							bind:value={articleForm.fraisPort}
							placeholder="Port"
							required
						/>
					</div>
					<GSelect
						bind:value={articleForm.categoryId}
						ariaLabel="Catégorie"
						placeholder="Choisir une catégorie"
						options={categories.map((category) => ({
							value: String(category.ID),
							label: category.name
						}))}
					/>
					<button class="btn" type="submit" disabled={soumissionArticle}>
						{soumissionArticle ? 'Création...' : "Publier l'article"}
					</button>
				</form>
			</section>
		</div>
	</div>
</div>

<GConfirmModal
	bind:open={confirmOpen}
	title={confirmTitle}
	message={confirmMessage}
	checkboxLabel={confirmCheckboxLabel}
	bind:checkboxChecked={confirmCheckboxChecked}
	confirmLabel="Retirer"
	danger
	onConfirm={() => confirmAction?.()}
/>

<style>
	.admin {
		max-width: 1200px;
		margin: 0 auto;
		padding: 32px 38px 60px;
	}

	.eyebrow {
		font-size: 11px;
		letter-spacing: 0.2em;
		text-transform: uppercase;
		font-weight: 600;
		color: #1e3b2c;
		margin-bottom: 10px;
	}

	.head {
		display: flex;
		align-items: flex-end;
		justify-content: space-between;
		gap: 24px;
		flex-wrap: wrap;
		border-bottom: 1px solid rgba(43, 38, 32, 0.1);
		padding-bottom: 24px;
		margin-bottom: 24px;
	}
	.title {
		font-family: var(--f-serif);
		font-weight: 500;
		font-size: 40px;
		line-height: 1;
		color: #2b2620;
		margin: 0 0 8px;
	}
	.subtitle {
		font-size: 13px;
		color: #8a7a64;
		margin: 0;
	}
	.stats {
		display: flex;
		gap: 12px;
	}
	.stat {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 2px;
		min-width: 74px;
		padding: 12px 16px;
		border: 1px solid rgba(43, 38, 32, 0.1);
		border-radius: 9px;
		background: rgba(43, 38, 32, 0.02);
	}
	.stat-val {
		font-family: var(--f-body);
		font-size: 24px;
		color: #2b2620;
	}
	.stat-label {
		font-size: 10.5px;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		color: #8a7a64;
	}

	.msg {
		padding: 10px 14px;
		border-radius: 7px;
		border: 1px solid;
		font-size: 13px;
		margin-bottom: 16px;
	}
	.msg-error {
		border-color: rgba(176, 67, 42, 0.3);
		background: rgba(176, 67, 42, 0.06);
		color: #b0432a;
	}
	.msg-success {
		border-color: rgba(63, 122, 82, 0.3);
		background: rgba(63, 122, 82, 0.06);
		color: #3f7a52;
	}

	/* Dashboard back-office — KPI */
	.kpis {
		display: grid;
		grid-template-columns: 1.4fr 1fr 1fr 1fr 1fr;
		gap: 12px;
		margin-bottom: 16px;
	}
	.kpi {
		display: flex;
		flex-direction: column;
		gap: 5px;
		padding: 16px 18px;
		border: 1px solid rgba(43, 38, 32, 0.1);
		border-radius: 10px;
		background: rgba(43, 38, 32, 0.02);
	}
	.kpi-hero {
		background: rgba(30, 59, 44, 0.06);
		border-color: rgba(30, 59, 44, 0.24);
	}
	.kpi-label {
		font-size: 10.5px;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		color: #8a7a64;
	}
	.kpi-val {
		font-family: var(--f-body);
		font-size: 25px;
		color: #1e3b2c;
	}
	.kpi-hero .kpi-val {
		font-size: 30px;
		color: #2b2620;
	}
	.kpi-sub {
		font-size: 11.5px;
		color: #8a7a64;
	}
	@media (max-width: 900px) {
		.kpis {
			grid-template-columns: repeat(2, 1fr);
		}
	}

	.dash-grid {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 14px;
		margin-bottom: 22px;
		align-items: start;
	}
	@media (max-width: 900px) {
		.dash-grid {
			grid-template-columns: 1fr;
		}
	}

	.funnel {
		display: flex;
		flex-direction: column;
		gap: 9px;
		margin-top: 6px;
	}
	.funnel-row {
		display: flex;
		align-items: center;
		gap: 12px;
	}
	.funnel-name {
		width: 96px;
		font-size: 12.5px;
		color: #2b2620;
		flex-shrink: 0;
	}
	.funnel-cancel {
		color: #b0432a;
	}
	.funnel-track {
		flex: 1;
		height: 8px;
		border-radius: 4px;
		background: rgba(43, 38, 32, 0.05);
		overflow: hidden;
	}
	.funnel-fill {
		display: block;
		height: 100%;
		background: #1e3b2c;
		border-radius: 4px;
		min-width: 2px;
	}
	.funnel-fill-cancel {
		background: #b0432a;
	}
	.funnel-count {
		font-family: var(--f-body);
		font-size: 12px;
		color: #8a7a64;
		width: 28px;
		text-align: right;
		flex-shrink: 0;
	}

	.orders {
		display: flex;
		flex-direction: column;
		margin-top: 6px;
	}
	.order-row {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 10px 0;
		border-bottom: 1px solid rgba(43, 38, 32, 0.08);
	}
	.order-date {
		font-family: var(--f-body);
		font-size: 11px;
		color: #8a7a64;
		flex-shrink: 0;
		width: 82px;
	}
	.order-name {
		flex: 1;
		font-size: 13px;
		color: #2b2620;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.order-status {
		font-size: 10.5px;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		padding: 3px 8px;
		border-radius: 4px;
		flex-shrink: 0;
		background: rgba(30, 59, 44, 0.1);
		color: #1e3b2c;
	}
	.order-cancelled {
		background: rgba(176, 67, 42, 0.1);
		color: #b0432a;
	}
	.order-price {
		font-family: var(--f-body);
		font-size: 12.5px;
		color: #8a7a64;
		flex-shrink: 0;
		width: 84px;
		text-align: right;
	}

	/* Recherche articles */
	.admin-search {
		display: flex;
		align-items: center;
		gap: 8px;
		background: rgba(43, 38, 32, 0.04);
		border: 1px solid rgba(43, 38, 32, 0.12);
		border-radius: 8px;
		padding: 0 12px;
		margin-bottom: 14px;
		transition:
			border-color 150ms,
			box-shadow 150ms;
	}
	.admin-search:focus-within {
		border-color: rgba(30, 59, 44, 0.5);
		box-shadow: 0 0 0 3px rgba(30, 59, 44, 0.08);
	}
	.admin-search-ico {
		color: #8a7a64;
		font-size: 15px;
	}
	.admin-search-in {
		flex: 1;
		background: none;
		border: none;
		outline: none;
		padding: 10px 0;
		color: #2b2620;
		font-family: var(--f-body);
		font-size: 14px;
	}
	.admin-search-in::placeholder {
		color: rgba(43, 38, 32, 0.28);
	}
	.admin-search-count {
		font-family: var(--f-body);
		font-size: 11px;
		color: #8a7a64;
		flex-shrink: 0;
	}

	/* Sélection groupée */
	.bulk-bar {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
		margin-bottom: 14px;
	}
	.bulk-select-all {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 12.5px;
		color: #8a7a64;
		cursor: pointer;
	}
	.bulk-delete-btn {
		padding: 7px 14px;
		border-radius: 6px;
		border: 1px solid rgba(176, 67, 42, 0.28);
		background: rgba(176, 67, 42, 0.06);
		color: #b0432a;
		font-size: 12px;
		font-weight: 600;
		cursor: pointer;
		transition:
			background 150ms,
			border-color 150ms;
	}
	.bulk-delete-btn:hover {
		background: rgba(176, 67, 42, 0.14);
		border-color: rgba(176, 67, 42, 0.5);
	}
	.card-head-left {
		display: flex;
		align-items: flex-start;
		gap: 10px;
		min-width: 0;
	}
	.card-select {
		display: flex;
		align-items: flex-start;
		padding-top: 3px;
		flex-shrink: 0;
	}
	.card-selected {
		border-color: rgba(30, 59, 44, 0.4);
		background: rgba(30, 59, 44, 0.05);
	}

	/* Case a cocher themee (remplace l'apparence native du navigateur) */
	.chk {
		appearance: none;
		-webkit-appearance: none;
		width: 16px;
		height: 16px;
		margin: 0;
		border-radius: 4px;
		border: 1px solid rgba(43, 38, 32, 0.25);
		background: rgba(43, 38, 32, 0.03);
		cursor: pointer;
		position: relative;
		flex-shrink: 0;
		transition:
			background 120ms,
			border-color 120ms;
	}
	.chk:hover {
		border-color: rgba(43, 38, 32, 0.42);
	}
	.chk:checked {
		background: #1e3b2c;
		border-color: #1e3b2c;
	}
	.chk:checked::after {
		content: '';
		position: absolute;
		left: 5px;
		top: 2px;
		width: 4px;
		height: 8px;
		border: solid #f6f1e6;
		border-width: 0 2px 2px 0;
		transform: rotate(45deg);
	}
	.chk:focus-visible {
		outline: none;
		box-shadow: 0 0 0 3px rgba(30, 59, 44, 0.25);
	}

	.grid {
		display: grid;
		grid-template-columns: 1.4fr 0.9fr;
		gap: 18px;
		align-items: start;
	}
	@media (max-width: 900px) {
		.grid {
			grid-template-columns: 1fr;
		}
	}

	.panel {
		background: #fffdf8;
		border: 1px solid rgba(43, 38, 32, 0.1);
		border-radius: 9px;
		padding: 20px;
	}
	.side {
		display: flex;
		flex-direction: column;
		gap: 18px;
	}
	.panel-title {
		font-family: var(--f-serif);
		font-weight: 500;
		font-size: 22px;
		color: #2b2620;
		margin: 0 0 16px;
	}

	.empty {
		border: 1px dashed rgba(43, 38, 32, 0.14);
		border-radius: 8px;
		padding: 18px;
		font-size: 13px;
		color: #8a7a64;
	}

	.cards {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
		gap: 14px;
	}
	.card {
		border: 1px solid rgba(43, 38, 32, 0.1);
		border-radius: 8px;
		padding: 16px;
		background: rgba(43, 38, 32, 0.02);
	}
	.card-head {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 10px;
		margin-bottom: 12px;
	}
	.card-id {
		font-size: 10.5px;
		letter-spacing: 0.14em;
		text-transform: uppercase;
		color: #8a7a64;
	}
	.card-name {
		font-family: var(--f-serif);
		font-size: 19px;
		color: #2b2620;
		margin: 4px 0 0;
	}
	.tag {
		flex-shrink: 0;
		font-family: var(--f-body);
		font-size: 10px;
		color: #1e3b2c;
		border: 1px solid rgba(30, 59, 44, 0.28);
		background: rgba(30, 59, 44, 0.08);
		border-radius: 4px;
		padding: 3px 7px;
	}
	.card-desc {
		font-size: 13px;
		line-height: 1.5;
		color: #8a7a64;
		margin: 0 0 14px;
	}
	.card-foot {
		display: flex;
		align-items: flex-end;
		justify-content: space-between;
		border-top: 1px dashed rgba(43, 38, 32, 0.12);
		padding-top: 12px;
	}
	.btn-delete {
		margin-top: 12px;
		width: 100%;
		padding: 8px;
		border-radius: 6px;
		border: 1px solid rgba(176, 67, 42, 0.28);
		background: rgba(176, 67, 42, 0.06);
		color: #b0432a;
		font-size: 12px;
		font-weight: 600;
		cursor: pointer;
		transition:
			background 150ms,
			border-color 150ms;
	}
	.btn-delete:hover {
		background: rgba(176, 67, 42, 0.14);
		border-color: rgba(176, 67, 42, 0.5);
	}
	.foot-right {
		text-align: right;
	}
	.foot-label {
		display: block;
		font-size: 10px;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		color: #8a7a64;
	}
	.foot-val {
		font-size: 15px;
		color: #2b2620;
	}
	.foot-price {
		font-family: var(--f-body);
		font-size: 20px;
		color: #1e3b2c;
	}

	.cats {
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.cat {
		border: 1px solid rgba(43, 38, 32, 0.1);
		border-radius: 7px;
		padding: 10px 12px;
		background: rgba(43, 38, 32, 0.02);
	}
	.cat-name {
		font-size: 14px;
		font-weight: 600;
		color: #2b2620;
		margin: 0;
	}
	.cat-desc {
		font-size: 12px;
		color: #8a7a64;
		margin: 2px 0 0;
	}

	.form {
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.form-row {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 10px;
	}
	.input {
		width: 100%;
		background: rgba(43, 38, 32, 0.04);
		border: 1px solid rgba(43, 38, 32, 0.12);
		border-radius: 7px;
		padding: 10px 12px;
		color: #2b2620;
		font-family: var(--f-body);
		font-size: 14px;
		outline: none;
		box-sizing: border-box;
		transition:
			border-color 150ms,
			box-shadow 150ms;
	}
	.input::placeholder {
		color: rgba(43, 38, 32, 0.25);
	}
	.input:focus {
		border-color: rgba(30, 59, 44, 0.5);
		box-shadow: 0 0 0 3px rgba(30, 59, 44, 0.08);
	}
	.textarea {
		min-height: 84px;
		resize: vertical;
	}
	.btn {
		margin-top: 2px;
		padding: 12px;
		border-radius: 7px;
		border: none;
		background: #1e3b2c;
		color: #f6f1e6;
		font-size: 13px;
		font-weight: 700;
		letter-spacing: 0.04em;
		cursor: pointer;
		transition:
			filter 150ms,
			opacity 150ms;
	}
	.btn:hover:not(:disabled) {
		filter: brightness(1.08);
	}
	.btn:disabled {
		opacity: 0.55;
		cursor: not-allowed;
	}
</style>
