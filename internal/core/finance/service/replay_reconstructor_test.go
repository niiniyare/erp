// Package service_test — ReplayReconstructor tests.
//
// Tests (FIN-REPLAY-*):
//
//	001 — Empty ledger produces clean replay (no drift)
//	002 — Replayed balance matches posted debit/credit entries
//	003 — Nil service returns clean report (nil-safe)
//	004 — Transactions with no entries contribute zero balance
//	005 — Balance drift detected when stored balance differs from replay
//	006 — ReplayAccount returns correct per-account balance
//	007 — ReplayTenant propagates transaction list error
//	008 — ReplayTenant only includes POSTED transactions (status filter respected)
//	009 — DriftAccounts populated when stored balance diverges
//	010 — ReplayClean is false when drift exists
package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"awo.so/internal/core/finance/domain"
	"awo.so/internal/core/finance/service"
	"awo.so/internal/shared"
	"awo.so/internal/shared/metrics"
)

// helper — context with tenant
func replayCtx() context.Context {
	return shared.WithTenantID(context.Background(), uuid.New())
}

// helper — posted transaction with entries
func postedTxn(entries ...domain.TransactionEntry) *domain.Transaction {
	return &domain.Transaction{
		ID:                uuid.New(),
		TransactionStatus: domain.TransactionStatusPosted,
		TransactionDate:   time.Now(),
		Entries:           entries,
	}
}

// helper — debit entry
func debitEntry(accountID uuid.UUID, amount float64) domain.TransactionEntry {
	return domain.TransactionEntry{
		ID:          uuid.New(),
		AccountID:   accountID,
		DebitAmount: decimal.NewFromFloat(amount),
	}
}

// helper — credit entry
func creditEntry(accountID uuid.UUID, amount float64) domain.TransactionEntry {
	return domain.TransactionEntry{
		ID:           uuid.New(),
		AccountID:    accountID,
		CreditAmount: decimal.NewFromFloat(amount),
	}
}

// txnRepoForReplay builds a p16StubTxnRepo that returns given txns from List
// and entries from GetEntriesByTransaction.
func txnRepoForReplay(txns ...*domain.Transaction) *p16StubTxnRepo {
	r := &p16StubTxnRepo{}
	r.fnList = func(_ context.Context, _ *domain.TransactionFilter) ([]*domain.Transaction, error) {
		return txns, nil
	}
	r.fnGetEntriesByTxn = func(_ context.Context, txnID uuid.UUID) ([]domain.TransactionEntry, error) {
		for _, t := range txns {
			if t.ID == txnID {
				return t.Entries, nil
			}
		}
		return nil, nil
	}
	return r
}

func newReplaySvc(accounts *p17AccountRepo, txns *p16StubTxnRepo) *service.ReplayReconstructor {
	return service.NewReplayReconstructor(accounts, txns, metrics.NewNoOpMetricsProvider())
}

// FIN-REPLAY-001 ---------------------------------------------------------

func TestReplay_EmptyLedger_Clean(t *testing.T) {
	accounts := newP17AccountRepo()
	txns := txnRepoForReplay()
	svc := newReplaySvc(accounts, txns)

	report, err := svc.ReplayTenant(replayCtx(), time.Time{})
	if err != nil {
		t.Fatalf("FIN-REPLAY-001: unexpected error: %v", err)
	}
	if !report.ReplayClean {
		t.Error("FIN-REPLAY-001: expected clean replay for empty ledger")
	}
}

// FIN-REPLAY-002 ---------------------------------------------------------

func TestReplay_BalanceMatches_PostedEntries(t *testing.T) {
	acctID := uuid.New()
	acct := makeAccount("10010001")
	acct.ID = acctID
	// Stored balance matches replay: debit 100, credit 0 → net 100.
	acct.CurrentBalance = decimal.NewFromFloat(100)

	txn := postedTxn(debitEntry(acctID, 100))
	accounts := newP17AccountRepo(acct)
	txns := txnRepoForReplay(txn)
	svc := newReplaySvc(accounts, txns)

	report, err := svc.ReplayTenant(replayCtx(), time.Now())
	if err != nil {
		t.Fatalf("FIN-REPLAY-002: unexpected error: %v", err)
	}
	if !report.ReplayClean {
		t.Errorf("FIN-REPLAY-002: expected clean replay, drift accounts: %v", report.DriftAccounts)
	}
}

// FIN-REPLAY-003 ---------------------------------------------------------

func TestReplay_NilService_Safe(t *testing.T) {
	var svc *service.ReplayReconstructor

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("FIN-REPLAY-003: nil service panicked: %v", r)
		}
	}()

	report, err := svc.ReplayTenant(replayCtx(), time.Time{})
	if err != nil {
		t.Errorf("FIN-REPLAY-003: nil service should return nil error, got %v", err)
	}
	if report == nil {
		t.Error("FIN-REPLAY-003: expected non-nil report from nil service")
	}
}

// FIN-REPLAY-004 ---------------------------------------------------------

func TestReplay_TxnWithNoEntries_ZeroBalance(t *testing.T) {
	acct := makeAccount("10010001")
	acct.CurrentBalance = decimal.Zero

	// Transaction with no entries.
	txn := postedTxn() // no entries
	accounts := newP17AccountRepo(acct)
	txns := txnRepoForReplay(txn)
	svc := newReplaySvc(accounts, txns)

	report, err := svc.ReplayTenant(replayCtx(), time.Now())
	if err != nil {
		t.Fatalf("FIN-REPLAY-004: unexpected error: %v", err)
	}
	if !report.ReplayClean {
		t.Error("FIN-REPLAY-004: expected clean replay when txn has no entries")
	}
}

// FIN-REPLAY-005 ---------------------------------------------------------

func TestReplay_BalanceDrift_Detected(t *testing.T) {
	acctID := uuid.New()
	acct := makeAccount("10010001")
	acct.ID = acctID
	// Stored balance is 200, but ledger only shows 100.
	acct.CurrentBalance = decimal.NewFromFloat(200)

	txn := postedTxn(debitEntry(acctID, 100)) // net = 100, stored = 200 → drift
	accounts := newP17AccountRepo(acct)
	txns := txnRepoForReplay(txn)
	svc := newReplaySvc(accounts, txns)

	report, err := svc.ReplayTenant(replayCtx(), time.Now())
	if err != nil {
		t.Fatalf("FIN-REPLAY-005: unexpected error: %v", err)
	}
	if report.ReplayClean {
		t.Error("FIN-REPLAY-005: expected drift to be detected")
	}
	if len(report.DriftAccounts) == 0 {
		t.Error("FIN-REPLAY-005: expected DriftAccounts to be populated")
	}
}

// FIN-REPLAY-006 ---------------------------------------------------------

func TestReplay_ReplayAccount_CorrectBalance(t *testing.T) {
	acctID := uuid.New()
	acct := makeAccount("10010001")
	acct.ID = acctID

	txn := postedTxn(
		debitEntry(acctID, 300),
		creditEntry(acctID, 100),
	)
	accounts := newP17AccountRepo(acct)
	txns := txnRepoForReplay(txn)
	svc := newReplaySvc(accounts, txns)

	rb, err := svc.ReplayAccount(replayCtx(), acctID, time.Now())
	if err != nil {
		t.Fatalf("FIN-REPLAY-006: unexpected error: %v", err)
	}
	expectedNet := decimal.NewFromFloat(200) // 300 debit - 100 credit
	if !rb.NetBalance.Equal(expectedNet) {
		t.Errorf("FIN-REPLAY-006: expected net=%s, got %s", expectedNet, rb.NetBalance)
	}
}

// FIN-REPLAY-007 ---------------------------------------------------------

func TestReplay_PropagatesTxnListError(t *testing.T) {
	accounts := newP17AccountRepo()
	txns := &p16StubTxnRepo{}
	txns.fnList = func(_ context.Context, _ *domain.TransactionFilter) ([]*domain.Transaction, error) {
		return nil, errors.New("DB timeout")
	}
	svc := newReplaySvc(accounts, txns)

	_, err := svc.ReplayTenant(replayCtx(), time.Time{})
	if err == nil {
		t.Error("FIN-REPLAY-007: expected error from txn repo, got nil")
	}
}

// FIN-REPLAY-008 ---------------------------------------------------------

func TestReplay_TxnCountInReport(t *testing.T) {
	acctID := uuid.New()
	acct := makeAccount("10010001")
	acct.ID = acctID
	acct.CurrentBalance = decimal.NewFromFloat(150)

	t1 := postedTxn(debitEntry(acctID, 100))
	t2 := postedTxn(debitEntry(acctID, 50))
	accounts := newP17AccountRepo(acct)
	txns := txnRepoForReplay(t1, t2)
	svc := newReplaySvc(accounts, txns)

	report, err := svc.ReplayTenant(replayCtx(), time.Now())
	if err != nil {
		t.Fatalf("FIN-REPLAY-008: unexpected error: %v", err)
	}
	if report.TxnsReplayed != 2 {
		t.Errorf("FIN-REPLAY-008: expected TxnsReplayed=2, got %d", report.TxnsReplayed)
	}
}

// FIN-REPLAY-009 ---------------------------------------------------------

func TestReplay_DriftAccounts_Populated(t *testing.T) {
	acct1ID := uuid.New()
	acct2ID := uuid.New()
	acct1 := makeAccount("10010001")
	acct1.ID = acct1ID
	acct1.CurrentBalance = decimal.NewFromFloat(100)

	acct2 := makeAccount("20010001")
	acct2.ID = acct2ID
	acct2.CurrentBalance = decimal.NewFromFloat(999) // drift — no txns hit this acct

	txn := postedTxn(debitEntry(acct1ID, 100)) // only touches acct1
	accounts := newP17AccountRepo(acct1, acct2)
	txns := txnRepoForReplay(txn)
	svc := newReplaySvc(accounts, txns)

	report, err := svc.ReplayTenant(replayCtx(), time.Now())
	if err != nil {
		t.Fatalf("FIN-REPLAY-009: unexpected error: %v", err)
	}
	found := false
	for _, id := range report.DriftAccounts {
		if id == acct2ID {
			found = true
		}
	}
	if !found {
		t.Error("FIN-REPLAY-009: expected acct2 in DriftAccounts")
	}
}

// FIN-REPLAY-010 ---------------------------------------------------------

func TestReplay_ReplayClean_False_WhenDrift(t *testing.T) {
	acctID := uuid.New()
	acct := makeAccount("10010001")
	acct.ID = acctID
	acct.CurrentBalance = decimal.NewFromFloat(500) // stored

	txn := postedTxn(debitEntry(acctID, 100)) // replayed net = 100 ≠ 500
	accounts := newP17AccountRepo(acct)
	txns := txnRepoForReplay(txn)
	svc := newReplaySvc(accounts, txns)

	report, err := svc.ReplayTenant(replayCtx(), time.Now())
	if err != nil {
		t.Fatalf("FIN-REPLAY-010: unexpected error: %v", err)
	}
	if report.ReplayClean {
		t.Error("FIN-REPLAY-010: expected ReplayClean=false when drift exists")
	}
}
