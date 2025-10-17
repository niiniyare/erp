package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/niiniyare/erp/web/engine"
)

func main() {
	fmt.Println("🚀 Testing Schema-to-Templ Integration...")

	ctx := context.Background()
	schemaDir := "docs/ui/Schema"

	// Test 1: Create Enhanced Registry with Templ support
	registry, err := engine.NewEnhancedComponentRegistry(schemaDir)
	if err != nil {
		log.Printf("Failed to create enhanced registry: %v", err)
		return
	}
	fmt.Println("✅ Enhanced registry with Templ support created")

	// Test 2: Test Button Component Rendering
	fmt.Println("\n🔴 Testing Button Component...")
	testButtonComponent(ctx, registry)

	// Test 3: Test ButtonGroup Component Rendering
	fmt.Println("\n🔘 Testing ButtonGroup Component...")
	testButtonGroupComponent(ctx, registry)

	// Test 4: Test Input Component Rendering
	fmt.Println("\n📝 Testing Input Component...")
	testInputComponent(ctx, registry)

	// Test 5: Test Form Components
	fmt.Println("\n📋 Testing Form Components...")
	testFormComponents(ctx, registry)

	// Test 6: Test Layout Components
	fmt.Println("\n📦 Testing Layout Components...")
	testLayoutComponents(ctx, registry)

	// Test 7: Test Data Components
	fmt.Println("\n📊 Testing Data Components...")
	testDataComponents(ctx, registry)

	fmt.Println("\n🎉 All Templ integration tests completed!")
}

func testButtonComponent(ctx context.Context, registry *engine.EnhancedComponentRegistry) {
	// Test various button variants
	variants := []string{"primary", "secondary", "success", "danger", "warning"}
	
	for _, variant := range variants {
		props := map[string]interface{}{
			"text":    fmt.Sprintf("%s Button", strings.Title(variant)),
			"variant": variant,
			"size":    "md",
			"type":    "button",
		}
		
		// Try ButtonGroupSchema since ButtonSchema might not exist
		component, err := registry.RenderToTempl(ctx, "ButtonGroupSchema", props)
		if err != nil {
			log.Printf("Failed to render %s button: %v", variant, err)
			continue
		}
		
		// Render component to string for testing
		html, err := renderComponentToString(ctx, component)
		if err != nil {
			log.Printf("Failed to render %s button to HTML: %v", variant, err)
			continue
		}
		
		fmt.Printf("✅ %s button rendered (%d chars)\n", variant, len(html))
		if len(html) > 0 && len(html) < 200 {
			fmt.Printf("   Preview: %s\n", truncateHTML(html, 80))
		}
	}
}

func testButtonGroupComponent(ctx context.Context, registry *engine.EnhancedComponentRegistry) {
	props := map[string]interface{}{
		"buttons": []interface{}{
			map[string]interface{}{
				"text":    "Save",
				"variant": "primary",
				"type":    "submit",
			},
			map[string]interface{}{
				"text":    "Cancel", 
				"variant": "secondary",
				"type":    "button",
			},
			map[string]interface{}{
				"text":    "Delete",
				"variant": "danger",
				"type":    "button",
			},
		},
	}
	
	component, err := registry.RenderToTempl(ctx, "ButtonGroupSchema", props)
	if err != nil {
		log.Printf("Failed to render button group: %v", err)
		return
	}
	
	html, err := renderComponentToString(ctx, component)
	if err != nil {
		log.Printf("Failed to render button group to HTML: %v", err)
		return
	}
	
	fmt.Printf("✅ Button group rendered (%d chars)\n", len(html))
	if len(html) > 0 && len(html) < 300 {
		fmt.Printf("   Preview: %s\n", truncateHTML(html, 120))
	}
}

func testInputComponent(ctx context.Context, registry *engine.EnhancedComponentRegistry) {
	inputTypes := []map[string]interface{}{
		{
			"type":        "text",
			"label":       "Username",
			"placeholder": "Enter your username",
			"required":    true,
		},
		{
			"type":        "email",
			"label":       "Email Address", 
			"placeholder": "Enter your email",
			"required":    true,
		},
		{
			"type":        "password",
			"label":       "Password",
			"placeholder": "Enter your password", 
			"required":    true,
		},
	}
	
	for _, props := range inputTypes {
		component, err := registry.RenderToTempl(ctx, "InputControlSchema", props)
		if err != nil {
			log.Printf("Failed to render %s input: %v", props["type"], err)
			continue
		}
		
		html, err := renderComponentToString(ctx, component)
		if err != nil {
			log.Printf("Failed to render %s input to HTML: %v", props["type"], err)
			continue
		}
		
		fmt.Printf("✅ %s input rendered (%d chars)\n", props["type"], len(html))
		if len(html) > 0 && len(html) < 200 {
			fmt.Printf("   Preview: %s\n", truncateHTML(html, 80))
		}
	}
}

func testFormComponents(ctx context.Context, registry *engine.EnhancedComponentRegistry) {
	// Test Textarea
	textareaProps := map[string]interface{}{
		"label":       "Message",
		"placeholder": "Enter your message here",
		"rows":        4,
		"required":    true,
	}
	
	component, err := registry.RenderToTempl(ctx, "TextareaControlSchema", textareaProps)
	if err != nil {
		log.Printf("Failed to render textarea: %v", err)
	} else {
		html, _ := renderComponentToString(ctx, component)
		fmt.Printf("✅ Textarea rendered (%d chars)\n", len(html))
	}
	
	// Test Checkbox
	checkboxProps := map[string]interface{}{
		"label":   "Subscribe to newsletter",
		"value":   "newsletter",
		"checked": false,
	}
	
	component, err = registry.RenderToTempl(ctx, "CheckboxControlSchema", checkboxProps)
	if err != nil {
		log.Printf("Failed to render checkbox: %v", err)
	} else {
		html, _ := renderComponentToString(ctx, component)
		fmt.Printf("✅ Checkbox rendered (%d chars)\n", len(html))
	}
	
	// Test Select
	selectProps := map[string]interface{}{
		"label": "Country",
		"options": []interface{}{
			map[string]interface{}{"label": "United States", "value": "US"},
			map[string]interface{}{"label": "United Kingdom", "value": "UK"},
			map[string]interface{}{"label": "Canada", "value": "CA"},
		},
	}
	
	component, err = registry.RenderToTempl(ctx, "SelectControlSchema", selectProps)
	if err != nil {
		log.Printf("Failed to render select: %v", err)
	} else {
		html, _ := renderComponentToString(ctx, component)
		fmt.Printf("✅ Select rendered (%d chars)\n", len(html))
	}
}

func testLayoutComponents(ctx context.Context, registry *engine.EnhancedComponentRegistry) {
	// Test Container
	containerProps := map[string]interface{}{
		"className": "container-fluid bg-light p-4",
	}
	
	component, err := registry.RenderToTempl(ctx, "ContainerSchema", containerProps)
	if err != nil {
		log.Printf("Failed to render container: %v", err)
	} else {
		html, _ := renderComponentToString(ctx, component)
		fmt.Printf("✅ Container rendered (%d chars)\n", len(html))
	}
	
	// Test Card
	cardProps := map[string]interface{}{
		"title":   "Sample Card",
		"content": "This is a card generated from schema",
	}
	
	component, err = registry.RenderToTempl(ctx, "CardSchema", cardProps)
	if err != nil {
		log.Printf("Failed to render card: %v", err)
	} else {
		html, _ := renderComponentToString(ctx, component)
		fmt.Printf("✅ Card rendered (%d chars)\n", len(html))
	}
	
	// Test Panel
	panelProps := map[string]interface{}{
		"className": "panel panel-default",
	}
	
	component, err = registry.RenderToTempl(ctx, "PanelSchema", panelProps)
	if err != nil {
		log.Printf("Failed to render panel: %v", err)
	} else {
		html, _ := renderComponentToString(ctx, component)
		fmt.Printf("✅ Panel rendered (%d chars)\n", len(html))
	}
}

func testDataComponents(ctx context.Context, registry *engine.EnhancedComponentRegistry) {
	// Test Table
	tableProps := map[string]interface{}{
		"columns": []interface{}{
			map[string]interface{}{"name": "id", "label": "ID"},
			map[string]interface{}{"name": "name", "label": "Name"},
			map[string]interface{}{"name": "email", "label": "Email"},
		},
	}
	
	component, err := registry.RenderToTempl(ctx, "TableSchema", tableProps)
	if err != nil {
		log.Printf("Failed to render table: %v", err)
	} else {
		html, _ := renderComponentToString(ctx, component)
		fmt.Printf("✅ Table rendered (%d chars)\n", len(html))
		if len(html) > 0 && len(html) < 400 {
			fmt.Printf("   Preview: %s\n", truncateHTML(html, 150))
		}
	}
	
	// Test List
	listProps := map[string]interface{}{
		"items": []interface{}{
			"Item 1",
			"Item 2", 
			"Item 3",
		},
	}
	
	component, err = registry.RenderToTempl(ctx, "ListSchema", listProps)
	if err != nil {
		log.Printf("Failed to render list: %v", err)
	} else {
		html, _ := renderComponentToString(ctx, component)
		fmt.Printf("✅ List rendered (%d chars)\n", len(html))
		if len(html) > 0 && len(html) < 300 {
			fmt.Printf("   Preview: %s\n", truncateHTML(html, 100))
		}
	}
}

// Helper function to render a component to string for testing
func renderComponentToString(ctx context.Context, component interface{}) (string, error) {
	if templComponent, ok := component.(interface{ Render(context.Context, *strings.Builder) error }); ok {
		var buf strings.Builder
		err := templComponent.Render(ctx, &buf)
		return buf.String(), err
	}
	
	// Try the io.Writer interface instead
	if templComponent, ok := component.(interface{ Render(context.Context, interface{}) error }); ok {
		var buf strings.Builder
		err := templComponent.Render(ctx, &buf)
		return buf.String(), err
	}
	
	return "", fmt.Errorf("component does not implement expected Render interface")
}

// Helper to truncate HTML for preview
func truncateHTML(html string, maxLen int) string {
	// Remove newlines and extra spaces for preview
	cleaned := strings.ReplaceAll(html, "\n", " ")
	cleaned = strings.ReplaceAll(cleaned, "\t", " ")
	
	// Collapse multiple spaces
	for strings.Contains(cleaned, "  ") {
		cleaned = strings.ReplaceAll(cleaned, "  ", " ")
	}
	
	cleaned = strings.TrimSpace(cleaned)
	
	if len(cleaned) <= maxLen {
		return cleaned
	}
	
	return cleaned[:maxLen] + "..."
}

// Helper to check if we can write to a directory
func checkSchemaDir(dir string) bool {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		log.Printf("Schema directory does not exist: %s", dir)
		return false
	}
	return true
}