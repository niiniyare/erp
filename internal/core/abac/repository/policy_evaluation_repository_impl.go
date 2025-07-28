package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/core/abac/models"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/niiniyare/erp/internal/shared/types"
)

// policyEvaluationRepository implements PolicyEvaluationRepository using SQLC and Store
type policyEvaluationRepository struct {
	store   sqlc.Store
	cache   cache.Service
	logger  logger.Logger
	metrics metrics.MetricsProvider
	tracer  tracing.TracingService
}

// NewPolicyEvaluationRepository creates a new policy evaluation repository implementation
func NewPolicyEvaluationRepository(
	store sqlc.Store,
	cache cache.Service,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.TracingService,
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
			tracing.StringAttribute("user_id", req.UserID.String()),
			tracing.StringAttribute("resource_type", req.ResourceType),
			tracing.StringAttribute("action", req.Action),
			tracing.StringAttribute("decision", string(req.Decision)),
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
		// Convert to a serializable format
		decisions := make([]map[string]interface{}, len(req.Result.PolicyDecisions))
		for i, decision := range req.Result.PolicyDecisions {
			decisions[i] = map[string]interface{}{
				"policy_id":   decision.PolicyID,
				"policy_name": decision.PolicyName,
				"effect":      string(decision.Effect),
				"reason":      decision.Reason,
			}
		}
		// In a real implementation, you would marshal this to JSON
		// For now, we'll use a simple approach
		policyDecisionsJSON = []byte("{}")
	}

	var cacheKey *string
	if req.ResourceID != nil {
		key := req.UserID.String() + ":" + req.ResourceType + ":" + req.ResourceID.String() + ":" + req.Action + ":" + req.ContextHash
		cacheKey = &key
	} else {
		key := req.UserID.String() + ":" + req.ResourceType + "::" + req.Action + ":" + req.ContextHash
		cacheKey = &key
	}

	params := sqlc.CacheEvaluationResultParams{
		UserID:             req.UserID,
		ResourceType:       req.ResourceType,
		ResourceID:         req.ResourceID,
		Action:             req.Action,
		EntityID:           req.EntityID,
		ContextHash:        req.ContextHash,
		Decision:           string(req.Decision),
		ApplicablePolicies: req.ApplicablePolicies,
		PolicyDecisions:    policyDecisionsJSON,
		EvaluationTimeMs:   sql.NullInt32{Int32: int32(req.EvaluationTimeMS), Valid: req.EvaluationTimeMS > 0},
		CacheKey:           cacheKey,
		ExpiresAt:          req.ExpiresAt,
	}

	// Extract tenant ID from context
	tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
	if !ok {
		return errors.NewBusinessError("TENANT_CONTEXT_REQUIRED", "Valid tenant context required")
	}

	// Use tenant-aware transaction
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, txStore sqlc.Store) error {
		return txStore.CacheEvaluationResult(ctx, params)
	})
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementErrorCount("policy_evaluation_repository", "cache_failed")
		return errors.NewBusinessErrorWithContext(ctx, "POLICY_EVALUATION_CACHE_FAILED", "Failed to cache policy evaluation result").WithErr(err)
	}

	r.metrics.IncrementSuccessCount("policy_evaluation_repository_cache")
	return nil
}

// GetCachedEvaluationResult retrieves a cached policy evaluation result
func (r *policyEvaluationRepository) GetCachedEvaluationResult(ctx context.Context, req *GetCachedEvaluationResultRequest) (*CachedEvaluationResult, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.GetCachedEvaluationResult",
		tracing.WithAttributes(
			tracing.StringAttribute("user_id", req.UserID.String()),
			tracing.StringAttribute("resource_type", req.ResourceType),
			tracing.StringAttribute("action", req.Action),
		))
	defer span.End()

	params := sqlc.GetCachedEvaluationResultParams{
		UserID:       req.UserID,
		ResourceType: req.ResourceType,
		Column3:      req.ResourceID,
		Action:       req.Action,
		ContextHash:  req.ContextHash,
	}

	// Extract tenant ID from context
	tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
	if !ok {
		return nil, errors.NewBusinessError("TENANT_CONTEXT_REQUIRED", "Valid tenant context required")
	}

	// Try cache first
	cacheKey := fmt.Sprintf("eval:%s:%s:%s:%s", req.UserID.String(), req.ResourceType, req.Action, req.ContextHash)
	var evaluation sqlc.GetCachedEvaluationResultRow
	if err := r.cache.Get(ctx, cacheKey, &evaluation); err == nil {
		r.metrics.IncrementCounter("policy_evaluation_repository_cache_hit")
		// Convert and return cached result
		var policyDecisions []*models.PolicyDecision
		result := &CachedEvaluationResult{
			Decision:        types.PolicyDecisionType(evaluation.Decision),
			PolicyDecisions: policyDecisions,
			CachedAt:        evaluation.EvaluatedAt,
			ExpiresAt:       evaluation.ExpiresAt,
		}
		return result, nil
	}

	// Cache miss - get from database
	evaluation, err := r.store.GetCachedEvaluationResult(ctx, params)
	if err != nil {
		if err == sql.ErrNoRows {
			r.metrics.IncrementCounter("policy_evaluation_repository_cache_miss")
			return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_EVALUATION_NOT_CACHED", "Policy evaluation result not found in cache")
		}
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementErrorCount("policy_evaluation_repository", "get_cached_failed")
		return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_EVALUATION_GET_CACHED_FAILED", "Failed to get cached policy evaluation result").WithErr(err)
	}

	r.metrics.IncrementCounter("policy_evaluation_repository_cache_hit")

	// Convert policy decisions from JSONB back to models
	var policyDecisions []*models.PolicyDecision
	// In a real implementation, you would unmarshal the JSON
	// For now, we'll return empty decisions
	policyDecisions = []*models.PolicyDecision{}

	result := &CachedEvaluationResult{
		Decision:        types.PolicyDecisionType(evaluation.Decision),
		PolicyDecisions: policyDecisions,
		CachedAt:        evaluation.EvaluatedAt,
		ExpiresAt:       evaluation.ExpiresAt,
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

	// Extract tenant ID from context
	tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
	if !ok {
		return errors.NewBusinessError("TENANT_CONTEXT_REQUIRED", "Valid tenant context required")
	}

	// Use tenant-aware transaction
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, txStore sqlc.Store) error {
		if req.InvalidateAll {
			return txStore.InvalidateAllEvaluations(ctx)
		} else if req.UserID != nil {
			return txStore.InvalidateUserEvaluations(ctx, *req.UserID)
		} else if req.ResourceType != nil {
			return txStore.InvalidateResourceEvaluations(ctx, sqlc.InvalidateResourceEvaluationsParams{
				ResourceType: *req.ResourceType,
				Column2:      req.ResourceID,
			})
		} else if req.Action != nil {
			return txStore.InvalidateActionEvaluations(ctx, *req.Action)
		} else if len(req.PolicyIDs) > 0 {
			return txStore.InvalidatePolicyEvaluations(ctx, req.PolicyIDs)
		}
		return nil
	})

	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementErrorCount("policy_evaluation_repository", "invalidate_failed")
		return errors.NewBusinessErrorWithContext(ctx, "POLICY_EVALUATION_INVALIDATE_FAILED", "Failed to invalidate policy evaluation cache").WithErr(err)
	}

	r.metrics.IncrementSuccessCount("policy_evaluation_repository_invalidate")
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
		r.metrics.IncrementErrorCount("policy_evaluation_repository", "cleanup_failed")
		return errors.NewBusinessErrorWithContext(ctx, "POLICY_EVALUATION_CLEANUP_FAILED", "Failed to cleanup expired policy evaluations").WithErr(err)
	}

	r.metrics.IncrementSuccessCount("policy_evaluation_repository_cleanup")
	return nil
}

// GetEvaluationCacheStats gets cache performance statistics
func (r *policyEvaluationRepository) GetEvaluationCacheStats(ctx context.Context) (*EvaluationCacheStats, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.GetEvaluationCacheStats")
	defer span.End()

	stats, err := r.store.GetEvaluationCacheStats(ctx)
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementErrorCount("policy_evaluation_repository", "get_stats_failed")
		return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_EVALUATION_STATS_FAILED", "Failed to get evaluation cache statistics").WithErr(err)
	}

	r.metrics.IncrementSuccessCount("policy_evaluation_repository_get_stats")

	result := &EvaluationCacheStats{
		TotalCachedEvaluations: int(stats.TotalCachedEvaluations),
		ActiveEvaluations:      int(stats.ActiveEvaluations.Int64),
		ExpiredEvaluations:     int(stats.ExpiredEvaluations.Int64),
		UniqueUsers:            int(stats.UniqueUsers.Int64),
		UniqueResourceTypes:    int(stats.UniqueResourceTypes.Int64),
		UniqueActions:          int(stats.UniqueActions.Int64),
		CacheHitRate:           float64(stats.CacheHitRate.Float64),
	}

	if stats.AvgEvaluationTimeMs.Valid {
		result.AvgEvaluationTimeMS = float64(stats.AvgEvaluationTimeMs.Float64)
	}
	if stats.MinEvaluationTimeMs.Valid {
		result.MinEvaluationTimeMS = int(stats.MinEvaluationTimeMs.Int32)
	}
	if stats.MaxEvaluationTimeMs.Valid {
		result.MaxEvaluationTimeMS = int(stats.MaxEvaluationTimeMs.Int32)
	}

	return result, nil
}

// GetUserEvaluationHistory gets evaluation history for a user
func (r *policyEvaluationRepository) GetUserEvaluationHistory(ctx context.Context, req *GetUserEvaluationHistoryRequest) ([]*models.PolicyEvaluation, error) {
	ctx, span := r.tracer.StartSpan(ctx, "abac.repository.GetUserEvaluationHistory",
		tracing.WithAttributes(
			tracing.StringAttribute("user_id", req.UserID.String()),
		))
	defer span.End()

	params := sqlc.GetUserEvaluationHistoryParams{
		UserID:  req.UserID,
		Column2: req.ResourceType,
		Column3: req.Action,
		Column4: int32(req.Limit),
		Column5: int32(req.Offset),
	}

	evaluations, err := r.store.GetUserEvaluationHistory(ctx, params)
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementErrorCount("policy_evaluation_repository", "get_user_history_failed")
		return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_EVALUATION_HISTORY_FAILED", "Failed to get user evaluation history").WithErr(err)
	}

	r.metrics.IncrementSuccessCount("policy_evaluation_repository_get_user_history")

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

	params := sqlc.GetEvaluationMetricsParams{
		Column1: req.StartTime,
		Column2: req.EndTime,
	}

	metrics, err := r.store.GetEvaluationMetrics(ctx, params)
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementErrorCount("policy_evaluation_repository", "get_metrics_failed")
		return nil, errors.NewBusinessErrorWithContext(ctx, "POLICY_EVALUATION_METRICS_FAILED", "Failed to get evaluation metrics").WithErr(err)
	}

	r.metrics.IncrementSuccessCount("policy_evaluation_repository_get_metrics")

	result := &EvaluationMetrics{
		TotalEvaluations: int(metrics.TotalEvaluations.Int64),
		UniqueUsers:      int(metrics.UniqueUsers.Int64),
		UniqueResources:  int(metrics.UniqueResources.Int64),
	}

	if metrics.AvgEvaluationTime.Valid {
		result.AvgEvaluationTimeMS = float64(metrics.AvgEvaluationTime.Float64)
	}
	if metrics.MedianEvaluationTime.Valid {
		result.MedianEvaluationTimeMS = float64(metrics.MedianEvaluationTime.Float64)
	}
	if metrics.P95EvaluationTime.Valid {
		result.P95EvaluationTimeMS = float64(metrics.P95EvaluationTime.Float64)
	}
	if metrics.P99EvaluationTime.Valid {
		result.P99EvaluationTimeMS = float64(metrics.P99EvaluationTime.Float64)
	}

	return result, nil
}

// Helper method to convert SQLC PolicyEvaluation to domain model
func (r *policyEvaluationRepository) convertSQLCPolicyEvaluationToModel(eval sqlc.PolicyEvaluation) *models.PolicyEvaluation {
	var resourceID *uuid.UUID
	if eval.ResourceID.Valid {
		resourceID = &eval.ResourceID.UUID
	}

	var entityID *uuid.UUID
	if eval.EntityID.Valid {
		entityID = &eval.EntityID.UUID
	}

	var evaluationTimeMS *int64
	if eval.EvaluationTimeMs.Valid {
		time := int64(eval.EvaluationTimeMs.Int32)
		evaluationTimeMS = &time
	}

	// Convert policy decisions from JSONB
	var policyDecisions []*models.PolicyDecision
	// In a real implementation, you would unmarshal the JSON
	// For now, we'll return empty decisions
	policyDecisions = []*models.PolicyDecision{}

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
		EvaluatedAt:        eval.EvaluatedAt,
		ExpiresAt:          eval.ExpiresAt,
	}
}
