package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"test-app/backend/internal/infrastructure/jwt"
)

const claimsKey = "jwt_claims"

// JWTAuth returns a middleware that validates JWT from cookie or Authorization header.
func JWTAuth(validator *jwt.Validator) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}
		if tokenString == "" {
			tokenString, _ = c.Cookie("wisa_access_token")
		}
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		claims, err := validator.Validate(c.Request.Context(), tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		c.Set(claimsKey, claims)
		c.Next()
	}
}

// GetClaims returns the JWT claims from context. Call only after JWTAuth middleware.
func GetClaims(c *gin.Context) map[string]interface{} {
	v, _ := c.Get(claimsKey)
	claims, _ := v.(map[string]interface{})
	return claims
}
