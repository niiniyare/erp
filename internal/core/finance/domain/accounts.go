package domain

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Accounts represents a single account in the chart of accounts
// This is the core entity for the accounting system's account structure
type Accounts struct {
	ID       uuid.UUID  `json:"id"`
	TenantID uuid.UUID  `json:"tenant_id"`
	EntityID *uuid.UUID `json:"entity_id,omitempty"` // Optional for multi-entity support

	// Account identification
	AccountCode        string  `json:"account_code"`                  // Unique identifier (e.g., "1000")
	AccountName        string  `json:"account_name"`                  // Display name (e.g., "Cash")
	AccountDescription *string `json:"account_description,omitempty"` // Optional detailed description

	// Account hierarchy - supports parent-child relationships
	ParentAccountID *uuid.UUID `json:"parent_account_id,omitempty"` // Reference to parent account
	AccountLevel    int32      `json:"account_level"`               // Depth in hierarchy (0 = root level)
	AccountPath     *string    `json:"account_path,omitempty"`      // Materialized path (e.g., "/1000/1100/")
	HasChildren     bool       `json:"has_children"`                // Whether account has child accounts
	IsLeafAccount   bool       `json:"is_leaf_account"`             // Whether account is a leaf node

	// Account grouping and organization
	AccountGroupID   *uuid.UUID `json:"account_group_id,omitempty"`   // Reference to account group
	AccountHeaderID  *uuid.UUID `json:"account_header_id,omitempty"`  // Reference to account header

	// Account classification following standard accounting taxonomy
	RootType        RootType `json:"root_type"`                  // Primary classification
	AccountType     string   `json:"account_type"`               // Secondary classification
	AccountSubtype  *string  `json:"account_subtype,omitempty"`  // Tertiary classification
	AccountCategory *string  `json:"account_category,omitempty"` // Account category for grouping
	SubCategory     *string  `json:"sub_category,omitempty"`     // Sub-category within main category

	// Financial attributes
	NormalBalance    NormalBalance `json:"normal_balance"`               // Debit or Credit normal balance
	IsControlAccount bool          `json:"is_control_account"`           // Summary account for sub-accounts
	ControlAccountID *uuid.UUID    `json:"control_account_id,omitempty"` // Reference to control account

	// Currency and localization support
	CurrencyCode                *string `json:"currency_code,omitempty"`       // ISO currency code
	IsMultiCurrency             bool    `json:"is_multi_currency"`             // Accepts multiple currencies
	CurrencyRevaluationRequired bool    `json:"currency_revaluation_required"` // Needs periodic revaluation

	// Operational settings
	IsActive           bool `json:"is_active"`            // Account is active for use
	IsSystemAccount    bool `json:"is_system_account"`    // Protected system account
	AllowManualEntries bool `json:"allow_manual_entries"` // Accepts manual journal entries
	RequireReference   bool `json:"require_reference"`    // Mandate reference for entries

	// Balance tracking - cached for performance
	CurrentBalance      decimal.Decimal `json:"current_balance"`                 // Real-time balance
	YTDBalance          decimal.Decimal `json:"ytd_balance"`                     // Year-to-date balance
	LastTransactionDate *time.Time      `json:"last_transaction_date,omitempty"` // Last activity

	// Reporting and analysis
	FinancialStatementLine *string `json:"financial_statement_line,omitempty"` // Report grouping
	ReportOrder            int32   `json:"report_order"`                       // Sort order in reports
	DisplayOrder           int32   `json:"display_order"`                      // Display order in UI/reports
	ShowInReports          bool    `json:"show_in_reports"`                    // Whether to include in standard reports
	ConsolidationAccount   *string `json:"consolidation_account,omitempty"`    // Consolidation mapping for multi-entity
	CashFlowType           *string `json:"cash_flow_type,omitempty"`           // Cash flow statement classification

	// Budgeting capabilities
	IsBudgetable            bool            `json:"is_budgetable"`             // Can have budget allocated
	BudgetVarianceThreshold decimal.Decimal `json:"budget_variance_threshold"` // Alert threshold %

	// Audit and validation
	Version           int32             `json:"version"`                       // Optimistic locking
	LastValidationRun *time.Time        `json:"last_validation_run,omitempty"` // Last validation check
	ValidationStatus  ValidationStatus  `json:"validation_status"`             // Current validation state
	ValidationErrors  []ValidationError `json:"validation_errors,omitempty"`   // Validation issues

	// Metadata - flexible attributes for extensions
	AccountAttributes map[string]any `json:"account_attributes,omitempty"`

	// Standard audit timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"` // Soft delete support
	CreatedBy uuid.UUID  `json:"created_by"`
	UpdatedBy *uuid.UUID `json:"updated_by,omitempty"`
}

// CreateAccountRequest represents the request to create a new account
type CreateAccountRequest struct {
	EntityID                    *uuid.UUID             `json:"entity_id,omitempty"`
	AccountCode                 string                 `json:"account_code" validate:"required,max=20"`
	AccountName                 string                 `json:"account_name" validate:"required,max=200"`
	AccountDescription          *string                `json:"account_description,omitempty" validate:"omitempty,max=1000"`
	
	// Hierarchy and grouping
	ParentAccountID             *uuid.UUID             `json:"parent_account_id,omitempty"`
	AccountGroupID              *uuid.UUID             `json:"account_group_id,omitempty"`
	AccountHeaderID             *uuid.UUID             `json:"account_header_id,omitempty"`
	
	// Classification
	RootType                    RootType               `json:"root_type" validate:"required"`
	AccountType                 string                 `json:"account_type" validate:"required,max=100"`
	AccountSubtype              *string                `json:"account_subtype,omitempty" validate:"omitempty,max=100"`
	AccountCategory             *string                `json:"account_category,omitempty" validate:"omitempty,max=100"`
	SubCategory                 *string                `json:"sub_category,omitempty" validate:"omitempty,max=100"`
	NormalBalance               NormalBalance          `json:"normal_balance" validate:"required"`
	
	// Control and relationships
	IsControlAccount            bool                   `json:"is_control_account"`
	ControlAccountID            *uuid.UUID             `json:"control_account_id,omitempty"`
	
	// Currency settings
	CurrencyCode                *string                `json:"currency_code,omitempty" validate:"omitempty,len=3"`
	IsMultiCurrency             bool                   `json:"is_multi_currency"`
	CurrencyRevaluationRequired bool                   `json:"currency_revaluation_required"`
	
	// Operational settings
	IsActive                    bool                   `json:"is_active"`
	AllowManualEntries          bool                   `json:"allow_manual_entries"`
	RequireReference            bool                   `json:"require_reference"`
	
	// Reporting and display
	FinancialStatementLine      *string                `json:"financial_statement_line,omitempty" validate:"omitempty,max=100"`
	ReportOrder                 int32                  `json:"report_order" validate:"min=0"`
	DisplayOrder                int32                  `json:"display_order" validate:"min=0"`
	ShowInReports               bool                   `json:"show_in_reports"`
	ConsolidationAccount        *string                `json:"consolidation_account,omitempty" validate:"omitempty,max=100"`
	CashFlowType                *string                `json:"cash_flow_type,omitempty" validate:"omitempty,max=50"`
	
	// Budgeting
	IsBudgetable                bool                   `json:"is_budgetable"`
	BudgetVarianceThreshold     decimal.Decimal        `json:"budget_variance_threshold" validate:"min=0,max=100"`
	
	// Extensions
	AccountAttributes           map[string]interface{} `json:"account_attributes,omitempty"`
}

// UpdateAccountRequest represents the request to update an account
type UpdateAccountRequest struct {
	AccountCode             *string                `json:"account_code,omitempty" validate:"omitempty,max=20"`
	AccountName             *string                `json:"account_name,omitempty" validate:"omitempty,max=200"`
	AccountDescription      *string                `json:"account_description,omitempty" validate:"omitempty,max=1000"`
	
	// Grouping and hierarchy (limited updates for data integrity)
	AccountGroupID          *uuid.UUID             `json:"account_group_id,omitempty"`
	AccountHeaderID         *uuid.UUID             `json:"account_header_id,omitempty"`
	
	// Classification updates
	AccountType             *string                `json:"account_type,omitempty" validate:"omitempty,max=100"`
	AccountSubtype          *string                `json:"account_subtype,omitempty" validate:"omitempty,max=100"`
	AccountCategory         *string                `json:"account_category,omitempty" validate:"omitempty,max=100"`
	SubCategory             *string                `json:"sub_category,omitempty" validate:"omitempty,max=100"`
	
	// Operational settings
	IsActive                *bool                  `json:"is_active,omitempty"`
	AllowManualEntries      *bool                  `json:"allow_manual_entries,omitempty"`
	RequireReference        *bool                  `json:"require_reference,omitempty"`
	
	// Reporting and display
	FinancialStatementLine  *string                `json:"financial_statement_line,omitempty" validate:"omitempty,max=100"`
	ReportOrder             *int32                 `json:"report_order,omitempty" validate:"omitempty,min=0"`
	DisplayOrder            *int32                 `json:"display_order,omitempty" validate:"omitempty,min=0"`
	ShowInReports           *bool                  `json:"show_in_reports,omitempty"`
	ConsolidationAccount    *string                `json:"consolidation_account,omitempty" validate:"omitempty,max=100"`
	CashFlowType            *string                `json:"cash_flow_type,omitempty" validate:"omitempty,max=50"`
	
	// Budgeting
	IsBudgetable            *bool                  `json:"is_budgetable,omitempty"`
	BudgetVarianceThreshold *decimal.Decimal       `json:"budget_variance_threshold,omitempty" validate:"omitempty,min=0,max=100"`
	
	// Extensions
	AccountAttributes       map[string]interface{} `json:"account_attributes,omitempty"`
}

// AccountBalance represents account balance information at a point in time
type AccountBalance struct {
	AccountID    uuid.UUID       `json:"account_id"`
	Account      Accounts `json:"account"`
	TotalDebits  decimal.Decimal `json:"total_debits"`  // Sum of all debit entries
	TotalCredits decimal.Decimal `json:"total_credits"` // Sum of all credit entries
	NetBalance   decimal.Decimal `json:"net_balance"`   // Calculated net balance
	AsOfDate     time.Time       `json:"as_of_date"`    // Balance calculation date
}

// TrialBalanceEntry represents a single line in the trial balance report
type TrialBalanceEntry struct {
	Account      Accounts `json:"account"`
	TotalDebits  decimal.Decimal `json:"total_debits"`
	TotalCredits decimal.Decimal `json:"total_credits"`
	NetBalance   decimal.Decimal `json:"net_balance"`
}

// Validate validates the chart of accounts entity
func (c *Accounts) Validate() []ValidationError {
	var errors []ValidationError

	// Validate required fields
	if c.TenantID == uuid.Nil {
		errors = append(errors, ValidationError{
			Field:   "tenant_id",
			Message: "Tenant ID is required",
			Code:    "REQUIRED_FIELD",
		})
	}

	if strings.TrimSpace(c.AccountCode) == "" {
		errors = append(errors, ValidationError{
			Field:   "account_code",
			Message: "Account code is required",
			Code:    "REQUIRED_FIELD",
		})
	} else if len(c.AccountCode) > 20 {
		errors = append(errors, ValidationError{
			Field:   "account_code",
			Message: "Account code must be 20 characters or less",
			Code:    "MAX_LENGTH_EXCEEDED",
		})
	}

	// Validate account code format (alphanumeric with specific patterns)
	if !isValidAccountCode(c.AccountCode) {
		errors = append(errors, ValidationError{
			Field:   "account_code",
			Message: "Account code must contain only alphanumeric characters and hyphens",
			Code:    "INVALID_FORMAT",
		})
	}

	if strings.TrimSpace(c.AccountName) == "" {
		errors = append(errors, ValidationError{
			Field:   "account_name",
			Message: "Account name is required",
			Code:    "REQUIRED_FIELD",
		})
	} else if len(c.AccountName) > 200 {
		errors = append(errors, ValidationError{
			Field:   "account_name",
			Message: "Account name must be 200 characters or less",
			Code:    "MAX_LENGTH_EXCEEDED",
		})
	}

	// Validate root type
	if !c.RootType.IsValid() {
		errors = append(errors, ValidationError{
			Field:   "root_type",
			Message: "Invalid root type",
			Code:    "INVALID_VALUE",
		})
	}

	// Validate normal balance
	if !c.NormalBalance.IsValid() {
		errors = append(errors, ValidationError{
			Field:   "normal_balance",
			Message: "Invalid normal balance",
			Code:    "INVALID_VALUE",
		})
	}

	// Validate normal balance matches root type
	expectedNormalBalance := GetNormalBalanceForRootType(c.RootType)
	if c.NormalBalance != expectedNormalBalance {
		errors = append(errors, ValidationError{
			Field:   "normal_balance",
			Message: fmt.Sprintf("Normal balance should be %s for root type %s", expectedNormalBalance, c.RootType),
			Code:    "INCONSISTENT_VALUE",
		})
	}

	// Validate currency code format if provided
	if c.CurrencyCode != nil && len(*c.CurrencyCode) != 3 {
		errors = append(errors, ValidationError{
			Field:   "currency_code",
			Message: "Currency code must be exactly 3 characters (ISO 4217)",
			Code:    "INVALID_FORMAT",
		})
	}

	// Business rule validations
	if c.IsControlAccount && c.AllowManualEntries {
		errors = append(errors, ValidationError{
			Field:   "allow_manual_entries",
			Message: "Control accounts should not allow manual entries",
			Code:    "BUSINESS_RULE_VIOLATION",
		})
	}

	if c.ControlAccountID != nil && c.IsControlAccount {
		errors = append(errors, ValidationError{
			Field:   "control_account_id",
			Message: "Control accounts cannot have a parent control account",
			Code:    "BUSINESS_RULE_VIOLATION",
		})
	}

	// Validate account level consistency
	if c.ParentAccountID == nil && c.AccountLevel != 0 {
		errors = append(errors, ValidationError{
			Field:   "account_level",
			Message: "Root accounts must have level 0",
			Code:    "INCONSISTENT_VALUE",
		})
	}

	if c.ParentAccountID != nil && c.AccountLevel <= 0 {
		errors = append(errors, ValidationError{
			Field:   "account_level",
			Message: "Child accounts must have level > 0",
			Code:    "INCONSISTENT_VALUE",
		})
	}

	// Validate budget variance threshold
	if c.BudgetVarianceThreshold.LessThan(decimal.Zero) || c.BudgetVarianceThreshold.GreaterThan(decimal.NewFromInt(100)) {
		errors = append(errors, ValidationError{
			Field:   "budget_variance_threshold",
			Message: "Budget variance threshold must be between 0 and 100",
			Code:    "VALUE_OUT_OF_RANGE",
		})
	}

	return errors
}

// isValidAccountCode validates the account code format
func isValidAccountCode(code string) bool {
	// Allow alphanumeric characters and hyphens
	matched, _ := regexp.MatchString(`^[A-Za-z0-9\-]+$`, code)
	return matched
}

// IsDescendantOf checks if this account is a descendant of the specified parent
func (c *Accounts) IsDescendantOf(parentID uuid.UUID) bool {
	if c.AccountPath == nil {
		return false
	}
	parentPath := fmt.Sprintf("/%s/", parentID.String())
	return strings.Contains(*c.AccountPath, parentPath)
}

// GetEffectiveBalance calculates the effective balance considering normal balance type
func (c *Accounts) GetEffectiveBalance() decimal.Decimal {
	if c.NormalBalance == NormalBalanceDebit {
		return c.CurrentBalance
	}
	// For credit normal balance accounts, negate the balance for reporting
	return c.CurrentBalance.Neg()
}

// CanAcceptManualEntries determines if manual entries are allowed
func (c *Accounts) CanAcceptManualEntries() bool {
	return c.IsActive && c.AllowManualEntries && !c.IsControlAccount
}

// CanBeDeactivated checks if the account can be marked as inactive
func (c *Accounts) CanBeDeactivated() bool {
	// Cannot deactivate if it has a non-zero balance
	if !c.CurrentBalance.IsZero() {
		return false
	}

	// Cannot deactivate system accounts
	if c.IsSystemAccount {
		return false
	}

	return true
}

// MarshalAccountAttributes marshals account attributes to JSON
func (c *Accounts) MarshalAccountAttributes() ([]byte, error) {
	if c.AccountAttributes == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(c.AccountAttributes)
}

// UnmarshalAccountAttributes unmarshals account attributes from JSON
func (c *Accounts) UnmarshalAccountAttributes(data []byte) error {
	if len(data) == 0 {
		c.AccountAttributes = make(map[string]any)
		return nil
	}
	return json.Unmarshal(data, &c.AccountAttributes)
}

// Validate validates the create account request
func (r *CreateAccountRequest) Validate() []ValidationError {
	// var errors []ValidationError

	// Create a temporary account for validation
	account := Accounts{
		TenantID:                    uuid.New(), // Dummy value for validation
		EntityID:                    r.EntityID,
		AccountCode:                 r.AccountCode,
		AccountName:                 r.AccountName,
		AccountDescription:          r.AccountDescription,
		ParentAccountID:             r.ParentAccountID,
		AccountGroupID:              r.AccountGroupID,
		AccountHeaderID:             r.AccountHeaderID,
		RootType:                    r.RootType,
		AccountType:                 r.AccountType,
		AccountSubtype:              r.AccountSubtype,
		AccountCategory:             r.AccountCategory,
		SubCategory:                 r.SubCategory,
		NormalBalance:               r.NormalBalance,
		IsControlAccount:            r.IsControlAccount,
		ControlAccountID:            r.ControlAccountID,
		CurrencyCode:                r.CurrencyCode,
		IsMultiCurrency:             r.IsMultiCurrency,
		CurrencyRevaluationRequired: r.CurrencyRevaluationRequired,
		IsActive:                    r.IsActive,
		AllowManualEntries:          r.AllowManualEntries,
		RequireReference:            r.RequireReference,
		FinancialStatementLine:      r.FinancialStatementLine,
		ReportOrder:                 r.ReportOrder,
		DisplayOrder:                r.DisplayOrder,
		ShowInReports:               r.ShowInReports,
		ConsolidationAccount:        r.ConsolidationAccount,
		CashFlowType:                r.CashFlowType,
		IsBudgetable:                r.IsBudgetable,
		BudgetVarianceThreshold:     r.BudgetVarianceThreshold,
		AccountAttributes:           r.AccountAttributes,
	}

	return account.Validate()
}

// Validate validates the update account request
func (r *UpdateAccountRequest) Validate() []ValidationError {
	var errors []ValidationError

	if r.AccountName != nil {
		if strings.TrimSpace(*r.AccountName) == "" {
			errors = append(errors, ValidationError{
				Field:   "account_name",
				Message: "Account name cannot be empty",
				Code:    "REQUIRED_FIELD",
			})
		} else if len(*r.AccountName) > 200 {
			errors = append(errors, ValidationError{
				Field:   "account_name",
				Message: "Account name must be 200 characters or less",
				Code:    "MAX_LENGTH_EXCEEDED",
			})
		}
	}

	if r.BudgetVarianceThreshold != nil {
		if r.BudgetVarianceThreshold.LessThan(decimal.Zero) || r.BudgetVarianceThreshold.GreaterThan(decimal.NewFromInt(100)) {
			errors = append(errors, ValidationError{
				Field:   "budget_variance_threshold",
				Message: "Budget variance threshold must be between 0 and 100",
				Code:    "VALUE_OUT_OF_RANGE",
			})
		}
	}

	return errors
}

// ValidateBusinessRules validates complex business rules that require external data
func (c *Accounts) ValidateBusinessRules(parentAccount *Accounts, hasChildren bool, hasTransactions bool) []ValidationError {
	var errors []ValidationError

	// Parent account validations
	if c.ParentAccountID != nil && parentAccount != nil {
		if parentAccount.RootType != c.RootType {
			errors = append(errors, ValidationError{
				Field:   "parent_account_id",
				Message: "Child account must have the same root type as parent",
				Code:    "INCONSISTENT_ROOT_TYPE",
			})
		}

		if c.AccountLevel != parentAccount.AccountLevel+1 {
			errors = append(errors, ValidationError{
				Field:   "account_level",
				Message: "Account level must be exactly one more than parent level",
				Code:    "INVALID_HIERARCHY_LEVEL",
			})
		}
	}

	// Control account business rules
	if c.IsControlAccount && !hasChildren {
		errors = append(errors, ValidationError{
			Field:   "is_control_account",
			Message: "Control accounts must have child accounts",
			Code:    "CONTROL_ACCOUNT_NO_CHILDREN",
		})
	}

	// Deactivation rules
	if !c.IsActive && hasTransactions && !c.CurrentBalance.IsZero() {
		errors = append(errors, ValidationError{
			Field:   "is_active",
			Message: "Cannot deactivate account with non-zero balance",
			Code:    "CANNOT_DEACTIVATE_WITH_BALANCE",
		})
	}

	return errors
}

// TODO: Add support for account templates and account hierarchies import/export
// TODO: Implement account code generation based on configurable patterns
// TODO: Add support for account aliases for different languages/regions
// TODO: Implement account archival instead of soft delete for better audit trail
// NOTE: Consider adding account category groupings for enhanced reporting
// NOTE: Future enhancement: Add support for account-specific posting rules
