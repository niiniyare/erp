package tenant

import (
	"github.com/go-playground/validator/v10"
	coreTenant "awo.so/internal/core/tenant"
	"awo.so/internal/shared/encryption"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
)

// TenantHandler handles tenant-related HTTP requests.
// It delegates business logic to the tenant service and uses helpers for
// validation, error handling, and response formatting.
type TenantHandler struct {
	service    coreTenant.Service
	logger     logger.Logger
	metrics    metrics.MetricsProvider
	tracer     tracing.Service
	validator  *validator.Validate
	encryption encryption.EncryptionService
}

// NewTenantHandler creates a new tenant handler with dependencies.
func NewTenantHandler(
	service coreTenant.Service,
	log logger.Logger,
	met metrics.MetricsProvider,
	trc tracing.Service,
) *TenantHandler {
	v := validator.New()
	registerCustomValidators(v)

	return &TenantHandler{
		service:    service,
		logger:     log,
		metrics:    met,
		tracer:     trc,
		validator:  v,
		encryption: nil, // To be added later
	}
}
