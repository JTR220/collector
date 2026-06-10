package repository

import (
	"context"

	"github.com/JTR220/collector/notification-service/internal/model"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type NotificationRepository struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS notifications (
		id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id    UUID NOT NULL,
		type       TEXT NOT NULL,
		title      TEXT NOT NULL,
		body       TEXT NOT NULL,
		item_id    UUID,
		read       BOOLEAN NOT NULL DEFAULT FALSE,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_notif_user_id ON notifications(user_id);
	CREATE INDEX IF NOT EXISTS idx_notif_read    ON notifications(read);
	CREATE INDEX IF NOT EXISTS idx_notif_created ON notifications(created_at DESC);
	`
	_, err := r.db.Exec(schema)
	return err
}

func (r *NotificationRepository) Save(ctx context.Context, n *model.Notification) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO notifications (id, user_id, type, title, body, item_id, read, created_at)
		VALUES (:id, :user_id, :type, :title, :body, :item_id, :read, :created_at)
	`, n)
	return err
}

func (r *NotificationRepository) GetByUser(ctx context.Context, userID uuid.UUID, limit int) ([]model.Notification, error) {
	var notifs []model.Notification
	err := r.db.SelectContext(ctx, &notifs,
		`SELECT * FROM notifications WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2`,
		userID, limit,
	)
	return notifs, err
}

func (r *NotificationRepository) MarkRead(ctx context.Context, notifID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE notifications SET read = TRUE WHERE id = $1`,
		notifID,
	)
	return err
}

func (r *NotificationRepository) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE notifications SET read = TRUE WHERE user_id = $1 AND read = FALSE`,
		userID,
	)
	return err
}

func (r *NotificationRepository) UnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count,
		`SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND read = FALSE`,
		userID,
	)
	return count, err
}
