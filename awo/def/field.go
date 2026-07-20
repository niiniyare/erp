package def

import "context"

// FieldDef describes a single field on an entity. The compiler translates a
// FieldDef into a SQL column definition, an index (if applicable), an amis
// form widget, a CapabilityGrant scope key, and a Filter DSL predicate builder.
type FieldDef struct {
	// Name is the stable snake_case identifier for this field. It is used in:
	//   - SQL column name
	//   - API JSON key
	//   - Filter DSL predicates
	//   - Redis cache keys
	//   - CapabilityGrant entity identifiers
	// Never rename a field after data has been persisted.
	Name string

	// Type is the semantic field type. See FieldType constants.
	Type FieldType

	// Label is the human-readable display name shown in the SDUI.
	Label string

	// Description is shown as help text in the SDUI form.
	Description string

	// Default is the zero-argument function that returns the field's default
	// value when a Create operation omits the field. Return nil to signal no
	// default. The function is called once per record, not once globally.
	Default func() any

	// Required causes the compiler to add a NOT NULL constraint (SQL) and a
	// validation error when the field is absent on Create.
	Required bool

	// Unique causes the compiler to add a UNIQUE constraint.
	Unique bool

	// Immutable marks the field as set-once: updates that include this field
	// produce a BusinessError "immutable_field".
	Immutable bool

	// Sensitive excludes this field from structured logs and standard API
	// responses. It is included only in responses to callers with explicit
	// "read_sensitive" permission.
	Sensitive bool

	// Searchable causes the compiler to create a GIN trigram index on this
	// field. Only effective for FieldTypeData and FieldTypeSmallText.
	Searchable bool

	// MaxLen is the maximum string length. Only effective for FieldTypeData
	// (default 255) and FieldTypeSmallText (default 1024).
	MaxLen int

	// Min and Max constrain numeric field values. Only effective for
	// FieldTypeInt, FieldTypeFloat, and FieldTypeCurrency.
	Min *float64
	Max *float64

	// Options declares the allowed values for FieldTypeSelect and
	// FieldTypeMultiSelect. The compiler generates a SQL CHECK constraint.
	Options []string

	// LinkTarget is the entity name (e.g. "finance_customer") that this
	// FieldTypeLink or FieldTypeLinkList references. The compiler generates a
	// FK constraint.
	LinkTarget string

	// Series is the naming series pattern for FieldTypeNamingSeries fields.
	// Format: "INV-{YYYY}-{SEQ:5}" where:
	//   {YYYY}  → 4-digit year
	//   {MM}    → 2-digit month
	//   {DD}    → 2-digit day
	//   {SEQ:N} → zero-padded sequential integer reset per the ResetPeriod
	Series string

	// TenantOverridable allows tenants to customise the naming series prefix
	// through the Settings module. Only effective for FieldTypeNamingSeries.
	TenantOverridable bool

	// Validators are custom field-level validation functions. They run during
	// the VALIDATE stage of the hook pipeline, before before_save.
	Validators []FieldValidator

	// Hidden removes this field from SDUI list and form views while keeping it
	// accessible via the API. Use for internal bookkeeping fields.
	Hidden bool

	// ReadOnly renders the field as read-only in SDUI edit views.
	ReadOnly bool
}

// FieldValidator is a custom validation function for a single field value.
// It receives the field value (typed per FieldType conventions) and returns
// an error message string. Return "" to indicate the value is valid.
//
// FieldValidators run before any database operation, inside the VALIDATE
// stage of the hook pipeline.
type FieldValidator func(ctx context.Context, value any) string
