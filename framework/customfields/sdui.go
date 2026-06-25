package customfields

import (
	"context"
)

// AMISControl returns an AMIS form control schema map for a CustomField.
// The control is nested under the "custom_fields" combo-field, so name is
// prefixed with "custom_fields." for proper AMIS data binding.
func AMISControl(cf *CustomField) map[string]any {
	ctrl := map[string]any{
		"name":  "custom_fields." + cf.Name,
		"label": cf.Label,
	}

	if cf.Required {
		ctrl["required"] = true
	}

	switch cf.Kind {
	case KindText:
		ctrl["type"] = "input-text"
	case KindNumber:
		ctrl["type"] = "input-number"
	case KindBool:
		ctrl["type"] = "switch"
	case KindDate:
		ctrl["type"] = "input-date"
		ctrl["format"] = "YYYY-MM-DD"
	case KindDateTime:
		ctrl["type"] = "input-datetime"
		ctrl["format"] = "YYYY-MM-DD HH:mm:ss"
	case KindSelect:
		ctrl["type"] = "select"
		opts := make([]map[string]any, len(cf.Options))
		for i, o := range cf.Options {
			opts[i] = map[string]any{"label": o, "value": o}
		}
		ctrl["options"] = opts
		ctrl["clearable"] = !cf.Required
	default:
		ctrl["type"] = "input-text"
	}

	return ctrl
}

// InjectControls appends AMIS form controls for all custom fields of
// (tenantID, entity) into the provided controls slice.
// Returns the extended slice. Safe to call with nil store (returns controls unchanged).
func InjectControls(ctx context.Context, store Store, tenantID, entity string, controls []any) ([]any, error) {
	if store == nil {
		return controls, nil
	}

	defs, err := store.ListForEntity(ctx, tenantID, entity)
	if err != nil {
		return controls, err
	}
	if len(defs) == 0 {
		return controls, nil
	}

	// Sort by Position (already ordered by DB query, but guard here).
	sortByPosition(defs)

	extra := make([]any, 0, len(defs)+1)
	// Visual separator
	extra = append(extra, map[string]any{
		"type":  "divider",
		"title": "Custom Fields",
	})
	for _, cf := range defs {
		extra = append(extra, AMISControl(&cf))
	}

	return append(controls, extra...), nil
}

// sortByPosition is a simple insertion sort — custom field counts are small.
func sortByPosition(defs []CustomField) {
	for i := 1; i < len(defs); i++ {
		for j := i; j > 0 && defs[j].Position < defs[j-1].Position; j-- {
			defs[j], defs[j-1] = defs[j-1], defs[j]
		}
	}
}
