# Plan de remédiation — Analyse sécurité et recommandations

## Méthode

Analyse basée sur (1) une revue du code des flux sensibles (authentification,
paiement, messagerie, publication d'annonce), (2) l'état des scans automatisés
du pipeline (Trivy, gosec, govulncheck — 0 CRITICAL/HIGH bloquant en CI), et
(3) la confrontation du comportement réel de l'application aux exigences
explicites du contexte métier (notamment l'interdiction d'échanger des
coordonnées personnelles entre acheteur et vendeur).

**Mise à jour (10/07/2026)** : les 5 constats de l'audit initial ont été
revérifiés directement dans le code et l'infrastructure courante. **Les 5 sont
désormais couverts** — 4 par des évolutions déjà livrées au fil des sessions
précédentes, la 5e (Kyverno) par le passage `Audit → Enforce` effectué dans le
cadre de cette mise à jour. Le détail de chaque remédiation est conservé
ci-dessous pour la soutenance (constat → remédiation → preuve), plutôt que
supprimé, afin de pouvoir présenter la démarche d'audit elle-même au jury.

## Bonnes pratiques déjà en place dans le POC

Pour mémoire (le sujet demande ≥2 bonnes pratiques intégrées, en voici plus
de deux) :

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
  signature à l'admission par Kyverno (désormais bloquante, cf. ci-dessous).
- **Anti double-vente** : transaction SQL avec `UPDATE ... WHERE sold=false`
  conditionnel, empêchant deux acheteurs concurrents de réserver la même
  pièce.

## État des remédiations (constat → action → preuve)

### 1. Critique — messagerie sans filtre anti-coordonnées personnelles — **Résolu**

**Constat initial** : le contexte exige explicitement d'empêcher l'échange
d'email ou de téléphone dans la messagerie acheteur↔vendeur, pour protéger la
commission de 5 % et la garantie qualité de Collector.shop. Aucun filtre
n'existait sur `SendMessage`.

**Remédiation en place** : détection par expression régulière (email +
téléphone FR/international) avant persistance du message, message rejeté avec
erreur explicite à l'expéditeur.

**Preuve** : [`notification-service/internal/pii/filter.go`](../../apps/backend/notification-service/internal/pii/filter.go)
(fonction `Detect`), testé par
[`filter_test.go`](../../apps/backend/notification-service/internal/pii/filter_test.go).

### 2. Élevé — jeton d'authentification en `localStorage` — **Résolu**

**Constat initial** : le JWT complet était stocké en `localStorage`,
exploitable par toute injection XSS pour un vol de session total.

**Remédiation en place** : le JWT ne transite plus jamais côté JavaScript.
Seul le profil utilisateur (non sensible) est mis en cache en `localStorage`
pour un affichage optimiste ; l'authentification effective repose sur le
cookie `httpOnly` posé par `auth-service`, envoyé via `credentials: include`.

**Preuve** : [`collector-front/src/lib/stores/auth.ts`](../../apps/collector-front/src/lib/stores/auth.ts)
(commentaire explicite ligne 10 : *« Le JWT lui-même ne transite plus jamais
par du JS »*).

### 3. Moyen — publication d'annonce sans contrôle avant mise en ligne — **Résolu (contrôle manuel ; automatisation encore à faire)**

**Constat initial** : le contexte exige un contrôle avant mise en vente
publique, si possible automatisé. `CreateArticle` publiait l'annonce
immédiatement.

**Remédiation en place** : chaque annonce créée passe par défaut en statut
`pending_review` (jamais le statut envoyé par le client, anti-auto-
approbation) ; seules les annonces `approved` apparaissent dans le catalogue
public ou sont consultables par lien direct ; un admin décide via
`ApproveArticle`/`RejectArticle` (décisions tracées par métrique Prometheus).

**Limite assumée** : la décision reste **100 % manuelle** — le premier niveau
de contrôle automatisé évoqué à titre d'exemple dans le constat initial (prix
aberrant vs médiane, description trop courte, image absente, auto-approbation
si aucun signal) n'est pas implémenté. C'est un axe d'amélioration réaliste à
mentionner à l'oral plutôt qu'un point à cacher : la porte de sécurité
(rien de non modéré n'est public) est fermée, l'optimisation de charge de
modération reste ouverte.

**Preuve** : [`catalog-service/models/article.go`](../../apps/backend/catalog-service/models/article.go)
(constantes `ArticleStatusPendingReview/Approved/Rejected`),
[`catalog-service/controllers/articleController.go`](../../apps/backend/catalog-service/controllers/articleController.go)
(`CreateArticle`, `GetArticle`, `decideArticleModeration`).

### 4. Moyen — Kyverno en mode Audit, pas Enforce — **Résolu**

**Constat initial** : les 4 policies Kyverno (signature cosign, non-root,
no-latest, requests/limits) journalisaient les violations sans bloquer
l'admission — un déploiement non conforme aurait pu atteindre le cluster.

**Remédiation appliquée le 10/07/2026** : `validationFailureAction` basculé de
`Audit` à `Enforce` sur les 4 policies.

**Preuve** : [`infra/policies/verify-images.yaml`](../../infra/policies/verify-images.yaml),
[`disallow-latest-tag.yaml`](../../infra/policies/disallow-latest-tag.yaml),
[`require-resources.yaml`](../../infra/policies/require-resources.yaml),
[`require-run-as-nonroot.yaml`](../../infra/policies/require-run-as-nonroot.yaml).

**Point de vigilance avant la soutenance** : ce changement n'a pas pu être
revérifié contre les `PolicyReport` du cluster réel depuis ce poste (pas
d'accès `kubectl`/`kubeseal` configuré ici). Avant de synchroniser Argo CD sur
ce commit, lancer `kubectl get policyreport -A` sur le cluster pour confirmer
zéro violation résiduelle — sinon des pods existants non conformes (image non
signée, ressources non définies) seraient bloqués au prochain rollout.

### 5. Faible — pas de composante d'observabilité applicative — **Résolu**

**Constat initial** : `kube-prometheus-stack` était déployé au niveau
infrastructure, mais aucun service Go n'exposait de métriques applicatives.

**Remédiation en place** : `catalog-service` expose désormais `/metrics`
(compteurs requêtes HTTP par route/statut, histogramme de latence, compteurs
métier — créations d'articles, commandes, décisions de modération, uploads
d'image).

**Preuve** : [`catalog-service/metrics/metrics.go`](../../apps/backend/catalog-service/metrics/metrics.go).

## Priorisation (état au 10/07/2026)

| # | Remédiation | Sévérité | Statut |
|---|---|---|---|
| 1 | Filtre anti-coordonnées personnelles (messagerie) | Critique | ✅ Résolu |
| 2 | Cookie httpOnly pour l'authentification | Élevé | ✅ Résolu |
| 3 | Contrôle de modération avant publication d'article | Moyen | ✅ Résolu (manuel) |
| 4 | Kyverno Audit → Enforce | Moyen | ✅ Résolu (à revalider sur cluster) |
| 5 | Instrumentation Prometheus applicative | Faible | ✅ Résolu (catalog-service) |

## Axes ouverts identifiés en repassant l'audit

Ces points ne sont pas des vulnérabilités mais des limites assumées, à
mentionner spontanément à l'oral pour montrer la maîtrise du périmètre :

- **Automatisation de la modération** (cf. #3) : premier niveau de règles
  auto-approve/à-signaler non implémenté, tout passe par un admin humain.
- **Instrumentation Prometheus incomplète** : seul `catalog-service` expose
  des métriques applicatives ; `auth-service`, `notification-service` et
  `price-tracker-service` restent à instrumenter pour une observabilité
  homogène.
- **Rotation de la clé privée sealed-secrets** non encore sauvegardée hors
  cluster (voir `infra/secrets/README.md`, section Rotation).
