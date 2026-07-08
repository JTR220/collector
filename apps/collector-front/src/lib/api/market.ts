import { apiRequest } from './http';
import { BASE_URL, type ArticleAPI } from './catalog';

export type OrderStatus = 'paid' | 'shipped' | 'delivered' | 'cancelled';

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
