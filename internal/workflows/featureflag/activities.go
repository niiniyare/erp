package featureflag

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/core/access/request"
	"github.com/niiniyare/erp/internal/core/featureflag"
	"github.com/niiniyare/erp/internal/featureflag/workflow"
	"github.com/niiniyare/erp/internal/shared/logger"
	"go.temporal.io/sdk/activity"
)

// FeatureFlagActivities contains all feature flag workflow activities
type FeatureFlagActivities struct {
	featureFlagService featureflag.Service
	accessRequestRepo  request.AccessRequestRepository
	webSocketService   WebSocketService
	auditService       AuditService
	policyService      PolicyService
}

// WebSocketService interface for real-time notifications
type WebSocketService interface {
	BroadcastToTenant(ctx context.Context, tenantID uuid.UUID, message *workflow.WebSocketMessage) error
	NotifyFlagChange(ctx context.Context, tenantID uuid.UUID, event *workflow.FeatureFlagChangeEvent) error
}

// AuditService interface for audit logging
type AuditService interface {
	CreateAuditEvent(ctx context.Context, req *workflow.AuditEventRequest) error
}

// PolicyService interface for approval policies
type PolicyService interface {
	GetApprovalPolicy(ctx context.Context, tenantID uuid.UUID, flagName string, changeType string) (*workflow.FeatureFlagApprovalPolicy, error)
	CheckApprovalRequired(ctx context.Context, tenantID uuid.UUID, flagName string, changeType string) (bool, error)
}

// NewFeatureFlagActivities creates a new instance of feature flag activities
func NewFeatureFlagActivities(
	featureFlagService featureflag.Service,
	accessRequestRepo request.AccessRequestRepository,
	webSocketService WebSocketService,
	auditService AuditService,
	policyService PolicyService,
) *FeatureFlagActivities {
	return &FeatureFlagActivities{
		featureFlagService: featureFlagService,
		accessRequestRepo:  accessRequestRepo,
		webSocketService:   webSocketService,
		auditService:       auditService,
		policyService:      policyService,
	}
}

// ValidateFeatureFlagChangeActivity validates a feature flag change request
func (a *FeatureFlagActivities) ValidateFeatureFlagChangeActivity(ctx context.Context, req *workflow.FeatureFlagChangeRequest) (*workflow.ValidationResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Validating feature flag change", "flag_name", req.FlagName, "change_type", req.ChangeType)

	result := &workflow.ValidationResult{IsValid: true}

	// Validate flag name
	if req.FlagName == "" {
		result.IsValid = false
		result.ErrorMessage = "Flag name is required"
		return result, nil
	}

	// Check if flag exists
	flag, err := a.featureFlagService.GetFeatureFlagByName(ctx, req.FlagName)
	if err != nil {
		result.IsValid = false
		result.ErrorMessage = fmt.Sprintf("Flag not found: %s", req.FlagName)
		return result, nil
	}

	// Validate change type
	validChangeTypes := map[string]bool{
		"enable":          true,
		"disable":         true,
		"update_rollout":  true,
		"update_config":   true,
		"rollback":        true,
	}

	if !validChangeTypes[req.ChangeType] {
		result.IsValid = false
		result.ErrorMessage = fmt.Sprintf("Invalid change type: %s", req.ChangeType)
		return result, nil
	}

	// Validate new value based on change type
	switch req.ChangeType {
	case "enable", "disable":
		if req.NewValue != true && req.NewValue != false {
			result.IsValid = false
			result.ErrorMessage = "New value must be boolean for enable/disable operations"
			return result, nil
		}
	case "update_rollout":
		if percentage, ok := req.NewValue.(float64); ok {
			if percentage < 0 || percentage > 100 {
				result.IsValid = false
				result.ErrorMessage = "Rollout percentage must be between 0 and 100"
				return result, nil
			}
		} else {
			result.IsValid = false
			result.ErrorMessage = "New value must be a number for rollout updates"
			return result, nil
		}
	}

	// Validate justification
	if len(req.Justification) < 10 {
		result.IsValid = false
		result.ErrorMessage = "Justification must be at least 10 characters"
		return result, nil
	}

	// Check if flag is in a valid state for the requested change
	if req.ChangeType == "enable" && flag.DefaultValue {
		result.IsValid = false
		result.ErrorMessage = "Flag is already enabled"
		return result, nil
	}

	if req.ChangeType == "disable" && !flag.DefaultValue {
		result.IsValid = false
		result.ErrorMessage = "Flag is already disabled"
		return result, nil
	}

	logger.Info("Feature flag change validation passed", "flag_name", req.FlagName)
	return result, nil
}

// CheckApprovalRequiredActivity checks if approval is required for the change
func (a *FeatureFlagActivities) CheckApprovalRequiredActivity(ctx context.Context, req *workflow.FeatureFlagChangeRequest) (bool, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Checking if approval is required", "flag_name", req.FlagName, "change_type", req.ChangeType)

	// Check approval policy
	requiresApproval, err := a.policyService.CheckApprovalRequired(ctx, req.TenantID, req.FlagName, req.ChangeType)
	if err != nil {
		logger.Error("Failed to check approval policy", "error", err)
		return false, fmt.Errorf("failed to check approval policy: %w", err)
	}

	logger.Info("Approval requirement determined", "flag_name", req.FlagName, "requires_approval", requiresApproval)
	return requiresApproval, nil
}

// CreateAccessRequestActivity creates an access request for the flag change
func (a *FeatureFlagActivities) CreateAccessRequestActivity(ctx context.Context, req *workflow.FeatureFlagChangeRequest) (*workflow.CreateAccessRequestResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Creating access request", "flag_name", req.FlagName, "requested_by", req.RequestedBy)

	// Create access request
	accessReq := &request.CreateAccessRequestRequest{
		EntityID:       uuid.New(), // Use a generated ID for the flag change request
		RequestType:    request.RequestTypeResourceAccess, // Treat as resource access
		Justification:  fmt.Sprintf("Feature flag change: %s\n%s", req.ChangeType, req.Justification),
		BusinessReason: &req.BusinessReason,
		DurationHours:  nil, // Flag changes are permanent until next change
	}

	accessRequestResult, err := a.accessRequestRepo.CreateAccessRequest(ctx, accessReq, req.RequestedBy)
	if err != nil {
		logger.Error("Failed to create access request", "error", err)
		return nil, fmt.Errorf("failed to create access request: %w", err)
	}

	result := &workflow.CreateAccessRequestResult{
		RequestID: accessRequestResult.ID,
	}

	logger.Info("Access request created", "request_id", result.RequestID, "flag_name", req.FlagName)
	return result, nil
}

// ApplyFeatureFlagChangeActivity applies the actual flag change
func (a *FeatureFlagActivities) ApplyFeatureFlagChangeActivity(ctx context.Context, req *workflow.FeatureFlagChangeRequest) (*workflow.ApplyChangeResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Applying feature flag change", "flag_name", req.FlagName, "change_type", req.ChangeType)

	// Get current flag state
	currentFlag, err := a.featureFlagService.GetFeatureFlagByName(ctx, req.FlagName)
	if err != nil {
		logger.Error("Failed to get current flag", "error", err)
		return nil, fmt.Errorf("failed to get current flag: %w", err)
	}

	oldValue := currentFlag.DefaultValue

	// Apply the change based on type
	var updatedFlag *featureflag.FeatureFlag
	switch req.ChangeType {
	case "enable":
		updateReq := &featureflag.UpdateFeatureFlagRequest{
			DefaultValue: func() *bool { v := true; return &v }(),
		}
		updatedFlag, err = a.featureFlagService.UpdateFeatureFlag(ctx, currentFlag.ID, updateReq)
	case "disable":
		updateReq := &featureflag.UpdateFeatureFlagRequest{
			DefaultValue: func() *bool { v := false; return &v }(),
		}
		updatedFlag, err = a.featureFlagService.UpdateFeatureFlag(ctx, currentFlag.ID, updateReq)
	case "update_rollout":
		if percentage, ok := req.NewValue.(float64); ok {
			rolloutPercentage := int32(percentage)
			updateReq := &featureflag.UpdateFeatureFlagRequest{
				RolloutPercentage: &rolloutPercentage,
			}
			updatedFlag, err = a.featureFlagService.UpdateFeatureFlag(ctx, currentFlag.ID, updateReq)
		} else {
			err = fmt.Errorf("invalid rollout percentage type")
		}
	case "rollback":
		// For rollback, we typically disable the flag
		updateReq := &featureflag.UpdateFeatureFlagRequest{
			DefaultValue: func() *bool { v := false; return &v }(),
		}
		updatedFlag, err = a.featureFlagService.UpdateFeatureFlag(ctx, currentFlag.ID, updateReq)
	default:
		err = fmt.Errorf("unsupported change type: %s", req.ChangeType)
	}

	if err != nil {
		logger.Error("Failed to apply flag change", "error", err)
		return nil, fmt.Errorf("failed to apply flag change: %w", err)
	}

	result := &workflow.ApplyChangeResult{
		FlagID:    updatedFlag.ID,
		OldValue:  oldValue,
		AppliedAt: time.Now(),
	}

	logger.Info("Feature flag change applied successfully", 
		"flag_id", result.FlagID, 
		"flag_name", req.FlagName, 
		"old_value", oldValue,
		"new_value", req.NewValue)

	return result, nil
}

// SendWebSocketNotificationActivity sends real-time notifications
func (a *FeatureFlagActivities) SendWebSocketNotificationActivity(ctx context.Context, req *workflow.NotificationRequest) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Sending WebSocket notification", "flag_name", req.FlagName, "change_type", req.ChangeType)

	event := &FeatureFlagChangeEvent{
		FlagName:        req.FlagName,
		ChangeType:      req.ChangeType,
		NewValue:        req.NewValue,
		ChangedBy:       req.AppliedBy,
		AccessRequestID: req.AccessRequestID,
		AppliedAt:       req.AppliedAt,
	}

	err := a.webSocketService.NotifyFlagChange(ctx, req.TenantID, event)
	if err != nil {
		logger.Error("Failed to send WebSocket notification", "error", err)
		return fmt.Errorf("failed to send WebSocket notification: %w", err)
	}

	logger.Info("WebSocket notification sent successfully", "flag_name", req.FlagName)
	return nil
}

// CompleteAccessRequestActivity marks an access request as completed
func (a *FeatureFlagActivities) CompleteAccessRequestActivity(ctx context.Context, accessRequestID uuid.UUID) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Completing access request", "request_id", accessRequestID)

	// Update access request status to completed
	updateReq := &request.UpdateAccessRequestRequest{
		ApprovalStatus:   request.ApprovalStatusApproved,
		ApprovalComments: func() *string { s := "Feature flag change completed successfully"; return &s }(),
	}

	_, err := a.accessRequestRepo.UpdateAccessRequestStatus(ctx, accessRequestID, updateReq, uuid.Nil) // System completion
	if err != nil {
		logger.Error("Failed to complete access request", "error", err)
		return fmt.Errorf("failed to complete access request: %w", err)
	}

	logger.Info("Access request completed successfully", "request_id", accessRequestID)
	return nil
}

// ExpireAccessRequestActivity expires an access request due to timeout
func (a *FeatureFlagActivities) ExpireAccessRequestActivity(ctx context.Context, accessRequestID uuid.UUID) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Expiring access request", "request_id", accessRequestID)

	err := a.accessRequestRepo.ExpireAccessRequest(ctx, accessRequestID)
	if err != nil {
		logger.Error("Failed to expire access request", "error", err)
		return fmt.Errorf("failed to expire access request: %w", err)
	}

	logger.Info("Access request expired successfully", "request_id", accessRequestID)
	return nil
}

// CreateAuditEventActivity creates an audit event for the flag change
func (a *FeatureFlagActivities) CreateAuditEventActivity(ctx context.Context, req *workflow.AuditEventRequest) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Creating audit event", "flag_name", req.FlagName, "change_type", req.ChangeType)

	err := a.auditService.CreateAuditEvent(ctx, req)
	if err != nil {
		logger.Error("Failed to create audit event", "error", err)
		return fmt.Errorf("failed to create audit event: %w", err)
	}

	logger.Info("Audit event created successfully", "flag_name", req.FlagName)
	return nil
}

// CheckRollbackNeededActivity checks if auto-rollback is still needed
func (a *FeatureFlagActivities) CheckRollbackNeededActivity(ctx context.Context, flagID uuid.UUID) (bool, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Checking if rollback is needed", "flag_id", flagID)

	// Get current flag state
	flag, err := a.featureFlagService.GetFeatureFlagByID(ctx, flagID)
	if err != nil {
		logger.Error("Failed to get flag", "error", err)
		return false, fmt.Errorf("failed to get flag: %w", err)
	}

	// Check if flag is still in a state that requires rollback
	// This logic can be customized based on business requirements
	needsRollback := flag.DefaultValue // If flag is still enabled, it might need rollback

	logger.Info("Rollback check completed", "flag_id", flagID, "needs_rollback", needsRollback)
	return needsRollback, nil
}

// Policy service implementation for checking approval requirements
type DefaultPolicyService struct{}

func NewDefaultPolicyService() PolicyService {
	return &DefaultPolicyService{}
}

// GetApprovalPolicy retrieves approval policy for a flag
func (s *DefaultPolicyService) GetApprovalPolicy(ctx context.Context, tenantID uuid.UUID, flagName string, changeType string) (*workflow.FeatureFlagApprovalPolicy, error) {
	// Default policy - require approval for production-like flags
	productionPattern := regexp.MustCompile(`(?i)prod|production|live|critical`)
	
	policy := &workflow.FeatureFlagApprovalPolicy{
		TenantID:             tenantID,
		FlagNamePattern:      ".*", // Match all flags by default
		ChangeTypes:          []string{"enable", "disable", "update_rollout"},
		RequiresApproval:     productionPattern.MatchString(flagName),
		MinApprovers:         1,
		ApprovalTimeoutHours: 24,
	}

	return policy, nil
}

// CheckApprovalRequired checks if approval is required
func (s *DefaultPolicyService) CheckApprovalRequired(ctx context.Context, tenantID uuid.UUID, flagName string, changeType string) (bool, error) {
	policy, err := s.GetApprovalPolicy(ctx, tenantID, flagName, changeType)
	if err != nil {
		return false, err
	}

	// Check if change type requires approval
	for _, ct := range policy.ChangeTypes {
		if ct == changeType {
			return policy.RequiresApproval, nil
		}
	}

	return false, nil
}