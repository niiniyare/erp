package tenant

import (
	"context"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/repo/tenant"
)

type TenantService struct {
	repo tenant.
}

// Primary Ports (Driving - Inbound)
type TenantService interface {
	// Admin operations (system-wide)
	CreateTenant(ctx context.Context, req CreateTenantRequest) (*Tenant, error)
	GetTenantByID(ctx context.Context, id int32) (*Tenant, error)
	GetTenantByUUID(ctx context.Context, uuid uuid.UUID) (*Tenant, error)
	GetTenantBySubdomain(ctx context.Context, subdomain string) (*Tenant, error)
	ListTenants(ctx context.Context, filter TenantFilter) ([]*Tenant, error)
	UpdateTenant(ctx context.Context, id int32, req UpdateTenantRequest) (*Tenant, error)
	DeleteTenant(ctx context.Context, id int32) error
	GetTenantStats(ctx context.Context) (*TenantStats, error)

	// Current tenant operations (RLS-aware)
	GetCurrentTenant(ctx context.Context) (*Tenant, error)
	UpdateCurrentTenant(ctx context.Context, req UpdateTenantRequest) (*Tenant, error)
}

type TenantConfigService interface {
	// Admin operations
	CreateTenantConfiguration(ctx context.Context, tenantID int32, req CreateTenantConfigRequest) (*TenantConfiguration, error)
	GetTenantConfiguration(ctx context.Context, tenantID int32) (*TenantConfiguration, error)
	UpdateTenantConfiguration(ctx context.Context, tenantID int32, req UpdateTenantConfigRequest) (*TenantConfiguration, error)
	DeleteTenantConfiguration(ctx context.Context, tenantID int32) error

	// Current tenant operations (RLS-aware)
	GetCurrentTenantConfiguration(ctx context.Context) (*TenantConfiguration, error)
	UpdateCurrentTenantConfiguration(ctx context.Context, req UpdateTenantConfigRequest) (*TenantConfiguration, error)
	CreateCurrentTenantConfiguration(ctx context.Context, req CreateTenantConfigRequest) (*TenantConfiguration, error)

	// Feature and module checks
	CheckCurrentTenantHasFeature(ctx context.Context, feature string) (bool, error)
	CheckCurrentTenantHasModule(ctx context.Context, module string) (bool, error)
}
