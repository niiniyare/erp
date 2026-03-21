package authz

//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"

import (
	"context"
	"time"

	"github.com/google/uuid"

	"awo/internal/core/iam/model"
)

// Service defines the authorization service interface
// Handles ABAC, RBAC, hybrid access control, and permission evaluation
type Service interface {
	// Permission Evaluation
	EvaluatePermission(ctx context.Context, req *PermissionEvaluationRequest) (*PermissionEvaluationResult, error)
	BulkEvaluatePermissions(ctx context.Context, req *BulkPermissionEvaluationRequest) (*BulkPermissionEvaluationResult, error)

	// User Permissions
	GetUserEffectivePermissions(ctx context.Context, userID uuid.UUID, entityID *uuid.UUID) (*UserEffectivePermissions, error)
	CalculateRoleHierarchy(ctx context.Context, userID uuid.UUID, entityID *uuid.UUID) (*RoleHierarchy, error)

	// Access Requests
	CreateAccessRequest(ctx context.Context, req *CreateAccessRequestRequest) (*model.AccessRequest, error)
	GetAccessRequest(ctx context.Context, requestID uuid.UUID) (*model.AccessRequest, error)
	ProcessAccessRequest(ctx context.Context, req *ProcessAccessRequestRequest) error
	ListAccessRequests(ctx context.Context, req *ListAccessRequestsRequest) (*ListAccessRequestsResult, error)

	// Approval Workflows
	CreateApprovalWorkflow(ctx context.Context, req *CreateApprovalWorkflowRequest) (*model.ApprovalWorkflow, error)
	GetApprovalWorkflow(ctx context.Context, workflowID uuid.UUID) (*model.ApprovalWorkflow, error)
	UpdateApprovalWorkflow(ctx context.Context, req *UpdateApprovalWorkflowRequest) error

	// Conditional Access
	EvaluateConditionalAccess(ctx context.Context, req *ConditionalAccessRequest) (*ConditionalAccessResult, error)
	CreateConditionalAccessPolicy(ctx context.Context, req *CreateConditionalAccessPolicyRequest) (*model.ConditionalAccessPolicy, error)
	UpdateConditionalAccessPolicy(ctx context.Context, req *UpdateConditionalAccessPolicyRequest) error

	// Permission Management
	GrantPermission(ctx context.Context, req *GrantPermissionRequest) error
	RevokePermission(ctx context.Context, req *RevokePermissionRequest) error
	ListUserPermissions(ctx context.Context, userID uuid.UUID, entityID *uuid.UUID) ([]*model.Permission, error)

	// Decision History and Audit
	GetDecisionHistory(ctx context.Context, userID uuid.UUID, limit int) ([]*DecisionHistoryEntry, error)

	// Cache Management
	InvalidateUserCache(ctx context.Context, userID uuid.UUID) error
	InvalidatePolicyCache(ctx context.Context, policyIDs []uuid.UUID) error
	GetCacheStatistics(ctx context.Context) (*CacheStatistics, error)
}

// PermissionEvaluationRequest is a Request/Response types for Permission Evaluation
type PermissionEvaluationRequest struct {
	UserID       uuid.UUID      `json:"user_id" validate:"required"`
	ResourceType string         `json:"resource_type" validate:"required"`
	ResourceID   *uuid.UUID     `json:"resource_id,omitempty"`
	Action       string         `json:"action" validate:"required"`
	EntityID     *uuid.UUID     `json:"entity_id,omitempty"`
	Context      map[string]any `json:"context,omitempty"`
	RequestID    string         `json:"request_id,omitempty"`
}

type PermissionEvaluationResult struct {
	Decision         model.PolicyDecisionType `json:"decision"`
	PolicyDecisions  []*model.PolicyDecision  `json:"policy_decisions"`
	EvaluationTimeMS int64                    `json:"evaluation_time_ms"`
	CacheHit         bool                     `json:"cache_hit"`
	RequestID        string                   `json:"request_id"`
	Timestamp        time.Time                `json:"timestamp"`
}

type BulkPermissionEvaluationRequest struct {
	Requests  []*PermissionEvaluationRequest `json:"requests" validate:"required,min=1,max=100"`
	RequestID string                         `json:"request_id,omitempty"`
}

type BulkPermissionEvaluationResult struct {
	Results         []*PermissionEvaluationResult `json:"results"`
	TotalRequests   int                           `json:"total_requests"`
	SuccessfulCount int                           `json:"successful_count"`
	FailedCount     int                           `json:"failed_count"`
	TotalTimeMS     int64                         `json:"total_time_ms"`
	AverageTimeMS   float64                       `json:"average_time_ms"`
	RequestID       string                        `json:"request_id"`
	Timestamp       time.Time                     `json:"timestamp"`
}

// UserEffectivePermissions is User Permissions types
type UserEffectivePermissions struct {
	UserID      uuid.UUID           `json:"user_id"`
	EntityID    *uuid.UUID          `json:"entity_id,omitempty"`
	Permissions map[string][]string `json:"permissions"` // resource_type -> actions
	Roles       []string            `json:"roles"`
	Timestamp   time.Time           `json:"timestamp"`
}

type RoleHierarchy struct {
	UserID    uuid.UUID           `json:"user_id"`
	EntityID  *uuid.UUID          `json:"entity_id,omitempty"`
	Roles     []string            `json:"roles"`
	Hierarchy map[string][]string `json:"hierarchy"` // role -> inherited roles
	Timestamp time.Time           `json:"timestamp"`
}

// CreateAccessRequestRequest is a  Access Request types
type CreateAccessRequestRequest struct {
	UserID        uuid.UUID      `json:"user_id" validate:"required"`
	ResourceType  string         `json:"resource_type" validate:"required"`
	ResourceID    *uuid.UUID     `json:"resource_id,omitempty"`
	Action        string         `json:"action" validate:"required"`
	Justification string         `json:"justification" validate:"required"`
	Duration      *time.Duration `json:"duration,omitempty"`
	Priority      string         `json:"priority" validate:"oneof=low medium high urgent"`
	Metadata      map[string]any `json:"metadata,omitempty"`
}

type ProcessAccessRequestRequest struct {
	RequestID  uuid.UUID `json:"request_id" validate:"required"`
	Action     string    `json:"action" validate:"required,oneof=approve reject"`
	ApproverID uuid.UUID `json:"approver_id" validate:"required"`
	Comments   string    `json:"comments,omitempty"`
	Conditions []string  `json:"conditions,omitempty"`
}

type ListAccessRequestsRequest struct {
	UserID   *uuid.UUID `json:"user_id,omitempty"`
	Status   *string    `json:"status,omitempty"`
	EntityID *uuid.UUID `json:"entity_id,omitempty"`
	Limit    int        `json:"limit" validate:"min=1,max=100"`
	Offset   int        `json:"offset" validate:"min=0"`
}

type ListAccessRequestsResult struct {
	Requests []*model.AccessRequest `json:"requests"`
	Total    int                    `json:"total"`
	Limit    int                    `json:"limit"`
	Offset   int                    `json:"offset"`
	HasMore  bool                   `json:"has_more"`
}

// CreateApprovalWorkflowRequest Approval Workflow types
type CreateApprovalWorkflowRequest struct {
	Name        string                `json:"name" validate:"required"`
	Description string                `json:"description"`
	Steps       []*model.ApprovalStep `json:"steps" validate:"required,min=1"`
	Metadata    map[string]any        `json:"metadata,omitempty"`
}

type UpdateApprovalWorkflowRequest struct {
	WorkflowID  uuid.UUID             `json:"workflow_id" validate:"required"`
	Name        *string               `json:"name,omitempty"`
	Description *string               `json:"description,omitempty"`
	Steps       []*model.ApprovalStep `json:"steps,omitempty"`
	Metadata    map[string]any        `json:"metadata,omitempty"`
}

// ConditionalAccessRequest Conditional Access types
type ConditionalAccessRequest struct {
	UserID       uuid.UUID            `json:"user_id" validate:"required"`
	ResourceType string               `json:"resource_type" validate:"required"`
	ResourceID   *uuid.UUID           `json:"resource_id,omitempty"`
	Action       string               `json:"action" validate:"required"`
	Context      *model.AccessContext `json:"context,omitempty"`
}

type ConditionalAccessResult struct {
	Allowed    bool                     `json:"allowed"`
	Conditions []*model.AccessCondition `json:"conditions,omitempty"`
	Reason     string                   `json:"reason,omitempty"`
	Metadata   map[string]any           `json:"metadata,omitempty"`
}

type CreateConditionalAccessPolicyRequest struct {
	Name        string                   `json:"name" validate:"required"`
	Description string                   `json:"description"`
	Conditions  []*model.PolicyCondition `json:"conditions" validate:"required,min=1"`
	Actions     []*model.PolicyAction    `json:"actions" validate:"required,min=1"`
	Priority    int                      `json:"priority" validate:"min=0"`
	Enabled     bool                     `json:"enabled"`
	Metadata    map[string]any           `json:"metadata,omitempty"`
}

type UpdateConditionalAccessPolicyRequest struct {
	PolicyID    uuid.UUID                `json:"policy_id" validate:"required"`
	Name        *string                  `json:"name,omitempty"`
	Description *string                  `json:"description,omitempty"`
	Conditions  []*model.PolicyCondition `json:"conditions,omitempty"`
	Actions     []*model.PolicyAction    `json:"actions,omitempty"`
	Priority    *int                     `json:"priority,omitempty"`
	Enabled     *bool                    `json:"enabled,omitempty"`
	Metadata    map[string]any           `json:"metadata,omitempty"`
}

// GrantPermissionRequest Permission Management types
type GrantPermissionRequest struct {
	UserID       uuid.UUID  `json:"user_id" validate:"required"`
	ResourceType string     `json:"resource_type" validate:"required"`
	ResourceID   *uuid.UUID `json:"resource_id,omitempty"`
	Action       string     `json:"action" validate:"required"`
	EntityID     *uuid.UUID `json:"entity_id,omitempty"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	Conditions   []string   `json:"conditions,omitempty"`
}

type RevokePermissionRequest struct {
	UserID       uuid.UUID  `json:"user_id" validate:"required"`
	ResourceType string     `json:"resource_type" validate:"required"`
	ResourceID   *uuid.UUID `json:"resource_id,omitempty"`
	Action       string     `json:"action" validate:"required"`
	EntityID     *uuid.UUID `json:"entity_id,omitempty"`
}

// DecisionHistoryEntry  History types
type DecisionHistoryEntry struct {
	ID             uuid.UUID                `json:"id"`
	UserID         uuid.UUID                `json:"user_id"`
	ResourceType   string                   `json:"resource_type"`
	ResourceID     *uuid.UUID               `json:"resource_id,omitempty"`
	Action         string                   `json:"action"`
	Decision       model.PolicyDecisionType `json:"decision"`
	Allowed        bool                     `json:"allowed"`
	EvaluationTime time.Duration            `json:"evaluation_time"`
	EvaluatedAt    time.Time                `json:"evaluated_at"`
	PolicyCount    int                      `json:"policy_count"`
	CacheHit       bool                     `json:"cache_hit"`
	RequestID      string                   `json:"request_id"`
}

// CacheStatistics  Statistics types
type CacheStatistics struct {
	HitRate        float64 `json:"hit_rate"`
	MissRate       float64 `json:"miss_rate"`
	TotalRequests  int64   `json:"total_requests"`
	CacheHits      int64   `json:"cache_hits"`
	CacheMisses    int64   `json:"cache_misses"`
	EvictionCount  int64   `json:"eviction_count"`
	AverageLatency int64   `json:"average_latency_ms"`
}
