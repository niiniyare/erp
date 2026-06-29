// Package pgorg provides a pgx-backed implementation of org.Tree.
//
// Storage model:
//   - org_units.org_unit_path  — materialized path (/root/parent/self/)
//   - org_unit_paths           — closure table (ancestor_id, descendant_id, depth)
//
// Both structures are maintained by this package via InsertPaths and RebuildPaths.
package pgorg

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/google/uuid"

	"awo.so/framework/org"
)

// PgTree implements org.Tree against PostgreSQL.
// When tx is non-nil, all operations run against that transaction.
type PgTree struct {
	pool *pgxpool.Pool
	tx   pgx.Tx // non-nil when scoped to a transaction
}

// New creates a PgTree backed by pool.
func New(pool *pgxpool.Pool) *PgTree {
	return &PgTree{pool: pool}
}

// InTx returns a PgTree scoped to tx. Use this when InsertPaths / RebuildPaths
// must share a transaction with the org_units INSERT.
func (t *PgTree) InTx(tx pgx.Tx) *PgTree {
	return &PgTree{pool: t.pool, tx: tx}
}

// querier returns the active querier (tx if set, else pool).
func (t *PgTree) querier() interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
} {
	if t.tx != nil {
		return t.tx
	}
	return t.pool
}

// ── org.Tree implementation ───────────────────────────────────────────────────

func (t *PgTree) IsAncestorOrEqual(ctx context.Context, tenantID, ancestorID, nodeID uuid.UUID) (bool, error) {
	if ancestorID == nodeID {
		return true, nil
	}
	var exists bool
	err := t.querier().QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM org_unit_paths
			WHERE tenant_id = $1 AND ancestor_id = $2 AND descendant_id = $3
		)`, tenantID, ancestorID, nodeID).Scan(&exists)
	return exists, err
}

func (t *PgTree) Ancestors(ctx context.Context, tenantID, nodeID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := t.querier().Query(ctx, `
		SELECT ancestor_id FROM org_unit_paths
		WHERE tenant_id = $1 AND descendant_id = $2
		ORDER BY depth ASC`, tenantID, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectUUIDs(rows)
}

func (t *PgTree) Descendants(ctx context.Context, tenantID, nodeID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := t.querier().Query(ctx, `
		SELECT descendant_id FROM org_unit_paths
		WHERE tenant_id = $1 AND ancestor_id = $2
		ORDER BY depth ASC`, tenantID, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectUUIDs(rows)
}

func (t *PgTree) SubtreePath(ctx context.Context, tenantID, nodeID uuid.UUID) (string, error) {
	var path *string
	err := t.querier().QueryRow(ctx,
		`SELECT org_unit_path FROM org_units WHERE uuid = $1 AND tenant_id = $2`,
		nodeID, tenantID).Scan(&path)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", org.ErrUnitNotFound
	}
	if err != nil {
		return "", err
	}
	if path == nil || *path == "" {
		return "", org.ErrUnitNotFound
	}
	return *path + "%", nil
}

func (t *PgTree) Unit(ctx context.Context, tenantID, unitID uuid.UUID) (*org.Unit, error) {
	u := &org.Unit{}
	err := t.querier().QueryRow(ctx, `
		SELECT uuid, tenant_id, parent_id, name, code, type, is_active, hidden,
		       accrual_method, fy_start_month, address, picture,
		       org_unit_path, org_level, settings, metadata, version,
		       validation_status, created_at, updated_at, deleted_at
		FROM org_units
		WHERE uuid = $1 AND tenant_id = $2 AND deleted_at IS NULL`,
		unitID, tenantID).Scan(
		&u.ID, &u.TenantID, &u.ParentID, &u.Name, &u.Code, &u.Type,
		&u.IsActive, &u.Hidden, &u.AccrualMethod, &u.FYStartMonth,
		&u.Address, &u.Picture, &u.OrgUnitPath, &u.OrgLevel,
		&u.Settings, &u.Metadata, &u.Version,
		&u.ValidationStatus, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, org.ErrUnitNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("pgorg.Unit: %w", err)
	}
	return u, nil
}

// InsertPaths inserts closure rows for a new unit.
// When called on a tx-scoped tree (InTx), uses that transaction.
// Otherwise wraps in its own transaction.
func (t *PgTree) InsertPaths(ctx context.Context, unit *org.Unit) error {
	if t.tx != nil {
		return insertPaths(ctx, t.tx, unit)
	}
	tx, err := t.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	if err := insertPaths(ctx, tx, unit); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// RebuildPaths rebuilds closure rows after a reparent.
// When called on a tx-scoped tree (InTx), uses that transaction.
// Otherwise wraps in its own transaction.
func (t *PgTree) RebuildPaths(ctx context.Context, unit *org.Unit) error {
	if t.tx != nil {
		return rebuildPaths(ctx, t.tx, unit)
	}
	tx, err := t.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	if err := rebuildPaths(ctx, tx, unit); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// ── Internal SQL helpers ──────────────────────────────────────────────────────

func insertPaths(ctx context.Context, tx pgx.Tx, unit *org.Unit) error {
	// Self-reference.
	if _, err := tx.Exec(ctx, `
		INSERT INTO org_unit_paths (tenant_id, ancestor_id, descendant_id, depth)
		VALUES ($1, $2, $2, 0)
		ON CONFLICT DO NOTHING`,
		unit.TenantID, unit.ID); err != nil {
		return fmt.Errorf("pgorg.insertPaths self: %w", err)
	}

	if unit.ParentID == nil {
		return nil
	}

	// Copy ancestor rows from parent, incrementing depth.
	if _, err := tx.Exec(ctx, `
		INSERT INTO org_unit_paths (tenant_id, ancestor_id, descendant_id, depth)
		SELECT tenant_id, ancestor_id, $1, depth + 1
		FROM org_unit_paths
		WHERE tenant_id = $2 AND descendant_id = $3
		ON CONFLICT DO NOTHING`,
		unit.ID, unit.TenantID, *unit.ParentID); err != nil {
		return fmt.Errorf("pgorg.insertPaths ancestors: %w", err)
	}

	// Update materialized path.
	if _, err := tx.Exec(ctx, `
		UPDATE org_units
		SET org_unit_path = (
			SELECT org_unit_path FROM org_units WHERE uuid = $1
		) || $2::text || '/'
		WHERE uuid = $3`,
		*unit.ParentID, unit.ID.String(), unit.ID); err != nil {
		return fmt.Errorf("pgorg.insertPaths mat-path: %w", err)
	}

	return nil
}

func rebuildPaths(ctx context.Context, tx pgx.Tx, unit *org.Unit) error {
	// Delete all existing paths where unit (or its descendants) is the descendant.
	if _, err := tx.Exec(ctx, `
		DELETE FROM org_unit_paths
		WHERE tenant_id = $1
		  AND descendant_id IN (
			SELECT descendant_id FROM org_unit_paths
			WHERE tenant_id = $1 AND ancestor_id = $2
		  )
		  AND ancestor_id NOT IN (
			SELECT descendant_id FROM org_unit_paths
			WHERE tenant_id = $1 AND ancestor_id = $2
		  )`,
		unit.TenantID, unit.ID); err != nil {
		return fmt.Errorf("pgorg.rebuildPaths delete: %w", err)
	}

	// Re-insert from new position.
	return insertPaths(ctx, tx, unit)
}

func collectUUIDs(rows pgx.Rows) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

var _ org.Tree = (*PgTree)(nil)
