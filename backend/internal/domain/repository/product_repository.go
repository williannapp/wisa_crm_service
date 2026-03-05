package repository

import (
	"context"

	"wisa-crm-service/backend/internal/domain/entity"
)

// ProductRepository defines the port for product persistence.
type ProductRepository interface {
	FindBySlug(ctx context.Context, slug string) (*entity.Product, error)
}
