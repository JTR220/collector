<script lang="ts">
	import { COLLECTIBLES } from '$lib/data/collectibles';
	import { eur, pct } from '$lib/utils/format';
	import HoloMeter from '$lib/components/holo/HoloMeter.svelte';

	const items = COLLECTIBLES;

	const meters = [
		{ label: 'TCG',      value: 42 },
		{ label: 'CONSOLES', value: 28 },
		{ label: 'COMICS',   value: 18 },
		{ label: 'VINYLES',  value: 12 }
	];

	type TiltState = { x: number; y: number; on: boolean; px: number; py: number };
	let tilts = $state<Record<string, TiltState>>({});

	function onMove(e: MouseEvent, id: string) {
		const el = e.currentTarget as HTMLElement;
		const r = el.getBoundingClientRect();
		const px = (e.clientX - r.left) / r.width;
		const py = (e.clientY - r.top) / r.height;
		tilts[id] = { x: (py - 0.5) * -10, y: (px - 0.5) * 14, on: true, px, py };
	}
	function onLeave(id: string) {
		tilts[id] = { x: 0, y: 0, on: false, px: 0.5, py: 0.5 };
	}

	function cardShadow(t: TiltState | undefined) {
		if (t?.on)
			return `0 24px 50px -22px rgba(168,200,228,0.22),0 6px 16px rgba(0,0,0,0.5),inset 0 0 0 1px rgba(168,200,228,0.28)`;
		return `0 12px 26px -16px rgba(0,0,0,0.6),inset 0 0 0 1px rgba(255,255,255,0.07)`;
	}
</script>

<svelte:head><title>Collector.shop · Holo Rares &amp; Scellés</title></svelte:head>

<!-- Hero -->
<section class="hero">
	<div class="hero-text">
		<div class="kicker">SAISON 03 · DROP DE LA SEMAINE</div>
		<h1 class="hero-title">HOLO RARES &amp; SCELLÉS</h1>
		<p class="hero-lede">
			Six pièces vérifiées · grading PSA / CGC · livraison <em>tracée</em> sous boîtier antichoc.
			Passe la souris sur les cartes pour faire bouger le foil.
		</p>
		<div class="hero-actions">
			<a href="/drops" class="cta-btn">OUVRIR LE BOOSTER</a>
			<a href="/quetes" class="ghost-btn">· voir les quêtes</a>
		</div>
	</div>
	<div class="hero-meters">
		{#each meters as m}
			<div class="meter-row">
				<span class="meter-label">{m.label}</span>
				<div class="meter-track"><HoloMeter value={m.value} /></div>
				<span class="meter-val">{m.value}</span>
			</div>
		{/each}
	</div>
</section>

<!-- Card grid -->
<section class="grid-section">
	{#each items as item (item.id)}
		{@const t = tilts[item.id]}
		<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
		<article
			class="card"
			style="
				transform:perspective(900px) rotateX({t?.x ?? 0}deg) rotateY({t?.y ?? 0}deg) translateZ(0);
				box-shadow:{cardShadow(t)};
				transition:transform 80ms linear,box-shadow 220ms ease;
			"
			onmousemove={(e) => onMove(e, item.id)}
			onmouseleave={() => onLeave(item.id)}
		>
			<!-- Steel foil sheen — monochrome ice, shifts with cursor -->
			<div
				class="card-sheen"
				style="
					opacity:{t?.on ? 0.32 : 0.12};
					transform:translate({((t?.px ?? 0.5)-0.5)*40}px,{((t?.py ?? 0.5)-0.5)*30}px);
				"
				aria-hidden="true"
			></div>
			<div class="card-scanlines" aria-hidden="true"></div>

			<!-- Header row -->
			<div class="card-top-row">
				<span class="card-id">{item.id}</span>
				<span class="card-cat">{item.category}</span>
			</div>

			<!-- Art zone — single azure gradient -->
			<div class="card-art">
				<div class="card-art-scanlines" aria-hidden="true"></div>
				<span class="card-glyph">{item.glyph}</span>
				<span class="card-rarity-badge">{item.rarity.toUpperCase()}</span>
			</div>

			<!-- Meta -->
			<div class="card-meta">
				<div class="name-price-row">
					<div>
						<p class="card-name">{item.name}</p>
						<p class="card-series">{item.series}</p>
					</div>
					<div style="text-align:right">
						<p class="price-label">PRIX</p>
						<p class="price-val">{eur(item.price)}</p>
					</div>
				</div>

				<div class="chip-row">
					<span class="chip">{item.grade}</span>
					<span class="chip">{item.year}</span>
					<span class="chip" style="color:{item.delta >= 0 ? '#7cd9a0' : '#e89a9a'}">{pct(item.delta)}</span>
					<span style="flex:1"></span>
					<button class="buy-btn">ENCHÉRIR</button>
				</div>
			</div>
		</article>
	{/each}
</section>

<!-- Ticker -->
<footer class="ticker">
	<span class="ticker-live">● LIVE</span>
	<div class="ticker-track">
		<div class="ticker-inner">
			{#each [...items, ...items] as item}
				<span class="ticker-item">
					<b style="color:#a8c8e4">{item.id}</b>
					<span style="color:#8a909a">{item.name}</span>
					<span style="color:{item.delta >= 0 ? '#7cd9a0' : '#e89a9a'}">{pct(item.delta)}</span>
					<span style="color:#e8eaed">{eur(item.price)}</span>
				</span>
			{/each}
		</div>
	</div>
</footer>

<style>
	/* Hero */
	.hero { display: grid; grid-template-columns: 1.4fr 1fr; gap: 48px; margin-bottom: 28px; align-items: center; padding: 24px 0 18px; position: relative; z-index: 2; }
	@media (max-width: 768px) { .hero { grid-template-columns: 1fr; gap: 24px; } }

	.kicker { font-family: 'JetBrains Mono', monospace; font-size: 10px; letter-spacing: 0.34em; color: #8a909a; margin-bottom: 14px; }
	.hero-title {
		font-family: 'Major Mono Display', monospace;
		font-size: clamp(32px, 5vw, 60px);
		line-height: 0.95; margin: 0 0 14px; letter-spacing: -0.02em;
		color: #e8eaed; text-shadow: 0 0 28px rgba(168,200,228,0.18);
	}
	.hero-lede { font-size: 13.5px; color: #8a909a; line-height: 1.55; margin-bottom: 22px; max-width: 500px; }
	.hero-actions { display: flex; gap: 18px; align-items: center; }
	.cta-btn {
		background: linear-gradient(135deg, #a8c8e4, #6a7280);
		color: #0e1014; padding: 13px 22px; border-radius: 8px;
		font-size: 11px; font-weight: 700; letter-spacing: 0.22em; text-decoration: none;
		box-shadow: 0 0 22px rgba(168,200,228,0.18); transition: filter 150ms;
	}
	.cta-btn:hover { filter: brightness(1.08); }
	.ghost-btn { color: #8a909a; font-size: 10.5px; letter-spacing: 0.18em; text-decoration: none; }

	.hero-meters { display: flex; flex-direction: column; gap: 9px; align-self: flex-end; }
	.meter-row { display: flex; align-items: center; gap: 14px; }
	.meter-label { width: 96px; font-size: 10px; letter-spacing: 0.20em; color: #8a909a; font-family: 'JetBrains Mono', monospace; }
	.meter-track { flex: 1; }
	.meter-val { width: 32px; text-align: right; font-family: 'JetBrains Mono', monospace; font-size: 11px; color: #8a909a; }

	/* Grid */
	.grid-section { display: grid; grid-template-columns: repeat(3, 1fr); gap: 16px; position: relative; z-index: 2; padding: 6px 0 18px; }
	@media (max-width: 900px) { .grid-section { grid-template-columns: repeat(2, 1fr); } }
	@media (max-width: 580px) { .grid-section { grid-template-columns: 1fr; } }

	/* Card */
	.card {
		position: relative; overflow: hidden; border-radius: 12px;
		background: linear-gradient(180deg, #181a20 0%, #0d0f13 100%);
		border: 1px solid rgba(255,255,255,0.07);
		padding: 13px; display: flex; flex-direction: column; gap: 11px;
		cursor: pointer;
	}
	.card-sheen {
		position: absolute; inset: 0; pointer-events: none;
		background: conic-gradient(from 220deg at 30% 30%,
			rgba(168,200,228,0.20),
			rgba(255,255,255,0.10),
			rgba(120,140,160,0.18),
			rgba(168,200,228,0.20));
		mix-blend-mode: color-dodge;
		filter: blur(8px) saturate(120%);
		transition: opacity 220ms ease, transform 80ms linear;
		z-index: 1;
	}
	.card-scanlines {
		position: absolute; inset: 0; pointer-events: none;
		background: repeating-linear-gradient(to bottom, rgba(255,255,255,0.04) 0 1px, transparent 1px 2px);
		mix-blend-mode: overlay; opacity: 0.4; z-index: 2;
	}

	.card-top-row { position: relative; z-index: 3; display: flex; justify-content: space-between; align-items: center; }
	.card-id { font-family: 'JetBrains Mono', monospace; font-size: 10px; letter-spacing: 0.16em; color: #a8c8e4; }
	.card-cat { font-size: 9px; letter-spacing: 0.30em; color: #8a909a; }

	.card-art {
		position: relative; z-index: 3; height: 104px; border-radius: 8px; overflow: hidden;
		display: flex; align-items: center; justify-content: center;
		background: radial-gradient(120% 90% at 30% 20%,
			oklch(0.55 0.08 215) 0%,
			oklch(0.32 0.06 215) 55%,
			oklch(0.18 0.04 215) 100%);
		box-shadow: inset 0 0 50px rgba(0,0,0,0.40);
	}
	.card-art-scanlines {
		position: absolute; inset: 0;
		background: repeating-linear-gradient(to bottom, rgba(255,255,255,0.07) 0 1px, transparent 1px 3px);
		mix-blend-mode: overlay; opacity: 0.5; pointer-events: none;
	}
	.card-glyph {
		position: relative; z-index: 1;
		font-family: 'Major Mono Display', monospace; font-size: 58px;
		color: rgba(255,255,255,0.94); text-shadow: 0 4px 22px rgba(0,0,0,0.45);
	}
	.card-rarity-badge {
		position: absolute; bottom: 7px; right: 7px; z-index: 2;
		font-size: 8px; letter-spacing: 0.22em; padding: 2px 7px;
		border: 1px solid rgba(255,255,255,0.45); border-radius: 3px;
		background: rgba(0,0,0,0.35); color: #fff;
		font-family: 'JetBrains Mono', monospace;
	}

	.card-meta { position: relative; z-index: 3; }
	.name-price-row { display: flex; gap: 10px; align-items: flex-end; margin-bottom: 10px; }
	.card-name { font-weight: 700; font-size: 15px; letter-spacing: -0.01em; color: #e8eaed; margin: 0 0 2px; }
	.card-series { font-size: 10.5px; color: #8a909a; margin: 0; }
	.price-label { font-size: 9px; letter-spacing: 0.20em; color: #5a606a; margin: 0; font-family: 'JetBrains Mono', monospace; }
	.price-val { font-family: 'JetBrains Mono', monospace; font-size: 20px; font-weight: 600; color: #a8c8e4; line-height: 1; margin: 0; }

	.chip-row { display: flex; gap: 5px; align-items: center; flex-wrap: wrap; }
	.chip {
		font-size: 9px; letter-spacing: 0.16em; padding: 3px 7px;
		border: 1px solid rgba(255,255,255,0.07); border-radius: 5px;
		background: rgba(255,255,255,0.04); color: #8a909a;
		font-family: 'JetBrains Mono', monospace;
	}
	.buy-btn {
		background: linear-gradient(135deg, #a8c8e4, #6a7280);
		color: #0e1014; border: none; padding: 6px 11px; border-radius: 5px;
		font-size: 9.5px; letter-spacing: 0.22em; font-weight: 700; cursor: pointer;
		font-family: 'Space Grotesk', sans-serif; transition: filter 150ms;
	}
	.buy-btn:hover { filter: brightness(1.08); }

	/* Ticker */
	.ticker {
		display: flex; align-items: center; gap: 18px;
		border-top: 1px solid rgba(255,255,255,0.07);
		padding: 12px 0; margin-top: 4px; overflow: hidden;
		font-family: 'JetBrains Mono', monospace; font-size: 11px;
		position: relative; z-index: 2;
	}
	.ticker-live {
		flex-shrink: 0; font-size: 11px; font-weight: 700; letter-spacing: 0.28em;
		color: #a8c8e4; text-shadow: 0 0 12px rgba(168,200,228,0.5);
		animation: holoPulse 1.8s ease-in-out infinite;
	}
	.ticker-track { flex: 1; overflow: hidden; }
	.ticker-inner { display: flex; animation: holoTicker 30s linear infinite; width: max-content; }
	.ticker-item { display: inline-flex; gap: 9px; margin-right: 34px; align-items: center; white-space: nowrap; }
</style>
