<script lang="ts">
	import GPanel from '$lib/components/galerie/GPanel.svelte';
	import Kicker from '$lib/components/galerie/Kicker.svelte';
</script>

<svelte:head><title>Conditions générales d'utilisation · Collector.shop</title></svelte:head>

<div class="legal-page">
	<Kicker>Le contrat</Kicker>
	<h1 class="legal-title">Conditions générales d'utilisation</h1>
	<p class="legal-updated">Dernière mise à jour : 9 juillet 2026</p>

	<GPanel style="margin-top:20px">
		<h2>1. Objet</h2>
		<p>
			Collector.shop est une place de marché permettant à ses membres d'acheter et de vendre des
			objets de collection entre particuliers. L'utilisation du site implique l'acceptation pleine
			et entière des présentes conditions.
		</p>

		<h2>2. Création de compte</h2>
		<p>
			L'inscription requiert un nom, une adresse email valide et un mot de passe. Chaque membre est
			responsable de la confidentialité de ses identifiants et de l'activité menée depuis son
			compte.
		</p>

		<h2>3. Annonces et transactions</h2>
		<ul>
			<li>Chaque annonce doit décrire fidèlement l'objet proposé à la vente.</li>
			<li>Toute annonce nouvellement publiée est soumise à une modération avant sa mise en ligne.</li>
			<li>Le vendeur choisit d'accepter ou de refuser une commande reçue sur l'une de ses annonces.</li>
			<li>Aucune donnée bancaire n'est collectée ni traitée par la plateforme.</li>
		</ul>

		<h2>4. Messagerie entre membres</h2>
		<p>
			La messagerie intégrée permet d'échanger au sujet d'une annonce. Le partage de coordonnées
			personnelles (email, téléphone) y est automatiquement bloqué afin de protéger les échanges et
			d'éviter les transactions hors plateforme.
		</p>

		<h2>5. Suspension et suppression de compte</h2>
		<p>
			En cas de manquement aux présentes conditions, un compte peut être suspendu par
			l'administration. Chaque membre peut par ailleurs supprimer définitivement son compte et ses
			données personnelles à tout moment depuis son <a href="/profil">profil</a> — voir notre
			<a href="/confidentialite">politique de confidentialité</a> pour le détail de ce traitement.
		</p>

		<h2>6. Responsabilité</h2>
		<p>
			Collector.shop met en relation acheteurs et vendeurs mais n'est pas partie aux transactions
			conclues entre membres. La plateforme s'efforce d'assurer la sécurité et la disponibilité du
			service, sans garantie de continuité absolue.
		</p>

		<h2>7. Contact</h2>
		<p>
			Pour toute question relative aux présentes conditions, contactez
			<a href="mailto:contact@collector.shop">contact@collector.shop</a>.
		</p>
	</GPanel>
</div>

<style>
	.legal-page {
		max-width: 760px;
		margin: 0 auto;
		padding: 24px 0 60px;
	}
	.legal-title {
		font-family: var(--f-serif);
		font-weight: 600;
		font-size: 34px;
		margin: 8px 0 4px;
		color: var(--c-text);
	}
	.legal-updated {
		font-family: var(--f-body);
		font-size: 12px;
		color: var(--c-text-muted);
		margin: 0;
	}
	:global(.legal-page h2) {
		font-family: var(--f-serif);
		font-size: 18px;
		font-weight: 600;
		color: var(--c-text);
		margin: 26px 0 8px;
	}
	:global(.legal-page h2:first-child) {
		margin-top: 0;
	}
	:global(.legal-page p),
	:global(.legal-page li) {
		font-family: var(--f-body);
		font-size: 13.5px;
		line-height: 1.6;
		color: var(--c-text-tertiary);
	}
	:global(.legal-page ul) {
		margin: 6px 0;
		padding-left: 20px;
	}
	:global(.legal-page a) {
		color: var(--c-ink);
		font-weight: 600;
	}
</style>
