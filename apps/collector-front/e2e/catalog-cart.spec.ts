import { test, expect, type Page } from '@playwright/test';
import { dismissCookieBanner, findAvailableArticleId, loginViaUi } from './helpers';

// Parcours marché -> panier, contre le vrai catalog-service : la recherche
// filtre la grille côté client sur des données réellement chargées via
// fetchArticles(), puis l'ajout au panier (localStorage) et le récapitulatif
// /panier retombent sur le même article et le même total. Le groupe de routes
// (holo) exige une session (garde dans (holo)/+page.ts) : on se connecte donc
// avec le compte de démo "Acheteur Demo", qui n'est propriétaire d'aucune
// pièce du catalogue seedé (cf catalog-service SeedData / backfillDemoOrders).
//
// Un seul login partagé pour tout le fichier (mode serial + page commune) :
// /login est protégé par un rate-limit anti brute-force côté auth-service
// (10 req/min/IP, cf middlewares/ratelimit.go) que la suite complète ne doit
// pas approcher.
test.describe('Catalogue et panier', () => {
	test.describe.configure({ mode: 'serial' });

	let page: Page;

	test.beforeAll(async ({ browser }) => {
		page = await browser.newPage();
		await dismissCookieBanner(page);
		await loginViaUi(page, 'acheteur@collector.shop', 'acheteur123');
	});

	test.afterAll(async () => {
		await page.close();
	});

	test('la vitrine charge le catalogue et la recherche filtre les résultats', async () => {
		await page.goto('/');

		const cards = page.locator('.grid-section a.card');
		await expect(cards.first()).toBeVisible();
		const total = await cards.count();
		expect(total).toBeGreaterThan(0);

		const firstName = (await cards.first().locator('.card-name').textContent())?.trim() ?? '';
		expect(firstName.length).toBeGreaterThan(0);

		await page.getByPlaceholder('Rechercher une pièce, une série, une référence…').fill(firstName);

		await expect(cards.first().locator('.card-name')).toHaveText(firstName);
	});

	test('ajoute un lot au panier et retrouve le bon total en récapitulatif', async () => {
		const articleId = await findAvailableArticleId(page);

		await page.goto(`/lot/${articleId}`);
		const articleName = (await page.locator('.lot-title').textContent())?.trim() ?? '';
		const priceText = (await page.locator('.lot-price').textContent())?.trim() ?? '';

		await page.getByRole('button', { name: 'Ajouter au panier' }).click();
		await expect(page.getByRole('button', { name: '✓ Dans le panier' })).toBeVisible();

		await page.goto('/panier');

		await expect(page.getByRole('link', { name: articleName })).toBeVisible();
		// Le sous-total (un seul article) est formaté par la même fonction eur()
		// que le prix affiché sur la fiche lot : les deux chaînes doivent matcher.
		await expect(page.locator('.summary-row').filter({ hasText: 'Sous-total' })).toContainText(
			priceText
		);

		await page.getByRole('button', { name: 'Passer à la commande' }).click();
		await expect(page).toHaveURL('/paiement');
		await expect(page.getByRole('heading', { name: 'Résumé de commande' })).toBeVisible();
		await expect(page.getByText(articleName, { exact: true })).toBeVisible();
	});
});
