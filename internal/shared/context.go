package shared

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const (
	TenantIDKey   contextKey = "tenant_id"
	UserIDKey     contextKey = "user_id"
	RequestCtxKey contextKey = "request_context"
)

// WithTenantID adds tenant ID to context
func WithTenantID(ctx context.Context, tenantID uuid.UUID) context.Context {
	return context.WithValue(ctx, TenantIDKey, tenantID)
}

// GetTenantID retrieves tenant ID from context
func GetTenantID(ctx context.Context) (uuid.UUID, bool) {
	tenantID, ok := ctx.Value(TenantIDKey).(uuid.UUID)
	return tenantID, ok
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
