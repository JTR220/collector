<script lang="ts">
	import { page } from '$app/stores';
	import HoloFrame from '$lib/components/holo/HoloFrame.svelte';
	import HoloHeader from '$lib/components/holo/HoloHeader.svelte';

	let { children } = $props();

	const routeToActive: Record<string, string> = {
		'/vitrine': 'VITRINE',
		'/profil': 'PROFIL',
		'/quetes': 'QUÊTES',
		'/ligue': 'LIGUE',
		'/drops': 'DROPS',
		'/journal': 'JOURNAL'
	};

	const active = $derived(routeToActive[$page.url.pathname] ?? 'VITRINE');
</script>

<svelte:head>
	<title>Collector.shop</title>
</svelte:head>

<HoloFrame>
	<HoloHeader {active} />
	{@render children()}
</HoloFrame>
