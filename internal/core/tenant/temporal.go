package tenant

import (
	"fmt"

	"github.com/niiniyare/erp/internal/core/tenant/activities"
	"github.com/niiniyare/erp/internal/core/tenant/workflow"
	"github.com/niiniyare/erp/internal/platform/temporal"
	loggerPkg "github.com/niiniyare/erp/internal/shared/logger"
	"go.temporal.io/sdk/client"
)

// TemporalIntegration handles registration of tenant workflows and activities.
type TemporalIntegration struct {
	activities     *activities.Activities
	temporalClient client.Client
	logger         loggerPkg.Logger
}

// TemporalConfig contains configuration for tenant Temporal integration.
type TemporalConfig struct {
	Service        Service
	TemporalClient client.Client
	Logger         loggerPkg.Logger
}

// NewTemporalIntegration creates a new Temporal integration for the tenant module.
func NewTemporalIntegration(cfg TemporalConfig) (*TemporalIntegration, error) {
	if cfg.Service == nil {
		return nil, fmt.Errorf("tenant service is required")
	}
	if cfg.TemporalClient == nil {
		return nil, fmt.Errorf("temporal client is required")
	}

	// Type assert to get internal adapter
	adapter, ok := cfg.Service.(*tenantServiceAdapter)
	if !ok {
		return nil, fmt.Errorf("tenant service must be created via tenant.NewService")
	}

	acts := activities.New(activities.Deps{
		TenantService:       adapter.tenant,
		ProvisioningService: adapter.provisioning,
		Repo:                adapter.repo,
		Tracer:              adapter.tracer,
	})

	return &TemporalIntegration{
		activities:     acts,
		temporalClient: cfg.TemporalClient,
		logger:         cfg.Logger,
	}, nil
}

// RegisterWithPlatform registers tenant workflows and activities with Temporal.
func (ti *TemporalIntegration) RegisterWithPlatform(platform *temporal.Platform) error {
	if platform == nil {
		return fmt.Errorf("temporal platform is required")
	}

	registrar := platform.NewModuleRegistrar("tenant")
	if !registrar.IsEnabled() {
		ti.logger.Info("Tenant module not enabled in Temporal, skipping")
		return nil
	}

	// Register workflows
	registrar.RegisterWorkflows(map[string]any{
		"TenantProvisioningWorkflow":  workflow.ProvisioningWorkflow,
		"TenantBulkOperationWorkflow": workflow.BulkOperationWorkflow,
	})

	// Register activities
	registrar.RegisterActivities(map[string]any{
		"ProvisionTenantActivity":          ti.activities.ProvisionTenantActivity,
		"CreateDefaultConfigActivity":      ti.activities.CreateDefaultConfigActivity,
		"InitUsageActivity":                ti.activities.InitUsageActivity,
		"ActivateTenantActivity":           ti.activities.ActivateTenantActivity,
		"CleanupTenantActivity":            ti.activities.CleanupTenantActivity,
		"SendWelcomeNotificationActivity":  ti.activities.SendWelcomeNotificationActivity,
		"BulkUpdateStatusActivity":         ti.activities.BulkUpdateStatusActivity,
		"BulkSoftDeleteActivity":           ti.activities.BulkSoftDeleteActivity,
	})

	ti.logger.Info("Tenant module registered with Temporal", loggerPkg.Fields{
		"module": "tenant",
		"status": "registered",
	})

	return nil
}

// GetTemporalClient returns the Temporal client.
func (ti *TemporalIntegration) GetTemporalClient() client.Client {
	return ti.temporalClient
}
