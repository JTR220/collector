package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/JTR220/collector/notification-service/config"
	"github.com/JTR220/collector/notification-service/internal/hub"
	"github.com/JTR220/collector/notification-service/internal/model"
	"github.com/JTR220/collector/notification-service/internal/repository"
)

// Setup declares all exchanges and queues needed by notification-service
func Setup(ch *amqp.Channel, cfg *config.RabbitMQConfig) error {
	// Listen on both exchanges
	for _, exchange := range []string{cfg.ExchangeEvents, cfg.ExchangeAlerts} {
		if err := ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
			return fmt.Errorf("declare exchange %s: %w", exchange, err)
		}
	}

	// Queue: price updates
	qPrice, err := ch.QueueDeclare(cfg.QueuePriceNotif, true, false, false, false, nil)
	if err != nil {
		return err
	}
	if err := ch.QueueBind(qPrice.Name, "price.updated", cfg.ExchangeEvents, false, nil); err != nil {
		return err
	}

	// Queue: fraud alerts
	qFraud, err := ch.QueueDeclare(cfg.QueueFraudNotif, true, false, false, false, nil)
	if err != nil {
		return err
	}
	if err := ch.QueueBind(qFraud.Name, "fraud.alert", cfg.ExchangeAlerts, false, nil); err != nil {
		return err
	}

	return nil
}

// Manager runs all consumers concurrently
type Manager struct {
	ch   *amqp.Channel
	hub  *hub.Hub
	repo *repository.NotificationRepository
	cfg  *config.Config
}

func NewManager(ch *amqp.Channel, h *hub.Hub, repo *repository.NotificationRepository, cfg *config.Config) *Manager {
	return &Manager{ch: ch, hub: h, repo: repo, cfg: cfg}
}

// Start launches all consumers — blocks until ctx is cancelled
func (m *Manager) Start(ctx context.Context) {
	go m.consumePriceUpdates(ctx)
	go m.consumeFraudAlerts(ctx)
	<-ctx.Done()
}

// ── Price Updates Consumer ───────────────────────────────────────────────────

func (m *Manager) consumePriceUpdates(ctx context.Context) {
	msgs, err := m.ch.Consume(m.cfg.RabbitMQ.QueuePriceNotif, "notif-price-consumer", false, false, false, false, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start price consumer")
	}
	log.Info().Msg("price update consumer started")

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-msgs:
			if !ok {
				return
			}
			m.handlePriceEvent(ctx, msg)
		}
	}
}

func (m *Manager) handlePriceEvent(ctx context.Context, msg amqp.Delivery) {
	var event model.PriceUpdatedEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Error().Err(err).Msg("invalid price event payload")
		_ = msg.Nack(false, false)
		return
	}

	// Determine notification type
	notifType := model.TypePriceSpike
	title := "⬆️ Hausse de prix"
	if event.NewPrice < event.OldPrice {
		notifType = model.TypePriceDrop
		title = "⬇️ Baisse de prix"
	}

	delta := event.NewPrice - event.OldPrice
	sign := "+"
	if delta < 0 {
		sign = ""
	}
	body := fmt.Sprintf("Article %s : %s%.2f€ (%.2f€ → %.2f€)",
		event.ItemID.String()[:8], sign, delta, event.OldPrice, event.NewPrice)

	// In a real app you'd query which users watch this item
	// Here we broadcast to ALL connected clients (demo-friendly)
	notif := &model.Notification{
		ID:        uuid.New(),
		UserID:    event.SellerID, // notify seller
		Type:      notifType,
		Title:     title,
		Body:      body,
		ItemID:    &event.ItemID,
		Read:      false,
		CreatedAt: time.Now(),
	}

	if err := m.repo.Save(ctx, notif); err != nil {
		log.Error().Err(err).Msg("failed to persist price notification")
	}

	// Build WebSocket message
	wsMsg := model.WebSocketMessage{
		Event: string(notifType),
		Data: map[string]interface{}{
			"notification_id": notif.ID,
			"item_id":         event.ItemID,
			"old_price":       event.OldPrice,
			"new_price":       event.NewPrice,
			"title":           title,
			"body":            body,
			"created_at":      notif.CreatedAt,
		},
	}
	payload, _ := json.Marshal(wsMsg)

	// Push to seller in real-time
	m.hub.SendToUser(event.SellerID, payload)
	// Also broadcast to all for demo purposes
	m.hub.SendToAll(payload)

	log.Info().
		Str("type", string(notifType)).
		Str("item_id", event.ItemID.String()).
		Msg("price notification dispatched")

	_ = msg.Ack(false)
}

// ── Fraud Alerts Consumer ────────────────────────────────────────────────────

func (m *Manager) consumeFraudAlerts(ctx context.Context) {
	msgs, err := m.ch.Consume(m.cfg.RabbitMQ.QueueFraudNotif, "notif-fraud-consumer", false, false, false, false, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start fraud consumer")
	}
	log.Info().Msg("fraud alert consumer started")

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-msgs:
			if !ok {
				return
			}
			m.handleFraudAlert(ctx, msg)
		}
	}
}

func (m *Manager) handleFraudAlert(ctx context.Context, msg amqp.Delivery) {
	var event model.FraudAlertEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Error().Err(err).Msg("invalid fraud alert payload")
		_ = msg.Nack(false, false)
		return
	}

	// Alerte pour l'admin — dans un vrai système on notifie le rôle admin
	adminUserID := uuid.MustParse("00000000-0000-0000-0000-000000000001") // placeholder admin UUID

	notif := &model.Notification{
		ID:        uuid.New(),
		UserID:    adminUserID,
		Type:      model.TypeFraudAlert,
		Title:     fmt.Sprintf("🚨 Fraude détectée : %s", event.Reason),
		Body:      fmt.Sprintf("Article %s — %s (%.2f€ → %.2f€)", event.ItemID.String()[:8], event.Detail, event.OldPrice, event.NewPrice),
		ItemID:    &event.ItemID,
		Read:      false,
		CreatedAt: time.Now(),
	}

	if err := m.repo.Save(ctx, notif); err != nil {
		log.Error().Err(err).Msg("failed to persist fraud notification")
	}

	wsMsg := model.WebSocketMessage{
		Event: "FRAUD_ALERT",
		Data: map[string]interface{}{
			"notification_id": notif.ID,
			"alert_id":        event.AlertID,
			"item_id":         event.ItemID,
			"seller_id":       event.SellerID,
			"reason":          event.Reason,
			"detail":          event.Detail,
			"old_price":       event.OldPrice,
			"new_price":       event.NewPrice,
			"title":           notif.Title,
			"body":            notif.Body,
			"created_at":      notif.CreatedAt,
		},
	}
	payload, _ := json.Marshal(wsMsg)

	// Broadcast fraud alerts to all admin connections
	m.hub.SendToAll(payload)

	log.Warn().
		Str("alert_id", event.AlertID.String()).
		Str("reason", event.Reason).
		Msg("fraud alert notification dispatched")

	_ = msg.Ack(false)
}
