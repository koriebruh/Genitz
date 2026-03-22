package config

import (
	"fmt"
	"log"

	"github.com/caarlos0/env/v11"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DatabaseConfig struct {
	Host     string `env:"DB_HOST"     envDefault:"localhost"`
	Port     int    `env:"DB_PORT"     envDefault:"5432"`
	Name     string `env:"DB_NAME"     envDefault:"mydb"`
	User     string `env:"DB_USER"     envDefault:"postgres"`
	Password string `env:"DB_PASSWORD" envDefault:""`
	SSLMode  string `env:"DB_SSLMODE"  envDefault:"disable"`
}

func LoadDatabaseConfig() DatabaseConfig {
	cfg := DatabaseConfig{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to load database config: %v", err)
	}
	return cfg
}

// NewDatabase creates a new GORM DB connection with the provided config.
func NewDatabase(cfg DatabaseConfig) *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Jakarta",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get sql.DB: %v", err)
	}

	// Connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	return db
}
