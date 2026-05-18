package cache

import (
	"context"
	"fmt"

	platformCache "awo.so/internal/platform/cache"
	sharedErrors "awo.so/internal/shared/errors"
)

// InvalidationScope controls which cache entries are evicted.
type InvalidationScope int

const (
	// InvalidateTenant evicts all schema cache entries for a tenant.
	// Use when: tenant config changes, subscription changes, bulk role reassignment.
	InvalidateTenant InvalidationScope = iota

	// InvalidateModule evicts all schema entries for one module under a tenant.
	// Use when: a module's page registry or DSL blocks change for a specific tenant.
	InvalidateModule

	// InvalidatePolicy bumps the PolicyGeneration counter.
	// Soft invalidation: no keys deleted. New keys are generated on next request;
	// old keys become orphans that expire by TTL.
	// Use when: IAM policies change for a tenant (role assignment, permission update).
	InvalidatePolicy

	// InvalidateHard performs a full pattern-delete for a tenant.
	// Equivalent to InvalidateTenant but signals intent — use when stale data
	// would cause a security or correctness issue that cannot wait for TTL.
	InvalidateHard
)

// String implements fmt.Stringer for logging.
func (s InvalidationScope) String() string {
	switch s {
	case InvalidateTenant:
		return "tenant"
	case InvalidateModule:
		return "module"
	case InvalidatePolicy:
		return "policy"
	case InvalidateHard:
		return "hard"
	default:
		return "unknown"
	}
}

// InvalidateRequest carries the parameters for a single invalidation operation.
type InvalidateRequest struct {
	Scope    InvalidationScope
	TenantID string
	// Module is required when Scope == InvalidateModule.
	// Must be a route prefix (e.g. "finance", "hr/payroll").
	Module string
	// VersionStore is required when Scope == InvalidatePolicy.
	// It holds and increments the PolicyGeneration counter.
	VersionStore VersionStore
}

// VersionStore abstracts reading and incrementing the PolicyGeneration counter.
// Implementations can use Redis INCR, a DB sequence, or an atomic in-process counter.
type VersionStore interface {
	// IncrPolicyGeneration increments the PolicyGeneration for tenantID and
	// returns the new value as a string.
	IncrPolicyGeneration(ctx context.Context, tenantID string) (string, error)
}

// InvalidateSchemaCache dispatches the invalidation strategy for the given request.
// All hard evictions use cache.Service.DeletePattern — never direct Redis calls.
//
// Returns a *sharedErrors.BusinessError with code CACHE_INVALIDATION_FAILED on error.
func InvalidateSchemaCache(
	ctx context.Context,
	svc platformCache.Service,
	req InvalidateRequest,
) error {
	switch req.Scope {
	case InvalidateTenant, InvalidateHard:
		return invalidateByPattern(ctx, svc, req.TenantID,
			TenantPattern(req.TenantID), req.Scope)

	case InvalidateModule:
		if req.Module == "" {
			return sharedErrors.NewBusinessError(
				"CACHE_INVALIDATION_FAILED",
				"Module is required for InvalidateModule scope",
			).WithHTTPStatus(400).
				WithCategory(sharedErrors.CategorySystem).
				WithDetail("scope", req.Scope.String())
		}
		pattern := ModulePattern(req.TenantID, req.Module)
		return invalidateByPattern(ctx, svc, req.TenantID, pattern, req.Scope)

	case InvalidatePolicy:
		if req.VersionStore == nil {
			return sharedErrors.NewBusinessError(
				"CACHE_INVALIDATION_FAILED",
				"VersionStore is required for InvalidatePolicy scope",
			).WithHTTPStatus(500).
				WithCategory(sharedErrors.CategorySystem).
				WithDetail("scope", req.Scope.String())
		}
		_, err := req.VersionStore.IncrPolicyGeneration(ctx, req.TenantID)
		if err != nil {
			return sharedErrors.NewBusinessError(
				"CACHE_INVALIDATION_FAILED",
				fmt.Sprintf("failed to increment policy generation for tenant %s", req.TenantID),
			).WithHTTPStatus(500).
				WithCategory(sharedErrors.CategorySystem).
				WithDetail("tenant_id", req.TenantID).
				WithCause(err)
		}
		// Soft invalidation: no keys deleted — next request generates new key
		// with the incremented generation; old keys expire by TTL.
		return nil

	default:
		return sharedErrors.NewBusinessError(
			"CACHE_INVALIDATION_FAILED",
			fmt.Sprintf("unknown invalidation scope: %d", req.Scope),
		).WithHTTPStatus(500).
			WithCategory(sharedErrors.CategorySystem)
	}
}

// invalidateByPattern deletes all keys matching pattern using cache.Service.DeletePattern.
func invalidateByPattern(
	ctx context.Context,
	svc platformCache.Service,
	tenantID, pattern string,
	scope InvalidationScope,
) error {
	if err := svc.DeletePattern(ctx, pattern); err != nil {
		return sharedErrors.NewBusinessError(
			"CACHE_INVALIDATION_FAILED",
			fmt.Sprintf("failed to delete schema cache for tenant %s (scope=%s)", tenantID, scope),
		).WithHTTPStatus(500).
			WithCategory(sharedErrors.CategorySystem).
			WithDetail("tenant_id", tenantID).
			WithDetail("scope", scope.String()).
			WithDetail("pattern", pattern).
			WithCause(err)
	}
	return nil
}
