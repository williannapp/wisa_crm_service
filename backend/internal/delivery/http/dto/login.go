package dto

// LoginRequest is the request body for POST /api/v1/auth/login.
type LoginRequest struct {
	Slug        string `json:"slug" binding:"required"`
	ProductSlug string `json:"product_slug" binding:"required"`
	UserEmail   string `json:"user_email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=1"`
}

// LoginResponse is the success response for login.
type LoginResponse struct {
	Token string `json:"token"`
}
