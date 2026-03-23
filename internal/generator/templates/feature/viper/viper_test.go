package config

import (
	"testing"
)

func TestNewViper_Defaults(t *testing.T) {
	v := NewViper()
	if v == nil {
		t.Fatal("Expected non-nil *viper.Viper")
	}

	if v.GetString("APP_NAME") == "" {
		t.Error("Expected APP_NAME default to be set")
	}
	if v.GetString("APP_PORT") == "" {
		t.Error("Expected APP_PORT default to be set")
	}
}

func TestLoadAppConfig_Defaults(t *testing.T) {
	v := NewViper()
	cfg := LoadAppConfig(v)

	if cfg.AppName == "" {
		t.Error("Expected AppName to have a default value")
	}
	if cfg.AppPort == "" {
		t.Error("Expected AppPort to have a default value")
	}
	if cfg.AppEnv == "" {
		t.Error("Expected AppEnv to have a default value")
	}
}
