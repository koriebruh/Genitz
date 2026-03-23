package router

import (
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestNewFiberApp(t *testing.T) {
	cfg := ServerConfig{
		Port:         "3000",
		ReadTimeout:  10,
		WriteTimeout: 10,
	}

	app := NewFiberApp(cfg)

	if app == nil {
		t.Fatal("Expected NewFiberApp to return a valid Fiber App instance")
	}

	if _, ok := interface{}(app).(*fiber.App); !ok {
		t.Fatal("Expected returned instance to be of type *fiber.App")
	}
}
