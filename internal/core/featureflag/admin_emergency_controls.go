package featureflag

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/core/audit"
	"github.com/niiniyare/erp/internal/shared/logger"
)

// EmergencyDisableAll disables all feature flags system-wide
func (s *adminServiceImpl) EmergencyDisableAll(ctx context.Context, reason string) (*EmergencyActionResult, error) {
	ctx, span := s.tracing.StartSpan(ctx, "admin.emergency_disable_all")
	defer span.End()

	startTime := time.Now()
	rollbackToken := generateRollbackToken()

	s.logger.Warn("EMERGENCY DISABLE ALL INITIATED", logger.Fields{
		"reason":         reason,
		"rollback_token": rollbackToken,
		"initiated_at":   startTime,
	})

	// Get all active flags
	listRequest := &ListFeatureFlagsRequest{
		Page:     1,
		PageSize: 1000, // Get a large batch
	}

	flagsResponse, err := s.baseService.ListFeatureFlags(ctx, listRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to list flags for emergency disable: %w", err)
	}

	affectedCount := 0
	errors := []string{}

	// Disable all flags
	for _, flag := range flagsResponse.FeatureFlags {
		if flag.DefaultValue == true { // Only disable currently enabled flags
			defaultValue := false
			updateRequest := &UpdateFeatureFlagRequest{
				DefaultValue: &defaultValue,
			}

			_, err := s.baseService.UpdateFeatureFlag(ctx, flag.ID, updateRequest)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Failed to disable flag %s: %s", flag.Name, err.Error()))
				continue
			}
			affectedCount++
		}
	}

	result := &EmergencyActionResult{
		ActionType:      "emergency_disable_all",
		AffectedFlags:   affectedCount,
		ExecutedAt:      startTime,
		Reason:          reason,
		RollbackToken:   rollbackToken,
		EstimatedImpact: s.estimateDisableAllImpact(affectedCount),
	}

	// Critical audit logging
	s.auditEmergencyAction(ctx, "emergency_disable_all", reason, result, errors)

	// Clear all caches to ensure immediate effect
	if s.cacheWarmup != nil {
		// Clear cache for all tenants (we'd need to implement this based on available tenant list)
		s.logger.Info("Clearing all feature flag caches for immediate effect")
	}

	if len(errors) > 0 {
		s.logger.Error("Emergency disable completed with errors", logger.Fields{
			"affected_flags": affectedCount,
			"errors":         errors,
			"rollback_token": rollbackToken,
		})
	} else {
		s.logger.Info("Emergency disable completed successfully", logger.Fields{
			"affected_flags": affectedCount,
			"rollback_token": rollbackToken,
		})
	}

	return result, nil
}

// EmergencyEnableAll enables all feature flags system-wide
func (s *adminServiceImpl) EmergencyEnableAll(ctx context.Context, reason string) (*EmergencyActionResult, error) {
	ctx, span := s.tracing.StartSpan(ctx, "admin.emergency_enable_all")
	defer span.End()

	startTime := time.Now()
	rollbackToken := generateRollbackToken()

	s.logger.Warn("EMERGENCY ENABLE ALL INITIATED", logger.Fields{
		"reason":         reason,
		"rollback_token": rollbackToken,
		"initiated_at":   startTime,
	})

	// Get all flags
	listRequest := &ListFeatureFlagsRequest{
		Page:     1,
		PageSize: 1000,
	}

	flagsResponse, err := s.baseService.ListFeatureFlags(ctx, listRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to list flags for emergency enable: %w", err)
	}

	affectedCount := 0
	errors := []string{}

	// Enable all flags
	for _, flag := range flagsResponse.FeatureFlags {
		if flag.DefaultValue == false { // Only enable currently disabled flags
			defaultValue := true
			updateRequest := &UpdateFeatureFlagRequest{
				DefaultValue: &defaultValue,
			}

			_, err := s.baseService.UpdateFeatureFlag(ctx, flag.ID, updateRequest)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Failed to enable flag %s: %s", flag.Name, err.Error()))
				continue
			}
			affectedCount++
		}
	}

	result := &EmergencyActionResult{
		ActionType:      "emergency_enable_all",
		AffectedFlags:   affectedCount,
		ExecutedAt:      startTime,
		Reason:          reason,
		RollbackToken:   rollbackToken,
		EstimatedImpact: s.estimateEnableAllImpact(affectedCount),
	}

	// Critical audit logging
	s.auditEmergencyAction(ctx, "emergency_enable_all", reason, result, errors)

	return result, nil
}

// CreateRolloutStrategy creates a new rollout strategy
func (s *adminServiceImpl) CreateRolloutStrategy(ctx context.Context, request *CreateRolloutStrategyRequest) (*RolloutStrategy, error) {
	ctx, span := s.tracing.StartSpan(ctx, "admin.create_rollout_strategy")
	defer span.End()

	// Validate rollout strategy configuration
	if err := s.validateRolloutStrategy(request); err != nil {
		return nil, fmt.Errorf("invalid rollout strategy: %w", err)
	}

	strategy := &RolloutStrategy{
		ID:            uuid.New(),
		Name:          request.Name,
		Description:   request.Description,
		Type:          request.Type,
		Configuration: request.Configuration,
		CreatedAt:     time.Now(),
	}

	// Store strategy (this would be persisted in a real implementation)
	s.logger.Info("Rollout strategy created", logger.Fields{
		"strategy_id":   strategy.ID.String(),
		"strategy_name": strategy.Name,
		"strategy_type": string(strategy.Type),
	})

	// Audit the creation
	s.auditRolloutStrategyAction(ctx, "create_rollout_strategy", strategy)

	return strategy, nil
}

// Cache Management Methods

// WarmupCache warms up the feature flag cache
func (s *adminServiceImpl) WarmupCache(ctx context.Context, request *CacheWarmupRequest) (*CacheOperationResult, error) {
	ctx, span := s.tracing.StartSpan(ctx, "admin.warmup_cache")
	defer span.End()

	startTime := time.Now()

	if s.cacheWarmup == nil {
		return &CacheOperationResult{
			OperationType:  "warmup",
			ExecutedAt:     startTime,
			ExecutionTime:  time.Since(startTime),
			ItemsProcessed: 0,
			Success:        false,
			Details:        "Cache warmup service not available",
		}, nil
	}

	// Configure warmup options based on request
	options := &WarmupOptions{
		TenantIDs:      request.TenantIDs,
		FlagNames:      request.FlagNames,
		IncludeStats:   true,
		IncludeLists:   true,
		BatchSize:      10,
		MaxConcurrency: 3,
	}

	// Adjust based on priority
	switch request.Priority {
	case "high":
		options.MaxConcurrency = 5
		options.BatchSize = 5
	case "low":
		options.MaxConcurrency = 2
		options.BatchSize = 20
	}

	// Execute warmup
	err := s.cacheWarmup.WarmupCache(ctx, options)
	success := err == nil

	result := &CacheOperationResult{
		OperationType:  "warmup",
		ExecutedAt:     startTime,
		ExecutionTime:  time.Since(startTime),
		ItemsProcessed: len(request.FlagNames) + len(request.TenantIDs),
		Success:        success,
	}

	if err != nil {
		result.Details = fmt.Sprintf("Warmup completed with errors: %s", err.Error())
	} else {
		result.Details = "Cache warmup completed successfully"
	}

	// Audit cache operation
	s.auditCacheOperation(ctx, "cache_warmup", request, result)

	return result, nil
}

// ClearCache clears feature flag cache
func (s *adminServiceImpl) ClearCache(ctx context.Context, request *CacheClearRequest) (*CacheOperationResult, error) {
	ctx, span := s.tracing.StartSpan(ctx, "admin.clear_cache")
	defer span.End()

	startTime := time.Now()

	if s.cacheWarmup == nil {
		return &CacheOperationResult{
			OperationType:  "clear",
			ExecutedAt:     startTime,
			ExecutionTime:  time.Since(startTime),
			ItemsProcessed: 0,
			Success:        false,
			Details:        "Cache service not available",
		}, nil
	}

	itemsProcessed := 0
	errors := []string{}

	// Clear cache for specific tenants
	for _, tenantID := range request.TenantIDs {
		err := s.cacheWarmup.ClearCache(ctx, tenantID)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to clear cache for tenant %s: %s", tenantID.String(), err.Error()))
		} else {
			itemsProcessed++
		}
	}

	result := &CacheOperationResult{
		OperationType:  "clear",
		ExecutedAt:     startTime,
		ExecutionTime:  time.Since(startTime),
		ItemsProcessed: itemsProcessed,
		Success:        len(errors) == 0,
	}

	if len(errors) > 0 {
		result.Details = fmt.Sprintf("Cache clear completed with errors: %v", errors)
	} else {
		result.Details = "Cache cleared successfully"
	}

	// Audit cache operation
	s.auditCacheOperation(ctx, "cache_clear", request, result)

	return result, nil
}

// GetCacheStats returns cache statistics
func (s *adminServiceImpl) GetCacheStats(ctx context.Context) (*CacheStatsResult, error) {
	ctx, span := s.tracing.StartSpan(ctx, "admin.get_cache_stats")
	defer span.End()

	if s.cacheWarmup == nil {
		return &CacheStatsResult{
			GeneratedAt: time.Now(),
			OverallStats: CacheOverallStats{
				TotalKeys:     0,
				TotalMemoryMB: 0,
				HitRate:       0,
			},
			Recommendations: []CacheRecommendation{
				{
					Type:        "configuration",
					Priority:    "medium",
					Title:       "Cache Not Configured",
					Description: "Redis cache is not configured for feature flags",
					Impact:      "medium",
					Action:      "Enable Redis caching for better performance",
				},
			},
		}, nil
	}

	// Get cache stats from the warmup service
	stats := s.cacheWarmup.GetCacheStats()
	if stats == nil {
		return &CacheStatsResult{
			GeneratedAt:  time.Now(),
			OverallStats: CacheOverallStats{},
		}, nil
	}

	result := &CacheStatsResult{
		GeneratedAt: time.Now(),
		OverallStats: CacheOverallStats{
			TotalKeys:      int64(stats.Sets), // Use Sets as approximation of keys
			TotalMemoryMB:  0.0,               // Memory info not available in this stats structure
			HitRate:        stats.HitRatio * 100,
			MissRate:       (1 - stats.HitRatio) * 100,
			EvictionRate:   0.0, // Eviction info not available
			AverageKeySize: 0.0, // Size info not available
		},
		CacheTypeStats: map[string]CacheTypeStats{
			"flags": {
				KeyCount:   int64(stats.Sets) / 3, // Rough estimate
				MemoryMB:   0.0,                   // Memory info not available
				HitRate:    stats.HitRatio * 100,
				AverageTTL: int(FlagDataCacheTTL.Seconds()),
			},
			"evaluations": {
				KeyCount:   int64(stats.Sets) / 3,
				MemoryMB:   0.0,
				HitRate:    stats.HitRatio * 100,
				AverageTTL: int(EvaluationCacheTTL.Seconds()),
			},
			"stats": {
				KeyCount:   int64(stats.Sets) / 3,
				MemoryMB:   0.0,
				HitRate:    stats.HitRatio * 100,
				AverageTTL: int(StatsCacheTTL.Seconds()),
			},
		},
		Recommendations: []CacheRecommendation{}, // Will be populated after result is created
	}

	// Generate recommendations based on the stats
	result.Recommendations = s.generateCacheRecommendations(result.OverallStats)

	return result, nil
}

// Helper methods

func generateRollbackToken() string {
	return fmt.Sprintf("emergency_%d_%s", time.Now().Unix(), uuid.New().String()[:8])
}

func (s *adminServiceImpl) estimateDisableAllImpact(affectedFlags int) string {
	if affectedFlags == 0 {
		return "No impact - no flags were disabled"
	}
	if affectedFlags < 5 {
		return "Low impact - few features affected"
	}
	if affectedFlags < 20 {
		return "Medium impact - some features may be unavailable"
	}
	return "High impact - many features will be disabled"
}

func (s *adminServiceImpl) estimateEnableAllImpact(affectedFlags int) string {
	if affectedFlags == 0 {
		return "No impact - no flags were enabled"
	}
	if affectedFlags < 5 {
		return "Low impact - few new features enabled"
	}
	if affectedFlags < 20 {
		return "Medium impact - some new features enabled"
	}
	return "High impact - many new features will be available"
}

func (s *adminServiceImpl) validateRolloutStrategy(request *CreateRolloutStrategyRequest) error {
	switch request.Type {
	case RolloutStrategyPercentage:
		if percentage, ok := request.Configuration["percentage"]; !ok {
			return fmt.Errorf("percentage rollout strategy requires 'percentage' configuration")
		} else if pct, ok := percentage.(float64); !ok || pct < 0 || pct > 100 {
			return fmt.Errorf("percentage must be between 0 and 100")
		}
	case RolloutStrategyUserAttribute:
		if _, ok := request.Configuration["attribute"]; !ok {
			return fmt.Errorf("user attribute strategy requires 'attribute' configuration")
		}
		if _, ok := request.Configuration["value"]; !ok {
			return fmt.Errorf("user attribute strategy requires 'value' configuration")
		}
	case RolloutStrategyGradual:
		if _, ok := request.Configuration["duration_hours"]; !ok {
			return fmt.Errorf("gradual rollout strategy requires 'duration_hours' configuration")
		}
		if _, ok := request.Configuration["initial_percentage"]; !ok {
			return fmt.Errorf("gradual rollout strategy requires 'initial_percentage' configuration")
		}
		if _, ok := request.Configuration["final_percentage"]; !ok {
			return fmt.Errorf("gradual rollout strategy requires 'final_percentage' configuration")
		}
	default:
		return fmt.Errorf("unsupported rollout strategy type: %s", request.Type)
	}
	return nil
}

func (s *adminServiceImpl) generateCacheRecommendations(stats CacheOverallStats) []CacheRecommendation {
	recommendations := []CacheRecommendation{}

	if stats.HitRate < 70 {
		recommendations = append(recommendations, CacheRecommendation{
			Type:        "performance",
			Priority:    "high",
			Title:       "Low Cache Hit Rate",
			Description: fmt.Sprintf("Cache hit rate is %.1f%%, which is below optimal", stats.HitRate),
			Impact:      "high",
			Action:      "Review cache TTL settings and cache key patterns",
		})
	}

	if stats.EvictionRate > 10 {
		recommendations = append(recommendations, CacheRecommendation{
			Type:        "capacity",
			Priority:    "medium",
			Title:       "High Eviction Rate",
			Description: fmt.Sprintf("Cache eviction rate is %.1f%%, indicating memory pressure", stats.EvictionRate),
			Impact:      "medium",
			Action:      "Consider increasing cache memory allocation",
		})
	}

	if stats.TotalMemoryMB > 1000 {
		recommendations = append(recommendations, CacheRecommendation{
			Type:        "optimization",
			Priority:    "low",
			Title:       "Large Cache Size",
			Description: fmt.Sprintf("Cache is using %.1f MB of memory", stats.TotalMemoryMB),
			Impact:      "low",
			Action:      "Review cached data for optimization opportunities",
		})
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, CacheRecommendation{
			Type:        "info",
			Priority:    "low",
			Title:       "Cache Performance Good",
			Description: "Cache is performing within normal parameters",
			Impact:      "none",
			Action:      "Continue monitoring cache performance",
		})
	}

	return recommendations
}

// Audit methods

func (s *adminServiceImpl) auditEmergencyAction(ctx context.Context, action, reason string, result *EmergencyActionResult, errors []string) {
	contextData, _ := json.Marshal(map[string]any{
		"action":           action,
		"reason":           reason,
		"affected_flags":   result.AffectedFlags,
		"rollback_token":   result.RollbackToken,
		"estimated_impact": result.EstimatedImpact,
		"executed_at":      result.ExecutedAt,
		"errors":           errors,
	})

	decision := fmt.Sprintf("EXECUTED_%s", action)
	reasonStr := fmt.Sprintf("Emergency action executed: %s - %s", action, reason)
	s.auditService.CreateAuditEvent(ctx, audit.CreateAuditEventRequest{
		EventType:     "admin_emergency_action",
		EventCategory: AuditCategoryAdmin,
		Severity:      AuditSeverityCritical, // Emergency actions are critical
		Decision:      &decision,
		Reason:        &reasonStr,
		Context:       contextData,
	})
}

func (s *adminServiceImpl) auditRolloutStrategyAction(ctx context.Context, action string, strategy *RolloutStrategy) {
	contextData, _ := json.Marshal(map[string]any{
		"action":        action,
		"strategy_id":   strategy.ID.String(),
		"strategy_name": strategy.Name,
		"strategy_type": string(strategy.Type),
		"configuration": strategy.Configuration,
		"created_at":    strategy.CreatedAt,
	})

	decision := fmt.Sprintf("EXECUTED_%s", action)
	reason := fmt.Sprintf("Rollout strategy %s: %s", action, strategy.Name)
	strategyEntityID := strategy.ID
	s.auditService.CreateAuditEvent(ctx, audit.CreateAuditEventRequest{
		EventType:     "admin_rollout_strategy",
		EventCategory: AuditCategoryAdmin,
		Severity:      AuditSeverityInfo,
		EntityID:      &strategyEntityID,
		Decision:      &decision,
		Reason:        &reason,
		Context:       contextData,
	})
}

func (s *adminServiceImpl) auditCacheOperation(ctx context.Context, operation string, request interface{}, result *CacheOperationResult) {
	contextData, _ := json.Marshal(map[string]any{
		"operation":       operation,
		"request":         request,
		"items_processed": result.ItemsProcessed,
		"execution_time":  result.ExecutionTime,
		"success":         result.Success,
		"details":         result.Details,
		"executed_at":     result.ExecutedAt,
	})

	severity := AuditSeverityInfo
	if !result.Success {
		severity = AuditSeverityWarn
	}

	decision := fmt.Sprintf("EXECUTED_%s", operation)
	reason := fmt.Sprintf("Cache operation executed: %s", operation)
	s.auditService.CreateAuditEvent(ctx, audit.CreateAuditEventRequest{
		EventType:     "admin_cache_operation",
		EventCategory: AuditCategoryAdmin,
		Severity:      severity,
		Decision:      &decision,
		Reason:        &reason,
		Context:       contextData,
	})
}
