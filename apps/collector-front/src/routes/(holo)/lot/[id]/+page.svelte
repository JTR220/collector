<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { auth, isAuthenticated } from '$lib/stores/auth';
	import { refreshStats } from '$lib/stores/stats';
	import { fetchArticle, articleImage, type ArticleAPI } from '$lib/api/catalog';
	import {
		createDropEntry,
		addToWishlist,
		removeFromWishlist,
		fetchMyWishlist,
		createJournalEntry,
		type DropEntryKind
	} from '$lib/api/engagement';
	import { buyArticle, uploadArticleImage, deleteListing } from '$lib/api/market';
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

	// Formulaire d'avis
	let rating = $state(0);
	let note = $state('');
	let reviewSent = $state(false);

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
	});

	function requireAuth(): string | null {
		if (!$isAuthenticated || !$auth.token) {
			goto('/login');
			return null;
		}
		return $auth.token;
	}

	async function enterDrop(kind: DropEntryKind) {
		const token = requireAuth();
		if (!token || !article) return;
		actionBusy = true;
		actionMsg = null;
		try {
			const res = await createDropEntry(token, article.ID, kind);
			if (res.already) {
				actionMsg = 'Vous êtes déjà inscrit à ce drop.';
			} else {
				if (res.seatsLeft != null) article.seatsLeft = res.seatsLeft;
				if (res.dropStatus) article.dropStatus = res.dropStatus as ArticleAPI['dropStatus'];
				actionMsg = `Inscription confirmée · +${res.xp ?? 0} XP`;
				refreshStats();
			}
		} catch (e) {
			actionMsg = e instanceof Error ? e.message : 'Erreur lors de l’inscription.';
		} finally {
			actionBusy = false;
		}
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
				actionMsg = res.already ? 'Déjà dans la wishlist.' : 'Ajouté à la wishlist · +30 XP';
				refreshStats();
			}
		} catch (e) {
			actionMsg = e instanceof Error ? e.message : 'Erreur wishlist.';
		} finally {
			actionBusy = false;
		}
	}

	// ── Marketplace : achat direct, photo, retrait d'annonce ──
	const isDirect = $derived(article?.saleType === 'direct');
	const isMine = $derived(
		!!article && article.sellerId !== 0 && article.sellerId === $auth.user?.id
	);

	async function buyNow() {
		const token = requireAuth();
		if (!token || !article) return;
		actionBusy = true;
		actionMsg = null;
		try {
			const res = await buyArticle(token, article.ID);
			article.sold = true;
			article.dropStatus = 'sold';
			actionMsg = `Achat confirmé · +${res.xp} XP — suivez la commande dans votre marché.`;
			refreshStats();
		} catch (e) {
			actionMsg = e instanceof Error ? e.message : 'Erreur lors de l’achat.';
		} finally {
			actionBusy = false;
		}
	}

	let photoInput = $state<HTMLInputElement | null>(null);

	async function sendPhoto() {
		const token = requireAuth();
		const file = photoInput?.files?.[0];
		if (!token || !article || !file) return;
		actionBusy = true;
		actionMsg = null;
		try {
			const res = await uploadArticleImage(token, article.ID, file);
			article.imageUrl = res.imageUrl;
			actionMsg = 'Photo mise à jour.';
			refreshStats();
		} catch (e) {
			actionMsg = e instanceof Error ? e.message : 'Erreur lors de l’upload.';
		} finally {
			actionBusy = false;
			if (photoInput) photoInput.value = '';
		}
	}

	async function removeListing() {
		const token = requireAuth();
		if (!token || !article) return;
		actionBusy = true;
		try {
			await deleteListing(token, article.ID);
			goto('/marche');
		} catch (e) {
			actionMsg = e instanceof Error ? e.message : 'Erreur lors du retrait.';
			actionBusy = false;
		}
	}

	async function sendReview() {
		const token = requireAuth();
		if (!token || !article || rating === 0) return;
		actionBusy = true;
		try {
			await createJournalEntry(token, { articleId: article.ID, kind: 'noté', rating, note });
			reviewSent = true;
			actionMsg = 'Avis publié dans votre journal · +30 XP';
			refreshStats();
		} catch (e) {
			actionMsg = e instanceof Error ? e.message : 'Erreur lors de la publication.';
		} finally {
			actionBusy = false;
		}
	}

	const ctaByStatus: Record<string, { label: string; kind: DropEntryKind }> = {
		live: { label: 'Acheter ce lot', kind: 'purchase' },
		next: { label: 'Entrer dans le raffle', kind: 'raffle' },
		soon: { label: '+ Rappel au lancement', kind: 'reminder' },
		sold: { label: 'Rejoindre la liste d’attente', kind: 'waitlist' }
	};

	const W = 560;
	const H = 140;
	const historyPath = $derived(article ? sparkPath(article.priceHistory ?? [], W, H) : '');
	const up = $derived((article?.delta ?? 0) >= 0);
</script>

<svelte:head><title>{article ? article.name : 'Lot'} · Collector.shop</title></svelte:head>

{#if loading}
	<p class="state-msg">Chargement du lot…</p>
{:else if error || !article}
	<p class="state-msg error">{error}</p>
	<a class="back-link" href="/">← Retour à la vitrine</a>
{:else}
	{@const cta = ctaByStatus[article.dropStatus] ?? ctaByStatus.soon}
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
				<GChip color={article.dropStatus === 'live' ? '#86b3a4' : '#a39a8c'}>{article.dropId}</GChip
				>
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

			{#if article.seatsTotal > 0 && article.dropStatus !== 'sold'}
				<p class="lot-seats">
					{article.seatsLeft} / {article.seatsTotal} places restantes · drop du {article.dropDate}
				</p>
			{/if}

			<div class="lot-actions">
				{#if isDirect}
					{#if article.sold}
						<button class="btn-primary" disabled>Vendu</button>
					{:else if isMine}
						<button class="btn-ghost" disabled>Votre annonce</button>
					{:else}
						<button class="btn-primary" disabled={actionBusy} onclick={buyNow}>
							Acheter maintenant · {eur(article.prix)}
						</button>
					{/if}
				{:else}
					<button class="btn-primary" disabled={actionBusy} onclick={() => enterDrop(cta.kind)}>
						{cta.label}
					</button>
				{/if}
				<button class="btn-ghost" disabled={actionBusy} onclick={toggleWishlist}>
					{inWishlist ? '♥ Dans la wishlist' : '♡ Wishlist'}
				</button>
			</div>

			{#if isMine}
				<div class="owner-panel">
					<Kicker>Gérer votre annonce</Kicker>
					<div class="owner-actions">
						<input
							bind:this={photoInput}
							type="file"
							accept=".jpg,.jpeg,.png,.webp"
							class="owner-file"
							onchange={sendPhoto}
						/>
						<button class="btn-ghost" disabled={actionBusy} onclick={() => photoInput?.click()}>
							{article.imageUrl ? 'Changer la photo' : '+ Ajouter une photo'}
						</button>
						{#if !article.sold}
							<button class="btn-danger" disabled={actionBusy} onclick={removeListing}>
								Retirer de la vente
							</button>
						{/if}
					</div>
				</div>
			{/if}

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
			{#each article.priceHistory ?? [] as p}
				<span>{eur(p)}</span>
			{/each}
		</div>
	</GPanel>

	<!-- Avis -->
	<GPanel style="margin-top:14px">
		<Kicker>Noter cette pièce</Kicker>
		{#if reviewSent}
			<p class="review-done">
				Merci, votre avis est publié dans votre <a href="/journal">journal</a>.
			</p>
		{:else}
			<div class="review-stars">
				{#each [1, 2, 3, 4, 5] as star}
					<button
						class="star-btn"
						class:star-on={rating >= star}
						onclick={() => (rating = star)}
						aria-label="{star} étoiles"
					>
						{rating >= star ? '★' : '☆'}
					</button>
				{/each}
			</div>
			<textarea
				class="review-input"
				bind:value={note}
				rows="3"
				placeholder="État, emballage, communication du vendeur…"
			></textarea>
			<button class="btn-primary" disabled={rating === 0 || actionBusy} onclick={sendReview}
				>Publier l'avis (+30 XP)</button
			>
		{/if}
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

	.lot-seats {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 11.5px;
		color: #a39a8c;
		margin-bottom: 14px;
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

	.btn-danger {
		padding: 12px 18px;
		border-radius: 7px;
		border: 1px solid rgba(215, 156, 134, 0.4);
		background: transparent;
		color: #d79c86;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13px;
		cursor: pointer;
		transition:
			border-color 120ms,
			background 120ms;
	}
	.btn-danger:hover:not(:disabled) {
		border-color: #d79c86;
		background: rgba(215, 156, 134, 0.08);
	}

	.owner-panel {
		margin-top: 18px;
	}
	.owner-actions {
		display: flex;
		gap: 10px;
		margin-top: 10px;
		flex-wrap: wrap;
	}
	.owner-file {
		display: none;
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

	.review-stars {
		display: flex;
		gap: 4px;
		margin: 12px 0;
	}
	.star-btn {
		background: none;
		border: none;
		font-size: 26px;
		color: #766d60;
		cursor: pointer;
		padding: 0 2px;
		transition: color 120ms;
	}
	.star-btn.star-on {
		color: #86b3a4;
	}
	.review-input {
		width: 100%;
		box-sizing: border-box;
		background: rgba(255, 255, 255, 0.04);
		border: 1px solid rgba(236, 229, 218, 0.12);
		border-radius: 7px;
		padding: 11px 14px;
		color: #ece5da;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13px;
		outline: none;
		resize: vertical;
		margin-bottom: 12px;
	}
	.review-input:focus {
		border-color: rgba(134, 179, 164, 0.5);
	}
	.review-done {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13px;
		color: #86c099;
		margin-top: 10px;
	}
	.review-done a {
		color: #86b3a4;
	}
</style>
