package tenant

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/storage/cache"
)

// TenantService provides business logic for tenant operations
type TenantService struct {
	tenantRepo TenantRepository
	cache      cache.CacheStore
}

// NewTenantService creates a new tenant service instance
func NewTenantService(tenantRepo TenantRepository, cache cache.CacheStore) *TenantService {
	return &TenantService{
		tenantRepo: tenantRepo,
		cache:      cache,
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

	// Cache TTL durations
	defaultCacheTTL = 15 * time.Minute
	shortCacheTTL   = 5 * time.Minute
	longCacheTTL    = 30 * time.Minute
)

// CreateTenant creates a new tenant and invalidates relevant caches
func (s *TenantService) CreateTenant(ctx context.Context, req *CreateTenantRequest) (*Tenant, error) {
	// Validate business rules
	if err := s.validateTenantCreation(ctx, req); err != nil {
		return nil, err
	}

	// Create tenant
	newTenant, err := s.tenantRepo.Create(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	// Invalidate relevant caches
	s.invalidateCountCaches()
	s.invalidateListCaches()

	// Cache the new tenant
	s.cacheTenant(*newTenant)

	return newTenant, nil
}

// GetTenant retrieves a tenant by ID with caching
func (s *TenantService) GetTenant(ctx context.Context, id int32) (*Tenant, error) {
	cacheKey := fmt.Sprintf(tenantByIDKey, id)

	// Try cache first
	if cached, err := s.cache.Get(cacheKey); err == nil {
		var t Tenant
		if json.Unmarshal([]byte(cached), &t) == nil {
			return &t, nil
		}
	}

	// Fetch from repository
	t, err := s.tenantRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant by ID: %w", err)
	}

	// Cache the result
	s.cacheTenant(*t)

	return t, nil
}

// GetTenantByUUID retrieves a tenant by UUID with caching
func (s *TenantService) GetTenantByUUID(ctx context.Context, tenantUUID uuid.UUID) (*Tenant, error) {
	cacheKey := fmt.Sprintf(tenantByUUIDKey, tenantUUID.String())

	// Try cache first
	if cached, err := s.cache.Get(cacheKey); err == nil {
		var t Tenant
		if json.Unmarshal([]byte(cached), &t) == nil {
			return &t, nil
		}
	}

	// Fetch from repository
	t, err := s.tenantRepo.GetByUUID(ctx, tenantUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant by UUID: %w", err)
	}

	// Cache the result
	s.cacheTenant(*t)

	return t, nil
}

// GetTenantBySubdomain retrieves a tenant by subdomain with caching
func (s *TenantService) GetTenantBySubdomain(ctx context.Context, subdomain string) (*Tenant, error) {
	cacheKey := fmt.Sprintf(tenantBySubdomainKey, subdomain)

	// Try cache first
	if cached, err := s.cache.Get(cacheKey); err == nil {
		var t Tenant
		if json.Unmarshal([]byte(cached), &t) == nil {
			return &t, nil
		}
	}

	// Fetch from repository
	t, err := s.tenantRepo.GetBySubdomain(ctx, subdomain)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant by subdomain: %w", err)
	}

	// Cache the result
	s.cacheTenant(*t)

	return t, nil
}

// ListTenants retrieves a paginated list of tenants
func (s *TenantService) ListTenants(ctx context.Context, limit, offset int32) ([]*Tenant, error) {
	tenants, err := s.tenantRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}
	return tenants, nil
}

// SearchTenants searches tenants by name
func (s *TenantService) SearchTenants(ctx context.Context, name string, limit, offset int32) ([]*Tenant, error) {
	tenants, err := s.tenantRepo.SearchByName(ctx, name, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search tenants: %w", err)
	}
	return tenants, nil
}

// FilterTenants retrieves filtered tenants
func (s *TenantService) FilterTenants(ctx context.Context, filter *TenantFilter) ([]*Tenant, error) {
	tenants, err := s.tenantRepo.Filter(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to filter tenants: %w", err)
	}
	return tenants, nil
}

// ActivateTenant activates a tenant
func (s *TenantService) ActivateTenant(ctx context.Context, id int32) (*Tenant, error) {
	tenant, err := s.tenantRepo.UpdateStatus(ctx, id, TenantStatusActive)
	if err != nil {
		return nil, fmt.Errorf("failed to activate tenant: %w", err)
	}
	
	// Invalidate caches
	s.invalidateTenantCaches(id)
	s.invalidateListCaches()
	
	// Cache the updated tenant
	s.cacheTenant(*tenant)
	
	return tenant, nil
}

// SuspendTenant suspends a tenant
func (s *TenantService) SuspendTenant(ctx context.Context, id int32) (*Tenant, error) {
	tenant, err := s.tenantRepo.UpdateStatus(ctx, id, TenantStatusSuspended)
	if err != nil {
		return nil, fmt.Errorf("failed to suspend tenant: %w", err)
	}
	
	// Invalidate caches
	s.invalidateTenantCaches(id)
	s.invalidateListCaches()
	
	// Cache the updated tenant
	s.cacheTenant(*tenant)
	
	return tenant, nil
}

// DeactivateTenant deactivates a tenant
func (s *TenantService) DeactivateTenant(ctx context.Context, id int32) (*Tenant, error) {
	tenant, err := s.tenantRepo.UpdateStatus(ctx, id, TenantStatusInactive)
	if err != nil {
		return nil, fmt.Errorf("failed to deactivate tenant: %w", err)
	}
	
	// Invalidate caches
	s.invalidateTenantCaches(id)
	s.invalidateListCaches()
	
	// Cache the updated tenant
	s.cacheTenant(*tenant)
	
	return tenant, nil
}

// ValidateSubdomain validates if a subdomain is available
func (s *TenantService) ValidateSubdomain(ctx context.Context, subdomain string) error {
	exists, err := s.tenantRepo.CheckSubdomainExists(ctx, subdomain)
	if err != nil {
		return fmt.Errorf("failed to check subdomain existence: %w", err)
	}
	if exists {
		return fmt.Errorf("subdomain '%s' already exists", subdomain)
	}
	return nil
}

// ValidateTenantName validates if a tenant name is available
func (s *TenantService) ValidateTenantName(ctx context.Context, name string) error {
	exists, err := s.tenantRepo.CheckNameExists(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to check tenant name existence: %w", err)
	}
	if exists {
		return fmt.Errorf("tenant name '%s' already exists", name)
	}
	return nil
}

// GetTenantStats retrieves tenant statistics
func (s *TenantService) GetTenantStats(ctx context.Context) (*TenantStats, error) {
	stats, err := s.tenantRepo.GetStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant stats: %w", err)
	}
	return stats, nil
}

// UpdateTenant updates a tenant and invalidates caches
func (s *TenantService) UpdateTenant(ctx context.Context, id int32, req *UpdateTenantRequest) (*Tenant, error) {
	// Validate business rules
	if err := s.validateTenantUpdate(ctx, id, req); err != nil {
		return nil, err
	}

	// Update tenant
	updatedTenant, err := s.tenantRepo.Update(ctx, id, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	// Invalidate caches for this tenant
	s.invalidateTenantCaches(id)
	s.invalidateListCaches()

	// Cache the updated tenant
	s.cacheTenant(*updatedTenant)

	return updatedTenant, nil
}

// DeleteTenant soft deletes a tenant and invalidates caches
func (s *TenantService) DeleteTenant(ctx context.Context, id int32) error {
	if err := s.tenantRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	// Invalidate caches
	s.invalidateTenantCaches(id)
	s.invalidateCountCaches()
	s.invalidateListCaches()

	return nil
}

// BulkDeleteTenants soft deletes multiple tenants
func (s *TenantService) BulkDeleteTenants(ctx context.Context, tenantIds []int32) error {
	if err := s.tenantRepo.BulkSoftDeleteTenants(ctx, tenantIds); err != nil {
		return fmt.Errorf("failed to bulk delete tenants: %w", err)
	}

	// Invalidate caches
	for _, id := range tenantIds {
		s.invalidateTenantCaches(id)
	}
	s.invalidateCountCaches()
	s.invalidateListCaches()

	return nil
}

// FilterTenants retrieves filtered tenants with caching for common filters
func (s *TenantService) FilterTenants(ctx context.Context, arg tenant.FilterTenants) ([]tenant.Tenant, error) {
	return s.tenantRepo.FilterTenants(ctx, arg)
}

// GetActiveTenants retrieves active tenants with caching
func (s *TenantService) GetActiveTenants(ctx context.Context) ([]tenant.Tenant, error) {
	// Try cache first
	if cached, err := s.cache.Get(activTenantsKey); err == nil {
		var tenants []tenant.Tenant
		if json.Unmarshal([]byte(cached), &tenants) == nil {
			return tenants, nil
		}
	}

	// Fetch from repository
	tenants, err := s.tenantRepo.GetActiveTenants(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active tenants: %w", err)
	}

	// Cache the result
	if data, err := json.Marshal(tenants); err == nil {
		s.cache.Set(activTenantsKey, string(data), shortCacheTTL)
	}

	return tenants, nil
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

	// Check repository
	exists, err := s.tenantRepo.CheckTenantExists(ctx, id)
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

	// Check repository
	exists, err := s.tenantRepo.CheckSubdomainExists(ctx, subdomain)
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

	// Fetch from repository
	count, err := s.tenantRepo.CountTenants(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count tenants: %w", err)
	}

	// Cache the result
	s.cache.Set(tenantCountKey, strconv.FormatInt(count, 10), defaultCacheTTL)

	return count, nil
}

// BulkUpdateTenantStatus updates status for multiple tenants
func (s *TenantService) BulkUpdateTenantStatus(ctx context.Context, arg tenant.BulkUpdateTenantStatus) error {
	if err := s.tenantRepo.BulkUpdateTenantStatus(ctx, arg); err != nil {
		return fmt.Errorf("failed to bulk update tenant status: %w", err)
	}

	// Invalidate caches for affected tenants
	for _, id := range arg.ID {
		s.invalidateTenantCaches(id)
	}
	s.invalidateListCaches()

	return nil
}

// Private helper methods

// validateTenantCreation validates business rules for tenant creation
func (s *TenantService) validateTenantCreation(ctx context.Context, req *CreateTenantRequest) error {
	// Check if name already exists
	if exists, err := s.tenantRepo.CheckNameExists(ctx, req.Name); err != nil {
		return fmt.Errorf("failed to check tenant name existence: %w", err)
	} else if exists {
		return fmt.Errorf("tenant name '%s' already exists", req.Name)
	}

	// Check if subdomain already exists (if provided)
	if req.Subdomain != nil {
		if exists, err := s.tenantRepo.CheckSubdomainExists(ctx, *req.Subdomain); err != nil {
			return fmt.Errorf("failed to check subdomain existence: %w", err)
		} else if exists {
			return fmt.Errorf("subdomain '%s' already exists", *req.Subdomain)
		}
	}

	return nil
}

// validateTenantUpdate validates business rules for tenant updates
func (s *TenantService) validateTenantUpdate(ctx context.Context, id int32, req *UpdateTenantRequest) error {
	// Check if tenant exists
	if exists, err := s.tenantRepo.CheckExists(ctx, id); err != nil {
		return fmt.Errorf("failed to check tenant existence: %w", err)
	} else if !exists {
		return fmt.Errorf("tenant with ID %d does not exist", id)
	}

	// Check subdomain uniqueness if being updated
	if req.Subdomain != nil {
		if exists, err := s.tenantRepo.CheckSubdomainExists(ctx, *req.Subdomain); err != nil {
			return fmt.Errorf("failed to check subdomain existence: %w", err)
		} else if exists {
			// Check if the subdomain belongs to the current tenant
			currentTenant, err := s.GetTenant(ctx, id)
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

// cacheTenant caches a tenant in multiple cache keys
func (s *TenantService) cacheTenant(t Tenant) {
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
