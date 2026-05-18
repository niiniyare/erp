package shared

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// SystemUserID is a well-known sentinel UUID used when the system itself
// performs an action on behalf of an automated workflow (e.g. Temporal SLA
// expiry, scheduled job). It is never a real user account.
// Value: 00000000-0000-0000-0000-000000000001
var SystemUserID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

type contextKey string

const (
	TenantIDKey          contextKey = "tenant_id"
	UserIDKey            contextKey = "user_id"
	EntityIDKey          contextKey = "entity_id"
	RequestCtxKey        contextKey = "request_context"
	CapabilityContextKey contextKey = "capability_context"
)

// WithTenantID adds tenant ID to context
func WithTenantID(ctx context.Context, tenantID uuid.UUID) context.Context {
	return context.WithValue(ctx, TenantIDKey, tenantID)
}

// GetTenantID retrieves tenant ID from context
func GetTenantID(ctx context.Context) (uuid.UUID, bool) {
	tenantID, ok := ctx.Value(TenantIDKey).(uuid.UUID)
	if !ok || tenantID == uuid.Nil {
		return uuid.Nil, false
	}
	return tenantID, true
}

// WithEntityID adds entity (legal entity / business unit) ID to context.
// Entity scoping is used to prevent cross-entity data leaks within a tenant.
func WithEntityID(ctx context.Context, entityID uuid.UUID) context.Context {
	return context.WithValue(ctx, EntityIDKey, entityID)
}

// GetEntityID retrieves entity ID from context.
// Returns (uuid.Nil, false) when no entity has been set — callers should treat
// this as "no entity scoping" (tenant-wide) rather than an error.
func GetEntityID(ctx context.Context) (uuid.UUID, bool) {
	entityID, ok := ctx.Value(EntityIDKey).(uuid.UUID)
	if !ok || entityID == uuid.Nil {
		return uuid.Nil, false
	}
	return entityID, true
}

// GetEntityIDPtr returns a pointer to the entity ID from context, or nil if absent.
// Useful for repository calls that accept *uuid.UUID for optional entity scoping.
func GetEntityIDPtr(ctx context.Context) *uuid.UUID {
	entityID, ok := GetEntityID(ctx)
	if !ok {
		return nil
	}
	return &entityID
}

// WithUserID adds user ID to context
func WithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// GetUserID retrieves user ID from context
func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return userID, ok
}

// GetUserIDPtr retrieves user ID from context
func GetUserIDPtr(ctx context.Context) *uuid.UUID {
	userID, _ := ctx.Value(UserIDKey).(uuid.UUID)
	return &userID
}

// RequestContext holds request-specific information
type RequestContext struct {
	Request   *http.Request
	UserAgent string
	IPAddress string
	SessionID string
	TraceID   string
}

// WithRequestContext adds request context to context
func WithRequestContext(ctx context.Context, reqCtx *RequestContext) context.Context {
	return context.WithValue(ctx, RequestCtxKey, reqCtx)
}

// GetRequestContext retrieves request context from context
func GetRequestContext(ctx context.Context) (*RequestContext, bool) {
	reqCtx, ok := ctx.Value(RequestCtxKey).(*RequestContext)
	return reqCtx, ok
}

// CapabilityContext carries non-authorization context about the request actor.
// Permissions, features, and module flags have been removed — use
// contract.SessionContext.FeatureEnabled / middleware.Authorize instead.
type CapabilityContext struct {
	UserID     string
	Role       string
	Scope      string
	TenantID   *string
	CustomerID *string
}

// WithCapabilityContext stores a CapabilityContext in the context.
func WithCapabilityContext(ctx context.Context, cap CapabilityContext) context.Context {
	return context.WithValue(ctx, CapabilityContextKey, cap)
}

// GetCapabilityContext retrieves a CapabilityContext from the context.
func GetCapabilityContext(ctx context.Context) (CapabilityContext, bool) {
	cap, ok := ctx.Value(CapabilityContextKey).(CapabilityContext)
	return cap, ok
}
