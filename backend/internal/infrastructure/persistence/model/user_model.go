package model

import (
	"time"

	"github.com/google/uuid"
)

// UserModel is the GORM model for wisa_crm_db.users.
type UserModel struct {
	ID          uuid.UUID  `gorm:"column:id;type:uuid;primaryKey"`
	TenantID    uuid.UUID  `gorm:"column:tenant_id;type:uuid;index"`
	Name        string     `gorm:"column:name;type:varchar(255)"`
	Email       string     `gorm:"column:email;type:varchar(320)"`
	PasswordHash string    `gorm:"column:password_hash;type:varchar(72)"`
	Status      string     `gorm:"column:status"`
	LastLoginAt *time.Time `gorm:"column:last_login_at"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at"`
}

// TableName returns the fully qualified table name.
func (UserModel) TableName() string {
	return "wisa_crm_db.users"
}
