package pgstore

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"awo.so/framework/org"
)

// CreateUnitInput carries the fields required to create a new org unit.
// Fields not set default to safe values (is_active=true, org_level computed).
type CreateUnitInput struct {
	// TenantID is required.
	TenantID uuid.UUID

	// ParentID is nil for root COMPANY units.
	ParentID *uuid.UUID

	// Name is the display name — must be unique within the tenant.
	Name string

	// Code is an optional short reference code, unique per tenant when set.
	Code *string

	// Type must be a valid org.UnitType value.
	Type org.UnitType

	// AccrualMethod: true = accrual, false = cash.
	AccrualMethod bool

	// FYStartMonth is 1–12. Defaults to 1 (January) when zero.
	FYStartMonth int

	// Address is a flexible JSON object.
	Address map[string]any

	// Picture is an optional logo path or URL.
	Picture *string

	// Settings holds per-unit configuration overrides.
	Settings map[string]any

	// Metadata holds arbitrary key-value extension data.
	Metadata map[string]any
}

// CreateUnit inserts a new org unit and its closure table rows in a single
// transaction. It also computes and stores org_unit_path and org_level.
//
// On success returns the fully populated *org.Unit including generated fields
// (ID, OrgLevel, OrgUnitPath, CreatedAt, UpdatedAt).
//
// The DB trigger check_org_unit_hierarchy_depth fires on INSERT and will
// return an error if the new unit would create a cycle or exceed depth 8.
func (t *PgTree) CreateUnit(ctx context.Context, in CreateUnitInput) (*org.Unit, error) {
	if in.TenantID == uuid.Nil {
		return nil, fmt.Errorf("org/pgstore: CreateUnit: tenant_id is required")
	}
	if !in.Type.IsValid() {
		return nil, fmt.Errorf("org/pgstore: CreateUnit: invalid unit type %q", in.Type)
	}
	if in.FYStartMonth == 0 {
		in.FYStartMonth = 1
	}
	if in.Address == nil {
		in.Address = map[string]any{}
	}
	if in.Settings == nil {
		in.Settings = map[string]any{}
	}
	if in.Metadata == nil {
		in.Metadata = map[string]any{}
	}

	id := uuid.New()

	tx, err := t.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("org/pgstore: CreateUnit begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	unit, err := insertUnitTx(ctx, tx, id, in)
	if err != nil {
		return nil, err
	}

	// InsertPaths requires a Tree — reuse t but operate through the tx connection.
	// We can't call t.InsertPaths directly (it uses the pool); call the tx variant.
	if err := insertPathsTx(ctx, tx, unit); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("org/pgstore: CreateUnit commit: %w", err)
	}
	return unit, nil
}

// insertUnitTx inserts the org_units row and computes org_unit_path + org_level.
func insertUnitTx(ctx context.Context, tx pgx.Tx, id uuid.UUID, in CreateUnitInput) (*org.Unit, error) {
	// Compute org_level: parent_level + 1, or 1 for root.
	orgLevel := 1
	var parentPath string
	if in.ParentID != nil {
		const qParent = `
			SELECT org_level, COALESCE(org_unit_path, '/' || uuid::text || '/')
			FROM org_units
			WHERE uuid = $1 AND tenant_id = $2 AND deleted_at IS NULL`
		if err := tx.QueryRow(ctx, qParent, *in.ParentID, in.TenantID).
			Scan(&orgLevel, &parentPath); err != nil {
			return nil, fmt.Errorf("org/pgstore: CreateUnit fetch parent: %w", err)
		}
		orgLevel++ // child is one deeper than parent
	}

	// Materialized path: parentPath + this_uuid + "/"
	// Root: "/uuid/"   Child: "/root/parent/uuid/"
	unitPath := parentPath + id.String() + "/"
	if in.ParentID == nil {
		unitPath = "/" + id.String() + "/"
	}

	const qInsert = `
		INSERT INTO org_units (
			uuid, tenant_id, parent_id, name, code, type,
			is_active, hidden, accrual_method, fy_start_month,
			address, picture, org_unit_path, org_level,
			settings, metadata, version,
			validation_status, validation_errors,
			created_at, updated_at
		) VALUES (
			$1,  $2,  $3,  $4,  $5,  $6,
			TRUE, FALSE, $7, $8,
			$9,  $10, $11, $12,
			$13, $14, 1,
			'PENDING', '[]',
			NOW(), NOW()
		)
		RETURNING created_at, updated_at`

	unit := &org.Unit{
		ID:               id,
		TenantID:         in.TenantID,
		ParentID:         in.ParentID,
		Name:             in.Name,
		Code:             in.Code,
		Type:             in.Type,
		IsActive:         true,
		Hidden:           false,
		AccrualMethod:    in.AccrualMethod,
		FYStartMonth:     in.FYStartMonth,
		Address:          in.Address,
		Picture:          in.Picture,
		OrgUnitPath:      &unitPath,
		OrgLevel:         orgLevel,
		Settings:         in.Settings,
		Metadata:         in.Metadata,
		Version:          1,
		ValidationStatus: org.ValidationStatusPending,
		ValidationErrors: []any{},
	}

	var createdAt, updatedAt time.Time
	err := tx.QueryRow(ctx, qInsert,
		id, in.TenantID, in.ParentID, in.Name, in.Code, string(in.Type),
		in.AccrualMethod, in.FYStartMonth,
		in.Address, in.Picture, unitPath, orgLevel,
		in.Settings, in.Metadata,
	).Scan(&createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("org/pgstore: CreateUnit insert: %w", err)
	}

	unit.CreatedAt = createdAt
	unit.UpdatedAt = updatedAt
	return unit, nil
}

// insertPathsTx is the transaction-local variant of InsertPaths.
// Called from CreateUnit where we already hold a pgx.Tx.
func insertPathsTx(ctx context.Context, tx pgx.Tx, unit *org.Unit) error {
	if unit.ParentID == nil {
		const q = `
			INSERT INTO org_unit_paths (tenant_id, ancestor_id, descendant_id, depth)
			VALUES ($1, $2, $2, 0)
			ON CONFLICT (tenant_id, ancestor_id, descendant_id) DO NOTHING`
		if _, err := tx.Exec(ctx, q, unit.TenantID, unit.ID); err != nil {
			return fmt.Errorf("org/pgstore: insertPathsTx (root): %w", err)
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

	if _, err := tx.Exec(ctx, q, unit.TenantID, unit.ID, *unit.ParentID); err != nil {
		return fmt.Errorf("org/pgstore: insertPathsTx: %w", err)
	}
	return nil
}

// ReparentUnit moves unit to a new parent within the same tenant.
// Calls RebuildPaths internally; the DB trigger re-validates depth and cycles.
func (t *PgTree) ReparentUnit(ctx context.Context, tenantID, unitID uuid.UUID, newParentID *uuid.UUID) (*org.Unit, error) {
	tx, err := t.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("org/pgstore: ReparentUnit begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Update parent_id — triggers check_org_unit_hierarchy_depth.
	const qUpdate = `
		UPDATE org_units
		SET parent_id = $1, updated_at = NOW(), version = version + 1
		WHERE uuid = $2 AND tenant_id = $3 AND deleted_at IS NULL
		RETURNING uuid, tenant_id, parent_id, org_unit_path, org_level`

	unit := &org.Unit{}
	err = tx.QueryRow(ctx, qUpdate, newParentID, unitID, tenantID).
		Scan(&unit.ID, &unit.TenantID, &unit.ParentID, &unit.OrgUnitPath, &unit.OrgLevel)
	if err != nil {
		return nil, fmt.Errorf("org/pgstore: ReparentUnit update: %w", err)
	}

	if err := rebuildPathsTx(ctx, tx, unit); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("org/pgstore: ReparentUnit commit: %w", err)
	}
	return t.Unit(ctx, tenantID, unitID)
}

// SoftDeleteUnit marks the unit deleted_at = NOW(). Does NOT cascade to
// children — caller must decide whether to reparent or also delete descendants.
func (t *PgTree) SoftDeleteUnit(ctx context.Context, tenantID, unitID uuid.UUID) error {
	const q = `
		UPDATE org_units
		SET deleted_at = NOW(), updated_at = NOW(), version = version + 1
		WHERE uuid = $1 AND tenant_id = $2 AND deleted_at IS NULL`

	tag, err := t.pool.Exec(ctx, q, unitID, tenantID)
	if err != nil {
		return fmt.Errorf("org/pgstore: SoftDeleteUnit: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return org.ErrUnitNotFound
	}
	return nil
}
