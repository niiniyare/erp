package activities

import (
	"awo/internal/core/audit"
	"awo/internal/core/featureflag"
	"awo/internal/core/finance/service"
	"awo/internal/core/iam"
	"awo/internal/core/notification"
	settingsService "awo/internal/core/settings/service"
	"awo/internal/platform/cache"
	"awo/internal/shared/logger"
	"awo/internal/shared/metrics"
	"awo/internal/shared/tracing"
	"go.temporal.io/sdk/worker"
)

// ActivityRegistry manages all finance-related Temporal activities
type ActivityRegistry struct {
	accountActivities      *AccountActivities
	transactionActivities  *TransactionActivities
	validationActivities   *ValidationActivities
	integrationActivities  *IntegrationActivities
	notificationActivities *NotificationActivities
}

// ActivityDependencies contains all required dependencies for finance activities
type ActivityDependencies struct {
	// Finance services
	AccountService          service.AccountService
	TransactionService      service.TransactionService
	TransactionEntryService service.TransactionEntryService

	// Platform service dependencies
	IAMService          iam.Service
	AuditService        audit.Service
	FeatureFlagService  featureflag.Service
	SettingsService     settingsService.ConfigurationService
	NotificationService notification.NotificationService
	CacheService        cache.Service

	// Infrastructure
	Logger  logger.Logger
	Metrics metrics.MetricsProvider
	Tracer  tracing.Service
}

// NewActivityRegistry creates a new activity registry with all dependencies
func NewActivityRegistry(deps ActivityDependencies) *ActivityRegistry {
	return &ActivityRegistry{
		accountActivities: NewAccountActivities(ActivityDeps{
			AccountService:     deps.AccountService,
			IAMService:         deps.IAMService,
			AuditService:       deps.AuditService,
			FeatureFlagService: deps.FeatureFlagService,
			SettingsService:    deps.SettingsService,
			CacheService:       deps.CacheService,
			Logger:             deps.Logger,
			Metrics:            deps.Metrics,
			Tracer:             deps.Tracer,
		}),
		transactionActivities: NewTransactionActivities(TransactionActivityDeps{
			TransactionService:      deps.TransactionService,
			TransactionEntryService: deps.TransactionEntryService,
			AccountService:          deps.AccountService,
			IAMService:              deps.IAMService,
			AuditService:            deps.AuditService,
			FeatureFlagService:      deps.FeatureFlagService,
			SettingsService:         deps.SettingsService,
			CacheService:            deps.CacheService,
			Logger:                  deps.Logger,
			Metrics:                 deps.Metrics,
			Tracer:                  deps.Tracer,
		}),
		validationActivities: NewValidationActivities(ValidationActivityDeps{
			AccountService:     deps.AccountService,
			TransactionService: deps.TransactionService,
			IAMService:         deps.IAMService,
			SettingsService:    deps.SettingsService,
			Logger:             deps.Logger,
			Metrics:            deps.Metrics,
			Tracer:             deps.Tracer,
		}),
		integrationActivities: NewIntegrationActivities(IntegrationActivityDeps{
			IAMService:         deps.IAMService,
			AuditService:       deps.AuditService,
			FeatureFlagService: deps.FeatureFlagService,
			SettingsService:    deps.SettingsService,
			CacheService:       deps.CacheService,
			Logger:             deps.Logger,
			Metrics:            deps.Metrics,
			Tracer:             deps.Tracer,
		}),
		notificationActivities: NewNotificationActivities(NotificationActivityDeps{
			NotificationService: deps.NotificationService,
			AuditService:        deps.AuditService,
			SettingsService:     deps.SettingsService,
			Logger:              deps.Logger,
			Metrics:             deps.Metrics,
			Tracer:              deps.Tracer,
		}),
	}
}

// RegisterWith registers all finance activities with a Temporal worker
func (r *ActivityRegistry) RegisterWith(w worker.Worker) {
	// Register account activities
	r.accountActivities.RegisterWith(w)

	// Register transaction activities
	r.transactionActivities.RegisterWith(w)

	// Register validation activities
	r.validationActivities.RegisterWith(w)

	// Register integration activities
	r.integrationActivities.RegisterWith(w)

	// Register notification activities
	r.notificationActivities.RegisterWith(w)
}

// GetAccountActivities returns account activities for workflow use
func (r *ActivityRegistry) GetAccountActivities() *AccountActivities {
	return r.accountActivities
}

// GetTransactionActivities returns transaction activities for workflow use
func (r *ActivityRegistry) GetTransactionActivities() *TransactionActivities {
	return r.transactionActivities
}

// GetValidationActivities returns validation activities for workflow use
func (r *ActivityRegistry) GetValidationActivities() *ValidationActivities {
	return r.validationActivities
}

// GetIntegrationActivities returns integration activities for workflow use
func (r *ActivityRegistry) GetIntegrationActivities() *IntegrationActivities {
	return r.integrationActivities
}

// GetNotificationActivities returns notification activities for workflow use
func (r *ActivityRegistry) GetNotificationActivities() *NotificationActivities {
	return r.notificationActivities
}
