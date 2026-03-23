package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewLogrus_JSON(t *testing.T) {
	cfg := LogrusConfig{Level: "info", Format: "json"}
	l := NewLogrus(cfg)
	if l == nil {
		t.Fatal("Expected non-nil *logrus.Logger")
	}
}

func TestNewLogrus_Text(t *testing.T) {
	cfg := LogrusConfig{Level: "debug", Format: "console"}
	l := NewLogrus(cfg)
	if l == nil {
		t.Fatal("Expected non-nil *logrus.Logger (text mode)")
	}
}

func TestNewLogrus_InvalidLevel_FallsBackToInfo(t *testing.T) {
	cfg := LogrusConfig{Level: "bogus", Format: "json"}
	l := NewLogrus(cfg)
	if l == nil {
		t.Fatal("Expected non-nil logger even with invalid level")
	}
}

func TestNewLogrusWithOutput_WritesToBuffer(t *testing.T) {
	var buf bytes.Buffer
	cfg := LogrusConfig{Level: "info", Format: "json"}
	l := NewLogrusWithOutput(cfg, &buf)

	l.Info("hello from test")

	if !strings.Contains(buf.String(), "hello from test") {
		t.Errorf("Expected log output to contain message, got: %s", buf.String())
	}
}
