package config

import (
	"os"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
)

func TestLoadFiberConfig(t *testing.T) {
	// Set environment variables for testing
	os.Setenv("BODY_LIMIT_MB", "10")
	os.Setenv("TRUST_PROXY", "true")
	os.Setenv("READ_TIMEOUT", "5s")

	defer func() {
		os.Unsetenv("BODY_LIMIT_MB")
		os.Unsetenv("TRUST_PROXY")
		os.Unsetenv("READ_TIMEOUT")
	}()

	cfg := LoadFiberConfig()

	if cfg.BodyLimitMB != 10 {
		t.Errorf("Expected BodyLimitMB to be 10, got %d", cfg.BodyLimitMB)
	}

	if cfg.TrustProxy != true {
		t.Errorf("Expected TrustProxy to be true, got %v", cfg.TrustProxy)
	}

	if cfg.ReadTimeout != 5*time.Second {
		t.Errorf("Expected ReadTimeout to be 5s, got %v", cfg.ReadTimeout)
	}
}

func TestNewFiberApp(t *testing.T) {
	cfg := FiberConfig{
		AppName:      "Test App",
		BodyLimitMB:  1,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
		IdleTimeout:  1 * time.Second,
	}

	app := NewFiber(cfg)

	if app == nil {
		t.Fatal("Expected NewFiber to return a valid Fiber App instance")
	}

	if _, ok := interface{}(app).(*fiber.App); !ok {
		t.Fatal("Expected returned instance to be of type *fiber.App")
	}
}
