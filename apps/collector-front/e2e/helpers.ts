import { expect, type Page } from '@playwright/test';

/**
 * Le bandeau RGPD (GCookieConsent.svelte) recouvre le bas de page tant qu'il
 * n'est pas acquitté et intercepte les clics sur les boutons qu'il chevauche
 * (ex. "Se connecter"). On pré-remplit le localStorage avant la première
 * navigation plutôt que de le fermer à chaque test : plus robuste et ça
 * garde les specs concentrées sur le parcours testé, pas sur le consentement.
 */
export async function dismissCookieBanner(page: Page): Promise<void> {
	await page.addInitScript(() => {
		window.localStorage.setItem('collector_cookie_consent', 'true');
	});
}

/**
 * Toutes les pages du groupe (holo) (marché, lot, panier...) sont gardées par
 * `(holo)/+page.ts` : sans `collector_user` en localStorage, on est renvoyé
 * vers /login avant même que la page ne s'affiche. On passe donc par le vrai
 * formulaire de connexion (hit réel sur auth-service) avant de tester le
 * catalogue ou le panier.
 */
export async function loginViaUi(page: Page, email: string, password: string): Promise<void> {
	await page.goto('/login');
	await page.getByLabel('Email').fill(email);
	await page.getByLabel('Mot de passe').fill(password);
	await page.getByRole('button', { name: 'Se connecter' }).click();
	await expect(page).toHaveURL('/', { timeout: 10000 });
}

/**
 * Le catalogue est seedé par catalog-service et peut évoluer (nouvelles
 * pièces, ventes déjà backfillées) : plutôt que de coder en dur l'ID d'un
 * article, on parcourt les cartes de la vitrine et on retient la première
 * dont la fiche lot expose bien un bouton "Ajouter au panier" actif (donc ni
 * vendue, ni propriété de l'utilisateur courant).
 */
export async function findAvailableArticleId(page: Page): Promise<string> {
	await page.goto('/');
	const cards = page.locator('.grid-section a.card');
	await expect(cards.first()).toBeVisible(); // attend la fin du fetchArticles() async

	// On récupère tous les hrefs avant de naviguer : une fois qu'on quitte '/'
	// pour une fiche lot, le locator `cards` ci-dessus n'a plus rien à cibler.
	const hrefs = await cards.evaluateAll((els) =>
		els.map((el) => el.getAttribute('href')).filter((h): h is string => !!h)
	);

	for (const href of hrefs) {
		const id = href.replace('/lot/', '');

		await page.goto(`/lot/${id}`);
		const addButton = page.getByRole('button', { name: 'Ajouter au panier' });
		// isVisible() renvoie l'état immédiat sans attendre : la fiche lot rend
		// sa CTA après un fetchArticle() async, donc on attend explicitement
		// (borné) au lieu de conclure "indisponible" sur un simple instantané.
		const appeared = await addButton
			.waitFor({ state: 'visible', timeout: 3000 })
			.then(() => true)
			.catch(() => false);
		if (appeared) {
			return id;
		}
	}

	throw new Error('Aucun article disponible trouvé dans le catalogue seedé pour le test e2e.');
}
