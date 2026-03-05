package errors

// Error codes (SCREAMING_SNAKE_CASE) for standardized API responses.
// Used for programmatic identification and logs.
const (
	CodeInvalidCredentials   = "INVALID_CREDENTIALS"
	CodeAccountLocked        = "ACCOUNT_LOCKED"
	CodeSubscriptionSuspended = "SUBSCRIPTION_SUSPENDED"
	CodeSubscriptionCanceled = "SUBSCRIPTION_CANCELED"
	CodeTenantNotFound       = "TENANT_NOT_FOUND"
	CodeUserNotFound         = "USER_NOT_FOUND"
	CodeRateLimitExceeded     = "RATE_LIMIT_EXCEEDED"
	CodeSubscriptionExpired   = "SUBSCRIPTION_EXPIRED"
	CodeInternalError         = "INTERNAL_ERROR"
)

// HTTP status constants for error responses.
const (
	HTTPBadRequest          = 400
	HTTPUnauthorized        = 401
	HTTPForbidden           = 403
	HTTPNotFound            = 404
	HTTPTooManyRequests     = 429
	HTTPInternalServerError = 500
)

// NewInvalidCredentials returns an AppError for invalid credentials (401).
// Per ADR-010: always use the same generic message for any auth failure.
func NewInvalidCredentials() *AppError {
	return NewAppError(
		CodeInvalidCredentials,
		"Credenciais inválidas.",
		"Verifique usuário e senha.",
		HTTPUnauthorized,
	)
}

// NewAccountLocked returns an AppError for locked account (403).
func NewAccountLocked() *AppError {
	return NewAppError(
		CodeAccountLocked,
		"Conta bloqueada.",
		"",
		HTTPForbidden,
	)
}

// NewSubscriptionSuspended returns an AppError for suspended subscription (403).
func NewSubscriptionSuspended() *AppError {
	return NewAppError(
		CodeSubscriptionSuspended,
		"Acesso suspenso por pendência financeira.",
		"",
		HTTPForbidden,
	)
}

// NewSubscriptionCanceled returns an AppError for canceled subscription (403).
func NewSubscriptionCanceled() *AppError {
	return NewAppError(
		CodeSubscriptionCanceled,
		"Assinatura cancelada.",
		"",
		HTTPForbidden,
	)
}

// NewSubscriptionExpired returns an AppError for expired subscription (403).
func NewSubscriptionExpired() *AppError {
	return NewAppError(
		CodeSubscriptionExpired,
		"Assinatura expirada.",
		"",
		HTTPForbidden,
	)
}

// NewTenantNotFound returns an AppError for tenant not found (404).
func NewTenantNotFound() *AppError {
	return NewAppError(
		CodeTenantNotFound,
		"Tenant não encontrado.",
		"",
		HTTPNotFound,
	)
}

// NewUserNotFound returns an AppError for user not found (404).
func NewUserNotFound() *AppError {
	return NewAppError(
		CodeUserNotFound,
		"Usuário não encontrado.",
		"",
		HTTPNotFound,
	)
}

// NewRateLimitExceeded returns an AppError for rate limit (429).
func NewRateLimitExceeded() *AppError {
	return NewAppError(
		CodeRateLimitExceeded,
		"Muitas tentativas. Tente novamente mais tarde.",
		"",
		HTTPTooManyRequests,
	)
}

// NewInternalError returns an AppError for unknown internal errors (500).
// Never expose internal error details to the client.
func NewInternalError() *AppError {
	return NewAppError(
		CodeInternalError,
		"Erro interno. Tente novamente.",
		"",
		HTTPInternalServerError,
	)
}
