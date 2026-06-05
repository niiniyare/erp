// Package wire - Service layer providers
package wire

import (
	temporalclient "go.temporal.io/sdk/client"

	"awo.so/internal/platform/cache"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"

	// Core services
	"awo.so/internal/core/contracts"
	financeRepo "awo.so/internal/core/finance/repository"
	financeService "awo.so/internal/core/finance/service"
	"awo.so/internal/core/iam"
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
// CONTRACTS SERVICE
// ============================================================================

// NewContractService wires the contracts service with its repository and cache.
func NewContractService(
	store db.Store,
	cacheSvc cache.Service,
	tracer tracing.Service,
	log logger.Logger,
) contracts.Service {
	return contracts.NewService(contracts.Dependencies{
		Store:  store,
		Cache:  cacheSvc,
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
	authzSvc iam.AuthzService,
) (*tenant.TemporalIntegration, error) {
	return tenant.NewTemporalIntegration(tenant.TemporalConfig{
		Service:        tenantService,
		TemporalClient: temporalClient,
		Logger:         log,
		AuthzService:   authzSvc,
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
	transactionRepo := financeRepo.NewTransactionRepository(store, tracer, log, met)
	periodRepo := financeRepo.NewPeriodRepository(store, tracer, log, met, cacheService)
	exchangeRateRepo := financeRepo.NewExchangeRateRepository(store, tracer)
	currencyRepo := financeRepo.NewCurrencyRepository(store, tracer)
	costCenterRepo := financeRepo.NewCostCenterRepository(store, tracer)
	budgetRepo := financeRepo.NewBudgetRepository(store, tracer)
	taxRepo := financeRepo.NewTaxRepository(store, tracer)
	reconciliationRepo := financeRepo.NewReconciliationRepository(store, tracer, log, met)

	return financeService.NewServices(financeService.Dependencies{
		AccountRepo:        accountRepo,
		TransactionRepo:    transactionRepo,
		PeriodRepo:         periodRepo,
		ExchangeRateRepo:   exchangeRateRepo,
		CurrencyRepo:       currencyRepo,
		CostCenterRepo:     costCenterRepo,
		BudgetRepo:         budgetRepo,
		TaxRepo:            taxRepo,
		ReconciliationRepo: reconciliationRepo,
		Tracing:            tracer,
		Metrics:            met,
	})
}
