package entity

import "github.com/google/uuid"

// User represents a user belonging to a tenant.
// Domain entity — no GORM tags.
type User struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	Email        string
	PasswordHash string
	Status       string // active, blocked
}
