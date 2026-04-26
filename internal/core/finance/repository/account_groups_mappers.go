package repository

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	db "awo.so/db/sqlc"
	"awo.so/internal/core/finance/domain"
	"awo.so/internal/shared/errors"
)

// Account Group Creation Mapping

func (r *accountsRepository) mapDomainAccountGroupToCreateParams(accountGroup *domain.AccountGroup) (db.CreateAccountGroupParams, error) {
	// Calculate hierarchy values if parent is specified
	groupLevel := int32(1)
	groupPath := "/" + accountGroup.GroupCode

	if accountGroup.ParentGroupID != nil {
		// TODO: Calculate proper group level and path based on parent
		// For now, use defaults
		groupLevel = 2
		groupPath = "/parent/" + accountGroup.GroupCode
	}

	// Map consolidation method
	var consolidationMethod *string
	if accountGroup.ConsolidationMethod != nil {
		method := string(*accountGroup.ConsolidationMethod)
		consolidationMethod = &method
	}

	// Map cash flow category
	var cashFlowCategory *string
	if accountGroup.CashFlowCategory != nil {
		category := string(*accountGroup.CashFlowCategory)
		cashFlowCategory = &category
	}

	// Set default values for required fields
	displayFormat := "STANDARD"
	if accountGroup.DisplayOrder > 0 {
		displayFormat = "CUSTOM"
	}

	// Convert Description string → *string (nullable DB column)
	var groupDescription *string
	if accountGroup.Description != "" {
		d := accountGroup.Description
		groupDescription = &d
	}

	// Convert GroupCategory string → *string
	var groupCategory *string
	if accountGroup.GroupCategory != "" {
		gc := accountGroup.GroupCategory
		groupCategory = &gc
	}

	return db.CreateAccountGroupParams{
		EntityID:                  accountGroup.EntityID,
		GroupCode:                 accountGroup.GroupCode,
		GroupName:                 accountGroup.GroupName,
		GroupDescription:          groupDescription,
		ParentGroupID:             accountGroup.ParentGroupID,
		GroupLevel:                groupLevel,
		GroupPath:                 &groupPath,
		RootType:                  string(accountGroup.RootType),
		GroupCategory:             groupCategory,
		FinancialStatementSection: accountGroup.FinancialStatementSection,
		ConsolidationMethod:       consolidationMethod,
		CashFlowCategory:    cashFlowCategory,
		StatementOrder:      func() *int32 { v := int32(accountGroup.DisplayOrder); return &v }(),
		DisplayFormat:       &displayFormat,
		IndentLevel:         func() *int32 { v := int32(accountGroup.IndentLevel); return &v }(),
		ShowTotals:          &accountGroup.ShowTotals,
		BoldDisplay:         &accountGroup.BoldDisplay,
		IsActive:            accountGroup.IsActive,
		CreatedBy:           nil, // Will be set by context in database
	}, nil
}

// Account Group Update Mapping

func (r *accountsRepository) mapDomainAccountGroupToUpdateParams(id uuid.UUID, accountGroup *domain.AccountGroup) db.UpdateAccountGroupParams {
	// Map consolidation method
	var consolidationMethod *string
	if accountGroup.ConsolidationMethod != nil {
		method := string(*accountGroup.ConsolidationMethod)
		consolidationMethod = &method
	}

	// Map cash flow category
	var cashFlowCategory *string
	if accountGroup.CashFlowCategory != nil {
		category := string(*accountGroup.CashFlowCategory)
		cashFlowCategory = &category
	}

	var updGroupDescription *string
	if accountGroup.Description != "" {
		d := accountGroup.Description
		updGroupDescription = &d
	}

	return db.UpdateAccountGroupParams{
		GroupID:                   id,
		EntityID:                  nil, // Allow cross-entity updates within tenant
		GroupName:                 &accountGroup.GroupName,
		GroupDescription:          updGroupDescription,
		ParentGroupID:             accountGroup.ParentGroupID,
		FinancialStatementSection: accountGroup.FinancialStatementSection,
		ConsolidationMethod:       consolidationMethod,
		CashFlowCategory:          cashFlowCategory,
		StatementOrder:            func() *int32 { v := int32(accountGroup.DisplayOrder); return &v }(),
		IndentLevel:               func() *int32 { v := int32(accountGroup.IndentLevel); return &v }(),
		ShowTotals:                &accountGroup.ShowTotals,
		BoldDisplay:               &accountGroup.BoldDisplay,
		IsActive:                  &accountGroup.IsActive,
		UpdatedBy:                 nil, // Will be set by context in database
	}
}

// SQLC to Domain Mapping

func (r *accountsRepository) mapSQLCAccountGroupToDomain(sqlcGroup *db.FinanceAccountGroup, domainGroup *domain.AccountGroup) error {
	// Basic fields
	domainGroup.ID = sqlcGroup.ID
	domainGroup.EntityID = sqlcGroup.EntityID
	domainGroup.GroupCode = sqlcGroup.GroupCode
	domainGroup.GroupName = sqlcGroup.GroupName
	if sqlcGroup.GroupDescription != nil {
		domainGroup.Description = *sqlcGroup.GroupDescription
	}
	domainGroup.RootType = domain.RootType(sqlcGroup.RootType)
	if sqlcGroup.GroupCategory != nil {
		domainGroup.GroupCategory = *sqlcGroup.GroupCategory
	}
	domainGroup.ParentGroupID = sqlcGroup.ParentGroupID
	domainGroup.FinancialStatementSection = sqlcGroup.FinancialStatementSection
	domainGroup.DisplayOrder = int(getInt32Value(sqlcGroup.StatementOrder))
	domainGroup.IsSystemDefined = sqlcGroup.IsSystemGroup
	domainGroup.IsActive = sqlcGroup.IsActive
	domainGroup.IsHeader = getBoolValue(sqlcGroup.ShowInSummary) // Map show_in_summary to IsHeader
	domainGroup.ShowTotals = getBoolValue(sqlcGroup.ShowTotals)
	domainGroup.IndentLevel = int(getInt32Value(sqlcGroup.IndentLevel))
	domainGroup.BoldDisplay = getBoolValue(sqlcGroup.BoldDisplay)

	// Map consolidation method
	if sqlcGroup.ConsolidationMethod != nil {
		method := domain.ConsolidationMethod(*sqlcGroup.ConsolidationMethod)
		domainGroup.ConsolidationMethod = &method
	}

	// Map cash flow category
	if sqlcGroup.CashFlowCategory != nil {
		category := domain.CashFlowCategory(*sqlcGroup.CashFlowCategory)
		domainGroup.CashFlowCategory = &category
	}

	// Hierarchy fields
	if sqlcGroup.GroupPath != nil {
		domainGroup.Path = *sqlcGroup.GroupPath
	}
	domainGroup.Level = int(sqlcGroup.GroupLevel)

	// Audit fields
	domainGroup.CreatedAt = sqlcGroup.CreatedAt
	domainGroup.UpdatedAt = sqlcGroup.UpdatedAt
	if sqlcGroup.CreatedBy != nil {
		domainGroup.CreatedBy = *sqlcGroup.CreatedBy
	}
	if sqlcGroup.UpdatedBy != nil {
		domainGroup.UpdatedBy = *sqlcGroup.UpdatedBy
	}

	return nil
}

// Filter Mapping

func (r *accountsRepository) mapFilterToListParams(filter *domain.AccountGroupFilter) db.ListAccountGroupsParams {
	params := db.ListAccountGroupsParams{
		EntityID:      filter.EntityID,
		RootType:      nil,
		GroupCategory: filter.GroupType,
		ParentGroupID: filter.ParentGroupID,
		IsActive:      filter.IsActive,
		SortBy:        filter.SortBy,
		LimitCount:    int32(getIntValue(filter.Limit, 50)), // Default limit
		OffsetCount:   int32(getIntValue(filter.Offset, 0)), // Default offset
	}

	return params
}

func (r *accountsRepository) mapFilterToCountParams(filter *domain.AccountGroupFilter) db.CountAccountGroupsParams {
	return db.CountAccountGroupsParams{
		EntityID:      filter.EntityID,
		RootType:      nil,
		GroupCategory: filter.GroupType,
		ParentGroupID: filter.ParentGroupID,
		IsActive:      filter.IsActive,
	}
}

// Cache Helper Functions

func (r *accountsRepository) isSimpleListFilter(filter *domain.AccountGroupFilter) bool {
	if filter == nil {
		return true
	}

	// Consider it simple if it only has basic filters (no complex joins or searches)
	complexFields := 0
	if filter.FinancialStatementSection != nil {
		complexFields++
	}
	if filter.CashFlowCategory != nil {
		complexFields++
	}
	// Note: AccountGroupFilter doesn't have SearchQuery field

	return complexFields <= 1
}

func (r *accountsRepository) generateFilterHash(filter *domain.AccountGroupFilter) string {
	// Create a simple hash of the filter parameters for caching
	if filter == nil {
		return "default"
	}

	// Serialize filter to JSON and hash it
	filterBytes, err := json.Marshal(filter)
	if err != nil {
		return "error"
	}

	hash := md5.Sum(filterBytes)
	return fmt.Sprintf("%x", hash)
}

// Error Mapping

func (r *accountsRepository) mapAccountGroupDatabaseError(err error, operation string) error {
	// Map database errors to domain errors for account groups
	if err == nil {
		return nil
	}

	// Check for PostgreSQL errors
	if pgErr, ok := err.(*pgconn.PgError); ok {
		switch pgErr.Code {
		case "23505": // Unique violation
			if strings.Contains(pgErr.ConstraintName, "group_code") {
				return errors.NewBusinessError("ACCOUNT_GROUP_CODE_EXISTS", "Account group code already exists")
			}
			return errors.NewBusinessError("ACCOUNT_GROUP_DUPLICATE", "Account group already exists")
		case "23503": // Foreign key violation
			return errors.NewBusinessError("INVALID_PARENT_GROUP", "Invalid parent group reference")
		case "23514": // Check constraint violation
			return errors.NewBusinessError("INVALID_ACCOUNT_GROUP_DATA", "Account group data violates business rules")
		}
	}

	// Return the original error wrapped
	return fmt.Errorf("account group %s failed: %w", operation, err)
}

// Helper Functions

func getInt32Value(val *int32) int32 {
	if val == nil {
		return 0
	}
	return *val
}

func getBoolValue(val *bool) bool {
	if val == nil {
		return false
	}
	return *val
}

func getIntValue(val *int, defaultVal int) int {
	if val == nil {
		return defaultVal
	}
	return *val
}

// Additional helper functions can be added here if needed
