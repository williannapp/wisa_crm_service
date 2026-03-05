package domain

import "errors"

// Domain errors for communication between layers.
// These are mapped to AppError in the delivery layer via ErrorMapper.
// Per ADR-010: auth failures (user not found, wrong password, tenant not found)
// must be converted to ErrInvalidCredentials before reaching the handler.
var (
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrTenantNotFound        = errors.New("tenant not found")
	ErrUserNotFound          = errors.New("user not found")
	ErrUserBlocked           = errors.New("user blocked")
	ErrProductNotFound       = errors.New("product not found")
	ErrProductUnavailable   = errors.New("product unavailable")
	ErrSubscriptionExpired   = errors.New("subscription expired")
	ErrSubscriptionSuspended = errors.New("subscription suspended")
	ErrSubscriptionCanceled  = errors.New("subscription canceled")
	ErrAccountLocked         = errors.New("account locked")
	ErrRateLimitExceeded     = errors.New("rate limit exceeded")
	ErrCodeInvalidOrExpired  = errors.New("authorization code invalid or expired")
	ErrAuthCodeStorageUnavailable = errors.New("authorization code storage unavailable")
)
