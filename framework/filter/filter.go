// Package filter provides a composable, database-agnostic predicate DSL for
// building query filters that a persistence layer can translate to parameterised
// SQL (or any other query language).
//
// # Design principles
//
//   - Nodes are value-like: constructed via package-level helpers, never by
//     filling struct literals directly.
//   - Composition is done with free functions ([And], [Or], [Not]) or the
//     equivalent method shortcuts on *Filter.
//   - The zero value of *Filter is treated as [None] — it matches every row.
//     Translators must guard with [Filter.IsNone] rather than a nil check.
//   - Generic helpers ([InSlice], [NotInSlice]) eliminate the []T → []any
//     conversion that the variadic [In] / [NotIn] would otherwise require.
//
// # Typical usage
//
//	f := filter.And(
//	    filter.Eq("status", "submitted"),
//	    filter.Gte("total_amount", 10_000),
//	    filter.Not(filter.InSlice("customer_id", excludedIDs)),
//	    filter.Between("created_at", from, to),
//	)
//
// The persistence layer walks the returned tree and emits parameterised SQL.
// See [Kind] for the full set of node types.
package filter

import "fmt"

// ── Node type ─────────────────────────────────────────────────────────────────

// Kind identifies the semantic type of a [Filter] node.
// Each Kind has a corresponding constructor documented below.
type Kind uint8

const (
	// Scalar comparisons — Field op Value.
	KindEq  Kind = iota + 1 // field = $n
	KindNeq                 // field != $n
	KindGt                  // field > $n
	KindGte                 // field >= $n
	KindLt                  // field < $n
	KindLte                 // field <= $n

	// Range — Field BETWEEN lo AND hi.
	KindBetween // field BETWEEN $n AND $m

	// Set membership — Field IN / NOT IN ($n, …).
	KindIn    // field IN ($n, ...)
	KindNotIn // field NOT IN ($n, ...)

	// String matching (all case-insensitive via ILIKE unless noted).
	KindContains   // field ILIKE '%value%'
	KindStartsWith // field ILIKE 'value%'
	KindEndsWith   // field ILIKE '%value'
	KindILike      // field ILIKE pattern  (caller supplies % wildcards)

	// Null checks.
	KindIsNull    // field IS NULL
	KindIsNotNull // field IS NOT NULL

	// Logical combinators.
	KindAnd // AND(children…)
	KindOr  // OR(children…)
	KindNot // NOT(inner)

	// Sentinel — no predicate; matches every row.
	KindNone

	// JSONB operators (PostgreSQL).
	KindJSONPath     // (column->'a'->>'b') op $n
	KindJSONContains // column @> $n::jsonb
)

// String returns a human-readable name for the kind, useful in error messages
// and debug output from translators.
func (k Kind) String() string {
	names := [...]string{
		KindEq:           "Eq",
		KindNeq:          "Neq",
		KindGt:           "Gt",
		KindGte:          "Gte",
		KindLt:           "Lt",
		KindLte:          "Lte",
		KindBetween:      "Between",
		KindIn:           "In",
		KindNotIn:        "NotIn",
		KindContains:     "Contains",
		KindStartsWith:   "StartsWith",
		KindEndsWith:     "EndsWith",
		KindILike:        "ILike",
		KindIsNull:       "IsNull",
		KindIsNotNull:    "IsNotNull",
		KindAnd:          "And",
		KindOr:           "Or",
		KindNot:          "Not",
		KindNone:         "None",
		KindJSONPath:     "JSONPath",
		KindJSONContains: "JSONContains",
	}
	if int(k) < len(names) && names[k] != "" {
		return names[k]
	}
	return fmt.Sprintf("Kind(%d)", k)
}

// ── Core type ─────────────────────────────────────────────────────────────────

// Filter is a single node in a predicate tree. Use the package-level
// constructors — never fill struct fields directly, as the internal layout may
// change between versions.
//
// The zero value (*Filter)(nil) is equivalent to [None] and must be handled by
// translators; use [Filter.IsNone] to test for it.
type Filter struct {
	// Kind identifies which predicate this node represents.
	Kind Kind

	// Field is the column name for scalar, string, null, and JSONB nodes.
	Field string

	// Value holds the single comparison value for Eq, Neq, Gt, Gte, Lt, Lte,
	// Contains, StartsWith, EndsWith, ILike, JSONContains.
	Value any

	// Lo and Hi are the inclusive bounds for Between.
	Lo, Hi any

	// Values holds the set for In / NotIn.
	Values []any

	// Children are the sub-filters for And / Or.
	Children []*Filter

	// Inner is the operand for Not, and the operator-plus-value node for
	// JSONPath (always a scalar-kind Filter with an empty Field).
	Inner *Filter

	// JSONPath-specific fields.

	// Column is the top-level JSONB column name for JSONPath / JSONContains.
	// (Field is reused for JSONContains; Column is the dedicated slot for
	// JSONPath so Field is free for the inner scalar's Field, which is "").
	Column string

	// Path is a dot-separated key sequence into the JSONB column, e.g.
	// "address.city" or "contact.phones.0.number".
	// Array indices are supported as numeric segments; translators convert them
	// to the appropriate -> / ->> chain.
	Path string
}

// ── Convenience predicates ────────────────────────────────────────────────────

// IsNone reports whether f represents a no-op filter (nil or KindNone).
// Translators should call this before emitting a WHERE clause.
func (f *Filter) IsNone() bool {
	return f == nil || f.Kind == KindNone
}

// ── Method chaining ───────────────────────────────────────────────────────────

// And returns a new AND node combining the receiver with other.
// Equivalent to [And](f, other).
func (f *Filter) And(other *Filter) *Filter { return And(f, other) }

// Or returns a new OR node combining the receiver with other.
// Equivalent to [Or](f, other).
func (f *Filter) Or(other *Filter) *Filter { return Or(f, other) }

// ── Scalar comparison constructors ───────────────────────────────────────────

// Eq matches rows where field = value.
//
//	filter.Eq("status", "active")  →  status = $1
func Eq(field string, value any) *Filter {
	return &Filter{Kind: KindEq, Field: field, Value: value}
}

// Neq matches rows where field != value.
//
//	filter.Neq("status", "deleted")  →  status != $1
func Neq(field string, value any) *Filter {
	return &Filter{Kind: KindNeq, Field: field, Value: value}
}

// Gt matches rows where field > value.
//
//	filter.Gt("age", 18)  →  age > $1
func Gt(field string, value any) *Filter {
	return &Filter{Kind: KindGt, Field: field, Value: value}
}

// Gte matches rows where field >= value.
//
//	filter.Gte("total_amount", 10_000)  →  total_amount >= $1
func Gte(field string, value any) *Filter {
	return &Filter{Kind: KindGte, Field: field, Value: value}
}

// Lt matches rows where field < value.
//
//	filter.Lt("stock", 5)  →  stock < $1
func Lt(field string, value any) *Filter {
	return &Filter{Kind: KindLt, Field: field, Value: value}
}

// Lte matches rows where field <= value.
//
//	filter.Lte("price", 999.99)  →  price <= $1
func Lte(field string, value any) *Filter {
	return &Filter{Kind: KindLte, Field: field, Value: value}
}

// ── Range constructor ─────────────────────────────────────────────────────────

// Between matches rows where lo <= field <= hi (inclusive on both ends).
// It is semantically equivalent to Gte(field, lo).And(Lte(field, hi)) but
// gives translators a single node to map to SQL's BETWEEN … AND … and makes
// date-range intent explicit at the call site.
//
//	filter.Between("created_at", startOfMonth, endOfMonth)
//	→  created_at BETWEEN $1 AND $2
func Between(field string, lo, hi any) *Filter {
	return &Filter{Kind: KindBetween, Field: field, Lo: lo, Hi: hi}
}

// ── Set membership constructors ───────────────────────────────────────────────

// In matches rows where field is any of the provided values.
// For typed slices prefer [InSlice], which avoids the []T → []any conversion.
//
//	filter.In("role", "admin", "manager")  →  role IN ($1, $2)
func In(field string, values ...any) *Filter {
	return &Filter{Kind: KindIn, Field: field, Values: values}
}

// NotIn matches rows where field is none of the provided values.
// For typed slices prefer [NotInSlice].
//
//	filter.NotIn("status", "deleted", "archived")
//	→  status NOT IN ($1, $2)
func NotIn(field string, values ...any) *Filter {
	return &Filter{Kind: KindNotIn, Field: field, Values: values}
}

// InSlice matches rows where field is any element of the typed slice values.
// It eliminates the manual []T → []any conversion required when using [In].
//
//	ids := []int64{1, 2, 3}
//	filter.InSlice("customer_id", ids)  →  customer_id IN ($1, $2, $3)
func InSlice[T any](field string, values []T) *Filter {
	return &Filter{Kind: KindIn, Field: field, Values: toAnySlice(values)}
}

// NotInSlice matches rows where field is none of the elements in values.
// It is the typed-slice companion to [NotIn].
//
//	filter.NotInSlice("customer_id", blockedIDs)
//	→  customer_id NOT IN ($1, $2, $3)
func NotInSlice[T any](field string, values []T) *Filter {
	return &Filter{Kind: KindNotIn, Field: field, Values: toAnySlice(values)}
}

// ── String matching constructors ──────────────────────────────────────────────
//
// All string matchers are case-insensitive (ILIKE) except where noted.
// For a raw pattern with explicit '%' wildcards use [ILike].

// Contains matches rows where field ILIKE '%value%'.
//
//	filter.Contains("name", "coffee")  →  name ILIKE '%coffee%'
func Contains(field, value string) *Filter {
	return &Filter{Kind: KindContains, Field: field, Value: value}
}

// StartsWith matches rows where field ILIKE 'value%'.
//
//	filter.StartsWith("code", "KE-")  →  code ILIKE 'KE-%'
func StartsWith(field, value string) *Filter {
	return &Filter{Kind: KindStartsWith, Field: field, Value: value}
}

// EndsWith matches rows where field ILIKE '%value'.
//
//	filter.EndsWith("email", "@example.com")  →  email ILIKE '%@example.com'
func EndsWith(field, value string) *Filter {
	return &Filter{Kind: KindEndsWith, Field: field, Value: value}
}

// ILike matches rows where field ILIKE pattern. The caller is responsible for
// including '%' wildcards in the pattern. Use this when none of [Contains],
// [StartsWith], or [EndsWith] expresses the desired match shape.
//
//	filter.ILike("ref", "INV-%-2025")  →  ref ILIKE $1
func ILike(field, pattern string) *Filter {
	return &Filter{Kind: KindILike, Field: field, Value: pattern}
}

// ── Null-check constructors ───────────────────────────────────────────────────

// IsNull matches rows where field IS NULL.
//
//	filter.IsNull("deleted_at")  →  deleted_at IS NULL
func IsNull(field string) *Filter {
	return &Filter{Kind: KindIsNull, Field: field}
}

// IsNotNull matches rows where field IS NOT NULL.
//
//	filter.IsNotNull("approved_by")  →  approved_by IS NOT NULL
func IsNotNull(field string) *Filter {
	return &Filter{Kind: KindIsNotNull, Field: field}
}

// ── Logical combinators ───────────────────────────────────────────────────────

// And returns a filter that requires all of the provided filters to match.
//
// Edge-case normalisation:
//   - And() with no arguments returns [None] (matches everything).
//   - And(f) with a single argument returns f unchanged.
//   - Nil children are treated as [None] and are dropped; if all children are
//     nil the result is [None].
//
// This means callers can safely build filters in a loop without tracking
// whether any real predicates have been added:
//
//	var clauses []*filter.Filter
//	if req.Status != "" {
//	    clauses = append(clauses, filter.Eq("status", req.Status))
//	}
//	f := filter.And(clauses...)
func And(filters ...*Filter) *Filter {
	children := compactNone(filters)
	switch len(children) {
	case 0:
		return None()
	case 1:
		return children[0]
	}
	return &Filter{Kind: KindAnd, Children: children}
}

// Or returns a filter that requires at least one of the provided filters to
// match.
//
// Edge-case normalisation:
//   - Or() with no arguments returns [None].
//   - Or(f) with a single argument returns f unchanged.
//   - Nil / [KindNone] children are dropped. Unlike [And], if any child is
//     [None] it does NOT short-circuit to None — an Or of real predicates and
//     a None is still the union, but a bare None child is meaningless and is
//     pruned for cleaner SQL.
func Or(filters ...*Filter) *Filter {
	children := compactNone(filters)
	switch len(children) {
	case 0:
		return None()
	case 1:
		return children[0]
	}
	return &Filter{Kind: KindOr, Children: children}
}

// Not negates inner. If inner is nil or [KindNone], Not returns [None] because
// NOT(match-everything) is match-nothing, which is rarely useful — callers
// should be explicit about that intent.
func Not(inner *Filter) *Filter {
	if inner.IsNone() {
		// NOT(no predicate) would mean "match nothing", which is almost
		// certainly a caller mistake. Return None to be safe.
		return None()
	}
	return &Filter{Kind: KindNot, Inner: inner}
}

// None returns a filter that matches every row. The persistence layer must
// omit the WHERE clause entirely when the root filter is None.
//
// A nil *Filter is also treated as None by [Filter.IsNone] and translators,
// but using None() makes intent explicit.
func None() *Filter {
	return &Filter{Kind: KindNone}
}

// ── JSONB constructors ────────────────────────────────────────────────────────

// JSONPath matches rows where the value at a dot-separated key path inside a
// JSONB column satisfies a scalar predicate.
//
// path is a dot-separated sequence of object keys or array indices, e.g.
// "address.city" or "contact.phones.0.number". Translators convert this into
// the appropriate PostgreSQL -> / ->> chain, using ->> for the final segment
// so the extracted value is treated as text for the comparison.
//
// op must be a scalar *Filter constructed with one of the Eq / Neq / Gt / Gte /
// Lt / Lte / Contains / StartsWith / EndsWith / ILike helpers; its Field is
// ignored — only Kind and Value are used. Use the package-level scalar helpers
// and leave the field argument empty:
//
//	filter.JSONPath("meta", "address.city", filter.Eq("", "Nairobi"))
//	→  (meta->'address'->>'city') = $1
//
//	filter.JSONPath("meta", "score", filter.Gte("", 90))
//	→  (meta->>'score')::numeric >= $1
//
// Translators are responsible for any type-casting on the extracted text value.
func JSONPath(column, path string, op *Filter) *Filter {
	return &Filter{
		Kind:   KindJSONPath,
		Column: column,
		Path:   path,
		Inner:  op,
	}
}

// JSONContains matches rows where column @> value::jsonb (the containment
// operator). value should be a Go map or struct that marshals to the JSON
// sub-document you want to test for:
//
//	filter.JSONContains("custom_fields", map[string]any{"tags": []string{"vip"}})
//	→  custom_fields @> '{"tags":["vip"]}'::jsonb
//
// This is most efficient when a GIN index exists on the column.
func JSONContains(column string, value any) *Filter {
	return &Filter{Kind: KindJSONContains, Field: column, Value: value}
}

// ── Internal helpers ──────────────────────────────────────────────────────────

// toAnySlice converts a typed slice to []any without the caller having to do
// the conversion manually. Used by InSlice and NotInSlice.
func toAnySlice[T any](values []T) []any {
	out := make([]any, len(values))
	for i, v := range values {
		out[i] = v
	}
	return out
}

// compactNone removes nil and KindNone entries from filters, returning a new
// slice of only the meaningful predicates. Used by And and Or to normalise
// their children.
func compactNone(filters []*Filter) []*Filter {
	out := filters[:0:len(filters)] // reuse backing array; safe because we only shrink
	for _, f := range filters {
		if !f.IsNone() {
			out = append(out, f)
		}
	}
	return out
}
