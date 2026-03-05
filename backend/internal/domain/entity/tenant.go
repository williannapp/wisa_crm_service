package entity

import "github.com/google/uuid"

// Tenant represents a client (multi-tenant) in the system.
// Domain entity — no GORM tags.
type Tenant struct {
	ID     uuid.UUID
	Slug   string
	Name   string
	Status string // active, inactive, blocked
}
