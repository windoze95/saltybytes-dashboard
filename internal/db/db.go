package db

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(databaseURL string) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		db, err = gorm.Open(postgres.Open(databaseURL), &gorm.Config{
			Logger:      logger.Default.LogMode(logger.Warn),
			PrepareStmt: true,
		})
		if err == nil {
			break
		}

		select {
		case <-timeout:
			return nil, fmt.Errorf("failed to connect to database after 30s: %w", err)
		case <-ticker.C:
			continue
		}
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetMaxIdleConns(2)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	return db, nil
}
