package auth_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kisssonik/hearts/apps/notification-service/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "test-secret"

// signToken builds a token the same way the API does, so these tests exercise
// the real shape rather than an invented one.
func signToken(t *testing.T, secret string, userID string, expiresIn time.Duration) string {
	t.Helper()
	claims := jwt.MapClaims{
		"userId": userID,
		"exp":    jwt.NewNumericDate(time.Now().Add(expiresIn)),
		"iat":    jwt.NewNumericDate(time.Now()),
		"iss":    "hearts-api",
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	require.NoError(t, err)
	return signed
}

func TestNewVerifier_RejectsEmptySecret(t *testing.T) {
	_, err := auth.NewVerifier("")
	assert.Error(t, err)
}

func TestUserID_ValidToken(t *testing.T) {
	v, err := auth.NewVerifier(testSecret)
	require.NoError(t, err)

	token := signToken(t, testSecret, "user-123", time.Minute)

	userID, err := v.UserID("Bearer " + token)
	assert.NoError(t, err)
	assert.Equal(t, "user-123", userID)
}

func TestUserID_Rejections(t *testing.T) {
	v, err := auth.NewVerifier(testSecret)
	require.NoError(t, err)

	t.Run("expired token", func(t *testing.T) {
		token := signToken(t, testSecret, "user-123", -time.Minute)
		_, err := v.UserID("Bearer " + token)
		assert.Error(t, err)
	})

	t.Run("signed with a different secret", func(t *testing.T) {
		token := signToken(t, "someone-elses-secret", "user-123", time.Minute)
		_, err := v.UserID("Bearer " + token)
		assert.Error(t, err)
	})

	// The classic JWT bypass: an attacker strips the signature and sets
	// alg to none. Accepting it would authenticate anyone as anyone.
	t.Run("alg none is refused", func(t *testing.T) {
		claims := jwt.MapClaims{
			"userId": "attacker",
			"exp":    jwt.NewNumericDate(time.Now().Add(time.Hour)),
		}
		unsigned, err := jwt.NewWithClaims(jwt.SigningMethodNone, claims).
			SignedString(jwt.UnsafeAllowNoneSignatureType)
		require.NoError(t, err)

		_, err = v.UserID("Bearer " + unsigned)
		assert.Error(t, err)
	})

	t.Run("token without a userId claim", func(t *testing.T) {
		claims := jwt.MapClaims{"exp": jwt.NewNumericDate(time.Now().Add(time.Hour))}
		token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(testSecret))
		require.NoError(t, err)

		_, err = v.UserID("Bearer " + token)
		assert.Error(t, err)
	})

	t.Run("malformed header", func(t *testing.T) {
		valid := signToken(t, testSecret, "user-123", time.Minute)

		for _, header := range []string{
			"",
			"Bearer",
			"Bearer ",
			valid,            // no scheme
			"Basic " + valid, // wrong scheme
			"Bearer not-a-token",
		} {
			_, err := v.UserID(header)
			assert.Error(t, err, "header %q must be rejected", header)
		}
	})
}
