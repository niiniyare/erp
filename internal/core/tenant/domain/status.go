package domain

import "fmt"

// TenantStatus represents tenant lifecycle states.
type TenantStatus string

const (
	StatusActive      TenantStatus = "ACTIVE"
	StatusSuspended   TenantStatus = "SUSPENDED"
	StatusPending     TenantStatus = "PENDING"
	StatusArchived    TenantStatus = "ARCHIVED"
	StatusDeactivated TenantStatus = "DEACTIVATED"
	StatusTrial       TenantStatus = "PENDING" // alias for backward compat
)

// validTransitions defines the allowed state machine transitions.
var validTransitions = map[TenantStatus][]TenantStatus{
	StatusPending:     {StatusActive, StatusArchived},
	StatusActive:      {StatusSuspended, StatusArchived, StatusDeactivated},
	StatusSuspended:   {StatusActive, StatusArchived},
	StatusDeactivated: {StatusArchived},
	StatusArchived:    {}, // terminal state
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

// ParseTenantStatus converts a string to TenantStatus, returning an error for unknown values.
func ParseTenantStatus(s string) (TenantStatus, error) {
	status := TenantStatus(s)
	if !status.Valid() {
		return "", fmt.Errorf("unknown tenant status: %q", s)
	}
	return status, nil
}
