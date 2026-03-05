package repository

import (
	"context"

	"github.com/google/uuid"

	"wisa-crm-service/backend/internal/domain/entity"
)

// UserRepository defines the port for user persistence.
type UserRepository interface {
	FindByEmailAndTenantID(ctx context.Context, tenantID uuid.UUID, email string) (*entity.User, error)
}
