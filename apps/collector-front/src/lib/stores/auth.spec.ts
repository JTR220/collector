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

// logout() appelle POST /logout (best-effort) : $app/environment.browser vaut
// false par defaut hors navigateur, donc cet appel ne se declenche pas dans
// ce test, mais on stub fetch quand meme pour ne pas dependre de ce detail.
vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true }));

import { auth, isAuthenticated } from './auth';

describe('auth store', () => {
	beforeEach(() => {
		store.clear();
		auth.logout();
	});

	it('login persiste le profil utilisateur (jamais de token, cookie httpOnly)', () => {
		auth.login({ id: 1, name: 'Alice', email: 'a@b.c' });
		const state = get(auth);
		expect(state.user?.name).toBe('Alice');
		expect(JSON.parse(localStorage.getItem('collector_user') ?? 'null')).toEqual({
			id: 1,
			name: 'Alice',
			email: 'a@b.c'
		});
		expect(get(isAuthenticated)).toBe(true);
	});

	it('logout efface le profil utilisateur en cache', () => {
		auth.login({ id: 1, name: 'Alice', email: 'a@b.c' });
		auth.logout();
		expect(get(auth).user).toBeNull();
		expect(localStorage.getItem('collector_user')).toBeNull();
		expect(get(isAuthenticated)).toBe(false);
	});
});
