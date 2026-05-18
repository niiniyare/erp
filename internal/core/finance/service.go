package finance

import (
	"awo.so/internal/core/audit"
	"awo.so/internal/core/finance/service"
	settingsService "awo.so/internal/core/settings/service"
	"awo.so/internal/platform/cache"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
	"go.temporal.io/sdk/client"
)

// Service defines the unified Finance service interface following the IAM pattern
type Service interface {
	// Get individual service instances
	Account() service.AccountService
	Transaction() service.TransactionService
	TransactionEntry() service.TransactionEntryService
}

// financeService implements the unified Finance service
type financeService struct {
	// Domain services
	accountService          service.AccountService
	transactionService      service.TransactionService
	transactionEntryService service.TransactionEntryService

	// External dependencies
	auditService    audit.Service
	settingsService settingsService.ConfigurationService
	cacheService    cache.Service
	temporalClient  client.Client

	// Shared infrastructure
	logger  logger.Logger
	metrics metrics.MetricsProvider
	tracer  tracing.Service
}

// NewService creates a new unified Finance service instance
func NewService(
	services *service.Services,
	auditService audit.Service,
	settingsService settingsService.ConfigurationService,
	cacheService cache.Service,
	temporalClient client.Client,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.Service,
) Service {
	return &financeService{
		accountService:          services.Account,
		transactionService:      services.Transaction,
		transactionEntryService: services.TransactionEntry,
		auditService:            auditService,
		settingsService:         settingsService,
		cacheService:            cacheService,
		temporalClient:          temporalClient,
		logger:                  logger,
		metrics:                 metrics,
		tracer:                  tracer,
	}
}

// Service access methods

func (s *financeService) Account() service.AccountService {
	return s.accountService
}

func (s *financeService) Transaction() service.TransactionService {
	return s.transactionService
}

func (s *financeService) TransactionEntry() service.TransactionEntryService {
	return s.transactionEntryService
}
