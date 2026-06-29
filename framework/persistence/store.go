package persistence

import (
	"context"

	"github.com/google/uuid"

	"awo.so/framework/def"
	"awo.so/framework/filter"
)

// ListOptions controls pagination, ordering, and filtering for List queries.
type ListOptions struct {
	// Predicate is a composable filter tree built with the filter package.
	// Takes precedence over all ad-hoc filters.
	// Example: filter.Eq("status", "active").And(filter.Gt("amount", 1000))
	Predicate *filter.Filter

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
	Records []def.Record
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
	FindByID(ctx context.Context, id uuid.UUID) (def.Record, error)

	// List returns a paginated, filtered page of records.
	List(ctx context.Context, opts ListOptions) (Page, error)

	// Create inserts a new record. The record's ID is assigned by the store
	// (uuid_generate_v4()) if not already set.
	Create(ctx context.Context, rec def.MutableRecord) error

	// Update applies changes in rec to the existing row identified by rec.ID().
	Update(ctx context.Context, rec def.MutableRecord) error

	// Delete removes the record. For soft-delete entities, sets deleted_at.
	Delete(ctx context.Context, id uuid.UUID) error

	// Exists reports whether any record matches the given filter.
	Exists(ctx context.Context, filter map[string]any) (bool, error)

	// Count returns the number of records matching filter (AND semantics, exact match).
	// More efficient than List+Total when only the count is needed.
	Count(ctx context.Context, filter map[string]any) (int64, error)

	// BulkCreate inserts multiple records in a single statement. Atomic — all
	// succeed or all fail. Does NOT invoke hooks; caller is responsible for
	// running before_validate, before_save, and after_save outside this call.
	BulkCreate(ctx context.Context, recs []def.MutableRecord) error

	// BulkUpdate applies the same field values to all records matching filter.
	// Does NOT invoke hooks — intended for system-level batch operations only.
	BulkUpdate(ctx context.Context, filter map[string]any, values map[string]any) (int64, error)
}

// ── Generic typed interface ────────────────────────────────────────────────────

// RecordMapper converts a dynamic def.Record into a typed domain value T.
// Module code provides this function when constructing a TypedRepository.
type RecordMapper[T any] func(def.Record) (T, error)

// EntityRepository[T any] is the typed persistence contract for one entity type.
//
// T is the caller's domain type. For dynamic/meta code use T = def.Record
// and adapt via AsRecordRepository(store). Module code passes a RecordMapper to
// NewTypedRepository to obtain a fully typed repository over the pgstore backend.
//
// All methods apply the entity's privacy policies and respect tenant context.
// Write methods do NOT invoke lifecycle hooks — the caller (handler or service)
// is responsible for running the hook chain around repository calls.
type EntityRepository[T any] interface {
	// Get returns the record with the given primary key.
	// Returns ErrNotFound when no row exists (or is soft-deleted).
	Get(ctx context.Context, id uuid.UUID) (T, error)

	// Query returns a paginated, filtered list of records plus the total count.
	Query(ctx context.Context, opts ListOptions) ([]T, int64, error)

	// Exists reports whether any record matches the given exact-match filter.
	Exists(ctx context.Context, filter map[string]any) (bool, error)

	// Count returns the count of records matching the given exact-match filter.
	Count(ctx context.Context, filter map[string]any) (int64, error)

	// Create inserts rec and returns the saved record (with assigned ID, timestamps).
	Create(ctx context.Context, rec def.MutableRecord) (T, error)

	// Update applies changes in rec to the existing row identified by rec.ID()
	// and returns the updated record.
	Update(ctx context.Context, rec def.MutableRecord) (T, error)

	// Delete removes the record (or sets deleted_at for soft-delete entities).
	Delete(ctx context.Context, id uuid.UUID) error

	// BulkCreate inserts multiple records atomically. Hooks do not run.
	BulkCreate(ctx context.Context, recs []def.MutableRecord) ([]T, error)

	// BulkUpdate applies values to all records matching filter. Hooks do not run.
	BulkUpdate(ctx context.Context, filter map[string]any, values map[string]any) (int64, error)
}

// TypedRepository[T] adapts an EntityStore to EntityRepository[T] via a RecordMapper.
// Construct with NewTypedRepository; use AsRecordRepository for def.Record.
type TypedRepository[T any] struct {
	store  EntityStore
	mapper RecordMapper[T]
}

// NewTypedRepository wraps store with mapper to produce an EntityRepository[T].
func NewTypedRepository[T any](store EntityStore, mapper RecordMapper[T]) *TypedRepository[T] {
	return &TypedRepository[T]{store: store, mapper: mapper}
}

// AsRecordRepository wraps store as EntityRepository[def.Record].
// Use when you need the generic interface but do not have a typed domain struct.
func AsRecordRepository(store EntityStore) *TypedRepository[def.Record] {
	return NewTypedRepository(store, func(r def.Record) (def.Record, error) {
		return r, nil
	})
}

func (r *TypedRepository[T]) Get(ctx context.Context, id uuid.UUID) (T, error) {
	rec, err := r.store.FindByID(ctx, id)
	if err != nil {
		var zero T
		return zero, err
	}
	return r.mapper(rec)
}

func (r *TypedRepository[T]) Query(ctx context.Context, opts ListOptions) ([]T, int64, error) {
	page, err := r.store.List(ctx, opts)
	if err != nil {
		return nil, 0, err
	}
	out := make([]T, 0, len(page.Records))
	for _, rec := range page.Records {
		t, err := r.mapper(rec)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, t)
	}
	return out, page.Total, nil
}

func (r *TypedRepository[T]) Exists(ctx context.Context, filter map[string]any) (bool, error) {
	return r.store.Exists(ctx, filter)
}

func (r *TypedRepository[T]) Count(ctx context.Context, filter map[string]any) (int64, error) {
	return r.store.Count(ctx, filter)
}

func (r *TypedRepository[T]) Create(ctx context.Context, rec def.MutableRecord) (T, error) {
	if err := r.store.Create(ctx, rec); err != nil {
		var zero T
		return zero, err
	}
	return r.mapper(rec)
}

func (r *TypedRepository[T]) Update(ctx context.Context, rec def.MutableRecord) (T, error) {
	if err := r.store.Update(ctx, rec); err != nil {
		var zero T
		return zero, err
	}
	return r.mapper(rec)
}

func (r *TypedRepository[T]) Delete(ctx context.Context, id uuid.UUID) error {
	return r.store.Delete(ctx, id)
}

func (r *TypedRepository[T]) BulkCreate(ctx context.Context, recs []def.MutableRecord) ([]T, error) {
	if err := r.store.BulkCreate(ctx, recs); err != nil {
		return nil, err
	}
	out := make([]T, 0, len(recs))
	for _, rec := range recs {
		t, err := r.mapper(rec)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, nil
}

func (r *TypedRepository[T]) BulkUpdate(ctx context.Context, filter map[string]any, values map[string]any) (int64, error) {
	return r.store.BulkUpdate(ctx, filter, values)
}

// compile-time check
var _ EntityRepository[def.Record] = (*TypedRepository[def.Record])(nil)

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
