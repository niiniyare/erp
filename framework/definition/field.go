package definition

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

// FieldType is the canonical type of a field. It determines the PostgreSQL
// column type, the Go value type in Record, the JSON wire format, and which
// AMIS widget the SDUI layer renders.
type FieldType string

const (
	// Scalar types
	FieldTypeData     FieldType = "data"      // varchar(n), default MaxLen 140
	FieldTypeSmallText FieldType = "smalltext" // varchar(1024), never indexed
	FieldTypeLongText  FieldType = "longtext"  // text, unbounded
	FieldTypeInt       FieldType = "int"       // bigint
	FieldTypeFloat     FieldType = "float"     // double precision — never for money
	FieldTypeCurrency  FieldType = "currency"  // numeric(20,4), serialised as string in JSON
	FieldTypeBool      FieldType = "bool"      // boolean NOT NULL
	FieldTypeDate      FieldType = "date"      // date (no time)
	FieldTypeDateTime  FieldType = "datetime"  // timestamptz, stored UTC
	FieldTypeTime      FieldType = "time"      // time without time zone
	FieldTypeUUID      FieldType = "uuid"      // uuid

	// Structured types
	FieldTypeSelect      FieldType = "select"       // text, validated against Options
	FieldTypeMultiSelect FieldType = "multiselect"  // text[], validated against Options
	FieldTypeJSON        FieldType = "json"         // jsonb

	// Relational types
	FieldTypeLink        FieldType = "link"         // uuid FK to another entity
	FieldTypeDynamicLink FieldType = "dynamiclink"  // polymorphic: {name}_type + {name}_id
	FieldTypeTable       FieldType = "table"        // child entity inline (one-to-many)

	// File types
	FieldTypeAttach      FieldType = "attach"       // text (file path or object key)
	FieldTypeAttachImage FieldType = "attachimage"  // text + metadata jsonb
)

// FieldValidator is a synchronous field-level validator.
// value is already coerced to the field's Go type.
// record provides read-only access to sibling fields for cross-field checks.
type FieldValidator func(value any, record Record) *FieldError

// AsyncFieldValidator is a validator that may query the database.
// Runs after all synchronous validators pass.
type AsyncFieldValidator func(ctx context.Context, store ReadStore, value any, record Record) *FieldError

// FieldError represents a single field-level validation failure.
type FieldError struct {
	Field   string
	Message string
}

func NewFieldError(field, message string) *FieldError {
	return &FieldError{Field: field, Message: message}
}

func (e *FieldError) Error() string {
	return e.Field + ": " + e.Message
}

// DefaultFunc produces a default value for a field at record assembly time.
// Must be fast and pure (no DB calls, no side effects).
type DefaultFunc func() any

// Now is the standard default function for DateTime fields.
var Now DefaultFunc = func() any { return time.Now().UTC() }

// FieldDef is the authoritative description of one attribute across every
// layer of the framework: storage, validation, API wire format, SDUI widget,
// privacy policy behaviour.
type FieldDef struct {
	Name        string
	Type        FieldType
	Label       string // human-readable, used by SDUI
	Description string

	// Constraints
	Required    bool
	Unique      bool
	Immutable   bool // set on create, rejected on update
	Searchable  bool // creates GIN pg_trgm index (Data fields only)
	Sensitive   bool // excluded from logs AND all API responses (FindByID, List)

	// Type-specific constraints
	MaxLen int            // Data, SmallText
	Min    *decimal.Decimal // Int, Float, Currency
	Max    *decimal.Decimal // Int, Float, Currency
	Options []string      // Select, MultiSelect

	// Defaults
	Default     any         // static default value
	DefaultFn   DefaultFunc // function default (takes precedence over Default)

	// Relational
	LinkedEntity string // Link, Table, DynamicLink
	ForeignKey   string // explicit FK column name (defaults to {Name}_id)

	// Validators
	Validators      []FieldValidator
	AsyncValidators []AsyncFieldValidator

	// Metadata for SDUI and privacy
	Hidden      bool   // exclude from SDUI list view columns; API still returns the field
	ReadOnly    bool   // rendered read-only in forms
	Section     string // form section grouping
	Width       int    // AMIS column width hint
}

// Option chaining helpers — return *FieldDef for fluent declaration.

func Field(name string) *FieldDef {
	return &FieldDef{Name: name}
}

func (f *FieldDef) OfType(t FieldType) *FieldDef     { f.Type = t; return f }
func (f *FieldDef) WithLabel(l string) *FieldDef      { f.Label = l; return f }
func (f *FieldDef) RequiredField() *FieldDef           { f.Required = true; return f }
func (f *FieldDef) UniqueField() *FieldDef             { f.Unique = true; return f }
func (f *FieldDef) ImmutableField() *FieldDef          { f.Immutable = true; return f }
func (f *FieldDef) SearchableField() *FieldDef         { f.Searchable = true; return f }
func (f *FieldDef) SensitiveField() *FieldDef          { f.Sensitive = true; return f }
func (f *FieldDef) WithMaxLen(n int) *FieldDef         { f.MaxLen = n; return f }
func (f *FieldDef) WithOptions(opts ...string) *FieldDef { f.Options = opts; return f }
func (f *FieldDef) WithDefault(v any) *FieldDef        { f.Default = v; return f }
func (f *FieldDef) WithDefaultFn(fn DefaultFunc) *FieldDef { f.DefaultFn = fn; return f }
func (f *FieldDef) LinksTo(entity string) *FieldDef    { f.LinkedEntity = entity; return f }
func (f *FieldDef) HiddenInList() *FieldDef            { f.Hidden = true; return f }
func (f *FieldDef) ReadOnlyField() *FieldDef           { f.ReadOnly = true; return f }
func (f *FieldDef) InSection(s string) *FieldDef       { f.Section = s; return f }
func (f *FieldDef) WithWidth(w int) *FieldDef          { f.Width = w; return f }

func (f *FieldDef) Validate(v FieldValidator) *FieldDef {
	f.Validators = append(f.Validators, v)
	return f
}

func (f *FieldDef) ValidateAsync(v AsyncFieldValidator) *FieldDef {
	f.AsyncValidators = append(f.AsyncValidators, v)
	return f
}

func (f *FieldDef) WithMin(v decimal.Decimal) *FieldDef { f.Min = &v; return f }
func (f *FieldDef) WithMax(v decimal.Decimal) *FieldDef { f.Max = &v; return f }
