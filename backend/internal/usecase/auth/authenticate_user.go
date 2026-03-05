package auth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"

	"wisa-crm-service/backend/internal/domain"
	"wisa-crm-service/backend/internal/domain/repository"
	"wisa-crm-service/backend/internal/domain/service"
)

// Dummy bcrypt hash for timing normalization when user does not exist.
// Per ADR-010: always run bcrypt.Compare to prevent user enumeration via timing.
const dummyBcryptHash = "$2a$12$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"

const authCodeTTLSeconds = 40

// LoginInput contains the login request data.
type LoginInput struct {
	Slug        string
	ProductSlug string
	UserEmail   string
	Password    string
	State       string // CSRF token for callback validation
}

// LoginOutput contains the login response (HTTP 302 redirect URL).
type LoginOutput struct {
	RedirectURL string
}

// AuthenticateUserUseCase orchestrates the login flow.
// Per ADR-010: returns redirect URL with code; client exchanges code for JWT via POST /auth/token.
type AuthenticateUserUseCase struct {
	tenantRepo          repository.TenantRepository
	productRepo         repository.ProductRepository
	userRepo            repository.UserRepository
	subscriptionRepo    repository.SubscriptionRepository
	userProductAccRepo  repository.UserProductAccessRepository
	passwordSvc         service.PasswordService
	authCodeStore       service.AuthCodeStore
	redirectBaseDomain  string
}

// NewAuthenticateUserUseCase creates a new AuthenticateUserUseCase.
func NewAuthenticateUserUseCase(
	tenantRepo repository.TenantRepository,
	productRepo repository.ProductRepository,
	userRepo repository.UserRepository,
	subscriptionRepo repository.SubscriptionRepository,
	userProductAccRepo repository.UserProductAccessRepository,
	passwordSvc service.PasswordService,
	authCodeStore service.AuthCodeStore,
	redirectBaseDomain string,
) *AuthenticateUserUseCase {
	return &AuthenticateUserUseCase{
		tenantRepo:         tenantRepo,
		productRepo:        productRepo,
		userRepo:           userRepo,
		subscriptionRepo:   subscriptionRepo,
		userProductAccRepo: userProductAccRepo,
		passwordSvc:        passwordSvc,
		authCodeStore:      authCodeStore,
		redirectBaseDomain: redirectBaseDomain,
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

	// 7. Generate authorization code and store in Redis
	aud := tenant.Slug + "." + uc.redirectBaseDomain
	code, err := generateAuthCode()
	if err != nil {
		return nil, err
	}
	data := &service.AuthCodeData{
		Subject:           user.ID.String(),
		Audience:          aud,
		TenantID:          tenant.ID.String(),
		ProductID:         product.ID.String(),
		UserAccessProfile: accessProfile,
	}
	if err := uc.authCodeStore.Store(ctx, code, data, authCodeTTLSeconds); err != nil {
		log.Printf("[AuthenticateUser] auth code store failed: %v", err)
		return nil, domain.ErrAuthCodeStorageUnavailable
	}

	// 8. Build redirect URL (tenant_slug.base_domain/product_slug/callback?code=...&state=...)
	stateEnc := url.QueryEscape(input.State)
	redirectURL := fmt.Sprintf("https://%s.%s/%s/callback?code=%s&state=%s",
		tenant.Slug, uc.redirectBaseDomain, product.Slug, code, stateEnc)

	return &LoginOutput{RedirectURL: redirectURL}, nil
}
