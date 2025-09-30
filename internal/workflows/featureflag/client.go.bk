package featureflag

// import (
// 	// "context"
// 	// "fmt"
// 	// "time"
// 	//
// 	// "github.com/docker/docker/registry/resumable"
// 	// "github.com/google/uuid"
// 	// "github.com/niiniyare/erp/internal/featureflag/workflow"
// 	// "github.com/niiniyare/erp/internal/shared/logger"
// 	"go.temporal.io/sdk/client"
// 	"go.temporal.io/sdk/workflow"
// )
//
// // workflowClient implements the workflow.Client interface
// type workflowClient struct {
// 	temporalClient client.Client
// }
//
// // NewWorkflowClient creates a new client for feature flag workflows
// func NewWorkflowClient(temporalClient client.Client) workflow.Client {
// 	return &workflowClient{
// 		temporalClient: temporalClient,
// 	}
// }
//
// // RequestFeatureFlagChange initiates a feature flag change workflow
// func (c *workflowClient) RequestFeatureFlagChange(ctx context.Context, req *workflow.FeatureFlagChangeRequestInput) (*workflow.FeatureFlagWorkflowResult, error) {
// 	// log := logger.WithFields(logger.Fields{
// 	// 	"client":    "workflowClient",
// 	// 	"method":    "RequestFeatureFlagChange",
// 	// 	"flag_name": req.FlagName,
// 	// })
// 	//
// 	// // NOTE:This is where the logic from the old service goes.
// 	// // In a real implementation, we would get tenant and user from context.
// 	// userID := uuid.New()   // Placeholder
// 	// tenantID := uuid.New() // Placeholder
// 	//
// 	// workflowReq := &workflow.FeatureFlagChangeRequest{
// 	// 	TenantID:             tenantID,
// 	// 	RequestedBy:          userID,
// 	// 	FlagName:             req.FlagName,
// 	// 	ChangeType:           req.ChangeType,
// 	// 	NewValue:             req.NewValue,
// 	// 	Justification:        req.Justification,
// 	// 	BusinessReason:       req.BusinessReason,
// 	// 	ApprovalTimeoutHours: req.ApprovalTimeoutHours,
// 	// 	Metadata:             req.Metadata,
// 	// }
// 	//
// 	// if workflowReq.ApprovalTimeoutHours == 0 {
// 	// 	workflowReq.ApprovalTimeoutHours = 24
// 	// }
// 	//
// 	// workflowID := fmt.Sprintf("feature-flag-change-%s-%d", req.FlagName, time.Now().Unix())
// 	//
// 	// workflowOptions := client.StartWorkflowOptions{
// 	// 	ID:        workflowID,
// 	// 	TaskQueue: "feature-flag-workflows",
// 	// }
// 	//
// 	// execution, err := c.temporalClient.ExecuteWorkflow(ctx, workflowOptions, FeatureFlagWorkflow, workflowReq)
// 	// if err != nil {
// 	// 	log.Error("Failed to start workflow", logger.Fields{"error": err})
// 	// 	return nil, fmt.Errorf("failed to start workflow: %w", err)
// 	// }
// 	//
// 	// result := &workflow.FeatureFlagWorkflowResult{
// 	// 	WorkflowID:       execution.GetID(),
// 	// 	Status:           "running",
// 	// 	FlagName:         req.FlagName,
// 	// 	ChangeType:       req.ChangeType,
// 	// 	RequiresApproval: true, // This will be determined by the workflow
// 	// 	CreatedAt:        time.Now(),
// 	// }
// 	//
// 	// logger.Info("Workflow started successfully")
// 	return nil, nil
// }
//
// func (c *workflowClient) RequestBulkFeatureFlagChange(ctx context.Context, req *workflow.BulkFeatureFlagChangeRequestInput) (*workflow.BulkFeatureFlagWorkflowResult, error) {
// 	// NOTE: Not implemented
// 	return nil, fmt.Errorf("not implemented")
// }
//
// func (c *workflowClient) ApproveFeatureFlagChange(ctx context.Context, workflowID string, approvalReq *workflow.ApprovalRequestInput) error {
// 	signal := workflow.ApprovalSignal{
// 		Approved:   true,
// 		ApproverID: approvalReq.ApproverID,
// 		Comments:   approvalReq.Comments,
// 		ApprovedAt: time.Now(),
// 	}
// 	return c.temporalClient.SignalWorkflow(ctx, workflowID, "", "approval-signal", signal)
// }
//
// func (c *workflowClient) RejectFeatureFlagChange(ctx context.Context, workflowID string, rejectionReq *workflow.RejectionRequestInput) error {
// 	signal := workflow.ApprovalSignal{
// 		Approved:   false,
// 		ApproverID: rejectionReq.ApproverID,
// 		Comments:   rejectionReq.Comments,
// 		ApprovedAt: time.Now(),
// 	}
// 	return c.temporalClient.SignalWorkflow(ctx, workflowID, "", "approval-signal", signal)
// }
//
// func (c *workflowClient) GetWorkflowStatus(ctx context.Context, workflowID string) (*workflow.WorkflowStatus, error) {
// 	// NOTE: Not implemented
// 	return nil, fmt.Errorf("not implemented")
// }
//
// func (c *workflowClient) ListPendingApprovals(ctx context.Context, req *workflow.ListPendingApprovalsRequest) (*workflow.ListPendingApprovalsResponse, error) {
// 	// NOTE: Not implemented
// 	return nil, fmt.Errorf("not implemented")
// }
//
// func (c *workflowClient) CancelWorkflow(ctx context.Context, workflowID string, reason string) error {
// 	return c.temporalClient.CancelWorkflow(ctx, workflowID, "")
// }
//
// func (c *workflowClient) ScheduleAutoRollback(ctx context.Context, req *workflow.ScheduleAutoRollbackRequest) (*workflow.AutoRollbackScheduleResult, error) {
// 	// NOTE: Not implemented
// 	return nil, fmt.Errorf("not implemented")
// }
//
// func (c *workflowClient) CancelAutoRollback(ctx context.Context, scheduleID string) error {
// 	// NOTE: Not implemented
// 	return fmt.Errorf("not implemented")
// }
