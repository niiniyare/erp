// Package wire - Service layer providers
package wire

import (
	temporalclient "go.temporal.io/sdk/client"

	"awo/internal/platform/cache"
	"awo/internal/shared/logger"
	"awo/internal/shared/metrics"
	"awo/internal/shared/tracing"

	// Core services
	financeService "awo/internal/core/finance/service"
	"awo/internal/core/tenant"

	db "awo/db/sqlc"
)

// ============================================================================
// TENANT SERVICE
// ============================================================================

// NewTenantService creates a new tenant service
func NewTenantService(
	store db.Store,
	cache cache.Service,
	tracer tracing.Service,
	log logger.Logger,
) tenant.Service {
	return tenant.NewService(tenant.Dependencies{
		Store:  store,
		Cache:  cache,
		Tracer: tracer,
		Logger: log,
	})
}

// ============================================================================
// TENANT TEMPORAL INTEGRATION
// ============================================================================

// NewTenantTemporalIntegration creates the tenant Temporal integration
func NewTenantTemporalIntegration(
	tenantService tenant.Service,
	temporalClient temporalclient.Client,
	log logger.Logger,
) (*tenant.TemporalIntegration, error) {
	return tenant.NewTemporalIntegration(tenant.TemporalConfig{
		Service:        tenantService,
		TemporalClient: temporalClient,
		Logger:         log,
	})
}

// ============================================================================
// FINANCE SERVICES
// ============================================================================

// NewFinanceServices creates finance service collection
func NewFinanceServices(
	store db.Store,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.Service,
) *financeService.Services {
	// TODO: Implement proper finance services
	return nil
}
