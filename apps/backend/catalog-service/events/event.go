package events

import (
	"fmt"
	"time"
)

// PriceUpdatedEvent correspond au format attendu par price-tracker-service
// et notification-service (voir internal/model/model.go de chacun) :
// les IDs y sont des uuid.UUID, d'ou le mapping deterministe ToEventUUID.
type PriceUpdatedEvent struct {
	ItemID    string    `json:"item_id"`
	SellerID  string    `json:"seller_id"`
	OldPrice  float64   `json:"old_price"`
	NewPrice  float64   `json:"new_price"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToEventUUID convertit un ID gorm (uint) en UUID deterministe et reversible,
// partage avec le front (src/lib/utils/eventId.ts).
func ToEventUUID(id uint) string {
	return fmt.Sprintf("00000000-0000-0000-0000-%012x", id)
}

// OrderCreatedEvent est publie quand un acheteur passe commande sur une
// annonce : le vendeur doit valider ou refuser avant que la vente soit actee.
type OrderCreatedEvent struct {
	OrderID   string    `json:"order_id"`
	ItemID    string    `json:"item_id"`
	ItemName  string    `json:"item_name"`
	BuyerID   string    `json:"buyer_id"`
	SellerID  string    `json:"seller_id"`
	Price     float64   `json:"price"`
	CreatedAt time.Time `json:"created_at"`
}

// OrderDecisionEvent est publie quand le vendeur accepte ou refuse une commande.
type OrderDecisionEvent struct {
	OrderID   string    `json:"order_id"`
	ItemID    string    `json:"item_id"`
	ItemName  string    `json:"item_name"`
	BuyerID   string    `json:"buyer_id"`
	SellerID  string    `json:"seller_id"`
	Price     float64   `json:"price"`
	Accepted  bool      `json:"accepted"`
	DecidedAt time.Time `json:"decided_at"`
}

// OfferCreatedEvent est publie quand un acheteur propose un prix negocie sur
// une annonce : le vendeur doit accepter ou refuser cette offre.
type OfferCreatedEvent struct {
	OfferID   string    `json:"offer_id"`
	ItemID    string    `json:"item_id"`
	ItemName  string    `json:"item_name"`
	BuyerID   string    `json:"buyer_id"`
	SellerID  string    `json:"seller_id"`
	Price     float64   `json:"price"`
	ListPrice float64   `json:"list_price"`
	CreatedAt time.Time `json:"created_at"`
}

// OfferDecisionEvent est publie quand le vendeur accepte ou refuse une offre.
type OfferDecisionEvent struct {
	OfferID   string    `json:"offer_id"`
	ItemID    string    `json:"item_id"`
	ItemName  string    `json:"item_name"`
	BuyerID   string    `json:"buyer_id"`
	SellerID  string    `json:"seller_id"`
	Price     float64   `json:"price"`
	Accepted  bool      `json:"accepted"`
	DecidedAt time.Time `json:"decided_at"`
}

// OfferPurchasedEvent est publie quand l'acheteur a paye une offre acceptee :
// la vente est finalisee au prix negocie.
type OfferPurchasedEvent struct {
	OfferID     string    `json:"offer_id"`
	OrderID     string    `json:"order_id"`
	ItemID      string    `json:"item_id"`
	ItemName    string    `json:"item_name"`
	BuyerID     string    `json:"buyer_id"`
	SellerID    string    `json:"seller_id"`
	Price       float64   `json:"price"`
	PurchasedAt time.Time `json:"purchased_at"`
}
