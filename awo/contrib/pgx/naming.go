// Package pgx provides the PostgreSQL implementation of awo/driver interfaces.
package pgx

import "strings"

// tableFor returns the database table name for an entity.
// Convention: entity name is already in snake_case module_noun format.
// The table name equals the entity name. No transformation needed.
func tableFor(entityName string) string {
	return entityName
}

// systemColumns are the mandatory columns on every system entity table.
// Custom entities store all non-meta fields in the "data" jsonb column.
var systemColumns = []string{
	"id",
	"tenant_id",
	"created_at",
	"updated_at",
}

// isReservedColumn reports whether a column name is a framework-managed
// system column that module authors must not declare as a FieldDef.
func isReservedColumn(name string) bool {
	switch strings.ToLower(name) {
	case "id", "tenant_id", "created_at", "updated_at", "custom_fields":
		return true
	}
	return false
}
