package auth

import (
	"context"
	"errors"

	"wisa-crm-service/backend/internal/domain"
	"wisa-crm-service/backend/internal/domain/service"
)

const jwtExpiresInSeconds = 900 // 15 min per ADR-006

// ExchangeCodeInput contains the token exchange request.
type ExchangeCodeInput struct {
	Code string
}

// ExchangeCodeOutput contains the token response.
type ExchangeCodeOutput struct {
	AccessToken string
	ExpiresIn   int
}

// ExchangeCodeForTokenUseCase exchanges an authorization code for a JWT.
type ExchangeCodeForTokenUseCase struct {
	authCodeStore service.AuthCodeStore
	jwtSvc        service.JWTService
}

// NewExchangeCodeForTokenUseCase creates a new ExchangeCodeForTokenUseCase.
func NewExchangeCodeForTokenUseCase(authCodeStore service.AuthCodeStore, jwtSvc service.JWTService) *ExchangeCodeForTokenUseCase {
	return &ExchangeCodeForTokenUseCase{
		authCodeStore: authCodeStore,
		jwtSvc:        jwtSvc,
	}
}

// Execute exchanges the code for a JWT. The code is consumed (single-use).
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

	return &ExchangeCodeOutput{
		AccessToken: token,
		ExpiresIn:   jwtExpiresInSeconds,
	}, nil
}
