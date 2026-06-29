package org

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// Tree is the read/write interface for the org_units hierarchy.
type Tree interface {
	IsAncestorOrEqual(ctx context.Context, tenantID, ancestorID, nodeID uuid.UUID) (bool, error)
	Ancestors(ctx context.Context, tenantID, nodeID uuid.UUID) ([]uuid.UUID, error)
	Descendants(ctx context.Context, tenantID, nodeID uuid.UUID) ([]uuid.UUID, error)
	SubtreePath(ctx context.Context, tenantID, nodeID uuid.UUID) (string, error)
	Unit(ctx context.Context, tenantID, unitID uuid.UUID) (*Unit, error)
	InsertPaths(ctx context.Context, unit *Unit) error
	RebuildPaths(ctx context.Context, unit *Unit) error
}

var (
	ErrUnitNotFound     = errors.New("org: unit not found")
	ErrCycleDetected    = errors.New("org: circular reference in hierarchy")
	ErrMaxDepthExceeded = errors.New("org: hierarchy depth exceeds maximum of 8 levels")
)

// AssertAncestor returns nil if ancestorID is an ancestor-or-equal of nodeID.
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
