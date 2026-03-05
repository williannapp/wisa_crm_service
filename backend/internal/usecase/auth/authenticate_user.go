package auth

import (
	"context"
	"errors"

	"wisa-crm-service/backend/internal/domain"
	"wisa-crm-service/backend/internal/domain/repository"
	"wisa-crm-service/backend/internal/domain/service"
)

// Dummy bcrypt hash for timing normalization when user does not exist.
// Per ADR-010: always run bcrypt.Compare to prevent user enumeration via timing.
const dummyBcryptHash = "$2a$12$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"

// LoginInput contains the login request data.
type LoginInput struct {
	Slug        string
	ProductSlug string
	UserEmail   string
	Password    string
}

// LoginOutput contains the login response.
type LoginOutput struct {
	Token string
}

// AuthenticateUserUseCase orchestrates the login flow.
type AuthenticateUserUseCase struct {
	tenantRepo         repository.TenantRepository
	productRepo        repository.ProductRepository
	userRepo           repository.UserRepository
	subscriptionRepo    repository.SubscriptionRepository
	userProductAccRepo repository.UserProductAccessRepository
	passwordSvc        service.PasswordService
	jwtSvc             service.JWTService
	audienceBaseDomain string
}

// NewAuthenticateUserUseCase creates a new AuthenticateUserUseCase.
func NewAuthenticateUserUseCase(
	tenantRepo repository.TenantRepository,
	productRepo repository.ProductRepository,
	userRepo repository.UserRepository,
	subscriptionRepo repository.SubscriptionRepository,
	userProductAccRepo repository.UserProductAccessRepository,
	passwordSvc service.PasswordService,
	jwtSvc service.JWTService,
	audienceBaseDomain string,
) *AuthenticateUserUseCase {
	return &AuthenticateUserUseCase{
		tenantRepo:         tenantRepo,
		productRepo:        productRepo,
		userRepo:           userRepo,
		subscriptionRepo:   subscriptionRepo,
		userProductAccRepo: userProductAccRepo,
		passwordSvc:        passwordSvc,
		jwtSvc:             jwtSvc,
		audienceBaseDomain: audienceBaseDomain,
	}
}

// Execute runs the authentication flow.
func (uc *AuthenticateUserUseCase) Execute(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	// 1. Tenant
	tenant, err := uc.tenantRepo.FindBySlug(ctx, input.Slug)
	if err != nil {
		if errors.Is(err, domain.ErrTenantNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	// 2. Product
	product, err := uc.productRepo.FindBySlug(ctx, input.ProductSlug)
	if err != nil {
		if errors.Is(err, domain.ErrProductNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}
	if product.Status != "active" {
		return nil, domain.ErrProductUnavailable
	}

	// 3. User + Password (timing-safe: always run bcrypt)
	user, err := uc.userRepo.FindByEmailAndTenantID(ctx, tenant.ID, input.UserEmail)
	hashToCompare := dummyBcryptHash
	if err == nil {
		hashToCompare = user.PasswordHash
	}
	if !uc.passwordSvc.Compare(input.Password, hashToCompare) {
		return nil, domain.ErrInvalidCredentials
	}
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	// 4. User status
	if user.Status != "active" {
		return nil, domain.ErrUserBlocked
	}

	// 5. Subscription
	subscription, err := uc.subscriptionRepo.FindByTenantAndProduct(ctx, tenant.ID, product.ID)
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

	// 6. Access profile (default "view" if not found)
	accessProfile := "view"
	upa, err := uc.userProductAccRepo.FindByUserAndProduct(ctx, user.ID, product.ID)
	if err == nil && upa != nil {
		accessProfile = upa.AccessProfile
	}

	// 7. Build JWT claims and sign
	aud := tenant.Slug + "." + uc.audienceBaseDomain
	claims := service.JWTClaims{
		Subject:           user.ID.String(),
		Audience:           aud,
		TenantID:           tenant.ID.String(),
		UserAccessProfile:  accessProfile,
	}
	token, err := uc.jwtSvc.Sign(ctx, claims)
	if err != nil {
		return nil, err
	}

	return &LoginOutput{Token: token}, nil
}
