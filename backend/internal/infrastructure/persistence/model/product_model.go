package model

import (
	"time"

	"github.com/google/uuid"
)

// ProductModel is the GORM model for wisa_crm_db.products.
type ProductModel struct {
	ID        uuid.UUID `gorm:"column:id;type:uuid;primaryKey"`
	Slug      string    `gorm:"column:slug;type:varchar(63);uniqueIndex"`
	Name      string    `gorm:"column:name;type:varchar(255)"`
	Status    string    `gorm:"column:status"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

// TableName returns the fully qualified table name.
func (ProductModel) TableName() string {
	return "wisa_crm_db.products"
}
