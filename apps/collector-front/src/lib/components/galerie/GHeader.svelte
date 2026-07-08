<script lang="ts">
	import { auth } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { unreadMessagesCount } from '$lib/stores/messages';
	import { cartCount } from '$lib/stores/cart';
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
		...($auth.user ? [{ label: 'Messages', href: '/messages' }] : []),
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
				{#if item.label === 'Messages' && $unreadMessagesCount > 0}
					<span class="g-nav-badge">{$unreadMessagesCount > 99 ? '99+' : $unreadMessagesCount}</span
					>
				{/if}
			</a>
		{/each}
	</nav>

	<div class="g-user">
		<a href="/panier" class="g-cart" title="Panier" aria-label="Panier">
			<span class="g-cart-ico">🛒</span>
			{#if $cartCount > 0}
				<span class="g-nav-badge g-cart-badge">{$cartCount > 99 ? '99+' : $cartCount}</span>
			{/if}
		</a>
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
	.g-nav-badge {
		display: inline-block;
		min-width: 15px;
		padding: 1px 4px;
		margin-left: 4px;
		border-radius: 999px;
		background: #86b3a4;
		color: #191714;
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 9px;
		font-weight: 700;
		text-align: center;
		vertical-align: middle;
	}

	/* User strip */
	.g-user {
		display: flex;
		align-items: center;
		gap: 14px;
		flex-shrink: 0;
	}
	.g-cart {
		position: relative;
		display: flex;
		align-items: center;
		justify-content: center;
		width: 34px;
		height: 34px;
		border-radius: 6px;
		border: 1px solid rgba(236, 229, 218, 0.1);
		text-decoration: none;
		transition: border-color 120ms;
	}
	.g-cart:hover {
		border-color: rgba(236, 229, 218, 0.22);
	}
	.g-cart-ico {
		font-size: 14px;
	}
	.g-cart-badge {
		position: absolute;
		top: -6px;
		right: -6px;
		margin-left: 0;
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
