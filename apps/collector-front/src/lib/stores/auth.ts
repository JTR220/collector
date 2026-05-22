import { writable, derived } from 'svelte/store';
import { browser } from '$app/environment';

export type AuthUser = { id: number; name: string; email: string };
type AuthState = { token: string | null; user: AuthUser | null };

function createAuth() {
	const initial: AuthState = browser
		? {
				token: localStorage.getItem('collector_token'),
				user: JSON.parse(localStorage.getItem('collector_user') ?? 'null')
			}
		: { token: null, user: null };

	const { subscribe, set } = writable<AuthState>(initial);

	return {
		subscribe,
		login(token: string, user: AuthUser) {
			localStorage.setItem('collector_token', token);
			localStorage.setItem('collector_user', JSON.stringify(user));
			set({ token, user });
		},
		logout() {
			localStorage.removeItem('collector_token');
			localStorage.removeItem('collector_user');
			set({ token: null, user: null });
		}
	};
}

export const auth = createAuth();
export const isAuthenticated = derived(auth, ($auth) => !!$auth.token);
