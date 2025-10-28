package featureflag

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/shared/logger"
)

// CacheWarmer provides utilities for warming up the feature flag cache
type CacheWarmer struct {
	service CachedFeatureFlagService
	logger  logger.Logger
}

// NewCacheWarmer creates a new cache warmer
func NewCacheWarmer(service *CachedFeatureFlagService, logger logger.Logger) *CacheWarmer {
	return &CacheWarmer{
		service: *service,
		logger:  logger,
	}
}

// WarmupOptions configures cache warmup behavior
type WarmupOptions struct {
	TenantIDs          []uuid.UUID          `json:"tenant_ids,omitempty"`
	FlagNames          []string             `json:"flag_names,omitempty"`
	IncludeStats       bool                 `json:"include_stats"`
	IncludeLists       bool                 `json:"include_lists"`
	BatchSize          int                  `json:"batch_size"`
	MaxConcurrency     int                  `json:"max_concurrency"`
	EvaluationContexts []*EvaluationContext `json:"evaluation_contexts,omitempty"`
}

// DefaultWarmupOptions returns sensible defaults for cache warmup
func DefaultWarmupOptions() *WarmupOptions {
	return &WarmupOptions{
		IncludeStats:   true,
		IncludeLists:   true,
		BatchSize:      10,
		MaxConcurrency: 3,
	}
}

// WarmupCache preloads frequently accessed data into the cache
func (w *CacheWarmer) WarmupCache(ctx context.Context, opts *WarmupOptions) error {
	if opts == nil {
		opts = DefaultWarmupOptions()
	}

	w.logger.Info("Starting feature flag cache warmup", logger.Fields{
		"tenant_count":    len(opts.TenantIDs),
		"flag_count":      len(opts.FlagNames),
		"include_stats":   opts.IncludeStats,
		"include_lists":   opts.IncludeLists,
		"batch_size":      opts.BatchSize,
		"max_concurrency": opts.MaxConcurrency,
	})

	startTime := time.Now()

	// Warmup individual flags
	if len(opts.FlagNames) > 0 {
		if err := w.warmupIndividualFlags(ctx, opts.FlagNames); err != nil {
			return fmt.Errorf("failed to warmup individual flags: %w", err)
		}
	}

	// Warmup flag lists for each tenant
	if opts.IncludeLists {
		if err := w.warmupFlagLists(ctx, opts.TenantIDs); err != nil {
			return fmt.Errorf("failed to warmup flag lists: %w", err)
		}
	}

	// Warmup stats for each tenant
	if opts.IncludeStats {
		if err := w.warmupStats(ctx, opts.TenantIDs); err != nil {
			return fmt.Errorf("failed to warmup stats: %w", err)
		}
	}

	// Warmup evaluations if contexts provided
	if len(opts.EvaluationContexts) > 0 && len(opts.FlagNames) > 0 {
		if err := w.warmupEvaluations(ctx, opts.FlagNames, opts.EvaluationContexts); err != nil {
			return fmt.Errorf("failed to warmup evaluations: %w", err)
		}
	}

	duration := time.Since(startTime)
	w.logger.Info("Feature flag cache warmup completed", logger.Fields{
		"duration_ms": duration.Milliseconds(),
		"duration":    duration.String(),
	})

	return nil
}

// WarmupForTenant warms up cache for a specific tenant
func (w *CacheWarmer) WarmupForTenant(ctx context.Context, tenantID uuid.UUID) error {
	w.logger.Info("Warming up cache for tenant", logger.Fields{
		"tenant_id": tenantID.String(),
	})

	// Add tenant context
	ctx = context.WithValue(ctx, "tenant_id", tenantID)

	// Get all flags for the tenant (this will cache them)
	listRequest := &ListFeatureFlagsRequest{
		Page:     1,
		PageSize: 100, // Get first 100 flags
	}

	flags, err := w.service.ListFeatureFlags(ctx, listRequest)
	if err != nil {
		return fmt.Errorf("failed to list flags for tenant %s: %w", tenantID.String(), err)
	}

	w.logger.Info("Cached flag list for tenant", logger.Fields{
		"tenant_id":  tenantID.String(),
		"flag_count": len(flags.FeatureFlags),
	})

	// Cache individual flags
	for _, flag := range flags.FeatureFlags {
		_, err := w.service.GetFeatureFlag(ctx, flag.Name)
		if err != nil {
			w.logger.Warn("Failed to cache individual flag", logger.Fields{
				"tenant_id": tenantID.String(),
				"flag_name": flag.Name,
				"error":     err.Error(),
			})
		}
	}

	// Cache stats
	_, err = w.service.GetFlagStats(ctx)
	if err != nil {
		w.logger.Warn("Failed to cache stats for tenant", logger.Fields{
			"tenant_id": tenantID.String(),
			"error":     err.Error(),
		})
	}

	return nil
}

// warmupIndividualFlags loads individual flags into cache
func (w *CacheWarmer) warmupIndividualFlags(ctx context.Context, flagNames []string) error {
	for _, name := range flagNames {
		_, err := w.service.GetFeatureFlag(ctx, name)
		if err != nil {
			w.logger.Warn("Failed to warmup flag", logger.Fields{
				"flag_name": name,
				"error":     err.Error(),
			})
			continue
		}
	}
	return nil
}

// warmupFlagLists loads flag lists into cache
func (w *CacheWarmer) warmupFlagLists(ctx context.Context, tenantIDs []uuid.UUID) error {
	for _, tenantID := range tenantIDs {
		tenantCtx := context.WithValue(ctx, "tenant_id", tenantID)

		// Cache different list configurations
		listConfigs := []*ListFeatureFlagsRequest{
			{Page: 1, PageSize: 20},                                 // Default pagination
			{Page: 1, PageSize: 50},                                 // Larger page
			{Page: 1, PageSize: 20, FlagType: stringPtr("boolean")}, // Boolean flags
			{Page: 1, PageSize: 20, FlagType: stringPtr("string")},  // String flags
		}

		for _, config := range listConfigs {
			_, err := w.service.ListFeatureFlags(tenantCtx, config)
			if err != nil {
				w.logger.Warn("Failed to warmup flag list", logger.Fields{
					"tenant_id": tenantID.String(),
					"config":    fmt.Sprintf("%+v", config),
					"error":     err.Error(),
				})
			}
		}
	}
	return nil
}

// warmupStats loads stats into cache
func (w *CacheWarmer) warmupStats(ctx context.Context, tenantIDs []uuid.UUID) error {
	for _, tenantID := range tenantIDs {
		tenantCtx := context.WithValue(ctx, "tenant_id", tenantID)

		_, err := w.service.GetFlagStats(tenantCtx)
		if err != nil {
			w.logger.Warn("Failed to warmup stats", logger.Fields{
				"tenant_id": tenantID.String(),
				"error":     err.Error(),
			})
		}
	}
	return nil
}

// warmupEvaluations loads evaluations into cache
func (w *CacheWarmer) warmupEvaluations(ctx context.Context, flagNames []string, evalCtxs []*EvaluationContext) error {
	for _, evalCtx := range evalCtxs {
		// Single evaluations
		for _, flagName := range flagNames {
			_, err := w.service.EvaluateFlag(ctx, flagName, evalCtx)
			if err != nil {
				w.logger.Warn("Failed to warmup evaluation", logger.Fields{
					"flag_name": flagName,
					"tenant_id": evalCtx.TenantID.String(),
					"error":     err.Error(),
				})
			}
		}

		// Bulk evaluation
		if len(flagNames) > 1 {
			_, err := w.service.EvaluateFlags(ctx, flagNames, evalCtx)
			if err != nil {
				w.logger.Warn("Failed to warmup bulk evaluation", logger.Fields{
					"flag_count": len(flagNames),
					"tenant_id":  evalCtx.TenantID.String(),
					"error":      err.Error(),
				})
			}
		}
	}
	return nil
}

// ClearCache invalidates all cached data
func (w *CacheWarmer) ClearCache(ctx context.Context, tenantID uuid.UUID) error {
	w.logger.Info("Clearing feature flag cache", logger.Fields{
		"tenant_id": tenantID.String(),
	})

	// This would use the cache invalidation patterns from the cached service
	patterns := []string{
		fmt.Sprintf("flag:%s:*", tenantID.String()),
		fmt.Sprintf("flag_by_id:%s:*", tenantID.String()),
		fmt.Sprintf("flag_list:%s:*", tenantID.String()),
		fmt.Sprintf("flag_stats:%s", tenantID.String()),
		fmt.Sprintf("flags_by_type:%s:*", tenantID.String()),
		"eval:*",      // Could be more specific to tenant
		"bulk_eval:*", // Could be more specific to tenant
	}

	for _, pattern := range patterns {
		if err := w.service.cache.DeletePattern(ctx, pattern); err != nil {
			w.logger.Warn("Failed to clear cache pattern", logger.Fields{
				"pattern": pattern,
				"error":   err.Error(),
			})
		}
	}

	return nil
}

// GetCacheStats returns cache statistics
func (w *CacheWarmer) GetCacheStats() *CacheStats {
	if w.service.cache != nil {
		stats := w.service.cache.Stats()
		return &CacheStats{
			Hits:              stats.Hits,
			Misses:            stats.Misses,
			Sets:              stats.Sets,
			Deletes:           stats.Deletes,
			Errors:            stats.Errors,
			HitRatio:          stats.HitRatio,
			AverageLatency:    stats.AverageLatency,
			ConnectionsActive: stats.ConnectionsActive,
			ConnectionsIdle:   stats.ConnectionsIdle,
			MemoryCacheSize:   stats.MemoryCacheSize,
			GlobalCacheSize:   stats.GlobalCacheSize,
			LastError:         stats.LastError,
			LastErrorTime:     stats.LastErrorTime,
		}
	}
	return nil
}

// Helper function
func stringPtr(s string) *string {
	return &s
}

// CacheStats type alias for easier access
type CacheStats = cache.CacheStats
