package api

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"awo.so/framework/def"
	"awo.so/framework/platform/org"
)

// ── test doubles ──────────────────────────────────────────────────────────────

type scopeViewer struct {
	orgUnitID uuid.UUID
	tenantID  string
	isSystem  bool
}

func (v *scopeViewer) ActorID() string       { return "test" }
func (v *scopeViewer) TenantID() string      { return v.tenantID }
func (v *scopeViewer) OrgUnitID() uuid.UUID  { return v.orgUnitID }
func (v *scopeViewer) HasRole(_ string) bool { return false }
func (v *scopeViewer) IsSystem() bool        { return v.isSystem }

// scopedRecord implements def.Record + def.OrgScoped.
type scopedRecord struct{ unitID uuid.UUID }

func (r *scopedRecord) Get(_ string) any           { return nil }
func (r *scopedRecord) ID() uuid.UUID              { return uuid.Nil }
func (r *scopedRecord) TenantID() uuid.UUID        { return uuid.Nil }
func (r *scopedRecord) EntityName() string         { return "test" }
func (r *scopedRecord) RecordOrgUnitID() uuid.UUID { return r.unitID }

// unscopedRecord implements def.Record only (no OrgScoped).
type unscopedRecord struct{}

func (r *unscopedRecord) Get(_ string) any    { return nil }
func (r *unscopedRecord) ID() uuid.UUID       { return uuid.Nil }
func (r *unscopedRecord) TenantID() uuid.UUID { return uuid.Nil }
func (r *unscopedRecord) EntityName() string  { return "test" }

// treeStub implements org.Tree with a fixed IsAncestorOrEqual result.
type treeStub struct {
	// ancestors maps ancestor → {descendants it IS an ancestor of}
	ancestors map[uuid.UUID]map[uuid.UUID]bool
}

func (t *treeStub) IsAncestorOrEqual(_ context.Context, _ uuid.UUID, ancestor, node uuid.UUID) (bool, error) {
	if ancestor == node {
		return true, nil
	}
	return t.ancestors[ancestor][node], nil
}
func (t *treeStub) Ancestors(context.Context, uuid.UUID, uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}
func (t *treeStub) Descendants(context.Context, uuid.UUID, uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}
func (t *treeStub) SubtreePath(context.Context, uuid.UUID, uuid.UUID) (string, error) { return "", nil }
func (t *treeStub) Unit(context.Context, uuid.UUID, uuid.UUID) (*org.Unit, error)     { return nil, nil }
func (t *treeStub) InsertPaths(context.Context, *org.Unit) error                      { return nil }
func (t *treeStub) RebuildPaths(context.Context, *org.Unit) error                     { return nil }

var _ org.Tree = (*treeStub)(nil)

// ── fixtures ──────────────────────────────────────────────────────────────────

var (
	scopeTenantID = uuid.MustParse("11111111-0000-0000-0000-000000000001")
	parent        = uuid.MustParse("22222222-0000-0000-0000-000000000001")
	child         = uuid.MustParse("22222222-0000-0000-0000-000000000002")
	sibling       = uuid.MustParse("22222222-0000-0000-0000-000000000003")

	hierarchy = &treeStub{ancestors: map[uuid.UUID]map[uuid.UUID]bool{
		parent: {child: true},
	}}
)

func unitHandler(tree org.Tree) *Handler {
	return &Handler{
		def:     &def.EntityDefinition{Name: "item", OrgScope: org.ScopeLevelUnit},
		orgTree: tree,
	}
}

// ── assertOrgScope ────────────────────────────────────────────────────────────

func TestAssertOrgScope_SameUnit(t *testing.T) {
	h := unitHandler(hierarchy)
	v := &scopeViewer{tenantID: scopeTenantID.String(), orgUnitID: parent}
	if err := h.assertOrgScope(context.Background(), scopeTenantID, v, &scopedRecord{unitID: parent}); err != nil {
		t.Errorf("same unit: want nil, got %v", err)
	}
}

func TestAssertOrgScope_AncestorAllowed(t *testing.T) {
	h := unitHandler(hierarchy)
	v := &scopeViewer{tenantID: scopeTenantID.String(), orgUnitID: parent}
	if err := h.assertOrgScope(context.Background(), scopeTenantID, v, &scopedRecord{unitID: child}); err != nil {
		t.Errorf("ancestor viewer: want nil, got %v", err)
	}
}

func TestAssertOrgScope_SiblingDenied(t *testing.T) {
	h := unitHandler(hierarchy)
	v := &scopeViewer{tenantID: scopeTenantID.String(), orgUnitID: child}
	if err := h.assertOrgScope(context.Background(), scopeTenantID, v, &scopedRecord{unitID: sibling}); err == nil {
		t.Error("sibling: want error, got nil")
	}
}

func TestAssertOrgScope_TenantWideViewerAllowed(t *testing.T) {
	h := unitHandler(hierarchy)
	v := &scopeViewer{tenantID: scopeTenantID.String(), orgUnitID: uuid.Nil}
	if err := h.assertOrgScope(context.Background(), scopeTenantID, v, &scopedRecord{unitID: sibling}); err != nil {
		t.Errorf("tenant-wide viewer: want nil, got %v", err)
	}
}

func TestAssertOrgScope_SystemViewerAllowed(t *testing.T) {
	h := unitHandler(hierarchy)
	v := &scopeViewer{tenantID: scopeTenantID.String(), orgUnitID: child, isSystem: true}
	if err := h.assertOrgScope(context.Background(), scopeTenantID, v, &scopedRecord{unitID: sibling}); err != nil {
		t.Errorf("system viewer: want nil, got %v", err)
	}
}

func TestAssertOrgScope_NoOrgTreeSkips(t *testing.T) {
	h := &Handler{
		def:     &def.EntityDefinition{Name: "item", OrgScope: org.ScopeLevelUnit},
		orgTree: nil,
	}
	v := &scopeViewer{tenantID: scopeTenantID.String(), orgUnitID: child}
	if err := h.assertOrgScope(context.Background(), scopeTenantID, v, &scopedRecord{unitID: sibling}); err != nil {
		t.Errorf("no tree: want nil, got %v", err)
	}
}

func TestAssertOrgScope_TenantScopedEntitySkips(t *testing.T) {
	h := &Handler{
		def:     &def.EntityDefinition{Name: "user"}, // OrgScope "" → ScopeLevelTenant
		orgTree: hierarchy,
	}
	v := &scopeViewer{tenantID: scopeTenantID.String(), orgUnitID: child}
	if err := h.assertOrgScope(context.Background(), scopeTenantID, v, &scopedRecord{unitID: sibling}); err != nil {
		t.Errorf("tenant-scoped entity: want nil, got %v", err)
	}
}

func TestAssertOrgScope_RecordWithNoUnit(t *testing.T) {
	h := unitHandler(hierarchy)
	v := &scopeViewer{tenantID: scopeTenantID.String(), orgUnitID: child}
	if err := h.assertOrgScope(context.Background(), scopeTenantID, v, &scopedRecord{unitID: uuid.Nil}); err != nil {
		t.Errorf("record with no unit: want nil, got %v", err)
	}
}

func TestAssertOrgScope_RecordNotOrgScoped(t *testing.T) {
	h := unitHandler(hierarchy)
	v := &scopeViewer{tenantID: scopeTenantID.String(), orgUnitID: child}
	if err := h.assertOrgScope(context.Background(), scopeTenantID, v, &unscopedRecord{}); err != nil {
		t.Errorf("non-OrgScoped record: want nil, got %v", err)
	}
}

// ── resolveOrgUnitID ──────────────────────────────────────────────────────────

func TestResolveOrgUnitID_UsesBodyString(t *testing.T) {
	v := &scopeViewer{orgUnitID: parent}
	body := map[string]any{"org_unit_id": child.String()}
	if got := resolveOrgUnitID(body, v); got != child {
		t.Errorf("want %v, got %v", child, got)
	}
}

func TestResolveOrgUnitID_UsesBodyUUID(t *testing.T) {
	v := &scopeViewer{orgUnitID: parent}
	body := map[string]any{"org_unit_id": child}
	if got := resolveOrgUnitID(body, v); got != child {
		t.Errorf("want %v, got %v", child, got)
	}
}

func TestResolveOrgUnitID_DefaultsToViewer(t *testing.T) {
	v := &scopeViewer{orgUnitID: parent}
	if got := resolveOrgUnitID(map[string]any{}, v); got != parent {
		t.Errorf("want viewer org %v, got %v", parent, got)
	}
}

func TestResolveOrgUnitID_InvalidStringFallsBack(t *testing.T) {
	v := &scopeViewer{orgUnitID: parent}
	body := map[string]any{"org_unit_id": "not-a-uuid"}
	if got := resolveOrgUnitID(body, v); got != parent {
		t.Errorf("invalid UUID: want viewer org %v, got %v", parent, got)
	}
}

func TestResolveOrgUnitID_NilUUIDFallsBack(t *testing.T) {
	v := &scopeViewer{orgUnitID: parent}
	body := map[string]any{"org_unit_id": uuid.Nil}
	if got := resolveOrgUnitID(body, v); got != parent {
		t.Errorf("nil uuid: want viewer org %v, got %v", parent, got)
	}
}
