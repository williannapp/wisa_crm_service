package crypto

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"math/big"
	"os"

	"github.com/golang-jwt/jwt/v5"

	"wisa-crm-service/backend/internal/domain/service"
)

// rsaJWKSProvider implements service.JWKSProvider using an RSA private key file.
// Per ADR-006: extracts public key and formats in RFC 7517 JWK.
type rsaJWKSProvider struct {
	keys []service.JWK
}

// NewRSAJWKSProvider creates a JWKSProvider from the same config as RSAJWTService.
// Fail-fast if the key cannot be loaded. Keys are computed once at construction.
func NewRSAJWKSProvider(cfg RSAJWTConfig) (service.JWKSProvider, error) {
	if cfg.PrivateKeyPath == "" {
		return nil, fmt.Errorf("JWT_PRIVATE_KEY_PATH is required")
	}
	data, err := os.ReadFile(cfg.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(data)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}
	kid := cfg.KeyID
	if kid == "" {
		kid = "key-2026-v1"
	}
	jwk, err := rsaPublicKeyToJWK(privateKey.Public().(*rsa.PublicKey), kid)
	if err != nil {
		return nil, err
	}
	return &rsaJWKSProvider{keys: []service.JWK{jwk}}, nil
}

// GetKeys returns the cached JWKs. Implements service.JWKSProvider.
func (p *rsaJWKSProvider) GetKeys(ctx context.Context) ([]service.JWK, error) {
	return p.keys, nil
}

// rsaPublicKeyToJWK converts an RSA public key to RFC 7517 JWK format.
// n and e are base64url-encoded (without padding) per RFC 7518.
func rsaPublicKeyToJWK(pub *rsa.PublicKey, kid string) (service.JWK, error) {
	// N (modulus): big-endian unsigned integer, base64url
	nBytes := pub.N.FillBytes(make([]byte, (pub.N.BitLen()+7)/8))
	// E (exponent): typically 65537 (0x01 0x00 0x01) -> base64url "AQAB"
	eBig := big.NewInt(int64(pub.E))
	eBytes := eBig.FillBytes(make([]byte, (eBig.BitLen()+7)/8))
	if len(eBytes) == 0 {
		eBytes = []byte{0x01, 0x00, 0x01}
	}
	return service.JWK{
		Kty: "RSA",
		Use: "sig",
		Alg: "RS256",
		Kid: kid,
		N:   base64.RawURLEncoding.EncodeToString(nBytes),
		E:   base64.RawURLEncoding.EncodeToString(eBytes),
	}, nil
}
