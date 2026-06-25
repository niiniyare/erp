package org

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// Tree is the read/write interface for the org_units hierarchy.
//
// Implementations back containment checks against either:
//   - the closure table (org_unit_paths) for O(1) depth lookups, or
//   - the materialized path column (org_unit_path) for O(1) LIKE subtree checks.
//
// The application layer is responsible for calling InsertPaths on every new unit
// and RebuildPaths on every reparent operation. The DB trigger enforces max depth
// (8 levels) and cycle detection — the Go layer must not bypass these.
type Tree interface {
	// IsAncestorOrEqual returns true when ancestor is at or above node in the tree.
	// Implemented as:
	//   SELECT 1 FROM org_unit_paths
	//   WHERE tenant_id = $1 AND ancestor_id = $2 AND descendant_id = $3
	//
	// Returns true when ancestorID == nodeID (self is always an ancestor of self, depth 0).
	IsAncestorOrEqual(ctx context.Context, tenantID, ancestorID, nodeID uuid.UUID) (bool, error)

	// Ancestors returns the chain of ancestor IDs from nodeID up to the root,
	// ordered by depth ASC (depth 0 = self, depth 1 = parent, …).
	// Implemented against org_unit_paths WHERE descendant_id = nodeID.
	Ancestors(ctx context.Context, tenantID, nodeID uuid.UUID) ([]uuid.UUID, error)

	// Descendants returns all descendant IDs under nodeID (inclusive of nodeID itself).
	// Implemented against org_unit_paths WHERE ancestor_id = nodeID.
	Descendants(ctx context.Context, tenantID, nodeID uuid.UUID) ([]uuid.UUID, error)

	// SubtreePath returns the materialized-path LIKE pattern for nodeID's subtree.
	// Returns ("", ErrUnitNotFound) if the unit has no org_unit_path set.
	// Example result: "/3fa85f64-…/%" — use with:
	//   WHERE org_unit_id IN (SELECT uuid FROM org_units WHERE org_unit_path LIKE $1)
	SubtreePath(ctx context.Context, tenantID, nodeID uuid.UUID) (string, error)

	// Unit fetches a single OrgUnit by its UUID. Returns ErrUnitNotFound when absent
	// or soft-deleted.
	Unit(ctx context.Context, tenantID, unitID uuid.UUID) (*Unit, error)

	// InsertPaths inserts closure table rows for a newly created unit.
	// Must be called inside the same transaction that inserts the org_units row.
	// Inserts:
	//   - self-reference row (ancestor = descendant = unit.ID, depth = 0)
	//   - one row per ancestor from parent up to root, incrementing depth
	InsertPaths(ctx context.Context, unit *Unit) error

	// RebuildPaths rebuilds the closure table rows for unit and all its descendants
	// after a reparent operation. Must be called inside the reparent transaction.
	// Also updates org_unit_path and org_level on all affected rows.
	RebuildPaths(ctx context.Context, unit *Unit) error
}

// Sentinel errors returned by Tree methods.
var (
	// ErrUnitNotFound is returned when the requested org unit does not exist
	// in the tenant or has been soft-deleted.
	ErrUnitNotFound = errors.New("org: unit not found")

	// ErrCycleDetected is returned when an insertion or reparent would create
	// a circular reference in the hierarchy.
	ErrCycleDetected = errors.New("org: circular reference in hierarchy")

	// ErrMaxDepthExceeded is returned when an insertion would exceed the
	// maximum allowed hierarchy depth (8 levels).
	ErrMaxDepthExceeded = errors.New("org: hierarchy depth exceeds maximum of 8 levels")
)

// AssertAncestor returns nil if ancestorID is an ancestor-or-equal of nodeID according
// to tree, or ErrScopeInsufficient if not. Use this in privacy policy functions.
func AssertAncestor(
	ctx context.Context,
	tree Tree,
	tenantID, viewerUnitID, recordUnitID uuid.UUID,
) error {
	if viewerUnitID == recordUnitID {
		return nil
	}
	ok, err := tree.IsAncestorOrEqual(ctx, tenantID, viewerUnitID, recordUnitID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrScopeInsufficient
	}
	return nil
}
