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

<div class="m-frame">
	{#if !hideHeader}
		<GHeader {active} />
	{/if}

	<div class="m-body" class:m-body-padded={!hideHeader}>
		{@render children()}
	</div>
</div>

<style>
	.m-frame {
		width: 100%;
		min-height: 100vh;
		background: var(--c-bg);
		color: var(--c-text);
		font-family: var(--f-body);
		box-sizing: border-box;
	}
	.m-body-padded {
		max-width: 1440px;
		margin: 0 auto;
		padding: 40px 48px;
	}
	@media (max-width: 640px) {
		.m-body-padded {
			padding: 24px 20px;
		}
	}
</style>
