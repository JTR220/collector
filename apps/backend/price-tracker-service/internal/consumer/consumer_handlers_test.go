package consumer

// Tests unitaires de handleMessage (le coeur de la fiabilite de la chaine
// evenementielle catalog -> price-tracker) : payload invalide, idempotence
// (doublon, erreur de verification), erreur de persistance, et detection de
// fraude declenchee ou non. Utilise sqlmock (comme detector_test.go) plutot
// qu'une vraie base, et un Acknowledger factice (comme
// notification-service/internal/consumer/consumer_handlers_test.go) plutot
// qu'un vrai canal AMQP.

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/JTR220/collector/price-tracker-service/config"
	"github.com/JTR220/collector/price-tracker-service/internal/detector"
	"github.com/JTR220/collector/price-tracker-service/internal/model"
	"github.com/JTR220/collector/price-tracker-service/internal/repository"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// fakeAcknowledger enregistre les appels Ack/Nack sans toucher a un vrai
// channel AMQP : handleMessage appelle msg.Ack/msg.Nack sans erreur.
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

func defaultRules() config.DetectionRules {
	return config.DetectionRules{
		SpikeThresholdPercent: 50.0,
		SpikeWindowHours:      24,
		FloodMaxUpdates:       5,
		FloodWindowMinutes:    60,
	}
}

// newTestConsumer cree un PriceConsumer branche sur une DB sqlmock. Le
// Publisher a un channel nil : PublishAlert echoue silencieusement (erreur
// journalisee, sans impact sur l'Ack), ce qui est le comportement voulu —
// une alerte fraude non publiee ne doit pas faire perdre l'evenement prix
// deja persiste (voir TestHandleMessage_AlertDetectedAndPublished).
func newTestConsumer(t *testing.T, rules config.DetectionRules) (*PriceConsumer, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := repository.NewPriceRepository(sqlxDB)
	det := detector.New(repo, rules)
	pub := NewPublisher(nil, "collector.alerts")
	c := NewPriceConsumer(repo, det, pub, &config.Config{})
	return c, mock
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

func makeEvent(oldPrice, newPrice float64) model.PriceUpdatedEvent {
	return model.PriceUpdatedEvent{
		ItemID:    uuid.New(),
		SellerID:  uuid.New(),
		OldPrice:  oldPrice,
		NewPrice:  newPrice,
		UpdatedAt: time.Now(),
	}
}

func TestHandleMessage_InvalidPayload(t *testing.T) {
	c, _ := newTestConsumer(t, defaultRules())
	ack := &fakeAcknowledger{}
	msg := amqp.Delivery{Body: []byte("not json"), Acknowledger: ack}

	c.handleMessage(context.Background(), msg)

	if ack.nacked != 1 || ack.requeue {
		t.Errorf("attendu Nack sans requeue pour un payload invalide, obtenu nack=%d requeue=%v", ack.nacked, ack.requeue)
	}
}

func TestHandleMessage_IdempotenceCheckErrorRequeues(t *testing.T) {
	c, mock := newTestConsumer(t, defaultRules())
	msg, ack := deliveryWithBody(t, makeEvent(100, 110))

	mock.ExpectExec("INSERT INTO processed_messages").WillReturnError(sql.ErrConnDone)

	c.handleMessage(context.Background(), msg)

	if ack.nacked != 1 || !ack.requeue {
		t.Errorf("attendu Nack avec requeue si la verification d'idempotence echoue, obtenu nack=%d requeue=%v", ack.nacked, ack.requeue)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("attentes SQL non satisfaites : %v", err)
	}
}

func TestHandleMessage_DuplicateIsAckedWithoutSave(t *testing.T) {
	c, mock := newTestConsumer(t, defaultRules())
	msg, ack := deliveryWithBody(t, makeEvent(100, 110))

	// RowsAffected=0 : message deja vu (ON CONFLICT DO NOTHING n'a rien insere).
	mock.ExpectExec("INSERT INTO processed_messages").WillReturnResult(sqlmock.NewResult(0, 0))

	c.handleMessage(context.Background(), msg)

	if ack.acked != 1 {
		t.Errorf("un doublon doit etre acquitte sans etre retraite, obtenu %d ack", ack.acked)
	}
	// Si le code retraitait le message a tort, SavePriceHistory executerait une
	// requete sans expectation correspondante : sqlmock la ferait echouer, ce
	// qui se traduirait par un Nack au lieu d'un Ack ci-dessus.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("attentes SQL non satisfaites (le message n'aurait pas du etre retraite) : %v", err)
	}
}

func TestHandleMessage_SavePriceHistoryErrorRequeues(t *testing.T) {
	c, mock := newTestConsumer(t, defaultRules())
	msg, ack := deliveryWithBody(t, makeEvent(100, 110))

	mock.ExpectExec("INSERT INTO processed_messages").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO price_history").WillReturnError(sql.ErrConnDone)

	c.handleMessage(context.Background(), msg)

	if ack.nacked != 1 || !ack.requeue {
		t.Errorf("attendu Nack avec requeue si la persistance de l'historique echoue, obtenu nack=%d requeue=%v", ack.nacked, ack.requeue)
	}
}

func TestHandleMessage_NoAlertSuccess(t *testing.T) {
	c, mock := newTestConsumer(t, defaultRules())
	msg, ack := deliveryWithBody(t, makeEvent(100, 110)) // +10%, sous le seuil de 50%

	mock.ExpectExec("INSERT INTO processed_messages").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO price_history").WillReturnResult(sqlmock.NewResult(1, 1))
	// Analyze : checkSpike (GetLastPrice) puis checkFlood (CountUpdatesInWindow).
	mock.ExpectQuery("SELECT new_price FROM price_history").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM price_history").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	c.handleMessage(context.Background(), msg)

	if ack.acked != 1 || ack.nacked != 0 {
		t.Errorf("attendu 1 Ack sans Nack pour un evenement normal, obtenu ack=%d nack=%d", ack.acked, ack.nacked)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("attentes SQL non satisfaites (INSERT INTO fraud_alerts inattendu ?) : %v", err)
	}
}

func TestHandleMessage_AlertDetectedAndPublished(t *testing.T) {
	c, mock := newTestConsumer(t, defaultRules())
	msg, ack := deliveryWithBody(t, makeEvent(100, 200)) // +100% -> SUSPICIOUS_SPIKE

	mock.ExpectExec("INSERT INTO processed_messages").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO price_history").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("SELECT new_price FROM price_history").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM price_history").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectExec("INSERT INTO fraud_alerts").WillReturnResult(sqlmock.NewResult(1, 1))

	c.handleMessage(context.Background(), msg)

	// Le Publisher a un channel nil dans ce test (voir newTestConsumer) :
	// PublishAlert echoue forcement. L'evenement doit malgre tout etre
	// acquitte, car l'historique de prix et l'alerte sont deja persistes en
	// base — seule la notification temps reel de l'alerte est perdue, ce qui
	// est prefere a un retraitement complet (double insertion) de l'evenement.
	if ack.acked != 1 || ack.nacked != 0 {
		t.Errorf("attendu 1 Ack meme si la publication de l'alerte echoue, obtenu ack=%d nack=%d", ack.acked, ack.nacked)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("attentes SQL non satisfaites (fraud_alerts non sauvegarde ?) : %v", err)
	}
}

func TestHandleMessage_MissingMessageIDIsDerivedFromBody(t *testing.T) {
	// Sans MessageId AMQP, handleMessage derive une cle stable (sha256 du
	// corps) : deux messages identiques sans MessageId doivent donc etre
	// traites comme le meme evenement (idempotence par contenu).
	c, mock := newTestConsumer(t, defaultRules())
	event := makeEvent(100, 110)
	body, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	ack := &fakeAcknowledger{}
	msg := amqp.Delivery{Body: body, Acknowledger: ack} // pas de MessageId

	mock.ExpectExec("INSERT INTO processed_messages").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO price_history").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("SELECT new_price FROM price_history").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM price_history").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	c.handleMessage(context.Background(), msg)

	if ack.acked != 1 {
		t.Errorf("attendu 1 Ack pour un message sans MessageId (cle derivee du contenu), obtenu %d", ack.acked)
	}
}
