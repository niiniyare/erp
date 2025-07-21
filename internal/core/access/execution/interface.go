package execution

import (
	"time"

	"github.com/google/uuid"
)

// RequestType represents the type of access request.
// This is a local representation to avoid circular dependencies.
type RequestType string

const (
	RequestTypeRoleAssignment  RequestType = "ROLE_ASSIGNMENT"
	RequestTypePermissionGrant RequestType = "PERMISSION_GRANT"
	RequestTypeResourceAccess  RequestType = "RESOURCE_ACCESS"
	RequestTypeElevation       RequestType = "ELEVATION"
)

// AccessRequest represents an access request in the system.
// This is a local representation to avoid circular dependencies.
type AccessRequest struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	RequestType    RequestType
	RequesterID    uuid.UUID
	TargetUserID   *uuid.UUID
	EntityID       uuid.UUID
	RoleID         *uuid.UUID
	PermissionID   *uuid.UUID
	ResourceID     *uuid.UUID
	ApprovalStatus string
	ApprovedBy     *uuid.UUID
	ExpiresAt      *time.Time
	Justification  string
	BusinessReason string
}

// ApprovalStatus represents the status of an access request approval.
// This is a local representation to avoid circular dependencies.
const (
	ApprovalStatusApproved = "APPROVED"
)
