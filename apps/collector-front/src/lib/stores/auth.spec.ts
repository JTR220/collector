import { describe, it, expect, beforeEach, vi } from 'vitest';
import { get } from 'svelte/store';

// localStorage en memoire (absent de l'environnement node)
const store = new Map<string, string>();
vi.stubGlobal('localStorage', {
	getItem: (k: string) => store.get(k) ?? null,
	setItem: (k: string, v: string) => void store.set(k, v),
	removeItem: (k: string) => void store.delete(k),
	clear: () => store.clear()
});

import { auth, isAuthenticated } from './auth';

describe('auth store', () => {
	beforeEach(() => {
		store.clear();
		auth.logout();
	});

	it('login persiste le token et l’utilisateur', () => {
		auth.login('tok-123', { id: 1, name: 'Alice', email: 'a@b.c' });
		const state = get(auth);
		expect(state.token).toBe('tok-123');
		expect(state.user?.name).toBe('Alice');
		expect(localStorage.getItem('collector_token')).toBe('tok-123');
		expect(get(isAuthenticated)).toBe(true);
	});

	it('logout efface le token', () => {
		auth.login('tok-123', { id: 1, name: 'Alice', email: 'a@b.c' });
		auth.logout();
		expect(get(auth).token).toBeNull();
		expect(localStorage.getItem('collector_token')).toBeNull();
		expect(get(isAuthenticated)).toBe(false);
	});
});
