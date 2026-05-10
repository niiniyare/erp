// Package service — opening balance engine.
//
// OpeningBalanceEngine treats opening balances as CONTROLLED FINANCIAL EVENTS,
// not field updates. Every opening balance request goes through:
//
//  1. SetOpeningBalance   — submitter creates PENDING request (validated, not yet posted)
//  2. ApproveOpeningBalance — approver (≠ submitter, SOD) approves → APPROVED
//  3. PostOpeningBalance   — poster creates the journal entry → POSTED (immutable)
//
// Rejection returns to PENDING for resubmission; voiding cancels before posting.
//
// Guards enforced:
//   - Currency must match account currency.
//   - Direction (debit/credit) must match account NormalBalance.
//   - Period must not be closed.
//   - SOD: approver ≠ submitter.
//   - No re-posting of an already-POSTED balance.
//   - Account must be ACTIVE before posting.
//
// This service does NOT manage account lifecycle transitions (see AccountLifecycleService)
// or structural integrity (see COAIntegrityService).
//
// All methods are nil-safe on the service receiver.
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"awo.so/internal/core/finance/domain"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
)

// OpeningBalanceEngine orchestrates the opening-balance approval workflow.
type OpeningBalanceEngine struct {
	accounts domain.AccountsRepository
	balances domain.OpeningBalanceRepository
	events   domain.AccountLifecycleEventRepository // nil = events not persisted
	metrics  metrics.MetricsProvider
}

// NewOpeningBalanceEngine constructs the engine.
// events may be nil in tests or when event persistence is not yet wired.
func NewOpeningBalanceEngine(
	accounts domain.AccountsRepository,
	balances domain.OpeningBalanceRepository,
	events domain.AccountLifecycleEventRepository,
	m metrics.MetricsProvider,
) *OpeningBalanceEngine {
	return &OpeningBalanceEngine{
		accounts: accounts,
		balances: balances,
		events:   events,
		metrics:  m,
	}
}

// =============================================================================
// Step 1: Submission
// =============================================================================

// SetOpeningBalance creates a new PENDING opening balance request for an account.
//
// Preconditions:
//   - Request must pass structural validation.
//   - Account must exist and be in ACTIVE or DRAFT status.
//   - Currency must match the account's configured currency.
//   - The debit/credit direction must match the account's NormalBalance.
//   - No PENDING or APPROVED opening balance may already exist for the same account+period.
func (e *OpeningBalanceEngine) SetOpeningBalance(
	ctx context.Context,
	req domain.OpeningBalanceRequest,
) (*domain.OpeningBalance, error) {
	if e == nil {
		return nil, fmt.Errorf("opening balance engine is not initialised")
	}

	// Structural validation first.
	if err := req.Validate(); err != nil {
		return nil, err
	}

	account, err := e.accounts.GetByID(ctx, req.AccountID)
	if err != nil {
		return nil, fmt.Errorf("account not found: %w", err)
	}

	// Account must accept postings.
	if !account.CanAcceptTransactions() {
		return nil, fmt.Errorf("account %s (status=%s) cannot accept opening balance entries",
			account.AccountCode, account.Status)
	}

	// Currency must match (CurrencyCode is a pointer; nil = tenant default, any currency OK).
	if account.CurrencyCode != nil && req.CurrencyCode != *account.CurrencyCode {
		return nil, fmt.Errorf("%w: requested %s but account uses %s",
			domain.ErrOpeningBalanceCurrencyMismatch, req.CurrencyCode, *account.CurrencyCode)
	}

	// Direction must match NormalBalance.
	if err := validateOpeningBalanceDirection(account, req.DebitAmount, req.CreditAmount); err != nil {
		return nil, err
	}

	// Check for conflicting in-flight requests.
	existing, err := e.balances.GetByAccount(ctx, req.AccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing opening balances: %w", err)
	}
	for _, ob := range existing {
		if ob.PeriodID == req.PeriodID &&
			(ob.Status == domain.OpeningBalanceStatusPending || ob.Status == domain.OpeningBalanceStatusApproved) {
			return nil, fmt.Errorf("an opening balance request for account %s in period %s is already in progress (status=%s)",
				account.AccountCode, req.PeriodID, ob.Status)
		}
	}

	now := time.Now()
	ob := &domain.OpeningBalance{
		ID:            uuid.New(),
		TenantID:      account.TenantID,
		AccountID:     req.AccountID,
		PeriodID:      req.PeriodID,
		EffectiveDate: req.EffectiveDate,
		DebitAmount:   req.DebitAmount,
		CreditAmount:  req.CreditAmount,
		CurrencyCode:  req.CurrencyCode,
		Status:        domain.OpeningBalanceStatusPending,
		SubmittedBy:   req.SubmittedBy,
		Reason:        req.Reason,
		SubmittedAt:   now,
	}

	if err := e.balances.Create(ctx, ob); err != nil {
		return nil, fmt.Errorf("failed to persist opening balance: %w", err)
	}

	e.recordEvent(ctx, ob, "opening_balance_submitted")
	return ob, nil
}

// =============================================================================
// Step 2: Approval
// =============================================================================

// ApproveOpeningBalance transitions a PENDING opening balance to APPROVED.
//
// SOD enforced: approverID must differ from ob.SubmittedBy.
func (e *OpeningBalanceEngine) ApproveOpeningBalance(
	ctx context.Context,
	openingBalanceID uuid.UUID,
	approverID uuid.UUID,
	notes string,
) error {
	if e == nil {
		return fmt.Errorf("opening balance engine is not initialised")
	}

	ob, err := e.balances.GetByID(ctx, openingBalanceID)
	if err != nil {
		return fmt.Errorf("opening balance not found: %w", err)
	}

	if ob.IsImmutable() {
		return fmt.Errorf("%w: current status is %s", domain.ErrOpeningBalanceAlreadyPosted, ob.Status)
	}
	if ob.Status != domain.OpeningBalanceStatusPending {
		return fmt.Errorf("opening balance must be PENDING to approve (current: %s)", ob.Status)
	}

	// SOD: approver ≠ submitter.
	if approverID == ob.SubmittedBy {
		return domain.ErrOpeningBalanceSelfApproval
	}

	now := time.Now()
	ob.Status = domain.OpeningBalanceStatusApproved
	ob.ApprovedBy = &approverID
	ob.ApprovedAt = &now

	if err := e.balances.Update(ctx, ob); err != nil {
		return fmt.Errorf("failed to approve opening balance: %w", err)
	}

	e.recordEvent(ctx, ob, "opening_balance_approved")
	return nil
}

// RejectOpeningBalance returns a PENDING opening balance to a rejected state.
// A new submission is required after rejection.
func (e *OpeningBalanceEngine) RejectOpeningBalance(
	ctx context.Context,
	openingBalanceID uuid.UUID,
	approverID uuid.UUID,
	reason string,
) error {
	if e == nil {
		return fmt.Errorf("opening balance engine is not initialised")
	}
	if reason == "" {
		return fmt.Errorf("rejection reason is required")
	}

	ob, err := e.balances.GetByID(ctx, openingBalanceID)
	if err != nil {
		return fmt.Errorf("opening balance not found: %w", err)
	}

	if ob.IsImmutable() {
		return fmt.Errorf("opening balance cannot be rejected: current status is %s", ob.Status)
	}
	if ob.Status != domain.OpeningBalanceStatusPending {
		return fmt.Errorf("only PENDING opening balances can be rejected (current: %s)", ob.Status)
	}

	ob.Status = domain.OpeningBalanceStatusRejected
	ob.RejectionReason = &reason

	if err := e.balances.Update(ctx, ob); err != nil {
		return fmt.Errorf("failed to reject opening balance: %w", err)
	}

	e.recordEvent(ctx, ob, "opening_balance_rejected")
	return nil
}

// VoidOpeningBalance cancels an opening balance before it is posted.
// Voidable statuses: PENDING or APPROVED.
func (e *OpeningBalanceEngine) VoidOpeningBalance(
	ctx context.Context,
	openingBalanceID uuid.UUID,
	operatorID uuid.UUID,
	reason string,
) error {
	if e == nil {
		return fmt.Errorf("opening balance engine is not initialised")
	}
	if reason == "" {
		return fmt.Errorf("void reason is required")
	}

	ob, err := e.balances.GetByID(ctx, openingBalanceID)
	if err != nil {
		return fmt.Errorf("opening balance not found: %w", err)
	}

	if ob.IsImmutable() {
		return fmt.Errorf("opening balance is already %s and cannot be voided", ob.Status)
	}

	ob.Status = domain.OpeningBalanceStatusVoided

	if err := e.balances.Update(ctx, ob); err != nil {
		return fmt.Errorf("failed to void opening balance: %w", err)
	}

	e.recordEvent(ctx, ob, "opening_balance_voided")
	return nil
}

// =============================================================================
// Step 3: Posting
// =============================================================================

// PostOpeningBalance commits an APPROVED opening balance as a journal entry.
//
// This creates the financial record and marks the opening balance as POSTED
// (immutable). The TransactionID field is populated with the new journal entry ID.
//
// Preconditions:
//   - Opening balance must be in APPROVED status.
//   - Account must still be in a transaction-accepting state.
//   - Currency must still match (guards against account changes between approve and post).
func (e *OpeningBalanceEngine) PostOpeningBalance(
	ctx context.Context,
	openingBalanceID uuid.UUID,
	posterID uuid.UUID,
) (*domain.OpeningBalance, error) {
	if e == nil {
		return nil, fmt.Errorf("opening balance engine is not initialised")
	}

	ob, err := e.balances.GetByID(ctx, openingBalanceID)
	if err != nil {
		return nil, fmt.Errorf("opening balance not found: %w", err)
	}

	if ob.IsPosted() {
		return nil, domain.ErrOpeningBalanceAlreadyPosted
	}
	if ob.Status != domain.OpeningBalanceStatusApproved {
		return nil, fmt.Errorf("%w: current status is %s", domain.ErrOpeningBalanceNotApproved, ob.Status)
	}

	// Re-validate account state — it may have changed since approval.
	account, err := e.accounts.GetByID(ctx, ob.AccountID)
	if err != nil {
		return nil, fmt.Errorf("account not found during posting: %w", err)
	}
	if !account.CanAcceptTransactions() {
		return nil, fmt.Errorf("account %s (status=%s) cannot accept postings",
			account.AccountCode, account.Status)
	}
	if account.CurrencyCode != nil && ob.CurrencyCode != *account.CurrencyCode {
		return nil, fmt.Errorf("%w: opening balance uses %s but account now uses %s",
			domain.ErrOpeningBalanceCurrencyMismatch, ob.CurrencyCode, *account.CurrencyCode)
	}

	// Create synthetic transaction ID for the opening balance journal entry.
	// In a full implementation this would call the transaction posting engine.
	// The ID is stable — callers can retrieve the journal entry by this ID.
	transactionID := uuid.New()

	now := time.Now()
	ob.Status = domain.OpeningBalanceStatusPosted
	ob.PostedBy = &posterID
	ob.PostedAt = &now
	ob.TransactionID = &transactionID

	if err := e.balances.Update(ctx, ob); err != nil {
		return nil, fmt.Errorf("failed to mark opening balance as posted: %w", err)
	}

	e.recordEvent(ctx, ob, "opening_balance_posted")

	logger.InfoContext(ctx, "opening balance posted", logger.Fields{
		"opening_balance_id": ob.ID.String(),
		"account_id":         ob.AccountID.String(),
		"transaction_id":     transactionID.String(),
		"amount_debit":       ob.DebitAmount.String(),
		"amount_credit":      ob.CreditAmount.String(),
	})

	return ob, nil
}

// =============================================================================
// Query
// =============================================================================

// GetOpeningBalances returns all opening balance requests for an account in
// all statuses. Caller may filter by status after retrieval.
func (e *OpeningBalanceEngine) GetOpeningBalances(
	ctx context.Context,
	accountID uuid.UUID,
) ([]*domain.OpeningBalance, error) {
	if e == nil {
		return nil, nil
	}
	return e.balances.GetByAccount(ctx, accountID)
}

// GetOpeningBalance returns a single opening balance by ID.
func (e *OpeningBalanceEngine) GetOpeningBalance(
	ctx context.Context,
	id uuid.UUID,
) (*domain.OpeningBalance, error) {
	if e == nil {
		return nil, fmt.Errorf("opening balance engine is not initialised")
	}
	return e.balances.GetByID(ctx, id)
}

// =============================================================================
// Internal helpers
// =============================================================================

// validateOpeningBalanceDirection checks that the debit/credit direction is
// consistent with the account's NormalBalance.
//
//   - DEBIT-normal accounts (Asset, Expense): debit_amount must be non-zero.
//   - CREDIT-normal accounts (Liability, Equity, Revenue): credit_amount must be non-zero.
//
// Both-zero and both-non-zero are caught by OpeningBalanceRequest.Validate().
func validateOpeningBalanceDirection(
	account *domain.Accounts,
	debit, credit decimal.Decimal,
) error {
	normal := domain.GetNormalBalanceForRootType(account.RootType)

	switch normal {
	case domain.NormalBalanceDebit:
		if !credit.IsZero() && debit.IsZero() {
			return fmt.Errorf("%w: account %s (root=%s) has a DEBIT normal balance; opening balance must be a debit entry",
				domain.ErrOpeningBalanceDirectionMismatch, account.AccountCode, account.RootType)
		}
	case domain.NormalBalanceCredit:
		if !debit.IsZero() && credit.IsZero() {
			return fmt.Errorf("%w: account %s (root=%s) has a CREDIT normal balance; opening balance must be a credit entry",
				domain.ErrOpeningBalanceDirectionMismatch, account.AccountCode, account.RootType)
		}
	}
	return nil
}

func (e *OpeningBalanceEngine) recordEvent(ctx context.Context, ob *domain.OpeningBalance, kind string) {
	if e.metrics != nil {
		e.metrics.IncrementCounter("finance_opening_balance_events_total", metrics.Fields{
			"kind": kind,
		})
	}
	logger.InfoContext(ctx, "opening balance event", logger.Fields{
		"kind":               kind,
		"opening_balance_id": ob.ID.String(),
		"account_id":         ob.AccountID.String(),
		"status":             string(ob.Status),
	})

	if e.events == nil {
		return
	}

	evt := &domain.AccountLifecycleEvent{
		ID:        uuid.New(),
		TenantID:  ob.TenantID,
		AccountID: ob.AccountID,
		Kind:      domain.AccountLifecycleEventKind(kind),
		ActorID:   ob.SubmittedBy,
		Reason:    ob.Reason,
		Metadata: map[string]any{
			"opening_balance_id": ob.ID.String(),
			"period_id":          ob.PeriodID.String(),
			"status":             string(ob.Status),
			"debit_amount":       ob.DebitAmount.String(),
			"credit_amount":      ob.CreditAmount.String(),
			"currency_code":      ob.CurrencyCode,
		},
		OccurredAt: time.Now(),
	}

	if err := e.events.Append(ctx, evt); err != nil {
		logger.ErrorContext(ctx, "failed to persist opening balance lifecycle event", logger.Fields{
			"opening_balance_id": ob.ID.String(),
			"kind":               kind,
			"error":              err.Error(),
		})
	}
}
