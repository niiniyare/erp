package abac

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"

	"github.com/niiniyare/erp/internal/core/abac/models"
	"github.com/niiniyare/erp/internal/core/abac/repository"
	"github.com/niiniyare/erp/internal/shared/convert"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/niiniyare/erp/internal/shared/types"
)

// Policy Lifecycle Management types and methods

// Policy Impact and Conflict Detection

type PolicyImpactRequest struct {
	PolicyID     uuid.UUID            `json:"policy_id" validate:"required"`
	Changes      *UpdatePolicyRequest `json:"changes,omitempty"`
	Scope        PolicyImpactScope    `json:"scope"`
	TimeRange    PolicyTimeRange      `json:"time_range"`
	AnalysisType string               `json:"analysis_type"` // "current", "projected", "historical"
}

type PolicyImpactScope struct {
	Users      []uuid.UUID `json:"users,omitempty"`
	Resources  []string    `json:"resources,omitempty"`
	Actions    []string    `json:"actions,omitempty"`
	Entities   []uuid.UUID `json:"entities,omitempty"`
	IncludeAll bool        `json:"include_all"`
}

type PolicyTimeRange struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

type PolicyImpactResult struct {
	PolicyID         uuid.UUID                    `json:"policy_id"`
	PolicyName       string                       `json:"policy_name"`
	ImpactSummary    PolicyImpactSummary          `json:"impact_summary"`
	UserImpacts      []PolicyUserImpact           `json:"user_impacts"`
	ResourceImpacts  []PolicyResourceImpact       `json:"resource_impacts"`
	EntityImpacts    []PolicyEntityImpact         `json:"entity_impacts"`
	TimelineAnalysis PolicyTimelineAnalysis       `json:"timeline_analysis"`
	Recommendations  []PolicyImpactRecommendation `json:"recommendations"`
	GeneratedAt      time.Time                    `json:"generated_at"`
}

type PolicyImpactSummary struct {
	TotalUsersAffected     int32   `json:"total_users_affected"`
	TotalResourcesAffected int32   `json:"total_resources_affected"`
	TotalEntitiesAffected  int32   `json:"total_entities_affected"`
	CriticalImpacts        int32   `json:"critical_impacts"`
	MediumImpacts          int32   `json:"medium_impacts"`
	LowImpacts             int32   `json:"low_impacts"`
	OverallSeverity        string  `json:"overall_severity"`
	AccessibilityChange    float64 `json:"accessibility_change_percent"`
}

type PolicyUserImpact struct {
	UserID          uuid.UUID `json:"user_id"`
	ImpactType      string    `json:"impact_type"` // "access_granted", "access_revoked", "access_modified"
	Severity        string    `json:"severity"`
	AffectedActions []string  `json:"affected_actions"`
	Description     string    `json:"description"`
}

type PolicyResourceImpact struct {
	ResourceType  string     `json:"resource_type"`
	ResourceID    *uuid.UUID `json:"resource_id,omitempty"`
	ImpactType    string     `json:"impact_type"`
	Severity      string     `json:"severity"`
	AffectedUsers int32      `json:"affected_users"`
	Description   string     `json:"description"`
}

type PolicyEntityImpact struct {
	EntityID          uuid.UUID `json:"entity_id"`
	ImpactType        string    `json:"impact_type"`
	Severity          string    `json:"severity"`
	AffectedUsers     int32     `json:"affected_users"`
	AffectedResources int32     `json:"affected_resources"`
	Description       string    `json:"description"`
}

type PolicyTimelineAnalysis struct {
	HistoricalTrends  []PolicyUsageTrend      `json:"historical_trends"`
	PeakUsagePeriods  []PolicyPeakPeriod      `json:"peak_usage_periods"`
	SeasonalPatterns  []PolicySeasonalPattern `json:"seasonal_patterns"`
	FutureProjections []PolicyUsageProjection `json:"future_projections"`
}

type PolicyUsageTrend struct {
	Period      string    `json:"period"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Evaluations int64     `json:"evaluations"`
	AllowRate   float64   `json:"allow_rate"`
	DenyRate    float64   `json:"deny_rate"`
	Trend       string    `json:"trend"` // "increasing", "decreasing", "stable"
}

type PolicyPeakPeriod struct {
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	PeakEvaluations int64     `json:"peak_evaluations"`
	TriggerReason   string    `json:"trigger_reason"`
	ImpactLevel     string    `json:"impact_level"`
}

type PolicySeasonalPattern struct {
	Pattern    string   `json:"pattern"` // "daily", "weekly", "monthly", "yearly"
	PeakTimes  []string `json:"peak_times"`
	LowTimes   []string `json:"low_times"`
	Variance   float64  `json:"variance"`
	Confidence float64  `json:"confidence"`
}

type PolicyUsageProjection struct {
	FutureDate           time.Time `json:"future_date"`
	ProjectedEvaluations int64     `json:"projected_evaluations"`
	ConfidenceLevel      float64   `json:"confidence_level"`
	Assumptions          []string  `json:"assumptions"`
}

type PolicyImpactRecommendation struct {
	Type        string `json:"type"`     // "action", "warning", "info"
	Priority    string `json:"priority"` // "high", "medium", "low"
	Title       string `json:"title"`
	Description string `json:"description"`
	Action      string `json:"action,omitempty"`
}

func (pm *policyManager) AnalyzePolicyImpact(ctx context.Context, req *PolicyImpactRequest) (*PolicyImpactResult, error) {
	ctx, span := pm.tracer.StartSpan(ctx, "abac.policy_manager.AnalyzePolicyImpact",
		tracing.WithAttributes(
			attribute.String("policy_id", req.PolicyID.String()),
			attribute.String("analysis_type", req.AnalysisType),
		),
	)
	defer span.End()

	pm.logger.InfoContext(ctx, "Starting policy impact analysis",
		logger.Fields{
			"policy_id":     req.PolicyID,
			"analysis_type": req.AnalysisType,
			"scope":         req.Scope,
		})

	// Get the policy
	policy, err := pm.policyRepo.GetPolicyByID(ctx, req.PolicyID)
	if err != nil {
		pm.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, fmt.Errorf("policy not found for impact analysis: %w", err)
	}

	// Analyze current impact
	userImpacts, err := pm.analyzeUserImpacts(ctx, policy, req)
	if err != nil {
		pm.logger.WarnContext(ctx, "Failed to analyze user impacts", logger.Fields{"error": err.Error()})
		userImpacts = []PolicyUserImpact{}
	}

	resourceImpacts, err := pm.analyzeResourceImpacts(ctx, policy, req)
	if err != nil {
		pm.logger.WarnContext(ctx, "Failed to analyze resource impacts", logger.Fields{"error": err.Error()})
		resourceImpacts = []PolicyResourceImpact{}
	}

	entityImpacts, err := pm.analyzeEntityImpacts(ctx, policy, req)
	if err != nil {
		pm.logger.WarnContext(ctx, "Failed to analyze entity impacts", logger.Fields{"error": err.Error()})
		entityImpacts = []PolicyEntityImpact{}
	}

	// Generate timeline analysis
	timelineAnalysis, err := pm.analyzeTimelinePatterns(ctx, policy, req)
	if err != nil {
		pm.logger.WarnContext(ctx, "Failed to analyze timeline patterns", logger.Fields{"error": err.Error()})
		timelineAnalysis = PolicyTimelineAnalysis{}
	}

	// Calculate summary
	impactSummary := pm.calculateImpactSummary(userImpacts, resourceImpacts, entityImpacts)

	// Generate recommendations
	recommendations := pm.generateImpactRecommendations(ctx, policy, impactSummary, req)

	// Record metrics
	// Record metrics
	impactCounter := pm.metrics.Counter(
		"abac_policy_impact_analysis_total",
		"Total policy impact analyses performed",
		"policy_id", "analysis_type",
	)
	impactCounter.Inc(metrics.Fields{
		"policy_id":     req.PolicyID.String(),
		"analysis_type": req.AnalysisType,
	})

	userGauge := pm.metrics.Gauge(
		"abac_policy_impact_users_affected",
		"Number of users affected by policy impact",
		"policy_id",
	)
	userGauge.Set(float64(impactSummary.TotalUsersAffected), metrics.Fields{
		"policy_id": req.PolicyID.String(),
	})

	result := &PolicyImpactResult{
		PolicyID:         req.PolicyID,
		PolicyName:       policy.Name,
		ImpactSummary:    impactSummary,
		UserImpacts:      userImpacts,
		ResourceImpacts:  resourceImpacts,
		EntityImpacts:    entityImpacts,
		TimelineAnalysis: timelineAnalysis,
		Recommendations:  recommendations,
		GeneratedAt:      time.Now(),
	}

	pm.logger.InfoContext(ctx, "Policy impact analysis completed",
		logger.Fields{
			"policy_id":        req.PolicyID,
			"users_affected":   impactSummary.TotalUsersAffected,
			"overall_severity": impactSummary.OverallSeverity,
			"recommendations":  len(recommendations),
		})

	return result, nil
}

// Policy Conflict Detection

type PolicyConflictRequest struct {
	PolicyID      *uuid.UUID           `json:"policy_id,omitempty"`
	PolicySet     []uuid.UUID          `json:"policy_set,omitempty"`
	NewPolicy     *CreatePolicyRequest `json:"new_policy,omitempty"`
	ConflictTypes []string             `json:"conflict_types,omitempty"` // "effect", "priority", "target_overlap"
	Scope         PolicyConflictScope  `json:"scope"`
	AnalysisDepth string               `json:"analysis_depth"` // "basic", "detailed", "comprehensive"
}

type PolicyConflictScope struct {
	IncludeInactive bool                  `json:"include_inactive"`
	EntityFilter    *uuid.UUID            `json:"entity_filter,omitempty"`
	CategoryFilter  *types.PolicyCategory `json:"category_filter,omitempty"`
}

type PolicyConflictResult struct {
	AnalyzedPolicies  int32                    `json:"analyzed_policies"`
	ConflictSummary   PolicyConflictSummary    `json:"conflict_summary"`
	DetectedConflicts []PolicyConflictDetail   `json:"detected_conflicts"`
	ResolutionPlan    PolicyConflictResolution `json:"resolution_plan"`
	GeneratedAt       time.Time                `json:"generated_at"`
}

type PolicyConflictSummary struct {
	TotalConflicts           int32 `json:"total_conflicts"`
	CriticalConflicts        int32 `json:"critical_conflicts"`
	MajorConflicts           int32 `json:"major_conflicts"`
	MinorConflicts           int32 `json:"minor_conflicts"`
	AutoResolvableConflicts  int32 `json:"auto_resolvable_conflicts"`
	ManualResolutionRequired int32 `json:"manual_resolution_required"`
}

type PolicyConflictResolution struct {
	AutoResolutionSteps []PolicyResolutionStep `json:"auto_resolution_steps"`
	ManualStepsRequired []PolicyResolutionStep `json:"manual_steps_required"`
	EstimatedEffort     string                 `json:"estimated_effort"`
	RiskAssessment      string                 `json:"risk_assessment"`
}

type PolicyResolutionStep struct {
	StepNumber    int32          `json:"step_number"`
	Action        string         `json:"action"`
	Description   string         `json:"description"`
	PolicyID      *uuid.UUID     `json:"policy_id,omitempty"`
	Changes       map[string]any `json:"changes,omitempty"`
	Priority      string         `json:"priority"`
	EstimatedTime string         `json:"estimated_time"`
}

func (pm *policyManager) DetectPolicyConflicts(ctx context.Context, req *PolicyConflictRequest) (*PolicyConflictResult, error) {
	ctx, span := pm.tracer.StartSpan(ctx, "abac.policy_manager.DetectPolicyConflicts",
		tracing.WithAttributes(
			attribute.String("analysis_depth", req.AnalysisDepth),
			attribute.Int("policy_set_size", len(req.PolicySet)),
		),
	)
	defer span.End()

	pm.logger.InfoContext(ctx, "Starting policy conflict detection",
		logger.Fields{
			"policy_id":      req.PolicyID,
			"policy_set":     len(req.PolicySet),
			"analysis_depth": req.AnalysisDepth,
			"conflict_types": req.ConflictTypes,
		})

	// Get policies to analyze
	var policiesToAnalyze []*models.Policy
	var err error

	if req.PolicyID != nil {
		// Analyze conflicts for a specific policy
		policy, err := pm.policyRepo.GetPolicyByID(ctx, *req.PolicyID)
		if err != nil {
			pm.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
			return nil, fmt.Errorf("policy not found for conflict analysis: %w", err)
		}
		policiesToAnalyze = []*models.Policy{policy}

		// Get all other policies to compare against
		listReq := &repository.ListPoliciesRequest{
			Limit:  1000, // Large limit to get all policies
			Offset: 0,
		}
		allPolicies, err := pm.policyRepo.ListPolicies(ctx, listReq)
		if err != nil {
			pm.logger.WarnContext(ctx, "Failed to get all policies for conflict analysis", logger.Fields{"error": err.Error()})
		} else {
			// Add other policies for comparison
			for _, p := range allPolicies {
				if p.ID != *req.PolicyID {
					policiesToAnalyze = append(policiesToAnalyze, p)
				}
			}
		}
	} else if len(req.PolicySet) > 0 {
		// Analyze conflicts within a specific policy set
		policiesToAnalyze, err = pm.policyRepo.GetPoliciesByIDs(ctx, req.PolicySet)
		if err != nil {
			pm.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
			return nil, fmt.Errorf("failed to get policies for conflict analysis: %w", err)
		}
	} else {
		// Analyze conflicts across all policies
		listReq := &repository.ListPoliciesRequest{
			Limit:  1000,
			Offset: 0,
		}
		policiesToAnalyze, err = pm.policyRepo.ListPolicies(ctx, listReq)
		if err != nil {
			pm.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
			return nil, fmt.Errorf("failed to list policies for conflict analysis: %w", err)
		}
	}

	// Detect conflicts
	detectedConflicts := pm.detectConflictsBetweenPolicies(ctx, policiesToAnalyze, req)

	// Generate conflict summary
	conflictSummary := pm.calculateConflictSummary(detectedConflicts)

	// Generate resolution plan
	resolutionPlan := pm.generateConflictResolutionPlan(ctx, detectedConflicts)

	// Record metrics
	// Record metrics
	conflictCounter := pm.metrics.Counter(
		"abac_policy_conflict_detection_total",
		"Total policy conflict detections performed",
		"analysis_depth",
	)
	conflictCounter.Inc(metrics.Fields{
		"analysis_depth": req.AnalysisDepth,
	})

	conflictGauge := pm.metrics.Gauge(
		"abac_policy_conflicts_detected",
		"Number of policy conflicts detected",
		"analysis_depth",
	)
	conflictGauge.Set(float64(conflictSummary.TotalConflicts), metrics.Fields{
		"analysis_depth": req.AnalysisDepth,
	})

	analyzedPolicies, err := convert.IntToInt32(len(policiesToAnalyze))
	if err != nil {
		analyzedPolicies = 0
	}

	result := &PolicyConflictResult{
		AnalyzedPolicies:  analyzedPolicies,
		ConflictSummary:   conflictSummary,
		DetectedConflicts: detectedConflicts,
		ResolutionPlan:    resolutionPlan,
		GeneratedAt:       time.Now(),
	}

	pm.logger.InfoContext(ctx, "Policy conflict detection completed",
		logger.Fields{
			"analyzed_policies":  len(policiesToAnalyze),
			"total_conflicts":    conflictSummary.TotalConflicts,
			"critical_conflicts": conflictSummary.CriticalConflicts,
		})

	return result, nil
}

// Policy Versioning and Lifecycle

type CreatePolicyVersionRequest struct {
	PolicyID      uuid.UUID            `json:"policy_id" validate:"required"`
	VersionType   string               `json:"version_type"` // "major", "minor", "patch"
	Changes       *UpdatePolicyRequest `json:"changes" validate:"required"`
	VersionNotes  string               `json:"version_notes,omitempty"`
	CreatedBy     *uuid.UUID           `json:"created_by,omitempty"`
	ScheduledDate *time.Time           `json:"scheduled_date,omitempty"`
}

type PolicyVersionResult struct {
	OriginalPolicy *models.Policy       `json:"original_policy"`
	NewVersion     *models.Policy       `json:"new_version"`
	VersionInfo    PolicyVersionInfo    `json:"version_info"`
	ChangesSummary PolicyChangesSummary `json:"changes_summary"`
}

type PolicyVersionInfo struct {
	VersionNumber string     `json:"version_number"`
	VersionType   string     `json:"version_type"`
	CreatedAt     time.Time  `json:"created_at"`
	CreatedBy     *uuid.UUID `json:"created_by"`
	VersionNotes  string     `json:"version_notes"`
	Status        string     `json:"status"` // "draft", "active", "deprecated", "archived"
}

type PolicyChangesSummary struct {
	ChangedFields      []string               `json:"changed_fields"`
	ImpactLevel        string                 `json:"impact_level"`
	BackwardCompatible bool                   `json:"backward_compatible"`
	BreakingChanges    []string               `json:"breaking_changes"`
	Changelog          []PolicyChangelogEntry `json:"changelog"`
}

type PolicyChangelogEntry struct {
	Field      string `json:"field"`
	OldValue   any    `json:"old_value"`
	NewValue   any    `json:"new_value"`
	ChangeType string `json:"change_type"` // "added", "modified", "removed"
	Impact     string `json:"impact"`
}

func (pm *policyManager) CreatePolicyVersion(ctx context.Context, req *CreatePolicyVersionRequest) (*PolicyVersionResult, error) {
	ctx, span := pm.tracer.StartSpan(ctx, "abac.policy_manager.CreatePolicyVersion",
		tracing.WithAttributes(
			attribute.String("policy_id", req.PolicyID.String()),
			attribute.String("version_type", req.VersionType),
		),
	)
	defer span.End()

	pm.logger.InfoContext(ctx, "Creating policy version",
		logger.Fields{
			"policy_id":    req.PolicyID,
			"version_type": req.VersionType,
			"notes":        req.VersionNotes,
		})

	// Get original policy
	originalPolicy, err := pm.policyRepo.GetPolicyByID(ctx, req.PolicyID)
	if err != nil {
		pm.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, fmt.Errorf("original policy not found for versioning: %w", err)
	}

	// Generate version number
	versionNumber := pm.generateVersionNumber(originalPolicy, req.VersionType)

	// Analyze changes
	changesSummary := pm.analyzeChanges(ctx, originalPolicy, req.Changes)

	// Create new version (this is simplified - in a real implementation, you'd have a versioning table)
	newVersionName := fmt.Sprintf("%s_v%s", originalPolicy.Name, versionNumber)
	if req.Changes.Name != nil {
		newVersionName = fmt.Sprintf("%s_v%s", *req.Changes.Name, versionNumber)
	}

	createReq := &repository.CreatePolicyRequest{
		Name:               newVersionName,
		DisplayName:        req.Changes.DisplayName,
		Description:        req.Changes.Description,
		PolicyType:         originalPolicy.PolicyType,
		Effect:             originalPolicy.Effect,
		Priority:           originalPolicy.Priority,
		Category:           originalPolicy.Category,
		Target:             originalPolicy.Target,
		Rule:               originalPolicy.Rule,
		Obligations:        originalPolicy.Obligations,
		Advice:             originalPolicy.Advice,
		CombiningAlgorithm: originalPolicy.CombiningAlgorithm,
		ExpiresAt:          originalPolicy.ExpiresAt,
		CreatedBy: func() uuid.UUID {
			if req.CreatedBy != nil {
				return *req.CreatedBy
			}
			return originalPolicy.CreatedBy
		}(),
	}

	// Apply changes
	if req.Changes.DisplayName != nil {
		createReq.DisplayName = req.Changes.DisplayName
	}
	if req.Changes.Description != nil {
		createReq.Description = req.Changes.Description
	}
	if req.Changes.PolicyType != nil {
		createReq.PolicyType = *req.Changes.PolicyType
	}
	if req.Changes.Effect != nil {
		createReq.Effect = *req.Changes.Effect
	}
	if req.Changes.Priority != nil {
		createReq.Priority = *req.Changes.Priority
	}
	if req.Changes.Category != nil {
		createReq.Category = *req.Changes.Category
	}
	if req.Changes.Target != nil {
		createReq.Target = req.Changes.Target
	}
	if req.Changes.Rule != nil {
		createReq.Rule = req.Changes.Rule
	}
	if req.Changes.Obligations != nil {
		createReq.Obligations = req.Changes.Obligations
	}
	if req.Changes.Advice != nil {
		createReq.Advice = req.Changes.Advice
	}

	newVersion, err := pm.policyRepo.CreatePolicy(ctx, createReq)
	if err != nil {
		pm.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		errorCounter := pm.metrics.Counter(
			"abac_policy_version_create_errors_total",
			"Total policy version creation errors",
			"error_type",
		)
		errorCounter.Inc(metrics.Fields{"error_type": "version_create_failed"})
		return nil, fmt.Errorf("failed to create policy version: %w", err)
	}

	versionInfo := PolicyVersionInfo{
		VersionNumber: versionNumber,
		VersionType:   req.VersionType,
		CreatedAt:     time.Now(),
		CreatedBy:     req.CreatedBy,
		VersionNotes:  req.VersionNotes,
		Status:        "draft",
	}

	// Record metrics
	// Record metrics
	versionCounter := pm.metrics.Counter(
		"abac_policy_version_created_total",
		"Total policy versions created",
		"version_type",
	)
	versionCounter.Inc(metrics.Fields{
		"version_type": req.VersionType,
	})
	pm.metrics.IncrementCounter("policy_versions_total",
		metrics.Fields{"version_type": req.VersionType})

	result := &PolicyVersionResult{
		OriginalPolicy: originalPolicy,
		NewVersion:     newVersion,
		VersionInfo:    versionInfo,
		ChangesSummary: changesSummary,
	}

	pm.logger.InfoContext(ctx, "Policy version created successfully",
		logger.Fields{
			"original_policy_id": req.PolicyID,
			"new_version_id":     newVersion.ID,
			"version_number":     versionNumber,
			"impact_level":       changesSummary.ImpactLevel,
		})

	return result, nil
}

type ActivatePolicyRequest struct {
	PolicyID      uuid.UUID  `json:"policy_id" validate:"required"`
	ActivatedBy   *uuid.UUID `json:"activated_by,omitempty"`
	Reason        string     `json:"reason,omitempty"`
	ScheduledTime *time.Time `json:"scheduled_time,omitempty"`
}

type DeactivatePolicyRequest struct {
	PolicyID      uuid.UUID  `json:"policy_id" validate:"required"`
	DeactivatedBy *uuid.UUID `json:"deactivated_by,omitempty"`
	Reason        string     `json:"reason,omitempty"`
	ScheduledTime *time.Time `json:"scheduled_time,omitempty"`
}

type ArchivePolicyRequest struct {
	PolicyID        uuid.UUID      `json:"policy_id" validate:"required"`
	ArchivedBy      *uuid.UUID     `json:"archived_by,omitempty"`
	Reason          string         `json:"reason,omitempty"`
	RetentionPeriod *time.Duration `json:"retention_period,omitempty"`
}

func (pm *policyManager) ActivatePolicy(ctx context.Context, req *ActivatePolicyRequest) error {
	ctx, span := pm.tracer.StartSpan(ctx, "abac.policy_manager.ActivatePolicy",
		tracing.WithAttributes(
			attribute.String("policy_id", req.PolicyID.String()),
		),
	)
	defer span.End()

	pm.logger.InfoContext(ctx, "Activating policy",
		logger.Fields{
			"policy_id": req.PolicyID,
			"reason":    req.Reason,
		})

	// Check if policy exists and is inactive
	policy, err := pm.policyRepo.GetPolicyByID(ctx, req.PolicyID)
	if err != nil {
		pm.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return err
	}

	if policy.IsActive {
		return fmt.Errorf("policy is already active")
	}

	// Activate the policy
	updateReq := &repository.UpdatePolicyRequest{
		IsActive: &[]bool{true}[0],
	}

	_, err = pm.policyRepo.UpdatePolicy(ctx, req.PolicyID, updateReq)
	if err != nil {
		pm.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		errorCounter := pm.metrics.Counter(
			"abac_policy_activate_errors_total",
			"Total policy activation errors",
			"error_type",
		)
		errorCounter.Inc(metrics.Fields{"error_type": "activate_failed"})
		return fmt.Errorf("failed to activate policy: %w", err)
	}

	// Invalidate related cache
	go func() {
		invalidateReq := &repository.InvalidateEvaluationCacheRequest{
			PolicyIDs: []uuid.UUID{req.PolicyID},
		}
		if err := pm.policyEvaluationRepo.InvalidateEvaluationCache(context.Background(), invalidateReq); err != nil {
			pm.logger.WarnContext(ctx, "Failed to invalidate cache after policy activation",
				logger.Fields{"policy_id": req.PolicyID, "error": err.Error()})
		}
	}()

	// Record metrics
	activateCounter := pm.metrics.Counter(
		"abac_policy_activate_total",
		"Total policy activations",
		"policy_id",
	)
	activateCounter.Inc(metrics.Fields{
		"policy_id": req.PolicyID.String(),
	})

	pm.logger.InfoContext(ctx, "Policy activated successfully",
		logger.Fields{
			"policy_id":   req.PolicyID,
			"policy_name": policy.Name,
		})

	return nil
}

func (pm *policyManager) DeactivatePolicy(ctx context.Context, req *DeactivatePolicyRequest) error {
	ctx, span := pm.tracer.StartSpan(ctx, "abac.policy_manager.DeactivatePolicy",
		tracing.WithAttributes(
			attribute.String("policy_id", req.PolicyID.String()),
		),
	)
	defer span.End()

	pm.logger.InfoContext(ctx, "Deactivating policy",
		logger.Fields{
			"policy_id": req.PolicyID,
			"reason":    req.Reason,
		})

	// Check if policy exists and is active
	policy, err := pm.policyRepo.GetPolicyByID(ctx, req.PolicyID)
	if err != nil {
		pm.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return err
	}

	if !policy.IsActive {
		return fmt.Errorf("policy is already inactive")
	}

	// Deactivate the policy
	updateReq := &repository.UpdatePolicyRequest{
		IsActive: &[]bool{false}[0],
	}

	_, err = pm.policyRepo.UpdatePolicy(ctx, req.PolicyID, updateReq)
	if err != nil {
		pm.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		errorCounter := pm.metrics.Counter(
			"abac_policy_deactivate_errors_total",
			"Total policy deactivation errors",
			"error_type",
		)
		errorCounter.Inc(metrics.Fields{"error_type": "deactivate_failed"})
		return fmt.Errorf("failed to deactivate policy: %w", err)
	}

	// Invalidate related cache
	go func() {
		invalidateReq := &repository.InvalidateEvaluationCacheRequest{
			PolicyIDs: []uuid.UUID{req.PolicyID},
		}
		if err := pm.policyEvaluationRepo.InvalidateEvaluationCache(context.Background(), invalidateReq); err != nil {
			pm.logger.WarnContext(ctx, "Failed to invalidate cache after policy deactivation",
				logger.Fields{"policy_id": req.PolicyID, "error": err.Error()})
		}
	}()

	// Record metrics
	deactivateCounter := pm.metrics.Counter(
		"abac_policy_deactivate_total",
		"Total policy deactivations",
		"policy_id",
	)
	deactivateCounter.Inc(metrics.Fields{
		"policy_id": req.PolicyID.String(),
	})

	pm.logger.InfoContext(ctx, "Policy deactivated successfully",
		logger.Fields{
			"policy_id":   req.PolicyID,
			"policy_name": policy.Name,
		})

	return nil
}

func (pm *policyManager) ArchivePolicy(ctx context.Context, req *ArchivePolicyRequest) error {
	ctx, span := pm.tracer.StartSpan(ctx, "abac.policy_manager.ArchivePolicy",
		tracing.WithAttributes(
			attribute.String("policy_id", req.PolicyID.String()),
		),
	)
	defer span.End()

	pm.logger.InfoContext(ctx, "Archiving policy",
		logger.Fields{
			"policy_id": req.PolicyID,
			"reason":    req.Reason,
		})

	// Check if policy exists
	policy, err := pm.policyRepo.GetPolicyByID(ctx, req.PolicyID)
	if err != nil {
		pm.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return err
	}

	// First deactivate if active
	if policy.IsActive {
		deactivateReq := &DeactivatePolicyRequest{
			PolicyID:      req.PolicyID,
			DeactivatedBy: req.ArchivedBy,
			Reason:        "Archived: " + req.Reason,
		}
		if err := pm.DeactivatePolicy(ctx, deactivateReq); err != nil {
			pm.logger.WarnContext(ctx, "Failed to deactivate policy before archiving", logger.Fields{"error": err.Error()})
		}
	}

	// Archive the policy (soft delete)
	err = pm.policyRepo.DeletePolicy(ctx, req.PolicyID)
	if err != nil {
		pm.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		errorCounter := pm.metrics.Counter(
			"abac_policy_archive_errors_total",
			"Total policy archive errors",
			"error_type",
		)
		errorCounter.Inc(metrics.Fields{"error_type": "archive_failed"})
		return fmt.Errorf("failed to archive policy: %w", err)
	}

	// Record metrics
	archiveCounter := pm.metrics.Counter(
		"abac_policy_archive_total",
		"Total policy archives",
		"policy_id",
	)
	archiveCounter.Inc(metrics.Fields{
		"policy_id": req.PolicyID.String(),
	})

	pm.logger.InfoContext(ctx, "Policy archived successfully",
		logger.Fields{
			"policy_id":   req.PolicyID,
			"policy_name": policy.Name,
		})

	return nil
}

// Helper methods for impact analysis, conflict detection, and versioning

func (pm *policyManager) analyzeUserImpacts(ctx context.Context, policy *models.Policy, req *PolicyImpactRequest) ([]PolicyUserImpact, error) {
	// This is a simplified implementation
	// In a real scenario, you would analyze historical evaluation data
	return []PolicyUserImpact{}, nil
}

func (pm *policyManager) analyzeResourceImpacts(ctx context.Context, policy *models.Policy, req *PolicyImpactRequest) ([]PolicyResourceImpact, error) {
	// This is a simplified implementation
	return []PolicyResourceImpact{}, nil
}

func (pm *policyManager) analyzeEntityImpacts(ctx context.Context, policy *models.Policy, req *PolicyImpactRequest) ([]PolicyEntityImpact, error) {
	// This is a simplified implementation
	return []PolicyEntityImpact{}, nil
}

func (pm *policyManager) analyzeTimelinePatterns(ctx context.Context, policy *models.Policy, req *PolicyImpactRequest) (PolicyTimelineAnalysis, error) {
	// This is a simplified implementation
	return PolicyTimelineAnalysis{}, nil
}

func (pm *policyManager) calculateImpactSummary(userImpacts []PolicyUserImpact, resourceImpacts []PolicyResourceImpact, entityImpacts []PolicyEntityImpact) PolicyImpactSummary {
	usersAffected, err := convert.IntToInt32(len(userImpacts))
	if err != nil {
		// Handle or log the error appropriately
		usersAffected = 0
	}
	resourcesAffected, err := convert.IntToInt32(len(resourceImpacts))
	if err != nil {
		resourcesAffected = 0
	}
	entitiesAffected, err := convert.IntToInt32(len(entityImpacts))
	if err != nil {
		entitiesAffected = 0
	}

	return PolicyImpactSummary{
		TotalUsersAffected:     usersAffected,
		TotalResourcesAffected: resourcesAffected,
		TotalEntitiesAffected:  entitiesAffected,
		OverallSeverity:        "low",
		AccessibilityChange:    0.0,
	}
}

func (pm *policyManager) generateImpactRecommendations(ctx context.Context, policy *models.Policy, summary PolicyImpactSummary, req *PolicyImpactRequest) []PolicyImpactRecommendation {
	var recommendations []PolicyImpactRecommendation

	if summary.TotalUsersAffected > 100 {
		recommendations = append(recommendations, PolicyImpactRecommendation{
			Type:        "warning",
			Priority:    "high",
			Title:       "High User Impact",
			Description: "This policy affects a large number of users. Consider gradual rollout.",
		})
	}

	return recommendations
}

func (pm *policyManager) detectConflictsBetweenPolicies(ctx context.Context, policies []*models.Policy, req *PolicyConflictRequest) []PolicyConflictDetail {
	var conflicts []PolicyConflictDetail

	// Simplified conflict detection
	for i := 0; i < len(policies); i++ {
		for j := i + 1; j < len(policies); j++ {
			if conflict := pm.checkPolicyPairConflict(policies[i], policies[j]); conflict != nil {
				conflicts = append(conflicts, *conflict)
			}
		}
	}

	return conflicts
}

func (pm *policyManager) checkPolicyPairConflict(policy1, policy2 *models.Policy) *PolicyConflictDetail {
	// Check for target overlap with conflicting effects
	if pm.hasTargetOverlap(policy1, policy2) && policy1.Effect != policy2.Effect {
		return &PolicyConflictDetail{
			ConflictType:        "effect_conflict",
			ConflictingPolicy:   policy2,
			ConflictDescription: fmt.Sprintf("Policies have overlapping targets but conflicting effects (%s vs %s)", policy1.Effect, policy2.Effect),
			Severity:            "high",
			AutoResolvable:      false,
		}
	}

	return nil
}

func (pm *policyManager) hasTargetOverlap(policy1, policy2 *models.Policy) bool {
	// Simplified target overlap detection
	// In a real implementation, you would compare target conditions in detail
	return true // Placeholder
}

func (pm *policyManager) calculateConflictSummary(conflicts []PolicyConflictDetail) PolicyConflictSummary {
	totalConflicts, err := convert.IntToInt32(len(conflicts))
	if err != nil {
		totalConflicts = 0
	}
	summary := PolicyConflictSummary{
		TotalConflicts: totalConflicts,
	}

	for _, conflict := range conflicts {
		switch conflict.Severity {
		case "critical":
			summary.CriticalConflicts++
		case "high":
			summary.MajorConflicts++
		default:
			summary.MinorConflicts++
		}

		if conflict.AutoResolvable {
			summary.AutoResolvableConflicts++
		} else {
			summary.ManualResolutionRequired++
		}
	}

	return summary
}

func (pm *policyManager) generateConflictResolutionPlan(ctx context.Context, conflicts []PolicyConflictDetail) PolicyConflictResolution {
	var autoSteps []PolicyResolutionStep
	var manualSteps []PolicyResolutionStep

	for i, conflict := range conflicts {
		if conflict.AutoResolvable {
			autoSteps = append(autoSteps, PolicyResolutionStep{
				StepNumber:  int32(i + 1),
				Action:      "auto_resolve",
				Description: fmt.Sprintf("Auto-resolve %s conflict", conflict.ConflictType),
				Priority:    conflict.Severity,
			})
		} else {
			manualSteps = append(manualSteps, PolicyResolutionStep{
				StepNumber:  int32(i + 1),
				Action:      "manual_review",
				Description: fmt.Sprintf("Manual review required for %s conflict", conflict.ConflictType),
				Priority:    conflict.Severity,
			})
		}
	}

	return PolicyConflictResolution{
		AutoResolutionSteps: autoSteps,
		ManualStepsRequired: manualSteps,
		EstimatedEffort:     "medium",
		RiskAssessment:      "low",
	}
}

func (pm *policyManager) generateVersionNumber(policy *models.Policy, versionType string) string {
	// Simplified version number generation
	// In a real implementation, you would track version history
	switch versionType {
	case "major":
		return "2.0.0"
	case "minor":
		return "1.1.0"
	case "patch":
		return "1.0.1"
	default:
		return "1.0.0"
	}
}

func (pm *policyManager) analyzeChanges(ctx context.Context, original *models.Policy, changes *UpdatePolicyRequest) PolicyChangesSummary {
	var changedFields []string
	var changelog []PolicyChangelogEntry

	if changes.Name != nil && *changes.Name != original.Name {
		changedFields = append(changedFields, "name")
		changelog = append(changelog, PolicyChangelogEntry{
			Field:      "name",
			OldValue:   original.Name,
			NewValue:   *changes.Name,
			ChangeType: "modified",
			Impact:     "low",
		})
	}

	if changes.Effect != nil && *changes.Effect != original.Effect {
		changedFields = append(changedFields, "effect")
		changelog = append(changelog, PolicyChangelogEntry{
			Field:      "effect",
			OldValue:   string(original.Effect),
			NewValue:   string(*changes.Effect),
			ChangeType: "modified",
			Impact:     "high",
		})
	}

	if changes.Priority != nil && *changes.Priority != original.Priority {
		changedFields = append(changedFields, "priority")
		changelog = append(changelog, PolicyChangelogEntry{
			Field:      "priority",
			OldValue:   original.Priority,
			NewValue:   *changes.Priority,
			ChangeType: "modified",
			Impact:     "medium",
		})
	}

	// Determine impact level
	impactLevel := "low"
	backwardCompatible := true
	var breakingChanges []string

	for _, entry := range changelog {
		if entry.Impact == "high" {
			impactLevel = "high"
			backwardCompatible = false
			breakingChanges = append(breakingChanges, fmt.Sprintf("Changed %s", entry.Field))
		} else if entry.Impact == "medium" && impactLevel == "low" {
			impactLevel = "medium"
		}
	}

	return PolicyChangesSummary{
		ChangedFields:      changedFields,
		ImpactLevel:        impactLevel,
		BackwardCompatible: backwardCompatible,
		BreakingChanges:    breakingChanges,
		Changelog:          changelog,
	}
}
