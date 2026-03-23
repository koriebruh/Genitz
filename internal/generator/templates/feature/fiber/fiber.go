package router

import (
	"github.com/gofiber/fiber/v3"
)

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Port         string `env:"APP_PORT"     envDefault:"8080"`
	ReadTimeout  int    `env:"READ_TIMEOUT" envDefault:"30"`
	WriteTimeout int    `env:"WRITE_TIMEOUT" envDefault:"30"`
}

// NewFiberApp creates a configured *fiber.App instance.
// Inject this as your HTTP server and register routes against it.
func NewFiberApp(cfg ServerConfig) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: "Genitz App",
	})

	// Default health-check endpoint
	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": cfg.Port,
		})
	})

	return app
}
