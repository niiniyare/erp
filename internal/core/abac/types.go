package abac

import (
	"time"

	"github.com/google/uuid"
)

// ─── Consolidated Types ───────────────────────────────────────────────────

// ValidationLevel defines the level of validation to be performed.
type ValidationLevel string

const (
	ValidationLevelBasic  ValidationLevel = "basic"
	ValidationLevelStrict ValidationLevel = "strict"
)

// FallbackStrategy defines the strategy to use when a primary evaluation fails.
type FallbackStrategy string

const (
	FallbackStrategyRBAC FallbackStrategy = "rbac"
	FallbackStrategyABAC FallbackStrategy = "abac"
	FallbackStrategyDeny FallbackStrategy = "deny"
)

// RetryPolicy defines the retry behavior for operations.
type RetryPolicy struct {
	MaxRetries        int32         `json:"max_retries"`
	InitialDelay      time.Duration `json:"initial_delay"`
	BackoffMultiplier float64       `json:"backoff_multiplier"`
	MaxDelay          time.Duration `json:"max_delay"`
	RetryConditions   []string      `json:"retry_conditions"`
	MaxAttempts       int           `json:"max_attempts"`
	Backoff           time.Duration `json:"backoff"`
}

// SourcePerformanceMetrics tracks performance of an external source.
type SourcePerformanceMetrics struct {
	SourceID        uuid.UUID     `json:"source_id"`
	RequestCount    int64         `json:"request_count"`
	SuccessCount    int64         `json:"success_count"`
	ErrorCount      int64         `json:"error_count"`
	AverageLatency  time.Duration `json:"average_latency"`
	LastRequestTime time.Time     `json:"last_request_time"`
}

// HybridEvaluationMode defines the mode for hybrid RBAC-ABAC evaluation.
type HybridEvaluationMode string

const (
	HybridEvaluationModeParallel   HybridEvaluationMode = "parallel"
	HybridEvaluationModeSequential HybridEvaluationMode = "sequential"
	HybridEvaluationModePrimary    HybridEvaluationMode = "primary"
)

// MigrationPlanRequest represents a request to create a migration plan.
type MigrationPlanRequest struct {
	SourceType string `json:"source_type"`
	TargetType string `json:"target_type"`
}

// MigrationPlanResult represents the result of a migration plan creation.
type MigrationPlanResult struct {
	PlanID string `json:"plan_id"`
	Steps  int    `json:"steps"`
}

// MigrationStepRequest represents a request to execute a migration step.
type MigrationStepRequest struct {
	PlanID string `json:"plan_id"`
	StepID string `json:"step_id"`
}

// MigrationStepResult represents the result of a migration step execution.
type MigrationStepResult struct {
	StepID  string `json:"step_id"`
	Success bool   `json:"success"`
}

// RBACCompatibilityRequest represents a request to evaluate RBAC compatibility.
type RBACCompatibilityRequest struct {
	UserID     uuid.UUID `json:"user_id"`
	ResourceID uuid.UUID `json:"resource_id"`
}

// RBACCompatibilityResult represents the result of an RBAC compatibility evaluation.
type RBACCompatibilityResult struct {
	Compatible bool   `json:"compatible"`
	Decision   string `json:"decision"`
}
