package detector_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/JTR220/collector/price-tracker-service/config"
	"github.com/JTR220/collector/price-tracker-service/internal/detector"
	"github.com/JTR220/collector/price-tracker-service/internal/model"
	"github.com/JTR220/collector/price-tracker-service/internal/repository"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func defaultRules() config.DetectionRules {
	return config.DetectionRules{
		SpikeThresholdPercent: 50.0,
		SpikeWindowHours:      24,
		FloodMaxUpdates:       5,
		FloodWindowMinutes:    60,
	}
}

func makeEvent(oldPrice, newPrice float64) *model.PriceUpdatedEvent {
	return &model.PriceUpdatedEvent{
		ItemID:    uuid.New(),
		SellerID:  uuid.New(),
		OldPrice:  oldPrice,
		NewPrice:  newPrice,
		UpdatedAt: time.Now(),
	}
}

// newMockDetector cree un detector branche sur une DB sqlmock : les tests
// exercent ainsi le vrai code d'Analyze / checkSpike / checkFlood / checkDumping
// plutot que de dupliquer la logique metier dans le test.
func newMockDetector(t *testing.T, rules config.DetectionRules) (*detector.Detector, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := repository.NewPriceRepository(sqlxDB)
	d := detector.New(repo, rules)
	return d, mock, func() { db.Close() }
}

func TestAnalyze_SpikeDetected(t *testing.T) {
	rules := defaultRules()
	d, mock, closeFn := newMockDetector(t, rules)
	defer closeFn()

	event := makeEvent(100.0, 200.0) // +100% -> depasse le seuil de 50%

	// GetLastPrice sans historique -> fallback sur event.OldPrice
	mock.ExpectQuery("SELECT new_price FROM price_history").
		WillReturnError(sql.ErrNoRows)
	rows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM price_history").WillReturnRows(rows)

	alerts := d.Analyze(context.Background(), event)

	found := false
	for _, a := range alerts {
		if a.Reason == model.ReasonSuspiciousSpike {
			found = true
		}
	}
	if !found {
		t.Errorf("attendu une alerte SUSPICIOUS_SPIKE, obtenu %+v", alerts)
	}
}

func TestAnalyze_NoSpikeBelowThreshold(t *testing.T) {
	rules := defaultRules()
	d, mock, closeFn := newMockDetector(t, rules)
	defer closeFn()

	event := makeEvent(100.0, 130.0) // +30% -> sous le seuil de 50%

	mock.ExpectQuery("SELECT new_price FROM price_history").
		WillReturnError(sql.ErrNoRows)
	rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM price_history").WillReturnRows(rows)

	alerts := d.Analyze(context.Background(), event)
	for _, a := range alerts {
		if a.Reason == model.ReasonSuspiciousSpike {
			t.Errorf("aucune alerte SUSPICIOUS_SPIKE attendue, obtenu %+v", a)
		}
	}
}

func TestAnalyze_SpikeSkippedWhenOldPriceZero(t *testing.T) {
	rules := defaultRules()
	d, mock, closeFn := newMockDetector(t, rules)
	defer closeFn()

	event := makeEvent(0.0, 200.0) // OldPrice <= 0 -> checkSpike renvoie nil sans requete

	rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM price_history").WillReturnRows(rows)

	alerts := d.Analyze(context.Background(), event)
	for _, a := range alerts {
		if a.Reason == model.ReasonSuspiciousSpike {
			t.Errorf("checkSpike devrait etre ignore quand OldPrice <= 0")
		}
	}
}

func TestAnalyze_FloodDetectedAtThreshold(t *testing.T) {
	rules := defaultRules()
	d, mock, closeFn := newMockDetector(t, rules)
	defer closeFn()

	event := makeEvent(100.0, 105.0)

	mock.ExpectQuery("SELECT new_price FROM price_history").
		WillReturnError(sql.ErrNoRows)
	rows := sqlmock.NewRows([]string{"count"}).AddRow(rules.FloodMaxUpdates)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM price_history").WillReturnRows(rows)

	alerts := d.Analyze(context.Background(), event)
	found := false
	for _, a := range alerts {
		if a.Reason == model.ReasonFloodPricing {
			found = true
		}
	}
	if !found {
		t.Errorf("attendu une alerte FLOOD_PRICING au seuil, obtenu %+v", alerts)
	}
}

func TestAnalyze_FloodQueryErrorIsIgnored(t *testing.T) {
	rules := defaultRules()
	d, mock, closeFn := newMockDetector(t, rules)
	defer closeFn()

	event := makeEvent(100.0, 105.0)

	mock.ExpectQuery("SELECT new_price FROM price_history").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM price_history").
		WillReturnError(sql.ErrNoRows)

	// Ne doit pas paniquer : une erreur de comptage ne declenche pas d'alerte.
	alerts := d.Analyze(context.Background(), event)
	for _, a := range alerts {
		if a.Reason == model.ReasonFloodPricing {
			t.Errorf("aucune alerte FLOOD_PRICING attendue en cas d'erreur de requete")
		}
	}
}

func TestAnalyze_DumpingDetected(t *testing.T) {
	rules := defaultRules()
	d, mock, closeFn := newMockDetector(t, rules)
	defer closeFn()

	event := makeEvent(100.0, 0.50) // < 1€ -> dumping

	mock.ExpectQuery("SELECT new_price FROM price_history").
		WillReturnError(sql.ErrNoRows)
	rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM price_history").WillReturnRows(rows)

	alerts := d.Analyze(context.Background(), event)
	found := false
	for _, a := range alerts {
		if a.Reason == model.ReasonDumping {
			found = true
		}
	}
	if !found {
		t.Errorf("attendu une alerte DUMPING, obtenu %+v", alerts)
	}
}

func TestAnalyze_NoAlertOnNormalEvent(t *testing.T) {
	rules := defaultRules()
	d, mock, closeFn := newMockDetector(t, rules)
	defer closeFn()

	event := makeEvent(100.0, 110.0) // +10%, prix normal

	mock.ExpectQuery("SELECT new_price FROM price_history").
		WillReturnError(sql.ErrNoRows)
	rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM price_history").WillReturnRows(rows)

	alerts := d.Analyze(context.Background(), event)
	if len(alerts) != 0 {
		t.Errorf("aucune alerte attendue pour un changement de prix normal, obtenu %+v", alerts)
	}
}

// Compile-time check : detector.New doit rester appelable avec un repo nil.
var _ = func() {
	var repo *repository.PriceRepository
	_ = detector.New(repo, defaultRules())
}
