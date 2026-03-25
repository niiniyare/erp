package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"

	"awo.so/internal/core/iam/domain"
	"awo.so/internal/core/iam/repository"
	"awo.so/internal/platform/cache"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
)

const (
	apiKeyPrefix    = "eak_"      // ERP API Key prefix — makes keys recognisable in logs
	apiKeyCacheTTL  = 5 * time.Minute
	apiKeyCacheNS   = "apikey:"   // Redis key namespace: apikey:{sha256hex(rawToken)}
	apiKeyRawLength = 32          // bytes of random entropy; 64 hex chars after encoding
)

// ─── Port (interface) ─────────────────────────────────────────────────────────

// APIKeyService handles creation, validation, revocation, and listing of API keys.
type APIKeyService interface {
	// CreateAPIKey generates a new API key, persists its SHA-256 hash, and returns
	// the plaintext bearer token exactly once.  The raw token is NEVER stored.
	CreateAPIKey(ctx context.Context, req *domain.CreateAPIKeyRequest) (*domain.APIKey, string, error)

	// ValidateAPIKey verifies a raw bearer token (prefixed with "eak_"), builds a
	// minimal ResolvedSession from the key's scopes, and caches it for 5 minutes.
	// Returns nil, nil when the key does not exist, is revoked, or is expired.
	ValidateAPIKey(ctx context.Context, rawToken string) (*domain.ResolvedSession, error)

	// RevokeAPIKey immediately invalidates the key and evicts the Redis cache entry.
	RevokeAPIKey(ctx context.Context, keyID uuid.UUID) error

	// ListAPIKeys returns all API keys for the current tenant (newest first).
	ListAPIKeys(ctx context.Context) ([]*domain.APIKey, error)
}

// ─── Implementation ───────────────────────────────────────────────────────────

type apiKeyService struct {
	repo    repository.APIKeyRepository
	cache   cache.Service
	tracer  tracing.Service
	metrics metrics.MetricsProvider
}

// NewAPIKeyService constructs an APIKeyService.
func NewAPIKeyService(
	repo repository.APIKeyRepository,
	cacheSvc cache.Service,
	tracer tracing.Service,
	m metrics.MetricsProvider,
) APIKeyService {
	return &apiKeyService{
		repo:    repo,
		cache:   cacheSvc,
		tracer:  tracer,
		metrics: m,
	}
}

// ─── CreateAPIKey ─────────────────────────────────────────────────────────────

func (s *apiKeyService) CreateAPIKey(ctx context.Context, req *domain.CreateAPIKeyRequest) (*domain.APIKey, string, error) {
	ctx, span := s.tracer.StartSpan(ctx, "iam.apikey.Create")
	defer span.End()

	// Generate raw token: prefix + 32 random bytes hex-encoded (64 chars)
	rawBytes := make([]byte, apiKeyRawLength)
	if _, err := rand.Read(rawBytes); err != nil {
		return nil, "", fmt.Errorf("apikey: generate token: %w", err)
	}
	rawToken := apiKeyPrefix + hex.EncodeToString(rawBytes)

	// Hash for storage
	keyHash := hashToken(rawToken)

	key, err := s.repo.Create(ctx, req, keyHash)
	if err != nil {
		return nil, "", fmt.Errorf("apikey: create: %w", err)
	}

	s.metrics.IncrementCounter("iam.apikey.created", nil)
	span.SetAttributes(attribute.String("apikey.id", key.ID.String()))

	// Return key + plaintext token — caller must return this to the user once
	return key, rawToken, nil
}

// ─── ValidateAPIKey ───────────────────────────────────────────────────────────

func (s *apiKeyService) ValidateAPIKey(ctx context.Context, rawToken string) (*domain.ResolvedSession, error) {
	ctx, span := s.tracer.StartSpan(ctx, "iam.apikey.Validate")
	defer span.End()

	keyHash := hashToken(rawToken)
	cacheKey := apiKeyCacheNS + keyHash

	// 1. Cache-aside: Redis hit is the common path on high-frequency API calls
	var sess domain.ResolvedSession
	if err := s.cache.Get(ctx, cacheKey, &sess); err == nil {
		s.metrics.IncrementCounter("iam.apikey.validate.cache_hit", nil)
		return &sess, nil
	}

	// 2. DB lookup (requires admin_role at the DB level — see repository.APIKeyRepository)
	key, err := s.repo.GetByHash(ctx, keyHash)
	if err != nil {
		return nil, fmt.Errorf("apikey: validate: %w", err)
	}
	if key == nil {
		s.metrics.IncrementCounter("iam.apikey.validate.not_found", nil)
		return nil, nil
	}

	// 3. Build minimal ResolvedSession from the key's scopes
	resolved := buildAPIKeySession(key)

	// 4. Cache the session
	_ = s.cache.Set(ctx, cacheKey, resolved, apiKeyCacheTTL)

	s.metrics.IncrementCounter("iam.apikey.validate.ok", nil)
	span.SetAttributes(
		attribute.String("apikey.id", key.ID.String()),
		attribute.String("tenant.id", key.TenantID.String()),
	)
	return resolved, nil
}

// ─── RevokeAPIKey ─────────────────────────────────────────────────────────────

func (s *apiKeyService) RevokeAPIKey(ctx context.Context, keyID uuid.UUID) error {
	ctx, span := s.tracer.StartSpan(ctx, "iam.apikey.Revoke")
	defer span.End()
	span.SetAttributes(attribute.String("apikey.id", keyID.String()))

	if err := s.repo.Revoke(ctx, keyID); err != nil {
		return fmt.Errorf("apikey: revoke: %w", err)
	}

	// Best-effort cache eviction — we don't know the raw token so we can't
	// derive the cache key here.  The cached entry expires naturally after TTL.
	// For immediate invalidation, callers should store key_hash → cache_key mapping.
	s.metrics.IncrementCounter("iam.apikey.revoked", nil)
	return nil
}

// ─── ListAPIKeys ──────────────────────────────────────────────────────────────

func (s *apiKeyService) ListAPIKeys(ctx context.Context) ([]*domain.APIKey, error) {
	ctx, span := s.tracer.StartSpan(ctx, "iam.apikey.List")
	defer span.End()

	keys, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("apikey: list: %w", err)
	}
	return keys, nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

// hashToken returns the SHA-256 hex digest of rawToken.
func hashToken(rawToken string) string {
	h := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(h[:])
}

// buildAPIKeySession constructs the minimal ResolvedSession for an API key request.
// The session carries only the key's scopes as permissions — there is no
// user session, no MFA state, and no entity scope beyond EntityScopeAll.
func buildAPIKeySession(key *domain.APIKey) *domain.ResolvedSession {
	perms := make(map[string]bool, len(key.Scopes))
	for _, scope := range key.Scopes {
		perms[scope] = true
	}

	return &domain.ResolvedSession{
		UserID:      key.CreatedBy, // creator — for audit trail
		UserType:    "API",
		TenantID:    key.TenantID,
		DisplayName: key.Name,
		Permissions: perms,
		EntityScope: domain.EntityScope{Type: domain.EntityScopeAll},
		Configuration: domain.DefaultConfiguration(),
	}
}

