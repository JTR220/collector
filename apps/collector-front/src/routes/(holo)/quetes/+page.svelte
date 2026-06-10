<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth, isAuthenticated } from '$lib/stores/auth';
	import { playerStats, refreshStats } from '$lib/stores/stats';
	import { fetchMyQuests, progressQuest, skipQuest, type Quest } from '$lib/api/engagement';
	import { fetchArticles } from '$lib/api/catalog';
	import GPanel from '$lib/components/galerie/GPanel.svelte';
	import GMeter from '$lib/components/galerie/GMeter.svelte';
	import GAvatar from '$lib/components/galerie/GAvatar.svelte';
	import Kicker from '$lib/components/galerie/Kicker.svelte';

	let quests = $state<Quest[]>([]);
	let eligibleCount = $state(0);
	let busy = $state(false);
	let toast = $state<string | null>(null);

	onMount(async () => {
		if (!$isAuthenticated || !$auth.token) {
			goto('/login');
			return;
		}
		try {
			quests = await fetchMyQuests($auth.token);
			const articles = await fetchArticles();
			eligibleCount = articles.filter((a) => a.category.name === 'TCG' && a.prix <= 300).length;
		} catch (e) {
			console.error(e);
		}
	});

	const mission = $derived(quests.find((q) => q.kind === 'mission') ?? null);
	const subs = $derived(quests.filter((q) => q.kind === 'daily'));
	const weekly = $derived(quests.filter((q) => q.kind === 'weekly'));
	const subsDone = $derived(subs.filter((q) => q.done).length);

	async function advance(quest: Quest) {
		if (!$auth.token || quest.done || busy) return;
		busy = true;
		try {
			const updated = await progressQuest($auth.token, quest.ID);
			quests = quests.map((q) => (q.ID === updated.ID ? updated : q));
			if (updated.done) {
				toast = `« ${updated.title} » complétée · +${updated.xp} XP`;
				refreshStats();
			}
		} catch (e) {
			toast = e instanceof Error ? e.message : 'Erreur quête.';
		} finally {
			busy = false;
		}
	}

	async function skipMission() {
		if (!$auth.token || !mission || mission.done || busy) return;
		busy = true;
		try {
			const res = await skipQuest($auth.token, mission.ID);
			quests = quests.map((q) => (q.ID === res.quest.ID ? res.quest : q));
			toast = `Mission passée · −30 gems (reste ${res.gems})`;
			refreshStats();
		} catch (e) {
			toast = e instanceof Error ? e.message : 'Erreur lors du passage.';
		} finally {
			busy = false;
		}
	}

	// Compte à rebours jusqu'à minuit (réinitialisation des quêtes journalières)
	const expiresAt = (() => {
		const d = new Date();
		d.setHours(24, 0, 0, 0);
		return d;
	})();
	let remaining = $state(0);
	let timerId: ReturnType<typeof setInterval>;
	function tick() {
		remaining = Math.max(0, expiresAt.getTime() - Date.now());
	}
	onMount(() => {
		tick();
		timerId = setInterval(tick, 1000);
	});
	onDestroy(() => clearInterval(timerId));
	const hh = $derived(Math.floor(remaining / 3_600_000));
	const mm = $derived(Math.floor((remaining % 3_600_000) / 60_000));
	const ss = $derived(Math.floor((remaining % 60_000) / 1_000));
	const pad = (n: number) => String(n).padStart(2, '0');

	const streak = $derived($playerStats?.streak ?? 0);
	const weekDays = ['L', 'M', 'M', 'J', 'V', 'S', 'D'];
	const todayIdx = (new Date().getDay() + 6) % 7;
	const weekDone = $derived(weekDays.map((_, i) => i < todayIdx || (i <= todayIdx && streak > 0)));
	const path = [
		{ s: 'done', t: 'Premier achat', d: '12.03' },
		{ s: 'done', t: 'Ligue Bronze', d: '04.04' },
		{ s: 'done', t: '5 pièces postées', d: '28.04' },
		{ s: 'cur', t: 'Compléter Base Set', d: '3 / 8 cartes' },
		{ s: 'lock', t: "Trader d'élite", d: 'niveau 15' },
		{ s: 'lock', t: 'Mythic Hunter', d: 'niveau 18' }
	];
</script>

<svelte:head><title>QUÊTES · Collector.shop</title></svelte:head>

<!-- Hero : assistant + mission du jour -->
<section class="hero-grid">
	<!-- Conseiller -->
	<GPanel style="display:flex;flex-direction:column;gap:14px">
		<div class="advisor-top">
			<GAvatar initials="◆" size={44} square />
			<div>
				<div class="advisor-name">Votre conseiller</div>
				<div class="advisor-ts">journal {pad(hh)}:{pad(mm)}</div>
			</div>
		</div>
		<p class="advisor-msg">
			Plus que <em>3 cartes</em> et votre Base Set est complet, Nina. Trois exemplaires éligibles sont
			passés sous les 300 € aujourd'hui.
		</p>
		<div class="advisor-footer">
			<span class="online-dot"></span>
			<span class="online-label">en ligne</span>
			<span class="advisor-sep">·</span>
			<button class="advisor-reply" onclick={() => goto('/journal')}>répondre</button>
		</div>
	</GPanel>

	<!-- Mission du jour -->
	<GPanel style="padding:22px">
		<div class="mission-body">
			<div class="mission-left">
				<Kicker>Mission du jour — expire dans</Kicker>
				<div class="countdown">
					<span class="cd-seg">{pad(hh)}<small>h</small></span>
					<span class="cd-sep">:</span>
					<span class="cd-seg">{pad(mm)}<small>min</small></span>
					<span class="cd-sep">:</span>
					<span class="cd-seg">{pad(ss)}<small>s</small></span>
				</div>
				<h2 class="mission-title">{mission?.title ?? 'Trouve une holo sous 300 €'}</h2>
				<p class="mission-desc">
					Filtre actif → catégorie TCG · moins de 300 €.
					<strong style="color:#ece5da"
						>{eligibleCount} pièce{eligibleCount > 1 ? 's' : ''} éligible{eligibleCount > 1
							? 's'
							: ''}.</strong
					>
				</p>
			</div>
			<div class="reward-box">
				<Kicker>Récompense</Kicker>
				<div class="reward-val">+150</div>
				<div class="reward-unit">gems</div>
				<div class="reward-bonus">+ 1 booster</div>
			</div>
		</div>
		<div class="mission-progress">
			<div class="progress-header">
				<span>{mission?.done ? 'Mission accomplie' : 'Mission en cours'}</span>
				<span style="color:#86b3a4">{mission?.done ? '100 %' : '0 %'}</span>
			</div>
			<GMeter value={mission?.done ? 100 : 0} height={6} />
		</div>
		<div class="mission-ctas">
			<button
				class="btn-primary"
				onclick={() => {
					if (mission && !mission.done) advance(mission);
					goto('/?cat=TCG&max=300');
				}}
			>
				Voir les {eligibleCount} pièce{eligibleCount > 1 ? 's' : ''} →
			</button>
			{#if mission && !mission.done}
				<button class="btn-skip" disabled={busy} onclick={skipMission}>passer (−30 gems)</button>
			{/if}
		</div>
		{#if toast}
			<p class="quest-toast">{toast}</p>
		{/if}
	</GPanel>
</section>

<!-- Sous-quêtes + Série -->
<section class="bottom-grid">
	<GPanel>
		<div class="sq-header">
			<Kicker>Sous-quêtes · {subsDone} / {subs.length} complétées</Kicker>
			<span class="sq-reset">réinit. 00:00</span>
		</div>
		<div class="sq-list">
			{#each subs as q (q.ID)}
				<button
					class="sq-row"
					style="opacity:{q.done ? 0.55 : 1};background:{q.done
						? 'transparent'
						: 'rgba(255,255,255,0.03)'}"
					disabled={q.done || busy}
					onclick={() => advance(q)}
					title={q.done ? 'Complétée' : 'Avancer cette quête'}
				>
					<span
						class="sq-check"
						style="
							border-color:{q.done ? '#86b3a4' : 'rgba(236,229,218,0.16)'};
							background:{q.done ? '#86b3a4' : 'transparent'};
							color:#191714
						">{q.done ? '✓' : ''}</span
					>
					<span class="sq-label" style="text-decoration:{q.done ? 'line-through' : 'none'}"
						>{q.title}</span
					>
					{#if q.target > 1}<span class="sq-prog">{q.progress}/{q.target}</span>{/if}
					<span class="sq-xp">+{q.xp} XP</span>
				</button>
			{/each}
			{#if subs.length === 0}
				<p class="sq-empty">Chargement des quêtes…</p>
			{/if}
		</div>

		<div class="weekly-section">
			<Kicker>Quêtes hebdomadaires</Kicker>
			<div class="weekly-grid">
				{#each weekly as w (w.ID)}
					<button
						class="weekly-card"
						disabled={w.done || busy}
						onclick={() => advance(w)}
						title={w.done ? 'Complétée' : 'Avancer cette quête'}
					>
						<div class="weekly-top">
							<span class="weekly-name">{w.title}</span>
							<span class="weekly-prog" style="color:{w.done ? '#86c099' : '#a39a8c'}"
								>{w.progress}/{w.target}</span
							>
						</div>
						<GMeter
							value={Math.round((w.progress / w.target) * 100)}
							color={w.done ? '#86c099' : '#86b3a4'}
						/>
					</button>
				{/each}
			</div>
		</div>
	</GPanel>

	<GPanel>
		<div class="streak-header">
			<Kicker>Série · {streak} jour{streak > 1 ? 's' : ''}</Kicker>
			<span class="streak-record">record · {Math.max(streak, 52)}</span>
		</div>
		<div class="streak-count">
			{streak}<span class="streak-unit"> jour{streak > 1 ? 's' : ''}</span>
		</div>
		<div class="week-grid">
			{#each weekDays as day, i}
				<div
					class="week-day"
					style="
						background:{weekDone[i] ? '#86b3a4' : 'transparent'};
						border:{i === todayIdx ? '1.5px dashed #86b3a4' : '1px solid rgba(236,229,218,0.10)'};
					"
				>
					<span
						class="week-label"
						style="color:{weekDone[i] ? '#191714' : i === todayIdx ? '#86b3a4' : '#a39a8c'}"
						>{day}</span
					>
					<span
						class="week-dot"
						style="color:{weekDone[i] ? '#191714' : i === todayIdx ? '#86b3a4' : '#766d60'}"
					>
						{weekDone[i] ? '●' : i === todayIdx ? '○' : '·'}
					</span>
				</div>
			{/each}
		</div>

		<div class="path-section">
			<Kicker>Parcours · saison Pokémon</Kicker>
			<div class="path-list">
				{#each path as node, i}
					<div class="path-item">
						<div class="path-node-wrap">
							<div
								class="path-node"
								style="
									border-color:{node.s === 'cur'
									? '#86b3a4'
									: node.s === 'done'
										? 'rgba(236,229,218,0.16)'
										: 'rgba(236,229,218,0.10)'};
									background:{node.s === 'cur' ? 'rgba(255,255,255,0.05)' : 'transparent'};
									color:{node.s === 'cur' ? '#86b3a4' : node.s === 'done' ? '#a39a8c' : '#766d60'};
								"
							>
								{node.s === 'done' ? '✓' : node.s === 'cur' ? '●' : '·'}
							</div>
							{#if i < path.length - 1}
								<div class="path-connector"></div>
							{/if}
						</div>
						<div class="path-text" style="opacity:{node.s === 'lock' ? 0.5 : 1}">
							<span
								class="path-title"
								style="font-weight:{node.s === 'cur' ? 600 : 400};color:{node.s === 'cur'
									? '#ece5da'
									: '#ece5da'}">{node.t}</span
							>
							<span class="path-date">{node.d}</span>
						</div>
					</div>
				{/each}
			</div>
		</div>
	</GPanel>
</section>

<style>
	/* ── Hero ── */
	.hero-grid {
		display: grid;
		grid-template-columns: 0.85fr 1.45fr;
		gap: 14px;
		padding: 4px 0 14px;
	}
	@media (max-width: 900px) {
		.hero-grid {
			grid-template-columns: 1fr;
		}
	}

	.advisor-top {
		display: flex;
		align-items: center;
		gap: 12px;
	}
	.advisor-name {
		font-family: 'Newsreader', Georgia, serif;
		font-size: 16px;
		color: #ece5da;
		white-space: nowrap;
	}
	.advisor-ts {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 10.5px;
		color: #766d60;
		margin-top: 2px;
	}
	.advisor-msg {
		font-family: 'Newsreader', Georgia, serif;
		font-size: 16px;
		line-height: 1.55;
		color: #ece5da;
		margin: 0;
	}
	.advisor-msg em {
		font-style: italic;
		font-weight: 500;
	}
	.advisor-footer {
		display: flex;
		align-items: center;
		gap: 8px;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 12px;
		color: #a39a8c;
		margin-top: auto;
	}
	.online-dot {
		width: 6px;
		height: 6px;
		border-radius: 50%;
		background: #86c099;
		display: inline-block;
	}
	.online-label {
		color: #a39a8c;
	}
	.advisor-sep {
		color: #766d60;
	}
	.advisor-reply {
		background: none;
		border: none;
		padding: 0;
		color: #86b3a4;
		cursor: pointer;
		font-family: inherit;
		font-size: inherit;
	}
	.btn-skip {
		background: none;
		border: none;
		padding: 0;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 12px;
		color: #a39a8c;
		cursor: pointer;
	}

	.mission-body {
		display: flex;
		justify-content: space-between;
		gap: 24px;
	}
	.mission-left {
		flex: 1;
	}
	.countdown {
		display: flex;
		align-items: baseline;
		gap: 6px;
		margin: 10px 0 14px;
	}
	.cd-seg {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 30px;
		color: #ece5da;
		line-height: 1;
	}
	.cd-seg small {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 11px;
		color: #766d60;
		margin-left: 2px;
	}
	.cd-sep {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 22px;
		color: #766d60;
	}
	.mission-title {
		font-family: 'Newsreader', Georgia, serif;
		font-weight: 500;
		font-size: 30px;
		line-height: 1.1;
		margin: 0;
		color: #ece5da;
	}
	.mission-desc {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13.5px;
		color: #a39a8c;
		margin-top: 10px;
		max-width: 480px;
		line-height: 1.5;
	}

	.reward-box {
		text-align: center;
		padding: 14px 18px;
		border-radius: 8px;
		border: 1px solid rgba(236, 229, 218, 0.1);
		background: rgba(255, 255, 255, 0.05);
		flex-shrink: 0;
		align-self: flex-start;
	}
	.reward-val {
		font-family: 'Newsreader', Georgia, serif;
		font-size: 34px;
		line-height: 1;
		margin-top: 6px;
		color: #ece5da;
	}
	.reward-unit {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 10.5px;
		letter-spacing: 0.16em;
		text-transform: uppercase;
		color: #a39a8c;
	}
	.reward-bonus {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 11px;
		color: #86b3a4;
		margin-top: 8px;
	}

	.mission-progress {
		margin-top: 16px;
	}
	.progress-header {
		display: flex;
		justify-content: space-between;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 11.5px;
		color: #a39a8c;
		margin-bottom: 6px;
	}
	.mission-ctas {
		display: flex;
		align-items: center;
		gap: 16px;
		margin-top: 18px;
	}
	.btn-primary {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 12.5px;
		font-weight: 600;
		padding: 11px 20px;
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

	/* ── Bottom ── */
	.bottom-grid {
		display: grid;
		grid-template-columns: 1.4fr 1fr;
		gap: 14px;
	}
	@media (max-width: 900px) {
		.bottom-grid {
			grid-template-columns: 1fr;
		}
	}

	.sq-header {
		display: flex;
		justify-content: space-between;
		align-items: baseline;
		margin-bottom: 12px;
	}
	.sq-reset {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 11px;
		color: #766d60;
	}
	.sq-list {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.sq-row {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 11px 13px;
		border-radius: 8px;
		border: 1px solid rgba(236, 229, 218, 0.1);
		cursor: pointer;
		width: 100%;
		text-align: left;
		font: inherit;
		color: inherit;
		transition: border-color 120ms;
	}
	.sq-row:hover:not(:disabled) {
		border-color: rgba(134, 179, 164, 0.4);
	}
	.sq-row:disabled {
		cursor: default;
	}
	.sq-empty {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 11px;
		color: #766d60;
	}
	.quest-toast {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 12.5px;
		color: #86b3a4;
		margin-top: 12px;
	}
	.sq-check {
		width: 18px;
		height: 18px;
		border-radius: 4px;
		flex-shrink: 0;
		border: 1.5px solid;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 12px;
	}
	.sq-label {
		flex: 1;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13px;
		color: #ece5da;
	}
	.sq-prog {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 11px;
		color: #a39a8c;
		flex-shrink: 0;
	}
	.sq-xp {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 12px;
		color: #86b3a4;
		flex-shrink: 0;
	}

	.weekly-section {
		margin-top: 18px;
	}
	.weekly-grid {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 8px;
		margin-top: 10px;
	}
	.weekly-card {
		padding: 11px 13px;
		border-radius: 8px;
		border: 1px solid rgba(236, 229, 218, 0.1);
		background: rgba(255, 255, 255, 0.03);
		cursor: pointer;
		text-align: left;
		font: inherit;
		color: inherit;
		transition: border-color 120ms;
	}
	.weekly-card:hover:not(:disabled) {
		border-color: rgba(134, 179, 164, 0.4);
	}
	.weekly-card:disabled {
		cursor: default;
	}
	.weekly-top {
		display: flex;
		justify-content: space-between;
		margin-bottom: 7px;
	}
	.weekly-name {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 12.5px;
		font-weight: 600;
		color: #ece5da;
	}
	.weekly-prog {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 11px;
	}

	/* Série */
	.streak-header {
		display: flex;
		justify-content: space-between;
		align-items: baseline;
	}
	.streak-record {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 11px;
		color: #766d60;
	}
	.streak-count {
		font-family: 'Newsreader', Georgia, serif;
		font-size: 66px;
		line-height: 1;
		margin-top: 6px;
		color: #ece5da;
	}
	.streak-unit {
		font-size: 26px;
		color: #a39a8c;
	}

	.week-grid {
		display: grid;
		grid-template-columns: repeat(7, 1fr);
		gap: 6px;
		margin-top: 14px;
	}
	.week-day {
		aspect-ratio: 1;
		border-radius: 7px;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 2px;
	}
	.week-label {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 9.5px;
		letter-spacing: 0.1em;
	}
	.week-dot {
		font-size: 13px;
	}

	.path-section {
		margin-top: 18px;
	}
	.path-list {
		display: flex;
		flex-direction: column;
		margin-top: 10px;
	}
	.path-item {
		display: flex;
		gap: 12px;
		align-items: flex-start;
	}
	.path-node-wrap {
		display: flex;
		flex-direction: column;
		align-items: center;
	}
	.path-node {
		width: 30px;
		height: 30px;
		border-radius: 50%;
		border: 1.5px solid;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 12px;
		flex-shrink: 0;
	}
	.path-connector {
		width: 1px;
		height: 14px;
		background: rgba(236, 229, 218, 0.1);
	}
	.path-text {
		flex: 1;
		padding: 4px 0 10px;
	}
	.path-title {
		display: block;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13px;
		color: #ece5da;
	}
	.path-date {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 10.5px;
		color: #766d60;
		margin-top: 1px;
		display: block;
	}
</style>
