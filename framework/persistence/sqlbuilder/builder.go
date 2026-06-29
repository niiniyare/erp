// Package sqlbuilder generates parameterised PostgreSQL statements from
// EntityDefinition metadata. All methods are pure functions; no DB calls occur here.
package sqlbuilder

import (
	"fmt"
	"strings"

	"github.com/google/uuid"

	"awo.so/framework/definition"
	"awo.so/framework/filter"
)

// SelectOne builds a SELECT statement that fetches a single record by primary key.
//
//	SELECT <cols> FROM <table> WHERE id = $1 [AND deleted_at IS NULL]
func SelectOne(def *definition.EntityDefinition) string {
	cols := columnList(def)
	q := fmt.Sprintf("SELECT %s FROM %s WHERE id = $1", cols, def.TableName())
	if def.SoftDelete {
		q += " AND deleted_at IS NULL"
	}
	return q
}

// SelectOpts controls filtering, search, ordering, and org-unit scoping for SelectList.
type SelectOpts struct {
	// Predicate is an optional composable filter tree (preferred over legacy Filter).
	Predicate  *filter.Filter
	Search     string
	OrderBy    string
	Ascending  bool
	// OrgUnitIDs restricts results to these org unit IDs via ANY($n).
	// Empty slice = no restriction.
	OrgUnitIDs []uuid.UUID
}

// SelectList builds a SELECT + COUNT(*) OVER() query with optional WHERE clause.
//
// Returns (query, filterArgs) where filterArgs holds all bound values EXCEPT
// limit and offset, which the caller appends:
//
//	query, args := sqlbuilder.SelectList(def, opts)
//	args = append(args, limit, offset)
//	rows, err := q.Query(ctx, query, args...)
func SelectList(def *definition.EntityDefinition, opts SelectOpts) (string, []any) {
	cols := columnList(def)
	var sb strings.Builder
	var args []any
	idx := 1

	sb.WriteString(fmt.Sprintf("SELECT %s, COUNT(*) OVER() AS __total FROM %s", cols, def.TableName()))

	clauses, args, idx := whereClauses(def, opts, idx)
	if len(clauses) > 0 {
		sb.WriteString(" WHERE ")
		sb.WriteString(strings.Join(clauses, " AND "))
	}

	sb.WriteString(" ORDER BY ")
	if opts.OrderBy != "" {
		sb.WriteString(pgIdent(opts.OrderBy))
	} else {
		sb.WriteString("created_at")
	}
	if opts.Ascending {
		sb.WriteString(" ASC")
	} else {
		sb.WriteString(" DESC")
	}

	sb.WriteString(fmt.Sprintf(" LIMIT $%d OFFSET $%d", idx, idx+1))

	return sb.String(), args
}

// Insert builds an INSERT statement for the given mutable field names.
// Returns (query, orderedColumns) so the caller can extract values in the
// correct placeholder order.
//
//	INSERT INTO <table> (col1, col2, ...) VALUES ($1, $2, ...) RETURNING id
func Insert(def *definition.EntityDefinition, fields []string) (query string, cols []string) {
	// Start with scope columns; trim based on entity's OrgScope.
	scopeCols := scopeColumns(def)
	reserved := map[string]bool{"id": true}
	for _, sc := range scopeCols {
		reserved[sc] = true
	}

	cols = make([]string, 0, len(fields)+len(scopeCols)+1)
	cols = append(cols, "id")
	cols = append(cols, scopeCols...)
	for _, f := range fields {
		if !reserved[f] {
			cols = append(cols, f)
		}
	}

	placeholders := make([]string, len(cols))
	for i := range cols {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query = fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s) RETURNING id",
		def.TableName(),
		strings.Join(cols, ", "),
		strings.Join(placeholders, ", "),
	)
	return query, cols
}

// immutableCols are columns that must never be changed after INSERT.
var immutableCols = map[string]bool{
	"id":          true,
	"tenant_id":   true,
	"org_unit_id": true, // reparenting goes through a dedicated reparent operation
	"created_at":  true,
}

// Update builds an UPDATE statement for the given mutable field names.
// Returns (query, orderedColumns). The last placeholder is always `id`.
//
//	UPDATE <table> SET col1=$1, col2=$2 WHERE id=$N
func Update(def *definition.EntityDefinition, fields []string) (query string, cols []string) {
	cols = make([]string, 0, len(fields))
	for _, f := range fields {
		if !immutableCols[f] {
			cols = append(cols, f)
		}
	}

	setClauses := make([]string, len(cols))
	for i, col := range cols {
		setClauses[i] = fmt.Sprintf("%s = $%d", pgIdent(col), i+1)
	}

	idPlaceholder := len(cols) + 1
	where := fmt.Sprintf("id = $%d", idPlaceholder)
	if def.SoftDelete {
		where += " AND deleted_at IS NULL"
	}
	query = fmt.Sprintf(
		"UPDATE %s SET %s WHERE %s",
		def.TableName(),
		strings.Join(setClauses, ", "),
		where,
	)
	return query, cols
}

// SoftDelete builds an UPDATE that sets deleted_at = NOW().
func SoftDelete(def *definition.EntityDefinition) string {
	return fmt.Sprintf(
		"UPDATE %s SET deleted_at = NOW() WHERE id = $1",
		def.TableName(),
	)
}

// HardDelete builds a DELETE statement.
func HardDelete(def *definition.EntityDefinition) string {
	return fmt.Sprintf("DELETE FROM %s WHERE id = $1", def.TableName())
}

// BulkInsert builds a multi-row INSERT statement for n records.
// Returns (query, orderedColumns) so the caller can extract values in the
// correct placeholder order (row-major: all cols for row 0, then row 1, …).
//
//	INSERT INTO <table> (col1, col2, …) VALUES ($1,$2,…),($3,$4,…) RETURNING id
func BulkInsert(def *definition.EntityDefinition, fields []string, n int) (query string, cols []string) {
	if n == 0 {
		return "", nil
	}
	scopeCols := scopeColumns(def)
	reserved := map[string]bool{"id": true}
	for _, sc := range scopeCols {
		reserved[sc] = true
	}
	cols = make([]string, 0, len(fields)+len(scopeCols)+1)
	cols = append(cols, "id")
	cols = append(cols, scopeCols...)
	for _, f := range fields {
		if !reserved[f] {
			cols = append(cols, f)
		}
	}

	rowWidth := len(cols)
	valueGroups := make([]string, n)
	for row := 0; row < n; row++ {
		placeholders := make([]string, rowWidth)
		for col := 0; col < rowWidth; col++ {
			placeholders[col] = fmt.Sprintf("$%d", row*rowWidth+col+1)
		}
		valueGroups[row] = "(" + strings.Join(placeholders, ", ") + ")"
	}
	query = fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s RETURNING id",
		def.TableName(),
		strings.Join(cols, ", "),
		strings.Join(valueGroups, ", "),
	)
	return query, cols
}

// Count builds a SELECT COUNT(*) query with optional exact-match WHERE clauses.
// Placeholder numbering starts at 1; caller appends filter values in field order.
func Count(def *definition.EntityDefinition, filterFields []string) string {
	clauses := make([]string, 0, len(filterFields)+1)
	if def.SoftDelete {
		clauses = append(clauses, "deleted_at IS NULL")
	}
	for i, f := range filterFields {
		clauses = append(clauses, fmt.Sprintf("%s = $%d", pgIdent(f), i+1))
	}
	q := fmt.Sprintf("SELECT COUNT(*) FROM %s", def.TableName())
	if len(clauses) > 0 {
		q += " WHERE " + strings.Join(clauses, " AND ")
	}
	return q
}

// Exists builds a cheap existence check.
func Exists(def *definition.EntityDefinition, filterFields []string) string {
	clauses := make([]string, len(filterFields))
	for i, f := range filterFields {
		clauses[i] = fmt.Sprintf("%s = $%d", pgIdent(f), i+1)
	}
	where := strings.Join(clauses, " AND ")
	if def.SoftDelete {
		where += " AND deleted_at IS NULL"
	}
	return fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE %s)", def.TableName(), where)
}

// ──────────────────────────────────────────────────────────────────
// Internal helpers
// ──────────────────────────────────────────────────────────────────

// columnList returns a comma-separated list of all field columns for SELECT.
func columnList(def *definition.EntityDefinition) string {
	cols := make([]string, 0, len(def.Fields)+5)
	cols = append(cols, "id")
	cols = append(cols, scopeColumns(def)...)
	cols = append(cols, "created_at", "updated_at")
	if def.SoftDelete {
		cols = append(cols, "deleted_at")
	}
	for _, f := range def.Fields {
		cols = append(cols, pgIdent(f.Name))
	}
	return strings.Join(cols, ", ")
}

// scopeColumns returns the scope-specific column names for an entity,
// based on its OrgScope level.
//
//   - Global   → [] (no scope columns)
//   - Tenant   → ["tenant_id"]
//   - Unit     → ["tenant_id", "org_unit_id"]
func scopeColumns(def *definition.EntityDefinition) []string {
	switch {
	case def.IsGlobal():
		return nil
	case def.IsUnitScoped():
		return []string{"tenant_id", "org_unit_id"}
	default: // ScopeLevelTenant
		return []string{"tenant_id"}
	}
}

// whereClauses builds AND clauses for filters, org-unit scoping, search, and
// soft-delete.
//
// Placeholder order:
//  1. soft-delete guard (no placeholder)
//  2. Predicate filter tree
//  3. org_unit_id = ANY($n) (one placeholder, []uuid.UUID)
//  4. search ILIKE (one placeholder per searchable field — OR-grouped)
//  5. LIMIT / OFFSET appended by SelectList
//
// Returns (clauses, args, nextIdx).
func whereClauses(def *definition.EntityDefinition, opts SelectOpts, startIdx int) ([]string, []any, int) {
	var clauses []string
	var args []any
	idx := startIdx

	if def.SoftDelete {
		clauses = append(clauses, "deleted_at IS NULL")
	}

	// Composable filter predicate.
	if opts.Predicate != nil && opts.Predicate.Kind != filter.KindNone {
		clause, filterArgs, nextIdx := ToSQL(opts.Predicate, idx)
		if clause != "" {
			clauses = append(clauses, clause)
			args = append(args, filterArgs...)
			idx = nextIdx
		}
	}

	if len(opts.OrgUnitIDs) > 0 {
		// pgx encodes []uuid.UUID as a Postgres UUID array; ANY($n) works directly.
		clauses = append(clauses, fmt.Sprintf("org_unit_id = ANY($%d)", idx))
		args = append(args, opts.OrgUnitIDs)
		idx++
	}

	if opts.Search != "" {
		var searchParts []string
		for _, f := range def.Fields {
			if f.IsSearchable {
				searchParts = append(searchParts, fmt.Sprintf("%s ILIKE $%d", pgIdent(f.Name), idx))
			}
		}
		if len(searchParts) > 0 {
			clauses = append(clauses, "("+strings.Join(searchParts, " OR ")+")")
			args = append(args, "%"+opts.Search+"%")
			idx++
		}
	}

	return clauses, args, idx
}

// pgIdent quotes an identifier to prevent SQL injection.
// Only allows [a-z0-9_] — panics on anything else.
func pgIdent(name string) string {
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_') {
			panic(fmt.Sprintf("sqlbuilder: unsafe identifier %q", name))
		}
	}
	return name
}
