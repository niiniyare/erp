package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// =============================================================================
// Supporting enumerations
// =============================================================================

// ConsolidationMethod defines how a group aggregates member balances when
// producing a financial statement line total.
// DB values: SUM | AVERAGE | MAX | MIN | CUSTOM (finance_account_groups.consolidation_method).
type ConsolidationMethod string

const (
	ConsolidationSum     ConsolidationMethod = "SUM"
	ConsolidationAverage ConsolidationMethod = "AVERAGE"
	ConsolidationMax     ConsolidationMethod = "MAX"
	ConsolidationMin     ConsolidationMethod = "MIN"
	ConsolidationCustom  ConsolidationMethod = "CUSTOM"
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

// CashFlowCategory classifies a group per IAS 7 / IFRS cash flow statement.
// DB constraint on finance_account_groups: OPERATING | INVESTING | FINANCING only.
// NON_CASH is intentionally excluded — the DB CHECK constraint does not permit it.
type CashFlowCategory string

const (
	CashFlowOperating CashFlowCategory = "OPERATING"
	CashFlowInvesting CashFlowCategory = "INVESTING"
	CashFlowFinancing CashFlowCategory = "FINANCING"
)

// IsValid returns true if c is a recognised cash flow category.
func (c CashFlowCategory) IsValid() bool {
	switch c {
	case CashFlowOperating, CashFlowInvesting, CashFlowFinancing:
		return true
	default:
		return false
	}
}

// StatementSection identifies which financial statement a group belongs to.
// Stored as free-text in finance_account_groups.financial_statement_section.
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
// Field-length constants (shared by AccountGroup and ReportingGroup)
// =============================================================================

const (
	AccountGroupCodeMaxLen        = 50
	AccountGroupNameMaxLen        = 255
	AccountGroupDescriptionMaxLen = 500
	AccountGroupIndentLevelMax    = 10
)

// AccountGroup —maps finance_account_groups (migration 000901)
// AccountGroup is the organisational grouping node used to structure the chart
// of accounts for display and simple balance roll-up. It is backed directly by
// the finance_account_groups DB table.
//
// For multi-scheme regulatory reporting (EPRA, KRA, IFRS), see ReportingGroup.
type AccountGroup struct {
	ID       uuid.UUID  `json:"id"`
	TenantID uuid.UUID  `json:"tenant_id"`
	EntityID *uuid.UUID `json:"entity_id,omitempty"`

	// Identity
	GroupCode   string `json:"group_code"`
	GroupName   string `json:"group_name"`
	Description string `json:"description,omitempty"` // DB: group_description TEXT (nullable)

	// Hierarchy — DB: parent_group_id, group_level, group_path
	ParentGroupID *uuid.UUID `json:"parent_group_id,omitempty"`
	Level         int        `json:"level"`
	Path          string     `json:"path,omitempty"` // materialised path, maintained by trigger

	// Classification — two separate DB columns
	RootType      RootType `json:"root_type"`                // DB: root_type NOT NULL CHECK(...)
	GroupCategory string   `json:"group_category,omitempty"` // DB: group_category (e.g. CURRENT_ASSETS)

	// Financial statement placement
	FinancialStatementSection *string `json:"financial_statement_section,omitempty"`

	// Consolidation
	ConsolidationMethod *ConsolidationMethod `json:"consolidation_method,omitempty"`
	CashFlowCategory    *CashFlowCategory    `json:"cash_flow_category,omitempty"`

	// Presentation — DB: statement_order, show_in_summary, indent_level, show_totals, bold_display
	DisplayOrder int  `json:"display_order"` // DB: statement_order
	IsHeader     bool `json:"is_header"`     // DB: show_in_summary
	ShowTotals   bool `json:"show_totals"`
	IndentLevel  int  `json:"indent_level"`
	BoldDisplay  bool `json:"bold_display"`

	// Lifecycle — DB: is_system_group, is_active
	IsSystemDefined bool `json:"is_system_defined"` // DB: is_system_group
	IsActive        bool `json:"is_active"`

	// Audit
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedBy uuid.UUID `json:"created_by"`
	UpdatedBy uuid.UUID `json:"updated_by"`
}

// AccountGroupFilter is the query predicate for listing account groups.
// Matches the parameters consumed by the repository mapper.
type AccountGroupFilter struct {
	EntityID                  *uuid.UUID        `json:"entity_id,omitempty"`
	GroupType                 *string           `json:"group_type,omitempty"` // filters on group_category
	ParentGroupID             *uuid.UUID        `json:"parent_group_id,omitempty"`
	IsActive                  *bool             `json:"is_active,omitempty"`
	FinancialStatementSection *string           `json:"financial_statement_section,omitempty"`
	CashFlowCategory          *CashFlowCategory `json:"cash_flow_category,omitempty"`
	SortBy                    *string           `json:"sort_by,omitempty"`
	Limit                     *int              `json:"limit,omitempty"`
	Offset                    *int              `json:"offset,omitempty"`
}

// ReportingGroup — multi-scheme reporting hierarchy (EPRA, KRA, IFRS, INTERNAL)
// ReportingGroup is a node in a reporting hierarchy that is completely
// independent of the ledger account tree. Multiple ReportingSchemes can coexist;
// the same ledger account can appear in different groups across schemes via
// AccountMapping.
//
// Note: ReportingGroup does not yet have a dedicated DB migration. It will be
// added when the multi-scheme reporting engine is implemented. Until then, the
// type lives here to support AccountMapping and ReportingScheme domain logic.
type ReportingGroup struct {
	ID       uuid.UUID  `json:"id"`
	TenantID uuid.UUID  `json:"tenant_id"`
	EntityID *uuid.UUID `json:"entity_id,omitempty"`

	// Identity
	GroupCode         string    `json:"group_code"`
	GroupName         string    `json:"group_name"`
	Description       *string   `json:"description,omitempty"`
	ReportingSchemeID uuid.UUID `json:"reporting_scheme_id"`

	// Hierarchy
	ParentGroupID *uuid.UUID       `json:"parent_group_id,omitempty"`
	Level         int              `json:"level"`
	Path          MaterialisedPath `json:"path"`

	// Classification
	StatementSection    *StatementSection    `json:"financial_statement_section,omitempty"`
	ConsolidationMethod *ConsolidationMethod `json:"consolidation_method,omitempty"`
	CashFlowCategory    *CashFlowCategory    `json:"cash_flow_category,omitempty"`

	// Presentation
	DisplayOrder    int  `json:"display_order"`
	IsHeader        bool `json:"is_header"`
	ShowTotals      bool `json:"show_totals"`
	IndentLevel     int  `json:"indent_level"`
	BoldDisplay     bool `json:"bold_display"`
	IsSystemDefined bool `json:"is_system_defined"`
	IsActive        bool `json:"is_active"`

	// Audit
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedBy uuid.UUID `json:"created_by"`
	UpdatedBy uuid.UUID `json:"updated_by"`
}

// IsRoot returns true when this group sits at the top of its scheme hierarchy.
func (rg *ReportingGroup) IsRoot() bool {
	return rg.ParentGroupID == nil && rg.Level == 0
}

// IsDescendantOf reports whether this group is a descendant of ancestorID.
func (rg *ReportingGroup) IsDescendantOf(ancestorID uuid.UUID) bool {
	return rg.Path.IsDescendantOf(ancestorID)
}

// CanBeModifiedByTenant returns true when a tenant user may edit or delete this group.
func (rg *ReportingGroup) CanBeModifiedByTenant() bool {
	return !rg.IsSystemDefined
}

// DisplayLabel returns the group name with bold markers when BoldDisplay is true.
func (rg *ReportingGroup) DisplayLabel() string {
	if rg.BoldDisplay {
		return "**" + rg.GroupName + "**"
	}
	return rg.GroupName
}

// Validate performs entity-level consistency checks.
func (rg *ReportingGroup) Validate() []ValidationError {
	var errs []ValidationError

	if rg.TenantID == uuid.Nil {
		errs = append(errs, ValidationError{
			Field: "tenant_id", Message: "tenant_id is required",
			Code: "REQUIRED_FIELD", Severity: ValidationSeverityError,
		})
	}
	if rg.ReportingSchemeID == uuid.Nil {
		errs = append(errs, ValidationError{
			Field: "reporting_scheme_id", Message: "reporting_scheme_id is required",
			Code: "REQUIRED_FIELD", Severity: ValidationSeverityError,
		})
	}
	switch {
	case strings.TrimSpace(rg.GroupCode) == "":
		errs = append(errs, ValidationError{
			Field: "group_code", Message: "group code is required",
			Code: "REQUIRED_FIELD", Severity: ValidationSeverityError,
		})
	case len(rg.GroupCode) > AccountGroupCodeMaxLen:
		errs = append(errs, ValidationError{
			Field:   "group_code",
			Message: fmt.Sprintf("group code must be %d characters or less", AccountGroupCodeMaxLen),
			Code:    "MAX_LENGTH_EXCEEDED", Severity: ValidationSeverityError,
		})
	}
	switch {
	case strings.TrimSpace(rg.GroupName) == "":
		errs = append(errs, ValidationError{
			Field: "group_name", Message: "group name is required",
			Code: "REQUIRED_FIELD", Severity: ValidationSeverityError,
		})
	case len(rg.GroupName) > AccountGroupNameMaxLen:
		errs = append(errs, ValidationError{
			Field:   "group_name",
			Message: fmt.Sprintf("group name must be %d characters or less", AccountGroupNameMaxLen),
			Code:    "MAX_LENGTH_EXCEEDED", Severity: ValidationSeverityError,
		})
	}
	if rg.Description != nil && len(*rg.Description) > AccountGroupDescriptionMaxLen {
		errs = append(errs, ValidationError{
			Field:   "description",
			Message: fmt.Sprintf("description must be %d characters or less", AccountGroupDescriptionMaxLen),
			Code:    "MAX_LENGTH_EXCEEDED", Severity: ValidationSeverityError,
		})
	}
	if rg.ParentGroupID != nil && *rg.ParentGroupID == rg.ID {
		errs = append(errs, ValidationError{
			Field: "parent_group_id", Message: "a reporting group cannot be its own parent",
			Code: "SELF_REFERENCE", Severity: ValidationSeverityError,
		})
	}
	if rg.ParentGroupID == nil && rg.Level != 0 {
		errs = append(errs, ValidationError{
			Field: "level", Message: "root group (no parent) must have level 0",
			Code: "INCONSISTENT_VALUE", Severity: ValidationSeverityError,
		})
	}
	if rg.ParentGroupID != nil && rg.Level <= 0 {
		errs = append(errs, ValidationError{
			Field: "level", Message: "child group (has parent) must have level > 0",
			Code: "INCONSISTENT_VALUE", Severity: ValidationSeverityError,
		})
	}
	if err := rg.Path.Validate(); err != nil {
		errs = append(errs, ValidationError{
			Field: "path", Message: err.Error(),
			Code: "INVALID_FORMAT", Severity: ValidationSeverityError,
		})
	}
	if rg.StatementSection != nil && !rg.StatementSection.IsValid() {
		errs = append(errs, ValidationError{
			Field:   "financial_statement_section",
			Message: fmt.Sprintf("invalid statement section: %q", *rg.StatementSection),
			Code:    "INVALID_VALUE", Severity: ValidationSeverityError,
		})
	}
	if rg.ConsolidationMethod != nil && !rg.ConsolidationMethod.IsValid() {
		errs = append(errs, ValidationError{
			Field:   "consolidation_method",
			Message: fmt.Sprintf("invalid consolidation method: %q", *rg.ConsolidationMethod),
			Code:    "INVALID_VALUE", Severity: ValidationSeverityError,
		})
	}
	if rg.CashFlowCategory != nil && !rg.CashFlowCategory.IsValid() {
		errs = append(errs, ValidationError{
			Field:   "cash_flow_category",
			Message: fmt.Sprintf("invalid cash flow category: %q", *rg.CashFlowCategory),
			Code:    "INVALID_VALUE", Severity: ValidationSeverityError,
		})
	}
	if rg.StatementSection != nil &&
		*rg.StatementSection == StatementSectionCashFlow &&
		rg.CashFlowCategory == nil {
		errs = append(errs, ValidationError{
			Field:   "cash_flow_category",
			Message: "cash_flow_category is required for cash flow statement groups",
			Code:    "REQUIRED_FIELD", Severity: ValidationSeverityError,
		})
	}
	if rg.DisplayOrder < 0 {
		errs = append(errs, ValidationError{
			Field: "display_order", Message: "display order must be non-negative",
			Code: "OUT_OF_RANGE", Severity: ValidationSeverityError,
		})
	}
	if rg.IndentLevel < 0 || rg.IndentLevel > AccountGroupIndentLevelMax {
		errs = append(errs, ValidationError{
			Field:   "indent_level",
			Message: fmt.Sprintf("indent level must be between 0 and %d", AccountGroupIndentLevelMax),
			Code:    "OUT_OF_RANGE", Severity: ValidationSeverityError,
		})
	}
	if rg.IsHeader && rg.ShowTotals {
		errs = append(errs, ValidationError{
			Field:   "show_totals",
			Message: "header groups are label rows and must not show totals",
			Code:    "INCONSISTENT_VALUE", Severity: ValidationSeverityWarning,
		})
	}

	return errs
}

// =============================================================================
// Request DTOs for ReportingGroup
// =============================================================================

// CreateReportingGroupRequest is the inbound payload for creating a new reporting group.
// Level and Path are derived by the service from ParentGroupID.
type CreateReportingGroupRequest struct {
	EntityID          *uuid.UUID `json:"entity_id,omitempty"`
	ReportingSchemeID uuid.UUID  `json:"reporting_scheme_id" validate:"required"`
	GroupCode         string     `json:"group_code"          validate:"required,max=50"`
	GroupName         string     `json:"group_name"          validate:"required,max=255"`
	Description       *string    `json:"description,omitempty" validate:"omitempty,max=500"`
	ParentGroupID     *uuid.UUID `json:"parent_group_id,omitempty"`

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
func (r *CreateReportingGroupRequest) Validate() []ValidationError {
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
			Field:   "financial_statement_section",
			Message: fmt.Sprintf("invalid statement section: %q", *r.StatementSection),
			Code:    "INVALID_VALUE", Severity: ValidationSeverityError,
		})
	}
	if r.StatementSection != nil &&
		*r.StatementSection == StatementSectionCashFlow &&
		r.CashFlowCategory == nil {
		errs = append(errs, ValidationError{
			Field:   "cash_flow_category",
			Message: "cash_flow_category is required for cash flow statement groups",
			Code:    "REQUIRED_FIELD", Severity: ValidationSeverityError,
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

// UpdateReportingGroupRequest is the partial-update payload.
// GroupCode and ReportingSchemeID are excluded — code changes require a uniqueness
// re-check; scheme changes require remapping all AccountMappings.
// ParentGroupID is excluded — reparenting requires a full subtree re-path via
// a dedicated Reparent operation.
type UpdateReportingGroupRequest struct {
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
// AccountGroup — request DTOs and validation
// =============================================================================

// Validate performs domain-level consistency checks on a simple AccountGroup.
func (ag *AccountGroup) Validate() []ValidationError {
	var errs []ValidationError
	if ag.TenantID == uuid.Nil {
		errs = append(errs, ValidationError{
			Field: "tenant_id", Message: "tenant_id is required",
			Code: "REQUIRED_FIELD", Severity: ValidationSeverityError,
		})
	}
	if strings.TrimSpace(ag.GroupCode) == "" {
		errs = append(errs, ValidationError{
			Field: "group_code", Message: "group code is required",
			Code: "REQUIRED_FIELD", Severity: ValidationSeverityError,
		})
	} else if len(ag.GroupCode) > AccountGroupCodeMaxLen {
		errs = append(errs, ValidationError{
			Field:   "group_code",
			Message: fmt.Sprintf("group code must be %d characters or less", AccountGroupCodeMaxLen),
			Code:    "MAX_LENGTH_EXCEEDED", Severity: ValidationSeverityError,
		})
	}
	if strings.TrimSpace(ag.GroupName) == "" {
		errs = append(errs, ValidationError{
			Field: "group_name", Message: "group name is required",
			Code: "REQUIRED_FIELD", Severity: ValidationSeverityError,
		})
	} else if len(ag.GroupName) > AccountGroupNameMaxLen {
		errs = append(errs, ValidationError{
			Field:   "group_name",
			Message: fmt.Sprintf("group name must be %d characters or less", AccountGroupNameMaxLen),
			Code:    "MAX_LENGTH_EXCEEDED", Severity: ValidationSeverityError,
		})
	}
	if !ag.RootType.IsValid() {
		errs = append(errs, ValidationError{
			Field:   "root_type",
			Message: fmt.Sprintf("invalid root type: %q", ag.RootType),
			Code:    "INVALID_VALUE", Severity: ValidationSeverityError,
		})
	}
	return errs
}

// CreateAccountGroupRequest is the inbound payload for creating a simple AccountGroup
// backed by the finance_account_groups DB table.
type CreateAccountGroupRequest struct {
	EntityID      *uuid.UUID `json:"entity_id,omitempty"`
	GroupCode     string     `json:"group_code"   validate:"required,max=50"`
	GroupName     string     `json:"group_name"   validate:"required,max=255"`
	Description   string     `json:"description,omitempty" validate:"omitempty,max=500"`
	ParentGroupID *uuid.UUID `json:"parent_group_id,omitempty"`

	// Classification
	RootType      RootType `json:"root_type" validate:"required"`
	GroupCategory string   `json:"group_category,omitempty"`

	// Financial statement placement
	FinancialStatementSection *string              `json:"financial_statement_section,omitempty"`
	ConsolidationMethod       *ConsolidationMethod `json:"consolidation_method,omitempty"`
	CashFlowCategory          *CashFlowCategory    `json:"cash_flow_category,omitempty"`

	// Presentation
	DisplayOrder int  `json:"display_order"`
	IsHeader     bool `json:"is_header"`
	ShowTotals   bool `json:"show_totals"`
	IndentLevel  int  `json:"indent_level"`
	BoldDisplay  bool `json:"bold_display"`
}

// UpdateAccountGroupRequest is the partial-update payload for a simple AccountGroup.
// All fields are optional; nil means keep existing value.
type UpdateAccountGroupRequest struct {
	GroupName   *string `json:"group_name,omitempty"   validate:"omitempty,max=255"`
	Description *string `json:"description,omitempty"  validate:"omitempty,max=500"`

	// Financial statement placement
	FinancialStatementSection *string              `json:"financial_statement_section,omitempty"`
	ConsolidationMethod       *ConsolidationMethod `json:"consolidation_method,omitempty"`
	CashFlowCategory          *CashFlowCategory    `json:"cash_flow_category,omitempty"`

	// Presentation
	DisplayOrder *int  `json:"display_order,omitempty"`
	IsActive     *bool `json:"is_active,omitempty"`
	IsHeader     *bool `json:"is_header,omitempty"`
	ShowTotals   *bool `json:"show_totals,omitempty"`
	IndentLevel  *int  `json:"indent_level,omitempty"`
	BoldDisplay  *bool `json:"bold_display,omitempty"`
}

// =============================================================================
// ReportingGroupFilter is the query predicate for listing reporting groups.
// =============================================================================

// ReportingGroupFilter is the query predicate for listing reporting groups.
type ReportingGroupFilter struct {
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

	Limit     *int    `json:"limit,omitempty"`
	Offset    *int    `json:"offset,omitempty"`
	SortBy    *string `json:"sort_by,omitempty"`
	SortOrder *string `json:"sort_order,omitempty"`
}

// =============================================================================
// View and analytics types (future reporting layer)
// These types are defined here to satisfy service/repository interfaces.
// Full implementations will be added with dedicated DB queries.
// =============================================================================

// UnifiedFilter is the predicate for unified account+group queries.
type UnifiedFilter struct {
	EntityID    *uuid.UUID `json:"entity_id,omitempty"`
	IsActive    *bool      `json:"is_active,omitempty"`
	SearchQuery *string    `json:"search_query,omitempty"`
	Limit       *int       `json:"limit,omitempty"`
	Offset      *int       `json:"offset,omitempty"`
}

// AccountNode is a lightweight node in a unified account/group tree.
type AccountNode struct {
	ID       uuid.UUID  `json:"id"`
	TenantID uuid.UUID  `json:"tenant_id"`
	EntityID *uuid.UUID `json:"entity_id,omitempty"`
	Code     string     `json:"code"`
	Name     string     `json:"name"`
	IsGroup  bool       `json:"is_group"`
	IsActive bool       `json:"is_active"`
	ParentID *uuid.UUID `json:"parent_id,omitempty"`
	Level    int        `json:"level"`
}

// AccountWithGroups combines a ledger account with its group metadata.
type AccountWithGroups struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	AccountCode string    `json:"account_code"`
	AccountName string    `json:"account_name"`
	RootType    RootType  `json:"root_type"`
	AccountType string    `json:"account_type"`
	GroupCode   *string   `json:"group_code,omitempty"`
	GroupName   *string   `json:"group_name,omitempty"`
	IsActive    bool      `json:"is_active"`
	IsLeaf      bool      `json:"is_leaf"`
}

// ChartOfAccountsComplete is the enriched account record used in reporting views.
type ChartOfAccountsComplete struct {
	ID               uuid.UUID         `json:"id"`
	TenantID         uuid.UUID         `json:"tenant_id"`
	AccountCode      string            `json:"account_code"`
	AccountName      string            `json:"account_name"`
	RootType         RootType          `json:"root_type"`
	AccountType      string            `json:"account_type"`
	GroupCode        *string           `json:"group_code,omitempty"`
	GroupName        *string           `json:"group_name,omitempty"`
	StatementSection *string           `json:"statement_section,omitempty"`
	CashFlowCategory *CashFlowCategory `json:"cash_flow_category,omitempty"`
	IsActive         bool              `json:"is_active"`
	IsLeaf           bool              `json:"is_leaf"`
	AccountLevel     int32             `json:"account_level"`
}

// TrialBalanceSummary is an enriched trial balance line for the reporting layer.
type TrialBalanceSummary struct {
	AccountID   uuid.UUID `json:"account_id"`
	AccountCode string    `json:"account_code"`
	AccountName string    `json:"account_name"`
	RootType    RootType  `json:"root_type"`
	GroupCode   *string   `json:"group_code,omitempty"`
	GroupName   *string   `json:"group_name,omitempty"`
}

// BalanceFilter is the predicate for balance-based account queries.
type BalanceFilter struct {
	EntityID    *uuid.UUID `json:"entity_id,omitempty"`
	RootType    *RootType  `json:"root_type,omitempty"`
	NonZeroOnly *bool      `json:"non_zero_only,omitempty"`
	Limit       *int       `json:"limit,omitempty"`
	Offset      *int       `json:"offset,omitempty"`
}

// CashFlowAccount is the view type returned by GetCashFlowAccounts.
type CashFlowAccount struct {
	AccountID        uuid.UUID        `json:"account_id"`
	AccountCode      string           `json:"account_code"`
	AccountName      string           `json:"account_name"`
	CashFlowCategory CashFlowCategory `json:"cash_flow_category"`
}

// AccountGroupSummary is an aggregated balance summary grouped by account group.
type AccountGroupSummary struct {
	GroupID   uuid.UUID `json:"group_id"`
	GroupCode string    `json:"group_code"`
	GroupName string    `json:"group_name"`
}

// AccountHierarchy represents an account with its position in the hierarchy tree.
type AccountHierarchy struct {
	ID              uuid.UUID       `json:"id"`
	TenantID        uuid.UUID       `json:"tenant_id"`
	AccountCode     string          `json:"account_code"`
	AccountName     string          `json:"account_name"`
	ParentAccountID *uuid.UUID      `json:"parent_account_id,omitempty"`
	RootType        string          `json:"root_type"`
	AccountType     string          `json:"account_type"`
	NormalBalance   string          `json:"normal_balance"`
	CurrentBalance  decimal.Decimal `json:"current_balance"`
	Level           int32           `json:"level"`
	FullPath        string          `json:"full_path"`
	FullName        string          `json:"full_name"`
	ChildCount      int64           `json:"child_count"`
}

// AccountActivityFilter is the predicate for activity-based account queries.
type AccountActivityFilter struct {
	EntityID        *uuid.UUID       `json:"entity_id,omitempty"`
	MinEntries      *int32           `json:"min_entries,omitempty"`
	MinDaysInactive *int32           `json:"min_days_inactive,omitempty"`
	MinBalance      *decimal.Decimal `json:"min_balance,omitempty"`
	ActivityLevel   *string          `json:"activity_level,omitempty"`
	Limit           *int             `json:"limit,omitempty"`
	Offset          *int             `json:"offset,omitempty"`
}

// AccountActivity is an account enriched with its recent transaction activity.
type AccountActivity struct {
	TenantID            uuid.UUID       `json:"tenant_id"`
	AccountID           uuid.UUID       `json:"account_id"`
	AccountCode         string          `json:"account_code"`
	AccountName         string          `json:"account_name"`
	CurrentBalance      decimal.Decimal `json:"current_balance"`
	LastTransactionDate *time.Time      `json:"last_transaction_date,omitempty"`
	TotalEntries        int64           `json:"total_entries"`
	EntriesLast30Days   int64           `json:"entries_last_30_days"`
	DebitsLast30Days    decimal.Decimal `json:"debits_last_30_days"`
	CreditsLast30Days   decimal.Decimal `json:"credits_last_30_days"`
}

// AccountActivitySummary provides categorised activity analysis for an account.
type AccountActivitySummary struct {
	AccountID           uuid.UUID       `json:"account_id"`
	AccountCode         string          `json:"account_code"`
	AccountName         string          `json:"account_name"`
	CurrentBalance      decimal.Decimal `json:"current_balance"`
	EntriesLast30Days   int64           `json:"entries_last_30_days"`
	DebitsLast30Days    decimal.Decimal `json:"debits_last_30_days"`
	CreditsLast30Days   decimal.Decimal `json:"credits_last_30_days"`
	TotalActivity30Days decimal.Decimal `json:"total_activity_30_days"`
	ActivityLevel       string          `json:"activity_level"`
}
