import { apiRequest } from './http';
import { BASE_URL, type ArticleAPI } from './catalog';

export type OrderStatus = 'pending' | 'paid' | 'shipped' | 'delivered' | 'cancelled';

export type Order = {
	ID: number;
	buyerId: number;
	sellerId: number;
	articleId: number;
	price: number;
	fraisPort: number;
	status: OrderStatus;
	article: ArticleAPI;
	CreatedAt: string;
	reviewed: boolean;
};

export const ORDER_STATUS_LABELS: Record<OrderStatus, string> = {
	pending: 'En attente de validation',
	paid: 'Payée',
	shipped: 'Expédiée',
	delivered: 'Livrée',
	cancelled: 'Annulée'
};

const request = <T>(path: string, init?: RequestInit) =>
	apiRequest<T>(BASE_URL, path, { init, errorPrefix: 'catalog-service' });

export const buyArticle = (articleId: number) =>
	request<{ order: Order }>(`/article/${articleId}/buy`, { method: 'POST' });

export const fetchMyOrders = () => request<Order[]>('/me/orders');

/** Ventes reçues (en tant que vendeur), y compris celles en attente de validation. */
export const fetchMySales = () => request<Order[]>('/me/sales');

export const acceptOrder = (orderId: number) =>
	request<{ order: Order }>(`/order/${orderId}/accept`, { method: 'PATCH' });

export const rejectOrder = (orderId: number) =>
	request<{ order: Order }>(`/order/${orderId}/reject`, { method: 'PATCH' });

// ── Avis vendeur ─────────────────────────────────────────────────────────────

export type Review = {
	ID: number;
	orderId: number;
	reviewerId: number;
	reviewerName: string;
	sellerId: number;
	rating: number;
	comment: string;
	CreatedAt: string;
};

export type SellerRating = { average: number; count: number };

export const leaveReview = (orderId: number, rating: number, comment: string) =>
	request<{ review: Review }>(`/order/${orderId}/review`, {
		method: 'POST',
		body: JSON.stringify({ rating, comment })
	});

export async function fetchSellerRating(sellerId: number): Promise<SellerRating> {
	const res = await fetch(`${BASE_URL}/sellers/${sellerId}/rating`);
	if (!res.ok) throw new Error(`catalog-service error: ${res.status}`);
	return res.json();
}

export async function fetchSellerReviews(sellerId: number): Promise<Review[]> {
	const res = await fetch(`${BASE_URL}/sellers/${sellerId}/reviews`);
	if (!res.ok) throw new Error(`catalog-service error: ${res.status}`);
	return res.json();
}
