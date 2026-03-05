package model

import (
	"time"

	"github.com/google/uuid"
)

// SubscriptionModel is the GORM model for wisa_crm_db.subscriptions.
type SubscriptionModel struct {
	ID        uuid.UUID `gorm:"column:id;type:uuid;primaryKey"`
	TenantID  uuid.UUID `gorm:"column:tenant_id;type:uuid;index"`
	ProductID uuid.UUID `gorm:"column:product_id;type:uuid;index"`
	Type      string    `gorm:"column:type"`
	Status    string    `gorm:"column:status"`
	StartDate time.Time `gorm:"column:start_date;type:date"`
	EndDate   time.Time `gorm:"column:end_date;type:date"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

// TableName returns the fully qualified table name.
func (SubscriptionModel) TableName() string {
	return "wisa_crm_db.subscriptions"
}
