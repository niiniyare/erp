package schema

import (
	"context"
	"fmt"
	"time"
)

// FieldOption for select/radio fields
type FieldOption struct {
	Label    string
	Value    string
	Selected bool
	Disabled bool
	Group    string
}

// FieldConditions for conditional field display
type FieldConditions struct {
	Show []Condition
	Hide []Condition
}

// Condition represents a single condition
type Condition struct {
	Field    string
	Operator string
	Value    any
}

// ConfirmDialog for action confirmation
type ConfirmDialog struct {
	Title   string
	Message string
	Confirm string
	Cancel  string
}

// SectionConfig for section customization
type SectionConfig struct {
	Classes  string
	Required bool
}

// FormConfig for form-level configuration
type FormConfig struct {
	Action   string
	Method   string
	Encoding string
	Classes  string
	Required bool
}

// Theme for form styling
type Theme struct {
	PrimaryColor string
	Spacing      string
	BorderRadius string
}

// CSRFConfig for CSRF protection
type CSRFConfig struct {
	Enabled    bool
	FieldName  string
	HeaderName string
}

// RateLimitConfig for rate limiting
type RateLimitConfig struct {
	Enabled     bool
	MaxRequests int
	Window      time.Duration
}

// EncryptionConfig for field encryption
type EncryptionConfig struct {
	Enabled   bool
	Fields    []string
	Algorithm string
}

// HTMXConfig for HTMX configuration
type HTMXConfig struct {
	Enabled  bool
	Post     string
	Get      string
	Target   string
	Swap     string
	Validate bool
	Headers  map[string]string
}

// AlpineConfig for Alpine.js configuration
type AlpineConfig struct {
	Enabled bool
	XData   string
	XInit   string
}

// Register adds a renderer for a field type
func (r *RendererRegistry) Register(fieldType FieldType, renderer FieldRenderer) {
	r.renderers[fieldType] = renderer
	r.assets[fieldType] = renderer.GetRequiredAssets()
}

// GetRenderer retrieves a renderer for a field type
func (r *RendererRegistry) GetRenderer(fieldType FieldType) FieldRenderer {
	return r.renderers[fieldType]
}

// Validate validates a field
func (f *Field) Validate(ctx context.Context) error {
	if f.Name == "" {
		return ErrValidation{Field: "Name", Message: "field name is required"}
	}

	if f.Type == "" {
		return ErrValidation{Field: "Type", Message: "field type is required"}
	}

	// Validate validation rules
	if f.Validation != nil {
		if f.Validation.Min != nil && f.Validation.Max != nil {
			if *f.Validation.Min > *f.Validation.Max {
				return ErrValidation{Field: "Validation", Message: "min cannot be greater than max"}
			}
		}

		if f.Validation.MinLength != nil && f.Validation.MaxLength != nil {
			if *f.Validation.MinLength > *f.Validation.MaxLength {
				return ErrValidation{Field: "Validation", Message: "minLength cannot be greater than maxLength"}
			}
		}
	}

	return nil
}

// Validate validates an action
func (a *Action) Validate(ctx context.Context) error {
	if a.ID == "" {
		return ErrValidation{Field: "ID", Message: "action ID is required"}
	}

	if a.Text == "" {
		return ErrValidation{Field: "Text", Message: "action text is required"}
	}

	if a.Type == "" {
		a.Type = ActionButton
	}

	return nil
}

// Error types
type ErrValidation struct {
	Field   string
	Message string
}

func (e ErrValidation) Error() string {
	return fmt.Sprintf("validation error in %s: %s", e.Field, e.Message)
}
