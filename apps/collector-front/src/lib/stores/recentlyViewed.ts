import { writable } from 'svelte/store';
import { browser } from '$app/environment';
import type { ArticleAPI } from '$lib/api/catalog';

const STORAGE_KEY = 'collector_recently_viewed';
const MAX_ITEMS = 8;

function loadInitial(): ArticleAPI[] {
	if (!browser) return [];
	try {
		const raw = localStorage.getItem(STORAGE_KEY);
		return raw ? (JSON.parse(raw) as ArticleAPI[]) : [];
	} catch {
		return [];
	}
}

function createRecentlyViewed() {
	const store = writable<ArticleAPI[]>(loadInitial());

	if (browser) {
		store.subscribe((items) => {
			localStorage.setItem(STORAGE_KEY, JSON.stringify(items));
		});
	}

	return {
		subscribe: store.subscribe,

		// Pousse l'article consulté en tête de liste (plus récent en premier),
		// sans doublon, plafonné à MAX_ITEMS.
		push(article: ArticleAPI) {
			store.update((items) =>
				[article, ...items.filter((i) => i.ID !== article.ID)].slice(0, MAX_ITEMS)
			);
		}
	};
}

export const recentlyViewed = createRecentlyViewed();
