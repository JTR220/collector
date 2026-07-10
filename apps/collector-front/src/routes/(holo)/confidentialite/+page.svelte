<script lang="ts">
	import GPanel from '$lib/components/galerie/GPanel.svelte';
	import Kicker from '$lib/components/galerie/Kicker.svelte';
</script>

<svelte:head><title>Politique de confidentialité · Collector.shop</title></svelte:head>

<div class="legal-page">
	<Kicker>Vos données</Kicker>
	<h1 class="legal-title">Politique de confidentialité</h1>
	<p class="legal-updated">Dernière mise à jour : 9 juillet 2026</p>

	<GPanel style="margin-top:20px">
		<h2>1. Qui traite vos données ?</h2>
		<p>
			Collector.shop est une place de marché dédiée aux objets de collection. Le responsable du
			traitement des données décrites ci-dessous est l'éditeur du site, contactable à
			<a href="mailto:contact@collector.shop">contact@collector.shop</a>.
		</p>

		<h2>2. Quelles données sont collectées</h2>
		<ul>
			<li>
				<strong>Compte</strong> : nom, adresse email, mot de passe (stocké haché, jamais en clair).
			</li>
			<li>
				<strong>Activité sur le marché</strong> : annonces publiées, achats, ventes, wishlist, avis laissés.
			</li>
			<li>
				<strong>Messagerie</strong> : messages échangés entre acheteurs et vendeurs au sujet d'une annonce.
				Un filtre automatique bloque l'envoi de coordonnées personnelles (email, téléphone) dans ces messages,
				pour éviter les échanges hors plateforme.
			</li>
			<li>
				<strong>Notifications</strong> : alertes de prix, de commande ou de fraude générées par votre
				activité.
			</li>
		</ul>
		<p>
			Nous ne collectons pas d'adresse postale, de numéro de téléphone ni de moyen de paiement :
			aucune donnée bancaire ne transite par nos serveurs.
		</p>

		<h2>3. Pourquoi ces données sont traitées</h2>
		<ul>
			<li>Créer et gérer votre compte, vous authentifier (exécution du contrat de service).</li>
			<li>Permettre la publication d'annonces et le suivi des commandes.</li>
			<li>
				Vous envoyer les notifications liées à votre activité (email transactionnel, temps réel).
			</li>
			<li>
				Assurer la sécurité de la plateforme (limitation des tentatives de connexion, détection de
				fraude sur les prix).
			</li>
		</ul>

		<h2>4. Durée de conservation</h2>
		<ul>
			<li>
				<strong>Données de compte</strong> (nom, email, mot de passe) : conservées tant que le compte
				existe, supprimées définitivement à la clôture du compte.
			</li>
			<li>
				<strong>Messages et notifications</strong> : conservés 12 mois, puis supprimés automatiquement
				par une purge périodique.
			</li>
		</ul>

		<h2>5. Cookies</h2>
		<p>
			Le site utilise uniquement un cookie de session strictement nécessaire à l'authentification (<code
				>httpOnly</code
			>, non accessible en JavaScript). Aucun cookie de mesure d'audience ni publicitaire n'est
			déposé.
		</p>

		<h2>6. Vos droits</h2>
		<p>Conformément au RGPD, vous disposez des droits suivants sur vos données :</p>
		<ul>
			<li>
				<strong>Accès et rectification</strong> : modifiez votre nom et votre email directement
				depuis votre <a href="/profil">profil</a>.
			</li>
			<li>
				<strong>Portabilité</strong> : téléchargez une copie de vos données personnelles depuis votre
				profil.
			</li>
			<li>
				<strong>Effacement</strong> : supprimez définitivement votre compte et vos données personnelles
				depuis votre profil, à tout moment.
			</li>
			<li>
				Pour toute autre demande (opposition, limitation), contactez <a
					href="mailto:contact@collector.shop">contact@collector.shop</a
				>.
			</li>
		</ul>

		<h2>7. Destinataires des données</h2>
		<p>
			Vos données ne sont jamais vendues ni partagées avec des tiers à des fins commerciales. Elles
			sont traitées uniquement par les services internes de la plateforme (authentification,
			catalogue, notifications) et, pour l'envoi d'email transactionnel, par le serveur SMTP
			configuré par l'éditeur.
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
	:global(.legal-page code) {
		font-size: 12px;
		background: var(--c-bg);
		border: 1px solid var(--c-border);
		border-radius: 4px;
		padding: 1px 5px;
	}
</style>
