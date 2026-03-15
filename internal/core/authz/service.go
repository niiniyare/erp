package authz

import (
	"context"
	"fmt"

	casbin "github.com/casbin/casbin/v2"
	casbinmodel "github.com/casbin/casbin/v2/model"

	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// Config holds the dependencies required to create a Service.
type Config struct {
	Store   db.Store           // required — wraps pgxpool for all DB operations
	Cache   cache.Service      // required — used by repo for role assignment caching
	Logger  logger.Logger      // required
	Metrics metrics.MetricsProvider // optional
	Tracer  tracing.Service    // optional
}

type service struct {
	enforcer *casbin.Enforcer
	repo     Repository
	log      logger.Logger
	metrics  metrics.MetricsProvider
	tracer   tracing.Service
}

// New creates a fully initialised Service backed by PostgreSQL via Casbin.
func New(cfg Config) (Service, error) {
	if cfg.Store == nil {
		return nil, fmt.Errorf("authz.New: store is required")
	}
	if cfg.Cache == nil {
		return nil, fmt.Errorf("authz.New: cache is required")
	}
	if cfg.Logger == nil {
		return nil, fmt.Errorf("authz.New: logger is required")
	}

	m, err := casbinmodel.NewModelFromString(casbinModel)
	if err != nil {
		return nil, fmt.Errorf("authz.New: build model: %w", err)
	}

	// Casbin pgx adapter uses the raw pool extracted from the Store.
	adapter := newPgxAdapter(cfg.Store.GetPool())
	e, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		return nil, fmt.Errorf("authz.New: create enforcer: %w", err)
	}
	e.EnableAutoSave(true)

	log := cfg.Logger.WithFields(logger.Fields{"component": "authz"})

	return &service{
		enforcer: e,
		repo:     newPgRepo(cfg.Store, cfg.Cache, log, cfg.Metrics, cfg.Tracer),
		log:      log,
		metrics:  cfg.Metrics,
		tracer:   cfg.Tracer,
	}, nil
}

// Enforce evaluates a single authorization request.
func (s *service) Enforce(ctx context.Context, r Request) (bool, error) {
	if r.Subject == "" || r.Domain == "" || r.Object == "" || r.Action == "" {
		return false, ErrInvalidRequest
	}

	// Lazy expiry: revoke any roles whose time has come before checking.
	if err := s.revokeExpiredRoles(ctx, r.Subject, r.Domain); err != nil {
		s.log.WarnContext(ctx, "authz: revokeExpiredRoles failed", logger.Fields{
			"subject": r.Subject,
			"domain":  r.Domain,
			"error":   err.Error(),
		})
		// Non-fatal: proceed with the check even if cleanup failed.
	}

	allowed, err := s.enforcer.Enforce(r.Subject, r.Domain, r.Object, r.Action)
	if err != nil {
		return false, fmt.Errorf("authz enforce: %w", err)
	}

	s.log.DebugContext(ctx, "authz enforce", logger.Fields{
		"subject": r.Subject,
		"domain":  r.Domain,
		"object":  r.Object,
		"action":  r.Action,
		"allowed": allowed,
	})

	return allowed, nil
}

// EnforceBatch evaluates multiple requests in a single call.
func (s *service) EnforceBatch(ctx context.Context, reqs []Request) ([]bool, error) {
	if len(reqs) == 0 {
		return nil, nil
	}

	batch := make([][]interface{}, len(reqs))
	for i, r := range reqs {
		if r.Subject == "" || r.Domain == "" || r.Object == "" || r.Action == "" {
			return nil, ErrInvalidRequest
		}
		batch[i] = []interface{}{r.Subject, r.Domain, r.Object, r.Action}
	}

	results, err := s.enforcer.BatchEnforce(batch)
	if err != nil {
		return nil, fmt.Errorf("authz batch enforce: %w", err)
	}
	return results, nil
}

// InvalidateCache reloads the policy from the database.
func (s *service) InvalidateCache(ctx context.Context) error {
	if err := s.enforcer.LoadPolicy(); err != nil {
		return fmt.Errorf("authz InvalidateCache: %w", err)
	}
	return nil
}
