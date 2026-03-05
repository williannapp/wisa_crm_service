package service

// RefreshTokenGenerator generates cryptographically secure refresh tokens.
// Per ADR-006: returns (plainToken, sha256HashHex). Hash is stored in DB; plain returned to client.
type RefreshTokenGenerator interface {
	Generate() (plain string, hash string, err error)
}
