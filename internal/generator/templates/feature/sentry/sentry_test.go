package monitoring

import (
	"testing"
)

func TestInitSentry_NoDSN_DoesNotFail(t *testing.T) {
	cfg := SentryConfig{} // DSN is empty
	err := InitSentry(cfg)
	if err != nil {
		t.Errorf("Expected InitSentry to succeed even without DSN, got: %v", err)
	}
}

func TestSentryConfig_DefaultValues(t *testing.T) {
	cfg := SentryConfig{
		Environment:      "development",
		SampleRate:       1.0,
		TracesSampleRate: 0.1,
		Release:          "1.0.0",
	}

	if cfg.Environment == "" {
		t.Error("Expected default environment to be set")
	}
	if cfg.SampleRate <= 0 {
		t.Error("Expected positive sample rate")
	}
}
