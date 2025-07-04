package tenant

import (
	"time"

	"github.com/google/uuid"
)

type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"
	TenantStatusSuspended TenantStatus = "suspended"
	TenantStatusInactive  TenantStatus = "inactive"
	TenantStatusPending   TenantStatus = "pending"
)

func (s TenantStatus) IsValid() bool {
	switch s {
	case TenantStatusActive, TenantStatusSuspended, TenantStatusInactive, TenantStatusPending:
		return true
	default:
		return false
	}
}

// Domain types (matching your provided structs)
type Tenant struct {
	ID        int32        `json:"id"`
	Uuid      uuid.UUID    `json:"uuid"`
	Name      string       `json:"name"`
	Subdomain string       `json:"subdomain"`
	Status    TenantStatus `json:"status"`
	Industry  string       `json:"industry"`
	Settings  *[]byte      `json:"settings"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	DeletedAt time.Time    `json:"deleted_at"`
}

type TenantConfiguration struct {
	TenantID       int32  `json:"tenant_id"`
	MaxUsers       int32  `json:"max_users"`
	StorageQuota   int64  `json:"storage_quota"`
	Features       []byte `json:"features"`
	ModulesEnabled []byte `json:"modules_enabled"`
}

// Request/Response types
type CreateTenantRequest struct {
	Name      string       `json:"name" validate:"required,min=2,max=100"`
	Subdomain *string      `json:"subdomain" validate:"omitempty,min=3,max=50,alphanum"`
	Status    TenantStatus `json:"status" validate:"required,oneof=active suspended inactive"`
	Industry  *string      `json:"industry" validate:"omitempty,max=50"`
}

type UpdateTenantRequest struct {
	Name      *string       `json:"name" validate:"omitempty,min=2,max=100"`
	Subdomain *string       `json:"subdomain" validate:"omitempty,min=3,max=50,alphanum"`
	Status    *TenantStatus `json:"status" validate:"omitempty,oneof=active suspended inactive"`
	Industry  *string       `json:"industry" validate:"omitempty,max=50"`
}

type TenantFilter struct {
	NameFilter     *string `json:"name_filter"`
	StatusFilter   *string `json:"status_filter"`
	IndustryFilter *string `json:"industry_filter"`
	SortBy         string  `json:"sort_by" validate:"oneof=name created_at updated_at"`
	Limit          int32   `json:"limit" validate:"min=1,max=100"`
	Offset         int32   `json:"offset" validate:"min=0"`
}

type TenantStats struct {
	TotalTenants     int64 `json:"total_tenants"`
	ActiveTenants    int64 `json:"active_tenants"`
	SuspendedTenants int64 `json:"suspended_tenants"`
	PendingTenants   int64 `json:"pending_tenants"`
}

type BulkUpdateStatusRequest struct {
	Status    TenantStatus `json:"status" validate:"required,oneof=active suspended inactive"`
	TenantIDs []int32      `json:"tenant_ids" validate:"required,min=1"`
}

type SearchTenantsRequest struct {
	Name   *string `json:"name" validate:"omitempty,min=1"`
	Limit  int32   `json:"limit" validate:"min=1,max=100"`
	Offset int32   `json:"offset" validate:"min=0"`
}

type ListTenantsRequest struct {
	Limit  int32 `json:"limit" validate:"min=1,max=100"`
	Offset int32 `json:"offset" validate:"min=0"`
}
