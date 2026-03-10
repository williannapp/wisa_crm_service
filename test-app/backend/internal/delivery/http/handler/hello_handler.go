package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HelloHandler returns a simple message for authenticated users.
func HelloHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Hello World"})
}
