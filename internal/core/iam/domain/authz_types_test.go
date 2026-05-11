package domain

import (
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// AZ-TYP-001 — Subject builders
// ---------------------------------------------------------------------------

func TestSubjectBuilders(t *testing.T) {
	id := "abc-123"
	cases := []struct {
		name    string
		got     string
		want    string
	}{
		{"PlatformSubject", PlatformSubject(id), "platform:abc-123"},
		{"TenantSubject", TenantSubject(id), "tenant:abc-123"},
		{"PortalSubject", PortalSubject(id), "portal:abc-123"},
		{"APISubject", APISubject(id), "api:abc-123"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Errorf("got %q, want %q", tc.got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// AZ-TYP-010 — Domain builders
// ---------------------------------------------------------------------------

func TestDomainBuilders(t *testing.T) {
	tid := "tenant-uuid"
	cases := []struct {
		name string
		got  string
		want string
	}{
		{"DomainPlatform constant", DomainPlatform, "_platform_"},
		{"PlatformDomain()", PlatformDomain(), "_platform_"},
		{"TenantDomain", TenantDomain(tid), "tenant-uuid"},
		{"PortalDomain", PortalDomain(tid), "tenant-uuid:portal"},
		{"APIDomain", APIDomain(tid), "tenant-uuid:api"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Errorf("got %q, want %q", tc.got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// AZ-TYP-020 — AssignOpts functional options
// ---------------------------------------------------------------------------

func TestAssignOpts(t *testing.T) {
	t.Run("WithAssignedBy", func(t *testing.T) {
		ao := ApplyAssignOpts([]AssignOpt{WithAssignedBy("tenant:user-1")})
		if ao.AssignedBy != "tenant:user-1" {
			t.Errorf("got %q, want %q", ao.AssignedBy, "tenant:user-1")
		}
	})

	t.Run("WithDelegatedBy", func(t *testing.T) {
		ao := ApplyAssignOpts([]AssignOpt{WithDelegatedBy("platform:admin")})
		if ao.DelegatedBy != "platform:admin" {
			t.Errorf("got %q", ao.DelegatedBy)
		}
	})

	t.Run("WithExpiry", func(t *testing.T) {
		exp := time.Now().Add(24 * time.Hour)
		ao := ApplyAssignOpts([]AssignOpt{WithExpiry(exp)})
		if ao.ExpiresAt == nil || !ao.ExpiresAt.Equal(exp) {
			t.Errorf("expiry not set correctly")
		}
	})

	t.Run("multiple opts", func(t *testing.T) {
		exp := time.Now().Add(time.Hour)
		ao := ApplyAssignOpts([]AssignOpt{
			WithAssignedBy("platform:system"),
			WithDelegatedBy("platform:admin"),
			WithExpiry(exp),
		})
		if ao.AssignedBy != "platform:system" {
			t.Errorf("AssignedBy: got %q", ao.AssignedBy)
		}
		if ao.DelegatedBy != "platform:admin" {
			t.Errorf("DelegatedBy: got %q", ao.DelegatedBy)
		}
		if ao.ExpiresAt == nil {
			t.Error("ExpiresAt nil")
		}
	})

	t.Run("zero opts returns empty struct", func(t *testing.T) {
		ao := ApplyAssignOpts(nil)
		if ao.AssignedBy != "" || ao.DelegatedBy != "" || ao.ExpiresAt != nil {
			t.Error("expected zero value AssignOpts")
		}
	})
}

// ---------------------------------------------------------------------------
// AZ-TYP-030 / AZ-TYP-031 — Error type
// ---------------------------------------------------------------------------

func TestErrorString(t *testing.T) {
	e := &Error{Code: "TEST_CODE", Message: "something failed", HTTPStatus: 400}
	got := e.Error()
	want := "[iam] TEST_CODE: something failed"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSentinelErrors(t *testing.T) {
	cases := []struct {
		name       string
		err        *Error
		wantStatus int
	}{
		{"ErrForbidden", ErrForbidden, 403},
		{"ErrUnauthorized", ErrUnauthorized, 401},
		{"ErrInvalidRequest", ErrInvalidRequest, 400},
		{"ErrPolicyConflict", ErrPolicyConflict, 409},
		{"ErrPolicyLimitExceeded", ErrPolicyLimitExceeded, 429},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err.HTTPStatus != tc.wantStatus {
				t.Errorf("HTTPStatus: got %d, want %d", tc.err.HTTPStatus, tc.wantStatus)
			}
			if tc.err.Error() == "" {
				t.Error("Error() must not be empty")
			}
		})
	}
}
