package workflow

import (
	"time"

	"github.com/google/uuid"
)

// FeatureFlagChangeRequest represents a request to change a feature flag
type FeatureFlagChangeRequest struct {
	TenantID             uuid.UUID      `json:"tenant_id"`
	RequestedBy          uuid.UUID      `json:"requested_by"`
	FlagName             string         `json:"flag_name"`
	ChangeType           string         `json:"change_type"` // enable, disable, update_rollout, update_config
	NewValue             any            `json:"new_value"`
	Justification        string         `json:"justification"`
	BusinessReason       string         `json:"business_reason"`
	ApprovalTimeoutHours int            `json:"approval_timeout_hours"`
	Metadata             map[string]any `json:"metadata,omitempty"`
}

// BulkFeatureFlagChangeRequest represents a bulk change request
type BulkFeatureFlagChangeRequest struct {
	TenantID             uuid.UUID               `json:"tenant_id"`
	RequestedBy          uuid.UUID               `json:"requested_by"`
	Changes              []*IndividualFlagChange `json:"changes"`
	Justification        string                  `json:"justification"`
	BusinessReason       string                  `json:"business_reason"`
	ApprovalTimeoutHours int                     `json:"approval_timeout_hours"`
}

// IndividualFlagChange represents a single flag change in bulk operation
type IndividualFlagChange struct {
	FlagName   string         `json:"flag_name"`
	ChangeType string         `json:"change_type"`
	NewValue   any            `json:"new_value"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

// FeatureFlagChangeResult represents the result of a flag change workflow
type FeatureFlagChangeResult struct {
	Status          string          `json:"status"` // completed, rejected, failed, validation_failed, approval_timeout
	FlagID          *uuid.UUID      `json:"flag_id,omitempty"`
	OldValue        any             `json:"old_value,omitempty"`
	NewValue        any             `json:"new_value,omitempty"`
	AppliedAt       *time.Time      `json:"applied_at,omitempty"`
	ApprovalDetails *ApprovalResult `json:"approval_details,omitempty"`
	Error           string          `json:"error,omitempty"`
}

// BulkFeatureFlagChangeResult represents the result of bulk changes
type BulkFeatureFlagChangeResult struct {
	TotalCount   int                                 `json:"total_count"`
	SuccessCount int                                 `json:"success_count"`
	FailureCount int                                 `json:"failure_count"`
	Results      map[string]*FeatureFlagChangeResult `json:"results"`
}

// ValidationResult represents flag validation result
type ValidationResult struct {
	IsValid      bool   `json:"is_valid"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// CreateAccessRequestResult represents access request creation result
type CreateAccessRequestResult struct {
	RequestID uuid.UUID `json:"request_id"`
}

// ApprovalResult represents the approval decision
type ApprovalResult struct {
	Approved   bool      `json:"approved"`
	ApproverID uuid.UUID `json:"approver_id"`
	Comments   string    `json:"comments,omitempty"`
	ApprovedAt time.Time `json:"approved_at"`
}

// ApprovalSignal represents the signal sent for approval decisions
type ApprovalSignal struct {
	Approved   bool      `json:"approved"`
	ApproverID uuid.UUID `json:"approver_id"`
	Comments   string    `json:"comments,omitempty"`
	ApprovedAt time.Time `json:"approved_at"`
}

// ApplyChangeResult represents the result of applying a flag change
type ApplyChangeResult struct {
	FlagID    uuid.UUID `json:"flag_id"`
	OldValue  any       `json:"old_value"`
	AppliedAt time.Time `json:"applied_at"`
}

// NotificationRequest represents a WebSocket notification request
type NotificationRequest struct {
	TenantID        uuid.UUID  `json:"tenant_id"`
	FlagName        string     `json:"flag_name"`
	ChangeType      string     `json:"change_type"`
	NewValue        any        `json:"new_value"`
	AppliedBy       uuid.UUID  `json:"applied_by"`
	AccessRequestID *uuid.UUID `json:"access_request_id,omitempty"`
	AppliedAt       time.Time  `json:"applied_at"`
}

// AuditEventRequest represents an audit event creation request
type AuditEventRequest struct {
	TenantID        uuid.UUID      `json:"tenant_id"`
	FlagName        string         `json:"flag_name"`
	ChangeType      string         `json:"change_type"`
	OldValue        any            `json:"old_value"`
	NewValue        any            `json:"new_value"`
	RequestedBy     uuid.UUID      `json:"requested_by"`
	AccessRequestID *uuid.UUID     `json:"access_request_id,omitempty"`
	AppliedAt       time.Time      `json:"applied_at"`
	Metadata        map[string]any `json:"metadata,omitempty"`
}

// AutoRollbackRequest represents an auto-rollback request
type AutoRollbackRequest struct {
	TenantID      uuid.UUID `json:"tenant_id"`
	FlagID        uuid.UUID `json:"flag_id"`
	FlagName      string    `json:"flag_name"`
	RollbackAt    time.Time `json:"rollback_at"`
	RollbackValue any       `json:"rollback_value"`
}

// AutoRollbackResult represents the result of auto-rollback
type AutoRollbackResult struct {
	Status       string     `json:"status"` // completed, failed, skipped
	FlagID       uuid.UUID  `json:"flag_id"`
	FlagName     string     `json:"flag_name"`
	ScheduledAt  time.Time  `json:"scheduled_at"`
	OldValue     any        `json:"old_value,omitempty"`
	NewValue     any        `json:"new_value,omitempty"`
	RolledBackAt *time.Time `json:"rolled_back_at,omitempty"`
	Reason       string     `json:"reason,omitempty"`
	Error        string     `json:"error,omitempty"`
}

// FeatureFlagApprovalPolicy represents approval requirements for flag changes
type FeatureFlagApprovalPolicy struct {
	TenantID             uuid.UUID   `json:"tenant_id"`
	FlagNamePattern      string      `json:"flag_name_pattern"` // regex pattern
	ChangeTypes          []string    `json:"change_types"`      // which change types require approval
	RequiresApproval     bool        `json:"requires_approval"`
	MinApprovers         int         `json:"min_approvers"`
	ApprovalTimeoutHours int         `json:"approval_timeout_hours"`
	AutoApproveFromUsers []uuid.UUID `json:"auto_approve_from_users,omitempty"`
	ApproverRoles        []string    `json:"approver_roles,omitempty"`
	ApproverUsers        []uuid.UUID `json:"approver_users,omitempty"`
}

// WebSocketMessage represents a real-time notification message
type WebSocketMessage struct {
	Type      string    `json:"type"`
	Event     string    `json:"event"`
	TenantID  uuid.UUID `json:"tenant_id"`
	Data      any       `json:"data"`
	Timestamp time.Time `json:"timestamp"`
}

// FeatureFlagChangeEvent represents a feature flag change event for WebSocket
type FeatureFlagChangeEvent struct {
	FlagID          uuid.UUID      `json:"flag_id"`
	FlagName        string         `json:"flag_name"`
	ChangeType      string         `json:"change_type"`
	OldValue        any            `json:"old_value,omitempty"`
	NewValue        any            `json:"new_value"`
	ChangedBy       uuid.UUID      `json:"changed_by"`
	AccessRequestID *uuid.UUID     `json:"access_request_id,omitempty"`
	AppliedAt       time.Time      `json:"applied_at"`
	Metadata        map[string]any `json:"metadata,omitempty"`
}
