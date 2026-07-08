# Backlog — Fonctionnalité métier : achat avec validation vendeur

## Reformulation des exigences fonctionnelles

Le contexte impose deux règles métier structurantes pour la marketplace :

1. « Un acheteur ne peut pas acheter sa propre annonce » (implicite dans le
   modèle acheteur/vendeur : la plateforme garantit la qualité des
   transactions entre tiers distincts).
2. Le paiement se fait *via* la plateforme et non en direct entre particuliers
   — ce qui suppose un contrôle du vendeur sur la transaction avant qu'elle ne
   soit actée, plutôt qu'un débit automatique et irréversible dès le clic
   acheteur.

La fonctionnalité retenue pour le POC formalise ce second point : **l'achat
direct passe par une étape de validation explicite du vendeur** avant d'être
considéré comme conclu, avec notification en temps réel et par email des deux
parties à chaque étape.

## User story

> **En tant que** vendeur sur collector.shop,
> **je veux** être notifié et pouvoir accepter ou refuser chaque demande
> d'achat sur mes annonces,
> **afin de** garder le contrôle sur mes ventes (disponibilité réelle de la
> pièce, fiabilité de l'acheteur) avant que la transaction ne soit finalisée.

> **En tant qu'** acheteur,
> **je veux** être informé par notification et par email dès que le vendeur
> accepte ou refuse ma demande,
> **afin de** savoir rapidement si mon achat est confirmé ou si je dois
> chercher la pièce ailleurs.

## Critères d'acceptation

| # | Critère | Test automatisé correspondant |
|---|---|---|
| 1 | Un utilisateur ne peut pas acheter sa propre annonce (400) | `TestAcceptance_CannotBuyOwnArticle`, `TestBuyOwnArticleReturns400` |
| 2 | Un achat crée une commande au statut `pending` et réserve la pièce (`sold=true`) | `TestAcceptance_BuyCreatesPendingOrderAndReservesArticle` |
| 3 | Seul le vendeur de la commande peut l'accepter ou la refuser (403 sinon) | `TestAcceptance_OnlySellerCanAcceptOrder` |
| 4 | Le vendeur accepte → la commande passe `paid`, visible dans `/me/orders` (acheteur) et `/me/sales` (vendeur) | `TestAcceptance_SellerAcceptsOrder` |
| 5 | Le vendeur refuse → la commande passe `cancelled`, la pièce redevient disponible et rachetable | `TestAcceptance_SellerRejectsOrderReleasesArticle` |
| 6 | Une commande déjà traitée ne peut pas être traitée une seconde fois (409) | `TestAcceptance_CannotDecideOrderTwice` |
| 7 | Le vendeur reçoit une notification temps réel (WebSocket) + un email dès la demande d'achat | vérifié manuellement (MailHog, `notification-service` events `order.created`) |
| 8 | L'acheteur reçoit une notification + un email dès la décision du vendeur | vérifié manuellement (`order.decided`) |

Les tests des critères 1 à 6 sont des **tests d'acceptation HTTP de bout en
bout** : ils passent par le vrai routeur Gin (`routes.InitRouter`), donc par
le vrai middleware JWT, et non par un appel direct au controller — c'est le
second type de test exigé par le sujet, en complément des tests unitaires
existants (`controllers_test.go`). Fichier :
[`apps/backend/catalog-service/routes/routes_integration_test.go`](../../apps/backend/catalog-service/routes/routes_integration_test.go).

Le critère 7/8 (notification + email) est couvert fonctionnellement par
l'intégration RabbitMQ → `notification-service` → SMTP (MailHog en dev,
`docker-compose.yml`) mais n'est pas encore automatisé en test — hors
périmètre du prototype (documenté comme limite connue).

## Schéma de flux

```
Acheteur                Catalog-service              Notification-service        Vendeur
   │  POST /article/:id/buy                                                         │
   ├──────────────────────►│                                                        │
   │                       │ Order{status=pending}, Article.sold=true               │
   │                       │  publish order.created ──────────────►│                │
   │                       │                                       │ email+notif WS │
   │                       │                                       ├───────────────►│
   │                       │                                       │                │
   │                       │◄──────────── PATCH /order/:id/accept ─────────────────┤
   │                       │ Order{status=paid}                                     │
   │                       │  publish order.decided ──────────────►│                │
   │  email+notif WS       │                                       │                │
   │◄──────────────────────┼───────────────────────────────────────┤                │
```

## Solutions techniques retenues

- **Modèle de données** : `Order.Status` (`pending → paid|cancelled`, GORM/PostgreSQL),
  réservation atomique de l'article via transaction SQL (`UPDATE ... WHERE sold=false`)
  pour éviter une double-vente en cas de requêtes concurrentes.
- **Communication interservice** : événements asynchrones RabbitMQ
  (`order.created`, `order.decided`) sur l'exchange `collector.events` —
  découplage total entre `catalog-service` (source de vérité transactionnelle)
  et `notification-service` (notification temps réel + email), cohérent avec
  l'architecture microservices existante (`price.updated` suit déjà ce
  patron).
- **Résolution d'email** : `notification-service` interroge un endpoint
  interne d'`auth-service` (`GET /internal/users/:id`, protégé par un secret
  partagé `X-Internal-Secret`) plutôt que de dupliquer les données
  utilisateur — respecte la séparation des responsabilités par service.
