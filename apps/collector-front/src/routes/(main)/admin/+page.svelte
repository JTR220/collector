<script lang="ts">
	import { env } from '$env/dynamic/public';
	import { onMount } from 'svelte';
	import type { Article, Category } from '$lib/types/catalog';

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

	let categories: Category[] = [];
	let articles: Article[] = [];

	let categoryForm = emptyCategoryForm();
	let articleForm = emptyArticleForm();

	let chargement = true;
	let soumissionCategory = false;
	let soumissionArticle = false;
	let erreur = '';
	let succes = '';

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

	async function chargerArticles() {
		const response = await fetch(`${catalogApiBaseUrl}/article`);
		if (!response.ok) {
			throw new Error('Impossible de charger les articles.');
		}
		const payload = (await response.json()) as Record<string, unknown>[];
		articles = payload.map(normalizeArticle);
	}

	async function chargerCatalogue() {
		chargement = true;
		erreur = '';

		try {
			await Promise.all([chargerCategories(), chargerArticles()]);
		} catch (error) {
			erreur = error instanceof Error ? error.message : 'Erreur inconnue.';
		} finally {
			chargement = false;
		}
	}

	async function soumettreCategory() {
		soumissionCategory = true;
		erreur = '';
		succes = '';

		try {
			const response = await fetch(`${catalogApiBaseUrl}/category`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
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
		soumissionArticle = true;
		erreur = '';
		succes = '';

		try {
			const response = await fetch(`${catalogApiBaseUrl}/article`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
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
			await chargerArticles();
		} catch (error) {
			erreur = error instanceof Error ? error.message : 'Erreur inconnue.';
		} finally {
			soumissionArticle = false;
		}
	}

	onMount(async () => {
		await chargerCatalogue();
	});
</script>

<svelte:head>
	<title>Collector.shop | Admin Catalogue</title>
	<meta name="description" content="Administration du catalogue collector.shop." />
</svelte:head>

<div
	class="min-h-screen bg-[radial-gradient(circle_at_top_left,_rgba(255,238,201,0.95),_rgba(255,255,255,1)_40%),linear-gradient(135deg,#0f172a_0%,#1f2937_45%,#f97316_45%,#fff7ed_100%)] text-slate-950"
>
	<div class="mx-auto flex min-h-screen max-w-7xl flex-col gap-8 px-4 py-8 md:px-8">
		<section
			class="rounded-[2rem] border border-white/70 bg-white/80 p-6 shadow-[0_30px_80px_rgba(15,23,42,0.18)] backdrop-blur md:p-8"
		>
			<div
				class="mb-8 flex flex-col gap-5 border-b border-slate-200 pb-6 lg:flex-row lg:items-end lg:justify-between"
			>
				<div class="max-w-3xl">
					<p
						class="mb-3 inline-flex rounded-full border border-orange-300 bg-orange-100 px-3 py-1 text-xs font-semibold tracking-[0.25em] text-orange-700 uppercase"
					>
						Collector.shop
					</p>
					<h1 class="text-4xl font-black tracking-tight text-slate-900 uppercase md:text-5xl">
						Catalogue collectors
					</h1>
					<p class="mt-3 text-sm leading-6 text-slate-600 md:text-base">
						Le front consomme le service catalogue sur
						<span class="font-semibold text-slate-900">{catalogApiBaseUrl}</span>.
					</p>
				</div>

				<div class="grid grid-cols-2 gap-3 text-center">
					<div class="rounded-2xl bg-slate-950 px-4 py-3 text-white">
						<p class="text-2xl font-black">{articles.length}</p>
						<p class="text-xs tracking-[0.2em] text-slate-300 uppercase">articles</p>
					</div>
					<div class="rounded-2xl bg-orange-500 px-4 py-3 text-white">
						<p class="text-2xl font-black">{categories.length}</p>
						<p class="text-xs tracking-[0.2em] text-orange-100 uppercase">categories</p>
					</div>
				</div>
			</div>

			{#if erreur}
				<div
					class="mb-4 rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm font-medium text-red-700"
				>
					{erreur}
				</div>
			{/if}

			{#if succes}
				<div
					class="mb-4 rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm font-medium text-emerald-700"
				>
					{succes}
				</div>
			{/if}

			<div class="grid gap-4 lg:grid-cols-[1.3fr_0.7fr]">
				<section
					class="rounded-3xl border border-slate-200 bg-white p-5 shadow-[0_18px_50px_rgba(15,23,42,0.08)]"
				>
					<div class="mb-5">
						<p class="text-xs font-semibold tracking-[0.25em] text-slate-400 uppercase">
							Catalogue
						</p>
						<h2 class="mt-2 text-2xl font-black text-slate-900">Articles en vitrine</h2>
					</div>

					{#if chargement}
						<div
							class="rounded-3xl border border-dashed border-slate-300 bg-slate-50 p-6 text-sm text-slate-500"
						>
							Chargement du catalogue...
						</div>
					{:else if articles.length === 0}
						<div
							class="rounded-3xl border border-dashed border-slate-300 bg-slate-50 p-6 text-sm text-slate-500"
						>
							Aucun article pour le moment. Cree une categorie puis un article depuis la colonne de
							droite.
						</div>
					{:else}
						<div class="grid gap-4 md:grid-cols-2">
							{#each articles as article}
								<article class="rounded-3xl border border-slate-200 bg-slate-50 p-5">
									<div class="mb-4 flex items-start justify-between gap-4">
										<div>
											<p class="text-xs font-semibold tracking-[0.25em] text-slate-400 uppercase">
												Article #{article.ID}
											</p>
											<h3 class="mt-2 text-2xl font-black text-slate-900">{article.name}</h3>
										</div>
										{#if article.category}
											<span
												class="rounded-full bg-orange-100 px-3 py-1 text-xs font-bold tracking-[0.2em] text-orange-700 uppercase"
											>
												{article.category.name}
											</span>
										{/if}
									</div>

									<p class="mb-4 text-sm leading-6 text-slate-600">{article.description}</p>

									<div
										class="flex items-end justify-between border-t border-dashed border-slate-300 pt-4"
									>
										<div>
											<p class="text-xs tracking-[0.2em] text-slate-400 uppercase">Port</p>
											<p class="text-lg font-bold text-slate-700">{article.fraisPort} EUR</p>
										</div>
										<div class="text-right">
											<p class="text-xs tracking-[0.2em] text-slate-400 uppercase">Prix</p>
											<p class="text-2xl font-black text-orange-600">{article.prix} EUR</p>
										</div>
									</div>
								</article>
							{/each}
						</div>
					{/if}
				</section>

				<section
					class="rounded-3xl border border-slate-200 bg-slate-950 p-5 text-white shadow-[0_18px_50px_rgba(15,23,42,0.18)]"
				>
					<p class="text-xs font-semibold tracking-[0.25em] text-slate-400 uppercase">Categories</p>
					<h2 class="mt-2 text-2xl font-black">Taxonomie actuelle</h2>

					<div class="mt-5 flex flex-wrap gap-3">
						{#if categories.length === 0}
							<p class="text-sm text-slate-400">Aucune categorie creee.</p>
						{:else}
							{#each categories as category}
								<div class="rounded-2xl border border-slate-800 bg-slate-900 px-4 py-3">
									<p class="text-sm font-bold text-white">{category.name}</p>
									<p class="mt-1 text-xs text-slate-400">{category.description}</p>
								</div>
							{/each}
						{/if}
					</div>
				</section>
			</div>
		</section>

		<section
			class="rounded-[2rem] border border-slate-900/10 bg-slate-950 p-6 text-white shadow-[0_30px_80px_rgba(15,23,42,0.28)] md:p-8"
		>
			<div class="mb-8">
				<p class="text-xs font-semibold tracking-[0.3em] text-orange-300 uppercase">Catalogue</p>
				<h2 class="mt-3 text-3xl font-black tracking-tight uppercase">
					Creer categories et articles
				</h2>
				<p class="mt-3 max-w-md text-sm leading-6 text-slate-300">
					L'interface est maintenant centree sur le vrai domaine produit : categories et articles de
					collection.
				</p>
			</div>

			<div class="grid gap-8 xl:grid-cols-2">
				<form class="space-y-4" on:submit|preventDefault={soumettreCategory}>
					<p class="text-sm font-semibold tracking-[0.2em] text-slate-400 uppercase">
						Nouvelle categorie
					</p>
					<input
						class="w-full rounded-2xl border border-slate-700 bg-slate-900 px-4 py-3 text-white placeholder:text-slate-500 focus:border-orange-400 focus:ring-0"
						type="text"
						bind:value={categoryForm.name}
						placeholder="Jeux video"
						required
					/>
					<textarea
						class="min-h-24 w-full rounded-2xl border border-slate-700 bg-slate-900 px-4 py-3 text-white placeholder:text-slate-500 focus:border-orange-400 focus:ring-0"
						bind:value={categoryForm.description}
						placeholder="Consoles, cartouches, accessoires..."
						required
					></textarea>
					<button
						class="w-full rounded-2xl bg-orange-500 px-5 py-3 text-sm font-black tracking-[0.2em] text-white uppercase transition hover:bg-orange-400 disabled:cursor-not-allowed disabled:bg-orange-300"
						type="submit"
						disabled={soumissionCategory}
					>
						{soumissionCategory ? 'Creation...' : 'Creer la categorie'}
					</button>
				</form>

				<form
					class="space-y-4 border-t border-slate-800 pt-8 xl:border-t-0 xl:border-l xl:pt-0 xl:pl-8"
					on:submit|preventDefault={soumettreArticle}
				>
					<p class="text-sm font-semibold tracking-[0.2em] text-slate-400 uppercase">
						Nouvel article
					</p>
					<input
						class="w-full rounded-2xl border border-slate-700 bg-slate-900 px-4 py-3 text-white placeholder:text-slate-500 focus:border-orange-400 focus:ring-0"
						type="text"
						bind:value={articleForm.name}
						placeholder="Game Boy Color - Edition Pikachu"
						required
					/>
					<textarea
						class="min-h-28 w-full rounded-2xl border border-slate-700 bg-slate-900 px-4 py-3 text-white placeholder:text-slate-500 focus:border-orange-400 focus:ring-0"
						bind:value={articleForm.description}
						placeholder="Description du collector..."
						required
					></textarea>
					<div class="grid gap-4 sm:grid-cols-2">
						<input
							class="w-full rounded-2xl border border-slate-700 bg-slate-900 px-4 py-3 text-white placeholder:text-slate-500 focus:border-orange-400 focus:ring-0"
							type="number"
							min="0"
							step="0.01"
							bind:value={articleForm.prix}
							placeholder="Prix"
							required
						/>
						<input
							class="w-full rounded-2xl border border-slate-700 bg-slate-900 px-4 py-3 text-white placeholder:text-slate-500 focus:border-orange-400 focus:ring-0"
							type="number"
							min="0"
							step="0.01"
							bind:value={articleForm.fraisPort}
							placeholder="Port"
							required
						/>
					</div>
					<select
						class="w-full rounded-2xl border border-slate-700 bg-slate-900 px-4 py-3 text-white focus:border-orange-400 focus:ring-0"
						bind:value={articleForm.categoryId}
						required
					>
						<option value="" disabled>Choisir une categorie</option>
						{#each categories as category}
							<option value={String(category.ID)}>{category.name}</option>
						{/each}
					</select>
					<button
						class="w-full rounded-2xl bg-emerald-500 px-5 py-3 text-sm font-black tracking-[0.2em] text-white uppercase transition hover:bg-emerald-400 disabled:cursor-not-allowed disabled:bg-emerald-300"
						type="submit"
						disabled={soumissionArticle}
					>
						{soumissionArticle ? 'Creation...' : "Publier l'article"}
					</button>
				</form>
			</div>
		</section>
	</div>
</div>
