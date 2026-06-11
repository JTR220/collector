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

export const buyArticle = (token: string, articleId: number) =>
	request<{ order: Order }>(`/article/${articleId}/buy`, token, { method: 'POST' });

export const fetchMyOrders = (token: string) => request<Order[]>('/me/orders', token);
