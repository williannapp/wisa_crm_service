package dto

// RefreshRequest is the request body for POST /api/v1/auth/refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
	ProductSlug  string `json:"product_slug" binding:"required"`
	TenantSlug   string `json:"tenant_slug" binding:"required"`
}
