package tenancy

//
// import (
// 	"context"
// 	"time"
//
// 	"github.com/google/uuid"
//
// 	db "github.com/niiniyare/erp/db/sqlc"
// )
//
// // Repository defines the interface for tenant data operations
// // It wraps the SQLC generated queries with domain-specific methods
// type Repository interface {
// 	// Tenant CRUD operations
// 	CreateTenant(ctx context.Context, req CreateTenantRequest) (*Tenant, error)
// 	GetTenant(ctx context.Context, id uuid.UUID) (*Tenant, error)
// 	GetTenantBySlug(ctx context.Context, slug string) (*Tenant, error)
// 	GetTenantBySubdomain(ctx context.Context, subdomain string) (*Tenant, error)
// 	GetCurrentTenant(ctx context.Context) (*Tenant, error)
// 	UpdateTenant(ctx context.Context, id uuid.UUID, req UpdateTenantRequest) (*Tenant, error)
// 	DeleteTenant(ctx context.Context, id uuid.UUID) error
//
// 	// Tenant listing and filtering
// 	ListTenants(ctx context.Context, filter TenantFilter) ([]*Tenant, error)
// 	CountTenants(ctx context.Context) (int64, error)
// 	SearchTenantsByName(ctx context.Context, query string, limit, offset int) ([]*Tenant, error)
//
// 	// Tenant existence checks
// 	CheckTenantExists(ctx context.Context, id uuid.UUID) (bool, error)
// 	CheckSubdomainExists(ctx context.Context, subdomain string) (bool, error)
// 	CheckTenantNameExists(ctx context.Context, name string) (bool, error)
//
// 	// Bulk operations
// 	BulkUpdateTenantStatus(ctx context.Context, tenantIDs []uuid.UUID, status string) error
// 	BulkSoftDeleteTenants(ctx context.Context, tenantIDs []uuid.UUID) error
//
// 	// Tenant configuration operations
// 	CreateTenantConfiguration(ctx context.Context, tenantID uuid.UUID) (*Configuration, error)
// 	GetTenantConfiguration(ctx context.Context, tenantID uuid.UUID) (*Configuration, error)
// 	UpdateTenantConfiguration(ctx context.Context, config *Configuration) (*Configuration, error)
//
// 	// Usage statistics operations
// 	CreateTenantUsageStats(ctx context.Context, stats *UsageStats) (*UsageStats, error)
// 	GetTenantUsageStats(ctx context.Context, tenantID uuid.UUID, periodStart time.Time) (*UsageStats, error)
// 	GetTenantUsageStatsRange(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]*UsageStats, error)
//
// 	// Analytics and reporting
// 	GetTenantGrowthStats(ctx context.Context, days int) ([]*GrowthStat, error)
// 	GetTenantsByIndustry(ctx context.Context) ([]*IndustryStat, error)
// 	GetTenantsByCompanySize(ctx context.Context) ([]*CompanySizeStat, error)
// 	GetTenantStatusDistribution(ctx context.Context) (*StatusDistribution, error)
//
// 	// Tenant context management
// 	SetTenantContext(ctx context.Context, tenantID uuid.UUID) error
// 	GetCurrentTenantID(ctx context.Context) (uuid.UUID, error)
// }
//
// // Analytics types for repository responses
// type GrowthStat struct {
// 	Date         time.Time `json:"date"`
// 	NewTenants   int64     `json:"new_tenants"`
// 	TotalTenants int64     `json:"total_tenants"`
// }
//
// type IndustryStat struct {
// 	Industry string `json:"industry"`
// 	Count    int64  `json:"count"`
// }
//
// type CompanySizeStat struct {
// 	CompanySize string `json:"company_size"`
// 	Count       int64  `json:"count"`
// }
//
// type StatusDistribution struct {
// 	Active    int64 `json:"active"`
// 	Suspended int64 `json:"suspended"`
// 	Pending   int64 `json:"pending"`
// }
//
// // sqlcRepository implements Repository using SQLC generated queries
// type sqlcRepository struct {
// 	queries *db.Queries
// }
//
// // NewRepository creates a new tenant repository using SQLC queries
// func NewRepository(queries *db.Queries) Repository {
// 	return &sqlcRepository{
// 		queries: queries,
// 	}
// }
//
// // Domain model conversion methods
//
// // tenantFromSQLC converts SQLC Tenant to domain Tenant
// func tenantFromSQLC(t *db.Tenant) *Tenant {
// 	if t == nil {
// 		return nil
// 	}
// 	return &Tenant{
// 		ID:                 t.ID,
// 		Slug:               t.Slug,
// 		Name:               t.Name,
// 		Email:              t.Email,
// 		Subdomain:          t.Subdomain,
// 		Status:             t.Status,
// 		Timezone:           t.Timezone,
// 		CurrencyCode:       t.CurrencyCode,
// 		Metadata:           t.Metadata,
// 		Industry:           t.Industry,
// 		CompanySize:        t.CompanySize,
// 		TaxID:              t.TaxID,
// 		RegistrationNumber: t.RegistrationNumber,
// 		LegalEntityType:    t.LegalEntityType,
// 		Settings:           t.Settings,
// 		CreatedAt:          t.CreatedAt,
// 		UpdatedAt:          t.UpdatedAt,
// 		DeletedAt:          t.DeletedAt,
// 	}
// }
//
// // configurationFromSQLC converts SQLC TenantConfiguration to domain Configuration
// func configurationFromSQLC(c *db.TenantConfiguration) *Configuration {
// 	if c == nil {
// 		return nil
// 	}
// 	return &Configuration{
// 		TenantID:                c.TenantID,
// 		MaxUsers:                c.MaxUsers,
// 		MaxEntities:             c.MaxEntities,
// 		MaxTransactionsPerMonth: c.MaxTransactionsPerMonth,
// 		StorageQuota:            c.StorageQuota,
// 		Features:                c.Features,
// 		ModulesEnabled:          c.ModulesEnabled,
// 		AccountingMethod:        c.AccountingMethod,
// 		FiscalYearStartMonth:    c.FiscalYearStartMonth,
// 		DefaultCurrency:         c.DefaultCurrency,
// 		DateFormat:              c.DateFormat,
// 		NumberFormat:            c.NumberFormat,
// 		LanguageCode:            c.LanguageCode,
// 		PasswordPolicy:          c.PasswordPolicy,
// 		WebhookEndpoints:        c.WebhookEndpoints,
// 		ApiRateLimits:           c.ApiRateLimits,
// 		CreatedAt:               c.CreatedAt,
// 		UpdatedAt:               c.UpdatedAt,
// 	}
// }
//
// // usageStatsFromSQLC converts SQLC TenantUsageStat to domain UsageStats
// func usageStatsFromSQLC(u *db.TenantUsageStat) *UsageStats {
// 	if u == nil {
// 		return nil
// 	}
//
// 	// Convert pgtype.Numeric to float64
// 	avgResponseTime, _ := u.AvgResponseTime.Float64Value()
// 	errorRate, _ := u.ErrorRate.Float64Value()
// 	monthlyRevenue, _ := u.MonthlyRevenue.Float64Value()
//
// 	return &UsageStats{
// 		TenantID:          u.TenantID,
// 		PeriodStart:       u.PeriodStart,
// 		PeriodEnd:         u.PeriodEnd,
// 		ActiveUsers:       u.ActiveUsers,
// 		TotalEntities:     u.TotalEntities,
// 		TotalTransactions: u.TotalTransactions,
// 		StorageUsed:       u.StorageUsed,
// 		ApiCalls:          u.ApiCalls,
// 		AvgResponseTime:   avgResponseTime.Float64,
// 		ErrorRate:         errorRate.Float64,
// 		MonthlyRevenue:    monthlyRevenue.Float64,
// 		CreatedAt:         u.CreatedAt,
// 	}
// }
//
