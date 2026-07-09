package hub

import (
	"io"
	"sync"
	"time"

	"github.com/JTR220/collector/notification-service/internal/model"
	"github.com/JTR220/collector/notification-service/internal/metrics"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
)

// Client represents a single WebSocket connection for a user
type Client struct {
	UserID uuid.UUID
	conn   *websocket.Conn
	send   chan []byte
	hub    *Hub
}

// Hub maintains all active WebSocket connections and broadcasts messages
type Hub struct {
	mu         sync.RWMutex
	clients    map[uuid.UUID][]*Client // userID → multiple tabs/devices
	register   chan *Client
	unregister chan *Client
	broadcast  chan *BroadcastMsg
}

// BroadcastMsg targets a specific user or all connected users
type BroadcastMsg struct {
	UserID  *uuid.UUID // nil = broadcast to all
	Payload []byte
}

func New() *Hub {
	return &Hub{
		clients:    make(map[uuid.UUID][]*Client),
		register:   make(chan *Client, 64),
		unregister: make(chan *Client, 64),
		broadcast:  make(chan *BroadcastMsg, 256),
	}
}

// Run starts the hub event loop — must be called in a goroutine
func (h *Hub) Run() {
	log.Info().Msg("WebSocket hub started")
	for {
		select {
		case client := <-h.register:
			h.handleRegister(client)
		case client := <-h.unregister:
			h.handleUnregister(client)
		case msg := <-h.broadcast:
			h.handleBroadcast(msg)
		}
	}
}

func (h *Hub) handleRegister(client *Client) {
	h.mu.Lock()
	h.clients[client.UserID] = append(h.clients[client.UserID], client)
	metrics.SetWebSocketActiveConnections(h.connectionCountLocked())
	h.mu.Unlock()
	log.Info().Str("user_id", client.UserID.String()).Msg("WebSocket client connected")
}

func (h *Hub) handleUnregister(client *Client) {
	h.mu.Lock()
	clients := h.clients[client.UserID]
	for i, c := range clients {
		if c == client {
			h.clients[client.UserID] = append(clients[:i], clients[i+1:]...)
			close(client.send)
			break
		}
	}
	if len(h.clients[client.UserID]) == 0 {
		delete(h.clients, client.UserID)
	}
	metrics.SetWebSocketActiveConnections(h.connectionCountLocked())
	h.mu.Unlock()
	log.Info().Str("user_id", client.UserID.String()).Msg("WebSocket client disconnected")
}

func (h *Hub) handleBroadcast(msg *BroadcastMsg) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if msg.UserID == nil {
		// Broadcast to ALL connected clients
		for _, clients := range h.clients {
			dispatchToClients(clients, msg.Payload)
		}
		return
	}
	// Send to a specific user (all their devices/tabs)
	dispatchToClients(h.clients[*msg.UserID], msg.Payload)
}

// dispatchToClients pousse payload sur le canal d'emission de chaque client ;
// si le buffer d'un client est plein, le message est abandonne pour lui sans
// bloquer ni impacter les autres clients.
func dispatchToClients(clients []*Client, payload []byte) {
	for _, c := range clients {
		select {
		case c.send <- payload:
		default:
			log.Warn().Str("user_id", c.UserID.String()).Msg("client send buffer full, dropping message")
		}
	}
}

// SendToUser sends a JSON payload to all connections of a specific user
func (h *Hub) SendToUser(userID uuid.UUID, payload []byte) {
	h.broadcast <- &BroadcastMsg{UserID: &userID, Payload: payload}
}

// SendToAll broadcasts a JSON payload to every connected client
func (h *Hub) SendToAll(payload []byte) {
	h.broadcast <- &BroadcastMsg{UserID: nil, Payload: payload}
}

// ConnectedCount returns the number of unique connected users
func (h *Hub) ConnectedCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

func (h *Hub) connectionCountLocked() int {
	total := 0
	for _, clients := range h.clients {
		total += len(clients)
	}
	return total
}

// NewClient creates a client and registers it with the hub
func (h *Hub) NewClient(userID uuid.UUID, conn *websocket.Conn) *Client {
	c := &Client{
		UserID: userID,
		conn:   conn,
		send:   make(chan []byte, 256),
		hub:    h,
	}
	h.register <- c
	return c
}

// WritePump pumps messages from the send channel to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
		c.hub.unregister <- c
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !c.writeMessage(message, ok) {
				return
			}
		case <-ticker.C:
			if !c.writePing() {
				return
			}
		}
	}
}

// writeMessage ecrit message sur la connexion (et vide le buffer restant en
// une seule frame WebSocket). Renvoie false si la connexion doit se fermer :
// canal ferme cote hub (ok == false) ou erreur d'ecriture.
func (c *Client) writeMessage(message []byte, ok bool) bool {
	_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
	if !ok {
		_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
		return false
	}
	w, err := c.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return false
	}
	if _, err := w.Write(message); err != nil {
		return false
	}
	if !c.flushPending(w) {
		return false
	}
	return w.Close() == nil
}

// flushPending vide le reste du buffer d'emission dans le writer courant,
// pour eviter une frame WebSocket par message quand plusieurs sont en attente.
func (c *Client) flushPending(w io.Writer) bool {
	n := len(c.send)
	for i := 0; i < n; i++ {
		if _, err := w.Write([]byte("\n")); err != nil {
			return false
		}
		if _, err := w.Write(<-c.send); err != nil {
			return false
		}
	}
	return true
}

func (c *Client) writePing() bool {
	_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
	return c.conn.WriteMessage(websocket.PingMessage, nil) == nil
}

// ReadPump pumps messages from the WebSocket to discard them (keeps conn alive)
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Warn().Err(err).Str("user_id", c.UserID.String()).Msg("WebSocket read error")
			}
			return
		}
		// Handle client-side messages (ex: mark notification as read)
		log.Debug().Str("user_id", c.UserID.String()).Bytes("msg", msg).Msg("received client message")
	}
}

// Ensure model is imported
var _ = model.TypePriceDrop
