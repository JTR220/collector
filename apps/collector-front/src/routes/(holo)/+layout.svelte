<script lang="ts">
	import { onDestroy } from 'svelte';
	import { page } from '$app/stores';
	import { auth } from '$lib/stores/auth';
	import { notifications } from '$lib/stores/notifications';
	import { messages } from '$lib/stores/messages';
	import GFrame from '$lib/components/galerie/GFrame.svelte';
	import GHeader from '$lib/components/galerie/GHeader.svelte';

	let { children } = $props();

	// Connexion WebSocket notifications liee a la session : ouverte au login,
	// fermee au logout ou en quittant le layout. L'authentification passe par
	// le cookie httpOnly (envoye automatiquement) : plus besoin de token cote JS.
	let wasAuthenticated = false;
	const unsubAuth = auth.subscribe(($auth) => {
		const isAuth = !!$auth.user;
		if (isAuth === wasAuthenticated) return;
		wasAuthenticated = isAuth;
		if (isAuth) {
			notifications.start();
			messages.start();
		} else {
			notifications.reset();
			messages.reset();
		}
	});

	onDestroy(() => {
		unsubAuth();
		notifications.stop();
		messages.stop();
	});

	const routeToActive: Record<string, string> = {
		'/': 'Marché',
		'/vendre': 'Vendre',
		'/profil': 'Profil',
		'/messages': 'Messages',
		'/admin': 'Admin',
		'/dashboard': 'Tableau de bord'
	};

	const active = $derived(routeToActive[$page.url.pathname] ?? 'Marché');
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
		max-width: 1440px;
		margin: 0 auto;
		padding: 0 48px 40px;
	}
	@media (max-width: 640px) {
		.g-main {
			padding: 0 20px 32px;
		}
	}
</style>
