package service

// safety_policy.go — Runtime Safety Policy Engine for the finance module.
//
// SafetyEnforcer evaluates configurable, tenant-aware policies before finance
// mutations to detect and contain abuse by compromised operators, runaway
// workflows, or malicious tenants.
//
// Enforcement actions:
//   - BLOCK  → returns a BusinessError; the mutation is rejected.
//   - WARN   → emits a metric and log; the mutation is allowed through.
//
// All methods are nil-safe: a nil *SafetyEnforcer passes all checks silently.
// This allows test wiring to pass nil without any setup overhead.

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"awo.so/internal/shared"
	"awo.so/internal/shared/errors"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
)

// SafetyAction is the enforcement action taken when a policy fires.
type SafetyAction string

const (
	SafetyActionBlock    SafetyAction = "BLOCK"
	SafetyActionWarn     SafetyAction = "WARN"
	SafetyActionEscalate SafetyAction = "ESCALATE"
)

// SafetyPolicy defines runtime thresholds for a tenant.
// A zero value for any threshold disables that check.
type SafetyPolicy struct {
	// MaxTransactionAmount is the maximum allowed single-transaction total debit.
	// Mutations exceeding this are BLOCKED.
	MaxTransactionAmount decimal.Decimal

	// MaxReversalsPerHour limits reversals per user per hour. BLOCKS on breach.
	MaxReversalsPerHour int

	// MaxApprovalVelocityPerHour limits approvals per user per hour. BLOCKS on breach.
	MaxApprovalVelocityPerHour int

	// MaxPostingsPerHour limits tenant-wide postings per hour. WARNS on breach.
	MaxPostingsPerHour int
}

// DefaultSafetyPolicy returns conservative production defaults.
func DefaultSafetyPolicy() SafetyPolicy {
	return SafetyPolicy{
		MaxTransactionAmount:       decimal.NewFromInt(10_000_000), // 10 M
		MaxReversalsPerHour:        20,
		MaxApprovalVelocityPerHour: 50,
		MaxPostingsPerHour:         500,
	}
}

// SafetyEnforcer evaluates runtime safety policies before finance mutations.
type SafetyEnforcer struct {
	mu            sync.RWMutex
	tenantPolicies map[uuid.UUID]SafetyPolicy
	defaultPolicy  SafetyPolicy
	counters       *safetyCounters
	metrics        metrics.MetricsProvider
}

// NewSafetyEnforcer creates an enforcer with the given default policy.
func NewSafetyEnforcer(defaultPolicy SafetyPolicy, m metrics.MetricsProvider) *SafetyEnforcer {
	return &SafetyEnforcer{
		tenantPolicies: make(map[uuid.UUID]SafetyPolicy),
		defaultPolicy:  defaultPolicy,
		counters:       newSafetyCounters(),
		metrics:        m,
	}
}

// SetTenantPolicy overrides the policy for a specific tenant.
func (e *SafetyEnforcer) SetTenantPolicy(tenantID uuid.UUID, p SafetyPolicy) {
	if e == nil {
		return
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.tenantPolicies[tenantID] = p
}

func (e *SafetyEnforcer) policyFor(tenantID uuid.UUID) SafetyPolicy {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if p, ok := e.tenantPolicies[tenantID]; ok {
		return p
	}
	return e.defaultPolicy
}

// CheckTransactionAmount blocks transactions whose total debit exceeds the configured max.
func (e *SafetyEnforcer) CheckTransactionAmount(ctx context.Context, totalDebit decimal.Decimal) error {
	if e == nil {
		return nil
	}
	tenantID, _ := shared.GetTenantID(ctx)
	p := e.policyFor(tenantID)
	if p.MaxTransactionAmount.IsZero() {
		return nil
	}
	if totalDebit.GreaterThan(p.MaxTransactionAmount) {
		e.metrics.IncrementCounter("finance_safety_violations_total", metrics.Fields{
			"policy": "MAX_TRANSACTION_AMOUNT",
			"action": string(SafetyActionBlock),
		})
		logger.WarnContext(ctx, "finance safety: transaction amount exceeds policy limit — BLOCKED", logger.Fields{
			"amount":    totalDebit.String(),
			"limit":     p.MaxTransactionAmount.String(),
			"tenant_id": tenantID.String(),
		})
		return errors.NewBusinessError("SAFETY_AMOUNT_EXCEEDED",
			fmt.Sprintf("transaction amount %s exceeds policy maximum %s; request secondary approval or raise the limit",
				totalDebit.StringFixed(2), p.MaxTransactionAmount.StringFixed(2))).
			WithHTTPStatus(http.StatusUnprocessableEntity).
			WithCategory(errors.CategoryBusiness)
	}
	return nil
}

// CheckReversalVelocity blocks users who exceed the per-hour reversal limit.
func (e *SafetyEnforcer) CheckReversalVelocity(ctx context.Context, userID uuid.UUID) error {
	if e == nil {
		return nil
	}
	tenantID, _ := shared.GetTenantID(ctx)
	p := e.policyFor(tenantID)
	if p.MaxReversalsPerHour == 0 {
		return nil
	}
	key := safetyCounterKey{tenantID: tenantID, userID: userID, event: "reversal"}
	count := e.counters.increment(key, time.Hour)
	if count > p.MaxReversalsPerHour {
		e.metrics.IncrementCounter("finance_safety_violations_total", metrics.Fields{
			"policy": "MAX_REVERSALS_PER_HOUR",
			"action": string(SafetyActionBlock),
		})
		logger.WarnContext(ctx, "finance safety: reversal velocity limit exceeded — BLOCKED", logger.Fields{
			"user_id":   userID.String(),
			"count":     count,
			"limit":     p.MaxReversalsPerHour,
			"tenant_id": tenantID.String(),
		})
		return errors.NewBusinessError("SAFETY_REVERSAL_VELOCITY",
			fmt.Sprintf("reversal rate limit: %d reversals in the last hour (max %d); possible abuse detected",
				count, p.MaxReversalsPerHour)).
			WithHTTPStatus(http.StatusTooManyRequests).
			WithCategory(errors.CategorySecurity)
	}
	return nil
}

// CheckApprovalVelocity blocks users approving at abnormal rates.
func (e *SafetyEnforcer) CheckApprovalVelocity(ctx context.Context, userID uuid.UUID) error {
	if e == nil {
		return nil
	}
	tenantID, _ := shared.GetTenantID(ctx)
	p := e.policyFor(tenantID)
	if p.MaxApprovalVelocityPerHour == 0 {
		return nil
	}
	key := safetyCounterKey{tenantID: tenantID, userID: userID, event: "approval"}
	count := e.counters.increment(key, time.Hour)
	if count > p.MaxApprovalVelocityPerHour {
		e.metrics.IncrementCounter("finance_safety_violations_total", metrics.Fields{
			"policy": "MAX_APPROVAL_VELOCITY",
			"action": string(SafetyActionBlock),
		})
		logger.WarnContext(ctx, "finance safety: approval velocity limit exceeded — BLOCKED", logger.Fields{
			"user_id":   userID.String(),
			"count":     count,
			"limit":     p.MaxApprovalVelocityPerHour,
			"tenant_id": tenantID.String(),
		})
		return errors.NewBusinessError("SAFETY_APPROVAL_VELOCITY",
			fmt.Sprintf("approval rate limit: %d approvals in the last hour (max %d); possible approval abuse",
				count, p.MaxApprovalVelocityPerHour)).
			WithHTTPStatus(http.StatusTooManyRequests).
			WithCategory(errors.CategorySecurity)
	}
	return nil
}

// CheckPostingVelocity emits a WARN metric when bulk posting exceeds threshold.
// Does not block — returns nil always. Upgrade to BLOCK by returning an error.
func (e *SafetyEnforcer) CheckPostingVelocity(ctx context.Context) {
	if e == nil {
		return
	}
	tenantID, _ := shared.GetTenantID(ctx)
	p := e.policyFor(tenantID)
	if p.MaxPostingsPerHour == 0 {
		return
	}
	key := safetyCounterKey{tenantID: tenantID, event: "posting"}
	count := e.counters.increment(key, time.Hour)
	if count > p.MaxPostingsPerHour {
		e.metrics.IncrementCounter("finance_safety_violations_total", metrics.Fields{
			"policy": "MAX_POSTINGS_PER_HOUR",
			"action": string(SafetyActionWarn),
		})
		logger.WarnContext(ctx, "finance safety: posting velocity anomaly — unusual bulk posting rate", logger.Fields{
			"count":     count,
			"limit":     p.MaxPostingsPerHour,
			"tenant_id": tenantID.String(),
		})
	}
}

// ── sliding-window counters ───────────────────────────────────────────────────

type safetyCounterKey struct {
	tenantID uuid.UUID
	userID   uuid.UUID // zero UUID for tenant-level (non-user) counters
	event    string
}

type windowEntry struct {
	mu         sync.Mutex
	timestamps []time.Time
}

type safetyCounters struct {
	mu      sync.RWMutex
	windows map[safetyCounterKey]*windowEntry
}

func newSafetyCounters() *safetyCounters {
	return &safetyCounters{windows: make(map[safetyCounterKey]*windowEntry)}
}

// increment records a new event and returns the count within the rolling window.
// Thread-safe. Window eviction is O(n) per call — acceptable for low event rates.
func (c *safetyCounters) increment(key safetyCounterKey, window time.Duration) int {
	c.mu.RLock()
	entry, ok := c.windows[key]
	c.mu.RUnlock()
	if !ok {
		c.mu.Lock()
		// Double-checked locking.
		if entry, ok = c.windows[key]; !ok {
			entry = &windowEntry{}
			c.windows[key] = entry
		}
		c.mu.Unlock()
	}

	now := time.Now()
	cutoff := now.Add(-window)

	entry.mu.Lock()
	defer entry.mu.Unlock()

	// Prune expired entries in-place (read-pos always >= write-pos → no aliasing).
	w := 0
	for _, t := range entry.timestamps {
		if t.After(cutoff) {
			entry.timestamps[w] = t
			w++
		}
	}
	entry.timestamps = append(entry.timestamps[:w], now)
	return len(entry.timestamps)
}
