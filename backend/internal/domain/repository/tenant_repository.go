package repository

import (
	"context"

	"wisa-crm-service/backend/internal/domain/entity"
)

// TenantRepository defines the port for tenant persistence.
type TenantRepository interface {
	FindBySlug(ctx context.Context, slug string) (*entity.Tenant, error)
}
