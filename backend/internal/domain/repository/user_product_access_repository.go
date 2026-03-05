package repository

import (
	"context"

	"github.com/google/uuid"

	"wisa-crm-service/backend/internal/domain/entity"
)

// UserProductAccessRepository defines the port for user-product access persistence.
type UserProductAccessRepository interface {
	FindByUserAndProduct(ctx context.Context, userID, productID uuid.UUID) (*entity.UserProductAccess, error)
}
