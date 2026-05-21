<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import HoloPanel from '$lib/components/holo/HoloPanel.svelte';
	import HoloMeter from '$lib/components/holo/HoloMeter.svelte';
	import HoloEyebrow from '$lib/components/holo/HoloEyebrow.svelte';
	import HoloTitle from '$lib/components/holo/HoloTitle.svelte';

	const expiresAt = new Date(Date.now() + 4 * 3600_000 + 27 * 60_000 + 44_000);
	let remaining = $state(0);
	let timerId: ReturnType<typeof setInterval>;
	function tick() { remaining = Math.max(0, expiresAt.getTime() - Date.now()); }
	onMount(() => { tick(); timerId = setInterval(tick, 1000); });
	onDestroy(() => clearInterval(timerId));
	const hh = $derived(Math.floor(remaining / 3600_000));
	const mm = $derived(Math.floor((remaining % 3600_000) / 60_000));
	const ss = $derived(Math.floor((remaining % 60_000) / 1000));
	const pad = (n: number) => String(n).padStart(2, '0');

	const subquests = [
		{ icon: '◆', label: 'Consulter 5 articles', progress: 3, target: 5, xp: 80,  done: false },
		{ icon: '★', label: 'Noter une pièce',       progress: 1, target: 1, xp: 50,  done: true  },
		{ icon: '◈', label: 'Ajouter à la wishlist', progress: 0, target: 3, xp: 60,  done: false },
		{ icon: '●', label: 'Partager une entrée',   progress: 0, target: 1, xp: 40,  done: false },
	];
	const weekly = [
		{ label: 'Compléter 5 missions',     current: 3, target: 5 },
		{ label: 'Acheter ou vendre 2 pièces',current: 1, target: 2 },
		{ label: 'Rédiger 3 notes journal',   current: 2, target: 3 },
		{ label: 'Streak 7j consécutifs',     current: 5, target: 7 },
	];
	const weekDays = ['L','M','M','J','V','S','D'];
	const weekDone = [true,true,true,true,true,false,false];
	const todayIdx = 5;
	const seasonNodes = [
		{ label: 'Départ',  done: true,  current: false },
		{ label: 'Niv. 10', done: true,  current: false },
		{ label: 'Niv. 12', done: true,  current: false },
		{ label: 'Niv. 14', done: false, current: true  },
		{ label: 'Niv. 16', done: false, current: false },
		{ label: 'Fin S3',  done: false, current: false },
	];
</script>

<svelte:head><title>QUÊTES · Collector.shop</title></svelte:head>

<div class="layout-top">
	<div class="orb-col">
		<div class="orb-wrap">
			<div class="orb-ring" aria-hidden="true"></div>
			<div class="orb"><span class="orb-glyph">◆</span></div>
		</div>
		<div class="speech-bubble">
			<div style="font-family:'JetBrains Mono',monospace;font-size:9px;letter-spacing:0.28em;color:#a8c8e4;margin-bottom:8px;text-transform:uppercase">transmission stable</div>
			<p style="font-size:13px;color:#e8eaed;line-height:1.6;margin:0 0 12px">
				Ta mission du jour expire dans <strong style="color:#cbd5e0">{pad(hh)}h{pad(mm)}</strong>. Bonne chance, collectionneur.
			</p>
			<div style="display:flex;justify-content:space-between;align-items:center">
				<span style="font-family:'JetBrains Mono',monospace;font-size:9px;color:#5a606a;letter-spacing:0.2em">transmission stable</span>
				<button style="padding:4px 12px;border:1px solid rgba(168,200,228,0.3);border-radius:6px;background:transparent;color:#a8c8e4;font-size:10px;font-weight:600;cursor:pointer">répondre</button>
			</div>
		</div>
	</div>

	<HoloPanel glow="#a8c8e4" style="flex:1.4">
		<HoloEyebrow color="#a8c8e4">MISSION DU JOUR</HoloEyebrow>
		<div class="countdown">
			<span class="cd-seg">{pad(hh)}<small>H</small></span>
			<span class="cd-sep">:</span>
			<span class="cd-seg">{pad(mm)}<small>M</small></span>
			<span class="cd-sep">:</span>
			<span class="cd-seg">{pad(ss)}<small>S</small></span>
		</div>
		<HoloTitle size={30}>Acquérir une pièce TCG gradée</HoloTitle>
		<div class="reward-box">
			<span style="font-family:'Major Mono Display',monospace;font-size:20px;color:#cbd5e0">◆</span>
			<span style="font-family:'JetBrains Mono',monospace;font-size:36px;font-weight:700;color:#cbd5e0;line-height:1">500</span>
			<span style="font-size:10px;letter-spacing:0.2em;color:#8a909a;font-family:'JetBrains Mono',monospace">XP · RÉCOMPENSE</span>
		</div>
		<div style="display:flex;justify-content:space-between;align-items:center;margin:14px 0 8px">
			<span style="font-size:12px;color:#8a909a">Progression</span>
			<span style="font-size:12px;font-family:'JetBrains Mono',monospace;color:#a8c8e4">66%</span>
		</div>
		<HoloMeter value={66} />
		<div style="display:flex;gap:10px;align-items:center;margin-top:16px">
			<button class="cta-quest">DÉMARRER LA MISSION</button>
			<button style="color:#5a606a;font-size:11px;background:none;border:none;cursor:pointer">passer</button>
		</div>
	</HoloPanel>
</div>

<div class="layout-bottom">
	<div style="display:flex;flex-direction:column;gap:14px;flex:1.4">
		<HoloPanel>
			<HoloEyebrow>SOUS-QUÊTES DU JOUR</HoloEyebrow>
			{#each subquests as q}
				<div class="subquest-row" class:sq-done={q.done}>
					<span class="sq-icon">{q.icon}</span>
					<span class="sq-label" class:line-through={q.done}>{q.label}</span>
					<span style="font-family:'JetBrains Mono',monospace;font-size:11px;color:#a8c8e4;flex-shrink:0">{q.progress}/{q.target}</span>
					<span style="font-family:'JetBrains Mono',monospace;font-size:14px;color:#cbd5e0;flex-shrink:0;font-weight:600">+{q.xp}</span>
					{#if q.done}<span style="color:#7cd9a0;flex-shrink:0">✓</span>{/if}
				</div>
			{/each}
		</HoloPanel>
		<HoloPanel>
			<HoloEyebrow color="#a8c8e4">QUÊTES HEBDOMADAIRES</HoloEyebrow>
			<div class="weekly-grid">
				{#each weekly as w}
					<div>
						<div style="display:flex;justify-content:space-between;margin-bottom:6px">
							<span style="font-size:12px;color:#e8eaed">{w.label}</span>
							<span style="font-family:'JetBrains Mono',monospace;font-size:11px;color:#8a909a">{w.current}/{w.target}</span>
						</div>
						<HoloMeter value={Math.round(w.current/w.target*100)} height={6} />
					</div>
				{/each}
			</div>
		</HoloPanel>
	</div>

	<div style="display:flex;flex-direction:column;gap:14px;flex:1">
		<HoloPanel glow="#a8c8e4">
			<HoloEyebrow color="#a8c8e4">STREAK</HoloEyebrow>
			<div class="streak-counter">×47</div>
			<div class="week-grid">
				{#each weekDays as day, i}
					<div class="week-day"
						class:week-done={weekDone[i]}
						class:week-today={i === todayIdx}
					>
						<span style="font-size:11px;font-weight:700;font-family:'JetBrains Mono',monospace;color:{weekDone[i] ? '#0e1014' : i===todayIdx ? '#a8c8e4' : '#5a606a'}">{day}</span>
					</div>
				{/each}
			</div>
		</HoloPanel>
		<HoloPanel>
			<HoloEyebrow>PARCOURS SAISONNIER</HoloEyebrow>
			<div class="season-track">
				{#each seasonNodes as node, i}
					<div style="display:flex;align-items:center;gap:10px">
						<div class="season-node"
							style={node.done ? 'border-color:#a8c8e4;background:rgba(168,200,228,0.12);color:#a8c8e4' :
								node.current ? 'border-color:#a8c8e4;color:#a8c8e4;box-shadow:0 0 10px rgba(168,200,228,0.3);animation:holoPulse 1.4s ease-in-out infinite' :
								'border-color:rgba(255,255,255,0.07);color:#5a606a;opacity:0.5'}
						>{node.done ? '✓' : node.current ? '◆' : '○'}</div>
						<span style="font-size:12px;color:{node.current ? '#a8c8e4' : node.done ? '#e8eaed' : '#5a606a'}">{node.label}</span>
					</div>
					{#if i < seasonNodes.length - 1}
						<div style="width:1px;height:18px;border-left:1px dashed {node.done ? 'rgba(168,200,228,0.3)' : 'rgba(255,255,255,0.07)'};margin-left:14px"></div>
					{/if}
				{/each}
			</div>
		</HoloPanel>
	</div>
</div>

<style>
	.layout-top { display:grid;grid-template-columns:0.9fr 1.4fr;gap:18px;margin-bottom:18px;align-items:start; }
	.layout-bottom { display:grid;grid-template-columns:1.4fr 1fr;gap:18px; }
	@media (max-width:900px) { .layout-top,.layout-bottom { grid-template-columns:1fr; } }

	.orb-col { display:flex;flex-direction:column;align-items:center;gap:16px; }
	.orb-wrap { position:relative;width:160px;height:160px; }
	.orb-ring { position:absolute;inset:-10px;border-radius:50%;background:conic-gradient(#a8c8e4,#6a7280,#cbd5e0,#8a909a,#a8c8e4);background-size:200% 100%;animation:holoSweep 6s linear infinite;filter:blur(8px);opacity:0.5; }
	.orb { position:relative;width:160px;height:160px;border-radius:50%;background:radial-gradient(60% 60% at 35% 30%,rgba(255,255,255,0.9) 0%,oklch(0.55 0.08 215) 35%,oklch(0.25 0.06 215) 70%,#0e1014 100%);display:flex;align-items:center;justify-content:center;animation:holoFloat 4s ease-in-out infinite;box-shadow:0 0 40px rgba(168,200,228,0.15); }
	.orb-glyph { font-family:'Major Mono Display',monospace;font-size:60px;color:#fff;text-shadow:0 0 20px rgba(168,200,228,0.6); }
	.speech-bubble { width:100%;background:linear-gradient(180deg,#181a20,#0d0f13);border:1px solid rgba(255,255,255,0.07);border-radius:12px;padding:14px; }

	.countdown { display:flex;align-items:baseline;gap:4px;margin-bottom:12px; }
	.cd-seg { font-family:'JetBrains Mono',monospace;font-size:42px;font-weight:700;line-height:1;color:#a8c8e4; }
	.cd-seg small { font-size:18px;opacity:0.7; }
	.cd-sep { font-family:'JetBrains Mono',monospace;font-size:30px;color:#5a606a; }

	.reward-box { display:inline-flex;align-items:center;gap:8px;margin:14px 0;padding:10px 16px;background:rgba(203,213,224,0.06);border:1px solid rgba(203,213,224,0.2);border-radius:8px; }
	.cta-quest { flex:1;padding:12px;border-radius:8px;border:none;background:linear-gradient(135deg,#a8c8e4,#6a7280);color:#0e1014;font-size:11px;font-weight:700;letter-spacing:0.2em;cursor:pointer; }

	.subquest-row { display:flex;align-items:center;gap:10px;padding:10px 12px;border-radius:8px;border:1px solid rgba(255,255,255,0.06);background:rgba(255,255,255,0.02);margin-bottom:8px; }
	.sq-done { opacity:0.45; }
	.sq-icon { width:28px;height:28px;border:1px solid rgba(168,200,228,0.25);border-radius:6px;display:flex;align-items:center;justify-content:center;font-family:'Major Mono Display',monospace;font-size:14px;flex-shrink:0;color:#a8c8e4;background:rgba(168,200,228,0.06); }
	.sq-label { flex:1;font-size:13px;color:#e8eaed; }
	.line-through { text-decoration:line-through; }
	.weekly-grid { display:grid;grid-template-columns:1fr 1fr;gap:12px; }

	.streak-counter { font-family:'Major Mono Display',monospace;font-size:72px;line-height:1;color:#a8c8e4;text-shadow:0 0 22px rgba(168,200,228,0.3);text-align:center;margin-bottom:16px; }
	.week-grid { display:grid;grid-template-columns:repeat(7,1fr);gap:6px; }
	.week-day { aspect-ratio:1;border-radius:6px;border:1px solid rgba(255,255,255,0.07);display:flex;align-items:center;justify-content:center;background:rgba(255,255,255,0.02); }
	.week-done { background:linear-gradient(135deg,#a8c8e4,#6a7280);border-color:transparent; }
	.week-today { border-style:dashed;border-color:#a8c8e4; }

	.season-track { display:flex;flex-direction:column; }
	.season-node { width:28px;height:28px;border-radius:50%;border:1px solid;flex-shrink:0;display:flex;align-items:center;justify-content:center;font-size:12px; }
</style>
