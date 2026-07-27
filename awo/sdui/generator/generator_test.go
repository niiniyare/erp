package generator_test

import (
	"testing"

	"github.com/google/uuid"

	"awo.so/awo/sdui/generator"
	"awo.so/awo/sdui/sduictx"
	"awo.so/awo/sdui/widget"
)

// ── Test helpers ──────────────────────────────────────────────────────────────

type stubViewer struct {
	perms map[string]bool
	admin bool
}

func (s *stubViewer) TenantID() uuid.UUID   { return uuid.New() }
func (s *stubViewer) Roles() []string       { return nil }
func (s *stubViewer) IsPlatformAdmin() bool { return s.admin }
func (s *stubViewer) HasPermission(p string) bool {
	if s.admin {
		return true
	}
	return s.perms[p]
}

func makeCtx(mode sduictx.ViewMode, viewer sduictx.ViewerContext) sduictx.GeneratorContext {
	ctx, _ := sduictx.NewGeneratorContext(
		uuid.New(), viewer, "finance_invoice", mode, "amis",
	).WithSchemaFingerprint("fp1").WithPermFingerprint("pf1").Build()
	return ctx
}

func allPermsViewer() *stubViewer {
	return &stubViewer{admin: true}
}

func noPermsViewer() *stubViewer {
	return &stubViewer{perms: map[string]bool{}}
}

func makeSchema() generator.EntitySchema {
	return generator.EntitySchema{
		Name:        "finance_invoice",
		Title:       "Invoice",
		PluralTitle: "Invoices",
		ListURL:     "/api/v1/finance/invoices",
		CreateURL:   "/api/v1/finance/invoices",
		EditURL:     "/api/v1/finance/invoices/{id}",
		DetailURL:   "/api/v1/finance/invoices/{id}",
		Permissions: map[string]string{
			"create": "finance.invoice.create",
			"read":   "finance.invoice.read",
			"update": "finance.invoice.update",
			"delete": "finance.invoice.delete",
		},
		Fields: []generator.FieldDef{
			{Name: "number", Label: "Number", FieldType: "data", InList: true, InForm: true, InDetail: true},
			{Name: "status", Label: "Status", FieldType: "select", InList: true, InForm: true, InDetail: true},
			{Name: "amount", Label: "Amount", FieldType: "currency", InList: true, InForm: true, InDetail: true},
			{Name: "date", Label: "Date", FieldType: "date", InList: true, InForm: true, InDetail: true},
			{Name: "notes", Label: "Notes", FieldType: "long_text", InForm: true, InDetail: true},
			{Name: "secret", Label: "Secret", FieldType: "data", Permission: "finance.invoice.admin",
				InList: true, InForm: true, InDetail: true},
		},
	}
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestGenerator_ListView(t *testing.T) {
	g := generator.New()
	ctx := makeCtx(sduictx.ViewModeList, allPermsViewer())
	schema := makeSchema()

	root, err := g.Generate(schema, ctx)
	if err != nil {
		t.Fatal(err)
	}
	if root == nil {
		t.Fatal("expected non-nil root")
	}
	if root.Kind != widget.NodePage {
		t.Errorf("expected NodePage root, got %q", root.Kind)
	}
	if len(root.Children) == 0 {
		t.Fatal("expected children on page")
	}
	listNode := root.Children[0]
	if listNode.Kind != widget.NodeList {
		t.Errorf("expected NodeList child, got %q", listNode.Kind)
	}
	// Should have 5 columns: number, status, amount, date, secret.
	// notes has InList=false (textarea fields don't appear in list view).
	if len(listNode.Children) != 5 {
		t.Errorf("expected 5 columns, got %d", len(listNode.Children))
	}
}

func TestGenerator_ListView_PermissionGating(t *testing.T) {
	g := generator.New()
	// Viewer has no permissions — secret field should be absent.
	ctx := makeCtx(sduictx.ViewModeList, noPermsViewer())
	root, err := g.Generate(makeSchema(), ctx)
	if err != nil {
		t.Fatal(err)
	}
	listNode := root.Children[0]
	// secret field is absent — 4 visible columns (number, status, amount, date).
	if len(listNode.Children) != 4 {
		t.Errorf("expected 4 columns (secret absent), got %d", len(listNode.Children))
	}
	// Verify secret is not in the list.
	for _, col := range listNode.Children {
		if col.Name == "secret" {
			t.Error("secret field should be absent for viewer without permission")
		}
	}
}

func TestGenerator_FormView_Create(t *testing.T) {
	g := generator.New()
	ctx := makeCtx(sduictx.ViewModeCreate, allPermsViewer())
	root, err := g.Generate(makeSchema(), ctx)
	if err != nil {
		t.Fatal(err)
	}
	if root.Kind != widget.NodePage {
		t.Errorf("expected NodePage, got %q", root.Kind)
	}
	formNode := root.Children[0]
	if formNode.Kind != widget.NodeForm {
		t.Errorf("expected NodeForm child, got %q", formNode.Kind)
	}
	if formNode.DataSource == nil {
		t.Fatal("expected DataSource on form")
	}
	if formNode.DataSource.Method != "POST" {
		t.Errorf("expected POST method for create, got %q", formNode.DataSource.Method)
	}
}

func TestGenerator_FormView_Edit(t *testing.T) {
	g := generator.New()
	ctx := makeCtx(sduictx.ViewModeEdit, allPermsViewer())
	root, err := g.Generate(makeSchema(), ctx)
	if err != nil {
		t.Fatal(err)
	}
	formNode := root.Children[0]
	if formNode.DataSource.Method != "PATCH" {
		t.Errorf("expected PATCH method for edit, got %q", formNode.DataSource.Method)
	}
	// Edit form should have initApi (ReadURL set).
	if formNode.DataSource.ReadURL == "" {
		t.Error("expected ReadURL set for edit form (initApi)")
	}
}

func TestGenerator_DetailView(t *testing.T) {
	g := generator.New()
	ctx := makeCtx(sduictx.ViewModeDetail, allPermsViewer())
	root, err := g.Generate(makeSchema(), ctx)
	if err != nil {
		t.Fatal(err)
	}
	// Detail page should have NodeSummaryCard + NodeForm.
	if len(root.Children) < 2 {
		t.Fatalf("expected at least 2 children on detail page, got %d", len(root.Children))
	}
	summaryCard := root.Children[0]
	if summaryCard.Kind != widget.NodeSummaryCard {
		t.Errorf("expected NodeSummaryCard first, got %q", summaryCard.Kind)
	}
}

func TestGenerator_DetailView_ReadOnly(t *testing.T) {
	g := generator.New()
	ctx := makeCtx(sduictx.ViewModeDetail, allPermsViewer())
	root, err := g.Generate(makeSchema(), ctx)
	if err != nil {
		t.Fatal(err)
	}
	// All field nodes on detail form must be ReadOnly=true.
	formNode := root.Children[1] // [0]=summaryCard, [1]=form
	widget.WalkAll(formNode, func(n *widget.Node) {
		if n.Kind == widget.NodeText || n.Kind == widget.NodeNumber || n.Kind == widget.NodeSelect {
			if !n.ReadOnly {
				t.Errorf("field %q on detail view must be ReadOnly", n.Name)
			}
		}
	})
}

func TestGenerator_UnknownViewMode(t *testing.T) {
	g := generator.New()
	ctx, _ := sduictx.NewGeneratorContext(
		uuid.New(), allPermsViewer(), "test", sduictx.ViewMode("invalid_mode"), "amis",
	).Build()
	_, err := g.Generate(makeSchema(), ctx)
	if err == nil {
		t.Error("expected error for unknown ViewMode")
	}
}

func TestGenerator_ListActions_CreateButton(t *testing.T) {
	g := generator.New()
	// Viewer has create permission.
	viewer := &stubViewer{perms: map[string]bool{"finance.invoice.create": true}}
	ctx := makeCtx(sduictx.ViewModeList, viewer)
	root, err := g.Generate(makeSchema(), ctx)
	if err != nil {
		t.Fatal(err)
	}
	listNode := root.Children[0]
	if len(listNode.Actions) == 0 {
		t.Fatal("expected at least one action (Create button)")
	}
	createBtn := listNode.Actions[0]
	if createBtn.ID != "create" {
		t.Errorf("expected create action, got %q", createBtn.ID)
	}
}

func TestGenerator_DetailActions_NoDeleteWithoutPerm(t *testing.T) {
	g := generator.New()
	// Viewer has update but not delete permission.
	viewer := &stubViewer{perms: map[string]bool{"finance.invoice.update": true}}
	ctx := makeCtx(sduictx.ViewModeDetail, viewer)
	root, err := g.Generate(makeSchema(), ctx)
	if err != nil {
		t.Fatal(err)
	}
	summaryCard := root.Children[0]
	for _, action := range summaryCard.Actions {
		if action.ID == "delete" {
			t.Error("delete action should be absent for viewer without delete permission")
		}
	}
}

func TestGenerator_FieldTypeMapping(t *testing.T) {
	g := generator.New()
	ctx := makeCtx(sduictx.ViewModeCreate, allPermsViewer())
	schema := generator.EntitySchema{
		Name:      "test",
		Title:     "Test",
		CreateURL: "/api/test",
		Fields: []generator.FieldDef{
			{Name: "f1", Label: "Text", FieldType: "data", InForm: true},
			{Name: "f2", Label: "Number", FieldType: "int", InForm: true},
			{Name: "f3", Label: "Date", FieldType: "date", InForm: true},
			{Name: "f4", Label: "Bool", FieldType: "bool", InForm: true},
			{Name: "f5", Label: "Select", FieldType: "select", InForm: true},
			{Name: "f6", Label: "Long", FieldType: "long_text", InForm: true},
		},
	}
	root, err := g.Generate(schema, ctx)
	if err != nil {
		t.Fatal(err)
	}
	formNode := root.Children[0]
	kinds := make(map[string]widget.NodeKind)
	widget.WalkAll(formNode, func(n *widget.Node) {
		if n.Name != "" {
			kinds[n.Name] = n.Kind
		}
	})
	expected := map[string]widget.NodeKind{
		"f1": widget.NodeText,
		"f2": widget.NodeNumber,
		"f3": widget.NodeDate,
		"f4": widget.NodeSwitch,
		"f5": widget.NodeSelect,
		"f6": widget.NodeTextArea,
	}
	for name, want := range expected {
		if got := kinds[name]; got != want {
			t.Errorf("field %q: kind=%q want %q", name, got, want)
		}
	}
}
