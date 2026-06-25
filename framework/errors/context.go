package errors

import (
	"context"

	"github.com/google/uuid"
)

func getTenantIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(TenantIDKey).(uuid.UUID); ok {
		return id.String()
	}
	if id, ok := ctx.Value(TenantIDKey).(string); ok {
		return id
	}
	return ""
}

func getUserIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(UserIDKey).(uuid.UUID); ok {
		return id.String()
	}
	if id, ok := ctx.Value(UserIDKey).(string); ok {
		return id
	}
	return ""
}

func getRequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(RequestIDKey).(string)
	return id
}

func getOperationFromContext(ctx context.Context) string {
	op, _ := ctx.Value(OperationKey).(string)
	return op
}

// Exported context helpers for callers to enrich errors with context.
func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, TenantIDKey, tenantID)
}

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

func WithOperation(ctx context.Context, op string) context.Context {
	return context.WithValue(ctx, OperationKey, op)
}

