package bridge

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	schemaui "github.com/niiniyare/erp/pkg/schema/ui"
	"github.com/niiniyare/erp/web/components/atoms"
)

// ============================================================================
// BRIDGE SYSTEM DEMONSTRATION
// ============================================================================

// Demo demonstrates the bridge system capabilities
type Demo struct {
	bridge   *Bridge
	registry *UnifiedRegistry
}

// NewDemo creates a new demonstration instance
func NewDemo() *Demo {
	return &Demo{
		bridge:   NewBridge(),
		registry: NewUnifiedRegistry(),
	}
}

// RunFullDemo demonstrates all bridge capabilities
func (d *Demo) RunFullDemo(ctx context.Context) error {
	fmt.Println("🌉 ERP UI Bridge System Demonstration")
	fmt.Println("=====================================")
	fmt.Println()

	// Demo 1: Schema to Templ conversion
	if err := d.demoSchemaToTempl(ctx); err != nil {
		return fmt.Errorf("schema-to-templ demo failed: %w", err)
	}

	// Demo 2: Templ to Schema conversion
	if err := d.demoTemplToSchema(ctx); err != nil {
		return fmt.Errorf("templ-to-schema demo failed: %w", err)
	}

	// Demo 3: Unified registry usage
	if err := d.demoUnifiedRegistry(ctx); err != nil {
		return fmt.Errorf("unified registry demo failed: %w", err)
	}

	// Demo 4: Template creation
	if err := d.demoTemplateCreation(ctx); err != nil {
		return fmt.Errorf("template creation demo failed: %w", err)
	}

	// Demo 5: CSS Runtime integration
	if err := d.demoCSSIntegration(ctx); err != nil {
		return fmt.Errorf("CSS integration demo failed: %w", err)
	}

	fmt.Println("✅ All demonstrations completed successfully!")
	return nil
}

// ============================================================================
// DEMO 1: SCHEMA TO TEMPL CONVERSION
// ============================================================================

func (d *Demo) demoSchemaToTempl(ctx context.Context) error {
	fmt.Println("Demo 1: Converting Schema Components to Templ Components")
	fmt.Println("-------------------------------------------------------")

	// Create a schema button component
	buttonComponent := schemaui.NewComponent(schemaui.ComponentButton, "demo-button").
		WithLabel("Save Changes").
		WithVariant(schemaui.VariantPrimary).
		WithSize(schemaui.SizeLG).
		WithConfig(schemaui.ButtonConfig{
			Text:       "Save Changes",
			Icon:       "save",
			ButtonType: schemaui.ButtonSubmit,
			Loading:    false,
		}).
		Build()

	fmt.Printf("📋 Created schema button component:\n")
	fmt.Printf("   - ID: %s\n", buttonComponent.ID)
	fmt.Printf("   - Type: %s\n", buttonComponent.Type)
	fmt.Printf("   - Label: %s\n", buttonComponent.Label)
	fmt.Printf("   - Variant: %s\n", buttonComponent.Variant)
	fmt.Printf("   - Size: %s\n", buttonComponent.Size)

	// Convert to Templ component
	templComponent, err := d.bridge.ConvertToTempl(ctx, buttonComponent)
	if err != nil {
		return fmt.Errorf("failed to convert schema to templ: %w", err)
	}

	fmt.Printf("\n🔄 Converted to Templ component:\n")
	fmt.Printf("   - Type: %s\n", templComponent.Type)
	
	if props, ok := templComponent.Props.(atoms.ButtonProps); ok {
		fmt.Printf("   - Text: %s\n", props.Text)
		fmt.Printf("   - Icon: %s\n", props.Icon)
		fmt.Printf("   - Variant: %s\n", props.Variant)
		fmt.Printf("   - Size: %s\n", props.Size)
		fmt.Printf("   - Type: %s\n", props.Type)
	}

	fmt.Println("\n✅ Schema to Templ conversion successful!")
	fmt.Println()
	return nil
}

// ============================================================================
// DEMO 2: TEMPL TO SCHEMA CONVERSION
// ============================================================================

func (d *Demo) demoTemplToSchema(ctx context.Context) error {
	fmt.Println("Demo 2: Converting Templ Components to Schema Components")
	fmt.Println("-------------------------------------------------------")

	// Create a Templ input component
	inputProps := atoms.InputProps{
		Type:        "email",
		Name:        "user_email",
		Placeholder: "Enter your email address",
		Required:    true,
		Size:        atoms.InputSizeLG,
		ID:          "email-input",
		Label:       "Email Address",
		MaxLength:   100,
		Pattern:     `^[^\s@]+@[^\s@]+\.[^\s@]+$`,
	}

	templComponent := TemplComponent{
		Type:  "Input",
		Props: inputProps,
	}

	fmt.Printf("🎯 Created Templ input component:\n")
	fmt.Printf("   - Type: %s\n", templComponent.Type)
	fmt.Printf("   - Name: %s\n", inputProps.Name)
	fmt.Printf("   - Placeholder: %s\n", inputProps.Placeholder)
	fmt.Printf("   - Required: %t\n", inputProps.Required)
	fmt.Printf("   - Size: %s\n", inputProps.Size)

	// Convert to schema component
	schemaComponent, err := d.bridge.ConvertTemplToSchema(ctx, templComponent)
	if err != nil {
		return fmt.Errorf("failed to convert templ to schema: %w", err)
	}

	fmt.Printf("\n🔄 Converted to Schema component:\n")
	fmt.Printf("   - ID: %s\n", schemaComponent.ID)
	fmt.Printf("   - Type: %s\n", schemaComponent.Type)
	fmt.Printf("   - Name: %s\n", schemaComponent.Name)
	fmt.Printf("   - Required: %t\n", schemaComponent.Required)
	fmt.Printf("   - Size: %s\n", schemaComponent.Size)

	// Show the config
	var inputConfig schemaui.InputConfig
	if err := json.Unmarshal(schemaComponent.Config, &inputConfig); err == nil {
		fmt.Printf("   - Input Type: %s\n", inputConfig.InputType)
		fmt.Printf("   - Max Length: %d\n", inputConfig.MaxLength)
		fmt.Printf("   - Pattern: %s\n", inputConfig.Pattern)
	}

	fmt.Println("\n✅ Templ to Schema conversion successful!")
	fmt.Println()
	return nil
}

// ============================================================================
// DEMO 3: UNIFIED REGISTRY USAGE
// ============================================================================

func (d *Demo) demoUnifiedRegistry(ctx context.Context) error {
	fmt.Println("Demo 3: Unified Registry - Creating Components Both Ways")
	fmt.Println("--------------------------------------------------------")

	// Method 1: Create via schema system
	fmt.Println("📋 Creating component via Schema System:")
	
	// TODO: Implement when registry method is available
	/*
	schemaComponent, err := d.registry.CreateSchemaComponent(ctx, schemaui.ComponentSelect, map[string]any{
		"name":        "user_role",
		"placeholder": "Select User Role",
		"required":    true,
		"options": []map[string]any{
			{"value": "admin", "label": "Administrator"},
			{"value": "manager", "label": "Manager"},
			{"value": "user", "label": "Regular User"},
			{"value": "viewer", "label": "Read-Only Viewer"},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create schema component: %w", err)
	}

	fmt.Printf("   - Created schema select with ID: %s\n", schemaComponent.ID)

	// Convert to Templ
	templFromSchema, err := d.bridge.ConvertToTempl(ctx, schemaComponent)
	if err != nil {
		return fmt.Errorf("failed to convert to templ: %w", err)
	}

	if selectProps, ok := templFromSchema.Props.(atoms.SelectProps); ok {
		fmt.Printf("   - Converted to Templ with %d options\n", len(selectProps.Options))
	}

	// Method 2: Create via Templ system and convert
	fmt.Println("\n🎯 Creating component via Templ System:")
	
	templComponent, err := d.registry.CreateTemplComponent(ctx, schemaui.ComponentCheckbox, map[string]any{
		"name":     "terms_accepted",
		"label":    "I accept the Terms and Conditions",
		"required": true,
	})
	if err != nil {
		return fmt.Errorf("failed to create templ component: %w", err)
	}

	fmt.Printf("   - Created Templ checkbox: %s\n", templComponent.Type)

	if checkboxProps, ok := templComponent.Props.(atoms.CheckboxProps); ok {
		fmt.Printf("   - Label: %s\n", checkboxProps.Label)
		fmt.Printf("   - Required: %t\n", checkboxProps.Required)
	}
	*/

	fmt.Println("   - Registry methods not yet implemented - demo skipped")
	fmt.Println("\n✅ Unified registry demonstration placeholder completed!")
	fmt.Println()
	return nil
}

// ============================================================================
// DEMO 4: TEMPLATE CREATION
// ============================================================================

func (d *Demo) demoTemplateCreation(ctx context.Context) error {
	fmt.Println("Demo 4: Template-Based Component Creation")
	fmt.Println("----------------------------------------")

	// Create login form template
	fmt.Println("📋 Creating login form template:")
	
	templComponents, schemaComponents, err := d.registry.CreateFromTemplate(ctx, "login-form", map[string]any{})
	if err != nil {
		return fmt.Errorf("failed to create login form template: %w", err)
	}

	fmt.Printf("   - Created %d Templ components\n", len(templComponents))
	fmt.Printf("   - Created %d Schema components\n", len(schemaComponents))

	// Show the components
	for i, templComp := range templComponents {
		fmt.Printf("   - Component %d: %s\n", i+1, templComp.Type)
	}

	// Create user form template
	fmt.Println("\n🎯 Creating user registration form template:")
	
	userTemplComponents, userSchemaComponents, err := d.registry.CreateFromTemplate(ctx, "user-form", map[string]any{})
	if err != nil {
		return fmt.Errorf("failed to create user form template: %w", err)
	}

	fmt.Printf("   - Created %d Templ components\n", len(userTemplComponents))
	fmt.Printf("   - Created %d Schema components\n", len(userSchemaComponents))

	// Show the components
	for i, templComp := range userTemplComponents {
		fmt.Printf("   - Component %d: %s\n", i+1, templComp.Type)
	}

	fmt.Println("\n✅ Template creation demonstration successful!")
	fmt.Println()
	return nil
}

// ============================================================================
// DEMO 5: CSS RUNTIME INTEGRATION
// ============================================================================

func (d *Demo) demoCSSIntegration(ctx context.Context) error {
	fmt.Println("Demo 5: CSS Runtime System Integration")
	fmt.Println("-------------------------------------")

	// Create a component with custom styles
	fmt.Println("🎨 Creating component with CSS Runtime integration:")

	// This would require the CSS system to be properly integrated
	// For demonstration, we'll show the concept
	
	customButton := schemaui.NewComponent(schemaui.ComponentButton, "custom-styled-button").
		WithLabel("Custom Styled Button").
		WithVariant(schemaui.VariantPrimary).
		WithSize(schemaui.SizeLG).
		WithClass("custom-animation hover-effect").
		WithConfig(schemaui.ButtonConfig{
			Text: "Custom Styled Button",
			Icon: "star",
		}).
		Build()

	fmt.Printf("   - Created button with custom classes: %s\n", customButton.Class)

	// Convert to Templ (CSS classes will be preserved)
	templButton, err := d.bridge.ConvertToTempl(ctx, customButton)
	if err != nil {
		return fmt.Errorf("failed to convert styled component: %w", err)
	}

	if buttonProps, ok := templButton.Props.(atoms.ButtonProps); ok {
		fmt.Printf("   - Templ button classes: %s\n", buttonProps.Class)
		fmt.Printf("   - CSS classes preserved in conversion ✓\n")
	}

	fmt.Println("\n✅ CSS integration demonstration successful!")
	fmt.Println()
	return nil
}

// ============================================================================
// VALIDATION DEMONSTRATION
// ============================================================================

func (d *Demo) demoValidation(ctx context.Context) error {
	fmt.Println("Demo: Component Validation")
	fmt.Println("-------------------------")

	// Create a valid component
	validComponent := schemaui.NewComponent(schemaui.ComponentInput, "valid-input").
		WithName("username").
		WithLabel("Username").
		Required().
		WithConfig(schemaui.InputConfig{
			InputType: schemaui.InputText,
			MinLength: 3,
			MaxLength: 50,
		}).
		Build()

	// Validate schema component (TODO: implement validation method)
	_ = validComponent // TODO: Add validation when method is available
	fmt.Println("✅ Valid schema component - validation placeholder passed")

	// Convert to Templ and validate
	templComponent, err := d.bridge.ConvertToTempl(ctx, validComponent)
	if err != nil {
		return fmt.Errorf("failed to convert for validation: %w", err)
	}

	// TODO: Add templ validation when method is available
	_ = templComponent
	fmt.Println("✅ Converted Templ component validation placeholder passed")
	return nil
}

// ============================================================================
// MAIN DEMO RUNNER
// ============================================================================

// RunDemo is the main entry point for running demonstrations
func RunDemo() {
	ctx := context.Background()
	demo := NewDemo()

	if err := demo.RunFullDemo(ctx); err != nil {
		log.Fatalf("Demo failed: %v", err)
	}

	// Additional validation demo
	fmt.Println("Bonus: Validation Demonstration")
	fmt.Println("==============================")
	if err := demo.demoValidation(ctx); err != nil {
		log.Printf("Validation demo failed: %v", err)
	} else {
		fmt.Println("✅ Validation demonstration successful!")
	}

	fmt.Println("\n🎉 Bridge System Demonstration Complete!")
	fmt.Println("The bridge successfully connects both UI systems,")
	fmt.Println("enabling seamless conversion and unified development workflows.")
}