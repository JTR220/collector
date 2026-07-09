import { describe, it, expect, beforeEach, vi } from 'vitest';

import { fetchMe } from './auth';

describe('fetchMe', () => {
	const fetchMock = vi.fn();

	beforeEach(() => {
		fetchMock.mockReset();
		vi.stubGlobal('fetch', fetchMock);
	});

	it('appelle /me avec credentials:include (cookie de session) et retourne le JSON', async () => {
		const user = { id: 1, name: 'Alice', email: 'a@b.c' };
		fetchMock.mockResolvedValue({
			ok: true,
			json: async () => user
		});

		const result = await fetchMe();

		expect(fetchMock).toHaveBeenCalledTimes(1);
		const [url, init] = fetchMock.mock.calls[0];
		expect(url).toContain('/me');
		expect(init.credentials).toBe('include');
		expect(result).toEqual(user);
	});

	it("rejette avec le statut HTTP quand la reponse n'est pas ok", async () => {
		fetchMock.mockResolvedValue({ ok: false, status: 401 });

		await expect(fetchMe()).rejects.toThrow(/401/);
	});
});
