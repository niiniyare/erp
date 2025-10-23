package bridge

import (
	"context"
	"fmt"
	"sync"

	schemaui "github.com/niiniyare/erp/pkg/schema/ui"
	"github.com/niiniyare/erp/web/components/atoms"
)

// ============================================================================
// UNIFIED REGISTRY
// ============================================================================

// UnifiedRegistry provides a unified interface for managing both schema and Templ components.
type UnifiedRegistry struct {
	schema schemaui.ComponentRegistry
	templ  *TemplRegistry
	bridge *Bridge
	mu     sync.RWMutex
}

// NewUnifiedRegistry creates a new unified registry instance.
func NewUnifiedRegistry() *UnifiedRegistry {
	return &UnifiedRegistry{
		schema: schemaui.NewRegistry(),
		templ:  newTemplRegistry(),
		bridge: NewBridge(),
	}
}

// CreateSchema creates a component using the schema system.
func (r *UnifiedRegistry) CreateSchema(ctx context.Context, typ schemaui.ComponentType, config map[string]any) (schemaui.Component, error) {
	return r.schema.Create(ctx, typ, config)
}

// CreateTempl creates a Templ component from schema configuration.
func (r *UnifiedRegistry) CreateTempl(ctx context.Context, typ schemaui.ComponentType, config map[string]any) (TemplComponent, error) {
	component, err := r.schema.Create(ctx, typ, config)
	if err != nil {
		return TemplComponent{}, fmt.Errorf("create schema component: %w", err)
	}

	return r.bridge.ConvertToTempl(ctx, component)
}

// ValidateSchema validates a schema component.
func (r *UnifiedRegistry) ValidateSchema(ctx context.Context, component schemaui.Component) error {
	return r.schema.Validate(ctx, component)
}

// ValidateTempl validates a Templ component by converting to schema first.
func (r *UnifiedRegistry) ValidateTempl(ctx context.Context, component TemplComponent) error {
	schema, err := r.bridge.ConvertTemplToSchema(ctx, component)
	if err != nil {
		return fmt.Errorf("convert to schema: %w", err)
	}

	return r.schema.Validate(ctx, schema)
}

// SupportedTypes returns all supported component types.
func (r *UnifiedRegistry) SupportedTypes() []schemaui.ComponentType {
	return r.schema.GetTypes()
}

// SchemaForType returns the schema definition for a component type.
func (r *UnifiedRegistry) SchemaForType(typ schemaui.ComponentType) (schemaui.ComponentSchema, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// TODO: Implement when GetAllSchemas method is available
	return schemaui.ComponentSchema{}, fmt.Errorf("GetAllSchemas method not implemented yet for type: %s", typ)
}

// ============================================================================
// TEMPLATE SYSTEM
// ============================================================================

// Template represents a reusable component template.
type Template interface {
	Name() string
	Build(ctx context.Context, data map[string]any, r *UnifiedRegistry) ([]TemplComponent, []schemaui.Component, error)
}

// templateRegistry holds registered templates.
var templateRegistry = make(map[string]Template)

func init() {
	RegisterTemplate(&loginFormTemplate{})
	RegisterTemplate(&userFormTemplate{})
	RegisterTemplate(&dataTableTemplate{})
}

// RegisterTemplate registers a template for use.
func RegisterTemplate(t Template) {
	templateRegistry[t.Name()] = t
}

// CreateFromTemplate creates components from a registered template.
func (r *UnifiedRegistry) CreateFromTemplate(ctx context.Context, name string, data map[string]any) ([]TemplComponent, []schemaui.Component, error) {
	template, exists := templateRegistry[name]
	if !exists {
		return nil, nil, fmt.Errorf("unknown template: %s", name)
	}

	return template.Build(ctx, data, r)
}

// ============================================================================
// TEMPLATE IMPLEMENTATIONS
// ============================================================================

type loginFormTemplate struct{}

func (t *loginFormTemplate) Name() string { return "login-form" }

func (t *loginFormTemplate) Build(ctx context.Context, data map[string]any, r *UnifiedRegistry) ([]TemplComponent, []schemaui.Component, error) {
	builder := newTemplateBuilder(ctx, r)

	builder.addInput(map[string]any{
		"input_type":  "email",
		"name":        "email",
		"placeholder": "Enter your email",
		"required":    true,
	})

	builder.addInput(map[string]any{
		"input_type":  "password",
		"name":        "password",
		"placeholder": "Enter your password",
		"required":    true,
	})

	builder.addButton(map[string]any{
		"text":        "Sign In",
		"button_type": "submit",
		"variant":     "primary",
	})

	return builder.build()
}

type userFormTemplate struct{}

func (t *userFormTemplate) Name() string { return "user-form" }

func (t *userFormTemplate) Build(ctx context.Context, data map[string]any, r *UnifiedRegistry) ([]TemplComponent, []schemaui.Component, error) {
	builder := newTemplateBuilder(ctx, r)

	builder.addInput(map[string]any{
		"input_type":  "text",
		"name":        "first_name",
		"placeholder": "First Name",
		"required":    true,
	})

	builder.addInput(map[string]any{
		"input_type":  "text",
		"name":        "last_name",
		"placeholder": "Last Name",
		"required":    true,
	})

	builder.addInput(map[string]any{
		"input_type":  "email",
		"name":        "email",
		"placeholder": "Email Address",
		"required":    true,
	})

	builder.addSelect(map[string]any{
		"name":        "role",
		"placeholder": "Select Role",
		"options": []map[string]any{
			{"value": "admin", "label": "Administrator"},
			{"value": "user", "label": "User"},
			{"value": "viewer", "label": "Viewer"},
		},
		"required": true,
	})

	builder.addButton(map[string]any{
		"text":        "Create User",
		"button_type": "submit",
		"variant":     "primary",
	})

	return builder.build()
}

type dataTableTemplate struct{}

func (t *dataTableTemplate) Name() string { return "data-table" }

func (t *dataTableTemplate) Build(ctx context.Context, data map[string]any, r *UnifiedRegistry) ([]TemplComponent, []schemaui.Component, error) {
	builder := newTemplateBuilder(ctx, r)

	builder.addInput(map[string]any{
		"input_type":  "search",
		"name":        "search",
		"placeholder": "Search...",
	})

	builder.addSelect(map[string]any{
		"name":        "status_filter",
		"placeholder": "Filter by status",
		"options": []map[string]any{
			{"value": "", "label": "All"},
			{"value": "active", "label": "Active"},
			{"value": "inactive", "label": "Inactive"},
		},
	})

	builder.addButton(map[string]any{
		"text":    "Add New",
		"icon":    "plus",
		"variant": "primary",
	})

	return builder.build()
}

// ============================================================================
// TEMPLATE BUILDER
// ============================================================================

// templateBuilder helps construct templates with error handling.
type templateBuilder struct {
	ctx      context.Context
	registry *UnifiedRegistry
	templ    []TemplComponent
	schema   []schemaui.Component
	err      error
}

func newTemplateBuilder(ctx context.Context, r *UnifiedRegistry) *templateBuilder {
	return &templateBuilder{
		ctx:      ctx,
		registry: r,
		templ:    make([]TemplComponent, 0),
		schema:   make([]schemaui.Component, 0),
	}
}

func (b *templateBuilder) addComponent(typ schemaui.ComponentType, config map[string]any) {
	if b.err != nil {
		return
	}

	schema, err := b.registry.CreateSchema(b.ctx, typ, config)
	if err != nil {
		b.err = fmt.Errorf("create %s: %w", typ, err)
		return
	}

	templ, err := b.registry.bridge.ConvertToTempl(b.ctx, schema)
	if err != nil {
		b.err = fmt.Errorf("convert %s: %w", typ, err)
		return
	}

	b.schema = append(b.schema, schema)
	b.templ = append(b.templ, templ)
}

func (b *templateBuilder) addButton(config map[string]any) {
	b.addComponent(schemaui.ComponentButton, config)
}

func (b *templateBuilder) addInput(config map[string]any) {
	b.addComponent(schemaui.ComponentInput, config)
}

func (b *templateBuilder) addSelect(config map[string]any) {
	b.addComponent(schemaui.ComponentSelect, config)
}

func (b *templateBuilder) addTextarea(config map[string]any) {
	b.addComponent(schemaui.ComponentTextarea, config)
}

func (b *templateBuilder) addCheckbox(config map[string]any) {
	b.addComponent(schemaui.ComponentCheckbox, config)
}

func (b *templateBuilder) addRadio(config map[string]any) {
	b.addComponent(schemaui.ComponentRadio, config)
}

func (b *templateBuilder) build() ([]TemplComponent, []schemaui.Component, error) {
	if b.err != nil {
		return nil, nil, b.err
	}
	return b.templ, b.schema, nil
}

// ============================================================================
// TEMPL REGISTRY
// ============================================================================

// TemplRegistry manages Templ component factories.
type TemplRegistry struct {
	factories map[string]Factory
	mu        sync.RWMutex
}

// Factory creates and validates Templ components.
type Factory interface {
	Create(ctx context.Context, config map[string]any) (TemplComponent, error)
	Validate(ctx context.Context, component TemplComponent) error
}

func newTemplRegistry() *TemplRegistry {
	r := &TemplRegistry{
		factories: make(map[string]Factory),
	}

	r.register("Button", &buttonFactory{})
	r.register("Input", &inputFactory{})
	r.register("Textarea", &textareaFactory{})
	r.register("Select", &selectFactory{})
	r.register("Checkbox", &checkboxFactory{})
	r.register("Radio", &radioFactory{})

	return r
}

func (r *TemplRegistry) register(typ string, f Factory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[typ] = f
}

func (r *TemplRegistry) Create(ctx context.Context, typ string, config map[string]any) (TemplComponent, error) {
	r.mu.RLock()
	factory, exists := r.factories[typ]
	r.mu.RUnlock()

	if !exists {
		return TemplComponent{}, fmt.Errorf("no factory for type: %s", typ)
	}

	return factory.Create(ctx, config)
}

// ============================================================================
// COMPONENT FACTORIES
// ============================================================================

type buttonFactory struct{}

func (f *buttonFactory) Create(ctx context.Context, config map[string]any) (TemplComponent, error) {
	props := atoms.ButtonProps{
		BaseProps: atoms.BaseProps{
			Name: getString(config, "name"),
		},
		InteractionProps: atoms.InteractionProps{
			Disabled: getBool(config, "disabled"),
		},
		Text:     getString(config, "text"),
		Icon:     atoms.IconProps{Name: getString(config, "icon")},
		Variant:  atoms.ButtonVariant(getString(config, "variant")),
		Size:     atoms.ButtonSize(getString(config, "size")),
	}

	return TemplComponent{Type: "Button", Props: props}, nil
}

func (f *buttonFactory) Validate(ctx context.Context, component TemplComponent) error {
	if component.Type != "Button" {
		return fmt.Errorf("expected Button, got %s", component.Type)
	}

	props, ok := component.Props.(atoms.ButtonProps)
	if !ok {
		return fmt.Errorf("invalid props type")
	}

	if props.Text == "" && props.Icon.Name == "" {
		return fmt.Errorf("button requires text or icon")
	}

	return nil
}

type inputFactory struct{}

func (f *inputFactory) Create(ctx context.Context, config map[string]any) (TemplComponent, error) {
	props := atoms.InputProps{
		BaseProps: atoms.BaseProps{
			Name: getString(config, "name"),
		},
		InteractionProps: atoms.InteractionProps{
			Required: getBool(config, "required"),
			Disabled: getBool(config, "disabled"),
		},
		PlaceholderProps: atoms.PlaceholderProps{
			Placeholder: getString(config, "placeholder"),
		},
		Type: stringToAtomsInputType(getString(config, "type")),
	}

	return TemplComponent{Type: "Input", Props: props}, nil
}

func (f *inputFactory) Validate(ctx context.Context, component TemplComponent) error {
	if component.Type != "Input" {
		return fmt.Errorf("expected Input, got %s", component.Type)
	}

	props, ok := component.Props.(atoms.InputProps)
	if !ok {
		return fmt.Errorf("invalid props type")
	}

	if props.BaseProps.Name == "" {
		return fmt.Errorf("input requires name")
	}

	return nil
}

type textareaFactory struct{}

func (f *textareaFactory) Create(ctx context.Context, config map[string]any) (TemplComponent, error) {
	props := atoms.TextareaProps{
		BaseProps: atoms.BaseProps{
			Name: getString(config, "name"),
		},
		InteractionProps: atoms.InteractionProps{
			Required: getBool(config, "required"),
			Disabled: getBool(config, "disabled"),
		},
		PlaceholderProps: atoms.PlaceholderProps{
			Placeholder: getString(config, "placeholder"),
		},
		Rows: getInt(config, "rows"),
	}

	return TemplComponent{Type: "Textarea", Props: props}, nil
}

func (f *textareaFactory) Validate(ctx context.Context, component TemplComponent) error {
	if component.Type != "Textarea" {
		return fmt.Errorf("expected Textarea, got %s", component.Type)
	}

	props, ok := component.Props.(atoms.TextareaProps)
	if !ok {
		return fmt.Errorf("invalid props type")
	}

	if props.BaseProps.Name == "" {
		return fmt.Errorf("textarea requires name")
	}

	return nil
}

type selectFactory struct{}

func (f *selectFactory) Create(ctx context.Context, config map[string]any) (TemplComponent, error) {
	props := atoms.SelectProps{
		BaseProps: atoms.BaseProps{
			Name: getString(config, "name"),
		},
		InteractionProps: atoms.InteractionProps{
			Required: getBool(config, "required"),
			Disabled: getBool(config, "disabled"),
		},
		PlaceholderProps: atoms.PlaceholderProps{
			Placeholder: getString(config, "placeholder"),
		},
		Multiple: getBool(config, "multiple"),
		Options:  parseSelectOptions(config["options"]),
	}

	return TemplComponent{Type: "Select", Props: props}, nil
}

func (f *selectFactory) Validate(ctx context.Context, component TemplComponent) error {
	if component.Type != "Select" {
		return fmt.Errorf("expected Select, got %s", component.Type)
	}

	props, ok := component.Props.(atoms.SelectProps)
	if !ok {
		return fmt.Errorf("invalid props type")
	}

	if props.BaseProps.Name == "" {
		return fmt.Errorf("select requires name")
	}

	if len(props.Options) == 0 {
		return fmt.Errorf("select requires at least one option")
	}

	return nil
}

type checkboxFactory struct{}

func (f *checkboxFactory) Create(ctx context.Context, config map[string]any) (TemplComponent, error) {
	props := atoms.CheckboxProps{
		BaseProps: atoms.BaseProps{
			Name: getString(config, "name"),
		},
		InteractionProps: atoms.InteractionProps{
			Required: getBool(config, "required"),
			Disabled: getBool(config, "disabled"),
		},
		LabelProps: atoms.LabelProps{
			Label: getString(config, "label"),
		},
		Checked: getBool(config, "checked"),
	}

	return TemplComponent{Type: "Checkbox", Props: props}, nil
}

func (f *checkboxFactory) Validate(ctx context.Context, component TemplComponent) error {
	if component.Type != "Checkbox" {
		return fmt.Errorf("expected Checkbox, got %s", component.Type)
	}

	props, ok := component.Props.(atoms.CheckboxProps)
	if !ok {
		return fmt.Errorf("invalid props type")
	}

	if props.BaseProps.Name == "" {
		return fmt.Errorf("checkbox requires name")
	}

	return nil
}

type radioFactory struct{}

func (f *radioFactory) Create(ctx context.Context, config map[string]any) (TemplComponent, error) {
	layout := "vertical"
	if getBool(config, "inline") {
		layout = "horizontal"
	}

	props := atoms.RadioGroupProps{
		BaseProps: atoms.BaseProps{
			Name: getString(config, "name"),
		},
		InteractionProps: atoms.InteractionProps{
			Required: getBool(config, "required"),
			Disabled: getBool(config, "disabled"),
		},
		Value:   getString(config, "value"),
		Layout:  layout,
		Options: parseRadioOptions(config["options"]),
	}

	return TemplComponent{Type: "RadioGroup", Props: props}, nil
}

func (f *radioFactory) Validate(ctx context.Context, component TemplComponent) error {
	if component.Type != "RadioGroup" {
		return fmt.Errorf("expected RadioGroup, got %s", component.Type)
	}

	props, ok := component.Props.(atoms.RadioGroupProps)
	if !ok {
		return fmt.Errorf("invalid props type")
	}

	if props.BaseProps.Name == "" {
		return fmt.Errorf("radio requires name")
	}

	if len(props.Options) == 0 {
		return fmt.Errorf("radio requires at least one option")
	}

	return nil
}

// ============================================================================
// UTILITY FUNCTIONS
// ============================================================================

func stringToAtomsInputType(inputType string) atoms.InputType {
	switch inputType {
	case "email":
		return atoms.InputEmail
	case "password":
		return atoms.InputPassword
	case "number":
		return atoms.InputNumber
	case "tel":
		return atoms.InputTel
	case "url":
		return atoms.InputURL
	case "search":
		return atoms.InputTypeSearch
	default:
		return atoms.InputText
	}
}

func getString(config map[string]any, key string) string {
	if v, ok := config[key].(string); ok {
		return v
	}
	return ""
}

func getBool(config map[string]any, key string) bool {
	if v, ok := config[key].(bool); ok {
		return v
	}
	return false
}

func getInt(config map[string]any, key string) int {
	if v, ok := config[key].(int); ok {
		return v
	}
	return 0
}

func parseSelectOptions(data any) []atoms.SelectOption {
	optionsData, ok := data.([]map[string]any)
	if !ok {
		return nil
	}

	options := make([]atoms.SelectOption, 0, len(optionsData))
	for _, opt := range optionsData {
		options = append(options, atoms.SelectOption{
			Value:    getString(opt, "value"),
			Label:    getString(opt, "label"),
			Disabled: getBool(opt, "disabled"),
		})
	}

	return options
}

func parseRadioOptions(data any) []atoms.RadioOption {
	optionsData, ok := data.([]map[string]any)
	if !ok {
		return nil
	}

	options := make([]atoms.RadioOption, 0, len(optionsData))
	for _, opt := range optionsData {
		options = append(options, atoms.RadioOption{
			Value:    getString(opt, "value"),
			Label:    getString(opt, "label"),
			Disabled: getBool(opt, "disabled"),
		})
	}

	return options
}
