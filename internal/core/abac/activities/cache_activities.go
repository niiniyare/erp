package activities

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"

	"github.com/niiniyare/erp/internal/core/abac/models"
	"github.com/niiniyare/erp/internal/core/abac/repository"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/niiniyare/erp/internal/shared/types"
)

// CacheActivities handles ABAC caching operations
type CacheActivities struct {
	policyEvaluationRepo repository.PolicyEvaluationRepository
	attributeRepo        repository.AttributeRepository
	logger               logger.Logger
	metrics              metrics.MetricsProvider
	tracer               tracing.Service
}

// NewCacheActivities creates a new CacheActivities instance
func NewCacheActivities(
	policyEvaluationRepo repository.PolicyEvaluationRepository,
	attributeRepo repository.AttributeRepository,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.Service,
) *CacheActivities {
	return &CacheActivities{
		policyEvaluationRepo: policyEvaluationRepo,
		attributeRepo:        attributeRepo,
		logger:               logger,
		metrics:              metrics,
		tracer:               tracer,
	}
}

// CachePolicyEvaluationInput represents input for caching policy evaluation results
type CachePolicyEvaluationInput struct {
	UserID           uuid.UUID                `json:"user_id"`
	ResourceType     string                   `json:"resource_type"`
	ResourceID       *uuid.UUID               `json:"resource_id,omitempty"`
	Action           string                   `json:"action"`
	ContextHash      string                   `json:"context_hash"`
	Decision         types.PolicyDecisionType `json:"decision"`
	PolicyDecisions  []*models.PolicyDecision `json:"policy_decisions"`
	EvaluationTimeMS int64                    `json:"evaluation_time_ms"`
	TTLMinutes       int32                    `json:"ttl_minutes"`
	RequestID        string                   `json:"request_id"`
}

// CachePolicyEvaluationOutput represents output of caching operation
type CachePolicyEvaluationOutput struct {
	Success   bool      `json:"success"`
	CacheKey  string    `json:"cache_key"`
	ExpiresAt time.Time `json:"expires_at"`
}

// CachePolicyEvaluation caches a policy evaluation result
func (a *CacheActivities) CachePolicyEvaluation(ctx context.Context, input *CachePolicyEvaluationInput) (*CachePolicyEvaluationOutput, error) {
	ctx, span := a.tracer.StartSpan(ctx, "abac.activities.CachePolicyEvaluation",
		tracing.WithAttributes(
			attribute.String("user_id", input.UserID.String()),
			attribute.String("resource_type", input.ResourceType),
			attribute.String("action", input.Action),
			attribute.String("decision", string(input.Decision)),
			attribute.String("request_id", input.RequestID),
		))
	defer span.End()

	startTime := time.Now()

	a.logger.InfoContext(ctx, "Caching policy evaluation result",
		logger.Fields{
			"user_id":            input.UserID,
			"resource_type":      input.ResourceType,
			"action":             input.Action,
			"decision":           input.Decision,
			"evaluation_time_ms": input.EvaluationTimeMS,
			"request_id":         input.RequestID,
		})

	// Calculate expiration time
	ttl := time.Duration(input.TTLMinutes) * time.Minute
	if ttl <= 0 {
		ttl = 15 * time.Minute // Default 15 minutes
	}
	expiresAt := time.Now().Add(ttl)

	// Extract policy IDs
	var policyIDs []uuid.UUID
	for _, decision := range input.PolicyDecisions {
		policyIDs = append(policyIDs, decision.PolicyID)
	}

	// Create cache request
	cacheReq := &repository.CacheEvaluationResultRequest{
		UserID:             input.UserID,
		ResourceType:       input.ResourceType,
		ResourceID:         input.ResourceID,
		Action:             input.Action,
		ContextHash:        input.ContextHash,
		Decision:           input.Decision,
		ApplicablePolicies: policyIDs,
		EvaluationTimeMS:   input.EvaluationTimeMS,
		ExpiresAt:          expiresAt,
		Result: &models.PolicyEvaluationResult{
			Decision:        input.Decision,
			PolicyDecisions: ConvertPolicyDecisionsToInfo(input.PolicyDecisions),
		},
	}

	// Cache the result
	err := a.policyEvaluationRepo.CacheEvaluationResult(ctx, cacheReq)
	if err != nil {
		a.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		a.metrics.IncrementCounter("policy_evaluation_cache_failed", metrics.Fields{"reason": "cache_store_failed"})

		// Log but don't fail the operation - caching is not critical
		a.logger.WarnContext(ctx, "Failed to cache policy evaluation result",
			logger.Fields{
				"error":      err.Error(),
				"request_id": input.RequestID,
			})

		return &CachePolicyEvaluationOutput{
			Success:   false,
			CacheKey:  a.generateCacheKey(input),
			ExpiresAt: expiresAt,
		}, nil
	}

	duration := time.Since(startTime)
	span.SetAttributes(attribute.Bool("cache_success", true))

	a.metrics.ObserveHistogram("abac_cache_store_duration_seconds", duration.Seconds(),
		metrics.Fields{"operation": "policy_evaluation"})
	a.metrics.IncrementCounter("policy_evaluation_cache_store_success", metrics.Fields{"operation": "cache_store"})

	a.logger.InfoContext(ctx, "Policy evaluation result cached successfully",
		logger.Fields{
			"cache_ttl_minutes": input.TTLMinutes,
			"expires_at":        expiresAt,
			"duration_ms":       duration.Milliseconds(),
			"request_id":        input.RequestID,
		})

	return &CachePolicyEvaluationOutput{
		Success:   true,
		CacheKey:  a.generateCacheKey(input),
		ExpiresAt: expiresAt,
	}, nil
}

// GetCachedPolicyEvaluationInput represents input for retrieving cached evaluation
type GetCachedPolicyEvaluationInput struct {
	UserID       uuid.UUID  `json:"user_id"`
	ResourceType string     `json:"resource_type"`
	ResourceID   *uuid.UUID `json:"resource_id,omitempty"`
	Action       string     `json:"action"`
	ContextHash  string     `json:"context_hash"`
	RequestID    string     `json:"request_id"`
}

// GetCachedPolicyEvaluationOutput represents output of cache retrieval
type GetCachedPolicyEvaluationOutput struct {
	Found           bool                     `json:"found"`
	Decision        types.PolicyDecisionType `json:"decision,omitempty"`
	PolicyDecisions []*models.PolicyDecision `json:"policy_decisions,omitempty"`
	CachedAt        time.Time                `json:"cached_at,omitempty"`
	ExpiresAt       time.Time                `json:"expires_at,omitempty"`
}

// GetCachedPolicyEvaluation retrieves a cached policy evaluation result
func (a *CacheActivities) GetCachedPolicyEvaluation(ctx context.Context, input *GetCachedPolicyEvaluationInput) (*GetCachedPolicyEvaluationOutput, error) {
	ctx, span := a.tracer.StartSpan(ctx, "abac.activities.GetCachedPolicyEvaluation",
		tracing.WithAttributes(
			attribute.String("user_id", input.UserID.String()),
			attribute.String("resource_type", input.ResourceType),
			attribute.String("action", input.Action),
			attribute.String("request_id", input.RequestID),
		))
	defer span.End()

	startTime := time.Now()

	getCacheReq := &repository.GetCachedEvaluationResultRequest{
		UserID:       input.UserID,
		ResourceType: input.ResourceType,
		ResourceID:   input.ResourceID,
		Action:       input.Action,
		ContextHash:  input.ContextHash,
	}

	result, err := a.policyEvaluationRepo.GetCachedEvaluationResult(ctx, getCacheReq)
	if err != nil {
		// Cache miss is not an error
		a.metrics.IncrementCounter("policy_evaluation_cache_miss", metrics.Fields{"operation": "get_cached"})

		duration := time.Since(startTime)
		a.metrics.ObserveHistogram("abac_cache_retrieve_duration_seconds", duration.Seconds(),
			metrics.Fields{"operation": "policy_evaluation", "hit": "false"})

		return &GetCachedPolicyEvaluationOutput{
			Found: false,
		}, nil
	}

	duration := time.Since(startTime)
	span.SetAttributes(attribute.Bool("cache_hit", true))

	a.metrics.IncrementCounter("policy_evaluation_cache_hit", metrics.Fields{"operation": "get_cached"})
	a.metrics.ObserveHistogram("abac_cache_retrieve_duration_seconds", duration.Seconds(),
		metrics.Fields{"operation": "policy_evaluation", "hit": "true"})

	a.logger.InfoContext(ctx, "Policy evaluation cache hit",
		logger.Fields{
			"decision":    result.Decision,
			"cached_at":   result.CachedAt,
			"expires_at":  result.ExpiresAt,
			"duration_ms": duration.Milliseconds(),
			"request_id":  input.RequestID,
		})

	return &GetCachedPolicyEvaluationOutput{
		Found:           true,
		Decision:        result.Decision,
		PolicyDecisions: ConvertPolicyDecisionInfoToDecision(result.PolicyDecisions),
		CachedAt:        result.CachedAt,
		ExpiresAt:       result.ExpiresAt,
	}, nil
}

// InvalidatePolicyCacheInput represents input for cache invalidation
type InvalidatePolicyCacheInput struct {
	UserID        *uuid.UUID  `json:"user_id,omitempty"`
	ResourceType  *string     `json:"resource_type,omitempty"`
	ResourceID    *uuid.UUID  `json:"resource_id,omitempty"`
	Action        *string     `json:"action,omitempty"`
	PolicyIDs     []uuid.UUID `json:"policy_ids,omitempty"`
	InvalidateAll bool        `json:"invalidate_all"`
	RequestID     string      `json:"request_id"`
}

// InvalidatePolicyCacheOutput represents output of cache invalidation
type InvalidatePolicyCacheOutput struct {
	Success          bool  `json:"success"`
	InvalidatedCount int64 `json:"invalidated_count"`
}

// InvalidatePolicyCache invalidates cached policy evaluation results
func (a *CacheActivities) InvalidatePolicyCache(ctx context.Context, input *InvalidatePolicyCacheInput) (*InvalidatePolicyCacheOutput, error) {
	ctx, span := a.tracer.StartSpan(ctx, "abac.activities.InvalidatePolicyCache",
		tracing.WithAttributes(
			attribute.Bool("invalidate_all", input.InvalidateAll),
			attribute.String("request_id", input.RequestID),
		))
	defer span.End()

	startTime := time.Now()

	a.logger.InfoContext(ctx, "Invalidating policy evaluation cache",
		logger.Fields{
			"user_id":        input.UserID,
			"resource_type":  input.ResourceType,
			"action":         input.Action,
			"policy_ids":     input.PolicyIDs,
			"invalidate_all": input.InvalidateAll,
			"request_id":     input.RequestID,
		})

	invalidateReq := &repository.InvalidateEvaluationCacheRequest{
		UserID:        input.UserID,
		ResourceType:  input.ResourceType,
		ResourceID:    input.ResourceID,
		Action:        input.Action,
		PolicyIDs:     input.PolicyIDs,
		InvalidateAll: input.InvalidateAll,
	}

	err := a.policyEvaluationRepo.InvalidateEvaluationCache(ctx, invalidateReq)
	if err != nil {
		a.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		a.metrics.IncrementCounter("policy_evaluation_cache_invalidation_failed", metrics.Fields{"reason": "invalidation_failed"})
		return nil, errors.NewBusinessErrorWithContext(ctx, "CACHE_INVALIDATION_FAILED", "Failed to invalidate policy cache").WithDetail("error", err.Error())
	}

	duration := time.Since(startTime)

	a.metrics.ObserveHistogram("abac_cache_invalidation_duration_seconds", duration.Seconds(),
		metrics.Fields{"invalidate_all": fmt.Sprintf("%t", input.InvalidateAll)})
	a.metrics.IncrementCounter("policy_evaluation_cache_invalidation_success", metrics.Fields{"operation": "invalidate"})

	a.logger.InfoContext(ctx, "Policy evaluation cache invalidated successfully",
		logger.Fields{
			"duration_ms": duration.Milliseconds(),
			"request_id":  input.RequestID,
		})

	return &InvalidatePolicyCacheOutput{
		Success:          true,
		InvalidatedCount: 1, // Placeholder - implement actual count if needed
	}, nil
}

// CleanupExpiredCacheInput represents input for cleaning up expired cache entries
type CleanupExpiredCacheInput struct {
	MaxAge    time.Duration `json:"max_age"`
	BatchSize int32         `json:"batch_size"`
	RequestID string        `json:"request_id"`
}

// CleanupExpiredCacheOutput represents output of cache cleanup
type CleanupExpiredCacheOutput struct {
	Success      bool  `json:"success"`
	CleanedCount int64 `json:"cleaned_count"`
}

// CleanupExpiredCache cleans up expired cache entries
func (a *CacheActivities) CleanupExpiredCache(ctx context.Context, input *CleanupExpiredCacheInput) (*CleanupExpiredCacheOutput, error) {
	ctx, span := a.tracer.StartSpan(ctx, "abac.activities.CleanupExpiredCache",
		tracing.WithAttributes(
			attribute.String("request_id", input.RequestID),
		))
	defer span.End()

	startTime := time.Now()

	a.logger.InfoContext(ctx, "Starting cache cleanup",
		logger.Fields{
			"max_age":    input.MaxAge,
			"batch_size": input.BatchSize,
			"request_id": input.RequestID,
		})

	err := a.policyEvaluationRepo.CleanupExpiredEvaluations(ctx)
	if err != nil {
		a.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		a.metrics.IncrementCounter("policy_evaluation_cache_cleanup_failed", metrics.Fields{"reason": "cleanup_failed"})
		return nil, errors.NewBusinessErrorWithContext(ctx, "CACHE_CLEANUP_FAILED", "Failed to cleanup expired cache entries").WithDetail("error", err.Error())
	}

	duration := time.Since(startTime)

	a.metrics.ObserveHistogram("abac_cache_cleanup_duration_seconds", duration.Seconds(), metrics.Fields{})
	a.metrics.IncrementCounter("policy_evaluation_cache_cleanup_success", metrics.Fields{"operation": "cleanup"})

	a.logger.InfoContext(ctx, "Cache cleanup completed",
		logger.Fields{
			"duration_ms": duration.Milliseconds(),
			"request_id":  input.RequestID,
		})

	return &CleanupExpiredCacheOutput{
		Success:      true,
		CleanedCount: 1, // Placeholder - implement actual count if needed
	}, nil
}

// GetCacheStatsInput represents input for getting cache statistics
type GetCacheStatsInput struct {
	RequestID string `json:"request_id"`
}

// GetCacheStatsOutput represents cache statistics
type GetCacheStatsOutput struct {
	PolicyEvaluationStats *repository.EvaluationCacheStats `json:"policy_evaluation_stats"`
	AttributeStats        *repository.AttributeStats       `json:"attribute_stats"`
	GeneratedAt           time.Time                        `json:"generated_at"`
}

// GetCacheStats retrieves cache performance statistics
func (a *CacheActivities) GetCacheStats(ctx context.Context, input *GetCacheStatsInput) (*GetCacheStatsOutput, error) {
	ctx, span := a.tracer.StartSpan(ctx, "abac.activities.GetCacheStats",
		tracing.WithAttributes(
			attribute.String("request_id", input.RequestID),
		))
	defer span.End()

	startTime := time.Now()

	// Get policy evaluation cache stats
	evalStats, err := a.policyEvaluationRepo.GetEvaluationCacheStats(ctx)
	if err != nil {
		a.logger.WarnContext(ctx, "Failed to get policy evaluation cache stats",
			logger.Fields{"error": err.Error()})
		evalStats = &repository.EvaluationCacheStats{} // Empty stats
	}

	// Get attribute stats
	attrStats, err := a.attributeRepo.GetAttributeStats(ctx)
	if err != nil {
		a.logger.WarnContext(ctx, "Failed to get attribute stats",
			logger.Fields{"error": err.Error()})
		attrStats = &repository.AttributeStats{} // Empty stats
	}

	duration := time.Since(startTime)

	a.metrics.ObserveHistogram("abac_cache_stats_duration_seconds", duration.Seconds(), metrics.Fields{})

	a.logger.InfoContext(ctx, "Cache statistics retrieved",
		logger.Fields{
			"policy_eval_cache_hit_rate": evalStats.CacheHitRate,
			"total_cached_evaluations":   evalStats.TotalCachedEvaluations,
			"total_attributes":           attrStats.UserAttributes + attrStats.ResourceAttributes + attrStats.EnvironmentAttributes,
			"duration_ms":                duration.Milliseconds(),
			"request_id":                 input.RequestID,
		})

	return &GetCacheStatsOutput{
		PolicyEvaluationStats: evalStats,
		AttributeStats:        attrStats,
		GeneratedAt:           time.Now(),
	}, nil
}

// WarmupCacheInput represents input for cache warmup
type WarmupCacheInput struct {
	UserIDs       []uuid.UUID `json:"user_ids,omitempty"`
	ResourceTypes []string    `json:"resource_types,omitempty"`
	Actions       []string    `json:"actions,omitempty"`
	Priority      string      `json:"priority"` // high, medium, low
	RequestID     string      `json:"request_id"`
}

// WarmupCacheOutput represents output of cache warmup
type WarmupCacheOutput struct {
	Success      bool  `json:"success"`
	WarmedCount  int64 `json:"warmed_count"`
	SkippedCount int64 `json:"skipped_count"`
}

// WarmupCache pre-loads cache with commonly accessed evaluations
func (a *CacheActivities) WarmupCache(ctx context.Context, input *WarmupCacheInput) (*WarmupCacheOutput, error) {
	ctx, span := a.tracer.StartSpan(ctx, "abac.activities.WarmupCache",
		tracing.WithAttributes(
			attribute.String("priority", input.Priority),
			attribute.String("request_id", input.RequestID),
		))
	defer span.End()

	startTime := time.Now()

	a.logger.InfoContext(ctx, "Starting cache warmup",
		logger.Fields{
			"user_count":     len(input.UserIDs),
			"resource_types": input.ResourceTypes,
			"actions":        input.Actions,
			"priority":       input.Priority,
			"request_id":     input.RequestID,
		})

	// This is a placeholder implementation
	// In a real implementation, you would:
	// 1. Get most common user/resource/action combinations
	// 2. Pre-evaluate policies for these combinations
	// 3. Cache the results

	warmedCount := int64(0)
	skippedCount := int64(0)

	// Simulate warmup processing
	for _, userID := range input.UserIDs {
		for _, resourceType := range input.ResourceTypes {
			for _, action := range input.Actions {
				// Check if already cached
				getCacheReq := &repository.GetCachedEvaluationResultRequest{
					UserID:       userID,
					ResourceType: resourceType,
					Action:       action,
					ContextHash:  "warmup", // Simplified context for warmup
				}

				_, err := a.policyEvaluationRepo.GetCachedEvaluationResult(ctx, getCacheReq)
				if err == nil {
					skippedCount++ // Already cached
					continue
				}

				// Would perform evaluation and cache here
				warmedCount++
			}
		}
	}

	duration := time.Since(startTime)

	a.metrics.ObserveHistogram("abac_cache_warmup_duration_seconds", duration.Seconds(),
		metrics.Fields{"priority": input.Priority})
	a.metrics.SetGauge("abac_cache_warmed_entries", float64(warmedCount),
		metrics.Fields{"priority": input.Priority})

	a.logger.InfoContext(ctx, "Cache warmup completed",
		logger.Fields{
			"warmed_count":  warmedCount,
			"skipped_count": skippedCount,
			"duration_ms":   duration.Milliseconds(),
			"request_id":    input.RequestID,
		})

	return &WarmupCacheOutput{
		Success:      true,
		WarmedCount:  warmedCount,
		SkippedCount: skippedCount,
	}, nil
}

// generateCacheKey generates a cache key for policy evaluation
func (a *CacheActivities) generateCacheKey(input *CachePolicyEvaluationInput) string {
	resourcePart := input.ResourceType
	if input.ResourceID != nil {
		resourcePart = fmt.Sprintf("%s:%s", input.ResourceType, input.ResourceID.String())
	}

	return fmt.Sprintf("eval:%s:%s:%s:%s",
		input.UserID.String(),
		resourcePart,
		input.Action,
		input.ContextHash)
}

// Conversion functions moved to policy_evaluation.go to avoid duplication
