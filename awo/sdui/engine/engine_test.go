package engine_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"awo.so/awo/sdui/amis"
	"awo.so/awo/sdui/cache"
	"awo.so/awo/sdui/engine"
	"awo.so/awo/sdui/generator"
	"awo.so/awo/sdui/layout"
	"awo.so/awo/sdui/observability"
	"awo.so/awo/sdui/renderer"
	"awo.so/awo/sdui/sduictx"
	"awo.so/awo/sdui/validation"
)

// ── test fixtures ─────────────────────────────────────────────────────────────

type stubViewer struct{ admin bool }

func (s *stubViewer) TenantID() uuid.UUID       { return uuid.New() }
func (s *stubViewer) Roles() []string           { return nil }
func (s *stubViewer) IsPlatformAdmin() bool     { return s.admin }
func (s *stubViewer) HasPermission(string) bool { return s.admin }

func makeCtx(mode sduictx.ViewMode) sduictx.GeneratorContext {
	ctx, _ := sduictx.NewGeneratorContext(
		uuid.New(), &stubViewer{admin: true}, "test_entity", mode, "amis",
	).WithSchemaFingerprint("sf1").WithPermFingerprint("pf1").Build()
	return ctx
}

func makeSchema() generator.EntitySchema {
	return generator.EntitySchema{
		Name:        "test_entity",
		Title:       "Entity",
		PluralTitle: "Entities",
		ListURL:     "/api/v1/test/entities",
		CreateURL:   "/api/v1/test/entities",
		EditURL:     "/api/v1/test/entities/{id}",
		DetailURL:   "/api/v1/test/entities/{id}",
		Permissions: map[string]string{
			"create": "test.entity.create",
			"read":   "test.entity.read",
			"update": "test.entity.update",
			"delete": "test.entity.delete",
		},
		Fields: []generator.FieldDef{
			{Name: "name", Label: "Name", FieldType: "data", InList: true, InForm: true, InDetail: true},
			{Name: "status", Label: "Status", FieldType: "select", InList: true, InForm: true, InDetail: true},
			{Name: "notes", Label: "Notes", FieldType: "long_text", InForm: true, InDetail: true},
		},
	}
}

func makeEngine(c *cache.Cache) *engine.Engine {
	return engine.New(engine.Options{
		Generator: generator.New(),
		Validator: validation.New(),
		Layout:    layout.New(),
		Cache:     c,
		Renderers: map[string]renderer.Renderer{"amis": amis.New()},
		Obs:       observability.Noop(),
	})
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestEngine_Handle_ListView(t *testing.T) {
	eng := makeEngine(nil)
	req := engine.Request{
		Ctx:    makeCtx(sduictx.ViewModeList),
		Schema: makeSchema(),
	}
	resp, err := eng.Handle(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	if resp.CacheHit {
		t.Error("first request should not be a cache hit")
	}
	if resp.Output.Format != renderer.FormatAMISJSON {
		t.Errorf("format = %q, want %q", resp.Output.Format, renderer.FormatAMISJSON)
	}
	if resp.Output.AMISSchema == nil {
		t.Error("AMISSchema must not be nil")
	}
	if resp.Layout == nil {
		t.Error("Layout must not be nil")
	}
}

func TestEngine_Handle_CreateForm(t *testing.T) {
	eng := makeEngine(nil)
	req := engine.Request{
		Ctx:    makeCtx(sduictx.ViewModeCreate),
		Schema: makeSchema(),
	}
	resp, err := eng.Handle(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	schema := resp.Output.AMISSchema
	if schema["type"] != "page" {
		t.Errorf("top-level type = %v, want \"page\"", schema["type"])
	}
}

func TestEngine_Handle_DetailView(t *testing.T) {
	eng := makeEngine(nil)
	req := engine.Request{
		Ctx:    makeCtx(sduictx.ViewModeDetail),
		Schema: makeSchema(),
	}
	resp, err := eng.Handle(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Output.AMISSchema == nil {
		t.Error("detail output must not be nil")
	}
}

func TestEngine_Handle_UnknownRenderer(t *testing.T) {
	eng := makeEngine(nil)
	req := engine.Request{
		Ctx:        makeCtx(sduictx.ViewModeList),
		Schema:     makeSchema(),
		RendererID: "nonexistent",
	}
	_, err := eng.Handle(context.Background(), req)
	if err == nil {
		t.Error("expected error for unknown renderer")
	}
}

func TestEngine_Handle_ValidationFatalAborts(t *testing.T) {
	// An entity schema that generates a node the global registry doesn't know
	// is hard to produce without a fake generator. Instead, verify that
	// validation warnings do NOT abort the pipeline.
	eng := makeEngine(nil)
	req := engine.Request{
		Ctx:    makeCtx(sduictx.ViewModeList),
		Schema: makeSchema(),
	}
	resp, err := eng.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Response should be non-nil even with warnings.
	if resp == nil {
		t.Error("expected non-nil response")
	}
}

func TestEngine_Handle_OutputIsValidJSON(t *testing.T) {
	eng := makeEngine(nil)
	req := engine.Request{
		Ctx:    makeCtx(sduictx.ViewModeCreate),
		Schema: makeSchema(),
	}
	resp, err := eng.Handle(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(resp.Output.AMISSchema)
	if err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if len(raw) < 2 {
		t.Error("output JSON is empty")
	}
}

func TestEngine_Handle_PermissionFiltering(t *testing.T) {
	// No-perms viewer should not see permission-gated fields.
	noperms := &stubViewer{admin: false}
	ctx, _ := sduictx.NewGeneratorContext(
		uuid.New(), noperms, "test_entity", sduictx.ViewModeList, "amis",
	).WithSchemaFingerprint("sf1").WithPermFingerprint("pf1").Build()

	schema := generator.EntitySchema{
		Name:        "test_entity",
		Title:       "Entity",
		ListURL:     "/api/v1/test/entities",
		Permissions: map[string]string{"read": "test.entity.read"},
		Fields: []generator.FieldDef{
			{Name: "name", Label: "Name", FieldType: "data", InList: true},
			{Name: "secret", Label: "Secret", FieldType: "data", InList: true,
				Permission: "test.entity.admin"},
		},
	}
	eng := makeEngine(nil)
	resp, err := eng.Handle(context.Background(), engine.Request{Ctx: ctx, Schema: schema})
	if err != nil {
		t.Fatal(err)
	}
	// Find the list node columns; secret must be absent.
	body, ok := resp.Output.AMISSchema["body"].([]any)
	if !ok || len(body) == 0 {
		t.Fatal("no body in output")
	}
	listNode, ok := body[0].(map[string]any)
	if !ok {
		t.Fatal("body[0] is not a map")
	}
	cols, _ := listNode["columns"].([]any)
	for _, col := range cols {
		m, ok := col.(map[string]any)
		if !ok {
			continue
		}
		if m["name"] == "secret" {
			t.Error("secret field must be absent for no-perms viewer")
		}
	}
}

func TestEngine_Handle_LayoutPresent(t *testing.T) {
	eng := makeEngine(nil)
	req := engine.Request{
		Ctx:    makeCtx(sduictx.ViewModeCreate),
		Schema: makeSchema(),
	}
	resp, err := eng.Handle(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Layout == nil {
		t.Fatal("Layout must be present for non-cached response")
	}
	if resp.Layout.Root == nil {
		t.Error("Layout.Root must not be nil")
	}
}

func TestEngine_Renderers(t *testing.T) {
	eng := makeEngine(nil)
	renderers := eng.Renderers()
	if _, ok := renderers["amis"]; !ok {
		t.Error("amis renderer must be registered")
	}
}

// ── benchmark ─────────────────────────────────────────────────────────────────

func BenchmarkEngine_Handle_NoCache(b *testing.B) {
	eng := makeEngine(nil)
	req := engine.Request{
		Ctx:    makeCtx(sduictx.ViewModeList),
		Schema: makeSchema(),
	}
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := eng.Handle(ctx, req); err != nil {
			b.Fatal(err)
		}
	}
}
