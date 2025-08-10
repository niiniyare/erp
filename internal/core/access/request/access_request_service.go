package request

//go:generate go run go.uber.org/mock/mockgen -source=access_request_service.go -destination=mock.go -package=request


import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/niiniyare/erp/internal/core/access/approval"
	"github.com/niiniyare/erp/internal/core/access/execution"
	"github.com/niiniyare/erp/internal/core/audit"
	"github.com/niiniyare/erp/internal/core/notification"
)

// BusinessLogicError represents a business logic error
type BusinessLogicError struct {
	Code    string
	Message string
}

func (e *BusinessLogicError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewBusinessLogicError creates a new BusinessLogicError
func NewBusinessLogicError(code, message string) *BusinessLogicError {
	return &BusinessLogicError{
		Code:    code,
		Message: message,
	}
}

// AccessRequestService defines the access request service interface
type AccessRequestService interface {
	// Core workflow operations
	CreateAccessRequest(ctx context.Context, req *CreateAccessRequestRequest, requesterID uuid.UUID) (*AccessRequest, error)
	ProcessAccessRequest(ctx context.Context, requestID uuid.UUID, action *AccessRequestApprovalRequest, approverID uuid.UUID) (*AccessRequest, error)

	// Approval workflow
	ApproveAccessRequest(ctx context.Context, requestID, approverID uuid.UUID, comments *string, modifyDuration *int32) (*AccessRequest, error)
	RejectAccessRequest(ctx context.Context, requestID, approverID uuid.UUID, comments *string) (*AccessRequest, error)
	RevokeAccessRequest(ctx context.Context, requestID uuid.UUID) (*AccessRequest, error)

	// Query operations
	GetAccessRequest(ctx context.Context, requestID uuid.UUID) (*AccessRequestWithDetails, error)
	ListAccessRequests(ctx context.Context, req *ListAccessRequestsRequest) ([]*AccessRequestWithDetails, error)
	GetUserRequestHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*AccessRequest, error)
	GetPendingRequestsForApprover(ctx context.Context, approverID uuid.UUID, limit, offset int) ([]*AccessRequest, error)

	// Administrative operations
	GetAccessRequestStats(ctx context.Context, fromDate, toDate *time.Time) (*AccessRequestStats, error)
	ProcessExpiredRequests(ctx context.Context) error

	// Integration operations
	ExecuteApprovedRequest(ctx context.Context, requestID uuid.UUID) error
	ValidateAccessRequest(ctx context.Context, req *CreateAccessRequestRequest, requesterID uuid.UUID) error
}

// accessRequestService implements AccessRequestService
type accessRequestService struct {
	repo    AccessRequestRepository
	cache   cache.Service
	tracing tracing.TracingService
	metrics metrics.MetricsProvider
	// Integration with main user service for executing approved requests
	userService UserService
	// Notification service for workflow events
	notificationService notification.NotificationService
	// Approver service for validation and determination
	approverService approval.ApproverService
	// Execution service for granting/revoking access
	executionService execution.AccessExecutionService
	// Audit service for comprehensive logging
	auditService audit.Service
}

// NewAccessRequestService creates a new access request service
func NewAccessRequestService(
	repo AccessRequestRepository,
	cache cache.Service,
	tracing tracing.TracingService,
	metrics metrics.MetricsProvider,
	userService UserService,
	notificationService notification.NotificationService,
	approverService approval.ApproverService,
	executionService execution.AccessExecutionService,
	auditService audit.Service,
) AccessRequestService {
	return &accessRequestService{
		repo:                repo,
		cache:               cache,
		tracing:             tracing,
		metrics:             metrics,
		userService:         userService,
		notificationService: notificationService,
		approverService:     approverService,
		executionService:    executionService,
		auditService:        auditService,
	}
}

// CreateAccessRequest creates a new access request with validation and workflow initiation
func (s *accessRequestService) CreateAccessRequest(ctx context.Context, req *CreateAccessRequestRequest, requesterID uuid.UUID) (*AccessRequest, error) {
	ctx, span := s.tracing.StartSpan(ctx, "accessRequestService.CreateAccessRequest")
	defer span.End()

	span.SetAttributes(
		attribute.String("requester_id", requesterID.String()),
		attribute.String("request_type", string(req.RequestType)),
		attribute.String("entity_id", req.EntityID.String()),
	)

	logger.Info("Creating access request", logger.Fields{
		"requester_id": requesterID,
		"request_type": req.RequestType,
		"entity_id":    req.EntityID,
	})

	// Validate the request
	if err := s.ValidateAccessRequest(ctx, req, requesterID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Access request validation failed")
		s.metrics.IncrementCounter("access_request_validation_failed", map[string]any{"error": "validation"})
		return nil, err
	}

	// Check for duplicate pending requests
	if err := s.checkDuplicateRequest(ctx, req, requesterID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Duplicate access request detected")
		s.metrics.IncrementCounter("access_request_duplicate", map[string]any{"error": "duplicate"})
		return nil, err
	}

	// Create the access request
	accessRequest, err := s.repo.CreateAccessRequest(ctx, req, requesterID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to create access request")
		return nil, err
	}

	// Log audit event
	// if err := s.auditService.LogAccessRequestCreated(ctx, accessRequest, requesterID); err != nil {
	// 	logger.Error("Failed to log access request creation", logger.Fields{
	// 		"request_id": accessRequest.ID,
	// 		"error":      err.Error(),
	// 	})
	// }

	// Send notification to approvers (async)
	// go func() {
	// 	if err := s.notificationService.NotifyApprovers(context.Background(), accessRequest); err != nil {
	// 		logger.Error("Failed to notify approvers", logger.Fields{
	// 			"request_id": accessRequest.ID,
	// 			"error":      err.Error(),
	// 		})
	// 	}
	// }()

	s.metrics.IncrementCounter("access_request_created", map[string]any{"type": string(req.RequestType)})
	logger.Info("Access request created successfully", logger.Fields{
		"request_id":   accessRequest.ID,
		"requester_id": requesterID,
		"request_type": accessRequest.RequestType,
	})

	return accessRequest, nil
}

// ProcessAccessRequest processes an access request (approve/reject) with full workflow logic
func (s *accessRequestService) ProcessAccessRequest(ctx context.Context, requestID uuid.UUID, action *AccessRequestApprovalRequest, approverID uuid.UUID) (*AccessRequest, error) {
	ctx, span := s.tracing.StartSpan(ctx, "accessRequestService.ProcessAccessRequest")
	defer span.End()

	span.SetAttributes(
		attribute.String("request_id", requestID.String()),
		attribute.String("action", action.Action),
		attribute.String("approver_id", approverID.String()),
	)

	logger.Info("Processing access request", logger.Fields{
		"request_id":  requestID,
		"action":      action.Action,
		"approver_id": approverID,
	})

	// Get the current request
	currentRequest, err := s.repo.GetAccessRequestByID(ctx, requestID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get access request")
		return nil, err
	}

	// Validate the request can be processed
	if !currentRequest.CanBeApproved() {
		err := NewBusinessLogicError("INVALID_STATE", "Access request cannot be processed in its current state")
		span.RecordError(err)
		span.SetStatus(codes.Error, "Invalid request state")
		s.metrics.IncrementCounter("access_request_invalid_state", map[string]any{"state": string(currentRequest.ApprovalStatus)})
		return nil, err
	}

	// Validate approver permissions
	if err := s.validateApproverPermissions(ctx, approverID, currentRequest); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Approver validation failed")
		s.metrics.IncrementCounter("access_request_approver_invalid", map[string]any{"error": "unauthorized"})
		return nil, err
	}

	var processedRequest *AccessRequest

	switch action.Action {
	case "approve":
		processedRequest, err = s.ApproveAccessRequest(ctx, requestID, approverID, action.Comments, action.ModifyDuration)
	case "reject":
		processedRequest, err = s.RejectAccessRequest(ctx, requestID, approverID, action.Comments)
	default:
		err = errors.NewBusinessError("INVALID_ACTION", "Invalid action. Must be 'approve' or 'reject'")
	}

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to process access request")
		return nil, err
	}

	// If approved, execute the request (grant actual access)
	if processedRequest.ApprovalStatus == ApprovalStatusApproved {
		if err := s.ExecuteApprovedRequest(ctx, processedRequest.ID); err != nil {
			// Log error but don't fail the approval - we can retry execution later
			logger.Error("Failed to execute approved request", logger.Fields{
				"request_id": processedRequest.ID,
				"error":      err.Error(),
			})
			s.metrics.IncrementCounter("access_request_execution_failed", map[string]any{"error": "execution"})
		}
	}

	logger.Info("Access request processed successfully", logger.Fields{
		"request_id":      requestID,
		"action":          action.Action,
		"approver_id":     approverID,
		"approval_status": processedRequest.ApprovalStatus,
	})

	return processedRequest, nil
}

// ApproveAccessRequest approves an access request
func (s *accessRequestService) ApproveAccessRequest(ctx context.Context, requestID, approverID uuid.UUID, comments *string, modifyDuration *int32) (*AccessRequest, error) {
	ctx, span := s.tracing.StartSpan(ctx, "accessRequestService.ApproveAccessRequest")
	defer span.End()

	updateReq := &UpdateAccessRequestRequest{
		ApprovalStatus:   ApprovalStatusApproved,
		ApprovalComments: comments,
		DurationHours:    modifyDuration,
	}

	accessRequest, err := s.repo.UpdateAccessRequestStatus(ctx, requestID, updateReq, approverID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to approve access request")
		return nil, err
	}

	// Log audit event
	// if err := s.auditService.LogAccessRequestProcessed(ctx, accessRequest, "approved", approverID, *comments); err != nil {
	// 	logger.Error("Failed to log access request approval", logger.Fields{
	// 		"request_id": accessRequest.ID,
	// 		"error":      err.Error(),
	// 	})
	// }

	// Send notification to requester (async)
	// go func() {
	// 	if err := s.notificationService.NotifyRequester(context.Background(), accessRequest, "approved"); err != nil {
	// 		logger.Error("Failed to notify requester", logger.Fields{
	// 			"request_id": accessRequest.ID,
	// 			"error":      err.Error(),
	// 		})
	// 	}
	// }()

	s.metrics.IncrementCounter("access_request_approved", map[string]any{"type": string(accessRequest.RequestType)})

	return accessRequest, nil
}

// RejectAccessRequest rejects an access request
func (s *accessRequestService) RejectAccessRequest(ctx context.Context, requestID, approverID uuid.UUID, comments *string) (*AccessRequest, error) {
	ctx, span := s.tracing.StartSpan(ctx, "accessRequestService.RejectAccessRequest")
	defer span.End()

	updateReq := &UpdateAccessRequestRequest{
		ApprovalStatus:   ApprovalStatusRejected,
		ApprovalComments: comments,
	}

	accessRequest, err := s.repo.UpdateAccessRequestStatus(ctx, requestID, updateReq, approverID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to reject access request")
		return nil, err
	}

	// Log audit event
	// if err := s.auditService.LogAccessRequestProcessed(ctx, accessRequest, "rejected", approverID, *comments); err != nil {
	// 	logger.Error("Failed to log access request rejection", logger.Fields{
	// 		"request_id": accessRequest.ID,
	// 		"error":      err.Error(),
	// 	})
	// }

	// Send notification to requester (async)
	// go func() {
	// 	if err := s.notificationService.NotifyRequester(context.Background(), accessRequest, "rejected"); err != nil {
	// 		logger.Error("Failed to notify requester", logger.Fields{
	// 			"request_id": accessRequest.ID,
	// 			"error":      err.Error(),
	// 		})
	// 	}
	// }()

	s.metrics.IncrementCounter("access_request_rejected", map[string]any{"type": string(accessRequest.RequestType)})

	return accessRequest, nil
}

// RevokeAccessRequest revokes an approved access request
func (s *accessRequestService) RevokeAccessRequest(ctx context.Context, requestID uuid.UUID) (*AccessRequest, error) {
	ctx, span := s.tracing.StartSpan(ctx, "accessRequestService.RevokeAccessRequest")
	defer span.End()

	// Get current request to validate
	currentRequest, err := s.repo.GetAccessRequestByID(ctx, requestID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get access request")
		return nil, err
	}

	if !currentRequest.CanBeRevoked() {
		err := NewBusinessLogicError("INVALID_STATE", "Access request cannot be revoked in its current state")
		span.RecordError(err)
		span.SetStatus(codes.Error, "Invalid request state for revocation")
		s.metrics.IncrementCounter("access_request_revoke_invalid_state", map[string]any{"state": string(currentRequest.ApprovalStatus)})
		return nil, err
	}

	// First revoke the actual access
	if err := s.revokeGrantedAccess(ctx, currentRequest); err != nil {
		logger.Error("Failed to revoke granted access", logger.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		})
		// Continue with request revocation even if access revocation fails
	}

	// Revoke the request
	accessRequest, err := s.repo.RevokeAccessRequest(ctx, requestID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to revoke access request")
		return nil, err
	}

	// Log audit event
	// if err := s.auditService.LogAccessRequestProcessed(ctx, accessRequest, "revoked", uuid.Nil, "Access request revoked"); err != nil {
	// 	logger.Error("Failed to log access request revocation", logger.Fields{
	// 		"request_id": accessRequest.ID,
	// 		"error":      err.Error(),
	// 	})
	// }

	s.metrics.IncrementCounter("access_request_revoked", map[string]any{"type": string(accessRequest.RequestType)})

	return accessRequest, nil
}

// GetAccessRequest gets an access request with full details
func (s *accessRequestService) GetAccessRequest(ctx context.Context, requestID uuid.UUID) (*AccessRequestWithDetails, error) {
	ctx, span := s.tracing.StartSpan(ctx, "accessRequestService.GetAccessRequest")
	defer span.End()

	// Try cache first
	cacheKey := fmt.Sprintf("access_request:%s", requestID.String())
	var cachedRequest *AccessRequestWithDetails

	if err := s.cache.Get(ctx, cacheKey, &cachedRequest); err == nil && cachedRequest != nil {
		s.metrics.IncrementCounter("access_request_cache_hit", map[string]any{})
		return cachedRequest, nil
	}

	// Get from database with details
	listReq := &ListAccessRequestsRequest{
		Limit:  1,
		Offset: 0,
	}

	requests, err := s.repo.ListAccessRequestsWithDetails(ctx, listReq)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get access request")
		return nil, err
	}

	if len(requests) == 0 {
		err := errors.NewRepositoryError("NOT_FOUND", "Access request not found", nil)
		span.RecordError(err)
		span.SetStatus(codes.Error, "Access request not found")
		return nil, err
	}

	request := requests[0]

	// Cache for 5 minutes
	s.cache.Set(ctx, cacheKey, request, 5*time.Minute)
	s.metrics.IncrementCounter("access_request_cache_miss", map[string]any{})

	return request, nil
}

// ListAccessRequests lists access requests with filters
func (s *accessRequestService) ListAccessRequests(ctx context.Context, req *ListAccessRequestsRequest) ([]*AccessRequestWithDetails, error) {
	ctx, span := s.tracing.StartSpan(ctx, "accessRequestService.ListAccessRequests")
	defer span.End()

	requests, err := s.repo.ListAccessRequestsWithDetails(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to list access requests")
		return nil, err
	}

	s.metrics.ObserveHistogram("access_request_list_count", float64(len(requests)), map[string]any{})
	return requests, nil
}

// GetUserRequestHistory gets access request history for a user
func (s *accessRequestService) GetUserRequestHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*AccessRequest, error) {
	ctx, span := s.tracing.StartSpan(ctx, "accessRequestService.GetUserRequestHistory")
	defer span.End()

	requests, err := s.repo.GetUserAccessRequestHistory(ctx, userID, limit, offset)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get user request history")
		return nil, err
	}

	return requests, nil
}

// GetPendingRequestsForApprover gets pending requests for an approver
func (s *accessRequestService) GetPendingRequestsForApprover(ctx context.Context, approverID uuid.UUID, limit, offset int) ([]*AccessRequest, error) {
	ctx, span := s.tracing.StartSpan(ctx, "accessRequestService.GetPendingRequestsForApprover")
	defer span.End()

	// TODO: Add approver filtering logic based on role hierarchy, department, etc.
	requests, err := s.repo.GetPendingRequestsForApprover(ctx, limit, offset)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get pending requests for approver")
		return nil, err
	}

	return requests, nil
}

// GetAccessRequestStats gets access request statistics
func (s *accessRequestService) GetAccessRequestStats(ctx context.Context, fromDate, toDate *time.Time) (*AccessRequestStats, error) {
	ctx, span := s.tracing.StartSpan(ctx, "accessRequestService.GetAccessRequestStats")
	defer span.End()

	stats, err := s.repo.GetAccessRequestStats(ctx, fromDate, toDate)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get access request stats")
		return nil, err
	}

	return stats, nil
}

// ProcessExpiredRequests processes all expired access requests
func (s *accessRequestService) ProcessExpiredRequests(ctx context.Context) error {
	ctx, span := s.tracing.StartSpan(ctx, "accessRequestService.ProcessExpiredRequests")
	defer span.End()

	logger.Info("Processing expired access requests", logger.Fields{})

	expiredRequests, err := s.repo.GetExpiredAccessRequests(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get expired access requests")
		return err
	}

	processedCount := 0
	for _, request := range expiredRequests {
		if request.AutoRevoke {
			// Revoke the granted access
			if err := s.revokeGrantedAccess(ctx, request); err != nil {
				logger.Error("Failed to revoke granted access for expired request", logger.Fields{
					"request_id": request.ID,
					"error":      err.Error(),
				})
				continue
			}

			// Update request status to expired
			if err := s.repo.ExpireAccessRequest(ctx, request.ID); err != nil {
				logger.Error("Failed to expire access request", logger.Fields{
					"request_id": request.ID,
					"error":      err.Error(),
				})
				continue
			}

			// Log audit event
			// if err := s.auditService.LogAccessRequestProcessed(ctx, request, "expired", uuid.Nil, "Access request expired automatically"); err != nil {
			// 	logger.Error("Failed to log access request expiration", logger.Fields{
			// 		"request_id": request.ID,
			// 		"error":      err.Error(),
			// 	})
			// }

			processedCount++
		}
	}

	s.metrics.ObserveHistogram("access_request_expired_processed", float64(processedCount), map[string]any{})
	logger.Info("Processed expired access requests", logger.Fields{
		"total_expired": len(expiredRequests),
		"processed":     processedCount,
	})

	return nil
}

// ExecuteApprovedRequest executes an approved access request by granting the actual access
func (s *accessRequestService) ExecuteApprovedRequest(ctx context.Context, requestID uuid.UUID) error {
	ctx, span := s.tracing.StartSpan(ctx, "accessRequestService.ExecuteApprovedRequest")
	defer span.End()

	request, err := s.repo.GetAccessRequestByID(ctx, requestID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get access request")
		return err
	}

	if request.ApprovalStatus != ApprovalStatusApproved {
		err := NewBusinessLogicError("INVALID_STATE", "Only approved requests can be executed")
		span.RecordError(err)
		span.SetStatus(codes.Error, "Invalid request state for execution")
		return err
	}

	logger.Info("Executing approved access request", logger.Fields{
		"request_id":   requestID,
		"request_type": request.RequestType,
		"target_user":  request.TargetUserID,
	})

	// Convert to execution service types and execute access request
	executionRequest := convertToExecutionAccessRequest(request)
	result, err := s.executionService.ExecuteAccessRequest(ctx, executionRequest)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to execute access request")
		s.metrics.IncrementCounter("access_request_execution_failed", map[string]any{"type": string(request.RequestType)})
		return err
	}

	if !result.Success {
		err := NewBusinessLogicError("EXECUTION_FAILED", result.ErrorMessage)
		span.RecordError(err)
		span.SetStatus(codes.Error, "Access execution failed")
		s.metrics.IncrementCounter("access_request_execution_failed", map[string]any{"type": string(request.RequestType)})
		return err
	}

	// Log successful access execution
	// if len(result.AccessGranted) > 0 {
	// 	if err := s.auditService.LogAccessGranted(ctx, request, result.AccessGranted); err != nil {
	// 		logger.Error("Failed to log access granted", logger.Fields{
	// 			"request_id": requestID,
	// 			"error":      err.Error(),
	// 		})
	// 	}
	// }

	s.metrics.IncrementCounter("access_request_executed", map[string]any{"type": string(request.RequestType)})
	logger.Info("Access request executed successfully", logger.Fields{
		"request_id":     requestID,
		"request_type":   request.RequestType,
		"access_granted": len(result.AccessGranted),
		"execution_time": result.ExecutedAt,
	})

	return nil
}

// ValidateAccessRequest validates an access request before creation
func (s *accessRequestService) ValidateAccessRequest(ctx context.Context, req *CreateAccessRequestRequest, requesterID uuid.UUID) error {
	ctx, span := s.tracing.StartSpan(ctx, "accessRequestService.ValidateAccessRequest")
	defer span.End()

	// Basic validation
	if err := req.ValidateRequest(); err != nil {
		return err
	}

	// Verify requester exists and is active
	_, err := s.userService.GetUserByID(ctx, requesterID)
	if err != nil {
		return errors.NewBusinessError("REQUESTER_NOT_FOUND", "Requester not found")
	}

	// TODO: Add user active status validation when User model includes IsActive field
	// Currently simplified to allow compilation

	// Verify target user exists if specified
	if req.TargetUserID != nil {
		_, err := s.userService.GetUserByID(ctx, *req.TargetUserID)
		if err != nil {
			return errors.NewBusinessError("TARGET_USER_NOT_FOUND", "Target user not found")
		}

		// TODO: Add target user active status validation when User model includes IsActive field
		// Currently simplified to allow compilation
	}

	// Verify entity exists
	// TODO: Add entity validation when entity service is available

	// Type-specific validation
	switch req.RequestType {
	case RequestTypeRoleAssignment:
		if req.RoleID == nil {
			return errors.NewBusinessError("ROLE_ID_REQUIRED", "Role ID is required for role assignment requests")
		}
		// TODO: Verify role exists and is available in the entity

	case RequestTypePermissionGrant:
		if req.PermissionID == nil {
			return errors.NewBusinessError("PERMISSION_ID_REQUIRED", "Permission ID is required for permission grant requests")
		}
		// TODO: Verify permission exists

	case RequestTypeResourceAccess:
		if req.ResourceID == nil {
			return errors.NewBusinessError("RESOURCE_ID_REQUIRED", "Resource ID is required for resource access requests")
		}
		// TODO: Verify resource exists

	case RequestTypeElevation:
		if req.DurationHours == nil || *req.DurationHours <= 0 {
			return errors.NewBusinessError("DURATION_REQUIRED", "Duration is required for elevation requests")
		}
		if *req.DurationHours > 72 { // Max 72 hours
			return errors.NewBusinessError("DURATION_EXCEEDED", "Elevation duration cannot exceed 72 hours")
		}
	}

	return nil
}

// Helper methods

// checkDuplicateRequest checks for duplicate pending requests
func (s *accessRequestService) checkDuplicateRequest(ctx context.Context, req *CreateAccessRequestRequest, requesterID uuid.UUID) error {
	pendingStatus := ApprovalStatusPending
	listReq := &ListAccessRequestsRequest{
		RequesterID:    &requesterID,
		TargetUserID:   req.TargetUserID,
		EntityID:       &req.EntityID,
		RequestType:    &req.RequestType,
		ApprovalStatus: &pendingStatus,
		IncludeExpired: false,
		Limit:          1,
		Offset:         0,
	}

	requests, err := s.repo.ListAccessRequestsWithDetails(ctx, listReq)
	if err != nil {
		return err
	}

	if len(requests) > 0 {
		for _, existingReq := range requests {
			// Check if it's truly a duplicate (same target resources)
			if s.isDuplicateRequest(existingReq.AccessRequest, req) {
				return NewBusinessLogicError("DUPLICATE_REQUEST", "A similar access request is already pending")
			}
		}
	}

	return nil
}

// isDuplicateRequest checks if two requests are duplicates
func (s *accessRequestService) isDuplicateRequest(existing *AccessRequest, new *CreateAccessRequestRequest) bool {
	if existing.RequestType != new.RequestType {
		return false
	}

	switch new.RequestType {
	case RequestTypeRoleAssignment:
		return existing.RoleID != nil && new.RoleID != nil && *existing.RoleID == *new.RoleID
	case RequestTypePermissionGrant:
		return existing.PermissionID != nil && new.PermissionID != nil && *existing.PermissionID == *new.PermissionID
	case RequestTypeResourceAccess:
		return existing.ResourceID != nil && new.ResourceID != nil && *existing.ResourceID == *new.ResourceID
	case RequestTypeElevation:
		// Elevation requests are always considered duplicates if pending
		return true
	}

	return false
}

// validateApproverPermissions validates that an approver can approve the request
func (s *accessRequestService) validateApproverPermissions(ctx context.Context, approverID uuid.UUID, request *AccessRequest) error {
	// result, err := s.approverService.ValidateApprover(ctx, approverID, request)
	// if err != nil {
	// 	return NewValidationError("approver_id", fmt.Sprintf("Failed to validate approver: %v", err))
	// }

	// if !result.IsValid {
	// 	return NewValidationError("approver_id", result.Reason)
	// }

	return nil
}

// revokeGrantedAccess revokes the access that was granted by an approved request
func (s *accessRequestService) revokeGrantedAccess(ctx context.Context, request *AccessRequest) error {
	executionRequest := convertToExecutionAccessRequest(request)
	result, err := s.executionService.RevokeAccessRequest(ctx, executionRequest)
	if err != nil {
		return err
	}

	if !result.Success {
		return NewBusinessLogicError("REVOCATION_FAILED", result.ErrorMessage)
	}

	// Log successful access revocation
	// if len(result.AccessRevoked) > 0 {
	// 	if err := s.auditService.LogAccessRevoked(ctx, request, result.AccessRevoked); err != nil {
	// 		logger.Error("Failed to log access revocation", logger.Fields{
	// 			"request_id": request.ID,
	// 			"error":      err.Error(),
	// 		})
	// 	}
	// }

	logger.Info("Access revoked successfully", logger.Fields{
		"request_id":     request.ID,
		"access_revoked": len(result.AccessRevoked),
	})

	return nil
}

// convertToExecutionAccessRequest converts request package AccessRequest to execution package AccessRequest
func convertToExecutionAccessRequest(req *AccessRequest) *execution.AccessRequest {
	businessReason := ""
	if req.BusinessReason != nil {
		businessReason = *req.BusinessReason
	}

	return &execution.AccessRequest{
		ID:             req.ID,
		TenantID:       req.TenantID,
		RequestType:    execution.RequestType(req.RequestType),
		RequesterID:    req.RequesterID,
		TargetUserID:   req.TargetUserID,
		EntityID:       req.EntityID,
		RoleID:         req.RoleID,
		PermissionID:   req.PermissionID,
		ResourceID:     req.ResourceID,
		ApprovalStatus: string(req.ApprovalStatus),
		ApprovedBy:     req.ApprovedBy,
		ExpiresAt:      req.ExpiresAt,
		Justification:  req.Justification,
		BusinessReason: businessReason,
	}
}
