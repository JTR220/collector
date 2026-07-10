import { env } from '$env/dynamic/public';
import { apiRequest } from './http';

const BASE_URL = env.PUBLIC_AUTH_API_BASE_URL ?? 'http://localhost:8080';

export type MeResponse = { id: number; name: string; email: string; role?: string };

export async function fetchMe(): Promise<MeResponse> {
	const res = await fetch(`${BASE_URL}/me`, { credentials: 'include' });
	if (!res.ok) throw new Error(`auth-service /me error: ${res.status}`);
	return res.json();
}

const meRequest = <T>(path: string, init?: RequestInit) =>
	apiRequest<T>(BASE_URL, path, { init, errorPrefix: 'auth-service' });

export type UpdateProfileInput = { name: string; email: string; password?: string };

// updateProfile modifie le profil de l'utilisateur connecte (droit de
// rectification, art. 16 RGPD).
export const updateProfile = (input: UpdateProfileInput) =>
	meRequest<MeResponse>('/me', { method: 'PATCH', body: JSON.stringify(input) });

export type DataExport = {
	exported_at: string;
	account: {
		id: number;
		name: string;
		email: string;
		role: string;
		suspended: boolean;
		created_at: string;
		updated_at: string;
	};
};

// exportMyData recupere l'integralite des donnees personnelles detenues sur
// le compte connecte (droit a la portabilite, art. 20 RGPD), a telecharger
// cote client (voir profil/+page.svelte).
export const exportMyData = () => meRequest<DataExport>('/me/export');

// deleteMyAccount supprime definitivement le compte et les donnees
// personnelles associees (droit a l'effacement, art. 17 RGPD).
export const deleteMyAccount = () => meRequest<{ message: string }>('/me', { method: 'DELETE' });

// ── Moderation (admin) ───────────────────────────────────────────────────────

export type AdminUser = {
	ID: number;
	name: string;
	email: string;
	role: string;
	suspended: boolean;
};

const adminRequest = <T>(path: string, init?: RequestInit) =>
	apiRequest<T>(BASE_URL, path, { init, errorPrefix: 'auth-service' });

export const fetchUsers = () => adminRequest<AdminUser[]>('/admin/users');

export const suspendUser = (id: number) =>
	adminRequest<{ id: number; suspended: boolean }>(`/admin/users/${id}/suspend`, {
		method: 'PATCH'
	});

export const unsuspendUser = (id: number) =>
	adminRequest<{ id: number; suspended: boolean }>(`/admin/users/${id}/unsuspend`, {
		method: 'PATCH'
	});
