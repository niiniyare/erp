package compiler

import (
	"awo.so/awo/def"
	"awo.so/awo/registry"
)

// entityAPIResource derives the plural URL path segment for a local entity name.
// Respects an explicit PluralName override when set.
// "organization" → "organizations", "org_assignment" → "org_assignments", "entry" → "entries".
func entityAPIResource(d def.EntityDefinition, localName string) string {
	if p := d.EntityPluralName(); p != "" {
		return p
	}
	return def.PluralizeLocal(localName)
}

// CompiledSchema is the immutable output of the compilation phase. It is the
// single source of truth for all runtime subsystems. Every subsystem receives
// a *CompiledSchema at startup and uses it for the lifetime of the process.
//
// CompiledSchema is safe for concurrent read access. It is never mutated after
// [Compile] returns.
type CompiledSchema struct {
	// Entities is the ordered list of compiled entity schemas.
	Entities []*EntitySchema

	// ByName provides O(1) lookup of EntitySchema by entity name.
	ByName map[string]*EntitySchema

	// Routes is the flat list of all auto-generated HTTP route descriptors,
	// consumed by the API layer to register Fiber routes.
	Routes []RouteDescriptor

	// CasbinPolicies is the list of Casbin policy assertions derived from
	// PermissionSet declarations. Consumed by the IAM module at startup.
	CasbinPolicies []CasbinPolicy

	// Diagnostics contains warnings and informational messages from the
	// compilation phase. Error-severity diagnostics cause Compile to return
	// an error; warnings are preserved here for tooling.
	Diagnostics Diagnostics
}

// EntitySchema is the compiled representation of a single EntityDefinition.
// It is the sole runtime metadata authority for its entity — no runtime code
// should call through to the original EntityDefinition after compilation.
type EntitySchema struct {
	// ── Identity ─────────────────────────────────────────────────────────────

	// QualifiedName is the globally unique identifier: module + "_" + local name.
	// e.g. "platform_organization", "iam_user", "finance_invoice".
	// Used as: DB table name (system entities), Casbin object, metric label.
	QualifiedName string

	// LocalName is the module-local identifier (EntityDefinition.Name without prefix).
	// e.g. "organization", "user", "invoice".
	LocalName string

	// Module is the owning module (EntityDefinition.Module).
	// e.g. "platform", "iam", "finance".
	Module string

	// IsSystem is true for SQL-backed system entities; false for JSONB custom entities.
	IsSystem bool

	// ── Display ──────────────────────────────────────────────────────────────

	// Label is the human-readable singular display name (e.g. "Invoice").
	Label string

	// LabelPlural is the human-readable plural display name (e.g. "Invoices").
	LabelPlural string

	// Description is the optional entity description for docs and OpenAPI.
	Description string

	// ── HTTP / API ───────────────────────────────────────────────────────────

	// APIResource is the plural URL path segment for this entity.
	// e.g. "organizations", "users", "invoices", "org_assignments".
	APIResource string

	// RoutePrefix is the base HTTP path for this entity's CRUD routes.
	// Format: /api/v1/{module}/{plural_local_name}
	// e.g. "/api/v1/platform/organizations", "/api/v1/iam/users".
	RoutePrefix string

	// OpenAPITag is the human-readable OpenAPI tag for this entity's module.
	// e.g. "Platform", "Iam", "Finance".
	OpenAPITag string

	// APISingular is the singular local name used in URL path segments.
	// e.g. "organization", "user", "invoice".
	APISingular string

	// ── Namespace identifiers (all derived once at compile time) ──────────────

	// EventNamespace is the dot-separated namespace for domain events.
	// Format: module + "." + local_name  (e.g. "iam.user", "finance.invoice").
	EventNamespace string

	// WorkflowNamespace is the Temporal namespace prefix for this entity's workflows.
	// Format: module + "." + local_name  (e.g. "iam.user", "finance.invoice").
	WorkflowNamespace string

	// PermissionNamespace is the Casbin object namespace for this entity.
	// Equals QualifiedName  (e.g. "iam_user", "finance_invoice").
	PermissionNamespace string

	// MetricNamespace is the Prometheus metric label prefix.
	// Format: module + "_" + local_name  (equals QualifiedName).
	// e.g. "iam_user", "finance_invoice".
	MetricNamespace string

	// CacheNamespace is the Redis key namespace prefix.
	// Format: module + ":" + local_name  (e.g. "iam:user", "finance:invoice").
	CacheNamespace string

	// TableName is the PostgreSQL table name (equals QualifiedName for system
	// entities; "custom_entity_records" for custom entities with entity_type filter).
	TableName string

	// ── Structural metadata (ordered slices for deterministic iteration) ──────

	// Fields is the ordered list of field definitions, mirroring EntityDefinition.Fields.
	Fields []def.FieldDef

	// Edges is the ordered list of edge definitions, mirroring EntityDefinition.Edges.
	Edges []def.EdgeDef

	// Actions is the ordered list of custom action definitions.
	Actions []def.ActionDef

	// Permissions is the RBAC permission set for this entity.
	Permissions def.PermissionSet

	// ── O(1) lookup maps ─────────────────────────────────────────────────────

	// FieldsByName provides O(1) lookup of FieldDef by name.
	FieldsByName map[string]def.FieldDef

	// EdgesByName provides O(1) lookup of EdgeDef by name.
	EdgesByName map[string]def.EdgeDef

	// ActionsByName provides O(1) lookup of ActionDef by name.
	ActionsByName map[string]def.ActionDef

	// DefaultValues maps field names to their resolved default value functions.
	// Nil entry means no default — field is zero-valued on omission.
	DefaultValues map[string]func() any

	// ── Field constraint sets (boolean index maps) ────────────────────────────

	// RequiredFields is the set of field names that are Required: true.
	RequiredFields map[string]bool

	// ImmutableFields is the set of field names that are Immutable: true.
	ImmutableFields map[string]bool

	// SensitiveFields is the set of field names that are Sensitive: true.
	SensitiveFields map[string]bool

	// SearchableFields is the set of field names that are Searchable: true.
	SearchableFields map[string]bool

	// FieldValidators maps field names to their custom validator functions.
	// Extracted once at compile time from FieldDef.Validators.
	// Runtime validation iterates this map instead of ranging over Fields.
	FieldValidators map[string][]def.FieldValidator

	// LinkTargets maps FieldTypeLink field names to their target EntitySchema.
	// Resolved at compile time; all link targets are guaranteed to exist.
	LinkTargets map[string]*EntitySchema

	// FieldLookups maps FieldTypeLink / FieldTypeLinkList field names to their
	// compiled lookup metadata. Consumed by the SDUI generator to emit amis
	// select controls with server-side search. Populated in Phase 2.5.
	FieldLookups map[string]*CompiledLookup

	// ── Runtime lifecycle data ────────────────────────────────────────────────

	// Hooks is the lifecycle hook set, extracted from EntityDefinition at
	// compile time. Runtime code reads this directly — never call through def.
	Hooks def.HookSet

	// WorkflowTriggers is the list of Temporal workflow bindings.
	// Extracted once at compile time. Use directly in the runtime; never call
	// es.def.EntityWorkflowTriggers() after startup.
	WorkflowTriggers []def.WorkflowTrigger

	// PageBuilders holds optional SDUI page builder overrides. When a builder
	// is set for a view kind, it replaces the auto-generated schema for that view.
	PageBuilders def.PageBuilderSet

	// ── Compile-time reference (not for runtime use) ──────────────────────────

	// def is the original EntityDefinition, retained for compile-phase link
	// resolution only. Runtime subsystems must use the promoted fields above.
	// Accessing def after Compile returns is an architectural violation.
	def def.EntityDefinition
}

// CompiledLookup carries all metadata the SDUI generator needs to render a
// FieldTypeLink field as an amis select control with server-side search.
// It is derived entirely from compile-time information; no runtime entity
// knowledge is hard-coded.
type CompiledLookup struct {
	// TargetQualifiedName is the fully-qualified name of the linked entity
	// (e.g. "finance_currency", "iam_user").
	TargetQualifiedName string

	// TargetLabel is the human-readable singular label of the target entity.
	TargetLabel string

	// SearchURL is the amis API endpoint for autocomplete search.
	// Format: /api/v1/{module}/{resource}?q=${keywords}&tenant_id=${tenant_id}
	SearchURL string

	// ValueField is the field name to use as the option value (always "id").
	ValueField string

	// LabelField is the field name to display as the option label.
	// Derived from the target entity's first Searchable or Data field named
	// "name", "code", or "title". Falls back to "id" when none are found.
	LabelField string

	// Multiple is true when the source field type is FieldTypeLinkList,
	// producing a multi-select control.
	Multiple bool
}

// RouteDescriptor describes a single HTTP route generated from an EntityDefinition.
// The API layer uses this to register Fiber route handlers.
type RouteDescriptor struct {
	// Method is the HTTP method (GET, POST, PATCH, DELETE).
	Method string

	// Path is the Fiber route path.
	// Format: /api/v1/{module}/{plural_local_name}[/:id][/{action}]
	// e.g. "/api/v1/finance/invoices/:id".
	Path string

	// EntityQualifiedName is the globally-unique entity identifier.
	// e.g. "finance_invoice". Use this to look up the EntitySchema via
	// CompiledSchema.ByName.
	EntityQualifiedName string

	// Module is the entity's owning module. e.g. "finance", "iam".
	Module string

	// Resource is the plural URL path segment. e.g. "invoices", "users".
	Resource string

	// Operation identifies the semantic operation (list, get, create, update,
	// delete, action).
	Operation string

	// ActionName is non-empty for custom action routes.
	ActionName string

	// RequiredPermission is the Casbin action string checked before executing
	// the route handler (e.g. "read", "write", "create", "delete", "submit").
	RequiredPermission string
}

// CasbinPolicy is a Casbin policy assertion derived from a PermissionSet.
type CasbinPolicy struct {
	// Subject is the role or user (e.g. "role:tenant.admin").
	Subject string

	// Object is the entity name.
	Object string

	// Action is the Casbin action ("read", "write", "create", "delete", or a
	// custom action name).
	Action string
}

// Compile transforms the sealed registry into a CompiledSchema.
// Returns an error if semantic analysis finds issues not caught by the registry
// validator (e.g. circular edge references, impossible default values).
//
// On non-recoverable errors the caller should treat the error as fatal.
func Compile(reg *registry.Registry) (*CompiledSchema, error) {
	c := &compiler{reg: reg}
	return c.compile()
}

type compiler struct {
	reg *registry.Registry
}

func (c *compiler) compile() (*CompiledSchema, error) {
	// Validate before building schema structures.
	ds := Validate(c.reg)
	if ds.HasErrors() {
		return nil, ds.AsError()
	}

	defs := c.reg.All()

	schema := &CompiledSchema{
		Entities:    make([]*EntitySchema, 0, len(defs)),
		ByName:      make(map[string]*EntitySchema, len(defs)),
		Diagnostics: ds,
	}

	// Phase 1: build EntitySchema stubs (without link resolution).
	for _, d := range defs {
		es := buildEntitySchema(d)
		schema.Entities = append(schema.Entities, es)
		schema.ByName[es.QualifiedName] = es
	}

	// Phase 2: resolve link targets (LinkTarget is always a QualifiedName).
	for _, es := range schema.Entities {
		for _, f := range es.Fields {
			if f.Type == def.FieldTypeLink || f.Type == def.FieldTypeLinkList {
				target, ok := schema.ByName[f.LinkTarget]
				if !ok {
					// Already validated by registry; should not happen.
					continue
				}
				es.LinkTargets[f.Name] = target
			}
		}
	}

	// Phase 2.5: build CompiledLookup for every Link / LinkList field.
	for _, es := range schema.Entities {
		for _, f := range es.Fields {
			if f.Type != def.FieldTypeLink && f.Type != def.FieldTypeLinkList {
				continue
			}
			target, ok := schema.ByName[f.LinkTarget]
			if !ok {
				continue
			}
			es.FieldLookups[f.Name] = buildLookup(target, f.Type == def.FieldTypeLinkList)
		}
	}

	// Phase 3: emit routes.
	for _, es := range schema.Entities {
		schema.Routes = append(schema.Routes, emitRoutes(es)...)
	}

	// Phase 4: emit Casbin policies.
	for _, es := range schema.Entities {
		schema.CasbinPolicies = append(schema.CasbinPolicies, emitPolicies(es)...)
	}

	return schema, nil
}

func buildEntitySchema(d def.EntityDefinition) *EntitySchema {
	qname := def.QualifiedName(d)
	local := def.LocalName(d)
	module := d.EntityModule()
	resource := entityAPIResource(d, local)
	dotNS := module + "." + local

	es := &EntitySchema{
		def:                 d,
		QualifiedName:       qname,
		LocalName:           local,
		Module:              module,
		IsSystem:            d.IsSystem(),
		Label:               d.EntityLabel(),
		LabelPlural:         d.EntityLabelPlural(),
		Description:         d.EntityDescription(),
		APIResource:         resource,
		APISingular:         local,
		RoutePrefix:         "/api/v1/" + module + "/" + resource,
		OpenAPITag:          def.OpenAPITag(module),
		EventNamespace:      dotNS,
		WorkflowNamespace:   dotNS,
		PermissionNamespace: qname,
		MetricNamespace:     qname,
		CacheNamespace:      module + ":" + local,
		Fields:           d.EntityFields(),
		Edges:            d.EntityEdges(),
		Actions:          d.EntityActions(),
		Permissions:      d.EntityPermissions(),
		Hooks:            d.EntityHooks(),
		WorkflowTriggers: d.EntityWorkflowTriggers(),
		PageBuilders:     d.EntityPageBuilders(),
		FieldsByName:     make(map[string]def.FieldDef),
		EdgesByName:      make(map[string]def.EdgeDef),
		ActionsByName:    make(map[string]def.ActionDef),
		FieldValidators:  make(map[string][]def.FieldValidator),
		DefaultValues:    make(map[string]func() any),
		RequiredFields:   make(map[string]bool),
		ImmutableFields:  make(map[string]bool),
		SensitiveFields:  make(map[string]bool),
		SearchableFields: make(map[string]bool),
		LinkTargets:      make(map[string]*EntitySchema),
		FieldLookups:     make(map[string]*CompiledLookup),
	}

	// Table name: system entities use QualifiedName as table name;
	// custom entities share "custom_entity_records" filtered by entity_type.
	if d.IsSystem() {
		es.TableName = qname
	} else {
		es.TableName = "custom_entity_records"
	}

	for _, f := range es.Fields {
		es.FieldsByName[f.Name] = f
		if f.Required {
			es.RequiredFields[f.Name] = true
		}
		if f.Immutable {
			es.ImmutableFields[f.Name] = true
		}
		if f.Sensitive {
			es.SensitiveFields[f.Name] = true
		}
		if f.Searchable {
			es.SearchableFields[f.Name] = true
		}
		if f.Default != nil {
			es.DefaultValues[f.Name] = f.Default
		}
		if len(f.Validators) > 0 {
			es.FieldValidators[f.Name] = f.Validators
		}
	}

	for _, e := range es.Edges {
		es.EdgesByName[e.Name] = e
	}

	for _, a := range es.Actions {
		es.ActionsByName[a.Name] = a
	}

	return es
}

// buildLookup constructs a CompiledLookup for a Link field targeting target.
func buildLookup(target *EntitySchema, multiple bool) *CompiledLookup {
	searchURL := target.RoutePrefix + "?q=${keywords}"

	// Heuristic: pick the best label field from the target entity.
	labelField := "id"
	priority := []string{"name", "code", "title", "label", "full_name", "account_name"}
	for _, candidate := range priority {
		if _, ok := target.FieldsByName[candidate]; ok {
			labelField = candidate
			break
		}
	}
	// Also accept any field marked Searchable as a fallback.
	if labelField == "id" {
		for name := range target.SearchableFields {
			labelField = name
			break
		}
	}

	return &CompiledLookup{
		TargetQualifiedName: target.QualifiedName,
		TargetLabel:         target.Label,
		SearchURL:           searchURL,
		ValueField:          "id",
		LabelField:          labelField,
		Multiple:            multiple,
	}
}

func emitRoutes(es *EntitySchema) []RouteDescriptor {
	qname := es.QualifiedName
	base := es.RoutePrefix

	routes := []RouteDescriptor{
		{Method: "GET", Path: base, EntityQualifiedName: qname, Module: es.Module, Resource: es.APIResource, Operation: "list", RequiredPermission: "read"},
		{Method: "GET", Path: base + "/:id", EntityQualifiedName: qname, Module: es.Module, Resource: es.APIResource, Operation: "get", RequiredPermission: "read"},
		{Method: "POST", Path: base, EntityQualifiedName: qname, Module: es.Module, Resource: es.APIResource, Operation: "create", RequiredPermission: "create"},
		{Method: "PATCH", Path: base + "/:id", EntityQualifiedName: qname, Module: es.Module, Resource: es.APIResource, Operation: "update", RequiredPermission: "write"},
		{Method: "DELETE", Path: base + "/:id", EntityQualifiedName: qname, Module: es.Module, Resource: es.APIResource, Operation: "delete", RequiredPermission: "delete"},
	}

	for _, action := range es.Actions {
		method := string(action.Method)
		if method == "" {
			method = "POST"
		}
		perm := action.Permission
		if perm == "" {
			perm = "write"
		}
		routes = append(routes, RouteDescriptor{
			Method:              method,
			Path:                base + "/:id/" + action.Name,
			EntityQualifiedName: qname,
			Module:              es.Module,
			Resource:            es.APIResource,
			Operation:           "action",
			ActionName:          action.Name,
			RequiredPermission:  perm,
		})
	}

	return routes
}

func emitPolicies(es *EntitySchema) []CasbinPolicy {
	perms := es.Permissions
	name := es.QualifiedName

	var policies []CasbinPolicy
	for _, subject := range perms.Create {
		policies = append(policies, CasbinPolicy{Subject: subject, Object: name, Action: "create"})
	}
	for _, subject := range perms.Read {
		policies = append(policies, CasbinPolicy{Subject: subject, Object: name, Action: "read"})
	}
	for _, subject := range perms.Write {
		policies = append(policies, CasbinPolicy{Subject: subject, Object: name, Action: "write"})
	}
	for _, subject := range perms.Delete {
		policies = append(policies, CasbinPolicy{Subject: subject, Object: name, Action: "delete"})
	}
	for actionName, subjects := range perms.Actions {
		for _, subject := range subjects {
			policies = append(policies, CasbinPolicy{Subject: subject, Object: name, Action: actionName})
		}
	}

	return policies
}
