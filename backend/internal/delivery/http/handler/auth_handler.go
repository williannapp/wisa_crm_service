package handler

import (
	"net/http"
	"strings"

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
	refreshToken         *auth.RefreshTokenUseCase
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(
	authenticateUser *auth.AuthenticateUserUseCase,
	exchangeCodeForToken *auth.ExchangeCodeForTokenUseCase,
	refreshToken *auth.RefreshTokenUseCase,
) *AuthHandler {
	return &AuthHandler{
		authenticateUser:     authenticateUser,
		exchangeCodeForToken: exchangeCodeForToken,
		refreshToken:         refreshToken,
	}
}

// Login handles POST /api/v1/auth/login.
// On success: if Accept contains application/json (SPA/XHR), responds with 200 + JSON { redirect_url };
// otherwise responds with HTTP 302 redirect for traditional form submissions.
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(apperrors.HTTPBadRequest, apperrors.NewInvalidRequest("Dados inválidos. Verifique tenant_slug, product_slug, email e senha."))
		return
	}

	out, err := h.authenticateUser.Execute(c.Request.Context(), auth.LoginInput{
		TenantSlug:  req.TenantSlug,
		ProductSlug: req.ProductSlug,
		UserEmail:   req.UserEmail,
		Password:    req.Password,
		State:       req.State,
	})
	if err != nil {
		deliveryerrors.RespondWithError(c, err)
		return
	}

	if strings.Contains(c.GetHeader("Accept"), "application/json") {
		c.JSON(http.StatusOK, dto.LoginResponse{RedirectURL: out.RedirectURL})
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
		AccessToken:      out.AccessToken,
		ExpiresIn:        out.ExpiresIn,
		RefreshToken:     out.RefreshToken,
		RefreshExpiresIn: out.RefreshExpiresIn,
	})
}

// Refresh handles POST /api/v1/auth/refresh.
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(apperrors.HTTPBadRequest, apperrors.NewInvalidRequest("Dados inválidos. Verifique refresh_token, product_slug e tenant_slug."))
		return
	}

	out, err := h.refreshToken.Execute(c.Request.Context(), auth.RefreshTokenInput{
		RefreshToken: req.RefreshToken,
		ProductSlug:  req.ProductSlug,
		TenantSlug:   req.TenantSlug,
	})
	if err != nil {
		deliveryerrors.RespondWithError(c, err)
		return
	}

	c.JSON(200, dto.TokenResponse{
		AccessToken:      out.AccessToken,
		ExpiresIn:        out.ExpiresIn,
		RefreshToken:     out.RefreshToken,
		RefreshExpiresIn: out.RefreshExpiresIn,
	})
}
