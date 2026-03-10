package handler

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	oauthStateCookie   = "oauth_state"
	oauthStateMaxAge   = 300
	accessTokenCookie  = "wisa_access_token"
	refreshTokenCookie = "wisa_refresh_token"
)

// AuthHandler handles OAuth redirect and callback.
type AuthHandler struct {
	authServerURL string
	tenantSlug    string
	productSlug   string
	appURL        string
	frontendURL   string // Optional: redirect here after callback when separate from backend
	httpClient    *http.Client
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler() *AuthHandler {
	authURL := os.Getenv("AUTH_SERVER_URL")
	if authURL == "" {
		authURL = "https://auth.wisa.labs.com.br"
	}
	if strings.HasSuffix(authURL, "/") {
		authURL = strings.TrimSuffix(authURL, "/")
	}
	return &AuthHandler{
		authServerURL: authURL,
		tenantSlug:    getEnv("TENANT_SLUG", "cliente1"),
		productSlug:   getEnv("PRODUCT_SLUG", "gestao-pocket"),
		appURL:        getEnv("APP_URL", "http://localhost:8081"),
		frontendURL:   getEnv("FRONTEND_URL", ""),
		httpClient:    &http.Client{Timeout: 15 * time.Second},
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// LoginRedirect initiates the auth flow: generates state, sets cookie, redirects to auth server.
func (h *AuthHandler) LoginRedirect(c *gin.Context) {
	state, err := generateState()
	if err != nil {
		c.Redirect(302, "/login?error=state_error")
		return
	}

	c.SetCookie(oauthStateCookie, state, oauthStateMaxAge, "/callback", "", false, true)

	redirectURL := fmt.Sprintf("%s/login?tenant_slug=%s&product_slug=%s&state=%s",
		h.authServerURL,
		url.QueryEscape(h.tenantSlug),
		url.QueryEscape(h.productSlug),
		url.QueryEscape(state),
	)
	c.Redirect(302, redirectURL)
}

// Callback handles the redirect from auth server: validates state, exchanges code for token, sets cookies.
func (h *AuthHandler) Callback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		c.Redirect(302, "/login?error=invalid_callback")
		return
	}

	savedState, err := c.Cookie(oauthStateCookie)
	if err != nil || savedState != state {
		c.SetCookie(oauthStateCookie, "", -1, "/callback", "", false, true)
		c.Redirect(302, "/login?error=invalid_callback")
		return
	}
	c.SetCookie(oauthStateCookie, "", -1, "/callback", "", false, true)

	tokenURL := h.authServerURL + "/api/v1/auth/token"
	body, _ := json.Marshal(map[string]string{"code": code})
	req, err := http.NewRequestWithContext(c.Request.Context(), "POST", tokenURL, bytes.NewReader(body))
	if err != nil {
		c.Redirect(302, "/login?error=auth_failed")
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		c.Redirect(302, "/login?error=auth_failed")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		c.Redirect(302, "/login?error=auth_failed")
		return
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		c.Redirect(302, "/login?error=auth_failed")
		return
	}

	var out struct {
		AccessToken      string `json:"access_token"`
		ExpiresIn        int    `json:"expires_in"`
		RefreshToken     string `json:"refresh_token"`
		RefreshExpiresIn int    `json:"refresh_expires_in"`
	}
	if err := json.Unmarshal(data, &out); err != nil || out.AccessToken == "" {
		c.Redirect(302, "/login?error=auth_failed")
		return
	}

	secure := strings.HasPrefix(h.appURL, "https://")
	maxAgeAccess := out.ExpiresIn
	if maxAgeAccess <= 0 {
		maxAgeAccess = 900
	}
	maxAgeRefresh := out.RefreshExpiresIn
	if maxAgeRefresh <= 0 {
		maxAgeRefresh = 604800
	}

	c.SetCookie(accessTokenCookie, out.AccessToken, maxAgeAccess, "/", "", secure, true)
	c.SetCookie(refreshTokenCookie, out.RefreshToken, maxAgeRefresh, "/", "", secure, true)

	redirectTo := "/"
	if h.frontendURL != "" {
		redirectTo = h.frontendURL
	}
	c.Redirect(302, redirectTo)
}

// Refresh proxies the refresh request to the auth server and updates cookies.
func (h *AuthHandler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie(refreshTokenCookie)
	if err != nil || refreshToken == "" {
		c.SetCookie(accessTokenCookie, "", -1, "/", "", false, true)
		c.SetCookie(refreshTokenCookie, "", -1, "/", "", false, true)
		c.AbortWithStatus(401)
		return
	}

	tokenURL := h.authServerURL + "/api/v1/auth/refresh"
	body, _ := json.Marshal(map[string]string{
		"refresh_token":  refreshToken,
		"tenant_slug":    h.tenantSlug,
		"product_slug":   h.productSlug,
	})
	req, err := http.NewRequestWithContext(c.Request.Context(), "POST", tokenURL, bytes.NewReader(body))
	if err != nil {
		c.AbortWithStatus(401)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		c.AbortWithStatus(503)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		if ra := resp.Header.Get("Retry-After"); ra != "" {
			c.Header("Retry-After", ra)
		}
		c.AbortWithStatus(429)
		return
	}

	if resp.StatusCode != 200 {
		c.SetCookie(accessTokenCookie, "", -1, "/", "", false, true)
		c.SetCookie(refreshTokenCookie, "", -1, "/", "", false, true)
		c.AbortWithStatus(401)
		return
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		c.AbortWithStatus(500)
		return
	}

	var out struct {
		AccessToken      string `json:"access_token"`
		ExpiresIn        int    `json:"expires_in"`
		RefreshToken     string `json:"refresh_token"`
		RefreshExpiresIn int    `json:"refresh_expires_in"`
	}
	if err := json.Unmarshal(data, &out); err != nil || out.AccessToken == "" {
		c.AbortWithStatus(500)
		return
	}

	secure := strings.HasPrefix(h.appURL, "https://")
	maxAgeAccess := out.ExpiresIn
	if maxAgeAccess <= 0 {
		maxAgeAccess = 900
	}
	maxAgeRefresh := out.RefreshExpiresIn
	if maxAgeRefresh <= 0 {
		maxAgeRefresh = 604800
	}

	c.SetCookie(accessTokenCookie, out.AccessToken, maxAgeAccess, "/", "", secure, true)
	c.SetCookie(refreshTokenCookie, out.RefreshToken, maxAgeRefresh, "/", "", secure, true)
	c.Status(200)
}

func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
