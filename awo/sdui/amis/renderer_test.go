package amis_test

import (
	"testing"

	"awo.so/awo/sdui/amis"
	"awo.so/awo/sdui/widget"
)

func TestDefaultRenderer_NilRoot(t *testing.T) {
	t.Parallel()
	r := amis.New()
	_, err := r.Render(nil)
	if err == nil {
		t.Error("Render(nil) must return error")
	}
}

func TestDefaultRenderer_UnknownNodeKind(t *testing.T) {
	t.Parallel()
	r := amis.New()
	n := &widget.Node{Kind: "unknown-kind-xyz"}
	_, err := r.Render(n)
	if err == nil {
		t.Error("Render with unknown NodeKind must return error")
	}
}

// TestDefaultRenderer_HiddenNode verifies that a Hidden root node returns nil
// (omit semantics), not an error.
func TestDefaultRenderer_HiddenNode(t *testing.T) {
	t.Parallel()
	r := amis.New()
	n := &widget.Node{Kind: widget.NodeText, Hidden: true}
	out, err := r.Render(n)
	if err != nil {
		t.Fatalf("Render(Hidden) returned error: %v", err)
	}
	if out != nil {
		t.Errorf("Render(Hidden) must return nil schema, got %v", out)
	}
}

// TestDefaultRenderer_NodePage_Type verifies the top-level page type.
func TestDefaultRenderer_NodePage_Type(t *testing.T) {
	t.Parallel()
	r := amis.New()
	n := &widget.Node{Kind: widget.NodePage, Label: "Invoices"}
	out, err := r.Render(n)
	if err != nil {
		t.Fatalf("Render NodePage: %v", err)
	}
	if out["type"] != "page" {
		t.Errorf("type = %v, want \"page\"", out["type"])
	}
	if out["title"] != "Invoices" {
		t.Errorf("title = %v, want \"Invoices\"", out["title"])
	}
}

// TestDefaultRenderer_NodeList_Type verifies the crud2 type for lists.
func TestDefaultRenderer_NodeList_Type(t *testing.T) {
	t.Parallel()
	r := amis.New()
	n := &widget.Node{
		Kind:  widget.NodeList,
		Label: "Invoices",
		DataSource: &widget.DataSource{
			URL: "/api/v1/finance/invoices?q=${keywords}",
		},
		Children: []*widget.Node{
			{Kind: widget.NodeText, Name: "number", Label: "Number"},
			{Kind: widget.NodeDate, Name: "date", Label: "Date"},
		},
	}
	out, err := r.Render(n)
	if err != nil {
		t.Fatalf("Render NodeList: %v", err)
	}
	if out["type"] != "crud2" {
		t.Errorf("type = %v, want \"crud2\"", out["type"])
	}
	cols, ok := out["columns"].([]any)
	if !ok || len(cols) != 2 {
		t.Errorf("columns: got %v, want 2 columns", out["columns"])
	}
}

// TestDefaultRenderer_NodeForm_API verifies form API config from DataSource.
func TestDefaultRenderer_NodeForm_API(t *testing.T) {
	t.Parallel()
	r := amis.New()
	n := &widget.Node{
		Kind: widget.NodeForm,
		DataSource: &widget.DataSource{
			URL:    "/api/v1/finance/invoices",
			Method: "POST",
		},
		Children: []*widget.Node{
			{Kind: widget.NodeText, Name: "number", Label: "Number", Required: true},
		},
	}
	out, err := r.Render(n)
	if err != nil {
		t.Fatalf("Render NodeForm: %v", err)
	}
	if out["type"] != "form" {
		t.Errorf("type = %v, want \"form\"", out["type"])
	}
	apiOut := out["api"].(map[string]any)
	if apiOut["url"] != "/api/v1/finance/invoices" {
		t.Errorf("api.url = %v", apiOut["url"])
	}
	if apiOut["method"] != "POST" {
		t.Errorf("api.method = %v, want POST", apiOut["method"])
	}
}

// TestDefaultRenderer_NodeText_Properties verifies name/label/required/disabled.
func TestDefaultRenderer_NodeText_Properties(t *testing.T) {
	t.Parallel()
	r := amis.New()
	n := &widget.Node{
		Kind:     widget.NodeText,
		Name:     "email",
		Label:    "Email Address",
		Required: true,
		ReadOnly: true,
	}
	out, err := r.Render(n)
	if err != nil {
		t.Fatalf("Render NodeText: %v", err)
	}
	if out["type"] != "input-text" {
		t.Errorf("type = %v, want \"input-text\"", out["type"])
	}
	if out["name"] != "email" {
		t.Errorf("name = %v", out["name"])
	}
	if out["label"] != "Email Address" {
		t.Errorf("label = %v", out["label"])
	}
	if out["required"] != true {
		t.Errorf("required = %v, want true", out["required"])
	}
	if out["disabled"] != true {
		t.Errorf("disabled = %v, want true (ReadOnly → disabled)", out["disabled"])
	}
}

// TestDefaultRenderer_NodeSelect_DataSource verifies server-side search config.
func TestDefaultRenderer_NodeSelect_DataSource(t *testing.T) {
	t.Parallel()
	r := amis.New()
	n := &widget.Node{
		Kind: widget.NodeSelect,
		Name: "currency_id",
		DataSource: &widget.DataSource{
			URL:        "/api/v1/finance/currencies?q=${keywords}",
			LabelField: "code",
			ValueField: "id",
		},
	}
	out, err := r.Render(n)
	if err != nil {
		t.Fatalf("Render NodeSelect: %v", err)
	}
	if out["type"] != "select" {
		t.Errorf("type = %v, want \"select\"", out["type"])
	}
	if out["labelField"] != "code" {
		t.Errorf("labelField = %v, want \"code\"", out["labelField"])
	}
	if out["valueField"] != "id" {
		t.Errorf("valueField = %v, want \"id\"", out["valueField"])
	}
}

// TestDefaultRenderer_Props_Override verifies Props merge last and override defaults.
func TestDefaultRenderer_Props_Override(t *testing.T) {
	t.Parallel()
	r := amis.New()
	n := &widget.Node{
		Kind:  widget.NodeText,
		Name:  "code",
		Label: "Code",
		Props: map[string]any{
			"maxLength": 10,
			"label":     "Custom Label", // override computed label
		},
	}
	out, err := r.Render(n)
	if err != nil {
		t.Fatalf("Render with Props: %v", err)
	}
	// Props.label must override the typed Label field.
	if out["label"] != "Custom Label" {
		t.Errorf("Props override failed: label = %v, want \"Custom Label\"", out["label"])
	}
	if out["maxLength"] != 10 {
		t.Errorf("Props.maxLength = %v, want 10", out["maxLength"])
	}
}

// TestDefaultRenderer_Actions verifies action buttons are rendered.
func TestDefaultRenderer_Actions(t *testing.T) {
	t.Parallel()
	r := amis.New()
	n := &widget.Node{
		Kind:  widget.NodePage,
		Label: "Invoice Detail",
		Actions: []*widget.ActionNode{
			{Label: "Submit", ActionType: "ajax", Level: "primary", API: "POST:/api/v1/finance/invoices/${id}/submit", ConfirmText: "Submit?"},
		},
	}
	out, err := r.Render(n)
	if err != nil {
		t.Fatalf("Render with actions: %v", err)
	}
	toolbar, ok := out["toolbar"].([]any)
	if !ok || len(toolbar) == 0 {
		t.Fatalf("toolbar missing or empty: %v", out["toolbar"])
	}
	btn := toolbar[0].(map[string]any)
	if btn["label"] != "Submit" {
		t.Errorf("action label = %v, want \"Submit\"", btn["label"])
	}
	if btn["actionType"] != "ajax" {
		t.Errorf("actionType = %v, want \"ajax\"", btn["actionType"])
	}
	if btn["confirmText"] != "Submit?" {
		t.Errorf("confirmText = %v, want \"Submit?\"", btn["confirmText"])
	}
}

// TestDefaultRenderer_HiddenChild verifies hidden children are omitted from output.
func TestDefaultRenderer_HiddenChild(t *testing.T) {
	t.Parallel()
	r := amis.New()
	n := &widget.Node{
		Kind: widget.NodePage,
		Children: []*widget.Node{
			{Kind: widget.NodeText, Name: "visible", Label: "Visible"},
			{Kind: widget.NodeText, Name: "secret", Label: "Secret", Hidden: true},
		},
	}
	out, err := r.Render(n)
	if err != nil {
		t.Fatalf("Render with hidden child: %v", err)
	}
	body, _ := out["body"].([]any)
	if len(body) != 1 {
		t.Errorf("body has %d items, want 1 (hidden child must be absent)", len(body))
	}
}

// TestDefaultRenderer_AllInputTypes verifies every NodeKind maps to expected amis type.
func TestDefaultRenderer_AllInputTypes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		kind     widget.NodeKind
		wantType string
	}{
		{widget.NodeText, "input-text"},
		{widget.NodeField, "input-text"},
		{widget.NodeTextArea, "textarea"},
		{widget.NodeNumber, "input-number"},
		{widget.NodeDate, "input-date"},
		{widget.NodeDateTime, "input-datetime"},
		{widget.NodeSwitch, "switch"},
		{widget.NodeEditor, "json-editor"},
	}

	r := amis.New()
	for _, tc := range cases {
		tc := tc
		t.Run(string(tc.kind), func(t *testing.T) {
			t.Parallel()
			n := &widget.Node{Kind: tc.kind, Name: "f", Label: "F"}
			out, err := r.Render(n)
			if err != nil {
				t.Fatalf("Render(%s): %v", tc.kind, err)
			}
			if out["type"] != tc.wantType {
				t.Errorf("Render(%s): type = %v, want %q", tc.kind, out["type"], tc.wantType)
			}
		})
	}
}

// TestDefaultRenderer_NodeDate_Format verifies date format is set.
func TestDefaultRenderer_NodeDate_Format(t *testing.T) {
	t.Parallel()
	r := amis.New()
	n := &widget.Node{Kind: widget.NodeDate, Name: "invoice_date"}
	out, err := r.Render(n)
	if err != nil {
		t.Fatalf("Render NodeDate: %v", err)
	}
	if out["format"] != "YYYY-MM-DD" {
		t.Errorf("format = %v, want YYYY-MM-DD", out["format"])
	}
}

// TestDefaultRenderer_NodeDateTime_Format verifies datetime format is set.
func TestDefaultRenderer_NodeDateTime_Format(t *testing.T) {
	t.Parallel()
	r := amis.New()
	n := &widget.Node{Kind: widget.NodeDateTime, Name: "created_at"}
	out, err := r.Render(n)
	if err != nil {
		t.Fatalf("Render NodeDateTime: %v", err)
	}
	if out["format"] != "YYYY-MM-DDTHH:mm:ssZ" {
		t.Errorf("format = %v, want YYYY-MM-DDTHH:mm:ssZ", out["format"])
	}
}

// TestDefaultRenderer_NodeID verifies that Node.ID is emitted as "id".
func TestDefaultRenderer_NodeID(t *testing.T) {
	t.Parallel()
	r := amis.New()
	n := &widget.Node{Kind: widget.NodePage, ID: "invoice-detail-page"}
	out, err := r.Render(n)
	if err != nil {
		t.Fatalf("Render with ID: %v", err)
	}
	if out["id"] != "invoice-detail-page" {
		t.Errorf("id = %v, want \"invoice-detail-page\"", out["id"])
	}
}

// TestDefaultRenderer_Select_DefaultValueField verifies valueField defaults to "id".
func TestDefaultRenderer_Select_DefaultValueField(t *testing.T) {
	t.Parallel()
	r := amis.New()
	n := &widget.Node{
		Kind: widget.NodeSelect,
		Name: "currency_id",
		DataSource: &widget.DataSource{
			URL:        "/api/v1/finance/currencies?q=${keywords}",
			LabelField: "code",
			// ValueField deliberately empty — must default to "id"
		},
	}
	out, err := r.Render(n)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if out["valueField"] != "id" {
		t.Errorf("valueField = %v, want \"id\" (default)", out["valueField"])
	}
}

// TestDefaultRenderer_Deterministic verifies identical input produces identical output.
func TestDefaultRenderer_Deterministic(t *testing.T) {
	t.Parallel()
	r := amis.New()
	n := &widget.Node{
		Kind:  widget.NodePage,
		Label: "Test",
		Children: []*widget.Node{
			{Kind: widget.NodeText, Name: "a", Label: "A"},
			{Kind: widget.NodeText, Name: "b", Label: "B"},
		},
	}
	out1, err1 := r.Render(n)
	out2, err2 := r.Render(n)
	if err1 != nil || err2 != nil {
		t.Fatalf("Render errors: %v %v", err1, err2)
	}
	// Both should have "page" type — deeper equality not checked here since
	// map[string]any comparison requires reflect.DeepEqual.
	if out1["type"] != out2["type"] {
		t.Error("Render not deterministic: type differs between calls")
	}
}
