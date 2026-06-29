package domain

import "errors"

// Validation errors.
var (
	ErrTitleRequired        = errors.New("contract title is required")
	ErrNumberRequired       = errors.New("contract number is required")
	ErrCounterpartyRequired = errors.New("counterparty name is required")
	ErrStartDateRequired    = errors.New("contract start date is required")
	ErrEndDateBeforeStart   = errors.New("end date must be after start date")
	ErrInvalidContractType  = errors.New("invalid contract type")
	ErrInvalidStatus        = errors.New("invalid contract status")
	ErrInvalidCurrency      = errors.New("currency code must be 3 characters")
	ErrNegativeValue        = errors.New("contract value cannot be negative")
	ErrInvalidRequest       = errors.New("invalid request")
)

// State & lifecycle errors.
var (
	ErrContractNotFound       = errors.New("contract not found")
	ErrContractAlreadyExists  = errors.New("contract number already exists")
	ErrInvalidTransition      = errors.New("invalid status transition")
	ErrAlreadyActive          = errors.New("contract is already active")
	ErrAlreadyExpired         = errors.New("contract is already expired")
	ErrAlreadyTerminated      = errors.New("contract is already terminated")
	ErrVersionConflict        = errors.New("contract was modified by another request (version conflict)")
	ErrCannotModifyTerminated = errors.New("terminated contract cannot be modified")
	ErrCannotModifyExpired    = errors.New("expired contract cannot be modified")
)
