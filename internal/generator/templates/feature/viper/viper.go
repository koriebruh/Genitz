package config

import (
	"log"

	"github.com/spf13/viper"
)

// AppConfig is the top-level application configuration struct.
// Add your own fields as your application grows.
type AppConfig struct {
	AppName  string `mapstructure:"APP_NAME"`
	AppEnv   string `mapstructure:"APP_ENV"`
	AppPort  string `mapstructure:"APP_PORT"`
	LogLevel string `mapstructure:"LOG_LEVEL"`
}

// NewViper creates and returns a fully-initialized *viper.Viper instance.
// It reads from environment variables AND a config file (.env / config.yaml).
func NewViper() *viper.Viper {
	v := viper.New()

	// Set defaults
	v.SetDefault("APP_NAME", "genitz-app")
	v.SetDefault("APP_ENV", "development")
	v.SetDefault("APP_PORT", "8080")
	v.SetDefault("LOG_LEVEL", "info")

	// Automatically read from environment variables
	v.AutomaticEnv()

	// Read from config file (prioritize internal/config/config.yaml, fallback to .)
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("internal/config")
	v.AddConfigPath(".")
	
	if err := v.ReadInConfig(); err != nil {
		log.Printf("ℹ️  No config.yaml found, using environment variables and defaults (%v)", err)
	}

	return v
}

// LoadAppConfig reads the AppConfig from a configured Viper instance.
func LoadAppConfig(v *viper.Viper) AppConfig {
	var cfg AppConfig
	if err := v.Unmarshal(&cfg); err != nil {
		log.Fatalf("failed to decode config into struct: %v", err)
	}
	return cfg
}
