// Package service_test — DeterministicReportingEngine tests.
//
// Tests (FIN-DREPORT-*):
//
//	001 — GenerateTrialBalance returns balanced report when debits == credits
//	002 — GenerateTrialBalance produces non-empty hash
//	003 — Two identical inputs produce identical hashes (deterministic)
//	004 — Different ledger produces different hash
//	005 — Nil service returns error
//	006 — GenerateTrialBalance fails when snapshot not found
//	007 — Report lines sorted by AccountCode
//	008 — TotalDebits and TotalCredits aggregate correctly
//	009 — VerifyReportHash returns true for reproducible report
//	010 — VerifyReportHash returns false when ledger mutated
//	011 — GenerateTrialBalance without tenant ID in context returns error
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

// helper — build a snap with given entries
func buildSnap(tenantID uuid.UUID, entries ...domain.COASnapshotEntry) *domain.COASnapshot {
	return &domain.COASnapshot{
		ID:         uuid.New(),
		TenantID:   tenantID,
		SnapshotAt: time.Now(),
		Accounts:   entries,
		Hash:       "testhash",
		EntryCount: len(entries),
		CreatedBy:  uuid.New(),
		CreatedAt:  time.Now(),
	}
}

// helper — build a snapshot entry from an account
func snapEntryFrom(a *domain.Accounts) domain.COASnapshotEntry {
	return domain.COASnapshotEntry{
		AccountID:     a.ID,
		AccountCode:   a.AccountCode,
		AccountName:   a.AccountName,
		RootType:      a.RootType,
		NormalBalance: a.NormalBalance,
		Status:        a.Status,
		IsActive:      a.IsActive,
	}
}

// helper — txn repo that returns entries for known txn IDs
func dreportTxnRepo(txns ...*domain.Transaction) *p16StubTxnRepo {
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

func newDReportSvc(snapRepo *p17StubSnapshotRepo, txns *p16StubTxnRepo) *service.DeterministicReportingEngine {
	return service.NewDeterministicReportingEngine(snapRepo, txns, metrics.NewNoOpMetricsProvider())
}

func dreportCtx() context.Context {
	return shared.WithTenantID(context.Background(), uuid.New())
}

// FIN-DREPORT-001 --------------------------------------------------------

func TestDReport_BalancedReport(t *testing.T) {
	acct1 := makeAccount("10010001")
	acct2 := makeAccount("20010001")

	snapRepo := &p17StubSnapshotRepo{}
	snap := buildSnap(uuid.New(), snapEntryFrom(acct1), snapEntryFrom(acct2))
	snapRepo.store = []*domain.COASnapshot{snap}

	// Balanced: 100 debit on acct1, 100 credit on acct2.
	txn := &domain.Transaction{
		ID:                uuid.New(),
		TransactionStatus: domain.TransactionStatusPosted,
		Entries: []domain.TransactionEntry{
			debitEntry(acct1.ID, 100),
			creditEntry(acct2.ID, 100),
		},
	}
	txns := dreportTxnRepo(txn)
	svc := newDReportSvc(snapRepo, txns)

	req := service.TrialBalanceRequest{SnapshotID: snap.ID, GeneratedBy: uuid.New()}
	report, err := svc.GenerateTrialBalance(dreportCtx(), req)
	if err != nil {
		t.Fatalf("FIN-DREPORT-001: unexpected error: %v", err)
	}
	if !report.Balanced {
		t.Errorf("FIN-DREPORT-001: expected balanced report, debits=%s credits=%s",
			report.TotalDebits, report.TotalCredits)
	}
}

// FIN-DREPORT-002 --------------------------------------------------------

func TestDReport_NonEmptyHash(t *testing.T) {
	acct1 := makeAccount("10010001")
	snapRepo := &p17StubSnapshotRepo{}
	snap := buildSnap(uuid.New(), snapEntryFrom(acct1))
	snapRepo.store = []*domain.COASnapshot{snap}

	svc := newDReportSvc(snapRepo, dreportTxnRepo())
	req := service.TrialBalanceRequest{SnapshotID: snap.ID, GeneratedBy: uuid.New()}
	report, err := svc.GenerateTrialBalance(dreportCtx(), req)
	if err != nil {
		t.Fatalf("FIN-DREPORT-002: unexpected error: %v", err)
	}
	if report.Hash == "" {
		t.Error("FIN-DREPORT-002: expected non-empty report hash")
	}
	if len(report.Hash) != 64 {
		t.Errorf("FIN-DREPORT-002: expected 64-char hash, got %d", len(report.Hash))
	}
}

// FIN-DREPORT-003 --------------------------------------------------------

func TestDReport_Deterministic_SameInputs(t *testing.T) {
	acct1 := makeAccount("10010001")
	snap := buildSnap(uuid.New(), snapEntryFrom(acct1))

	snapRepo1 := &p17StubSnapshotRepo{store: []*domain.COASnapshot{snap}}
	snapRepo2 := &p17StubSnapshotRepo{store: []*domain.COASnapshot{snap}}

	txn := postedTxn(debitEntry(acct1.ID, 100))
	svc1 := newDReportSvc(snapRepo1, dreportTxnRepo(txn))
	svc2 := newDReportSvc(snapRepo2, dreportTxnRepo(txn))

	req := service.TrialBalanceRequest{SnapshotID: snap.ID, GeneratedBy: uuid.New()}
	r1, err1 := svc1.GenerateTrialBalance(dreportCtx(), req)
	r2, err2 := svc2.GenerateTrialBalance(dreportCtx(), req)
	if err1 != nil || err2 != nil {
		t.Fatalf("FIN-DREPORT-003: errors: %v / %v", err1, err2)
	}
	if r1.Hash != r2.Hash {
		t.Errorf("FIN-DREPORT-003: hashes differ for identical inputs: %s vs %s", r1.Hash, r2.Hash)
	}
}

// FIN-DREPORT-004 --------------------------------------------------------

func TestDReport_DifferentLedger_DifferentHash(t *testing.T) {
	acct1 := makeAccount("10010001")
	snap := buildSnap(uuid.New(), snapEntryFrom(acct1))

	snapRepo1 := &p17StubSnapshotRepo{store: []*domain.COASnapshot{snap}}
	snapRepo2 := &p17StubSnapshotRepo{store: []*domain.COASnapshot{snap}}

	txn1 := postedTxn(debitEntry(acct1.ID, 100))
	txn2 := postedTxn(debitEntry(acct1.ID, 200)) // different amount

	svc1 := newDReportSvc(snapRepo1, dreportTxnRepo(txn1))
	svc2 := newDReportSvc(snapRepo2, dreportTxnRepo(txn2))

	req := service.TrialBalanceRequest{SnapshotID: snap.ID, GeneratedBy: uuid.New()}
	r1, _ := svc1.GenerateTrialBalance(dreportCtx(), req)
	r2, _ := svc2.GenerateTrialBalance(dreportCtx(), req)

	if r1.Hash == r2.Hash {
		t.Error("FIN-DREPORT-004: expected different hashes for different ledgers")
	}
}

// FIN-DREPORT-005 --------------------------------------------------------

func TestDReport_NilService_ReturnsError(t *testing.T) {
	var svc *service.DeterministicReportingEngine
	_, err := svc.GenerateTrialBalance(dreportCtx(), service.TrialBalanceRequest{})
	if err == nil {
		t.Error("FIN-DREPORT-005: expected error from nil service, got nil")
	}
}

// FIN-DREPORT-006 --------------------------------------------------------

func TestDReport_SnapshotNotFound_ReturnsError(t *testing.T) {
	snapRepo := &p17StubSnapshotRepo{
		fnGetByID: func(_ context.Context, _ uuid.UUID) (*domain.COASnapshot, error) {
			return nil, errors.New("not found")
		},
	}
	svc := newDReportSvc(snapRepo, dreportTxnRepo())
	req := service.TrialBalanceRequest{SnapshotID: uuid.New(), GeneratedBy: uuid.New()}
	_, err := svc.GenerateTrialBalance(dreportCtx(), req)
	if err == nil {
		t.Error("FIN-DREPORT-006: expected error when snapshot not found")
	}
}

// FIN-DREPORT-007 --------------------------------------------------------

func TestDReport_Lines_SortedByAccountCode(t *testing.T) {
	// Deliberately insert in reverse order to verify sorting.
	a1 := makeAccount("10010001")
	a2 := makeAccount("20010001")
	a3 := makeAccount("30010001")

	snap := buildSnap(uuid.New(), snapEntryFrom(a3), snapEntryFrom(a1), snapEntryFrom(a2))
	snapRepo := &p17StubSnapshotRepo{store: []*domain.COASnapshot{snap}}

	svc := newDReportSvc(snapRepo, dreportTxnRepo())
	req := service.TrialBalanceRequest{SnapshotID: snap.ID, GeneratedBy: uuid.New()}
	report, err := svc.GenerateTrialBalance(dreportCtx(), req)
	if err != nil {
		t.Fatalf("FIN-DREPORT-007: unexpected error: %v", err)
	}

	for i := 1; i < len(report.Lines); i++ {
		if report.Lines[i].AccountCode < report.Lines[i-1].AccountCode {
			t.Errorf("FIN-DREPORT-007: lines not sorted at index %d: %s < %s",
				i, report.Lines[i].AccountCode, report.Lines[i-1].AccountCode)
		}
	}
}

// FIN-DREPORT-008 --------------------------------------------------------

func TestDReport_TotalsAggregate(t *testing.T) {
	a1 := makeAccount("10010001")
	a2 := makeAccount("20010001")
	snap := buildSnap(uuid.New(), snapEntryFrom(a1), snapEntryFrom(a2))
	snapRepo := &p17StubSnapshotRepo{store: []*domain.COASnapshot{snap}}

	txn := &domain.Transaction{
		ID:                uuid.New(),
		TransactionStatus: domain.TransactionStatusPosted,
		Entries: []domain.TransactionEntry{
			debitEntry(a1.ID, 250),
			creditEntry(a2.ID, 250),
		},
	}
	svc := newDReportSvc(snapRepo, dreportTxnRepo(txn))
	req := service.TrialBalanceRequest{SnapshotID: snap.ID, GeneratedBy: uuid.New()}
	report, err := svc.GenerateTrialBalance(dreportCtx(), req)
	if err != nil {
		t.Fatalf("FIN-DREPORT-008: unexpected error: %v", err)
	}

	expected := decimal.NewFromFloat(250)
	if !report.TotalDebits.Equal(expected) {
		t.Errorf("FIN-DREPORT-008: expected TotalDebits=%s, got %s", expected, report.TotalDebits)
	}
	if !report.TotalCredits.Equal(expected) {
		t.Errorf("FIN-DREPORT-008: expected TotalCredits=%s, got %s", expected, report.TotalCredits)
	}
}

// FIN-DREPORT-009 --------------------------------------------------------

func TestDReport_VerifyHash_Reproducible(t *testing.T) {
	acct1 := makeAccount("10010001")
	snap := buildSnap(uuid.New(), snapEntryFrom(acct1))
	snapRepo := &p17StubSnapshotRepo{store: []*domain.COASnapshot{snap}}

	txn := postedTxn(debitEntry(acct1.ID, 100))
	svc := newDReportSvc(snapRepo, dreportTxnRepo(txn))

	req := service.TrialBalanceRequest{SnapshotID: snap.ID, GeneratedBy: uuid.New()}
	report, err := svc.GenerateTrialBalance(dreportCtx(), req)
	if err != nil {
		t.Fatalf("FIN-DREPORT-009: generate error: %v", err)
	}

	// Build a stored hash record that points to the same inputs.
	stored := &domain.FinancialReportHash{
		ID:         uuid.New(),
		SnapshotID: snap.ID,
		AsOfDate:   req.AsOfDate,
		ReportHash: report.Hash,
	}

	valid, err := svc.VerifyReportHash(dreportCtx(), stored, uuid.New())
	if err != nil {
		t.Fatalf("FIN-DREPORT-009: verify error: %v", err)
	}
	if !valid {
		t.Error("FIN-DREPORT-009: expected hash to be reproducible")
	}
}

// FIN-DREPORT-010 --------------------------------------------------------

func TestDReport_VerifyHash_LedgerMutated(t *testing.T) {
	acct1 := makeAccount("10010001")
	snap := buildSnap(uuid.New(), snapEntryFrom(acct1))

	// Original ledger.
	txnOrig := postedTxn(debitEntry(acct1.ID, 100))
	snapRepoOrig := &p17StubSnapshotRepo{store: []*domain.COASnapshot{snap}}
	svcOrig := newDReportSvc(snapRepoOrig, dreportTxnRepo(txnOrig))
	req := service.TrialBalanceRequest{SnapshotID: snap.ID, GeneratedBy: uuid.New()}
	origReport, err := svcOrig.GenerateTrialBalance(dreportCtx(), req)
	if err != nil {
		t.Fatalf("FIN-DREPORT-010: original generate error: %v", err)
	}

	// Now use a different (mutated) ledger for verification.
	txnMutated := postedTxn(debitEntry(acct1.ID, 999))
	snapRepoMutated := &p17StubSnapshotRepo{store: []*domain.COASnapshot{snap}}
	svcMutated := newDReportSvc(snapRepoMutated, dreportTxnRepo(txnMutated))

	stored := &domain.FinancialReportHash{
		ID:         uuid.New(),
		SnapshotID: snap.ID,
		AsOfDate:   req.AsOfDate,
		ReportHash: origReport.Hash,
	}

	valid, err := svcMutated.VerifyReportHash(dreportCtx(), stored, uuid.New())
	if err != nil {
		t.Fatalf("FIN-DREPORT-010: verify error: %v", err)
	}
	if valid {
		t.Error("FIN-DREPORT-010: expected hash mismatch for mutated ledger")
	}
}

// FIN-DREPORT-011 --------------------------------------------------------

func TestDReport_NoTenantInContext_ReturnsError(t *testing.T) {
	snap := buildSnap(uuid.New())
	snapRepo := &p17StubSnapshotRepo{store: []*domain.COASnapshot{snap}}
	svc := newDReportSvc(snapRepo, dreportTxnRepo())

	req := service.TrialBalanceRequest{SnapshotID: snap.ID, GeneratedBy: uuid.New()}
	_, err := svc.GenerateTrialBalance(context.Background(), req) // no tenant
	if err == nil {
		t.Error("FIN-DREPORT-011: expected error when tenant ID missing, got nil")
	}
}
