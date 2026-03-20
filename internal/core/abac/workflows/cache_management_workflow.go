package workflows

import (
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"awo/internal/core/abac/activities"
)

// CacheCleanupWorkflowName is the name of the cache cleanup workflow
const CacheCleanupWorkflowName = "CacheCleanupWorkflow"

// CacheCleanupWorkflowInput represents input for cache cleanup workflow
type CacheCleanupWorkflowInput struct {
	MaxAge               time.Duration `json:"max_age"`
	BatchSize            int32         `json:"batch_size"`
	CleanupIntervalHours int           `json:"cleanup_interval_hours"`
	RequestID            string        `json:"request_id"`
}

// CacheCleanupWorkflowOutput represents output of cache cleanup workflow
type CacheCleanupWorkflowOutput struct {
	CleanupRuns     int       `json:"cleanup_runs"`
	TotalCleaned    int64     `json:"total_cleaned"`
	LastCleanupTime time.Time `json:"last_cleanup_time"`
}

// CacheCleanupWorkflow runs periodic cache cleanup operations
func CacheCleanupWorkflow(ctx workflow.Context, input CacheCleanupWorkflowInput) (*CacheCleanupWorkflowOutput, error) {
	logger := workflow.GetLogger(ctx)

	logger.Info("Starting cache cleanup workflow",
		"max_age", input.MaxAge,
		"batch_size", input.BatchSize,
		"cleanup_interval_hours", input.CleanupIntervalHours,
		"request_id", input.RequestID)

	// Set default values
	if input.MaxAge == 0 {
		input.MaxAge = 24 * time.Hour // Default 24 hours
	}
	if input.BatchSize == 0 {
		input.BatchSize = 1000 // Default batch size
	}
	if input.CleanupIntervalHours == 0 {
		input.CleanupIntervalHours = 4 // Default every 4 hours
	}

	cleanupRuns := 0
	totalCleaned := int64(0)

	// Set activity options for cleanup operations
	cleanupActivityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    10 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    2 * time.Minute,
			MaximumAttempts:    3,
		},
	}

	// Run continuous cleanup loop
	for {
		cleanupCtx := workflow.WithActivityOptions(ctx, cleanupActivityOptions)

		// Perform cleanup
		cleanupInput := &activities.CleanupExpiredCacheInput{
			MaxAge:    input.MaxAge,
			BatchSize: input.BatchSize,
			RequestID: input.RequestID,
		}

		var cleanupResult *activities.CleanupExpiredCacheOutput
		err := workflow.ExecuteActivity(cleanupCtx, "CacheActivities.CleanupExpiredCache", cleanupInput).Get(ctx, &cleanupResult)
		if err != nil {
			logger.Error("Cache cleanup failed", "error", err, "run", cleanupRuns)
		} else {
			totalCleaned += cleanupResult.CleanedCount
			cleanupRuns++

			logger.Info("Cache cleanup completed",
				"run", cleanupRuns,
				"cleaned_count", cleanupResult.CleanedCount,
				"total_cleaned", totalCleaned)
		}

		// Wait for next cleanup interval
		err = workflow.Sleep(ctx, time.Duration(input.CleanupIntervalHours)*time.Hour)
		if err != nil {
			// Workflow was cancelled or completed
			break
		}
	}

	return &CacheCleanupWorkflowOutput{
		CleanupRuns:     cleanupRuns,
		TotalCleaned:    totalCleaned,
		LastCleanupTime: workflow.Now(ctx),
	}, nil
}

// CacheWarmupWorkflowName is the name of the cache warmup workflow
const CacheWarmupWorkflowName = "CacheWarmupWorkflow"

// CacheWarmupWorkflowInput represents input for cache warmup workflow
type CacheWarmupWorkflowInput struct {
	WarmupConfigs []*CacheWarmupConfig `json:"warmup_configs"`
	RequestID     string               `json:"request_id"`
}

// CacheWarmupConfig represents configuration for cache warmup
type CacheWarmupConfig struct {
	UserIDs       []string `json:"user_ids"`
	ResourceTypes []string `json:"resource_types"`
	Actions       []string `json:"actions"`
	Priority      string   `json:"priority"` // high, medium, low
	BatchSize     int      `json:"batch_size"`
}

// CacheWarmupWorkflowOutput represents output of cache warmup workflow
type CacheWarmupWorkflowOutput struct {
	TotalWarmed      int64 `json:"total_warmed"`
	TotalSkipped     int64 `json:"total_skipped"`
	ConfigsProcessed int   `json:"configs_processed"`
}

// CacheWarmupWorkflow pre-loads frequently accessed policy evaluations into cache
func CacheWarmupWorkflow(ctx workflow.Context, input CacheWarmupWorkflowInput) (*CacheWarmupWorkflowOutput, error) {
	logger := workflow.GetLogger(ctx)

	logger.Info("Starting cache warmup workflow",
		"config_count", len(input.WarmupConfigs),
		"request_id", input.RequestID)

	totalWarmed := int64(0)
	totalSkipped := int64(0)
	configsProcessed := 0

	// Set activity options for warmup operations
	warmupActivityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    5 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    2,
		},
	}
	warmupCtx := workflow.WithActivityOptions(ctx, warmupActivityOptions)

	// Process each warmup configuration
	for i, config := range input.WarmupConfigs {
		logger.Info("Processing warmup configuration",
			"config_index", i,
			"priority", config.Priority,
			"user_count", len(config.UserIDs),
			"resource_types", len(config.ResourceTypes),
			"actions", len(config.Actions))

		// Convert string UUIDs to proper UUID format (if needed)
		warmupInput := &activities.WarmupCacheInput{
			UserIDs:       convertStringUUIDs(config.UserIDs),
			ResourceTypes: config.ResourceTypes,
			Actions:       config.Actions,
			Priority:      config.Priority,
			RequestID:     input.RequestID,
		}

		var warmupResult *activities.WarmupCacheOutput
		err := workflow.ExecuteActivity(warmupCtx, "CacheActivities.WarmupCache", warmupInput).Get(ctx, &warmupResult)
		if err != nil {
			logger.Error("Cache warmup failed for configuration",
				"config_index", i,
				"error", err)
			continue
		}

		totalWarmed += warmupResult.WarmedCount
		totalSkipped += warmupResult.SkippedCount
		configsProcessed++

		logger.Info("Warmup configuration completed",
			"config_index", i,
			"warmed_count", warmupResult.WarmedCount,
			"skipped_count", warmupResult.SkippedCount)

		// Add small delay between configurations to avoid overwhelming the system
		if i < len(input.WarmupConfigs)-1 {
			_ = workflow.Sleep(ctx, 1*time.Second)
		}
	}

	logger.Info("Cache warmup workflow completed",
		"total_warmed", totalWarmed,
		"total_skipped", totalSkipped,
		"configs_processed", configsProcessed,
		"request_id", input.RequestID)

	return &CacheWarmupWorkflowOutput{
		TotalWarmed:      totalWarmed,
		TotalSkipped:     totalSkipped,
		ConfigsProcessed: configsProcessed,
	}, nil
}

// CacheInvalidationWorkflowName is the name of the cache invalidation workflow
const CacheInvalidationWorkflowName = "CacheInvalidationWorkflow"

// CacheInvalidationWorkflowInput represents input for cache invalidation workflow
type CacheInvalidationWorkflowInput struct {
	InvalidationRequests []*activities.InvalidatePolicyCacheInput `json:"invalidation_requests"`
	RequestID            string                                   `json:"request_id"`
}

// CacheInvalidationWorkflowOutput represents output of cache invalidation workflow
type CacheInvalidationWorkflowOutput struct {
	TotalInvalidated  int64 `json:"total_invalidated"`
	RequestsProcessed int   `json:"requests_processed"`
	FailedRequests    int   `json:"failed_requests"`
}

// CacheInvalidationWorkflow handles bulk cache invalidation operations
func CacheInvalidationWorkflow(ctx workflow.Context, input CacheInvalidationWorkflowInput) (*CacheInvalidationWorkflowOutput, error) {
	logger := workflow.GetLogger(ctx)

	logger.Info("Starting cache invalidation workflow",
		"request_count", len(input.InvalidationRequests),
		"request_id", input.RequestID)

	totalInvalidated := int64(0)
	requestsProcessed := 0
	failedRequests := 0

	// Set activity options for invalidation operations
	invalidationActivityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    2 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    20 * time.Second,
			MaximumAttempts:    3,
		},
	}
	invalidationCtx := workflow.WithActivityOptions(ctx, invalidationActivityOptions)

	// Process invalidation requests concurrently
	futures := make([]workflow.Future, len(input.InvalidationRequests))

	for i, invalidationReq := range input.InvalidationRequests {
		// Set request ID for tracking
		invalidationReq.RequestID = input.RequestID

		// Execute invalidation activity
		future := workflow.ExecuteActivity(invalidationCtx, "CacheActivities.InvalidatePolicyCache", invalidationReq)
		futures[i] = future
	}

	// Wait for all invalidations to complete
	for i, future := range futures {
		var invalidationResult *activities.InvalidatePolicyCacheOutput
		err := future.Get(ctx, &invalidationResult)
		if err != nil {
			logger.Error("Cache invalidation failed",
				"request_index", i,
				"error", err)
			failedRequests++
		} else {
			totalInvalidated += invalidationResult.InvalidatedCount
			requestsProcessed++

			logger.Debug("Cache invalidation completed",
				"request_index", i,
				"invalidated_count", invalidationResult.InvalidatedCount)
		}
	}

	logger.Info("Cache invalidation workflow completed",
		"total_invalidated", totalInvalidated,
		"requests_processed", requestsProcessed,
		"failed_requests", failedRequests,
		"request_id", input.RequestID)

	return &CacheInvalidationWorkflowOutput{
		TotalInvalidated:  totalInvalidated,
		RequestsProcessed: requestsProcessed,
		FailedRequests:    failedRequests,
	}, nil
}

// Helper functions

// convertStringUUIDs converts string UUIDs to UUID objects (placeholder implementation)
func convertStringUUIDs(stringUUIDs []string) []uuid.UUID {
	// This is a placeholder implementation
	// In a real implementation, you would parse the string UUIDs
	uuids := make([]uuid.UUID, 0, len(stringUUIDs))
	for _, s := range stringUUIDs {
		if parsed, err := uuid.Parse(s); err == nil {
			uuids = append(uuids, parsed)
		}
	}
	return uuids
}
