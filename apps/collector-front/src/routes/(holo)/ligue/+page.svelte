<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth, isAuthenticated } from '$lib/stores/auth';
	import { refreshStats } from '$lib/stores/stats';
	import { fetchLeague, challengeRival, type LeagueRow } from '$lib/api/engagement';
	import { sparkPath } from '$lib/utils/format';
	import GPanel from '$lib/components/galerie/GPanel.svelte';
	import GMeter from '$lib/components/galerie/GMeter.svelte';
	import GSpark from '$lib/components/galerie/GSpark.svelte';
	import GAvatar from '$lib/components/galerie/GAvatar.svelte';
	import GChip from '$lib/components/galerie/GChip.svelte';
	import Kicker from '$lib/components/galerie/Kicker.svelte';

	let league = $state<LeagueRow[]>([]);
	let challengeMsg = $state<string | null>(null);
	let busy = $state(false);

	onMount(async () => {
		if (!$isAuthenticated || !$auth.token) {
			goto('/login');
			return;
		}
		try {
			league = await fetchLeague($auth.token);
		} catch (e) {
			console.error(e);
		}
	});

	const tiers = [
		{ label: 'Bronze',  active: false, done: true  },
		{ label: 'Argent',  active: false, done: true  },
		{ label: 'Or',      active: true,  done: true  },
		{ label: 'Diamant', active: false, done: false },
		{ label: 'Master',  active: false, done: false },
		{ label: 'Légende', active: false, done: false },
	];

	const myIndex = $derived(league.findIndex((r) => r.me));
	const myRow = $derived(myIndex >= 0 ? league[myIndex] : null);
	// Rival : le joueur juste au-dessus (ou en dessous si premier)
	const rival = $derived(
		myIndex > 0 ? league[myIndex - 1] : myIndex === 0 && league.length > 1 ? league[1] : null
	);
	const gapToTop7 = $derived(
		myIndex >= 0 && league.length > 7
			? (myIndex < 7 ? (myRow!.xp - league[7].xp) : (league[6].xp - myRow!.xp))
			: 0
	);

	const mini = $derived([
		{ label: 'Position',    value: myIndex >= 0 ? `${myIndex + 1}e / ${league.length}` : '—' },
		{ label: 'XP totale',   value: myRow ? myRow.xp.toLocaleString('fr-FR') : '—' },
		{ label: myIndex >= 0 && myIndex < 7 ? 'Avance top 7' : 'Écart top 7', value: `${gapToTop7 >= 0 ? '+' : ''}${gapToTop7} XP` },
		{ label: 'Fin dans',    value: '2 j 14 h' },
	]);

	// Sparkline synthétique dérivée de l'XP (progression sur 7 jours)
	function fakeSpark(xp: number): number[] {
		const base = Math.max(xp * 0.4, 1);
		return [base, base * 1.1, base * 1.25, base * 1.4, base * 1.6, base * 1.8, xp * 0.92, xp];
	}

	async function challenge() {
		if (!$auth.token || !rival || busy) return;
		busy = true;
		try {
			await challengeRival($auth.token, rival.name);
			challengeMsg = `Défi lancé : dépasser @${rival.name} (+200 XP à la clé). Suivez-le dans vos quêtes.`;
			refreshStats();
		} catch (e) {
			challengeMsg = e instanceof Error ? e.message : 'Erreur lors du défi.';
		} finally {
			busy = false;
		}
	}

	const rewards = [
		{ range: 'Top 1',  desc: '5 000 XP + Badge Légende' },
		{ range: 'Top 3',  desc: '3 000 XP + Badge Master'  },
		{ range: 'Top 7',  desc: '1 500 XP + Promotion'     },
		{ range: '8–23',   desc: '500 XP · maintien'        },
		{ range: 'Bas 7',  desc: 'Relégation · −200 XP'     },
	];

	const W = 280, H = 130;
	const rivalPath = $derived(rival ? sparkPath(fakeSpark(rival.xp), W, H) : '');
	const mePath    = $derived(myRow ? sparkPath(fakeSpark(myRow.xp), W, H) : '');

	function sepColor(kind: string) {
		if (kind === 'promo') return '#86b3a4';
		if (kind === 'rel')   return '#d79c86';
		return '#766d60';
	}
</script>

<svelte:head><title>LIGUE · Collector.shop</title></svelte:head>

<!-- Hero -->
<section class="hero">
	<div>
		<Kicker>Ligue Or · Division 2 · Semaine 3/4</Kicker>
		<h1 class="hero-title">Or</h1>
		<p class="hero-lede">
			Le <strong style="color:#ece5da">top 7</strong> est promu en Diamant.
			{#if myIndex >= 0}
				Vous êtes <strong style="color:#86b3a4">{myIndex + 1}<sup>e</sup></strong> —
				{myIndex < 7 ? "gardez l'écart." : 'gagnez de l’XP pour remonter.'}
			{/if}
		</p>

		<div class="mini-stats">
			{#each mini as m}
				<div class="mini-card">
					<div class="mini-label">{m.label}</div>
					<div class="mini-val">{m.value}</div>
				</div>
			{/each}
		</div>

		<div class="tiers">
			{#each tiers as t}
				<div
					class="tier-chip"
					style="
						border-color: {t.active ? '#86b3a4' : 'rgba(236,229,218,0.10)'};
						background: {t.active ? 'rgba(255,255,255,0.05)' : 'transparent'};
						opacity: {t.done ? 1 : 0.45};
					"
				>
					<span class="tier-diamond" style="background:{t.active ? '#86b3a4' : '#766d60'}"></span>
					<span class="tier-label" style="color:{t.active ? '#86b3a4' : '#a39a8c'}">{t.label}</span>
				</div>
			{/each}
		</div>
	</div>

	<!-- Médaillon -->
	<div class="medallion-col">
		<div class="medallion">
			<span class="medallion-diamond"></span>
		</div>
		<div class="medallion-caption">Trophée · Saison 03</div>
	</div>
</section>

<!-- Board + Sidebar -->
<section class="board-section">
	<GPanel style="overflow:hidden">
		<div class="board-header-row">
			<Kicker>Classement Ligue Or · {league.length} chasseurs</Kicker>
			<span class="board-ts">à jour · {new Date().toLocaleTimeString('fr-FR', { hour: '2-digit', minute: '2-digit' })}</span>
		</div>
		<div class="board-cols">
			<span style="width:32px;text-align:center">#</span>
			<span style="flex:1">Chasseur</span>
			<span style="width:60px">Niv.</span>
			<span style="width:110px;text-align:center">7 derniers j.</span>
			<span style="width:80px;text-align:right">XP</span>
			<span style="width:52px;text-align:right">Δ</span>
		</div>
		{#if league.length === 0}
			<p class="board-empty">Chargement du classement…</p>
		{/if}
		{#each league as row, i (row.name)}
			<div
				class="board-row"
				class:board-me={row.me}
				style={row.me ? 'background:rgba(255,255,255,0.05);border:1px solid rgba(134,179,164,0.35)' : 'border:1px solid transparent'}
			>
				<span class="board-rank" style="color:{row.me ? '#86b3a4' : '#a39a8c'}">{i + 1}</span>
				<span class="board-name" style="font-weight:{row.me ? 600 : 400}">
					{#if row.me}
						<a href="/profil" class="board-me-link">{row.name} (vous)</a>
					{:else}
						@{row.name}
					{/if}
				</span>
				<span class="board-level">niv. {row.level}</span>
				<span class="board-spark">
					<GSpark values={fakeSpark(row.xp)} color={row.me ? '#86b3a4' : '#766d60'} w={100} h={22} dot={false} />
				</span>
				<span class="board-xp">{row.xp.toLocaleString('fr-FR')}</span>
				<span class="board-delta" style="color:{row.delta >= 0 ? '#86c099' : '#d79c86'}">
					{row.delta >= 0 ? '+' : ''}{row.delta}
				</span>
			</div>
			{#if i === 6 && league.length > 7}
				<div class="board-sep" style="color:{sepColor('promo')};border-color:{sepColor('promo')}66">
					Promotion Diamant · top 7
				</div>
			{/if}
		{/each}
	</GPanel>

	<div class="sidebar">
		<!-- Course XP -->
		<GPanel>
			<Kicker>Course XP · 7 derniers jours</Kicker>
			<div class="chart-wrap">
				<svg viewBox="0 0 {W} {H}" width="100%" height={H} preserveAspectRatio="none" style="display:block;margin-top:12px">
					{#each [0,1,2,3,4] as i}
						<line x1="0" y1={H*i/4} x2={W} y2={H*i/4} stroke="rgba(236,229,218,0.06)" stroke-width="0.5"/>
					{/each}
					<line x1="0" y1="32" x2={W} y2="32" stroke="#86b3a4" stroke-dasharray="3 4" stroke-width="1" opacity="0.6"/>
					<path d={rivalPath} stroke="#766d60" stroke-width="1.4" fill="none"/>
					<path d={mePath}    stroke="#86b3a4" stroke-width="2"   fill="none"/>
				</svg>
				<div class="chart-days">
					{#each ['L','M','M','J','V','S','D'] as d}
						<span>{d}</span>
					{/each}
				</div>
				<div class="chart-legend">
					<span class="legend-item"><span class="legend-line" style="background:#86b3a4"></span>vous · +120/j</span>
					<span class="legend-item"><span class="legend-line" style="background:#766d60"></span>rival · +50/j</span>
				</div>
			</div>
		</GPanel>

		<!-- Rival -->
		{#if rival && myRow}
			<GPanel>
				<Kicker>Rival de la semaine</Kicker>
				<div class="rival-row">
					<GAvatar initials={rival.name.slice(0, 2).toUpperCase()} size={50} square />
					<div class="rival-info">
						<div class="rival-name">@{rival.name}</div>
						<div class="rival-meta">niveau {rival.level} · {league.indexOf(rival) + 1}<sup>e</sup> place</div>
						<p class="rival-desc">
							{#if myRow.xp >= rival.xp}
								Vous menez <strong style="color:#86c099">+{myRow.xp - rival.xp} XP</strong>. Gardez le rythme.
							{:else}
								Il vous manque <strong style="color:#d79c86">{rival.xp - myRow.xp} XP</strong> pour le dépasser.
							{/if}
						</p>
					</div>
				</div>
				<div class="rival-btns">
					<button class="btn-primary" onclick={() => goto('/profil')}>Voir mon profil</button>
					<button class="btn-ghost" disabled={busy} onclick={challenge}>Défier</button>
				</div>
				{#if challengeMsg}
					<p class="challenge-msg">{challengeMsg}</p>
				{/if}
			</GPanel>
		{/if}

		<!-- Récompenses -->
		<GPanel>
			<Kicker>Récompenses de fin de ligue</Kicker>
			<div class="rewards">
				{#each rewards as r, i}
					<div class="reward-row" style="border-top:{i>0?'1px solid rgba(236,229,218,0.10)':'none'}">
						<span class="reward-range" style="color:{i===4?'#d79c86':'#86b3a4'}">{r.range}</span>
						<span class="reward-desc">{r.desc}</span>
					</div>
				{/each}
			</div>
		</GPanel>
	</div>
</section>

<style>
	/* ── Hero ── */
	.hero {
		display: grid;
		grid-template-columns: 1.5fr 0.6fr;
		gap: 30px;
		padding: 24px 0 20px;
		align-items: center;
	}
	@media (max-width: 768px) { .hero { grid-template-columns: 1fr; } }

	.hero-title {
		font-family: 'Newsreader', Georgia, serif;
		font-weight: 500;
		font-size: 64px;
		line-height: 1;
		margin: 8px 0 0;
		color: #ece5da;
	}
	.hero-lede {
		font-size: 14.5px;
		color: #a39a8c;
		margin-top: 12px;
		line-height: 1.55;
		max-width: 540px;
	}

	.mini-stats {
		display: flex;
		gap: 10px;
		flex-wrap: wrap;
		margin-top: 18px;
	}
	.mini-card {
		padding: 10px 14px;
		border-radius: 8px;
		border: 1px solid rgba(236,229,218,0.10);
		background: rgba(255,255,255,0.05);
		min-width: 110px;
	}
	.mini-label {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 10px;
		letter-spacing: 0.14em;
		text-transform: uppercase;
		color: #766d60;
	}
	.mini-val {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 20px;
		color: #ece5da;
		margin-top: 3px;
	}

	.tiers {
		display: flex;
		gap: 8px;
		margin-top: 20px;
		flex-wrap: wrap;
	}
	.tier-chip {
		flex: 1;
		padding: 10px;
		border-radius: 8px;
		border: 1px solid;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 5px;
		min-width: 60px;
	}
	.tier-diamond {
		width: 9px;
		height: 9px;
		border-radius: 2px;
		transform: rotate(45deg);
		display: inline-block;
	}
	.tier-label {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 10px;
		letter-spacing: 0.12em;
		text-transform: uppercase;
	}

	/* Médaillon */
	.medallion-col {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 14px;
	}
	.medallion {
		width: 130px;
		height: 130px;
		border-radius: 50%;
		border: 2px solid #86b3a4;
		background: rgba(255,255,255,0.05);
		box-shadow: 0 0 0 8px rgba(255,255,255,0.03);
		display: flex;
		align-items: center;
		justify-content: center;
	}
	.medallion-diamond {
		width: 40px;
		height: 40px;
		border-radius: 6px;
		transform: rotate(45deg);
		background: #86b3a4;
		opacity: 0.85;
	}
	.medallion-caption {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 10.5px;
		letter-spacing: 0.2em;
		text-transform: uppercase;
		color: #a39a8c;
	}

	/* ── Board ── */
	.board-section {
		display: grid;
		grid-template-columns: 1.6fr 1fr;
		gap: 14px;
		margin-top: 8px;
	}
	@media (max-width: 900px) { .board-section { grid-template-columns: 1fr; } }

	.board-header-row {
		display: flex;
		justify-content: space-between;
		align-items: baseline;
		margin-bottom: 12px;
	}
	.board-ts {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 11px;
		color: #766d60;
	}
	.board-cols {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 0 8px 8px;
		border-bottom: 1px solid rgba(236,229,218,0.10);
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 10px;
		letter-spacing: 0.12em;
		text-transform: uppercase;
		color: #766d60;
	}
	.board-row {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 9px 8px;
		border-radius: 7px;
		margin-bottom: 2px;
	}
	.board-empty {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 11px;
		color: #766d60;
		padding: 14px 8px;
	}
	.board-me-link { color: inherit; text-decoration: none; }
	.board-me-link:hover { text-decoration: underline; }
	.challenge-msg {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 12px;
		color: #86b3a4;
		margin-top: 10px;
		line-height: 1.5;
	}
	.board-sep {
		text-align: center;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 10px;
		letter-spacing: 0.18em;
		text-transform: uppercase;
		padding: 8px 0;
		margin: 3px 0;
		border-top: 1px dashed;
		border-bottom: 1px dashed;
	}
	.board-rank {
		width: 32px;
		text-align: center;
		font-family: 'Newsreader', Georgia, serif;
		font-size: 16px;
		flex-shrink: 0;
	}
	.board-name {
		flex: 1;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13px;
		color: #ece5da;
	}
	.board-level {
		width: 60px;
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 11.5px;
		color: #a39a8c;
		flex-shrink: 0;
	}
	.board-spark {
		width: 110px;
		display: flex;
		justify-content: center;
		flex-shrink: 0;
	}
	.board-xp {
		width: 80px;
		text-align: right;
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 13px;
		color: #ece5da;
		flex-shrink: 0;
	}
	.board-delta {
		width: 52px;
		text-align: right;
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 11.5px;
		flex-shrink: 0;
	}

	/* ── Sidebar ── */
	.sidebar { display: flex; flex-direction: column; gap: 14px; }

	.chart-wrap { margin-top: 0; }
	.chart-days {
		display: flex;
		justify-content: space-between;
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 10px;
		color: #766d60;
		margin-top: 6px;
	}
	.chart-legend {
		display: flex;
		flex-direction: column;
		gap: 5px;
		margin-top: 12px;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 12px;
		color: #a39a8c;
	}
	.legend-item { display: flex; align-items: center; gap: 8px; }
	.legend-line { width: 14px; height: 2px; border-radius: 2px; display: inline-block; flex-shrink: 0; }

	.rival-row {
		display: flex;
		gap: 14px;
		align-items: flex-start;
		margin-top: 10px;
	}
	.rival-info { flex: 1; min-width: 0; }
	.rival-name {
		font-family: 'Newsreader', Georgia, serif;
		font-size: 17px;
		color: #ece5da;
	}
	.rival-meta {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 11.5px;
		color: #a39a8c;
		margin-top: 2px;
	}
	.rival-desc {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 12.5px;
		color: #a39a8c;
		margin-top: 8px;
		line-height: 1.5;
	}
	.rival-btns {
		display: flex;
		gap: 8px;
		margin-top: 14px;
	}
	.btn-primary {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 12px;
		font-weight: 600;
		padding: 9px 16px;
		border-radius: 7px;
		border: none;
		cursor: pointer;
		background: #86b3a4;
		color: #191714;
		transition: filter 120ms;
	}
	.btn-primary:hover { filter: brightness(1.08); }
	.btn-ghost {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 12px;
		padding: 9px 16px;
		border-radius: 7px;
		border: 1px solid rgba(236,229,218,0.10);
		cursor: pointer;
		background: transparent;
		color: #a39a8c;
		transition: border-color 120ms, color 120ms;
	}
	.btn-ghost:hover { border-color: rgba(236,229,218,0.22); color: #ece5da; }

	.rewards { display: flex; flex-direction: column; margin-top: 8px; }
	.reward-row {
		display: flex;
		gap: 12px;
		align-items: baseline;
		padding: 8px 0;
	}
	.reward-range {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 11px;
		letter-spacing: 0.08em;
		width: 52px;
		flex-shrink: 0;
	}
	.reward-desc {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 12.5px;
		color: #ece5da;
	}
</style>
