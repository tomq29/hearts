package handler

import (
	"encoding/json"
	"net/http"

	"github.com/kisssonik/hearts/internal/chat/service"
	"github.com/kisssonik/hearts/pkg/auth"
	"go.uber.org/zap"
)

type ChatHandler struct {
	service service.ChatService
	logger  *zap.Logger
}

func NewChatHandler(s service.ChatService, l *zap.Logger) *ChatHandler {
	return &ChatHandler{service: s, logger: l}
}

// GetHistory handles retrieving chat history.
// @Summary Get chat history
// @Description Get messages between the authenticated user and another user.
// @Tags chat
// @Produce json
// @Security ApiKeyAuth
// @Param otherUserID path string true "Other User ID"
// @Success 200 {array} chat.Message
// @Failure 400 {string} string "User ID required"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Not matched"
// @Failure 500 {string} string "Internal server error"
// @Router /chats/{otherUserID}/messages [get]
func (h *ChatHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	otherUserID := r.PathValue("otherUserID")
	if otherUserID == "" {
		http.Error(w, "Other User ID required", http.StatusBadRequest)
		return
	}

	messages, err := h.service.GetHistory(r.Context(), userID, otherUserID)
	if err != nil {
		if err.Error() == "users are not matched" {
			http.Error(w, "Not matched", http.StatusForbidden)
			return
		}
		h.logger.Error("Failed to get chat history", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}
