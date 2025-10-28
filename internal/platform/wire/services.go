// Package wire - Service layer providers
package wire

import (
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/shared/tracing"
	
	// Core services
	"github.com/niiniyare/erp/internal/core/tenant"
)

// ============================================================================
// TENANT SERVICE
// ============================================================================

// NewTenantService creates a new tenant service
func NewTenantService(
	repo tenant.Repository,
	cache cache.Service,
	tracer tracing.Service,
) tenant.Service {
	return tenant.NewService(repo, cache, tracer)
}