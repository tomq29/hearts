package handler

import (
	"encoding/json"
	"net/http"

	"github.com/kisssonik/hearts/internal/notification/service"
	"github.com/kisssonik/hearts/pkg/auth"
	"go.uber.org/zap"
)

type NotificationHandler struct {
	service service.NotificationService
	logger  *zap.Logger
}

func NewNotificationHandler(s service.NotificationService, l *zap.Logger) *NotificationHandler {
	return &NotificationHandler{service: s, logger: l}
}

// List handles listing notifications.
// @Summary List notifications
// @Description Get all notifications for the authenticated user.
// @Tags notifications
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} notification.Notification
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /notifications [get]
func (h *NotificationHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	notifications, err := h.service.GetNotifications(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get notifications", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notifications)
}
