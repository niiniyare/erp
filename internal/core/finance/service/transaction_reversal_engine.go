package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.opentelemetry.io/otel/attribute"

	"github.com/niiniyare/erp/internal/core/finance/domain"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// TransactionReversalEngine handles the reversal of posted financial transactions
type TransactionReversalEngine interface {
	// CreateReversalTransaction creates a reversal transaction for a posted transaction
	CreateReversalTransaction(ctx context.Context, req CreateReversalRequest) (*ReversalTransactionResult, error)

	// ProcessCompleteReversal processes a complete reversal (original + reversal transactions)
	ProcessCompleteReversal(ctx context.Context, req CompleteReversalRequest) (*CompleteReversalResult, error)

	// ValidateReversalEligibility validates if a transaction can be reversed
	ValidateReversalEligibility(ctx context.Context, transactionID uuid.UUID) (*ReversalValidationResult, error)

	// GetReversalHistory gets the reversal history for a transaction
	GetReversalHistory(ctx context.Context, transactionID uuid.UUID) (*ReversalHistoryResult, error)

	// CreateCorrectionEntry creates a correction entry for partial corrections
	CreateCorrectionEntry(ctx context.Context, req CorrectionEntryRequest) (*CorrectionEntryResult, error)

	// BatchReverseTransactions reverses multiple transactions in a batch
	BatchReverseTransactions(ctx context.Context, req BatchReversalRequest) (*BatchReversalResult, error)
}

// CreateReversalRequest represents a request to create a reversal transaction
type CreateReversalRequest struct {
	OriginalTransactionID uuid.UUID    `json:"original_transaction_id"`
	ReversedBy            uuid.UUID    `json:"reversed_by"`
	ReversalReason        string       `json:"reversal_reason"`
	ReversalDate          *time.Time   `json:"reversal_date,omitempty"` // If nil, uses current date
	ReversalType          ReversalType `json:"reversal_type"`
	CustomReversalNumber  *string      `json:"custom_reversal_number,omitempty"`
	AutoPost              bool         `json:"auto_post"`              // Automatically post the reversal
	CopyOriginalMetadata  bool         `json:"copy_original_metadata"` // Copy metadata from original
}

// CompleteReversalRequest represents a request for complete reversal processing
type CompleteReversalRequest struct {
	OriginalTransactionID uuid.UUID  `json:"original_transaction_id"`
	ReversedBy            uuid.UUID  `json:"reversed_by"`
	ReversalReason        string     `json:"reversal_reason"`
	ReversalDate          *time.Time `json:"reversal_date,omitempty"`
	AutoPost              bool       `json:"auto_post"`
}

// CorrectionEntryRequest represents a request to create a correction entry
type CorrectionEntryRequest struct {
	OriginalTransactionID uuid.UUID                   `json:"original_transaction_id"`
	CorrectionEntries     []CorrectionEntryDefinition `json:"correction_entries"`
	CorrectedBy           uuid.UUID                   `json:"corrected_by"`
	CorrectionReason      string                      `json:"correction_reason"`
	CorrectionDate        *time.Time                  `json:"correction_date,omitempty"`
	AutoPost              bool                        `json:"auto_post"`
}

// BatchReversalRequest represents a batch reversal request
type BatchReversalRequest struct {
	TransactionIDs  []uuid.UUID `json:"transaction_ids"`
	ReversedBy      uuid.UUID   `json:"reversed_by"`
	ReversalReason  string      `json:"reversal_reason"`
	ReversalDate    *time.Time  `json:"reversal_date,omitempty"`
	ContinueOnError bool        `json:"continue_on_error"`
	AutoPost        bool        `json:"auto_post"`
}

// ReversalType represents the type of reversal
type ReversalType string

const (
	ReversalTypeFull       ReversalType = "FULL"       // Full reversal of entire transaction
	ReversalTypePartial    ReversalType = "PARTIAL"    // Partial reversal of specific entries
	ReversalTypeCorrection ReversalType = "CORRECTION" // Correction with new values
)

// CorrectionEntryDefinition defines a correction entry
type CorrectionEntryDefinition struct {
	OriginalEntryID uuid.UUID       `json:"original_entry_id"`
	AccountID       uuid.UUID       `json:"account_id"`
	NewDebitAmount  decimal.Decimal `json:"new_debit_amount"`
	NewCreditAmount decimal.Decimal `json:"new_credit_amount"`
	CorrectionType  CorrectionType  `json:"correction_type"`
	Description     string          `json:"description"`
}

// CorrectionType represents the type of correction
type CorrectionType string

const (
	CorrectionTypeAmount      CorrectionType = "AMOUNT"      // Amount correction
	CorrectionTypeAccount     CorrectionType = "ACCOUNT"     // Account correction
	CorrectionTypeDescription CorrectionType = "DESCRIPTION" // Description correction
)

// ReversalTransactionResult represents the result of creating a reversal transaction
type ReversalTransactionResult struct {
	ReversalTransaction *domain.Transaction       `json:"reversal_transaction"`
	OriginalTransaction *domain.Transaction       `json:"original_transaction"`
	Success             bool                      `json:"success"`
	PostingResult       *PostTransactionResult    `json:"posting_result,omitempty"`
	ValidationResult    *ReversalValidationResult `json:"validation_result,omitempty"`
	Errors              []string                  `json:"errors,omitempty"`
	Warnings            []string                  `json:"warnings,omitempty"`
}

// CompleteReversalResult represents the result of complete reversal processing
type CompleteReversalResult struct {
	OriginalTransactionID uuid.UUID              `json:"original_transaction_id"`
	ReversalTransactionID uuid.UUID              `json:"reversal_transaction_id"`
	Success               bool                   `json:"success"`
	ProcessingTime        time.Duration          `json:"processing_time"`
	BalanceUpdates        []AccountBalanceUpdate `json:"balance_updates"`
	ReversalTransaction   *domain.Transaction    `json:"reversal_transaction"`
	Errors                []string               `json:"errors,omitempty"`
	Warnings              []string               `json:"warnings,omitempty"`
}

// CorrectionEntryResult represents the result of creating correction entries
type CorrectionEntryResult struct {
	CorrectionTransaction *domain.Transaction    `json:"correction_transaction"`
	Success               bool                   `json:"success"`
	PostingResult         *PostTransactionResult `json:"posting_result,omitempty"`
	Errors                []string               `json:"errors,omitempty"`
}

// BatchReversalResult represents the result of batch reversal
type BatchReversalResult struct {
	TotalTransactions   int                         `json:"total_transactions"`
	SuccessfulReversals int                         `json:"successful_reversals"`
	FailedReversals     int                         `json:"failed_reversals"`
	Results             []ReversalTransactionResult `json:"results"`
	ProcessingTime      time.Duration               `json:"processing_time"`
}

// ReversalValidationResult represents validation results for reversal
type ReversalValidationResult struct {
	CanReverse          bool                 `json:"can_reverse"`
	ValidationLevel     ValidationLevel      `json:"validation_level"`
	Errors              []ValidationError    `json:"errors,omitempty"`
	Warnings            []ValidationWarning  `json:"warnings,omitempty"`
	ReversalConstraints []ReversalConstraint `json:"reversal_constraints,omitempty"`
}

// ReversalConstraint represents a constraint that affects reversal eligibility
type ReversalConstraint struct {
	Type        ReversalConstraintType `json:"type"`
	Description string                 `json:"description"`
	Blocking    bool                   `json:"blocking"` // If true, prevents reversal
}

// ReversalConstraintType represents types of reversal constraints
type ReversalConstraintType string

const (
	ReversalConstraintPeriodClosed    ReversalConstraintType = "PERIOD_CLOSED"
	ReversalConstraintAlreadyReversed ReversalConstraintType = "ALREADY_REVERSED"
	ReversalConstraintNotPosted       ReversalConstraintType = "NOT_POSTED"
	ReversalConstraintHasReversals    ReversalConstraintType = "HAS_REVERSALS"
	ReversalConstraintTimeLimit       ReversalConstraintType = "TIME_LIMIT"
)

// ReversalHistoryResult represents reversal history for a transaction
type ReversalHistoryResult struct {
	TransactionID uuid.UUID              `json:"transaction_id"`
	ReversalChain []ReversalRecord       `json:"reversal_chain"`
	HasReversals  bool                   `json:"has_reversals"`
	IsReversed    bool                   `json:"is_reversed"`
	NetEffect     []AccountBalanceUpdate `json:"net_effect"`
}

// ReversalRecord represents a single reversal record
type ReversalRecord struct {
	ReversalID            uuid.UUID    `json:"reversal_id"`
	ReversalTransactionID uuid.UUID    `json:"reversal_transaction_id"`
	ReversalDate          time.Time    `json:"reversal_date"`
	ReversedBy            uuid.UUID    `json:"reversed_by"`
	ReversalReason        string       `json:"reversal_reason"`
	ReversalType          ReversalType `json:"reversal_type"`
	Status                string       `json:"status"`
}

// transactionReversalEngine implements TransactionReversalEngine
type transactionReversalEngine struct {
	transactionRepository domain.TransactionRepository
	numberingService      TransactionNumberingService
	postingEngine         TransactionPostingEngine
	validator             DoubleEntryValidator
	tracing               tracing.Service
}

// TransactionReversalEngineDeps represents dependencies for the reversal engine
type TransactionReversalEngineDeps struct {
	TransactionRepository domain.TransactionRepository
	NumberingService      TransactionNumberingService
	PostingEngine         TransactionPostingEngine
	Validator             DoubleEntryValidator
	Tracing               tracing.Service
}

// NewTransactionReversalEngine creates a new transaction reversal engine
func NewTransactionReversalEngine(deps TransactionReversalEngineDeps) TransactionReversalEngine {
	return &transactionReversalEngine{
		transactionRepository: deps.TransactionRepository,
		numberingService:      deps.NumberingService,
		postingEngine:         deps.PostingEngine,
		validator:             deps.Validator,
		tracing:               deps.Tracing,
	}
}

// CreateReversalTransaction creates a reversal transaction for a posted transaction
func (e *transactionReversalEngine) CreateReversalTransaction(ctx context.Context, req CreateReversalRequest) (*ReversalTransactionResult, error) {
	ctx, span := e.tracing.StartSpan(ctx, "TransactionReversalEngine.CreateReversalTransaction")
	defer span.End()

	result := &ReversalTransactionResult{
		Success:  false,
		Errors:   []string{},
		Warnings: []string{},
	}

	span.SetAttributes(
		attribute.String("original_transaction_id", req.OriginalTransactionID.String()),
		attribute.String("reversed_by", req.ReversedBy.String()),
		attribute.String("reversal_type", string(req.ReversalType)),
	)

	// 1. Load and validate the original transaction
	originalTransaction, err := e.transactionRepository.GetByID(ctx, req.OriginalTransactionID)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to load original transaction: %v", err))
		return result, err
	}

	result.OriginalTransaction = originalTransaction

	// 2. Validate reversal eligibility
	validationResult, err := e.ValidateReversalEligibility(ctx, req.OriginalTransactionID)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to validate reversal eligibility: %v", err))
		return result, err
	}

	result.ValidationResult = validationResult

	if !validationResult.CanReverse {
		result.Errors = append(result.Errors, "Transaction cannot be reversed due to validation errors")
		for _, validationErr := range validationResult.Errors {
			result.Errors = append(result.Errors, validationErr.Message)
		}
		return result, fmt.Errorf("transaction cannot be reversed")
	}

	// 3. Create the reversal transaction
	reversalTransaction, err := e.createReversalFromOriginal(ctx, originalTransaction, req)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to create reversal transaction: %v", err))
		return result, err
	}

	result.ReversalTransaction = reversalTransaction

	// 4. Save the reversal transaction
	err = e.transactionRepository.Create(ctx, reversalTransaction)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to save reversal transaction: %v", err))
		return result, err
	}

	// 5. Update the original transaction to mark as reversed
	originalTransaction.IsReversed = true
	originalTransaction.ReversedByTransactionID = &reversalTransaction.ID
	originalTransaction.ReversalReason = &req.ReversalReason

	err = e.transactionRepository.Update(ctx, originalTransaction)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to update original transaction: %v", err))
		return result, err
	}

	// 6. Auto-post the reversal if requested
	if req.AutoPost {
		postReq := PostTransactionRequest{
			TransactionID:         reversalTransaction.ID,
			PostedBy:              req.ReversedBy,
			PostingDate:           req.ReversalDate,
			ForcePost:             false,
			ValidateBeforePosting: true,
		}

		postResult, err := e.postingEngine.PostTransaction(ctx, postReq)
		if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to auto-post reversal: %v", err))
		} else {
			result.PostingResult = postResult
		}
	}

	result.Success = true

	span.SetAttributes(
		attribute.String("reversal_transaction_id", reversalTransaction.ID.String()),
		attribute.Bool("auto_posted", req.AutoPost),
		attribute.Bool("success", result.Success),
	)

	return result, nil
}

// ProcessCompleteReversal processes a complete reversal (original + reversal transactions)
func (e *transactionReversalEngine) ProcessCompleteReversal(ctx context.Context, req CompleteReversalRequest) (*CompleteReversalResult, error) {
	ctx, span := e.tracing.StartSpan(ctx, "TransactionReversalEngine.ProcessCompleteReversal")
	defer span.End()

	startTime := time.Now()

	result := &CompleteReversalResult{
		OriginalTransactionID: req.OriginalTransactionID,
		Success:               false,
		BalanceUpdates:        []AccountBalanceUpdate{},
		Errors:                []string{},
		Warnings:              []string{},
	}

	// 1. Create reversal transaction
	createReq := CreateReversalRequest{
		OriginalTransactionID: req.OriginalTransactionID,
		ReversedBy:            req.ReversedBy,
		ReversalReason:        req.ReversalReason,
		ReversalDate:          req.ReversalDate,
		ReversalType:          ReversalTypeFull,
		AutoPost:              req.AutoPost,
		CopyOriginalMetadata:  true,
	}

	reversalResult, err := e.CreateReversalTransaction(ctx, createReq)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to create reversal: %v", err))
		return result, err
	}

	result.ReversalTransactionID = reversalResult.ReversalTransaction.ID
	result.ReversalTransaction = reversalResult.ReversalTransaction

	// 2. Collect balance updates from posting if auto-posted
	if reversalResult.PostingResult != nil {
		result.BalanceUpdates = reversalResult.PostingResult.BalanceUpdates
	}

	// 3. Merge errors and warnings
	result.Errors = append(result.Errors, reversalResult.Errors...)
	result.Warnings = append(result.Warnings, reversalResult.Warnings...)

	result.Success = reversalResult.Success
	result.ProcessingTime = time.Since(startTime)

	return result, nil
}

// ValidateReversalEligibility validates if a transaction can be reversed
func (e *transactionReversalEngine) ValidateReversalEligibility(ctx context.Context, transactionID uuid.UUID) (*ReversalValidationResult, error) {
	ctx, span := e.tracing.StartSpan(ctx, "TransactionReversalEngine.ValidateReversalEligibility")
	defer span.End()

	result := &ReversalValidationResult{
		CanReverse:          true,
		ValidationLevel:     ValidationLevelValid,
		Errors:              []ValidationError{},
		Warnings:            []ValidationWarning{},
		ReversalConstraints: []ReversalConstraint{},
	}

	// Load the transaction
	transaction, err := e.transactionRepository.GetByID(ctx, transactionID)
	if err != nil {
		result.CanReverse = false
		result.Errors = append(result.Errors, ValidationError{
			Field:    "transaction_id",
			Message:  "Transaction not found",
			Code:     "TRANSACTION_NOT_FOUND",
			Severity: "ERROR",
			Value:    transactionID,
		})
		return result, nil
	}

	// 1. Check if transaction is posted
	if transaction.TransactionStatus != domain.TransactionStatusPosted {
		result.CanReverse = false
		result.Errors = append(result.Errors, ValidationError{
			Field:    "transaction_status",
			Message:  "Only posted transactions can be reversed",
			Code:     "TRANSACTION_NOT_POSTED",
			Severity: "ERROR",
			Value:    transaction.TransactionStatus,
		})
		result.ReversalConstraints = append(result.ReversalConstraints, ReversalConstraint{
			Type:        ReversalConstraintNotPosted,
			Description: "Transaction must be posted to be reversed",
			Blocking:    true,
		})
	}

	// 2. Check if already reversed
	if transaction.IsReversed {
		result.CanReverse = false
		result.Errors = append(result.Errors, ValidationError{
			Field:    "is_reversed",
			Message:  "Transaction has already been reversed",
			Code:     "ALREADY_REVERSED",
			Severity: "ERROR",
			Value:    transaction.IsReversed,
		})
		result.ReversalConstraints = append(result.ReversalConstraints, ReversalConstraint{
			Type:        ReversalConstraintAlreadyReversed,
			Description: "Transaction has already been reversed",
			Blocking:    true,
		})
	}

	// 3. Check posting period (simplified - in production, check actual accounting periods)
	if transaction.PostingDate != nil {
		// Check if more than 30 days old (example business rule)
		daysSincePosting := time.Since(*transaction.PostingDate).Hours() / 24
		if daysSincePosting > 30 {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:   "posting_date",
				Message: "Transaction is more than 30 days old",
				Code:    "OLD_TRANSACTION",
				Value:   transaction.PostingDate,
			})
			result.ReversalConstraints = append(result.ReversalConstraints, ReversalConstraint{
				Type:        ReversalConstraintTimeLimit,
				Description: fmt.Sprintf("Transaction is %.0f days old", daysSincePosting),
				Blocking:    false,
			})
		}
	}

	// 4. Check for existing reversals (this transaction reversing others)
	// TODO: Implement check for whether this transaction has already reversed others

	// Set final validation level
	if len(result.Errors) > 0 {
		result.CanReverse = false
		result.ValidationLevel = ValidationLevelError
	} else if len(result.Warnings) > 0 {
		result.ValidationLevel = ValidationLevelWarning
	}

	return result, nil
}

// GetReversalHistory gets the reversal history for a transaction
func (e *transactionReversalEngine) GetReversalHistory(ctx context.Context, transactionID uuid.UUID) (*ReversalHistoryResult, error) {
	ctx, span := e.tracing.StartSpan(ctx, "TransactionReversalEngine.GetReversalHistory")
	defer span.End()

	result := &ReversalHistoryResult{
		TransactionID: transactionID,
		ReversalChain: []ReversalRecord{},
		HasReversals:  false,
		IsReversed:    false,
		NetEffect:     []AccountBalanceUpdate{},
	}

	// Load the transaction
	transaction, err := e.transactionRepository.GetByID(ctx, transactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to load transaction: %w", err)
	}

	result.IsReversed = transaction.IsReversed

	// TODO: In a full implementation, this would query a reversal audit table
	// For now, just check if the transaction has a reversal reference
	if transaction.ReversedByTransactionID != nil {
		result.HasReversals = true

		reversalRecord := ReversalRecord{
			ReversalID:            uuid.New(), // Would be from audit table
			ReversalTransactionID: *transaction.ReversedByTransactionID,
			ReversalDate:          time.Now(), // Would be from audit table
			ReversedBy:            uuid.New(), // Would be from audit table
			ReversalReason:        *transaction.ReversalReason,
			ReversalType:          ReversalTypeFull,
			Status:                "COMPLETED",
		}

		result.ReversalChain = append(result.ReversalChain, reversalRecord)
	}

	return result, nil
}

// CreateCorrectionEntry creates a correction entry for partial corrections
func (e *transactionReversalEngine) CreateCorrectionEntry(ctx context.Context, req CorrectionEntryRequest) (*CorrectionEntryResult, error) {
	ctx, span := e.tracing.StartSpan(ctx, "TransactionReversalEngine.CreateCorrectionEntry")
	defer span.End()

	result := &CorrectionEntryResult{
		Success: false,
		Errors:  []string{},
	}

	// Load original transaction
	originalTransaction, err := e.transactionRepository.GetByID(ctx, req.OriginalTransactionID)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to load original transaction: %v", err))
		return result, err
	}

	// Create correction transaction
	correctionTransaction := &domain.Transaction{
		ID:                uuid.New(),
		TenantID:          originalTransaction.TenantID,
		EntityID:          originalTransaction.EntityID,
		TransactionType:   domain.TransactionTypeAdjustment,
		TransactionStatus: domain.TransactionStatusDraft,
		TransactionDate:   time.Now(),
		Description:       fmt.Sprintf("Correction for %s - %s", originalTransaction.TransactionNumber, req.CorrectionReason),
		CurrencyCode:      originalTransaction.CurrencyCode,
		ExchangeRate:      originalTransaction.ExchangeRate,
		ApprovalRequired:  true,
		ApprovalStatus:    domain.ApprovalStatusPending,
		ValidationStatus:  domain.ValidationStatusPending,
		CreatedBy:         req.CorrectedBy,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// Generate transaction number
	numberReq := GenerateNumberRequest{
		TenantID:        correctionTransaction.TenantID,
		EntityID:        correctionTransaction.EntityID,
		TransactionType: correctionTransaction.TransactionType,
		TransactionDate: correctionTransaction.TransactionDate,
	}

	transactionNumber, err := e.numberingService.GenerateTransactionNumber(ctx, numberReq)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to generate transaction number: %v", err))
		return result, err
	}

	correctionTransaction.TransactionNumber = transactionNumber

	// Create correction entries
	entries := []domain.TransactionEntry{}
	for i, correctionDef := range req.CorrectionEntries {
		entry := domain.TransactionEntry{
			ID:            uuid.New(),
			TenantID:      correctionTransaction.TenantID,
			TransactionID: correctionTransaction.ID,
			EntryNumber:   int32(i + 1),
			AccountID:     correctionDef.AccountID,
			DebitAmount:   correctionDef.NewDebitAmount,
			CreditAmount:  correctionDef.NewCreditAmount,
			Description:   correctionDef.Description,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		entries = append(entries, entry)
	}

	correctionTransaction.Entries = entries
	correctionTransaction.CalculateTotals()

	// Validate the correction transaction
	accounts, err := e.loadTransactionAccounts(ctx, correctionTransaction)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to load accounts: %v", err))
		return result, err
	}

	validationResult := e.validator.ValidateTransaction(ctx, correctionTransaction, accounts)
	if !validationResult.IsValid {
		result.Errors = append(result.Errors, "Correction transaction validation failed")
		for _, validationErr := range validationResult.Errors {
			result.Errors = append(result.Errors, validationErr.Message)
		}
		return result, fmt.Errorf("correction transaction validation failed")
	}

	// Save the correction transaction
	err = e.transactionRepository.Create(ctx, correctionTransaction)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to save correction transaction: %v", err))
		return result, err
	}

	result.CorrectionTransaction = correctionTransaction

	// Auto-post if requested
	if req.AutoPost {
		postReq := PostTransactionRequest{
			TransactionID:         correctionTransaction.ID,
			PostedBy:              req.CorrectedBy,
			PostingDate:           req.CorrectionDate,
			ValidateBeforePosting: true,
		}

		postResult, err := e.postingEngine.PostTransaction(ctx, postReq)
		if err == nil && postResult.Success {
			result.PostingResult = postResult
		}
	}

	result.Success = true
	return result, nil
}

// BatchReverseTransactions reverses multiple transactions in a batch
func (e *transactionReversalEngine) BatchReverseTransactions(ctx context.Context, req BatchReversalRequest) (*BatchReversalResult, error) {
	ctx, span := e.tracing.StartSpan(ctx, "TransactionReversalEngine.BatchReverseTransactions")
	defer span.End()

	startTime := time.Now()

	result := &BatchReversalResult{
		TotalTransactions:   len(req.TransactionIDs),
		SuccessfulReversals: 0,
		FailedReversals:     0,
		Results:             []ReversalTransactionResult{},
	}

	// Process each transaction
	for _, transactionID := range req.TransactionIDs {
		reversalReq := CreateReversalRequest{
			OriginalTransactionID: transactionID,
			ReversedBy:            req.ReversedBy,
			ReversalReason:        req.ReversalReason,
			ReversalDate:          req.ReversalDate,
			ReversalType:          ReversalTypeFull,
			AutoPost:              req.AutoPost,
			CopyOriginalMetadata:  true,
		}

		reversalResult, err := e.CreateReversalTransaction(ctx, reversalReq)
		if err != nil && !req.ContinueOnError {
			// Stop processing on first error if not continuing
			break
		}

		result.Results = append(result.Results, *reversalResult)

		if reversalResult.Success {
			result.SuccessfulReversals++
		} else {
			result.FailedReversals++
		}
	}

	result.ProcessingTime = time.Since(startTime)

	span.SetAttributes(
		attribute.Int64("total_transactions", int64(result.TotalTransactions)),
		attribute.Int64("successful_reversals", int64(result.SuccessfulReversals)),
		attribute.Int64("failed_reversals", int64(result.FailedReversals)),
		attribute.String("processing_time", result.ProcessingTime.String()),
	)

	return result, nil
}

// Helper methods

func (e *transactionReversalEngine) createReversalFromOriginal(ctx context.Context, original *domain.Transaction, req CreateReversalRequest) (*domain.Transaction, error) {
	reversal := &domain.Transaction{
		ID:                uuid.New(),
		TenantID:          original.TenantID,
		EntityID:          original.EntityID,
		TransactionType:   domain.TransactionTypeAdjustment,
		TransactionStatus: domain.TransactionStatusDraft,
		TransactionDate:   time.Now(),
		Description:       fmt.Sprintf("Reversal of %s - %s", original.TransactionNumber, req.ReversalReason),
		CurrencyCode:      original.CurrencyCode,
		ExchangeRate:      original.ExchangeRate,
		ApprovalRequired:  true,
		ApprovalStatus:    domain.ApprovalStatusPending,
		ValidationStatus:  domain.ValidationStatusPending,
		CreatedBy:         req.ReversedBy,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// Set reversal date
	if req.ReversalDate != nil {
		reversal.TransactionDate = *req.ReversalDate
	}

	// Generate transaction number
	if req.CustomReversalNumber != nil {
		reversal.TransactionNumber = *req.CustomReversalNumber
	} else {
		numberReq := GenerateNumberRequest{
			TenantID:        reversal.TenantID,
			EntityID:        reversal.EntityID,
			TransactionType: reversal.TransactionType,
			TransactionDate: reversal.TransactionDate,
		}

		transactionNumber, err := e.numberingService.GenerateTransactionNumber(ctx, numberReq)
		if err != nil {
			return nil, fmt.Errorf("failed to generate transaction number: %w", err)
		}
		reversal.TransactionNumber = transactionNumber
	}

	// Copy metadata if requested
	if req.CopyOriginalMetadata {
		reversal.SourceModule = original.SourceModule
		reversal.TransactionAttributes = original.TransactionAttributes
		reversal.Tags = original.Tags
	}

	// Create reversed entries (swap debit/credit amounts)
	entries := []domain.TransactionEntry{}
	for i, originalEntry := range original.Entries {
		reversalEntry := domain.TransactionEntry{
			ID:            uuid.New(),
			TenantID:      reversal.TenantID,
			TransactionID: reversal.ID,
			EntryNumber:   int32(i + 1),
			AccountID:     originalEntry.AccountID,
			DebitAmount:   originalEntry.CreditAmount, // Swap amounts
			CreditAmount:  originalEntry.DebitAmount,  // Swap amounts
			Description:   fmt.Sprintf("Reversal: %s", originalEntry.Description),
			Reference:     originalEntry.Reference,
			CostCenter:    originalEntry.CostCenter,
			Department:    originalEntry.Department,
			ProjectID:     originalEntry.ProjectID,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		entries = append(entries, reversalEntry)
	}

	reversal.Entries = entries
	reversal.CalculateTotals()

	return reversal, nil
}

func (e *transactionReversalEngine) loadTransactionAccounts(ctx context.Context, transaction *domain.Transaction) (map[uuid.UUID]*domain.Accounts, error) {
	// This is a simplified version - in production, you'd inject the account repository
	// For now, return empty map since we don't have direct access to account repository
	return make(map[uuid.UUID]*domain.Accounts), nil
}
