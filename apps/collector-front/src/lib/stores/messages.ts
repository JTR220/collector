import { writable, derived, get } from 'svelte/store';
import {
	fetchConversations,
	connectMessages,
	type ConversationAPI,
	type MessageSocket
} from '$lib/api/messages';

type MessagesState = {
	conversations: ConversationAPI[];
	unreadCount: number;
};

function createMessages() {
	const store = writable<MessagesState>({ conversations: [], unreadCount: 0 });
	let socket: MessageSocket | null = null;

	async function refresh() {
		try {
			const conversations = await fetchConversations();
			const unreadCount = conversations.reduce((sum, c) => sum + c.unread_count, 0);
			store.set({ conversations, unreadCount });
		} catch {
			// notification-service indisponible : on garde l'etat vide
		}
	}

	return {
		subscribe: store.subscribe,

		async start() {
			this.stop();
			await refresh();
			socket = connectMessages(() => {
				// Nouveau message (envoye ou recu) : on rafraichit la liste des fils
				// pour recalculer apercu + compteur, plus simple et fiable que du
				// merge manuel cote client.
				refresh();
			});
		},

		stop() {
			socket?.close();
			socket = null;
		},

		reset() {
			this.stop();
			store.set({ conversations: [], unreadCount: 0 });
		},

		refresh
	};
}

export const messages = createMessages();
export const unreadMessagesCount = derived(messages, ($m) => $m.unreadCount);
