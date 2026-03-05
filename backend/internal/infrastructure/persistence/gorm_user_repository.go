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

type gormUserRepository struct {
	db *gorm.DB
}

// NewGormUserRepository creates a new GORM implementation of UserRepository.
func NewGormUserRepository(db *gorm.DB) repository.UserRepository {
	return &gormUserRepository{db: db}
}

func (r *gormUserRepository) FindByEmailAndTenantID(ctx context.Context, tenantID uuid.UUID, email string) (*entity.User, error) {
	var m model.UserModel
	result := r.db.WithContext(ctx).Where("tenant_id = ? AND email = ?", tenantID, email).First(&m)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, domain.ErrUserNotFound
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return userModelToEntity(&m), nil
}

func userModelToEntity(m *model.UserModel) *entity.User {
	if m == nil {
		return nil
	}
	return &entity.User{
		ID:           m.ID,
		TenantID:     m.TenantID,
		Email:        m.Email,
		PasswordHash: m.PasswordHash,
		Status:       m.Status,
	}
}
