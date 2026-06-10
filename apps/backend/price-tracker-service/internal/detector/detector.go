package detector

import (
	"context"
	"fmt"
	"time"

	"github.com/JTR220/collector/price-tracker-service/config"
	"github.com/JTR220/collector/price-tracker-service/internal/model"
	"github.com/JTR220/collector/price-tracker-service/internal/repository"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// Detector runs all fraud detection rules against a price event
type Detector struct {
	repo  *repository.PriceRepository
	rules config.DetectionRules
}

func New(repo *repository.PriceRepository, rules config.DetectionRules) *Detector {
	return &Detector{repo: repo, rules: rules}
}

// Analyze runs all rules and returns any alerts triggered
func (d *Detector) Analyze(ctx context.Context, event *model.PriceUpdatedEvent) []model.FraudAlert {
	var alerts []model.FraudAlert

	// Rule 1 — SUSPICIOUS_SPIKE : prix augmente de X% en moins de N heures
	if alert := d.checkSpike(ctx, event); alert != nil {
		alerts = append(alerts, *alert)
	}

	// Rule 2 — FLOOD_PRICING : plus de N modifications en M minutes
	if alert := d.checkFlood(ctx, event); alert != nil {
		alerts = append(alerts, *alert)
	}

	// Rule 3 — DUMPING : prix anormalement bas (< 1€)
	if alert := d.checkDumping(ctx, event); alert != nil {
		alerts = append(alerts, *alert)
	}

	return alerts
}

func (d *Detector) checkSpike(ctx context.Context, event *model.PriceUpdatedEvent) *model.FraudAlert {
	if event.OldPrice <= 0 {
		return nil
	}

	deltaPercent := ((event.NewPrice - event.OldPrice) / event.OldPrice) * 100

	// Cherche le prix de référence dans la fenêtre glissante
	refWindow := time.Duration(d.rules.SpikeWindowHours) * time.Hour
	refPrice, err := d.repo.GetLastPrice(ctx, event.ItemID, refWindow)
	if err != nil || refPrice <= 0 {
		refPrice = event.OldPrice
	}

	windowDelta := ((event.NewPrice - refPrice) / refPrice) * 100

	if windowDelta >= d.rules.SpikeThresholdPercent {
		log.Warn().
			Str("item_id", event.ItemID.String()).
			Float64("delta_percent", windowDelta).
			Msg("SUSPICIOUS_SPIKE détecté")

		return &model.FraudAlert{
			ID:        uuid.New(),
			ItemID:    event.ItemID,
			SellerID:  event.SellerID,
			Reason:    model.ReasonSuspiciousSpike,
			Detail:    fmt.Sprintf("Hausse de %.1f%% en %dh (seuil: %.0f%%)", windowDelta, d.rules.SpikeWindowHours, d.rules.SpikeThresholdPercent),
			OldPrice:  refPrice,
			NewPrice:  event.NewPrice,
			Resolved:  false,
			CreatedAt: time.Now(),
		}
	}
	return nil
}

func (d *Detector) checkFlood(ctx context.Context, event *model.PriceUpdatedEvent) *model.FraudAlert {
	count, err := d.repo.CountUpdatesInWindow(ctx, event.ItemID, d.rules.FloodWindowMinutes)
	if err != nil {
		log.Error().Err(err).Msg("flood check failed")
		return nil
	}

	if count >= d.rules.FloodMaxUpdates {
		log.Warn().
			Str("item_id", event.ItemID.String()).
			Int("count", count).
			Msg("FLOOD_PRICING détecté")

		return &model.FraudAlert{
			ID:        uuid.New(),
			ItemID:    event.ItemID,
			SellerID:  event.SellerID,
			Reason:    model.ReasonFloodPricing,
			Detail:    fmt.Sprintf("%d modifications en %d minutes (seuil: %d)", count, d.rules.FloodWindowMinutes, d.rules.FloodMaxUpdates),
			OldPrice:  event.OldPrice,
			NewPrice:  event.NewPrice,
			Resolved:  false,
			CreatedAt: time.Now(),
		}
	}
	return nil
}

func (d *Detector) checkDumping(ctx context.Context, event *model.PriceUpdatedEvent) *model.FraudAlert {
	const dumpingThreshold = 1.0 // en euros

	if event.NewPrice < dumpingThreshold {
		log.Warn().
			Str("item_id", event.ItemID.String()).
			Float64("new_price", event.NewPrice).
			Msg("DUMPING détecté")

		return &model.FraudAlert{
			ID:        uuid.New(),
			ItemID:    event.ItemID,
			SellerID:  event.SellerID,
			Reason:    model.ReasonDumping,
			Detail:    fmt.Sprintf("Prix anormalement bas: %.2f€ (seuil: %.2f€)", event.NewPrice, dumpingThreshold),
			OldPrice:  event.OldPrice,
			NewPrice:  event.NewPrice,
			Resolved:  false,
			CreatedAt: time.Now(),
		}
	}
	return nil
}
