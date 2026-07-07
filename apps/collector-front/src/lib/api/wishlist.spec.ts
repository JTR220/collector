import { describe, it, expect, beforeEach, vi } from 'vitest';

import { fetchMyWishlist, addToWishlist, removeFromWishlist } from './wishlist';

describe('wishlist API', () => {
	const fetchMock = vi.fn();

	beforeEach(() => {
		fetchMock.mockReset();
		vi.stubGlobal('fetch', fetchMock);
	});

	it('fetchMyWishlist envoie le token Bearer', async () => {
		fetchMock.mockResolvedValue({ ok: true, json: async () => [] });

		await fetchMyWishlist('tok-abc');

		const [url, init] = fetchMock.mock.calls[0];
		expect(url).toContain('/me/wishlist');
		expect((init.headers as Record<string, string>).Authorization).toBe('Bearer tok-abc');
	});

	it('addToWishlist poste articleId et renvoie item/already', async () => {
		fetchMock.mockResolvedValue({
			ok: true,
			json: async () => ({ item: { ID: 1, articleId: 5 }, already: false })
		});

		const result = await addToWishlist('tok', 5);

		const [, init] = fetchMock.mock.calls[0];
		expect(init.method).toBe('POST');
		expect(JSON.parse(init.body as string)).toEqual({ articleId: 5 });
		expect(result.already).toBe(false);
	});

	it('removeFromWishlist envoie DELETE sur /me/wishlist/:articleId', async () => {
		fetchMock.mockResolvedValue({ ok: true, json: async () => ({ message: 'ok' }) });

		await removeFromWishlist('tok', 5);

		const [url, init] = fetchMock.mock.calls[0];
		expect(url).toContain('/me/wishlist/5');
		expect(init.method).toBe('DELETE');
	});

	it("rejette avec le message d'erreur du backend", async () => {
		fetchMock.mockResolvedValue({
			ok: false,
			status: 404,
			json: async () => ({ error: 'Article introuvable' })
		});

		await expect(addToWishlist('tok', 999)).rejects.toThrow('Article introuvable');
	});
});
