import { redirect } from '@sveltejs/kit';

export const ssr = false;

// Le JWT est dans un cookie httpOnly (invisible en JS) : cette garde est
// optimiste, basee sur le profil utilisateur non sensible mis en cache
// (voir lib/stores/auth.ts). Si le cookie est absent/expire malgre tout, les
// appels API protegés echoueront en 401 et redirigeront vers /login.
export function load() {
	if (!localStorage.getItem('collector_user')) {
		throw redirect(307, '/login');
	}
}
