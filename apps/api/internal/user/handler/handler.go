package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/kisssonik/hearts/internal/user"
	"github.com/kisssonik/hearts/internal/user/repository"
	"github.com/kisssonik/hearts/internal/user/service"
	"github.com/kisssonik/hearts/pkg/auth"
)

// RegisterInput represents the input for registration.
type RegisterInput struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	User        *user.User `json:"user"`
	AccessToken string     `json:"accessToken"`
}

// LoginInput represents the input for login.
type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents the response for login.
type LoginResponse struct {
	AccessToken string `json:"accessToken"`
}

// UserHandler handles HTTP requests for user-related actions.
type UserHandler struct {
	service     service.UserService
	authService auth.AuthService
	logger      *zap.Logger
}

// NewUserHandler creates a new instance of UserHandler.
func NewUserHandler(s service.UserService, as auth.AuthService, l *zap.Logger) *UserHandler {
	return &UserHandler{
		service:     s,
		authService: as,
		logger:      l,
	}
}

// Register is the handler for user registration.
// @Summary Register a new user
// @Description Register a new user with email, username, and password.
// @Tags users
// @Accept json
// @Produce json
// @Param input body RegisterInput true "Registration input"
// @Success 201 {object} user.User
// @Failure 400 {string} string "Invalid request body"
// @Failure 409 {string} string "Conflict"
// @Failure 500 {string} string "Internal server error"
// @Router /users/register [post]
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input RegisterInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Basic validation
	if input.Email == "" || input.Username == "" || input.Password == "" {
		h.logger.Warn("Validation failed for registration", zap.String("email", input.Email), zap.String("username", input.Username))
		http.Error(w, "Email, username, and password are required", http.StatusBadRequest)
		return
	}

	user, err := h.service.RegisterUser(r.Context(), input.Email, input.Username, input.Password)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateEmailOrUsername) {
			h.logger.Warn("Registration conflict", zap.String("email", input.Email), zap.String("username", input.Username), zap.Error(err))
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		h.logger.Error("Failed to register user", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("User registered successfully", zap.String("userID", user.ID))

	// Auto-login after registration
	loginResp, err := h.service.Login(r.Context(), input.Email, input.Password)
	if err != nil {
		h.logger.Error("Auto-login failed after registration", zap.Error(err))
		// Fallback: return success but force user to login manually.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(user)
		return
	}

	// Set the refresh token in a secure, HttpOnly cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    loginResp.RefreshToken,
		Expires:  time.Now().Add(auth.RefreshTokenCookieLifetime),
		HttpOnly: true,
		Secure:   true,                   // Should be true in production
		Path:     "/api/v1/auth/refresh", // Make sure this path is correct for your refresh endpoint
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(RegisterResponse{
		User:        user,
		AccessToken: loginResp.AccessToken,
	})
}

// Login is the handler for user authentication.
// @Summary Login a user
// @Description Login a user with email and password.
// @Tags users
// @Accept json
// @Produce json
// @Param input body LoginInput true "Login input"
// @Success 200 {object} LoginResponse
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /users/login [post]
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input LoginInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Error("Failed to decode login request body", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	loginResp, err := h.service.Login(r.Context(), input.Email, input.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			h.logger.Warn("Invalid login attempt", zap.String("email", input.Email), zap.Error(err))
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		h.logger.Error("Login failed", zap.String("email", input.Email), zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set the refresh token in a secure, HttpOnly cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    loginResp.RefreshToken,
		Expires:  time.Now().Add(auth.RefreshTokenCookieLifetime),
		HttpOnly: true,
		Secure:   true,                   // Should be true in production
		Path:     "/api/v1/auth/refresh", // Make sure this path is correct for your refresh endpoint
		SameSite: http.SameSiteLaxMode,
	})

	h.logger.Info("User logged in successfully", zap.String("email", input.Email))
	// Return the access token in the response body
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LoginResponse{
		AccessToken: loginResp.AccessToken,
	})
}

// RefreshToken is the handler for refreshing an access token.
// @Summary Refresh access token
// @Description Refresh the access token using the refresh token cookie.
// @Tags auth
// @Produce json
// @Success 200 {object} LoginResponse
// @Failure 401 {string} string "Unauthorized"
// @Router /auth/refresh [post]
func (h *UserHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// 1. Get the refresh token from the cookie.
	cookie, err := r.Cookie("refreshToken")
	if err != nil {
		h.logger.Warn("Refresh token cookie not found", zap.Error(err))
		http.Error(w, "Refresh token not found", http.StatusUnauthorized)
		return
	}
	refreshToken := cookie.Value

	// 2. Call the service to validate the token and get new ones.
	loginResp, err := h.service.RefreshToken(r.Context(), refreshToken)
	if err != nil {
		h.logger.Warn("Failed to refresh token", zap.Error(err))
		// This could be due to an invalid, expired, or already used token.
		http.Error(w, "Invalid or expired refresh token", http.StatusUnauthorized)
		return
	}

	// 3. Set the new refresh token in the cookie.
	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    loginResp.RefreshToken,
		Expires:  time.Now().Add(auth.RefreshTokenCookieLifetime),
		HttpOnly: true,
		Secure:   true,
		Path:     "/api/v1/auth/refresh",
		SameSite: http.SameSiteLaxMode,
	})

	h.logger.Info("Token refreshed successfully")
	// 4. Return the new access token in the response body.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LoginResponse{
		AccessToken: loginResp.AccessToken,
	})
}

// Me is the handler for retrieving the current user's profile.
// @Summary Get current user
// @Description Get the profile of the currently authenticated user.
// @Tags users
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} user.User
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "User not found"
// @Failure 500 {string} string "Internal server error"
// @Router /users/me [get]
func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok {
		h.logger.Error("User ID not found in context")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	user, err := h.service.GetUserByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.logger.Warn("User not found", zap.String("userID", userID))
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		h.logger.Error("Failed to get user", zap.String("userID", userID), zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}
