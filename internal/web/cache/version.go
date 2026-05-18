// Package cache provides generation-aware cache key construction and invalidation
// for the UI schema compilation pipeline.
//
// Why generation versioning?
// The original cache key used only tenantID + route + permFP + flagFP + SchemaVersion,
// where SchemaVersion was a compile-time constant. This is insufficient:
//
//   - IAM policy updates do not increment SchemaVersion → users see stale schemas
//     until the 5-minute TTL expires (privilege escalation window).
//   - Schema/builder changes require a manual constant bump and redeploy.
//   - Multi-instance deployments can have mixed versions in cache.
//
// CacheVersions makes all components injectable at startup from config/env,
// so any generation increment automatically busts the relevant cache entries
// without a Redis SCAN+DEL sweep.
package cache

import (
	"fmt"
	"strings"
)

// CacheVersions holds all generation and version components included in the
// UI schema cache key. Injected once at application startup via NewUIPipeline.
//
// Changing any field causes all cache keys that include it to produce a different
// string — making existing cached schemas invisible (soft invalidation).
//
// Fields are intentionally strings so they can hold semver, git SHA, epoch
// timestamps, or monotonic counters depending on the invalidation strategy.
type CacheVersions struct {
	// CompilerVersion identifies the UI compiler build.
	// Set from BUILD_VERSION env or git SHA at startup.
	// Increment when CompileTree or any Node.Compile() output format changes.
	CompilerVersion string

	// ASTVersion identifies the AST node schema contract version.
	// Increment when node Compile() output changes in a breaking way.
	ASTVersion string

	// PolicyGeneration is a monotonic counter incremented when IAM policies change.
	// Soft-invalidates all cached schemas without eviction: the old keys simply
	// become unreachable (new key → cache miss → recompile).
	//
	// Source: read from a shared atomic counter, Redis key, or IAM event timestamp.
	// Updated by InvalidatePolicy in invalidation.go.
	PolicyGeneration string

	// SchemaGeneration is a monotonic counter incremented when the page registry
	// changes (e.g. after a deploy that adds/removes pages or changes DSL blocks).
	// Set from BUILD_VERSION or a deploy sequence number at startup.
	SchemaGeneration string
}

// Validate checks that all required version fields are populated.
// Called at startup — missing fields are programmer errors.
func (v CacheVersions) Validate() error {
	var missing []string
	if v.CompilerVersion == "" {
		missing = append(missing, "CompilerVersion")
	}
	if v.ASTVersion == "" {
		missing = append(missing, "ASTVersion")
	}
	if v.PolicyGeneration == "" {
		missing = append(missing, "PolicyGeneration")
	}
	if v.SchemaGeneration == "" {
		missing = append(missing, "SchemaGeneration")
	}
	if len(missing) > 0 {
		return fmt.Errorf("uicache: CacheVersions missing required fields: %s", strings.Join(missing, ", "))
	}
	return nil
}

// DefaultVersions returns a CacheVersions populated with safe defaults.
// Use in tests and local development. Production must inject real values.
func DefaultVersions() CacheVersions {
	return CacheVersions{
		CompilerVersion:  "dev",
		ASTVersion:       "v1",
		PolicyGeneration: "0",
		SchemaGeneration: "0",
	}
}

// Key builds the canonical 8-component UI schema cache key.
//
// Format: ui:schema:{tenantID}:{route}:{permFP}:{flagFP}:{compilerVer}:{astVer}:{policyGen}:{schemaGen}
//
// All 8 components are required. Missing any one risks:
//   - tenantID absent       → cross-tenant schema serving (security defect)
//   - permFP absent         → wrong permissions for a session (security defect)
//   - flagFP absent         → wrong feature flags for a session
//   - compilerVersion absent → stale schemas after compiler changes
//   - astVersion absent     → stale schemas after node format changes
//   - policyGeneration absent→ stale schemas after IAM policy changes
//   - schemaGeneration absent→ stale schemas after page registry changes
//
// Route is normalised: leading slash removed, remaining slashes replaced with ":".
func Key(tenantID, route, permFP, flagFP string, v CacheVersions) string {
	route = strings.TrimPrefix(route, "/")
	route = strings.ReplaceAll(route, "/", ":")
	return fmt.Sprintf("ui:schema:%s:%s:%s:%s:%s:%s:%s:%s",
		tenantID, route, permFP, flagFP,
		v.CompilerVersion, v.ASTVersion, v.PolicyGeneration, v.SchemaGeneration,
	)
}

// TenantPattern returns the Redis glob pattern that matches all cached schemas
// for a given tenant, regardless of generation or permission set.
// Used by tenant-scoped hard invalidation.
func TenantPattern(tenantID string) string {
	return fmt.Sprintf("ui:schema:%s:*", tenantID)
}

// ModulePattern returns the Redis glob pattern matching all schemas for a
// given module prefix under a tenant (e.g. "finance" matches /finance/*).
func ModulePattern(tenantID, modulePrefix string) string {
	modulePrefix = strings.TrimPrefix(modulePrefix, "/")
	modulePrefix = strings.ReplaceAll(modulePrefix, "/", ":")
	return fmt.Sprintf("ui:schema:%s:%s:*", tenantID, modulePrefix)
}
