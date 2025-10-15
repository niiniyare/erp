package bridge

import (
	"context"
	"fmt"
	"sync"

	schemaui "github.com/niiniyare/erp/pkg/schema/ui"
	"github.com/niiniyare/erp/web/components/atoms"
)

// ============================================================================
// UNIFIED BRIDGE REGISTRY
// ============================================================================

// UnifiedRegistry manages both schema components and Templ components through a single interface
type UnifiedRegistry struct {
	schemaRegistry schemaui.ComponentRegistry
	templRegistry  *TemplRegistry
	bridge         *Bridge
	mu             sync.RWMutex
}

// NewUnifiedRegistry creates a registry that supports both schema and Templ components
func NewUnifiedRegistry() *UnifiedRegistry {
	bridge := NewBridge()
	return &UnifiedRegistry{
		schemaRegistry: schemaui.NewRegistry(),
		templRegistry:  NewTemplRegistry(),
		bridge:         bridge,
	}
}

// ============================================================================
// UNIFIED COMPONENT CREATION
// ============================================================================

// CreateSchemaComponent creates a component using the schema system
func (r *UnifiedRegistry) CreateSchemaComponent(ctx context.Context, componentType schemaui.ComponentType, config map[string]any) (schemaui.Component, error) {
	return r.schemaRegistry.Create(ctx, componentType, config)
}

// CreateTemplComponent creates a Templ component from schema configuration
func (r *UnifiedRegistry) CreateTemplComponent(ctx context.Context, componentType schemaui.ComponentType, config map[string]any) (TemplComponent, error) {
	// First create schema component
	schemaComponent, err := r.schemaRegistry.Create(ctx, componentType, config)
	if err != nil {
		return TemplComponent{}, fmt.Errorf("failed to create schema component: %w", err)
	}

	// Convert to Templ component
	return r.bridge.ConvertSchemaToTempl(ctx, schemaComponent)
}

// CreateFromTemplate creates components from a pre-defined template
func (r *UnifiedRegistry) CreateFromTemplate(ctx context.Context, templateName string, data map[string]any) ([]TemplComponent, []schemaui.Component, error) {
	// This would load a template and create both representations
	// For now, we'll implement a basic form template
	
	switch templateName {
	case "login-form":
		return r.createLoginFormTemplate(ctx, data)
	case "user-form":
		return r.createUserFormTemplate(ctx, data)
	case "data-table":
		return r.createDataTableTemplate(ctx, data)
	default:
		return nil, nil, fmt.Errorf("unknown template: %s", templateName)
	}
}

// ============================================================================
// COMPONENT VALIDATION
// ============================================================================

// ValidateSchemaComponent validates a schema component
func (r *UnifiedRegistry) ValidateSchemaComponent(ctx context.Context, component schemaui.Component) error {
	return r.schemaRegistry.Validate(ctx, component)
}

// ValidateTemplComponent validates a Templ component by converting to schema first
func (r *UnifiedRegistry) ValidateTemplComponent(ctx context.Context, templComponent TemplComponent) error {
	// Convert to schema component
	schemaComponent, err := r.bridge.ConvertTemplToSchema(ctx, templComponent)
	if err != nil {
		return fmt.Errorf("failed to convert to schema for validation: %w", err)
	}

	// Validate using schema system
	return r.schemaRegistry.Validate(ctx, schemaComponent)
}

// ============================================================================
// REGISTRY INFORMATION
// ============================================================================

// GetSupportedTypes returns all supported component types from both systems
func (r *UnifiedRegistry) GetSupportedTypes() []schemaui.ComponentType {
	return r.schemaRegistry.GetTypes()
}

// GetSchemaForType returns the schema definition for a component type
func (r *UnifiedRegistry) GetSchemaForType(componentType schemaui.ComponentType) (schemaui.ComponentSchema, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Get all schemas and find the one for this type
	schemas := r.schemaRegistry.GetAllSchemas()
	if schema, exists := schemas[componentType]; exists {
		return schema, nil
	}

	return schemaui.ComponentSchema{}, fmt.Errorf("no schema found for component type: %s", componentType)
}

// ============================================================================
// TEMPLATE IMPLEMENTATIONS
// ============================================================================

// createLoginFormTemplate creates a complete login form in both representations
func (r *UnifiedRegistry) createLoginFormTemplate(ctx context.Context, data map[string]any) ([]TemplComponent, []schemaui.Component, error) {
	var templComponents []TemplComponent
	var schemaComponents []schemaui.Component

	// Email input
	emailSchema, err := r.schemaRegistry.Create(ctx, schemaui.ComponentInput, map[string]any{
		"input_type":  "email",
		"name":        "email",
		"placeholder": "Enter your email",
		"required":    true,
	})
	if err != nil {
		return nil, nil, err
	}
	schemaComponents = append(schemaComponents, emailSchema)

	emailTempl, err := r.bridge.ConvertSchemaToTempl(ctx, emailSchema)
	if err != nil {
		return nil, nil, err
	}
	templComponents = append(templComponents, emailTempl)

	// Password input
	passwordSchema, err := r.schemaRegistry.Create(ctx, schemaui.ComponentInput, map[string]any{
		"input_type":  "password",
		"name":        "password",
		"placeholder": "Enter your password",
		"required":    true,
	})
	if err != nil {
		return nil, nil, err
	}
	schemaComponents = append(schemaComponents, passwordSchema)

	passwordTempl, err := r.bridge.ConvertSchemaToTempl(ctx, passwordSchema)
	if err != nil {
		return nil, nil, err
	}
	templComponents = append(templComponents, passwordTempl)

	// Submit button
	submitSchema, err := r.schemaRegistry.Create(ctx, schemaui.ComponentButton, map[string]any{
		"text":        "Sign In",
		"button_type": "submit",
		"variant":     "primary",
	})
	if err != nil {
		return nil, nil, err
	}
	schemaComponents = append(schemaComponents, submitSchema)

	submitTempl, err := r.bridge.ConvertSchemaToTempl(ctx, submitSchema)
	if err != nil {
		return nil, nil, err
	}
	templComponents = append(templComponents, submitTempl)

	return templComponents, schemaComponents, nil
}

// createUserFormTemplate creates a complete user registration form
func (r *UnifiedRegistry) createUserFormTemplate(ctx context.Context, data map[string]any) ([]TemplComponent, []schemaui.Component, error) {
	var templComponents []TemplComponent
	var schemaComponents []schemaui.Component

	// First name input
	firstNameSchema, err := r.schemaRegistry.Create(ctx, schemaui.ComponentInput, map[string]any{
		"input_type":  "text",
		"name":        "first_name",
		"placeholder": "First Name",
		"required":    true,
	})
	if err != nil {
		return nil, nil, err
	}
	schemaComponents = append(schemaComponents, firstNameSchema)

	firstNameTempl, err := r.bridge.ConvertSchemaToTempl(ctx, firstNameSchema)
	if err != nil {
		return nil, nil, err
	}
	templComponents = append(templComponents, firstNameTempl)

	// Last name input
	lastNameSchema, err := r.schemaRegistry.Create(ctx, schemaui.ComponentInput, map[string]any{
		"input_type":  "text",
		"name":        "last_name",
		"placeholder": "Last Name",
		"required":    true,
	})
	if err != nil {
		return nil, nil, err
	}
	schemaComponents = append(schemaComponents, lastNameSchema)

	lastNameTempl, err := r.bridge.ConvertSchemaToTempl(ctx, lastNameSchema)
	if err != nil {
		return nil, nil, err
	}
	templComponents = append(templComponents, lastNameTempl)

	// Email input
	emailSchema, err := r.schemaRegistry.Create(ctx, schemaui.ComponentInput, map[string]any{
		"input_type":  "email",
		"name":        "email",
		"placeholder": "Email Address",
		"required":    true,
	})
	if err != nil {
		return nil, nil, err
	}
	schemaComponents = append(schemaComponents, emailSchema)

	emailTempl, err := r.bridge.ConvertSchemaToTempl(ctx, emailSchema)
	if err != nil {
		return nil, nil, err
	}
	templComponents = append(templComponents, emailTempl)

	// Role select
	roleSchema, err := r.schemaRegistry.Create(ctx, schemaui.ComponentSelect, map[string]any{
		"name":        "role",
		"placeholder": "Select Role",
		"options": []map[string]any{
			{"value": "admin", "label": "Administrator"},
			{"value": "user", "label": "User"},
			{"value": "viewer", "label": "Viewer"},
		},
		"required": true,
	})
	if err != nil {
		return nil, nil, err
	}
	schemaComponents = append(schemaComponents, roleSchema)

	roleTempl, err := r.bridge.ConvertSchemaToTempl(ctx, roleSchema)
	if err != nil {
		return nil, nil, err
	}
	templComponents = append(templComponents, roleTempl)

	// Submit button
	submitSchema, err := r.schemaRegistry.Create(ctx, schemaui.ComponentButton, map[string]any{
		"text":        "Create User",
		"button_type": "submit",
		"variant":     "primary",
	})
	if err != nil {
		return nil, nil, err
	}
	schemaComponents = append(schemaComponents, submitSchema)

	submitTempl, err := r.bridge.ConvertSchemaToTempl(ctx, submitSchema)
	if err != nil {
		return nil, nil, err
	}
	templComponents = append(templComponents, submitTempl)

	return templComponents, schemaComponents, nil
}

// createDataTableTemplate creates a data table configuration
func (r *UnifiedRegistry) createDataTableTemplate(ctx context.Context, data map[string]any) ([]TemplComponent, []schemaui.Component, error) {
	var templComponents []TemplComponent
	var schemaComponents []schemaui.Component

	// Search input
	searchSchema, err := r.schemaRegistry.Create(ctx, schemaui.ComponentInput, map[string]any{
		"input_type":  "search",
		"name":        "search",
		"placeholder": "Search...",
	})
	if err != nil {
		return nil, nil, err
	}
	schemaComponents = append(schemaComponents, searchSchema)

	searchTempl, err := r.bridge.ConvertSchemaToTempl(ctx, searchSchema)
	if err != nil {
		return nil, nil, err
	}
	templComponents = append(templComponents, searchTempl)

	// Filter select
	filterSchema, err := r.schemaRegistry.Create(ctx, schemaui.ComponentSelect, map[string]any{
		"name":        "status_filter",
		"placeholder": "Filter by status",
		"options": []map[string]any{
			{"value": "", "label": "All"},
			{"value": "active", "label": "Active"},
			{"value": "inactive", "label": "Inactive"},
		},
	})
	if err != nil {
		return nil, nil, err
	}
	schemaComponents = append(schemaComponents, filterSchema)

	filterTempl, err := r.bridge.ConvertSchemaToTempl(ctx, filterSchema)
	if err != nil {
		return nil, nil, err
	}
	templComponents = append(templComponents, filterTempl)

	// Add button
	addSchema, err := r.schemaRegistry.Create(ctx, schemaui.ComponentButton, map[string]any{
		"text":    "Add New",
		"icon":    "plus",
		"variant": "primary",
	})
	if err != nil {
		return nil, nil, err
	}
	schemaComponents = append(schemaComponents, addSchema)

	addTempl, err := r.bridge.ConvertSchemaToTempl(ctx, addSchema)
	if err != nil {
		return nil, nil, err
	}
	templComponents = append(templComponents, addTempl)

	return templComponents, schemaComponents, nil
}

// ============================================================================
// TEMPL REGISTRY FOR COMPONENT FACTORIES
// ============================================================================

// TemplRegistry manages Templ component creation and validation
type TemplRegistry struct {
	factories map[string]TemplComponentFactory
	mu        sync.RWMutex
}

// TemplComponentFactory creates Templ components
type TemplComponentFactory interface {
	Create(ctx context.Context, config map[string]any) (TemplComponent, error)
	Validate(ctx context.Context, component TemplComponent) error
}

// NewTemplRegistry creates a new Templ component registry
func NewTemplRegistry() *TemplRegistry {
	registry := &TemplRegistry{
		factories: make(map[string]TemplComponentFactory),
	}

	// Register Templ component factories
	registry.Register("Button", &ButtonTemplFactory{})
	registry.Register("Input", &InputTemplFactory{})
	registry.Register("Textarea", &TextareaTemplFactory{})
	registry.Register("Select", &SelectTemplFactory{})
	registry.Register("Checkbox", &CheckboxTemplFactory{})
	registry.Register("Radio", &RadioTemplFactory{})

	return registry
}

// Register registers a Templ component factory
func (r *TemplRegistry) Register(componentType string, factory TemplComponentFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[componentType] = factory
}

// Create creates a Templ component using the registered factory
func (r *TemplRegistry) Create(ctx context.Context, componentType string, config map[string]any) (TemplComponent, error) {
	r.mu.RLock()
	factory, exists := r.factories[componentType]
	r.mu.RUnlock()

	if !exists {
		return TemplComponent{}, fmt.Errorf("no factory registered for Templ component type: %s", componentType)
	}

	return factory.Create(ctx, config)
}

// ============================================================================
// TEMPL COMPONENT FACTORIES
// ============================================================================

// ButtonTemplFactory creates Button Templ components
type ButtonTemplFactory struct{}

func (f *ButtonTemplFactory) Create(ctx context.Context, config map[string]any) (TemplComponent, error) {
	props := atoms.ButtonProps{}
	
	if text, ok := config["text"].(string); ok {
		props.Text = text
	}
	if icon, ok := config["icon"].(string); ok {
		props.Icon = icon
	}
	if variant, ok := config["variant"].(string); ok {
		props.Variant = atoms.ButtonVariant(variant)
	}
	if size, ok := config["size"].(string); ok {
		props.Size = atoms.ButtonSize(size)
	}
	if disabled, ok := config["disabled"].(bool); ok {
		props.Disabled = disabled
	}

	return TemplComponent{
		Type:  "Button",
		Props: props,
	}, nil
}

func (f *ButtonTemplFactory) Validate(ctx context.Context, component TemplComponent) error {
	if component.Type != "Button" {
		return fmt.Errorf("invalid component type for ButtonTemplFactory: %s", component.Type)
	}
	
	props, ok := component.Props.(atoms.ButtonProps)
	if !ok {
		return fmt.Errorf("invalid props type for Button component")
	}
	
	if props.Text == "" && props.Icon == "" {
		return fmt.Errorf("button must have either text or icon")
	}
	
	return nil
}

// InputTemplFactory creates Input Templ components
type InputTemplFactory struct{}

func (f *InputTemplFactory) Create(ctx context.Context, config map[string]any) (TemplComponent, error) {
	props := atoms.InputProps{}
	
	if inputType, ok := config["type"].(string); ok {
		props.Type = inputType
	}
	if name, ok := config["name"].(string); ok {
		props.Name = name
	}
	if placeholder, ok := config["placeholder"].(string); ok {
		props.Placeholder = placeholder
	}
	if required, ok := config["required"].(bool); ok {
		props.Required = required
	}
	if disabled, ok := config["disabled"].(bool); ok {
		props.Disabled = disabled
	}

	return TemplComponent{
		Type:  "Input",
		Props: props,
	}, nil
}

func (f *InputTemplFactory) Validate(ctx context.Context, component TemplComponent) error {
	if component.Type != "Input" {
		return fmt.Errorf("invalid component type for InputTemplFactory: %s", component.Type)
	}
	
	props, ok := component.Props.(atoms.InputProps)
	if !ok {
		return fmt.Errorf("invalid props type for Input component")
	}
	
	if props.Name == "" {
		return fmt.Errorf("input must have a name")
	}
	
	return nil
}

// TextareaTemplFactory creates Textarea Templ components
type TextareaTemplFactory struct{}

func (f *TextareaTemplFactory) Create(ctx context.Context, config map[string]any) (TemplComponent, error) {
	props := atoms.TextareaProps{}
	
	if name, ok := config["name"].(string); ok {
		props.Name = name
	}
	if placeholder, ok := config["placeholder"].(string); ok {
		props.Placeholder = placeholder
	}
	if rows, ok := config["rows"].(int); ok {
		props.Rows = rows
	}
	if required, ok := config["required"].(bool); ok {
		props.Required = required
	}
	if disabled, ok := config["disabled"].(bool); ok {
		props.Disabled = disabled
	}

	return TemplComponent{
		Type:  "Textarea",
		Props: props,
	}, nil
}

func (f *TextareaTemplFactory) Validate(ctx context.Context, component TemplComponent) error {
	if component.Type != "Textarea" {
		return fmt.Errorf("invalid component type for TextareaTemplFactory: %s", component.Type)
	}
	
	props, ok := component.Props.(atoms.TextareaProps)
	if !ok {
		return fmt.Errorf("invalid props type for Textarea component")
	}
	
	if props.Name == "" {
		return fmt.Errorf("textarea must have a name")
	}
	
	return nil
}

// SelectTemplFactory creates Select Templ components
type SelectTemplFactory struct{}

func (f *SelectTemplFactory) Create(ctx context.Context, config map[string]any) (TemplComponent, error) {
	props := atoms.SelectProps{}
	
	if name, ok := config["name"].(string); ok {
		props.Name = name
	}
	if placeholder, ok := config["placeholder"].(string); ok {
		props.Placeholder = placeholder
	}
	if multiple, ok := config["multiple"].(bool); ok {
		props.Multiple = multiple
	}
	if required, ok := config["required"].(bool); ok {
		props.Required = required
	}
	if disabled, ok := config["disabled"].(bool); ok {
		props.Disabled = disabled
	}

	// Handle options
	if optionsData, ok := config["options"].([]map[string]any); ok {
		var options []atoms.SelectOption
		for _, optionData := range optionsData {
			option := atoms.SelectOption{}
			if value, ok := optionData["value"].(string); ok {
				option.Value = value
			}
			if label, ok := optionData["label"].(string); ok {
				option.Label = label
			}
			if disabled, ok := optionData["disabled"].(bool); ok {
				option.Disabled = disabled
			}
			options = append(options, option)
		}
		props.Options = options
	}

	return TemplComponent{
		Type:  "Select",
		Props: props,
	}, nil
}

func (f *SelectTemplFactory) Validate(ctx context.Context, component TemplComponent) error {
	if component.Type != "Select" {
		return fmt.Errorf("invalid component type for SelectTemplFactory: %s", component.Type)
	}
	
	props, ok := component.Props.(atoms.SelectProps)
	if !ok {
		return fmt.Errorf("invalid props type for Select component")
	}
	
	if props.Name == "" {
		return fmt.Errorf("select must have a name")
	}
	
	if len(props.Options) == 0 {
		return fmt.Errorf("select must have at least one option")
	}
	
	return nil
}

// CheckboxTemplFactory creates Checkbox Templ components
type CheckboxTemplFactory struct{}

func (f *CheckboxTemplFactory) Create(ctx context.Context, config map[string]any) (TemplComponent, error) {
	props := atoms.CheckboxProps{}
	
	if name, ok := config["name"].(string); ok {
		props.Name = name
	}
	if label, ok := config["label"].(string); ok {
		props.Label = label
	}
	if checked, ok := config["checked"].(bool); ok {
		props.Checked = checked
	}
	if required, ok := config["required"].(bool); ok {
		props.Required = required
	}
	if disabled, ok := config["disabled"].(bool); ok {
		props.Disabled = disabled
	}

	return TemplComponent{
		Type:  "Checkbox",
		Props: props,
	}, nil
}

func (f *CheckboxTemplFactory) Validate(ctx context.Context, component TemplComponent) error {
	if component.Type != "Checkbox" {
		return fmt.Errorf("invalid component type for CheckboxTemplFactory: %s", component.Type)
	}
	
	props, ok := component.Props.(atoms.CheckboxProps)
	if !ok {
		return fmt.Errorf("invalid props type for Checkbox component")
	}
	
	if props.Name == "" {
		return fmt.Errorf("checkbox must have a name")
	}
	
	return nil
}

// RadioTemplFactory creates Radio Templ components
type RadioTemplFactory struct{}

func (f *RadioTemplFactory) Create(ctx context.Context, config map[string]any) (TemplComponent, error) {
	props := atoms.RadioProps{}
	
	if name, ok := config["name"].(string); ok {
		props.Name = name
	}
	if value, ok := config["value"].(string); ok {
		props.Value = value
	}
	if inline, ok := config["inline"].(bool); ok {
		props.Inline = inline
	}
	if required, ok := config["required"].(bool); ok {
		props.Required = required
	}
	if disabled, ok := config["disabled"].(bool); ok {
		props.Disabled = disabled
	}

	// Handle options
	if optionsData, ok := config["options"].([]map[string]any); ok {
		var options []atoms.RadioOption
		for _, optionData := range optionsData {
			option := atoms.RadioOption{}
			if value, ok := optionData["value"].(string); ok {
				option.Value = value
			}
			if label, ok := optionData["label"].(string); ok {
				option.Label = label
			}
			if disabled, ok := optionData["disabled"].(bool); ok {
				option.Disabled = disabled
			}
			options = append(options, option)
		}
		props.Options = options
	}

	return TemplComponent{
		Type:  "Radio",
		Props: props,
	}, nil
}

func (f *RadioTemplFactory) Validate(ctx context.Context, component TemplComponent) error {
	if component.Type != "Radio" {
		return fmt.Errorf("invalid component type for RadioTemplFactory: %s", component.Type)
	}
	
	props, ok := component.Props.(atoms.RadioProps)
	if !ok {
		return fmt.Errorf("invalid props type for Radio component")
	}
	
	if props.Name == "" {
		return fmt.Errorf("radio must have a name")
	}
	
	if len(props.Options) == 0 {
		return fmt.Errorf("radio must have at least one option")
	}
	
	return nil
}