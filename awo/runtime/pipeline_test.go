package runtime_test

import (
	"context"
	"testing"

	"awo.so/awo/audit"
	"awo.so/awo/compiler"
	"awo.so/awo/def"
	"awo.so/awo/registry"
	"awo.so/awo/runtime"
)

// buildTestSchema compiles a minimal CompiledSchema from a single EntityDefinition.
func buildTestSchema(t *testing.T, d def.EntityDefinition) *compiler.CompiledSchema {
	t.Helper()
	reg, err := registry.BuildFrom([]def.EntityDefinition{d})
	if err != nil {
		t.Fatalf("registry.BuildFrom: %v", err)
	}
	schema, err := compiler.Compile(reg)
	if err != nil {
		t.Fatalf("compiler.Compile: %v", err)
	}
	return schema
}

func TestPipeline_RunBeforeCreate_AppliesDefaults(t *testing.T) {
	d := &def.SystemDefinition{
		Name:   "test_order",
		Module: "test",
		Label:  "Order",
		Fields: []def.FieldDef{
			{
				Name:    "status",
				Type:    def.FieldTypeSelect,
				Options: []string{"pending", "completed"},
				Default: func() any { return "pending" },
			},
		},
	}

	schema := buildTestSchema(t, d)
	pipeline := runtime.NewPipeline(schema, audit.NoopAuditWriter{})

	pctx := &runtime.CreateContext{
		Ctx:        context.Background(),
		EntityName: "test_order",
		Data:       map[string]any{}, // no status provided
		Actor:      nil,
	}

	record, err := pipeline.RunBeforeCreate(pctx)
	if err != nil {
		t.Fatalf("RunBeforeCreate failed: %v", err)
	}
	if record.GetString("status") != "pending" {
		t.Errorf("default not applied: got %q, want %q", record.GetString("status"), "pending")
	}
}

func TestPipeline_RunBeforeCreate_RequiredFieldMissing(t *testing.T) {
	d := &def.SystemDefinition{
		Name:   "test_item",
		Module: "test",
		Label:  "Item",
		Fields: []def.FieldDef{
			{Name: "name", Type: def.FieldTypeData, Required: true},
		},
	}

	schema := buildTestSchema(t, d)
	pipeline := runtime.NewPipeline(schema, audit.NoopAuditWriter{})

	pctx := &runtime.CreateContext{
		Ctx:        context.Background(),
		EntityName: "test_item",
		Data:       map[string]any{}, // name missing
		Actor:      nil,
	}

	_, err := pipeline.RunBeforeCreate(pctx)
	if err == nil {
		t.Fatal("expected validation error for missing required field, got nil")
	}
	if !runtime.IsValidation(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestPipeline_RunBeforeUpdate_ImmutableField(t *testing.T) {
	d := &def.SystemDefinition{
		Name:   "test_contract",
		Module: "test",
		Label:  "Contract",
		Fields: []def.FieldDef{
			{Name: "reference", Type: def.FieldTypeData, Immutable: true},
			{Name: "value", Type: def.FieldTypeCurrency},
		},
	}

	schema := buildTestSchema(t, d)
	pipeline := runtime.NewPipeline(schema, audit.NoopAuditWriter{})

	current := &def.EntityRecord{
		EntityName: "test_contract",
		Data:       map[string]any{"reference": "CTR-001", "value": nil},
	}

	pctx := &runtime.UpdateContext{
		Ctx:        context.Background(),
		EntityName: "test_contract",
		Data:       map[string]any{"reference": "CTR-002"}, // attempt to change immutable
		Actor:      nil,
	}

	_, err := pipeline.RunBeforeUpdate(pctx, current)
	if err == nil {
		t.Fatal("expected error for immutable field update, got nil")
	}
	if !runtime.IsValidation(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestPipeline_RunBeforeCreate_HookAborts(t *testing.T) {
	hook := &rejectAllHook{}
	d := &def.SystemDefinition{
		Name:   "test_blocked",
		Module: "test",
		Label:  "Blocked",
		Hooks: def.HookSet{
			BeforeCreate: []def.BeforeCreateHook{hook},
		},
	}

	schema := buildTestSchema(t, d)
	pipeline := runtime.NewPipeline(schema, audit.NoopAuditWriter{})

	pctx := &runtime.CreateContext{
		Ctx:        context.Background(),
		EntityName: "test_blocked",
		Data:       map[string]any{},
		Actor:      nil,
	}

	_, err := pipeline.RunBeforeCreate(pctx)
	if err == nil {
		t.Fatal("expected hook to abort create, got nil error")
	}
}

// rejectAllHook is a test BeforeCreateHook that always rejects.
type rejectAllHook struct{}

func (h *rejectAllHook) BeforeCreate(_ context.Context, _ *def.EntityRecord) error {
	return &runtime.BusinessError{
		Code:    "test.always_rejected",
		Message: "always rejected",
		Status:  400,
	}
}
