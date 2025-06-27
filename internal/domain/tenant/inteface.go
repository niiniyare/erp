package tenant

import ()

// // TenantRepository defines the interface for tenant data operations
// type TenantRepository interface {
// 	// Basic CRUD operations
// 	Create(ctx context.Context, req *CreateTenantRequest) (*Tenant, error)
// 	GetByID(ctx context.Context, id int32) (*Tenant, error)
// 	GetByUUID(ctx context.Context, uuid uuid.UUID) (*Tenant, error)
// 	GetBySubdomain(ctx context.Context, subdomain string) (*Tenant, error)
// 	Update(ctx context.Context, id int32, req *UpdateTenantRequest) (*Tenant, error)
// 	Delete(ctx context.Context, id int32) error
//
// 	// List and search operations
// 	List(ctx context.Context, limit, offset int32) ([]*Tenant, error)
// 	GetActiveTenants(ctx context.Context) ([]*Tenant, error)
// 	SearchByName(ctx context.Context, name string, limit, offset int32) ([]*Tenant, error)
// 	Filter(ctx context.Context, filter *TenantFilter) ([]*Tenant, error)
//
// 	// Status operations
// 	UpdateStatus(ctx context.Context, id int32, status TenantStatus) (*Tenant, error)
// 	BulkUpdateStatus(ctx context.Context, ids []int32, status TenantStatus) error
//
// 	// Validation operations
// 	CheckExists(ctx context.Context, id int32) (bool, error)
// 	CheckSubdomainExists(ctx context.Context, subdomain string) (bool, error)
// 	CheckNameExists(ctx context.Context, name string) (bool, error)
//
// 	// Statistics and reporting
// 	Count(ctx context.Context) (int64, error)
// 	GetStats(ctx context.Context) (*TenantStats, error)
// }
//
// // TenantService defines the interface for tenant business logic
// type TenantService interface {
// 	// Primary business operations
// 	CreateTenant(ctx context.Context, req *CreateTenantRequest) (*Tenant, error)
// 	GetTenant(ctx context.Context, id int32) (*Tenant, error)
// 	GetTenantByUUID(ctx context.Context, uuid uuid.UUID) (*Tenant, error)
// 	GetTenantBySubdomain(ctx context.Context, subdomain string) (*Tenant, error)
// 	UpdateTenant(ctx context.Context, id int32, req *UpdateTenantRequest) (*Tenant, error)
// 	DeleteTenant(ctx context.Context, id int32) error
//
// 	// List and search operations
// 	ListTenants(ctx context.Context, limit, offset int32) ([]*Tenant, error)
// 	SearchTenants(ctx context.Context, name string, limit, offset int32) ([]*Tenant, error)
// 	FilterTenants(ctx context.Context, filter *TenantFilter) ([]*Tenant, error)
//
// 	// Status management
// 	ActivateTenant(ctx context.Context, id int32) (*Tenant, error)
// 	SuspendTenant(ctx context.Context, id int32) (*Tenant, error)
// 	DeactivateTenant(ctx context.Context, id int32) (*Tenant, error)
//
// 	// Validation operations
// 	ValidateSubdomain(ctx context.Context, subdomain string) error
// 	ValidateTenantName(ctx context.Context, name string) error
//
// 	// Statistics and reporting
// 	GetTenantStats(ctx context.Context) (*TenantStats, error)
// }
//
// // Primary Ports (Driving - Inbound)
