package featureflag

//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"
//
// import (
// 	"context"
//
// 	db "awo/db/sqlc"
// 	"awo/internal/core/access/request"
// 	"awo/internal/core/tenant"
// 	"awo/internal/featureflag/workflow"
// 	"awo/internal/shared/logger"
// )
//
// // WorkflowService handles feature flag workflows with approval processes.
// // It acts as a client to the underlying workflow engine.
// type WorkflowService interface {
// 	RequestFeatureFlagChange(ctx context.Context, req *workflow.FeatureFlagChangeRequestInput) (*workflow.FeatureFlagWorkflowResult, error)
// 	RequestBulkFeatureFlagChange(ctx context.Context, req *workflow.BulkFeatureFlagChangeRequestInput) (*workflow.BulkFeatureFlagWorkflowResult, error)
// 	ApproveFeatureFlagChange(ctx context.Context, workflowID string, approvalReq *workflow.ApprovalRequestInput) error
// 	RejectFeatureFlagChange(ctx context.Context, workflowID string, rejectionReq *workflow.RejectionRequestInput) error
// 	GetWorkflowStatus(ctx context.Context, workflowID string) (*workflow.WorkflowStatus, error)
// 	ListPendingApprovals(ctx context.Context, req *workflow.ListPendingApprovalsRequest) (*workflow.ListPendingApprovalsResponse, error)
// 	CancelWorkflow(ctx context.Context, workflowID string, reason string) error
// 	ScheduleAutoRollback(ctx context.Context, req *workflow.ScheduleAutoRollbackRequest) (*workflow.AutoRollbackScheduleResult, error)
// 	CancelAutoRollback(ctx context.Context, scheduleID string) error
// }
//
// // workflowService implements WorkflowService
// type workflowService struct {
// 	workflowClient     workflow.Client
// 	featureFlagService Service
// 	accessRequestRepo  request.AccessRequestRepository
// 	tenantService      tenant.Service
// 	store              db.Store
// 	webSocketService   WebSocketService
// }
//
// // NewWorkflowService creates a new workflow service
// func NewWorkflowService(
// 	workflowClient workflow.Client,
// 	featureFlagService Service,
// 	accessRequestRepo request.AccessRequestRepository,
// 	tenantService tenant.Service,
// 	store db.Store,
// 	webSocketService WebSocketService,
// ) WorkflowService {
// 	return &workflowService{
// 		workflowClient:     workflowClient,
// 		featureFlagService: featureFlagService,
// 		accessRequestRepo:  accessRequestRepo,
// 		tenantService:      tenantService,
// 		store:              store,
// 		webSocketService:   webSocketService,
// 	}
// }
//
// // RequestFeatureFlagChange initiates a feature flag change workflow
// func (s *workflowService) RequestFeatureFlagChange(ctx context.Context, req *workflow.FeatureFlagChangeRequestInput) (*workflow.FeatureFlagWorkflowResult, error) {
// 	logger := logger.WithFields(logger.Fields{
// 		"service":   "workflowService",
// 		"method":    "RequestFeatureFlagChange",
// 		"flag_name": req.FlagName,
// 	})
// 	logger.Info("Forwarding request to workflow client")
// 	return s.workflowClient.RequestFeatureFlagChange(ctx, req)
// }
//
// // RequestBulkFeatureFlagChange initiates a bulk feature flag change workflow
// func (s *workflowService) RequestBulkFeatureFlagChange(ctx context.Context, req *workflow.BulkFeatureFlagChangeRequestInput) (*workflow.BulkFeatureFlagWorkflowResult, error) {
// 	logger := logger.WithFields(logger.Fields{
// 		"service":      "workflowService",
// 		"method":       "RequestBulkFeatureFlagChange",
// 		"change_count": len(req.Changes),
// 	})
// 	logger.Info("Forwarding bulk request to workflow client")
// 	return s.workflowClient.RequestBulkFeatureFlagChange(ctx, req)
// }
//
// // ApproveFeatureFlagChange approves a pending feature flag change
// func (s *workflowService) ApproveFeatureFlagChange(ctx context.Context, workflowID string, approvalReq *workflow.ApprovalRequestInput) error {
// 	logger := logger.WithFields(logger.Fields{
// 		"service":     "workflowService",
// 		"method":      "ApproveFeatureFlagChange",
// 		"workflow_id": workflowID,
// 	})
// 	logger.Info("Forwarding approval to workflow client")
// 	return s.workflowClient.ApproveFeatureFlagChange(ctx, workflowID, approvalReq)
// }
//
// // RejectFeatureFlagChange rejects a pending feature flag change
// func (s *workflowService) RejectFeatureFlagChange(ctx context.Context, workflowID string, rejectionReq *workflow.RejectionRequestInput) error {
// 	logger := logger.WithFields(logger.Fields{
// 		"service":     "workflowService",
// 		"method":      "RejectFeatureFlagChange",
// 		"workflow_id": workflowID,
// 	})
// 	logger.Info("Forwarding rejection to workflow client")
// 	return s.workflowClient.RejectFeatureFlagChange(ctx, workflowID, rejectionReq)
// }
//
// // GetWorkflowStatus retrieves the current status of a workflow
// func (s *workflowService) GetWorkflowStatus(ctx context.Context, workflowID string) (*workflow.WorkflowStatus, error) {
// 	return s.workflowClient.GetWorkflowStatus(ctx, workflowID)
// }
//
// // ListPendingApprovals lists pending approval requests
// func (s *workflowService) ListPendingApprovals(ctx context.Context, req *workflow.ListPendingApprovalsRequest) (*workflow.ListPendingApprovalsResponse, error) {
// 	return s.workflowClient.ListPendingApprovals(ctx, req)
// }
//
// // CancelWorkflow cancels a running workflow
// func (s *workflowService) CancelWorkflow(ctx context.Context, workflowID string, reason string) error {
// 	return s.workflowClient.CancelWorkflow(ctx, workflowID, reason)
// }
//
// // ScheduleAutoRollback schedules automatic rollback for a feature flag
// func (s *workflowService) ScheduleAutoRollback(ctx context.Context, req *workflow.ScheduleAutoRollbackRequest) (*workflow.AutoRollbackScheduleResult, error) {
// 	return s.workflowClient.ScheduleAutoRollback(ctx, req)
// }
//
// // CancelAutoRollback cancels a scheduled auto-rollback
// func (s *workflowService) CancelAutoRollback(ctx context.Context, scheduleID string) error {
// 	return s.workflowClient.CancelAutoRollback(ctx, scheduleID)
// }
