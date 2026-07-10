import { env } from '$env/dynamic/public';
import { toEventUuid } from '$lib/utils/eventId';
import { apiRequest } from './http';

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

const request = <T>(path: string, init?: RequestInit) =>
	apiRequest<T>(BASE_URL, path, { init, errorPrefix: `price-tracker ${path}` });

// Historique de prix : route publique (fiche lot visible sans connexion).
export async function fetchPriceHistory(articleId: number): Promise<PriceHistoryEntry[]> {
	const data = await request<{ count: number; history: PriceHistoryEntry[] | null }>(
		`/api/v1/items/${toEventUuid(articleId)}/price-history`
	);
	return data.history ?? [];
}

// Alertes de fraude : reservees au dashboard admin (cookie de session requis).
export async function fetchAlerts(unresolved = false): Promise<FraudAlertAPI[]> {
	const qs = unresolved ? '?unresolved=true' : '';
	const data = await request<{ count: number; alerts: FraudAlertAPI[] | null }>(
		`/api/v1/alerts${qs}`
	);
	return data.alerts ?? [];
}

export async function resolveAlert(id: string): Promise<void> {
	await request(`/api/v1/alerts/${id}/resolve`, { method: 'PUT' });
}
