<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth, isAuthenticated } from '$lib/stores/auth';
	import { refreshStats } from '$lib/stores/stats';
	import { fetchArticles, type ArticleAPI } from '$lib/api/catalog';
	import { createDropEntry, fetchMyDropEntries, type DropEntryKind } from '$lib/api/engagement';
	import { eur } from '$lib/utils/format';
	import GPanel from '$lib/components/galerie/GPanel.svelte';
	import GChip from '$lib/components/galerie/GChip.svelte';
	import GMeter from '$lib/components/galerie/GMeter.svelte';
	import Kicker from '$lib/components/galerie/Kicker.svelte';

	let articles = $state<ArticleAPI[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let entered = $state<Record<string, boolean>>({});
	let busy = $state(false);
	let toast = $state<string | null>(null);

	onMount(async () => {
		try {
			articles = await fetchArticles();
			if ($auth.token) {
				const entries = await fetchMyDropEntries($auth.token);
				const map: Record<string, boolean> = {};
				for (const e of entries) map[`${e.articleId}:${e.kind}`] = true;
				entered = map;
			}
		} catch (e) {
			error = 'Impossible de charger les drops.';
			console.error(e);
		} finally {
			loading = false;
		}
	});

	const kindByStatus: Record<string, DropEntryKind> = {
		live: 'purchase',
		next: 'raffle',
		soon: 'reminder',
		sold: 'waitlist'
	};
	const enteredLabel: Record<DropEntryKind, string> = {
		purchase: '✓ Acheté',
		raffle: '✓ Inscrit au raffle',
		reminder: '✓ Rappel actif',
		waitlist: '✓ En liste d’attente'
	};

	function isEntered(article: ArticleAPI, kind: DropEntryKind) {
		return !!entered[`${article.ID}:${kind}`];
	}

	async function act(article: ArticleAPI, kind: DropEntryKind) {
		if (!$isAuthenticated || !$auth.token) {
			goto('/login');
			return;
		}
		if (isEntered(article, kind) || busy) return;
		busy = true;
		toast = null;
		try {
			const res = await createDropEntry($auth.token, article.ID, kind);
			entered = { ...entered, [`${article.ID}:${kind}`]: true };
			if (res.seatsLeft != null) article.seatsLeft = res.seatsLeft;
			if (res.dropStatus) article.dropStatus = res.dropStatus as ArticleAPI['dropStatus'];
			toast = res.already
				? `${article.name} : déjà inscrit.`
				: `${article.name} : ${enteredLabel[kind].replace('✓ ', '').toLowerCase()} · +${res.xp ?? 0} XP`;
			refreshStats();
		} catch (e) {
			toast = e instanceof Error ? e.message : 'Erreur lors de l’inscription.';
		} finally {
			busy = false;
		}
	}

	const featured = $derived(
		articles.find((a) => a.dropStatus === 'next') ??
			articles.find((a) => a.dropStatus === 'live') ??
			null
	);

	const dropAt = (() => {
		const d = new Date();
		d.setDate(d.getDate() + ((5 - d.getDay() + 7) % 7 || 7));
		d.setHours(18, 0, 0, 0);
		return d;
	})();
	let remaining = $state(0);
	let timerId: ReturnType<typeof setInterval>;
	function tick() {
		remaining = Math.max(0, dropAt.getTime() - Date.now());
	}
	onMount(() => {
		tick();
		timerId = setInterval(tick, 1000);
	});
	onDestroy(() => clearInterval(timerId));
	const days = $derived(Math.floor(remaining / 86400_000));
	const hours = $derived(Math.floor((remaining % 86400_000) / 3600_000));
	const mins = $derived(Math.floor((remaining % 3600_000) / 60_000));
	const secs = $derived(Math.floor((remaining % 60_000) / 1000));
	const pad = (n: number) => String(n).padStart(2, '0');

	type DropStatus = 'live' | 'next' | 'sold' | 'soon';
	const statusLabel: Record<DropStatus, string> = {
		live: 'Live',
		next: 'Prochain',
		sold: 'Épuisé',
		soon: 'Bientôt'
	};
	const statusColor: Record<DropStatus, string> = {
		live: '#86b3a4',
		next: '#a39a8c',
		sold: '#766d60',
		soon: '#a39a8c'
	};
</script>

<svelte:head><title>Drops · Collector.shop</title></svelte:head>

{#if loading}
	<p class="state-msg">Chargement des drops…</p>
{:else if error}
	<p class="state-msg error">{error}</p>
{:else if featured}
	<div class="hero-grid">
		<div>
			<Kicker color="#86b3a4"
				>{featured.dropId} · {statusLabel[featured.dropStatus as DropStatus]} ouvert</Kicker
			>
			<h1 class="hero-title">{featured.name}</h1>
			<p class="hero-sub">{featured.series} · {featured.grade} · Livraison sous 48 h</p>

			<div class="countdown">
				{#each [{ val: days, label: 'Jours' }, { val: hours, label: 'Heures' }, { val: mins, label: 'Min' }, { val: secs, label: 'Sec' }] as seg}
					<div class="cd-box">
						<span class="cd-val">{pad(seg.val)}</span>
						<span class="cd-label">{seg.label}</span>
					</div>
				{/each}
			</div>

			<div class="hero-actions">
				<button
					class="btn-primary"
					disabled={busy || isEntered(featured, kindByStatus[featured.dropStatus])}
					onclick={() => act(featured, kindByStatus[featured.dropStatus])}
				>
					{isEntered(featured, kindByStatus[featured.dropStatus])
						? enteredLabel[kindByStatus[featured.dropStatus]]
						: featured.dropStatus === 'live'
							? 'Acheter maintenant'
							: 'Entrer dans le raffle'}
				</button>
				<button
					class="btn-ghost"
					disabled={busy || isEntered(featured, 'reminder')}
					onclick={() => act(featured, 'reminder')}
				>
					{isEntered(featured, 'reminder') ? '✓ Rappel actif' : '+ Rappel'}
				</button>
			</div>
			<p class="hero-meta">
				{featured.seatsTotal} places · {eur(featured.prix)} drop · {eur(featured.resellPrice)} resell
			</p>
		</div>

		<div class="hero-art">
			<span class="hero-glyph">{featured.glyph}</span>
			<div class="hero-badges">
				<GChip>{featured.grade}</GChip>
				<GChip>{featured.rarity}</GChip>
			</div>
		</div>
	</div>
{/if}

{#if !loading && !error}
	{#if toast}
		<div class="toast">{toast}</div>
	{/if}
	<div class="section-head">
		<Kicker>Calendrier des drops</Kicker>
	</div>
	<div class="calendar-grid">
		{#each articles as article (article.ID)}
			{@const status = article.dropStatus as DropStatus}
			<GPanel style={status === 'sold' ? 'opacity:0.55' : ''}>
				<div class="card-head">
					<div>
						<p class="card-id">{article.dropId}</p>
						<p class="card-date">{article.dropDate}</p>
					</div>
					<GChip color={statusColor[status]}>{statusLabel[status]}</GChip>
				</div>

				<div
					class="card-art"
					style={status === 'sold' ? 'filter:grayscale(1) brightness(0.5)' : ''}
				>
					<span class="card-glyph">{article.glyph}</span>
				</div>

				<a class="card-name" href={`/lot/${article.ID}`}>{article.name}</a>
				<p class="card-series">{article.series}</p>

				{#if status !== 'sold'}
					<div class="card-seats">
						{#if article.seatsLeft > 0}
							<GMeter value={(article.seatsLeft / article.seatsTotal) * 100} height={3} />
							<span class="seats-label">{article.seatsLeft}/{article.seatsTotal} restants</span>
						{/if}
					</div>
				{/if}

				<div class="card-prices">
					<div>
						<p class="price-label">Drop</p>
						<p class="price-val">{eur(article.prix)}</p>
					</div>
					<div style="text-align:right">
						<p class="price-label">Resell</p>
						<p class="price-val resell">{eur(article.resellPrice)}</p>
					</div>
				</div>

				{#if status === 'live'}
					<button
						class="cal-cta cta-live"
						disabled={busy || isEntered(article, 'purchase')}
						onclick={() => act(article, 'purchase')}
					>
						{isEntered(article, 'purchase') ? '✓ Acheté' : 'Acheter'}
					</button>
				{:else if status === 'next'}
					<button
						class="cal-cta cta-next"
						disabled={busy || isEntered(article, 'raffle')}
						onclick={() => act(article, 'raffle')}
					>
						{isEntered(article, 'raffle') ? '✓ Inscrit' : 'Entrer raffle'}
					</button>
				{:else if status === 'sold'}
					<button
						class="cal-cta cta-soon"
						disabled={busy || isEntered(article, 'waitlist')}
						onclick={() => act(article, 'waitlist')}
					>
						{isEntered(article, 'waitlist') ? '✓ En liste' : 'Rejoindre WL'}
					</button>
				{:else}
					<button
						class="cal-cta cta-soon"
						disabled={busy || isEntered(article, 'reminder')}
						onclick={() => act(article, 'reminder')}
					>
						{isEntered(article, 'reminder') ? '✓ Rappel actif' : '+ Rappel'}
					</button>
				{/if}
			</GPanel>
		{/each}
	</div>
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

	/* Hero */
	.hero-grid {
		display: grid;
		grid-template-columns: 1.3fr 0.7fr;
		gap: 28px;
		margin-bottom: 32px;
		align-items: center;
	}
	@media (max-width: 768px) {
		.hero-grid {
			grid-template-columns: 1fr;
		}
	}

	.hero-title {
		font-family: 'Newsreader', Georgia, serif;
		font-size: 42px;
		font-weight: 500;
		line-height: 1.05;
		color: #ece5da;
		margin: 6px 0 4px;
	}
	.hero-sub {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13px;
		color: #766d60;
		margin: 0 0 20px;
	}

	.countdown {
		display: flex;
		gap: 10px;
		margin-bottom: 20px;
	}
	.cd-box {
		display: flex;
		flex-direction: column;
		align-items: center;
		padding: 10px 16px;
		border: 1px solid rgba(236, 229, 218, 0.08);
		border-radius: 8px;
		background: #221f1b;
	}
	.cd-val {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 38px;
		font-weight: 600;
		line-height: 1;
		color: #ece5da;
	}
	.cd-label {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 10px;
		letter-spacing: 0.12em;
		text-transform: uppercase;
		color: #766d60;
		margin-top: 3px;
	}

	.hero-actions {
		display: flex;
		gap: 10px;
		flex-wrap: wrap;
		margin-bottom: 14px;
	}
	.btn-primary {
		flex: 1;
		padding: 12px 22px;
		border-radius: 7px;
		border: none;
		background: #86b3a4;
		color: #191714;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13px;
		font-weight: 600;
		letter-spacing: 0.04em;
		cursor: pointer;
		transition: filter 120ms;
	}
	.btn-primary:hover {
		filter: brightness(1.08);
	}
	.btn-ghost {
		padding: 12px 16px;
		border-radius: 7px;
		border: 1px solid rgba(236, 229, 218, 0.14);
		background: transparent;
		color: #a39a8c;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13px;
		font-weight: 500;
		cursor: pointer;
		transition:
			border-color 120ms,
			color 120ms;
	}
	.btn-ghost:hover {
		border-color: rgba(236, 229, 218, 0.28);
		color: #ece5da;
	}
	.hero-meta {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 11px;
		color: #766d60;
	}

	.hero-art {
		aspect-ratio: 3/4;
		border-radius: 12px;
		position: relative;
		overflow: hidden;
		background: radial-gradient(120% 90% at 30% 20%, #4a6a5a 0%, #2a3a32 55%, #191714 100%);
		display: flex;
		align-items: center;
		justify-content: center;
		border: 1px solid rgba(236, 229, 218, 0.1);
	}
	.hero-glyph {
		font-family: 'Newsreader', Georgia, serif;
		font-size: 130px;
		color: rgba(236, 229, 218, 0.85);
		position: relative;
		z-index: 1;
	}
	.hero-badges {
		position: absolute;
		bottom: 12px;
		right: 12px;
		display: flex;
		flex-direction: column;
		gap: 4px;
		z-index: 2;
	}

	/* Calendar */
	.section-head {
		margin-bottom: 14px;
	}
	.calendar-grid {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 12px;
	}
	@media (max-width: 900px) {
		.calendar-grid {
			grid-template-columns: repeat(2, 1fr);
		}
	}
	@media (max-width: 580px) {
		.calendar-grid {
			grid-template-columns: 1fr;
		}
	}

	.card-head {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: 12px;
	}
	.card-id {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 10px;
		color: #86b3a4;
		margin: 0;
	}
	.card-date {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 10px;
		color: #766d60;
		margin: 2px 0 0;
	}

	.card-art {
		height: 90px;
		border-radius: 7px;
		display: flex;
		align-items: center;
		justify-content: center;
		overflow: hidden;
		background: radial-gradient(120% 90% at 30% 20%, #3a5a4a 0%, #221f1b 100%);
		margin-bottom: 10px;
	}
	.card-glyph {
		font-family: 'Newsreader', Georgia, serif;
		font-size: 44px;
		color: rgba(236, 229, 218, 0.8);
	}

	.card-name {
		display: block;
		font-family: 'Newsreader', Georgia, serif;
		font-size: 15px;
		font-weight: 500;
		color: #ece5da;
		margin: 0 0 2px;
		line-height: 1.2;
		text-decoration: none;
	}
	.card-name:hover {
		color: #86b3a4;
	}

	.toast {
		padding: 10px 14px;
		border-radius: 7px;
		border: 1px solid rgba(134, 179, 164, 0.3);
		background: rgba(134, 179, 164, 0.06);
		color: #86b3a4;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13px;
		margin-bottom: 14px;
	}

	.cal-cta:disabled {
		opacity: 0.6;
		cursor: default;
	}
	.card-series {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 11px;
		color: #766d60;
		margin: 0 0 8px;
	}

	.card-seats {
		margin-bottom: 8px;
	}
	.seats-label {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 10px;
		color: #a39a8c;
		display: block;
		margin-top: 4px;
	}

	.card-prices {
		display: flex;
		justify-content: space-between;
		align-items: flex-end;
		margin-bottom: 12px;
	}
	.price-label {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 10px;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		color: #766d60;
		margin: 0;
	}
	.price-val {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 17px;
		font-weight: 600;
		color: #ece5da;
		margin: 0;
	}
	.price-val.resell {
		color: #86b3a4;
	}

	.cal-cta {
		width: 100%;
		padding: 9px;
		border-radius: 7px;
		border: none;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 12px;
		font-weight: 600;
		cursor: pointer;
		transition: filter 120ms;
	}
	.cta-live {
		background: #86b3a4;
		color: #191714;
	}
	.cta-live:hover {
		filter: brightness(1.08);
	}
	.cta-next {
		background: rgba(236, 229, 218, 0.1);
		color: #ece5da;
		border: 1px solid rgba(236, 229, 218, 0.14);
	}
	.cta-next:hover {
		background: rgba(236, 229, 218, 0.16);
	}
	.cta-soon {
		background: transparent;
		border: 1px solid rgba(134, 179, 164, 0.28);
		color: #86b3a4;
	}
	.cta-soon:hover {
		background: rgba(134, 179, 164, 0.08);
	}
</style>
