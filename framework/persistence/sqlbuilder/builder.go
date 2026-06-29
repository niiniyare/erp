// Package sqlbuilder generates parameterised PostgreSQL statements from
// EntityDefinition metadata.
//
// All exported functions are pure: they accept metadata and options, return
// SQL strings and argument slices, and never touch the database.
//
// # Placeholder numbering
//
// Functions that return both a query and args use 1-based PostgreSQL $n
// placeholders. When a caller needs to compose fragments (e.g. appending LIMIT
// and OFFSET), [SelectList] documents exactly which $n slots it occupies so
// the caller can append limit/offset at the right indices.
//
// # Identifier safety
//
// Column and table names are passed through [pgIdent], which double-quote-wraps
// the identifier and returns an error on characters outside [a-zA-Z0-9_].
// Callers that provide dynamic OrderBy values should validate them against
// [EntityDefinition.FieldNames] before calling [SelectList].
package sqlbuilder

import (
	"fmt"
	"strings"

	"github.com/google/uuid"

	"awo.so/framework/def"
	"awo.so/framework/filter"
)

// TotalColumn is the alias emitted by [SelectList] for the window-function
// total count. Callers scan this column to determine total result count without
// a separate COUNT query.
const TotalColumn = "__total"

// ── SELECT ────────────────────────────────────────────────────────────────────

// SelectOne builds a SELECT statement that fetches a single record by primary key.
//
//	SELECT <cols> FROM <table> WHERE id = $1 [AND deleted_at IS NULL]
//
// The caller supplies the id value as the sole query argument.
func SelectOne(def *def.EntityDefinition) (string, error) {
	cols, err := columnList(def)
	if err != nil {
		return "", err
	}
	table, err := pgIdent(def.TableName())
	if err != nil {
		return "", fmt.Errorf("sqlbuilder.SelectOne: %w", err)
	}
	q := fmt.Sprintf("SELECT %s FROM %s WHERE id = $1", cols, table)
	if def.SoftDelete {
		q += " AND deleted_at IS NULL"
	}
	return q, nil
}

// SelectOpts controls filtering, ordering, search, and org-unit scoping for
// [SelectList]. All fields are optional.
type SelectOpts struct {
	// Predicate is an optional composable filter tree built with the filter
	// package. It is ANDed with the implicit soft-delete and org-unit clauses.
	Predicate *filter.Filter

	// Search is a free-text string applied as ILIKE '%search%' across all
	// fields where FieldDefinition.IsSearchable is true. Special characters
	// (%, _) are escaped so the value is treated as a literal substring.
	Search string

	// OrderBy is the column to sort by. It must be a valid column identifier;
	// [SelectList] validates it with [pgIdent] and returns an error if it
	// contains unsafe characters. Defaults to "created_at" when empty.
	OrderBy string

	// Ascending controls sort direction. Default (false) is DESC.
	Ascending bool

	// OrgUnitIDs restricts results to the given org-unit IDs via
	//   org_unit_id = ANY($n)
	// An empty slice means no org-unit restriction.
	OrgUnitIDs []uuid.UUID
}

// SelectList builds a paginated SELECT with a window-function total count.
//
// The returned args slice covers all WHERE clause parameters. The caller must
// append limit and offset values (in that order) before executing:
//
//	query, args, err := sqlbuilder.SelectList(def, opts)
//	if err != nil { … }
//	args = append(args, limit, offset)
//	rows, err := pool.Query(ctx, query, args...)
//
// The result set includes a [TotalColumn] ("__total") column that holds the
// total number of matching rows, available on every row via the OVER() window.
func SelectList(def *def.EntityDefinition, opts SelectOpts) (string, []any, error) {
	cols, err := columnList(def)
	if err != nil {
		return "", nil, err
	}
	table, err := pgIdent(def.TableName())
	if err != nil {
		return "", nil, fmt.Errorf("sqlbuilder.SelectList: table name: %w", err)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(
		"SELECT %s, COUNT(*) OVER() AS %s FROM %s",
		cols, TotalColumn, table,
	))

	clauses, args, nextIdx, err := whereClauses(def, opts, 1)
	if err != nil {
		return "", nil, err
	}
	if len(clauses) > 0 {
		sb.WriteString(" WHERE ")
		sb.WriteString(strings.Join(clauses, " AND "))
	}

	// ORDER BY — validate the column name before interpolating.
	orderCol := "created_at"
	if opts.OrderBy != "" {
		if _, err := pgIdent(opts.OrderBy); err != nil {
			return "", nil, fmt.Errorf("sqlbuilder.SelectList: OrderBy: %w", err)
		}
		orderCol = opts.OrderBy
	}
	dir := "DESC"
	if opts.Ascending {
		dir = "ASC"
	}
	sb.WriteString(fmt.Sprintf(" ORDER BY %s %s", orderCol, dir))

	// LIMIT and OFFSET are NOT in args — the caller appends them.
	sb.WriteString(fmt.Sprintf(" LIMIT $%d OFFSET $%d", nextIdx, nextIdx+1))

	return sb.String(), args, nil
}

// ── INSERT ────────────────────────────────────────────────────────────────────

// Insert builds a single-row INSERT statement for the given mutable fields.
//
// Returns (query, cols, error) where cols is the ordered slice of column names
// whose values the caller must supply as positional args. The slice order
// matches the $n placeholders in query.
//
// The tenant_id column — if present for this entity's scope — is always written
// as the SQL function call current_tenant_id() and is never a caller parameter.
//
//	INSERT INTO <table> (id, tenant_id, col1, …)
//	VALUES ($1, current_tenant_id(), $2, …)
//	RETURNING id
func Insert(def *def.EntityDefinition, fields []string) (query string, cols []string, err error) {
	sqlCols, paramCols, err := buildColumnLists(def, fields)
	if err != nil {
		return "", nil, fmt.Errorf("sqlbuilder.Insert: %w", err)
	}

	placeholders, err := buildPlaceholders(sqlCols, 1)
	if err != nil {
		return "", nil, fmt.Errorf("sqlbuilder.Insert: %w", err)
	}

	table, err := pgIdent(def.TableName())
	if err != nil {
		return "", nil, fmt.Errorf("sqlbuilder.Insert: %w", err)
	}

	quotedCols, err := quoteIdents(sqlCols)
	if err != nil {
		return "", nil, fmt.Errorf("sqlbuilder.Insert: %w", err)
	}

	query = fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s) RETURNING id",
		table,
		strings.Join(quotedCols, ", "),
		strings.Join(placeholders, ", "),
	)
	return query, paramCols, nil
}

// BulkInsert builds a multi-row INSERT for n records.
//
// Placeholders are laid out in row-major order: all parameter columns for row 0
// (at $1…$k), then row 1 (at $k+1…$2k), and so on.
//
//	INSERT INTO <table> (id, tenant_id, col1, …)
//	VALUES ($1, current_tenant_id(), $2, …), ($k+1, current_tenant_id(), $k+2, …)
//	RETURNING id
//
// Returns (query, cols, error) where cols is the ordered parameter column list
// (same for every row). If n == 0 both query and cols are empty.
func BulkInsert(def *def.EntityDefinition, fields []string, n int) (query string, cols []string, err error) {
	if n == 0 {
		return "", nil, nil
	}

	sqlCols, paramCols, err := buildColumnLists(def, fields)
	if err != nil {
		return "", nil, fmt.Errorf("sqlbuilder.BulkInsert: %w", err)
	}

	table, err := pgIdent(def.TableName())
	if err != nil {
		return "", nil, fmt.Errorf("sqlbuilder.BulkInsert: %w", err)
	}

	quotedCols, err := quoteIdents(sqlCols)
	if err != nil {
		return "", nil, fmt.Errorf("sqlbuilder.BulkInsert: %w", err)
	}

	// paramWidth: number of $N slots consumed per row (tenant_id is a literal).
	paramWidth := len(paramCols)
	valueGroups := make([]string, n)
	for row := 0; row < n; row++ {
		placeholders, err := buildPlaceholders(sqlCols, row*paramWidth+1)
		if err != nil {
			return "", nil, fmt.Errorf("sqlbuilder.BulkInsert row %d: %w", row, err)
		}
		valueGroups[row] = "(" + strings.Join(placeholders, ", ") + ")"
	}

	query = fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s RETURNING id",
		table,
		strings.Join(quotedCols, ", "),
		strings.Join(valueGroups, ", "),
	)
	return query, paramCols, nil
}

// ── UPDATE ────────────────────────────────────────────────────────────────────

// immutableCols are columns that [Update] will silently skip even if the caller
// includes them in fields. Reparenting an entity to a different org unit must
// go through a dedicated reparent operation.
var immutableCols = map[string]bool{
	"id":          true,
	"tenant_id":   true,
	"org_unit_id": true,
	"created_at":  true,
}

// Update builds an UPDATE statement for the given mutable fields.
//
// Returns (query, cols, error). cols is the ordered list of columns whose
// values the caller supplies as $1…$k; id is always the final placeholder $k+1.
//
//	UPDATE <table> SET col1=$1, col2=$2, … WHERE id=$k+1 [AND deleted_at IS NULL]
func Update(def *def.EntityDefinition, fields []string) (query string, cols []string, err error) {
	mutable := make([]string, 0, len(fields))
	for _, f := range fields {
		if !immutableCols[f] {
			mutable = append(mutable, f)
		}
	}
	if len(mutable) == 0 {
		return "", nil, fmt.Errorf("sqlbuilder.Update: no mutable fields provided")
	}

	setClauses := make([]string, len(mutable))
	for i, col := range mutable {
		quoted, err := pgIdent(col)
		if err != nil {
			return "", nil, fmt.Errorf("sqlbuilder.Update: field %q: %w", col, err)
		}
		setClauses[i] = fmt.Sprintf("%s = $%d", quoted, i+1)
	}

	table, err := pgIdent(def.TableName())
	if err != nil {
		return "", nil, fmt.Errorf("sqlbuilder.Update: %w", err)
	}

	idParam := len(mutable) + 1
	where := fmt.Sprintf("id = $%d", idParam)
	if def.SoftDelete {
		where += " AND deleted_at IS NULL"
	}

	query = fmt.Sprintf(
		"UPDATE %s SET %s WHERE %s",
		table,
		strings.Join(setClauses, ", "),
		where,
	)
	return query, mutable, nil
}

// ── DELETE ────────────────────────────────────────────────────────────────────

// SoftDelete builds an UPDATE that stamps deleted_at = NOW() for a single record.
//
// The caller supplies the record id as $1.
func SoftDelete(def *def.EntityDefinition) (string, error) {
	table, err := pgIdent(def.TableName())
	if err != nil {
		return "", fmt.Errorf("sqlbuilder.SoftDelete: %w", err)
	}
	return fmt.Sprintf("UPDATE %s SET deleted_at = NOW() WHERE id = $1", table), nil
}

// HardDelete builds a DELETE statement for a single record.
//
// The caller supplies the record id as $1.
func HardDelete(def *def.EntityDefinition) (string, error) {
	table, err := pgIdent(def.TableName())
	if err != nil {
		return "", fmt.Errorf("sqlbuilder.HardDelete: %w", err)
	}
	return fmt.Sprintf("DELETE FROM %s WHERE id = $1", table), nil
}

// ── COUNT / EXISTS ────────────────────────────────────────────────────────────

// Count builds a SELECT COUNT(*) query.
//
// filterFields is an optional list of columns for exact-match equality
// predicates ($1, $2, …). The soft-delete guard (if any) does not consume
// a placeholder.
//
// The caller supplies values for filterFields (in order) as query args.
func Count(def *def.EntityDefinition, filterFields []string) (string, error) {
	table, err := pgIdent(def.TableName())
	if err != nil {
		return "", fmt.Errorf("sqlbuilder.Count: %w", err)
	}

	// soft-delete guard has no placeholder; field predicates start at $1.
	clauses := make([]string, 0, len(filterFields)+1)
	if def.SoftDelete {
		clauses = append(clauses, "deleted_at IS NULL")
	}
	for i, f := range filterFields {
		quoted, err := pgIdent(f)
		if err != nil {
			return "", fmt.Errorf("sqlbuilder.Count: field %q: %w", f, err)
		}
		clauses = append(clauses, fmt.Sprintf("%s = $%d", quoted, i+1))
	}

	q := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
	if len(clauses) > 0 {
		q += " WHERE " + strings.Join(clauses, " AND ")
	}
	return q, nil
}

// Exists builds a cheap SELECT EXISTS(…) query.
//
// filterFields are exact-match equality predicates ($1, $2, …). Soft-delete
// guard (if any) is appended after the parameterised fields and does not
// consume a placeholder. The caller supplies values for filterFields as args.
func Exists(def *def.EntityDefinition, filterFields []string) (string, error) {
	table, err := pgIdent(def.TableName())
	if err != nil {
		return "", fmt.Errorf("sqlbuilder.Exists: %w", err)
	}

	clauses := make([]string, 0, len(filterFields)+1)
	for i, f := range filterFields {
		quoted, err := pgIdent(f)
		if err != nil {
			return "", fmt.Errorf("sqlbuilder.Exists: field %q: %w", f, err)
		}
		clauses = append(clauses, fmt.Sprintf("%s = $%d", quoted, i+1))
	}
	if def.SoftDelete {
		clauses = append(clauses, "deleted_at IS NULL")
	}

	return fmt.Sprintf(
		"SELECT EXISTS(SELECT 1 FROM %s WHERE %s)",
		table,
		strings.Join(clauses, " AND "),
	), nil
}

// ── Internal helpers ──────────────────────────────────────────────────────────

// columnList returns a comma-separated, quoted column list for SELECT statements.
// System columns (id, scope cols, timestamps) come first; entity fields follow.
func columnList(def *def.EntityDefinition) (string, error) {
	scope := scopeColumns(def)
	cols := make([]string, 0, len(def.Fields)+5)

	// System columns — these are known-safe identifiers; quoting is defensive.
	for _, sys := range append([]string{"id"}, scope...) {
		q, err := pgIdent(sys)
		if err != nil {
			return "", fmt.Errorf("sqlbuilder: system column %q: %w", sys, err)
		}
		cols = append(cols, q)
	}
	for _, ts := range []string{"created_at", "updated_at"} {
		q, _ := pgIdent(ts) // known-safe
		cols = append(cols, q)
	}
	if def.SoftDelete {
		cols = append(cols, "deleted_at")
	}
	for _, f := range def.Fields {
		q, err := pgIdent(f.Name)
		if err != nil {
			return "", fmt.Errorf("sqlbuilder: field %q: %w", f.Name, err)
		}
		cols = append(cols, q)
	}
	return strings.Join(cols, ", "), nil
}

// scopeColumns returns the scope-enforcement columns for the entity's OrgScope:
//
//   - Global   → nil
//   - Tenant   → ["tenant_id"]
//   - Unit     → ["tenant_id", "org_unit_id"]
func scopeColumns(def *def.EntityDefinition) []string {
	switch {
	case def.IsGlobal():
		return nil
	case def.IsUnitScoped():
		return []string{"tenant_id", "org_unit_id"}
	default:
		return []string{"tenant_id"}
	}
}

// buildColumnLists derives two parallel slices from a field list:
//
//   - sqlCols: every column that appears in the INSERT column list,
//     including id, scope cols, and the supplied fields.
//   - paramCols: columns whose values are caller-supplied $N parameters
//     (tenant_id is excluded because it maps to the current_tenant_id() literal).
func buildColumnLists(def *def.EntityDefinition, fields []string) (sqlCols, paramCols []string, err error) {
	scopeCols := scopeColumns(def)
	reserved := make(map[string]bool, len(scopeCols)+1)
	reserved["id"] = true
	for _, sc := range scopeCols {
		reserved[sc] = true
	}

	sqlCols = make([]string, 0, len(fields)+len(scopeCols)+1)
	paramCols = make([]string, 0, len(fields)+len(scopeCols)+1)

	sqlCols = append(sqlCols, "id")
	paramCols = append(paramCols, "id")

	for _, sc := range scopeCols {
		sqlCols = append(sqlCols, sc)
		if sc != "tenant_id" {
			paramCols = append(paramCols, sc)
		}
	}
	for _, f := range fields {
		if reserved[f] {
			continue
		}
		if _, verr := pgIdent(f); verr != nil {
			return nil, nil, fmt.Errorf("field %q: %w", f, verr)
		}
		sqlCols = append(sqlCols, f)
		paramCols = append(paramCols, f)
	}
	return sqlCols, paramCols, nil
}

// buildPlaceholders returns a placeholder slice aligned with sqlCols.
// tenant_id maps to the literal current_tenant_id(); all other columns get
// a $N placeholder. startParam is the first $N value to use.
func buildPlaceholders(sqlCols []string, startParam int) ([]string, error) {
	placeholders := make([]string, len(sqlCols))
	paramIdx := startParam
	for i, col := range sqlCols {
		if col == "tenant_id" {
			placeholders[i] = "current_tenant_id()"
		} else {
			placeholders[i] = fmt.Sprintf("$%d", paramIdx)
			paramIdx++
		}
	}
	return placeholders, nil
}

// quoteIdents returns a new slice with each identifier double-quote-wrapped.
func quoteIdents(names []string) ([]string, error) {
	out := make([]string, len(names))
	for i, n := range names {
		q, err := pgIdent(n)
		if err != nil {
			return nil, err
		}
		out[i] = q
	}
	return out, nil
}

// whereClauses assembles the AND fragments for a SelectList WHERE clause.
//
// Placeholder allocation order:
//  1. soft-delete guard (no placeholder)
//  2. filter.Predicate ($startIdx … $n)
//  3. org_unit_id = ANY($n+1)  — one slice param
//  4. search ILIKE ($n+2)      — shared across all searchable fields
//
// Returns (clauses, args, nextIdx, error). nextIdx is the first unused $N
// after all WHERE clause parameters (LIMIT / OFFSET are appended at nextIdx
// and nextIdx+1 by [SelectList]).
func whereClauses(def *def.EntityDefinition, opts SelectOpts, startIdx int) ([]string, []any, int, error) {
	var clauses []string
	var args []any
	idx := startIdx

	if def.SoftDelete {
		clauses = append(clauses, "deleted_at IS NULL")
	}

	// Composable predicate from filter package.
	if !opts.Predicate.IsNone() {
		clause, filterArgs, err := filterToSQL(opts.Predicate, idx)
		if err != nil {
			return nil, nil, 0, fmt.Errorf("sqlbuilder: predicate: %w", err)
		}
		if clause != "" {
			clauses = append(clauses, clause)
			args = append(args, filterArgs...)
			idx += len(filterArgs)
		}
	}

	// Org-unit restriction — pgx encodes []uuid.UUID natively for ANY($n).
	if len(opts.OrgUnitIDs) > 0 {
		clauses = append(clauses, fmt.Sprintf("org_unit_id = ANY($%d)", idx))
		args = append(args, opts.OrgUnitIDs)
		idx++
	}

	// Full-text search across searchable fields.
	if opts.Search != "" {
		var searchParts []string
		for _, f := range def.Fields {
			if !f.IsSearchable {
				continue
			}
			quoted, err := pgIdent(f.Name)
			if err != nil {
				return nil, nil, 0, fmt.Errorf("sqlbuilder: search field %q: %w", f.Name, err)
			}
			searchParts = append(searchParts, fmt.Sprintf("%s ILIKE $%d", quoted, idx))
		}
		if len(searchParts) > 0 {
			clauses = append(clauses, "("+strings.Join(searchParts, " OR ")+")")
			// Escape literal % and _ so the search term is treated as a plain
			// substring, not an ILIKE pattern.
			args = append(args, "%"+escapeLike(opts.Search)+"%")
			idx++
		}
	}

	return clauses, args, idx, nil
}

// escapeLike escapes ILIKE special characters in a literal search term.
func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, "%", `\%`)
	s = strings.ReplaceAll(s, "_", `\_`)
	return s
}

// pgIdent double-quote-wraps a PostgreSQL identifier and returns an error if it
// contains characters outside [a-zA-Z0-9_]. It does not allow schema-qualified
// names (no dots); callers must handle qualification themselves.
//
// This is a security boundary: all caller-supplied names (field names, OrderBy)
// must pass through pgIdent before being interpolated into SQL.
func pgIdent(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("sqlbuilder: empty identifier")
	}
	for _, c := range name {
		ok := (c >= 'a' && c <= 'z') ||
			(c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') ||
			c == '_'
		if !ok {
			return "", fmt.Errorf("sqlbuilder: unsafe identifier %q (char %q)", name, c)
		}
	}
	return `"` + name + `"`, nil
}
