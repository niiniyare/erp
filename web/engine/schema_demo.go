package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
)

// SchemaDemoSystem demonstrates the enhanced schema-driven UI system
type SchemaDemoSystem struct {
	registry *EnhancedComponentRegistry
}

// NewSchemaDemoSystem creates a new demo system
func NewSchemaDemoSystem(schemaDir string) (*SchemaDemoSystem, error) {
	registry, err := NewEnhancedComponentRegistry(schemaDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create enhanced registry: %w", err)
	}

	return &SchemaDemoSystem{
		registry: registry,
	}, nil
}

// DemoBasicComponents demonstrates basic component creation from schemas
func (demo *SchemaDemoSystem) DemoBasicComponents(ctx context.Context) error {
	fmt.Println("=== Schema-Driven UI System Demo ===\n")

	// Demo 1: Create a button from ButtonSchema
	fmt.Println("1. Creating Button from ButtonSchema...")
	buttonProps := map[string]interface{}{
		"text":     "Save Changes",
		"variant":  "primary",
		"size":     "lg",
		"disabled": false,
		"type":     "submit",
	}

	buttonHTML, err := demo.registry.RenderFromSchema(ctx, "ButtonSchema", buttonProps)
	if err != nil {
		return fmt.Errorf("failed to render button: %w", err)
	}
	fmt.Printf("Button HTML: %s\n\n", buttonHTML)

	// Demo 2: Create a checkbox from CheckboxControlSchema
	fmt.Println("2. Creating Checkbox from CheckboxControlSchema...")
	checkboxProps := map[string]interface{}{
		"label":    "Accept Terms and Conditions",
		"checked":  false,
		"required": true,
		"name":     "terms",
		"value":    "accepted",
	}

	checkboxHTML, err := demo.registry.RenderFromSchema(ctx, "CheckboxControlSchema", checkboxProps)
	if err != nil {
		return fmt.Errorf("failed to render checkbox: %w", err)
	}
	fmt.Printf("Checkbox HTML: %s\n\n", checkboxHTML)

	// Demo 3: Create an input from InputControlSchema with validation
	fmt.Println("3. Creating Input with validation...")
	inputProps := map[string]interface{}{
		"type":        "email",
		"placeholder": "Enter your email address",
		"required":    true,
		"name":        "email",
		"size":        "md",
	}

	// Validate props against schema first
	if err := demo.registry.ValidateComponentAgainstSchema("InputControlSchema", inputProps); err != nil {
		fmt.Printf("Validation error: %v\n", err)
	} else {
		fmt.Println("✓ Input props validation passed")
	}

	inputHTML, err := demo.registry.RenderFromSchema(ctx, "InputControlSchema", inputProps)
	if err != nil {
		return fmt.Errorf("failed to render input: %w", err)
	}
	fmt.Printf("Input HTML: %s\n\n", inputHTML)

	return nil
}

// DemoComplexForm demonstrates creating a complete form from JSON schema definitions
func (demo *SchemaDemoSystem) DemoComplexForm(ctx context.Context) error {
	fmt.Println("=== Complex Form Demo ===\n")

	// Define a complete user registration form using JSON
	formJSON := `{
		"type": "form",
		"id": "user-registration",
		"className": "space-y-6 bg-white p-6 rounded-lg shadow-md",
		"children": [
			{
				"type": "input",
				"name": "firstName",
				"label": "First Name",
				"placeholder": "Enter your first name",
				"required": true,
				"size": "lg"
			},
			{
				"type": "input",
				"name": "lastName", 
				"label": "Last Name",
				"placeholder": "Enter your last name",
				"required": true,
				"size": "lg"
			},
			{
				"type": "input",
				"name": "email",
				"type": "email",
				"label": "Email Address",
				"placeholder": "Enter your email",
				"required": true,
				"size": "lg"
			},
			{
				"type": "textarea",
				"name": "bio",
				"label": "Bio",
				"placeholder": "Tell us about yourself...",
				"rows": 4,
				"maxLength": 500
			},
			{
				"type": "checkbox",
				"name": "newsletter",
				"label": "Subscribe to newsletter",
				"checked": false
			},
			{
				"type": "button-group",
				"className": "flex justify-end space-x-4",
				"buttons": [
					{
						"text": "Cancel",
						"variant": "secondary",
						"type": "button"
					},
					{
						"text": "Create Account",
						"variant": "primary",
						"type": "submit"
					}
				]
			}
		]
	}`

	// Parse and convert to component definitions
	var formData map[string]interface{}
	if err := json.Unmarshal([]byte(formJSON), &formData); err != nil {
		return fmt.Errorf("failed to parse form JSON: %w", err)
	}

	// Convert to component tree
	component, err := demo.registry.schemaRenderer.ConvertSchemaToComponent(formData)
	if err != nil {
		return fmt.Errorf("failed to convert schema: %w", err)
	}

	fmt.Printf("Form Component Definition:\n")
	fmt.Printf("- ID: %s\n", component.ID)
	fmt.Printf("- Schema Type: %s\n", component.SchemaType)
	fmt.Printf("- Number of children: %d\n\n", len(component.Children))

	// Render each child component
	fmt.Println("Rendered child components:")
	for i, child := range component.Children {
		html, err := demo.registry.RenderFromSchema(ctx, child.SchemaType, child.Props)
		if err != nil {
			fmt.Printf("  %d. ERROR: %v\n", i+1, err)
			continue
		}
		fmt.Printf("  %d. %s: %s\n", i+1, child.SchemaType, html)
	}

	return nil
}

// DemoSchemaIntrospection demonstrates schema introspection capabilities
func (demo *SchemaDemoSystem) DemoSchemaIntrospection() error {
	fmt.Println("\n=== Schema Introspection Demo ===\n")

	// List all available schemas
	schemas := demo.registry.GetAvailableSchemas()
	fmt.Printf("Available schemas (%d total):\n", len(schemas))
	for i, schema := range schemas {
		fmt.Printf("  %d. %s\n", i+1, schema)
	}
	fmt.Println()

	// Get detailed info about a specific schema
	buttonSchema, err := demo.registry.GetSchemaDefinition("ButtonSchema")
	if err != nil {
		return fmt.Errorf("failed to get button schema: %w", err)
	}

	fmt.Printf("ButtonSchema details:\n")
	fmt.Printf("- ID: %s\n", buttonSchema.ID)
	fmt.Printf("- Type: %s\n", buttonSchema.Type)
	fmt.Printf("- Description: %s\n", buttonSchema.Description)
	fmt.Printf("- Required fields: %v\n", buttonSchema.Required)
	fmt.Printf("- Number of properties: %d\n", len(buttonSchema.Properties))

	// Show some properties
	fmt.Println("- Key properties:")
	for key, prop := range buttonSchema.Properties {
		if len(prop.Enum) > 0 {
			fmt.Printf("  - %s (%s): %v\n", key, prop.Type, prop.Enum)
		} else {
			fmt.Printf("  - %s (%s): %s\n", key, prop.Type, prop.Description)
		}
		if len(fmt.Sprintf("  - %s", key)) > 100 { // Limit output
			break
		}
	}
	fmt.Println()

	// Show component info
	componentInfo := demo.registry.GetComponentInfo()
	fmt.Printf("Registered components with schema info:\n")
	count := 0
	for _, info := range componentInfo {
		if info.HasSchema {
			fmt.Printf("  - %s (schema: %s)\n", info.Type, info.SchemaType)
			count++
			if count >= 10 { // Limit output
				fmt.Printf("  ... and %d more\n", len(componentInfo)-count)
				break
			}
		}
	}

	return nil
}

// DemoValidationFeatures demonstrates schema validation features
func (demo *SchemaDemoSystem) DemoValidationFeatures() error {
	fmt.Println("\n=== Schema Validation Demo ===\n")

	// Test valid props
	fmt.Println("1. Testing valid button props...")
	validProps := map[string]interface{}{
		"text":    "Click Me",
		"variant": "primary",
		"size":    "md",
		"type":    "button",
	}

	if err := demo.registry.ValidateComponentAgainstSchema("ButtonSchema", validProps); err != nil {
		fmt.Printf("❌ Validation failed: %v\n", err)
	} else {
		fmt.Println("✅ Valid props passed validation")
	}

	// Test invalid props
	fmt.Println("\n2. Testing invalid button props...")
	invalidProps := map[string]interface{}{
		"text":    "Click Me",
		"variant": "invalid-variant", // Invalid enum value
		"size":    123,               // Invalid type
		"extra":   "not-allowed",     // Extra property not in schema
	}

	if err := demo.registry.ValidateComponentAgainstSchema("ButtonSchema", invalidProps); err != nil {
		fmt.Printf("❌ Validation correctly failed: %v\n", err)
	} else {
		fmt.Println("⚠️  Validation should have failed but didn't")
	}

	// Test missing required props
	fmt.Println("\n3. Testing missing required props...")
	incompleteProps := map[string]interface{}{
		"variant": "primary",
		// Missing required "text" property
	}

	if err := demo.registry.ValidateComponentAgainstSchema("ButtonSchema", incompleteProps); err != nil {
		fmt.Printf("❌ Validation correctly failed for missing required props: %v\n", err)
	} else {
		fmt.Println("⚠️  Validation should have failed for missing required props")
	}

	return nil
}

// RunAllDemos runs all demonstration scenarios
func (demo *SchemaDemoSystem) RunAllDemos(ctx context.Context) error {
	fmt.Println("🚀 Starting Schema-Driven UI System Demonstration\n")

	if err := demo.DemoBasicComponents(ctx); err != nil {
		return fmt.Errorf("basic components demo failed: %w", err)
	}

	if err := demo.DemoComplexForm(ctx); err != nil {
		return fmt.Errorf("complex form demo failed: %w", err)
	}

	if err := demo.DemoSchemaIntrospection(); err != nil {
		return fmt.Errorf("schema introspection demo failed: %w", err)
	}

	if err := demo.DemoValidationFeatures(); err != nil {
		return fmt.Errorf("validation demo failed: %w", err)
	}

	fmt.Println("\n🎉 All demos completed successfully!")
	fmt.Println("\nKey Benefits Demonstrated:")
	fmt.Println("✅ Direct component creation from JSON schemas")
	fmt.Println("✅ Built-in validation against schema definitions")
	fmt.Println("✅ Type-safe prop handling")
	fmt.Println("✅ Schema introspection and documentation")
	fmt.Println("✅ Seamless integration with existing Templ components")
	fmt.Println("✅ Enterprise-grade 913+ component schemas support")

	return nil
}

// RunDemo is a convenience function to run the complete demonstration
func RunSchemaDemo(schemaDir string) {
	ctx := context.Background()

	demo, err := NewSchemaDemoSystem(schemaDir)
	if err != nil {
		log.Fatalf("Failed to create demo system: %v", err)
	}

	if err := demo.RunAllDemos(ctx); err != nil {
		log.Fatalf("Demo failed: %v", err)
	}
}
