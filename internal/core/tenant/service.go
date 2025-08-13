package tenant

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"github.com/niiniyare/erp/internal/platform/cache"
	sharedErrors "github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// Service defines tenant business logic interface
// This is the main interface that other system services should use
type Service interface {
	// Tenant Provisioning
	ProvisionTenant(ctx context.Context, req ProvisionTenantRequest) (*ProvisionedTenantInfo, error)

	// Core CRUD operations
	CreateTenant(ctx context.Context, req CreateTenantRequest) (*Tenant, error)
	GetTenantByID(ctx context.Context, id uuid.UUID) (*Tenant, error) // Get specific tenant by ID
	GetCurrentTenant(ctx context.Context) (*Tenant, error)            // Get current tenant from database session context
	GetTenantBySubdomain(ctx context.Context, subdomain string) (*Tenant, error)
	UpdateTenant(ctx context.Context, id uuid.UUID, req UpdateTenantRequest) (*Tenant, error)
	DeactivateTenant(ctx context.Context, id uuid.UUID) error
	ActivateTenant(ctx context.Context, id uuid.UUID) error
	DeleteTenant(ctx context.Context, id uuid.UUID) error
	ListTenants(ctx context.Context, offset, limit int) ([]*Tenant, error)

	// Tenant resolution and validation (for other services)
	ResolveTenantID(ctx context.Context, subdomain string) (uuid.UUID, error)
	ValidateTenantAccess(ctx context.Context, tenantID uuid.UUID) error
	ExistsTenant(ctx context.Context, tenantID uuid.UUID) (bool, error)

	// Context operations (for middleware and other services)
	ValidateCurrentTenant(ctx context.Context) error
	WithTenantContext(ctx context.Context, tenantID uuid.UUID, fn func(context.Context) error) error

	// Cache management
	ClearTenantCache(ctx context.Context, tenantID uuid.UUID) error
	ClearSubdomainCache(ctx context.Context, subdomain string) error
	WarmupCache(ctx context.Context, tenantID uuid.UUID) error

	// Database session context management
	SetTenant(ctx context.Context, tenantID uuid.UUID) error
	ResetTenant(ctx context.Context) error
}

// service implements Service interface
type service struct {
	repo   Repository
	cache  cache.Service
	tracer tracing.TracingService
}

// NewService creates a new tenant service
func NewService(repo Repository, cache cache.Service, tracer tracing.TracingService) Service {
	return &service{
		repo:   repo,
		cache:  cache,
		tracer: tracer,
	}
}

// CreateTenant implements Service.CreateTenant
func (s *service) CreateTenant(ctx context.Context, req CreateTenantRequest) (*Tenant, error) {
	ctx, span := s.tracer.StartSpan(ctx, "tenant.service.CreateTenant")
	defer span.End()

	// Add tracing attributes
	span.SetAttributes(
		attribute.String("tenant.name", req.Name),
		attribute.String("tenant.email", req.Email),
	)
	if req.Subdomain != nil {
		span.SetAttributes(attribute.String("tenant.subdomain", *req.Subdomain))
	}

	logger.InfoContext(ctx, "Creating new tenant", logger.Fields{
		"tenant_name": req.Name,
		"subdomain":   req.Subdomain,
		"industry":    req.Industry,
		"operation":   "service.CreateTenant",
	})

	// Validate input
	if err := s.validateCreateTenantRequest(req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Invalid request")
		return nil, err
	}

	// Validate subdomain uniqueness if provided
	if req.Subdomain != nil && *req.Subdomain != "" {
		// Try to get tenant by subdomain - if found, subdomain exists
		_, err := s.repo.GetBySubdomain(ctx, *req.Subdomain)
		if err == nil {
			// Tenant found, subdomain exists
			logger.WarnContext(ctx, "Subdomain already exists", logger.Fields{
				"subdomain": *req.Subdomain,
			})
			return nil, sharedErrors.ErrSubdomainAlreadyExists
		}
		// If error is not "tenant not found", it's a real error
		if !sharedErrors.IsTenantNotFound(err) {
			logger.ErrorContext(ctx, "Failed to check subdomain existence", logger.Fields{
				"subdomain": *req.Subdomain,
				"error":     err.Error(),
			})
			return nil, fmt.Errorf("failed to check subdomain existence: %w", err)
		}
		// If tenant not found error, subdomain is available (good)
	}

	// Set default status if not provided
	status := req.Status
	if status == "" {
		status = StatusActive
	}

	// Create tenant entity
	tenant := &Tenant{
		ID:                 uuid.New(),
		Slug:               slug.Make(req.Name),
		Name:               req.Name,
		Email:              req.Email,
		Subdomain:          req.Subdomain,
		Status:             status,
		Timezone:           "UTC", // Default timezone
		CurrencyCode:       "USD", // Default currency
		Industry:           req.Industry,
		CompanySize:        req.CompanySize,
		TaxID:              req.TaxID,
		RegistrationNumber: req.RegistrationNumber,
		LegalEntityType:    req.LegalEntityType,
		Settings:           req.Settings,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	// Save to database
	if err := s.repo.Create(ctx, tenant); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to create tenant")

		logger.ErrorContext(ctx, "Failed to create tenant in database", logger.Fields{
			"tenant_id":   tenant.ID.String(),
			"tenant_name": tenant.Name,
			"error":       err.Error(),
			"operation":   "service.CreateTenant",
		})
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	span.SetAttributes(attribute.String("tenant.created_id", tenant.ID.String()))

	logger.InfoContext(ctx, "Tenant created successfully", logger.Fields{
		"tenant_id":   tenant.ID.String(),
		"tenant_name": tenant.Name,
		"subdomain":   tenant.Subdomain,
		"status":      string(tenant.Status),
		"operation":   "service.CreateTenant",
	})

	// Cache the tenant
	if err := s.cacheTenant(ctx, tenant); err != nil {
		logger.WarnContext(ctx, "Failed to cache new tenant", logger.Fields{
			"tenant_id": tenant.ID.String(),
			"error":     err.Error(),
			"operation": "service.CreateTenant",
		})
	}

	return tenant, nil
}

// GetTenantBySubdomain implements Service.GetTenantBySubdomain with caching
func (s *service) GetTenantBySubdomain(ctx context.Context, subdomain string) (*Tenant, error) {
	ctx, span := s.tracer.StartSpan(ctx, "tenant.service.GetTenantBySubdomain")
	defer span.End()

	span.SetAttributes(attribute.String("tenant.subdomain", subdomain))

	logger.DebugContext(ctx, "Getting tenant by subdomain", logger.Fields{
		"subdomain": subdomain,
		"operation": "service.GetTenantBySubdomain",
	})

	// Check cache first
	cacheKey := s.getSubdomainCacheKey(subdomain)
	var tenant Tenant
	if err := s.cache.Get(ctx, cacheKey, &tenant); err == nil {
		span.SetAttributes(attribute.Bool("cache.hit", true))
		logger.DebugContext(ctx, "Tenant found in cache", logger.Fields{
			"subdomain": subdomain,
			"tenant_id": tenant.ID.String(),
			"operation": "service.GetTenantBySubdomain",
		})
		return &tenant, nil
	}

	span.SetAttributes(attribute.Bool("cache.hit", false))

	// Get from database
	dbTenant, err := s.repo.GetBySubdomain(ctx, subdomain)
	if err != nil {
		span.RecordError(err)

		if sharedErrors.IsTenantNotFound(err) {
			span.SetStatus(codes.Error, "Tenant not found")
			logger.WarnContext(ctx, "Tenant not found by subdomain", logger.Fields{
				"subdomain": subdomain,
				"operation": "service.GetTenantBySubdomain",
			})
		} else {
			span.SetStatus(codes.Error, "Database error")
			logger.ErrorContext(ctx, "Failed to get tenant by subdomain", logger.Fields{
				"subdomain": subdomain,
				"error":     err.Error(),
				"operation": "service.GetTenantBySubdomain",
			})
		}
		return nil, err
	}

	span.SetAttributes(attribute.String("tenant.found_id", dbTenant.ID.String()))

	logger.DebugContext(ctx, "Tenant found in database", logger.Fields{
		"subdomain":   subdomain,
		"tenant_id":   dbTenant.ID.String(),
		"tenant_name": dbTenant.Name,
		"operation":   "service.GetTenantBySubdomain",
	})

	// Cache result
	if err := s.cacheTenant(ctx, dbTenant); err != nil {
		logger.WarnContext(ctx, "Failed to cache tenant", logger.Fields{
			"subdomain": subdomain,
			"tenant_id": dbTenant.ID.String(),
			"error":     err.Error(),
			"operation": "service.GetTenantBySubdomain",
		})
	}

	return dbTenant, nil
}

// GetTenant implements Service.GetTenant
// If no ID provided, gets the current tenant from database session context
func (s *service) GetTenantByID(ctx context.Context, id uuid.UUID) (*Tenant, error) {
	ctx, span := s.tracer.StartSpan(ctx, "tenant.service.GetTenant")
	defer span.End()

	var tenantID uuid.UUID
	var err error

	// Determine which tenant to get
	if len(id) > 0 && id != uuid.Nil {
		// Get specific tenant by ID
		tenantID = id
		span.SetAttributes(attribute.String("tenant.id", tenantID.String()))

		logger.DebugContext(ctx, "Getting tenant by ID", logger.Fields{
			"tenant_id": tenantID.String(),
			"operation": "service.GetTenant",
		})
	} else {
		// Get current tenant from database session context
		tenantID, err = s.repo.GetTenant(ctx)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "Failed to get tenant from context")

			logger.ErrorContext(ctx, "Failed to get current tenant from context", logger.Fields{
				"error":     err.Error(),
				"operation": "service.GetTenant",
			})
			return nil, fmt.Errorf("failed to get current tenant from context: %w", err)
		}

		span.SetAttributes(attribute.String("tenant.context_id", tenantID.String()))

		logger.DebugContext(ctx, "Getting current tenant from database session", logger.Fields{
			"tenant_id": tenantID.String(),
			"operation": "service.GetTenant",
		})
	}

	// Check cache first
	cacheKey := s.getTenantCacheKey(tenantID)
	var tenant Tenant
	if err := s.cache.Get(ctx, cacheKey, &tenant); err == nil {
		span.SetAttributes(attribute.Bool("cache.hit", true))
		logger.DebugContext(ctx, "Tenant found in cache", logger.Fields{
			"tenant_id": tenantID.String(),
			"operation": "service.GetTenant",
		})
		return &tenant, nil
	}

	span.SetAttributes(attribute.Bool("cache.hit", false))

	// Get from database
	dbTenant, err := s.repo.GetByID(ctx, tenantID)
	if err != nil {
		span.RecordError(err)

		if sharedErrors.IsTenantNotFound(err) {
			span.SetStatus(codes.Error, "Tenant not found")
		} else {
			span.SetStatus(codes.Error, "Database error")
		}
		return nil, err
	}

	// Cache result
	if err := s.cacheTenant(ctx, dbTenant); err != nil {
		logger.WarnContext(ctx, "Failed to cache tenant", logger.Fields{
			"tenant_id": tenantID.String(),
			"error":     err.Error(),
			"operation": "service.GetTenant",
		})
	}

	return dbTenant, nil
}

// UpdateTenant implements Service.UpdateTenant
func (s *service) UpdateTenant(ctx context.Context, id uuid.UUID, req UpdateTenantRequest) (*Tenant, error) {
	ctx, span := s.tracer.StartSpan(ctx, "tenant.service.UpdateTenant")
	defer span.End()

	span.SetAttributes(attribute.String("tenant.id", id.String()))

	logger.InfoContext(ctx, "Updating tenant", logger.Fields{
		"tenant_id": id.String(),
		"operation": "service.UpdateTenant",
	})

	// Validate input
	if err := s.validateUpdateTenantRequest(req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Invalid request")
		return nil, err
	}

	// Check if subdomain is changing and if new subdomain exists
	if req.Subdomain != nil && *req.Subdomain != "" {
		existingTenant, err := s.repo.GetBySubdomain(ctx, *req.Subdomain)
		if err == nil && existingTenant.ID != id {
			span.SetStatus(codes.Error, "Subdomain already exists")
			return nil, sharedErrors.ErrSubdomainAlreadyExists
		}
		if err != nil && !sharedErrors.IsTenantNotFound(err) {
			span.RecordError(err)
			span.SetStatus(codes.Error, "Failed to validate subdomain")
			return nil, fmt.Errorf("failed to validate subdomain: %w", err)
		}
	}

	// Clear cache before update
	if err := s.ClearTenantCache(ctx, id); err != nil {
		logger.WarnContext(ctx, "Failed to clear tenant cache before update", logger.Fields{
			"tenant_id": id.String(),
			"error":     err.Error(),
		})
	}

	// Update in database
	if err := s.repo.Update(ctx, id, req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to update tenant")
		return nil, err
	}

	// Get updated tenant
	updatedTenant, err := s.repo.GetByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to retrieve updated tenant")
		return nil, fmt.Errorf("failed to retrieve updated tenant: %w", err)
	}

	// Cache updated tenant
	if err := s.cacheTenant(ctx, updatedTenant); err != nil {
		logger.WarnContext(ctx, "Failed to cache updated tenant", logger.Fields{
			"tenant_id": id.String(),
			"error":     err.Error(),
		})
	}

	logger.InfoContext(ctx, "Tenant updated successfully", logger.Fields{
		"tenant_id":   id.String(),
		"tenant_name": updatedTenant.Name,
		"operation":   "service.UpdateTenant",
	})

	return updatedTenant, nil
}

// DeactivateTenant implements Service.DeactivateTenant
func (s *service) DeactivateTenant(ctx context.Context, id uuid.UUID) error {
	ctx, span := s.tracer.StartSpan(ctx, "tenant.service.DeactivateTenant")
	defer span.End()

	span.SetAttributes(attribute.String("tenant.id", id.String()))

	logger.InfoContext(ctx, "Deactivating tenant", logger.Fields{
		"tenant_id": id.String(),
		"operation": "service.DeactivateTenant",
	})

	// Update status to suspended (since there's no inactive status)
	suspendedStatus := StatusSuspended
	updateReq := UpdateTenantRequest{
		Status: &suspendedStatus,
	}

	_, err := s.UpdateTenant(ctx, id, updateReq)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to deactivate tenant")
		return fmt.Errorf("failed to deactivate tenant: %w", err)
	}

	logger.InfoContext(ctx, "Tenant deactivated successfully", logger.Fields{
		"tenant_id": id.String(),
		"operation": "service.DeactivateTenant",
	})

	return nil
}

// ListTenants implements Service.ListTenants
func (s *service) ListTenants(ctx context.Context, offset, limit int) ([]*Tenant, error) {
	ctx, span := s.tracer.StartSpan(ctx, "tenant.service.ListTenants")
	defer span.End()

	span.SetAttributes(
		attribute.Int("pagination.offset", offset),
		attribute.Int("pagination.limit", limit),
	)

	logger.DebugContext(ctx, "Listing tenants", logger.Fields{
		"offset":    offset,
		"limit":     limit,
		"operation": "service.ListTenants",
	})

	// Validate pagination parameters
	if offset < 0 {
		span.SetStatus(codes.Error, "Invalid offset")
		return nil, sharedErrors.NewBusinessError("INVALID_OFFSET", "Offset must be non-negative")
	}
	if limit <= 0 || limit > 100 {
		span.SetStatus(codes.Error, "Invalid limit")
		return nil, sharedErrors.NewBusinessError("INVALID_LIMIT", "Limit must be between 1 and 100")
	}

	tenants, err := s.repo.List(ctx, offset, limit)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to list tenants")
		return nil, err
	}

	span.SetAttributes(attribute.Int("result.count", len(tenants)))

	logger.DebugContext(ctx, "Tenants listed successfully", logger.Fields{
		"offset":    offset,
		"limit":     limit,
		"count":     len(tenants),
		"operation": "service.ListTenants",
	})

	return tenants, nil
}

// ActivateTenant implements Service.ActivateTenant
func (s *service) ActivateTenant(ctx context.Context, id uuid.UUID) error {
	ctx, span := s.tracer.StartSpan(ctx, "tenant.service.ActivateTenant")
	defer span.End()

	span.SetAttributes(attribute.String("tenant.id", id.String()))

	logger.InfoContext(ctx, "Activating tenant", logger.Fields{
		"tenant_id": id.String(),
		"operation": "service.ActivateTenant",
	})

	// Update status to active
	activeStatus := StatusActive
	updateReq := UpdateTenantRequest{
		Status: &activeStatus,
	}

	_, err := s.UpdateTenant(ctx, id, updateReq)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to activate tenant")
		return fmt.Errorf("failed to activate tenant: %w", err)
	}

	logger.InfoContext(ctx, "Tenant activated successfully", logger.Fields{
		"tenant_id": id.String(),
		"operation": "service.ActivateTenant",
	})

	return nil
}

// DeleteTenant implements Service.DeleteTenant
func (s *service) DeleteTenant(ctx context.Context, id uuid.UUID) error {
	ctx, span := s.tracer.StartSpan(ctx, "tenant.service.DeleteTenant")
	defer span.End()

	span.SetAttributes(attribute.String("tenant.id", id.String()))

	logger.InfoContext(ctx, "Deleting tenant", logger.Fields{
		"tenant_id": id.String(),
		"operation": "service.DeleteTenant",
	})

	// Clear cache before delete
	if err := s.ClearTenantCache(ctx, id); err != nil {
		logger.WarnContext(ctx, "Failed to clear tenant cache before delete", logger.Fields{
			"tenant_id": id.String(),
			"error":     err.Error(),
		})
	}

	// Soft delete from database
	if err := s.repo.Delete(ctx, id); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to delete tenant")
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	logger.InfoContext(ctx, "Tenant deleted successfully", logger.Fields{
		"tenant_id": id.String(),
		"operation": "service.DeleteTenant",
	})

	return nil
}

// ResolveTenantID implements Service.ResolveTenantID
// This is used by other services to resolve subdomain to tenant ID
func (s *service) ResolveTenantID(ctx context.Context, subdomain string) (uuid.UUID, error) {
	ctx, span := s.tracer.StartSpan(ctx, "tenant.service.ResolveTenantID")
	defer span.End()

	span.SetAttributes(attribute.String("tenant.subdomain", subdomain))

	logger.DebugContext(ctx, "Resolving subdomain to tenant ID", logger.Fields{
		"subdomain": subdomain,
		"operation": "service.ResolveTenantID",
	})

	// Check cache first
	cacheKey := s.getSubdomainIDCacheKey(subdomain)
	var tenantID uuid.UUID
	if err := s.cache.Get(ctx, cacheKey, &tenantID); err == nil {
		span.SetAttributes(attribute.Bool("cache.hit", true))
		logger.DebugContext(ctx, "Tenant ID found in cache", logger.Fields{
			"subdomain": subdomain,
			"tenant_id": tenantID.String(),
			"operation": "service.ResolveTenantID",
		})
		return tenantID, nil
	}

	span.SetAttributes(attribute.Bool("cache.hit", false))

	// Resolve from repository
	resolvedID, err := s.repo.ResolveSubdomainToID(ctx, subdomain)
	if err != nil {
		span.RecordError(err)
		if sharedErrors.IsTenantNotFound(err) {
			span.SetStatus(codes.Error, "Tenant not found")
		} else {
			span.SetStatus(codes.Error, "Resolution failed")
		}
		return uuid.Nil, err
	}

	span.SetAttributes(attribute.String("tenant.resolved_id", resolvedID.String()))

	// Cache the resolved ID
	if err := s.cache.Set(ctx, cacheKey, resolvedID, 30*time.Minute); err != nil {
		logger.WarnContext(ctx, "Failed to cache resolved tenant ID", logger.Fields{
			"subdomain": subdomain,
			"tenant_id": resolvedID.String(),
			"error":     err.Error(),
		})
	}

	logger.DebugContext(ctx, "Subdomain resolved to tenant ID", logger.Fields{
		"subdomain": subdomain,
		"tenant_id": resolvedID.String(),
		"operation": "service.ResolveTenantID",
	})

	return resolvedID, nil
}

// ValidateTenantAccess implements Service.ValidateTenantAccess
// This is used by other services to validate tenant access
func (s *service) ValidateTenantAccess(ctx context.Context, tenantID uuid.UUID) error {
	ctx, span := s.tracer.StartSpan(ctx, "tenant.service.ValidateTenantAccess")
	defer span.End()

	span.SetAttributes(attribute.String("tenant.id", tenantID.String()))

	logger.DebugContext(ctx, "Validating tenant access", logger.Fields{
		"tenant_id": tenantID.String(),
		"operation": "service.ValidateTenantAccess",
	})

	// Get tenant to check if it exists and is active
	tenant, err := s.GetTenantByID(ctx, tenantID)
	if err != nil {
		span.RecordError(err)
		if sharedErrors.IsTenantNotFound(err) {
			span.SetStatus(codes.Error, "Tenant not found")
		} else {
			span.SetStatus(codes.Error, "Access validation failed")
		}
		return err
	}

	// Check if tenant is active
	if tenant.Status != StatusActive {
		span.SetStatus(codes.Error, "Tenant not active")
		logger.WarnContext(ctx, "Access denied - tenant not active", logger.Fields{
			"tenant_id": tenantID.String(),
			"status":    string(tenant.Status),
			"operation": "service.ValidateTenantAccess",
		})
		return sharedErrors.NewBusinessError("TENANT_NOT_ACTIVE", "Tenant is not active")
	}

	logger.DebugContext(ctx, "Tenant access validated", logger.Fields{
		"tenant_id": tenantID.String(),
		"operation": "service.ValidateTenantAccess",
	})

	return nil
}

// ExistsTenant implements Service.ExistsTenant
func (s *service) ExistsTenant(ctx context.Context, tenantID uuid.UUID) (bool, error) {
	ctx, span := s.tracer.StartSpan(ctx, "tenant.service.ExistsTenant")
	defer span.End()

	span.SetAttributes(attribute.String("tenant.id", tenantID.String()))

	return s.repo.Exists(ctx, tenantID)
}

// GetCurrentTenant implements Service.GetCurrentTenant
func (s *service) GetCurrentTenant(ctx context.Context) (*Tenant, error) {
	ctx, span := s.tracer.StartSpan(ctx, "tenant.service.GetCurrentTenant")
	defer span.End()

	logger.DebugContext(ctx, "Getting current tenant from context", logger.Fields{
		"operation": "service.GetCurrentTenant",
	})

	return s.repo.GetCurrentTenant(ctx)
}

// ValidateCurrentTenant implements Service.ValidateCurrentTenant
func (s *service) ValidateCurrentTenant(ctx context.Context) error {
	ctx, span := s.tracer.StartSpan(ctx, "tenant.service.ValidateCurrentTenant")
	defer span.End()

	return s.repo.ValidateCurrentTenant(ctx)
}

// WithTenantContext implements Service.WithTenantContext
func (s *service) WithTenantContext(ctx context.Context, tenantID uuid.UUID, fn func(context.Context) error) error {
	ctx, span := s.tracer.StartSpan(ctx, "tenant.service.WithTenantContext")
	defer span.End()

	span.SetAttributes(attribute.String("tenant.id", tenantID.String()))

	return s.repo.WithTenantContext(ctx, tenantID, fn)
}

// ClearTenantCache implements Service.ClearTenantCache
func (s *service) ClearTenantCache(ctx context.Context, tenantID uuid.UUID) error {
	ctx, span := s.tracer.StartSpan(ctx, "tenant.service.ClearTenantCache")
	defer span.End()

	span.SetAttributes(attribute.String("tenant.id", tenantID.String()))

	logger.DebugContext(ctx, "Clearing tenant cache", logger.Fields{
		"tenant_id": tenantID.String(),
		"operation": "service.ClearTenantCache",
	})

	// Get tenant to find subdomain for cache clearing
	tenant, err := s.repo.GetByID(ctx, tenantID)
	if err != nil && !sharedErrors.IsTenantNotFound(err) {
		// Log warning but don't fail the cache clear
		logger.WarnContext(ctx, "Could not get tenant for cache clearing", logger.Fields{
			"tenant_id": tenantID.String(),
			"error":     err.Error(),
		})
	}

	var cacheErrors []error

	// Clear main tenant cache
	tenantCacheKey := s.getTenantCacheKey(tenantID)
	if err := s.cache.Delete(ctx, tenantCacheKey); err != nil {
		cacheErrors = append(cacheErrors, fmt.Errorf("failed to clear tenant cache: %w", err))
	}

	// Clear subdomain cache if tenant has subdomain
	if tenant != nil && tenant.Subdomain != nil {
		subdomainCacheKey := s.getSubdomainCacheKey(*tenant.Subdomain)
		if err := s.cache.Delete(ctx, subdomainCacheKey); err != nil {
			cacheErrors = append(cacheErrors, fmt.Errorf("failed to clear subdomain cache: %w", err))
		}

		// Clear subdomain ID cache
		subdomainIDCacheKey := s.getSubdomainIDCacheKey(*tenant.Subdomain)
		if err := s.cache.Delete(ctx, subdomainIDCacheKey); err != nil {
			cacheErrors = append(cacheErrors, fmt.Errorf("failed to clear subdomain ID cache: %w", err))
		}
	}

	if len(cacheErrors) > 0 {
		span.RecordError(fmt.Errorf("cache clear errors: %v", cacheErrors))
		logger.WarnContext(ctx, "Some cache clear operations failed", logger.Fields{
			"tenant_id": tenantID.String(),
			"errors":    fmt.Sprintf("%v", cacheErrors),
		})
	}

	return nil
}

// ClearSubdomainCache implements Service.ClearSubdomainCache
func (s *service) ClearSubdomainCache(ctx context.Context, subdomain string) error {
	ctx, span := s.tracer.StartSpan(ctx, "tenant.service.ClearSubdomainCache")
	defer span.End()

	span.SetAttributes(attribute.String("tenant.subdomain", subdomain))

	var cacheErrors []error

	// Clear subdomain cache
	subdomainCacheKey := s.getSubdomainCacheKey(subdomain)
	if err := s.cache.Delete(ctx, subdomainCacheKey); err != nil {
		cacheErrors = append(cacheErrors, fmt.Errorf("failed to clear subdomain cache: %w", err))
	}

	// Clear subdomain ID cache
	subdomainIDCacheKey := s.getSubdomainIDCacheKey(subdomain)
	if err := s.cache.Delete(ctx, subdomainIDCacheKey); err != nil {
		cacheErrors = append(cacheErrors, fmt.Errorf("failed to clear subdomain ID cache: %w", err))
	}

	if len(cacheErrors) > 0 {
		span.RecordError(fmt.Errorf("cache clear errors: %v", cacheErrors))
		return fmt.Errorf("failed to clear subdomain cache: %v", cacheErrors)
	}

	return nil
}

// WarmupCache implements Service.WarmupCache
func (s *service) WarmupCache(ctx context.Context, tenantID uuid.UUID) error {
	ctx, span := s.tracer.StartSpan(ctx, "tenant.service.WarmupCache")
	defer span.End()

	span.SetAttributes(attribute.String("tenant.id", tenantID.String()))

	logger.DebugContext(ctx, "Warming up tenant cache", logger.Fields{
		"tenant_id": tenantID.String(),
		"operation": "service.WarmupCache",
	})

	// Get tenant from database (bypassing cache)
	tenant, err := s.repo.GetByID(ctx, tenantID)
	if err != nil {
		span.RecordError(err)
		return err
	}

	// Cache the tenant
	if err := s.cacheTenant(ctx, tenant); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to warmup cache: %w", err)
	}

	logger.DebugContext(ctx, "Tenant cache warmed up", logger.Fields{
		"tenant_id": tenantID.String(),
		"operation": "service.WarmupCache",
	})

	return nil
}

// ProvisionTenant creates a new tenant along with its default configuration and
// initial usage statistics in a single atomic transaction.
func (s *service) ProvisionTenant(ctx context.Context, req ProvisionTenantRequest) (*ProvisionedTenantInfo, error) {
	ctx, span := s.tracer.StartSpan(ctx, "tenant.service.ProvisionTenant")
	defer span.End()

	// 1. Validate input request
	if err := s.validateCreateTenantRequest(CreateTenantRequest{
		Name:         req.Name,
		Email:        req.Email,
		Subdomain:    req.Subdomain,
		CountryCode:  req.CountryCode,
		CurrencyCode: req.CurrencyCode,
	}); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Invalid provisioning request")
		return nil, err
	}

	// 2. Call repository to perform provisioning within a single transaction
	provisionedInfo, err := s.repo.ProvisionTenant(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Tenant provisioning failed")
		return nil, fmt.Errorf("tenant provisioning transaction failed: %w", err)
	}

	return provisionedInfo, nil
}

// Helper methods for caching and validation

// cacheTenant caches a tenant in multiple cache keys
func (s *service) cacheTenant(ctx context.Context, tenant *Tenant) error {
	var cacheErrors []error

	// Cache by tenant ID
	tenantCacheKey := s.getTenantCacheKey(tenant.ID)
	if err := s.cache.Set(ctx, tenantCacheKey, tenant, 30*time.Minute); err != nil {
		cacheErrors = append(cacheErrors, fmt.Errorf("failed to cache tenant by ID: %w", err))
	}

	// Cache by subdomain if available
	if tenant.Subdomain != nil && *tenant.Subdomain != "" {
		subdomainCacheKey := s.getSubdomainCacheKey(*tenant.Subdomain)
		if err := s.cache.Set(ctx, subdomainCacheKey, tenant, 30*time.Minute); err != nil {
			cacheErrors = append(cacheErrors, fmt.Errorf("failed to cache tenant by subdomain: %w", err))
		}

		// Cache subdomain -> ID mapping
		subdomainIDCacheKey := s.getSubdomainIDCacheKey(*tenant.Subdomain)
		if err := s.cache.Set(ctx, subdomainIDCacheKey, tenant.ID, 30*time.Minute); err != nil {
			cacheErrors = append(cacheErrors, fmt.Errorf("failed to cache subdomain ID mapping: %w", err))
		}
	}

	if len(cacheErrors) > 0 {
		return fmt.Errorf("cache errors: %v", cacheErrors)
	}

	return nil
}

// Cache key generators
func (s *service) getTenantCacheKey(tenantID uuid.UUID) string {
	return fmt.Sprintf("tenant:id:%s", tenantID.String())
}

func (s *service) getSubdomainCacheKey(subdomain string) string {
	return fmt.Sprintf("tenant:subdomain:%s", subdomain)
}

func (s *service) getSubdomainIDCacheKey(subdomain string) string {
	return fmt.Sprintf("tenant:subdomain:id:%s", subdomain)
}

// validateCreateTenantRequest validates the create tenant request
func (s *service) validateCreateTenantRequest(req CreateTenantRequest) error {
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		for _, e := range validationErrors {
			switch e.Tag() {
			case "required":
				if e.Field() == "Name" {
					return sharedErrors.NewBusinessError("INVALID_TENANT_NAME", "Tenant name is required")
				} else if e.Field() == "Email" {
					return sharedErrors.NewBusinessError("INVALID_EMAIL", "Email is required")
				}
			case "email":
				return sharedErrors.NewBusinessError("INVALID_EMAIL", "Invalid email format")
			case "iso3166_1_alpha2":
				return sharedErrors.NewBusinessError("INVALID_COUNTRY_CODE", "Invalid country code format")
			case "iso4217":
				return sharedErrors.NewBusinessError("INVALID_CURRENCY_CODE", "Invalid currency code format")
			}
		}
		return err // Fallback for other validation errors
	}

	if req.Subdomain != nil && *req.Subdomain != "" {
		if len(*req.Subdomain) > 63 {
			return sharedErrors.NewBusinessError("INVALID_SUBDOMAIN", "Subdomain must be 63 characters or less")
		}
		// Basic subdomain validation (alphanumeric and hyphens)
		for _, char := range *req.Subdomain {
			if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '-') {
				return sharedErrors.NewBusinessError("INVALID_SUBDOMAIN", "Subdomain can only contain alphanumeric characters and hyphens")
			}
		}
	}
	return nil
}

// validateUpdateTenantRequest validates the update tenant request
func (s *service) validateUpdateTenantRequest(req UpdateTenantRequest) error {
	if req.Name != nil && *req.Name == "" {
		return sharedErrors.NewBusinessError("INVALID_TENANT_NAME", "Tenant name cannot be empty")
	}
	if req.Name != nil && len(*req.Name) > 255 {
		return sharedErrors.NewBusinessError("INVALID_TENANT_NAME", "Tenant name must be 255 characters or less")
	}
	if req.Subdomain != nil && *req.Subdomain != "" {
		if len(*req.Subdomain) > 63 {
			return sharedErrors.NewBusinessError("INVALID_SUBDOMAIN", "Subdomain must be 63 characters or less")
		}
		// Basic subdomain validation
		for _, char := range *req.Subdomain {
			if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '-') {
				return sharedErrors.NewBusinessError("INVALID_SUBDOMAIN", "Subdomain can only contain alphanumeric characters and hyphens")
			}
		}
	}
	return nil
}

// SetTenant implements Service.SetTenant
// Sets the tenant context in the database session
func (s *service) SetTenant(ctx context.Context, tenantID uuid.UUID) error {
	ctx, span := s.tracer.StartSpan(ctx, "tenant.service.SetTenant")
	defer span.End()

	span.SetAttributes(attribute.String("tenant.id", tenantID.String()))

	logger.DebugContext(ctx, "Setting tenant context via service", logger.Fields{
		"tenant_id": tenantID.String(),
		"operation": "service.SetTenant",
	})

	// Validate tenant exists and is active before setting context
	if err := s.ValidateTenantAccess(ctx, tenantID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Tenant validation failed")
		return fmt.Errorf("failed to validate tenant before setting context: %w", err)
	}

	// Set in repository (database session)
	err := s.repo.SetTenant(ctx, tenantID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to set tenant context")
		return err
	}

	logger.InfoContext(ctx, "Successfully set tenant context", logger.Fields{
		"tenant_id": tenantID.String(),
		"operation": "service.SetTenant",
	})

	return nil
}

// ResetTenant implements Service.ResetTenant
// Resets/clears the tenant context in the database session
func (s *service) ResetTenant(ctx context.Context) error {
	ctx, span := s.tracer.StartSpan(ctx, "tenant.service.ResetTenant")
	defer span.End()

	logger.DebugContext(ctx, "Resetting tenant context via service", logger.Fields{
		"operation": "service.ResetTenant",
	})

	// Reset in repository (database session)
	err := s.repo.ResetTenant(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to reset tenant context")
		return err
	}

	logger.InfoContext(ctx, "Successfully reset tenant context", logger.Fields{
		"operation": "service.ResetTenant",
	})

	return nil
}
