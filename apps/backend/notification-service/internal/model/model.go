package model

import (
	"time"

	"github.com/google/uuid"
)

// NotificationType defines what kind of event triggered the notification
type NotificationType string

const (
	TypePriceDrop     NotificationType = "PRICE_DROP"     // baisse de prix
	TypePriceSpike    NotificationType = "PRICE_SPIKE"    // hausse de prix
	TypeFraudAlert    NotificationType = "FRAUD_ALERT"    // alerte fraude détectée
	TypeNewItem       NotificationType = "NEW_ITEM"       // nouvel article dans une catégorie suivie
	TypeItemSold      NotificationType = "ITEM_SOLD"      // article vendu (pour le vendeur)
	TypeOrderPending  NotificationType = "ORDER_PENDING"  // nouvelle commande a valider (pour le vendeur)
	TypeOrderAccepted NotificationType = "ORDER_ACCEPTED" // commande acceptee (pour l'acheteur)
	TypeOrderRejected NotificationType = "ORDER_REJECTED" // commande refusee (pour l'acheteur)

	TypeOfferReceived  NotificationType = "OFFER_RECEIVED"  // nouvelle offre de prix a traiter (pour le vendeur)
	TypeOfferAccepted  NotificationType = "OFFER_ACCEPTED"  // offre acceptee, paiement possible (pour l'acheteur)
	TypeOfferRejected  NotificationType = "OFFER_REJECTED"  // offre refusee (pour l'acheteur)
	TypeOfferPurchased NotificationType = "OFFER_PURCHASED" // offre payee, vente finalisee (pour le vendeur)
)

// Notification is persisted in the database and sent via WebSocket
type Notification struct {
	ID        uuid.UUID        `db:"id"          json:"id"`
	UserID    uuid.UUID        `db:"user_id"     json:"user_id"`
	Type      NotificationType `db:"type"        json:"type"`
	Title     string           `db:"title"       json:"title"`
	Body      string           `db:"body"        json:"body"`
	ItemID    *uuid.UUID       `db:"item_id"     json:"item_id,omitempty"`
	Read      bool             `db:"read"        json:"read"`
	CreatedAt time.Time        `db:"created_at"  json:"created_at"`
}

// Message is a direct message between two users (ex: acheteur ↔ vendeur au
// sujet d'une annonce), persiste et pousse en temps reel via WebSocket.
type Message struct {
	ID             uuid.UUID  `db:"id"              json:"id"`
	ConversationID uuid.UUID  `db:"conversation_id" json:"conversation_id"`
	SenderID       uuid.UUID  `db:"sender_id"        json:"sender_id"`
	SenderName     string     `db:"sender_name"      json:"sender_name"`
	RecipientID    uuid.UUID  `db:"recipient_id"     json:"recipient_id"`
	RecipientName  string     `db:"recipient_name"   json:"recipient_name"`
	ArticleID      *uuid.UUID `db:"article_id"       json:"article_id,omitempty"`
	ArticleName    string     `db:"article_name"     json:"article_name,omitempty"`
	Body           string     `db:"body"             json:"body"`
	Read           bool       `db:"read"             json:"read"`
	CreatedAt      time.Time  `db:"created_at"       json:"created_at"`
}

// ConversationSummary resume le dernier message d'un fil pour l'utilisateur
// courant (liste des conversations).
type ConversationSummary struct {
	ConversationID uuid.UUID  `db:"conversation_id" json:"conversation_id"`
	OtherUserID    uuid.UUID  `db:"other_user_id"   json:"other_user_id"`
	OtherUserName  string     `db:"other_user_name" json:"other_user_name"`
	ArticleID      *uuid.UUID `db:"article_id"      json:"article_id,omitempty"`
	ArticleName    string     `db:"article_name"    json:"article_name,omitempty"`
	LastMessage    string     `db:"last_message"    json:"last_message"`
	LastAt         time.Time  `db:"last_at"         json:"last_at"`
	UnreadCount    int        `db:"-"               json:"unread_count"`
}

// WebSocketMessage is the payload sent to the browser client
type WebSocketMessage struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

// PriceUpdatedEvent from catalog-service (via collector.events)
type PriceUpdatedEvent struct {
	ItemID    uuid.UUID `json:"item_id"`
	SellerID  uuid.UUID `json:"seller_id"`
	OldPrice  float64   `json:"old_price"`
	NewPrice  float64   `json:"new_price"`
	UpdatedAt time.Time `json:"updated_at"`
}

// OrderCreatedEvent from catalog-service (via collector.events) : un acheteur
// vient de passer commande, le vendeur doit valider ou refuser.
type OrderCreatedEvent struct {
	OrderID   uuid.UUID `json:"order_id"`
	ItemID    uuid.UUID `json:"item_id"`
	ItemName  string    `json:"item_name"`
	BuyerID   uuid.UUID `json:"buyer_id"`
	SellerID  uuid.UUID `json:"seller_id"`
	Price     float64   `json:"price"`
	CreatedAt time.Time `json:"created_at"`
}

// OrderDecisionEvent from catalog-service (via collector.events) : le vendeur
// a accepte ou refuse une commande.
type OrderDecisionEvent struct {
	OrderID   uuid.UUID `json:"order_id"`
	ItemID    uuid.UUID `json:"item_id"`
	ItemName  string    `json:"item_name"`
	BuyerID   uuid.UUID `json:"buyer_id"`
	SellerID  uuid.UUID `json:"seller_id"`
	Price     float64   `json:"price"`
	Accepted  bool      `json:"accepted"`
	DecidedAt time.Time `json:"decided_at"`
}

// OfferCreatedEvent from catalog-service (via collector.events) : un acheteur
// vient de proposer un prix negocie, le vendeur doit accepter ou refuser.
type OfferCreatedEvent struct {
	OfferID   uuid.UUID `json:"offer_id"`
	ItemID    uuid.UUID `json:"item_id"`
	ItemName  string    `json:"item_name"`
	BuyerID   uuid.UUID `json:"buyer_id"`
	SellerID  uuid.UUID `json:"seller_id"`
	Price     float64   `json:"price"`
	ListPrice float64   `json:"list_price"`
	CreatedAt time.Time `json:"created_at"`
}

// OfferDecisionEvent from catalog-service (via collector.events) : le vendeur
// a accepte ou refuse une offre.
type OfferDecisionEvent struct {
	OfferID   uuid.UUID `json:"offer_id"`
	ItemID    uuid.UUID `json:"item_id"`
	ItemName  string    `json:"item_name"`
	BuyerID   uuid.UUID `json:"buyer_id"`
	SellerID  uuid.UUID `json:"seller_id"`
	Price     float64   `json:"price"`
	Accepted  bool      `json:"accepted"`
	DecidedAt time.Time `json:"decided_at"`
}

// OfferPurchasedEvent from catalog-service (via collector.events) : l'acheteur
// a paye une offre acceptee, la vente est finalisee.
type OfferPurchasedEvent struct {
	OfferID     uuid.UUID `json:"offer_id"`
	OrderID     uuid.UUID `json:"order_id"`
	ItemID      uuid.UUID `json:"item_id"`
	ItemName    string    `json:"item_name"`
	BuyerID     uuid.UUID `json:"buyer_id"`
	SellerID    uuid.UUID `json:"seller_id"`
	Price       float64   `json:"price"`
	PurchasedAt time.Time `json:"purchased_at"`
}

// FraudAlertEvent from price-tracker-service (via collector.alerts)
type FraudAlertEvent struct {
	AlertID   uuid.UUID `json:"alert_id"`
	ItemID    uuid.UUID `json:"item_id"`
	SellerID  uuid.UUID `json:"seller_id"`
	Reason    string    `json:"reason"`
	Detail    string    `json:"detail"`
	OldPrice  float64   `json:"old_price"`
	NewPrice  float64   `json:"new_price"`
	CreatedAt time.Time `json:"created_at"`
}
