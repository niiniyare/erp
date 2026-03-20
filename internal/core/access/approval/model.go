package approval

import (
	"github.com/google/uuid"
	"awo/internal/shared/types"
)

// Type aliases for shared types
type RequestType = types.RequestType

const (
	RequestTypeRoleAssignment  = types.RequestTypeRoleAssignment
	RequestTypePermissionGrant = types.RequestTypePermissionGrant
	RequestTypeResourceAccess  = types.RequestTypeResourceAccess
	RequestTypeElevation       = types.RequestTypeElevation
)

// ApprovalRule represents a rule for determining approvers
type ApprovalRule struct {
	ID          uuid.UUID             `json:"id"`
	TenantID    uuid.UUID             `json:"tenant_id"`
	EntityID    uuid.UUID             `json:"entity_id"`
	Name        string                `json:"name"`
	RequestType RequestType           `json:"request_type"`
	Conditions  map[string]any        `json:"conditions"` // ABAC-style conditions
	Approvers   ApproverConfiguration `json:"approvers"`
	Priority    int                   `json:"priority"` // Higher priority rules evaluated first
	IsActive    bool                  `json:"is_active"`
}

// ApproverConfiguration defines who can approve requests
type ApproverConfiguration struct {
	Type               string              `json:"type"` // ROLE_BASED, HIERARCHY_BASED, SPECIFIC_USERS, MIXED
	RequiredApprovals  int                 `json:"required_approvals"`
	RoleBasedApprovers *RoleBasedApprovers `json:"role_based_approvers,omitempty"`
	HierarchyApprovers *HierarchyApprovers `json:"hierarchy_approvers,omitempty"`
	SpecificApprovers  []uuid.UUID         `json:"specific_approvers,omitempty"`
	FallbackApprovers  []uuid.UUID         `json:"fallback_approvers,omitempty"`
}

// RoleBasedApprovers defines approval based on roles
type RoleBasedApprovers struct {
	RequiredRoles    []uuid.UUID `json:"required_roles"`               // Users must have one of these roles
	MinSecurityLevel *int        `json:"min_security_level,omitempty"` // Minimum security clearance
	SameEntity       bool        `json:"same_entity"`                  // Must be in same entity
	ExcludeRequester bool        `json:"exclude_requester"`            // Cannot approve own requests
}

// HierarchyApprovers defines approval based on organizational hierarchy
type HierarchyApprovers struct {
	ManagerLevelsUp      int  `json:"manager_levels_up"`      // How many levels up in hierarchy
	IncludePeers         bool `json:"include_peers"`          // Include users at same level
	IncludeSubordinates  bool `json:"include_subordinates"`   // Include direct reports
	RequireDirectManager bool `json:"require_direct_manager"` // Must be direct manager
}

// ApproverValidationResult represents the result of approver validation
type ApproverValidationResult struct {
	IsValid           bool           `json:"is_valid"`
	Reason            string         `json:"reason,omitempty"`
	ApprovalRules     []ApprovalRule `json:"approval_rules"`
	RequiredApprovals int            `json:"required_approvals"`
	SecurityLevel     *int           `json:"security_level,omitempty"`
}

// UserHierarchy represents a user's position in organizational hierarchy
type UserHierarchy struct {
	UserID        uuid.UUID   `json:"user_id"`
	ManagerID     *uuid.UUID  `json:"manager_id,omitempty"`
	Level         int         `json:"level"`          // Organizational level (0 = top)
	SecurityLevel int         `json:"security_level"` // Security clearance level
	DepartmentID  *uuid.UUID  `json:"department_id,omitempty"`
	DirectReports []uuid.UUID `json:"direct_reports"`
	ManagerChain  []uuid.UUID `json:"manager_chain"` // All managers up the chain
}
