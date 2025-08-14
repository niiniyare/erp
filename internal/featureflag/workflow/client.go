package workflow

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Client defines the interface for interacting with feature flag workflows.
// This abstracts the underlying workflow engine (e.g., Temporal).
type Client interface {
	// Workflow management
	RequestFeatureFlagChange(ctx context.Context, req *FeatureFlagChangeRequestInput) (*FeatureFlagWorkflowResult, error)
	RequestBulkFeatureFlagChange(ctx context.Context, req *BulkFeatureFlagChangeRequestInput) (*BulkFeatureFlagWorkflowResult, error)

	// Approval handling
	ApproveFeatureFlagChange(ctx context.Context, workflowID string, approvalReq *ApprovalRequestInput) error
	RejectFeatureFlagChange(ctx context.Context, workflowID string, rejectionReq *RejectionRequestInput) error

	// Workflow monitoring
	GetWorkflowStatus(ctx context.Context, workflowID string) (*WorkflowStatus, error)
	ListPendingApprovals(ctx context.Context, req *ListPendingApprovalsRequest) (*ListPendingApprovalsResponse, error)
	CancelWorkflow(ctx context.Context, workflowID string, reason string) error

	// Auto-rollback scheduling
	ScheduleAutoRollback(ctx context.Context, req *ScheduleAutoRollbackRequest) (*AutoRollbackScheduleResult, error)
	CancelAutoRollback(ctx context.Context, scheduleID string) error
}

// --- Data Transfer Objects (DTOs) for the Client interface ---

// FeatureFlagChangeRequestInput represents input for flag change workflow
type FeatureFlagChangeRequestInput struct {
	FlagName             string                 `json:"flag_name" validate:"required"`
	ChangeType           string                 `json:"change_type" validate:"required"`
	NewValue             interface{}            `json:"new_value" validate:"required"`
	Justification        string                 `json:"justification" validate:"required,min=10"`
	BusinessReason       string                 `json:"business_reason"`
	ApprovalTimeoutHours int                    `json:"approval_timeout_hours"`
	AutoRollback         *AutoRollbackConfig    `json:"auto_rollback,omitempty"`
	Metadata             map[string]interface{} `json:"metadata,omitempty"`
}

// BulkFeatureFlagChangeRequestInput represents bulk change request
type BulkFeatureFlagChangeRequestInput struct {
	Changes              []*IndividualFlagChangeInput `json:"changes" validate:"required,min=1"`
	Justification        string                       `json:"justification" validate:"required,min=10"`
	BusinessReason       string                       `json:"business_reason"`
	ApprovalTimeoutHours int                          `json:"approval_timeout_hours"`
	Metadata             map[string]interface{}       `json:"metadata,omitempty"`
}

// IndividualFlagChangeInput represents a single flag change
type IndividualFlagChangeInput struct {
	FlagName     string                 `json:"flag_name" validate:"required"`
	ChangeType   string                 `json:"change_type" validate:"required"`
	NewValue     interface{}            `json:"new_value" validate:"required"`
	AutoRollback *AutoRollbackConfig    `json:"auto_rollback,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// AutoRollbackConfig specifies automatic rollback settings
type AutoRollbackConfig struct {
	Enabled       bool          `json:"enabled"`
	Duration      time.Duration `json:"duration"`
	RollbackValue interface{}   `json:"rollback_value"`
}

// FeatureFlagWorkflowResult represents the result of a workflow
type FeatureFlagWorkflowResult struct {
	WorkflowID        string                   `json:"workflow_id"`
	Status            string                   `json:"status"`
	FlagName          string                   `json:"flag_name"`
	ChangeType        string                   `json:"change_type"`
	RequiresApproval  bool                     `json:"requires_approval"`
	AccessRequestID   *uuid.UUID               `json:"access_request_id,omitempty"`
	EstimatedDuration time.Duration            `json:"estimated_duration"`
	CreatedAt         time.Time                `json:"created_at"`
	CompletedAt       *time.Time               `json:"completed_at,omitempty"`
	Result            *FeatureFlagChangeResult `json:"result,omitempty"`
}

// BulkFeatureFlagWorkflowResult represents bulk workflow result
type BulkFeatureFlagWorkflowResult struct {
	WorkflowID        string                              `json:"workflow_id"`
	Status            string                              `json:"status"`
	TotalChanges      int                                 `json:"total_changes"`
	CompletedChanges  int                                 `json:"completed_changes"`
	FailedChanges     int                                 `json:"failed_changes"`
	RequiresApproval  bool                                `json:"requires_approval"`
	AccessRequestIDs  []uuid.UUID                         `json:"access_request_ids,omitempty"`
	EstimatedDuration time.Duration                       `json:"estimated_duration"`
	CreatedAt         time.Time                           `json:"created_at"`
	CompletedAt       *time.Time                          `json:"completed_at,omitempty"`
	Results           map[string]*FeatureFlagChangeResult `json:"results,omitempty"`
}

// ApprovalRequestInput represents approval input
type ApprovalRequestInput struct {
	ApproverID uuid.UUID `json:"approver_id" validate:"required"`
	Comments   string    `json:"comments"`
}

// RejectionRequestInput represents rejection input
type RejectionRequestInput struct {
	ApproverID uuid.UUID `json:"approver_id" validate:"required"`
	Comments   string    `json:"comments" validate:"required,min=10"`
	Reason     string    `json:"reason"`
}

// WorkflowStatus represents current workflow status
type WorkflowStatus struct {
	WorkflowID        string                 `json:"workflow_id"`
	Status            string                 `json:"status"` // running, completed, failed, cancelled
	FlagName          string                 `json:"flag_name"`
	ChangeType        string                 `json:"change_type"`
	RequiresApproval  bool                   `json:"requires_approval"`
	AccessRequestID   *uuid.UUID             `json:"access_request_id,omitempty"`
	CurrentStep       string                 `json:"current_step"`
	Progress          float64                `json:"progress"` // 0.0 to 1.0
	StartedAt         time.Time              `json:"started_at"`
	LastUpdatedAt     time.Time              `json:"last_updated_at"`
	EstimatedTimeLeft *time.Duration         `json:"estimated_time_left,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// ListPendingApprovalsRequest represents request for pending approvals
type ListPendingApprovalsRequest struct {
	ApproverID *uuid.UUID `json:"approver_id,omitempty"`
	FlagName   *string    `json:"flag_name,omitempty"`
	Limit      int        `json:"limit"`
	Offset     int        `json:"offset"`
}

// ListPendingApprovalsResponse represents response with pending approvals
type ListPendingApprovalsResponse struct {
	Approvals []*PendingApproval `json:"approvals"`
	Total     int64              `json:"total"`
	Page      int                `json:"page"`
	PageSize  int                `json:"page_size"`
}

// PendingApproval represents a pending approval request
type PendingApproval struct {
	WorkflowID       string      `json:"workflow_id"`
	AccessRequestID  uuid.UUID   `json:"access_request_id"`
	FlagName         string      `json:"flag_name"`
	ChangeType       string      `json:"change_type"`
	NewValue         interface{} `json:"new_value"`
	RequestedBy      uuid.UUID   `json:"requested_by"`
	Justification    string      `json:"justification"`
	BusinessReason   string      `json:"business_reason,omitempty"`
	CreatedAt        time.Time   `json:"created_at"`
	ExpiresAt        *time.Time  `json:"expires_at,omitempty"`
	Priority         string      `json:"priority"` // low, medium, high, critical
	ImpactAssessment string      `json:"impact_assessment,omitempty"`
}

// ScheduleAutoRollbackRequest represents auto-rollback scheduling request
type ScheduleAutoRollbackRequest struct {
	FlagID        uuid.UUID   `json:"flag_id" validate:"required"`
	FlagName      string      `json:"flag_name" validate:"required"`
	RollbackAt    time.Time   `json:"rollback_at" validate:"required"`
	RollbackValue interface{} `json:"rollback_value" validate:"required"`
	Reason        string      `json:"reason"`
}

// AutoRollbackScheduleResult represents auto-rollback schedule result
type AutoRollbackScheduleResult struct {
	ScheduleID string    `json:"schedule_id"`
	WorkflowID string    `json:"workflow_id"`
	FlagID     uuid.UUID `json:"flag_id"`
	FlagName   string    `json:"flag_name"`
	RollbackAt time.Time `json:"rollback_at"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}
