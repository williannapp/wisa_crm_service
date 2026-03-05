package crypto

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"wisa-crm-service/backend/internal/domain/service"
)

// RSAJWTService implements domain.JWTService using RS256.
// Per ADR-006: RSA 4096 bits, kid in header, 15 min expiration.
type RSAJWTService struct {
	privateKey []byte
	issuer     string
	expMinutes int
	keyID      string
}

// RSAJWTConfig holds configuration for RSAJWTService.
type RSAJWTConfig struct {
	PrivateKeyPath string
	Issuer         string
	ExpMinutes     int
	KeyID          string
}

// NewRSAJWTService creates a new RSAJWTService. Fail fast if key cannot be loaded.
func NewRSAJWTService(cfg RSAJWTConfig) (*RSAJWTService, error) {
	if cfg.PrivateKeyPath == "" {
		return nil, fmt.Errorf("JWT_PRIVATE_KEY_PATH is required")
	}
	data, err := os.ReadFile(cfg.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}
	if cfg.ExpMinutes <= 0 {
		cfg.ExpMinutes = 15
	}
	if cfg.Issuer == "" {
		cfg.Issuer = "wisa-crm-service"
	}
	if cfg.KeyID == "" {
		cfg.KeyID = "key-2026-v1"
	}
	return &RSAJWTService{
		privateKey: data,
		issuer:     cfg.Issuer,
		expMinutes: cfg.ExpMinutes,
		keyID:      cfg.KeyID,
	}, nil
}

// Sign creates a JWT with RS256 and returns the encoded token string.
func (s *RSAJWTService) Sign(ctx context.Context, claims service.JWTClaims) (string, error) {
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(s.privateKey)
	if err != nil {
		return "", fmt.Errorf("invalid private key: %w", err)
	}

	// Use claims from input; ensure jti if empty
	jti := claims.JTI
	if jti == "" {
		jti = uuid.New().String()
	}
	now := time.Now()
	iat := claims.IssuedAt
	if iat == 0 {
		iat = now.Unix()
	}
	exp := claims.ExpiresAt
	if exp == 0 {
		exp = iat + int64(s.expMinutes*60)
	}
	nbf := claims.NotBefore
	if nbf == 0 {
		nbf = iat
	}

	mapClaims := jwt.MapClaims{
		"iss":                 s.issuer,
		"sub":                 claims.Subject,
		"aud":                 claims.Audience,
		"tenant_id":           claims.TenantID,
		"user_access_profile": claims.UserAccessProfile,
		"jti":                 jti,
		"iat":                 iat,
		"exp":                 exp,
		"nbf":                 nbf,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, mapClaims)
	token.Header["kid"] = s.keyID

	return token.SignedString(privateKey)
}
