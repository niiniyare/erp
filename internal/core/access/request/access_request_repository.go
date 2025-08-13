package request

//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"

import (
	"context"
	"database/sql"
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
	tracing tracing.TracingService
	metrics metrics.MetricsProvider
}

// NewAccessRequestRepository creates a new access request repository
func NewAccessRequestRepository(database db.Store, tracing tracing.TracingService, metrics metrics.MetricsProvider) AccessRequestRepository {
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
	tenantID, err := r.db.GetCurrentTenantID(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get tenant ID from context")
		r.metrics.IncrementCounter("access_request_create_error", map[string]any{"error": "no_tenant_id"})
		return nil, errors.NewRepositoryError("TENANT_REQUIRED", "Tenant ID is required", err)
	}

	params, err := req.ToSQLCCreateParams(tenantID, requesterID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to convert request params")
		r.metrics.IncrementCounter("access_request_create_error", map[string]any{"error": "param_conversion"})
		return nil, errors.NewRepositoryError("PARAM_CONVERSION_FAILED", "Failed to convert request parameters", err)
	}

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

// ListAccessRequests lists access requests with filters - simplified implementation
func (r *accessRequestRepository) ListAccessRequests(ctx context.Context, req *ListAccessRequestsRequest) ([]*AccessRequest, error) {
	ctx, span := r.tracing.StartSpan(ctx, "accessRequestRepository.ListAccessRequests")
	defer span.End()

	// TODO: Implement proper list functionality with raw SQL or add SQLC queries
	// For now, return empty list to allow compilation
	logger.Info("ListAccessRequests called - simplified implementation", logger.Fields{
		"limit":  req.Limit,
		"offset": req.Offset,
	})

	return []*AccessRequest{}, nil
}

// ListAccessRequestsWithDetails lists access requests with detailed information - simplified implementation
func (r *accessRequestRepository) ListAccessRequestsWithDetails(ctx context.Context, req *ListAccessRequestsRequest) ([]*AccessRequestWithDetails, error) {
	ctx, span := r.tracing.StartSpan(ctx, "accessRequestRepository.ListAccessRequestsWithDetails")
	defer span.End()

	// TODO: Implement proper detailed list functionality with raw SQL or add SQLC queries
	// For now, return empty list to allow compilation
	logger.Info("ListAccessRequestsWithDetails called - simplified implementation", logger.Fields{
		"limit":  req.Limit,
		"offset": req.Offset,
	})

	return []*AccessRequestWithDetails{}, nil
}

// GetPendingRequests gets all pending access requests - simplified implementation
func (r *accessRequestRepository) GetPendingRequests(ctx context.Context, limit, offset int) ([]*AccessRequest, error) {
	ctx, span := r.tracing.StartSpan(ctx, "accessRequestRepository.GetPendingRequests")
	defer span.End()

	// TODO: Implement proper pending requests query with raw SQL or add SQLC queries
	// For now, return empty list to allow compilation
	logger.Info("GetPendingRequests called - simplified implementation", logger.Fields{
		"limit":  limit,
		"offset": offset,
	})

	return []*AccessRequest{}, nil
}

// GetPendingRequestsForApprover gets pending requests for approver view - simplified implementation
func (r *accessRequestRepository) GetPendingRequestsForApprover(ctx context.Context, limit, offset int) ([]*AccessRequest, error) {
	ctx, span := r.tracing.StartSpan(ctx, "accessRequestRepository.GetPendingRequestsForApprover")
	defer span.End()

	// TODO: Implement proper pending requests for approver query with raw SQL or add SQLC queries
	// For now, return empty list to allow compilation
	logger.Info("GetPendingRequestsForApprover called - simplified implementation", logger.Fields{
		"limit":  limit,
		"offset": offset,
	})

	return []*AccessRequest{}, nil
}

// GetUserAccessRequestHistory gets access request history for a user - simplified implementation
func (r *accessRequestRepository) GetUserAccessRequestHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*AccessRequest, error) {
	ctx, span := r.tracing.StartSpan(ctx, "accessRequestRepository.GetUserAccessRequestHistory")
	defer span.End()

	span.SetAttributes(attribute.String("user_id", userID.String()))

	// TODO: Implement proper user history query with raw SQL or add SQLC queries
	// For now, return empty list to allow compilation
	logger.Info("GetUserAccessRequestHistory called - simplified implementation", logger.Fields{
		"user_id": userID,
		"limit":   limit,
		"offset":  offset,
	})

	return []*AccessRequest{}, nil
}

// ApproveAccessRequest approves an access request - using UpdateAccessRequestStatus
func (r *accessRequestRepository) ApproveAccessRequest(ctx context.Context, id, approverID uuid.UUID, comments *string) (*AccessRequest, error) {
	ctx, span := r.tracing.StartSpan(ctx, "accessRequestRepository.ApproveAccessRequest")
	defer span.End()

	// Use the existing UpdateAccessRequestStatus with approval data
	approvedAt := sql.NullTime{Time: time.Now(), Valid: true}
	commentsStr := ""
	if comments != nil {
		commentsStr = *comments
	}

	params := &db.UpdateAccessRequestStatusParams{
		ID:               id,
		ApprovalStatus:   func() *string { s := "APPROVED"; return &s }(),
		ApprovedBy:       &approverID,
		ApprovedAt:       approvedAt,
		ApprovalComments: commentsStr,
		DurationHours:    nil,            // Keep existing duration
		ExpiresAt:        sql.NullTime{}, // Keep existing expiry
	}

	sqlcRequest, err := r.db.UpdateAccessRequestStatus(ctx, *params)
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

// RejectAccessRequest rejects an access request - using UpdateAccessRequestStatus
func (r *accessRequestRepository) RejectAccessRequest(ctx context.Context, id, approverID uuid.UUID, comments *string) (*AccessRequest, error) {
	ctx, span := r.tracing.StartSpan(ctx, "accessRequestRepository.RejectAccessRequest")
	defer span.End()

	// Use the existing UpdateAccessRequestStatus with rejection data
	approvedAt := sql.NullTime{Time: time.Now(), Valid: true}
	commentsStr := ""
	if comments != nil {
		commentsStr = *comments
	}

	params := &db.UpdateAccessRequestStatusParams{
		ID:               id,
		ApprovalStatus:   func() *string { s := "REJECTED"; return &s }(),
		ApprovedBy:       &approverID,
		ApprovedAt:       approvedAt,
		ApprovalComments: commentsStr,
		DurationHours:    nil,            // Keep existing duration
		ExpiresAt:        sql.NullTime{}, // Keep existing expiry
	}

	sqlcRequest, err := r.db.UpdateAccessRequestStatus(ctx, *params)
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

// RevokeAccessRequest revokes an approved access request - using UpdateAccessRequestStatus
func (r *accessRequestRepository) RevokeAccessRequest(ctx context.Context, id uuid.UUID) (*AccessRequest, error) {
	ctx, span := r.tracing.StartSpan(ctx, "accessRequestRepository.RevokeAccessRequest")
	defer span.End()

	// Use the existing UpdateAccessRequestStatus with revocation data
	approvedAt := sql.NullTime{Time: time.Now(), Valid: true}

	params := &db.UpdateAccessRequestStatusParams{
		ID:               id,
		ApprovalStatus:   func() *string { s := "REVOKED"; return &s }(),
		ApprovedBy:       nil, // No specific approver for revocation
		ApprovedAt:       approvedAt,
		ApprovalComments: "Access request revoked",
		DurationHours:    nil,            // Keep existing duration
		ExpiresAt:        sql.NullTime{}, // Keep existing expiry
	}

	sqlcRequest, err := r.db.UpdateAccessRequestStatus(ctx, *params)
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

// ExpireAccessRequest expires an access request - using UpdateAccessRequestStatus
func (r *accessRequestRepository) ExpireAccessRequest(ctx context.Context, id uuid.UUID) error {
	ctx, span := r.tracing.StartSpan(ctx, "accessRequestRepository.ExpireAccessRequest")
	defer span.End()

	// Use the existing UpdateAccessRequestStatus with expiration data
	approvedAt := sql.NullTime{Time: time.Now(), Valid: true}

	params := &db.UpdateAccessRequestStatusParams{
		ID:               id,
		ApprovalStatus:   func() *string { s := "EXPIRED"; return &s }(),
		ApprovedBy:       nil, // No specific approver for expiration
		ApprovedAt:       approvedAt,
		ApprovalComments: "Access request expired",
		DurationHours:    nil,            // Keep existing duration
		ExpiresAt:        sql.NullTime{}, // Keep existing expiry
	}

	_, err := r.db.UpdateAccessRequestStatus(ctx, *params)
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

// GetAccessRequestStats gets access request statistics - simplified implementation
func (r *accessRequestRepository) GetAccessRequestStats(ctx context.Context, fromDate, toDate *time.Time) (*AccessRequestStats, error) {
	ctx, span := r.tracing.StartSpan(ctx, "accessRequestRepository.GetAccessRequestStats")
	defer span.End()

	// TODO: Implement proper stats query with raw SQL or add SQLC queries
	// For now, return empty stats to allow compilation
	logger.Info("GetAccessRequestStats called - simplified implementation", logger.Fields{
		"from_date": fromDate,
		"to_date":   toDate,
	})

	stats := &AccessRequestStats{
		TotalRequests:        0,
		PendingRequests:      0,
		ApprovedRequests:     0,
		RejectedRequests:     0,
		ExpiredRequests:      0,
		AvgApprovalTimeHours: 0,
		RequestsByType: map[RequestType]int64{
			RequestTypeRoleAssignment:  0,
			RequestTypePermissionGrant: 0,
			RequestTypeResourceAccess:  0,
			RequestTypeElevation:       0,
		},
	}

	return stats, nil
}

// GetExpiredAccessRequests gets all expired access requests that should be auto-revoked - simplified implementation
func (r *accessRequestRepository) GetExpiredAccessRequests(ctx context.Context) ([]*AccessRequest, error) {
	ctx, span := r.tracing.StartSpan(ctx, "accessRequestRepository.GetExpiredAccessRequests")
	defer span.End()

	// TODO: Implement proper expired requests query with raw SQL or add SQLC queries
	// For now, return empty list to allow compilation
	logger.Info("GetExpiredAccessRequests called - simplified implementation")

	return []*AccessRequest{}, nil
}
