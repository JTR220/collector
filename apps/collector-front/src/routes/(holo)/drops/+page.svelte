<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { COLLECTIBLES } from '$lib/data/collectibles';
	import { eur } from '$lib/utils/format';
	import HoloPanel from '$lib/components/holo/HoloPanel.svelte';
	import HoloEyebrow from '$lib/components/holo/HoloEyebrow.svelte';
	import HoloTitle from '$lib/components/holo/HoloTitle.svelte';

	const dropAt = (() => {
		const d = new Date();
		d.setDate(d.getDate() + ((5 - d.getDay() + 7) % 7 || 7));
		d.setHours(18, 0, 0, 0);
		return d;
	})();
	let remaining = $state(0);
	let timerId: ReturnType<typeof setInterval>;
	function tick() { remaining = Math.max(0, dropAt.getTime() - Date.now()); }
	onMount(() => { tick(); timerId = setInterval(tick, 1000); });
	onDestroy(() => clearInterval(timerId));
	const days  = $derived(Math.floor(remaining / 86400_000));
	const hours = $derived(Math.floor((remaining % 86400_000) / 3600_000));
	const mins  = $derived(Math.floor((remaining % 3600_000) / 60_000));
	const secs  = $derived(Math.floor((remaining % 60_000) / 1000));
	const pad = (n: number) => String(n).padStart(2, '0');

	type DropStatus = 'live' | 'next' | 'sold' | 'soon';
	type CalDrop = { id: string; name: string; series: string; date: string; status: DropStatus; seats: string; seatsLeft: number; seatsTotal: number; price: number; resell: number; delta: number; glyph: string; };

	const calendar: CalDrop[] = [
		{ id:'DRP-041', name:'Charizard Holo',      series:'Base Set 1ère éd.',  date:'21 mai', status:'live', seats:'OPEN',   seatsLeft:3,  seatsTotal:12,  price:18400, resell:22000, delta:19.6, glyph:'卡' },
		{ id:'DRP-042', name:'Game Boy Pikachu',     series:'Sealed NTSC',        date:'24 mai', status:'next', seats:'RAFFLE', seatsLeft:48, seatsTotal:200, price:1290,  resell:2100,  delta:62.8, glyph:'電' },
		{ id:'DRP-039', name:'Action Comics',        series:'DC 1988 CGC 9.6',    date:'17 mai', status:'sold', seats:'COMPLET',seatsLeft:0,  seatsTotal:8,   price:640,   resell:800,   delta:25,   glyph:'S'  },
		{ id:'DRP-043', name:'Daft Punk Discovery',  series:'Vinyle 1ère presse', date:'28 mai', status:'soon', seats:'BIENTÔT',seatsLeft:0,  seatsTotal:30,  price:320,   resell:480,   delta:50,   glyph:'♪'  },
		{ id:'DRP-044', name:'Bearbrick 1000%',      series:'Andy Warhol 2022',   date:'31 mai', status:'soon', seats:'BIENTÔT',seatsLeft:0,  seatsTotal:5,   price:1180,  resell:1800,  delta:52.5, glyph:'★'  },
		{ id:'DRP-045', name:'Casio F-91W NATO',     series:'Custom bleu 1991',   date:'07 juin',status:'soon', seats:'BIENTÔT',seatsLeft:0,  seatsTotal:50,  price:89,    resell:150,   delta:68.5, glyph:'◷'  },
	];

	const tickets = [
		{ label:'#TICKET-0847', drop:'Game Boy Pikachu',    ratio:'1/3' },
		{ label:'#TICKET-0812', drop:'Bearbrick 1000%',     ratio:'1/1' },
		{ label:'#TICKET-0791', drop:'Daft Punk Discovery', ratio:'2/3' },
	];

	const statusLabel: Record<DropStatus, string> = { live:'LIVE', next:'PROCHAIN', sold:'ÉPUISÉ', soon:'BIENTÔT' };
	const statusColor: Record<DropStatus, string> = { live:'#a8c8e4', next:'#cbd5e0', sold:'#3a3a40', soon:'#8a909a' };
</script>

<svelte:head><title>DROPS · Collector.shop</title></svelte:head>

<div class="hero-grid">
	<div>
		<HoloEyebrow color="#a8c8e4">DRP-042 · RAFFLE OUVERT</HoloEyebrow>
		<HoloTitle size={36}>GAME BOY PIKACHU</HoloTitle>
		<p style="color:#8a909a;font-size:13px;margin:8px 0 18px">Édition scellée NTSC · Mint · Livraison sous 48h</p>

		<div class="countdown">
			{#each [{ val: days, label: 'JOURS' }, { val: hours, label: 'HEURES' }, { val: mins, label: 'MIN' }, { val: secs, label: 'SEC' }] as seg}
				<div class="cd-box">
					<span class="cd-val">{pad(seg.val)}</span>
					<span class="cd-label">{seg.label}</span>
				</div>
			{/each}
		</div>

		<div style="display:flex;gap:10px;align-items:center;margin-bottom:16px;flex-wrap:wrap">
			<button class="raffle-btn">◆ ENTRER DANS LE RAFFLE</button>
			<button class="remind-btn">+ RAPPEL</button>
		</div>
		<p style="font-family:'JetBrains Mono',monospace;font-size:11px;color:#5a606a;letter-spacing:0.1em">
			3 482 entrants · 0.34% · 1 290€ drop · 2 100€ resell (+62.8%)
		</p>
	</div>

	<div class="hero-card">
		<div class="hero-card-sheen" aria-hidden="true"></div>
		<div class="hero-card-scanlines" aria-hidden="true"></div>
		<span class="hero-card-glyph">電</span>
		<div style="position:absolute;bottom:12px;right:12px;display:flex;flex-direction:column;gap:4px;z-index:2">
			<span class="seal">MINT</span>
			<span class="seal">HOLO RARE</span>
		</div>
	</div>
</div>

<HoloEyebrow color="#8a909a">CALENDRIER DES DROPS</HoloEyebrow>
<div class="calendar-grid">
	{#each calendar as drop}
		<HoloPanel style={drop.status === 'sold' ? 'opacity:0.55' : ''}>
			<div style="display:flex;justify-content:space-between;align-items:flex-start;margin-bottom:10px">
				<div>
					<span style="font-family:'JetBrains Mono',monospace;font-size:9px;color:#a8c8e4">{drop.id}</span>
					<p style="font-family:'JetBrains Mono',monospace;font-size:10px;color:#5a606a;margin:2px 0 0">{drop.date}</p>
				</div>
				<span class="status-tag" style="border-color:{statusColor[drop.status]}55;color:{statusColor[drop.status]};background:{statusColor[drop.status]}14">{statusLabel[drop.status]}</span>
			</div>
			<div class="mini-art" style={drop.status === 'sold' ? 'filter:grayscale(1) brightness(0.5)' : ''}>
				<span class="mini-glyph">{drop.glyph}</span>
			</div>
			<p style="font-family:'Major Mono Display',monospace;font-size:14px;color:#e8eaed;margin:8px 0 2px;line-height:1.1">{drop.name}</p>
			<p style="font-size:10px;color:#5a606a;font-family:'JetBrains Mono',monospace;margin:0 0 8px">{drop.series}</p>
			{#if drop.status !== 'sold'}
				<p style="font-size:11px;font-family:'JetBrains Mono',monospace;color:#a8c8e4;margin-bottom:8px">{drop.seatsLeft > 0 ? `${drop.seatsLeft}/${drop.seatsTotal} restants · ` : ''}{drop.seats}</p>
			{/if}
			<div style="display:flex;justify-content:space-between;align-items:flex-end;margin-bottom:10px">
				<div>
					<p style="font-size:9px;color:#5a606a;font-family:'JetBrains Mono',monospace;margin:0">DROP</p>
					<p style="font-family:'JetBrains Mono',monospace;font-size:18px;font-weight:600;color:#a8c8e4;margin:0">{eur(drop.price)}</p>
				</div>
				<div style="text-align:right">
					<p style="font-size:9px;color:#5a606a;font-family:'JetBrains Mono',monospace;margin:0">RESELL</p>
					<p style="font-family:'JetBrains Mono',monospace;font-size:18px;font-weight:600;color:#7cd9a0;margin:0">{eur(drop.resell)}</p>
				</div>
			</div>
			{#if drop.status === 'live'}
				<button class="cal-cta" style="background:linear-gradient(135deg,#a8c8e4,#6a7280);color:#0e1014">ACHETER</button>
			{:else if drop.status === 'next'}
				<button class="cal-cta" style="background:linear-gradient(135deg,#cbd5e0,#8a909a);color:#0e1014">ENTRER RAFFLE</button>
			{:else if drop.status === 'sold'}
				<button class="cal-cta" style="background:rgba(255,255,255,0.04);border:1px solid rgba(255,255,255,0.07);cursor:not-allowed;color:#5a606a" disabled>REJOINDRE WL</button>
			{:else}
				<button class="cal-cta" style="border:1px solid rgba(168,200,228,0.25);background:transparent;color:#a8c8e4">+ RAPPEL</button>
			{/if}
		</HoloPanel>
	{/each}
</div>

<HoloPanel glow="#a8c8e4" style="margin-top:18px">
	<HoloEyebrow color="#a8c8e4">TES TICKETS ACTIFS</HoloEyebrow>
	<div class="tickets-grid">
		{#each tickets as t}
			<div class="ticket-card">
				<span style="font-family:'JetBrains Mono',monospace;font-size:10px;color:#a8c8e4">{t.label}</span>
				<p style="font-size:13px;color:#e8eaed;margin:6px 0 4px">{t.drop}</p>
				<span style="font-family:'JetBrains Mono',monospace;font-size:20px;color:#a8c8e4;font-weight:700">{t.ratio}</span>
			</div>
		{/each}
	</div>
</HoloPanel>

<style>
	.hero-grid { display:grid;grid-template-columns:1.3fr 0.7fr;gap:24px;margin-bottom:24px;align-items:center; }
	@media (max-width:768px) { .hero-grid { grid-template-columns:1fr; } }

	.countdown { display:flex;gap:12px;margin-bottom:18px; }
	.cd-box { display:flex;flex-direction:column;align-items:center;padding:10px 14px;border:1px solid rgba(255,255,255,0.07);border-radius:8px;background:rgba(168,200,228,0.04); }
	.cd-val { font-family:'JetBrains Mono',monospace;font-size:42px;font-weight:700;line-height:1;color:#a8c8e4; }
	.cd-label { font-size:9px;font-family:'JetBrains Mono',monospace;letter-spacing:0.2em;color:#5a606a; }

	.raffle-btn { flex:1;padding:13px 22px;border-radius:8px;border:none;background:linear-gradient(135deg,#a8c8e4,#6a7280);color:#0e1014;font-size:12px;font-weight:700;letter-spacing:0.22em;cursor:pointer;box-shadow:0 0 22px rgba(168,200,228,0.15); }
	.remind-btn { padding:13px 16px;border-radius:8px;border:1px solid rgba(168,200,228,0.25);background:transparent;color:#a8c8e4;font-size:12px;font-weight:600;cursor:pointer; }

	.hero-card { aspect-ratio:3/4;border-radius:12px;position:relative;overflow:hidden;background:radial-gradient(120% 90% at 30% 20%,oklch(0.55 0.08 215) 0%,oklch(0.32 0.06 215) 55%,oklch(0.18 0.04 215) 100%);display:flex;align-items:center;justify-content:center;box-shadow:0 0 0 1px rgba(168,200,228,0.2),0 20px 50px -20px rgba(168,200,228,0.15); }
	.hero-card-sheen { position:absolute;inset:0;pointer-events:none;background:conic-gradient(from 220deg at 30% 30%,rgba(168,200,228,0.20),rgba(255,255,255,0.10),rgba(120,140,160,0.18),rgba(168,200,228,0.20));mix-blend-mode:color-dodge;filter:blur(12px);opacity:0.5; }
	.hero-card-scanlines { position:absolute;inset:0;background:repeating-linear-gradient(to bottom,rgba(255,255,255,0.07) 0 1px,transparent 1px 3px);mix-blend-mode:overlay;opacity:0.5; }
	.hero-card-glyph { position:relative;z-index:1;font-family:'Major Mono Display',monospace;font-size:140px;color:rgba(255,255,255,0.88);text-shadow:0 4px 40px rgba(0,0,0,0.6); }
	.seal { padding:3px 8px;border:1px solid rgba(255,255,255,0.35);border-radius:4px;font-size:9px;font-weight:700;letter-spacing:0.16em;background:rgba(0,0,0,0.5);color:#fff;font-family:'JetBrains Mono',monospace; }

	.calendar-grid { display:grid;grid-template-columns:repeat(3,1fr);gap:12px;margin-bottom:0; }
	@media (max-width:900px) { .calendar-grid { grid-template-columns:repeat(2,1fr); } }
	@media (max-width:580px) { .calendar-grid { grid-template-columns:1fr; } }

	.status-tag { padding:3px 8px;border:1px solid;border-radius:4px;font-size:9px;font-weight:700;letter-spacing:0.16em;font-family:'JetBrains Mono',monospace; }
	.mini-art { height:90px;border-radius:8px;display:flex;align-items:center;justify-content:center;overflow:hidden;background:radial-gradient(120% 90% at 30% 20%,oklch(0.55 0.08 215) 0%,oklch(0.32 0.06 215) 55%,oklch(0.18 0.04 215) 100%); }
	.mini-glyph { font-family:'Major Mono Display',monospace;font-size:40px;color:rgba(255,255,255,0.88); }
	.cal-cta { width:100%;padding:9px;border-radius:8px;border:none;font-size:10px;font-weight:700;letter-spacing:0.18em;cursor:pointer;font-family:'Space Grotesk',sans-serif; }

	.tickets-grid { display:grid;grid-template-columns:repeat(3,1fr);gap:12px; }
	@media (max-width:700px) { .tickets-grid { grid-template-columns:1fr; } }
	.ticket-card { border:2px dashed rgba(168,200,228,0.3);border-radius:8px;padding:14px;display:flex;flex-direction:column;gap:2px; }
</style>
