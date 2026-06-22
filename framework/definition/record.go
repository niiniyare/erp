package definition

import (
	"context"

	"github.com/google/uuid"
)

// Record provides read-only access to the assembled field values of an entity
// record. It is passed to FieldValidators and HookFuncs so they can inspect
// sibling fields without triggering another DB round-trip.
type Record interface {
	// Get returns the value of the named field, or nil if absent.
	Get(field string) any

	// ID returns the primary key of the record (zero UUID on create before save).
	ID() uuid.UUID

	// TenantID returns the owning tenant (zero UUID for global entities).
	TenantID() uuid.UUID

	// EntityName returns the canonical entity name this record belongs to.
	EntityName() string
}

// ReadStore is the minimal read-only data-access interface available to
// AsyncFieldValidators. It allows uniqueness checks and existence lookups
// without exposing mutating operations.
//
// The concrete implementation is provided by the host application and injected
// at registration time; the framework never directly imports app code.
type ReadStore interface {
	// Exists reports whether any record of the given entity matches the filter.
	// filter is a map of field name → exact value (AND semantics).
	Exists(ctx context.Context, entity string, filter map[string]any) (bool, error)

	// FindByID returns the record for the given entity and primary key.
	// Returns ErrNotFound when no row exists.
	FindByID(ctx context.Context, entity string, id uuid.UUID) (Record, error)
}
