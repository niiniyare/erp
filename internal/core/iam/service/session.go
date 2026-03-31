package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"

	"awo.so/internal/core/iam/domain"
	"awo.so/internal/core/iam/repository"
	"awo.so/internal/platform/cache"
	"awo.so/internal/shared/errors"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
)

// Port (interface)

// SessionService handles Login / ValidateSession / Logout.
type SessionService interface {
	// Login authenticates the user.  When MFA is enabled for the user, the
	// returned error is ErrMFARequired and the token string is a short-lived
	// pending token.  The caller must redirect to CompleteMFALogin.
	Login(ctx context.Context, email, password string) (*domain.ResolvedSession, string, error)

	// CompleteMFALogin exchanges a pending MFA token + TOTP code for a full
	// session.  Call this after Login returns ErrMFARequired.
	CompleteMFALogin(ctx context.Context, pendingToken, mfaCode string) (*domain.ResolvedSession, string, error)

	ValidateSession(ctx context.Context, token string) (*domain.ResolvedSession, error)
	Logout(ctx context.Context, token string) error

	// LogoutAllForUser invalidates all active DB sessions for the given user.
	// Existing cache entries expire naturally within the session TTL window.
	LogoutAllForUser(ctx context.Context, userID uuid.UUID) error

	// LogoutAllForTenant invalidates all active DB sessions for the given tenant.
	// Existing cache entries expire naturally within the session TTL window.
	LogoutAllForTenant(ctx context.Context, tenantID uuid.UUID) error

	// LoginWithSSO creates a full session for a user that was authenticated via
	// an external OAuth/OIDC provider.  The caller (SSOService) is responsible
	// for verifying the OAuth exchange before calling this.
	// MFA is intentionally skipped — the IdP is the second factor.
	LoginWithSSO(ctx context.Context, user *domain.User) (*domain.ResolvedSession, string, error)
}

// Implementation

type sessionService struct {
	identity UserService
	authz    AuthzService
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
	scopedLog := log.WithFields(logger.Fields{"component": "iam.session"})
	return &sessionService{
		identity: identitySvc,
		authz:    authzSvc,
		repo:     repo,
		tracer:   tracer,
		metrics:  m,
		log:      scopedLog,
		cfg:      cfg,
	}
}

// Login authenticates the user and either creates a full session or,
// when MFA is enabled, returns (nil, pendingToken, ErrMFARequired).
// In the MFA case, the caller must call CompleteMFALogin with the pending token
// and the user's TOTP code.
// 
// NOTE(tenant-context): ctx must carry tenant_id via cache.TenantIDKey.
func (s *sessionService) Login(ctx context.Context, email, password string) (*domain.ResolvedSession, string, error) {
	ctx, span := s.tracer.StartSpan(ctx, "iam.session.Login")
	defer span.End()

	timer := s.metrics.Timer("iam_session_login_duration", nil)
	defer timer.Stop()

	user, err := s.identity.Authenticate(ctx, email, password)
	if err != nil {
		s.metrics.IncrementCounter("iam.session.login.failure", nil)
		return nil, "", err
	}

	span.SetAttributes(
		attribute.String("user.id", user.ID.String()),
		attribute.String("user.type", user.UserType),
	)

	// MFA step 1: if enabled, return a short-lived pending token.
	if user.MfaEnabled {
		rawPending, _, err := generateToken()
		if err != nil {
			span.RecordError(err)
			return nil, "", fmt.Errorf("iam session: generate mfa pending token: %w", err)
		}
		if err := s.repo.StorePendingMFA(ctx, rawPending, user.ID); err != nil {
			span.RecordError(err)
			return nil, "", fmt.Errorf("iam session: store mfa pending: %w", err)
		}
		s.metrics.IncrementCounter("iam.session.mfa.pending", nil)
		return nil, rawPending, errors.ErrMFARequired
	}

	return s.buildAndPersistSession(ctx, user)
}

// CompleteMFALogin validates a TOTP code against a pending login token and,
// on success, creates and returns a full session.
func (s *sessionService) CompleteMFALogin(ctx context.Context, pendingToken, mfaCode string) (*domain.ResolvedSession, string, error) {
	ctx, span := s.tracer.StartSpan(ctx, "iam.session.CompleteMFALogin")
	defer span.End()

	userID, err := s.repo.GetPendingMFA(ctx, pendingToken)
	if err != nil {
		return nil, "", errors.ErrAuthenticationFailed
	}

	ok, err := s.identity.ValidateMFACode(ctx, userID, mfaCode)
	if err != nil {
		span.RecordError(err)
		return nil, "", fmt.Errorf("iam session: validate mfa code: %w", err)
	}
	if !ok {
		s.metrics.IncrementCounter("iam.session.mfa.invalid", nil)
		return nil, "", errors.ErrMFAInvalid
	}

	// Pending token is single-use — delete it regardless of subsequent errors.
	_ = s.repo.DeletePendingMFA(ctx, pendingToken)

	user, err := s.identity.GetUserByID(ctx, userID)
	if err != nil {
		span.RecordError(err)
		return nil, "", fmt.Errorf("iam session: get user after mfa: %w", err)
	}

	s.metrics.IncrementCounter("iam.session.mfa.success", nil)
	return s.buildAndPersistSession(ctx, user)
}

// LoginWithSSO creates a full session for an SSO-authenticated user.
// MFA is intentionally skipped — the identity provider is the second factor.
func (s *sessionService) LoginWithSSO(ctx context.Context, user *domain.User) (*domain.ResolvedSession, string, error) {
	ctx, span := s.tracer.StartSpan(ctx, "iam.session.LoginWithSSO")
	defer span.End()
	span.SetAttributes(attribute.String("user.id", user.ID.String()))

	s.metrics.IncrementCounter("iam.session.sso.login", nil)
	return s.buildAndPersistSession(ctx, user)
}

// buildAndPersistSession creates the ResolvedSession and persists it.
// Called by Login (non-MFA path), CompleteMFALogin, and LoginWithSSO.
func (s *sessionService) buildAndPersistSession(ctx context.Context, user *domain.User) (*domain.ResolvedSession, string, error) {
	ctx, span := s.tracer.StartSpan(ctx, "iam.session.buildAndPersistSession")
	defer span.End()

	perms, err := s.buildPermissions(ctx, user)
	if err != nil {
		s.log.WarnContext(ctx, "buildPermissions failed, proceeding with empty map", logger.Fields{
			"user_id":  user.ID.String(),
			"error":    err.Error(),
			"trace_id": s.tracer.GetTraceID(ctx),
		})
		perms = make(map[string]bool)
	}

	rawToken, hash, err := generateToken()
	if err != nil {
		span.RecordError(err)
		s.log.ErrorContext(ctx, "failed to generate session token", logger.Fields{
			"user_id":  user.ID.String(),
			"error":    err.Error(),
			"trace_id": s.tracer.GetTraceID(ctx),
		})
		return nil, "", fmt.Errorf("iam session: generate token: %w", err)
	}

	tenantID := tenantIDFromCtx(ctx)

	entityScope, err := s.repo.ResolveEntityScope(ctx, user.EntityID)
	if err != nil {
		s.log.WarnContext(ctx, "failed to resolve entity scope, using entity fallback", logger.Fields{
			"user_id":   user.ID.String(),
			"entity_id": user.EntityID.String(),
			"error":     err.Error(),
			"trace_id":  s.tracer.GetTraceID(ctx),
		})
		entityScope = domain.EntityScope{Type: domain.EntityScopeEntity, EntityID: user.EntityID.String()}
	}

	configuration, err := s.repo.LoadLoginConfig(ctx, user.ID, tenantID)
	if err != nil {
		s.log.WarnContext(ctx, "failed to load login config, using defaults", logger.Fields{
			"user_id":  user.ID.String(),
			"error":    err.Error(),
			"trace_id": s.tracer.GetTraceID(ctx),
		})
		configuration = domain.DefaultConfiguration()
	}

	ttl := s.resolveTTL(configuration)
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
		ExpiresAt:     now.Add(ttl),
		LastSeenAt:    now,
	}

	if err := s.repo.Create(ctx, sess); err != nil {
		span.RecordError(err)
		s.log.ErrorContext(ctx, "failed to persist session", logger.Fields{
			"user_id":  user.ID.String(),
			"error":    err.Error(),
			"trace_id": s.tracer.GetTraceID(ctx),
		})
		return nil, "", fmt.Errorf("iam session: persist session: %w", err)
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

	s.repo.CacheResolved(ctx, hash, resolved, ttl)

	s.metrics.IncrementCounter("iam.session.login.success", nil)
	s.log.DebugContext(ctx, "session created", logger.Fields{
		"user_id":   user.ID.String(),
		"user_type": user.UserType,
		"ttl":       ttl.String(),
	})
	return resolved, rawToken, nil
}

// ValidateSession checks the repo (which handles cache-aside internally).
func (s *sessionService) ValidateSession(ctx context.Context, token string) (*domain.ResolvedSession, error) {
	ctx, span := s.tracer.StartSpan(ctx, "iam.session.ValidateSession")
	defer span.End()

	timer := s.metrics.Timer("iam_session_validate_duration", nil)
	defer timer.Stop()

	hash := sha256hex(token)
	resolved, err := s.repo.ValidateToken(ctx, hash)
	if err != nil {
		span.RecordError(err)
		s.log.ErrorContext(ctx, "session validation error", logger.Fields{
			"error":    err.Error(),
			"trace_id": s.tracer.GetTraceID(ctx),
		})
		return nil, fmt.Errorf("iam session: validate token: %w", err)
	}
	if resolved == nil {
		s.metrics.IncrementCounter("iam.session.validate.miss", nil)
		return nil, errors.ErrAuthenticationFailed
	}

	s.metrics.IncrementCounter("iam.session.validate.hit", nil)
	return resolved, nil
}

// Logout invalidates the session via the repo (DB + cache eviction).
func (s *sessionService) Logout(ctx context.Context, token string) error {
	ctx, span := s.tracer.StartSpan(ctx, "iam.session.Logout")
	defer span.End()

	timer := s.metrics.Timer("iam_session_logout_duration", nil)
	defer timer.Stop()

	hash := sha256hex(token)
	if err := s.repo.Invalidate(ctx, hash); err != nil {
		span.RecordError(err)
		s.log.ErrorContext(ctx, "failed to invalidate session", logger.Fields{
			"error":    err.Error(),
			"trace_id": s.tracer.GetTraceID(ctx),
		})
		return fmt.Errorf("iam session: logout: %w", err)
	}

	s.metrics.IncrementCounter("iam.session.logout", nil)
	return nil
}

func (s *sessionService) LogoutAllForUser(ctx context.Context, userID uuid.UUID) error {
	ctx, span := s.tracer.StartSpan(ctx, "iam.session.LogoutAllForUser")
	defer span.End()
	span.SetAttributes(attribute.String("user.id", userID.String()))

	timer := s.metrics.Timer("iam_session_logout_all_duration", metrics.Fields{"scope": "user"})
	defer timer.Stop()

	if err := s.repo.InvalidateByUser(ctx, userID); err != nil {
		span.RecordError(err)
		s.log.ErrorContext(ctx, "failed to invalidate user sessions", logger.Fields{
			"user_id":  userID.String(),
			"error":    err.Error(),
			"trace_id": s.tracer.GetTraceID(ctx),
		})
		return err
	}

	s.log.DebugContext(ctx, "all user sessions invalidated", logger.Fields{"user_id": userID.String()})
	s.metrics.IncrementCounter("iam.session.logout.bulk", metrics.Fields{"scope": "user"})
	return nil
}

func (s *sessionService) LogoutAllForTenant(ctx context.Context, tenantID uuid.UUID) error {
	ctx, span := s.tracer.StartSpan(ctx, "iam.session.LogoutAllForTenant")
	defer span.End()
	span.SetAttributes(attribute.String("tenant.id", tenantID.String()))

	timer := s.metrics.Timer("iam_session_logout_all_duration", metrics.Fields{"scope": "tenant"})
	defer timer.Stop()

	if err := s.repo.InvalidateByTenant(ctx, tenantID); err != nil {
		span.RecordError(err)
		s.log.ErrorContext(ctx, "failed to invalidate tenant sessions", logger.Fields{
			"tenant_id": tenantID.String(),
			"error":     err.Error(),
			"trace_id":  s.tracer.GetTraceID(ctx),
		})
		return err
	}

	s.log.DebugContext(ctx, "all tenant sessions invalidated", logger.Fields{"tenant_id": tenantID.String()})
	s.metrics.IncrementCounter("iam.session.logout.bulk", metrics.Fields{"scope": "tenant"})
	return nil
}

// Permission computation

func (s *sessionService) buildPermissions(ctx context.Context, user *domain.User) (map[string]bool, error) {
	ctx, span := s.tracer.StartSpan(ctx, "iam.session.buildPermissions")
	defer span.End()

	subject := subjectForUser(user)
	domainName := domainForUser(user)

	span.SetAttributes(
		attribute.String("authz.subject", subject),
		attribute.String("authz.domain", domainName),
	)

	// GetImplicitRoles traverses the full role inheritance chain, unlike
	// GetRoles which only returns directly-assigned roles.
	roles, err := s.authz.GetImplicitRoles(ctx, subject, domainName)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("buildPermissions: GetImplicitRoles: %w", err)
	}

	// Just-in-time bootstrap: if a tenant user has no roles yet (e.g. they
	// were created before the bootstrap logic existed), seed the default
	// tenant_admin role now so the first login isn't permission-less.
	if len(roles) == 0 && domain.ActorTypeFromUserType(user.UserType) == domain.ActorTenant && user.TenantID != uuid.Nil {
		if bErr := s.authz.BootstrapTenantAdmin(ctx, user.TenantID, user.ID); bErr != nil {
			s.log.WarnContext(ctx, "jit bootstrap tenant_admin failed", logger.Fields{
				"user_id": user.ID.String(), "error": bErr.Error(),
			})
		} else {
			// Reload roles after bootstrap.
			roles, _ = s.authz.GetImplicitRoles(ctx, subject, domainName)
		}
	}

	policies, err := s.authz.GetPolicies(ctx, domainName)
	if err != nil {
		span.RecordError(err)
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

	// Collapse wildcard: a "*.*" policy (Object="*", Action="*") means the
	// role has blanket access. Replace it with the special "*" sentinel that
	// ResolvedSession.Can() recognises so every permission check short-circuits.
	if perms["*.*"] && !denies["*.*"] {
		delete(perms, "*.*")
		perms["*"] = true
	}

	span.SetAttributes(
		attribute.Int("authz.roles_count", len(roles)),
		attribute.Int("authz.permissions_count", len(perms)),
	)
	s.metrics.IncrementCounter("iam.session.permissions_computed", nil)
	return perms, nil
}

// Helpers

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

// resolveTTL returns the session TTL to use for this login.
// It reads "iam.session_ttl_hours" from the pre-computed tenant settings
// (already loaded into cfg.Settings by LoadLoginConfig) and converts it to a
// Duration. Falls back to the process-level default when the setting is absent,
// zero, or unparseable.
func (s *sessionService) resolveTTL(cfg domain.Configuration) time.Duration {
	if v, ok := cfg.Settings["iam.session_ttl_hours"]; ok && v != "" {
		if hours, err := strconv.Atoi(v); err == nil && hours > 0 {
			return time.Duration(hours) * time.Hour
		}
	}
	return s.cfg.SessionTTL
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
