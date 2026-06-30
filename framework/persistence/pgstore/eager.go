package pgstore

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"awo.so/framework/def"
	"awo.so/framework/filter"
	"awo.so/framework/persistence"
)

// loadEagerEdges fetches all EagerLoad edges for rec and embeds them into the
// returned map under the edge name.
//
// Supported cardinalities:
//   - CardinalityOne  — FK on this entity; fetches the single related row.
//   - CardinalityMany — FK on the target entity; fetches all matching rows.
//   - CardinalityManyToMany — not yet implemented; skipped with a TODO.
//
// Non-fatal errors (unknown target entity, missing FK value) are embedded as
// {"error": "…"} in the edge slot rather than failing the whole request.
// This preserves the primary record while surfacing the edge failure visibly.
func (s *EntityStore) loadEagerEdges(ctx context.Context, rec def.Record, base map[string]any) map[string]any {
	if len(s.def.Edges) == 0 {
		return base
	}

	for _, edge := range s.def.Edges {
		if !edge.EagerLoad {
			continue
		}

		switch edge.Cardinality {
		case def.CardinalityOne:
			base[edge.Name] = s.loadOne(ctx, rec, edge)

		case def.CardinalityMany:
			base[edge.Name] = s.loadMany(ctx, rec, edge)

		case def.CardinalityManyToMany:
			// TODO: implement many-to-many eager load via join table.
			base[edge.Name] = map[string]any{"error": "many_to_many eager load not yet implemented"}
		}
	}
	return base
}

// loadOne fetches a single related record for a CardinalityOne edge.
// FK lives on this entity's record; its value is rec.Get(fkColumn).
func (s *EntityStore) loadOne(ctx context.Context, rec def.Record, edge *def.EdgeDef) any {
	fkCol := edge.ForeignKey
	if fkCol == "" {
		fkCol = edge.Name + "_id"
	}

	fkRaw := rec.Get(fkCol)
	if fkRaw == nil {
		return nil
	}
	relatedID, err := toUUID(fkRaw)
	if err != nil {
		return map[string]any{"error": fmt.Sprintf("edge %q: invalid FK value %v", edge.Name, fkRaw)}
	}

	targetDef := def.Lookup(edge.TargetEntity)
	if targetDef == nil {
		return map[string]any{"error": fmt.Sprintf("edge %q: unknown entity %q", edge.Name, edge.TargetEntity)}
	}

	targetStore := New(targetDef, s.tenantID, s.q)
	related, err := targetStore.FindByID(ctx, relatedID)
	if err != nil {
		if err == persistence.ErrNotFound {
			return nil
		}
		return map[string]any{"error": fmt.Sprintf("edge %q: %v", edge.Name, err)}
	}
	return recordToFlatMap(related, targetDef)
}

// loadMany fetches all related records for a CardinalityMany edge.
// FK lives on the target entity's table; filter: WHERE {fkCol} = thisRecord.ID().
func (s *EntityStore) loadMany(ctx context.Context, rec def.Record, edge *def.EdgeDef) any {
	fkCol := edge.ForeignKey
	if fkCol == "" {
		fkCol = s.def.Name + "_id"
	}

	targetDef := def.Lookup(edge.TargetEntity)
	if targetDef == nil {
		return map[string]any{"error": fmt.Sprintf("edge %q: unknown entity %q", edge.Name, edge.TargetEntity)}
	}

	targetStore := New(targetDef, s.tenantID, s.q)
	page, err := targetStore.List(ctx, persistence.ListOptions{
		Predicate: filter.Eq(fkCol, rec.ID()),
		Limit:     500, // safety cap; deep pagination not supported for eager loads
	})
	if err != nil {
		return map[string]any{"error": fmt.Sprintf("edge %q: %v", edge.Name, err)}
	}

	rows := make([]map[string]any, len(page.Records))
	for i, r := range page.Records {
		rows[i] = recordToFlatMap(r, targetDef)
	}
	return rows
}

// recordToFlatMap converts a def.Record to a map[string]any for edge embedding.
// Mirrors api.recordToMap but lives in pgstore to avoid an import cycle.
func recordToFlatMap(rec def.Record, entDef *def.EntityDefinition) map[string]any {
	m := map[string]any{
		"id":         rec.ID(),
		"created_at": rec.Get("created_at"),
		"updated_at": rec.Get("updated_at"),
	}
	if !entDef.IsGlobal() {
		m["tenant_id"] = rec.TenantID()
	}
	for _, f := range entDef.Fields {
		if !f.IsSensitive {
			m[f.Name] = rec.Get(f.Name)
		}
	}
	return m
}

// toUUID coerces a raw value to uuid.UUID.
func toUUID(v any) (uuid.UUID, error) {
	switch val := v.(type) {
	case uuid.UUID:
		return val, nil
	case string:
		return uuid.Parse(val)
	case []byte:
		return uuid.ParseBytes(val)
	default:
		return uuid.Nil, fmt.Errorf("cannot convert %T to uuid.UUID", v)
	}
}
