package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"awo.so/internal/core/finance/domain"
)

// ============================================================================
// FIN-PER-010: AccountingPeriod posting gates (AllowsPosting / IsOpen)
// ============================================================================

func TestAccountingPeriod_CanPost(t *testing.T) {
	// AllowsPosting → true only for OPEN; soft/hard/locked all reject
	tests := []struct {
		status    domain.PeriodStatus
		canPost   bool
		note      string
	}{
		{domain.PeriodStatusOpen, true, "OPEN accepts normal postings"},
		{domain.PeriodStatusSoftClosed, false, "SOFT_CLOSED blocks normal postings (finance role uses AllowsAdjustment)"},
		{domain.PeriodStatusHardClosed, false, "HARD_CLOSED blocks all postings"},
		{domain.PeriodStatusLocked, false, "LOCKED blocks all postings — even CFO cannot post"},
	}

	for _, tc := range tests {
		t.Run(string(tc.status), func(t *testing.T) {
			period := &domain.AccountingPeriod{Status: tc.status}
			assert.Equal(t, tc.canPost, period.IsOpen(), tc.note)
			assert.Equal(t, tc.canPost, tc.status.AllowsPosting(), tc.note)
		})
	}
}

// ============================================================================
// FIN-PER-011: Finance role has a broader posting window (AllowsAdjustment)
// ============================================================================

func TestAccountingPeriod_CanFinancePost(t *testing.T) {
	// AllowsAdjustment is the "finance role" gate — allows posting to SOFT_CLOSED
	tests := []struct {
		status          domain.PeriodStatus
		canAdjust       bool
		note            string
	}{
		{domain.PeriodStatusOpen, true, "OPEN: finance role can always post"},
		{domain.PeriodStatusSoftClosed, true, "SOFT_CLOSED: finance role may still post adjusting entries"},
		{domain.PeriodStatusHardClosed, false, "HARD_CLOSED: nobody can post"},
		{domain.PeriodStatusLocked, false, "LOCKED: nobody can post — read-only forever"},
	}

	for _, tc := range tests {
		t.Run(string(tc.status), func(t *testing.T) {
			assert.Equal(t, tc.canAdjust, tc.status.AllowsAdjustment(), tc.note)
		})
	}
}

// ============================================================================
// PeriodStatus.IsValid — accepts only known values
// ============================================================================

func TestPeriodStatus_IsValid(t *testing.T) {
	valid := []domain.PeriodStatus{
		domain.PeriodStatusOpen,
		domain.PeriodStatusSoftClosed,
		domain.PeriodStatusHardClosed,
		domain.PeriodStatusLocked,
	}
	for _, s := range valid {
		assert.True(t, s.IsValid(), "expected %q to be valid", s)
	}

	invalid := []domain.PeriodStatus{"CLOSED", "ARCHIVED", "PENDING", ""}
	for _, s := range invalid {
		assert.False(t, s.IsValid(), "expected %q to be invalid", s)
	}
}

// ============================================================================
// AccountingPeriod.Contains — date-within-period check
// ============================================================================

func TestAccountingPeriod_Contains(t *testing.T) {
	jan2025 := &domain.AccountingPeriod{
		StartDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC),
		Status:    domain.PeriodStatusOpen,
	}

	assert.True(t, jan2025.Contains(time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)), "mid-month should be inside period")
	assert.True(t, jan2025.Contains(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)), "start date should be inside (inclusive)")
	assert.True(t, jan2025.Contains(time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC)), "end date should be inside (inclusive)")
	assert.False(t, jan2025.Contains(time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)), "first day of next month is outside")
	assert.False(t, jan2025.Contains(time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)), "last day of prior month is outside")
}

// ============================================================================
// AccountingPeriod state machine — CanTransitionTo
// ============================================================================

func TestAccountingPeriod_StateTransitions(t *testing.T) {
	period := &domain.AccountingPeriod{Status: domain.PeriodStatusOpen}

	assert.True(t, period.CanTransitionTo(domain.PeriodStatusSoftClosed), "OPEN → SOFT_CLOSED should be allowed")
	assert.False(t, period.CanTransitionTo(domain.PeriodStatusHardClosed), "OPEN → HARD_CLOSED is not a direct transition")
	assert.False(t, period.CanTransitionTo(domain.PeriodStatusLocked), "OPEN → LOCKED is not a direct transition")

	period.Status = domain.PeriodStatusSoftClosed
	assert.True(t, period.CanTransitionTo(domain.PeriodStatusOpen), "SOFT_CLOSED → OPEN (reopen) should be allowed")
	assert.True(t, period.CanTransitionTo(domain.PeriodStatusHardClosed), "SOFT_CLOSED → HARD_CLOSED should be allowed")

	period.Status = domain.PeriodStatusHardClosed
	assert.True(t, period.CanTransitionTo(domain.PeriodStatusOpen), "HARD_CLOSED → OPEN (reopen) should be allowed")
	assert.True(t, period.CanTransitionTo(domain.PeriodStatusLocked), "HARD_CLOSED → LOCKED should be allowed")

	period.Status = domain.PeriodStatusLocked
	// Locked → Open is intentionally absent; requires out-of-band support escalation
	assert.False(t, period.CanTransitionTo(domain.PeriodStatusOpen), "LOCKED → OPEN must not be a normal transition")
	assert.False(t, period.CanTransitionTo(domain.PeriodStatusSoftClosed), "LOCKED → SOFT_CLOSED must not be allowed")
}
