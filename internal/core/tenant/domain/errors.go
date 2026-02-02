package domain

import "errors"

var (
	ErrTenantNotFound    = errors.New("tenant not found")
	ErrSubdomainTaken    = errors.New("subdomain already taken")
	ErrInvalidTransition = errors.New("invalid status transition")
	ErrTenantSuspended   = errors.New("tenant is suspended")
	ErrLimitExceeded     = errors.New("usage limit exceeded")
	ErrInvalidRequest    = errors.New("invalid request")
)
