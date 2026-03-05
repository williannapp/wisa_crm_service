package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"wisa-crm-service/backend/internal/domain/entity"
	"wisa-crm-service/backend/internal/domain/repository"
	"wisa-crm-service/backend/internal/infrastructure/persistence/model"
)

type gormRefreshTokenRepository struct {
	db *gorm.DB
}

// NewGormRefreshTokenRepository creates a new GORM implementation of RefreshTokenRepository.
func NewGormRefreshTokenRepository(db *gorm.DB) repository.RefreshTokenRepository {
	return &gormRefreshTokenRepository{db: db}
}

func (r *gormRefreshTokenRepository) Create(ctx context.Context, token *entity.RefreshToken) error {
	if token.ID == uuid.Nil {
		token.ID = uuid.New()
	}
	m := refreshTokenEntityToModel(token)
	result := r.db.WithContext(ctx).Create(&m)
	return result.Error
}

func (r *gormRefreshTokenRepository) FindByHashAndTenantAndProduct(ctx context.Context, tokenHash string, tenantID, productID uuid.UUID) (*entity.RefreshToken, error) {
	var m model.RefreshTokenModel
	result := r.db.WithContext(ctx).
		Where("token_hash = ? AND tenant_id = ? AND product_id = ?", tokenHash, tenantID, productID).
		Where("revoked_at IS NULL").
		Where("expires_at > ?", time.Now().UTC()).
		First(&m)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return refreshTokenModelToEntity(&m), nil
}

func (r *gormRefreshTokenRepository) RevokeByID(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	result := r.db.WithContext(ctx).
		Model(&model.RefreshTokenModel{}).
		Where("id = ?", id).
		Update("revoked_at", now)
	return result.Error
}

func (r *gormRefreshTokenRepository) Rotate(ctx context.Context, oldTokenID uuid.UUID, newToken *entity.RefreshToken) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()
		if result := tx.Model(&model.RefreshTokenModel{}).
			Where("id = ?", oldTokenID).
			Update("revoked_at", now); result.Error != nil {
			return result.Error
		}
		if newToken.ID == uuid.Nil {
			newToken.ID = uuid.New()
		}
		m := refreshTokenEntityToModel(newToken)
		return tx.Create(&m).Error
	})
}

func refreshTokenEntityToModel(e *entity.RefreshToken) *model.RefreshTokenModel {
	return &model.RefreshTokenModel{
		ID:        e.ID,
		UserID:    e.UserID,
		TenantID:  e.TenantID,
		ProductID: e.ProductID,
		TokenHash: e.TokenHash,
		ExpiresAt: e.ExpiresAt,
		RevokedAt: e.RevokedAt,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}

func refreshTokenModelToEntity(m *model.RefreshTokenModel) *entity.RefreshToken {
	if m == nil {
		return nil
	}
	return &entity.RefreshToken{
		ID:        m.ID,
		UserID:    m.UserID,
		TenantID:  m.TenantID,
		ProductID: m.ProductID,
		TokenHash: m.TokenHash,
		ExpiresAt: m.ExpiresAt,
		RevokedAt: m.RevokedAt,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
