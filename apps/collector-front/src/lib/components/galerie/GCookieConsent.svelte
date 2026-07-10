<script lang="ts">
	import { browser } from '$app/environment';

	const STORAGE_KEY = 'collector_cookie_consent';

	// Le seul cookie pose par l'application est le cookie de session httpOnly
	// (authentification), strictement necessaire au fonctionnement du site :
	// il est exempte de consentement (art. 82 loi Informatique et Libertes /
	// directive ePrivacy). Ce bandeau est donc informatif — il n'existe pas
	// de cookie de mesure d'audience ou publicitaire a activer/desactiver.
	let visible = $state(browser ? localStorage.getItem(STORAGE_KEY) !== 'true' : false);

	function acknowledge() {
		if (browser) localStorage.setItem(STORAGE_KEY, 'true');
		visible = false;
	}
</script>

{#if visible}
	<div class="g-cookie-banner" role="dialog" aria-label="Information sur les cookies">
		<p class="g-cookie-text">
			Collector.shop utilise uniquement un cookie de session strictement necessaire a la
			connexion (aucun cookie de mesure d'audience ni publicitaire). En savoir plus dans notre
			<a href="/confidentialite">politique de confidentialité</a>.
		</p>
		<button class="g-cookie-btn" onclick={acknowledge}>J'ai compris</button>
	</div>
{/if}

<style>
	.g-cookie-banner {
		position: fixed;
		left: 0;
		right: 0;
		bottom: 0;
		z-index: 9998;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 20px;
		flex-wrap: wrap;
		padding: 14px 24px;
		background: var(--c-ink);
		color: var(--c-bg);
	}
	.g-cookie-text {
		margin: 0;
		font-family: var(--f-body);
		font-size: 12.5px;
		line-height: 1.5;
		max-width: 720px;
	}
	.g-cookie-text a {
		color: var(--c-bg);
		text-decoration: underline;
	}
	.g-cookie-btn {
		flex-shrink: 0;
		font-family: var(--f-body);
		font-size: 12.5px;
		font-weight: 600;
		padding: 8px 20px;
		border-radius: 7px;
		border: 1px solid var(--c-bg);
		background: transparent;
		color: var(--c-bg);
		cursor: pointer;
		transition:
			background 120ms,
			color 120ms;
	}
	.g-cookie-btn:hover {
		background: var(--c-bg);
		color: var(--c-ink);
	}
</style>
