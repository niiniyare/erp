package domain

import "fmt"

// TenantStatus represents tenant lifecycle states.
type TenantStatus string

const (
	StatusActive    TenantStatus = "ACTIVE"
	StatusSuspended TenantStatus = "SUSPENDED"
	StatusPending   TenantStatus = "PENDING"
	StatusArchived  TenantStatus = "ARCHIVED"
	StatusTrial     TenantStatus = "PENDING" // alias for backward compat with middleware
)

// validTransitions defines the allowed state machine transitions.
// Per business rules doc:
//   - PENDING  → ACTIVE only (must activate before any other transition)
//   - ACTIVE   → SUSPENDED, ARCHIVED
//   - SUSPENDED → ACTIVE (reactivation), ARCHIVED
//   - ARCHIVED → (terminal state, no transitions allowed)
var validTransitions = map[TenantStatus][]TenantStatus{
	StatusPending:   {StatusActive},
	StatusActive:    {StatusSuspended, StatusArchived},
	StatusSuspended: {StatusActive, StatusArchived},
	StatusArchived:  {}, // terminal state
}

// Valid returns true if the status is a known value.
func (s TenantStatus) Valid() bool {
	_, ok := validTransitions[s]
	return ok
}

// CanTransitionTo returns true if transitioning from s to target is allowed.
func (s TenantStatus) CanTransitionTo(target TenantStatus) bool {
	allowed, ok := validTransitions[s]
	if !ok {
		return false
	}
	for _, a := range allowed {
		if a == target {
			return true
		}
	}
	return false
}

// String returns string representation.
func (s TenantStatus) String() string {
	return string(s)
}

// ParseTenantStatus converts a string to TenantStatus, returning an error for unknown values.
func ParseTenantStatus(s string) (TenantStatus, error) {
	status := TenantStatus(s)
	if !status.Valid() {
		return "", fmt.Errorf("invalid tenant status: %q", s)
	}
	return status, nil
}
