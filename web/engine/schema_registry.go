package engine

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/a-h/templ"
	"github.com/niiniyare/erp/web/components/atoms"
	"github.com/niiniyare/erp/web/schemas"
)

// SchemaComponentRegistry extends ComponentRegistry with schema-driven capabilities
type SchemaComponentRegistry struct {
	*ComponentRegistry                      // Embed existing registry for backward compatibility
	schemaFactories map[string]SchemaFactory // Schema-driven component factories
	tagParsers      map[string]TagParser     // Tag parsing functions for different component types
	generators      map[string]Generator     // Code generators for different output formats
	mu              sync.RWMutex             // Protects schema-specific maps
}

// SchemaFactory creates components from schema definitions with full tag support
type SchemaFactory func(schemaJSON string, ctx context.Context) (templ.Component, error)

// TagParser converts struct tags to component schemas
type TagParser func(fieldName string, fieldType reflect.Type, tags map[string]string) (schemas.ComponentDefinition, error)

// Generator creates different output formats from schemas
type Generator func(schema schemas.ComponentDefinition) (string, error)

// NewSchemaComponentRegistry creates a registry with both traditional and schema support
func NewSchemaComponentRegistry() *SchemaComponentRegistry {
	registry := &SchemaComponentRegistry{
		ComponentRegistry: NewComponentRegistry(), // Initialize base registry
		schemaFactories:   make(map[string]SchemaFactory),
		tagParsers:        make(map[string]TagParser),
		generators:        make(map[string]Generator),
	}

	// Register schema-driven factories
	registry.registerSchemaFactories()
	
	// Register tag parsers for auto-generation
	registry.registerTagParsers()
	
	// Register code generators
	registry.registerGenerators()

	return registry
}

// RegisterSchemaFactory adds a schema-driven component factory
func (r *SchemaComponentRegistry) RegisterSchemaFactory(componentType string, factory SchemaFactory) error {
	if componentType == "" {
		return fmt.Errorf("component type cannot be empty")
	}
	if factory == nil {
		return fmt.Errorf("schema factory cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.schemaFactories[componentType] = factory
	return nil
}

// RegisterTagParser adds a tag parser for a component type
func (r *SchemaComponentRegistry) RegisterTagParser(componentType string, parser TagParser) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tagParsers[componentType] = parser
	return nil
}

// RegisterGenerator adds a code generator for a specific format
func (r *SchemaComponentRegistry) RegisterGenerator(format string, generator Generator) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.generators[format] = generator
	return nil
}

// ResolveFromSchema creates a component from JSON schema string
func (r *SchemaComponentRegistry) ResolveFromSchema(componentType string, schemaJSON string, ctx context.Context) (templ.Component, error) {
	r.mu.RLock()
	factory, exists := r.schemaFactories[componentType]
	r.mu.RUnlock()

	if !exists {
		// Fallback to traditional factory if schema factory doesn't exist
		return r.ComponentRegistry.Resolve(componentType)
	}

	return factory(schemaJSON, ctx)
}

// ResolveFromTags creates a component from Go struct tags (for auto-generation)
func (r *SchemaComponentRegistry) ResolveFromTags(componentType string, fieldName string, fieldType reflect.Type, tags map[string]string, ctx context.Context) (templ.Component, error) {
	r.mu.RLock()
	parser, exists := r.tagParsers[componentType]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no tag parser found for component type: %s", componentType)
	}

	// Parse tags into schema
	schema, err := parser(fieldName, fieldType, tags)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tags: %w", err)
	}

	// Convert schema to JSON and resolve
	schemaJSON, err := schema.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize schema: %w", err)
	}

	return r.ResolveFromSchema(componentType, schemaJSON, ctx)
}

// GenerateCode creates code in the specified format from a schema
func (r *SchemaComponentRegistry) GenerateCode(format string, schema schemas.ComponentDefinition) (string, error) {
	r.mu.RLock()
	generator, exists := r.generators[format]
	r.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("no generator found for format: %s", format)
	}

	return generator(schema)
}

// AnalyzeStruct extracts component schemas from a Go struct using reflection
func (r *SchemaComponentRegistry) AnalyzeStruct(structType reflect.Type) ([]schemas.ComponentDefinition, error) {
	if structType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct type, got %s", structType.Kind())
	}

	var components []schemas.ComponentDefinition

	// Iterate through struct fields
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		
		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Parse struct tags
		tags := r.parseStructTags(field)
		
		// Determine component type from ui: tag or field type
		componentType := r.determineComponentType(field, tags)
		if componentType == "" {
			continue // Skip fields without UI components
		}

		// Parse tags using appropriate parser
		if parser, exists := r.tagParsers[componentType]; exists {
			schema, err := parser(field.Name, field.Type, tags)
			if err != nil {
				continue // Skip fields with parsing errors
			}
			components = append(components, schema)
		}
	}

	return components, nil
}

// Hybrid support methods for gradual migration

// ResolveHybrid attempts schema resolution first, falls back to traditional
func (r *SchemaComponentRegistry) ResolveHybrid(componentType string, props map[string]any, ctx context.Context) (templ.Component, error) {
	// Try schema-driven approach first
	if schemaFactory, exists := r.schemaFactories[componentType]; exists {
		// Convert props to schema JSON
		schema := schemas.ComponentDefinition{
			Type:  componentType,
			Props: props,
		}
		
		schemaJSON, err := schema.ToJSON()
		if err == nil {
			if component, err := schemaFactory(schemaJSON, ctx); err == nil {
				return component, nil
			}
		}
	}

	// Fallback to traditional ComponentRegistry
	return r.ComponentRegistry.ResolveWithDefinition(&schemas.ComponentDefinition{
		Type:  componentType,
		Props: props,
	}, ctx)
}

// Internal registration methods

func (r *SchemaComponentRegistry) registerSchemaFactories() {
	// Register button schema factory
	r.RegisterSchemaFactory("atoms.button", func(schemaJSON string, ctx context.Context) (templ.Component, error) {
		schema, err := atoms.ButtonSchemaFromJSON(schemaJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to parse button schema: %w", err)
		}
		
		// Validate schema
		if err := schema.Validate(); err != nil {
			return nil, fmt.Errorf("invalid button schema: %w", err)
		}
		
		return atoms.ButtonFromSchema(schema), nil
	})

	// Register variants for different button types
	r.RegisterSchemaFactory("atoms.icon-button", func(schemaJSON string, ctx context.Context) (templ.Component, error) {
		schema, err := atoms.ButtonSchemaFromJSON(schemaJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to parse icon button schema: %w", err)
		}
		
		// Override component type for icon button
		schema.UITags.Component = "icon-button"
		return atoms.ButtonFromSchema(schema), nil
	})

	r.RegisterSchemaFactory("atoms.link-button", func(schemaJSON string, ctx context.Context) (templ.Component, error) {
		schema, err := atoms.ButtonSchemaFromJSON(schemaJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to parse link button schema: %w", err)
		}
		
		// Override component type for link button
		schema.UITags.Component = "link-button"
		return atoms.ButtonFromSchema(schema), nil
	})

	// TODO: Add more schema factories for other components
	// r.RegisterSchemaFactory("atoms.input", inputSchemaFactory)
	// r.RegisterSchemaFactory("molecules.field", fieldSchemaFactory)
}

func (r *SchemaComponentRegistry) registerTagParsers() {
	// Register button tag parser
	r.RegisterTagParser("atoms.button", func(fieldName string, fieldType reflect.Type, tags map[string]string) (schemas.ComponentDefinition, error) {
		buttonSchema := atoms.ButtonSchemaFromTags(fieldName, tags)
		
		// Convert to generic ComponentDefinition
		definition := schemas.ComponentDefinition{
			Type:     buttonSchema.Type,
			ID:       buttonSchema.ID,
			Class:    buttonSchema.Class,
			Disabled: buttonSchema.Disabled,
			Props:    buttonSchema.Props,
		}
		
		// Add component-specific data
		definition.Props["uiTags"] = buttonSchema.UITags
		definition.Props["formTags"] = buttonSchema.FormTags
		definition.Props["tableTags"] = buttonSchema.TableTags
		
		return definition, nil
	})

	// Register input tag parser
	r.RegisterTagParser("atoms.input", func(fieldName string, fieldType reflect.Type, tags map[string]string) (schemas.ComponentDefinition, error) {
		// Basic input parsing - extend based on your input schema implementation
		definition := schemas.ComponentDefinition{
			Type:  "atoms.input",
			Props: make(map[string]any),
		}
		
		// Parse ui: tag for inputs
		if uiTag, exists := tags["ui"]; exists {
			pairs := strings.Split(uiTag, ";")
			for _, pair := range pairs {
				if kv := strings.SplitN(pair, "=", 2); len(kv) == 2 {
					key := strings.TrimSpace(kv[0])
					value := strings.TrimSpace(kv[1])
					
					switch key {
					case "component":
						if value != "input" && value != "email" && value != "password" && value != "number" {
							continue // Not an input component
						}
						definition.Props["inputType"] = value
					case "label":
						definition.Props["label"] = value
					case "placeholder":
						definition.Props["placeholder"] = value
					case "required":
						definition.Props["required"] = value == "true"
					}
				}
			}
		}
		
		// Set default label if not specified
		if _, exists := definition.Props["label"]; !exists {
			definition.Props["label"] = humanizeFieldName(fieldName)
		}
		
		return definition, nil
	})

	// TODO: Add more tag parsers
	// r.RegisterTagParser("atoms.select", selectTagParser)
	// r.RegisterTagParser("molecules.field", fieldTagParser)
}

func (r *SchemaComponentRegistry) registerGenerators() {
	// JSON generator
	r.RegisterGenerator("json", func(schema schemas.ComponentDefinition) (string, error) {
		return schema.ToJSON()
	})

	// Go code generator
	r.RegisterGenerator("go", func(schema schemas.ComponentDefinition) (string, error) {
		// Generate Go code for the component
		var code strings.Builder
		
		code.WriteString(fmt.Sprintf("// Auto-generated component: %s\n", schema.Type))
		code.WriteString(fmt.Sprintf("templ %sComponent() {\n", strings.ReplaceAll(schema.Type, ".", "")))
		code.WriteString(fmt.Sprintf("\t@%s(%sProps{\n", strings.Title(strings.Split(schema.Type, ".")[1]), strings.Title(strings.Split(schema.Type, ".")[1])))
		
		// Add props
		for key, value := range schema.Props {
			if str, ok := value.(string); ok {
				code.WriteString(fmt.Sprintf("\t\t%s: \"%s\",\n", strings.Title(key), str))
			} else if b, ok := value.(bool); ok {
				code.WriteString(fmt.Sprintf("\t\t%s: %t,\n", strings.Title(key), b))
			}
		}
		
		code.WriteString("\t})\n")
		code.WriteString("}\n")
		
		return code.String(), nil
	})

	// YAML generator
	r.RegisterGenerator("yaml", func(schema schemas.ComponentDefinition) (string, error) {
		var yaml strings.Builder
		
		yaml.WriteString(fmt.Sprintf("type: %s\n", schema.Type))
		if schema.ID != "" {
			yaml.WriteString(fmt.Sprintf("id: %s\n", schema.ID))
		}
		if schema.Class != "" {
			yaml.WriteString(fmt.Sprintf("class: %s\n", schema.Class))
		}
		yaml.WriteString("props:\n")
		
		for key, value := range schema.Props {
			yaml.WriteString(fmt.Sprintf("  %s: %v\n", key, value))
		}
		
		return yaml.String(), nil
	})
}

// Helper methods

func (r *SchemaComponentRegistry) parseStructTags(field reflect.StructField) map[string]string {
	tags := make(map[string]string)
	
	// Common struct tags to parse
	tagNames := []string{"ui", "form", "table", "validate", "db", "json"}
	
	for _, tagName := range tagNames {
		if tagValue := field.Tag.Get(tagName); tagValue != "" {
			tags[tagName] = tagValue
		}
	}
	
	return tags
}

func (r *SchemaComponentRegistry) determineComponentType(field reflect.StructField, tags map[string]string) string {
	// Check ui: tag first
	if uiTag, exists := tags["ui"]; exists {
		pairs := strings.Split(uiTag, ";")
		for _, pair := range pairs {
			if kv := strings.SplitN(pair, "=", 2); len(kv) == 2 {
				key := strings.TrimSpace(kv[0])
				value := strings.TrimSpace(kv[1])
				
				if key == "component" {
					// Map component types to full component names
					switch value {
					case "button", "submit":
						return "atoms.button"
					case "icon-button":
						return "atoms.icon-button"
					case "link-button":
						return "atoms.link-button"
					case "text", "email", "password", "number":
						return "atoms.input"
					case "textarea":
						return "atoms.textarea"
					case "select":
						return "atoms.select"
					case "checkbox":
						return "atoms.checkbox"
					case "radio":
						return "atoms.radio"
					case "toggle":
						return "atoms.toggle"
					}
				}
			}
		}
	}
	
	// Fallback to field type inference
	switch field.Type.Kind() {
	case reflect.String:
		return "atoms.input"
	case reflect.Bool:
		return "atoms.checkbox"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		 reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		 reflect.Float32, reflect.Float64:
		return "atoms.input"
	}
	
	// Check for common field names that indicate buttons
	fieldNameLower := strings.ToLower(field.Name)
	if strings.Contains(fieldNameLower, "button") || 
	   strings.Contains(fieldNameLower, "action") || 
	   strings.Contains(fieldNameLower, "submit") {
		return "atoms.button"
	}
	
	return "" // No component type determined
}

// humanizeFieldName converts camelCase/PascalCase to human readable form
func humanizeFieldName(fieldName string) string {
	if fieldName == "" {
		return ""
	}
	
	var result []rune
	for i, r := range fieldName {
		if i > 0 && 'A' <= r && r <= 'Z' {
			result = append(result, ' ')
		}
		result = append(result, r)
	}
	
	return string(result)
}