package def

import (
	"errors"
	"fmt"

	"awo.so/framework/platform/org"
)

// EntityDefinition is the authoritative meta-model for one entity type.
//
// It drives every layer of the framework automatically:
//   - Persistence  — table name, column list, insert/update/delete SQL
//   - API          — REST CRUD routes with pagination and filtering
//   - SDUI         — AMIS page and form schemas
//   - Privacy      — policy chain evaluation
//   - Hooks        — before/after lifecycle callbacks
//   - Migration    — generated UP/DOWN SQL with RLS policies
//   - Audit log    — automatic write tracking when Audited = true
//
// EntityDefinitions MUST be immutable after registration. Mutating a definition
// after calling def.Register produces undefined behaviour because the
// registry stores a pointer and the API/persistence layers cache field lists.
type EntityDefinition struct {
	// ── Identity ──────────────────────────────────────────────────────────────

	// Type controls which persistence backend stores this entity's records.
	// Defaults to EntityTypeSystem (dedicated SQL table) when zero.
	// Set to EntityTypeCustom for tenant-defined runtime entities stored in JSONB.
	Type EntityType

	// Name is the canonical machine identifier used in URLs, DB tables, and logs.
	// Convention: singular, snake_case. Must be unique across the registry.
	// Example: "sales_order", "finance_account"
	Name string

	// Label is the human-readable display name shown in SDUI navigation and forms.
	// Example: "Sales Order", "Chart of Accounts"
	Label string

	// LabelPlural is the plural form of Label used in list views and navigation.
	// Defaults to Label+"s" when empty.
	// Example: "Sales Orders", "Chart of Accounts" (irregular — must be explicit)
	LabelPlural string

	// Description is shown in generated API documentation and SDUI tooltips.
	Description string

	// Module is the logical grouping for SDUI sidebar navigation.
	// Example: "Finance", "HR", "CRM", "Platform"
	Module string

	// ── Storage ───────────────────────────────────────────────────────────────

	// Table is the PostgreSQL table name.
	// Defaults to "{Name}s" when empty (e.g. "finance_account" → "finance_accounts").
	Table string

	// OrgScope determines how rows are scoped within the organisational hierarchy.
	// The persistence layer generates the correct columns, WHERE clauses, and RLS
	// policies based on this value.
	//
	//   ScopeLevelGlobal — no tenant_id or org_unit_id column.
	//                       Rows are shared across all tenants.
	//                       Example: currencies, countries, language codes.
	//
	//   ScopeLevelTenant — rows carry tenant_id only; visible to all org units
	//                       within the tenant.
	//                       Example: users, roles, subscription features.
	//
	//   ScopeLevelUnit   — rows carry tenant_id + org_unit_id.
	//                       Access is tree-based: viewer's unit must be an
	//                       ancestor-or-equal of the record's unit.
	//                       Example: GL accounts, invoices, employees, budgets.
	//
	// Defaults to ScopeLevelTenant when zero-value.
	OrgScope org.ScopeLevel

	// ── Schema ────────────────────────────────────────────────────────────────

	// Fields declares all scalar and relational attributes.
	// Order matters: it determines column order in generated SQL and field order
	// in generated SDUI forms.
	Fields []*FieldDef

	// Edges declares named relationships to other EntityDefinitions.
	// Used by SDUI link pickers and API response embedding (future).
	Edges []*EdgeDef

	// ── Behaviour ─────────────────────────────────────────────────────────────

	// Hooks are lifecycle callbacks invoked synchronously within write transactions.
	// Hooks run in registration order. A BeforeHook error aborts the transaction.
	Hooks []HookDef

	// Policies control read/write access per operation.
	// Evaluated in order; fail-closed (deny if no policy returns ErrAllow).
	// See def.PolicyFunc for the full contract.
	Policies []PolicyDef

	// SoftDelete enables soft-delete via a `deleted_at timestamptz` column.
	// When true: Delete sets deleted_at rather than removing the row; FindByID,
	// List, and Exists automatically exclude deleted records.
	SoftDelete bool

	// Audited enables automatic audit log entries on every Create/Update/Delete.
	// The audit log records actor, tenant/company/division, timestamp, and a
	// JSON diff of changed fields.
	Audited bool

	// NamingSeries configures automatic document numbering for this entity.
	// When non-nil, the naming package stamps a generated series value onto
	// the target field during Create (before RunBefore hooks).
	// Example: &NamingSeries{Field: "name", Prefix: "INV-", Padding: 5}
	// produces "INV-00001", "INV-00002", …
	NamingSeries *NamingSeriesDef

	// EntityValidators are cross-field validators run after all per-field
	// validators pass. Each receives the full record and returns a slice
	// of FieldErrors (may reference multiple fields).
	// Example use: end_date must be after start_date.
	EntityValidators []EntityValidator

	// ── Workflow ───────────────────────────────────────────────────────────────

	// WorkflowTriggers declares Temporal workflow bindings for this entity.
	// Each trigger specifies which mutation operations launch which workflow type.
	// Triggers fire after a successful save (via HookAfterSave) so the entity
	// row is committed before the workflow attempts to read it.
	WorkflowTriggers []WorkflowTrigger

	// ── UI Extensions ─────────────────────────────────────────────────────────

	// Actions declares custom action buttons shown in the SDUI list and detail views.
	// Each action maps to a POST /api/{entity}/{id}/{action-name} route registered
	// automatically at bootstrap time.
	Actions []ActionDef

	// PageBuilders overrides the default AMIS schema generators per view mode.
	// When non-nil, the SDUI handler calls the appropriate builder instead of
	// the default CRUDPage / FormPage generators.
	PageBuilders *PageBuilderSet

	// ── Permissions ───────────────────────────────────────────────────────────

	// Permissions maps role names to the set of operations that role may perform.
	// Used by RequireRole policies and the SDUI layer to filter UI elements.
	// Example: {"finance_manager": OpAll, "auditor": OpRead}
	Permissions map[string]Op
}

// Validation errors returned by Validate.
var (
	ErrMissingName       = errors.New("entity definition: Name is required")
	ErrDuplicateField    = errors.New("entity definition: duplicate field name")
	ErrDuplicateEdge     = errors.New("entity definition: duplicate edge name")
	ErrMissingEdgeTarget = errors.New("entity definition: edge missing TargetEntity")
	ErrInvalidFieldType  = errors.New("entity definition: unknown FieldType")
	ErrInvalidOrgScope   = errors.New("entity definition: invalid OrgScope level")
)

// Validate checks the definition for structural correctness.
// The registry calls this automatically at registration time.
func (d *EntityDefinition) Validate() error {
	if d.Name == "" {
		return ErrMissingName
	}

	// Validate org scope level.
	switch d.effectiveOrgScope() {
	case org.ScopeLevelGlobal, org.ScopeLevelTenant, org.ScopeLevelUnit:
		// valid
	default:
		return fmt.Errorf("%w: %q on entity %q", ErrInvalidOrgScope, d.OrgScope, d.Name)
	}

	// Validate fields — uniqueness and known types.
	seen := make(map[string]struct{}, len(d.Fields))
	for _, f := range d.Fields {
		if _, dup := seen[f.Name]; dup {
			return fmt.Errorf("%w: %q in entity %q", ErrDuplicateField, f.Name, d.Name)
		}
		seen[f.Name] = struct{}{}
		if !validFieldType(f.Type) {
			return fmt.Errorf("%w: %q on field %q of entity %q",
				ErrInvalidFieldType, f.Type, f.Name, d.Name)
		}
	}

	// Validate edges — uniqueness and non-empty target.
	edgeSeen := make(map[string]struct{}, len(d.Edges))
	for _, e := range d.Edges {
		if _, dup := edgeSeen[e.Name]; dup {
			return fmt.Errorf("%w: %q in entity %q", ErrDuplicateEdge, e.Name, d.Name)
		}
		edgeSeen[e.Name] = struct{}{}
		if e.TargetEntity == "" {
			return fmt.Errorf("%w: edge %q in entity %q", ErrMissingEdgeTarget, e.Name, d.Name)
		}
	}

	return nil
}

// TableName returns the resolved PostgreSQL table name.
// Defaults to "{Name}s" when Table is empty.
// For EntityTypeCustom the store always uses custom_entity_records; this method
// returns the logical table name used in SQL builders for system entities.
func (d *EntityDefinition) TableName() string {
	if d.Table != "" {
		return d.Table
	}
	return d.Name + "s"
}

// LabelPluralName returns the plural display name.
// Defaults to Label+"s" when LabelPlural is not set.
func (d *EntityDefinition) LabelPluralName() string {
	if d.LabelPlural != "" {
		return d.LabelPlural
	}
	if d.Label != "" {
		return d.Label + "s"
	}
	return d.Name + "s"
}

// IsCustom reports whether this is a runtime custom entity stored in JSONB.
func (d *EntityDefinition) IsCustom() bool {
	return d.Type.IsCustom()
}

// FieldSet returns a map of field name → struct for fast allowlist lookups.
func (d *EntityDefinition) FieldSet() map[string]struct{} {
	m := make(map[string]struct{}, len(d.Fields))
	for _, f := range d.Fields {
		m[f.Name] = struct{}{}
	}
	return m
}

// FieldByName returns the FieldDef with the given name, or nil.
func (d *EntityDefinition) FieldByName(name string) *FieldDef {
	for _, f := range d.Fields {
		if f.Name == name {
			return f
		}
	}
	return nil
}

// EdgeByName returns the EdgeDef with the given name, or nil.
func (d *EntityDefinition) EdgeByName(name string) *EdgeDef {
	for _, e := range d.Edges {
		if e.Name == name {
			return e
		}
	}
	return nil
}

// IsGlobal reports whether this entity has no tenant or unit scoping.
// Equivalent to OrgScope == ScopeLevelGlobal.
func (d *EntityDefinition) IsGlobal() bool {
	return d.effectiveOrgScope() == org.ScopeLevelGlobal
}

// IsUnitScoped reports whether rows carry an org_unit_id column.
// Tree-based access control applies to unit-scoped entities.
func (d *EntityDefinition) IsUnitScoped() bool {
	return d.effectiveOrgScope() == org.ScopeLevelUnit
}

// IsTenantScoped reports whether rows carry a tenant_id but no org_unit_id.
func (d *EntityDefinition) IsTenantScoped() bool {
	return d.effectiveOrgScope() == org.ScopeLevelTenant
}

// effectiveOrgScope returns the OrgScope with the zero-value defaulted to ScopeLevelTenant.
func (d *EntityDefinition) effectiveOrgScope() org.ScopeLevel {
	if d.OrgScope == "" {
		return org.ScopeLevelTenant
	}
	return d.OrgScope
}

// RequiredRoles returns all role names that have been granted the given Op
// in Permissions. Returns nil when Permissions is empty or no role matches op.
func (d *EntityDefinition) RequiredRoles(op Op) []string {
	if len(d.Permissions) == 0 {
		return nil
	}
	var roles []string
	for role, ops := range d.Permissions {
		if ops.Is(op) {
			roles = append(roles, role)
		}
	}
	return roles
}

// validFieldType reports whether t is a recognised FieldType constant.
func validFieldType(t FieldType) bool {
	switch t {
	case FieldTypeData, FieldTypeSmallText, FieldTypeLongText,
		FieldTypeInt, FieldTypeFloat, FieldTypeCurrency,
		FieldTypeBool, FieldTypeDate, FieldTypeDateTime, FieldTypeTime,
		FieldTypeUUID, FieldTypeSelect, FieldTypeMultiSelect, FieldTypeJSON,
		FieldTypeLink, FieldTypeDynamicLink, FieldTypeTable,
		FieldTypeAttach, FieldTypeAttachImage:
		return true
	}
	return false
}
