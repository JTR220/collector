package repository_test

// Tests d'integration de la messagerie contre une vraie base Postgres (les
// requetes utilisent des fonctionnalites specifiques : gen_random_uuid(),
// colonnes UUID/TIMESTAMPTZ, DISTINCT ON) qu'une base en memoire ne peut pas
// emuler fidelement. Ils sont donc gardes derriere TEST_DATABASE_DSN : ils
// s'auto-desactivent (t.Skip) si aucune base n'est disponible en local, et
// tournent reellement en CI contre le service "postgres" du workflow
// backend.yml (matrix notification-service).
//
// Voir aussi ../handler/handler_integration_test.go pour le test
// d'acceptation HTTP correspondant (au niveau du routeur complet).

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/JTR220/collector/notification-service/internal/model"
	"github.com/JTR220/collector/notification-service/internal/repository"
)

func newTestRepo(t *testing.T) *repository.NotificationRepository {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN non defini : test d'integration Postgres ignore (voir CI backend.yml pour l'execution reelle)")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connexion Postgres de test : %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	repo := repository.New(db)
	if err := repo.Migrate(); err != nil {
		t.Fatalf("migration : %v", err)
	}
	return repo
}

func newMessage(convID, sender, recipient uuid.UUID, body string) *model.Message {
	return &model.Message{
		ID:             uuid.New(),
		ConversationID: convID,
		SenderID:       sender,
		SenderName:     "Alice",
		RecipientID:    recipient,
		RecipientName:  "Bob",
		Body:           body,
		Read:           false,
		CreatedAt:      time.Now(),
	}
}

// ── Critère d'acceptation 1 : un message envoye apparait dans la liste des conversations du destinataire, non lu ──

func TestIntegration_SendMessageAppearsInRecipientConversations(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	alice, bob := uuid.New(), uuid.New()
	convID := uuid.New()

	if err := repo.SaveMessage(ctx, newMessage(convID, alice, bob, "Bonjour, la piece est-elle toujours dispo ?")); err != nil {
		t.Fatalf("SaveMessage : %v", err)
	}

	convs, err := repo.GetConversations(ctx, bob)
	if err != nil {
		t.Fatalf("GetConversations (destinataire) : %v", err)
	}
	if len(convs) != 1 {
		t.Fatalf("attendu 1 conversation pour le destinataire, obtenu %d", len(convs))
	}
	if convs[0].UnreadCount != 1 {
		t.Errorf("attendu 1 message non lu, obtenu %d", convs[0].UnreadCount)
	}
	if convs[0].OtherUserID != alice {
		t.Errorf("l'autre participant attendu est alice, obtenu %s", convs[0].OtherUserID)
	}
}

// ── Critère d'acceptation 2 : marquer un fil comme lu remet son compteur non-lu a zero ──

func TestIntegration_MarkConversationReadClearsUnreadCount(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	alice, bob := uuid.New(), uuid.New()
	convID := uuid.New()
	if err := repo.SaveMessage(ctx, newMessage(convID, alice, bob, "Toujours interesse ?")); err != nil {
		t.Fatalf("SaveMessage : %v", err)
	}

	if err := repo.MarkConversationRead(ctx, convID, bob); err != nil {
		t.Fatalf("MarkConversationRead : %v", err)
	}

	unread, err := repo.UnreadMessagesCount(ctx, bob)
	if err != nil {
		t.Fatalf("UnreadMessagesCount : %v", err)
	}
	if unread != 0 {
		t.Errorf("compteur non-lu attendu 0 apres lecture, obtenu %d", unread)
	}
}

// ── Critère d'acceptation 3 : un utilisateur etranger au fil ne voit aucun message ──

func TestIntegration_GetMessagesReturnsEmptyForNonParticipant(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	alice, bob, stranger := uuid.New(), uuid.New(), uuid.New()
	convID := uuid.New()
	if err := repo.SaveMessage(ctx, newMessage(convID, alice, bob, "message prive")); err != nil {
		t.Fatalf("SaveMessage : %v", err)
	}

	msgs, err := repo.GetMessages(ctx, convID, stranger, 50)
	if err != nil {
		t.Fatalf("GetMessages : %v", err)
	}
	if len(msgs) != 0 {
		t.Errorf("un utilisateur hors conversation ne devrait voir aucun message, obtenu %d", len(msgs))
	}

	// Les deux participants, eux, voient bien le message.
	msgs, err = repo.GetMessages(ctx, convID, bob, 50)
	if err != nil {
		t.Fatalf("GetMessages (participant) : %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("le destinataire devrait voir 1 message, obtenu %d", len(msgs))
	}
}
