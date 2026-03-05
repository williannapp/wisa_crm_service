package middleware

import (
	"log"

	"github.com/gin-gonic/gin"

	apperrors "wisa-crm-service/backend/pkg/errors"
)

// Recovery returns a Gin middleware that captures panics, logs them internally,
// and responds with a generic INTERNAL_ERROR (500) without exposing stack traces.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log internally for debugging — never expose to client
				log.Printf("[Recovery] panic recovered: %v", err)
				if !c.Writer.Written() {
					appErr := apperrors.NewInternalError()
					c.JSON(appErr.HTTPStatus, appErr)
				}
				c.Abort()
			}
		}()
		c.Next()
	}
}
