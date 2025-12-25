package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDatabase инициализирует подключение к базе данных
func InitDatabase(dbPath string) error {
	// Создаем директорию для БД если её нет
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	// Открываем соединение с БД
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Включаем WAL mode для лучшей производительности
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	if _, err := sqlDB.Exec("PRAGMA journal_mode=WAL"); err != nil {
		log.Printf("Warning: failed to enable WAL mode: %v", err)
	}

	// Автоматическая миграция схемы
	if err := db.AutoMigrate(&User{}); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	DB = db
	log.Printf("Database initialized successfully at %s", dbPath)
	return nil
}

// CloseDatabase закрывает соединение с базой данных
func CloseDatabase() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	log.Println("Database connection closed")
	return nil
}

// GetDB возвращает экземпляр базы данных
func GetDB() *gorm.DB {
	return DB
}
