package tenant

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

type TenantFilter struct {
	NameFilter     *string `json:"name_filter"`
	StatusFilter   *string `json:"status_filter"`
	IndustryFilter *string `json:"industry_filter"`
	SortBy         string  `json:"sort_by" validate:"oneof=name created_at updated_at"`
	Limit          int32   `json:"limit" validate:"min=1,max=100"`
	Offset         int32   `json:"offset" validate:"min=0"`
}

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

type TenantStats struct {
	TotalTenants     int64 `json:"total_tenants"`
	ActiveTenants    int64 `json:"active_tenants"`
	SuspendedTenants int64 `json:"suspended_tenants"`
	PendingTenants   int64 `json:"pending_tenants"`
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
