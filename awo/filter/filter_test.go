package filter_test

import (
	"testing"

	"awo.so/awo/filter"
)

func TestEq_String(t *testing.T) {
	f := filter.Eq("status", "Draft")
	got := f.String()
	if got != `status eq "Draft"` {
		t.Errorf("got %q", got)
	}
}

func TestEq_Nil_BecomesIsNull(t *testing.T) {
	f := filter.Eq("deleted_at", nil)
	if f.Kind != filter.KindIsNull {
		t.Errorf("expected KindIsNull, got %v", f.Kind)
	}
}

func TestAnd_SingleFilter_Unwrapped(t *testing.T) {
	inner := filter.Eq("status", "Draft")
	result := filter.And(inner)
	if result != inner {
		t.Error("And with one filter should return that filter unchanged")
	}
}

func TestAnd_NilFilters_SkipsNil(t *testing.T) {
	a := filter.Eq("status", "Draft")
	result := filter.And(a, nil, nil)
	if result != a {
		t.Error("And should skip nil filters and return the single non-nil filter")
	}
}

func TestAnd_AllNil_ReturnsNil(t *testing.T) {
	result := filter.And(nil, nil)
	if result != nil {
		t.Error("And with all-nil filters should return nil")
	}
}

func TestOr_MultipleFilters(t *testing.T) {
	a := filter.Eq("status", "Draft")
	b := filter.Eq("status", "Submitted")
	result := filter.Or(a, b)
	if result.Kind != filter.KindOr {
		t.Errorf("expected KindOr, got %v", result.Kind)
	}
	if len(result.Sub) != 2 {
		t.Errorf("expected 2 sub-filters, got %d", len(result.Sub))
	}
}

func TestNot_NilInput_ReturnsNil(t *testing.T) {
	result := filter.Not(nil)
	if result != nil {
		t.Error("Not(nil) should return nil")
	}
}

func TestNot_WrapsFilter(t *testing.T) {
	inner := filter.Eq("archived", true)
	result := filter.Not(inner)
	if result.Kind != filter.KindNot {
		t.Errorf("expected KindNot, got %v", result.Kind)
	}
	if len(result.Sub) != 1 || result.Sub[0] != inner {
		t.Error("Not should wrap inner filter in Sub[0]")
	}
}

func TestIn_Empty_StillBuilds(t *testing.T) {
	f := filter.In("status") // no values
	if f == nil {
		t.Fatal("In with no values should return a non-nil Filter")
	}
	if f.Kind != filter.KindIn {
		t.Errorf("expected KindIn, got %v", f.Kind)
	}
}

func TestBetween(t *testing.T) {
	f := filter.Between("age", int64(18), int64(65))
	if f.Kind != filter.KindBetween {
		t.Errorf("expected KindBetween, got %v", f.Kind)
	}
	if f.Lo != int64(18) || f.Hi != int64(65) {
		t.Errorf("Lo/Hi mismatch: got %v/%v", f.Lo, f.Hi)
	}
}

func TestFilterImplementsDefFilter(t *testing.T) {
	// *filter.Filter satisfies def.Filter (empty interface) — verified by the
	// compiler whenever a *Filter is assigned to a def.Filter variable.
	var _ interface{} = (*filter.Filter)(nil)
}
