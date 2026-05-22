package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/JTR220/collector/price-tracker-service/internal/model"
)

type PriceRepository struct {
	db *sqlx.DB
}

func NewPriceRepository(db *sqlx.DB) *PriceRepository {
	return &PriceRepository{db: db}
}

// Migrate creates the tables if they don't exist
func (r *PriceRepository) Migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS price_history (
		id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		item_id    UUID NOT NULL,
		seller_id  UUID NOT NULL,
		old_price  NUMERIC(12,2) NOT NULL,
		new_price  NUMERIC(12,2) NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_price_history_item_id ON price_history(item_id);
	CREATE INDEX IF NOT EXISTS idx_price_history_created_at ON price_history(created_at);

	CREATE TABLE IF NOT EXISTS fraud_alerts (
		id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		item_id    UUID NOT NULL,
		seller_id  UUID NOT NULL,
		reason     TEXT NOT NULL,
		detail     TEXT,
		old_price  NUMERIC(12,2) NOT NULL,
		new_price  NUMERIC(12,2) NOT NULL,
		resolved   BOOLEAN NOT NULL DEFAULT FALSE,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_fraud_alerts_item_id ON fraud_alerts(item_id);
	CREATE INDEX IF NOT EXISTS idx_fraud_alerts_resolved ON fraud_alerts(resolved);
	`
	_, err := r.db.Exec(schema)
	return err
}

// SavePriceHistory persists a price change
func (r *PriceRepository) SavePriceHistory(ctx context.Context, h *model.PriceHistory) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO price_history (id, item_id, seller_id, old_price, new_price, created_at)
		VALUES (:id, :item_id, :seller_id, :old_price, :new_price, :created_at)
	`, h)
	return err
}

// GetPriceHistory returns full price history for an item
func (r *PriceRepository) GetPriceHistory(ctx context.Context, itemID uuid.UUID) ([]model.PriceHistory, error) {
	var history []model.PriceHistory
	err := r.db.SelectContext(ctx, &history,
		`SELECT * FROM price_history WHERE item_id = $1 ORDER BY created_at DESC`,
		itemID,
	)
	return history, err
}

// CountUpdatesInWindow counts updates for an item in the last N minutes
func (r *PriceRepository) CountUpdatesInWindow(ctx context.Context, itemID uuid.UUID, minutes int) (int, error) {
	var count int
	since := time.Now().Add(-time.Duration(minutes) * time.Minute)
	err := r.db.GetContext(ctx, &count,
		`SELECT COUNT(*) FROM price_history WHERE item_id = $1 AND created_at >= $2`,
		itemID, since,
	)
	return count, err
}

// GetLastPrice returns the last recorded price before the current event
func (r *PriceRepository) GetLastPrice(ctx context.Context, itemID uuid.UUID, since time.Duration) (float64, error) {
	var price float64
	cutoff := time.Now().Add(-since)
	err := r.db.GetContext(ctx, &price,
		`SELECT new_price FROM price_history 
		 WHERE item_id = $1 AND created_at >= $2
		 ORDER BY created_at ASC LIMIT 1`,
		itemID, cutoff,
	)
	return price, err
}

// SaveAlert persists a fraud alert
func (r *PriceRepository) SaveAlert(ctx context.Context, alert *model.FraudAlert) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO fraud_alerts (id, item_id, seller_id, reason, detail, old_price, new_price, resolved, created_at)
		VALUES (:id, :item_id, :seller_id, :reason, :detail, :old_price, :new_price, :resolved, :created_at)
	`, alert)
	return err
}

// GetAlerts returns all unresolved fraud alerts
func (r *PriceRepository) GetAlerts(ctx context.Context, onlyUnresolved bool) ([]model.FraudAlert, error) {
	var alerts []model.FraudAlert
	query := `SELECT * FROM fraud_alerts`
	if onlyUnresolved {
		query += ` WHERE resolved = FALSE`
	}
	query += ` ORDER BY created_at DESC`
	err := r.db.SelectContext(ctx, &alerts, query)
	return alerts, err
}

// ResolveAlert marks an alert as resolved
func (r *PriceRepository) ResolveAlert(ctx context.Context, alertID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE fraud_alerts SET resolved = TRUE WHERE id = $1`,
		alertID,
	)
	return err
}
