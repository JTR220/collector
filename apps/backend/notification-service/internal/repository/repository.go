package repository

import (
	"context"
	"sort"

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

	CREATE TABLE IF NOT EXISTS processed_messages (
		message_id   TEXT PRIMARY KEY,
		processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS messages (
		id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		conversation_id UUID NOT NULL,
		sender_id       UUID NOT NULL,
		sender_name     TEXT NOT NULL,
		recipient_id    UUID NOT NULL,
		recipient_name  TEXT NOT NULL,
		article_id      UUID,
		article_name    TEXT,
		body            TEXT NOT NULL,
		read            BOOLEAN NOT NULL DEFAULT FALSE,
		created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_messages_conversation ON messages(conversation_id, created_at);
	CREATE INDEX IF NOT EXISTS idx_messages_participants  ON messages(sender_id, recipient_id);
	CREATE INDEX IF NOT EXISTS idx_messages_recipient_unread ON messages(recipient_id, read);
	`

	ctx := context.Background()
	conn, err := r.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	// Verrou consultatif Postgres : en CI, plusieurs packages de test
	// (internal/repository, internal/handler) se connectent au meme
	// TEST_DATABASE_DSN et appellent Migrate() en parallele. Sans verrou,
	// deux "CREATE TABLE IF NOT EXISTS" concurrents peuvent constater tous
	// les deux que la table n'existe pas encore et se marcher dessus sur le
	// catalogue systeme (pg_type_typname_nsp_index). Le verrou serialise
	// l'execution du DDL ; les advisory locks sont scoped a la connexion,
	// d'ou l'usage explicite de conn plutot que r.db pour lock/DDL/unlock.
	const migrateLockKey = 727163
	if _, err := conn.ExecContext(ctx, "SELECT pg_advisory_lock($1)", migrateLockKey); err != nil {
		return err
	}
	defer func() { _, _ = conn.ExecContext(ctx, "SELECT pg_advisory_unlock($1)", migrateLockKey) }()

	_, err = conn.ExecContext(ctx, schema)
	return err
}

// MarkProcessed enregistre un message comme traite. Renvoie (true, nil) la
// premiere fois qu'on voit ce messageID, (false, nil) s'il a deja ete traite —
// l'appelant doit alors sauter le traitement (idempotence sur redelivery).
func (r *NotificationRepository) MarkProcessed(ctx context.Context, messageID string) (bool, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO processed_messages (message_id) VALUES ($1) ON CONFLICT DO NOTHING`,
		messageID,
	)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
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

// MarkRead marque une notification comme lue, uniquement si elle appartient
// a userID (evite qu'un utilisateur marque comme lue une notification d'autrui).
// found=false signifie que la notification n'existe pas ou n'appartient pas
// a cet utilisateur.
func (r *NotificationRepository) MarkRead(ctx context.Context, notifID, userID uuid.UUID) (found bool, err error) {
	res, err := r.db.ExecContext(ctx,
		`UPDATE notifications SET read = TRUE WHERE id = $1 AND user_id = $2`,
		notifID, userID,
	)
	if err != nil {
		return false, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
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

// ── Messagerie ──────────────────────────────────────────────────────────────

func (r *NotificationRepository) SaveMessage(ctx context.Context, m *model.Message) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO messages (id, conversation_id, sender_id, sender_name, recipient_id, recipient_name, article_id, article_name, body, read, created_at)
		VALUES (:id, :conversation_id, :sender_id, :sender_name, :recipient_id, :recipient_name, :article_id, :article_name, :body, :read, :created_at)
	`, m)
	return err
}

// GetConversations renvoie, pour chaque fil ou l'utilisateur est participant,
// le dernier message et le nombre de messages non lus recus dans ce fil.
func (r *NotificationRepository) GetConversations(ctx context.Context, userID uuid.UUID) ([]model.ConversationSummary, error) {
	var convs []model.ConversationSummary
	err := r.db.SelectContext(ctx, &convs, `
		SELECT DISTINCT ON (conversation_id)
			conversation_id,
			CASE WHEN sender_id = $1 THEN recipient_id ELSE sender_id END AS other_user_id,
			CASE WHEN sender_id = $1 THEN recipient_name ELSE sender_name END AS other_user_name,
			article_id, article_name,
			body AS last_message,
			created_at AS last_at
		FROM messages
		WHERE sender_id = $1 OR recipient_id = $1
		ORDER BY conversation_id, created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}

	type unreadRow struct {
		ConversationID uuid.UUID `db:"conversation_id"`
		Count          int       `db:"count"`
	}
	var unread []unreadRow
	if err := r.db.SelectContext(ctx, &unread, `
		SELECT conversation_id, COUNT(*) as count FROM messages
		WHERE recipient_id = $1 AND read = FALSE
		GROUP BY conversation_id
	`, userID); err != nil {
		return nil, err
	}
	unreadByConv := make(map[uuid.UUID]int, len(unread))
	for _, u := range unread {
		unreadByConv[u.ConversationID] = u.Count
	}
	for i := range convs {
		convs[i].UnreadCount = unreadByConv[convs[i].ConversationID]
	}

	// Tri par derniere activite (les DISTINCT ON precedents sont ordonnes par
	// conversation_id, pas par date).
	sort.Slice(convs, func(i, j int) bool { return convs[i].LastAt.After(convs[j].LastAt) })

	return convs, nil
}

// GetMessages renvoie les messages d'un fil, uniquement si userID y participe.
func (r *NotificationRepository) GetMessages(ctx context.Context, conversationID, userID uuid.UUID, limit int) ([]model.Message, error) {
	var msgs []model.Message
	err := r.db.SelectContext(ctx, &msgs, `
		SELECT * FROM messages
		WHERE conversation_id = $1 AND (sender_id = $2 OR recipient_id = $2)
		ORDER BY created_at ASC
		LIMIT $3
	`, conversationID, userID, limit)
	return msgs, err
}

// MarkConversationRead marque comme lus les messages recus par userID dans
// ce fil.
func (r *NotificationRepository) MarkConversationRead(ctx context.Context, conversationID, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE messages SET read = TRUE WHERE conversation_id = $1 AND recipient_id = $2 AND read = FALSE`,
		conversationID, userID,
	)
	return err
}

func (r *NotificationRepository) UnreadMessagesCount(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count,
		`SELECT COUNT(*) FROM messages WHERE recipient_id = $1 AND read = FALSE`,
		userID,
	)
	return count, err
}
