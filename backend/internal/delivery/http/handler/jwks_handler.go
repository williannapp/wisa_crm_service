package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"wisa-crm-service/backend/internal/domain/service"
)

// JWKSHandler handles the public key discovery endpoint.
type JWKSHandler struct {
	provider service.JWKSProvider
}

// NewJWKSHandler creates a new JWKSHandler.
func NewJWKSHandler(provider service.JWKSProvider) *JWKSHandler {
	return &JWKSHandler{provider: provider}
}

// GetJWKS handles GET /.well-known/jwks.json.
// Returns JWKS in RFC 7517 format. No authentication required.
// Cache-Control: public, max-age=86400 (24 hours) per ADR-006.
func (h *JWKSHandler) GetJWKS(c *gin.Context) {
	keys, err := h.provider.GetKeys(c.Request.Context())
	if err != nil {
		log.Printf("JWKS provider error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Service temporarily unavailable"})
		return
	}
	c.Header("Cache-Control", "public, max-age=86400")
	c.JSON(http.StatusOK, service.JWKS{Keys: keys})
}
