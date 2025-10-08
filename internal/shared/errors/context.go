package errors

import (
	"context"

	"github.com/google/uuid"
)

// ─── CONTEXT HELPERS ─────────────────────────────────────────────

func getTenantIDFromContext(ctx context.Context) string {
	if tenantID, ok := ctx.Value(TenantIDKey).(uuid.UUID); ok {
		return tenantID.String()
	}
	if tenantID, ok := ctx.Value(TenantIDKey).(string); ok {
		return tenantID
	}
	return ""
}

func getUserIDFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value(UserIDKey).(uuid.UUID); ok {
		return userID.String()
	}
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		return userID
	}
	return ""
}

func getRequestIDFromContext(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}

func getOperationFromContext(ctx context.Context) string {
	if operation, ok := ctx.Value(OperationKey).(string); ok {
		return operation
	}
	return ""
}
