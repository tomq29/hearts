package config_test

import (
	"testing"

	"github.com/kisssonik/hearts/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("JWT_SECRET_KEY", "test-secret")

	cfg, err := config.Load()
	require.NoError(t, err)

	assert.Equal(t, []string{"kafka:29092"}, cfg.Kafka.Brokers)
	assert.Equal(t, []string{"http://localhost:5173", "http://localhost:4173"}, cfg.Server.CORSAllowedOrigins)
	assert.Equal(t, []string{"5", "10", "25", "50"}, cfg.Profile.DistanceBands)
}

func TestLoad_DistanceBandsFromEnv(t *testing.T) {
	t.Setenv("JWT_SECRET_KEY", "test-secret")
	t.Setenv("HEARTS_PROFILE_DISTANCE_BANDS", "2,10,50")

	cfg, err := config.Load()
	require.NoError(t, err)

	assert.Equal(t, []string{"2", "10", "50"}, cfg.Profile.DistanceBands)
}

func TestLoad_EnvOverridesWithCommaSeparatedLists(t *testing.T) {
	t.Setenv("JWT_SECRET_KEY", "test-secret")
	// The plain form docker-compose passes (no HEARTS_ prefix).
	t.Setenv("KAFKA_BROKERS", "broker1:9092,broker2:9092")
	t.Setenv("HEARTS_SERVER_CORS_ALLOWED_ORIGINS", "https://hearts.example.com")

	cfg, err := config.Load()
	require.NoError(t, err)

	assert.Equal(t, []string{"broker1:9092", "broker2:9092"}, cfg.Kafka.Brokers)
	assert.Equal(t, []string{"https://hearts.example.com"}, cfg.Server.CORSAllowedOrigins)
}

// Regression test: storage.secret_access_key has no SetDefault, and viper's
// AutomaticEnv only resolves keys it already knows about — without the
// explicit BindEnv the env var would be silently ignored.
func TestLoad_StorageSecretFromEnv(t *testing.T) {
	t.Setenv("JWT_SECRET_KEY", "test-secret")
	t.Setenv("HEARTS_STORAGE_SECRET_ACCESS_KEY", "s3-secret-from-env")

	cfg, err := config.Load()
	require.NoError(t, err)
	assert.Equal(t, "s3-secret-from-env", cfg.Storage.SecretAccessKey)
}

func TestLoad_MissingJWTSecretFails(t *testing.T) {
	cfg, err := config.Load()
	assert.Error(t, err)
	assert.Empty(t, cfg.Auth.JWTSecret)
}
