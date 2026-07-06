<script lang="ts">
	import favicon from '$lib/assets/favicon.svg';
	import { page } from '$app/stores';
	import GHeader from '$lib/components/galerie/GHeader.svelte';

	let { children } = $props();

	// Le header est masque sur la page de connexion (mise en page plein ecran dediee).
	const hideHeader = $derived($page.url.pathname === '/login');

	const routeToActive: Record<string, string> = {
		'/admin': 'Admin',
		'/dashboard': 'Tableau de bord'
	};
	const active = $derived(routeToActive[$page.url.pathname] ?? '');
</script>

<svelte:head><link rel="icon" href={favicon} /></svelte:head>

<div class="m-frame" class:m-frame-padded={!hideHeader}>
	{#if !hideHeader}
		<GHeader {active} />
	{/if}

	{@render children()}
</div>

<style>
	.m-frame {
		width: 100%;
		min-height: 100vh;
		background: #191714;
		color: #ece5da;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		box-sizing: border-box;
	}
	.m-frame-padded {
		padding: 26px 38px 40px;
	}
</style>
