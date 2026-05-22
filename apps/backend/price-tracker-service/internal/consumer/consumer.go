package consumer

import (
	"context"
	"encoding/json"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/JTR220/collector/price-tracker-service/config"
	"github.com/JTR220/collector/price-tracker-service/internal/detector"
	"github.com/JTR220/collector/price-tracker-service/internal/model"
	"github.com/JTR220/collector/price-tracker-service/internal/repository"
)

// Publisher sends fraud alerts to RabbitMQ
type Publisher struct {
	ch       *amqp.Channel
	exchange string
}

func NewPublisher(ch *amqp.Channel, exchange string) *Publisher {
	return &Publisher{ch: ch, exchange: exchange}
}

func (p *Publisher) PublishAlert(alert model.FraudAlertEvent) error {
	body, err := json.Marshal(alert)
	if err != nil {
		return err
	}
	return p.ch.Publish(
		p.exchange, // exchange
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
	)
}

// PriceConsumer listens to price.updated events from catalog-service
type PriceConsumer struct {
	repo      *repository.PriceRepository
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

// Start begins consuming messages — blocks until context is cancelled
func (c *PriceConsumer) Start(ctx context.Context, ch *amqp.Channel) error {
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
		return err
	}

	log.Info().Msg("price-tracker consumer started, waiting for price.updated events...")

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("consumer stopped")
			return nil
		case msg, ok := <-msgs:
			if !ok {
				log.Warn().Msg("RabbitMQ channel closed")
				return nil
			}
			c.handleMessage(ctx, msg)
		}
	}
}

func (c *PriceConsumer) handleMessage(ctx context.Context, msg amqp.Delivery) {
	log.Debug().RawJSON("body", msg.Body).Msg("received price.updated")

	var event model.PriceUpdatedEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Error().Err(err).Msg("failed to unmarshal price event — nacking")
		_ = msg.Nack(false, false)
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
		log.Error().Err(err).Msg("failed to save price history — requeuing")
		_ = msg.Nack(false, true)
		return
	}

	// 2. Lance les règles de détection
	alerts := c.detector.Analyze(ctx, &event)

	for _, alert := range alerts {
		// Persiste l'alerte en BDD
		if err := c.repo.SaveAlert(ctx, &alert); err != nil {
			log.Error().Err(err).Str("reason", string(alert.Reason)).Msg("failed to save fraud alert")
			continue
		}

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

	_ = msg.Ack(false)
}
