<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { fetchArticles, type ArticleAPI } from '$lib/api/catalog';
	import { eur } from '$lib/utils/format';
	import HoloPanel from '$lib/components/holo/HoloPanel.svelte';
	import HoloEyebrow from '$lib/components/holo/HoloEyebrow.svelte';
	import HoloTitle from '$lib/components/holo/HoloTitle.svelte';

	let articles = $state<ArticleAPI[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);

	onMount(async () => {
		try {
			articles = await fetchArticles();
		} catch (e) {
			error = 'Impossible de charger les drops.';
			console.error(e);
		} finally {
			loading = false;
		}
	});

	// Prochain drop = premier article avec status 'next', sinon 'live'
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
	function tick() { remaining = Math.max(0, dropAt.getTime() - Date.now()); }
	onMount(() => { tick(); timerId = setInterval(tick, 1000); });
	onDestroy(() => clearInterval(timerId));
	const days  = $derived(Math.floor(remaining / 86400_000));
	const hours = $derived(Math.floor((remaining % 86400_000) / 3600_000));
	const mins  = $derived(Math.floor((remaining % 3600_000) / 60_000));
	const secs  = $derived(Math.floor((remaining % 60_000) / 1000));
	const pad = (n: number) => String(n).padStart(2, '0');

	type DropStatus = 'live' | 'next' | 'sold' | 'soon';
	const statusLabel: Record<DropStatus, string> = { live: 'LIVE', next: 'PROCHAIN', sold: 'ÉPUISÉ', soon: 'BIENTÔT' };
	const statusColor: Record<DropStatus, string> = { live: '#a8c8e4', next: '#cbd5e0', sold: '#3a3a40', soon: '#8a909a' };
</script>

<svelte:head><title>DROPS · Collector.shop</title></svelte:head>

{#if loading}
	<div class="state-msg">Chargement des drops…</div>
{:else if error}
	<div class="state-msg error">{error}</div>
{:else if featured}
	<div class="hero-grid">
		<div>
			<HoloEyebrow color="#a8c8e4">{featured.dropId} · {statusLabel[featured.dropStatus]} OUVERT</HoloEyebrow>
			<HoloTitle size={36}>{featured.name.toUpperCase()}</HoloTitle>
			<p style="color:#8a909a;font-size:13px;margin:8px 0 18px">{featured.series} · {featured.grade} · Livraison sous 48h</p>

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
				{featured.seatsTotal} places · {eur(featured.prix)} drop · {eur(featured.resellPrice)} resell
			</p>
		</div>

		<div class="hero-card">
			<div class="hero-card-sheen" aria-hidden="true"></div>
			<div class="hero-card-scanlines" aria-hidden="true"></div>
			<span class="hero-card-glyph">{featured.glyph}</span>
			<div style="position:absolute;bottom:12px;right:12px;display:flex;flex-direction:column;gap:4px;z-index:2">
				<span class="seal">{featured.grade}</span>
				<span class="seal">{featured.rarity.toUpperCase()}</span>
			</div>
		</div>
	</div>
{/if}

{#if !loading && !error}
	<HoloEyebrow color="#8a909a">CALENDRIER DES DROPS</HoloEyebrow>
	<div class="calendar-grid">
		{#each articles as article (article.ID)}
			{@const status = article.dropStatus as DropStatus}
			<HoloPanel style={status === 'sold' ? 'opacity:0.55' : ''}>
				<div style="display:flex;justify-content:space-between;align-items:flex-start;margin-bottom:10px">
					<div>
						<span style="font-family:'JetBrains Mono',monospace;font-size:9px;color:#a8c8e4">{article.dropId}</span>
						<p style="font-family:'JetBrains Mono',monospace;font-size:10px;color:#5a606a;margin:2px 0 0">{article.dropDate}</p>
					</div>
					<span class="status-tag" style="border-color:{statusColor[status]}55;color:{statusColor[status]};background:{statusColor[status]}14">{statusLabel[status]}</span>
				</div>
				<div class="mini-art" style={status === 'sold' ? 'filter:grayscale(1) brightness(0.5)' : ''}>
					<span class="mini-glyph">{article.glyph}</span>
				</div>
				<p style="font-family:'Major Mono Display',monospace;font-size:14px;color:#e8eaed;margin:8px 0 2px;line-height:1.1">{article.name}</p>
				<p style="font-size:10px;color:#5a606a;font-family:'JetBrains Mono',monospace;margin:0 0 8px">{article.series}</p>
				{#if status !== 'sold'}
					<p style="font-size:11px;font-family:'JetBrains Mono',monospace;color:#a8c8e4;margin-bottom:8px">
						{article.seatsLeft > 0 ? `${article.seatsLeft}/${article.seatsTotal} restants · ` : ''}{status === 'live' ? 'OPEN' : status === 'next' ? 'RAFFLE' : 'BIENTÔT'}
					</p>
				{/if}
				<div style="display:flex;justify-content:space-between;align-items:flex-end;margin-bottom:10px">
					<div>
						<p style="font-size:9px;color:#5a606a;font-family:'JetBrains Mono',monospace;margin:0">DROP</p>
						<p style="font-family:'JetBrains Mono',monospace;font-size:18px;font-weight:600;color:#a8c8e4;margin:0">{eur(article.prix)}</p>
					</div>
					<div style="text-align:right">
						<p style="font-size:9px;color:#5a606a;font-family:'JetBrains Mono',monospace;margin:0">RESELL</p>
						<p style="font-family:'JetBrains Mono',monospace;font-size:18px;font-weight:600;color:#7cd9a0;margin:0">{eur(article.resellPrice)}</p>
					</div>
				</div>
				{#if status === 'live'}
					<button class="cal-cta" style="background:linear-gradient(135deg,#a8c8e4,#6a7280);color:#0e1014">ACHETER</button>
				{:else if status === 'next'}
					<button class="cal-cta" style="background:linear-gradient(135deg,#cbd5e0,#8a909a);color:#0e1014">ENTRER RAFFLE</button>
				{:else if status === 'sold'}
					<button class="cal-cta" style="background:rgba(255,255,255,0.04);border:1px solid rgba(255,255,255,0.07);cursor:not-allowed;color:#5a606a" disabled>REJOINDRE WL</button>
				{:else}
					<button class="cal-cta" style="border:1px solid rgba(168,200,228,0.25);background:transparent;color:#a8c8e4">+ RAPPEL</button>
				{/if}
			</HoloPanel>
		{/each}
	</div>
{/if}

<style>
	.state-msg { text-align: center; padding: 60px 0; font-family: 'JetBrains Mono', monospace; font-size: 12px; color: #8a909a; letter-spacing: 0.18em; }
	.state-msg.error { color: #e89a9a; }

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
</style>
