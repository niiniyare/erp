package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// CreateContractRequest is the input for creating a new contract.
type CreateContractRequest struct {
	EntityID          uuid.UUID       `json:"entity_id"           validate:"required"`
	Title             string          `json:"title"               validate:"required,min=2,max=500"`
	ContractType      string          `json:"contract_type"       validate:"required"`
	CounterpartyName  string          `json:"counterparty_name"   validate:"required,min=2,max=255"`
	CounterpartyEmail *string         `json:"counterparty_email"  validate:"omitempty,email"`
	StartDate         time.Time       `json:"start_date"          validate:"required"`
	EndDate           *time.Time      `json:"end_date"`
	Value             decimal.Decimal `json:"value"`
	CurrencyCode      string          `json:"currency_code"       validate:"omitempty,len=3"`
	Description       *string         `json:"description"`
	Terms             *string         `json:"terms"`
}

// UpdateContractRequest is the input for updating a contract.
// All pointer fields are optional — only non-nil fields are applied (COALESCE in SQL).
type UpdateContractRequest struct {
	Title             *string          `json:"title"              validate:"omitempty,min=2,max=500"`
	ContractType      *string          `json:"contract_type"`
	CounterpartyName  *string          `json:"counterparty_name"  validate:"omitempty,min=2,max=255"`
	CounterpartyEmail *string          `json:"counterparty_email" validate:"omitempty,email"`
	StartDate         *time.Time       `json:"start_date"`
	EndDate           *time.Time       `json:"end_date"`
	Value             *decimal.Decimal `json:"value"`
	CurrencyCode      *string          `json:"currency_code"      validate:"omitempty,len=3"`
	Description       *string          `json:"description"`
	Terms             *string          `json:"terms"`
	// Status is set internally by lifecycle transitions (Submit/Approve/Reject/etc.).
	// Callers should use the dedicated service methods rather than setting this directly.
	Status   *string    `json:"status,omitempty"`
	SignedBy *uuid.UUID `json:"signed_by,omitempty"`
	SignedAt *time.Time `json:"signed_at,omitempty"`
	// Version is required for optimistic locking — must match the current row version.
	Version int32 `json:"version" validate:"required,min=1"`
}

// ContractFilter is used to filter/paginate contract listings.
type ContractFilter struct {
	EntityID     *uuid.UUID
	Status       *string
	ContractType *string
	Search       *string
	Limit        int32
	Offset       int32
	SortBy       string
}
