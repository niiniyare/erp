package stages

import (
	"context"
	"errors"
	"fmt"
	"time"

	"awo.so/internal/pipeline"
	"awo.so/internal/platform/cache"
	"awo.so/internal/web/authz"
	uicache "awo.so/internal/web/cache"
	"awo.so/internal/web/ui"
)

// Cache TTL for compiled schemas.
// Schemas are invalidated earlier via DeletePattern on role change.
const schemaCacheTTL = 5 * time.Minute

// ─── TASK 3a — CACHE LOOKUP STAGE ────────────────────────────────────────────
//
// CacheLookupStage computes the canonical cache key and performs a cache lookup.
// On a hit it writes the cached schema to DataKeySchema and jumps directly to
// the ResponseStage, skipping Registry → Compile → Normalize → Validate → Store.
//
// DESCRIPTION:
// Cache key = "ui:schema:{tenantID}:{route}:{permFP}:{flagFP}:{version}"
// All five components are required. Key is computed after AuthzStage so fingerprints
// are available. Cache lookup uses cache.Service.GetMemory (in-process LRU) first,
// then Redis. The cache.Service is tenant-aware — tenant context injected via
// cache.TenantIDKey in the Go context.
//
// WHY:
// After AuthzStage runs (once per request), the cache key is deterministic.
// Two users with the same role set and the same enabled feature flags will share
// one cache entry, dramatically reducing Casbin + PageFn invocations at scale.
//
// SECURITY: Cache key built AFTER AuthzStage. Building it before would mean
// the route serves whatever schema was cached for the LAST request with the
// same route, regardless of permissions — a privilege escalation defect.
//
// RISKS:
// Non-required: a cache failure does not abort schema compilation.
// Cache miss is not an error — it is the expected cold path.

// CacheLookupStage is Priority 30, Required false.
type CacheLookupStage struct {
	pipeline.BaseStage
	cacheSvc cache.Service
}

// NewCacheLookupStage constructs a CacheLookupStage.
func NewCacheLookupStage(svc cache.Service) *CacheLookupStage {
	return &CacheLookupStage{
		BaseStage: pipeline.BaseStage{
			StageName:       "ui.cache_lookup",
			StageOperations: []string{ui.OperationKey, ui.AppOperationKey},
			StagePriority:   ui.PriorityCache,
			StageRequired:   false, // cache failure must not abort pipeline
			StageDependsOn:  []string{"ui.authz"},
		},
		cacheSvc: svc,
	}
}

// Execute looks up the schema in cache. On hit, sets DataKeySchema + DataKeyCacheHit
// and jumps to ResponseStage. On miss, stores the cache key for CacheStoreStage.
func (s *CacheLookupStage) Execute(opCtx *pipeline.OperationContext) (pipeline.StageResult, error) {
	input, ok := opCtx.Input.(ui.UISchemaInput)
	if !ok || input.Route == "" {
		return pipeline.StageResult{Status: "skipped", Message: "no UISchemaInput in opCtx.Input"}, nil
	}

	permFP, _ := opCtx.Data[ui.DataKeyPermFingerprint].(string)
	flagFP, _ := opCtx.Data[ui.DataKeyFlagFingerprint].(string)

	if permFP == "" || flagFP == "" {
		// AuthzStage must run first. This is a pipeline ordering defect.
		return pipeline.StageResult{Status: "skipped", Message: "fingerprints not set — AuthzStage must run before CacheLookupStage"}, nil
	}

	cacheKey := buildCacheKey(opCtx, input.Route, permFP, flagFP)

	// Inject tenant context required by cache.Service.
	cacheCtx := withTenantContext(opCtx.Ctx, opCtx.TenantID.String())

	var schema ui.Schema
	err := s.cacheSvc.Get(cacheCtx, cacheKey, &schema)
	if err != nil {
		if errors.Is(err, cache.ErrCacheMiss) {
			// Normal cache miss — store key for CacheStoreStage and continue.
			return pipeline.StageResult{
				Status: "completed",
				Outputs: map[string]any{
					ui.DataKeyCacheKey: cacheKey,
					ui.DataKeyCacheHit: false,
				},
			}, nil
		}
		// Other errors (circuit open, Redis unavailable) — log and continue.
		return pipeline.StageResult{
			Status:  "skipped",
			Message: fmt.Sprintf("cache lookup error (non-fatal): %v", err),
			Outputs: map[string]any{
				ui.DataKeyCacheKey: cacheKey,
				ui.DataKeyCacheHit: false,
			},
		}, nil
	}

	// Cache hit — jump to ResponseStage, skip compile stages.
	return pipeline.StageResult{
		Status:      "completed",
		NextStageID: "ui.response",
		Message:     fmt.Sprintf("cache hit for key %s", cacheKey),
		Outputs: map[string]any{
			ui.DataKeyCacheKey: cacheKey,
			ui.DataKeyCacheHit: true,
			ui.DataKeySchema:   schema,
		},
	}, nil
}

// ─── TASK 3b — CACHE STORE STAGE ─────────────────────────────────────────────
//
// CacheStoreStage writes the compiled schema to cache after successful compilation
// and validation. Non-required: a cache write failure must not fail the request.
//
// DESCRIPTION:
// Reads DataKeySchema and DataKeyCacheKey. Writes to cache.Service with schemaCacheTTL.
// Does NOT write if DataKeyCacheHit is true (schema came from cache — no point storing).
//
// RISKS:
// Non-required: failure is logged by PipelineBuilder but does not abort.

// CacheStoreStage is Priority 80, Required false.
type CacheStoreStage struct {
	pipeline.BaseStage
	cacheSvc cache.Service
	ttl      time.Duration
}

// NewCacheStoreStage constructs a CacheStoreStage.
func NewCacheStoreStage(svc cache.Service) *CacheStoreStage {
	return &CacheStoreStage{
		BaseStage: pipeline.BaseStage{
			StageName:       "ui.cache_store",
			StageOperations: []string{ui.OperationKey, ui.AppOperationKey},
			StagePriority:   ui.PriorityCacheStore,
			StageRequired:   false,
			StageDependsOn:  []string{"ui.validate"},
		},
		cacheSvc: svc,
		ttl:      schemaCacheTTL,
	}
}

// Execute writes the compiled schema to cache.
func (s *CacheStoreStage) Execute(opCtx *pipeline.OperationContext) (pipeline.StageResult, error) {
	// Skip if this was a cache hit — schema already in cache.
	if hit, _ := opCtx.Data[ui.DataKeyCacheHit].(bool); hit {
		return pipeline.StageResult{Status: "skipped", Message: "cache hit — no store needed"}, nil
	}

	cacheKey, _ := opCtx.Data[ui.DataKeyCacheKey].(string)
	if cacheKey == "" {
		return pipeline.StageResult{Status: "skipped", Message: "no cache key — CacheLookupStage may have skipped"}, nil
	}

	schema, ok := opCtx.Data[ui.DataKeySchema].(ui.Schema)
	if !ok || len(schema) == 0 {
		return pipeline.StageResult{Status: "skipped", Message: "no compiled schema to store"}, nil
	}

	cacheCtx := withTenantContext(opCtx.Ctx, opCtx.TenantID.String())
	if err := s.cacheSvc.Set(cacheCtx, cacheKey, schema, s.ttl); err != nil {
		// Non-fatal: log via StageResult message, return nil error.
		return pipeline.StageResult{
			Status:  "completed",
			Message: fmt.Sprintf("cache store failed (non-fatal): %v", err),
		}, nil
	}

	return pipeline.StageResult{
		Status:  "completed",
		Message: fmt.Sprintf("schema cached at key %s ttl=%s", cacheKey, s.ttl),
	}, nil
}

// buildCacheKey constructs the cache key for the current request.
// Uses the generation-aware 8-component key when CacheVersions are injected
// (DataKeyCacheVersions set by wire.go). Falls back to the legacy 5-component
// key when versions are absent — safe during the migration window.
func buildCacheKey(opCtx *pipeline.OperationContext, route, permFP, flagFP string) string {
	tenantID := opCtx.TenantID.String()
	if v, ok := opCtx.Data[ui.DataKeyCacheVersions].(uicache.CacheVersions); ok {
		return uicache.Key(tenantID, route, permFP, flagFP, v)
	}
	// Legacy fallback — remove once all deployments inject CacheVersions.
	return authz.CacheKey(tenantID, route, permFP, flagFP)
}

// withTenantContext injects the tenant ID into ctx for cache.Service operations.
// cache.Service uses cache.TenantIDKey (contextKey("tenant_id")) for tenant isolation.
func withTenantContext(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, cache.TenantIDKey, tenantID)
}

var (
	_ pipeline.Stage = (*CacheLookupStage)(nil)
	_ pipeline.Stage = (*CacheStoreStage)(nil)
)
