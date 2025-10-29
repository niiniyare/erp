package schema

// Package-level helpers for common operations

// Builder provides a fluent interface for building schemas programmatically
type Builder struct {
	schema *Schema
}

// NewBuilder creates a new schema builder
func NewBuilder(id string, schemaType Type, title string) *Builder {
	return &Builder{
		schema: NewSchema(id, schemaType, title),
	}
}

// WithDescription adds a description
func (b *Builder) WithDescription(desc string) *Builder {
	b.schema.Description = desc
	return b
}

// WithVersion sets the version
func (b *Builder) WithVersion(version string) *Builder {
	b.schema.Version = version
	return b
}

// WithCategory sets the category
func (b *Builder) WithCategory(category string) *Builder {
	b.schema.Category = category
	return b
}

// WithModule sets the module
func (b *Builder) WithModule(module string) *Builder {
	b.schema.Module = module
	return b
}

// WithTags adds tags
func (b *Builder) WithTags(tags ...string) *Builder {
	b.schema.Tags = append(b.schema.Tags, tags...)
	return b
}

// WithConfig sets the config
func (b *Builder) WithConfig(config *Config) *Builder {
	b.schema.Config = config
	return b
}

// WithLayout sets the layout
func (b *Builder) WithLayout(layout *Layout) *Builder {
	b.schema.Layout = layout
	return b
}

// WithTenant enables tenant isolation
func (b *Builder) WithTenant(field, isolation string) *Builder {
	b.schema.Tenant = &Tenant{
		Enabled:   true,
		Field:     field,
		Isolation: isolation,
	}
	return b
}

// WithSecurity adds security configuration
func (b *Builder) WithSecurity(security *Security) *Builder {
	b.schema.Security = security
	return b
}

// WithCSRF enables CSRF protection
func (b *Builder) WithCSRF() *Builder {
	if b.schema.Security == nil {
		b.schema.Security = &Security{}
	}
	b.schema.Security.CSRF = &CSRF{
		Enabled:    true,
		FieldName:  "_csrf",
		HeaderName: "X-CSRF-Token",
	}
	return b
}

// WithRateLimit adds rate limiting
func (b *Builder) WithRateLimit(maxRequests int, windowSeconds int64) *Builder {
	if b.schema.Security == nil {
		b.schema.Security = &Security{}
	}
	b.schema.Security.RateLimit = &RateLimit{
		Enabled:     true,
		MaxRequests: maxRequests,
		ByUser:      true,
	}
	return b
}

// WithHTMX configures HTMX
func (b *Builder) WithHTMX(post, target string) *Builder {
	b.schema.HTMX = &HTMX{
		Enabled: true,
		Post:    post,
		Target:  target,
		Swap:    "innerHTML",
	}
	return b
}

// WithAlpine enables Alpine.js with initial data
func (b *Builder) WithAlpine(xData string) *Builder {
	b.schema.Alpine = &Alpine{
		Enabled: true,
		XData:   xData,
	}
	return b
}

// WithI18n enables internationalization
func (b *Builder) WithI18n(defaultLocale string, supported ...string) *Builder {
	b.schema.I18n = &I18n{
		Enabled:          true,
		DefaultLocale:    defaultLocale,
		SupportedLocales: supported,
	}
	return b
}

// AddTextField adds a text field
func (b *Builder) AddTextField(name, label string, required bool) *Builder {
	b.schema.AddField(Field{
		Name:     name,
		Type:     FieldText,
		Label:    label,
		Required: required,
	})
	return b
}

// AddEmailField adds an email field
func (b *Builder) AddEmailField(name, label string, required bool) *Builder {
	b.schema.AddField(Field{
		Name:     name,
		Type:     FieldEmail,
		Label:    label,
		Required: required,
	})
	return b
}

// AddPasswordField adds a password field
func (b *Builder) AddPasswordField(name, label string, required bool) *Builder {
	b.schema.AddField(Field{
		Name:     name,
		Type:     FieldPassword,
		Label:    label,
		Required: required,
	})
	return b
}

// AddNumberField adds a number field
func (b *Builder) AddNumberField(name, label string, required bool, min, max *float64) *Builder {
	field := Field{
		Name:     name,
		Type:     FieldNumber,
		Label:    label,
		Required: required,
	}
	if min != nil || max != nil {
		field.Validation = &FieldValidation{
			Min: min,
			Max: max,
		}
	}
	b.schema.AddField(field)
	return b
}

// AddSelectField adds a select field with options
func (b *Builder) AddSelectField(name, label string, required bool, options []Option) *Builder {
	b.schema.AddField(Field{
		Name:     name,
		Type:     FieldSelect,
		Label:    label,
		Required: required,
		Options:  options,
	})
	return b
}

// AddTextareaField adds a textarea field
func (b *Builder) AddTextareaField(name, label string, required bool, rows int) *Builder {
	field := Field{
		Name:     name,
		Type:     FieldTextarea,
		Label:    label,
		Required: required,
	}
	if rows > 0 {
		field.Config = map[string]any{"rows": rows}
	}
	b.schema.AddField(field)
	return b
}

// AddCheckboxField adds a checkbox field
func (b *Builder) AddCheckboxField(name, label string, defaultValue bool) *Builder {
	b.schema.AddField(Field{
		Name:    name,
		Type:    FieldCheckbox,
		Label:   label,
		Default: defaultValue,
	})
	return b
}

// AddDateField adds a date field
func (b *Builder) AddDateField(name, label string, required bool) *Builder {
	b.schema.AddField(Field{
		Name:     name,
		Type:     FieldDate,
		Label:    label,
		Required: required,
	})
	return b
}

// AddFieldWithConfig adds a fully configured field
func (b *Builder) AddFieldWithConfig(field Field) *Builder {
	b.schema.AddField(field)
	return b
}

// AddSubmitButton adds a submit button
func (b *Builder) AddSubmitButton(text string) *Builder {
	b.schema.AddAction(Action{
		ID:      "submit",
		Type:    ActionSubmit,
		Text:    text,
		Variant: "primary",
		Size:    "md",
	})
	return b
}

// AddResetButton adds a reset button
func (b *Builder) AddResetButton(text string) *Builder {
	b.schema.AddAction(Action{
		ID:      "reset",
		Type:    ActionReset,
		Text:    text,
		Variant: "secondary",
		Size:    "md",
	})
	return b
}

// AddButton adds a generic button
func (b *Builder) AddButton(id, text, variant string) *Builder {
	b.schema.AddAction(Action{
		ID:      id,
		Type:    ActionButton,
		Text:    text,
		Variant: variant,
		Size:    "md",
	})
	return b
}

// AddActionWithConfig adds a fully configured action
func (b *Builder) AddActionWithConfig(action Action) *Builder {
	b.schema.AddAction(action)
	return b
}

// Build returns the constructed schema
func (b *Builder) Build() (*Schema, error) {
	// Validate before returning
	if err := b.schema.Validate(); err != nil {
		return nil, err
	}
	return b.schema, nil
}

// MustBuild returns the schema or panics on error (useful for static definitions)
func (b *Builder) MustBuild() *Schema {
	schema, err := b.Build()
	if err != nil {
		panic(err)
	}
	return schema
}

// Helper functions for creating common configurations

// NewGridLayout creates a grid layout with specified columns
func NewGridLayout(columns int) *Layout {
	return &Layout{
		Type:       LayoutGrid,
		Columns:    columns,
		Gap:        "1rem",
		Responsive: true,
	}
}

// NewTabLayout creates a tabbed layout
func NewTabLayout(tabs []Tab) *Layout {
	return &Layout{
		Type: LayoutTabs,
		Tabs: tabs,
	}
}

// NewStepLayout creates a multi-step wizard layout
func NewStepLayout(steps []Step) *Layout {
	return &Layout{
		Type:  LayoutSteps,
		Steps: steps,
	}
}

// NewSimpleConfig creates a basic POST configuration
func NewSimpleConfig(action, method string) *Config {
	return &Config{
		Action:   action,
		Method:   method,
		Encoding: "application/json",
		Timeout:  30000,
	}
}

// CreateOption is a helper to create Option structs
func CreateOption(value, label string) Option {
	return Option{
		Value: value,
		Label: label,
	}
}

// CreateOptionWithIcon creates an option with an icon
func CreateOptionWithIcon(value, label, icon string) Option {
	return Option{
		Value: value,
		Label: label,
		Icon:  icon,
	}
}

// CreateGroupedOption creates an option with a group
func CreateGroupedOption(value, label, group string) Option {
	return Option{
		Value: value,
		Label: label,
		Group: group,
	}
}

// Field builder helpers

// FieldBuilder provides fluent interface for building fields
type FieldBuilder struct {
	field Field
}

// NewField starts building a field
func NewField(name string, fieldType FieldType) *FieldBuilder {
	return &FieldBuilder{
		field: Field{
			Name: name,
			Type: fieldType,
		},
	}
}

// WithLabel sets the label
func (fb *FieldBuilder) WithLabel(label string) *FieldBuilder {
	fb.field.Label = label
	return fb
}

// WithDescription sets the description
func (fb *FieldBuilder) WithDescription(desc string) *FieldBuilder {
	fb.field.Description = desc
	return fb
}

// WithPlaceholder sets the placeholder
func (fb *FieldBuilder) WithPlaceholder(placeholder string) *FieldBuilder {
	fb.field.Placeholder = placeholder
	return fb
}

// WithHelp sets help text
func (fb *FieldBuilder) WithHelp(help string) *FieldBuilder {
	fb.field.Help = help
	return fb
}

// Required marks field as required
func (fb *FieldBuilder) Required() *FieldBuilder {
	fb.field.Required = true
	return fb
}

// Disabled marks field as disabled
func (fb *FieldBuilder) Disabled() *FieldBuilder {
	fb.field.Disabled = true
	return fb
}

// Readonly marks field as readonly
func (fb *FieldBuilder) Readonly() *FieldBuilder {
	fb.field.Readonly = true
	return fb
}

// WithDefault sets default value
func (fb *FieldBuilder) WithDefault(value any) *FieldBuilder {
	fb.field.Default = value
	return fb
}

// WithOptions sets options (for select fields)
func (fb *FieldBuilder) WithOptions(options []Option) *FieldBuilder {
	fb.field.Options = options
	return fb
}

// WithValidation sets validation rules
func (fb *FieldBuilder) WithValidation(validation *FieldValidation) *FieldBuilder {
	fb.field.Validation = validation
	return fb
}

// Build returns the constructed field
func (fb *FieldBuilder) Build() Field {
	return fb.field
}
