import { describe, it, expect, beforeEach, vi } from 'vitest';

import { fetchPriceHistory, fetchAlerts, resolveAlert } from './priceTracker';

describe('priceTracker API', () => {
	const fetchMock = vi.fn();

	beforeEach(() => {
		fetchMock.mockReset();
		vi.stubGlobal('fetch', fetchMock);
	});

	it('fetchPriceHistory route publique (aucune coordonnee particuliere requise)', async () => {
		fetchMock.mockResolvedValue({
			ok: true,
			json: async () => ({ count: 1, history: [{ new_price: 42 }] })
		});

		const history = await fetchPriceHistory(1);

		const [url] = fetchMock.mock.calls[0];
		expect(url).toContain('/api/v1/items/');
		expect(url).toContain('/price-history');
		expect(history).toEqual([{ new_price: 42 }]);
	});

	it('fetchPriceHistory renvoie [] si history est null', async () => {
		fetchMock.mockResolvedValue({ ok: true, json: async () => ({ count: 0, history: null }) });
		expect(await fetchPriceHistory(1)).toEqual([]);
	});

	it('fetchAlerts envoie credentials:include et le filtre unresolved', async () => {
		fetchMock.mockResolvedValue({ ok: true, json: async () => ({ count: 0, alerts: [] }) });

		await fetchAlerts(true);

		const [url, init] = fetchMock.mock.calls[0];
		expect(url).toContain('/api/v1/alerts?unresolved=true');
		expect(init.credentials).toBe('include');
	});

	it('resolveAlert envoie PUT avec credentials:include', async () => {
		fetchMock.mockResolvedValue({ ok: true, json: async () => ({}) });

		await resolveAlert('alert-1');

		const [url, init] = fetchMock.mock.calls[0];
		expect(url).toContain('/api/v1/alerts/alert-1/resolve');
		expect(init.method).toBe('PUT');
		expect(init.credentials).toBe('include');
	});

	it('rejette sur reponse non-ok', async () => {
		fetchMock.mockResolvedValue({ ok: false, status: 403 });
		await expect(fetchAlerts(false)).rejects.toThrow(/403/);
	});
});
