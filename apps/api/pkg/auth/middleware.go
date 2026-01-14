package auth

import (
	"context"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

type contextKey string

const UserIDKey contextKey = "userID"

// Middleware creates a new authentication middleware.
func Middleware(authService AuthService, logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header is required", http.StatusUnauthorized)
				return
			}

			headerParts := strings.Split(authHeader, " ")
			if len(headerParts) != 2 || headerParts[0] != "Bearer" {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			tokenString := headerParts[1]
			claims, err := authService.ValidateAccessToken(tokenString)
			if err != nil {
				logger.Warn("Invalid access token", zap.Error(err))
				http.Error(w, "Invalid access token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
