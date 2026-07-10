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
		Collector<span class="logo-dot">.</span>shop
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
			<svg
				class="g-cart-ico"
				aria-hidden="true"
				viewBox="0 0 24 24"
				fill="none"
				stroke="currentColor"
				stroke-width="2"
				stroke-linecap="round"
				stroke-linejoin="round"
			>
				<circle cx="9" cy="21" r="1" />
				<circle cx="20" cy="21" r="1" />
				<path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6" />
			</svg>
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
		gap: 32px;
		padding: 20px 48px;
		background: var(--c-surface);
		border-bottom: 1px solid var(--c-border);
		flex-wrap: wrap;
	}

	/* Logo */
	.logo {
		font-family: var(--f-serif);
		font-size: 24px;
		font-weight: 700;
		letter-spacing: -0.5px;
		color: var(--c-ink);
		text-decoration: none;
		flex-shrink: 0;
	}
	.logo-dot {
		color: var(--c-accent);
	}

	/* Nav */
	.g-nav {
		display: flex;
		gap: 24px;
		flex: 1;
		flex-wrap: wrap;
	}
	.g-nav-link {
		font-family: var(--f-body);
		font-size: 14px;
		font-weight: 500;
		color: var(--c-text-tertiary);
		text-decoration: none;
		transition: color 120ms;
	}
	.g-nav-link:hover {
		color: var(--c-accent);
	}
	.g-nav-link.active {
		color: var(--c-ink);
		font-weight: 600;
	}
	.g-nav-badge {
		display: inline-block;
		min-width: 15px;
		padding: 1px 4px;
		margin-left: 4px;
		border-radius: var(--r-pill);
		background: var(--c-accent);
		color: #fff;
		font-family: var(--f-body);
		font-size: 9px;
		font-weight: 700;
		text-align: center;
		vertical-align: middle;
	}

	/* User strip */
	.g-user {
		display: flex;
		align-items: center;
		gap: 18px;
		flex-shrink: 0;
	}
	.g-cart {
		position: relative;
		display: flex;
		align-items: center;
		justify-content: center;
		width: 22px;
		height: 22px;
		color: var(--c-text-tertiary);
		text-decoration: none;
		transition: color 120ms;
	}
	.g-cart:hover {
		color: var(--c-ink);
	}
	.g-cart-ico {
		width: 20px;
		height: 20px;
	}
	.g-cart-badge {
		position: absolute;
		top: -8px;
		right: -8px;
		margin-left: 0;
	}
	.g-avatar {
		width: 32px;
		height: 32px;
		border-radius: 50%;
		background: var(--c-ink);
		color: var(--c-bg);
		display: flex;
		align-items: center;
		justify-content: center;
		font-family: var(--f-body);
		font-size: 12px;
		font-weight: 600;
		flex-shrink: 0;
	}
	.g-logout {
		background: none;
		border: 1px solid var(--c-border);
		border-radius: 6px;
		padding: 4px 8px;
		font-size: 13px;
		color: var(--c-text-tertiary);
		cursor: pointer;
		transition:
			color 120ms,
			border-color 120ms;
	}
	.g-logout:hover {
		color: var(--c-ink);
		border-color: var(--c-ink);
	}
	.g-login-btn {
		font-family: var(--f-body);
		font-size: 13px;
		font-weight: 600;
		padding: 9px 18px;
		border-radius: 8px;
		background: var(--c-accent);
		color: #fff;
		text-decoration: none;
		transition: filter 120ms;
	}
	.g-login-btn:hover {
		filter: brightness(1.08);
	}
</style>
