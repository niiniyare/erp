//go:build ignore

package approval

//go:generate go run go.uber.org/mock/mockgen -source=approver_service.go -destination=mock.go -package=approval

import (
	"context"

	"github.com/google/uuid"
	"awo.so/internal/core/iam"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
	"awo.so/internal/shared/types"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// Type aliases for external types
type (
	User          = iam.User
	UserService   = iam.UserService
	AccessRequest = types.AccessRequest
)

// ApproverService handles approver validation and determination
type ApproverService interface {
	// Approver validation
	ValidateApprover(ctx context.Context, approverID uuid.UUID, req *AccessRequest) (*ApproverValidationResult, error)
	GetEligibleApprovers(ctx context.Context, req *AccessRequest) ([]uuid.UUID, error)

	// Approval rules management
	CreateApprovalRule(ctx context.Context, rule *ApprovalRule) error
	GetApprovalRules(ctx context.Context, entityID uuid.UUID, reqType RequestType) ([]*ApprovalRule, error)
	UpdateApprovalRule(ctx context.Context, ruleID uuid.UUID, rule *ApprovalRule) error
	DeleteApprovalRule(ctx context.Context, ruleID uuid.UUID) error

	// Hierarchy operations
	GetUserHierarchy(ctx context.Context, userID uuid.UUID) (*UserHierarchy, error)
	GetManagerChain(ctx context.Context, userID uuid.UUID, levels int) ([]uuid.UUID, error)
	GetDirectReports(ctx context.Context, managerID uuid.UUID) ([]uuid.UUID, error)

	// Permission validation
	CheckApprovalPermissions(ctx context.Context, approverID uuid.UUID, targetResource string, reqType RequestType) (bool, error)
}

// approverService implements ApproverService
type approverService struct {
	userService UserService
	tracing     tracing.Service
	metrics     metrics.MetricsProvider
	// TODO: Add approval rules repository when implemented
	// approvalRulesRepo ApprovalRulesRepository
}

// NewApproverService creates a new approver service
func NewApproverService(
	userService UserService,
	tracing tracing.Service,
	metrics metrics.MetricsProvider,
) ApproverService {
	return &approverService{
		userService: userService,
		tracing:     tracing,
		metrics:     metrics,
	}
}

// ValidateApprover validates if a user can approve a specific access req
func (s *approverService) ValidateApprover(ctx context.Context, approverID uuid.UUID, req *AccessRequest) (*ApproverValidationResult, error) {
	ctx, span := s.tracing.StartSpan(ctx, "approverService.ValidateApprover")
	defer span.End()

	span.SetAttributes(
		attribute.String("approver_id", approverID.String()),
		attribute.String("request_id", req.ID.String()),
		attribute.String("request_type", string(req.RequestType)),
	)

	logger.Info("Validating approver", logger.Fields{
		"approver_id":  approverID,
		"request_id":   req.ID,
		"request_type": req.RequestType,
	})

	// Get approver user details
	approver, err := s.userService.GetUserByID(ctx, approverID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get approver user")
		return &ApproverValidationResult{
			IsValid: false,
			Reason:  "Approver not found",
		}, err
	}

	// Basic validation: approver must be active
	if !approver.IsActive || string(approver.AccountStatus) != "ACTIVE" {
		return &ApproverValidationResult{
			IsValid: false,
			Reason:  "Approver account is not active",
		}, nil
	}

	// Get reqer details
	reqer, err := s.userService.GetUserByID(ctx, req.RequesterID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get reqer user")
		return &ApproverValidationResult{
			IsValid: false,
			Reason:  "Requester not found",
		}, err
	}

	// Self-approval check: users cannot approve their own reqs
	if approverID == req.RequesterID {
		return &ApproverValidationResult{
			IsValid: false,
			Reason:  "Users cannot approve their own reqs",
		}, nil
	}

	// Get approval rules for this req type and entity
	rules, err := s.GetApprovalRules(ctx, req.EntityID, RequestType(req.RequestType))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get approval rules")
		return &ApproverValidationResult{
			IsValid: false,
			Reason:  "Failed to retrieve approval rules",
		}, err
	}

	// If no specific rules exist, use default validation
	if len(rules) == 0 {
		return s.performDefaultValidation(ctx, approver, reqer, req)
	}

	// Evaluate approval rules
	for _, rule := range rules {
		if s.evaluateApprovalRule(ctx, rule, approver, reqer, req) {
			s.metrics.IncrementCounter("approver_validation_success", map[string]any{"rule_id": rule.ID.String()})
			return &ApproverValidationResult{
				IsValid:           true,
				ApprovalRules:     []ApprovalRule{*rule},
				RequiredApprovals: rule.Approvers.RequiredApprovals,
			}, nil
		}
	}

	s.metrics.IncrementCounter("approver_validation_failed", map[string]any{"request_type": string(req.RequestType)})
	return &ApproverValidationResult{
		IsValid: false,
		Reason:  "No approval rules match this approver for the reqed access",
	}, nil
}

// GetEligibleApprovers gets all eligible approvers for an access req
func (s *approverService) GetEligibleApprovers(ctx context.Context, req *AccessRequest) ([]uuid.UUID, error) {
	ctx, span := s.tracing.StartSpan(ctx, "approverService.GetEligibleApprovers")
	defer span.End()

	var eligibleApprovers []uuid.UUID

	// Get approval rules
	rules, err := s.GetApprovalRules(ctx, req.EntityID, RequestType(req.RequestType))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get approval rules")
		return nil, err
	}

	// If no rules exist, use default logic
	if len(rules) == 0 {
		return s.getDefaultApprovers(ctx, req)
	}

	// Collect approvers from all applicable rules
	approverSet := make(map[uuid.UUID]bool)

	for _, rule := range rules {
		ruleApprovers, err := s.getApproversFromRule(ctx, rule, req)
		if err != nil {
			logger.Error("Failed to get approvers from rule", logger.Fields{
				"rule_id": rule.ID,
				"error":   err.Error(),
			})
			continue
		}

		for _, approverID := range ruleApprovers {
			approverSet[approverID] = true
		}
	}

	// Convert set to slice
	for approverID := range approverSet {
		eligibleApprovers = append(eligibleApprovers, approverID)
	}

	// Remove the reqer from eligible approvers
	filteredApprovers := make([]uuid.UUID, 0)
	for _, approverID := range eligibleApprovers {
		if approverID != req.RequesterID {
			filteredApprovers = append(filteredApprovers, approverID)
		}
	}

	s.metrics.ObserveHistogram("eligible_approvers_count", float64(len(filteredApprovers)), map[string]any{"request_type": string(req.RequestType)})
	return filteredApprovers, nil
}

// CreateApprovalRule creates a new approval rule
func (s *approverService) CreateApprovalRule(ctx context.Context, rule *ApprovalRule) error {
	// TODO: Implement when approval rules repository is available
	logger.Info("Creating approval rule", logger.Fields{
		"rule_name":    rule.Name,
		"entity_id":    rule.EntityID,
		"request_type": rule.RequestType,
	})
	return nil
}

// GetApprovalRules gets approval rules for entity and req type
func (s *approverService) GetApprovalRules(ctx context.Context, entityID uuid.UUID, reqType RequestType) ([]*ApprovalRule, error) {
	// TODO: Implement database storage for approval rules
	// For now, return default rules based on req type

	defaultRules := s.getDefaultApprovalRules(entityID, reqType)
	return defaultRules, nil
}

// UpdateApprovalRule updates an existing approval rule
func (s *approverService) UpdateApprovalRule(ctx context.Context, ruleID uuid.UUID, rule *ApprovalRule) error {
	// TODO: Implement when approval rules repository is available
	logger.Info("Updating approval rule", logger.Fields{
		"rule_id": ruleID,
	})
	return nil
}

// DeleteApprovalRule deletes an approval rule
func (s *approverService) DeleteApprovalRule(ctx context.Context, ruleID uuid.UUID) error {
	// TODO: Implement when approval rules repository is available
	logger.Info("Deleting approval rule", logger.Fields{
		"rule_id": ruleID,
	})
	return nil
}

// GetUserHierarchy gets a user's organizational hierarchy information
func (s *approverService) GetUserHierarchy(ctx context.Context, userID uuid.UUID) (*UserHierarchy, error) {
	ctx, span := s.tracing.StartSpan(ctx, "approverService.GetUserHierarchy")
	defer span.End()

	// TODO: Implement hierarchy retrieval
	// For now, return basic hierarchy information
	_, err := s.userService.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	hierarchy := &UserHierarchy{
		UserID:        userID,
		Level:         0, // TODO: Calculate from hierarchy
		SecurityLevel: 0, // TODO: Get from employee record
		DirectReports: make([]uuid.UUID, 0),
		ManagerChain:  make([]uuid.UUID, 0),
	}

	return hierarchy, nil
}

// GetManagerChain gets the manager chain for a user
func (s *approverService) GetManagerChain(ctx context.Context, userID uuid.UUID, levels int) ([]uuid.UUID, error) {
	// TODO: Implement hierarchical manager chain retrieval
	// This would traverse the employee table's manager_id relationships
	return []uuid.UUID{}, nil
}

// GetDirectReports gets direct reports for a manager
func (s *approverService) GetDirectReports(ctx context.Context, managerID uuid.UUID) ([]uuid.UUID, error) {
	// TODO: Implement direct reports retrieval
	// This would query employees where manager_id = managerID
	return []uuid.UUID{}, nil
}

// CheckApprovalPermissions checks if an approver has permission to approve a specific type of req
func (s *approverService) CheckApprovalPermissions(ctx context.Context, approverID uuid.UUID, targetResource string, reqType RequestType) (bool, error) {
	ctx, span := s.tracing.StartSpan(ctx, "approverService.CheckApprovalPermissions")
	defer span.End()

	// TODO: Implement permission checking
	// This would integrate with the permission evaluation system to check if the approver
	// has the necessary permissions to approve access to the target resource

	logger.Info("Checking approval permissions", logger.Fields{
		"approver_id":     approverID,
		"target_resource": targetResource,
		"req_type":        reqType,
	})

	// For now, assume all active users can approve
	approver, err := s.userService.GetUserByID(ctx, approverID)
	if err != nil {
		return false, err
	}

	return approver.IsActive && string(approver.AccountStatus) == "ACTIVE", nil
}

// Helper methods

// performDefaultValidation performs default approver validation when no specific rules exist
func (s *approverService) performDefaultValidation(ctx context.Context, approver, reqer *User, req *AccessRequest) (*ApproverValidationResult, error) {
	// Default validation logic:
	// 1. Approver must be active
	// 2. Approver cannot be the reqer
	// 3. Approver should be in the same entity or a parent entity
	// 4. For sensitive operations, approver should have higher security level

	if !approver.IsActive || string(approver.AccountStatus) != "ACTIVE" {
		return &ApproverValidationResult{
			IsValid: false,
			Reason:  "Approver account is not active",
		}, nil
	}

	if approver.ID == reqer.ID {
		return &ApproverValidationResult{
			IsValid: false,
			Reason:  "Users cannot approve their own reqs",
		}, nil
	}

	// TODO: Add entity scope validation
	// TODO: Add security level validation for sensitive reqs

	return &ApproverValidationResult{
		IsValid:           true,
		RequiredApprovals: 1,
	}, nil
}

// evaluateApprovalRule evaluates if an approval rule applies to an approver for a req
func (s *approverService) evaluateApprovalRule(ctx context.Context, rule *ApprovalRule, approver, reqer *User, req *AccessRequest) bool {
	// Check if rule is active
	if !rule.IsActive {
		return false
	}

	// Check req type match
	if rule.RequestType != RequestType(req.RequestType) {
		return false
	}

	// Evaluate based on approver configuration type
	switch rule.Approvers.Type {
	case "ROLE_BASED":
		return s.evaluateRoleBasedApproval(ctx, rule.Approvers.RoleBasedApprovers, approver, reqer, req)
	case "HIERARCHY_BASED":
		return s.evaluateHierarchyBasedApproval(ctx, rule.Approvers.HierarchyApprovers, approver, reqer, req)
	case "SPECIFIC_USERS":
		return s.evaluateSpecificUserApproval(rule.Approvers.SpecificApprovers, approver.ID)
	case "MIXED":
		// Evaluate multiple criteria (AND logic)
		return s.evaluateRoleBasedApproval(ctx, rule.Approvers.RoleBasedApprovers, approver, reqer, req) ||
			s.evaluateHierarchyBasedApproval(ctx, rule.Approvers.HierarchyApprovers, approver, reqer, req) ||
			s.evaluateSpecificUserApproval(rule.Approvers.SpecificApprovers, approver.ID)
	default:
		logger.Warn("Unknown approver configuration type", logger.Fields{
			"type":    rule.Approvers.Type,
			"rule_id": rule.ID,
		})
		return false
	}
}

// evaluateRoleBasedApproval evaluates role-based approval rules
func (s *approverService) evaluateRoleBasedApproval(ctx context.Context, config *RoleBasedApprovers, approver, reqer *User, req *AccessRequest) bool {
	if config == nil {
		return false
	}

	// Check if reqer is excluded from approving their own req
	if config.ExcludeRequester && approver.ID == reqer.ID {
		return false
	}

	// TODO: Implement role checking
	// This would check if the approver has one of the required roles
	// and meets the minimum security level requirements

	return true // Placeholder
}

// evaluateHierarchyBasedApproval evaluates hierarchy-based approval rules
func (s *approverService) evaluateHierarchyBasedApproval(ctx context.Context, config *HierarchyApprovers, approver, reqer *User, req *AccessRequest) bool {
	if config == nil {
		return false
	}

	// TODO: Implement hierarchy checking
	// This would check the organizational relationship between approver and reqer

	return true // Placeholder
}

// evaluateSpecificUserApproval evaluates specific user approval rules
func (s *approverService) evaluateSpecificUserApproval(specificApprovers []uuid.UUID, approverID uuid.UUID) bool {
	for _, allowedApprover := range specificApprovers {
		if allowedApprover == approverID {
			return true
		}
	}
	return false
}

// getApproversFromRule gets eligible approvers from a specific rule
func (s *approverService) getApproversFromRule(ctx context.Context, rule *ApprovalRule, req *AccessRequest) ([]uuid.UUID, error) {
	var approvers []uuid.UUID

	switch rule.Approvers.Type {
	case "SPECIFIC_USERS":
		approvers = append(approvers, rule.Approvers.SpecificApprovers...)
	case "ROLE_BASED":
		// TODO: Query users with specific roles
		// This would find all users with the required roles in the entity
	case "HIERARCHY_BASED":
		// TODO: Find users based on hierarchy
		// This would find managers, peers, or subordinates based on configuration
	}

	// Add fallback approvers if primary approvers are not available
	if len(approvers) == 0 && len(rule.Approvers.FallbackApprovers) > 0 {
		approvers = append(approvers, rule.Approvers.FallbackApprovers...)
	}

	return approvers, nil
}

// getDefaultApprovers returns default approvers when no rules are configured
func (s *approverService) getDefaultApprovers(ctx context.Context, req *AccessRequest) ([]uuid.UUID, error) {
	// TODO: Implement default approver logic
	// This could include:
	// - Entity administrators
	// - Department managers
	// - System administrators for the tenant

	return []uuid.UUID{}, nil
}

// getDefaultApprovalRules returns default approval rules for a req type
func (s *approverService) getDefaultApprovalRules(entityID uuid.UUID, reqType RequestType) []*ApprovalRule {
	rules := make([]*ApprovalRule, 0)

	// Create default rules based on req type
	switch reqType {
	case RequestTypeRoleAssignment:
		rule := &ApprovalRule{
			ID:          uuid.New(),
			EntityID:    entityID,
			Name:        "Default Role Assignment Approval",
			RequestType: reqType,
			Approvers: ApproverConfiguration{
				Type:              "HIERARCHY_BASED",
				RequiredApprovals: 1,
				HierarchyApprovers: &HierarchyApprovers{
					ManagerLevelsUp:      1,
					RequireDirectManager: true,
				},
			},
			Priority: 100,
			IsActive: true,
		}
		rules = append(rules, rule)

	case RequestTypeElevation:
		rule := &ApprovalRule{
			ID:          uuid.New(),
			EntityID:    entityID,
			Name:        "Default Privilege Elevation Approval",
			RequestType: reqType,
			Approvers: ApproverConfiguration{
				Type:              "ROLE_BASED",
				RequiredApprovals: 2, // Require two approvals for elevation
				RoleBasedApprovers: &RoleBasedApprovers{
					MinSecurityLevel: intPtr(5), // Require high security level
					ExcludeRequester: true,
				},
			},
			Priority: 200,
			IsActive: true,
		}
		rules = append(rules, rule)
	}

	return rules
}

// Helper function to create int pointer
func intPtr(i int) *int {
	return &i
}
