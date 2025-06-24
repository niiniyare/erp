// internal/domain/tenant/service.go
package tenant

import (
	"context"
	"database/sql"
	"errors"

	"github.com/niiniyare/erp/internal/storage/postgres"
)

type Service struct {
	repo    Repo
	queries *postgres.Queries
	// workflow WorkflowClient
}

func NewService(repo Repository, queries *postgres.Queries, workflow WorkflowClient) *Service {
	return &Service{repo, queries, workflow}
}

func (s *Service) ProvisionTenant(ctx context.Context, params ProvisionParams) (*Tenant, error) {
	// Validate business rules
	if err := validateProvisionParams(params); err != nil {
		return nil, err
	}

	// Start provisioning workflow
	workflowID := "tenant-provisioning-" + params.Subdomain
	options := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: "tenant-provisioning",
	}
	exec, err := s.workflow.ExecuteWorkflow(ctx, options, TenantProvisioningWorkflow, params)
	if err != nil {
		return nil, err
	}

	// Return immediately while workflow runs asynchronously
	return &Tenant{
		Name:      params.Name,
		Subdomain: params.Subdomain,
		Status:    "provisioning",
		Workflow:  workflowID,
	}, nil
}

func (s *Service) SuspendTenant(ctx context.Context, tenantID int) error {
	return s.repo.WithTx(ctx, func(tx *sql.Tx) error {
		// Get current status
		t, err := s.queries.GetTenant(ctx, int32(tenantID))
		if err != nil {
			return err
		}

		if t.Status == "suspended" {
			return errors.New("tenant already suspended")
		}

		// Suspend all users
		if err := s.queries.SuspendTenantUsers(ctx, int32(tenantID)); err != nil {
			return err
		}

		// Update tenant status
		return s.queries.UpdateTenantStatus(ctx, postgres.UpdateTenantStatusParams{
			ID:     int32(tenantID),
			Status: "suspended",
		})
	})
}

// import (
//
//	"context"
//
//	"github.com/google/uuid"
//
// )
//
// Primary Ports (Driving - Inbound)
type TenantService interface {
	// Admin operations (system-wide)
	CreateTenant(ctx context.Context, req CreateTenantRequest) (*Tenant, error)
	GetTenantByID(ctx context.Context, id int32) (*Tenant, error)
	GetTenantByUUID(ctx context.Context, uuid uuid.UUID) (*Tenant, error)
	GetTenantBySubdomain(ctx context.Context, subdomain string) (*Tenant, error)
	ListTenants(ctx context.Context, filter TenantFilter) ([]*Tenant, error)
	UpdateTenant(ctx context.Context, id int32, req UpdateTenantRequest) (*Tenant, error)
	DeleteTenant(ctx context.Context, id int32) error
	GetTenantStats(ctx context.Context) (*TenantStats, error)

	// Current tenant operations (RLS-aware)
	GetCurrentTenant(ctx context.Context) (*Tenant, error)
	UpdateCurrentTenant(ctx context.Context, req UpdateTenantRequest) (*Tenant, error)
}

type TenantConfigService interface {
	// Admin operations
	CreateTenantConfiguration(ctx context.Context, tenantID int32, req CreateTenantConfigRequest) (*TenantConfiguration, error)
	GetTenantConfiguration(ctx context.Context, tenantID int32) (*TenantConfiguration, error)
	UpdateTenantConfiguration(ctx context.Context, tenantID int32, req UpdateTenantConfigRequest) (*TenantConfiguration, error)
	DeleteTenantConfiguration(ctx context.Context, tenantID int32) error

	// Current tenant operations (RLS-aware)
	GetCurrentTenantConfiguration(ctx context.Context) (*TenantConfiguration, error)
	UpdateCurrentTenantConfiguration(ctx context.Context, req UpdateTenantConfigRequest) (*TenantConfiguration, error)
	CreateCurrentTenantConfiguration(ctx context.Context, req CreateTenantConfigRequest) (*TenantConfiguration, error)

	// Feature and module checks
	CheckCurrentTenantHasFeature(ctx context.Context, feature string) (bool, error)
	CheckCurrentTenantHasModule(ctx context.Context, module string) (bool, error)
}
