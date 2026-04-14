package repository

import (
	"context"
	"testing"

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

func SetupTestDB(t *testing.T, models ...interface{}) *gorm.DB {
	db := NewTestDB(t)
	for _, m := range models {
		if err := db.AutoMigrate(m); err != nil {
			t.Fatalf("failed to migrate model: %v", err)
		}
	}
	return db
}

func CleanupTable(db *gorm.DB, tableName string) error {
	return db.Exec("DELETE FROM " + tableName).Error
}

type TransactionTestHelper struct {
	db *gorm.DB
}

func NewTransactionHelper(db *gorm.DB) *TransactionTestHelper {
	return &TransactionTestHelper{db: db}
}

func (h *TransactionTestHelper) WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}
