<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { refreshStats } from '$lib/stores/stats';
	import GFrame from '$lib/components/galerie/GFrame.svelte';
	import GHeader from '$lib/components/galerie/GHeader.svelte';

	let { children } = $props();

	onMount(() => {
		refreshStats();
	});

	const routeToActive: Record<string, string> = {
		'/': 'Vitrine',
		'/vendre': 'Vendre',
		'/marche': 'Marché',
		'/profil': 'Profil',
		'/quetes': 'Quêtes',
		'/ligue': 'Ligue',
		'/drops': 'Drops',
		'/journal': 'Journal'
	};

	const active = $derived(routeToActive[$page.url.pathname] ?? 'Vitrine');
</script>

<svelte:head>
	<title>Collector.shop</title>
</svelte:head>

<GFrame>
	<GHeader {active} />
	<main class="g-main">
		{@render children()}
	</main>
</GFrame>

<style>
	.g-main {
		padding-top: 26px;
	}
</style>
