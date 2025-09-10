package finance

import (
	"context"
	"fmt"

	"github.com/niiniyare/erp/internal/core/audit"
	"github.com/niiniyare/erp/internal/core/featureflag"
	"github.com/niiniyare/erp/internal/core/finance/activities"
	"github.com/niiniyare/erp/internal/core/finance/service"
	"github.com/niiniyare/erp/internal/core/finance/workflows"
	"github.com/niiniyare/erp/internal/core/iam"
	"github.com/niiniyare/erp/internal/core/notification"
	settingsService "github.com/niiniyare/erp/internal/core/settings/service"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/platform/temporal"
	loggerPkg "github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.temporal.io/sdk/client"
)

// TemporalIntegration handles the registration and management of finance module
// activities and workflows with the Temporal platform
type TemporalIntegration struct {
	activityRegistry *activities.ActivityRegistry
	workflowRegistry *workflows.WorkflowRegistry
	temporalClient   client.Client
	logger           loggerPkg.Logger
}

// TemporalIntegrationConfig contains configuration for finance Temporal integration
type TemporalIntegrationConfig struct {
	// Finance services
	Services *service.Services

	// External service dependencies
	IAMService          iam.Service
	AuditService        audit.Service
	FeatureFlagService  featureflag.Service
	SettingsService     settingsService.ConfigurationService
	NotificationService notification.NotificationService
	CacheService        cache.Service

	// Infrastructure
	TemporalClient client.Client
	Logger         loggerPkg.Logger
	Metrics        metrics.MetricsProvider
	Tracer         tracing.TracingService
}

// NewTemporalIntegration creates a new Temporal integration for the finance module
func NewTemporalIntegration(config TemporalIntegrationConfig) (*TemporalIntegration, error) {
	if config.Services == nil {
		return nil, fmt.Errorf("finance services are required")
	}
	if config.TemporalClient == nil {
		return nil, fmt.Errorf("temporal client is required")
	}
	if config.Logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	// Create activity registry
	activityRegistry := activities.NewActivityRegistry(activities.ActivityDependencies{
		AccountService:          config.Services.Account,
		TransactionService:      config.Services.Transaction,
		TransactionEntryService: config.Services.TransactionEntry,
		IAMService:              config.IAMService,
		AuditService:            config.AuditService,
		FeatureFlagService:      config.FeatureFlagService,
		SettingsService:         config.SettingsService,
		NotificationService:     config.NotificationService,
		CacheService:            config.CacheService,
		Logger:                  config.Logger,
		Metrics:                 config.Metrics,
		Tracer:                  config.Tracer,
	})

	// Create workflow registry
	workflowRegistry := workflows.NewWorkflowRegistry(workflows.WorkflowDependencies{
		Services:            config.Services,
		IAMService:          config.IAMService,
		AuditService:        config.AuditService,
		FeatureFlagService:  config.FeatureFlagService,
		SettingsService:     config.SettingsService,
		NotificationService: config.NotificationService,
		CacheService:        config.CacheService,
		Logger:              config.Logger,
		Metrics:             config.Metrics,
		Tracer:              config.Tracer,
	})

	return &TemporalIntegration{
		activityRegistry: activityRegistry,
		workflowRegistry: workflowRegistry,
		temporalClient:   config.TemporalClient,
		logger:           config.Logger,
	}, nil
}

// RegisterWithPlatform registers finance module activities and workflows with the Temporal platform
func (ti *TemporalIntegration) RegisterWithPlatform(platform *temporal.Platform) error {
	if platform == nil {
		return fmt.Errorf("temporal platform is required")
	}

	// Get module registrar for finance
	registrar := platform.NewModuleRegistrar("finance")

	if !registrar.IsEnabled() {
		ti.logger.Info("Finance module not enabled in Temporal configuration, skipping registration")
		return nil
	}

	// Register activities
	activities := ti.getActivitiesForRegistration()
	registrar.RegisterActivities(activities)

	// Register workflows
	workflows := ti.getWorkflowsForRegistration()
	registrar.RegisterWorkflows(workflows)

	ti.logger.Info("✅ Finance module registered with Temporal platform", loggerPkg.Fields{
		"module":         "finance",
		"activity_count": len(activities),
		"workflow_count": len(workflows),
		"status":         "registered",
	})

	return nil
}

// getActivitiesForRegistration returns a map of activities for registration with Temporal
func (ti *TemporalIntegration) getActivitiesForRegistration() map[string]any {
	activities := make(map[string]any)

	// Account activities
	accountActivities := ti.activityRegistry.GetAccountActivities()
	activities["ValidateAccountCreationActivity"] = accountActivities.ValidateAccountCreationActivity
	activities["CreateAccountActivity"] = accountActivities.CreateAccountActivity
	activities["GetAccountActivity"] = accountActivities.GetAccountActivity
	activities["UpdateAccountActivity"] = accountActivities.UpdateAccountActivity
	activities["DeactivateAccountActivity"] = accountActivities.DeactivateAccountActivity
	activities["ValidateAccountHierarchyActivity"] = accountActivities.ValidateAccountHierarchyActivity
	activities["CheckAccountPermissionsActivity"] = accountActivities.CheckAccountPermissionsActivity
	activities["CacheAccountActivity"] = accountActivities.CacheAccountActivity
	activities["InvalidateAccountCacheActivity"] = accountActivities.InvalidateAccountCacheActivity

	// Transaction activities
	transactionActivities := ti.activityRegistry.GetTransactionActivities()
	activities["ValidateTransactionActivity"] = transactionActivities.ValidateTransactionActivity
	activities["CreateTransactionActivity"] = transactionActivities.CreateTransactionActivity
	activities["ProcessTransactionEntriesActivity"] = transactionActivities.ProcessTransactionEntriesActivity
	activities["PostTransactionActivity"] = transactionActivities.PostTransactionActivity
	activities["ReverseTransactionActivity"] = transactionActivities.ReverseTransactionActivity
	activities["ValidateDoubleEntryActivity"] = transactionActivities.ValidateDoubleEntryActivity
	activities["CheckTransactionPermissionsActivity"] = transactionActivities.CheckTransactionPermissionsActivity
	activities["UpdateTransactionStatusActivity"] = transactionActivities.UpdateTransactionStatusActivity
	activities["CacheTransactionActivity"] = transactionActivities.CacheTransactionActivity
	activities["ProcessBulkTransactionsActivity"] = transactionActivities.ProcessBulkTransactionsActivity

	// Validation activities
	validationActivities := ti.activityRegistry.GetValidationActivities()
	activities["ValidateBusinessRulesActivity"] = validationActivities.ValidateBusinessRulesActivity
	activities["ValidateAccountingPeriodActivity"] = validationActivities.ValidateAccountingPeriodActivity
	activities["ValidateExchangeRateActivity"] = validationActivities.ValidateExchangeRateActivity
	activities["ValidateBudgetConstraintsActivity"] = validationActivities.ValidateBudgetConstraintsActivity
	activities["ValidateApprovalLimitsActivity"] = validationActivities.ValidateApprovalLimitsActivity
	activities["ValidateComplianceRulesActivity"] = validationActivities.ValidateComplianceRulesActivity

	// Integration activities
	integrationActivities := ti.activityRegistry.GetIntegrationActivities()
	activities["CheckFeatureFlagActivity"] = integrationActivities.CheckFeatureFlagActivity
	activities["GetSettingsActivity"] = integrationActivities.GetSettingsActivity
	activities["ValidateUserPermissionsActivity"] = integrationActivities.ValidateUserPermissionsActivity
	activities["LogAuditEventActivity"] = integrationActivities.LogAuditEventActivity
	activities["InvalidateCacheActivity"] = integrationActivities.InvalidateCacheActivity
	activities["SetCacheActivity"] = integrationActivities.SetCacheActivity
	activities["GetUserContextActivity"] = integrationActivities.GetUserContextActivity
	activities["ValidateEntityAccessActivity"] = integrationActivities.ValidateEntityAccessActivity

	// Notification activities
	notificationActivities := ti.activityRegistry.GetNotificationActivities()
	activities["SendAccountCreatedNotificationActivity"] = notificationActivities.SendAccountCreatedNotificationActivity
	activities["SendTransactionPostedNotificationActivity"] = notificationActivities.SendTransactionPostedNotificationActivity
	activities["SendApprovalRequestNotificationActivity"] = notificationActivities.SendApprovalRequestNotificationActivity
	activities["SendBudgetExceededNotificationActivity"] = notificationActivities.SendBudgetExceededNotificationActivity
	activities["SendComplianceAlertNotificationActivity"] = notificationActivities.SendComplianceAlertNotificationActivity
	activities["SendPeriodClosingNotificationActivity"] = notificationActivities.SendPeriodClosingNotificationActivity
	activities["SendErrorNotificationActivity"] = notificationActivities.SendErrorNotificationActivity
	activities["SendBulkOperationCompletedNotificationActivity"] = notificationActivities.SendBulkOperationCompletedNotificationActivity

	return activities
}

// getWorkflowsForRegistration returns a map of workflows for registration with Temporal
func (ti *TemporalIntegration) getWorkflowsForRegistration() map[string]any {
	workflows := make(map[string]any)

	// Get workflow registry
	workflowReg := ti.workflowRegistry

	// Transaction workflows
	workflows["TransactionApprovalWorkflow"] = workflowReg.GetTransactionApprovalWorkflow()
	workflows["TransactionProcessingWorkflow"] = workflowReg.GetTransactionProcessingWorkflow()
	workflows["TransactionReversalWorkflow"] = workflowReg.GetTransactionReversalWorkflow()
	workflows["BulkTransactionWorkflow"] = workflowReg.GetBulkTransactionWorkflow()

	// Account workflows
	workflows["AccountCreationWorkflow"] = workflowReg.GetAccountCreationWorkflow()
	workflows["AccountClosureWorkflow"] = workflowReg.GetAccountClosureWorkflow()
	workflows["AccountReconciliationWorkflow"] = workflowReg.GetAccountReconciliationWorkflow()

	// Compliance workflows
	workflows["ComplianceAuditWorkflow"] = workflowReg.GetComplianceAuditWorkflow()
	workflows["FraudDetectionWorkflow"] = workflowReg.GetFraudDetectionWorkflow()

	// Periodic workflows
	workflows["MonthEndClosingWorkflow"] = workflowReg.GetMonthEndClosingWorkflow()
	workflows["YearEndClosingWorkflow"] = workflowReg.GetYearEndClosingWorkflow()

	return workflows
}

// GetTemporalClient returns the Temporal client for direct workflow operations
func (ti *TemporalIntegration) GetTemporalClient() client.Client {
	return ti.temporalClient
}

// GetActivityRegistry returns the activity registry for manual registration
func (ti *TemporalIntegration) GetActivityRegistry() *activities.ActivityRegistry {
	return ti.activityRegistry
}

// GetWorkflowRegistry returns the workflow registry for manual registration
func (ti *TemporalIntegration) GetWorkflowRegistry() *workflows.WorkflowRegistry {
	return ti.workflowRegistry
}

// StartWorkflow provides a convenience method to start finance workflows
func (ti *TemporalIntegration) StartWorkflow(ctx context.Context, workflowType string, input any, options ...client.StartWorkflowOptions) (client.WorkflowRun, error) {
	// Default task queue for finance workflows
	defaultOptions := client.StartWorkflowOptions{
		TaskQueue: "finance.standard",
	}

	// Merge options
	if len(options) > 0 {
		opts := options[0]
		if opts.TaskQueue != "" {
			defaultOptions.TaskQueue = opts.TaskQueue
		}
		if opts.ID != "" {
			defaultOptions.ID = opts.ID
		}
		if opts.WorkflowExecutionTimeout != 0 {
			defaultOptions.WorkflowExecutionTimeout = opts.WorkflowExecutionTimeout
		}
	}

	return ti.temporalClient.ExecuteWorkflow(ctx, defaultOptions, workflowType, input)
}
