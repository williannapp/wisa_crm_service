package jwt

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

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

// JWKSFetcher fetches and caches JWKS from the auth server.
type JWKSFetcher struct {
	authServerURL string
	cache         map[string]*rsa.PublicKey
	cacheExpiry   time.Time
	mu            sync.RWMutex
	httpClient    *http.Client
}

// NewJWKSFetcher creates a new JWKSFetcher.
func NewJWKSFetcher() *JWKSFetcher {
	authURL := os.Getenv("AUTH_SERVER_URL")
	if authURL == "" {
		authURL = "https://auth.wisa.labs.com.br"
	}
	if strings.HasSuffix(authURL, "/") {
		authURL = strings.TrimSuffix(authURL, "/")
	}
	return &JWKSFetcher{
		authServerURL: authURL,
		cache:        make(map[string]*rsa.PublicKey),
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

// GetPublicKey returns the RSA public key for the given kid.
// Fetches JWKS if cache is stale (older than 24h).
func (f *JWKSFetcher) GetPublicKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	f.mu.RLock()
	if time.Now().Before(f.cacheExpiry) {
		if pub, ok := f.cache[kid]; ok {
			f.mu.RUnlock()
			return pub, nil
		}
	}
	f.mu.RUnlock()

	f.mu.Lock()
	defer f.mu.Unlock()

	if time.Now().Before(f.cacheExpiry) {
		if pub, ok := f.cache[kid]; ok {
			return pub, nil
		}
	}

	url := f.authServerURL + "/.well-known/jwks.json"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("jwks request: %w", err)
	}

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("jwks fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("jwks status %d", resp.StatusCode)
	}

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("jwks decode: %w", err)
	}

	f.cache = make(map[string]*rsa.PublicKey)
	for _, jwk := range jwks.Keys {
		if jwk.Kty != "RSA" {
			continue
		}
		pub, err := jwkToRSA(&jwk)
		if err != nil {
			continue
		}
		f.cache[jwk.Kid] = pub
	}
	f.cacheExpiry = time.Now().Add(24 * time.Hour)

	if pub, ok := f.cache[kid]; ok {
		return pub, nil
	}
	return nil, fmt.Errorf("key not found: %s", kid)
}

func jwkToRSA(jwk *JWK) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, err
	}
	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)
	ei := int(e.Int64())
	if ei == 0 {
		ei = 65537
	}
	return &rsa.PublicKey{N: n, E: ei}, nil
}
