// Package wire - Repository layer providers
package wire

import (
	"github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/shared/tracing"
	
	// Core domain repositories
	"github.com/niiniyare/erp/internal/core/audit"
	"github.com/niiniyare/erp/internal/core/featureflag"
	"github.com/niiniyare/erp/internal/core/finance/domain"
	"github.com/niiniyare/erp/internal/core/finance/repository"
	"github.com/niiniyare/erp/internal/core/iam/repo"
	"github.com/niiniyare/erp/internal/core/settings/repository"
	"github.com/niiniyare/erp/internal/core/tenant"
)

// ============================================================================
// TENANT REPOSITORIES
// ============================================================================

// NewTenantRepository creates a new tenant repository
func NewTenantRepository(
	store *sqlc.Store,
	cache cache.Service,
	tracer tracing.TracingService,
) tenant.Repository {
	return tenant.NewRepository(store, cache, tracer)
}

// ============================================================================
// IAM REPOSITORIES
// ============================================================================

// NewUserRepository creates a new user repository
func NewUserRepository(
	store *sqlc.Store,
	cache cache.Service,
	tracer tracing.TracingService,
) repo.UserRepository {
	return repo.NewUserRepository(store, cache, tracer)
}

// NewPersonRepository creates a new person repository
func NewPersonRepository(
	store *sqlc.Store,
	cache cache.Service,
	tracer tracing.TracingService,
) repo.PersonRepository {
	return repo.NewPersonRepository(store, cache, tracer)
}

// NewEmployeeRepository creates a new employee repository
func NewEmployeeRepository(
	store *sqlc.Store,
	cache cache.Service,
	tracer tracing.TracingService,
) repo.EmployeeRepository {
	return repo.NewEmployeeRepository(store, cache, tracer)
}

// NewRoleRepository creates a new role repository
func NewRoleRepository(
	store *sqlc.Store,
	cache cache.Service,
	tracer tracing.TracingService,
) repo.RoleRepository {
	return repo.NewRoleRepository(store, cache, tracer)
}

// NewPermissionRepository creates a new permission repository
func NewPermissionRepository(
	store *sqlc.Store,
	cache cache.Service,
	tracer tracing.TracingService,
) repo.PermissionRepository {
	return repo.NewPermissionRepository(store, cache, tracer)
}

// ============================================================================
// FINANCE REPOSITORIES
// ============================================================================

// NewAccountRepository creates a new account repository
func NewAccountRepository(
	store *sqlc.Store,
	cache cache.Service,
	tracer tracing.TracingService,
) domain.AccountsRepository {
	return repository.NewAccountRepository(store, cache, tracer)
}

// NewAccountGroupRepository creates a new account group repository
func NewAccountGroupRepository(
	store *sqlc.Store,
	cache cache.Service,
	tracer tracing.TracingService,
) domain.AccountGroupRepository {
	return repository.NewAccountGroupRepository(store, cache, tracer)
}

// NewTransactionRepository creates a new transaction repository
func NewTransactionRepository(
	store *sqlc.Store,
	cache cache.Service,
	tracer tracing.TracingService,
) domain.TransactionRepository {
	return repository.NewTransactionRepository(store, cache, tracer)
}

// ============================================================================
// FEATURE FLAG REPOSITORIES
// ============================================================================

// NewFeatureFlagRepository creates a new feature flag repository
func NewFeatureFlagRepository(
	store *sqlc.Store,
	cache cache.Service,
	tracer tracing.TracingService,
) featureflag.Repository {
	return featureflag.NewRepository(store, cache, tracer)
}

// ============================================================================
// AUDIT REPOSITORIES
// ============================================================================

// NewAuditRepository creates a new audit repository
func NewAuditRepository(
	store *sqlc.Store,
	cache cache.Service,
	tracer tracing.TracingService,
) audit.Repository {
	return audit.NewRepository(store, cache, tracer)
}

// ============================================================================
// SETTINGS REPOSITORIES
// ============================================================================

// NewSettingsRepository creates a new settings repository
func NewSettingsRepository(
	store *sqlc.Store,
	cache cache.Service,
	tracer tracing.TracingService,
) settingsrepo.ConfigurationRepository {
	return settingsrepo.NewConfigurationRepository(store, cache, tracer)
}