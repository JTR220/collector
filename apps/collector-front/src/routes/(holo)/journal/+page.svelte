<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { auth, isAuthenticated } from '$lib/stores/auth';
	import { refreshStats } from '$lib/stores/stats';
	import { fetchArticles, type ArticleAPI } from '$lib/api/catalog';
	import {
		fetchMyJournal,
		fetchMyWishlist,
		likeJournalEntry,
		removeFromWishlist,
		type JournalEntry,
		type WishlistItem
	} from '$lib/api/engagement';
	import { eur } from '$lib/utils/format';
	import GPanel from '$lib/components/galerie/GPanel.svelte';
	import GMeter from '$lib/components/galerie/GMeter.svelte';
	import GChip from '$lib/components/galerie/GChip.svelte';
	import GAvatar from '$lib/components/galerie/GAvatar.svelte';
	import Kicker from '$lib/components/galerie/Kicker.svelte';

	const tabs = ['Feed', 'Collection', 'Wishlist', 'Listes', 'Notes', 'Trades'];
	let activeTab = $state('Feed');
	let search = $state('');

	let articles = $state<ArticleAPI[]>([]);
	let journal = $state<JournalEntry[]>([]);
	let wishlist = $state<WishlistItem[]>([]);
	let loading = $state(true);
	let shareMsg = $state<number | null>(null);

	onMount(async () => {
		if (!$isAuthenticated || !$auth.token) {
			goto('/login');
			return;
		}
		// Onglet pré-sélectionné via /journal?tab=Wishlist
		const tabParam = $page.url.searchParams.get('tab');
		if (tabParam && tabs.includes(tabParam)) activeTab = tabParam;
		try {
			[articles, journal, wishlist] = await Promise.all([
				fetchArticles(),
				fetchMyJournal($auth.token),
				fetchMyWishlist($auth.token)
			]);
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	});

	type DiaryKind = 'acquis' | 'trade' | 'wishlist' | 'noté' | 'vendu';
	const kindColor: Record<DiaryKind, string> = {
		acquis: '#86b3a4',
		trade: '#a39a8c',
		wishlist: '#86b3a4',
		noté: '#a39a8c',
		vendu: '#d79c86'
	};
	const tabKind: Record<string, DiaryKind | null> = {
		Feed: null,
		Collection: 'acquis',
		Notes: 'noté',
		Trades: 'trade'
	};

	const matchSearch = (name: string | undefined) =>
		!search || (name ?? '').toLowerCase().includes(search.toLowerCase());

	const diary = $derived(
		journal.filter((entry) => {
			const kind = tabKind[activeTab];
			if (activeTab === 'Wishlist' || activeTab === 'Listes') return false;
			return (!kind || entry.kind === kind) && matchSearch(entry.article?.name);
		})
	);

	const filteredWishlist = $derived(wishlist.filter((w) => matchSearch(w.article?.name)));

	// Pièces fétiches : dernières acquisitions, complétées par le catalogue (sans doublon)
	const favs = $derived(
		(() => {
			const acquired = journal
				.filter((j) => j.kind === 'acquis' && j.article)
				.map((j) => j.article as ArticleAPI);
			const seen = new Set<number>();
			const list: ArticleAPI[] = [];
			for (const a of [...acquired, ...articles]) {
				if (seen.has(a.ID)) continue;
				seen.add(a.ID);
				list.push(a);
				if (list.length === 4) break;
			}
			return list;
		})()
	);

	async function like(entry: JournalEntry) {
		if (!$auth.token) return;
		try {
			const res = await likeJournalEntry($auth.token, entry.ID);
			journal = journal.map((j) => (j.ID === entry.ID ? { ...j, likes: res.likes } : j));
		} catch (e) {
			console.error(e);
		}
	}

	async function share(entry: JournalEntry) {
		const url = `${location.origin}/lot/${entry.articleId}`;
		try {
			await navigator.clipboard.writeText(url);
			shareMsg = entry.ID;
			setTimeout(() => (shareMsg = null), 2000);
		} catch {
			/* clipboard indisponible */
		}
	}

	async function removeWish(item: WishlistItem) {
		if (!$auth.token) return;
		try {
			await removeFromWishlist($auth.token, item.articleId);
			wishlist = wishlist.filter((w) => w.ID !== item.ID);
			refreshStats();
		} catch (e) {
			console.error(e);
		}
	}

	function entryDate(entry: JournalEntry): { month: string; day: number } {
		const d = new Date(entry.CreatedAt);
		return {
			month: d.toLocaleDateString('fr-FR', { month: 'short' }).replace('.', ''),
			day: d.getDate()
		};
	}

	const popularLists = [
		{ title: 'Top TCG Holo 1ère édition', cat: 'TCG', handle: 'holo_king', count: 24 },
		{ title: 'Consoles scellées all-time', cat: 'Console', handle: 'pack_ripper', count: 18 },
		{ title: 'Vinyles cultes année 2000', cat: 'Vinyle', handle: 'groove_atlas', count: 31 },
		{ title: 'Designer toys édition limitée', cat: 'Designer Toy', handle: 'soho_pulse', count: 12 }
	];

	const friendActivity = [
		{ handle: 'holo_king', action: 'a acquis', target: 'Charizard', slug: 'PKM-001', rating: null },
		{
			handle: 'pack_ripper',
			action: 'a noté ★★★★★',
			target: 'Game Boy Color',
			slug: 'GBC-014',
			rating: 5
		},
		{
			handle: 'groove_atlas',
			action: 'a mis en wishlist',
			target: 'Daft Punk Discovery',
			slug: 'VNL-022',
			rating: null
		},
		{
			handle: 'arcade_twin',
			action: 'a publié une note',
			target: 'Action Comics #1',
			slug: 'CMX-007',
			rating: 3
		},
		{
			handle: 'soho_pulse',
			action: 'a vendu',
			target: 'Bearbrick 1000%',
			slug: 'FIG-101',
			rating: null
		}
	];

	function friendLink(slug: string): string {
		const a = articles.find((x) => x.slug === slug);
		return a ? `/lot/${a.ID}` : '/';
	}

	const ratingDist = $derived(
		[5, 4, 3, 2, 1].map((stars) => {
			const rated = journal.filter((j) => j.rating > 0);
			const count = rated.filter((j) => j.rating === stars).length;
			return { stars, count, pct: rated.length ? Math.round((count / rated.length) * 100) : 0 };
		})
	);
</script>

<svelte:head><title>Journal · Collector.shop</title></svelte:head>

<div class="tabs-bar">
	<div class="tabs">
		{#each tabs as tab}
			<button class="tab-btn" class:tab-active={activeTab === tab} onclick={() => (activeTab = tab)}
				>{tab}</button
			>
		{/each}
	</div>
	<input class="search-input" placeholder="Rechercher…" type="search" bind:value={search} />
</div>

{#if favs.length > 0 && activeTab === 'Feed'}
	<div style="margin-bottom:22px">
		<Kicker>Pièces fétiches</Kicker>
		<div class="fav-grid">
			{#each favs as article (article.ID)}
				<a class="fav-card" href={`/lot/${article.ID}`}>
					<div class="fav-art">
						<span class="fav-glyph">{article.glyph}</span>
					</div>
					<div class="fav-info">
						<p class="fav-id">{article.slug}</p>
						<div class="fav-stars">
							{'★'.repeat(article.rarityScore)}{'☆'.repeat(5 - article.rarityScore)}
						</div>
						<p class="fav-name">{article.name}</p>
						<p class="fav-sub">{article.year} · {article.grade}</p>
					</div>
				</a>
			{/each}
		</div>
	</div>
{/if}

<div class="body-grid">
	<div>
		{#if activeTab === 'Wishlist'}
			<Kicker style="margin-bottom:14px"
				>Ma wishlist · {filteredWishlist.length} pièce{filteredWishlist.length > 1
					? 's'
					: ''}</Kicker
			>
			{#each filteredWishlist as item (item.ID)}
				<div class="diary-entry">
					<div class="diary-art">
						<span class="diary-glyph">{item.article.glyph}</span>
					</div>
					<div class="diary-content">
						<GChip color="#86b3a4">wishlist</GChip>
						<a class="entry-name" href={`/lot/${item.article.ID}`}>{item.article.name}</a>
						<p class="entry-sub">{item.article.series} · {eur(item.article.prix)}</p>
						<div class="entry-actions">
							<a class="action-link" href={`/lot/${item.article.ID}`}>voir le lot</a>
							<button class="action-btn" onclick={() => removeWish(item)}>× retirer</button>
						</div>
					</div>
				</div>
			{:else}
				<p class="empty-msg">
					{#if loading}Chargement…{:else}Wishlist vide — ajoutez des pièces depuis la <a href="/"
							>vitrine</a
						>.{/if}
				</p>
			{/each}
		{:else if activeTab === 'Listes'}
			<Kicker style="margin-bottom:14px">Listes populaires</Kicker>
			{#each popularLists as list}
				<a class="list-big-row" href={`/?cat=${encodeURIComponent(list.cat)}`}>
					<div class="mini-covers">
						{#each [0, 1, 2] as i}
							<div
								class="mini-cover"
								style="background:radial-gradient(120% 90% at 30% 20%,#3a5a4a,#221f1b);margin-left:{i >
								0
									? '-14px'
									: '0'};z-index:{3 - i}"
							></div>
						{/each}
					</div>
					<div>
						<p class="list-title">{list.title}</p>
						<p class="list-meta">
							par @{list.handle} · {list.count} pièces · voir la catégorie {list.cat} →
						</p>
					</div>
				</a>
			{/each}
		{:else}
			<Kicker style="margin-bottom:14px">
				{activeTab === 'Feed' ? 'Mon journal' : activeTab} · {diary.length} entrée{diary.length > 1
					? 's'
					: ''}
			</Kicker>
			{#each diary as entry (entry.ID)}
				{@const color = kindColor[entry.kind as DiaryKind] ?? '#a39a8c'}
				{@const d = entryDate(entry)}
				<div class="diary-entry">
					<div class="diary-date">
						<span class="date-month">{d.month}</span>
						<span class="date-day">{d.day}</span>
					</div>

					<div class="diary-art">
						<span class="diary-glyph">{entry.article?.glyph ?? '◆'}</span>
					</div>

					<div class="diary-content">
						<GChip {color}>{entry.kind}</GChip>
						<a class="entry-name" href={`/lot/${entry.articleId}`}
							>{entry.article?.name ?? `#${entry.articleId}`}</a
						>
						<p class="entry-sub">{entry.article?.series ?? ''} · {eur(entry.article?.prix ?? 0)}</p>
						{#if entry.rating}
							<div class="entry-stars">
								{'★'.repeat(entry.rating)}{'☆'.repeat(5 - entry.rating)}
							</div>
						{/if}
						{#if entry.note}
							<p class="entry-note">«&nbsp;{entry.note}&nbsp;»</p>
						{/if}
						<div class="entry-actions">
							<button class="action-btn" onclick={() => like(entry)}>♡ {entry.likes}</button>
							<button class="action-btn" onclick={() => share(entry)}>
								{shareMsg === entry.ID ? '✓ lien copié' : '↗ partager'}
							</button>
							{#if entry.xp}<span class="entry-xp">+{entry.xp} XP</span>{/if}
						</div>
					</div>
				</div>
			{:else}
				<p class="empty-msg">
					{#if loading}
						Chargement du journal…
					{:else}
						Aucune entrée pour l'instant — achetez un lot ou publiez un avis depuis une
						<a href="/">page de lot</a> pour alimenter votre journal.
					{/if}
				</p>
			{/each}
		{/if}
	</div>

	<div class="aside">
		<GPanel style="margin-bottom:14px">
			<Kicker style="margin-bottom:14px" color="#766d60">Listes populaires</Kicker>
			<div class="list-stack">
				{#each popularLists as list}
					<a class="list-row" href={`/?cat=${encodeURIComponent(list.cat)}`}>
						<div class="mini-covers">
							{#each [0, 1, 2] as i}
								<div
									class="mini-cover"
									style="background:radial-gradient(120% 90% at 30% 20%,#3a5a4a,#221f1b);margin-left:{i >
									0
										? '-14px'
										: '0'};z-index:{3 - i}"
								></div>
							{/each}
						</div>
						<div>
							<p class="list-title">{list.title}</p>
							<p class="list-meta">par @{list.handle} · {list.count} pièces</p>
						</div>
					</a>
				{/each}
			</div>
		</GPanel>

		<GPanel style="margin-bottom:14px">
			<Kicker style="margin-bottom:14px" color="#86b3a4">Vu chez tes amis</Kicker>
			<div class="friend-stack">
				{#each friendActivity as f}
					<div class="friend-row">
						<GAvatar initials={f.handle[0].toUpperCase()} size={32} square />
						<div class="friend-text">
							<span class="friend-handle">@{f.handle}</span>
							<span class="friend-action"> {f.action} </span>
							<a class="friend-target" href={friendLink(f.slug)}>{f.target}</a>
							{#if f.rating}<span class="friend-stars"> {'★'.repeat(f.rating)}</span>{/if}
						</div>
					</div>
				{/each}
			</div>
		</GPanel>

		<GPanel>
			<Kicker style="margin-bottom:14px" color="#a39a8c">Répartition de mes notes</Kicker>
			<div class="rating-stack">
				{#each ratingDist as r}
					<div class="rating-row">
						<span class="rating-stars">{'★'.repeat(r.stars)}</span>
						<div style="flex:1"><GMeter value={r.pct} height={5} /></div>
						<span class="rating-count">{r.count}</span>
					</div>
				{/each}
			</div>
		</GPanel>
	</div>
</div>

<style>
	/* Tabs */
	.tabs-bar {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 24px;
		gap: 12px;
		flex-wrap: wrap;
	}
	.tabs {
		display: flex;
		gap: 2px;
		flex-wrap: wrap;
	}
	.tab-btn {
		padding: 6px 14px;
		background: none;
		border: none;
		border-bottom: 2px solid transparent;
		cursor: pointer;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13px;
		font-weight: 500;
		color: #766d60;
		transition:
			color 120ms,
			border-color 120ms;
	}
	.tab-btn:hover {
		color: #a39a8c;
	}
	.tab-active {
		color: #ece5da !important;
		border-bottom-color: #86b3a4 !important;
		font-weight: 600;
	}
	.search-input {
		padding: 7px 14px;
		border: 1px solid rgba(236, 229, 218, 0.1);
		border-radius: 7px;
		background: #221f1b;
		color: #ece5da;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13px;
		outline: none;
		min-width: 160px;
		transition: border-color 120ms;
	}
	.search-input:focus {
		border-color: rgba(134, 179, 164, 0.4);
	}
	.search-input::placeholder {
		color: #766d60;
	}

	/* Fav grid */
	.fav-grid {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 12px;
		margin-top: 12px;
	}
	@media (max-width: 900px) {
		.fav-grid {
			grid-template-columns: repeat(2, 1fr);
		}
	}
	.fav-card {
		display: flex;
		flex-direction: column;
		gap: 0;
		text-decoration: none;
	}
	.fav-card:hover .fav-name {
		color: #86b3a4;
	}
	.fav-art {
		height: 120px;
		border-radius: 8px;
		background: radial-gradient(120% 90% at 30% 20%, #3a5a4a 0%, #221f1b 100%);
		border: 1px solid rgba(236, 229, 218, 0.08);
		display: flex;
		align-items: center;
		justify-content: center;
		overflow: hidden;
	}
	.fav-glyph {
		font-family: 'Newsreader', Georgia, serif;
		font-size: 52px;
		color: rgba(236, 229, 218, 0.8);
	}
	.fav-info {
		padding: 8px 0 4px;
	}
	.fav-id {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 10px;
		color: #86b3a4;
		margin: 0;
	}
	.fav-stars {
		font-size: 13px;
		color: #a39a8c;
		margin: 2px 0;
	}
	.fav-name {
		font-family: 'Newsreader', Georgia, serif;
		font-size: 14px;
		font-weight: 500;
		color: #ece5da;
		margin: 0;
		line-height: 1.2;
		transition: color 120ms;
	}
	.fav-sub {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 11px;
		color: #766d60;
		margin: 2px 0 0;
	}

	/* Body */
	.body-grid {
		display: grid;
		grid-template-columns: 1.45fr 1fr;
		gap: 18px;
	}
	@media (max-width: 900px) {
		.body-grid {
			grid-template-columns: 1fr;
		}
	}

	/* Diary */
	.diary-entry {
		display: flex;
		align-items: flex-start;
		gap: 14px;
		border-top: 1px solid rgba(236, 229, 218, 0.07);
		padding: 16px 0;
	}
	.diary-entry:first-of-type {
		border-top: none;
		padding-top: 0;
	}
	.diary-date {
		display: flex;
		flex-direction: column;
		align-items: center;
		flex-shrink: 0;
		width: 44px;
	}
	.date-month {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 9px;
		letter-spacing: 0.16em;
		text-transform: uppercase;
		color: #766d60;
	}
	.date-day {
		font-family: 'Newsreader', Georgia, serif;
		font-size: 30px;
		font-weight: 500;
		color: #ece5da;
		line-height: 0.95;
	}
	.diary-art {
		width: 72px;
		height: 72px;
		flex-shrink: 0;
		border-radius: 8px;
		background: radial-gradient(120% 90% at 30% 20%, #3a5a4a 0%, #221f1b 100%);
		border: 1px solid rgba(236, 229, 218, 0.08);
		display: flex;
		align-items: center;
		justify-content: center;
	}
	.diary-glyph {
		font-family: 'Newsreader', Georgia, serif;
		font-size: 32px;
		color: rgba(236, 229, 218, 0.8);
	}
	.diary-content {
		flex: 1;
	}
	.entry-name {
		display: block;
		font-family: 'Newsreader', Georgia, serif;
		font-size: 16px;
		font-weight: 500;
		color: #ece5da;
		margin: 6px 0 2px;
		line-height: 1.15;
		text-decoration: none;
	}
	.entry-name:hover {
		color: #86b3a4;
	}
	.entry-sub {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 10px;
		color: #766d60;
		margin: 0;
	}
	.entry-stars {
		font-size: 14px;
		color: #a39a8c;
		margin: 4px 0;
	}
	.entry-note {
		font-family: 'Newsreader', Georgia, serif;
		font-size: 13px;
		font-style: italic;
		color: #a39a8c;
		margin: 6px 0 8px;
		line-height: 1.55;
	}
	.entry-actions {
		display: flex;
		gap: 12px;
		flex-wrap: wrap;
		align-items: center;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 11px;
		color: #766d60;
		margin-top: 6px;
	}
	.action-btn {
		background: none;
		border: none;
		padding: 0;
		font: inherit;
		color: #766d60;
		cursor: pointer;
		transition: color 120ms;
	}
	.action-btn:hover {
		color: #86b3a4;
	}
	.action-link {
		color: #766d60;
		text-decoration: none;
	}
	.action-link:hover {
		color: #a39a8c;
	}
	.entry-xp {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 10px;
		color: #86b3a4;
	}
	.empty-msg {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13px;
		color: #766d60;
		padding: 20px 0;
		line-height: 1.6;
	}
	.empty-msg a {
		color: #86b3a4;
	}

	/* Aside */
	.list-stack {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}
	.list-row {
		display: flex;
		align-items: center;
		gap: 10px;
		text-decoration: none;
	}
	.list-row:hover .list-title {
		color: #86b3a4;
	}
	.list-big-row {
		display: flex;
		align-items: center;
		gap: 14px;
		padding: 14px;
		border: 1px solid rgba(236, 229, 218, 0.1);
		border-radius: 9px;
		text-decoration: none;
		margin-bottom: 10px;
		transition: border-color 120ms;
	}
	.list-big-row:hover {
		border-color: rgba(134, 179, 164, 0.4);
	}
	.list-big-row:hover .list-title {
		color: #86b3a4;
	}
	.mini-covers {
		display: flex;
		position: relative;
		flex-shrink: 0;
	}
	.mini-cover {
		width: 32px;
		height: 32px;
		border-radius: 4px;
		border: 2px solid #191714;
		flex-shrink: 0;
	}
	.list-title {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13px;
		color: #ece5da;
		margin: 0;
		transition: color 120ms;
	}
	.list-meta {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 10px;
		color: #766d60;
		margin: 2px 0 0;
	}

	.friend-stack {
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.friend-row {
		display: flex;
		align-items: center;
		gap: 8px;
	}
	.friend-text {
		font-size: 12px;
		flex: 1;
		line-height: 1.4;
	}
	.friend-handle {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 11px;
		color: #86b3a4;
	}
	.friend-action {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		color: #766d60;
		font-style: italic;
	}
	.friend-target {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		color: #ece5da;
		text-decoration: none;
	}
	.friend-target:hover {
		color: #86b3a4;
	}
	.friend-stars {
		color: #a39a8c;
	}

	.rating-stack {
		display: flex;
		flex-direction: column;
		gap: 7px;
	}
	.rating-row {
		display: flex;
		align-items: center;
		gap: 8px;
	}
	.rating-stars {
		font-size: 13px;
		color: #a39a8c;
		width: 52px;
		flex-shrink: 0;
	}
	.rating-count {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 10px;
		color: #766d60;
		width: 24px;
		text-align: right;
	}
</style>
