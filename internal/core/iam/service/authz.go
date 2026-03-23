package service

import (
	"context"
	"fmt"

	casbin "github.com/casbin/casbin/v2"
	casbinmodel "github.com/casbin/casbin/v2/model"
	"github.com/google/uuid"

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
	GetRoles(ctx context.Context, subject, domainName string) ([]string, error)
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
	Store   db.Store               // required
	Cache   cache.Service          // required (passed to repo)
	Logger  logger.Logger          // required
	Metrics metrics.MetricsProvider // optional
	Tracer  tracing.Service        // optional
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

	log := cfg.Logger.WithFields(logger.Fields{"component": "authz"})
	repo := repository.NewAuthzRepository(cfg.Store, cfg.Cache, log, cfg.Metrics, cfg.Tracer)

	return &authzService{enforcer: e, repo: repo, log: log, metrics: cfg.Metrics, tracer: cfg.Tracer}, nil
}

// NewInMemoryAuthzService creates an AuthzService backed by a pure in-memory
// Casbin enforcer with AutoSave disabled. Useful for unit tests that do not
// need a database — writes (AddPolicy, AssignRole) only affect in-memory state.
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
	return &authzService{enforcer: e, repo: repo, log: log}, nil
}

// ─── Enforcement ─────────────────────────────────────────────────────────────

func (s *authzService) Enforce(ctx context.Context, r domain.Request) (bool, error) {
	if r.Subject == "" || r.Domain == "" || r.Object == "" || r.Action == "" {
		return false, domain.ErrInvalidRequest
	}

	if err := s.revokeExpiredRoles(ctx, r.Subject, r.Domain); err != nil {
		s.log.WarnContext(ctx, "authz: revokeExpiredRoles failed", logger.Fields{
			"subject": r.Subject, "domain": r.Domain, "error": err.Error(),
		})
	}

	allowed, err := s.enforcer.Enforce(r.Subject, r.Domain, r.Object, r.Action)
	if err != nil {
		return false, fmt.Errorf("authz: enforce: %w", err)
	}

	s.log.DebugContext(ctx, "authz enforce", logger.Fields{
		"subject": r.Subject, "domain": r.Domain,
		"object": r.Object, "action": r.Action, "allowed": allowed,
	})
	return allowed, nil
}

func (s *authzService) EnforceBatch(ctx context.Context, reqs []domain.Request) ([]bool, error) {
	if len(reqs) == 0 {
		return nil, nil
	}
	batch := make([][]interface{}, len(reqs))
	for i, r := range reqs {
		if r.Subject == "" || r.Domain == "" || r.Object == "" || r.Action == "" {
			return nil, domain.ErrInvalidRequest
		}
		batch[i] = []interface{}{r.Subject, r.Domain, r.Object, r.Action}
	}
	results, err := s.enforcer.BatchEnforce(batch)
	if err != nil {
		return nil, fmt.Errorf("authz: batch enforce: %w", err)
	}
	return results, nil
}

// ─── Role management ─────────────────────────────────────────────────────────

func (s *authzService) AssignRole(ctx context.Context, tenantID, subject, role, domainName string, opts ...domain.AssignOpt) error {
	if subject == "" || role == "" || domainName == "" {
		return domain.ErrInvalidRequest
	}

	if _, err := s.enforcer.AddGroupingPolicy(subject, role, domainName); err != nil {
		return fmt.Errorf("authz: add grouping policy: %w", err)
	}

	tid, _ := uuid.Parse(tenantID)
	expiresAt, assignedBy, delegatedBy := domain.ApplyAssignOpts(opts)
	return s.repo.UpsertRoleAssignment(ctx, tid, subject, role, domainName, assignedBy, delegatedBy, expiresAt)
}

func (s *authzService) RevokeRole(ctx context.Context, subject, role, domainName string) error {
	if subject == "" || role == "" || domainName == "" {
		return domain.ErrInvalidRequest
	}
	s.enforcer.DeleteRoleForUserInDomain(subject, role, domainName)
	return s.repo.DeactivateRoleAssignment(ctx, subject, role, domainName)
}

func (s *authzService) GetRoles(_ context.Context, subject, domainName string) ([]string, error) {
	return s.enforcer.GetRolesForUserInDomain(subject, domainName), nil
}

func (s *authzService) HasRole(_ context.Context, subject, role, domainName string) (bool, error) {
	roles := s.enforcer.GetRolesForUserInDomain(subject, domainName)
	for _, r := range roles {
		if r == role {
			return true, nil
		}
	}
	return false, nil
}

func (s *authzService) GetAssignments(ctx context.Context, subject, domainName string) ([]domain.RoleAssignment, error) {
	return s.repo.ListRoleAssignments(ctx, subject, domainName)
}

// ─── Policy management ────────────────────────────────────────────────────────

func (s *authzService) AddPolicy(_ context.Context, p domain.Policy) error {
	added, err := s.enforcer.AddPolicy(p.Subject, p.Domain, p.Object, p.Action, p.Effect)
	if err != nil {
		return fmt.Errorf("authz: add policy: %w", err)
	}
	if !added {
		return domain.ErrPolicyConflict
	}
	return nil
}

func (s *authzService) RemovePolicy(_ context.Context, p domain.Policy) error {
	_, err := s.enforcer.RemovePolicy(p.Subject, p.Domain, p.Object, p.Action, p.Effect)
	return err
}

func (s *authzService) GetPolicies(_ context.Context, domainName string) ([]domain.Policy, error) {
	// GetFilteredPolicy(fieldIndex=1, fieldValue=domain) returns all p-rules for domain.
	rows, err := s.enforcer.GetFilteredPolicy(1, domainName)
	if err != nil {
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
	return out, nil
}

// ─── Cache ────────────────────────────────────────────────────────────────────

func (s *authzService) InvalidateCache(_ context.Context) error {
	if err := s.enforcer.LoadPolicy(); err != nil {
		return fmt.Errorf("authz: reload policy: %w", err)
	}
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
