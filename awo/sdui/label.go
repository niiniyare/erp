package sdui

import (
	"strings"

	"awo.so/awo/def"
)

// fieldLabel returns the display label for a field: FieldDef.Label if set,
// otherwise a title-cased version of the field name.
func fieldLabel(f def.FieldDef) string {
	if f.Label != "" {
		return f.Label
	}
	return toTitle(f.Name)
}

// toTitle converts snake_case to Title Case.
// "invoice_date" → "Invoice Date".
func toTitle(s string) string {
	parts := strings.Split(s, "_")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, " ")
}
