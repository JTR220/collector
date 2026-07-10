// Package idconv convertit entre les IDs numeriques gorm (auth-service,
// catalog-service) et les UUID deterministes utilises dans les evenements
// AMQP et les claims JWT ("00000000-0000-0000-0000-<id hex>", voir
// events.ToEventUUID cote catalog-service et le claim "sub" pose au login).
package idconv

import (
	"fmt"

	"github.com/google/uuid"
)

// FromUUID inverse le mapping deterministe ID numerique -> UUID : renvoie 0
// si l'UUID ne suit pas ce format (par exemple un vrai UUID aleatoire).
func FromUUID(id uuid.UUID) uint {
	s := id.String()
	if len(s) < 12 {
		return 0
	}
	var n uint64
	_, _ = fmt.Sscanf(s[len(s)-12:], "%x", &n)
	return uint(n)
}

// ToUUID applique le mapping deterministe direct ID numerique -> UUID,
// symetrique de FromUUID et identique a events.ToEventUUID cote
// catalog-service (les deux doivent rester en phase).
func ToUUID(id uint) uuid.UUID {
	return uuid.MustParse(fmt.Sprintf("00000000-0000-0000-0000-%012x", id))
}
