package generator

import "github.com/koriebruh/Genitz/internal/tui"

// configLibIDs are the DependencyRegistry IDs (internal/tui/dependencies.go)
// that get an automatic config/config.go stub — no toggle, purely derived
// from what's already selected, same spirit as Docker's compose mapping.
var configLibIDs = map[string]bool{
	"viper":     true,
	"godotenv":  true,
	"koanf":     true,
	"envconfig": true,
}

// detectConfigLib returns the single selected config-loader dependency, if
// exactly one was picked. Zero or more-than-one match is ambiguous — no
// guessing, nothing gets generated.
func detectConfigLib(deps map[int]tui.Dependency) (tui.Dependency, bool) {
	var match tui.Dependency
	count := 0
	for _, dep := range deps {
		if configLibIDs[dep.ID] {
			match = dep
			count++
		}
	}
	if count != 1 {
		return tui.Dependency{}, false
	}
	return match, true
}

// configStubContent returns a minimal config/config.go for dep, using only
// well-established, stable API for each library — kept deliberately simple
// rather than guessing at a fuller pattern.
func configStubContent(dep tui.Dependency) (string, bool) {
	switch dep.ID {
	case "viper":
		return `package config

import "github.com/spf13/viper"

type Config struct {
	AppEnv string
}

func Load() *Config {
	viper.AutomaticEnv()
	return &Config{
		AppEnv: viper.GetString("APP_ENV"),
	}
}
`, true

	case "godotenv":
		return `package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv string
}

func Load() *Config {
	_ = godotenv.Load()
	return &Config{
		AppEnv: os.Getenv("APP_ENV"),
	}
}
`, true

	case "envconfig":
		return `package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
	AppEnv string ` + "`envconfig:\"APP_ENV\"`" + `
}

func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
`, true

	case "koanf":
		return `package config

import (
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	AppEnv string
}

func Load() (*Config, error) {
	k := koanf.New(".")
	if err := k.Load(env.Provider(".", env.Opt{}), nil); err != nil {
		return nil, err
	}
	return &Config{
		AppEnv: k.String("APP_ENV"),
	}, nil
}
`, true
	}
	return "", false
}
