// Package wire - Service layer providers
package wire

import (
	temporalclient "go.temporal.io/sdk/client"

	"awo.so/internal/platform/cache"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"

	// Core services
	financeRepo "awo.so/internal/core/finance/repository"
	financeService "awo.so/internal/core/finance/service"
	"awo.so/internal/core/tenant"

	db "awo.so/db/sqlc"
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
	cacheService cache.Service,
	log logger.Logger,
	met metrics.MetricsProvider,
	tracer tracing.Service,
) *financeService.Services {
	accountRepo := financeRepo.NewAccountsRepository(store, cacheService, tracer)
	transactionRepo := financeRepo.NewTransactionRepository(store, tracer)
	periodRepo := financeRepo.NewPeriodRepository(store, tracer)
	exchangeRateRepo := financeRepo.NewExchangeRateRepository(store, tracer)
	currencyRepo := financeRepo.NewCurrencyRepository(store, tracer)
	costCenterRepo := financeRepo.NewCostCenterRepository(store, tracer)

	_ = log // available for future use

	return financeService.NewServices(financeService.Dependencies{
		AccountRepo:      accountRepo,
		TransactionRepo:  transactionRepo,
		PeriodRepo:       periodRepo,
		ExchangeRateRepo: exchangeRateRepo,
		CurrencyRepo:     currencyRepo,
		CostCenterRepo:   costCenterRepo,
		Tracing:          tracer,
		Metrics:          met,
	})
}
