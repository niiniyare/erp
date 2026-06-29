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

func New(pool *pgxpool.Pool) *PgTree        { return &PgTree{pool: pool} }
func (t *PgTree) InTx(tx pgx.Tx) *PgTree   { return &PgTree{pool: t.pool, tx: tx} }

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
			SELECT 1 FROM org_unit_paths
			WHERE ancestor_id = $1 AND descendant_id = $2
		)`, ancestorID, nodeID).Scan(&exists)
	return exists, err
}

func (t *PgTree) Ancestors(ctx context.Context, tenantID, nodeID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := t.querier().Query(ctx, `
		SELECT ancestor_id FROM org_unit_paths
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
		SELECT descendant_id FROM org_unit_paths
		WHERE ancestor_id = $1
		ORDER BY depth ASC`, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectUUIDs(rows)
}

func (t *PgTree) SubtreePath(ctx context.Context, tenantID, nodeID uuid.UUID) (string, error) {
	var path *string
	err := t.querier().QueryRow(ctx,
		`SELECT path FROM org_units WHERE id = $1`,
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
		SELECT id, tenant_id, parent_id, name, code, path
		FROM org_units
		WHERE id = $1`,
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

func insertPaths(ctx context.Context, tx pgx.Tx, unit *org.Unit) error {
	if _, err := tx.Exec(ctx, `
		INSERT INTO org_unit_paths (ancestor_id, descendant_id, depth)
		VALUES ($1, $1, 0)
		ON CONFLICT DO NOTHING`,
		unit.ID); err != nil {
		return fmt.Errorf("pgorg.insertPaths self: %w", err)
	}

	if unit.ParentID == nil {
		return nil
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO org_unit_paths (ancestor_id, descendant_id, depth)
		SELECT ancestor_id, $1, depth + 1
		FROM org_unit_paths
		WHERE descendant_id = $2
		ON CONFLICT DO NOTHING`,
		unit.ID, *unit.ParentID); err != nil {
		return fmt.Errorf("pgorg.insertPaths ancestors: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		UPDATE org_units
		SET path = (SELECT path FROM org_units WHERE id = $1) || '.' || $2
		WHERE id = $3`,
		*unit.ParentID, unit.ID.String(), unit.ID); err != nil {
		return fmt.Errorf("pgorg.insertPaths path: %w", err)
	}

	return nil
}

func rebuildPaths(ctx context.Context, tx pgx.Tx, unit *org.Unit) error {
	if _, err := tx.Exec(ctx, `
		DELETE FROM org_unit_paths
		WHERE descendant_id IN (
			SELECT descendant_id FROM org_unit_paths WHERE ancestor_id = $1
		)
		AND ancestor_id NOT IN (
			SELECT descendant_id FROM org_unit_paths WHERE ancestor_id = $1
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
