package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// PeriodStatus tracks the lifecycle of an accounting period.
// Values are lowercase to match the database CHECK constraint.
type PeriodStatus string

const (
	PeriodStatusOpen       PeriodStatus = "open"        // Accepting transactions
	PeriodStatusSoftClosed PeriodStatus = "soft_closed" // Finance managers only
	PeriodStatusHardClosed PeriodStatus = "hard_closed" // All checklist items passed; nobody can post
	PeriodStatusLocked     PeriodStatus = "locked"      // Audit/year-end lock; read-only, terminal
)

// IsValid returns true if the PeriodStatus is a recognised value.
func (ps PeriodStatus) IsValid() bool {
	switch ps {
	case PeriodStatusOpen, PeriodStatusSoftClosed, PeriodStatusHardClosed, PeriodStatusLocked:
		return true
	default:
		return false
	}
}

// String returns the string representation of PeriodStatus.
func (ps PeriodStatus) String() string { return string(ps) }

// AllowsPosting returns true when normal users can post to this period.
// SoftClosed is handled at the application layer (role check); the DB trigger
// blocks only hard_closed and locked.
func (ps PeriodStatus) AllowsPosting() bool {
	return ps == PeriodStatusOpen
}

// AllowsAdjustment returns true when finance-manager-level adjustments are permitted.
func (ps PeriodStatus) AllowsAdjustment() bool {
	return ps == PeriodStatusOpen || ps == PeriodStatusSoftClosed
}

// PeriodTransition defines a valid state transition for a period.
type PeriodTransition struct {
	From               PeriodStatus
	To                 PeriodStatus
	RequiredPermission string
}

// AllowedPeriodTransitions is the complete state machine for accounting periods.
var AllowedPeriodTransitions = []PeriodTransition{
	{PeriodStatusOpen, PeriodStatusSoftClosed, "finance.periods.soft_close"},
	{PeriodStatusSoftClosed, PeriodStatusOpen, "finance.periods.reopen"},
	{PeriodStatusSoftClosed, PeriodStatusHardClosed, "finance.periods.hard_close"},
	{PeriodStatusHardClosed, PeriodStatusOpen, "finance.periods.reopen"},
	{PeriodStatusHardClosed, PeriodStatusLocked, "finance.periods.lock"},
	// Locked → Open is intentionally absent; requires out-of-band support escalation.
}

// FiscalYear represents a tenant's fiscal year configuration.
type FiscalYear struct {
	ID       uuid.UUID `json:"id"`
	TenantID uuid.UUID `json:"tenant_id"`

	Name      string    `json:"name"`       // e.g. "FY2025"
	Year      int       `json:"year"`       // Calendar year the FY starts in (derived from start_date)
	StartDate time.Time `json:"start_date"` // Inclusive start of the fiscal year
	EndDate   time.Time `json:"end_date"`   // Inclusive end of the fiscal year
	IsClosed  bool      `json:"is_closed"`  // All periods hard_closed or locked
	IsLocked  bool      `json:"is_locked"`  // No transactions whatsoever; requires CFO+CEO to unlock

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	CreatedBy uuid.UUID  `json:"created_by"`
	UpdatedBy *uuid.UUID `json:"updated_by,omitempty"`
}

// AccountingPeriod represents a single month (or custom sub-period) within a FiscalYear.
type AccountingPeriod struct {
	ID           uuid.UUID `json:"id"`
	TenantID     uuid.UUID `json:"tenant_id"`
	FiscalYearID uuid.UUID `json:"fiscal_year_id"`

	PeriodNumber int          `json:"period_number"` // 1-based index within the fiscal year
	Name         string       `json:"name"`          // e.g. "January 2025", "Period 1"
	StartDate    time.Time    `json:"start_date"`
	EndDate      time.Time    `json:"end_date"`
	Status       PeriodStatus `json:"status"`

	// Closing metadata
	ClosedAt  *time.Time `json:"closed_at,omitempty"`
	ClosedBy  *uuid.UUID `json:"closed_by,omitempty"`
	LockedAt  *time.Time `json:"locked_at,omitempty"`
	LockedBy  *uuid.UUID `json:"locked_by,omitempty"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	CreatedBy uuid.UUID  `json:"created_by"`
	UpdatedBy *uuid.UUID `json:"updated_by,omitempty"`
}

// IsOpen returns true when new transactions may be posted to this period.
func (p *AccountingPeriod) IsOpen() bool {
	return p.Status.AllowsPosting()
}

// Contains reports whether the given date falls within this period (inclusive).
func (p *AccountingPeriod) Contains(date time.Time) bool {
	d := date.Truncate(24 * time.Hour)
	start := p.StartDate.Truncate(24 * time.Hour)
	end := p.EndDate.Truncate(24 * time.Hour)
	return !d.Before(start) && !d.After(end)
}

// CanTransitionTo reports whether the period may move to newStatus.
func (p *AccountingPeriod) CanTransitionTo(newStatus PeriodStatus) bool {
	for _, t := range AllowedPeriodTransitions {
		if t.From == p.Status && t.To == newStatus {
			return true
		}
	}
	return false
}

// TransitionTo attempts to move the period to newStatus, returning the matching
// PeriodTransition or an error if the move is not permitted.
func (p *AccountingPeriod) TransitionTo(newStatus PeriodStatus) (PeriodTransition, error) {
	for _, t := range AllowedPeriodTransitions {
		if t.From == p.Status && t.To == newStatus {
			return t, nil
		}
	}
	return PeriodTransition{}, fmt.Errorf("period transition from %s to %s is not allowed", p.Status, newStatus)
}

// SoftClose transitions Open → SoftClosed. Only finance_manager/cfo should call this.
func (p *AccountingPeriod) SoftClose(by uuid.UUID, now time.Time) error {
	if p.Status != PeriodStatusOpen {
		return fmt.Errorf("period transition from %s to %s is not allowed", p.Status, PeriodStatusSoftClosed)
	}
	p.Status = PeriodStatusSoftClosed
	p.ClosedAt = &now
	p.ClosedBy = &by
	return nil
}

// HardClose transitions SoftClosed → HardClosed. Requires all close checklist items to pass.
func (p *AccountingPeriod) HardClose(by uuid.UUID, now time.Time, checksPassed bool) error {
	if p.Status != PeriodStatusSoftClosed {
		return fmt.Errorf("period must be soft_closed before hard-closing (current: %s)", p.Status)
	}
	if !checksPassed {
		return fmt.Errorf("period close checklist has unresolved blocking items")
	}
	p.Status = PeriodStatusHardClosed
	p.ClosedAt = &now
	p.ClosedBy = &by
	return nil
}

// Reopen transitions SoftClosed or HardClosed → Open. Locked periods cannot be reopened.
func (p *AccountingPeriod) Reopen(by uuid.UUID, now time.Time) error {
	if p.Status == PeriodStatusLocked {
		return fmt.Errorf("locked period cannot be reopened via normal workflow")
	}
	if p.Status != PeriodStatusSoftClosed && p.Status != PeriodStatusHardClosed {
		return fmt.Errorf("period is not closed (current: %s)", p.Status)
	}
	p.Status = PeriodStatusOpen
	p.ClosedAt = nil
	p.ClosedBy = nil
	return nil
}

// Lock transitions HardClosed → Locked. CFO-only operation.
func (p *AccountingPeriod) Lock(by uuid.UUID, now time.Time) error {
	if p.Status != PeriodStatusHardClosed {
		return fmt.Errorf("period must be hard_closed before locking (current: %s)", p.Status)
	}
	p.Status = PeriodStatusLocked
	p.LockedAt = &now
	p.LockedBy = &by
	return nil
}

// Validate returns a slice of ValidationErrors for the AccountingPeriod.
func (p *AccountingPeriod) Validate() []ValidationError {
	var errs []ValidationError

	if p.TenantID == uuid.Nil {
		errs = append(errs, ValidationError{Field: "tenant_id", Message: "tenant_id is required", Code: "REQUIRED_FIELD", Severity: ValidationSeverityError})
	}
	if p.FiscalYearID == uuid.Nil {
		errs = append(errs, ValidationError{Field: "fiscal_year_id", Message: "fiscal_year_id is required", Code: "REQUIRED_FIELD", Severity: ValidationSeverityError})
	}
	if p.Name == "" {
		errs = append(errs, ValidationError{Field: "name", Message: "period name is required", Code: "REQUIRED_FIELD", Severity: ValidationSeverityError})
	}
	if !p.EndDate.After(p.StartDate) {
		errs = append(errs, ValidationError{Field: "end_date", Message: "end_date must be after start_date", Code: "INVALID_DATE_RANGE", Severity: ValidationSeverityError})
	}
	if p.PeriodNumber < 1 {
		errs = append(errs, ValidationError{Field: "period_number", Message: "period_number must be ≥ 1", Code: "VALUE_OUT_OF_RANGE", Severity: ValidationSeverityError})
	}
	if !p.Status.IsValid() {
		errs = append(errs, ValidationError{Field: "status", Message: "invalid period status", Code: "INVALID_VALUE", Severity: ValidationSeverityError})
	}

	return errs
}
