package helpers

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/niiniyare/erp/internal/core/iam"
	"github.com/niiniyare/erp/internal/core/iam/authz"
	"github.com/niiniyare/erp/internal/shared"
)

// PermissionHelper provides template helper functions for ABAC permission checks
type PermissionHelper struct {
	iamService iam.Service
}

// NewPermissionHelper creates a new permission helper instance
func NewPermissionHelper(iamService iam.Service) *PermissionHelper {
	return &PermissionHelper{
		iamService: iamService,
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

// HasPermission checks if the current user has permission to perform an action
func (h *PermissionHelper) HasPermission(ctx context.Context, permCtx *PermissionContext) bool {
	if permCtx == nil || permCtx.UserID == uuid.Nil {
		return false
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
		return false
	}

	return result.Decision == "allow"
}

// HasAnyPermission checks if the current user has any of the specified permissions
func (h *PermissionHelper) HasAnyPermission(ctx context.Context, permissions []*PermissionContext) bool {
	for _, perm := range permissions {
		if h.HasPermission(ctx, perm) {
			return true
		}
	}
	return false
}

// HasAllPermissions checks if the current user has all of the specified permissions
func (h *PermissionHelper) HasAllPermissions(ctx context.Context, permissions []*PermissionContext) bool {
	for _, perm := range permissions {
		if !h.HasPermission(ctx, perm) {
			return false
		}
	}
	return true
}

// CanRead checks if the current user can read a resource
func (h *PermissionHelper) CanRead(ctx context.Context, userID uuid.UUID, resourceType string, resourceID *uuid.UUID, entityID *uuid.UUID) bool {
	return h.HasPermission(ctx, &PermissionContext{
		UserID:       userID,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Action:       "read",
		EntityID:     entityID,
	})
}

// CanWrite checks if the current user can write/modify a resource
func (h *PermissionHelper) CanWrite(ctx context.Context, userID uuid.UUID, resourceType string, resourceID *uuid.UUID, entityID *uuid.UUID) bool {
	return h.HasPermission(ctx, &PermissionContext{
		UserID:       userID,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Action:       "write",
		EntityID:     entityID,
	})
}

// CanDelete checks if the current user can delete a resource
func (h *PermissionHelper) CanDelete(ctx context.Context, userID uuid.UUID, resourceType string, resourceID *uuid.UUID, entityID *uuid.UUID) bool {
	return h.HasPermission(ctx, &PermissionContext{
		UserID:       userID,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Action:       "delete",
		EntityID:     entityID,
	})
}

// CanCreate checks if the current user can create a resource
func (h *PermissionHelper) CanCreate(ctx context.Context, userID uuid.UUID, resourceType string, entityID *uuid.UUID) bool {
	return h.HasPermission(ctx, &PermissionContext{
		UserID:       userID,
		ResourceType: resourceType,
		Action:       "create",
		EntityID:     entityID,
	})
}

// CanManage checks if the current user can manage (full access) a resource
func (h *PermissionHelper) CanManage(ctx context.Context, userID uuid.UUID, resourceType string, resourceID *uuid.UUID, entityID *uuid.UUID) bool {
	return h.HasPermission(ctx, &PermissionContext{
		UserID:       userID,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Action:       "manage",
		EntityID:     entityID,
	})
}

// GetUserFromContext extracts user information from request context
func (h *PermissionHelper) GetUserFromContext(ctx context.Context) (uuid.UUID, *uuid.UUID, error) {
	userID, ok := shared.GetUserID(ctx)
	if !ok {
		return uuid.Nil, nil, ErrUserNotFound
	}

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return userID, nil, nil
	}

	return userID, &tenantID, nil
}

// GetUserEffectivePermissions retrieves all effective permissions for a user
func (h *PermissionHelper) GetUserEffectivePermissions(ctx context.Context, userID uuid.UUID, entityID *uuid.UUID) (*authz.UserEffectivePermissions, error) {
	return h.iamService.Authorization().GetUserEffectivePermissions(ctx, userID, entityID)
}

// Utility functions for common permission patterns

// IsSystemAdmin checks if user has system-level admin permissions
func (h *PermissionHelper) IsSystemAdmin(ctx context.Context, userID uuid.UUID) bool {
	return h.HasPermission(ctx, &PermissionContext{
		UserID:       userID,
		ResourceType: "system",
		Action:       "admin",
	})
}

// IsTenantAdmin checks if user has tenant-level admin permissions
func (h *PermissionHelper) IsTenantAdmin(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) bool {
	return h.HasPermission(ctx, &PermissionContext{
		UserID:       userID,
		ResourceType: "tenant",
		Action:       "admin",
		EntityID:     &tenantID,
	})
}

// CanAccessConsole checks if user can access the admin console
func (h *PermissionHelper) CanAccessConsole(ctx context.Context, userID uuid.UUID) bool {
	return h.HasPermission(ctx, &PermissionContext{
		UserID:       userID,
		ResourceType: "console",
		Action:       "access",
	})
}

// CanAccessWorkspace checks if user can access a specific workspace
func (h *PermissionHelper) CanAccessWorkspace(ctx context.Context, userID uuid.UUID, workspaceID uuid.UUID, tenantID uuid.UUID) bool {
	return h.HasPermission(ctx, &PermissionContext{
		UserID:       userID,
		ResourceType: "workspace",
		ResourceID:   &workspaceID,
		Action:       "access",
		EntityID:     &tenantID,
	})
}

// Error definitions
var (
	ErrUserNotFound = errors.New("user not found in context")
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

	// Actions
	ActionRead   = "read"
	ActionWrite  = "write"
	ActionCreate = "create"
	ActionDelete = "delete"
	ActionManage = "manage"
	ActionAdmin  = "admin"
	ActionAccess = "access"
)
