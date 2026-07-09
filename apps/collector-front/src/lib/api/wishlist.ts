import { apiRequest } from './http';
import { BASE_URL, type ArticleAPI } from './catalog';

export type WishlistItem = {
	ID: number;
	articleId: number;
	article: ArticleAPI;
	CreatedAt: string;
};

const request = <T>(path: string, init?: RequestInit) =>
	apiRequest<T>(BASE_URL, path, { init, errorPrefix: 'catalog-service' });

export const fetchMyWishlist = () => request<WishlistItem[]>('/me/wishlist');

export const addToWishlist = (articleId: number) =>
	request<{ item: WishlistItem; already?: boolean }>('/me/wishlist', {
		method: 'POST',
		body: JSON.stringify({ articleId })
	});

export const removeFromWishlist = (articleId: number) =>
	request<{ message: string }>(`/me/wishlist/${articleId}`, { method: 'DELETE' });
