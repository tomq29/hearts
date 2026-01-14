package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	mapstructure "github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

// Config captures application configuration derived from files and environment variables.
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Logger   LoggerConfig   `mapstructure:"logger"`
	Storage  StorageConfig  `mapstructure:"storage"`
	Kafka    KafkaConfig    `mapstructure:"kafka"`
}

// AppConfig captures application-wide settings.
type AppConfig struct {
	Environment string `mapstructure:"environment"`
}

// ServerConfig contains HTTP server settings.
type ServerConfig struct {
	Address string `mapstructure:"address"`
}

// DatabaseConfig contains Postgres connection pool settings.
type DatabaseConfig struct {
	URL             string        `mapstructure:"url"`
	MaxConnections  int32         `mapstructure:"max_connections"`
	MinConnections  int32         `mapstructure:"min_connections"`
	MaxConnLifetime time.Duration `mapstructure:"max_conn_lifetime"`
}

// AuthConfig contains authentication-related configuration.
type AuthConfig struct {
	JWTSecret string `mapstructure:"jwt_secret"`
}

// LoggerConfig describes the zap logger configuration.
type LoggerConfig struct {
	Level string `mapstructure:"level"`
	Mode  string `mapstructure:"mode"`
}

// StorageConfig contains MinIO/S3 settings.
type StorageConfig struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	UseSSL          bool   `mapstructure:"use_ssl"`
	BucketName      string `mapstructure:"bucket_name"`
	Location        string `mapstructure:"location"`
}

// KafkaConfig contains Kafka connection settings.
type KafkaConfig struct {
	Brokers []string `mapstructure:"brokers"`
	Topic   string   `mapstructure:"topic"`
	GroupID string   `mapstructure:"group_id"`
}

// Load reads configuration using Viper, applying sane defaults and environment overrides.
func Load() (Config, error) {
	v := viper.New()

	if configFile := os.Getenv("CONFIG_FILE"); configFile != "" {
		v.SetConfigFile(configFile)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./config")
		v.AddConfigPath("./configs")
	}

	v.SetEnvPrefix("HEARTS")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Defaults
	v.SetDefault("app.environment", "development")
	v.SetDefault("server.address", ":8080")
	v.SetDefault("database.max_connections", 10)
	v.SetDefault("database.min_connections", 2)
	v.SetDefault("database.max_conn_lifetime", "1h")
	v.SetDefault("logger.level", "info")
	v.SetDefault("logger.mode", "development")
	v.SetDefault("storage.endpoint", "localhost:9000")
	v.SetDefault("storage.access_key_id", "minioadmin")
	v.SetDefault("storage.secret_access_key", "minioadmin")
	v.SetDefault("storage.use_ssl", false)
	v.SetDefault("storage.bucket_name", "hearts-photos")
	v.SetDefault("storage.location", "us-east-1")
	v.SetDefault("kafka.brokers", []string{"kafka:29092"})
	v.SetDefault("kafka.topic", "match-checks")
	v.SetDefault("kafka.group_id", "hearts-match-checker")

	// Explicit environment bindings for commonly overridden keys.
	_ = v.BindEnv("database.url", "DATABASE_URL")
	_ = v.BindEnv("auth.jwt_secret", "JWT_SECRET_KEY")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return Config{}, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg, viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
	))); err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	if cfg.Auth.JWTSecret == "" {
		return Config{}, fmt.Errorf("jwt secret not configured; set auth.jwt_secret or JWT_SECRET_KEY")
	}

	return cfg, nil
}
