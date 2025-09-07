package activities

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/niiniyare/erp/internal/core/audit"
	"github.com/niiniyare/erp/internal/core/featureflag"
	"github.com/niiniyare/erp/internal/core/finance/domain"
	"github.com/niiniyare/erp/internal/core/finance/service"
	"github.com/niiniyare/erp/internal/core/iam"
	"github.com/niiniyare/erp/internal/core/iam/authz"
	"github.com/niiniyare/erp/internal/core/iam/model"
	settingsService "github.com/niiniyare/erp/internal/core/settings/service"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/worker"
)

// TransactionActivities handles all transaction-related Temporal activities
type TransactionActivities struct {
	transactionService      service.TransactionService
	transactionEntryService service.TransactionEntryService
	accountService          service.AccountService
	iamService              iam.Service
	auditService            audit.Service
	featureFlagService      featureflag.Service
	settingsService         settingsService.ConfigurationService
	cacheService            cache.Service
	logger                  logger.Logger
	metrics                 metrics.MetricsProvider
	tracer                  tracing.TracingService
}

// TransactionActivityDeps contains dependencies for transaction activities
type TransactionActivityDeps struct {
	TransactionService      service.TransactionService
	TransactionEntryService service.TransactionEntryService
	AccountService          service.AccountService
	IAMService              iam.Service
	AuditService            audit.Service
	FeatureFlagService      featureflag.Service
	SettingsService         settingsService.ConfigurationService
	CacheService            cache.Service
	Logger                  logger.Logger
	Metrics                 metrics.MetricsProvider
	Tracer                  tracing.TracingService
}

// NewTransactionActivities creates a new transaction activities instance
func NewTransactionActivities(deps TransactionActivityDeps) *TransactionActivities {
	return &TransactionActivities{
		transactionService:      deps.TransactionService,
		transactionEntryService: deps.TransactionEntryService,
		accountService:          deps.AccountService,
		iamService:              deps.IAMService,
		auditService:            deps.AuditService,
		featureFlagService:      deps.FeatureFlagService,
		settingsService:         deps.SettingsService,
		cacheService:            deps.CacheService,
		logger:                  deps.Logger,
		metrics:                 deps.Metrics,
		tracer:                  deps.Tracer,
	}
}

// RegisterWith registers transaction activities with a Temporal worker
func (t *TransactionActivities) RegisterWith(w worker.Worker) {
	w.RegisterActivity(t.ValidateTransactionActivity)
	w.RegisterActivity(t.CreateTransactionActivity)
	w.RegisterActivity(t.ProcessTransactionEntriesActivity)
	w.RegisterActivity(t.PostTransactionActivity)
	w.RegisterActivity(t.ReverseTransactionActivity)
	w.RegisterActivity(t.ValidateDoubleEntryActivity)
	w.RegisterActivity(t.CheckTransactionPermissionsActivity)
	w.RegisterActivity(t.UpdateTransactionStatusActivity)
	w.RegisterActivity(t.CacheTransactionActivity)
	w.RegisterActivity(t.ProcessBulkTransactionsActivity)
}

// Transaction Activity Input/Output Types

// CreateTransactionActivityInput represents the input for creating a transaction
type CreateTransactionActivityInput struct {
	TransactionNumber string                   `json:"transaction_number"`
	TransactionDate   time.Time                `json:"transaction_date"`
	Description       string                   `json:"description"`
	Reference         string                   `json:"reference"`
	TransactionType   domain.TransactionType   `json:"transaction_type"`
	Status            domain.TransactionStatus `json:"status"`
	Currency          *string                  `json:"currency"`
	ExchangeRate      decimal.Decimal          `json:"exchange_rate"`
	Entries           []TransactionEntryInput  `json:"entries"`
	Metadata          map[string]interface{}   `json:"metadata"`
}

// TransactionEntryInput represents a transaction entry
type TransactionEntryInput struct {
	AccountID   uuid.UUID              `json:"account_id"`
	EntryType   domain.TransactionType `json:"entry_type"`
	Amount      decimal.Decimal        `json:"amount"`
	Currency    *string                `json:"currency"`
	Description string                 `json:"description"`
	Reference   string                 `json:"reference"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// TransactionActivityOutput represents the output of transaction operations
type TransactionActivityOutput struct {
	Transaction      *domain.Transaction        `json:"transaction,omitempty"`
	Entries          []*domain.TransactionEntry `json:"entries,omitempty"`
	Success          bool                       `json:"success"`
	Message          string                     `json:"message"`
	ErrorCode        string                     `json:"error_code,omitempty"`
	ValidationErrors []string                   `json:"validation_errors,omitempty"`
}

// BulkTransactionActivityInput represents input for bulk transaction processing
type BulkTransactionActivityInput struct {
	Transactions    []CreateTransactionActivityInput `json:"transactions"`
	BatchSize       int                              `json:"batch_size"`
	ConcurrentLimit int                              `json:"concurrent_limit"`
}

// Transaction Activities Implementation

// ValidateTransactionActivity validates a transaction before creation
func (t *TransactionActivities) ValidateTransactionActivity(ctx context.Context, input CreateTransactionActivityInput) (*TransactionActivityOutput, error) {
	ctx, span := t.tracer.StartSpan(ctx, domain.ActivityTypeTransactionValidation)
	defer span.End()

	info := activity.GetInfo(ctx)
	activityLogger := t.logger.WithFields(logger.Fields{
		"activity_id":        info.ActivityID,
		"workflow_id":        info.WorkflowExecution.ID,
		"activity_type":      domain.ActivityTypeTransactionValidation,
		"transaction_number": input.TransactionNumber,
	})

	activityLogger.InfoContext(ctx, "Starting transaction validation")
	t.metrics.Counter(domain.MetricActivityExecutions, "Total activity executions").Add(1, nil)

	validationErrors := []string{}

	// Validate basic transaction data
	if input.TransactionNumber == "" {
		validationErrors = append(validationErrors, "transaction number is required")
	}

	if input.Description == "" {
		validationErrors = append(validationErrors, "transaction description is required")
	}

	if len(input.Entries) == 0 {
		validationErrors = append(validationErrors, "transaction must have at least one entry")
	}

	// Validate double-entry constraint
	if len(input.Entries) > 0 {
		totalDebits := decimal.Zero
		totalCredits := decimal.Zero

		for _, entry := range input.Entries {
			switch entry.EntryType {
			case domain.TransactionType(domain.NormalBalanceDebit):
				totalDebits = totalDebits.Add(entry.Amount)
			case domain.TransactionType(domain.NormalBalanceCredit):
				totalCredits = totalCredits.Add(entry.Amount)
			}
		}

		if !totalDebits.Equal(totalCredits) {
			validationErrors = append(validationErrors, "transaction entries must balance (debits must equal credits)")
		}
	}

	// Validate all accounts exist and are active
	for i, entry := range input.Entries {
		account, err := t.accountService.GetAccountByID(ctx, entry.AccountID)
		if err != nil || account == nil {
			validationErrors = append(validationErrors, "account not found for entry "+string(rune(i)))
		} else if !account.IsActive {
			validationErrors = append(validationErrors, "account is inactive for entry "+string(rune(i)))
		}
	}

	// Check for duplicate transaction number
	if input.TransactionNumber != "" {
		existing, _ := t.transactionService.GetTransactionByNumber(ctx, input.TransactionNumber)
		if existing != nil {
			validationErrors = append(validationErrors, "transaction number already exists")
		}
	}

	// Check feature flags for advanced validation
	advancedValidationEnabled, _ := t.featureFlagService.IsEnabled(ctx, domain.FeatureFlagApprovalWorkflow, nil)
	if advancedValidationEnabled {
		activityLogger.InfoContext(ctx, "Performing advanced transaction validation")
		// Additional advanced validation logic
	}

	if len(validationErrors) > 0 {
		activityLogger.WarnContext(ctx, "Transaction validation failed", logger.Fields{
			"errors": validationErrors,
		})
		t.metrics.Counter(domain.MetricValidationErrors, "Validation errors").Add(1, nil)
		return &TransactionActivityOutput{
			Success:          false,
			Message:          "Transaction validation failed",
			ErrorCode:        domain.ErrCodeValidationFailed,
			ValidationErrors: validationErrors,
		}, nil
	}

	activityLogger.InfoContext(ctx, "Transaction validation successful")
	t.metrics.Counter(domain.MetricValidationSuccess, "Validation success").Add(1, nil)

	return &TransactionActivityOutput{
		Success: true,
		Message: "Transaction validation successful",
	}, nil
}

// CreateTransactionActivity creates a new transaction
func (t *TransactionActivities) CreateTransactionActivity(ctx context.Context, input CreateTransactionActivityInput) (*TransactionActivityOutput, error) {
	ctx, span := t.tracer.StartSpan(ctx, domain.ActivityTypeTransactionCreation)
	defer span.End()

	info := activity.GetInfo(ctx)
	activityLogger := t.logger.WithFields(logger.Fields{
		"activity_id":        info.ActivityID,
		"workflow_id":        info.WorkflowExecution.ID,
		"activity_type":      domain.ActivityTypeTransactionCreation,
		"transaction_number": input.TransactionNumber,
	})

	activityLogger.InfoContext(ctx, "Creating new transaction")
	t.metrics.Counter(domain.MetricActivityExecutions, "Total activity executions").Add(1, nil)

	// Convert input to service request
	req := domain.CreateTransactionRequest{
		TransactionNumber:     input.TransactionNumber,
		TransactionDate:       input.TransactionDate,
		Description:           input.Description,
		ReferenceNumber:       &input.Reference,
		TransactionType:       input.TransactionType,
		TransactionStatus:     input.Status,
		CurrencyCode:          *input.Currency,
		ExchangeRate:          input.ExchangeRate,
		TransactionAttributes: input.Metadata,
	}

	// Convert entries
	for _, entryInput := range input.Entries {
		var debitAmount, creditAmount decimal.Decimal
		
		// Convert entry type to debit/credit amounts
		// Note: This assumes EntryType indicates whether it's a debit or credit
		// You may need to adjust this logic based on actual business requirements
		if string(entryInput.EntryType) == string(domain.NormalBalanceDebit) {
			debitAmount = entryInput.Amount
			creditAmount = decimal.Zero
		} else if string(entryInput.EntryType) == string(domain.NormalBalanceCredit) {
			debitAmount = decimal.Zero
			creditAmount = entryInput.Amount
		} else {
			// Default to debit if type is unclear
			debitAmount = entryInput.Amount
			creditAmount = decimal.Zero
		}
		
		req.Entries = append(req.Entries, domain.CreateEntryRequest{
			AccountID:        entryInput.AccountID,
			DebitAmount:      debitAmount,
			CreditAmount:     creditAmount,
			Description:      entryInput.Description,
			Reference:        &entryInput.Reference,
			OriginalCurrency: entryInput.Currency,
			OriginalAmount:   entryInput.Amount,
			ExchangeRate:     decimal.NewFromInt(1), // Default exchange rate
		})
	}

	// Call transaction service
	transaction, err := t.transactionService.CreateTransaction(ctx, req)
	if err != nil {
		activityLogger.ErrorContext(ctx, "Failed to create transaction", logger.Fields{
			"error": err.Error(),
		})
		t.metrics.Counter(domain.MetricActivityFailures, "Transaction errors").Add(1, nil)
		return &TransactionActivityOutput{
			Success:   false,
			Message:   "Transaction creation failed",
			ErrorCode: domain.ErrCodeInternalError,
		}, err
	}

	activityLogger.InfoContext(ctx, "Transaction created successfully", logger.Fields{
		"transaction_id": transaction.ID,
	})
	t.metrics.Counter(domain.MetricTransactionsProcessed, "Transactions processed").Add(1, nil)

	return &TransactionActivityOutput{
		Transaction: transaction,
		Success:     true,
		Message:     "Transaction created successfully",
	}, nil
}

// ProcessTransactionEntriesActivity processes transaction entries
func (t *TransactionActivities) ProcessTransactionEntriesActivity(ctx context.Context, transactionID uuid.UUID, entries []TransactionEntryInput) (*TransactionActivityOutput, error) {
	ctx, span := t.tracer.StartSpan(ctx, domain.ActivityTypeEntryGeneration)
	defer span.End()

	info := activity.GetInfo(ctx)
	activityLogger := t.logger.WithFields(logger.Fields{
		"activity_id":    info.ActivityID,
		"workflow_id":    info.WorkflowExecution.ID,
		"activity_type":  domain.ActivityTypeEntryGeneration,
		"transaction_id": transactionID,
		"entries_count":  len(entries),
	})

	activityLogger.InfoContext(ctx, "Processing transaction entries")
	t.metrics.Counter(domain.MetricActivityExecutions, "Total activity executions").Add(1, nil)

	processedEntries := []*domain.TransactionEntry{}

	for i, entryInput := range entries {
		var debitAmount, creditAmount decimal.Decimal
		
		// Convert entry type to debit/credit amounts
		if string(entryInput.EntryType) == string(domain.NormalBalanceDebit) {
			debitAmount = entryInput.Amount
			creditAmount = decimal.Zero
		} else if string(entryInput.EntryType) == string(domain.NormalBalanceCredit) {
			debitAmount = decimal.Zero
			creditAmount = entryInput.Amount
		} else {
			// Default to debit if type is unclear
			debitAmount = entryInput.Amount
			creditAmount = decimal.Zero
		}
		
		entryReq := &domain.TransactionEntry{
			ID:               uuid.New(),
			AccountID:        entryInput.AccountID,
			DebitAmount:      debitAmount,
			CreditAmount:     creditAmount,
			Description:      entryInput.Description,
			Reference:        &entryInput.Reference,
			OriginalCurrency: entryInput.Currency,
			OriginalAmount:   entryInput.Amount,
			ExchangeRate:     decimal.NewFromInt(1), // Default exchange rate
		}

		err := t.transactionEntryService.CreateEntry(ctx, entryReq)
		if err != nil {
			activityLogger.ErrorContext(ctx, "Failed to create transaction entry", logger.Fields{
				"error":     err.Error(),
				"entry_idx": i,
			})
			return &TransactionActivityOutput{
				Success:   false,
				Message:   "Failed to process transaction entry",
				ErrorCode: "ENTRY_PROCESSING_FAILED",
			}, err
		}

		processedEntries = append(processedEntries, entryReq)
	}

	activityLogger.InfoContext(ctx, "Transaction entries processed successfully")
	t.metrics.Counter(domain.MetricActivityExecutions, "Entries processed").Add(float64(len(processedEntries)), nil)

	return &TransactionActivityOutput{
		Entries: processedEntries,
		Success: true,
		Message: "Transaction entries processed successfully",
	}, nil
}

// PostTransactionActivity posts a transaction to the general ledger
func (t *TransactionActivities) PostTransactionActivity(ctx context.Context, transactionID uuid.UUID) (*TransactionActivityOutput, error) {
	ctx, span := t.tracer.StartSpan(ctx, domain.ActivityTypeTransactionPosting)
	defer span.End()

	info := activity.GetInfo(ctx)
	activityLogger := t.logger.WithFields(logger.Fields{
		"activity_id":    info.ActivityID,
		"workflow_id":    info.WorkflowExecution.ID,
		"activity_type":  domain.ActivityTypeTransactionPosting,
		"transaction_id": transactionID,
	})

	activityLogger.InfoContext(ctx, "Posting transaction to general ledger")
	t.metrics.Counter(domain.MetricActivityExecutions, "Total activity executions").Add(1, nil)

	now := time.Now()
	_, err := t.transactionService.PostTransaction(ctx, transactionID, &now)
	if err != nil {
		activityLogger.ErrorContext(ctx, "Failed to post transaction", logger.Fields{
			"error": err.Error(),
		})
		t.metrics.Counter(domain.MetricActivityFailures, "Transaction posting errors").Add(1, nil)
		return &TransactionActivityOutput{
			Success:   false,
			Message:   "Transaction posting failed",
			ErrorCode: "TRANSACTION_POSTING_FAILED",
		}, err
	}

	activityLogger.InfoContext(ctx, "Transaction posted successfully")
	t.metrics.Counter(domain.MetricTransactionsPosted, "Transactions posted").Add(1, nil)

	return &TransactionActivityOutput{
		Success: true,
		Message: "Transaction posted successfully",
	}, nil
}

// ReverseTransactionActivity creates a reversal transaction
func (t *TransactionActivities) ReverseTransactionActivity(ctx context.Context, originalTransactionID uuid.UUID, reason string) (*TransactionActivityOutput, error) {
	ctx, span := t.tracer.StartSpan(ctx, domain.ActivityTypeTransactionReversal)
	defer span.End()

	info := activity.GetInfo(ctx)
	activityLogger := t.logger.WithFields(logger.Fields{
		"activity_id":             info.ActivityID,
		"workflow_id":             info.WorkflowExecution.ID,
		"activity_type":           domain.ActivityTypeTransactionReversal,
		"original_transaction_id": originalTransactionID,
		"reversal_reason":         reason,
	})

	activityLogger.InfoContext(ctx, "Creating transaction reversal")
	t.metrics.Counter(domain.MetricActivityExecutions, "Total activity executions").Add(1, nil)

	reversalTransaction, err := t.transactionService.ReverseTransaction(ctx, originalTransactionID, reason)
	if err != nil {
		activityLogger.ErrorContext(ctx, "Failed to create transaction reversal", logger.Fields{
			"error": err.Error(),
		})
		t.metrics.Counter(domain.MetricActivityFailures, "Transaction reversal errors").Add(1, nil)
		return &TransactionActivityOutput{
			Success:   false,
			Message:   "Transaction reversal failed",
			ErrorCode: "TRANSACTION_REVERSAL_FAILED",
		}, err
	}

	activityLogger.InfoContext(ctx, "Transaction reversal created successfully", logger.Fields{
		"reversal_transaction_id": reversalTransaction.ID,
	})
	t.metrics.Counter(domain.MetricTransactionsReversed, "Transactions reversed").Add(1, nil)

	return &TransactionActivityOutput{
		Transaction: reversalTransaction,
		Success:     true,
		Message:     "Transaction reversal created successfully",
	}, nil
}

// ValidateDoubleEntryActivity validates double-entry bookkeeping rules
func (t *TransactionActivities) ValidateDoubleEntryActivity(ctx context.Context, entries []TransactionEntryInput) (*TransactionActivityOutput, error) {
	ctx, span := t.tracer.StartSpan(ctx, "finance.activity.double.entry.validation")
	defer span.End()

	info := activity.GetInfo(ctx)
	activityLogger := t.logger.WithFields(logger.Fields{
		"activity_id":   info.ActivityID,
		"workflow_id":   info.WorkflowExecution.ID,
		"activity_type": "finance.activity.double.entry.validation",
		"entries_count": len(entries),
	})

	activityLogger.InfoContext(ctx, "Validating double-entry bookkeeping rules")
	t.metrics.Counter(domain.MetricActivityExecutions, "Total activity executions").Add(1, nil)

	validationErrors := []string{}

	if len(entries) < 2 {
		validationErrors = append(validationErrors, "transaction must have at least 2 entries for double-entry bookkeeping")
	}

	// Calculate totals by entry type
	totalDebits := decimal.Zero
	totalCredits := decimal.Zero
	debitCount := 0
	creditCount := 0

	for _, entry := range entries {
		switch string(entry.EntryType) {
		case string(domain.NormalBalanceDebit):
			totalDebits = totalDebits.Add(entry.Amount)
			debitCount++
		case string(domain.NormalBalanceCredit):
			totalCredits = totalCredits.Add(entry.Amount)
			creditCount++
		default:
			validationErrors = append(validationErrors, "invalid entry type")
		}
	}

	// Validate double-entry balance
	if !totalDebits.Equal(totalCredits) {
		validationErrors = append(validationErrors, "total debits must equal total credits")
	}

	// Ensure both debit and credit entries exist
	if debitCount == 0 {
		validationErrors = append(validationErrors, "transaction must have at least one debit entry")
	}
	if creditCount == 0 {
		validationErrors = append(validationErrors, "transaction must have at least one credit entry")
	}

	if len(validationErrors) > 0 {
		activityLogger.WarnContext(ctx, "Double-entry validation failed", logger.Fields{
			"errors":        validationErrors,
			"total_debits":  totalDebits,
			"total_credits": totalCredits,
		})
		t.metrics.Counter(domain.MetricActivityFailures, "Double-entry validation errors").Add(1, nil)
		return &TransactionActivityOutput{
			Success:          false,
			Message:          "Double-entry validation failed",
			ErrorCode:        "DOUBLE_ENTRY_VALIDATION_FAILED",
			ValidationErrors: validationErrors,
		}, nil
	}

	activityLogger.InfoContext(ctx, "Double-entry validation successful", logger.Fields{
		"total_debits":  totalDebits,
		"total_credits": totalCredits,
	})
	t.metrics.Counter(domain.MetricValidationSuccess, "Double-entry validation success").Add(1, nil)

	return &TransactionActivityOutput{
		Success: true,
		Message: "Double-entry validation successful",
	}, nil
}

// CheckTransactionPermissionsActivity checks user permissions for transaction operations
func (t *TransactionActivities) CheckTransactionPermissionsActivity(ctx context.Context, transactionID *uuid.UUID, action string) (*TransactionActivityOutput, error) {
	ctx, span := t.tracer.StartSpan(ctx, domain.ActivityTypePermissionCheck)
	defer span.End()

	info := activity.GetInfo(ctx)
	activityLogger := t.logger.WithFields(logger.Fields{
		"activity_id":    info.ActivityID,
		"workflow_id":    info.WorkflowExecution.ID,
		"activity_type":  domain.ActivityTypePermissionCheck,
		"transaction_id": transactionID,
		"action":         action,
	})

	activityLogger.InfoContext(ctx, "Checking transaction permissions")
	t.metrics.Counter(domain.MetricActivityExecutions, "Total activity executions").Add(1, nil)

	// Extract user context
	userID := getUserIDFromContext(ctx)
	if userID == uuid.Nil {
		return &TransactionActivityOutput{
			Success:   false,
			Message:   "User context not found",
			ErrorCode: domain.ErrCodeUnauthorized,
		}, nil
	}

	// Check permissions via IAM service
	permReq := &authz.PermissionEvaluationRequest{
		UserID:       userID,
		ResourceType: "finance_transaction",
		ResourceID:   transactionID,
		Action:       action,
	}
	permResult, err := t.iamService.Authorization().EvaluatePermission(ctx, permReq)
	if err != nil {
		activityLogger.ErrorContext(ctx, "Permission check failed", logger.Fields{
			"error": err.Error(),
		})
		return &TransactionActivityOutput{
			Success:   false,
			Message:   "Permission check failed",
			ErrorCode: "PERMISSION_CHECK_FAILED",
		}, err
	}

	hasPermission := permResult.Decision == model.PolicyDecisionAllow
	if !hasPermission {
		activityLogger.WarnContext(ctx, "Permission denied")
		t.metrics.Counter(domain.MetricPermissionDenied, "Permission denied").Add(1, nil)
		return &TransactionActivityOutput{
			Success:   false,
			Message:   "Permission denied",
			ErrorCode: domain.ErrCodePermissionDenied,
		}, nil
	}

	activityLogger.InfoContext(ctx, "Permission check successful")
	t.metrics.Counter(domain.MetricPermissionGranted, "Permission granted").Add(1, nil)

	return &TransactionActivityOutput{
		Success: true,
		Message: "Permission granted",
	}, nil
}

// UpdateTransactionStatusActivity updates transaction status
func (t *TransactionActivities) UpdateTransactionStatusActivity(ctx context.Context, transactionID uuid.UUID, newStatus domain.TransactionStatus, reason string) (*TransactionActivityOutput, error) {
	ctx, span := t.tracer.StartSpan(ctx, "finance.activity.status.update")
	defer span.End()

	info := activity.GetInfo(ctx)
	activityLogger := t.logger.WithFields(logger.Fields{
		"activity_id":    info.ActivityID,
		"workflow_id":    info.WorkflowExecution.ID,
		"activity_type":  "finance.activity.status.update",
		"transaction_id": transactionID,
		"new_status":     newStatus,
		"reason":         reason,
	})

	activityLogger.InfoContext(ctx, "Updating transaction status")
	t.metrics.Counter(domain.MetricActivityExecutions, "Total activity executions").Add(1, nil)

	// Get current transaction
	transaction, err := t.transactionService.GetTransactionByID(ctx, transactionID)
	if err != nil {
		activityLogger.ErrorContext(ctx, "Failed to get transaction", logger.Fields{
			"error": err.Error(),
		})
		return &TransactionActivityOutput{
			Success:   false,
			Message:   "Failed to get transaction",
			ErrorCode: "TRANSACTION_NOT_FOUND",
		}, err
	}

	// Update transaction status
	transaction.TransactionStatus = newStatus
	updatedTransaction, err := t.transactionService.UpdateTransaction(ctx, transactionID, *transaction)
	if err != nil {
		activityLogger.ErrorContext(ctx, "Failed to update transaction status", logger.Fields{
			"error": err.Error(),
		})
		return &TransactionActivityOutput{
			Success:   false,
			Message:   "Transaction status update failed",
			ErrorCode: "STATUS_UPDATE_FAILED",
		}, err
	}

	activityLogger.InfoContext(ctx, "Transaction status updated successfully")
	t.metrics.Counter(domain.MetricActivityExecutions, "Transaction status updated").Add(1, nil)

	return &TransactionActivityOutput{
		Transaction: updatedTransaction,
		Success:     true,
		Message:     "Transaction status updated successfully",
	}, nil
}

// CacheTransactionActivity caches transaction data
func (t *TransactionActivities) CacheTransactionActivity(ctx context.Context, transaction *domain.Transaction, ttl int) (*TransactionActivityOutput, error) {
	ctx, span := t.tracer.StartSpan(ctx, domain.ActivityTypeCacheOperation)
	defer span.End()

	info := activity.GetInfo(ctx)
	activityLogger := t.logger.WithFields(logger.Fields{
		"activity_id":    info.ActivityID,
		"workflow_id":    info.WorkflowExecution.ID,
		"activity_type":  domain.ActivityTypeCacheOperation,
		"transaction_id": transaction.ID,
	})

	activityLogger.InfoContext(ctx, "Caching transaction data")

	// Generate cache key with tenant and entity context
	tenantID := getTenantIDFromContext(ctx)
	entityID := getEntityIDFromContext(ctx)
	cacheKey := domain.GetTransactionCacheKey(tenantID.String(), entityID.String(), transaction.ID.String())

	err := t.cacheService.Set(ctx, cacheKey, transaction, time.Duration(ttl)*time.Second)
	if err != nil {
		activityLogger.ErrorContext(ctx, "Failed to cache transaction", logger.Fields{
			"error": err.Error(),
		})
		return &TransactionActivityOutput{
			Success:   false,
			Message:   "Cache operation failed",
			ErrorCode: domain.ErrCodeCacheOperationFailed,
		}, err
	}

	activityLogger.InfoContext(ctx, "Transaction cached successfully")
	return &TransactionActivityOutput{
		Success: true,
		Message: "Transaction cached successfully",
	}, nil
}

// ProcessBulkTransactionsActivity processes multiple transactions in bulk
func (t *TransactionActivities) ProcessBulkTransactionsActivity(ctx context.Context, input BulkTransactionActivityInput) (*TransactionActivityOutput, error) {
	ctx, span := t.tracer.StartSpan(ctx, "finance.activity.bulk.processing")
	defer span.End()

	info := activity.GetInfo(ctx)
	activityLogger := t.logger.WithFields(logger.Fields{
		"activity_id":        info.ActivityID,
		"workflow_id":        info.WorkflowExecution.ID,
		"activity_type":      "finance.activity.bulk.processing",
		"transactions_count": len(input.Transactions),
		"batch_size":         input.BatchSize,
		"concurrent_limit":   input.ConcurrentLimit,
	})

	activityLogger.InfoContext(ctx, "Processing bulk transactions")
	t.metrics.Counter(domain.MetricActivityExecutions, "Total activity executions").Add(1, nil)

	successCount := 0
	errorCount := 0
	validationErrors := []string{}

	// Process transactions in batches
	for i := 0; i < len(input.Transactions); i += input.BatchSize {
		end := i + input.BatchSize
		if end > len(input.Transactions) {
			end = len(input.Transactions)
		}

		batch := input.Transactions[i:end]
		activityLogger.InfoContext(ctx, "Processing transaction batch", logger.Fields{
			"batch_start": i,
			"batch_end":   end,
			"batch_size":  len(batch),
		})

		for _, txInput := range batch {
			// Validate transaction
			validationResult, err := t.ValidateTransactionActivity(ctx, txInput)
			if err != nil || !validationResult.Success {
				errorCount++
				if validationResult != nil {
					validationErrors = append(validationErrors, validationResult.ValidationErrors...)
				}
				continue
			}

			// Create transaction
			createResult, err := t.CreateTransactionActivity(ctx, txInput)
			if err != nil || !createResult.Success {
				errorCount++
				continue
			}

			successCount++
		}
	}

	activityLogger.InfoContext(ctx, "Bulk transaction processing completed", logger.Fields{
		"success_count": successCount,
		"error_count":   errorCount,
		"total_count":   len(input.Transactions),
	})

	t.metrics.Counter(domain.MetricTransactionsProcessed, "Bulk transactions processed").Add(float64(successCount), nil)
	t.metrics.Counter(domain.MetricActivityFailures, "Bulk transaction errors").Add(float64(errorCount), nil)

	return &TransactionActivityOutput{
		Success:          successCount > 0,
		Message:          "Bulk transaction processing completed",
		ValidationErrors: validationErrors,
	}, nil
}
