package handler

import (
	"encoding/json"
	"net/http"

	"github.com/kisssonik/hearts/internal/like/service"
	"github.com/kisssonik/hearts/internal/profile"
	"github.com/kisssonik/hearts/pkg/auth"
	"github.com/kisssonik/hearts/pkg/storage"
	"go.uber.org/zap"
)

// LikeResponse represents the response for a like action.
type LikeResponse struct {
	IsMatch bool `json:"isMatch"`
}

type LikeHandler struct {
	service service.LikeService
	storage storage.Provider
	logger  *zap.Logger
}

func NewLikeHandler(s service.LikeService, storage storage.Provider, l *zap.Logger) *LikeHandler {
	return &LikeHandler{service: s, storage: storage, logger: l}
}

// Like handles liking or passing a user.
// @Summary Like or pass a user
// @Description Like or pass a user. Returns isMatch=true if it's a mutual like.
// @Tags likes
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param input body service.LikeInput true "Like input"
// @Success 200 {object} LikeResponse
// @Failure 400 {string} string "Invalid request body or self-like"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /likes [post]
func (h *LikeHandler) Like(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var input service.LikeInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	isMatch, err := h.service.LikeUser(r.Context(), userID, input)
	if err != nil {
		if err == service.ErrSelfLike {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.logger.Error("Failed to process like", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LikeResponse{
		IsMatch: isMatch,
	})
}

// GetMatches handles retrieving matches.
// @Summary Get matches
// @Description Get a list of profiles that the user has matched with.
// @Tags likes
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} profile.Profile
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /matches [get]
func (h *LikeHandler) GetMatches(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	matches, err := h.service.GetMatches(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get matches", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Enrich matches with presigned URLs
	for _, p := range matches {
		for i, photoKey := range p.Photos {
			url, err := h.storage.GetPresignedURL(r.Context(), photoKey)
			if err != nil {
				h.logger.Error("Failed to generate presigned URL", zap.String("key", photoKey), zap.Error(err))
				continue
			}
			p.Photos[i] = url
		}
	}

	if matches == nil {
		matches = []*profile.Profile{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(matches)
}
