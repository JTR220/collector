import { defineConfig, devices } from '@playwright/test';

// Suite E2E : pilote l'app SvelteKit servie par `docker compose` (front +
// auth-service + catalog-service + ...) via un vrai navigateur. Contrairement
// aux specs vitest (src/**/*.svelte.spec.ts), ces tests ne mockent rien : ils
// valident un parcours utilisateur de bout en bout à travers la stack réelle.
//
// Prérequis local : `docker compose up -d` depuis apps/ (le service
// collector-front écoute déjà sur :5173, PUBLIC_*_API_BASE_URL pointent vers
// les autres services). Puis `npm run test:e2e` depuis apps/collector-front.
export default defineConfig({
	testDir: './e2e',
	// `vite dev` (Dockerfile.dev, HMR) transforme chaque module à la demande et
	// sert des centaines de requêtes non bundlées : plusieurs pages qui se
	// chargent en parallèle sur ce même serveur créent une ruée qui dépasse
	// largement les timeouts par défaut. Un seul worker + des timeouts plus
	// généreux éliminent cette source de flakiness (aucun impact en usage
	// normal, où les modules sont déjà transformés/cache après le 1er coup).
	fullyParallel: false,
	workers: 1,
	forbidOnly: !!process.env.CI,
	retries: process.env.CI ? 2 : 0,
	reporter: process.env.CI ? [['html', { open: 'never' }], ['list']] : 'list',
	expect: { timeout: 10000 },

	use: {
		baseURL: process.env.PLAYWRIGHT_BASE_URL ?? 'http://localhost:5173',
		trace: 'on-first-retry',
		screenshot: 'only-on-failure',
		navigationTimeout: 20000,
		actionTimeout: 10000
	},

	projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }]
});
