package tenant

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	tenantrepo "github.com/niiniyare/erp/internal/repo/tenant_repo"

	// Import the tenant package from its location
	"github.com/niiniyare/erp/internal/storage/cache"
)

// TenantStatus represents the status of a tenant
type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"
	TenantStatusSuspended TenantStatus = "suspended"
	TenantStatusInactive  TenantStatus = "inactive"
)

// Domain types (matching your provided structs)
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

// TenantService provides business logic for tenant operations
type TenantService struct {
	dbManager tenantrepo.TenantRepo
	cache     cache.CacheStore
}

// NewTenantService creates a new tenant service instance
func NewTenantService(dbManager tenantrepo.TenantRepo, cache cache.CacheStore) *TenantService {
	return &TenantService{
		dbManager: dbManager,
		cache:     cache,
	}
}

// Cache key constants
const (
	tenantByIDKey        = "tenant:id:%d"
	tenantByUUIDKey      = "tenant:uuid:%s"
	tenantBySubdomainKey = "tenant:subdomain:%s"
	currentTenantKey     = "tenant:current"
	tenantExistsKey      = "tenant:exists:%d"
	subdomainExistsKey   = "subdomain:exists:%s"
	tenantCountKey       = "tenant:count"
	activTenantsKey      = "tenants:active"
	tenantStatsKey       = "tenant:stats"

	// Cache TTL durations
	defaultCacheTTL = 15 * time.Minute
	shortCacheTTL   = 5 * time.Minute
	longCacheTTL    = 30 * time.Minute
)

// CreateTenant creates a new tenant and invalidates relevant caches
func (s *TenantService) CreateTenant(ctx context.Context, req CreateTenantRequest) (Tenant, error) {
	// Validate business rules
	if err := s.validateTenantCreation(ctx, req); err != nil {
		return Tenant{}, err
	}

	// Convert request to repository type
	createArg := tenantrepo.CreateTenant{
		Name:      req.Name,
		Subdomain: req.Subdomain,
		Status:    string(req.Status),
		Industry:  req.Industry,
	}

	// Create tenant
	repoTenant, err := s.dbManager.CreateTenant(ctx, createArg)
	if err != nil {
		return Tenant{}, fmt.Errorf("failed to create tenant: %w", err)
	}

	// Convert to domain type
	domainTenant := s.convertToDomainTenant(repoTenant)

	// Invalidate relevant caches
	s.invalidateCountCaches()
	s.invalidateListCaches()
	s.invalidateStatsCaches()

	// Cache the new tenant
	s.cacheDomainTenant(domainTenant)

	return domainTenant, nil
}

// GetTenantByID retrieves a tenant by ID with caching
func (s *TenantService) GetTenantByID(ctx context.Context, id int32) (Tenant, error) {
	cacheKey := fmt.Sprintf(tenantByIDKey, id)

	// Try cache first
	if cached, err := s.cache.Get(cacheKey); err == nil {
		var t Tenant
		if json.Unmarshal([]byte(cached), &t) == nil {
			return t, nil
		}
	}

	// Fetch from database
	repoTenant, err := s.dbManager.GetTenantByID(ctx, id)
	if err != nil {
		return Tenant{}, fmt.Errorf("failed to get tenant by ID: %w", err)
	}

	// Convert to domain type
	domainTenant := s.convertToDomainTenant(repoTenant)

	// Cache the result
	s.cacheDomainTenant(domainTenant)

	return domainTenant, nil
}

// GetTenantByUUID retrieves a tenant by UUID with caching
func (s *TenantService) GetTenantByUUID(ctx context.Context, tenantUUID uuid.UUID) (Tenant, error) {
	cacheKey := fmt.Sprintf(tenantByUUIDKey, tenantUUID.String())

	// Try cache first
	if cached, err := s.cache.Get(cacheKey); err == nil {
		var t Tenant
		if json.Unmarshal([]byte(cached), &t) == nil {
			return t, nil
		}
	}

	// Fetch from database
	repoTenant, err := s.dbManager.GetTenantByUUID(ctx, tenantUUID)
	if err != nil {
		return Tenant{}, fmt.Errorf("failed to get tenant by UUID: %w", err)
	}

	// Convert to domain type
	domainTenant := s.convertToDomainTenant(repoTenant)

	// Cache the result
	s.cacheDomainTenant(domainTenant)

	return domainTenant, nil
}

// GetTenantBySubdomain retrieves a tenant by subdomain with caching
func (s *TenantService) GetTenantBySubdomain(ctx context.Context, subdomain string) (Tenant, error) {
	if subdomain == "" {
		return Tenant{}, fmt.Errorf("subdomain cannot be empty")
	}

	cacheKey := fmt.Sprintf(tenantBySubdomainKey, subdomain)

	// Try cache first
	if cached, err := s.cache.Get(cacheKey); err == nil {
		var t Tenant
		if json.Unmarshal([]byte(cached), &t) == nil {
			return t, nil
		}
	}

	// Fetch from database
	subdomainPtr := &subdomain
	repoTenant, err := s.dbManager.GetTenantBySubdomain(ctx, subdomainPtr)
	if err != nil {
		return Tenant{}, fmt.Errorf("failed to get tenant by subdomain: %w", err)
	}

	// Convert to domain type
	domainTenant := s.convertToDomainTenant(repoTenant)

	// Cache the result
	s.cacheDomainTenant(domainTenant)

	return domainTenant, nil
}

// GetCurrentTenant retrieves the current tenant with caching
func (s *TenantService) GetCurrentTenant(ctx context.Context) (Tenant, error) {
	// Try cache first
	if cached, err := s.cache.Get(currentTenantKey); err == nil {
		var t Tenant
		if json.Unmarshal([]byte(cached), &t) == nil {
			return t, nil
		}
	}

	// Fetch from database
	repoTenant, err := s.dbManager.GetCurrentTenant(ctx)
	if err != nil {
		return Tenant{}, fmt.Errorf("failed to get current tenant: %w", err)
	}

	// Convert to domain type
	domainTenant := s.convertToDomainTenant(repoTenant)

	// Cache the result
	if data, err := json.Marshal(domainTenant); err == nil {
		s.cache.Set(currentTenantKey, string(data), defaultCacheTTL)
	}

	return domainTenant, nil
}

// UpdateTenant updates a tenant and invalidates caches
func (s *TenantService) UpdateTenant(ctx context.Context, id int32, req UpdateTenantRequest) (Tenant, error) {
	// Validate business rules
	if err := s.validateTenantUpdate(ctx, id, req); err != nil {
		return Tenant{}, err
	}

	// Convert request to repository type
	updateArg := tenantrepo.UpdateTenant{
		ID:        id,
		Name:      req.Name,
		Subdomain: req.Subdomain,
		Industry:  req.Industry,
	}

	if req.Status != nil {
		status := string(*req.Status)
		updateArg.Status = &status
	}

	// Update tenant
	repoTenant, err := s.dbManager.UpdateTenant(ctx, updateArg)
	if err != nil {
		return Tenant{}, fmt.Errorf("failed to update tenant: %w", err)
	}

	// Convert to domain type
	domainTenant := s.convertToDomainTenant(repoTenant)

	// Invalidate caches for this tenant
	s.invalidateTenantCaches(id)
	s.invalidateListCaches()
	s.invalidateStatsCaches()

	// Cache the updated tenant
	s.cacheDomainTenant(domainTenant)

	return domainTenant, nil
}

// DeleteTenant soft deletes a tenant and invalidates caches
func (s *TenantService) DeleteTenant(ctx context.Context, id int32) error {
	if err := s.dbManager.SoftDeleteTenant(ctx, id); err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	// Invalidate caches
	s.invalidateTenantCaches(id)
	s.invalidateCountCaches()
	s.invalidateListCaches()
	s.invalidateStatsCaches()

	return nil
}

// BulkDeleteTenants soft deletes multiple tenants
func (s *TenantService) BulkDeleteTenants(ctx context.Context, tenantIds []int32) error {
	if err := s.dbManager.BulkSoftDeleteTenants(ctx, tenantIds); err != nil {
		return fmt.Errorf("failed to bulk delete tenants: %w", err)
	}

	// Invalidate caches
	for _, id := range tenantIds {
		s.invalidateTenantCaches(id)
	}
	s.invalidateCountCaches()
	s.invalidateListCaches()
	s.invalidateStatsCaches()

	return nil
}

// BulkUpdateTenantStatus updates status for multiple tenants
func (s *TenantService) BulkUpdateTenantStatus(ctx context.Context, req BulkUpdateStatusRequest) error {
	// Convert request to repository type
	bulkUpdateArg := tenantrepo.BulkUpdateTenantStatus{
		Status: string(req.Status),
		ID:     req.TenantIDs,
	}

	if err := s.dbManager.BulkUpdateTenantStatus(ctx, bulkUpdateArg); err != nil {
		return fmt.Errorf("failed to bulk update tenant status: %w", err)
	}

	// Invalidate caches for affected tenants
	for _, id := range req.TenantIDs {
		s.invalidateTenantCaches(id)
	}
	s.invalidateListCaches()
	s.invalidateStatsCaches()

	return nil
}

// FilterTenants retrieves filtered tenants
func (s *TenantService) FilterTenants(ctx context.Context, filter TenantFilter) ([]Tenant, error) {
	// Convert filter to repository type
	sortField := tenantrepo.TenantSortField(filter.SortBy)
	filterArg := tenantrepo.FilterTenants{
		NameFilter:     filter.NameFilter,
		StatusFilter:   filter.StatusFilter,
		IndustryFilter: filter.IndustryFilter,
		SortBy:         &sortField,
		OffsetCount:    filter.Offset,
		LimitCount:     filter.Limit,
	}

	repoTenants, err := s.dbManager.FilterTenants(ctx, filterArg)
	if err != nil {
		return nil, err
	}

	return s.convertToDomainTenants(repoTenants), nil
}

// GetActiveTenants retrieves active tenants with caching
func (s *TenantService) GetActiveTenants(ctx context.Context) ([]Tenant, error) {
	// Try cache first
	if cached, err := s.cache.Get(activTenantsKey); err == nil {
		var tenants []Tenant
		if json.Unmarshal([]byte(cached), &tenants) == nil {
			return tenants, nil
		}
	}

	// Fetch from database
	repoTenants, err := s.dbManager.GetActiveTenants(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active tenants: %w", err)
	}

	// Convert to domain types
	domainTenants := s.convertToDomainTenants(repoTenants)

	// Cache the result
	if data, err := json.Marshal(domainTenants); err == nil {
		s.cache.Set(activTenantsKey, string(data), shortCacheTTL)
	}

	return domainTenants, nil
}

// ListTenants retrieves tenants with pagination
func (s *TenantService) ListTenants(ctx context.Context, req ListTenantsRequest) ([]Tenant, error) {
	repoTenants, err := s.dbManager.ListTenants(ctx, req.Limit, req.Offset)
	if err != nil {
		return nil, err
	}

	return s.convertToDomainTenants(repoTenants), nil
}

// SearchTenantsByName searches tenants by name
func (s *TenantService) SearchTenantsByName(ctx context.Context, req SearchTenantsRequest) ([]Tenant, error) {
	searchArg := tenantrepo.SearchTenantsByName{
		Name:   req.Name,
		Offset: req.Offset,
		Limit:  req.Limit,
	}

	repoTenants, err := s.dbManager.SearchTenantsByName(ctx, searchArg)
	if err != nil {
		return nil, err
	}

	return s.convertToDomainTenants(repoTenants), nil
}

// CheckTenantExists checks if a tenant exists with caching
func (s *TenantService) CheckTenantExists(ctx context.Context, id int32) (bool, error) {
	cacheKey := fmt.Sprintf(tenantExistsKey, id)

	// Try cache first
	if cached, err := s.cache.Get(cacheKey); err == nil {
		if exists, err := strconv.ParseBool(cached); err == nil {
			return exists, nil
		}
	}

	// Check database
	exists, err := s.dbManager.CheckTenantExists(ctx, id)
	if err != nil {
		return false, fmt.Errorf("failed to check tenant existence: %w", err)
	}

	// Cache the result
	s.cache.Set(cacheKey, strconv.FormatBool(exists), shortCacheTTL)

	return exists, nil
}

// CheckSubdomainExists checks if a subdomain exists with caching
func (s *TenantService) CheckSubdomainExists(ctx context.Context, subdomain *string) (bool, error) {
	if subdomain == nil {
		return false, nil
	}

	cacheKey := fmt.Sprintf(subdomainExistsKey, *subdomain)

	// Try cache first
	if cached, err := s.cache.Get(cacheKey); err == nil {
		if exists, err := strconv.ParseBool(cached); err == nil {
			return exists, nil
		}
	}

	// Check database
	exists, err := s.dbManager.CheckSubdomainExists(ctx, subdomain)
	if err != nil {
		return false, fmt.Errorf("failed to check subdomain existence: %w", err)
	}

	// Cache the result
	s.cache.Set(cacheKey, strconv.FormatBool(exists), shortCacheTTL)

	return exists, nil
}

// GetTenantCount retrieves tenant count with caching
func (s *TenantService) GetTenantCount(ctx context.Context) (int64, error) {
	// Try cache first
	if cached, err := s.cache.Get(tenantCountKey); err == nil {
		if count, err := strconv.ParseInt(cached, 10, 64); err == nil {
			return count, nil
		}
	}

	// Fetch from database
	count, err := s.dbManager.CountTenants(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count tenants: %w", err)
	}

	// Cache the result
	s.cache.Set(tenantCountKey, strconv.FormatInt(count, 10), defaultCacheTTL)

	return count, nil
}

// GetTenantStats retrieves comprehensive tenant statistics with caching
func (s *TenantService) GetTenantStats(ctx context.Context) (TenantStats, error) {
	// Try cache first
	if cached, err := s.cache.Get(tenantStatsKey); err == nil {
		var stats TenantStats
		if json.Unmarshal([]byte(cached), &stats) == nil {
			return stats, nil
		}
	}

	// Calculate stats from database
	totalCount, err := s.dbManager.CountTenants(ctx)
	if err != nil {
		return TenantStats{}, fmt.Errorf("failed to get total tenant count: %w", err)
	}

	// Get active tenants count
	activeTenants, err := s.dbManager.GetActiveTenants(ctx)
	if err != nil {
		return TenantStats{}, fmt.Errorf("failed to get active tenants: %w", err)
	}

	// You would need to implement additional queries for suspended/pending counts
	// For now, we'll use placeholder logic
	stats := TenantStats{
		TotalTenants:     totalCount,
		ActiveTenants:    int64(len(activeTenants)),
		SuspendedTenants: 0, // Implement based on your needs
		PendingTenants:   0, // Implement based on your needs
	}

	// Cache the result
	if data, err := json.Marshal(stats); err == nil {
		s.cache.Set(tenantStatsKey, string(data), defaultCacheTTL)
	}

	return stats, nil
}

// GetTenantsByIndustry retrieves tenants grouped by industry
func (s *TenantService) GetTenantsByIndustry(ctx context.Context) ([]tenantrepo.GetTenantsByIndustryRow, error) {
	return s.dbManager.GetTenantsByIndustry(ctx)
}

// CountFilteredTenants counts tenants based on filters
func (s *TenantService) CountFilteredTenants(ctx context.Context, filter TenantFilter) (int64, error) {
	countArg := tenantrepo.CountFilteredTenants{
		NameFilter:     filter.NameFilter,
		StatusFilter:   filter.StatusFilter,
		IndustryFilter: filter.IndustryFilter,
	}

	return s.dbManager.CountFilteredTenants(ctx, countArg)
}

// Private helper methods

// validateTenantCreation validates business rules for tenant creation
func (s *TenantService) validateTenantCreation(ctx context.Context, req CreateTenantRequest) error {
	// Check if name already exists
	if exists, err := s.dbManager.CheckTenantNameExists(ctx, req.Name); err != nil {
		return fmt.Errorf("failed to check tenant name existence: %w", err)
	} else if exists {
		return fmt.Errorf("tenant name '%s' already exists", req.Name)
	}

	// Check if subdomain already exists (if provided)
	if req.Subdomain != nil {
		if exists, err := s.CheckSubdomainExists(ctx, req.Subdomain); err != nil {
			return fmt.Errorf("failed to check subdomain existence: %w", err)
		} else if exists {
			return fmt.Errorf("subdomain '%s' already exists", *req.Subdomain)
		}
	}

	return nil
}

// validateTenantUpdate validates business rules for tenant updates
func (s *TenantService) validateTenantUpdate(ctx context.Context, id int32, req UpdateTenantRequest) error {
	// Check if tenant exists
	if exists, err := s.CheckTenantExists(ctx, id); err != nil {
		return fmt.Errorf("failed to check tenant existence: %w", err)
	} else if !exists {
		return fmt.Errorf("tenant with ID %d does not exist", id)
	}

	// Check subdomain uniqueness if being updated
	if req.Subdomain != nil {
		if exists, err := s.CheckSubdomainExists(ctx, req.Subdomain); err != nil {
			return fmt.Errorf("failed to check subdomain existence: %w", err)
		} else if exists {
			// Check if the subdomain belongs to the current tenant
			currentTenant, err := s.GetTenantByID(ctx, id)
			if err != nil {
				return fmt.Errorf("failed to get current tenant: %w", err)
			}
			if currentTenant.Subdomain != *req.Subdomain {
				return fmt.Errorf("subdomain '%s' already exists", *req.Subdomain)
			}
		}
	}

	return nil
}

// Helper methods for type conversion

// convertToDomainTenant converts repository tenant to domain tenant
func (s *TenantService) convertToDomainTenant(repoTenant tenantrepo.Tenant) Tenant {
	domainTenant := Tenant{
		ID:        repoTenant.ID,
		Uuid:      repoTenant.Uuid,
		Name:      repoTenant.Name,
		Status:    TenantStatus(repoTenant.Status),
		CreatedAt: repoTenant.CreatedAt,
		UpdatedAt: repoTenant.UpdatedAt,
		DeletedAt: repoTenant.DeletedAt,
	}

	// Handle subdomain (convert from pointer to string)
	if repoTenant.Subdomain != nil {
		domainTenant.Subdomain = *repoTenant.Subdomain
	}

	// Handle industry (convert from pointer to string)
	if repoTenant.Industry != nil {
		domainTenant.Industry = *repoTenant.Industry
	}

	return domainTenant
}

// convertToDomainTenants converts slice of repository tenants to domain tenants
func (s *TenantService) convertToDomainTenants(repoTenants []tenantrepo.Tenant) []Tenant {
	domainTenants := make([]Tenant, len(repoTenants))
	for i, repoTenant := range repoTenants {
		domainTenants[i] = s.convertToDomainTenant(repoTenant)
	}
	return domainTenants
}

// cacheDomainTenant caches a domain tenant in multiple cache keys
func (s *TenantService) cacheDomainTenant(t Tenant) {
	if data, err := json.Marshal(t); err == nil {
		dataStr := string(data)

		// Cache by ID
		s.cache.Set(fmt.Sprintf(tenantByIDKey, t.ID), dataStr, defaultCacheTTL)

		// Cache by UUID if available
		if t.Uuid != uuid.Nil {
			s.cache.Set(fmt.Sprintf(tenantByUUIDKey, t.Uuid.String()), dataStr, defaultCacheTTL)
		}

		// Cache by subdomain if available
		if t.Subdomain != "" {
			s.cache.Set(fmt.Sprintf(tenantBySubdomainKey, t.Subdomain), dataStr, defaultCacheTTL)
		}
	}
}

// cacheTenant caches a tenant in multiple cache keys (kept for compatibility)
func (s *TenantService) cacheTenant(t tenantrepo.Tenant) {
	domainTenant := s.convertToDomainTenant(t)
	s.cacheDomainTenant(domainTenant)
}

// invalidateTenantCaches removes all cache entries for a specific tenant
func (s *TenantService) invalidateTenantCaches(tenantID int32) {
	s.cache.Delete(fmt.Sprintf(tenantByIDKey, tenantID))
	s.cache.Delete(fmt.Sprintf(tenantExistsKey, tenantID))
	s.cache.Delete(currentTenantKey)
}

// invalidateCountCaches removes count-related cache entries
func (s *TenantService) invalidateCountCaches() {
	s.cache.Delete(tenantCountKey)
}

// invalidateListCaches removes list-related cache entries
func (s *TenantService) invalidateListCaches() {
	s.cache.Delete(activTenantsKey)
}

// invalidateStatsCaches removes statistics cache entries
func (s *TenantService) invalidateStatsCaches() {
	s.cache.Delete(tenantStatsKey)
}

// TenantConfiguration methods

// GetTenantConfiguration retrieves tenant configuration with caching
func (s *TenantService) GetTenantConfiguration(ctx context.Context, tenantID int32) (TenantConfiguration, error) {
	cacheKey := fmt.Sprintf("tenant:config:%d", tenantID)

	// Try cache first
	if cached, err := s.cache.Get(cacheKey); err == nil {
		var config TenantConfiguration
		if json.Unmarshal([]byte(cached), &config) == nil {
			return config, nil
		}
	}

	// For now, return empty config as the repository interface doesn't include config methods
	// You would need to add these methods to your TenantRepo interface
	return TenantConfiguration{}, fmt.Errorf("tenant configuration methods not implemented in repository")
}

// UpdateTenantConfiguration updates tenant configuration
func (s *TenantService) UpdateTenantConfiguration(ctx context.Context, config TenantConfiguration) error {
	// You would need to add this method to your TenantRepo interface
	// For now, just invalidate cache
	cacheKey := fmt.Sprintf("tenant:config:%d", config.TenantID)
	s.cache.Delete(cacheKey)

	return fmt.Errorf("tenant configuration methods not implemented in repository")
}
