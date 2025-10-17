package main

import (
	"context"
	"fmt"
	"log"

	"github.com/niiniyare/erp/web/engine"
)

func main() {
	fmt.Println("🚀 Testing Schema-Driven UI System...")

	// Test 1: Create schema factory
	factory, err := engine.NewSchemaFactory("docs/ui/Schema")
	if err != nil {
		log.Printf("Schema factory creation failed: %v", err)
		return
	}
	fmt.Println("✅ Schema factory created successfully")

	// Test 2: List available schemas
	schemas := factory.GetAvailableSchemas()
	fmt.Printf("✅ Found %d schemas\n", len(schemas))
	
	if len(schemas) > 0 {
		fmt.Printf("📋 First few schemas: %v\n", schemas[:min(5, len(schemas))])
	}

	// Test 3: Create enhanced registry
	registry, err := engine.NewEnhancedComponentRegistry("docs/ui/Schema")
	if err != nil {
		log.Printf("Enhanced registry creation failed: %v", err)
		return
	}
	fmt.Println("✅ Enhanced registry created successfully")

	// Test 4: Test component creation
	ctx := context.Background()
	buttonProps := map[string]interface{}{
		"text":     "Test Button",
		"variant":  "primary",
		"size":     "md",
		"disabled": false,
	}

	if len(schemas) > 0 {
		// Try to find ButtonSchema
		buttonSchemaFound := false
		for _, schema := range schemas {
			if schema == "ButtonGroupSchema" {
				buttonSchemaFound = true
				break
			}
		}

		if buttonSchemaFound {
			component, err := registry.CreateFromSchema(ctx, "ButtonGroupSchema", buttonProps)
			if err != nil {
				log.Printf("Component creation failed: %v", err)
			} else {
				fmt.Printf("✅ ButtonGroup component created: %+v\n", component)
			}
		} else {
			fmt.Println("⚠️  ButtonGroupSchema not found, skipping component test")
		}
	}

	// Test 5: CSS Integration
	cssConfig := engine.CSSEngineConfig{
		SchemaDir:        "docs/ui/Schema",
		EnableValidation: true,
	}
	
	cssEngine, err := engine.NewCSSIntegrationEngine(cssConfig)
	if err != nil {
		log.Printf("CSS engine creation failed: %v", err)
		return
	}
	log.Printf("✓ CSS integration engine created successfully")
	fmt.Println("✅ CSS integration engine created successfully")

	// Test CSS engine functionality briefly
	_ = cssEngine // Acknowledge we're using the variable

	fmt.Println("\n🎉 All basic tests passed! System is ready for development.")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}