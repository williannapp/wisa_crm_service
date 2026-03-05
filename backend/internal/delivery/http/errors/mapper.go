package errors

import (
	"errors"
	"log"

	apperrors "wisa-crm-service/backend/pkg/errors"

	"wisa-crm-service/backend/internal/domain"
)

// MapToAppError converts domain errors to standardized AppError for HTTP responses.
// Unknown errors (GORM, bcrypt, etc.) are mapped to INTERNAL_ERROR (500).
// Per ADR-010: auth failures always return INVALID_CREDENTIALS with generic message.
// Never expose err.Error() or internal details to the client.
func MapToAppError(err error) *apperrors.AppError {
	if err == nil {
		return nil
	}

	// Specific errors first (order matters for wrapped errors)
	switch {
	case errors.Is(err, domain.ErrAccountLocked):
		return apperrors.NewAccountLocked()
	case errors.Is(err, domain.ErrSubscriptionSuspended):
		return apperrors.NewSubscriptionSuspended()
	case errors.Is(err, domain.ErrSubscriptionCanceled):
		return apperrors.NewSubscriptionCanceled()
	case errors.Is(err, domain.ErrSubscriptionExpired):
		return apperrors.NewSubscriptionExpired()
	case errors.Is(err, domain.ErrRateLimitExceeded):
		return apperrors.NewRateLimitExceeded()
	case errors.Is(err, domain.ErrInvalidCredentials):
		// Auth context: use case converts UserNotFound/TenantNotFound to this per ADR-010
		return apperrors.NewInvalidCredentials()
	case errors.Is(err, domain.ErrTenantNotFound):
		return apperrors.NewTenantNotFound()
	case errors.Is(err, domain.ErrUserNotFound):
		return apperrors.NewUserNotFound()
	default:
		// Unknown errors: log internally, never expose to client
		log.Printf("[MapToAppError] unknown error (internal): %v", err)
		return apperrors.NewInternalError()
	}
}
