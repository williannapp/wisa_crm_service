package entity

import "github.com/google/uuid"

// Subscription links a tenant to a product.
// Domain entity — no GORM tags.
type Subscription struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	ProductID uuid.UUID
	Status    string // pending, active, suspended, canceled
}
