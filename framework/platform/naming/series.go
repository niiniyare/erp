// Package naming provides atomic document-number generation backed by PostgreSQL.
//
// Sequences are stored in the awo_naming_sequences table (one row per entity
// per tenant). Each call to Next atomically increments the counter and returns
// the new value in a single round-trip via INSERT … ON CONFLICT … UPDATE.
//
// Required migration: see db/migrations/000002_framework.up.sql
package naming

import (
	"fmt"

	"github.com/google/uuid"

	"awo.so/framework/definition"
)

// ExecOneRow executes a single-row query and scans the result into dest.
// Callers provide this to decouple the naming package from pgx.
type ExecOneRow func(sql string, args []any, dest []any) error

const upsertSeq = `
INSERT INTO awo_naming_sequences (tenant_id, entity, current_seq)
VALUES ($1, $2, 1)
ON CONFLICT (tenant_id, entity) DO UPDATE
    SET current_seq = awo_naming_sequences.current_seq + 1
RETURNING current_seq`

// Next atomically increments the sequence for (tenantID, entityName) and
// returns the new counter value. exec must be backed by the same transaction
// as the Create INSERT so the counter rolls back on failure.
func Next(exec ExecOneRow, tenantID uuid.UUID, entityName string) (int64, error) {
	var seq int64
	if err := exec(upsertSeq, []any{tenantID, entityName}, []any{&seq}); err != nil {
		return 0, fmt.Errorf("naming.Next %s/%s: %w", entityName, tenantID, err)
	}
	return seq, nil
}

// Format formats seq according to the naming series definition.
// Padding defaults to 5 when ≤ 0.
func Format(ns *definition.NamingSeriesDef, seq int64) string {
	pad := ns.Padding
	if pad <= 0 {
		pad = 5
	}
	return fmt.Sprintf("%s%0*d", ns.Prefix, pad, seq)
}

// Stamp generates the next document number and writes it onto rec's target
// field. No-op when def.NamingSeries is nil.
func Stamp(exec ExecOneRow, tenantID uuid.UUID, def *definition.EntityDefinition, rec definition.MutableRecord) error {
	if def.NamingSeries == nil {
		return nil
	}
	seq, err := Next(exec, tenantID, def.Name)
	if err != nil {
		return err
	}
	rec.Set(def.NamingSeries.Field, Format(def.NamingSeries, seq))
	return nil
}
