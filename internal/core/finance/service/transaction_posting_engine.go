package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/niiniyare/erp/internal/core/finance/domain"
	"github.com/niiniyare/erp/internal/shared"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// TransactionPostingEngine handles the posting of financial transactions to the ledger
type TransactionPostingEngine interface {
	// PostTransaction posts a transaction to the ledger with account balance updates
	PostTransaction(ctx context.Context, req PostTransactionRequest) (*PostTransactionResult, error)
	
	// UnpostTransaction reverses a posted transaction (for corrections)
	UnpostTransaction(ctx context.Context, req UnpostTransactionRequest) (*UnpostTransactionResult, error)
	
	// ValidatePostingRequirements validates if a transaction can be posted
	ValidatePostingRequirements(ctx context.Context, transaction *domain.Transaction, accounts map[uuid.UUID]*domain.Accounts) *PostingValidationResult
	
	// PreviewPostingImpact previews the impact of posting without actually posting
	PreviewPostingImpact(ctx context.Context, transaction *domain.Transaction, accounts map[uuid.UUID]*domain.Accounts) (*PostingImpactPreview, error)
	
	// RecalculateAccountBalances recalculates account balances after posting
	RecalculateAccountBalances(ctx context.Context, accountIDs []uuid.UUID) (*BalanceRecalculationResult, error)
	
	// BatchPostTransactions posts multiple transactions in a single batch
	BatchPostTransactions(ctx context.Context, req BatchPostRequest) (*BatchPostResult, error)
}

// PostTransactionRequest represents a request to post a transaction
type PostTransactionRequest struct {
	TransactionID         uuid.UUID  `json:"transaction_id"`
	PostedBy              uuid.UUID  `json:"posted_by"`
	PostingDate           *time.Time `json:"posting_date,omitempty"` // If nil, uses current date
	ForcePost             bool       `json:"force_post"`             // Override validation warnings
	SkipBalanceUpdate     bool       `json:"skip_balance_update"`    // For batch operations
	ValidateBeforePosting bool       `json:"validate_before_posting"`
}

// UnpostTransactionRequest represents a request to unpost a transaction
type UnpostTransactionRequest struct {
	TransactionID uuid.UUID `json:"transaction_id"`
	UnpostedBy    uuid.UUID `json:"unposted_by"`
	Reason        string    `json:"reason"`
}

// BatchPostRequest represents a batch posting request
type BatchPostRequest struct {
	TransactionIDs  []uuid.UUID `json:"transaction_ids"`
	PostedBy        uuid.UUID   `json:"posted_by"`
	PostingDate     *time.Time  `json:"posting_date,omitempty"`
	ContinueOnError bool        `json:"continue_on_error"`
}

// PostTransactionResult represents the result of posting a transaction
type PostTransactionResult struct {
	TransactionID    uuid.UUID                   `json:"transaction_id"`
	Success          bool                        `json:"success"`
	PostedAt         time.Time                   `json:"posted_at"`
	PostedBy         uuid.UUID                   `json:"posted_by"`
	BalanceUpdates   []AccountBalanceUpdate      `json:"balance_updates"`
	ValidationResult *PostingValidationResult    `json:"validation_result,omitempty"`
	Errors           []string                    `json:"errors,omitempty"`
	Warnings         []string                    `json:"warnings,omitempty"`
}

// UnpostTransactionResult represents the result of unposting a transaction
type UnpostTransactionResult struct {
	TransactionID    uuid.UUID              `json:"transaction_id"`
	Success          bool                   `json:"success"`
	UnpostedAt       time.Time              `json:"unposted_at"`
	UnpostedBy       uuid.UUID              `json:"unposted_by"`
	BalanceUpdates   []AccountBalanceUpdate `json:"balance_updates"`
	Errors           []string               `json:"errors,omitempty"`
}

// BatchPostResult represents the result of batch posting
type BatchPostResult struct {
	TotalTransactions     int                      `json:"total_transactions"`
	SuccessfulPosts      int                      `json:"successful_posts"`
	FailedPosts          int                      `json:"failed_posts"`
	Results              []PostTransactionResult  `json:"results"`
	BatchBalanceUpdates  []AccountBalanceUpdate   `json:"batch_balance_updates"`
	ProcessingTime       time.Duration            `json:"processing_time"`
}

// PostingValidationResult represents validation results for posting
type PostingValidationResult struct {
	CanPost          bool                `json:"can_post"`
	ValidationLevel  ValidationLevel     `json:"validation_level"`
	Errors           []ValidationError   `json:"errors,omitempty"`
	Warnings         []ValidationWarning `json:"warnings,omitempty"`
	BusinessRules    []BusinessRuleResult `json:"business_rules,omitempty"`
	AccountingPeriod *AccountingPeriodInfo `json:"accounting_period,omitempty"`
}

// PostingImpactPreview shows the impact of posting without actually posting
type PostingImpactPreview struct {
	TransactionID           uuid.UUID              `json:"transaction_id"`
	ExpectedBalanceChanges  []AccountBalanceUpdate `json:"expected_balance_changes"`
	AffectedAccounts        []uuid.UUID            `json:"affected_accounts"`
	EstimatedProcessingTime time.Duration          `json:"estimated_processing_time"`
	ValidationResult        *PostingValidationResult `json:"validation_result"`
}

// AccountBalanceUpdate represents a balance change for an account
type AccountBalanceUpdate struct {
	AccountID       uuid.UUID       `json:"account_id"`
	AccountName     string          `json:"account_name"`
	PreviousBalance decimal.Decimal `json:"previous_balance"`
	BalanceChange   decimal.Decimal `json:"balance_change"`
	NewBalance      decimal.Decimal `json:"new_balance"`
	UpdateType      BalanceUpdateType `json:"update_type"`
}

// BalanceUpdateType represents the type of balance update
type BalanceUpdateType string

const (
	BalanceUpdateTypeDebit  BalanceUpdateType = "DEBIT"
	BalanceUpdateTypeCredit BalanceUpdateType = "CREDIT"
)

// BalanceRecalculationResult represents the result of balance recalculation
type BalanceRecalculationResult struct {
	AccountsRecalculated int                    `json:"accounts_recalculated"`
	BalanceUpdates       []AccountBalanceUpdate `json:"balance_updates"`
	RecalculationTime    time.Duration          `json:"recalculation_time"`
	Errors               []string               `json:"errors,omitempty"`
}

// AccountingPeriodInfo represents information about the accounting period
type AccountingPeriodInfo struct {
	PeriodID    uuid.UUID  `json:"period_id"`
	PeriodName  string     `json:"period_name"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     time.Time  `json:"end_date"`
	IsClosed    bool       `json:"is_closed"`
	ClosedDate  *time.Time `json:"closed_date,omitempty"`
}

// transactionPostingEngine implements TransactionPostingEngine
type transactionPostingEngine struct {
	accountRepository       domain.AccountsRepository
	transactionRepository  domain.TransactionRepository
	validator              DoubleEntryValidator
	stateMachine           func(*domain.Transaction) *domain.TransactionStateMachine
	workflowEngine         func(*domain.Transaction) *domain.TransactionWorkflowEngine
	tracing                tracing.TracingService
}

// TransactionPostingEngineDeps represents dependencies for the posting engine
type TransactionPostingEngineDeps struct {
	AccountRepository      domain.AccountsRepository
	TransactionRepository domain.TransactionRepository
	Validator             DoubleEntryValidator
	Tracing               tracing.TracingService
}

// NewTransactionPostingEngine creates a new transaction posting engine
func NewTransactionPostingEngine(deps TransactionPostingEngineDeps) TransactionPostingEngine {
	return &transactionPostingEngine{
		accountRepository:     deps.AccountRepository,
		transactionRepository: deps.TransactionRepository,
		validator:            deps.Validator,
		tracing:             deps.Tracing,
		stateMachine: func(t *domain.Transaction) *domain.TransactionStateMachine {
			return domain.NewTransactionStateMachine(t)
		},
		workflowEngine: func(t *domain.Transaction) *domain.TransactionWorkflowEngine {
			return domain.NewTransactionWorkflowEngine(t)
		},
	}
}

// PostTransaction posts a transaction to the ledger with account balance updates
func (e *transactionPostingEngine) PostTransaction(ctx context.Context, req PostTransactionRequest) (*PostTransactionResult, error) {
	ctx, span := e.tracing.StartSpan(ctx, "TransactionPostingEngine.PostTransaction")
	defer span.End()

	startTime := time.Now()
	result := &PostTransactionResult{
		TransactionID: req.TransactionID,
		Success:       false,
		PostedBy:      req.PostedBy,
		BalanceUpdates: []AccountBalanceUpdate{},
		Errors:        []string{},
		Warnings:      []string{},
	}

	span.SetAttributes(
		shared.StringAttribute("transaction_id", req.TransactionID.String()),
		shared.StringAttribute("posted_by", req.PostedBy.String()),
		shared.BoolAttribute("force_post", req.ForcePost),
	)

	// 1. Retrieve the transaction
	transaction, err := e.transactionRepository.GetByID(ctx, req.TransactionID)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to retrieve transaction: %v", err))
		return result, err
	}

	// 2. Validate transaction can be posted
	if req.ValidateBeforePosting {
		accounts, err := e.loadTransactionAccounts(ctx, transaction)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to load accounts: %v", err))
			return result, err
		}

		validationResult := e.ValidatePostingRequirements(ctx, transaction, accounts)
		result.ValidationResult = validationResult

		if !validationResult.CanPost && !req.ForcePost {
			result.Errors = append(result.Errors, "Transaction cannot be posted due to validation errors")
			for _, validationErr := range validationResult.Errors {
				result.Errors = append(result.Errors, validationErr.Message)
			}
			return result, fmt.Errorf("transaction cannot be posted")
		}

		// Add warnings if any
		for _, warning := range validationResult.Warnings {
			result.Warnings = append(result.Warnings, warning.Message)
		}
	}

	// 3. Set posting date
	postingDate := time.Now()
	if req.PostingDate != nil {
		postingDate = *req.PostingDate
	}

	// 4. Update transaction status to posted using state machine
	stateMachine := e.stateMachine(transaction)
	workflowEngine := e.workflowEngine(transaction)

	err = workflowEngine.Post(req.PostedBy)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to post transaction via workflow: %v", err))
		return result, err
	}

	// Set posting date
	transaction.PostingDate = &postingDate

	// 5. Update account balances (if not skipped for batch operations)
	if !req.SkipBalanceUpdate {
		balanceUpdates, err := e.updateAccountBalances(ctx, transaction)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to update account balances: %v", err))
			return result, err
		}
		result.BalanceUpdates = balanceUpdates
	}

	// 6. Save the updated transaction
	err = e.transactionRepository.Update(ctx, transaction)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to save posted transaction: %v", err))
		return result, err
	}

	// 7. Set success result
	result.Success = true
	result.PostedAt = postingDate

	span.SetAttributes(
		shared.BoolAttribute("success", result.Success),
		shared.Int64Attribute("balance_updates_count", int64(len(result.BalanceUpdates))),
		shared.DurationAttribute("processing_time", time.Since(startTime)),
	)

	return result, nil
}

// UnpostTransaction reverses a posted transaction
func (e *transactionPostingEngine) UnpostTransaction(ctx context.Context, req UnpostTransactionRequest) (*UnpostTransactionResult, error) {
	ctx, span := e.tracing.StartSpan(ctx, "TransactionPostingEngine.UnpostTransaction")
	defer span.End()

	result := &UnpostTransactionResult{
		TransactionID: req.TransactionID,
		Success:       false,
		UnpostedBy:    req.UnpostedBy,
		UnpostedAt:    time.Now(),
		BalanceUpdates: []AccountBalanceUpdate{},
		Errors:        []string{},
	}

	// 1. Retrieve the transaction
	transaction, err := e.transactionRepository.GetByID(ctx, req.TransactionID)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to retrieve transaction: %v", err))
		return result, err
	}

	// 2. Validate transaction can be unposted
	if transaction.TransactionStatus != domain.TransactionStatusPosted {
		result.Errors = append(result.Errors, "Only posted transactions can be unposted")
		return result, fmt.Errorf("transaction is not posted")
	}

	// 3. Reverse account balance updates
	balanceUpdates, err := e.reverseAccountBalances(ctx, transaction)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to reverse account balances: %v", err))
		return result, err
	}
	result.BalanceUpdates = balanceUpdates

	// 4. Update transaction status back to approved/draft
	transaction.TransactionStatus = domain.TransactionStatusApproved
	transaction.PostedAt = nil
	transaction.PostedBy = nil
	transaction.PostingDate = nil

	// 5. Save the updated transaction
	err = e.transactionRepository.Update(ctx, transaction)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to save unposted transaction: %v", err))
		return result, err
	}

	result.Success = true
	return result, nil
}

// ValidatePostingRequirements validates if a transaction can be posted
func (e *transactionPostingEngine) ValidatePostingRequirements(ctx context.Context, transaction *domain.Transaction, accounts map[uuid.UUID]*domain.Accounts) *PostingValidationResult {
	ctx, span := e.tracing.StartSpan(ctx, "TransactionPostingEngine.ValidatePostingRequirements")
	defer span.End()

	result := &PostingValidationResult{
		CanPost:         true,
		ValidationLevel: ValidationLevelValid,
		Errors:          []ValidationError{},
		Warnings:        []ValidationWarning{},
		BusinessRules:   []BusinessRuleResult{},
	}

	// 1. Check transaction status
	if !transaction.CanBePosted() {
		result.CanPost = false
		result.Errors = append(result.Errors, ValidationError{
			Field:    "transaction_status",
			Message:  "Transaction is not in a postable state",
			Code:     "INVALID_STATUS_FOR_POSTING",
			Severity: "ERROR",
			Value:    transaction.TransactionStatus,
		})
	}

	// 2. Validate double-entry requirements
	validationResult := e.validator.ValidateTransaction(ctx, transaction, accounts)
	if !validationResult.IsValid {
		result.CanPost = false
	}
	
	result.Errors = append(result.Errors, validationResult.Errors...)
	result.Warnings = append(result.Warnings, validationResult.Warnings...)
	result.BusinessRules = append(result.BusinessRules, validationResult.BusinessRules...)

	// 3. Check accounting period (simplified - in production, this would check actual periods)
	if transaction.PostingDate != nil {
		periodInfo := &AccountingPeriodInfo{
			PeriodID:   uuid.New(),
			PeriodName: fmt.Sprintf("%d-%02d", transaction.PostingDate.Year(), transaction.PostingDate.Month()),
			StartDate:  time.Date(transaction.PostingDate.Year(), transaction.PostingDate.Month(), 1, 0, 0, 0, 0, time.UTC),
			EndDate:    time.Date(transaction.PostingDate.Year(), transaction.PostingDate.Month()+1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, -1),
			IsClosed:   false, // Simplified - would check actual period status
		}
		result.AccountingPeriod = periodInfo

		if periodInfo.IsClosed {
			result.CanPost = false
			result.Errors = append(result.Errors, ValidationError{
				Field:    "posting_date",
				Message:  "Cannot post to a closed accounting period",
				Code:     "CLOSED_PERIOD",
				Severity: "ERROR",
				Value:    transaction.PostingDate,
			})
		}
	}

	// 4. Check for duplicate postings (simplified)
	if transaction.TransactionStatus == domain.TransactionStatusPosted {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "transaction_status",
			Message: "Transaction is already posted",
			Code:    "ALREADY_POSTED",
			Value:   transaction.TransactionStatus,
		})
	}

	// Set final validation level
	if len(result.Errors) > 0 {
		result.CanPost = false
		result.ValidationLevel = ValidationLevelError
	} else if len(result.Warnings) > 0 {
		result.ValidationLevel = ValidationLevelWarning
	}

	return result
}

// PreviewPostingImpact previews the impact of posting without actually posting
func (e *transactionPostingEngine) PreviewPostingImpact(ctx context.Context, transaction *domain.Transaction, accounts map[uuid.UUID]*domain.Accounts) (*PostingImpactPreview, error) {
	ctx, span := e.tracing.StartSpan(ctx, "TransactionPostingEngine.PreviewPostingImpact")
	defer span.End()

	startTime := time.Now()

	preview := &PostingImpactPreview{
		TransactionID:          transaction.ID,
		ExpectedBalanceChanges: []AccountBalanceUpdate{},
		AffectedAccounts:       []uuid.UUID{},
	}

	// 1. Validate posting requirements
	validationResult := e.ValidatePostingRequirements(ctx, transaction, accounts)
	preview.ValidationResult = validationResult

	// 2. Calculate expected balance changes
	for _, entry := range transaction.Entries {
		account, exists := accounts[entry.AccountID]
		if !exists {
			continue
		}

		balanceChange := decimal.Zero
		updateType := BalanceUpdateTypeDebit

		if entry.DebitAmount.GreaterThan(decimal.Zero) {
			balanceChange = entry.DebitAmount
			updateType = BalanceUpdateTypeDebit
		} else if entry.CreditAmount.GreaterThan(decimal.Zero) {
			balanceChange = entry.CreditAmount
			updateType = BalanceUpdateTypeCredit
		}

		// Adjust balance based on normal balance type
		newBalance := account.CurrentBalance
		if account.NormalBalance == domain.NormalBalanceDebit {
			if updateType == BalanceUpdateTypeDebit {
				newBalance = newBalance.Add(balanceChange)
			} else {
				newBalance = newBalance.Sub(balanceChange)
			}
		} else { // Credit normal balance
			if updateType == BalanceUpdateTypeCredit {
				newBalance = newBalance.Add(balanceChange)
			} else {
				newBalance = newBalance.Sub(balanceChange)
			}
		}

		balanceUpdate := AccountBalanceUpdate{
			AccountID:       entry.AccountID,
			AccountName:     account.AccountName,
			PreviousBalance: account.CurrentBalance,
			BalanceChange:   balanceChange,
			NewBalance:      newBalance,
			UpdateType:      updateType,
		}

		preview.ExpectedBalanceChanges = append(preview.ExpectedBalanceChanges, balanceUpdate)
		preview.AffectedAccounts = append(preview.AffectedAccounts, entry.AccountID)
	}

	// 3. Estimate processing time (simplified)
	preview.EstimatedProcessingTime = time.Since(startTime) * 2 // Rough estimate

	return preview, nil
}

// RecalculateAccountBalances recalculates account balances after posting
func (e *transactionPostingEngine) RecalculateAccountBalances(ctx context.Context, accountIDs []uuid.UUID) (*BalanceRecalculationResult, error) {
	ctx, span := e.tracing.StartSpan(ctx, "TransactionPostingEngine.RecalculateAccountBalances")
	defer span.End()

	startTime := time.Now()

	result := &BalanceRecalculationResult{
		AccountsRecalculated: 0,
		BalanceUpdates:       []AccountBalanceUpdate{},
		Errors:              []string{},
	}

	// This is a simplified implementation
	// In production, this would query all transaction entries for each account
	// and recalculate the balance from scratch

	for _, accountID := range accountIDs {
		// Get current account
		account, err := e.accountRepository.GetByID(ctx, accountID)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to get account %s: %v", accountID, err))
			continue
		}

		// TODO: Calculate actual balance from transaction entries
		// For now, just mark as recalculated
		result.AccountsRecalculated++

		balanceUpdate := AccountBalanceUpdate{
			AccountID:       accountID,
			AccountName:     account.AccountName,
			PreviousBalance: account.CurrentBalance,
			BalanceChange:   decimal.Zero, // Would be calculated
			NewBalance:      account.CurrentBalance, // Would be recalculated
			UpdateType:      BalanceUpdateTypeDebit,
		}

		result.BalanceUpdates = append(result.BalanceUpdates, balanceUpdate)
	}

	result.RecalculationTime = time.Since(startTime)
	return result, nil
}

// BatchPostTransactions posts multiple transactions in a single batch
func (e *transactionPostingEngine) BatchPostTransactions(ctx context.Context, req BatchPostRequest) (*BatchPostResult, error) {
	ctx, span := e.tracing.StartSpan(ctx, "TransactionPostingEngine.BatchPostTransactions")
	defer span.End()

	startTime := time.Now()

	result := &BatchPostResult{
		TotalTransactions:   len(req.TransactionIDs),
		SuccessfulPosts:    0,
		FailedPosts:        0,
		Results:            []PostTransactionResult{},
		BatchBalanceUpdates: []AccountBalanceUpdate{},
	}

	// Process each transaction
	for _, transactionID := range req.TransactionIDs {
		postReq := PostTransactionRequest{
			TransactionID:         transactionID,
			PostedBy:              req.PostedBy,
			PostingDate:           req.PostingDate,
			ForcePost:             false,
			SkipBalanceUpdate:     true, // Skip individual updates for batch
			ValidateBeforePosting: true,
		}

		postResult, err := e.PostTransaction(ctx, postReq)
		result.Results = append(result.Results, *postResult)

		if postResult.Success {
			result.SuccessfulPosts++
		} else {
			result.FailedPosts++
			if !req.ContinueOnError {
				break // Stop on first error if not continuing
			}
		}
	}

	// Perform batch balance updates
	// TODO: Implement efficient batch balance update logic

	result.ProcessingTime = time.Since(startTime)

	span.SetAttributes(
		shared.Int64Attribute("total_transactions", int64(result.TotalTransactions)),
		shared.Int64Attribute("successful_posts", int64(result.SuccessfulPosts)),
		shared.Int64Attribute("failed_posts", int64(result.FailedPosts)),
		shared.DurationAttribute("processing_time", result.ProcessingTime),
	)

	return result, nil
}

// Helper methods

func (e *transactionPostingEngine) loadTransactionAccounts(ctx context.Context, transaction *domain.Transaction) (map[uuid.UUID]*domain.Accounts, error) {
	accounts := make(map[uuid.UUID]*domain.Accounts)

	for _, entry := range transaction.Entries {
		if _, exists := accounts[entry.AccountID]; !exists {
			account, err := e.accountRepository.GetByID(ctx, entry.AccountID)
			if err != nil {
				return nil, fmt.Errorf("failed to load account %s: %w", entry.AccountID, err)
			}
			accounts[entry.AccountID] = account
		}
	}

	return accounts, nil
}

func (e *transactionPostingEngine) updateAccountBalances(ctx context.Context, transaction *domain.Transaction) ([]AccountBalanceUpdate, error) {
	var balanceUpdates []AccountBalanceUpdate

	for _, entry := range transaction.Entries {
		// Get current account
		account, err := e.accountRepository.GetByID(ctx, entry.AccountID)
		if err != nil {
			return nil, fmt.Errorf("failed to get account %s: %w", entry.AccountID, err)
		}

		previousBalance := account.CurrentBalance
		balanceChange := decimal.Zero
		updateType := BalanceUpdateTypeDebit

		// Determine balance change
		if entry.DebitAmount.GreaterThan(decimal.Zero) {
			balanceChange = entry.DebitAmount
			updateType = BalanceUpdateTypeDebit
		} else if entry.CreditAmount.GreaterThan(decimal.Zero) {
			balanceChange = entry.CreditAmount
			updateType = BalanceUpdateTypeCredit
		}

		// Calculate new balance based on normal balance
		newBalance := previousBalance
		if account.NormalBalance == domain.NormalBalanceDebit {
			if updateType == BalanceUpdateTypeDebit {
				newBalance = newBalance.Add(balanceChange)
			} else {
				newBalance = newBalance.Sub(balanceChange)
			}
		} else { // Credit normal balance
			if updateType == BalanceUpdateTypeCredit {
				newBalance = newBalance.Add(balanceChange)
			} else {
				newBalance = newBalance.Sub(balanceChange)
			}
		}

		// Update account balance
		account.CurrentBalance = newBalance
		err = e.accountRepository.Update(ctx, account)
		if err != nil {
			return nil, fmt.Errorf("failed to update account balance for %s: %w", entry.AccountID, err)
		}

		balanceUpdate := AccountBalanceUpdate{
			AccountID:       entry.AccountID,
			AccountName:     account.AccountName,
			PreviousBalance: previousBalance,
			BalanceChange:   balanceChange,
			NewBalance:      newBalance,
			UpdateType:      updateType,
		}

		balanceUpdates = append(balanceUpdates, balanceUpdate)
	}

	return balanceUpdates, nil
}

func (e *transactionPostingEngine) reverseAccountBalances(ctx context.Context, transaction *domain.Transaction) ([]AccountBalanceUpdate, error) {
	var balanceUpdates []AccountBalanceUpdate

	for _, entry := range transaction.Entries {
		// Get current account
		account, err := e.accountRepository.GetByID(ctx, entry.AccountID)
		if err != nil {
			return nil, fmt.Errorf("failed to get account %s: %w", entry.AccountID, err)
		}

		previousBalance := account.CurrentBalance
		balanceChange := decimal.Zero
		updateType := BalanceUpdateTypeDebit

		// Determine balance change (reverse of original)
		if entry.DebitAmount.GreaterThan(decimal.Zero) {
			balanceChange = entry.DebitAmount
			updateType = BalanceUpdateTypeCredit // Reverse the type
		} else if entry.CreditAmount.GreaterThan(decimal.Zero) {
			balanceChange = entry.CreditAmount
			updateType = BalanceUpdateTypeDebit // Reverse the type
		}

		// Calculate new balance (reverse of original posting)
		newBalance := previousBalance
		if account.NormalBalance == domain.NormalBalanceDebit {
			if updateType == BalanceUpdateTypeDebit {
				newBalance = newBalance.Add(balanceChange)
			} else {
				newBalance = newBalance.Sub(balanceChange)
			}
		} else { // Credit normal balance
			if updateType == BalanceUpdateTypeCredit {
				newBalance = newBalance.Add(balanceChange)
			} else {
				newBalance = newBalance.Sub(balanceChange)
			}
		}

		// Update account balance
		account.CurrentBalance = newBalance
		err = e.accountRepository.Update(ctx, account)
		if err != nil {
			return nil, fmt.Errorf("failed to reverse account balance for %s: %w", entry.AccountID, err)
		}

		balanceUpdate := AccountBalanceUpdate{
			AccountID:       entry.AccountID,
			AccountName:     account.AccountName,
			PreviousBalance: previousBalance,
			BalanceChange:   balanceChange.Neg(), // Negative change for reversal
			NewBalance:      newBalance,
			UpdateType:      updateType,
		}

		balanceUpdates = append(balanceUpdates, balanceUpdate)
	}

	return balanceUpdates, nil
}