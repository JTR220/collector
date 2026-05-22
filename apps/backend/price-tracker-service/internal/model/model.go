package model

import (
	"time"

	"github.com/google/uuid"
)

// PriceHistory stores each price change for an item
type PriceHistory struct {
	ID        uuid.UUID `db:"id"         json:"id"`
	ItemID    uuid.UUID `db:"item_id"    json:"item_id"`
	SellerID  uuid.UUID `db:"seller_id"  json:"seller_id"`
	OldPrice  float64   `db:"old_price"  json:"old_price"`
	NewPrice  float64   `db:"new_price"  json:"new_price"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// FraudAlert is raised when a detection rule is triggered
type FraudAlert struct {
	ID        uuid.UUID  `db:"id"          json:"id"`
	ItemID    uuid.UUID  `db:"item_id"     json:"item_id"`
	SellerID  uuid.UUID  `db:"seller_id"   json:"seller_id"`
	Reason    AlertReason `db:"reason"     json:"reason"`
	Detail    string     `db:"detail"      json:"detail"`
	OldPrice  float64    `db:"old_price"   json:"old_price"`
	NewPrice  float64    `db:"new_price"   json:"new_price"`
	Resolved  bool       `db:"resolved"    json:"resolved"`
	CreatedAt time.Time  `db:"created_at"  json:"created_at"`
}

type AlertReason string

const (
	ReasonSuspiciousSpike AlertReason = "SUSPICIOUS_SPIKE"  // prix +50% en <24h
	ReasonFloodPricing    AlertReason = "FLOOD_PRICING"     // >5 modifs en 1h
	ReasonDumping         AlertReason = "DUMPING"           // prix anormalement bas
)

// PriceUpdatedEvent matches the message published by catalog-service
type PriceUpdatedEvent struct {
	ItemID    uuid.UUID `json:"item_id"`
	SellerID  uuid.UUID `json:"seller_id"`
	OldPrice  float64   `json:"old_price"`
	NewPrice  float64   `json:"new_price"`
	UpdatedAt time.Time `json:"updated_at"`
}

// FraudAlertEvent is published to collector.alerts exchange
type FraudAlertEvent struct {
	AlertID   uuid.UUID   `json:"alert_id"`
	ItemID    uuid.UUID   `json:"item_id"`
	SellerID  uuid.UUID   `json:"seller_id"`
	Reason    AlertReason `json:"reason"`
	Detail    string      `json:"detail"`
	OldPrice  float64     `json:"old_price"`
	NewPrice  float64     `json:"new_price"`
	CreatedAt time.Time   `json:"created_at"`
}
