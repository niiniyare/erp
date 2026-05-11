package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	db "awo.so/db/sqlc"
	"awo.so/internal/core/iam/domain"
	"awo.so/internal/platform/cache"
	"awo.so/internal/shared"
	sharedErrors "awo.so/internal/shared/errors"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
)

const roleAssignmentCacheTTL = 5 * time.Minute

// Port (interface)

// AuthzRepository handles persistence for Casbin role assignments.
// Cache is managed internally — callers never touch cache directly.
type AuthzRepository interface {
	UpsertRoleAssignment(ctx context.Context, tenantID uuid.UUID, subject, role, domainName string, assignedBy, delegatedBy *string, expiresAt *time.Time) error
	DeactivateRoleAssignment(ctx context.Context, subject, role, domainName string) error
	ListRoleAssignments(ctx context.Context, subject, domainName string) ([]domain.RoleAssignment, error)
	ListExpiredActiveRoleNames(ctx context.Context, subject, domainName string) ([]string, error)
}

// Adapter (implementation)

type authzRepo struct {
	store   db.Store
	cache   cache.Service
	log     logger.Logger
	metrics metrics.MetricsProvider
	tracer  tracing.Service
}

// NewAuthzRepository constructs a cache-backed Postgres AuthzRepository.
func NewAuthzRepository(
	store db.Store,
	cacheSvc cache.Service,
	log logger.Logger,
	m metrics.MetricsProvider,
	tracer tracing.Service,
) AuthzRepository {
	return &authzRepo{store: store, cache: cacheSvc, log: log, metrics: m, tracer: tracer}
}

func (r *authzRepo) UpsertRoleAssignment(
	ctx context.Context,
	tenantID uuid.UUID,
	subject, role, domainName string,
	assignedBy, delegatedBy *string,
	expiresAt *time.Time,
) error {
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
			Domain:      domainName,
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
			WithDetail("domain", domainName)
	}

	_ = r.cache.Delete(ctx, roleAssignmentCacheKey(subject, domainName))
	r.metrics.IncrementCounter("authz.repo.upsert_role_assignment.ok", nil)
	return nil
}

func (r *authzRepo) DeactivateRoleAssignment(ctx context.Context, subject, role, domainName string) error {
	ctx, span := r.tracer.StartSpan(ctx, "authz.repo.DeactivateRoleAssignment")
	defer span.End()

	exec := func(ctx context.Context, txStore db.Store) error {
		return txStore.DeactivateRoleAssignment(ctx, db.DeactivateRoleAssignmentParams{
			Subject:  subject,
			RoleName: role,
			Domain:   domainName,
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
			WithDetail("domain", domainName)
	}

	_ = r.cache.Delete(ctx, roleAssignmentCacheKey(subject, domainName))
	r.metrics.IncrementCounter("authz.repo.deactivate_role_assignment.ok", nil)
	return nil
}

func (r *authzRepo) ListRoleAssignments(ctx context.Context, subject, domainName string) ([]domain.RoleAssignment, error) {
	ctx, span := r.tracer.StartSpan(ctx, "authz.repo.ListRoleAssignments")
	defer span.End()

	cacheKey := roleAssignmentCacheKey(subject, domainName)

	var cached []domain.RoleAssignment
	if err := r.cache.Get(ctx, cacheKey, &cached); err == nil {
		r.metrics.IncrementCounter("authz.repo.list_role_assignments.cache_hit", nil)
		return cached, nil
	}

	rows, err := r.store.ListRoleAssignments(ctx, db.ListRoleAssignmentsParams{
		Subject: subject,
		Domain:  domainName,
	})
	if err != nil {
		r.metrics.IncrementCounter("authz.repo.list_role_assignments.error", nil)
		return nil, sharedErrors.NewRepositoryErrorWithContext(ctx, "LIST_ROLE_ASSIGNMENTS_FAILED",
			"failed to list role assignments", err).
			WithOperation("select").
			WithTable("role_assignments").
			WithDetail("subject", subject).
			WithDetail("domain", domainName)
	}

	out := make([]domain.RoleAssignment, 0, len(rows))
	for _, row := range rows {
		ra := domain.RoleAssignment{
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

	if err := r.cache.Set(ctx, cacheKey, out, roleAssignmentCacheTTL); err != nil {
		r.log.WarnContext(ctx, "authz repo: cache.Set failed for role assignments", logger.Fields{"error": err.Error()})
	}

	r.metrics.IncrementCounter("authz.repo.list_role_assignments.db_hit", nil)
	return out, nil
}

func (r *authzRepo) ListExpiredActiveRoleNames(ctx context.Context, subject, domainName string) ([]string, error) {
	ctx, span := r.tracer.StartSpan(ctx, "authz.repo.ListExpiredActiveRoleNames")
	defer span.End()

	rows, err := r.store.ListExpiredActiveRoleNames(ctx, db.ListExpiredActiveRoleNamesParams{
		Subject: subject,
		Domain:  domainName,
	})
	if err != nil {
		return nil, sharedErrors.NewRepositoryErrorWithContext(ctx, "LIST_EXPIRED_ROLES_FAILED",
			"failed to list expired role names", err).
			WithOperation("select").
			WithTable("role_assignments")
	}
	return rows, nil
}

func roleAssignmentCacheKey(subject, domainName string) string {
	return "authz:ra:" + subject + ":" + domainName
}

// Casbin pgx Adapter

// NewPgxAdapter constructs a Casbin persist.BatchAdapter backed by pgx.
// Used only by NewAuthzService during initialization.
func NewPgxAdapter(pool *pgxpool.Pool) persist.BatchAdapter {
	return &pgxAdapter{pool: pool}
}

type pgxAdapter struct {
	pool *pgxpool.Pool
}

func (a *pgxAdapter) LoadPolicy(m model.Model) error {
	ctx := context.Background()
	rows, err := a.pool.Query(ctx,
		`SELECT ptype, v0, v1, v2, v3, v4, v5 FROM casbin_rule ORDER BY ptype`)
	if err != nil {
		return fmt.Errorf("authz adapter LoadPolicy: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var ptype, v0, v1, v2, v3, v4, v5 string
		if err := rows.Scan(&ptype, &v0, &v1, &v2, &v3, &v4, &v5); err != nil {
			return fmt.Errorf("authz adapter LoadPolicy scan: %w", err)
		}
		rule := filterEmpty([]string{v0, v1, v2, v3, v4, v5})
		persist.LoadPolicyLine(fmt.Sprintf("%s, %s", ptype, joinRule(rule)), m)
	}
	return rows.Err()
}

func (a *pgxAdapter) SavePolicy(m model.Model) error {
	ctx := context.Background()
	tx, err := a.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("authz adapter SavePolicy begin: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if _, err := tx.Exec(ctx, `TRUNCATE casbin_rule`); err != nil {
		return fmt.Errorf("authz adapter SavePolicy truncate: %w", err)
	}

	for ptype, assertions := range m["p"] {
		for _, rule := range assertions.Policy {
			v0, v1, v2, v3, v4, v5 := ruleToValues(rule)
			if _, err := tx.Exec(ctx,
				`INSERT INTO casbin_rule(ptype,v0,v1,v2,v3,v4,v5) VALUES($1,$2,$3,$4,$5,$6,$7) ON CONFLICT DO NOTHING`,
				ptype, v0, v1, v2, v3, v4, v5,
			); err != nil {
				return fmt.Errorf("authz adapter SavePolicy insert p: %w", err)
			}
		}
	}
	for ptype, assertions := range m["g"] {
		for _, rule := range assertions.Policy {
			v0, v1, v2, v3, v4, v5 := ruleToValues(rule)
			if _, err := tx.Exec(ctx,
				`INSERT INTO casbin_rule(ptype,v0,v1,v2,v3,v4,v5) VALUES($1,$2,$3,$4,$5,$6,$7) ON CONFLICT DO NOTHING`,
				ptype, v0, v1, v2, v3, v4, v5,
			); err != nil {
				return fmt.Errorf("authz adapter SavePolicy insert g: %w", err)
			}
		}
	}
	return tx.Commit(ctx)
}

func (a *pgxAdapter) AddPolicy(sec, ptype string, rule []string) error {
	v0, v1, v2, v3, v4, v5 := ruleToValues(rule)
	_, err := a.pool.Exec(context.Background(),
		`INSERT INTO casbin_rule(ptype,v0,v1,v2,v3,v4,v5) VALUES($1,$2,$3,$4,$5,$6,$7) ON CONFLICT DO NOTHING`,
		ptype, v0, v1, v2, v3, v4, v5,
	)
	return err
}

func (a *pgxAdapter) AddPolicies(sec, ptype string, rules [][]string) error {
	ctx := context.Background()
	tx, err := a.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	for _, rule := range rules {
		v0, v1, v2, v3, v4, v5 := ruleToValues(rule)
		if _, err := tx.Exec(ctx,
			`INSERT INTO casbin_rule(ptype,v0,v1,v2,v3,v4,v5) VALUES($1,$2,$3,$4,$5,$6,$7) ON CONFLICT DO NOTHING`,
			ptype, v0, v1, v2, v3, v4, v5,
		); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (a *pgxAdapter) RemovePolicy(sec, ptype string, rule []string) error {
	v0, v1, v2, v3, v4, v5 := ruleToValues(rule)
	_, err := a.pool.Exec(context.Background(),
		`DELETE FROM casbin_rule WHERE ptype=$1 AND v0=$2 AND v1=$3 AND v2=$4 AND v3=$5 AND v4=$6 AND v5=$7`,
		ptype, v0, v1, v2, v3, v4, v5,
	)
	return err
}

func (a *pgxAdapter) RemovePolicies(sec, ptype string, rules [][]string) error {
	ctx := context.Background()
	tx, err := a.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	for _, rule := range rules {
		v0, v1, v2, v3, v4, v5 := ruleToValues(rule)
		if _, err := tx.Exec(ctx,
			`DELETE FROM casbin_rule WHERE ptype=$1 AND v0=$2 AND v1=$3 AND v2=$4 AND v3=$5 AND v4=$6 AND v5=$7`,
			ptype, v0, v1, v2, v3, v4, v5,
		); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (a *pgxAdapter) RemoveFilteredPolicy(sec, ptype string, fieldIndex int, fieldValues ...string) error {
	query := `DELETE FROM casbin_rule WHERE ptype=$1`
	args := []any{ptype}
	cols := []string{"v0", "v1", "v2", "v3", "v4", "v5"}

	for i, val := range fieldValues {
		if val == "" {
			continue
		}
		col := cols[fieldIndex+i]
		args = append(args, val)
		query += fmt.Sprintf(" AND %s=$%d", col, len(args))
	}

	_, err := a.pool.Exec(context.Background(), query, args...)
	return err
}

func ruleToValues(rule []string) (v0, v1, v2, v3, v4, v5 string) {
	padded := make([]string, 6)
	copy(padded, rule)
	return padded[0], padded[1], padded[2], padded[3], padded[4], padded[5]
}

func filterEmpty(ss []string) []string {
	last := len(ss) - 1
	for last >= 0 && ss[last] == "" {
		last--
	}
	return ss[:last+1]
}

func joinRule(rule []string) string {
	out := ""
	for i, v := range rule {
		if i > 0 {
			out += ", "
		}
		out += v
	}
	return out
}
