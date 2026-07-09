package consumer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/JTR220/collector/price-tracker-service/config"
	"github.com/JTR220/collector/price-tracker-service/internal/detector"
	"github.com/JTR220/collector/price-tracker-service/internal/metrics"
	"github.com/JTR220/collector/price-tracker-service/internal/model"
	"github.com/JTR220/collector/price-tracker-service/internal/repository"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

// MessageStore tracks which message IDs have already been processed, to make
// consumption idempotent across redeliveries (e.g. after a nack/requeue or a
// consumer restart before the ack reached the broker). Satisfied by
// *repository.PriceRepository; a fake is used in tests.
type MessageStore interface {
	MarkProcessed(ctx context.Context, messageID string) (bool, error)
}

// Publisher sends fraud alerts to RabbitMQ. Le channel sous-jacent est
// remplace apres chaque reconnexion (voir PriceConsumer.Start), d'ou le mutex.
type Publisher struct {
	mu       sync.RWMutex
	ch       *amqp.Channel
	exchange string
}

func NewPublisher(ch *amqp.Channel, exchange string) *Publisher {
	return &Publisher{ch: ch, exchange: exchange}
}

// SetChannel remplace le channel AMQP utilise pour publier, typiquement
// apres une reconnexion au broker.
func (p *Publisher) SetChannel(ch *amqp.Channel) {
	p.mu.Lock()
	p.ch = ch
	p.mu.Unlock()
}

func (p *Publisher) PublishAlert(alert model.FraudAlertEvent) error {
	p.mu.RLock()
	ch := p.ch
	p.mu.RUnlock()
	if ch == nil {
		metrics.RecordRabbitMQError("publish_alert")
		return fmt.Errorf("publisher: pas de channel AMQP disponible (broker deconnecte)")
	}

	body, err := json.Marshal(alert)
	if err != nil {
		metrics.RecordRabbitMQError("marshal_alert")
		return err
	}
	if err := ch.Publish(
		p.exchange,    // exchange
		"fraud.alert", // routing key
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			MessageId:    alert.AlertID.String(),
			Timestamp:    time.Now(),
			Body:         body,
		},
	); err != nil {
		metrics.RecordRabbitMQError("publish_alert")
		return err
	}
	return nil
}

// PriceConsumer listens to price.updated events from catalog-service
type PriceConsumer struct {
	repo      *repository.PriceRepository
	store     MessageStore
	detector  *detector.Detector
	publisher *Publisher
	cfg       *config.Config
}

func NewPriceConsumer(
	repo *repository.PriceRepository,
	det *detector.Detector,
	pub *Publisher,
	cfg *config.Config,
) *PriceConsumer {
	return &PriceConsumer{
		repo:      repo,
		store:     repo,
		detector:  det,
		publisher: pub,
		cfg:       cfg,
	}
}

// Setup declares exchange and queue, then binds them
func Setup(ch *amqp.Channel, cfg *config.RabbitMQConfig) error {
	// Exchange events (catalog-service publie ici)
	if err := ch.ExchangeDeclare(cfg.ExchangeEvents, "topic", true, false, false, false, nil); err != nil {
		return err
	}

	// Exchange alerts (price-tracker publie ici pour notification-service)
	if err := ch.ExchangeDeclare(cfg.ExchangeAlerts, "topic", true, false, false, false, nil); err != nil {
		return err
	}

	// Queue dédiée au price-tracker
	q, err := ch.QueueDeclare(cfg.QueuePriceUpdate, true, false, false, false, nil)
	if err != nil {
		return err
	}

	// Bind : écoute price.updated sur l'exchange events
	return ch.QueueBind(q.Name, "price.updated", cfg.ExchangeEvents, false, nil)
}

// Start consomme les messages en boucle jusqu'a annulation du contexte. Si la
// connexion au broker est perdue en cours de route, elle est retablie avec un
// backoff exponentiel (le service ne s'arrete pas sur une coupure RabbitMQ
// transitoire) : reconnexion, re-declaration des exchanges/queue (Setup) et
// re-attachement du publisher d'alertes fraude au nouveau channel.
func (c *PriceConsumer) Start(ctx context.Context, url string) error {
	backoff := time.Second
	for {
		if ctx.Err() != nil {
			return nil
		}

		conn, ch, err := dialAndSetup(url, &c.cfg.RabbitMQ)
		if err != nil {
			metrics.RecordRabbitMQError("connect")
			log.Error().Err(err).Msg("connexion RabbitMQ (consumer) echouee, nouvel essai")
			if !sleepOrDone(ctx, backoff) {
				return nil
			}
			backoff = nextBackoff(backoff)
			continue
		}
		backoff = time.Second

		if c.publisher != nil {
			c.publisher.SetChannel(ch)
		}

		closeCh := conn.NotifyClose(make(chan *amqp.Error, 1))
		lost := c.consumeUntilClosed(ctx, ch, closeCh)
		_ = ch.Close()
		_ = conn.Close()

		if !lost {
			log.Info().Msg("consumer stopped")
			return nil
		}
		log.Warn().Msg("connexion RabbitMQ perdue (consumer), reconnexion...")
	}
}

// consumeUntilClosed consomme jusqu'a annulation du contexte (retourne false)
// ou perte de connexion (retourne true, pour declencher une reconnexion).
func (c *PriceConsumer) consumeUntilClosed(ctx context.Context, ch *amqp.Channel, closeCh <-chan *amqp.Error) bool {
	msgs, err := ch.Consume(
		c.cfg.RabbitMQ.QueuePriceUpdate,
		"price-tracker-consumer",
		false, // manual ack
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		metrics.RecordRabbitMQError("consume")
		log.Error().Err(err).Msg("impossible de demarrer la consommation")
		return true
	}

	log.Info().Msg("price-tracker consumer started, waiting for price.updated events...")

	for {
		select {
		case <-ctx.Done():
			return false
		case <-closeCh:
			return true
		case msg, ok := <-msgs:
			if !ok {
				return true
			}
			c.handleMessage(ctx, msg)
		}
	}
}

func dialAndSetup(url string, cfg *config.RabbitMQConfig) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, nil, err
	}
	if err := Setup(ch, cfg); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, nil, err
	}
	return conn, ch, nil
}

func sleepOrDone(ctx context.Context, d time.Duration) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(d):
		return true
	}
}

func nextBackoff(d time.Duration) time.Duration {
	if d >= 30*time.Second {
		return 30 * time.Second
	}
	return d * 2
}

func (c *PriceConsumer) handleMessage(ctx context.Context, msg amqp.Delivery) {
	log.Debug().RawJSON("body", msg.Body).Msg("received price.updated")

	var event model.PriceUpdatedEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		metrics.RecordPriceEvent("invalid_payload")
		log.Error().Err(err).Msg("failed to unmarshal price event — nacking")
		_ = msg.Nack(false, false)
		return
	}

	// Idempotence : un message redistribue (nack/requeue, restart consumer
	// avant l'ack) ne doit pas re-declencher persistance/detection/publication.
	// A defaut de MessageId AMQP, on derive une cle stable du contenu.
	messageID := msg.MessageId
	if messageID == "" {
		sum := sha256.Sum256(msg.Body)
		messageID = fmt.Sprintf("price.updated:%s", hex.EncodeToString(sum[:]))
	}
	firstSeen, err := c.store.MarkProcessed(ctx, messageID)
	if err != nil {
		metrics.RecordPriceEvent("idempotence_error")
		log.Error().Err(err).Msg("failed to check message idempotence — requeuing")
		_ = msg.Nack(false, true)
		return
	}
	if !firstSeen {
		metrics.RecordPriceEvent("duplicate")
		log.Info().Str("message_id", messageID).Msg("duplicate price.updated ignored")
		_ = msg.Ack(false)
		return
	}

	// 1. Persiste l'historique de prix
	history := &model.PriceHistory{
		ID:        uuid.New(),
		ItemID:    event.ItemID,
		SellerID:  event.SellerID,
		OldPrice:  event.OldPrice,
		NewPrice:  event.NewPrice,
		CreatedAt: event.UpdatedAt,
	}
	if err := c.repo.SavePriceHistory(ctx, history); err != nil {
		metrics.RecordPriceEvent("save_error")
		log.Error().Err(err).Msg("failed to save price history — requeuing")
		_ = msg.Nack(false, true)
		return
	}

	// 2. Lance les règles de détection
	alerts := c.detector.Analyze(ctx, &event)

	for _, alert := range alerts {
		// Persiste l'alerte en BDD
		if err := c.repo.SaveAlert(ctx, &alert); err != nil {
			metrics.RecordRabbitMQError("save_alert")
			log.Error().Err(err).Str("reason", string(alert.Reason)).Msg("failed to save fraud alert")
			continue
		}
		metrics.RecordFraudAlert(string(alert.Reason))

		// Publie l'alerte sur collector.alerts pour notification-service
		alertEvent := model.FraudAlertEvent{
			AlertID:   alert.ID,
			ItemID:    alert.ItemID,
			SellerID:  alert.SellerID,
			Reason:    alert.Reason,
			Detail:    alert.Detail,
			OldPrice:  alert.OldPrice,
			NewPrice:  alert.NewPrice,
			CreatedAt: alert.CreatedAt,
		}
		if err := c.publisher.PublishAlert(alertEvent); err != nil {
			log.Error().Err(err).Msg("failed to publish fraud alert event")
		} else {
			log.Info().
				Str("alert_id", alert.ID.String()).
				Str("reason", string(alert.Reason)).
				Msg("fraud alert published")
		}
	}

	metrics.RecordPriceEvent("success")
	_ = msg.Ack(false)
}
