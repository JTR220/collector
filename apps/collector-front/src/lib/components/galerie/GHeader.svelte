<script lang="ts">
	import { auth } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import GNotifBell from './GNotifBell.svelte';

	type Props = {
		active?: string;
	};
	let { active = 'Marché' }: Props = $props();

	// « Vendre » n'apparait que pour un utilisateur connecte ; « Admin » uniquement
	// pour les administrateurs, et reste visible sur toutes les pages holo.
	const nav = $derived([
		{ label: 'Marché', href: '/' },
		...($auth.user ? [{ label: 'Vendre', href: '/vendre' }] : []),
		{ label: 'Profil', href: '/profil' },
		...($auth.user?.role === 'admin'
			? [
					{ label: 'Admin', href: '/admin' },
					{ label: 'Tableau de bord', href: '/dashboard' }
				]
			: [])
	]);

	const initials = $derived(
		$auth.user?.name
			? $auth.user.name
					.split(' ')
					.map((w: string) => w[0])
					.join('')
					.toUpperCase()
					.slice(0, 2)
			: 'NK'
	);

	function logout() {
		auth.logout();
		goto('/login');
	}
</script>

<header class="g-header">
	<a href="/" class="logo">
		<span class="logo-diamond"></span>
		<span class="logo-text">Collector<span class="logo-dim">.shop</span></span>
	</a>

	<nav class="g-nav">
		{#each nav as item}
			<a href={item.href} class="g-nav-link" class:active={item.label === active}>
				{item.label}
			</a>
		{/each}
	</nav>

	<div class="g-user">
		{#if $auth.user}
			<GNotifBell />
			<div class="g-avatar">{initials}</div>
			<button class="g-logout" onclick={logout} title="Se déconnecter">↩</button>
		{:else}
			<a href="/login" class="g-login-btn">Se connecter</a>
		{/if}
	</div>
</header>

<style>
	.g-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 24px;
		padding-bottom: 16px;
		border-bottom: 1px solid rgba(236, 229, 218, 0.1);
		margin-bottom: 0;
		flex-wrap: wrap;
	}

	/* Logo */
	.logo {
		display: flex;
		align-items: center;
		gap: 10px;
		text-decoration: none;
		flex-shrink: 0;
	}
	.logo-diamond {
		width: 9px;
		height: 9px;
		background: #86b3a4;
		border-radius: 2px;
		transform: rotate(45deg);
		display: inline-block;
		flex-shrink: 0;
	}
	.logo-text {
		font-family: 'Newsreader', Georgia, serif;
		font-size: 20px;
		letter-spacing: 0.01em;
		color: #ece5da;
	}
	.logo-dim {
		color: #766d60;
	}

	/* Nav */
	.g-nav {
		display: flex;
		gap: 26px;
	}
	.g-nav-link {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13.5px;
		font-weight: 400;
		color: #a39a8c;
		text-decoration: none;
		border-bottom: 1.5px solid transparent;
		padding-bottom: 4px;
		transition:
			color 120ms,
			border-color 120ms;
	}
	.g-nav-link:hover {
		color: #ece5da;
	}
	.g-nav-link.active {
		color: #ece5da;
		font-weight: 600;
		border-bottom-color: #86b3a4;
	}

	/* User strip */
	.g-user {
		display: flex;
		align-items: center;
		gap: 14px;
		flex-shrink: 0;
	}
	.g-avatar {
		width: 34px;
		height: 34px;
		border-radius: 50%;
		border: 1px solid rgba(236, 229, 218, 0.16);
		background: rgba(255, 255, 255, 0.05);
		color: #86b3a4;
		display: flex;
		align-items: center;
		justify-content: center;
		font-family: 'Newsreader', Georgia, serif;
		font-size: 14px;
		flex-shrink: 0;
	}
	.g-logout {
		background: none;
		border: 1px solid rgba(236, 229, 218, 0.1);
		border-radius: 6px;
		padding: 4px 8px;
		font-size: 13px;
		color: #a39a8c;
		cursor: pointer;
		transition:
			color 120ms,
			border-color 120ms;
	}
	.g-logout:hover {
		color: #ece5da;
		border-color: rgba(236, 229, 218, 0.22);
	}
	.g-login-btn {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 12.5px;
		font-weight: 600;
		padding: 9px 18px;
		border-radius: 7px;
		background: #86b3a4;
		color: #191714;
		text-decoration: none;
		transition: filter 120ms;
	}
	.g-login-btn:hover {
		filter: brightness(1.08);
	}
</style>
