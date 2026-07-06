<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth, isAuthenticated } from '$lib/stores/auth';
	import {
		fetchCategories,
		createArticle,
		type CategoryAPI,
		type NewArticleInput
	} from '$lib/api/catalog';
	import GPanel from '$lib/components/galerie/GPanel.svelte';
	import GSelect from '$lib/components/galerie/GSelect.svelte';
	import Kicker from '$lib/components/galerie/Kicker.svelte';

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

	onMount(async () => {
		if (!$isAuthenticated || !$auth.token) {
			goto('/login');
			return;
		}
		try {
			categories = await fetchCategories();
		} catch {
			error = 'Impossible de charger les catégories. Le catalog-service est-il démarré ?';
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
			goto(`/lot/${created.ID}`);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Mise en vente impossible.';
			submitting = false;
		}
	}
</script>

<svelte:head><title>Vendre · Collector.shop</title></svelte:head>

<section class="head">
	<Kicker>Mettre en vente</Kicker>
	<h1 class="title">Vendez votre pièce</h1>
	<p class="sub">
		Renseignez la fiche : elle sera publiée à votre nom sur le marché, en vente directe. Vous
		pourrez la retirer à tout moment depuis votre profil.
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
					<input class="in" type="number" min="0" step="0.01" bind:value={prix} placeholder="1290" required />
				</label>

				<label class="field">
					<span class="lbl">Frais de port (€) *</span>
					<input class="in" type="number" min="0" step="0.01" bind:value={fraisPort} placeholder="12" required />
				</label>

				<label class="field span-2">
					<span class="lbl">Description *</span>
					<textarea
						class="in area"
						bind:value={description}
						placeholder="État, provenance, particularités, authentification…"
						required
					></textarea>
				</label>
			</div>
		</GPanel>

		<GPanel>
			<Kicker>Détails (optionnel)</Kicker>
			<div class="fields">
				<label class="field">
					<span class="lbl">Série / édition</span>
					<input class="in" bind:value={series} placeholder="Base Set, 1re édition" />
				</label>
				<label class="field">
					<span class="lbl">Année</span>
					<input class="in" type="number" bind:value={year} placeholder="1999" />
				</label>
				<label class="field">
					<span class="lbl">Rareté</span>
					<input class="in" bind:value={rarity} placeholder="Holo Rare" />
				</label>
				<label class="field">
					<span class="lbl">Grade</span>
					<input class="in" bind:value={grade} placeholder="PSA 9" />
				</label>
				<label class="field span-2">
					<span class="lbl">URL de la photo (https uniquement)</span>
					<input
						class="in"
						type="url"
						bind:value={imageUrl}
						placeholder="https://…  (laissez vide pour une photo par défaut)"
					/>
				</label>
			</div>

			<div class="actions">
				<a class="btn-ghost" href="/">Annuler</a>
				<button class="btn" type="submit" disabled={submitting}>
					{submitting ? 'Publication…' : 'Mettre en vente'}
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
		font-family: 'Newsreader', Georgia, serif;
		font-weight: 500;
		font-size: clamp(28px, 4vw, 40px);
		color: #ece5da;
		margin: 8px 0 10px;
	}
	.sub {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 14px;
		color: #a39a8c;
		line-height: 1.55;
		max-width: 560px;
		margin: 0;
	}
	.state {
		text-align: center;
		padding: 60px 0;
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 12px;
		color: #a39a8c;
		letter-spacing: 0.16em;
	}
	.msg-error {
		padding: 11px 15px;
		border-radius: 7px;
		border: 1px solid rgba(215, 156, 134, 0.3);
		background: rgba(215, 156, 134, 0.06);
		color: #d79c86;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
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
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 11.5px;
		letter-spacing: 0.04em;
		color: #a39a8c;
	}
	.in {
		width: 100%;
		box-sizing: border-box;
		background: rgba(255, 255, 255, 0.04);
		border: 1px solid rgba(236, 229, 218, 0.12);
		border-radius: 7px;
		padding: 10px 12px;
		color: #ece5da;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 14px;
		outline: none;
		transition:
			border-color 150ms,
			box-shadow 150ms;
	}
	.in::placeholder {
		color: rgba(236, 229, 218, 0.25);
	}
	.in:focus {
		border-color: rgba(134, 179, 164, 0.5);
		box-shadow: 0 0 0 3px rgba(134, 179, 164, 0.08);
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
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13px;
		padding: 11px 20px;
		border-radius: 8px;
		border: 1px solid rgba(236, 229, 218, 0.12);
		color: #a39a8c;
		text-decoration: none;
	}
	.btn-ghost:hover {
		color: #ece5da;
		border-color: rgba(236, 229, 218, 0.24);
	}
	.btn {
		padding: 11px 24px;
		border-radius: 8px;
		border: none;
		background: #86b3a4;
		color: #191714;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
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
