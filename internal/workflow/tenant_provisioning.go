package workflow

import (
	"context"
	"time"

	"github.com/niiniyare/erp/internal/domain/tenant"
	"go.temporal.io/sdk/workflow"
)

func TenantProvisioningWorkflow(ctx workflow.Context, params tenant.ProvisionParams) error {
	options := workflow.ActivityOptions{
		StartToCloseTimeout: 15 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, options)

	// Execute provisioning activities in sequence
	var tenantID int
	if err := workflow.ExecuteActivity(ctx, activities.CreateTenantRecord, params).Get(ctx, &tenantID); err != nil {
		return err
	}

	if err := workflow.ExecuteActivity(ctx, activities.CreateDefaultEntity, tenantID).Get(ctx, nil); err != nil {
		return err
	}

	if err := workflow.ExecuteActivity(ctx, activities.CreateAdminUser, tenantID).Get(ctx, nil); err != nil {
		return err
	}

	if err := workflow.ExecuteActivity(ctx, activities.ConfigureDefaults, tenantID).Get(ctx, nil); err != nil {
		return err
	}

	return workflow.ExecuteActivity(ctx, activities.SendWelcomeEmail, tenantID).Get(ctx, nil)
}
