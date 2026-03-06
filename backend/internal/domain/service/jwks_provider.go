package service

import "context"

// JWK represents a single key in JSON Web Key format (RFC 7517).
type JWK struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// JWKS is the JSON Web Key Set (RFC 7517).
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// JWKSProvider provides public keys for JWT signature verification.
// Per ADR-006: supports multiple keys for zero-downtime rotation.
type JWKSProvider interface {
	GetKeys(ctx context.Context) ([]JWK, error)
}
