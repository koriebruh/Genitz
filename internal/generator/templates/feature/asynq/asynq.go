package worker

import (
	"log"

	"github.com/hibiken/asynq"
)

// AsynqConfig holds configuration for the Asynq task queue (backed by Redis).
type AsynqConfig struct {
	RedisAddr     string `env:"REDIS_ADDR"         envDefault:"localhost:6379"`
	RedisPassword string `env:"REDIS_PASSWORD"`
	RedisDB       int    `env:"REDIS_DB"           envDefault:"0"`
	Concurrency   int    `env:"ASYNQ_CONCURRENCY"  envDefault:"10"`
}

// NewAsynqClient creates an Asynq client for enqueuing tasks.
// Inject this into your service/use-case layer to schedule background jobs.
func NewAsynqClient(cfg AsynqConfig) *asynq.Client {
	redisOpt := asynq.RedisClientOpt{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	}
	client := asynq.NewClient(redisOpt)
	log.Println("✅ Asynq client initialized")
	return client
}

// NewAsynqServer creates an Asynq server for processing tasks.
// Call server.Run(mux) to start processing queued tasks.
func NewAsynqServer(cfg AsynqConfig) *asynq.Server {
	redisOpt := asynq.RedisClientOpt{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	}
	srv := asynq.NewServer(redisOpt, asynq.Config{
		Concurrency: cfg.Concurrency,
		Queues: map[string]int{
			"critical": 6,
			"default":  3,
			"low":      1,
		},
	})
	return srv
}
