import { writable, derived, get } from 'svelte/store';
import {
	fetchNotifications,
	fetchUnreadCount,
	markRead,
	markAllRead,
	connectNotifications,
	type NotificationAPI,
	type NotificationSocket
} from '$lib/api/notifications';

type NotificationsState = {
	items: NotificationAPI[];
	unreadCount: number;
};

function createNotifications() {
	const store = writable<NotificationsState>({ items: [], unreadCount: 0 });
	let socket: NotificationSocket | null = null;

	return {
		subscribe: store.subscribe,

		// Hydrate en REST puis ouvre le WebSocket. Appele au login / au mount.
		async start(token: string) {
			this.stop();
			try {
				const [{ notifications }, unreadCount] = await Promise.all([
					fetchNotifications(token),
					fetchUnreadCount(token)
				]);
				store.set({ items: notifications, unreadCount });
			} catch {
				// services notifications indisponibles : on garde l'etat vide
			}

			socket = connectNotifications(token, (msg) => {
				const data = msg.data ?? {};
				const incoming: NotificationAPI = {
					id: String(data.notification_id ?? crypto.randomUUID()),
					user_id: '',
					type: (msg.event as NotificationAPI['type']) ?? 'NEW_ITEM',
					title: String(data.title ?? msg.event),
					body: String(data.body ?? ''),
					item_id: data.item_id ? String(data.item_id) : undefined,
					read: false,
					created_at: String(data.created_at ?? new Date().toISOString())
				};
				store.update((s) => ({
					items: [incoming, ...s.items].slice(0, 100),
					unreadCount: s.unreadCount + 1
				}));
			});
		},

		stop() {
			socket?.close();
			socket = null;
		},

		reset() {
			this.stop();
			store.set({ items: [], unreadCount: 0 });
		},

		async markRead(token: string, id: string) {
			const state = get(store);
			const target = state.items.find((n) => n.id === id);
			if (!target || target.read) return;
			store.update((s) => ({
				items: s.items.map((n) => (n.id === id ? { ...n, read: true } : n)),
				unreadCount: Math.max(0, s.unreadCount - 1)
			}));
			try {
				await markRead(token, id);
			} catch {
				// echec silencieux : l'etat local reste optimiste
			}
		},

		async markAllRead(token: string) {
			store.update((s) => ({
				items: s.items.map((n) => ({ ...n, read: true })),
				unreadCount: 0
			}));
			try {
				await markAllRead(token);
			} catch {
				// echec silencieux
			}
		}
	};
}

export const notifications = createNotifications();
export const unreadCount = derived(notifications, ($n) => $n.unreadCount);
