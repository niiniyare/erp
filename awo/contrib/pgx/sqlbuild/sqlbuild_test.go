package sqlbuild_test

import (
	"testing"

	"awo.so/awo/contrib/pgx/sqlbuild"
	"awo.so/awo/filter"
)

func TestBuild_Nil(t *testing.T) {
	r, err := sqlbuild.Build(nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	if r.Clause != "" {
		t.Errorf("expected empty clause for nil filter, got %q", r.Clause)
	}
	if len(r.Args) != 0 {
		t.Errorf("expected no args, got %v", r.Args)
	}
}

func TestBuild_Eq(t *testing.T) {
	r, err := sqlbuild.Build(filter.Eq("status", "active"), 0)
	if err != nil {
		t.Fatal(err)
	}
	if r.Clause != `"status" = $1` {
		t.Errorf("unexpected clause: %q", r.Clause)
	}
	if len(r.Args) != 1 || r.Args[0] != "active" {
		t.Errorf("unexpected args: %v", r.Args)
	}
}

func TestBuild_EqNil(t *testing.T) {
	r, err := sqlbuild.Build(filter.Eq("deleted_at", nil), 0)
	if err != nil {
		t.Fatal(err)
	}
	if r.Clause != `"deleted_at" IS NULL` {
		t.Errorf("unexpected clause: %q", r.Clause)
	}
}

func TestBuild_In_Empty(t *testing.T) {
	r, err := sqlbuild.Build(filter.In("id"), 0)
	if err != nil {
		t.Fatal(err)
	}
	if r.Clause != "FALSE" {
		t.Errorf("expected FALSE for empty In, got %q", r.Clause)
	}
}

func TestBuild_And(t *testing.T) {
	f := filter.And(filter.Eq("a", 1), filter.Eq("b", 2))
	r, err := sqlbuild.Build(f, 0)
	if err != nil {
		t.Fatal(err)
	}
	expected := `("a" = $1) AND ("b" = $2)`
	if r.Clause != expected {
		t.Errorf("got %q, want %q", r.Clause, expected)
	}
}

func TestBuild_ParamOffset(t *testing.T) {
	// offset=2 → params start at $3
	r, err := sqlbuild.Build(filter.Eq("x", 99), 2)
	if err != nil {
		t.Fatal(err)
	}
	if r.Clause != `"x" = $3` {
		t.Errorf("unexpected clause: %q", r.Clause)
	}
}

func TestBuild_Between(t *testing.T) {
	r, err := sqlbuild.Build(filter.Between("amount", 100, 999), 0)
	if err != nil {
		t.Fatal(err)
	}
	if r.Clause != `"amount" BETWEEN $1 AND $2` {
		t.Errorf("unexpected clause: %q", r.Clause)
	}
}

func TestBuild_Not(t *testing.T) {
	r, err := sqlbuild.Build(filter.Not(filter.Eq("status", "draft")), 0)
	if err != nil {
		t.Fatal(err)
	}
	expected := `NOT ("status" = $1)`
	if r.Clause != expected {
		t.Errorf("got %q, want %q", r.Clause, expected)
	}
}
