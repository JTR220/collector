import type { ArticleAPI } from './catalog';

const BASE_URL = import.meta.env.VITE_CATALOG_API_BASE_URL ?? 'http://localhost:8081';

export type PlayerStats = {
	xp: number;
	gems: number;
	streak: number;
	level: number;
	xpToNext: number;
	wishlistCount: number;
	journalCount: number;
};

export type DropEntryKind = 'purchase' | 'raffle' | 'reminder' | 'waitlist';

export type DropEntry = {
	ID: number;
	userId: number;
	articleId: number;
	kind: DropEntryKind;
};

export type WishlistItem = {
	ID: number;
	articleId: number;
	article: ArticleAPI;
	CreatedAt: string;
};

export type JournalEntryKind = 'acquis' | 'vendu' | 'noté' | 'trade' | 'wishlist';

export type JournalEntry = {
	ID: number;
	articleId: number;
	kind: JournalEntryKind;
	rating: number;
	note: string;
	likes: number;
	xp: number;
	article: ArticleAPI;
	CreatedAt: string;
};

export type Quest = {
	ID: number;
	code: string;
	title: string;
	kind: 'daily' | 'weekly' | 'mission';
	xp: number;
	target: number;
	progress: number;
	done: boolean;
};

export type LeagueRow = {
	name: string;
	level: number;
	xp: number;
	delta: number;
	me: boolean;
};

async function request<T>(path: string, token: string, init?: RequestInit): Promise<T> {
	const res = await fetch(`${BASE_URL}${path}`, {
		...init,
		headers: {
			'Content-Type': 'application/json',
			Authorization: `Bearer ${token}`,
			...init?.headers
		}
	});
	const data = await res.json().catch(() => ({}));
	if (!res.ok) throw new Error(data.error ?? `catalog-service error: ${res.status}`);
	return data as T;
}

export const fetchMyStats = (token: string) => request<PlayerStats>('/me/stats', token);

export const fetchMyDropEntries = (token: string) => request<DropEntry[]>('/me/entries', token);

export const createDropEntry = (token: string, articleId: number, kind: DropEntryKind) =>
	request<{
		entry: DropEntry;
		xp?: number;
		seatsLeft?: number;
		dropStatus?: string;
		already?: boolean;
	}>(`/article/${articleId}/entry`, token, { method: 'POST', body: JSON.stringify({ kind }) });

export const fetchMyWishlist = (token: string) => request<WishlistItem[]>('/me/wishlist', token);

export const addToWishlist = (token: string, articleId: number) =>
	request<{ item: WishlistItem; xp?: number; already?: boolean }>('/me/wishlist', token, {
		method: 'POST',
		body: JSON.stringify({ articleId })
	});

export const removeFromWishlist = (token: string, articleId: number) =>
	request<{ message: string }>(`/me/wishlist/${articleId}`, token, { method: 'DELETE' });

export const fetchMyJournal = (token: string) => request<JournalEntry[]>('/me/journal', token);

export const createJournalEntry = (
	token: string,
	input: { articleId: number; kind: JournalEntryKind; rating?: number; note?: string }
) => request<JournalEntry>('/me/journal', token, { method: 'POST', body: JSON.stringify(input) });

export const likeJournalEntry = (token: string, entryId: number) =>
	request<{ id: number; likes: number }>(`/me/journal/${entryId}/like`, token, { method: 'POST' });

export const fetchMyQuests = (token: string) => request<Quest[]>('/me/quests', token);

export const progressQuest = (token: string, questId: number) =>
	request<Quest>(`/me/quests/${questId}/progress`, token, { method: 'POST' });

export const skipQuest = (token: string, questId: number) =>
	request<{ quest: Quest; gems: number }>(`/me/quests/${questId}/skip`, token, { method: 'POST' });

export const fetchLeague = (token: string) => request<LeagueRow[]>('/league', token);

export const challengeRival = (token: string, rival: string) =>
	request<Quest | { quest: Quest; already: boolean }>('/league/challenge', token, {
		method: 'POST',
		body: JSON.stringify({ rival })
	});
