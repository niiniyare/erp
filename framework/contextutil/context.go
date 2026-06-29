// Package contextutil provides context helpers for propagating organisational
// scope and request identity through the Awo Framework call chain.
//
// # Org scope context
//
// Every authenticated request carries an def.Scope that locates the request
// within the organisational tree:
//
//	Tenant ──► OrgUnit (any node in the unit tree)
//
// Use WithOrgScope at the request entry point (middleware or handler) and
// GetOrgScope wherever the scope is needed:
//
//	// In auth middleware:
//	ctx = contextutil.WithOrgScope(ctx, org.WithUnit(tenantID, unitID))
//
//	// In a service or repository:
//	scope, ok := contextutil.GetOrgScope(ctx)
//	if !ok { return ErrUnauthenticated }
//
// # Standard context keys
//
// The package uses unexported typed context keys to avoid collisions with
// third-party packages that use string keys.
package contextutil

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"awo.so/framework/def"
)

// SystemUserID is a well-known sentinel UUID used when the system itself
// initiates an operation (Temporal workflow, scheduled job, internal service).
// It is never a real user account and must never appear in user-facing queries.
//
// Value: 00000000-0000-0000-0000-000000000001
var SystemUserID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

// contextKey is a private type to prevent collisions with other packages
// that use plain strings as context keys.
type contextKey string

const (
	keyOrgScope   contextKey = "awo.org_scope"
	keyActorID    contextKey = "awo.actor_id"
	keyRequestCtx contextKey = "awo.request_ctx"
)

// ── Org scope ─────────────────────────────────────────────────────────────────

// WithOrgScope stores an def.Scope in the context.
// Call this early in the request pipeline (auth middleware) so downstream
// layers (service, repository, privacy policies) can retrieve it.
func WithOrgScope(ctx context.Context, scope def.Scope) context.Context {
	return context.WithValue(ctx, keyOrgScope, scope)
}

// GetOrgScope retrieves the def.Scope from the context.
// Returns (zero, false) if no scope has been set or if TenantID is missing.
func GetOrgScope(ctx context.Context) (def.Scope, bool) {
	scope, ok := ctx.Value(keyOrgScope).(def.Scope)
	if !ok || scope.TenantID == uuid.Nil {
		return def.Scope{}, false
	}
	return scope, true
}

// MustGetOrgScope retrieves the def.Scope or panics.
// Intended for internal framework use where scope absence is a programming error.
func MustGetOrgScope(ctx context.Context) def.Scope {
	scope, ok := GetOrgScope(ctx)
	if !ok {
		panic("contextutil: org scope not in context — ensure auth middleware ran")
	}
	return scope
}

// WithTenantID is a convenience wrapper that sets a tenant-only scope with no
// specific org unit. Use WithOrgScope(ctx, org.WithUnit(…)) for unit-scoped requests.
func WithTenantID(ctx context.Context, tenantID uuid.UUID) context.Context {
	return WithOrgScope(ctx, def.TenantOnly(tenantID))
}

// GetTenantID retrieves just the tenant ID from the org scope.
// Returns (uuid.Nil, false) if no scope is present.
func GetTenantID(ctx context.Context) (uuid.UUID, bool) {
	scope, ok := GetOrgScope(ctx)
	if !ok {
		return uuid.Nil, false
	}
	return scope.TenantID, true
}

// GetOrgUnitID retrieves the org unit ID from the scope.
// Returns (uuid.Nil, false) when no scope is set or the viewer is tenant-wide.
func GetOrgUnitID(ctx context.Context) (uuid.UUID, bool) {
	scope, ok := GetOrgScope(ctx)
	if !ok || scope.UnitID == uuid.Nil {
		return uuid.Nil, false
	}
	return scope.UnitID, true
}

// ── Actor identity ────────────────────────────────────────────────────────────

// WithActorID stores the authenticated user UUID in the context.
func WithActorID(ctx context.Context, actorID uuid.UUID) context.Context {
	return context.WithValue(ctx, keyActorID, actorID)
}

// GetActorID retrieves the authenticated user UUID from the context.
// Returns (uuid.Nil, false) for unauthenticated requests.
func GetActorID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(keyActorID).(uuid.UUID)
	if !ok || id == uuid.Nil {
		return uuid.Nil, false
	}
	return id, true
}

// IsSystemActor reports whether the context actor is the system sentinel UUID.
// Use this to detect framework-internal or workflow-initiated operations.
func IsSystemActor(ctx context.Context) bool {
	id, ok := GetActorID(ctx)
	return ok && id == SystemUserID
}

// ── Request metadata ──────────────────────────────────────────────────────────

// RequestContext holds HTTP-layer metadata about the incoming request.
// It is stored in context by middleware and read by loggers, audit log writers,
// and tracing integrations.
type RequestContext struct {
	// Request is the underlying HTTP request. May be nil in non-HTTP contexts
	// (e.g. Temporal activity, gRPC).
	Request *http.Request

	// UserAgent is the value of the User-Agent header.
	UserAgent string

	// IPAddress is the client's IP address (after proxy header resolution).
	IPAddress string

	// SessionID is the authenticated session identifier, if applicable.
	SessionID string

	// TraceID is the distributed trace ID for correlating logs across services.
	TraceID string

	// RequestID is a per-request UUID injected by the gateway or middleware.
	RequestID string
}

// WithRequestContext stores request metadata in the context.
func WithRequestContext(ctx context.Context, req *RequestContext) context.Context {
	return context.WithValue(ctx, keyRequestCtx, req)
}

// GetRequestContext retrieves the request metadata from the context.
// Returns (nil, false) when no request context has been set.
func GetRequestContext(ctx context.Context) (*RequestContext, bool) {
	req, ok := ctx.Value(keyRequestCtx).(*RequestContext)
	return req, ok
}

// GetTraceID is a convenience function that extracts the trace ID from the
// request context without requiring callers to unpack the full RequestContext.
func GetTraceID(ctx context.Context) string {
	if req, ok := GetRequestContext(ctx); ok {
		return req.TraceID
	}
	return ""
}

// GetRequestID extracts the request ID from the request context.
func GetRequestID(ctx context.Context) string {
	if req, ok := GetRequestContext(ctx); ok {
		return req.RequestID
	}
	return ""
}
