package schema

import (
	"context"
	"encoding/json"
	"fmt"
)

// FormSchema provides a specialized interface for form-specific operations
// while leveraging the unified Schema architecture
type FormSchema struct {
	*Schema
	renderer FormRenderer
}

// FormRenderer interface for rendering form-specific templates
type FormRenderer interface {
	RenderForm(ctx context.Context, schema *Schema, data map[string]interface{}) (string, error)
	RenderField(ctx context.Context, field *Field, value interface{}) (string, error)
	RenderAction(ctx context.Context, action *Action) (string, error)
}

// NewFormSchema creates a new form schema with basic field support
func NewFormSchema(id, title string) *FormSchema {
	schema := NewSchema(id, TypeForm, title)

	// Set form-specific defaults
	if schema.Config == nil {
		schema.Config = &Config{
			Method:   "POST",
			Encoding: "application/json",
			Timeout:  30000,
		}
	}

	if schema.Layout == nil {
		schema.Layout = &Layout{
			Type:       LayoutGrid,
			Columns:    1,
			Gap:        "1rem",
			Direction:  "column",
			Responsive: true,
		}
	}

	return &FormSchema{
		Schema: schema,
	}
}

// AddBasicField adds a field with basic configuration
func (fs *FormSchema) AddBasicField(name string, fieldType FieldType, label string, required bool) *FormSchema {
	field := Field{
		Name:     name,
		Type:     fieldType,
		Label:    label,
		Required: required,
	}

	// Add basic validation if required
	if required {
		field.Validation = &FieldValidation{
			Required: true,
		}
	}

	fs.AddField(field)
	return fs
}

// AddTextField adds a text input field
func (fs *FormSchema) AddTextField(name, label, placeholder string, required bool) *FormSchema {
	field := Field{
		Name:        name,
		Type:        FieldText,
		Label:       label,
		Placeholder: placeholder,
		Required:    required,
	}

	if required {
		field.Validation = &FieldValidation{
			Required: true,
		}
	}

	fs.AddField(field)
	return fs
}

// AddEmailField adds an email input field with validation
func (fs *FormSchema) AddEmailField(name, label string, required bool) *FormSchema {
	field := Field{
		Name:        name,
		Type:        FieldEmail,
		Label:       label,
		Placeholder: "user@example.com",
		Required:    required,
		Validation: &FieldValidation{
			Required: required,
			Email:    true,
		},
	}

	fs.AddField(field)
	return fs
}

// AddNumberField adds a number input field with optional range validation
func (fs *FormSchema) AddNumberField(name, label string, required bool, min, max *float64) *FormSchema {
	field := Field{
		Name:     name,
		Type:     FieldNumber,
		Label:    label,
		Required: required,
		Validation: &FieldValidation{
			Required: required,
			Min:      min,
			Max:      max,
		},
	}

	fs.AddField(field)
	return fs
}

// AddSelectField adds a select field with options
func (fs *FormSchema) AddSelectField(name, label string, required bool, options []Option) *FormSchema {
	field := Field{
		Name:     name,
		Type:     FieldSelect,
		Label:    label,
		Required: required,
		Options:  options,
	}

	if required {
		field.Validation = &FieldValidation{
			Required: true,
		}
	}

	fs.AddField(field)
	return fs
}

// AddTextareaField adds a textarea field with optional length validation
func (fs *FormSchema) AddTextareaField(name, label string, required bool, minLength, maxLength *int) *FormSchema {
	field := Field{
		Name:     name,
		Type:     FieldTextarea,
		Label:    label,
		Required: required,
		Validation: &FieldValidation{
			Required:  required,
			MinLength: minLength,
			MaxLength: maxLength,
		},
	}

	fs.AddField(field)
	return fs
}

// AddDateField adds a date input field
func (fs *FormSchema) AddDateField(name, label string, required bool) *FormSchema {
	field := Field{
		Name:     name,
		Type:     FieldDate,
		Label:    label,
		Required: required,
	}

	if required {
		field.Validation = &FieldValidation{
			Required: true,
		}
	}

	fs.AddField(field)
	return fs
}

// AddCheckboxField adds a checkbox field
func (fs *FormSchema) AddCheckboxField(name, label string, required bool) *FormSchema {
	field := Field{
		Name:     name,
		Type:     FieldCheckbox,
		Label:    label,
		Required: required,
	}

	if required {
		field.Validation = &FieldValidation{
			Required: true,
		}
	}

	fs.AddField(field)
	return fs
}

// AddSubmitAction adds a submit button
func (fs *FormSchema) AddSubmitAction(text string, variant string) *FormSchema {
	action := Action{
		ID:      "submit",
		Type:    ActionSubmit,
		Text:    text,
		Variant: variant,
	}

	fs.AddAction(action)
	return fs
}

// AddResetAction adds a reset button
func (fs *FormSchema) AddResetAction(text string) *FormSchema {
	action := Action{
		ID:      "reset",
		Type:    ActionReset,
		Text:    text,
		Variant: "outline",
	}

	fs.AddAction(action)
	return fs
}

// SetFormAction sets the form submission URL and method
func (fs *FormSchema) SetFormAction(url, method string) *FormSchema {
	if fs.Config == nil {
		fs.Config = &Config{}
	}

	fs.Config.Action = url
	fs.Config.Method = method

	return fs
}

// SetLayout configures the form layout
func (fs *FormSchema) SetLayout(layoutType LayoutType, columns int) *FormSchema {
	if fs.Layout == nil {
		fs.Layout = &Layout{}
	}

	fs.Layout.Type = layoutType
	fs.Layout.Columns = columns

	return fs
}

// AddSection creates a new section with fields
func (fs *FormSchema) AddSection(id, title string, fieldNames []string) *FormSchema {
	section := Section{
		ID:     id,
		Title:  title,
		Fields: fieldNames,
	}

	if fs.Layout == nil {
		fs.Layout = &Layout{
			Type: LayoutSections,
		}
	}

	fs.Layout.Sections = append(fs.Layout.Sections, section)
	return fs
}

// EnableHTMX enables HTMX support for the form
func (fs *FormSchema) EnableHTMX(target, swap string) *FormSchema {
	fs.HTMX = &HTMX{
		Enabled: true,
		Target:  target,
		Swap:    swap,
	}

	return fs
}

// EnableAlpine enables Alpine.js support for the form
func (fs *FormSchema) EnableAlpine(xData string) *FormSchema {
	fs.Alpine = &Alpine{
		Enabled: true,
		XData:   xData,
	}

	return fs
}

// SetCSRFProtection enables CSRF protection
func (fs *FormSchema) SetCSRFProtection(tokenField string) *FormSchema {
	if fs.Security == nil {
		fs.Security = &Security{}
	}

	fs.Security.CSRF = &CSRF{
		Enabled:    true,
		TokenField: tokenField,
	}

	return fs
}

// SetValidationMode sets how validation is triggered
func (fs *FormSchema) SetValidationMode(mode ValidationMode) *FormSchema {
	if fs.Validation == nil {
		fs.Validation = &Validation{}
	}

	fs.Validation.Mode = mode
	return fs
}

// GetFieldByName returns a field by its name
func (fs *FormSchema) GetFieldByName(name string) (*Field, error) {
	for i, field := range fs.Fields {
		if field.Name == name {
			return &fs.Fields[i], nil
		}
	}
	return nil, fmt.Errorf("field '%s' not found", name)
}

// GetActionByID returns an action by its ID
func (fs *FormSchema) GetActionByID(id string) (*Action, error) {
	for i, action := range fs.Actions {
		if action.ID == id {
			return &fs.Actions[i], nil
		}
	}
	return nil, fmt.Errorf("action '%s' not found", id)
}

// ValidateForm validates the form schema structure
func (fs *FormSchema) ValidateForm(ctx context.Context) error {
	// Use the unified validation system
	if err := fs.Validate(); err != nil {
		return err
	}

	// Form-specific validations
	if len(fs.Fields) == 0 {
		return NewValidationError("no_fields", "form must have at least one field")
	}

	// Validate that form has at least one submit action
	hasSubmit := false
	for _, action := range fs.Actions {
		if action.Type == ActionSubmit {
			hasSubmit = true
			break
		}
	}

	if !hasSubmit {
		return NewValidationError("no_submit_action", "form must have at least one submit action")
	}

	return nil
}

// ValidateFormData validates form submission data
func (fs *FormSchema) ValidateFormData(ctx context.Context, data map[string]interface{}) (*ValidationResult, error) {
	// Use the comprehensive validation system from validation.go
	validator := NewValidator()
	return validator.ValidateData(ctx, fs.Schema, data), nil
}

// ToJSON converts the form schema to JSON
func (fs *FormSchema) ToJSON() ([]byte, error) {
	return json.MarshalIndent(fs.Schema, "", "  ")
}

// FromJSON loads a form schema from JSON
func (fs *FormSchema) FromJSON(data []byte) error {
	return json.Unmarshal(data, fs.Schema)
}

// SetRenderer sets the form renderer
func (fs *FormSchema) SetRenderer(renderer FormRenderer) {
	fs.renderer = renderer
}

// Render renders the complete form
func (fs *FormSchema) Render(ctx context.Context, data map[string]interface{}) (string, error) {
	if fs.renderer == nil {
		return "", NewRenderError("no_renderer", "form renderer not set")
	}

	return fs.renderer.RenderForm(ctx, fs.Schema, data)
}

// RenderField renders a specific field
func (fs *FormSchema) RenderField(ctx context.Context, fieldName string, value interface{}) (string, error) {
	if fs.renderer == nil {
		return "", NewRenderError("no_renderer", "form renderer not set")
	}

	field, err := fs.GetFieldByName(fieldName)
	if err != nil {
		return "", NewNotFoundError("field", err.Error())
	}

	return fs.renderer.RenderField(ctx, field, value)
}

// Clone creates a deep copy of the form schema
func (fs *FormSchema) Clone() *FormSchema {
	data, _ := fs.ToJSON()
	newSchema := &FormSchema{
		Schema: NewSchema(fs.ID+"_copy", TypeForm, fs.Title+" (Copy)"),
	}
	_ = newSchema.FromJSON(data)
	return newSchema
}

// GetFormStatistics returns basic form statistics
func (fs *FormSchema) GetFormStatistics() map[string]interface{} {
	stats := make(map[string]interface{})

	// Field counts
	fieldCounts := make(map[FieldType]int)
	requiredFields := 0

	for _, field := range fs.Fields {
		fieldCounts[field.Type]++
		if field.Required {
			requiredFields++
		}
	}

	stats["total_fields"] = len(fs.Fields)
	stats["required_fields"] = requiredFields
	stats["field_types"] = fieldCounts
	stats["total_actions"] = len(fs.Actions)
	stats["has_validation"] = fs.Validation != nil
	stats["has_workflow"] = fs.HasWorkflow()
	stats["has_htmx"] = fs.HasHTMX()
	stats["has_alpine"] = fs.HasAlpine()
	stats["has_security"] = fs.Security != nil
	stats["is_multi_tenant"] = fs.HasMultiTenant()

	if fs.Meta != nil {
		stats["created_at"] = fs.Meta.CreatedAt
		stats["updated_at"] = fs.Meta.UpdatedAt
	}

	return stats
}

// Builder pattern helpers for common form patterns

// NewContactForm creates a basic contact form
func NewContactForm() *FormSchema {
	fs := NewFormSchema("contact-form", "Contact Us")

	return fs.
		AddTextField("name", "Full Name", "Enter your full name", true).
		AddEmailField("email", "Email Address", true).
		AddTextField("subject", "Subject", "Enter subject", true).
		AddTextareaField("message", "Message", true, IntPtr(10), IntPtr(1000)).
		AddSubmitAction("Send Message", "primary").
		SetFormAction("/api/contact", "POST").
		SetValidationMode(ValidationOnChange)
}

// NewUserRegistrationForm creates a user registration form
func NewUserRegistrationForm() *FormSchema {
	fs := NewFormSchema("user-registration", "Create Account")

	return fs.
		AddTextField("firstName", "First Name", "Enter first name", true).
		AddTextField("lastName", "Last Name", "Enter last name", true).
		AddEmailField("email", "Email Address", true).
		AddBasicField("password", FieldPassword, "Password", true).
		AddBasicField("confirmPassword", FieldPassword, "Confirm Password", true).
		AddCheckboxField("terms", "I agree to the terms and conditions", true).
		AddSubmitAction("Create Account", "primary").
		AddResetAction("Clear Form").
		SetFormAction("/api/users/register", "POST").
		SetCSRFProtection("_token").
		SetValidationMode(ValidationOnBlur)
}

// NewProductForm creates a product creation form
func NewProductForm() *FormSchema {
	categories := []Option{
		{Value: "electronics", Label: "Electronics"},
		{Value: "clothing", Label: "Clothing"},
		{Value: "books", Label: "Books"},
		{Value: "home", Label: "Home & Garden"},
	}

	fs := NewFormSchema("product-form", "Add Product")

	return fs.
		AddTextField("name", "Product Name", "Enter product name", true).
		AddTextareaField("description", "Description", false, IntPtr(10), IntPtr(500)).
		AddSelectField("category", "Category", true, categories).
		AddNumberField("price", "Price", true, Float64Ptr(0), nil).
		AddNumberField("stock", "Stock Quantity", true, Float64Ptr(0), Float64Ptr(10000)).
		AddBasicField("image", FieldImage, "Product Image", false).
		AddSubmitAction("Save Product", "primary").
		AddResetAction("Clear Form").
		SetFormAction("/api/products", "POST").
		SetLayout(LayoutGrid, 2).
		SetValidationMode(ValidationOnChange)
}

// Form validation helpers
func (fs *FormSchema) HasRequiredFields() bool {
	for _, field := range fs.Fields {
		if field.Required {
			return true
		}
	}
	return false
}

func (fs *FormSchema) GetRequiredFieldNames() []string {
	var required []string
	for _, field := range fs.Fields {
		if field.Required {
			required = append(required, field.Name)
		}
	}
	return required
}

func (fs *FormSchema) GetFieldNames() []string {
	var names []string
	for _, field := range fs.Fields {
		names = append(names, field.Name)
	}
	return names
}

// UpdateField updates an existing field
func (fs *FormSchema) UpdateField(name string, updater func(*Field)) error {
	for i := range fs.Fields {
		if fs.Fields[i].Name == name {
			updater(&fs.Fields[i])
			fs.updateTimestamp()
			return nil
		}
	}
	return fmt.Errorf("field '%s' not found", name)
}

// RemoveField removes a field by name
func (fs *FormSchema) RemoveField(name string) error {
	for i, field := range fs.Fields {
		if field.Name == name {
			fs.Fields = append(fs.Fields[:i], fs.Fields[i+1:]...)
			fs.updateTimestamp()
			return nil
		}
	}
	return fmt.Errorf("field '%s' not found", name)
}
