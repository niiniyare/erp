package tenant

import (
	"awo.so/internal/core/entity"
	"awo.so/internal/core/iam/contract"
	coreTenant "awo.so/internal/core/tenant"
	"awo.so/internal/shared/encryption"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
	workflowsTenant "awo.so/internal/workflows/tenant"
	"github.com/go-playground/validator/v10"
)

// TenantHandler handles tenant-related HTTP requests.
// It delegates business logic to the tenant service and uses helpers for
// validation, error handling, and response formatting.
type TenantHandler struct {
	service        coreTenant.Service
	userSvc        contract.UserService  // nil = sync onboarding skips user creation
	authzSvc       contract.AuthzService // nil = sync onboarding skips IAM seeding
	entitySvc      entity.Service        // nil = sync onboarding skips root entity creation
	logger         logger.Logger
	metrics        metrics.MetricsProvider
	tracer         tracing.Service
	validator      *validator.Validate
	encryption     encryption.EncryptionService
	onboardStarter *workflowsTenant.Starter // nil = Temporal not wired; Onboard returns 503
}

// NewTenantHandler creates a new tenant handler with dependencies.
func NewTenantHandler(
	service coreTenant.Service,
	log logger.Logger,
	met metrics.MetricsProvider,
	trc tracing.Service,
	onboardStarter *workflowsTenant.Starter,
) *TenantHandler {
	v := validator.New()
	registerCustomValidators(v)

	return &TenantHandler{
		service:        service,
		logger:         log,
		metrics:        met,
		tracer:         trc,
		validator:      v,
		encryption:     nil,
		onboardStarter: onboardStarter,
	}
}

// WithOnboardServices injects the user, authz, and entity services needed for
// synchronous onboarding. Call this after NewTenantHandler when services are available.
func (h *TenantHandler) WithOnboardServices(userSvc contract.UserService, authzSvc contract.AuthzService, entitySvc entity.Service) *TenantHandler {
	h.userSvc = userSvc
	h.authzSvc = authzSvc
	h.entitySvc = entitySvc
	return h
}
