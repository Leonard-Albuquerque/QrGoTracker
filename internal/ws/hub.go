package ws

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[*websocket.Conn]bool
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]map[*websocket.Conn]bool),
	}
}

func (h *Hub) Register(code string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[code]; !ok {
		h.clients[code] = make(map[*websocket.Conn]bool)
	}
	h.clients[code][conn] = true
}

func (h *Hub) Unregister(code string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if conns, ok := h.clients[code]; ok {
		delete(conns, conn)
		if len(conns) == 0 {
			delete(h.clients, code)
		}
	}
}

func (h *Hub) Broadcast(code string, clickCount int64) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	conns, ok := h.clients[code]
	if !ok {
		return
	}

	payload := map[string]interface{}{
		"code":        code,
		"click_count": clickCount,
	}

	for conn := range conns {
		_ = conn.WriteJSON(payload)
	}
}
