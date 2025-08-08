package featureflag

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/core/access/request"
	"github.com/niiniyare/erp/internal/core/tenant"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/workflows/featureflag"
	db "github.com/niiniyare/erp/db/sqlc"
	"go.temporal.io/sdk/client"
)

// WorkflowService handles feature flag workflows with approval processes
type WorkflowService interface {
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

// workflowService implements WorkflowService
type workflowService struct {
	temporalClient      client.Client
	featureFlagService  SimpleService
	accessRequestRepo   request.AccessRequestRepository
	tenantService       tenant.Service
	store               db.Store
	webSocketService    WebSocketService
}

// NewWorkflowService creates a new workflow service
func NewWorkflowService(
	temporalClient client.Client,
	featureFlagService SimpleService,
	accessRequestRepo request.AccessRequestRepository,
	tenantService tenant.Service,
	store db.Store,
	webSocketService WebSocketService,
) WorkflowService {
	return &workflowService{
		temporalClient:     temporalClient,
		featureFlagService: featureFlagService,
		accessRequestRepo:  accessRequestRepo,
		tenantService:      tenantService,
		store:              store,
		webSocketService:   webSocketService,
	}
}

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
	WorkflowID        string                         `json:"workflow_id"`
	Status            string                         `json:"status"`
	FlagName          string                         `json:"flag_name"`
	ChangeType        string                         `json:"change_type"`
	RequiresApproval  bool                           `json:"requires_approval"`
	AccessRequestID   *uuid.UUID                     `json:"access_request_id,omitempty"`
	EstimatedDuration time.Duration                  `json:"estimated_duration"`
	CreatedAt         time.Time                      `json:"created_at"`
	CompletedAt       *time.Time                     `json:"completed_at,omitempty"`
	Result            *featureflag.FeatureFlagChangeResult `json:"result,omitempty"`
}

// BulkFeatureFlagWorkflowResult represents bulk workflow result
type BulkFeatureFlagWorkflowResult struct {
	WorkflowID        string                            `json:"workflow_id"`
	Status            string                            `json:"status"`
	TotalChanges      int                               `json:"total_changes"`
	CompletedChanges  int                               `json:"completed_changes"`
	FailedChanges     int                               `json:"failed_changes"`
	RequiresApproval  bool                              `json:"requires_approval"`
	AccessRequestIDs  []uuid.UUID                       `json:"access_request_ids,omitempty"`
	EstimatedDuration time.Duration                     `json:"estimated_duration"`
	CreatedAt         time.Time                         `json:"created_at"`
	CompletedAt       *time.Time                        `json:"completed_at,omitempty"`
	Results           map[string]*featureflag.FeatureFlagChangeResult `json:"results,omitempty"`
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
	WorkflowID        string     `json:"workflow_id"`
	AccessRequestID   uuid.UUID  `json:"access_request_id"`
	FlagName          string     `json:"flag_name"`
	ChangeType        string     `json:"change_type"`
	NewValue          interface{} `json:"new_value"`
	RequestedBy       uuid.UUID  `json:"requested_by"`
	Justification     string     `json:"justification"`
	BusinessReason    string     `json:"business_reason,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty"`
	Priority          string     `json:"priority"` // low, medium, high, critical
	ImpactAssessment  string     `json:"impact_assessment,omitempty"`
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

// RequestFeatureFlagChange initiates a feature flag change workflow
func (s *workflowService) RequestFeatureFlagChange(ctx context.Context, req *FeatureFlagChangeRequestInput) (*FeatureFlagWorkflowResult, error) {
	logger := logger.WithFields(logger.Fields{
		"service": "workflowService",
		"method":  "RequestFeatureFlagChange",
		"flag_name": req.FlagName,
		"change_type": req.ChangeType,
	})

	// Validate tenant context
	tenant, err := s.tenantService.GetCurrentTenant(ctx)
	if err != nil {
		logger.Error("Failed to get current tenant", logger.Fields{"error": err})
		return nil, fmt.Errorf("invalid tenant context: %w", err)
	}

	// Get current user (would typically come from JWT/auth context)
	userID := uuid.New() // This should be extracted from auth context

	// Create Temporal workflow request
	workflowReq := &featureflag.FeatureFlagChangeRequest{
		TenantID:             tenant.ID,
		RequestedBy:          userID,
		FlagName:             req.FlagName,
		ChangeType:           req.ChangeType,
		NewValue:             req.NewValue,
		Justification:        req.Justification,
		BusinessReason:       req.BusinessReason,
		ApprovalTimeoutHours: req.ApprovalTimeoutHours,
		Metadata:             req.Metadata,
	}

	// Set defaults
	if workflowReq.ApprovalTimeoutHours == 0 {
		workflowReq.ApprovalTimeoutHours = 24 // Default 24 hours
	}

	// Generate workflow ID
	workflowID := fmt.Sprintf("feature-flag-change-%s-%d", req.FlagName, time.Now().Unix())

	// Start the workflow
	workflowOptions := client.StartWorkflowOptions{
		ID:           workflowID,
		TaskQueue:    "feature-flag-workflows",
		WorkflowExecutionTimeout: 48 * time.Hour, // Maximum 48 hours
	}

	execution, err := s.temporalClient.ExecuteWorkflow(ctx, workflowOptions, featureflag.FeatureFlagWorkflow, workflowReq)
	if err != nil {
		logger.Error("Failed to start workflow", logger.Fields{"error": err})
		return nil, fmt.Errorf("failed to start workflow: %w", err)
	}

	// Create result
	result := &FeatureFlagWorkflowResult{
		WorkflowID:        execution.GetID(),
		Status:            "running",
		FlagName:          req.FlagName,
		ChangeType:        req.ChangeType,
		RequiresApproval:  true, // Will be determined by the workflow
		EstimatedDuration: time.Duration(workflowReq.ApprovalTimeoutHours) * time.Hour,
		CreatedAt:         time.Now(),
	}

	// Schedule auto-rollback if configured
	if req.AutoRollback != nil && req.AutoRollback.Enabled {
		go s.scheduleAutoRollbackForWorkflow(context.Background(), workflowID, req.FlagName, req.AutoRollback)
	}

	logger.Info("Feature flag change workflow started", logger.Fields{
		"workflow_id": workflowID,
		"flag_name":   req.FlagName,
		"change_type": req.ChangeType,
	})

	return result, nil
}

// RequestBulkFeatureFlagChange initiates a bulk feature flag change workflow
func (s *workflowService) RequestBulkFeatureFlagChange(ctx context.Context, req *BulkFeatureFlagChangeRequestInput) (*BulkFeatureFlagWorkflowResult, error) {
	logger := logger.WithFields(logger.Fields{
		"service": "workflowService",
		"method":  "RequestBulkFeatureFlagChange",
		"change_count": len(req.Changes),
	})

	// Validate tenant context
	tenant, err := s.tenantService.GetCurrentTenant(ctx)
	if err != nil {
		logger.Error("Failed to get current tenant", logger.Fields{"error": err})
		return nil, fmt.Errorf("invalid tenant context: %w", err)
	}

	// Get current user
	userID := uuid.New() // This should be extracted from auth context

	// Convert individual changes
	changes := make([]*featureflag.IndividualFlagChange, len(req.Changes))
	for i, change := range req.Changes {
		changes[i] = &featureflag.IndividualFlagChange{
			FlagName:   change.FlagName,
			ChangeType: change.ChangeType,
			NewValue:   change.NewValue,
			Metadata:   change.Metadata,
		}
	}

	// Create Temporal workflow request
	workflowReq := &featureflag.BulkFeatureFlagChangeRequest{
		TenantID:             tenant.ID,
		RequestedBy:          userID,
		Changes:              changes,
		Justification:        req.Justification,
		BusinessReason:       req.BusinessReason,
		ApprovalTimeoutHours: req.ApprovalTimeoutHours,
	}

	// Set defaults
	if workflowReq.ApprovalTimeoutHours == 0 {
		workflowReq.ApprovalTimeoutHours = 24 // Default 24 hours
	}

	// Generate workflow ID
	workflowID := fmt.Sprintf("bulk-feature-flag-change-%d-%d", len(req.Changes), time.Now().Unix())

	// Start the workflow
	workflowOptions := client.StartWorkflowOptions{
		ID:           workflowID,
		TaskQueue:    "feature-flag-workflows",
		WorkflowExecutionTimeout: 72 * time.Hour, // Maximum 72 hours for bulk operations
	}

	execution, err := s.temporalClient.ExecuteWorkflow(ctx, workflowOptions, featureflag.BulkFeatureFlagWorkflow, workflowReq)
	if err != nil {
		logger.Error("Failed to start bulk workflow", logger.Fields{"error": err})
		return nil, fmt.Errorf("failed to start bulk workflow: %w", err)
	}

	// Create result
	result := &BulkFeatureFlagWorkflowResult{
		WorkflowID:        execution.GetID(),
		Status:            "running",
		TotalChanges:      len(req.Changes),
		RequiresApproval:  true, // Will be determined by the workflow
		EstimatedDuration: time.Duration(workflowReq.ApprovalTimeoutHours) * time.Hour,
		CreatedAt:         time.Now(),
	}

	logger.Info("Bulk feature flag change workflow started", logger.Fields{
		"workflow_id":  workflowID,
		"change_count": len(req.Changes),
	})

	return result, nil
}

// ApproveFeatureFlagChange approves a pending feature flag change
func (s *workflowService) ApproveFeatureFlagChange(ctx context.Context, workflowID string, approvalReq *ApprovalRequestInput) error {
	logger := logger.WithFields(logger.Fields{
		"service":     "workflowService",
		"method":      "ApproveFeatureFlagChange",
		"workflow_id": workflowID,
		"approver_id": approvalReq.ApproverID,
	})

	// Create approval signal
	signal := featureflag.ApprovalSignal{
		Approved:   true,
		ApproverID: approvalReq.ApproverID,
		Comments:   approvalReq.Comments,
		ApprovedAt: time.Now(),
	}

	// Send signal to workflow
	err := s.temporalClient.SignalWorkflow(ctx, workflowID, "", "approval-signal", signal)
	if err != nil {
		logger.Error("Failed to send approval signal", logger.Fields{"error": err})
		return fmt.Errorf("failed to send approval signal: %w", err)
	}

	logger.Info("Feature flag change approved", logger.Fields{
		"workflow_id": workflowID,
		"approver_id": approvalReq.ApproverID,
	})

	return nil
}

// RejectFeatureFlagChange rejects a pending feature flag change
func (s *workflowService) RejectFeatureFlagChange(ctx context.Context, workflowID string, rejectionReq *RejectionRequestInput) error {
	logger := logger.WithFields(logger.Fields{
		"service":     "workflowService",
		"method":      "RejectFeatureFlagChange",
		"workflow_id": workflowID,
		"approver_id": rejectionReq.ApproverID,
	})

	// Create rejection signal
	signal := featureflag.ApprovalSignal{
		Approved:   false,
		ApproverID: rejectionReq.ApproverID,
		Comments:   rejectionReq.Comments,
		ApprovedAt: time.Now(),
	}

	// Send signal to workflow
	err := s.temporalClient.SignalWorkflow(ctx, workflowID, "", "approval-signal", signal)
	if err != nil {
		logger.Error("Failed to send rejection signal", logger.Fields{"error": err})
		return fmt.Errorf("failed to send rejection signal: %w", err)
	}

	logger.Info("Feature flag change rejected", logger.Fields{
		"workflow_id": workflowID,
		"approver_id": rejectionReq.ApproverID,
		"reason":      rejectionReq.Reason,
	})

	return nil
}

// GetWorkflowStatus retrieves the current status of a workflow
func (s *workflowService) GetWorkflowStatus(ctx context.Context, workflowID string) (*WorkflowStatus, error) {
	logger := logger.WithFields(logger.Fields{
		"service":     "workflowService",
		"method":      "GetWorkflowStatus",
		"workflow_id": workflowID,
	})

	// Get workflow execution
	execution := s.temporalClient.GetWorkflow(ctx, workflowID, "")

	// Get workflow description
	desc, err := execution.Describe(ctx)
	if err != nil {
		logger.Error("Failed to describe workflow", logger.Fields{"error": err})
		return nil, fmt.Errorf("failed to get workflow status: %w", err)
	}

	// Map Temporal status to our status
	var status string
	switch desc.WorkflowExecutionInfo.Status {
	case 1: // Running
		status = "running"
	case 2: // Completed
		status = "completed"
	case 3: // Failed
		status = "failed"
	case 4: // Canceled
		status = "cancelled"
	case 5: // Terminated
		status = "terminated"
	default:
		status = "unknown"
	}

	// Calculate progress (simplified)
	progress := 0.0
	if status == "completed" {
		progress = 1.0
	} else if status == "running" {
		// Could be more sophisticated based on workflow state
		progress = 0.5
	}

	workflowStatus := &WorkflowStatus{
		WorkflowID:      workflowID,
		Status:          status,
		CurrentStep:     "processing", // Could be extracted from workflow state
		Progress:        progress,
		StartedAt:       desc.WorkflowExecutionInfo.StartTime,
		LastUpdatedAt:   time.Now(),
	}

	if desc.WorkflowExecutionInfo.CloseTime != nil {
		workflowStatus.EstimatedTimeLeft = nil
	} else {
		// Estimate remaining time (simplified)
		remaining := 30 * time.Minute
		workflowStatus.EstimatedTimeLeft = &remaining
	}

	return workflowStatus, nil
}

// ListPendingApprovals lists pending approval requests
func (s *workflowService) ListPendingApprovals(ctx context.Context, req *ListPendingApprovalsRequest) (*ListPendingApprovalsResponse, error) {
	// This would typically query the access request repository for pending requests
	// and correlate with workflow data
	
	// Simplified implementation for now
	return &ListPendingApprovalsResponse{
		Approvals: []*PendingApproval{},
		Total:     0,
		Page:      1,
		PageSize:  req.Limit,
	}, nil
}

// CancelWorkflow cancels a running workflow
func (s *workflowService) CancelWorkflow(ctx context.Context, workflowID string, reason string) error {
	logger := logger.WithFields(logger.Fields{
		"service":     "workflowService",
		"method":      "CancelWorkflow",
		"workflow_id": workflowID,
		"reason":      reason,
	})

	err := s.temporalClient.CancelWorkflow(ctx, workflowID, "")
	if err != nil {
		logger.Error("Failed to cancel workflow", logger.Fields{"error": err})
		return fmt.Errorf("failed to cancel workflow: %w", err)
	}

	logger.Info("Workflow cancelled", logger.Fields{
		"workflow_id": workflowID,
		"reason":      reason,
	})

	return nil
}

// ScheduleAutoRollback schedules automatic rollback for a feature flag
func (s *workflowService) ScheduleAutoRollback(ctx context.Context, req *ScheduleAutoRollbackRequest) (*AutoRollbackScheduleResult, error) {
	logger := logger.WithFields(logger.Fields{
		"service":     "workflowService",
		"method":      "ScheduleAutoRollback",
		"flag_id":     req.FlagID,
		"rollback_at": req.RollbackAt,
	})

	// Validate tenant context
	tenant, err := s.tenantService.GetCurrentTenant(ctx)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant context: %w", err)
	}

	// Create auto-rollback workflow request
	rollbackReq := &featureflag.AutoRollbackRequest{
		TenantID:      tenant.ID,
		FlagID:        req.FlagID,
		FlagName:      req.FlagName,
		RollbackAt:    req.RollbackAt,
		RollbackValue: req.RollbackValue,
	}

	// Generate workflow ID
	scheduleID := fmt.Sprintf("auto-rollback-%s-%d", req.FlagName, req.RollbackAt.Unix())
	workflowID := fmt.Sprintf("workflow-%s", scheduleID)

	// Start the auto-rollback workflow
	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: "feature-flag-workflows",
	}

	execution, err := s.temporalClient.ExecuteWorkflow(ctx, workflowOptions, featureflag.AutoRollbackWorkflow, rollbackReq)
	if err != nil {
		logger.Error("Failed to schedule auto-rollback", logger.Fields{"error": err})
		return nil, fmt.Errorf("failed to schedule auto-rollback: %w", err)
	}

	result := &AutoRollbackScheduleResult{
		ScheduleID: scheduleID,
		WorkflowID: execution.GetID(),
		FlagID:     req.FlagID,
		FlagName:   req.FlagName,
		RollbackAt: req.RollbackAt,
		Status:     "scheduled",
		CreatedAt:  time.Now(),
	}

	logger.Info("Auto-rollback scheduled", logger.Fields{
		"schedule_id": scheduleID,
		"workflow_id": workflowID,
		"flag_name":   req.FlagName,
		"rollback_at": req.RollbackAt,
	})

	return result, nil
}

// CancelAutoRollback cancels a scheduled auto-rollback
func (s *workflowService) CancelAutoRollback(ctx context.Context, scheduleID string) error {
	logger := logger.WithFields(logger.Fields{
		"service":     "workflowService",
		"method":      "CancelAutoRollback",
		"schedule_id": scheduleID,
	})

	workflowID := fmt.Sprintf("workflow-%s", scheduleID)
	
	err := s.temporalClient.CancelWorkflow(ctx, workflowID, "")
	if err != nil {
		logger.Error("Failed to cancel auto-rollback", logger.Fields{"error": err})
		return fmt.Errorf("failed to cancel auto-rollback: %w", err)
	}

	logger.Info("Auto-rollback cancelled", logger.Fields{
		"schedule_id": scheduleID,
		"workflow_id": workflowID,
	})

	return nil
}

// scheduleAutoRollbackForWorkflow schedules auto-rollback after a workflow completes
func (s *workflowService) scheduleAutoRollbackForWorkflow(ctx context.Context, workflowID, flagName string, config *AutoRollbackConfig) {
	logger := logger.WithFields(logger.Fields{
		"workflow_id": workflowID,
		"flag_name":   flagName,
		"duration":    config.Duration,
	})

	// Wait for the main workflow to complete
	execution := s.temporalClient.GetWorkflow(ctx, workflowID, "")
	
	var result featureflag.FeatureFlagChangeResult
	err := execution.Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to get workflow result for auto-rollback", logger.Fields{"error": err})
		return
	}

	// Only schedule rollback if the workflow was successful
	if result.Status != "completed" || result.FlagID == nil {
		logger.Info("Skipping auto-rollback - workflow not completed successfully")
		return
	}

	// Schedule the rollback
	rollbackAt := time.Now().Add(config.Duration)
	
	rollbackReq := &ScheduleAutoRollbackRequest{
		FlagID:        *result.FlagID,
		FlagName:      flagName,
		RollbackAt:    rollbackAt,
		RollbackValue: config.RollbackValue,
		Reason:        fmt.Sprintf("Auto-rollback after %v from workflow %s", config.Duration, workflowID),
	}

	_, err = s.ScheduleAutoRollback(ctx, rollbackReq)
	if err != nil {
		logger.Error("Failed to schedule auto-rollback", logger.Fields{"error": err})
	} else {
		logger.Info("Auto-rollback scheduled successfully", logger.Fields{
			"rollback_at": rollbackAt,
			"flag_name":   flagName,
		})
	}
}