import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';

// connectNotifications court-circuite hors navigateur : on force browser=true.
vi.mock('$app/environment', () => ({ browser: true }));

import {
	fetchNotifications,
	fetchUnreadCount,
	connectNotifications,
	type NotificationSocket
} from './notifications';

describe('notifications REST', () => {
	const fetchMock = vi.fn();

	beforeEach(() => {
		fetchMock.mockReset();
		vi.stubGlobal('fetch', fetchMock);
	});

	it('fetchNotifications appelle la bonne URL avec le token Bearer', async () => {
		fetchMock.mockResolvedValue({
			ok: true,
			json: async () => ({ count: 0, notifications: [] })
		});

		await fetchNotifications('tok-abc', 10);

		const [url, init] = fetchMock.mock.calls[0];
		expect(url).toContain('/api/v1/notifications?limit=10');
		expect((init.headers as Record<string, string>).Authorization).toBe('Bearer tok-abc');
	});

	it('fetchUnreadCount extrait unread_count', async () => {
		fetchMock.mockResolvedValue({
			ok: true,
			json: async () => ({ unread_count: 7 })
		});

		const count = await fetchUnreadCount('tok-abc');
		expect(count).toBe(7);
	});

	it('rejette sur reponse non-ok', async () => {
		fetchMock.mockResolvedValue({ ok: false, status: 500 });
		await expect(fetchUnreadCount('tok')).rejects.toThrow(/500/);
	});
});

describe('connectNotifications (reconnexion)', () => {
	class FakeWebSocket {
		static instances: FakeWebSocket[] = [];
		url: string;
		onopen: (() => void) | null = null;
		onmessage: ((e: { data: string }) => void) | null = null;
		onclose: (() => void) | null = null;
		onerror: (() => void) | null = null;
		constructor(url: string) {
			this.url = url;
			FakeWebSocket.instances.push(this);
		}
		close() {}
	}

	let socket: NotificationSocket;

	beforeEach(() => {
		FakeWebSocket.instances = [];
		vi.useFakeTimers();
		vi.stubGlobal('WebSocket', FakeWebSocket as unknown as typeof WebSocket);
	});

	afterEach(() => {
		socket?.close();
		vi.useRealTimers();
	});

	it('ouvre une connexion WebSocket vers /ws?token=', () => {
		socket = connectNotifications('tok-xyz', () => {});
		expect(FakeWebSocket.instances).toHaveLength(1);
		// nosemgrep: javascript.lang.security.detect-insecure-websocket -- URL de dev attendue dans le test
		expect(FakeWebSocket.instances[0].url).toBe('ws://localhost:8083/ws?token=tok-xyz');
	});

	it('se reconnecte apres une fermeture', () => {
		socket = connectNotifications('tok-xyz', () => {});
		expect(FakeWebSocket.instances).toHaveLength(1);

		// Simule une perte de connexion : un nouveau socket apparait apres le backoff.
		FakeWebSocket.instances[0].onclose?.();
		vi.advanceTimersByTime(1000);
		expect(FakeWebSocket.instances).toHaveLength(2);
	});

	it('ne se reconnecte pas apres close()', () => {
		socket = connectNotifications('tok-xyz', () => {});
		socket.close();
		FakeWebSocket.instances[0].onclose?.();
		vi.advanceTimersByTime(5000);
		expect(FakeWebSocket.instances).toHaveLength(1);
	});
});
