// Package customfields allows tenants to attach extra fields to any registered
// EntityDefinition at runtime without schema migrations on the core tables.
//
// Storage model: one JSONB column `custom_fields` on every tenant table holds
// all custom field values as a flat key→value map. The package provides:
//   - A CustomField registry per tenant (loaded from DB at startup / cache-refreshed)
//   - A before_save HookFunc that validates incoming custom field values
//   - A helper to merge custom field values into Record.Get responses
package customfields

import (
	"context"
	"fmt"
)

// FieldKind mirrors the subset of definition.FieldType that custom fields support.
// Complex types (table, attach, dynamiclink) are intentionally excluded.
type FieldKind string

const (
	KindText     FieldKind = "text"
	KindNumber   FieldKind = "number"
	KindBool     FieldKind = "bool"
	KindDate     FieldKind = "date"
	KindDateTime FieldKind = "datetime"
	KindSelect   FieldKind = "select"
)

// CustomField is a tenant-defined field attached to a named entity.
type CustomField struct {
	// ID is the stable DB identifier for this custom field definition.
	ID string `json:"id"`

	// TenantID scopes the field to a specific tenant.
	TenantID string `json:"tenant_id"`

	// Entity is the EntityDefinition.Name this field is attached to.
	Entity string `json:"entity"`

	// Name is the snake_case key used in custom_fields JSONB and API payloads.
	// Must be unique within (tenant_id, entity).
	Name string `json:"name"`

	// Label is the human-readable name shown in SDUI.
	Label string `json:"label"`

	// Kind determines validation and AMIS control type.
	Kind FieldKind `json:"kind"`

	// Required causes validation to reject blank values.
	Required bool `json:"required"`

	// Options is the allowed value set for KindSelect fields.
	Options []string `json:"options,omitempty"`

	// Position controls sort order in generated forms (ascending).
	Position int `json:"position"`
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
// Returns nil if valid.
func (cf *CustomField) Validate(value any) error {
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

// Store is the interface the host application must implement to persist and
// retrieve CustomField definitions per tenant.
type Store interface {
	// ListForEntity returns all CustomField definitions for (tenantID, entity).
	ListForEntity(ctx context.Context, tenantID, entity string) ([]CustomField, error)

	// Save upserts a CustomField definition.
	Save(ctx context.Context, cf *CustomField) error

	// Delete removes a CustomField definition by ID.
	Delete(ctx context.Context, id string) error
}
