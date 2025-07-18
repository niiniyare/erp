package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// ApprovalRule represents a rule for determining approvers
type ApprovalRule struct {
	ID           uuid.UUID                `json:"id"`
	TenantID     uuid.UUID                `json:"tenant_id"`
	EntityID     uuid.UUID                `json:"entity_id"`
	Name         string                   `json:"name"`
	RequestType  RequestType              `json:"request_type"`
	Conditions   map[string]any   `json:"conditions"`   // ABAC-style conditions
	Approvers    ApproverConfiguration    `json:"approvers"`
	Priority     int                      `json:"priority"`     // Higher priority rules evaluated first
	IsActive     bool                     `json:"is_active"`
}

// ApproverConfiguration defines who can approve requests
type ApproverConfiguration struct {
	Type                string               `json:"type"` // ROLE_BASED, HIERARCHY_BASED, SPECIFIC_USERS, MIXED
	RequiredApprovals   int                  `json:"required_approvals"`
	RoleBasedApprovers  *RoleBasedApprovers  `json:"role_based_approvers,omitempty"`
	HierarchyApprovers  *HierarchyApprovers  `json:"hierarchy_approvers,omitempty"`
	SpecificApprovers   []uuid.UUID          `json:"specific_approvers,omitempty"`
	FallbackApprovers   []uuid.UUID          `json:"fallback_approvers,omitempty"`
}

// RoleBasedApprovers defines approval based on roles
type RoleBasedApprovers struct {
	RequiredRoles     []uuid.UUID `json:"required_roles"`     // Users must have one of these roles
	MinSecurityLevel  *int        `json:"min_security_level,omitempty"` // Minimum security clearance
	SameEntity        bool        `json:"same_entity"`        // Must be in same entity
	ExcludeRequester  bool        `json:"exclude_requester"`  // Cannot approve own requests
}

// HierarchyApprovers defines approval based on organizational hierarchy
type HierarchyApprovers struct {
	ManagerLevelsUp   int  `json:"manager_levels_up"`   // How many levels up in hierarchy
	IncludePeers      bool `json:"include_peers"`       // Include users at same level
	IncludeSubordinates bool `json:"include_subordinates"` // Include direct reports
	RequireDirectManager bool `json:"require_direct_manager"` // Must be direct manager
}

// ApproverValidationResult represents the result of approver validation
type ApproverValidationResult struct {
	IsValid           bool     `json:"is_valid"`
	Reason            string   `json:"reason,omitempty"`
	ApprovalRules     []ApprovalRule `json:"approval_rules"`
	RequiredApprovals int      `json:"required_approvals"`
	SecurityLevel     *int     `json:"security_level,omitempty"`
}

// ApproverService handles approver validation and determination
type ApproverService interface {
	// Approver validation
	ValidateApprover(ctx context.Context, approverID uuid.UUID, request *AccessRequest) (*ApproverValidationResult, error)
	GetEligibleApprovers(ctx context.Context, request *AccessRequest) ([]uuid.UUID, error)
	
	// Approval rules management
	CreateApprovalRule(ctx context.Context, rule *ApprovalRule) error
	GetApprovalRules(ctx context.Context, entityID uuid.UUID, requestType RequestType) ([]*ApprovalRule, error)
	UpdateApprovalRule(ctx context.Context, ruleID uuid.UUID, rule *ApprovalRule) error
	DeleteApprovalRule(ctx context.Context, ruleID uuid.UUID) error
	
	// Hierarchy operations
	GetUserHierarchy(ctx context.Context, userID uuid.UUID) (*UserHierarchy, error)
	GetManagerChain(ctx context.Context, userID uuid.UUID, levels int) ([]uuid.UUID, error)
	GetDirectReports(ctx context.Context, managerID uuid.UUID) ([]uuid.UUID, error)
	
	// Permission validation
	CheckApprovalPermissions(ctx context.Context, approverID uuid.UUID, targetResource string, requestType RequestType) (bool, error)
}

// UserHierarchy represents a user's position in organizational hierarchy
type UserHierarchy struct {
	UserID           uuid.UUID   `json:"user_id"`
	ManagerID        *uuid.UUID  `json:"manager_id,omitempty"`
	Level            int         `json:"level"`            // Organizational level (0 = top)
	SecurityLevel    int         `json:"security_level"`   // Security clearance level
	DepartmentID     *uuid.UUID  `json:"department_id,omitempty"`
	DirectReports    []uuid.UUID `json:"direct_reports"`
	ManagerChain     []uuid.UUID `json:"manager_chain"`    // All managers up the chain
}

// approverService implements ApproverService
type approverService struct {
	userRepo     Repository
	tracing      *tracing.TracingService
	metrics      *metrics.MetricsService
	// TODO: Add approval rules repository when implemented
	// approvalRulesRepo ApprovalRulesRepository
}

// NewApproverService creates a new approver service
func NewApproverService(
	userRepo Repository,
	tracing *tracing.TracingService,
	metrics *metrics.MetricsService,
) ApproverService {
	return &approverService{
		userRepo: userRepo,
		tracing:  tracing,
		metrics:  metrics,
	}
}

// ValidateApprover validates if a user can approve a specific access request
func (s *approverService) ValidateApprover(ctx context.Context, approverID uuid.UUID, request *AccessRequest) (*ApproverValidationResult, error) {
	ctx, span := s.tracing.StartSpan(ctx, "approverService.ValidateApprover")
	defer span.End()

	span.SetAttributes(
		attribute.String("approver_id", approverID.String()),
		attribute.String("request_id", request.ID.String()),
		attribute.String("request_type", string(request.RequestType)),
	)

	logger.Info("Validating approver", logger.Fields{
		"approver_id": approverID,
		"request_id":  request.ID,
		"request_type": request.RequestType,
	})

	// Get approver user details
	approver, err := s.userRepo.GetUserByID(ctx, approverID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get approver user")
		return &ApproverValidationResult{
			IsValid: false,
			Reason:  "Approver not found",
		}, err
	}

	// Basic validation: approver must be active
	if !approver.IsActive || approver.AccountStatus != AccountStatusActive {
		return &ApproverValidationResult{
			IsValid: false,
			Reason:  "Approver account is not active",
		}, nil
	}

	// Get requester details
	requester, err := s.userRepo.GetUserByID(ctx, request.RequesterID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get requester user")
		return &ApproverValidationResult{
			IsValid: false,
			Reason:  "Requester not found",
		}, err
	}

	// Self-approval check: users cannot approve their own requests
	if approverID == request.RequesterID {
		return &ApproverValidationResult{
			IsValid: false,
			Reason:  "Users cannot approve their own requests",
		}, nil
	}

	// Get approval rules for this request type and entity
	rules, err := s.GetApprovalRules(ctx, request.EntityID, request.RequestType)
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
		return s.performDefaultValidation(ctx, approver, requester, request)
	}

	// Evaluate approval rules
	for _, rule := range rules {
		if s.evaluateApprovalRule(ctx, rule, approver, requester, request) {
			s.metrics.IncrementCounter("approver_validation_success", map[string]any{"rule_id": rule.ID.String()})
			return &ApproverValidationResult{
				IsValid:           true,
				ApprovalRules:     []ApprovalRule{*rule},
				RequiredApprovals: rule.Approvers.RequiredApprovals,
			}, nil
		}
	}

	s.metrics.IncrementCounter("approver_validation_failed", map[string]any{"request_type": string(request.RequestType)})
	return &ApproverValidationResult{
		IsValid: false,
		Reason:  "No approval rules match this approver for the requested access",
	}, nil
}

// GetEligibleApprovers gets all eligible approvers for an access request
func (s *approverService) GetEligibleApprovers(ctx context.Context, request *AccessRequest) ([]uuid.UUID, error) {
	ctx, span := s.tracing.StartSpan(ctx, "approverService.GetEligibleApprovers")
	defer span.End()

	var eligibleApprovers []uuid.UUID

	// Get approval rules
	rules, err := s.GetApprovalRules(ctx, request.EntityID, request.RequestType)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get approval rules")
		return nil, err
	}

	// If no rules exist, use default logic
	if len(rules) == 0 {
		return s.getDefaultApprovers(ctx, request)
	}

	// Collect approvers from all applicable rules
	approverSet := make(map[uuid.UUID]bool)
	
	for _, rule := range rules {
		ruleApprovers, err := s.getApproversFromRule(ctx, rule, request)
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

	// Remove the requester from eligible approvers
	filteredApprovers := make([]uuid.UUID, 0)
	for _, approverID := range eligibleApprovers {
		if approverID != request.RequesterID {
			filteredApprovers = append(filteredApprovers, approverID)
		}
	}

	s.metrics.ObserveHistogram("eligible_approvers_count", float64(len(filteredApprovers)), map[string]any{"request_type": string(request.RequestType)})
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

// GetApprovalRules gets approval rules for entity and request type
func (s *approverService) GetApprovalRules(ctx context.Context, entityID uuid.UUID, requestType RequestType) ([]*ApprovalRule, error) {
	// TODO: Implement database storage for approval rules
	// For now, return default rules based on request type
	
	defaultRules := s.getDefaultApprovalRules(entityID, requestType)
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

	// TODO: Implement comprehensive hierarchy retrieval
	// For now, return basic hierarchy information
	_, err := s.userRepo.GetUserByID(ctx, userID)
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

// CheckApprovalPermissions checks if an approver has permission to approve a specific type of request
func (s *approverService) CheckApprovalPermissions(ctx context.Context, approverID uuid.UUID, targetResource string, requestType RequestType) (bool, error) {
	ctx, span := s.tracing.StartSpan(ctx, "approverService.CheckApprovalPermissions")
	defer span.End()

	// TODO: Implement comprehensive permission checking
	// This would integrate with the permission evaluation system to check if the approver
	// has the necessary permissions to approve access to the target resource
	
	logger.Info("Checking approval permissions", logger.Fields{
		"approver_id":     approverID,
		"target_resource": targetResource,
		"request_type":    requestType,
	})

	// For now, assume all active users can approve
	approver, err := s.userRepo.GetUserByID(ctx, approverID)
	if err != nil {
		return false, err
	}

	return approver.IsActive && approver.AccountStatus == AccountStatusActive, nil
}

// Helper methods

// performDefaultValidation performs default approver validation when no specific rules exist
func (s *approverService) performDefaultValidation(ctx context.Context, approver, requester *User, request *AccessRequest) (*ApproverValidationResult, error) {
	// Default validation logic:
	// 1. Approver must be active
	// 2. Approver cannot be the requester
	// 3. Approver should be in the same entity or a parent entity
	// 4. For sensitive operations, approver should have higher security level

	if !approver.IsActive || approver.AccountStatus != AccountStatusActive {
		return &ApproverValidationResult{
			IsValid: false,
			Reason:  "Approver account is not active",
		}, nil
	}

	if approver.ID == requester.ID {
		return &ApproverValidationResult{
			IsValid: false,
			Reason:  "Users cannot approve their own requests",
		}, nil
	}

	// TODO: Add entity scope validation
	// TODO: Add security level validation for sensitive requests

	return &ApproverValidationResult{
		IsValid:           true,
		RequiredApprovals: 1,
	}, nil
}

// evaluateApprovalRule evaluates if an approval rule applies to an approver for a request
func (s *approverService) evaluateApprovalRule(ctx context.Context, rule *ApprovalRule, approver, requester *User, request *AccessRequest) bool {
	// Check if rule is active
	if !rule.IsActive {
		return false
	}

	// Check request type match
	if rule.RequestType != request.RequestType {
		return false
	}

	// Evaluate based on approver configuration type
	switch rule.Approvers.Type {
	case "ROLE_BASED":
		return s.evaluateRoleBasedApproval(ctx, rule.Approvers.RoleBasedApprovers, approver, requester, request)
	case "HIERARCHY_BASED":
		return s.evaluateHierarchyBasedApproval(ctx, rule.Approvers.HierarchyApprovers, approver, requester, request)
	case "SPECIFIC_USERS":
		return s.evaluateSpecificUserApproval(rule.Approvers.SpecificApprovers, approver.ID)
	case "MIXED":
		// Evaluate multiple criteria (AND logic)
		return s.evaluateRoleBasedApproval(ctx, rule.Approvers.RoleBasedApprovers, approver, requester, request) ||
			   s.evaluateHierarchyBasedApproval(ctx, rule.Approvers.HierarchyApprovers, approver, requester, request) ||
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
func (s *approverService) evaluateRoleBasedApproval(ctx context.Context, config *RoleBasedApprovers, approver, requester *User, request *AccessRequest) bool {
	if config == nil {
		return false
	}

	// Check if requester is excluded from approving their own request
	if config.ExcludeRequester && approver.ID == requester.ID {
		return false
	}

	// TODO: Implement role checking
	// This would check if the approver has one of the required roles
	// and meets the minimum security level requirements

	return true // Placeholder
}

// evaluateHierarchyBasedApproval evaluates hierarchy-based approval rules
func (s *approverService) evaluateHierarchyBasedApproval(ctx context.Context, config *HierarchyApprovers, approver, requester *User, request *AccessRequest) bool {
	if config == nil {
		return false
	}

	// TODO: Implement hierarchy checking
	// This would check the organizational relationship between approver and requester
	
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
func (s *approverService) getApproversFromRule(ctx context.Context, rule *ApprovalRule, request *AccessRequest) ([]uuid.UUID, error) {
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
func (s *approverService) getDefaultApprovers(ctx context.Context, request *AccessRequest) ([]uuid.UUID, error) {
	// TODO: Implement default approver logic
	// This could include:
	// - Entity administrators
	// - Department managers
	// - System administrators for the tenant
	
	return []uuid.UUID{}, nil
}

// getDefaultApprovalRules returns default approval rules for a request type
func (s *approverService) getDefaultApprovalRules(entityID uuid.UUID, requestType RequestType) []*ApprovalRule {
	rules := make([]*ApprovalRule, 0)

	// Create default rules based on request type
	switch requestType {
	case RequestTypeRoleAssignment:
		rule := &ApprovalRule{
			ID:          uuid.New(),
			EntityID:    entityID,
			Name:        "Default Role Assignment Approval",
			RequestType: requestType,
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
			RequestType: requestType,
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