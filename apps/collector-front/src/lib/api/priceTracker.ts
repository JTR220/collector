import { env } from '$env/dynamic/public';
import { toEventUuid } from '$lib/utils/eventId';

const BASE_URL = env.PUBLIC_PRICE_TRACKER_API_BASE_URL ?? 'http://localhost:8082';

export type PriceHistoryEntry = {
	id: string;
	item_id: string;
	seller_id: string;
	old_price: number;
	new_price: number;
	created_at: string;
};

export type FraudAlertAPI = {
	id: string;
	item_id: string;
	seller_id: string;
	reason: 'SUSPICIOUS_SPIKE' | 'FLOOD_PRICING' | 'DUMPING';
	detail: string;
	old_price: number;
	new_price: number;
	resolved: boolean;
	created_at: string;
};

async function request<T>(path: string, init?: RequestInit, token?: string): Promise<T> {
	const res = await fetch(`${BASE_URL}${path}`, {
		...init,
		headers: {
			...(token ? { Authorization: `Bearer ${token}` } : {}),
			...init?.headers
		}
	});
	if (!res.ok) throw new Error(`price-tracker ${path} error: ${res.status}`);
	return res.json();
}

// Historique de prix : route publique (fiche lot visible sans connexion).
export async function fetchPriceHistory(articleId: number): Promise<PriceHistoryEntry[]> {
	const data = await request<{ count: number; history: PriceHistoryEntry[] | null }>(
		`/api/v1/items/${toEventUuid(articleId)}/price-history`
	);
	return data.history ?? [];
}

// Alertes de fraude : reservees au dashboard admin, token requis.
export async function fetchAlerts(token: string, unresolved = false): Promise<FraudAlertAPI[]> {
	const qs = unresolved ? '?unresolved=true' : '';
	const data = await request<{ count: number; alerts: FraudAlertAPI[] | null }>(
		`/api/v1/alerts${qs}`,
		undefined,
		token
	);
	return data.alerts ?? [];
}

export async function resolveAlert(token: string, id: string): Promise<void> {
	await request(`/api/v1/alerts/${id}/resolve`, { method: 'PUT' }, token);
}
