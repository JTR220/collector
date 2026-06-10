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

export type ListingInput = {
	name: string;
	description: string;
	series?: string;
	year?: number;
	rarity?: string;
	grade?: string;
	prix: number;
	fraisPort?: number;
	categoryId: number;
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

export const createListing = (token: string, input: ListingInput) =>
	request<{ article: ArticleAPI; xp: number }>('/market/listings', token, {
		method: 'POST',
		body: JSON.stringify(input)
	});

export const deleteListing = (token: string, articleId: number) =>
	request<{ message: string }>(`/market/listings/${articleId}`, token, { method: 'DELETE' });

export const fetchMyListings = (token: string) => request<ArticleAPI[]>('/me/listings', token);

/** Upload multipart de la photo d'un article — ne pas fixer Content-Type, le navigateur gère le boundary. */
export async function uploadArticleImage(
	token: string,
	articleId: number,
	file: File
): Promise<{ imageUrl: string }> {
	const form = new FormData();
	form.append('image', file);
	const res = await fetch(`${BASE_URL}/article/${articleId}/image`, {
		method: 'POST',
		headers: { Authorization: `Bearer ${token}` },
		body: form
	});
	const data = await res.json().catch(() => ({}));
	if (!res.ok) throw new Error(data.error ?? `catalog-service error: ${res.status}`);
	return data as { imageUrl: string };
}

export const buyArticle = (token: string, articleId: number) =>
	request<{ order: Order; xp: number }>(`/article/${articleId}/buy`, token, { method: 'POST' });

export const fetchMyOrders = (token: string) => request<Order[]>('/me/orders', token);

export const fetchMySales = (token: string) => request<Order[]>('/me/sales', token);

export const updateOrderStatus = (token: string, orderId: number, status: OrderStatus) =>
	request<Order>(`/orders/${orderId}/status`, token, {
		method: 'POST',
		body: JSON.stringify({ status })
	});
