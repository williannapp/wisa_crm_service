package dto

// LoginRequest is the request body for POST /api/v1/auth/login.
type LoginRequest struct {
	TenantSlug  string `json:"tenant_slug" binding:"required"`
	ProductSlug string `json:"product_slug" binding:"required"`
	UserEmail   string `json:"user_email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=1"`
	State       string `json:"state"` // CSRF token; client should send for validation on callback
}

// LoginResponse is the success response for login (when Accept: application/json).
// For normal flow, the handler responds with HTTP 302 redirect; no JSON body.
type LoginResponse struct {
	RedirectURL string `json:"redirect_url"`
}
