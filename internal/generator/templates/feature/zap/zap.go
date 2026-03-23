package logger

import (
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LoggerConfig holds configuration for the Zap logger.
type LoggerConfig struct {
	// Level: "debug", "info", "warn", "error"
	Level string `env:"LOG_LEVEL" envDefault:"info"`
	// Format: "json" or "console"
	Format string `env:"LOG_FORMAT" envDefault:"json"`
}

// NewLogger creates a configured *zap.Logger.
// In production, JSON format is recommended.
// In development (console), a human-friendly format is used instead.
func NewLogger(cfg LoggerConfig) *zap.Logger {
	var zapCfg zap.Config

	if cfg.Format == "console" {
		zapCfg = zap.NewDevelopmentConfig()
	} else {
		zapCfg = zap.NewProductionConfig()
	}

	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}
	zapCfg.Level = zap.NewAtomicLevelAt(level)

	logger, err := zapCfg.Build()
	if err != nil {
		log.Fatalf("failed to initialize zap logger: %v", err)
	}

	return logger
}
