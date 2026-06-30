package workflow

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// ── Activity input/output types ────────────────────────────────────────────────

// CreateAdminUserInput carries the parameters for CreateAdminUserActivity.
type CreateAdminUserInput struct {
	TenantID uuid.UUID `json:"tenant_id"`
	Email    string    `json:"email"`
}

// DeleteUserInput carries the parameters for DeleteUserActivity.
type DeleteUserInput struct {
	TenantID uuid.UUID `json:"tenant_id"`
	UserID   uuid.UUID `json:"user_id"`
}

// ActivateFeatureFlagsInput carries the parameters for ActivateFeatureFlagsActivity.
type ActivateFeatureFlagsInput struct {
	TenantID uuid.UUID `json:"tenant_id"`
	Plan     string    `json:"plan"`
}

// SendWelcomeEmailInput carries the parameters for SendWelcomeEmailActivity.
type SendWelcomeEmailInput struct {
	TenantID   uuid.UUID `json:"tenant_id"`
	Email      string    `json:"email"`
	TenantName string    `json:"tenant_name"`
}

// ── TenantActivities ──────────────────────────────────────────────────────────

// TenantActivities holds injected dependencies for all tenant-provisioning
// activities. Register via worker.RegisterActivity(activities.CreateTenantRecordActivity).
//
// All activity methods follow the Temporal activity signature:
//
//	func(ctx context.Context, ...) error
//
// I/O (DB writes, email sends) is performed here, NOT in the workflow function.
type TenantActivities struct {
	// TenantRepo provides Create/Update/Delete for the core tenants table.
	// TODO: inject concrete *pgstore.TenantRepository once the repo type is defined.
	TenantRepo interface {
		Create(ctx context.Context, tenantID uuid.UUID, name string) error
		Delete(ctx context.Context, tenantID uuid.UUID) error
		Activate(ctx context.Context, tenantID uuid.UUID) error
	}

	// IAMRepo provisions roles and users within a tenant.
	// TODO: inject concrete *pgstore.IAMRepository once the repo type is defined.
	IAMRepo interface {
		SeedRoles(ctx context.Context, tenantID uuid.UUID) error
		CleanupRoles(ctx context.Context, tenantID uuid.UUID) error
		CreateAdminUser(ctx context.Context, tenantID uuid.UUID, email string) (uuid.UUID, error)
		DeleteUser(ctx context.Context, tenantID, userID uuid.UUID) error
	}

	// SettingsRepo seeds per-tenant system settings.
	// TODO: inject concrete *pgstore.SettingsRepository once the repo type is defined.
	SettingsRepo interface {
		SeedDefaults(ctx context.Context, tenantID uuid.UUID, locale string) error
		Cleanup(ctx context.Context, tenantID uuid.UUID) error
	}

	// FlagRepo activates plan-level feature flags for a tenant.
	// TODO: inject concrete *pgstore.FeatureFlagRepository once the repo type is defined.
	FlagRepo interface {
		ActivatePlanFlags(ctx context.Context, tenantID uuid.UUID, plan string) error
	}

	// EmailClient sends transactional email.
	// TODO: inject concrete notifications.EmailClient once the package is defined.
	EmailClient interface {
		SendWelcome(ctx context.Context, to, tenantName string) error
	}
}

// ── Activity functions (standalone, registered by name) ───────────────────────
// These are package-level variables so the workflow can reference them as
// workflow.ExecuteActivity(ctx, CreateTenantRecordActivity, ...) without
// needing a struct receiver, and so tests can swap them easily.

// CreateTenantRecordActivity inserts the tenant row in PENDING status.
var CreateTenantRecordActivity = func(ctx context.Context, input ProvisioningInput) error {
	// TODO: inject TenantActivities via activity context or use a module-level var.
	// Placeholder: replace with real DB call via injected TenantRepo.
	return fmt.Errorf("CreateTenantRecordActivity: not implemented — wire TenantRepo")
}

// DeleteTenantRecordActivity removes the tenant row (saga compensation for step 1).
var DeleteTenantRecordActivity = func(ctx context.Context, tenantID uuid.UUID) error {
	// TODO: implement via injected TenantRepo.Delete
	return fmt.Errorf("DeleteTenantRecordActivity: not implemented — wire TenantRepo")
}

// SeedRolesActivity seeds the standard role set for a new tenant.
var SeedRolesActivity = func(ctx context.Context, tenantID uuid.UUID) error {
	// TODO: implement via injected IAMRepo.SeedRoles
	return fmt.Errorf("SeedRolesActivity: not implemented — wire IAMRepo")
}

// CleanupRolesActivity removes seeded roles (saga compensation for step 2).
var CleanupRolesActivity = func(ctx context.Context, tenantID uuid.UUID) error {
	// TODO: implement via injected IAMRepo.CleanupRoles
	return fmt.Errorf("CleanupRolesActivity: not implemented — wire IAMRepo")
}

// SeedSystemSettingsActivity inserts per-tenant system settings with locale defaults.
var SeedSystemSettingsActivity = func(ctx context.Context, tenantID uuid.UUID, locale string) error {
	// TODO: implement via injected SettingsRepo.SeedDefaults
	return fmt.Errorf("SeedSystemSettingsActivity: not implemented — wire SettingsRepo")
}

// CleanupSettingsActivity removes seeded settings (saga compensation for step 3).
var CleanupSettingsActivity = func(ctx context.Context, tenantID uuid.UUID) error {
	// TODO: implement via injected SettingsRepo.Cleanup
	return fmt.Errorf("CleanupSettingsActivity: not implemented — wire SettingsRepo")
}

// CreateAdminUserActivity creates the initial admin user and returns its UUID.
var CreateAdminUserActivity = func(ctx context.Context, input CreateAdminUserInput) (uuid.UUID, error) {
	// TODO: implement via injected IAMRepo.CreateAdminUser
	return uuid.Nil, fmt.Errorf("CreateAdminUserActivity: not implemented — wire IAMRepo")
}

// DeleteUserActivity removes a user (saga compensation for step 4).
var DeleteUserActivity = func(ctx context.Context, input DeleteUserInput) error {
	// TODO: implement via injected IAMRepo.DeleteUser
	return fmt.Errorf("DeleteUserActivity: not implemented — wire IAMRepo")
}

// ActivateFeatureFlagsActivity enables plan-level feature flags for the tenant.
var ActivateFeatureFlagsActivity = func(ctx context.Context, input ActivateFeatureFlagsInput) error {
	// TODO: implement via injected FlagRepo.ActivatePlanFlags
	return fmt.Errorf("ActivateFeatureFlagsActivity: not implemented — wire FlagRepo")
}

// ActivateTenantActivity transitions the tenant from PENDING → ACTIVE.
var ActivateTenantActivity = func(ctx context.Context, tenantID uuid.UUID) error {
	// TODO: implement via injected TenantRepo.Activate + StatusService.SetStatus for cache invalidation
	return fmt.Errorf("ActivateTenantActivity: not implemented — wire TenantRepo + StatusService")
}

// SendWelcomeEmailActivity sends the onboarding email to the new admin.
// Non-critical: failure is logged by the workflow but does not trigger compensation.
var SendWelcomeEmailActivity = func(ctx context.Context, input SendWelcomeEmailInput) error {
	// TODO: implement via injected EmailClient.SendWelcome
	return fmt.Errorf("SendWelcomeEmailActivity: not implemented — wire EmailClient")
}
