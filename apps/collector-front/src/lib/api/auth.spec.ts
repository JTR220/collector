import { describe, it, expect, beforeEach, vi } from 'vitest';

import { fetchMe, updateProfile, exportMyData, deleteMyAccount } from './auth';

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

describe('updateProfile', () => {
	const fetchMock = vi.fn();

	beforeEach(() => {
		fetchMock.mockReset();
		vi.stubGlobal('fetch', fetchMock);
	});

	it('envoie un PATCH /me avec les champs modifies', async () => {
		const updated = { id: 1, name: 'Alice B.', email: 'alice.b@example.com' };
		fetchMock.mockResolvedValue({ ok: true, json: async () => updated });

		const result = await updateProfile({ name: 'Alice B.', email: 'alice.b@example.com' });

		expect(fetchMock).toHaveBeenCalledTimes(1);
		const [url, init] = fetchMock.mock.calls[0];
		expect(url).toContain('/me');
		expect(init.method).toBe('PATCH');
		expect(init.credentials).toBe('include');
		expect(JSON.parse(init.body)).toEqual({ name: 'Alice B.', email: 'alice.b@example.com' });
		expect(result).toEqual(updated);
	});
});

describe('exportMyData', () => {
	const fetchMock = vi.fn();

	beforeEach(() => {
		fetchMock.mockReset();
		vi.stubGlobal('fetch', fetchMock);
	});

	it('recupere les donnees personnelles depuis /me/export', async () => {
		const payload = { exported_at: '2026-07-09T00:00:00Z', account: { id: 1 } };
		fetchMock.mockResolvedValue({ ok: true, json: async () => payload });

		const result = await exportMyData();

		const [url] = fetchMock.mock.calls[0];
		expect(url).toContain('/me/export');
		expect(result).toEqual(payload);
	});
});

describe('deleteMyAccount', () => {
	const fetchMock = vi.fn();

	beforeEach(() => {
		fetchMock.mockReset();
		vi.stubGlobal('fetch', fetchMock);
	});

	it('envoie un DELETE /me', async () => {
		fetchMock.mockResolvedValue({ ok: true, json: async () => ({ message: 'ok' }) });

		await deleteMyAccount();

		const [url, init] = fetchMock.mock.calls[0];
		expect(url).toContain('/me');
		expect(init.method).toBe('DELETE');
	});
});
