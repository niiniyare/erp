package service

// OAuth/OIDC SSO service.
//
// Flow:
// 1. BeginOAuth  — build the provider's authorization URL; store CSRF state in Redis.
// 2. ResolveUser — exchange code for tokens; fetch userinfo; JIT-provision user if
// provider.AutoProvision is true; return the User.
// 3. Caller passes the User to SessionService.LoginWithSSO to build a full session.
//
// Supported providers: google, microsoft.
// Implemented using stdlib net/http only — no external OAuth2 library required.

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"awo.so/internal/core/iam/domain"
	"awo.so/internal/core/iam/repository"
	"awo.so/internal/platform/cache"
	sharedErrors "awo.so/internal/shared/errors"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
)

// Endpoints (hardcoded per provider)

type oauthEndpoints struct {
	AuthURL  string
	TokenURL string
	UserInfo string
}

func providerEndpoints(p domain.OAuthProvider, extra map[string]string) oauthEndpoints {
	switch p {
	case domain.OAuthProviderGoogle:
		return oauthEndpoints{
			AuthURL:  "https://accounts.google.com/o/oauth2/v2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
			UserInfo: "https://www.googleapis.com/oauth2/v3/userinfo",
		}
	case domain.OAuthProviderMicrosoft:
		tenant := "common"
		if t, ok := extra["tenant"]; ok && t != "" {
			tenant = t
		}
		base := "https://login.microsoftonline.com/" + tenant
		return oauthEndpoints{
			AuthURL:  base + "/oauth2/v2.0/authorize",
			TokenURL: base + "/oauth2/v2.0/token",
			UserInfo: "https://graph.microsoft.com/oidc/userinfo",
		}
	default:
		return oauthEndpoints{}
	}
}

// Port (interface)

// SSOService manages OAuth/OIDC provider configurations and the login exchange flow.
type SSOService interface {
	// Provider management — ctx must carry tenant_id (cache.TenantIDKey).
	UpsertProvider(ctx context.Context, req *domain.CreateSSOProviderRequest) (*domain.SSOProvider, error)
	GetProvider(ctx context.Context, provider domain.OAuthProvider) (*domain.SSOProvider, error)
	ListProviders(ctx context.Context) ([]*domain.SSOProvider, error)
	DeactivateProvider(ctx context.Context, provider domain.OAuthProvider) error

	// OAuth flow
	// BeginOAuth generates a CSRF state, stores it in Redis, and returns the
	// redirect URL for the given provider. tenantID is taken from the query
	// param and injected into ctx before the repo call.
	BeginOAuth(ctx context.Context, tenantID uuid.UUID, provider domain.OAuthProvider) (authURL string, err error)

	// ResolveUser validates the OAuth callback (code + state), exchanges the code
	// for tokens, fetches user info from the provider, and returns a User.
	// Tenant context is recovered from the CSRF state stored in Redis.
	// Returns ErrForbidden if the user does not exist and AutoProvision is false.
	ResolveUser(ctx context.Context, provider domain.OAuthProvider, code, state string) (*domain.User, error)
}

// Config

// SSOConfig holds the constructor config for SSOService.
type SSOConfig struct {
	// EncryptionKey is the 32-byte AES-256 key used to encrypt client secrets at rest.
	// Must be set from a secure secret store (Vault, KMS) — never hard-code.
	EncryptionKey []byte

	// Cache is the Redis cache service used to store OAuth CSRF state tokens.
	Cache cache.Service

	// HTTPClient is the HTTP client for provider API calls.
	// Defaults to http.DefaultClient when nil.
	HTTPClient *http.Client
}

// Implementation

const ssoStateTTL = 10 * time.Minute

type ssoService struct {
	repo     repository.SSORepository
	identity UserService
	tracer   tracing.Service
	metrics  metrics.MetricsProvider
	log      logger.Logger
	cfg      SSOConfig
	http     *http.Client
}

// NewSSOService constructs an SSOService.
func NewSSOService(
	repo repository.SSORepository,
	identity UserService,
	tracer tracing.Service,
	m metrics.MetricsProvider,
	log logger.Logger,
	cfg SSOConfig,
) SSOService {
	h := cfg.HTTPClient
	if h == nil {
		h = &http.Client{Timeout: 15 * time.Second}
	}
	return &ssoService{
		repo:     repo,
		identity: identity,
		tracer:   tracer,
		metrics:  m,
		log:      log.WithFields(logger.Fields{"component": "iam.sso"}),
		cfg:      cfg,
		http:     h,
	}
}

// Provider management

func (s *ssoService) UpsertProvider(ctx context.Context, req *domain.CreateSSOProviderRequest) (*domain.SSOProvider, error) {
	ctx, span := s.tracer.StartSpan(ctx, "iam.sso.UpsertProvider")
	defer span.End()

	enc, err := encryptSecret(s.cfg.EncryptionKey, req.ClientSecret)
	if err != nil {
		return nil, fmt.Errorf("iam sso: encrypt client secret: %w", err)
	}
	return s.repo.UpsertProvider(ctx, req, enc)
}

func (s *ssoService) GetProvider(ctx context.Context, provider domain.OAuthProvider) (*domain.SSOProvider, error) {
	return s.repo.GetProvider(ctx, provider)
}

func (s *ssoService) ListProviders(ctx context.Context) ([]*domain.SSOProvider, error) {
	return s.repo.ListProviders(ctx)
}

func (s *ssoService) DeactivateProvider(ctx context.Context, provider domain.OAuthProvider) error {
	return s.repo.DeactivateProvider(ctx, provider)
}

// OAuth flow

// ssoState is stored in Redis under `sso:state:{token}` for CSRF validation.
type ssoState struct {
	TenantID uuid.UUID            `json:"tenant_id"`
	Provider domain.OAuthProvider `json:"provider"`
}

func (s *ssoService) BeginOAuth(ctx context.Context, tenantID uuid.UUID, provider domain.OAuthProvider) (string, error) {
	ctx, span := s.tracer.StartSpan(ctx, "iam.sso.BeginOAuth")
	defer span.End()

	// Inject tenant ID so the repo can call WithTenantFromCtx.
	ctx = context.WithValue(ctx, cache.TenantIDKey, tenantID)

	prov, err := s.repo.GetProvider(ctx, provider)
	if err != nil {
		return "", fmt.Errorf("iam sso: get provider: %w", err)
	}
	if prov == nil {
		return "", fmt.Errorf("iam sso: no active %s provider configured for tenant", provider)
	}

	// Generate a cryptographically random CSRF state token.
	stateBytes := make([]byte, 24)
	if _, err := rand.Read(stateBytes); err != nil {
		return "", fmt.Errorf("iam sso: generate state: %w", err)
	}
	stateToken := base64.URLEncoding.EncodeToString(stateBytes)

	// Persist state in Redis (global namespace — no tenant context required).
	statePayload := ssoState{TenantID: tenantID, Provider: provider}
	if err := s.cfg.Cache.Set(ctx, ssoStateKey(stateToken), statePayload, ssoStateTTL); err != nil {
		return "", fmt.Errorf("iam sso: store state: %w", err)
	}

	endpoints := providerEndpoints(provider, prov.ExtraParams)
	params := url.Values{
		"client_id":     {prov.ClientID},
		"redirect_uri":  {prov.RedirectURI},
		"response_type": {"code"},
		"scope":         {strings.Join(prov.Scopes, " ")},
		"state":         {stateToken},
		"access_type":   {"offline"}, // Google: request refresh token
		"prompt":        {"select_account"},
	}

	s.metrics.IncrementCounter("iam.sso.begin", metrics.Fields{"provider": string(provider)})
	return endpoints.AuthURL + "?" + params.Encode(), nil
}

func (s *ssoService) ResolveUser(ctx context.Context, provider domain.OAuthProvider, code, state string) (*domain.User, error) {
	ctx, span := s.tracer.StartSpan(ctx, "iam.sso.ResolveUser")
	defer span.End()

	// 1. Validate CSRF state.
	var statePayload ssoState
	if err := s.cfg.Cache.Get(ctx, ssoStateKey(state), &statePayload); err != nil {
		s.metrics.IncrementCounter("iam.sso.state.invalid", nil)
		return nil, sharedErrors.ErrUnauthorized
	}
	// State is single-use — delete immediately.
	_ = s.cfg.Cache.Delete(ctx, ssoStateKey(state))

	if statePayload.Provider != provider {
		return nil, sharedErrors.ErrUnauthorized
	}

	// 2. Inject tenant_id into context for all subsequent repo/service calls.
	ctx = context.WithValue(ctx, cache.TenantIDKey, statePayload.TenantID)

	// 3. Load provider config.
	prov, err := s.repo.GetProvider(ctx, provider)
	if err != nil || prov == nil {
		return nil, fmt.Errorf("iam sso: load provider: %w", err)
	}

	secret, err := decryptSecret(s.cfg.EncryptionKey, prov.ClientSecretEnc)
	if err != nil {
		return nil, fmt.Errorf("iam sso: decrypt secret: %w", err)
	}

	endpoints := providerEndpoints(provider, prov.ExtraParams)

	// 4. Exchange code for tokens.
	accessToken, err := s.exchangeCode(ctx, endpoints.TokenURL, prov.ClientID, secret, prov.RedirectURI, code)
	if err != nil {
		s.metrics.IncrementCounter("iam.sso.exchange.failure", metrics.Fields{"provider": string(provider)})
		return nil, fmt.Errorf("iam sso: token exchange: %w", err)
	}

	// 5. Fetch user info.
	userInfo, err := s.fetchUserInfo(ctx, endpoints.UserInfo, accessToken)
	if err != nil {
		return nil, fmt.Errorf("iam sso: fetch userinfo: %w", err)
	}
	userInfo.Provider = provider

	// 6. Look up or JIT-provision the user (ctx already has tenant_id).
	user, err := s.identity.GetUserByEmail(ctx, userInfo.Email)
	if err == nil {
		// Existing user — done.
		s.metrics.IncrementCounter("iam.sso.login.existing", metrics.Fields{"provider": string(provider)})
		return user, nil
	}

	if !prov.AutoProvision {
		s.metrics.IncrementCounter("iam.sso.login.forbidden", metrics.Fields{"provider": string(provider)})
		return nil, sharedErrors.ErrForbidden
	}

	if prov.DefaultEntityID == uuid.Nil {
		return nil, fmt.Errorf("iam sso: auto_provision=true but default_entity_id is not configured for provider %s", provider)
	}

	// JIT provision: create a new user with a random placeholder password.
	// The placeholder is bcrypt-hashed; the user cannot log in with a password
	// because they will never know the raw value.
	rawPlaceholder, err := generatePlaceholderPassword()
	if err != nil {
		return nil, fmt.Errorf("iam sso: generate placeholder password: %w", err)
	}

	displayName := userInfo.Name
	createReq := &domain.CreateUserRequest{
		EntityID:    prov.DefaultEntityID,
		Email:       userInfo.Email,
		Username:    strings.ToLower(strings.ReplaceAll(userInfo.Email, "@", "_at_")),
		Password:    rawPlaceholder,
		UserType:    "CUSTOMER",
		DisplayName: &displayName,
	}
	if err := createReq.Validate(); err != nil {
		return nil, fmt.Errorf("iam sso: jit provision validate: %w", err)
	}

	user, err = s.identity.RegisterNewUser(ctx, createReq)
	if err != nil {
		return nil, fmt.Errorf("iam sso: jit provision: %w", err)
	}

	s.log.InfoContext(ctx, "SSO JIT provisioned user", logger.Fields{
		"user_id":  user.ID.String(),
		"email":    user.Email,
		"provider": string(provider),
	})
	s.metrics.IncrementCounter("iam.sso.login.provisioned", metrics.Fields{"provider": string(provider)})
	return user, nil
}

// Internal helpers

// tokenResponse is the OAuth2 token endpoint response.
type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

func (s *ssoService) exchangeCode(ctx context.Context, tokenURL, clientID, clientSecret, redirectURI, code string) (string, error) {
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"redirect_uri":  {redirectURI},
		"code":          {code},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := s.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token endpoint returned %d: %s", resp.StatusCode, body)
	}

	var tok tokenResponse
	if err := json.Unmarshal(body, &tok); err != nil {
		return "", fmt.Errorf("parse token response: %w", err)
	}
	if tok.AccessToken == "" {
		return "", fmt.Errorf("token endpoint returned empty access_token")
	}
	return tok.AccessToken, nil
}

// userInfoResponse is the common subset returned by Google and Microsoft userinfo endpoints.
type userInfoResponse struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func (s *ssoService) fetchUserInfo(ctx context.Context, userInfoURL, accessToken string) (*domain.SSOUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, userInfoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := s.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo endpoint returned %d", resp.StatusCode)
	}

	var info userInfoResponse
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("parse userinfo: %w", err)
	}
	if info.Email == "" {
		return nil, fmt.Errorf("userinfo: email is empty")
	}

	return &domain.SSOUserInfo{
		Sub:   info.Sub,
		Email: strings.ToLower(strings.TrimSpace(info.Email)),
		Name:  info.Name,
	}, nil
}

func ssoStateKey(state string) string { return "sso:state:" + state }

// generatePlaceholderPassword creates a 32-byte random placeholder password for
// JIT-provisioned SSO users. It is hashed immediately; the user never sees it.
func generatePlaceholderPassword() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
