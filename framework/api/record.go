package api

// record.go — in-memory MutableRecord implementation and factory functions.
//
// The persistence layer returns [def.Record] values (read-only).  HTTP
// handlers need a mutable view to apply body changes, set derived fields, and
// pass the mutated record to hooks and validation.  simpleRecord is that
// mutable view; it is intentionally private to this package.
//
// Two factory functions cover the two mutation cases:
//
//   - [newMutableFromMap]    — create: builds a record from a raw JSON body map.
//   - [newMutableFromRecord] — update: overlays body changes onto an existing record.

import (
	"fmt"

	"github.com/google/uuid"

	"awo.so/framework/def"
)

// ──────────────────────────────────────────────────────────────────────────────
// simpleRecord
// ──────────────────────────────────────────────────────────────────────────────

// simpleRecord is a package-private [def.MutableRecord] backed by a
// plain map.  It is used inside HTTP handlers as a staging area for applying
// body changes and setting server-controlled fields before persistence.
//
// It is not safe for concurrent use.
type simpleRecord struct {
	// id is the entity's primary key.  May be uuid.Nil before the persistence
	// layer assigns one on INSERT.
	id uuid.UUID

	// tenantID is injected from the resolved viewer context; never taken from
	// user-supplied body data.
	tenantID uuid.UUID

	// entityName matches [def.Entitydef.Name] and is used by
	// newMutableFromRecord to look up the entity's field list.
	entityName string

	// data holds all field values, including system fields (org_unit_id,
	// created_at, updated_at) that are mirrored in from the base record on
	// update.
	data map[string]any
}

// Compile-time check: simpleRecord must satisfy MutableRecord.
var _ def.MutableRecord = (*simpleRecord)(nil)

func (r *simpleRecord) Get(field string) any      { return r.data[field] }
func (r *simpleRecord) Set(field string, val any) { r.data[field] = val }
func (r *simpleRecord) ID() uuid.UUID             { return r.id }
func (r *simpleRecord) TenantID() uuid.UUID       { return r.tenantID }
func (r *simpleRecord) EntityName() string        { return r.entityName }

// ──────────────────────────────────────────────────────────────────────────────
// Factory functions
// ──────────────────────────────────────────────────────────────────────────────

// newMutableFromMap constructs a [simpleRecord] from a raw JSON body map for
// use in create operations.
//
// tenantID is always taken from the resolved viewer context — the body value
// (if any) is silently ignored, preventing cross-tenant injection.
//
// If the body contains an "id" key with a valid UUID string, that value is
// used as the record ID (useful when the client generates IDs client-side,
// e.g. for offline-first sync patterns).  Otherwise a new UUID v4 is generated
// so that the record is never inserted with a nil primary key.
func newMutableFromMap(
	entityName string,
	tenantID uuid.UUID,
	body map[string]any,
) def.MutableRecord {
	rec := &simpleRecord{
		entityName: entityName,
		tenantID:   tenantID,
		data:       make(map[string]any, len(body)),
	}

	// Attempt to use a client-supplied ID.
	if raw, ok := body["id"]; ok {
		if s, ok := raw.(string); ok {
			rec.id, _ = uuid.Parse(s)
		}
	}

	// Generate an ID if none was supplied or the supplied value was invalid.
	// This ensures the record always has a valid primary key before reaching
	// the persistence layer, regardless of whether the store auto-generates IDs.
	if rec.id == uuid.Nil {
		rec.id = uuid.New()
	}

	// Copy all body fields.  Caller is responsible for overwriting privileged
	// fields (tenant_id, org_unit_id) after this call.
	for k, v := range body {
		rec.data[k] = v
	}

	return rec
}

// newMutableFromRecord constructs a [simpleRecord] for use in update
// operations by copying all fields from base and then overlaying changes.
//
// def is required to enumerate the entity's declared fields so that the copy
// is complete.  An error is returned if def does not match base.EntityName(),
// indicating a programming error in the caller.
//
// System fields (id, tenant_id, created_at) are preserved from base and
// cannot be overwritten via changes.  This prevents a client from, for example,
// changing the record's tenant or resetting its creation timestamp.
func newMutableFromRecord(
	def *def.EntityDefinition,
	base def.Record,
	changes map[string]any,
) (def.MutableRecord, error) {
	// Guard against mismatched entity names — a symptom of a routing or
	// handler-configuration bug.
	if def.Name != base.EntityName() {
		return nil, fmt.Errorf(
			"api: entity name mismatch in newMutableFromRecord: handler=%q record=%q",
			def.Name, base.EntityName(),
		)
	}

	rec := &simpleRecord{
		id:         base.ID(),
		tenantID:   base.TenantID(),
		entityName: base.EntityName(),
		data:       make(map[string]any, len(def.Fields)+4), // +4 for system fields
	}

	// Mirror system / metadata fields from the existing record.
	// These are intentionally excluded from the changes map below.
	for _, sys := range []string{"created_at", "updated_at", "org_unit_id"} {
		rec.data[sys] = base.Get(sys)
	}

	// Copy declared entity fields from the existing record as the baseline.
	for _, f := range def.Fields {
		rec.data[f.Name] = base.Get(f.Name)
	}

	// Apply user-supplied changes on top of the baseline.
	// System fields (created_at, tenant_id, id) present in changes are
	// silently ignored because they are never in def.Fields.
	for k, v := range changes {
		// Only allow overwriting keys that are already present (i.e. declared
		// fields or the system fields seeded above).  Unknown keys in the body
		// are dropped, preventing mass-assignment of undeclared columns.
		if _, exists := rec.data[k]; exists {
			rec.data[k] = v
		}
	}

	return rec, nil
}
