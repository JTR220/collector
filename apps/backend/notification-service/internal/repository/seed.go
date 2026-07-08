package repository

import (
	"context"
	"time"

	"github.com/JTR220/collector/notification-service/internal/idconv"
	"github.com/JTR220/collector/notification-service/internal/model"
	"github.com/google/uuid"
)

// SeedDemoData relie les comptes de demo (auth-service : admin=1, test=2,
// vendeur=3, acheteur=4 sur une base fraiche) et le catalogue de demo
// (catalog-service : VNL-022=4, FIG-101=5, WAT-045=6, voir
// backfillSellerAssignments/backfillDemoOrders cote catalog-service) a des
// notifications et messages illustrant le flux complet sans avoir a
// rejouer les actions a la main :
//   - Testeur recoit une notification ORDER_PENDING : "Vendeur Demo" veut
//     acheter le Bearbrick 1000%, en attente de validation.
//   - Vendeur Demo recoit une notification ORDER_ACCEPTED : sa commande sur
//     le Daft Punk — Discovery a ete acceptee.
//   - Acheteur Demo envoie un message de negociation a Testeur au sujet de
//     la Casio F-91W, encore en vente.
//
// Idempotent : chaque insertion est gardee par un check d'existence, donc
// SeedDemoData peut tourner a chaque demarrage du service sans dupliquer.
func (r *NotificationRepository) SeedDemoData(ctx context.Context) error {
	const (
		testDemoID     = 2
		vendeurDemoID  = 3
		acheteurDemoID = 4

		vnlArticleID = 4 // VNL-022, Daft Punk — Discovery
		figArticleID = 5 // FIG-101, Bearbrick 1000%
		watArticleID = 6 // WAT-045, Casio F-91W
	)

	testUUID := idconv.ToUUID(testDemoID)
	vendeurUUID := idconv.ToUUID(vendeurDemoID)
	acheteurUUID := idconv.ToUUID(acheteurDemoID)

	figItemID := idconv.ToUUID(figArticleID)
	if err := r.seedNotificationOnce(ctx, &model.Notification{
		ID:        uuid.New(),
		UserID:    testUUID,
		Type:      model.TypeOrderPending,
		Title:     "🛒 Nouvelle demande d'achat",
		Body:      "Un acheteur souhaite acquérir \"Bearbrick 1000%\" pour 1180.00€. Validez ou refusez la commande depuis votre profil.",
		ItemID:    &figItemID,
		Read:      false,
		CreatedAt: time.Now(),
	}); err != nil {
		return err
	}

	vnlItemID := idconv.ToUUID(vnlArticleID)
	if err := r.seedNotificationOnce(ctx, &model.Notification{
		ID:        uuid.New(),
		UserID:    vendeurUUID,
		Type:      model.TypeOrderAccepted,
		Title:     "✅ Commande acceptée",
		Body:      "Votre commande sur \"Daft Punk — Discovery\" (320.00€) a été acceptée par le vendeur.",
		ItemID:    &vnlItemID,
		Read:      false,
		CreatedAt: time.Now(),
	}); err != nil {
		return err
	}

	watItemID := idconv.ToUUID(watArticleID)
	if err := r.seedMessageOnce(ctx, &model.Message{
		ID:             uuid.New(),
		ConversationID: uuid.New(),
		SenderID:       acheteurUUID,
		SenderName:     "Acheteur Demo",
		RecipientID:    testUUID,
		RecipientName:  "Testeur",
		ArticleID:      &watItemID,
		ArticleName:    "Casio F-91W",
		Body:           "Bonjour ! Je suis intéressé par la Casio F-91W, seriez-vous prêt à descendre à 75€ ? Merci.",
		Read:           false,
		CreatedAt:      time.Now(),
	}); err != nil {
		return err
	}

	return nil
}

func (r *NotificationRepository) seedNotificationOnce(ctx context.Context, n *model.Notification) error {
	var count int
	if err := r.db.GetContext(ctx, &count,
		`SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND type = $2 AND item_id = $3`,
		n.UserID, n.Type, n.ItemID,
	); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return r.Save(ctx, n)
}

func (r *NotificationRepository) seedMessageOnce(ctx context.Context, m *model.Message) error {
	var count int
	if err := r.db.GetContext(ctx, &count,
		`SELECT COUNT(*) FROM messages WHERE sender_id = $1 AND recipient_id = $2 AND article_id = $3`,
		m.SenderID, m.RecipientID, m.ArticleID,
	); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return r.SaveMessage(ctx, m)
}
