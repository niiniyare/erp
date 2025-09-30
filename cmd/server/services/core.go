package services

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"

	"github.com/google/uuid"
	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/core/abac"
	"github.com/niiniyare/erp/internal/core/abac/repository"
	"github.com/niiniyare/erp/internal/core/access/approval"
	"github.com/niiniyare/erp/internal/core/access/conditional"
	"github.com/niiniyare/erp/internal/core/access/execution"
	"github.com/niiniyare/erp/internal/core/access/request"
	"github.com/niiniyare/erp/internal/core/analytics"
	"github.com/niiniyare/erp/internal/core/audit"
	"github.com/niiniyare/erp/internal/core/entity"
	"github.com/niiniyare/erp/internal/core/featureflag"
	"github.com/niiniyare/erp/internal/core/identity"
	"github.com/niiniyare/erp/internal/core/notification"
	"github.com/niiniyare/erp/internal/core/tenant"
	"github.com/niiniyare/erp/internal/platform/cache"
	loggerPkg "github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// Simple bridge structures to eliminate adapters
type userServiceBridge struct {
	identityService identity.Service
}

func (u *userServiceBridge) GetUserByID(ctx context.Context, id uuid.UUID) (*request.User, error) {
	identityUser, err := u.identityService.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &request.User{
		ID:       identityUser.ID,
		Username: identityUser.Username,
		Email:    identityUser.Email,
	}, nil
}

type auditServiceBridge struct {
	auditService audit.Service
}

func (a *auditServiceBridge) LogPermissionEvaluation(ctx context.Context, evaluation *conditional.PermissionEvaluationAudit) error {
	contextData, _ := json.Marshal(map[string]any{
		"resource_name":      evaluation.ResourceName,
		"action_name":        evaluation.ActionName,
		"policy_decisions":   evaluation.PolicyDecisions,
		"evaluation_time_ms": evaluation.EvaluationTimeMS,
		"context":            evaluation.Context,
		"risk_factors":       evaluation.RiskFactors,
	})

	auditEvent := audit.CreateAuditEventRequest{
		UserID:        &evaluation.UserID,
		EventType:     "permission_evaluation",
		EventCategory: "access_control",
		Severity:      "info",
		Decision:      &evaluation.Decision,
		Reason:        stringPtr("Permission evaluation completed"),
		Context:       contextData,
	}

	_, err := a.auditService.CreateAuditEvent(ctx, auditEvent)
	return err
}

func stringPtr(s string) *string {
	return &s
}

type CoreServices struct {
	// Core services  
	TenantService             tenant.Service
	TenantProvisioningService tenant.ProvisioningService
	EntityService             entity.Service
	IdentityService           identity.Service
	ABACService               abac.Service
	AuditService              audit.Service
	FeatureFlagService        featureflag.Service
	AdminFeatureFlagService   featureflag.AdminService

	// Access services
	AccessRequestService     request.AccessRequestService
	ConditionalAccessService conditional.ConditionalAccessService
	AnalyticsService         analytics.UserAnalyticsService
	NotificationService      notification.NotificationService
	ApproverService          approval.ApproverService
	ExecutionService         execution.AccessExecutionService
}

func InitializeCoreServices(store db.Store, redisClient cache.Service, logger loggerPkg.Logger, metricsService *metrics.MetricsService, tracingService tracing.TracingService) (*CoreServices, error) {
	// Initialize repositories
	tenantRepo := tenant.NewRepository(store, tracingService)
	entityRepo := entity.NewRepository(store, tracingService, metricsService)
	identityRepo := identity.NewRepository(store, tracingService, metricsService)
	auditRepo := audit.NewRepository(store, logger, tracingService, metricsService)
	notificationRepo := notification.NewRepository(store)

	// Initialize ABAC repositories
	policyRepo := repository.NewPolicyRepository(store, redisClient, logger.WithFields(loggerPkg.Fields{}), metricsService, tracingService)
	attributeRepo := repository.NewAttributeRepository(store, redisClient, logger.WithFields(loggerPkg.Fields{}), metricsService, tracingService)
	policyEvaluationRepo := repository.NewPolicyEvaluationRepository(store, redisClient, logger.WithFields(loggerPkg.Fields{}), metricsService, tracingService)

	// Initialize core business services
	tenantService := tenant.NewService(tenantRepo, redisClient, tracingService)
	entityService := entity.NewService(entityRepo, tracingService, metricsService)
	identityService := identity.NewService(identityRepo, redisClient, tracingService, metricsService)

	// Initialize ABAC service
	abacService := abac.NewService(
		policyRepo,
		attributeRepo,
		policyEvaluationRepo,
		identityService,
		tenantService,
		logger.WithFields(loggerPkg.Fields{}),
		metricsService,
		tracingService,
	)

	// Initialize domain services first
	auditService := audit.NewService(auditRepo, redisClient, logger, tracingService, metricsService)

	// Initialize feature flag service with audit logging
	featureFlagRepo := featureflag.NewRepository(store)
	webSocketService := featureflag.NewWebSocketService(store, tracingService, metricsService)
	baseFeatureFlagService := featureflag.NewService(featureFlagRepo, tenantService, store, auditService, webSocketService)

	// Wrap with caching if Redis is available
	var featureFlagService featureflag.Service
	if redisClient != nil {
		featureFlagService = featureflag.NewCachedFeatureFlagService(
			baseFeatureFlagService,
			redisClient,
			logger.WithFields(loggerPkg.Fields{"service": "featureflag_cached"}),
			metricsService,
			tracingService,
		)
		logger.Info("Feature flag service initialized with Redis caching", loggerPkg.Fields{
			"service": "featureflag",
			"caching": "enabled",
		})
	} else {
		featureFlagService = baseFeatureFlagService
		logger.Info("Feature flag service initialized without caching", loggerPkg.Fields{
			"service": "featureflag",
			"caching": "disabled",
		})
	}

	// Initialize admin feature flag service
	adminFeatureFlagService := featureflag.NewAdminService(
		featureFlagService,
		auditService,
		logger.WithFields(loggerPkg.Fields{"service": "admin_featureflag"}),
		metricsService,
		tracingService,
		nil, // CacheWarmer can be nil for now
	)
	logger.Info("Admin feature flag service initialized", loggerPkg.Fields{
		"service": "admin_featureflag",
		"status":  "ready",
	})

	// Create bridge instances
	userBridge := &userServiceBridge{identityService: identityService}
	auditBridge := &auditServiceBridge{auditService: auditService}

	approverService := approval.NewApproverService(identityService, tracingService, metricsService)
	notificationService := notification.NewNotificationService(tracingService, metricsService, nil, nil, notificationRepo)
	executionService := execution.NewAccessExecutionService(identityRepo, identityService, tracingService, metricsService)

	accessRequestService := request.NewAccessRequestService(
		nil, // TODO: Implement AccessRequestRepository
		redisClient,
		tracingService,
		metricsService,
		userBridge, // Use bridge instead of direct service
		notificationService,
		approverService,
		executionService,
		auditService,
	)
	conditionalAccessService := conditional.NewConditionalAccessService(tracingService, metricsService, auditBridge) // Use bridge
	analyticsService := analytics.NewUserAnalyticsService(tracingService, metricsService, auditService)

	// Initialize tenant provisioning service (after other services are created)
	tenantProvisioningService := tenant.NewProvisioningService(
		tenantService,
		identityService,
		nil, // AuditLogger - can be nil for now
		nil, // NotificationSender - can be nil for now
	)

	// Create services struct
	services := &CoreServices{
		TenantService:             tenantService,
		TenantProvisioningService: tenantProvisioningService,
		EntityService:             entityService,
		IdentityService:           identityService,
		ABACService:               abacService,
		AuditService:              auditService,
		FeatureFlagService:        featureFlagService,
		AdminFeatureFlagService:   adminFeatureFlagService,
		AccessRequestService:      accessRequestService,
		ConditionalAccessService:  conditionalAccessService,
		AnalyticsService:          analyticsService,
		NotificationService:       notificationService,
		ApproverService:           approverService,
		ExecutionService:          executionService,
	}

	// Log all initialized services dynamically
	logInitializedServices(services, logger)

	return services, nil
}

// logInitializedServices dynamically logs all services using reflection
func logInitializedServices(services *CoreServices, logger loggerPkg.Logger) {
	serviceCount := 0
	coreServices := make(map[string]any)
	accessServices := make(map[string]any)
	supportServices := make(map[string]any)

	// Use reflection to get all service fields
	v := reflect.ValueOf(services).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		// Skip unexported fields or nil services
		if !field.IsExported() || value.IsNil() {
			continue
		}

		serviceCount++
		serviceName := formatServiceName(field.Name)
		status := "ready"

		// Categorize services based on name patterns
		fieldName := strings.ToLower(field.Name)
		switch {
		case strings.Contains(fieldName, "access") || strings.Contains(fieldName, "approval") || strings.Contains(fieldName, "execution"):
			accessServices[serviceName] = status
		case strings.Contains(fieldName, "analytics") || strings.Contains(fieldName, "notification"):
			supportServices[serviceName] = status
		default:
			coreServices[serviceName] = status
		}
	}

	// Log summary
	logger.Info("🚀 Core Services Initialized", loggerPkg.Fields{
		"total_services": serviceCount,
		"status":         "ready",
	})

	// Log core services
	if len(coreServices) > 0 {
		logger.Info("  → Core Services", coreServices)
	}

	// Log access services
	if len(accessServices) > 0 {
		logger.Info("  → Access Management", accessServices)
	}

	// Log support services
	if len(supportServices) > 0 {
		logger.Info("  → Support Services", supportServices)
	}
}

// formatServiceName converts field names like "TenantService" to "tenant"
func formatServiceName(fieldName string) string {
	// Remove "Service" suffix
	name := strings.TrimSuffix(fieldName, "Service")

	// Convert from PascalCase to lowercase
	var result strings.Builder
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}

	return strings.ToLower(result.String())
}
