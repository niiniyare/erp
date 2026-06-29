package sqlbuilder

// filterToSQL.go — internal bridge between filter.Filter trees and the
// parameterised SQL style used by this package.
//
// Why not use filter.ToSQL directly?
//
// filter.ToSQL (in the filter package) uses an internal parameter counter that
// starts at $1. sqlbuilder composes filter fragments into larger queries that
// already occupy $1…$k (e.g. for org-unit and search args). The stateful
// translator below threads the current placeholder index explicitly so all $n
// values in a composed query are globally unique and contiguous.
//
// If you ever make filter.ToSQL accept a startIdx, delete this file and call
// that instead.

import (
	"encoding/json"
	"fmt"
	"strings"

	"awo.so/framework/filter"
)

// ToSQL translates a filter tree to a PostgreSQL WHERE-clause fragment.
// Returns (clause, args, nextIdx) where nextIdx = startIdx + len(args).
// Panics if the filter tree contains an unsupported node kind.
func ToSQL(f *filter.Filter, startIdx int) (clause string, args []any, nextIdx int) {
	clause, args, err := filterToSQL(f, startIdx)
	if err != nil {
		panic("sqlbuilder.ToSQL: " + err.Error())
	}
	return clause, args, startIdx + len(args)
}

// filterToSQL translates a filter tree to a PostgreSQL WHERE-clause fragment,
// starting placeholder numbering at startIdx.
//
// Returns (clause, args, error):
//   - clause is the SQL fragment (empty when f is nil/KindNone)
//   - args are the values for the $n placeholders, in order
//
// The number of placeholders consumed is always len(args); the caller advances
// its own counter by that amount.
func filterToSQL(f *filter.Filter, startIdx int) (clause string, args []any, err error) {
	if f.IsNone() {
		return "", nil, nil
	}
	b := &indexedBuilder{idx: startIdx}
	clause, err = b.node(f)
	return clause, b.args, err
}

// indexedBuilder carries the mutable placeholder counter and accumulated args.
type indexedBuilder struct {
	idx  int
	args []any
}

// param registers v as the next query argument and returns its $n placeholder.
func (b *indexedBuilder) param(v any) string {
	b.args = append(b.args, v)
	s := fmt.Sprintf("$%d", b.idx)
	b.idx++
	return s
}

// node translates a single Filter node, recursing as needed.
func (b *indexedBuilder) node(f *filter.Filter) (string, error) {
	if f.IsNone() {
		return "", nil
	}

	switch f.Kind {

	// ── Scalar comparisons ────────────────────────────────────────────────────

	case filter.KindEq:
		col, err := pgIdent(f.Field)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s = %s", col, b.param(f.Value)), nil

	case filter.KindNeq:
		col, err := pgIdent(f.Field)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s != %s", col, b.param(f.Value)), nil

	case filter.KindGt:
		col, err := pgIdent(f.Field)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s > %s", col, b.param(f.Value)), nil

	case filter.KindGte:
		col, err := pgIdent(f.Field)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s >= %s", col, b.param(f.Value)), nil

	case filter.KindLt:
		col, err := pgIdent(f.Field)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s < %s", col, b.param(f.Value)), nil

	case filter.KindLte:
		col, err := pgIdent(f.Field)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s <= %s", col, b.param(f.Value)), nil

	// ── Range ─────────────────────────────────────────────────────────────────

	case filter.KindBetween:
		col, err := pgIdent(f.Field)
		if err != nil {
			return "", err
		}
		lo := b.param(f.Lo)
		hi := b.param(f.Hi)
		return fmt.Sprintf("%s BETWEEN %s AND %s", col, lo, hi), nil

	// ── Set membership ────────────────────────────────────────────────────────

	case filter.KindIn:
		if len(f.Values) == 0 {
			return "FALSE", nil // IN () is a SQL syntax error; semantically always false
		}
		return b.inClause(f.Field, f.Values, "IN")

	case filter.KindNotIn:
		if len(f.Values) == 0 {
			return "TRUE", nil // NOT IN () is vacuously true
		}
		return b.inClause(f.Field, f.Values, "NOT IN")

	// ── String matching (all ILIKE for case-insensitivity) ───────────────────

	case filter.KindContains:
		return b.ilike(f.Field, "%"+escapeLike(mustString(f.Value))+"%")

	case filter.KindStartsWith:
		return b.ilike(f.Field, escapeLike(mustString(f.Value))+"%")

	case filter.KindEndsWith:
		return b.ilike(f.Field, "%"+escapeLike(mustString(f.Value)))

	case filter.KindILike:
		// Caller supplies % wildcards; do NOT escape.
		return b.ilike(f.Field, mustString(f.Value))

	// ── Null checks ───────────────────────────────────────────────────────────

	case filter.KindIsNull:
		col, err := pgIdent(f.Field)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s IS NULL", col), nil

	case filter.KindIsNotNull:
		col, err := pgIdent(f.Field)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s IS NOT NULL", col), nil

	// ── Logical combinators ───────────────────────────────────────────────────

	case filter.KindAnd:
		return b.joinChildren(f.Children, "AND")

	case filter.KindOr:
		return b.joinChildren(f.Children, "OR")

	case filter.KindNot:
		inner, err := b.node(f.Inner)
		if err != nil {
			return "", err
		}
		if inner == "" {
			return "", nil
		}
		return "NOT (" + inner + ")", nil

	// ── JSONB ─────────────────────────────────────────────────────────────────

	case filter.KindJSONPath:
		return b.jsonPath(f)

	case filter.KindJSONContains:
		col, err := pgIdent(f.Field)
		if err != nil {
			return "", err
		}
		raw, err := json.Marshal(f.Value)
		if err != nil {
			return "", fmt.Errorf("filterToSQL: JSONContains: cannot marshal value: %w", err)
		}
		return fmt.Sprintf("%s @> %s::jsonb", col, b.param(string(raw))), nil

	default:
		return "", fmt.Errorf("filterToSQL: unsupported filter kind %s", f.Kind)
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// inClause emits "field IN ($n, …)" or "field NOT IN ($n, …)".
func (b *indexedBuilder) inClause(field string, values []any, op string) (string, error) {
	col, err := pgIdent(field)
	if err != nil {
		return "", err
	}
	placeholders := make([]string, len(values))
	for i, v := range values {
		placeholders[i] = b.param(v)
	}
	return fmt.Sprintf("%s %s (%s)", col, op, strings.Join(placeholders, ", ")), nil
}

// ilike emits "field ILIKE $n" with pattern as the argument.
func (b *indexedBuilder) ilike(field, pattern string) (string, error) {
	col, err := pgIdent(field)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s ILIKE %s", col, b.param(pattern)), nil
}

// joinChildren translates each child and joins them with the given SQL operator.
// Each child fragment is wrapped in parentheses for correct precedence.
func (b *indexedBuilder) joinChildren(children []*filter.Filter, op string) (string, error) {
	parts := make([]string, 0, len(children))
	for _, child := range children {
		part, err := b.node(child)
		if err != nil {
			return "", err
		}
		if part != "" {
			parts = append(parts, "("+part+")")
		}
	}
	if len(parts) == 0 {
		return "", nil
	}
	return strings.Join(parts, " "+op+" "), nil
}

// jsonPath converts a KindJSONPath node to a PostgreSQL JSON navigation
// expression.
//
// Dot-separated path segments are converted to -> / ->> operators:
//   - Intermediate segments: col->'seg'
//   - Final segment: col->>'seg'  (text extraction for comparison)
//
// The inner filter's Kind and Value determine the comparison operator; its
// Field is ignored (the expression built from Column+Path is used instead).
func (b *indexedBuilder) jsonPath(f *filter.Filter) (string, error) {
	if f.Inner == nil {
		return "", fmt.Errorf("filterToSQL: JSONPath column=%q path=%q has nil inner op", f.Column, f.Path)
	}
	if f.Path == "" {
		return "", fmt.Errorf("filterToSQL: JSONPath column=%q has empty path", f.Column)
	}

	col, err := pgIdent(f.Column)
	if err != nil {
		return "", fmt.Errorf("filterToSQL: JSONPath column: %w", err)
	}

	// Build the -> / ->> navigation chain from the dot-separated path.
	segments := strings.Split(f.Path, ".")
	expr := col
	for i, seg := range segments {
		// Escape single-quotes inside the key literal (SQL standard).
		escaped := strings.ReplaceAll(seg, "'", "''")
		if i < len(segments)-1 {
			expr = fmt.Sprintf("(%s->'%s')", expr, escaped)
		} else {
			expr = fmt.Sprintf("(%s->>'%s')", expr, escaped)
		}
	}

	// Apply the inner scalar operator against the JSON-extracted expression.
	// The inner filter's Field is intentionally empty; we use expr directly.
	op := f.Inner
	switch op.Kind {
	case filter.KindEq:
		return fmt.Sprintf("%s = %s", expr, b.param(op.Value)), nil
	case filter.KindNeq:
		return fmt.Sprintf("%s != %s", expr, b.param(op.Value)), nil
	case filter.KindGt:
		return fmt.Sprintf("%s > %s", expr, b.param(op.Value)), nil
	case filter.KindGte:
		return fmt.Sprintf("%s >= %s", expr, b.param(op.Value)), nil
	case filter.KindLt:
		return fmt.Sprintf("%s < %s", expr, b.param(op.Value)), nil
	case filter.KindLte:
		return fmt.Sprintf("%s <= %s", expr, b.param(op.Value)), nil
	case filter.KindContains:
		return fmt.Sprintf("%s ILIKE %s", expr, b.param("%"+escapeLike(mustString(op.Value))+"%")), nil
	case filter.KindStartsWith:
		return fmt.Sprintf("%s ILIKE %s", expr, b.param(escapeLike(mustString(op.Value))+"%")), nil
	case filter.KindEndsWith:
		return fmt.Sprintf("%s ILIKE %s", expr, b.param("%"+escapeLike(mustString(op.Value)))), nil
	case filter.KindIsNull:
		return fmt.Sprintf("%s IS NULL", expr), nil
	case filter.KindIsNotNull:
		return fmt.Sprintf("%s IS NOT NULL", expr), nil
	default:
		return "", fmt.Errorf("filterToSQL: JSONPath inner op %s is not a supported scalar kind", op.Kind)
	}
}

// mustString casts v to string. It is used for filter values that the DSL
// constructors already enforce to be strings (Contains, StartsWith, etc.).
// If v is somehow not a string (caller bypassed the constructor), the returned
// empty string will produce an obviously wrong but safe query rather than a panic.
func mustString(v any) string {
	s, _ := v.(string)
	return s
}
