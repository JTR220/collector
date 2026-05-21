<script lang="ts">
	import { sparkPath } from '$lib/utils/format';
	import HoloPanel from '$lib/components/holo/HoloPanel.svelte';
	import HoloMeter from '$lib/components/holo/HoloMeter.svelte';
	import HoloEyebrow from '$lib/components/holo/HoloEyebrow.svelte';
	import HoloTitle from '$lib/components/holo/HoloTitle.svelte';

	const tiers = [
		{ label: 'BRONZE',  active: false, done: true  },
		{ label: 'ARGENT',  active: false, done: true  },
		{ label: 'OR',      active: true,  done: true  },
		{ label: 'DIAMANT', active: false, done: false },
		{ label: 'MASTER',  active: false, done: false },
		{ label: 'LÉGENDE', active: false, done: false },
	];

	type Player = { rank: number; name: string; level: number; spark: number[]; weekXp: number; delta: number; me?: boolean };
	const players: Player[] = [
		{ rank:1,  name:'holo_king',      level:18, spark:[800,900,950,1100,1200,1350,1480,1600], weekXp:1600, delta:12.4 },
		{ rank:2,  name:'pack_ripper',    level:16, spark:[700,750,800,900,1000,1100,1250,1420],  weekXp:1420, delta:8.1  },
		{ rank:3,  name:'arcade_twin',    level:15, spark:[600,680,750,820,880,950,1100,1310],    weekXp:1310, delta:5.3  },
		{ rank:4,  name:'kanto_archive',  level:12, spark:[500,560,620,700,800,960,1100,1240],    weekXp:1240, delta:3.8, me:true },
		{ rank:5,  name:'volt_tacticien', level:14, spark:[600,640,680,720,800,900,1050,1190],    weekXp:1190, delta:2.1  },
		{ rank:6,  name:'neon_ranger',    level:11, spark:[400,460,520,580,650,750,900,1050],     weekXp:1050, delta:1.4  },
		{ rank:7,  name:'dust_seeker',    level:10, spark:[350,400,450,500,580,680,800,980],      weekXp:980,  delta:-1.2 },
		{ rank:8,  name:'chrome_fox',     level:9,  spark:[300,350,380,420,480,560,680,820],      weekXp:820,  delta:-2.4 },
		{ rank:9,  name:'rift_waltz',     level:8,  spark:[280,300,330,360,400,480,580,720],      weekXp:720,  delta:-4.1 },
		{ rank:10, name:'static_mono',    level:7,  spark:[200,220,250,280,320,380,460,560],      weekXp:560,  delta:-8.2 },
		{ rank:11, name:'low_orbit',      level:6,  spark:[150,180,200,230,260,300,380,480],      weekXp:480,  delta:-11.5},
	];

	const rewards = [
		{ label: 'TOP 1',  reward: '5 000 XP + Badge Légende' },
		{ label: 'TOP 3',  reward: '3 000 XP + Badge Master'  },
		{ label: 'TOP 7',  reward: '1 500 XP + Promotion'     },
		{ label: '8–23',   reward: 'Zone sûre · 500 XP'       },
		{ label: 'BAS 7',  reward: 'Relégation · -200 XP'     },
	];

	const W = 300, H = 140;
	const rivalSpark = [600,640,680,720,800,900,1050,1190];
	const meSpark    = [500,560,620,700,800,960,1100,1240];
	const rivalPath  = sparkPath(rivalSpark, W, H);
	const mePath     = sparkPath(meSpark, W, H);
	const threshold  = H - (H-8) * (980 - Math.min(...meSpark)) / (Math.max(...meSpark) - Math.min(...meSpark));
</script>

<svelte:head><title>LIGUE · Collector.shop</title></svelte:head>

<div class="hero-row">
	<div>
		<HoloEyebrow color="#a8c8e4">SAISON 03 · SEMAINE 3</HoloEyebrow>
		<HoloTitle size={40}>LIGUE OR</HoloTitle>
		<p style="font-size:13px;color:#8a909a;margin:10px 0 20px">Division II · Rang #4 · 1 240 XP cette semaine</p>
		<div class="tier-row">
			{#each tiers as t}
				<div class="tier-chip" style={t.active
					? 'border-color:#a8c8e4;background:rgba(168,200,228,0.12);color:#a8c8e4;box-shadow:0 0 14px rgba(168,200,228,0.15)'
					: t.done ? 'color:#8a909a;border-color:rgba(255,255,255,0.12)' : 'opacity:0.3;border-color:rgba(255,255,255,0.06);color:#5a606a'}>
					{t.label}
				</div>
			{/each}
		</div>
	</div>
	<div class="trophy-col">
		<div class="trophy-aura" aria-hidden="true"></div>
		<div class="trophy"><span class="trophy-glyph">◆</span></div>
		<div class="trophy-base" aria-hidden="true"></div>
	</div>
</div>

<div class="bottom-row">
	<HoloPanel style="flex:1.6">
		<HoloEyebrow>CLASSEMENT SEMAINE</HoloEyebrow>
		<div class="lb-header">
			<span style="width:36px">#</span>
			<span style="flex:1">JOUEUR</span>
			<span style="width:40px">NIV</span>
			<span style="width:130px">XP 7J</span>
			<span style="width:90px;text-align:right">XP SEM.</span>
			<span style="width:70px;text-align:right">DELTA</span>
		</div>
		{#each players as p}
			{#if p.rank === 8}
				<div class="lb-sep" style="color:#8a909a;border-color:rgba(255,255,255,0.1)">─ ZONE SÛRE ─</div>
			{/if}
			{#if p.rank === 10}
				<div class="lb-sep" style="color:#e89a9a;border-color:rgba(232,154,154,0.2)">─ RELÉGATION ─</div>
			{/if}
			<div class="lb-row" class:lb-me={p.me} style={p.me ? 'border:1px solid rgba(168,200,228,0.25);background:rgba(168,200,228,0.06)' : ''}>
				<span class="lb-rank" style="color:{p.rank<=3?'#a8c8e4':'#5a606a'};font-family:'JetBrains Mono',monospace;font-size:14px;font-weight:700">
					{p.rank <= 3 ? ['◆','◇','◈'][p.rank-1] : p.rank}
				</span>
				<span class="lb-name" style="color:{p.me ? '#a8c8e4' : '#e8eaed'};font-weight:{p.me ? 700 : 400}">{p.name}{p.me ? ' (vous)' : ''}</span>
				<span style="width:40px;font-family:'JetBrains Mono',monospace;font-size:11px;color:#5a606a;flex-shrink:0">{p.level}</span>
				<span style="width:130px;flex-shrink:0">
					<svg width="120" height="26">
						<path d={sparkPath(p.spark, 120, 26)} fill="none" stroke={p.me ? '#a8c8e4' : '#8a909a'} stroke-width="1.5" opacity="0.85"/>
					</svg>
				</span>
				<span style="width:90px;font-family:'JetBrains Mono',monospace;font-size:14px;color:#cbd5e0;text-align:right;font-weight:600;flex-shrink:0">{p.weekXp.toLocaleString('fr-FR')}</span>
				<span style="width:70px;font-family:'JetBrains Mono',monospace;font-size:11px;text-align:right;color:{p.delta>=0?'#7cd9a0':'#e89a9a'};flex-shrink:0">{p.delta>=0?'+':''}{p.delta}%</span>
			</div>
		{/each}
	</HoloPanel>

	<div style="flex:1;display:flex;flex-direction:column;gap:14px">
		<HoloPanel>
			<HoloEyebrow color="#a8c8e4">COURSE XP</HoloEyebrow>
			<svg width={W} height={H} style="display:block;width:100%;height:auto">
				{#each [0,1,2,3,4] as i}
					<line x1="4" y1={8+(H-16)*i/4} x2={W-4} y2={8+(H-16)*i/4} stroke="rgba(255,255,255,0.04)" stroke-width="0.5"/>
				{/each}
				<line x1="4" y1={threshold} x2={W-4} y2={threshold} stroke="#8a909a" stroke-width="1" stroke-dasharray="4,4"/>
				<text x={W-6} y={threshold-4} fill="#8a909a" font-size="8" text-anchor="end" font-family="JetBrains Mono,monospace">SEUIL TOP 7</text>
				<path d={rivalPath} fill="none" stroke="#6a7280" stroke-width="1.4"/>
				<path d={`${mePath} L${W-4},${H} L4,${H} Z`} fill="url(#meGrad)" opacity="0.18"/>
				<path d={mePath} fill="none" stroke="#a8c8e4" stroke-width="2"/>
				<defs>
					<linearGradient id="meGrad" x1="0" y1="0" x2="0" y2="1">
						<stop offset="0%" stop-color="#a8c8e4"/>
						<stop offset="100%" stop-color="transparent"/>
					</linearGradient>
				</defs>
			</svg>
			<div style="display:flex;gap:14px;margin-top:8px;font-size:10px;font-family:'JetBrains Mono',monospace">
				<span style="color:#a8c8e4">— vous</span>
				<span style="color:#6a7280">— rival</span>
			</div>
		</HoloPanel>
		<HoloPanel>
			<HoloEyebrow color="#8a909a">RIVAL DE LA SEMAINE</HoloEyebrow>
			<div style="display:flex;align-items:center;gap:12px">
				<div style="width:56px;height:56px;border-radius:8px;background:linear-gradient(135deg,#6a7280,#a8c8e4);display:flex;align-items:center;justify-content:center;font-family:'Major Mono Display',monospace;font-size:20px;color:#0e1014;box-shadow:0 0 20px rgba(168,200,228,0.12)">VT</div>
				<div>
					<p style="color:#e8eaed;font-size:14px;font-weight:600">volt_tacticien</p>
					<p style="color:#8a909a;font-size:11px;font-family:'JetBrains Mono',monospace">Rang #5 · Écart : +50 XP</p>
				</div>
			</div>
		</HoloPanel>
		<HoloPanel>
			<HoloEyebrow>RÉCOMPENSES</HoloEyebrow>
			{#each rewards as r, i}
				<div style="display:flex;justify-content:space-between;align-items:center;font-size:12px;padding:4px 0;border-top:{i>0?'1px solid rgba(255,255,255,0.05)':'none'}">
					<span style="font-family:'JetBrains Mono',monospace;font-size:10px;color:#a8c8e4;letter-spacing:0.18em">{r.label}</span>
					<span style="color:#8a909a">{r.reward}</span>
				</div>
			{/each}
		</HoloPanel>
	</div>
</div>

<style>
	.hero-row { display:grid;grid-template-columns:1.4fr 0.7fr;gap:24px;margin-bottom:22px;align-items:center; }
	@media (max-width:768px) { .hero-row { grid-template-columns:1fr; } }
	.tier-row { display:flex;gap:8px;flex-wrap:wrap; }
	.tier-chip { padding:6px 14px;border-radius:8px;border:1px solid;font-size:10px;font-weight:700;letter-spacing:0.18em;font-family:'JetBrains Mono',monospace; }

	.trophy-col { display:flex;flex-direction:column;align-items:center;position:relative; }
	.trophy-aura { position:absolute;top:10px;width:120px;height:120px;border-radius:50%;background:radial-gradient(circle,rgba(168,200,228,0.18) 0%,transparent 70%);filter:blur(20px); }
	.trophy { position:relative;width:140px;height:140px;border-radius:50%;background:conic-gradient(#a8c8e4,#6a7280,#cbd5e0,#a8c8e4);background-size:200% 100%;animation:holoSweep 6s linear infinite,holoFloat 4s ease-in-out infinite;display:flex;align-items:center;justify-content:center;box-shadow:0 0 40px rgba(168,200,228,0.18); }
	.trophy-glyph { font-family:'Major Mono Display',monospace;font-size:80px;color:#0e1014; }
	.trophy-base { width:80px;height:10px;border-radius:50%;background:rgba(168,200,228,0.12);filter:blur(8px);margin-top:4px; }

	.bottom-row { display:flex;gap:14px; }
	@media (max-width:900px) { .bottom-row { flex-direction:column; } }

	.lb-header { display:flex;align-items:center;gap:8px;font-size:9px;font-family:'JetBrains Mono',monospace;letter-spacing:0.2em;color:#5a606a;padding:0 8px 8px;border-bottom:1px solid rgba(255,255,255,0.06);margin-bottom:8px; }
	.lb-row { display:flex;align-items:center;gap:8px;padding:8px;border-radius:8px;margin-bottom:4px; }
	.lb-sep { text-align:center;font-size:10px;font-family:'JetBrains Mono',monospace;letter-spacing:0.2em;padding:6px 0;border-top:1px dashed;border-bottom:1px dashed;margin:6px 0;opacity:0.8; }
	.lb-rank { width:36px;flex-shrink:0; }
	.lb-name { flex:1;font-size:13px;font-family:'Space Grotesk',sans-serif; }
</style>
