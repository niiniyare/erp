package naming_test

import (
	"testing"
	"time"

	"awo.so/framework/def"
	"awo.so/framework/platform/naming"
)

func TestExpandPrefix_NoTokens(t *testing.T) {
	ns := &def.NamingSeriesDef{Prefix: "INV-"}
	got := naming.ExpandPrefix(ns, time.Date(2025, 6, 30, 0, 0, 0, 0, time.UTC))
	if got != "INV-" {
		t.Errorf("want INV-, got %q", got)
	}
}

func TestExpandPrefix_AllTokens(t *testing.T) {
	ns := &def.NamingSeriesDef{Prefix: "{YYYY}-{YY}-{MM}-{DD}-"}
	got := naming.ExpandPrefix(ns, time.Date(2025, 6, 7, 0, 0, 0, 0, time.UTC))
	want := "2025-25-06-07-"
	if got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestExpandPrefix_YYYY(t *testing.T) {
	ns := &def.NamingSeriesDef{Prefix: "INV-{YYYY}-"}
	got := naming.ExpandPrefix(ns, time.Date(2025, 6, 30, 0, 0, 0, 0, time.UTC))
	if got != "INV-2025-" {
		t.Errorf("want INV-2025-, got %q", got)
	}
}

func TestFormat_DefaultPadding(t *testing.T) {
	ns := &def.NamingSeriesDef{Prefix: "INV-", Padding: 0}
	got := naming.Format(ns, 42, time.Time{})
	if got != "INV-00042" {
		t.Errorf("want INV-00042, got %q", got)
	}
}

func TestFormat_CustomPadding(t *testing.T) {
	ns := &def.NamingSeriesDef{Prefix: "PO-", Padding: 3}
	got := naming.Format(ns, 7, time.Time{})
	if got != "PO-007" {
		t.Errorf("want PO-007, got %q", got)
	}
}

func TestPeriodKey_Never(t *testing.T) {
	ns := &def.NamingSeriesDef{ResetPeriod: def.ResetNever}
	got := naming.PeriodKey(ns, time.Now())
	if got != "" {
		t.Errorf("want empty string for ResetNever, got %q", got)
	}
}

func TestPeriodKey_Yearly(t *testing.T) {
	ns := &def.NamingSeriesDef{ResetPeriod: def.ResetYearly}
	got := naming.PeriodKey(ns, time.Date(2025, 6, 30, 0, 0, 0, 0, time.UTC))
	if got != "2025" {
		t.Errorf("want 2025, got %q", got)
	}
}

func TestPeriodKey_Monthly(t *testing.T) {
	ns := &def.NamingSeriesDef{ResetPeriod: def.ResetMonthly}
	got := naming.PeriodKey(ns, time.Date(2025, 6, 30, 0, 0, 0, 0, time.UTC))
	if got != "2025-06" {
		t.Errorf("want 2025-06, got %q", got)
	}
}
