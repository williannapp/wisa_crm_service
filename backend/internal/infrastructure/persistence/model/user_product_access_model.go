package model

import (
	"time"

	"github.com/google/uuid"
)

// UserProductAccessModel is the GORM model for wisa_crm_db.user_product_access.
type UserProductAccessModel struct {
	ID            uuid.UUID `gorm:"column:id;type:uuid;primaryKey"`
	UserID        uuid.UUID `gorm:"column:user_id;type:uuid;index"`
	ProductID     uuid.UUID `gorm:"column:product_id;type:uuid;index"`
	TenantID      uuid.UUID `gorm:"column:tenant_id;type:uuid;index"`
	AccessProfile string    `gorm:"column:access_profile"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

// TableName returns the fully qualified table name.
func (UserProductAccessModel) TableName() string {
	return "wisa_crm_db.user_product_access"
}
