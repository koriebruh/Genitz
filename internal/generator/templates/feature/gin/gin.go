package config

import (
	"log"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/gin-gonic/gin"
)

type GinConfig struct {
	AppName      string        `env:"APP_NAME"       envDefault:"Genitz App"`
	Mode         string        `env:"GIN_MODE"       envDefault:"debug"`
	ReadTimeout  time.Duration `env:"READ_TIMEOUT"   envDefault:"10s"`
	WriteTimeout time.Duration `env:"WRITE_TIMEOUT"  envDefault:"10s"`
	Port         string        `env:"PORT"           envDefault:"8080"`
}

func LoadGinConfig() GinConfig {
	cfg := GinConfig{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to load gin config: %v", err)
	}
	return cfg
}

// NewGin creates a new Gin engine with production-ready configuration.
func NewGin(cfg GinConfig) *gin.Engine {
	gin.SetMode(cfg.Mode)

	r := gin.New()

	// Recovery middleware recovers from any panics and writes a 500 error.
	r.Use(gin.Recovery())

	// Logger middleware logs request/response details.
	r.Use(gin.Logger())

	return r
}
