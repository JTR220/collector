import { apiRequest } from './http';
import { BASE_URL, type ArticleAPI } from './catalog';
import type { Order } from './market';

export type OfferStatus = 'pending' | 'accepted' | 'rejected' | 'purchased';

export type Offer = {
	ID: number;
	articleId: number;
	buyerId: number;
	sellerId: number;
	price: number;
	message: string;
	status: OfferStatus;
	article: ArticleAPI;
	CreatedAt: string;
};

export const OFFER_STATUS_LABELS: Record<OfferStatus, string> = {
	pending: 'En attente',
	accepted: 'Acceptée',
	rejected: 'Refusée',
	purchased: 'Payée'
};

const request = <T>(path: string, init?: RequestInit) =>
	apiRequest<T>(BASE_URL, path, { init, errorPrefix: 'catalog-service' });

export const createOffer = (articleId: number, price: number, message: string) =>
	request<{ offer: Offer }>(`/article/${articleId}/offer`, {
		method: 'POST',
		body: JSON.stringify({ price, message })
	});

/** Offres en attente reçues (en tant que vendeur). */
export const fetchReceivedOffers = () => request<Offer[]>('/me/offers/received');

/** Offres envoyées (en tant qu'acheteur), tous statuts confondus. */
export const fetchSentOffers = () => request<Offer[]>('/me/offers/sent');

export const acceptOffer = (offerId: number) =>
	request<{ offer: Offer }>(`/offer/${offerId}/accept`, { method: 'PATCH' });

export const rejectOffer = (offerId: number) =>
	request<{ offer: Offer }>(`/offer/${offerId}/reject`, { method: 'PATCH' });

export const payOffer = (offerId: number) =>
	request<{ order: Order }>(`/offer/${offerId}/pay`, { method: 'POST' });
