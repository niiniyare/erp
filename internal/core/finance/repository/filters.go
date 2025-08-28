package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/core/finance/domain"
)

// AccountFilter defines filtering options for chart of accounts queries
type AccountFilter struct {
	// Pagination
	Limit  int32 `json:"limit,omitempty"`
	Offset int32 `json:"offset,omitempty"`

	// Entity filtering
	EntityID *uuid.UUID `json:"entity_id,omitempty"`

	// Account classification filters
	RootType    *domain.RootType `json:"root_type,omitempty"`
	AccountType *string          `json:"account_type,omitempty"`

	// Status filters
	IsActive *bool `json:"is_active,omitempty"`

	// Hierarchy filters
	ParentAccountID *uuid.UUID `json:"parent_account_id,omitempty"`
	AccountLevel    *int32     `json:"account_level,omitempty"`

	// Special account types
	IsControlAccount *bool `json:"is_control_account,omitempty"`
	IsSystemAccount  *bool `json:"is_system_account,omitempty"`

	// Search filters
	SearchTerm *string `json:"search_term,omitempty"` // Search in account code/name

	// Currency filters
	CurrencyCode    *string `json:"currency_code,omitempty"`
	IsMultiCurrency *bool   `json:"is_multi_currency,omitempty"`
}

// TransactionFilter defines filtering options for transaction queries
type TransactionFilter struct {
	// Pagination
	Limit  int32 `json:"limit,omitempty"`
	Offset int32 `json:"offset,omitempty"`

	// Entity filtering
	EntityID *uuid.UUID `json:"entity_id,omitempty"`

	// Date range filters
	FromDate *time.Time `json:"from_date,omitempty"`
	ToDate   *time.Time `json:"to_date,omitempty"`

	// Status and type filters
	Status          *domain.TransactionStatus `json:"status,omitempty"`
	TransactionType *domain.TransactionType   `json:"transaction_type,omitempty"`

	// Account filters
	AccountID *uuid.UUID `json:"account_id,omitempty"`

	// Search filters
	SearchTerm      *string `json:"search_term,omitempty"` // Search in description/reference
	ReferenceNumber *string `json:"reference_number,omitempty"`

	// Amount filters
	MinAmount *string `json:"min_amount,omitempty"` // Use string for decimal handling
	MaxAmount *string `json:"max_amount,omitempty"`

	// User filters
	CreatedBy *uuid.UUID `json:"created_by,omitempty"`
	PostedBy  *uuid.UUID `json:"posted_by,omitempty"`
}

// EntryFilter defines filtering options for transaction entry queries
type EntryFilter struct {
	// Pagination
	Limit  int32 `json:"limit,omitempty"`
	Offset int32 `json:"offset,omitempty"`

	// Transaction filters
	TransactionID *uuid.UUID `json:"transaction_id,omitempty"`

	// Account filters
	AccountID *uuid.UUID `json:"account_id,omitempty"`

	// Date range filters
	FromDate *time.Time `json:"from_date,omitempty"`
	ToDate   *time.Time `json:"to_date,omitempty"`

	// Amount filters
	MinAmount *string `json:"min_amount,omitempty"`
	MaxAmount *string `json:"max_amount,omitempty"`

	// Reconciliation filters
	IsReconciled *bool `json:"is_reconciled,omitempty"`

	// Reference filters
	Reference *string `json:"reference,omitempty"`
}

// AuditFilter defines filtering options for audit trail queries
type AuditFilter struct {
	// Pagination
	Limit  int32 `json:"limit,omitempty"`
	Offset int32 `json:"offset,omitempty"`

	// Date range filters
	FromDate *time.Time `json:"from_date,omitempty"`
	ToDate   *time.Time `json:"to_date,omitempty"`

	// Entity filters
	EntityType *string    `json:"entity_type,omitempty"`
	EntityID   *string    `json:"entity_id,omitempty"`
	UserID     *uuid.UUID `json:"user_id,omitempty"`

	// Action filters
	Action *string `json:"action,omitempty"`

	// Search filters
	SearchTerm *string `json:"search_term,omitempty"`
}
