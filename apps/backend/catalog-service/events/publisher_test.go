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
