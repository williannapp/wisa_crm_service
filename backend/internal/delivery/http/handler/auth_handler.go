package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"wisa-crm-service/backend/internal/delivery/http/dto"
	deliveryerrors "wisa-crm-service/backend/internal/delivery/http/errors"
	apperrors "wisa-crm-service/backend/pkg/errors"
	"wisa-crm-service/backend/internal/usecase/auth"
)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	authenticateUser     *auth.AuthenticateUserUseCase
	exchangeCodeForToken *auth.ExchangeCodeForTokenUseCase
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authenticateUser *auth.AuthenticateUserUseCase, exchangeCodeForToken *auth.ExchangeCodeForTokenUseCase) *AuthHandler {
	return &AuthHandler{
		authenticateUser:     authenticateUser,
		exchangeCodeForToken: exchangeCodeForToken,
	}
}

// Login handles POST /api/v1/auth/login.
// On success, responds with HTTP 302 redirect to client callback URL with code and state.
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
		State:       req.State,
	})
	if err != nil {
		deliveryerrors.RespondWithError(c, err)
		return
	}

	c.Redirect(http.StatusFound, out.RedirectURL)
}

// Token handles POST /api/v1/auth/token (exchange authorization code for JWT).
func (h *AuthHandler) Token(c *gin.Context) {
	var req dto.TokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(apperrors.HTTPBadRequest, apperrors.NewInvalidRequest("Dados inválidos. O campo code é obrigatório."))
		return
	}

	out, err := h.exchangeCodeForToken.Execute(c.Request.Context(), auth.ExchangeCodeInput{
		Code: req.Code,
	})
	if err != nil {
		deliveryerrors.RespondWithError(c, err)
		return
	}

	c.JSON(200, dto.TokenResponse{
		AccessToken: out.AccessToken,
		ExpiresIn:   out.ExpiresIn,
	})
}
