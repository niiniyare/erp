// Package filter provides a composable, database-agnostic predicate DSL.
//
// Filters are immutable tree nodes. Build complex predicates by composing
// package-level constructors and method chaining:
//
//	f := filter.Eq("status", "submitted").
//	    And(filter.Gte("total_amount", 10000)).
//	    And(filter.Not(filter.In("customer_id", excludedIDs)))
//
// The persistence layer translates Filter trees to parameterised SQL.
package filter

// Kind identifies the type of predicate node.
type Kind uint8

const (
	KindEq           Kind = iota + 1 // field = $n
	KindNeq                           // field != $n
	KindGt                            // field > $n
	KindGte                           // field >= $n
	KindLt                            // field < $n
	KindLte                           // field <= $n
	KindIn                            // field IN ($n, ...)
	KindNotIn                         // field NOT IN ($n, ...)
	KindContains                      // field ILIKE '%val%'
	KindStartsWith                    // field LIKE 'val%'
	KindEndsWith                      // field LIKE '%val'
	KindILike                         // field ILIKE pattern (caller supplies %)
	KindIsNull                        // field IS NULL
	KindIsNotNull                     // field IS NOT NULL
	KindAnd                           // children joined with AND
	KindOr                            // children joined with OR
	KindNot                           // NOT (inner)
	KindNone                          // no predicate — matches every row
	KindJSONPath                      // (column->>'path') op $n
	KindJSONContains                  // column @> $n::jsonb
)

// Filter is a single node in a predicate tree.
// Construct with the package-level helper functions; never set fields directly.
type Filter struct {
	Kind     Kind
	Field    string    // column name (Eq, Neq, Gt, …, Contains, IsNull, …)
	Value    any       // single comparison value
	Values   []any     // for In / NotIn
	Children []*Filter // for And / Or
	Inner    *Filter   // for Not / JSONPath
	Path     string    // dot-separated key for JSONPath (e.g. "address.city")
}

// ── Method chaining ────────────────────────────────────────────────────────────

// And returns a new AND node combining f and other.
func (f *Filter) And(other *Filter) *Filter { return And(f, other) }

// Or returns a new OR node combining f and other.
func (f *Filter) Or(other *Filter) *Filter { return Or(f, other) }

// ── Constructors — equality / ordering ────────────────────────────────────────

// Eq matches rows where field = value.
func Eq(field string, value any) *Filter {
	return &Filter{Kind: KindEq, Field: field, Value: value}
}

// Neq matches rows where field != value.
func Neq(field string, value any) *Filter {
	return &Filter{Kind: KindNeq, Field: field, Value: value}
}

// Gt matches rows where field > value.
func Gt(field string, value any) *Filter {
	return &Filter{Kind: KindGt, Field: field, Value: value}
}

// Gte matches rows where field >= value.
func Gte(field string, value any) *Filter {
	return &Filter{Kind: KindGte, Field: field, Value: value}
}

// Lt matches rows where field < value.
func Lt(field string, value any) *Filter {
	return &Filter{Kind: KindLt, Field: field, Value: value}
}

// Lte matches rows where field <= value.
func Lte(field string, value any) *Filter {
	return &Filter{Kind: KindLte, Field: field, Value: value}
}

// ── Constructors — set membership ─────────────────────────────────────────────

// In matches rows where field is any of values.
func In(field string, values ...any) *Filter {
	return &Filter{Kind: KindIn, Field: field, Values: values}
}

// NotIn matches rows where field is none of values.
func NotIn(field string, values ...any) *Filter {
	return &Filter{Kind: KindNotIn, Field: field, Values: values}
}

// ── Constructors — string matching ────────────────────────────────────────────

// Contains matches rows where field ILIKE '%value%'.
func Contains(field string, value string) *Filter {
	return &Filter{Kind: KindContains, Field: field, Value: value}
}

// StartsWith matches rows where field LIKE 'value%'.
func StartsWith(field string, value string) *Filter {
	return &Filter{Kind: KindStartsWith, Field: field, Value: value}
}

// EndsWith matches rows where field LIKE '%value'.
func EndsWith(field string, value string) *Filter {
	return &Filter{Kind: KindEndsWith, Field: field, Value: value}
}

// ILike matches rows where field ILIKE pattern. Caller supplies '%' wildcards.
func ILike(field string, pattern string) *Filter {
	return &Filter{Kind: KindILike, Field: field, Value: pattern}
}

// ── Constructors — null checks ────────────────────────────────────────────────

// IsNull matches rows where field IS NULL.
func IsNull(field string) *Filter {
	return &Filter{Kind: KindIsNull, Field: field}
}

// IsNotNull matches rows where field IS NOT NULL.
func IsNotNull(field string) *Filter {
	return &Filter{Kind: KindIsNotNull, Field: field}
}

// ── Constructors — logical ─────────────────────────────────────────────────────

// And returns a node that requires all children to match.
func And(filters ...*Filter) *Filter {
	return &Filter{Kind: KindAnd, Children: filters}
}

// Or returns a node that requires at least one child to match.
func Or(filters ...*Filter) *Filter {
	return &Filter{Kind: KindOr, Children: filters}
}

// Not negates inner.
func Not(inner *Filter) *Filter {
	return &Filter{Kind: KindNot, Inner: inner}
}

// None returns a filter that matches every row (no WHERE clause generated).
func None() *Filter {
	return &Filter{Kind: KindNone}
}

// ── Constructors — JSONB ───────────────────────────────────────────────────────

// JSONPath matches rows where the value at dot-separated path within column
// satisfies inner. Example:
//
//	filter.JSONPath("custom_fields", "address.city", filter.Eq("", "Nairobi"))
//	→ custom_fields->'address'->>'city' = $n
func JSONPath(column string, path string, inner *Filter) *Filter {
	return &Filter{Kind: KindJSONPath, Field: column, Path: path, Inner: inner}
}

// JSONContains matches rows where column @> value::jsonb (containment).
// value should be a Go map or struct that marshals to the JSON you expect.
//
//	filter.JSONContains("custom_fields", map[string]any{"tags": []string{"vip"}})
func JSONContains(column string, value any) *Filter {
	return &Filter{Kind: KindJSONContains, Field: column, Value: value}
}
