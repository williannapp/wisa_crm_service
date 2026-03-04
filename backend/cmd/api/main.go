package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"wisa-crm-service/backend/internal/delivery/http/handler"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found, using system environment")
	}

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
