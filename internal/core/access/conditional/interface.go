package conditional

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// AccessRequest represents an access request in the system.
// This is a local representation to avoid circular dependencies.
type AccessRequest struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	RequestType    string
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

// PermissionEvaluationAudit represents permission evaluation audit data
// This is a local representation to avoid circular dependencies.
type PermissionEvaluationAudit struct {
	UserID           uuid.UUID
	ResourceName     string
	ActionName       string
	Decision         string
	PolicyDecisions  []string
	EvaluationTimeMS int
	Context          map[string]any
	RiskFactors      []string
}

// AuditService defines the interface for audit logging needed by the conditional access service.
// This is a local representation to avoid circular dependencies.
type AuditService interface {
	LogPermissionEvaluation(ctx context.Context, evaluation *PermissionEvaluationAudit) error
}
