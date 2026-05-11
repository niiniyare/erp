//go:build ignore

package permission

import (
	"time"

	"github.com/google/uuid"
)

// Role represents a role in the system
type Role struct {
	ID           uuid.UUID      `json:"id"`
	TenantID     uuid.UUID      `json:"tenant_id"`
	EntityID     uuid.UUID      `json:"entity_id"`
	Name         string         `json:"name"`
	DisplayName  *string        `json:"display_name,omitempty"`
	Description  *string        `json:"description,omitempty"`
	RoleType     string         `json:"role_type"`
	ParentRoleID *uuid.UUID     `json:"parent_role_id,omitempty"`
	Level        int32          `json:"level"`
	Permissions  map[string]any `json:"permissions"`
	EntityScope  map[string]any `json:"entity_scope"`
	Conditions   map[string]any `json:"conditions"`
	IsActive     bool           `json:"is_active"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// Permission represents a permission in the system
type Permission struct {
	ID                uuid.UUID      `json:"id"`
	TenantID          uuid.UUID      `json:"tenant_id"`
	ResourceID        uuid.UUID      `json:"resource_id"`
	ActionID          uuid.UUID      `json:"action_id"`
	Name              string         `json:"name"`
	DisplayName       *string        `json:"display_name,omitempty"`
	Description       *string        `json:"description,omitempty"`
	Effect            string         `json:"effect"` // ALLOW or DENY
	Conditions        map[string]any `json:"conditions"`
	DataFilters       map[string]any `json:"data_filters"`
	FieldRestrictions map[string]any `json:"field_restrictions"`
	IsActive          bool           `json:"is_active"`
	CreatedAt         time.Time      `json:"created_at"`
}

// Policy represents an ABAC policy
type Policy struct {
	ID          uuid.UUID      `json:"id"`
	TenantID    uuid.UUID      `json:"tenant_id"`
	EntityID    *uuid.UUID     `json:"entity_id,omitempty"`
	Name        string         `json:"name"`
	DisplayName *string        `json:"display_name,omitempty"`
	Description *string        `json:"description,omitempty"`
	PolicyType  string         `json:"policy_type"`
	Effect      string         `json:"effect"` // ALLOW or DENY
	Priority    int32          `json:"priority"`
	Category    string         `json:"category"`
	Target      map[string]any `json:"target"`
	Rule        map[string]any `json:"rule"`
	Obligations map[string]any `json:"obligations"`
	Advice      map[string]any `json:"advice"`
	IsActive    bool           `json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// EffectivePermission represents a permission granted to a user through roles
type EffectivePermission struct {
	Permission     *Permission `json:"permission"`
	GrantedByRole  *Role       `json:"granted_by_role"`
	EntityID       uuid.UUID   `json:"entity_id"`
	AssignmentType string      `json:"assignment_type"`
	ExpiresAt      *time.Time  `json:"expires_at,omitempty"`
}

// UserRole represents a role assigned to a user
type UserRole struct {
	ID             uuid.UUID  `json:"id"`
	UserID         uuid.UUID  `json:"user_id"`
	RoleID         uuid.UUID  `json:"role_id"`
	EntityID       uuid.UUID  `json:"entity_id"`
	AssignmentType string     `json:"assignment_type"`
	AssignedAt     time.Time  `json:"assigned_at"`
	AssignedBy     *uuid.UUID `json:"assigned_by,omitempty"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	IsActive       bool       `json:"is_active"`
}

// PermissionEvaluationRequest represents a permission evaluation request
type PermissionEvaluationRequest struct {
	UserID       uuid.UUID      `json:"user_id" validate:"required"`
	ResourceName string         `json:"resource_name" validate:"required"`
	ActionName   string         `json:"action_name" validate:"required"`
	EntityID     *uuid.UUID     `json:"entity_id,omitempty"`
	Context      map[string]any `json:"context,omitempty"`
}

// PermissionEvaluationResult represents the result of permission evaluation
type PermissionEvaluationResult struct {
	Allowed          bool                  `json:"allowed"`
	PolicyDecisions  []string              `json:"policy_decisions"`
	EffectiveRoles   []string              `json:"effective_roles"`
	EvaluationTimeMS int                   `json:"evaluation_time_ms"`
	CacheHit         bool                  `json:"cache_hit"`
	RBACResult       *RBACEvaluationResult `json:"rbac_result,omitempty"`
	ABACResult       *ABACEvaluationResult `json:"abac_result,omitempty"`
}

// RBACEvaluationResult represents RBAC evaluation result
type RBACEvaluationResult struct {
	Allowed         bool     `json:"allowed"`
	PolicyDecisions []string `json:"policy_decisions"`
}

// ABACEvaluationRequest represents ABAC evaluation request
type ABACEvaluationRequest struct {
	UserID       uuid.UUID      `json:"user_id" validate:"required"`
	ResourceName string         `json:"resource_name" validate:"required"`
	ActionName   string         `json:"action_name" validate:"required"`
	EntityID     *uuid.UUID     `json:"entity_id,omitempty"`
	Context      map[string]any `json:"context,omitempty"`
}

// ABACEvaluationResult represents ABAC evaluation result
type ABACEvaluationResult struct {
	Allowed            bool           `json:"allowed"`
	PolicyDecisions    []string       `json:"policy_decisions"`
	ApplicablePolicies []string       `json:"applicable_policies"`
	EvaluationDetails  map[string]any `json:"evaluation_details,omitempty"`
}

// BulkPermissionEvaluationRequest represents multiple permission evaluations
type BulkPermissionEvaluationRequest struct {
	Requests []*PermissionEvaluationRequest `json:"requests" validate:"required,min=1"`
}

// PolicyTestRequest represents a policy test request
type PolicyTestRequest struct {
	UserID       uuid.UUID      `json:"user_id" validate:"required"`
	ResourceName string         `json:"resource_name" validate:"required"`
	ActionName   string         `json:"action_name" validate:"required"`
	EntityID     *uuid.UUID     `json:"entity_id,omitempty"`
	Context      map[string]any `json:"context,omitempty"`
}

// PolicyTestResult represents the result of policy testing
type PolicyTestResult struct {
	PolicyID      uuid.UUID      `json:"policy_id"`
	PolicyName    string         `json:"policy_name"`
	TargetMatches bool           `json:"target_matches"`
	RuleResult    bool           `json:"rule_result"`
	Effect        string         `json:"effect"`
	Details       map[string]any `json:"details"`
}
