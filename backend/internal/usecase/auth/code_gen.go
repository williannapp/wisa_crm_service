package auth

import (
	"crypto/rand"
	"encoding/hex"
)

// generateAuthCode creates a cryptographically secure random authorization code.
// Returns 64 hex characters (32 bytes). Per ADR-010: must be unpredictable.
func generateAuthCode() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
