package config

import (
	"log"
	"os"

	"github.com/caarlos0/env/v11"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LoggerConfig struct {
	Level      string `env:"LOG_LEVEL"      envDefault:"info"`
	Encoding   string `env:"LOG_ENCODING"   envDefault:"json"`
	OutputPath string `env:"LOG_OUTPUT"     envDefault:"stdout"`
}

func LoadLoggerConfig() LoggerConfig {
	cfg := LoggerConfig{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to load logger config: %v", err)
	}
	return cfg
}

// NewLogger creates a new production-ready zap.Logger.
func NewLogger(cfg LoggerConfig) *zap.Logger {
	level := zapcore.InfoLevel
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		log.Printf("invalid log level %q, defaulting to info", cfg.Level)
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "time"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder

	var encoder zapcore.Encoder
	if cfg.Encoding == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	}

	syncer := zapcore.AddSync(os.Stdout)

	core := zapcore.NewCore(encoder, syncer, level)
	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}
