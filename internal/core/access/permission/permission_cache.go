package permission

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.opentelemetry.io/otel/attribute"
)

// PermissionCacheService provides specialized caching for permission evaluation results
type PermissionCacheService struct {
	cache   cache.Service
	metrics metrics.MetricsProvider
	tracing tracing.Service
}

// NewPermissionCacheService creates a new permission cache service
func NewPermissionCacheService(cache cache.Service, metrics metrics.MetricsProvider, tracing tracing.Service) *PermissionCacheService {
	return &PermissionCacheService{
		cache:   cache,
		metrics: metrics,
		tracing: tracing,
	}
}

// Cache key patterns for different permission data
const (
	// Permission evaluation results
	PermissionEvaluationCacheKey = "perm:eval:%s"      // Hash of evaluation request
	UserPermissionsCacheKey      = "user:perms:%s:%s"  // user_id:entity_id
	UserRolesCacheKey            = "user:roles:%s"     // user_id
	RolePermissionsCacheKey      = "role:perms:%s"     // role_id
	RoleHierarchyCacheKey        = "role:hierarchy:%s" // role_id
	PolicyEvaluationCacheKey     = "policy:eval:%s"    // Hash of policy evaluation request

	// User session and context caching
	UserSessionCacheKey = "user:session:%s" // session_id
	UserContextCacheKey = "user:context:%s" // user_id

	// Bulk evaluation results
	BulkEvaluationCacheKey = "bulk:eval:%s" // Hash of bulk evaluation request

	// Statistical and analytics caching
	UserBehaviorCacheKey   = "user:behavior:%s"   // user_id
	RiskAssessmentCacheKey = "risk:assessment:%s" // user_id
)

// Cache TTL configurations
const (
	// Short-lived cache for frequently changing data
	ShortCacheTTL = 5 * time.Minute

	// Medium-lived cache for moderately stable data
	MediumCacheTTL = 15 * time.Minute

	// Long-lived cache for stable data
	LongCacheTTL = 1 * time.Hour

	// Session cache TTL (matches typical session timeout)
	SessionCacheTTL = 30 * time.Minute

	// Behavioral data cache TTL
	BehaviorCacheTTL = 24 * time.Hour
)

// CachedPermissionEvaluationResult represents a cached permission evaluation result
type CachedPermissionEvaluationResult struct {
	Result       *PermissionEvaluationResult `json:"result"`
	CachedAt     time.Time                   `json:"cached_at"`
	ExpiresAt    time.Time                   `json:"expires_at"`
	CacheVersion string                      `json:"cache_version"`
	UserID       uuid.UUID                   `json:"user_id"`
	ResourceName string                      `json:"resource_name"`
	ActionName   string                      `json:"action_name"`
	EntityID     *uuid.UUID                  `json:"entity_id,omitempty"`
	ContextHash  string                      `json:"context_hash"`
}

// CachedUserPermissions represents cached user permissions
type CachedUserPermissions struct {
	UserID       uuid.UUID              `json:"user_id"`
	EntityID     *uuid.UUID             `json:"entity_id,omitempty"`
	Permissions  []*EffectivePermission `json:"permissions"`
	CachedAt     time.Time              `json:"cached_at"`
	ExpiresAt    time.Time              `json:"expires_at"`
	CacheVersion string                 `json:"cache_version"`
}

// CachedUserRoles represents cached user roles
type CachedUserRoles struct {
	UserID       uuid.UUID   `json:"user_id"`
	Roles        []*UserRole `json:"roles"`
	CachedAt     time.Time   `json:"cached_at"`
	ExpiresAt    time.Time   `json:"expires_at"`
	CacheVersion string      `json:"cache_version"`
}

// CachedRoleHierarchy represents cached role hierarchy
type CachedRoleHierarchy struct {
	RoleID       uuid.UUID `json:"role_id"`
	Hierarchy    []*Role   `json:"hierarchy"`
	CachedAt     time.Time `json:"cached_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	CacheVersion string    `json:"cache_version"`
}

// GeneratePermissionEvaluationKey generates a cache key for permission evaluation
func (pcs *PermissionCacheService) GeneratePermissionEvaluationKey(req *PermissionEvaluationRequest) string {
	// Create a hash of the request parameters for a stable cache key
	hasher := sha256.New()
	hasher.Write([]byte(fmt.Sprintf("%s:%s:%s", req.UserID.String(), req.ResourceName, req.ActionName)))

	if req.EntityID != nil {
		hasher.Write([]byte(fmt.Sprintf(":%s", req.EntityID.String())))
	}

	// Include context in the hash if present
	if req.Context != nil {
		for k, v := range req.Context {
			hasher.Write([]byte(fmt.Sprintf(":%s=%v", k, v)))
		}
	}

	hash := hex.EncodeToString(hasher.Sum(nil))
	return fmt.Sprintf(PermissionEvaluationCacheKey, hash)
}

// GetPermissionEvaluationResult retrieves cached permission evaluation result
func (pcs *PermissionCacheService) GetPermissionEvaluationResult(ctx context.Context, req *PermissionEvaluationRequest) (*PermissionEvaluationResult, bool, error) {
	ctx, span := pcs.tracing.StartSpan(ctx, "cache.get_permission_evaluation",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("user_id", req.UserID.String()),
			attribute.String("resource_name", req.ResourceName),
			attribute.String("action_name", req.ActionName),
		))
	defer span.End()

	key := pcs.GeneratePermissionEvaluationKey(req)

	timer := pcs.metrics.Timer("cache_get_duration", metrics.Fields{
		"cache_type": "permission_evaluation",
		"operation":  "get",
	})
	defer timer.Stop()

	var cached CachedPermissionEvaluationResult
	err := pcs.cache.Get(ctx, key, &cached)
	if err != nil {
		// Cache miss
		pcs.metrics.IncrementCounter("cache_operations_total", metrics.Fields{
			"cache_type": "permission_evaluation",
			"operation":  "get",
			"result":     "miss",
		})
		return nil, false, nil
	}

	// Check if cache entry is expired
	if time.Now().After(cached.ExpiresAt) {
		pcs.metrics.IncrementCounter("cache_operations_total", metrics.Fields{
			"cache_type": "permission_evaluation",
			"operation":  "get",
			"result":     "expired",
		})
		// Delete expired entry
		go pcs.cache.Delete(context.Background(), key)
		return nil, false, nil
	}

	// Cache hit
	pcs.metrics.IncrementCounter("cache_operations_total", metrics.Fields{
		"cache_type": "permission_evaluation",
		"operation":  "get",
		"result":     "hit",
	})

	// Update cache hit information in the result
	cached.Result.CacheHit = true
	cached.Result.EvaluationTimeMS = 0 // Cache hit has no evaluation time

	logger.DebugContext(ctx, "Permission evaluation cache hit", logger.Fields{
		"user_id":       req.UserID.String(),
		"resource_name": req.ResourceName,
		"action_name":   req.ActionName,
		"cache_key":     key,
	})

	return cached.Result, true, nil
}

// SetPermissionEvaluationResult caches permission evaluation result
func (pcs *PermissionCacheService) SetPermissionEvaluationResult(ctx context.Context, req *PermissionEvaluationRequest, result *PermissionEvaluationResult) error {
	ctx, span := pcs.tracing.StartSpan(ctx, "cache.set_permission_evaluation",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("user_id", req.UserID.String()),
			attribute.String("resource_name", req.ResourceName),
			attribute.String("action_name", req.ActionName),
			attribute.Bool("permission_allowed", result.Allowed),
		))
	defer span.End()

	key := pcs.GeneratePermissionEvaluationKey(req)

	timer := pcs.metrics.Timer("cache_set_duration", metrics.Fields{
		"cache_type": "permission_evaluation",
		"operation":  "set",
	})
	defer timer.Stop()

	cached := CachedPermissionEvaluationResult{
		Result:       result,
		CachedAt:     time.Now(),
		ExpiresAt:    time.Now().Add(MediumCacheTTL),
		CacheVersion: "1.0",
		UserID:       req.UserID,
		ResourceName: req.ResourceName,
		ActionName:   req.ActionName,
		EntityID:     req.EntityID,
		ContextHash:  pcs.GeneratePermissionEvaluationKey(req),
	}

	err := pcs.cache.Set(ctx, key, cached, MediumCacheTTL)
	if err != nil {
		pcs.metrics.IncrementCounter("cache_operations_total", metrics.Fields{
			"cache_type": "permission_evaluation",
			"operation":  "set",
			"result":     "error",
		})
		return err
	}

	pcs.metrics.IncrementCounter("cache_operations_total", metrics.Fields{
		"cache_type": "permission_evaluation",
		"operation":  "set",
		"result":     "success",
	})

	logger.DebugContext(ctx, "Permission evaluation result cached", logger.Fields{
		"user_id":       req.UserID.String(),
		"resource_name": req.ResourceName,
		"action_name":   req.ActionName,
		"cache_key":     key,
		"allowed":       result.Allowed,
	})

	return nil
}

// GetUserPermissions retrieves cached user permissions
func (pcs *PermissionCacheService) GetUserPermissions(ctx context.Context, userID uuid.UUID, entityID *uuid.UUID) ([]*EffectivePermission, bool, error) {
	ctx, span := pcs.tracing.StartSpan(ctx, "cache.get_user_permissions",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("user_id", userID.String()),
		))
	defer span.End()

	entityIDStr := ""
	if entityID != nil {
		entityIDStr = entityID.String()
	}

	key := fmt.Sprintf(UserPermissionsCacheKey, userID.String(), entityIDStr)

	timer := pcs.metrics.Timer("cache_get_duration", metrics.Fields{
		"cache_type": "user_permissions",
		"operation":  "get",
	})
	defer timer.Stop()

	var cached CachedUserPermissions
	err := pcs.cache.Get(ctx, key, &cached)
	if err != nil {
		pcs.metrics.IncrementCounter("cache_operations_total", metrics.Fields{
			"cache_type": "user_permissions",
			"operation":  "get",
			"result":     "miss",
		})
		return nil, false, nil
	}

	// Check if cache entry is expired
	if time.Now().After(cached.ExpiresAt) {
		pcs.metrics.IncrementCounter("cache_operations_total", metrics.Fields{
			"cache_type": "user_permissions",
			"operation":  "get",
			"result":     "expired",
		})
		go pcs.cache.Delete(context.Background(), key)
		return nil, false, nil
	}

	pcs.metrics.IncrementCounter("cache_operations_total", metrics.Fields{
		"cache_type": "user_permissions",
		"operation":  "get",
		"result":     "hit",
	})

	return cached.Permissions, true, nil
}

// SetUserPermissions caches user permissions
func (pcs *PermissionCacheService) SetUserPermissions(ctx context.Context, userID uuid.UUID, entityID *uuid.UUID, permissions []*EffectivePermission) error {
	ctx, span := pcs.tracing.StartSpan(ctx, "cache.set_user_permissions",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("user_id", userID.String()),
			attribute.Int("permissions_count", len(permissions)),
		))
	defer span.End()

	entityIDStr := ""
	if entityID != nil {
		entityIDStr = entityID.String()
	}

	key := fmt.Sprintf(UserPermissionsCacheKey, userID.String(), entityIDStr)

	timer := pcs.metrics.Timer("cache_set_duration", metrics.Fields{
		"cache_type": "user_permissions",
		"operation":  "set",
	})
	defer timer.Stop()

	cached := CachedUserPermissions{
		UserID:       userID,
		EntityID:     entityID,
		Permissions:  permissions,
		CachedAt:     time.Now(),
		ExpiresAt:    time.Now().Add(LongCacheTTL),
		CacheVersion: "1.0",
	}

	err := pcs.cache.Set(ctx, key, cached, LongCacheTTL)
	if err != nil {
		pcs.metrics.IncrementCounter("cache_operations_total", metrics.Fields{
			"cache_type": "user_permissions",
			"operation":  "set",
			"result":     "error",
		})
		return err
	}

	pcs.metrics.IncrementCounter("cache_operations_total", metrics.Fields{
		"cache_type": "user_permissions",
		"operation":  "set",
		"result":     "success",
	})

	return nil
}

// GetUserRoles retrieves cached user roles
func (pcs *PermissionCacheService) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*UserRole, bool, error) {
	key := fmt.Sprintf(UserRolesCacheKey, userID.String())

	timer := pcs.metrics.Timer("cache_get_duration", metrics.Fields{
		"cache_type": "user_roles",
		"operation":  "get",
	})
	defer timer.Stop()

	var cached CachedUserRoles
	err := pcs.cache.Get(ctx, key, &cached)
	if err != nil {
		pcs.metrics.IncrementCounter("cache_operations_total", metrics.Fields{
			"cache_type": "user_roles",
			"operation":  "get",
			"result":     "miss",
		})
		return nil, false, nil
	}

	// Check if cache entry is expired
	if time.Now().After(cached.ExpiresAt) {
		pcs.metrics.IncrementCounter("cache_operations_total", metrics.Fields{
			"cache_type": "user_roles",
			"operation":  "get",
			"result":     "expired",
		})
		go pcs.cache.Delete(context.Background(), key)
		return nil, false, nil
	}

	pcs.metrics.IncrementCounter("cache_operations_total", metrics.Fields{
		"cache_type": "user_roles",
		"operation":  "get",
		"result":     "hit",
	})

	return cached.Roles, true, nil
}

// SetUserRoles caches user roles
func (pcs *PermissionCacheService) SetUserRoles(ctx context.Context, userID uuid.UUID, roles []*UserRole) error {
	key := fmt.Sprintf(UserRolesCacheKey, userID.String())

	timer := pcs.metrics.Timer("cache_set_duration", metrics.Fields{
		"cache_type": "user_roles",
		"operation":  "set",
	})
	defer timer.Stop()

	cached := CachedUserRoles{
		UserID:       userID,
		Roles:        roles,
		CachedAt:     time.Now(),
		ExpiresAt:    time.Now().Add(LongCacheTTL),
		CacheVersion: "1.0",
	}

	err := pcs.cache.Set(ctx, key, cached, LongCacheTTL)
	if err != nil {
		pcs.metrics.IncrementCounter("cache_operations_total", metrics.Fields{
			"cache_type": "user_roles",
			"operation":  "set",
			"result":     "error",
		})
		return err
	}

	pcs.metrics.IncrementCounter("cache_operations_total", metrics.Fields{
		"cache_type": "user_roles",
		"operation":  "set",
		"result":     "success",
	})

	return nil
}

// GetRoleHierarchy retrieves cached role hierarchy
func (pcs *PermissionCacheService) GetRoleHierarchy(ctx context.Context, roleID uuid.UUID) ([]*Role, bool, error) {
	key := fmt.Sprintf(RoleHierarchyCacheKey, roleID.String())

	timer := pcs.metrics.Timer("cache_get_duration", metrics.Fields{
		"cache_type": "role_hierarchy",
		"operation":  "get",
	})
	defer timer.Stop()

	var cached CachedRoleHierarchy
	err := pcs.cache.Get(ctx, key, &cached)
	if err != nil {
		pcs.metrics.IncrementCounter("cache_operations_total", metrics.Fields{
			"cache_type": "role_hierarchy",
			"operation":  "get",
			"result":     "miss",
		})
		return nil, false, nil
	}

	// Check if cache entry is expired
	if time.Now().After(cached.ExpiresAt) {
		pcs.metrics.IncrementCounter("cache_operations_total", metrics.Fields{
			"cache_type": "role_hierarchy",
			"operation":  "get",
			"result":     "expired",
		})
		go pcs.cache.Delete(context.Background(), key)
		return nil, false, nil
	}

	pcs.metrics.IncrementCounter("cache_operations_total", metrics.Fields{
		"cache_type": "role_hierarchy",
		"operation":  "get",
		"result":     "hit",
	})

	return cached.Hierarchy, true, nil
}

// SetRoleHierarchy caches role hierarchy
func (pcs *PermissionCacheService) SetRoleHierarchy(ctx context.Context, roleID uuid.UUID, hierarchy []*Role) error {
	key := fmt.Sprintf(RoleHierarchyCacheKey, roleID.String())

	timer := pcs.metrics.Timer("cache_set_duration", metrics.Fields{
		"cache_type": "role_hierarchy",
		"operation":  "set",
	})
	defer timer.Stop()

	cached := CachedRoleHierarchy{
		RoleID:       roleID,
		Hierarchy:    hierarchy,
		CachedAt:     time.Now(),
		ExpiresAt:    time.Now().Add(LongCacheTTL),
		CacheVersion: "1.0",
	}

	err := pcs.cache.Set(ctx, key, cached, LongCacheTTL)
	if err != nil {
		pcs.metrics.IncrementCounter("cache_operations_total", metrics.Fields{
			"cache_type": "role_hierarchy",
			"operation":  "set",
			"result":     "error",
		})
		return err
	}

	pcs.metrics.IncrementCounter("cache_operations_total", metrics.Fields{
		"cache_type": "role_hierarchy",
		"operation":  "set",
		"result":     "success",
	})

	return nil
}

// InvalidateUserPermissions invalidates all cached permissions for a user
func (pcs *PermissionCacheService) InvalidateUserPermissions(ctx context.Context, userID uuid.UUID) error {
	ctx, span := pcs.tracing.StartSpan(ctx, "cache.invalidate_user_permissions",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("user_id", userID.String()),
		))
	defer span.End()

	// Delete user permissions cache
	userPermKey := fmt.Sprintf(UserPermissionsCacheKey, userID.String(), "")
	pcs.cache.Delete(ctx, userPermKey)

	// Delete user roles cache
	userRolesKey := fmt.Sprintf(UserRolesCacheKey, userID.String())
	pcs.cache.Delete(ctx, userRolesKey)

	// Note: In a production system, you might want to implement a more sophisticated
	// cache invalidation strategy using cache tags or patterns

	pcs.metrics.IncrementCounter("cache_operations_total", metrics.Fields{
		"cache_type": "user_permissions",
		"operation":  "invalidate",
		"result":     "success",
	})

	logger.InfoContext(ctx, "Invalidated user permissions cache", logger.Fields{
		"user_id": userID.String(),
	})

	return nil
}

// InvalidateRoleCache invalidates all cached data for a role
func (pcs *PermissionCacheService) InvalidateRoleCache(ctx context.Context, roleID uuid.UUID) error {
	ctx, span := pcs.tracing.StartSpan(ctx, "cache.invalidate_role_cache",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("role_id", roleID.String()),
		))
	defer span.End()

	// Delete role hierarchy cache
	hierarchyKey := fmt.Sprintf(RoleHierarchyCacheKey, roleID.String())
	pcs.cache.Delete(ctx, hierarchyKey)

	// Delete role permissions cache
	permissionsKey := fmt.Sprintf(RolePermissionsCacheKey, roleID.String())
	pcs.cache.Delete(ctx, permissionsKey)

	pcs.metrics.IncrementCounter("cache_operations_total", metrics.Fields{
		"cache_type": "role_cache",
		"operation":  "invalidate",
		"result":     "success",
	})

	logger.InfoContext(ctx, "Invalidated role cache", logger.Fields{
		"role_id": roleID.String(),
	})

	return nil
}

// FlushPermissionCache flushes all permission-related cache entries
func (pcs *PermissionCacheService) FlushPermissionCache(ctx context.Context) error {
	ctx, span := pcs.tracing.StartSpan(ctx, "cache.flush_permission_cache",
		tracing.WithSpanKind(tracing.SpanKindInternal))
	defer span.End()

	// In a production system, you would implement pattern-based cache deletion
	// For now, we'll flush the entire cache (not recommended for production)
	err := pcs.cache.Flush(ctx)
	if err != nil {
		pcs.metrics.IncrementCounter("cache_operations_total", metrics.Fields{
			"cache_type": "permission_cache",
			"operation":  "flush",
			"result":     "error",
		})
		return err
	}

	pcs.metrics.IncrementCounter("cache_operations_total", metrics.Fields{
		"cache_type": "permission_cache",
		"operation":  "flush",
		"result":     "success",
	})

	logger.InfoContext(ctx, "Flushed permission cache", logger.Fields{})

	return nil
}

// GetCacheStats returns cache statistics
func (pcs *PermissionCacheService) GetCacheStats(ctx context.Context) map[string]any {
	// In a real implementation, you would collect and return cache statistics
	// This is a simplified version
	return map[string]any{
		"cache_type": "permission_cache",
		"status":     "active",
		"ttl_config": map[string]any{
			"short_ttl":    ShortCacheTTL.String(),
			"medium_ttl":   MediumCacheTTL.String(),
			"long_ttl":     LongCacheTTL.String(),
			"session_ttl":  SessionCacheTTL.String(),
			"behavior_ttl": BehaviorCacheTTL.String(),
		},
	}
}
