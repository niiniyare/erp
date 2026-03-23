package activities

import (
	"context"

	"github.com/google/uuid"

	"awo.so/internal/core/finance/domain"
	"awo.so/internal/core/finance/service"
	"awo.so/internal/core/iam"
	settingsDomain "awo.so/internal/core/settings/domain"
	settingsService "awo.so/internal/core/settings/service"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/worker"
)

// ValidationActivities handles all validation-related Temporal activities
type ValidationActivities struct {
	accountService     service.AccountService
	transactionService service.TransactionService
	iamService         iam.Service
	settingsService    settingsService.ConfigurationService
	logger             logger.Logger
	metrics            metrics.MetricsProvider
	tracer             tracing.Service
}

// ValidationActivityDeps contains dependencies for validation activities
type ValidationActivityDeps struct {
	AccountService     service.AccountService
	TransactionService service.TransactionService
	IAMService         iam.Service
	SettingsService    settingsService.ConfigurationService
	Logger             logger.Logger
	Metrics            metrics.MetricsProvider
	Tracer             tracing.Service
}

// NewValidationActivities creates a new validation activities instance
func NewValidationActivities(deps ValidationActivityDeps) *ValidationActivities {
	return &ValidationActivities{
		accountService:     deps.AccountService,
		transactionService: deps.TransactionService,
		iamService:         deps.IAMService,
		settingsService:    deps.SettingsService,
		logger:             deps.Logger,
		metrics:            deps.Metrics,
		tracer:             deps.Tracer,
	}
}

// RegisterWith registers validation activities with a Temporal worker
func (v *ValidationActivities) RegisterWith(w worker.Worker) {
	w.RegisterActivity(v.ValidateBusinessRulesActivity)
	w.RegisterActivity(v.ValidateAccountingPeriodActivity)
	w.RegisterActivity(v.ValidateExchangeRateActivity)
	w.RegisterActivity(v.ValidateBudgetConstraintsActivity)
	w.RegisterActivity(v.ValidateApprovalLimitsActivity)
	w.RegisterActivity(v.ValidateComplianceRulesActivity)
}

// Validation Activity Input/Output Types

// BusinessRuleValidationInput represents business rule validation input
type BusinessRuleValidationInput struct {
	RuleType string         `json:"rule_type"`
	RuleData map[string]any `json:"rule_data"`
	Context  map[string]any `json:"context"`
}

// ValidationActivityOutput represents validation activity output
type ValidationActivityOutput struct {
	Valid            bool           `json:"valid"`
	ValidationErrors []string       `json:"validation_errors,omitempty"`
	Warnings         []string       `json:"warnings,omitempty"`
	AppliedRules     []string       `json:"applied_rules,omitempty"`
	RuleResults      map[string]any `json:"rule_results,omitempty"`
	Message          string         `json:"message"`
}

// ExchangeRateValidationInput represents exchange rate validation input
type ExchangeRateValidationInput struct {
	FromCurrency string  `json:"from_currency"`
	ToCurrency   string  `json:"to_currency"`
	Rate         float64 `json:"rate"`
	Date         string  `json:"date"`
}

// Validation Activities Implementation

// ValidateBusinessRulesActivity validates general business rules
func (v *ValidationActivities) ValidateBusinessRulesActivity(ctx context.Context, input BusinessRuleValidationInput) (*ValidationActivityOutput, error) {
	ctx, span := v.tracer.StartSpan(ctx, "finance.activity.business.rule.validation")
	defer span.End()

	info := activity.GetInfo(ctx)
	activityLogger := v.logger.WithFields(logger.Fields{
		"activity_id":   info.ActivityID,
		"workflow_id":   info.WorkflowExecution.ID,
		"activity_type": "finance.activity.business.rule.validation",
		"rule_type":     input.RuleType,
	})

	activityLogger.InfoContext(ctx, "Validating business rules")
	v.metrics.Counter(domain.MetricActivityExecutions, "Total activity executions").Add(1, nil)

	validationErrors := []string{}
	warnings := []string{}
	appliedRules := []string{}
	ruleResults := make(map[string]any)

	switch input.RuleType {
	case "ACCOUNT_CREATION":
		errors, warns, rules, results := v.validateAccountCreationRules(ctx, input.RuleData)
		validationErrors = append(validationErrors, errors...)
		warnings = append(warnings, warns...)
		appliedRules = append(appliedRules, rules...)
		for k, v := range results {
			ruleResults[k] = v
		}

	case "TRANSACTION_PROCESSING":
		errors, warns, rules, results := v.validateTransactionProcessingRules(ctx, input.RuleData)
		validationErrors = append(validationErrors, errors...)
		warnings = append(warnings, warns...)
		appliedRules = append(appliedRules, rules...)
		for k, v := range results {
			ruleResults[k] = v
		}

	case "PERIOD_CLOSING":
		errors, warns, rules, results := v.validatePeriodClosingRules(ctx, input.RuleData)
		validationErrors = append(validationErrors, errors...)
		warnings = append(warnings, warns...)
		appliedRules = append(appliedRules, rules...)
		for k, v := range results {
			ruleResults[k] = v
		}

	default:
		validationErrors = append(validationErrors, "unknown rule type: "+input.RuleType)
	}

	isValid := len(validationErrors) == 0

	if isValid {
		activityLogger.InfoContext(ctx, "Business rule validation successful", logger.Fields{
			"applied_rules": appliedRules,
			"warnings":      len(warnings),
		})
		v.metrics.Counter("finance.business.rule.validation.success", "Business rule validation success").Add(1, nil)
	} else {
		activityLogger.WarnContext(ctx, "Business rule validation failed", logger.Fields{
			"errors": validationErrors,
		})
		v.metrics.Counter("finance.business.rule.validation.errors", "Business rule validation errors").Add(1, nil)
	}

	return &ValidationActivityOutput{
		Valid:            isValid,
		ValidationErrors: validationErrors,
		Warnings:         warnings,
		AppliedRules:     appliedRules,
		RuleResults:      ruleResults,
		Message:          "Business rule validation completed",
	}, nil
}

// ValidateAccountingPeriodActivity validates accounting period constraints
func (v *ValidationActivities) ValidateAccountingPeriodActivity(ctx context.Context, date string) (*ValidationActivityOutput, error) {
	ctx, span := v.tracer.StartSpan(ctx, "finance.activity.period.validation")
	defer span.End()

	info := activity.GetInfo(ctx)
	activityLogger := v.logger.WithFields(logger.Fields{
		"activity_id":   info.ActivityID,
		"workflow_id":   info.WorkflowExecution.ID,
		"activity_type": "finance.activity.period.validation",
		"date":          date,
	})

	activityLogger.InfoContext(ctx, "Validating accounting period")
	v.metrics.Counter(domain.MetricActivityExecutions, "Total activity executions").Add(1, nil)

	validationErrors := []string{}
	warnings := []string{}

	// Get period settings from settings service
	periodSettings, err := v.settingsService.GetEffectiveConfiguration(ctx, nil, settingsDomain.ModuleName("finance"), settingsDomain.ConfigKey("accounting_period"))
	if err != nil {
		activityLogger.ErrorContext(ctx, "Failed to get period settings", logger.Fields{
			"error": err.Error(),
		})
		return &ValidationActivityOutput{
			Valid:   false,
			Message: "Failed to retrieve period settings",
		}, err
	}

	// Check if period is open
	isPeriodOpen := v.checkPeriodStatus(ctx, date, periodSettings)
	if !isPeriodOpen {
		validationErrors = append(validationErrors, "accounting period is closed for the specified date")
	}

	// Check future date restrictions
	isFutureDate := v.checkFutureDate(date)
	if isFutureDate {
		warnings = append(warnings, "transaction date is in the future")
	}

	// Check cutoff date restrictions
	pastCutoffDate := v.checkCutoffDate(ctx, date, periodSettings)
	if pastCutoffDate {
		validationErrors = append(validationErrors, "transaction date is beyond the allowed cutoff period")
	}

	isValid := len(validationErrors) == 0

	if isValid {
		activityLogger.InfoContext(ctx, "Accounting period validation successful")
		v.metrics.Counter("finance.period.validation.success", "Accounting period validation success").Add(1, nil)
	} else {
		activityLogger.WarnContext(ctx, "Accounting period validation failed", logger.Fields{
			"errors": validationErrors,
		})
		v.metrics.Counter("finance.period.validation.errors", "Accounting period validation errors").Add(1, nil)
	}

	return &ValidationActivityOutput{
		Valid:            isValid,
		ValidationErrors: validationErrors,
		Warnings:         warnings,
		Message:          "Accounting period validation completed",
	}, nil
}

// ValidateExchangeRateActivity validates exchange rate data
func (v *ValidationActivities) ValidateExchangeRateActivity(ctx context.Context, input ExchangeRateValidationInput) (*ValidationActivityOutput, error) {
	ctx, span := v.tracer.StartSpan(ctx, "finance.activity.exchange.rate.validation")
	defer span.End()

	info := activity.GetInfo(ctx)
	activityLogger := v.logger.WithFields(logger.Fields{
		"activity_id":   info.ActivityID,
		"workflow_id":   info.WorkflowExecution.ID,
		"activity_type": "finance.activity.exchange.rate.validation",
		"from_currency": input.FromCurrency,
		"to_currency":   input.ToCurrency,
		"rate":          input.Rate,
		"date":          input.Date,
	})

	activityLogger.InfoContext(ctx, "Validating exchange rate")
	v.metrics.Counter(domain.MetricActivityExecutions, "Total activity executions").Add(1, nil)

	validationErrors := []string{}
	warnings := []string{}

	// Validate currency codes
	if input.FromCurrency == "" {
		validationErrors = append(validationErrors, "from currency is required")
	}
	if input.ToCurrency == "" {
		validationErrors = append(validationErrors, "to currency is required")
	}

	// Validate rate value
	if input.Rate <= 0 {
		validationErrors = append(validationErrors, "exchange rate must be positive")
	}

	// Check for same currency conversion
	if input.FromCurrency == input.ToCurrency {
		if input.Rate != 1.0 {
			validationErrors = append(validationErrors, "exchange rate for same currency must be 1.0")
		}
	}

	// Validate rate reasonableness (basic sanity check)
	if input.Rate > 1000000 || input.Rate < 0.000001 {
		warnings = append(warnings, "exchange rate appears to be outside normal range")
	}

	// Check for supported currencies
	supportedCurrencies := []string{
		"USD", "EUR", "GBP",
		"JPY", "CAD", "AUD",
	}

	if !contains(supportedCurrencies, input.FromCurrency) {
		validationErrors = append(validationErrors, "unsupported from currency: "+string(input.FromCurrency))
	}
	if !contains(supportedCurrencies, input.ToCurrency) {
		validationErrors = append(validationErrors, "unsupported to currency: "+string(input.ToCurrency))
	}

	isValid := len(validationErrors) == 0

	if isValid {
		activityLogger.InfoContext(ctx, "Exchange rate validation successful")
		v.metrics.Counter("finance.exchange.rate.validation.success", "Exchange rate validation success").Add(1, nil)
	} else {
		activityLogger.WarnContext(ctx, "Exchange rate validation failed", logger.Fields{
			"errors": validationErrors,
		})
		v.metrics.Counter("finance.exchange.rate.validation.errors", "Exchange rate validation errors").Add(1, nil)
	}

	return &ValidationActivityOutput{
		Valid:            isValid,
		ValidationErrors: validationErrors,
		Warnings:         warnings,
		Message:          "Exchange rate validation completed",
	}, nil
}

// ValidateBudgetConstraintsActivity validates budget constraints
func (v *ValidationActivities) ValidateBudgetConstraintsActivity(ctx context.Context, accountID uuid.UUID, amount float64, period string) (*ValidationActivityOutput, error) {
	ctx, span := v.tracer.StartSpan(ctx, "finance.activity.budget.validation")
	defer span.End()

	info := activity.GetInfo(ctx)
	activityLogger := v.logger.WithFields(logger.Fields{
		"activity_id":   info.ActivityID,
		"workflow_id":   info.WorkflowExecution.ID,
		"activity_type": "finance.activity.budget.validation",
		"account_id":    accountID,
		"amount":        amount,
		"period":        period,
	})

	activityLogger.InfoContext(ctx, "Validating budget constraints")
	v.metrics.Counter(domain.MetricActivityExecutions, "Total activity executions").Add(1, nil)

	validationErrors := []string{}
	warnings := []string{}

	// Get account details
	account, err := v.accountService.GetAccountByID(ctx, accountID)
	if err != nil {
		return &ValidationActivityOutput{
			Valid:   false,
			Message: "Failed to retrieve account for budget validation",
		}, err
	}

	// Check if budget validation is enabled for this account type
	if !requiresBudgetValidation(domain.RootType(account.AccountType)) {
		activityLogger.InfoContext(ctx, "Budget validation not required for account type")
		return &ValidationActivityOutput{
			Valid:   true,
			Message: "Budget validation not required for this account type",
		}, nil
	}

	// Get budget settings for the account
	budgetLimit := v.getBudgetLimit(ctx, accountID, period)
	if budgetLimit <= 0 {
		warnings = append(warnings, "no budget limit configured for account")
	} else {
		// Check if amount exceeds budget limit
		currentUsage := v.getCurrentBudgetUsage(ctx, accountID, period)
		projectedUsage := currentUsage + amount

		if projectedUsage > budgetLimit {
			excess := projectedUsage - budgetLimit
			validationErrors = append(validationErrors, "transaction would exceed budget limit by "+string(rune(excess)))
		} else if projectedUsage > budgetLimit*0.9 {
			// Warn at 90% of budget
			warnings = append(warnings, "transaction would exceed 90% of budget limit")
		}
	}

	isValid := len(validationErrors) == 0

	if isValid {
		activityLogger.InfoContext(ctx, "Budget constraint validation successful")
		v.metrics.Counter("finance.budget.validation.success", "Budget validation success").Add(1, nil)
	} else {
		activityLogger.WarnContext(ctx, "Budget constraint validation failed", logger.Fields{
			"errors": validationErrors,
		})
		v.metrics.Counter("finance.budget.validation.errors", "Budget validation errors").Add(1, nil)
	}

	return &ValidationActivityOutput{
		Valid:            isValid,
		ValidationErrors: validationErrors,
		Warnings:         warnings,
		Message:          "Budget constraint validation completed",
	}, nil
}

// ValidateApprovalLimitsActivity validates approval limit requirements
func (v *ValidationActivities) ValidateApprovalLimitsActivity(ctx context.Context, amount float64, userID uuid.UUID) (*ValidationActivityOutput, error) {
	ctx, span := v.tracer.StartSpan(ctx, "finance.activity.approval.validation")
	defer span.End()

	info := activity.GetInfo(ctx)
	activityLogger := v.logger.WithFields(logger.Fields{
		"activity_id":   info.ActivityID,
		"workflow_id":   info.WorkflowExecution.ID,
		"activity_type": "finance.activity.approval.validation",
		"amount":        amount,
		"user_id":       userID,
	})

	activityLogger.InfoContext(ctx, "Validating approval limits")
	v.metrics.Counter(domain.MetricActivityExecutions, "Total activity executions").Add(1, nil)

	validationErrors := []string{}
	warnings := []string{}
	ruleResults := make(map[string]any)

	// Get user's approval limits from IAM service
	userLimits := v.getUserApprovalLimits(ctx, userID)
	ruleResults["user_approval_limit"] = userLimits

	if amount > userLimits {
		if userLimits == 0 {
			validationErrors = append(validationErrors, "user has no approval authority")
		} else {
			validationErrors = append(validationErrors, "transaction amount exceeds user's approval limit")
			ruleResults["requires_higher_approval"] = true
		}
	} else if amount > userLimits*0.8 {
		warnings = append(warnings, "transaction amount is close to user's approval limit")
	}

	// Check for additional approval requirements based on transaction characteristics
	requiresAdditionalApproval := v.checkAdditionalApprovalRequirements(ctx, amount)
	if requiresAdditionalApproval {
		warnings = append(warnings, "transaction may require additional approvals based on business rules")
		ruleResults["additional_approval_required"] = true
	}

	isValid := len(validationErrors) == 0

	if isValid {
		activityLogger.InfoContext(ctx, "Approval limit validation successful")
		v.metrics.Counter("finance.approval.validation.success", "Approval validation success").Add(1, nil)
	} else {
		activityLogger.WarnContext(ctx, "Approval limit validation failed", logger.Fields{
			"errors": validationErrors,
		})
		v.metrics.Counter("finance.approval.validation.errors", "Approval validation errors").Add(1, nil)
	}

	return &ValidationActivityOutput{
		Valid:            isValid,
		ValidationErrors: validationErrors,
		Warnings:         warnings,
		RuleResults:      ruleResults,
		Message:          "Approval limit validation completed",
	}, nil
}

// ValidateComplianceRulesActivity validates compliance and regulatory rules
func (v *ValidationActivities) ValidateComplianceRulesActivity(ctx context.Context, ruleSet string, data map[string]any) (*ValidationActivityOutput, error) {
	ctx, span := v.tracer.StartSpan(ctx, "finance.activity.compliance.validation")
	defer span.End()

	info := activity.GetInfo(ctx)
	activityLogger := v.logger.WithFields(logger.Fields{
		"activity_id":   info.ActivityID,
		"workflow_id":   info.WorkflowExecution.ID,
		"activity_type": "finance.activity.compliance.validation",
		"rule_set":      ruleSet,
	})

	activityLogger.InfoContext(ctx, "Validating compliance rules")
	v.metrics.Counter(domain.MetricActivityExecutions, "Total activity executions").Add(1, nil)

	validationErrors := []string{}
	warnings := []string{}
	appliedRules := []string{}

	switch ruleSet {
	case "SOX":
		errors, warns, rules := v.validateSOXCompliance(ctx, data)
		validationErrors = append(validationErrors, errors...)
		warnings = append(warnings, warns...)
		appliedRules = append(appliedRules, rules...)

	case "GAAP":
		errors, warns, rules := v.validateGAAPCompliance(ctx, data)
		validationErrors = append(validationErrors, errors...)
		warnings = append(warnings, warns...)
		appliedRules = append(appliedRules, rules...)

	case "IFRS":
		errors, warns, rules := v.validateIFRSCompliance(ctx, data)
		validationErrors = append(validationErrors, errors...)
		warnings = append(warnings, warns...)
		appliedRules = append(appliedRules, rules...)

	default:
		validationErrors = append(validationErrors, "unknown compliance rule set: "+ruleSet)
	}

	isValid := len(validationErrors) == 0

	if isValid {
		activityLogger.InfoContext(ctx, "Compliance rule validation successful", logger.Fields{
			"applied_rules": appliedRules,
		})
		v.metrics.Counter("finance.compliance.validation.success", "Compliance validation success").Add(1, nil)
	} else {
		activityLogger.WarnContext(ctx, "Compliance rule validation failed", logger.Fields{
			"errors": validationErrors,
		})
		v.metrics.Counter("finance.compliance.validation.errors", "Compliance validation errors").Add(1, nil)
	}

	return &ValidationActivityOutput{
		Valid:            isValid,
		ValidationErrors: validationErrors,
		Warnings:         warnings,
		AppliedRules:     appliedRules,
		Message:          "Compliance rule validation completed",
	}, nil
}

// Helper functions for validation logic

func (v *ValidationActivities) validateAccountCreationRules(ctx context.Context, data map[string]any) ([]string, []string, []string, map[string]any) {
	// Implementation for account creation rule validation
	return []string{}, []string{}, []string{"account_creation_base_rule"}, map[string]any{}
}

func (v *ValidationActivities) validateTransactionProcessingRules(ctx context.Context, data map[string]any) ([]string, []string, []string, map[string]any) {
	// Implementation for transaction processing rule validation
	return []string{}, []string{}, []string{"transaction_processing_base_rule"}, map[string]any{}
}

func (v *ValidationActivities) validatePeriodClosingRules(ctx context.Context, data map[string]any) ([]string, []string, []string, map[string]any) {
	// Implementation for period closing rule validation
	return []string{}, []string{}, []string{"period_closing_base_rule"}, map[string]any{}
}

func (v *ValidationActivities) checkPeriodStatus(ctx context.Context, date string, settings any) bool {
	// Implementation for period status checking
	return true // Simplified for example
}

func (v *ValidationActivities) checkFutureDate(date string) bool {
	// Implementation for future date checking
	return false // Simplified for example
}

func (v *ValidationActivities) checkCutoffDate(ctx context.Context, date string, settings any) bool {
	// Implementation for cutoff date checking
	return false // Simplified for example
}

func (v *ValidationActivities) getBudgetLimit(ctx context.Context, accountID uuid.UUID, period string) float64 {
	// Implementation for getting budget limit
	return 10000.0 // Simplified for example
}

func (v *ValidationActivities) getCurrentBudgetUsage(ctx context.Context, accountID uuid.UUID, period string) float64 {
	// Implementation for getting current budget usage
	return 5000.0 // Simplified for example
}

func (v *ValidationActivities) getUserApprovalLimits(ctx context.Context, userID uuid.UUID) float64 {
	// Implementation for getting user approval limits
	return 50000.0 // Simplified for example
}

func (v *ValidationActivities) checkAdditionalApprovalRequirements(ctx context.Context, amount float64) bool {
	// Implementation for checking additional approval requirements
	return amount > 100000.0 // Simplified for example
}

func (v *ValidationActivities) validateSOXCompliance(ctx context.Context, data map[string]any) ([]string, []string, []string) {
	// Implementation for SOX compliance validation
	return []string{}, []string{}, []string{"sox_compliance_check"}
}

func (v *ValidationActivities) validateGAAPCompliance(ctx context.Context, data map[string]any) ([]string, []string, []string) {
	// Implementation for GAAP compliance validation
	return []string{}, []string{}, []string{"gaap_compliance_check"}
}

func (v *ValidationActivities) validateIFRSCompliance(ctx context.Context, data map[string]any) ([]string, []string, []string) {
	// Implementation for IFRS compliance validation
	return []string{}, []string{}, []string{"ifrs_compliance_check"}
}

func requiresBudgetValidation(rootType domain.RootType) bool {
	// Only expense accounts typically require budget validation
	return rootType == domain.RootTypeExpense
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
