package logger

import (
	"io"
	"log"
	"os"

	"github.com/sirupsen/logrus"
)

// LogrusConfig holds configuration for the Logrus logger.
type LogrusConfig struct {
	Level  string `env:"LOG_LEVEL"  envDefault:"info"`
	Format string `env:"LOG_FORMAT" envDefault:"json"`
}

// NewLogrus creates a configured *logrus.Logger ready for injection.
func NewLogrus(cfg LogrusConfig) *logrus.Logger {
	l := logrus.New()

	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	l.SetLevel(level)

	if cfg.Format == "console" || cfg.Format == "text" {
		l.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	} else {
		l.SetFormatter(&logrus.JSONFormatter{})
	}

	l.SetOutput(os.Stdout)
	return l
}

// NewLogrusWithOutput allows injecting a custom io.Writer — useful in tests.
func NewLogrusWithOutput(cfg LogrusConfig, out io.Writer) *logrus.Logger {
	l := NewLogrus(cfg)
	l.SetOutput(out)
	return l
}

// WithField is a convenience wrapper that returns a *logrus.Entry.
func WithField(l *logrus.Logger, key, value string) *logrus.Entry {
	if l == nil {
		log.Panic("logrus.Logger is nil")
	}
	return l.WithField(key, value)
}
