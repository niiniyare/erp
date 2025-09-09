package workflows

import (
	"github.com/niiniyare/erp/internal/core/audit"
	"github.com/niiniyare/erp/internal/core/featureflag"
	"github.com/niiniyare/erp/internal/core/finance/service"
	"github.com/niiniyare/erp/internal/core/iam"
	"github.com/niiniyare/erp/internal/core/notification"
	settingsService "github.com/niiniyare/erp/internal/core/settings/service"
	"github.com/niiniyare/erp/internal/platform/cache"
	loggerPkg "github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
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
	Tracer  tracing.TracingService
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