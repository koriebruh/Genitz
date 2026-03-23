package database

import (
	"log"

	"github.com/caarlos0/env/v11"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DatabaseConfig struct {
	Name string `env:"DB_NAME" envDefault:"mydb.sqlite"`
}

func LoadDatabaseConfig() DatabaseConfig {
	cfg := DatabaseConfig{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to load database config: %v", err)
	}
	return cfg
}

// NewDatabase creates a new GORM DB connection using SQLite.
func NewDatabase(cfg DatabaseConfig) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(cfg.Name), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get sql.DB: %v", err)
	}

	// Connection pool settings for SQLite
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetMaxOpenConns(1)

	return db
}
