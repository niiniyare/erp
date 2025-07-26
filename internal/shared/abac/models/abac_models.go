package abac

import (
	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/shared/utils"
	"time"
)

// Task Queues
const (
	PermissionEvaluationTaskQueue = "ABAC_PERMISSION_EVALUATION_TASK_QUEUE"
	AccessRequestTaskQueue        = "ABAC_ACCESS_REQUEST_TASK_QUEUE"
	PolicyManagementTaskQueue     = "ABAC_POLICY_MANAGEMENT_TASK_QUEUE"
)

// Workflow Request/Response Models

// PermissionEvaluationWorkflowRequest is the input for the permission evaluation workflow.
type PermissionEvaluationWorkflowRequest struct {
	UserID     uuid.UUID              `json:"user_id" validate:"required,uuid"`
	ResourceID uuid.UUID              `json:"resource_id" validate:"required,uuid"`
	ActionID   uuid.UUID              `json:"action_id" validate:"required,uuid"`
	EntityID   *uuid.UUID             `json:"entity_id,omitempty" validate:"omitempty,uuid"`
	Context    map[string]interface{} `json:"context,omitempty"` // Additional context attributes
	RequestID  string                 `json:"request_id" validate:"required"`
}

// Validate performs validation on the request struct.
func (r *PermissionEvaluationWorkflowRequest) Validate() error {
	return utils.ValidateStruct(r)
}

// PermissionEvaluationWorkflowResponse is the result of the permission evaluation workflow.
type PermissionEvaluationWorkflowResponse struct {
	Allowed          bool             `json:"allowed"`
	PolicyDecisions  []PolicyDecision `json:"policy_decisions"`
	EffectiveRoles   []string         `json:"effective_roles"`
	EvaluationTimeMS int              `json:"evaluation_time_ms"`
	CacheHit         bool             `json:"cache_hit"`
	Error            string           `json:"error,omitempty"`
}

// PolicyDecision represents a single policy's evaluation outcome.
type PolicyDecision struct {
	PolicyID   uuid.UUID              `json:"policy_id"`
	PolicyName string                 `json:"policy_name"`
	Effect     string                 `json:"effect"` // "ALLOW", "DENY", "NOT_APPLICABLE"
	Reason     string                 `json:"reason,omitempty"`
	Details    map[string]interface{} `json:"details,omitempty"`
}

// AccessRequestWorkflowRequest is the input for the access request workflow.
type AccessRequestWorkflowRequest struct {
	RequesterID    uuid.UUID  `json:"requester_id" validate:"required,uuid"`
	TargetUserID   *uuid.UUID `json:"target_user_id,omitempty" validate:"omitempty,uuid"`
	EntityID       uuid.UUID  `json:"entity_id" validate:"required,uuid"`
	RequestType    string     `json:"request_type" validate:"required"` // e.g., "ROLE_ASSIGNMENT", "PERMISSION_GRANT"
	RoleID         *uuid.UUID `json:"role_id,omitempty" validate:"omitempty,uuid"`
	PermissionID   *uuid.UUID `json:"permission_id,omitempty" validate:"omitempty,uuid"`
	ResourceID     *uuid.UUID `json:"resource_id,omitempty" validate:"omitempty,uuid"`
	Justification  string     `json:"justification" validate:"required,min=10"`
	BusinessReason *string    `json:"business_reason,omitempty"`
	DurationHours  *int32     `json:"duration_hours,omitempty" validate:"omitempty,gt=0"`
	RequestID      string     `json:"request_id" validate:"required"`
}

// Validate performs validation on the request struct.
func (r *AccessRequestWorkflowRequest) Validate() error {
	return utils.ValidateStruct(r)
}

// AccessRequestWorkflowResponse is the result of the access request workflow.
type AccessRequestWorkflowResponse struct {
	AccessRequestID uuid.UUID `json:"access_request_id"`
	Status          string    `json:"status"` // "PENDING", "APPROVED", "REJECTED", "EXPIRED"
	Message         string    `json:"message,omitempty"`
	Error           string    `json:"error,omitempty"`
}

// PolicyTestWorkflowRequest is the input for testing a policy.
type PolicyTestWorkflowRequest struct {
	PolicyID    uuid.UUID              `json:"policy_id" validate:"required,uuid"`
	TestContext map[string]interface{} `json:"test_context" validate:"required"` // Attributes to test against
	RequestID   string                 `json:"request_id" validate:"required"`
}

// Validate performs validation on the request struct.
func (r *PolicyTestWorkflowRequest) Validate() error {
	return utils.ValidateStruct(r)
}

// PolicyTestWorkflowResponse is the result of testing a policy.
type PolicyTestWorkflowResponse struct {
	PolicyID         uuid.UUID      `json:"policy_id"`
	PolicyName       string         `json:"policy_name"`
	EvaluationResult PolicyDecision `json:"evaluation_result"`
	EvaluationTimeMS int            `json:"evaluation_time_ms"`
	Error            string         `json:"error,omitempty"`
}

// Activity Input/Output Models

// CollectUserAttributesActivityInput is the input for collecting user attributes.
type CollectUserAttributesActivityInput struct {
	UserID   uuid.UUID `json:"user_id"`
	TenantID uuid.UUID `json:"tenant_id"`
}

// CollectUserAttributesActivityOutput is the output of collecting user attributes.
type CollectUserAttributesActivityOutput struct {
	Attributes map[string]interface{} `json:"attributes"`
}

// CollectResourceAttributesActivityInput is the input for collecting resource attributes.
type CollectResourceAttributesActivityInput struct {
	ResourceID uuid.UUID `json:"resource_id"`
	TenantID   uuid.UUID `json:"tenant_id"`
}

// CollectResourceAttributesActivityOutput is the output of collecting resource attributes.
type CollectResourceAttributesActivityOutput struct {
	Attributes map[string]interface{} `json:"attributes"`
}

// BuildEnvironmentContextActivityInput is the input for building environment context.
type BuildEnvironmentContextActivityInput struct {
	IPAddress string                 `json:"ip_address,omitempty"`
	UserAgent string                 `json:"user_agent,omitempty"`
	Location  map[string]interface{} `json:"location,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// BuildEnvironmentContextActivityOutput is the output of building environment context.
type BuildEnvironmentContextActivityOutput struct {
	Context map[string]interface{} `json:"context"`
}

// EvaluatePoliciesActivityInput is the input for evaluating policies.
type EvaluatePoliciesActivityInput struct {
	TenantID   uuid.UUID              `json:"tenant_id"`
	ResourceID uuid.UUID              `json:"resource_id"`
	ActionID   uuid.UUID              `json:"action_id"`
	Attributes map[string]interface{} `json:"attributes"` // Combined user, resource, environment attributes
}

// EvaluatePoliciesActivityOutput is the output of evaluating policies.
type EvaluatePoliciesActivityOutput struct {
	Allowed         bool             `json:"allowed"`
	PolicyDecisions []PolicyDecision `json:"policy_decisions"`
}

// CachePolicyResultActivityInput is the input for caching policy results.
type CachePolicyResultActivityInput struct {
	TenantID           uuid.UUID   `json:"tenant_id"`
	UserID             uuid.UUID   `json:"user_id"`
	ResourceID         uuid.UUID   `json:"resource_id"`
	ActionID           uuid.UUID   `json:"action_id"`
	ContextHash        string      `json:"context_hash"`
	Decision           string      `json:"decision"` // "ALLOW", "DENY", "NOT_APPLICABLE"
	ApplicablePolicies []uuid.UUID `json:"applicable_policies"`
	EvaluationTimeMS   int         `json:"evaluation_time_ms"`
	ExpiresAt          time.Time   `json:"expires_at"`
}

// CachePolicyResultActivityOutput is the output of caching policy results.
type CachePolicyResultActivityOutput struct {
	Success bool `json:"success"`
}
