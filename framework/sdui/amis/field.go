// Package amis builds AMIS JSON schema fragments from EntityDefinition metadata.
package amis

import (
	"awo.so/framework/def"
)

// FormControl converts a FieldDef into an AMIS form control schema map.
func FormControl(f *def.FieldDef) map[string]any {
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
func ColumnDef(f *def.FieldDef) map[string]any {
	col := map[string]any{
		"name":  f.Name,
		"label": label(f),
	}

	switch f.Type {
	case def.FieldTypeBool:
		col["type"] = "status"
	case def.FieldTypeCurrency:
		col["type"] = "number"
		col["prefix"] = ""
	case def.FieldTypeDate:
		col["type"] = "date"
		col["format"] = "YYYY-MM-DD"
	case def.FieldTypeDateTime:
		col["type"] = "datetime"
		col["format"] = "YYYY-MM-DD HH:mm:ss"
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

func label(f *def.FieldDef) string {
	if f.Label != "" {
		return f.Label
	}
	return f.Name
}

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
		// Stored as string (decimal); rendered as number input.
		ctrl["type"] = "input-number"
		ctrl["precision"] = 2
		ctrl["description"] = appendNote(f.Description,
			"Value serialised as string to preserve decimal precision.")
		applyMinMax(ctrl, f)

	case def.FieldTypeBool:
		ctrl["type"] = "switch"

	case def.FieldTypeDate:
		ctrl["type"] = "input-date"
		ctrl["format"] = "YYYY-MM-DD"

	case def.FieldTypeDateTime:
		ctrl["type"] = "input-datetime"
		ctrl["format"] = "YYYY-MM-DD HH:mm:ss"

	case def.FieldTypeTime:
		ctrl["type"] = "input-time"
		ctrl["format"] = "HH:mm:ss"

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
		ctrl["type"] = "input-text"
		if f.LinkedEntity != "" {
			ctrl["type"] = "select"
			ctrl["source"] = "${API_BASE}/api/" + f.LinkedEntity + "s?limit=100"
			ctrl["valueField"] = "id"
			ctrl["labelField"] = "name"
		}

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

func applyValidation(ctrl map[string]any, f *def.FieldDef) {
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

func applyMinMax(ctrl map[string]any, f *def.FieldDef) {
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
