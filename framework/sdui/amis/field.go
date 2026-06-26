// Package amis builds AMIS JSON schema fragments from EntityDefinition metadata.
package amis

import (
	"awo.so/framework/definition"
)

// FormControl converts a FieldDef into an AMIS form control schema map.
func FormControl(f *definition.FieldDef) map[string]any {
	ctrl := map[string]any{
		"name":  f.Name,
		"label": label(f),
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
func ColumnDef(f *definition.FieldDef) map[string]any {
	col := map[string]any{
		"name":  f.Name,
		"label": label(f),
	}

	switch f.Type {
	case definition.FieldTypeBool:
		col["type"] = "status"
	case definition.FieldTypeCurrency:
		col["type"] = "number"
		col["prefix"] = ""
	case definition.FieldTypeDate:
		col["type"] = "date"
		col["format"] = "YYYY-MM-DD"
	case definition.FieldTypeDateTime:
		col["type"] = "datetime"
		col["format"] = "YYYY-MM-DD HH:mm:ss"
	case definition.FieldTypeAttachImage:
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

func label(f *definition.FieldDef) string {
	if f.Label != "" {
		return f.Label
	}
	return f.Name
}

func applyType(ctrl map[string]any, f *definition.FieldDef) {
	switch f.Type {
	case definition.FieldTypeSmallText, definition.FieldTypeData:
		ctrl["type"] = "input-text"
		if f.MaxLength > 0 {
			ctrl["maxLength"] = f.MaxLength
		}

	case definition.FieldTypeLongText:
		ctrl["type"] = "textarea"
		if f.MaxLength > 0 {
			ctrl["maxLength"] = f.MaxLength
		}

	case definition.FieldTypeInt:
		ctrl["type"] = "input-number"
		ctrl["precision"] = 0
		applyMinMax(ctrl, f)

	case definition.FieldTypeFloat:
		ctrl["type"] = "input-number"
		applyMinMax(ctrl, f)

	case definition.FieldTypeCurrency:
		// Stored as string (decimal); rendered as number input.
		ctrl["type"] = "input-number"
		ctrl["precision"] = 2
		ctrl["description"] = appendNote(f.Description,
			"Value serialised as string to preserve decimal precision.")
		applyMinMax(ctrl, f)

	case definition.FieldTypeBool:
		ctrl["type"] = "switch"

	case definition.FieldTypeDate:
		ctrl["type"] = "input-date"
		ctrl["format"] = "YYYY-MM-DD"

	case definition.FieldTypeDateTime:
		ctrl["type"] = "input-datetime"
		ctrl["format"] = "YYYY-MM-DD HH:mm:ss"

	case definition.FieldTypeTime:
		ctrl["type"] = "input-time"
		ctrl["format"] = "HH:mm:ss"

	case definition.FieldTypeUUID:
		ctrl["type"] = "input-text"
		ctrl["placeholder"] = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"

	case definition.FieldTypeSelect:
		ctrl["type"] = "select"
		ctrl["options"] = optionItems(f.Options)
		ctrl["clearable"] = !f.IsRequired

	case definition.FieldTypeMultiSelect:
		ctrl["type"] = "select"
		ctrl["multiple"] = true
		ctrl["options"] = optionItems(f.Options)

	case definition.FieldTypeJSON:
		ctrl["type"] = "json-editor"

	case definition.FieldTypeLink:
		ctrl["type"] = "input-text"
		if f.LinkedEntity != "" {
			ctrl["type"] = "select"
			ctrl["source"] = "${API_BASE}/api/" + f.LinkedEntity + "s?limit=100"
			ctrl["valueField"] = "id"
			ctrl["labelField"] = "name"
		}

	case definition.FieldTypeDynamicLink:
		ctrl["type"] = "select"
		ctrl["source"] = "${DYNAMIC_OPTIONS_API}/" + f.Name

	case definition.FieldTypeTable:
		ctrl["type"] = "input-table"

	case definition.FieldTypeAttach:
		ctrl["type"] = "input-file"
		ctrl["multiple"] = true

	case definition.FieldTypeAttachImage:
		ctrl["type"] = "input-image"
		ctrl["multiple"] = true
		ctrl["accept"] = ".jpg,.jpeg,.png,.webp"

	default:
		ctrl["type"] = "input-text"
	}
}

func applyValidation(ctrl map[string]any, f *definition.FieldDef) {
	rules := map[string]any{}
	msgs := map[string]any{}

	if f.IsRequired {
		rules["isRequired"] = true
		msgs["isRequired"] = f.Label + " is required"
	}
	if f.MaxLength > 0 {
		rules["maxLength"] = f.MaxLength
		msgs["maxLength"] = f.Label + " must be at most " + itoa(f.MaxLength) + " characters"
	}
	if f.MinVal != nil {
		rules["minimum"] = f.MinVal.String()
	}
	if f.MaxVal != nil {
		rules["maximum"] = f.MaxVal.String()
	}

	if len(rules) > 0 {
		ctrl["validations"] = rules
		ctrl["validationErrors"] = msgs
	}
}

func applyMinMax(ctrl map[string]any, f *definition.FieldDef) {
	if f.MinVal != nil {
		ctrl["min"] = f.MinVal.String()
	}
	if f.MaxVal != nil {
		ctrl["max"] = f.MaxVal.String()
	}
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

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := make([]byte, 0, 10)
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	return string(buf)
}
