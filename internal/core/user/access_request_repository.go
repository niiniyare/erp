package user

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// AccessRequestRepository defines the access request repository interface
type AccessRequestRepository interface {
	// Core CRUD operations
	CreateAccessRequest(ctx context.Context, req *CreateAccessRequestRequest, requesterID uuid.UUID) (*AccessRequest, error)
	GetAccessRequestByID(ctx context.Context, id uuid.UUID) (*AccessRequest, error)
	UpdateAccessRequestStatus(ctx context.Context, id uuid.UUID, req *UpdateAccessRequestRequest, approverID uuid.UUID) (*AccessRequest, error)

	// List and search operations
	ListAccessRequests(ctx context.Context, req *ListAccessRequestsRequest) ([]*AccessRequest, error)
	ListAccessRequestsWithDetails(ctx context.Context, req *ListAccessRequestsRequest) ([]*AccessRequestWithDetails, error)
	GetPendingRequests(ctx context.Context, limit, offset int) ([]*AccessRequest, error)
	GetPendingRequestsForApprover(ctx context.Context, limit, offset int) ([]*AccessRequest, error)

	// User-specific operations
	GetUserAccessRequestHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*AccessRequest, error)

	// Workflow operations
	ApproveAccessRequest(ctx context.Context, id, approverID uuid.UUID, comments *string) (*AccessRequest, error)
	RejectAccessRequest(ctx context.Context, id, approverID uuid.UUID, comments *string) (*AccessRequest, error)
	RevokeAccessRequest(ctx context.Context, id uuid.UUID) (*AccessRequest, error)
	ExpireAccessRequest(ctx context.Context, id uuid.UUID) error

	// Statistics and monitoring
	GetAccessRequestStats(ctx context.Context, fromDate, toDate *time.Time) (*AccessRequestStats, error)
	GetExpiredAccessRequests(ctx context.Context) ([]*AccessRequest, error)
}

// accessRequestRepository implements AccessRequestRepository
type accessRequestRepository struct {
	db      db.Store
	tracing *tracing.TracingService
	metrics *metrics.MetricsService
}

// NewAccessRequestRepository creates a new access request repository
func NewAccessRequestRepository(database db.Store, tracing *tracing.TracingService, metrics *metrics.MetricsService) AccessRequestRepository {
	return &accessRequestRepository{
		db:      database,
		tracing: tracing,
		metrics: metrics,
	}
}

// CreateAccessRequest creates a new access request
func (r *accessRequestRepository) CreateAccessRequest(ctx context.Context, req *CreateAccessRequestRequest, requesterID uuid.UUID) (*AccessRequest, error) {
	ctx, span := r.tracing.StartSpan(ctx, "accessRequestRepository.CreateAccessRequest")
	defer span.End()

	logger.Info("Creating access request", logger.Fields{
		"requester_id": requesterID,
		"request_type": req.RequestType,
		"entity_id":    req.EntityID,
	})

	// Get current tenant ID from context - this should be set by middleware
	tenantIDInterface, err := r.db.GetCurrentTenantID(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get tenant ID from context")
		r.metrics.IncrementCounter("access_request_create_error", map[string]any{"error": "no_tenant_id"})
		return nil, errors.NewRepositoryError("TENANT_REQUIRED", "Tenant ID is required", err)
	}

	// Convert tenant ID to UUID
	tenantID, ok := tenantIDInterface.(uuid.UUID)
	if !ok {
		err = fmt.Errorf("invalid tenant ID type: %T", tenantIDInterface)
		span.RecordError(err)
		span.SetStatus(codes.Error, "Invalid tenant ID type")
		r.metrics.IncrementCounter("access_request_create_error", map[string]any{"error": "invalid_tenant_id"})
		return nil, errors.NewRepositoryError("TENANT_INVALID", "Invalid tenant ID", err)
	}

	params := req.ToSQLCCreateParams(tenantID, requesterID)

	sqlcRequest, err := r.db.CreateAccessRequest(ctx, *params)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to create access request")
		r.metrics.IncrementCounter("access_request_create_error", map[string]any{"error": "db_error"})
		return nil, errors.NewRepositoryError("CREATE_FAILED", "Failed to create access request", err)
	}

	accessRequest, err := FromSQLCAccessRequest(sqlcRequest)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to convert access request")
		return nil, errors.NewRepositoryError("CONVERSION_FAILED", "Failed to convert access request", err)
	}

	r.metrics.IncrementCounter("access_request_created", map[string]any{"type": string(req.RequestType)})
	logger.Info("Access request created successfully", logger.Fields{
		"request_id":   accessRequest.ID,
		"requester_id": requesterID,
		"request_type": accessRequest.RequestType,
	})

	return accessRequest, nil
}

// GetAccessRequestByID retrieves an access request by ID
func (r *accessRequestRepository) GetAccessRequestByID(ctx context.Context, id uuid.UUID) (*AccessRequest, error) {
	ctx, span := r.tracing.StartSpan(ctx, "accessRequestRepository.GetAccessRequestByID")
	defer span.End()

	span.SetAttributes(attribute.String("access_request.id", id.String()))

	sqlcRequest, err := r.db.GetAccessRequestByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get access request")
		r.metrics.IncrementCounter("access_request_get_error", map[string]any{"error": "not_found"})
		return nil, errors.NewRepositoryError("NOT_FOUND", "Access request not found", err)
	}

	accessRequest, err := FromSQLCAccessRequest(sqlcRequest)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to convert access request")
		return nil, errors.NewRepositoryError("CONVERSION_FAILED", "Failed to convert access request", err)
	}

	return accessRequest, nil
}

// UpdateAccessRequestStatus updates the status of an access request
func (r *accessRequestRepository) UpdateAccessRequestStatus(ctx context.Context, id uuid.UUID, req *UpdateAccessRequestRequest, approverID uuid.UUID) (*AccessRequest, error) {
	ctx, span := r.tracing.StartSpan(ctx, "accessRequestRepository.UpdateAccessRequestStatus")
	defer span.End()

	span.SetAttributes(
		attribute.String("access_request.id", id.String()),
		attribute.String("approval_status", string(req.ApprovalStatus)),
		attribute.String("approver_id", approverID.String()),
	)

	logger.Info("Updating access request status", logger.Fields{
		"request_id":      id,
		"approval_status": req.ApprovalStatus,
		"approver_id":     approverID,
	})

	var expiresAt *time.Time
	var approvedAt *time.Time
	var approvedBy *uuid.UUID

	if req.ApprovalStatus == ApprovalStatusApproved {
		now := time.Now()
		approvedAt = &now
		approvedBy = &approverID

		if req.DurationHours != nil && *req.DurationHours > 0 {
			expiry := now.Add(time.Duration(*req.DurationHours) * time.Hour)
			expiresAt = &expiry
		}
	}

	// Convert fields to match SQLC types
	approvalStatusStr := string(req.ApprovalStatus)
	approvalComments := ""
	if req.ApprovalComments != nil {
		approvalComments = *req.ApprovalComments
	}

	var approvedAtSql sql.NullTime
	if approvedAt != nil {
		approvedAtSql = sql.NullTime{Time: *approvedAt, Valid: true}
	}

	var expiresAtSql sql.NullTime
	if expiresAt != nil {
		expiresAtSql = sql.NullTime{Time: *expiresAt, Valid: true}
	}

	params := &db.UpdateAccessRequestStatusParams{
		ID:               id,
		ApprovalStatus:   &approvalStatusStr,
		ApprovedBy:       approvedBy,
		ApprovedAt:       approvedAtSql,
		ApprovalComments: approvalComments,
		DurationHours:    req.DurationHours,
		ExpiresAt:        expiresAtSql,
	}

	sqlcRequest, err := r.db.UpdateAccessRequestStatus(ctx, *params)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to update access request status")
		r.metrics.IncrementCounter("access_request_update_error", map[string]any{"error": "db_error"})
		return nil, errors.NewRepositoryError("UPDATE_FAILED", "Failed to update access request status", err)
	}

	accessRequest, err := FromSQLCAccessRequest(sqlcRequest)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to convert access request")
		return nil, errors.NewRepositoryError("CONVERSION_FAILED", "Failed to convert access request", err)
	}

	r.metrics.IncrementCounter("access_request_status_updated", map[string]any{"status": string(req.ApprovalStatus)})
	logger.Info("Access request status updated successfully", logger.Fields{
		"request_id":      id,
		"approval_status": req.ApprovalStatus,
		"approver_id":     approverID,
	})

	return accessRequest, nil
}

// ListAccessRequests lists access requests with filters
func (r *accessRequestRepository) ListAccessRequests(ctx context.Context, req *ListAccessRequestsRequest) ([]*AccessRequest, error) {
	ctx, span := r.tracing.StartSpan(ctx, "accessRequestRepository.ListAccessRequests")
	defer span.End()

	var approvalStatus string
	if req.ApprovalStatus != nil {
		approvalStatus = string(*req.ApprovalStatus)
	}

	// Convert optional UUID pointers to required UUIDs (use zero value for nil)
	var requesterID uuid.UUID
	if req.RequesterID != nil {
		requesterID = *req.RequesterID
	}

	var targetUserID uuid.UUID
	if req.TargetUserID != nil {
		targetUserID = *req.TargetUserID
	}

	params := &db.ListAccessRequestsParams{
		Column1: approvalStatus, // approval_status
		Column2: requesterID,    // requester_id
		Column3: targetUserID,   // target_user_id
		Limit:   int32(req.Limit),
		Offset:  int32(req.Offset),
	}

	sqlcRequests, err := r.db.ListAccessRequests(ctx, *params)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to list access requests")
		r.metrics.IncrementCounter("access_request_list_error", map[string]any{"error": "db_error"})
		return nil, errors.NewRepositoryError("LIST_FAILED", "Failed to list access requests", err)
	}

	accessRequests := make([]*AccessRequest, len(sqlcRequests))
	for i, sqlcReq := range sqlcRequests {
		accessRequest, err := FromSQLCAccessRequest(sqlcReq)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "Failed to convert access request")
			return nil, errors.NewRepositoryError("CONVERSION_FAILED", "Failed to convert access request", err)
		}
		accessRequests[i] = accessRequest
	}

	r.metrics.ObserveHistogram("access_request_list_count", float64(len(accessRequests)), map[string]any{})
	return accessRequests, nil
}

// ListAccessRequestsWithDetails lists access requests with detailed information
func (r *accessRequestRepository) ListAccessRequestsWithDetails(ctx context.Context, req *ListAccessRequestsRequest) ([]*AccessRequestWithDetails, error) {
	ctx, span := r.tracing.StartSpan(ctx, "accessRequestRepository.ListAccessRequestsWithDetails")
	defer span.End()

	var approvalStatus string
	if req.ApprovalStatus != nil {
		approvalStatus = string(*req.ApprovalStatus)
	}

	var requestType string
	if req.RequestType != nil {
		requestType = string(*req.RequestType)
	}

	// Convert optional UUID pointers to required UUIDs (use zero value for nil)
	var requesterID uuid.UUID
	if req.RequesterID != nil {
		requesterID = *req.RequesterID
	}

	var targetUserID uuid.UUID
	if req.TargetUserID != nil {
		targetUserID = *req.TargetUserID
	}

	var entityID uuid.UUID
	if req.EntityID != nil {
		entityID = *req.EntityID
	}

	params := &db.GetAccessRequestsWithDetailsParams{
		Column1: approvalStatus,     // approval_status
		Column2: requesterID,        // requester_id
		Column3: targetUserID,       // target_user_id
		Column4: entityID,           // entity_id
		Column5: requestType,        // request_type
		Column6: req.IncludeExpired, // include_expired
		Limit:   int32(req.Limit),
		Offset:  int32(req.Offset),
	}

	sqlcRows, err := r.db.GetAccessRequestsWithDetails(ctx, *params)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get access requests with details")
		r.metrics.IncrementCounter("access_request_list_details_error", map[string]any{"error": "db_error"})
		return nil, errors.NewRepositoryError("LIST_FAILED", "Failed to get access requests with details", err)
	}

	details := make([]*AccessRequestWithDetails, len(sqlcRows))
	for i, row := range sqlcRows {
		// Convert the base access request
		baseRequest := &db.AccessRequest{
			ID:               row.ID,
			TenantID:         row.TenantID,
			RequesterID:      row.RequesterID,
			TargetUserID:     row.TargetUserID,
			EntityID:         row.EntityID,
			RequestType:      row.RequestType,
			RoleID:           row.RoleID,
			PermissionID:     row.PermissionID,
			ResourceID:       row.ResourceID,
			Justification:    row.Justification,
			BusinessReason:   row.BusinessReason,
			DurationHours:    row.DurationHours,
			ApprovalStatus:   row.ApprovalStatus,
			ApprovedBy:       row.ApprovedBy,
			ApprovedAt:       row.ApprovedAt,
			ApprovalComments: row.ApprovalComments,
			ExpiresAt:        row.ExpiresAt,
			AutoRevoke:       row.AutoRevoke,
			CreatedAt:        row.CreatedAt,
			UpdatedAt:        row.UpdatedAt,
		}

		accessRequest, err := FromSQLCAccessRequest(baseRequest)
		if err != nil {
			span.RecordError(err)
			return nil, errors.NewRepositoryError("CONVERSION_FAILED", "Failed to convert access request", err)
		}

		// Build the detailed view
		detail := &AccessRequestWithDetails{
			AccessRequest: accessRequest,
		}

		// Add requester details if available
		if row.RequesterUsername != nil {
			detail.RequesterDetails = &User{
				ID:       row.RequesterID,
				Username: *row.RequesterUsername,
				Email:    *row.RequesterEmail,
			}
		}

		// Add target user details if available
		if row.TargetUsername != nil && row.TargetUserID != nil {
			detail.TargetUserDetails = &User{
				ID:       *row.TargetUserID,
				Username: *row.TargetUsername,
				Email:    *row.TargetEmail,
			}
		}

		// Add role details if available
		if row.RoleName != nil && row.RoleID != nil {
			detail.RoleDetails = &Role{
				ID:          *row.RoleID,
				Name:        *row.RoleName,
				DisplayName: row.RoleDisplayName,
			}
		}

		// Add permission details if available
		if row.PermissionName != nil && row.PermissionID != nil {
			detail.PermissionDetails = &Permission{
				ID:          *row.PermissionID,
				Name:        *row.PermissionName,
				DisplayName: row.PermissionDisplayName,
			}
		}

		// Add resource details if available
		if row.ResourceName != nil && row.ResourceID != nil {
			detail.ResourceDetails = &db.Resource{
				ID:          *row.ResourceID,
				Name:        *row.ResourceName,
				DisplayName: row.ResourceDisplayName,
			}
		}

		// Add approver details if available
		if row.ApproverUsername != nil && row.ApprovedBy != nil {
			detail.ApproverDetails = &User{
				ID:       *row.ApprovedBy,
				Username: *row.ApproverUsername,
				Email:    *row.ApproverEmail,
			}
		}

		details[i] = detail
	}

	r.metrics.ObserveHistogram("access_request_list_details_count", float64(len(details)), map[string]any{})
	return details, nil
}

// GetPendingRequests gets all pending access requests
func (r *accessRequestRepository) GetPendingRequests(ctx context.Context, limit, offset int) ([]*AccessRequest, error) {
	ctx, span := r.tracing.StartSpan(ctx, "accessRequestRepository.GetPendingRequests")
	defer span.End()

	sqlcRequests, err := r.db.GetPendingAccessRequests(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get pending access requests")
		r.metrics.IncrementCounter("access_request_pending_error", map[string]any{"error": "db_error"})
		return nil, errors.NewRepositoryError("LIST_FAILED", "Failed to get pending access requests", err)
	}

	// Apply manual pagination if needed
	start := offset
	end := offset + limit
	if start > len(sqlcRequests) {
		return []*AccessRequest{}, nil
	}
	if end > len(sqlcRequests) {
		end = len(sqlcRequests)
	}

	paginatedRequests := sqlcRequests[start:end]
	accessRequests := make([]*AccessRequest, len(paginatedRequests))

	for i, sqlcReq := range paginatedRequests {
		accessRequest, err := FromSQLCAccessRequest(sqlcReq)
		if err != nil {
			span.RecordError(err)
			return nil, errors.NewRepositoryError("CONVERSION_FAILED", "Failed to convert access request", err)
		}
		accessRequests[i] = accessRequest
	}

	return accessRequests, nil
}

// GetPendingRequestsForApprover gets pending requests for approver view
func (r *accessRequestRepository) GetPendingRequestsForApprover(ctx context.Context, limit, offset int) ([]*AccessRequest, error) {
	ctx, span := r.tracing.StartSpan(ctx, "accessRequestRepository.GetPendingRequestsForApprover")
	defer span.End()

	params := &db.GetPendingRequestsForApproverParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	sqlcRows, err := r.db.GetPendingRequestsForApprover(ctx, *params)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get pending requests for approver")
		r.metrics.IncrementCounter("access_request_pending_approver_error", map[string]any{"error": "db_error"})
		return nil, errors.NewRepositoryError("LIST_FAILED", "Failed to get pending requests for approver", err)
	}

	accessRequests := make([]*AccessRequest, len(sqlcRows))
	for i, row := range sqlcRows {
		// Convert to base access request structure
		baseRequest := &db.AccessRequest{
			ID:               row.ID,
			TenantID:         row.TenantID,
			RequesterID:      row.RequesterID,
			TargetUserID:     row.TargetUserID,
			EntityID:         row.EntityID,
			RequestType:      row.RequestType,
			RoleID:           row.RoleID,
			PermissionID:     row.PermissionID,
			ResourceID:       row.ResourceID,
			Justification:    row.Justification,
			BusinessReason:   row.BusinessReason,
			DurationHours:    row.DurationHours,
			ApprovalStatus:   row.ApprovalStatus,
			ApprovedBy:       row.ApprovedBy,
			ApprovedAt:       row.ApprovedAt,
			ApprovalComments: row.ApprovalComments,
			ExpiresAt:        row.ExpiresAt,
			AutoRevoke:       row.AutoRevoke,
			CreatedAt:        row.CreatedAt,
			UpdatedAt:        row.UpdatedAt,
		}

		accessRequest, err := FromSQLCAccessRequest(baseRequest)
		if err != nil {
			span.RecordError(err)
			return nil, errors.NewRepositoryError("CONVERSION_FAILED", "Failed to convert access request", err)
		}
		accessRequests[i] = accessRequest
	}

	return accessRequests, nil
}

// GetUserAccessRequestHistory gets access request history for a user
func (r *accessRequestRepository) GetUserAccessRequestHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*AccessRequest, error) {
	ctx, span := r.tracing.StartSpan(ctx, "accessRequestRepository.GetUserAccessRequestHistory")
	defer span.End()

	span.SetAttributes(attribute.String("user_id", userID.String()))

	params := &db.GetUserAccessRequestHistoryParams{
		RequesterID: userID,
		Limit:       int32(limit),
		Offset:      int32(offset),
	}

	sqlcRows, err := r.db.GetUserAccessRequestHistory(ctx, *params)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get user access request history")
		r.metrics.IncrementCounter("access_request_history_error", map[string]any{"error": "db_error"})
		return nil, errors.NewRepositoryError("LIST_FAILED", "Failed to get user access request history", err)
	}

	accessRequests := make([]*AccessRequest, len(sqlcRows))
	for i, row := range sqlcRows {
		// Convert the base fields to AccessRequest
		baseRequest := &db.AccessRequest{
			ID:               row.ID,
			TenantID:         row.TenantID,
			RequesterID:      row.RequesterID,
			TargetUserID:     row.TargetUserID,
			EntityID:         row.EntityID,
			RequestType:      row.RequestType,
			RoleID:           row.RoleID,
			PermissionID:     row.PermissionID,
			ResourceID:       row.ResourceID,
			Justification:    row.Justification,
			BusinessReason:   row.BusinessReason,
			DurationHours:    row.DurationHours,
			ApprovalStatus:   row.ApprovalStatus,
			ApprovedBy:       row.ApprovedBy,
			ApprovedAt:       row.ApprovedAt,
			ApprovalComments: row.ApprovalComments,
			ExpiresAt:        row.ExpiresAt,
			AutoRevoke:       row.AutoRevoke,
			CreatedAt:        row.CreatedAt,
			UpdatedAt:        row.UpdatedAt,
		}

		accessRequest, err := FromSQLCAccessRequest(baseRequest)
		if err != nil {
			span.RecordError(err)
			return nil, errors.NewRepositoryError("CONVERSION_FAILED", "Failed to convert access request", err)
		}
		accessRequests[i] = accessRequest
	}

	return accessRequests, nil
}

// ApproveAccessRequest approves an access request
func (r *accessRequestRepository) ApproveAccessRequest(ctx context.Context, id, approverID uuid.UUID, comments *string) (*AccessRequest, error) {
	ctx, span := r.tracing.StartSpan(ctx, "accessRequestRepository.ApproveAccessRequest")
	defer span.End()

	params := &db.ApproveAccessRequestParams{
		ID:               id,
		ApprovedBy:       &approverID,
		ApprovalComments: *comments,
	}

	sqlcRequest, err := r.db.ApproveAccessRequest(ctx, *params)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to approve access request")
		r.metrics.IncrementCounter("access_request_approve_error", map[string]any{"error": "db_error"})
		return nil, errors.NewRepositoryError("APPROVE_FAILED", "Failed to approve access request", err)
	}

	accessRequest, err := FromSQLCAccessRequest(sqlcRequest)
	if err != nil {
		span.RecordError(err)
		return nil, errors.NewRepositoryError("CONVERSION_FAILED", "Failed to convert access request", err)
	}

	r.metrics.IncrementCounter("access_request_approved", map[string]any{})
	logger.Info("Access request approved", logger.Fields{
		"request_id":  id,
		"approver_id": approverID,
	})

	return accessRequest, nil
}

// RejectAccessRequest rejects an access request
func (r *accessRequestRepository) RejectAccessRequest(ctx context.Context, id, approverID uuid.UUID, comments *string) (*AccessRequest, error) {
	ctx, span := r.tracing.StartSpan(ctx, "accessRequestRepository.RejectAccessRequest")
	defer span.End()

	params := &db.RejectAccessRequestParams{
		ID:               id,
		ApprovedBy:       &approverID,
		ApprovalComments: *comments,
	}

	sqlcRequest, err := r.db.RejectAccessRequest(ctx, *params)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to reject access request")
		r.metrics.IncrementCounter("access_request_reject_error", map[string]any{"error": "db_error"})
		return nil, errors.NewRepositoryError("REJECT_FAILED", "Failed to reject access request", err)
	}

	accessRequest, err := FromSQLCAccessRequest(sqlcRequest)
	if err != nil {
		span.RecordError(err)
		return nil, errors.NewRepositoryError("CONVERSION_FAILED", "Failed to convert access request", err)
	}

	r.metrics.IncrementCounter("access_request_rejected", map[string]any{})
	logger.Info("Access request rejected", logger.Fields{
		"request_id":  id,
		"approver_id": approverID,
	})

	return accessRequest, nil
}

// RevokeAccessRequest revokes an approved access request
func (r *accessRequestRepository) RevokeAccessRequest(ctx context.Context, id uuid.UUID) (*AccessRequest, error) {
	ctx, span := r.tracing.StartSpan(ctx, "accessRequestRepository.RevokeAccessRequest")
	defer span.End()

	sqlcRequest, err := r.db.RevokeAccessRequest(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to revoke access request")
		r.metrics.IncrementCounter("access_request_revoke_error", map[string]any{"error": "db_error"})
		return nil, errors.NewRepositoryError("REVOKE_FAILED", "Failed to revoke access request", err)
	}

	accessRequest, err := FromSQLCAccessRequest(sqlcRequest)
	if err != nil {
		span.RecordError(err)
		return nil, errors.NewRepositoryError("CONVERSION_FAILED", "Failed to convert access request", err)
	}

	r.metrics.IncrementCounter("access_request_revoked", map[string]any{})
	logger.Info("Access request revoked", logger.Fields{"request_id": id})

	return accessRequest, nil
}

// ExpireAccessRequest expires an access request
func (r *accessRequestRepository) ExpireAccessRequest(ctx context.Context, id uuid.UUID) error {
	ctx, span := r.tracing.StartSpan(ctx, "accessRequestRepository.ExpireAccessRequest")
	defer span.End()

	err := r.db.ExpireAccessRequest(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to expire access request")
		r.metrics.IncrementCounter("access_request_expire_error", map[string]any{"error": "db_error"})
		return errors.NewRepositoryError("EXPIRE_FAILED", "Failed to expire access request", err)
	}

	r.metrics.IncrementCounter("access_request_expired", map[string]any{})
	logger.Info("Access request expired", logger.Fields{"request_id": id})

	return nil
}

// GetAccessRequestStats gets access request statistics
func (r *accessRequestRepository) GetAccessRequestStats(ctx context.Context, fromDate, toDate *time.Time) (*AccessRequestStats, error) {
	ctx, span := r.tracing.StartSpan(ctx, "accessRequestRepository.GetAccessRequestStats")
	defer span.End()

	params := &db.GetAccessRequestStatsParams{
		FromDate: *fromDate,
		ToDate:   *toDate,
	}

	sqlcStats, err := r.db.GetAccessRequestStats(ctx, *params)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get access request stats")
		r.metrics.IncrementCounter("access_request_stats_error", map[string]any{"error": "db_error"})
		return nil, errors.NewRepositoryError("STATS_FAILED", "Failed to get access request stats", err)
	}

	// Convert pgtype.Numeric to int32 for average approval time
	var avgApprovalTimeHours int32
	if sqlcStats.AvgApprovalTimeHours.Valid {
		avgApprovalTimeHours = int32(sqlcStats.AvgApprovalTimeHours.Int.Int64())
	}

	stats := &AccessRequestStats{
		TotalRequests:        sqlcStats.TotalRequests,
		PendingRequests:      sqlcStats.PendingRequests,
		ApprovedRequests:     sqlcStats.ApprovedRequests,
		RejectedRequests:     sqlcStats.RejectedRequests,
		ExpiredRequests:      sqlcStats.ExpiredRequests,
		AvgApprovalTimeHours: avgApprovalTimeHours,
		RequestsByType: map[RequestType]int64{
			RequestTypeRoleAssignment:  sqlcStats.RoleAssignmentRequests,
			RequestTypePermissionGrant: sqlcStats.PermissionGrantRequests,
			RequestTypeResourceAccess:  sqlcStats.ResourceAccessRequests,
			RequestTypeElevation:       sqlcStats.ElevationRequests,
		},
	}

	return stats, nil
}

// GetExpiredAccessRequests gets all expired access requests that should be auto-revoked
func (r *accessRequestRepository) GetExpiredAccessRequests(ctx context.Context) ([]*AccessRequest, error) {
	ctx, span := r.tracing.StartSpan(ctx, "accessRequestRepository.GetExpiredAccessRequests")
	defer span.End()

	sqlcRequests, err := r.db.GetExpiredAccessRequests(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get expired access requests")
		r.metrics.IncrementCounter("access_request_expired_get_error", map[string]any{"error": "db_error"})
		return nil, errors.NewRepositoryError("LIST_FAILED", "Failed to get expired access requests", err)
	}

	accessRequests := make([]*AccessRequest, len(sqlcRequests))
	for i, sqlcReq := range sqlcRequests {
		accessRequest, err := FromSQLCAccessRequest(sqlcReq)
		if err != nil {
			span.RecordError(err)
			return nil, errors.NewRepositoryError("CONVERSION_FAILED", "Failed to convert access request", err)
		}
		accessRequests[i] = accessRequest
	}

	r.metrics.ObserveHistogram("access_request_expired_count", float64(len(accessRequests)), map[string]any{})
	return accessRequests, nil
}
