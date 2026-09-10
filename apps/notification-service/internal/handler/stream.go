package handler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kisssonik/hearts/apps/notification-service/internal/auth"
	"github.com/kisssonik/hearts/apps/notification-service/internal/hub"
	"go.uber.org/zap"
)

// keepaliveInterval keeps idle streams open through proxies that drop
// connections carrying no traffic, typically after 30–60 seconds.
const keepaliveInterval = 25 * time.Second

// StreamHandler serves a user's notifications as Server-Sent Events.
//
// SSE rather than a WebSocket: notifications only flow server to client, and
// a fetch-based SSE client can send an Authorization header, which a browser
// WebSocket handshake cannot — so the credential never has to go in the URL.
type StreamHandler struct {
	hub            *hub.Hub
	verifier       *auth.Verifier
	logger         *zap.Logger
	allowedOrigins map[string]bool
}

func NewStreamHandler(h *hub.Hub, v *auth.Verifier, logger *zap.Logger, allowedOrigins []string) *StreamHandler {
	allowed := make(map[string]bool, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		allowed[origin] = true
	}
	return &StreamHandler{hub: h, verifier: v, logger: logger, allowedOrigins: allowed}
}

func (h *StreamHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.setCORSHeaders(w, r)

	// A cross-origin request carrying Authorization is preflighted, and the
	// preflight never carries credentials — so it must be answered before auth.
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// The user comes from the verified token, never from the request itself:
	// this endpoint previously trusted a userId query parameter outright.
	userID, err := h.verifier.UserID(r.Header.Get("Authorization"))
	if err != nil {
		h.logger.Warn("Rejected notification stream", zap.Error(err))
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Subscribe before telling the client the stream is open: once it sees the
	// 200 it may reasonably expect delivery, so nothing published after that
	// point may be missed.
	messages, unsubscribe := h.hub.Subscribe(userID)
	defer unsubscribe()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	h.logger.Info("Notification stream opened", zap.String("userId", userID))

	keepalive := time.NewTicker(keepaliveInterval)
	defer keepalive.Stop()

	for {
		select {
		case <-r.Context().Done():
			h.logger.Info("Notification stream closed", zap.String("userId", userID))
			return
		case msg := <-messages:
			writeEvent(w, msg)
			flusher.Flush()
		case <-keepalive.C:
			// Lines starting with ":" are comments; clients ignore them.
			fmt.Fprint(w, ": keepalive\n\n")
			flusher.Flush()
		}
	}
}

// writeEvent frames data as one SSE event. Each line needs its own "data:"
// prefix, or a newline inside the payload would end the event early.
func writeEvent(w io.Writer, data []byte) {
	for _, line := range bytes.Split(data, []byte("\n")) {
		fmt.Fprintf(w, "data: %s\n", line)
	}
	fmt.Fprint(w, "\n")
}

func (h *StreamHandler) setCORSHeaders(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if !h.allowedOrigins[origin] {
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Headers", "Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	// Responses differ per Origin, so caches must key on it.
	w.Header().Add("Vary", "Origin")
}
