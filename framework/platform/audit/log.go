// Package audit writes structured audit log entries for Create, Update, and Delete
// operations on Audited entities.
//
// Required migration: see db/migrations/000002_framework.up.sql
package audit

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"awo.so/framework/def"
)

// Entry is one audit log record. tenant_id is populated by the DB via current_tenant_id().
type Entry struct {
	Entity   string         `json:"entity"`
	RecordID uuid.UUID      `json:"record_id"`
	Op       string         `json:"op"`
	ActorID  string         `json:"actor_id"`
	Changes  map[string]any `json:"changes"`
}

// WriteFunc persists an audit entry. The host application provides this,
// typically by writing to awo_audit_log inside the same transaction.
type WriteFunc func(ctx context.Context, e *Entry) error

const insertAudit = `
INSERT INTO awo_audit_log (tenant_id, entity, record_id, op, actor_id, changes)
VALUES (current_tenant_id(), $1, $2, $3, $4, $5)`

// ExecFunc executes a fire-and-forget SQL statement (no result scanning).
// Implementations use the store's transaction querier so audit rows roll back
// on failure.
type ExecFunc func(ctx context.Context, sql string, args []any) error

// Write builds and persists an audit entry for mutation m.
// No-op when def.Audited is false.
func Write(ctx context.Context, exec ExecFunc, entity *def.EntityDefinition, m *def.Mutation) error {
	if !entity.Audited {
		return nil
	}

	recordID := uuid.Nil
	op := m.Op.String()

	changes := diff(entity, m)

	switch m.Op {
	case def.OpCreate:
		if m.After != nil {
			recordID = m.After.ID()
		}
	case def.OpUpdate:
		if m.After != nil {
			recordID = m.After.ID()
		}
	case def.OpDelete:
		if m.Before != nil {
			recordID = m.Before.ID()
		}
	}

	changesJSON, err := json.Marshal(changes)
	if err != nil {
		return fmt.Errorf("audit: marshal changes: %w", err)
	}

	return exec(ctx, insertAudit, []any{entity.Name, recordID, op, m.ActorID, string(changesJSON)})
}

// diff builds a changes map from the mutation.
// For Create: {"after": {field: value, …}}
// For Update: {"before": {…}, "after": {…}} — only changed fields
// For Delete: {"before": {field: value, …}}
func diff(entity *def.EntityDefinition, m *def.Mutation) map[string]any {
	switch m.Op {
	case def.OpCreate:
		return map[string]any{"after": recordSnapshot(entity, m.After)}
	case def.OpDelete:
		return map[string]any{"before": recordSnapshot(entity, m.Before)}
	case def.OpUpdate:
		before := recordSnapshot(entity, m.Before)
		after := recordSnapshot(entity, m.After)
		changed := make(map[string]any)
		for k, av := range after {
			bv := before[k]
			if fmt.Sprintf("%v", av) != fmt.Sprintf("%v", bv) {
				changed[k] = av
			}
		}
		return map[string]any{
			"before": filterKeys(before, changed),
			"after":  changed,
		}
	}
	return map[string]any{}
}

// recordSnapshot captures field values from rec, skipping sensitive fields.
func recordSnapshot(entity *def.EntityDefinition, rec def.Record) map[string]any {
	if rec == nil {
		return map[string]any{}
	}
	m := make(map[string]any, len(entity.Fields))
	for _, f := range entity.Fields {
		if f.IsSensitive {
			continue
		}
		m[f.Name] = rec.Get(f.Name)
	}
	return m
}

// filterKeys returns a copy of src containing only keys present in keys.
func filterKeys(src map[string]any, keys map[string]any) map[string]any {
	out := make(map[string]any, len(keys))
	for k := range keys {
		out[k] = src[k]
	}
	return out
}
