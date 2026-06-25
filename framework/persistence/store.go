package persistence

import (
	"context"

	"github.com/google/uuid"
	"awo.so/framework/definition"
)

// ListOptions controls pagination, ordering, and filtering for List queries.
type ListOptions struct {
	// Filter is a map of field name → value (AND semantics, exact match).
	Filter map[string]any

	// Search is a full-text search string applied to Searchable fields via pg_trgm.
	Search string

	// OrderBy is the field name to sort by. Defaults to "created_at DESC".
	OrderBy string

	// Ascending reverses the default descending sort.
	Ascending bool

	// Limit caps the number of returned records (default 50, max 500).
	Limit int

	// Offset is the zero-based row offset for pagination.
	Offset int

	// OrgUnitIDs restricts results to records whose org_unit_id is in this set.
	// When non-empty the store emits: WHERE org_unit_id = ANY($n).
	// Set by the API handler for ScopeLevelUnit entities based on the viewer's
	// subtree (all descendants of viewer.OrgUnitID()). Ignored for non-unit-scoped
	// entities.
	OrgUnitIDs []uuid.UUID
}

// Page is a paginated result from EntityStore.List.
type Page struct {
	Records []definition.Record
	Total   int64
	Limit   int
	Offset  int
}

// EntityStore is the persistence interface for a single EntityDefinition.
// One EntityStore instance corresponds to one entity type.
//
// The host application provides the concrete implementation (typically backed
// by pgx) and registers it alongside the EntityDefinition at startup.
// The framework never imports app-internal packages.
type EntityStore interface {
	// FindByID returns the record with the given primary key.
	// Returns ErrNotFound when no row exists (or has been soft-deleted).
	FindByID(ctx context.Context, id uuid.UUID) (definition.Record, error)

	// List returns a paginated, filtered page of records.
	List(ctx context.Context, opts ListOptions) (Page, error)

	// Create inserts a new record. The record's ID is assigned by the store
	// (uuid_generate_v4()) if not already set.
	Create(ctx context.Context, rec definition.MutableRecord) error

	// Update applies changes in rec to the existing row identified by rec.ID().
	Update(ctx context.Context, rec definition.MutableRecord) error

	// Delete removes the record. For soft-delete entities, sets deleted_at.
	Delete(ctx context.Context, id uuid.UUID) error

	// Exists reports whether any record matches the given filter.
	Exists(ctx context.Context, filter map[string]any) (bool, error)

	// BulkUpdate applies the same field values to all records matching filter.
	// Does NOT invoke hooks — intended for system-level batch operations only.
	BulkUpdate(ctx context.Context, filter map[string]any, values map[string]any) (int64, error)
}

// TenantStore is the top-level store factory. The host application provides
// one implementation (usually wrapping a *pgxpool.Pool or db.SQLStore) and
// registers it with the framework at startup.
type TenantStore interface {
	// ForEntity returns an EntityStore scoped to the given tenant and entity.
	// The implementation must call set_tenant_context(tenantID) before any query.
	ForEntity(ctx context.Context, tenantID uuid.UUID, entity string) (EntityStore, error)

	// WithTx executes fn within a serialisable transaction, scoped to tenantID.
	// The EntityStore passed to fn shares the transaction connection.
	WithTx(ctx context.Context, tenantID uuid.UUID, fn func(tx TenantTx) error) error
}

// TenantTx is the transactional variant of EntityStore access.
// Obtained from TenantStore.WithTx; commits on nil return, rolls back on error.
type TenantTx interface {
	ForEntity(entity string) EntityStore
}
