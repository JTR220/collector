package model

import (
	"time"

	"github.com/google/uuid"
)

// NotificationType defines what kind of event triggered the notification
type NotificationType string

const (
	TypePriceDrop    NotificationType = "PRICE_DROP"    // baisse de prix
	TypePriceSpike   NotificationType = "PRICE_SPIKE"   // hausse de prix
	TypeFraudAlert   NotificationType = "FRAUD_ALERT"   // alerte fraude détectée
	TypeNewItem      NotificationType = "NEW_ITEM"       // nouvel article dans une catégorie suivie
	TypeItemSold     NotificationType = "ITEM_SOLD"      // article vendu (pour le vendeur)
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
