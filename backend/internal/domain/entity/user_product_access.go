package entity

import "github.com/google/uuid"

// UserProductAccess represents the access profile of a user for a product.
// Domain entity — no GORM tags.
// AccessProfile values: admin, operator, view (from enum access_profile).
type UserProductAccess struct {
	UserID        uuid.UUID
	ProductID     uuid.UUID
	TenantID      uuid.UUID
	AccessProfile string // admin, operator, view
}
