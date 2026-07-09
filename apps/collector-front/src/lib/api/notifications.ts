import { env } from '$env/dynamic/public';
import { browser } from '$app/environment';
import { apiRequest } from './http';

const BASE_URL = env.PUBLIC_NOTIFICATION_API_BASE_URL ?? 'http://localhost:8083';

export type NotificationType =
	| 'PRICE_DROP'
	| 'PRICE_SPIKE'
	| 'FRAUD_ALERT'
	| 'NEW_ITEM'
	| 'ITEM_SOLD'
	| 'ORDER_PENDING'
	| 'ORDER_ACCEPTED'
	| 'ORDER_REJECTED';

export type NotificationAPI = {
	id: string;
	user_id: string;
	type: NotificationType;
	title: string;
	body: string;
	item_id?: string;
	read: boolean;
	created_at: string;
};

export type WebSocketMessage = {
	event: string;
	data: Record<string, unknown>;
};

const request = <T>(path: string, init?: RequestInit) =>
	apiRequest<T>(BASE_URL, path, { init, errorPrefix: `notification-service ${path}` });

export async function fetchNotifications(
	limit = 50
): Promise<{ count: number; notifications: NotificationAPI[] }> {
	const data = await request<{ count: number; notifications: NotificationAPI[] | null }>(
		`/api/v1/notifications?limit=${limit}`
	);
	return { count: data.count, notifications: data.notifications ?? [] };
}

export async function markRead(id: string): Promise<void> {
	await request(`/api/v1/notifications/${id}/read`, { method: 'PUT' });
}

export async function markAllRead(): Promise<void> {
	await request('/api/v1/notifications/read-all', { method: 'PUT' });
}

export async function fetchUnreadCount(): Promise<number> {
	const data = await request<{ unread_count: number }>('/api/v1/notifications/unread-count');
	return data.unread_count;
}

// ── WebSocket avec reconnexion ────────────────────────────────────────────────

export type NotificationSocket = { close: () => void };

const INITIAL_BACKOFF_MS = 1000;
const MAX_BACKOFF_MS = 30000;

// Authentifie par le cookie httpOnly de session : le navigateur le joint
// automatiquement a la requete d'upgrade WS (meme domaine/sous-domaine),
// impossible de toute facon de le lire en JS pour le poser en ?token=.
export function connectNotifications(onMessage: (msg: WebSocketMessage) => void): NotificationSocket {
	if (!browser) return { close: () => {} };

	const wsUrl = BASE_URL.replace(/^http/, 'ws') + '/ws';
	let socket: WebSocket | null = null;
	let backoff = INITIAL_BACKOFF_MS;
	let closed = false;
	let retryTimer: ReturnType<typeof setTimeout> | null = null;

	function connect() {
		if (closed) return;
		socket = new WebSocket(wsUrl);

		socket.onopen = () => {
			backoff = INITIAL_BACKOFF_MS;
		};

		socket.onmessage = (e) => {
			try {
				onMessage(JSON.parse(e.data));
			} catch {
				// message non JSON ignore
			}
		};

		socket.onclose = () => {
			socket = null;
			if (closed) return;
			retryTimer = setTimeout(connect, backoff);
			backoff = Math.min(backoff * 2, MAX_BACKOFF_MS);
		};

		socket.onerror = () => {
			socket?.close();
		};
	}

	connect();

	return {
		close() {
			closed = true;
			if (retryTimer) clearTimeout(retryTimer);
			socket?.close();
		}
	};
}
