import { env } from '$env/dynamic/public';
import { apiRequest } from './http';

const BASE_URL = env.PUBLIC_AUTH_API_BASE_URL ?? 'http://localhost:8080';

export type MeResponse = { id: number; name: string; email: string; role?: string };

export async function fetchMe(token: string): Promise<MeResponse> {
	const res = await fetch(`${BASE_URL}/me`, {
		headers: { Authorization: `Bearer ${token}` }
	});
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

const adminRequest = <T>(path: string, token: string, init?: RequestInit) =>
	apiRequest<T>(BASE_URL, path, { token, init, errorPrefix: 'auth-service' });

export const fetchUsers = (token: string) => adminRequest<AdminUser[]>('/admin/users', token);

export const suspendUser = (token: string, id: number) =>
	adminRequest<{ id: number; suspended: boolean }>(`/admin/users/${id}/suspend`, token, {
		method: 'PATCH'
	});

export const unsuspendUser = (token: string, id: number) =>
	adminRequest<{ id: number; suspended: boolean }>(`/admin/users/${id}/unsuspend`, token, {
		method: 'PATCH'
	});
