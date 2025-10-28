// Package wire - Repository layer providers
package wire

import (
	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/shared/tracing"
	
	// Core domain repositories
	"github.com/niiniyare/erp/internal/core/tenant"
)

// ============================================================================
// TENANT REPOSITORIES
// ============================================================================

// NewTenantRepository creates a new tenant repository
func NewTenantRepository(
	store db.Store,
	tracer tracing.TracingService,
) tenant.Repository {
	return tenant.NewRepository(store, tracer)
}