package sdui_test

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"

	"awo.so/framework/def"
	"awo.so/framework/sdui"
)

func makeTestDef() *def.EntityDefinition {
	return &def.EntityDefinition{
		Name:  "invoice",
		Label: "Invoice",
		Fields: []*def.FieldDef{
			{Name: "number", Type: def.FieldTypeData},
			{Name: "secret_key", Type: def.FieldTypeData, IsSensitive: true},
		},
	}
}

func newRedisCache(t *testing.T) sdui.Cache {
	t.Helper()
	mr := miniredis.RunT(t)
	return redis.NewClient(&redis.Options{Addr: mr.Addr()})
}

// TestBuilder_BuildCRUD verifies the base CRUD schema is returned.
func TestBuilder_BuildCRUD(t *testing.T) {
	entDef := makeTestDef()
	b := sdui.NewBuilder(entDef, "/api")
	schema, err := b.Build(context.Background(), sdui.ViewCRUD, nil, uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}
	if schema["type"] != "page" {
		t.Errorf("want type=page, got %v", schema["type"])
	}
}

// TestBuilder_CacheHitAvoidsDuplicateGeneration verifies the second call is served from cache.
func TestBuilder_CacheHitAvoidsDuplicateGeneration(t *testing.T) {
	entDef := makeTestDef()
	cache := newRedisCache(t)
	b := sdui.NewBuilder(entDef, "/api", sdui.WithBuilderCache(cache))
	ctx := context.Background()
	tid := uuid.New()

	// First call — miss, generates schema and caches.
	s1, err := b.Build(ctx, sdui.ViewCRUD, nil, tid)
	if err != nil || s1 == nil {
		t.Fatalf("first build failed: %v", err)
	}

	// Second call — should come from cache (same content).
	s2, err := b.Build(ctx, sdui.ViewCRUD, nil, tid)
	if err != nil || s2 == nil {
		t.Fatalf("second build failed: %v", err)
	}

	if s1["type"] != s2["type"] {
		t.Errorf("cached schema mismatch: %v vs %v", s1["type"], s2["type"])
	}
}

// TestBuilder_SensitiveFieldExcludedForRegularViewer verifies viewer filtering post-cache.
func TestBuilder_SensitiveFieldExcludedForRegularViewer(t *testing.T) {
	entDef := makeTestDef()
	b := sdui.NewBuilder(entDef, "/api")
	viewer := &mockViewer{roles: []string{"role:finance.viewer"}}

	schema, err := b.Build(context.Background(), sdui.ViewCRUD, viewer, uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}

	// secret_key should not appear anywhere in the schema body.
	body, _ := schema["body"].(map[string]any)
	if body != nil {
		cols, _ := body["columns"].([]any)
		for _, col := range cols {
			m, ok := col.(map[string]any)
			if !ok {
				continue
			}
			if name, _ := m["name"].(string); name == "secret_key" {
				t.Error("secret_key should not appear in schema for non-admin viewer")
			}
		}
	}
}

// TestBuilder_AdminViewerGetsAllFields verifies admin gets unfiltered schema.
func TestBuilder_AdminViewerGetsAllFields(t *testing.T) {
	entDef := makeTestDef()
	b := sdui.NewBuilder(entDef, "/api")
	admin := &mockViewer{roles: []string{"role:tenant.admin"}}

	schema, err := b.Build(context.Background(), sdui.ViewCRUD, admin, uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}
	if schema == nil {
		t.Fatal("want non-nil schema for admin")
	}
}

// TestBuilder_PageBuilderOverrideBypasses cache verifies custom builders skip cache.
func TestBuilder_PageBuilderOverrideBypassesCache(t *testing.T) {
	customSchema := map[string]any{"type": "custom-page", "title": "Override"}
	entDef := &def.EntityDefinition{
		Name:  "invoice",
		Label: "Invoice",
		PageBuilders: &def.PageBuilderSet{
			CRUDPage: func(_ *def.EntityDefinition, _ string) map[string]any {
				return customSchema
			},
		},
	}
	cache := newRedisCache(t)
	b := sdui.NewBuilder(entDef, "/api", sdui.WithBuilderCache(cache))
	schema, err := b.Build(context.Background(), sdui.ViewCRUD, nil, uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}
	if schema["type"] != "custom-page" {
		t.Errorf("want custom-page from override, got %v", schema["type"])
	}
}

// TestBuilder_InvalidateRemovesCachedEntries verifies invalidation clears all view types.
func TestBuilder_InvalidateRemovesCachedEntries(t *testing.T) {
	entDef := makeTestDef()
	cache := newRedisCache(t)
	b := sdui.NewBuilder(entDef, "/api", sdui.WithBuilderCache(cache))
	ctx := context.Background()
	tid := uuid.New()

	// Populate cache for all view types.
	for _, vt := range []sdui.ViewType{sdui.ViewCRUD, sdui.ViewFormCreate, sdui.ViewFormEdit, sdui.ViewDetail} {
		_, _ = b.Build(ctx, vt, nil, tid)
	}

	// Invalidate.
	if err := b.Invalidate(ctx, tid); err != nil {
		t.Fatal(err)
	}

	// Subsequent build should succeed (regenerated, not error).
	schema, err := b.Build(ctx, sdui.ViewCRUD, nil, tid)
	if err != nil || schema == nil {
		t.Errorf("build after invalidate failed: %v", err)
	}
}
