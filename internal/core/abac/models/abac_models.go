package models

import (
	"time"

	"github.com/google/uuid"
)

// PermissionEvaluationRequest for permission checking
type PermissionEvaluationRequest struct {
	UserID       uuid.UUID      `json:"user_id"`
	Permission   string         `json:"permission"`
	Resource     string         `json:"resource"`
	ResourceName string         `json:"resource_name"`
	ActionName   string         `json:"action_name"`
	EntityID     uuid.UUID      `json:"entity_id"`
	Context      map[string]any `json:"context,omitempty"`
}

// PermissionEvaluationResult contains permission check result
type PermissionEvaluationResult struct {
	Allowed          bool     `json:"allowed"`
	Permissions      []string `json:"permissions"`
	PolicyDecisions  []string `json:"policy_decisions"`
	EffectiveRoles   []string `json:"effective_roles"`
	EvaluationTimeMS int64    `json:"evaluation_time_ms"`
	CacheHit         bool     `json:"cache_hit"`
}

// BulkPermissionEvaluationRequest for bulk permission checking
type BulkPermissionEvaluationRequest struct {
	Requests []*PermissionEvaluationRequest `json:"requests"`
}

// PolicyTestRequest for policy testing
type PolicyTestRequest struct {
	PolicyID     string         `json:"policy_id"`
	UserID       uuid.UUID      `json:"user_id"`
	Context      map[string]any `json:"context"`
	ResourceName string         `json:"resource_name"`
	ActionName   string         `json:"action_name"`
	EntityID     uuid.UUID      `json:"entity_id"`
}

// PolicyTestResult represents the result of policy testing
type PolicyTestResult struct {
	PolicyID      string         `json:"policy_id"`
	PolicyName    string         `json:"policy_name"`
	Effect        string         `json:"effect"`
	TargetMatches bool           `json:"target_matches"`
	RuleResult    bool           `json:"rule_result"`
	Details       map[string]any `json:"details"`
}

// UserPermission represents an effective permission for a user
type UserPermission struct {
	Permission  string    `json:"permission"`
	Resource    string    `json:"resource"`
	Effect      string    `json:"effect"`
	GrantedBy   string    `json:"granted_by"`
	GrantedAt   time.Time `json:"granted_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// RoleHierarchy represents a role in a hierarchy
type RoleHierarchy struct {
	RoleID     uuid.UUID `json:"role_id"`
	RoleName   string    `json:"role_name"`
	Level      int       `json:"level"`
	ParentID   *uuid.UUID `json:"parent_id,omitempty"`
	Children   []*RoleHierarchy `json:"children,omitempty"`
}

// --- Temporal Workflow Models ---

// PermissionEvaluationWorkflowRequest represents a workflow request for permission evaluation
type PermissionEvaluationWorkflowRequest struct {
	RequestID    uuid.UUID                   `json:"request_id"`
	UserID       uuid.UUID                   `json:"user_id"`
	ResourceName string                      `json:"resource_name"`
	ActionName   string                      `json:"action_name"`
	EntityID     uuid.UUID                   `json:"entity_id"`
	Context      map[string]interface{}      `json:"context,omitempty"`
	CacheKey     string                      `json:"cache_key,omitempty"`
}

// PermissionEvaluationWorkflowResult represents the result of permission evaluation workflow
type PermissionEvaluationWorkflowResult struct {
	RequestID        uuid.UUID `json:"request_id"`
	Allowed          bool      `json:"allowed"`
	PolicyDecisions  []string  `json:"policy_decisions"`
	EffectiveRoles   []string  `json:"effective_roles"`
	EvaluationTimeMS int64     `json:"evaluation_time_ms"`
	CacheHit         bool      `json:"cache_hit"`
	ErrorMessage     string    `json:"error_message,omitempty"`
}

// AccessRequestWorkflowRequest represents a workflow request for access requests
type AccessRequestWorkflowRequest struct {
	RequestID      uuid.UUID              `json:"request_id"`
	RequesterID    uuid.UUID              `json:"requester_id"`
	TargetUserID   *uuid.UUID             `json:"target_user_id,omitempty"`
	EntityID       uuid.UUID              `json:"entity_id"`
	RequestType    string                 `json:"request_type"` // ROLE_ASSIGNMENT, PERMISSION_GRANT, RESOURCE_ACCESS, ELEVATION
	RoleID         *uuid.UUID             `json:"role_id,omitempty"`
	PermissionID   *uuid.UUID             `json:"permission_id,omitempty"`
	ResourceID     *uuid.UUID             `json:"resource_id,omitempty"`
	Justification  string                 `json:"justification"`
	BusinessReason string                 `json:"business_reason,omitempty"`
	DurationHours  *int                   `json:"duration_hours,omitempty"`
	Context        map[string]interface{} `json:"context,omitempty"`
}

// AccessRequestWorkflowResult represents the result of access request workflow
type AccessRequestWorkflowResult struct {
	RequestID     uuid.UUID `json:"request_id"`
	Success       bool      `json:"success"`
	ApprovalStatus string   `json:"approval_status"` // PENDING, APPROVED, REJECTED, EXPIRED, REVOKED
	ApprovedBy    *uuid.UUID `json:"approved_by,omitempty"`
	GrantedAt     *time.Time `json:"granted_at,omitempty"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	ErrorMessage  string    `json:"error_message,omitempty"`
}

// PolicyManagementWorkflowRequest represents a workflow request for policy management
type PolicyManagementWorkflowRequest struct {
	RequestID    uuid.UUID              `json:"request_id"`
	Operation    string                 `json:"operation"` // CREATE, UPDATE, DELETE, TEST
	PolicyID     *uuid.UUID             `json:"policy_id,omitempty"`
	Policy       map[string]interface{} `json:"policy,omitempty"`
	TestContext  map[string]interface{} `json:"test_context,omitempty"`
	CreatedBy    uuid.UUID              `json:"created_by"`
}

// --- Activity Models ---

// AttributeCollectionResult represents the result of attribute collection
type AttributeCollectionResult struct {
	UserAttributes        map[string]interface{} `json:"user_attributes"`
	PersonAttributes      map[string]interface{} `json:"person_attributes,omitempty"`
	EmployeeAttributes    map[string]interface{} `json:"employee_attributes,omitempty"`
	ResourceAttributes    map[string]interface{} `json:"resource_attributes,omitempty"`
	EnvironmentAttributes map[string]interface{} `json:"environment_attributes"`
	CollectionTimeMS      int64                  `json:"collection_time_ms"`
}

// PolicyEvaluationResult represents the result of policy evaluation
type PolicyEvaluationResult struct {
	PolicyID         uuid.UUID              `json:"policy_id"`
	PolicyName       string                 `json:"policy_name"`
	Effect           string                 `json:"effect"` // ALLOW, DENY, NOT_APPLICABLE
	Decision         string                 `json:"decision"`
	TargetMatches    bool                   `json:"target_matches"`
	RuleResult       bool                   `json:"rule_result"`
	Priority         int                    `json:"priority"`
	EvaluationTimeMS int64                  `json:"evaluation_time_ms"`
	Details          map[string]interface{} `json:"details,omitempty"`
	ErrorMessage     string                 `json:"error_message,omitempty"`
}

// ContextEnrichmentResult represents the result of context enrichment
type ContextEnrichmentResult struct {
	DeviceContext    map[string]interface{} `json:"device_context,omitempty"`
	LocationContext  map[string]interface{} `json:"location_context,omitempty"`
	TimeContext      map[string]interface{} `json:"time_context"`
	NetworkContext   map[string]interface{} `json:"network_context,omitempty"`
	RiskContext      map[string]interface{} `json:"risk_context,omitempty"`
	EnrichmentTimeMS int64                  `json:"enrichment_time_ms"`
}

// PolicyEvaluationCacheEntry represents a cached policy evaluation result
type PolicyEvaluationCacheEntry struct {
	UserID           uuid.UUID `json:"user_id"`
	ResourceName     string    `json:"resource_name"`
	ActionName       string    `json:"action_name"`
	ContextHash      string    `json:"context_hash"`
	Decision         string    `json:"decision"`
	PolicyDecisions  []string  `json:"policy_decisions"`
	EvaluationTimeMS int64     `json:"evaluation_time_ms"`
	EvaluatedAt      time.Time `json:"evaluated_at"`
	ExpiresAt        time.Time `json:"expires_at"`
}