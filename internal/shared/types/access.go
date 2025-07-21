package types

import (
	"time"

	"github.com/google/uuid"
)

// RequestType represents the type of access request
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

// ApprovalStatus represents the status of an approval request
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

// AccessRequest represents an access request (domain model)
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
