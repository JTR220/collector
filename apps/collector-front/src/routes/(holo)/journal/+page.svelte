<script lang="ts">
	import { COLLECTIBLES } from '$lib/data/collectibles';
	import { eur } from '$lib/utils/format';
	import HoloPanel from '$lib/components/holo/HoloPanel.svelte';
	import HoloMeter from '$lib/components/holo/HoloMeter.svelte';
	import HoloEyebrow from '$lib/components/holo/HoloEyebrow.svelte';
	import HoloArt from '$lib/components/holo/HoloArt.svelte';

	const tabs = ['FEED', 'COLLECTION', 'WISHLIST', 'LISTES', 'NOTES', 'TRADES'];
	let activeTab = $state('FEED');
	const items = COLLECTIBLES;

	type DiaryKind = 'acquis' | 'trade' | 'wishlist' | 'noté' | 'vendu';
	const kindColor: Record<DiaryKind, string> = {
		acquis: '#7cd9a0', trade: '#8a909a', wishlist: '#a8c8e4', noté: '#cbd5e0', vendu: '#e89a9a'
	};

	const diary = [
		{ month:'MAI', day:20, kind:'acquis' as DiaryKind,  item:items[0], rating:5, note:'Exemplaire en parfait état, reflets holographiques intacts. Vendeur sérieux, emballage triple couche.', likes:42, comments:7, xp:420 },
		{ month:'MAI', day:18, kind:'vendu' as DiaryKind,   item:items[1], rating:null, note:null, likes:12, comments:3, xp:280 },
		{ month:'MAI', day:16, kind:'noté' as DiaryKind,    item:items[2], rating:4, note:'Très bon état pour un reprint 88. Quelques pliures mineures sur la couverture, mais rien de rédhibitoire.', likes:28, comments:5, xp:80 },
		{ month:'MAI', day:14, kind:'trade' as DiaryKind,   item:items[4], rating:null, note:null, likes:8, comments:2, xp:150 },
		{ month:'MAI', day:12, kind:'wishlist' as DiaryKind, item:items[3], rating:null, note:null, likes:5, comments:0, xp:30 },
	];

	const popularLists = [
		{ title:'Top TCG Holo 1ère édition',      handle:'holo_king',    count:24 },
		{ title:'Consoles scellées all-time',      handle:'pack_ripper',  count:18 },
		{ title:'Vinyles cultes année 2000',       handle:'groove_atlas', count:31 },
		{ title:'Designer toys édition limitée',  handle:'soho_pulse',   count:12 },
	];

	const friendActivity = [
		{ handle:'holo_king',    action:'a acquis',         target:'Charizard PSA 10',    rating:null },
		{ handle:'pack_ripper',  action:'a noté ★★★★★',     target:'Game Boy Pikachu',    rating:5    },
		{ handle:'groove_atlas', action:'a mis en wishlist', target:'Daft Punk Discovery', rating:null },
		{ handle:'arcade_twin',  action:'a publié une note', target:'Action Comics #1',    rating:3    },
		{ handle:'soho_pulse',   action:'a vendu',           target:'Bearbrick 1000%',     rating:null },
	];

	const ratingDist = [
		{ stars:5, count:48, pct:62 },
		{ stars:4, count:18, pct:23 },
		{ stars:3, count:8,  pct:10 },
		{ stars:2, count:3,  pct:4  },
		{ stars:1, count:1,  pct:1  },
	];
</script>

<svelte:head><title>JOURNAL · Collector.shop</title></svelte:head>

<div class="tabs-bar">
	<div class="tabs">
		{#each tabs as tab}
			<button class="tab-btn" class:tab-active={activeTab === tab} onclick={() => (activeTab = tab)}>{tab}</button>
		{/each}
	</div>
	<input class="search-input" placeholder="Rechercher…" type="search" />
</div>

<div style="margin-bottom:18px">
	<HoloEyebrow>PIÈCES FÉTICHES</HoloEyebrow>
	<div class="fav-grid">
		{#each items.slice(0, 4) as item}
			<div class="fav-card">
				<HoloArt {item} size={140} />
				<div style="display:flex;flex-direction:column;gap:2px;padding:6px 0">
					<span style="font-family:'JetBrains Mono',monospace;font-size:10px;color:#a8c8e4">{item.id}</span>
					<div style="font-size:14px;color:#cbd5e0">{'★'.repeat(item.rarityScore)}{'☆'.repeat(5-item.rarityScore)}</div>
					<p style="font-family:'Major Mono Display',monospace;font-size:14px;color:#e8eaed;margin:2px 0;line-height:1.1">{item.name}</p>
					<p style="font-size:10px;color:#5a606a;font-family:'JetBrains Mono',monospace">{item.year} · {item.grade}</p>
				</div>
			</div>
		{/each}
	</div>
</div>

<div class="body-grid">
	<div>
		<HoloEyebrow>MON JOURNAL</HoloEyebrow>
		{#each diary as entry}
			{@const color = kindColor[entry.kind]}
			<div class="diary-entry">
				<div class="diary-date">
					<span style="font-family:'JetBrains Mono',monospace;font-size:9px;color:#5a606a;letter-spacing:0.2em">{entry.month}</span>
					<span style="font-family:'JetBrains Mono',monospace;font-size:32px;font-weight:700;color:#e8eaed;line-height:0.9">{entry.day}</span>
				</div>
				<HoloArt item={entry.item} size={90} />
				<div class="diary-content">
					<span class="kind-tag" style="border-color:{color}44;color:{color};background:{color}12">{entry.kind}</span>
					<p style="font-family:'Major Mono Display',monospace;font-size:16px;color:#e8eaed;margin:4px 0 2px;line-height:1.1">{entry.item.name}</p>
					<p style="font-family:'JetBrains Mono',monospace;font-size:10px;color:#5a606a">{entry.item.series} · {eur(entry.item.price)}</p>
					{#if entry.rating}
						<div style="margin:4px 0;font-size:15px;color:#cbd5e0">{'★'.repeat(entry.rating)}{'☆'.repeat(5-entry.rating)}</div>
					{/if}
					{#if entry.note}
						<p style="font-size:12px;color:#8a909a;font-style:italic;margin:4px 0 8px;line-height:1.5">«&nbsp;{entry.note}&nbsp;»</p>
					{/if}
					<div style="display:flex;gap:12px;flex-wrap:wrap">
						<span style="color:#5a606a;font-size:11px">♡ {entry.likes}</span>
						<span style="color:#5a606a;font-size:11px">◌ {entry.comments}</span>
						<span style="color:#5a606a;font-size:11px">↗ partager</span>
						{#if entry.xp}<span style="font-family:'JetBrains Mono',monospace;font-size:10px;color:#a8c8e4">+{entry.xp} XP</span>{/if}
					</div>
				</div>
			</div>
		{/each}
	</div>

	<div class="aside">
		<HoloPanel style="margin-bottom:14px">
			<HoloEyebrow color="#8a909a">LISTES POPULAIRES</HoloEyebrow>
			<div style="display:flex;flex-direction:column;gap:12px">
				{#each popularLists as list}
					<div style="display:flex;align-items:center;gap:10px">
						<div class="mini-covers">
							{#each [0,1,2] as i}
								<div class="mini-cover" style="background:radial-gradient(120% 90% at 30% 20%,oklch(0.55 0.08 {215+i*15}) 0%,oklch(0.25 0.06 {215+i*15}) 100%);margin-left:{i>0?'-16px':'0'};z-index:{3-i}"></div>
							{/each}
						</div>
						<div>
							<p style="font-size:13px;color:#e8eaed;margin:0">{list.title}</p>
							<p style="font-size:10px;color:#5a606a;font-family:'JetBrains Mono',monospace;margin:2px 0 0">par @{list.handle} · {list.count} pièces</p>
						</div>
					</div>
				{/each}
			</div>
		</HoloPanel>

		<HoloPanel style="margin-bottom:14px">
			<HoloEyebrow color="#a8c8e4">VU CHEZ TES AMIS</HoloEyebrow>
			<div style="display:flex;flex-direction:column;gap:10px">
				{#each friendActivity as f}
					<div style="display:flex;align-items:center;gap:8px">
						<div style="width:32px;height:32px;border-radius:6px;background:linear-gradient(135deg,rgba(168,200,228,0.18),rgba(168,200,228,0.06));border:1px solid rgba(168,200,228,0.2);display:flex;align-items:center;justify-content:center;font-size:11px;color:#a8c8e4;font-family:'Major Mono Display',monospace;flex-shrink:0">{f.handle[0].toUpperCase()}</div>
						<div style="font-size:11px;flex:1">
							<span style="font-family:'JetBrains Mono',monospace;color:#a8c8e4">@{f.handle}</span>
							<span style="color:#5a606a;font-style:italic"> {f.action} </span>
							<span style="color:#e8eaed">{f.target}</span>
							{#if f.rating}<span style="margin-left:4px;color:#cbd5e0">{'★'.repeat(f.rating)}</span>{/if}
						</div>
					</div>
				{/each}
			</div>
		</HoloPanel>

		<HoloPanel>
			<HoloEyebrow color="#cbd5e0">RÉPARTITION NOTES</HoloEyebrow>
			<div style="display:flex;flex-direction:column;gap:8px">
				{#each ratingDist as r}
					<div style="display:flex;align-items:center;gap:8px">
						<span style="font-family:'JetBrains Mono',monospace;font-size:14px;color:#cbd5e0;width:52px;flex-shrink:0;font-weight:600">{'★'.repeat(r.stars)}</span>
						<div style="flex:1"><HoloMeter value={r.pct} height={6} /></div>
						<span style="font-family:'JetBrains Mono',monospace;font-size:10px;color:#5a606a;width:28px;text-align:right">{r.count}</span>
					</div>
				{/each}
			</div>
		</HoloPanel>
	</div>
</div>

<style>
	.tabs-bar { display:flex;justify-content:space-between;align-items:center;margin-bottom:22px;gap:12px;flex-wrap:wrap; }
	.tabs { display:flex;gap:4px;flex-wrap:wrap; }
	.tab-btn { padding:6px 14px;background:none;border:none;border-bottom:2px solid transparent;cursor:pointer;font-family:'Space Grotesk',sans-serif;font-size:12px;font-weight:600;letter-spacing:0.14em;color:#5a606a;transition:color 150ms,border-color 150ms; }
	.tab-active { color:#a8c8e4;border-bottom-color:#a8c8e4; }
	.search-input { padding:7px 14px;border:1px solid rgba(255,255,255,0.07);border-radius:8px;background:rgba(255,255,255,0.03);color:#e8eaed;font-size:12px;outline:none;font-family:'Space Grotesk',sans-serif;min-width:160px; }
	.search-input:focus { border-color:rgba(168,200,228,0.3); }
	.search-input::placeholder { color:#5a606a; }

	.fav-grid { display:grid;grid-template-columns:repeat(4,1fr);gap:12px; }
	@media (max-width:900px) { .fav-grid { grid-template-columns:repeat(2,1fr); } }
	.fav-card { display:flex;flex-direction:column;gap:0; }

	.body-grid { display:grid;grid-template-columns:1.45fr 1fr;gap:18px; }
	@media (max-width:900px) { .body-grid { grid-template-columns:1fr; } }

	.diary-entry { display:flex;align-items:flex-start;gap:14px;border-top:1px solid rgba(255,255,255,0.06);padding:16px 0; }
	.diary-entry:first-child { border-top:none;padding-top:0; }
	.diary-date { display:flex;flex-direction:column;align-items:center;flex-shrink:0;width:44px; }
	.diary-content { flex:1; }
	.kind-tag { display:inline-block;padding:2px 8px;border:1px solid;border-radius:4px;font-size:10px;font-weight:700;letter-spacing:0.16em;font-family:'JetBrains Mono',monospace;margin-bottom:4px; }

	.mini-covers { display:flex;position:relative; }
	.mini-cover { width:36px;height:36px;border-radius:4px;border:2px solid #0e1014;flex-shrink:0; }
	.aside {}
</style>
