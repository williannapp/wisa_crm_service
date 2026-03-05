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

type gormUserProductAccessRepository struct {
	db *gorm.DB
}

// NewGormUserProductAccessRepository creates a new GORM implementation of UserProductAccessRepository.
func NewGormUserProductAccessRepository(db *gorm.DB) repository.UserProductAccessRepository {
	return &gormUserProductAccessRepository{db: db}
}

func (r *gormUserProductAccessRepository) FindByUserAndProduct(ctx context.Context, userID, productID uuid.UUID) (*entity.UserProductAccess, error) {
	var m model.UserProductAccessModel
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND product_id = ?", userID, productID).
		First(&m)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, domain.ErrUserNotFound
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return userProductAccessModelToEntity(&m), nil
}

func userProductAccessModelToEntity(m *model.UserProductAccessModel) *entity.UserProductAccess {
	if m == nil {
		return nil
	}
	return &entity.UserProductAccess{
		UserID:        m.UserID,
		ProductID:     m.ProductID,
		TenantID:      m.TenantID,
		AccessProfile: m.AccessProfile,
	}
}
