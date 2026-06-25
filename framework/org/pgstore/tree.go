// Package pgstore provides a PostgreSQL-backed implementation of org.Tree.
//
// It uses two complementary structures maintained in the database:
//
//   - org_unit_paths (closure table) — every ancestor–descendant pair with depth.
//     Powers IsAncestorOrEqual, Ancestors, Descendants in O(1) / index-seek.
//
//   - org_units.org_unit_path (materialized path) — '/root/parent/self/' string.
//     Powers SubtreePath for fast LIKE-based subtree filtering.
//
// The closure table is NOT self-maintained by DB triggers. Callers MUST invoke
// InsertPaths after every org unit INSERT and RebuildPaths after every reparent.
// The DB trigger (check_org_unit_hierarchy_depth) enforces max depth (8) and
// cycle detection; this layer does not duplicate those checks.
package pgstore

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"awo.so/framework/org"

	"github.com/google/uuid"
)

// PgTree is a PostgreSQL-backed org.Tree.
type PgTree struct {
	pool *pgxpool.Pool
}

// New returns a PgTree backed by pool.
func New(pool *pgxpool.Pool) *PgTree {
	return &PgTree{pool: pool}
}

// Ensure PgTree satisfies org.Tree at compile time.
var _ org.Tree = (*PgTree)(nil)

// ── org.Tree implementation ────────────────────────────────────────────────

// IsAncestorOrEqual returns true when ancestorID is at or above nodeID in the
// org_unit_paths closure table, including the self-reference case (depth = 0).
//
// SQL: SELECT 1 FROM org_unit_paths
//
//	WHERE tenant_id = $1 AND ancestor_id = $2 AND descendant_id = $3
func (t *PgTree) IsAncestorOrEqual(
	ctx context.Context,
	tenantID, ancestorID, nodeID uuid.UUID,
) (bool, error) {
	if ancestorID == nodeID {
		return true, nil // self is always an ancestor of self
	}
	const q = `
		SELECT 1 FROM org_unit_paths
		WHERE tenant_id    = $1
		  AND ancestor_id  = $2
		  AND descendant_id = $3
		LIMIT 1`

	var dummy int
	err := t.pool.QueryRow(ctx, q, tenantID, ancestorID, nodeID).Scan(&dummy)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("org/pgstore: IsAncestorOrEqual: %w", err)
	}
	return true, nil
}

// Ancestors returns the chain of ancestor UUIDs from nodeID up to the root,
// ordered by depth ASC (depth 0 = self, depth 1 = parent, …).
//
// SQL: SELECT ancestor_id FROM org_unit_paths
//
//	WHERE tenant_id = $1 AND descendant_id = $2
//	ORDER BY depth ASC
func (t *PgTree) Ancestors(
	ctx context.Context,
	tenantID, nodeID uuid.UUID,
) ([]uuid.UUID, error) {
	const q = `
		SELECT ancestor_id
		FROM   org_unit_paths
		WHERE  tenant_id     = $1
		  AND  descendant_id = $2
		ORDER  BY depth ASC`

	rows, err := t.pool.Query(ctx, q, tenantID, nodeID)
	if err != nil {
		return nil, fmt.Errorf("org/pgstore: Ancestors: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("org/pgstore: Ancestors scan: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// Descendants returns all descendant UUIDs under nodeID (inclusive of nodeID
// itself at depth 0), ordered by depth ASC.
//
// SQL: SELECT descendant_id FROM org_unit_paths
//
//	WHERE tenant_id = $1 AND ancestor_id = $2
//	ORDER BY depth ASC
func (t *PgTree) Descendants(
	ctx context.Context,
	tenantID, nodeID uuid.UUID,
) ([]uuid.UUID, error) {
	const q = `
		SELECT descendant_id
		FROM   org_unit_paths
		WHERE  tenant_id   = $1
		  AND  ancestor_id = $2
		ORDER  BY depth ASC`

	rows, err := t.pool.Query(ctx, q, tenantID, nodeID)
	if err != nil {
		return nil, fmt.Errorf("org/pgstore: Descendants: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("org/pgstore: Descendants scan: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// SubtreePath returns the LIKE pattern for the org unit's subtree.
// Example: "/3fa85f64-…/" → "/3fa85f64-…/%"
//
// Returns org.ErrUnitNotFound when the unit is absent, soft-deleted, or has
// no org_unit_path set.
func (t *PgTree) SubtreePath(
	ctx context.Context,
	tenantID, nodeID uuid.UUID,
) (string, error) {
	const q = `
		SELECT org_unit_path
		FROM   org_units
		WHERE  uuid       = $1
		  AND  tenant_id  = $2
		  AND  deleted_at IS NULL`

	var path *string
	err := t.pool.QueryRow(ctx, q, nodeID, tenantID).Scan(&path)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", org.ErrUnitNotFound
	}
	if err != nil {
		return "", fmt.Errorf("org/pgstore: SubtreePath: %w", err)
	}
	if path == nil || *path == "" {
		return "", org.ErrUnitNotFound
	}
	return *path + "%", nil
}

// Unit fetches a single org unit by UUID. Returns org.ErrUnitNotFound when
// absent or soft-deleted.
func (t *PgTree) Unit(
	ctx context.Context,
	tenantID, unitID uuid.UUID,
) (*org.Unit, error) {
	const q = `
		SELECT
			uuid, tenant_id, parent_id, name, code, type,
			is_active, hidden, accrual_method, fy_start_month,
			address, picture, org_unit_path, org_level,
			settings, metadata, version,
			validation_status, validation_errors, last_validation_run,
			created_at, updated_at, deleted_at
		FROM org_units
		WHERE uuid      = $1
		  AND tenant_id = $2
		  AND deleted_at IS NULL`

	u := &org.Unit{}
	err := t.pool.QueryRow(ctx, q, unitID, tenantID).Scan(
		&u.ID, &u.TenantID, &u.ParentID, &u.Name, &u.Code, &u.Type,
		&u.IsActive, &u.Hidden, &u.AccrualMethod, &u.FYStartMonth,
		&u.Address, &u.Picture, &u.OrgUnitPath, &u.OrgLevel,
		&u.Settings, &u.Metadata, &u.Version,
		&u.ValidationStatus, &u.ValidationErrors, &u.LastValidationRun,
		&u.CreatedAt, &u.UpdatedAt, &u.DeletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, org.ErrUnitNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("org/pgstore: Unit: %w", err)
	}
	return u, nil
}

// InsertPaths inserts closure table rows for a newly created org unit.
// Must be called inside the same transaction that inserted the org_units row.
//
// Inserts:
//   - self-reference row (ancestor = descendant = unit.ID, depth = 0)
//   - one row per ancestor copied from the parent's closure rows, depth + 1
//
// SQL pattern:
//
//	INSERT INTO org_unit_paths (tenant_id, ancestor_id, descendant_id, depth)
//	  -- self-reference
//	  SELECT $1, $2, $2, 0
//	  UNION ALL
//	  -- all ancestors of parent, at depth + 1
//	  SELECT $1, ancestor_id, $2, depth + 1
//	  FROM org_unit_paths
//	  WHERE tenant_id = $1 AND descendant_id = $3  -- $3 = parent_id
func (t *PgTree) InsertPaths(ctx context.Context, unit *org.Unit) error {
	if unit.ParentID == nil {
		// Root unit — only self-reference needed.
		const q = `
			INSERT INTO org_unit_paths (tenant_id, ancestor_id, descendant_id, depth)
			VALUES ($1, $2, $2, 0)
			ON CONFLICT (tenant_id, ancestor_id, descendant_id) DO NOTHING`
		_, err := t.pool.Exec(ctx, q, unit.TenantID, unit.ID)
		if err != nil {
			return fmt.Errorf("org/pgstore: InsertPaths (root): %w", err)
		}
		return nil
	}

	const q = `
		INSERT INTO org_unit_paths (tenant_id, ancestor_id, descendant_id, depth)
		SELECT $1, ancestor_id, $2, depth + 1
		FROM   org_unit_paths
		WHERE  tenant_id     = $1
		  AND  descendant_id = $3
		UNION ALL
		SELECT $1, $2, $2, 0
		ON CONFLICT (tenant_id, ancestor_id, descendant_id) DO NOTHING`

	_, err := t.pool.Exec(ctx, q, unit.TenantID, unit.ID, *unit.ParentID)
	if err != nil {
		return fmt.Errorf("org/pgstore: InsertPaths: %w", err)
	}
	return nil
}

// RebuildPaths rebuilds closure table rows for unit and all its descendants
// after a reparent operation. Also recomputes org_unit_path and org_level on
// all affected rows.
//
// Strategy (executed in a transaction):
//  1. Delete all closure rows where descendant is in unit's subtree but
//     ancestor is NOT in that subtree (i.e. rows that connected old parent chain).
//  2. Re-insert ancestor rows by joining the new parent's closure rows.
//  3. Recompute org_unit_path and org_level for all descendants.
func (t *PgTree) RebuildPaths(ctx context.Context, unit *org.Unit) error {
	tx, err := t.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("org/pgstore: RebuildPaths begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if err := rebuildPathsTx(ctx, tx, unit); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func rebuildPathsTx(ctx context.Context, tx pgx.Tx, unit *org.Unit) error {
	// 1. Collect all descendant IDs (including self) for this subtree.
	const getDescendants = `
		SELECT descendant_id FROM org_unit_paths
		WHERE tenant_id = $1 AND ancestor_id = $2`

	rows, err := tx.Query(ctx, getDescendants, unit.TenantID, unit.ID)
	if err != nil {
		return fmt.Errorf("org/pgstore: RebuildPaths query descendants: %w", err)
	}
	var descendants []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		descendants = append(descendants, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	// 2. Delete stale closure rows: any row where descendant ∈ subtree but
	//    ancestor ∉ subtree (these were connections to the old parent chain).
	const deleteStale = `
		DELETE FROM org_unit_paths
		WHERE  tenant_id     = $1
		  AND  descendant_id = ANY($2)
		  AND  ancestor_id   NOT IN (
		    SELECT descendant_id FROM org_unit_paths
		    WHERE tenant_id = $1 AND ancestor_id = $3
		  )`
	if _, err := tx.Exec(ctx, deleteStale, unit.TenantID, descendants, unit.ID); err != nil {
		return fmt.Errorf("org/pgstore: RebuildPaths delete stale: %w", err)
	}

	// 3. Re-insert: for each descendant, insert ancestor rows from new parent chain.
	if unit.ParentID != nil {
		const reinsert = `
			INSERT INTO org_unit_paths (tenant_id, ancestor_id, descendant_id, depth)
			SELECT p.tenant_id, p.ancestor_id, c.descendant_id, p.depth + c.depth + 1
			FROM   org_unit_paths p
			JOIN   org_unit_paths c ON c.ancestor_id = $2 -- $2 = unit.ID (subtree root)
			WHERE  p.tenant_id     = $1
			  AND  p.descendant_id = $3 -- $3 = new parent_id
			ON CONFLICT (tenant_id, ancestor_id, descendant_id) DO UPDATE
			  SET depth = EXCLUDED.depth, updated_at = NOW()`
		if _, err := tx.Exec(ctx, reinsert, unit.TenantID, unit.ID, *unit.ParentID); err != nil {
			return fmt.Errorf("org/pgstore: RebuildPaths reinsert: %w", err)
		}
	}

	// 4. Recompute org_level and org_unit_path for all descendants.
	const updateMeta = `
		UPDATE org_units ou
		SET
			org_level = (
				SELECT COUNT(*) FROM org_unit_paths
				WHERE tenant_id = $1 AND descendant_id = ou.uuid
			),
			org_unit_path = (
				WITH chain AS (
					SELECT ancestor_id FROM org_unit_paths
					WHERE tenant_id = $1 AND descendant_id = ou.uuid
					ORDER BY depth DESC
				)
				SELECT '/' || string_agg(ancestor_id::text, '/' ORDER BY ordinality) || '/'
				FROM   chain WITH ORDINALITY
			),
			updated_at = NOW()
		WHERE tenant_id = $1
		  AND uuid = ANY($2)`
	if _, err := tx.Exec(ctx, updateMeta, unit.TenantID, descendants); err != nil {
		return fmt.Errorf("org/pgstore: RebuildPaths update meta: %w", err)
	}

	return nil
}
