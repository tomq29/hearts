package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	// AccessTokenLifetime is the duration for which the access token is valid.
	AccessTokenLifetime = 15 * time.Minute
	// RefreshTokenLifetime is the duration for which the refresh token is valid.
	RefreshTokenLifetime = 7 * 24 * time.Hour
	// RefreshTokenCookieLifetime is the duration for the refresh token cookie's "Expires" attribute.
	RefreshTokenCookieLifetime = 7 * 24 * time.Hour
)

// Claims represents the JWT claims.
type Claims struct {
	UserID string `json:"userId"`
	jwt.RegisteredClaims
}

// AuthService defines the interface for token operations.
type AuthService interface {
	GenerateTokens(userID string) (accessToken, refreshToken, refreshTokenHash string, refreshTokenExpiresAt time.Time, err error)
	ValidateAccessToken(tokenString string) (*Claims, error)
}

// jwtService is the implementation of AuthService.
type jwtService struct {
	secretKey []byte
}

// NewAuthService creates a new instance of jwtService.
func NewAuthService(secret string) (AuthService, error) {
	if secret == "" {
		return nil, fmt.Errorf("JWT secret key cannot be empty")
	}
	return &jwtService{secretKey: []byte(secret)}, nil
}

// GenerateTokens creates a new access token and a new refresh token.
func (s *jwtService) GenerateTokens(userID string) (string, string, string, time.Time, error) {
	// 1. Create Access Token
	accessTokenExp := time.Now().Add(AccessTokenLifetime)
	accessTokenClaims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessTokenExp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "hearts-api",
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	signedAccessToken, err := accessToken.SignedString(s.secretKey)
	if err != nil {
		return "", "", "", time.Time{}, err
	}

	// 2. Create Refresh Token
	refreshTokenBytes := make([]byte, 32)
	if _, err := rand.Read(refreshTokenBytes); err != nil {
		return "", "", "", time.Time{}, err
	}

	// 3. Hash the refresh token bytes before encoding them for the client.
	refreshTokenHash := sha256.Sum256(refreshTokenBytes)
	refreshTokenHashString := base64.URLEncoding.EncodeToString(refreshTokenHash[:])

	// 4. Encode the raw refresh token bytes for transmission to the client.
	refreshTokenString := base64.URLEncoding.EncodeToString(refreshTokenBytes)
	refreshTokenExpiresAt := time.Now().Add(RefreshTokenLifetime)

	return signedAccessToken, refreshTokenString, refreshTokenHashString, refreshTokenExpiresAt, nil
}

// ValidateAccessToken validates the given access token string.
func (s *jwtService) ValidateAccessToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
