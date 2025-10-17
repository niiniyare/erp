package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/a-h/templ"
	"github.com/niiniyare/erp/web/schemas"
)

// EnhancedComponentRegistry extends the basic ComponentRegistry with schema-driven capabilities
// It provides seamless integration between JSON schemas and Templ components
type EnhancedComponentRegistry struct {
	*ComponentRegistry // Embed existing registry
	schemaRenderer     *JSONSchemaRenderer
	schemaFactory      *SchemaFactory
	mu                 sync.RWMutex
	schemaBindings     map[string]string // schemaType -> componentType mapping
}

// NewEnhancedComponentRegistry creates a registry with schema-driven capabilities
func NewEnhancedComponentRegistry(schemaDir string) (*EnhancedComponentRegistry, error) {
	baseRegistry := NewComponentRegistry()
	
	schemaRenderer, err := NewJSONSchemaRenderer(schemaDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create schema renderer: %w", err)
	}

	schemaFactory, err := NewSchemaFactory(schemaDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create schema factory: %w", err)
	}

	registry := &EnhancedComponentRegistry{
		ComponentRegistry: baseRegistry,
		schemaRenderer:    schemaRenderer,
		schemaFactory:     schemaFactory,
		schemaBindings:    make(map[string]string),
	}

	// Register schema bindings
	registry.registerSchemaBindings()

	return registry, nil
}

// registerSchemaBindings creates mappings between schema types and component types
func (er *EnhancedComponentRegistry) registerSchemaBindings() {
	// Form component bindings
	er.schemaBindings["ButtonSchema"] = "button"
	er.schemaBindings["ButtonGroupSchema"] = "button-group"
	er.schemaBindings["CheckboxControlSchema"] = "checkbox"
	er.schemaBindings["RadioControlSchema"] = "radio"
	er.schemaBindings["InputControlSchema"] = "input"
	er.schemaBindings["TextareaControlSchema"] = "textarea"
	er.schemaBindings["SelectControlSchema"] = "select"

	// Layout component bindings
	er.schemaBindings["ContainerSchema"] = "container"
	er.schemaBindings["CardSchema"] = "card"
	er.schemaBindings["PanelSchema"] = "panel"
	er.schemaBindings["TabsSchema"] = "tabs"
	er.schemaBindings["ModalSchema"] = "modal"

	// Data display bindings
	er.schemaBindings["TableSchema"] = "table"
	er.schemaBindings["ListSchema"] = "list"
	er.schemaBindings["TreeSchema"] = "tree"
	er.schemaBindings["ChartSchema"] = "chart"

	// Navigation bindings
	er.schemaBindings["NavSchema"] = "nav"
	er.schemaBindings["BreadcrumbSchema"] = "breadcrumb"
	er.schemaBindings["PaginationSchema"] = "pagination"
}

// CreateFromSchema creates a component instance from a JSON schema definition
func (er *EnhancedComponentRegistry) CreateFromSchema(ctx context.Context, schemaType string, props map[string]interface{}) (schemas.ComponentDefinition, error) {
	er.mu.RLock()
	defer er.mu.RUnlock()

	// First try to render using the schema factory
	templComponent, err := er.schemaFactory.RenderFromSchema(ctx, schemaType, props)
	if err != nil {
		return schemas.ComponentDefinition{}, fmt.Errorf("schema factory failed: %w", err)
	}

	// Convert TemplComponent to ComponentDefinition
	componentDef := schemas.ComponentDefinition{
		Type:  templComponent.Type,
		Class: templComponent.CSS,
	}
	
	// Handle Props conversion - it's an interface{} that we need to convert
	if propsMap, ok := templComponent.Props.(map[string]interface{}); ok {
		componentDef.Props = propsMap
	} else {
		// If it's not a map, create an empty map
		componentDef.Props = make(map[string]interface{})
	}

	// Add component type mapping if available
	if componentType, exists := er.schemaBindings[schemaType]; exists {
		componentDef.Type = componentType
	}

	return componentDef, nil
}

// RenderFromSchema renders a component directly from JSON schema
func (er *EnhancedComponentRegistry) RenderFromSchema(ctx context.Context, schemaType string, props map[string]interface{}) (string, error) {
	return er.schemaRenderer.RenderComponent(ctx, schemaType, props)
}

// GetSchemaDefinition returns the JSON schema definition for a component type
func (er *EnhancedComponentRegistry) GetSchemaDefinition(schemaType string) (*JsonSchema, error) {
	return er.schemaFactory.GetSchemaDefinition(schemaType)
}

// ValidateComponentAgainstSchema validates component props against its JSON schema
func (er *EnhancedComponentRegistry) ValidateComponentAgainstSchema(schemaType string, props map[string]interface{}) error {
	return er.schemaFactory.ValidateAgainstSchema(schemaType, props)
}

// GetAvailableSchemas returns all available schema types
func (er *EnhancedComponentRegistry) GetAvailableSchemas() []string {
	return er.schemaFactory.GetAvailableSchemas()
}

// GetSchemaBinding returns the component type for a schema type
func (er *EnhancedComponentRegistry) GetSchemaBinding(schemaType string) (string, bool) {
	er.mu.RLock()
	defer er.mu.RUnlock()
	
	componentType, exists := er.schemaBindings[schemaType]
	return componentType, exists
}

// RegisterSchemaBinding adds a new schema-to-component binding
func (er *EnhancedComponentRegistry) RegisterSchemaBinding(schemaType, componentType string) {
	er.mu.Lock()
	defer er.mu.Unlock()
	
	er.schemaBindings[schemaType] = componentType
}

// CreateComponentFromJSON creates a component from JSON string
func (er *EnhancedComponentRegistry) CreateComponentFromJSON(ctx context.Context, jsonData string) (schemas.ComponentDefinition, error) {
	// Parse the JSON to extract schema type and props
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return schemas.ComponentDefinition{}, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Extract schema type
	schemaType, ok := data["type"].(string)
	if !ok {
		return schemas.ComponentDefinition{}, fmt.Errorf("schema type not found in JSON")
	}

	// Append "Schema" suffix if not present
	if !strings.HasSuffix(schemaType, "Schema") {
		schemaType += "Schema"
	}

	// Remove type from props
	props := make(map[string]interface{})
	for key, value := range data {
		if key != "type" {
			props[key] = value
		}
	}

	return er.CreateFromSchema(ctx, schemaType, props)
}

// RenderComponentTree renders a tree of components from schema definitions
func (er *EnhancedComponentRegistry) RenderComponentTree(ctx context.Context, components []ComponentDefinition) (string, error) {
	return er.schemaRenderer.RenderComponentTree(ctx, components)
}

// ============================================================================
// SCHEMA-DRIVEN COMPONENT FACTORY
// ============================================================================

// SchemaComponentFactory creates components using JSON schema validation
type SchemaComponentFactory struct {
	schemaType    string
	schemaFactory *SchemaFactory
	componentType string
}

// NewSchemaComponentFactory creates a new schema-driven component factory
func NewSchemaComponentFactory(schemaType, componentType string, schemaFactory *SchemaFactory) *SchemaComponentFactory {
	return &SchemaComponentFactory{
		schemaType:    schemaType,
		schemaFactory: schemaFactory,
		componentType: componentType,
	}
}

// Create creates a component instance with schema validation
func (scf *SchemaComponentFactory) Create(ctx context.Context, props map[string]interface{}) (schemas.ComponentDefinition, error) {
	// Validate props against schema
	if err := scf.schemaFactory.ValidateAgainstSchema(scf.schemaType, props); err != nil {
		return schemas.ComponentDefinition{}, fmt.Errorf("schema validation failed: %w", err)
	}

	// Create component using schema factory
	templComponent, err := scf.schemaFactory.RenderFromSchema(ctx, scf.schemaType, props)
	if err != nil {
		return schemas.ComponentDefinition{}, fmt.Errorf("component creation failed: %w", err)
	}

	// Convert to ComponentDefinition
	componentDef := schemas.ComponentDefinition{
		Type:  scf.componentType,
		Class: templComponent.CSS,
	}
	
	// Handle Props conversion
	if propsMap, ok := templComponent.Props.(map[string]interface{}); ok {
		componentDef.Props = propsMap
	} else {
		componentDef.Props = make(map[string]interface{})
	}
	
	return componentDef, nil
}

// Validate validates a component definition against the schema
func (scf *SchemaComponentFactory) Validate(ctx context.Context, component schemas.ComponentDefinition) error {
	// This is a simplified approach - in reality, you'd need to convert
	// the component props back to a map format for validation
	// For now, we'll assume basic validation
	
	if component.Type != scf.componentType {
		return fmt.Errorf("component type mismatch: expected %s, got %s", scf.componentType, component.Type)
	}

	return nil
}

// GetSchema returns the JSON schema for this component type
func (scf *SchemaComponentFactory) GetSchema() (*JsonSchema, error) {
	return scf.schemaFactory.GetSchemaDefinition(scf.schemaType)
}

// ============================================================================
// ENHANCED REGISTRY UTILITIES
// ============================================================================

// RegisterSchemaComponentFactory registers a schema-driven component factory
func (er *EnhancedComponentRegistry) RegisterSchemaComponentFactory(schemaType, componentType string) error {
	factory := NewSchemaComponentFactory(schemaType, componentType, er.schemaFactory)
	
	// Register with the base registry using the adapter
	adapter := er.createSchemaFactoryAdapter(factory)
	return er.ComponentRegistry.Register(componentType, adapter)
}

// SchemaFactoryAdapter adapts SchemaComponentFactory to schemas.ComponentFactory function type
func (er *EnhancedComponentRegistry) createSchemaFactoryAdapter(factory *SchemaComponentFactory) schemas.ComponentFactory {
	return func(def *schemas.ComponentDefinition, ctx context.Context) (templ.Component, error) {
		// Convert ComponentDefinition to our internal format and create the component
		props := def.Props
		if props == nil {
			props = make(map[string]interface{})
		}

		// Instead of using the factory.Create which returns ComponentDefinition,
		// directly use the schema factory to get a TemplComponent and render it
		schemaType := factory.schemaType
		templComponent, err := factory.schemaFactory.RenderFromSchema(ctx, schemaType, props)
		if err != nil {
			return nil, fmt.Errorf("failed to create component: %w", err)
		}

		// Use the schema factory's Templ renderer to convert to real Templ component
		templRenderer := NewSchemaTemplRenderer()
		return templRenderer.RenderComponent(templComponent), nil
	}
}

// RenderToTempl creates a real Templ component from schema type and props
func (er *EnhancedComponentRegistry) RenderToTempl(ctx context.Context, schemaType string, props map[string]interface{}) (templ.Component, error) {
	return er.schemaFactory.RenderToTempl(ctx, schemaType, props)
}

// GetComponentInfo returns information about all registered components
func (er *EnhancedComponentRegistry) GetComponentInfo() map[string]ComponentInfo {
	er.mu.RLock()
	defer er.mu.RUnlock()

	info := make(map[string]ComponentInfo)

	// Get info from base registry
	for componentType := range er.ComponentRegistry.factories {
		info[componentType] = ComponentInfo{
			Type:        componentType,
			HasSchema:   false,
			SchemaType:  "",
			Description: fmt.Sprintf("Component: %s", componentType),
		}
	}

	// Add schema information
	for schemaType, componentType := range er.schemaBindings {
		if existing, exists := info[componentType]; exists {
			existing.HasSchema = true
			existing.SchemaType = schemaType
			if schema, err := er.schemaFactory.GetSchemaDefinition(schemaType); err == nil {
				existing.Description = schema.Description
			}
			info[componentType] = existing
		} else {
			// Create new entry for schema-only components
			description := fmt.Sprintf("Schema-driven component: %s", schemaType)
			if schema, err := er.schemaFactory.GetSchemaDefinition(schemaType); err == nil {
				description = schema.Description
			}
			
			info[componentType] = ComponentInfo{
				Type:        componentType,
				HasSchema:   true,
				SchemaType:  schemaType,
				Description: description,
			}
		}
	}

	return info
}

// ComponentInfo provides information about a registered component
type ComponentInfo struct {
	Type        string `json:"type"`
	HasSchema   bool   `json:"hasSchema"`
	SchemaType  string `json:"schemaType,omitempty"`
	Description string `json:"description"`
}