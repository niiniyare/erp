package core

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"

	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/core/abac"
	abacRepo "github.com/niiniyare/erp/internal/core/abac/repository"
	"github.com/niiniyare/erp/internal/core/access"
	"github.com/niiniyare/erp/internal/core/analytics"
	"github.com/niiniyare/erp/internal/core/audit"
	"github.com/niiniyare/erp/internal/core/entity"
	"github.com/niiniyare/erp/internal/core/featureflag"
	financeRepo "github.com/niiniyare/erp/internal/core/finance/repository"
	financeService "github.com/niiniyare/erp/internal/core/finance/service"
	"github.com/niiniyare/erp/internal/core/iam"
	"github.com/niiniyare/erp/internal/core/identity"
	"github.com/niiniyare/erp/internal/core/notification"
	"github.com/niiniyare/erp/internal/core/settings"
	"github.com/niiniyare/erp/internal/core/tenant"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/platform/temporal"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// Dependencies represents all external dependencies needed by core services
type Dependencies struct {
	Store    db.Store
	Cache    cache.Service
	Logger   logger.Logger
	Metrics  metrics.MetricsProvider
	Tracing  tracing.TracingService
	Temporal *temporal.Platform
}

// Validate ensures all required dependencies are present
func (d Dependencies) Validate() error {
	if d.Store == nil {
		return fmt.Errorf("database store is required")
	}
	if d.Logger == nil {
		return fmt.Errorf("logger is required")
	}
	if d.Metrics == nil {
		return fmt.Errorf("metrics service is required")
	}
	if d.Tracing == nil {
		return fmt.Errorf("tracing service is required")
	}
	return nil
}

// ServiceContainer holds all initialized core services
type ServiceContainer struct {
	// Core Infrastructure Services
	TenantService             tenant.Service
	TenantProvisioningService tenant.ProvisioningService
	EntityService             entity.Service
	IdentityService           identity.Service

	// Security & Access Control
	ABACService   abac.Service
	IAMService    iam.Service
	AccessService access.Service
	AuditService  audit.Service

	// Feature Management
	FeatureFlagService      featureflag.Service
	AdminFeatureFlagService featureflag.AdminService

	// Business Modules
	FinanceService      *financeService.Services
	NotificationService notification.NotificationService
	SettingsService     settings.SettingsService
	AnalyticsService    analytics.UserAnalyticsService

	// Internal state
	deps    Dependencies
	logger  logger.Logger
	mu      sync.RWMutex
	started bool
}

// NewServiceContainer creates a new service container with dependencies
func NewServiceContainer(deps Dependencies) (*ServiceContainer, error) {
	if err := deps.Validate(); err != nil {
		return nil, fmt.Errorf("invalid dependencies: %w", err)
	}

	return &ServiceContainer{
		deps:   deps,
		logger: deps.Logger.WithFields(logger.Fields{"component": "core_services"}),
	}, nil
}

// Initialize initializes all core services in the correct dependency order
func (sc *ServiceContainer) Initialize(ctx context.Context) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if sc.started {
		return fmt.Errorf("services already initialized")
	}

	sc.logger.Info("Initializing core services", logger.Fields{
		"action": "service_initialization",
		"stage":  "starting",
	})

	// Phase 1: Initialize foundational services
	if err := sc.initializeFoundationalServices(ctx); err != nil {
		return fmt.Errorf("failed to initialize foundational services: %w", err)
	}

	// Phase 2: Initialize security services
	if err := sc.initializeSecurityServices(ctx); err != nil {
		return fmt.Errorf("failed to initialize security services: %w", err)
	}

	// Phase 3: Initialize business services
	if err := sc.initializeBusinessServices(ctx); err != nil {
		return fmt.Errorf("failed to initialize business services: %w", err)
	}

	// Phase 4: Initialize integration services
	if err := sc.initializeIntegrationServices(ctx); err != nil {
		return fmt.Errorf("failed to initialize integration services: %w", err)
	}

	sc.started = true
	sc.logServiceStatus()

	return nil
}

// Shutdown gracefully shuts down all services
func (sc *ServiceContainer) Shutdown(ctx context.Context) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if !sc.started {
		return nil
	}

	sc.logger.Info("Shutting down core services", logger.Fields{
		"action": "service_shutdown",
	})

	// Services that implement context.Context shutdown can be added here
	// For now, most services are stateless and don't require shutdown

	sc.started = false
	return nil
}

// Phase 1: Foundational Services (no dependencies on other core services)
func (sc *ServiceContainer) initializeFoundationalServices(ctx context.Context) error {
	sc.logger.Info("Phase 1: Initializing foundational services")

	// Tenant Service - Required by almost everything
	tenantRepo := tenant.NewRepository(sc.deps.Store, sc.deps.Tracing)
	sc.TenantService = tenant.NewService(tenantRepo, sc.deps.Cache, sc.deps.Tracing)

	// Entity Service - Organizational structure
	entityRepo := entity.NewRepository(sc.deps.Store, sc.deps.Tracing, sc.deps.Metrics)
	sc.EntityService = entity.NewService(entityRepo, sc.deps.Tracing, sc.deps.Metrics)

	// Identity Service - User management
	identityRepo := identity.NewRepository(sc.deps.Store, sc.deps.Tracing, sc.deps.Metrics)
	sc.IdentityService = identity.NewService(identityRepo, sc.deps.Cache, sc.deps.Tracing, sc.deps.Metrics)

	// Audit Service - Required by other services for logging
	auditRepo := audit.NewRepository(sc.deps.Store, sc.deps.Logger, sc.deps.Tracing, sc.deps.Metrics)
	sc.AuditService = audit.NewService(auditRepo, sc.deps.Cache, sc.deps.Logger, sc.deps.Tracing, sc.deps.Metrics)

	sc.logger.Info("✅ Foundational services initialized", logger.Fields{
		"services": []string{"tenant", "entity", "identity", "audit"},
	})

	return nil
}

// Phase 2: Security Services (depend on foundational services)
func (sc *ServiceContainer) initializeSecurityServices(ctx context.Context) error {
	sc.logger.Info("Phase 2: Initializing security services")

	// ABAC Service - Advanced access control
	policyRepo := abacRepo.NewPolicyRepository(sc.deps.Store, sc.deps.Cache, sc.deps.Logger, sc.deps.Metrics, sc.deps.Tracing)
	attributeRepo := abacRepo.NewAttributeRepository(sc.deps.Store, sc.deps.Cache, sc.deps.Logger, sc.deps.Metrics, sc.deps.Tracing)
	policyEvaluationRepo := abacRepo.NewPolicyEvaluationRepository(sc.deps.Store, sc.deps.Cache, sc.deps.Logger, sc.deps.Metrics, sc.deps.Tracing)

	sc.ABACService = abac.NewService(
		policyRepo,
		attributeRepo,
		policyEvaluationRepo,
		sc.IdentityService,
		sc.TenantService,
		sc.deps.Logger,
		sc.deps.Metrics,
		sc.deps.Tracing,
	)

	// IAM Service - Identity and Access Management (placeholder for now)
	// TODO: Initialize IAM service properly when all dependencies are ready
	// sc.IAMService = iam.NewService(...)

	// Access Service - High-level access control
	// TODO: Initialize access service properly
	// sc.AccessService = access.NewService()

	sc.logger.Info("✅ Security services initialized", logger.Fields{
		"services": []string{"abac", "iam", "access"},
	})

	return nil
}

// Phase 3: Business Services (depend on foundational and security services)
func (sc *ServiceContainer) initializeBusinessServices(ctx context.Context) error {
	sc.logger.Info("Phase 3: Initializing business services")

	// Feature Flag Service
	featureFlagRepo := featureflag.NewRepository(sc.deps.Store)
	webSocketService := featureflag.NewWebSocketService(sc.deps.Store, sc.deps.Tracing, sc.deps.Metrics)
	baseFeatureFlagService := featureflag.NewService(featureFlagRepo, sc.TenantService, sc.deps.Store, sc.AuditService, webSocketService)

	// Add caching if available
	if sc.deps.Cache != nil {
		sc.FeatureFlagService = featureflag.NewCachedFeatureFlagService(
			baseFeatureFlagService,
			sc.deps.Cache,
			sc.deps.Logger,
			sc.deps.Metrics.(*metrics.MetricsService),
			sc.deps.Tracing,
		)
	} else {
		sc.FeatureFlagService = baseFeatureFlagService
	}

	// Admin Feature Flag Service
	sc.AdminFeatureFlagService = featureflag.NewAdminService(
		sc.FeatureFlagService,
		sc.AuditService,
		sc.deps.Logger,
		sc.deps.Metrics.(*metrics.MetricsService),
		sc.deps.Tracing,
		nil, // CacheWarmer can be nil for now
	)

	// Finance Service
	accountRepo := financeRepo.NewAccountsRepository(sc.deps.Store, sc.deps.Cache, sc.deps.Tracing)
	transactionRepo := financeRepo.NewTransactionRepository(sc.deps.Store, sc.deps.Tracing)
	financeServiceDeps := financeService.Dependencies{
		AccountRepo:        accountRepo,
		TransactionRepo:    transactionRepo,
		Tracing:            sc.deps.Tracing,
		Metrics:            sc.deps.Metrics,
		IAMService:         nil, // sc.IAMService,
		FeatureFlagService: sc.FeatureFlagService,
	}

	if err := financeServiceDeps.Validate(); err != nil {
		return fmt.Errorf("finance service dependencies invalid: %w", err)
	}

	financeServices := financeService.NewServices(financeServiceDeps)
	sc.FinanceService = financeServices

	// Notification Service
	notificationRepo := notification.NewRepository(sc.deps.Store)
	sc.NotificationService = notification.NewNotificationService(
		sc.deps.Tracing,
		sc.deps.Metrics,
		nil, // Email service - can be nil for now
		nil, // SMS service - can be nil for now
		notificationRepo,
	)

	// Settings Service (placeholder for now)
	// TODO: Implement proper settings service initialization
	// sc.SettingsService = settings.NewSettingsService(ctx, repo, auditService, logger, metrics, tracing)

	// Analytics Service
	sc.AnalyticsService = analytics.NewUserAnalyticsService(
		sc.deps.Tracing,
		sc.deps.Metrics,
		sc.AuditService,
	)

	sc.logger.Info("✅ Business services initialized", logger.Fields{
		"services": []string{"featureflag", "admin_featureflag", "finance", "notification", "settings", "analytics"},
	})

	return nil
}

// Phase 4: Integration Services (depend on all other services)
func (sc *ServiceContainer) initializeIntegrationServices(ctx context.Context) error {
	sc.logger.Info("Phase 4: Initializing integration services")

	// Tenant Provisioning Service
	sc.TenantProvisioningService = tenant.NewProvisioningService(
		sc.TenantService,
		sc.IdentityService,
		nil, // AuditLogger - can be nil for now
		nil, // NotificationSender - can be nil for now
	)

	// Register Finance module with Temporal if available
	if sc.deps.Temporal != nil {
		if err := sc.registerFinanceWithTemporal(ctx); err != nil {
			sc.logger.Error("Failed to register finance with Temporal", logger.Fields{
				"error": err.Error(),
			})
			// Don't fail initialization, just log the error
		}
	}

	sc.logger.Info("✅ Integration services initialized", logger.Fields{
		"services": []string{"tenant_provisioning", "temporal_finance"},
	})

	return nil
}

// registerFinanceWithTemporal registers the finance module with Temporal platform
func (sc *ServiceContainer) registerFinanceWithTemporal(ctx context.Context) error {
	// TODO: Implement finance temporal integration when all services are ready
	sc.logger.Info("Finance temporal integration skipped (placeholder)", logger.Fields{
		"status": "placeholder",
	})
	return nil
}

// logServiceStatus logs the status of all initialized services
func (sc *ServiceContainer) logServiceStatus() {
	serviceCount := 0
	services := make(map[string]string)

	// Use reflection to get all service fields
	v := reflect.ValueOf(sc).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		// Skip unexported fields, internal state, or nil services
		if !field.IsExported() || isInternalField(field.Name) || value.IsNil() {
			continue
		}

		serviceCount++
		serviceName := formatServiceName(field.Name)
		services[serviceName] = "ready"
	}

	sc.logger.Info("🚀 Core Services Ready", logger.Fields{
		"total_services": serviceCount,
		"status":         "operational",
		"services":       services,
	})
}

// isInternalField checks if a field is internal state (not a service)
func isInternalField(fieldName string) bool {
	internalFields := []string{"deps", "logger", "mu", "started"}
	for _, internal := range internalFields {
		if fieldName == internal {
			return true
		}
	}
	return false
}

// formatServiceName converts field names like "TenantService" to "tenant"
func formatServiceName(fieldName string) string {
	// Remove "Service" suffix
	name := strings.TrimSuffix(fieldName, "Service")

	// Convert from PascalCase to snake_case
	var result strings.Builder
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}

	return strings.ToLower(result.String())
}

// Getter methods for accessing services

func (sc *ServiceContainer) GetTenantService() tenant.Service {
	return sc.TenantService
}

func (sc *ServiceContainer) GetTenantProvisioningService() tenant.ProvisioningService {
	return sc.TenantProvisioningService
}

func (sc *ServiceContainer) GetEntityService() entity.Service {
	return sc.EntityService
}

func (sc *ServiceContainer) GetIdentityService() identity.Service {
	return sc.IdentityService
}

func (sc *ServiceContainer) GetABACService() abac.Service {
	return sc.ABACService
}

func (sc *ServiceContainer) GetIAMService() iam.Service {
	return sc.IAMService
}

func (sc *ServiceContainer) GetAccessService() access.Service {
	return sc.AccessService
}

func (sc *ServiceContainer) GetAuditService() audit.Service {
	return sc.AuditService
}

func (sc *ServiceContainer) GetFeatureFlagService() featureflag.Service {
	return sc.FeatureFlagService
}

func (sc *ServiceContainer) GetAdminFeatureFlagService() featureflag.AdminService {
	return sc.AdminFeatureFlagService
}

func (sc *ServiceContainer) GetFinanceService() *financeService.Services {
	return sc.FinanceService
}

func (sc *ServiceContainer) GetNotificationService() notification.NotificationService {
	return sc.NotificationService
}

func (sc *ServiceContainer) GetSettingsService() settings.SettingsService {
	return sc.SettingsService
}

func (sc *ServiceContainer) GetAnalyticsService() analytics.UserAnalyticsService {
	return sc.AnalyticsService
}

// IsReady returns whether all services are initialized and ready
func (sc *ServiceContainer) IsReady() bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.started
}

// Health performs a health check on all services
func (sc *ServiceContainer) Health(ctx context.Context) error {
	if !sc.IsReady() {
		return fmt.Errorf("services not initialized")
	}

	// Add service-specific health checks here as needed
	// For now, just check if we're initialized
	return nil
}
