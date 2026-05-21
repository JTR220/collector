<script lang="ts">
	import favicon from '$lib/assets/favicon.svg';
	let { children } = $props();

	const nav = [
		{ href: '/', label: 'Catalogue' },
		{ href: '/dashboard', label: 'Tableau de bord' },
		{ href: '/login', label: 'Connexion' },
		{ href: '/vitrine', label: 'Vitrine Holo' }
	];
</script>

<svelte:head><link rel="icon" href={favicon} /></svelte:head>

<nav class="sticky top-0 z-50 border-b border-slate-800 bg-slate-950 px-4 py-3 shadow-lg">
	<div class="mx-auto flex max-w-7xl items-center justify-between">
		<a href="/" class="flex items-center gap-2">
			<span class="rounded-full bg-orange-500 px-2 py-0.5 text-xs font-black uppercase tracking-widest text-white">
				collector
			</span>
			<span class="text-sm font-bold text-white">.shop</span>
		</a>
		<ul class="flex gap-1">
			{#each nav as item}
				<li>
					<a
						href={item.href}
						class="rounded-xl px-4 py-2 text-xs font-semibold uppercase tracking-widest text-slate-400 transition hover:bg-slate-800 hover:text-white"
					>
						{item.label}
					</a>
				</li>
			{/each}
		</ul>
	</div>
</nav>

{@render children()}
