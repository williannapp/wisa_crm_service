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

type gormProductRepository struct {
	db *gorm.DB
}

// NewGormProductRepository creates a new GORM implementation of ProductRepository.
func NewGormProductRepository(db *gorm.DB) repository.ProductRepository {
	return &gormProductRepository{db: db}
}

func (r *gormProductRepository) FindBySlug(ctx context.Context, slug string) (*entity.Product, error) {
	var m model.ProductModel
	result := r.db.WithContext(ctx).Where("slug = ?", slug).First(&m)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, domain.ErrProductNotFound
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return productModelToEntity(&m), nil
}

func productModelToEntity(m *model.ProductModel) *entity.Product {
	if m == nil {
		return nil
	}
	return &entity.Product{
		ID:     m.ID,
		Slug:   m.Slug,
		Name:   m.Name,
		Status: m.Status,
	}
}
