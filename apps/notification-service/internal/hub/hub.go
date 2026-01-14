package hub

import (
"log"
"sync"

"github.com/gorilla/websocket"
)

type Hub struct {
	clients map[string]*websocket.Conn // userID -> conn
	mu      sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]*websocket.Conn),
	}
}

func (h *Hub) Register(userID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[userID] = conn
}

func (h *Hub) Unregister(userID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if conn, ok := h.clients[userID]; ok {
		conn.Close()
		delete(h.clients, userID)
	}
}

func (h *Hub) Send(userID string, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if conn, ok := h.clients[userID]; ok {
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("Failed to send to %s: %v", userID, err)
			conn.Close()
			delete(h.clients, userID)
		}
	}
}
