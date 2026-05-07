// Package service_test — integrity service unit tests.
//
// Proves that:
// - ScanPostedTransactions calls all registered checks per transaction.
// - Built-in balance check raises CRITICAL violation on unbalanced transactions.
// - Built-in no-entries check raises HIGH violation on POSTED transaction with no entries.
// - IntegrityReport.HasCritical() correctly gates period-close decisions.
// No database required — uses stub repository.
package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	gomock "go.uber.org/mock/gomock"

	"awo.so/internal/core/finance/domain"
	"awo.so/internal/core/finance/service"
	"awo.so/internal/shared/metrics"
)

// ============================================================================
// FIN-INT-001: ScanPostedTransactions calls registered checks per transaction
// ============================================================================

// TestScanPostedTransactions_CallsChecks verifies that each transaction
// returned by the repo is passed to every registered Check.Execute call.
func TestScanPostedTransactions_CallsChecks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	txnID := uuid.New()
	txn := &domain.Transaction{
		ID:                txnID,
		TransactionStatus: domain.TransactionStatusPosted,
		TenantID:          uuid.New(),
	}

	repo := &stubTxnRepo{
		fnList: func(_ context.Context, _ *domain.TransactionFilter) ([]*domain.Transaction, error) {
			// Return one page with one transaction, then empty to stop pagination.
			return []*domain.Transaction{txn}, nil
		},
		fnGetEntriesByTxn: func(_ context.Context, _ uuid.UUID) ([]domain.TransactionEntry, error) {
			return []domain.TransactionEntry{}, nil
		},
	}

	// Override fnList to return empty on second call (stop pagination).
	callCount := 0
	repo.fnList = func(_ context.Context, f *domain.TransactionFilter) ([]*domain.Transaction, error) {
		callCount++
		if callCount == 1 {
			return []*domain.Transaction{txn}, nil
		}
		return []*domain.Transaction{}, nil
	}

	mockCheck := service.NewMockCheck(ctrl)
	mockCheck.EXPECT().Kind().Return("TEST_CHECK").AnyTimes()
	// Execute must be called once — for the single transaction.
	mockCheck.EXPECT().
		Execute(gomock.Any(), txn, gomock.Any(), gomock.Any()).
		Times(1)

	intSvc := service.NewIntegrityServiceWithChecks(repo, nil, metrics.NewNoOpMetricsProvider(), mockCheck)
	_, err := intSvc.ScanPostedTransactions(context.Background(), nil)
	if err != nil {
		t.Fatalf("ScanPostedTransactions failed: %v", err)
	}
}

// ============================================================================
// FIN-INT-002: Built-in checks detect unbalanced transaction
// ============================================================================

// TestIntegrityService_UnbalancedTransaction verifies that scanning a POSTED
// transaction where debits ≠ credits produces a CRITICAL violation.
func TestIntegrityService_UnbalancedTransaction(t *testing.T) {
	txnID := uuid.New()
	txn := &domain.Transaction{
		ID:                txnID,
		TenantID:          uuid.New(),
		TransactionStatus: domain.TransactionStatusPosted,
		TransactionNumber: "TXN-0001",
	}
	entries := []domain.TransactionEntry{
		{DebitAmount: decimal.NewFromFloat(100), CreditAmount: decimal.Zero},
		{DebitAmount: decimal.Zero, CreditAmount: decimal.NewFromFloat(90)}, // off by 10
	}

	pageCount := 0
	repo := &stubTxnRepo{
		fnList: func(_ context.Context, _ *domain.TransactionFilter) ([]*domain.Transaction, error) {
			pageCount++
			if pageCount == 1 {
				return []*domain.Transaction{txn}, nil
			}
			return nil, nil
		},
		fnGetEntriesByTxn: func(_ context.Context, _ uuid.UUID) ([]domain.TransactionEntry, error) {
			return entries, nil
		},
	}

	// NewIntegrityService includes built-in checks.
	intSvc := service.NewIntegrityService(repo, nil, metrics.NewNoOpMetricsProvider())
	report, err := intSvc.ScanPostedTransactions(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !report.HasCritical() {
		t.Fatalf("expected CRITICAL violation for unbalanced transaction, violations: %+v", report.Violations)
	}

	found := false
	for _, v := range report.Violations {
		if v.Kind == "UNBALANCED_TRANSACTION" && v.Severity == service.SeverityCritical {
			found = true
		}
	}
	if !found {
		t.Errorf("expected UNBALANCED_TRANSACTION/CRITICAL violation, violations: %+v", report.Violations)
	}
}

// ============================================================================
// FIN-INT-003: Built-in checks detect POSTED transaction with no entries
// ============================================================================

func TestIntegrityService_PostedWithNoEntries(t *testing.T) {
	txn := &domain.Transaction{
		ID:                uuid.New(),
		TenantID:          uuid.New(),
		TransactionStatus: domain.TransactionStatusPosted,
		TransactionNumber: "TXN-0002",
	}

	pageCount := 0
	repo := &stubTxnRepo{
		fnList: func(_ context.Context, _ *domain.TransactionFilter) ([]*domain.Transaction, error) {
			pageCount++
			if pageCount == 1 {
				return []*domain.Transaction{txn}, nil
			}
			return nil, nil
		},
		fnGetEntriesByTxn: func(_ context.Context, _ uuid.UUID) ([]domain.TransactionEntry, error) {
			return []domain.TransactionEntry{}, nil // no entries
		},
	}

	intSvc := service.NewIntegrityService(repo, nil, metrics.NewNoOpMetricsProvider())
	report, err := intSvc.ScanPostedTransactions(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, v := range report.Violations {
		if v.Kind == "POSTED_WITHOUT_ENTRIES" && v.Severity == service.SeverityHigh {
			found = true
		}
	}
	if !found {
		t.Errorf("expected POSTED_WITHOUT_ENTRIES/HIGH violation, violations: %+v", report.Violations)
	}
}

// ============================================================================
// FIN-INT-004: Balanced transaction produces no violations
// ============================================================================

func TestIntegrityService_BalancedTransaction_NoViolations(t *testing.T) {
	txn := &domain.Transaction{
		ID:                uuid.New(),
		TenantID:          uuid.New(),
		TransactionStatus: domain.TransactionStatusPosted,
		TransactionNumber: "TXN-0003",
	}
	entries := []domain.TransactionEntry{
		{DebitAmount: decimal.NewFromFloat(250), CreditAmount: decimal.Zero},
		{DebitAmount: decimal.Zero, CreditAmount: decimal.NewFromFloat(250)},
	}

	pageCount := 0
	repo := &stubTxnRepo{
		fnList: func(_ context.Context, _ *domain.TransactionFilter) ([]*domain.Transaction, error) {
			pageCount++
			if pageCount == 1 {
				return []*domain.Transaction{txn}, nil
			}
			return nil, nil
		},
		fnGetEntriesByTxn: func(_ context.Context, _ uuid.UUID) ([]domain.TransactionEntry, error) {
			return entries, nil
		},
	}

	intSvc := service.NewIntegrityService(repo, nil, metrics.NewNoOpMetricsProvider())
	report, err := intSvc.ScanPostedTransactions(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.HasCritical() {
		t.Errorf("expected no CRITICAL violations for balanced transaction, got: %+v", report.Violations)
	}
	if len(report.Violations) > 0 {
		t.Errorf("expected zero violations for balanced transaction with entries, got: %+v", report.Violations)
	}
}

// ============================================================================
// FIN-INT-005: HasCritical gates correctly
// ============================================================================

func TestIntegrityReport_HasCritical(t *testing.T) {
	// Empty report: not critical.
	empty := &service.IntegrityReport{}
	if empty.HasCritical() {
		t.Error("empty report should not be critical")
	}

	// Report with only HIGH violation: not critical.
	highOnly := &service.IntegrityReport{
		Violations: []service.IntegrityViolation{
			{Kind: "POSTED_WITHOUT_ENTRIES", Severity: service.SeverityHigh},
		},
	}
	if highOnly.HasCritical() {
		t.Error("HIGH-only report should not be critical")
	}

	// Report with CRITICAL violation: is critical.
	withCritical := &service.IntegrityReport{
		Violations: []service.IntegrityViolation{
			{Kind: "UNBALANCED_TRANSACTION", Severity: service.SeverityCritical},
		},
	}
	if !withCritical.HasCritical() {
		t.Error("report with CRITICAL violation should return HasCritical=true")
	}
}

// ============================================================================
// FIN-INT-006: Custom check is called via NewIntegrityServiceWithChecks
// ============================================================================

// TestIntegrityService_CustomCheck proves that custom checks are executed
// and their violations appear in the report.
func TestIntegrityService_CustomCheck(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	txn := &domain.Transaction{
		ID:                uuid.New(),
		TenantID:          uuid.New(),
		TransactionStatus: domain.TransactionStatusPosted,
	}

	pageCount := 0
	repo := &stubTxnRepo{
		fnList: func(_ context.Context, _ *domain.TransactionFilter) ([]*domain.Transaction, error) {
			pageCount++
			if pageCount == 1 {
				return []*domain.Transaction{txn}, nil
			}
			return nil, nil
		},
		fnGetEntriesByTxn: func(_ context.Context, _ uuid.UUID) ([]domain.TransactionEntry, error) {
			return []domain.TransactionEntry{}, nil
		},
	}

	// Custom check that always adds a MEDIUM violation.
	mockCheck := service.NewMockCheck(ctrl)
	mockCheck.EXPECT().Kind().Return("CUSTOM_CHECK").AnyTimes()
	mockCheck.EXPECT().Execute(gomock.Any(), txn, gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, txn *domain.Transaction, _ []domain.TransactionEntry, report *service.IntegrityReport) {
			report.Violations = append(report.Violations, service.IntegrityViolation{
				Kind:     "CUSTOM_CHECK",
				Severity: service.SeverityMedium,
				EntityID: txn.ID,
				Detail:   "custom check triggered",
			})
		}).Times(1)

	intSvc := service.NewIntegrityServiceWithChecks(repo, nil, metrics.NewNoOpMetricsProvider(), mockCheck)
	report, err := intSvc.ScanPostedTransactions(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, v := range report.Violations {
		if v.Kind == "CUSTOM_CHECK" {
			found = true
		}
	}
	if !found {
		t.Errorf("custom check violation not found in report: %+v", report.Violations)
	}
}
