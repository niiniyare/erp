package definition

import (
	"errors"
	"fmt"
)

// EntityDefinition is the authoritative meta-model for one entity type.
// It drives dynamic SQL generation, REST API registration, SDUI schema
// building, privacy enforcement, and lifecycle hooks.
//
// EntityDefinitions are immutable after registration. Mutating a definition
// after calling Register produces undefined behaviour.
type EntityDefinition struct {
	// Name is the canonical machine name used in URLs, DB tables, and logs.
	// Convention: singular, snake_case (e.g. "sales_order").
	Name string

	// Label is the human-readable display name (e.g. "Sales Order").
	Label string

	// Description is shown in generated API docs and SDUI tooltips.
	Description string

	// Table is the PostgreSQL table name. Defaults to "{Name}s" if empty.
	Table string

	// Global marks the entity as tenant-agnostic (no tenant_id column, no RLS).
	// Examples: timezones, currencies, countries.
	Global bool

	// Fields declares all scalar and relational attributes of the entity.
	Fields []*FieldDef

	// Edges declares relationships to other entities.
	Edges []*EdgeDef

	// Hooks are lifecycle callbacks invoked within write transactions.
	Hooks []HookDef

	// Policies control read/write access per operation.
	// The chain is evaluated in order; fail-closed (deny if no ErrAllow).
	Policies []PolicyDef

	// SoftDelete enables soft-delete via a `deleted_at` timestamptz column.
	// FindByID, List, and Exists automatically filter out deleted records.
	SoftDelete bool

	// Audited enables automatic creation of audit log entries on every write.
	Audited bool

	// Module is the logical grouping (e.g. "finance", "hr") for SDUI navigation.
	Module string
}

// Validation errors returned by EntityDefinition.Validate.
var (
	ErrMissingName      = errors.New("entity definition: Name is required")
	ErrDuplicateField   = errors.New("entity definition: duplicate field name")
	ErrDuplicateEdge    = errors.New("entity definition: duplicate edge name")
	ErrMissingEdgeTarget = errors.New("entity definition: edge missing TargetEntity")
	ErrInvalidFieldType = errors.New("entity definition: unknown FieldType")
)

// Validate checks the definition for structural correctness.
// Called automatically by the registry at registration time.
func (d *EntityDefinition) Validate() error {
	if d.Name == "" {
		return ErrMissingName
	}

	seen := make(map[string]struct{}, len(d.Fields))
	for _, f := range d.Fields {
		if _, dup := seen[f.Name]; dup {
			return fmt.Errorf("%w: %q in entity %q", ErrDuplicateField, f.Name, d.Name)
		}
		seen[f.Name] = struct{}{}
		if !validFieldType(f.Type) {
			return fmt.Errorf("%w: %q on field %q", ErrInvalidFieldType, f.Type, f.Name)
		}
	}

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
