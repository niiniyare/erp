package runtime_test

import (
	"context"
	"testing"

	"awo.so/awo/compiler"
	"awo.so/awo/def"
	"awo.so/awo/runtime"
)

// buildTestSchema constructs a minimal CompiledSchema for test use.
func buildTestSchema(d def.EntityDefinition) *compiler.CompiledSchema {
	// Re-implement buildEntitySchema inline to avoid importing compiler internals.
	// In real tests, use registry.Build() + compiler.Compile().
	schema := &compiler.CompiledSchema{
		ByName: map[string]*compiler.EntitySchema{},
	}
	es := &compiler.EntitySchema{
		Def:             d,
		FieldsByName:    map[string]def.FieldDef{},
		EdgesByName:     map[string]def.EdgeDef{},
		ActionsByName:   map[string]def.ActionDef{},
		DefaultValues:   map[string]func() any{},
		RequiredFields:  map[string]bool{},
		ImmutableFields: map[string]bool{},
		SensitiveFields: map[string]bool{},
		SearchableFields: map[string]bool{},
		LinkTargets:     map[string]*compiler.EntitySchema{},
		TableName:       d.EntityName(),
	}
	for _, f := range d.EntityFields() {
		es.FieldsByName[f.Name] = f
		if f.Required {
			es.RequiredFields[f.Name] = true
		}
		if f.Immutable {
			es.ImmutableFields[f.Name] = true
		}
		if f.Default != nil {
			es.DefaultValues[f.Name] = f.Default
		}
	}
	schema.Entities = []*compiler.EntitySchema{es}
	schema.ByName[d.EntityName()] = es
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

	schema := buildTestSchema(d)
	pipeline := runtime.NewPipeline(schema)

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

	schema := buildTestSchema(d)
	pipeline := runtime.NewPipeline(schema)

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

	schema := buildTestSchema(d)
	pipeline := runtime.NewPipeline(schema)

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

	schema := buildTestSchema(d)
	pipeline := runtime.NewPipeline(schema)

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
