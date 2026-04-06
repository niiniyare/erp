package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// PeriodStatus tracks the lifecycle of an accounting period.
type PeriodStatus string

const (
	PeriodStatusOpen       PeriodStatus = "OPEN"        // Accepting transactions
	PeriodStatusSoftClosed PeriodStatus = "SOFT_CLOSED" // Closed to normal users; adjustments by finance managers only
	PeriodStatusClosed     PeriodStatus = "CLOSED"      // Permanently closed; no further postings
	PeriodStatusLocked     PeriodStatus = "LOCKED"      // Audit/year-end lock; read-only
)

// IsValid returns true if the PeriodStatus is a recognised value.
func (ps PeriodStatus) IsValid() bool {
	switch ps {
	case PeriodStatusOpen, PeriodStatusSoftClosed, PeriodStatusClosed, PeriodStatusLocked:
		return true
	default:
		return false
	}
}

// String returns the string representation of PeriodStatus.
func (ps PeriodStatus) String() string { return string(ps) }

// AllowsPosting returns true when transactions can be posted to this period.
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
	{PeriodStatusSoftClosed, PeriodStatusClosed, "finance.periods.close"},
	{PeriodStatusClosed, PeriodStatusLocked, "finance.periods.lock"},
	// Closed → Open is intentionally absent; only locked periods can be appealed via support.
}

// FiscalYear represents a tenant's fiscal year configuration.
type FiscalYear struct {
	ID       uuid.UUID `json:"id"`
	TenantID uuid.UUID `json:"tenant_id"`

	Year      int       `json:"year"`       // Calendar year the FY starts in (e.g. 2025)
	StartDate time.Time `json:"start_date"` // Inclusive start of the fiscal year
	EndDate   time.Time `json:"end_date"`   // Inclusive end of the fiscal year
	IsClosed  bool      `json:"is_closed"`  // Whether the year-end close has been run

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
