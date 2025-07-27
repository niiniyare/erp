package services

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

	// Pattern-based operations
	InvalidateByPattern(ctx context.Context, pattern string) error
	InvalidateByUserID(ctx context.Context, userID uuid.UUID) error
	InvalidateByResourceType(ctx context.Context, resourceType string) error

	// Cache management
	GetCacheStats(ctx context.Context) (*AttributeCacheStats, error)
	CleanupExpiredEntries(ctx context.Context) error
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

// AttributeCacheWrapper wraps an attribute value with cache metadata
type AttributeCacheWrapper struct {
	AttributeValue *models.AttributeValue `json:"attribute_value"`
	CachedAt       time.Time              `json:"cached_at"`
	ExpiresAt      time.Time              `json:"expires_at"`
	AccessCount    int64                  `json:"access_count"`
	LastAccessed   time.Time              `json:"last_accessed"`
}

// AttributeCollectionWrapper wraps an attribute collection with cache metadata
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

// attributeCacheService implements AttributeCacheService
type attributeCacheService struct {
	cache   cache.Service
	tracing tracing.TracingService
	metrics metrics.Provider
	logger  logger.Logger
}

// NewAttributeCacheService creates a new attribute cache service
func NewAttributeCacheService(
	cache cache.Service,
	tracing tracing.TracingService,
	metrics metrics.Provider,
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

	s.recordCacheMetrics(ctx, "get", "hit", time.Since(startTime))
	return cacheWrapper.AttributeValue, nil
}

// InvalidateAttribute removes an attribute from cache
func (s *attributeCacheService) InvalidateAttribute(ctx context.Context, key string) error {
	ctx, span := s.tracing.StartSpan(ctx, "attributeCacheService.InvalidateAttribute",
		tracing.WithAttributes(attribute.String("cache.key", key)))
	defer span.End()

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

	s.recordCacheMetrics(ctx, "get_collection", "hit", time.Since(startTime))
	return collectionWrapper.Attributes, nil
}

// InvalidateAttributeCollection removes an attribute collection from cache
func (s *attributeCacheService) InvalidateAttributeCollection(ctx context.Context, collectionKey string) error {
	ctx, span := s.tracing.StartSpan(ctx, "attributeCacheService.InvalidateAttributeCollection",
		tracing.WithAttributes(attribute.String("cache.collection_key", collectionKey)))
	defer span.End()

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
	return nil
}

// CacheAttributeContext caches an attribute context
func (s *attributeCacheService) CacheAttributeContext(ctx context.Context, contextKey string, attrContext *models.AttributeContext, ttl time.Duration) error {
	ctx, span := s.tracing.StartSpan(ctx, "attributeCacheService.CacheAttributeContext",
		tracing.WithAttributes(attribute.String("cache.context_key", contextKey)))
	defer span.End()

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

	s.recordCacheMetrics(ctx, "get_context", "hit", time.Since(startTime))
	return contextWrapper.AttributeContext, nil
}

// InvalidateAttributeContext removes an attribute context from cache
func (s *attributeCacheService) InvalidateAttributeContext(ctx context.Context, contextKey string) error {
	ctx, span := s.tracing.StartSpan(ctx, "attributeCacheService.InvalidateAttributeContext",
		tracing.WithAttributes(attribute.String("cache.context_key", contextKey)))
	defer span.End()

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
	return nil
}

// InvalidateByPattern removes cached attributes matching a pattern
func (s *attributeCacheService) InvalidateByPattern(ctx context.Context, pattern string) error {
	ctx, span := s.tracing.StartSpan(ctx, "attributeCacheService.InvalidateByPattern",
		tracing.WithAttributes(attribute.String("cache.pattern", pattern)))
	defer span.End()

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
	return nil
}

// InvalidateByUserID removes cached attributes for a specific user
func (s *attributeCacheService) InvalidateByUserID(ctx context.Context, userID uuid.UUID) error {
	ctx, span := s.tracing.StartSpan(ctx, "attributeCacheService.InvalidateByUserID",
		tracing.WithAttributes(attribute.String("user.id", userID.String())))
	defer span.End()

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

	return nil
}

// InvalidateByResourceType removes cached attributes for a specific resource type
func (s *attributeCacheService) InvalidateByResourceType(ctx context.Context, resourceType string) error {
	ctx, span := s.tracing.StartSpan(ctx, "attributeCacheService.InvalidateByResourceType",
		tracing.WithAttributes(attribute.String("resource.type", resourceType)))
	defer span.End()

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

	return nil
}

// GetCacheStats returns cache statistics
func (s *attributeCacheService) GetCacheStats(ctx context.Context) (*AttributeCacheStats, error) {
	ctx, span := s.tracing.StartSpan(ctx, "attributeCacheService.GetCacheStats")
	defer span.End()

	// TODO: Implement cache statistics collection
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

	startTime := time.Now()

	// TODO: Implement expired entry cleanup
	// This would require extending the cache service to support cleanup operations

	s.recordCacheMetrics(ctx, "cleanup", "success", time.Since(startTime))
	return nil
}

// recordCacheMetrics records cache operation metrics
func (s *attributeCacheService) recordCacheMetrics(ctx context.Context, operation, status string, duration time.Duration) {
	// Cache operation counter
	counter := s.metrics.Counter(
		"abac_attribute_cache_operations_total",
		"Total number of attribute cache operations",
		"operation", "status",
	)

	counter.Inc(ctx, metrics.Fields{
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

	histogram.Observe(ctx, duration.Seconds(), metrics.Fields{
		"operation": operation,
		"status":    status,
	})
}
