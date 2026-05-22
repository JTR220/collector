// Les données du catalogue viennent désormais du catalog-service via $lib/api/catalog.ts
// Ce fichier est conservé pour compatibilité historique uniquement.
export type Collectible = Record<string, never>;
export const COLLECTIBLES: Collectible[] = [];
