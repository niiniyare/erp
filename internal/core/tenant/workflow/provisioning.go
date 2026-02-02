package workflow

import (
	"fmt"
	"time"

	"github.com/niiniyare/erp/internal/core/tenant/domain"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	TaskQueue              = "tenant.provisioning"
	ProvisioningWorkflowID = "TenantProvisioningWorkflow"
)

// ProvisioningWorkflow orchestrates tenant provisioning as a saga.
// Steps: create tenant → create config → init usage → activate.
// On failure, compensating actions soft-delete the tenant.
func ProvisioningWorkflow(ctx workflow.Context, input domain.ProvisioningInput) (*domain.ProvisioningResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting tenant provisioning", "name", input.Name)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Step 1: Provision tenant (creates tenant + schema via DB function)
	var result domain.ProvisioningResult
	if err := workflow.ExecuteActivity(ctx, "ProvisionTenantActivity", input).Get(ctx, &result); err != nil {
		return nil, fmt.Errorf("provision tenant failed: %w", err)
	}

	// Step 2: Create default configuration
	if err := workflow.ExecuteActivity(ctx, "CreateDefaultConfigActivity", result.TenantID).Get(ctx, nil); err != nil {
		logger.Error("Create config failed, compensating", "error", err)
		_ = workflow.ExecuteActivity(ctx, "CleanupTenantActivity", result.TenantID).Get(ctx, nil)
		return nil, fmt.Errorf("create config failed: %w", err)
	}

	// Step 3: Initialize usage stats
	if err := workflow.ExecuteActivity(ctx, "InitUsageActivity", result.TenantID).Get(ctx, nil); err != nil {
		logger.Error("Init usage failed, compensating", "error", err)
		_ = workflow.ExecuteActivity(ctx, "CleanupTenantActivity", result.TenantID).Get(ctx, nil)
		return nil, fmt.Errorf("init usage failed: %w", err)
	}

	// Step 4: Activate tenant
	if err := workflow.ExecuteActivity(ctx, "ActivateTenantActivity", result.TenantID).Get(ctx, nil); err != nil {
		logger.Error("Activate tenant failed, compensating", "error", err)
		_ = workflow.ExecuteActivity(ctx, "CleanupTenantActivity", result.TenantID).Get(ctx, nil)
		return nil, fmt.Errorf("activate tenant failed: %w", err)
	}

	// Step 5: Send welcome notification (best-effort)
	if err := workflow.ExecuteActivity(ctx, "SendWelcomeNotificationActivity", result.TenantID).Get(ctx, nil); err != nil {
		logger.Warn("Welcome notification failed (non-fatal)", "error", err)
	}

	logger.Info("Tenant provisioning complete", "tenant_id", result.TenantID.String())
	return &result, nil
}
