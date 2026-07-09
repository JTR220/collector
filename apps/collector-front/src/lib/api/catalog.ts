import { env } from '$env/dynamic/public';
import { apiRequest } from './http';

export const BASE_URL = env.PUBLIC_CATALOG_API_BASE_URL ?? 'http://localhost:8081';

export type CategoryAPI = {
	ID: number;
	name: string;
	description: string;
};

export type ArticleAPI = {
	ID: number;
	slug: string;
	name: string;
	description: string;
	series: string;
	year: number;
	rarity: string;
	rarityScore: number;
	grade: string;
	prix: number;
	fraisPort: number;
	seller: string;
	sellerId: number;
	sellerScore: number;
	imageUrl: string;
	images: string[];
	saleType: 'drop' | 'direct';
	sold: boolean;
	views: number;
	delta: number;
	priceHistory: number[];
	glyph: string;
	dropId: string;
	dropStatus: 'live' | 'next' | 'sold' | 'soon';
	dropDate: string;
	seatsLeft: number;
	seatsTotal: number;
	resellPrice: number;
	categoryId: number;
	category: CategoryAPI;
};

function resolveImageURL(url: string): string {
	return url.startsWith('http') ? url : `${BASE_URL}${url}`;
}

/** Résout l'URL d'affichage de la photo d'un article (URL absolue, ou chemin relatif hérité résolu sur le catalog-service). */
export function articleImage(article: Pick<ArticleAPI, 'imageUrl'>): string | null {
	if (!article.imageUrl) return null;
	return resolveImageURL(article.imageUrl);
}

/** Résout la galerie complète d'un article (retombe sur la seule couverture si aucune galerie n'a été renseignée). */
export function articleImages(article: Pick<ArticleAPI, 'imageUrl' | 'images'>): string[] {
	const gallery = (article.images ?? []).filter(Boolean);
	if (gallery.length > 0) return gallery.map(resolveImageURL);
	return article.imageUrl ? [resolveImageURL(article.imageUrl)] : [];
}

export async function fetchArticle(id: number | string): Promise<ArticleAPI> {
	const res = await fetch(`${BASE_URL}/article/${id}`);
	if (!res.ok) throw new Error(`catalog-service error: ${res.status}`);
	return res.json();
}

export async function fetchArticles(): Promise<ArticleAPI[]> {
	const res = await fetch(`${BASE_URL}/article`);
	if (!res.ok) throw new Error(`catalog-service error: ${res.status}`);
	return res.json();
}

export async function fetchCategories(): Promise<CategoryAPI[]> {
	const res = await fetch(`${BASE_URL}/category`);
	if (!res.ok) throw new Error(`catalog-service error: ${res.status}`);
	return res.json();
}

export type NewArticleInput = {
	name: string;
	description: string;
	prix: number;
	fraisPort: number;
	categoryId: number;
	series?: string;
	year?: number;
	rarity?: string;
	grade?: string;
	imageUrl?: string;
};

/** Met une pièce en vente. Le vendeur (sellerId) est déduit de la session côté serveur. */
export async function createArticle(input: NewArticleInput): Promise<ArticleAPI> {
	const data = await apiRequest<{ article: ArticleAPI }>(BASE_URL, '/article', {
		init: { method: 'POST', body: JSON.stringify(input) },
		errorPrefix: 'catalog-service'
	});
	return data.article;
}

export type EditArticleInput = {
	name: string;
	description: string;
	prix: number;
	fraisPort: number;
	categoryId: number;
	imageUrl?: string;
};

/** Modifie une annonce existante (nom, description, prix, port, catégorie, photo). */
export async function updateArticle(id: number, input: EditArticleInput): Promise<ArticleAPI> {
	const data = await apiRequest<{ article: ArticleAPI }>(BASE_URL, `/article/${id}`, {
		init: { method: 'PUT', body: JSON.stringify(input) },
		errorPrefix: 'catalog-service'
	});
	return data.article;
}

/** Retire définitivement une annonce du catalogue. */
export async function deleteArticle(id: number): Promise<void> {
	await apiRequest(BASE_URL, `/article/${id}`, {
		init: { method: 'DELETE' },
		errorPrefix: 'catalog-service'
	});
}

/** Annonces de l'utilisateur courant (vendues incluses), pour la gestion depuis son profil. */
export async function fetchMyArticles(): Promise<ArticleAPI[]> {
	return apiRequest<ArticleAPI[]>(BASE_URL, '/me/articles', { errorPrefix: 'catalog-service' });
}

/** Tout le catalogue (vendues incluses, tous vendeurs), pour la modération back-office. */
export async function fetchAllArticlesAdmin(): Promise<ArticleAPI[]> {
	return apiRequest<ArticleAPI[]>(BASE_URL, '/admin/articles', { errorPrefix: 'catalog-service' });
}

/**
 * Envoie une photo pour une annonce existante (multipart). Ne passe pas par
 * apiRequest : celui-ci force Content-Type: application/json dès qu'un body
 * est présent, ce qui casserait le multipart (le navigateur doit fixer
 * lui-même le boundary).
 */
export async function uploadArticleImage(id: number, file: File): Promise<ArticleAPI> {
	const form = new FormData();
	form.append('image', file);
	const res = await fetch(`${BASE_URL}/article/${id}/image`, {
		method: 'POST',
		credentials: 'include',
		body: form
	});
	const data = await res.json().catch(() => ({}) as { article?: ArticleAPI; error?: string });
	if (!res.ok) {
		throw new Error(data.error ?? `catalog-service error: ${res.status}`);
	}
	return data.article as ArticleAPI;
}
