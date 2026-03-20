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

// PeriodicWorkflows contains all periodic/scheduled Temporal workflows
type PeriodicWorkflows struct {
	deps PeriodicWorkflowDeps
}

// PeriodicWorkflowDeps contains dependencies for periodic workflows
type PeriodicWorkflowDeps struct {
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

// NewPeriodicWorkflows creates new periodic workflows
func NewPeriodicWorkflows(deps PeriodicWorkflowDeps) *PeriodicWorkflows {
	return &PeriodicWorkflows{
		deps: deps,
	}
}

// MonthEndClosingWorkflow handles month-end closing processes
func (pw *PeriodicWorkflows) MonthEndClosingWorkflow(ctx workflow.Context, input domain.MonthEndClosingWorkflowInput) (*domain.MonthEndClosingWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting month-end closing workflow", "period", input.ClosingPeriod)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Hour * 4, // Month-end can take several hours
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 2,
			InitialInterval: time.Minute * 5,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	result := &domain.MonthEndClosingWorkflowResult{
		ClosingPeriod: input.ClosingPeriod,
		StartTime:     workflow.Now(ctx),
		Steps:         make(map[string]domain.ClosingStepResult),
	}

	// Step 1: Pre-closing validation
	logger.Info("Step 1: Pre-closing validation")
	var preClosingValidation domain.PreClosingValidationResult
	err := workflow.ExecuteActivity(ctx, "ValidatePreClosingConditions", domain.PreClosingValidationInput{
		ClosingPeriod:   input.ClosingPeriod,
		TenantID:        input.TenantID,
		ValidationRules: input.ValidationRules,
	}).Get(ctx, &preClosingValidation)

	if err != nil || !preClosingValidation.IsValid {
		logger.Error("Pre-closing validation failed", "error", err)
		result.Status = domain.ClosingStatusFailed
		result.Error = err.Error()
		result.ValidationErrors = preClosingValidation.ValidationErrors
		return result, nil
	}

	result.Steps["pre_validation"] = domain.ClosingStepResult{
		StepName:    "Pre-closing Validation",
		Status:      domain.StepStatusCompleted,
		CompletedAt: workflow.Now(ctx),
	}

	// Step 2: Reconcile all accounts
	logger.Info("Step 2: Account reconciliation")
	var reconciliationResult domain.AccountReconciliationBatchResult
	err = workflow.ExecuteActivity(ctx, "ReconcileAllAccounts", domain.AccountReconciliationBatchInput{
		ClosingPeriod: input.ClosingPeriod,
		TenantID:      input.TenantID,
		AccountFilter: input.AccountFilter,
	}).Get(ctx, &reconciliationResult)
	if err != nil {
		logger.Error("Account reconciliation failed", "error", err)
		result.Status = domain.ClosingStatusFailed
		result.Error = err.Error()
		return result, nil
	}

	result.Steps["reconciliation"] = domain.ClosingStepResult{
		StepName:         "Account Reconciliation",
		Status:           domain.StepStatusCompleted,
		CompletedAt:      workflow.Now(ctx),
		ProcessedCount:   reconciliationResult.ProcessedAccounts,
		DiscrepancyCount: reconciliationResult.DiscrepancyCount,
	}

	// Step 3: Generate adjusting entries
	logger.Info("Step 3: Adjusting entries")
	var adjustingEntries domain.AdjustingEntriesResult
	err = workflow.ExecuteActivity(ctx, "GenerateAdjustingEntries", domain.AdjustingEntriesInput{
		ClosingPeriod:         input.ClosingPeriod,
		TenantID:              input.TenantID,
		ReconciliationResults: reconciliationResult.Results,
		AdjustmentRules:       input.AdjustmentRules,
	}).Get(ctx, &adjustingEntries)
	if err != nil {
		logger.Error("Adjusting entries generation failed", "error", err)
		result.Status = domain.ClosingStatusFailed
		result.Error = err.Error()
		return result, nil
	}

	result.Steps["adjusting_entries"] = domain.ClosingStepResult{
		StepName:       "Adjusting Entries",
		Status:         domain.StepStatusCompleted,
		CompletedAt:    workflow.Now(ctx),
		ProcessedCount: len(adjustingEntries.GeneratedEntries),
	}

	// Step 4: Calculate depreciation
	logger.Info("Step 4: Depreciation calculation")
	var depreciationResult domain.DepreciationCalculationResult
	err = workflow.ExecuteActivity(ctx, "CalculateDepreciation", domain.DepreciationCalculationInput{
		ClosingPeriod: input.ClosingPeriod,
		TenantID:      input.TenantID,
		AssetFilter:   input.AssetFilter,
	}).Get(ctx, &depreciationResult)
	if err != nil {
		logger.Error("Depreciation calculation failed", "error", err)
		result.Status = domain.ClosingStatusFailed
		result.Error = err.Error()
		return result, nil
	}

	result.Steps["depreciation"] = domain.ClosingStepResult{
		StepName:       "Depreciation Calculation",
		Status:         domain.StepStatusCompleted,
		CompletedAt:    workflow.Now(ctx),
		ProcessedCount: depreciationResult.ProcessedAssets,
	}

	// Step 5: Generate closing entries
	logger.Info("Step 5: Closing entries")
	var closingEntries domain.ClosingEntriesResult
	err = workflow.ExecuteActivity(ctx, "GenerateClosingEntries", domain.ClosingEntriesInput{
		ClosingPeriod:       input.ClosingPeriod,
		TenantID:            input.TenantID,
		AdjustingEntries:    adjustingEntries.GeneratedEntries,
		DepreciationEntries: depreciationResult.DepreciationEntries,
	}).Get(ctx, &closingEntries)
	if err != nil {
		logger.Error("Closing entries generation failed", "error", err)
		result.Status = domain.ClosingStatusFailed
		result.Error = err.Error()
		return result, nil
	}

	result.Steps["closing_entries"] = domain.ClosingStepResult{
		StepName:       "Closing Entries",
		Status:         domain.StepStatusCompleted,
		CompletedAt:    workflow.Now(ctx),
		ProcessedCount: len(closingEntries.GeneratedEntries),
	}

	// Step 6: Generate financial statements
	logger.Info("Step 6: Financial statements")
	var financialStatements domain.FinancialStatementsResult
	err = workflow.ExecuteActivity(ctx, "GenerateFinancialStatements", domain.FinancialStatementsInput{
		ClosingPeriod:  input.ClosingPeriod,
		TenantID:       input.TenantID,
		StatementTypes: input.StatementTypes,
	}).Get(ctx, &financialStatements)
	if err != nil {
		logger.Error("Financial statements generation failed", "error", err)
		result.Status = domain.ClosingStatusFailed
		result.Error = err.Error()
		return result, nil
	}

	result.Steps["financial_statements"] = domain.ClosingStepResult{
		StepName:       "Financial Statements",
		Status:         domain.StepStatusCompleted,
		CompletedAt:    workflow.Now(ctx),
		ProcessedCount: len(financialStatements.GeneratedStatements),
	}

	// Step 7: Finalize closing
	logger.Info("Step 7: Finalize closing")
	err = workflow.ExecuteActivity(ctx, "FinalizeMonthEndClosing", domain.FinalizeClosingInput{
		ClosingPeriod:       input.ClosingPeriod,
		TenantID:            input.TenantID,
		ClosingEntries:      closingEntries.GeneratedEntries,
		FinancialStatements: financialStatements.GeneratedStatements,
		FinalizedBy:         input.InitiatedBy,
	}).Get(ctx, nil)
	if err != nil {
		logger.Error("Closing finalization failed", "error", err)
		result.Status = domain.ClosingStatusFailed
		result.Error = err.Error()
		return result, nil
	}

	result.Steps["finalization"] = domain.ClosingStepResult{
		StepName:    "Finalization",
		Status:      domain.StepStatusCompleted,
		CompletedAt: workflow.Now(ctx),
	}

	// Step 8: Send notifications
	err = workflow.ExecuteActivity(ctx, "SendClosingNotifications", domain.ClosingNotificationInput{
		ClosingPeriod:       input.ClosingPeriod,
		ClosingStatus:       domain.ClosingStatusCompleted,
		FinancialStatements: financialStatements.GeneratedStatements,
		Recipients:          input.NotificationRecipients,
	}).Get(ctx, nil)
	if err != nil {
		logger.Warn("Failed to send closing notifications", "error", err)
		// Don't fail the workflow for notification errors
	}

	result.Status = domain.ClosingStatusCompleted
	result.FinancialStatements = financialStatements.GeneratedStatements
	result.CompletedTime = workflow.Now(ctx)
	result.Duration = result.CompletedTime.Sub(result.StartTime)

	logger.Info("Month-end closing workflow completed successfully",
		"period", input.ClosingPeriod,
		"steps_completed", len(result.Steps),
		"duration", result.Duration)

	return result, nil
}

// YearEndClosingWorkflow handles year-end closing processes
func (pw *PeriodicWorkflows) YearEndClosingWorkflow(ctx workflow.Context, input domain.YearEndClosingWorkflowInput) (*domain.YearEndClosingWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting year-end closing workflow", "fiscal_year", input.FiscalYear)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Hour * 8, // Year-end can take much longer
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 2,
			InitialInterval: time.Minute * 10,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	result := &domain.YearEndClosingWorkflowResult{
		FiscalYear: input.FiscalYear,
		StartTime:  workflow.Now(ctx),
		Steps:      make(map[string]domain.ClosingStepResult),
	}

	// Step 1: Pre-year-end validation
	logger.Info("Step 1: Pre-year-end validation")
	var preYearEndValidation domain.PreYearEndValidationResult
	err := workflow.ExecuteActivity(ctx, "ValidatePreYearEndConditions", domain.PreYearEndValidationInput{
		FiscalYear:      input.FiscalYear,
		TenantID:        input.TenantID,
		ValidationRules: input.ValidationRules,
	}).Get(ctx, &preYearEndValidation)

	if err != nil || !preYearEndValidation.IsValid {
		logger.Error("Pre-year-end validation failed", "error", err)
		result.Status = domain.ClosingStatusFailed
		result.Error = err.Error()
		result.ValidationErrors = preYearEndValidation.ValidationErrors
		return result, nil
	}

	result.Steps["pre_validation"] = domain.ClosingStepResult{
		StepName:    "Pre-Year-End Validation",
		Status:      domain.StepStatusCompleted,
		CompletedAt: workflow.Now(ctx),
	}

	// Step 2: Ensure all months are closed
	logger.Info("Step 2: Validate monthly closings")
	var monthlyClosingCheck domain.MonthlyClosingCheckResult
	err = workflow.ExecuteActivity(ctx, "ValidateMonthlyClosings", domain.MonthlyClosingCheckInput{
		FiscalYear: input.FiscalYear,
		TenantID:   input.TenantID,
	}).Get(ctx, &monthlyClosingCheck)

	if err != nil || !monthlyClosingCheck.AllMonthsClosed {
		logger.Error("Monthly closing validation failed", "error", err)
		result.Status = domain.ClosingStatusFailed
		result.Error = "Not all months are closed for fiscal year"
		result.UnclosedMonths = monthlyClosingCheck.UnclosedMonths
		return result, nil
	}

	result.Steps["monthly_validation"] = domain.ClosingStepResult{
		StepName:    "Monthly Closing Validation",
		Status:      domain.StepStatusCompleted,
		CompletedAt: workflow.Now(ctx),
	}

	// Step 3: Calculate annual depreciation adjustments
	logger.Info("Step 3: Annual depreciation adjustments")
	var annualDepreciation domain.AnnualDepreciationResult
	err = workflow.ExecuteActivity(ctx, "CalculateAnnualDepreciation", domain.AnnualDepreciationInput{
		FiscalYear:  input.FiscalYear,
		TenantID:    input.TenantID,
		AssetFilter: input.AssetFilter,
	}).Get(ctx, &annualDepreciation)
	if err != nil {
		logger.Error("Annual depreciation calculation failed", "error", err)
		result.Status = domain.ClosingStatusFailed
		result.Error = err.Error()
		return result, nil
	}

	result.Steps["annual_depreciation"] = domain.ClosingStepResult{
		StepName:       "Annual Depreciation",
		Status:         domain.StepStatusCompleted,
		CompletedAt:    workflow.Now(ctx),
		ProcessedCount: annualDepreciation.ProcessedAssets,
	}

	// Step 4: Process year-end accruals
	logger.Info("Step 4: Year-end accruals")
	var accrualResult domain.YearEndAccrualsResult
	err = workflow.ExecuteActivity(ctx, "ProcessYearEndAccruals", domain.YearEndAccrualsInput{
		FiscalYear:   input.FiscalYear,
		TenantID:     input.TenantID,
		AccrualRules: input.AccrualRules,
	}).Get(ctx, &accrualResult)
	if err != nil {
		logger.Error("Year-end accruals processing failed", "error", err)
		result.Status = domain.ClosingStatusFailed
		result.Error = err.Error()
		return result, nil
	}

	result.Steps["accruals"] = domain.ClosingStepResult{
		StepName:       "Year-End Accruals",
		Status:         domain.StepStatusCompleted,
		CompletedAt:    workflow.Now(ctx),
		ProcessedCount: len(accrualResult.ProcessedAccruals),
	}

	// Step 5: Generate year-end closing entries
	logger.Info("Step 5: Year-end closing entries")
	var yearEndClosingEntries domain.YearEndClosingEntriesResult
	err = workflow.ExecuteActivity(ctx, "GenerateYearEndClosingEntries", domain.YearEndClosingEntriesInput{
		FiscalYear:          input.FiscalYear,
		TenantID:            input.TenantID,
		DepreciationEntries: annualDepreciation.DepreciationEntries,
		AccrualEntries:      accrualResult.AccrualEntries,
	}).Get(ctx, &yearEndClosingEntries)
	if err != nil {
		logger.Error("Year-end closing entries generation failed", "error", err)
		result.Status = domain.ClosingStatusFailed
		result.Error = err.Error()
		return result, nil
	}

	result.Steps["closing_entries"] = domain.ClosingStepResult{
		StepName:       "Year-End Closing Entries",
		Status:         domain.StepStatusCompleted,
		CompletedAt:    workflow.Now(ctx),
		ProcessedCount: len(yearEndClosingEntries.GeneratedEntries),
	}

	// Step 6: Generate annual financial statements
	logger.Info("Step 6: Annual financial statements")
	var annualStatements domain.AnnualFinancialStatementsResult
	err = workflow.ExecuteActivity(ctx, "GenerateAnnualFinancialStatements", domain.AnnualFinancialStatementsInput{
		FiscalYear:       input.FiscalYear,
		TenantID:         input.TenantID,
		StatementTypes:   input.StatementTypes,
		IncludePriorYear: true,
	}).Get(ctx, &annualStatements)
	if err != nil {
		logger.Error("Annual financial statements generation failed", "error", err)
		result.Status = domain.ClosingStatusFailed
		result.Error = err.Error()
		return result, nil
	}

	result.Steps["financial_statements"] = domain.ClosingStepResult{
		StepName:       "Annual Financial Statements",
		Status:         domain.StepStatusCompleted,
		CompletedAt:    workflow.Now(ctx),
		ProcessedCount: len(annualStatements.GeneratedStatements),
	}

	// Step 7: Archive fiscal year data
	logger.Info("Step 7: Archive fiscal year data")
	var archiveResult domain.FiscalYearArchiveResult
	err = workflow.ExecuteActivity(ctx, "ArchiveFiscalYearData", domain.FiscalYearArchiveInput{
		FiscalYear:      input.FiscalYear,
		TenantID:        input.TenantID,
		ArchiveSettings: input.ArchiveSettings,
	}).Get(ctx, &archiveResult)
	if err != nil {
		logger.Error("Fiscal year data archival failed", "error", err)
		result.Status = domain.ClosingStatusFailed
		result.Error = err.Error()
		return result, nil
	}

	result.Steps["archival"] = domain.ClosingStepResult{
		StepName:       "Data Archival",
		Status:         domain.StepStatusCompleted,
		CompletedAt:    workflow.Now(ctx),
		ProcessedCount: archiveResult.ArchivedRecords,
	}

	// Step 8: Finalize year-end closing
	logger.Info("Step 8: Finalize year-end closing")
	err = workflow.ExecuteActivity(ctx, "FinalizeYearEndClosing", domain.FinalizeYearEndClosingInput{
		FiscalYear:          input.FiscalYear,
		TenantID:            input.TenantID,
		ClosingEntries:      yearEndClosingEntries.GeneratedEntries,
		FinancialStatements: annualStatements.GeneratedStatements,
		ArchiveReference:    archiveResult.ArchiveReference,
		FinalizedBy:         input.InitiatedBy,
	}).Get(ctx, nil)
	if err != nil {
		logger.Error("Year-end closing finalization failed", "error", err)
		result.Status = domain.ClosingStatusFailed
		result.Error = err.Error()
		return result, nil
	}

	result.Steps["finalization"] = domain.ClosingStepResult{
		StepName:    "Finalization",
		Status:      domain.StepStatusCompleted,
		CompletedAt: workflow.Now(ctx),
	}

	result.Status = domain.ClosingStatusCompleted
	result.FinancialStatements = annualStatements.GeneratedStatements
	result.ArchiveReference = archiveResult.ArchiveReference
	result.CompletedTime = workflow.Now(ctx)
	result.Duration = result.CompletedTime.Sub(result.StartTime)

	logger.Info("Year-end closing workflow completed successfully",
		"fiscal_year", input.FiscalYear,
		"steps_completed", len(result.Steps),
		"archive_ref", result.ArchiveReference,
		"duration", result.Duration)

	return result, nil
}
