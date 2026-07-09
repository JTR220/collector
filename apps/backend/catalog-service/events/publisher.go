package events

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	exchangeEvents         = "collector.events"
	routingKeyPrice        = "price.updated"
	routingKeyOrderCreated = "order.created"
	routingKeyOrderDecided = "order.decided"
)

// Publisher publie les evenements metier du catalogue.
type Publisher interface {
	PublishPriceUpdated(itemID, sellerID uint, oldPrice, newPrice float64)
	PublishOrderCreated(orderID, itemID, buyerID, sellerID uint, itemName string, price float64)
	PublishOrderDecision(orderID, itemID, buyerID, sellerID uint, itemName string, price float64, accepted bool)
	Close()
}

// Current est le publisher actif, initialise dans main.go (meme style que repository.DB).
var Current Publisher = NoopPublisher{}

// NoopPublisher est utilise quand RABBITMQ_URL n'est pas defini :
// le service fonctionne normalement, sans publication.
type NoopPublisher struct{}

func (NoopPublisher) PublishPriceUpdated(itemID, sellerID uint, oldPrice, newPrice float64) {
	// no-op : RABBITMQ_URL absent, aucune publication attendue.
}
func (NoopPublisher) PublishOrderCreated(orderID, itemID, buyerID, sellerID uint, itemName string, price float64) {
	// no-op : RABBITMQ_URL absent, aucune publication attendue.
}
func (NoopPublisher) PublishOrderDecision(orderID, itemID, buyerID, sellerID uint, itemName string, price float64, accepted bool) {
	// no-op : RABBITMQ_URL absent, aucune publication attendue.
}
func (NoopPublisher) Close() {
	// no-op : aucune connexion a fermer.
}

// AMQPPublisher publie sur RabbitMQ. La connexion s'etablit en arriere-plan
// avec retry : le demarrage du service ne depend jamais du broker.
type AMQPPublisher struct {
	url    string
	mu     sync.Mutex
	ch     *amqp.Channel
	conn   *amqp.Connection
	closed bool
}

func NewAMQPPublisher(url string) *AMQPPublisher {
	p := &AMQPPublisher{url: url}
	go p.connectLoop()
	return p
}

func (p *AMQPPublisher) connectLoop() {
	backoff := time.Second
	for !p.isClosed() {
		conn, ch, err := p.dialAndDeclare()
		if err != nil {
			log.Printf("RabbitMQ indisponible (%v), nouvel essai dans %s", err, backoff)
			time.Sleep(backoff)
			backoff = nextPublisherBackoff(backoff)
			continue
		}

		p.setActive(conn, ch)
		log.Printf("RabbitMQ connecte, exchange %q pret", exchangeEvents)
		<-conn.NotifyClose(make(chan *amqp.Error, 1))

		if p.clearActive() {
			return
		}
		log.Printf("Connexion RabbitMQ perdue, reconnexion...")
		backoff = time.Second
	}
}

func (p *AMQPPublisher) isClosed() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.closed
}

// dialAndDeclare ouvre une connexion et un canal AMQP, puis declare
// l'exchange collector.events. En cas d'erreur a n'importe quelle etape,
// la connexion ouverte est refermee et l'erreur remontee telle quelle.
func (p *AMQPPublisher) dialAndDeclare() (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(p.url)
	if err != nil {
		return nil, nil, err
	}
	ch, err := conn.Channel()
	if err == nil {
		err = ch.ExchangeDeclare(exchangeEvents, "topic", true, false, false, false, nil)
	}
	if err != nil {
		_ = conn.Close()
		return nil, nil, err
	}
	return conn, ch, nil
}

func (p *AMQPPublisher) setActive(conn *amqp.Connection, ch *amqp.Channel) {
	p.mu.Lock()
	p.conn = conn
	p.ch = ch
	p.mu.Unlock()
}

// clearActive efface la connexion/canal courants et renvoie true si le
// publisher a ete ferme entre-temps (Close), auquel cas connectLoop doit
// s'arreter plutot que de tenter une reconnexion.
func (p *AMQPPublisher) clearActive() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.conn = nil
	p.ch = nil
	return p.closed
}

func nextPublisherBackoff(d time.Duration) time.Duration {
	if d >= 30*time.Second {
		return 30 * time.Second
	}
	return d * 2
}

// PublishPriceUpdated publie l'evenement price.updated. Si le broker est
// indisponible, l'evenement est abandonne avec un warn : on ne bloque jamais
// la requete HTTP appelante.
func (p *AMQPPublisher) PublishPriceUpdated(itemID, sellerID uint, oldPrice, newPrice float64) {
	event := PriceUpdatedEvent{
		ItemID:    ToEventUUID(itemID),
		SellerID:  ToEventUUID(sellerID),
		OldPrice:  oldPrice,
		NewPrice:  newPrice,
		UpdatedAt: time.Now().UTC(),
	}

	body, err := json.Marshal(event)
	if err != nil {
		log.Printf("price.updated non publie (marshal) : %v", err)
		return
	}

	p.mu.Lock()
	ch := p.ch
	p.mu.Unlock()
	if ch == nil {
		log.Printf("price.updated non publie (broker deconnecte) : article %d", itemID)
		return
	}

	err = ch.Publish(exchangeEvents, routingKeyPrice, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		// MessageId deterministe : permet aux consumers de dedupliquer les
		// livraisons redelivrees (meme evenement => meme MessageId).
		MessageId: messageID(event),
		Timestamp: time.Now(),
		Body:      body,
	})
	if err != nil {
		log.Printf("price.updated non publie (publish) : %v", err)
		return
	}
	log.Printf("price.updated publie : article %d, %.2f -> %.2f", itemID, oldPrice, newPrice)
}

// messageID derive un identifiant stable et deterministe pour un evenement
// price.updated, utilise comme AMQP MessageId. Meme article + memes prix +
// meme horodatage => meme MessageId, ce qui permet aux consumers de
// deduplicquer une livraison redelivree par RabbitMQ (idempotence).
func messageID(event PriceUpdatedEvent) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%.2f|%.2f|%s",
		event.ItemID, event.SellerID, event.OldPrice, event.NewPrice, event.UpdatedAt.UTC().Format(time.RFC3339Nano))))
	return hex.EncodeToString(h[:])
}

// PublishOrderCreated publie order.created quand un acheteur passe commande :
// le vendeur doit valider ou refuser (notification-service s'en charge, avec
// notification + email au vendeur).
func (p *AMQPPublisher) PublishOrderCreated(orderID, itemID, buyerID, sellerID uint, itemName string, price float64) {
	event := OrderCreatedEvent{
		OrderID:   ToEventUUID(orderID),
		ItemID:    ToEventUUID(itemID),
		ItemName:  itemName,
		BuyerID:   ToEventUUID(buyerID),
		SellerID:  ToEventUUID(sellerID),
		Price:     price,
		CreatedAt: time.Now().UTC(),
	}
	p.publishJSON(routingKeyOrderCreated, event, fmt.Sprintf("order.created:%d", orderID))
}

// PublishOrderDecision publie order.decided quand le vendeur accepte ou
// refuse une commande : l'acheteur est notifie du resultat.
func (p *AMQPPublisher) PublishOrderDecision(orderID, itemID, buyerID, sellerID uint, itemName string, price float64, accepted bool) {
	event := OrderDecisionEvent{
		OrderID:   ToEventUUID(orderID),
		ItemID:    ToEventUUID(itemID),
		ItemName:  itemName,
		BuyerID:   ToEventUUID(buyerID),
		SellerID:  ToEventUUID(sellerID),
		Price:     price,
		Accepted:  accepted,
		DecidedAt: time.Now().UTC(),
	}
	p.publishJSON(routingKeyOrderDecided, event, fmt.Sprintf("order.decided:%d:%v", orderID, accepted))
}

// publishJSON serialise et publie un evenement sur l'exchange collector.events.
// Si le broker est indisponible, l'evenement est abandonne avec un warn : on
// ne bloque jamais la requete HTTP appelante.
func (p *AMQPPublisher) publishJSON(routingKey string, event interface{}, msgID string) {
	body, err := json.Marshal(event)
	if err != nil {
		log.Printf("%s non publie (marshal) : %v", routingKey, err)
		return
	}

	p.mu.Lock()
	ch := p.ch
	p.mu.Unlock()
	if ch == nil {
		log.Printf("%s non publie (broker deconnecte)", routingKey)
		return
	}

	err = ch.Publish(exchangeEvents, routingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		MessageId:    msgID,
		Timestamp:    time.Now(),
		Body:         body,
	})
	if err != nil {
		log.Printf("%s non publie (publish) : %v", routingKey, err)
		return
	}
	log.Printf("%s publie", routingKey)
}

func (p *AMQPPublisher) Close() {
	p.mu.Lock()
	p.closed = true
	conn := p.conn
	p.conn = nil
	p.ch = nil
	p.mu.Unlock()
	if conn != nil {
		_ = conn.Close()
	}
}

// Init configure events.Current selon RABBITMQ_URL (vide => Noop).
func Init(rabbitURL string) {
	if rabbitURL == "" {
		log.Printf("RABBITMQ_URL non defini : publication d'evenements desactivee")
		Current = NoopPublisher{}
		return
	}
	Current = NewAMQPPublisher(rabbitURL)
}
