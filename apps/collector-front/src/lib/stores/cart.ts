import { writable, derived } from 'svelte/store';
import { browser } from '$app/environment';
import type { ArticleAPI } from '$lib/api/catalog';

const STORAGE_KEY = 'collector_cart';

function loadInitial(): ArticleAPI[] {
	if (!browser) return [];
	try {
		const raw = localStorage.getItem(STORAGE_KEY);
		return raw ? (JSON.parse(raw) as ArticleAPI[]) : [];
	} catch {
		return [];
	}
}

function createCart() {
	const store = writable<ArticleAPI[]>(loadInitial());

	if (browser) {
		store.subscribe((items) => {
			localStorage.setItem(STORAGE_KEY, JSON.stringify(items));
		});
	}

	return {
		subscribe: store.subscribe,

		add(article: ArticleAPI) {
			store.update((items) =>
				items.some((i) => i.ID === article.ID) ? items : [...items, article]
			);
		},

		remove(articleId: number) {
			store.update((items) => items.filter((i) => i.ID !== articleId));
		},

		clear() {
			store.set([]);
		},

		has(articleId: number, items: ArticleAPI[]) {
			return items.some((i) => i.ID === articleId);
		}
	};
}

export const cart = createCart();
export const cartCount = derived(cart, ($c) => $c.length);
export const cartTotal = derived(cart, ($c) =>
	$c.reduce((sum, a) => sum + a.prix + a.fraisPort, 0)
);
