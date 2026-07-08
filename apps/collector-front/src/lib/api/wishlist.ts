import { apiRequest } from './http';
import { BASE_URL, type ArticleAPI } from './catalog';

export type WishlistItem = {
	ID: number;
	articleId: number;
	article: ArticleAPI;
	CreatedAt: string;
};

const request = <T>(path: string, token: string, init?: RequestInit) =>
	apiRequest<T>(BASE_URL, path, { token, init, errorPrefix: 'catalog-service' });

export const fetchMyWishlist = (token: string) => request<WishlistItem[]>('/me/wishlist', token);

export const addToWishlist = (token: string, articleId: number) =>
	request<{ item: WishlistItem; already?: boolean }>('/me/wishlist', token, {
		method: 'POST',
		body: JSON.stringify({ articleId })
	});

export const removeFromWishlist = (token: string, articleId: number) =>
	request<{ message: string }>(`/me/wishlist/${articleId}`, token, { method: 'DELETE' });
