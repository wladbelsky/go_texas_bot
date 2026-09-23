package db

import (
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

// Init opens (creating if needed) the sqlite database at path and runs migrations.
func Init(path string) error {
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	database, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return err
	}

	if err = database.AutoMigrate(&ArkCollectionEntry{}, &ArkStats{}, &ArkPity{}, &GuildSettings{}, &GerStats{}, &UserGerCounter{}); err != nil {
		return err
	}

	DB = database
	return nil
}

// Close releases the underlying database connection.
func Close() error {
	if DB == nil {
		return nil
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
