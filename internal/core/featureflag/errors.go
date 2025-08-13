package featureflag

import "errors"

// Domain-specific errors for feature flags
var (
	// ErrFeatureFlagNotFound is returned when a feature flag is not found
	ErrFeatureFlagNotFound = errors.New("feature flag not found")

	// ErrFeatureFlagAlreadyExists is returned when trying to create a flag that already exists
	ErrFeatureFlagAlreadyExists = errors.New("feature flag already exists")

	// ErrTenantOverrideNotFound is returned when a tenant override is not found
	ErrTenantOverrideNotFound = errors.New("tenant override not found")

	// ErrTenantOverrideAlreadyExists is returned when trying to create an override that already exists
	ErrTenantOverrideAlreadyExists = errors.New("tenant override already exists")

	// ErrInvalidFlagType is returned when an invalid flag type is provided
	ErrInvalidFlagType = errors.New("invalid flag type")

	// ErrInvalidRolloutPercentage is returned when rollout percentage is out of range
	ErrInvalidRolloutPercentage = errors.New("rollout percentage must be between 0 and 100")

	// ErrInvalidEvaluationContext is returned when evaluation context is invalid
	ErrInvalidEvaluationContext = errors.New("invalid evaluation context")

	// ErrFlagEvaluationFailed is returned when flag evaluation fails
	ErrFlagEvaluationFailed = errors.New("flag evaluation failed")

	// ML Optimization errors
	ErrInsufficientHistoricalData = errors.New("insufficient historical data for ML optimization")
	ErrMLOptimizationFailed       = errors.New("ML optimization analysis failed")
	ErrAnomalyDetectionFailed     = errors.New("anomaly detection failed")
	ErrInvalidMLConfiguration     = errors.New("invalid ML optimization configuration")
	ErrMLModelNotReady            = errors.New("ML model not ready for predictions")
	ErrStatisticalAnalysisFailed  = errors.New("statistical analysis failed")

	// A/B Test Analytics errors
	ErrInvalidStatisticalTest     = errors.New("invalid statistical test configuration")
	ErrInsufficientSampleSize     = errors.New("insufficient sample size for statistical significance")
	ErrInvalidConfidenceLevel     = errors.New("confidence level must be between 0 and 1")
	ErrInvalidVariantData         = errors.New("invalid or missing variant data")
	ErrExperimentNotFound         = errors.New("experiment not found")
	ErrStatisticalTestFailed      = errors.New("statistical test execution failed")
	ErrBayesianAnalysisFailed     = errors.New("Bayesian analysis failed")
	ErrPowerAnalysisFailed        = errors.New("power analysis failed")
)
