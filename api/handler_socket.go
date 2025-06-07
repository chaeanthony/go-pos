package api

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// WebSocket upgrader with permissive origin checking (adjust for production)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Hub maintains active clients and broadcasts messages
type Hub struct {
	clients map[*websocket.Conn]bool
	mu      sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[*websocket.Conn]bool),
	}
}

func (h *Hub) AddClient(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[conn] = true
}

func (h *Hub) RemoveClient(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, conn)
	conn.Close()
}

// Broadcast message to all clients
func (h *Hub) Broadcast(message []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for client := range h.clients {
		err := client.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("Error sending message to client: %v", err)
			client.Close()
			delete(h.clients, client)
		}
	}
}

func (cfg *APIConfig) WsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		cfg.Logger.Errorf("Upgrade error: %v", err)
		return
	}
	defer func() {
		cfg.Hub.RemoveClient(conn)
		conn.Close()
		cfg.Logger.Infof("Client disconnected: IP=%s", r.RemoteAddr)
	}()

	cfg.Hub.AddClient(conn)
	cfg.Logger.Infof("Client connected: IP=%s", r.RemoteAddr)

	for {
		// Wait for next message or close
		if _, _, err := conn.NextReader(); err != nil {
			cfg.Logger.Infof("Client disconnected: IP=%s, error: %v", r.RemoteAddr, err)
			break
		}
	}

}
