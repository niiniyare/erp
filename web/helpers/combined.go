package helpers

import (
	"context"
	"sync"

	"github.com/google/uuid"

	"github.com/niiniyare/erp/internal/core/featureflag"
	"github.com/niiniyare/erp/internal/core/iam"
	"github.com/niiniyare/erp/internal/shared/logger"
)

// CombinedHelper provides unified access to both permissions and feature flags
type CombinedHelper struct {
	permissionHelper  *PermissionHelper
	featureFlagHelper *FeatureFlagHelper
	logger            logger.Logger
	mu                sync.RWMutex
	config            *CombinedHelperConfig
}

// CombinedHelperConfig holds configuration for the combined helper
type CombinedHelperConfig struct {
	CheckPermissionsFirst bool // If true, check permissions before feature flags
	FailFast              bool // If true, stop on first failure
	EnableAuditLog        bool
}

// DefaultCombinedHelperConfig returns default configuration
func DefaultCombinedHelperConfig() *CombinedHelperConfig {
	return &CombinedHelperConfig{
		CheckPermissionsFirst: true,
		FailFast:              true,
		EnableAuditLog:        true,
	}
}

// NewCombinedHelper creates a new combined helper instance
func NewCombinedHelper(iamService iam.Service, featureFlagService featureflag.Service) *CombinedHelper {
	return &CombinedHelper{
		permissionHelper:  NewPermissionHelper(iamService),
		featureFlagHelper: NewFeatureFlagHelper(featureFlagService),
		logger:            nil,
		config:            DefaultCombinedHelperConfig(),
	}
}

// NewCombinedHelperWithLogger creates a combined helper with logger
func NewCombinedHelperWithLogger(iamService iam.Service, featureFlagService featureflag.Service, log logger.Logger) *CombinedHelper {
	return &CombinedHelper{
		permissionHelper:  NewPermissionHelperWithLogger(iamService, log),
		featureFlagHelper: NewFeatureFlagHelperWithLogger(featureFlagService, log),
		logger:            log,
		config:            DefaultCombinedHelperConfig(),
	}
}

// NewCombinedHelperWithConfig creates a combined helper with custom config
func NewCombinedHelperWithConfig(iamService iam.Service, featureFlagService featureflag.Service, log logger.Logger, config *CombinedHelperConfig) *CombinedHelper {
	if config == nil {
		config = DefaultCombinedHelperConfig()
	}
	return &CombinedHelper{
		permissionHelper:  NewPermissionHelperWithLogger(iamService, log),
		featureFlagHelper: NewFeatureFlagHelperWithLogger(featureFlagService, log),
		logger:            log,
		config:            config,
	}
}

// SetLogger sets the logger for both helpers
func (h *CombinedHelper) SetLogger(log logger.Logger) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.logger = log
	h.permissionHelper.SetLogger(log)
	h.featureFlagHelper.SetLogger(log)
}

// GetConfig returns a copy of the current configuration
func (h *CombinedHelper) GetConfig() CombinedHelperConfig {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return *h.config
}

// UpdateConfig updates the configuration
func (h *CombinedHelper) UpdateConfig(config *CombinedHelperConfig) {
	if config == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.config = config

	h.logDebug("Combined helper config updated", logger.Fields{
		"check_permissions_first": config.CheckPermissionsFirst,
		"fail_fast":               config.FailFast,
		"enable_audit_log":        config.EnableAuditLog,
	})
}

// Logging helpers
func (h *CombinedHelper) logDebug(msg string, fields ...logger.Fields) {
	if h.logger != nil {
		h.logger.Debug(msg, fields...)
	}
}

func (h *CombinedHelper) logInfo(msg string, fields ...logger.Fields) {
	if h.logger != nil {
		h.logger.Info(msg, fields...)
	}
}

func (h *CombinedHelper) logWarn(msg string, fields ...logger.Fields) {
	if h.logger != nil {
		h.logger.Warn(msg, fields...)
	}
}

func (h *CombinedHelper) logError(msg string, fields ...logger.Fields) {
	if h.logger != nil {
		h.logger.Error(msg, fields...)
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

// Validate checks if the combined access context is valid
func (c *CombinedAccessContext) Validate() error {
	if c.PermissionContext != nil {
		if err := c.PermissionContext.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// CombinedAccessContextBuilder provides a fluent API for building combined contexts
type CombinedAccessContextBuilder struct {
	context CombinedAccessContext
}

// NewCombinedAccessContext creates a new combined access context builder
func NewCombinedAccessContext() *CombinedAccessContextBuilder {
	return &CombinedAccessContextBuilder{}
}

func (b *CombinedAccessContextBuilder) WithPermissionContext(permCtx *PermissionContext) *CombinedAccessContextBuilder {
	b.context.PermissionContext = permCtx
	return b
}

func (b *CombinedAccessContextBuilder) WithFeatureFlag(flagName string) *CombinedAccessContextBuilder {
	b.context.FeatureFlags = append(b.context.FeatureFlags, flagName)
	return b
}

func (b *CombinedAccessContextBuilder) WithFeatureFlags(flags []string) *CombinedAccessContextBuilder {
	b.context.FeatureFlags = flags
	return b
}

func (b *CombinedAccessContextBuilder) RequireAllFlags() *CombinedAccessContextBuilder {
	b.context.RequireAllFlags = true
	return b
}

func (b *CombinedAccessContextBuilder) RequireAnyFlag() *CombinedAccessContextBuilder {
	b.context.RequireAllFlags = false
	return b
}

func (b *CombinedAccessContextBuilder) Build() *CombinedAccessContext {
	ctx := b.context
	return &ctx
}

// HasAccessWithFeatureFlags checks both permissions and feature flags
func (h *CombinedHelper) HasAccessWithFeatureFlags(ctx context.Context, accessCtx *CombinedAccessContext) bool {
	if accessCtx == nil {
		h.logWarn("HasAccessWithFeatureFlags called with nil context")
		return false
	}

	h.mu.RLock()
	checkPermissionsFirst := h.config.CheckPermissionsFirst
	failFast := h.config.FailFast
	enableAudit := h.config.EnableAuditLog
	h.mu.RUnlock()

	var hasPermission, hasFeatureFlag bool

	// Check in configured order
	if checkPermissionsFirst {
		// Check permissions first
		if accessCtx.PermissionContext != nil {
			hasPermission = h.permissionHelper.HasPermission(ctx, accessCtx.PermissionContext)
			if failFast && !hasPermission {
				h.logAccessDecision(ctx, accessCtx, false, "permission_denied", enableAudit)
				return false
			}
		} else {
			hasPermission = true // No permission check needed
		}

		// Check feature flags
		if len(accessCtx.FeatureFlags) > 0 {
			if accessCtx.RequireAllFlags {
				hasFeatureFlag = h.featureFlagHelper.HasAllFeaturesEnabled(ctx, accessCtx.FeatureFlags)
			} else {
				hasFeatureFlag = h.featureFlagHelper.HasAnyFeatureEnabled(ctx, accessCtx.FeatureFlags)
			}
		} else {
			hasFeatureFlag = true // No feature flag check needed
		}
	} else {
		// Check feature flags first
		if len(accessCtx.FeatureFlags) > 0 {
			if accessCtx.RequireAllFlags {
				hasFeatureFlag = h.featureFlagHelper.HasAllFeaturesEnabled(ctx, accessCtx.FeatureFlags)
			} else {
				hasFeatureFlag = h.featureFlagHelper.HasAnyFeatureEnabled(ctx, accessCtx.FeatureFlags)
			}
			if failFast && !hasFeatureFlag {
				h.logAccessDecision(ctx, accessCtx, false, "feature_flag_disabled", enableAudit)
				return false
			}
		} else {
			hasFeatureFlag = true
		}

		// Check permissions
		if accessCtx.PermissionContext != nil {
			hasPermission = h.permissionHelper.HasPermission(ctx, accessCtx.PermissionContext)
		} else {
			hasPermission = true
		}
	}

	allowed := hasPermission && hasFeatureFlag
	h.logAccessDecision(ctx, accessCtx, allowed, "", enableAudit)
	return allowed
}

// logAccessDecision logs the access decision
func (h *CombinedHelper) logAccessDecision(ctx context.Context, accessCtx *CombinedAccessContext, allowed bool, reason string, enableAudit bool) {
	if !enableAudit {
		return
	}

	fields := logger.Fields{
		"allowed": allowed,
		"reason":  reason,
	}

	if accessCtx.PermissionContext != nil {
		fields["resource_type"] = accessCtx.PermissionContext.ResourceType
		fields["action"] = accessCtx.PermissionContext.Action
		fields["user_id"] = accessCtx.PermissionContext.UserID.String()
	}

	if len(accessCtx.FeatureFlags) > 0 {
		fields["feature_flags"] = accessCtx.FeatureFlags
		fields["require_all_flags"] = accessCtx.RequireAllFlags
	}

	h.logInfo("Combined access check", fields)
}

// Combined utility functions

// CanAccessFeature checks if user has both permission and feature flag enabled
func (h *CombinedHelper) CanAccessFeature(ctx context.Context, resourceType, action string, resourceID *uuid.UUID, featureFlag string) bool {
	userID, entityID, err := h.permissionHelper.GetUserFromContext(ctx)
	if err != nil {
		h.logWarn("Failed to get user from context", logger.Fields{
			"error": err.Error(),
		})
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
		h.logDebug("Permission check failed", logger.Fields{
			"user_id":       userID.String(),
			"resource_type": resourceType,
			"action":        action,
		})
		return false
	}

	// Check feature flag
	enabled := h.featureFlagHelper.IsFeatureEnabled(ctx, featureFlag)
	if !enabled {
		h.logDebug("Feature flag disabled", logger.Fields{
			"user_id":   userID.String(),
			"flag_name": featureFlag,
		})
	}

	return enabled
}

// CanAccessModuleWithPermission checks module access with specific permission
func (h *CombinedHelper) CanAccessModuleWithPermission(ctx context.Context, moduleName, action string) bool {
	// Check if module is enabled via feature flag
	if !h.featureFlagHelper.IsModuleEnabled(ctx, moduleName) {
		h.logDebug("Module disabled", logger.Fields{
			"module_name": moduleName,
		})
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
		h.logDebug("UI feature disabled", logger.Fields{
			"feature_name": featureName,
		})
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
		h.logWarn("Failed to get user from context for access status", logger.Fields{
			"error": err.Error(),
		})
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

	isSystemAdmin := h.permissionHelper.IsSystemAdmin(ctx, userID)
	isTenantAdmin := false
	if entityID != nil {
		isTenantAdmin = h.permissionHelper.IsTenantAdmin(ctx, userID, *entityID)
	}
	isBetaUser := h.featureFlagHelper.IsBetaFeatureEnabled(ctx)

	h.logDebug("Generated access status", logger.Fields{
		"user_id":         userID.String(),
		"resource_type":   resourceType,
		"is_system_admin": isSystemAdmin,
		"is_tenant_admin": isTenantAdmin,
		"is_beta_user":    isBetaUser,
	})

	return &AccessStatus{
		UserID:        userID,
		EntityID:      entityID,
		Permissions:   permissionStatus,
		FeatureFlags:  featureStatus,
		IsSystemAdmin: isSystemAdmin,
		IsTenantAdmin: isTenantAdmin,
		IsBetaUser:    isBetaUser,
	}
}

// GetDetailedAccessStatus returns more detailed access status
func (h *CombinedHelper) GetDetailedAccessStatus(ctx context.Context, resourceType string, resourceID *uuid.UUID, featureFlagNames []string) *DetailedAccessStatus {
	basicStatus := h.GetAccessStatus(ctx, resourceType, resourceID)

	// Get additional feature flags
	additionalFlags := make(map[string]bool)
	for _, flag := range featureFlagNames {
		additionalFlags[flag] = h.featureFlagHelper.IsFeatureEnabled(ctx, flag)
	}

	return &DetailedAccessStatus{
		AccessStatus:           basicStatus,
		AdditionalFeatureFlags: additionalFlags,
		ResourceType:           resourceType,
		ResourceID:             resourceID,
	}
}

// BatchAccessCheck checks access for multiple resources
func (h *CombinedHelper) BatchAccessCheck(ctx context.Context, resourceType string, resourceIDs []uuid.UUID, action string, featureFlag string) map[uuid.UUID]bool {
	userID, entityID, err := h.permissionHelper.GetUserFromContext(ctx)
	if err != nil {
		h.logError("Failed to get user from context for batch check", logger.Fields{
			"error": err.Error(),
		})
		return nil
	}

	// Check feature flag once
	if featureFlag != "" && !h.featureFlagHelper.IsFeatureEnabled(ctx, featureFlag) {
		h.logDebug("Feature flag disabled for batch check", logger.Fields{
			"flag_name": featureFlag,
		})
		// Return all false
		results := make(map[uuid.UUID]bool, len(resourceIDs))
		for _, id := range resourceIDs {
			results[id] = false
		}
		return results
	}

	// Check permissions for each resource
	return h.permissionHelper.CheckMultipleResources(ctx, userID, resourceType, resourceIDs, action, entityID)
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

// DetailedAccessStatus provides more detailed access information
type DetailedAccessStatus struct {
	*AccessStatus
	AdditionalFeatureFlags map[string]bool `json:"additional_feature_flags,omitempty"`
	ResourceType           string          `json:"resource_type"`
	ResourceID             *uuid.UUID      `json:"resource_id,omitempty"`
}

// HasPermission checks if any of the permissions in the status allow the action
func (a *AccessStatus) HasPermission(action string) bool {
	if a.Permissions == nil {
		return false
	}
	allowed, exists := a.Permissions["can"+action]
	return exists && allowed
}

// HasFeatureFlag checks if a feature flag is enabled
func (a *AccessStatus) HasFeatureFlag(flagName string) bool {
	if a.FeatureFlags == nil {
		return false
	}
	enabled, exists := a.FeatureFlags[flagName]
	return exists && enabled
}

// IsAdmin checks if user has any admin privileges
func (a *AccessStatus) IsAdmin() bool {
	return a.IsSystemAdmin || a.IsTenantAdmin
}
