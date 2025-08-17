package adapters

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/core/access/conditional"
	"github.com/niiniyare/erp/internal/core/access/request"
	"github.com/niiniyare/erp/internal/core/audit"
	"github.com/niiniyare/erp/internal/core/identity"
	"github.com/niiniyare/erp/internal/shared"
)

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}

// UserServiceAdapter adapts the identity.Service to request.UserService interface
type UserServiceAdapter struct {
	identityService identity.Service
}

// NewUserServiceAdapter creates a new adapter
func NewUserServiceAdapter(identityService identity.Service) request.UserService {
	return &UserServiceAdapter{
		identityService: identityService,
	}
}

// GetUserByID converts identity.User to request.User
func (a *UserServiceAdapter) GetUserByID(ctx context.Context, id uuid.UUID) (*request.User, error) {
	identityUser, err := a.identityService.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Convert identity.User to request.User
	return &request.User{
		ID:       identityUser.ID,
		Username: identityUser.Username,
		Email:    identityUser.Email,
	}, nil
}

// AuditServiceAdapter adapts the audit.Service to conditional.AuditService interface
type AuditServiceAdapter struct {
	auditService audit.Service
}

// NewAuditServiceAdapter creates a new adapter
func NewAuditServiceAdapter(auditService audit.Service) conditional.AuditService {
	return &AuditServiceAdapter{
		auditService: auditService,
	}
}

// LogPermissionEvaluation converts PermissionEvaluationAudit to audit.AuditEvent
func (a *AuditServiceAdapter) LogPermissionEvaluation(ctx context.Context, evaluation *conditional.PermissionEvaluationAudit) error {
	contextData, _ := json.Marshal(map[string]any{
		"resource_name":      evaluation.ResourceName,
		"action_name":        evaluation.ActionName,
		"policy_decisions":   evaluation.PolicyDecisions,
		"evaluation_time_ms": evaluation.EvaluationTimeMS,
		"context":            evaluation.Context,
		"risk_factors":       evaluation.RiskFactors,
	})

	_, _ = shared.GetTenantID(ctx) // Keep for consistency but no longer needed
	req := audit.CreateAuditEventRequest{
		UserID:        &evaluation.UserID,
		EventType:     "permission_evaluation",
		EventCategory: "access_control",
		Severity:      "info",
		Decision:      &evaluation.Decision,
		Reason:        stringPtr("Permission evaluation completed"),
		Context:       contextData,
	}
	_, err := a.auditService.CreateAuditEvent(ctx, req)
	return err
}
