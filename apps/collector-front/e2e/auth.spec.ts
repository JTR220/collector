import { test, expect } from '@playwright/test';
import { dismissCookieBanner } from './helpers';

// Parcours d'authentification, contre le vrai auth-service (pas de mock) :
// identifiants invalides -> message d'erreur renvoyé par l'API ; identifiants
// valides (compte de démo seedé par auth-service) -> session cookie httpOnly
// posée, navigation débloquée (lien "Vendre" réservé aux connectés) ;
// déconnexion -> retour à l'état invité.
test.describe('Authentification', () => {
	test.beforeEach(async ({ page }) => {
		await dismissCookieBanner(page);
	});

	test('refuse un mot de passe incorrect avec le message du serveur', async ({ page }) => {
		await page.goto('/login');

		await page.getByLabel('Email').fill('test@collector.shop');
		await page.getByLabel('Mot de passe').fill('mauvais-mot-de-passe');
		await page.getByRole('button', { name: 'Se connecter' }).click();

		await expect(page.getByText('Email ou mot de passe incorrect')).toBeVisible();
		await expect(page).toHaveURL(/\/login$/);
	});

	test('connecte un compte de démo et permet de se déconnecter', async ({ page }) => {
		await page.goto('/login');

		await page.getByLabel('Email').fill('test@collector.shop');
		await page.getByLabel('Mot de passe').fill('test123');
		await page.getByRole('button', { name: 'Se connecter' }).click();

		// Redirigé hors de /login, et la nav expose "Vendre" (visible seulement connecté).
		await expect(page).toHaveURL('/');
		await expect(page.getByRole('link', { name: 'Vendre' })).toBeVisible();

		await page.getByTitle('Se déconnecter').click();

		// Le layout (main)/login n'affiche pas le header applicatif : on vérifie
		// qu'on est bien retombé sur le formulaire de connexion.
		await expect(page).toHaveURL(/\/login$/);
		await expect(page.getByRole('heading', { name: 'Accès' })).toBeVisible();
	});
});
