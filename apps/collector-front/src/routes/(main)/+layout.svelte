<script lang="ts">
	import favicon from '$lib/assets/favicon.svg';
	import { auth } from '$lib/stores/auth';
	import { goto } from '$app/navigation';

	let { children } = $props();

	const nav = [
		{ href: '/', label: '← Collector.shop' },
		{ href: '/admin', label: 'Admin Catalogue' },
		{ href: '/dashboard', label: 'Tableau de bord' }
	];

	function logout() {
		auth.logout();
		goto('/login');
	}
</script>

<svelte:head><link rel="icon" href={favicon} /></svelte:head>

<nav class="sticky top-0 z-50 border-b border-slate-800 bg-slate-950 px-4 py-3 shadow-lg">
	<div class="mx-auto flex max-w-7xl items-center justify-between">
		<a href="/" class="flex items-center gap-2">
			<span
				class="rounded-full bg-orange-500 px-2 py-0.5 text-xs font-black tracking-widest text-white uppercase"
			>
				collector
			</span>
			<span class="text-sm font-bold text-white">.shop</span>
		</a>
		<ul class="flex items-center gap-1">
			{#each nav as item}
				<li>
					<a
						href={item.href}
						class="rounded-xl px-4 py-2 text-xs font-semibold tracking-widest text-slate-400 uppercase transition hover:bg-slate-800 hover:text-white"
					>
						{item.label}
					</a>
				</li>
			{/each}
			{#if $auth.user}
				<li class="ml-2 flex items-center gap-2">
					<span class="text-xs text-slate-500">{$auth.user.name}</span>
					<button
						onclick={logout}
						class="rounded-xl px-3 py-2 text-xs font-semibold tracking-widest text-red-400 uppercase transition hover:bg-slate-800"
					>
						Déconnexion
					</button>
				</li>
			{:else}
				<li>
					<a
						href="/login"
						class="rounded-xl px-4 py-2 text-xs font-semibold tracking-widest text-orange-400 uppercase transition hover:bg-slate-800"
					>
						Connexion
					</a>
				</li>
			{/if}
		</ul>
	</div>
</nav>

{@render children()}
