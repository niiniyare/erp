package def

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// FieldType is the canonical type of a field. It determines the PostgreSQL
// column type, the Go value type in Record, the JSON wire format, and which
// AMIS widget the SDUI layer renders.
type FieldType string

const (
	// Scalar types
	FieldTypeData      FieldType = "data"      // varchar(n), default MaxLen 140
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
	FieldTypeSelect      FieldType = "select"      // text, validated against Options
	FieldTypeMultiSelect FieldType = "multiselect" // text[], validated against Options
	FieldTypeJSON        FieldType = "json"        // jsonb

	// Relational types
	FieldTypeLink        FieldType = "link"        // uuid FK to another entity
	FieldTypeDynamicLink FieldType = "dynamiclink" // polymorphic: {name}_type + {name}_id
	FieldTypeTable       FieldType = "table"       // child entity inline (one-to-many)

	// File types
	FieldTypeAttach      FieldType = "attach"      // text (file path or object key)
	FieldTypeAttachImage FieldType = "attachimage" // text + metadata jsonb
)

// EntityValidator is a cross-field validator run after all per-field validators pass.
// It receives the full record and returns zero or more FieldErrors.
// Use for invariants that span multiple fields (e.g. end_date > start_date).
type EntityValidator func(record Record) []*FieldError

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
	IsRequired     bool
	IsUnique       bool
	IsImmutable    bool // set on create, rejected on update
	IsSearchable   bool // creates GIN pg_trgm index (Data fields only)
	IsSensitive    bool // excluded from logs AND all API responses (FindByID, List)
	IsTranslatable bool // value stored in default locale; companion JSONB holds per-locale overrides

	// Type-specific constraints
	MaxLength int              // Data, SmallText
	MinVal    *decimal.Decimal // Int, Float, Currency
	MaxVal    *decimal.Decimal // Int, Float, Currency
	Options   []string         // Select, MultiSelect

	// Defaults
	DefaultVal any         // static default value
	DefaultFn  DefaultFunc // function default (takes precedence over DefaultVal)

	// Relational
	LinkedEntity string // Link, Table, DynamicLink
	ForeignKey   string // explicit FK column name (defaults to {Name}_id)

	// Validators
	Validators      []FieldValidator
	AsyncValidators []AsyncFieldValidator

	// Metadata for SDUI and privacy
	Hidden   bool   // exclude from SDUI list view columns; API still returns the field
	ReadOnly bool   // rendered read-only in forms
	Section  string // form section grouping
	Width    int    // AMIS column width hint

	StorageKeyVal string // Custom storage key / database type mapping (e.g. "currency", "date")
}

// Fluent helper constructors matching the documentation's simplified builder API.

// String creates a standard text field mapping to FieldTypeData (varchar).
func String(name string) *FieldDef {
	return Field(name).OfType(FieldTypeData)
}

// Text creates an unbounded prose field mapping to FieldTypeLongText (text).
func Text(name string) *FieldDef {
	return Field(name).OfType(FieldTypeLongText)
}

// Int creates a signed 64-bit integer field mapping to FieldTypeInt (bigint).
func Int(name string) *FieldDef {
	return Field(name).OfType(FieldTypeInt)
}

// Float creates a 64-bit IEEE 754 float field mapping to FieldTypeFloat.
func Float(name string) *FieldDef {
	return Field(name).OfType(FieldTypeFloat)
}

// Bool creates a boolean field mapping to FieldTypeBool.
func Bool(name string) *FieldDef {
	return Field(name).OfType(FieldTypeBool)
}

// Time creates a datetime/time field mapping to FieldTypeDateTime.
func Time(name string) *FieldDef {
	return Field(name).OfType(FieldTypeDateTime)
}

// UUID creates a UUID field mapping to FieldTypeUUID.
func UUID(name string, dummy ...any) *FieldDef {
	return Field(name).OfType(FieldTypeUUID)
}

// Enum creates a single-select choice field mapping to FieldTypeSelect.
func Enum(name string) *FieldDef {
	return Field(name).OfType(FieldTypeSelect)
}

// Strings creates an array-of-strings field mapping to FieldTypeMultiSelect.
func Strings(name string) *FieldDef {
	return Field(name).OfType(FieldTypeMultiSelect)
}

// JSON creates a structured JSONB document field mapping to FieldTypeJSON.
func JSON(name string, dummy ...any) *FieldDef {
	return Field(name).OfType(FieldTypeJSON)
}

// Other creates a custom-typed or schema-driven field, with type detection.
func Other(name string, schema any) *FieldDef {
	f := Field(name)
	typeName := fmt.Sprintf("%T", schema)
	if strings.Contains(strings.ToLower(typeName), "currency") {
		f.Type = FieldTypeCurrency
	} else if strings.Contains(strings.ToLower(typeName), "time") {
		f.Type = FieldTypeTime
	} else {
		f.Type = FieldTypeData
	}
	return f
}

// Option chaining helpers — return *FieldDef for fluent declaration.

func Field(name string) *FieldDef {
	return &FieldDef{Name: name}
}

func (f *FieldDef) OfType(t FieldType) *FieldDef           { f.Type = t; return f }
func (f *FieldDef) WithLabel(l string) *FieldDef           { f.Label = l; return f }
func (f *FieldDef) RequiredField() *FieldDef               { f.IsRequired = true; return f }
func (f *FieldDef) UniqueField() *FieldDef                 { f.IsUnique = true; return f }
func (f *FieldDef) ImmutableField() *FieldDef              { f.IsImmutable = true; return f }
func (f *FieldDef) SearchableField() *FieldDef             { f.IsSearchable = true; return f }
func (f *FieldDef) SensitiveField() *FieldDef              { f.IsSensitive = true; return f }
func (f *FieldDef) WithMaxLen(n int) *FieldDef             { f.MaxLength = n; return f }
func (f *FieldDef) WithOptions(opts ...string) *FieldDef   { f.Options = opts; return f }
func (f *FieldDef) WithDefault(v any) *FieldDef            { f.DefaultVal = v; return f }
func (f *FieldDef) WithDefaultFn(fn DefaultFunc) *FieldDef { f.DefaultFn = fn; return f }
func (f *FieldDef) LinksTo(entity string) *FieldDef        { f.LinkedEntity = entity; return f }
func (f *FieldDef) HiddenInList() *FieldDef                { f.Hidden = true; return f }
func (f *FieldDef) ReadOnlyField() *FieldDef               { f.ReadOnly = true; return f }
func (f *FieldDef) InSection(s string) *FieldDef           { f.Section = s; return f }
func (f *FieldDef) WithWidth(w int) *FieldDef              { f.Width = w; return f }

func (f *FieldDef) Validate(v FieldValidator) *FieldDef {
	f.Validators = append(f.Validators, v)
	return f
}

func (f *FieldDef) ValidateAsync(v AsyncFieldValidator) *FieldDef {
	f.AsyncValidators = append(f.AsyncValidators, v)
	return f
}

func (f *FieldDef) WithMin(v decimal.Decimal) *FieldDef { f.MinVal = &v; return f }
func (f *FieldDef) WithMax(v decimal.Decimal) *FieldDef { f.MaxVal = &v; return f }

// Documentation-aligned fluent builder API aliases and helpers

// NotEmpty marks the field as required. Alias for RequiredField.
func (f *FieldDef) NotEmpty() *FieldDef {
	return f.RequiredField()
}

// Required marks the field as required. Alias for RequiredField.
func (f *FieldDef) Required() *FieldDef {
	return f.RequiredField()
}

// Unique marks the field as unique. Alias for UniqueField.
func (f *FieldDef) Unique() *FieldDef {
	return f.UniqueField()
}

// Immutable marks the field as immutable. Alias for ImmutableField.
func (f *FieldDef) Immutable() *FieldDef {
	return f.ImmutableField()
}

// Searchable marks the field as searchable. Alias for SearchableField.
func (f *FieldDef) Searchable() *FieldDef {
	return f.SearchableField()
}

// Sensitive marks the field as sensitive. Alias for SensitiveField.
func (f *FieldDef) Sensitive() *FieldDef {
	return f.SensitiveField()
}

// Translatable marks the field as translatable. The default-locale value is stored
// in the standard column; per-locale overrides are stored in a companion
// {field}_translations jsonb column resolved at response time via Accept-Language.
func (f *FieldDef) Translatable() *FieldDef {
	f.IsTranslatable = true
	return f
}

// Optional marks the field as optional (IsRequired = false).
func (f *FieldDef) Optional() *FieldDef {
	f.IsRequired = false
	return f
}

// MaxLen sets the maximum length constraints. Alias for WithMaxLen.
func (f *FieldDef) MaxLen(n int) *FieldDef {
	return f.WithMaxLen(n)
}

// Values sets the available options for select fields. Alias for WithOptions.
func (f *FieldDef) Values(opts ...string) *FieldDef {
	return f.WithOptions(opts...)
}

// Min sets the minimum value constraint, accepting numeric types or decimal.Decimal.
func (f *FieldDef) Min(v any) *FieldDef {
	var d decimal.Decimal
	switch val := v.(type) {
	case int:
		d = decimal.NewFromInt(int64(val))
	case int32:
		d = decimal.NewFromInt(int64(val))
	case int64:
		d = decimal.NewFromInt(val)
	case float32:
		d = decimal.NewFromFloat(float64(val))
	case float64:
		d = decimal.NewFromFloat(val)
	case string:
		var err error
		d, err = decimal.NewFromString(val)
		if err != nil {
			panic(fmt.Sprintf("invalid Min string value: %v", err))
		}
	case decimal.Decimal:
		d = val
	default:
		panic(fmt.Sprintf("unsupported Min type: %T", v))
	}
	f.MinVal = &d
	return f
}

// Max sets the maximum value constraint, accepting numeric types or decimal.Decimal.
func (f *FieldDef) Max(v any) *FieldDef {
	var d decimal.Decimal
	switch val := v.(type) {
	case int:
		d = decimal.NewFromInt(int64(val))
	case int32:
		d = decimal.NewFromInt(int64(val))
	case int64:
		d = decimal.NewFromInt(val)
	case float32:
		d = decimal.NewFromFloat(float64(val))
	case float64:
		d = decimal.NewFromFloat(val)
	case string:
		var err error
		d, err = decimal.NewFromString(val)
		if err != nil {
			panic(fmt.Sprintf("invalid Max string value: %v", err))
		}
	case decimal.Decimal:
		d = val
	default:
		panic(fmt.Sprintf("unsupported Max type: %T", v))
	}
	f.MaxVal = &d
	return f
}

// Default sets either a static default value or a lazy default function based on type.
func (f *FieldDef) Default(v any) *FieldDef {
	if fn, ok := v.(DefaultFunc); ok {
		f.DefaultFn = fn
		return f
	}
	if fn, ok := v.(func() any); ok {
		f.DefaultFn = fn
		return f
	}
	if fn, ok := v.(func() time.Time); ok {
		f.DefaultFn = func() any { return fn() }
		return f
	}
	if fn, ok := v.(func() string); ok {
		f.DefaultFn = func() any { return fn() }
		return f
	}
	f.DefaultVal = v
	return f
}

// StorageKey configures custom database types/column overrides (e.g. "currency", "date").
func (f *FieldDef) StorageKey(key string) *FieldDef {
	f.StorageKeyVal = key
	if key == "currency" {
		f.Type = FieldTypeCurrency
	} else if key == "date" {
		f.Type = FieldTypeDate
	}
	return f
}
