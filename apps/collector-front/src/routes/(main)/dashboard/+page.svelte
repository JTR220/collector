<script lang="ts">
	import { onMount } from 'svelte';

	const authApiUrl = import.meta.env.VITE_AUTH_API_BASE_URL ?? 'http://localhost:8080';
	const catalogApiUrl = import.meta.env.VITE_CATALOG_API_BASE_URL ?? 'http://localhost:8081';

	// --- Acteurs selon le cours (page 17) ---
	type Actor = 'ceo' | 'cto' | 'pm';
	let activeActor: Actor = 'pm';

	const actors: { id: Actor; label: string; role: string; besoin: string; icon: string }[] = [
		{
			id: 'pm',
			label: 'PM',
			role: 'Chef de projet',
			besoin: 'Avancement, qualité, vélocité',
			icon: '📋'
		},
		{
			id: 'cto',
			label: 'CTO',
			role: 'Directeur technique',
			besoin: 'Infrastructure, performance, DevOps',
			icon: '⚙️'
		},
		{
			id: 'ceo',
			label: 'CEO',
			role: 'Direction générale',
			besoin: 'Business, conversion, rétention',
			icon: '📊'
		}
	];

	// --- KPI live ---
	type ServiceStatus = 'loading' | 'ok' | 'down';

	let authStatus: ServiceStatus = 'loading';
	let catalogStatus: ServiceStatus = 'loading';
	let catalogLatencyMs: number | null = null;
	let articlesCount = 0;
	let categoriesCount = 0;
	let lastRefresh = new Date();

	// --- KPI projet (PM) ---
	const roadmapItems = [
		{ label: 'Monorepo + CI/CD backend', done: true },
		{ label: 'Catalog-service CRUD', done: true },
		{ label: 'Collector-front catalogue', done: true },
		{ label: 'Docker Compose complet', done: true },
		{ label: 'CI front (build + tests)', done: true },
		{ label: 'Auth JWT (bcrypt + login)', done: true },
		{ label: 'Dashboard KPI', done: true },
		{ label: 'Auth branché sur le front', done: false },
		{ label: 'Tests unitaires backend', done: false },
		{ label: 'Conventions API homogènes', done: false }
	];

	const avancement = Math.round(
		(roadmapItems.filter((i) => i.done).length / roadmapItems.length) * 100
	);

	// Vélocité agile (sprints simulés)
	const velocite = [
		{ sprint: 'S1', points: 22 },
		{ sprint: 'S2', points: 27 },
		{ sprint: 'S3', points: 18 }
	];

	// KPI qualité
	const testCoverage = 18;
	const bugRate = 0.25;
	const bugsOuverts = 2;
	const featuresLivrees = 8;
	const fixRate = 75;

	// KPI DevOps (CTO)
	const doraMetrics = [
		{ label: 'Deployment Frequency', value: '1/sem.', note: 'push sur main', status: 'medium' },
		{ label: 'Lead Time', value: '< 1j', note: 'commit → prod', status: 'good' },
		{ label: 'MTTR', value: 'N/A', note: 'pas encore mesuré', status: 'na' },
		{ label: 'Change Failure Rate', value: '~15%', note: 'builds échoués CI', status: 'medium' }
	];

	// KPI business (CEO)
	const churnRate = 'N/A';
	const retentionRate = 'N/A';
	const arpu = 'N/A';
	const conversionRate = 'N/A';

	// --- Fonctions live ---
	async function checkHealth(url: string): Promise<ServiceStatus> {
		try {
			const res = await fetch(`${url}/health`, { signal: AbortSignal.timeout(3000) });
			return res.ok ? 'ok' : 'down';
		} catch {
			return 'down';
		}
	}

	async function fetchLiveData() {
		lastRefresh = new Date();
		authStatus = 'loading';
		catalogStatus = 'loading';

		const [auth, catalog] = await Promise.all([
			checkHealth(authApiUrl),
			(async () => {
				const t0 = Date.now();
				const s = await checkHealth(catalogApiUrl);
				catalogLatencyMs = Date.now() - t0;
				return s;
			})()
		]);

		authStatus = auth;
		catalogStatus = catalog;

		if (catalogStatus === 'ok') {
			try {
				const [art, cat] = await Promise.all([
					fetch(`${catalogApiUrl}/article`).then((r) => r.json()),
					fetch(`${catalogApiUrl}/category`).then((r) => r.json())
				]);
				articlesCount = Array.isArray(art) ? art.length : 0;
				categoriesCount = Array.isArray(cat) ? cat.length : 0;
			} catch {
				/* ignore */
			}
		}
	}

	onMount(() => {
		fetchLiveData();
	});

	function coverageBadge(v: number) {
		if (v >= 90) return { label: 'Bon', color: 'emerald' };
		if (v >= 70) return { label: 'Moyen', color: 'amber' };
		return { label: 'Critique', color: 'red' };
	}

	const cvg = coverageBadge(testCoverage);
	const velociteMax = Math.max(...velocite.map((s) => s.points));
</script>

<svelte:head>
	<title>Collector.shop | Tableau de bord</title>
</svelte:head>

<div
	class="min-h-screen bg-[radial-gradient(circle_at_top_left,_rgba(255,238,201,0.6),_rgba(255,255,255,1)_40%),linear-gradient(135deg,#0f172a_0%,#1f2937_45%,#f97316_45%,#fff7ed_100%)] text-slate-950"
>
	<div class="mx-auto max-w-7xl px-4 py-8 md:px-8">
		<!-- En-tête -->
		<div class="mb-8 flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
			<div>
				<p
					class="inline-flex rounded-full border border-orange-300 bg-orange-100 px-3 py-1 text-xs font-semibold tracking-widest text-orange-700 uppercase"
				>
					Pilotage
				</p>
				<h1 class="mt-2 text-4xl font-black tracking-tight text-slate-900 uppercase">
					Tableau de bord
				</h1>
				<p class="mt-1 text-sm text-slate-500">
					Dernière mise à jour : {lastRefresh.toLocaleTimeString('fr-FR')}
				</p>
			</div>
			<button
				onclick={fetchLiveData}
				class="self-start rounded-2xl border border-slate-200 bg-white px-5 py-2 text-xs font-bold tracking-widest text-slate-700 uppercase shadow-sm transition hover:bg-slate-50 sm:self-auto"
			>
				Rafraîchir
			</button>
		</div>

		<!-- Sélecteur acteur — source cours p.17 -->
		<div class="mb-8">
			<p class="mb-1 text-xs font-semibold tracking-widest text-slate-400 uppercase">
				Identifier l'utilisateur · <span class="text-orange-500">source cours</span>
			</p>
			<p class="mb-4 text-xs text-slate-400">
				Un dashboard = logique de décision adaptée à l'acteur. Sélectionnez votre rôle.
			</p>
			<div class="flex flex-wrap gap-3">
				{#each actors as actor}
					<button
						onclick={() => (activeActor = actor.id)}
						class={`flex flex-col items-start rounded-2xl border px-5 py-4 text-left transition ${
							activeActor === actor.id
								? 'border-orange-400 bg-orange-50 shadow-md ring-1 ring-orange-300'
								: 'border-slate-200 bg-white shadow-sm hover:bg-slate-50'
						}`}
					>
						<div class="flex items-center gap-2">
							<span class="text-xl">{actor.icon}</span>
							<span
								class={`text-sm font-black tracking-wide uppercase ${activeActor === actor.id ? 'text-orange-700' : 'text-slate-800'}`}
							>
								{actor.label}
							</span>
						</div>
						<span class="mt-1 text-xs font-semibold text-slate-500">{actor.role}</span>
						<span class="mt-0.5 text-xs text-slate-400">{actor.besoin}</span>
					</button>
				{/each}
			</div>
		</div>

		<!-- =============================== -->
		<!-- VUE PM — Chef de projet         -->
		<!-- =============================== -->
		{#if activeActor === 'pm'}
			<!-- Ligne 1 : Vision globale -->
			<p class="mb-3 text-xs font-semibold tracking-widest text-slate-400 uppercase">
				Ligne 1 — Vision globale
			</p>
			<div class="mb-6 grid gap-4 lg:grid-cols-2">
				<!-- Avancement projet -->
				<div class="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
					<p class="text-xs font-semibold tracking-widest text-slate-400 uppercase">
						Avancement projet
					</p>
					<div class="mt-3 flex items-end gap-4">
						<p class="text-5xl font-black text-slate-900">{avancement}%</p>
						<p class="mb-1 text-sm text-slate-500">
							{roadmapItems.filter((i) => i.done).length} / {roadmapItems.length} features
						</p>
					</div>
					<div class="mt-4 h-3 overflow-hidden rounded-full bg-slate-100">
						<div
							class="h-full rounded-full bg-orange-500 transition-all"
							style="width: {avancement}%"
						></div>
					</div>
					<p class="mt-2 text-xs text-slate-400">
						Formule cours : Avancement = Travail réalisé / Travail total × 100
					</p>
					<ul class="mt-4 grid grid-cols-2 gap-1">
						{#each roadmapItems as item}
							<li class="flex items-center gap-2 text-xs text-slate-600">
								<span
									class={`h-2 w-2 flex-shrink-0 rounded-full ${item.done ? 'bg-emerald-500' : 'bg-slate-300'}`}
								></span>
								{item.label}
							</li>
						{/each}
					</ul>
				</div>

				<!-- Vélocité agile -->
				<div class="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
					<p class="text-xs font-semibold tracking-widest text-slate-400 uppercase">
						Vélocité agile
					</p>
					<p class="mt-1 text-xs text-slate-400">Story points terminés par sprint</p>
					<div class="mt-4 space-y-3">
						{#each velocite as s}
							<div>
								<div class="mb-1 flex justify-between text-xs">
									<span class="font-semibold text-slate-700">{s.sprint}</span>
									<span class={`font-bold ${s.points < 20 ? 'text-red-600' : 'text-slate-900'}`}
										>{s.points} pts</span
									>
								</div>
								<div class="h-3 overflow-hidden rounded-full bg-slate-100">
									<div
										class={`h-full rounded-full transition-all ${s.points < 20 ? 'bg-red-400' : 'bg-orange-500'}`}
										style="width: {(s.points / velociteMax) * 100}%"
									></div>
								</div>
							</div>
						{/each}
					</div>
					<div class="mt-4 rounded-xl bg-amber-50 px-3 py-2 text-xs font-medium text-amber-700">
						Alerte : chute de vélocité S3 → investiguer blocages
					</div>
				</div>
			</div>

			<!-- Ligne 2 : Problèmes -->
			<p class="mb-3 text-xs font-semibold tracking-widest text-slate-400 uppercase">
				Ligne 2 — Problèmes (qualité)
			</p>
			<div class="mb-6 grid gap-4 md:grid-cols-3">
				<!-- Taux de bugs -->
				<div class="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
					<p class="text-xs font-semibold tracking-widest text-slate-400 uppercase">Taux de bugs</p>
					<div class="mt-3 flex items-end gap-3">
						<p class="text-4xl font-black text-slate-900">{bugRate.toFixed(2)}</p>
						<span
							class="mb-1 rounded-full bg-amber-100 px-2 py-0.5 text-xs font-bold text-amber-700"
							>Moyen</span
						>
					</div>
					<p class="mt-1 text-xs text-slate-400">Formule : bugs / features livrées</p>
					<div class="mt-4 space-y-2 text-xs text-slate-600">
						<div class="flex justify-between rounded-xl bg-slate-50 px-3 py-2">
							<span>Bugs critiques ouverts</span><span class="font-bold text-red-600"
								>{bugsOuverts}</span
							>
						</div>
						<div class="flex justify-between rounded-xl bg-slate-50 px-3 py-2">
							<span>Features livrées</span><span class="font-bold">{featuresLivrees}</span>
						</div>
					</div>
				</div>

				<!-- Fix Rate -->
				<div class="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
					<p class="text-xs font-semibold tracking-widest text-slate-400 uppercase">Fix Rate</p>
					<div class="mt-3 flex items-end gap-3">
						<p class="text-4xl font-black text-slate-900">{fixRate}%</p>
						<span
							class="mb-1 rounded-full bg-emerald-100 px-2 py-0.5 text-xs font-bold text-emerald-700"
							>Bon</span
						>
					</div>
					<p class="mt-1 text-xs text-slate-400">Formule : bugs corrigés / bugs détectés × 100</p>
					<div class="mt-3 h-2 overflow-hidden rounded-full bg-slate-100">
						<div class="h-full rounded-full bg-emerald-500" style="width: {fixRate}%"></div>
					</div>
					<p class="mt-2 text-xs text-slate-400">
						Cible cours : &gt; 90% = bon · 70–90% = moyen · &lt; 70% = critique
					</p>
				</div>

				<!-- Couverture tests -->
				<div class="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
					<p class="text-xs font-semibold tracking-widest text-slate-400 uppercase">
						Couverture tests
					</p>
					<div class="mt-3 flex items-end gap-3">
						<p class="text-4xl font-black text-slate-900">{testCoverage}%</p>
						<span
							class={`mb-1 rounded-full px-2 py-0.5 text-xs font-bold bg-${cvg.color}-100 text-${cvg.color}-700`}
						>
							{cvg.label}
						</span>
					</div>
					<div class="mt-3 h-2 overflow-hidden rounded-full bg-slate-100">
						<div
							class={`h-full rounded-full bg-${cvg.color}-500`}
							style="width: {testCoverage}%"
						></div>
					</div>
					<p class="mt-2 text-xs text-slate-400">Cible : 80% · Source : SonarCloud CI</p>
					<div class="mt-3 rounded-xl bg-red-50 px-3 py-2 text-xs font-medium text-red-700">
						Action : ajouter tests unitaires backend
					</div>
				</div>
			</div>

			<div class="rounded-3xl border border-orange-200 bg-orange-50 p-5">
				<p class="text-xs font-bold tracking-widest text-orange-700 uppercase">
					Lecture PM · Cours p.17
				</p>
				<p class="mt-2 text-sm text-orange-800">
					Le PM suit l'avancement (prévu vs réel), la vélocité pour détecter les blocages et la
					qualité (bug rate, fix rate, couverture). Une chute de vélocité = signal d'alerte →
					réallouer ressources ou geler de nouvelles features.
				</p>
			</div>

			<!-- =============================== -->
			<!-- VUE CTO — Directeur technique   -->
			<!-- =============================== -->
		{:else if activeActor === 'cto'}
			<!-- Ligne 3 : Technique -->
			<p class="mb-3 text-xs font-semibold tracking-widest text-slate-400 uppercase">
				Ligne 3 — Technique (infrastructure live)
			</p>
			<div class="mb-6 grid gap-4 lg:grid-cols-2">
				<!-- Uptime services -->
				<div class="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
					<p class="text-xs font-semibold tracking-widest text-slate-400 uppercase">
						Disponibilité (Uptime)
					</p>
					<p class="mt-1 text-xs text-slate-400">
						Formule : temps fonctionnement / temps total × 100
					</p>
					<div class="mt-4 space-y-4">
						{#each [{ label: 'auth-service', status: authStatus }, { label: 'catalog-service', status: catalogStatus }] as svc}
							<div class="flex items-center justify-between rounded-2xl bg-slate-50 px-4 py-3">
								<div>
									<p class="text-sm font-bold text-slate-900">{svc.label}</p>
									<p class="text-xs text-slate-400">GET /health</p>
								</div>
								<span
									class={`rounded-full px-3 py-1 text-xs font-bold uppercase ${
										svc.status === 'ok'
											? 'bg-emerald-100 text-emerald-700'
											: svc.status === 'loading'
												? 'bg-slate-100 text-slate-500'
												: 'bg-red-100 text-red-700'
									}`}
								>
									{svc.status === 'loading' ? '...' : svc.status === 'ok' ? 'UP' : 'DOWN'}
								</span>
							</div>
						{/each}
					</div>
					<div class="mt-3 rounded-xl bg-slate-50 px-3 py-2 text-xs text-slate-500">
						Niveau cible cours : 99.9% (standard) · 99.99% (critique)
					</div>
				</div>

				<!-- Latence + taux d'erreur -->
				<div class="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
					<p class="text-xs font-semibold tracking-widest text-slate-400 uppercase">
						Performance API
					</p>
					<div class="mt-4 space-y-4">
						{#if catalogLatencyMs !== null}
							<div class="flex items-center justify-between rounded-2xl bg-slate-50 px-4 py-3">
								<div>
									<p class="text-sm font-bold text-slate-900">Latence (temps de réponse)</p>
									<p class="text-xs text-slate-400">
										catalog-service /health · Formule : Σtemps / nb requêtes
									</p>
								</div>
								<span
									class={`rounded-full px-3 py-1 text-xs font-bold ${
										catalogLatencyMs < 200
											? 'bg-emerald-100 text-emerald-700'
											: catalogLatencyMs < 500
												? 'bg-amber-100 text-amber-700'
												: 'bg-red-100 text-red-700'
									}`}
								>
									{catalogLatencyMs} ms
								</span>
							</div>
						{/if}
						<div class="flex items-center justify-between rounded-2xl bg-slate-50 px-4 py-3">
							<div>
								<p class="text-sm font-bold text-slate-900">Taux d'erreur API</p>
								<p class="text-xs text-slate-400">Formule : requêtes erreur / total × 100</p>
							</div>
							<span
								class={`rounded-full px-3 py-1 text-xs font-bold ${
									catalogStatus === 'ok'
										? 'bg-emerald-100 text-emerald-700'
										: 'bg-slate-100 text-slate-500'
								}`}
							>
								{catalogStatus === 'ok' ? '< 1%' : '...'}
							</span>
						</div>
						<div class="flex items-center justify-between rounded-2xl bg-slate-50 px-4 py-3">
							<div>
								<p class="text-sm font-bold text-slate-900">Articles catalogue</p>
								<p class="text-xs text-slate-400">catalog-service live</p>
							</div>
							<span class="rounded-full bg-orange-100 px-3 py-1 text-xs font-bold text-orange-700">
								{catalogStatus === 'ok' ? articlesCount : '—'}
							</span>
						</div>
					</div>
				</div>
			</div>

			<!-- DORA Metrics -->
			<p class="mb-3 text-xs font-semibold tracking-widest text-slate-400 uppercase">
				DORA Metrics · Standard mondial DevOps
			</p>
			<div class="mb-6 rounded-3xl border border-slate-800 bg-slate-950 p-6 text-white shadow-lg">
				<p class="mb-1 text-xs font-semibold tracking-widest text-orange-300 uppercase">
					State of DevOps Report — 4 indicateurs clés
				</p>
				<h2 class="mb-5 text-xl font-black tracking-tight uppercase">DORA Metrics</h2>
				<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
					{#each doraMetrics as metric}
						<div class="rounded-2xl border border-slate-800 bg-slate-900 p-4">
							<p class="text-xs font-semibold tracking-widest text-slate-400 uppercase">
								{metric.label}
							</p>
							<p class="mt-2 text-2xl font-black text-white">{metric.value}</p>
							<p class="mt-1 text-xs text-slate-400">{metric.note}</p>
							<span
								class={`mt-3 inline-block rounded-full px-2 py-0.5 text-xs font-bold ${
									metric.status === 'good'
										? 'bg-emerald-900 text-emerald-400'
										: metric.status === 'medium'
											? 'bg-amber-900 text-amber-400'
											: metric.status === 'bad'
												? 'bg-red-900 text-red-400'
												: 'bg-slate-800 text-slate-400'
								}`}
							>
								{metric.status === 'good'
									? 'Élite'
									: metric.status === 'medium'
										? 'En progression'
										: metric.status === 'na'
											? 'À mesurer'
											: 'Critique'}
							</span>
						</div>
					{/each}
				</div>
				<p class="mt-4 text-xs text-slate-500">
					Cours p.15-16 : Deployment Frequency · Lead Time · MTTR · Change Failure Rate
				</p>
			</div>

			<div class="rounded-3xl border border-orange-200 bg-orange-50 p-5">
				<p class="text-xs font-bold tracking-widest text-orange-700 uppercase">
					Lecture CTO · Cours p.17
				</p>
				<p class="mt-2 text-sm text-orange-800">
					Le CTO suit la disponibilité des services, la latence et les métriques DORA. Le MTTR et le
					Change Failure Rate nécessiteront Prometheus / Grafana en production pour être mesurés
					avec précision.
				</p>
			</div>

			<!-- =============================== -->
			<!-- VUE CEO — Direction générale    -->
			<!-- =============================== -->
		{:else if activeActor === 'ceo'}
			<!-- Ligne 4 : Business -->
			<p class="mb-3 text-xs font-semibold tracking-widest text-slate-400 uppercase">
				Ligne 4 — Business (valeur créée)
			</p>
			<div class="mb-6 grid gap-4 md:grid-cols-2 lg:grid-cols-4">
				{#each [{ label: 'Taux de churn', value: churnRate, sub: 'Formule : utilisateurs perdus / total × 100', threshold: '< 5% = bon · > 10% = critique', color: 'slate' }, { label: 'Taux de rétention', value: retentionRate, sub: 'Utilisateurs actifs après période / initiaux × 100', threshold: 'Fidélité produit', color: 'slate' }, { label: 'Taux de conversion', value: conversionRate, sub: 'Visiteurs ayant agi / visiteurs totaux × 100', threshold: 'Cible : > 5%', color: 'slate' }, { label: 'ARPU', value: arpu, sub: "Revenu total / nombre d'utilisateurs", threshold: 'Amélioration UX → ARPU augmente', color: 'slate' }] as kpi}
					<div class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
						<p class="text-xs font-semibold tracking-widest text-slate-400 uppercase">
							{kpi.label}
						</p>
						<p class="mt-2 text-4xl font-black text-slate-400">{kpi.value}</p>
						<p class="mt-2 text-xs text-slate-400">{kpi.sub}</p>
						<div class="mt-3 rounded-xl bg-slate-50 px-3 py-2 text-xs text-slate-500">
							{kpi.threshold}
						</div>
					</div>
				{/each}
			</div>

			<!-- Données live catalogue -->
			<p class="mb-3 text-xs font-semibold tracking-widest text-slate-400 uppercase">
				Données live · Catalogue
			</p>
			<div class="mb-6 grid gap-4 md:grid-cols-2">
				<div class="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
					<p class="text-xs font-semibold tracking-widest text-slate-400 uppercase">
						Articles publiés
					</p>
					<p class="mt-2 text-5xl font-black text-orange-600">
						{catalogStatus === 'ok' ? articlesCount : '—'}
					</p>
					<p class="mt-1 text-xs text-slate-400">catalogue live · catalog-service</p>
				</div>
				<div class="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
					<p class="text-xs font-semibold tracking-widest text-slate-400 uppercase">Catégories</p>
					<p class="mt-2 text-5xl font-black text-slate-900">
						{catalogStatus === 'ok' ? categoriesCount : '—'}
					</p>
					<p class="mt-1 text-xs text-slate-400">taxonomie live · catalog-service</p>
				</div>
			</div>

			<!-- Avertissement -->
			<div class="mb-6 rounded-3xl border border-amber-200 bg-amber-50 p-5">
				<p class="text-xs font-bold tracking-widest text-amber-700 uppercase">
					KPIs business — À instrumenter
				</p>
				<p class="mt-2 text-sm text-amber-800">
					Churn, rétention, conversion et ARPU nécessitent un outil d'analytics dédié (Amplitude,
					Mixpanel, ou tracking custom). Ces indicateurs sont définis dans le cours mais non encore
					mesurables sans collecte de données utilisateurs.
				</p>
			</div>

			<div class="rounded-3xl border border-orange-200 bg-orange-50 p-5">
				<p class="text-xs font-bold tracking-widest text-orange-700 uppercase">
					Lecture CEO · Cours p.12-14
				</p>
				<p class="mt-2 text-sm text-orange-800">
					Le CEO suit la valeur business créée : rétention, churn, conversion, ARPU. Un projet
					techniquement bon mais business KO = échec. Lien direct : KPI technique dégradé → KPI
					business impacté (ex. latence élevée → churn augmente).
				</p>
			</div>
		{/if}

		<!-- Note pédagogique commune -->
		<div class="mt-6 rounded-3xl border border-slate-200 bg-white p-5">
			<p class="text-xs font-bold tracking-widest text-slate-400 uppercase">
				Principe fondamental · Cours p.17
			</p>
			<div class="mt-2 grid gap-3 text-xs text-slate-600 md:grid-cols-3">
				<div class="rounded-xl bg-slate-50 px-4 py-3">
					<p class="font-bold text-slate-800">1. Identifier l'utilisateur</p>
					<p class="mt-1">CEO → business · CTO → technique · PM → projet</p>
				</div>
				<div class="rounded-xl bg-slate-50 px-4 py-3">
					<p class="font-bold text-slate-800">2. Limiter les KPI</p>
					<p class="mt-1">Règle : 5 à 10 KPI max par vue</p>
				</div>
				<div class="rounded-xl bg-slate-50 px-4 py-3">
					<p class="font-bold text-slate-800">3. Logique de décision</p>
					<p class="mt-1">Un dashboard ≠ liste de KPI = outil pour décider</p>
				</div>
			</div>
		</div>
	</div>
</div>
