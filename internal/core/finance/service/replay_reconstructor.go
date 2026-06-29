// Package service — Ledger Replay Reconstructor.
//
// ReplayReconstructor rebuilds account balances by replaying all POSTED
// transaction entries from the ledger. The result is compared against stored
// balances to detect drift — the primary signal of silent data corruption.
//
// # Replay contract
//
//  1. Only POSTED transactions are included in the replay.
//  2. Entries are aggregated per account as (totalDebit, totalCredit).
//  3. NetBalance = totalDebit - totalCredit (sign convention: raw, not
//     sign-adjusted for normal balance).
//  4. A drift is detected when the replayed net balance differs from the
//     account's stored CurrentBalance.
//
// # Why this matters
//
// Direct balance updates that bypass the transaction ledger will produce
// a drift. This surfaces direct-SQL manipulation, bug-induced balance
// resets, and migration defects before they contaminate financial reports.
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"awo.so/internal/core/finance/domain"
	"awo.so/internal/shared"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
)

// ReplayReconstructor replays the transaction ledger to verify stored balances.
type ReplayReconstructor struct {
	accounts     domain.AccountsRepository
	transactions domain.TransactionRepository
	metrics      metrics.MetricsProvider
}

// NewReplayReconstructor constructs the reconstructor.
func NewReplayReconstructor(
	accounts domain.AccountsRepository,
	transactions domain.TransactionRepository,
	m metrics.MetricsProvider,
) *ReplayReconstructor {
	return &ReplayReconstructor{
		accounts:     accounts,
		transactions: transactions,
		metrics:      m,
	}
}

// =============================================================================
// Full tenant replay
// =============================================================================

// ReplayTenant replays all posted transactions for the tenant in ctx and
// builds a ReplayReport. asOfDate = zero means replay up to now.
func (r *ReplayReconstructor) ReplayTenant(
	ctx context.Context,
	asOfDate time.Time,
) (*domain.ReplayReport, error) {
	if r == nil {
		return &domain.ReplayReport{ReplayClean: true, GeneratedAt: time.Now()}, nil
	}

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("replay tenant: tenant ID missing from context")
	}

	if asOfDate.IsZero() {
		asOfDate = time.Now().UTC()
	}

	// Load all posted transactions up to asOfDate.
	posted := transactionStatusPosted()
	filter := &domain.TransactionFilter{
		Status: &posted,
		DateRange: &domain.DateRange{
			StartDate: time.Time{}, // beginning of time
			EndDate:   &asOfDate,
		},
	}
	transactions, err := r.transactions.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("replay tenant: load transactions: %w", err)
	}

	// Accumulate debit/credit totals per account.
	type accumulator struct {
		debit  decimal.Decimal
		credit decimal.Decimal
		count  int
	}
	acc := make(map[uuid.UUID]*accumulator)
	totalEntries := 0

	for _, txn := range transactions {
		entries, err := r.transactions.GetEntriesByTransaction(ctx, txn.ID)
		if err != nil {
			return nil, fmt.Errorf("replay tenant: load entries for txn %s: %w", txn.ID, err)
		}
		for _, e := range entries {
			if _, ok := acc[e.AccountID]; !ok {
				acc[e.AccountID] = &accumulator{}
			}
			acc[e.AccountID].debit = acc[e.AccountID].debit.Add(e.DebitAmount)
			acc[e.AccountID].credit = acc[e.AccountID].credit.Add(e.CreditAmount)
			acc[e.AccountID].count++
			totalEntries++
		}
	}

	// Load all accounts to compare replayed balance vs stored balance.
	accounts, err := r.accounts.List(ctx, &domain.AccountFilter{})
	if err != nil {
		return nil, fmt.Errorf("replay tenant: load accounts: %w", err)
	}

	balances := make([]domain.ReplayedBalance, 0, len(acc))
	var driftAccounts []uuid.UUID

	for _, a := range accounts {
		a := a
		entry, hasEntries := acc[a.ID]

		var debit, credit decimal.Decimal
		txnCount := 0
		if hasEntries {
			debit = entry.debit
			credit = entry.credit
			txnCount = entry.count
		}

		net := debit.Sub(credit)
		rb := domain.ReplayedBalance{
			AccountID:      a.ID,
			AccountCode:    a.AccountCode,
			ReplayedDebit:  debit,
			ReplayedCredit: credit,
			NetBalance:     net,
			TxnCount:       txnCount,
		}
		balances = append(balances, rb)

		// Detect drift: stored balance differs from replayed net.
		if !net.Equal(a.CurrentBalance) {
			driftAccounts = append(driftAccounts, a.ID)
		}
	}

	report := &domain.ReplayReport{
		TenantID:        tenantID,
		AsOfDate:        asOfDate,
		GeneratedAt:     time.Now().UTC(),
		Balances:        balances,
		TxnsReplayed:    len(transactions),
		EntriesReplayed: totalEntries,
		DriftAccounts:   driftAccounts,
		ReplayClean:     len(driftAccounts) == 0,
	}

	r.emitReplayMetrics(ctx, report)

	logger.InfoContext(ctx, "ledger replay complete", logger.Fields{
		"tenant_id":   tenantID.String(),
		"txns":        report.TxnsReplayed,
		"entries":     report.EntriesReplayed,
		"drift_count": len(driftAccounts),
		"clean":       report.ReplayClean,
	})

	return report, nil
}

// =============================================================================
// Single account replay
// =============================================================================

// ReplayAccount replays the ledger for a single account and returns the
// reconstructed balance. Does not compare against stored balance.
func (r *ReplayReconstructor) ReplayAccount(
	ctx context.Context,
	accountID uuid.UUID,
	asOfDate time.Time,
) (*domain.ReplayedBalance, error) {
	if r == nil {
		return nil, fmt.Errorf("ReplayReconstructor is nil")
	}

	if asOfDate.IsZero() {
		asOfDate = time.Now().UTC()
	}

	a, err := r.accounts.GetByID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("replay account: load account: %w", err)
	}

	posted := transactionStatusPosted()
	filter := &domain.TransactionFilter{
		Status: &posted,
		DateRange: &domain.DateRange{
			StartDate: time.Time{},
			EndDate:   &asOfDate,
		},
	}
	transactions, err := r.transactions.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("replay account: load transactions: %w", err)
	}

	var totalDebit, totalCredit decimal.Decimal
	txnCount := 0

	for _, txn := range transactions {
		entries, err := r.transactions.GetEntriesByTransaction(ctx, txn.ID)
		if err != nil {
			return nil, fmt.Errorf("replay account: load entries for txn %s: %w", txn.ID, err)
		}
		for _, e := range entries {
			if e.AccountID != accountID {
				continue
			}
			totalDebit = totalDebit.Add(e.DebitAmount)
			totalCredit = totalCredit.Add(e.CreditAmount)
			txnCount++
		}
	}

	return &domain.ReplayedBalance{
		AccountID:      a.ID,
		AccountCode:    a.AccountCode,
		ReplayedDebit:  totalDebit,
		ReplayedCredit: totalCredit,
		NetBalance:     totalDebit.Sub(totalCredit),
		TxnCount:       txnCount,
	}, nil
}

// =============================================================================
// Helpers
// =============================================================================

// transactionStatusPosted returns the POSTED status value.
// Centralised to avoid stringly-typed scattered references.
func transactionStatusPosted() domain.TransactionStatus {
	return domain.TransactionStatusPosted
}

// =============================================================================
// Metrics
// =============================================================================

func (r *ReplayReconstructor) emitReplayMetrics(ctx context.Context, report *domain.ReplayReport) {
	if r.metrics == nil {
		return
	}
	r.metrics.IncrementCounter("finance_ledger_replays_total", metrics.Fields{
		"tenant_id": report.TenantID.String(),
		"clean":     fmt.Sprintf("%v", report.ReplayClean),
	})
	if len(report.DriftAccounts) > 0 {
		r.metrics.IncrementCounter("finance_ledger_drift_accounts_total", metrics.Fields{
			"tenant_id": report.TenantID.String(),
		})
	}
}
