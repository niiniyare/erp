// Package service_test — COA integrity service tests.
//
// Tests (FIN-INTEGRITY-*):
//
//	001 — Orphan account detected (parent doesn't exist)
//	002 — Root-type mismatch between parent and child
//	003 — Depth exceeded (level > MaxAccountHierarchyDepth)
//	004 — Dead account flagged (ACTIVE, no transactions, old enough)
//	005 — Balance in CLOSED account is CRITICAL
//	006 — Balance in ARCHIVED account is CRITICAL
//	007 — Normal-balance mismatch detected
//	008 — Duplicate account code detected
//	009 — DRAFT account with non-zero balance flagged
//	010 — Cyclic hierarchy detected
//	011 — Healthy COA produces empty violations list
//	012 — Nil service receiver — ScanTenant nil-safe
//	013 — CheckSingleAccount returns violations for single bad account
//	014 — ScanTenant propagates repo error
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

// =============================================================================
// Stub that controls List output
// =============================================================================

type integrityAccountRepo struct {
	p16StubAccountRepo
	accounts    []*domain.Accounts
	getByIDRepo map[uuid.UUID]*domain.Accounts
	listErr     error
}

func newIntegrityRepo(accounts ...*domain.Accounts) *integrityAccountRepo {
	r := &integrityAccountRepo{
		getByIDRepo: make(map[uuid.UUID]*domain.Accounts),
	}
	for _, a := range accounts {
		r.accounts = append(r.accounts, a)
		r.getByIDRepo[a.ID] = a
	}
	return r
}

func (r *integrityAccountRepo) List(_ context.Context, _ *domain.AccountFilter) ([]*domain.Accounts, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	return r.accounts, nil
}

func (r *integrityAccountRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Accounts, error) {
	if a, ok := r.getByIDRepo[id]; ok {
		return a, nil
	}
	return nil, errors.New("not found")
}

func intCtx() context.Context {
	return shared.WithTenantID(context.Background(), uuid.New())
}

func newIntegritySvc(repo *integrityAccountRepo) *service.COAIntegrityService {
	return service.NewCOAIntegrityService(repo, nil, metrics.NewNoOpMetricsProvider(), 0)
}

func baseAccount(code string, status domain.AccountStatus) *domain.Accounts {
	curr := "USD"
	return &domain.Accounts{
		ID:             uuid.New(),
		TenantID:       uuid.New(),
		AccountCode:    code,
		AccountName:    "Test " + code,
		RootType:       domain.RootTypeAsset,
		NormalBalance:  domain.NormalBalanceDebit,
		AccountType:    "Current Asset",
		Status:         status,
		CurrentBalance: decimal.Zero,
		CurrencyCode:   &curr,
		CreatedAt:      time.Now().Add(-2 * 365 * 24 * time.Hour), // 2 years old
		UpdatedAt:      time.Now(),
	}
}

func containsViolationKind(violations []domain.COAIntegrityViolation, kind domain.COAViolationKind) bool {
	for _, v := range violations {
		if v.Kind == kind {
			return true
		}
	}
	return false
}

// =============================================================================
// FIN-INTEGRITY-001: Orphan account detected
// =============================================================================

func TestIntegrity_OrphanAccount_Detected(t *testing.T) {
	orphan := baseAccount("10010001", domain.AccountStatusActive)
	missingParent := uuid.New()
	orphan.ParentAccountID = &missingParent // parent doesn't exist

	repo := newIntegrityRepo(orphan)
	svc := newIntegritySvc(repo)

	report, err := svc.ScanTenant(intCtx())
	if err != nil {
		t.Fatalf("FIN-INTEGRITY-001: unexpected error: %v", err)
	}
	if !containsViolationKind(report.Violations, domain.COAViolationOrphanAccount) {
		t.Error("FIN-INTEGRITY-001: expected ORPHAN_ACCOUNT violation")
	}
}

// =============================================================================
// FIN-INTEGRITY-002: Root-type mismatch
// =============================================================================

func TestIntegrity_RootTypeMismatch_Detected(t *testing.T) {
	parent := baseAccount("10000000", domain.AccountStatusActive)
	parent.RootType = domain.RootTypeAsset

	child := baseAccount("10010001", domain.AccountStatusActive)
	child.RootType = domain.RootTypeExpense // mismatch
	child.ParentAccountID = &parent.ID
	child.AccountLevel = 1

	repo := newIntegrityRepo(parent, child)
	svc := newIntegritySvc(repo)

	report, err := svc.ScanTenant(intCtx())
	if err != nil {
		t.Fatalf("FIN-INTEGRITY-002: unexpected error: %v", err)
	}
	if !containsViolationKind(report.Violations, domain.COAViolationRootTypeMismatch) {
		t.Error("FIN-INTEGRITY-002: expected ROOT_TYPE_MISMATCH violation")
	}
}

// =============================================================================
// FIN-INTEGRITY-003: Depth exceeded
// =============================================================================

func TestIntegrity_DepthExceeded_Detected(t *testing.T) {
	deep := baseAccount("10010001", domain.AccountStatusActive)
	deep.AccountLevel = int32(domain.MaxAccountHierarchyDepth + 1)

	repo := newIntegrityRepo(deep)
	svc := newIntegritySvc(repo)

	report, err := svc.ScanTenant(intCtx())
	if err != nil {
		t.Fatalf("FIN-INTEGRITY-003: unexpected error: %v", err)
	}
	if !containsViolationKind(report.Violations, domain.COAViolationDepthExceeded) {
		t.Errorf("FIN-INTEGRITY-003: expected DEPTH_EXCEEDED violation (level=%d, max=%d)",
			deep.AccountLevel, domain.MaxAccountHierarchyDepth)
	}
}

// =============================================================================
// FIN-INTEGRITY-004: Dead account flagged
// =============================================================================

func TestIntegrity_DeadAccount_Flagged(t *testing.T) {
	dead := baseAccount("10010001", domain.AccountStatusActive)
	dead.CreatedAt = time.Now().Add(-2 * 365 * 24 * time.Hour) // 2 years old
	dead.LastTransactionDate = nil                             // never transacted
	// Default dead threshold = 1 year, so this account should be flagged.

	repo := newIntegrityRepo(dead)
	svc := newIntegritySvc(repo)

	report, err := svc.ScanTenant(intCtx())
	if err != nil {
		t.Fatalf("FIN-INTEGRITY-004: unexpected error: %v", err)
	}
	if !containsViolationKind(report.Violations, domain.COAViolationDeadAccount) {
		t.Error("FIN-INTEGRITY-004: expected DEAD_ACCOUNT violation")
	}
}

// =============================================================================
// FIN-INTEGRITY-005: Balance in CLOSED account is CRITICAL
// =============================================================================

func TestIntegrity_BalanceInClosed_Critical(t *testing.T) {
	closed := baseAccount("10010001", domain.AccountStatusClosed)
	closed.CurrentBalance = decimal.NewFromFloat(100.00)

	repo := newIntegrityRepo(closed)
	svc := newIntegritySvc(repo)

	report, err := svc.ScanTenant(intCtx())
	if err != nil {
		t.Fatalf("FIN-INTEGRITY-005: unexpected error: %v", err)
	}

	found := false
	for _, v := range report.Violations {
		if v.Kind == domain.COAViolationBalanceInClosedAccount && v.Severity == domain.COASeverityCritical {
			found = true
		}
	}
	if !found {
		t.Error("FIN-INTEGRITY-005: expected CRITICAL BALANCE_IN_CLOSED_ACCOUNT violation")
	}
	if report.Healthy {
		t.Error("FIN-INTEGRITY-005: report should not be healthy with CRITICAL violation")
	}
}

// =============================================================================
// FIN-INTEGRITY-006: Balance in ARCHIVED account is CRITICAL
// =============================================================================

func TestIntegrity_BalanceInArchived_Critical(t *testing.T) {
	archived := baseAccount("10010001", domain.AccountStatusArchived)
	archived.CurrentBalance = decimal.NewFromFloat(50.00)

	repo := newIntegrityRepo(archived)
	svc := newIntegritySvc(repo)

	report, err := svc.ScanTenant(intCtx())
	if err != nil {
		t.Fatalf("FIN-INTEGRITY-006: unexpected error: %v", err)
	}
	if !containsViolationKind(report.Violations, domain.COAViolationBalanceInClosedAccount) {
		t.Error("FIN-INTEGRITY-006: expected BALANCE_IN_CLOSED_ACCOUNT violation for archived account")
	}
}

// =============================================================================
// FIN-INTEGRITY-007: Normal-balance mismatch
// =============================================================================

func TestIntegrity_NormalBalanceMismatch_Detected(t *testing.T) {
	account := baseAccount("10010001", domain.AccountStatusActive)
	account.RootType = domain.RootTypeAsset
	account.NormalBalance = domain.NormalBalanceCredit // wrong — Asset should be DEBIT

	repo := newIntegrityRepo(account)
	svc := newIntegritySvc(repo)

	report, err := svc.ScanTenant(intCtx())
	if err != nil {
		t.Fatalf("FIN-INTEGRITY-007: unexpected error: %v", err)
	}
	if !containsViolationKind(report.Violations, domain.COAViolationNormalBalanceMismatch) {
		t.Error("FIN-INTEGRITY-007: expected NORMAL_BALANCE_MISMATCH violation")
	}
}

// =============================================================================
// FIN-INTEGRITY-008: Duplicate code detected
// =============================================================================

func TestIntegrity_DuplicateCode_Detected(t *testing.T) {
	a1 := baseAccount("10010001", domain.AccountStatusActive)
	a2 := baseAccount("10010001", domain.AccountStatusActive) // same code

	repo := newIntegrityRepo(a1, a2)
	svc := newIntegritySvc(repo)

	report, err := svc.ScanTenant(intCtx())
	if err != nil {
		t.Fatalf("FIN-INTEGRITY-008: unexpected error: %v", err)
	}
	if !containsViolationKind(report.Violations, domain.COAViolationDuplicateCode) {
		t.Error("FIN-INTEGRITY-008: expected DUPLICATE_CODE violation")
	}
}

// =============================================================================
// FIN-INTEGRITY-009: DRAFT with non-zero balance
// =============================================================================

func TestIntegrity_DraftWithBalance_Detected(t *testing.T) {
	draft := baseAccount("10010001", domain.AccountStatusDraft)
	draft.CurrentBalance = decimal.NewFromFloat(200.00)

	repo := newIntegrityRepo(draft)
	svc := newIntegritySvc(repo)

	report, err := svc.ScanTenant(intCtx())
	if err != nil {
		t.Fatalf("FIN-INTEGRITY-009: unexpected error: %v", err)
	}
	if !containsViolationKind(report.Violations, domain.COAViolationInvalidStatusForBalance) {
		t.Error("FIN-INTEGRITY-009: expected INVALID_STATUS_FOR_BALANCE violation")
	}
}

// =============================================================================
// FIN-INTEGRITY-010: Cyclic hierarchy detected
// =============================================================================

func TestIntegrity_CyclicHierarchy_Detected(t *testing.T) {
	a := baseAccount("10000000", domain.AccountStatusActive)
	b := baseAccount("10010000", domain.AccountStatusActive)

	// a's parent = b, b's parent = a → cycle
	a.ParentAccountID = &b.ID
	a.AccountLevel = 1
	b.ParentAccountID = &a.ID
	b.AccountLevel = 1

	repo := newIntegrityRepo(a, b)
	svc := newIntegritySvc(repo)

	report, err := svc.ScanTenant(intCtx())
	if err != nil {
		t.Fatalf("FIN-INTEGRITY-010: unexpected error: %v", err)
	}
	if !containsViolationKind(report.Violations, domain.COAViolationCyclicHierarchy) {
		t.Error("FIN-INTEGRITY-010: expected CYCLIC_HIERARCHY violation")
	}
}

// =============================================================================
// FIN-INTEGRITY-011: Healthy COA — no violations
// =============================================================================

func TestIntegrity_HealthyCOA_NoViolations(t *testing.T) {
	account := baseAccount("10010001", domain.AccountStatusActive)
	account.CreatedAt = time.Now().Add(-7 * 24 * time.Hour) // new account — not dead
	now := time.Now()
	account.LastTransactionDate = &now // recently transacted

	repo := newIntegrityRepo(account)
	svc := newIntegritySvc(repo)

	report, err := svc.ScanTenant(intCtx())
	if err != nil {
		t.Fatalf("FIN-INTEGRITY-011: unexpected error: %v", err)
	}
	if len(report.Violations) > 0 {
		t.Errorf("FIN-INTEGRITY-011: expected no violations for healthy COA, got %d: %+v",
			len(report.Violations), report.Violations)
	}
	if !report.Healthy {
		t.Error("FIN-INTEGRITY-011: expected Healthy=true for clean COA")
	}
}

// =============================================================================
// FIN-INTEGRITY-012: Nil service — nil-safe
// =============================================================================

func TestIntegrity_NilService_Safe(t *testing.T) {
	var svc *service.COAIntegrityService

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("FIN-INTEGRITY-012: nil service panicked: %v", r)
		}
	}()

	report, err := svc.ScanTenant(intCtx())
	if err != nil {
		t.Errorf("FIN-INTEGRITY-012: nil service should return nil error, got %v", err)
	}
	if report == nil {
		t.Error("FIN-INTEGRITY-012: nil service should return non-nil report")
	}
}

// =============================================================================
// FIN-INTEGRITY-013: CheckSingleAccount returns violations
// =============================================================================

func TestIntegrity_CheckSingleAccount_ReturnsViolations(t *testing.T) {
	account := baseAccount("10010001", domain.AccountStatusClosed)
	account.CurrentBalance = decimal.NewFromFloat(500.00) // balance in closed account

	repo := newIntegrityRepo(account)
	svc := newIntegritySvc(repo)

	violations, err := svc.CheckSingleAccount(intCtx(), account.ID)
	if err != nil {
		t.Fatalf("FIN-INTEGRITY-013: unexpected error: %v", err)
	}
	if !containsViolationKind(violations, domain.COAViolationBalanceInClosedAccount) {
		t.Error("FIN-INTEGRITY-013: expected BALANCE_IN_CLOSED_ACCOUNT violation from CheckSingleAccount")
	}
}

// =============================================================================
// FIN-INTEGRITY-014: ScanTenant propagates repo error
// =============================================================================

func TestIntegrity_ScanTenant_PropagatesRepoError(t *testing.T) {
	repo := newIntegrityRepo()
	repo.listErr = errors.New("DB unavailable")
	svc := newIntegritySvc(repo)

	_, err := svc.ScanTenant(intCtx())
	if err == nil {
		t.Error("FIN-INTEGRITY-014: expected error propagation from repo, got nil")
	}
}
