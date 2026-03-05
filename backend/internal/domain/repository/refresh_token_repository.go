package repository

import (
	"context"

	"github.com/google/uuid"

	"wisa-crm-service/backend/internal/domain/entity"
)

// RefreshTokenRepository defines the port for refresh token persistence.
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *entity.RefreshToken) error
	FindByHashAndTenantAndProduct(ctx context.Context, tokenHash string, tenantID, productID uuid.UUID) (*entity.RefreshToken, error)
	RevokeByID(ctx context.Context, id uuid.UUID) error
	Rotate(ctx context.Context, oldTokenID uuid.UUID, newToken *entity.RefreshToken) error
}
