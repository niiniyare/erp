package abac

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/niiniyare/erp/internal/core/abac/models"
	"github.com/niiniyare/erp/internal/core/abac/repository"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/niiniyare/erp/internal/shared/types"
)

// HybridEvaluator provides RBAC-ABAC hybrid evaluation capabilities
type HybridEvaluator interface {
	// Hybrid Evaluation
	EvaluateHybrid(ctx context.Context, req *HybridEvaluationRequest) (*HybridEvaluationResult, error)

	// Role-Based Attribute Inheritance
	InheritAttributesFromRoles(ctx context.Context, req *RoleAttributeInheritanceRequest) (*RoleAttributeInheritanceResult, error)

	// Permission Elevation
	EvaluatePermissionElevation(ctx context.Context, req *PermissionElevationRequest) (*PermissionElevationResult, error)

	// Migration Support
	CreateMigrationPlan(ctx context.Context, req *MigrationPlanRequest) (*MigrationPlanResult, error)
	ExecuteMigrationStep(ctx context.Context, req *MigrationStepRequest) (*MigrationStepResult, error)

	// Compatibility
	EvaluateRBACCompatibility(ctx context.Context, req *RBACCompatibilityRequest) (*RBACCompatibilityResult, error)
}

// hybridEvaluator implements HybridEvaluator
type hybridEvaluator struct {
	policyRepo     repository.PolicyRepository
	attributeRepo  repository.AttributeRepository
	evaluationRepo repository.PolicyEvaluationRepository
	logger         logger.Logger
	metrics        metrics.MetricsProvider
	tracer         tracing.TracingService
}

// NewHybridEvaluator creates a new hybrid evaluator instance
func NewHybridEvaluator(
	policyRepo repository.PolicyRepository,
	attributeRepo repository.AttributeRepository,
	evaluationRepo repository.PolicyEvaluationRepository,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.TracingService,
) HybridEvaluator {
	return &hybridEvaluator{
		policyRepo:     policyRepo,
		attributeRepo:  attributeRepo,
		evaluationRepo: evaluationRepo,
		logger:         logger,
		metrics:        metrics,
		tracer:         tracer,
	}
}

// Hybrid Evaluation Types

type HybridEvaluationRequest struct {
	UserID         uuid.UUID               `json:"user_id" validate:"required"`
	ResourceType   string                  `json:"resource_type" validate:"required"`
	ResourceID     *uuid.UUID              `json:"resource_id,omitempty"`
	Action         string                  `json:"action" validate:"required"`
	EntityID       *uuid.UUID              `json:"entity_id,omitempty"`
	Context        map[string]interface{}  `json:"context,omitempty"`
	EvaluationMode HybridEvaluationMode    `json:"evaluation_mode"`
	Options        HybridEvaluationOptions `json:"options"`
}

type HybridEvaluationMode string

const (
	HybridModeRBACFirst  HybridEvaluationMode = "rbac_first"  // Try RBAC first, fallback to ABAC
	HybridModeABACFirst  HybridEvaluationMode = "abac_first"  // Try ABAC first, fallback to RBAC
	HybridModeStrictRBAC HybridEvaluationMode = "strict_rbac" // RBAC only with attribute enrichment
	HybridModeStrictABAC HybridEvaluationMode = "strict_abac" // ABAC only
	HybridModeUnion      HybridEvaluationMode = "union"       // Allow if either system allows
	HybridModeIntersect  HybridEvaluationMode = "intersect"   // Allow only if both systems allow
)

type HybridEvaluationOptions struct {
	IncludeRoleInfo           bool `json:"include_role_info"`
	IncludeAttributeDetails   bool `json:"include_attribute_details"`
	EnablePermissionElevation bool `json:"enable_permission_elevation"`
	CacheResults              bool `json:"cache_results"`
	TraceEvaluation           bool `json:"trace_evaluation"`
}

type HybridEvaluationResult struct {
	Decision            types.PolicyDecisionType `json:"decision"`
	DecisionSource      HybridDecisionSource     `json:"decision_source"`
	RBACResult          *RBACEvaluationResult    `json:"rbac_result,omitempty"`
	ABACResult          *ABACEvaluationResult    `json:"abac_result,omitempty"`
	CombinedAnalysis    *CombinedAnalysis        `json:"combined_analysis,omitempty"`
	PermissionElevation *PermissionElevationInfo `json:"permission_elevation,omitempty"`
	ExecutionTime       time.Duration            `json:"execution_time"`
	Timestamp           time.Time                `json:"timestamp"`
}

type HybridDecisionSource string

const (
	DecisionSourceRBAC            HybridDecisionSource = "rbac"
	DecisionSourceABAC            HybridDecisionSource = "abac"
	DecisionSourceHybridUnion     HybridDecisionSource = "hybrid_union"
	DecisionSourceHybridIntersect HybridDecisionSource = "hybrid_intersect"
	DecisionSourceElevation       HybridDecisionSource = "elevation"
	DecisionSourceFallback        HybridDecisionSource = "fallback"
)

type RBACEvaluationResult struct {
	Decision        types.PolicyDecisionType `json:"decision"`
	ApplicableRoles []RoleInfo               `json:"applicable_roles"`
	Permissions     []PermissionInfo         `json:"permissions"`
	ExecutionTime   time.Duration            `json:"execution_time"`
	CacheHit        bool                     `json:"cache_hit"`
}

type ABACEvaluationResult struct {
	Decision           types.PolicyDecisionType `json:"decision"`
	ApplicablePolicies []*models.PolicyDecision `json:"applicable_policies"`
	AttributesUsed     map[string]interface{}   `json:"attributes_used"`
	ExecutionTime      time.Duration            `json:"execution_time"`
	CacheHit           bool                     `json:"cache_hit"`
}

type RoleInfo struct {
	RoleID      uuid.UUID `json:"role_id"`
	RoleName    string    `json:"role_name"`
	Source      string    `json:"source"` // "direct", "inherited", "group"
	Priority    int32     `json:"priority"`
	Permissions []string  `json:"permissions"`
}

type PermissionInfo struct {
	Permission   string                 `json:"permission"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   *uuid.UUID             `json:"resource_id,omitempty"`
	Constraints  map[string]interface{} `json:"constraints,omitempty"`
	Source       string                 `json:"source"`
}

type CombinedAnalysis struct {
	ConflictDetected      bool                     `json:"conflict_detected"`
	ConflictDetails       []EvaluationConflict     `json:"conflict_details,omitempty"`
	RecommendedDecision   types.PolicyDecisionType `json:"recommended_decision"`
	RecommendationReason  string                   `json:"recommendation_reason"`
	RiskAssessment        HybridRiskAssessment     `json:"risk_assessment"`
	PerformanceComparison PerformanceComparison    `json:"performance_comparison"`
}

type EvaluationConflict struct {
	ConflictType  string `json:"conflict_type"` // "decision_mismatch", "permission_overlap", "attribute_conflict"
	Description   string `json:"description"`
	RBACComponent string `json:"rbac_component"`
	ABACComponent string `json:"abac_component"`
	Severity      string `json:"severity"`
	Resolution    string `json:"resolution"`
}

type HybridRiskAssessment struct {
	RiskLevel            string   `json:"risk_level"` // "low", "medium", "high", "critical"
	SecurityImplications []string `json:"security_implications"`
	ComplianceImpact     []string `json:"compliance_impact"`
	AuditRequirements    []string `json:"audit_requirements"`
}

type PerformanceComparison struct {
	RBACExecutionTime time.Duration        `json:"rbac_execution_time"`
	ABACExecutionTime time.Duration        `json:"abac_execution_time"`
	HybridOverhead    time.Duration        `json:"hybrid_overhead"`
	RecommendedMode   HybridEvaluationMode `json:"recommended_mode"`
}

func (he *hybridEvaluator) EvaluateHybrid(ctx context.Context, req *HybridEvaluationRequest) (*HybridEvaluationResult, error) {
	ctx, span := he.tracer.StartSpan(ctx, "abac.hybrid_evaluator.EvaluateHybrid",
		tracing.WithAttributes(
			tracing.StringAttribute("user_id", req.UserID.String()),
			tracing.StringAttribute("resource_type", req.ResourceType),
			tracing.StringAttribute("action", req.Action),
			tracing.StringAttribute("evaluation_mode", string(req.EvaluationMode)),
		))
	defer span.End()

	startTime := time.Now()

	he.logger.InfoContext(ctx, "Starting hybrid evaluation",
		logger.Fields{
			"user_id":         req.UserID,
			"resource_type":   req.ResourceType,
			"action":          req.Action,
			"evaluation_mode": req.EvaluationMode,
		})

	var rbacResult *RBACEvaluationResult
	var abacResult *ABACEvaluationResult
	var finalDecision types.PolicyDecisionType
	var decisionSource HybridDecisionSource
	var err error

	// Execute evaluation based on mode
	switch req.EvaluationMode {
	case HybridModeRBACFirst:
		rbacResult, abacResult, finalDecision, decisionSource, err = he.evaluateRBACFirst(ctx, req)
	case HybridModeABACFirst:
		rbacResult, abacResult, finalDecision, decisionSource, err = he.evaluateABACFirst(ctx, req)
	case HybridModeStrictRBAC:
		rbacResult, finalDecision, decisionSource, err = he.evaluateStrictRBAC(ctx, req)
	case HybridModeStrictABAC:
		abacResult, finalDecision, decisionSource, err = he.evaluateStrictABAC(ctx, req)
	case HybridModeUnion:
		rbacResult, abacResult, finalDecision, decisionSource, err = he.evaluateUnion(ctx, req)
	case HybridModeIntersect:
		rbacResult, abacResult, finalDecision, decisionSource, err = he.evaluateIntersect(ctx, req)
	default:
		return nil, errors.NewBusinessErrorWithContext(ctx, "INVALID_EVALUATION_MODE", "Invalid hybrid evaluation mode")
	}

	if err != nil {
		he.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, err
	}

	executionTime := time.Since(startTime)

	// Generate combined analysis if both systems were evaluated
	var combinedAnalysis *CombinedAnalysis
	if rbacResult != nil && abacResult != nil {
		combinedAnalysis = he.generateCombinedAnalysis(ctx, req, rbacResult, abacResult, executionTime)
	}

	// Check for permission elevation if enabled
	var permissionElevation *PermissionElevationInfo
	if req.Options.EnablePermissionElevation && finalDecision == types.PolicyDecisionDeny {
		elevationResult, err := he.checkPermissionElevation(ctx, req, rbacResult, abacResult)
		if err != nil {
			he.logger.WarnContext(ctx, "Failed to check permission elevation", logger.Fields{"error": err.Error()})
		} else if elevationResult.ElevationAvailable {
			permissionElevation = elevationResult
			finalDecision = types.PolicyDecisionAllow
			decisionSource = DecisionSourceElevation
		}
	}

	// Record metrics
	he.metrics.IncrementSuccessCount("hybrid_evaluator_evaluation")
	he.metrics.RecordGauge("hybrid_evaluation_decision_allow",
		func() float64 {
			if finalDecision == types.PolicyDecisionAllow {
				return 1
			}
			return 0
		}(),
		metrics.Fields{"mode": string(req.EvaluationMode), "source": string(decisionSource)})
	he.metrics.ObserveHistogram("hybrid_evaluation_duration_seconds", executionTime.Seconds(),
		metrics.Fields{"mode": string(req.EvaluationMode)})

	result := &HybridEvaluationResult{
		Decision:            finalDecision,
		DecisionSource:      decisionSource,
		RBACResult:          rbacResult,
		ABACResult:          abacResult,
		CombinedAnalysis:    combinedAnalysis,
		PermissionElevation: permissionElevation,
		ExecutionTime:       executionTime,
		Timestamp:           time.Now(),
	}

	he.logger.InfoContext(ctx, "Hybrid evaluation completed",
		logger.Fields{
			"user_id":         req.UserID,
			"decision":        finalDecision,
			"decision_source": decisionSource,
			"execution_time":  executionTime.Milliseconds(),
		})

	return result, nil
}

// Role-Based Attribute Inheritance

type RoleAttributeInheritanceRequest struct {
	UserID            uuid.UUID  `json:"user_id" validate:"required"`
	EntityID          *uuid.UUID `json:"entity_id,omitempty"`
	IncludeGroups     bool       `json:"include_groups"`
	IncludeTransitive bool       `json:"include_transitive"`
}

type RoleAttributeInheritanceResult struct {
	UserID              uuid.UUID                     `json:"user_id"`
	InheritedAttributes map[string]AttributeValue     `json:"inherited_attributes"`
	AttributeSources    map[string]AttributeSource    `json:"attribute_sources"`
	RoleHierarchy       []RoleHierarchyLevel          `json:"role_hierarchy"`
	ConflictResolution  []AttributeConflictResolution `json:"conflict_resolution,omitempty"`
	ExecutionTime       time.Duration                 `json:"execution_time"`
	Timestamp           time.Time                     `json:"timestamp"`
}

type AttributeValue struct {
	Value       interface{}            `json:"value"`
	DataType    string                 `json:"data_type"`
	Source      string                 `json:"source"`
	Priority    int32                  `json:"priority"`
	Constraints map[string]interface{} `json:"constraints,omitempty"`
}

type AttributeSource struct {
	SourceType      string    `json:"source_type"` // "direct_role", "inherited_role", "group", "computed"
	SourceID        uuid.UUID `json:"source_id"`
	SourceName      string    `json:"source_name"`
	InheritancePath []string  `json:"inheritance_path,omitempty"`
}

type RoleHierarchyLevel struct {
	Level      int32       `json:"level"`
	RoleID     uuid.UUID   `json:"role_id"`
	RoleName   string      `json:"role_name"`
	Attributes []string    `json:"attributes"`
	Children   []uuid.UUID `json:"children,omitempty"`
}

type AttributeConflictResolution struct {
	AttributeName     string           `json:"attribute_name"`
	ConflictingValues []AttributeValue `json:"conflicting_values"`
	ResolvedValue     AttributeValue   `json:"resolved_value"`
	ResolutionMethod  string           `json:"resolution_method"`
	ResolutionReason  string           `json:"resolution_reason"`
}

func (he *hybridEvaluator) InheritAttributesFromRoles(ctx context.Context, req *RoleAttributeInheritanceRequest) (*RoleAttributeInheritanceResult, error) {
	ctx, span := he.tracer.StartSpan(ctx, "abac.hybrid_evaluator.InheritAttributesFromRoles",
		tracing.WithAttributes(
			tracing.StringAttribute("user_id", req.UserID.String()),
			tracing.BoolAttribute("include_groups", req.IncludeGroups),
			tracing.BoolAttribute("include_transitive", req.IncludeTransitive),
		))
	defer span.End()

	startTime := time.Now()

	he.logger.InfoContext(ctx, "Starting role-based attribute inheritance",
		logger.Fields{
			"user_id":            req.UserID,
			"include_groups":     req.IncludeGroups,
			"include_transitive": req.IncludeTransitive,
		})

	// This is a simplified implementation
	// In a real scenario, you would:
	// 1. Get all roles for the user
	// 2. Build role hierarchy
	// 3. Collect attributes from each role
	// 4. Resolve conflicts using priority/precedence rules
	// 5. Return the final attribute set

	inheritedAttributes := make(map[string]AttributeValue)
	attributeSources := make(map[string]AttributeSource)
	var roleHierarchy []RoleHierarchyLevel
	var conflictResolution []AttributeConflictResolution

	// Placeholder implementation
	inheritedAttributes["department"] = AttributeValue{
		Value:    "engineering",
		DataType: "string",
		Source:   "role_employee",
		Priority: 100,
	}
	inheritedAttributes["clearance_level"] = AttributeValue{
		Value:    5,
		DataType: "number",
		Source:   "role_senior_engineer",
		Priority: 200,
	}

	attributeSources["department"] = AttributeSource{
		SourceType: "direct_role",
		SourceID:   uuid.New(),
		SourceName: "Employee",
	}

	executionTime := time.Since(startTime)

	he.metrics.IncrementSuccessCount("hybrid_evaluator_attribute_inheritance")
	he.metrics.RecordGauge("inherited_attributes_count", float64(len(inheritedAttributes)),
		metrics.Fields{"user_id": req.UserID.String()})

	result := &RoleAttributeInheritanceResult{
		UserID:              req.UserID,
		InheritedAttributes: inheritedAttributes,
		AttributeSources:    attributeSources,
		RoleHierarchy:       roleHierarchy,
		ConflictResolution:  conflictResolution,
		ExecutionTime:       executionTime,
		Timestamp:           time.Now(),
	}

	he.logger.InfoContext(ctx, "Role-based attribute inheritance completed",
		logger.Fields{
			"user_id":            req.UserID,
			"attributes_count":   len(inheritedAttributes),
			"conflicts_resolved": len(conflictResolution),
			"execution_time":     executionTime.Milliseconds(),
		})

	return result, nil
}

// Permission Elevation

type PermissionElevationRequest struct {
	UserID        uuid.UUID              `json:"user_id" validate:"required"`
	ResourceType  string                 `json:"resource_type" validate:"required"`
	ResourceID    *uuid.UUID             `json:"resource_id,omitempty"`
	Action        string                 `json:"action" validate:"required"`
	EntityID      *uuid.UUID             `json:"entity_id,omitempty"`
	Context       map[string]interface{} `json:"context,omitempty"`
	Justification string                 `json:"justification" validate:"required"`
	RequestedBy   *uuid.UUID             `json:"requested_by,omitempty"`
}

type PermissionElevationResult struct {
	ElevationID        uuid.UUID               `json:"elevation_id"`
	UserID             uuid.UUID               `json:"user_id"`
	ElevationStatus    ElevationStatus         `json:"elevation_status"`
	ElevationMethod    ElevationMethod         `json:"elevation_method"`
	GrantedPermissions []ElevatedPermission    `json:"granted_permissions"`
	RequiredApprovals  []ApprovalRequirement   `json:"required_approvals,omitempty"`
	ExpirationTime     *time.Time              `json:"expiration_time,omitempty"`
	AuditTrail         []ElevationAuditEntry   `json:"audit_trail"`
	RiskAssessment     ElevationRiskAssessment `json:"risk_assessment"`
	ExecutionTime      time.Duration           `json:"execution_time"`
	Timestamp          time.Time               `json:"timestamp"`
}

type ElevationStatus string

const (
	ElevationStatusGranted ElevationStatus = "granted"
	ElevationStatusPending ElevationStatus = "pending_approval"
	ElevationStatusDenied  ElevationStatus = "denied"
	ElevationStatusExpired ElevationStatus = "expired"
	ElevationStatusRevoked ElevationStatus = "revoked"
)

type ElevationMethod string

const (
	ElevationMethodAutomatic    ElevationMethod = "automatic"
	ElevationMethodApprovalFlow ElevationMethod = "approval_flow"
	ElevationMethodBreakGlass   ElevationMethod = "break_glass"
	ElevationMethodTemporary    ElevationMethod = "temporary"
)

type ElevatedPermission struct {
	Permission   string                 `json:"permission"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   *uuid.UUID             `json:"resource_id,omitempty"`
	Constraints  map[string]interface{} `json:"constraints,omitempty"`
	GrantedAt    time.Time              `json:"granted_at"`
	ExpiresAt    *time.Time             `json:"expires_at,omitempty"`
}

type ApprovalRequirement struct {
	ApprovalType     string     `json:"approval_type"`
	RequiredApprover *uuid.UUID `json:"required_approver,omitempty"`
	ApprovalRole     string     `json:"approval_role,omitempty"`
	Deadline         *time.Time `json:"deadline,omitempty"`
	Priority         string     `json:"priority"`
}

type ElevationAuditEntry struct {
	Action       string                 `json:"action"`
	ActorID      *uuid.UUID             `json:"actor_id,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
	Details      map[string]interface{} `json:"details,omitempty"`
	SystemAction bool                   `json:"system_action"`
}

type ElevationRiskAssessment struct {
	RiskScore          float64  `json:"risk_score"`
	RiskLevel          string   `json:"risk_level"`
	RiskFactors        []string `json:"risk_factors"`
	MitigationMeasures []string `json:"mitigation_measures"`
	MonitoringRequired bool     `json:"monitoring_required"`
}

type PermissionElevationInfo struct {
	ElevationAvailable bool                       `json:"elevation_available"`
	ElevationMethods   []AvailableElevationMethod `json:"elevation_methods"`
	Recommendations    []string                   `json:"recommendations"`
	RiskWarnings       []string                   `json:"risk_warnings"`
}

type AvailableElevationMethod struct {
	Method            ElevationMethod       `json:"method"`
	RequiredApprovals []ApprovalRequirement `json:"required_approvals,omitempty"`
	EstimatedTime     *time.Duration        `json:"estimated_time,omitempty"`
	RiskLevel         string                `json:"risk_level"`
}

func (he *hybridEvaluator) EvaluatePermissionElevation(ctx context.Context, req *PermissionElevationRequest) (*PermissionElevationResult, error) {
	ctx, span := he.tracer.StartSpan(ctx, "abac.hybrid_evaluator.EvaluatePermissionElevation",
		tracing.WithAttributes(
			tracing.StringAttribute("user_id", req.UserID.String()),
			tracing.StringAttribute("resource_type", req.ResourceType),
			tracing.StringAttribute("action", req.Action),
		))
	defer span.End()

	startTime := time.Now()

	he.logger.InfoContext(ctx, "Evaluating permission elevation request",
		logger.Fields{
			"user_id":       req.UserID,
			"resource_type": req.ResourceType,
			"action":        req.Action,
			"justification": req.Justification,
		})

	// This is a simplified implementation
	// In a real scenario, you would:
	// 1. Analyze the requested permission
	// 2. Check user's current permissions and roles
	// 3. Determine available elevation methods
	// 4. Assess risk and security implications
	// 5. Apply elevation policies
	// 6. Create approval workflows if needed

	elevationID := uuid.New()
	executionTime := time.Since(startTime)

	// Simplified risk assessment
	riskAssessment := ElevationRiskAssessment{
		RiskScore:          0.3,
		RiskLevel:          "medium",
		RiskFactors:        []string{"Sensitive resource access", "Outside normal hours"},
		MitigationMeasures: []string{"Time-limited access", "Enhanced monitoring"},
		MonitoringRequired: true,
	}

	result := &PermissionElevationResult{
		ElevationID:     elevationID,
		UserID:          req.UserID,
		ElevationStatus: ElevationStatusGranted,
		ElevationMethod: ElevationMethodTemporary,
		GrantedPermissions: []ElevatedPermission{
			{
				Permission:   req.Action,
				ResourceType: req.ResourceType,
				ResourceID:   req.ResourceID,
				GrantedAt:    time.Now(),
				ExpiresAt:    func() *time.Time { t := time.Now().Add(1 * time.Hour); return &t }(),
			},
		},
		AuditTrail: []ElevationAuditEntry{
			{
				Action:       "elevation_granted",
				ActorID:      req.RequestedBy,
				Timestamp:    time.Now(),
				SystemAction: true,
			},
		},
		RiskAssessment: riskAssessment,
		ExecutionTime:  executionTime,
		Timestamp:      time.Now(),
	}

	he.metrics.IncrementSuccessCount("hybrid_evaluator_permission_elevation")
	he.metrics.RecordGauge("permission_elevation_risk_score", riskAssessment.RiskScore,
		metrics.Fields{"user_id": req.UserID.String(), "risk_level": riskAssessment.RiskLevel})

	he.logger.InfoContext(ctx, "Permission elevation evaluated",
		logger.Fields{
			"elevation_id":     elevationID,
			"user_id":          req.UserID,
			"elevation_status": result.ElevationStatus,
			"risk_level":       riskAssessment.RiskLevel,
			"execution_time":   executionTime.Milliseconds(),
		})

	return result, nil
}

// Helper methods for hybrid evaluation modes

func (he *hybridEvaluator) evaluateRBACFirst(ctx context.Context, req *HybridEvaluationRequest) (*RBACEvaluationResult, *ABACEvaluationResult, types.PolicyDecisionType, HybridDecisionSource, error) {
	// Try RBAC first
	rbacResult, err := he.evaluateRBAC(ctx, req)
	if err != nil {
		return nil, nil, types.PolicyDecisionDeny, DecisionSourceRBAC, err
	}

	if rbacResult.Decision == types.PolicyDecisionAllow {
		return rbacResult, nil, types.PolicyDecisionAllow, DecisionSourceRBAC, nil
	}

	// Fallback to ABAC
	abacResult, err := he.evaluateABAC(ctx, req)
	if err != nil {
		return rbacResult, nil, types.PolicyDecisionDeny, DecisionSourceFallback, err
	}

	return rbacResult, abacResult, abacResult.Decision, DecisionSourceFallback, nil
}

func (he *hybridEvaluator) evaluateABACFirst(ctx context.Context, req *HybridEvaluationRequest) (*RBACEvaluationResult, *ABACEvaluationResult, types.PolicyDecisionType, HybridDecisionSource, error) {
	// Try ABAC first
	abacResult, err := he.evaluateABAC(ctx, req)
	if err != nil {
		return nil, nil, types.PolicyDecisionDeny, DecisionSourceABAC, err
	}

	if abacResult.Decision == types.PolicyDecisionAllow {
		return nil, abacResult, types.PolicyDecisionAllow, DecisionSourceABAC, nil
	}

	// Fallback to RBAC
	rbacResult, err := he.evaluateRBAC(ctx, req)
	if err != nil {
		return nil, abacResult, types.PolicyDecisionDeny, DecisionSourceFallback, err
	}

	return rbacResult, abacResult, rbacResult.Decision, DecisionSourceFallback, nil
}

func (he *hybridEvaluator) evaluateStrictRBAC(ctx context.Context, req *HybridEvaluationRequest) (*RBACEvaluationResult, types.PolicyDecisionType, HybridDecisionSource, error) {
	rbacResult, err := he.evaluateRBAC(ctx, req)
	if err != nil {
		return nil, types.PolicyDecisionDeny, DecisionSourceRBAC, err
	}
	return rbacResult, rbacResult.Decision, DecisionSourceRBAC, nil
}

func (he *hybridEvaluator) evaluateStrictABAC(ctx context.Context, req *HybridEvaluationRequest) (*ABACEvaluationResult, types.PolicyDecisionType, HybridDecisionSource, error) {
	abacResult, err := he.evaluateABAC(ctx, req)
	if err != nil {
		return nil, types.PolicyDecisionDeny, DecisionSourceABAC, err
	}
	return abacResult, abacResult.Decision, DecisionSourceABAC, nil
}

func (he *hybridEvaluator) evaluateUnion(ctx context.Context, req *HybridEvaluationRequest) (*RBACEvaluationResult, *ABACEvaluationResult, types.PolicyDecisionType, HybridDecisionSource, error) {
	rbacResult, rbacErr := he.evaluateRBAC(ctx, req)
	abacResult, abacErr := he.evaluateABAC(ctx, req)

	if rbacErr != nil && abacErr != nil {
		return rbacResult, abacResult, types.PolicyDecisionDeny, DecisionSourceHybridUnion, rbacErr
	}

	finalDecision := types.PolicyDecisionDeny
	if (rbacResult != nil && rbacResult.Decision == types.PolicyDecisionAllow) ||
		(abacResult != nil && abacResult.Decision == types.PolicyDecisionAllow) {
		finalDecision = types.PolicyDecisionAllow
	}

	return rbacResult, abacResult, finalDecision, DecisionSourceHybridUnion, nil
}

func (he *hybridEvaluator) evaluateIntersect(ctx context.Context, req *HybridEvaluationRequest) (*RBACEvaluationResult, *ABACEvaluationResult, types.PolicyDecisionType, HybridDecisionSource, error) {
	rbacResult, rbacErr := he.evaluateRBAC(ctx, req)
	abacResult, abacErr := he.evaluateABAC(ctx, req)

	if rbacErr != nil || abacErr != nil {
		return rbacResult, abacResult, types.PolicyDecisionDeny, DecisionSourceHybridIntersect, fmt.Errorf("evaluation error")
	}

	finalDecision := types.PolicyDecisionDeny
	if rbacResult.Decision == types.PolicyDecisionAllow && abacResult.Decision == types.PolicyDecisionAllow {
		finalDecision = types.PolicyDecisionAllow
	}

	return rbacResult, abacResult, finalDecision, DecisionSourceHybridIntersect, nil
}

// Simplified RBAC and ABAC evaluation methods (placeholders)
func (he *hybridEvaluator) evaluateRBAC(ctx context.Context, req *HybridEvaluationRequest) (*RBACEvaluationResult, error) {
	// Simplified RBAC evaluation
	return &RBACEvaluationResult{
		Decision: types.PolicyDecisionAllow,
		ApplicableRoles: []RoleInfo{
			{
				RoleID:   uuid.New(),
				RoleName: "user",
				Source:   "direct",
				Priority: 100,
			},
		},
		ExecutionTime: 10 * time.Millisecond,
		CacheHit:      false,
	}, nil
}

func (he *hybridEvaluator) evaluateABAC(ctx context.Context, req *HybridEvaluationRequest) (*ABACEvaluationResult, error) {
	// Simplified ABAC evaluation
	return &ABACEvaluationResult{
		Decision:           types.PolicyDecisionAllow,
		ApplicablePolicies: []*models.PolicyDecision{},
		AttributesUsed: map[string]interface{}{
			"user.department": "engineering",
			"resource.type":   req.ResourceType,
		},
		ExecutionTime: 25 * time.Millisecond,
		CacheHit:      false,
	}, nil
}

func (he *hybridEvaluator) generateCombinedAnalysis(ctx context.Context, req *HybridEvaluationRequest, rbacResult *RBACEvaluationResult, abacResult *ABACEvaluationResult, totalTime time.Duration) *CombinedAnalysis {
	// Detect conflicts
	conflictDetected := rbacResult.Decision != abacResult.Decision
	var conflicts []EvaluationConflict

	if conflictDetected {
		conflicts = append(conflicts, EvaluationConflict{
			ConflictType:  "decision_mismatch",
			Description:   fmt.Sprintf("RBAC decision: %s, ABAC decision: %s", rbacResult.Decision, abacResult.Decision),
			RBACComponent: "primary_evaluation",
			ABACComponent: "primary_evaluation",
			Severity:      "high",
			Resolution:    "Review policy configuration and role assignments",
		})
	}

	// Risk assessment
	riskLevel := "low"
	if conflictDetected {
		riskLevel = "medium"
	}

	return &CombinedAnalysis{
		ConflictDetected:     conflictDetected,
		ConflictDetails:      conflicts,
		RecommendedDecision:  abacResult.Decision, // ABAC takes precedence in conflicts
		RecommendationReason: "ABAC provides more granular control",
		RiskAssessment: HybridRiskAssessment{
			RiskLevel:            riskLevel,
			SecurityImplications: []string{"Decision conflict may indicate policy misconfiguration"},
			ComplianceImpact:     []string{"Ensure audit trail captures both evaluations"},
		},
		PerformanceComparison: PerformanceComparison{
			RBACExecutionTime: rbacResult.ExecutionTime,
			ABACExecutionTime: abacResult.ExecutionTime,
			HybridOverhead:    totalTime - rbacResult.ExecutionTime - abacResult.ExecutionTime,
			RecommendedMode:   HybridModeABACFirst,
		},
	}
}

func (he *hybridEvaluator) checkPermissionElevation(ctx context.Context, req *HybridEvaluationRequest, rbacResult *RBACEvaluationResult, abacResult *ABACEvaluationResult) (*PermissionElevationInfo, error) {
	// Simplified elevation check
	return &PermissionElevationInfo{
		ElevationAvailable: true,
		ElevationMethods: []AvailableElevationMethod{
			{
				Method:        ElevationMethodTemporary,
				RiskLevel:     "medium",
				EstimatedTime: func() *time.Duration { d := 5 * time.Minute; return &d }(),
			},
		},
		Recommendations: []string{"Consider requesting temporary access with justification"},
		RiskWarnings:    []string{"Elevated access will be logged and monitored"},
	}, nil
}
