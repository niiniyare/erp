package errors

import (
	"context"

	"github.com/google/uuid"
)

// contextKey is a private type for context keys defined by this package,
// following the standard library's own convention (see context.WithValue
// docs) to guarantee these keys can never collide with a key defined by
// another package using a plain string or int.
type contextKey string

const (
	TenantIDKey  contextKey = "tenant_id"
	EntityIDKey  contextKey = "entity_id"
	UserIDKey    contextKey = "user_id"
	RequestIDKey contextKey = "request_id"
	OperationKey contextKey = "operation"
)

// Each getXFromContext helper accepts either a uuid.UUID or a string
// under the given key, since different layers of the codebase store
// context values slightly differently (middleware setting a parsed
// uuid.UUID vs. a handler setting a raw string from a header) — this
// keeps NewFromContext robust to either without callers needing to know
// which representation is in play at a given point in the request path.

func getTenantIDFromContext(ctx context.Context) string {
	return getIDFromContext(ctx, TenantIDKey)
}

func getEntityIDFromContext(ctx context.Context) string {
	return getIDFromContext(ctx, EntityIDKey)
}

func getUserIDFromContext(ctx context.Context) string {
	return getIDFromContext(ctx, UserIDKey)
}

func getIDFromContext(ctx context.Context, key contextKey) string {
	if id, ok := ctx.Value(key).(uuid.UUID); ok {
		return id.String()
	}
	if id, ok := ctx.Value(key).(string); ok {
		return id
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

// NewFromContext creates an AwoError for code and automatically attaches
// whatever tenant/user/entity/request context is available on ctx. This
// is the recommended constructor for handler/service code that already
// has a context.Context in scope, since it saves repeating
// `.WithTenant(...).WithUser(...)` at every call site:
//
//	return nil, errors.NewFromContext(ctx, errors.CodeUserNotFound)
//
// If an entity ID is present on the context, it is attached as a detail
// (there is no dedicated AwoError field for it, since entity is a
// business concept specific to the org-hierarchy module rather than a
// framework-wide concern like tenant/user).
func NewFromContext(ctx context.Context, code Code) *AwoError {
	err := New(code)

	if tenantID := getTenantIDFromContext(ctx); tenantID != "" {
		err = err.WithTenant(tenantID)
	}
	if userID := getUserIDFromContext(ctx); userID != "" {
		err = err.WithUser(userID)
	}
	if entityID := getEntityIDFromContext(ctx); entityID != "" {
		err = err.WithDetail("entity_id", entityID)
	}

	return err
}

// WrapFromContext behaves like NewFromContext but also attaches cause,
// exactly mirroring the relationship between New and Wrap.
func WrapFromContext(ctx context.Context, code Code, cause error) *AwoError {
	return NewFromContext(ctx, code).WithCause(cause)
}
