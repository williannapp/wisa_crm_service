package model

import (
	"time"

	"github.com/google/uuid"
)

// RefreshTokenModel is the GORM model for wisa_crm_db.refresh_tokens.
type RefreshTokenModel struct {
	ID        uuid.UUID  `gorm:"column:id;type:uuid;primaryKey"`
	UserID    uuid.UUID  `gorm:"column:user_id;type:uuid;not null"`
	TenantID  uuid.UUID  `gorm:"column:tenant_id;type:uuid;not null"`
	ProductID uuid.UUID  `gorm:"column:product_id;type:uuid;not null"`
	TokenHash string     `gorm:"column:token_hash;type:char(64);not null;uniqueIndex"`
	ExpiresAt time.Time  `gorm:"column:expires_at;type:timestamptz;not null"`
	RevokedAt *time.Time `gorm:"column:revoked_at;type:timestamptz"`
	CreatedAt time.Time  `gorm:"column:created_at;type:timestamptz;not null"`
	UpdatedAt time.Time  `gorm:"column:updated_at;type:timestamptz;not null"`
}

// TableName returns the fully qualified table name.
func (RefreshTokenModel) TableName() string {
	return "wisa_crm_db.refresh_tokens"
}
