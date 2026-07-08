package consumer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"

	"github.com/JTR220/collector/notification-service/config"
	"github.com/JTR220/collector/notification-service/internal/authclient"
	"github.com/JTR220/collector/notification-service/internal/hub"
	"github.com/JTR220/collector/notification-service/internal/idconv"
	"github.com/JTR220/collector/notification-service/internal/mailer"
	"github.com/JTR220/collector/notification-service/internal/model"
	"github.com/JTR220/collector/notification-service/internal/repository"
)

// messageIDOf renvoie un identifiant stable pour deduplicquer une livraison
// AMQP redistribuee : le MessageId pose par le publisher si present, sinon
// un hash du corps (compatibilite avec un publisher qui n'en fixe pas).
func messageIDOf(msg amqp.Delivery, prefix string) string {
	if msg.MessageId != "" {
		return msg.MessageId
	}
	sum := sha256.Sum256(msg.Body)
	return fmt.Sprintf("%s:%s", prefix, hex.EncodeToString(sum[:]))
}

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

	// Queue: commandes (achat en attente de validation vendeur + decision)
	qOrders, err := ch.QueueDeclare("notification-service.order.events", true, false, false, false, nil)
	if err != nil {
		return err
	}
	if err := ch.QueueBind(qOrders.Name, "order.created", cfg.ExchangeEvents, false, nil); err != nil {
		return err
	}
	if err := ch.QueueBind(qOrders.Name, "order.decided", cfg.ExchangeEvents, false, nil); err != nil {
		return err
	}

	return nil
}

// Manager runs all consumers concurrently
type Manager struct {
	hub  *hub.Hub
	repo *repository.NotificationRepository
	cfg  *config.Config
	mail mailer.Mailer
	auth *authclient.Client
}

func NewManager(h *hub.Hub, repo *repository.NotificationRepository, cfg *config.Config, mail mailer.Mailer, auth *authclient.Client) *Manager {
	return &Manager{hub: h, repo: repo, cfg: cfg, mail: mail, auth: auth}
}

// Start consomme les deux files (price/fraud) jusqu'a annulation du contexte.
// Si la connexion au broker est perdue en cours de route, elle est retablie
// avec un backoff exponentiel : reconnexion, re-declaration des exchanges/
// queues (Setup), puis relance des deux consumers sur le nouveau channel.
func (m *Manager) Start(ctx context.Context, url string) {
	backoff := time.Second
	for ctx.Err() == nil {
		conn, ch, err := dialAndSetup(url, &m.cfg.RabbitMQ)
		if err != nil {
			log.Error().Err(err).Msg("connexion RabbitMQ (consumer) echouee, nouvel essai")
			if !sleepOrDone(ctx, backoff) {
				return
			}
			backoff = nextBackoff(backoff)
			continue
		}
		backoff = time.Second

		lost := m.runUntilClosed(ctx, ch, conn.NotifyClose(make(chan *amqp.Error, 1)))
		_ = ch.Close()
		_ = conn.Close()

		if !lost {
			return
		}
		log.Warn().Msg("connexion RabbitMQ perdue (consumer), reconnexion...")
	}
}

// runUntilClosed lance les deux consumers sur ch et attend soit l'annulation
// du contexte parent (retourne false), soit la perte du channel/connexion
// (retourne true, pour declencher une reconnexion dans Start).
func (m *Manager) runUntilClosed(ctx context.Context, ch *amqp.Channel, closeCh <-chan *amqp.Error) bool {
	cycleCtx, cancelCycle := context.WithCancel(ctx)
	defer cancelCycle()

	go func() {
		select {
		case <-closeCh:
			cancelCycle()
		case <-cycleCtx.Done():
		}
	}()

	var wg sync.WaitGroup
	wg.Add(3)
	go func() { defer wg.Done(); m.consumePriceUpdates(cycleCtx, ch) }()
	go func() { defer wg.Done(); m.consumeFraudAlerts(cycleCtx, ch) }()
	go func() { defer wg.Done(); m.consumeOrderEvents(cycleCtx, ch) }()
	wg.Wait()

	return ctx.Err() == nil
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

// ── Price Updates Consumer ───────────────────────────────────────────────────

func (m *Manager) consumePriceUpdates(ctx context.Context, ch *amqp.Channel) {
	msgs, err := ch.Consume(m.cfg.RabbitMQ.QueuePriceNotif, "notif-price-consumer", false, false, false, false, nil)
	if err != nil {
		log.Error().Err(err).Msg("failed to start price consumer")
		return
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

	firstSeen, err := m.repo.MarkProcessed(ctx, messageIDOf(msg, "price.updated"))
	if err != nil {
		log.Error().Err(err).Msg("failed to check message idempotence — requeuing")
		_ = msg.Nack(false, true)
		return
	}
	if !firstSeen {
		log.Info().Msg("duplicate price.updated ignored")
		_ = msg.Ack(false)
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

func (m *Manager) consumeFraudAlerts(ctx context.Context, ch *amqp.Channel) {
	msgs, err := ch.Consume(m.cfg.RabbitMQ.QueueFraudNotif, "notif-fraud-consumer", false, false, false, false, nil)
	if err != nil {
		log.Error().Err(err).Msg("failed to start fraud consumer")
		return
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

	firstSeen, err := m.repo.MarkProcessed(ctx, messageIDOf(msg, "fraud.alert"))
	if err != nil {
		log.Error().Err(err).Msg("failed to check message idempotence — requeuing")
		_ = msg.Nack(false, true)
		return
	}
	if !firstSeen {
		log.Info().Msg("duplicate fraud.alert ignored")
		_ = msg.Ack(false)
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

// ── Order Events Consumer ────────────────────────────────────────────────────

const queueOrderEvents = "notification-service.order.events"

func (m *Manager) consumeOrderEvents(ctx context.Context, ch *amqp.Channel) {
	msgs, err := ch.Consume(queueOrderEvents, "notif-order-consumer", false, false, false, false, nil)
	if err != nil {
		log.Error().Err(err).Msg("failed to start order consumer")
		return
	}
	log.Info().Msg("order consumer started")

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-msgs:
			if !ok {
				return
			}
			switch msg.RoutingKey {
			case "order.created":
				m.handleOrderCreated(ctx, msg)
			case "order.decided":
				m.handleOrderDecided(ctx, msg)
			default:
				_ = msg.Ack(false)
			}
		}
	}
}

// notifyAndEmail persiste une notification, la pousse en temps reel au
// destinataire, puis envoie un email si son adresse est resolue via
// auth-service. L'echec de l'email ne bloque jamais le flux (log uniquement).
func (m *Manager) notifyAndEmail(ctx context.Context, notif *model.Notification, recipientNumericID uint, emailSubject, emailBody string) {
	if err := m.repo.Save(ctx, notif); err != nil {
		log.Error().Err(err).Msg("failed to persist order notification")
	}

	wsMsg := model.WebSocketMessage{
		Event: string(notif.Type),
		Data: map[string]interface{}{
			"notification_id": notif.ID,
			"item_id":         notif.ItemID,
			"title":           notif.Title,
			"body":            notif.Body,
			"created_at":      notif.CreatedAt,
		},
	}
	payload, _ := json.Marshal(wsMsg)
	m.hub.SendToUser(notif.UserID, payload)

	if m.auth == nil {
		return
	}
	user, err := m.auth.GetUser(ctx, recipientNumericID)
	if err != nil {
		log.Warn().Err(err).Uint("user_id", recipientNumericID).Msg("email non envoye : resolution utilisateur echouee")
		return
	}
	m.mail.Send(user.Email, emailSubject, emailBody)
}

func (m *Manager) handleOrderCreated(ctx context.Context, msg amqp.Delivery) {
	var event model.OrderCreatedEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Error().Err(err).Msg("invalid order.created payload")
		_ = msg.Nack(false, false)
		return
	}

	firstSeen, err := m.repo.MarkProcessed(ctx, messageIDOf(msg, "order.created"))
	if err != nil {
		log.Error().Err(err).Msg("failed to check message idempotence — requeuing")
		_ = msg.Nack(false, true)
		return
	}
	if !firstSeen {
		log.Info().Msg("duplicate order.created ignored")
		_ = msg.Ack(false)
		return
	}

	title := "🛒 Nouvelle demande d'achat"
	body := fmt.Sprintf("Un acheteur souhaite acquérir \"%s\" pour %.2f€. Validez ou refusez la commande depuis votre profil.", event.ItemName, event.Price)

	notif := &model.Notification{
		ID:        uuid.New(),
		UserID:    event.SellerID,
		Type:      model.TypeOrderPending,
		Title:     title,
		Body:      body,
		ItemID:    &event.ItemID,
		Read:      false,
		CreatedAt: time.Now(),
	}

	emailBody := fmt.Sprintf(
		"Bonjour,\n\nUn acheteur souhaite acquérir votre annonce \"%s\" pour %.2f€.\n\nConnectez-vous à votre profil Collector.shop pour accepter ou refuser cette commande.\n\n— Collector.shop",
		event.ItemName, event.Price,
	)
	m.notifyAndEmail(ctx, notif, idconv.FromUUID(event.SellerID), title, emailBody)

	log.Info().Str("item_id", event.ItemID.String()).Msg("order.created notification dispatched")
	_ = msg.Ack(false)
}

func (m *Manager) handleOrderDecided(ctx context.Context, msg amqp.Delivery) {
	var event model.OrderDecisionEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Error().Err(err).Msg("invalid order.decided payload")
		_ = msg.Nack(false, false)
		return
	}

	firstSeen, err := m.repo.MarkProcessed(ctx, messageIDOf(msg, "order.decided"))
	if err != nil {
		log.Error().Err(err).Msg("failed to check message idempotence — requeuing")
		_ = msg.Nack(false, true)
		return
	}
	if !firstSeen {
		log.Info().Msg("duplicate order.decided ignored")
		_ = msg.Ack(false)
		return
	}

	notifType := model.TypeOrderRejected
	title := "❌ Commande refusée"
	emailBody := fmt.Sprintf("Bonjour,\n\nLe vendeur a refusé votre commande sur \"%s\" (%.2f€). La pièce est de nouveau disponible à la vente.\n\n— Collector.shop", event.ItemName, event.Price)
	if event.Accepted {
		notifType = model.TypeOrderAccepted
		title = "✅ Commande acceptée"
		emailBody = fmt.Sprintf("Bonjour,\n\nLe vendeur a accepté votre commande sur \"%s\" (%.2f€). Retrouvez le détail dans votre profil.\n\n— Collector.shop", event.ItemName, event.Price)
	}
	body := fmt.Sprintf("Votre commande sur \"%s\" (%.2f€) a été %s par le vendeur.", event.ItemName, event.Price, map[bool]string{true: "acceptée", false: "refusée"}[event.Accepted])

	notif := &model.Notification{
		ID:        uuid.New(),
		UserID:    event.BuyerID,
		Type:      notifType,
		Title:     title,
		Body:      body,
		ItemID:    &event.ItemID,
		Read:      false,
		CreatedAt: time.Now(),
	}

	m.notifyAndEmail(ctx, notif, idconv.FromUUID(event.BuyerID), title, emailBody)

	log.Info().Str("item_id", event.ItemID.String()).Bool("accepted", event.Accepted).Msg("order.decided notification dispatched")
	_ = msg.Ack(false)
}
