package featureflag

//
// import (
// 	"time"
//
// 	"github.com/google/uuid"
// 	wf "github.com/niiniyare/erp/internal/featureflag/workflow"
// 	"go.temporal.io/sdk/workflow"
// )
//
// // FeatureFlagWorkflow manages the workflow for feature flag changes requiring approval
// func FeatureFlagWorkflow(ctx workflow.Context, req *workflow.FeatureFlagChangeRequest) (*workflow.FeatureFlagChangeResult, error) {
// 	logger := workflow.GetLogger(ctx)
// 	logger.Info("Starting feature flag workflow", "flag_name", req.FlagName, "change_type", req.ChangeType)
//
// 	var result wf.FeatureFlagChangeResult
//
// 	// Set workflow options
// 	ao := workflow.ActivityOptions{
// 		StartToCloseTimeout: 5 * time.Minute,
// 		RetryPolicy: &RetryPolicy{
// 			MaximumAttempts: 3,
// 		},
// 	}
// 	ctx = workflow.WithActivityOptions(ctx, ao)
//
// 	// Step 1: Validate flag change request
// 	var validationResult workflow.ValidationResult
// 	err := workflow.ExecuteActivity(ctx, ValidateFeatureFlagChangeActivity, req).Get(ctx, &validationResult)
// 	if err != nil {
// 		logger.Error("Flag validation failed", "error", err)
// 		result.Status = "failed"
// 		result.Error = err.Error()
// 		return &result, err
// 	}
//
// 	if !validationResult.IsValid {
// 		result.Status = "validation_failed"
// 		result.Error = validationResult.ErrorMessage
// 		return &result, nil
// 	}
//
// 	// Step 2: Check if approval is required
// 	var requiresApproval bool
// 	err = workflow.ExecuteActivity(ctx, CheckApprovalRequiredActivity, req).Get(ctx, &requiresApproval)
// 	if err != nil {
// 		logger.Error("Failed to check approval requirements", "error", err)
// 		return nil, err
// 	}
//
// 	var accessRequestID *uuid.UUID
// 	if requiresApproval {
// 		// Step 3: Create access request for flag change
// 		var createAccessRequestResult workflow.CreateAccessRequestResult
// 		err = workflow.ExecuteActivity(ctx, CreateAccessRequestActivity, req).Get(ctx, &createAccessRequestResult)
// 		if err != nil {
// 			logger.Error("Failed to create access request", "error", err)
// 			result.Status = "access_request_failed"
// 			result.Error = err.Error()
// 			return &result, err
// 		}
//
// 		accessRequestID = &createAccessRequestResult.RequestID
// 		logger.Info("Access request created", "request_id", createAccessRequestResult.RequestID)
//
// 		// Step 4: Wait for approval with timeout
// 		approvalTimeout := req.ApprovalTimeoutHours
// 		if approvalTimeout == 0 {
// 			approvalTimeout = 24 // Default 24 hours
// 		}
//
// 		selector := workflow.NewSelector(ctx)
// 		// Wait for approval signal
// 		var approvalResult workflow.ApprovalResult
// 		approvalFuture := workflow.NewFuture(ctx)
// 		selector.AddFuture(approvalFuture, func(f workflow.Future) {
// 			// This will be completed by signal
// 		})
//
// 		// Set up approval signal handler
// 		var signalChan workflow.ReceiveChannel = workflow.GetSignalChannel(ctx, "approval-signal")
// 		selector.AddReceive(signalChan, func(c workflow.ReceiveChannel, more bool) {
// 			var signal workflow.ApprovalSignal
// 			c.Receive(ctx, &signal)
// 			approvalResult = workflow.ApprovalResult{
// 				Approved:   signal.Approved,
// 				ApproverID: signal.ApproverID,
// 				Comments:   signal.Comments,
// 				ApprovedAt: signal.ApprovedAt,
// 			}
// 			approvalFuture.Set(approvalResult, nil)
// 		})
//
// 		// Add timeout
// 		timeoutCtx, cancel := workflow.WithTimeout(ctx, time.Duration(approvalTimeout)*time.Hour)
// 		defer cancel()
//
// 		selector.Select(timeoutCtx)
//
// 		if timeoutCtx.Err() != nil {
// 			// Timeout occurred - expire the access request
// 			logger.Info("Approval timeout occurred", "request_id", accessRequestID)
// 			workflow.ExecuteActivity(ctx, ExpireAccessRequestActivity, *accessRequestID)
// 			result.Status = "approval_timeout"
// 			result.Error = "Approval timeout exceeded"
// 			return &result, nil
// 		}
//
// 		if !approvalResult.Approved {
// 			logger.Info("Flag change rejected", "request_id", accessRequestID, "approver_id", approvalResult.ApproverID)
// 			result.Status = "rejected"
// 			result.Error = "Change rejected by approver"
// 			result.ApprovalDetails = &approvalResult
// 			return &result, nil
// 		}
//
// 		logger.Info("Flag change approved", "request_id", accessRequestID, "approver_id", approvalResult.ApproverID)
// 		result.ApprovalDetails = &approvalResult
// 	}
//
// 	// Step 5: Apply the feature flag change
// 	var applyResult workflow.ApplyChangeResult
// 	err = workflow.ExecuteActivity(ctx, ApplyFeatureFlagChangeActivity, req).Get(ctx, &applyResult)
// 	if err != nil {
// 		logger.Error("Failed to apply flag change", "error", err)
// 		result.Status = "apply_failed"
// 		result.Error = err.Error()
// 		return &result, err
// 	}
//
// 	// Step 6: Send real-time notifications via WebSocket
// 	notificationReq := workflow.NotificationRequest{
// 		TenantID:        req.TenantID,
// 		FlagName:        req.FlagName,
// 		ChangeType:      req.ChangeType,
// 		NewValue:        req.NewValue,
// 		AppliedBy:       req.RequestedBy,
// 		AccessRequestID: accessRequestID,
// 		AppliedAt:       applyResult.AppliedAt,
// 	}
// 	workflow.ExecuteActivity(ctx, SendWebSocketNotificationActivity, notificationReq)
//
// 	// Step 7: Update access request to completed if it exists
// 	if accessRequestID != nil {
// 		workflow.ExecuteActivity(ctx, CompleteAccessRequestActivity, *accessRequestID)
// 	}
//
// 	// Step 8: Log audit event
// 	auditReq := workflow.AuditEventRequest{
// 		TenantID:        req.TenantID,
// 		FlagName:        req.FlagName,
// 		ChangeType:      req.ChangeType,
// 		OldValue:        applyResult.OldValue,
// 		NewValue:        req.NewValue,
// 		RequestedBy:     req.RequestedBy,
// 		AccessRequestID: accessRequestID,
// 		AppliedAt:       applyResult.AppliedAt,
// 		Metadata:        req.Metadata,
// 	}
// 	workflow.ExecuteActivity(ctx, CreateAuditEventActivity, auditReq)
//
// 	result.Status = "completed"
// 	result.FlagID = applyResult.FlagID
// 	result.OldValue = applyResult.OldValue
// 	result.NewValue = req.NewValue
// 	result.AppliedAt = applyResult.AppliedAt
//
// 	logger.Info("Feature flag workflow completed successfully", "flag_name", req.FlagName, "flag_id", result.FlagID)
// 	return &result, nil
// }
//
// // BulkFeatureFlagWorkflow handles bulk changes to multiple flags with approval
// func BulkFeatureFlagWorkflow(ctx workflow.Context, req *workflow.BulkFeatureFlagChangeRequest) (*workflow.BulkFeatureFlagChangeResult, error) {
// 	logger := workflow.GetLogger(ctx)
// 	logger.Info("Starting bulk feature flag workflow", "flag_count", len(req.Changes))
//
// 	var result workflow.BulkFeatureFlagChangeResult
// 	result.Results = make(map[string]*workflow.FeatureFlagChangeResult)
//
// 	// Set workflow options for child workflows
// 	cwo := workflow.ChildWorkflowOptions{
// 		WorkflowExecutionTimeout: 2 * time.Hour,
// 	}
// 	ctx = workflow.WithChildOptions(ctx, cwo)
//
// 	// Execute individual flag changes as child workflows
// 	var futures []workflow.ChildWorkflowFuture
// 	for _, change := range req.Changes {
// 		// Create individual change request
// 		individualReq := &workflow.FeatureFlagChangeRequest{
// 			TenantID:             req.TenantID,
// 			RequestedBy:          req.RequestedBy,
// 			FlagName:             change.FlagName,
// 			ChangeType:           change.ChangeType,
// 			NewValue:             change.NewValue,
// 			Justification:        req.Justification,
// 			BusinessReason:       req.BusinessReason,
// 			ApprovalTimeoutHours: req.ApprovalTimeoutHours,
// 			Metadata:             change.Metadata,
// 		}
//
// 		future := workflow.ExecuteChildWorkflow(ctx, FeatureFlagWorkflow, individualReq)
// 		futures = append(futures, future)
// 	}
//
// 	// Wait for all child workflows to complete
// 	for i, future := range futures {
// 		var childResult workflow.FeatureFlagChangeResult
// 		err := future.Get(ctx, &childResult)
// 		flagName := req.Changes[i].FlagName
// 		if err != nil {
// 			logger.Error("Child workflow failed", "flag_name", flagName, "error", err)
// 			childResult.Status = "workflow_failed"
// 			childResult.Error = err.Error()
// 		}
//
// 		result.Results[flagName] = &childResult
// 		if childResult.Status == "completed" {
// 			result.SuccessCount++
// 		} else {
// 			result.FailureCount++
// 		}
// 	}
//
// 	result.TotalCount = len(req.Changes)
//
// 	logger.Info("Bulk feature flag workflow completed",
// 		"total", result.TotalCount,
// 		"success", result.SuccessCount,
// 		"failures", result.FailureCount)
//
// 	return &result, nil
// }
//
// // AutoRollbackWorkflow handles automatic rollback of feature flags after expiration
// func AutoRollbackWorkflow(ctx workflow.Context, req *workflow.AutoRollbackRequest) (*workflow.AutoRollbackResult, error) {
// 	logger := workflow.GetLogger(ctx)
// 	logger.Info("Starting auto-rollback workflow", "flag_name", req.FlagName, "rollback_at", req.RollbackAt)
//
// 	// Wait until rollback time
// 	err := workflow.Sleep(ctx, req.RollbackAt.Sub(workflow.Now(ctx)))
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	// Set activity options
// 	ao := workflow.ActivityOptions{
// 		StartToCloseTimeout: 2 * time.Minute,
// 		RetryPolicy: &workflow.RetryPolicy{
// 			MaximumAttempts: 3,
// 		},
// 	}
// 	ctx = workflow.WithActivityOptions(ctx, ao)
//
// 	// Check if flag still needs rollback
// 	var needsRollback bool
// 	err = workflow.ExecuteActivity(ctx, CheckRollbackNeededActivity, req.FlagID).Get(ctx, &needsRollback)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	var result workflow.AutoRollbackResult
// 	result.FlagID = req.FlagID
// 	result.FlagName = req.FlagName
// 	result.ScheduledAt = req.RollbackAt
//
// 	if !needsRollback {
// 		result.Status = "skipped"
// 		result.Reason = "Flag configuration changed, rollback no longer needed"
// 		return &result, nil
// 	}
//
// 	// Perform rollback
// 	rollbackReq := &workflow.FeatureFlagChangeRequest{
// 		TenantID:       req.TenantID,
// 		RequestedBy:    uuid.Nil, // System-initiated
// 		FlagName:       req.FlagName,
// 		ChangeType:     "rollback",
// 		NewValue:       req.RollbackValue,
// 		Justification:  "Automatic rollback after expiration",
// 		BusinessReason: "Scheduled rollback",
// 		Metadata: map[string]interface{}{
// 			"auto_rollback":         true,
// 			"original_scheduled_at": req.RollbackAt,
// 		},
// 	}
//
// 	var changeResult workflow.ApplyChangeResult
// 	err = workflow.ExecuteActivity(ctx, ApplyFeatureFlagChangeActivity, rollbackReq).Get(ctx, &changeResult)
// 	if err != nil {
// 		result.Status = "failed"
// 		result.Error = err.Error()
// 		return &result, err
// 	}
//
// 	// Send notification
// 	notificationReq := workflow.NotificationRequest{
// 		TenantID:   req.TenantID,
// 		FlagName:   req.FlagName,
// 		ChangeType: "auto_rollback",
// 		NewValue:   req.RollbackValue,
// 		AppliedBy:  uuid.Nil,
// 		AppliedAt:  changeResult.AppliedAt,
// 	}
// 	workflow.ExecuteActivity(ctx, SendWebSocketNotificationActivity, notificationReq)
//
// 	// Create audit event
// 	auditReq := workflow.AuditEventRequest{
// 		TenantID:    req.TenantID,
// 		FlagName:    req.FlagName,
// 		ChangeType:  "auto_rollback",
// 		OldValue:    changeResult.OldValue,
// 		NewValue:    req.RollbackValue,
// 		RequestedBy: uuid.Nil,
// 		AppliedAt:   changeResult.AppliedAt,
// 		Metadata: map[string]interface{}{
// 			"auto_rollback": true,
// 			"scheduled_at":  req.RollbackAt,
// 		},
// 	}
// 	workflow.ExecuteActivity(ctx, CreateAuditEventActivity, auditReq)
//
// 	result.Status = "completed"
// 	result.OldValue = changeResult.OldValue
// 	result.NewValue = req.RollbackValue
// 	result.RolledBackAt = changeResult.AppliedAt
//
// 	logger.Info("Auto-rollback completed successfully", "flag_name", req.FlagName)
// 	return &result, nil
// }
