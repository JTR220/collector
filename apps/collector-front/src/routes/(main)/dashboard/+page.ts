import { redirect } from '@sveltejs/kit';

export const ssr = false;

export function load() {
	if (!localStorage.getItem('collector_token')) {
		throw redirect(307, '/login');
	}
}
