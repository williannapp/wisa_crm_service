package handler

import (
	"github.com/gin-gonic/gin"

	"wisa-crm-service/backend/internal/delivery/http/dto"
	deliveryerrors "wisa-crm-service/backend/internal/delivery/http/errors"
	apperrors "wisa-crm-service/backend/pkg/errors"
	"wisa-crm-service/backend/internal/usecase/auth"
)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	authenticateUser *auth.AuthenticateUserUseCase
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authenticateUser *auth.AuthenticateUserUseCase) *AuthHandler {
	return &AuthHandler{authenticateUser: authenticateUser}
}

// Login handles POST /api/v1/auth/login.
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(apperrors.HTTPBadRequest, apperrors.NewInvalidRequest("Dados inválidos. Verifique slug, product_slug, email e senha."))
		return
	}

	out, err := h.authenticateUser.Execute(c.Request.Context(), auth.LoginInput{
		Slug:        req.Slug,
		ProductSlug: req.ProductSlug,
		UserEmail:   req.UserEmail,
		Password:    req.Password,
	})
	if err != nil {
		deliveryerrors.RespondWithError(c, err)
		return
	}

	c.JSON(200, dto.LoginResponse{Token: out.Token})
}
