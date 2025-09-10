package finance

import (
	"context"

	"github.com/google/uuid"

	"github.com/niiniyare/erp/internal/core/audit"
	"github.com/niiniyare/erp/internal/core/featureflag"
	"github.com/niiniyare/erp/internal/core/finance/service"
	"github.com/niiniyare/erp/internal/core/iam"
	settingsService "github.com/niiniyare/erp/internal/core/settings/service"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
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
	iamService         iam.Service
	auditService       audit.Service
	featureFlagService featureflag.Service
	settingsService    settingsService.ConfigurationService
	cacheService       cache.Service
	temporalClient     client.Client

	// Shared infrastructure
	logger  logger.Logger
	metrics metrics.MetricsProvider
	tracer  tracing.TracingService
}

// NewService creates a new unified Finance service instance
func NewService(
	services *service.Services,
	iamService iam.Service,
	auditService audit.Service,
	featureFlagService featureflag.Service,
	settingsService settingsService.ConfigurationService,
	cacheService cache.Service,
	temporalClient client.Client,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.TracingService,
) Service {
	return &financeService{
		accountService:          services.Account,
		transactionService:      services.Transaction,
		transactionEntryService: services.TransactionEntry,
		iamService:              iamService,
		auditService:            auditService,
		featureFlagService:      featureFlagService,
		settingsService:         settingsService,
		cacheService:            cacheService,
		temporalClient:          temporalClient,
		logger:                  logger,
		metrics:                 metrics,
		tracer:                  tracer,
	}
}

// Helper methods for getting tenant context
func (s *financeService) getCurrentTenantID(ctx context.Context) uuid.UUID {
	// Extract tenant ID from context (set by middleware)
	if tenantID, ok := ctx.Value("tenant_id").(uuid.UUID); ok {
		return tenantID
	}
	s.logger.WarnContext(ctx, "No tenant ID found in context")
	return uuid.Nil
}

func (s *financeService) getCurrentEntityID(ctx context.Context) uuid.UUID {
	// Extract entity ID from context (set by middleware)
	if entityID, ok := ctx.Value("entity_id").(uuid.UUID); ok {
		return entityID
	}
	s.logger.WarnContext(ctx, "No entity ID found in context")
	return uuid.Nil
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
