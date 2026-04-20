package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	db "awo.so/db/sqlc"
	"awo.so/internal/core/finance/domain"
	"awo.so/internal/shared"
	"awo.so/internal/shared/errors"
)

// Cache TTL constants
const (
	accountGroupCacheTTL      = 30 * time.Minute
	accountGroupListCacheTTL  = 15 * time.Minute
	accountGroupCountCacheTTL = 10 * time.Minute
)

/*
type a countsRepository struct {
	store   db.Store
	cache   cache.Service
	tracing tracing.Service
}

func NewaccountsRepository(store db.Store, cache cache.Service, tracing tracing.Service) domain.accountsRepository {
	return &accountsRepository{
		store:   store,
		cache:   cache,
		tracing: tracing,
	}
}
*/
// Cache key generators
func (r *accountsRepository) getCacheKey(prefix string, tenantID uuid.UUID, keys ...string) string {
	keyParts := []string{"account_group", prefix, tenantID.String()}
	keyParts = append(keyParts, keys...)
	return strings.Join(keyParts, ":")
}

func (r *accountsRepository) invalidateGroupCaches(ctx context.Context, tenantID uuid.UUID, groupID *uuid.UUID) {
	// Invalidate specific group cache if groupID provided
	if groupID != nil {
		r.cache.Delete(ctx, r.getCacheKey("id", tenantID, groupID.String()))
	}

	// Invalidate list and count caches
	r.cache.DeletePattern(ctx, r.getCacheKey("list", tenantID, "*"))
	r.cache.DeletePattern(ctx, r.getCacheKey("count", tenantID, "*"))
	r.cache.DeletePattern(ctx, r.getCacheKey("hierarchy", tenantID, "*"))
	r.cache.Delete(ctx, r.getCacheKey("all", tenantID))
}

// Basic CRUD Operations

func (r *accountsRepository) CreateAccountGroup(ctx context.Context, accountGroup *domain.AccountGroup) error {
	ctx, span := r.tracing.StartSpan(ctx, "accountsRepository.Create")
	defer span.End()

	// Get tenant ID from context
	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	// Use tenant-aware transaction for proper isolation
	return r.store.WithTenantFromCtx(ctx, func(ctx context.Context, s db.Store) error {
		// Validate the account group
		if err := accountGroup.Validate(); err != nil {
			return fmt.Errorf("%w", err)
		}

		// Map domain account group to SQLC parameters
		params, err := r.mapDomainAccountGroupToCreateParams(accountGroup)
		if err != nil {
			return fmt.Errorf("failed to map create account group request: %w", err)
		}

		// Execute SQLC query within tenant context
		sqlcAccountGroup, err := s.CreateAccountGroup(ctx, params)
		if err != nil {
			return r.mapAccountGroupDatabaseError(err, "create_account_group")
		}

		// Update the account group with generated fields
		err = r.mapSQLCAccountGroupToDomain(sqlcAccountGroup, accountGroup)
		if err != nil {
			return fmt.Errorf("failed to map created account group: %w", err)
		}

		// Invalidate caches after successful creation
		r.invalidateGroupCaches(ctx, tenantID, &accountGroup.ID)

		return nil
	})
}

func (r *accountsRepository) GetAccountGroupByID(ctx context.Context, id uuid.UUID) (*domain.AccountGroup, error) {
	ctx, span := r.tracing.StartSpan(ctx, "accountsRepository.GetByID")
	defer span.End()

	// Get tenant ID from context
	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Try cache first
	cacheKey := r.getCacheKey("id", tenantID, id.String())
	var cachedGroup domain.AccountGroup
	if err := r.cache.Get(ctx, cacheKey, &cachedGroup); err == nil {
		return &cachedGroup, nil
	}

	// Cache miss - get from database
	var accountGroup *domain.AccountGroup
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		sqlcAccountGroup, err := s.GetAccountGroup(ctx, db.GetAccountGroupParams{
			GroupID:  id,
			EntityID: nil, // Allow cross-entity access within tenant
		})
		if err != nil {
			if err == db.ErrNoRows {
				return errors.NewBusinessError("ACCOUNT_GROUP_NOT_FOUND", "Account group not found")
			}
			return r.mapAccountGroupDatabaseError(err, "get_account_group")
		}

		// Map SQLC result to domain
		accountGroup = &domain.AccountGroup{}
		err = r.mapSQLCAccountGroupToDomain(sqlcAccountGroup, accountGroup)
		if err != nil {
			return fmt.Errorf("failed to map account group: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Cache the result
	r.cache.Set(ctx, cacheKey, accountGroup, accountGroupCacheTTL)

	return accountGroup, nil
}

func (r *accountsRepository) GetAccountGroupByCode(ctx context.Context, code string, entityID *uuid.UUID) (*domain.AccountGroup, error) {
	ctx, span := r.tracing.StartSpan(ctx, "accountsRepository.GetByCode")
	defer span.End()

	// Get tenant ID from context
	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var accountGroup *domain.AccountGroup
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		sqlcAccountGroup, err := s.GetAccountGroupByCode(ctx, db.GetAccountGroupByCodeParams{
			GroupCode: code,
			EntityID:  entityID,
		})
		if err != nil {
			if err == db.ErrNoRows {
				return errors.NewBusinessError("ACCOUNT_GROUP_NOT_FOUND", "Account group not found")
			}
			return r.mapAccountGroupDatabaseError(err, "get_account_group_by_code")
		}

		// Map SQLC result to domain
		accountGroup = &domain.AccountGroup{}
		err = r.mapSQLCAccountGroupToDomain(sqlcAccountGroup, accountGroup)
		if err != nil {
			return fmt.Errorf("failed to map account group: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return accountGroup, nil
}

func (r *accountsRepository) UpdateAccountGroup(ctx context.Context, id uuid.UUID, accountGroup *domain.AccountGroup) error {
	ctx, span := r.tracing.StartSpan(ctx, "accountsRepository.Update")
	defer span.End()

	// Get tenant ID from context
	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Validate the account group
		if err := accountGroup.Validate(); err != nil {
			if len(err) > 0 {
				return domain.ValidationErrors(err)
			}
			return nil
		}

		// Map to update parameters
		params := r.mapDomainAccountGroupToUpdateParams(id, accountGroup)

		// Execute update
		sqlcAccountGroup, err := s.UpdateAccountGroup(ctx, params)
		if err != nil {
			if err == db.ErrNoRows {
				return errors.NewBusinessError("ACCOUNT_GROUP_NOT_FOUND", "Account group not found")
			}
			return r.mapAccountGroupDatabaseError(err, "update_account_group")
		}

		// Update the domain object with returned data
		err = r.mapSQLCAccountGroupToDomain(sqlcAccountGroup, accountGroup)
		if err != nil {
			return fmt.Errorf("failed to map updated account group: %w", err)
		}

		// Invalidate caches after successful update
		r.invalidateGroupCaches(ctx, tenantID, &id)

		return nil
	})
}

func (r *accountsRepository) DeleteAccountGroup(ctx context.Context, id uuid.UUID, entityID *uuid.UUID) error {
	ctx, span := r.tracing.StartSpan(ctx, "accountsRepository.Delete")
	defer span.End()

	// Get tenant ID from context
	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Check if group has children
		hasChildren, err := s.CheckGroupHasChildren(ctx, &id)
		if err != nil {
			return r.mapAccountGroupDatabaseError(err, "check_group_children")
		}

		if hasChildren {
			return errors.NewBusinessError("ACCOUNT_GROUP_HAS_CHILDREN", "Cannot delete account group with child groups")
		}

		userId, ok := shared.GetUserID(ctx)
		if !ok {
			return r.mapAccountGroupDatabaseError(fmt.Errorf("user ID not found"), "userId_required")
		}
		// Soft delete the account group
		err = s.SoftDeleteAccountGroup(ctx, db.SoftDeleteAccountGroupParams{
			DeletedAt: sql.NullTime{Time: time.Now(), Valid: true},
			DeletedBy: &userId,
			GroupID:   id,
			EntityID:  entityID,
		})
		if err != nil {
			return r.mapAccountGroupDatabaseError(err, "delete_account_group")
		}

		// Invalidate caches after successful deletion
		r.invalidateGroupCaches(ctx, tenantID, &id)

		return nil
	})
}

func (r *accountsRepository) ListAccountGroups(ctx context.Context, filter *domain.AccountGroupFilter) ([]*domain.AccountGroup, error) {
	ctx, span := r.tracing.StartSpan(ctx, "accountsRepository.List")
	defer span.End()

	// Get tenant ID from context
	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Try cache for simple list requests
	if r.isSimpleListFilter(filter) {
		cacheKey := r.getCacheKey("list", tenantID, r.generateFilterHash(filter))
		var cachedGroups []*domain.AccountGroup
		if err := r.cache.Get(ctx, cacheKey, &cachedGroups); err == nil {
			return cachedGroups, nil
		}
	}

	// Cache miss or complex filter - get from database
	var accountGroups []*domain.AccountGroup
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		params := r.mapFilterToListParams(filter)
		sqlcGroups, err := s.ListAccountGroups(ctx, params)
		if err != nil {
			return r.mapAccountGroupDatabaseError(err, "list_account_groups")
		}

		accountGroups = make([]*domain.AccountGroup, len(sqlcGroups))
		for i, sqlcGroup := range sqlcGroups {
			accountGroups[i] = &domain.AccountGroup{}
			err = r.mapSQLCAccountGroupToDomain(sqlcGroup, accountGroups[i])
			if err != nil {
				return fmt.Errorf("failed to map account group at index %d: %w", i, err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Cache simple list results
	if r.isSimpleListFilter(filter) {
		cacheKey := r.getCacheKey("list", tenantID, r.generateFilterHash(filter))
		r.cache.Set(ctx, cacheKey, accountGroups, accountGroupListCacheTTL)
	}

	return accountGroups, nil
}

func (r *accountsRepository) CountAccountGroups(ctx context.Context, filter *domain.AccountGroupFilter) (int64, error) {
	ctx, span := r.tracing.StartSpan(ctx, "accountsRepository.Count")
	defer span.End()

	// Get tenant ID from context
	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return 0, fmt.Errorf("tenant ID not found in context")
	}

	// Try cache for simple count requests
	if r.isSimpleListFilter(filter) {
		cacheKey := r.getCacheKey("count", tenantID, r.generateFilterHash(filter))
		var cachedCount int64
		if err := r.cache.Get(ctx, cacheKey, &cachedCount); err == nil {
			return cachedCount, nil
		}
	}

	var count int64
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		params := r.mapFilterToCountParams(filter)
		var err error
		count, err = s.CountAccountGroups(ctx, params)
		if err != nil {
			return r.mapAccountGroupDatabaseError(err, "count_account_groups")
		}
		return nil
	})
	if err != nil {
		return 0, err
	}

	// Cache simple count results
	if r.isSimpleListFilter(filter) {
		cacheKey := r.getCacheKey("count", tenantID, r.generateFilterHash(filter))
		r.cache.Set(ctx, cacheKey, count, accountGroupCountCacheTTL)
	}

	return count, nil
}

// Specialized query methods

func (r *accountsRepository) GetAccountGroupHierarchy(ctx context.Context, rootGroupID *uuid.UUID, entityID *uuid.UUID) ([]*domain.AccountGroup, error) {
	ctx, span := r.tracing.StartSpan(ctx, "accountsRepository.GetHierarchy")
	defer span.End()

	// Get tenant ID from context
	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var accountGroups []*domain.AccountGroup
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		sqlcGroups, err := s.GetAccountGroupHierarchy(ctx, db.GetAccountGroupHierarchyParams{
			EntityID:    entityID,
			RootGroupID: rootGroupID,
		})
		if err != nil {
			return r.mapAccountGroupDatabaseError(err, "get_account_group_hierarchy")
		}

		accountGroups = make([]*domain.AccountGroup, len(sqlcGroups))
		for i, sqlcGroup := range sqlcGroups {
			accountGroups[i] = &domain.AccountGroup{}
			err = r.mapSQLCAccountGroupToDomain(sqlcGroup, accountGroups[i])
			if err != nil {
				return fmt.Errorf("failed to map account group at index %d: %w", i, err)
			}
		}

		return nil
	})

	return accountGroups, err
}

func (r *accountsRepository) GetGroupsByFinancialStatement(ctx context.Context, statementType string, entityID *uuid.UUID) ([]*domain.AccountGroup, error) {
	ctx, span := r.tracing.StartSpan(ctx, "accountsRepository.GetByFinancialStatement")
	defer span.End()

	// Get tenant ID from context

	var accountGroups []*domain.AccountGroup
	err := r.store.WithTenantFromCtx(ctx, func(ctx context.Context, s db.Store) error {
		// Note: We need to update the generated SQLC query name
		sqlcGroups, err := s.GetGroupsByStatementSection(ctx, db.GetGroupsByStatementSectionParams{
			Column1:                   uuid.UUID{}, // This needs to be fixed in SQL query
			FinancialStatementSection: &statementType,
		})
		if err != nil {
			return r.mapAccountGroupDatabaseError(err, "get_groups_by_financial_statement")
		}

		accountGroups = make([]*domain.AccountGroup, len(sqlcGroups))
		for i, sqlcGroup := range sqlcGroups {
			accountGroups[i] = &domain.AccountGroup{}
			err = r.mapSQLCAccountGroupToDomain(sqlcGroup, accountGroups[i])
			if err != nil {
				return fmt.Errorf("failed to map account group at index %d: %w", i, err)
			}
		}

		return nil
	})

	return accountGroups, err
}

func (r *accountsRepository) GetGroupsByCashFlowCategory(ctx context.Context, category string, entityID *uuid.UUID) ([]*domain.AccountGroup, error) {
	ctx, span := r.tracing.StartSpan(ctx, "accountsRepository.GetByCashFlowCategory")
	defer span.End()

	// Get tenant ID from context
	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var accountGroups []*domain.AccountGroup
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Use the GetGroupsByCategory query for cash flow category
		sqlcGroups, err := s.GetGroupsByCategory(ctx, db.GetGroupsByCategoryParams{
			Column1:       uuid.UUID{}, // This needs to be fixed in SQL query
			GroupCategory: &category,
		})
		if err != nil {
			return r.mapAccountGroupDatabaseError(err, "get_groups_by_cash_flow_category")
		}

		accountGroups = make([]*domain.AccountGroup, len(sqlcGroups))
		for i, sqlcGroup := range sqlcGroups {
			accountGroups[i] = &domain.AccountGroup{}
			err = r.mapSQLCAccountGroupToDomain(sqlcGroup, accountGroups[i])
			if err != nil {
				return fmt.Errorf("failed to map account group at index %d: %w", i, err)
			}
		}

		return nil
	})

	return accountGroups, err
}

func (r *accountsRepository) ValidateAccountGroupCode(ctx context.Context, code string, excludeID *uuid.UUID, entityID *uuid.UUID) error {
	ctx, span := r.tracing.StartSpan(ctx, "accountsRepository.ValidateCode")
	defer span.End()

	// Get tenant ID from context
	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	var codeExists bool
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		var err error
		codeExists, err = s.ValidateAccountGroupCode(ctx, db.ValidateAccountGroupCodeParams{
			EntityID:  entityID,
			GroupCode: code,
			ExcludeID: excludeID,
		})
		if err != nil {
			return r.mapAccountGroupDatabaseError(err, "validate_account_group_code")
		}
		return nil
	})
	if err != nil {
		return err
	}

	if codeExists {
		return errors.NewBusinessError("ACCOUNT_GROUP_CODE_EXISTS", "Account group code already exists")
	}

	return nil
}
