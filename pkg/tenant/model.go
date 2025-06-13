package tenant

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// =====================================================
// DOMAIN MODELS (Core)
// =====================================================

type TenantStatus string

const (
	StatusActive    TenantStatus = "active"
	StatusSuspended TenantStatus = "suspended"
	StatusPending   TenantStatus = "pending"
)

type Tenant struct {
	ID         int32     `json:"id"`
	UUID       uuid.UUID `json:"uuid"`
	Name       string    `json:"name"`
	Subdomain  *string   `json:"subdomain,omitempty"`
	Status     string    `json:"status"`
	Industry   *string   `json:"industry,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
}

type TenantConfiguration struct {
	TenantID       int32           `json:"tenant_id"`
	MaxUsers       int32           `json:"max_users"`
	StorageQuota   int64           `json:"storage_quota"`
	Features       json.RawMessage `json:"features"`
	ModulesEnabled json.RawMessage `json:"modules_enabled"`
}

type TenantWithConfig struct {
	Tenant
	Configuration *TenantConfiguration `json:"configuration,omitempty"`
}

// =====================================================
// DOMAIN VALUE OBJECTS
// =====================================================

// =====================================================
// DOMAIN ERRORS
// =====================================================

type DomainError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

func (e DomainError) Error() string {
	return e.Message
}

var (
	ErrTenantNotFound           = DomainError{Code: "TENANT_NOT_FOUND", Message: "tenant not found"}
	ErrTenantAlreadyExists      = DomainError{Code: "TENANT_ALREADY_EXISTS", Message: "tenant already exists"}
	ErrSubdomainAlreadyExists   = DomainError{Code: "SUBDOMAIN_ALREADY_EXISTS", Message: "subdomain already exists"}
	ErrConfigurationNotFound    = DomainError{Code: "CONFIGURATION_NOT_FOUND", Message: "tenant configuration not found"}
	ErrInvalidTenantStatus      = DomainError{Code: "INVALID_TENANT_STATUS", Message: "invalid tenant status"}
	ErrUnauthorized            = DomainError{Code: "UNAUTHORIZED", Message: "unauthorized access"}
	ErrValidationFailed        = DomainError{Code: "VALIDATION_FAILED", Message: "validation failed"}
)

// =====================================================
// PORT INTERFACES (Application Layer)
// =====================================================


// =====================================================
// APPLICATION SERVICE IMPLEMENTATION
// =====================================================

type tenantService struct {
	tenantRepo       TenantRepository
	tenantConfigRepo TenantConfigRepository
	validator        Validator
}

type tenantConfigService struct {
	tenantConfigRepo TenantConfigRepository
	validator        Validator
}

// Validator interface for input validation
type Validator interface {
	Validate(interface{}) error
}

// Constructor functions
func NewTenantService(tenantRepo TenantRepository, tenantConfigRepo TenantConfigRepository, validator Validator) TenantService {
	return &tenantService{
		tenantRepo:       tenantRepo,
		tenantConfigRepo: tenantConfigRepo,
		validator:        validator,
	}
}

func NewTenantConfigService(tenantConfigRepo TenantConfigRepository, validator Validator) TenantConfigService {
	return &tenantConfigService{
		tenantConfigRepo: tenantConfigRepo,
		validator:        validator,
	}
}

// =====================================================
// TENANT SERVICE IMPLEMENTATION
// =====================================================

func (s *tenantService) CreateTenant(ctx context.Context, req CreateTenantRequest) (*Tenant, error) {
	if err := s.validator.Validate(req); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrValidationFailed, err)
	}

	// Check if tenant name already exists
	exists, err := s.tenantRepo.CheckTenantNameExists(ctx, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check tenant name existence: %w", err)
	}
	if exists {
		return nil, ErrTenantAlreadyExists
	}

	// Check if subdomain already exists (if provided)
	if req.Subdomain != nil {
		exists, err := s.tenantRepo.CheckSubdomainExists(ctx, *req.Subdomain)
		if err != nil {
			return nil, fmt.Errorf("failed to check subdomain existence: %w", err)
		}
		if exists {
			return nil, ErrSubdomainAlreadyExists
		}
	}

	params := CreateTenantParams{
		Name:      req.Name,
		Subdomain: req.Subdomain,
		Status:    req.Status,
		Industry:  req.Industry,
	}

	tenant, err := s.tenantRepo.CreateTenant(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	return tenant, nil
}

func (s *tenantService) GetTenantByID(ctx context.Context, id int32) (*Tenant, error) {
	tenant, err := s.tenantRepo.GetTenantByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrTenantNotFound
		}
		return nil, fmt.Errorf("failed to get tenant by ID: %w", err)
	}
	return tenant, nil
}

func (s *tenantService) GetTenantByUUID(ctx context.Context, uuid uuid.UUID) (*Tenant, error) {
	tenant, err := s.tenantRepo.GetTenantByUUID(ctx, uuid)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrTenantNotFound
		}
		return nil, fmt.Errorf("failed to get tenant by UUID: %w", err)
	}
	return tenant, nil
}

func (s *tenantService) GetTenantBySubdomain(ctx context.Context, subdomain string) (*Tenant, error) {
	tenant, err := s.tenantRepo.GetTenantBySubdomain(ctx, subdomain)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrTenantNotFound
		}
		return nil, fmt.Errorf("failed to get tenant by subdomain: %w", err)
	}
	return tenant, nil
}

func (s *tenantService) ListTenants(ctx context.Context, filter TenantFilter) ([]*Tenant, error) {
	if err := s.validator.Validate(filter); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrValidationFailed, err)
	}

	params := ListTenantsParams{
		Limit:  filter.Limit,
		Offset: filter.Offset,
	}

	tenants, err := s.tenantRepo.ListTenants(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}

	return tenants, nil
}

func (s *tenantService) UpdateTenant(ctx context.Context, id int32, req UpdateTenantRequest) (*Tenant, error) {
	if err := s.validator.Validate(req); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrValidationFailed, err)
	}

	// Check if tenant exists
	exists, err := s.tenantRepo.CheckTenantExists(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to check tenant existence: %w", err)
	}
	if !exists {
		return nil, ErrTenantNotFound
	}

	params := UpdateTenantParams{
		ID:        id,
		Name:      req.Name,
		Subdomain: req.Subdomain,
		Status:    req.Status,
		Industry:  req.Industry,
	}

	tenant, err := s.tenantRepo.UpdateTenant(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	return tenant, nil
}

func (s *tenantService) DeleteTenant(ctx context.Context, id int32) error {
	// Check if tenant exists
	exists, err := s.tenantRepo.CheckTenantExists(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check tenant existence: %w", err)
	}
	if !exists {
		return ErrTenantNotFound
	}

	if err := s.tenantRepo.SoftDeleteTenant(ctx, id); err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	return nil
}

func (s *tenantService) GetTenantStats(ctx context.Context) (*TenantStats, error) {
	stats, err := s.tenantRepo.GetTenantStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant stats: %w", err)
	}
	return stats, nil
}

func (s *tenantService) GetCurrentTenant(ctx context.Context) (*Tenant, error) {
	tenant, err := s.tenantRepo.GetCurrentTenant(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrTenantNotFound
		}
		return nil, fmt.Errorf("failed to get current tenant: %w", err)
	}
	return tenant, nil
}

func (s *tenantService) UpdateCurrentTenant(ctx context.Context, req UpdateTenantRequest) (*Tenant, error) {
	if err := s.validator.Validate(req); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrValidationFailed, err)
	}

	params := UpdateCurrentTenantParams{
		Name:      req.Name,
		Subdomain: req.Subdomain,
		Status:    req.Status,
		Industry:  req.Industry,
	}

	tenant, err := s.tenantRepo.UpdateCurrentTenant(ctx, params)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrTenantNotFound
		}
		return nil, fmt.Errorf("failed to update current tenant: %w", err)
	}

	return tenant, nil
}

// =====================================================
// TENANT CONFIG SERVICE IMPLEMENTATION
// =====================================================

func (s *tenantConfigService) CreateTenantConfiguration(ctx context.Context, tenantID int32, req CreateTenantConfigRequest) (*TenantConfiguration, error) {
	if err := s.validator.Validate(req); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrValidationFailed, err)
	}

	params := CreateTenantConfigurationParams{
		TenantID:       tenantID,
		MaxUsers:       req.MaxUsers,
		StorageQuota:   req.StorageQuota,
		Features:       req.Features,
		ModulesEnabled: req.ModulesEnabled,
	}

	config, err := s.tenantConfigRepo.CreateTenantConfiguration(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant configuration: %w", err)
	}

	return config, nil
}

func (s *tenantConfigService) GetTenantConfiguration(ctx context.Context, tenantID int32) (*TenantConfiguration, error) {
	config, err := s.tenantConfigRepo.GetTenantConfiguration(ctx, tenantID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrConfigurationNotFound
		}
		return nil, fmt.Errorf("failed to get tenant configuration: %w", err)
	}
	return config, nil
}

func (s *tenantConfigService) UpdateTenantConfiguration(ctx context.Context, tenantID int32, req UpdateTenantConfigRequest) (*TenantConfiguration, error) {
	if err := s.validator.Validate(req); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrValidationFailed, err)
	}

	// Get current configuration to merge updates
	current, err := s.tenantConfigRepo.GetTenantConfiguration(ctx, tenantID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrConfigurationNotFound
		}
		return nil, fmt.Errorf("failed to get current configuration: %w", err)
	}

	// Merge updates with current values
	params := UpdateTenantConfigurationParams{
		TenantID:       tenantID,
		MaxUsers:       current.MaxUsers,
		StorageQuota:   current.StorageQuota,
		Features:       current.Features,
		ModulesEnabled: current.ModulesEnabled,
	}

	if req.MaxUsers != nil {
		params.MaxUsers = *req.MaxUsers
	}
	if req.StorageQuota != nil {
		params.StorageQuota = *req.StorageQuota
	}
	if req.Features != nil {
		params.Features = *req.Features
	}
	if req.ModulesEnabled != nil {
		params.ModulesEnabled = *req.ModulesEnabled
	}

	config, err := s.tenantConfigRepo.UpdateTenantConfiguration(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update tenant configuration: %w", err)
	}

	return config, nil
}

func (s *tenantConfigService) DeleteTenantConfiguration(ctx context.Context, tenantID int32) error {
	if err := s.tenantConfigRepo.DeleteTenantConfiguration(ctx, tenantID); err != nil {
		return fmt.Errorf("failed to delete tenant configuration: %w", err)
	}
	return nil
}

func (s *tenantConfigService) GetCurrentTenantConfiguration(ctx context.Context) (*TenantConfiguration, error) {
	config, err := s.tenantConfigRepo.GetCurrentTenantConfiguration(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrConfigurationNotFound
		}
		return nil, fmt.Errorf("failed to get current tenant configuration: %w", err)
	}
	return config, nil
}

func (s *tenantConfigService) UpdateCurrentTenantConfiguration(ctx context.Context, req UpdateTenantConfigRequest) (*TenantConfiguration, error) {
	if err := s.validator.Validate(req); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrValidationFailed, err)
	}

	var config *TenantConfiguration
	var err error

	// Apply updates one by one based on what's provided
	if req.MaxUsers != nil {
		config, err = s.tenantConfigRepo.UpdateCurrentTenantMaxUsers(ctx, *req.MaxUsers)
		if err != nil {
			return nil, fmt.Errorf("failed to update max users: %w", err)
		}
	}

	if req.StorageQuota != nil {
		config, err = s.tenantConfigRepo.UpdateCurrentTenantStorageQuota(ctx, *req.StorageQuota)
		if err != nil {
			return nil, fmt.Errorf("failed to update storage quota: %w", err)
		}
	}

	if req.Features != nil {
		config, err = s.tenantConfigRepo.UpdateCurrentTenantFeatures(ctx, *req.Features)
		if err != nil {
			return nil, fmt.Errorf("failed to update features: %w", err)
		}
	}

	if req.ModulesEnabled != nil {
		config, err = s.tenantConfigRepo.UpdateCurrentTenantModules(ctx, *req.ModulesEnabled)
		if err != nil {
			return nil, fmt.Errorf("failed to update modules: %w", err)
		}
	}

	// If no specific updates were made, get the current configuration
	if config == nil {
		config, err = s.tenantConfigRepo.GetCurrentTenantConfiguration(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get current configuration: %w", err)
		}
	}

	return config, nil
}

func (s *tenantConfigService) CreateCurrentTenantConfiguration(ctx context.Context, req CreateTenantConfigRequest) (*TenantConfiguration, error) {
	if err := s.validator.Validate(req); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrValidationFailed, err)
	}

	params := CreateCurrentTenantConfigurationParams{
		MaxUsers:       req.MaxUsers,
		StorageQuota:   req.StorageQuota,
		Features:       req.Features,
		ModulesEnabled: req.ModulesEnabled,
	}

	config, err := s.tenantConfigRepo.CreateCurrentTenantConfiguration(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create current tenant configuration: %w", err)
	}

	return config, nil
}

func (s *tenantConfigService) CheckCurrentTenantHasFeature(ctx context.Context, feature string) (bool, error) {
	hasFeature, err := s.tenantConfigRepo.CheckCurrentTenantHasFeature(ctx, feature)
	if err != nil {
		return false, fmt.Errorf("failed to check tenant feature: %w", err)
	}
	return hasFeature, nil
}

func (s *tenantConfigService) CheckCurrentTenantHasModule(ctx context.Context, module string) (bool, error) {
	hasModule, err := s.tenantConfigRepo.CheckCurrentTenantHasModule(ctx, module)
	if err != nil {
		return false, fmt.Errorf("failed to check tenant module: %w", err)
	}
	return hasModule, nil
}
