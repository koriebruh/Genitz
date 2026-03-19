package config

import (
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
)

type FiberConfig struct {
	AppName      string
	BodyLimitMB  int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	ProxyHeader  string // isi kalau di belakang reverse proxy (nginx, cloudflare)
	TrustProxy   bool
}

func LoadFiberConfig() FiberConfig {
	bodyLimit, _ := strconv.Atoi(getFiberEnv("BODY_LIMIT_MB", "4"))
	trustProxy, _ := strconv.ParseBool(getFiberEnv("TRUST_PROXY", "false"))

	return FiberConfig{
		AppName:      getFiberEnv("APP_NAME", "Genitz App"),
		BodyLimitMB:  bodyLimit,
		ReadTimeout:  parseFiberDuration(getFiberEnv("READ_TIMEOUT", "10s")),
		WriteTimeout: parseFiberDuration(getFiberEnv("WRITE_TIMEOUT", "10s")),
		IdleTimeout:  parseFiberDuration(getFiberEnv("IDLE_TIMEOUT", "120s")),
		ProxyHeader:  getFiberEnv("PROXY_HEADER", ""),
		TrustProxy:   trustProxy,
	}
}

func getFiberEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func parseFiberDuration(val string) time.Duration {
	d, _ := time.ParseDuration(val)
	return d
}

func NewFiber(cfg FiberConfig) *fiber.App {
	return fiber.New(fiber.Config{
		AppName:      cfg.AppName,
		BodyLimit:    cfg.BodyLimitMB * 1024 * 1024,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
		ProxyHeader:  cfg.ProxyHeader,
	})
}
