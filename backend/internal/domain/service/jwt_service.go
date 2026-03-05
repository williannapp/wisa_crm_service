package service

import "context"

// JWTClaims contains the claims for the JWT payload.
// Per ADR-006: iss, sub, aud, tenant_id, user_access_profile, jti, iat, exp, nbf.
type JWTClaims struct {
	Issuer            string `json:"iss"`
	Subject           string `json:"sub"`
	Audience          string `json:"aud"`
	TenantID          string `json:"tenant_id"`
	UserAccessProfile string `json:"user_access_profile"`
	JTI               string `json:"jti"`
	IssuedAt          int64  `json:"iat"`
	ExpiresAt         int64  `json:"exp"`
	NotBefore         int64  `json:"nbf"`
}

// JWTService defines the port for JWT signing.
// Per ADR-006: RS256, 15 min expiration, kid in header.
type JWTService interface {
	Sign(ctx context.Context, claims JWTClaims) (string, error)
}
