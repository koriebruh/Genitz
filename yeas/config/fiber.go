package config

import (
	"log"
	"time"

	"github.com/bytedance/sonic"
	"github.com/caarlos0/env/v11"
	"github.com/gofiber/fiber/v3"
)

type FiberConfig struct {
	AppName      string        `env:"APP_NAME"       envDefault:"Genitz App"`
	BodyLimitMB  int           `env:"BODY_LIMIT_MB"  envDefault:"4"`
	ReadTimeout  time.Duration `env:"READ_TIMEOUT"   envDefault:"10s"`
	WriteTimeout time.Duration `env:"WRITE_TIMEOUT"  envDefault:"10s"`
	IdleTimeout  time.Duration `env:"IDLE_TIMEOUT"   envDefault:"120s"`
	ProxyHeader  string        `env:"PROXY_HEADER"`
	TrustProxy   bool          `env:"TRUST_PROXY"    envDefault:"false"`
}

func LoadFiberConfig() FiberConfig {
	cfg := FiberConfig{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to load fiber config: %v", err)
	}
	return cfg
}

// NewFiber creates a new Fiber app with production-ready configuration.
func NewFiber(cfg FiberConfig) *fiber.App {
	return fiber.New(fiber.Config{
		AppName:       cfg.AppName,
		ServerHeader:  "",
		StrictRouting: true,
		CaseSensitive: true,

		// Body & Buffer
		BodyLimit:       cfg.BodyLimitMB * 1024 * 1024,
		ReadBufferSize:  8192,
		WriteBufferSize: 8192,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		IdleTimeout:     cfg.IdleTimeout,
		TrustProxy:      cfg.TrustProxy,
		ProxyHeader:     cfg.ProxyHeader,
		TrustProxyConfig: fiber.TrustProxyConfig{
			Proxies: []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"},
		},

		// Performance
		JSONEncoder:        sonic.Marshal,
		JSONDecoder:        sonic.Unmarshal,
		Immutable:          true,
		EnableIPValidation: true,
		ErrorHandler:       fiberErrorHandler,
	})
}

// fiberErrorHandler is a custom error handler that prevents raw errors from being exposed to clients.
func fiberErrorHandler(c fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "internal server error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	return c.Status(code).JSON(fiber.Map{
		"success": false,
		"error":   message,
	})
}
