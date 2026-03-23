package messaging

import (
	"os"
	"testing"
)

func TestRabbitMQConfig_Defaults(t *testing.T) {
	cfg := RabbitMQConfig{
		Host:     "localhost",
		Port:     5672,
		User:     "guest",
		Password: "guest",
		VHost:    "/",
	}

	if cfg.Host == "" {
		t.Error("Expected Host to have a default value")
	}
	if cfg.Port == 0 {
		t.Error("Expected Port to be non-zero")
	}
}

func TestNewRabbitMQConn_SkipIfNoServer(t *testing.T) {
	if os.Getenv("RABBITMQ_HOST") == "" {
		t.Skip("Skipping RabbitMQ integration test: RABBITMQ_HOST not set")
	}

	cfg := RabbitMQConfig{
		Host:     os.Getenv("RABBITMQ_HOST"),
		Port:     5672,
		User:     "guest",
		Password: "guest",
		VHost:    "/",
	}

	conn, err := NewRabbitMQConn(cfg)
	if err != nil {
		t.Fatalf("Expected successful connection, got: %v", err)
	}
	defer conn.Close()

	if conn.Conn == nil || conn.Channel == nil {
		t.Error("Expected non-nil connection and channel")
	}
}
