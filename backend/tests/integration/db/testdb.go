package db

import (
	"testing"

	model "backend/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}
	return db
}

func SetupTestDB(t *testing.T) *gorm.DB {
	db := NewTestDB(t)
	if err := model.AutoMigrate(db); err != nil {
		t.Fatalf("failed to migrate models: %v", err)
	}
	return db
}

func CleanupTables(db *gorm.DB, tableNames ...string) {
	for _, tableName := range tableNames {
		db.Exec("DELETE FROM " + tableName)
	}
}
