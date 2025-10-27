// Package wire contains Google Wire provider sets for dependency injection
// in the Awo ERP system. This follows Clean Architecture boundaries and
// maintains multi-tenant isolation.
package wire

import (
	"github.com/google/wire"
	
	// Platform layer
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/platform/config"
	"github.com/niiniyare/erp/internal/platform/temporal"
	
	// Database layer
	"github.com/niiniyare/erp/db/sqlc"
	
	// Shared infrastructure
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	
	// Core services
	"github.com/niiniyare/erp/internal/core/audit"
	"github.com/niiniyare/erp/internal/core/featureflag"
	"github.com/niiniyare/erp/internal/core/finance/service"
	"github.com/niiniyare/erp/internal/core/iam"
	"github.com/niiniyare/erp/internal/core/settings"
	"github.com/niiniyare/erp/internal/core/tenant"
	
	// API layer
	"github.com/niiniyare/erp/internal/api/handlers"
	"github.com/niiniyare/erp/internal/api/middleware"
)

// ============================================================================
// PLATFORM PROVIDER SET - Foundation Layer
// ============================================================================

// PlatformProviderSet provides all platform-level dependencies
var PlatformProviderSet = wire.NewSet(
	// Configuration
	config.Load,
	
	// Database connection and store
	NewDatabaseConnection,
	NewDBStore,
	
	// Cache service
	NewCacheService,
	
	// Observability stack
	NewLogger,
	NewMetricsProvider,
	NewTracingService,
	
	// Temporal client
	NewTemporalClient,
)

// ============================================================================
// REPOSITORY PROVIDER SET - Data Access Layer
// ============================================================================

// RepositoryProviderSet provides all repository implementations
var RepositoryProviderSet = wire.NewSet(
	// Tenant repositories
	NewTenantRepository,
	
	// IAM repositories
	NewUserRepository,
	NewPersonRepository,
	NewEmployeeRepository,
	NewRoleRepository,
	NewPermissionRepository,
	
	// Finance repositories
	NewAccountRepository,
	NewAccountGroupRepository,
	NewTransactionRepository,
	
	// Feature flag repositories
	NewFeatureFlagRepository,
	
	// Audit repositories
	NewAuditRepository,
	
	// Settings repositories
	NewSettingsRepository,
)

// ============================================================================
// CORE SERVICE PROVIDER SET - Business Logic Layer
// ============================================================================

// CoreServiceProviderSet provides core business services
// Note: Order matters here to avoid circular dependencies
var CoreServiceProviderSet = wire.NewSet(
	// Tenant service (no business service dependencies)
	NewTenantService,
	
	// Settings service (depends on tenant)
	NewSettingsService,
	
	// Audit service (depends on tenant)
	NewAuditService,
	
	// Feature flag service (depends on tenant, audit)
	NewFeatureFlagService,
	
	// IAM services (depends on tenant, settings, feature flags)
	NewAuthenticationService,
	NewAuthorizationService,
	NewPolicyService,
	NewIAMService,
	
	// Finance services (depends on IAM, feature flags)
	NewFinanceServiceDependencies,
	NewFinanceServices,
)

// ============================================================================
// API PROVIDER SET - Presentation Layer
// ============================================================================

// APIProviderSet provides API layer components
var APIProviderSet = wire.NewSet(
	// Middleware dependencies
	NewMiddlewareConfig,
	
	// Handler dependencies
	NewHandlerDependencies,
	NewRouter,
	
	// Fiber app
	NewFiberApp,
)

// ============================================================================
// APPLICATION PROVIDER SET - Complete Dependency Graph
// ============================================================================

// ApplicationProviderSet combines all provider sets for the complete application
var ApplicationProviderSet = wire.NewSet(
	PlatformProviderSet,
	RepositoryProviderSet,
	CoreServiceProviderSet,
	APIProviderSet,
)

// ============================================================================
// MULTI-TENANT SCOPED PROVIDER SET
// ============================================================================

// TenantScopedProviderSet provides tenant-scoped dependencies
// These are created per-tenant context and include tenant-aware components
var TenantScopedProviderSet = wire.NewSet(
	// Tenant-scoped database operations
	NewTenantScopedDBStore,
	
	// Tenant-scoped cache keys
	NewTenantScopedCache,
	
	// Tenant-aware repositories
	NewTenantAwareRepositories,
)