<script lang="ts">
	import { env } from '$env/dynamic/public';
	import favicon from '$lib/assets/favicon.svg';
	import './layout.css';
	let { children } = $props();
	const isStaging = env.PUBLIC_ENV === 'staging';
</script>

<svelte:head><link rel="icon" href={favicon} /></svelte:head>

{#if isStaging}
	<div class="staging-banner">STAGING — environnement de test, pas la production</div>
{/if}
{@render children()}

<style>
	.staging-banner {
		position: sticky;
		top: 0;
		z-index: 9999;
		padding: 0.4rem 1rem;
		text-align: center;
		font-size: 0.85rem;
		font-weight: 600;
		background: #f5a623;
		color: #1a1a1a;
	}
</style>
