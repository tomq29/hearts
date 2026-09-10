package auth

import (
	"errors"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// ErrUnauthorized is returned for every rejection, so callers cannot leak which
// specific check failed back to the client.
var ErrUnauthorized = errors.New("unauthorized")

// claims mirrors the API's token claims. The userId key must stay in step with
// apps/api/pkg/auth/auth.go — this service verifies tokens that service mints,
// and the two modules cannot share code (see the Dockerfile build context).
type claims struct {
	UserID string `json:"userId"`
	jwt.RegisteredClaims
}

// Verifier validates access tokens issued by the API.
type Verifier struct {
	secret []byte
}

// NewVerifier fails on an empty secret rather than silently accepting nothing:
// a service that verifies against "" would authenticate every request.
func NewVerifier(secret string) (*Verifier, error) {
	if secret == "" {
		return nil, fmt.Errorf("jwt secret cannot be empty")
	}
	return &Verifier{secret: []byte(secret)}, nil
}

// UserID extracts and verifies the bearer token from an Authorization header,
// returning the user it belongs to.
func (v *Verifier) UserID(authorizationHeader string) (string, error) {
	rawToken, err := bearerToken(authorizationHeader)
	if err != nil {
		return "", err
	}

	var c claims
	token, err := jwt.ParseWithClaims(rawToken, &c, func(t *jwt.Token) (interface{}, error) {
		// Without this check a token signed with "alg: none", or one signed
		// with our secret treated as an RSA public key, would be accepted.
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return v.secret, nil
	})
	if err != nil || !token.Valid {
		return "", ErrUnauthorized
	}
	if c.UserID == "" {
		return "", ErrUnauthorized
	}

	return c.UserID, nil
}

func bearerToken(header string) (string, error) {
	parts := strings.Split(header, " ")
	if len(parts) != 2 || parts[0] != "Bearer" || parts[1] == "" {
		return "", ErrUnauthorized
	}
	return parts[1], nil
}
