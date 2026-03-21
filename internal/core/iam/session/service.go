package session

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"awo/internal/core/iam"
	"awo/internal/platform/cache"
	"awo/internal/shared/errors"
	"awo/internal/shared/logger"
	"awo/internal/shared/metrics"
	"awo/internal/shared/tracing"
)

// Config holds session-related settings.
// TODO(settings): read SessionTTL from settings key "iam.session_ttl_hours" per tenant
// FIXME(feature-flags): MFAEnabled should be read from feature-flag "iam.mfa_enabled"
type Config struct {
	SessionTTL time.Duration // default: 8h
	CookieName string        // default: "session"
}

// DefaultConfig returns safe defaults for development.
func DefaultConfig() Config {
	return Config{
		SessionTTL: 8 * time.Hour,
		CookieName: "session",
	}
}

// Service handles Login / ValidateSession / Logout.
type Service interface {
	Login(ctx context.Context, email, password string) (*ResolvedSession, string, error)
	ValidateSession(ctx context.Context, token string) (*ResolvedSession, error)
	Logout(ctx context.Context, token string) error
}

type service struct {
	identity iam.UserService
	authz    iam.Service
	repo     Repository
	cache    cache.Service
	tracer   tracing.Service
	metrics  metrics.MetricsProvider
	log      logger.Logger
	cfg      Config
}

// New constructs a session Service with default config.
func New(
	identitySvc iam.UserService,
	authzSvc iam.Service,
	repo Repository,
	cacheSvc cache.Service,
	tracer tracing.Service,
	metricsProv metrics.MetricsProvider,
	log logger.Logger,
) Service {
	return NewWithConfig(identitySvc, authzSvc, repo, cacheSvc, tracer, metricsProv, log, DefaultConfig())
}

// NewWithConfig constructs a session Service with explicit config.
func NewWithConfig(
	identitySvc iam.UserService,
	authzSvc iam.Service,
	repo Repository,
	cacheSvc cache.Service,
	tracer tracing.Service,
	metricsProv metrics.MetricsProvider,
	log logger.Logger,
	cfg Config,
) Service {
	return &service{
		identity: identitySvc,
		authz:    authzSvc,
		repo:     repo,
		cache:    cacheSvc,
		tracer:   tracer,
		metrics:  metricsProv,
		log:      log,
		cfg:      cfg,
	}
}

// Login authenticates the user, computes their permission map, persists the
// session, caches it, and returns the raw token to the caller.
//
// NOTE(tenant-context): ctx must carry tenant_id via cache.TenantIDKey.
func (s *service) Login(ctx context.Context, email, password string) (*ResolvedSession, string, error) {
	ctx, span := s.tracer.StartSpan(ctx, "session.Login")
	defer span.End()

	user, err := s.identity.Authenticate(ctx, email, password)
	if err != nil {
		s.metrics.IncrementCounter("session.login.failure", nil)
		return nil, "", err
	}

	perms, err := s.buildPermissions(ctx, user)
	if err != nil {
		s.log.WarnContext(ctx, "session: buildPermissions failed, continuing with empty map", logger.Fields{
			"user_id": user.ID.String(),
			"error":   err.Error(),
		})
		perms = make(map[string]bool)
	}

	rawToken, hash, err := generateToken()
	if err != nil {
		return nil, "", fmt.Errorf("session: generate token: %w", err)
	}

	tenantID := tenantIDFromCtx(ctx)

	// TODO(entity-scope): compute real EntityScope from entity hierarchy when entity service is wired.
	entityScope := EntityScope{
		Type:     EntityScopeEntity,
		EntityID: user.EntityID.String(),
	}
	// TODO(configuration): compute real flags+settings+prefs once services are wired.
	configuration := DefaultConfiguration()

	now := time.Now()
	sess := Session{
		UserID:        user.ID,
		TenantID:      tenantID,
		UserType:      user.UserType,
		TokenHash:     hash,
		Permissions:   perms,
		EntityScope:   entityScope,
		Configuration: configuration,
		IsActive:      true,
		ExpiresAt:     now.Add(s.cfg.SessionTTL),
		LastSeenAt:    now,
		RiskScore:     0,
	}
	if err := s.repo.CreateSession(ctx, sess); err != nil {
		return nil, "", fmt.Errorf("session: persist session: %w", err)
	}

	resolved := &ResolvedSession{
		UserID:      user.ID,
		UserType:    user.UserType,
		TenantID:    tenantID,
		DisplayName: displayName(user),
		Permissions: perms,
	}

	if err := s.cache.Set(ctx, sessionCacheKey(hash), resolved, s.cfg.SessionTTL); err != nil {
		s.log.WarnContext(ctx, "session: cache.Set failed", logger.Fields{"error": err.Error()})
	}

	s.metrics.IncrementCounter("session.login.success", nil)
	return resolved, rawToken, nil
}

// ValidateSession checks the cache first, then falls back to the DB.
func (s *service) ValidateSession(ctx context.Context, token string) (*ResolvedSession, error) {
	ctx, span := s.tracer.StartSpan(ctx, "session.ValidateSession")
	defer span.End()

	hash := sha256hex(token)

	var resolved ResolvedSession
	if err := s.cache.Get(ctx, sessionCacheKey(hash), &resolved); err == nil {
		s.repo.UpdateLastSeen(ctx, hash)
		return &resolved, nil
	}

	sess, err := s.repo.GetByTokenHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("session: get by hash: %w", err)
	}
	if sess == nil {
		return nil, errors.ErrAuthenticationFailed
	}

	r := &ResolvedSession{
		UserID:      sess.UserID,
		UserType:    sess.UserType,
		TenantID:    sess.TenantID,
		PrincipalID: sess.PrincipalID,
		Permissions: sess.Permissions,
	}

	ttl := time.Until(sess.ExpiresAt)
	if ttl > 0 {
		if err := s.cache.Set(ctx, sessionCacheKey(hash), r, ttl); err != nil {
			s.log.WarnContext(ctx, "session: re-populate cache failed", logger.Fields{"error": err.Error()})
		}
	}

	return r, nil
}

// Logout invalidates the DB session and removes the cache entry.
func (s *service) Logout(ctx context.Context, token string) error {
	ctx, span := s.tracer.StartSpan(ctx, "session.Logout")
	defer span.End()

	hash := sha256hex(token)

	if err := s.repo.Invalidate(ctx, hash); err != nil {
		return fmt.Errorf("session: logout invalidate: %w", err)
	}
	if err := s.cache.Delete(ctx, sessionCacheKey(hash)); err != nil {
		s.log.WarnContext(ctx, "session: cache.Delete failed on logout", logger.Fields{"error": err.Error()})
	}
	return nil
}

// buildPermissions uses authz.GetRoles + authz.GetPolicies to compute a flat
// permission map at login time.
func (s *service) buildPermissions(ctx context.Context, user *iam.User) (map[string]bool, error) {
	subject := subjectForUser(user)
	domain := domainForUser(user)

	roles, err := s.authz.GetRoles(ctx, subject, domain)
	if err != nil {
		return nil, fmt.Errorf("buildPermissions: GetRoles: %w", err)
	}

	policies, err := s.authz.GetPolicies(ctx, domain)
	if err != nil {
		return nil, fmt.Errorf("buildPermissions: GetPolicies: %w", err)
	}

	roleSet := make(map[string]bool, len(roles))
	for _, r := range roles {
		roleSet[r] = true
	}

	perms := make(map[string]bool)
	for _, p := range policies {
		if roleSet[p.Subject] && p.Effect == "allow" {
			perms[p.Object+"."+p.Action] = true
		}
	}

	s.metrics.IncrementCounter("session.permissions_computed", nil)
	return perms, nil
}

// --- helpers ---

func generateToken() (rawToken, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	rawToken = hex.EncodeToString(b)
	hash = sha256hex(rawToken)
	return rawToken, hash, nil
}

func sha256hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func tenantIDFromCtx(ctx context.Context) uuid.UUID {
	if v, ok := ctx.Value(cache.TenantIDKey).(string); ok && v != "" {
		if id, err := uuid.Parse(v); err == nil {
			return id
		}
	}
	return uuid.Nil
}

func subjectForUser(user *iam.User) string {
	id := user.ID.String()
	switch user.UserType {
	case string(iam.ActorPlatform):
		return iam.PlatformSubject(id)
	case string(iam.ActorPortal):
		return iam.PortalSubject(id)
	default:
		return iam.TenantSubject(id)
	}
}

func domainForUser(user *iam.User) string {
	switch user.UserType {
	case string(iam.ActorPlatform):
		return iam.DomainPlatform
	default:
		return iam.TenantDomain(user.TenantID.String())
	}
}

// displayName returns a human-readable name for the user.
// Prefers DisplayName, falls back to Username, then Email.
func displayName(user *iam.User) string {
	if user.DisplayName != nil && *user.DisplayName != "" {
		return *user.DisplayName
	}
	if user.Username != "" {
		return user.Username
	}
	return user.Email
}
