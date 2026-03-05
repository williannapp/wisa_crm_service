package entity

import (
	"time"

	"github.com/google/uuid"
)

// RefreshToken represents a refresh token record in the database.
// Domain entity — no GORM tags.
// Per ADR-006: token stored as SHA-256 hash, 7 days TTL, rotated on use.
type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TenantID  uuid.UUID
	ProductID uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
