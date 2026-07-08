/**
 * Helper HTTP partagé par tous les clients API (catalog, market, wishlist,
 * notifications, price-tracker) : en-têtes JSON + Bearer, parsing de la
 * réponse et remontée du message d'erreur renvoyé par le backend.
 */
export async function apiRequest<T>(
	baseUrl: string,
	path: string,
	options: { token?: string; init?: RequestInit; errorPrefix?: string } = {}
): Promise<T> {
	const { token, init, errorPrefix = 'API' } = options;

	const res = await fetch(`${baseUrl}${path}`, {
		...init,
		headers: {
			...(init?.body ? { 'Content-Type': 'application/json' } : {}),
			...(token ? { Authorization: `Bearer ${token}` } : {}),
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
