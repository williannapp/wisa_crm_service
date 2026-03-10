package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"test-app/backend/internal/delivery/http/handler"
	"test-app/backend/internal/delivery/http/middleware"
	"test-app/backend/internal/infrastructure/jwt"
)

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found, using system environment")
	}

	port := getEnv("PORT", "8081")

	jwksFetcher := jwt.NewJWKSFetcher()
	jwtValidator := jwt.NewValidator(jwksFetcher)

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	authHandler := handler.NewAuthHandler()

	router.GET("/health", handler.HealthHandler)
	router.GET("/login", authHandler.LoginRedirect)
	router.GET("/callback", authHandler.Callback)
	router.GET("/:product/callback", authHandler.Callback)

	api := router.Group("/api")
	api.Use(middleware.JWTAuth(jwtValidator))
	{
		api.GET("/hello", handler.HelloHandler)
	}

	authAPI := router.Group("/api/auth")
	{
		authAPI.POST("/refresh", authHandler.Refresh)
	}

	addr := ":" + port
	log.Printf("Test app backend starting on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
