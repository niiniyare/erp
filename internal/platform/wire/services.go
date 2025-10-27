// Package wire - Service layer providers
package wire

import (
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	
	// Core services
	"github.com/niiniyare/erp/internal/core/audit"
	"github.com/niiniyare/erp/internal/core/featureflag"
	"github.com/niiniyare/erp/internal/core/finance/service"
	"github.com/niiniyare/erp/internal/core/iam"
	"github.com/niiniyare/erp/internal/core/iam/authn"
	"github.com/niiniyare/erp/internal/core/iam/authz"
	"github.com/niiniyare/erp/internal/core/iam/policy"
	"github.com/niiniyare/erp/internal/core/iam/repo"
	"github.com/niiniyare/erp/internal/core/settings"
	settingsservice "github.com/niiniyare/erp/internal/core/settings/service"
	settingsrepo "github.com/niiniyare/erp/internal/core/settings/repository"
	"github.com/niiniyare/erp/internal/core/tenant"
	
	// Finance domain
	"github.com/niiniyare/erp/internal/core/finance/domain"
)

// ============================================================================
// TENANT SERVICE
// ============================================================================

// NewTenantService creates a new tenant service
func NewTenantService(
	repo tenant.Repository,
	cache cache.Service,
	tracer tracing.TracingService,
) tenant.Service {
	return tenant.NewService(repo, cache, tracer)
}

// ============================================================================
// SETTINGS SERVICE
// ============================================================================

// NewSettingsService creates a new settings service
func NewSettingsService(
	configRepo settingsrepo.ConfigurationRepository,
	cache cache.Service,
	tracer tracing.TracingService,
	metrics metrics.MetricsProvider,
	log logger.Logger,
) settings.SettingsService {
	return settingsservice.NewConfigurationService(
		configRepo,
		cache,
		tracer,
		metrics,
		log,
	)
}

// ============================================================================
// AUDIT SERVICE
// ============================================================================

// NewAuditService creates a new audit service
func NewAuditService(
	repo audit.Repository,
	cache cache.Service,
	log logger.Logger,
	tracer tracing.TracingService,
	metrics metrics.MetricsProvider,
) audit.Service {
	return audit.NewService(repo, cache, log, tracer, metrics)
}

// ============================================================================
// FEATURE FLAG SERVICE
// ============================================================================

// NewFeatureFlagService creates a new feature flag service
func NewFeatureFlagService(
	repo featureflag.Repository,
	tenantService tenant.Service,
	auditService audit.Service,
	cache cache.Service,
	log logger.Logger,
	tracer tracing.TracingService,
	metrics metrics.MetricsProvider,
) featureflag.Service {
	return featureflag.NewService(
		repo,
		tenantService,
		auditService,
		cache,
		log,
		tracer,
		metrics,
	)
}

// ============================================================================
// IAM SERVICES
// ============================================================================

// NewAuthenticationService creates a new authentication service
func NewAuthenticationService(
	userRepo repo.UserRepository,
	personRepo repo.PersonRepository,
	employeeRepo repo.EmployeeRepository,
	cache cache.Service,
	log logger.Logger,
	tracer tracing.TracingService,
	metrics metrics.MetricsProvider,
) authn.Service {
	return authn.NewService(
		userRepo,
		personRepo,
		employeeRepo,
		cache,
		log,
		tracer,
		metrics,
	)
}

// NewAuthorizationService creates a new authorization service
func NewAuthorizationService(
	permissionRepo repo.PermissionRepository,
	roleRepo repo.RoleRepository,
	userRepo repo.UserRepository,
	cache cache.Service,
	log logger.Logger,
	tracer tracing.TracingService,
	metrics metrics.MetricsProvider,
) authz.Service {
	return authz.NewService(
		permissionRepo,
		roleRepo,
		userRepo,
		cache,
		log,
		tracer,
		metrics,
	)
}

// NewPolicyService creates a new policy service
func NewPolicyService(
	permissionRepo repo.PermissionRepository,
	roleRepo repo.RoleRepository,
	cache cache.Service,
	log logger.Logger,
	tracer tracing.TracingService,
	metrics metrics.MetricsProvider,
) policy.Service {
	return policy.NewService(
		permissionRepo,
		roleRepo,
		cache,
		log,
		tracer,
		metrics,
	)
}

// NewIAMService creates a unified IAM service
func NewIAMService(
	authnService authn.Service,
	authzService authz.Service,
	policyService policy.Service,
	tenantService tenant.Service,
	auditService audit.Service,
	featureFlagService featureflag.Service,
	settingsService settings.SettingsService,
	cache cache.Service,
	log logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.TracingService,
) iam.Service {
	return iam.NewService(
		authnService,
		authzService,
		policyService,
		tenantService,
		auditService,
		featureFlagService,
		settingsService,
		cache,
		log,
		metrics,
		tracer,
	)
}

// ============================================================================
// FINANCE SERVICES
// ============================================================================

// NewFinanceServiceDependencies creates finance service dependencies struct
func NewFinanceServiceDependencies(
	accountRepo domain.AccountsRepository,
	accountGroupRepo domain.AccountGroupRepository,
	transactionRepo domain.TransactionRepository,
	tracer tracing.TracingService,
	metrics metrics.MetricsProvider,
	iamService iam.Service,
	featureFlagService featureflag.Service,
) service.Dependencies {
	return service.Dependencies{
		AccountRepo:        accountRepo,
		AccountGroupRepo:   accountGroupRepo,
		TransactionRepo:    transactionRepo,
		Tracing:           tracer,
		Metrics:           metrics,
		IAMService:        iamService,
		FeatureFlagService: featureFlagService,
	}
}

// NewFinanceServices creates a new finance services aggregate
func NewFinanceServices(deps service.Dependencies) *service.Services {
	return service.NewServices(deps)
}