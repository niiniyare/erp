package sqlbuilder

import (
	"encoding/json"
	"fmt"
	"strings"

	"awo.so/framework/filter"
)

// ToSQL translates a Filter tree to a parameterised SQL WHERE clause fragment.
// startIdx is the first $N placeholder number to use.
//
// Returns (clause, args, nextIdx):
//   - clause is the SQL fragment (empty string when f is nil or KindNone)
//   - args are the values bound to the placeholders in order
//   - nextIdx is the next available placeholder number after this fragment
//
// The caller wraps the clause in WHERE (…) as appropriate.
func ToSQL(f *filter.Filter, startIdx int) (clause string, args []any, nextIdx int) {
	if f == nil || f.Kind == filter.KindNone {
		return "", nil, startIdx
	}

	idx := startIdx

	switch f.Kind {

	// ── Single-field comparisons ──────────────────────────────────────────────

	case filter.KindEq:
		return fmt.Sprintf("%s = $%d", pgIdent(f.Field), idx), []any{f.Value}, idx + 1

	case filter.KindNeq:
		return fmt.Sprintf("%s != $%d", pgIdent(f.Field), idx), []any{f.Value}, idx + 1

	case filter.KindGt:
		return fmt.Sprintf("%s > $%d", pgIdent(f.Field), idx), []any{f.Value}, idx + 1

	case filter.KindGte:
		return fmt.Sprintf("%s >= $%d", pgIdent(f.Field), idx), []any{f.Value}, idx + 1

	case filter.KindLt:
		return fmt.Sprintf("%s < $%d", pgIdent(f.Field), idx), []any{f.Value}, idx + 1

	case filter.KindLte:
		return fmt.Sprintf("%s <= $%d", pgIdent(f.Field), idx), []any{f.Value}, idx + 1

	// ── Set membership ────────────────────────────────────────────────────────

	case filter.KindIn:
		if len(f.Values) == 0 {
			return "FALSE", nil, idx // IN () is always false
		}
		placeholders := make([]string, len(f.Values))
		for i := range f.Values {
			placeholders[i] = fmt.Sprintf("$%d", idx+i)
		}
		clause = fmt.Sprintf("%s IN (%s)", pgIdent(f.Field), strings.Join(placeholders, ", "))
		return clause, f.Values, idx + len(f.Values)

	case filter.KindNotIn:
		if len(f.Values) == 0 {
			return "TRUE", nil, idx // NOT IN () is always true
		}
		placeholders := make([]string, len(f.Values))
		for i := range f.Values {
			placeholders[i] = fmt.Sprintf("$%d", idx+i)
		}
		clause = fmt.Sprintf("%s NOT IN (%s)", pgIdent(f.Field), strings.Join(placeholders, ", "))
		return clause, f.Values, idx + len(f.Values)

	// ── String matching ───────────────────────────────────────────────────────

	case filter.KindContains:
		val := "%" + fmt.Sprintf("%v", f.Value) + "%"
		return fmt.Sprintf("%s ILIKE $%d", pgIdent(f.Field), idx), []any{val}, idx + 1

	case filter.KindStartsWith:
		val := fmt.Sprintf("%v", f.Value) + "%"
		return fmt.Sprintf("%s LIKE $%d", pgIdent(f.Field), idx), []any{val}, idx + 1

	case filter.KindEndsWith:
		val := "%" + fmt.Sprintf("%v", f.Value)
		return fmt.Sprintf("%s LIKE $%d", pgIdent(f.Field), idx), []any{val}, idx + 1

	case filter.KindILike:
		return fmt.Sprintf("%s ILIKE $%d", pgIdent(f.Field), idx), []any{f.Value}, idx + 1

	// ── Null checks ───────────────────────────────────────────────────────────

	case filter.KindIsNull:
		return fmt.Sprintf("%s IS NULL", pgIdent(f.Field)), nil, idx

	case filter.KindIsNotNull:
		return fmt.Sprintf("%s IS NOT NULL", pgIdent(f.Field)), nil, idx

	// ── Logical ───────────────────────────────────────────────────────────────

	case filter.KindAnd:
		return joinChildren(f.Children, "AND", idx)

	case filter.KindOr:
		return joinChildren(f.Children, "OR", idx)

	case filter.KindNot:
		inner, innerArgs, nextIdx := ToSQL(f.Inner, idx)
		if inner == "" {
			return "", nil, nextIdx
		}
		return "NOT (" + inner + ")", innerArgs, nextIdx

	// ── JSONB ─────────────────────────────────────────────────────────────────

	case filter.KindJSONPath:
		// Translate dot-separated path to PostgreSQL -> / ->> chain.
		// The last segment uses ->> (text extraction); intermediate use ->.
		parts := strings.Split(f.Path, ".")
		col := pgIdent(f.Field)
		for i, p := range parts {
			if i < len(parts)-1 {
				col = fmt.Sprintf("(%s->'%s')", col, p)
			} else {
				col = fmt.Sprintf("(%s->>'%s')", col, p)
			}
		}
		// Translate inner filter using col as the field name.
		// We temporarily replace the inner's Field for rendering.
		innerClause, innerArgs, nextIdx := jsonPathInner(f.Inner, col, idx)
		return innerClause, innerArgs, nextIdx

	case filter.KindJSONContains:
		b, err := json.Marshal(f.Value)
		if err != nil {
			// Fall back to a safe FALSE rather than panicking.
			return "FALSE", nil, idx
		}
		clause = fmt.Sprintf("%s @> $%d::jsonb", pgIdent(f.Field), idx)
		return clause, []any{string(b)}, idx + 1
	}

	// Unknown kind — emit a safe FALSE rather than producing invalid SQL.
	return "FALSE", nil, idx
}

// joinChildren renders a list of Filter children joined with op (AND/OR).
func joinChildren(children []*filter.Filter, op string, startIdx int) (string, []any, int) {
	if len(children) == 0 {
		return "", nil, startIdx
	}
	parts := make([]string, 0, len(children))
	var allArgs []any
	idx := startIdx
	for _, child := range children {
		clause, args, next := ToSQL(child, idx)
		if clause == "" {
			continue
		}
		parts = append(parts, "("+clause+")")
		allArgs = append(allArgs, args...)
		idx = next
	}
	if len(parts) == 0 {
		return "", nil, idx
	}
	return strings.Join(parts, " "+op+" "), allArgs, idx
}

// jsonPathInner renders the inner filter using col as the already-built column expression.
func jsonPathInner(f *filter.Filter, col string, idx int) (string, []any, int) {
	if f == nil {
		return "", nil, idx
	}
	// Substitute col into the simple comparison operators.
	switch f.Kind {
	case filter.KindEq:
		return fmt.Sprintf("%s = $%d", col, idx), []any{f.Value}, idx + 1
	case filter.KindNeq:
		return fmt.Sprintf("%s != $%d", col, idx), []any{f.Value}, idx + 1
	case filter.KindContains:
		val := "%" + fmt.Sprintf("%v", f.Value) + "%"
		return fmt.Sprintf("%s ILIKE $%d", col, idx), []any{val}, idx + 1
	case filter.KindIsNull:
		return fmt.Sprintf("%s IS NULL", col), nil, idx
	case filter.KindIsNotNull:
		return fmt.Sprintf("%s IS NOT NULL", col), nil, idx
	default:
		// For complex inner filters, fall back to ToSQL with a synthesised filter.
		synth := *f
		synth.Field = col
		return ToSQL(&synth, idx)
	}
}
