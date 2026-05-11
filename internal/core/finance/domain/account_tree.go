// Package domain  contains read-side projections for the chart of accounts
// and reporting hierarchy. Nothing in this package participates in the write
// path. These types are assembled by query handlers from the domain entities;
// they should never be passed to service methods that create or mutate data.
//
// The separation exists because the shape of data required for rendering a
// tree in the UI (flattened, annotated with children counts, carrying both
// ledger and reporting labels) is fundamentally different from — and more
// volatile than — the core domain entities. Keeping these projections here
// prevents the domain package from growing UI-driven fields.
package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// =============================================================================
// AccountTreeNode
// =============================================================================

// AccountTreeNode is the unified read projection used to render the chart of
// accounts tree in the UI and to drive tree-picker components.
//
// It is assembled by the ChartOfAccountsQueryHandler from Account and
// ReportingGroup data. It must never be stored or passed to write-side services.
//
// IsGroup discriminates the payload:
//   - IsGroup == false → AccountData is populated; GroupData is nil.
//   - IsGroup == true  → GroupData is populated; AccountData is nil.
type AccountTreeNode struct {
	// Common identity fields
	ID       uuid.UUID  `json:"id"`
	TenantID uuid.UUID  `json:"tenant_id"`
	EntityID *uuid.UUID `json:"entity_id,omitempty"`
	Code     string     `json:"code"`
	Name     string     `json:"name"`
	IsActive bool       `json:"is_active"`

	// IsGroup is true when this node represents an AccountGroup; false for an Account.
	IsGroup bool `json:"is_group"`

	// Hierarchy metadata — always populated regardless of node type.
	ParentID    *uuid.UUID       `json:"parent_id,omitempty"`
	Level       int              `json:"level"`
	Path        MaterialisedPath `json:"path"`
	HasChildren bool             `json:"has_children"`
	ChildCount  int              `json:"child_count"`

	// AccountData is populated when IsGroup == false.
	AccountData *AccountNodeData `json:"account,omitempty"`

	// GroupData is populated when IsGroup == true.
	GroupData *GroupNodeData `json:"group,omitempty"`

	// Audit timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	CreatedBy uuid.UUID  `json:"created_by"`
	UpdatedBy *uuid.UUID `json:"updated_by,omitempty"`
}

// AccountNodeData carries account-specific projection fields.
// These are derived from the Account entity and cached balance tables.
type AccountNodeData struct {
	// Classification
	RootType      RootType      `json:"root_type"`
	AccountType   string        `json:"account_type"`
	NormalBalance NormalBalance `json:"normal_balance"`

	// Control & structural flags
	IsControlAccount bool `json:"is_control_account"`
	IsLeaf           bool `json:"is_leaf"`

	// Currency
	CurrencyCode    *string `json:"currency_code,omitempty"`
	IsMultiCurrency bool    `json:"is_multi_currency"`

	// Cached balance — derived from the journal engine, not the Account entity.
	// Always reflect a point-in-time snapshot; never use for accounting calculations.
	CurrentBalance decimal.Decimal `json:"current_balance"`

	// Status
	Status           AccountStatus    `json:"status"`
	ValidationStatus ValidationStatus `json:"validation_status"`

	// Reconciliation
	RequiresReconciliation bool       `json:"requires_reconciliation"`
	LastReconciledAt       *time.Time `json:"last_reconciled_at,omitempty"`

	// Tax
	TaxCode *string          `json:"tax_code,omitempty"`
	TaxRate *decimal.Decimal `json:"tax_rate,omitempty"`
}

// GroupNodeData carries reporting-group-specific projection fields.
// These are derived from the ReportingGroup entity.
type GroupNodeData struct {
	ReportingSchemeID uuid.UUID `json:"reporting_scheme_id"`
	SchemeCode        string    `json:"scheme_code"` // Denormalised from ReportingScheme for display
	SchemeName        string    `json:"scheme_name"`

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

	// Aggregated balance across all mapped accounts — computed by the query handler.
	AggregatedBalance decimal.Decimal `json:"aggregated_balance"`
}

// =============================================================================
// AccountNodeList
// =============================================================================

// AccountNodeList is a paginated list of AccountTreeNodes returned by list queries.
type AccountNodeList struct {
	Nodes      []*AccountTreeNode `json:"nodes"`
	TotalCount int                `json:"total_count"`
	HasMore    bool               `json:"has_more"`
	PageInfo   *PageInfo          `json:"page_info,omitempty"`
}

// PageInfo provides pagination context for list responses.
type PageInfo struct {
	Page       int  `json:"page"`
	PageSize   int  `json:"page_size"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_previous"`
}

// =============================================================================
// ChartOfAccountsView
// =============================================================================

// ChartOfAccountsView is a flattened, denormalised row used to populate the
// full chart of accounts report. It is backed by a database view that joins
// accounts, groups, and balance snapshots.
//
// This type is read-only and is never used as a write payload.
type ChartOfAccountsView struct {
	AccountID   uuid.UUID `json:"account_id"`
	AccountCode string    `json:"account_code"`
	AccountName string    `json:"account_name"`

	// Classification
	RootType      string `json:"root_type"`
	AccountType   string `json:"account_type"`
	NormalBalance string `json:"normal_balance"`

	// Hierarchy
	ParentAccountID *uuid.UUID `json:"parent_account_id,omitempty"`
	AccountLevel    int32      `json:"account_level"`
	IsLeaf          bool       `json:"is_leaf"`

	// Balance
	CurrentBalance decimal.Decimal `json:"current_balance"`

	// Reporting group (from the primary/internal scheme mapping)
	GroupCode  *string `json:"group_code,omitempty"`
	GroupName  *string `json:"group_name,omitempty"`
	SchemeCode *string `json:"scheme_code,omitempty"`

	// Statement placement
	StatementSection       *string `json:"statement_section,omitempty"`
	CashFlowClassification *string `json:"cash_flow_classification,omitempty"`

	// Display
	DisplayOrder  int32 `json:"display_order"`
	ShowInReports bool  `json:"show_in_reports"`
	IsActive      bool  `json:"is_active"`

	TenantID  uuid.UUID `json:"tenant_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// =============================================================================
// Query filters
// =============================================================================

// TreeFilter is the query predicate for fetching an AccountTreeNode list.
// It covers both account nodes and group nodes in a unified query.
type TreeFilter struct {
	EntityID          *uuid.UUID `json:"entity_id,omitempty"`
	ReportingSchemeID *uuid.UUID `json:"reporting_scheme_id,omitempty"`
	IsGroup           *bool      `json:"is_group,omitempty"` // nil = both accounts and groups
	ParentID          *uuid.UUID `json:"parent_id,omitempty"`
	IsActive          *bool      `json:"is_active,omitempty"`
	SearchQuery       *string    `json:"search_query,omitempty"`

	// Account-specific predicates
	RootType    *RootType `json:"root_type,omitempty"`
	AccountType *string   `json:"account_type,omitempty"`
	IsLeaf      *bool     `json:"is_leaf,omitempty"`

	// Group-specific predicates
	StatementSection *StatementSection `json:"financial_statement_section,omitempty"`
	CashFlowCategory *CashFlowCategory `json:"cash_flow_category,omitempty"`

	// Hierarchy depth limit
	MaxLevel        *int `json:"max_level,omitempty"`
	IncludeChildren bool `json:"include_children"`

	// Pagination
	Limit  *int `json:"limit,omitempty"`
	Offset *int `json:"offset,omitempty"`

	// Sorting
	SortBy    *string `json:"sort_by,omitempty"`
	SortOrder *string `json:"sort_order,omitempty"`
}

// ChartOfAccountsFilter is the predicate for the full chart of accounts view.
type ChartOfAccountsFilter struct {
	EntityID          *uuid.UUID        `json:"entity_id,omitempty"`
	ReportingSchemeID *uuid.UUID        `json:"reporting_scheme_id,omitempty"`
	StatementSection  *StatementSection `json:"statement_section,omitempty"`
	IncludeInactive   *bool             `json:"include_inactive,omitempty"`
	NonZeroOnly       *bool             `json:"non_zero_only,omitempty"`
	Limit             *int              `json:"limit,omitempty"`
	Offset            *int              `json:"offset,omitempty"`
}
