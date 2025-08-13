package featureflag

import (
	"context"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/core/abac"
	"github.com/niiniyare/erp/internal/shared/types"
)

// AdminFeatureFlagPermissions defines the ABAC permissions for admin feature flag operations
const (
	// Resource types
	ResourceTypeFeatureFlag       = "feature_flag"
	ResourceTypeFeatureFlagBulk   = "feature_flag_bulk"
	ResourceTypeFeatureFlagSystem = "feature_flag_system"

	// Actions for individual feature flags
	ActionFeatureFlagRead   = "read"
	ActionFeatureFlagCreate = "create"
	ActionFeatureFlagUpdate = "update"
	ActionFeatureFlagDelete = "delete"

	// Actions for bulk operations
	ActionBulkEnable  = "bulk_enable"
	ActionBulkDisable = "bulk_disable"
	ActionBulkDelete  = "bulk_delete"
	ActionBulkRollout = "bulk_rollout"

	// Actions for system management
	ActionSystemHealth     = "system_health"
	ActionSystemMetrics    = "system_metrics"
	ActionEmergencyControl = "emergency_control"
	ActionCacheManagement  = "cache_management"

	// Roles with admin privileges
	RoleFeatureFlagAdmin    = "feature_flag_admin"
	RoleSystemAdmin         = "system_admin"
	RoleSuperAdmin          = "super_admin"
	RoleTenantAdmin         = "tenant_admin"
	RoleFeatureFlagOperator = "feature_flag_operator"

	// Attribute keys for context evaluation
	AttrUserRoles         = "user.roles"
	AttrUserTenantID      = "user.tenant_id"
	AttrResourceTenantID  = "resource.tenant_id"
	AttrTimeOfDay         = "environment.time_of_day"
	AttrIPAddress         = "environment.ip_address"
	AttrRequestReason     = "request.reason"
	AttrBulkOperationSize = "request.bulk_operation_size"
)

// AdminPermissionEvaluator provides ABAC evaluation for admin feature flag operations
type AdminPermissionEvaluator struct {
	abacService abac.Service
}

// NewAdminPermissionEvaluator creates a new admin permission evaluator
func NewAdminPermissionEvaluator(abacService abac.Service) *AdminPermissionEvaluator {
	return &AdminPermissionEvaluator{
		abacService: abacService,
	}
}

// EvaluateBulkOperationPermission evaluates permission for bulk operations
func (e *AdminPermissionEvaluator) EvaluateBulkOperationPermission(
	ctx context.Context,
	userID uuid.UUID,
	action string,
	tenantID uuid.UUID,
	bulkSize int,
	reason string,
) (*abac.PermissionEvaluationResult, error) {
	req := &abac.PermissionEvaluationRequest{
		UserID:       userID,
		ResourceType: ResourceTypeFeatureFlagBulk,
		Action:       action,
		EntityID:     &tenantID,
		Context: map[string]any{
			AttrBulkOperationSize: bulkSize,
			AttrRequestReason:     reason,
		},
	}

	return e.abacService.EvaluatePermission(ctx, req)
}

// EvaluateSystemOperationPermission evaluates permission for system operations
func (e *AdminPermissionEvaluator) EvaluateSystemOperationPermission(
	ctx context.Context,
	userID uuid.UUID,
	action string,
	tenantID uuid.UUID,
) (*abac.PermissionEvaluationResult, error) {
	req := &abac.PermissionEvaluationRequest{
		UserID:       userID,
		ResourceType: ResourceTypeFeatureFlagSystem,
		Action:       action,
		EntityID:     &tenantID,
		Context: map[string]any{
			"system_operation": true,
		},
	}

	return e.abacService.EvaluatePermission(ctx, req)
}

// EvaluateEmergencyOperationPermission evaluates permission for emergency operations
func (e *AdminPermissionEvaluator) EvaluateEmergencyOperationPermission(
	ctx context.Context,
	userID uuid.UUID,
	tenantID uuid.UUID,
	reason string,
) (*abac.PermissionEvaluationResult, error) {
	req := &abac.PermissionEvaluationRequest{
		UserID:       userID,
		ResourceType: ResourceTypeFeatureFlagSystem,
		Action:       ActionEmergencyControl,
		EntityID:     &tenantID,
		Context: map[string]any{
			AttrRequestReason:     reason,
			"emergency_operation": true,
			"requires_approval":   true,
		},
	}

	return e.abacService.EvaluatePermission(ctx, req)
}

// AdminPolicyDefinitions provides predefined ABAC policies for admin operations
type AdminPolicyDefinitions struct{}

// GetBulkOperationPolicies returns policies for bulk operations
func (p *AdminPolicyDefinitions) GetBulkOperationPolicies() []*PolicyDefinition {
	return []*PolicyDefinition{
		{
			Name:         "FeatureFlag_Admin_BulkOperations_Allow",
			Description:  "Allow feature flag admins to perform bulk operations",
			Effect:       types.PolicyEffectAllow,
			ResourceType: ResourceTypeFeatureFlagBulk,
			Actions:      []string{ActionBulkEnable, ActionBulkDisable, ActionBulkRollout},
			Conditions: []PolicyCondition{
				{
					Attribute: AttrUserRoles,
					Operator:  "contains",
					Value:     RoleFeatureFlagAdmin,
				},
				{
					Attribute: AttrUserTenantID,
					Operator:  "equals",
					Value:     "${resource.tenant_id}",
				},
				{
					Attribute: AttrBulkOperationSize,
					Operator:  "less_than_or_equal",
					Value:     100, // Maximum bulk operation size
				},
			},
		},
		{
			Name:         "SystemAdmin_BulkOperations_Allow",
			Description:  "Allow system admins to perform any bulk operations",
			Effect:       types.PolicyEffectAllow,
			ResourceType: ResourceTypeFeatureFlagBulk,
			Actions:      []string{ActionBulkEnable, ActionBulkDisable, ActionBulkDelete, ActionBulkRollout},
			Conditions: []PolicyCondition{
				{
					Attribute: AttrUserRoles,
					Operator:  "contains_any",
					Value:     []string{RoleSystemAdmin, RoleSuperAdmin},
				},
			},
		},
		{
			Name:         "BulkDelete_Restricted",
			Description:  "Restrict bulk delete operations to super admins only",
			Effect:       types.PolicyEffectDeny,
			ResourceType: ResourceTypeFeatureFlagBulk,
			Actions:      []string{ActionBulkDelete},
			Conditions: []PolicyCondition{
				{
					Attribute: AttrUserRoles,
					Operator:  "not_contains",
					Value:     RoleSuperAdmin,
				},
			},
		},
	}
}

// GetSystemManagementPolicies returns policies for system management operations
func (p *AdminPolicyDefinitions) GetSystemManagementPolicies() []*PolicyDefinition {
	return []*PolicyDefinition{
		{
			Name:         "FeatureFlag_SystemHealth_Allow",
			Description:  "Allow operators and admins to check system health",
			Effect:       types.PolicyEffectAllow,
			ResourceType: ResourceTypeFeatureFlagSystem,
			Actions:      []string{ActionSystemHealth, ActionSystemMetrics},
			Conditions: []PolicyCondition{
				{
					Attribute: AttrUserRoles,
					Operator:  "contains_any",
					Value: []string{
						RoleFeatureFlagOperator,
						RoleFeatureFlagAdmin,
						RoleSystemAdmin,
						RoleSuperAdmin,
					},
				},
			},
		},
		{
			Name:         "CacheManagement_Admin_Allow",
			Description:  "Allow admins to manage feature flag cache",
			Effect:       types.PolicyEffectAllow,
			ResourceType: ResourceTypeFeatureFlagSystem,
			Actions:      []string{ActionCacheManagement},
			Conditions: []PolicyCondition{
				{
					Attribute: AttrUserRoles,
					Operator:  "contains_any",
					Value: []string{
						RoleFeatureFlagAdmin,
						RoleSystemAdmin,
						RoleSuperAdmin,
					},
				},
			},
		},
		{
			Name:         "EmergencyControl_SuperAdmin_Only",
			Description:  "Emergency controls restricted to super admins",
			Effect:       types.PolicyEffectAllow,
			ResourceType: ResourceTypeFeatureFlagSystem,
			Actions:      []string{ActionEmergencyControl},
			Conditions: []PolicyCondition{
				{
					Attribute: AttrUserRoles,
					Operator:  "contains",
					Value:     RoleSuperAdmin,
				},
				{
					Attribute: AttrRequestReason,
					Operator:  "not_empty",
					Value:     nil,
				},
			},
		},
	}
}

// GetTimeBasedPolicies returns time-based access policies
func (p *AdminPolicyDefinitions) GetTimeBasedPolicies() []*PolicyDefinition {
	return []*PolicyDefinition{
		{
			Name:         "BusinessHours_BulkOperations_Restrict",
			Description:  "Restrict large bulk operations to business hours",
			Effect:       types.PolicyEffectDeny,
			ResourceType: ResourceTypeFeatureFlagBulk,
			Actions:      []string{ActionBulkEnable, ActionBulkDisable},
			Conditions: []PolicyCondition{
				{
					Attribute: AttrBulkOperationSize,
					Operator:  "greater_than",
					Value:     50,
				},
				{
					Attribute: AttrTimeOfDay,
					Operator:  "not_between",
					Value:     []string{"09:00", "17:00"},
				},
				{
					Attribute: AttrUserRoles,
					Operator:  "not_contains",
					Value:     RoleSuperAdmin,
				},
			},
		},
	}
}

// PolicyDefinition represents an ABAC policy definition
type PolicyDefinition struct {
	Name         string             `json:"name"`
	Description  string             `json:"description"`
	Effect       types.PolicyEffect `json:"effect"`
	ResourceType string             `json:"resource_type"`
	Actions      []string           `json:"actions"`
	Conditions   []PolicyCondition  `json:"conditions"`
	Priority     int                `json:"priority"`
	IsActive     bool               `json:"is_active"`
}

// PolicyCondition represents a condition within an ABAC policy
type PolicyCondition struct {
	Attribute string `json:"attribute"`
	Operator  string `json:"operator"`
	Value     any    `json:"value"`
}

// RequiredRolesForAction returns the roles required for a specific action
func RequiredRolesForAction(action string) []string {
	switch action {
	case ActionBulkEnable, ActionBulkDisable:
		return []string{RoleFeatureFlagAdmin, RoleSystemAdmin, RoleSuperAdmin}
	case ActionBulkDelete:
		return []string{RoleSystemAdmin, RoleSuperAdmin}
	case ActionBulkRollout:
		return []string{RoleFeatureFlagAdmin, RoleSystemAdmin, RoleSuperAdmin}
	case ActionSystemHealth, ActionSystemMetrics:
		return []string{RoleFeatureFlagOperator, RoleFeatureFlagAdmin, RoleSystemAdmin, RoleSuperAdmin}
	case ActionCacheManagement:
		return []string{RoleFeatureFlagAdmin, RoleSystemAdmin, RoleSuperAdmin}
	case ActionEmergencyControl:
		return []string{RoleSuperAdmin}
	default:
		return []string{RoleSystemAdmin, RoleSuperAdmin}
	}
}

// IsHighRiskOperation determines if an operation is considered high risk
func IsHighRiskOperation(action string, bulkSize int) bool {
	switch action {
	case ActionBulkDelete:
		return true
	case ActionEmergencyControl:
		return true
	case ActionBulkEnable, ActionBulkDisable:
		return bulkSize > 50
	case ActionBulkRollout:
		return bulkSize > 30
	default:
		return false
	}
}

// GetMinimumRoleForOperation returns the minimum role required for an operation
func GetMinimumRoleForOperation(action string, isEmergency bool) string {
	if isEmergency {
		return RoleSuperAdmin
	}

	switch action {
	case ActionSystemHealth, ActionSystemMetrics:
		return RoleFeatureFlagOperator
	case ActionBulkEnable, ActionBulkDisable, ActionBulkRollout, ActionCacheManagement:
		return RoleFeatureFlagAdmin
	case ActionBulkDelete, ActionEmergencyControl:
		return RoleSuperAdmin
	default:
		return RoleSystemAdmin
	}
}
