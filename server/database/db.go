package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/anveesa/proxera/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init(dsn string) error {
	// Ensure directory exists
	dir := filepath.Dir(dsn)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	// Append pragmas for WAL mode and foreign keys
	dsnWithPragmas := dsn + "?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)"

	db, err := gorm.Open(sqlite.Open(dsnWithPragmas), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxOpenConns(1) // SQLite only supports one writer at a time

	// AutoMigrate
	if err := db.AutoMigrate(
		&models.Server{},
		&models.Route{},
		&models.Alert{},
	); err != nil {
		return fmt.Errorf("automigrate failed: %w", err)
	}

	log.Printf("Database initialized at %s", dsn)
	DB = db
	return nil
}
