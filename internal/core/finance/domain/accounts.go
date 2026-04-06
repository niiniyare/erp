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
	AccountGroupID  *uuid.UUID `json:"account_group_id,omitempty"`  // Reference to account group
	AccountHeaderID *uuid.UUID `json:"account_header_id,omitempty"` // Reference to account header

	// Account classification following standard accounting taxonomy
	RootType        RootType      `json:"root_type"`                  // Primary classification
	AccountType     string        `json:"account_type"`               // Secondary classification
	AccountSubtype  *string       `json:"account_subtype,omitempty"`  // Tertiary classification
	AccountCategory *string       `json:"account_category,omitempty"` // Account category for grouping
	SubCategory     *string       `json:"sub_category,omitempty"`     // Sub-category within main category
	Status          AccountStatus `json:"account_status"`             // AccountState which can be based on the other tAttributes

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

// StatusTransition defines a valid state transition between account statuses
type StatusTransition struct {
	From               AccountStatus // Starting status for the transition
	To                 AccountStatus // Target status for the transition
	RequiredPermission string        // Permission required to perform this transition
	ValidationRequired bool          // Whether validation is required before transition
}

// AllowedTransitions defines all permitted status transitions for accounts.
// Each row encodes the business rule: who can trigger the transition and whether
// the account must pass a full validation run before the move is allowed.
var AllowedTransitions = []StatusTransition{
	// ── Initial lifecycle ───────────────────────────────────────────────────────
	// DRAFT → PENDING_APPROVAL: submit for approval
	{AccountStatusDraft, AccountStatusPendingApproval, "finance.accounts.submit", true},
	// DRAFT → ACTIVE: direct activation (trusted/system workflows, e.g. seed scripts)
	{AccountStatusDraft, AccountStatusActive, "finance.accounts.activate", true},

	// ── Approval flow ───────────────────────────────────────────────────────────
	// PENDING_APPROVAL → ACTIVE: approver accepts
	{AccountStatusPendingApproval, AccountStatusActive, "finance.accounts.approve", true},
	// PENDING_APPROVAL → DRAFT: approver rejects; submitter must revise
	{AccountStatusPendingApproval, AccountStatusDraft, "finance.accounts.reject", false},

	// ── Normal operational transitions ──────────────────────────────────────────
	// ACTIVE → INACTIVE: voluntary deactivation (balance must be zero)
	{AccountStatusActive, AccountStatusInactive, "finance.accounts.deactivate", false},
	// ACTIVE → SUSPENDED: administrative suspension (e.g. compliance hold)
	{AccountStatusActive, AccountStatusSuspended, "finance.accounts.suspend", false},
	// ACTIVE → RESTRICTED: limit to read-only / approval-only entries
	{AccountStatusActive, AccountStatusRestricted, "finance.accounts.restrict", false},
	// ACTIVE → FROZEN: hard freeze — no transactions at all
	{AccountStatusActive, AccountStatusFrozen, "finance.accounts.freeze", false},
	// ACTIVE → UNDER_REVIEW: flag for internal audit or compliance review
	{AccountStatusActive, AccountStatusUnderReview, "finance.accounts.review", false},
	// ACTIVE → YEAR_END_PROCESSING: lock for year-end close; only closing entries
	{AccountStatusActive, AccountStatusYearEndProcessing, "finance.accounts.year_end", true},
	// ACTIVE → AUDIT_LOCK: external audit in progress; read-only
	{AccountStatusActive, AccountStatusAuditLock, "finance.accounts.audit_lock", false},
	// ACTIVE → COMPLIANCE_HOLD: regulatory / sanctions hold
	{AccountStatusActive, AccountStatusComplianceHold, "finance.accounts.compliance_hold", false},
	// ACTIVE → SYSTEM_MAINTENANCE: brief technical window (e.g. migration)
	{AccountStatusActive, AccountStatusSystemMaintenance, "finance.accounts.maintenance", false},
	// ACTIVE → DATA_ERROR: automated detection of data integrity issue
	{AccountStatusActive, AccountStatusDataError, "finance.accounts.flag_error", false},

	// ── Releasing administrative holds ──────────────────────────────────────────
	// SUSPENDED → ACTIVE: lift suspension after resolution
	{AccountStatusSuspended, AccountStatusActive, "finance.accounts.activate", true},
	// SUSPENDED → INACTIVE: deactivate without restoring (avoids re-entry)
	{AccountStatusSuspended, AccountStatusInactive, "finance.accounts.deactivate", false},
	// RESTRICTED → ACTIVE: lift restriction
	{AccountStatusRestricted, AccountStatusActive, "finance.accounts.activate", true},
	// RESTRICTED → FROZEN: escalate restriction to hard freeze
	{AccountStatusRestricted, AccountStatusFrozen, "finance.accounts.freeze", false},
	// FROZEN → ACTIVE: unfreeze
	{AccountStatusFrozen, AccountStatusActive, "finance.accounts.activate", true},
	// FROZEN → RESTRICTED: partially unfreeze (back to restricted mode)
	{AccountStatusFrozen, AccountStatusRestricted, "finance.accounts.restrict", false},
	// UNDER_REVIEW → ACTIVE: review completed, no issues found
	{AccountStatusUnderReview, AccountStatusActive, "finance.accounts.activate", true},
	// UNDER_REVIEW → RESTRICTED: review found issues; restrict pending resolution
	{AccountStatusUnderReview, AccountStatusRestricted, "finance.accounts.restrict", false},
	// UNDER_REVIEW → FROZEN: review found serious issues
	{AccountStatusUnderReview, AccountStatusFrozen, "finance.accounts.freeze", false},
	// YEAR_END_PROCESSING → ACTIVE: year-end close completed
	{AccountStatusYearEndProcessing, AccountStatusActive, "finance.accounts.activate", true},
	// AUDIT_LOCK → ACTIVE: audit completed
	{AccountStatusAuditLock, AccountStatusActive, "finance.accounts.activate", true},
	// COMPLIANCE_HOLD → ACTIVE: hold lifted by compliance officer
	{AccountStatusComplianceHold, AccountStatusActive, "finance.accounts.activate", true},
	// SYSTEM_MAINTENANCE → ACTIVE: maintenance window finished
	{AccountStatusSystemMaintenance, AccountStatusActive, "finance.accounts.activate", false},
	// DATA_ERROR → ACTIVE: data corrected and validated
	{AccountStatusDataError, AccountStatusActive, "finance.accounts.activate", true},
	// DATA_ERROR → DRAFT: reset for re-submission after data correction
	{AccountStatusDataError, AccountStatusDraft, "finance.accounts.reset", true},

	// ── Terminal transitions ─────────────────────────────────────────────────────
	// INACTIVE → ACTIVE: reactivate a dormant account
	{AccountStatusInactive, AccountStatusActive, "finance.accounts.activate", true},
	// INACTIVE → SUSPENDED: re-suspend before final closure
	{AccountStatusInactive, AccountStatusSuspended, "finance.accounts.suspend", false},
	// INACTIVE → CLOSED: permanently close (balance must be zero)
	{AccountStatusInactive, AccountStatusClosed, "finance.accounts.close", true},
	// CLOSED → ARCHIVED: move to long-term storage
	{AccountStatusClosed, AccountStatusArchived, "finance.accounts.archive", false},
}

// CanTransitionTo reports whether a transition from the account's current status to
// newStatus is defined in AllowedTransitions.
func (a *Accounts) CanTransitionTo(newStatus AccountStatus) bool {
	for _, t := range AllowedTransitions {
		if t.From == a.Status && t.To == newStatus {
			return true
		}
	}
	return false
}

// TransitionTo attempts to move the account to newStatus.
// It returns the matching StatusTransition (for the caller to enforce the required
// permission) or an error if the transition is not permitted.
func (a *Accounts) TransitionTo(newStatus AccountStatus) (StatusTransition, error) {
	for _, t := range AllowedTransitions {
		if t.From == a.Status && t.To == newStatus {
			return t, nil
		}
	}
	return StatusTransition{}, fmt.Errorf("transition from %s to %s is not allowed", a.Status, newStatus)
}

// CanAcceptTransactions determines if the account can receive financial transactions
func (a *Accounts) CanAcceptTransactions() bool {
	return a.Status == AccountStatusActive &&
		a.IsActive &&
		a.AllowManualEntries &&
		a.DeletedAt == nil &&
		a.ValidationStatus == ValidationStatusValid
}

// CanBeUsedInReports determines if the account should be included in financial reports
func (a *Accounts) CanBeUsedInReports() bool {
	return a.ShowInReports &&
		a.DeletedAt == nil &&
		a.Status != AccountStatusDraft
}

// RequiresApproval checks if the account is in a state requiring administrative approval
func (a *Accounts) RequiresApproval() bool {
	return a.Status == AccountStatusPendingApproval ||
		a.Status == AccountStatusUnderReview
}

// GetTransactionRestrictions returns a list of transaction restrictions based on account status and settings
func (a *Accounts) GetTransactionRestrictions() []string {
	var restrictions []string

	// Status-based restrictions
	switch a.Status {
	case AccountStatusRestricted:
		restrictions = append(restrictions, "Manual entries only with approval")
	case AccountStatusFrozen:
		restrictions = append(restrictions, "No new transactions allowed")
	case AccountStatusYearEndProcessing:
		restrictions = append(restrictions, "Only closing entries permitted")
	case AccountStatusAuditLock:
		restrictions = append(restrictions, "Read-only during audit")
	}

	// Settings-based restrictions
	if !a.AllowManualEntries {
		restrictions = append(restrictions, "System entries only")
	}

	if a.RequireReference {
		restrictions = append(restrictions, "Reference number mandatory")
	}

	return restrictions
}

// CreateAccountRequest represents the request to create a new account
type CreateAccountRequest struct {
	EntityID           *uuid.UUID `json:"entity_id,omitempty"`
	AccountCode        string     `json:"account_code" validate:"required,max=20"`
	AccountName        string     `json:"account_name" validate:"required,max=200"`
	AccountDescription *string    `json:"account_description,omitempty" validate:"omitempty,max=1000"`

	// Hierarchy and grouping
	ParentAccountID *uuid.UUID `json:"parent_account_id,omitempty"`
	AccountGroupID  *uuid.UUID `json:"account_group_id,omitempty"`
	AccountHeaderID *uuid.UUID `json:"account_header_id,omitempty"`

	// Classification
	RootType        RootType      `json:"root_type" validate:"required"`
	AccountType     string        `json:"account_type" validate:"required,max=100"`
	AccountSubtype  *string       `json:"account_subtype,omitempty" validate:"omitempty,max=100"`
	AccountCategory *string       `json:"account_category,omitempty" validate:"omitempty,max=100"`
	SubCategory     *string       `json:"sub_category,omitempty" validate:"omitempty,max=100"`
	NormalBalance   NormalBalance `json:"normal_balance" validate:"required"`

	// Control and relationships
	IsControlAccount bool       `json:"is_control_account"`
	ControlAccountID *uuid.UUID `json:"control_account_id,omitempty"`

	// Currency settings
	CurrencyCode                *string `json:"currency_code,omitempty" validate:"omitempty,len=3"`
	IsMultiCurrency             bool    `json:"is_multi_currency"`
	CurrencyRevaluationRequired bool    `json:"currency_revaluation_required"`

	// Operational settings
	IsActive           bool `json:"is_active"`
	AllowManualEntries bool `json:"allow_manual_entries"`
	RequireReference   bool `json:"require_reference"`

	// Reporting and display
	FinancialStatementLine *string `json:"financial_statement_line,omitempty" validate:"omitempty,max=100"`
	ReportOrder            int32   `json:"report_order" validate:"min=0"`
	DisplayOrder           int32   `json:"display_order" validate:"min=0"`
	ShowInReports          bool    `json:"show_in_reports"`
	ConsolidationAccount   *string `json:"consolidation_account,omitempty" validate:"omitempty,max=100"`
	CashFlowType           *string `json:"cash_flow_type,omitempty" validate:"omitempty,max=50"`

	// Budgeting
	IsBudgetable            bool            `json:"is_budgetable"`
	BudgetVarianceThreshold decimal.Decimal `json:"budget_variance_threshold" validate:"min=0,max=100"`

	// Extensions
	AccountAttributes map[string]any `json:"account_attributes,omitempty"`
}

// UpdateAccountRequest represents the request to update an account
type UpdateAccountRequest struct {
	AccountCode        *string `json:"account_code,omitempty" validate:"omitempty,max=20"`
	AccountName        *string `json:"account_name,omitempty" validate:"omitempty,max=200"`
	AccountDescription *string `json:"account_description,omitempty" validate:"omitempty,max=1000"`

	// Grouping and hierarchy (limited updates for data integrity)
	AccountGroupID  *uuid.UUID `json:"account_group_id,omitempty"`
	AccountHeaderID *uuid.UUID `json:"account_header_id,omitempty"`

	// Classification updates
	AccountType     *string `json:"account_type,omitempty" validate:"omitempty,max=100"`
	AccountSubtype  *string `json:"account_subtype,omitempty" validate:"omitempty,max=100"`
	AccountCategory *string `json:"account_category,omitempty" validate:"omitempty,max=100"`
	SubCategory     *string `json:"sub_category,omitempty" validate:"omitempty,max=100"`

	// Operational settings
	IsActive           *bool `json:"is_active,omitempty"`
	AllowManualEntries *bool `json:"allow_manual_entries,omitempty"`
	RequireReference   *bool `json:"require_reference,omitempty"`

	// Reporting and display
	FinancialStatementLine *string `json:"financial_statement_line,omitempty" validate:"omitempty,max=100"`
	ReportOrder            *int32  `json:"report_order,omitempty" validate:"omitempty,min=0"`
	DisplayOrder           *int32  `json:"display_order,omitempty" validate:"omitempty,min=0"`
	ShowInReports          *bool   `json:"show_in_reports,omitempty"`
	ConsolidationAccount   *string `json:"consolidation_account,omitempty" validate:"omitempty,max=100"`
	CashFlowType           *string `json:"cash_flow_type,omitempty" validate:"omitempty,max=50"`

	// Budgeting
	IsBudgetable            *bool            `json:"is_budgetable,omitempty"`
	BudgetVarianceThreshold *decimal.Decimal `json:"budget_variance_threshold,omitempty" validate:"omitempty,min=0,max=100"`

	// Extensions
	AccountAttributes map[string]any `json:"account_attributes,omitempty"`
}

// AccountBalance represents account balance information at a point in time
type AccountBalance struct {
	AccountID    uuid.UUID       `json:"account_id"`
	Account      Accounts        `json:"account"`
	TotalDebits  decimal.Decimal `json:"total_debits"`  // Sum of all debit entries
	TotalCredits decimal.Decimal `json:"total_credits"` // Sum of all credit entries
	NetBalance   decimal.Decimal `json:"net_balance"`   // Calculated net balance
	AsOfDate     time.Time       `json:"as_of_date"`    // Balance calculation date
}

// TrialBalanceEntry represents a single line in the trial balance report
type TrialBalanceEntry struct {
	Account      Accounts        `json:"account"`
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
// AccountWithGroups represents an account with its group and header information from views

type AccountWithGroups struct {
	// All fields from Accounts
	Accounts

	// Additional fields from v_finance_accounts_with_groups
	GroupCode             *string `json:"group_code,omitempty"`
	GroupName             *string `json:"group_name,omitempty"`
	GroupCategory         *string `json:"group_category,omitempty"`
	HeaderCode            *string `json:"header_code,omitempty"`
	HeaderName            *string `json:"header_name,omitempty"`
	EffectiveDisplayOrder int32   `json:"effective_display_order"`
}

// ChartOfAccountsComplete represents the complete chart of accounts view data
type ChartOfAccountsComplete struct {
	// Core account information
	AccountID   uuid.UUID `json:"account_id"`
	AccountCode string    `json:"account_code"`
	AccountName string    `json:"account_name"`

	// Account classification
	RootType      string `json:"root_type"`
	AccountType   string `json:"account_type"`
	NormalBalance string `json:"normal_balance"`

	// Balance information
	CurrentBalance decimal.Decimal `json:"current_balance"`

	// Group and header information
	GroupCode     *string `json:"group_code,omitempty"`
	GroupName     *string `json:"group_name,omitempty"`
	GroupCategory *string `json:"group_category,omitempty"`
	HeaderCode    *string `json:"header_code,omitempty"`
	HeaderName    *string `json:"header_name,omitempty"`

	// Financial statement information
	StatementSection       *string `json:"statement_section,omitempty"`
	CashFlowClassification *string `json:"cash_flow_classification,omitempty"`

	// Display and reporting
	DisplayOrder     int32 `json:"display_order"`
	IncludeInReports bool  `json:"include_in_reports"`
	IsActive         bool  `json:"is_active"`
	IsLeafAccount    bool  `json:"is_leaf_account"`

	// Audit information
	TenantID  uuid.UUID `json:"tenant_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TrialBalanceSummary represents trial balance data from views
type TrialBalanceSummary struct {
	AccountID        uuid.UUID       `json:"account_id"`
	AccountCode      string          `json:"account_code"`
	AccountName      string          `json:"account_name"`
	RootType         string          `json:"root_type"`
	NormalBalance    string          `json:"normal_balance"`
	CurrentBalance   decimal.Decimal `json:"current_balance"`
	GroupName        *string         `json:"group_name,omitempty"`
	StatementSection *string         `json:"statement_section,omitempty"`
}

// CashFlowAccount represents cash flow account information from views
type CashFlowAccount struct {
	AccountID              uuid.UUID       `json:"account_id"`
	AccountCode            string          `json:"account_code"`
	AccountName            string          `json:"account_name"`
	CurrentBalance         decimal.Decimal `json:"current_balance"`
	CashFlowClassification *string         `json:"cash_flow_classification,omitempty"`
	GroupName              *string         `json:"group_name,omitempty"`
}

// AccountGroupSummary represents grouped account summary data
type AccountGroupSummary struct {
	GroupCode          string          `json:"group_code"`
	GroupName          string          `json:"group_name"`
	GroupCategory      *string         `json:"group_category,omitempty"`
	StatementSection   *string         `json:"statement_section,omitempty"`
	AccountCount       int64           `json:"account_count"`
	ActiveAccountCount int64           `json:"active_account_count"`
	TotalBalance       decimal.Decimal `json:"total_balance"`
	ActiveBalance      decimal.Decimal `json:"active_balance"`
}

// ChartOfAccountsFilter represents filtering options for complete chart of accounts
type ChartOfAccountsFilter struct {
	EntityID         *uuid.UUID        `json:"entity_id,omitempty"`
	StatementSection *StatementSection `json:"statement_section,omitempty"`
	IncludeInactive  *bool             `json:"include_inactive,omitempty"`
	IncludeInReports *bool             `json:"include_in_reports,omitempty"`
	Limit            *int              `json:"limit,omitempty"`
	Offset           *int              `json:"offset,omitempty"`
}

// BalanceFilter represents filtering options for accounts with balances
type BalanceFilter struct {
	EntityID    *uuid.UUID `json:"entity_id,omitempty"`
	NonZeroOnly *bool      `json:"non_zero_only,omitempty"`
	RootType    *RootType  `json:"root_type,omitempty"`
	Limit       *int       `json:"limit,omitempty"`
	Offset      *int       `json:"offset,omitempty"`
}

// Validate validates the ChartOfAccountsFilter
func (f *ChartOfAccountsFilter) Validate() []ValidationError {
	var errors []ValidationError

	if f.StatementSection != nil && !f.StatementSection.IsValid() {
		errors = append(errors, ValidationError{
			Field:   "statement_section",
			Message: "Invalid statement section",
			Code:    "INVALID_VALUE",
		})
	}

	if f.Limit != nil && *f.Limit < 0 {
		errors = append(errors, ValidationError{
			Field:   "limit",
			Message: "Limit must be non-negative",
			Code:    "INVALID_VALUE",
		})
	}

	if f.Offset != nil && *f.Offset < 0 {
		errors = append(errors, ValidationError{
			Field:   "offset",
			Message: "Offset must be non-negative",
			Code:    "INVALID_VALUE",
		})
	}

	return errors
}

// Validate validates the BalanceFilter
func (f *BalanceFilter) Validate() []ValidationError {
	var errors []ValidationError

	if f.RootType != nil && !f.RootType.IsValid() {
		errors = append(errors, ValidationError{
			Field:   "root_type",
			Message: "Invalid root type",
			Code:    "INVALID_VALUE",
		})
	}

	if f.Limit != nil && *f.Limit < 0 {
		errors = append(errors, ValidationError{
			Field:   "limit",
			Message: "Limit must be non-negative",
			Code:    "INVALID_VALUE",
		})
	}

	if f.Offset != nil && *f.Offset < 0 {
		errors = append(errors, ValidationError{
			Field:   "offset",
			Message: "Offset must be non-negative",
			Code:    "INVALID_VALUE",
		})
	}

	return errors
}

// NOTE: Consider adding account category groupings for enhanced reporting
// NOTE: Future enhancement: Add support for account-specific posting rules
