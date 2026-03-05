package dto

// TokenRequest is the request body for POST /api/v1/auth/token.
type TokenRequest struct {
	Code string `json:"code" binding:"required"`
}

// TokenResponse is the success response for token exchange.
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}
