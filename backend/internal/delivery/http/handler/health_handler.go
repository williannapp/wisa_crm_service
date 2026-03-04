package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthHandler responds to GET /health with HTTP 200 and {"status":"ok"}.
// Used by load balancers, Kubernetes probes, and health checks.
func HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
