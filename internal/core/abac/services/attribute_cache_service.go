package services

//go:generate go run go.uber.org/mock/mockgen -source=attribute_cache_service.go -destination=mock.go -package=services

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/niiniyare/erp/internal/core/abac/models"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/niiniyare/erp/internal/shared/types"
)

// AttributeCacheService defines the interface for attribute caching
type AttributeCacheService interface {
	// Single attribute operations
	CacheAttribute(ctx context.Context, key string, attrValue *models.AttributeValue, ttl time.Duration) error
	GetCachedAttribute(ctx context.Context, key string) (*models.AttributeValue, error)
	InvalidateAttribute(ctx context.Context, key string) error

	// Attribute collection operations
	CacheAttributeCollection(ctx context.Context, collectionKey string, attributes map[string]*models.AttributeValue, ttl time.Duration) error
	GetCachedAttributeCollection(ctx context.Context, collectionKey string) (map[string]*models.AttributeValue, error)
	InvalidateAttributeCollection(ctx context.Context, collectionKey string) error

	// Context operations
	CacheAttributeContext(ctx context.Context, contextKey string, attrContext *models.AttributeContext, ttl time.Duration) error
	GetCachedAttributeContext(ctx context.Context, contextKey string) (*models.AttributeContext, error)
	InvalidateAttributeContext(ctx context.Context, contextKey string) error

	// Bulk operations
	CacheMultipleAttributes(ctx context.Context, items map[string]*CacheItem) error
	GetMultipleCachedAttributes(ctx context.Context, keys []string) (map[string]*models.AttributeValue, error)
	InvalidateMultipleAttributes(ctx context.Context, keys []string) error

	// Pattern-based operations
	InvalidateByPattern(ctx context.Context, pattern string) error
	InvalidateByUserID(ctx context.Context, userID uuid.UUID) error
	InvalidateByResourceType(ctx context.Context, resourceType string) error

	// Cache management
	GetCacheStats(ctx context.Context) (*AttributeCacheStats, error)
	CleanupExpiredEntries(ctx context.Context) error
}

// CacheItem represents an item to be cached
type CacheItem struct {
	Key        string                    `json:"key"`
	Value      *models.AttributeValue    `json:"value"`
	TTL        time.Duration             `json:"ttl"`
	Categories []types.AttributeCategory `json:"categories"`
}

// AttributeCacheStats represents cache statistics
type AttributeCacheStats struct {
	TotalEntries         int64            `json:"total_entries"`
	HitRate              float64          `json:"hit_rate"`
	MissRate             float64          `json:"miss_rate"`
	ExpiredEntries       int64            `json:"expired_entries"`
	EntriesByCategory    map[string]int64 `json:"entries_by_category"`
	EntriesBySource      map[string]int64 `json:"entries_by_source"`
	AverageRetrievalTime time.Duration    `json:"average_retrieval_time"`
	CacheSizeBytes       int64            `json:"cache_size_bytes"`
}

// AttributeCacheWrapper wraps an attribute with cache metadata
type AttributeCacheWrapper struct {
	AttributeValue *models.AttributeValue `json:"attribute_value"`
	CachedAt       time.Time              `json:"cached_at"`
	ExpiresAt      time.Time              `json:"expires_at"`
	AccessCount    int64                  `json:"access_count"`
	LastAccessed   time.Time              `json:"last_accessed"`
}

// AttributeCollectionWrapper wraps a collection of attributes with cache metadata
type AttributeCollectionWrapper struct {
	Attributes   map[string]*models.AttributeValue `json:"attributes"`
	CachedAt     time.Time                         `json:"cached_at"`
	ExpiresAt    time.Time                         `json:"expires_at"`
	AccessCount  int64                             `json:"access_count"`
	LastAccessed time.Time                         `json:"last_accessed"`
}

// AttributeContextWrapper wraps an attribute context with cache metadata
type AttributeContextWrapper struct {
	AttributeContext *models.AttributeContext `json:"attribute_context"`
	CachedAt         time.Time                `json:"cached_at"`
	ExpiresAt        time.Time                `json:"expires_at"`
	AccessCount      int64                    `json:"access_count"`
	LastAccessed     time.Time                `json:"last_accessed"`
}

// CacheKeyBuilder provides methods for building cache keys
type CacheKeyBuilder struct {
	tenantID uuid.UUID
}

// NewCacheKeyBuilder creates a new cache key builder
func NewCacheKeyBuilder(tenantID uuid.UUID) *CacheKeyBuilder {
	return &CacheKeyBuilder{tenantID: tenantID}
}

// attributeCacheService implements AttributeCacheService
type attributeCacheService struct {
	cache   cache.Service
	tracing tracing.TracingService
	metrics metrics.MetricsProvider
	logger  logger.Logger
}

// NewAttributeCacheService creates a new attribute cache service
func NewAttributeCacheService(
	cache cache.Service,
	tracing tracing.TracingService,
	metrics metrics.MetricsProvider,
	logger logger.Logger,
) AttributeCacheService {
	return &attributeCacheService{
		cache:   cache,
		tracing: tracing,
		metrics: metrics,
		logger:  logger,
	}
}

// CacheAttribute caches a single attribute value
func (s *attributeCacheService) CacheAttribute(ctx context.Context, key string, attrValue *models.AttributeValue, ttl time.Duration) error {
	ctx, span := s.tracing.StartSpan(ctx, "attributeCacheService.CacheAttribute",
		tracing.WithAttributes(
			attribute.String("cache.key", key),
			attribute.String("attribute.name", attrValue.Name),
			attribute.String("attribute.category", string(attrValue.Category)),
		))
	defer span.End()

	s.logger.InfoContext(ctx, "Caching attribute",
		logger.Fields{
			"cache_key":      key,
			"attribute_name": attrValue.Name,
			"ttl_seconds":    ttl.Seconds(),
		})

	startTime := time.Now()

	// Create cache wrapper with metadata
	cacheWrapper := &AttributeCacheWrapper{
		AttributeValue: attrValue,
		CachedAt:       time.Now(),
		ExpiresAt:      time.Now().Add(ttl),
		AccessCount:    0,
		LastAccessed:   time.Now(),
	}

	// Cache the wrapped attribute
	if err := s.cache.Set(ctx, key, cacheWrapper, ttl); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to cache attribute")
		s.recordCacheMetrics(ctx, "cache", "error", time.Since(startTime))
		return errors.NewBusinessError("ATTRIBUTE_CACHE_FAILED", "Failed to cache attribute").
			WithDetail("cache_key", key).
			WithDetail("error", err.Error())
	}

	s.recordCacheMetrics(ctx, "cache", "success", time.Since(startTime))
	s.logger.InfoContext(ctx, "Attribute cached successfully",
		logger.Fields{
			"cache_key":      key,
			"attribute_name": attrValue.Name,
			"cache_time_ms":  time.Since(startTime).Milliseconds(),
		})

	return nil
}

// GetCachedAttribute retrieves a cached attribute value
func (s *attributeCacheService) GetCachedAttribute(ctx context.Context, key string) (*models.AttributeValue, error) {
	ctx, span := s.tracing.StartSpan(ctx, "attributeCacheService.GetCachedAttribute",
		tracing.WithAttributes(attribute.String("cache.key", key)))
	defer span.End()

	startTime := time.Now()

	var cacheWrapper AttributeCacheWrapper
	if err := s.cache.Get(ctx, key, &cacheWrapper); err != nil {
		s.recordCacheMetrics(ctx, "get", "miss", time.Since(startTime))
		return nil, errors.NewBusinessError("ATTRIBUTE_CACHE_MISS", "Attribute not found in cache").
			WithDetail("cache_key", key)
	}

	// Check if expired
	if time.Now().After(cacheWrapper.ExpiresAt) {
		s.recordCacheMetrics(ctx, "get", "expired", time.Since(startTime))
		// Remove expired entry
		_ = s.cache.Delete(ctx, key)
		return nil, errors.NewBusinessError("ATTRIBUTE_CACHE_EXPIRED", "Cached attribute has expired").
			WithDetail("cache_key", key)
	}

	// Update access metadata
	cacheWrapper.AccessCount++
	cacheWrapper.LastAccessed = time.Now()

	// Update the cache with new metadata (fire and forget)
	go func() {
		_ = s.cache.Set(context.Background(), key, &cacheWrapper, time.Until(cacheWrapper.ExpiresAt))
	}()

	s.recordCacheMetrics(ctx, "get", "hit", time.Since(startTime))
	s.logger.InfoContext(ctx, "Attribute retrieved from cache",
		logger.Fields{
			"cache_key":      key,
			"attribute_name": cacheWrapper.AttributeValue.Name,
			"access_count":   cacheWrapper.AccessCount,
			"retrieval_time": time.Since(startTime).Milliseconds(),
		})

	return cacheWrapper.AttributeValue, nil
}

// InvalidateAttribute removes an attribute from cache
func (s *attributeCacheService) InvalidateAttribute(ctx context.Context, key string) error {
	ctx, span := s.tracing.StartSpan(ctx, "attributeCacheService.InvalidateAttribute",
		tracing.WithAttributes(attribute.String("cache.key", key)))
	defer span.End()

	s.logger.InfoContext(ctx, "Invalidating cached attribute", logger.Fields{"cache_key": key})
	startTime := time.Now()

	if err := s.cache.Delete(ctx, key); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to invalidate attribute")
		s.recordCacheMetrics(ctx, "invalidate", "error", time.Since(startTime))
		return errors.NewBusinessError("ATTRIBUTE_CACHE_INVALIDATE_FAILED", "Failed to invalidate cached attribute").
			WithDetail("cache_key", key).
			WithDetail("error", err.Error())
	}

	s.recordCacheMetrics(ctx, "invalidate", "success", time.Since(startTime))
	s.logger.InfoContext(ctx, "Attribute invalidated successfully", logger.Fields{"cache_key": key})

	return nil
}

// CacheAttributeCollection caches a collection of attributes
func (s *attributeCacheService) CacheAttributeCollection(ctx context.Context, collectionKey string, attributes map[string]*models.AttributeValue, ttl time.Duration) error {
	ctx, span := s.tracing.StartSpan(ctx, "attributeCacheService.CacheAttributeCollection",
		tracing.WithAttributes(
			attribute.String("cache.collection_key", collectionKey),
			attribute.Int("attributes.count", len(attributes)),
		))
	defer span.End()

	s.logger.InfoContext(ctx, "Caching attribute collection",
		logger.Fields{
			"collection_key":  collectionKey,
			"attribute_count": len(attributes),
			"ttl_seconds":     ttl.Seconds(),
		})

	startTime := time.Now()

	// Create collection wrapper
	collectionWrapper := &AttributeCollectionWrapper{
		Attributes:   attributes,
		CachedAt:     time.Now(),
		ExpiresAt:    time.Now().Add(ttl),
		AccessCount:  0,
		LastAccessed: time.Now(),
	}

	// Cache the collection
	if err := s.cache.Set(ctx, collectionKey, collectionWrapper, ttl); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to cache attribute collection")
		s.recordCacheMetrics(ctx, "cache_collection", "error", time.Since(startTime))
		return errors.NewBusinessError("ATTRIBUTE_COLLECTION_CACHE_FAILED", "Failed to cache attribute collection").
			WithDetail("collection_key", collectionKey).
			WithDetail("error", err.Error())
	}

	s.recordCacheMetrics(ctx, "cache_collection", "success", time.Since(startTime))
	s.logger.InfoContext(ctx, "Attribute collection cached successfully",
		logger.Fields{
			"collection_key":  collectionKey,
			"attribute_count": len(attributes),
			"cache_time_ms":   time.Since(startTime).Milliseconds(),
		})

	return nil
}

// GetCachedAttributeCollection retrieves a cached attribute collection
func (s *attributeCacheService) GetCachedAttributeCollection(ctx context.Context, collectionKey string) (map[string]*models.AttributeValue, error) {
	ctx, span := s.tracing.StartSpan(ctx, "attributeCacheService.GetCachedAttributeCollection",
		tracing.WithAttributes(attribute.String("cache.collection_key", collectionKey)))
	defer span.End()

	startTime := time.Now()

	var collectionWrapper AttributeCollectionWrapper
	if err := s.cache.Get(ctx, collectionKey, &collectionWrapper); err != nil {
		s.recordCacheMetrics(ctx, "get_collection", "miss", time.Since(startTime))
		return nil, errors.NewBusinessError("ATTRIBUTE_COLLECTION_CACHE_MISS", "Attribute collection not found in cache").
			WithDetail("collection_key", collectionKey)
	}

	// Check if expired
	if time.Now().After(collectionWrapper.ExpiresAt) {
		s.recordCacheMetrics(ctx, "get_collection", "expired", time.Since(startTime))
		// Remove expired entry
		_ = s.cache.Delete(ctx, collectionKey)
		return nil, errors.NewBusinessError("ATTRIBUTE_COLLECTION_CACHE_EXPIRED", "Cached attribute collection has expired").
			WithDetail("collection_key", collectionKey)
	}

	// Update access metadata
	collectionWrapper.AccessCount++
	collectionWrapper.LastAccessed = time.Now()

	// Update the cache with new metadata (fire and forget)
	go func() {
		_ = s.cache.Set(context.Background(), collectionKey, &collectionWrapper, time.Until(collectionWrapper.ExpiresAt))
	}()

	s.recordCacheMetrics(ctx, "get_collection", "hit", time.Since(startTime))
	s.logger.InfoContext(ctx, "Attribute collection retrieved from cache",
		logger.Fields{
			"collection_key":  collectionKey,
			"attribute_count": len(collectionWrapper.Attributes),
			"access_count":    collectionWrapper.AccessCount,
			"retrieval_time":  time.Since(startTime).Milliseconds(),
		})

	return collectionWrapper.Attributes, nil
}

// InvalidateAttributeCollection removes an attribute collection from cache
func (s *attributeCacheService) InvalidateAttributeCollection(ctx context.Context, collectionKey string) error {
	ctx, span := s.tracing.StartSpan(ctx, "attributeCacheService.InvalidateAttributeCollection",
		tracing.WithAttributes(attribute.String("cache.collection_key", collectionKey)))
	defer span.End()

	s.logger.InfoContext(ctx, "Invalidating cached attribute collection", logger.Fields{"collection_key": collectionKey})
	startTime := time.Now()

	if err := s.cache.Delete(ctx, collectionKey); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to invalidate attribute collection")
		s.recordCacheMetrics(ctx, "invalidate_collection", "error", time.Since(startTime))
		return errors.NewBusinessError("ATTRIBUTE_COLLECTION_CACHE_INVALIDATE_FAILED", "Failed to invalidate cached attribute collection").
			WithDetail("collection_key", collectionKey).
			WithDetail("error", err.Error())
	}

	s.recordCacheMetrics(ctx, "invalidate_collection", "success", time.Since(startTime))
	s.logger.InfoContext(ctx, "Attribute collection invalidated successfully", logger.Fields{"collection_key": collectionKey})

	return nil
}

// CacheAttributeContext caches an attribute context
func (s *attributeCacheService) CacheAttributeContext(ctx context.Context, contextKey string, attrContext *models.AttributeContext, ttl time.Duration) error {
	ctx, span := s.tracing.StartSpan(ctx, "attributeCacheService.CacheAttributeContext",
		tracing.WithAttributes(attribute.String("cache.context_key", contextKey)))
	defer span.End()

	totalAttrs := len(attrContext.UserAttributes) + len(attrContext.ResourceAttributes) +
		len(attrContext.EnvironmentAttributes) + len(attrContext.SessionAttributes) +
		len(attrContext.EntityAttributes) + len(attrContext.ActionAttributes)

	s.logger.InfoContext(ctx, "Caching attribute context",
		logger.Fields{
			"context_key":      contextKey,
			"total_attributes": totalAttrs,
			"ttl_seconds":      ttl.Seconds(),
		})

	startTime := time.Now()

	// Create context wrapper
	contextWrapper := &AttributeContextWrapper{
		AttributeContext: attrContext,
		CachedAt:         time.Now(),
		ExpiresAt:        time.Now().Add(ttl),
		AccessCount:      0,
		LastAccessed:     time.Now(),
	}

	// Cache the context
	if err := s.cache.Set(ctx, contextKey, contextWrapper, ttl); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to cache attribute context")
		s.recordCacheMetrics(ctx, "cache_context", "error", time.Since(startTime))
		return errors.NewBusinessError("ATTRIBUTE_CONTEXT_CACHE_FAILED", "Failed to cache attribute context").
			WithDetail("context_key", contextKey).
			WithDetail("error", err.Error())
	}

	s.recordCacheMetrics(ctx, "cache_context", "success", time.Since(startTime))
	s.logger.InfoContext(ctx, "Attribute context cached successfully",
		logger.Fields{
			"context_key":      contextKey,
			"total_attributes": totalAttrs,
			"cache_time_ms":    time.Since(startTime).Milliseconds(),
		})

	return nil
}

// GetCachedAttributeContext retrieves a cached attribute context
func (s *attributeCacheService) GetCachedAttributeContext(ctx context.Context, contextKey string) (*models.AttributeContext, error) {
	ctx, span := s.tracing.StartSpan(ctx, "attributeCacheService.GetCachedAttributeContext",
		tracing.WithAttributes(attribute.String("cache.context_key", contextKey)))
	defer span.End()

	startTime := time.Now()

	var contextWrapper AttributeContextWrapper
	if err := s.cache.Get(ctx, contextKey, &contextWrapper); err != nil {
		s.recordCacheMetrics(ctx, "get_context", "miss", time.Since(startTime))
		return nil, errors.NewBusinessError("ATTRIBUTE_CONTEXT_CACHE_MISS", "Attribute context not found in cache").
			WithDetail("context_key", contextKey)
	}

	// Check if expired
	if time.Now().After(contextWrapper.ExpiresAt) {
		s.recordCacheMetrics(ctx, "get_context", "expired", time.Since(startTime))
		// Remove expired entry
		_ = s.cache.Delete(ctx, contextKey)
		return nil, errors.NewBusinessError("ATTRIBUTE_CONTEXT_CACHE_EXPIRED", "Cached attribute context has expired").
			WithDetail("context_key", contextKey)
	}

	// Update access metadata
	contextWrapper.AccessCount++
	contextWrapper.LastAccessed = time.Now()

	// Update the cache with new metadata (fire and forget)
	go func() {
		_ = s.cache.Set(context.Background(), contextKey, &contextWrapper, time.Until(contextWrapper.ExpiresAt))
	}()

	totalAttrs := len(contextWrapper.AttributeContext.GetAllAttributes())
	s.recordCacheMetrics(ctx, "get_context", "hit", time.Since(startTime))
	s.logger.InfoContext(ctx, "Attribute context retrieved from cache",
		logger.Fields{
			"context_key":      contextKey,
			"total_attributes": totalAttrs,
			"access_count":     contextWrapper.AccessCount,
			"retrieval_time":   time.Since(startTime).Milliseconds(),
		})

	return contextWrapper.AttributeContext, nil
}

// InvalidateAttributeContext removes an attribute context from cache
func (s *attributeCacheService) InvalidateAttributeContext(ctx context.Context, contextKey string) error {
	ctx, span := s.tracing.StartSpan(ctx, "attributeCacheService.InvalidateAttributeContext",
		tracing.WithAttributes(attribute.String("cache.context_key", contextKey)))
	defer span.End()

	s.logger.InfoContext(ctx, "Invalidating cached attribute context", logger.Fields{"context_key": contextKey})
	startTime := time.Now()

	if err := s.cache.Delete(ctx, contextKey); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to invalidate attribute context")
		s.recordCacheMetrics(ctx, "invalidate_context", "error", time.Since(startTime))
		return errors.NewBusinessError("ATTRIBUTE_CONTEXT_CACHE_INVALIDATE_FAILED", "Failed to invalidate cached attribute context").
			WithDetail("context_key", contextKey).
			WithDetail("error", err.Error())
	}

	s.recordCacheMetrics(ctx, "invalidate_context", "success", time.Since(startTime))
	s.logger.InfoContext(ctx, "Attribute context invalidated successfully", logger.Fields{"context_key": contextKey})

	return nil
}

// CacheMultipleAttributes caches multiple attributes in a batch
func (s *attributeCacheService) CacheMultipleAttributes(ctx context.Context, items map[string]*CacheItem) error {
	ctx, span := s.tracing.StartSpan(ctx, "attributeCacheService.CacheMultipleAttributes",
		tracing.WithAttributes(attribute.Int("items.count", len(items))))
	defer span.End()

	s.logger.InfoContext(ctx, "Caching multiple attributes", logger.Fields{"item_count": len(items)})
	startTime := time.Now()

	// Cache each item
	errors := make([]error, 0)
	successCount := 0

	for key, item := range items {
		if err := s.CacheAttribute(ctx, key, item.Value, item.TTL); err != nil {
			errors = append(errors, err)
			s.logger.WarnContext(ctx, "Failed to cache attribute in batch",
				logger.Fields{"cache_key": key, "error": err.Error()})
		} else {
			successCount++
		}
	}

	if len(errors) > 0 && successCount == 0 {
		// All failed
		s.recordCacheMetrics(ctx, "cache_multiple", "error", time.Since(startTime))
		return errors[0] // Return first error
	}

	if len(errors) > 0 {
		// Partial failure
		s.recordCacheMetrics(ctx, "cache_multiple", "partial", time.Since(startTime))
		s.logger.WarnContext(ctx, "Partial failure in batch cache operation",
			logger.Fields{
				"total_items":   len(items),
				"success_count": successCount,
				"failure_count": len(errors),
			})
	} else {
		// All succeeded
		s.recordCacheMetrics(ctx, "cache_multiple", "success", time.Since(startTime))
	}

	s.logger.InfoContext(ctx, "Multiple attributes cache operation completed",
		logger.Fields{
			"total_items":   len(items),
			"success_count": successCount,
			"failure_count": len(errors),
			"cache_time_ms": time.Since(startTime).Milliseconds(),
		})

	return nil
}

// GetMultipleCachedAttributes retrieves multiple cached attributes
func (s *attributeCacheService) GetMultipleCachedAttributes(ctx context.Context, keys []string) (map[string]*models.AttributeValue, error) {
	ctx, span := s.tracing.StartSpan(ctx, "attributeCacheService.GetMultipleCachedAttributes",
		tracing.WithAttributes(attribute.Int("keys.count", len(keys))))
	defer span.End()

	s.logger.InfoContext(ctx, "Getting multiple cached attributes", logger.Fields{"key_count": len(keys)})
	startTime := time.Now()

	result := make(map[string]*models.AttributeValue)
	hitCount := 0
	missCount := 0

	for _, key := range keys {
		if attrValue, err := s.GetCachedAttribute(ctx, key); err == nil {
			result[key] = attrValue
			hitCount++
		} else {
			missCount++
		}
	}

	// Calculate hit rate
	hitRate := float64(hitCount) / float64(len(keys))

	s.recordBatchCacheMetrics(ctx, "get_multiple", hitCount, missCount, time.Since(startTime))
	s.logger.InfoContext(ctx, "Multiple attributes retrieval completed",
		logger.Fields{
			"total_keys":     len(keys),
			"hit_count":      hitCount,
			"miss_count":     missCount,
			"hit_rate":       hitRate,
			"retrieval_time": time.Since(startTime).Milliseconds(),
		})

	return result, nil
}

// InvalidateMultipleAttributes removes multiple attributes from cache
func (s *attributeCacheService) InvalidateMultipleAttributes(ctx context.Context, keys []string) error {
	ctx, span := s.tracing.StartSpan(ctx, "attributeCacheService.InvalidateMultipleAttributes",
		tracing.WithAttributes(attribute.Int("keys.count", len(keys))))
	defer span.End()

	s.logger.InfoContext(ctx, "Invalidating multiple cached attributes", logger.Fields{"key_count": len(keys)})
	startTime := time.Now()

	errors := make([]error, 0)
	successCount := 0

	for _, key := range keys {
		if err := s.InvalidateAttribute(ctx, key); err != nil {
			errors = append(errors, err)
			s.logger.WarnContext(ctx, "Failed to invalidate attribute in batch",
				logger.Fields{"cache_key": key, "error": err.Error()})
		} else {
			successCount++
		}
	}

	if len(errors) > 0 && successCount == 0 {
		// All failed
		s.recordCacheMetrics(ctx, "invalidate_multiple", "error", time.Since(startTime))
		return errors[0] // Return first error
	}

	if len(errors) > 0 {
		// Partial failure
		s.recordCacheMetrics(ctx, "invalidate_multiple", "partial", time.Since(startTime))
	} else {
		// All succeeded
		s.recordCacheMetrics(ctx, "invalidate_multiple", "success", time.Since(startTime))
	}

	s.logger.InfoContext(ctx, "Multiple attributes invalidation completed",
		logger.Fields{
			"total_keys":         len(keys),
			"success_count":      successCount,
			"failure_count":      len(errors),
			"invalidate_time_ms": time.Since(startTime).Milliseconds(),
		})

	return nil
}

// InvalidateByPattern removes cached attributes matching a pattern
func (s *attributeCacheService) InvalidateByPattern(ctx context.Context, pattern string) error {
	ctx, span := s.tracing.StartSpan(ctx, "attributeCacheService.InvalidateByPattern",
		tracing.WithAttributes(attribute.String("cache.pattern", pattern)))
	defer span.End()

	s.logger.InfoContext(ctx, "Invalidating cached attributes by pattern", logger.Fields{"pattern": pattern})
	startTime := time.Now()

	if err := s.cache.DeletePattern(ctx, pattern); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to invalidate by pattern")
		s.recordCacheMetrics(ctx, "invalidate_pattern", "error", time.Since(startTime))
		return errors.NewBusinessError("ATTRIBUTE_CACHE_PATTERN_INVALIDATE_FAILED", "Failed to invalidate cached attributes by pattern").
			WithDetail("pattern", pattern).
			WithDetail("error", err.Error())
	}

	s.recordCacheMetrics(ctx, "invalidate_pattern", "success", time.Since(startTime))
	s.logger.InfoContext(ctx, "Attributes invalidated by pattern successfully",
		logger.Fields{
			"pattern":         pattern,
			"invalidate_time": time.Since(startTime).Milliseconds(),
		})

	return nil
}

// InvalidateByUserID removes cached attributes for a specific user
func (s *attributeCacheService) InvalidateByUserID(ctx context.Context, userID uuid.UUID) error {
	ctx, span := s.tracing.StartSpan(ctx, "attributeCacheService.InvalidateByUserID",
		tracing.WithAttributes(attribute.String("user.id", userID.String())))
	defer span.End()

	s.logger.InfoContext(ctx, "Invalidating cached attributes by user ID", logger.Fields{"user_id": userID})

	// Build patterns for user-related cache entries
	patterns := []string{
		fmt.Sprintf("user_attributes:%s", userID),
		fmt.Sprintf("attr_ctx:user:%s:*", userID),
		fmt.Sprintf("attr_collection:user:%s:*", userID),
	}

	for _, pattern := range patterns {
		if err := s.InvalidateByPattern(ctx, pattern); err != nil {
			s.logger.WarnContext(ctx, "Failed to invalidate pattern for user",
				logger.Fields{"user_id": userID, "pattern": pattern, "error": err.Error()})
		}
	}

	s.logger.InfoContext(ctx, "User attributes invalidated successfully", logger.Fields{"user_id": userID})
	return nil
}

// InvalidateByResourceType removes cached attributes for a specific resource type
func (s *attributeCacheService) InvalidateByResourceType(ctx context.Context, resourceType string) error {
	ctx, span := s.tracing.StartSpan(ctx, "attributeCacheService.InvalidateByResourceType",
		tracing.WithAttributes(attribute.String("resource.type", resourceType)))
	defer span.End()

	s.logger.InfoContext(ctx, "Invalidating cached attributes by resource type", logger.Fields{"resource_type": resourceType})

	// Build patterns for resource-related cache entries
	patterns := []string{
		fmt.Sprintf("resource_attributes:%s:*", resourceType),
		fmt.Sprintf("attr_ctx:resource:%s:*", resourceType),
		fmt.Sprintf("attr_collection:resource:%s:*", resourceType),
	}

	for _, pattern := range patterns {
		if err := s.InvalidateByPattern(ctx, pattern); err != nil {
			s.logger.WarnContext(ctx, "Failed to invalidate pattern for resource type",
				logger.Fields{"resource_type": resourceType, "pattern": pattern, "error": err.Error()})
		}
	}

	s.logger.InfoContext(ctx, "Resource type attributes invalidated successfully", logger.Fields{"resource_type": resourceType})
	return nil
}

// GetCacheStats returns cache statistics
func (s *attributeCacheService) GetCacheStats(ctx context.Context) (*AttributeCacheStats, error) {
	ctx, span := s.tracing.StartSpan(ctx, "attributeCacheService.GetCacheStats")
	defer span.End()

	// TODO: Implement cache statistics collection
	// This would require extending the cache service to provide statistics
	stats := &AttributeCacheStats{
		TotalEntries:         0,
		HitRate:              0.0,
		MissRate:             0.0,
		ExpiredEntries:       0,
		EntriesByCategory:    make(map[string]int64),
		EntriesBySource:      make(map[string]int64),
		AverageRetrievalTime: 0,
		CacheSizeBytes:       0,
	}

	return stats, nil
}

// CleanupExpiredEntries removes expired cache entries
func (s *attributeCacheService) CleanupExpiredEntries(ctx context.Context) error {
	ctx, span := s.tracing.StartSpan(ctx, "attributeCacheService.CleanupExpiredEntries")
	defer span.End()

	s.logger.InfoContext(ctx, "Cleaning up expired cache entries")
	startTime := time.Now()

	// TODO: Implement expired entry cleanup
	// This would require extending the cache service to support cleanup operations

	s.recordCacheMetrics(ctx, "cleanup", "success", time.Since(startTime))
	s.logger.InfoContext(ctx, "Cache cleanup completed",
		logger.Fields{"cleanup_time": time.Since(startTime).Milliseconds()})

	return nil
}

// ─── CACHE KEY BUILDERS ─────────────────────────────────────────────

// BuildUserAttributeKey builds a cache key for user attributes
func (b *CacheKeyBuilder) BuildUserAttributeKey(userID uuid.UUID) string {
	return fmt.Sprintf("user_attributes:%s:%s", b.tenantID, userID)
}

// BuildResourceAttributeKey builds a cache key for resource attributes
func (b *CacheKeyBuilder) BuildResourceAttributeKey(resourceType string, resourceID *uuid.UUID) string {
	if resourceID != nil {
		return fmt.Sprintf("resource_attributes:%s:%s:%s", b.tenantID, resourceType, *resourceID)
	}
	return fmt.Sprintf("resource_attributes:%s:%s:global", b.tenantID, resourceType)
}

// BuildEnvironmentAttributeKey builds a cache key for environment attributes
func (b *CacheKeyBuilder) BuildEnvironmentAttributeKey(contextHash string) string {
	return fmt.Sprintf("env_attributes:%s:%s", b.tenantID, contextHash)
}

// BuildSessionAttributeKey builds a cache key for session attributes
func (b *CacheKeyBuilder) BuildSessionAttributeKey(sessionID string) string {
	return fmt.Sprintf("session_attributes:%s:%s", b.tenantID, sessionID)
}

// BuildActionAttributeKey builds a cache key for action attributes
func (b *CacheKeyBuilder) BuildActionAttributeKey(action string) string {
	return fmt.Sprintf("action_attributes:%s:%s", b.tenantID, action)
}

// BuildEntityAttributeKey builds a cache key for entity attributes
func (b *CacheKeyBuilder) BuildEntityAttributeKey(entityID uuid.UUID) string {
	return fmt.Sprintf("entity_attributes:%s:%s", b.tenantID, entityID)
}

// BuildAttributeContextKey builds a cache key for attribute context
func (b *CacheKeyBuilder) BuildAttributeContextKey(req *AttributeCollectionRequest) string {
	// Create a hash of the request parameters for consistent key generation
	hash := b.hashAttributeCollectionRequest(req)
	return fmt.Sprintf("attr_ctx:%s:%s", b.tenantID, hash)
}

// hashAttributeCollectionRequest creates a hash of the collection request for cache key
func (b *CacheKeyBuilder) hashAttributeCollectionRequest(req *AttributeCollectionRequest) string {
	// Create a consistent string representation of the request
	var keyParts []string
	keyParts = append(keyParts, fmt.Sprintf("user:%s", req.UserID))
	keyParts = append(keyParts, fmt.Sprintf("resource:%s", req.ResourceType))

	if req.ResourceID != nil {
		keyParts = append(keyParts, fmt.Sprintf("resource_id:%s", *req.ResourceID))
	}

	keyParts = append(keyParts, fmt.Sprintf("action:%s", req.Action))

	if req.EntityID != nil {
		keyParts = append(keyParts, fmt.Sprintf("entity:%s", *req.EntityID))
	}

	// Add environment data hash if present
	if req.EnvironmentData != nil {
		envBytes, _ := json.Marshal(req.EnvironmentData)
		envHash := fmt.Sprintf("%x", sha256.Sum256(envBytes))
		keyParts = append(keyParts, fmt.Sprintf("env:%s", envHash[:16])) // Use first 16 chars
	}

	// Add session data hash if present
	if req.SessionData != nil {
		sessionBytes, _ := json.Marshal(req.SessionData)
		sessionHash := fmt.Sprintf("%x", sha256.Sum256(sessionBytes))
		keyParts = append(keyParts, fmt.Sprintf("session:%s", sessionHash[:16])) // Use first 16 chars
	}

	// Join all parts and create final hash
	keyString := strings.Join(keyParts, "|")
	finalHash := fmt.Sprintf("%x", sha256.Sum256([]byte(keyString)))

	return finalHash[:32] // Use first 32 characters for reasonable key length
}

// ─── HELPER METHODS ─────────────────────────────────────────────

// recordCacheMetrics records cache operation metrics
func (s *attributeCacheService) recordCacheMetrics(ctx context.Context, operation, status string, duration time.Duration) {
	// Cache operation counter
	counter := s.metrics.Counter(
		"abac_attribute_cache_operations_total",
		"Total number of attribute cache operations",
		"operation", "status",
	)

	counter.Inc(metrics.Fields{
		"operation": operation,
		"status":    status,
	})

	// Cache operation duration histogram
	histogram := s.metrics.Histogram(
		"abac_attribute_cache_operation_duration_seconds",
		"Duration of attribute cache operations",
		metrics.StandardHTTPDurationBuckets(),
		"operation", "status",
	)

	histogram.Observe(duration.Seconds(), metrics.Fields{
		"operation": operation,
		"status":    status,
	})
}

// recordBatchCacheMetrics records batch cache operation metrics
func (s *attributeCacheService) recordBatchCacheMetrics(ctx context.Context, operation string, hitCount, missCount int, duration time.Duration) {
	totalCount := hitCount + missCount
	hitRate := 0.0
	if totalCount > 0 {
		hitRate = float64(hitCount) / float64(totalCount)
	}

	// Batch operation counter
	counter := s.metrics.Counter(
		"abac_attribute_cache_batch_operations_total",
		"Total number of batch attribute cache operations",
		"operation",
	)

	counter.Inc(metrics.Fields{
		"operation": operation,
	})

	// Batch hit rate histogram
	hitRateHist := s.metrics.Histogram(
		"abac_attribute_cache_batch_hit_rate",
		"Hit rate for batch cache operations",
		[]float64{0.0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0},
		"operation",
	)

	hitRateHist.Observe(hitRate, metrics.Fields{
		"operation": operation,
	})

	// Batch size histogram
	sizeHist := s.metrics.Histogram(
		"abac_attribute_cache_batch_size",
		"Size of batch cache operations",
		[]float64{1, 5, 10, 20, 50, 100, 200, 500},
		"operation",
	)

	sizeHist.Observe(float64(totalCount), metrics.Fields{
		"operation": operation,
	})

	// Duration histogram
	durationHist := s.metrics.Histogram(
		"abac_attribute_cache_batch_duration_seconds",
		"Duration of batch cache operations",
		metrics.StandardHTTPDurationBuckets(),
		"operation",
	)

	durationHist.Observe(duration.Seconds(), metrics.Fields{
		"operation": operation,
	})
}

//
//
//   // hashAttributeCollectionRequest creates a hash of the collection request for cache key
//  func (b *CacheKeyBuilder) hashAttributeCollectionRequest(req *AttributeCollectionRequest) string {
//    // Create a consistent string representation of the request
//    var keyParts []string
//    keyParts = append(keyParts, fmt.Sprintf("user:%s",
// req.UserID))
//    keyParts = append(keyParts, fmt.Sprintf("resource:%s",
// req.ResourceType))
//
//    if req.ResourceID != nil {
//      keyParts = append(keyParts,
// fmt.Sprintf("resource_id:%s", *req.ResourceID))
//    }
//
//    keyParts = append(keyParts, fmt.Sprintf("action:%s",
// req.Action))
//
//    if req.EntityID != nil {
//      keyParts = append(keyParts, fmt.Sprintf("entity:%s",
// *req.EntityID))
//    }
//
//    // Add environment data hash if present
//    if req.EnvironmentData != nil {
//      envBytes, _ := json.Marshal(req.EnvironmentData)
//      envHash := fmt.Sprintf("%x", sha256.Sum256(envBytes))
//      keyParts = append(keyParts, fmt.Sprintf("env:%s",
// envHash[:16])) // Use first 16 chars
//    }
//
//    // Add session data hash if present
//    if req.SessionData != nil {
//      sessionBytes, _ := json.Marshal(req.SessionData)
//      sessionHash := fmt.Sprintf("%x",
// sha256.Sum256(sessionBytes))
//      keyParts = append(keyParts, fmt.Sprintf("session:%s",
// sessionHash[:16])) // Use first 16 chars
//    }
//
//    // Join all parts and create final hash
//    keyString := strings.Join(keyParts, "|")
//    finalHash := fmt.Sprintf("%x",
// sha256.Sum256([]byte(keyString)))
//
//    return finalHash[:32] // Use first 32 characters for reasonable key length
//  }
//
//  // ─── CACHE WRAPPER TYPES
// //─────────────────────────────────────────────
//
//  // AttributeCacheWrapper wraps an attribute value with 	cache metadata
//  type AttributeCacheWrapper struct {
//    AttributeValue *models.AttributeValue 	`json:"attribute_value"`
//    CachedAt       time.Time `json:"cached_at"`
//    ExpiresAt      time.Time `json:"expires_at"`
//    AccessCount    int64 `json:"access_count"`
//    LastAccessed   time.Time `json:"last_accessed"`
//  }
//
//  // AttributeCollectionWrapper wraps an attribute collection with cache metadata
//  type AttributeCollectionWrapper struct {
//    Attributes   map[string]*models.AttributeValue `json:"attributes"`
//    CachedAt     time.Time `json:"cached_at"`
//    ExpiresAt    time.Time `json:"expires_at"`
//    AccessCount  int64 `json:"access_count"`
//    LastAccessed time.Time `json:"last_accessed"`
//  }
//
//  // AttributeContextWrapper wraps an attribute context	 cache metadata
//  type AttributeContextWrapper struct {
//    AttributeContext *models.AttributeContext `json:"attribute_context"`
//    CachedAt         time.Time `json:"cached_at"`
//    ExpiresAt        time.Time `json:"expires_at"`
//    AccessCount      int64 `json:"access_count"`
//    LastAccessed     time.Time `json:"last_accessed"`
//  }
//
//  // ─── HELPER METHODS
// //─────────────────────────────────────────────
//
//  // recordCacheMetrics records cache operation metrics
//  func (s *attributeCacheService) recordCacheMetrics(ctx context.Context, operation, status string, duration time.Duration) {
//    // Cache operation counter
//    counter := s.metrics.Counter(
//      "abac_attribute_cache_operations_total",
//      "Total number of attribute cache operations",
//      "operation", "status",
//    )
//
//    counter.Inc(ctx, metrics.Fields{
//      "operation": operation,
//      "status":    status,
//    })
//
//    // Cache operation duration histogram
//    histogram := s.metrics.Histogram(
//      "abac_attribute_cache_operation_duration_seconds",
//      "Duration of attribute cache operations",
//      metrics.StandardHTTPDurationBuckets(),
//      "operation", "status",
//    )
// 	 histogram.Observe(ctx, duration.Seconds(),
// metrics.Fields{
//      "operation": operation,
//      "status":    status,
//    })
//  }
//
//  // recordBatchCacheMetrics records batch cache operation metrics
//  func (s *attributeCacheService) recordBatchCacheMetrics(ctx context.Context, operation string, hitCount, missCount int, duration time.Duration) {
//    totalCount := hitCount + missCount
//    hitRate := 0.0
//    if totalCount > 0 {
//      hitRate = float64(hitCount) / float64(totalCount)
//    }
//
//    // Batch operation counter
//    counter := s.metrics.Counter(
//      "abac_attribute_cache_batch_operations_total",
//      "Total number of batch attribute cache operations",
//      "operation",
//    )
//
//    counter.Inc(ctx, metrics.Fields{
//      "operation": operation,
//    })
//
//    // Batch hit rate histogram
//    hitRateHist := s.metrics.Histogram(
//      "abac_attribute_cache_batch_hit_rate",
//      "Hit rate for batch cache operations",
//      []float64{0.0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7,
//
//  0.9, 1.0},
//      "operation",
//    )
//
//    hitRateHist.Observe(ctx, hitRate, metrics.Fields{
//      "operation": operation,
//    })
//
//    // Batch size histogram
//    sizeHist := s.metrics.Histogram(
//      "abac_attribute_cache_batch_size",
//      "Size of batch cache operations",
//      []float64{1, 5, 10, 20, 50, 100, 200, 500},
//      "operation",
//    )
//
//    sizeHist.Observe(ctx, float64(totalCount),
// metrics.Fields{
//      "operation": operation,
//    })
//
//    // Duration histogram
//    durationHist := s.metrics.Histogram(
//      "abac_attribute_cache_batch_duration_seconds",
//      "Duration of batch cache operations",
//      metrics.StandardHTTPDurationBuckets(),
// 		      "operation",
//    )
//
//    durationHist.Observe(ctx, duration.Seconds(),
// metrics.Fields{
//      "operation": operation,
//    })
//  }
