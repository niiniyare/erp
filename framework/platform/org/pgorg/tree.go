// Package pgorg provides a pgx-backed implementation of org.Tree.
package pgorg

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"awo.so/framework/platform/org"
)

// PgTree implements org.Tree against PostgreSQL.
type PgTree struct {
	pool *pgxpool.Pool
	tx   pgx.Tx
}

func New(pool *pgxpool.Pool) *PgTree     { return &PgTree{pool: pool} }
func (t *PgTree) InTx(tx pgx.Tx) *PgTree { return &PgTree{pool: t.pool, tx: tx} }

func (t *PgTree) querier() interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
} {
	if t.tx != nil {
		return t.tx
	}
	return t.pool
}

func (t *PgTree) IsAncestorOrEqual(ctx context.Context, tenantID, ancestorID, nodeID uuid.UUID) (bool, error) {
	if ancestorID == nodeID {
		return true, nil
	}
	var exists bool
	err := t.querier().QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM hierarchy_paths
			WHERE ancestor_id = $1 AND descendant_id = $2
		)`, ancestorID, nodeID).Scan(&exists)
	return exists, err
}

func (t *PgTree) Ancestors(ctx context.Context, tenantID, nodeID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := t.querier().Query(ctx, `
		SELECT ancestor_id FROM hierarchy_paths
		WHERE descendant_id = $1
		ORDER BY depth ASC`, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectUUIDs(rows)
}

func (t *PgTree) Descendants(ctx context.Context, tenantID, nodeID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := t.querier().Query(ctx, `
		SELECT descendant_id FROM hierarchy_paths
		WHERE ancestor_id = $1
		ORDER BY depth ASC`, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectUUIDs(rows)
}

// SubtreePath returns the org_path of nodeID with a trailing '%' suitable for
// a LIKE predicate: WHERE org_path LIKE SubtreePath(nodeID).
func (t *PgTree) SubtreePath(ctx context.Context, tenantID, nodeID uuid.UUID) (string, error) {
	var path *string
	err := t.querier().QueryRow(ctx,
		`SELECT org_path FROM orgunits WHERE uuid = $1`,
		nodeID).Scan(&path)
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
		SELECT uuid, tenant_id, parent_id, name, code, org_path
		FROM orgunits
		WHERE uuid = $1`,
		unitID).Scan(
		&u.ID, &u.TenantID, &u.ParentID, &u.Name, &u.Code, &u.OrgUnitPath,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, org.ErrUnitNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("pgorg.Unit: %w", err)
	}
	return u, nil
}

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

// insertPaths writes the closure table entries for unit and updates its
// materialized org_path in orgunits.
//
// hierarchy_paths stores every ancestor-descendant pair (including self at
// depth 0) with tenant_id for RLS. org_path in orgunits follows the format:
//
//	/root_uuid/parent_uuid/this_uuid/
//
// Root units (ParentID == nil) skip the path update — their org_path is set
// during provisioning when the org unit row is first created.
func insertPaths(ctx context.Context, tx pgx.Tx, unit *org.Unit) error {
	// Self-reference at depth 0.
	if _, err := tx.Exec(ctx, `
		INSERT INTO hierarchy_paths (tenant_id, ancestor_id, descendant_id, depth)
		VALUES ($1, $2, $2, 0)
		ON CONFLICT DO NOTHING`,
		unit.TenantID, unit.ID); err != nil {
		return fmt.Errorf("pgorg.insertPaths self: %w", err)
	}

	if unit.ParentID == nil {
		return nil
	}

	// Inherit all ancestor rows from the parent, incrementing depth.
	if _, err := tx.Exec(ctx, `
		INSERT INTO hierarchy_paths (tenant_id, ancestor_id, descendant_id, depth)
		SELECT tenant_id, ancestor_id, $1, depth + 1
		FROM hierarchy_paths
		WHERE descendant_id = $2
		ON CONFLICT DO NOTHING`,
		unit.ID, *unit.ParentID); err != nil {
		return fmt.Errorf("pgorg.insertPaths ancestors: %w", err)
	}

	// Extend the materialized path: parent_path + this_uuid + '/'.
	// e.g. parent = '/root/' → child = '/root/child_uuid/'
	if _, err := tx.Exec(ctx, `
		UPDATE orgunits
		SET org_path = (SELECT org_path FROM orgunits WHERE uuid = $1) || $2 || '/'
		WHERE uuid = $3`,
		*unit.ParentID, unit.ID.String(), unit.ID); err != nil {
		return fmt.Errorf("pgorg.insertPaths org_path: %w", err)
	}

	return nil
}

// rebuildPaths removes stale closure-table entries for the subtree rooted at
// unit and re-inserts them via insertPaths. Used on reparenting.
func rebuildPaths(ctx context.Context, tx pgx.Tx, unit *org.Unit) error {
	// Delete all entries that point INTO the subtree from OUTSIDE it.
	// Specifically: descendant is in our subtree AND ancestor is not.
	if _, err := tx.Exec(ctx, `
		DELETE FROM hierarchy_paths
		WHERE descendant_id IN (
			SELECT descendant_id FROM hierarchy_paths WHERE ancestor_id = $1
		)
		AND ancestor_id NOT IN (
			SELECT descendant_id FROM hierarchy_paths WHERE ancestor_id = $1
		)`,
		unit.ID); err != nil {
		return fmt.Errorf("pgorg.rebuildPaths delete: %w", err)
	}
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
