<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth, isAuthenticated } from '$lib/stores/auth';
	import { refreshStats } from '$lib/stores/stats';
	import { fetchCategories, type CategoryAPI } from '$lib/api/catalog';
	import { createListing, uploadArticleImage } from '$lib/api/market';
	import GPanel from '$lib/components/galerie/GPanel.svelte';
	import Kicker from '$lib/components/galerie/Kicker.svelte';

	let categories = $state<CategoryAPI[]>([]);
	let error = $state<string | null>(null);
	let busy = $state(false);

	// Champs du formulaire
	let name = $state('');
	let description = $state('');
	let series = $state('');
	let year = $state<number | null>(null);
	let rarity = $state('');
	let grade = $state('');
	let prix = $state<number | null>(null);
	let fraisPort = $state<number | null>(null);
	let categoryId = $state<number>(0);

	// Photo
	let photoFile = $state<File | null>(null);
	let photoPreview = $state<string | null>(null);

	const rarities = ['Common', 'Rare', 'Holo Rare', 'Near Mint', 'Limited', 'Sealed'];

	onMount(async () => {
		if (!$isAuthenticated) {
			goto('/login');
			return;
		}
		try {
			categories = await fetchCategories();
			if (categories.length > 0) categoryId = categories[0].ID;
		} catch (e) {
			error = 'Impossible de charger les catégories. Vérifiez que le catalog-service est démarré.';
			console.error(e);
		}
	});

	function onPhotoChange(e: Event) {
		const file = (e.currentTarget as HTMLInputElement).files?.[0] ?? null;
		photoFile = file;
		if (photoPreview) URL.revokeObjectURL(photoPreview);
		photoPreview = file ? URL.createObjectURL(file) : null;
	}

	const canSubmit = $derived(
		name.trim().length > 0 &&
			description.trim().length > 0 &&
			(prix ?? 0) > 0 &&
			categoryId > 0
	);

	async function submit(e: SubmitEvent) {
		e.preventDefault();
		if (!$auth.token || !canSubmit || busy) return;
		busy = true;
		error = null;
		try {
			const res = await createListing($auth.token, {
				name: name.trim(),
				description: description.trim(),
				series: series.trim(),
				year: year ?? 0,
				rarity,
				grade: grade.trim(),
				prix: prix ?? 0,
				fraisPort: fraisPort ?? 0,
				categoryId
			});
			if (photoFile) {
				try {
					await uploadArticleImage($auth.token, res.article.ID, photoFile);
				} catch (err) {
					console.error('Upload photo échoué, annonce créée sans photo', err);
				}
			}
			refreshStats();
			goto(`/lot/${res.article.ID}`);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Erreur lors de la création de l’annonce.';
			busy = false;
		}
	}
</script>

<svelte:head><title>Vendre une pièce · Collector.shop</title></svelte:head>

<div class="sell-head">
	<Kicker color="#86b3a4">Espace vendeur</Kicker>
	<h1 class="sell-title">Mettre une pièce en vente</h1>
	<p class="sell-lede">
		Votre annonce est publiée en vente directe sur la vitrine. Ajoutez une photo nette :
		les pièces avec photo se vendent mieux (et la quête photo vous attend). +60 XP à la publication.
	</p>
</div>

{#if error}
	<p class="sell-error">{error}</p>
{/if}

<form class="sell-grid" onsubmit={submit}>
	<!-- Photo -->
	<GPanel>
		<Kicker>Photo de la pièce</Kicker>
		<label class="photo-drop" class:has-photo={!!photoPreview}>
			{#if photoPreview}
				<img class="photo-preview" src={photoPreview} alt="Aperçu de la pièce" />
			{:else}
				<span class="photo-hint">+ Choisir une image<br /><small>jpg, png ou webp · 5 Mo max</small></span>
			{/if}
			<input type="file" accept=".jpg,.jpeg,.png,.webp" class="photo-input" onchange={onPhotoChange} />
		</label>
	</GPanel>

	<!-- Champs -->
	<GPanel>
		<Kicker>Détails de l'annonce</Kicker>
		<div class="fields">
			<label class="field">
				<span>Nom de la pièce *</span>
				<input bind:value={name} placeholder="ex. Charizard Base Set" required />
			</label>
			<label class="field">
				<span>Description *</span>
				<textarea bind:value={description} rows="3" placeholder="État, provenance, défauts éventuels…" required></textarea>
			</label>
			<div class="field-row">
				<label class="field">
					<span>Série / édition</span>
					<input bind:value={series} placeholder="ex. Base Set, 1ère édition" />
				</label>
				<label class="field">
					<span>Année</span>
					<input type="number" bind:value={year} min="1900" max="2030" placeholder="1999" />
				</label>
			</div>
			<div class="field-row">
				<label class="field">
					<span>Catégorie *</span>
					<select bind:value={categoryId}>
						{#each categories as cat}
							<option value={cat.ID}>{cat.name}</option>
						{/each}
					</select>
				</label>
				<label class="field">
					<span>Rareté</span>
					<select bind:value={rarity}>
						<option value="">—</option>
						{#each rarities as r}
							<option value={r}>{r}</option>
						{/each}
					</select>
				</label>
				<label class="field">
					<span>Grade</span>
					<input bind:value={grade} placeholder="ex. PSA 9, Mint…" />
				</label>
			</div>
			<div class="field-row">
				<label class="field">
					<span>Prix (€) *</span>
					<input type="number" bind:value={prix} min="1" step="0.01" placeholder="120" required />
				</label>
				<label class="field">
					<span>Frais de port (€)</span>
					<input type="number" bind:value={fraisPort} min="0" step="0.01" placeholder="8" />
				</label>
			</div>

			<button class="btn-primary" type="submit" disabled={!canSubmit || busy}>
				{busy ? 'Publication…' : 'Publier l’annonce (+60 XP)'}
			</button>
		</div>
	</GPanel>
</form>

<style>
	.sell-head { margin-bottom: 20px; }
	.sell-title {
		font-family: 'Newsreader', Georgia, serif;
		font-weight: 500;
		font-size: 36px;
		color: #ece5da;
		margin: 8px 0 6px;
	}
	.sell-lede {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 14px;
		color: #a39a8c;
		line-height: 1.6;
		max-width: 560px;
		margin: 0;
	}
	.sell-error {
		padding: 10px 14px;
		border-radius: 7px;
		border: 1px solid rgba(215, 156, 134, 0.3);
		background: rgba(215, 156, 134, 0.06);
		color: #d79c86;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13px;
		margin-bottom: 16px;
	}

	.sell-grid {
		display: grid;
		grid-template-columns: 0.8fr 1.2fr;
		gap: 18px;
		align-items: start;
	}
	@media (max-width: 768px) { .sell-grid { grid-template-columns: 1fr; } }

	/* Photo */
	.photo-drop {
		display: flex;
		align-items: center;
		justify-content: center;
		aspect-ratio: 4/3;
		margin-top: 12px;
		border: 1px dashed rgba(236, 229, 218, 0.22);
		border-radius: 9px;
		cursor: pointer;
		overflow: hidden;
		position: relative;
		background: rgba(255, 255, 255, 0.02);
		transition: border-color 120ms;
	}
	.photo-drop:hover { border-color: #86b3a4; }
	.photo-drop.has-photo { border-style: solid; }
	.photo-preview { width: 100%; height: 100%; object-fit: cover; }
	.photo-hint {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13px;
		color: #a39a8c;
		text-align: center;
		line-height: 1.8;
	}
	.photo-hint small { color: #766d60; font-size: 11px; }
	.photo-input { display: none; }

	/* Champs */
	.fields { display: flex; flex-direction: column; gap: 14px; margin-top: 12px; }
	.field-row { display: flex; gap: 12px; }
	.field-row .field { flex: 1; }
	.field { display: flex; flex-direction: column; gap: 5px; }
	.field span {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 10.5px;
		letter-spacing: 0.12em;
		text-transform: uppercase;
		color: #766d60;
	}
	.field input,
	.field textarea,
	.field select {
		background: rgba(255, 255, 255, 0.04);
		border: 1px solid rgba(236, 229, 218, 0.12);
		border-radius: 7px;
		padding: 10px 12px;
		color: #ece5da;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13px;
		outline: none;
	}
	.field input:focus,
	.field textarea:focus,
	.field select:focus { border-color: rgba(134, 179, 164, 0.5); }
	.field textarea { resize: vertical; }
	.field select option { background: #221f1b; }

	.btn-primary {
		margin-top: 4px;
		padding: 12px 22px;
		border-radius: 7px;
		border: none;
		background: #86b3a4;
		color: #191714;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13px;
		font-weight: 600;
		cursor: pointer;
		transition: filter 120ms, opacity 120ms;
	}
	.btn-primary:hover:not(:disabled) { filter: brightness(1.08); }
	.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
</style>
