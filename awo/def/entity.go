package def

// EntityDefinition is the interface implemented by both [SystemDefinition]
// and [CustomDefinition]. The registry, compiler, and runtime work against
// this interface so they do not need to distinguish between entity kinds for
// most operations.
type EntityDefinition interface {
	// EntityName returns the stable snake_case name. Format: {module}_{noun}.
	// Never rename after data is persisted.
	EntityName() string

	// EntityModule returns the module this entity belongs to (e.g. "finance").
	EntityModule() string

	// EntityLabel returns the human-readable singular label.
	EntityLabel() string

	// EntityLabelPlural returns the human-readable plural label.
	EntityLabelPlural() string

	// EntityFields returns all field definitions.
	EntityFields() []FieldDef

	// EntityEdges returns all edge definitions.
	EntityEdges() []EdgeDef

	// EntityHooks returns the hook set.
	EntityHooks() HookSet

	// EntityPermissions returns the permission set.
	EntityPermissions() PermissionSet

	// EntityActions returns custom action definitions.
	EntityActions() []ActionDef

	// EntityWorkflowTriggers returns workflow trigger bindings.
	EntityWorkflowTriggers() []WorkflowTrigger

	// EntityPageBuilders returns optional SDUI page builder overrides.
	EntityPageBuilders() PageBuilderSet

	// IsSystem returns true for SQL-backed system entities (typed columns),
	// false for JSONB-backed custom entities.
	IsSystem() bool
}

// SystemDefinition declares a system entity: Go-struct-backed, typed SQL
// columns, full financial and IAM eligibility.
//
// System entities are mandatory for: ledger entries, stock moves, payments,
// users, tenants, journal entries, and tax entries. Use [CustomDefinition]
// for all other entity shapes.
type SystemDefinition struct {
	// Name is the stable entity identifier. Format: {module}_{noun}.
	// Embedded in migration filenames, Temporal workflow IDs (stored for
	// years), and Redis cache keys. Never rename.
	Name string

	// Module is the business domain this entity belongs to (e.g. "finance",
	// "inventory", "hr"). Used for route namespacing and Casbin policy scope.
	Module string

	// Label is the human-readable singular display name (e.g. "Invoice").
	Label string

	// LabelPlural is the human-readable plural display name (e.g. "Invoices").
	LabelPlural string

	// Fields declares all typed columns for this entity.
	Fields []FieldDef

	// Edges declares all relationships to other entities.
	Edges []EdgeDef

	// Hooks declares the lifecycle hook implementations.
	Hooks HookSet

	// Permissions declares the RBAC gates for CRUD operations.
	Permissions PermissionSet

	// Actions declares custom action routes beyond standard CRUD.
	Actions []ActionDef

	// WorkflowTriggers declares Temporal workflow starts bound to lifecycle
	// events.
	WorkflowTriggers []WorkflowTrigger

	// PageBuilders optionally overrides auto-generated SDUI page schemas.
	PageBuilders PageBuilderSet
}

// Ensure SystemDefinition implements EntityDefinition at compile time.
var _ EntityDefinition = (*SystemDefinition)(nil)

func (d *SystemDefinition) EntityName() string                   { return d.Name }
func (d *SystemDefinition) EntityModule() string                 { return d.Module }
func (d *SystemDefinition) EntityLabel() string                  { return d.Label }
func (d *SystemDefinition) EntityLabelPlural() string            { return d.LabelPlural }
func (d *SystemDefinition) EntityFields() []FieldDef             { return d.Fields }
func (d *SystemDefinition) EntityEdges() []EdgeDef               { return d.Edges }
func (d *SystemDefinition) EntityHooks() HookSet                 { return d.Hooks }
func (d *SystemDefinition) EntityPermissions() PermissionSet     { return d.Permissions }
func (d *SystemDefinition) EntityActions() []ActionDef           { return d.Actions }
func (d *SystemDefinition) EntityWorkflowTriggers() []WorkflowTrigger {
	return d.WorkflowTriggers
}
func (d *SystemDefinition) EntityPageBuilders() PageBuilderSet { return d.PageBuilders }
func (d *SystemDefinition) IsSystem() bool                     { return true }

// CustomDefinition declares a custom entity: JSONB-backed, tenant-specific
// schema, extensible at runtime via the Metadata module.
//
// Use for tenant-specific data that evolves frequently, participates in no
// financial or inventory accounting, and has low write rates.
//
// Escalate to SystemDefinition when: >10M records, fields used in financial
// calculations, or FK constraints to system entity PKs are required.
type CustomDefinition struct {
	// Name is the stable entity identifier. Same naming rules as
	// SystemDefinition.Name.
	Name string

	// Module is the business domain.
	Module string

	// Label is the human-readable singular display name.
	Label string

	// LabelPlural is the human-readable plural display name.
	LabelPlural string

	// Fields declares the logical fields. Each field maps to a key in the
	// JSONB document. The compiler generates GIN indexes for Searchable fields.
	Fields []FieldDef

	// Edges declares relationships to other entities.
	Edges []EdgeDef

	// Hooks declares the lifecycle hook implementations.
	Hooks HookSet

	// Permissions declares the RBAC gates.
	Permissions PermissionSet

	// Actions declares custom action routes.
	Actions []ActionDef

	// WorkflowTriggers declares Temporal workflow starts.
	WorkflowTriggers []WorkflowTrigger

	// PageBuilders optionally overrides auto-generated SDUI pages.
	PageBuilders PageBuilderSet
}

// Ensure CustomDefinition implements EntityDefinition at compile time.
var _ EntityDefinition = (*CustomDefinition)(nil)

func (d *CustomDefinition) EntityName() string                   { return d.Name }
func (d *CustomDefinition) EntityModule() string                 { return d.Module }
func (d *CustomDefinition) EntityLabel() string                  { return d.Label }
func (d *CustomDefinition) EntityLabelPlural() string            { return d.LabelPlural }
func (d *CustomDefinition) EntityFields() []FieldDef             { return d.Fields }
func (d *CustomDefinition) EntityEdges() []EdgeDef               { return d.Edges }
func (d *CustomDefinition) EntityHooks() HookSet                 { return d.Hooks }
func (d *CustomDefinition) EntityPermissions() PermissionSet     { return d.Permissions }
func (d *CustomDefinition) EntityActions() []ActionDef           { return d.Actions }
func (d *CustomDefinition) EntityWorkflowTriggers() []WorkflowTrigger {
	return d.WorkflowTriggers
}
func (d *CustomDefinition) EntityPageBuilders() PageBuilderSet { return d.PageBuilders }
func (d *CustomDefinition) IsSystem() bool                     { return false }
