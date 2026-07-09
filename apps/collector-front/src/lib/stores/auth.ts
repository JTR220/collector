import { writable, derived } from 'svelte/store';
import { browser } from '$app/environment';
import { env } from '$env/dynamic/public';

const AUTH_BASE_URL = env.PUBLIC_AUTH_API_BASE_URL ?? 'http://localhost:8080';

export type AuthUser = { id: number; name: string; email: string; role?: string };
type AuthState = { user: AuthUser | null };

// Le JWT lui-meme ne transite plus jamais par du JS (cookie httpOnly de
// session, voir auth-service) : seul le profil utilisateur — non sensible —
// est mis en cache ici pour un affichage optimiste (garde de route
// synchrone, en-tete) sans attendre un aller-retour reseau. Il n'a aucune
// valeur d'authentification : chaque appel API protege reste verifie
// serveur via le cookie (credentials:'include').
function createAuth() {
	const initial: AuthState = browser
		? { user: JSON.parse(localStorage.getItem('collector_user') ?? 'null') }
		: { user: null };

	const { subscribe, set } = writable<AuthState>(initial);

	return {
		subscribe,
		login(user: AuthUser) {
			localStorage.setItem('collector_user', JSON.stringify(user));
			set({ user });
		},
		logout() {
			localStorage.removeItem('collector_user');
			set({ user: null });
			// Efface le cookie de session cote serveur. Best-effort (fire-and-
			// forget) : meme si l'appel echoue, l'etat local est deja nettoye.
			if (browser) {
				fetch(`${AUTH_BASE_URL}/logout`, { method: 'POST', credentials: 'include' }).catch(
					() => {}
				);
			}
		}
	};
}

export const auth = createAuth();
export const isAuthenticated = derived(auth, ($auth) => !!$auth.user);
