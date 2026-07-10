# Preparation soutenance - Projet Collector

Document de revision aligne sur l'etat du depot au 09/07/2026. Il sert a
repondre aux questions de jury sur l'architecture, les choix techniques, les
fonctionnalites livrees, la securite, les tests et le deploiement GitOps.

---

## 1. Vue d'ensemble

**collector.shop** est une plateforme de marketplace pour objets de collection.
Le projet est organise en monorepo :

- `apps/collector-front/` : frontend SvelteKit 2 / Svelte 5 / TypeScript /
  Tailwind.
- `apps/backend/` : quatre microservices Go.
- `apps/docker-compose.yml` : stack locale avec PostgreSQL, RabbitMQ, MailHog
  et les services applicatifs.
- `infra/` : manifests Kubernetes, Kustomize, Argo CD, Kyverno, cert-manager,
  Sealed Secrets et monitoring.
- `docs/` : livrables d'evaluation, journal de bord et documentation
  DevSecOps.

**Question jury : pourquoi un monorepo ?**

Un seul historique Git permet de garder synchrones le front, les APIs, les
contrats evenementiels et l'infrastructure. Pour une equipe reduite, c'est plus
simple a maintenir qu'une constellation de repos separes.

---

## 2. Architecture actuelle

```text
collector/
|-- apps/
|   |-- collector-front/             # UI SvelteKit, clients API, stores
|   |-- backend/
|   |   |-- auth-service/            # comptes, JWT, roles, endpoint interne users
|   |   |-- catalog-service/         # catalogue, marketplace, achats, avis, images
|   |   |-- price-tracker-service/   # historique prix, detection fraude
|   |   `-- notification-service/    # notifications, messages, WebSocket, email
|   |-- postgres-init/               # bases locales price / notifications
|   |-- docker-compose.yml
|   `-- DOCKER.md
|-- infra/
|   |-- k8s/base/                    # deployments/services communs
|   |-- k8s/overlays/staging,prod/   # differences par environnement
|   |-- argocd/                      # App-of-Apps + applications enfants
|   |-- policies/                    # policies Kyverno en Audit
|   |-- cert-manager/
|   |-- secrets/
|   `-- traefik/
`-- docs/
    |-- evaluation-bloc/
    `-- journal/
```

Le backend est decoupe par domaine. Chaque service expose une API HTTP et, pour
les flux asynchrones, publie ou consomme des evenements RabbitMQ.

---

## 3. Services backend

### auth-service

Responsabilites :

- inscription, login, logout et profil `/me` ;
- JWT signe en HS256, role utilisateur/admin ;
- cookie `collector_token` pose au login, en plus du token utilise par le front ;
- rate-limiting en memoire sur `/login` et `/utilisateur` ;
- endpoint interne `GET /internal/users/:id`, protege par
  `X-Internal-Secret`, pour permettre au service de notification de resoudre un
  utilisateur sans dupliquer la table des comptes ;
- back-office admin : liste des comptes, suspension, reactivation.

Fichiers a citer :

- `apps/backend/auth-service/routes/routes.go`
- `apps/backend/auth-service/middlewares/auth.go`
- `apps/backend/auth-service/middlewares/ratelimit.go`
- `apps/backend/auth-service/controllers/adminController.go`

### catalog-service

Responsabilites :

- catalogue public : articles, categories, details, note vendeur ;
- gestion vendeur : creation, edition, suppression, `GET /me/articles` ;
- upload photo securise via `POST /article/:id/image` ;
- marketplace : achat direct, validation ou refus vendeur, commandes acheteur
  et ventes vendeur ;
- avis vendeur apres commande payee/livree ;
- wishlist ;
- back-office : statistiques et moderation catalogue ;
- publication RabbitMQ : `price.updated`, `order.created`, `order.decided`.

Points importants :

- les articles vendus sont exclus du catalogue public ;
- l'achat cree une commande `pending` et reserve l'article ;
- le vendeur accepte ou refuse ensuite la commande ;
- la reservation utilise une logique transactionnelle pour eviter la double
  vente ;
- les IDs `uint` du catalogue sont convertis en UUID deterministes pour les
  services consommateurs.

Fichiers a citer :

- `apps/backend/catalog-service/routes/routes.go`
- `apps/backend/catalog-service/controllers/marketplaceController.go`
- `apps/backend/catalog-service/controllers/uploadController.go`
- `apps/backend/catalog-service/controllers/reviewController.go`
- `apps/backend/catalog-service/events/publisher.go`

### price-tracker-service

Responsabilites :

- consommer `price.updated` depuis RabbitMQ ;
- historiser les prix ;
- detecter les anomalies : spike, flood pricing, dumping ;
- publier `fraud.alert` ;
- exposer l'historique de prix et la resolution d'alertes.

Endpoints :

- `GET /api/v1/health`
- `GET /api/v1/items/:id/price-history`
- `GET /api/v1/alerts`
- `PUT /api/v1/alerts/:id/resolve` avec role admin

### notification-service

Responsabilites :

- consommer `price.updated`, `fraud.alert`, `order.created`,
  `order.decided` ;
- persister les notifications ;
- pousser les notifications en temps reel via WebSocket `/ws?token=...` ;
- envoyer les emails via SMTP/MailHog en local ;
- gerer la messagerie directe acheteur/vendeur ;
- pousser aussi les nouveaux messages en WebSocket.

Endpoints principaux :

- `GET /api/v1/notifications`
- `PUT /api/v1/notifications/:id/read`
- `PUT /api/v1/notifications/read-all`
- `GET /api/v1/notifications/unread-count`
- `POST /api/v1/messages`
- `GET /api/v1/conversations`
- `GET /api/v1/conversations/:id/messages`
- `PUT /api/v1/conversations/:id/read`

Limite connue importante : la messagerie limite la taille et interdit l'envoi a
soi-meme, mais ne filtre pas encore les emails ou numeros de telephone dans le
corps du message. C'est le risque P0 du plan de remediation.

---

## 4. Flux evenementiels a expliquer

### Flux prix et fraude

```text
catalog-service -- price.updated --> RabbitMQ collector.events
        |                                      |
        |                                      +--> notification-service
        |
        +--> price-tracker-service -- fraud.alert --> RabbitMQ collector.alerts
                                                  |
                                                  +--> notification-service
                                                          |
                                                          +--> WebSocket + email
```

Interet : le catalogue ne depend pas directement du price-tracker ni des
notifications. Chaque service peut etre teste, deploye et scale separement.

### Flux achat avec validation vendeur

```text
Acheteur -> POST /article/:id/buy -> catalog-service
catalog-service -> Order pending + Article sold=true
catalog-service -> order.created -> notification-service -> vendeur
Vendeur -> PATCH /order/:id/accept ou /reject -> catalog-service
catalog-service -> order.decided -> notification-service -> acheteur
```

Ce flux repond au besoin metier : le vendeur garde la main avant finalisation de
la transaction.

---

## 5. Frontend SvelteKit

Technos :

- SvelteKit `^2.69.1`, Svelte `^5.56.4`, Vite `^7.2.6` ;
- TypeScript `^6.0.3` ;
- Tailwind CSS 4 ;
- Vitest pour les tests unitaires.

Organisation :

- `src/routes/(main)/` : dashboard et administration ;
- `src/routes/(holo)/` : experience principale marketplace : catalogue,
  profil, messages, vente, etc. Le nom entre parentheses est un route group et
  n'apparait pas dans l'URL ;
- `src/lib/api/` : clients API types par domaine ;
- `src/lib/stores/` : session, panier, notifications, messages, recents ;
- `src/lib/types/` : types partages.

Fonctionnalites visibles :

- catalogue public et detail article ;
- creation/edition d'annonce, upload ou URL d'image ;
- profil vendeur/acheteur avec annonces, commandes, ventes et avis ;
- wishlist ;
- negotiation de prix par message ;
- notifications temps reel ;
- espace admin.

Question jury : **comment le front recoit-il les notifications ?**

Le front ouvre une connexion WebSocket vers `notification-service` avec un JWT.
Les messages recus mettent a jour les stores Svelte, ce qui rafraichit l'UI sans
rechargement de page.

---

## 6. Securite

Mesures deja en place :

- Go `1.26.5` sur les quatre services, suite a la remediation de l'alerte
  GO-2026-5856 ;
- JWT obligatoire : les services echouent ou refusent les requetes si
  `JWT_SECRET` est absent ;
- rate-limiting sur les endpoints d'authentification ;
- endpoint interne auth protege par secret partage ;
- CORS borne par `FRONTEND_ORIGIN` ;
- WebSocket protege contre le Cross-Site WebSocket Hijacking via controle
  d'`Origin` ;
- upload image durci : type detecte par les octets, formats autorises,
  limite 5 Mo, nom genere serveur, stockage statique, `nosniff` ;
- secrets Kubernetes via Sealed Secrets ;
- images distroless non-root, scan Trivy, SBOM Syft, signature cosign ;
- policies Kyverno de gouvernance cluster.

Risques suivis :

- P0 : filtre anti-coordonnees personnelles dans la messagerie ;
- P1 : remplacer l'usage front du JWT en `localStorage` par le cookie httpOnly
  pour tous les services ;
- P1 : moderation avant publication d'annonce ;
- P2 : passer Kyverno de `Audit` a `Enforce` ;
- P2 : ajouter des metriques applicatives Prometheus.

---

## 7. Tests et qualite

Backend :

- tests unitaires sur controllers, middlewares, detector, hub, mailer,
  authclient, repository ;
- tests d'acceptation HTTP sur le vrai routeur Gin ;
- tests d'integration repository pour notification-service ;
- tests de concurrence et de comportement pour la reservation et les alertes.

Frontend :

- tests Vitest sur clients API et stores ;
- `npm run check` pour typecheck Svelte/TypeScript ;
- `npm run lint` pour Prettier + ESLint.

CI/CD :

- workflows backend, frontend et gitleaks ;
- `go test -race -covermode=atomic` ;
- govulncheck, gosec, Trivy ;
- SonarCloud pour dette technique/couverture ;
- build Docker, scan, signature et deploiement GitOps.

---

## 8. Deploiement et exploitation

### Local

Depuis `apps/` :

```powershell
docker compose up --build
```

URLs locales :

- Front : `http://localhost:5173`
- Auth : `http://localhost:8080`
- Catalogue : `http://localhost:8081`
- Price tracker : `http://localhost:8082`
- Notifications : `http://localhost:8083`
- RabbitMQ UI : `http://localhost:15672`
- MailHog : selon la configuration Docker Compose

### Kubernetes / GitOps

- `infra/k8s/base` contient la configuration commune ;
- `infra/k8s/overlays/staging` et `infra/k8s/overlays/prod` portent les
  differences d'environnement ;
- `infra/argocd/bootstrap/root-app.yaml` applique le pattern App-of-Apps ;
- staging est synchronise automatiquement, prod reste en synchro manuelle ;
- PostgreSQL et RabbitMQ sont deployes dans la base Kustomize ;
- `catalog-service` a un PVC `catalog-uploads` pour les images vendeurs, donc
  `replicas: 1` tant que le volume est `ReadWriteOnce`.

Commandes utiles :

```bash
kubectl get pods -n <namespace>
kubectl logs -f deploy/<service> -n <namespace>
kubectl describe pod <pod> -n <namespace>
kubectl kustomize infra/k8s/overlays/staging
argocd app list
argocd app get collector-staging
argocd app diff collector-staging
kubectl get certificate -n <namespace>
kubectl get sealedsecrets -n <namespace>
```

### Montee en charge (Siege)

Voir [`loadtest/README.md`](loadtest/README.md) pour le protocole complet et
les resultats. **Etat au 10/07/2026** : `/api/health` tient 100% de
disponibilite jusqu'a 100 utilisateurs concurrents sur staging, mais
`/api/article` et `/api/category` renvoient **500** ("Impossible de
recuperer les articles/categories") sur `collector-staging` actuellement —
probablement un decalage entre le code deploye (filtre de moderation
`pending_review`) et le schema/donnees reels de la base staging. **A
corriger avant la soutenance** : rejouer le test complet (`urls.txt.tpl`,
tous les endpoints) une fois le 500 resolu, pour avoir un vrai chiffre de
montee en charge sur le catalogue metier et pas seulement sur `/health`.

---

## 9. Reponses courtes a preparer

**Pourquoi Go ?**

Services simples a deployer, binaires statiques, concurrence native utile pour
WebSocket/RabbitMQ, ecosysteme solide pour HTTP, JWT, SQL et tests.

**Pourquoi RabbitMQ ?**

Il decouple les domaines : une mise a jour de prix peut alimenter l'historique,
la detection fraude et les notifications sans bloquer la requete catalogue.

**Pourquoi Kustomize ?**

Le `base` evite la duplication, les `overlays` isolent les differences staging
et prod : domaines, secrets, certificats, politiques de synchro.

**Pourquoi Argo CD ?**

Le cluster converge vers l'etat de Git. Une modification d'infra passe par une
revue, un commit et une synchronisation tracable.

**Limite la plus importante du POC ?**

La messagerie ne bloque pas encore les coordonnees personnelles. C'est le point
de remediation prioritaire car il touche directement le modele economique et
l'exigence metier.

---

## 10. Documents a citer

- `README.md` : vue d'ensemble et lancement local.
- `ARCHITECTURE.md` : architecture applicative.
- `docs/DEVSECOPS.md` : pipeline CI/CD, GitOps et securite.
- `docs/evaluation-bloc/01-protocole-experimentation.md`
- `docs/evaluation-bloc/02-backlog-fonctionnalite-metier.md`
- `docs/evaluation-bloc/03-indicateurs-qualite.md`
- `docs/evaluation-bloc/04-cartographie-competences-formation.md`
- `docs/evaluation-bloc/05-plan-remediation-securite.md`
- `docs/journal/2026-07-09.md` : changements les plus recents.
