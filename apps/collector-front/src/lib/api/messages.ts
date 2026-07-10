import { env } from '$env/dynamic/public';
import { browser } from '$app/environment';
import { apiRequest } from './http';

const BASE_URL = env.PUBLIC_NOTIFICATION_API_BASE_URL ?? 'http://localhost:8083';

export type MessageAPI = {
	id: string;
	conversation_id: string;
	sender_id: string;
	sender_name: string;
	recipient_id: string;
	recipient_name: string;
	article_id?: string;
	article_name?: string;
	body: string;
	read: boolean;
	created_at: string;
};

export type ConversationAPI = {
	conversation_id: string;
	other_user_id: string;
	other_user_name: string;
	article_id?: string;
	article_name?: string;
	last_message: string;
	last_at: string;
	unread_count: number;
};

const request = <T>(path: string, init?: RequestInit) =>
	apiRequest<T>(BASE_URL, path, { init, errorPrefix: `notification-service ${path}` });

export type SendMessageInput = {
	recipientId: string;
	body: string;
	articleId?: string | number;
	articleName?: string;
};

export async function sendMessage(input: SendMessageInput): Promise<MessageAPI> {
	const data = await request<{ message: MessageAPI }>('/api/v1/messages', {
		method: 'POST',
		body: JSON.stringify({
			recipient_id: input.recipientId,
			body: input.body,
			article_id: input.articleId !== undefined ? String(input.articleId) : undefined,
			article_name: input.articleName
		})
	});
	return data.message;
}

export async function fetchConversations(): Promise<ConversationAPI[]> {
	const data = await request<{ conversations: ConversationAPI[] | null }>('/api/v1/conversations');
	return data.conversations ?? [];
}

export async function fetchConversationMessages(conversationId: string): Promise<MessageAPI[]> {
	const data = await request<{ messages: MessageAPI[] | null }>(
		`/api/v1/conversations/${conversationId}/messages`
	);
	return data.messages ?? [];
}

export async function markConversationRead(conversationId: string): Promise<void> {
	await request(`/api/v1/conversations/${conversationId}/read`, { method: 'PUT' });
}

/** Convertit un id numerique (auth-service) en UUID deterministe partagé par tous les services. */
export function toUserUUID(id: number): string {
	return `00000000-0000-0000-0000-${id.toString(16).padStart(12, '0')}`;
}

// ── WebSocket dedie a la messagerie (connexion separee de celle des notifications) ──

export type MessageSocket = { close: () => void };

const INITIAL_BACKOFF_MS = 1000;
const MAX_BACKOFF_MS = 30000;

// Authentifie par le cookie httpOnly de session (voir connectNotifications).
export function connectMessages(onMessage: (msg: MessageAPI) => void): MessageSocket {
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
				const parsed = JSON.parse(e.data);
				if (parsed.event === 'NEW_MESSAGE' && parsed.data) {
					onMessage(parsed.data as MessageAPI);
				}
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
