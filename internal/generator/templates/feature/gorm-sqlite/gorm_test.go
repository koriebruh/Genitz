package database

import (
	"testing"
	"os"
)

func TestNewDatabase(t *testing.T) {
	// Skip integration test if no real database available
	if os.Getenv("DB_HOST") == "" {
		t.Skip("Skipping database integration test: DB_HOST not set")
	}

	cfg := LoadDatabaseConfig()
	db := NewDatabase(cfg)
	if db == nil {
		t.Fatal("Expected non-nil *gorm.DB")
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get sql.DB: %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		t.Errorf("Database ping failed: %v", err)
	}
}

func TestLoadDatabaseConfig(t *testing.T) {
	cfg := LoadDatabaseConfig()
	if cfg.Name == "" {
		t.Error("Expected default DB_NAME to be set for SQLite")
	}
}
