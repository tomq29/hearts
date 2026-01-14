package handler

import (
	"encoding/json"
	"net/http"

	"github.com/kisssonik/hearts/internal/review/service"
	"github.com/kisssonik/hearts/pkg/auth"
	"go.uber.org/zap"
)

type ReviewHandler struct {
	service service.ReviewService
	logger  *zap.Logger
}

func NewReviewHandler(s service.ReviewService, l *zap.Logger) *ReviewHandler {
	return &ReviewHandler{service: s, logger: l}
}

func (h *ReviewHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var input service.CreateReviewInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	rev, err := h.service.CreateReview(r.Context(), userID, input)
	if err != nil {
		if err == service.ErrSelfReview {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err == service.ErrInvalidRating {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err == service.ErrNoMatch {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		h.logger.Error("Failed to create review", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rev)
}

func (h *ReviewHandler) List(w http.ResponseWriter, r *http.Request) {
	targetID := r.PathValue("userID")
	if targetID == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	reviews, err := h.service.GetReviewsForUser(r.Context(), targetID)
	if err != nil {
		h.logger.Error("Failed to get reviews", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reviews)
}
