package workflows

import (
	"awo/internal/core/audit"
	"awo/internal/core/featureflag"
	"awo/internal/core/finance/service"
	"awo/internal/core/iam"
	"awo/internal/core/notification"
	settingsService "awo/internal/core/settings/service"
	"awo/internal/platform/cache"
	loggerPkg "awo/internal/shared/logger"
	"awo/internal/shared/metrics"
	"awo/internal/shared/tracing"
)

// WorkflowRegistry manages all finance-related Temporal workflows
type WorkflowRegistry struct {
	transactionWorkflows *TransactionWorkflows
	accountWorkflows     *AccountWorkflows
	complianceWorkflows  *ComplianceWorkflows
	periodicWorkflows    *PeriodicWorkflows
}

// WorkflowDependencies contains all required dependencies for finance workflows
type WorkflowDependencies struct {
	// Finance services
	Services *service.Services

	// Platform service dependencies
	IAMService          iam.Service
	AuditService        audit.Service
	FeatureFlagService  featureflag.Service
	SettingsService     settingsService.ConfigurationService
	NotificationService notification.NotificationService
	CacheService        cache.Service

	// Infrastructure
	Logger  loggerPkg.Logger
	Metrics metrics.MetricsProvider
	Tracer  tracing.Service
}

// NewWorkflowRegistry creates a new workflow registry with all dependencies
func NewWorkflowRegistry(deps WorkflowDependencies) *WorkflowRegistry {
	return &WorkflowRegistry{
		transactionWorkflows: NewTransactionWorkflows(TransactionWorkflowDeps{
			Services:            deps.Services,
			IAMService:          deps.IAMService,
			AuditService:        deps.AuditService,
			FeatureFlagService:  deps.FeatureFlagService,
			SettingsService:     deps.SettingsService,
			NotificationService: deps.NotificationService,
			CacheService:        deps.CacheService,
			Logger:              deps.Logger,
			Metrics:             deps.Metrics,
			Tracer:              deps.Tracer,
		}),
		accountWorkflows: NewAccountWorkflows(AccountWorkflowDeps{
			Services:            deps.Services,
			IAMService:          deps.IAMService,
			AuditService:        deps.AuditService,
			FeatureFlagService:  deps.FeatureFlagService,
			SettingsService:     deps.SettingsService,
			NotificationService: deps.NotificationService,
			CacheService:        deps.CacheService,
			Logger:              deps.Logger,
			Metrics:             deps.Metrics,
			Tracer:              deps.Tracer,
		}),
		complianceWorkflows: NewComplianceWorkflows(ComplianceWorkflowDeps{
			Services:            deps.Services,
			IAMService:          deps.IAMService,
			AuditService:        deps.AuditService,
			FeatureFlagService:  deps.FeatureFlagService,
			SettingsService:     deps.SettingsService,
			NotificationService: deps.NotificationService,
			CacheService:        deps.CacheService,
			Logger:              deps.Logger,
			Metrics:             deps.Metrics,
			Tracer:              deps.Tracer,
		}),
		periodicWorkflows: NewPeriodicWorkflows(PeriodicWorkflowDeps{
			Services:            deps.Services,
			IAMService:          deps.IAMService,
			AuditService:        deps.AuditService,
			FeatureFlagService:  deps.FeatureFlagService,
			SettingsService:     deps.SettingsService,
			NotificationService: deps.NotificationService,
			CacheService:        deps.CacheService,
			Logger:              deps.Logger,
			Metrics:             deps.Metrics,
			Tracer:              deps.Tracer,
		}),
	}
}

// Transaction workflow getters
func (wr *WorkflowRegistry) GetTransactionApprovalWorkflow() any {
	return wr.transactionWorkflows.TransactionApprovalWorkflow
}

func (wr *WorkflowRegistry) GetTransactionProcessingWorkflow() any {
	return wr.transactionWorkflows.TransactionProcessingWorkflow
}

func (wr *WorkflowRegistry) GetTransactionReversalWorkflow() any {
	return wr.transactionWorkflows.TransactionReversalWorkflow
}

func (wr *WorkflowRegistry) GetBulkTransactionWorkflow() any {
	return wr.transactionWorkflows.BulkTransactionWorkflow
}

// Account workflow getters
func (wr *WorkflowRegistry) GetAccountCreationWorkflow() any {
	return wr.accountWorkflows.AccountCreationWorkflow
}

func (wr *WorkflowRegistry) GetAccountClosureWorkflow() any {
	return wr.accountWorkflows.AccountClosureWorkflow
}

func (wr *WorkflowRegistry) GetAccountReconciliationWorkflow() any {
	return wr.accountWorkflows.AccountReconciliationWorkflow
}

// Compliance workflow getters
func (wr *WorkflowRegistry) GetComplianceAuditWorkflow() any {
	return wr.complianceWorkflows.ComplianceAuditWorkflow
}

func (wr *WorkflowRegistry) GetFraudDetectionWorkflow() any {
	return wr.complianceWorkflows.FraudDetectionWorkflow
}

// Periodic workflow getters
func (wr *WorkflowRegistry) GetMonthEndClosingWorkflow() any {
	return wr.periodicWorkflows.MonthEndClosingWorkflow
}

func (wr *WorkflowRegistry) GetYearEndClosingWorkflow() any {
	return wr.periodicWorkflows.YearEndClosingWorkflow
}
