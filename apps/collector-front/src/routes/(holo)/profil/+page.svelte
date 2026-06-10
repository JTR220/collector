<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth, isAuthenticated } from '$lib/stores/auth';
	import { playerStats, refreshStats } from '$lib/stores/stats';
	import { fetchMe, type MeResponse } from '$lib/api/auth';
	import { fetchMyJournal, type JournalEntry } from '$lib/api/engagement';
	import GPanel from '$lib/components/galerie/GPanel.svelte';
	import GMeter from '$lib/components/galerie/GMeter.svelte';
	import GSpark from '$lib/components/galerie/GSpark.svelte';
	import GAvatar from '$lib/components/galerie/GAvatar.svelte';
	import GChip from '$lib/components/galerie/GChip.svelte';
	import Kicker from '$lib/components/galerie/Kicker.svelte';

	let me = $state<MeResponse | null>(null);
	let loading = $state(true);
	let journal = $state<JournalEntry[]>([]);

	const initials = $derived(
		me?.name
			? me.name
					.split(' ')
					.map((w) => w[0])
					.join('')
					.toUpperCase()
					.slice(0, 2)
			: '??'
	);
	const handle = $derived(me?.name ? me.name.toLowerCase().replace(/\s+/g, '_') : '…');

	onMount(async () => {
		if (!$isAuthenticated || !$auth.token) {
			goto('/login');
			return;
		}
		try {
			me = await fetchMe($auth.token);
			await refreshStats();
			journal = await fetchMyJournal($auth.token);
		} catch {
			auth.logout();
			goto('/login');
		} finally {
			loading = false;
		}
	});

	function logout() {
		auth.logout();
		goto('/login');
	}

	const xp = $derived($playerStats?.xp ?? 0);
	const xpToNext = $derived($playerStats?.xpToNext ?? 350);
	const level = $derived($playerStats?.level ?? 1);
	// Piste du saison-pass centrée sur le niveau courant
	const tiers = $derived(Array.from({ length: 11 }, (_, i) => Math.max(1, level - 2) + i));

	const countKind = (kind: string) => journal.filter((j) => j.kind === kind).length;
	// Sparkline cumulative : progression du compteur dans le temps
	function cumulSpark(total: number): number[] {
		return Array.from({ length: 8 }, (_, i) => Math.round((total * (i + 1)) / 8));
	}

	const stats = $derived([
		{ label: 'Pièces', value: String(countKind('acquis')), spark: cumulSpark(countKind('acquis')) },
		{
			label: 'Wishlist',
			value: String($playerStats?.wishlistCount ?? 0),
			spark: cumulSpark($playerStats?.wishlistCount ?? 0)
		},
		{ label: 'Vendues', value: String(countKind('vendu')), spark: cumulSpark(countKind('vendu')) },
		{ label: 'Notes', value: String(countKind('noté')), spark: cumulSpark(countKind('noté')) },
		{ label: 'XP', value: String(xp), spark: cumulSpark(xp) },
		{ label: 'Trades', value: String(countKind('trade')), spark: cumulSpark(countKind('trade')) }
	]);

	const badges = $derived([
		{ n: 'Holo Hunter', s: '1 holo acquise', on: countKind('acquis') >= 1 },
		{ n: 'Critique', s: '1 avis écrit', on: countKind('noté') >= 1 },
		{
			n: 'Wishlist Addict',
			s: '3 pièces en wishlist',
			on: ($playerStats?.wishlistCount ?? 0) >= 3
		},
		{ n: 'Membre fondateur', s: 'compte saison 03', on: true },
		{ n: 'Speed Trader', s: '1 trade validé', on: countKind('trade') >= 1 },
		{ n: 'Vendeur', s: '1 pièce vendue', on: countKind('vendu') >= 1 },
		{ n: 'Niveau 5', s: 'atteindre niv. 5', on: level >= 5 },
		{ n: 'Mille XP', s: '1 000 XP cumulés', on: xp >= 1000 }
	]);

	const kindLabel: Record<string, string> = {
		acquis: 'Acquis',
		vendu: 'Vendu',
		noté: 'Noté',
		trade: 'Trade',
		wishlist: 'Wishlist'
	};
	const activity = $derived(
		journal.slice(0, 8).map((j) => ({
			date: new Date(j.CreatedAt).toLocaleDateString('fr-FR', { day: '2-digit', month: '2-digit' }),
			kind: kindLabel[j.kind] ?? j.kind,
			label: j.article?.name ?? `#${j.articleId}`,
			xp: j.xp ? `+${j.xp} XP` : null,
			articleId: j.articleId
		}))
	);
</script>

<svelte:head><title>PROFIL · Collector.shop</title></svelte:head>

{#if loading}
	<div class="state-msg">Chargement du profil…</div>
{:else if me}
	<!-- Bandeau identité -->
	<section class="identity">
		<GAvatar {initials} size={96} />
		<div class="identity-text">
			<Kicker>Saison 03 — chasseuse holo</Kicker>
			<h1 class="identity-name">{me.name}</h1>
			<div class="identity-meta">
				<span>@{handle}</span>
				<span class="meta-sep">·</span>
				<span style="color:#86b3a4">membre fondateur</span>
			</div>
			<p class="identity-bio">« Chasse les holos et les pièces rares. Trades ouverts. »</p>
		</div>
		<div class="identity-btns">
			<button class="btn-primary" onclick={() => goto('/journal')}>Mon journal</button>
			<button class="btn-ghost" onclick={logout}>Se déconnecter</button>
		</div>
	</section>

	<!-- Saison-pass -->
	<GPanel style="padding:22px;margin-bottom:14px">
		<div class="pass-top">
			<div>
				<Kicker>Saison-pass · 03</Kicker>
				<div class="pass-title">
					Niveau {level}
					<span class="pass-sep">·</span>
					<span class="pass-xp"
						>{xp.toLocaleString('fr-FR')} / {xpToNext.toLocaleString('fr-FR')} XP</span
					>
				</div>
			</div>
			<div style="text-align:right">
				<Kicker>Fin de saison</Kicker>
				<div class="pass-end">23 j 14 h</div>
			</div>
		</div>
		<div style="margin-top:14px">
			<GMeter value={Math.round((xp / xpToNext) * 100)} height={6} />
		</div>
		<div class="pass-track">
			{#each tiers as lvl}
				{@const done = lvl < level}
				{@const cur = lvl === level}
				<div class="pass-node-wrap">
					<div
						class="pass-node"
						style="
							border-color:{cur ? '#86b3a4' : done ? 'rgba(236,229,218,0.16)' : 'rgba(236,229,218,0.10)'};
							background:{cur ? 'rgba(255,255,255,0.05)' : 'transparent'};
							box-shadow:{cur ? '0 0 0 4px rgba(255,255,255,0.04)' : 'none'};
							color:{cur ? '#86b3a4' : done ? '#a39a8c' : '#766d60'};
						"
					>
						{lvl}
					</div>
					<span class="pass-node-label" style="color:{cur ? '#86b3a4' : '#766d60'}">
						{cur ? 'Ici' : done ? 'Reçu' : `+${(tiers.indexOf(lvl) + 1) * 80}`}
					</span>
				</div>
			{/each}
		</div>
	</GPanel>

	<!-- 3 colonnes -->
	<div class="three-col">
		<!-- Statistiques -->
		<GPanel>
			<Kicker>Statistiques</Kicker>
			<div class="stats-grid">
				{#each stats as s}
					<div class="stat-cell">
						<span class="stat-label">{s.label}</span>
						<span class="stat-val">{s.value}</span>
						<GSpark values={s.spark} w={130} h={26} dot={false} />
					</div>
				{/each}
			</div>
		</GPanel>

		<!-- Badges -->
		<GPanel>
			<Kicker>Badges · 6 / 8</Kicker>
			<div class="badge-grid">
				{#each badges as b}
					<div
						class="badge-card"
						style="
							border-color:{b.on ? 'rgba(236,229,218,0.16)' : 'rgba(236,229,218,0.10)'};
							background:{b.on ? 'rgba(255,255,255,0.05)' : 'transparent'};
							opacity:{b.on ? 1 : 0.5};
						"
					>
						<span
							class="badge-diamond"
							style="background:{b.on ? '#86b3a4' : 'transparent'};border:{b.on
								? 'none'
								: '1px solid #766d60'}"
						></span>
						<div class="badge-name">{b.n}</div>
						<div class="badge-desc">{b.s}</div>
					</div>
				{/each}
			</div>
		</GPanel>

		<!-- Activité -->
		<GPanel>
			<Kicker>Activité récente</Kicker>
			<div class="activity-list">
				{#each activity as a}
					<div class="activity-row">
						<span class="activity-date">{a.date}</span>
						<GChip>{a.kind}</GChip>
						<a class="activity-label" href={`/lot/${a.articleId}`}>{a.label}</a>
						{#if a.xp}
							<span class="activity-xp">{a.xp}</span>
						{/if}
					</div>
				{:else}
					<p class="activity-empty">
						Aucune activité pour l'instant — achetez, notez ou ajoutez une pièce en wishlist depuis
						la <a href="/">vitrine</a>.
					</p>
				{/each}
			</div>
		</GPanel>
	</div>
{/if}

<style>
	.state-msg {
		text-align: center;
		padding: 80px 0;
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 12px;
		color: #a39a8c;
		letter-spacing: 0.18em;
	}

	/* Identité */
	.identity {
		display: flex;
		align-items: center;
		gap: 24px;
		padding: 10px 0 22px;
		flex-wrap: wrap;
	}
	.identity-text {
		flex: 1;
		min-width: 200px;
	}
	.identity-name {
		font-family: 'Newsreader', Georgia, serif;
		font-weight: 500;
		font-size: 46px;
		line-height: 1.05;
		margin: 8px 0 0;
		letter-spacing: -0.01em;
		color: #ece5da;
	}
	.identity-meta {
		display: flex;
		gap: 10px;
		align-items: center;
		margin-top: 8px;
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 12.5px;
		color: #a39a8c;
	}
	.meta-sep {
		color: #766d60;
	}
	.identity-bio {
		font-family: 'Newsreader', Georgia, serif;
		font-style: italic;
		font-size: 15px;
		color: #a39a8c;
		margin-top: 10px;
		max-width: 560px;
		line-height: 1.5;
	}
	.identity-btns {
		display: flex;
		flex-direction: column;
		gap: 8px;
		flex-shrink: 0;
	}
	.btn-primary {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 12.5px;
		font-weight: 600;
		padding: 10px 22px;
		border-radius: 7px;
		border: none;
		cursor: pointer;
		background: #86b3a4;
		color: #191714;
		transition: filter 120ms;
	}
	.btn-primary:hover {
		filter: brightness(1.08);
	}
	.btn-ghost {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 12.5px;
		padding: 9px 22px;
		border-radius: 7px;
		border: 1px solid rgba(236, 229, 218, 0.1);
		cursor: pointer;
		background: transparent;
		color: #a39a8c;
		transition:
			border-color 120ms,
			color 120ms;
	}
	.btn-ghost:hover {
		border-color: rgba(236, 229, 218, 0.22);
		color: #ece5da;
	}

	/* Pass */
	.pass-top {
		display: flex;
		justify-content: space-between;
		align-items: flex-end;
	}
	.pass-title {
		font-family: 'Newsreader', Georgia, serif;
		font-size: 24px;
		margin-top: 6px;
		color: #ece5da;
	}
	.pass-sep {
		color: #766d60;
	}
	.pass-xp {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 19px;
	}
	.pass-end {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 22px;
		color: #ece5da;
		margin-top: 4px;
	}

	.pass-track {
		display: grid;
		grid-template-columns: repeat(11, 1fr);
		gap: 6px;
		margin-top: 22px;
	}
	.pass-node-wrap {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 6px;
	}
	.pass-node {
		width: 40px;
		height: 40px;
		border-radius: 50%;
		border: 1.5px solid;
		display: flex;
		align-items: center;
		justify-content: center;
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 13px;
	}
	.pass-node-label {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 9.5px;
		letter-spacing: 0.08em;
		text-align: center;
		line-height: 1.3;
	}

	/* 3 colonnes */
	.three-col {
		display: grid;
		grid-template-columns: 1.1fr 1fr 0.95fr;
		gap: 14px;
	}
	@media (max-width: 1024px) {
		.three-col {
			grid-template-columns: 1fr 1fr;
		}
	}
	@media (max-width: 640px) {
		.three-col {
			grid-template-columns: 1fr;
		}
	}

	/* Stats */
	.stats-grid {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 10px;
		margin-top: 12px;
	}
	.stat-cell {
		padding: 11px 13px;
		border: 1px solid rgba(236, 229, 218, 0.1);
		border-radius: 8px;
		background: rgba(255, 255, 255, 0.03);
		display: flex;
		flex-direction: column;
		gap: 4px;
	}
	.stat-label {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 10.5px;
		letter-spacing: 0.14em;
		text-transform: uppercase;
		color: #766d60;
	}
	.stat-val {
		font-family: 'Newsreader', Georgia, serif;
		font-size: 32px;
		line-height: 1;
		margin-top: 3px;
		color: #ece5da;
	}

	/* Badges */
	.badge-grid {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 8px;
		margin-top: 12px;
	}
	.badge-card {
		padding: 12px 13px 14px;
		border-radius: 8px;
		border: 1px solid;
	}
	.badge-diamond {
		width: 11px;
		height: 11px;
		display: inline-block;
		border-radius: 2px;
		transform: rotate(45deg);
	}
	.badge-name {
		font-family: 'Newsreader', Georgia, serif;
		font-size: 14px;
		margin-top: 9px;
		color: #ece5da;
	}
	.badge-desc {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 11px;
		color: #a39a8c;
		margin-top: 2px;
	}

	/* Activité */
	.activity-list {
		display: flex;
		flex-direction: column;
		margin-top: 6px;
	}
	.activity-row {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 10px 0;
		border-bottom: 1px solid rgba(236, 229, 218, 0.1);
	}
	.activity-date {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 11px;
		color: #766d60;
		width: 38px;
		flex-shrink: 0;
	}
	.activity-label {
		flex: 1;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 12.5px;
		color: #ece5da;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		text-decoration: none;
	}
	.activity-label:hover {
		color: #86b3a4;
	}
	.activity-empty {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 12.5px;
		color: #766d60;
		line-height: 1.5;
		padding: 12px 0;
	}
	.activity-empty a {
		color: #86b3a4;
	}
	.activity-xp {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 11.5px;
		color: #86b3a4;
		flex-shrink: 0;
	}
</style>
