package customfields

import (
	"context"
	"fmt"

	"awo.so/framework/definition"
)

// ValidationHook returns a definition.HookFunc that validates the
// `custom_fields` map in a mutation's After record against the tenant's
// registered CustomField definitions.
//
// Attach as a BeforeHook so invalid values are rejected before DB write:
//
//	definition.BeforeHook(definition.OpWrite, customfields.ValidationHook(store))
func ValidationHook(store Store) definition.HookFunc {
	return func(ctx context.Context, m *definition.Mutation) error {
		if m.After == nil {
			return nil
		}

		raw := m.After.Get("custom_fields")
		if raw == nil {
			return nil
		}

		values, ok := raw.(map[string]any)
		if !ok {
			return fmt.Errorf("custom_fields must be a JSON object")
		}

		defs, err := store.ListForEntity(ctx, m.TenantID, m.After.EntityName())
		if err != nil {
			return fmt.Errorf("custom_fields: load definitions: %w", err)
		}

		// Index by name for O(1) lookup.
		byName := make(map[string]*CustomField, len(defs))
		for i := range defs {
			byName[defs[i].Name] = &defs[i]
		}

		// Validate each provided value.
		for key, val := range values {
			cf, known := byName[key]
			if !known {
				return fmt.Errorf("custom_fields: unknown field %q for entity %q", key, m.After.EntityName())
			}
			if err := cf.Validate(val); err != nil {
				return err
			}
		}

		// Check required fields that were omitted entirely.
		for _, cf := range defs {
			if cf.Required {
				if _, present := values[cf.Name]; !present {
					// On update, missing key means "don't change" — only enforce on create.
					if m.Op == definition.OpCreate {
						return &ValidationError{Field: cf.Name, Message: "required"}
					}
				}
			}
		}

		return nil
	}
}
