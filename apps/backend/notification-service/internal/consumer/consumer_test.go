package consumer

import (
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func TestMessageIDOfUsesMessageIdWhenPresent(t *testing.T) {
	msg := amqp.Delivery{MessageId: "explicit-id"}
	if got := messageIDOf(msg, "price.updated"); got != "explicit-id" {
		t.Errorf("attendu 'explicit-id', obtenu %q", got)
	}
}

func TestMessageIDOfHashesBodyWhenMessageIdMissing(t *testing.T) {
	msg := amqp.Delivery{Body: []byte(`{"item_id":"1"}`)}
	got := messageIDOf(msg, "order.created")
	if got == "" {
		t.Fatal("un identifiant non vide est attendu")
	}
	// Deterministe : le meme corps doit toujours produire le meme identifiant
	// (garantit la deduplication meme sans MessageId pose par le publisher).
	again := messageIDOf(msg, "order.created")
	if got != again {
		t.Errorf("messageIDOf devrait etre deterministe : %q != %q", got, again)
	}
	// Le prefixe distingue les types d'evenements pour eviter une collision
	// fortuite entre deux payloads identiques d'origines differentes.
	other := messageIDOf(msg, "order.decided")
	if got == other {
		t.Error("deux prefixes differents ne devraient pas produire le meme identifiant")
	}
}

func TestNextBackoffDoublesUpToCeiling(t *testing.T) {
	d := time.Second
	for i := 0; i < 10; i++ {
		d = nextBackoff(d)
	}
	if d != 30*time.Second {
		t.Errorf("le backoff devrait plafonner a 30s, obtenu %s", d)
	}
}

func TestNextBackoffFromZero(t *testing.T) {
	if got := nextBackoff(time.Second); got != 2*time.Second {
		t.Errorf("attendu 2s, obtenu %s", got)
	}
}
