package persistence

import (
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewDatabase opens a connection to PostgreSQL and configures the connection pool.
// Never log the DSN or credentials — only generic success/error messages.
func NewDatabase(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("Database connection failed: %v", err)
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("Failed to get underlying sql.DB: %v", err)
		return nil, err
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	sqlDB.SetConnMaxIdleTime(2 * time.Minute)

	log.Print("Database connected successfully")
	return db, nil
}
