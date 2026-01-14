package logger

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config controls how the zap logger is initialised.
type Config struct {
	Level string
	Mode  string
}

// New creates and configures a new zap logger based on the supplied configuration.
func New(cfg Config) (*zap.Logger, error) {
	var zapConfig zap.Config

	switch strings.ToLower(cfg.Mode) {
	case "production", "prod":
		zapConfig = zap.NewProductionConfig()
	default:
		zapConfig = zap.NewDevelopmentConfig()
	}

	if cfg.Level != "" {
		var level zapcore.Level
		if err := level.UnmarshalText([]byte(strings.ToLower(cfg.Level))); err != nil {
			return nil, fmt.Errorf("invalid log level %q: %w", cfg.Level, err)
		}
		zapConfig.Level = zap.NewAtomicLevelAt(level)
	}

	return zapConfig.Build()
}
