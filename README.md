# collector.shop

`collector.shop` est un monorepo organise autour d'un front web et de microservices backend evenementiels.

L'objectif du projet est de construire une plateforme orientee collection : catalogue d'articles, categories, comptes utilisateurs, marketplace, suivi de prix et notifications temps reel, le tout via une interface web unifiee.

## Structure

```text
collector/
├── .github/
│   └── workflows/         # backend.yml, frontend.yml, gitleaks.yml
├── apps/
│   ├── backend/
│   │   ├── auth-service/            # identite + JWT
│   │   ├── catalog-service/         # catalogue, marketplace, gamification, publisher d'events
│   │   ├── price-tracker-service/   # suivi de prix + detection de fraude
│   │   └── notification-service/    # notifications temps reel (WebSocket)
│   ├── collector-front/   # SvelteKit + TypeScript + Tailwind
│   ├── postgres-init/     # creation des bases price / notifications
│   ├── docker-compose.yml
│   └── DOCKER.md
├── infra/                 # K8s (base + overlays), ArgoCD, Kyverno, secrets
├── ARCHITECTURE.md
├── PROJECT_OVERVIEW.md
└── ROADMAP.md
```

## Composants

### Front

- `apps/collector-front` — SvelteKit + TypeScript + Tailwind
- interface catalogue, marketplace et gamification branchee sur les APIs
- cloche de notifications temps reel (WebSocket) et historique de prix

### Backend

- `apps/backend/auth-service` (8080) — Go + Gin + Gorm. Login JWT (HS256), comptes utilisateurs, `/health`.
- `apps/backend/catalog-service` (8081) — Go + Gin + Gorm. CRUD articles/categories, marketplace, gamification. Publie `price.updated` sur RabbitMQ quand un prix change.
- `apps/backend/price-tracker-service` (8082) — Go. Consomme `price.updated`, historise les prix, detecte la fraude (spike, flood, dumping) et publie `fraud.alert`.
- `apps/backend/notification-service` (8083) — Go. Consomme `price.updated` et `fraud.alert`, persiste les notifications et les pousse en WebSocket (`/ws?token=...`).

### Flux evenementiel (RabbitMQ)

```text
catalog-service ──price.updated──▶ collector.events ──▶ price-tracker-service
                                          │                      │
                                          └──▶ notification ◀──fraud.alert── collector.alerts
                                                   │
                                                   └──WebSocket──▶ collector-front
```

Les IDs `uint` du catalogue sont convertis en UUID deterministe (`00000000-0000-0000-0000-<hex>`) pour les services consommateurs. Ce mapping est partage entre le Go (`catalog-service/events/event.go`) et le front (`src/lib/utils/eventId.ts`).

### Infra locale

- PostgreSQL (bases `collector`, `collector_price`, `collector_notifications`)
- RabbitMQ (3.13-management)
- Docker Desktop via `apps/docker-compose.yml`

### Infra cible (GitOps)

- Kubernetes (`infra/k8s/base` + overlays `staging`/`prod`)
- ArgoCD (`infra/argocd`) : staging en sync auto, prod en sync manuel
- Politiques Kyverno (images signees, ressources, non-root)
- Secrets via Sealed Secrets (`infra/secrets`)

## Lancement local avec Docker

Depuis `C:\Users\2577407\Documents\collector\apps` :

```powershell
docker compose up --build
```

URLs locales :

- Front : `http://localhost:5173`
- API Auth : `http://localhost:8080` (health `/health`)
- API Catalogue : `http://localhost:8081` (health `/health`)
- API Price Tracker : `http://localhost:8082` (health `/api/v1/health`)
- API Notifications : `http://localhost:8083` (health `/api/v1/health`, WebSocket `/ws`)
- PostgreSQL : `localhost:5432`
- RabbitMQ UI : `http://localhost:15672` (guest/guest)

Variables front utiles : `VITE_AUTH_API_BASE_URL`, `VITE_CATALOG_API_BASE_URL`, `VITE_PRICE_TRACKER_API_BASE_URL`, `VITE_NOTIFICATION_API_BASE_URL`.

## Documentation

- Vue d'ensemble projet : [PROJECT_OVERVIEW.md](PROJECT_OVERVIEW.md)
- Architecture : [ARCHITECTURE.md](ARCHITECTURE.md)
- Roadmap : [ROADMAP.md](ROADMAP.md)
- Docker local : [apps/DOCKER.md](apps/DOCKER.md)

## Resume

`collector.shop` est un monorepo microservices evenementiel : un front SvelteKit unifie, quatre services Go, une chaine RabbitMQ (catalogue → price-tracker → notifications), une stack Docker locale complete et une infra GitOps (K8s/ArgoCD/Kyverno).
