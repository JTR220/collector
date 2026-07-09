<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { auth, isAuthenticated } from '$lib/stores/auth';
	import {
		fetchCategories,
		fetchArticle,
		createArticle,
		updateArticle,
		uploadArticleImage,
		type CategoryAPI,
		type NewArticleInput
	} from '$lib/api/catalog';
	import GPanel from '$lib/components/galerie/GPanel.svelte';
	import GSelect from '$lib/components/galerie/GSelect.svelte';
	import Kicker from '$lib/components/galerie/Kicker.svelte';

	const editId = $derived($page.url.searchParams.get('edit'));

	let categories = $state<CategoryAPI[]>([]);
	let loading = $state(true);
	let submitting = $state(false);
	let error = $state('');

	// Formulaire de mise en vente.
	let name = $state('');
	let categoryId = $state('');
	let prix = $state('');
	let fraisPort = $state('');
	let description = $state('');
	let series = $state('');
	let year = $state('');
	let rarity = $state('');
	let grade = $state('');
	let imageUrl = $state('');
	let photoFile = $state<File | null>(null);
	let photoError = $state('');

	const MAX_PHOTO_SIZE = 5 * 1024 * 1024;

	function onPhotoChange(e: Event) {
		const input = e.currentTarget as HTMLInputElement;
		const f = input.files?.[0] ?? null;
		photoError = '';
		if (f && f.size > MAX_PHOTO_SIZE) {
			photoError = 'Fichier trop volumineux (5 Mo max).';
			photoFile = null;
			input.value = '';
			return;
		}
		photoFile = f;
	}

	onMount(async () => {
		if (!$isAuthenticated || !$auth.token) {
			goto('/login');
			return;
		}
		try {
			categories = await fetchCategories();
			if (editId) {
				const a = await fetchArticle(editId);
				name = a.name;
				categoryId = String(a.categoryId);
				prix = String(a.prix);
				fraisPort = String(a.fraisPort);
				description = a.description;
				series = a.series ?? '';
				year = a.year ? String(a.year) : '';
				rarity = a.rarity ?? '';
				grade = a.grade ?? '';
				imageUrl = a.imageUrl ?? '';
			}
		} catch {
			error = editId
				? 'Impossible de charger cette annonce.'
				: 'Impossible de charger les catégories. Le catalog-service est-il démarré ?';
		} finally {
			loading = false;
		}
	});

	async function submit() {
		if (!$auth.token) return;
		if (!categoryId) {
			error = 'Choisissez une catégorie.';
			return;
		}
		submitting = true;
		error = '';
		try {
			if (editId) {
				await updateArticle($auth.token, Number(editId), {
					name: name.trim(),
					description: description.trim(),
					prix: Number(prix),
					fraisPort: Number(fraisPort),
					categoryId: Number(categoryId),
					imageUrl: imageUrl.trim() || undefined
				});
				if (photoFile) await uploadArticleImage($auth.token, Number(editId), photoFile);
				goto(`/lot/${editId}`);
				return;
			}
			const input: NewArticleInput = {
				name: name.trim(),
				description: description.trim(),
				prix: Number(prix),
				fraisPort: Number(fraisPort),
				categoryId: Number(categoryId),
				series: series.trim() || undefined,
				year: year ? Number(year) : undefined,
				rarity: rarity.trim() || undefined,
				grade: grade.trim() || undefined,
				imageUrl: imageUrl.trim() || undefined
			};
			const created = await createArticle($auth.token, input);
			if (photoFile) await uploadArticleImage($auth.token, created.ID, photoFile);
			goto(`/lot/${created.ID}`);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Opération impossible.';
			submitting = false;
		}
	}
</script>

<svelte:head><title>{editId ? 'Modifier' : 'Vendre'} · Collector.shop</title></svelte:head>

<section class="head">
	<Kicker>{editId ? "Modifier l'annonce" : 'Mettre en vente'}</Kicker>
	<h1 class="title">{editId ? 'Modifiez votre annonce' : 'Vendez votre pièce'}</h1>
	<p class="sub">
		{editId
			? 'Le titre, la description, le prix, les frais de port, la catégorie et la photo sont modifiables. Série, année, rareté et grade restent fixés à la création.'
			: 'Renseignez la fiche : elle sera publiée à votre nom sur le marché, en vente directe. Vous pourrez la retirer à tout moment depuis votre profil.'}
	</p>
</section>

{#if loading}
	<div class="state">Chargement…</div>
{:else}
	{#if error}
		<div class="msg-error">{error}</div>
	{/if}

	<form
		class="form-grid"
		onsubmit={(e) => {
			e.preventDefault();
			submit();
		}}
	>
		<GPanel>
			<Kicker>Essentiel</Kicker>
			<div class="fields">
				<label class="field span-2">
					<span class="lbl">Titre de l'annonce *</span>
					<input class="in" bind:value={name} placeholder="Charizard holo — Base Set" required />
				</label>

				<label class="field">
					<span class="lbl">Catégorie *</span>
					<GSelect
						bind:value={categoryId}
						ariaLabel="Catégorie"
						placeholder="Choisir…"
						options={categories.map((c) => ({ value: String(c.ID), label: c.name }))}
					/>
				</label>

				<label class="field">
					<span class="lbl">Prix (€) *</span>
					<input
						class="in"
						type="number"
						min="0"
						step="0.01"
						bind:value={prix}
						placeholder="1290"
						required
					/>
				</label>

				<label class="field">
					<span class="lbl">Frais de port (€) *</span>
					<input
						class="in"
						type="number"
						min="0"
						step="0.01"
						bind:value={fraisPort}
						placeholder="12"
						required
					/>
				</label>

				<label class="field span-2">
					<span class="lbl">Description *</span>
					<textarea
						class="in area"
						bind:value={description}
						placeholder="État, provenance, particularités, authentification…"
						required></textarea>
				</label>
			</div>
		</GPanel>

		<GPanel>
			<Kicker
				>{editId
					? 'Détails (série/année/rareté/grade fixés à la création)'
					: 'Détails (optionnel)'}</Kicker
			>
			<div class="fields">
				<label class="field">
					<span class="lbl">Série / édition</span>
					<input
						class="in"
						bind:value={series}
						placeholder="Base Set, 1re édition"
						disabled={!!editId}
					/>
				</label>
				<label class="field">
					<span class="lbl">Année</span>
					<input
						class="in"
						type="number"
						bind:value={year}
						placeholder="1999"
						disabled={!!editId}
					/>
				</label>
				<label class="field">
					<span class="lbl">Rareté</span>
					<input class="in" bind:value={rarity} placeholder="Holo Rare" disabled={!!editId} />
				</label>
				<label class="field">
					<span class="lbl">Grade</span>
					<input class="in" bind:value={grade} placeholder="PSA 9" disabled={!!editId} />
				</label>
				<label class="field span-2">
					<span class="lbl">Photo (fichier, 5 Mo max)</span>
					<input
						class="in in-file"
						type="file"
						accept="image/jpeg,image/png,image/webp,image/gif"
						onchange={onPhotoChange}
					/>
					{#if photoFile}<span class="photo-picked">{photoFile.name} sélectionné</span>{/if}
					{#if photoError}<span class="photo-err">{photoError}</span>{/if}
				</label>
				<label class="field span-2">
					<span class="lbl">…ou URL de la photo (https uniquement)</span>
					<input
						class="in"
						type="url"
						bind:value={imageUrl}
						disabled={!!photoFile}
						placeholder={editId
							? 'https://…  (laissez vide pour garder la photo actuelle)'
							: 'https://…  (laissez vide pour une photo par défaut)'}
					/>
				</label>
			</div>

			<div class="actions">
				<a class="btn-ghost" href={editId ? `/lot/${editId}` : '/'}>Annuler</a>
				<button class="btn" type="submit" disabled={submitting}>
					{submitting ? 'Enregistrement…' : editId ? 'Enregistrer' : 'Mettre en vente'}
				</button>
			</div>
		</GPanel>
	</form>
{/if}

<style>
	.head {
		padding: 20px 0 18px;
	}
	.title {
		font-family: var(--f-serif);
		font-weight: 600;
		font-size: clamp(26px, 4vw, 36px);
		color: var(--c-text);
		margin: 8px 0 10px;
	}
	.sub {
		font-family: var(--f-body);
		font-size: 14px;
		color: var(--c-text-muted);
		line-height: 1.55;
		max-width: 560px;
		margin: 0;
	}
	.state {
		text-align: center;
		padding: 60px 0;
		font-family: var(--f-serif);
		font-style: italic;
		font-size: 15px;
		color: var(--c-text-muted);
	}
	.msg-error {
		padding: 11px 15px;
		border-radius: 7px;
		border: 1px solid rgba(176, 67, 42, 0.3);
		background: #fbe9e3;
		color: var(--c-error);
		font-family: var(--f-body);
		font-size: 13px;
		margin-bottom: 16px;
	}
	.form-grid {
		display: grid;
		grid-template-columns: 1.3fr 1fr;
		gap: 14px;
		align-items: start;
	}
	@media (max-width: 820px) {
		.form-grid {
			grid-template-columns: 1fr;
		}
	}
	.fields {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 12px;
		margin-top: 8px;
	}
	.field {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	.span-2 {
		grid-column: 1 / -1;
	}
	.lbl {
		font-family: var(--f-body);
		font-size: 11.5px;
		letter-spacing: 0.04em;
		color: var(--c-text-muted);
	}
	.in {
		width: 100%;
		box-sizing: border-box;
		background: var(--c-bg);
		border: 1px solid var(--c-border);
		border-radius: 7px;
		padding: 10px 12px;
		color: var(--c-text);
		font-family: var(--f-body);
		font-size: 14px;
		outline: none;
		transition: border-color 150ms;
	}
	.in::placeholder {
		color: var(--c-text-muted);
	}
	.in:focus {
		border-color: var(--c-ink);
	}
	.in:disabled {
		opacity: 0.45;
		cursor: not-allowed;
	}
	.in-file {
		padding: 8px 10px;
		font-size: 13px;
	}
	.photo-picked {
		font-family: var(--f-body);
		font-size: 12px;
		color: var(--c-ink);
	}
	.photo-err {
		font-family: var(--f-body);
		font-size: 12px;
		color: var(--c-error);
	}
	.area {
		min-height: 92px;
		resize: vertical;
	}
	.actions {
		display: flex;
		justify-content: flex-end;
		gap: 10px;
		margin-top: 16px;
	}
	.btn-ghost {
		display: inline-flex;
		align-items: center;
		font-family: var(--f-body);
		font-size: 13px;
		padding: 11px 20px;
		border-radius: 8px;
		border: 1px solid var(--c-border);
		color: var(--c-text-tertiary);
		text-decoration: none;
	}
	.btn-ghost:hover {
		color: var(--c-ink);
		border-color: var(--c-ink);
	}
	.btn {
		padding: 11px 24px;
		border-radius: 8px;
		border: none;
		background: var(--c-accent);
		color: #fff;
		font-family: var(--f-body);
		font-size: 13px;
		font-weight: 700;
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
