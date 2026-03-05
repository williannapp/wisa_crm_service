package persistence

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"wisa-crm-service/backend/internal/domain"
	"wisa-crm-service/backend/internal/domain/entity"
	"wisa-crm-service/backend/internal/domain/repository"
	"wisa-crm-service/backend/internal/infrastructure/persistence/model"
)

type gormTenantRepository struct {
	db *gorm.DB
}

// NewGormTenantRepository creates a new GORM implementation of TenantRepository.
func NewGormTenantRepository(db *gorm.DB) repository.TenantRepository {
	return &gormTenantRepository{db: db}
}

func (r *gormTenantRepository) FindBySlug(ctx context.Context, slug string) (*entity.Tenant, error) {
	var m model.TenantModel
	result := r.db.WithContext(ctx).Where("slug = ?", slug).First(&m)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, domain.ErrTenantNotFound
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return tenantModelToEntity(&m), nil
}

func tenantModelToEntity(m *model.TenantModel) *entity.Tenant {
	if m == nil {
		return nil
	}
	return &entity.Tenant{
		ID:     m.ID,
		Slug:   m.Slug,
		Name:   m.Name,
		Status: m.Status,
	}
}
