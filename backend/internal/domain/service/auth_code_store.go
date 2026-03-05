package service

import "context"

// AuthCodeData contains the data needed to issue the JWT when exchanging the code.
type AuthCodeData struct {
	Subject           string // user_id (ULID)
	Audience          string
	TenantID          string
	UserAccessProfile string
}

// AuthCodeStore stores authorization codes temporarily.
type AuthCodeStore interface {
	Store(ctx context.Context, code string, data *AuthCodeData, ttlSeconds int) error
	GetAndDelete(ctx context.Context, code string) (*AuthCodeData, error)
}
