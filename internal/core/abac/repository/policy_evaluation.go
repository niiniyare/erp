package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"

	db "awo/db/sqlc"
	"awo/internal/core/abac/models"
	"awo/internal/platform/cache"
	"awo/internal/shared/convert"
	"awo/internal/shared/errors"
	"awo/internal/shared/logger"
	"awo/internal/shared/metrics"
	"awo/internal/shared/tracing"
	"awo/internal/shared/types"
)

// policyEvaluationRepository implements PolicyEvaluationRepository using SQLC and Store
type policyEvaluationRepository struct {
	store   db.Store
	cache   cache.Service
	logger  logger.Logger
	metrics metrics.MetricsProvider
	tracer  tracing.Service
}

// NewPolicyEvaluationRepository creates a new policy evaluation repository implementation
func NewPolicyEvaluationRepository(
	store db.Store,
	cache cache.Service,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.Service,
) PolicyEvaluationRepository {
	return &policyEvaluationRepository{
		store:   store,
		cache:   cache,
		logger:  logger,
		metrics: metrics,
		tracer:  tracer,
	}
}

// CacheEvaluationResult caches a policy evaluation result
func (r *policyEvaluationRepository) CacheEvaluationResult(ctx context.Context, req *CacheEvaluationResultRequest) error {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.CacheEvaluationResult",
		tracing.WithAttributes(
			attribute.String("user_id", req.UserID.String()),
			attribute.String("resource_type", req.ResourceType),
			attribute.String("action", req.Action),
			attribute.String("decision", string(req.Decision)),
		))
	defer span.End()

	r.logger.DebugContext(ctx, "Caching policy evaluation result",
		logger.Fields{
			"user_id":       req.UserID,
			"resource_type": req.ResourceType,
			"action":        req.Action,
			"decision":      req.Decision,
			"expires_at":    req.ExpiresAt,
		})

	// Convert models.PolicyDecision to JSONB
	var policyDecisionsJSON []byte
	if req.Result != nil && len(req.Result.PolicyDecisions) > 0 {
		policyDecisionsJSON, _ = json.Marshal(req.Result.PolicyDecisions)
	}

	var cacheKey *string
	if req.ResourceID != nil {
		key := req.UserID.String() + ":" + req.ResourceType + ":" + req.ResourceID.String() + ":" + req.Action + ":" + req.ContextHash
		cacheKey = &key
	} else {
		key := req.UserID.String() + ":" + req.ResourceType + "::" + req.Action + ":" + req.ContextHash
		cacheKey = &key
	}

	evaluationTimeMs, err := convert.Int64ToInt32(req.EvaluationTimeMS)
	if err != nil {
		return fmt.Errorf("invalid evaluation time: %w", err)
	}
	params := db.CacheEvaluationResultParams{
		UserID:             req.UserID,
		ResourceType:       req.ResourceType,
		ResourceID:         req.ResourceID,
		Action:             req.Action,
		ContextHash:        req.ContextHash,
		Decision:           string(req.Decision),
		ApplicablePolicies: req.ApplicablePolicies,
		PolicyDecisions:    policyDecisionsJSON,
		EvaluationTimeMs:   &evaluationTimeMs,
		CacheKey:           cacheKey,
		ExpiresAt:          sql.NullTime{Time: req.ExpiresAt, Valid: !req.ExpiresAt.IsZero()},
	}

	err = r.store.CacheEvaluationResult(ctx, params)
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.Counter("policy_evaluation_repository_cache_failed", "Total failed policy evaluation cache operations").Inc(nil)
		return errors.NewBusinessErrorWithContext(ctx, "POLICY_EVALUATION_CACHE_FAILED", "Failed to cache policy evaluation result").WithDetail("error", err.Error())
	}

	r.metrics.Counter("policy_evaluation_repository_cache_success", "Total successful policy evaluation cache operations").Inc(nil)
	return nil
}

// GetCachedEvaluationResult retrieves a cached policy evaluation result
func (r *policyEvaluationRepository) GetCachedEvaluationResult(ctx context.Context, req *GetCachedEvaluationResultRequest) (*models.PolicyEvaluationResult, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.GetCachedEvaluationResult",
		tracing.WithAttributes(
			attribute.String("user_id", req.UserID.String()),
			attribute.String("resource_type", req.ResourceType),
			attribute.String("action", req.Action),
		))
	defer span.End()

	params := db.GetCachedEvaluationResultParams{
		UserID:       req.UserID,
		ResourceType: req.ResourceType,
		Column3:      *req.ResourceID,
		Action:       req.Action,
		ContextHash:  req.ContextHash,
	}

	// Try cache first
	cacheKey := fmt.Sprintf("eval:%s:%s:%s:%s", req.UserID.String(), req.ResourceType, req.Action, req.ContextHash)
	var evaluation *db.PolicyEvaluation
	if err := r.cache.Get(ctx, cacheKey, &evaluation); err == nil {
		r.metrics.Counter("policy_evaluation_repository_cache_hit", "Total policy evaluation cache hits").Inc(nil)
		// Convert and return cached result
		var policyDecisions []models.PolicyDecisionInfo
		if err := json.Unmarshal(evaluation.PolicyDecisions, &policyDecisions); err != nil {
			return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_EVALUATION_UNMARSHAL_FAILED", "Failed to unmarshal policy decisions").WithDetail("error", err.Error())
		}
		result := &models.PolicyEvaluationResult{
			Decision:        types.PolicyDecisionType(evaluation.Decision),
			PolicyDecisions: policyDecisions,
			EvaluatedAt:     evaluation.EvaluatedAt.Time,
			CacheHit:        true,
		}
		return result, nil
	}

	// Cache miss - get from database
	evaluation, err := r.store.GetCachedEvaluationResult(ctx, params)
	if err != nil {
		if err == sql.ErrNoRows {
			r.metrics.Counter("policy_evaluation_repository_cache_miss", "Total policy evaluation cache misses").Inc(nil)
			return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_EVALUATION_NOT_CACHED", "Policy evaluation result not found in cache")
		}
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.Counter("policy_evaluation_repository_get_cached_failed", "Total failed get cached policy evaluation operations").Inc(nil)
		return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_EVALUATION_GET_CACHED_FAILED", "Failed to get cached policy evaluation result").WithDetail("error", err.Error())
	}

	r.metrics.Counter("policy_evaluation_repository_db_hit", "Total policy evaluation db hits").Inc(nil)

	// Convert policy decisions from JSONB back to models
	var policyDecisions []models.PolicyDecisionInfo
	if err := json.Unmarshal(evaluation.PolicyDecisions, &policyDecisions); err != nil {
		return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_EVALUATION_UNMARSHAL_FAILED", "Failed to unmarshal policy decisions").WithDetail("error", err.Error())
	}

	result := &models.PolicyEvaluationResult{
		Decision:        types.PolicyDecisionType(evaluation.Decision),
		PolicyDecisions: policyDecisions,
		EvaluatedAt:     evaluation.EvaluatedAt.Time,
		CacheHit:        false,
	}

	return result, nil
}

// InvalidateEvaluationCache invalidates cached evaluation results
func (r *policyEvaluationRepository) InvalidateEvaluationCache(ctx context.Context, req *InvalidateEvaluationCacheRequest) error {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.InvalidateEvaluationCache")
	defer span.End()

	r.logger.InfoContext(ctx, "Invalidating policy evaluation cache",
		logger.Fields{
			"user_id":        req.UserID,
			"resource_type":  req.ResourceType,
			"action":         req.Action,
			"policy_ids":     req.PolicyIDs,
			"invalidate_all": req.InvalidateAll,
		})

	if req.InvalidateAll {
		err := r.store.InvalidateAllEvaluations(ctx)
		if err != nil {
			r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
			r.metrics.Counter("policy_evaluation_repository_invalidate_failed", "Total failed policy evaluation invalidate operations").Inc(nil)
			return errors.NewBusinessErrorWithContext(ctx, "POLICY_EVALUATION_INVALIDATE_FAILED", "Failed to invalidate all policy evaluation cache").WithDetail("error", err.Error())
		}
	} else if req.UserID != nil {
		err := r.store.InvalidateUserEvaluations(ctx, *req.UserID)
		if err != nil {
			r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
			r.metrics.Counter("policy_evaluation_repository_invalidate_failed", "Total failed policy evaluation invalidate operations").Inc(nil)
			return errors.NewBusinessErrorWithContext(ctx, "POLICY_EVALUATION_INVALIDATE_FAILED", "Failed to invalidate user policy evaluation cache").WithDetail("error", err.Error())
		}
	} else if req.ResourceType != nil {
		err := r.store.InvalidateResourceEvaluations(ctx, db.InvalidateResourceEvaluationsParams{
			ResourceType: *req.ResourceType,
			Column2:      *req.ResourceID,
		})
		if err != nil {
			r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
			r.metrics.Counter("policy_evaluation_repository_invalidate_failed", "Total failed policy evaluation invalidate operations").Inc(nil)
			return errors.NewBusinessErrorWithContext(ctx, "POLICY_EVALUATION_INVALIDATE_FAILED", "Failed to invalidate resource policy evaluation cache").WithDetail("error", err.Error())
		}
	} else if req.Action != nil {
		err := r.store.InvalidateActionEvaluations(ctx, *req.Action)
		if err != nil {
			r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
			r.metrics.Counter("policy_evaluation_repository_invalidate_failed", "Total failed policy evaluation invalidate operations").Inc(nil)
			return errors.NewBusinessErrorWithContext(ctx, "POLICY_EVALUATION_INVALIDATE_FAILED", "Failed to invalidate action policy evaluation cache").WithDetail("error", err.Error())
		}
	} else if len(req.PolicyIDs) > 0 {
		err := r.store.InvalidatePolicyEvaluations(ctx, req.PolicyIDs)
		if err != nil {
			r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
			r.metrics.Counter("policy_evaluation_repository_invalidate_failed", "Total failed policy evaluation invalidate operations").Inc(nil)
			return errors.NewBusinessErrorWithContext(ctx, "POLICY_EVALUATION_INVALIDATE_FAILED", "Failed to invalidate policy evaluation cache").WithDetail("error", err.Error())
		}
	}

	r.metrics.Counter("policy_evaluation_repository_invalidate_success", "Total successful policy evaluation invalidate operations").Inc(nil)
	return nil
}

// CleanupExpiredEvaluations removes expired evaluation results
func (r *policyEvaluationRepository) CleanupExpiredEvaluations(ctx context.Context) error {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.CleanupExpiredEvaluations")
	defer span.End()

	r.logger.InfoContext(ctx, "Cleaning up expired policy evaluations")

	err := r.store.CleanupExpiredEvaluations(ctx)
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.Counter("policy_evaluation_repository_cleanup_failed", "Total failed policy evaluation cleanup operations").Inc(nil)
		return errors.NewBusinessErrorWithContext(ctx, "POLICY_EVALUATION_CLEANUP_FAILED", "Failed to cleanup expired policy evaluations").WithDetail("error", err.Error())
	}

	r.metrics.Counter("policy_evaluation_repository_cleanup_success", "Total successful policy evaluation cleanup operations").Inc(nil)
	return nil
}

// GetEvaluationCacheStats gets cache performance statistics
func (r *policyEvaluationRepository) GetEvaluationCacheStats(ctx context.Context) (*EvaluationCacheStats, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.GetEvaluationCacheStats")
	defer span.End()

	stats, err := r.store.GetEvaluationCacheStats(ctx)
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.Counter("policy_evaluation_repository_get_stats_failed", "Total failed get evaluation cache stats operations").Inc(nil)
		return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_EVALUATION_STATS_FAILED", "Failed to get evaluation cache statistics").WithDetail("error", err.Error())
	}

	r.metrics.Counter("policy_evaluation_repository_get_stats_success", "Total successful get evaluation cache stats operations").Inc(nil)

	result := &EvaluationCacheStats{
		TotalCachedEvaluations: stats.TotalCachedEvaluations,
		CacheHitRate:           0, // stats.CacheHitRate,
		CacheMissRate:          0, // This needs to be calculated
		ExpiredEvaluations:     stats.ExpiredEvaluations,
		AverageEvaluationTime:  time.Duration(stats.AvgEvaluationTimeMs) * time.Millisecond,
		EvaluationsByDecision:  nil, // This needs to be calculated
		EvaluationsByResource:  nil, // This needs to be calculated
	}

	return result, nil
}

// GetUserEvaluationHistory gets evaluation history for a user
func (r *policyEvaluationRepository) GetUserEvaluationHistory(ctx context.Context, req *GetUserEvaluationHistoryRequest) ([]*models.PolicyEvaluation, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.GetUserEvaluationHistory",
		tracing.WithAttributes(
			attribute.String("user_id", req.UserID.String()),
		))
	defer span.End()

	limit, err := convert.IntToInt32(req.Limit)
	if err != nil {
		return nil, fmt.Errorf("invalid limit: %w", err)
	}

	offset, err := convert.IntToInt32(req.Offset)
	if err != nil {
		return nil, fmt.Errorf("invalid offset: %w", err)
	}

	params := db.GetUserEvaluationHistoryParams{
		UserID:  req.UserID,
		Column2: *req.ResourceType,
		Column3: *req.Action,
		Limit:   limit,
		Offset:  offset,
	}

	evaluations, err := r.store.GetUserEvaluationHistory(ctx, params)
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.Counter("policy_evaluation_repository_get_user_history_failed", "Total failed get user evaluation history operations").Inc(nil)
		return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_EVALUATION_HISTORY_FAILED", "Failed to get user evaluation history").WithDetail("error", err.Error())
	}

	r.metrics.Counter("policy_evaluation_repository_get_user_history_success", "Total successful get user evaluation history operations").Inc(nil)

	result := make([]*models.PolicyEvaluation, len(evaluations))
	for i, evaluation := range evaluations {
		result[i] = r.convertSQLCPolicyEvaluationToModel(evaluation)
	}

	return result, nil
}

// GetEvaluationMetrics gets evaluation performance metrics
func (r *policyEvaluationRepository) GetEvaluationMetrics(ctx context.Context, req *GetEvaluationMetricsRequest) (*EvaluationMetrics, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.GetEvaluationMetrics")
	defer span.End()

	params := db.GetEvaluationMetricsParams{
		EvaluatedAt:   sql.NullTime{Time: req.StartTime, Valid: !req.StartTime.IsZero()},
		EvaluatedAt_2: sql.NullTime{Time: req.EndTime, Valid: !req.EndTime.IsZero()},
	}

	metrics, err := r.store.GetEvaluationMetrics(ctx, params)
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.Counter("policy_evaluation_repository_get_metrics_failed", "Total failed get evaluation metrics operations").Inc(nil)
		return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_EVALUATION_METRICS_FAILED", "Failed to get evaluation metrics").WithDetail("error", err.Error())
	}

	r.metrics.Counter("policy_evaluation_repository_get_metrics_success", "Total successful get evaluation metrics operations").Inc(nil)

	result := &EvaluationMetrics{
		TotalEvaluations:       int(metrics.TotalEvaluations),
		UniqueUsers:            int(metrics.UniqueUsers),
		UniqueResources:        int(metrics.UniqueResources),
		AvgEvaluationTimeMS:    metrics.AvgEvaluationTime,
		MedianEvaluationTimeMS: metrics.MedianEvaluationTime,
		P95EvaluationTimeMS:    metrics.P95EvaluationTime,
		P99EvaluationTimeMS:    metrics.P99EvaluationTime,
	}

	return result, nil
}

// Helper method to convert SQLC PolicyEvaluation to domain model
func (r *policyEvaluationRepository) convertSQLCPolicyEvaluationToModel(eval *db.PolicyEvaluation) *models.PolicyEvaluation {
	var resourceID *uuid.UUID
	if eval.ResourceID != nil {
		resourceID = eval.ResourceID
	}

	var entityID *uuid.UUID
	if eval.EntityID != nil {
		entityID = eval.EntityID
	}

	var evaluationTimeMS *int64
	if eval.EvaluationTimeMs != nil {
		time := int64(*eval.EvaluationTimeMs)
		evaluationTimeMS = &time
	}

	// Convert policy decisions from JSONB
	var policyDecisions []*models.PolicyDecision
	if err := json.Unmarshal(eval.PolicyDecisions, &policyDecisions); err != nil {
		// In a real implementation, you would handle this error properly
	}

	return &models.PolicyEvaluation{
		ID:                 eval.ID,
		TenantID:           eval.TenantID,
		UserID:             eval.UserID,
		ResourceType:       eval.ResourceType,
		ResourceID:         resourceID,
		Action:             eval.Action,
		EntityID:           entityID,
		ContextHash:        eval.ContextHash,
		Decision:           types.PolicyDecisionType(eval.Decision),
		ApplicablePolicies: eval.ApplicablePolicies,
		PolicyDecisions:    policyDecisions,
		EvaluationTimeMS:   evaluationTimeMS,
		EvaluatedAt:        eval.EvaluatedAt.Time,
		ExpiresAt:          eval.ExpiresAt.Time,
	}
}

func (r *policyEvaluationRepository) GetCachedEvaluationResults(ctx context.Context, requests []*GetCachedEvaluationResultRequest) ([]*models.PolicyEvaluationResult, error) {
	return nil, nil
}

func (r *policyEvaluationRepository) InvalidateEvaluationCacheByPolicyID(ctx context.Context, policyID uuid.UUID) error {
	return nil
}

func (r *policyEvaluationRepository) InvalidateEvaluationCacheByUserID(ctx context.Context, userID uuid.UUID) error {
	return nil
}
