// Package wire contains Google Wire provider sets for dependency injection
// in the Awo ERP system. This follows Clean Architecture boundaries and
// maintains multi-tenant isolation.
package wire

import (
	"github.com/google/wire"
	"awo.so/internal/platform/config"
)

// ============================================================================
// PLATFORM PROVIDER SET - Foundation Layer
// ============================================================================

// PlatformProviderSet provides all platform-level dependencies
var PlatformProviderSet = wire.NewSet(
	// Configuration
	config.Load,

	// Database and cache
	NewDatabaseConnection,
	NewDBStore,
	NewCacheService,

	// Observability
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
var RepositoryProviderSet = wire.NewSet()

// ============================================================================
// CORE SERVICE PROVIDER SET - Business Logic Layer
// ============================================================================

// CoreServiceProviderSet provides core business services
var CoreServiceProviderSet = wire.NewSet(
	// Tenant service and Temporal integration
	NewTenantService,
	NewTenantTemporalIntegration,

	// Identity, authz, session
	NewIdentityRepository,
	NewIdentityService,
	NewAuthzService,
	NewSessionRepository,
	NewSessionService,

	// Audit
	NewAuditRepository,
	NewAuditService,

	// API keys
	NewAPIKeyRepository,
	NewAPIKeyService,

	// SSO (OAuth/OIDC)
	NewSSORepository,
	NewSSOService,

	// Finance services
	NewFinanceServices,

	// Contracts service
	NewContractService,
)

// ============================================================================
// API PROVIDER SET - Presentation Layer
// ============================================================================

// APIProviderSet provides API layer components
var APIProviderSet = wire.NewSet(
	// Fiber app
	NewFiberApp,

	// Middleware components
	NewTenantMiddlewareConfig,
	NewTenantMiddleware,
	NewRouteSecurityManager,

	// Handler dependencies
	NewHandlerDependencies,
	NewRouter,
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
