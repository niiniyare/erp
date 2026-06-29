// Package service_test — Phase 16 shared test stubs.
//
// All types use the "p16" prefix to avoid conflicts with stubs defined in
// earlier-phase test files within the same package.
package service_test

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"awo.so/internal/core/finance/domain"
)

// ── p16StubTxnRepo ──────────────────────────────────────────────────────────
// Minimal stub for domain.TransactionRepository.
// All fields are function hooks — nil hooks return zero values with no error.
// Used by Phase 16 tests that exercise IntegrityService and TransactionService
// without a real database.

type p16StubTxnRepo struct {
	fnCreate          func(ctx context.Context, t *domain.Transaction) error
	fnGetByID         func(ctx context.Context, id uuid.UUID) (*domain.Transaction, error)
	fnGetByNumber     func(ctx context.Context, entityID *uuid.UUID, number string) (*domain.Transaction, error)
	fnUpdate          func(ctx context.Context, t *domain.Transaction) error
	fnDelete          func(ctx context.Context, id uuid.UUID) error
	fnList            func(ctx context.Context, filter *domain.TransactionFilter) ([]*domain.Transaction, error)
	fnGetEntriesByTxn func(ctx context.Context, transactionID uuid.UUID) ([]domain.TransactionEntry, error)
	fnPost            func(ctx context.Context, transactionID uuid.UUID, postedBy uuid.UUID, postedAt time.Time) error
	fnApprove         func(ctx context.Context, transactionID uuid.UUID, approvedBy uuid.UUID, approvedAt time.Time, notes *string) error
	fnReject          func(ctx context.Context, transactionID uuid.UUID, rejectedBy uuid.UUID, rejectedAt time.Time, reason domain.RejectionReason, notes *string) error
	fnReverse         func(ctx context.Context, originalID, reversalID uuid.UUID, reason string) error
	fnIsNumberUnique  func(ctx context.Context, entityID *uuid.UUID, number string, excludeID *uuid.UUID) (bool, error)
	fnGetNextNumber   func(ctx context.Context, entityID *uuid.UUID, txType domain.TransactionType) (string, error)
}

func (r *p16StubTxnRepo) Create(ctx context.Context, t *domain.Transaction) error {
	if r.fnCreate != nil {
		return r.fnCreate(ctx, t)
	}
	return nil
}

func (r *p16StubTxnRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Transaction, error) {
	if r.fnGetByID != nil {
		return r.fnGetByID(ctx, id)
	}
	return nil, nil
}

func (r *p16StubTxnRepo) GetByNumber(ctx context.Context, entityID *uuid.UUID, number string) (*domain.Transaction, error) {
	if r.fnGetByNumber != nil {
		return r.fnGetByNumber(ctx, entityID, number)
	}
	return nil, nil
}

func (r *p16StubTxnRepo) Update(ctx context.Context, t *domain.Transaction) error {
	if r.fnUpdate != nil {
		return r.fnUpdate(ctx, t)
	}
	return nil
}

func (r *p16StubTxnRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if r.fnDelete != nil {
		return r.fnDelete(ctx, id)
	}
	return nil
}

func (r *p16StubTxnRepo) List(ctx context.Context, filter *domain.TransactionFilter) ([]*domain.Transaction, error) {
	if r.fnList != nil {
		return r.fnList(ctx, filter)
	}
	return nil, nil
}

func (r *p16StubTxnRepo) Count(ctx context.Context, filter *domain.TransactionFilter) (int64, error) {
	return 0, nil
}

func (r *p16StubTxnRepo) ListByAccount(ctx context.Context, accountID uuid.UUID, filter *domain.TransactionFilter) ([]*domain.Transaction, error) {
	return nil, nil
}

func (r *p16StubTxnRepo) ListByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*domain.Transaction, error) {
	return nil, nil
}

func (r *p16StubTxnRepo) GetByStatus(ctx context.Context, status domain.TransactionStatus, limit int) ([]*domain.Transaction, error) {
	return nil, nil
}

func (r *p16StubTxnRepo) GetPendingApproval(ctx context.Context, userID *uuid.UUID) ([]*domain.Transaction, error) {
	return nil, nil
}

func (r *p16StubTxnRepo) GetRecurringTransactions(ctx context.Context, dueDate time.Time) ([]*domain.Transaction, error) {
	return nil, nil
}

func (r *p16StubTxnRepo) UpdateNextRecurringDate(ctx context.Context, transactionID uuid.UUID, nextDate time.Time) error {
	return nil
}

func (r *p16StubTxnRepo) CreateEntry(ctx context.Context, entry *domain.TransactionEntry) error {
	return nil
}

func (r *p16StubTxnRepo) CreateEntries(ctx context.Context, entries []*domain.TransactionEntry) error {
	return nil
}

func (r *p16StubTxnRepo) GetEntryByID(ctx context.Context, id uuid.UUID) (*domain.TransactionEntry, error) {
	return nil, nil
}

func (r *p16StubTxnRepo) GetEntriesByTransaction(ctx context.Context, transactionID uuid.UUID) ([]domain.TransactionEntry, error) {
	if r.fnGetEntriesByTxn != nil {
		return r.fnGetEntriesByTxn(ctx, transactionID)
	}
	return nil, nil
}

func (r *p16StubTxnRepo) GetEntriesByAccount(ctx context.Context, accountID uuid.UUID, filter *domain.EntryFilter) ([]domain.TransactionEntry, error) {
	return nil, nil
}

func (r *p16StubTxnRepo) UpdateEntry(ctx context.Context, entry *domain.TransactionEntry) error {
	return nil
}

func (r *p16StubTxnRepo) DeleteEntry(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (r *p16StubTxnRepo) SearchEntries(ctx context.Context, query string, filters *domain.EntryFilter, limit int, offset int) ([]*domain.TransactionEntry, error) {
	return nil, nil
}

func (r *p16StubTxnRepo) UpdateReconciliationStatus(ctx context.Context, entryID uuid.UUID, reconciled bool, reconciledDate *time.Time, reconciliationRef *string) error {
	return nil
}

func (r *p16StubTxnRepo) GetUnreconciledEntries(ctx context.Context, accountID uuid.UUID, cutoffDate *time.Time) ([]*domain.TransactionEntry, error) {
	return nil, nil
}

func (r *p16StubTxnRepo) GetEntrySummary(ctx context.Context, accountID uuid.UUID, startDate, endDate time.Time) (*domain.TransactionSummary, error) {
	return nil, nil
}

func (r *p16StubTxnRepo) CalculateAccountBalance(ctx context.Context, accountID uuid.UUID, asOfDate *time.Time) (decimal.Decimal, error) {
	return decimal.Zero, nil
}

func (r *p16StubTxnRepo) GetAccountTransactionSummary(ctx context.Context, accountID uuid.UUID, dateRange *domain.DateRange) (*domain.TransactionSummary, error) {
	return nil, nil
}

func (r *p16StubTxnRepo) Post(ctx context.Context, transactionID uuid.UUID, postedBy uuid.UUID, postedAt time.Time) error {
	if r.fnPost != nil {
		return r.fnPost(ctx, transactionID, postedBy, postedAt)
	}
	return nil
}

func (r *p16StubTxnRepo) Approve(ctx context.Context, transactionID uuid.UUID, approvedBy uuid.UUID, approvedAt time.Time, notes *string) error {
	if r.fnApprove != nil {
		return r.fnApprove(ctx, transactionID, approvedBy, approvedAt, notes)
	}
	return nil
}

func (r *p16StubTxnRepo) Reject(ctx context.Context, transactionID uuid.UUID, rejectedBy uuid.UUID, rejectedAt time.Time, reason domain.RejectionReason, notes *string) error {
	if r.fnReject != nil {
		return r.fnReject(ctx, transactionID, rejectedBy, rejectedAt, reason, notes)
	}
	return nil
}

func (r *p16StubTxnRepo) Reverse(ctx context.Context, originalID, reversalID uuid.UUID, reason string) error {
	if r.fnReverse != nil {
		return r.fnReverse(ctx, originalID, reversalID, reason)
	}
	return nil
}

func (r *p16StubTxnRepo) IsTransactionNumberUnique(ctx context.Context, entityID *uuid.UUID, number string, excludeID *uuid.UUID) (bool, error) {
	if r.fnIsNumberUnique != nil {
		return r.fnIsNumberUnique(ctx, entityID, number, excludeID)
	}
	return true, nil
}

func (r *p16StubTxnRepo) ValidateAccountsExist(ctx context.Context, accountIDs []uuid.UUID) error {
	return nil
}

func (r *p16StubTxnRepo) GetNextTransactionNumber(ctx context.Context, entityID *uuid.UUID, txType domain.TransactionType) (string, error) {
	if r.fnGetNextNumber != nil {
		return r.fnGetNextNumber(ctx, entityID, txType)
	}
	return "TXN-P16-AUTO", nil
}
