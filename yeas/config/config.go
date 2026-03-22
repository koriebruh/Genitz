package config

import "os"

type Config struct {
	AppName string

	Fiber FiberConfig

	Database DatabaseConfig

	Redis RedisConfig

	Logger LoggerConfig
}

func LoadConfig() Config {
	return Config{
		AppName: os.Getenv("APP_NAME"),

		Fiber: LoadFiberConfig(),

		Database: LoadDatabaseConfig(),

		Redis: LoadRedisConfig(),

		Logger: LoadLoggerConfig(),
	}
}
