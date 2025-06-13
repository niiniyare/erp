package repositories

import (
	"encoding/`json"
)

type CreateTenantRequest struct {
	Name      string  `json:"name" validate:"required,min=1,max=255"`
	Subdomain *string `json:"subdomain,omitempty" validate:"omitempty,min=1,max=63"`
	Status    string  `json:"status" validate:"required,oneof=active suspended pending"`
	Industry  *string `json:"industry,omitempty" validate:"omitempty,max=50"`
}

type UpdateTenantRequest struct {
	Name      string  `json:"name" validate:"required,min=1,max=255"`
	Subdomain *string `json:"subdomain,omitempty" validate:"omitempty,min=1,max=63"`
	Status    string  `json:"status" validate:"required,oneof=active suspended pending"`
	Industry  *string `json:"industry,omitempty" validate:"omitempty,max=50"`
}

type CreateTenantConfigRequest struct {
	MaxUsers       int32           `json:"max_users" validate:"min=1"`
	StorageQuota   int64           `json:"storage_quota" validate:"min=1"`
	Features       json.RawMessage `json:"features"`
	ModulesEnabled json.RawMessage `json:"modules_enabled"`
}

type UpdateTenantConfigRequest struct {
	MaxUsers       *int32           `json:"max_users,omitempty" validate:"omitempty,min=1"`
	StorageQuota   *int64           `json:"storage_quota,omitempty" validate:"omitempty,min=1"`
	Features       *json.RawMessage `json:"features,omitempty"`
	ModulesEnabled *json.RawMessage `json:"modules_enabled,omitempty"`
}

type TenantFilter struct {
	Name     *string `json:"name,omitempty"`
	Status   *string `json:"status,omitempty"`
	Industry *string `json:"industry,omitempty"`
	SortBy   string  `json:"sort_by,omitempty"`
	Limit    int32   `json:"limit"`
	Offset   int32   `json:"offset"`
}

type TenantStats struct {
	TotalTenants     int64 `json:"total_tenants"`
	ActiveTenants    int64 `json:"active_tenants"`
	SuspendedTenants int64 `json:"suspended_tenants"`
	PendingTenants   int64 `json:"pending_tenants"`
}



// Secondary Ports (Driven - Outbound)
type TenantRepository interface {
	// Admin operations
	CreateTenant(ctx context.Context, arg CreateTenantParams) (*Tenant, error)
	GetTenantByID(ctx context.Context, id int32) (*Tenant, error)
	GetTenantByUUID(ctx context.Context, uuid uuid.UUID) (*Tenant, error)
	GetTenantBySubdomain(ctx context.Context, subdomain string) (*Tenant, error)
	ListTenants(ctx context.Context, arg ListTenantsParams) ([]*Tenant, error)
	UpdateTenant(ctx context.Context, arg UpdateTenantParams) (*Tenant, error)
	SoftDeleteTenant(ctx context.Context, id int32) error
	GetTenantStats(ctx context.Context) (*TenantStats, error)
	CheckTenantExists(ctx context.Context, id int32) (bool, error)
	CheckSubdomainExists(ctx context.Context, subdomain string) (bool, error)
	CheckTenantNameExists(ctx context.Context, name string) (bool, error)
	
	// Current tenant operations
	GetCurrentTenant(ctx context.Context) (*Tenant, error)
	UpdateCurrentTenant(ctx context.Context, arg UpdateCurrentTenantParams) (*Tenant, error)
}

type TenantConfigRepository interface {
	// Admin operations
	CreateTenantConfiguration(ctx context.Context, arg CreateTenantConfigurationParams) (*TenantConfiguration, error)
	GetTenantConfiguration(ctx context.Context, tenantID int32) (*TenantConfiguration, error)
	UpdateTenantConfiguration(ctx context.Context, arg UpdateTenantConfigurationParams) (*TenantConfiguration, error)
	DeleteTenantConfiguration(ctx context.Context, tenantID int32) error
	
	// Current tenant operations
	GetCurrentTenantConfiguration(ctx context.Context) (*TenantConfiguration, error)
	UpdateCurrentTenantMaxUsers(ctx context.Context, maxUsers int32) (*TenantConfiguration, error)
	UpdateCurrentTenantStorageQuota(ctx context.Context, storageQuota int64) (*TenantConfiguration, error)
	UpdateCurrentTenantFeatures(ctx context.Context, features json.RawMessage) (*TenantConfiguration, error)
	UpdateCurrentTenantModules(ctx context.Context, modules json.RawMessage) (*TenantConfiguration, error)
	CreateCurrentTenantConfiguration(ctx context.Context, arg CreateCurrentTenantConfigurationParams) (*TenantConfiguration, error)
	CheckCurrentTenantHasFeature(ctx context.Context, feature string) (bool, error)
	CheckCurrentTenantHasModule(ctx context.Context, module string) (bool, error)
}

// =====================================================
// REPOSITORY PARAMETER TYPES
// =====================================================

type CreateTenantParams struct {
	Name      string
	Subdomain *string
	Status    string
	Industry  *string
}

type UpdateTenantParams struct {
	ID        int32
	Name      string
	Subdomain *string
	Status    string
	Industry  *string
}

type UpdateCurrentTenantParams struct {
	Name      string
	Subdomain *string
	Status    string
	Industry  *string
}

type ListTenantsParams struct {
	Limit  int32
	Offset int32
}

type CreateTenantConfigurationParams struct {
	TenantID       int32
	MaxUsers       int32
	StorageQuota   int64
	Features       json.RawMessage
	ModulesEnabled json.RawMessage
}

type UpdateTenantConfigurationParams struct {
	TenantID       int32
	MaxUsers       int32
	StorageQuota   int64
	Features       json.RawMessage
	ModulesEnabled json.RawMessage
}

type CreateCurrentTenantConfigurationParams struct {
	MaxUsers       int32
	StorageQuota   int64
	Features       json.RawMessage
	ModulesEnabled json.RawMessage
}

