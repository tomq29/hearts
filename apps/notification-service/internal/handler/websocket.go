package handler

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/kisssonik/hearts/apps/notification-service/internal/hub"
	"go.uber.org/zap"
)

type WebSocketHandler struct {
	hub      *hub.Hub
	logger   *zap.Logger
	upgrader *websocket.Upgrader
}

// NewWebSocketHandler builds the handler. allowedOrigins restricts which
// browser origins may open a socket: WebSocket handshakes are not covered by
// the same-origin policy, so without the check any page could connect. An
// absent Origin header means a non-browser client and is allowed.
func NewWebSocketHandler(hub *hub.Hub, logger *zap.Logger, allowedOrigins []string) *WebSocketHandler {
	allowed := make(map[string]bool, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		allowed[origin] = true
	}
	return &WebSocketHandler{
		hub:    hub,
		logger: logger,
		upgrader: &websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				if origin == "" {
					return true
				}
				return allowed[origin]
			},
		},
	}
}

func (h *WebSocketHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// SECURITY: this endpoint is unauthenticated — it trusts the userId query
	// parameter, so anyone who knows a user ID can read that user's
	// notification stream. The frontend does not use this socket yet (matches
	// are delivered over the API's own /ws); it must validate a JWT or a
	// one-time ticket before being exposed publicly.
	userID := r.URL.Query().Get("userId")
	if userID == "" {
		http.Error(w, "UserId required", http.StatusBadRequest)
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
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
