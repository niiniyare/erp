package pipeline

import (
	"fmt"
	"net/http"
	"time"

	"awo.so/internal/core/finance/domain"
	corePipeline "awo.so/internal/pipeline"
	sharederrors "awo.so/internal/shared/errors"
)

// ── LoadTransactionStage ──────────────────────────────────────────────────────

// LoadTransactionStage fetches the transaction header and its entries from the
// repository and stores them in the OperationContext data map.
// All subsequent stages read from the data map rather than hitting the DB again.
type LoadTransactionStage struct {
	corePipeline.BaseStage
	repo domain.TransactionRepository
}

func NewLoadTransactionStage(repo domain.TransactionRepository) *LoadTransactionStage {
	return &LoadTransactionStage{
		BaseStage: corePipeline.BaseStage{
			StageName:       "gl.load_transaction",
			StageOperations: []string{"finance.transaction.post"},
			StagePriority:   100,
			StageRequired:   true,
		},
		repo: repo,
	}
}

func (s *LoadTransactionStage) Execute(opCtx *corePipeline.OperationContext) (corePipeline.StageResult, error) {
	input, ok := opCtx.Input.(*PostTransactionInput)
	if !ok || input == nil {
		return corePipeline.StageResult{Status: "failed"}, fmt.Errorf("gl.load_transaction: missing or invalid PostTransactionInput")
	}

	txn, err := s.repo.GetByID(opCtx.Ctx, input.TransactionID)
	if err != nil {
		return corePipeline.StageResult{Status: "failed"}, fmt.Errorf("gl.load_transaction: %w", err)
	}

	entries, err := s.repo.GetEntriesByTransaction(opCtx.Ctx, input.TransactionID)
	if err != nil {
		return corePipeline.StageResult{Status: "failed"}, fmt.Errorf("gl.load_transaction: get entries: %w", err)
	}

	opCtx.SetData(KeyTransaction, txn)
	opCtx.SetData(KeyEntries, entries)

	return corePipeline.StageResult{
		Status:  "completed",
		Message: fmt.Sprintf("loaded transaction %s with %d entries", txn.TransactionNumber, len(entries)),
		Outputs: map[string]any{
			KeyTransaction: txn,
			KeyEntries:     entries,
		},
	}, nil
}

// ── BalanceCheckStage ─────────────────────────────────────────────────────────

// BalanceCheckStage validates that the transaction's entries are balanced
// (total debits == total credits) and that there are at least two entries.
type BalanceCheckStage struct {
	corePipeline.BaseStage
}

func NewBalanceCheckStage() *BalanceCheckStage {
	return &BalanceCheckStage{
		BaseStage: corePipeline.BaseStage{
			StageName:       "gl.balance_check",
			StageOperations: []string{"finance.transaction.post"},
			StagePriority:   200,
			StageRequired:   true,
		},
	}
}

func (s *BalanceCheckStage) Execute(opCtx *corePipeline.OperationContext) (corePipeline.StageResult, error) {
	entries, ok := opCtx.Data[KeyEntries].([]domain.TransactionEntry)
	if !ok {
		return corePipeline.StageResult{Status: "failed"}, fmt.Errorf("gl.balance_check: entries not loaded")
	}

	if len(entries) < 2 {
		return corePipeline.StageResult{Status: "failed"},
			sharederrors.NewBusinessError("INSUFFICIENT_ENTRIES", "transaction must have at least two entries").
				WithHTTPStatus(http.StatusUnprocessableEntity)
	}

	// Delegate to domain validator — covers: balance check, both-debit-and-credit, no-amount rules.
	validator := domain.NewBusinessRuleValidator()
	validationErrors := validator.ValidateTransactionBalance(entries)

	if len(validationErrors) > 0 {
		first := validationErrors[0]
		return corePipeline.StageResult{Status: "failed"},
			sharederrors.NewBusinessError(first.Code, first.Message).
				WithHTTPStatus(http.StatusUnprocessableEntity)
	}

	return corePipeline.StageResult{Status: "completed", Message: "entries are balanced"}, nil
}

// ── PeriodCheckStage ──────────────────────────────────────────────────────────

// PeriodCheckStage verifies that the posting date falls within an open accounting
// period. Requires a PeriodRepository; if nil the stage is skipped gracefully.
type PeriodCheckStage struct {
	corePipeline.BaseStage
	periodRepo domain.PeriodRepository
}

func NewPeriodCheckStage(periodRepo domain.PeriodRepository) *PeriodCheckStage {
	return &PeriodCheckStage{
		BaseStage: corePipeline.BaseStage{
			StageName:       "gl.period_check",
			StageOperations: []string{"finance.transaction.post"},
			StagePriority:   300,
			StageRequired:   true,
		},
		periodRepo: periodRepo,
	}
}

func (s *PeriodCheckStage) Execute(opCtx *corePipeline.OperationContext) (corePipeline.StageResult, error) {
	if s.periodRepo == nil {
		return corePipeline.StageResult{Status: "skipped", Message: "no PeriodRepository configured"}, nil
	}

	input, ok := opCtx.Input.(*PostTransactionInput)
	if !ok || input == nil {
		return corePipeline.StageResult{Status: "failed"}, fmt.Errorf("gl.period_check: missing PostTransactionInput")
	}

	postingDate := input.PostingDate
	if postingDate.IsZero() {
		postingDate = time.Now()
	}

	period, err := s.periodRepo.GetPeriodForDate(opCtx.Ctx, input.TenantID, postingDate)
	if err != nil {
		return corePipeline.StageResult{Status: "failed"},
			sharederrors.NewBusinessError("PERIOD_NOT_FOUND",
				fmt.Sprintf("no accounting period found for date %s", postingDate.Format("2006-01-02"))).
				WithHTTPStatus(http.StatusUnprocessableEntity)
	}

	if !period.Status.AllowsPosting() {
		return corePipeline.StageResult{Status: "failed"},
			sharederrors.NewBusinessError("PERIOD_CLOSED",
				fmt.Sprintf("accounting period %q (%s) is %s — posting not allowed",
					period.Name, postingDate.Format("2006-01-02"), period.Status)).
				WithHTTPStatus(http.StatusUnprocessableEntity)
	}

	opCtx.SetData(KeyPeriod, period)

	return corePipeline.StageResult{
		Status:  "completed",
		Message: fmt.Sprintf("period %q is open for posting", period.Name),
		Outputs: map[string]any{KeyPeriod: period},
	}, nil
}

// ── AccountValidateStage ──────────────────────────────────────────────────────

// AccountValidateStage verifies that every entry references an account that
// exists, is active, and accepts manual entries.
type AccountValidateStage struct {
	corePipeline.BaseStage
	accountRepo domain.AccountsRepository
}

func NewAccountValidateStage(accountRepo domain.AccountsRepository) *AccountValidateStage {
	return &AccountValidateStage{
		BaseStage: corePipeline.BaseStage{
			StageName:       "gl.account_validate",
			StageOperations: []string{"finance.transaction.post"},
			StagePriority:   400,
			StageRequired:   true,
		},
		accountRepo: accountRepo,
	}
}

func (s *AccountValidateStage) Execute(opCtx *corePipeline.OperationContext) (corePipeline.StageResult, error) {
	entries, ok := opCtx.Data[KeyEntries].([]domain.TransactionEntry)
	if !ok {
		return corePipeline.StageResult{Status: "failed"}, fmt.Errorf("gl.account_validate: entries not loaded")
	}

	for i, entry := range entries {
		account, err := s.accountRepo.GetByID(opCtx.Ctx, entry.AccountID)
		if err != nil {
			return corePipeline.StageResult{Status: "failed"},
				sharederrors.NewBusinessError("INVALID_ACCOUNT",
					fmt.Sprintf("entries[%d].account_id: account %s not found", i, entry.AccountID)).
					WithHTTPStatus(http.StatusUnprocessableEntity)
		}

		if !account.CanAcceptTransactions() {
			return corePipeline.StageResult{Status: "failed"},
				sharederrors.NewBusinessError("ACCOUNT_NOT_ACTIVE",
					fmt.Sprintf("entries[%d].account_id: account %s (%s) cannot accept transactions — status: %s",
						i, account.AccountCode, account.AccountName, account.Status)).
					WithHTTPStatus(http.StatusUnprocessableEntity)
		}
	}

	return corePipeline.StageResult{
		Status:  "completed",
		Message: fmt.Sprintf("validated %d accounts", len(entries)),
	}, nil
}

// ── GLPostStage ───────────────────────────────────────────────────────────────

// GLPostStage marks the transaction as POSTED in the database and updates
// account balances. This stage should run last (highest priority) since it
// performs the irreversible write to the general ledger.
type GLPostStage struct {
	corePipeline.BaseStage
	repo domain.TransactionRepository
}

func NewGLPostStage(repo domain.TransactionRepository) *GLPostStage {
	return &GLPostStage{
		BaseStage: corePipeline.BaseStage{
			StageName:       "gl.post",
			StageOperations: []string{"finance.transaction.post"},
			StagePriority:   600,
			StageRequired:   true,
		},
		repo: repo,
	}
}

func (s *GLPostStage) Execute(opCtx *corePipeline.OperationContext) (corePipeline.StageResult, error) {
	input, ok := opCtx.Input.(*PostTransactionInput)
	if !ok || input == nil {
		return corePipeline.StageResult{Status: "failed"}, fmt.Errorf("gl.post: missing PostTransactionInput")
	}

	// Use the state machine to confirm posting is valid from the current status.
	txn, hasTxn := opCtx.Data[KeyTransaction].(*domain.Transaction)
	if hasTxn && txn != nil {
		tsm := domain.NewTransactionStateMachine(txn)
		if canPost, reason := tsm.CanTransitionTo(domain.TransactionStatusPosted); !canPost {
			return corePipeline.StageResult{Status: "failed"},
				sharederrors.NewBusinessError("CANNOT_POST", reason).
					WithHTTPStatus(http.StatusUnprocessableEntity)
		}
	}

	postingDate := input.PostingDate
	if postingDate.IsZero() {
		postingDate = time.Now()
	}

	if err := s.repo.Post(opCtx.Ctx, input.TransactionID, input.PostedBy, postingDate); err != nil {
		return corePipeline.StageResult{Status: "failed"},
			fmt.Errorf("gl.post: failed to post transaction: %w", err)
	}

	opCtx.SetFlag("gl_posted", true)
	opCtx.SetData("gl.posting_date", postingDate)

	return corePipeline.StageResult{
		Status:  "completed",
		Message: fmt.Sprintf("transaction posted on %s", postingDate.Format("2006-01-02")),
		Outputs: map[string]any{
			"gl.posting_date": postingDate,
			"gl.posted":       true,
		},
	}, nil
}
