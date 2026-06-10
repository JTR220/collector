<script lang="ts">
	import GPanel from '$lib/components/galerie/GPanel.svelte';
	import GMeter from '$lib/components/galerie/GMeter.svelte';
	import Kicker from '$lib/components/galerie/Kicker.svelte';

	const steps = [
		{
			n: '01',
			t: 'Réception sécurisée',
			d: 'Votre pièce arrive sous boîtier antichoc tracé. Photos d’état à la réception, horodatées et archivées.'
		},
		{
			n: '02',
			t: 'Authentification',
			d: 'Vérification d’authenticité par nos experts partenaires : impression, hologrammes, numéros de série, provenance.'
		},
		{
			n: '03',
			t: 'Notation PSA / CGC',
			d: 'La pièce est notée selon le barème du grader choisi : centrage, coins, bords, surface. Note de 1 à 10.'
		},
		{
			n: '04',
			t: 'Mise sous slab',
			d: 'Encapsulation scellée avec étiquette certifiée. La note est définitive et consultable en ligne.'
		},
		{
			n: '05',
			t: 'Mise en vente',
			d: 'Le lot rejoint la vitrine avec sa cote actualisée. Vous suivez son delta de prix en temps réel.'
		}
	];

	const scale = [
		{
			grade: 'PSA 10 · Gem Mint',
			pct: 100,
			desc: 'Parfaite. Centrage 55/45 max, aucun défaut visible.'
		},
		{ grade: 'PSA 9 · Mint', pct: 90, desc: 'Un défaut mineur toléré (léger point d’impression).' },
		{ grade: 'PSA 8 · NM-Mint', pct: 80, desc: 'Coins nets, légère usure de surface possible.' },
		{ grade: 'PSA 7 · Near Mint', pct: 70, desc: 'Légère usure visible de près, rien de majeur.' },
		{
			grade: 'PSA 6 et moins',
			pct: 50,
			desc: 'Usure franche : plis, rayures, blanchiment des bords.'
		}
	];
</script>

<svelte:head><title>Grading · Collector.shop</title></svelte:head>

<section class="hero">
	<Kicker color="#86b3a4">Guide · authentification &amp; notation</Kicker>
	<h1 class="hero-title">Comment fonctionne<br />le grading.</h1>
	<p class="hero-lede">
		Chaque lot vendu sur Collector.shop est authentifié puis noté par un organisme indépendant (PSA
		pour les cartes, CGC pour les comics). Voici le parcours d'une pièce, de la réception à la mise
		en vente.
	</p>
</section>

<section class="steps-grid">
	{#each steps as step}
		<GPanel>
			<span class="step-n">{step.n}</span>
			<h2 class="step-title">{step.t}</h2>
			<p class="step-desc">{step.d}</p>
		</GPanel>
	{/each}
</section>

<GPanel style="margin-top:18px">
	<Kicker>Échelle de notation</Kicker>
	<div class="scale-list">
		{#each scale as s}
			<div class="scale-row">
				<span class="scale-grade">{s.grade}</span>
				<div style="flex:1"><GMeter value={s.pct} height={5} /></div>
				<span class="scale-desc">{s.desc}</span>
			</div>
		{/each}
	</div>
</GPanel>

<div class="cta-row">
	<a href="/" class="btn-primary">Parcourir les lots gradés</a>
	<a href="/drops" class="btn-link">Voir les prochains drops →</a>
</div>

<style>
	.hero {
		padding: 24px 0 26px;
		max-width: 640px;
	}
	.hero-title {
		font-family: 'Newsreader', Georgia, serif;
		font-weight: 500;
		font-size: clamp(34px, 5vw, 46px);
		line-height: 1.05;
		color: #ece5da;
		margin: 10px 0 14px;
	}
	.hero-lede {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 14.5px;
		color: #a39a8c;
		line-height: 1.6;
	}

	.steps-grid {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 12px;
	}
	@media (max-width: 900px) {
		.steps-grid {
			grid-template-columns: repeat(2, 1fr);
		}
	}
	@media (max-width: 580px) {
		.steps-grid {
			grid-template-columns: 1fr;
		}
	}

	.step-n {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 11px;
		color: #86b3a4;
		letter-spacing: 0.16em;
	}
	.step-title {
		font-family: 'Newsreader', Georgia, serif;
		font-size: 19px;
		font-weight: 500;
		color: #ece5da;
		margin: 8px 0 6px;
	}
	.step-desc {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 12.5px;
		color: #a39a8c;
		line-height: 1.55;
		margin: 0;
	}

	.scale-list {
		display: flex;
		flex-direction: column;
		gap: 12px;
		margin-top: 14px;
	}
	.scale-row {
		display: flex;
		align-items: center;
		gap: 14px;
	}
	.scale-grade {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 12px;
		color: #ece5da;
		width: 150px;
		flex-shrink: 0;
	}
	.scale-desc {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 11.5px;
		color: #766d60;
		width: 280px;
		flex-shrink: 0;
	}
	@media (max-width: 768px) {
		.scale-desc {
			display: none;
		}
	}

	.cta-row {
		display: flex;
		align-items: center;
		gap: 20px;
		margin-top: 24px;
	}
	.btn-primary {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13px;
		font-weight: 600;
		padding: 12px 22px;
		border-radius: 7px;
		background: #86b3a4;
		color: #191714;
		text-decoration: none;
		transition: filter 120ms;
	}
	.btn-primary:hover {
		filter: brightness(1.08);
	}
	.btn-link {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13px;
		color: #a39a8c;
		text-decoration: underline;
		text-underline-offset: 3px;
	}
	.btn-link:hover {
		color: #ece5da;
	}
</style>
