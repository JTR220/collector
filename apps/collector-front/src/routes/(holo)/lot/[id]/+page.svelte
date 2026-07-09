<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { auth, isAuthenticated } from '$lib/stores/auth';
	import { fetchArticle, fetchArticles, articleImage, articleImages, type ArticleAPI } from '$lib/api/catalog';
	import { recentlyViewed } from '$lib/stores/recentlyViewed';
	import { addToWishlist, removeFromWishlist, fetchMyWishlist } from '$lib/api/wishlist';
	import {
		buyArticle,
		fetchSellerRating,
		fetchSellerReviews,
		type SellerRating,
		type Review
	} from '$lib/api/market';
	import { fetchPriceHistory } from '$lib/api/priceTracker';
	import { sendMessage, toUserUUID } from '$lib/api/messages';
	import { cart } from '$lib/stores/cart';
	import { eur, eurC, pct, sparkPath } from '$lib/utils/format';
	import GPanel from '$lib/components/galerie/GPanel.svelte';
	import Kicker from '$lib/components/galerie/Kicker.svelte';

	let article = $state<ArticleAPI | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let inWishlist = $state(false);
	let actionMsg = $state<string | null>(null);
	let actionBusy = $state(false);
	let contactOpen = $state(false);
	let contactDraft = $state('');
	let contactBusy = $state(false);
	let negotiateOpen = $state(false);
	let negotiatePrice = $state('');
	let negotiateComment = $state('');
	let negotiateBusy = $state(false);

	let trackedPrices = $state<number[]>([]);
	let sellerRating = $state<SellerRating | null>(null);
	let sellerReviews = $state<Review[]>([]);
	let related = $state<ArticleAPI[]>([]);
	let mainImageIndex = $state(0);

	// Suggestions "déjà vus", excluant l'article courant.
	const recentSuggestions = $derived(
		$recentlyViewed.filter((a) => a.ID !== article?.ID).slice(0, 4)
	);

	onMount(async () => {
		try {
			article = await fetchArticle($page.params.id ?? '');
			if ($auth.token) {
				const wishlist = await fetchMyWishlist($auth.token);
				inWishlist = wishlist.some((w) => w.articleId === article?.ID);
			}
			if (article) recentlyViewed.push(article);
		} catch (e) {
			error = 'Lot introuvable ou catalog-service indisponible.';
			console.error(e);
		} finally {
			loading = false;
		}

		if (article) {
			try {
				const hist = await fetchPriceHistory(article.ID);
				trackedPrices = hist.map((h) => h.new_price);
			} catch {
				trackedPrices = [];
			}

			try {
				[sellerRating, sellerReviews] = await Promise.all([
					fetchSellerRating(article.sellerId),
					fetchSellerReviews(article.sellerId)
				]);
			} catch {
				sellerRating = null;
				sellerReviews = [];
			}

			try {
				const all = await fetchArticles();
				related = all
					.filter((a) => a.ID !== article!.ID && a.category.name === article!.category.name)
					.slice(0, 4);
			} catch {
				related = [];
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

	const isOwnArticle = $derived(!!article && !!$auth.user && article.sellerId === $auth.user.id);

	function addToCart() {
		if (!article) return;
		cart.add(article);
	}

	async function buyNow() {
		const token = requireAuth();
		if (!token || !article || isOwnArticle) return;
		actionBusy = true;
		actionMsg = null;
		try {
			await buyArticle(token, article.ID);
			article.sold = true;
			actionMsg =
				'Demande d’achat envoyée — le vendeur doit la valider. Vous recevrez une notification.';
		} catch (e) {
			actionMsg = e instanceof Error ? e.message : 'Erreur lors de l’achat.';
		} finally {
			actionBusy = false;
		}
	}

	async function sendContactMessage() {
		const token = requireAuth();
		const body = contactDraft.trim();
		if (!token || !article || !body || isOwnArticle) return;
		contactBusy = true;
		try {
			const sent = await sendMessage(token, {
				recipientId: toUserUUID(article.sellerId),
				body,
				articleId: article.ID,
				articleName: article.name
			});
			goto(`/messages/${sent.conversation_id}`);
		} catch (e) {
			actionMsg = e instanceof Error ? e.message : "Erreur lors de l'envoi du message.";
		} finally {
			contactBusy = false;
		}
	}

	function openNegotiate() {
		if (!article) return;
		negotiatePrice = String(article.prix);
		negotiateOpen = !negotiateOpen;
	}

	async function sendNegotiation() {
		const token = requireAuth();
		if (!token || !article || isOwnArticle) return;
		const offer = Number(negotiatePrice);
		if (!offer || offer <= 0) {
			actionMsg = 'Indiquez un prix proposé valide.';
			return;
		}
		negotiateBusy = true;
		try {
			let body = `Bonjour, seriez-vous prêt à accepter ${eur(offer)} pour "${article.name}" (prix affiché : ${eur(article.prix)}) ?`;
			if (negotiateComment.trim()) body += `\n${negotiateComment.trim()}`;
			const sent = await sendMessage(token, {
				recipientId: toUserUUID(article.sellerId),
				body,
				articleId: article.ID,
				articleName: article.name
			});
			goto(`/messages/${sent.conversation_id}`);
		} catch (e) {
			actionMsg = e instanceof Error ? e.message : "Erreur lors de l'envoi de l'offre.";
		} finally {
			negotiateBusy = false;
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
	{@const gallery = articleImages(article)}
	{@const mainImg = gallery[mainImageIndex] ?? gallery[0]}
	<a class="back-link" href="/">Accueil / {article.category.name} / <span>{article.name}</span></a>

	<section class="lot-grid">
		<!-- Galerie photo -->
		<div class="gallery">
			<div class="gallery-main">
				{#if mainImg}
					<img
						class="gallery-img"
						src={mainImg}
						alt={article.name}
						onerror={(e) => ((e.currentTarget as HTMLImageElement).style.display = 'none')}
					/>
				{:else}
					<span class="gallery-glyph">{article.glyph}</span>
				{/if}
				{#if article.sold}
					<span class="lot-sold">vendu</span>
				{/if}
			</div>
			{#if gallery.length > 1}
				<div class="gallery-thumbs">
					{#each gallery as thumb, i}
						<button
							class="gallery-thumb"
							class:gallery-thumb-active={i === mainImageIndex}
							onclick={() => (mainImageIndex = i)}
							aria-label={`Photo ${i + 1}`}
						>
							<img src={thumb} alt="" />
						</button>
					{/each}
				</div>
			{/if}
		</div>

		<!-- Infos -->
		<div class="lot-info">
			<div class="lot-badges">
				{#if article.grade}<span class="badge-condition">{article.grade}</span>{/if}
				<span class="badge-verified">✓ Vérifié par Collector</span>
				{#if article.rarity}<span class="badge-rarity">{article.rarity}</span>{/if}
			</div>

			<h1 class="lot-title">{article.name}</h1>
			{#if article.series}<p class="lot-series">{article.series}</p>{/if}
			<div class="lot-price">{eur(article.prix)}</div>

			<div class="seller-card">
				<div class="seller-avatar">{article.seller.slice(0, 2).toUpperCase()}</div>
				<div class="seller-meta">
					<span class="seller-name">{article.seller}</span>
					<span class="seller-rating">
						Particulier vérifié
						{#if sellerRating && sellerRating.count > 0}
							· {sellerRating.average.toFixed(1)} ★ ({sellerRating.count} ventes)
						{/if}
					</span>
				</div>
				{#if !isOwnArticle}
					<button class="btn-outline" onclick={() => (contactOpen = !contactOpen)}>
						Contacter
					</button>
				{/if}
			</div>

			<div class="cta-row">
				{#if article.sold}
					<button class="btn-primary" disabled>Vendu</button>
				{:else if isOwnArticle}
					<button class="btn-primary" disabled title="Vous ne pouvez pas acheter votre propre annonce">
						Votre annonce
					</button>
				{:else}
					{@const currentId = article.ID}
					{@const inCart = $cart.some((i) => i.ID === currentId)}
					<button class="btn-primary" disabled={inCart} onclick={addToCart}>
						{inCart ? '✓ Dans le panier' : 'Ajouter au panier'}
					</button>
					<button class="btn-primary-alt" disabled={actionBusy} onclick={buyNow}>
						Acheter maintenant
					</button>
				{/if}
				<button
					class="btn-square"
					disabled={actionBusy}
					onclick={toggleWishlist}
					title={inWishlist ? 'Retirer de la wishlist' : 'Ajouter à la wishlist'}
					aria-label="Wishlist"
				>
					{inWishlist ? '♥' : '♡'}
				</button>
			</div>

			{#if !isOwnArticle && !article.sold}
				<button class="btn-outline btn-negotiate" onclick={openNegotiate}>
					💬 Négocier le prix
				</button>
			{/if}

			{#if contactOpen && !isOwnArticle}
				<div class="contact-box">
					<textarea
						placeholder={`Bonjour, votre annonce "${article.name}" m'intéresse…`}
						bind:value={contactDraft}
						disabled={contactBusy}
						rows="2"></textarea>
					<button
						class="btn-send"
						disabled={contactBusy || !contactDraft.trim()}
						onclick={sendContactMessage}
					>
						Envoyer
					</button>
				</div>
			{/if}

			{#if negotiateOpen && !isOwnArticle && !article.sold}
				<div class="negotiate-box">
					<label class="negotiate-field">
						<span class="negotiate-lbl">Votre offre (€)</span>
						<input
							class="negotiate-in"
							type="number"
							min="1"
							step="0.01"
							bind:value={negotiatePrice}
							disabled={negotiateBusy}
						/>
					</label>
					<textarea
						class="negotiate-comment"
						placeholder="Message facultatif…"
						bind:value={negotiateComment}
						disabled={negotiateBusy}
						rows="2"></textarea>
					<button class="btn-send" disabled={negotiateBusy || !negotiatePrice} onclick={sendNegotiation}>
						Envoyer l'offre
					</button>
				</div>
			{/if}

			{#if actionMsg}
				<p class="action-msg">{actionMsg}</p>
			{/if}

			<p class="lot-desc">{article.description}</p>

			<div class="trust-row">
				<span>Livraison assurée sous 3 à 5 jours</span>
				<span class="dot">·</span>
				<span>Paiement 100% sécurisé</span>
				<span class="dot">·</span>
				<span>Frais de port {eurC(article.fraisPort)}</span>
			</div>

			{#if sellerReviews.length > 0}
				<div class="review-list">
					{#each sellerReviews.slice(0, 5) as r (r.ID)}
						<div class="review-row">
							<span class="review-stars">{'★'.repeat(r.rating)}{'☆'.repeat(5 - r.rating)}</span>
							<span class="review-author">{r.reviewerName}</span>
							{#if r.comment}<span class="review-comment">{r.comment}</span>{/if}
						</div>
					{/each}
				</div>
			{/if}
		</div>
	</section>

	<!-- Historique de prix -->
	<GPanel style="margin-top:24px">
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
					stroke="#e4dcc8"
					stroke-width="0.5"
				/>
			{/each}
			<path d={historyPath} stroke={up ? '#3f7a52' : '#b0432a'} stroke-width="2" fill="none" />
		</svg>
		<div class="history-vals">
			{#each historySeries as p}
				<span>{eur(p)}</span>
			{/each}
		</div>
	</GPanel>

	<!-- Déjà consultés -->
	{#if recentSuggestions.length > 0}
		<section class="related">
			<h2 class="related-title">Récemment consultés</h2>
			<div class="related-grid">
				{@render productCards(recentSuggestions)}
			</div>
		</section>
	{/if}

	<!-- Vous aimerez aussi -->
	{#if related.length > 0}
		<section class="related">
			<h2 class="related-title">Vous aimerez aussi</h2>
			<div class="related-grid">
				{@render productCards(related)}
			</div>
		</section>
	{/if}
{/if}

{#snippet productCards(items: ArticleAPI[])}
	{#each items as r (r.ID)}
		{@const rImg = articleImage(r)}
		<a class="related-card" href={`/lot/${r.ID}`}>
			<div class="related-art">
				{#if rImg}<img src={rImg} alt={r.name} />{/if}
			</div>
			<div class="related-body">
				<p class="related-name">{r.name}</p>
				<p class="related-price">{eur(r.prix)}</p>
			</div>
		</a>
	{/each}
{/snippet}

<style>
	.state-msg {
		text-align: center;
		padding: 60px 0;
		font-family: var(--f-serif);
		font-style: italic;
		font-size: 15px;
		color: var(--c-text-muted);
	}
	.state-msg.error {
		color: var(--c-error);
	}

	.back-link {
		display: inline-block;
		font-family: var(--f-body);
		font-size: 13px;
		color: var(--c-text-muted);
		text-decoration: none;
		margin: 20px 0 16px;
	}
	.back-link span {
		color: var(--c-text);
	}
	.back-link:hover {
		color: var(--c-ink);
	}

	.lot-grid {
		display: grid;
		grid-template-columns: 1.1fr 0.9fr;
		gap: 56px;
		align-items: start;
	}
	@media (max-width: 900px) {
		.lot-grid {
			grid-template-columns: 1fr;
			gap: 24px;
		}
	}

	/* Galerie */
	.gallery {
		display: flex;
		flex-direction: column;
		gap: 14px;
	}
	.gallery-main {
		position: relative;
		height: 460px;
		border-radius: 16px;
		background: radial-gradient(120% 90% at 30% 20%, #2a4e3a 0%, #1e3b2c 55%, #16301f 100%);
		display: flex;
		align-items: center;
		justify-content: center;
		overflow: hidden;
	}
	@media (max-width: 640px) {
		.gallery-main {
			height: 320px;
		}
	}
	.gallery-glyph {
		font-family: var(--f-serif);
		font-size: 110px;
		color: rgba(246, 241, 230, 0.85);
	}
	.gallery-img {
		position: absolute;
		inset: 0;
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	.lot-sold {
		position: absolute;
		top: 14px;
		left: 14px;
		font-family: var(--f-body);
		font-size: 10px;
		font-weight: 600;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		padding: 4px 10px;
		border-radius: 5px;
		color: #fff;
		background: var(--c-accent);
	}
	.gallery-thumbs {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 12px;
	}
	.gallery-thumb {
		height: 90px;
		padding: 0;
		border-radius: 10px;
		background: var(--c-bg);
		border: 1px solid var(--c-border);
		overflow: hidden;
		cursor: pointer;
		transition: border-color 120ms;
	}
	.gallery-thumb:hover {
		border-color: var(--c-ink);
	}
	.gallery-thumb-active {
		border: 2px solid var(--c-ink);
	}
	.gallery-thumb img {
		width: 100%;
		height: 100%;
		object-fit: cover;
		display: block;
	}

	/* Infos */
	.lot-info {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}
	.lot-badges {
		display: flex;
		gap: 8px;
		flex-wrap: wrap;
	}
	.badge-condition,
	.badge-verified,
	.badge-rarity {
		font-family: var(--f-body);
		font-size: 11px;
		font-weight: 600;
		padding: 4px 10px;
		border-radius: 6px;
	}
	.badge-condition {
		color: var(--c-ink);
		background: var(--c-badge-verified-bg);
	}
	.badge-verified {
		color: var(--c-ink);
		background: var(--c-badge-moderation-bg);
	}
	.badge-rarity {
		color: var(--c-text-tertiary);
		background: var(--c-bg);
		border: 1px solid var(--c-border);
	}
	.lot-title {
		font-family: var(--f-serif);
		font-weight: 600;
		font-size: 28px;
		line-height: 1.25;
		color: var(--c-text);
		margin: 0;
	}
	.lot-series {
		font-family: var(--f-body);
		font-size: 13px;
		color: var(--c-text-muted);
		margin: -8px 0 0;
	}
	.lot-price {
		font-family: var(--f-serif);
		font-size: 32px;
		font-weight: 700;
		color: var(--c-ink);
	}

	.seller-card {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 16px;
		border: 1px solid var(--c-border);
		border-radius: 12px;
		background: var(--c-surface);
	}
	.seller-avatar {
		width: 44px;
		height: 44px;
		border-radius: 50%;
		background: var(--c-ink);
		color: var(--c-bg);
		font-family: var(--f-body);
		font-size: 14px;
		font-weight: 600;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}
	.seller-meta {
		flex: 1;
		display: flex;
		flex-direction: column;
		gap: 2px;
		min-width: 0;
	}
	.seller-name {
		font-family: var(--f-body);
		font-size: 14px;
		font-weight: 600;
		color: var(--c-text);
	}
	.seller-rating {
		font-family: var(--f-body);
		font-size: 12px;
		color: var(--c-text-muted);
	}
	.btn-outline {
		flex-shrink: 0;
		padding: 9px 16px;
		border: 1px solid var(--c-ink);
		border-radius: 8px;
		background: transparent;
		color: var(--c-ink);
		font-family: var(--f-body);
		font-size: 13px;
		font-weight: 600;
		cursor: pointer;
		transition: background 120ms;
	}
	.btn-outline:hover {
		background: var(--c-badge-verified-bg);
	}

	.cta-row {
		display: flex;
		gap: 12px;
	}
	.btn-primary {
		flex: 1;
		padding: 15px;
		border: none;
		border-radius: 10px;
		background: var(--c-accent);
		color: #fff;
		font-family: var(--f-body);
		font-size: 15px;
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
	.btn-primary-alt {
		flex: 1;
		padding: 15px;
		border: 1px solid var(--c-ink);
		border-radius: 10px;
		background: transparent;
		color: var(--c-ink);
		font-family: var(--f-body);
		font-size: 15px;
		font-weight: 600;
		cursor: pointer;
		transition: background 120ms;
	}
	.btn-primary-alt:hover:not(:disabled) {
		background: var(--c-badge-verified-bg);
	}
	.btn-primary-alt:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}
	.btn-square {
		width: 52px;
		height: 52px;
		flex-shrink: 0;
		border: 1px solid var(--c-border);
		border-radius: 10px;
		background: transparent;
		font-size: 20px;
		color: var(--c-accent);
		cursor: pointer;
		transition: border-color 120ms;
	}
	.btn-square:hover:not(:disabled) {
		border-color: var(--c-ink);
	}

	.btn-negotiate {
		align-self: flex-start;
	}

	.btn-send {
		align-self: flex-start;
		padding: 10px 22px;
		border: none;
		border-radius: 8px;
		background: var(--c-accent);
		color: #fff;
		font-family: var(--f-body);
		font-size: 13px;
		font-weight: 600;
		cursor: pointer;
		transition: filter 120ms;
	}
	.btn-send:hover:not(:disabled) {
		filter: brightness(1.08);
	}
	.btn-send:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.contact-box {
		display: flex;
		gap: 8px;
		align-items: flex-start;
	}
	.contact-box textarea {
		flex: 1;
		resize: vertical;
		background: var(--c-surface);
		border: 1px solid var(--c-border);
		border-radius: 8px;
		padding: 9px 12px;
		color: var(--c-text);
		font-family: var(--f-body);
		font-size: 13px;
	}
	.contact-box textarea:focus {
		outline: none;
		border-color: var(--c-ink);
	}

	.negotiate-box {
		display: flex;
		flex-direction: column;
		gap: 8px;
		padding: 14px;
		border: 1px solid var(--c-border);
		border-radius: 10px;
		background: var(--c-surface);
	}
	.negotiate-field {
		display: flex;
		flex-direction: column;
		gap: 4px;
		max-width: 180px;
	}
	.negotiate-lbl {
		font-family: var(--f-body);
		font-size: 11px;
		color: var(--c-text-muted);
	}
	.negotiate-in {
		background: var(--c-bg);
		border: 1px solid var(--c-border);
		border-radius: 7px;
		padding: 8px 10px;
		color: var(--c-text);
		font-family: var(--f-body);
		font-size: 13px;
	}
	.negotiate-in:focus {
		outline: none;
		border-color: var(--c-ink);
	}
	.negotiate-comment {
		resize: vertical;
		background: var(--c-bg);
		border: 1px solid var(--c-border);
		border-radius: 8px;
		padding: 9px 12px;
		color: var(--c-text);
		font-family: var(--f-body);
		font-size: 13px;
	}
	.negotiate-comment:focus {
		outline: none;
		border-color: var(--c-ink);
	}

	.action-msg {
		font-family: var(--f-body);
		font-size: 12.5px;
		color: var(--c-ink);
		font-weight: 600;
	}

	.lot-desc {
		font-family: var(--f-body);
		font-size: 14px;
		line-height: 1.6;
		color: var(--c-text-tertiary);
		border-top: 1px solid var(--c-border);
		padding-top: 16px;
		margin: 0;
	}

	.trust-row {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
		font-family: var(--f-body);
		font-size: 12.5px;
		color: var(--c-text-muted);
	}
	.trust-row .dot {
		color: var(--c-border);
	}

	.review-list {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.review-row {
		display: flex;
		flex-wrap: wrap;
		align-items: baseline;
		gap: 8px;
		padding: 10px 12px;
		border: 1px solid var(--c-border);
		border-radius: 8px;
		background: var(--c-surface);
	}
	.review-stars {
		color: #c1552f;
		font-size: 12px;
		letter-spacing: 1px;
	}
	.review-author {
		font-family: var(--f-body);
		font-size: 12px;
		font-weight: 600;
		color: var(--c-text);
	}
	.review-comment {
		flex-basis: 100%;
		font-family: var(--f-body);
		font-size: 12.5px;
		color: var(--c-text-muted);
	}

	.history-vals {
		display: flex;
		justify-content: space-between;
		font-family: var(--f-body);
		font-size: 10px;
		color: var(--c-text-muted);
		margin-top: 6px;
	}

	/* Vous aimerez aussi */
	.related {
		margin: 40px 0 24px;
	}
	.related-title {
		font-family: var(--f-serif);
		font-size: 19px;
		font-weight: 600;
		color: var(--c-ink);
		margin: 0 0 20px;
	}
	.related-grid {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 24px;
	}
	@media (max-width: 780px) {
		.related-grid {
			grid-template-columns: repeat(2, 1fr);
			gap: 14px;
		}
	}
	.related-card {
		display: block;
		background: var(--c-surface);
		border: 1px solid var(--c-border);
		border-radius: var(--r-card);
		overflow: hidden;
		text-decoration: none;
	}
	.related-art {
		height: 150px;
		background: var(--c-bg);
	}
	.related-art img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	.related-body {
		padding: 14px;
	}
	.related-name {
		font-family: var(--f-body);
		font-size: 13px;
		font-weight: 600;
		color: var(--c-text);
		line-height: 1.3;
		margin: 0 0 6px;
	}
	.related-price {
		font-family: var(--f-serif);
		font-size: 16px;
		font-weight: 600;
		color: var(--c-ink);
		margin: 0;
	}
</style>
