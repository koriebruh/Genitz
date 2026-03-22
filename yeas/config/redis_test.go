package config

import (
	"context"
	"testing"
	"time"
)

func TestRedisConnection(t *testing.T) {
	cfg := LoadRedisConfig()

	// Create client
	client, err := NewRedisClient(cfg)
	if err != nil {
		t.Fatalf("Failed to create redis client: %v", err)
	}
	defer client.Close()

	// Test Ping
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx); err != nil {
		t.Errorf("Redis Ping failed: %v", err)
	}

	// Test Set/Get
	key := "test_key"
	value := "hello genitz"

	if err := client.Set(ctx, key, value, 10*time.Second); err != nil {
		t.Errorf("Failed to set key: %v", err)
	}

	val, err := client.Get(ctx, key)
	if err != nil {
		t.Errorf("Failed to get key: %v", err)
	}

	if val != value {
		t.Errorf("Expected %s, got %s", value, val)
	}

	// Cleanup
	_ = client.Delete(ctx, key)
}
