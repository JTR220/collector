import type { ArticleAPI } from './catalog';

const BASE_URL = import.meta.env.VITE_CATALOG_API_BASE_URL ?? 'http://localhost:8081';

export type WishlistItem = {
	ID: number;
	articleId: number;
	article: ArticleAPI;
	CreatedAt: string;
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

export const fetchMyWishlist = (token: string) => request<WishlistItem[]>('/me/wishlist', token);

export const addToWishlist = (token: string, articleId: number) =>
	request<{ item: WishlistItem; already?: boolean }>('/me/wishlist', token, {
		method: 'POST',
		body: JSON.stringify({ articleId })
	});

export const removeFromWishlist = (token: string, articleId: number) =>
	request<{ message: string }>(`/me/wishlist/${articleId}`, token, { method: 'DELETE' });
