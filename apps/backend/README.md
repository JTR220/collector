# Collector — price-tracker-service & notification-service

## Architecture du flux event-driven

```
catalog-service
      │ PATCH /items/:id/price
      │ publie → price.updated
      ▼
[RabbitMQ — collector.events]
      │
      ├──► price-tracker-service :8082
      │         │ analyse règles anti-fraude
      │         │ publie → fraud.alert
      │         ▼
      │    [RabbitMQ — collector.alerts]
      │         │
      └──►      ▼
           notification-service :8083
                │ consomme price.updated + fraud.alert
                │ WebSocket push
                ▼
           🖥️ Client navigateur (temps réel)
```

## Démarrage rapide

```bash
# Depuis le dossier racine (là où se trouve docker-compose.yml)
docker compose up --build
```

Services disponibles :
| Service                | URL                                   |
|------------------------|---------------------------------------|
| price-tracker API      | http://localhost:8082/api/v1          |
| notification-service   | http://localhost:8083/api/v1          |
| WebSocket              | ws://localhost:8083/ws?token=<jwt>    |
| RabbitMQ Management    | http://localhost:15672 (guest/guest)  |

## Scénario de démo (10 min)

### 1. Ouvrir le client WebSocket

Dans un navigateur, ouvre la console et connecte-toi :
```javascript
// Token JWT de test (valide avec JWT_SECRET=change-me-in-production)
// sub = seller UUID, exp = 2099
const token = "VOTRE_JWT_ICI"
const ws = new WebSocket(`ws://localhost:8083/ws?token=${token}`)
ws.onmessage = (e) => console.log("📬 NOTIF REÇUE:", JSON.parse(e.data))
ws.onopen = () => console.log("✅ WebSocket connecté")
```

### 2. Simuler une hausse de prix (SPIKE +80%)

```bash
# Publie directement un event price.updated sur RabbitMQ
# (simule ce que catalog-service ferait)
curl -X POST http://localhost:15672/api/exchanges/%2F/collector.events/publish \
  -u guest:guest \
  -H "Content-Type: application/json" \
  -d '{
    "properties": {"content_type": "application/json", "delivery_mode": 2},
    "routing_key": "price.updated",
    "payload": "{\"item_id\":\"11111111-1111-1111-1111-111111111111\",\"seller_id\":\"22222222-2222-2222-2222-222222222222\",\"old_price\":100.00,\"new_price\":180.00,\"updated_at\":\"2025-03-05T10:00:00Z\"}",
    "payload_encoding": "string"
  }'
```

### 3. Observer

- **Dans le navigateur** : notification `PRICE_SPIKE` apparaît en temps réel
- **Dans les logs** : `SUSPICIOUS_SPIKE détecté`
- **RabbitMQ UI** : messages publiés sur `collector.alerts`
- **Navigateur** : notification `FRAUD_ALERT` arrive juste après

### 4. Consulter l'historique

```bash
# Historique des prix
curl http://localhost:8082/api/v1/items/11111111-1111-1111-1111-111111111111/price-history

# Alertes fraude
curl http://localhost:8082/api/v1/alerts

# Notifs en base
curl http://localhost:8083/api/v1/notifications \
  -H "Authorization: Bearer VOTRE_JWT_ICI"
```

## Règles de détection anti-fraude

| Règle             | Déclencheur                             | Configurable via ENV         |
|-------------------|-----------------------------------------|------------------------------|
| SUSPICIOUS_SPIKE  | Hausse > 50% en moins de 24h            | SPIKE_THRESHOLD_PERCENT      |
| FLOOD_PRICING     | Plus de 5 modifications en 60 minutes  | FLOOD_MAX_UPDATES            |
| DUMPING           | Prix < 1.00€                            | (hardcodé, extensible)       |

## Endpoints price-tracker-service

| Méthode | Route                            | Description                     |
|---------|----------------------------------|---------------------------------|
| GET     | /api/v1/health                   | Health check                    |
| GET     | /api/v1/items/:id/price-history  | Historique des prix d'un article|
| GET     | /api/v1/alerts?unresolved=true   | Liste les alertes fraude        |
| PUT     | /api/v1/alerts/:id/resolve       | Résout une alerte               |

## Endpoints notification-service

| Méthode | Route                               | Description                        |
|---------|-------------------------------------|------------------------------------|
| GET     | /ws?token=<jwt>                     | WebSocket (temps réel)             |
| GET     | /api/v1/health                      | Health check + nb connexions WS    |
| GET     | /api/v1/notifications               | Historique des notifs              |
| PUT     | /api/v1/notifications/:id/read      | Marquer comme lue                  |
| PUT     | /api/v1/notifications/read-all      | Tout marquer comme lu              |
| GET     | /api/v1/notifications/unread-count  | Compteur non lues                  |

## Structure du projet

```
collector/
├── docker-compose.yml
├── price-tracker-service/
│   ├── cmd/main.go
│   ├── config/config.go
│   ├── internal/
│   │   ├── consumer/consumer.go     # RabbitMQ consumer + publisher
│   │   ├── detector/detector.go     # Règles anti-fraude
│   │   ├── detector/detector_test.go
│   │   ├── handler/handler.go       # Routes Gin
│   │   ├── model/model.go
│   │   └── repository/repository.go
│   ├── Dockerfile
│   └── go.mod
└── notification-service/
    ├── cmd/main.go
    ├── config/config.go
    ├── internal/
    │   ├── consumer/consumer.go     # 2 consumers RabbitMQ
    │   ├── handler/handler.go       # Routes Gin + WebSocket upgrade
    │   ├── hub/hub.go               # WebSocket Hub (goroutine-safe)
    │   ├── hub/hub_test.go
    │   ├── model/model.go
    │   └── repository/repository.go
    ├── Dockerfile
    └── go.mod
```
