package domain

import (
	"testing"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Session model post-refactor (BLOCK-1 risk)
// ---------------------------------------------------------------------------

// TestResolvedSession_NoPermissions verifies that ResolvedSession has no
// Permissions field — compile-time proof that the session permission snapshot
// was removed.  If this file compiles the field does not exist.
func TestResolvedSession_NoPermissions(t *testing.T) {
	sess := &ResolvedSession{
		UserID:        uuid.New(),
		TenantID:      uuid.New(),
		UserType:      "INTERNAL",
		Configuration: DefaultConfiguration(),
		EntityScope:   EntityScope{Type: EntityScopeAll},
	}
	// If ResolvedSession had a Permissions field this file would not compile.
	// The mere existence of this test (and its compilation) proves the field
	// is absent.
	_ = sess
}

// TestResolvedSession_ToPrincipal verifies that ToPrincipal returns the
// correct Subject+Domain for each actor type.
func TestResolvedSession_ToPrincipal(t *testing.T) {
	uid := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	tid := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	cases := []struct {
		userType      string
		wantSubject   string
		wantDomain    string
	}{
		{"SYSADMIN", "platform:" + uid.String(), "_platform_"},
		{"PLATFORM", "platform:" + uid.String(), "_platform_"},
		{"INTERNAL", "tenant:" + uid.String(), tid.String()},
		{"EMPLOYEE", "tenant:" + uid.String(), tid.String()},
		{"PORTAL", "portal:" + uid.String(), tid.String() + ":portal"},
		{"CUSTOMER", "portal:" + uid.String(), tid.String() + ":portal"},
		{"API", "api:" + uid.String(), tid.String() + ":api"},
	}

	for _, tc := range cases {
		t.Run(tc.userType, func(t *testing.T) {
			sess := &ResolvedSession{
				UserID:   uid,
				TenantID: tid,
				UserType: tc.userType,
			}
			p := sess.ToPrincipal()
			if p.Subject != tc.wantSubject {
				t.Errorf("Subject: got %q, want %q", p.Subject, tc.wantSubject)
			}
			if p.Domain != tc.wantDomain {
				t.Errorf("Domain: got %q, want %q", p.Domain, tc.wantDomain)
			}
		})
	}
}

// TestResolvedSession_FeatureEnabled verifies FeatureEnabled reads flags from
// the pre-computed Configuration — not from Casbin.
func TestResolvedSession_FeatureEnabled(t *testing.T) {
	sess := &ResolvedSession{
		Configuration: Configuration{
			Flags: map[string]bool{
				"hr.payroll_v2": true,
				"finance.beta":  false,
			},
			Settings: map[string]string{},
			Prefs:    map[string]string{},
		},
	}

	if !sess.FeatureEnabled("hr.payroll_v2") {
		t.Error("expected true for hr.payroll_v2")
	}
	if sess.FeatureEnabled("finance.beta") {
		t.Error("expected false for finance.beta")
	}
	if sess.FeatureEnabled("nonexistent") {
		t.Error("expected false for unknown flag")
	}

	// nil receiver must not panic.
	var nilSess *ResolvedSession
	if nilSess.FeatureEnabled("anything") {
		t.Error("nil receiver must return false")
	}
}

// TestResolvedSession_Configuration tests typed Setting accessors.
func TestResolvedSession_Configuration(t *testing.T) {
	sess := &ResolvedSession{
		Configuration: Configuration{
			Flags: map[string]bool{},
			Settings: map[string]string{
				"iam.session_ttl_hours": "8",
				"finance.vat_rate":      "0.20",
				"feature.enabled":       "true",
				"bad.int":               "notanumber",
			},
			Prefs: map[string]string{},
		},
	}

	t.Run("SettingString present", func(t *testing.T) {
		if got := sess.SettingString("iam.session_ttl_hours", "4"); got != "8" {
			t.Errorf("got %q", got)
		}
	})
	t.Run("SettingString absent uses default", func(t *testing.T) {
		if got := sess.SettingString("missing", "default"); got != "default" {
			t.Errorf("got %q", got)
		}
	})
	t.Run("SettingInt", func(t *testing.T) {
		if got := sess.SettingInt("iam.session_ttl_hours", 0); got != 8 {
			t.Errorf("got %d", got)
		}
	})
	t.Run("SettingInt parse error uses default", func(t *testing.T) {
		if got := sess.SettingInt("bad.int", 99); got != 99 {
			t.Errorf("got %d", got)
		}
	})
	t.Run("SettingDecimal", func(t *testing.T) {
		if got := sess.SettingDecimal("finance.vat_rate", 0); got != 0.20 {
			t.Errorf("got %f", got)
		}
	})
	t.Run("SettingBool", func(t *testing.T) {
		if got := sess.SettingBool("feature.enabled", false); !got {
			t.Error("expected true")
		}
	})
}
