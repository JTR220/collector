/**
 * Helper HTTP partagé par tous les clients API (catalog, market, wishlist,
 * notifications, price-tracker) : en-têtes JSON, parsing de la réponse et
 * remontée du message d'erreur renvoyé par le backend. L'authentification
 * passe par le cookie httpOnly de session (credentials:'include'), jamais
 * par un token manipulé en JS — voir lib/stores/auth.ts.
 */
export async function apiRequest<T>(
	baseUrl: string,
	path: string,
	options: { init?: RequestInit; errorPrefix?: string } = {}
): Promise<T> {
	const { init, errorPrefix = 'API' } = options;

	const res = await fetch(`${baseUrl}${path}`, {
		...init,
		credentials: 'include',
		headers: {
			...(init?.body ? { 'Content-Type': 'application/json' } : {}),
			...init?.headers
		}
	});

	// Corps absent ou non-JSON (204, erreurs proxy...) : on retombe sur {}.
	const data = await Promise.resolve()
		.then(() => res.json())
		.catch(() => ({}));
	if (!res.ok) {
		throw new Error((data as { error?: string }).error ?? `${errorPrefix} error: ${res.status}`);
	}
	return data as T;
}
