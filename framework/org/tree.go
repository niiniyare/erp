// Package org is a compatibility shim. New code should import awo.so/framework/platform/org.
package org

import (
	"context"

	"github.com/google/uuid"

	platformorg "awo.so/framework/platform/org"
)

// Tree re-exported from platform/org.
type Tree = platformorg.Tree

var (
	ErrUnitNotFound     = platformorg.ErrUnitNotFound
	ErrCycleDetected    = platformorg.ErrCycleDetected
	ErrMaxDepthExceeded = platformorg.ErrMaxDepthExceeded
)

// AssertAncestor re-exported from platform/org.
func AssertAncestor(ctx context.Context, tree Tree, tenantID, viewerUnitID, recordUnitID uuid.UUID) error {
	return platformorg.AssertAncestor(ctx, tree, tenantID, viewerUnitID, recordUnitID)
}
