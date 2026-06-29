package domain

import "errors"

// Validation errors.
var (
	ErrTenantNameRequired      = errors.New("tenant name is required")
	ErrTenantEmailRequired     = errors.New("tenant email is required")
	ErrInvalidEmail            = errors.New("invalid email format")
	ErrInvalidSubdomain        = errors.New("invalid subdomain format")
	ErrInvalidCompanySize      = errors.New("invalid company size")
	ErrInvalidAccountingMethod = errors.New("accounting method must be ACCRUAL or CASH")
	ErrInvalidRequest          = errors.New("invalid request")
)

// State & lifecycle errors.
var (
	ErrTenantNotFound               = errors.New("tenant not found")
	ErrTenantAlreadyExists          = errors.New("tenant already exists")
	ErrSubdomainTaken               = errors.New("subdomain already taken")
	ErrInvalidTransition            = errors.New("invalid status transition")
	ErrTenantSuspended              = errors.New("tenant is suspended")
	ErrAlreadyActive                = errors.New("tenant is already active")
	ErrAlreadySuspended             = errors.New("tenant is already suspended")
	ErrAlreadyArchived              = errors.New("tenant is already archived")
	ErrCannotActivateArchivedTenant = errors.New("cannot activate archived tenant")
	ErrCannotSuspendArchivedTenant  = errors.New("cannot suspend archived tenant")
)

// Resource limit errors.
var (
	ErrLimitExceeded            = errors.New("usage limit exceeded")
	ErrUserLimitExceeded        = errors.New("user limit exceeded")
	ErrEntityLimitExceeded      = errors.New("entity limit exceeded")
	ErrTransactionLimitExceeded = errors.New("transaction limit exceeded")
	ErrStorageLimitExceeded     = errors.New("storage limit exceeded")
)

// Configuration errors.
var (
	ErrConfigurationNotFound = errors.New("tenant configuration not found")
	ErrInvalidConfiguration  = errors.New("invalid tenant configuration")
)
