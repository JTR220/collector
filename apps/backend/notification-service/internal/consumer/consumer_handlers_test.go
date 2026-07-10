package consumer

// Tests unitaires des handlers metier (handlePriceEvent, handleFraudAlert,
// handleOrderCreated, handleOrderDecided) via une DB sqlmock et un
// Acknowledger factice : complete consumer_test.go qui ne testait que les
// fonctions utilitaires pures (messageIDOf, nextBackoff).

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/JTR220/collector/notification-service/config"
	"github.com/JTR220/collector/notification-service/internal/hub"
	"github.com/JTR220/collector/notification-service/internal/mailer"
	"github.com/JTR220/collector/notification-service/internal/model"
	"github.com/JTR220/collector/notification-service/internal/repository"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// fakeAcknowledger enregistre les appels Ack/Nack/Reject sans toucher a un
// vrai channel AMQP : les handlers appellent msg.Ack/msg.Nack sans erreur.
type fakeAcknowledger struct {
	acked   int
	nacked  int
	requeue bool
}

func (f *fakeAcknowledger) Ack(tag uint64, multiple bool) error { f.acked++; return nil }
func (f *fakeAcknowledger) Nack(tag uint64, multiple, requeue bool) error {
	f.nacked++
	f.requeue = requeue
	return nil
}
func (f *fakeAcknowledger) Reject(tag uint64, requeue bool) error { return nil }

func newTestManager(t *testing.T) (*Manager, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := repository.New(sqlxDB)
	h := hub.New()
	go h.Run()
	m := NewManager(h, repo, &config.Config{}, mailer.NoopMailer{}, nil)
	return m, mock
}

func deliveryWithBody(t *testing.T, v interface{}) (amqp.Delivery, *fakeAcknowledger) {
	t.Helper()
	body, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	ack := &fakeAcknowledger{}
	return amqp.Delivery{Body: body, Acknowledger: ack, MessageId: uuid.NewString()}, ack
}

func TestHandlePriceEvent_InvalidPayload(t *testing.T) {
	m, _ := newTestManager(t)
	ack := &fakeAcknowledger{}
	msg := amqp.Delivery{Body: []byte("not json"), Acknowledger: ack}
	m.handlePriceEvent(t.Context(), msg)
	if ack.nacked != 1 {
		t.Errorf("attendu 1 Nack pour un payload invalide, obtenu %d", ack.nacked)
	}
}

func TestHandlePriceEvent_SpikeSavesAndBroadcasts(t *testing.T) {
	m, mock := newTestManager(t)
	event := model.PriceUpdatedEvent{
		ItemID: uuid.New(), SellerID: uuid.New(),
		OldPrice: 100, NewPrice: 150, UpdatedAt: time.Now(),
	}
	msg, ack := deliveryWithBody(t, event)

	mock.ExpectExec("INSERT INTO processed_messages").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO notifications").WillReturnResult(sqlmock.NewResult(1, 1))

	m.handlePriceEvent(t.Context(), msg)

	if ack.acked != 1 {
		t.Errorf("attendu 1 Ack, obtenu %d", ack.acked)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("attentes SQL non satisfaites : %v", err)
	}
}

func TestHandlePriceEvent_DropDetected(t *testing.T) {
	m, mock := newTestManager(t)
	event := model.PriceUpdatedEvent{
		ItemID: uuid.New(), SellerID: uuid.New(),
		OldPrice: 100, NewPrice: 80, UpdatedAt: time.Now(),
	}
	msg, ack := deliveryWithBody(t, event)

	mock.ExpectExec("INSERT INTO processed_messages").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO notifications").WillReturnResult(sqlmock.NewResult(1, 1))

	m.handlePriceEvent(t.Context(), msg)
	if ack.acked != 1 {
		t.Errorf("attendu 1 Ack, obtenu %d", ack.acked)
	}
}

func TestHandlePriceEvent_DuplicateIsAckedWithoutSave(t *testing.T) {
	m, mock := newTestManager(t)
	event := model.PriceUpdatedEvent{ItemID: uuid.New(), SellerID: uuid.New(), OldPrice: 100, NewPrice: 90}
	msg, ack := deliveryWithBody(t, event)

	mock.ExpectExec("INSERT INTO processed_messages").WillReturnResult(sqlmock.NewResult(0, 0)) // deja vu

	m.handlePriceEvent(t.Context(), msg)
	if ack.acked != 1 {
		t.Errorf("un message duplique doit quand meme etre acquitte, obtenu %d ack", ack.acked)
	}
}

func TestHandlePriceEvent_IdempotenceCheckErrorRequeues(t *testing.T) {
	m, mock := newTestManager(t)
	event := model.PriceUpdatedEvent{ItemID: uuid.New(), SellerID: uuid.New(), OldPrice: 100, NewPrice: 90}
	msg, ack := deliveryWithBody(t, event)

	mock.ExpectExec("INSERT INTO processed_messages").WillReturnError(sql.ErrConnDone)

	m.handlePriceEvent(t.Context(), msg)
	if ack.nacked != 1 || !ack.requeue {
		t.Errorf("attendu Nack avec requeue, obtenu nack=%d requeue=%v", ack.nacked, ack.requeue)
	}
}

func TestHandleFraudAlert_InvalidPayload(t *testing.T) {
	m, _ := newTestManager(t)
	ack := &fakeAcknowledger{}
	msg := amqp.Delivery{Body: []byte("bad"), Acknowledger: ack}
	m.handleFraudAlert(t.Context(), msg)
	if ack.nacked != 1 {
		t.Errorf("attendu 1 Nack, obtenu %d", ack.nacked)
	}
}

func TestHandleFraudAlert_SavesAndBroadcasts(t *testing.T) {
	m, mock := newTestManager(t)
	event := model.FraudAlertEvent{
		AlertID: uuid.New(), ItemID: uuid.New(), SellerID: uuid.New(),
		Reason: "SUSPICIOUS_SPIKE", Detail: "test", OldPrice: 10, NewPrice: 100,
	}
	msg, ack := deliveryWithBody(t, event)

	mock.ExpectExec("INSERT INTO processed_messages").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO notifications").WillReturnResult(sqlmock.NewResult(1, 1))

	m.handleFraudAlert(t.Context(), msg)
	if ack.acked != 1 {
		t.Errorf("attendu 1 Ack, obtenu %d", ack.acked)
	}
}

func TestHandleOrderCreated_InvalidPayload(t *testing.T) {
	m, _ := newTestManager(t)
	ack := &fakeAcknowledger{}
	msg := amqp.Delivery{Body: []byte("bad"), Acknowledger: ack}
	m.handleOrderCreated(t.Context(), msg)
	if ack.nacked != 1 {
		t.Errorf("attendu 1 Nack, obtenu %d", ack.nacked)
	}
}

func TestHandleOrderCreated_SavesNotifiesWithoutEmail(t *testing.T) {
	// auth=nil dans newTestManager -> notifyAndEmail ne tente pas d'email.
	m, mock := newTestManager(t)
	event := model.OrderCreatedEvent{
		OrderID: uuid.New(), ItemID: uuid.New(), ItemName: "Charizard PSA9",
		BuyerID: uuid.New(), SellerID: uuid.New(), Price: 250,
	}
	msg, ack := deliveryWithBody(t, event)

	mock.ExpectExec("INSERT INTO processed_messages").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO notifications").WillReturnResult(sqlmock.NewResult(1, 1))

	m.handleOrderCreated(t.Context(), msg)
	if ack.acked != 1 {
		t.Errorf("attendu 1 Ack, obtenu %d", ack.acked)
	}
}

func TestHandleOrderDecided_Accepted(t *testing.T) {
	m, mock := newTestManager(t)
	event := model.OrderDecisionEvent{
		OrderID: uuid.New(), ItemID: uuid.New(), ItemName: "Charizard PSA9",
		BuyerID: uuid.New(), SellerID: uuid.New(), Price: 250, Accepted: true,
	}
	msg, ack := deliveryWithBody(t, event)

	mock.ExpectExec("INSERT INTO processed_messages").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO notifications").WillReturnResult(sqlmock.NewResult(1, 1))

	m.handleOrderDecided(t.Context(), msg)
	if ack.acked != 1 {
		t.Errorf("attendu 1 Ack, obtenu %d", ack.acked)
	}
}

func TestHandleOrderDecided_Rejected(t *testing.T) {
	m, mock := newTestManager(t)
	event := model.OrderDecisionEvent{
		OrderID: uuid.New(), ItemID: uuid.New(), ItemName: "Charizard PSA9",
		BuyerID: uuid.New(), SellerID: uuid.New(), Price: 250, Accepted: false,
	}
	msg, ack := deliveryWithBody(t, event)

	mock.ExpectExec("INSERT INTO processed_messages").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO notifications").WillReturnResult(sqlmock.NewResult(1, 1))

	m.handleOrderDecided(t.Context(), msg)
	if ack.acked != 1 {
		t.Errorf("attendu 1 Ack, obtenu %d", ack.acked)
	}
}

func TestHandleOrderDecided_DuplicateIgnored(t *testing.T) {
	m, mock := newTestManager(t)
	event := model.OrderDecisionEvent{OrderID: uuid.New(), ItemID: uuid.New(), BuyerID: uuid.New(), SellerID: uuid.New()}
	msg, ack := deliveryWithBody(t, event)

	mock.ExpectExec("INSERT INTO processed_messages").WillReturnResult(sqlmock.NewResult(0, 0))

	m.handleOrderDecided(t.Context(), msg)
	if ack.acked != 1 {
		t.Errorf("un doublon doit quand meme etre acquitte, obtenu %d ack", ack.acked)
	}
}
