import { describe, it, expect, beforeEach, vi } from 'vitest';

import { fetchMe } from './auth';

describe('fetchMe', () => {
	const fetchMock = vi.fn();

	beforeEach(() => {
		fetchMock.mockReset();
		vi.stubGlobal('fetch', fetchMock);
	});

	it('appelle /me avec le token Bearer et retourne le JSON', async () => {
		const user = { id: 1, name: 'Alice', email: 'a@b.c' };
		fetchMock.mockResolvedValue({
			ok: true,
			json: async () => user
		});

		const result = await fetchMe('tok-123');

		expect(fetchMock).toHaveBeenCalledTimes(1);
		const [url, init] = fetchMock.mock.calls[0];
		expect(url).toContain('/me');
		expect((init.headers as Record<string, string>).Authorization).toBe('Bearer tok-123');
		expect(result).toEqual(user);
	});

	it("rejette avec le statut HTTP quand la reponse n'est pas ok", async () => {
		fetchMock.mockResolvedValue({ ok: false, status: 401 });

		await expect(fetchMe('bad-token')).rejects.toThrow(/401/);
	});
});
