// Package finance contains the domain models and business logic for the finance module
package finance

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// RootType represents the high-level account classification
type RootType string

const (
	RootTypeAsset     RootType = "ASSET"
	RootTypeLiability RootType = "LIABILITY"
	RootTypeEquity    RootType = "EQUITY"
	RootTypeRevenue   RootType = "REVENUE"
	RootTypeExpense   RootType = "EXPENSE"
)

// NormalBalance represents the normal balance type for accounts
type NormalBalance string

const (
	NormalBalanceDebit  NormalBalance = "DEBIT"
	NormalBalanceCredit NormalBalance = "CREDIT"
)

// TransactionType represents the type of financial transaction
type TransactionType string

const (
	TransactionTypeManual     TransactionType = "MANUAL"
	TransactionTypeSystem     TransactionType = "SYSTEM"
	TransactionTypeImported   TransactionType = "IMPORTED"
	TransactionTypeRecurring  TransactionType = "RECURRING"
	TransactionTypeAdjustment TransactionType = "ADJUSTMENT"
	TransactionTypeClosing    TransactionType = "CLOSING"
)

// TransactionStatus represents the current status of a transaction
type TransactionStatus string

const (
	TransactionStatusDraft           TransactionStatus = "DRAFT"
	TransactionStatusPendingApproval TransactionStatus = "PENDING_APPROVAL"
	TransactionStatusApproved        TransactionStatus = "APPROVED"
	TransactionStatusPosted          TransactionStatus = "POSTED"
	TransactionStatusCancelled       TransactionStatus = "CANCELLED"
	TransactionStatusReversed        TransactionStatus = "REVERSED"
)

// ApprovalStatus represents the approval status of a transaction
type ApprovalStatus string

const (
	ApprovalStatusNotRequired ApprovalStatus = "NOT_REQUIRED"
	ApprovalStatusPending     ApprovalStatus = "PENDING"
	ApprovalStatusApproved    ApprovalStatus = "APPROVED"
	ApprovalStatusRejected    ApprovalStatus = "REJECTED"
)

// ValidationStatus represents the validation status of an entity
type ValidationStatus string

const (
	ValidationStatusPending ValidationStatus = "PENDING"
	ValidationStatusValid   ValidationStatus = "VALID"
	ValidationStatusWarning ValidationStatus = "WARNING"
	ValidationStatusError   ValidationStatus = "ERROR"
)

// ChartOfAccounts represents a single account in the chart of accounts
type ChartOfAccounts struct {
	ID       uuid.UUID  `json:"id"`
	TenantID uuid.UUID  `json:"tenant_id"`
	EntityID *uuid.UUID `json:"entity_id,omitempty"`

	// Account identification
	AccountCode        string  `json:"account_code"`
	AccountName        string  `json:"account_name"`
	AccountDescription *string `json:"account_description,omitempty"`

	// Account hierarchy
	ParentAccountID *uuid.UUID `json:"parent_account_id,omitempty"`
	AccountLevel    int32      `json:"account_level"`
	AccountPath     *string    `json:"account_path,omitempty"`

	// Account classification
	RootType       RootType `json:"root_type"`
	AccountType    string   `json:"account_type"`
	AccountSubtype *string  `json:"account_subtype,omitempty"`

	// Financial attributes
	NormalBalance    NormalBalance `json:"normal_balance"`
	IsControlAccount bool          `json:"is_control_account"`
	ControlAccountID *uuid.UUID    `json:"control_account_id,omitempty"`

	// Currency and localization
	CurrencyCode                *string `json:"currency_code,omitempty"`
	IsMultiCurrency             bool    `json:"is_multi_currency"`
	CurrencyRevaluationRequired bool    `json:"currency_revaluation_required"`

	// Operational settings
	IsActive           bool `json:"is_active"`
	IsSystemAccount    bool `json:"is_system_account"`
	AllowManualEntries bool `json:"allow_manual_entries"`
	RequireReference   bool `json:"require_reference"`

	// Balance tracking
	CurrentBalance      decimal.Decimal `json:"current_balance"`
	YTDBalance          decimal.Decimal `json:"ytd_balance"`
	LastTransactionDate *time.Time      `json:"last_transaction_date,omitempty"`

	// Reporting and analysis
	FinancialStatementLine *string `json:"financial_statement_line,omitempty"`
	ReportOrder            int32   `json:"report_order"`

	// Budgeting
	IsBudgetable            bool            `json:"is_budgetable"`
	BudgetVarianceThreshold decimal.Decimal `json:"budget_variance_threshold"`

	// Audit and validation
	Version           int32             `json:"version"`
	LastValidationRun *time.Time        `json:"last_validation_run,omitempty"`
	ValidationStatus  ValidationStatus  `json:"validation_status"`
	ValidationErrors  []ValidationError `json:"validation_errors,omitempty"`

	// Metadata
	AccountAttributes map[string]any `json:"account_attributes,omitempty"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	CreatedBy uuid.UUID  `json:"created_by"`
	UpdatedBy *uuid.UUID `json:"updated_by,omitempty"`
}

// Transaction represents a financial transaction header
type Transaction struct {
	ID       uuid.UUID  `json:"id"`
	TenantID uuid.UUID  `json:"tenant_id"`
	EntityID *uuid.UUID `json:"entity_id,omitempty"`

	// Transaction identification
	TransactionNumber string            `json:"transaction_number"`
	TransactionType   TransactionType   `json:"transaction_type"`
	TransactionStatus TransactionStatus `json:"transaction_status"`

	// Transaction dates
	TransactionDate time.Time  `json:"transaction_date"`
	PostingDate     *time.Time `json:"posting_date,omitempty"`
	DueDate         *time.Time `json:"due_date,omitempty"`

	// Transaction details
	Description       string  `json:"description"`
	ReferenceNumber   *string `json:"reference_number,omitempty"`
	ExternalReference *string `json:"external_reference,omitempty"`

	// Financial information
	CurrencyCode      string          `json:"currency_code"`
	ExchangeRate      decimal.Decimal `json:"exchange_rate"`
	TotalDebitAmount  decimal.Decimal `json:"total_debit_amount"`
	TotalCreditAmount decimal.Decimal `json:"total_credit_amount"`

	// Source and traceability
	SourceModule       *string    `json:"source_module,omitempty"`
	SourceDocumentType *string    `json:"source_document_type,omitempty"`
	SourceDocumentID   *uuid.UUID `json:"source_document_id,omitempty"`
	BatchID            *uuid.UUID `json:"batch_id,omitempty"`

	// Approval workflow
	ApprovalRequired bool           `json:"approval_required"`
	ApprovalStatus   ApprovalStatus `json:"approval_status"`
	ApprovedBy       *uuid.UUID     `json:"approved_by,omitempty"`
	ApprovedAt       *time.Time     `json:"approved_at,omitempty"`
	ApprovalNotes    *string        `json:"approval_notes,omitempty"`

	// Recurring transaction
	IsRecurring        bool       `json:"is_recurring"`
	RecurringFrequency *string    `json:"recurring_frequency,omitempty"`
	NextRecurringDate  *time.Time `json:"next_recurring_date,omitempty"`

	// Reversal tracking
	IsReversed              bool       `json:"is_reversed"`
	ReversedByTransactionID *uuid.UUID `json:"reversed_by_transaction_id,omitempty"`
	ReversalReason          *string    `json:"reversal_reason,omitempty"`

	// Audit and validation
	Version          int32             `json:"version"`
	ValidationStatus ValidationStatus  `json:"validation_status"`
	ValidationErrors []ValidationError `json:"validation_errors,omitempty"`

	// Metadata
	TransactionAttributes map[string]any `json:"transaction_attributes,omitempty"`

	// Transaction entries
	Entries []TransactionEntry `json:"entries,omitempty"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	CreatedBy uuid.UUID  `json:"created_by"`
	UpdatedBy *uuid.UUID `json:"updated_by,omitempty"`
	PostedBy  *uuid.UUID `json:"posted_by,omitempty"`
	PostedAt  *time.Time `json:"posted_at,omitempty"`
}

// TransactionEntry represents a journal entry line within a transaction
type TransactionEntry struct {
	ID            uuid.UUID `json:"id"`
	TenantID      uuid.UUID `json:"tenant_id"`
	TransactionID uuid.UUID `json:"transaction_id"`
	EntryNumber   int32     `json:"entry_number"`
	AccountID     uuid.UUID `json:"account_id"`

	// Entry amounts
	DebitAmount  decimal.Decimal `json:"debit_amount"`
	CreditAmount decimal.Decimal `json:"credit_amount"`

	// Entry details
	Description string  `json:"description"`
	Reference   *string `json:"reference,omitempty"`

	// Dimensional analysis
	CostCenter *string    `json:"cost_center,omitempty"`
	Department *string    `json:"department,omitempty"`
	ProjectID  *uuid.UUID `json:"project_id,omitempty"`

	// Multi-currency support
	OriginalCurrency *string         `json:"original_currency,omitempty"`
	OriginalAmount   decimal.Decimal `json:"original_amount"`
	ExchangeRate     decimal.Decimal `json:"exchange_rate"`

	// Tax information
	TaxCode   *string         `json:"tax_code,omitempty"`
	TaxRate   decimal.Decimal `json:"tax_rate"`
	TaxAmount decimal.Decimal `json:"tax_amount"`

	// Reconciliation
	Reconciled              bool       `json:"reconciled"`
	ReconciledDate          *time.Time `json:"reconciled_date,omitempty"`
	ReconciliationReference *string    `json:"reconciliation_reference,omitempty"`

	// Related account information (populated in queries)
	Account *ChartOfAccounts `json:"account,omitempty"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// AccountBalance represents account balance information
type AccountBalance struct {
	AccountID    uuid.UUID       `json:"account_id"`
	Account      ChartOfAccounts `json:"account"`
	TotalDebits  decimal.Decimal `json:"total_debits"`
	TotalCredits decimal.Decimal `json:"total_credits"`
	NetBalance   decimal.Decimal `json:"net_balance"`
	AsOfDate     time.Time       `json:"as_of_date"`
}

// TrialBalanceEntry represents a single line in the trial balance
type TrialBalanceEntry struct {
	Account      ChartOfAccounts `json:"account"`
	TotalDebits  decimal.Decimal `json:"total_debits"`
	TotalCredits decimal.Decimal `json:"total_credits"`
	NetBalance   decimal.Decimal `json:"net_balance"`
}

// TransactionSummary represents aggregated transaction data
type TransactionSummary struct {
	TransactionType   TransactionType   `json:"transaction_type"`
	TransactionStatus TransactionStatus `json:"transaction_status"`
	TransactionCount  int64             `json:"transaction_count"`
	TotalDebit        decimal.Decimal   `json:"total_debit"`
	TotalCredit       decimal.Decimal   `json:"total_credit"`
}

// CreateAccountRequest represents the request to create a new account
type CreateAccountRequest struct {
	EntityID                    *uuid.UUID             `json:"entity_id,omitempty"`
	AccountCode                 string                 `json:"account_code"`
	AccountName                 string                 `json:"account_name"`
	AccountDescription          *string                `json:"account_description,omitempty"`
	ParentAccountID             *uuid.UUID             `json:"parent_account_id,omitempty"`
	RootType                    RootType               `json:"root_type"`
	AccountType                 string                 `json:"account_type"`
	AccountSubtype              *string                `json:"account_subtype,omitempty"`
	NormalBalance               NormalBalance          `json:"normal_balance"`
	IsControlAccount            bool                   `json:"is_control_account"`
	ControlAccountID            *uuid.UUID             `json:"control_account_id,omitempty"`
	CurrencyCode                *string                `json:"currency_code,omitempty"`
	IsMultiCurrency             bool                   `json:"is_multi_currency"`
	CurrencyRevaluationRequired bool                   `json:"currency_revaluation_required"`
	IsActive                    bool                   `json:"is_active"`
	AllowManualEntries          bool                   `json:"allow_manual_entries"`
	RequireReference            bool                   `json:"require_reference"`
	FinancialStatementLine      *string                `json:"financial_statement_line,omitempty"`
	ReportOrder                 int32                  `json:"report_order"`
	IsBudgetable                bool                   `json:"is_budgetable"`
	BudgetVarianceThreshold     decimal.Decimal        `json:"budget_variance_threshold"`
	AccountAttributes           map[string]interface{} `json:"account_attributes,omitempty"`
}

// UpdateAccountRequest represents the request to update an account
type UpdateAccountRequest struct {
	AccountName             *string                `json:"account_name,omitempty"`
	AccountDescription      *string                `json:"account_description,omitempty"`
	AccountType             *string                `json:"account_type,omitempty"`
	AccountSubtype          *string                `json:"account_subtype,omitempty"`
	IsActive                *bool                  `json:"is_active,omitempty"`
	AllowManualEntries      *bool                  `json:"allow_manual_entries,omitempty"`
	RequireReference        *bool                  `json:"require_reference,omitempty"`
	FinancialStatementLine  *string                `json:"financial_statement_line,omitempty"`
	ReportOrder             *int32                 `json:"report_order,omitempty"`
	IsBudgetable            *bool                  `json:"is_budgetable,omitempty"`
	BudgetVarianceThreshold *decimal.Decimal       `json:"budget_variance_threshold,omitempty"`
	AccountAttributes       map[string]interface{} `json:"account_attributes,omitempty"`
}

// CreateTransactionRequest represents the request to create a new transaction
type CreateTransactionRequest struct {
	EntityID              *uuid.UUID           `json:"entity_id,omitempty"`
	TransactionNumber     string               `json:"transaction_number"`
	TransactionType       TransactionType      `json:"transaction_type"`
	TransactionDate       time.Time            `json:"transaction_date"`
	PostingDate           *time.Time           `json:"posting_date,omitempty"`
	DueDate               *time.Time           `json:"due_date,omitempty"`
	Description           string               `json:"description"`
	ReferenceNumber       *string              `json:"reference_number,omitempty"`
	ExternalReference     *string              `json:"external_reference,omitempty"`
	CurrencyCode          string               `json:"currency_code"`
	ExchangeRate          decimal.Decimal      `json:"exchange_rate"`
	SourceModule          *string              `json:"source_module,omitempty"`
	SourceDocumentType    *string              `json:"source_document_type,omitempty"`
	SourceDocumentID      *uuid.UUID           `json:"source_document_id,omitempty"`
	BatchID               *uuid.UUID           `json:"batch_id,omitempty"`
	ApprovalRequired      bool                 `json:"approval_required"`
	IsRecurring           bool                 `json:"is_recurring"`
	RecurringFrequency    *string              `json:"recurring_frequency,omitempty"`
	NextRecurringDate     *time.Time           `json:"next_recurring_date,omitempty"`
	TransactionAttributes map[string]any       `json:"transaction_attributes,omitempty"`
	Entries               []CreateEntryRequest `json:"entries"`
}

// CreateEntryRequest represents the request to create a transaction entry
type CreateEntryRequest struct {
	AccountID        uuid.UUID       `json:"account_id"`
	DebitAmount      decimal.Decimal `json:"debit_amount"`
	CreditAmount     decimal.Decimal `json:"credit_amount"`
	Description      string          `json:"description"`
	Reference        *string         `json:"reference,omitempty"`
	CostCenter       *string         `json:"cost_center,omitempty"`
	Department       *string         `json:"department,omitempty"`
	ProjectID        *uuid.UUID      `json:"project_id,omitempty"`
	OriginalCurrency *string         `json:"original_currency,omitempty"`
	OriginalAmount   decimal.Decimal `json:"original_amount"`
	ExchangeRate     decimal.Decimal `json:"exchange_rate"`
	TaxCode          *string         `json:"tax_code,omitempty"`
	TaxRate          decimal.Decimal `json:"tax_rate"`
	TaxAmount        decimal.Decimal `json:"tax_amount"`
}

// IsBalanced checks if the transaction debits equal credits
func (t *Transaction) IsBalanced() bool {
	return t.TotalDebitAmount.Equal(t.TotalCreditAmount)
}

// CanBePosted checks if a transaction can be posted
func (t *Transaction) CanBePosted() bool {
	if !t.IsBalanced() {
		return false
	}

	switch t.TransactionStatus {
	case TransactionStatusApproved, TransactionStatusDraft:
		if t.ApprovalRequired && t.ApprovalStatus != ApprovalStatusApproved {
			return false
		}
		return true
	default:
		return false
	}
}

// CanBeReversed checks if a transaction can be reversed
func (t *Transaction) CanBeReversed() bool {
	return t.TransactionStatus == TransactionStatusPosted && !t.IsReversed
}

// GetNormalBalanceForRootType returns the normal balance for a root type
func GetNormalBalanceForRootType(rootType RootType) NormalBalance {
	switch rootType {
	case RootTypeAsset, RootTypeExpense:
		return NormalBalanceDebit
	case RootTypeLiability, RootTypeEquity, RootTypeRevenue:
		return NormalBalanceCredit
	default:
		return NormalBalanceDebit
	}
}

// MarshalAccountAttributes marshals account attributes to JSON
func (c *ChartOfAccounts) MarshalAccountAttributes() ([]byte, error) {
	if c.AccountAttributes == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(c.AccountAttributes)
}

// UnmarshalAccountAttributes unmarshals account attributes from JSON
func (c *ChartOfAccounts) UnmarshalAccountAttributes(data []byte) error {
	if len(data) == 0 {
		c.AccountAttributes = make(map[string]any)
		return nil
	}
	return json.Unmarshal(data, &c.AccountAttributes)
}

// MarshalTransactionAttributes marshals transaction attributes to JSON
func (t *Transaction) MarshalTransactionAttributes() ([]byte, error) {
	if t.TransactionAttributes == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(t.TransactionAttributes)
}

// UnmarshalTransactionAttributes unmarshals transaction attributes from JSON
func (t *Transaction) UnmarshalTransactionAttributes(data []byte) error {
	if len(data) == 0 {
		t.TransactionAttributes = make(map[string]any)
		return nil
	}
	return json.Unmarshal(data, &t.TransactionAttributes)
}

