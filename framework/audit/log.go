// Package audit writes structured audit log entries for Create, Update, and Delete
// operations on Audited entities.
//
// Required migration:
//
//	CREATE TABLE IF NOT EXISTS awo_audit_log (
//	    id         UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
//	    tenant_id  UUID,
//	    entity     TEXT        NOT NULL,
//	    record_id  UUID        NOT NULL,
//	    op         TEXT        NOT NULL,  -- "create" | "update" | "delete"
//	    actor_id   TEXT        NOT NULL,
//	    changes    JSONB       NOT NULL DEFAULT '{}',
//	    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
//	);
//	CREATE INDEX IF NOT EXISTS awo_audit_log_record_idx
//	    ON awo_audit_log (entity, record_id);
//	CREATE INDEX IF NOT EXISTS awo_audit_log_tenant_idx
//	    ON awo_audit_log (tenant_id, created_at DESC);
package audit

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"awo.so/framework/definition"
)

// Entry is one audit log record.
type Entry struct {
	TenantID uuid.UUID      `json:"tenant_id"`
	Entity   string         `json:"entity"`
	RecordID uuid.UUID      `json:"record_id"`
	Op       string         `json:"op"`
	ActorID  string         `json:"actor_id"`
	Changes  map[string]any `json:"changes"`
}

// WriteFunc persists an audit entry. The host application provides this,
// typically by writing to awo_audit_log inside the same transaction.
//
//	func(ctx context.Context, e *audit.Entry) error {
//	    _, err := tx.Exec(ctx,
//	        `INSERT INTO awo_audit_log (tenant_id,entity,record_id,op,actor_id,changes)
//	         VALUES ($1,$2,$3,$4,$5,$6)`,
//	        e.TenantID, e.Entity, e.RecordID, e.Op, e.ActorID, changesJSON)
//	    return err
//	}
type WriteFunc func(ctx context.Context, e *Entry) error

const insertAudit = `
INSERT INTO awo_audit_log (tenant_id, entity, record_id, op, actor_id, changes)
VALUES ($1, $2, $3, $4, $5, $6)`

// ExecFunc executes a fire-and-forget SQL statement (no result scanning).
// Implementations use the store's transaction querier so audit rows roll back
// on failure.
type ExecFunc func(ctx context.Context, sql string, args []any) error

// Write builds and persists an audit entry for mutation m.
// No-op when def.Audited is false.
func Write(ctx context.Context, exec ExecFunc, def *definition.EntityDefinition, m *definition.Mutation) error {
	if !def.Audited {
		return nil
	}

	tenantID, _ := uuid.Parse(m.TenantID)
	recordID := uuid.Nil
	op := string(m.Op)

	changes := diff(def, m)

	switch m.Op {
	case definition.OpCreate:
		if m.After != nil {
			recordID = m.After.ID()
		}
	case definition.OpUpdate:
		if m.After != nil {
			recordID = m.After.ID()
		}
	case definition.OpDelete:
		if m.Before != nil {
			recordID = m.Before.ID()
		}
	}

	changesJSON, err := json.Marshal(changes)
	if err != nil {
		return fmt.Errorf("audit: marshal changes: %w", err)
	}

	return exec(ctx, insertAudit, []any{tenantID, def.Name, recordID, op, m.ActorID, string(changesJSON)})
}

// diff builds a changes map from the mutation.
// For Create: {"after": {field: value, …}}
// For Update: {"before": {…}, "after": {…}} — only changed fields
// For Delete: {"before": {field: value, …}}
func diff(def *definition.EntityDefinition, m *definition.Mutation) map[string]any {
	switch m.Op {
	case definition.OpCreate:
		return map[string]any{"after": recordSnapshot(def, m.After)}
	case definition.OpDelete:
		return map[string]any{"before": recordSnapshot(def, m.Before)}
	case definition.OpUpdate:
		before := recordSnapshot(def, m.Before)
		after := recordSnapshot(def, m.After)
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
func recordSnapshot(def *definition.EntityDefinition, rec definition.Record) map[string]any {
	if rec == nil {
		return map[string]any{}
	}
	m := make(map[string]any, len(def.Fields))
	for _, f := range def.Fields {
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
