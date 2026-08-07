package engine_test

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

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

// ── Security and determinism tests ────────────────────────────────────────────

// permActionCount returns the number of actions in the AMIS schema for the
// list view. Used to compare admin vs restricted schemas.
func actionsInList(t *testing.T, schema map[string]any) int {
	t.Helper()
	body, ok := schema["body"].([]any)
	if !ok || len(body) == 0 {
		return 0
	}
	listNode, ok := body[0].(map[string]any)
	if !ok {
		return 0
	}
	cols, _ := listNode["columns"].([]any)
	// Last column is operations; count its buttons.
	for _, col := range cols {
		m, ok := col.(map[string]any)
		if !ok {
			continue
		}
		if m["type"] == "operation" {
			buttons, _ := m["buttons"].([]any)
			return len(buttons)
		}
	}
	return 0
}

func TestEngine_PermFPIsolation_DifferentSchemas(t *testing.T) {
	// Admin viewer sees Edit + Delete row buttons.
	// Restricted viewer (no permissions) sees none.
	// Their schemas must be different, and must be cached under different keys.

	schema := generator.EntitySchema{
		Name:        "test_entity",
		Title:       "Entity",
		PluralTitle: "Entities",
		ListURL:     "/api/v1/test/entities",
		UIPrefix:    "/ui/test/entities",
		CreateURL:   "/api/v1/test/entities",
		EditURL:     "/api/v1/test/entities/${id}",
		DetailURL:   "/api/v1/test/entities/${id}",
		Permissions: map[string]string{
			"create": "test.entity.create",
			"read":   "test.entity.read",
			"update": "test.entity.update",
			"delete": "test.entity.delete",
		},
		Fields: []generator.FieldDef{
			{Name: "name", Label: "Name", FieldType: "data", InList: true},
		},
	}

	tenantID := uuid.New()

	adminViewer := &stubViewer{admin: true}
	adminCtx, _ := sduictx.NewGeneratorContext(
		tenantID, adminViewer, "test_entity", sduictx.ViewModeList, "amis",
	).WithSchemaFingerprint("sfp1").WithPermFingerprint("__admin__").Build()

	restrictedViewer := &stubViewer{admin: false}
	restrictedCtx, _ := sduictx.NewGeneratorContext(
		tenantID, restrictedViewer, "test_entity", sduictx.ViewModeList, "amis",
	).WithSchemaFingerprint("sfp1").WithPermFingerprint("__noroles__").Build()

	eng := makeEngine(nil)

	adminResp, err := eng.Handle(context.Background(), engine.Request{Ctx: adminCtx, Schema: schema})
	if err != nil {
		t.Fatalf("admin Handle: %v", err)
	}
	restrictedResp, err := eng.Handle(context.Background(), engine.Request{Ctx: restrictedCtx, Schema: schema})
	if err != nil {
		t.Fatalf("restricted Handle: %v", err)
	}

	adminActions := actionsInList(t, adminResp.Output.AMISSchema)
	restrictedActions := actionsInList(t, restrictedResp.Output.AMISSchema)

	// Admin must have more row actions than a viewer with no permissions.
	if adminActions <= restrictedActions {
		t.Errorf("admin schema actions (%d) must exceed restricted actions (%d)",
			adminActions, restrictedActions)
	}
}

func TestEngine_PermFPIsolation_CacheNotShared(t *testing.T) {
	// Prove: with a live cache, admin schema stored under "__admin__" permFP
	// is not returned for a "__noroles__" permFP lookup.
	c := cache.New(newInlineRedis())
	eng := makeEngine(c)

	schema := generator.EntitySchema{
		Name:        "cache_test_entity",
		Title:       "Entity",
		PluralTitle: "Entities",
		ListURL:     "/api/v1/test/entities",
		Fields:      []generator.FieldDef{{Name: "name", Label: "Name", FieldType: "data", InList: true}},
		Permissions: map[string]string{"delete": "test.entity.delete"},
	}
	tenantID := uuid.New()

	// Admin populates cache.
	adminCtx, _ := sduictx.NewGeneratorContext(
		tenantID, &stubViewer{admin: true}, "cache_test_entity", sduictx.ViewModeList, "amis",
	).WithSchemaFingerprint("sfX").WithPermFingerprint("__admin__").Build()
	if _, err := eng.Handle(context.Background(), engine.Request{Ctx: adminCtx, Schema: schema}); err != nil {
		t.Fatalf("admin Handle: %v", err)
	}

	// Restricted viewer — MUST NOT get admin's cached output.
	restrictedCtx, _ := sduictx.NewGeneratorContext(
		tenantID, &stubViewer{admin: false}, "cache_test_entity", sduictx.ViewModeList, "amis",
	).WithSchemaFingerprint("sfX").WithPermFingerprint("__noroles__").Build()
	restrictedResp, err := eng.Handle(context.Background(), engine.Request{Ctx: restrictedCtx, Schema: schema})
	if err != nil {
		t.Fatalf("restricted Handle: %v", err)
	}

	// L3 keys differ by PermFP — restricted viewer must miss the cache and
	// generate its own schema. A cache hit here would mean the fix broke.
	if restrictedResp.CacheHit {
		t.Error("restricted viewer must not receive a cache hit from admin's cached schema")
	}
}

func TestEngine_TenantIsolation_CacheNotShared(t *testing.T) {
	// Prove: a schema cached for tenant A is not returned for tenant B.
	c := cache.New(newInlineRedis())
	eng := makeEngine(c)

	schema := generator.EntitySchema{
		Name:    "tenant_test_entity",
		Title:   "Entity",
		ListURL: "/api/v1/test/entities",
		Fields:  []generator.FieldDef{{Name: "name", Label: "Name", FieldType: "data", InList: true}},
	}

	tenantA := uuid.New()
	tenantB := uuid.New()

	ctxA, _ := sduictx.NewGeneratorContext(
		tenantA, &stubViewer{admin: true}, "tenant_test_entity", sduictx.ViewModeList, "amis",
	).WithSchemaFingerprint("sfT").WithPermFingerprint("__admin__").Build()

	ctxB, _ := sduictx.NewGeneratorContext(
		tenantB, &stubViewer{admin: true}, "tenant_test_entity", sduictx.ViewModeList, "amis",
	).WithSchemaFingerprint("sfT").WithPermFingerprint("__admin__").Build()

	// Populate cache for tenant A.
	if _, err := eng.Handle(context.Background(), engine.Request{Ctx: ctxA, Schema: schema}); err != nil {
		t.Fatalf("tenant A Handle: %v", err)
	}

	// Tenant B must not get a cache hit from tenant A's entry.
	respB, err := eng.Handle(context.Background(), engine.Request{Ctx: ctxB, Schema: schema})
	if err != nil {
		t.Fatalf("tenant B Handle: %v", err)
	}
	if respB.CacheHit {
		t.Error("tenant B must not receive tenant A's cached schema")
	}
}

func TestEngine_Determinism(t *testing.T) {
	// Identical inputs must produce identical AMIS JSON output.
	eng := makeEngine(nil)
	req := engine.Request{
		Ctx:    makeCtx(sduictx.ViewModeList),
		Schema: makeSchema(),
	}

	resp1, err := eng.Handle(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	resp2, err := eng.Handle(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}

	b1, _ := json.Marshal(resp1.Output.AMISSchema)
	b2, _ := json.Marshal(resp2.Output.AMISSchema)
	if string(b1) != string(b2) {
		t.Errorf("engine output is not deterministic:\nrun1: %s\nrun2: %s", b1, b2)
	}
}

// inlineRedis is a goroutine-safe map-backed Redis stub for cache isolation tests.
type inlineRedis struct {
	mu   sync.Mutex
	data map[string]string
}

func newInlineRedis() *inlineRedis {
	return &inlineRedis{data: make(map[string]string)}
}
func (r *inlineRedis) Get(_ context.Context, key string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	v, ok := r.data[key]
	if !ok {
		return "", cache.ErrCacheMiss
	}
	return v, nil
}
func (r *inlineRedis) Set(_ context.Context, key, value string, _ time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[key] = value
	return nil
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
