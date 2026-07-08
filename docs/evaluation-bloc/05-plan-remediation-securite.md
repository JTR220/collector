# Plan de remédiation — Analyse sécurité et recommandations

## Méthode

Analyse basée sur (1) une revue du code des flux sensibles (authentification,
paiement, messagerie), (2) l'état des scans automatisés du pipeline (Trivy,
gosec, govulncheck — 0 CRITICAL/HIGH bloquant en CI au moment de la
rédaction), et (3) la confrontation du comportement réel de l'application aux
exigences explicites du contexte métier (notamment l'interdiction d'échanger
des coordonnées personnelles entre acheteur et vendeur).

## Bonnes pratiques déjà en place dans le POC

Pour mémoire (le sujet demande ≥2 bonnes pratiques intégrées, en voici plus
de deux, déjà en production dans le code) :

- **JWT fail-fast** : les 4 services refusent de démarrer sans `JWT_SECRET`
  explicite (aucun secret par défaut compilé en dur), et rejettent toute
  requête si le secret est vide au runtime (503).
- **Anti brute-force** : rate-limiting en mémoire (10 req/min/IP) sur
  `/login` et `/utilisateur` (`auth-service/middlewares/ratelimit.go`).
- **HTTPS** de bout en bout via cert-manager + Ingress (voir
  `01-protocole-experimentation.md`).
- **Anti mass-assignment** : les entrées HTTP passent par des DTO dédiés
  (`dto.RegisterInput`), jamais un bind direct sur le modèle de persistance ;
  l'identité vendeur (`sellerId`) est toujours dérivée du token JWT, jamais
  acceptée depuis le corps de la requête client.
- **Communication interservice authentifiée** : l'endpoint interne
  `GET /internal/users/:id` (résolution d'email pour les notifications) est
  protégé par un secret partagé `X-Internal-Secret`, refusé par défaut si le
  secret n'est pas configuré.
- **Chaîne de confiance des images** : distroless non-root, scan Trivy
  bloquant avant push, signature cosign keyless, SBOM Syft, vérification de
  signature à l'admission par Kyverno.
- **Anti double-vente** : transaction SQL avec `UPDATE ... WHERE sold=false`
  conditionnel, empêchant deux acheteurs concurrents de réserver la même
  pièce.

## Vulnérabilités / risques identifiés

### Critique — messagerie sans filtre anti-coordonnées personnelles

Le contexte est explicite : *« Le système doit donc empêcher l'échange
d'informations personnelles telles que l'email, un numéro de téléphone »*.
La messagerie directe acheteur↔vendeur actuellement implémentée
(`notification-service`, table `messages`) ne filtre **aucun** contenu — un
utilisateur peut transmettre librement un email ou un numéro de téléphone
pour organiser un paiement hors plateforme, contournant la commission de 5 %
et la garantie qualité assurée par Collector.shop.

**Impact** : contournement direct du modèle économique (perte de commission)
et de la garantie transactionnelle mise en avant par la direction.

**Remédiation proposée** : validation du corps du message côté
`notification-service` (`SendMessage`) avant persistance — détection par
expression régulière des motifs email et numéro de téléphone (formats FR/
internationaux courants), rejet ou masquage du message avec message
d'erreur explicite à l'expéditeur. Complément possible : détection de
motifs contournés (espacement volontaire, « at » écrit en toutes lettres) en
V2, hors périmètre du prototype.

### Élevé — jeton d'authentification en `localStorage`

Le front stocke le JWT en `localStorage` (`src/lib/stores/auth.ts`),
accessible à tout script exécuté dans la page — une faille XSS, même
mineure, entraînerait un vol de session complet. Un cookie `httpOnly` est
déjà utilisé en parallèle côté `auth-service` (`collector_token`) pour la
session navigateur, mais le front continue d'utiliser le token JavaScript
pour les appels `Authorization: Bearer` vers les autres services.

**Remédiation proposée** : faire porter l'authentification inter-services
par le cookie `httpOnly` existant (déjà posé au login) plutôt que par un
`Authorization` header lu depuis `localStorage` — nécessite que chaque
service backend accepte le cookie en plus du header (patron déjà présent
côté `auth-service`, `TokenFromRequest`), et une configuration CORS
`credentials: include` cohérente entre tous les services.

### Moyen — publication d'annonce sans contrôle avant mise en ligne

Le contexte exige : *« La mise en ligne d'un article est proposé à la vente
qu'après contrôle de Collector. Le contrôle doit pouvoir être automatisé le
plus possible. »* `CreateArticle` publie aujourd'hui l'annonce immédiatement,
sans état intermédiaire de modération.

**Remédiation proposée** : ajouter un statut `Article.ModerationStatus`
(`pending_review` par défaut → `approved`/`rejected`), avec un contrôle
automatisé de premier niveau (prix aberrant vs médiane de la catégorie,
description trop courte, image absente) réutilisant la logique déjà en place
dans `price-tracker-service` pour la détection d'anomalies de prix ; passage
en `approved` automatique si aucun signal, sinon file de modération admin.

### Moyen — Kyverno en mode Audit, pas Enforce

Les politiques Kyverno (vérification de signature cosign, interdiction de
`:latest`, non-root obligatoire, limites de ressources) sont déployées mais
en mode `Audit` (`docs/DEVSECOPS.md`) : elles journalisent les violations
sans les bloquer. Un déploiement non conforme pourrait donc atteindre le
cluster sans être arrêté.

**Remédiation proposée** : purger les `PolicyReport` existants
(`kubectl get policyreport -A`), confirmer zéro violation résiduelle, puis
basculer `validationFailureAction: Enforce` — déjà identifié comme prochain
chantier dans `docs/DEVSECOPS.md`.

### Faible — pas de composante d'observabilité applicative

Le stack `kube-prometheus-stack` est déployé au niveau infrastructure, mais
aucun service Go n'expose de métriques applicatives (`/metrics`). En cas
d'incident (ex. pic de commandes refusées, latence de la messagerie), il n'y
a aujourd'hui aucune donnée applicative pour diagnostiquer rapidement.

**Remédiation proposée** : instrumenter au minimum `catalog-service` et
`notification-service` avec le client Prometheus Go standard (compteur de
requêtes par statut, histogramme de latence), pré-requis déjà noté dans
`docs/DEVSECOPS.md` pour l'analyse canary automatique.

## Priorisation

| # | Remédiation | Sévérité | Effort estimé | Priorité |
|---|---|---|---|---|
| 1 | Filtre anti-coordonnées personnelles (messagerie) | Critique | Faible (regex + test) | **P0** |
| 2 | Cookie httpOnly pour tous les appels inter-services | Élevé | Moyen (tous les clients front) | P1 |
| 3 | Contrôle de modération avant publication d'article | Moyen | Moyen | P1 |
| 4 | Kyverno Audit → Enforce | Moyen | Faible (une fois purgé) | P2 |
| 5 | Instrumentation Prometheus applicative | Faible (mais bloque la démo de charge outillée) | Moyen | P2 |

La priorisation reflète à la fois la sévérité métier (P0 = contournement
direct du modèle économique et d'une exigence contractuelle explicite du
contexte) et le rapport effort/risque (les remédiations P1/P2 sont
importantes mais n'exposent pas à un contournement immédiat de l'exigence
métier).
