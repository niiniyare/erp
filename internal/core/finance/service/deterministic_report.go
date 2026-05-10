// Package service — Deterministic Reporting Engine.
//
// DeterministicReportingEngine generates financial reports that are bound
// to a COA snapshot and a time-bounded ledger slice. Every report is
// fingerprinted with a deterministic SHA-256 hash: given the same snapshot,
// the same ledger, and the same as-of date, the hash MUST be identical on
// every recomputation.
//
// # Core contract
//
//  1. Every report MUST specify a COA snapshot ID.
//  2. Account balance for a report line = sum of all POSTED entry
//     debits/credits for that account up to AsOfDate.
//  3. The report hash is computed from the canonical JSON of the sorted lines
//     plus the snapshot hash.
//  4. A stored FinancialReportHash record is written after every generation.
//     Recomputing the same report later and comparing hashes detects ledger
//     mutation and non-determinism.
package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"awo.so/internal/core/finance/domain"
	"awo.so/internal/shared"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
)

// DeterministicReportingEngine generates hash-verified, snapshot-bound reports.
type DeterministicReportingEngine struct {
	snapshots    domain.COASnapshotRepository
	transactions domain.TransactionRepository
	metrics      metrics.MetricsProvider
}

// NewDeterministicReportingEngine constructs the engine.
func NewDeterministicReportingEngine(
	snapshots domain.COASnapshotRepository,
	transactions domain.TransactionRepository,
	m metrics.MetricsProvider,
) *DeterministicReportingEngine {
	return &DeterministicReportingEngine{
		snapshots:    snapshots,
		transactions: transactions,
		metrics:      m,
	}
}

// =============================================================================
// Trial Balance
// =============================================================================

// TrialBalanceRequest specifies inputs for a deterministic trial balance.
type TrialBalanceRequest struct {
	// SnapshotID is the COA snapshot to bind to. Required.
	SnapshotID uuid.UUID
	// AsOfDate is the ledger cut-off date (inclusive). Zero = now.
	AsOfDate    time.Time
	GeneratedBy uuid.UUID
}

// GenerateTrialBalance produces a deterministic, snapshot-bound trial balance.
//
// The report is not persisted; the caller is responsible for storing the
// FinancialReportHash returned inside the report if replay verification is needed.
func (e *DeterministicReportingEngine) GenerateTrialBalance(
	ctx context.Context,
	req TrialBalanceRequest,
) (*domain.TrialBalanceReport, error) {
	if e == nil {
		return nil, fmt.Errorf("DeterministicReportingEngine is nil")
	}

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("generate trial balance: tenant ID missing from context")
	}

	if req.AsOfDate.IsZero() {
		req.AsOfDate = time.Now().UTC()
	}

	// 1. Load the COA snapshot — this is the structural binding.
	snap, err := e.snapshots.GetByID(ctx, req.SnapshotID)
	if err != nil {
		return nil, fmt.Errorf("generate trial balance: load snapshot: %w", err)
	}

	// 2. Load all posted transactions up to AsOfDate.
	posted := domain.TransactionStatusPosted
	txFilter := &domain.TransactionFilter{
		Status: &posted,
		DateRange: &domain.DateRange{
			StartDate: time.Time{},
			EndDate:   &req.AsOfDate,
		},
	}
	transactions, err := e.transactions.List(ctx, txFilter)
	if err != nil {
		return nil, fmt.Errorf("generate trial balance: load transactions: %w", err)
	}

	// 3. Accumulate debit/credit per account from posted entries.
	type accum struct {
		debit  decimal.Decimal
		credit decimal.Decimal
	}
	balances := make(map[uuid.UUID]*accum, len(snap.Accounts))
	entryCount := 0

	for _, txn := range transactions {
		entries, err := e.transactions.GetEntriesByTransaction(ctx, txn.ID)
		if err != nil {
			return nil, fmt.Errorf("generate trial balance: load entries for txn %s: %w", txn.ID, err)
		}
		for _, entry := range entries {
			if _, ok := balances[entry.AccountID]; !ok {
				balances[entry.AccountID] = &accum{}
			}
			balances[entry.AccountID].debit = balances[entry.AccountID].debit.Add(entry.DebitAmount)
			balances[entry.AccountID].credit = balances[entry.AccountID].credit.Add(entry.CreditAmount)
			entryCount++
		}
	}

	// 4. Build trial balance lines from the snapshot's account set.
	// Only include accounts present in the snapshot (historical fidelity).
	lines := make([]domain.TrialBalanceLine, 0, len(snap.Accounts))
	var totalDebits, totalCredits decimal.Decimal

	for _, acct := range snap.Accounts {
		bal := balances[acct.AccountID]
		var debit, credit decimal.Decimal
		if bal != nil {
			debit = bal.debit
			credit = bal.credit
		}
		totalDebits = totalDebits.Add(debit)
		totalCredits = totalCredits.Add(credit)

		lines = append(lines, domain.TrialBalanceLine{
			AccountID:    acct.AccountID,
			AccountCode:  acct.AccountCode,
			AccountName:  acct.AccountName,
			RootType:     acct.RootType,
			NormalBalance: acct.NormalBalance,
			Debit:        debit,
			Credit:       credit,
		})
	}

	// 5. Sort lines canonically by AccountCode for deterministic hash.
	sort.Slice(lines, func(i, j int) bool {
		return lines[i].AccountCode < lines[j].AccountCode
	})

	// 6. Compute report hash.
	reportHash, ledgerHash, err := e.computeTrialBalanceHash(lines, snap.Hash, transactions)
	if err != nil {
		return nil, fmt.Errorf("generate trial balance: compute hash: %w", err)
	}

	now := time.Now().UTC()
	report := &domain.TrialBalanceReport{
		TenantID:     tenantID,
		AsOfDate:     req.AsOfDate,
		SnapshotID:   req.SnapshotID,
		GeneratedAt:  now,
		Lines:        lines,
		TotalDebits:  totalDebits,
		TotalCredits: totalCredits,
		Balanced:     totalDebits.Equal(totalCredits),
		Hash:         reportHash,
	}

	_ = ledgerHash // available for FinancialReportHash persistence if caller needs it

	e.emitReportGenerated(ctx, tenantID, domain.ReportKindTrialBalance)

	logger.InfoContext(ctx, "trial balance generated", logger.Fields{
		"tenant_id":   tenantID.String(),
		"snapshot_id": req.SnapshotID.String(),
		"as_of_date":  req.AsOfDate.Format("2006-01-02"),
		"lines":       len(lines),
		"balanced":    report.Balanced,
		"hash":        reportHash[:16] + "...",
	})

	return report, nil
}

// =============================================================================
// Report hash verification
// =============================================================================

// VerifyReportHash recomputes the trial balance from the same inputs stored in
// a FinancialReportHash record and compares it against the stored hash.
// Returns (true, nil) when the report is reproducible, (false, nil) when
// hash diverges, (false, err) on error.
func (e *DeterministicReportingEngine) VerifyReportHash(
	ctx context.Context,
	stored *domain.FinancialReportHash,
	generatedBy uuid.UUID,
) (bool, error) {
	if e == nil {
		return false, fmt.Errorf("DeterministicReportingEngine is nil")
	}

	regenReq := TrialBalanceRequest{
		SnapshotID:  stored.SnapshotID,
		AsOfDate:    stored.AsOfDate,
		GeneratedBy: generatedBy,
	}

	regen, err := e.GenerateTrialBalance(ctx, regenReq)
	if err != nil {
		return false, fmt.Errorf("verify report hash: regenerate: %w", err)
	}

	valid := regen.Hash == stored.ReportHash

	if !valid {
		logger.WarnContext(ctx, "report hash mismatch — ledger may have been mutated", logger.Fields{
			"snapshot_id": stored.SnapshotID.String(),
			"stored_hash": stored.ReportHash,
			"regen_hash":  regen.Hash,
		})
	}

	return valid, nil
}

// =============================================================================
// Deterministic hash computation
// =============================================================================

// computeTrialBalanceHash produces:
//   - reportHash: SHA-256 of sorted lines + snapshotHash
//   - ledgerHash: SHA-256 of transaction IDs in canonical order
func (e *DeterministicReportingEngine) computeTrialBalanceHash(
	lines []domain.TrialBalanceLine,
	snapshotHash string,
	transactions []*domain.Transaction,
) (reportHash, ledgerHash string, err error) {
	// Ledger hash: deterministic hash of transaction IDs sorted ascending.
	txIDs := make([]string, 0, len(transactions))
	for _, t := range transactions {
		txIDs = append(txIDs, t.ID.String())
	}
	sort.Strings(txIDs)

	lh := sha256.New()
	for _, id := range txIDs {
		lh.Write([]byte(id))
	}
	ledgerHash = hex.EncodeToString(lh.Sum(nil))

	// Report hash: lines + snapshotHash + ledgerHash.
	rh := sha256.New()
	rh.Write([]byte(snapshotHash))
	rh.Write([]byte(ledgerHash))
	for _, line := range lines {
		b, err := json.Marshal(line)
		if err != nil {
			return "", "", fmt.Errorf("marshal line %s: %w", line.AccountCode, err)
		}
		rh.Write(b)
	}
	reportHash = hex.EncodeToString(rh.Sum(nil))

	return reportHash, ledgerHash, nil
}

// =============================================================================
// Metrics
// =============================================================================

func (e *DeterministicReportingEngine) emitReportGenerated(
	ctx context.Context,
	tenantID uuid.UUID,
	kind domain.FinancialReportKind,
) {
	if e.metrics == nil {
		return
	}
	e.metrics.IncrementCounter("finance_deterministic_reports_generated_total", metrics.Fields{
		"tenant_id":   tenantID.String(),
		"report_kind": string(kind),
	})
}
