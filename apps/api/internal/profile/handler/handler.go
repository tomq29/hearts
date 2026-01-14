package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/kisssonik/hearts/internal/profile"
	"github.com/kisssonik/hearts/internal/profile/repository"
	"github.com/kisssonik/hearts/internal/profile/service"
	"github.com/kisssonik/hearts/pkg/auth"
	"github.com/kisssonik/hearts/pkg/storage"
	"go.uber.org/zap"
)

// UploadPhotoResponse represents the response for photo upload.
type UploadPhotoResponse struct {
	Key string `json:"key"`
	URL string `json:"url"`
}

type ProfileHandler struct {
	service service.ProfileService
	storage storage.Provider
	logger  *zap.Logger
}

func NewProfileHandler(s service.ProfileService, storage storage.Provider, l *zap.Logger) *ProfileHandler {
	return &ProfileHandler{service: s, storage: storage, logger: l}
}

func (h *ProfileHandler) enrichProfile(ctx context.Context, p *profile.Profile) {
	if p == nil {
		return
	}
	for i, photoKey := range p.Photos {
		url, err := h.storage.GetPresignedURL(ctx, photoKey)
		if err != nil {
			h.logger.Error("Failed to generate presigned URL", zap.String("key", photoKey), zap.Error(err))
			continue
		}
		p.Photos[i] = url
	}
}

// UploadPhoto handles photo upload.
// @Summary Upload a photo
// @Description Upload a photo for the user profile.
// @Tags profiles
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth
// @Param photo formData file true "Photo file"
// @Success 200 {object} UploadPhotoResponse
// @Failure 400 {string} string "Invalid file"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /profiles/upload [post]
func (h *ProfileHandler) UploadPhoto(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Limit upload size to 10MB
	r.ParseMultipartForm(10 << 20)

	file, header, err := r.FormFile("photo")
	if err != nil {
		h.logger.Error("Failed to get file from form", zap.Error(err))
		http.Error(w, "Invalid file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Generate a unique filename
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("%s/%s%s", userID, uuid.New().String(), ext)

	// Upload to storage
	key, err := h.storage.Upload(r.Context(), file, header.Size, header.Header.Get("Content-Type"), filename)
	if err != nil {
		h.logger.Error("Failed to upload file", zap.Error(err))
		http.Error(w, "Failed to upload file", http.StatusInternalServerError)
		return
	}

	// Generate a presigned URL for immediate display (optional)
	// or just return the key/url. Since we are using MinIO locally,
	// we might want to return a presigned URL or a public URL if configured.
	// For now, let's return the key and a presigned URL.
	url, err := h.storage.GetPresignedURL(r.Context(), key)
	if err != nil {
		h.logger.Error("Failed to generate presigned URL", zap.Error(err))
		// We can still return the key, but the frontend might not be able to display it immediately without a URL.
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(UploadPhotoResponse{
		Key: key,
		URL: url,
	})
}

// Create handles profile creation.
// @Summary Create a profile
// @Description Create a new profile for the authenticated user.
// @Tags profiles
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param input body service.CreateProfileInput true "Profile creation input"
// @Success 201 {object} profile.Profile
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Unauthorized"
// @Failure 409 {string} string "Profile already exists"
// @Failure 500 {string} string "Internal server error"
// @Router /profiles [post]
func (h *ProfileHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var input service.CreateProfileInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if input.FirstName == "" {
		http.Error(w, "First name is required", http.StatusBadRequest)
		return
	}

	p, err := h.service.CreateProfile(r.Context(), userID, input)
	if err != nil {
		if errors.Is(err, service.ErrProfileAlreadyExists) {
			http.Error(w, "Profile already exists", http.StatusConflict)
			return
		}
		h.logger.Error("Failed to create profile", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.enrichProfile(r.Context(), p)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

// Get handles retrieving a profile by user ID.
// @Summary Get a profile
// @Description Get a user's profile by their user ID.
// @Tags profiles
// @Produce json
// @Param userID path string true "User ID"
// @Success 200 {object} profile.Profile
// @Failure 400 {string} string "User ID required"
// @Failure 404 {string} string "Profile not found"
// @Failure 500 {string} string "Internal server error"
// @Router /profiles/{userID} [get]
func (h *ProfileHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userID")
	if userID == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	p, err := h.service.GetProfileByUserID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "Profile not found", http.StatusNotFound)
			return
		}
		h.logger.Error("Failed to get profile", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.enrichProfile(r.Context(), p)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

// Update handles profile updates.
// @Summary Update a profile
// @Description Update the authenticated user's profile.
// @Tags profiles
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param input body service.UpdateProfileInput true "Profile update input"
// @Success 200 {object} profile.Profile
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Profile not found"
// @Failure 500 {string} string "Internal server error"
// @Router /profiles [put]
func (h *ProfileHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var input service.UpdateProfileInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	p, err := h.service.UpdateProfile(r.Context(), userID, input)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "Profile not found", http.StatusNotFound)
			return
		}
		h.logger.Error("Failed to update profile", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.enrichProfile(r.Context(), p)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

// Search handles searching for profiles.
// @Summary Search profiles
// @Description Search for profiles based on age, gender, height and location.
// @Tags profiles
// @Produce json
// @Security ApiKeyAuth
// @Param minAge query int false "Minimum age"
// @Param maxAge query int false "Maximum age"
// @Param gender query string false "Gender"
// @Param minHeight query int false "Minimum height"
// @Param maxHeight query int false "Maximum height"
// @Param radius query float64 false "Radius in KM"
// @Success 200 {array} profile.Profile
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /profiles/search [get]
func (h *ProfileHandler) Search(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var params service.SearchParams
	query := r.URL.Query()

	if minAgeStr := query.Get("minAge"); minAgeStr != "" {
		var minAge int
		if _, err := fmt.Sscanf(minAgeStr, "%d", &minAge); err == nil {
			params.MinAge = &minAge
		}
	}

	if maxAgeStr := query.Get("maxAge"); maxAgeStr != "" {
		var maxAge int
		if _, err := fmt.Sscanf(maxAgeStr, "%d", &maxAge); err == nil {
			params.MaxAge = &maxAge
		}
	}

	if gender := query.Get("gender"); gender != "" {
		params.Gender = &gender
	}

	if minHeightStr := query.Get("minHeight"); minHeightStr != "" {
		var minHeight int
		if _, err := fmt.Sscanf(minHeightStr, "%d", &minHeight); err == nil {
			params.MinHeight = &minHeight
		}
	}

	if maxHeightStr := query.Get("maxHeight"); maxHeightStr != "" {
		var maxHeight int
		if _, err := fmt.Sscanf(maxHeightStr, "%d", &maxHeight); err == nil {
			params.MaxHeight = &maxHeight
		}
	}

	if radiusStr := query.Get("radius"); radiusStr != "" {
		var radius float64
		if _, err := fmt.Sscanf(radiusStr, "%f", &radius); err == nil {
			params.RadiusKM = &radius
		}
	}

	profiles, err := h.service.SearchProfiles(r.Context(), userID, params)
	if err != nil {
		h.logger.Error("Failed to search profiles", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	for _, p := range profiles {
		h.enrichProfile(r.Context(), p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profiles)
}
