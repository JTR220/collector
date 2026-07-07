import { describe, it, expect, beforeEach, vi } from 'vitest';

import { fetchPriceHistory, fetchAlerts, resolveAlert } from './priceTracker';

describe('priceTracker API', () => {
	const fetchMock = vi.fn();

	beforeEach(() => {
		fetchMock.mockReset();
		vi.stubGlobal('fetch', fetchMock);
	});

	it('fetchPriceHistory ne requiert pas de token (route publique)', async () => {
		fetchMock.mockResolvedValue({
			ok: true,
			json: async () => ({ count: 1, history: [{ new_price: 42 }] })
		});

		const history = await fetchPriceHistory(1);

		const [url, init] = fetchMock.mock.calls[0];
		expect(url).toContain('/api/v1/items/');
		expect(url).toContain('/price-history');
		expect((init?.headers as Record<string, string> | undefined)?.Authorization).toBeUndefined();
		expect(history).toEqual([{ new_price: 42 }]);
	});

	it('fetchPriceHistory renvoie [] si history est null', async () => {
		fetchMock.mockResolvedValue({ ok: true, json: async () => ({ count: 0, history: null }) });
		expect(await fetchPriceHistory(1)).toEqual([]);
	});

	it('fetchAlerts envoie le token Bearer et le filtre unresolved', async () => {
		fetchMock.mockResolvedValue({ ok: true, json: async () => ({ count: 0, alerts: [] }) });

		await fetchAlerts('tok-abc', true);

		const [url, init] = fetchMock.mock.calls[0];
		expect(url).toContain('/api/v1/alerts?unresolved=true');
		expect((init.headers as Record<string, string>).Authorization).toBe('Bearer tok-abc');
	});

	it('resolveAlert envoie PUT avec le token Bearer', async () => {
		fetchMock.mockResolvedValue({ ok: true, json: async () => ({}) });

		await resolveAlert('tok-abc', 'alert-1');

		const [url, init] = fetchMock.mock.calls[0];
		expect(url).toContain('/api/v1/alerts/alert-1/resolve');
		expect(init.method).toBe('PUT');
		expect((init.headers as Record<string, string>).Authorization).toBe('Bearer tok-abc');
	});

	it('rejette sur reponse non-ok', async () => {
		fetchMock.mockResolvedValue({ ok: false, status: 403 });
		await expect(fetchAlerts('tok', false)).rejects.toThrow(/403/);
	});
});
