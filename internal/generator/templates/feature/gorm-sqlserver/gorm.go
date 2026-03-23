package database

import (
	"fmt"
	"log"

	"github.com/caarlos0/env/v11"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DatabaseConfig struct {
	Host     string `env:"DB_HOST"     envDefault:"localhost"`
	Port     int    `env:"DB_PORT"     envDefault:"1433"`
	Name     string `env:"DB_NAME"     envDefault:"mydb"`
	User     string `env:"DB_USER"     envDefault:"sa"`
	Password string `env:"DB_PASSWORD" envDefault:""`
}

func LoadDatabaseConfig() DatabaseConfig {
	cfg := DatabaseConfig{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to load database config: %v", err)
	}
	return cfg
}

// NewDatabase creates a new GORM DB connection with SQL Server.
func NewDatabase(cfg DatabaseConfig) *gorm.DB {
	dsn := fmt.Sprintf(
		"sqlserver://%s:%s@%s:%d?database=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name,
	)

	db, err := gorm.Open(sqlserver.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get sql.DB: %v", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	return db
}
