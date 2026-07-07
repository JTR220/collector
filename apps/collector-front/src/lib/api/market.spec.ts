import { describe, it, expect, beforeEach, vi } from 'vitest';

import { buyArticle, fetchMyOrders } from './market';

describe('market API', () => {
	const fetchMock = vi.fn();

	beforeEach(() => {
		fetchMock.mockReset();
		vi.stubGlobal('fetch', fetchMock);
	});

	it('buyArticle poste sur /article/:id/buy avec le token Bearer', async () => {
		fetchMock.mockResolvedValue({
			ok: true,
			json: async () => ({ order: { ID: 1, articleId: 5 } })
		});

		const result = await buyArticle('tok-abc', 5);

		const [url, init] = fetchMock.mock.calls[0];
		expect(url).toContain('/article/5/buy');
		expect(init.method).toBe('POST');
		expect((init.headers as Record<string, string>).Authorization).toBe('Bearer tok-abc');
		expect(result.order.ID).toBe(1);
	});

	it('buyArticle rejette avec le message backend (ex: deja vendu)', async () => {
		fetchMock.mockResolvedValue({
			ok: false,
			status: 409,
			json: async () => ({ error: 'Cette piece est deja vendue' })
		});

		await expect(buyArticle('tok', 5)).rejects.toThrow('Cette piece est deja vendue');
	});

	it('fetchMyOrders appelle /me/orders avec le token', async () => {
		fetchMock.mockResolvedValue({ ok: true, json: async () => [] });

		await fetchMyOrders('tok-abc');

		const [url, init] = fetchMock.mock.calls[0];
		expect(url).toContain('/me/orders');
		expect((init.headers as Record<string, string>).Authorization).toBe('Bearer tok-abc');
	});
});
