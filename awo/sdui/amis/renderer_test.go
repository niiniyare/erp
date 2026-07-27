package amis_test

import (
	"testing"

	"github.com/google/uuid"

	"awo.so/awo/sdui/amis"
	"awo.so/awo/sdui/expression"
	"awo.so/awo/sdui/renderer"
	"awo.so/awo/sdui/sduictx"
	"awo.so/awo/sdui/widget"
)

// ── Test helpers ──────────────────────────────────────────────────────────────

type stubViewer struct{}

func (stubViewer) TenantID() uuid.UUID       { return uuid.New() }
func (stubViewer) Roles() []string           { return nil }
func (stubViewer) IsPlatformAdmin() bool     { return false }
func (stubViewer) HasPermission(string) bool { return true }

func makeCtx() renderer.RendererContext {
	genCtx, _ := sduictx.NewGeneratorContext(
		uuid.New(), stubViewer{}, "test_entity", sduictx.ViewModeList, amis.RendererID,
	).Build()
	return renderer.RendererContext{GenCtx: genCtx}
}

func render(t *testing.T, n *widget.Node) map[string]any {
	t.Helper()
	r := amis.New()
	out, err := r.Render(n, makeCtx())
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	return out.AMISSchema
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestRenderer_NilRoot(t *testing.T) {
	t.Parallel()
	r := amis.New()
	_, err := r.Render(nil, makeCtx())
	if err == nil {
		t.Error("Render(nil) must return error")
	}
}

func TestRenderer_UnknownNodeKind(t *testing.T) {
	t.Parallel()
	r := amis.New()
	_, err := r.Render(&widget.Node{Kind: "unknown-kind-xyz"}, makeCtx())
	if err == nil {
		t.Error("Render with unknown NodeKind must return error")
	}
}

func TestRenderer_HiddenRootStillErrors(t *testing.T) {
	// A hidden root emits nil schema — the pipeline should have removed it before calling Render.
	// This documents the behaviour: Render on a hidden root returns (empty output, nil error)
	// because hidden == omit.
	t.Parallel()
	r := amis.New()
	out, err := r.Render(&widget.Node{Kind: widget.NodeText, Name: "x", Hidden: true}, makeCtx())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.AMISSchema != nil {
		t.Errorf("expected nil AMISSchema for hidden root, got %v", out.AMISSchema)
	}
}

func TestRenderer_PageType(t *testing.T) {
	t.Parallel()
	out := render(t, &widget.Node{Kind: widget.NodePage, Label: "Invoices"})
	if out["type"] != "page" {
		t.Errorf("type=%v want page", out["type"])
	}
	if out["title"] != "Invoices" {
		t.Errorf("title=%v want Invoices", out["title"])
	}
}

func TestRenderer_FormAPI(t *testing.T) {
	t.Parallel()
	n := &widget.Node{
		Kind: widget.NodeForm,
		DataSource: &widget.DataSource{
			URL:    "/api/v1/finance/invoices",
			Method: "POST",
		},
		Children: []*widget.Node{{Kind: widget.NodeText, Name: "number", Label: "Number"}},
	}
	out := render(t, n)
	if out["type"] != "form" {
		t.Errorf("type=%v want form", out["type"])
	}
	api := out["api"].(map[string]any)
	if api["url"] != "/api/v1/finance/invoices" {
		t.Errorf("api.url=%v", api["url"])
	}
	if api["method"] != "POST" {
		t.Errorf("api.method=%v want POST", api["method"])
	}
}

func TestRenderer_ListColumns(t *testing.T) {
	t.Parallel()
	n := &widget.Node{
		Kind:       widget.NodeList,
		DataSource: &widget.DataSource{URL: "/api/v1/invoices"},
		Children: []*widget.Node{
			{Kind: widget.NodeText, Name: "number", Label: "Number"},
			{Kind: widget.NodeDate, Name: "date", Label: "Date"},
		},
	}
	out := render(t, n)
	if out["type"] != "crud2" {
		t.Errorf("type=%v want crud2", out["type"])
	}
	cols := out["columns"].([]any)
	if len(cols) != 2 {
		t.Errorf("expected 2 columns, got %d", len(cols))
	}
}

func TestRenderer_TextFieldProperties(t *testing.T) {
	t.Parallel()
	out := render(t, &widget.Node{
		Kind:     widget.NodeText,
		Name:     "email",
		Label:    "Email",
		Required: true,
		ReadOnly: true,
	})
	if out["type"] != "input-text" {
		t.Errorf("type=%v want input-text", out["type"])
	}
	if out["name"] != "email" {
		t.Errorf("name=%v", out["name"])
	}
	if out["required"] != true {
		t.Errorf("required=%v want true", out["required"])
	}
	if out["disabled"] != true {
		t.Errorf("disabled=%v want true (ReadOnly→disabled)", out["disabled"])
	}
}

func TestRenderer_SelectValueField_Default(t *testing.T) {
	t.Parallel()
	out := render(t, &widget.Node{
		Kind: widget.NodeSelect,
		Name: "currency_id",
		DataSource: &widget.DataSource{
			URL:        "/api/v1/currencies",
			LabelField: "code",
			// ValueField empty → must default to "id"
		},
	})
	if out["valueField"] != "id" {
		t.Errorf("valueField=%v want id (default)", out["valueField"])
	}
}

func TestRenderer_PropsOverride(t *testing.T) {
	t.Parallel()
	out := render(t, &widget.Node{
		Kind:  widget.NodeText,
		Name:  "code",
		Label: "Code",
		Props: map[string]any{"label": "Custom Label"},
	})
	if out["label"] != "Custom Label" {
		t.Errorf("Props.label should override: got %v", out["label"])
	}
}

func TestRenderer_IDEmitted(t *testing.T) {
	t.Parallel()
	out := render(t, &widget.Node{Kind: widget.NodePage, ID: "invoice-page"})
	if out["id"] != "invoice-page" {
		t.Errorf("id=%v want invoice-page", out["id"])
	}
}

func TestRenderer_HiddenChildOmitted(t *testing.T) {
	t.Parallel()
	out := render(t, &widget.Node{
		Kind: widget.NodePage,
		Children: []*widget.Node{
			{Kind: widget.NodeText, Name: "visible"},
			{Kind: widget.NodeText, Name: "secret", Hidden: true},
		},
	})
	body := out["body"].([]any)
	if len(body) != 1 {
		t.Errorf("body has %d items, want 1 (hidden child must be absent)", len(body))
	}
}

func TestRenderer_XSSSanitized(t *testing.T) {
	t.Parallel()
	out := render(t, &widget.Node{Kind: widget.NodePage, Label: "<script>alert('xss')</script>"})
	title := out["title"].(string)
	if title == "<script>alert('xss')</script>" {
		t.Error("XSS: label not sanitized")
	}
}

func TestRenderer_VisibleOnExpression(t *testing.T) {
	t.Parallel()
	exprNode := expression.Eq(expression.Field("status"), expression.Lit("active"))
	out := render(t, &widget.Node{
		Kind:      widget.NodePage,
		VisibleOn: &widget.ExpressionRef{Expr: exprNode},
	})
	got := out["visibleOn"]
	if got != "data.status === 'active'" {
		t.Errorf("visibleOn=%q want data.status === 'active'", got)
	}
}

func TestRenderer_DateFormat(t *testing.T) {
	t.Parallel()
	out := render(t, &widget.Node{Kind: widget.NodeDate, Name: "date"})
	if out["format"] != "YYYY-MM-DD" {
		t.Errorf("format=%v", out["format"])
	}
}

func TestRenderer_DateTimeFormat(t *testing.T) {
	t.Parallel()
	out := render(t, &widget.Node{Kind: widget.NodeDateTime, Name: "ts"})
	if out["format"] != "YYYY-MM-DDTHH:mm:ssZ" {
		t.Errorf("format=%v", out["format"])
	}
}

func TestRenderer_PageActions(t *testing.T) {
	t.Parallel()
	out := render(t, &widget.Node{
		Kind:  widget.NodePage,
		Label: "Detail",
		Actions: []*widget.ActionNode{
			{Label: "Submit", ActionType: "ajax", Level: "primary", API: "/api/submit", ConfirmText: "Sure?"},
		},
	})
	toolbar := out["toolbar"].([]any)
	if len(toolbar) != 1 {
		t.Fatalf("expected 1 action, got %d", len(toolbar))
	}
	btn := toolbar[0].(map[string]any)
	if btn["label"] != "Submit" {
		t.Errorf("label=%v", btn["label"])
	}
	if btn["confirmText"] != "Sure?" {
		t.Errorf("confirmText=%v", btn["confirmText"])
	}
}

func TestRenderer_TabPaneInsideTabs(t *testing.T) {
	t.Parallel()
	out := render(t, &widget.Node{
		Kind: widget.NodePage,
		Children: []*widget.Node{{
			Kind: widget.NodeTabs,
			Children: []*widget.Node{{
				Kind:     widget.NodeTabPane,
				Label:    "Details",
				Children: []*widget.Node{{Kind: widget.NodeText, Name: "name"}},
			}},
		}},
	})
	body := out["body"].([]any)
	tabs := body[0].(map[string]any)
	if tabs["type"] != "tabs" {
		t.Errorf("type=%v want tabs", tabs["type"])
	}
	tabList := tabs["tabs"].([]any)
	if len(tabList) != 1 {
		t.Fatalf("expected 1 tab, got %d", len(tabList))
	}
	tab := tabList[0].(map[string]any)
	if tab["title"] != "Details" {
		t.Errorf("tab title=%v want Details", tab["title"])
	}
}

func TestRenderer_GridEditable(t *testing.T) {
	t.Parallel()
	out := render(t, &widget.Node{
		Kind:         widget.NodeGrid,
		Name:         "line_items",
		Label:        "Line Items",
		GridEditable: true,
		Children: []*widget.Node{
			{Kind: widget.NodeText, Name: "description", Label: "Description"},
			{Kind: widget.NodeNumber, Name: "quantity", Label: "Qty"},
		},
	})
	if out["type"] != "input-table" {
		t.Errorf("type=%v want input-table", out["type"])
	}
	if out["addable"] != true || out["editable"] != true || out["removable"] != true {
		t.Error("editable grid must have addable/editable/removable=true")
	}
}

func TestRenderer_AllInputKinds(t *testing.T) {
	t.Parallel()
	cases := []struct {
		kind     widget.NodeKind
		wantType string
	}{
		{widget.NodeText, "input-text"},
		{widget.NodeField, "input-text"},
		{widget.NodeTextArea, "textarea"},
		{widget.NodeRichText, "rich-text"},
		{widget.NodeNumber, "input-number"},
		{widget.NodeDate, "input-date"},
		{widget.NodeDateTime, "input-datetime"},
		{widget.NodeSwitch, "switch"},
		{widget.NodeEditor, "json-editor"},
		{widget.NodeColor, "input-color"},
		{widget.NodeFileUpload, "input-file"},
	}
	r := amis.New()
	ctx := makeCtx()
	for _, tc := range cases {
		tc := tc
		t.Run(string(tc.kind), func(t *testing.T) {
			t.Parallel()
			out, err := r.Render(&widget.Node{Kind: tc.kind, Name: "f", Label: "F"}, ctx)
			if err != nil {
				t.Fatalf("Render(%s): %v", tc.kind, err)
			}
			if out.AMISSchema["type"] != tc.wantType {
				t.Errorf("Render(%s): type=%v want %q", tc.kind, out.AMISSchema["type"], tc.wantType)
			}
		})
	}
}

// Ensure _ imports compile when expression package is used in test file.
var _ = expression.Field
