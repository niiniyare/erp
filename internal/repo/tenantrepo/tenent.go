package tenantrepo

import (
	"context"

	"github.com/google/uuid"
)

type TenantRepo interface {
	CreateTenant(ctx context.Context, arg CreateTenant) (Tenant, error)
	BulkSoftDeleteTenants(ctx context.Context, tenantIds []int32) error
	// =====================================================
	// BULK OPERATIONS
	// =====================================================
	BulkUpdateTenantStatus(ctx context.Context, arg BulkUpdateTenantStatus) error
	CheckCurrentTenantExists(ctx context.Context) (bool, error)
	CheckSubdomainExists(ctx context.Context, subdomain *string) (bool, error)
	CheckTenantExists(ctx context.Context, id int32) (bool, error)
	CheckTenantNameExists(ctx context.Context, name string) (bool, error)
	CountFilteredTenants(ctx context.Context, arg CountFilteredTenants) (int64, error)
	CountTenants(ctx context.Context) (int64, error)
	DeleteTenant(ctx context.Context) error
	FilterTenants(ctx context.Context, arg FilterTenants) ([]Tenant, error)
	GetActiveTenants(ctx context.Context) ([]Tenant, error)
	GetCurrentTenant(ctx context.Context) (Tenant, error)
	GetCurrentTenantID(ctx context.Context) (int32, error)
	// GetCurrentTenantStorageUsage(ctx context.Context) (GetCurrentTenantStorageUsageRow, error)
	GetTenantByID(ctx context.Context, id int32) (Tenant, error)
	GetTenantBySubdomain(ctx context.Context, subdomain *string) (Tenant, error)
	GetTenantByUUID(ctx context.Context, argUuid uuid.UUID) (Tenant, error)
	// Admin utilities (system-wide)
	// GetTenantStats(ctx context.Context) (GetTenantStatsRow, error)
	GetTenantsByIndustry(ctx context.Context) ([]GetTenantsByIndustryRow, error)
	// GetTenantsCreatedInDateRange(ctx context.Context, arg GetTenantsCreatedInDateRange) ([]Tenant, error)
	ListTenants(ctx context.Context, limit, offset int32) ([]Tenant, error)
	SearchTenantsByName(ctx context.Context, arg SearchTenantsByName) ([]Tenant, error)
	SetCurrentTenant(ctx context.Context, dollar_1 interface{}) error
	SoftDeleteTenant(ctx context.Context, id int32) error
	// UpdateCurrentTenant(ctx context.Context, arg UpdateCurrentTenant) (Tenant, error)
	UpdateTenant(ctx context.Context, arg UpdateTenant) (Tenant, error)
	// UpdateTenantIndustry(ctx context.Context, arg UpdateTenantIndustry) (Tenant, error)
	// UpdateTenantName(ctx context.Context, arg UpdateTenantName) (Tenant, error)
	//
	// UpdateTenantStatus(ctx context.Context, arg UpdateTenantStatus) (Tenant, error)
	// UpdateTenantSubdomain(ctx context.Context, arg UpdateTenantSubdomain) (Tenant, error)
}
type UpdateCurrentTenant struct {
	Name      string  `json:"name"`
	Subdomain *string `json:"subdomain"`
	Status    string  `json:"status"`
	Industry  *string `json:"industry"`
}

type CreateTenant struct {
	Name      string  `json:"name"`
	Subdomain *string `json:"subdomain"`
	Status    string  `json:"status"`
	Industry  *string `json:"industry"`
}

type BulkUpdateTenantStatus struct {
	Status string  `json:"status"`
	ID     []int32 `json:"id"`
}
type TenantSortField string

const (
	SortByName      TenantSortField = "name"
	SortByCreatedAt TenantSortField = "created_at"
	SortByUpdatedAt TenantSortField = "updated_at"
)

type FilterTenants struct {
	NameFilter     *string          `json:"name_filter"`
	StatusFilter   *string          `json:"status_filter"`
	IndustryFilter *string          `json:"industry_filter"`
	SortBy         *TenantSortField `json:"sort_by"` // Pointer to enum
	OffsetCount    int32            `json:"offset_count"`
	LimitCount     int32            `json:"limit_count"`
}

type CountFilteredTenants struct {
	NameFilter     *string `json:"name_filter"`
	StatusFilter   *string `json:"status_filter"`
	IndustryFilter *string `json:"industry_filter"`
}

type GetTenantsByIndustryRow struct {
	Industry    *string `json:"industry"`
	TenantCount int64   `json:"tenant_count"`
}

type List struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

type SearchTenantsByName struct {
	Name   *string `json:"name"`
	Offset int32   `json:"offset"`
	Limit  int32   `json:"limit"`
}

type UpdateTenant struct {
	Name      *string `json:"name"`
	Subdomain *string `json:"subdomain"`
	Status    *string `json:"status"`
	Industry  *string `json:"industry"`
	ID        int32   `json:"id"`
}
