import { describe, it, expect, beforeEach, vi } from 'vitest';

import {
	fetchArticle,
	fetchArticles,
	fetchCategories,
	createArticle,
	articleImage
} from './catalog';

describe('catalog API', () => {
	const fetchMock = vi.fn();

	beforeEach(() => {
		fetchMock.mockReset();
		vi.stubGlobal('fetch', fetchMock);
	});

	it('fetchArticle appelle /article/:id et renvoie le JSON brut', async () => {
		const article = { ID: 1, name: 'Carte rare' };
		fetchMock.mockResolvedValue({ ok: true, json: async () => article });

		const result = await fetchArticle(1);

		expect(fetchMock.mock.calls[0][0]).toContain('/article/1');
		expect(result).toEqual(article);
	});

	it('fetchArticle rejette sur reponse non-ok', async () => {
		fetchMock.mockResolvedValue({ ok: false, status: 404 });
		await expect(fetchArticle(999)).rejects.toThrow(/404/);
	});

	it('fetchArticles appelle /article sans id', async () => {
		fetchMock.mockResolvedValue({ ok: true, json: async () => [] });
		await fetchArticles();
		expect(fetchMock.mock.calls[0][0]).toMatch(/\/article$/);
	});

	it('fetchCategories appelle /category', async () => {
		fetchMock.mockResolvedValue({ ok: true, json: async () => [] });
		await fetchCategories();
		expect(fetchMock.mock.calls[0][0]).toMatch(/\/category$/);
	});

	it('createArticle envoie le token Bearer et extrait .article de la reponse', async () => {
		const article = { ID: 2, name: 'Nouvelle carte' };
		fetchMock.mockResolvedValue({
			ok: true,
			json: async () => ({ status: 'created', article, message: 'ok' })
		});

		const result = await createArticle('tok-abc', {
			name: 'Nouvelle carte',
			description: '',
			prix: 10,
			fraisPort: 2,
			categoryId: 1
		});

		const [url, init] = fetchMock.mock.calls[0];
		expect(url).toContain('/article');
		expect((init.headers as Record<string, string>).Authorization).toBe('Bearer tok-abc');
		expect(result).toEqual(article);
	});

	it("createArticle rejette avec le message d'erreur du backend", async () => {
		fetchMock.mockResolvedValue({
			ok: false,
			status: 400,
			json: async () => ({ error: 'Donnees invalides' })
		});

		await expect(
			createArticle('tok', { name: '', description: '', prix: 0, fraisPort: 0, categoryId: 1 })
		).rejects.toThrow('Donnees invalides');
	});

	it('articleImage prefixe une URL relative avec BASE_URL', () => {
		const url = articleImage({ imageUrl: '/uploads/foo.jpg' });
		expect(url).toContain('/uploads/foo.jpg');
	});

	it('articleImage laisse une URL absolue inchangee', () => {
		expect(articleImage({ imageUrl: 'https://cdn.example.com/foo.jpg' })).toBe(
			'https://cdn.example.com/foo.jpg'
		);
	});

	it('articleImage renvoie null sans image', () => {
		expect(articleImage({ imageUrl: '' })).toBeNull();
	});
});
