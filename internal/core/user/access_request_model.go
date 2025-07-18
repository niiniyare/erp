package user

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	db "github.com/niiniyare/erp/db/sqlc"
)

// ValidationError represents a validation error with field and message
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// Access request workflow types and constants following existing patterns
type RequestType string

const (
	RequestTypeRoleAssignment  RequestType = "ROLE_ASSIGNMENT"
	RequestTypePermissionGrant RequestType = "PERMISSION_GRANT"
	RequestTypeResourceAccess  RequestType = "RESOURCE_ACCESS"
	RequestTypeElevation       RequestType = "ELEVATION"
)

var validRequestTypes = map[RequestType]struct{}{
	RequestTypeRoleAssignment:  {},
	RequestTypePermissionGrant: {},
	RequestTypeResourceAccess:  {},
	RequestTypeElevation:       {},
}

func (rt RequestType) IsValid() bool {
	_, ok := validRequestTypes[rt]
	return ok
}

func (rt RequestType) String() string {
	return string(rt)
}

func AllRequestTypes() []RequestType {
	return []RequestType{
		RequestTypeRoleAssignment,
		RequestTypePermissionGrant,
		RequestTypeResourceAccess,
		RequestTypeElevation,
	}
}

type ApprovalStatus string

const (
	ApprovalStatusPending  ApprovalStatus = "PENDING"
	ApprovalStatusApproved ApprovalStatus = "APPROVED"
	ApprovalStatusRejected ApprovalStatus = "REJECTED"
	ApprovalStatusExpired  ApprovalStatus = "EXPIRED"
	ApprovalStatusRevoked  ApprovalStatus = "REVOKED"
)

var validApprovalStatuses = map[ApprovalStatus]struct{}{
	ApprovalStatusPending:  {},
	ApprovalStatusApproved: {},
	ApprovalStatusRejected: {},
	ApprovalStatusExpired:  {},
	ApprovalStatusRevoked:  {},
}

func (as ApprovalStatus) IsValid() bool {
	_, ok := validApprovalStatuses[as]
	return ok
}

func (as ApprovalStatus) String() string {
	return string(as)
}

func AllApprovalStatuses() []ApprovalStatus {
	return []ApprovalStatus{
		ApprovalStatusPending,
		ApprovalStatusApproved,
		ApprovalStatusRejected,
		ApprovalStatusExpired,
		ApprovalStatusRevoked,
	}
}

// AccessRequest represents an access request in the system (domain model)
type AccessRequest struct {
	ID               uuid.UUID      `json:"id"`
	TenantID         uuid.UUID      `json:"tenant_id"`
	RequesterID      uuid.UUID      `json:"requester_id"`
	TargetUserID     *uuid.UUID     `json:"target_user_id,omitempty"`
	EntityID         uuid.UUID      `json:"entity_id"`
	RequestType      RequestType    `json:"request_type"`
	RoleID           *uuid.UUID     `json:"role_id,omitempty"`
	PermissionID     *uuid.UUID     `json:"permission_id,omitempty"`
	ResourceID       *uuid.UUID     `json:"resource_id,omitempty"`
	Justification    string         `json:"justification"`
	BusinessReason   *string        `json:"business_reason,omitempty"`
	DurationHours    *int32         `json:"duration_hours,omitempty"`
	ApprovalStatus   ApprovalStatus `json:"approval_status"`
	ApprovedBy       *uuid.UUID     `json:"approved_by,omitempty"`
	ApprovedAt       *time.Time     `json:"approved_at,omitempty"`
	ApprovalComments *string        `json:"approval_comments,omitempty"`
	ExpiresAt        *time.Time     `json:"expires_at,omitempty"`
	AutoRevoke       bool           `json:"auto_revoke"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

// AccessRequestWithDetails includes related entity details for comprehensive view
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
	ID            uuid.UUID              `json:"id"`
	RequestID     uuid.UUID              `json:"request_id"`
	EventType     string                 `json:"event_type"` // CREATED, SUBMITTED, APPROVED, REJECTED, EXPIRED, REVOKED
	ActorID       *uuid.UUID             `json:"actor_id,omitempty"`
	PreviousState *ApprovalStatus        `json:"previous_state,omitempty"`
	NewState      ApprovalStatus         `json:"new_state"`
	Comments      *string                `json:"comments,omitempty"`
	Metadata      map[string]any `json:"metadata,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
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
	if sqlcRequest.ApprovalComments != "" {
		approvalComments = &sqlcRequest.ApprovalComments
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

// ToSQLCAccessRequestParams converts CreateAccessRequestRequest to SQLC params
func (req *CreateAccessRequestRequest) ToSQLCCreateParams(tenantID, requesterID uuid.UUID) *db.CreateAccessRequestParams {
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
	}
}

// ToSQLCUpdateParams converts UpdateAccessRequestRequest to SQLC params
func (req *UpdateAccessRequestRequest) ToSQLCUpdateParams(requestID, approverID uuid.UUID) *db.UpdateAccessRequestStatusParams {
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
		ApprovalComments: approvalComments,
		DurationHours:    req.DurationHours,
		ExpiresAt:        expiresAtSql,
	}
}

// ValidateRequest validates the access request based on its type
func (req *CreateAccessRequestRequest) ValidateRequest() error {
	if !req.RequestType.IsValid() {
		return NewValidationError("request_type", "Invalid request type")
	}

	switch req.RequestType {
	case RequestTypeRoleAssignment:
		if req.RoleID == nil {
			return NewValidationError("role_id", "Role ID is required for role assignment requests")
		}
	case RequestTypePermissionGrant:
		if req.PermissionID == nil {
			return NewValidationError("permission_id", "Permission ID is required for permission grant requests")
		}
	case RequestTypeResourceAccess:
		if req.ResourceID == nil {
			return NewValidationError("resource_id", "Resource ID is required for resource access requests")
		}
	case RequestTypeElevation:
		if req.DurationHours == nil || *req.DurationHours <= 0 {
			return NewValidationError("duration_hours", "Duration is required for elevation requests")
		}
		if *req.DurationHours > 72 { // Max 72 hours for elevation
			return NewValidationError("duration_hours", "Elevation duration cannot exceed 72 hours")
		}
	}

	return nil
}

// IsExpired checks if the access request has expired
func (ar *AccessRequest) IsExpired() bool {
	return ar.ExpiresAt != nil && ar.ExpiresAt.Before(time.Now())
}

// CanBeApproved checks if the request is in a state that allows approval
func (ar *AccessRequest) CanBeApproved() bool {
	return ar.ApprovalStatus == ApprovalStatusPending && !ar.IsExpired()
}

// CanBeRevoked checks if the request can be revoked
func (ar *AccessRequest) CanBeRevoked() bool {
	return ar.ApprovalStatus == ApprovalStatusApproved && !ar.IsExpired()
}

// GetRequestTypeDisplayName returns human-readable request type name
func (rt RequestType) GetDisplayName() string {
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
func (as ApprovalStatus) GetDisplayName() string {
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

