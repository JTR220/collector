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
