package sqlbuilder_test

import (
	"strings"
	"testing"

	"awo.so/framework/filter"
	"awo.so/framework/persistence/sqlbuilder"
)

// ── pgIdent ───────────────────────────────────────────────────────────────────

func TestPgIdent_Safe(t *testing.T) {
	cases := []struct{ in, want string }{
		{"status", `"status"`},
		{"org_unit_id", `"org_unit_id"`},
		{"col123", `"col123"`},
		{"CamelCase", `"CamelCase"`},
	}
	for _, tc := range cases {
		got, err := sqlbuilder.ExportedPgIdent(tc.in)
		if err != nil {
			t.Errorf("pgIdent(%q): unexpected error: %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("pgIdent(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestPgIdent_Unsafe(t *testing.T) {
	bad := []string{"a.b", "col; DROP TABLE", `"x"`, "col-name", ""}
	for _, name := range bad {
		if _, err := sqlbuilder.ExportedPgIdent(name); err == nil {
			t.Errorf("pgIdent(%q): expected error, got nil", name)
		}
	}
}

// ── filterToSQL ───────────────────────────────────────────────────────────────

func TestFilterToSQL_Scalars(t *testing.T) {
	tests := []struct {
		name     string
		f        *filter.Filter
		startIdx int
		wantSQL  string
		wantArgs []any
	}{
		{
			name:     "Eq at $1",
			f:        filter.Eq("status", "active"),
			startIdx: 1,
			wantSQL:  `"status" = $1`,
			wantArgs: []any{"active"},
		},
		{
			name:     "Eq at $3 (composed query)",
			f:        filter.Eq("status", "active"),
			startIdx: 3,
			wantSQL:  `"status" = $3`,
			wantArgs: []any{"active"},
		},
		{
			name:     "Neq",
			f:        filter.Neq("status", "deleted"),
			startIdx: 1,
			wantSQL:  `"status" != $1`,
			wantArgs: []any{"deleted"},
		},
		{
			name:     "Gt",
			f:        filter.Gt("age", 18),
			startIdx: 1,
			wantSQL:  `"age" > $1`,
			wantArgs: []any{18},
		},
		{
			name:     "Gte",
			f:        filter.Gte("total", 1000),
			startIdx: 1,
			wantSQL:  `"total" >= $1`,
			wantArgs: []any{1000},
		},
		{
			name:     "Lt",
			f:        filter.Lt("stock", 5),
			startIdx: 1,
			wantSQL:  `"stock" < $1`,
			wantArgs: []any{5},
		},
		{
			name:     "Lte",
			f:        filter.Lte("price", 99.99),
			startIdx: 1,
			wantSQL:  `"price" <= $1`,
			wantArgs: []any{99.99},
		},
		{
			name:     "Between",
			f:        filter.Between("created_at", "2025-01-01", "2025-12-31"),
			startIdx: 2,
			wantSQL:  `"created_at" BETWEEN $2 AND $3`,
			wantArgs: []any{"2025-01-01", "2025-12-31"},
		},
		{
			name:     "Contains escapes wildcards",
			f:        filter.Contains("note", "50% off"),
			startIdx: 1,
			wantSQL:  `"note" ILIKE $1`,
			wantArgs: []any{`%50\% off%`},
		},
		{
			name:     "StartsWith uses ILIKE",
			f:        filter.StartsWith("code", "KE-"),
			startIdx: 1,
			wantSQL:  `"code" ILIKE $1`,
			wantArgs: []any{`KE-%`},
		},
		{
			name:     "EndsWith uses ILIKE",
			f:        filter.EndsWith("email", "@test.com"),
			startIdx: 1,
			wantSQL:  `"email" ILIKE $1`,
			wantArgs: []any{`%@test.com`},
		},
		{
			name:     "IsNull",
			f:        filter.IsNull("deleted_at"),
			startIdx: 1,
			wantSQL:  `"deleted_at" IS NULL`,
			wantArgs: nil,
		},
		{
			name:     "IsNotNull",
			f:        filter.IsNotNull("approved_by"),
			startIdx: 1,
			wantSQL:  `"approved_by" IS NOT NULL`,
			wantArgs: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assertFilterSQL(t, tc.f, tc.startIdx, tc.wantSQL, tc.wantArgs)
		})
	}
}

func TestFilterToSQL_InSlice(t *testing.T) {
	f := filter.InSlice("customer_id", []int64{1, 2, 3})
	assertFilterSQL(
		t, f, 1,
		`"customer_id" IN ($1, $2, $3)`,
		[]any{int64(1), int64(2), int64(3)},
	)
}

func TestFilterToSQL_InEmpty(t *testing.T) {
	clause, args, err := sqlbuilder.ExportedFilterToSQL(filter.In("id"), 1)
	assertNoErr(t, err)
	if clause != "FALSE" {
		t.Errorf("In() empty: got %q, want FALSE", clause)
	}
	if len(args) != 0 {
		t.Errorf("In() empty: unexpected args %v", args)
	}
}

func TestFilterToSQL_NotInEmpty(t *testing.T) {
	clause, args, err := sqlbuilder.ExportedFilterToSQL(filter.NotIn("id"), 1)
	assertNoErr(t, err)
	if clause != "TRUE" {
		t.Errorf("NotIn() empty: got %q, want TRUE", clause)
	}
	if len(args) != 0 {
		t.Errorf("NotIn() empty: unexpected args %v", args)
	}
}

func TestFilterToSQL_LogicalAnd(t *testing.T) {
	f := filter.And(
		filter.Eq("status", "active"),
		filter.Gte("total", 100),
	)
	assertFilterSQL(
		t, f, 1,
		`("status" = $1) AND ("total" >= $2)`,
		[]any{"active", 100},
	)
}

func TestFilterToSQL_LogicalOr(t *testing.T) {
	f := filter.Or(
		filter.Eq("type", "A"),
		filter.Eq("type", "B"),
	)
	assertFilterSQL(
		t, f, 5,
		`("type" = $5) OR ("type" = $6)`,
		[]any{"A", "B"},
	)
}

func TestFilterToSQL_Not(t *testing.T) {
	f := filter.Not(filter.Eq("status", "deleted"))
	assertFilterSQL(
		t, f, 1,
		`NOT ("status" = $1)`,
		[]any{"deleted"},
	)
}

func TestFilterToSQL_Nested(t *testing.T) {
	f := filter.And(
		filter.Eq("status", "submitted"),
		filter.Or(
			filter.Lt("amount", 500),
			filter.Gte("amount", 10_000),
		),
	)
	assertFilterSQL(
		t, f, 1,
		`("status" = $1) AND (("amount" < $2) OR ("amount" >= $3))`,
		[]any{"submitted", 500, 10_000},
	)
}

func TestFilterToSQL_JSONPath_MultiSegment(t *testing.T) {
	f := filter.JSONPath("meta", "address.city", filter.Eq("", "Nairobi"))
	assertFilterSQL(
		t, f, 1,
		`("meta"->'address'->>'city') = $1`,
		[]any{"Nairobi"},
	)
}

func TestFilterToSQL_JSONPath_SingleSegment(t *testing.T) {
	f := filter.JSONPath("config", "theme", filter.Eq("", "dark"))
	assertFilterSQL(
		t, f, 1,
		`("config"->>'theme') = $1`,
		[]any{"dark"},
	)
}

func TestFilterToSQL_JSONPath_AtNonOneIdx(t *testing.T) {
	// Ensure JSON path respects the offset startIdx.
	f := filter.JSONPath("meta", "score", filter.Gte("", 90))
	assertFilterSQL(
		t, f, 4,
		`("meta"->>'score') >= $4`,
		[]any{90},
	)
}

func TestFilterToSQL_None(t *testing.T) {
	clause, args, err := sqlbuilder.ExportedFilterToSQL(filter.None(), 1)
	assertNoErr(t, err)
	if clause != "" {
		t.Errorf("None(): got clause %q, want empty", clause)
	}
	if len(args) != 0 {
		t.Errorf("None(): got args %v, want none", args)
	}
}

func TestFilterToSQL_NilFilter(t *testing.T) {
	clause, args, err := sqlbuilder.ExportedFilterToSQL(nil, 1)
	assertNoErr(t, err)
	if clause != "" || len(args) != 0 {
		t.Error("nil filter should produce empty clause and no args")
	}
}

func TestFilterToSQL_PlaceholderContinuity(t *testing.T) {
	// Simulates composing after $1 is already occupied (e.g. by tenant_id).
	f := filter.And(
		filter.Eq("status", "open"),
		filter.Lt("amount", 500),
	)
	clause, args, err := sqlbuilder.ExportedFilterToSQL(f, 2)
	assertNoErr(t, err)
	if !strings.Contains(clause, "$2") || !strings.Contains(clause, "$3") {
		t.Errorf("expected $2 and $3 in clause, got: %s", clause)
	}
	if len(args) != 2 {
		t.Errorf("expected 2 args, got %d: %v", len(args), args)
	}
}

func TestFilterToSQL_UnsafeFieldName(t *testing.T) {
	f := filter.Eq("bad-field", "x")
	_, _, err := sqlbuilder.ExportedFilterToSQL(f, 1)
	if err == nil {
		t.Error("expected error for unsafe field name, got nil")
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func assertFilterSQL(t *testing.T, f *filter.Filter, startIdx int, wantSQL string, wantArgs []any) {
	t.Helper()
	clause, args, err := sqlbuilder.ExportedFilterToSQL(f, startIdx)
	assertNoErr(t, err)
	if clause != wantSQL {
		t.Errorf("SQL mismatch\n got:  %s\n want: %s", clause, wantSQL)
	}
	if len(args) != len(wantArgs) {
		t.Fatalf("args len: got %d, want %d\n got:  %v\n want: %v",
			len(args), len(wantArgs), args, wantArgs)
	}
	for i := range args {
		if args[i] != wantArgs[i] {
			t.Errorf("arg[%d]: got %v, want %v", i, args[i], wantArgs[i])
		}
	}
}

func assertNoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
