package iam

import (
	"context"

	"github.com/google/uuid"

	"github.com/niiniyare/erp/internal/core/audit"
	"github.com/niiniyare/erp/internal/core/featureflag"
	"github.com/niiniyare/erp/internal/core/iam/authn"
	"github.com/niiniyare/erp/internal/core/iam/authz"
	"github.com/niiniyare/erp/internal/core/iam/policy"
	"github.com/niiniyare/erp/internal/core/settings"
	"github.com/niiniyare/erp/internal/core/tenant"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// Service defines the unified IAM service interface that consolidates
// authentication, authorization, and policy management functionality
type Service interface {
	// Get individual service instances
	Authentication() authn.Service
	Authorization() authz.Service
	Policy() policy.Service
}

// service implements the unified IAM service
type service struct {
	// Domain services
	authnService  authn.Service
	authzService  authz.Service
	policyService policy.Service

	// External dependencies
	tenantService      tenant.Service
	auditService       audit.Service
	featureFlagService featureflag.Service
	cache              cache.Service
	setting            settings.SettingsService

	// Shared infrastructure
	logger  logger.Logger
	metrics metrics.MetricsProvider
	tracer  tracing.TracingService
}

// NewService creates a new unified IAM service instance
func NewService(
	authnService authn.Service,
	authzService authz.Service,
	policyService policy.Service,
	tenantService tenant.Service,
	auditService audit.Service,
	featureFlagService featureflag.Service,
	setting settings.SettingsService,
	cache cache.Service,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.TracingService,
) Service {
	return &service{
		authnService:       authnService,
		authzService:       authzService,
		policyService:      policyService,
		tenantService:      tenantService,
		auditService:       auditService,
		featureFlagService: featureFlagService,
		setting:            setting,
		cache:              cache,
		logger:             logger,
		metrics:            metrics,
		tracer:             tracer,
	}
}

// Helper methods for getting tenant context
func (s *service) getCurrentTenantID(ctx context.Context) uuid.UUID {
	tenant, err := s.tenantService.GetCurrentTenant(ctx)
	if err != nil {
		s.logger.WarnContext(ctx, "Failed to get current tenant", logger.Fields{"error": err})
		return uuid.Nil
	}
	return tenant.ID
}

func (s *service) getCurrentTenantSlug(ctx context.Context) string {
	tenant, err := s.tenantService.GetCurrentTenant(ctx)
	if err != nil {
		s.logger.WarnContext(ctx, "Failed to get current tenant", logger.Fields{"error": err})
		return ""
	}
	return tenant.Slug
}

// Service access methods

func (s *service) Authentication() authn.Service {
	return s.authnService
}

func (s *service) Authorization() authz.Service {
	return s.authzService
}

func (s *service) Policy() policy.Service {
	return s.policyService
}
