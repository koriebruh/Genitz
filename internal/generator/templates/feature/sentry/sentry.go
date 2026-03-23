package monitoring

import (
	"log"
	"time"

	"github.com/getsentry/sentry-go"
)

// SentryConfig holds configuration for Sentry error monitoring.
type SentryConfig struct {
	DSN              string  `env:"SENTRY_DSN"`
	Environment      string  `env:"SENTRY_ENVIRONMENT"    envDefault:"development"`
	SampleRate       float64 `env:"SENTRY_SAMPLE_RATE"    envDefault:"1.0"`
	TracesSampleRate float64 `env:"SENTRY_TRACES_RATE"    envDefault:"0.1"`
	Release          string  `env:"SENTRY_RELEASE"        envDefault:"1.0.0"`
}

// InitSentry initializes the Sentry SDK.
// Call this once at application startup.
func InitSentry(cfg SentryConfig) error {
	if cfg.DSN == "" {
		log.Println("⚠️  SENTRY_DSN not set, Sentry will be disabled")
		return nil
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.DSN,
		Environment:      cfg.Environment,
		Release:          cfg.Release,
		TracesSampleRate: cfg.TracesSampleRate,
		SampleRate:       cfg.SampleRate,
	})
	if err != nil {
		return err
	}

	log.Println("✅ Sentry initialized")
	return nil
}

// Flush waits for buffered events to be sent (call on app shutdown).
func Flush(timeout time.Duration) {
	sentry.Flush(timeout)
}

// CaptureError sends a Go error to Sentry with current context.
func CaptureError(err error) {
	sentry.CaptureException(err)
}
