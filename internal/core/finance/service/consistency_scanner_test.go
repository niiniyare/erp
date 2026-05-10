// Package service_test — FinancialConsistencyScanner tests.
//
// Tests (FIN-SCAN-*):
//
//	001 — Empty system produces healthy report with no violations
//	002 — Nil service returns healthy report (nil-safe)
//	003 — Orphan posting detected when entry references unknown account
//	004 — Balance drift detected for closed account with non-zero balance and ledger activity
//	005 — No snapshot for tenant flags COA_MUTATION_WITHOUT_SNAPSHOT (HIGH)
//	006 — HasCritical returns true when critical violation present
//	007 — Healthy=false when critical violation present
//	008 — Healthy=false when HIGH violation present
//	009 — Clean system with snapshot returns healthy=true
//	010 — ScanSystem propagates accounts list error
//	011 — ScanSystem without tenant ID returns error
//	012 — LedgerTxnCount and LedgerEntryCount populated correctly
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

// helper — consistency scanner context with tenant
func scanCtx() context.Context {
	return shared.WithTenantID(context.Background(), uuid.New())
}

// helper — consistency scanner with given deps
func newScannerSvc(
	accounts *p17AccountRepo,
	txns *p16StubTxnRepo,
	snapshots *p17StubSnapshotRepo,
) *service.FinancialConsistencyScanner {
	return service.NewFinancialConsistencyScanner(accounts, txns, snapshots, metrics.NewNoOpMetricsProvider())
}

// helper — active account with a given balance
func activeAccountWithBalance(balance float64) *domain.Accounts {
	a := makeAccount("10010001")
	a.CurrentBalance = decimal.NewFromFloat(balance)
	return a
}

// helper — closed account with non-zero balance
func closedAccountWithBalance(balance float64) *domain.Accounts {
	a := makeAccount("10010001")
	a.Status = domain.AccountStatusClosed
	a.CurrentBalance = decimal.NewFromFloat(balance)
	return a
}

// FIN-SCAN-001 -----------------------------------------------------------

func TestScan_EmptySystem_Healthy(t *testing.T) {
	accounts := newP17AccountRepo()
	txns := txnRepoForReplay()
	snapshots := &p17StubSnapshotRepo{}

	svc := newScannerSvc(accounts, txns, snapshots)
	report, err := svc.ScanSystem(scanCtx(), time.Now())
	if err != nil {
		t.Fatalf("FIN-SCAN-001: unexpected error: %v", err)
	}
	if !report.Healthy {
		t.Errorf("FIN-SCAN-001: expected healthy for empty system, violations: %+v", report.Violations)
	}
}

// FIN-SCAN-002 -----------------------------------------------------------

func TestScan_NilService_Safe(t *testing.T) {
	var svc *service.FinancialConsistencyScanner

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("FIN-SCAN-002: nil service panicked: %v", r)
		}
	}()

	report, err := svc.ScanSystem(scanCtx(), time.Time{})
	if err != nil {
		t.Errorf("FIN-SCAN-002: nil service should return nil error, got %v", err)
	}
	if report == nil {
		t.Error("FIN-SCAN-002: expected non-nil report from nil service")
	}
}

// FIN-SCAN-003 -----------------------------------------------------------

func TestScan_OrphanPosting_Detected(t *testing.T) {
	// No accounts in COA, but a txn entry references one.
	unknownAcctID := uuid.New()
	txn := postedTxn(debitEntry(unknownAcctID, 100))

	accounts := newP17AccountRepo() // empty COA
	txns := txnRepoForReplay(txn)
	snapshots := &p17StubSnapshotRepo{}

	svc := newScannerSvc(accounts, txns, snapshots)
	report, err := svc.ScanSystem(scanCtx(), time.Now())
	if err != nil {
		t.Fatalf("FIN-SCAN-003: unexpected error: %v", err)
	}

	found := false
	for _, v := range report.Violations {
		if v.Kind == domain.ConsistencyOrphanPosting {
			found = true
		}
	}
	if !found {
		t.Error("FIN-SCAN-003: expected ORPHAN_POSTING violation")
	}
}

// FIN-SCAN-004 -----------------------------------------------------------

func TestScan_BalanceDrift_ClosedAccountWithLedgerActivity(t *testing.T) {
	acctID := uuid.New()
	acct := makeAccount("10010001")
	acct.ID = acctID
	acct.Status = domain.AccountStatusClosed
	acct.CurrentBalance = decimal.NewFromFloat(100)
	acct.IsActive = false

	// Txn entry references this closed account.
	txn := postedTxn(debitEntry(acctID, 100))
	txn.Entries = append(txn.Entries, debitEntry(acctID, 100)) // put in txn.Entries so inline check works

	accounts := newP17AccountRepo(acct)
	txns := txnRepoForReplay(txn)
	snapshots := &p17StubSnapshotRepo{}

	svc := newScannerSvc(accounts, txns, snapshots)
	report, err := svc.ScanSystem(scanCtx(), time.Now())
	if err != nil {
		t.Fatalf("FIN-SCAN-004: unexpected error: %v", err)
	}

	found := false
	for _, v := range report.Violations {
		if v.Kind == domain.ConsistencyBalanceDrift {
			found = true
		}
	}
	if !found {
		t.Error("FIN-SCAN-004: expected BALANCE_DRIFT violation for closed account with ledger activity")
	}
}

// FIN-SCAN-005 -----------------------------------------------------------

func TestScan_NoSnapshot_FlagsMutationWithoutSnapshot(t *testing.T) {
	// Has accounts but NO snapshot.
	accounts := newP17AccountRepo(makeAccount("10010001"))
	txns := txnRepoForReplay()
	snapshots := &p17StubSnapshotRepo{} // empty store

	svc := newScannerSvc(accounts, txns, snapshots)
	report, err := svc.ScanSystem(scanCtx(), time.Now())
	if err != nil {
		t.Fatalf("FIN-SCAN-005: unexpected error: %v", err)
	}

	found := false
	for _, v := range report.Violations {
		if v.Kind == domain.ConsistencyCOAMutationWithoutSnapshot {
			found = true
		}
	}
	if !found {
		t.Error("FIN-SCAN-005: expected COA_MUTATION_WITHOUT_SNAPSHOT violation when no snapshot exists")
	}
}

// FIN-SCAN-006 -----------------------------------------------------------

func TestScan_HasCritical_True(t *testing.T) {
	report := &domain.SystemConsistencyReport{
		Violations: []domain.ConsistencyViolation{
			{Kind: domain.ConsistencyOrphanPosting, Severity: domain.COASeverityCritical},
		},
	}
	if !report.HasCritical() {
		t.Error("FIN-SCAN-006: expected HasCritical=true")
	}
}

// FIN-SCAN-007 -----------------------------------------------------------

func TestScan_Healthy_False_WhenCriticalViolation(t *testing.T) {
	unknownID := uuid.New()
	txn := postedTxn(debitEntry(unknownID, 100))

	accounts := newP17AccountRepo()
	txns := txnRepoForReplay(txn)
	snapshots := &p17StubSnapshotRepo{}

	svc := newScannerSvc(accounts, txns, snapshots)
	report, err := svc.ScanSystem(scanCtx(), time.Now())
	if err != nil {
		t.Fatalf("FIN-SCAN-007: unexpected error: %v", err)
	}
	if report.Healthy {
		t.Error("FIN-SCAN-007: expected Healthy=false when CRITICAL violation exists")
	}
}

// FIN-SCAN-008 -----------------------------------------------------------

func TestScan_Healthy_False_WhenHighViolation(t *testing.T) {
	// No snapshot → HIGH violation.
	accounts := newP17AccountRepo(makeAccount("10010001"))
	txns := txnRepoForReplay()
	snapshots := &p17StubSnapshotRepo{}

	svc := newScannerSvc(accounts, txns, snapshots)
	report, err := svc.ScanSystem(scanCtx(), time.Now())
	if err != nil {
		t.Fatalf("FIN-SCAN-008: unexpected error: %v", err)
	}
	if report.Healthy {
		t.Error("FIN-SCAN-008: expected Healthy=false when HIGH violation exists")
	}
}

// FIN-SCAN-009 -----------------------------------------------------------

func TestScan_CleanSystemWithSnapshot_Healthy(t *testing.T) {
	tenantID := uuid.New()
	acct := makeAccount("10010001")
	acct.TenantID = tenantID
	acct.CurrentBalance = decimal.Zero

	// Provide a snapshot so coverage check passes.
	snap := &domain.COASnapshot{
		ID:         uuid.New(),
		TenantID:   tenantID,
		SnapshotAt: time.Now().Add(-1 * time.Hour),
		Hash:       "testhash",
		CreatedBy:  uuid.New(),
		CreatedAt:  time.Now(),
	}
	snapshots := &p17StubSnapshotRepo{store: []*domain.COASnapshot{snap}}
	accounts := newP17AccountRepo(acct)
	txns := txnRepoForReplay() // no transactions → no orphan or drift violations

	ctx := shared.WithTenantID(context.Background(), tenantID)
	svc := newScannerSvc(accounts, txns, snapshots)
	report, err := svc.ScanSystem(ctx, time.Now())
	if err != nil {
		t.Fatalf("FIN-SCAN-009: unexpected error: %v", err)
	}
	if !report.Healthy {
		t.Errorf("FIN-SCAN-009: expected healthy system, violations: %+v", report.Violations)
	}
}

// FIN-SCAN-010 -----------------------------------------------------------

func TestScan_PropagatesAccountListError(t *testing.T) {
	accounts := &p17AccountRepo{listErr: errors.New("accounts DB unavailable")}
	txns := txnRepoForReplay()
	snapshots := &p17StubSnapshotRepo{}

	svc := newScannerSvc(accounts, txns, snapshots)
	_, err := svc.ScanSystem(scanCtx(), time.Time{})
	if err == nil {
		t.Error("FIN-SCAN-010: expected error from account repo, got nil")
	}
}

// FIN-SCAN-011 -----------------------------------------------------------

func TestScan_NoTenantInContext_ReturnsError(t *testing.T) {
	accounts := newP17AccountRepo()
	txns := txnRepoForReplay()
	snapshots := &p17StubSnapshotRepo{}

	svc := newScannerSvc(accounts, txns, snapshots)
	_, err := svc.ScanSystem(context.Background(), time.Time{}) // no tenant
	if err == nil {
		t.Error("FIN-SCAN-011: expected error when tenant ID missing, got nil")
	}
}

// FIN-SCAN-012 -----------------------------------------------------------

func TestScan_TxnAndEntryCountsPopulated(t *testing.T) {
	tenantID := uuid.New()
	acctID := uuid.New()
	acct := makeAccount("10010001")
	acct.ID = acctID
	acct.TenantID = tenantID

	t1 := postedTxn(debitEntry(acctID, 100), creditEntry(acctID, 50))
	t2 := postedTxn(debitEntry(acctID, 200))

	snap := &domain.COASnapshot{
		ID:         uuid.New(),
		TenantID:   tenantID,
		SnapshotAt: time.Now().Add(-1 * time.Hour),
		Hash:       "h",
		CreatedBy:  uuid.New(),
		CreatedAt:  time.Now(),
	}
	snapshots := &p17StubSnapshotRepo{store: []*domain.COASnapshot{snap}}
	accounts := newP17AccountRepo(acct)
	txns := txnRepoForReplay(t1, t2)

	ctx := shared.WithTenantID(context.Background(), tenantID)
	svc := newScannerSvc(accounts, txns, snapshots)
	report, err := svc.ScanSystem(ctx, time.Now())
	if err != nil {
		t.Fatalf("FIN-SCAN-012: unexpected error: %v", err)
	}
	if report.LedgerTxnCount != 2 {
		t.Errorf("FIN-SCAN-012: expected LedgerTxnCount=2, got %d", report.LedgerTxnCount)
	}
	if report.LedgerEntryCount != 3 {
		t.Errorf("FIN-SCAN-012: expected LedgerEntryCount=3, got %d", report.LedgerEntryCount)
	}
}
