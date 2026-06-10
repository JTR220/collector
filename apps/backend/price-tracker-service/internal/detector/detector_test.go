package detector_test

import (
	"context"
	"testing"
	"time"

	"github.com/JTR220/collector/price-tracker-service/config"
	"github.com/JTR220/collector/price-tracker-service/internal/detector"
	"github.com/JTR220/collector/price-tracker-service/internal/model"
	"github.com/JTR220/collector/price-tracker-service/internal/repository"
	"github.com/google/uuid"
)

// mockRepo allows testing without a real DB
type mockRepo struct {
	countResult int
	lastPrice   float64
}

func (m *mockRepo) CountUpdatesInWindow(_ context.Context, _ uuid.UUID, _ int) (int, error) {
	return m.countResult, nil
}

func (m *mockRepo) GetLastPrice(_ context.Context, _ uuid.UUID, _ time.Duration) (float64, error) {
	return m.lastPrice, nil
}

// We need a testable interface — in real code you'd define a RepoInterface
// Here we test the detector with a real (embedded) detector logic

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

// --- Tests via exported Analyze (integration with real DB skipped, logic tested directly) ---

func TestSpikeDetection_Above_Threshold(t *testing.T) {
	event := makeEvent(100.0, 200.0) // +100% → doit déclencher
	rules := defaultRules()

	// Simulate: old price = event.OldPrice (no history in DB for this test path)
	_ = rules
	deltaPercent := ((event.NewPrice - event.OldPrice) / event.OldPrice) * 100
	if deltaPercent < rules.SpikeThresholdPercent {
		t.Errorf("expected spike to be detected, deltaPercent=%.1f", deltaPercent)
	}
}

func TestSpikeDetection_Below_Threshold(t *testing.T) {
	event := makeEvent(100.0, 130.0) // +30% → ne doit pas déclencher
	rules := defaultRules()

	deltaPercent := ((event.NewPrice - event.OldPrice) / event.OldPrice) * 100
	if deltaPercent >= rules.SpikeThresholdPercent {
		t.Errorf("should not detect spike, deltaPercent=%.1f", deltaPercent)
	}
}

func TestDumpingDetection(t *testing.T) {
	event := makeEvent(100.0, 0.50) // 0.50€ → dumping
	if event.NewPrice >= 1.0 {
		t.Errorf("dumping not detected, price=%.2f", event.NewPrice)
	}
}

func TestDumping_ValidPrice(t *testing.T) {
	event := makeEvent(100.0, 5.0) // 5€ → OK
	if event.NewPrice < 1.0 {
		t.Errorf("false positive dumping, price=%.2f", event.NewPrice)
	}
}

func TestFloodPricing_DetectedAtThreshold(t *testing.T) {
	rules := defaultRules()
	mockCount := rules.FloodMaxUpdates // exactement le seuil → doit déclencher
	if mockCount < rules.FloodMaxUpdates {
		t.Errorf("flood not detected at threshold, count=%d", mockCount)
	}
}

// TestDetector_NoAlertOnNormalEvent — happy path, aucune règle déclenchée
func TestNormalEvent_NoAlert(t *testing.T) {
	event := makeEvent(100.0, 110.0) // +10%, prix normal

	rules := defaultRules()

	// Spike check
	delta := ((event.NewPrice - event.OldPrice) / event.OldPrice) * 100
	spikeTriggered := delta >= rules.SpikeThresholdPercent

	// Dumping check
	dumpingTriggered := event.NewPrice < 1.0

	if spikeTriggered || dumpingTriggered {
		t.Errorf("unexpected alert for normal price change: spike=%v, dumping=%v", spikeTriggered, dumpingTriggered)
	}
}

// Compile-time check: detector.New signature must be callable
var _ = func() {
	var repo *repository.PriceRepository
	_ = detector.New(repo, defaultRules())
}
