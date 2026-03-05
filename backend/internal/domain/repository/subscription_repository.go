package repository

import (
	"context"

	"github.com/google/uuid"

	"wisa-crm-service/backend/internal/domain/entity"
)

// SubscriptionRepository defines the port for subscription persistence.
type SubscriptionRepository interface {
	FindByTenantAndProduct(ctx context.Context, tenantID, productID uuid.UUID) (*entity.Subscription, error)
}
