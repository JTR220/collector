<script lang="ts">
	import { env } from '$env/dynamic/public';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/stores/auth';
	import { fetchAlerts, resolveAlert, type FraudAlertAPI } from '$lib/api/priceTracker';
	import { fromEventUuid } from '$lib/utils/eventId';
	import { fetchUsers, suspendUser, unsuspendUser, type AdminUser } from '$lib/api/auth';

	const authApiUrl = env.PUBLIC_AUTH_API_BASE_URL ?? 'http://localhost:8080';
	const catalogApiUrl = env.PUBLIC_CATALOG_API_BASE_URL ?? 'http://localhost:8081';

	type AdminStats = {
		gmv: number;
		totalOrders: number;
		avgOrderValue: number;
		ordersByStatus: { paid: number; shipped: number; delivered: number; cancelled: number };
		totalArticles: number;
		activeListings: number;
		soldArticles: number;
		sellThrough: number;
		avgListing: number;
		categories: number;
		activeSellers: number;
		byCategory: { name: string; count: number }[];
		recentOrders: {
			id: number;
			article: string;
			price: number;
			status: string;
			buyerId: number;
			createdAt: string;
		}[];
	};

	const ORDER_STATUS_FR: Record<string, string> = {
		paid: 'Payée',
		shipped: 'Expédiée',
		delivered: 'Livrée',
		cancelled: 'Annulée'
	};
	const eur = (n: number) => `${n.toLocaleString('fr-FR')} €`;

	type ServiceStatus = 'loading' | 'ok' | 'down';

	let authStatus: ServiceStatus = 'loading';
	let catalogStatus: ServiceStatus = 'loading';
	let catalogLatencyMs: number | null = null;
	let stats: AdminStats | null = null;
	let statsErreur = '';
	let lastRefresh = new Date();

	function authHeaders(): Record<string, string> {
		return $auth.token ? { Authorization: `Bearer ${$auth.token}` } : {};
	}

	async function checkHealth(url: string): Promise<ServiceStatus> {
		try {
			const res = await fetch(`${url}/health`, { signal: AbortSignal.timeout(3000) });
			return res.ok ? 'ok' : 'down';
		} catch {
			return 'down';
		}
	}

	async function chargerStats() {
		try {
			const response = await fetch(`${catalogApiUrl}/admin/stats`, { headers: authHeaders() });
			if (!response.ok) throw new Error();
			stats = (await response.json()) as AdminStats;
			statsErreur = '';
		} catch {
			stats = null;
			statsErreur = 'Statistiques indisponibles.';
		}
	}

	async function refreshAll() {
		lastRefresh = new Date();
		authStatus = 'loading';
		catalogStatus = 'loading';

		const [a, c] = await Promise.all([
			checkHealth(authApiUrl),
			(async () => {
				const t0 = Date.now();
				const s = await checkHealth(catalogApiUrl);
				catalogLatencyMs = Date.now() - t0;
				return s;
			})()
		]);

		authStatus = a;
		catalogStatus = c;

		if (catalogStatus === 'ok') {
			await chargerStats();
		} else {
			stats = null;
			statsErreur = 'catalog-service indisponible.';
		}
	}

	// --- Alertes fraude (price-tracker) ---
	let fraudAlerts: FraudAlertAPI[] = [];
	let fraudTrackerDown = false;

	async function fetchFraudAlerts() {
		if (!$auth.token) return;
		try {
			fraudAlerts = await fetchAlerts($auth.token, true);
			fraudTrackerDown = false;
		} catch {
			fraudAlerts = [];
			fraudTrackerDown = true;
		}
	}

	async function onResolveAlert(id: string) {
		if (!$auth.token) return;
		try {
			await resolveAlert($auth.token, id);
			fraudAlerts = fraudAlerts.filter((a) => a.id !== id);
		} catch {
			/* ignore */
		}
	}

	const reasonLabels: Record<FraudAlertAPI['reason'], string> = {
		SUSPICIOUS_SPIKE: 'Hausse suspecte',
		FLOOD_PRICING: 'Modifications en rafale',
		DUMPING: 'Prix anormalement bas'
	};

	// --- Modération des comptes (suspension) ---
	let users: AdminUser[] = [];
	let usersDown = false;
	let userBusyId: number | null = null;
	let moderationMsg = '';

	async function fetchUsersList() {
		if (!$auth.token) return;
		try {
			users = await fetchUsers($auth.token);
			usersDown = false;
		} catch {
			users = [];
			usersDown = true;
		}
	}

	async function onToggleSuspend(u: AdminUser) {
		if (!$auth.token) return;
		userBusyId = u.ID;
		moderationMsg = '';
		try {
			const { suspended } = u.suspended
				? await unsuspendUser($auth.token, u.ID)
				: await suspendUser($auth.token, u.ID);
			users = users.map((x) => (x.ID === u.ID ? { ...x, suspended } : x));
		} catch (e) {
			moderationMsg = e instanceof Error ? e.message : 'Action impossible.';
		} finally {
			userBusyId = null;
		}
	}

	onMount(() => {
		// Page reservee aux administrateurs.
		if ($auth.user?.role !== 'admin') {
			goto('/login');
			return;
		}
		refreshAll();
		fetchFraudAlerts();
		fetchUsersList();
	});
</script>

<svelte:head>
	<title>Collector.shop | Tableau de bord</title>
</svelte:head>

<div class="dash">
	<header class="head">
		<div>
			<div class="eyebrow">Pilotage</div>
			<h1 class="title">Tableau de bord</h1>
			<p class="subtitle">
				Dernière mise à jour : {lastRefresh.toLocaleTimeString('fr-FR')}
			</p>
		</div>
		<button class="btn-refresh" type="button" onclick={refreshAll}>Rafraîchir</button>
	</header>

	<!-- Infrastructure -->
	<section class="panel">
		<div class="eyebrow">Infrastructure</div>
		<h2 class="panel-title">Disponibilité des services</h2>
		<div class="infra-grid">
			{#each [{ label: 'auth-service', status: authStatus }, { label: 'catalog-service', status: catalogStatus }] as svc}
				<div class="infra-row">
					<div>
						<span class="infra-name">{svc.label}</span>
						<span class="infra-sub">GET /health</span>
					</div>
					<span class="infra-badge infra-{svc.status}">
						{svc.status === 'loading' ? '…' : svc.status === 'ok' ? 'UP' : 'DOWN'}
					</span>
				</div>
			{/each}
			{#if catalogLatencyMs !== null}
				<div class="infra-row">
					<div>
						<span class="infra-name">Latence catalog-service</span>
						<span class="infra-sub">temps de réponse /health</span>
					</div>
					<span
						class="infra-badge"
						class:infra-ok={catalogLatencyMs < 200}
						class:infra-warn={catalogLatencyMs >= 200 && catalogLatencyMs < 500}
						class:infra-down={catalogLatencyMs >= 500}
					>
						{catalogLatencyMs} ms
					</span>
				</div>
			{/if}
		</div>
	</section>

	{#if statsErreur}
		<div class="msg msg-error">{statsErreur}</div>
	{/if}

	{#if stats}
		<section class="kpis">
			<div class="kpi kpi-hero">
				<span class="kpi-label">Volume d'affaires (GMV)</span>
				<span class="kpi-val">{eur(stats.gmv)}</span>
				<span class="kpi-sub">Panier moyen {eur(Math.round(stats.avgOrderValue))}</span>
			</div>
			<div class="kpi">
				<span class="kpi-label">Commandes</span>
				<span class="kpi-val">{stats.totalOrders}</span>
				<span class="kpi-sub">{stats.ordersByStatus.paid} à expédier</span>
			</div>
			<div class="kpi">
				<span class="kpi-label">Taux d'écoulement</span>
				<span class="kpi-val">{stats.sellThrough.toFixed(1)}%</span>
				<span class="kpi-sub">{stats.soldArticles}/{stats.totalArticles} vendues</span>
			</div>
			<div class="kpi">
				<span class="kpi-label">Annonces actives</span>
				<span class="kpi-val">{stats.activeListings}</span>
				<span class="kpi-sub">prix moyen {eur(Math.round(stats.avgListing))}</span>
			</div>
			<div class="kpi">
				<span class="kpi-label">Vendeurs actifs</span>
				<span class="kpi-val">{stats.activeSellers}</span>
				<span class="kpi-sub">{stats.categories} catégories</span>
			</div>
		</section>

		<div class="dash-grid">
			<section class="panel">
				<div class="eyebrow">Suivi</div>
				<h2 class="panel-title">Commandes par statut</h2>
				<div class="funnel">
					{#each ['paid', 'shipped', 'delivered', 'cancelled'] as st}
						{@const n = stats.ordersByStatus[st as keyof typeof stats.ordersByStatus]}
						<div class="funnel-row">
							<span class="funnel-name" class:funnel-cancel={st === 'cancelled'}>
								{ORDER_STATUS_FR[st]}
							</span>
							<span class="funnel-track">
								<span
									class="funnel-fill"
									class:funnel-fill-cancel={st === 'cancelled'}
									style={`width:${stats.totalOrders ? (n / stats.totalOrders) * 100 : 0}%`}
								></span>
							</span>
							<span class="funnel-count">{n}</span>
						</div>
					{/each}
				</div>

				<div class="eyebrow" style="margin-top:20px">Catalogue</div>
				<h2 class="panel-title">Répartition par catégorie</h2>
				<div class="funnel">
					{#each stats.byCategory as row}
						{#if row.name}
							<div class="funnel-row">
								<span class="funnel-name">{row.name}</span>
								<span class="funnel-track">
									<span
										class="funnel-fill"
										style={`width:${stats.totalArticles ? (row.count / stats.totalArticles) * 100 : 0}%`}
									></span>
								</span>
								<span class="funnel-count">{row.count}</span>
							</div>
						{/if}
					{/each}
				</div>
			</section>

			<section class="panel">
				<div class="eyebrow">Activité</div>
				<h2 class="panel-title">Dernières commandes</h2>
				{#if stats.recentOrders.length === 0}
					<div class="empty">Aucune commande pour l'instant.</div>
				{:else}
					<div class="orders">
						{#each stats.recentOrders as o}
							<div class="order-row">
								<span class="order-date">{o.createdAt}</span>
								<span class="order-name">{o.article || `Lot #${o.id}`}</span>
								<span class="order-status order-{o.status}">
									{ORDER_STATUS_FR[o.status] ?? o.status}
								</span>
								<span class="order-price">{eur(o.price)}</span>
							</div>
						{/each}
					</div>
				{/if}
			</section>
		</div>
	{/if}

	<!-- Modération des comptes -->
	{#if moderationMsg}
		<div class="msg msg-error">{moderationMsg}</div>
	{/if}
	<section class="panel">
		<div class="eyebrow">Modération</div>
		<h2 class="panel-title">Comptes utilisateurs</h2>
		{#if usersDown}
			<div class="empty">auth-service indisponible.</div>
		{:else if users.length === 0}
			<div class="empty">Aucun utilisateur.</div>
		{:else}
			<div class="mod-list mod-list-scroll">
				{#each users as u (u.ID)}
					<div class="mod-row">
						<div class="mod-info">
							<span class="mod-name">{u.name}</span>
							<span class="mod-sub">{u.email} · {u.role}</span>
						</div>
						{#if u.suspended}
							<span class="mod-badge mod-badge-suspended">Suspendu</span>
						{/if}
						<button
							class="mod-action"
							class:mod-action-danger={!u.suspended}
							disabled={userBusyId === u.ID}
							onclick={() => onToggleSuspend(u)}
						>
							{u.suspended ? 'Réactiver' : 'Suspendre'}
						</button>
					</div>
				{/each}
			</div>
		{/if}
	</section>

	<!-- Alertes fraude -->
	<section class="panel">
		<div class="eyebrow">Sécurité</div>
		<h2 class="panel-title">Alertes fraude</h2>
		{#if fraudTrackerDown}
			<div class="empty">price-tracker-service indisponible.</div>
		{:else if fraudAlerts.length === 0}
			<div class="empty">Aucune alerte non résolue.</div>
		{:else}
			<div class="alerts">
				{#each fraudAlerts as alert (alert.id)}
					<div class="alert-row">
						<div>
							<p class="alert-title">
								{reasonLabels[alert.reason] ?? alert.reason}
								<span class="alert-item">article #{fromEventUuid(alert.item_id)}</span>
							</p>
							<p class="alert-detail">{alert.detail}</p>
							<p class="alert-meta">
								{alert.old_price.toFixed(2)} € → {alert.new_price.toFixed(2)} € · {new Date(
									alert.created_at
								).toLocaleString('fr-FR')}
							</p>
						</div>
						<button type="button" class="alert-resolve" onclick={() => onResolveAlert(alert.id)}>
							Résoudre
						</button>
					</div>
				{/each}
			</div>
		{/if}
	</section>
</div>

<style>
	.dash {
		max-width: 1200px;
		margin: 0 auto;
		padding-bottom: 60px;
	}

	.eyebrow {
		font-size: 11px;
		letter-spacing: 0.2em;
		text-transform: uppercase;
		font-weight: 600;
		color: #86b3a4;
		margin-bottom: 10px;
	}

	.head {
		display: flex;
		align-items: flex-end;
		justify-content: space-between;
		gap: 24px;
		flex-wrap: wrap;
		border-bottom: 1px solid rgba(236, 229, 218, 0.1);
		padding-bottom: 24px;
		margin-bottom: 24px;
	}
	.title {
		font-family: 'Newsreader', Georgia, serif;
		font-weight: 500;
		font-size: 40px;
		line-height: 1;
		color: #ece5da;
		margin: 0 0 8px;
	}
	.subtitle {
		font-size: 13px;
		color: #a39a8c;
		margin: 0;
	}
	.btn-refresh {
		padding: 10px 18px;
		border-radius: 7px;
		border: 1px solid rgba(236, 229, 218, 0.14);
		background: rgba(255, 255, 255, 0.02);
		color: #ece5da;
		font-size: 12px;
		font-weight: 700;
		letter-spacing: 0.04em;
		text-transform: uppercase;
		cursor: pointer;
		transition: background 150ms;
	}
	.btn-refresh:hover {
		background: rgba(255, 255, 255, 0.05);
	}

	.msg {
		padding: 10px 14px;
		border-radius: 7px;
		border: 1px solid;
		font-size: 13px;
		margin-bottom: 16px;
	}
	.msg-error {
		border-color: rgba(215, 156, 134, 0.3);
		background: rgba(215, 156, 134, 0.06);
		color: #d79c86;
	}

	.panel {
		background: #221f1b;
		border: 1px solid rgba(236, 229, 218, 0.1);
		border-radius: 9px;
		padding: 20px;
		margin-bottom: 22px;
	}
	.panel-title {
		font-family: 'Newsreader', Georgia, serif;
		font-weight: 500;
		font-size: 22px;
		color: #ece5da;
		margin: 0 0 16px;
	}
	.empty {
		border: 1px dashed rgba(236, 229, 218, 0.14);
		border-radius: 8px;
		padding: 18px;
		font-size: 13px;
		color: #a39a8c;
	}

	/* Infrastructure */
	.infra-grid {
		display: flex;
		flex-direction: column;
		gap: 10px;
		margin-top: 6px;
	}
	.infra-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
		padding: 12px 14px;
		border-radius: 8px;
		background: rgba(255, 255, 255, 0.03);
	}
	.infra-name {
		display: block;
		font-size: 13.5px;
		font-weight: 600;
		color: #ece5da;
	}
	.infra-sub {
		display: block;
		font-size: 11.5px;
		color: #766d60;
		margin-top: 2px;
	}
	.infra-badge {
		flex-shrink: 0;
		font-size: 11px;
		font-weight: 700;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		padding: 4px 10px;
		border-radius: 20px;
		background: rgba(255, 255, 255, 0.06);
		color: #a39a8c;
	}
	.infra-ok {
		background: rgba(134, 179, 164, 0.14);
		color: #86b3a4;
	}
	.infra-warn {
		background: rgba(224, 178, 96, 0.14);
		color: #e0b260;
	}
	.infra-down {
		background: rgba(215, 156, 134, 0.14);
		color: #d79c86;
	}

	/* KPI */
	.kpis {
		display: grid;
		grid-template-columns: 1.4fr 1fr 1fr 1fr 1fr;
		gap: 12px;
		margin-bottom: 16px;
	}
	.kpi {
		display: flex;
		flex-direction: column;
		gap: 5px;
		padding: 16px 18px;
		border: 1px solid rgba(236, 229, 218, 0.1);
		border-radius: 10px;
		background: rgba(255, 255, 255, 0.02);
	}
	.kpi-hero {
		background: rgba(134, 179, 164, 0.06);
		border-color: rgba(134, 179, 164, 0.24);
	}
	.kpi-label {
		font-size: 10.5px;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		color: #766d60;
	}
	.kpi-val {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 25px;
		color: #86b3a4;
	}
	.kpi-hero .kpi-val {
		font-size: 30px;
		color: #ece5da;
	}
	.kpi-sub {
		font-size: 11.5px;
		color: #a39a8c;
	}
	@media (max-width: 900px) {
		.kpis {
			grid-template-columns: repeat(2, 1fr);
		}
	}

	.dash-grid {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 14px;
		margin-bottom: 22px;
		align-items: start;
	}
	.dash-grid .panel {
		margin-bottom: 0;
	}
	@media (max-width: 900px) {
		.dash-grid {
			grid-template-columns: 1fr;
		}
	}

	/* Funnel */
	.funnel {
		display: flex;
		flex-direction: column;
		gap: 9px;
		margin-top: 6px;
	}
	.funnel-row {
		display: flex;
		align-items: center;
		gap: 12px;
	}
	.funnel-name {
		width: 96px;
		font-size: 12.5px;
		color: #ece5da;
		flex-shrink: 0;
	}
	.funnel-cancel {
		color: #d79c86;
	}
	.funnel-track {
		flex: 1;
		height: 8px;
		border-radius: 4px;
		background: rgba(255, 255, 255, 0.05);
		overflow: hidden;
	}
	.funnel-fill {
		display: block;
		height: 100%;
		background: #86b3a4;
		border-radius: 4px;
		min-width: 2px;
	}
	.funnel-fill-cancel {
		background: #d79c86;
	}
	.funnel-count {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 12px;
		color: #a39a8c;
		width: 28px;
		text-align: right;
		flex-shrink: 0;
	}

	/* Commandes */
	.orders {
		display: flex;
		flex-direction: column;
		margin-top: 6px;
	}
	.order-row {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 10px 0;
		border-bottom: 1px solid rgba(236, 229, 218, 0.08);
	}
	.order-date {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 11px;
		color: #766d60;
		flex-shrink: 0;
		width: 82px;
	}
	.order-name {
		flex: 1;
		font-size: 13px;
		color: #ece5da;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.order-status {
		font-size: 10.5px;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		padding: 3px 8px;
		border-radius: 4px;
		flex-shrink: 0;
		background: rgba(134, 179, 164, 0.1);
		color: #86b3a4;
	}
	.order-cancelled {
		background: rgba(215, 156, 134, 0.1);
		color: #d79c86;
	}
	.order-price {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 12.5px;
		color: #a39a8c;
		flex-shrink: 0;
		width: 84px;
		text-align: right;
	}

	/* Modération */
	.mod-list {
		display: flex;
		flex-direction: column;
		margin-top: 6px;
	}
	.mod-list-scroll {
		max-height: 320px;
		overflow-y: auto;
	}
	.mod-row {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 10px 0;
		border-bottom: 1px solid rgba(236, 229, 218, 0.08);
	}
	.mod-info {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 2px;
	}
	.mod-name {
		font-size: 13px;
		color: #ece5da;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.mod-sub {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 11px;
		color: #766d60;
	}
	.mod-badge {
		flex-shrink: 0;
		font-size: 10px;
		font-weight: 700;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		padding: 3px 8px;
		border-radius: 20px;
	}
	.mod-badge-suspended {
		background: rgba(215, 156, 134, 0.14);
		color: #d79c86;
	}
	.mod-action {
		flex-shrink: 0;
		padding: 6px 12px;
		border-radius: 6px;
		border: 1px solid rgba(236, 229, 218, 0.14);
		background: rgba(255, 255, 255, 0.02);
		color: #a39a8c;
		font-size: 11px;
		font-weight: 700;
		letter-spacing: 0.03em;
		cursor: pointer;
		transition: background 150ms;
	}
	.mod-action:hover:not(:disabled) {
		background: rgba(255, 255, 255, 0.06);
	}
	.mod-action:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}
	.mod-action-danger {
		border-color: rgba(215, 156, 134, 0.4);
		color: #d79c86;
	}

	/* Alertes fraude */
	.alerts {
		display: flex;
		flex-direction: column;
		gap: 10px;
		margin-top: 6px;
	}
	.alert-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 16px;
		padding: 14px 16px;
		border: 1px solid rgba(215, 156, 134, 0.24);
		background: rgba(215, 156, 134, 0.05);
		border-radius: 9px;
		flex-wrap: wrap;
	}
	.alert-title {
		font-size: 13px;
		font-weight: 700;
		color: #d79c86;
		text-transform: uppercase;
		letter-spacing: 0.02em;
		margin: 0;
	}
	.alert-item {
		margin-left: 8px;
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 11px;
		font-weight: 400;
		color: #a39a8c;
		text-transform: none;
	}
	.alert-detail {
		margin: 4px 0 0;
		font-size: 13px;
		color: #cbc3b6;
	}
	.alert-meta {
		margin: 4px 0 0;
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 11px;
		color: #766d60;
	}
	.alert-resolve {
		flex-shrink: 0;
		padding: 8px 16px;
		border-radius: 7px;
		border: 1px solid rgba(215, 156, 134, 0.4);
		background: rgba(255, 255, 255, 0.02);
		color: #d79c86;
		font-size: 11.5px;
		font-weight: 700;
		letter-spacing: 0.04em;
		text-transform: uppercase;
		cursor: pointer;
		transition: background 150ms;
	}
	.alert-resolve:hover {
		background: rgba(215, 156, 134, 0.12);
	}
</style>
