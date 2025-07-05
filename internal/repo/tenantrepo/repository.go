package tenantrepo

// import (
// 	"context"
// 	"fmt"
//
// 	"github.com/google/uuid"
// 	"github.com/jackc/pgx/v5/pgtype"
// 	db "github.com/niiniyare/erp/db/sqlc"
// )
//
// // TenantRepository implements the TenantRepo interface using SQLC generated queries
// type TenantRepository struct {
// 	queries *db.Queries
// }
//
// // NewTenantRepository creates a new tenant repository instance
// func NewTenantRepository(queries *db.Queries) TenantRepo {
// 	return &TenantRepository{
// 		queries: queries,
// 	}
// }
//
// // CreateTenant creates a new tenant
// func (r *TenantRepository) CreateTenant(ctx context.Context, arg CreateTenant) (Tenant, error) {
// 	params := db.CreateTenantParams{
// 		Name:      arg.Name,
// 		Subdomain: arg.Subdomain,
// 		Status:    arg.Status,
// 		Industry:  arg.Industry,
// 	}
//
// 	dbTenant, err := r.queries.CreateTenant(ctx, params)
// 	if err != nil {
// 		return Tenant{}, fmt.Errorf("failed to create tenant: %w", err)
// 	}
//
// 	return r.convertFromDBTenant(dbTenant), nil
// }
//
// // BulkSoftDeleteTenants soft deletes multiple tenants
// func (r *TenantRepository) BulkSoftDeleteTenants(ctx context.Context, tenantIds []int32) error {
// 	return r.queries.BulkSoftDeleteTenants(ctx, tenantIds)
// }
//
// // BulkUpdateTenantStatus updates status for multiple tenants
// func (r *TenantRepository) BulkUpdateTenantStatus(ctx context.Context, arg BulkUpdateTenantStatus) error {
// 	params := db.BulkUpdateTenantStatusParams{
// 		Status: arg.Status,
// 		ID:     arg.ID,
// 	}
//
// 	return r.queries.BulkUpdateTenantStatus(ctx, params)
// }
//
// // CheckCurrentTenantExists checks if current tenant exists
// func (r *TenantRepository) CheckCurrentTenantExists(ctx context.Context) (bool, error) {
// 	return r.queries.CheckCurrentTenantExists(ctx)
// }
//
// // CheckSubdomainExists checks if a subdomain exists
// func (r *TenantRepository) CheckSubdomainExists(ctx context.Context, subdomain *string) (bool, error) {
// 	return r.queries.CheckSubdomainExists(ctx, subdomain)
// }
//
// // CheckTenantExists checks if a tenant exists
// func (r *TenantRepository) CheckTenantExists(ctx context.Context, id int32) (bool, error) {
// 	return r.queries.CheckTenantExists(ctx, id)
// }
//
// // CheckTenantNameExists checks if a tenant name exists
// func (r *TenantRepository) CheckTenantNameExists(ctx context.Context, name string) (bool, error) {
// 	return r.queries.CheckTenantNameExists(ctx, name)
// }
//
// // CountFilteredTenants counts tenants based on filters
// func (r *TenantRepository) CountFilteredTenants(ctx context.Context, arg CountFilteredTenants) (int64, error) {
// 	params := db.CountFilteredTenantsParams{
// 		NameFilter:     arg.NameFilter,
// 		StatusFilter:   arg.StatusFilter,
// 		IndustryFilter: arg.IndustryFilter,
// 	}
//
// 	return r.queries.CountFilteredTenants(ctx, params)
// }
//
// // CountTenants returns total count of tenants
// func (r *TenantRepository) CountTenants(ctx context.Context) (int64, error) {
// 	return r.queries.CountTenants(ctx)
// }
//
// // DeleteTenant deletes current tenant (hard delete)
// func (r *TenantRepository) DeleteTenant(ctx context.Context) error {
// 	return r.queries.DeleteTenant(ctx)
// }
//
// // FilterTenants retrieves tenants based on filters
// func (r *TenantRepository) FilterTenants(ctx context.Context, arg FilterTenants) ([]Tenant, error) {
// 	sortBy := ""
// 	if arg.SortBy != nil {
// 		sortBy = string(*arg.SortBy)
// 	}
//
// 	params := db.FilterTenantsParams{
// 		NameFilter:     arg.NameFilter,
// 		StatusFilter:   arg.StatusFilter,
// 		IndustryFilter: arg.IndustryFilter,
// 		SortBy:         sortBy,
// 		OffsetCount:    arg.OffsetCount,
// 		LimitCount:     arg.LimitCount,
// 	}
//
// 	dbTenants, err := r.queries.FilterTenants(ctx, params)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to filter tenants: %w", err)
// 	}
//
// 	return r.convertFromDBTenants(dbTenants), nil
// }
//
// // GetActiveTenants retrieves all active tenants
// func (r *TenantRepository) GetActiveTenants(ctx context.Context) ([]Tenant, error) {
// 	dbTenants, err := r.queries.GetActiveTenants(ctx)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get active tenants: %w", err)
// 	}
//
// 	return r.convertFromDBTenants(dbTenants), nil
// }
//
// // GetCurrentTenant retrieves the current tenant
// func (r *TenantRepository) GetCurrentTenant(ctx context.Context) (Tenant, error) {
// 	dbTenant, err := r.queries.GetCurrentTenant(ctx)
// 	if err != nil {
// 		return Tenant{}, fmt.Errorf("failed to get current tenant: %w", err)
// 	}
//
// 	return r.convertFromDBTenant(dbTenant), nil
// }
//
// // GetCurrentTenantID retrieves the current tenant ID
// func (r *TenantRepository) GetCurrentTenantID(ctx context.Context) (int32, error) {
// 	return r.queries.GetCurrentTenantID(ctx)
// }
//
// // GetTenantByID retrieves a tenant by ID
// func (r *TenantRepository) GetTenantByID(ctx context.Context, id int32) (Tenant, error) {
// 	dbTenant, err := r.queries.GetTenantByID(ctx, id)
// 	if err != nil {
// 		return Tenant{}, fmt.Errorf("failed to get tenant by ID: %w", err)
// 	}
//
// 	return r.convertFromDBTenant(dbTenant), nil
// }
//
// // GetTenantBySubdomain retrieves a tenant by subdomain
// func (r *TenantRepository) GetTenantBySubdomain(ctx context.Context, subdomain *string) (Tenant, error) {
// 	dbTenant, err := r.queries.GetTenantBySubdomain(ctx, subdomain)
// 	if err != nil {
// 		return Tenant{}, fmt.Errorf("failed to get tenant by subdomain: %w", err)
// 	}
//
// 	return r.convertFromDBTenant(dbTenant), nil
// }
//
// // GetTenantByUUID retrieves a tenant by UUID
// func (r *TenantRepository) GetTenantByUUID(ctx context.Context, argUuid uuid.UUID) (Tenant, error) {
// 	dbTenant, err := r.queries.GetTenantByUUID(ctx, argUuid)
// 	if err != nil {
// 		return Tenant{}, fmt.Errorf("failed to get tenant by UUID: %w", err)
// 	}
//
// 	return r.convertFromDBTenant(dbTenant), nil
// }
//
// // GetTenantsByIndustry retrieves tenants grouped by industry
// func (r *TenantRepository) GetTenantsByIndustry(ctx context.Context) ([]GetTenantsByIndustryRow, error) {
// 	dbRows, err := r.queries.GetTenantsByIndustry(ctx)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get tenants by industry: %w", err)
// 	}
//
// 	rows := make([]GetTenantsByIndustryRow, len(dbRows))
// 	for i, dbRow := range dbRows {
// 		rows[i] = GetTenantsByIndustryRow{
// 			Industry:    dbRow.Industry,
// 			TenantCount: dbRow.TenantCount,
// 		}
// 	}
//
// 	return rows, nil
// }
//
// // ListTenants retrieves tenants with pagination
// func (r *TenantRepository) ListTenants(ctx context.Context, limit, offset int32) ([]Tenant, error) {
// 	params := db.ListTenantsParams{
// 		Limit:  limit,
// 		Offset: offset,
// 	}
//
// 	dbTenants, err := r.queries.ListTenants(ctx, params)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to list tenants: %w", err)
// 	}
//
// 	return r.convertFromDBTenants(dbTenants), nil
// }
//
// // SearchTenantsByName searches tenants by name
// func (r *TenantRepository) SearchTenantsByName(ctx context.Context, arg SearchTenantsByName) ([]Tenant, error) {
// 	params := db.SearchTenantsByNameParams{
// 		Name:   arg.Name,
// 		Offset: arg.Offset,
// 		Limit:  arg.Limit,
// 	}
//
// 	dbTenants, err := r.queries.SearchTenantsByName(ctx, params)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to search tenants by name: %w", err)
// 	}
//
// 	return r.convertFromDBTenants(dbTenants), nil
// }
//
// // SetCurrentTenant sets the current tenant
// func (r *TenantRepository) SetCurrentTenant(ctx context.Context, tenantID interface{}) error {
// 	return r.queries.SetCurrentTenant(ctx, tenantID)
// }
//
// // SoftDeleteTenant soft deletes a tenant
// func (r *TenantRepository) SoftDeleteTenant(ctx context.Context, id int32) error {
// 	return r.queries.SoftDeleteTenant(ctx, id)
// }
//
// // UpdateTenant updates a tenant
// func (r *TenantRepository) UpdateTenant(ctx context.Context, arg UpdateTenant) (Tenant, error) {
// 	params := db.UpdateTenantParams{
// 		Name:      arg.Name,
// 		Subdomain: arg.Subdomain,
// 		Status:    arg.Status,
// 		Industry:  arg.Industry,
// 		ID:        arg.ID,
// 	}
//
// 	dbTenant, err := r.queries.UpdateTenant(ctx, params)
// 	if err != nil {
// 		return Tenant{}, fmt.Errorf("failed to update tenant: %w", err)
// 	}
//
// 	return r.convertFromDBTenant(dbTenant), nil
// }
//
// // Additional methods that exist in SQLC but not in the interface
// // These can be used for extended functionality
//
// // UpdateCurrentTenant updates the current tenant
// func (r *TenantRepository) UpdateCurrentTenant(ctx context.Context, arg UpdateCurrentTenant) (Tenant, error) {
// 	params := db.UpdateCurrentTenantParams{
// 		Name:      arg.Name,
// 		Subdomain: arg.Subdomain,
// 		Status:    arg.Status,
// 		Industry:  arg.Industry,
// 	}
//
// 	dbTenant, err := r.queries.UpdateCurrentTenant(ctx, params)
// 	if err != nil {
// 		return Tenant{}, fmt.Errorf("failed to update current tenant: %w", err)
// 	}
//
// 	return r.convertFromDBTenant(dbTenant), nil
// }
//
// // GetTenantStats retrieves tenant statistics
// func (r *TenantRepository) GetTenantStats(ctx context.Context) (db.GetTenantStatsRow, error) {
// 	return r.queries.GetTenantStats(ctx)
// }
//
// // UpdateTenantName updates only the tenant name
// func (r *TenantRepository) UpdateTenantName(ctx context.Context, id int32, name string) (Tenant, error) {
// 	params := db.UpdateTenantNameParams{
// 		ID:   id,
// 		Name: name,
// 	}
//
// 	dbTenant, err := r.queries.UpdateTenantName(ctx, params)
// 	if err != nil {
// 		return Tenant{}, fmt.Errorf("failed to update tenant name: %w", err)
// 	}
//
// 	return r.convertFromDBTenant(dbTenant), nil
// }
//
// // UpdateTenantStatus updates only the tenant status
// func (r *TenantRepository) UpdateTenantStatus(ctx context.Context, id int32, status string) (Tenant, error) {
// 	params := db.UpdateTenantStatusParams{
// 		ID:     id,
// 		Status: status,
// 	}
//
// 	dbTenant, err := r.queries.UpdateTenantStatus(ctx, params)
// 	if err != nil {
// 		return Tenant{}, fmt.Errorf("failed to update tenant status: %w", err)
// 	}
//
// 	return r.convertFromDBTenant(dbTenant), nil
// }
//
// // UpdateTenantSubdomain updates only the tenant subdomain
// func (r *TenantRepository) UpdateTenantSubdomain(ctx context.Context, id int32, subdomain *string) (Tenant, error) {
// 	params := db.UpdateTenantSubdomainParams{
// 		ID:        id,
// 		Subdomain: subdomain,
// 	}
//
// 	dbTenant, err := r.queries.UpdateTenantSubdomain(ctx, params)
// 	if err != nil {
// 		return Tenant{}, fmt.Errorf("failed to update tenant subdomain: %w", err)
// 	}
//
// 	return r.convertFromDBTenant(dbTenant), nil
// }
//
// // UpdateTenantIndustry updates only the tenant industry
// func (r *TenantRepository) UpdateTenantIndustry(ctx context.Context, id int32, industry *string) (Tenant, error) {
// 	params := db.UpdateTenantIndustryParams{
// 		ID:       id,
// 		Industry: industry,
// 	}
//
// 	dbTenant, err := r.queries.UpdateTenantIndustry(ctx, params)
// 	if err != nil {
// 		return Tenant{}, fmt.Errorf("failed to update tenant industry: %w", err)
// 	}
//
// 	return r.convertFromDBTenant(dbTenant), nil
// }
//
// // GetCurrentTenantStorageUsage retrieves current tenant storage usage
// func (r *TenantRepository) GetCurrentTenantStorageUsage(ctx context.Context) (db.GetCurrentTenantStorageUsageRow, error) {
// 	return r.queries.GetCurrentTenantStorageUsage(ctx)
// }
//
// // GetTenantsCreatedInDateRange retrieves tenants created in a date range
// func (r *TenantRepository) GetTenantsCreatedInDateRange(ctx context.Context, arg db.GetTenantsCreatedInDateRangeParams) ([]Tenant, error) {
// 	dbTenants, err := r.queries.GetTenantsCreatedInDateRange(ctx, arg)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get tenants created in date range: %w", err)
// 	}
//
// 	return r.convertFromDBTenants(dbTenants), nil
// }
//
// // Private helper methods for type conversion
//
// // convertFromDBTenant converts a SQLC db.Tenant to repository Tenant
// func (r *TenantRepository) convertFromDBTenant(dbTenant db.Tenant) Tenant {
// 	return Tenant{
// 		ID:        dbTenant.ID,
// 		Uuid:      dbTenant.Uuid,
// 		Name:      dbTenant.Name,
// 		Subdomain: *dbTenant.Subdomain,
// 		Status:    dbTenant.Status,
// 		Industry:  *dbTenant.Industry,
// 		CreatedAt: dbTenant.CreatedAt.Time,
// 		UpdatedAt: dbTenant.UpdatedAt.Time,
// 		DeletedAt: dbTenant.DeletedAt.Time,
// 	}
// }
//
// // convertFromDBTenants converts a slice of SQLC db.Tenant to repository Tenant slice
// func (r *TenantRepository) convertFromDBTenants(dbTenants []db.Tenant) []Tenant {
// 	tenants := make([]Tenant, len(dbTenants))
// 	for i, dbTenant := range dbTenants {
// 		tenants[i] = r.convertFromDBTenant(dbTenant)
// 	}
// 	return tenants
// }
//
// // convertToDBTenant converts a repository Tenant to SQLC db.Tenant (if needed)
// func (r *TenantRepository) convertToDBTenant(tenant Tenant) db.Tenant {
// 	return db.Tenant{
// 		ID:        tenant.ID,
// 		Uuid:      tenant.Uuid,
// 		Name:      tenant.Name,
// 		Subdomain: &tenant.Subdomain,
// 		Status:    tenant.Status,
// 		Industry:  &tenant.Industry,
// 		CreatedAt: pgtype.Timestamptz{
// 			Time:  tenant.CreatedAt,
// 			Valid: true},
// 		UpdatedAt: pgtype.Timestamptz{
// 			Time:  tenant.UpdatedAt,
// 			Valid: true},
// 		DeletedAt: pgtype.Timestamptz{
// 			Time:  tenant.DeletedAt,
// 			Valid: true},
// 	}
// }
