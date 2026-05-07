package service

// integrity.go — Financial drift detection and ledger consistency verification.
//
// The IntegrityService scans the finance ledger for corruption that cannot be
// prevented by runtime guards alone: orphaned entries, unbalanced posted
// transactions, broken reversal chains, and duplicated postings.
//
// # Architecture
//
// Checks are registered via the Check interface. Add a new check by implementing
// Check and passing it to NewIntegrityServiceWithChecks. Built-in checks are
// in the builtinChecks() constructor helper.
//
// # Pagination
//
// All scans page through the DB using MaxIntegrityScanPage to prevent OOM on
// large tenants. Callers receive a combined report across all pages.
//
// # Operational use
//
//   - Run on a scheduled basis (nightly Temporal cron)
//   - Run after any bulk import or migration
//   - Run as part of period-close sign-off
//   - Run in incident response when ledger correctness is in doubt
//
// All methods are READ-ONLY — they detect drift but never auto-correct.
// Correction requires a human-initiated reversal/amendment flow.

//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"awo.so/internal/core/finance/domain"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
)

// ── Severity ─────────────────────────────────────────────────────────────────

// ViolationSeverity classifies the operational urgency of a detected violation.
type ViolationSeverity string

const (
	// SeverityCritical: Immediate action required — ledger correctness is
	// compromised. Trigger a P0 incident. Examples: unbalanced posted transaction.
	SeverityCritical ViolationSeverity = "CRITICAL"

	// SeverityHigh: Action required before period close. Examples: broken
	// reversal chain, reversal transaction not posted.
	SeverityHigh ViolationSeverity = "HIGH"

	// SeverityMedium: Investigate within 48 hours. Examples: duplicate posting.
	SeverityMedium ViolationSeverity = "MEDIUM"

	// SeverityLow: Background hygiene issue. Examples: posted transaction with
	// zero entries (stale cleanup task).
	SeverityLow ViolationSeverity = "LOW"
)

// ── IntegrityViolation ────────────────────────────────────────────────────────

// IntegrityViolation describes a single detected inconsistency.
type IntegrityViolation struct {
	// Kind is a machine-readable violation code (e.g. "UNBALANCED_TRANSACTION").
	Kind string `json:"kind"`

	// Severity classifies operational urgency.
	Severity ViolationSeverity `json:"severity"`

	// EntityID is the UUID of the primary affected entity.
	EntityID uuid.UUID `json:"entity_id"`

	// Detail is a human-readable explanation suitable for operator review.
	Detail string `json:"detail"`

	// RepairAction describes the recommended remediation step.
	// Empty string means no automated path exists — escalate to engineering.
	RepairAction string `json:"repair_action,omitempty"`
}

func (v IntegrityViolation) Error() string {
	return fmt.Sprintf("[%s/%s] %s (entity=%s)", v.Severity, v.Kind, v.Detail, v.EntityID)
}

// IntegrityReport aggregates violations found by a single integrity scan pass.
type IntegrityReport struct {
	Violations []IntegrityViolation `json:"violations"`

	// ScannedTransactions is the number of transactions examined.
	ScannedTransactions int `json:"scanned_transactions"`

	// ScannedEntries is the number of entries examined.
	ScannedEntries int `json:"scanned_entries"`

	// ChecksRun lists the check kind strings that executed in this scan.
	ChecksRun []string `json:"checks_run"`
}

func (r *IntegrityReport) add(v IntegrityViolation) {
	r.Violations = append(r.Violations, v)
}

// HasCritical returns true if any CRITICAL violation was found.
// Use this to gate period-close or deployment sign-off.
func (r *IntegrityReport) HasCritical() bool {
	for _, v := range r.Violations {
		if v.Severity == SeverityCritical {
			return true
		}
	}
	return false
}

// ── Check interface ───────────────────────────────────────────────────────────

// Check is a pluggable integrity check. Implement this interface to add
// custom verification logic without modifying the core IntegrityService.
//
// Kind returns a stable machine-readable identifier used in metrics and reports.
// Execute performs the check and appends any violations to report.
type Check interface {
	Kind() string
	Execute(ctx context.Context, txn *domain.Transaction, entries []domain.TransactionEntry, report *IntegrityReport)
}

// ── IntegrityService ──────────────────────────────────────────────────────────

// IntegrityService performs read-only consistency scans of the finance ledger.
type IntegrityService interface {
	// ScanPostedTransactions checks every posted transaction within the supplied
	// filter for double-entry balance violations and runs all registered checks.
	// Paginated: loads at most MaxIntegrityScanPage transactions per DB round-trip.
	ScanPostedTransactions(ctx context.Context, filter *domain.TransactionFilter) (*IntegrityReport, error)

	// ScanReversalChains verifies that every transaction marked REVERSED has a
	// valid reversal record and that the reversal transaction exists and is POSTED.
	ScanReversalChains(ctx context.Context, filter *domain.TransactionFilter) (*IntegrityReport, error)

	// ScanDuplicatePostings finds posted transactions whose transaction number
	// appears more than once within the result set.
	// Paginated: loads at most MaxIntegrityScanPage transactions per DB round-trip.
	ScanDuplicatePostings(ctx context.Context, tenantID uuid.UUID) (*IntegrityReport, error)
}

// ── implementation ────────────────────────────────────────────────────────────

type integrityService struct {
	txnRepo             domain.TransactionRepository
	reversalHistoryRepo domain.ReversalHistoryRepository // nil → reversal scan skipped
	checks              []Check
	metrics             metrics.MetricsProvider
}

// NewIntegrityService constructs an IntegrityService with the built-in checks.
// reversalHistoryRepo may be nil — reversal chain scanning will be skipped.
func NewIntegrityService(
	txnRepo domain.TransactionRepository,
	reversalHistoryRepo domain.ReversalHistoryRepository,
	metrics metrics.MetricsProvider,
) IntegrityService {
	return NewIntegrityServiceWithChecks(txnRepo, reversalHistoryRepo, metrics, builtinChecks()...)
}

// NewIntegrityServiceWithChecks constructs an IntegrityService with a custom
// check set. Use this to inject domain-specific or tenant-specific checks.
func NewIntegrityServiceWithChecks(
	txnRepo domain.TransactionRepository,
	reversalHistoryRepo domain.ReversalHistoryRepository,
	metrics metrics.MetricsProvider,
	checks ...Check,
) IntegrityService {
	return &integrityService{
		txnRepo:             txnRepo,
		reversalHistoryRepo: reversalHistoryRepo,
		checks:              checks,
		metrics:             metrics,
	}
}

// builtinChecks returns the default set of integrity checks applied on every scan.
func builtinChecks() []Check {
	return []Check{
		&balanceCheck{},
		&noEntriesCheck{},
	}
}

// ── built-in checks ───────────────────────────────────────────────────────────

// balanceCheck verifies ∑debit == ∑credit for a posted transaction.
type balanceCheck struct{}

func (c *balanceCheck) Kind() string { return "UNBALANCED_TRANSACTION" }

func (c *balanceCheck) Execute(_ context.Context, txn *domain.Transaction, entries []domain.TransactionEntry, report *IntegrityReport) {
	if len(entries) == 0 {
		return // handled by noEntriesCheck
	}
	var totalDebit, totalCredit decimal.Decimal
	for _, e := range entries {
		totalDebit = totalDebit.Add(e.DebitAmount)
		totalCredit = totalCredit.Add(e.CreditAmount)
	}
	if !totalDebit.Equal(totalCredit) {
		report.add(IntegrityViolation{
			Kind:     c.Kind(),
			Severity: SeverityCritical,
			EntityID: txn.ID,
			Detail: fmt.Sprintf(
				"transaction %s: debits=%s credits=%s diff=%s",
				txn.TransactionNumber, totalDebit, totalCredit,
				totalDebit.Sub(totalCredit).Abs(),
			),
			RepairAction: "Identify and reverse the unbalanced transaction. DO NOT UPDATE ENTRIES DIRECTLY. Use ReverseTransaction then create a correcting journal entry.",
		})
	}
}

// noEntriesCheck detects posted transactions with no entries — a sign of
// partial commit or migration corruption.
type noEntriesCheck struct{}

func (c *noEntriesCheck) Kind() string { return "POSTED_WITHOUT_ENTRIES" }

func (c *noEntriesCheck) Execute(_ context.Context, txn *domain.Transaction, entries []domain.TransactionEntry, report *IntegrityReport) {
	if len(entries) == 0 {
		report.add(IntegrityViolation{
			Kind:     c.Kind(),
			Severity: SeverityHigh,
			EntityID: txn.ID,
			Detail:   fmt.Sprintf("transaction %s posted with no entries", txn.TransactionNumber),
			RepairAction: "Cancel or reverse the transaction if it appears in reports. Investigate whether a migration partially committed the header without entries.",
		})
	}
}

// ── ScanPostedTransactions ────────────────────────────────────────────────────

// ScanPostedTransactions pages through all posted transactions and runs all
// registered checks per transaction.
func (s *integrityService) ScanPostedTransactions(ctx context.Context, filter *domain.TransactionFilter) (*IntegrityReport, error) {
	postedStatus := domain.TransactionStatusPosted
	scanFilter := &domain.TransactionFilter{}
	if filter != nil {
		*scanFilter = *filter
	}
	scanFilter.Status = &postedStatus

	report := &IntegrityReport{}
	for _, c := range s.checks {
		report.ChecksRun = append(report.ChecksRun, c.Kind())
	}

	// Paginate to avoid loading the entire ledger into memory.
	pageSize := domain.MaxIntegrityScanPage
	offset := 0
	scanFilter.Limit = &pageSize

	for {
		scanFilter.Offset = &offset
		page, err := s.txnRepo.List(ctx, scanFilter)
		if err != nil {
			return nil, fmt.Errorf("integrity scan: list posted page offset=%d: %w", offset, err)
		}
		if len(page) == 0 {
			break
		}

		report.ScannedTransactions += len(page)

		for _, txn := range page {
			entries, err := s.txnRepo.GetEntriesByTransaction(ctx, txn.ID)
			if err != nil {
				logger.ErrorContext(ctx, "integrity scan: failed to fetch entries",
					logger.Fields{"transaction_id": txn.ID.String(), "error": err.Error()})
				report.add(IntegrityViolation{
					Kind:         "ENTRY_FETCH_FAILED",
					Severity:     SeverityHigh,
					EntityID:     txn.ID,
					Detail:       fmt.Sprintf("could not fetch entries for %s: %v", txn.TransactionNumber, err),
					RepairAction: "Check DB connectivity and re-run the scan.",
				})
				continue
			}
			report.ScannedEntries += len(entries)

			for _, check := range s.checks {
				check.Execute(ctx, txn, entries, report)
			}
		}

		if len(page) < pageSize {
			break // last page
		}
		offset += pageSize
	}

	s.emitScanMetrics(ctx, "posted_transactions", report)
	return report, nil
}

// ── ScanReversalChains ────────────────────────────────────────────────────────

func (s *integrityService) ScanReversalChains(ctx context.Context, filter *domain.TransactionFilter) (*IntegrityReport, error) {
	if s.reversalHistoryRepo == nil {
		logger.WarnContext(ctx, "Integrity scan: reversal chain scan skipped — no reversalHistoryRepo configured")
		return &IntegrityReport{ChecksRun: []string{"REVERSAL_CHAIN"}}, nil
	}

	reversedStatus := domain.TransactionStatusReversed
	scanFilter := &domain.TransactionFilter{}
	if filter != nil {
		*scanFilter = *filter
	}
	scanFilter.Status = &reversedStatus

	report := &IntegrityReport{ChecksRun: []string{"REVERSAL_CHAIN"}}

	pageSize := domain.MaxIntegrityScanPage
	offset := 0
	scanFilter.Limit = &pageSize

	for {
		scanFilter.Offset = &offset
		page, err := s.txnRepo.List(ctx, scanFilter)
		if err != nil {
			return nil, fmt.Errorf("integrity scan: list reversed page offset=%d: %w", offset, err)
		}
		if len(page) == 0 {
			break
		}
		report.ScannedTransactions += len(page)

		for _, txn := range page {
			records, err := s.reversalHistoryRepo.GetByOriginal(ctx, txn.ID)
			if err != nil {
				report.add(IntegrityViolation{
					Kind:         "REVERSAL_HISTORY_FETCH_FAILED",
					Severity:     SeverityHigh,
					EntityID:     txn.ID,
					Detail:       fmt.Sprintf("could not fetch reversal history for %s: %v", txn.TransactionNumber, err),
					RepairAction: "Check DB connectivity and re-run the scan.",
				})
				continue
			}

			if len(records) == 0 {
				s.metrics.IncrementCounter("integrity_violations_total", metrics.Fields{"kind": "missing_reversal_record"})
				report.add(IntegrityViolation{
					Kind:         "MISSING_REVERSAL_RECORD",
					Severity:     SeverityHigh,
					EntityID:     txn.ID,
					Detail:       fmt.Sprintf("transaction %s marked REVERSED but has no reversal history record", txn.TransactionNumber),
					RepairAction: "Verify whether a reversal transaction exists by searching for REV- prefix. If found, insert the missing history record. If not found, this may indicate a partial commit — escalate to engineering.",
				})
				continue
			}

			for _, rec := range records {
				reversal, err := s.txnRepo.GetByID(ctx, rec.ReversalTransactionID)
				if err != nil {
					s.metrics.IncrementCounter("integrity_violations_total", metrics.Fields{"kind": "reversal_transaction_missing"})
					report.add(IntegrityViolation{
						Kind:         "REVERSAL_TRANSACTION_MISSING",
						Severity:     SeverityCritical,
						EntityID:     rec.ReversalTransactionID,
						Detail:       fmt.Sprintf("reversal history for %s points to reversal %s which does not exist", txn.TransactionNumber, rec.ReversalTransactionID),
						RepairAction: "This indicates a partial commit. The original is marked reversed but the reversal transaction is gone. Create a compensating entry manually after approval by finance controller.",
					})
					continue
				}
				if reversal.TransactionStatus != domain.TransactionStatusPosted {
					s.metrics.IncrementCounter("integrity_violations_total", metrics.Fields{"kind": "reversal_not_posted"})
					report.add(IntegrityViolation{
						Kind:         "REVERSAL_NOT_POSTED",
						Severity:     SeverityHigh,
						EntityID:     reversal.ID,
						Detail:       fmt.Sprintf("reversal %s for original %s has status %s (expected POSTED)", reversal.TransactionNumber, txn.TransactionNumber, reversal.TransactionStatus),
						RepairAction: "Post the reversal transaction if it is in APPROVED status. If it is in DRAFT, investigate whether the posting step was interrupted.",
					})
				}
			}
		}

		if len(page) < pageSize {
			break
		}
		offset += pageSize
	}

	s.emitScanMetrics(ctx, "reversal_chains", report)
	return report, nil
}

// ── ScanDuplicatePostings ─────────────────────────────────────────────────────

func (s *integrityService) ScanDuplicatePostings(ctx context.Context, _ uuid.UUID) (*IntegrityReport, error) {
	postedStatus := domain.TransactionStatusPosted
	report := &IntegrityReport{ChecksRun: []string{"DUPLICATE_POSTING"}}

	// Collect transaction numbers across all pages.
	seen := make(map[string][]uuid.UUID)
	pageSize := domain.MaxIntegrityScanPage
	offset := 0

	for {
		lim := pageSize
		off := offset
		page, err := s.txnRepo.List(ctx, &domain.TransactionFilter{
			Status: &postedStatus,
			Limit:  &lim,
			Offset: &off,
		})
		if err != nil {
			return nil, fmt.Errorf("integrity scan: duplicate check page offset=%d: %w", offset, err)
		}
		if len(page) == 0 {
			break
		}
		report.ScannedTransactions += len(page)
		for _, txn := range page {
			seen[txn.TransactionNumber] = append(seen[txn.TransactionNumber], txn.ID)
		}
		if len(page) < pageSize {
			break
		}
		offset += pageSize
	}

	for number, ids := range seen {
		if len(ids) > 1 {
			s.metrics.IncrementCounter("integrity_violations_total", metrics.Fields{"kind": "duplicate_posting"})
			report.add(IntegrityViolation{
				Kind:     "DUPLICATE_POSTING",
				Severity: SeverityMedium,
				EntityID: ids[0],
				Detail:   fmt.Sprintf("transaction number %q posted %d times: %v", number, len(ids), ids),
				RepairAction: "Identify which posting is the authoritative one. Reverse the duplicates. Investigate the idempotency failure that allowed both to be created.",
			})
		}
	}

	s.emitScanMetrics(ctx, "duplicate_postings", report)
	return report, nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

func (s *integrityService) emitScanMetrics(ctx context.Context, scanType string, report *IntegrityReport) {
	s.metrics.IncrementCounter("integrity_scans_total", metrics.Fields{
		"scan_type":  scanType,
		"violations": fmt.Sprintf("%d", len(report.Violations)),
	})
	for _, v := range report.Violations {
		s.metrics.IncrementCounter("integrity_violations_total", metrics.Fields{
			"kind":      v.Kind,
			"severity":  string(v.Severity),
			"scan_type": scanType,
		})
	}

	if len(report.Violations) > 0 {
		critCount := 0
		for _, v := range report.Violations {
			if v.Severity == SeverityCritical {
				critCount++
			}
		}
		logger.ErrorContext(ctx, "INTEGRITY SCAN: violations detected — immediate review required",
			logger.Fields{
				"scan_type":            scanType,
				"violations_count":     len(report.Violations),
				"critical_count":       critCount,
				"scanned_transactions": report.ScannedTransactions,
				"scanned_entries":      report.ScannedEntries,
			})
	} else {
		logger.InfoContext(ctx, "Integrity scan passed — no violations",
			logger.Fields{
				"scan_type": scanType,
				"scanned":   report.ScannedTransactions,
			})
	}
}
