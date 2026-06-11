<script lang="ts">
	import { onDestroy } from 'svelte';
	import { page } from '$app/stores';
	import { auth } from '$lib/stores/auth';
	import { notifications } from '$lib/stores/notifications';
	import GFrame from '$lib/components/galerie/GFrame.svelte';
	import GHeader from '$lib/components/galerie/GHeader.svelte';

	let { children } = $props();

	// Connexion WebSocket notifications liee a la session : ouverte au login,
	// fermee au logout ou en quittant le layout.
	let currentToken: string | null = null;
	const unsubAuth = auth.subscribe(($auth) => {
		if ($auth.token === currentToken) return;
		currentToken = $auth.token;
		if ($auth.token) {
			notifications.start($auth.token);
		} else {
			notifications.reset();
		}
	});

	onDestroy(() => {
		unsubAuth();
		notifications.stop();
	});

	const routeToActive: Record<string, string> = {
		'/': 'Vitrine',
		'/profil': 'Profil'
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
