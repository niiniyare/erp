package sdui_test

import (
	"testing"

	"github.com/google/uuid"

	"awo.so/framework/def"
	"awo.so/framework/sdui"
)

// ── Mock viewer ───────────────────────────────────────────────────────────────

type mockViewer struct {
	roles []string
}

var _ def.ViewerContext = (*mockViewer)(nil)

func (m *mockViewer) ActorID() string      { return "user-1" }
func (m *mockViewer) TenantID() string     { return "tenant-1" }
func (m *mockViewer) OrgUnitID() uuid.UUID { return uuid.Nil }
func (m *mockViewer) IsSystem() bool       { return false }
func (m *mockViewer) HasRole(role string) bool {
	for _, r := range m.roles {
		if r == role {
			return true
		}
	}
	return false
}

// ── FilteredOpts tests ────────────────────────────────────────────────────────

func TestFilteredOpts_NilViewer_ExcludesSensitive(t *testing.T) {
	entDef := &def.EntityDefinition{
		Name: "invoice",
		Fields: []*def.FieldDef{
			{Name: "number"},
			{Name: "secret_key", IsSensitive: true},
		},
	}
	excluded := sdui.FilteredOpts(entDef, nil)
	if !excluded["secret_key"] {
		t.Error("want secret_key excluded for nil viewer")
	}
	if excluded["number"] {
		t.Error("want non-sensitive field included")
	}
}

func TestFilteredOpts_AdminBypassesFilter(t *testing.T) {
	entDef := &def.EntityDefinition{
		Name: "invoice",
		Fields: []*def.FieldDef{
			{Name: "secret_key", IsSensitive: true},
		},
	}
	admin := &mockViewer{roles: []string{"role:tenant.admin"}}
	excluded := sdui.FilteredOpts(entDef, admin)
	if len(excluded) != 0 {
		t.Errorf("want no exclusions for tenant.admin, got %v", excluded)
	}
}

func TestFilteredOpts_PlatformAdminBypassesFilter(t *testing.T) {
	entDef := &def.EntityDefinition{
		Name: "invoice",
		Fields: []*def.FieldDef{
			{Name: "secret_key", IsSensitive: true},
		},
	}
	padmin := &mockViewer{roles: []string{"role:platform-admin"}}
	excluded := sdui.FilteredOpts(entDef, padmin)
	if len(excluded) != 0 {
		t.Errorf("want no exclusions for platform-admin, got %v", excluded)
	}
}

func TestFilteredOpts_RegularUserExcludesSensitive(t *testing.T) {
	entDef := &def.EntityDefinition{
		Name: "invoice",
		Fields: []*def.FieldDef{
			{Name: "total", IsSensitive: false},
			{Name: "bank_account", IsSensitive: true},
		},
	}
	user := &mockViewer{roles: []string{"role:finance.viewer"}}
	excluded := sdui.FilteredOpts(entDef, user)
	if !excluded["bank_account"] {
		t.Error("want bank_account excluded for regular user")
	}
	if excluded["total"] {
		t.Error("want non-sensitive total included")
	}
}

// ── FilteredReadOnly tests ─────────────────────────────────────────────────────

func TestFilteredReadOnly_NilViewer(t *testing.T) {
	entDef := &def.EntityDefinition{Name: "invoice"}
	if !sdui.FilteredReadOnly(entDef, nil, def.OpUpdate) {
		t.Error("nil viewer should produce read-only schema")
	}
}

func TestFilteredReadOnly_AdminNotReadOnly(t *testing.T) {
	entDef := &def.EntityDefinition{Name: "invoice"}
	admin := &mockViewer{roles: []string{"role:tenant.admin"}}
	if sdui.FilteredReadOnly(entDef, admin, def.OpUpdate) {
		t.Error("tenant.admin should not produce read-only schema")
	}
}
