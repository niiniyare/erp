package schema

import (
	"context"

	"github.com/niiniyare/erp/pkg/condition"
)

// Builder provides a fluent interface for building schemas programmatically
type Builder struct {
	schema       *Schema
	mixinSupport *MixinRegistry
	validator    Validator // Use interface instead of concrete type
	ruleEngine   *BusinessRuleEngine
	evaluator    *condition.Evaluator
}

// NewBuilder creates a new schema builder
func NewBuilder(id string, schemaType Type, title string) *Builder {
	return &Builder{
		schema:       NewSchema(id, schemaType, title),
		mixinSupport: NewMixinRegistry(),
		validator:    nil, // Will be set via WithValidator
		ruleEngine:   NewBusinessRuleEngine(),
	}
}

// WithEvaluator sets the condition evaluator for all fields
func (b *Builder) WithEvaluator(evaluator *condition.Evaluator) *Builder {
	b.evaluator = evaluator
	return b
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
	field := Field{
		Name:     name,
		Type:     FieldText,
		Label:    label,
		Required: required,
	}
	if b.evaluator != nil {
		field.SetEvaluator(b.evaluator)
	}
	b.schema.AddField(field)
	return b
}

// AddEmailField adds an email field
func (b *Builder) AddEmailField(name, label string, required bool) *Builder {
	field := Field{
		Name:     name,
		Type:     FieldEmail,
		Label:    label,
		Required: required,
	}
	if b.evaluator != nil {
		field.SetEvaluator(b.evaluator)
	}
	b.schema.AddField(field)
	return b
}

// AddPasswordField adds a password field
func (b *Builder) AddPasswordField(name, label string, required bool) *Builder {
	field := Field{
		Name:     name,
		Type:     FieldPassword,
		Label:    label,
		Required: required,
	}
	if b.evaluator != nil {
		field.SetEvaluator(b.evaluator)
	}
	b.schema.AddField(field)
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
	if b.evaluator != nil {
		field.SetEvaluator(b.evaluator)
	}
	b.schema.AddField(field)
	return b
}

// AddSelectField adds a select field with options
func (b *Builder) AddSelectField(name, label string, required bool, options []Option) *Builder {
	field := Field{
		Name:     name,
		Type:     FieldSelect,
		Label:    label,
		Required: required,
		Options:  options,
	}
	if b.evaluator != nil {
		field.SetEvaluator(b.evaluator)
	}
	b.schema.AddField(field)
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
	if b.evaluator != nil {
		field.SetEvaluator(b.evaluator)
	}
	b.schema.AddField(field)
	return b
}

// AddCheckboxField adds a checkbox field
func (b *Builder) AddCheckboxField(name, label string, defaultValue bool) *Builder {
	field := Field{
		Name:    name,
		Type:    FieldCheckbox,
		Label:   label,
		Default: defaultValue,
	}
	if b.evaluator != nil {
		field.SetEvaluator(b.evaluator)
	}
	b.schema.AddField(field)
	return b
}

// AddDateField adds a date field
func (b *Builder) AddDateField(name, label string, required bool) *Builder {
	field := Field{
		Name:     name,
		Type:     FieldDate,
		Label:    label,
		Required: required,
	}
	if b.evaluator != nil {
		field.SetEvaluator(b.evaluator)
	}
	b.schema.AddField(field)
	return b
}

// AddFieldWithConfig adds a fully configured field
func (b *Builder) AddFieldWithConfig(field Field) *Builder {
	if b.evaluator != nil {
		field.SetEvaluator(b.evaluator)
	}
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

// Build returns the constructed schema with evaluator set on all fields
func (b *Builder) Build(ctx context.Context) (*Schema, error) {
	// Set evaluator on all fields
	if b.evaluator != nil {
		for i := range b.schema.Fields {
			b.schema.Fields[i].SetEvaluator(b.evaluator)
		}
	}

	// Validate before returning
	if err := b.schema.Validate(ctx); err != nil {
		return nil, err
	}
	return b.schema, nil
}

// MustBuild returns the schema or panics on error (useful for static definitions)
func (b *Builder) MustBuild(ctx context.Context) *Schema {
	schema, err := b.Build(ctx)
	if err != nil {
		panic(err)
	}
	return schema
}

// WithMixin applies a mixin to the schema with evaluator support
func (b *Builder) WithMixin(mixinID string) *Builder {
	if err := b.mixinSupport.ApplyMixinWithEvaluator(b.schema, mixinID, "", b.evaluator); err == nil {
		// Track applied mixin
		if b.schema.Meta == nil {
			b.schema.Meta = &Meta{CustomData: make(map[string]any)}
		} else if b.schema.Meta.CustomData == nil {
			b.schema.Meta.CustomData = make(map[string]any)
		}
		if appliedMixins, exists := b.schema.Meta.CustomData["applied_mixins"]; exists {
			if mixinList, ok := appliedMixins.([]string); ok {
				b.schema.Meta.CustomData["applied_mixins"] = append(mixinList, mixinID)
			}
		} else {
			b.schema.Meta.CustomData["applied_mixins"] = []string{mixinID}
		}
	}
	return b
}

// WithCustomMixin applies a custom mixin
func (b *Builder) WithCustomMixin(mixin *Mixin) *Builder {
	if err := b.mixinSupport.Register(mixin); err == nil {
		b.WithMixin(mixin.ID)
	}
	return b
}

// WithRepeatable adds a repeatable field with evaluator support
func (b *Builder) WithRepeatable(field *RepeatableField) *Builder {
	if b.evaluator != nil {
		field.SetEvaluator(b.evaluator)
	}
	b.schema.AddField(field.Field)
	return b
}

// WithValidator sets a custom validator
func (b *Builder) WithValidator(validator Validator) *Builder {
	b.validator = validator
	return b
}

// WithCustomValidator adds a custom validator (placeholder - implement with concrete validator)
func (b *Builder) WithCustomValidator(name string, validatorFunc func(context.Context, any, map[string]any) error) *Builder {
	// Implementation depends on concrete validator type
	return b
}

// GetMixinRegistry returns the mixin registry for advanced operations
func (b *Builder) GetMixinRegistry() *MixinRegistry {
	return b.mixinSupport
}

// GetValidator returns the validator for advanced operations
func (b *Builder) GetValidator() Validator {
	return b.validator
}

// WithBusinessRule adds a business rule to the schema
func (b *Builder) WithBusinessRule(rule *BusinessRule) *Builder {
	b.ruleEngine.AddRule(rule)
	return b
}

// WithBusinessRuleBuilder adds a business rule using the builder pattern
func (b *Builder) WithBusinessRuleBuilder(ruleBuilder *BusinessRuleBuilder) *Builder {
	if rule, err := ruleBuilder.Build(); err == nil {
		b.ruleEngine.AddRule(rule)
	}
	return b
}

// GetBusinessRuleEngine returns the business rule engine for advanced operations
func (b *Builder) GetBusinessRuleEngine() *BusinessRuleEngine {
	return b.ruleEngine
}

// ApplyBusinessRules applies business rules to the schema based on data
func (b *Builder) ApplyBusinessRules(ctx context.Context, data map[string]any) (*Schema, error) {
	return b.ruleEngine.ApplyRules(ctx, b.schema, data)
}

// BuildWithRules builds the schema and applies business rules based on provided data
func (b *Builder) BuildWithRules(ctx context.Context, data map[string]any) (*Schema, error) {
	// First build the base schema
	schema, err := b.Build(ctx)
	if err != nil {
		return nil, err
	}

	// Then apply business rules
	return b.ruleEngine.ApplyRules(ctx, schema, data)
}

// Helper functions for creating common configurations

// NewGridLayout creates a grid layout with specified columns
func NewGridLayout(columns int) *Layout {
	return &Layout{
		Type:       "grid",
		Columns:    columns,
		Gap:        "1rem",
		Responsive: true,
	}
}

// NewTabLayout creates a tabbed layout
func NewTabLayout(tabs []Tab) *Layout {
	return &Layout{
		Type: "tabs",
		Tabs: tabs,
	}
}

// NewStepLayout creates a multi-step wizard layout
func NewStepLayout(steps []Step) *Layout {
	return &Layout{
		Type:  "steps",
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
	field     Field
	evaluator *condition.Evaluator
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

// WithEvaluator sets the condition evaluator
func (fb *FieldBuilder) WithEvaluator(evaluator *condition.Evaluator) *FieldBuilder {
	fb.evaluator = evaluator
	return fb
}

// Build returns the constructed field
func (fb *FieldBuilder) Build() Field {
	if fb.evaluator != nil {
		fb.field.SetEvaluator(fb.evaluator)
	}
	return fb.field
}

// Foundation helper methods for common schema patterns

// NewInvoiceFormBuilder creates a builder for invoice forms with line items
func NewInvoiceFormBuilder() *Builder {
	builder := NewBuilder("invoice_form", TypeForm, "Invoice Form")

	// Apply audit fields mixin
	builder.WithMixin("audit_fields")

	// Add basic invoice fields
	builder.AddTextField("invoice_number", "Invoice Number", true)
	builder.AddDateField("invoice_date", "Invoice Date", true)
	builder.AddDateField("due_date", "Due Date", true)

	// Add customer information
	builder.AddTextField("customer_name", "Customer Name", true)
	builder.AddTextField("customer_email", "Customer Email", false)

	// Add repeatable line items
	lineItems := CreateInvoiceLineItemsField()
	builder.WithRepeatable(lineItems)

	// Add totals (readonly calculated fields)
	builder.AddFieldWithConfig(Field{
		Name:     "subtotal",
		Type:     FieldCurrency,
		Label:    "Subtotal",
		Readonly: true,
	})
	builder.AddFieldWithConfig(Field{
		Name:     "tax_amount",
		Type:     FieldCurrency,
		Label:    "Tax Amount",
		Readonly: true,
	})
	builder.AddFieldWithConfig(Field{
		Name:     "total",
		Type:     FieldCurrency,
		Label:    "Total",
		Readonly: true,
	})

	// Add actions
	builder.AddSubmitButton("Save Invoice")
	builder.AddButton("preview", "Preview", "secondary")
	builder.AddButton("send", "Send", "primary")

	return builder
}

// NewContactFormBuilder creates a builder for contact forms
func NewContactFormBuilder() *Builder {
	builder := NewBuilder("contact_form", TypeForm, "Contact Form")

	// Apply contact fields and audit fields mixins
	builder.WithMixin("contact_fields")
	builder.WithMixin("audit_fields")

	// Add additional contact-specific fields
	builder.AddSelectField("contact_type", "Contact Type", true, []Option{
		CreateOption("customer", "Customer"),
		CreateOption("supplier", "Supplier"),
		CreateOption("employee", "Employee"),
		CreateOption("other", "Other"),
	})

	builder.AddTextareaField("notes", "Notes", false, 4)

	// Add custom validation
	builder.WithCustomValidator("phone_required_for_customers", func(ctx context.Context, value any, params map[string]any) error {
		// Custom business logic
		return nil
	})

	builder.AddSubmitButton("Save Contact")

	return builder
}

// NewAddressFormBuilder creates a builder for address forms
func NewAddressFormBuilder() *Builder {
	builder := NewBuilder("address_form", TypeForm, "Address Form")

	// Apply address fields mixin
	builder.WithMixin("address_fields")

	// Add address type selection
	builder.AddSelectField("address_type", "Address Type", true, []Option{
		CreateOption("billing", "Billing Address"),
		CreateOption("shipping", "Shipping Address"),
		CreateOption("main", "Main Address"),
	})

	builder.AddCheckboxField("is_default", "Set as Default", false)

	builder.AddSubmitButton("Save Address")

	return builder
}

// NewUserRegistrationBuilder creates a builder for user registration with validation
func NewUserRegistrationBuilder() *Builder {
	builder := NewBuilder("user_registration", TypeForm, "User Registration")

	// Add user fields
	builder.AddTextField("username", "Username", true)
	builder.AddEmailField("email", "Email Address", true)
	builder.AddPasswordField("password", "Password", true)
	builder.AddPasswordField("password_confirmation", "Confirm Password", true)

	// Add personal information
	builder.AddTextField("first_name", "First Name", true)
	builder.AddTextField("last_name", "Last Name", true)
	builder.AddDateField("birth_date", "Date of Birth", false)

	// Add custom validators (would implement with concrete validator)
	builder.WithCustomValidator("username_available", func(ctx context.Context, value any, params map[string]any) error {
		// Would check database for username availability
		return nil
	})

	builder.WithCustomValidator("email_available", func(ctx context.Context, value any, params map[string]any) error {
		// Would check database for email availability
		return nil
	})

	builder.AddCheckboxField("terms_accepted", "I accept the Terms of Service", false)
	builder.AddSubmitButton("Create Account")

	return builder
}

// NewTaskFormBuilder creates a builder for task management forms
func NewTaskFormBuilder() *Builder {
	builder := NewBuilder("task_form", TypeForm, "Task Form")

	// Apply audit fields for tracking
	builder.WithMixin("audit_fields")

	// Add task fields
	builder.AddTextField("title", "Task Title", true)
	builder.AddTextareaField("description", "Description", false, 6)

	builder.AddSelectField("priority", "Priority", true, []Option{
		CreateOption("low", "Low"),
		CreateOption("medium", "Medium"),
		CreateOption("high", "High"),
		CreateOption("urgent", "Urgent"),
	})

	builder.AddSelectField("status", "Status", true, []Option{
		CreateOption("todo", "To Do"),
		CreateOption("in_progress", "In Progress"),
		CreateOption("review", "Under Review"),
		CreateOption("done", "Completed"),
	})

	builder.AddDateField("due_date", "Due Date", false)
	builder.AddTextField("assigned_to", "Assigned To", false)

	// Add subtasks as repeatable field
	subtaskTemplate := []Field{
		{
			Name:     "title",
			Type:     FieldText,
			Label:    "Subtask Title",
			Required: true,
		},
		{
			Name:    "completed",
			Type:    FieldCheckbox,
			Label:   "Completed",
			Default: false,
		},
	}

	subtasks, _ := NewRepeatableField("subtasks", "Subtasks").
		WithTemplate(subtaskTemplate).
		WithMinItems(0).
		WithMaxItems(20).
		WithItemLabel("Subtask {index}").
		WithTexts("Add Subtask", "Remove").
		Build()

	builder.WithRepeatable(subtasks)

	builder.AddSubmitButton("Save Task")
	builder.AddButton("assign", "Assign", "secondary")

	return builder
}
