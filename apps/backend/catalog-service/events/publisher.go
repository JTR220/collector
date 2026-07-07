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
	exchangeEvents  = "collector.events"
	routingKeyPrice = "price.updated"
)

// Publisher publie les evenements metier du catalogue.
type Publisher interface {
	PublishPriceUpdated(itemID, sellerID uint, oldPrice, newPrice float64)
	Close()
}

// Current est le publisher actif, initialise dans main.go (meme style que repository.DB).
var Current Publisher = NoopPublisher{}

// NoopPublisher est utilise quand RABBITMQ_URL n'est pas defini :
// le service fonctionne normalement, sans publication.
type NoopPublisher struct{}

func (NoopPublisher) PublishPriceUpdated(itemID, sellerID uint, oldPrice, newPrice float64) {}
func (NoopPublisher) Close()                                                                {}

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
	for {
		p.mu.Lock()
		if p.closed {
			p.mu.Unlock()
			return
		}
		p.mu.Unlock()

		conn, err := amqp.Dial(p.url)
		if err == nil {
			ch, chErr := conn.Channel()
			if chErr == nil {
				chErr = ch.ExchangeDeclare(exchangeEvents, "topic", true, false, false, false, nil)
			}
			if chErr == nil {
				p.mu.Lock()
				p.conn = conn
				p.ch = ch
				p.mu.Unlock()
				log.Printf("RabbitMQ connecte, exchange %q pret", exchangeEvents)

				closeCh := conn.NotifyClose(make(chan *amqp.Error, 1))
				<-closeCh

				p.mu.Lock()
				p.conn = nil
				p.ch = nil
				closed := p.closed
				p.mu.Unlock()
				if closed {
					return
				}
				log.Printf("Connexion RabbitMQ perdue, reconnexion...")
				backoff = time.Second
				continue
			}
			_ = conn.Close()
			err = chErr
		}

		log.Printf("RabbitMQ indisponible (%v), nouvel essai dans %s", err, backoff)
		time.Sleep(backoff)
		if backoff < 30*time.Second {
			backoff *= 2
		}
	}
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
