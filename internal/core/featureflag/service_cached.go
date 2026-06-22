package featureflag

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"awo.so/internal/platform/cache"
	"awo.so/internal/shared"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
)

// CachedFeatureFlagService wraps the simple service with Redis caching
type CachedFeatureFlagService struct {
	service Service
	cache   cache.Service
	logger  logger.Logger
	metrics *metrics.MetricsService
	tracing tracing.Service
}

// NewCachedFeatureFlagService creates a new cached feature flag service
func NewCachedFeatureFlagService(
	service Service,
	cache cache.Service,
	logger logger.Logger,
	metrics *metrics.MetricsService,
	tracing tracing.Service,
) Service {
	return &CachedFeatureFlagService{
		service: service,
		cache:   cache,
		logger:  logger,
		metrics: metrics,
		tracing: tracing,
	}
}

// Cache key patterns
const (
	// Feature flag data cache keys
	FlagCacheKey        = "flag:%s:%s"          // tenant_id:flag_name
	FlagByCacheKey      = "flag_by_id:%s:%s"    // tenant_id:flag_id
	FlagListCacheKey    = "flag_list:%s:%s"     // tenant_id:hash_of_params
	FlagStatsCacheKey   = "flag_stats:%s"       // tenant_id
	FlagsByTypeCacheKey = "flags_by_type:%s:%s" // tenant_id:flag_type

	// Evaluation cache keys
	EvaluationCacheKey     = "eval:%s"      // hash_of_evaluation_request
	BulkEvaluationCacheKey = "bulk_eval:%s" // hash_of_bulk_evaluation_request

	// Cache TTL configurations
	FlagDataCacheTTL   = 15 * time.Minute // Flag data changes less frequently
	EvaluationCacheTTL = 5 * time.Minute  // Evaluations can be cached shorter
	StatsCacheTTL      = 30 * time.Second // Stats change frequently
	ListCacheTTL       = 2 * time.Minute  // Lists can be cached briefly
)

// Feature Flag Management (with caching)

func (s *CachedFeatureFlagService) CreateFeatureFlag(ctx context.Context, request *CreateFeatureFlagRequest) (*FeatureFlag, error) {
	ctx, span := s.tracing.StartSpan(ctx, "featureflag.cached.create")
	defer span.End()

	// Create the flag through the underlying service
	flag, err := s.service.CreateFeatureFlag(ctx, request)
	if err != nil {
		return nil, err
	}

	// Invalidate related caches
	s.invalidateFlagCaches(ctx, flag.TenantID, flag.Name)

	s.logger.Info("Feature flag created and caches invalidated", logger.Fields{
		"flag_id":   flag.ID.String(),
		"flag_name": flag.Name,
		"tenant_id": flag.TenantID.String(),
	})

	return flag, nil
}

func (s *CachedFeatureFlagService) GetFeatureFlag(ctx context.Context, name string) (*FeatureFlag, error) {
	ctx, span := s.tracing.StartSpan(ctx, "featureflag.cached.get")
	defer span.End()

	tenantID, err := s.getTenantIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	cacheKey := fmt.Sprintf(FlagCacheKey, tenantID.String(), name)

	// Try to get from cache first
	var cachedFlag FeatureFlag
	err = s.cache.Get(ctx, cacheKey, &cachedFlag)
	if err == nil {
		s.recordCacheHit("feature_flag_get")
		s.logger.Debug("Feature flag cache hit", logger.Fields{
			"flag_name": name,
			"cache_key": cacheKey,
		})
		return &cachedFlag, nil
	}

	s.recordCacheMiss("feature_flag_get")

	// Cache miss - get from underlying service
	flag, err := s.service.GetFeatureFlag(ctx, name)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if err := s.cache.Set(ctx, cacheKey, flag, FlagDataCacheTTL); err != nil {
		s.logger.Warn("Failed to cache feature flag", logger.Fields{
			"error":     err.Error(),
			"flag_name": name,
			"cache_key": cacheKey,
		})
	}

	return flag, nil
}

func (s *CachedFeatureFlagService) GetFeatureFlagByID(ctx context.Context, id uuid.UUID) (*FeatureFlag, error) {
	ctx, span := s.tracing.StartSpan(ctx, "featureflag.cached.get_by_id")
	defer span.End()

	tenantID, err := s.getTenantIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	cacheKey := fmt.Sprintf(FlagByCacheKey, tenantID.String(), id.String())

	// Try to get from cache first
	var cachedFlag FeatureFlag
	err = s.cache.Get(ctx, cacheKey, &cachedFlag)
	if err == nil {
		s.recordCacheHit("feature_flag_get_by_id")
		return &cachedFlag, nil
	}

	s.recordCacheMiss("feature_flag_get_by_id")

	// Cache miss - get from underlying service
	flag, err := s.service.GetFeatureFlagByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cache the result with both ID and name keys
	if err := s.cache.Set(ctx, cacheKey, flag, FlagDataCacheTTL); err != nil {
		s.logger.Warn("Failed to cache feature flag by ID", logger.Fields{
			"error":   err.Error(),
			"flag_id": id.String(),
		})
	}

	// Also cache by name for consistency
	nameCacheKey := fmt.Sprintf(FlagCacheKey, tenantID.String(), flag.Name)
	if err := s.cache.Set(ctx, nameCacheKey, flag, FlagDataCacheTTL); err != nil {
		s.logger.Warn("Failed to cache feature flag by name", logger.Fields{
			"error":     err.Error(),
			"flag_name": flag.Name,
		})
	}

	return flag, nil
}

func (s *CachedFeatureFlagService) UpdateFeatureFlag(ctx context.Context, id uuid.UUID, request *UpdateFeatureFlagRequest) (*FeatureFlag, error) {
	ctx, span := s.tracing.StartSpan(ctx, "featureflag.cached.update")
	defer span.End()

	// Update through underlying service
	flag, err := s.service.UpdateFeatureFlag(ctx, id, request)
	if err != nil {
		return nil, err
	}

	// Invalidate caches
	s.invalidateFlagCaches(ctx, flag.TenantID, flag.Name)

	s.logger.Info("Feature flag updated and caches invalidated", logger.Fields{
		"flag_id":   flag.ID.String(),
		"flag_name": flag.Name,
	})

	return flag, nil
}

func (s *CachedFeatureFlagService) DeleteFeatureFlag(ctx context.Context, id uuid.UUID) error {
	ctx, span := s.tracing.StartSpan(ctx, "featureflag.cached.delete")
	defer span.End()

	// Get the flag first to know which caches to invalidate
	flag, err := s.service.GetFeatureFlagByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete through underlying service
	err = s.service.DeleteFeatureFlag(ctx, id)
	if err != nil {
		return err
	}

	// Invalidate caches
	s.invalidateFlagCaches(ctx, flag.TenantID, flag.Name)

	return nil
}

func (s *CachedFeatureFlagService) ListFeatureFlags(ctx context.Context, request *ListFeatureFlagsRequest) (*ListFeatureFlagsResponse, error) {
	ctx, span := s.tracing.StartSpan(ctx, "featureflag.cached.list")
	defer span.End()

	tenantID, err := s.getTenantIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Create cache key based on request parameters
	cacheKey := s.buildListCacheKey(tenantID, request)

	// Try to get from cache first
	var cachedResponse ListFeatureFlagsResponse
	err = s.cache.Get(ctx, cacheKey, &cachedResponse)
	if err == nil {
		s.recordCacheHit("feature_flag_list")
		return &cachedResponse, nil
	}

	s.recordCacheMiss("feature_flag_list")

	// Cache miss - get from underlying service
	response, err := s.service.ListFeatureFlags(ctx, request)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if err := s.cache.Set(ctx, cacheKey, response, ListCacheTTL); err != nil {
		s.logger.Warn("Failed to cache feature flag list", logger.Fields{
			"error": err.Error(),
		})
	}

	return response, nil
}

// Flag Evaluation (with caching)

func (s *CachedFeatureFlagService) EvaluateFlag(ctx context.Context, name string, evalCtx *EvaluationContext) (*EvaluationResult, error) {
	ctx, span := s.tracing.StartSpan(ctx, "featureflag.cached.evaluate")
	defer span.End()

	// Create cache key based on evaluation parameters
	cacheKey := s.buildEvaluationCacheKey(name, evalCtx)

	// Try to get from cache first
	var cachedResult EvaluationResult
	err := s.cache.Get(ctx, cacheKey, &cachedResult)
	if err == nil {
		s.recordCacheHit("feature_flag_evaluation")
		// Update cache hit in metadata
		cachedResult.Metadata.CacheHit = true
		return &cachedResult, nil
	}

	s.recordCacheMiss("feature_flag_evaluation")

	// Cache miss - evaluate through underlying service
	result, err := s.service.EvaluateFlag(ctx, name, evalCtx)
	if err != nil {
		return nil, err
	}

	// Mark as not from cache
	result.Metadata.CacheHit = false

	// Cache the result
	if err := s.cache.Set(ctx, cacheKey, result, EvaluationCacheTTL); err != nil {
		s.logger.Warn("Failed to cache evaluation result", logger.Fields{
			"error":     err.Error(),
			"flag_name": name,
		})
	}

	return result, nil
}

func (s *CachedFeatureFlagService) EvaluateFlags(ctx context.Context, names []string, evalCtx *EvaluationContext) (*BulkEvaluationResponse, error) {
	ctx, span := s.tracing.StartSpan(ctx, "featureflag.cached.evaluate_bulk")
	defer span.End()

	// Create cache key for bulk evaluation
	cacheKey := s.buildBulkEvaluationCacheKey(names, evalCtx)

	// Try to get from cache first
	var cachedResponse BulkEvaluationResponse
	err := s.cache.Get(ctx, cacheKey, &cachedResponse)
	if err == nil {
		s.recordCacheHit("feature_flag_bulk_evaluation")
		// Mark all results as cache hits
		for _, result := range cachedResponse.Results {
			result.Metadata.CacheHit = true
		}
		return &cachedResponse, nil
	}

	s.recordCacheMiss("feature_flag_bulk_evaluation")

	// Cache miss - evaluate through underlying service
	response, err := s.service.EvaluateFlags(ctx, names, evalCtx)
	if err != nil {
		return nil, err
	}

	// Mark all results as not from cache
	for _, result := range response.Results {
		result.Metadata.CacheHit = false
	}

	// Cache the result
	if err := s.cache.Set(ctx, cacheKey, response, EvaluationCacheTTL); err != nil {
		s.logger.Warn("Failed to cache bulk evaluation result", logger.Fields{
			"error":      err.Error(),
			"flag_count": len(names),
		})
	}

	return response, nil
}

func (s *CachedFeatureFlagService) IsEnabled(ctx context.Context, name string, evalCtx *EvaluationContext) (bool, error) {
	result, err := s.EvaluateFlag(ctx, name, evalCtx)
	if err != nil {
		return false, err
	}
	return result.Enabled, nil
}

// Analytics and Monitoring (with caching)

func (s *CachedFeatureFlagService) GetFlagStats(ctx context.Context) (*FlagStats, error) {
	ctx, span := s.tracing.StartSpan(ctx, "featureflag.cached.get_stats")
	defer span.End()

	tenantID, err := s.getTenantIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	cacheKey := fmt.Sprintf(FlagStatsCacheKey, tenantID.String())

	// Try to get from cache first
	var cachedStats FlagStats
	err = s.cache.Get(ctx, cacheKey, &cachedStats)
	if err == nil {
		s.recordCacheHit("feature_flag_stats")
		return &cachedStats, nil
	}

	s.recordCacheMiss("feature_flag_stats")

	// Cache miss - get from underlying service
	stats, err := s.service.GetFlagStats(ctx)
	if err != nil {
		return nil, err
	}

	// Cache the result with shorter TTL since stats change frequently
	if err := s.cache.Set(ctx, cacheKey, stats, StatsCacheTTL); err != nil {
		s.logger.Warn("Failed to cache flag stats", logger.Fields{
			"error": err.Error(),
		})
	}

	return stats, nil
}

func (s *CachedFeatureFlagService) SearchFlags(ctx context.Context, query string, limit, offset int32) ([]*FeatureFlag, error) {
	// Search results are less predictable, so we don't cache them
	// But we could implement a simple cache based on query parameters if needed
	return s.service.SearchFlags(ctx, query, limit, offset)
}

func (s *CachedFeatureFlagService) GetFlagsByType(ctx context.Context, flagType string) ([]*FeatureFlag, error) {
	ctx, span := s.tracing.StartSpan(ctx, "featureflag.cached.get_by_type")
	defer span.End()

	tenantID, err := s.getTenantIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	cacheKey := fmt.Sprintf(FlagsByTypeCacheKey, tenantID.String(), flagType)

	// Try to get from cache first
	var cachedFlags []*FeatureFlag
	err = s.cache.Get(ctx, cacheKey, &cachedFlags)
	if err == nil {
		s.recordCacheHit("feature_flag_by_type")
		return cachedFlags, nil
	}

	s.recordCacheMiss("feature_flag_by_type")

	// Cache miss - get from underlying service
	flags, err := s.service.GetFlagsByType(ctx, flagType)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if err := s.cache.Set(ctx, cacheKey, flags, FlagDataCacheTTL); err != nil {
		s.logger.Warn("Failed to cache flags by type", logger.Fields{
			"error":     err.Error(),
			"flag_type": flagType,
		})
	}

	return flags, nil
}

// Cache management helper methods

func (s *CachedFeatureFlagService) invalidateFlagCaches(ctx context.Context, tenantID uuid.UUID, flagName string) {
	patterns := []string{
		fmt.Sprintf("flag:%s:*", tenantID.String()),          // All flag data
		fmt.Sprintf("flag_by_id:%s:*", tenantID.String()),    // Flag by ID
		fmt.Sprintf("flag_list:%s:*", tenantID.String()),     // Flag lists
		fmt.Sprintf("flag_stats:%s", tenantID.String()),      // Stats
		fmt.Sprintf("flags_by_type:%s:*", tenantID.String()), // Flags by type
		"eval:*",      // All evaluations (could be more specific)
		"bulk_eval:*", // All bulk evaluations
	}

	for _, pattern := range patterns {
		if err := s.cache.DeletePattern(ctx, pattern); err != nil {
			s.logger.Warn("Failed to invalidate cache pattern", logger.Fields{
				"error":   err.Error(),
				"pattern": pattern,
			})
		}
	}

	s.logger.Debug("Invalidated feature flag caches", logger.Fields{
		"tenant_id": tenantID.String(),
		"flag_name": flagName,
	})
}

func (s *CachedFeatureFlagService) buildListCacheKey(tenantID uuid.UUID, request *ListFeatureFlagsRequest) string {
	// Create a hash of the request parameters
	data, _ := json.Marshal(request)
	hash := sha256.Sum256(data)
	hashStr := hex.EncodeToString(hash[:])

	return fmt.Sprintf(FlagListCacheKey, tenantID.String(), hashStr)
}

func (s *CachedFeatureFlagService) buildEvaluationCacheKey(name string, evalCtx *EvaluationContext) string {
	// Create a hash of the evaluation parameters
	keyData := struct {
		Name    string             `json:"name"`
		Context *EvaluationContext `json:"context"`
	}{
		Name:    name,
		Context: evalCtx,
	}

	data, _ := json.Marshal(keyData)
	hash := sha256.Sum256(data)
	hashStr := hex.EncodeToString(hash[:])

	return fmt.Sprintf(EvaluationCacheKey, hashStr)
}

func (s *CachedFeatureFlagService) buildBulkEvaluationCacheKey(names []string, evalCtx *EvaluationContext) string {
	// Create a hash of the bulk evaluation parameters
	keyData := struct {
		Names   []string           `json:"names"`
		Context *EvaluationContext `json:"context"`
	}{
		Names:   names,
		Context: evalCtx,
	}

	data, _ := json.Marshal(keyData)
	hash := sha256.Sum256(data)
	hashStr := hex.EncodeToString(hash[:])

	return fmt.Sprintf(BulkEvaluationCacheKey, hashStr)
}

func (s *CachedFeatureFlagService) getTenantIDFromContext(ctx context.Context) (uuid.UUID, error) {
	if id, ok := shared.GetTenantID(ctx); ok {
		return id, nil
	}
	return uuid.Nil, fmt.Errorf("tenant ID not found in context")
}

func (s *CachedFeatureFlagService) recordCacheHit(operation string) {
	if s.metrics != nil {
		s.metrics.IncrementCounter("feature_flag_cache_hits", metrics.Fields{
			"operation": operation,
		})
	}
}

func (s *CachedFeatureFlagService) recordCacheMiss(operation string) {
	if s.metrics != nil {
		s.metrics.IncrementCounter("feature_flag_cache_misses", metrics.Fields{
			"operation": operation,
		})
	}
}
