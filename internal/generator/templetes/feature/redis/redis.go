package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Enabled      bool
	Host         string
	Port         string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	MaxRetries   int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func LoadRedisConfig() RedisConfig {
	enabled, _ := strconv.ParseBool(getEnv("REDIS_ENABLED", "true"))
	db, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))
	poolSize, _ := strconv.Atoi(getEnv("REDIS_POOL_SIZE", "10"))
	minIdle, _ := strconv.Atoi(getEnv("REDIS_MIN_IDLE", "2"))
	maxRetries, _ := strconv.Atoi(getEnv("REDIS_MAX_RETRIES", "3"))

	return RedisConfig{
		Enabled:      enabled,
		Host:         getEnv("REDIS_HOST", "localhost"),
		Port:         getEnv("REDIS_PORT", "6379"),
		Password:     getEnv("REDIS_PASSWORD", ""),
		DB:           db,
		PoolSize:     poolSize,
		MinIdleConns: minIdle,
		MaxRetries:   maxRetries,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}
}

// Helper to read env with fallback
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// RedisClient wraps redis.Client.
type RedisClient struct {
	Client *redis.Client
}

// NewRedisClient creates a new Redis client with connection pool.
func NewRedisClient(cfg RedisConfig) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password:     cfg.Password, // no password set
		DB:           cfg.DB,       // use default DB
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), cfg.DialTimeout)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("could not connect to redis: %v", err)
		return nil, err
	}

	return &RedisClient{Client: rdb}, nil
}

// Close closes the Redis connection.
func (r *RedisClient) Close() error {
	if err := r.Client.Close(); err != nil {
		log.Println("❌ Failed to close Redis connection:", err)
		return err
	}
	log.Println("Redis connection closed")
	return nil
}

// Ping checks if Redis is alive.
func (r *RedisClient) Ping(ctx context.Context) error {
	return r.Client.Ping(ctx).Err()
}

// GetStats returns Redis connection pool statistics.
func (r *RedisClient) GetStats() *redis.PoolStats {
	return r.Client.PoolStats()
}

// ─── Common Redis Operations ──────────────────────────────────────────────────

// Set sets a key-value pair with expiration.
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.Client.Set(ctx, key, value, expiration).Err()
}

// Get retrieves a value by key.
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return r.Client.Get(ctx, key).Result()
}

// Delete removes one or more keys.
func (r *RedisClient) Delete(ctx context.Context, keys ...string) error {
	return r.Client.Del(ctx, keys...).Err()
}

// Exists reports whether at least one of the given keys exists.
func (r *RedisClient) Exists(ctx context.Context, keys ...string) (bool, error) {
	result, err := r.Client.Exists(ctx, keys...).Result()
	return result > 0, err
}

// SetNX sets a key only if it does not exist (useful for distributed locks).
func (r *RedisClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return r.Client.SetNX(ctx, key, value, expiration).Result()
}

// Increment atomically increments a counter key by 1.
func (r *RedisClient) Increment(ctx context.Context, key string) (int64, error) {
	return r.Client.Incr(ctx, key).Result()
}

// Expire sets a TTL on an existing key.
func (r *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.Client.Expire(ctx, key, expiration).Err()
}
