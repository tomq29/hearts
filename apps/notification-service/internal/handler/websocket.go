package handler

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/kisssonik/hearts/apps/notification-service/internal/hub"
	"go.uber.org/zap"
)

type WebSocketHandler struct {
	hub    *hub.Hub
	logger *zap.Logger
}

func NewWebSocketHandler(hub *hub.Hub, logger *zap.Logger) *WebSocketHandler {
	return &WebSocketHandler{
		hub:    hub,
		logger: logger,
	}
}

func (h *WebSocketHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// In a real app, validate JWT token here!
	userID := r.URL.Query().Get("userId")
	if userID == "" {
		http.Error(w, "UserId required", http.StatusBadRequest)
		return
	}

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("Upgrade failed", zap.Error(err))
		return
	}

	h.hub.Register(userID, conn)
	h.logger.Info("Client connected", zap.String("userId", userID))

	// Keep connection alive
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			h.hub.Unregister(userID)
			break
		}
	}
}
