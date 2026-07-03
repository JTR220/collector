import { env } from '$env/dynamic/public';
const BASE_URL = env.PUBLIC_AUTH_API_BASE_URL ?? 'http://localhost:8080';

export type MeResponse = { id: number; name: string; email: string };

export async function fetchMe(token: string): Promise<MeResponse> {
	const res = await fetch(`${BASE_URL}/me`, {
		headers: { Authorization: `Bearer ${token}` }
	});
	if (!res.ok) throw new Error(`auth-service /me error: ${res.status}`);
	return res.json();
}
