package tenant

//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"

import (
	"context"

	"github.com/google/uuid"
	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/core/tenant/domain"
	"github.com/niiniyare/erp/internal/core/tenant/repository"
	"github.com/niiniyare/erp/internal/core/tenant/service"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// Re-export domain types for backward compatibility.
type (
	Tenant              = domain.Tenant
	TenantStatus        = domain.TenantStatus
	TenantConfiguration = domain.TenantConfiguration
	TenantUsage         = domain.TenantUsage
	CreateTenantRequest = domain.CreateTenantRequest
	UpdateTenantRequest = domain.UpdateTenantRequest
	TenantFilter        = domain.TenantFilter
	TenantLimits        = domain.TenantLimits
	PlanType            = domain.PlanType
	CompanySize         = domain.CompanySize
	AccountingMethod    = domain.AccountingMethod
	ProvisioningInput   = domain.ProvisioningInput
	ProvisioningResult  = domain.ProvisioningResult
	GrowthStat          = domain.GrowthStat
	StatusCount         = domain.StatusCount
)

// Re-export status constants.
const (
	StatusActive    = domain.StatusActive
	StatusSuspended = domain.StatusSuspended
	StatusPending   = domain.StatusPending
	StatusArchived  = domain.StatusArchived
	StatusTrial     = domain.StatusTrial
)

// Re-export plan constants.
const (
	PlanBasic        = domain.PlanBasic
	PlanProfessional = domain.PlanProfessional
	PlanEnterprise   = domain.PlanEnterprise
)

// Re-export company size constants.
const (
	CompanySizeStartup    = domain.CompanySizeStartup
	CompanySizeSmall      = domain.CompanySizeSmall
	CompanySizeMedium     = domain.CompanySizeMedium
	CompanySizeLarge      = domain.CompanySizeLarge
	CompanySizeEnterprise = domain.CompanySizeEnterprise
)

// Re-export errors.
var (
	ErrTenantNotFound               = domain.ErrTenantNotFound
	ErrTenantAlreadyExists          = domain.ErrTenantAlreadyExists
	ErrSubdomainTaken               = domain.ErrSubdomainTaken
	ErrInvalidTransition            = domain.ErrInvalidTransition
	ErrTenantSuspended              = domain.ErrTenantSuspended
	ErrLimitExceeded                = domain.ErrLimitExceeded
	ErrTenantNameRequired           = domain.ErrTenantNameRequired
	ErrTenantEmailRequired          = domain.ErrTenantEmailRequired
	ErrInvalidEmail                 = domain.ErrInvalidEmail
	ErrInvalidSubdomain             = domain.ErrInvalidSubdomain
	ErrInvalidCompanySize           = domain.ErrInvalidCompanySize
	ErrAlreadyActive                = domain.ErrAlreadyActive
	ErrAlreadySuspended             = domain.ErrAlreadySuspended
	ErrAlreadyArchived              = domain.ErrAlreadyArchived
	ErrCannotActivateArchivedTenant = domain.ErrCannotActivateArchivedTenant
	ErrCannotSuspendArchivedTenant  = domain.ErrCannotSuspendArchivedTenant
	ErrConfigurationNotFound        = domain.ErrConfigurationNotFound
)

// Service is the backward-compatible tenant service interface that
// matches what handlers currently expect.
type Service interface {
	// Core CRUD
	CreateTenant(ctx context.Context, req CreateTenantRequest) (*Tenant, error)
	GetTenantByID(ctx context.Context, id uuid.UUID) (*Tenant, error)
	GetTenantBySubdomain(ctx context.Context, subdomain string) (*Tenant, error)
	UpdateTenant(ctx context.Context, id uuid.UUID, req UpdateTenantRequest) (*Tenant, error)
	DeleteTenant(ctx context.Context, id uuid.UUID) error
	ActivateTenant(ctx context.Context, id uuid.UUID) error
	SuspendTenant(ctx context.Context, id uuid.UUID, reason string) error
	ArchiveTenant(ctx context.Context, id uuid.UUID) error
	DeactivateTenant(ctx context.Context, id uuid.UUID) error
	ListTenants(ctx context.Context, filter TenantFilter) ([]*Tenant, int64, error)

	// Provisioning
	ProvisionTenant(ctx context.Context, req ProvisionTenantRequest) (*ProvisionedTenantInfo, error)

	// Tenant resolution
	ResolveTenantID(ctx context.Context, subdomain string) (uuid.UUID, error)
	ValidateTenantAccess(ctx context.Context, tenantID uuid.UUID) error
	ExistsTenant(ctx context.Context, tenantID uuid.UUID) (bool, error)

	// Context ops
	GetCurrentTenant(ctx context.Context) (*Tenant, error)
	ValidateCurrentTenant(ctx context.Context) error
	WithTenantContext(ctx context.Context, tenantID uuid.UUID, fn func(context.Context) error) error

	// Cache
	ClearTenantCache(ctx context.Context, tenantID uuid.UUID) error
	ClearSubdomainCache(ctx context.Context, subdomain string) error
	WarmupCache(ctx context.Context, tenantID uuid.UUID) error

	// RLS session
	SetTenant(ctx context.Context, tenantID uuid.UUID) error
	ResetTenant(ctx context.Context) error
}

// ProvisionTenantRequest — kept for backward compat with old provisioning types.
type ProvisionTenantRequest = domain.ProvisioningInput

// ProvisionedTenantInfo — kept for backward compat.
type ProvisionedTenantInfo = domain.ProvisioningResult

// Repository is the backward-compatible repository interface.
type Repository = repository.Repository

// NewRepository creates a new tenant repository (backward compat).
func NewRepository(store db.Store, tracer tracing.Service) Repository {
	return repository.NewPostgres(store, tracer)
}

// Dependencies holds all dependencies needed to create a tenant service.
type Dependencies struct {
	Store  db.Store
	Cache  cache.Service
	Tracer tracing.Service
	Logger logger.Logger
}

// tenantServiceAdapter wraps the new service layer to implement the old Service interface.
type tenantServiceAdapter struct {
	tenant       *service.TenantService
	provisioning *service.ProvisioningService
	analytics    *service.AnalyticsService
	repo         repository.Repository
	store        db.Store
	cache        cache.Service
	tracer       tracing.Service
	log          logger.Logger
}

// NewService wires everything together and returns the backward-compatible Service.
func NewService(deps Dependencies) Service {
	repo := repository.NewPostgres(deps.Store, deps.Tracer)
	return &tenantServiceAdapter{
		tenant:       service.NewTenantService(repo, deps.Cache, deps.Tracer, deps.Logger),
		provisioning: service.NewProvisioningService(repo, deps.Tracer),
		analytics:    service.NewAnalyticsService(repo, deps.Cache, deps.Tracer, deps.Logger),
		repo:         repo,
		store:        deps.Store,
		cache:        deps.Cache,
		tracer:       deps.Tracer,
		log:          deps.Logger,
	}
}

// Tenant returns the TenantService for callers that want the new API.
func (a *tenantServiceAdapter) Tenant() *service.TenantService { return a.tenant }

// Provisioning returns the ProvisioningService.
func (a *tenantServiceAdapter) Provisioning() *service.ProvisioningService { return a.provisioning }

// Analytics returns the AnalyticsService.
func (a *tenantServiceAdapter) Analytics() *service.AnalyticsService { return a.analytics }

// --- backward-compatible method implementations ---

func (a *tenantServiceAdapter) CreateTenant(ctx context.Context, req CreateTenantRequest) (*Tenant, error) {
	return a.tenant.Create(ctx, req)
}

func (a *tenantServiceAdapter) GetTenantByID(ctx context.Context, id uuid.UUID) (*Tenant, error) {
	return a.tenant.GetByID(ctx, id)
}

func (a *tenantServiceAdapter) GetTenantBySubdomain(ctx context.Context, subdomain string) (*Tenant, error) {
	return a.tenant.GetBySubdomain(ctx, subdomain)
}

func (a *tenantServiceAdapter) UpdateTenant(ctx context.Context, id uuid.UUID, req UpdateTenantRequest) (*Tenant, error) {
	return a.tenant.Update(ctx, id, req)
}

func (a *tenantServiceAdapter) DeleteTenant(ctx context.Context, id uuid.UUID) error {
	return a.tenant.Delete(ctx, id)
}

func (a *tenantServiceAdapter) ActivateTenant(ctx context.Context, id uuid.UUID) error {
	return a.tenant.Activate(ctx, id)
}

func (a *tenantServiceAdapter) SuspendTenant(ctx context.Context, id uuid.UUID, reason string) error {
	return a.tenant.Suspend(ctx, id, reason)
}

func (a *tenantServiceAdapter) ArchiveTenant(ctx context.Context, id uuid.UUID) error {
	return a.tenant.Archive(ctx, id)
}

func (a *tenantServiceAdapter) DeactivateTenant(ctx context.Context, id uuid.UUID) error {
	return a.tenant.Suspend(ctx, id, "deactivated")
}

func (a *tenantServiceAdapter) ListTenants(ctx context.Context, filter TenantFilter) ([]*Tenant, int64, error) {
	return a.tenant.List(ctx, filter)
}

func (a *tenantServiceAdapter) ProvisionTenant(ctx context.Context, req ProvisionTenantRequest) (*ProvisionedTenantInfo, error) {
	return a.provisioning.Provision(ctx, req)
}

func (a *tenantServiceAdapter) ResolveTenantID(ctx context.Context, subdomain string) (uuid.UUID, error) {
	return a.tenant.ResolveTenantID(ctx, subdomain)
}

func (a *tenantServiceAdapter) ValidateTenantAccess(ctx context.Context, tenantID uuid.UUID) error {
	return a.tenant.ValidateTenantAccess(ctx, tenantID)
}

func (a *tenantServiceAdapter) ExistsTenant(ctx context.Context, tenantID uuid.UUID) (bool, error) {
	return a.tenant.ExistsTenant(ctx, tenantID)
}

func (a *tenantServiceAdapter) GetCurrentTenant(ctx context.Context) (*Tenant, error) {
	return a.repo.GetCurrentTenant(ctx)
}

func (a *tenantServiceAdapter) ValidateCurrentTenant(ctx context.Context) error {
	return a.repo.ValidateCurrentTenant(ctx)
}

func (a *tenantServiceAdapter) WithTenantContext(ctx context.Context, tenantID uuid.UUID, fn func(context.Context) error) error {
	return a.store.WithTenant(ctx, tenantID, func(ctx context.Context, _ db.Store) error {
		return fn(ctx)
	})
}

func (a *tenantServiceAdapter) ClearTenantCache(ctx context.Context, tenantID uuid.UUID) error {
	cctx := cache.WithTenantID(ctx, tenantID)
	cctx = cache.WithNamespace(cctx, "tenant")
	return a.cache.Delete(cctx, "id:"+tenantID.String())
}

func (a *tenantServiceAdapter) ClearSubdomainCache(ctx context.Context, subdomain string) error {
	return nil
}

func (a *tenantServiceAdapter) WarmupCache(ctx context.Context, tenantID uuid.UUID) error {
	t, err := a.repo.GetByID(ctx, tenantID)
	if err != nil {
		return err
	}
	cctx := cache.WithTenantID(ctx, t.ID)
	cctx = cache.WithNamespace(cctx, "tenant")
	return a.cache.Set(cctx, "id:"+t.ID.String(), t, 30*60*1e9) // 30 min
}

func (a *tenantServiceAdapter) SetTenant(ctx context.Context, tenantID uuid.UUID) error {
	if err := a.tenant.ValidateTenantAccess(ctx, tenantID); err != nil {
		return err
	}
	return a.store.SetTenantContext(ctx, tenantID)
}

func (a *tenantServiceAdapter) ResetTenant(ctx context.Context) error {
	return a.store.ResetTenantContext(ctx)
}
