package domain

import (
	"time"

	"github.com/google/uuid"
	"awo/internal/shared/errors"
	"github.com/shopspring/decimal"
)

// AccountNode represents a unified type for both accounts and account groups
// Uses discriminator pattern for type safety and clear API contracts
type AccountNode struct {
	// Common fields
	ID          uuid.UUID  `json:"id"`
	Code        string     `json:"code"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	IsActive    bool       `json:"is_active"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`

	// Discriminator field to identify the type
	NodeType AccountNodeType `json:"node_type"`

	// Account-specific fields (populated only when NodeType = "account")
	Account *AccountNodeData `json:"account,omitempty"`

	// AccountGroup-specific fields (populated only when NodeType = "group")
	Group *GroupNodeData `json:"group,omitempty"`

	// Hierarchy metadata
	Level       int    `json:"level"`
	HasChildren bool   `json:"has_children"`
	ChildCount  int    `json:"child_count"`
	Path        string `json:"path"` // Materialized path for hierarchy

	// Audit fields
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedBy uuid.UUID `json:"created_by"`
	UpdatedBy uuid.UUID `json:"updated_by"`
}

// AccountNodeType discriminates between accounts and groups
type AccountNodeType string

const (
	NodeTypeAccount AccountNodeType = "account"
	NodeTypeGroup   AccountNodeType = "group"
)

// AccountNodeData contains account-specific information
type AccountNodeData struct {
	AccountType      string          `json:"account_type"`
	RootType         RootType        `json:"root_type"`
	NormalBalance    string          `json:"normal_balance"`
	IsControlAccount bool            `json:"is_control_account"`
	IsGroup          bool            `json:"is_group"`
	CurrencyCode     string          `json:"currency_code"`
	CurrentBalance   decimal.Decimal `json:"current_balance"`

	// Tax information
	TaxCode *string          `json:"tax_code,omitempty"`
	TaxRate *decimal.Decimal `json:"tax_rate,omitempty"`

	// Reconciliation
	RequiresReconciliation bool       `json:"requires_reconciliation"`
	LastReconciledAt       *time.Time `json:"last_reconciled_at,omitempty"`
}

// GroupNodeData contains account group-specific information
type GroupNodeData struct {
	GroupType                 string               `json:"group_type"`
	FinancialStatementSection *string              `json:"financial_statement_section,omitempty"`
	ConsolidationMethod       *ConsolidationMethod `json:"consolidation_method,omitempty"`
	CashFlowCategory          *CashFlowCategory    `json:"cash_flow_category,omitempty"`
	DisplayOrder              int                  `json:"display_order"`
	IsHeader                  bool                 `json:"is_header"`
	ShowTotals                bool                 `json:"show_totals"`
	IndentLevel               int                  `json:"indent_level"`
	BoldDisplay               bool                 `json:"bold_display"`
}

// ConsolidationMethod defines how group balances are calculated
type ConsolidationMethod string

const (
	ConsolidationSum     ConsolidationMethod = "SUM"
	ConsolidationAverage ConsolidationMethod = "AVERAGE"
	ConsolidationMax     ConsolidationMethod = "MAX"
	ConsolidationMin     ConsolidationMethod = "MIN"
	ConsolidationCustom  ConsolidationMethod = "CUSTOM"
)

// CashFlowCategory is defined in types.go to avoid redeclaration

// AccountGroup represents account grouping for financial statements
type AccountGroup struct {
	ID                        uuid.UUID            `json:"id"`
	EntityID                  *uuid.UUID           `json:"entity_id,omitempty"`
	GroupCode                 string               `json:"group_code"`
	GroupName                 string               `json:"group_name"`
	Description               *string              `json:"description,omitempty"`
	GroupType                 string               `json:"group_type"`
	ParentGroupID             *uuid.UUID           `json:"parent_group_id,omitempty"`
	FinancialStatementSection *string              `json:"financial_statement_section,omitempty"`
	ConsolidationMethod       *ConsolidationMethod `json:"consolidation_method,omitempty"`
	CashFlowCategory          *CashFlowCategory    `json:"cash_flow_category,omitempty"`
	DisplayOrder              int                  `json:"display_order"`
	IsSystemDefined           bool                 `json:"is_system_defined"`
	IsActive                  bool                 `json:"is_active"`
	IsHeader                  bool                 `json:"is_header"`
	ShowTotals                bool                 `json:"show_totals"`
	IndentLevel               int                  `json:"indent_level"`
	BoldDisplay               bool                 `json:"bold_display"`

	// Hierarchy support (materialized path)
	Path  string `json:"path"`
	Level int    `json:"level"`

	// Audit fields
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedBy uuid.UUID `json:"created_by"`
	UpdatedBy uuid.UUID `json:"updated_by"`
	Version   int       `json:"version"`
}

// Request/Response Types for Account Groups

// CreateAccountGroupRequest for creating new account groups
type CreateAccountGroupRequest struct {
	EntityID                  *uuid.UUID           `json:"entity_id,omitempty"`
	GroupCode                 string               `json:"group_code" validate:"required,max=50"`
	GroupName                 string               `json:"group_name" validate:"required,max=255"`
	Description               *string              `json:"description,omitempty" validate:"omitempty,max=500"`
	GroupType                 string               `json:"group_type" validate:"required"`
	ParentGroupID             *uuid.UUID           `json:"parent_group_id,omitempty"`
	FinancialStatementSection *string              `json:"financial_statement_section,omitempty"`
	ConsolidationMethod       *ConsolidationMethod `json:"consolidation_method,omitempty"`
	CashFlowCategory          *CashFlowCategory    `json:"cash_flow_category,omitempty"`
	DisplayOrder              int                  `json:"display_order"`
	IsHeader                  bool                 `json:"is_header"`
	ShowTotals                bool                 `json:"show_totals"`
	IndentLevel               int                  `json:"indent_level"`
	BoldDisplay               bool                 `json:"bold_display"`
}

// UpdateAccountGroupRequest for updating existing account groups
type UpdateAccountGroupRequest struct {
	GroupName                 *string              `json:"group_name,omitempty" validate:"omitempty,max=255"`
	Description               *string              `json:"description,omitempty" validate:"omitempty,max=500"`
	GroupType                 *string              `json:"group_type,omitempty"`
	ParentGroupID             *uuid.UUID           `json:"parent_group_id,omitempty"`
	FinancialStatementSection *string              `json:"financial_statement_section,omitempty"`
	ConsolidationMethod       *ConsolidationMethod `json:"consolidation_method,omitempty"`
	CashFlowCategory          *CashFlowCategory    `json:"cash_flow_category,omitempty"`
	DisplayOrder              *int                 `json:"display_order,omitempty"`
	IsActive                  *bool                `json:"is_active,omitempty"`
	IsHeader                  *bool                `json:"is_header,omitempty"`
	ShowTotals                *bool                `json:"show_totals,omitempty"`
	IndentLevel               *int                 `json:"indent_level,omitempty"`
	BoldDisplay               *bool                `json:"bold_display,omitempty"`
}

// Filter types

// AccountGroupFilter for filtering account groups
type AccountGroupFilter struct {
	EntityID                  *uuid.UUID        `json:"entity_id,omitempty"`
	GroupType                 *string           `json:"group_type,omitempty"`
	ParentGroupID             *uuid.UUID        `json:"parent_group_id,omitempty"`
	FinancialStatementSection *string           `json:"financial_statement_section,omitempty"`
	CashFlowCategory          *CashFlowCategory `json:"cash_flow_category,omitempty"`
	IsActive                  *bool             `json:"is_active,omitempty"`
	IsSystemDefined           *bool             `json:"is_system_defined,omitempty"`
	IsHeader                  *bool             `json:"is_header,omitempty"`

	// Pagination
	Limit  *int `json:"limit,omitempty"`
	Offset *int `json:"offset,omitempty"`

	// Sorting
	SortBy    *string `json:"sort_by,omitempty"`
	SortOrder *string `json:"sort_order,omitempty"`
}

// UnifiedFilter for filtering both accounts and groups
type UnifiedFilter struct {
	EntityID    *uuid.UUID        `json:"entity_id,omitempty"`
	NodeTypes   []AccountNodeType `json:"node_types,omitempty"` // Filter by account, group, or both
	ParentID    *uuid.UUID        `json:"parent_id,omitempty"`
	IsActive    *bool             `json:"is_active,omitempty"`
	SearchQuery *string           `json:"search_query,omitempty"`

	// Account-specific filters
	AccountType *string   `json:"account_type,omitempty"`
	RootType    *RootType `json:"root_type,omitempty"`

	// Group-specific filters
	GroupType                 *string           `json:"group_type,omitempty"`
	FinancialStatementSection *string           `json:"financial_statement_section,omitempty"`
	CashFlowCategory          *CashFlowCategory `json:"cash_flow_category,omitempty"`

	// Hierarchy filters
	MaxLevel        *int `json:"max_level,omitempty"`
	IncludeChildren bool `json:"include_children"`

	// Pagination
	Limit  *int `json:"limit,omitempty"`
	Offset *int `json:"offset,omitempty"`

	// Sorting
	SortBy    *string `json:"sort_by,omitempty"`
	SortOrder *string `json:"sort_order,omitempty"`
}

// Validation methods

// Validate performs business rule validation for AccountGroup
func (ag *AccountGroup) Validate() error {
	var errs errors.ValidationErrors

	// Required fields
	if ag.GroupCode == "" {
		errs.Add("group_code", "Group code is required")
	}
	if ag.GroupName == "" {
		errs.Add("group_name", "Group name is required")
	}

	// Code format validation
	if len(ag.GroupCode) > 50 {
		errs.Add("group_code", "Group code must be 50 characters or less")
	}

	// Name length validation
	if len(ag.GroupName) > 255 {
		errs.Add("group_name", "Group name must be 255 characters or less")
	}

	// Hierarchy validation
	if ag.ParentGroupID != nil && *ag.ParentGroupID == ag.ID {
		errs.Add("parent_group_id", "Group cannot be its own parent")
	}

	// Display order validation
	if ag.DisplayOrder < 0 {
		errs.Add("display_order", "Display order must be non-negative")
	}

	// Indent level validation
	if ag.IndentLevel < 0 || ag.IndentLevel > 10 {
		errs.Add("indent_level", "Indent level must be between 0 and 10")
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// Business methods

// IsRoot returns true if this is a root-level group
func (ag *AccountGroup) IsRoot() bool {
	return ag.ParentGroupID == nil
}

// CanHaveChildren returns true if this group can have child groups
func (ag *AccountGroup) CanHaveChildren() bool {
	return ag.IsActive && !ag.IsSystemDefined
}

// GetDisplayName returns the formatted name for display
func (ag *AccountGroup) GetDisplayName() string {
	if ag.BoldDisplay {
		return "**" + ag.GroupName + "**"
	}
	return ag.GroupName
}

// AccountNodeList represents a collection of account nodes with metadata
type AccountNodeList struct {
	Nodes      []*AccountNode `json:"nodes"`
	TotalCount int            `json:"total_count"`
	HasMore    bool           `json:"has_more"`
	PageInfo   *PageInfo      `json:"page_info,omitempty"`
}

// PageInfo provides pagination metadata
type PageInfo struct {
	Page       int  `json:"page"`
	PageSize   int  `json:"page_size"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_previous"`
}
