package access

//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/niiniyare/erp/internal/core/access/conditional"
	"github.com/niiniyare/erp/internal/shared/types"
)

// Service defines the access management service interface
// This aggregates all access-related functionality
type Service interface {
	// Access Request Management
	CreateAccessRequest(ctx context.Context, req *CreateAccessRequestRequest) (*AccessRequest, error)
	GetAccessRequest(ctx context.Context, requestID uuid.UUID) (*AccessRequest, error)
	ProcessAccessRequest(ctx context.Context, req *ProcessAccessRequestRequest) error
	ListAccessRequests(ctx context.Context, req *ListAccessRequestsRequest) (*ListAccessRequestsResult, error)

	// Approval Workflow Management
	CreateApprovalWorkflow(ctx context.Context, req *CreateApprovalWorkflowRequest) (*ApprovalWorkflow, error)
	GetApprovalWorkflow(ctx context.Context, workflowID uuid.UUID) (*ApprovalWorkflow, error)
	UpdateApprovalWorkflow(ctx context.Context, req *UpdateApprovalWorkflowRequest) error

	// Conditional Access
	EvaluateConditionalAccess(ctx context.Context, req *EvaluateConditionalAccessRequest) (*ConditionalAccessResult, error)
	CreateConditionalAccessPolicy(ctx context.Context, req *CreateConditionalAccessPolicyRequest) (*ConditionalAccessPolicy, error)
	UpdateConditionalAccessPolicy(ctx context.Context, req *UpdateConditionalAccessPolicyRequest) error

	// Permission Management
	GrantPermission(ctx context.Context, req *GrantPermissionRequest) error
	RevokePermission(ctx context.Context, req *RevokePermissionRequest) error
	ListUserPermissions(ctx context.Context, userID uuid.UUID, entityID *uuid.UUID) ([]*Permission, error)
}

// ─── ACCESS REQUEST TYPES ──────────────────────────────────────────────────

// Type aliases from request package
type AccessRequest = types.AccessRequest
type RequestType = types.RequestType
type ApprovalStatus = types.ApprovalStatus

// CreateAccessRequestRequest represents a request to create an access request
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

// ProcessAccessRequestRequest represents a request to process an access request
type ProcessAccessRequestRequest struct {
	RequestID  uuid.UUID `json:"request_id" validate:"required"`
	Action     string    `json:"action" validate:"required,oneof=approve reject"`
	ApproverID uuid.UUID `json:"approver_id" validate:"required"`
	Comments   string    `json:"comments,omitempty"`
	Conditions []string  `json:"conditions,omitempty"`
}

// ListAccessRequestsRequest represents a request to list access requests
type ListAccessRequestsRequest struct {
	UserID   *uuid.UUID `json:"user_id,omitempty"`
	Status   *string    `json:"status,omitempty"`
	EntityID *uuid.UUID `json:"entity_id,omitempty"`
	Limit    int        `json:"limit" validate:"min=1,max=100"`
	Offset   int        `json:"offset" validate:"min=0"`
}

// ListAccessRequestsResult represents the result of listing access requests
type ListAccessRequestsResult struct {
	Requests []*AccessRequest `json:"requests"`
	Total    int              `json:"total"`
	Limit    int              `json:"limit"`
	Offset   int              `json:"offset"`
	HasMore  bool             `json:"has_more"`
}

// ─── APPROVAL WORKFLOW TYPES ───────────────────────────────────────────────

// ApprovalWorkflow represents an approval workflow
type ApprovalWorkflow struct {
	ID          uuid.UUID       `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Steps       []*ApprovalStep `json:"steps"`
	IsActive    bool            `json:"is_active"`
	CreatedBy   uuid.UUID       `json:"created_by"`
	Metadata    map[string]any  `json:"metadata,omitempty"`
	EntityID    *uuid.UUID      `json:"entity_id,omitempty"`
	TenantID    uuid.UUID       `json:"tenant_id"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// ApprovalStep represents a step in an approval workflow
type ApprovalStep struct {
	ID             uuid.UUID            `json:"id"`
	Order          int                  `json:"order"`
	Name           string               `json:"name"`
	Description    string               `json:"description"`
	ApproverType   string               `json:"approver_type"`
	ApproverIDs    []uuid.UUID          `json:"approver_ids"`
	RequiredCount  int                  `json:"required_count"`
	TimeoutMinutes int                  `json:"timeout_minutes"`
	Conditions     []*ApprovalCondition `json:"conditions,omitempty"`
	Metadata       map[string]any       `json:"metadata,omitempty"`
}

// ApprovalCondition represents a condition for approval
type ApprovalCondition struct {
	Type        string `json:"type"`
	Operator    string `json:"operator"`
	Value       any    `json:"value"`
	Description string `json:"description"`
}

// CreateApprovalWorkflowRequest represents a request to create an approval workflow
type CreateApprovalWorkflowRequest struct {
	Name        string          `json:"name" validate:"required"`
	Description string          `json:"description"`
	Steps       []*ApprovalStep `json:"steps" validate:"required,min=1"`
	Metadata    map[string]any  `json:"metadata,omitempty"`
}

// UpdateApprovalWorkflowRequest represents a request to update an approval workflow
type UpdateApprovalWorkflowRequest struct {
	WorkflowID  uuid.UUID       `json:"workflow_id" validate:"required"`
	Name        *string         `json:"name,omitempty"`
	Description *string         `json:"description,omitempty"`
	Steps       []*ApprovalStep `json:"steps,omitempty"`
	Metadata    map[string]any  `json:"metadata,omitempty"`
}

// ─── CONDITIONAL ACCESS TYPES ──────────────────────────────────────────────

// Type aliases from conditional package
type ConditionalAccessPolicy = conditional.ConditionalAccessPolicy

// Define missing types locally
type PolicyCondition struct {
	Type       string         `json:"type"`
	Operator   string         `json:"operator"`
	Value      any            `json:"value"`
	Expression string         `json:"expression,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

type PolicyAction struct {
	Type        string         `json:"type"`
	Value       any            `json:"value"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type AccessContext struct {
	IPAddress       string           `json:"ip_address,omitempty"`
	UserAgent       string           `json:"user_agent,omitempty"`
	DeviceInfo      *DeviceInfo      `json:"device_info,omitempty"`
	LocationInfo    *LocationInfo    `json:"location_info,omitempty"`
	SecurityContext *SecurityContext `json:"security_context,omitempty"`
	TimeContext     *TimeContext     `json:"time_context,omitempty"`
	Metadata        map[string]any   `json:"metadata,omitempty"`
}

type DeviceInfo struct {
	DeviceID       string `json:"device_id"`
	DeviceType     string `json:"device_type"`
	OS             string `json:"os"`
	OSVersion      string `json:"os_version"`
	Browser        string `json:"browser"`
	BrowserVersion string `json:"browser_version"`
	IsTrusted      bool   `json:"is_trusted"`
	IsManaged      bool   `json:"is_managed"`
}

type LocationInfo struct {
	Country    string         `json:"country"`
	Region     string         `json:"region"`
	City       string         `json:"city"`
	Latitude   *float64       `json:"latitude,omitempty"`
	Longitude  *float64       `json:"longitude,omitempty"`
	TimeZone   string         `json:"time_zone"`
	IsTrusted  bool           `json:"is_trusted"`
	RiskLevel  string         `json:"risk_level"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

type SecurityContext struct {
	ThreatLevel         string         `json:"threat_level"`
	AuthenticationLevel string         `json:"authentication_level"`
	EncryptionLevel     string         `json:"encryption_level"`
	SecurityFlags       []string       `json:"security_flags"`
	RiskScore           float64        `json:"risk_score"`
	Attributes          map[string]any `json:"attributes"`
}

type TimeContext struct {
	RequestTime   time.Time      `json:"request_time"`
	TimeZone      string         `json:"time_zone"`
	BusinessHours *BusinessHours `json:"business_hours,omitempty"`
	IsWeekend     bool           `json:"is_weekend"`
	IsHoliday     bool           `json:"is_holiday"`
}

type BusinessHours struct {
	Start string   `json:"start"`
	End   string   `json:"end"`
	Days  []string `json:"days"`
}

type AccessCondition struct {
	Type        string         `json:"type"`
	Operator    string         `json:"operator"`
	Value       any            `json:"value"`
	Description string         `json:"description"`
	IsMandatory bool           `json:"is_mandatory"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// EvaluateConditionalAccessRequest represents a request to evaluate conditional access
type EvaluateConditionalAccessRequest struct {
	UserID       uuid.UUID      `json:"user_id" validate:"required"`
	ResourceType string         `json:"resource_type" validate:"required"`
	ResourceID   *uuid.UUID     `json:"resource_id,omitempty"`
	Action       string         `json:"action" validate:"required"`
	Context      *AccessContext `json:"context,omitempty"`
}

// ConditionalAccessResult represents the result of conditional access evaluation
type ConditionalAccessResult struct {
	Allowed    bool               `json:"allowed"`
	Conditions []*AccessCondition `json:"conditions,omitempty"`
	Reason     string             `json:"reason,omitempty"`
	Metadata   map[string]any     `json:"metadata,omitempty"`
}

// CreateConditionalAccessPolicyRequest represents a request to create a conditional access policy
type CreateConditionalAccessPolicyRequest struct {
	Name        string             `json:"name" validate:"required"`
	Description string             `json:"description"`
	Conditions  []*PolicyCondition `json:"conditions" validate:"required,min=1"`
	Actions     []*PolicyAction    `json:"actions" validate:"required,min=1"`
	Priority    int                `json:"priority" validate:"min=0"`
	Enabled     bool               `json:"enabled"`
	Metadata    map[string]any     `json:"metadata,omitempty"`
}

// UpdateConditionalAccessPolicyRequest represents a request to update a conditional access policy
type UpdateConditionalAccessPolicyRequest struct {
	PolicyID    uuid.UUID          `json:"policy_id" validate:"required"`
	Name        *string            `json:"name,omitempty"`
	Description *string            `json:"description,omitempty"`
	Conditions  []*PolicyCondition `json:"conditions,omitempty"`
	Actions     []*PolicyAction    `json:"actions,omitempty"`
	Priority    *int               `json:"priority,omitempty"`
	Enabled     *bool              `json:"enabled,omitempty"`
	Metadata    map[string]any     `json:"metadata,omitempty"`
}

// ─── PERMISSION MANAGEMENT TYPES ───────────────────────────────────────────

// Permission represents a permission in the system
type Permission struct {
	ID           uuid.UUID      `json:"id"`
	UserID       uuid.UUID      `json:"user_id"`
	ResourceType string         `json:"resource_type"`
	ResourceID   *uuid.UUID     `json:"resource_id,omitempty"`
	Action       string         `json:"action"`
	Effect       string         `json:"effect"`
	Conditions   []string       `json:"conditions,omitempty"`
	GrantedBy    *uuid.UUID     `json:"granted_by,omitempty"`
	GrantedAt    *time.Time     `json:"granted_at,omitempty"`
	ExpiresAt    *time.Time     `json:"expires_at,omitempty"`
	IsActive     bool           `json:"is_active"`
	Metadata     map[string]any `json:"metadata,omitempty"`
	EntityID     *uuid.UUID     `json:"entity_id,omitempty"`
	TenantID     uuid.UUID      `json:"tenant_id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// GrantPermissionRequest represents a request to grant permission
type GrantPermissionRequest struct {
	UserID       uuid.UUID  `json:"user_id" validate:"required"`
	ResourceType string     `json:"resource_type" validate:"required"`
	ResourceID   *uuid.UUID `json:"resource_id,omitempty"`
	Action       string     `json:"action" validate:"required"`
	EntityID     *uuid.UUID `json:"entity_id,omitempty"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	Conditions   []string   `json:"conditions,omitempty"`
}

// RevokePermissionRequest represents a request to revoke permission
type RevokePermissionRequest struct {
	UserID       uuid.UUID  `json:"user_id" validate:"required"`
	ResourceType string     `json:"resource_type" validate:"required"`
	ResourceID   *uuid.UUID `json:"resource_id,omitempty"`
	Action       string     `json:"action" validate:"required"`
	EntityID     *uuid.UUID `json:"entity_id,omitempty"`
}
