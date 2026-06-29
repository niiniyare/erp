package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"awo.so/internal/core/tenant/domain"
	"awo.so/internal/core/tenant/repository"
	"awo.so/internal/platform/cache"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/tracing"
	"awo.so/internal/shared/utils"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

const tenantCacheTTL = 30 * time.Minute

// TenantService handles core tenant CRUD and lifecycle operations.
type TenantService struct {
	repo   repository.Repository
	cache  cache.Service
	tracer tracing.Service
	log    logger.Logger
}

// NewTenantService creates a new TenantService.
func NewTenantService(repo repository.Repository, c cache.Service, tracer tracing.Service, log logger.Logger) *TenantService {
	return &TenantService{repo: repo, cache: c, tracer: tracer, log: log}
}

// cacheCtx returns a context enriched with tenant cache namespace.
func cacheCtx(ctx context.Context, tenantID uuid.UUID) context.Context {
	ctx = cache.WithTenantID(ctx, tenantID)
	ctx = cache.WithNamespace(ctx, "tenant")
	return ctx
}

// Create validates and creates a new tenant.
func (s *TenantService) Create(ctx context.Context, req domain.CreateTenantRequest) (*domain.Tenant, error) {
	ctx, span := s.tracer.StartSpan(ctx, "tenant.svc.Create")
	defer span.End()
	span.SetAttributes(attribute.String("tenant.name", req.Name))

	// Validate
	if err := s.validateCreate(req); err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Check subdomain uniqueness
	if req.Subdomain != nil && *req.Subdomain != "" {
		exists, err := s.repo.SubdomainExists(ctx, *req.Subdomain)
		if err != nil {
			return nil, fmt.Errorf("failed to check subdomain: %w", err)
		}
		if exists {
			return nil, domain.ErrSubdomainTaken
		}
	}

	status := req.Status
	if status == "" {
		status = domain.StatusPending
	}

	t, err := domain.NewTenant(req.Name, req.Email,
		domain.WithStatus(status),
	)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	if req.Slug != "" {
		t.Slug = req.Slug
	}
	t.Subdomain = req.Subdomain
	t.Industry = req.Industry
	t.CompanySize = req.CompanySize
	t.TaxID = req.TaxID
	t.RegistrationNumber = req.RegistrationNumber
	t.LegalEntityType = req.LegalEntityType
	t.Settings = req.Settings
	if req.CurrencyCode != "" {
		t.CurrencyCode = req.CurrencyCode
	}

	if err := s.repo.Create(ctx, t); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "create failed")
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	// Populate cache
	s.cacheTenant(ctx, t)
	return t, nil
}

// GetByID returns a tenant by ID, checking cache first.
func (s *TenantService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	ctx, span := s.tracer.StartSpan(ctx, "tenant.svc.GetByID")
	defer span.End()
	span.SetAttributes(attribute.String("tenant.id", id.String()))

	// Try cache
	cctx := cacheCtx(ctx, id)
	var cached domain.Tenant
	if err := s.cache.Get(cctx, "id:"+id.String(), &cached); err == nil {
		span.SetAttributes(attribute.Bool("cache.hit", true))
		return &cached, nil
	}
	span.SetAttributes(attribute.Bool("cache.hit", false))

	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	s.cacheTenant(ctx, t)
	return t, nil
}

// GetBySubdomain returns a tenant by subdomain, checking cache first.
func (s *TenantService) GetBySubdomain(ctx context.Context, subdomain string) (*domain.Tenant, error) {
	ctx, span := s.tracer.StartSpan(ctx, "tenant.svc.GetBySubdomain")
	defer span.End()
	span.SetAttributes(attribute.String("tenant.subdomain", subdomain))

	t, err := s.repo.GetBySubdomain(ctx, subdomain)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	s.cacheTenant(ctx, t)
	return t, nil
}

// Update applies changes to a tenant.
func (s *TenantService) Update(ctx context.Context, id uuid.UUID, req domain.UpdateTenantRequest) (*domain.Tenant, error) {
	ctx, span := s.tracer.StartSpan(ctx, "tenant.svc.Update")
	defer span.End()
	span.SetAttributes(attribute.String("tenant.id", id.String()))

	if err := s.validateUpdate(req); err != nil {
		return nil, err
	}

	// Check subdomain uniqueness if changing
	if req.Subdomain != nil && *req.Subdomain != "" {
		existing, err := s.repo.GetBySubdomain(ctx, *req.Subdomain)
		if err == nil && existing.ID != id {
			return nil, domain.ErrSubdomainTaken
		}
		if err != nil && err != domain.ErrTenantNotFound {
			return nil, fmt.Errorf("failed to check subdomain: %w", err)
		}
	}

	s.invalidateCache(ctx, id)

	if err := s.repo.Update(ctx, id, req); err != nil {
		span.RecordError(err)
		return nil, err
	}

	updated, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	s.cacheTenant(ctx, updated)
	return updated, nil
}

// Activate transitions a tenant to Active status.
func (s *TenantService) Activate(ctx context.Context, id uuid.UUID) error {
	ctx, span := s.tracer.StartSpan(ctx, "tenant.svc.Activate")
	defer span.End()

	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := t.Activate(); err != nil {
		return err
	}

	status := domain.StatusActive
	if err := s.repo.Update(ctx, id, domain.UpdateTenantRequest{Status: &status}); err != nil {
		return err
	}

	s.invalidateCache(ctx, id)
	return nil
}

// Suspend transitions a tenant to Suspended status.
func (s *TenantService) Suspend(ctx context.Context, id uuid.UUID, reason string) error {
	ctx, span := s.tracer.StartSpan(ctx, "tenant.svc.Suspend")
	defer span.End()

	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := t.Suspend(reason); err != nil {
		return err
	}

	status := domain.StatusSuspended
	if err := s.repo.Update(ctx, id, domain.UpdateTenantRequest{Status: &status}); err != nil {
		return err
	}

	s.invalidateCache(ctx, id)
	return nil
}

// Archive soft-deletes a tenant.
func (s *TenantService) Archive(ctx context.Context, id uuid.UUID) error {
	ctx, span := s.tracer.StartSpan(ctx, "tenant.svc.Archive")
	defer span.End()

	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := t.Archive(); err != nil {
		return err
	}

	if err := s.repo.SoftDelete(ctx, id); err != nil {
		return err
	}

	s.invalidateCache(ctx, id)
	return nil
}

// Delete soft-deletes a tenant (alias for SoftDelete).
func (s *TenantService) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, span := s.tracer.StartSpan(ctx, "tenant.svc.Delete")
	defer span.End()

	s.invalidateCache(ctx, id)
	return s.repo.SoftDelete(ctx, id)
}

// List returns tenants matching the filter with SQL pagination.
func (s *TenantService) List(ctx context.Context, filter domain.TenantFilter) ([]*domain.Tenant, int64, error) {
	ctx, span := s.tracer.StartSpan(ctx, "tenant.svc.List")
	defer span.End()

	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 20
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	return s.repo.List(ctx, filter)
}

// ResolveTenantID resolves a subdomain to a tenant UUID.
func (s *TenantService) ResolveTenantID(ctx context.Context, subdomain string) (uuid.UUID, error) {
	t, err := s.GetBySubdomain(ctx, subdomain)
	if err != nil {
		return uuid.Nil, err
	}
	return t.ID, nil
}

// ValidateTenantAccess checks that a tenant exists and is active.
func (s *TenantService) ValidateTenantAccess(ctx context.Context, tenantID uuid.UUID) error {
	t, err := s.GetByID(ctx, tenantID)
	if err != nil {
		return err
	}
	if !t.IsActive() {
		return domain.ErrTenantSuspended
	}
	return nil
}

// ExistsTenant checks if a tenant exists.
func (s *TenantService) ExistsTenant(ctx context.Context, tenantID uuid.UUID) (bool, error) {
	return s.repo.Exists(ctx, tenantID)
}

// --- cache helpers ---

func (s *TenantService) cacheTenant(ctx context.Context, t *domain.Tenant) {
	cctx := cacheCtx(ctx, t.ID)
	if err := s.cache.Set(cctx, "id:"+t.ID.String(), t, tenantCacheTTL); err != nil {
		s.log.WarnContext(ctx, "failed to cache tenant", logger.Fields{"tenant_id": t.ID.String(), "error": err.Error()})
	}
}

func (s *TenantService) invalidateCache(ctx context.Context, id uuid.UUID) {
	cctx := cacheCtx(ctx, id)
	if err := s.cache.Delete(cctx, "id:"+id.String()); err != nil {
		s.log.WarnContext(ctx, "failed to invalidate cache", logger.Fields{"tenant_id": id.String(), "error": err.Error()})
	}
}

// --- validation ---

func (s *TenantService) validateCreate(req domain.CreateTenantRequest) error {
	if err := utils.ValidateStruct(req); err != nil {
		return fmt.Errorf("%w: %v", domain.ErrInvalidRequest, err)
	}
	if req.Subdomain != nil && *req.Subdomain != "" {
		if len(*req.Subdomain) > 63 {
			return fmt.Errorf("%w: subdomain too long", domain.ErrInvalidRequest)
		}
		// Check reserved subdomains
		if domain.ReservedSubdomains[strings.ToLower(*req.Subdomain)] {
			return domain.ErrInvalidSubdomain
		}
	}
	// Validate company size if provided
	if req.CompanySize != nil && *req.CompanySize != "" {
		if !domain.ValidCompanySize(*req.CompanySize) {
			return domain.ErrInvalidCompanySize
		}
	}
	return nil
}

func (s *TenantService) validateUpdate(req domain.UpdateTenantRequest) error {
	if req.Name != nil && *req.Name == "" {
		return fmt.Errorf("%w: name cannot be empty", domain.ErrInvalidRequest)
	}
	if req.Name != nil && len(*req.Name) > 255 {
		return fmt.Errorf("%w: name too long", domain.ErrInvalidRequest)
	}
	if req.Subdomain != nil && *req.Subdomain != "" && len(*req.Subdomain) > 63 {
		return fmt.Errorf("%w: subdomain too long", domain.ErrInvalidRequest)
	}
	return nil
}
