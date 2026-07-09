import { env } from '$env/dynamic/public';
import { apiRequest } from './http';

const BASE_URL = env.PUBLIC_AUTH_API_BASE_URL ?? 'http://localhost:8080';

export type MeResponse = { id: number; name: string; email: string; role?: string };

export async function fetchMe(): Promise<MeResponse> {
	const res = await fetch(`${BASE_URL}/me`, { credentials: 'include' });
	if (!res.ok) throw new Error(`auth-service /me error: ${res.status}`);
	return res.json();
}

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
