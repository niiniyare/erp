package sdui_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"awo.so/awo/auth"
	"awo.so/awo/cache"
	"awo.so/awo/compiler"
	"awo.so/awo/def"
	"awo.so/awo/sdui"
	"awo.so/awo/sdui/widget"
)

// ── Test helpers ──────────────────────────────────────────────────────────────

// testSchema builds a minimal CompiledSchema with one entity for use in tests.
func testSchema(es *compiler.EntitySchema) *compiler.CompiledSchema {
	return &compiler.CompiledSchema{
		Entities: []*compiler.EntitySchema{es},
		ByName:   map[string]*compiler.EntitySchema{es.QualifiedName: es},
	}
}

// invoiceSchema returns a realistic EntitySchema for testing all SDUI paths.
func invoiceSchema() *compiler.EntitySchema {
	return &compiler.EntitySchema{
		QualifiedName: "finance_invoice",
		LocalName:     "invoice",
		Module:        "finance",
		Label:         "Invoice",
		LabelPlural:   "Invoices",
		APIResource:   "invoices",
		RoutePrefix:   "/api/v1/finance/invoices",
		Fields: []def.FieldDef{
			{Name: "id", Type: def.FieldTypeData, Hidden: true},
			{Name: "tenant_id", Type: def.FieldTypeData, Hidden: true},
			{Name: "number", Label: "Number", Type: def.FieldTypeNamingSeries, Immutable: true},
			{Name: "customer_id", Label: "Customer", Type: def.FieldTypeLink, LinkTarget: "crm_customer", Required: true},
			{Name: "status", Label: "Status", Type: def.FieldTypeSelect, Options: []string{"Draft", "Submitted", "Paid"}},
			{Name: "total", Label: "Total", Type: def.FieldTypeCurrency, Required: true, Description: "Invoice total in KES"},
			{Name: "notes", Label: "Notes", Type: def.FieldTypeLongText},
			{Name: "internal_ref", Type: def.FieldTypeData, Sensitive: true},
		},
		Actions: []def.ActionDef{
			{Name: "submit", Label: "Submit", Method: def.ActionMethodPost, ConfirmMessage: "Submit this invoice?"},
		},
		Permissions: def.PermissionSet{
			Create: []string{"finance.invoice.create"},
			Read:   []string{"finance.invoice.read"},
			Write:  []string{"finance.invoice.write"},
			Delete: []string{"finance.invoice.delete"},
		},
		FieldLookups: map[string]*compiler.CompiledLookup{
			"customer_id": {
				SearchURL:  "/api/v1/crm/customers?q=${keywords}&tenant_id=${tenant_id}",
				LabelField: "name",
				ValueField: "id",
			},
		},
	}
}

// allowAllEvaluator permits every action.
type allowAllEvaluator struct{}

func (allowAllEvaluator) CanPerform(_ context.Context, _ auth.ViewerContext, _, _ string) (bool, error) {
	return true, nil
}

// denyAllEvaluator denies every action.
type denyAllEvaluator struct{}

func (denyAllEvaluator) CanPerform(_ context.Context, _ auth.ViewerContext, _, _ string) (bool, error) {
	return false, nil
}

// stubViewer is a minimal ViewerContext for tests.
type stubViewer struct {
	roles []string
}

func (v *stubViewer) TenantID() uuid.UUID         { return uuid.Nil }
func (v *stubViewer) UserID() uuid.UUID           { return uuid.Nil }
func (v *stubViewer) ServiceAccountID() uuid.UUID { return uuid.Nil }
func (v *stubViewer) Roles() []string             { return v.roles }
func (v *stubViewer) HasRole(r string) bool {
	for _, role := range v.roles {
		if role == r {
			return true
		}
	}
	return false
}
func (v *stubViewer) IsPlatformAdmin() bool { return false }
func (v *stubViewer) Actor() *def.Actor {
	return &def.Actor{Roles: append([]string(nil), v.roles...)}
}

// countingCache records Set calls without storing anything that causes cache hits.
type countingCache struct {
	sets int
}

func (c *countingCache) Get(_ context.Context, _ string, _ any) error {
	return cache.ErrMiss
}
func (c *countingCache) Set(_ context.Context, _ string, _ any, _ time.Duration) error {
	c.sets++
	return nil
}
func (c *countingCache) Delete(_ context.Context, _ string) error         { return nil }
func (c *countingCache) DeletePrefix(_ context.Context, _ string) error   { return nil }
func (c *countingCache) Exists(_ context.Context, _ string) (bool, error) { return false, nil }

// ── List page tests ───────────────────────────────────────────────────────────

func TestGenerator_ListPage_Type(t *testing.T) {
	t.Parallel()
	gen := sdui.New(testSchema(invoiceSchema()), allowAllEvaluator{}, nil)

	page, err := gen.GetPage(context.Background(), "finance_invoice", def.PageKindList, nil)
	if err != nil {
		t.Fatalf("GetPage list: %v", err)
	}
	if page["type"] != "page" {
		t.Errorf("type = %v, want page", page["type"])
	}
	if page["title"] != "Invoices" {
		t.Errorf("title = %v, want Invoices", page["title"])
	}
}

func TestGenerator_ListPage_HasCrud2(t *testing.T) {
	t.Parallel()
	gen := sdui.New(testSchema(invoiceSchema()), allowAllEvaluator{}, nil)

	page, _ := gen.GetPage(context.Background(), "finance_invoice", def.PageKindList, nil)
	body := page["body"].([]any)
	crud := body[0].(map[string]any)
	if crud["type"] != "crud2" {
		t.Errorf("expected crud2, got: %v", crud["type"])
	}
}

func TestGenerator_ListPage_ExcludesSensitiveField(t *testing.T) {
	t.Parallel()
	gen := sdui.New(testSchema(invoiceSchema()), allowAllEvaluator{}, nil)

	page, _ := gen.GetPage(context.Background(), "finance_invoice", def.PageKindList, nil)
	body := page["body"].([]any)
	crud := body[0].(map[string]any)
	cols := crud["columns"].([]any)
	for _, c := range cols {
		col := c.(map[string]any)
		if col["name"] == "internal_ref" {
			t.Error("sensitive field must not appear in list columns")
		}
	}
}

func TestGenerator_ListPage_ExcludesLongTextField(t *testing.T) {
	t.Parallel()
	gen := sdui.New(testSchema(invoiceSchema()), allowAllEvaluator{}, nil)

	page, _ := gen.GetPage(context.Background(), "finance_invoice", def.PageKindList, nil)
	body := page["body"].([]any)
	crud := body[0].(map[string]any)
	cols := crud["columns"].([]any)
	for _, c := range cols {
		col := c.(map[string]any)
		if col["name"] == "notes" {
			t.Error("LongText field must not appear in list columns")
		}
	}
}

func TestGenerator_ListPage_ToolbarWithCreatePermission(t *testing.T) {
	t.Parallel()
	gen := sdui.New(testSchema(invoiceSchema()), allowAllEvaluator{}, nil)

	page, _ := gen.GetPage(context.Background(), "finance_invoice", def.PageKindList, nil)
	toolbar, ok := page["toolbar"].([]any)
	if !ok || len(toolbar) == 0 {
		t.Fatal("toolbar must have buttons when viewer has create permission")
	}
	btn := toolbar[0].(map[string]any)
	if btn["label"] != "New Invoice" {
		t.Errorf("toolbar label = %v, want 'New Invoice'", btn["label"])
	}
}

func TestGenerator_ListPage_NoToolbarWhenDenied(t *testing.T) {
	t.Parallel()
	gen := sdui.New(testSchema(invoiceSchema()), denyAllEvaluator{}, nil)

	page, _ := gen.GetPage(context.Background(), "finance_invoice", def.PageKindList, &stubViewer{})
	if _, ok := page["toolbar"]; ok {
		t.Error("toolbar must be absent when viewer lacks create permission")
	}
}

func TestGenerator_ListPage_RowActionsWithPermission(t *testing.T) {
	t.Parallel()
	gen := sdui.New(testSchema(invoiceSchema()), allowAllEvaluator{}, nil)

	page, _ := gen.GetPage(context.Background(), "finance_invoice", def.PageKindList, nil)
	body := page["body"].([]any)
	crud := body[0].(map[string]any)
	rowActions, ok := crud["rowActions"].([]any)
	if !ok || len(rowActions) == 0 {
		t.Fatal("rowActions must be present when viewer has permissions")
	}
}

func TestGenerator_ListPage_NoRowActionsWhenDenied(t *testing.T) {
	t.Parallel()
	gen := sdui.New(testSchema(invoiceSchema()), denyAllEvaluator{}, nil)

	page, _ := gen.GetPage(context.Background(), "finance_invoice", def.PageKindList, &stubViewer{})
	body := page["body"].([]any)
	crud := body[0].(map[string]any)
	if _, ok := crud["rowActions"]; ok {
		t.Error("rowActions must be absent when viewer has no permissions")
	}
}

// ── Create page tests ─────────────────────────────────────────────────────────

func TestGenerator_CreatePage_Type(t *testing.T) {
	t.Parallel()
	gen := sdui.New(testSchema(invoiceSchema()), allowAllEvaluator{}, nil)

	page, err := gen.GetPage(context.Background(), "finance_invoice", def.PageKindCreate, nil)
	if err != nil {
		t.Fatalf("GetPage create: %v", err)
	}
	if page["title"] != "Create Invoice" {
		t.Errorf("title = %v, want 'Create Invoice'", page["title"])
	}
}

func TestGenerator_CreatePage_PostAPI(t *testing.T) {
	t.Parallel()
	gen := sdui.New(testSchema(invoiceSchema()), allowAllEvaluator{}, nil)

	page, _ := gen.GetPage(context.Background(), "finance_invoice", def.PageKindCreate, nil)
	body := page["body"].([]any)
	form := body[0].(map[string]any)
	api := form["api"].(map[string]any)
	if api["method"] != "POST" {
		t.Errorf("api.method = %v, want POST", api["method"])
	}
}

func TestGenerator_CreatePage_NoInitApi(t *testing.T) {
	t.Parallel()
	gen := sdui.New(testSchema(invoiceSchema()), allowAllEvaluator{}, nil)

	page, _ := gen.GetPage(context.Background(), "finance_invoice", def.PageKindCreate, nil)
	body := page["body"].([]any)
	form := body[0].(map[string]any)
	if _, ok := form["initApi"]; ok {
		t.Error("create form must not have initApi — no existing record to load")
	}
}

func TestGenerator_CreatePage_RequiredField(t *testing.T) {
	t.Parallel()
	gen := sdui.New(testSchema(invoiceSchema()), allowAllEvaluator{}, nil)

	page, _ := gen.GetPage(context.Background(), "finance_invoice", def.PageKindCreate, nil)
	body := page["body"].([]any)
	form := body[0].(map[string]any)
	fields := form["body"].([]any)

	for _, f := range fields {
		field := f.(map[string]any)
		if field["name"] == "total" {
			if field["required"] != true {
				t.Error("required field 'total' must have required: true")
			}
			return
		}
	}
	t.Error("field 'total' not found in create form")
}

// ── Edit page tests ───────────────────────────────────────────────────────────

func TestGenerator_EditPage_PatchAPI(t *testing.T) {
	t.Parallel()
	gen := sdui.New(testSchema(invoiceSchema()), allowAllEvaluator{}, nil)

	page, err := gen.GetPage(context.Background(), "finance_invoice", def.PageKindEdit, nil)
	if err != nil {
		t.Fatalf("GetPage edit: %v", err)
	}
	body := page["body"].([]any)
	form := body[0].(map[string]any)
	api := form["api"].(map[string]any)
	if api["method"] != "PATCH" {
		t.Errorf("api.method = %v, want PATCH", api["method"])
	}
}

func TestGenerator_EditPage_HasInitApi(t *testing.T) {
	t.Parallel()
	gen := sdui.New(testSchema(invoiceSchema()), allowAllEvaluator{}, nil)

	page, _ := gen.GetPage(context.Background(), "finance_invoice", def.PageKindEdit, nil)
	body := page["body"].([]any)
	form := body[0].(map[string]any)
	initApi, ok := form["initApi"]
	if !ok {
		t.Fatal("edit form must have initApi to pre-populate with existing record")
	}
	if initApi == "" {
		t.Error("initApi must not be empty")
	}
}

func TestGenerator_EditPage_InitApiNoDuplicateID(t *testing.T) {
	t.Parallel()
	gen := sdui.New(testSchema(invoiceSchema()), allowAllEvaluator{}, nil)

	page, _ := gen.GetPage(context.Background(), "finance_invoice", def.PageKindEdit, nil)
	body := page["body"].([]any)
	form := body[0].(map[string]any)
	initApi := form["initApi"].(string)

	count := 0
	for i := 0; i <= len(initApi)-5; i++ {
		if initApi[i:i+5] == "${id}" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("initApi %q has ${id} %d times, want exactly 1", initApi, count)
	}
}

func TestGenerator_EditPage_ImmutableFieldDisabled(t *testing.T) {
	t.Parallel()
	gen := sdui.New(testSchema(invoiceSchema()), allowAllEvaluator{}, nil)

	page, _ := gen.GetPage(context.Background(), "finance_invoice", def.PageKindEdit, nil)
	body := page["body"].([]any)
	form := body[0].(map[string]any)
	fields := form["body"].([]any)

	for _, f := range fields {
		field := f.(map[string]any)
		if field["name"] == "number" {
			if field["disabled"] != true {
				t.Error("NamingSeries field 'number' must be disabled in edit form")
			}
			return
		}
	}
	t.Error("field 'number' not found in edit form")
}

// ── Detail page tests ─────────────────────────────────────────────────────────

func TestGenerator_DetailPage_Type(t *testing.T) {
	t.Parallel()
	gen := sdui.New(testSchema(invoiceSchema()), allowAllEvaluator{}, nil)

	page, err := gen.GetPage(context.Background(), "finance_invoice", def.PageKindDetail, nil)
	if err != nil {
		t.Fatalf("GetPage detail: %v", err)
	}
	if page["type"] != "page" {
		t.Errorf("type = %v, want page", page["type"])
	}
}

func TestGenerator_DetailPage_HasInitApi(t *testing.T) {
	t.Parallel()
	gen := sdui.New(testSchema(invoiceSchema()), allowAllEvaluator{}, nil)

	page, _ := gen.GetPage(context.Background(), "finance_invoice", def.PageKindDetail, nil)
	body := page["body"].([]any)
	form := body[0].(map[string]any)
	if _, ok := form["initApi"]; !ok {
		t.Error("detail form must have initApi to load record data")
	}
}

func TestGenerator_DetailPage_NoSubmitAPI(t *testing.T) {
	t.Parallel()
	gen := sdui.New(testSchema(invoiceSchema()), allowAllEvaluator{}, nil)

	page, _ := gen.GetPage(context.Background(), "finance_invoice", def.PageKindDetail, nil)
	body := page["body"].([]any)
	form := body[0].(map[string]any)
	if _, ok := form["api"]; ok {
		t.Error("detail form must not have submit api — read-only view")
	}
}

func TestGenerator_DetailPage_AllFieldsReadOnly(t *testing.T) {
	t.Parallel()
	gen := sdui.New(testSchema(invoiceSchema()), allowAllEvaluator{}, nil)

	page, _ := gen.GetPage(context.Background(), "finance_invoice", def.PageKindDetail, nil)
	body := page["body"].([]any)
	form := body[0].(map[string]any)
	fields := form["body"].([]any)
	if len(fields) == 0 {
		t.Fatal("detail form body is empty")
	}
	for _, f := range fields {
		field := f.(map[string]any)
		if field["disabled"] != true {
			t.Errorf("detail field %q must be disabled (read-only)", field["name"])
		}
	}
}

func TestGenerator_DetailPage_EditActionPresent(t *testing.T) {
	t.Parallel()
	gen := sdui.New(testSchema(invoiceSchema()), allowAllEvaluator{}, nil)

	page, _ := gen.GetPage(context.Background(), "finance_invoice", def.PageKindDetail, nil)
	toolbar, ok := page["toolbar"].([]any)
	if !ok || len(toolbar) == 0 {
		t.Fatal("detail page toolbar must have actions when viewer has update permission")
	}
	var hasEdit bool
	for _, a := range toolbar {
		btn := a.(map[string]any)
		if btn["label"] == "Edit" {
			hasEdit = true
		}
	}
	if !hasEdit {
		t.Error("detail toolbar must contain Edit button when update permission granted")
	}
}

// ── Field mapping tests ───────────────────────────────────────────────────────

func TestGenerator_FieldMapping_CurrencyPrecision(t *testing.T) {
	t.Parallel()
	gen := sdui.New(testSchema(invoiceSchema()), allowAllEvaluator{}, nil)

	page, _ := gen.GetPage(context.Background(), "finance_invoice", def.PageKindCreate, nil)
	body := page["body"].([]any)
	form := body[0].(map[string]any)
	fields := form["body"].([]any)

	for _, f := range fields {
		field := f.(map[string]any)
		if field["name"] == "total" {
			if field["type"] != "input-number" {
				t.Errorf("currency type = %v, want input-number", field["type"])
			}
			if field["precision"] != 4 {
				t.Errorf("currency precision = %v, want 4", field["precision"])
			}
			return
		}
	}
	t.Error("field 'total' not found")
}

func TestGenerator_FieldMapping_SelectOptions(t *testing.T) {
	t.Parallel()
	gen := sdui.New(testSchema(invoiceSchema()), allowAllEvaluator{}, nil)

	page, _ := gen.GetPage(context.Background(), "finance_invoice", def.PageKindCreate, nil)
	body := page["body"].([]any)
	form := body[0].(map[string]any)
	fields := form["body"].([]any)

	for _, f := range fields {
		field := f.(map[string]any)
		if field["name"] == "status" {
			if field["type"] != "select" {
				t.Errorf("select type = %v, want select", field["type"])
			}
			return
		}
	}
	t.Error("field 'status' not found")
}

func TestGenerator_FieldMapping_LinkDataSource(t *testing.T) {
	t.Parallel()
	gen := sdui.New(testSchema(invoiceSchema()), allowAllEvaluator{}, nil)

	page, _ := gen.GetPage(context.Background(), "finance_invoice", def.PageKindCreate, nil)
	body := page["body"].([]any)
	form := body[0].(map[string]any)
	fields := form["body"].([]any)

	for _, f := range fields {
		field := f.(map[string]any)
		if field["name"] == "customer_id" {
			src, ok := field["source"].(map[string]any)
			if !ok {
				t.Fatalf("Link field must have source, got: %v", field["source"])
			}
			if src["url"] == "" {
				t.Error("Link field source URL must not be empty")
			}
			if field["labelField"] != "name" {
				t.Errorf("labelField = %v, want name", field["labelField"])
			}
			if field["valueField"] != "id" {
				t.Errorf("valueField = %v, want id", field["valueField"])
			}
			return
		}
	}
	t.Error("link field 'customer_id' not found")
}

func TestGenerator_FieldMapping_DescriptionPropagated(t *testing.T) {
	t.Parallel()
	gen := sdui.New(testSchema(invoiceSchema()), allowAllEvaluator{}, nil)

	page, _ := gen.GetPage(context.Background(), "finance_invoice", def.PageKindCreate, nil)
	body := page["body"].([]any)
	form := body[0].(map[string]any)
	fields := form["body"].([]any)

	for _, f := range fields {
		field := f.(map[string]any)
		if field["name"] == "total" {
			if field["description"] != "Invoice total in KES" {
				t.Errorf("description = %v, want 'Invoice total in KES'", field["description"])
			}
			return
		}
	}
	t.Error("field 'total' not found")
}

func TestGenerator_FieldMapping_AllFieldTypes(t *testing.T) {
	t.Parallel()
	cases := []struct {
		fieldType def.FieldType
		wantAmis  string
	}{
		{def.FieldTypeData, "input-text"},
		{def.FieldTypeSmallText, "textarea"},
		{def.FieldTypeLongText, "textarea"},
		{def.FieldTypeInt, "input-number"},
		{def.FieldTypeFloat, "input-number"},
		{def.FieldTypeCurrency, "input-number"},
		{def.FieldTypeBool, "switch"},
		{def.FieldTypeDate, "input-date"},
		{def.FieldTypeDateTime, "input-datetime"},
		{def.FieldTypeSelect, "select"},
		{def.FieldTypeMultiSelect, "select"},
		{def.FieldTypeJSON, "json-editor"},
		{def.FieldTypeNamingSeries, "input-text"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(string(tc.fieldType), func(t *testing.T) {
			t.Parallel()
			es := &compiler.EntitySchema{
				QualifiedName: "test_entity",
				LocalName:     "entity",
				Module:        "test",
				Label:         "Entity",
				LabelPlural:   "Entities",
				APIResource:   "entities",
				RoutePrefix:   "/api/v1/test/entities",
				Fields: []def.FieldDef{
					{Name: "field1", Type: tc.fieldType},
				},
				Permissions: def.PermissionSet{Read: []string{"test.entity.read"}},
			}
			schema := testSchema(es)
			gen := sdui.New(schema, allowAllEvaluator{}, nil)

			page, err := gen.GetPage(context.Background(), "test_entity", def.PageKindCreate, nil)
			if err != nil {
				t.Fatalf("GetPage: %v", err)
			}
			body := page["body"].([]any)
			form := body[0].(map[string]any)
			fields, _ := form["body"].([]any)
			if len(fields) == 0 {
				t.Fatalf("no fields in form for type %s", tc.fieldType)
			}
			field := fields[0].(map[string]any)
			if field["type"] != tc.wantAmis {
				t.Errorf("FieldType %s: amis type = %v, want %v", tc.fieldType, field["type"], tc.wantAmis)
			}
		})
	}
}

// ── Error handling tests ──────────────────────────────────────────────────────

func TestGenerator_UnknownEntity_ReturnsError(t *testing.T) {
	t.Parallel()
	gen := sdui.New(testSchema(invoiceSchema()), nil, nil)

	_, err := gen.GetPage(context.Background(), "nonexistent", def.PageKindList, nil)
	if err == nil {
		t.Error("expected error for unknown entity, got nil")
	}
}

func TestGenerator_UnknownPageKind_ReturnsError(t *testing.T) {
	t.Parallel()
	gen := sdui.New(testSchema(invoiceSchema()), nil, nil)

	_, err := gen.GetPage(context.Background(), "finance_invoice", def.PageKind("bogus"), nil)
	if err == nil {
		t.Error("expected error for unknown PageKind, got nil")
	}
}

// ── Cache behaviour tests ─────────────────────────────────────────────────────

func TestGenerator_Cache_WritesOnMiss(t *testing.T) {
	t.Parallel()
	c := &countingCache{}
	gen := sdui.New(testSchema(invoiceSchema()), allowAllEvaluator{}, c)

	_, err := gen.GetPage(context.Background(), "finance_invoice", def.PageKindList, nil)
	if err != nil {
		t.Fatalf("GetPage: %v", err)
	}
	if c.sets == 0 {
		t.Error("expected at least one cache Set on miss")
	}
}

func TestGenerator_Cache_NilCacheNoError(t *testing.T) {
	t.Parallel()
	gen := sdui.New(testSchema(invoiceSchema()), allowAllEvaluator{}, nil)

	_, err := gen.GetPage(context.Background(), "finance_invoice", def.PageKindList, nil)
	if err != nil {
		t.Errorf("nil cache must not cause error: %v", err)
	}
}

func TestGenerator_Invalidate_NilCache(t *testing.T) {
	t.Parallel()
	gen := sdui.New(testSchema(invoiceSchema()), nil, nil)

	if err := gen.Invalidate(context.Background(), "finance_invoice"); err != nil {
		t.Errorf("Invalidate with nil cache: %v", err)
	}
}

func TestGenerator_Invalidate_WithCache(t *testing.T) {
	t.Parallel()
	c := &countingCache{}
	gen := sdui.New(testSchema(invoiceSchema()), allowAllEvaluator{}, c)

	// Should not error.
	if err := gen.Invalidate(context.Background(), "finance_invoice"); err != nil {
		t.Errorf("Invalidate: %v", err)
	}
}

// ── Navigation tests ──────────────────────────────────────────────────────────

func TestGenerator_Nav_GroupsByModule(t *testing.T) {
	t.Parallel()
	gen := sdui.New(testSchema(invoiceSchema()), nil, nil)

	nav := gen.Nav()
	if len(nav) == 0 {
		t.Fatal("Nav must return at least one module")
	}
	if nav[0].Module != "finance" {
		t.Errorf("module = %v, want finance", nav[0].Module)
	}
	if len(nav[0].Entries) == 0 {
		t.Error("module must have at least one entry")
	}
}

func TestGenerator_Nav_ExcludesNoReadPermission(t *testing.T) {
	t.Parallel()
	es := invoiceSchema()
	es.Permissions.Read = nil
	gen := sdui.New(testSchema(es), nil, nil)

	nav := gen.Nav()
	if len(nav) != 0 {
		t.Error("Nav must exclude entities with no Read permission declared")
	}
}

func TestGenerator_Nav_CorrectListURL(t *testing.T) {
	t.Parallel()
	gen := sdui.New(testSchema(invoiceSchema()), nil, nil)

	nav := gen.Nav()
	entry := nav[0].Entries[0]
	if entry.ListURL != "/ui/finance/invoices" {
		t.Errorf("listUrl = %v, want /ui/finance/invoices", entry.ListURL)
	}
}

func TestGenerator_Nav_EntityName(t *testing.T) {
	t.Parallel()
	gen := sdui.New(testSchema(invoiceSchema()), nil, nil)

	nav := gen.Nav()
	entry := nav[0].Entries[0]
	if entry.Entity != "finance_invoice" {
		t.Errorf("entity = %v, want finance_invoice", entry.Entity)
	}
}

// ── WidgetTree tests ──────────────────────────────────────────────────────────

func TestGenerator_GetWidgetTree_ReturnsNodePage(t *testing.T) {
	t.Parallel()
	gen := sdui.New(testSchema(invoiceSchema()), allowAllEvaluator{}, nil)

	root, err := gen.GetWidgetTree(context.Background(), "finance_invoice", def.PageKindList, nil)
	if err != nil {
		t.Fatalf("GetWidgetTree: %v", err)
	}
	if root == nil {
		t.Fatal("GetWidgetTree returned nil root")
	}
	if root.Kind != widget.NodePage {
		t.Errorf("root kind = %v, want NodePage", root.Kind)
	}
}

// ── PageBuilder override test ─────────────────────────────────────────────────

func TestGenerator_PageBuilderOverride_ReturnsCustom(t *testing.T) {
	t.Parallel()
	es := invoiceSchema()
	es.PageBuilders = def.PageBuilderSet{
		List: func(_ context.Context, _ def.PageContext) (map[string]any, error) {
			return map[string]any{"type": "page", "title": "Custom Page"}, nil
		},
	}
	gen := sdui.New(testSchema(es), nil, nil)

	page, err := gen.GetPage(context.Background(), "finance_invoice", def.PageKindList, nil)
	if err != nil {
		t.Fatalf("GetPage with PageBuilder: %v", err)
	}
	if page["title"] != "Custom Page" {
		t.Errorf("PageBuilder override not applied: title = %v", page["title"])
	}
}

func TestGenerator_PageBuilderNilFallsThrough(t *testing.T) {
	t.Parallel()
	es := invoiceSchema()
	es.PageBuilders = def.PageBuilderSet{
		List: func(_ context.Context, _ def.PageContext) (map[string]any, error) {
			return nil, nil // nil → use auto-generation
		},
	}
	gen := sdui.New(testSchema(es), allowAllEvaluator{}, nil)

	page, err := gen.GetPage(context.Background(), "finance_invoice", def.PageKindList, nil)
	if err != nil {
		t.Fatalf("GetPage with nil PageBuilder result: %v", err)
	}
	// Auto-generated page should have "Invoices" title
	if page["title"] != "Invoices" {
		t.Errorf("nil PageBuilder must fall through to auto-generation: title = %v", page["title"])
	}
}

// ── Benchmark ─────────────────────────────────────────────────────────────────

func BenchmarkGenerator_ListPage(b *testing.B) {
	gen := sdui.New(testSchema(invoiceSchema()), allowAllEvaluator{}, nil)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := gen.GetPage(ctx, "finance_invoice", def.PageKindList, nil); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGenerator_CreatePage(b *testing.B) {
	gen := sdui.New(testSchema(invoiceSchema()), allowAllEvaluator{}, nil)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := gen.GetPage(ctx, "finance_invoice", def.PageKindCreate, nil); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGenerator_AllFourViews(b *testing.B) {
	gen := sdui.New(testSchema(invoiceSchema()), allowAllEvaluator{}, nil)
	ctx := context.Background()
	views := []def.PageKind{def.PageKindList, def.PageKindCreate, def.PageKindEdit, def.PageKindDetail}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, v := range views {
			if _, err := gen.GetPage(ctx, "finance_invoice", v, nil); err != nil {
				b.Fatal(err)
			}
		}
	}
}

// ── Layout Engine tests ───────────────────────────────────────────────────────

// withFieldsByName populates EntitySchema.FieldsByName from Fields.
// Required for layout-engine tests because the test helper builds EntitySchema
// directly (without the compiler), so lookup maps are not auto-populated.
func withFieldsByName(es *compiler.EntitySchema) *compiler.EntitySchema {
	es.FieldsByName = make(map[string]def.FieldDef, len(es.Fields))
	for _, f := range es.Fields {
		es.FieldsByName[f.Name] = f
	}
	return es
}

// invoiceSchemaWithLayout returns invoiceSchema with a two-tab layout declared.
func invoiceSchemaWithLayout() *compiler.EntitySchema {
	es := withFieldsByName(invoiceSchema())
	es.Layout = def.LayoutDef{
		Tabs: []def.TabDef{
			{
				Name:  "general",
				Label: "General",
				Sections: []def.SectionDef{
					{
						Name:  "header",
						Label: "Header",
						Columns: []def.ColumnDef{
							{Span: 6, Fields: []string{"number", "status"}},
							{Span: 6, Fields: []string{"customer_id", "total"}},
						},
					},
				},
			},
			{
				Name:  "notes_tab",
				Label: "Notes",
				Sections: []def.SectionDef{
					{
						Name:        "notes_section",
						Label:       "Internal Notes",
						Collapsible: true,
						Collapsed:   false,
						Columns: []def.ColumnDef{
							{Fields: []string{"notes"}},
						},
					},
				},
			},
		},
	}
	return es
}

func TestLayout_TabsGenerateNodeTabs(t *testing.T) {
	t.Parallel()
	es := invoiceSchemaWithLayout()
	gen := sdui.New(testSchema(es), allowAllEvaluator{}, nil)

	page, err := gen.GetPage(context.Background(), "finance_invoice", def.PageKindCreate, nil)
	if err != nil {
		t.Fatalf("GetPage with layout: %v", err)
	}
	body := page["body"].([]any)
	form := body[0].(map[string]any)
	formBody := form["body"].([]any)
	if len(formBody) == 0 {
		t.Fatal("form body is empty")
	}
	first := formBody[0].(map[string]any)
	if first["type"] != "tabs" {
		t.Errorf("layout with Tabs must produce tabs component, got %v", first["type"])
	}
}

func TestLayout_TabsHaveCorrectTitles(t *testing.T) {
	t.Parallel()
	es := invoiceSchemaWithLayout()
	gen := sdui.New(testSchema(es), allowAllEvaluator{}, nil)

	page, _ := gen.GetPage(context.Background(), "finance_invoice", def.PageKindCreate, nil)
	body := page["body"].([]any)
	form := body[0].(map[string]any)
	formBody := form["body"].([]any)
	tabsNode := formBody[0].(map[string]any)
	tabs := tabsNode["tabs"].([]any)

	if len(tabs) != 2 {
		t.Fatalf("expected 2 tabs, got %d", len(tabs))
	}
	if tabs[0].(map[string]any)["title"] != "General" {
		t.Errorf("tab[0].title = %v, want General", tabs[0].(map[string]any)["title"])
	}
	if tabs[1].(map[string]any)["title"] != "Notes" {
		t.Errorf("tab[1].title = %v, want Notes", tabs[1].(map[string]any)["title"])
	}
}

func TestLayout_SectionRendersAsFieldSet(t *testing.T) {
	t.Parallel()
	es := withFieldsByName(invoiceSchema())
	es.Layout = def.LayoutDef{
		Sections: []def.SectionDef{
			{
				Name:    "header",
				Label:   "Invoice Header",
				Columns: []def.ColumnDef{{Fields: []string{"number", "status"}}},
			},
		},
	}
	gen := sdui.New(testSchema(es), allowAllEvaluator{}, nil)

	page, err := gen.GetPage(context.Background(), "finance_invoice", def.PageKindCreate, nil)
	if err != nil {
		t.Fatalf("GetPage: %v", err)
	}
	body := page["body"].([]any)
	form := body[0].(map[string]any)
	formBody := form["body"].([]any)
	if len(formBody) == 0 {
		t.Fatal("form body empty")
	}
	sec := formBody[0].(map[string]any)
	if sec["type"] != "fieldSet" {
		t.Errorf("labeled section must render as fieldSet, got %v", sec["type"])
	}
	if sec["title"] != "Invoice Header" {
		t.Errorf("fieldSet title = %v, want 'Invoice Header'", sec["title"])
	}
}

func TestLayout_CollapsibleSection(t *testing.T) {
	t.Parallel()
	es := withFieldsByName(invoiceSchema())
	es.Layout = def.LayoutDef{
		Sections: []def.SectionDef{
			{
				Name:        "notes",
				Label:       "Notes",
				Collapsible: true,
				Collapsed:   true,
				Columns:     []def.ColumnDef{{Fields: []string{"notes"}}},
			},
		},
	}
	gen := sdui.New(testSchema(es), allowAllEvaluator{}, nil)

	page, _ := gen.GetPage(context.Background(), "finance_invoice", def.PageKindCreate, nil)
	body := page["body"].([]any)
	form := body[0].(map[string]any)
	formBody := form["body"].([]any)
	sec := formBody[0].(map[string]any)

	if sec["collapsable"] != true {
		t.Error("collapsible section must have collapsable: true")
	}
	if sec["collapsed"] != true {
		t.Error("collapsed section must have collapsed: true")
	}
}

func TestLayout_MultiColumnSpanClass(t *testing.T) {
	t.Parallel()
	es := withFieldsByName(invoiceSchema())
	es.Layout = def.LayoutDef{
		Sections: []def.SectionDef{
			{
				Name:  "two_col",
				Label: "Two Columns",
				Columns: []def.ColumnDef{
					{Span: 8, Fields: []string{"number"}},
					{Span: 4, Fields: []string{"status"}},
				},
			},
		},
	}
	gen := sdui.New(testSchema(es), allowAllEvaluator{}, nil)

	page, _ := gen.GetPage(context.Background(), "finance_invoice", def.PageKindCreate, nil)
	body := page["body"].([]any)
	form := body[0].(map[string]any)
	formBody := form["body"].([]any)
	sec := formBody[0].(map[string]any)
	fields := sec["body"].([]any)

	if len(fields) != 2 {
		t.Fatalf("expected 2 fields in two-column section, got %d", len(fields))
	}
	f0 := fields[0].(map[string]any)
	f1 := fields[1].(map[string]any)
	if f0["columnClassName"] != "col-md-8" {
		t.Errorf("first column columnClassName = %v, want col-md-8", f0["columnClassName"])
	}
	if f1["columnClassName"] != "col-md-4" {
		t.Errorf("second column columnClassName = %v, want col-md-4", f1["columnClassName"])
	}
}

func TestLayout_ZeroSpanAutoDistributes(t *testing.T) {
	t.Parallel()
	es := withFieldsByName(invoiceSchema())
	es.Layout = def.LayoutDef{
		Sections: []def.SectionDef{
			{
				Name:  "equal",
				Label: "Equal Columns",
				Columns: []def.ColumnDef{
					{Span: 0, Fields: []string{"number"}},
					{Span: 0, Fields: []string{"status"}},
					{Span: 0, Fields: []string{"total"}},
				},
			},
		},
	}
	gen := sdui.New(testSchema(es), allowAllEvaluator{}, nil)

	page, _ := gen.GetPage(context.Background(), "finance_invoice", def.PageKindCreate, nil)
	body := page["body"].([]any)
	form := body[0].(map[string]any)
	formBody := form["body"].([]any)
	sec := formBody[0].(map[string]any)
	fields := sec["body"].([]any)

	for i, f := range fields {
		field := f.(map[string]any)
		if field["columnClassName"] != "col-md-4" {
			t.Errorf("field[%d] columnClassName = %v, want col-md-4 (3 equal cols = 12/3)", i, field["columnClassName"])
		}
	}
}

func TestLayout_EmptyLayoutFallsBackToFlatList(t *testing.T) {
	t.Parallel()
	// invoiceSchema has no Layout set — should remain flat.
	gen := sdui.New(testSchema(invoiceSchema()), allowAllEvaluator{}, nil)

	page, _ := gen.GetPage(context.Background(), "finance_invoice", def.PageKindCreate, nil)
	body := page["body"].([]any)
	form := body[0].(map[string]any)
	formBody := form["body"].([]any)
	if len(formBody) == 0 {
		t.Fatal("flat layout form body must not be empty")
	}
	// First element must be a field, not a tabs/fieldSet container.
	first := formBody[0].(map[string]any)
	if first["type"] == "tabs" || first["type"] == "fieldSet" {
		t.Errorf("flat layout must not produce tabs/fieldSet, got %v", first["type"])
	}
}

func TestLayout_DetailPageUsesLayout(t *testing.T) {
	t.Parallel()
	es := invoiceSchemaWithLayout()
	gen := sdui.New(testSchema(es), allowAllEvaluator{}, nil)

	page, err := gen.GetPage(context.Background(), "finance_invoice", def.PageKindDetail, nil)
	if err != nil {
		t.Fatalf("GetPage detail with layout: %v", err)
	}
	body := page["body"].([]any)
	form := body[0].(map[string]any)
	formBody := form["body"].([]any)
	if len(formBody) == 0 {
		t.Fatal("detail page with layout must not be empty")
	}
	first := formBody[0].(map[string]any)
	if first["type"] != "tabs" {
		t.Errorf("detail page with tab layout must produce tabs, got %v", first["type"])
	}
}

func TestLayout_DetailPageAllReadOnly(t *testing.T) {
	t.Parallel()
	es := withFieldsByName(invoiceSchema())
	es.Layout = def.LayoutDef{
		Sections: []def.SectionDef{
			{
				Name:    "header",
				Label:   "Header",
				Columns: []def.ColumnDef{{Fields: []string{"number", "status", "total"}}},
			},
		},
	}
	gen := sdui.New(testSchema(es), allowAllEvaluator{}, nil)

	page, _ := gen.GetPage(context.Background(), "finance_invoice", def.PageKindDetail, nil)
	body := page["body"].([]any)
	form := body[0].(map[string]any)
	formBody := form["body"].([]any)
	sec := formBody[0].(map[string]any)
	fields := sec["body"].([]any)

	for _, f := range fields {
		field := f.(map[string]any)
		if field["disabled"] != true {
			t.Errorf("detail field %q must be disabled in layout-based detail page", field["name"])
		}
	}
}

// ── Expression Engine tests ───────────────────────────────────────────────────

func TestExpression_VisibleOnPassedThrough(t *testing.T) {
	t.Parallel()
	es := &compiler.EntitySchema{
		QualifiedName: "test_entity",
		LocalName:     "entity",
		Module:        "test",
		Label:         "Entity",
		LabelPlural:   "Entities",
		APIResource:   "entities",
		RoutePrefix:   "/api/v1/test/entities",
		Fields: []def.FieldDef{
			{Name: "status", Type: def.FieldTypeSelect, Options: []string{"Draft", "Active"}},
			{Name: "notes", Type: def.FieldTypeLongText, VisibleOn: "data.status === 'Active'"},
		},
		Permissions: def.PermissionSet{Read: []string{"test.entity.read"}},
	}
	gen := sdui.New(testSchema(es), allowAllEvaluator{}, nil)

	page, err := gen.GetPage(context.Background(), "test_entity", def.PageKindCreate, nil)
	if err != nil {
		t.Fatalf("GetPage: %v", err)
	}
	body := page["body"].([]any)
	form := body[0].(map[string]any)
	fields := form["body"].([]any)

	for _, f := range fields {
		field := f.(map[string]any)
		if field["name"] == "notes" {
			if field["visibleOn"] != "data.status === 'Active'" {
				t.Errorf("visibleOn = %v, want \"data.status === 'Active'\"", field["visibleOn"])
			}
			return
		}
	}
	t.Error("field 'notes' not found")
}

func TestExpression_HiddenOnPassedThrough(t *testing.T) {
	t.Parallel()
	es := &compiler.EntitySchema{
		QualifiedName: "test_entity",
		LocalName:     "entity",
		Module:        "test",
		Label:         "Entity",
		LabelPlural:   "Entities",
		APIResource:   "entities",
		RoutePrefix:   "/api/v1/test/entities",
		Fields: []def.FieldDef{
			{Name: "internal_note", Type: def.FieldTypeData, HiddenOn: "!data.is_admin"},
		},
		Permissions: def.PermissionSet{Read: []string{"test.entity.read"}},
	}
	gen := sdui.New(testSchema(es), allowAllEvaluator{}, nil)

	page, _ := gen.GetPage(context.Background(), "test_entity", def.PageKindCreate, nil)
	body := page["body"].([]any)
	form := body[0].(map[string]any)
	fields := form["body"].([]any)
	field := fields[0].(map[string]any)

	if field["hiddenOn"] != "!data.is_admin" {
		t.Errorf("hiddenOn = %v, want '!data.is_admin'", field["hiddenOn"])
	}
}

func TestExpression_DisabledOnOverridesReadOnly(t *testing.T) {
	t.Parallel()
	es := &compiler.EntitySchema{
		QualifiedName: "test_entity",
		LocalName:     "entity",
		Module:        "test",
		Label:         "Entity",
		LabelPlural:   "Entities",
		APIResource:   "entities",
		RoutePrefix:   "/api/v1/test/entities",
		Fields: []def.FieldDef{
			// ReadOnly: true AND DisabledOn set — DisabledOn must win.
			{Name: "amount", Type: def.FieldTypeCurrency, ReadOnly: true, DisabledOn: "data.status !== 'draft'"},
		},
		Permissions: def.PermissionSet{Read: []string{"test.entity.read"}},
	}
	gen := sdui.New(testSchema(es), allowAllEvaluator{}, nil)

	page, _ := gen.GetPage(context.Background(), "test_entity", def.PageKindCreate, nil)
	body := page["body"].([]any)
	form := body[0].(map[string]any)
	fields := form["body"].([]any)
	field := fields[0].(map[string]any)

	// disabledOn must be present, disabled: true must NOT be (expression wins).
	if _, ok := field["disabled"]; ok {
		t.Error("disabled: true must not be set when disabledOn expression is present")
	}
	if field["disabledOn"] != "data.status !== 'draft'" {
		t.Errorf("disabledOn = %v, want 'data.status !== 'draft''", field["disabledOn"])
	}
}

func TestExpression_RequiredOnOverridesRequired(t *testing.T) {
	t.Parallel()
	es := &compiler.EntitySchema{
		QualifiedName: "test_entity",
		LocalName:     "entity",
		Module:        "test",
		Label:         "Entity",
		LabelPlural:   "Entities",
		APIResource:   "entities",
		RoutePrefix:   "/api/v1/test/entities",
		Fields: []def.FieldDef{
			// Required: true AND RequiredOn set — RequiredOn must win.
			{Name: "tax_id", Type: def.FieldTypeData, Required: true, RequiredOn: "data.is_company === true"},
		},
		Permissions: def.PermissionSet{Read: []string{"test.entity.read"}},
	}
	gen := sdui.New(testSchema(es), allowAllEvaluator{}, nil)

	page, _ := gen.GetPage(context.Background(), "test_entity", def.PageKindCreate, nil)
	body := page["body"].([]any)
	form := body[0].(map[string]any)
	fields := form["body"].([]any)
	field := fields[0].(map[string]any)

	if _, ok := field["required"]; ok {
		t.Error("required: true must not be set when requiredOn expression is present")
	}
	if field["requiredOn"] != "data.is_company === true" {
		t.Errorf("requiredOn = %v, want 'data.is_company === true'", field["requiredOn"])
	}
}

func TestExpression_EmptyExpressionFallsBackToBoolean(t *testing.T) {
	t.Parallel()
	es := &compiler.EntitySchema{
		QualifiedName: "test_entity",
		LocalName:     "entity",
		Module:        "test",
		Label:         "Entity",
		LabelPlural:   "Entities",
		APIResource:   "entities",
		RoutePrefix:   "/api/v1/test/entities",
		Fields: []def.FieldDef{
			{Name: "amount", Type: def.FieldTypeCurrency, Required: true, ReadOnly: true},
		},
		Permissions: def.PermissionSet{Read: []string{"test.entity.read"}},
	}
	gen := sdui.New(testSchema(es), allowAllEvaluator{}, nil)

	page, _ := gen.GetPage(context.Background(), "test_entity", def.PageKindCreate, nil)
	body := page["body"].([]any)
	form := body[0].(map[string]any)
	fields := form["body"].([]any)
	field := fields[0].(map[string]any)

	if field["required"] != true {
		t.Error("required: true must be set when Required=true and no RequiredOn")
	}
	if field["disabled"] != true {
		t.Error("disabled: true must be set when ReadOnly=true and no DisabledOn")
	}
}

func BenchmarkGenerator_LayoutWithTabs(b *testing.B) {
	gen := sdui.New(testSchema(invoiceSchemaWithLayout()), allowAllEvaluator{}, nil)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := gen.GetPage(ctx, "finance_invoice", def.PageKindCreate, nil); err != nil {
			b.Fatal(err)
		}
	}
}
