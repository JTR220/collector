<script lang="ts">
	import HoloPanel from '$lib/components/holo/HoloPanel.svelte';
	import HoloMeter from '$lib/components/holo/HoloMeter.svelte';
	import HoloSpark from '$lib/components/holo/HoloSpark.svelte';
	import HoloEyebrow from '$lib/components/holo/HoloEyebrow.svelte';
	import HoloTitle from '$lib/components/holo/HoloTitle.svelte';
	import HoloChip from '$lib/components/holo/HoloChip.svelte';

	const user = { handle: 'kanto_archive', displayName: 'Kanto Archive', bio: 'Collectionneur TCG · PSA / CGC · Spécialiste 1ère édition', level: 12, xp: 3820, xpToNext: 5000, streak: 47, league: 'OR', rank: 4, passLevel: 14 };

	const passNodes = Array.from({ length: 11 }, (_, i) => ({ level: 10 + i, state: i < 4 ? 'done' : i === 4 ? 'current' : 'locked' }));

	const stats = [
		{ label: 'PIÈCES',  value: '127', spark: [80,95,100,108,110,115,120,127] },
		{ label: 'VENTES',  value: '43',  spark: [20,24,28,30,32,36,40,43] },
		{ label: 'VOL. EUR',value: '84K', spark: [40,48,55,58,62,68,75,84] },
		{ label: 'ENCHÈRES',value: '18',  spark: [8,10,11,12,13,14,16,18] },
		{ label: 'SCORE',   value: '4.98',spark: [4.7,4.8,4.85,4.88,4.9,4.92,4.96,4.98] },
		{ label: 'STREAK',  value: '47j', spark: [10,18,25,30,35,38,44,47] },
	];

	const badges = [
		{ icon: '◆', label: 'HOLOGRAPHE', desc: '10 pièces holo', active: true },
		{ icon: '◈', label: 'PSA MASTER',  desc: 'Note moy. PSA 9+', active: true },
		{ icon: '★', label: 'VENDEUR TOP', desc: '4.9+ sur 30 ventes', active: true },
		{ icon: '●', label: 'STREAK 30',   desc: '30 jours consécutifs', active: true },
		{ icon: '◇', label: 'SCELLÉ HUNT', desc: '5 scellés acquis', active: false },
		{ icon: '◷', label: 'HORLOGERIE',  desc: '3 montres notées', active: false },
		{ icon: '♪', label: 'VINYLOPHILE', desc: '5 vinyles', active: false },
		{ icon: 'S', label: 'COMICS FAN',  desc: '5 comics notés', active: false },
	];

	const activity = [
		{ date: '20 mai', kind: 'ACQUIS', label: 'Charizard PSA 9',        xp: 420, pos: true },
		{ date: '18 mai', kind: 'VENDU',  label: 'Game Boy Pikachu',        xp: 280, pos: false },
		{ date: '16 mai', kind: 'NOTÉ',   label: 'Action Comics #1 — ★★★★', xp: 80,  pos: true },
		{ date: '14 mai', kind: 'QUÊTE',  label: 'Mission hebdo complétée', xp: 500, pos: true },
		{ date: '12 mai', kind: 'TRADE',  label: 'Bearbrick ↔ Vinyl',       xp: 150, pos: true },
		{ date: '10 mai', kind: 'LIGUE',  label: 'Promotion OR',            xp: 600, pos: true },
		{ date: '08 mai', kind: 'ACQUIS', label: 'Casio F-91W custom',      xp: 60,  pos: true },
	];
</script>

<svelte:head><title>PROFIL · Collector.shop</title></svelte:head>

<div class="identity">
	<div class="avatar-wrap">
		<div class="avatar-ring" aria-hidden="true"></div>
		<div class="avatar">KA</div>
	</div>
	<div class="identity-text">
		<HoloEyebrow color="#8a909a">@{user.handle}</HoloEyebrow>
		<HoloTitle size={36}>{user.displayName}</HoloTitle>
		<p class="bio">{user.bio}</p>
		<div style="display:flex;gap:8px;margin-top:10px;flex-wrap:wrap">
			<HoloChip color="#a8c8e4">LIGUE {user.league}</HoloChip>
			<HoloChip color="#8a909a">RANG #{user.rank}</HoloChip>
			<HoloChip color="#cbd5e0">STREAK {user.streak}j</HoloChip>
		</div>
	</div>
	<button class="follow-btn">+ SUIVRE</button>
</div>

<HoloPanel glow="#a8c8e4" style="margin-bottom:18px">
	<HoloEyebrow color="#a8c8e4">SAISON PASS · SAISON 03</HoloEyebrow>
	<div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:14px">
		<p style="font-size:13px;color:#8a909a">Niveau {user.passLevel} · {user.xp.toLocaleString('fr-FR')} / {user.xpToNext.toLocaleString('fr-FR')} XP</p>
		<span style="font-family:'JetBrains Mono',monospace;font-size:14px;color:#cbd5e0;font-weight:600">{Math.round(user.xp/user.xpToNext*100)}%</span>
	</div>
	<HoloMeter value={Math.round(user.xp / user.xpToNext * 100)} height={10} />
	<div class="pass-track">
		{#each passNodes as node}
			<div class="pass-node-wrap">
				<div class="pass-node"
					class:pass-done={node.state === 'done'}
					class:pass-current={node.state === 'current'}
					class:pass-locked={node.state === 'locked'}
				>
					{node.state === 'done' ? '✓' : node.state === 'current' ? '◆' : node.level}
				</div>
				<span class="pass-node-label">{node.level}</span>
			</div>
		{/each}
	</div>
</HoloPanel>

<div class="three-col">
	<HoloPanel>
		<HoloEyebrow>STATISTIQUES</HoloEyebrow>
		<div class="stats-grid">
			{#each stats as s}
				<div class="stat-cell">
					<span style="font-size:10px;letter-spacing:0.24em;font-family:'JetBrains Mono',monospace;color:#8a909a">{s.label}</span>
					<span style="font-family:'JetBrains Mono',monospace;font-size:28px;font-weight:700;color:#a8c8e4;line-height:1">{s.value}</span>
					<HoloSpark values={s.spark} />
				</div>
			{/each}
		</div>
	</HoloPanel>

	<HoloPanel>
		<HoloEyebrow>BADGES</HoloEyebrow>
		<div class="badge-grid">
			{#each badges as b}
				<div class="badge-card" class:badge-locked={!b.active}
					style={b.active ? 'background:rgba(168,200,228,0.06);border-color:rgba(168,200,228,0.2)' : ''}>
					<span class="badge-icon" style={b.active ? 'color:#a8c8e4' : 'color:#2a2e35'}>{b.icon}</span>
					<span class="badge-label" style={b.active ? 'color:#a8c8e4' : 'color:#2a2e35'}>{b.label}</span>
					<span class="badge-desc">{b.desc}</span>
				</div>
			{/each}
		</div>
	</HoloPanel>

	<HoloPanel>
		<HoloEyebrow>ACTIVITÉ RÉCENTE</HoloEyebrow>
		<div class="activity-list">
			{#each activity as a}
				<div class="activity-row">
					<span style="font-family:'JetBrains Mono',monospace;color:#5a606a;font-size:10px;flex-shrink:0">{a.date}</span>
					<span class="activity-kind" style="color:{a.pos ? '#7cd9a0' : '#e89a9a'};border-color:{a.pos ? '#7cd9a044' : '#e89a9a44'};background:{a.pos ? '#7cd9a014' : '#e89a9a14'}">{a.kind}</span>
					<span style="flex:1;font-size:11px;color:#e8eaed">{a.label}</span>
					<span style="font-family:'JetBrains Mono',monospace;font-size:11px;color:#cbd5e0;flex-shrink:0">+{a.xp}XP</span>
				</div>
			{/each}
		</div>
	</HoloPanel>
</div>

<style>
	.identity { display:flex;align-items:center;gap:22px;margin-bottom:22px;flex-wrap:wrap; }
	.avatar-wrap { position:relative;width:120px;height:120px;flex-shrink:0; }
	.avatar-ring {
		position:absolute;inset:-6px;border-radius:50%;
		background:conic-gradient(#a8c8e4,#6a7280,#cbd5e0,#8a909a,#a8c8e4);
		background-size:200% 100%;animation:holoSweep 6s linear infinite;
		filter:blur(4px);opacity:0.6;
	}
	.avatar {
		position:relative;width:120px;height:120px;border-radius:50%;
		background:linear-gradient(135deg,#6a7280,#a8c8e4);
		display:flex;align-items:center;justify-content:center;
		font-family:'Major Mono Display',monospace;font-size:32px;color:#0e1014;
		box-shadow:0 0 30px rgba(168,200,228,0.2);
	}
	.identity-text { flex:1;min-width:200px; }
	.bio { font-size:13px;color:#8a909a;margin:8px 0 0; }
	.follow-btn {
		padding:10px 20px;border:1px solid rgba(168,200,228,0.33);border-radius:8px;
		background:transparent;color:#a8c8e4;font-size:11px;font-weight:700;
		letter-spacing:0.2em;cursor:pointer;transition:background 150ms;
	}
	.follow-btn:hover { background:rgba(168,200,228,0.08); }

	.pass-track { display:grid;grid-template-columns:repeat(11,1fr);gap:6px;margin-top:16px; }
	.pass-node-wrap { display:flex;flex-direction:column;align-items:center;gap:4px; }
	.pass-node {
		width:32px;height:32px;border-radius:50%;display:flex;align-items:center;
		justify-content:center;font-size:11px;font-weight:700;
		border:1px solid rgba(255,255,255,0.07);background:rgba(255,255,255,0.03);color:#5a606a;
	}
	.pass-done { background:rgba(168,200,228,0.12);border-color:rgba(168,200,228,0.4);color:#a8c8e4; }
	.pass-current { background:rgba(168,200,228,0.16);border-color:#a8c8e4;color:#a8c8e4;box-shadow:0 0 10px rgba(168,200,228,0.3);animation:holoPulse 1.4s ease-in-out infinite; }
	.pass-locked { opacity:0.35; }
	.pass-node-label { font-size:9px;color:#5a606a;font-family:'JetBrains Mono',monospace; }

	.three-col { display:grid;grid-template-columns:1.1fr 1fr 0.9fr;gap:14px; }
	@media (max-width:1024px) { .three-col { grid-template-columns:1fr 1fr; } }
	@media (max-width:640px)  { .three-col { grid-template-columns:1fr; } }

	.stats-grid { display:grid;grid-template-columns:1fr 1fr;gap:14px; }
	.stat-cell { display:flex;flex-direction:column;gap:4px; }

	.badge-grid { display:grid;grid-template-columns:1fr 1fr;gap:8px; }
	.badge-card { border:1px solid rgba(255,255,255,0.07);border-radius:8px;padding:10px;display:flex;flex-direction:column;gap:3px; }
	.badge-locked { opacity:0.35;filter:grayscale(1); }
	.badge-icon { font-family:'Major Mono Display',monospace;font-size:22px; }
	.badge-label { font-size:9px;font-weight:700;letter-spacing:0.18em;font-family:'JetBrains Mono',monospace; }
	.badge-desc { font-size:10px;color:#5a606a; }

	.activity-list { display:flex;flex-direction:column;gap:8px; }
	.activity-row { display:flex;align-items:center;gap:8px;font-size:11px;border-top:1px solid rgba(255,255,255,0.06);padding-top:8px; }
	.activity-row:first-child { border-top:none;padding-top:0; }
	.activity-kind { padding:1px 6px;border:1px solid;border-radius:3px;font-size:9px;font-weight:700;letter-spacing:0.14em;flex-shrink:0;font-family:'JetBrains Mono',monospace; }
</style>
