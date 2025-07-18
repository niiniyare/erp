package tenant

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/platform/cache"
	sharedErrors "github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
)

// Service defines tenant business logic interface
type Service interface {
	CreateTenant(ctx context.Context, req CreateTenantRequest) (*Tenant, error)
	GetTenant(ctx context.Context, id uuid.UUID) (*Tenant, error)
	GetTenantBySubdomain(ctx context.Context, subdomain string) (*Tenant, error)
	UpdateTenant(ctx context.Context, id uuid.UUID, req UpdateTenantRequest) error
	DeactivateTenant(ctx context.Context, id uuid.UUID) error
	ListTenants(ctx context.Context, offset, limit int) ([]*Tenant, error)
}

// service implements Service interface
type service struct {
	repo  Repository
	cache cache.Service
}

// NewService creates a new tenant service
func NewService(repo Repository, cache cache.Service) Service {
	return &service{
		repo:  repo,
		cache: cache,
	}
}

// CreateTenant implements Service.CreateTenant
func (s *service) CreateTenant(ctx context.Context, req CreateTenantRequest) (*Tenant, error) {
	logger.InfoContext(ctx, "Creating new tenant", logger.Fields{
		"tenant_name": req.Name,
		"subdomain":   req.Subdomain,
		"industry":    req.Industry,
	})

	// Validate subdomain uniqueness if provided
	if req.Subdomain != nil && *req.Subdomain != "" {
		exists, err := s.repo.Exists(ctx, *req.Subdomain)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to check subdomain existence", logger.Fields{
				"subdomain": *req.Subdomain,
				"error":     err.Error(),
			})
			return nil, fmt.Errorf("failed to check subdomain existence: %w", err)
		}
		if exists {
			logger.WarnContext(ctx, "Subdomain already exists", logger.Fields{
				"subdomain": *req.Subdomain,
			})
			return nil, sharedErrors.ErrSubdomainAlreadyExists
		}
	}

	// Set default status if not provided
	status := req.Status
	if status == "" {
		status = StatusActive
	}

	// Create tenant entity
	tenant := &Tenant{
		ID:                 uuid.New(),
		Slug:               req.Slug,
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
		logger.ErrorContext(ctx, "Failed to create tenant in database", logger.Fields{
			"tenant_id":   tenant.ID.String(),
			"tenant_name": tenant.Name,
			"error":       err.Error(),
		})
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	logger.InfoContext(ctx, "Tenant created successfully", logger.Fields{
		"tenant_id":   tenant.ID.String(),
		"tenant_name": tenant.Name,
		"subdomain":   tenant.Subdomain,
		"status":      string(tenant.Status),
	})

	// Cache the tenant if subdomain is provided
	if tenant.Subdomain != nil {
		cacheKey := fmt.Sprintf("tenant:subdomain:%s", *tenant.Subdomain)
		if err := s.cache.Set(ctx, cacheKey, tenant, 30*time.Minute); err != nil {
			logger.WarnContext(ctx, "Failed to cache tenant", logger.Fields{
				"tenant_id": tenant.ID.String(),
				"cache_key": cacheKey,
				"error":     err.Error(),
			})
		}
	}

	return tenant, nil
}

// GetTenantBySubdomain implements Service.GetTenantBySubdomain with caching
func (s *service) GetTenantBySubdomain(ctx context.Context, subdomain string) (*Tenant, error) {
	logger.DebugContext(ctx, "Getting tenant by subdomain", logger.Fields{
		"subdomain": subdomain,
	})

	// Check cache first
	cacheKey := fmt.Sprintf("tenant:subdomain:%s", subdomain)
	var tenant Tenant
	if err := s.cache.Get(ctx, cacheKey, &tenant); err == nil {
		logger.DebugContext(ctx, "Tenant found in cache", logger.Fields{
			"subdomain": subdomain,
			"tenant_id": tenant.ID.String(),
		})
		return &tenant, nil
	}

	// Get from database
	dbTenant, err := s.repo.GetBySubdomain(ctx, subdomain)
	if err != nil {
		if errors.Is(err, sharedErrors.ErrTenantNotFound) {
			logger.WarnContext(ctx, "Tenant not found by subdomain", logger.Fields{
				"subdomain": subdomain,
			})
		} else {
			logger.ErrorContext(ctx, "Failed to get tenant by subdomain", logger.Fields{
				"subdomain": subdomain,
				"error":     err.Error(),
			})
		}
		return nil, err
	}

	logger.InfoContext(ctx, "Tenant found in database", logger.Fields{
		"subdomain":   subdomain,
		"tenant_id":   dbTenant.ID.String(),
		"tenant_name": dbTenant.Name,
	})

	// Cache result
	if err := s.cache.Set(ctx, cacheKey, dbTenant, 30*time.Minute); err != nil {
		logger.WarnContext(ctx, "Failed to cache tenant", logger.Fields{
			"subdomain": subdomain,
			"tenant_id": dbTenant.ID.String(),
			"error":     err.Error(),
		})
	}

	return dbTenant, nil
}

// GetTenant implements Service.GetTenant
func (s *service) GetTenant(ctx context.Context, id uuid.UUID) (*Tenant, error) {
	return s.repo.GetByID(ctx, id)
}

// UpdateTenant implements Service.UpdateTenant
func (s *service) UpdateTenant(ctx context.Context, id uuid.UUID, req UpdateTenantRequest) error {
	return s.repo.Update(ctx, id, req)
}

// DeactivateTenant implements Service.DeactivateTenant
func (s *service) DeactivateTenant(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// ListTenants implements Service.ListTenants
func (s *service) ListTenants(ctx context.Context, offset, limit int) ([]*Tenant, error) {
	return s.repo.List(ctx, offset, limit)
}
