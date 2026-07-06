package runtime

import (
	"context"
	"fmt"

	"awo.so/awo/compiler"
	"awo.so/awo/def"
)

// Pipeline executes the entity lifecycle hook sequence for a mutation
// operation. It is the authoritative implementation of:
//
//	ASSEMBLE → BeforeValidate → VALIDATE → AUTHORIZE
//	         → BeforeSave / BeforeCreate / BeforeUpdate
//	         → [TX begins]  (delegated to the driver)
//	         → PERSIST      (delegated to the driver)
//	         → AfterSave / AfterCreate / AfterUpdate
//	         → [TX commits] (delegated to the driver)
//	         → Workflow start (outside TX)
//
// The Pipeline does not open database transactions directly. It calls the
// driver's Create/Update/Delete, which manage transactions internally and
// call the after_* hook stage from within the transaction.
type Pipeline struct {
	schema *compiler.CompiledSchema
}

// NewPipeline creates a Pipeline bound to the compiled schema.
func NewPipeline(schema *compiler.CompiledSchema) *Pipeline {
	return &Pipeline{schema: schema}
}

// CreateContext carries all inputs for a Create pipeline run.
type CreateContext struct {
	// Ctx is the request context, carrying TenantContext.
	Ctx context.Context

	// EntityName identifies the entity definition to use.
	EntityName string

	// Data is the raw field values from the API request body.
	Data map[string]any

	// CustomFields is the custom field values from the API request body.
	CustomFields map[string]any

	// Actor is the authenticated principal.
	Actor *def.Actor
}

// UpdateContext carries all inputs for an Update pipeline run.
type UpdateContext struct {
	Ctx        context.Context
	EntityName string
	Data       map[string]any
	CustomFields map[string]any
	Actor      *def.Actor
}

// DeleteContext carries all inputs for a Delete pipeline run.
type DeleteContext struct {
	Ctx        context.Context
	EntityName string
	Actor      *def.Actor
}

// RunBeforeCreate executes the pre-persist stages of the Create pipeline.
// The caller (driver) invokes this before opening the database transaction.
//
// Stages executed:
//  1. ApplyDefaults — populate missing fields from FieldDef.Default
//  2. BeforeValidate hooks
//  3. VALIDATE — field-level and FieldValidator checks
//  4. ImmutableField check (no-op on Create)
//  5. BeforeSave hooks
//  6. BeforeCreate hooks
//
// Returns the assembled EntityRecord. The driver then persists it and calls
// [RunAfterCreate] from within the database transaction.
func (p *Pipeline) RunBeforeCreate(pctx *CreateContext) (*def.EntityRecord, error) {
	es, err := p.lookupSchema(pctx.EntityName)
	if err != nil {
		return nil, err
	}

	record := &def.EntityRecord{
		EntityName:   pctx.EntityName,
		Data:         cloneMap(pctx.Data),
		CustomFields: cloneMap(pctx.CustomFields),
		Meta: def.RecordMeta{
			Actor:         pctx.Actor,
			OperationType: def.OperationCreate,
		},
	}

	// Stage 1: apply defaults for missing fields.
	p.applyDefaults(record, es)

	hooks := es.Def.EntityHooks()

	// Stage 2: BeforeValidate
	for _, h := range hooks.BeforeValidate {
		if err := h.BeforeValidate(pctx.Ctx, record); err != nil {
			return nil, fmt.Errorf("before_validate: %w", err)
		}
	}

	// Stage 3: VALIDATE
	if err := p.validateFields(pctx.Ctx, record, es); err != nil {
		return nil, err
	}

	// Stage 5: BeforeSave
	for _, h := range hooks.BeforeSave {
		if err := h.BeforeSave(pctx.Ctx, record); err != nil {
			return nil, fmt.Errorf("before_save: %w", err)
		}
	}

	// Stage 6: BeforeCreate
	for _, h := range hooks.BeforeCreate {
		if err := h.BeforeCreate(pctx.Ctx, record); err != nil {
			return nil, fmt.Errorf("before_create: %w", err)
		}
	}

	return record, nil
}

// RunAfterCreate executes the post-persist stages of the Create pipeline.
// Called by the driver FROM WITHIN the open database transaction.
//
// If any hook returns an error, the driver must roll back the transaction.
func (p *Pipeline) RunAfterCreate(ctx context.Context, record *def.EntityRecord) error {
	es, err := p.lookupSchema(record.EntityName)
	if err != nil {
		return err
	}
	hooks := es.Def.EntityHooks()

	for _, h := range hooks.AfterSave {
		if err := h.AfterSave(ctx, record); err != nil {
			return fmt.Errorf("after_save: %w", err)
		}
	}
	for _, h := range hooks.AfterCreate {
		if err := h.AfterCreate(ctx, record); err != nil {
			return fmt.Errorf("after_create: %w", err)
		}
	}
	return nil
}

// RunBeforeUpdate executes the pre-persist stages of the Update pipeline.
// prev is the current record state before the update is applied.
func (p *Pipeline) RunBeforeUpdate(pctx *UpdateContext, current *def.EntityRecord) (*def.EntityRecord, error) {
	es, err := p.lookupSchema(pctx.EntityName)
	if err != nil {
		return nil, err
	}

	// Build the proposed record: current values + patch.
	proposed := &def.EntityRecord{
		ID:           current.ID,
		TenantID:     current.TenantID,
		EntityName:   current.EntityName,
		Data:         mergeMaps(current.Data, pctx.Data),
		CustomFields: mergeMaps(current.CustomFields, pctx.CustomFields),
		CreatedAt:    current.CreatedAt,
		Meta: def.RecordMeta{
			Actor:         pctx.Actor,
			OperationType: def.OperationUpdate,
		},
	}

	// Check immutable fields.
	ve := &ValidationError{}
	for field := range pctx.Data {
		if es.ImmutableFields[field] {
			ve.AddField(field, fmt.Sprintf("%s cannot be changed after creation", field))
		}
	}
	if !ve.IsEmpty() {
		return nil, ve
	}

	hooks := es.Def.EntityHooks()

	for _, h := range hooks.BeforeValidate {
		if err := h.BeforeValidate(pctx.Ctx, proposed); err != nil {
			return nil, fmt.Errorf("before_validate: %w", err)
		}
	}

	if err := p.validateFields(pctx.Ctx, proposed, es); err != nil {
		return nil, err
	}

	for _, h := range hooks.BeforeSave {
		if err := h.BeforeSave(pctx.Ctx, proposed); err != nil {
			return nil, fmt.Errorf("before_save: %w", err)
		}
	}

	for _, h := range hooks.BeforeUpdate {
		if err := h.BeforeUpdate(pctx.Ctx, proposed, current); err != nil {
			return nil, fmt.Errorf("before_update: %w", err)
		}
	}

	return proposed, nil
}

// RunAfterUpdate executes the post-persist stages of the Update pipeline.
// Called from within the open database transaction.
func (p *Pipeline) RunAfterUpdate(ctx context.Context, record, prev *def.EntityRecord) error {
	es, err := p.lookupSchema(record.EntityName)
	if err != nil {
		return err
	}
	hooks := es.Def.EntityHooks()

	for _, h := range hooks.AfterSave {
		if err := h.AfterSave(ctx, record); err != nil {
			return fmt.Errorf("after_save: %w", err)
		}
	}
	for _, h := range hooks.AfterUpdate {
		if err := h.AfterUpdate(ctx, record, prev); err != nil {
			return fmt.Errorf("after_update: %w", err)
		}
	}
	return nil
}

// RunBeforeDelete executes the pre-delete stages.
func (p *Pipeline) RunBeforeDelete(ctx context.Context, record *def.EntityRecord) error {
	es, err := p.lookupSchema(record.EntityName)
	if err != nil {
		return err
	}
	hooks := es.Def.EntityHooks()
	for _, h := range hooks.BeforeDelete {
		if err := h.BeforeDelete(ctx, record); err != nil {
			return fmt.Errorf("before_delete: %w", err)
		}
	}
	return nil
}

// RunAfterDelete executes the post-delete stages from within the transaction.
func (p *Pipeline) RunAfterDelete(ctx context.Context, record *def.EntityRecord) error {
	es, err := p.lookupSchema(record.EntityName)
	if err != nil {
		return err
	}
	hooks := es.Def.EntityHooks()
	for _, h := range hooks.AfterDelete {
		if err := h.AfterDelete(ctx, record); err != nil {
			return fmt.Errorf("after_delete: %w", err)
		}
	}
	return nil
}

// --- Internal helpers ---

func (p *Pipeline) lookupSchema(entityName string) (*compiler.EntitySchema, error) {
	es := p.schema.ByName[entityName]
	if es == nil {
		return nil, fmt.Errorf("runtime: entity %q not found in compiled schema", entityName)
	}
	return es, nil
}

func (p *Pipeline) applyDefaults(record *def.EntityRecord, es *compiler.EntitySchema) {
	for name, defaultFn := range es.DefaultValues {
		if _, exists := record.Data[name]; !exists {
			if record.Data == nil {
				record.Data = make(map[string]any)
			}
			record.Data[name] = defaultFn()
		}
	}
}

func (p *Pipeline) validateFields(ctx context.Context, record *def.EntityRecord, es *compiler.EntitySchema) error {
	ve := &ValidationError{}

	for name := range es.RequiredFields {
		val := record.Get(name)
		if val == nil || val == "" {
			ve.AddField(name, "this field is required")
		}
	}

	for _, f := range es.Def.EntityFields() {
		val := record.Get(f.Name)
		for _, validator := range f.Validators {
			if msg := validator(ctx, val); msg != "" {
				ve.AddField(f.Name, msg)
			}
		}
	}

	if !ve.IsEmpty() {
		return ve
	}
	return nil
}

func cloneMap(m map[string]any) map[string]any {
	if m == nil {
		return nil
	}
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// mergeMaps returns a new map with base values overlaid by patch.
func mergeMaps(base, patch map[string]any) map[string]any {
	out := make(map[string]any, len(base)+len(patch))
	for k, v := range base {
		out[k] = v
	}
	for k, v := range patch {
		out[k] = v
	}
	return out
}
