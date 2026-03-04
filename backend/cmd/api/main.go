package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"wisa-crm-service/backend/internal/delivery/http/handler"
	"wisa-crm-service/backend/internal/infrastructure/persistence"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found, using system environment")
	}

	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "development"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	var db *gorm.DB
	if databaseURL != "" {
		conn, err := persistence.NewDatabase(databaseURL)
		if err != nil {
			log.Fatalf("Database connection failed: %v", err)
		}
		db = conn
	} else if appEnv == "production" {
		log.Fatal("DATABASE_URL is required in production")
	} else {
		log.Print("DATABASE_URL is empty; running without database (development mode)")
	}
	_ = db // reserved for future use by handlers/repositories

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router := gin.Default()
	router.GET("/health", handler.HealthHandler)

	addr := ":" + port
	log.Printf("Server starting on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
