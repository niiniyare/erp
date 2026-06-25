package definition

import (
	"errors"
	"fmt"

	"awo.so/framework/org"
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
// after calling definition.Register produces undefined behaviour because the
// registry stores a pointer and the API/persistence layers cache field lists.
type EntityDefinition struct {
	// ── Identity ──────────────────────────────────────────────────────────────

	// Name is the canonical machine identifier used in URLs, DB tables, and logs.
	// Convention: singular, snake_case. Must be unique across the registry.
	// Example: "sales_order", "finance_account"
	Name string

	// Label is the human-readable display name shown in SDUI navigation and forms.
	// Example: "Sales Order", "Chart of Accounts"
	Label string

	// Description is shown in generated API documentation and SDUI tooltips.
	Description string

	// Module is the logical grouping for SDUI sidebar navigation.
	// Example: "Finance", "HR", "CRM", "Platform"
	Module string

	// ── Storage ───────────────────────────────────────────────────────────────

	// Table is the PostgreSQL table name.
	// Defaults to "{Name}s" when empty (e.g. "finance_account" → "finance_accounts").
	Table string

	// OrgScope determines the level of the organisational hierarchy that scopes
	// this entity's data rows. The persistence layer adds the correct WHERE
	// clauses and the migration generator creates matching RLS policies.
	//
	//   ScopeLevelGlobal   — no tenant_id column; shared across all tenants.
	//                         Example: currencies, countries, languages.
	//
	//   ScopeLevelTenant   — rows have tenant_id; visible across all companies.
	//                         Example: users, roles, feature flags.
	//
	//   ScopeLevelCompany  — rows have tenant_id + company_id.
	//                         Example: GL accounts, employees, fiscal years.
	//
	//   ScopeLevelDivision — rows have tenant_id + company_id + division_id.
	//                         Example: sales targets, divisional budgets.
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
	// See definition.PolicyFunc for the full contract.
	Policies []PolicyDef

	// SoftDelete enables soft-delete via a `deleted_at timestamptz` column.
	// When true: Delete sets deleted_at rather than removing the row; FindByID,
	// List, and Exists automatically exclude deleted records.
	SoftDelete bool

	// Audited enables automatic audit log entries on every Create/Update/Delete.
	// The audit log records actor, tenant/company/division, timestamp, and a
	// JSON diff of changed fields.
	Audited bool
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
	case org.ScopeLevelGlobal, org.ScopeLevelTenant,
		org.ScopeLevelCompany, org.ScopeLevelDivision:
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
func (d *EntityDefinition) TableName() string {
	if d.Table != "" {
		return d.Table
	}
	return d.Name + "s"
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

// IsGlobal reports whether this entity has no tenant scoping.
// Equivalent to OrgScope == ScopeLevelGlobal.
func (d *EntityDefinition) IsGlobal() bool {
	return d.effectiveOrgScope() == org.ScopeLevelGlobal
}

// IsCompanyScoped reports whether rows carry a company_id column.
func (d *EntityDefinition) IsCompanyScoped() bool {
	s := d.effectiveOrgScope()
	return s == org.ScopeLevelCompany || s == org.ScopeLevelDivision
}

// IsDivisionScoped reports whether rows carry a division_id column.
func (d *EntityDefinition) IsDivisionScoped() bool {
	return d.effectiveOrgScope() == org.ScopeLevelDivision
}

// effectiveOrgScope returns the OrgScope with the zero-value defaulted to Tenant.
func (d *EntityDefinition) effectiveOrgScope() org.ScopeLevel {
	if d.OrgScope == "" {
		return org.ScopeLevelTenant
	}
	return d.OrgScope
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
