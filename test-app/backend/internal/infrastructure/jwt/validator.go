package jwt

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// leeway for clock skew per ADR-006
const leeway = 30 * time.Second

// Validator validates JWT tokens using JWKS.
type Validator struct {
	jwksFetcher *JWKSFetcher
	issuer      string
	audience    string
}

// NewValidator creates a new JWT Validator.
func NewValidator(jwks *JWKSFetcher) *Validator {
	authURL := os.Getenv("AUTH_SERVER_URL")
	if authURL == "" {
		authURL = "https://auth.wisa.labs.com.br"
	}
	if strings.HasSuffix(authURL, "/") {
		authURL = strings.TrimSuffix(authURL, "/")
	}
	issuer := os.Getenv("JWT_ISSUER")
	if issuer == "" {
		issuer = authURL
	}
	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		appURL = "http://localhost:8081"
	}
	audience := os.Getenv("JWT_AUDIENCE")
	if audience == "" {
		audience = strings.TrimPrefix(appURL, "https://")
		audience = strings.TrimPrefix(audience, "http://")
	}
	return &Validator{
		jwksFetcher: jwks,
		issuer:      issuer,
		audience:    audience,
	}
}

// Validate parses and validates the token. Returns claims or error.
func (v *Validator) Validate(ctx context.Context, tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		alg, ok := token.Header["alg"].(string)
		if !ok || alg != "RS256" {
			return nil, fmt.Errorf("invalid alg: expected RS256")
		}
		kid, _ := token.Header["kid"].(string)
		if kid == "" {
			return nil, fmt.Errorf("missing kid")
		}
		return v.jwksFetcher.GetPublicKey(ctx, kid)
	}, jwt.WithValidMethods([]string{"RS256"}))

	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("token invalid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}

	if iss, _ := claims["iss"].(string); iss != v.issuer {
		return nil, fmt.Errorf("invalid iss: expected %s", v.issuer)
	}

	if aud, _ := claims["aud"].(string); aud != v.audience {
		return nil, fmt.Errorf("invalid aud: expected %s", v.audience)
	}

	exp, ok := claims["exp"]
	if !ok {
		return nil, fmt.Errorf("missing exp")
	}
	expNum, ok := exp.(float64)
	if !ok {
		return nil, fmt.Errorf("invalid exp")
	}
	if time.Now().Unix() > int64(expNum)+int64(leeway.Seconds()) {
		return nil, fmt.Errorf("token expired")
	}

	if nbf, ok := claims["nbf"]; ok {
		nbfNum, ok := nbf.(float64)
		if ok && time.Now().Unix() < int64(nbfNum)-int64(leeway.Seconds()) {
			return nil, fmt.Errorf("token not yet valid")
		}
	}

	return claims, nil
}
