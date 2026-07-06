package privacy_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"awo.so/framework/def"
	"awo.so/framework/platform/org"
	"awo.so/framework/privacy"
)

// ── test doubles ──────────────────────────────────────────────────────────────

type testViewer struct {
	orgUnitID uuid.UUID
	tenantID  string
	isSystem  bool
}

func (v *testViewer) ActorID() string       { return "test-actor" }
func (v *testViewer) TenantID() string      { return v.tenantID }
func (v *testViewer) OrgUnitID() uuid.UUID  { return v.orgUnitID }
func (v *testViewer) HasRole(_ string) bool { return false }
func (v *testViewer) IsSystem() bool        { return v.isSystem }

// orgRecord implements def.Record and def.OrgScoped.
type orgRecord struct{ unitID uuid.UUID }

func (r *orgRecord) Get(_ string) any           { return nil }
func (r *orgRecord) ID() uuid.UUID              { return uuid.Nil }
func (r *orgRecord) TenantID() uuid.UUID        { return uuid.Nil }
func (r *orgRecord) EntityName() string         { return "test" }
func (r *orgRecord) RecordOrgUnitID() uuid.UUID { return r.unitID }

// plainRecord implements def.Record but NOT def.OrgScoped.
type plainRecord struct{}

func (r *plainRecord) Get(_ string) any    { return nil }
func (r *plainRecord) ID() uuid.UUID       { return uuid.Nil }
func (r *plainRecord) TenantID() uuid.UUID { return uuid.Nil }
func (r *plainRecord) EntityName() string  { return "test" }

// mockTree implements org.Tree. Only IsAncestorOrEqual is exercised here.
type mockTree struct {
	// edges[ancestor] = set of descendants (not including ancestor itself)
	edges map[uuid.UUID][]uuid.UUID
	err   error
}

func (t *mockTree) IsAncestorOrEqual(_ context.Context, _ uuid.UUID, ancestor, node uuid.UUID) (bool, error) {
	if t.err != nil {
		return false, t.err
	}
	if ancestor == node {
		return true, nil
	}
	for _, d := range t.edges[ancestor] {
		if d == node {
			return true, nil
		}
	}
	return false, nil
}

func (t *mockTree) Ancestors(context.Context, uuid.UUID, uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}

func (t *mockTree) Descendants(context.Context, uuid.UUID, uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}
func (t *mockTree) SubtreePath(context.Context, uuid.UUID, uuid.UUID) (string, error) { return "", nil }
func (t *mockTree) Unit(context.Context, uuid.UUID, uuid.UUID) (*org.Unit, error)     { return nil, nil }
func (t *mockTree) InsertPaths(context.Context, *org.Unit) error                      { return nil }
func (t *mockTree) RebuildPaths(context.Context, *org.Unit) error                     { return nil }

var _ org.Tree = (*mockTree)(nil)

// ── fixtures ──────────────────────────────────────────────────────────────────

var (
	testTenantID = "aaaaaaaa-0000-0000-0000-000000000001"
	unitParent   = uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000001")
	unitChild    = uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000002")
	unitSibling  = uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000003")

	// tree: unitParent → unitChild; unitSibling is unrelated
	tree = &mockTree{edges: map[uuid.UUID][]uuid.UUID{
		unitParent: {unitChild},
	}}
)

// ── AllowWithinOrgScope tests ─────────────────────────────────────────────────

func TestAllowWithinOrgScope_SameUnit(t *testing.T) {
	fn := privacy.AllowWithinOrgScope(tree)
	viewer := &testViewer{tenantID: testTenantID, orgUnitID: unitParent}
	rec := &orgRecord{unitID: unitParent}
	err := fn(context.Background(), viewer, def.OpRead, rec)
	if !errors.Is(err, def.ErrAllow) {
		t.Errorf("same unit: want ErrAllow, got %v", err)
	}
}

func TestAllowWithinOrgScope_AncestorCanRead(t *testing.T) {
	fn := privacy.AllowWithinOrgScope(tree)
	viewer := &testViewer{tenantID: testTenantID, orgUnitID: unitParent}
	rec := &orgRecord{unitID: unitChild}
	err := fn(context.Background(), viewer, def.OpRead, rec)
	if !errors.Is(err, def.ErrAllow) {
		t.Errorf("ancestor viewer: want ErrAllow, got %v", err)
	}
}

func TestAllowWithinOrgScope_SiblingDenied(t *testing.T) {
	fn := privacy.AllowWithinOrgScope(tree)
	viewer := &testViewer{tenantID: testTenantID, orgUnitID: unitChild}
	rec := &orgRecord{unitID: unitSibling}
	err := fn(context.Background(), viewer, def.OpRead, rec)
	if !errors.Is(err, def.ErrDeny) {
		t.Errorf("sibling access: want ErrDeny, got %v", err)
	}
}

func TestAllowWithinOrgScope_TenantWideViewerAllowed(t *testing.T) {
	fn := privacy.AllowWithinOrgScope(tree)
	viewer := &testViewer{tenantID: testTenantID, orgUnitID: uuid.Nil}
	rec := &orgRecord{unitID: unitChild}
	err := fn(context.Background(), viewer, def.OpRead, rec)
	if !errors.Is(err, def.ErrAllow) {
		t.Errorf("tenant-wide viewer: want ErrAllow, got %v", err)
	}
}

func TestAllowWithinOrgScope_SystemViewerAllowed(t *testing.T) {
	fn := privacy.AllowWithinOrgScope(tree)
	viewer := &testViewer{tenantID: testTenantID, orgUnitID: unitChild, isSystem: true}
	rec := &orgRecord{unitID: unitSibling}
	err := fn(context.Background(), viewer, def.OpRead, rec)
	if !errors.Is(err, def.ErrAllow) {
		t.Errorf("system viewer: want ErrAllow, got %v", err)
	}
}

func TestAllowWithinOrgScope_NilRecordSkips(t *testing.T) {
	fn := privacy.AllowWithinOrgScope(tree)
	viewer := &testViewer{tenantID: testTenantID, orgUnitID: unitParent}
	err := fn(context.Background(), viewer, def.OpRead, nil)
	if !errors.Is(err, def.ErrSkip) {
		t.Errorf("nil record: want ErrSkip, got %v", err)
	}
}

func TestAllowWithinOrgScope_RecordWithNoUnitSkips(t *testing.T) {
	fn := privacy.AllowWithinOrgScope(tree)
	viewer := &testViewer{tenantID: testTenantID, orgUnitID: unitParent}
	rec := &orgRecord{unitID: uuid.Nil}
	err := fn(context.Background(), viewer, def.OpRead, rec)
	if !errors.Is(err, def.ErrSkip) {
		t.Errorf("record with no unit: want ErrSkip, got %v", err)
	}
}

func TestAllowWithinOrgScope_NonOrgScopedRecordSkips(t *testing.T) {
	fn := privacy.AllowWithinOrgScope(tree)
	viewer := &testViewer{tenantID: testTenantID, orgUnitID: unitParent}
	rec := &plainRecord{} // does not implement OrgScoped
	err := fn(context.Background(), viewer, def.OpRead, rec)
	if !errors.Is(err, def.ErrSkip) {
		t.Errorf("non-OrgScoped record: want ErrSkip, got %v", err)
	}
}

func TestAllowWithinOrgScope_TreeErrorPropagates(t *testing.T) {
	boom := errors.New("tree lookup failed")
	errTree := &mockTree{err: boom}
	fn := privacy.AllowWithinOrgScope(errTree)
	viewer := &testViewer{tenantID: testTenantID, orgUnitID: unitChild}
	rec := &orgRecord{unitID: unitSibling}
	err := fn(context.Background(), viewer, def.OpRead, rec)
	if !errors.Is(err, boom) {
		t.Errorf("tree error: want wrapped boom, got %v", err)
	}
}
