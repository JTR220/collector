package events

import (
	"encoding/json"
	"testing"
)

func TestToEventUUID(t *testing.T) {
	cases := map[uint]string{
		0:   "00000000-0000-0000-0000-000000000000",
		1:   "00000000-0000-0000-0000-000000000001",
		255: "00000000-0000-0000-0000-0000000000ff",
	}
	for id, want := range cases {
		if got := ToEventUUID(id); got != want {
			t.Errorf("ToEventUUID(%d) = %q, attendu %q", id, got, want)
		}
	}
}

func TestPriceUpdatedEventJSONTags(t *testing.T) {
	body, err := json.Marshal(PriceUpdatedEvent{
		ItemID:   ToEventUUID(1),
		SellerID: ToEventUUID(2),
		OldPrice: 10,
		NewPrice: 20,
	})
	if err != nil {
		t.Fatalf("marshal : %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		t.Fatalf("unmarshal : %v", err)
	}

	// Ces clefs doivent matcher model.PriceUpdatedEvent des consumers.
	for _, key := range []string{"item_id", "seller_id", "old_price", "new_price", "updated_at"} {
		if _, ok := raw[key]; !ok {
			t.Errorf("clef JSON %q absente du payload", key)
		}
	}
}

func TestNoopPublisherDoesNotPanic(t *testing.T) {
	var p Publisher = NoopPublisher{}
	p.PublishPriceUpdated(1, 2, 10, 20)
	p.PublishOrderCreated(1, 2, 3, 4, "Lot", 60)
	p.PublishOrderDecision(1, 2, 3, 4, "Lot", 60, true)
	p.Close()
}

// Sans connexion active (ch nil, comme avant tout Dial reussi ou apres perte
// de connexion), les methodes Publish* ne doivent jamais paniquer : elles se
// contentent de logger et abandonner l'evenement.
func TestAMQPPublisherDisconnectedDoesNotPanic(t *testing.T) {
	p := &AMQPPublisher{}
	p.PublishPriceUpdated(1, 2, 10, 20)
	p.PublishOrderCreated(1, 2, 3, 4, "Lot", 60)
	p.PublishOrderDecision(1, 2, 3, 4, "Lot", 60, false)
	p.Close()
}

func TestOrderCreatedEventJSONTags(t *testing.T) {
	body, err := json.Marshal(OrderCreatedEvent{
		OrderID:  ToEventUUID(1),
		ItemID:   ToEventUUID(2),
		ItemName: "Dracaufeu",
		BuyerID:  ToEventUUID(3),
		SellerID: ToEventUUID(4),
		Price:    100,
	})
	if err != nil {
		t.Fatalf("marshal : %v", err)
	}
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		t.Fatalf("unmarshal : %v", err)
	}
	for _, key := range []string{"order_id", "item_id", "item_name", "buyer_id", "seller_id", "price", "created_at"} {
		if _, ok := raw[key]; !ok {
			t.Errorf("clef JSON %q absente du payload OrderCreatedEvent", key)
		}
	}
}

func TestOrderDecisionEventJSONTags(t *testing.T) {
	body, err := json.Marshal(OrderDecisionEvent{
		OrderID:  ToEventUUID(1),
		ItemID:   ToEventUUID(2),
		ItemName: "Dracaufeu",
		BuyerID:  ToEventUUID(3),
		SellerID: ToEventUUID(4),
		Price:    100,
		Accepted: true,
	})
	if err != nil {
		t.Fatalf("marshal : %v", err)
	}
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		t.Fatalf("unmarshal : %v", err)
	}
	for _, key := range []string{"order_id", "item_id", "item_name", "buyer_id", "seller_id", "price", "accepted", "decided_at"} {
		if _, ok := raw[key]; !ok {
			t.Errorf("clef JSON %q absente du payload OrderDecisionEvent", key)
		}
	}
}

func TestInitWithEmptyURLUsesNoop(t *testing.T) {
	Init("")
	if _, ok := Current.(NoopPublisher); !ok {
		t.Errorf("avec RABBITMQ_URL vide, Current devrait etre NoopPublisher, obtenu %T", Current)
	}
}

// TestInitWithURLCreatesAMQPPublisher verifie que Init bascule Current vers un
// AMQPPublisher des qu'une URL est fournie, meme si le broker est injoignable
// (connectLoop tourne en arriere-plan et n'empeche jamais le demarrage).
func TestInitWithURLCreatesAMQPPublisher(t *testing.T) {
	Init("amqp://guest:guest@localhost:1/") // port invalide, jamais joignable
	p, ok := Current.(*AMQPPublisher)
	if !ok {
		t.Fatalf("avec une URL non vide, Current devrait etre *AMQPPublisher, obtenu %T", Current)
	}
	p.Close()
	Current = NoopPublisher{}
}

// TestMessageIDIsDeterministic garantit que deux evenements identiques
// produisent le meme MessageId AMQP (necessaire a la deduplication cote
// consumer), et que des evenements differents produisent des ids differents.
func TestMessageIDIsDeterministic(t *testing.T) {
	e1 := PriceUpdatedEvent{ItemID: ToEventUUID(1), SellerID: ToEventUUID(2), OldPrice: 10, NewPrice: 20}
	e2 := PriceUpdatedEvent{ItemID: ToEventUUID(1), SellerID: ToEventUUID(2), OldPrice: 10, NewPrice: 20}
	if messageID(e1) != messageID(e2) {
		t.Error("deux evenements identiques devraient produire le meme MessageId")
	}

	e3 := PriceUpdatedEvent{ItemID: ToEventUUID(1), SellerID: ToEventUUID(2), OldPrice: 10, NewPrice: 30}
	if messageID(e1) == messageID(e3) {
		t.Error("deux evenements differents ne devraient pas produire le meme MessageId")
	}
}

// TestAMQPPublisherClose_Idempotent verifie que Close peut etre appele sans
// connexion active (avant tout Dial reussi) sans paniquer, et que closed=true
// empeche connectLoop de retenter une connexion.
func TestAMQPPublisherClose_Idempotent(t *testing.T) {
	p := &AMQPPublisher{}
	p.Close()
	p.Close() // deuxieme appel : ne doit pas paniquer
	if !p.closed {
		t.Error("closed devrait etre true apres Close()")
	}
}
