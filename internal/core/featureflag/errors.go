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
)
