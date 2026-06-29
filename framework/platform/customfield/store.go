package customfield

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"awo.so/framework/definition"
)

// FieldKind mirrors the subset of definition.FieldType that custom fields support.
type FieldKind string

const (
	KindText     FieldKind = "text"
	KindNumber   FieldKind = "number"
	KindBool     FieldKind = "bool"
	KindDate     FieldKind = "date"
	KindDateTime FieldKind = "datetime"
	KindSelect   FieldKind = "select"
)

// Field is a tenant-defined field attached to a named entity.
type Field struct {
	ID       string    `json:"id"`
	TenantID string    `json:"tenant_id"`
	Entity   string    `json:"entity"`
	Name     string    `json:"name"`
	Label    string    `json:"label"`
	Kind     FieldKind `json:"kind"`
	Required bool      `json:"required"`
	Options  []string  `json:"options,omitempty"`
	Position int       `json:"position"`
}

// ValidationError is returned when a custom field value fails validation.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("custom field %q: %s", e.Field, e.Message)
}

// Validate checks value against the field's constraints.
func (cf *Field) Validate(value any) error {
	if value == nil || value == "" {
		if cf.Required {
			return &ValidationError{Field: cf.Name, Message: "required"}
		}
		return nil
	}

	switch cf.Kind {
	case KindSelect:
		s, ok := value.(string)
		if !ok {
			return &ValidationError{Field: cf.Name, Message: "must be a string"}
		}
		for _, opt := range cf.Options {
			if opt == s {
				return nil
			}
		}
		return &ValidationError{Field: cf.Name, Message: fmt.Sprintf("must be one of %v", cf.Options)}

	case KindBool:
		if _, ok := value.(bool); !ok {
			return &ValidationError{Field: cf.Name, Message: "must be a boolean"}
		}

	case KindNumber:
		switch value.(type) {
		case float64, float32, int, int64, int32:
			// ok
		default:
			return &ValidationError{Field: cf.Name, Message: "must be a number"}
		}
	}

	return nil
}

// Store is the persistence interface for Field definitions.
type Store interface {
	ListForEntity(ctx context.Context, tenantID, entity string) ([]Field, error)
	Save(ctx context.Context, cf *Field) error
	Delete(ctx context.Context, id string) error
}

// ValidationHook returns a HookFunc that validates the `custom_fields` map
// against the tenant's registered Field definitions.
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
			return fmt.Errorf("customfield: load definitions: %w", err)
		}

		byName := make(map[string]*Field, len(defs))
		for i := range defs {
			byName[defs[i].Name] = &defs[i]
		}

		for key, val := range values {
			cf, known := byName[key]
			if !known {
				return fmt.Errorf("customfield: unknown field %q for entity %q", key, m.After.EntityName())
			}
			if err := cf.Validate(val); err != nil {
				return err
			}
		}

		for _, cf := range defs {
			if cf.Required {
				if _, present := values[cf.Name]; !present {
					if m.Op == definition.OpCreate {
						return &ValidationError{Field: cf.Name, Message: "required"}
					}
				}
			}
		}

		return nil
	}
}

// RegisterRoutes mounts CRUD routes for Field management under prefix.
func RegisterRoutes(router fiber.Router, prefix string, store Store, viewerFn func(*fiber.Ctx) (string, error)) {
	g := router.Group(prefix)

	g.Get("/:entity", func(c *fiber.Ctx) error {
		tenantID, err := viewerFn(c)
		if err != nil {
			return fiber.ErrUnauthorized
		}
		entity := c.Params("entity")
		if definition.Lookup(entity) == nil {
			return fiber.NewError(fiber.StatusNotFound, "unknown entity: "+entity)
		}
		fields, err := store.ListForEntity(c.Context(), tenantID, entity)
		if err != nil {
			return err
		}
		return c.JSON(fields)
	})

	g.Post("/:entity", func(c *fiber.Ctx) error {
		tenantID, err := viewerFn(c)
		if err != nil {
			return fiber.ErrUnauthorized
		}
		entity := c.Params("entity")
		if definition.Lookup(entity) == nil {
			return fiber.NewError(fiber.StatusNotFound, "unknown entity: "+entity)
		}

		var cf Field
		if err := c.BodyParser(&cf); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid body")
		}
		cf.TenantID = tenantID
		cf.Entity = entity
		if cf.ID == "" {
			cf.ID = uuid.New().String()
		}

		if err := store.Save(c.Context(), &cf); err != nil {
			return err
		}
		return c.Status(fiber.StatusCreated).JSON(cf)
	})

	g.Put("/:entity/:id", func(c *fiber.Ctx) error {
		tenantID, err := viewerFn(c)
		if err != nil {
			return fiber.ErrUnauthorized
		}

		var cf Field
		if err := c.BodyParser(&cf); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid body")
		}
		cf.ID = c.Params("id")
		cf.TenantID = tenantID
		cf.Entity = c.Params("entity")

		if err := store.Save(c.Context(), &cf); err != nil {
			return err
		}
		return c.JSON(cf)
	})

	g.Delete("/:entity/:id", func(c *fiber.Ctx) error {
		_, err := viewerFn(c)
		if err != nil {
			return fiber.ErrUnauthorized
		}
		if err := store.Delete(c.Context(), c.Params("id")); err != nil {
			return err
		}
		return c.SendStatus(fiber.StatusNoContent)
	})
}

// AMISControl returns an AMIS form control schema map for a Field.
func AMISControl(cf *Field) map[string]any {
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
// (tenantID, entity) into controls. Safe to call with nil store.
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
	sortByPosition(defs)

	extra := make([]any, 0, len(defs)+1)
	extra = append(extra, map[string]any{"type": "divider", "title": "Custom Fields"})
	for _, cf := range defs {
		extra = append(extra, AMISControl(&cf))
	}
	return append(controls, extra...), nil
}

func sortByPosition(defs []Field) {
	for i := 1; i < len(defs); i++ {
		for j := i; j > 0 && defs[j].Position < defs[j-1].Position; j-- {
			defs[j], defs[j-1] = defs[j-1], defs[j]
		}
	}
}
