package hub_test

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/JTR220/collector/notification-service/internal/hub"
	"github.com/JTR220/collector/notification-service/internal/model"
)

func TestHub_Run_StartsWithoutPanic(t *testing.T) {
	h := hub.New()
	done := make(chan struct{})
	go func() {
		defer close(done)
		// Hub.Run() blocks — we just verify it starts
		go h.Run()
		time.Sleep(10 * time.Millisecond)
	}()
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		// OK — hub is blocking as expected
	}
}

func TestHub_ConnectedCount_StartsAtZero(t *testing.T) {
	h := hub.New()
	go h.Run()
	time.Sleep(10 * time.Millisecond)

	if count := h.ConnectedCount(); count != 0 {
		t.Errorf("expected 0 connected clients, got %d", count)
	}
}

func TestWebSocketMessage_Serialization(t *testing.T) {
	itemID := uuid.New()
	msg := model.WebSocketMessage{
		Event: string(model.TypePriceDrop),
		Data: map[string]interface{}{
			"item_id":   itemID,
			"old_price": 100.0,
			"new_price": 80.0,
		},
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("failed to marshal WebSocketMessage: %v", err)
	}

	var decoded model.WebSocketMessage
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("failed to unmarshal WebSocketMessage: %v", err)
	}

	if decoded.Event != string(model.TypePriceDrop) {
		t.Errorf("expected event %s, got %s", model.TypePriceDrop, decoded.Event)
	}
}

func TestHub_SendToAll_NoPanic(t *testing.T) {
	h := hub.New()
	go h.Run()
	time.Sleep(10 * time.Millisecond)

	// No clients connected — sending to all should not panic
	payload, _ := json.Marshal(model.WebSocketMessage{Event: "TEST", Data: "hello"})
	h.SendToAll(payload)
}

func TestHub_SendToUser_NoPanic(t *testing.T) {
	h := hub.New()
	go h.Run()
	time.Sleep(10 * time.Millisecond)

	// User not connected — sending to unknown user should not panic
	fakeUserID := uuid.New()
	payload, _ := json.Marshal(model.WebSocketMessage{Event: "TEST", Data: "hello"})
	h.SendToUser(fakeUserID, payload)
}

func TestHub_ConcurrentSend_NoPanic(t *testing.T) {
	h := hub.New()
	go h.Run()
	time.Sleep(10 * time.Millisecond)

	payload, _ := json.Marshal(model.WebSocketMessage{Event: "TEST", Data: "concurrent"})

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			h.SendToAll(payload)
		}()
	}
	wg.Wait()
}

func TestNotificationTypes_AreWellDefined(t *testing.T) {
	types := []model.NotificationType{
		model.TypePriceDrop,
		model.TypePriceSpike,
		model.TypeFraudAlert,
		model.TypeNewItem,
		model.TypeItemSold,
	}

	seen := make(map[model.NotificationType]bool)
	for _, nt := range types {
		if seen[nt] {
			t.Errorf("duplicate NotificationType: %s", nt)
		}
		if nt == "" {
			t.Error("empty NotificationType found")
		}
		seen[nt] = true
	}
}
