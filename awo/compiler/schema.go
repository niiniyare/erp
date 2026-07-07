package compiler

import (
	"awo.so/awo/def"
	"awo.so/awo/registry"
)

// entityAPIResource derives the plural URL path segment for a local entity name.
// "organization" → "organizations", "org_assignment" → "org_assignments", "entry" → "entries".
func entityAPIResource(localName string) string {
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
type EntitySchema struct {
	// Def is the original EntityDefinition. Never mutated after compilation.
	Def def.EntityDefinition

	// QualifiedName is the globally unique identifier: module + "_" + local name.
	// e.g. "platform_organization", "iam_user", "finance_invoice".
	// Used as: DB table name (system entities), Casbin object, Temporal namespace,
	// Redis cache key prefix, event namespace.
	QualifiedName string

	// LocalName is the module-local identifier (EntityDefinition.Name without prefix).
	// e.g. "organization", "user", "invoice".
	LocalName string

	// Module is the owning module (EntityDefinition.Module).
	// e.g. "platform", "iam", "finance".
	Module string

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

	// FieldsByName provides O(1) lookup of FieldDef by name.
	FieldsByName map[string]def.FieldDef

	// EdgesByName provides O(1) lookup of EdgeDef by name.
	EdgesByName map[string]def.EdgeDef

	// ActionsByName provides O(1) lookup of ActionDef by name.
	ActionsByName map[string]def.ActionDef

	// DefaultValues maps field names to their resolved default value functions.
	// Nil entry means no default — field is zero-valued on omission.
	DefaultValues map[string]func() any

	// RequiredFields is the set of field names that are Required: true.
	RequiredFields map[string]bool

	// ImmutableFields is the set of field names that are Immutable: true.
	ImmutableFields map[string]bool

	// SensitiveFields is the set of field names that are Sensitive: true.
	SensitiveFields map[string]bool

	// SearchableFields is the set of field names that are Searchable: true.
	SearchableFields map[string]bool

	// LinkTargets maps FieldTypeLink field names to their target EntitySchema.
	// Resolved at compile time; all link targets are guaranteed to exist.
	LinkTargets map[string]*EntitySchema

	// TableName is the PostgreSQL table name (equals QualifiedName for system
	// entities; "custom_entity_records" for custom entities with entity_type filter).
	TableName string
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

	// EntityName is the QualifiedName of the entity this route operates on.
	// e.g. "finance_invoice".
	EntityName string

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
		for _, f := range es.Def.EntityFields() {
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
	resource := entityAPIResource(local)

	es := &EntitySchema{
		Def:              d,
		QualifiedName:    qname,
		LocalName:        local,
		Module:           module,
		APIResource:      resource,
		RoutePrefix:      "/api/v1/" + module + "/" + resource,
		OpenAPITag:       def.OpenAPITag(module),
		FieldsByName:     make(map[string]def.FieldDef),
		EdgesByName:      make(map[string]def.EdgeDef),
		ActionsByName:    make(map[string]def.ActionDef),
		DefaultValues:    make(map[string]func() any),
		RequiredFields:   make(map[string]bool),
		ImmutableFields:  make(map[string]bool),
		SensitiveFields:  make(map[string]bool),
		SearchableFields: make(map[string]bool),
		LinkTargets:      make(map[string]*EntitySchema),
	}

	// Table name: system entities use QualifiedName as table name;
	// custom entities share "custom_entity_records" filtered by entity_type.
	if d.IsSystem() {
		es.TableName = qname
	} else {
		es.TableName = "custom_entity_records"
	}

	for _, f := range d.EntityFields() {
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
	}

	for _, e := range d.EntityEdges() {
		es.EdgesByName[e.Name] = e
	}

	for _, a := range d.EntityActions() {
		es.ActionsByName[a.Name] = a
	}

	return es
}

func emitRoutes(es *EntitySchema) []RouteDescriptor {
	qname := es.QualifiedName
	base := es.RoutePrefix

	routes := []RouteDescriptor{
		{Method: "GET", Path: base, EntityName: qname, Operation: "list", RequiredPermission: "read"},
		{Method: "GET", Path: base + "/:id", EntityName: qname, Operation: "get", RequiredPermission: "read"},
		{Method: "POST", Path: base, EntityName: qname, Operation: "create", RequiredPermission: "create"},
		{Method: "PATCH", Path: base + "/:id", EntityName: qname, Operation: "update", RequiredPermission: "write"},
		{Method: "DELETE", Path: base + "/:id", EntityName: qname, Operation: "delete", RequiredPermission: "delete"},
	}

	for _, action := range es.Def.EntityActions() {
		method := string(action.Method)
		if method == "" {
			method = "POST"
		}
		perm := action.Permission
		if perm == "" {
			perm = "write"
		}
		routes = append(routes, RouteDescriptor{
			Method:             method,
			Path:               base + "/:id/" + action.Name,
			EntityName:         qname,
			Operation:          "action",
			ActionName:         action.Name,
			RequiredPermission: perm,
		})
	}

	return routes
}

func emitPolicies(es *EntitySchema) []CasbinPolicy {
	perms := es.Def.EntityPermissions()
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
