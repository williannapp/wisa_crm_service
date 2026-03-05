package persistence

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"wisa-crm-service/backend/internal/domain"
	"wisa-crm-service/backend/internal/domain/entity"
	"wisa-crm-service/backend/internal/domain/repository"
	"wisa-crm-service/backend/internal/infrastructure/persistence/model"
)

type gormSubscriptionRepository struct {
	db *gorm.DB
}

// NewGormSubscriptionRepository creates a new GORM implementation of SubscriptionRepository.
func NewGormSubscriptionRepository(db *gorm.DB) repository.SubscriptionRepository {
	return &gormSubscriptionRepository{db: db}
}

func (r *gormSubscriptionRepository) FindByTenantAndProduct(ctx context.Context, tenantID, productID uuid.UUID) (*entity.Subscription, error) {
	var m model.SubscriptionModel
	result := r.db.WithContext(ctx).
		Where("tenant_id = ? AND product_id = ?", tenantID, productID).
		First(&m)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, domain.ErrSubscriptionExpired
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return subscriptionModelToEntity(&m), nil
}

func subscriptionModelToEntity(m *model.SubscriptionModel) *entity.Subscription {
	if m == nil {
		return nil
	}
	return &entity.Subscription{
		ID:        m.ID,
		TenantID:  m.TenantID,
		ProductID: m.ProductID,
		Status:    m.Status,
	}
}
