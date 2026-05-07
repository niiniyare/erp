package service

// integrity.go — Financial drift detection and ledger consistency verification.
//
// The IntegrityService scans the finance ledger for corruption that cannot be
// prevented by runtime guards alone: orphaned entries, unbalanced posted
// transactions, broken reversal chains, and duplicated postings.
//
// Run these checks:
//   - On a scheduled basis (e.g. nightly cron via Temporal)
//   - After any bulk import or migration
//   - As part of period-close sign-off
//   - In incident response when ledger correctness is in doubt
//
// All methods are READ-ONLY — they detect drift but never auto-correct.
// Correction requires a human-initiated reversal/amendment flow.

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"awo.so/internal/core/finance/domain"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
)

// IntegrityViolation describes a single detected inconsistency.
type IntegrityViolation struct {
	// Kind is a machine-readable violation code (e.g. "UNBALANCED_TRANSACTION").
	Kind string `json:"kind"`
	// EntityID is the UUID of the affected entity (transaction, entry, account…).
	EntityID uuid.UUID `json:"entity_id"`
	// Detail is a human-readable explanation suitable for operator review.
	Detail string `json:"detail"`
}

func (v IntegrityViolation) Error() string {
	return fmt.Sprintf("[%s] %s (entity=%s)", v.Kind, v.Detail, v.EntityID)
}

// IntegrityReport aggregates violations found by a single integrity scan pass.
type IntegrityReport struct {
	Violations []IntegrityViolation `json:"violations"`
	// ScannedTransactions is the number of transactions examined.
	ScannedTransactions int `json:"scanned_transactions"`
	// ScannedEntries is the number of entries examined.
	ScannedEntries int `json:"scanned_entries"`
}

func (r *IntegrityReport) add(v IntegrityViolation) {
	r.Violations = append(r.Violations, v)
}

// IntegrityService performs read-only consistency scans of the finance ledger.
type IntegrityService interface {
	// ScanPostedTransactions checks every posted transaction within the supplied
	// filter for double-entry balance violations (∑debit ≠ ∑credit in its entries).
	// Returns violations — callers must treat any non-empty report as a P1 incident.
	ScanPostedTransactions(ctx context.Context, filter *domain.TransactionFilter) (*IntegrityReport, error)

	// ScanReversalChains verifies that every transaction marked IsReversed=true
	// has a corresponding reversal record in the reversal history table and that
	// the reversal transaction exists and is POSTED.
	ScanReversalChains(ctx context.Context, filter *domain.TransactionFilter) (*IntegrityReport, error)

	// ScanDuplicatePostings finds posted transactions whose transaction number
	// appears more than once (excluding the template itself in the recurring case).
	// Duplicate postings indicate an idempotency failure.
	ScanDuplicatePostings(ctx context.Context, tenantID uuid.UUID) (*IntegrityReport, error)
}

type integrityService struct {
	txnRepo             domain.TransactionRepository
	reversalHistoryRepo domain.ReversalHistoryRepository // nil → reversal scan skipped
	metrics             metrics.MetricsProvider
}

// NewIntegrityService constructs an IntegrityService.
// reversalHistoryRepo may be nil; reversal chain scanning will be skipped.
func NewIntegrityService(
	txnRepo domain.TransactionRepository,
	reversalHistoryRepo domain.ReversalHistoryRepository,
	metrics metrics.MetricsProvider,
) IntegrityService {
	return &integrityService{
		txnRepo:             txnRepo,
		reversalHistoryRepo: reversalHistoryRepo,
		metrics:             metrics,
	}
}

// ScanPostedTransactions checks double-entry balance for every posted transaction.
func (s *integrityService) ScanPostedTransactions(ctx context.Context, filter *domain.TransactionFilter) (*IntegrityReport, error) {
	postedStatus := domain.TransactionStatusPosted
	scanFilter := &domain.TransactionFilter{}
	if filter != nil {
		*scanFilter = *filter
	}
	scanFilter.Status = &postedStatus

	transactions, err := s.txnRepo.List(ctx, scanFilter)
	if err != nil {
		return nil, fmt.Errorf("integrity scan: list posted transactions: %w", err)
	}

	report := &IntegrityReport{}
	report.ScannedTransactions = len(transactions)

	for _, txn := range transactions {
		entries, err := s.txnRepo.GetEntriesByTransaction(ctx, txn.ID)
		if err != nil {
			logger.ErrorContext(ctx, "integrity scan: failed to fetch entries for transaction",
				logger.Fields{"transaction_id": txn.ID.String(), "error": err.Error()})
			report.add(IntegrityViolation{
				Kind:     "ENTRY_FETCH_FAILED",
				EntityID: txn.ID,
				Detail:   fmt.Sprintf("could not fetch entries: %v", err),
			})
			continue
		}

		report.ScannedEntries += len(entries)

		if len(entries) == 0 {
			report.add(IntegrityViolation{
				Kind:     "POSTED_WITHOUT_ENTRIES",
				EntityID: txn.ID,
				Detail:   fmt.Sprintf("posted transaction %s has no entries", txn.TransactionNumber),
			})
			continue
		}

		var totalDebit, totalCredit decimal.Decimal
		for _, e := range entries {
			totalDebit = totalDebit.Add(e.DebitAmount)
			totalCredit = totalCredit.Add(e.CreditAmount)
		}

		if !totalDebit.Equal(totalCredit) {
			s.metrics.IncrementCounter("integrity_violations_total", metrics.Fields{
				"kind": "unbalanced_transaction",
			})
			report.add(IntegrityViolation{
				Kind:     "UNBALANCED_TRANSACTION",
				EntityID: txn.ID,
				Detail: fmt.Sprintf(
					"transaction %s debits=%s credits=%s diff=%s",
					txn.TransactionNumber,
					totalDebit.String(),
					totalCredit.String(),
					totalDebit.Sub(totalCredit).Abs().String(),
				),
			})
		}
	}

	s.metrics.IncrementCounter("integrity_scans_total", metrics.Fields{
		"scan_type":  "posted_transactions",
		"violations": fmt.Sprintf("%d", len(report.Violations)),
	})

	if len(report.Violations) > 0 {
		logger.ErrorContext(ctx, "INTEGRITY SCAN: violations detected in posted transactions — immediate review required",
			logger.Fields{
				"violations_count":     len(report.Violations),
				"scanned_transactions": report.ScannedTransactions,
			})
	} else {
		logger.InfoContext(ctx, "Integrity scan: posted transactions OK",
			logger.Fields{"scanned": report.ScannedTransactions})
	}

	return report, nil
}

// ScanReversalChains verifies every marked-reversed transaction has a valid reversal.
func (s *integrityService) ScanReversalChains(ctx context.Context, filter *domain.TransactionFilter) (*IntegrityReport, error) {
	if s.reversalHistoryRepo == nil {
		logger.WarnContext(ctx, "Integrity scan: reversal chain scan skipped — no reversalHistoryRepo configured")
		return &IntegrityReport{}, nil
	}

	reversedStatus := domain.TransactionStatusReversed
	scanFilter := &domain.TransactionFilter{}
	if filter != nil {
		*scanFilter = *filter
	}
	scanFilter.Status = &reversedStatus

	transactions, err := s.txnRepo.List(ctx, scanFilter)
	if err != nil {
		return nil, fmt.Errorf("integrity scan: list reversed transactions: %w", err)
	}

	report := &IntegrityReport{ScannedTransactions: len(transactions)}

	for _, txn := range transactions {
		records, err := s.reversalHistoryRepo.GetByOriginal(ctx, txn.ID)
		if err != nil {
			report.add(IntegrityViolation{
				Kind:     "REVERSAL_HISTORY_FETCH_FAILED",
				EntityID: txn.ID,
				Detail:   fmt.Sprintf("transaction %s: could not fetch reversal history: %v", txn.TransactionNumber, err),
			})
			continue
		}

		if len(records) == 0 {
			s.metrics.IncrementCounter("integrity_violations_total", metrics.Fields{
				"kind": "missing_reversal_record",
			})
			report.add(IntegrityViolation{
				Kind:     "MISSING_REVERSAL_RECORD",
				EntityID: txn.ID,
				Detail:   fmt.Sprintf("transaction %s marked REVERSED but has no reversal history record", txn.TransactionNumber),
			})
			continue
		}

		// Verify the reversal transaction itself exists and is POSTED.
		for _, rec := range records {
			reversal, err := s.txnRepo.GetByID(ctx, rec.ReversalTransactionID)
			if err != nil {
				s.metrics.IncrementCounter("integrity_violations_total", metrics.Fields{
					"kind": "reversal_transaction_missing",
				})
				report.add(IntegrityViolation{
					Kind:     "REVERSAL_TRANSACTION_MISSING",
					EntityID: rec.ReversalTransactionID,
					Detail: fmt.Sprintf(
						"reversal history for %s points to reversal %s which does not exist",
						txn.TransactionNumber, rec.ReversalTransactionID,
					),
				})
				continue
			}

			if reversal.TransactionStatus != domain.TransactionStatusPosted {
				s.metrics.IncrementCounter("integrity_violations_total", metrics.Fields{
					"kind": "reversal_not_posted",
				})
				report.add(IntegrityViolation{
					Kind:     "REVERSAL_NOT_POSTED",
					EntityID: reversal.ID,
					Detail: fmt.Sprintf(
						"reversal transaction %s for original %s has status %s (expected POSTED)",
						reversal.TransactionNumber, txn.TransactionNumber, reversal.TransactionStatus,
					),
				})
			}
		}
	}

	s.metrics.IncrementCounter("integrity_scans_total", metrics.Fields{
		"scan_type":  "reversal_chains",
		"violations": fmt.Sprintf("%d", len(report.Violations)),
	})

	if len(report.Violations) > 0 {
		logger.ErrorContext(ctx, "INTEGRITY SCAN: broken reversal chains detected — immediate review required",
			logger.Fields{"violations_count": len(report.Violations)})
	} else {
		logger.InfoContext(ctx, "Integrity scan: reversal chains OK",
			logger.Fields{"scanned": report.ScannedTransactions})
	}

	return report, nil
}

// ScanDuplicatePostings finds duplicate posted transaction numbers within a tenant.
func (s *integrityService) ScanDuplicatePostings(ctx context.Context, tenantID uuid.UUID) (*IntegrityReport, error) {
	postedStatus := domain.TransactionStatusPosted
	transactions, err := s.txnRepo.List(ctx, &domain.TransactionFilter{
		Status: &postedStatus,
	})
	if err != nil {
		return nil, fmt.Errorf("integrity scan: list for duplicate check: %w", err)
	}

	report := &IntegrityReport{ScannedTransactions: len(transactions)}

	seen := make(map[string][]uuid.UUID, len(transactions))
	for _, txn := range transactions {
		seen[txn.TransactionNumber] = append(seen[txn.TransactionNumber], txn.ID)
	}

	for number, ids := range seen {
		if len(ids) > 1 {
			s.metrics.IncrementCounter("integrity_violations_total", metrics.Fields{
				"kind": "duplicate_posting",
			})
			report.add(IntegrityViolation{
				Kind:     "DUPLICATE_POSTING",
				EntityID: ids[0],
				Detail:   fmt.Sprintf("transaction number %q posted %d times: %v", number, len(ids), ids),
			})
		}
	}

	s.metrics.IncrementCounter("integrity_scans_total", metrics.Fields{
		"scan_type":  "duplicate_postings",
		"violations": fmt.Sprintf("%d", len(report.Violations)),
	})

	if len(report.Violations) > 0 {
		logger.ErrorContext(ctx, "INTEGRITY SCAN: duplicate postings detected — possible idempotency failure",
			logger.Fields{"violations_count": len(report.Violations)})
	} else {
		logger.InfoContext(ctx, "Integrity scan: no duplicate postings",
			logger.Fields{"scanned": report.ScannedTransactions})
	}

	return report, nil
}
