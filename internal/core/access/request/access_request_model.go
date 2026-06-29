//go:build ignore

package request

import (
	"database/sql"
	"time"

	db "awo.so/db/sqlc"
	"awo.so/internal/shared/errors"
	"awo.so/internal/shared/types"
	"github.com/google/uuid"
)

// Type aliases for shared types
type (
	RequestType    = types.RequestType
	ApprovalStatus = types.ApprovalStatus
	AccessRequest  = types.AccessRequest
)

const (
	RequestTypeRoleAssignment  = types.RequestTypeRoleAssignment
	RequestTypePermissionGrant = types.RequestTypePermissionGrant
	RequestTypeResourceAccess  = types.RequestTypeResourceAccess
	RequestTypeElevation       = types.RequestTypeElevation
)

const (
	ApprovalStatusPending  = types.ApprovalStatusPending
	ApprovalStatusApproved = types.ApprovalStatusApproved
	ApprovalStatusRejected = types.ApprovalStatusRejected
	ApprovalStatusExpired  = types.ApprovalStatusExpired
	ApprovalStatusRevoked  = types.ApprovalStatusRevoked
)

// User represents a user in the system for details view.
type User struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
}

// Role represents a role in the system for details view.
type Role struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	DisplayName *string   `json:"display_name,omitempty"`
}

// Permission represents a permission in the system for details view.
type Permission struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	DisplayName *string   `json:"display_name,omitempty"`
}

// AccessRequestWithDetails includes related entity details for view
type AccessRequestWithDetails struct {
	*AccessRequest
	RequesterDetails  *User        `json:"requester_details,omitempty"`
	TargetUserDetails *User        `json:"target_user_details,omitempty"`
	RoleDetails       *Role        `json:"role_details,omitempty"`
	PermissionDetails *Permission  `json:"permission_details,omitempty"`
	ResourceDetails   *db.Resource `json:"resource_details,omitempty"`
	ApproverDetails   *User        `json:"approver_details,omitempty"`
}

// CreateAccessRequestRequest represents access request creation request
type CreateAccessRequestRequest struct {
	TargetUserID   *uuid.UUID  `json:"target_user_id,omitempty"`
	EntityID       uuid.UUID   `json:"entity_id" validate:"required"`
	RequestType    RequestType `json:"request_type" validate:"required"`
	RoleID         *uuid.UUID  `json:"role_id,omitempty"`
	PermissionID   *uuid.UUID  `json:"permission_id,omitempty"`
	ResourceID     *uuid.UUID  `json:"resource_id,omitempty"`
	Justification  string      `json:"justification" validate:"required,min=10,max=2000"`
	BusinessReason *string     `json:"business_reason,omitempty"`
	DurationHours  *int32      `json:"duration_hours,omitempty"`
}

// UpdateAccessRequestRequest represents access request update request (for approvers)
type UpdateAccessRequestRequest struct {
	ApprovalStatus   ApprovalStatus `json:"approval_status" validate:"required"`
	ApprovalComments *string        `json:"approval_comments,omitempty"`
	DurationHours    *int32         `json:"duration_hours,omitempty"` // Approver can modify duration
}

// ListAccessRequestsRequest represents access request list request with filters
type ListAccessRequestsRequest struct {
	RequesterID    *uuid.UUID      `json:"requester_id,omitempty"`
	TargetUserID   *uuid.UUID      `json:"target_user_id,omitempty"`
	EntityID       *uuid.UUID      `json:"entity_id,omitempty"`
	RequestType    *RequestType    `json:"request_type,omitempty"`
	ApprovalStatus *ApprovalStatus `json:"approval_status,omitempty"`
	IncludeExpired bool            `json:"include_expired"`
	Limit          int             `json:"limit"`
	Offset         int             `json:"offset"`
}

// AccessRequestApprovalRequest represents approval/rejection action
type AccessRequestApprovalRequest struct {
	Action           string  `json:"action" validate:"required,oneof=approve reject"`
	Comments         *string `json:"comments,omitempty"`
	ModifyDuration   *int32  `json:"modify_duration,omitempty"`
	ConditionalTerms *string `json:"conditional_terms,omitempty"`
}

// AccessRequestWorkflowEvent represents workflow state changes for audit trail
type AccessRequestWorkflowEvent struct {
	ID            uuid.UUID       `json:"id"`
	RequestID     uuid.UUID       `json:"request_id"`
	EventType     string          `json:"event_type"` // CREATED, SUBMITTED, APPROVED, REJECTED, EXPIRED, REVOKED
	ActorID       *uuid.UUID      `json:"actor_id,omitempty"`
	PreviousState *ApprovalStatus `json:"previous_state,omitempty"`
	NewState      ApprovalStatus  `json:"new_state"`
	Comments      *string         `json:"comments,omitempty"`
	Metadata      map[string]any  `json:"metadata,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
}

// AccessRequestStats represents access request statistics for monitoring
type AccessRequestStats struct {
	TotalRequests        int64                 `json:"total_requests"`
	PendingRequests      int64                 `json:"pending_requests"`
	ApprovedRequests     int64                 `json:"approved_requests"`
	RejectedRequests     int64                 `json:"rejected_requests"`
	ExpiredRequests      int64                 `json:"expired_requests"`
	AvgApprovalTimeHours int32                 `json:"avg_approval_time_hours"`
	RequestsByType       map[RequestType]int64 `json:"requests_by_type"`
}

// Conversion functions following existing patterns

// FromSQLCAccessRequest converts SQLC AccessRequest to domain AccessRequest
func FromSQLCAccessRequest(sqlcRequest *db.AccessRequest) (*AccessRequest, error) {
	var approvalStatus ApprovalStatus
	if sqlcRequest.ApprovalStatus != nil {
		approvalStatus = ApprovalStatus(*sqlcRequest.ApprovalStatus)
	}

	var approvedAt *time.Time
	if sqlcRequest.ApprovedAt.Valid {
		approvedAt = &sqlcRequest.ApprovedAt.Time
	}

	var approvalComments *string
	if sqlcRequest.ApprovalComments != nil && *sqlcRequest.ApprovalComments != "" {
		approvalComments = sqlcRequest.ApprovalComments
	}

	var expiresAt *time.Time
	if sqlcRequest.ExpiresAt.Valid {
		expiresAt = &sqlcRequest.ExpiresAt.Time
	}

	var autoRevoke bool
	if sqlcRequest.AutoRevoke != nil {
		autoRevoke = *sqlcRequest.AutoRevoke
	}

	var createdAt time.Time
	if sqlcRequest.CreatedAt.Valid {
		createdAt = sqlcRequest.CreatedAt.Time
	}

	var updatedAt time.Time
	if sqlcRequest.UpdatedAt.Valid {
		updatedAt = sqlcRequest.UpdatedAt.Time
	}

	return &AccessRequest{
		ID:               sqlcRequest.ID,
		TenantID:         sqlcRequest.TenantID,
		RequesterID:      sqlcRequest.RequesterID,
		TargetUserID:     sqlcRequest.TargetUserID,
		EntityID:         sqlcRequest.EntityID,
		RequestType:      RequestType(sqlcRequest.RequestType),
		RoleID:           sqlcRequest.RoleID,
		PermissionID:     sqlcRequest.PermissionID,
		ResourceID:       sqlcRequest.ResourceID,
		Justification:    sqlcRequest.Justification,
		BusinessReason:   sqlcRequest.BusinessReason,
		DurationHours:    sqlcRequest.DurationHours,
		ApprovalStatus:   approvalStatus,
		ApprovedBy:       sqlcRequest.ApprovedBy,
		ApprovedAt:       approvedAt,
		ApprovalComments: approvalComments,
		ExpiresAt:        expiresAt,
		AutoRevoke:       autoRevoke,
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
	}, nil
}

// ToSQLCCreateParams converts CreateAccessRequestRequest to SQLC params
func (req *CreateAccessRequestRequest) ToSQLCCreateParams(tenantID, requesterID uuid.UUID) (*db.CreateAccessRequestParams, error) {
	var expiresAt *time.Time
	if req.DurationHours != nil && *req.DurationHours > 0 {
		expiry := time.Now().Add(time.Duration(*req.DurationHours) * time.Hour)
		expiresAt = &expiry
	}

	var expiresAtSql sql.NullTime
	if expiresAt != nil {
		expiresAtSql = sql.NullTime{Time: *expiresAt, Valid: true}
	}

	autoRevoke := true // Default to true for security

	return &db.CreateAccessRequestParams{
		RequesterID:    requesterID,
		TargetUserID:   req.TargetUserID,
		EntityID:       req.EntityID,
		RequestType:    string(req.RequestType),
		RoleID:         req.RoleID,
		PermissionID:   req.PermissionID,
		ResourceID:     req.ResourceID,
		Justification:  req.Justification,
		BusinessReason: req.BusinessReason,
		DurationHours:  req.DurationHours,
		ExpiresAt:      expiresAtSql,
		AutoRevoke:     &autoRevoke,
	}, nil
}

// ToSQLCUpdateParams converts UpdateAccessRequestRequest to SQLC params
func (req *UpdateAccessRequestRequest) ToSQLCUpdateParams(requestID, approverID uuid.UUID) (*db.UpdateAccessRequestStatusParams, error) {
	var expiresAt *time.Time
	if req.DurationHours != nil && *req.DurationHours > 0 && req.ApprovalStatus == ApprovalStatusApproved {
		expiry := time.Now().Add(time.Duration(*req.DurationHours) * time.Hour)
		expiresAt = &expiry
	}

	var approvedAt *time.Time
	var approvedBy *uuid.UUID
	if req.ApprovalStatus == ApprovalStatusApproved {
		now := time.Now()
		approvedAt = &now
		approvedBy = &approverID
	}

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

	return &db.UpdateAccessRequestStatusParams{
		ID:               requestID,
		ApprovalStatus:   &approvalStatusStr,
		ApprovedBy:       approvedBy,
		ApprovedAt:       approvedAtSql,
		ApprovalComments: &approvalComments,
		DurationHours:    req.DurationHours,
		ExpiresAt:        expiresAtSql,
	}, nil
}

// ValidateRequest validates the access request based on its type
func (req *CreateAccessRequestRequest) ValidateRequest() error {
	if !req.RequestType.IsValid() {
		return errors.NewBusinessError("INVALID_REQUEST_TYPE", "Invalid request type").WithDetail("request_type", req.RequestType)
	}

	switch req.RequestType {
	case RequestTypeRoleAssignment:
		if req.RoleID == nil {
			return errors.NewBusinessError("ROLE_ID_REQUIRED", "Role ID is required for role assignment requests")
		}
	case RequestTypePermissionGrant:
		if req.PermissionID == nil {
			return errors.NewBusinessError("PERMISSION_ID_REQUIRED", "Permission ID is required for permission grant requests")
		}
	case RequestTypeResourceAccess:
		if req.ResourceID == nil {
			return errors.NewBusinessError("RESOURCE_ID_REQUIRED", "Resource ID is required for resource access requests")
		}
	case RequestTypeElevation:
		if req.DurationHours == nil || *req.DurationHours <= 0 {
			return errors.NewBusinessError("DURATION_REQUIRED", "Duration is required for elevation requests")
		}
		if *req.DurationHours > 72 { // Max 72 hours for elevation
			return errors.NewBusinessError("DURATION_EXCEEDED", "Elevation duration cannot exceed 72 hours").WithDetail("duration_hours", *req.DurationHours)
		}
	}

	return nil
}

// GetRequestTypeDisplayName returns human-readable request type name
func GetRequestTypeDisplayName(rt RequestType) string {
	switch rt {
	case RequestTypeRoleAssignment:
		return "Role Assignment"
	case RequestTypePermissionGrant:
		return "Permission Grant"
	case RequestTypeResourceAccess:
		return "Resource Access"
	case RequestTypeElevation:
		return "Privilege Elevation"
	default:
		return string(rt)
	}
}

// GetApprovalStatusDisplayName returns human-readable approval status name
func GetApprovalStatusDisplayName(as ApprovalStatus) string {
	switch as {
	case ApprovalStatusPending:
		return "Pending Approval"
	case ApprovalStatusApproved:
		return "Approved"
	case ApprovalStatusRejected:
		return "Rejected"
	case ApprovalStatusExpired:
		return "Expired"
	case ApprovalStatusRevoked:
		return "Revoked"
	default:
		return string(as)
	}
}
