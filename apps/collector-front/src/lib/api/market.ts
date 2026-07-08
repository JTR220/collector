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
};

export const ORDER_STATUS_LABELS: Record<OrderStatus, string> = {
	pending: 'En attente de validation',
	paid: 'Payée',
	shipped: 'Expédiée',
	delivered: 'Livrée',
	cancelled: 'Annulée'
};

const request = <T>(path: string, token: string, init?: RequestInit) =>
	apiRequest<T>(BASE_URL, path, { token, init, errorPrefix: 'catalog-service' });

export const buyArticle = (token: string, articleId: number) =>
	request<{ order: Order }>(`/article/${articleId}/buy`, token, { method: 'POST' });

export const fetchMyOrders = (token: string) => request<Order[]>('/me/orders', token);

/** Ventes reçues (en tant que vendeur), y compris celles en attente de validation. */
export const fetchMySales = (token: string) => request<Order[]>('/me/sales', token);

export const acceptOrder = (token: string, orderId: number) =>
	request<{ order: Order }>(`/order/${orderId}/accept`, token, { method: 'PATCH' });

export const rejectOrder = (token: string, orderId: number) =>
	request<{ order: Order }>(`/order/${orderId}/reject`, token, { method: 'PATCH' });
