package handlers

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"zorion/internal/auth"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type WebSocketHandler struct {
	hub *WebSocketHub
}

func NewWebSocketHandler(hub *WebSocketHub) *WebSocketHandler {
	return &WebSocketHandler{hub: hub}
}

func (h *WebSocketHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Регистрируем соединение в хабе
	h.hub.Register(userID, conn)
	defer func() {
		h.hub.Unregister(userID)
		conn.Close()
	}()

	log.Printf("WebSocket connected: user %s", userID)

	// Чтение сообщений от клиента (пока просто игнорируем)
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}