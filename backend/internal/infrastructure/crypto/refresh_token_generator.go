package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"

	"wisa-crm-service/backend/internal/domain/service"
)

// RefreshTokenGeneratorImpl implements service.RefreshTokenGenerator.
var _ service.RefreshTokenGenerator = (*RefreshTokenGeneratorImpl)(nil)

// RefreshTokenGeneratorImpl generates cryptographically random refresh tokens.
type RefreshTokenGeneratorImpl struct{}

// NewRefreshTokenGenerator creates a new RefreshTokenGenerator.
func NewRefreshTokenGenerator() *RefreshTokenGeneratorImpl {
	return &RefreshTokenGeneratorImpl{}
}

// Generate creates a cryptographically random refresh token and its SHA-256 hash.
// Returns (plainToken, hashHex, error). hashHex is 64 chars (SHA-256 in hex).
func (g *RefreshTokenGeneratorImpl) Generate() (plain string, hash string, err error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}
	plain = hex.EncodeToString(bytes)
	h := sha256.Sum256([]byte(plain))
	hash = hex.EncodeToString(h[:])
	return plain, hash, nil
}
