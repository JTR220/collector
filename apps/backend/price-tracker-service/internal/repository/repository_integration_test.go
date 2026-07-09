package repository_test

// Tests d'integration contre une vraie base Postgres (types specifiques :
// UUID, TIMESTAMPTZ, NUMERIC, contrainte ON CONFLICT sur processed_messages)
// qu'une base en memoire ne peut pas emuler fidelement. Gardes derriere
// TEST_DATABASE_DSN, comme notification-service/internal/repository : ils
// s'auto-desactivent (t.Skip) en local sans base disponible, et tournent
// reellement en CI contre le service "postgres" du workflow backend.yml
// (matrix price-tracker-service).

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/JTR220/collector/price-tracker-service/internal/model"
	"github.com/JTR220/collector/price-tracker-service/internal/repository"
)

func newTestRepo(t *testing.T) *repository.PriceRepository {
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

	repo := repository.NewPriceRepository(db)
	if err := repo.Migrate(); err != nil {
		t.Fatalf("migration : %v", err)
	}
	return repo
}

func newHistory(itemID, sellerID uuid.UUID, oldPrice, newPrice float64, createdAt time.Time) *model.PriceHistory {
	return &model.PriceHistory{
		ID:        uuid.New(),
		ItemID:    itemID,
		SellerID:  sellerID,
		OldPrice:  oldPrice,
		NewPrice:  newPrice,
		CreatedAt: createdAt,
	}
}

func TestIntegration_SavePriceHistoryThenGetPriceHistory(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	itemID, sellerID := uuid.New(), uuid.New()
	if err := repo.SavePriceHistory(ctx, newHistory(itemID, sellerID, 100, 120, time.Now())); err != nil {
		t.Fatalf("SavePriceHistory : %v", err)
	}
	if err := repo.SavePriceHistory(ctx, newHistory(itemID, sellerID, 120, 90, time.Now())); err != nil {
		t.Fatalf("SavePriceHistory (2e entree) : %v", err)
	}

	history, err := repo.GetPriceHistory(ctx, itemID)
	if err != nil {
		t.Fatalf("GetPriceHistory : %v", err)
	}
	if len(history) != 2 {
		t.Fatalf("attendu 2 entrees d'historique, obtenu %d", len(history))
	}
	// ORDER BY created_at DESC : la plus recente (120 -> 90) en premier.
	if history[0].NewPrice != 90 {
		t.Errorf("attendu la modification la plus recente en premier (90), obtenu %.2f", history[0].NewPrice)
	}
}

func TestIntegration_GetPriceHistoryIsolatedByItem(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	itemA, itemB, seller := uuid.New(), uuid.New(), uuid.New()
	if err := repo.SavePriceHistory(ctx, newHistory(itemA, seller, 100, 110, time.Now())); err != nil {
		t.Fatalf("SavePriceHistory (item A) : %v", err)
	}
	if err := repo.SavePriceHistory(ctx, newHistory(itemB, seller, 200, 210, time.Now())); err != nil {
		t.Fatalf("SavePriceHistory (item B) : %v", err)
	}

	historyA, err := repo.GetPriceHistory(ctx, itemA)
	if err != nil {
		t.Fatalf("GetPriceHistory (item A) : %v", err)
	}
	if len(historyA) != 1 || historyA[0].NewPrice != 110 {
		t.Errorf("l'historique de l'item A ne devrait contenir que sa propre entree, obtenu %+v", historyA)
	}
}

func TestIntegration_CountUpdatesInWindow(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	itemID, seller := uuid.New(), uuid.New()
	now := time.Now()
	// 2 modifications recentes, 1 ancienne (hors fenetre de 60 minutes).
	if err := repo.SavePriceHistory(ctx, newHistory(itemID, seller, 100, 105, now.Add(-10*time.Minute))); err != nil {
		t.Fatalf("SavePriceHistory (recent 1) : %v", err)
	}
	if err := repo.SavePriceHistory(ctx, newHistory(itemID, seller, 105, 108, now.Add(-5*time.Minute))); err != nil {
		t.Fatalf("SavePriceHistory (recent 2) : %v", err)
	}
	if err := repo.SavePriceHistory(ctx, newHistory(itemID, seller, 90, 100, now.Add(-3*time.Hour))); err != nil {
		t.Fatalf("SavePriceHistory (ancien) : %v", err)
	}

	count, err := repo.CountUpdatesInWindow(ctx, itemID, 60)
	if err != nil {
		t.Fatalf("CountUpdatesInWindow : %v", err)
	}
	if count != 2 {
		t.Errorf("attendu 2 modifications dans la fenetre de 60 minutes, obtenu %d", count)
	}
}

func TestIntegration_GetLastPrice(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	itemID, seller := uuid.New(), uuid.New()
	now := time.Now()
	if err := repo.SavePriceHistory(ctx, newHistory(itemID, seller, 100, 130, now.Add(-2*time.Hour))); err != nil {
		t.Fatalf("SavePriceHistory (ancien) : %v", err)
	}
	if err := repo.SavePriceHistory(ctx, newHistory(itemID, seller, 130, 150, now.Add(-1*time.Hour))); err != nil {
		t.Fatalf("SavePriceHistory (recent) : %v", err)
	}

	// La plus ancienne entree DANS la fenetre (ORDER BY created_at ASC LIMIT 1).
	price, err := repo.GetLastPrice(ctx, itemID, 3*time.Hour)
	if err != nil {
		t.Fatalf("GetLastPrice : %v", err)
	}
	if price != 130 {
		t.Errorf("attendu le prix de reference le plus ancien dans la fenetre (130), obtenu %.2f", price)
	}
}

func TestIntegration_MarkProcessedIsIdempotent(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	messageID := "price.updated:" + uuid.NewString()

	firstSeen, err := repo.MarkProcessed(ctx, messageID)
	if err != nil {
		t.Fatalf("MarkProcessed (1er passage) : %v", err)
	}
	if !firstSeen {
		t.Error("le premier passage devrait etre signale comme nouveau (firstSeen=true)")
	}

	firstSeen, err = repo.MarkProcessed(ctx, messageID)
	if err != nil {
		t.Fatalf("MarkProcessed (redelivree) : %v", err)
	}
	if firstSeen {
		t.Error("une redelivraison du meme message_id ne devrait pas etre signalee comme nouvelle")
	}
}

func TestIntegration_SaveAlertThenResolve(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	alert := &model.FraudAlert{
		ID:        uuid.New(),
		ItemID:    uuid.New(),
		SellerID:  uuid.New(),
		Reason:    model.ReasonSuspiciousSpike,
		Detail:    "test",
		OldPrice:  100,
		NewPrice:  250,
		Resolved:  false,
		CreatedAt: time.Now(),
	}
	if err := repo.SaveAlert(ctx, alert); err != nil {
		t.Fatalf("SaveAlert : %v", err)
	}

	unresolved, err := repo.GetAlerts(ctx, true)
	if err != nil {
		t.Fatalf("GetAlerts(true) : %v", err)
	}
	found := false
	for _, a := range unresolved {
		if a.ID == alert.ID {
			found = true
		}
	}
	if !found {
		t.Fatal("l'alerte creee devrait apparaitre parmi les alertes non resolues")
	}

	if err := repo.ResolveAlert(ctx, alert.ID); err != nil {
		t.Fatalf("ResolveAlert : %v", err)
	}

	unresolved, err = repo.GetAlerts(ctx, true)
	if err != nil {
		t.Fatalf("GetAlerts(true) apres resolution : %v", err)
	}
	for _, a := range unresolved {
		if a.ID == alert.ID {
			t.Error("une alerte resolue ne devrait plus apparaitre parmi les alertes non resolues")
		}
	}
}
