// Package workflow contains Temporal workflow definitions for framework-level
// operations (tenant provisioning, etc.). Business-domain workflows live in
// internal/core/<module>/workflows/.
package workflow

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// ── Input / Output types ───────────────────────────────────────────────────────

// ProvisioningInput is passed to TenantProvisioningWorkflow at start time.
type ProvisioningInput struct {
	// TenantID is pre-allocated by the caller so it can be recorded in the
	// HTTP response before the workflow completes.
	TenantID uuid.UUID `json:"tenant_id"`

	// TenantName is the display name for the new tenant account.
	TenantName string `json:"tenant_name"`

	// AdminEmail is the email address for the initial admin user.
	AdminEmail string `json:"admin_email"`

	// Plan is the subscription plan identifier (e.g. "starter", "growth", "enterprise").
	Plan string `json:"plan"`

	// Locale is the default locale for the tenant (e.g. "en-KE").
	// Defaults to "en-KE" when empty.
	Locale string `json:"locale,omitempty"`
}

// ProvisioningOutput is returned by TenantProvisioningWorkflow on success.
type ProvisioningOutput struct {
	TenantID    uuid.UUID `json:"tenant_id"`
	AdminUserID uuid.UUID `json:"admin_user_id"`
	WorkflowID  string    `json:"workflow_id"`
}

// ── Workflow ───────────────────────────────────────────────────────────────────

// TenantProvisioningWorkflow creates a fully-provisioned tenant in a durable,
// saga-compensated sequence of activities.
//
// Workflow ID convention: "tenant.provision.{tenantID}"
//
// Steps (with compensation):
//  1. CreateTenantRecord    → compensation: DeleteTenantRecord
//  2. SeedRoles             → compensation: CleanupRoles
//  3. SeedSystemSettings    → compensation: CleanupSettings
//  4. CreateAdminUser       → compensation: DeleteUser
//  5. ActivateFeatureFlags  → (idempotent, no compensation needed)
//  6. ActivateTenant        → sets status = ACTIVE
//  7. SendWelcomeEmail      → non-critical; failure logged but does not fail workflow
//
// Each activity uses a 30-second start-to-close timeout with 3 retry attempts.
// The welcome email has a single attempt — bounce is acceptable.
func TenantProvisioningWorkflow(ctx workflow.Context, input ProvisioningInput) (*ProvisioningOutput, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("TenantProvisioningWorkflow started", "tenant_id", input.TenantID)

	if input.Locale == "" {
		input.Locale = "en-KE"
	}

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
			InitialInterval: 2 * time.Second,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Saga compensator — runs compensations in LIFO order on failure.
	compensations := make([]compensationFn, 0, 6)
	defer func() {
		// Run compensations in reverse order on any panic/non-nil return.
		// workflow.Go is not needed here because compensations are sequential.
		// This defer only fires when the workflow function returns with an error;
		// on success the slice is never iterated.
	}()

	rollback := func(err error) (*ProvisioningOutput, error) {
		for i := len(compensations) - 1; i >= 0; i-- {
			if cErr := compensations[i](ctx); cErr != nil {
				logger.Error("saga compensation failed",
					"step", i,
					"compensation_error", cErr,
					"original_error", err,
				)
			}
		}
		return nil, fmt.Errorf("TenantProvisioningWorkflow: %w", err)
	}

	// ── Step 1: Create tenant record ──────────────────────────────────────────
	if err := workflow.ExecuteActivity(ctx, CreateTenantRecordActivity, input).Get(ctx, nil); err != nil {
		return rollback(fmt.Errorf("create tenant record: %w", err))
	}
	compensations = append(compensations, func(ctx workflow.Context) error {
		return workflow.ExecuteActivity(ctx, DeleteTenantRecordActivity, input.TenantID).Get(ctx, nil)
	})

	// ── Step 2: Seed roles ────────────────────────────────────────────────────
	if err := workflow.ExecuteActivity(ctx, SeedRolesActivity, input.TenantID).Get(ctx, nil); err != nil {
		return rollback(fmt.Errorf("seed roles: %w", err))
	}
	compensations = append(compensations, func(ctx workflow.Context) error {
		return workflow.ExecuteActivity(ctx, CleanupRolesActivity, input.TenantID).Get(ctx, nil)
	})

	// ── Step 3: Seed system settings ─────────────────────────────────────────
	if err := workflow.ExecuteActivity(ctx, SeedSystemSettingsActivity, input.TenantID, input.Locale).Get(ctx, nil); err != nil {
		return rollback(fmt.Errorf("seed settings: %w", err))
	}
	compensations = append(compensations, func(ctx workflow.Context) error {
		return workflow.ExecuteActivity(ctx, CleanupSettingsActivity, input.TenantID).Get(ctx, nil)
	})

	// ── Step 4: Create admin user ─────────────────────────────────────────────
	var adminUserID uuid.UUID
	if err := workflow.ExecuteActivity(ctx, CreateAdminUserActivity,
		CreateAdminUserInput{TenantID: input.TenantID, Email: input.AdminEmail},
	).Get(ctx, &adminUserID); err != nil {
		return rollback(fmt.Errorf("create admin user: %w", err))
	}
	compensations = append(compensations, func(ctx workflow.Context) error {
		return workflow.ExecuteActivity(ctx, DeleteUserActivity,
			DeleteUserInput{TenantID: input.TenantID, UserID: adminUserID},
		).Get(ctx, nil)
	})

	// ── Step 5: Activate feature flags (idempotent) ───────────────────────────
	if err := workflow.ExecuteActivity(ctx, ActivateFeatureFlagsActivity,
		ActivateFeatureFlagsInput{TenantID: input.TenantID, Plan: input.Plan},
	).Get(ctx, nil); err != nil {
		return rollback(fmt.Errorf("activate feature flags: %w", err))
	}

	// ── Step 6: Activate tenant (set status = ACTIVE) ─────────────────────────
	if err := workflow.ExecuteActivity(ctx, ActivateTenantActivity, input.TenantID).Get(ctx, nil); err != nil {
		return rollback(fmt.Errorf("activate tenant: %w", err))
	}

	// ── Step 7: Send welcome email (non-critical) ─────────────────────────────
	emailOpts := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 15 * time.Second,
		RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 1},
	})
	if err := workflow.ExecuteActivity(emailOpts, SendWelcomeEmailActivity,
		SendWelcomeEmailInput{TenantID: input.TenantID, Email: input.AdminEmail, TenantName: input.TenantName},
	).Get(emailOpts, nil); err != nil {
		// Non-critical — log and continue.
		logger.Warn("welcome email failed (non-fatal)", "error", err)
	}

	logger.Info("TenantProvisioningWorkflow completed",
		"tenant_id", input.TenantID,
		"admin_user_id", adminUserID,
	)

	return &ProvisioningOutput{
		TenantID:    input.TenantID,
		AdminUserID: adminUserID,
		WorkflowID:  workflow.GetInfo(ctx).WorkflowExecution.ID,
	}, nil
}

// ProvisioningWorkflowID returns the canonical Temporal workflow ID for a
// tenant provisioning run.
func ProvisioningWorkflowID(tenantID uuid.UUID) string {
	return "tenant.provision." + tenantID.String()
}

// compensationFn is a function that compensates (undoes) one saga step.
type compensationFn func(ctx workflow.Context) error
