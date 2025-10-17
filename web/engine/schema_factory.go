package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/a-h/templ"
	"github.com/niiniyare/erp/pkg/schema/ui/css"
	"github.com/niiniyare/erp/web/components/atoms"
)

// SchemaFactory creates Templ components directly from JSON schema definitions
// This provides a schema-first approach that leverages the 913+ JSON schemas in docs/ui/Schema/definitions/
type SchemaFactory struct {
	schemaRegistry   map[string]*JsonSchema
	componentFactory map[string]ComponentRenderer
	cssFactory       *css.Factory
	validationEngine *ValidationEngine
	templRenderer    *SchemaTemplRenderer
	schemaDir        string
}

// JsonSchema represents a parsed JSON schema from docs/ui/Schema/definitions/
type JsonSchema struct {
	ID                   string                 `json:"$id"`
	Schema               string                 `json:"$schema"`
	Type                 string                 `json:"type"`
	Properties           map[string]Property    `json:"properties"`
	Required             []string               `json:"required"`
	AdditionalProperties bool                   `json:"additionalProperties"`
	Description          string                 `json:"description"`
	RawSchema            map[string]interface{} `json:"-"`
}

// Property represents a schema property definition
type Property struct {
	Type        interface{} `json:"type"`
	Description string      `json:"description"`
	Enum        []string    `json:"enum,omitempty"`
	Const       interface{} `json:"const,omitempty"`
	Ref         string      `json:"$ref,omitempty"`
	AnyOf       []Property  `json:"anyOf,omitempty"`
	OneOf       []Property  `json:"oneOf,omitempty"`
	Items       *Property   `json:"items,omitempty"`
	Format      string      `json:"format,omitempty"`
	Pattern     string      `json:"pattern,omitempty"`
	Minimum     *float64    `json:"minimum,omitempty"`
	Maximum     *float64    `json:"maximum,omitempty"`
	MinLength   *int        `json:"minLength,omitempty"`
	MaxLength   *int        `json:"maxLength,omitempty"`
	Default     interface{} `json:"default,omitempty"`
}

// ComponentRenderer defines how to render a specific component type
type ComponentRenderer interface {
	Render(ctx context.Context, props map[string]interface{}, schema *JsonSchema) (TemplComponent, error)
	GetSchemaType() string
	ValidateProps(props map[string]interface{}, schema *JsonSchema) error
}

// TemplComponent represents a rendered component ready for Templ
type TemplComponent struct {
	Type       string                 `json:"type"`
	Props      interface{}            `json:"props"`
	Children   []TemplComponent       `json:"children,omitempty"`
	Attributes map[string]string      `json:"attributes,omitempty"`
	Events     map[string]string      `json:"events,omitempty"`
	CSS        string                 `json:"css,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// ValidationEngine handles schema-based validation
type ValidationEngine struct {
	rules map[string]ValidationRule
}

// ValidationRule defines a validation rule
type ValidationRule struct {
	Type      string
	Validator func(value interface{}) error
	Message   string
	Severity  string
}

// NewSchemaFactory creates a new schema factory with the enhanced capabilities
func NewSchemaFactory(schemaDir string) (*SchemaFactory, error) {
	factory := &SchemaFactory{
		schemaRegistry:   make(map[string]*JsonSchema),
		componentFactory: make(map[string]ComponentRenderer),
		cssFactory:       css.NewFactory(schemaDir),
		validationEngine: NewValidationEngine(),
		templRenderer:    NewSchemaTemplRenderer(),
		schemaDir:        schemaDir,
	}

	// Register built-in component renderers
	factory.registerComponentRenderers()

	// Load JSON schema definitions
	if err := factory.loadJsonSchemas(); err != nil {
		return nil, fmt.Errorf("failed to load JSON schemas: %w", err)
	}

	return factory, nil
}

// loadJsonSchemas loads all JSON schema definitions from docs/ui/Schema/definitions/
func (f *SchemaFactory) loadJsonSchemas() error {
	definitionsDir := filepath.Join(f.schemaDir, "definitions")
	if _, err := os.Stat(definitionsDir); os.IsNotExist(err) {
		return fmt.Errorf("schema definitions directory not found: %s", definitionsDir)
	}

	pattern := filepath.Join(definitionsDir, "*.json")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to glob schema files: %w", err)
	}

	fmt.Printf("Loading %d JSON schema definitions from %s...\n", len(matches), definitionsDir)

	for _, file := range matches {
		if err := f.loadSingleSchema(file); err != nil {
			fmt.Printf("Warning: Failed to load schema %s: %v\n", filepath.Base(file), err)
			continue
		}
	}

	fmt.Printf("Successfully loaded %d JSON schemas\n", len(f.schemaRegistry))
	return nil
}

// loadSingleSchema loads a single JSON schema file
func (f *SchemaFactory) loadSingleSchema(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse into map first to preserve all data
	var rawSchema map[string]interface{}
	if err := json.Unmarshal(data, &rawSchema); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Parse into structured schema
	var schema JsonSchema
	if err := json.Unmarshal(data, &schema); err != nil {
		return fmt.Errorf("failed to parse structured JSON: %w", err)
	}

	schema.RawSchema = rawSchema

	// Extract schema name from filename
	basename := filepath.Base(filePath)
	schemaName := strings.TrimSuffix(basename, ".json")

	f.schemaRegistry[schemaName] = &schema
	return nil
}

// registerComponentRenderers registers built-in component renderers
func (f *SchemaFactory) registerComponentRenderers() {
	// Button components
	f.componentFactory["ButtonSchema"] = &ButtonRenderer{}
	f.componentFactory["ButtonGroupSchema"] = &ButtonGroupRenderer{}

	// Form components
	f.componentFactory["CheckboxControlSchema"] = &CheckboxRenderer{}
	f.componentFactory["RadioControlSchema"] = &RadioRenderer{}
	f.componentFactory["TextareaControlSchema"] = &TextareaRenderer{}
	f.componentFactory["InputControlSchema"] = &InputRenderer{}
	f.componentFactory["SelectControlSchema"] = &SelectRenderer{}

	// Layout components
	f.componentFactory["ContainerSchema"] = &ContainerRenderer{}
	f.componentFactory["CardSchema"] = &CardRenderer{}
	f.componentFactory["PanelSchema"] = &PanelRenderer{}

	// Data components
	f.componentFactory["TableSchema"] = &TableRenderer{}
	f.componentFactory["ListSchema"] = &ListRenderer{}
}

// RenderFromSchema creates a Templ component from a JSON schema definition
func (f *SchemaFactory) RenderFromSchema(ctx context.Context, schemaType string, props map[string]interface{}) (TemplComponent, error) {
	// Get schema definition
	schema, exists := f.schemaRegistry[schemaType]
	if !exists {
		return TemplComponent{}, fmt.Errorf("schema not found: %s", schemaType)
	}

	// Get component renderer
	renderer, exists := f.componentFactory[schemaType]
	if !exists {
		return TemplComponent{}, fmt.Errorf("renderer not found for schema: %s", schemaType)
	}

	// Validate props against schema
	if err := renderer.ValidateProps(props, schema); err != nil {
		return TemplComponent{}, fmt.Errorf("validation failed: %w", err)
	}

	// Render component
	component, err := renderer.Render(ctx, props, schema)
	if err != nil {
		return TemplComponent{}, fmt.Errorf("render failed: %w", err)
	}

	// Apply CSS if available
	if err := f.applyCSSToComponent(&component, props); err != nil {
		fmt.Printf("Warning: CSS application failed: %v\n", err)
	}

	return component, nil
}

// RenderToTempl creates a real Templ component from a JSON schema definition
func (f *SchemaFactory) RenderToTempl(ctx context.Context, schemaType string, props map[string]interface{}) (templ.Component, error) {
	// First create the TemplComponent
	templComponent, err := f.RenderFromSchema(ctx, schemaType, props)
	if err != nil {
		return nil, fmt.Errorf("failed to create TemplComponent: %w", err)
	}
	
	// Convert to actual Templ component
	return f.templRenderer.RenderComponent(templComponent), nil
}

// applyCSSToComponent applies CSS styling to a component
func (f *SchemaFactory) applyCSSToComponent(component *TemplComponent, props map[string]interface{}) error {
	// Extract CSS-related properties
	if className, ok := props["className"].(string); ok && className != "" {
		component.CSS = className
	}

	if style, ok := props["style"].(map[string]interface{}); ok {
		// Convert style object to CSS string using the CSS factory
		if styles, err := f.convertStyleMapToCSS(style); err == nil {
			if component.CSS != "" {
				component.CSS += " " + styles
			} else {
				component.CSS = styles
			}
		}
	}

	return nil
}

// convertStyleMapToCSS converts a style map to CSS string
func (f *SchemaFactory) convertStyleMapToCSS(styleMap map[string]interface{}) (string, error) {
	// Convert to css.Styles format and use the factory
	cssStyles := &css.Styles{}

	// Map common style properties
	if color, ok := styleMap["color"].(string); ok {
		cssStyles.Color = color
	}
	if backgroundColor, ok := styleMap["backgroundColor"].(string); ok {
		cssStyles.BackgroundColor = backgroundColor
	}
	if fontSize, ok := styleMap["fontSize"].(string); ok {
		cssStyles.FontSize = fontSize
	}
	// Add more style mappings as needed

	return f.cssFactory.GenerateCSS(*cssStyles)
}

// GetAvailableSchemas returns all available schema types
func (f *SchemaFactory) GetAvailableSchemas() []string {
	var schemas []string
	for name := range f.schemaRegistry {
		schemas = append(schemas, name)
	}
	return schemas
}

// GetSchemaDefinition returns the JSON schema definition for a type
func (f *SchemaFactory) GetSchemaDefinition(schemaType string) (*JsonSchema, error) {
	schema, exists := f.schemaRegistry[schemaType]
	if !exists {
		return nil, fmt.Errorf("schema not found: %s", schemaType)
	}
	return schema, nil
}

// ValidateAgainstSchema validates data against a JSON schema
func (f *SchemaFactory) ValidateAgainstSchema(schemaType string, data map[string]interface{}) error {
	schema, err := f.GetSchemaDefinition(schemaType)
	if err != nil {
		return err
	}

	return f.validationEngine.ValidateData(data, schema)
}

// ============================================================================
// COMPONENT RENDERERS
// ============================================================================

// ButtonRenderer renders button components from ButtonSchema
type ButtonRenderer struct{}

func (r *ButtonRenderer) GetSchemaType() string {
	return "ButtonSchema"
}

func (r *ButtonRenderer) ValidateProps(props map[string]interface{}, schema *JsonSchema) error {
	// Validate required props based on schema
	for _, required := range schema.Required {
		if _, exists := props[required]; !exists {
			return fmt.Errorf("required property missing: %s", required)
		}
	}
	return nil
}

func (r *ButtonRenderer) Render(ctx context.Context, props map[string]interface{}, schema *JsonSchema) (TemplComponent, error) {
	// Convert props to ButtonProps
	buttonProps := atoms.ButtonProps{
		Text:     getStringProp(props, "text", ""),
		Type:     getStringProp(props, "type", "button"),
		Variant:  atoms.ButtonVariant(getStringProp(props, "variant", "primary")),
		Size:     atoms.ButtonSize(getStringProp(props, "size", "md")),
		Disabled: getBoolProp(props, "disabled", false),
		Loading:  getBoolProp(props, "loading", false),
		ID:       getStringProp(props, "id", ""),
		Class:    getStringProp(props, "className", ""),
		OnClick:  getStringProp(props, "onClick", ""),
	}

	return TemplComponent{
		Type:  "Button",
		Props: buttonProps,
		Attributes: map[string]string{
			"type": buttonProps.Type,
			"id":   buttonProps.ID,
		},
		Events: map[string]string{
			"click": buttonProps.OnClick,
		},
	}, nil
}

// CheckboxRenderer renders checkbox components from CheckboxControlSchema
type CheckboxRenderer struct{}

func (r *CheckboxRenderer) GetSchemaType() string {
	return "CheckboxControlSchema"
}

func (r *CheckboxRenderer) ValidateProps(props map[string]interface{}, schema *JsonSchema) error {
	return nil // Basic validation, can be enhanced
}

func (r *CheckboxRenderer) Render(ctx context.Context, props map[string]interface{}, schema *JsonSchema) (TemplComponent, error) {
	checkboxProps := atoms.CheckboxProps{
		Label:    getStringProp(props, "label", ""),
		Checked:  getBoolProp(props, "checked", false),
		Value:    getStringProp(props, "value", ""),
		Size:     atoms.CheckboxSize(getStringProp(props, "size", "md")),
		Required: getBoolProp(props, "required", false),
		Disabled: getBoolProp(props, "disabled", false),
		ID:       getStringProp(props, "id", ""),
		Name:     getStringProp(props, "name", ""),
		Class:    getStringProp(props, "className", ""),
	}

	return TemplComponent{
		Type:  "Checkbox",
		Props: checkboxProps,
		Attributes: map[string]string{
			"name":     checkboxProps.Name,
			"id":       checkboxProps.ID,
			"required": fmt.Sprintf("%v", checkboxProps.Required),
		},
	}, nil
}

// Add placeholder renderers for other component types
type ButtonGroupRenderer struct{}

func (r *ButtonGroupRenderer) GetSchemaType() string { return "ButtonGroupSchema" }
func (r *ButtonGroupRenderer) ValidateProps(props map[string]interface{}, schema *JsonSchema) error {
	return nil
}

func (r *ButtonGroupRenderer) Render(ctx context.Context, props map[string]interface{}, schema *JsonSchema) (TemplComponent, error) {
	return TemplComponent{Type: "ButtonGroup", Props: props}, nil
}

type RadioRenderer struct{}

func (r *RadioRenderer) GetSchemaType() string { return "RadioControlSchema" }
func (r *RadioRenderer) ValidateProps(props map[string]interface{}, schema *JsonSchema) error {
	return nil
}

func (r *RadioRenderer) Render(ctx context.Context, props map[string]interface{}, schema *JsonSchema) (TemplComponent, error) {
	return TemplComponent{Type: "Radio", Props: props}, nil
}

type TextareaRenderer struct{}

func (r *TextareaRenderer) GetSchemaType() string { return "TextareaControlSchema" }
func (r *TextareaRenderer) ValidateProps(props map[string]interface{}, schema *JsonSchema) error {
	return nil
}

func (r *TextareaRenderer) Render(ctx context.Context, props map[string]interface{}, schema *JsonSchema) (TemplComponent, error) {
	return TemplComponent{Type: "Textarea", Props: props}, nil
}

type InputRenderer struct{}

func (r *InputRenderer) GetSchemaType() string { return "InputControlSchema" }
func (r *InputRenderer) ValidateProps(props map[string]interface{}, schema *JsonSchema) error {
	return nil
}

func (r *InputRenderer) Render(ctx context.Context, props map[string]interface{}, schema *JsonSchema) (TemplComponent, error) {
	return TemplComponent{Type: "Input", Props: props}, nil
}

type SelectRenderer struct{}

func (r *SelectRenderer) GetSchemaType() string { return "SelectControlSchema" }
func (r *SelectRenderer) ValidateProps(props map[string]interface{}, schema *JsonSchema) error {
	return nil
}

func (r *SelectRenderer) Render(ctx context.Context, props map[string]interface{}, schema *JsonSchema) (TemplComponent, error) {
	return TemplComponent{Type: "Select", Props: props}, nil
}

type ContainerRenderer struct{}

func (r *ContainerRenderer) GetSchemaType() string { return "ContainerSchema" }
func (r *ContainerRenderer) ValidateProps(props map[string]interface{}, schema *JsonSchema) error {
	return nil
}

func (r *ContainerRenderer) Render(ctx context.Context, props map[string]interface{}, schema *JsonSchema) (TemplComponent, error) {
	return TemplComponent{Type: "Container", Props: props}, nil
}

type CardRenderer struct{}

func (r *CardRenderer) GetSchemaType() string { return "CardSchema" }
func (r *CardRenderer) ValidateProps(props map[string]interface{}, schema *JsonSchema) error {
	return nil
}

func (r *CardRenderer) Render(ctx context.Context, props map[string]interface{}, schema *JsonSchema) (TemplComponent, error) {
	return TemplComponent{Type: "Card", Props: props}, nil
}

type PanelRenderer struct{}

func (r *PanelRenderer) GetSchemaType() string { return "PanelSchema" }
func (r *PanelRenderer) ValidateProps(props map[string]interface{}, schema *JsonSchema) error {
	return nil
}

func (r *PanelRenderer) Render(ctx context.Context, props map[string]interface{}, schema *JsonSchema) (TemplComponent, error) {
	return TemplComponent{Type: "Panel", Props: props}, nil
}

type TableRenderer struct{}

func (r *TableRenderer) GetSchemaType() string { return "TableSchema" }
func (r *TableRenderer) ValidateProps(props map[string]interface{}, schema *JsonSchema) error {
	return nil
}

func (r *TableRenderer) Render(ctx context.Context, props map[string]interface{}, schema *JsonSchema) (TemplComponent, error) {
	return TemplComponent{Type: "Table", Props: props}, nil
}

type ListRenderer struct{}

func (r *ListRenderer) GetSchemaType() string { return "ListSchema" }
func (r *ListRenderer) ValidateProps(props map[string]interface{}, schema *JsonSchema) error {
	return nil
}

func (r *ListRenderer) Render(ctx context.Context, props map[string]interface{}, schema *JsonSchema) (TemplComponent, error) {
	return TemplComponent{Type: "List", Props: props}, nil
}

// ============================================================================
// VALIDATION ENGINE
// ============================================================================

// NewValidationEngine creates a new validation engine
func NewValidationEngine() *ValidationEngine {
	return &ValidationEngine{
		rules: make(map[string]ValidationRule),
	}
}

// ValidateData validates data against a schema
func (ve *ValidationEngine) ValidateData(data map[string]interface{}, schema *JsonSchema) error {
	// Validate required fields
	for _, required := range schema.Required {
		if _, exists := data[required]; !exists {
			return fmt.Errorf("required field missing: %s", required)
		}
	}

	// Validate each property
	for key, value := range data {
		if property, exists := schema.Properties[key]; exists {
			if err := ve.validateProperty(key, value, property); err != nil {
				return fmt.Errorf("property %s: %w", key, err)
			}
		}
	}

	return nil
}

// validateProperty validates a single property
func (ve *ValidationEngine) validateProperty(key string, value interface{}, property Property) error {
	// Type validation
	if property.Type != nil {
		if err := ve.validateType(value, property.Type); err != nil {
			return fmt.Errorf("type validation failed: %w", err)
		}
	}

	// Enum validation
	if len(property.Enum) > 0 {
		if err := ve.validateEnum(value, property.Enum); err != nil {
			return fmt.Errorf("enum validation failed: %w", err)
		}
	}

	// Pattern validation
	if property.Pattern != "" {
		if err := ve.validatePattern(value, property.Pattern); err != nil {
			return fmt.Errorf("pattern validation failed: %w", err)
		}
	}

	return nil
}

// validateType validates the type of a value
func (ve *ValidationEngine) validateType(value interface{}, expectedType interface{}) error {
	switch expectedType {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean, got %T", value)
		}
	case "number":
		switch value.(type) {
		case int, int64, float32, float64:
			// Valid number types
		default:
			return fmt.Errorf("expected number, got %T", value)
		}
	}
	return nil
}

// validateEnum validates that a value is in the allowed enum values
func (ve *ValidationEngine) validateEnum(value interface{}, allowedValues []string) error {
	valueStr := fmt.Sprintf("%v", value)
	for _, allowed := range allowedValues {
		if valueStr == allowed {
			return nil
		}
	}
	return fmt.Errorf("value %v not in allowed values: %v", value, allowedValues)
}

// validatePattern validates a string against a regex pattern
func (ve *ValidationEngine) validatePattern(value interface{}, pattern string) error {
	// For now, just return nil - can be enhanced with actual regex validation
	return nil
}

// ============================================================================
// UTILITY FUNCTIONS
// ============================================================================

// Helper functions to extract typed values from props map (defined in registry.go)

func getArrayProp(props map[string]interface{}, key string) []interface{} {
	if value, exists := props[key]; exists {
		if arr, ok := value.([]interface{}); ok {
			return arr
		}
	}
	return nil
}

