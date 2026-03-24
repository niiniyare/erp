package service

import (
	"context"
	"fmt"

	casbin "github.com/casbin/casbin/v2"
	casbinmodel "github.com/casbin/casbin/v2/model"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"

	db "awo.so/db/sqlc"
	"awo.so/internal/core/iam/domain"
	"awo.so/internal/core/iam/repository"
	"awo.so/internal/platform/cache"
	sharedErrors "awo.so/internal/shared/errors"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
)

// ─── Port (interface) ─────────────────────────────────────────────────────────

// AuthzService is the application service for authorization operations.
// Note: HTTP middleware is NOT part of this interface — see api/middleware.
type AuthzService interface {
	// Enforcement
	Enforce(ctx context.Context, r domain.Request) (bool, error)
	EnforceBatch(ctx context.Context, reqs []domain.Request) ([]bool, error)

	// Role management
	AssignRole(ctx context.Context, tenantID, subject, role, domainName string, opts ...domain.AssignOpt) error
	RevokeRole(ctx context.Context, subject, role, domainName string) error
	// GetRoles returns directly-assigned roles only (no inheritance traversal).
	GetRoles(ctx context.Context, subject, domainName string) ([]string, error)
	// GetImplicitRoles returns all effective roles including those inherited
	// through the role hierarchy. Use this for permission pre-computation at login.
	GetImplicitRoles(ctx context.Context, subject, domainName string) ([]string, error)
	HasRole(ctx context.Context, subject, role, domainName string) (bool, error)
	GetAssignments(ctx context.Context, subject, domainName string) ([]domain.RoleAssignment, error)

	// Policy management (admin)
	AddPolicy(ctx context.Context, p domain.Policy) error
	RemovePolicy(ctx context.Context, p domain.Policy) error
	GetPolicies(ctx context.Context, domainName string) ([]domain.Policy, error)

	// Cache
	InvalidateCache(ctx context.Context) error
}

// ─── Config ───────────────────────────────────────────────────────────────────

// AuthzConfig holds the dependencies required to create an AuthzService.
type AuthzConfig struct {
	Store   db.Store                // required
	Cache   cache.Service           // required (passed to repo)
	Logger  logger.Logger           // required
	Metrics metrics.MetricsProvider // optional
	Tracer  tracing.Service         // optional
}

// ─── Implementation ───────────────────────────────────────────────────────────

type authzService struct {
	enforcer *casbin.Enforcer
	repo     repository.AuthzRepository
	log      logger.Logger
	metrics  metrics.MetricsProvider
	tracer   tracing.Service
}

// NewAuthzService creates a fully initialised AuthzService backed by PostgreSQL via Casbin.
func NewAuthzService(cfg AuthzConfig) (AuthzService, error) {
	if cfg.Store == nil {
		return nil, fmt.Errorf("authz: store is required")
	}
	if cfg.Cache == nil {
		return nil, fmt.Errorf("authz: cache is required")
	}
	if cfg.Logger == nil {
		return nil, fmt.Errorf("authz: logger is required")
	}

	m, err := casbinmodel.NewModelFromString(domain.CasbinModel)
	if err != nil {
		return nil, fmt.Errorf("authz: build casbin model: %w", err)
	}

	adapter := repository.NewPgxAdapter(cfg.Store.GetPool())
	e, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		return nil, fmt.Errorf("authz: create casbin enforcer: %w", err)
	}
	e.EnableAutoSave(true)

	log := cfg.Logger.WithFields(logger.Fields{"component": "iam.authz"})
	repo := repository.NewAuthzRepository(cfg.Store, cfg.Cache, log, cfg.Metrics, cfg.Tracer)

	return &authzService{enforcer: e, repo: repo, log: log, metrics: cfg.Metrics, tracer: cfg.Tracer}, nil
}

// NewInMemoryAuthzService creates an AuthzService backed by a pure in-memory
// Casbin enforcer with AutoSave disabled. Useful for unit tests that do not
// need a database — writes (AddPolicy, AssignRole) only affect in-memory state.
// No-op tracer and metrics are wired in so all methods are safe to call without
// a real observability stack.
func NewInMemoryAuthzService(repo repository.AuthzRepository, log logger.Logger) (AuthzService, error) {
	m, err := casbinmodel.NewModelFromString(domain.CasbinModel)
	if err != nil {
		return nil, fmt.Errorf("authz: build casbin model: %w", err)
	}
	e, err := casbin.NewEnforcer(m)
	if err != nil {
		return nil, fmt.Errorf("authz: create in-memory enforcer: %w", err)
	}
	e.EnableAutoSave(false)
	scopedLog := log.WithFields(logger.Fields{"component": "iam.authz"})
	return &authzService{
		enforcer: e,
		repo:     repo,
		log:      scopedLog,
		metrics:  metrics.NewNoOpMetricsProvider(),
		tracer:   tracing.NewNoOpService(),
	}, nil
}

// ─── Enforcement ──────────────────────────────────────────────────────────────

func (s *authzService) Enforce(ctx context.Context, r domain.Request) (bool, error) {
	ctx, span := s.tracer.StartSpan(ctx, "iam.authz.Enforce")
	defer span.End()

	timer := s.metrics.Timer("iam_authz_enforce_duration", nil)
	defer timer.Stop()

	if r.Subject == "" || r.Domain == "" || r.Object == "" || r.Action == "" {
		return false, domain.ErrInvalidRequest
	}

	span.SetAttributes(
		attribute.String("authz.subject", r.Subject),
		attribute.String("authz.domain", r.Domain),
		attribute.String("authz.object", r.Object),
		attribute.String("authz.action", r.Action),
	)

	if err := s.revokeExpiredRoles(ctx, r.Subject, r.Domain); err != nil {
		s.log.WarnContext(ctx, "authz: revokeExpiredRoles failed", logger.Fields{
			"subject": r.Subject, "domain": r.Domain, "error": err.Error(),
		})
	}

	allowed, err := s.enforcer.Enforce(r.Subject, r.Domain, r.Object, r.Action)
	if err != nil {
		span.RecordError(err)
		s.log.ErrorContext(ctx, "authz enforce failed", logger.Fields{
			"subject":  r.Subject,
			"domain":   r.Domain,
			"object":   r.Object,
			"action":   r.Action,
			"error":    err.Error(),
			"trace_id": s.tracer.GetTraceID(ctx),
		})
		return false, fmt.Errorf("authz: enforce: %w", err)
	}

	span.SetAttributes(attribute.Bool("authz.allowed", allowed))
	s.metrics.IncrementCounter("iam.authz.enforce", metrics.Fields{"allowed": fmt.Sprintf("%t", allowed)})
	s.log.DebugContext(ctx, "authz enforce", logger.Fields{
		"subject": r.Subject, "domain": r.Domain,
		"object": r.Object, "action": r.Action, "allowed": allowed,
	})
	return allowed, nil
}

func (s *authzService) EnforceBatch(ctx context.Context, reqs []domain.Request) ([]bool, error) {
	ctx, span := s.tracer.StartSpan(ctx, "iam.authz.EnforceBatch")
	defer span.End()

	timer := s.metrics.Timer("iam_authz_enforce_batch_duration", nil)
	defer timer.Stop()

	if len(reqs) == 0 {
		return nil, nil
	}

	span.SetAttributes(attribute.Int("authz.batch_size", len(reqs)))

	batch := make([][]interface{}, len(reqs))
	for i, r := range reqs {
		if r.Subject == "" || r.Domain == "" || r.Object == "" || r.Action == "" {
			return nil, domain.ErrInvalidRequest
		}
		batch[i] = []interface{}{r.Subject, r.Domain, r.Object, r.Action}
	}
	results, err := s.enforcer.BatchEnforce(batch)
	if err != nil {
		span.RecordError(err)
		s.log.ErrorContext(ctx, "authz batch enforce failed", logger.Fields{
			"batch_size": len(reqs),
			"error":      err.Error(),
			"trace_id":   s.tracer.GetTraceID(ctx),
		})
		return nil, fmt.Errorf("authz: batch enforce: %w", err)
	}

	s.metrics.IncrementCounter("iam.authz.enforce_batch", nil)
	return results, nil
}

// ─── Role management ──────────────────────────────────────────────────────────

func (s *authzService) AssignRole(ctx context.Context, tenantID, subject, role, domainName string, opts ...domain.AssignOpt) error {
	ctx, span := s.tracer.StartSpan(ctx, "iam.authz.AssignRole")
	defer span.End()

	timer := s.metrics.Timer("iam_authz_role_assign_duration", nil)
	defer timer.Stop()

	if subject == "" || role == "" || domainName == "" {
		return domain.ErrInvalidRequest
	}

	span.SetAttributes(
		attribute.String("authz.subject", subject),
		attribute.String("authz.role", role),
		attribute.String("authz.domain", domainName),
	)

	if _, err := s.enforcer.AddGroupingPolicy(subject, role, domainName); err != nil {
		span.RecordError(err)
		s.log.ErrorContext(ctx, "authz assign role: add grouping policy failed", logger.Fields{
			"subject":  subject,
			"role":     role,
			"domain":   domainName,
			"error":    err.Error(),
			"trace_id": s.tracer.GetTraceID(ctx),
		})
		return fmt.Errorf("authz: add grouping policy: %w", err)
	}

	tid, _ := uuid.Parse(tenantID)
	ao := domain.ApplyAssignOpts(opts)

	// repo expects *string for nullable audit columns;
	// empty string means not provided → pass nil.
	var assignedBy, delegatedBy *string
	if ao.AssignedBy != "" {
		assignedBy = &ao.AssignedBy
	}
	if ao.DelegatedBy != "" {
		delegatedBy = &ao.DelegatedBy
	}

	if err := s.repo.UpsertRoleAssignment(ctx, tid, subject, role, domainName, assignedBy, delegatedBy, ao.ExpiresAt); err != nil {
		span.RecordError(err)
		s.log.ErrorContext(ctx, "authz assign role: upsert failed", logger.Fields{
			"subject":  subject,
			"role":     role,
			"domain":   domainName,
			"error":    err.Error(),
			"trace_id": s.tracer.GetTraceID(ctx),
		})
		return err
	}

	s.metrics.IncrementCounter("iam.authz.role.assigned", nil)
	s.log.DebugContext(ctx, "role assigned", logger.Fields{
		"subject": subject, "role": role, "domain": domainName,
	})
	return nil
}

func (s *authzService) RevokeRole(ctx context.Context, subject, role, domainName string) error {
	ctx, span := s.tracer.StartSpan(ctx, "iam.authz.RevokeRole")
	defer span.End()

	timer := s.metrics.Timer("iam_authz_role_revoke_duration", nil)
	defer timer.Stop()

	if subject == "" || role == "" || domainName == "" {
		return domain.ErrInvalidRequest
	}

	span.SetAttributes(
		attribute.String("authz.subject", subject),
		attribute.String("authz.role", role),
		attribute.String("authz.domain", domainName),
	)

	s.enforcer.DeleteRoleForUserInDomain(subject, role, domainName)

	if err := s.repo.DeactivateRoleAssignment(ctx, subject, role, domainName); err != nil {
		span.RecordError(err)
		s.log.ErrorContext(ctx, "authz revoke role: deactivate failed", logger.Fields{
			"subject":  subject,
			"role":     role,
			"domain":   domainName,
			"error":    err.Error(),
			"trace_id": s.tracer.GetTraceID(ctx),
		})
		return err
	}

	s.metrics.IncrementCounter("iam.authz.role.revoked", nil)
	s.log.DebugContext(ctx, "role revoked", logger.Fields{
		"subject": subject, "role": role, "domain": domainName,
	})
	return nil
}

func (s *authzService) GetRoles(ctx context.Context, subject, domainName string) ([]string, error) {
	ctx, span := s.tracer.StartSpan(ctx, "iam.authz.GetRoles")
	defer span.End()

	span.SetAttributes(
		attribute.String("authz.subject", subject),
		attribute.String("authz.domain", domainName),
	)

	roles := s.enforcer.GetRolesForUserInDomain(subject, domainName)
	span.SetAttributes(attribute.Int("authz.roles_count", len(roles)))
	return roles, nil
}

func (s *authzService) GetImplicitRoles(ctx context.Context, subject, domainName string) ([]string, error) {
	ctx, span := s.tracer.StartSpan(ctx, "iam.authz.GetImplicitRoles")
	defer span.End()

	span.SetAttributes(
		attribute.String("authz.subject", subject),
		attribute.String("authz.domain", domainName),
	)

	roles, err := s.enforcer.GetImplicitRolesForUser(subject, domainName)
	if err != nil {
		span.RecordError(err)
		s.log.ErrorContext(ctx, "authz get implicit roles failed", logger.Fields{
			"subject":  subject,
			"domain":   domainName,
			"error":    err.Error(),
			"trace_id": s.tracer.GetTraceID(ctx),
		})
		return nil, fmt.Errorf("authz: get implicit roles: %w", err)
	}

	span.SetAttributes(attribute.Int("authz.roles_count", len(roles)))
	return roles, nil
}

func (s *authzService) HasRole(ctx context.Context, subject, role, domainName string) (bool, error) {
	ctx, span := s.tracer.StartSpan(ctx, "iam.authz.HasRole")
	defer span.End()

	span.SetAttributes(
		attribute.String("authz.subject", subject),
		attribute.String("authz.role", role),
		attribute.String("authz.domain", domainName),
	)

	for _, r := range s.enforcer.GetRolesForUserInDomain(subject, domainName) {
		if r == role {
			span.SetAttributes(attribute.Bool("authz.has_role", true))
			return true, nil
		}
	}
	span.SetAttributes(attribute.Bool("authz.has_role", false))
	return false, nil
}

func (s *authzService) GetAssignments(ctx context.Context, subject, domainName string) ([]domain.RoleAssignment, error) {
	ctx, span := s.tracer.StartSpan(ctx, "iam.authz.GetAssignments")
	defer span.End()

	timer := s.metrics.Timer("iam_authz_get_assignments_duration", nil)
	defer timer.Stop()

	span.SetAttributes(
		attribute.String("authz.subject", subject),
		attribute.String("authz.domain", domainName),
	)

	assignments, err := s.repo.ListRoleAssignments(ctx, subject, domainName)
	if err != nil {
		span.RecordError(err)
		s.log.ErrorContext(ctx, "authz get assignments failed", logger.Fields{
			"subject":  subject,
			"domain":   domainName,
			"error":    err.Error(),
			"trace_id": s.tracer.GetTraceID(ctx),
		})
		return nil, err
	}

	span.SetAttributes(attribute.Int("authz.assignments_count", len(assignments)))
	return assignments, nil
}

// ─── Policy management ────────────────────────────────────────────────────────

func (s *authzService) AddPolicy(ctx context.Context, p domain.Policy) error {
	if p.Subject == "" || p.Domain == "" || p.Object == "" || p.Action == "" {
		return domain.ErrInvalidRequest
	}
	if p.Effect != "allow" && p.Effect != "deny" {
		return fmt.Errorf("authz: invalid policy effect %q: must be \"allow\" or \"deny\"", p.Effect)
	}

	ctx, span := s.tracer.StartSpan(ctx, "iam.authz.AddPolicy")
	defer span.End()

	span.SetAttributes(
		attribute.String("authz.subject", p.Subject),
		attribute.String("authz.domain", p.Domain),
		attribute.String("authz.object", p.Object),
		attribute.String("authz.action", p.Action),
		attribute.String("authz.effect", p.Effect),
	)

	added, err := s.enforcer.AddPolicy(p.Subject, p.Domain, p.Object, p.Action, p.Effect)
	if err != nil {
		span.RecordError(err)
		s.log.ErrorContext(ctx, "authz add policy failed", logger.Fields{
			"subject":  p.Subject,
			"domain":   p.Domain,
			"object":   p.Object,
			"action":   p.Action,
			"error":    err.Error(),
			"trace_id": s.tracer.GetTraceID(ctx),
		})
		return fmt.Errorf("authz: add policy: %w", err)
	}
	if !added {
		return domain.ErrPolicyConflict
	}

	s.metrics.IncrementCounter("iam.authz.policy.added", nil)
	s.log.DebugContext(ctx, "policy added", logger.Fields{
		"subject": p.Subject, "domain": p.Domain, "object": p.Object, "action": p.Action, "effect": p.Effect,
	})
	return nil
}

func (s *authzService) RemovePolicy(ctx context.Context, p domain.Policy) error {
	if p.Subject == "" || p.Domain == "" || p.Object == "" || p.Action == "" {
		return domain.ErrInvalidRequest
	}

	ctx, span := s.tracer.StartSpan(ctx, "iam.authz.RemovePolicy")
	defer span.End()

	span.SetAttributes(
		attribute.String("authz.subject", p.Subject),
		attribute.String("authz.domain", p.Domain),
		attribute.String("authz.object", p.Object),
		attribute.String("authz.action", p.Action),
	)

	_, err := s.enforcer.RemovePolicy(p.Subject, p.Domain, p.Object, p.Action, p.Effect)
	if err != nil {
		span.RecordError(err)
		s.log.ErrorContext(ctx, "authz remove policy failed", logger.Fields{
			"subject":  p.Subject,
			"domain":   p.Domain,
			"object":   p.Object,
			"action":   p.Action,
			"error":    err.Error(),
			"trace_id": s.tracer.GetTraceID(ctx),
		})
		return err
	}

	s.metrics.IncrementCounter("iam.authz.policy.removed", nil)
	return nil
}

func (s *authzService) GetPolicies(ctx context.Context, domainName string) ([]domain.Policy, error) {
	ctx, span := s.tracer.StartSpan(ctx, "iam.authz.GetPolicies")
	defer span.End()

	timer := s.metrics.Timer("iam_authz_get_policies_duration", nil)
	defer timer.Stop()

	span.SetAttributes(attribute.String("authz.domain", domainName))

	// GetFilteredPolicy(fieldIndex=1, fieldValue=domain) returns all p-rules for domain.
	rows, err := s.enforcer.GetFilteredPolicy(1, domainName)
	if err != nil {
		span.RecordError(err)
		s.log.ErrorContext(ctx, "authz get policies failed", logger.Fields{
			"domain":   domainName,
			"error":    err.Error(),
			"trace_id": s.tracer.GetTraceID(ctx),
		})
		return nil, sharedErrors.NewBusinessError("GET_POLICIES_FAILED", "failed to retrieve policies").
			WithCategory(sharedErrors.CategorySecurity).
			WithDetail("domain", domainName).
			WithCause(err)
	}
	out := make([]domain.Policy, 0, len(rows))
	for _, row := range rows {
		if len(row) < 5 {
			continue
		}
		out = append(out, domain.Policy{
			Subject: row[0],
			Domain:  row[1],
			Object:  row[2],
			Action:  row[3],
			Effect:  row[4],
		})
	}

	span.SetAttributes(attribute.Int("authz.policies_count", len(out)))
	return out, nil
}

// ─── Cache ────────────────────────────────────────────────────────────────────

func (s *authzService) InvalidateCache(ctx context.Context) error {
	ctx, span := s.tracer.StartSpan(ctx, "iam.authz.InvalidateCache")
	defer span.End()

	if err := s.enforcer.LoadPolicy(); err != nil {
		span.RecordError(err)
		s.log.ErrorContext(ctx, "authz reload policy failed", logger.Fields{
			"error":    err.Error(),
			"trace_id": s.tracer.GetTraceID(ctx),
		})
		return fmt.Errorf("authz: reload policy: %w", err)
	}

	s.metrics.IncrementCounter("iam.authz.cache.invalidated", nil)
	s.log.DebugContext(ctx, "authz policy cache reloaded", nil)
	return nil
}

// ─── Internal helpers ─────────────────────────────────────────────────────────

// revokeExpiredRoles lazily removes roles whose expiry has passed.
// Non-fatal: enforce proceeds even if cleanup fails.
func (s *authzService) revokeExpiredRoles(ctx context.Context, subject, domainName string) error {
	names, err := s.repo.ListExpiredActiveRoleNames(ctx, subject, domainName)
	if err != nil {
		return err
	}
	for _, name := range names {
		s.enforcer.DeleteRoleForUserInDomain(subject, name, domainName)
		if err := s.repo.DeactivateRoleAssignment(ctx, subject, name, domainName); err != nil {
			s.log.WarnContext(ctx, "authz: deactivate expired role failed", logger.Fields{
				"subject": subject, "role": name, "domain": domainName, "error": err.Error(),
			})
		}
	}
	return nil
}
