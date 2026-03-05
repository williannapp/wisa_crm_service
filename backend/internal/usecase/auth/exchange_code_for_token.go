package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"wisa-crm-service/backend/internal/domain"
	"wisa-crm-service/backend/internal/domain/entity"
	"wisa-crm-service/backend/internal/domain/repository"
	"wisa-crm-service/backend/internal/domain/service"
)

const jwtExpiresInSeconds = 900       // 15 min per ADR-006
const refreshTokenExpiresInSeconds = 7 * 24 * 3600 // 7 days per ADR-006

// ExchangeCodeInput contains the token exchange request.
type ExchangeCodeInput struct {
	Code string
}

// ExchangeCodeOutput contains the token response.
type ExchangeCodeOutput struct {
	AccessToken      string
	ExpiresIn        int
	RefreshToken     string
	RefreshExpiresIn int
}

// ExchangeCodeForTokenUseCase exchanges an authorization code for a JWT and refresh token.
type ExchangeCodeForTokenUseCase struct {
	authCodeStore     service.AuthCodeStore
	jwtSvc            service.JWTService
	refreshTokenRepo  repository.RefreshTokenRepository
	refreshTokenGen   service.RefreshTokenGenerator
}

// NewExchangeCodeForTokenUseCase creates a new ExchangeCodeForTokenUseCase.
func NewExchangeCodeForTokenUseCase(
	authCodeStore service.AuthCodeStore,
	jwtSvc service.JWTService,
	refreshTokenRepo repository.RefreshTokenRepository,
	refreshTokenGen service.RefreshTokenGenerator,
) *ExchangeCodeForTokenUseCase {
	return &ExchangeCodeForTokenUseCase{
		authCodeStore:    authCodeStore,
		jwtSvc:           jwtSvc,
		refreshTokenRepo: refreshTokenRepo,
		refreshTokenGen:  refreshTokenGen,
	}
}

// Execute exchanges the code for a JWT and refresh token. The code is consumed (single-use).
func (uc *ExchangeCodeForTokenUseCase) Execute(ctx context.Context, input ExchangeCodeInput) (*ExchangeCodeOutput, error) {
	data, err := uc.authCodeStore.GetAndDelete(ctx, input.Code)
	if err != nil {
		if errors.Is(err, domain.ErrCodeInvalidOrExpired) {
			return nil, domain.ErrCodeInvalidOrExpired
		}
		return nil, err
	}
	if data == nil {
		return nil, domain.ErrCodeInvalidOrExpired
	}

	// ProductID required for refresh token; old codes without it are invalid
	if data.ProductID == "" {
		return nil, domain.ErrCodeInvalidOrExpired
	}

	userID, err := uuid.Parse(data.Subject)
	if err != nil {
		return nil, domain.ErrCodeInvalidOrExpired
	}
	tenantID, err := uuid.Parse(data.TenantID)
	if err != nil {
		return nil, domain.ErrCodeInvalidOrExpired
	}
	productID, err := uuid.Parse(data.ProductID)
	if err != nil {
		return nil, domain.ErrCodeInvalidOrExpired
	}

	claims := service.JWTClaims{
		Subject:           data.Subject,
		Audience:          data.Audience,
		TenantID:          data.TenantID,
		UserAccessProfile: data.UserAccessProfile,
	}
	token, err := uc.jwtSvc.Sign(ctx, claims)
	if err != nil {
		return nil, err
	}

	plainRT, hashRT, err := uc.refreshTokenGen.Generate()
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	expiresAt := now.Add(7 * 24 * time.Hour)
	rt := &entity.RefreshToken{
		UserID:    userID,
		TenantID:  tenantID,
		ProductID: productID,
		TokenHash: hashRT,
		ExpiresAt: expiresAt,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := uc.refreshTokenRepo.Create(ctx, rt); err != nil {
		return nil, err
	}

	return &ExchangeCodeOutput{
		AccessToken:      token,
		ExpiresIn:        jwtExpiresInSeconds,
		RefreshToken:     plainRT,
		RefreshExpiresIn: refreshTokenExpiresInSeconds,
	}, nil
}
