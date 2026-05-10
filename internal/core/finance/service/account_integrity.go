// Package service — COA integrity service.
//
// COAIntegrityService performs scheduled and on-demand structural integrity
// scans of the Chart of Accounts. It detects:
//
//   - Orphan accounts   (parent referenced but not found)
//   - Cyclic hierarchy  (materialised path is inconsistent with parentage)
//   - Root-type mismatch between parent and child
//   - Depth violations  (level > MaxAccountHierarchyDepth)
//   - Dead accounts     (ACTIVE but no transactions within a configurable threshold)
//   - Balance in closed accounts (CLOSED/ARCHIVED with non-zero balance)
//   - Normal-balance/root-type inconsistency
//   - Duplicate account codes within a tenant
//   - Balance on DRAFT/PENDING accounts (should be zero)
//
// Each finding is tagged with a COAViolationSeverity (CRITICAL/HIGH/MEDIUM/LOW)
// so the governance dashboard can gate deployments on critical findings.
//
// Nil-safe on the service receiver.
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"awo.so/internal/core/finance/domain"
	"awo.so/internal/shared"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
)

// COAIntegrityService scans the chart of accounts for structural violations.
type COAIntegrityService struct {
	accounts      domain.AccountsRepository
	events        domain.AccountLifecycleEventRepository // nil = no event persistence
	metrics       metrics.MetricsProvider
	deadThreshold time.Duration // minimum inactivity period to flag a dead account
}

// DefaultDeadAccountThreshold is the inactivity period after which an ACTIVE
// account with no transactions is flagged as potentially dead.
const DefaultDeadAccountThreshold = 365 * 24 * time.Hour // 1 year

// NewCOAIntegrityService constructs the integrity scanner.
// deadThreshold = 0 means use DefaultDeadAccountThreshold.
func NewCOAIntegrityService(
	accounts domain.AccountsRepository,
	events domain.AccountLifecycleEventRepository,
	m metrics.MetricsProvider,
	deadThreshold time.Duration,
) *COAIntegrityService {
	if deadThreshold == 0 {
		deadThreshold = DefaultDeadAccountThreshold
	}
	return &COAIntegrityService{
		accounts:      accounts,
		events:        events,
		metrics:       m,
		deadThreshold: deadThreshold,
	}
}

// =============================================================================
// Full scan
// =============================================================================

// ScanTenant performs a full COA integrity scan for the tenant in ctx.
// Returns a COAIntegrityReport regardless of whether violations are found.
// The Healthy field is false when any CRITICAL or HIGH violations exist.
func (s *COAIntegrityService) ScanTenant(ctx context.Context) (*domain.COAIntegrityReport, error) {
	if s == nil {
		return &domain.COAIntegrityReport{Healthy: true, GeneratedAt: time.Now()}, nil
	}

	tenantID, _ := shared.GetTenantID(ctx)

	// Load all accounts for the tenant (tenant scoping is enforced by the
	// repository via ctx, so no TenantID field on the filter).
	filter := &domain.AccountFilter{}
	accounts, err := s.accounts.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("COA integrity scan: failed to load accounts: %w", err)
	}

	report := &domain.COAIntegrityReport{
		TenantID:        tenantID,
		GeneratedAt:     time.Now(),
		AccountsScanned: len(accounts),
	}

	// Build lookup maps for efficient cross-account checks.
	byID := make(map[uuid.UUID]*domain.Accounts, len(accounts))
	codeCount := make(map[string]int, len(accounts))
	for _, a := range accounts {
		byID[a.ID] = a
		codeCount[a.AccountCode]++
	}

	// Run all checks.
	for _, a := range accounts {
		report.Violations = append(report.Violations, s.checkOrphan(a, byID)...)
		report.Violations = append(report.Violations, s.checkRootTypeMismatch(a, byID)...)
		report.Violations = append(report.Violations, s.checkDepth(a)...)
		report.Violations = append(report.Violations, s.checkDeadAccount(a)...)
		report.Violations = append(report.Violations, s.checkBalanceInClosed(a)...)
		report.Violations = append(report.Violations, s.checkNormalBalanceMismatch(a)...)
		report.Violations = append(report.Violations, s.checkInvalidStatusBalance(a)...)
	}

	// Duplicate code check (cross-account, run once).
	report.Violations = append(report.Violations, s.checkDuplicateCodes(accounts, codeCount)...)

	// Cycle check via materialised path consistency.
	report.Violations = append(report.Violations, s.checkCycles(accounts, byID)...)

	report.Healthy = !report.HasHigh()

	s.emitScanMetrics(ctx, report)

	logger.InfoContext(ctx, "COA integrity scan complete", logger.Fields{
		"tenant_id":        tenantID.String(),
		"accounts_scanned": report.AccountsScanned,
		"violations":       len(report.Violations),
		"healthy":          report.Healthy,
	})

	return report, nil
}

// =============================================================================
// Individual checks (each returns 0 or more violations)
// =============================================================================

// checkOrphan detects accounts whose ParentID references a non-existent account.
func (s *COAIntegrityService) checkOrphan(
	a *domain.Accounts,
	byID map[uuid.UUID]*domain.Accounts,
) []domain.COAIntegrityViolation {
	if a.ParentAccountID == nil {
		return nil
	}
	if _, ok := byID[*a.ParentAccountID]; !ok {
		return []domain.COAIntegrityViolation{{
			Kind:      domain.COAViolationOrphanAccount,
			Severity:  domain.COASeverityCritical,
			AccountID: a.ID,
			Detail: fmt.Sprintf("account %s references parent %s which does not exist in this tenant",
				a.AccountCode, *a.ParentAccountID),
		}}
	}
	return nil
}

// checkRootTypeMismatch detects accounts whose root type differs from their parent.
func (s *COAIntegrityService) checkRootTypeMismatch(
	a *domain.Accounts,
	byID map[uuid.UUID]*domain.Accounts,
) []domain.COAIntegrityViolation {
	if a.ParentAccountID == nil {
		return nil
	}
	parent, ok := byID[*a.ParentAccountID]
	if !ok {
		return nil // orphan already flagged
	}
	if parent.RootType != a.RootType {
		return []domain.COAIntegrityViolation{{
			Kind:      domain.COAViolationRootTypeMismatch,
			Severity:  domain.COASeverityHigh,
			AccountID: a.ID,
			Detail: fmt.Sprintf("account %s (root=%s) has parent %s (root=%s) — root types must match",
				a.AccountCode, a.RootType, parent.AccountCode, parent.RootType),
		}}
	}
	return nil
}

// checkDepth detects accounts whose AccountLevel exceeds MaxAccountHierarchyDepth.
func (s *COAIntegrityService) checkDepth(a *domain.Accounts) []domain.COAIntegrityViolation {
	if int(a.AccountLevel) > domain.MaxAccountHierarchyDepth {
		return []domain.COAIntegrityViolation{{
			Kind:      domain.COAViolationDepthExceeded,
			Severity:  domain.COASeverityHigh,
			AccountID: a.ID,
			Detail: fmt.Sprintf("account %s is at depth %d (max %d)",
				a.AccountCode, a.AccountLevel, domain.MaxAccountHierarchyDepth),
		}}
	}
	return nil
}

// checkDeadAccount flags ACTIVE accounts that have had no transactions within
// the configured dead threshold.
func (s *COAIntegrityService) checkDeadAccount(a *domain.Accounts) []domain.COAIntegrityViolation {
	if a.Status != domain.AccountStatusActive {
		return nil
	}
	// If the account has never transacted AND was created long ago, flag it.
	if a.LastTransactionDate == nil {
		age := time.Since(a.CreatedAt)
		if age > s.deadThreshold {
			return []domain.COAIntegrityViolation{{
				Kind:      domain.COAViolationDeadAccount,
				Severity:  domain.COASeverityLow,
				AccountID: a.ID,
				Detail: fmt.Sprintf("account %s has been ACTIVE for %d days with no transactions",
					a.AccountCode, int(age.Hours()/24)),
			}}
		}
		return nil
	}
	// Account transacted before but not recently.
	sinceLastTxn := time.Since(*a.LastTransactionDate)
	if sinceLastTxn > s.deadThreshold {
		return []domain.COAIntegrityViolation{{
			Kind:      domain.COAViolationDeadAccount,
			Severity:  domain.COASeverityLow,
			AccountID: a.ID,
			Detail: fmt.Sprintf("account %s has had no transactions for %d days (last: %s)",
				a.AccountCode, int(sinceLastTxn.Hours()/24),
				a.LastTransactionDate.Format("2006-01-02")),
		}}
	}
	return nil
}

// checkBalanceInClosed detects CLOSED or ARCHIVED accounts with non-zero balances.
func (s *COAIntegrityService) checkBalanceInClosed(a *domain.Accounts) []domain.COAIntegrityViolation {
	isClosed := a.Status == domain.AccountStatusClosed || a.Status == domain.AccountStatusArchived
	if !isClosed {
		return nil
	}
	if !a.CurrentBalance.IsZero() {
		return []domain.COAIntegrityViolation{{
			Kind:      domain.COAViolationBalanceInClosedAccount,
			Severity:  domain.COASeverityCritical,
			AccountID: a.ID,
			Detail: func() string {
				cur := ""
				if a.CurrencyCode != nil {
					cur = " " + *a.CurrencyCode
				}
				return fmt.Sprintf("account %s has status=%s but carries a non-zero balance of %s%s",
					a.AccountCode, a.Status, a.CurrentBalance.String(), cur)
			}(),
		}}
	}
	return nil
}

// checkNormalBalanceMismatch detects accounts where NormalBalance is inconsistent
// with their RootType (e.g. Asset with a CREDIT normal balance).
func (s *COAIntegrityService) checkNormalBalanceMismatch(a *domain.Accounts) []domain.COAIntegrityViolation {
	expected := domain.GetNormalBalanceForRootType(a.RootType)
	if a.NormalBalance != expected {
		return []domain.COAIntegrityViolation{{
			Kind:      domain.COAViolationNormalBalanceMismatch,
			Severity:  domain.COASeverityHigh,
			AccountID: a.ID,
			Detail: fmt.Sprintf("account %s has root_type=%s but normal_balance=%s (expected %s)",
				a.AccountCode, a.RootType, a.NormalBalance, expected),
		}}
	}
	return nil
}

// checkInvalidStatusBalance flags DRAFT or PENDING_APPROVAL accounts with
// non-zero balances (they should not accept postings yet).
func (s *COAIntegrityService) checkInvalidStatusBalance(a *domain.Accounts) []domain.COAIntegrityViolation {
	isPre := a.Status == domain.AccountStatusDraft || a.Status == domain.AccountStatusPendingApproval
	if !isPre {
		return nil
	}
	if !a.CurrentBalance.IsZero() {
		return []domain.COAIntegrityViolation{{
			Kind:      domain.COAViolationInvalidStatusForBalance,
			Severity:  domain.COASeverityHigh,
			AccountID: a.ID,
			Detail: fmt.Sprintf("account %s has status=%s but carries balance %s — DRAFT/PENDING accounts must have zero balance",
				a.AccountCode, a.Status, a.CurrentBalance.String()),
		}}
	}
	return nil
}

// checkDuplicateCodes detects accounts that share the same account code within
// a tenant. All duplicates are flagged (not just one).
func (s *COAIntegrityService) checkDuplicateCodes(
	accounts []*domain.Accounts,
	codeCount map[string]int,
) []domain.COAIntegrityViolation {
	var violations []domain.COAIntegrityViolation
	for _, a := range accounts {
		if codeCount[a.AccountCode] > 1 {
			violations = append(violations, domain.COAIntegrityViolation{
				Kind:      domain.COAViolationDuplicateCode,
				Severity:  domain.COASeverityCritical,
				AccountID: a.ID,
				Detail: fmt.Sprintf("account code %s appears %d times in this tenant",
					a.AccountCode, codeCount[a.AccountCode]),
			})
		}
	}
	return violations
}

// checkCycles detects accounts where the materialised path is inconsistent
// with the parent chain (i.e. cycle would be reached by following ParentIDs).
//
// Strategy: for each account, walk the parent chain. If we visit more than
// len(accounts) nodes before reaching the root, a cycle exists.
func (s *COAIntegrityService) checkCycles(
	accounts []*domain.Accounts,
	byID map[uuid.UUID]*domain.Accounts,
) []domain.COAIntegrityViolation {
	maxDepth := len(accounts) + 1
	var violations []domain.COAIntegrityViolation

	for _, a := range accounts {
		if a.ParentAccountID == nil {
			continue
		}
		visited := make(map[uuid.UUID]bool)
		visited[a.ID] = true
		cur := a

		for steps := 0; steps < maxDepth; steps++ {
			if cur.ParentAccountID == nil {
				break // reached root — no cycle
			}
			if visited[*cur.ParentAccountID] {
				violations = append(violations, domain.COAIntegrityViolation{
					Kind:      domain.COAViolationCyclicHierarchy,
					Severity:  domain.COASeverityCritical,
					AccountID: a.ID,
					Detail: fmt.Sprintf("account %s is part of a cyclic hierarchy (cycle detected at parent %s)",
						a.AccountCode, *cur.ParentAccountID),
				})
				break
			}
			parent, ok := byID[*cur.ParentAccountID]
			if !ok {
				break // orphan — already handled
			}
			visited[*cur.ParentAccountID] = true
			cur = parent
		}
	}
	return violations
}

// =============================================================================
// Targeted checks (for use during mutation governance)
// =============================================================================

// CheckSingleAccount runs all integrity checks for a single account and returns
// any violations found. Useful in pre-mutation validation paths.
func (s *COAIntegrityService) CheckSingleAccount(
	ctx context.Context,
	accountID uuid.UUID,
) ([]domain.COAIntegrityViolation, error) {
	if s == nil {
		return nil, nil
	}

	a, err := s.accounts.GetByID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("account not found: %w", err)
	}

	// Build minimal byID map for cross-account checks.
	byID := map[uuid.UUID]*domain.Accounts{a.ID: a}
	if a.ParentAccountID != nil {
		parent, err := s.accounts.GetByID(ctx, *a.ParentAccountID)
		if err == nil {
			byID[*a.ParentAccountID] = parent
		}
	}

	var violations []domain.COAIntegrityViolation
	violations = append(violations, s.checkOrphan(a, byID)...)
	violations = append(violations, s.checkRootTypeMismatch(a, byID)...)
	violations = append(violations, s.checkDepth(a)...)
	violations = append(violations, s.checkDeadAccount(a)...)
	violations = append(violations, s.checkBalanceInClosed(a)...)
	violations = append(violations, s.checkNormalBalanceMismatch(a)...)
	violations = append(violations, s.checkInvalidStatusBalance(a)...)

	return violations, nil
}

// =============================================================================
// Metrics
// =============================================================================

func (s *COAIntegrityService) emitScanMetrics(ctx context.Context, report *domain.COAIntegrityReport) {
	if s.metrics == nil {
		return
	}

	counts := map[domain.COAViolationSeverity]int{}
	for _, v := range report.Violations {
		counts[v.Severity]++
	}

	s.metrics.IncrementCounter("finance_coa_integrity_scans_total", metrics.Fields{
		"tenant_id": report.TenantID.String(),
		"healthy":   fmt.Sprintf("%v", report.Healthy),
	})

	if len(report.Violations) > 0 {
		for sev, count := range counts {
			for i := 0; i < count; i++ {
				s.metrics.IncrementCounter("finance_coa_integrity_violations_total", metrics.Fields{
					"severity":  string(sev),
					"tenant_id": report.TenantID.String(),
				})
			}
		}
	}
}
