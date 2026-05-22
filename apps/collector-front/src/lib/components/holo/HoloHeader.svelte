<script lang="ts">
	import { auth } from '$lib/stores/auth';
	import { goto } from '$app/navigation';

	type Props = {
		active?: string;
		credits?: string;
		level?: number;
		xp?: number;
		streak?: number;
	};
	let { active = 'VITRINE', credits = '04 820', level = 12, xp = 3820, streak = 47 }: Props = $props();

	const nav = [
		{ label: 'VITRINE', href: '/' },
		{ label: 'PROFIL', href: '/profil' },
		{ label: 'QUÊTES', href: '/quetes' },
		{ label: 'LIGUE', href: '/ligue' },
		{ label: 'DROPS', href: '/drops' },
		{ label: 'JOURNAL', href: '/journal' }
	];

	function logout() {
		auth.logout();
		goto('/login');
	}
</script>

<header class="header">
	<a href="/vitrine" class="logo-block">
		<span class="logo-glyph">◆</span>
		<span class="logo-word">COLLECTOR<span style="color:#a8c8e4">.</span>SHOP</span>
	</a>

	<nav class="nav">
		{#each nav as item}
			<a href={item.href} class="nav-link" class:nav-link-active={item.label === active}>
				{item.label}
			</a>
		{/each}
	</nav>

	<div class="user-strip">
		{#if $auth.user}
			<div
				class="tip-wrap"
				data-tip="Votre niveau de collectionneur. Montez en XP pour débloquer des rangs de ligue et des récompenses exclusives."
			>
				<div class="pill" style="border-color:#a8c8e455;background:#a8c8e40e">
					<span class="pill-icon" style="color:#a8c8e4">◈</span>
					<div class="pill-inner">
						<span class="pill-label">NIV</span>
						<span class="pill-val" style="color:#a8c8e4">{level}</span>
					</div>
				</div>
			</div>

			<div
				class="tip-wrap"
				data-tip="Points d'expérience gagnés en achetant, notant des pièces et en complétant des quêtes. Progressez vers le niveau suivant."
			>
				<div class="pill" style="border-color:#cbd5e055;background:#cbd5e00e">
					<span class="pill-icon" style="color:#cbd5e0">⚡</span>
					<div class="pill-inner">
						<span class="pill-label">XP</span>
						<span class="pill-val" style="color:#cbd5e0">{xp.toLocaleString('fr-FR')}</span>
					</div>
				</div>
			</div>

			<div
				class="tip-wrap"
				data-tip="Jours consécutifs d'activité. Ne cassez pas votre streak pour conserver vos bonus de quêtes quotidiennes !"
			>
				<div class="pill" style="border-color:#8a909a55;background:#8a909a0e">
					<span class="pill-icon pulse" style="color:#8a909a">●</span>
					<div class="pill-inner">
						<span class="pill-label">STREAK</span>
						<span class="pill-val" style="color:#8a909a">{streak}j</span>
					</div>
				</div>
			</div>

			<div
				class="tip-wrap"
				data-tip="Monnaie virtuelle de la plateforme. Utilisez vos crédits pour participer aux raffles et accéder aux drops exclusifs."
			>
				<div class="coin">
					<span class="coin-label">CRÉDITS</span>
					<span class="coin-val">{credits}</span>
				</div>
			</div>

			<div class="user-block">
				<div class="user-info">
					<span class="user-label">COMPTE</span>
					<span class="user-name">{$auth.user.name}</span>
				</div>
				<button class="logout-btn" onclick={logout}>↩</button>
			</div>
		{:else}
			<a href="/login" class="login-btn">SE CONNECTER</a>
		{/if}
	</div>
</header>

<style>
	.header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 16px;
		padding: 0 0 18px;
		border-bottom: 1px solid rgba(255,255,255,0.07);
		margin-bottom: 22px;
		flex-wrap: wrap;
		position: relative;
		z-index: 2;
	}
	.logo-block {
		display: flex;
		align-items: center;
		gap: 10px;
		text-decoration: none;
	}
	.logo-glyph {
		font-family: 'Major Mono Display', monospace;
		font-size: 20px;
		color: #a8c8e4;
		text-shadow: 0 0 12px rgba(168,200,228,0.5);
	}
	.logo-word {
		font-family: 'Space Grotesk', sans-serif;
		font-weight: 700;
		font-size: 13px;
		letter-spacing: 0.20em;
		color: #e8eaed;
	}
	.nav {
		display: flex;
		gap: 4px;
		flex-wrap: wrap;
	}
	.nav-link {
		padding: 7px 11px;
		border-radius: 999px;
		font-size: 10px;
		font-weight: 600;
		letter-spacing: 0.28em;
		color: #8a909a;
		border: 1px solid rgba(255,255,255,0.07);
		text-decoration: none;
		transition: border-color 120ms, color 120ms, background 120ms;
	}
	.nav-link:hover {
		border-color: rgba(168,200,228,0.35);
		color: #e8eaed;
	}
	.nav-link-active {
		background: linear-gradient(135deg, #a8c8e4, #6a7280);
		color: #0e1014;
		border-color: transparent;
		font-weight: 700;
		box-shadow: 0 0 14px rgba(168,200,228,0.18);
	}
	.user-strip {
		display: flex;
		align-items: center;
		gap: 8px;
		flex-wrap: wrap;
	}
	.pill {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 6px 10px;
		border: 1px solid;
		border-radius: 8px;
	}
	.pill-icon { font-size: 13px; }
	.pill-inner {
		display: flex;
		flex-direction: column;
		line-height: 1;
	}
	.pill-label {
		font-size: 9px;
		letter-spacing: 0.22em;
		color: #5a606a;
	}
	.pill-val {
		font-family: 'JetBrains Mono', monospace;
		font-size: 14px;
		font-weight: 600;
	}
	.pulse { animation: holoPulse 1.6s ease-in-out infinite; }
	.coin {
		display: flex;
		flex-direction: column;
		align-items: flex-end;
		padding: 6px 12px;
		border: 1px solid rgba(203,213,224,0.33);
		border-radius: 8px;
		background: rgba(203,213,224,0.06);
	}
	.coin-label {
		font-size: 9px;
		letter-spacing: 0.24em;
		color: #8a909a;
	}
	.coin-val {
		font-family: 'JetBrains Mono', monospace;
		font-size: 16px;
		font-weight: 600;
		letter-spacing: 0.06em;
		color: #cbd5e0;
	}

	/* ── Tooltips ── */
	.tip-wrap { position: relative; cursor: default; }

	.tip-wrap::after {
		content: attr(data-tip);
		position: absolute;
		top: calc(100% + 10px);
		right: 50%;
		transform: translateX(50%) translateY(-4px);
		width: 220px;
		padding: 9px 12px;
		border-radius: 8px;
		background: #181a20;
		border: 1px solid rgba(255,255,255,0.14);
		box-shadow:
			0 0 0 1px rgba(168,200,228,0.12),
			0 12px 32px -8px rgba(0,0,0,0.7);
		font-family: 'Space Grotesk', sans-serif;
		font-size: 11px;
		font-weight: 400;
		line-height: 1.5;
		letter-spacing: 0;
		color: #8a909a;
		text-align: center;
		white-space: normal;
		pointer-events: none;
		opacity: 0;
		transition: opacity 140ms ease, transform 140ms ease;
		z-index: 10000;
	}
	.tip-wrap::before {
		content: '';
		position: absolute;
		top: calc(100% + 4px);
		right: 50%;
		transform: translateX(50%);
		border: 5px solid transparent;
		border-bottom-color: rgba(255,255,255,0.14);
		pointer-events: none;
		opacity: 0;
		transition: opacity 140ms ease;
		z-index: 10001;
	}
	.tip-wrap:hover::after {
		opacity: 1;
		transform: translateX(50%) translateY(0);
	}
	.tip-wrap:hover::before { opacity: 1; }

	/* ── Auth ── */
	.user-block {
		display: flex; align-items: center; gap: 8px;
		padding: 6px 10px;
		border: 1px solid rgba(168,200,228,0.22);
		border-radius: 8px;
		background: rgba(168,200,228,0.06);
	}
	.user-info { display: flex; flex-direction: column; line-height: 1; }
	.user-label { font-size: 9px; letter-spacing: 0.22em; color: #5a606a; }
	.user-name {
		font-family: 'Space Grotesk', sans-serif;
		font-size: 13px; font-weight: 600;
		color: #a8c8e4; max-width: 120px;
		overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
	}
	.logout-btn {
		background: none; border: 1px solid rgba(255,255,255,0.10);
		border-radius: 6px; padding: 4px 7px;
		font-size: 13px; color: #8a909a; cursor: pointer;
		transition: color 120ms, border-color 120ms;
	}
	.logout-btn:hover { color: #e8eaed; border-color: rgba(255,255,255,0.22); }

	.login-btn {
		padding: 7px 13px; border-radius: 8px;
		font-size: 10px; font-weight: 700; letter-spacing: 0.22em;
		color: #0e1014; text-decoration: none;
		background: linear-gradient(135deg, #a8c8e4, #6a7280);
		box-shadow: 0 0 14px rgba(168,200,228,0.15);
		transition: filter 120ms;
	}
	.login-btn:hover { filter: brightness(1.08); }
</style>
