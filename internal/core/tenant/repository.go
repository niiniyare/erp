package tenant

//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// Repository defines the interface for tenant data access
// All operations use UUID-based tenant identification following RLS patterns
type Repository interface {
	// Tenant Provisioning
	ProvisionTenant(ctx context.Context, req ProvisionTenantRequest) (*ProvisionedTenantInfo, error)

	// Core CRUD operations (UUID-based)
	Create(ctx context.Context, tenant *Tenant) error
	GetByID(ctx context.Context, id uuid.UUID) (*Tenant, error)
	Update(ctx context.Context, id uuid.UUID, updates UpdateTenantRequest) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetTenantConfiguration(ctx context.Context) (*db.TenantConfiguration, error)

	// Query operations (UUID-based with RLS)
	List(ctx context.Context, offset, limit int) ([]*Tenant, error)
	Exists(ctx context.Context, tenantID uuid.UUID) (bool, error)

	// Subdomain resolution operations (resolve to UUID)
	GetBySubdomain(ctx context.Context, subdomain string) (*Tenant, error)
	ResolveSubdomainToID(ctx context.Context, subdomain string) (uuid.UUID, error)

	// Context-aware operations (for RLS)
	WithTenantContext(ctx context.Context, tenantID uuid.UUID, fn func(context.Context) error) error
	GetCurrentTenant(ctx context.Context) (*Tenant, error)
	ValidateCurrentTenant(ctx context.Context) error

	// Database session context management
	SetTenant(ctx context.Context, tenantID uuid.UUID) error
	GetTenant(ctx context.Context) (uuid.UUID, error)
	ResetTenant(ctx context.Context) error
}

// repository implements Repository interface
type repository struct {
	store  db.Store
	tracer tracing.TracingService
}

// NewRepository creates a new tenant repository
func NewRepository(store db.Store, tracer tracing.TracingService) Repository {
	return &repository{store: store, tracer: tracer}
}

// Create implements Repository.Create
// Creates a new tenant using SQLC-generated functions with proper tracing
func (r *repository) Create(ctx context.Context, tenant *Tenant) error {
	ctx, span := r.tracer.StartSpan(ctx, "tenant.repository.Create")
	defer span.End()

	// Add tracing attributes
	span.SetAttributes(
		attribute.String("tenant.id", tenant.ID.String()),
		attribute.String("tenant.name", tenant.Name),
		attribute.String("tenant.subdomain", *tenant.Subdomain),
	)

	logger.DebugContext(ctx, "Creating tenant in database", logger.Fields{
		"tenant_id":   tenant.ID.String(),
		"tenant_name": tenant.Name,
		"subdomain":   tenant.Subdomain,
		"operation":   "repository.Create",
	})

	// Use SQLC-generated CreateTenant function
	params := db.CreateTenantParams{
		Name:      tenant.Name,
		Slug:      tenant.Slug,
		Email:     tenant.Email,
		Subdomain: tenant.Subdomain,
		Status:    string(tenant.Status),
		Industry:  tenant.Industry,
	}

	createdTenant, err := r.store.CreateTenant(ctx, params)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to create tenant")

		logger.ErrorContext(ctx, "Failed to create tenant in database", logger.Fields{
			"tenant_id": tenant.ID.String(),
			"error":     err.Error(),
			"operation": "repository.Create",
		})
		return fmt.Errorf("failed to create tenant: %w", err)
	}

	// Update the tenant with the generated ID
	tenant.ID = createdTenant.ID

	logger.InfoContext(ctx, "Successfully created tenant", logger.Fields{
		"tenant_id":   tenant.ID.String(),
		"tenant_name": tenant.Name,
		"operation":   "repository.Create",
	})

	return nil
}

// GetByID implements Repository.GetByID
// Retrieves tenant by UUID with RLS enforcement and comprehensive tracing
func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Tenant, error) {
	ctx, span := r.tracer.StartSpan(ctx, "tenant.repository.GetByID")
	defer span.End()

	span.SetAttributes(attribute.String("tenant.id", id.String()))

	logger.DebugContext(ctx, "Retrieving tenant by ID", logger.Fields{
		"tenant_id": id.String(),
		"operation": "repository.GetByID",
	})

	// Use SQLC-generated GetTenantByID function (RLS will be applied automatically)
	sqlcTenant, err := r.store.GetTenantByID(ctx, id)
	if err != nil {
		span.RecordError(err)

		if err.Error() == "no rows in result set" {
			span.SetStatus(codes.Error, "Tenant not found")
			logger.WarnContext(ctx, "Tenant not found", logger.Fields{
				"tenant_id": id.String(),
				"operation": "repository.GetByID",
			})
			return nil, errors.ErrTenantNotFound
		}

		span.SetStatus(codes.Error, "Database error")
		logger.ErrorContext(ctx, "Failed to get tenant by ID", logger.Fields{
			"tenant_id": id.String(),
			"error":     err.Error(),
			"operation": "repository.GetByID",
		})
		return nil, fmt.Errorf("failed to get tenant by ID: %w", err)
	}

	tenant, err := FromSQLCTenant(sqlcTenant)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Conversion error")
		return nil, fmt.Errorf("failed to convert tenant: %w", err)
	}

	logger.DebugContext(ctx, "Successfully retrieved tenant by ID", logger.Fields{
		"tenant_id":   id.String(),
		"tenant_name": tenant.Name,
		"operation":   "repository.GetByID",
	})

	return tenant, nil
}

// GetBySubdomain implements Repository.GetBySubdomain
// Resolves subdomain to tenant (does not enforce RLS as this is for resolution)
func (r *repository) GetBySubdomain(ctx context.Context, subdomain string) (*Tenant, error) {
	ctx, span := r.tracer.StartSpan(ctx, "tenant.repository.GetBySubdomain")
	defer span.End()

	span.SetAttributes(attribute.String("tenant.subdomain", subdomain))

	logger.DebugContext(ctx, "Retrieving tenant by subdomain", logger.Fields{
		"subdomain": subdomain,
		"operation": "repository.GetBySubdomain",
	})

	// Use SQLC-generated function - note: this doesn't enforce RLS since it's for subdomain resolution
	sqlcTenant, err := r.store.GetTenantByUUID(ctx, &subdomain)
	if err != nil {
		span.RecordError(err)

		if err.Error() == "no rows in result set" {
			span.SetStatus(codes.Error, "Tenant not found")
			logger.WarnContext(ctx, "Tenant not found by subdomain", logger.Fields{
				"subdomain": subdomain,
				"operation": "repository.GetBySubdomain",
			})
			return nil, errors.ErrTenantNotFound
		}

		span.SetStatus(codes.Error, "Database error")
		logger.ErrorContext(ctx, "Failed to get tenant by subdomain", logger.Fields{
			"subdomain": subdomain,
			"error":     err.Error(),
			"operation": "repository.GetBySubdomain",
		})
		return nil, fmt.Errorf("failed to get tenant by subdomain: %w", err)
	}

	tenant, err := FromSQLCTenant(sqlcTenant)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Conversion error")
		return nil, fmt.Errorf("failed to convert tenant: %w", err)
	}

	logger.DebugContext(ctx, "Successfully retrieved tenant by subdomain", logger.Fields{
		"subdomain":   subdomain,
		"tenant_id":   tenant.ID.String(),
		"tenant_name": tenant.Name,
		"operation":   "repository.GetBySubdomain",
	})

	return tenant, nil
}

// Update implements Repository.Update
// Updates tenant using SQLC-generated functions with RLS enforcement
func (r *repository) Update(ctx context.Context, id uuid.UUID, updates UpdateTenantRequest) error {
	ctx, span := r.tracer.StartSpan(ctx, "tenant.repository.Update")
	defer span.End()

	span.SetAttributes(attribute.String("tenant.id", id.String()))

	logger.DebugContext(ctx, "Updating tenant", logger.Fields{
		"tenant_id": id.String(),
		"operation": "repository.Update",
	})

	// Convert status if provided
	var status *string
	if updates.Status != nil {
		statusStr := string(*updates.Status)
		status = &statusStr
		span.SetAttributes(attribute.String("tenant.new_status", statusStr))
	}

	// Use SQLC-generated UpdateTenant function (RLS will be applied)
	params := db.UpdateTenantParams{
		Name:      updates.Name,
		Subdomain: updates.Subdomain,
		Status:    status,
		Industry:  updates.Industry,
		ID:        id,
	}

	updatedTenant, err := r.store.UpdateTenant(ctx, params)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to update tenant")

		logger.ErrorContext(ctx, "Failed to update tenant", logger.Fields{
			"tenant_id": id.String(),
			"error":     err.Error(),
			"operation": "repository.Update",
		})
		return fmt.Errorf("failed to update tenant: %w", err)
	}

	span.SetAttributes(attribute.String("tenant.updated_name", updatedTenant.Name))

	logger.InfoContext(ctx, "Successfully updated tenant", logger.Fields{
		"tenant_id":   id.String(),
		"tenant_name": updatedTenant.Name,
		"operation":   "repository.Update",
	})

	return nil
}

// Delete implements Repository.Delete
// Soft deletes tenant using SQLC-generated functions with RLS enforcement
func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, span := r.tracer.StartSpan(ctx, "tenant.repository.Delete")
	defer span.End()

	span.SetAttributes(attribute.String("tenant.id", id.String()))

	logger.DebugContext(ctx, "Soft deleting tenant", logger.Fields{
		"tenant_id": id.String(),
		"operation": "repository.Delete",
	})

	// Use SQLC-generated BulkSoftDeleteTenants function (RLS will be applied)
	err := r.store.SoftDeleteTenant(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to delete tenant")

		logger.ErrorContext(ctx, "Failed to soft delete tenant", logger.Fields{
			"tenant_id": id.String(),
			"error":     err.Error(),
			"operation": "repository.Delete",
		})
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	logger.InfoContext(ctx, "Successfully soft deleted tenant", logger.Fields{
		"tenant_id": id.String(),
		"operation": "repository.Delete",
	})

	return nil
}

func (r *repository) GetTenantConfiguration(ctx context.Context) (*db.TenantConfiguration, error) {
	ctx, span := r.tracer.StartSpan(ctx, "tenant.repository.GetTenantConfiguration")
	defer span.End()

	config, err := r.store.GetTenantConfiguration(ctx)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get tenant configuration: %w", err)
	}

	return config, nil
}

// List implements Repository.List
// Lists tenants with RLS enforcement and pagination
func (r *repository) List(ctx context.Context, offset, limit int) ([]*Tenant, error) {
	ctx, span := r.tracer.StartSpan(ctx, "tenant.repository.List")
	defer span.End()

	span.SetAttributes(
		attribute.Int("pagination.offset", offset),
		attribute.Int("pagination.limit", limit),
	)

	logger.DebugContext(ctx, "Listing tenants with pagination", logger.Fields{
		"offset":    offset,
		"limit":     limit,
		"operation": "repository.List",
	})

	// Use SQLC-generated GetActiveTenants function (RLS will be applied)
	// TODO: Add proper pagination support in SQL queries for better performance
	sqlcTenants, err := r.store.GetActiveTenants(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to list tenants")

		logger.ErrorContext(ctx, "Failed to list tenants", logger.Fields{
			"offset":    offset,
			"limit":     limit,
			"error":     err.Error(),
			"operation": "repository.List",
		})
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}

	// Apply manual pagination (ideally this should be done in SQL)
	start := offset
	end := offset + limit
	if start > len(sqlcTenants) {
		span.SetAttributes(attribute.Int("result.count", 0))
		return []*Tenant{}, nil
	}
	if end > len(sqlcTenants) {
		end = len(sqlcTenants)
	}

	paginatedSQLCTenants := sqlcTenants[start:end]

	var tenants []*Tenant
	for _, sqlcTenant := range paginatedSQLCTenants {
		tenant, err := FromSQLCTenant(sqlcTenant)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "Conversion error")
			return nil, fmt.Errorf("failed to convert tenant: %w", err)
		}
		tenants = append(tenants, tenant)
	}

	span.SetAttributes(attribute.Int("result.count", len(tenants)))

	logger.DebugContext(ctx, "Successfully listed tenants", logger.Fields{
		"offset":         offset,
		"limit":          limit,
		"returned_count": len(tenants),
		"total_count":    len(sqlcTenants),
		"operation":      "repository.List",
	})

	return tenants, nil
}

// Exists implements Repository.Exists
// Checks if tenant exists by UUID with RLS enforcement
func (r *repository) Exists(ctx context.Context, tenantID uuid.UUID) (bool, error) {
	ctx, span := r.tracer.StartSpan(ctx, "tenant.repository.Exists")
	defer span.End()

	span.SetAttributes(attribute.String("tenant.id", tenantID.String()))

	logger.DebugContext(ctx, "Checking tenant existence", logger.Fields{
		"tenant_id": tenantID.String(),
		"operation": "repository.Exists",
	})

	// Use SQLC-generated CheckTenantExists function
	exists, err := r.store.CheckTenantExists(ctx, tenantID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Database error")

		logger.ErrorContext(ctx, "Failed to check tenant existence", logger.Fields{
			"tenant_id": tenantID.String(),
			"error":     err.Error(),
			"operation": "repository.Exists",
		})
		return false, fmt.Errorf("failed to check tenant existence: %w", err)
	}

	logger.DebugContext(ctx, "Tenant existence check completed", logger.Fields{
		"tenant_id": tenantID.String(),
		"exists":    exists,
		"operation": "repository.Exists",
	})

	return exists, nil
}

// ResolveSubdomainToID implements Repository.ResolveSubdomainToID
// Resolves subdomain to tenant UUID (does not enforce RLS)
func (r *repository) ResolveSubdomainToID(ctx context.Context, subdomain string) (uuid.UUID, error) {
	ctx, span := r.tracer.StartSpan(ctx, "tenant.repository.ResolveSubdomainToID")
	defer span.End()

	span.SetAttributes(attribute.String("tenant.subdomain", subdomain))

	logger.DebugContext(ctx, "Resolving subdomain to tenant ID", logger.Fields{
		"subdomain": subdomain,
		"operation": "repository.ResolveSubdomainToID",
	})

	tenant, err := r.GetBySubdomain(ctx, subdomain)
	if err != nil {
		return uuid.Nil, err
	}

	span.SetAttributes(attribute.String("tenant.resolved_id", tenant.ID.String()))

	logger.DebugContext(ctx, "Successfully resolved subdomain to tenant ID", logger.Fields{
		"subdomain": subdomain,
		"tenant_id": tenant.ID.String(),
		"operation": "repository.ResolveSubdomainToID",
	})

	return tenant.ID, nil
}

// WithTenantContext implements Repository.WithTenantContext
// Executes a function with tenant context set using store's WithTenant
func (r *repository) WithTenantContext(ctx context.Context, tenantID uuid.UUID, fn func(context.Context) error) error {
	ctx, span := r.tracer.StartSpan(ctx, "tenant.repository.WithTenantContext")
	defer span.End()

	span.SetAttributes(attribute.String("tenant.id", tenantID.String()))

	logger.DebugContext(ctx, "Executing operation with tenant context", logger.Fields{
		"tenant_id": tenantID.String(),
		"operation": "repository.WithTenantContext",
	})

	// Use store's WithTenant method to set tenant context and execute function
	return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, store db.Store) error {
		return fn(ctx)
	})
}

// GetCurrentTenant implements Repository.GetCurrentTenant
// Gets the current tenant from the database context (for RLS)
func (r *repository) GetCurrentTenant(ctx context.Context) (*Tenant, error) {
	ctx, span := r.tracer.StartSpan(ctx, "tenant.repository.GetCurrentTenant")
	defer span.End()

	logger.DebugContext(ctx, "Getting current tenant from context", logger.Fields{
		"operation": "repository.GetCurrentTenant",
	})

	// Use SQLC-generated function to get current tenant
	// This leverages the current_tenant_id() function and RLS
	sqlcTenant, err := r.store.GetCurrentTenant(ctx)
	if err != nil {
		span.RecordError(err)

		if err.Error() == "no rows in result set" {
			span.SetStatus(codes.Error, "No current tenant in context")
			logger.WarnContext(ctx, "No current tenant in context", logger.Fields{
				"operation": "repository.GetCurrentTenant",
			})
			return nil, errors.ErrTenantIDNotInContext
		}

		span.SetStatus(codes.Error, "Database error")
		logger.ErrorContext(ctx, "Failed to get current tenant", logger.Fields{
			"error":     err.Error(),
			"operation": "repository.GetCurrentTenant",
		})
		return nil, fmt.Errorf("failed to get current tenant: %w", err)
	}

	tenant, err := FromSQLCTenant(sqlcTenant)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Conversion error")
		return nil, fmt.Errorf("failed to convert current tenant: %w", err)
	}

	span.SetAttributes(attribute.String("tenant.current_id", tenant.ID.String()))

	logger.DebugContext(ctx, "Successfully retrieved current tenant", logger.Fields{
		"tenant_id":   tenant.ID.String(),
		"tenant_name": tenant.Name,
		"operation":   "repository.GetCurrentTenant",
	})

	return tenant, nil
}

// ValidateCurrentTenant implements Repository.ValidateCurrentTenant
// Validates that current tenant context exists and is valid
func (r *repository) ValidateCurrentTenant(ctx context.Context) error {
	ctx, span := r.tracer.StartSpan(ctx, "tenant.repository.ValidateCurrentTenant")
	defer span.End()

	logger.DebugContext(ctx, "Validating current tenant context", logger.Fields{
		"operation": "repository.ValidateCurrentTenant",
	})

	// Use SQLC-generated function to check if current tenant exists
	exists, err := r.store.CheckCurrentTenantExists(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Database error")

		logger.ErrorContext(ctx, "Failed to validate current tenant", logger.Fields{
			"error":     err.Error(),
			"operation": "repository.ValidateCurrentTenant",
		})
		return fmt.Errorf("failed to validate current tenant: %w", err)
	}

	if !exists {
		span.SetStatus(codes.Error, "Invalid tenant context")

		logger.WarnContext(ctx, "Current tenant context is invalid", logger.Fields{
			"operation": "repository.ValidateCurrentTenant",
		})
		return errors.ErrTenantIDNotInContext
	}

	logger.DebugContext(ctx, "Current tenant context is valid", logger.Fields{
		"operation": "repository.ValidateCurrentTenant",
	})

	return nil
}

// SetTenant implements Repository.SetTenant
// Sets the tenant context in the database session
func (r *repository) SetTenant(ctx context.Context, tenantID uuid.UUID) error {
	ctx, span := r.tracer.StartSpan(ctx, "tenant.repository.SetTenant")
	defer span.End()

	span.SetAttributes(attribute.String("tenant.id", tenantID.String()))

	logger.DebugContext(ctx, "Setting tenant context in database session", logger.Fields{
		"tenant_id": tenantID.String(),
		"operation": "repository.SetTenant",
	})

	// Use store's SetTenantContext method
	err := r.store.SetTenantContext(ctx, tenantID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to set tenant context")

		logger.ErrorContext(ctx, "Failed to set tenant context", logger.Fields{
			"tenant_id": tenantID.String(),
			"error":     err.Error(),
			"operation": "repository.SetTenant",
		})
		return fmt.Errorf("failed to set tenant context: %w", err)
	}

	logger.DebugContext(ctx, "Successfully set tenant context", logger.Fields{
		"tenant_id": tenantID.String(),
		"operation": "repository.SetTenant",
	})

	return nil
}

// GetTenant implements Repository.GetTenant
// Gets the current tenant ID from the database session context
func (r *repository) GetTenant(ctx context.Context) (uuid.UUID, error) {
	ctx, span := r.tracer.StartSpan(ctx, "tenant.repository.GetTenant")
	defer span.End()

	logger.DebugContext(ctx, "Getting current tenant ID from database session", logger.Fields{
		"operation": "repository.GetTenant",
	})

	// Use SQLC-generated GetCurrentTenantID function
	tenantID, err := r.store.GetCurrentTenantID(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get current tenant ID")

		logger.ErrorContext(ctx, "Failed to get current tenant ID", logger.Fields{
			"error":     err.Error(),
			"operation": "repository.GetTenant",
		})
		return uuid.Nil, fmt.Errorf("failed to get current tenant ID: %w", err)
	}

	// Handle nil case (no tenant context set) - check if UUID is nil
	if tenantID == uuid.Nil {
		span.SetStatus(codes.Error, "No tenant context set")

		logger.WarnContext(ctx, "No tenant context set in database session", logger.Fields{
			"operation": "repository.GetTenant",
		})
		return uuid.Nil, fmt.Errorf("no tenant context set in database session")
	}

	span.SetAttributes(attribute.String("tenant.current_id", tenantID.String()))

	logger.DebugContext(ctx, "Successfully retrieved current tenant ID", logger.Fields{
		"tenant_id": tenantID.String(),
		"operation": "repository.GetTenant",
	})

	return tenantID, nil
}

// ResetTenant implements Repository.ResetTenant
// Resets/clears the tenant context in the database session
func (r *repository) ResetTenant(ctx context.Context) error {
	ctx, span := r.tracer.StartSpan(ctx, "tenant.repository.ResetTenant")
	defer span.End()

	logger.DebugContext(ctx, "Resetting tenant context in database session", logger.Fields{
		"operation": "repository.ResetTenant",
	})

	// Use SQLC-generated ResetTenantContext function
	err := r.store.ResetTenantContext(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to reset tenant context")

		logger.ErrorContext(ctx, "Failed to reset tenant context", logger.Fields{
			"error":     err.Error(),
			"operation": "repository.ResetTenant",
		})
		return fmt.Errorf("failed to reset tenant context: %w", err)
	}

	logger.DebugContext(ctx, "Successfully reset tenant context", logger.Fields{
		"operation": "repository.ResetTenant",
	})

	return nil
}

// ProvisionTenant implements Repository.ProvisionTenant
// This method uses a transaction to ensure that tenant creation and its
// initial configuration are atomic.
func (r *repository) ProvisionTenant(ctx context.Context, req ProvisionTenantRequest) (*ProvisionedTenantInfo, error) {
	ctx, span := r.tracer.StartSpan(ctx, "tenant.repository.ProvisionTenant")
	defer span.End()

	// Convert *string to string for SQLC parameters, handling nil
	subdomain := ""
	if req.Subdomain != nil {
		subdomain = *req.Subdomain
	}
	industry := ""
	if req.Industry != nil {
		industry = *req.Industry
	}
	companySize := "small" // Default value
	if req.CompanySize != nil {
		companySize = *req.CompanySize
	}

	// Call the SQLC-generated ProvisionTenant function
	provisionedID, err := r.store.ProvisionTenant(ctx, db.ProvisionTenantParams{
		PName:         req.Name,
		PEmail:        req.Email,
		PSubdomain:    subdomain,
		PIndustry:     industry,
		PCompanySize:  companySize,
		PCurrencyCode: req.CurrencyCode,
		PTimezone:     "UTC",        // Default timezone
		PSettings:     []byte("{}"), // Default settings
	})

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to provision tenant via DB function")
		return nil, fmt.Errorf("failed to provision tenant: %w", err)
	}

	return &ProvisionedTenantInfo{
		TenantID:  provisionedID, // Use provisionedID directly
		Slug:      "",            // Set to empty string
		Subdomain: nil,           // Set to nil
		Status:    "pending",     // Set to default status
	}, nil
}
