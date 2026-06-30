// Package naming provides atomic document-number generation backed by PostgreSQL.
//
// Sequences are stored in the awo_naming_sequences table (one row per entity
// per tenant per period key). Each call to Next atomically increments the counter
// and returns the new value in a single round-trip via INSERT … ON CONFLICT … UPDATE.
//
// Required migration: see db/migrations/000002_framework.up.sql
package naming

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"awo.so/framework/def"
)

// ExecOneRow executes a single-row query and scans the result into dest.
// Callers provide this to decouple the naming package from pgx.
type ExecOneRow func(sql string, args []any, dest []any) error

const upsertSeq = `
INSERT INTO awo_naming_sequences (tenant_id, entity, period_key, current_seq)
VALUES ($1, $2, $3, 1)
ON CONFLICT (tenant_id, entity, period_key) DO UPDATE
    SET current_seq = awo_naming_sequences.current_seq + 1
RETURNING current_seq`

// Next atomically increments the sequence for (tenantID, entityName, periodKey) and
// returns the new counter value. exec must be backed by the same transaction
// as the Create INSERT so the counter rolls back on failure.
//
// periodKey encodes the reset boundary (e.g. "2025" for yearly, "2025-06" for monthly,
// "" for never). Rows with different periodKey values are independent sequences.
func Next(exec ExecOneRow, tenantID uuid.UUID, entityName, periodKey string) (int64, error) {
	var seq int64
	if err := exec(upsertSeq, []any{tenantID, entityName, periodKey}, []any{&seq}); err != nil {
		return 0, fmt.Errorf("naming.Next %s/%s[%s]: %w", entityName, tenantID, periodKey, err)
	}
	return seq, nil
}

// PeriodKey returns the period key for ns at the given time.
// Returns "" (no reset), "2025" (yearly), or "2025-06" (monthly).
func PeriodKey(ns *def.NamingSeriesDef, t time.Time) string {
	switch ns.ResetPeriod {
	case def.ResetYearly:
		return fmt.Sprintf("%04d", t.Year())
	case def.ResetMonthly:
		return fmt.Sprintf("%04d-%02d", t.Year(), t.Month())
	default:
		return ""
	}
}

// ExpandPrefix replaces date tokens in ns.Prefix with values derived from t.
//
// Supported tokens:
//
//	{YYYY} → 4-digit year
//	{YY}   → 2-digit year
//	{MM}   → 2-digit month
//	{DD}   → 2-digit day
func ExpandPrefix(ns *def.NamingSeriesDef, t time.Time) string {
	p := ns.Prefix
	p = strings.ReplaceAll(p, "{YYYY}", fmt.Sprintf("%04d", t.Year()))
	p = strings.ReplaceAll(p, "{YY}", fmt.Sprintf("%02d", t.Year()%100))
	p = strings.ReplaceAll(p, "{MM}", fmt.Sprintf("%02d", int(t.Month())))
	p = strings.ReplaceAll(p, "{DD}", fmt.Sprintf("%02d", t.Day()))
	return p
}

// Format formats seq according to ns using the expanded prefix at time t.
// Padding defaults to 5 when ≤ 0.
func Format(ns *def.NamingSeriesDef, seq int64, t time.Time) string {
	pad := ns.Padding
	if pad <= 0 {
		pad = 5
	}
	return fmt.Sprintf("%s%0*d", ExpandPrefix(ns, t), pad, seq)
}

// Stamp generates the next document number and writes it onto rec's target field.
// No-op when def.NamingSeries is nil.
// now is injected so callers can control the timestamp (use time.Now().UTC() in production).
func Stamp(exec ExecOneRow, tenantID uuid.UUID, entDef *def.EntityDefinition, rec def.MutableRecord, now time.Time) error {
	if entDef.NamingSeries == nil {
		return nil
	}
	ns := entDef.NamingSeries
	periodKey := PeriodKey(ns, now)
	seq, err := Next(exec, tenantID, entDef.Name, periodKey)
	if err != nil {
		return err
	}
	rec.Set(ns.Field, Format(ns, seq, now))
	return nil
}
