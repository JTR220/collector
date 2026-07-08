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

/** Résout l'URL d'affichage de la photo d'un article (URL absolue, ou chemin relatif hérité résolu sur le catalog-service). */
export function articleImage(article: Pick<ArticleAPI, 'imageUrl'>): string | null {
	if (!article.imageUrl) return null;
	return article.imageUrl.startsWith('http') ? article.imageUrl : `${BASE_URL}${article.imageUrl}`;
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

/** Met une pièce en vente. Le vendeur (sellerId) est déduit du token côté serveur. */
export async function createArticle(token: string, input: NewArticleInput): Promise<ArticleAPI> {
	const data = await apiRequest<{ article: ArticleAPI }>(BASE_URL, '/article', {
		token,
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
export async function updateArticle(
	token: string,
	id: number,
	input: EditArticleInput
): Promise<ArticleAPI> {
	const data = await apiRequest<{ article: ArticleAPI }>(BASE_URL, `/article/${id}`, {
		token,
		init: { method: 'PUT', body: JSON.stringify(input) },
		errorPrefix: 'catalog-service'
	});
	return data.article;
}

/** Retire définitivement une annonce du catalogue. */
export async function deleteArticle(token: string, id: number): Promise<void> {
	await apiRequest(BASE_URL, `/article/${id}`, {
		token,
		init: { method: 'DELETE' },
		errorPrefix: 'catalog-service'
	});
}

/** Annonces de l'utilisateur courant (vendues incluses), pour la gestion depuis son profil. */
export async function fetchMyArticles(token: string): Promise<ArticleAPI[]> {
	return apiRequest<ArticleAPI[]>(BASE_URL, '/me/articles', {
		token,
		errorPrefix: 'catalog-service'
	});
}
