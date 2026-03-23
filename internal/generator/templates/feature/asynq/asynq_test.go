package worker

import (
	"testing"
)

func TestAsynqConfig_Defaults(t *testing.T) {
	cfg := AsynqConfig{
		RedisAddr:   "localhost:6379",
		Concurrency: 10,
	}

	if cfg.RedisAddr == "" {
		t.Error("Expected RedisAddr to be set")
	}
	if cfg.Concurrency <= 0 {
		t.Error("Expected Concurrency to be greater than 0")
	}
}

func TestNewAsynqClient_CreatesWithoutPanic(t *testing.T) {
	cfg := AsynqConfig{
		RedisAddr:   "localhost:6379",
		Concurrency: 5,
	}
	// NewAsynqClient doesn't dial Redis immediately — safe to call without server
	client := NewAsynqClient(cfg)
	if client == nil {
		t.Fatal("Expected non-nil *asynq.Client")
	}
	client.Close() //nolint:errcheck
}

func TestNewAsynqServer_CreatesWithoutPanic(t *testing.T) {
	cfg := AsynqConfig{
		RedisAddr:   "localhost:6379",
		Concurrency: 3,
	}
	srv := NewAsynqServer(cfg)
	if srv == nil {
		t.Fatal("Expected non-nil *asynq.Server")
	}
}
