package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"

	"wisa-crm-service/backend/internal/domain"
	"wisa-crm-service/backend/internal/domain/entity"
	"wisa-crm-service/backend/internal/domain/repository"
	"wisa-crm-service/backend/internal/domain/service"
)

// RefreshTokenInput contains the refresh request data.
type RefreshTokenInput struct {
	RefreshToken string
	ProductSlug  string
	TenantSlug   string
}

// RefreshTokenOutput contains the refresh response.
type RefreshTokenOutput struct {
	AccessToken      string
	ExpiresIn        int
	RefreshToken     string
	RefreshExpiresIn int
}

// RefreshTokenUseCase orchestrates the refresh token rotation flow.
type RefreshTokenUseCase struct {
	tenantRepo          repository.TenantRepository
	productRepo         repository.ProductRepository
	subscriptionRepo    repository.SubscriptionRepository
	userProductAccRepo  repository.UserProductAccessRepository
	refreshTokenRepo    repository.RefreshTokenRepository
	jwtSvc              service.JWTService
	refreshTokenGen     service.RefreshTokenGenerator
	redirectBaseDomain  string
}

// NewRefreshTokenUseCase creates a new RefreshTokenUseCase.
func NewRefreshTokenUseCase(
	tenantRepo repository.TenantRepository,
	productRepo repository.ProductRepository,
	subscriptionRepo repository.SubscriptionRepository,
	userProductAccRepo repository.UserProductAccessRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	jwtSvc service.JWTService,
	refreshTokenGen service.RefreshTokenGenerator,
	redirectBaseDomain string,
) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{
		tenantRepo:         tenantRepo,
		productRepo:        productRepo,
		subscriptionRepo:   subscriptionRepo,
		userProductAccRepo: userProductAccRepo,
		refreshTokenRepo:   refreshTokenRepo,
		jwtSvc:             jwtSvc,
		refreshTokenGen:    refreshTokenGen,
		redirectBaseDomain: redirectBaseDomain,
	}
}

// Execute runs the refresh token rotation. Returns 401 for any invalid token (generic message).
func (uc *RefreshTokenUseCase) Execute(ctx context.Context, input RefreshTokenInput) (*RefreshTokenOutput, error) {
	// Always compute hash first (timing constant)
	hashBytes := sha256.Sum256([]byte(input.RefreshToken))
	tokenHash := hex.EncodeToString(hashBytes[:])

	tenant, err := uc.tenantRepo.FindBySlug(ctx, input.TenantSlug)
	tenantID := uuid.Nil
	if err == nil && tenant != nil {
		tenantID = tenant.ID
	} else {
		// Tenant not found — use Nil; query will not match; return 401
		if !errors.Is(err, domain.ErrTenantNotFound) && err != nil {
			return nil, err
		}
		rt, _ := uc.refreshTokenRepo.FindByHashAndTenantAndProduct(ctx, tokenHash, uuid.Nil, uuid.Nil)
		if rt == nil {
			return nil, domain.ErrRefreshTokenInvalid
		}
	}

	product, err := uc.productRepo.FindBySlug(ctx, input.ProductSlug)
	productID := uuid.Nil
	if err == nil && product != nil {
		productID = product.ID
	} else {
		if !errors.Is(err, domain.ErrProductNotFound) && err != nil {
			return nil, err
		}
		rt, _ := uc.refreshTokenRepo.FindByHashAndTenantAndProduct(ctx, tokenHash, tenantID, uuid.Nil)
		if rt == nil {
			return nil, domain.ErrRefreshTokenInvalid
		}
	}

	rt, err := uc.refreshTokenRepo.FindByHashAndTenantAndProduct(ctx, tokenHash, tenantID, productID)
	if err != nil {
		return nil, err
	}
	if rt == nil {
		return nil, domain.ErrRefreshTokenInvalid
	}

	// Verify subscription is active
	subscription, err := uc.subscriptionRepo.FindByTenantAndProduct(ctx, tenantID, productID)
	if err != nil {
		if errors.Is(err, domain.ErrSubscriptionExpired) {
			return nil, domain.ErrSubscriptionExpired
		}
		return nil, err
	}
	switch subscription.Status {
	case "suspended":
		return nil, domain.ErrSubscriptionSuspended
	case "canceled":
		return nil, domain.ErrSubscriptionCanceled
	case "active", "pending":
		// ok
	default:
		return nil, domain.ErrSubscriptionExpired
	}

	// Access profile for JWT
	accessProfile := "view"
	upa, err := uc.userProductAccRepo.FindByUserAndProduct(ctx, rt.UserID, rt.ProductID)
	if err == nil && upa != nil {
		accessProfile = upa.AccessProfile
	}

	aud := tenant.Slug + "." + uc.redirectBaseDomain
	claims := service.JWTClaims{
		Subject:           rt.UserID.String(),
		Audience:          aud,
		TenantID:          rt.TenantID.String(),
		UserAccessProfile: accessProfile,
	}
	accessToken, err := uc.jwtSvc.Sign(ctx, claims)
	if err != nil {
		return nil, err
	}

	plainRT, hashRT, err := uc.refreshTokenGen.Generate()
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	expiresAt := now.Add(7 * 24 * time.Hour)
	newRT := &entity.RefreshToken{
		UserID:    rt.UserID,
		TenantID:  rt.TenantID,
		ProductID: rt.ProductID,
		TokenHash: hashRT,
		ExpiresAt: expiresAt,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := uc.refreshTokenRepo.Rotate(ctx, rt.ID, newRT); err != nil {
		return nil, err
	}

	return &RefreshTokenOutput{
		AccessToken:      accessToken,
		ExpiresIn:        jwtExpiresInSeconds,
		RefreshToken:     plainRT,
		RefreshExpiresIn: refreshTokenExpiresInSeconds,
	}, nil
}
