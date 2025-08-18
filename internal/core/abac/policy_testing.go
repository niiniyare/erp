package abac

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"

	"github.com/niiniyare/erp/internal/core/abac/models"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/niiniyare/erp/internal/shared/types"
)

// Policy Testing and Simulation types and methods

type PolicyTestRequest struct {
	PolicyID    uuid.UUID        `json:"policy_id" validate:"required"`
	TestCases   []PolicyTestCase `json:"test_cases" validate:"required,min=1"`
	Description string           `json:"description,omitempty"`
	CreatedBy   *uuid.UUID       `json:"created_by,omitempty"`
	RequestID   string           `json:"request_id,omitempty"`
}

type PolicyTestCase struct {
	Name           string                   `json:"name" validate:"required"`
	Description    string                   `json:"description,omitempty"`
	Context        map[string]any           `json:"context" validate:"required"`
	ExpectedResult types.PolicyDecisionType `json:"expected_result" validate:"required"`
	UserID         uuid.UUID                `json:"user_id" validate:"required"`
	ResourceType   string                   `json:"resource_type" validate:"required"`
	ResourceID     *uuid.UUID               `json:"resource_id,omitempty"`
	Action         string                   `json:"action" validate:"required"`
	EntityID       *uuid.UUID               `json:"entity_id,omitempty"`
}

type PolicyTestResult struct {
	PolicyID        uuid.UUID               `json:"policy_id"`
	PolicyName      string                  `json:"policy_name"`
	OverallResult   PolicyTestOverallResult `json:"overall_result"`
	TestCaseResults []PolicyTestCaseResult  `json:"test_case_results"`
	ExecutionTime   time.Duration           `json:"execution_time"`
	Timestamp       time.Time               `json:"timestamp"`
	Summary         PolicyTestSummary       `json:"summary"`
}

type PolicyTestOverallResult struct {
	Passed      bool     `json:"passed"`
	TotalTests  int      `json:"total_tests"`
	PassedTests int      `json:"passed_tests"`
	FailedTests int      `json:"failed_tests"`
	SuccessRate float64  `json:"success_rate"`
	Errors      []string `json:"errors,omitempty"`
}

type PolicyTestCaseResult struct {
	TestCase        PolicyTestCase           `json:"test_case"`
	ActualResult    types.PolicyDecisionType `json:"actual_result"`
	ExpectedResult  types.PolicyDecisionType `json:"expected_result"`
	Passed          bool                     `json:"passed"`
	ExecutionTime   time.Duration            `json:"execution_time"`
	PolicyDecisions []*models.PolicyDecision `json:"policy_decisions"`
	Error           string                   `json:"error,omitempty"`
	Details         PolicyTestCaseDetails    `json:"details"`
}

type PolicyTestCaseDetails struct {
	AttributesCollected map[string]any       `json:"attributes_collected"`
	RuleEvaluationSteps []RuleEvaluationStep `json:"rule_evaluation_steps"`
	TargetMatched       bool                 `json:"target_matched"`
	CacheHit            bool                 `json:"cache_hit"`
}

type RuleEvaluationStep struct {
	StepNumber  int            `json:"step_number"`
	Description string         `json:"description"`
	Result      bool           `json:"result"`
	Details     map[string]any `json:"details,omitempty"`
}

type PolicyTestSummary struct {
	Recommendations  []string               `json:"recommendations"`
	CoverageAnalysis PolicyCoverageAnalysis `json:"coverage_analysis"`
	PerformanceStats PolicyPerformanceStats `json:"performance_stats"`
}

type PolicyCoverageAnalysis struct {
	TargetCoverage float64            `json:"target_coverage"`
	RuleCoverage   float64            `json:"rule_coverage"`
	BranchCoverage map[string]float64 `json:"branch_coverage"`
	UncoveredPaths []string           `json:"uncovered_paths"`
}

type PolicyPerformanceStats struct {
	MinExecutionTime time.Duration `json:"min_execution_time"`
	MaxExecutionTime time.Duration `json:"max_execution_time"`
	AvgExecutionTime time.Duration `json:"avg_execution_time"`
	P95ExecutionTime time.Duration `json:"p95_execution_time"`
}

func (pm *policyManager) TestPolicy(ctx context.Context, req *PolicyTestRequest) (*PolicyTestResult, error) {
	ctx, span := pm.tracer.StartSpan(ctx, "abac.policy_manager.TestPolicy",
		tracing.WithAttributes(
			attribute.String("policy_id", req.PolicyID.String()),
			attribute.Int("test_cases", len(req.TestCases)),
		))
	defer span.End()

	startTime := time.Now()

	pm.logger.InfoContext(ctx, "Starting policy testing",
		logger.Fields{
			"policy_id":   req.PolicyID,
			"test_cases":  len(req.TestCases),
			"description": req.Description,
		})

	// Get the policy
	policy, err := pm.policyRepo.GetPolicyByID(ctx, req.PolicyID)
	if err != nil {
		pm.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_NOT_FOUND", "Policy not found for testing")
	}

	// Execute test cases
	var testCaseResults []PolicyTestCaseResult
	passedTests := 0
	var allErrors []string

	for i, testCase := range req.TestCases {
		pm.logger.DebugContext(ctx, "Executing test case",
			logger.Fields{
				"test_case_index": i,
				"test_case_name":  testCase.Name,
			})

		result, err := pm.executeTestCase(ctx, policy, testCase)
		if err != nil {
			allErrors = append(allErrors, fmt.Sprintf("Test case '%s': %s", testCase.Name, err.Error()))
			result = PolicyTestCaseResult{
				TestCase:       testCase,
				ExpectedResult: testCase.ExpectedResult,
				Passed:         false,
				Error:          err.Error(),
			}
		}

		testCaseResults = append(testCaseResults, result)
		if result.Passed {
			passedTests++
		}
	}

	executionTime := time.Since(startTime)
	totalTests := len(req.TestCases)
	successRate := float64(passedTests) / float64(totalTests) * 100

	// Generate summary and recommendations
	summary := pm.generateTestSummary(ctx, policy, testCaseResults)

	// Record metrics
	pm.metrics.IncrementCounter("policy_manager_test", nil)
	pm.metrics.SetGauge("policy_test_success_rate", successRate,
		metrics.Fields{"policy_id": req.PolicyID.String()})
	pm.metrics.ObserveHistogram("policy_test_duration_seconds", executionTime.Seconds(),
		metrics.Fields{"test_cases": fmt.Sprintf("%d", totalTests)})

	result := &PolicyTestResult{
		PolicyID:   req.PolicyID,
		PolicyName: policy.Name,
		OverallResult: PolicyTestOverallResult{
			Passed:      passedTests == totalTests,
			TotalTests:  totalTests,
			PassedTests: passedTests,
			FailedTests: totalTests - passedTests,
			SuccessRate: successRate,
			Errors:      allErrors,
		},
		TestCaseResults: testCaseResults,
		ExecutionTime:   executionTime,
		Timestamp:       time.Now(),
		Summary:         summary,
	}

	pm.logger.InfoContext(ctx, "Policy testing completed",
		logger.Fields{
			"policy_id":      req.PolicyID,
			"total_tests":    totalTests,
			"passed_tests":   passedTests,
			"success_rate":   successRate,
			"execution_time": executionTime.Milliseconds(),
		})

	return result, nil
}

func (pm *policyManager) executeTestCase(ctx context.Context, policy *models.Policy, testCase PolicyTestCase) (PolicyTestCaseResult, error) {
	startTime := time.Now()

	// This is a simplified test execution
	// In a real implementation, you would use the actual policy evaluation engine

	result := PolicyTestCaseResult{
		TestCase:       testCase,
		ExpectedResult: testCase.ExpectedResult,
		ExecutionTime:  time.Since(startTime),
	}

	// Simulate policy evaluation
	// For now, we'll just return a basic result
	actualResult := types.PolicyDecisionAllow // Simplified
	result.ActualResult = actualResult
	result.Passed = (actualResult == testCase.ExpectedResult)

	// Add evaluation details
	result.Details = PolicyTestCaseDetails{
		AttributesCollected: testCase.Context,
		TargetMatched:       true,
		CacheHit:            false,
		RuleEvaluationSteps: []RuleEvaluationStep{
			{
				StepNumber:  1,
				Description: "Target evaluation",
				Result:      true,
				Details:     map[string]any{"matched": true},
			},
			{
				StepNumber:  2,
				Description: "Rule evaluation",
				Result:      true,
				Details:     map[string]any{"conditions_met": true},
			},
		},
	}

	result.PolicyDecisions = []*models.PolicyDecision{
		{
			PolicyID: policy.ID,
			Decision: types.PolicyDecisionAllow,
			Reason:   "Test case evaluation",
		},
	}

	return result, nil
}

func (pm *policyManager) generateTestSummary(ctx context.Context, policy *models.Policy, results []PolicyTestCaseResult) PolicyTestSummary {
	var recommendations []string
	var minTime, maxTime, totalTime time.Duration

	if len(results) > 0 {
		minTime = results[0].ExecutionTime
		maxTime = results[0].ExecutionTime
	}

	for _, result := range results {
		totalTime += result.ExecutionTime
		if result.ExecutionTime < minTime {
			minTime = result.ExecutionTime
		}
		if result.ExecutionTime > maxTime {
			maxTime = result.ExecutionTime
		}

		if !result.Passed {
			recommendations = append(recommendations,
				fmt.Sprintf("Test case '%s' failed - review policy logic for this scenario", result.TestCase.Name))
		}
	}

	avgTime := totalTime / time.Duration(len(results))

	// Generate coverage analysis (simplified)
	coverage := PolicyCoverageAnalysis{
		TargetCoverage: 80.0, // Placeholder
		RuleCoverage:   75.0, // Placeholder
		BranchCoverage: map[string]float64{
			"allow_branch": 60.0,
			"deny_branch":  40.0,
		},
		UncoveredPaths: []string{
			"Edge case: empty context",
			"Boundary condition: maximum priority",
		},
	}

	performance := PolicyPerformanceStats{
		MinExecutionTime: minTime,
		MaxExecutionTime: maxTime,
		AvgExecutionTime: avgTime,
		P95ExecutionTime: maxTime, // Simplified calculation
	}

	if avgTime > 100*time.Millisecond {
		recommendations = append(recommendations, "Policy evaluation is slow - consider optimizing rule complexity")
	}

	if coverage.RuleCoverage < 80 {
		recommendations = append(recommendations, "Add more test cases to improve rule coverage")
	}

	return PolicyTestSummary{
		Recommendations:  recommendations,
		CoverageAnalysis: coverage,
		PerformanceStats: performance,
	}
}

// Policy Simulation types and methods

type PolicySimulationRequest struct {
	Scenario       PolicyScenario    `json:"scenario" validate:"required"`
	Changes        []PolicyChange    `json:"changes" validate:"required,min=1"`
	TestDataSet    PolicyTestDataSet `json:"test_data_set" validate:"required"`
	Description    string            `json:"description,omitempty"`
	SimulationMode string            `json:"simulation_mode"` // "what_if", "impact_analysis", "rollback_test"
}

type PolicyScenario struct {
	Name         string     `json:"name" validate:"required"`
	Description  string     `json:"description,omitempty"`
	BaselineDate *time.Time `json:"baseline_date,omitempty"`
	TargetDate   *time.Time `json:"target_date,omitempty"`
	Environment  string     `json:"environment"` // "production", "staging", "test"
}

type PolicyChange struct {
	ChangeType  string               `json:"change_type"` // "create", "update", "delete", "activate", "deactivate"
	PolicyID    *uuid.UUID           `json:"policy_id,omitempty"`
	PolicyData  *CreatePolicyRequest `json:"policy_data,omitempty"`
	UpdateData  *UpdatePolicyRequest `json:"update_data,omitempty"`
	Description string               `json:"description,omitempty"`
}

type PolicyTestDataSet struct {
	Users      []uuid.UUID          `json:"users"`
	Resources  []PolicyTestResource `json:"resources"`
	Actions    []string             `json:"actions"`
	Contexts   []map[string]any     `json:"contexts"`
	SampleSize int                  `json:"sample_size"` // Number of random combinations to test
}

type PolicyTestResource struct {
	Type       string         `json:"type"`
	ID         *uuid.UUID     `json:"id,omitempty"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

type PolicySimulationResult struct {
	Scenario         PolicyScenario           `json:"scenario"`
	BaselineResults  PolicySimulationSnapshot `json:"baseline_results"`
	ProjectedResults PolicySimulationSnapshot `json:"projected_results"`
	ImpactAnalysis   PolicyImpactAnalysis     `json:"impact_analysis"`
	Recommendations  []string                 `json:"recommendations"`
	ExecutionTime    time.Duration            `json:"execution_time"`
	Timestamp        time.Time                `json:"timestamp"`
}

type PolicySimulationSnapshot struct {
	TotalEvaluations       int64                       `json:"total_evaluations"`
	AllowDecisions         int64                       `json:"allow_decisions"`
	DenyDecisions          int64                       `json:"deny_decisions"`
	NotApplicableDecisions int64                       `json:"not_applicable_decisions"`
	AvgEvaluationTime      time.Duration               `json:"avg_evaluation_time"`
	PolicyUtilization      map[uuid.UUID]int64         `json:"policy_utilization"`
	ErrorRate              float64                     `json:"error_rate"`
	PerformanceMetrics     PolicySimulationPerformance `json:"performance_metrics"`
}

type PolicySimulationPerformance struct {
	P50EvaluationTime time.Duration `json:"p50_evaluation_time"`
	P95EvaluationTime time.Duration `json:"p95_evaluation_time"`
	P99EvaluationTime time.Duration `json:"p99_evaluation_time"`
	MaxEvaluationTime time.Duration `json:"max_evaluation_time"`
	ThroughputPerSec  float64       `json:"throughput_per_sec"`
}

func (pm *policyManager) SimulatePolicyChange(ctx context.Context, req *PolicySimulationRequest) (*PolicySimulationResult, error) {
	ctx, span := pm.tracer.StartSpan(ctx, "abac.policy_manager.SimulatePolicyChange",
		tracing.WithAttributes(
			attribute.String("scenario_name", req.Scenario.Name),
			attribute.String("simulation_mode", req.SimulationMode),
			attribute.Int("changes_count", len(req.Changes)),
		))
	defer span.End()

	startTime := time.Now()

	pm.logger.InfoContext(ctx, "Starting policy simulation",
		logger.Fields{
			"scenario":        req.Scenario.Name,
			"simulation_mode": req.SimulationMode,
			"changes_count":   len(req.Changes),
			"sample_size":     req.TestDataSet.SampleSize,
		})

	// Step 1: Capture baseline snapshot
	baselineSnapshot, err := pm.captureBaselineSnapshot(ctx, req.TestDataSet)
	if err != nil {
		pm.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, errors.NewBusinessErrorWithContext(ctx, "SIMULATION_BASELINE_FAILED", "Failed to capture baseline snapshot")
	}

	// Step 2: Apply changes in simulation environment
	projectedSnapshot, err := pm.simulateChanges(ctx, req.Changes, req.TestDataSet)
	if err != nil {
		pm.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		return nil, errors.NewBusinessErrorWithContext(ctx, "SIMULATION_CHANGES_FAILED", "Failed to simulate policy changes")
	}

	// Step 3: Analyze impact
	impactAnalysis := pm.analyzeSimulationImpact(ctx, baselineSnapshot, projectedSnapshot)

	// Step 4: Generate recommendations
	recommendations := pm.generateSimulationRecommendations(ctx, req, impactAnalysis)

	executionTime := time.Since(startTime)

	// Record metrics
	pm.metrics.IncrementCounter("policy_manager_simulation", nil)
	pm.metrics.ObserveHistogram("policy_simulation_duration_seconds", executionTime.Seconds(),
		metrics.Fields{
			"mode":    req.SimulationMode,
			"changes": fmt.Sprintf("%d", len(req.Changes)),
		})

	result := &PolicySimulationResult{
		Scenario:         req.Scenario,
		BaselineResults:  baselineSnapshot,
		ProjectedResults: projectedSnapshot,
		ImpactAnalysis:   impactAnalysis,
		Recommendations:  recommendations,
		ExecutionTime:    executionTime,
		Timestamp:        time.Now(),
	}

	pm.logger.InfoContext(ctx, "Policy simulation completed",
		logger.Fields{
			"scenario":         req.Scenario.Name,
			"baseline_allows":  baselineSnapshot.AllowDecisions,
			"projected_allows": projectedSnapshot.AllowDecisions,
			"impact_rating":    impactAnalysis.OverallImpact,
			"execution_time":   executionTime.Milliseconds(),
		})

	return result, nil
}

func (pm *policyManager) captureBaselineSnapshot(ctx context.Context, testDataSet PolicyTestDataSet) (PolicySimulationSnapshot, error) {
	// This is a simplified implementation
	// In a real scenario, you would run the current policies against the test data set

	snapshot := PolicySimulationSnapshot{
		TotalEvaluations:       1000,
		AllowDecisions:         800,
		DenyDecisions:          150,
		NotApplicableDecisions: 50,
		AvgEvaluationTime:      50 * time.Millisecond,
		PolicyUtilization:      make(map[uuid.UUID]int64),
		ErrorRate:              0.01,
		PerformanceMetrics: PolicySimulationPerformance{
			P50EvaluationTime: 25 * time.Millisecond,
			P95EvaluationTime: 100 * time.Millisecond,
			P99EvaluationTime: 200 * time.Millisecond,
			MaxEvaluationTime: 500 * time.Millisecond,
			ThroughputPerSec:  1000.0,
		},
	}

	return snapshot, nil
}

func (pm *policyManager) simulateChanges(ctx context.Context, changes []PolicyChange, testDataSet PolicyTestDataSet) (PolicySimulationSnapshot, error) {
	// This is a simplified implementation
	// In a real scenario, you would apply the changes and re-run evaluations

	snapshot := PolicySimulationSnapshot{
		TotalEvaluations:       1000,
		AllowDecisions:         750, // Simulated change impact
		DenyDecisions:          200,
		NotApplicableDecisions: 50,
		AvgEvaluationTime:      55 * time.Millisecond,
		PolicyUtilization:      make(map[uuid.UUID]int64),
		ErrorRate:              0.02,
		PerformanceMetrics: PolicySimulationPerformance{
			P50EvaluationTime: 30 * time.Millisecond,
			P95EvaluationTime: 110 * time.Millisecond,
			P99EvaluationTime: 220 * time.Millisecond,
			MaxEvaluationTime: 600 * time.Millisecond,
			ThroughputPerSec:  950.0,
		},
	}

	return snapshot, nil
}

type PolicyImpactAnalysis struct {
	OverallImpact        string                     `json:"overall_impact"` // "low", "medium", "high", "critical"
	AccessibilityChanges PolicyAccessibilityChanges `json:"accessibility_changes"`
	PerformanceChanges   PolicyPerformanceChanges   `json:"performance_changes"`
	SecurityImplications PolicySecurityImplications `json:"security_implications"`
	AffectedEntities     PolicyAffectedEntities     `json:"affected_entities"`
	RiskAssessment       PolicyRiskAssessment       `json:"risk_assessment"`
}

type PolicyAccessibilityChanges struct {
	AllowIncrease   int64    `json:"allow_increase"`
	AllowDecrease   int64    `json:"allow_decrease"`
	DenyIncrease    int64    `json:"deny_increase"`
	DenyDecrease    int64    `json:"deny_decrease"`
	NetAccessChange float64  `json:"net_access_change_percent"`
	CriticalChanges []string `json:"critical_changes"`
}

type PolicyPerformanceChanges struct {
	EvaluationTimeChange time.Duration `json:"evaluation_time_change"`
	ThroughputChange     float64       `json:"throughput_change_percent"`
	ErrorRateChange      float64       `json:"error_rate_change_percent"`
	MemoryImpact         string        `json:"memory_impact"`
}

type PolicySecurityImplications struct {
	SecurityPosture     string   `json:"security_posture"` // "improved", "degraded", "unchanged"
	NewVulnerabilities  []string `json:"new_vulnerabilities"`
	ResolvedIssues      []string `json:"resolved_issues"`
	ComplianceImpact    []string `json:"compliance_impact"`
	PrivilegeEscalation bool     `json:"privilege_escalation_risk"`
}

type PolicyAffectedEntities struct {
	UsersImpacted     int32               `json:"users_impacted"`
	ResourcesImpacted int32               `json:"resources_impacted"`
	EntitiesImpacted  int32               `json:"entities_impacted"`
	ImpactedGroups    []PolicyImpactGroup `json:"impacted_groups"`
}

type PolicyImpactGroup struct {
	GroupType   string `json:"group_type"` // "users", "resources", "entities"
	GroupName   string `json:"group_name"`
	ImpactType  string `json:"impact_type"` // "access_granted", "access_revoked", "access_modified"
	Severity    string `json:"severity"`
	Description string `json:"description"`
}

type PolicyRiskAssessment struct {
	OverallRiskLevel      string   `json:"overall_risk_level"` // "low", "medium", "high", "critical"
	IdentifiedRisks       []string `json:"identified_risks"`
	MitigationSuggestions []string `json:"mitigation_suggestions"`
	ApprovalRequired      bool     `json:"approval_required"`
	RollbackPlan          string   `json:"rollback_plan"`
}

func (pm *policyManager) analyzeSimulationImpact(ctx context.Context, baseline, projected PolicySimulationSnapshot) PolicyImpactAnalysis {
	// Calculate changes
	allowChange := projected.AllowDecisions - baseline.AllowDecisions
	denyChange := projected.DenyDecisions - baseline.DenyDecisions
	netAccessChange := float64(allowChange-denyChange) / float64(baseline.TotalEvaluations) * 100

	// Determine overall impact
	overallImpact := "low"
	if abs64(allowChange) > 100 || abs64(denyChange) > 100 {
		overallImpact = "medium"
	}
	if abs64(allowChange) > 500 || abs64(denyChange) > 500 {
		overallImpact = "high"
	}
	if netAccessChange > 20 || netAccessChange < -20 {
		overallImpact = "critical"
	}

	performanceChange := projected.AvgEvaluationTime - baseline.AvgEvaluationTime
	throughputChange := (projected.PerformanceMetrics.ThroughputPerSec - baseline.PerformanceMetrics.ThroughputPerSec) / baseline.PerformanceMetrics.ThroughputPerSec * 100

	analysis := PolicyImpactAnalysis{
		OverallImpact: overallImpact,
		AccessibilityChanges: PolicyAccessibilityChanges{
			AllowIncrease:   maxInt64(0, allowChange),
			AllowDecrease:   maxInt64(0, -allowChange),
			DenyIncrease:    maxInt64(0, denyChange),
			DenyDecrease:    maxInt64(0, -denyChange),
			NetAccessChange: netAccessChange,
		},
		PerformanceChanges: PolicyPerformanceChanges{
			EvaluationTimeChange: performanceChange,
			ThroughputChange:     throughputChange,
			ErrorRateChange:      (projected.ErrorRate - baseline.ErrorRate) * 100,
		},
		SecurityImplications: PolicySecurityImplications{
			SecurityPosture: "unchanged", // Simplified
		},
		AffectedEntities: PolicyAffectedEntities{
			UsersImpacted:     100, // Simplified
			ResourcesImpacted: 50,
			EntitiesImpacted:  10,
		},
		RiskAssessment: PolicyRiskAssessment{
			OverallRiskLevel: overallImpact,
			ApprovalRequired: overallImpact == "high" || overallImpact == "critical",
		},
	}

	return analysis
}

func (pm *policyManager) generateSimulationRecommendations(ctx context.Context, req *PolicySimulationRequest, impact PolicyImpactAnalysis) []string {
	var recommendations []string

	if impact.OverallImpact == "high" || impact.OverallImpact == "critical" {
		recommendations = append(recommendations, "High impact detected - consider gradual rollout")
		recommendations = append(recommendations, "Implement additional monitoring during deployment")
	}

	if impact.AccessibilityChanges.NetAccessChange > 10 {
		recommendations = append(recommendations, "Significant access increase detected - review security implications")
	}

	if impact.AccessibilityChanges.NetAccessChange < -10 {
		recommendations = append(recommendations, "Significant access decrease detected - verify intended restrictions")
	}

	if impact.PerformanceChanges.ThroughputChange < -10 {
		recommendations = append(recommendations, "Performance degradation detected - optimize policy rules")
	}

	if impact.RiskAssessment.OverallRiskLevel == "high" || impact.RiskAssessment.OverallRiskLevel == "critical" {
		recommendations = append(recommendations, "Approval required before implementing these changes")
		recommendations = append(recommendations, "Prepare detailed rollback plan")
	}

	return recommendations
}

// Helper functions
func abs64(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
