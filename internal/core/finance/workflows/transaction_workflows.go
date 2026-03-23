package workflows

import (
	"time"

	"awo.so/internal/core/audit"
	"awo.so/internal/core/featureflag"
	"awo.so/internal/core/finance/domain"
	"awo.so/internal/core/finance/service"
	"awo.so/internal/core/iam"
	"awo.so/internal/core/notification"
	settingsService "awo.so/internal/core/settings/service"
	"awo.so/internal/platform/cache"
	loggerPkg "awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// TransactionWorkflows contains all transaction-related Temporal workflows
type TransactionWorkflows struct {
	deps TransactionWorkflowDeps
}

// TransactionWorkflowDeps contains dependencies for transaction workflows
type TransactionWorkflowDeps struct {
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

// NewTransactionWorkflows creates new transaction workflows
func NewTransactionWorkflows(deps TransactionWorkflowDeps) *TransactionWorkflows {
	return &TransactionWorkflows{
		deps: deps,
	}
}

// TransactionApprovalWorkflow handles the complete approval process for transactions
func (tw *TransactionWorkflows) TransactionApprovalWorkflow(ctx workflow.Context, input domain.TransactionApprovalWorkflowInput) (*domain.TransactionApprovalWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting transaction approval workflow", "transaction_id", input.TransactionID)

	// Set workflow options with timeouts from domain constants
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 5,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
			InitialInterval: time.Second * 1,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	result := &domain.TransactionApprovalWorkflowResult{
		TransactionID: input.TransactionID,
		StartTime:     workflow.Now(ctx),
	}

	// Step 1: Validate transaction request
	var validationResult domain.TransactionValidationResult
	err := workflow.ExecuteActivity(ctx, "ValidateTransactionRequest", domain.TransactionValidationInput{
		TransactionID:  input.TransactionID,
		ValidationType: domain.ValidationTypeApproval,
	}).Get(ctx, &validationResult)
	if err != nil {
		logger.Error("Transaction validation failed", "error", err)
		result.Status = domain.ApprovalStatusRejected
		result.Error = err.Error()
		return result, nil
	}

	if !validationResult.IsValid {
		logger.Info("Transaction validation failed", "reasons", validationResult.ValidationErrors)
		result.Status = domain.ApprovalStatusRejected
		result.ValidationErrors = validationResult.ValidationErrors
		return result, nil
	}

	// Step 2: Determine approval requirements
	var approvalRequirements domain.ApprovalRequirementsResult
	err = workflow.ExecuteActivity(ctx, "DetermineApprovalRequirements", domain.ApprovalRequirementsInput{
		TransactionID: input.TransactionID,
		Amount:        input.Amount,
		AccountType:   input.AccountType,
	}).Get(ctx, &approvalRequirements)
	if err != nil {
		logger.Error("Failed to determine approval requirements", "error", err)
		result.Status = domain.ApprovalStatusRejected
		result.Error = err.Error()
		return result, nil
	}

	// Step 3: Process approvals based on requirements
	if len(approvalRequirements.RequiredApprovers) == 0 {
		// Auto-approve if no approvers required
		result.Status = domain.ApprovalStatusApproved
		result.ApprovedBy = []string{"system"}
		logger.Info("Transaction auto-approved (no approvers required)")
	} else {
		// Manual approval process
		approvalResult := tw.processManualApproval(ctx, input, approvalRequirements)
		result.Status = approvalResult.Status
		result.ApprovedBy = approvalResult.ApprovedBy
		result.RejectedBy = approvalResult.RejectedBy
		result.Comments = approvalResult.Comments
	}

	// Step 4: Send notifications
	err = workflow.ExecuteActivity(ctx, "SendApprovalNotification", domain.ApprovalNotificationInput{
		TransactionID: input.TransactionID,
		Status:        result.Status,
		Recipients:    append(approvalRequirements.RequiredApprovers, input.SubmittedBy),
		Comments:      result.Comments,
	}).Get(ctx, nil)
	if err != nil {
		logger.Warn("Failed to send approval notification", "error", err)
		// Don't fail the workflow for notification errors
	}

	result.CompletedTime = workflow.Now(ctx)
	result.Duration = result.CompletedTime.Sub(result.StartTime)

	logger.Info("Transaction approval workflow completed",
		"transaction_id", input.TransactionID,
		"status", result.Status,
		"duration", result.Duration)

	return result, nil
}

// processManualApproval handles the manual approval process
func (tw *TransactionWorkflows) processManualApproval(ctx workflow.Context, input domain.TransactionApprovalWorkflowInput, requirements domain.ApprovalRequirementsResult) domain.ApprovalProcessResult {
	logger := workflow.GetLogger(ctx)

	result := domain.ApprovalProcessResult{
		Status: domain.ApprovalStatusPending,
	}

	// Wait for approvals with timeout
	approvalTimeout := time.Hour * 24 // Default 24 hours
	if requirements.ApprovalTimeout > 0 {
		approvalTimeout = requirements.ApprovalTimeout
	}

	selector := workflow.NewSelector(ctx)
	approvalChannel := workflow.GetSignalChannel(ctx, "approval-signal")
	rejectionChannel := workflow.GetSignalChannel(ctx, "rejection-signal")

	approvedCount := 0
	requiredApprovals := requirements.RequiredApprovalCount
	if requiredApprovals == 0 {
		requiredApprovals = len(requirements.RequiredApprovers)
	}

	// Set up signal handlers
	selector.AddReceive(approvalChannel, func(c workflow.ReceiveChannel, more bool) {
		var approval domain.ApprovalSignal
		c.Receive(ctx, &approval)

		logger.Info("Received approval", "approver", approval.ApproverID, "transaction_id", input.TransactionID)
		result.ApprovedBy = append(result.ApprovedBy, approval.ApproverID)
		result.Comments = append(result.Comments, approval.Comment)
		approvedCount++

		if approvedCount >= requiredApprovals {
			result.Status = domain.ApprovalStatusApproved
		}
	})

	selector.AddReceive(rejectionChannel, func(c workflow.ReceiveChannel, more bool) {
		var rejection domain.RejectionSignal
		c.Receive(ctx, &rejection)

		logger.Info("Received rejection", "rejector", rejection.RejectorID, "transaction_id", input.TransactionID)
		result.RejectedBy = append(result.RejectedBy, rejection.RejectorID)
		result.Comments = append(result.Comments, rejection.Comment)
		result.Status = domain.ApprovalStatusRejected
	})

	// Wait for completion or timeout
	for result.Status == domain.ApprovalStatusPending {
		selector.Select(ctx)

		// Check timeout
		if workflow.Now(ctx).Sub(input.SubmittedAt) > approvalTimeout {
			logger.Warn("Approval workflow timed out", "transaction_id", input.TransactionID)
			result.Status = domain.ApprovalStatusExpired
			break
		}

		// Check if we have enough approvals or any rejection
		if result.Status != domain.ApprovalStatusPending {
			break
		}
	}

	return result
}

// TransactionProcessingWorkflow handles the complete transaction processing
func (tw *TransactionWorkflows) TransactionProcessingWorkflow(ctx workflow.Context, input domain.TransactionProcessingWorkflowInput) (*domain.TransactionProcessingWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting transaction processing workflow", "transaction_id", input.TransactionID)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 10,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
			InitialInterval: time.Second * 2,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	result := &domain.TransactionProcessingWorkflowResult{
		TransactionID: input.TransactionID,
		StartTime:     workflow.Now(ctx),
	}

	// Step 1: Validate double-entry
	var doubleEntryResult domain.DoubleEntryValidationResult
	err := workflow.ExecuteActivity(ctx, "ValidateDoubleEntry", domain.DoubleEntryValidationInput{
		TransactionID: input.TransactionID,
	}).Get(ctx, &doubleEntryResult)

	if err != nil || !doubleEntryResult.IsValid {
		logger.Error("Double-entry validation failed", "error", err)
		result.Status = domain.ProcessingStatusFailed
		result.Error = err.Error()
		return result, nil
	}

	// Step 2: Post to ledger
	var postingResult domain.LedgerPostingResult
	err = workflow.ExecuteActivity(ctx, "PostTransactionToLedger", domain.LedgerPostingInput{
		TransactionID: input.TransactionID,
	}).Get(ctx, &postingResult)
	if err != nil {
		logger.Error("Ledger posting failed", "error", err)
		result.Status = domain.ProcessingStatusFailed
		result.Error = err.Error()
		return result, nil
	}

	// Step 3: Update account balances
	err = workflow.ExecuteActivity(ctx, "UpdateAccountBalance", domain.BalanceUpdateInput{
		TransactionID: input.TransactionID,
		Entries:       postingResult.ProcessedEntries,
	}).Get(ctx, nil)
	if err != nil {
		logger.Error("Balance update failed", "error", err)
		// Attempt to reverse the posting
		workflow.ExecuteActivity(ctx, "ReverseTransaction", domain.TransactionReversalInput{
			OriginalTransactionID: input.TransactionID,
			ReversalReason:        "Balance update failed",
		}).Get(ctx, nil)

		result.Status = domain.ProcessingStatusFailed
		result.Error = err.Error()
		return result, nil
	}

	result.Status = domain.ProcessingStatusCompleted
	result.PostingReference = postingResult.PostingReference
	result.CompletedTime = workflow.Now(ctx)
	result.Duration = result.CompletedTime.Sub(result.StartTime)

	logger.Info("Transaction processing workflow completed successfully",
		"transaction_id", input.TransactionID,
		"posting_ref", postingResult.PostingReference,
		"duration", result.Duration)

	return result, nil
}

// TransactionReversalWorkflow handles transaction reversals
func (tw *TransactionWorkflows) TransactionReversalWorkflow(ctx workflow.Context, input domain.TransactionReversalWorkflowInput) (*domain.TransactionReversalWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting transaction reversal workflow", "original_transaction_id", input.OriginalTransactionID)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 5,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
			InitialInterval: time.Second * 1,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	result := &domain.TransactionReversalWorkflowResult{
		OriginalTransactionID: input.OriginalTransactionID,
		StartTime:             workflow.Now(ctx),
	}

	// Step 1: Validate reversal eligibility
	var validationResult domain.ReversalValidationResult
	err := workflow.ExecuteActivity(ctx, "ValidateReversalEligibility", domain.ReversalValidationInput{
		TransactionID: input.OriginalTransactionID,
		Reason:        input.ReversalReason,
	}).Get(ctx, &validationResult)

	if err != nil || !validationResult.IsEligible {
		logger.Error("Reversal validation failed", "error", err)
		result.Status = domain.ReversalStatusRejected
		result.Error = err.Error()
		return result, nil
	}

	// Step 2: Create reversal transaction
	var reversalResult domain.ReversalCreationResult
	err = workflow.ExecuteActivity(ctx, "CreateReversalTransaction", domain.ReversalCreationInput{
		OriginalTransactionID: input.OriginalTransactionID,
		ReversalReason:        input.ReversalReason,
		InitiatedBy:           input.InitiatedBy,
	}).Get(ctx, &reversalResult)
	if err != nil {
		logger.Error("Reversal creation failed", "error", err)
		result.Status = domain.ReversalStatusFailed
		result.Error = err.Error()
		return result, nil
	}

	result.ReversalTransactionID = reversalResult.ReversalTransactionID

	// Step 3: Process the reversal transaction
	processResult, err := tw.TransactionProcessingWorkflow(ctx, domain.TransactionProcessingWorkflowInput{
		TransactionID: reversalResult.ReversalTransactionID,
	})

	if err != nil || processResult.Status != domain.ProcessingStatusCompleted {
		logger.Error("Reversal processing failed", "error", err)
		result.Status = domain.ReversalStatusFailed
		result.Error = err.Error()
		return result, nil
	}

	result.Status = domain.ReversalStatusCompleted
	result.PostingReference = processResult.PostingReference
	result.CompletedTime = workflow.Now(ctx)
	result.Duration = result.CompletedTime.Sub(result.StartTime)

	logger.Info("Transaction reversal workflow completed successfully",
		"original_transaction_id", input.OriginalTransactionID,
		"reversal_transaction_id", result.ReversalTransactionID,
		"duration", result.Duration)

	return result, nil
}

// BulkTransactionWorkflow handles bulk transaction processing
func (tw *TransactionWorkflows) BulkTransactionWorkflow(ctx workflow.Context, input domain.BulkTransactionWorkflowInput) (*domain.BulkTransactionWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting bulk transaction workflow", "batch_id", input.BatchID, "transaction_count", len(input.TransactionIDs))

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 30, // Longer timeout for bulk operations
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 2,
			InitialInterval: time.Second * 5,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	result := &domain.BulkTransactionWorkflowResult{
		BatchID:   input.BatchID,
		StartTime: workflow.Now(ctx),
		Results:   make(map[string]domain.TransactionProcessingWorkflowResult),
	}

	// Process transactions in parallel with controlled concurrency
	concurrency := 10 // Process up to 10 transactions concurrently
	if input.Concurrency > 0 {
		concurrency = input.Concurrency
	}

	// Create child workflows for each transaction
	childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: time.Minute * 15,
	})

	futures := make([]workflow.ChildWorkflowFuture, len(input.TransactionIDs))
	semaphore := workflow.NewSemaphore(childCtx, int64(concurrency))

	for i, transactionID := range input.TransactionIDs {
		// Acquire semaphore to limit concurrency
		err := semaphore.Acquire(childCtx, 1)
		if err != nil {
			logger.Error("Failed to acquire semaphore", "error", err)
			continue
		}

		futures[i] = workflow.ExecuteChildWorkflow(childCtx, tw.TransactionProcessingWorkflow, domain.TransactionProcessingWorkflowInput{
			TransactionID: transactionID,
		})

		// Release semaphore when done
		workflow.Go(childCtx, func(ctx workflow.Context) {
			defer semaphore.Release(1)
			var res domain.TransactionProcessingWorkflowResult
			err := futures[i].Get(ctx, &res)
			if err != nil {
				logger.Error("Child workflow failed", "transaction_id", transactionID, "error", err)
				result.FailedTransactions = append(result.FailedTransactions, transactionID)
			} else {
				result.Results[transactionID.String()] = res
				if res.Status == domain.ProcessingStatusCompleted {
					result.SuccessfulTransactions = append(result.SuccessfulTransactions, transactionID)
				} else {
					result.FailedTransactions = append(result.FailedTransactions, transactionID)
				}
			}
		})
	}

	// Wait for all child workflows to complete
	for _, future := range futures {
		if future != nil {
			future.Get(childCtx, nil) // Wait for completion
		}
	}

	result.CompletedTime = workflow.Now(ctx)
	result.Duration = result.CompletedTime.Sub(result.StartTime)
	result.TotalProcessed = len(result.SuccessfulTransactions) + len(result.FailedTransactions)
	result.SuccessRate = float64(len(result.SuccessfulTransactions)) / float64(len(input.TransactionIDs)) * 100

	logger.Info("Bulk transaction workflow completed",
		"batch_id", input.BatchID,
		"total", len(input.TransactionIDs),
		"successful", len(result.SuccessfulTransactions),
		"failed", len(result.FailedTransactions),
		"success_rate", result.SuccessRate,
		"duration", result.Duration)

	return result, nil
}
