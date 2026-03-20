package workflows

import (
	"time"

	"awo/internal/core/audit"
	"awo/internal/core/featureflag"
	"awo/internal/core/finance/domain"
	"awo/internal/core/finance/service"
	"awo/internal/core/iam"
	"awo/internal/core/notification"
	settingsService "awo/internal/core/settings/service"
	"awo/internal/platform/cache"
	loggerPkg "awo/internal/shared/logger"
	"awo/internal/shared/metrics"
	"awo/internal/shared/tracing"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// AccountWorkflows contains all account-related Temporal workflows
type AccountWorkflows struct {
	deps AccountWorkflowDeps
}

// AccountWorkflowDeps contains dependencies for account workflows
type AccountWorkflowDeps struct {
	Services            *service.Services
	IAMService          iam.Service
	AuditService        audit.Service
	FeatureFlagService  featureflag.Service
	SettingsService     settingsService.ConfigurationService
	NotificationService notification.NotificationService
	CacheService        cache.Service
	Logger              loggerPkg.Logger
	Metrics             metrics.MetricsProvider
	Tracer              tracing.Service
}

// NewAccountWorkflows creates new account workflows
func NewAccountWorkflows(deps AccountWorkflowDeps) *AccountWorkflows {
	return &AccountWorkflows{
		deps: deps,
	}
}

// AccountCreationWorkflow handles the complete account creation process
func (aw *AccountWorkflows) AccountCreationWorkflow(ctx workflow.Context, input domain.AccountCreationWorkflowInput) (*domain.AccountCreationWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting account creation workflow", "account_code", input.AccountCode)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 5,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
			InitialInterval: time.Second * 1,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	result := &domain.AccountCreationWorkflowResult{
		AccountCode: input.AccountCode,
		StartTime:   workflow.Now(ctx),
	}

	// Step 1: Validate account code
	var codeValidation domain.AccountCodeValidationResult
	err := workflow.ExecuteActivity(ctx, "ValidateAccountCode", domain.AccountCodeValidationInput{
		AccountCode: input.AccountCode,
		AccountType: input.AccountType,
	}).Get(ctx, &codeValidation)

	if err != nil || !codeValidation.IsValid {
		logger.Error("Account code validation failed", "error", err)
		result.Status = domain.AccountCreationStatusFailed
		result.Error = err.Error()
		result.ValidationErrors = codeValidation.ValidationErrors
		return result, nil
	}

	// Step 2: Validate account hierarchy
	var hierarchyValidation domain.AccountHierarchyValidationResult
	err = workflow.ExecuteActivity(ctx, "ValidateAccountHierarchy", domain.AccountHierarchyValidationInput{
		AccountCode:  input.AccountCode,
		ParentCode:   input.ParentAccountCode,
		AccountType:  input.AccountType,
		AccountClass: input.AccountClass,
	}).Get(ctx, &hierarchyValidation)

	if err != nil || !hierarchyValidation.IsValid {
		logger.Error("Account hierarchy validation failed", "error", err)
		result.Status = domain.AccountCreationStatusFailed
		result.Error = err.Error()
		result.ValidationErrors = hierarchyValidation.ValidationErrors
		return result, nil
	}

	// Step 3: Create account with validation
	var creationResult domain.AccountCreationResult
	err = workflow.ExecuteActivity(ctx, "CreateAccountWithValidation", domain.AccountCreationInput{
		AccountCode:        input.AccountCode,
		AccountName:        input.AccountName,
		AccountType:        input.AccountType,
		AccountClass:       input.AccountClass,
		ParentAccountCode:  input.ParentAccountCode,
		CurrencyCode:       input.CurrencyCode,
		Description:        input.Description,
		IsActive:           input.IsActive,
		AllowManualJournal: input.AllowManualJournal,
		CreatedBy:          input.CreatedBy,
	}).Get(ctx, &creationResult)
	if err != nil {
		logger.Error("Account creation failed", "error", err)
		result.Status = domain.AccountCreationStatusFailed
		result.Error = err.Error()
		return result, nil
	}

	result.AccountID = creationResult.AccountID
	result.Status = domain.AccountCreationStatusCompleted

	// Step 4: Send notification
	err = workflow.ExecuteActivity(ctx, "SendAccountCreationNotification", domain.AccountNotificationInput{
		AccountID:   creationResult.AccountID,
		AccountCode: input.AccountCode,
		Action:      "created",
		Recipients:  []string{input.CreatedBy},
	}).Get(ctx, nil)
	if err != nil {
		logger.Warn("Failed to send account creation notification", "error", err)
		// Don't fail workflow for notification errors
	}

	result.CompletedTime = workflow.Now(ctx)
	result.Duration = result.CompletedTime.Sub(result.StartTime)

	logger.Info("Account creation workflow completed",
		"account_code", input.AccountCode,
		"account_id", result.AccountID,
		"duration", result.Duration)

	return result, nil
}

// AccountClosureWorkflow handles the account closure process
func (aw *AccountWorkflows) AccountClosureWorkflow(ctx workflow.Context, input domain.AccountClosureWorkflowInput) (*domain.AccountClosureWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting account closure workflow", "account_id", input.AccountID)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 10,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
			InitialInterval: time.Second * 2,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	result := &domain.AccountClosureWorkflowResult{
		AccountID: input.AccountID,
		StartTime: workflow.Now(ctx),
	}

	// Step 1: Check account balance
	var balanceCheck domain.AccountBalanceCheckResult
	err := workflow.ExecuteActivity(ctx, "CheckAccountBalance", domain.AccountBalanceCheckInput{
		AccountID: input.AccountID,
	}).Get(ctx, &balanceCheck)
	if err != nil {
		logger.Error("Balance check failed", "error", err)
		result.Status = domain.AccountClosureStatusFailed
		result.Error = err.Error()
		return result, nil
	}

	if !balanceCheck.IsZero {
		logger.Error("Account has non-zero balance", "balance", balanceCheck.Balance)
		result.Status = domain.AccountClosureStatusFailed
		result.Error = "Cannot close account with non-zero balance"
		return result, nil
	}

	// Step 2: Check for dependent accounts
	var dependencyCheck domain.AccountDependencyCheckResult
	err = workflow.ExecuteActivity(ctx, "CheckAccountDependencies", domain.AccountDependencyCheckInput{
		AccountID: input.AccountID,
	}).Get(ctx, &dependencyCheck)
	if err != nil {
		logger.Error("Dependency check failed", "error", err)
		result.Status = domain.AccountClosureStatusFailed
		result.Error = err.Error()
		return result, nil
	}

	if dependencyCheck.HasDependencies {
		logger.Error("Account has dependencies", "dependencies", dependencyCheck.Dependencies)
		result.Status = domain.AccountClosureStatusFailed
		result.Error = "Cannot close account with dependencies"
		result.Dependencies = dependencyCheck.Dependencies
		return result, nil
	}

	// Step 3: Close account
	err = workflow.ExecuteActivity(ctx, "CloseAccount", domain.AccountClosureInput{
		AccountID:     input.AccountID,
		ClosureReason: input.ClosureReason,
		ClosedBy:      input.ClosedBy,
		ClosureDate:   input.ClosureDate,
	}).Get(ctx, nil)
	if err != nil {
		logger.Error("Account closure failed", "error", err)
		result.Status = domain.AccountClosureStatusFailed
		result.Error = err.Error()
		return result, nil
	}

	result.Status = domain.AccountClosureStatusCompleted
	result.CompletedTime = workflow.Now(ctx)
	result.Duration = result.CompletedTime.Sub(result.StartTime)

	logger.Info("Account closure workflow completed",
		"account_id", input.AccountID,
		"duration", result.Duration)

	return result, nil
}

// AccountReconciliationWorkflow handles account reconciliation processes
func (aw *AccountWorkflows) AccountReconciliationWorkflow(ctx workflow.Context, input domain.AccountReconciliationWorkflowInput) (*domain.AccountReconciliationWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting account reconciliation workflow", "account_id", input.AccountID, "period", input.ReconciliationPeriod)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 15,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 2,
			InitialInterval: time.Second * 5,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	result := &domain.AccountReconciliationWorkflowResult{
		AccountID: input.AccountID,
		Period:    input.ReconciliationPeriod,
		StartTime: workflow.Now(ctx),
	}

	// Step 1: Get account transactions for period
	var transactionData domain.AccountTransactionDataResult
	err := workflow.ExecuteActivity(ctx, "GetAccountTransactionData", domain.AccountTransactionDataInput{
		AccountID:  input.AccountID,
		PeriodFrom: input.PeriodFrom,
		PeriodTo:   input.PeriodTo,
	}).Get(ctx, &transactionData)
	if err != nil {
		logger.Error("Failed to get transaction data", "error", err)
		result.Status = domain.ReconciliationStatusFailed
		result.Error = err.Error()
		return result, nil
	}

	// Step 2: Calculate expected balance
	var balanceCalculation domain.BalanceCalculationResult
	err = workflow.ExecuteActivity(ctx, "CalculateExpectedBalance", domain.BalanceCalculationInput{
		AccountID:      input.AccountID,
		Transactions:   transactionData.Transactions,
		OpeningBalance: transactionData.OpeningBalance,
	}).Get(ctx, &balanceCalculation)
	if err != nil {
		logger.Error("Balance calculation failed", "error", err)
		result.Status = domain.ReconciliationStatusFailed
		result.Error = err.Error()
		return result, nil
	}

	// Step 3: Compare with actual balance
	var actualBalance domain.ActualBalanceResult
	err = workflow.ExecuteActivity(ctx, "GetActualAccountBalance", domain.ActualBalanceInput{
		AccountID: input.AccountID,
		AsOfDate:  input.PeriodTo,
	}).Get(ctx, &actualBalance)
	if err != nil {
		logger.Error("Failed to get actual balance", "error", err)
		result.Status = domain.ReconciliationStatusFailed
		result.Error = err.Error()
		return result, nil
	}

	// Step 4: Analyze discrepancies
	result.ExpectedBalance = balanceCalculation.ExpectedBalance
	result.ActualBalance = actualBalance.Balance
	result.Discrepancy = result.ActualBalance.Sub(result.ExpectedBalance)

	if result.Discrepancy.IsZero() {
		result.Status = domain.ReconciliationStatusReconciled
		logger.Info("Account reconciled successfully - no discrepancies")
	} else {
		result.Status = domain.ReconciliationStatusDiscrepancy
		logger.Warn("Account reconciliation found discrepancies",
			"expected", result.ExpectedBalance,
			"actual", result.ActualBalance,
			"discrepancy", result.Discrepancy)

		// Step 5: Generate reconciliation report
		var reportResult domain.ReconciliationReportResult
		err = workflow.ExecuteActivity(ctx, "GenerateReconciliationReport", domain.ReconciliationReportInput{
			AccountID:       input.AccountID,
			Period:          input.ReconciliationPeriod,
			ExpectedBalance: result.ExpectedBalance,
			ActualBalance:   result.ActualBalance,
			Discrepancy:     result.Discrepancy,
			Transactions:    transactionData.Transactions,
		}).Get(ctx, &reportResult)

		if err != nil {
			logger.Error("Failed to generate reconciliation report", "error", err)
		} else {
			result.ReportID = reportResult.ReportID
			result.ReportPath = reportResult.ReportPath
		}
	}

	result.CompletedTime = workflow.Now(ctx)
	result.Duration = result.CompletedTime.Sub(result.StartTime)

	logger.Info("Account reconciliation workflow completed",
		"account_id", input.AccountID,
		"status", result.Status,
		"discrepancy", result.Discrepancy,
		"duration", result.Duration)

	return result, nil
}
