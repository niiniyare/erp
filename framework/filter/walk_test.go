package filter_test

import (
	"errors"
	"testing"

	"awo.so/framework/filter"
)

func TestWalk_NilFilter_NoOp(t *testing.T) {
	var f *filter.Filter
	called := 0
	err := f.Walk(func(_ *filter.Filter) error { called++; return nil })
	if err != nil || called != 0 {
		t.Errorf("nil filter: want no call, got called=%d err=%v", called, err)
	}
}

func TestWalk_ScalarNode_VisitorCalledOnce(t *testing.T) {
	f := filter.Eq("status", "active")
	called := 0
	_ = f.Walk(func(n *filter.Filter) error {
		called++
		return nil
	})
	if called != 1 {
		t.Errorf("want 1 call, got %d", called)
	}
}

func TestWalk_And_VisitsAllChildren(t *testing.T) {
	f := filter.And(
		filter.Eq("status", "active"),
		filter.Gt("amount", 100),
		filter.IsNull("deleted_at"),
	)
	visited := 0
	_ = f.Walk(func(_ *filter.Filter) error { visited++; return nil })
	// 1 And node + 3 leaf nodes = 4 total.
	if visited != 4 {
		t.Errorf("want 4 visited, got %d", visited)
	}
}

func TestWalk_Or_VisitsAllChildren(t *testing.T) {
	f := filter.Or(filter.Eq("a", 1), filter.Eq("b", 2))
	visited := 0
	_ = f.Walk(func(_ *filter.Filter) error { visited++; return nil })
	if visited != 3 { // Or + 2 children
		t.Errorf("want 3 visited, got %d", visited)
	}
}

func TestWalk_Not_VisitsInner(t *testing.T) {
	f := filter.Not(filter.Eq("status", "deleted"))
	visited := 0
	_ = f.Walk(func(_ *filter.Filter) error { visited++; return nil })
	if visited != 2 { // Not + Eq
		t.Errorf("want 2 visited, got %d", visited)
	}
}

func TestWalk_ErrorStopsTraversal(t *testing.T) {
	f := filter.And(
		filter.Eq("a", 1),
		filter.Eq("b", 2),
	)
	sentinel := errors.New("stop")
	calls := 0
	err := f.Walk(func(_ *filter.Filter) error {
		calls++
		return sentinel // stop after first node (the And parent)
	})
	if !errors.Is(err, sentinel) {
		t.Errorf("want sentinel error, got %v", err)
	}
	if calls != 1 {
		t.Errorf("want 1 call before stop, got %d", calls)
	}
}

func TestFields_ReturnsLeafFieldNames(t *testing.T) {
	f := filter.And(
		filter.Eq("status", "active"),
		filter.Gt("amount", 100),
		filter.Eq("status", "active"), // duplicate — should deduplicate
	)
	fields := f.Fields()
	seen := map[string]bool{}
	for _, name := range fields {
		if seen[name] {
			t.Errorf("duplicate field %q in Fields()", name)
		}
		seen[name] = true
	}
	if !seen["status"] || !seen["amount"] {
		t.Errorf("want status and amount in fields, got %v", fields)
	}
}

func TestFields_JSONPath_UsesColumn(t *testing.T) {
	f := filter.JSONPath("meta", "address.city", filter.Eq("", "Nairobi"))
	fields := f.Fields()
	if len(fields) != 1 || fields[0] != "meta" {
		t.Errorf("want [meta], got %v", fields)
	}
}
