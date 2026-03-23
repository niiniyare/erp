package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"

	"awo.so/internal/core/iam/domain"
	"awo.so/internal/core/iam/repository"
	"awo.so/internal/platform/cache"
	"awo.so/internal/shared/errors"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
)

//  Port (interface)

// SessionService handles Login / ValidateSession / Logout.
type SessionService interface {
	Login(ctx context.Context, email, password string) (*domain.ResolvedSession, string, error)
	ValidateSession(ctx context.Context, token string) (*domain.ResolvedSession, error)
	Logout(ctx context.Context, token string) error

	// LogoutAllForUser invalidates all active DB sessions for the given user.
	// Existing cache entries expire naturally within the session TTL window.
	LogoutAllForUser(ctx context.Context, userID uuid.UUID) error

	// LogoutAllForTenant invalidates all active DB sessions for the given tenant.
	// Existing cache entries expire naturally within the session TTL window.
	LogoutAllForTenant(ctx context.Context, tenantID uuid.UUID) error
}

//  Implementation

type sessionService struct {
	identity UserService  // same package — no import needed
	authz    AuthzService // same package
	repo     repository.SessionRepository
	tracer   tracing.Service
	metrics  metrics.MetricsProvider
	log      logger.Logger
	cfg      domain.SessionConfig
}

// NewSessionService constructs a SessionService with default config.
func NewSessionService(
	identitySvc UserService,
	authzSvc AuthzService,
	repo repository.SessionRepository,
	tracer tracing.Service,
	m metrics.MetricsProvider,
	log logger.Logger,
) SessionService {
	return NewSessionServiceWithConfig(identitySvc, authzSvc, repo, tracer, m, log, domain.DefaultSessionConfig())
}

// NewSessionServiceWithConfig constructs a SessionService with explicit config.
func NewSessionServiceWithConfig(
	identitySvc UserService,
	authzSvc AuthzService,
	repo repository.SessionRepository,
	tracer tracing.Service,
	m metrics.MetricsProvider,
	log logger.Logger,
	cfg domain.SessionConfig,
) SessionService {
	return &sessionService{
		identity: identitySvc,
		authz:    authzSvc,
		repo:     repo,
		tracer:   tracer,
		metrics:  m,
		log:      log,
		cfg:      cfg,
	}
}

// Login authenticates the user, computes their permission map, persists the
// session, caches it, and returns the raw token to the caller.
//
// NOTE(tenant-context): ctx must carry tenant_id via cache.TenantIDKey.
func (s *sessionService) Login(ctx context.Context, email, password string) (*domain.ResolvedSession, string, error) {
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
			"user_id": user.ID.String(), "error": err.Error(),
		})
		perms = make(map[string]bool)
	}

	rawToken, hash, err := generateToken()
	if err != nil {
		return nil, "", fmt.Errorf("session: generate token: %w", err)
	}

	tenantID := tenantIDFromCtx(ctx)

	// TODO(entity-scope): compute real EntityScope from entity hierarchy.
	entityScope := domain.EntityScope{
		Type:     domain.EntityScopeEntity,
		EntityID: user.EntityID.String(),
	}
	// TODO(configuration): compute real flags+settings+prefs once wired.
	configuration := domain.DefaultConfiguration()

	now := time.Now()
	sess := domain.Session{
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
	}
	if err := s.repo.Create(ctx, sess); err != nil {
		return nil, "", fmt.Errorf("session: persist session: %w", err)
	}

	resolved := &domain.ResolvedSession{
		UserID:        user.ID,
		UserType:      user.UserType,
		TenantID:      tenantID,
		DisplayName:   displayName(user),
		Permissions:   perms,
		EntityScope:   entityScope,
		Configuration: configuration,
	}

	// Populate cache with the full resolved session (including DisplayName).
	s.repo.CacheResolved(ctx, hash, resolved, s.cfg.SessionTTL)

	s.metrics.IncrementCounter("session.login.success", nil)
	return resolved, rawToken, nil
}

// ValidateSession checks the repo (which handles cache-aside internally).
func (s *sessionService) ValidateSession(ctx context.Context, token string) (*domain.ResolvedSession, error) {
	ctx, span := s.tracer.StartSpan(ctx, "session.ValidateSession")
	defer span.End()

	hash := sha256hex(token)
	resolved, err := s.repo.ValidateToken(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("session: validate token: %w", err)
	}
	if resolved == nil {
		return nil, errors.ErrAuthenticationFailed
	}
	return resolved, nil
}

// Logout invalidates the session via the repo (DB + cache eviction).
func (s *sessionService) Logout(ctx context.Context, token string) error {
	ctx, span := s.tracer.StartSpan(ctx, "session.Logout")
	defer span.End()

	hash := sha256hex(token)
	if err := s.repo.Invalidate(ctx, hash); err != nil {
		return fmt.Errorf("session: logout: %w", err)
	}
	return nil
}

//  Permission computation

func (s *sessionService) buildPermissions(ctx context.Context, user *domain.User) (map[string]bool, error) {
	subject := subjectForUser(user)
	domainName := domainForUser(user)

	// GetImplicitRoles traverses the full role inheritance chain, unlike
	// GetRoles which only returns directly-assigned roles.
	roles, err := s.authz.GetImplicitRoles(ctx, subject, domainName)
	if err != nil {
		return nil, fmt.Errorf("buildPermissions: GetImplicitRoles: %w", err)
	}

	policies, err := s.authz.GetPolicies(ctx, domainName)
	if err != nil {
		return nil, fmt.Errorf("buildPermissions: GetPolicies: %w", err)
	}

	roleSet := make(map[string]bool, len(roles))
	for _, r := range roles {
		roleSet[r] = true
	}

	// Separate allows and denies; deny-override: one deny beats all allows.
	allows := make(map[string]bool)
	denies := make(map[string]bool)
	for _, p := range policies {
		if !roleSet[p.Subject] {
			continue
		}
		key := p.Object + "." + p.Action
		switch p.Effect {
		case "allow":
			allows[key] = true
		case "deny":
			denies[key] = true
		}
	}

	perms := make(map[string]bool, len(allows))
	for key := range allows {
		if !denies[key] {
			perms[key] = true
		}
	}

	s.metrics.IncrementCounter("session.permissions_computed", nil)
	return perms, nil
}

//  Helpers

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

func subjectForUser(user *domain.User) string {
	id := user.ID.String()
	switch domain.ActorTypeFromUserType(user.UserType) {
	case domain.ActorPlatform:
		return domain.PlatformSubject(id)
	case domain.ActorPortal:
		return domain.PortalSubject(id)
	case domain.ActorAPI:
		return domain.APISubject(id)
	default:
		return domain.TenantSubject(id)
	}
}

func domainForUser(user *domain.User) string {
	tenantID := user.TenantID.String()
	switch domain.ActorTypeFromUserType(user.UserType) {
	case domain.ActorPlatform:
		return domain.DomainPlatform
	case domain.ActorPortal:
		return domain.PortalDomain(tenantID)
	case domain.ActorAPI:
		return domain.APIDomain(tenantID)
	default:
		return domain.TenantDomain(tenantID)
	}
}

func (s *sessionService) LogoutAllForUser(ctx context.Context, userID uuid.UUID) error {
	ctx, span := s.tracer.StartSpan(ctx, "session.LogoutAllForUser")
	defer span.End()
	return s.repo.InvalidateByUser(ctx, userID)
}

func (s *sessionService) LogoutAllForTenant(ctx context.Context, tenantID uuid.UUID) error {
	ctx, span := s.tracer.StartSpan(ctx, "session.LogoutAllForTenant")
	defer span.End()
	return s.repo.InvalidateByTenant(ctx, tenantID)
}

func displayName(user *domain.User) string {
	if user.DisplayName != nil && *user.DisplayName != "" {
		return *user.DisplayName
	}
	if user.Username != "" {
		return user.Username
	}
	return user.Email
}
