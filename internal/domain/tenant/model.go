package tenant

import (
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/repo/tenant"
)

type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"
	TenantStatusSuspended TenantStatus = "suspended"
	TenantStatusInactive  TenantStatus = "inactive"
)

type Tenant struct {
	ID        int32        `json:"id"`
	Uuid      uuid.UUID    `json:"uuid"`
	Name      string       `json:"name"`
	Subdomain string       `json:"subdomain"`
	Status    TenantStatus `json:"status"`
	Industry  string       `json:"industry"`
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
