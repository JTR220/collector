import { writable, get } from 'svelte/store';
import { auth } from './auth';
import { fetchMyStats, type PlayerStats } from '$lib/api/engagement';

// Stats du joueur connecté, affichées dans le header et les pages gamifiées.
// refreshStats() est appelé après chaque action qui rapporte XP / gems.
export const playerStats = writable<PlayerStats | null>(null);

export async function refreshStats(): Promise<void> {
	const token = get(auth).token;
	if (!token) {
		playerStats.set(null);
		return;
	}
	try {
		playerStats.set(await fetchMyStats(token));
	} catch {
		playerStats.set(null);
	}
}
