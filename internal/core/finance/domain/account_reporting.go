package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// =============================================================================
// Supporting enumerations
// =============================================================================

// ConsolidationMethod defines how a AccountGroup aggregates the balances of
// its members when producing a financial statement line total.
type ConsolidationMethod string

const (
	// ConsolidationSum totals all member balances. Used for almost all
	// standard P&L and balance sheet groups.
	ConsolidationSum ConsolidationMethod = "SUM"

	// ConsolidationAverage takes the arithmetic mean of member balances.
	// Useful for period-average exchange rate reporting.
	ConsolidationAverage ConsolidationMethod = "AVERAGE"

	// ConsolidationMax / ConsolidationMin select the highest or lowest
	// individual balance. Uncommon; used in specific regulatory templates.
	ConsolidationMax ConsolidationMethod = "MAX"
	ConsolidationMin ConsolidationMethod = "MIN"

	// ConsolidationCustom defers the computation to a named formula registered
	// in the reporting engine configuration.
	ConsolidationCustom ConsolidationMethod = "CUSTOM"
)

// IsValid returns true if m is one of the recognised constants.
func (m ConsolidationMethod) IsValid() bool {
	switch m {
	case ConsolidationSum, ConsolidationAverage,
		ConsolidationMax, ConsolidationMin, ConsolidationCustom:
		return true
	default:
		return false
	}
}

// CashFlowCategory classifies a reporting group's position in the cash flow
// statement following IAS 7 / IFRS framework.
type CashFlowCategory string

const (
	CashFlowOperating CashFlowCategory = "OPERATING"
	CashFlowInvesting CashFlowCategory = "INVESTING"
	CashFlowFinancing CashFlowCategory = "FINANCING"
	CashFlowNonCash   CashFlowCategory = "NON_CASH"
)

// IsValid returns true if c is a recognised cash flow category.
func (c CashFlowCategory) IsValid() bool {
	switch c {
	case CashFlowOperating, CashFlowInvesting, CashFlowFinancing, CashFlowNonCash:
		return true
	default:
		return false
	}
}

// StatementSection identifies the financial statement a reporting group belongs to.
type StatementSection string

const (
	StatementSectionBalanceSheet    StatementSection = "BALANCE_SHEET"
	StatementSectionIncomeStatement StatementSection = "INCOME_STATEMENT"
	StatementSectionCashFlow        StatementSection = "CASH_FLOW"
	StatementSectionEquity          StatementSection = "EQUITY"
	StatementSectionNotes           StatementSection = "NOTES"
)

// IsValid returns true if s is a recognised statement section.
func (s StatementSection) IsValid() bool {
	switch s {
	case StatementSectionBalanceSheet, StatementSectionIncomeStatement,
		StatementSectionCashFlow, StatementSectionEquity, StatementSectionNotes:
		return true
	default:
		return false
	}
}

// =============================================================================
// Field-length constants
// =============================================================================

const (
	AccountGroupCodeMaxLen        = 50
	AccountGroupNameMaxLen        = 255
	AccountGroupDescriptionMaxLen = 500
	AccountGroupIndentLevelMax    = 10
)

// =============================================================================
// AccountGroup
// =============================================================================

// AccountGroup is a node in a reporting hierarchy that is completely
// independent of the ledger account tree.
//
// # Separation from the ledger tree
//
// Ledger accounts (Account) are arranged for double-entry correctness: they
// must maintain RootType homogeneity and roll up balances through a single
// structural tree.
//
// AccountGroups are arranged for presentation: a P&L might group "Fuel Sales"
// and "Lubrication Sales" under "Petroleum Revenue" regardless of their
// ledger codes or account types. The same Account can be mapped to multiple
// AccountGroups across different ReportingSchemes (EPRA, KRA, internal).
//
// The link between an Account and a AccountGroup is AccountMapping — never
// a direct foreign key from Account to AccountGroup.
//
// # Hierarchy
//
// AccountGroups form their own tree via ParentGroupID. Level and Path use
// MaterialisedPath for efficient subtree queries, identical to the Account tree.
//
// # Display
//
// DisplayOrder, IndentLevel, BoldDisplay, ShowTotals, and IsHeader are purely
// presentational; they tell the report renderer how to lay out this group's
// line in the output statement. They have no effect on balance computation.
//
// # System groups
//
// IsSystemDefined groups are provisioned by the chart of accounts seed for each
// tenant. They cannot be renamed, reparented, or deleted by tenant users.
type AccountGroup struct {
	ID       uuid.UUID  `json:"id"`
	TenantID uuid.UUID  `json:"tenant_id"`
	EntityID *uuid.UUID `json:"entity_id,omitempty"` // Scoped to a specific entity, or nil for tenant-wide

	// -------------------------------------------------------------------------
	// Identity
	// -------------------------------------------------------------------------

	// GroupCode is a unique mnemonic within the tenant + scheme, e.g. "REV-FUEL".
	// Maximum AccountGroupCodeMaxLen characters.
	GroupCode string `json:"group_code"`

	// GroupName is the label printed on the financial statement.
	// Maximum AccountGroupNameMaxLen characters.
	GroupName string `json:"group_name"`

	// Description is an optional explanation of what this group covers.
	Description *string `json:"description,omitempty"`

	// -------------------------------------------------------------------------
	// Hierarchy
	// -------------------------------------------------------------------------

	// ParentGroupID is the direct parent within the same reporting scheme tree.
	// Nil for root groups.
	ParentGroupID *uuid.UUID `json:"parent_group_id,omitempty"`

	// Level is the zero-based depth (0 = root group in a scheme).
	Level int `json:"level"`

	// Path is the materialised ancestor path maintained by the repository.
	Path MaterialisedPath `json:"path"`

	// -------------------------------------------------------------------------
	// Classification
	// -------------------------------------------------------------------------

	// ReportingSchemeID references the ReportingScheme this group belongs to
	// (e.g. EPRA, KRA, IFRS, INTERNAL). A group belongs to exactly one scheme.
	// Cross-scheme mapping is done at the AccountMapping level.
	ReportingSchemeID uuid.UUID `json:"reporting_scheme_id"`

	// StatementSection places this group on a specific financial statement.
	StatementSection *StatementSection `json:"financial_statement_section,omitempty"`

	// ConsolidationMethod overrides the scheme-level default for this group.
	// Nil means inherit the scheme default (usually SUM).
	ConsolidationMethod *ConsolidationMethod `json:"consolidation_method,omitempty"`

	// CashFlowCategory classifies the group for the cash flow statement.
	// Required when StatementSection == CASH_FLOW.
	CashFlowCategory *CashFlowCategory `json:"cash_flow_category,omitempty"`

	// -------------------------------------------------------------------------
	// Presentation
	// -------------------------------------------------------------------------

	// DisplayOrder controls the sort position of this group within its parent.
	// Lower numbers appear first.
	DisplayOrder int `json:"display_order"`

	// IsHeader marks this group as a heading row only. Header groups do not
	// display a balance total; they are separators in the statement layout.
	IsHeader bool `json:"is_header"`

	// ShowTotals controls whether the group's aggregated balance is printed.
	// Typically true for subtotal and total rows.
	ShowTotals bool `json:"show_totals"`

	// IndentLevel is the visual indentation depth in the printed statement.
	// Range [0, AccountGroupIndentLevelMax].
	IndentLevel int `json:"indent_level"`

	// BoldDisplay renders the group label in bold in the report output.
	BoldDisplay bool `json:"bold_display"`

	// -------------------------------------------------------------------------
	// Lifecycle
	// -------------------------------------------------------------------------

	// IsSystemDefined marks groups provisioned by the seed for standard schemes
	// (EPRA, KRA). Tenant users may not modify or delete system-defined groups.
	IsSystemDefined bool `json:"is_system_defined"`

	// IsActive controls whether this group is included in report generation.
	// Inactive groups are hidden from reports but retained for history.
	IsActive bool `json:"is_active"`

	// -------------------------------------------------------------------------
	// Audit
	// -------------------------------------------------------------------------

	Version   int       `json:"version"` // Optimistic locking counter
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedBy uuid.UUID `json:"created_by"`
	UpdatedBy uuid.UUID `json:"updated_by"` // Always set; populated from CreatedBy on first update
}

// =============================================================================
// Behaviour methods
// =============================================================================

// IsRoot returns true when this group sits at the top of its scheme hierarchy.
func (rg *AccountGroup) IsRoot() bool {
	return rg.ParentGroupID == nil && rg.Level == 0
}

// IsDescendantOf reports whether this group is a descendant of the group
// identified by ancestorID using the materialised Path.
func (rg *AccountGroup) IsDescendantOf(ancestorID uuid.UUID) bool {
	return rg.Path.IsDescendantOf(ancestorID)
}

// CanBeModifiedByTenant returns true when a tenant user (as opposed to a system
// administrator) may edit or delete this group.
func (rg *AccountGroup) CanBeModifiedByTenant() bool {
	return !rg.IsSystemDefined
}

// DisplayLabel returns the group name formatted for statement output,
// applying bold markers when BoldDisplay is true.
// The formatting tokens ("**") are interpreted by the report renderer.
func (rg *AccountGroup) DisplayLabel() string {
	if rg.BoldDisplay {
		return "**" + rg.GroupName + "**"
	}
	return rg.GroupName
}

// =============================================================================
// Validation
// =============================================================================

// Validate performs entity-level consistency checks.
// Returns []ValidationError for consistency with the rest of the domain package.
func (rg *AccountGroup) Validate() []ValidationError {
	var errs []ValidationError

	// -- Tenant ---------------------------------------------------------------
	if rg.TenantID == uuid.Nil {
		errs = append(errs, ValidationError{
			Field: "tenant_id", Message: "tenant_id is required",
			Code: "REQUIRED_FIELD", Severity: ValidationSeverityError,
		})
	}

	// -- Scheme ---------------------------------------------------------------
	if rg.ReportingSchemeID == uuid.Nil {
		errs = append(errs, ValidationError{
			Field: "reporting_scheme_id", Message: "reporting_scheme_id is required",
			Code: "REQUIRED_FIELD", Severity: ValidationSeverityError,
		})
	}

	// -- Code -----------------------------------------------------------------
	switch {
	case strings.TrimSpace(rg.GroupCode) == "":
		errs = append(errs, ValidationError{
			Field: "group_code", Message: "group code is required",
			Code: "REQUIRED_FIELD", Severity: ValidationSeverityError,
		})
	case len(rg.GroupCode) > AccountGroupCodeMaxLen:
		errs = append(errs, ValidationError{
			Field: "group_code",
			Message: fmt.Sprintf(
				"group code must be %d characters or less", AccountGroupCodeMaxLen,
			),
			Code: "MAX_LENGTH_EXCEEDED", Severity: ValidationSeverityError,
		})
	}

	// -- Name -----------------------------------------------------------------
	switch {
	case strings.TrimSpace(rg.GroupName) == "":
		errs = append(errs, ValidationError{
			Field: "group_name", Message: "group name is required",
			Code: "REQUIRED_FIELD", Severity: ValidationSeverityError,
		})
	case len(rg.GroupName) > AccountGroupNameMaxLen:
		errs = append(errs, ValidationError{
			Field: "group_name",
			Message: fmt.Sprintf(
				"group name must be %d characters or less", AccountGroupNameMaxLen,
			),
			Code: "MAX_LENGTH_EXCEEDED", Severity: ValidationSeverityError,
		})
	}

	// -- Description ----------------------------------------------------------
	if rg.Description != nil && len(*rg.Description) > AccountGroupDescriptionMaxLen {
		errs = append(errs, ValidationError{
			Field: "description",
			Message: fmt.Sprintf(
				"description must be %d characters or less", AccountGroupDescriptionMaxLen,
			),
			Code: "MAX_LENGTH_EXCEEDED", Severity: ValidationSeverityError,
		})
	}

	// -- Hierarchy consistency ------------------------------------------------
	if rg.ParentGroupID != nil && *rg.ParentGroupID == rg.ID {
		errs = append(errs, ValidationError{
			Field:    "parent_group_id",
			Message:  "a reporting group cannot be its own parent",
			Code:     "SELF_REFERENCE",
			Severity: ValidationSeverityError,
		})
	}
	if rg.ParentGroupID == nil && rg.Level != 0 {
		errs = append(errs, ValidationError{
			Field:    "level",
			Message:  "root reporting group (no parent) must have level 0",
			Code:     "INCONSISTENT_VALUE",
			Severity: ValidationSeverityError,
		})
	}
	if rg.ParentGroupID != nil && rg.Level <= 0 {
		errs = append(errs, ValidationError{
			Field:    "level",
			Message:  "child reporting group (has parent) must have level > 0",
			Code:     "INCONSISTENT_VALUE",
			Severity: ValidationSeverityError,
		})
	}
	if err := rg.Path.Validate(); err != nil {
		errs = append(errs, ValidationError{
			Field:    "path",
			Message:  err.Error(),
			Code:     "INVALID_FORMAT",
			Severity: ValidationSeverityError,
		})
	}

	// -- Classification -------------------------------------------------------
	if rg.StatementSection != nil && !rg.StatementSection.IsValid() {
		errs = append(errs, ValidationError{
			Field:    "financial_statement_section",
			Message:  fmt.Sprintf("invalid statement section: %q", *rg.StatementSection),
			Code:     "INVALID_VALUE",
			Severity: ValidationSeverityError,
		})
	}
	if rg.ConsolidationMethod != nil && !rg.ConsolidationMethod.IsValid() {
		errs = append(errs, ValidationError{
			Field:    "consolidation_method",
			Message:  fmt.Sprintf("invalid consolidation method: %q", *rg.ConsolidationMethod),
			Code:     "INVALID_VALUE",
			Severity: ValidationSeverityError,
		})
	}
	if rg.CashFlowCategory != nil && !rg.CashFlowCategory.IsValid() {
		errs = append(errs, ValidationError{
			Field:    "cash_flow_category",
			Message:  fmt.Sprintf("invalid cash flow category: %q", *rg.CashFlowCategory),
			Code:     "INVALID_VALUE",
			Severity: ValidationSeverityError,
		})
	}
	// Cash flow category is required when the group is placed on the cash flow statement.
	if rg.StatementSection != nil &&
		*rg.StatementSection == StatementSectionCashFlow &&
		rg.CashFlowCategory == nil {
		errs = append(errs, ValidationError{
			Field:    "cash_flow_category",
			Message:  "cash_flow_category is required for cash flow statement groups",
			Code:     "REQUIRED_FIELD",
			Severity: ValidationSeverityError,
		})
	}

	// -- Presentation ---------------------------------------------------------
	if rg.DisplayOrder < 0 {
		errs = append(errs, ValidationError{
			Field: "display_order", Message: "display order must be non-negative",
			Code: "OUT_OF_RANGE", Severity: ValidationSeverityError,
		})
	}
	if rg.IndentLevel < 0 || rg.IndentLevel > AccountGroupIndentLevelMax {
		errs = append(errs, ValidationError{
			Field: "indent_level",
			Message: fmt.Sprintf(
				"indent level must be between 0 and %d", AccountGroupIndentLevelMax,
			),
			Code: "OUT_OF_RANGE", Severity: ValidationSeverityError,
		})
	}
	// A header group should not claim to show totals — it is a label row.
	if rg.IsHeader && rg.ShowTotals {
		errs = append(errs, ValidationError{
			Field:    "show_totals",
			Message:  "header groups are label rows and must not show totals",
			Code:     "INCONSISTENT_VALUE",
			Severity: ValidationSeverityWarning,
		})
	}

	return errs
}

// =============================================================================
// Request DTOs
// =============================================================================

// CreateAccountGroupRequest is the inbound payload for creating a new group.
// Level and Path are derived by the service from ParentGroupID.
type CreateAccountGroupRequest struct {
	EntityID          *uuid.UUID `json:"entity_id,omitempty"`
	ReportingSchemeID uuid.UUID  `json:"reporting_scheme_id" validate:"required"`
	GroupCode         string     `json:"group_code"          validate:"required,max=50"`
	GroupName         string     `json:"group_name"          validate:"required,max=255"`
	Description       *string    `json:"description,omitempty" validate:"omitempty,max=500"`

	ParentGroupID *uuid.UUID `json:"parent_group_id,omitempty"`

	StatementSection    *StatementSection    `json:"financial_statement_section,omitempty"`
	ConsolidationMethod *ConsolidationMethod `json:"consolidation_method,omitempty"`
	CashFlowCategory    *CashFlowCategory    `json:"cash_flow_category,omitempty"`

	DisplayOrder int  `json:"display_order"`
	IsHeader     bool `json:"is_header"`
	ShowTotals   bool `json:"show_totals"`
	IndentLevel  int  `json:"indent_level"`
	BoldDisplay  bool `json:"bold_display"`
}

// Validate checks request-level self-consistency.
func (r *CreateAccountGroupRequest) Validate() []ValidationError {
	var errs []ValidationError

	if r.ReportingSchemeID == uuid.Nil {
		errs = append(errs, ValidationError{
			Field: "reporting_scheme_id", Message: "reporting_scheme_id is required",
			Code: "REQUIRED_FIELD", Severity: ValidationSeverityError,
		})
	}
	if strings.TrimSpace(r.GroupCode) == "" {
		errs = append(errs, ValidationError{
			Field: "group_code", Message: "group code is required",
			Code: "REQUIRED_FIELD", Severity: ValidationSeverityError,
		})
	} else if len(r.GroupCode) > AccountGroupCodeMaxLen {
		errs = append(errs, ValidationError{
			Field:   "group_code",
			Message: fmt.Sprintf("group code must be %d characters or less", AccountGroupCodeMaxLen),
			Code:    "MAX_LENGTH_EXCEEDED", Severity: ValidationSeverityError,
		})
	}
	if strings.TrimSpace(r.GroupName) == "" {
		errs = append(errs, ValidationError{
			Field: "group_name", Message: "group name is required",
			Code: "REQUIRED_FIELD", Severity: ValidationSeverityError,
		})
	} else if len(r.GroupName) > AccountGroupNameMaxLen {
		errs = append(errs, ValidationError{
			Field:   "group_name",
			Message: fmt.Sprintf("group name must be %d characters or less", AccountGroupNameMaxLen),
			Code:    "MAX_LENGTH_EXCEEDED", Severity: ValidationSeverityError,
		})
	}
	if r.StatementSection != nil && !r.StatementSection.IsValid() {
		errs = append(errs, ValidationError{
			Field:    "financial_statement_section",
			Message:  fmt.Sprintf("invalid statement section: %q", *r.StatementSection),
			Code:     "INVALID_VALUE",
			Severity: ValidationSeverityError,
		})
	}
	if r.StatementSection != nil &&
		*r.StatementSection == StatementSectionCashFlow &&
		r.CashFlowCategory == nil {
		errs = append(errs, ValidationError{
			Field:    "cash_flow_category",
			Message:  "cash_flow_category is required for cash flow statement groups",
			Code:     "REQUIRED_FIELD",
			Severity: ValidationSeverityError,
		})
	}
	if r.IndentLevel < 0 || r.IndentLevel > AccountGroupIndentLevelMax {
		errs = append(errs, ValidationError{
			Field:   "indent_level",
			Message: fmt.Sprintf("indent level must be between 0 and %d", AccountGroupIndentLevelMax),
			Code:    "OUT_OF_RANGE", Severity: ValidationSeverityError,
		})
	}

	return errs
}

// UpdateAccountGroupRequest is the partial-update payload.
//
// GroupCode and ReportingSchemeID are excluded:
//   - Code changes require a uniqueness re-check via a dedicated service operation.
//   - Scheme changes require remapping all AccountMappings; use a migration operation.
//
// ParentGroupID is excluded: reparenting requires a full subtree re-path via a
// dedicated Reparent operation.
type UpdateAccountGroupRequest struct {
	GroupName   *string `json:"group_name,omitempty"  validate:"omitempty,max=255"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`

	StatementSection    *StatementSection    `json:"financial_statement_section,omitempty"`
	ConsolidationMethod *ConsolidationMethod `json:"consolidation_method,omitempty"`
	CashFlowCategory    *CashFlowCategory    `json:"cash_flow_category,omitempty"`

	DisplayOrder *int  `json:"display_order,omitempty"`
	IsActive     *bool `json:"is_active,omitempty"`
	IsHeader     *bool `json:"is_header,omitempty"`
	ShowTotals   *bool `json:"show_totals,omitempty"`
	IndentLevel  *int  `json:"indent_level,omitempty"`
	BoldDisplay  *bool `json:"bold_display,omitempty"`
}

// =============================================================================
// Filter type
// =============================================================================

// AccountGroupFilter is the query predicate for listing reporting groups.
type AccountGroupFilter struct {
	TenantID          *uuid.UUID        `json:"tenant_id,omitempty"`
	EntityID          *uuid.UUID        `json:"entity_id,omitempty"`
	ReportingSchemeID *uuid.UUID        `json:"reporting_scheme_id,omitempty"`
	ParentGroupID     *uuid.UUID        `json:"parent_group_id,omitempty"`
	StatementSection  *StatementSection `json:"financial_statement_section,omitempty"`
	CashFlowCategory  *CashFlowCategory `json:"cash_flow_category,omitempty"`
	IsActive          *bool             `json:"is_active,omitempty"`
	IsSystemDefined   *bool             `json:"is_system_defined,omitempty"`
	IsHeader          *bool             `json:"is_header,omitempty"`
	MaxLevel          *int              `json:"max_level,omitempty"`

	// Pagination
	Limit  *int `json:"limit,omitempty"`
	Offset *int `json:"offset,omitempty"`

	// Sorting
	SortBy    *string `json:"sort_by,omitempty"`
	SortOrder *string `json:"sort_order,omitempty"`
}
