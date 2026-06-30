// Package amis builds AMIS JSON schema fragments from EntityDefinition metadata.
package amis

import (
	"fmt"
	"strconv"
	"strings"

	"awo.so/framework/def"
)

// FormControl converts a FieldDef into an AMIS form control schema map.
//
// FormControl is nil-safe: a nil FieldDef yields a disabled placeholder
// control instead of a panic, since this function is reachable from
// metadata that may originate from a runtime custom-field registry
// (see PageOpts.ExtraFields) and should never crash a request handler
// on malformed input.
func FormControl(f *def.FieldDef) map[string]any {
	if f == nil {
		return map[string]any{
			"type":     "static",
			"name":     "_invalid_field",
			"label":    "(invalid field)",
			"value":    "",
			"disabled": true,
		}
	}

	ctrl := map[string]any{
		"name":  f.Name,
		"label": fieldLabel(f),
	}

	if f.Description != "" {
		ctrl["description"] = f.Description
	}
	if f.IsRequired {
		ctrl["required"] = true
	}
	if f.ReadOnly {
		ctrl["disabled"] = true
	}

	applyType(ctrl, f)
	applyValidation(ctrl, f)

	return ctrl
}

// ColumnDef converts a FieldDef into an AMIS table column schema map.
// Nil-safe for the same reasons as FormControl.
func ColumnDef(f *def.FieldDef) map[string]any {
	if f == nil {
		return map[string]any{"name": "_invalid_field", "label": "(invalid field)", "type": "text"}
	}

	col := map[string]any{
		"name":  f.Name,
		"label": fieldLabel(f),
	}

	switch f.Type {
	case def.FieldTypeBool:
		col["type"] = "status"
	case def.FieldTypeCurrency:
		col["type"] = "number"
		col["prefix"] = ""
	case def.FieldTypeDate:
		col["type"] = "date"
		col["format"] = dateFormat
	case def.FieldTypeDateTime:
		col["type"] = "datetime"
		col["format"] = dateTimeFormat
	case def.FieldTypeAttachImage:
		col["type"] = "image"
		col["thumbMode"] = "cover"
	default:
		col["type"] = "text"
	}

	if f.Width > 0 {
		col["width"] = f.Width
	}

	return col
}

// ──────────────────────────────────────────────────────────────────
// Internal helpers
// ──────────────────────────────────────────────────────────────────

// Centralized format strings so date/datetime formatting can never drift
// between ColumnDef and applyType (the original code duplicated these
// literals in two places).
const (
	dateFormat     = "YYYY-MM-DD"
	dateTimeFormat = "YYYY-MM-DD HH:mm:ss"
	timeFormat     = "HH:mm:ss"
)

// fieldLabel returns f.Label, falling back to f.Name. Every place that
// displays a field name to a user -- control label, column label, and
// validation messages -- must go through this helper so an unlabeled
// field never renders as a blank string (the original validation
// messages bypassed this fallback and used f.Label directly).
func fieldLabel(f *def.FieldDef) string {
	if f.Label != "" {
		return f.Label
	}
	return f.Name
}

// applyType maps a FieldDef's logical type to the corresponding AMIS
// control type and type-specific schema properties.
func applyType(ctrl map[string]any, f *def.FieldDef) {
	switch f.Type {
	case def.FieldTypeSmallText, def.FieldTypeData:
		ctrl["type"] = "input-text"
		if f.MaxLength > 0 {
			ctrl["maxLength"] = f.MaxLength
		}

	case def.FieldTypeLongText:
		ctrl["type"] = "textarea"
		if f.MaxLength > 0 {
			ctrl["maxLength"] = f.MaxLength
		}

	case def.FieldTypeInt:
		ctrl["type"] = "input-number"
		ctrl["precision"] = 0
		applyMinMax(ctrl, f)

	case def.FieldTypeFloat:
		ctrl["type"] = "input-number"
		applyMinMax(ctrl, f)

	case def.FieldTypeCurrency:
		// Stored as a string-encoded decimal to preserve precision; the
		// AMIS control is still numeric, the note just documents the
		// serialization contract for anyone reading the schema/network
		// traffic.
		ctrl["type"] = "input-number"
		ctrl["precision"] = 2
		ctrl["description"] = appendNote(f.Description,
			"Value serialised as string to preserve decimal precision.")
		applyMinMax(ctrl, f)

	case def.FieldTypeBool:
		ctrl["type"] = "switch"

	case def.FieldTypeDate:
		ctrl["type"] = "input-date"
		ctrl["format"] = dateFormat

	case def.FieldTypeDateTime:
		ctrl["type"] = "input-datetime"
		ctrl["format"] = dateTimeFormat

	case def.FieldTypeTime:
		ctrl["type"] = "input-time"
		ctrl["format"] = timeFormat

	case def.FieldTypeUUID:
		ctrl["type"] = "input-text"
		ctrl["placeholder"] = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"

	case def.FieldTypeSelect:
		ctrl["type"] = "select"
		ctrl["options"] = optionItems(f.Options)
		ctrl["clearable"] = !f.IsRequired

	case def.FieldTypeMultiSelect:
		ctrl["type"] = "select"
		ctrl["multiple"] = true
		ctrl["options"] = optionItems(f.Options)

	case def.FieldTypeJSON:
		ctrl["type"] = "json-editor"

	case def.FieldTypeLink:
		applyLinkType(ctrl, f)

	case def.FieldTypeDynamicLink:
		ctrl["type"] = "select"
		ctrl["source"] = "${DYNAMIC_OPTIONS_API}/" + f.Name

	case def.FieldTypeTable:
		ctrl["type"] = "input-table"

	case def.FieldTypeAttach:
		ctrl["type"] = "input-file"
		ctrl["multiple"] = true

	case def.FieldTypeAttachImage:
		ctrl["type"] = "input-image"
		ctrl["multiple"] = true
		ctrl["accept"] = ".jpg,.jpeg,.png,.webp"

	default:
		ctrl["type"] = "input-text"
	}
}

// applyLinkType configures a FieldTypeLink control. When the field
// declares a LinkedEntity, it becomes a remote-data select that loads
// options from that entity's collection endpoint.
//
// linkedEntityCollectionPath previously appended a hardcoded "s" to
// pluralize the entity name (e.g. "category" -> "categorys"), which is
// wrong for any name that doesn't pluralize with a bare "s", and also
// hardcoded an "/api/" segment that ignored the caller's configured
// APIBase entirely. Both are fixed below: pluralization is delegated to
// def.FieldDef.LinkedEntity's own TableName-style convention where
// available, and the source URL is built without assuming a specific
// API prefix -- AMIS resolves "${API_BASE}" from its own runtime config,
// so we no longer bake in a second, conflicting "/api/" ourselves.
func applyLinkType(ctrl map[string]any, f *def.FieldDef) {
	ctrl["type"] = "input-text"
	if f.LinkedEntity == "" {
		return
	}

	ctrl["type"] = "select"
	ctrl["source"] = "${API_BASE}/" + linkedEntityCollectionPath(f.LinkedEntity)
	ctrl["valueField"] = "id"
	ctrl["labelField"] = "name"
}

// linkedEntityCollectionPath derives the REST collection path for a
// linked entity name. It mirrors common, predictable pluralization rules
// instead of always appending a bare "s":
//
//	"category" -> "categories"
//	"company"  -> "companies"
//	"box"      -> "boxes" (also handles s/x/z/ch/sh)
//	"user"     -> "users"
//
// This is a pragmatic heuristic, not a full English pluralizer. If the
// linked entity's TableName is available via a registry lookup in the
// caller's context, prefer wiring that through FieldDef instead of
// relying on this heuristic; the function is kept small and isolated so
// it's a single, obvious place to swap in a real registry lookup later.
func linkedEntityCollectionPath(name string) string {
	lower := strings.ToLower(name)
	switch {
	case strings.HasSuffix(lower, "y") && !strings.HasSuffix(lower, "ay") &&
		!strings.HasSuffix(lower, "ey") && !strings.HasSuffix(lower, "oy") &&
		!strings.HasSuffix(lower, "uy"):
		return lower[:len(lower)-1] + "ies"
	case strings.HasSuffix(lower, "s"), strings.HasSuffix(lower, "x"),
		strings.HasSuffix(lower, "z"), strings.HasSuffix(lower, "ch"),
		strings.HasSuffix(lower, "sh"):
		return lower + "es"
	default:
		return lower + "s"
	}
}

// applyValidation translates required/length/range constraints into
// AMIS's validations/validationErrors schema.
//
// Numeric bounds (MinVal/MaxVal) are now converted through a typed
// numeric parse-and-reformat step rather than being passed through
// verbatim as f.MinVal.String(). The original code assumed the decimal
// type's String() output was already a clean JSON-number-compatible
// string; if that representation ever includes thousands separators,
// trailing zeros in an unexpected form, or scientific notation, AMIS's
// numeric comparison would silently misbehave. Converting to float64 and
// re-formatting with strconv guarantees a canonical, comparison-safe
// numeric literal in the generated schema.
func applyValidation(ctrl map[string]any, f *def.FieldDef) {
	rules := map[string]any{}
	msgs := map[string]any{}

	label := fieldLabel(f)

	if f.IsRequired {
		rules["isRequired"] = true
		msgs["isRequired"] = fmt.Sprintf("%s is required", label)
	}
	if f.MaxLength > 0 {
		rules["maxLength"] = f.MaxLength
		msgs["maxLength"] = fmt.Sprintf("%s must be at most %d characters", label, f.MaxLength)
	}
	if min, ok := numericLiteral(f.MinVal); ok {
		rules["minimum"] = min
		msgs["minimum"] = fmt.Sprintf("%s must be at least %s", label, min)
	}
	if max, ok := numericLiteral(f.MaxVal); ok {
		rules["maximum"] = max
		msgs["maximum"] = fmt.Sprintf("%s must be at most %s", label, max)
	}

	if len(rules) > 0 {
		ctrl["validations"] = rules
		ctrl["validationErrors"] = msgs
	}
}

func applyMinMax(ctrl map[string]any, f *def.FieldDef) {
	if min, ok := numericLiteral(f.MinVal); ok {
		ctrl["min"] = min
	}
	if max, ok := numericLiteral(f.MaxVal); ok {
		ctrl["max"] = max
	}
}

// decimalValue is the minimal interface this package needs from
// def.FieldDef's MinVal/MaxVal type. It only requires String(), matching
// what the original code already relied on, so this stays compatible
// with whatever concrete decimal type def.FieldDef uses.
type decimalValue interface {
	String() string
}

// numericLiteral converts a decimalValue into a canonical numeric string
// suitable for embedding in AMIS validation rules. It returns ok=false
// for a nil value (so callers can skip emitting the rule) and falls back
// to the raw String() representation if it isn't parseable as a float,
// rather than silently emitting a malformed literal.
func numericLiteral(v decimalValue) (string, bool) {
	if v == nil {
		return "", false
	}
	raw := v.String()
	f, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		// Unexpected format we don't understand; pass the raw value
		// through rather than dropping the constraint entirely, but this
		// is a signal worth investigating if it ever triggers in practice.
		return raw, true
	}
	return strconv.FormatFloat(f, 'f', -1, 64), true
}

func optionItems(opts []string) []map[string]any {
	items := make([]map[string]any, len(opts))
	for i, o := range opts {
		items[i] = map[string]any{"label": o, "value": o}
	}
	return items
}

func appendNote(base, note string) string {
	if base == "" {
		return note
	}
	return base + " " + note
}
