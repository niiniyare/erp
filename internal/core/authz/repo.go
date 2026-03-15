package authz

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/shared"
	sharedErrors "github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

const roleAssignmentCacheTTL = 5 * time.Minute

// Repository handles persistence for Casbin role assignments.
type Repository interface {
	UpsertRoleAssignment(ctx context.Context, tenantID uuid.UUID, subject, role, domain string, assignedBy, delegatedBy *string, expiresAt *time.Time) error
	DeactivateRoleAssignment(ctx context.Context, subject, role, domain string) error
	ListRoleAssignments(ctx context.Context, subject, domain string) ([]RoleAssignment, error)
	ListExpiredActiveRoleNames(ctx context.Context, subject, domain string) ([]string, error)
}

type pgRepo struct {
	store   db.Store
	cache   cache.Service
	log     logger.Logger
	metrics metrics.MetricsProvider
	tracer  tracing.Service
}

func newPgRepo(store db.Store, cacheSvc cache.Service, log logger.Logger, m metrics.MetricsProvider, t tracing.Service) Repository {
	return &pgRepo{store: store, cache: cacheSvc, log: log, metrics: m, tracer: t}
}

// UpsertRoleAssignment inserts or updates a role assignment inside a tenant-scoped
// transaction. The caller (service layer) must inject shared.WithTenantID(ctx, tenantID)
// before calling so that store.WithTenantFromCtx can set the Postgres session variable.
func (r *pgRepo) UpsertRoleAssignment(ctx context.Context, tenantID uuid.UUID, subject, role, domain string, assignedBy, delegatedBy *string, expiresAt *time.Time) error {
	ctx, span := r.tracer.StartSpan(ctx, "authz.repo.UpsertRoleAssignment")
	defer span.End()

	var exp sql.NullTime
	if expiresAt != nil {
		exp = sql.NullTime{Time: *expiresAt, Valid: true}
	}

	err := r.store.WithTenantFromCtx(ctx, func(ctx context.Context, txStore db.Store) error {
		return txStore.UpsertRoleAssignment(ctx, db.UpsertRoleAssignmentParams{
			ID:          uuid.New(),
			TenantID:    tenantID,
			Subject:     subject,
			RoleName:    role,
			Domain:      domain,
			AssignedBy:  assignedBy,
			DelegatedBy: delegatedBy,
			ExpiresAt:   exp,
		})
	})
	if err != nil {
		r.metrics.IncrementCounter("authz.repo.upsert_role_assignment.error", nil)
		return sharedErrors.NewRepositoryErrorWithContext(ctx, "UPSERT_ROLE_ASSIGNMENT_FAILED",
			"failed to upsert role assignment", err).
			WithOperation("upsert").
			WithTable("role_assignments").
			WithDetail("subject", subject).
			WithDetail("role", role).
			WithDetail("domain", domain)
	}

	r.cache.Delete(ctx, roleAssignmentCacheKey(subject, domain)) //nolint:errcheck — best-effort
	r.metrics.IncrementCounter("authz.repo.upsert_role_assignment.ok", nil)
	return nil
}

// DeactivateRoleAssignment marks a role assignment inactive.
// Uses WithTenantFromCtx when a tenant UUID is in context (tenant-scoped roles);
// falls back to WithTx for platform-level revocations where no tenant is present.
func (r *pgRepo) DeactivateRoleAssignment(ctx context.Context, subject, role, domain string) error {
	ctx, span := r.tracer.StartSpan(ctx, "authz.repo.DeactivateRoleAssignment")
	defer span.End()

	exec := func(ctx context.Context, txStore db.Store) error {
		return txStore.DeactivateRoleAssignment(ctx, db.DeactivateRoleAssignmentParams{
			Subject:  subject,
			RoleName: role,
			Domain:   domain,
		})
	}

	var err error
	if _, ok := shared.GetTenantID(ctx); ok {
		err = r.store.WithTenantFromCtx(ctx, exec)
	} else {
		err = r.store.WithTx(ctx, exec)
	}
	if err != nil {
		r.metrics.IncrementCounter("authz.repo.deactivate_role_assignment.error", nil)
		return sharedErrors.NewRepositoryErrorWithContext(ctx, "DEACTIVATE_ROLE_ASSIGNMENT_FAILED",
			"failed to deactivate role assignment", err).
			WithOperation("update").
			WithTable("role_assignments").
			WithDetail("subject", subject).
			WithDetail("role", role).
			WithDetail("domain", domain)
	}

	r.cache.Delete(ctx, roleAssignmentCacheKey(subject, domain)) //nolint:errcheck — best-effort
	r.metrics.IncrementCounter("authz.repo.deactivate_role_assignment.ok", nil)
	return nil
}

// ListRoleAssignments returns all role assignment records for a subject+domain pair.
// Results are served from cache when available; cache is populated on miss.
func (r *pgRepo) ListRoleAssignments(ctx context.Context, subject, domain string) ([]RoleAssignment, error) {
	ctx, span := r.tracer.StartSpan(ctx, "authz.repo.ListRoleAssignments")
	defer span.End()

	cacheKey := roleAssignmentCacheKey(subject, domain)

	var cached []RoleAssignment
	if err := r.cache.Get(ctx, cacheKey, &cached); err == nil {
		r.metrics.IncrementCounter("authz.repo.list_role_assignments.cache_hit", nil)
		return cached, nil
	}

	rows, err := r.store.ListRoleAssignments(ctx, db.ListRoleAssignmentsParams{
		Subject: subject,
		Domain:  domain,
	})
	if err != nil {
		r.metrics.IncrementCounter("authz.repo.list_role_assignments.error", nil)
		return nil, sharedErrors.NewRepositoryErrorWithContext(ctx, "LIST_ROLE_ASSIGNMENTS_FAILED",
			"failed to list role assignments", err).
			WithOperation("select").
			WithTable("role_assignments").
			WithDetail("subject", subject).
			WithDetail("domain", domain)
	}

	out := make([]RoleAssignment, 0, len(rows))
	for _, row := range rows {
		ra := RoleAssignment{
			ID:       row.ID.String(),
			Subject:  row.Subject,
			Role:     row.RoleName,
			Domain:   row.Domain,
			TenantID: row.TenantID.String(),
			IsActive: row.IsActive != nil && *row.IsActive,
		}
		if row.AssignedBy != nil {
			ra.AssignedBy = *row.AssignedBy
		}
		if row.DelegatedBy != nil {
			ra.DelegatedBy = *row.DelegatedBy
		}
		if row.ExpiresAt.Valid {
			t := row.ExpiresAt.Time
			ra.ExpiresAt = &t
		}
		if row.CreatedAt.Valid {
			ra.CreatedAt = row.CreatedAt.Time
		}
		out = append(out, ra)
	}

	if setErr := r.cache.Set(ctx, cacheKey, out, roleAssignmentCacheTTL); setErr != nil {
		r.log.WarnContext(ctx, "authz repo: cache.Set failed for role assignments", logger.Fields{"error": setErr.Error()})
	}

	r.metrics.IncrementCounter("authz.repo.list_role_assignments.db_hit", nil)
	return out, nil
}

// ListExpiredActiveRoleNames returns role names whose expires_at has passed.
// Intentionally not cached — expiry checks must reflect the current clock.
func (r *pgRepo) ListExpiredActiveRoleNames(ctx context.Context, subject, domain string) ([]string, error) {
	ctx, span := r.tracer.StartSpan(ctx, "authz.repo.ListExpiredActiveRoleNames")
	defer span.End()

	rows, err := r.store.ListExpiredActiveRoleNames(ctx, db.ListExpiredActiveRoleNamesParams{
		Subject: subject,
		Domain:  domain,
	})
	if err != nil {
		return nil, sharedErrors.NewRepositoryErrorWithContext(ctx, "LIST_EXPIRED_ROLES_FAILED",
			"failed to list expired role names", err).
			WithOperation("select").
			WithTable("role_assignments")
	}
	return rows, nil
}

// roleAssignmentCacheKey builds the Redis key for a subject+domain assignment list.
func roleAssignmentCacheKey(subject, domain string) string {
	return "authz:ra:" + subject + ":" + domain
}

// nullableString returns nil for empty strings so pgx stores SQL NULL.
func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
