package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewZerolog_JSON(t *testing.T) {
	cfg := ZerologConfig{Level: "info", Format: "json"}
	l := NewZerolog(cfg)
	// zerolog.Logger is a value type — check it writes
	var buf bytes.Buffer
	tl := NewZerologWithOutput(cfg, &buf)
	tl.Info().Msg("test message")

	_ = l // ensure constructor runs without panic
	if !strings.Contains(buf.String(), "test message") {
		t.Errorf("Expected log output, got: %s", buf.String())
	}
}

func TestNewZerolog_InvalidLevel(t *testing.T) {
	cfg := ZerologConfig{Level: "invalid", Format: "json"}
	var buf bytes.Buffer
	l := NewZerologWithOutput(cfg, &buf)
	l.Info().Msg("fallback works")

	if !strings.Contains(buf.String(), "fallback works") {
		t.Errorf("Expected log output after invalid level fallback, got: %s", buf.String())
	}
}
