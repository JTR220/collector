package hub_test

// Complete hub_test.go : exerce NewClient / WritePump / ReadPump avec une
// vraie connexion WebSocket (httptest.Server + gorilla/websocket), pour
// couvrir le cycle de vie complet d'un client (register, envoi, unregister)
// qui n'etait pas atteint par les tests portant uniquement sur le Hub.

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JTR220/collector/notification-service/internal/hub"
	"github.com/JTR220/collector/notification-service/internal/model"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var testUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// newTestServer demarre un serveur HTTP qui upgrade toute connexion en
// WebSocket et l'enregistre aupres du hub, comme le fait le vrai handler.
func newTestServer(t *testing.T, h *hub.Hub, userID uuid.UUID) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := testUpgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade: %v", err)
			return
		}
		client := h.NewClient(userID, conn)
		go client.WritePump()
		go client.ReadPump()
	}))
	t.Cleanup(srv.Close)
	return srv
}

func dialWS(t *testing.T, srv *httptest.Server) *websocket.Conn {
	t.Helper()
	url := "ws" + srv.URL[len("http"):]
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

// waitForConnectedCount poll le hub jusqu'a ce que ConnectedCount atteigne
// la valeur attendue (evite les tests flaky sur un simple sleep fixe : le
// register passe par un channel asynchrone traite dans la goroutine Run).
func waitForConnectedCount(t *testing.T, h *hub.Hub, want int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if h.ConnectedCount() == want {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("attendu %d client(s) connecte(s), obtenu %d", want, h.ConnectedCount())
}

func TestClient_RegistersAndReceivesMessage(t *testing.T) {
	h := hub.New()
	go h.Run()
	time.Sleep(10 * time.Millisecond)

	userID := uuid.New()
	srv := newTestServer(t, h, userID)
	conn := dialWS(t, srv)

	// Laisse le temps au client de s'enregistrer aupres du hub.
	waitForConnectedCount(t, h, 1)

	payload, _ := json.Marshal(model.WebSocketMessage{Event: "TEST", Data: "hello"})
	h.SendToUser(userID, payload)

	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("lecture message : %v", err)
	}
	var decoded model.WebSocketMessage
	if err := json.Unmarshal(msg, &decoded); err != nil {
		t.Fatalf("decodage message : %v", err)
	}
	if decoded.Event != "TEST" {
		t.Errorf("attendu event TEST, obtenu %s", decoded.Event)
	}
}

func TestClient_DisconnectUnregisters(t *testing.T) {
	h := hub.New()
	go h.Run()
	time.Sleep(10 * time.Millisecond)

	userID := uuid.New()
	srv := newTestServer(t, h, userID)
	conn := dialWS(t, srv)
	waitForConnectedCount(t, h, 1)

	_ = conn.Close()
	// Laisse le temps a ReadPump de detecter la fermeture et de se
	// desenregistrer aupres du hub.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if h.ConnectedCount() == 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("attendu 0 client connecte apres deconnexion, obtenu %d", h.ConnectedCount())
}

func TestClient_MultipleTabsSameUser(t *testing.T) {
	h := hub.New()
	go h.Run()
	time.Sleep(10 * time.Millisecond)

	userID := uuid.New()
	srv := newTestServer(t, h, userID)
	conn1 := dialWS(t, srv)
	conn2 := dialWS(t, srv)
	// Un seul utilisateur connecte, meme avec deux onglets/connexions.
	waitForConnectedCount(t, h, 1)

	payload, _ := json.Marshal(model.WebSocketMessage{Event: "BROADCAST", Data: "hi"})
	h.SendToUser(userID, payload)

	for _, conn := range []*websocket.Conn{conn1, conn2} {
		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		if _, _, err := conn.ReadMessage(); err != nil {
			t.Errorf("lecture message sur une des connexions : %v", err)
		}
	}
}
