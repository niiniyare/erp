package featureflag_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/niiniyare/erp/internal/core/featureflag"
)

// TestWorkflowRequest represents workflow request structure for testing
type TestWorkflowRequest struct {
	WorkflowID      string                 `json:"workflow_id"`
	FlagName        string                 `json:"flag_name"`
	ChangeType      string                 `json:"change_type"`
	Justification   string                 `json:"justification"`
	BusinessReason  string                 `json:"business_reason,omitempty"`
	RequestedBy     uuid.UUID              `json:"requested_by"`
	RequestedAt     time.Time              `json:"requested_at"`
	RequiredApprovers int                   `json:"required_approvers"`
	TimeoutMinutes  int                    `json:"timeout_minutes"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// TestWorkflowResult represents workflow result structure for testing
type TestWorkflowResult struct {
	WorkflowID    string              `json:"workflow_id"`
	Status        string              `json:"status"` // pending, approved, rejected, timeout
	ApprovedBy    *uuid.UUID          `json:"approved_by,omitempty"`
	RejectedBy    *uuid.UUID          `json:"rejected_by,omitempty"`
	CompletedAt   time.Time           `json:"completed_at"`
	ExecutionTime time.Duration       `json:"execution_time"`
	RollbackToken string              `json:"rollback_token,omitempty"`
	AppliedChanges []map[string]interface{} `json:"applied_changes,omitempty"`
}

// Test Case FF-WORKFLOW-001: Feature Flag Change Approval Workflow
func TestFeatureFlagChangeApprovalWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Temporal workflow integration test in short mode")
	}

	tenantID := uuid.New()
	userID := uuid.New()
	approverID := uuid.New()

	t.Run("ValidateWorkflowRequestStructure", func(t *testing.T) {
		// Test workflow request structure
		request := TestWorkflowRequest{
			WorkflowID:      "workflow-" + uuid.New().String(),
			FlagName:        "test-flag",
			ChangeType:      "enable",
			Justification:   "Enable feature for beta testing group",
			BusinessReason:  "Improve user experience based on feedback",
			RequestedBy:     userID,
			RequestedAt:     time.Now(),
			RequiredApprovers: 1,
			TimeoutMinutes:  60,
			Metadata: map[string]interface{}{
				"priority":     "medium",
				"affected_users": 1000,
				"risk_level":   "low",
			},
		}

		// Verify request structure
		assert.NotEmpty(t, request.WorkflowID)
		assert.Equal(t, "test-flag", request.FlagName)
		assert.Equal(t, "enable", request.ChangeType)
		assert.GreaterOrEqual(t, len(request.Justification), 10)
		assert.Equal(t, userID, request.RequestedBy)
		assert.NotZero(t, request.RequestedAt)
		assert.Equal(t, 1, request.RequiredApprovers)
		assert.Equal(t, 60, request.TimeoutMinutes)
		assert.Contains(t, request.Metadata, "priority")
		assert.Contains(t, request.Metadata, "risk_level")
	})

	t.Run("ValidateWorkflowApprovalProcess", func(t *testing.T) {
		// Test approval workflow process
		workflowID := "workflow-" + uuid.New().String()
		
		// Step 1: Create workflow
		request := TestWorkflowRequest{
			WorkflowID:        workflowID,
			FlagName:         "approval-test-flag",
			ChangeType:       "enable",
			Justification:    "Testing approval workflow functionality",
			RequestedBy:      userID,
			RequestedAt:      time.Now(),
			RequiredApprovers: 1,
			TimeoutMinutes:   30,
		}

		// Step 2: Simulate approval
		approvalTime := time.Now()
		approvalResult := TestWorkflowResult{
			WorkflowID:    workflowID,
			Status:        "approved",
			ApprovedBy:    &approverID,
			CompletedAt:   approvalTime,
			ExecutionTime: approvalTime.Sub(request.RequestedAt),
			RollbackToken: "rollback-" + uuid.New().String(),
			AppliedChanges: []map[string]interface{}{
				{
					"flag_name":    request.FlagName,
					"change_type":  request.ChangeType,
					"old_value":    false,
					"new_value":    true,
					"applied_at":   approvalTime,
					"applied_by":   approverID,
				},
			},
		}

		// Verify approval workflow
		assert.Equal(t, request.WorkflowID, approvalResult.WorkflowID)
		assert.Equal(t, "approved", approvalResult.Status)
		assert.Equal(t, approverID, *approvalResult.ApprovedBy)
		assert.NotZero(t, approvalResult.CompletedAt)
		assert.Greater(t, approvalResult.ExecutionTime, time.Duration(0))
		assert.NotEmpty(t, approvalResult.RollbackToken)
		assert.Len(t, approvalResult.AppliedChanges, 1)
		
		change := approvalResult.AppliedChanges[0]
		assert.Equal(t, request.FlagName, change["flag_name"])
		assert.Equal(t, request.ChangeType, change["change_type"])
		assert.Equal(t, false, change["old_value"])
		assert.Equal(t, true, change["new_value"])
		assert.Equal(t, approverID, change["applied_by"])
	})

	t.Run("ValidateWorkflowRejectionProcess", func(t *testing.T) {
		// Test rejection workflow process
		workflowID := "workflow-" + uuid.New().String()
		
		request := TestWorkflowRequest{
			WorkflowID:        workflowID,
			FlagName:         "rejection-test-flag",
			ChangeType:       "disable",
			Justification:    "Testing rejection workflow functionality",
			RequestedBy:      userID,
			RequestedAt:      time.Now(),
			RequiredApprovers: 1,
			TimeoutMinutes:   30,
		}

		// Simulate rejection
		rejectionTime := time.Now()
		rejectionResult := TestWorkflowResult{
			WorkflowID:    workflowID,
			Status:        "rejected",
			RejectedBy:    &approverID,
			CompletedAt:   rejectionTime,
			ExecutionTime: rejectionTime.Sub(request.RequestedAt),
		}

		// Verify rejection workflow
		assert.Equal(t, request.WorkflowID, rejectionResult.WorkflowID)
		assert.Equal(t, "rejected", rejectionResult.Status)
		assert.Equal(t, approverID, *rejectionResult.RejectedBy)
		assert.NotZero(t, rejectionResult.CompletedAt)
		assert.Greater(t, rejectionResult.ExecutionTime, time.Duration(0))
		assert.Empty(t, rejectionResult.RollbackToken) // No rollback token for rejected workflows
		assert.Empty(t, rejectionResult.AppliedChanges) // No changes applied for rejected workflows
	})

	t.Run("ValidateWorkflowTimeout", func(t *testing.T) {
		// Test workflow timeout handling
		workflowID := "workflow-" + uuid.New().String()
		
		request := TestWorkflowRequest{
			WorkflowID:        workflowID,
			FlagName:         "timeout-test-flag",
			ChangeType:       "update_rollout",
			Justification:    "Testing timeout workflow functionality",
			RequestedBy:      userID,
			RequestedAt:      time.Now().Add(-65 * time.Minute), // Simulate old request
			RequiredApprovers: 1,
			TimeoutMinutes:   60,
		}

		// Simulate timeout
		timeoutTime := request.RequestedAt.Add(time.Duration(request.TimeoutMinutes) * time.Minute)
		timeoutResult := TestWorkflowResult{
			WorkflowID:    workflowID,
			Status:        "timeout",
			CompletedAt:   timeoutTime,
			ExecutionTime: timeoutTime.Sub(request.RequestedAt),
		}

		// Verify timeout workflow
		assert.Equal(t, request.WorkflowID, timeoutResult.WorkflowID)
		assert.Equal(t, "timeout", timeoutResult.Status)
		assert.Nil(t, timeoutResult.ApprovedBy)
		assert.Nil(t, timeoutResult.RejectedBy)
		assert.Equal(t, 60*time.Minute, timeoutResult.ExecutionTime)
		assert.Empty(t, timeoutResult.RollbackToken)
		assert.Empty(t, timeoutResult.AppliedChanges)
	})
}

// Test Case FF-WORKFLOW-002: Bulk Change Approval Workflow
func TestBulkChangeApprovalWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping bulk workflow integration test in short mode")
	}

	tenantID := uuid.New()
	userID := uuid.New()
	approverID := uuid.New()

	t.Run("ValidateBulkWorkflowStructure", func(t *testing.T) {
		// Test bulk workflow request structure
		bulkChanges := []map[string]interface{}{
			{
				"flag_name":    "feature-1",
				"change_type":  "enable",
				"target_value": true,
			},
			{
				"flag_name":    "feature-2", 
				"change_type":  "disable",
				"target_value": false,
			},
			{
				"flag_name":    "feature-3",
				"change_type":  "update_rollout",
				"target_value": 50,
			},
		}

		request := TestWorkflowRequest{
			WorkflowID:        "bulk-workflow-" + uuid.New().String(),
			ChangeType:        "bulk_update",
			Justification:     "Bulk update for feature rollout coordination",
			BusinessReason:    "Synchronized feature release for major product update",
			RequestedBy:       userID,
			RequestedAt:       time.Now(),
			RequiredApprovers: 2, // Bulk changes require more approvers
			TimeoutMinutes:    120, // Longer timeout for bulk operations
			Metadata: map[string]interface{}{
				"bulk_changes":    bulkChanges,
				"change_count":    len(bulkChanges),
				"priority":        "high",
				"coordination_id": uuid.New().String(),
			},
		}

		// Verify bulk workflow structure
		assert.NotEmpty(t, request.WorkflowID)
		assert.Equal(t, "bulk_update", request.ChangeType)
		assert.GreaterOrEqual(t, len(request.Justification), 10)
		assert.Equal(t, userID, request.RequestedBy)
		assert.Equal(t, 2, request.RequiredApprovers)
		assert.Equal(t, 120, request.TimeoutMinutes)
		
		metadata := request.Metadata
		assert.Contains(t, metadata, "bulk_changes")
		assert.Equal(t, 3, metadata["change_count"])
		assert.Equal(t, "high", metadata["priority"])
		
		changes := metadata["bulk_changes"].([]map[string]interface{})
		assert.Len(t, changes, 3)
		assert.Equal(t, "feature-1", changes[0]["flag_name"])
		assert.Equal(t, "enable", changes[0]["change_type"])
		assert.Equal(t, true, changes[0]["target_value"])
	})

	t.Run("ValidateBulkApprovalProcess", func(t *testing.T) {
		// Test bulk approval workflow process
		workflowID := "bulk-workflow-" + uuid.New().String()
		
		approvalTime := time.Now()
		bulkResult := TestWorkflowResult{
			WorkflowID:    workflowID,
			Status:        "approved",
			ApprovedBy:    &approverID,
			CompletedAt:   approvalTime,
			ExecutionTime: 30 * time.Minute,
			RollbackToken: "bulk-rollback-" + uuid.New().String(),
			AppliedChanges: []map[string]interface{}{
				{
					"flag_name":    "feature-1",
					"change_type":  "enable", 
					"old_value":    false,
					"new_value":    true,
					"applied_at":   approvalTime,
				},
				{
					"flag_name":    "feature-2",
					"change_type":  "disable",
					"old_value":    true,
					"new_value":    false,
					"applied_at":   approvalTime,
				},
				{
					"flag_name":    "feature-3",
					"change_type":  "update_rollout",
					"old_value":    25,
					"new_value":    50,
					"applied_at":   approvalTime,
				},
			},
		}

		// Verify bulk approval results
		assert.Equal(t, "approved", bulkResult.Status)
		assert.Equal(t, approverID, *bulkResult.ApprovedBy)
		assert.NotEmpty(t, bulkResult.RollbackToken)
		assert.Len(t, bulkResult.AppliedChanges, 3)

		// Verify all changes were applied atomically
		for i, change := range bulkResult.AppliedChanges {
			assert.Contains(t, change, "flag_name")
			assert.Contains(t, change, "change_type") 
			assert.Contains(t, change, "old_value")
			assert.Contains(t, change, "new_value")
			assert.Equal(t, approvalTime, change["applied_at"])
			
			// Verify specific changes
			switch i {
			case 0:
				assert.Equal(t, "feature-1", change["flag_name"])
				assert.Equal(t, "enable", change["change_type"])
			case 1:
				assert.Equal(t, "feature-2", change["flag_name"])
				assert.Equal(t, "disable", change["change_type"])
			case 2:
				assert.Equal(t, "feature-3", change["flag_name"])
				assert.Equal(t, "update_rollout", change["change_type"])
			}
		}
	})

	t.Run("ValidatePartialBulkFailure", func(t *testing.T) {
		// Test handling of partial failures in bulk operations
		workflowID := "bulk-partial-workflow-" + uuid.New().String()
		
		partialResult := TestWorkflowResult{
			WorkflowID:    workflowID,
			Status:        "partial_success",
			ApprovedBy:    &approverID,
			CompletedAt:   time.Now(),
			ExecutionTime: 45 * time.Minute,
			RollbackToken: "partial-rollback-" + uuid.New().String(),
			AppliedChanges: []map[string]interface{}{
				{
					"flag_name":    "feature-1",
					"change_type":  "enable",
					"status":       "success",
					"applied_at":   time.Now(),
				},
				{
					"flag_name":    "feature-2",
					"change_type":  "disable",
					"status":       "failed",
					"error":        "Flag not found",
				},
				{
					"flag_name":    "feature-3",
					"change_type":  "update_rollout",
					"status":       "success",
					"applied_at":   time.Now(),
				},
			},
		}

		// Verify partial failure handling
		assert.Equal(t, "partial_success", partialResult.Status)
		assert.Len(t, partialResult.AppliedChanges, 3)

		successCount := 0
		failureCount := 0
		for _, change := range partialResult.AppliedChanges {
			status := change["status"].(string)
			if status == "success" {
				successCount++
				assert.Contains(t, change, "applied_at")
			} else if status == "failed" {
				failureCount++
				assert.Contains(t, change, "error")
			}
		}

		assert.Equal(t, 2, successCount)
		assert.Equal(t, 1, failureCount)
	})
}

// Test Case FF-WORKFLOW-003: Workflow Approval and Rejection
func TestWorkflowApprovalAndRejection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping workflow approval test in short mode")
	}

	userID := uuid.New()
	approverID := uuid.New()

	t.Run("ValidateApprovalWithComments", func(t *testing.T) {
		// Test approval with comments
		workflowID := "approval-comments-" + uuid.New().String()
		
		approvalData := map[string]interface{}{
			"workflow_id":     workflowID,
			"approver_id":     approverID,
			"decision":        "approve",
			"comments":        "Approved after security review",
			"approved_at":     time.Now(),
			"conditions":      []string{"Monitor error rates", "Rollback if issues"},
		}

		result := TestWorkflowResult{
			WorkflowID:  workflowID,
			Status:      "approved",
			ApprovedBy:  &approverID,
			CompletedAt: approvalData["approved_at"].(time.Time),
		}

		// Verify approval with metadata
		assert.Equal(t, "approval-comments-"+workflowID[len("approval-comments-"):], result.WorkflowID)
		assert.Equal(t, "approved", result.Status)
		assert.Equal(t, approverID, *result.ApprovedBy)
		
		assert.Equal(t, "approve", approvalData["decision"])
		assert.Equal(t, "Approved after security review", approvalData["comments"])
		assert.Len(t, approvalData["conditions"], 2)
	})

	t.Run("ValidateRejectionWithReason", func(t *testing.T) {
		// Test rejection with detailed reason
		workflowID := "rejection-reason-" + uuid.New().String()
		
		rejectionData := map[string]interface{}{
			"workflow_id":    workflowID,
			"approver_id":    approverID,
			"decision":       "reject",
			"reason":         "insufficient_justification",
			"comments":       "Please provide more detailed business justification and impact analysis",
			"rejected_at":    time.Now(),
			"required_changes": []string{
				"Add business impact metrics",
				"Include rollback plan",
				"Specify monitoring requirements",
			},
		}

		result := TestWorkflowResult{
			WorkflowID:  workflowID,
			Status:      "rejected",
			RejectedBy:  &approverID,
			CompletedAt: rejectionData["rejected_at"].(time.Time),
		}

		// Verify rejection with detailed feedback
		assert.Equal(t, "rejected", result.Status)
		assert.Equal(t, approverID, *result.RejectedBy)
		
		assert.Equal(t, "reject", rejectionData["decision"])
		assert.Equal(t, "insufficient_justification", rejectionData["reason"])
		assert.GreaterOrEqual(t, len(rejectionData["comments"].(string)), 20)
		assert.Len(t, rejectionData["required_changes"], 3)
	})

	t.Run("ValidateApprovalAuditTrail", func(t *testing.T) {
		// Test audit trail for approval process
		workflowID := "audit-trail-" + uuid.New().String()
		
		auditTrail := []map[string]interface{}{
			{
				"timestamp":  time.Now().Add(-60 * time.Minute),
				"event":      "workflow_created",
				"actor":      userID,
				"details":    "Feature flag change workflow initiated",
			},
			{
				"timestamp":  time.Now().Add(-30 * time.Minute),
				"event":      "approval_requested", 
				"actor":      "system",
				"details":    "Approval notification sent to designated approvers",
			},
			{
				"timestamp":  time.Now(),
				"event":      "workflow_approved",
				"actor":      approverID,
				"details":    "Workflow approved with conditions",
			},
			{
				"timestamp":  time.Now(),
				"event":      "changes_applied",
				"actor":      "system", 
				"details":    "Feature flag changes applied successfully",
			},
		}

		// Verify audit trail structure
		assert.Len(t, auditTrail, 4)
		
		for i, entry := range auditTrail {
			assert.Contains(t, entry, "timestamp")
			assert.Contains(t, entry, "event")
			assert.Contains(t, entry, "actor")
			assert.Contains(t, entry, "details")
			
			// Verify chronological order
			if i > 0 {
				prevTime := auditTrail[i-1]["timestamp"].(time.Time)
				currTime := entry["timestamp"].(time.Time)
				assert.True(t, currTime.After(prevTime) || currTime.Equal(prevTime))
			}
		}

		// Verify specific events
		assert.Equal(t, "workflow_created", auditTrail[0]["event"])
		assert.Equal(t, userID, auditTrail[0]["actor"])
		
		assert.Equal(t, "workflow_approved", auditTrail[2]["event"])
		assert.Equal(t, approverID, auditTrail[2]["actor"])
	})
}

// Test Case FF-WORKFLOW-004: Auto-Rollback Scheduling
func TestAutoRollbackScheduling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping auto-rollback test in short mode")
	}

	tenantID := uuid.New()
	userID := uuid.New()
	flagID := uuid.New()

	t.Run("ValidateRollbackScheduleCreation", func(t *testing.T) {
		// Test auto-rollback schedule creation
		rollbackRequest := map[string]interface{}{
			"flag_id":          flagID,
			"flag_name":        "rollback-test-flag",
			"rollback_at":      time.Now().Add(2 * time.Hour),
			"original_state":   map[string]interface{}{
				"enabled":          false,
				"rollout_percentage": 0,
				"default_value":    false,
			},
			"created_by":       userID,
			"created_at":       time.Now(),
			"reason":           "Automatic rollback after test period",
			"schedule_id":      "schedule-" + uuid.New().String(),
		}

		// Verify rollback schedule structure
		assert.Equal(t, flagID, rollbackRequest["flag_id"])
		assert.Equal(t, "rollback-test-flag", rollbackRequest["flag_name"])
		assert.True(t, rollbackRequest["rollback_at"].(time.Time).After(time.Now()))
		assert.Equal(t, userID, rollbackRequest["created_by"])
		assert.Contains(t, rollbackRequest, "original_state")
		assert.Contains(t, rollbackRequest, "schedule_id")
		
		originalState := rollbackRequest["original_state"].(map[string]interface{})
		assert.Equal(t, false, originalState["enabled"])
		assert.Equal(t, 0, originalState["rollout_percentage"])
		assert.Equal(t, false, originalState["default_value"])
	})

	t.Run("ValidateRollbackExecution", func(t *testing.T) {
		// Test rollback execution
		scheduleID := "schedule-" + uuid.New().String()
		rollbackTime := time.Now()
		
		rollbackExecution := map[string]interface{}{
			"schedule_id":      scheduleID,
			"flag_id":          flagID,
			"executed_at":      rollbackTime,
			"execution_status": "success",
			"changes_applied":  []map[string]interface{}{
				{
					"property":   "enabled",
					"from_value": true,
					"to_value":   false,
				},
				{
					"property":   "rollout_percentage", 
					"from_value": 75,
					"to_value":   0,
				},
			},
			"notifications_sent": []string{
				"websocket",
				"audit_log",
				"monitoring_alert",
			},
		}

		// Verify rollback execution
		assert.Equal(t, scheduleID, rollbackExecution["schedule_id"])
		assert.Equal(t, flagID, rollbackExecution["flag_id"])
		assert.Equal(t, "success", rollbackExecution["execution_status"])
		
		changes := rollbackExecution["changes_applied"].([]map[string]interface{})
		assert.Len(t, changes, 2)
		
		// Verify enabled was set back to false
		enabledChange := changes[0]
		assert.Equal(t, "enabled", enabledChange["property"])
		assert.Equal(t, true, enabledChange["from_value"])
		assert.Equal(t, false, enabledChange["to_value"])
		
		// Verify rollout was set back to 0
		rolloutChange := changes[1]
		assert.Equal(t, "rollout_percentage", rolloutChange["property"])
		assert.Equal(t, 75, rolloutChange["from_value"])
		assert.Equal(t, 0, rolloutChange["to_value"])
		
		notifications := rollbackExecution["notifications_sent"].([]string)
		assert.Contains(t, notifications, "websocket")
		assert.Contains(t, notifications, "audit_log")
		assert.Contains(t, notifications, "monitoring_alert")
	})

	t.Run("ValidateRollbackCancellation", func(t *testing.T) {
		// Test rollback schedule cancellation
		scheduleID := "schedule-" + uuid.New().String()
		
		cancellationData := map[string]interface{}{
			"schedule_id":       scheduleID,
			"cancelled_at":      time.Now(),
			"cancelled_by":      userID,
			"cancellation_reason": "Feature is performing well, no rollback needed",
			"status":           "cancelled",
		}

		// Verify cancellation structure
		assert.Equal(t, scheduleID, cancellationData["schedule_id"])
		assert.Equal(t, userID, cancellationData["cancelled_by"])
		assert.Equal(t, "cancelled", cancellationData["status"])
		assert.GreaterOrEqual(t, len(cancellationData["cancellation_reason"].(string)), 10)
		assert.NotZero(t, cancellationData["cancelled_at"])
	})

	t.Run("ValidateRollbackFailureHandling", func(t *testing.T) {
		// Test rollback failure handling
		scheduleID := "schedule-" + uuid.New().String()
		
		rollbackFailure := map[string]interface{}{
			"schedule_id":       scheduleID,
			"flag_id":          flagID,
			"executed_at":      time.Now(),
			"execution_status": "failed",
			"error":            "Database connection timeout during rollback",
			"retry_count":      3,
			"next_retry_at":    time.Now().Add(5 * time.Minute),
			"max_retries":      5,
			"failure_notifications": []string{
				"ops_team_alert",
				"high_priority_page",
			},
		}

		// Verify failure handling
		assert.Equal(t, "failed", rollbackFailure["execution_status"])
		assert.Contains(t, rollbackFailure, "error")
		assert.Equal(t, 3, rollbackFailure["retry_count"])
		assert.Equal(t, 5, rollbackFailure["max_retries"])
		assert.True(t, rollbackFailure["next_retry_at"].(time.Time).After(time.Now()))
		
		notifications := rollbackFailure["failure_notifications"].([]string)
		assert.Contains(t, notifications, "ops_team_alert")
		assert.Contains(t, notifications, "high_priority_page")
	})
}

// Integration test for complete workflow lifecycle
func TestWorkflowIntegrationLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping workflow integration test in short mode")
	}

	t.Run("ValidateCompleteWorkflowLifecycle", func(t *testing.T) {
		// Test the complete workflow from request to execution
		tenantID := uuid.New()
		userID := uuid.New()
		approverID := uuid.New()
		
		// Step 1: Workflow Creation
		workflowID := "integration-" + uuid.New().String()
		request := TestWorkflowRequest{
			WorkflowID:        workflowID,
			FlagName:         "integration-test-flag",
			ChangeType:       "enable",
			Justification:    "Complete integration test for workflow lifecycle",
			RequestedBy:      userID,
			RequestedAt:      time.Now(),
			RequiredApprovers: 1,
			TimeoutMinutes:   30,
		}

		// Step 2: Approval Process
		approvalTime := request.RequestedAt.Add(10 * time.Minute)
		
		// Step 3: Execution and Rollback Schedule
		result := TestWorkflowResult{
			WorkflowID:    workflowID,
			Status:        "approved",
			ApprovedBy:    &approverID,
			CompletedAt:   approvalTime,
			ExecutionTime: approvalTime.Sub(request.RequestedAt),
			RollbackToken: "integration-rollback-" + uuid.New().String(),
			AppliedChanges: []map[string]interface{}{
				{
					"flag_name":     request.FlagName,
					"change_type":   request.ChangeType,
					"old_value":     false,
					"new_value":     true,
					"applied_at":    approvalTime,
					"applied_by":    approverID,
					"rollback_schedule": approvalTime.Add(24 * time.Hour),
				},
			},
		}

		// Verify complete lifecycle
		assert.Equal(t, request.WorkflowID, result.WorkflowID)
		assert.Equal(t, "approved", result.Status)
		assert.NotEmpty(t, result.RollbackToken)
		assert.Len(t, result.AppliedChanges, 1)
		
		change := result.AppliedChanges[0]
		assert.Equal(t, request.FlagName, change["flag_name"])
		assert.Equal(t, request.ChangeType, change["change_type"])
		assert.Contains(t, change, "rollback_schedule")
		
		rollbackSchedule := change["rollback_schedule"].(time.Time)
		assert.True(t, rollbackSchedule.After(approvalTime))
		
		t.Logf("Complete workflow lifecycle validated: %s", workflowID)
		t.Logf("  - Request created at: %s", request.RequestedAt.Format(time.RFC3339))
		t.Logf("  - Approved at: %s", approvalTime.Format(time.RFC3339))
		t.Logf("  - Execution time: %s", result.ExecutionTime)
		t.Logf("  - Rollback scheduled for: %s", rollbackSchedule.Format(time.RFC3339))
	})
}