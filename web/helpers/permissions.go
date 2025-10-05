package helpers

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"

	"github.com/niiniyare/erp/internal/core/iam"
	"github.com/niiniyare/erp/internal/core/iam/authz"
	"github.com/niiniyare/erp/internal/shared"
	"github.com/niiniyare/erp/internal/shared/logger"
)

// PermissionHelper provides template helper functions for ABAC permission checks
type PermissionHelper struct {
	iamService iam.Service
	logger     logger.Logger
	mu         sync.RWMutex // For thread-safe operations if needed
	config     *PermissionHelperConfig
}

// PermissionHelperConfig holds configuration for the permission helper
type PermissionHelperConfig struct {
	EnableCaching    bool
	CacheTTL         int // seconds
	EnableAuditLog   bool
	StrictValidation bool
}

// DefaultPermissionHelperConfig returns default configuration
func DefaultPermissionHelperConfig() *PermissionHelperConfig {
	return &PermissionHelperConfig{
		EnableCaching:    false,
		CacheTTL:         300, // 5 minutes
		EnableAuditLog:   true,
		StrictValidation: false,
	}
}

// NewPermissionHelper creates a new permission helper instance
func NewPermissionHelper(iamService iam.Service) *PermissionHelper {
	return &PermissionHelper{
		iamService: iamService,
		logger:     nil,
		config:     DefaultPermissionHelperConfig(),
	}
}

// NewPermissionHelperWithLogger creates a new permission helper with logger
func NewPermissionHelperWithLogger(iamService iam.Service, log logger.Logger) *PermissionHelper {
	return &PermissionHelper{
		iamService: iamService,
		logger:     log,
		config:     DefaultPermissionHelperConfig(),
	}
}

// NewPermissionHelperWithConfig creates a new permission helper with custom config
func NewPermissionHelperWithConfig(iamService iam.Service, log logger.Logger, config *PermissionHelperConfig) *PermissionHelper {
	if config == nil {
		config = DefaultPermissionHelperConfig()
	}
	return &PermissionHelper{
		iamService: iamService,
		logger:     log,
		config:     config,
	}
}

// SetLogger sets the logger instance
func (h *PermissionHelper) SetLogger(log logger.Logger) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.logger = log
}

// GetConfig returns a copy of the current configuration
func (h *PermissionHelper) GetConfig() PermissionHelperConfig {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return *h.config
}

// UpdateConfig updates the configuration
func (h *PermissionHelper) UpdateConfig(config *PermissionHelperConfig) {
	if config == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.config = config

	h.logDebug("Permission helper config updated", logger.Fields{
		"enable_caching":    config.EnableCaching,
		"cache_ttl":         config.CacheTTL,
		"enable_audit_log":  config.EnableAuditLog,
		"strict_validation": config.StrictValidation,
	})
}

// Logging helpers
func (h *PermissionHelper) logDebug(msg string, fields ...logger.Fields) {
	if h.logger != nil {
		h.logger.Debug(msg, fields...)
	}
}

func (h *PermissionHelper) logInfo(msg string, fields ...logger.Fields) {
	if h.logger != nil {
		h.logger.Info(msg, fields...)
	}
}

func (h *PermissionHelper) logWarn(msg string, fields ...logger.Fields) {
	if h.logger != nil {
		h.logger.Warn(msg, fields...)
	}
}

func (h *PermissionHelper) logError(msg string, fields ...logger.Fields) {
	if h.logger != nil {
		h.logger.Error(msg, fields...)
	}
}

// PermissionContext holds the context needed for permission evaluation
type PermissionContext struct {
	UserID       uuid.UUID
	EntityID     *uuid.UUID
	ResourceType string
	ResourceID   *uuid.UUID
	Action       string
	Context      map[string]any
}

// Validate checks if the permission context is valid
func (pc *PermissionContext) Validate() error {
	if pc.UserID == uuid.Nil {
		return ErrInvalidUserID
	}
	if pc.ResourceType == "" {
		return ErrInvalidResourceType
	}
	if pc.Action == "" {
		return ErrInvalidAction
	}
	return nil
}

// PermissionContextBuilder provides a fluent API for building permission contexts
type PermissionContextBuilder struct {
	context PermissionContext
}

// NewPermissionContext creates a new permission context builder
func NewPermissionContext() *PermissionContextBuilder {
	return &PermissionContextBuilder{
		context: PermissionContext{
			Context: make(map[string]any),
		},
	}
}

func (b *PermissionContextBuilder) WithUserID(userID uuid.UUID) *PermissionContextBuilder {
	b.context.UserID = userID
	return b
}

func (b *PermissionContextBuilder) WithEntityID(entityID uuid.UUID) *PermissionContextBuilder {
	b.context.EntityID = &entityID
	return b
}

func (b *PermissionContextBuilder) WithResourceType(resourceType string) *PermissionContextBuilder {
	b.context.ResourceType = resourceType
	return b
}

func (b *PermissionContextBuilder) WithResourceID(resourceID uuid.UUID) *PermissionContextBuilder {
	b.context.ResourceID = &resourceID
	return b
}

func (b *PermissionContextBuilder) WithAction(action string) *PermissionContextBuilder {
	b.context.Action = action
	return b
}

func (b *PermissionContextBuilder) WithContext(key string, value any) *PermissionContextBuilder {
	if b.context.Context == nil {
		b.context.Context = make(map[string]any)
	}
	b.context.Context[key] = value
	return b
}

func (b *PermissionContextBuilder) WithContextMap(contextMap map[string]any) *PermissionContextBuilder {
	b.context.Context = contextMap
	return b
}

func (b *PermissionContextBuilder) Build() *PermissionContext {
	ctx := b.context
	return &ctx
}

// HasPermission checks if the current user has permission to perform an action
func (h *PermissionHelper) HasPermission(ctx context.Context, permCtx *PermissionContext) bool {
	if permCtx == nil {
		h.logWarn("HasPermission called with nil permission context")
		return false
	}

	if permCtx.UserID == uuid.Nil {
		h.logWarn("HasPermission called with nil UserID")
		return false
	}

	// Validate if strict validation is enabled
	h.mu.RLock()
	strictValidation := h.config.StrictValidation
	h.mu.RUnlock()

	if strictValidation {
		if err := permCtx.Validate(); err != nil {
			h.logError("Permission context validation failed", logger.Fields{
				"error":   err.Error(),
				"user_id": permCtx.UserID.String(),
			})
			return false
		}
	}

	// Prepare permission evaluation request
	req := &authz.PermissionEvaluationRequest{
		UserID:       permCtx.UserID,
		ResourceType: permCtx.ResourceType,
		ResourceID:   permCtx.ResourceID,
		Action:       permCtx.Action,
		EntityID:     permCtx.EntityID,
		Context:      permCtx.Context,
	}

	// Evaluate permission using authorization service
	result, err := h.iamService.Authorization().EvaluatePermission(ctx, req)
	if err != nil {
		h.logError("Permission evaluation failed", logger.Fields{
			"error":         err.Error(),
			"user_id":       permCtx.UserID.String(),
			"resource_type": permCtx.ResourceType,
			"action":        permCtx.Action,
		})
		return false
	}

	allowed := result.Decision == "allow"

	// Audit logging
	h.mu.RLock()
	enableAudit := h.config.EnableAuditLog
	h.mu.RUnlock()

	if enableAudit {
		h.logInfo("Permission check", logger.Fields{
			"user_id":       permCtx.UserID.String(),
			"resource_type": permCtx.ResourceType,
			"resource_id":   h.formatResourceID(permCtx.ResourceID),
			"action":        permCtx.Action,
			"decision":      result.Decision,
			"allowed":       allowed,
		})
	}

	return allowed
}

// HasPermissionWithValidation checks permission and returns detailed error
func (h *PermissionHelper) HasPermissionWithValidation(ctx context.Context, permCtx *PermissionContext) (bool, error) {
	if permCtx == nil {
		return false, ErrNilPermissionContext
	}

	if err := permCtx.Validate(); err != nil {
		h.logError("Permission context validation failed", logger.Fields{
			"error": err.Error(),
		})
		return false, fmt.Errorf("validation failed: %w", err)
	}

	allowed := h.HasPermission(ctx, permCtx)
	if !allowed {
		return false, ErrPermissionDenied
	}

	return true, nil
}

// HasAnyPermission checks if the current user has any of the specified permissions
func (h *PermissionHelper) HasAnyPermission(ctx context.Context, permissions []*PermissionContext) bool {
	if len(permissions) == 0 {
		h.logWarn("HasAnyPermission called with empty permissions list")
		return false
	}

	for i, perm := range permissions {
		if h.HasPermission(ctx, perm) {
			h.logDebug("HasAnyPermission granted", logger.Fields{
				"matched_index": i,
				"resource_type": perm.ResourceType,
				"action":        perm.Action,
			})
			return true
		}
	}

	h.logDebug("HasAnyPermission denied for all permissions", logger.Fields{
		"permission_count": len(permissions),
	})
	return false
}

// HasAllPermissions checks if the current user has all of the specified permissions
func (h *PermissionHelper) HasAllPermissions(ctx context.Context, permissions []*PermissionContext) bool {
	if len(permissions) == 0 {
		h.logWarn("HasAllPermissions called with empty permissions list")
		return true // Empty set - vacuous truth
	}

	for i, perm := range permissions {
		if !h.HasPermission(ctx, perm) {
			h.logDebug("HasAllPermissions denied", logger.Fields{
				"failed_index":  i,
				"resource_type": perm.ResourceType,
				"action":        perm.Action,
			})
			return false
		}
	}

	h.logDebug("HasAllPermissions granted", logger.Fields{
		"permission_count": len(permissions),
	})
	return true
}

// CRUD Permission Helpers

// CanRead checks if the current user can read a resource
func (h *PermissionHelper) CanRead(ctx context.Context, userID uuid.UUID, resourceType string, resourceID *uuid.UUID, entityID *uuid.UUID) bool {
	return h.HasPermission(ctx, &PermissionContext{
		UserID:       userID,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Action:       ActionRead,
		EntityID:     entityID,
	})
}

// CanWrite checks if the current user can write/modify a resource
func (h *PermissionHelper) CanWrite(ctx context.Context, userID uuid.UUID, resourceType string, resourceID *uuid.UUID, entityID *uuid.UUID) bool {
	return h.HasPermission(ctx, &PermissionContext{
		UserID:       userID,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Action:       ActionWrite,
		EntityID:     entityID,
	})
}

// CanDelete checks if the current user can delete a resource
func (h *PermissionHelper) CanDelete(ctx context.Context, userID uuid.UUID, resourceType string, resourceID *uuid.UUID, entityID *uuid.UUID) bool {
	return h.HasPermission(ctx, &PermissionContext{
		UserID:       userID,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Action:       ActionDelete,
		EntityID:     entityID,
	})
}

// CanCreate checks if the current user can create a resource
func (h *PermissionHelper) CanCreate(ctx context.Context, userID uuid.UUID, resourceType string, entityID *uuid.UUID) bool {
	return h.HasPermission(ctx, &PermissionContext{
		UserID:       userID,
		ResourceType: resourceType,
		Action:       ActionCreate,
		EntityID:     entityID,
	})
}

// CanManage checks if the current user can manage (full access) a resource
func (h *PermissionHelper) CanManage(ctx context.Context, userID uuid.UUID, resourceType string, resourceID *uuid.UUID, entityID *uuid.UUID) bool {
	return h.HasPermission(ctx, &PermissionContext{
		UserID:       userID,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Action:       ActionManage,
		EntityID:     entityID,
	})
}

// Context Extraction

// GetUserFromContext extracts user information from request context
func (h *PermissionHelper) GetUserFromContext(ctx context.Context) (uuid.UUID, *uuid.UUID, error) {
	userID, ok := shared.GetUserID(ctx)
	if !ok {
		h.logWarn("Failed to get user ID from context")
		return uuid.Nil, nil, ErrUserNotFound
	}

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		h.logDebug("No tenant ID in context", logger.Fields{
			"user_id": userID.String(),
		})
		return userID, nil, nil
	}

	h.logDebug("Extracted user from context", logger.Fields{
		"user_id":   userID.String(),
		"tenant_id": tenantID.String(),
	})

	return userID, &tenantID, nil
}

// GetUserEffectivePermissions retrieves all effective permissions for a user
func (h *PermissionHelper) GetUserEffectivePermissions(ctx context.Context, userID uuid.UUID, entityID *uuid.UUID) (*authz.UserEffectivePermissions, error) {
	if userID == uuid.Nil {
		return nil, ErrInvalidUserID
	}

	h.logDebug("Getting effective permissions", logger.Fields{
		"user_id":   userID.String(),
		"entity_id": h.formatEntityID(entityID),
	})

	perms, err := h.iamService.Authorization().GetUserEffectivePermissions(ctx, userID, entityID)
	if err != nil {
		h.logError("Failed to get effective permissions", logger.Fields{
			"error":   err.Error(),
			"user_id": userID.String(),
		})
		return nil, fmt.Errorf("failed to get effective permissions: %w", err)
	}

	return perms, nil
}

// Role and Admin Checks

// IsSystemAdmin checks if user has system-level admin permissions
func (h *PermissionHelper) IsSystemAdmin(ctx context.Context, userID uuid.UUID) bool {
	if userID == uuid.Nil {
		h.logWarn("IsSystemAdmin called with nil user ID")
		return false
	}

	isAdmin := h.HasPermission(ctx, &PermissionContext{
		UserID:       userID,
		ResourceType: ResourceTypeSystem,
		Action:       ActionAdmin,
	})

	h.logDebug("System admin check", logger.Fields{
		"user_id":  userID.String(),
		"is_admin": isAdmin,
	})

	return isAdmin
}

// IsTenantAdmin checks if user has tenant-level admin permissions
func (h *PermissionHelper) IsTenantAdmin(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) bool {
	if userID == uuid.Nil || tenantID == uuid.Nil {
		h.logWarn("IsTenantAdmin called with nil ID", logger.Fields{
			"user_id_nil":   userID == uuid.Nil,
			"tenant_id_nil": tenantID == uuid.Nil,
		})
		return false
	}

	isAdmin := h.HasPermission(ctx, &PermissionContext{
		UserID:       userID,
		ResourceType: ResourceTypeTenant,
		Action:       ActionAdmin,
		EntityID:     &tenantID,
	})

	h.logDebug("Tenant admin check", logger.Fields{
		"user_id":   userID.String(),
		"tenant_id": tenantID.String(),
		"is_admin":  isAdmin,
	})

	return isAdmin
}

// Access Checks

// CanAccessConsole checks if user can access the admin console
func (h *PermissionHelper) CanAccessConsole(ctx context.Context, userID uuid.UUID) bool {
	return h.HasPermission(ctx, &PermissionContext{
		UserID:       userID,
		ResourceType: ResourceTypeConsole,
		Action:       ActionAccess,
	})
}

// CanAccessWorkspace checks if user can access a specific workspace
func (h *PermissionHelper) CanAccessWorkspace(ctx context.Context, userID uuid.UUID, workspaceID uuid.UUID, tenantID uuid.UUID) bool {
	if userID == uuid.Nil || workspaceID == uuid.Nil || tenantID == uuid.Nil {
		h.logWarn("CanAccessWorkspace called with nil ID")
		return false
	}

	return h.HasPermission(ctx, &PermissionContext{
		UserID:       userID,
		ResourceType: ResourceTypeWorkspace,
		ResourceID:   &workspaceID,
		Action:       ActionAccess,
		EntityID:     &tenantID,
	})
}

// Batch Permission Checks

// CheckMultipleResources checks permissions for multiple resources of the same type
func (h *PermissionHelper) CheckMultipleResources(ctx context.Context, userID uuid.UUID, resourceType string, resourceIDs []uuid.UUID, action string, entityID *uuid.UUID) map[uuid.UUID]bool {
	results := make(map[uuid.UUID]bool, len(resourceIDs))

	for _, resourceID := range resourceIDs {
		rid := resourceID
		allowed := h.HasPermission(ctx, &PermissionContext{
			UserID:       userID,
			ResourceType: resourceType,
			ResourceID:   &rid,
			Action:       action,
			EntityID:     entityID,
		})
		results[resourceID] = allowed
	}

	h.logDebug("Batch permission check completed", logger.Fields{
		"user_id":        userID.String(),
		"resource_type":  resourceType,
		"resource_count": len(resourceIDs),
		"action":         action,
	})

	return results
}

// GetResourcesWithPermission filters resources based on permission
func (h *PermissionHelper) GetResourcesWithPermission(ctx context.Context, userID uuid.UUID, resourceType string, resourceIDs []uuid.UUID, action string, entityID *uuid.UUID) []uuid.UUID {
	var allowedResources []uuid.UUID

	for _, resourceID := range resourceIDs {
		rid := resourceID
		if h.HasPermission(ctx, &PermissionContext{
			UserID:       userID,
			ResourceType: resourceType,
			ResourceID:   &rid,
			Action:       action,
			EntityID:     entityID,
		}) {
			allowedResources = append(allowedResources, resourceID)
		}
	}

	h.logDebug("Filtered resources by permission", logger.Fields{
		"user_id":         userID.String(),
		"resource_type":   resourceType,
		"total_resources": len(resourceIDs),
		"allowed_count":   len(allowedResources),
		"action":          action,
	})

	return allowedResources
}

// Utility functions

func (h *PermissionHelper) formatResourceID(resourceID *uuid.UUID) string {
	if resourceID == nil {
		return "nil"
	}
	return resourceID.String()
}

func (h *PermissionHelper) formatEntityID(entityID *uuid.UUID) string {
	if entityID == nil {
		return "nil"
	}
	return entityID.String()
}

// Error definitions
var (
	ErrUserNotFound         = errors.New("user not found in context")
	ErrInvalidUserID        = errors.New("invalid user ID")
	ErrInvalidResourceType  = errors.New("invalid resource type")
	ErrInvalidAction        = errors.New("invalid action")
	ErrNilPermissionContext = errors.New("permission context cannot be nil")
	ErrPermissionDenied     = errors.New("permission denied")
	ErrServiceNotAvailable  = errors.New("IAM service not available")
)

// Common permission constants
const (
	// Resource types
	ResourceTypeUser          = "user"
	ResourceTypeTenant        = "tenant"
	ResourceTypeWorkspace     = "workspace"
	ResourceTypeFinancialData = "financial_data"
	ResourceTypeReport        = "report"
	ResourceTypeSystem        = "system"
	ResourceTypeConsole       = "console"
	ResourceTypeModule        = "module"
	ResourceTypeIntegration   = "integration"

	// Actions
	ActionRead    = "read"
	ActionWrite   = "write"
	ActionCreate  = "create"
	ActionDelete  = "delete"
	ActionManage  = "manage"
	ActionAdmin   = "admin"
	ActionAccess  = "access"
	ActionExecute = "execute"
	ActionExport  = "export"
	ActionImport  = "import"
)
