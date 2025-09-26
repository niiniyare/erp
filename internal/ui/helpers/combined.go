package helpers

import (
	"context"

	"github.com/google/uuid"

	"github.com/niiniyare/erp/internal/core/featureflag"
	"github.com/niiniyare/erp/internal/core/iam"
)

// CombinedHelper provides unified access to both permissions and feature flags
type CombinedHelper struct {
	permissionHelper  *PermissionHelper
	featureFlagHelper *FeatureFlagHelper
}

// NewCombinedHelper creates a new combined helper instance
func NewCombinedHelper(iamService iam.Service, featureFlagService featureflag.Service) *CombinedHelper {
	return &CombinedHelper{
		permissionHelper:  NewPermissionHelper(iamService),
		featureFlagHelper: NewFeatureFlagHelper(featureFlagService),
	}
}

// Permissions returns the permission helper
func (h *CombinedHelper) Permissions() *PermissionHelper {
	return h.permissionHelper
}

// FeatureFlags returns the feature flag helper
func (h *CombinedHelper) FeatureFlags() *FeatureFlagHelper {
	return h.featureFlagHelper
}

// CombinedAccessContext holds both permission and feature flag context
type CombinedAccessContext struct {
	PermissionContext *PermissionContext
	FeatureFlags      []string
	RequireAllFlags   bool
}

// HasAccessWithFeatureFlags checks both permissions and feature flags
func (h *CombinedHelper) HasAccessWithFeatureFlags(ctx context.Context, accessCtx *CombinedAccessContext) bool {
	// Check permissions first
	if accessCtx.PermissionContext != nil {
		if !h.permissionHelper.HasPermission(ctx, accessCtx.PermissionContext) {
			return false
		}
	}

	// Check feature flags
	if len(accessCtx.FeatureFlags) > 0 {
		if accessCtx.RequireAllFlags {
			return h.featureFlagHelper.HasAllFeaturesEnabled(ctx, accessCtx.FeatureFlags)
		} else {
			return h.featureFlagHelper.HasAnyFeatureEnabled(ctx, accessCtx.FeatureFlags)
		}
	}

	return true
}

// Combined utility functions

// CanAccessFeature checks if user has both permission and feature flag enabled
func (h *CombinedHelper) CanAccessFeature(ctx context.Context, resourceType, action string, resourceID *uuid.UUID, featureFlag string) bool {
	userID, entityID, err := h.permissionHelper.GetUserFromContext(ctx)
	if err != nil {
		return false
	}

	// Check permission
	hasPermission := h.permissionHelper.HasPermission(ctx, &PermissionContext{
		UserID:       userID,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Action:       action,
		EntityID:     entityID,
	})

	if !hasPermission {
		return false
	}

	// Check feature flag
	return h.featureFlagHelper.IsFeatureEnabled(ctx, featureFlag)
}

// CanAccessModuleWithPermission checks module access with specific permission
func (h *CombinedHelper) CanAccessModuleWithPermission(ctx context.Context, moduleName, action string) bool {
	// Check if module is enabled via feature flag
	if !h.featureFlagHelper.IsModuleEnabled(ctx, moduleName) {
		return false
	}

	// Check permission for the module
	userID, entityID, err := h.permissionHelper.GetUserFromContext(ctx)
	if err != nil {
		return false
	}

	return h.permissionHelper.HasPermission(ctx, &PermissionContext{
		UserID:       userID,
		ResourceType: moduleName,
		Action:       action,
		EntityID:     entityID,
	})
}

// CanAccessUIFeatureWithPermission checks UI feature access with permission
func (h *CombinedHelper) CanAccessUIFeatureWithPermission(ctx context.Context, featureName, resourceType, action string) bool {
	// Check if UI feature is enabled
	if !h.featureFlagHelper.IsUIFeatureEnabled(ctx, featureName) {
		return false
	}

	// Check permission
	userID, entityID, err := h.permissionHelper.GetUserFromContext(ctx)
	if err != nil {
		return false
	}

	return h.permissionHelper.HasPermission(ctx, &PermissionContext{
		UserID:       userID,
		ResourceType: resourceType,
		Action:       action,
		EntityID:     entityID,
	})
}

// GetAccessStatus returns comprehensive access status for templates
func (h *CombinedHelper) GetAccessStatus(ctx context.Context, resourceType string, resourceID *uuid.UUID) *AccessStatus {
	userID, entityID, err := h.permissionHelper.GetUserFromContext(ctx)
	if err != nil {
		return &AccessStatus{}
	}

	// Get permission status by checking common actions
	permissionStatus := map[string]bool{
		"canRead":   h.permissionHelper.CanRead(ctx, userID, resourceType, resourceID, entityID),
		"canWrite":  h.permissionHelper.CanWrite(ctx, userID, resourceType, resourceID, entityID),
		"canCreate": h.permissionHelper.CanCreate(ctx, userID, resourceType, entityID),
		"canDelete": h.permissionHelper.CanDelete(ctx, userID, resourceType, resourceID, entityID),
		"canManage": h.permissionHelper.CanManage(ctx, userID, resourceType, resourceID, entityID),
	}

	// Get common feature flags
	commonFlags := []string{
		FeatureFlagBetaFeatures,
		FeatureFlagNewDashboard,
		FeatureFlagAdvancedReports,
		FeatureFlagBulkOperations,
	}

	featureStatus := make(map[string]bool)
	for _, flag := range commonFlags {
		featureStatus[flag] = h.featureFlagHelper.IsFeatureEnabled(ctx, flag)
	}

	return &AccessStatus{
		UserID:        userID,
		EntityID:      entityID,
		Permissions:   permissionStatus,
		FeatureFlags:  featureStatus,
		IsSystemAdmin: h.permissionHelper.IsSystemAdmin(ctx, userID),
		IsTenantAdmin: entityID != nil && h.permissionHelper.IsTenantAdmin(ctx, userID, *entityID),
		IsBetaUser:    h.featureFlagHelper.IsBetaFeatureEnabled(ctx),
	}
}

// AccessStatus provides comprehensive access information for templates
type AccessStatus struct {
	UserID        uuid.UUID       `json:"user_id"`
	EntityID      *uuid.UUID      `json:"entity_id,omitempty"`
	Permissions   map[string]bool `json:"permissions"`
	FeatureFlags  map[string]bool `json:"feature_flags"`
	IsSystemAdmin bool            `json:"is_system_admin"`
	IsTenantAdmin bool            `json:"is_tenant_admin"`
	IsBetaUser    bool            `json:"is_beta_user"`
}
