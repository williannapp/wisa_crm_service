package main

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"wisa-crm-service/backend/internal/delivery/http/handler"
	"wisa-crm-service/backend/internal/infrastructure/cache"
	"wisa-crm-service/backend/internal/infrastructure/crypto"
	"wisa-crm-service/backend/internal/infrastructure/http/middleware"
	"wisa-crm-service/backend/internal/infrastructure/persistence"
	"wisa-crm-service/backend/internal/usecase/auth"
)

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil {
			return n
		}
	}
	return def
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found, using system environment")
	}

	appEnv := getEnv("APP_ENV", "development")

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

	port := getEnv("PORT", "8080")

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(middleware.Recovery())
	router.GET("/health", handler.HealthHandler)

	if db != nil {
		jwtPrivateKeyPath := os.Getenv("JWT_PRIVATE_KEY_PATH")
		if jwtPrivateKeyPath == "" && appEnv == "production" {
			log.Fatal("JWT_PRIVATE_KEY_PATH is required in production")
		}
		if jwtPrivateKeyPath != "" {
			redisURL := os.Getenv("REDIS_URL")
			if redisURL == "" {
				if appEnv == "production" {
					log.Fatal("REDIS_URL is required in production")
				}
				redisURL = "redis://localhost:6379/0"
			}
			redisClient, err := cache.NewRedisClient(context.Background(), redisURL)
			if err != nil {
				log.Fatalf("Redis connection failed: %v", err)
			}
			defer redisClient.Close()
			tenantRepo := persistence.NewGormTenantRepository(db)
			productRepo := persistence.NewGormProductRepository(db)
			userRepo := persistence.NewGormUserRepository(db)
			subscriptionRepo := persistence.NewGormSubscriptionRepository(db)
			userProductAccRepo := persistence.NewGormUserProductAccessRepository(db)
			passwordSvc := crypto.NewBcryptPasswordService()

			jwtSvc, err := crypto.NewRSAJWTService(crypto.RSAJWTConfig{
				PrivateKeyPath: jwtPrivateKeyPath,
				Issuer:         getEnv("JWT_ISSUER", "wisa-crm-service"),
				ExpMinutes:      getEnvInt("JWT_EXPIRATION_MINUTES", 15),
				KeyID:          getEnv("JWT_KEY_ID", "key-2026-v1"),
			})
			if err != nil {
				log.Fatalf("JWT service initialization failed: %v", err)
			}

			jwksProvider, err := crypto.NewRSAJWKSProvider(crypto.RSAJWTConfig{
				PrivateKeyPath: jwtPrivateKeyPath,
				KeyID:          getEnv("JWT_KEY_ID", "key-2026-v1"),
			})
			if err != nil {
				log.Fatalf("JWKS provider initialization failed: %v", err)
			}
			jwksHandler := handler.NewJWKSHandler(jwksProvider)
			router.GET("/.well-known/jwks.json", jwksHandler.GetJWKS)

			authCodeStore := cache.NewRedisAuthCodeStore(redisClient)
			redirectBaseDomain := getEnv("JWT_AUD_BASE_DOMAIN", "app.wisa-crm.com")

			authenticateUser := auth.NewAuthenticateUserUseCase(
				tenantRepo,
				productRepo,
				userRepo,
				subscriptionRepo,
				userProductAccRepo,
				passwordSvc,
				authCodeStore,
				redirectBaseDomain,
			)
			refreshTokenRepo := persistence.NewGormRefreshTokenRepository(db)
			refreshTokenGen := crypto.NewRefreshTokenGenerator()
			exchangeCodeForToken := auth.NewExchangeCodeForTokenUseCase(
				authCodeStore,
				jwtSvc,
				refreshTokenRepo,
				refreshTokenGen,
			)
			refreshTokenUseCase := auth.NewRefreshTokenUseCase(
				tenantRepo,
				productRepo,
				subscriptionRepo,
				userProductAccRepo,
				refreshTokenRepo,
				jwtSvc,
				refreshTokenGen,
				redirectBaseDomain,
			)
			authHandler := handler.NewAuthHandler(authenticateUser, exchangeCodeForToken, refreshTokenUseCase)

			authGroup := router.Group("/api/v1/auth")
			authGroup.POST("/login", authHandler.Login)
			authGroup.POST("/token", authHandler.Token)
			authGroup.POST("/refresh", authHandler.Refresh)
		}
	}

	addr := ":" + port
	log.Printf("Server starting on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
