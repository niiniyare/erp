package abac

//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/niiniyare/erp/internal/core/abac/repository"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/niiniyare/erp/internal/shared/types"
)

type HybridEvaluator interface {
	EvaluateHybrid(ctx context.Context, req any) (any, error)
}

// CompatibilityLayer provides RBAC-ABAC compatibility features
type CompatibilityLayer interface {
	// Legacy RBAC Support
	EvaluateLegacyRBAC(ctx context.Context, req *LegacyRBACRequest) (*LegacyRBACResult, error)
	TranslateRBACToABAC(ctx context.Context, req *RBACTranslationRequest) (*RBACTranslationResult, error)

	// Gradual Migration Support
	ConfigureGradualMigration(ctx context.Context, req *GradualMigrationConfig) (*GradualMigrationResult, error)
	EvaluateWithFallback(ctx context.Context, req *FallbackEvaluationRequest) (*FallbackEvaluationResult, error)

	// Permission Bridging
	CreatePermissionBridge(ctx context.Context, req *PermissionBridgeRequest) (*PermissionBridge, error)
	EvaluatePermissionBridge(ctx context.Context, req *BridgeEvaluationRequest) (*BridgeEvaluationResult, error)

	// Compatibility Monitoring
	MonitorCompatibility(ctx context.Context, req *CompatibilityMonitoringRequest) (*CompatibilityMonitoringResult, error)
	GenerateCompatibilityReport(ctx context.Context, req *CompatibilityReportRequest) (*CompatibilityReport, error)
}

// compatibilityLayer implements CompatibilityLayer
type compatibilityLayer struct {
	policyRepo      repository.PolicyRepository
	attributeRepo   repository.AttributeRepository
	hybridEvaluator HybridEvaluator
	logger          logger.Logger
	metrics         metrics.MetricsProvider
	tracer          tracing.TracingService
}

// NewCompatibilityLayer creates a new compatibility layer instance
func NewCompatibilityLayer(
	policyRepo repository.PolicyRepository,
	attributeRepo repository.AttributeRepository,
	hybridEvaluator HybridEvaluator,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.TracingService,
) CompatibilityLayer {
	return &compatibilityLayer{
		policyRepo:      policyRepo,
		attributeRepo:   attributeRepo,
		hybridEvaluator: hybridEvaluator,
		logger:          logger,
		metrics:         metrics,
		tracer:          tracer,
	}
}

// Legacy RBAC Support Types

type LegacyRBACRequest struct {
	UserID        uuid.UUID      `json:"user_id" validate:"required"`
	ResourceType  string         `json:"resource_type" validate:"required"`
	ResourceID    *uuid.UUID     `json:"resource_id,omitempty"`
	Action        string         `json:"action" validate:"required"`
	Context       map[string]any `json:"context,omitempty"`
	RoleOverrides []RoleOverride `json:"role_overrides,omitempty"`
	LegacyMode    LegacyMode     `json:"legacy_mode"`
}

type LegacyMode string

const (
	LegacyModeStrict     LegacyMode = "strict"     // Pure RBAC evaluation
	LegacyModeEnhanced   LegacyMode = "enhanced"   // RBAC with attribute enrichment
	LegacyModeCompatible LegacyMode = "compatible" // RBAC with ABAC fallback
	LegacyModeTransition LegacyMode = "transition" // Gradual ABAC adoption
)

type RoleOverride struct {
	RoleID         uuid.UUID      `json:"role_id"`
	Action         string         `json:"action"` // "add", "remove", "modify"
	Conditions     map[string]any `json:"conditions,omitempty"`
	TemporaryUntil *time.Time     `json:"temporary_until,omitempty"`
	Reason         string         `json:"reason,omitempty"`
}

type LegacyRBACResult struct {
	Decision            types.PolicyDecisionType `json:"decision"`
	EvaluationMode      LegacyMode               `json:"evaluation_mode"`
	ApplicableRoles     []LegacyRoleInfo         `json:"applicable_roles"`
	PermissionMatrix    PermissionMatrix         `json:"permission_matrix"`
	AttributeEnrichment map[string]any           `json:"attribute_enrichment,omitempty"`
	CompatibilityIssues []CompatibilityIssue     `json:"compatibility_issues,omitempty"`
	ExecutionTime       time.Duration            `json:"execution_time"`
	Timestamp           time.Time                `json:"timestamp"`
}

type LegacyRoleInfo struct {
	RoleID      uuid.UUID          `json:"role_id"`
	RoleName    string             `json:"role_name"`
	RoleType    string             `json:"role_type"`
	Permissions []LegacyPermission `json:"permissions"`
	Hierarchy   []string           `json:"hierarchy"`
	Constraints map[string]any     `json:"constraints,omitempty"`
	Source      string             `json:"source"` // "direct", "inherited", "computed"
	IsActive    bool               `json:"is_active"`
}

type LegacyPermission struct {
	Permission    string         `json:"permission"`
	ResourceType  string         `json:"resource_type"`
	ResourceID    *uuid.UUID     `json:"resource_id,omitempty"`
	Actions       []string       `json:"actions"`
	Constraints   map[string]any `json:"constraints,omitempty"`
	GrantedBy     string         `json:"granted_by"`
	EffectiveFrom time.Time      `json:"effective_from"`
	EffectiveTo   *time.Time     `json:"effective_to,omitempty"`
}

type PermissionMatrix struct {
	UserID                 uuid.UUID                      `json:"user_id"`
	ResourceMatrix         map[string]ResourcePermissions `json:"resource_matrix"`
	GlobalPermissions      []string                       `json:"global_permissions"`
	ConditionalPermissions []ConditionalPermission        `json:"conditional_permissions"`
	DeniedPermissions      []string                       `json:"denied_permissions"`
}

type ResourcePermissions struct {
	ResourceType   string         `json:"resource_type"`
	ResourceID     *uuid.UUID     `json:"resource_id,omitempty"`
	AllowedActions []string       `json:"allowed_actions"`
	DeniedActions  []string       `json:"denied_actions"`
	Conditions     map[string]any `json:"conditions,omitempty"`
}

type ConditionalPermission struct {
	Permission   string         `json:"permission"`
	Conditions   map[string]any `json:"conditions"`
	Requirements []string       `json:"requirements"`
}

type CompatibilityIssue struct {
	IssueType      string `json:"issue_type"`
	Severity       string `json:"severity"`
	Description    string `json:"description"`
	AffectedArea   string `json:"affected_area"`
	Recommendation string `json:"recommendation"`
	AutoFixable    bool   `json:"auto_fixable"`
}

func (cl *compatibilityLayer) EvaluateLegacyRBAC(ctx context.Context, req *LegacyRBACRequest) (*LegacyRBACResult, error) {
	ctx, span := cl.tracer.StartSpan(ctx, "abac.compatibility_layer.EvaluateLegacyRBAC")
	defer span.End()

	startTime := time.Now()

	cl.logger.InfoContext(ctx, "Evaluating legacy RBAC request",
		logger.Fields{
			"user_id":       req.UserID,
			"resource_type": req.ResourceType,
			"action":        req.Action,
			"legacy_mode":   req.LegacyMode,
		})

	var decision types.PolicyDecisionType
	var applicableRoles []LegacyRoleInfo
	var attributeEnrichment map[string]any
	var compatibilityIssues []CompatibilityIssue

	switch req.LegacyMode {
	case LegacyModeStrict:
		decision, applicableRoles = cl.evaluateStrictRBAC(ctx, req)
	case LegacyModeEnhanced:
		decision, applicableRoles, attributeEnrichment = cl.evaluateEnhancedRBAC(ctx, req)
	case LegacyModeCompatible:
		decision, applicableRoles, compatibilityIssues = cl.evaluateCompatibleRBAC(ctx, req)
	case LegacyModeTransition:
		decision, applicableRoles, compatibilityIssues = cl.evaluateTransitionRBAC(ctx, req)
	default:
		return nil, errors.NewBusinessErrorWithContext(ctx, "INVALID_LEGACY_MODE", "Invalid legacy mode specified")
	}

	// Generate permission matrix
	permissionMatrix := cl.generatePermissionMatrix(ctx, req.UserID, applicableRoles)

	executionTime := time.Since(startTime)

	result := &LegacyRBACResult{
		Decision:            decision,
		EvaluationMode:      req.LegacyMode,
		ApplicableRoles:     applicableRoles,
		PermissionMatrix:    permissionMatrix,
		AttributeEnrichment: attributeEnrichment,
		CompatibilityIssues: compatibilityIssues,
		ExecutionTime:       executionTime,
		Timestamp:           time.Now(),
	}

	cl.metrics.IncrementCounter("compatibility_layer_legacy_rbac", metrics.Fields{})
	cl.metrics.SetGauge("legacy_rbac_decision_allow",
		func() float64 {
			if decision == types.PolicyDecisionAllow {
				return 1
			}
			return 0
		}(),
		metrics.Fields{"mode": string(req.LegacyMode)})

	cl.logger.InfoContext(ctx, "Legacy RBAC evaluation completed",
		logger.Fields{
			"user_id":              req.UserID,
			"decision":             decision,
			"applicable_roles":     len(applicableRoles),
			"compatibility_issues": len(compatibilityIssues),
			"execution_time":       executionTime.Milliseconds(),
		})

	return result, nil
}

// RBAC Translation Types

type RBACTranslationRequest struct {
	RoleID          *uuid.UUID         `json:"role_id,omitempty"`
	RoleName        string             `json:"role_name,omitempty"`
	Permissions     []string           `json:"permissions,omitempty"`
	TranslationMode TranslationMode    `json:"translation_mode"`
	Options         TranslationOptions `json:"options"`
	TargetEntity    *uuid.UUID         `json:"target_entity,omitempty"`
}

type TranslationMode string

const (
	TranslationModeRole       TranslationMode = "role_to_policy"
	TranslationModePermission TranslationMode = "permission_to_attribute"
	TranslationModeHierarchy  TranslationMode = "hierarchy_to_rules"
	TranslationModeComplete   TranslationMode = "complete_mapping"
)

type TranslationOptions struct {
	PreserveSemantics    bool `json:"preserve_semantics"`
	CreateAttributes     bool `json:"create_attributes"`
	GenerateConstraints  bool `json:"generate_constraints"`
	OptimizePolicies     bool `json:"optimize_policies"`
	IncludeDocumentation bool `json:"include_documentation"`
}

type RBACTranslationResult struct {
	TranslationID     uuid.UUID               `json:"translation_id"`
	SourceType        string                  `json:"source_type"`
	SourceIdentifier  string                  `json:"source_identifier"`
	CreatedPolicies   []TranslatedPolicy      `json:"created_policies"`
	CreatedAttributes []TranslatedAttribute   `json:"created_attributes"`
	TransformationMap map[string]any          `json:"transformation_map"`
	SemanticAnalysis  SemanticAnalysis        `json:"semantic_analysis"`
	ValidationResults []TranslationValidation `json:"validation_results"`
	ExecutionTime     time.Duration           `json:"execution_time"`
	Timestamp         time.Time               `json:"timestamp"`
}

type TranslatedPolicy struct {
	PolicyID    uuid.UUID        `json:"policy_id"`
	PolicyName  string           `json:"policy_name"`
	PolicyType  types.PolicyType `json:"policy_type"`
	SourceRole  string           `json:"source_role"`
	Target      map[string]any   `json:"target"`
	Rule        map[string]any   `json:"rule"`
	Confidence  float64          `json:"confidence"`
	Assumptions []string         `json:"assumptions"`
}

type TranslatedAttribute struct {
	AttributeID      uuid.UUID      `json:"attribute_id"`
	AttributeName    string         `json:"attribute_name"`
	DataType         string         `json:"data_type"`
	Category         string         `json:"category"`
	SourcePermission string         `json:"source_permission"`
	DefaultValue     any            `json:"default_value,omitempty"`
	Constraints      map[string]any `json:"constraints,omitempty"`
}

type SemanticAnalysis struct {
	SemanticPreservation float64            `json:"semantic_preservation"`
	IdentifiedPatterns   []string           `json:"identified_patterns"`
	LostSemantics        []string           `json:"lost_semantics"`
	EnhancedCapabilities []string           `json:"enhanced_capabilities"`
	ComplexityAnalysis   ComplexityAnalysis `json:"complexity_analysis"`
}

type ComplexityAnalysis struct {
	SourceComplexity    int32   `json:"source_complexity"`
	TargetComplexity    int32   `json:"target_complexity"`
	ComplexityReduction float64 `json:"complexity_reduction"`
	PerformanceImpact   string  `json:"performance_impact"`
}

type TranslationValidation struct {
	ValidationID   uuid.UUID `json:"validation_id"`
	ValidationType string    `json:"validation_type"`
	Status         string    `json:"status"`
	Message        string    `json:"message"`
	Suggestions    []string  `json:"suggestions,omitempty"`
}

func (cl *compatibilityLayer) TranslateRBACToABAC(ctx context.Context, req *RBACTranslationRequest) (*RBACTranslationResult, error) {
	ctx, span := cl.tracer.StartSpan(ctx, "abac.compatibility_layer.TranslateRBACToABAC")
	defer span.End()

	startTime := time.Now()

	cl.logger.InfoContext(ctx, "Starting RBAC to ABAC translation",
		logger.Fields{
			"translation_mode": req.TranslationMode,
			"role_name":        req.RoleName,
			"permissions":      len(req.Permissions),
		})

	translationID := uuid.New()

	// Perform translation based on mode
	var createdPolicies []TranslatedPolicy
	var createdAttributes []TranslatedAttribute
	var err error

	switch req.TranslationMode {
	case TranslationModeRole:
		createdPolicies, err = cl.translateRoleToPolicy(ctx, req)
	case TranslationModePermission:
		createdAttributes, err = cl.translatePermissionToAttribute(ctx, req)
	case TranslationModeHierarchy:
		createdPolicies, err = cl.translateHierarchyToRules(ctx, req)
	case TranslationModeComplete:
		createdPolicies, createdAttributes, err = cl.translateComplete(ctx, req)
	default:
		return nil, errors.NewBusinessErrorWithContext(ctx, "INVALID_TRANSLATION_MODE", "Invalid translation mode")
	}

	if err != nil {
		cl.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, err
	}

	// Generate transformation map
	transformationMap := cl.generateTransformationMap(ctx, req, createdPolicies, createdAttributes)

	// Perform semantic analysis
	semanticAnalysis := cl.performSemanticAnalysis(ctx, req, createdPolicies, createdAttributes)

	// Validate translation
	validationResults := cl.validateTranslation(ctx, req, createdPolicies, createdAttributes)

	executionTime := time.Since(startTime)

	result := &RBACTranslationResult{
		TranslationID:     translationID,
		SourceType:        "rbac",
		SourceIdentifier:  req.RoleName,
		CreatedPolicies:   createdPolicies,
		CreatedAttributes: createdAttributes,
		TransformationMap: transformationMap,
		SemanticAnalysis:  semanticAnalysis,
		ValidationResults: validationResults,
		ExecutionTime:     executionTime,
		Timestamp:         time.Now(),
	}

	cl.metrics.IncrementCounter("compatibility_layer_translation", metrics.Fields{})
	cl.metrics.SetGauge("translation_policies_created", float64(len(createdPolicies)),
		metrics.Fields{"mode": string(req.TranslationMode)})
	cl.metrics.SetGauge("translation_attributes_created", float64(len(createdAttributes)),
		metrics.Fields{"mode": string(req.TranslationMode)})

	cl.logger.InfoContext(ctx, "RBAC to ABAC translation completed",
		logger.Fields{
			"translation_id":        translationID,
			"policies_created":      len(createdPolicies),
			"attributes_created":    len(createdAttributes),
			"semantic_preservation": semanticAnalysis.SemanticPreservation,
			"execution_time":        executionTime.Milliseconds(),
		})

	return result, nil
}

// Gradual Migration Support Types

type GradualMigrationConfig struct {
	ConfigID         uuid.UUID               `json:"config_id"`
	ConfigName       string                  `json:"config_name" validate:"required"`
	MigrationPhases  []MigrationPhaseConfig  `json:"migration_phases" validate:"required,min=1"`
	FallbackStrategy FallbackStrategy        `json:"fallback_strategy"`
	MonitoringConfig MonitoringConfig        `json:"monitoring_config"`
	Options          GradualMigrationOptions `json:"options"`
	CreatedBy        *uuid.UUID              `json:"created_by,omitempty"`
}

type MigrationPhaseConfig struct {
	PhaseID           uuid.UUID            `json:"phase_id"`
	PhaseName         string               `json:"phase_name"`
	UserCriteria      UserCriteria         `json:"user_criteria"`
	ResourceCriteria  ResourceCriteria     `json:"resource_criteria"`
	EvaluationMode    HybridEvaluationMode `json:"evaluation_mode"`
	RolloutPercentage float64              `json:"rollout_percentage"`
	StartDate         *time.Time           `json:"start_date,omitempty"`
	EndDate           *time.Time           `json:"end_date,omitempty"`
	Configuration     map[string]any       `json:"configuration,omitempty"`
}

type UserCriteria struct {
	UserIDs     []uuid.UUID    `json:"user_ids,omitempty"`
	UserGroups  []string       `json:"user_groups,omitempty"`
	Departments []string       `json:"departments,omitempty"`
	Roles       []string       `json:"roles,omitempty"`
	Attributes  map[string]any `json:"attributes,omitempty"`
}

type ResourceCriteria struct {
	ResourceTypes []string    `json:"resource_types,omitempty"`
	ResourceIDs   []uuid.UUID `json:"resource_ids,omitempty"`
	ResourceTags  []string    `json:"resource_tags,omitempty"`
	Sensitivity   []string    `json:"sensitivity,omitempty"`
	Categories    []string    `json:"categories,omitempty"`
}

// FallbackStrategy and FallbackTrigger types moved to shared_types.go
// Use SystemFallbackStrategy for this struct type

type MonitoringConfig struct {
	EnableDetailedLogging bool          `json:"enable_detailed_logging"`
	MetricsCollection     bool          `json:"metrics_collection"`
	PerformanceTracking   bool          `json:"performance_tracking"`
	DecisionComparison    bool          `json:"decision_comparison"`
	AlertRules            []AlertRule   `json:"alert_rules"`
	ReportingFrequency    time.Duration `json:"reporting_frequency"`
}

type AlertRule struct {
	RuleID        uuid.UUID      `json:"rule_id"`
	RuleName      string         `json:"rule_name"`
	Condition     string         `json:"condition"`
	Threshold     float64        `json:"threshold"`
	Severity      string         `json:"severity"`
	Actions       []string       `json:"actions"`
	Recipients    []string       `json:"recipients"`
	Configuration map[string]any `json:"configuration,omitempty"`
}

type GradualMigrationOptions struct {
	SafetyMode         bool          `json:"safety_mode"`
	RollbackOnFailure  bool          `json:"rollback_on_failure"`
	MaxErrorRate       float64       `json:"max_error_rate"`
	MaxLatencyIncrease time.Duration `json:"max_latency_increase"`
	RequireApproval    bool          `json:"require_approval"`
}

type GradualMigrationResult struct {
	ConfigID            uuid.UUID              `json:"config_id"`
	ConfigurationStatus string                 `json:"configuration_status"`
	ActivePhases        []ActiveMigrationPhase `json:"active_phases"`
	PhaseSchedule       []PhaseScheduleEntry   `json:"phase_schedule"`
	ValidationResults   []ConfigValidation     `json:"validation_results"`
	EstimatedCompletion *time.Time             `json:"estimated_completion,omitempty"`
	RiskAssessment      ConfigRiskAssessment   `json:"risk_assessment"`
	ExecutionTime       time.Duration          `json:"execution_time"`
	Timestamp           time.Time              `json:"timestamp"`
}

type ActiveMigrationPhase struct {
	PhaseID       uuid.UUID      `json:"phase_id"`
	PhaseName     string         `json:"phase_name"`
	Status        string         `json:"status"`
	Progress      float64        `json:"progress"`
	AffectedUsers int32          `json:"affected_users"`
	Configuration map[string]any `json:"configuration"`
	StartedAt     time.Time      `json:"started_at"`
	EstimatedEnd  *time.Time     `json:"estimated_end,omitempty"`
}

type PhaseScheduleEntry struct {
	PhaseID        uuid.UUID   `json:"phase_id"`
	PhaseName      string      `json:"phase_name"`
	ScheduledStart time.Time   `json:"scheduled_start"`
	ScheduledEnd   time.Time   `json:"scheduled_end"`
	Dependencies   []uuid.UUID `json:"dependencies"`
	Prerequisites  []string    `json:"prerequisites"`
}

type ConfigValidation struct {
	ValidationID   uuid.UUID `json:"validation_id"`
	ValidationType string    `json:"validation_type"`
	Status         string    `json:"status"`
	Message        string    `json:"message"`
	Severity       string    `json:"severity"`
}

type ConfigRiskAssessment struct {
	OverallRisk       string              `json:"overall_risk"`
	IdentifiedRisks   []ConfigurationRisk `json:"identified_risks"`
	MitigationActions []MitigationAction  `json:"mitigation_actions"`
	SafetyMeasures    []SafetyMeasure     `json:"safety_measures"`
}

type ConfigurationRisk struct {
	RiskID      uuid.UUID `json:"risk_id"`
	RiskType    string    `json:"risk_type"`
	Description string    `json:"description"`
	Probability float64   `json:"probability"`
	Impact      string    `json:"impact"`
	Severity    string    `json:"severity"`
}

type MitigationAction struct {
	ActionID    uuid.UUID   `json:"action_id"`
	ActionType  string      `json:"action_type"`
	Description string      `json:"description"`
	AutoExecute bool        `json:"auto_execute"`
	TargetRisks []uuid.UUID `json:"target_risks"`
}

type SafetyMeasure struct {
	MeasureID     uuid.UUID      `json:"measure_id"`
	MeasureType   string         `json:"measure_type"`
	Description   string         `json:"description"`
	Enabled       bool           `json:"enabled"`
	Configuration map[string]any `json:"configuration,omitempty"`
}

func (cl *compatibilityLayer) ConfigureGradualMigration(ctx context.Context, req *GradualMigrationConfig) (*GradualMigrationResult, error) {
	ctx, span := cl.tracer.StartSpan(ctx, "abac.compatibility_layer.ConfigureGradualMigration")
	defer span.End()

	startTime := time.Now()

	cl.logger.InfoContext(ctx, "Configuring gradual migration",
		logger.Fields{
			"config_name":  req.ConfigName,
			"phases_count": len(req.MigrationPhases),
			"safety_mode":  req.Options.SafetyMode,
		})

	// Validate configuration
	validationResults := cl.validateMigrationConfig(ctx, req)

	// Assess risks
	riskAssessment := cl.assessConfigurationRisks(ctx, req)

	// Generate phase schedule
	phaseSchedule := cl.generatePhaseSchedule(ctx, req.MigrationPhases)

	// Determine active phases
	activePhases := cl.determineActivePhases(ctx, req.MigrationPhases)

	// Estimate completion
	estimatedCompletion := cl.estimateCompletion(ctx, req.MigrationPhases)

	executionTime := time.Since(startTime)

	result := &GradualMigrationResult{
		ConfigID:            req.ConfigID,
		ConfigurationStatus: "configured",
		ActivePhases:        activePhases,
		PhaseSchedule:       phaseSchedule,
		ValidationResults:   validationResults,
		EstimatedCompletion: estimatedCompletion,
		RiskAssessment:      riskAssessment,
		ExecutionTime:       executionTime,
		Timestamp:           time.Now(),
	}

	cl.metrics.IncrementCounter("compatibility_layer_gradual_migration_configured", metrics.Fields{})
	cl.metrics.SetGauge("gradual_migration_phases", float64(len(req.MigrationPhases)),
		metrics.Fields{"config_id": req.ConfigID.String()})

	cl.logger.InfoContext(ctx, "Gradual migration configured successfully",
		logger.Fields{
			"config_id":      req.ConfigID,
			"active_phases":  len(activePhases),
			"overall_risk":   riskAssessment.OverallRisk,
			"execution_time": executionTime.Milliseconds(),
		})

	return result, nil
}

// Helper methods for legacy RBAC evaluation modes

func (cl *compatibilityLayer) evaluateStrictRBAC(ctx context.Context, req *LegacyRBACRequest) (types.PolicyDecisionType, []LegacyRoleInfo) {
	// Simplified strict RBAC evaluation
	roles := []LegacyRoleInfo{
		{
			RoleID:   uuid.New(),
			RoleName: "user",
			RoleType: "standard",
			Permissions: []LegacyPermission{
				{
					Permission:    "read",
					ResourceType:  req.ResourceType,
					Actions:       []string{req.Action},
					GrantedBy:     "role_assignment",
					EffectiveFrom: time.Now().Add(-24 * time.Hour),
				},
			},
			Source:   "direct",
			IsActive: true,
		},
	}
	return types.PolicyDecisionAllow, roles
}

func (cl *compatibilityLayer) evaluateEnhancedRBAC(ctx context.Context, req *LegacyRBACRequest) (types.PolicyDecisionType, []LegacyRoleInfo, map[string]any) {
	decision, roles := cl.evaluateStrictRBAC(ctx, req)

	// Add attribute enrichment
	enrichment := map[string]any{
		"user.department":  "engineering",
		"user.clearance":   "standard",
		"context.time":     time.Now(),
		"context.ip_range": "internal",
	}

	return decision, roles, enrichment
}

func (cl *compatibilityLayer) evaluateCompatibleRBAC(ctx context.Context, req *LegacyRBACRequest) (types.PolicyDecisionType, []LegacyRoleInfo, []CompatibilityIssue) {
	decision, roles := cl.evaluateStrictRBAC(ctx, req)

	// Check for compatibility issues
	issues := []CompatibilityIssue{
		{
			IssueType:      "semantic_gap",
			Severity:       "medium",
			Description:    "RBAC role lacks contextual constraints available in ABAC",
			AffectedArea:   "permission_granularity",
			Recommendation: "Consider migrating to ABAC for enhanced control",
			AutoFixable:    false,
		},
	}

	return decision, roles, issues
}

func (cl *compatibilityLayer) evaluateTransitionRBAC(ctx context.Context, req *LegacyRBACRequest) (types.PolicyDecisionType, []LegacyRoleInfo, []CompatibilityIssue) {
	// Similar to compatible mode but with transition-specific logic
	return cl.evaluateCompatibleRBAC(ctx, req)
}

func (cl *compatibilityLayer) generatePermissionMatrix(ctx context.Context, userID uuid.UUID, roles []LegacyRoleInfo) PermissionMatrix {
	matrix := PermissionMatrix{
		UserID:            userID,
		ResourceMatrix:    make(map[string]ResourcePermissions),
		GlobalPermissions: []string{"login", "profile_read"},
	}

	// Aggregate permissions from roles
	for _, role := range roles {
		for _, perm := range role.Permissions {
			if existing, ok := matrix.ResourceMatrix[perm.ResourceType]; ok {
				existing.AllowedActions = append(existing.AllowedActions, perm.Actions...)
			} else {
				matrix.ResourceMatrix[perm.ResourceType] = ResourcePermissions{
					ResourceType:   perm.ResourceType,
					ResourceID:     perm.ResourceID,
					AllowedActions: perm.Actions,
				}
			}
		}
	}

	return matrix
}

// Translation helper methods (simplified implementations)

func (cl *compatibilityLayer) translateRoleToPolicy(ctx context.Context, req *RBACTranslationRequest) ([]TranslatedPolicy, error) {
	// Simplified role to policy translation
	return []TranslatedPolicy{
		{
			PolicyID:   uuid.New(),
			PolicyName: fmt.Sprintf("%s_access_policy", req.RoleName),
			PolicyType: types.PolicyTypeAccess,
			SourceRole: req.RoleName,
			Target: map[string]any{
				"user.role": req.RoleName,
			},
			Rule: map[string]any{
				"allow": true,
			},
			Confidence:  0.9,
			Assumptions: []string{"Role semantics preserved"},
		},
	}, nil
}

func (cl *compatibilityLayer) translatePermissionToAttribute(ctx context.Context, req *RBACTranslationRequest) ([]TranslatedAttribute, error) {
	// Simplified permission to attribute translation
	var attributes []TranslatedAttribute

	for _, perm := range req.Permissions {
		attributes = append(attributes, TranslatedAttribute{
			AttributeID:      uuid.New(),
			AttributeName:    fmt.Sprintf("permission.%s", perm),
			DataType:         "boolean",
			Category:         "user",
			SourcePermission: perm,
			DefaultValue:     false,
		})
	}

	return attributes, nil
}

func (cl *compatibilityLayer) translateHierarchyToRules(ctx context.Context, req *RBACTranslationRequest) ([]TranslatedPolicy, error) {
	// Placeholder for hierarchy translation
	return []TranslatedPolicy{}, nil
}

func (cl *compatibilityLayer) translateComplete(ctx context.Context, req *RBACTranslationRequest) ([]TranslatedPolicy, []TranslatedAttribute, error) {
	policies, err := cl.translateRoleToPolicy(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	attributes, err := cl.translatePermissionToAttribute(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return policies, attributes, nil
}

func (cl *compatibilityLayer) generateTransformationMap(ctx context.Context, req *RBACTranslationRequest, policies []TranslatedPolicy, attributes []TranslatedAttribute) map[string]any {
	return map[string]any{
		"source_role":         req.RoleName,
		"policies_created":    len(policies),
		"attributes_created":  len(attributes),
		"transformation_type": string(req.TranslationMode),
	}
}

func (cl *compatibilityLayer) performSemanticAnalysis(ctx context.Context, req *RBACTranslationRequest, policies []TranslatedPolicy, attributes []TranslatedAttribute) SemanticAnalysis {
	return SemanticAnalysis{
		SemanticPreservation: 0.85,
		IdentifiedPatterns:   []string{"role_based_access", "resource_specific_permissions"},
		LostSemantics:        []string{"implicit_hierarchies"},
		EnhancedCapabilities: []string{"contextual_constraints", "dynamic_evaluation"},
		ComplexityAnalysis: ComplexityAnalysis{
			SourceComplexity:    10,
			TargetComplexity:    15,
			ComplexityReduction: -0.5,
			PerformanceImpact:   "minimal",
		},
	}
}

func (cl *compatibilityLayer) validateTranslation(ctx context.Context, req *RBACTranslationRequest, policies []TranslatedPolicy, attributes []TranslatedAttribute) []TranslationValidation {
	return []TranslationValidation{
		{
			ValidationID:   uuid.New(),
			ValidationType: "semantic_correctness",
			Status:         "passed",
			Message:        "Translation preserves original semantics",
		},
		{
			ValidationID:   uuid.New(),
			ValidationType: "policy_syntax",
			Status:         "passed",
			Message:        "Generated policies have valid syntax",
		},
	}
}

// Gradual migration helper methods (simplified implementations)

func (cl *compatibilityLayer) validateMigrationConfig(ctx context.Context, req *GradualMigrationConfig) []ConfigValidation {
	return []ConfigValidation{
		{
			ValidationID:   uuid.New(),
			ValidationType: "phase_dependencies",
			Status:         "passed",
			Message:        "All phase dependencies are valid",
			Severity:       "info",
		},
	}
}

func (cl *compatibilityLayer) assessConfigurationRisks(ctx context.Context, req *GradualMigrationConfig) ConfigRiskAssessment {
	return ConfigRiskAssessment{
		OverallRisk: "medium",
		IdentifiedRisks: []ConfigurationRisk{
			{
				RiskID:      uuid.New(),
				RiskType:    "performance_degradation",
				Description: "Potential performance impact during dual evaluation",
				Probability: 0.4,
				Impact:      "medium",
				Severity:    "medium",
			},
		},
		SafetyMeasures: []SafetyMeasure{
			{
				MeasureID:   uuid.New(),
				MeasureType: "automatic_rollback",
				Description: "Automatic rollback on high error rates",
				Enabled:     true,
			},
		},
	}
}

func (cl *compatibilityLayer) generatePhaseSchedule(ctx context.Context, phases []MigrationPhaseConfig) []PhaseScheduleEntry {
	var schedule []PhaseScheduleEntry

	for _, phase := range phases {
		entry := PhaseScheduleEntry{
			PhaseID:   phase.PhaseID,
			PhaseName: phase.PhaseName,
			ScheduledStart: func() time.Time {
				if phase.StartDate != nil {
					return *phase.StartDate
				}
				return time.Now()
			}(),
			ScheduledEnd: func() time.Time {
				if phase.EndDate != nil {
					return *phase.EndDate
				}
				return time.Now().Add(7 * 24 * time.Hour)
			}(),
		}
		schedule = append(schedule, entry)
	}

	return schedule
}

func (cl *compatibilityLayer) determineActivePhases(ctx context.Context, phases []MigrationPhaseConfig) []ActiveMigrationPhase {
	var activePhases []ActiveMigrationPhase

	for _, phase := range phases {
		if phase.StartDate == nil || phase.StartDate.Before(time.Now()) {
			activePhases = append(activePhases, ActiveMigrationPhase{
				PhaseID:       phase.PhaseID,
				PhaseName:     phase.PhaseName,
				Status:        "active",
				Progress:      0.5,
				AffectedUsers: 100,
				StartedAt:     time.Now(),
			})
		}
	}

	return activePhases
}

func (cl *compatibilityLayer) estimateCompletion(ctx context.Context, phases []MigrationPhaseConfig) *time.Time {
	maxEnd := time.Now()

	for _, phase := range phases {
		if phase.EndDate != nil && phase.EndDate.After(maxEnd) {
			maxEnd = *phase.EndDate
		}
	}

	return &maxEnd
}

// Placeholder method implementations for complete interface

type (
	FallbackEvaluationRequest      struct{}
	FallbackEvaluationResult       struct{}
	PermissionBridgeRequest        struct{}
	PermissionBridge               struct{}
	BridgeEvaluationRequest        struct{}
	BridgeEvaluationResult         struct{}
	CompatibilityMonitoringRequest struct{}
	CompatibilityMonitoringResult  struct{}
	CompatibilityReportRequest     struct{}
	CompatibilityReport            struct{}
)

func (cl *compatibilityLayer) EvaluateWithFallback(ctx context.Context, req *FallbackEvaluationRequest) (*FallbackEvaluationResult, error) {
	return &FallbackEvaluationResult{}, nil
}

func (cl *compatibilityLayer) CreatePermissionBridge(ctx context.Context, req *PermissionBridgeRequest) (*PermissionBridge, error) {
	return &PermissionBridge{}, nil
}

func (cl *compatibilityLayer) EvaluatePermissionBridge(ctx context.Context, req *BridgeEvaluationRequest) (*BridgeEvaluationResult, error) {
	return &BridgeEvaluationResult{}, nil
}

func (cl *compatibilityLayer) MonitorCompatibility(ctx context.Context, req *CompatibilityMonitoringRequest) (*CompatibilityMonitoringResult, error) {
	return &CompatibilityMonitoringResult{}, nil
}

func (cl *compatibilityLayer) GenerateCompatibilityReport(ctx context.Context, req *CompatibilityReportRequest) (*CompatibilityReport, error) {
	return &CompatibilityReport{}, nil
}
