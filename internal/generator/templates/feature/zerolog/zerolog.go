package logger

import (
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// ZerologConfig holds configuration for the Zerolog logger.
type ZerologConfig struct {
	Level  string `env:"LOG_LEVEL"  envDefault:"info"`
	Format string `env:"LOG_FORMAT" envDefault:"json"` // "json" or "console"
}

// NewZerolog creates a configured zerolog.Logger.
// It returns the logger and a sync function to flush pending output.
func NewZerolog(cfg ZerologConfig) zerolog.Logger {
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}

	var writer io.Writer = os.Stdout
	if cfg.Format == "console" {
		writer = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "15:04:05"}
	}

	logger := zerolog.New(writer).
		Level(level).
		With().
		Timestamp().
		Caller().
		Logger()

	// Replace the global logger for packages that use log.Logger directly
	log.Logger = logger

	return logger
}

// NewZerologWithOutput allows injecting a custom writer for testability.
func NewZerologWithOutput(cfg ZerologConfig, writer io.Writer) zerolog.Logger {
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	return zerolog.New(writer).Level(level).With().Timestamp().Logger()
}
