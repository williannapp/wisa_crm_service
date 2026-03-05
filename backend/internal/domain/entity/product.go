package entity

import "github.com/google/uuid"

// Product represents a product/plan in the catalog.
// Domain entity — no GORM tags.
type Product struct {
	ID     uuid.UUID
	Slug   string
	Name   string
	Status string // active, inactive, blocked
}
