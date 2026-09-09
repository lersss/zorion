package handlers

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type WebSocketHub struct {
	mu         sync.RWMutex
	clients    map[string]*websocket.Conn // key: userID
}

func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		clients: make(map[string]*websocket.Conn),
	}
}

func (h *WebSocketHub) Register(userID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	// Закрываем существующее соединение, если есть
	if old, ok := h.clients[userID]; ok {
		old.Close()
	}
	h.clients[userID] = conn
	log.Printf("WebSocket registered: user %s", userID)
}

func (h *WebSocketHub) Unregister(userID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, userID)
}

func (h *WebSocketHub) SendToUser(userID string, message []byte) error {
	h.mu.RLock()
	conn, ok := h.clients[userID]
	h.mu.RUnlock()
	if !ok {
		return nil // пользователь не в сети
	}
	return conn.WriteMessage(websocket.TextMessage, message)
}