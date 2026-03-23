package logger

import (
	"testing"
)

func TestNewLogger_JSON(t *testing.T) {
	cfg := LoggerConfig{Level: "info", Format: "json"}
	logger := NewLogger(cfg)
	if logger == nil {
		t.Fatal("Expected non-nil *zap.Logger")
	}
	defer logger.Sync() //nolint:errcheck
	logger.Info("TestNewLogger_JSON passed")
}

func TestNewLogger_Console(t *testing.T) {
	cfg := LoggerConfig{Level: "debug", Format: "console"}
	logger := NewLogger(cfg)
	if logger == nil {
		t.Fatal("Expected non-nil *zap.Logger (console mode)")
	}
	defer logger.Sync() //nolint:errcheck
}

func TestNewLogger_InvalidLevel(t *testing.T) {
	cfg := LoggerConfig{Level: "invalid-level", Format: "json"}
	// Should fallback to info, not panic
	logger := NewLogger(cfg)
	if logger == nil {
		t.Fatal("Expected non-nil *zap.Logger even with invalid log level")
	}
	defer logger.Sync() //nolint:errcheck
}
