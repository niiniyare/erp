// Test file for schema-to-CSS integration
package css

import (
	"strings"
	"testing"
)

// MockButtonSchema represents a button component schema for testing
type MockButtonSchema struct {
	Type      string `json:"type"`
	Size      string `json:"size,omitempty"`
	Variant   string `json:"variant,omitempty"`
	Disabled  bool   `json:"disabled,omitempty"`
	ClassName string `json:"className,omitempty"`
	Style     map[string]any `json:"style,omitempty"`
}

// MockInputSchema represents an input component schema for testing
type MockInputSchema struct {
	Type      string `json:"type"`
	Size      string `json:"size,omitempty"`
	Required  bool   `json:"required,omitempty"`
	ReadOnly  bool   `json:"readOnly,omitempty"`
	ClassName string `json:"className,omitempty"`
}

// MockFormSchema represents a form component schema for testing
type MockFormSchema struct {
	Type      string `json:"type"`
	Mode      string `json:"mode,omitempty"`
	ClassName string `json:"className,omitempty"`
}

// TestSchemaToCSS tests basic schema-to-CSS conversion
func TestSchemaToCSS(t *testing.T) {
	converter := NewSchemaToCSS()
	
	// Test button schema conversion
	buttonSchema := MockButtonSchema{
		Type:    "button",
		Size:    "lg",
		Variant: "primary",
	}
	
	css, err := converter.GenerateCSSFromSchema(buttonSchema)
	if err != nil {
		t.Fatalf("Failed to generate CSS from button schema: %v", err)
	}
	
	// Check that button CSS was generated
	if !strings.Contains(css, ".ui-btn-primary-lg") {
		t.Error("Generated CSS should contain button class")
	}
	
	// Check for expected button properties
	if !strings.Contains(css, "display: inline-flex") {
		t.Error("Button should have inline-flex display")
	}
}

// TestSchemaToCSSSStateClasses tests state class extraction
func TestSchemaToCSSSStateClasses(t *testing.T) {
	converter := NewSchemaToCSS()
	
	// Test input with required and readonly states
	inputSchema := MockInputSchema{
		Type:     "input-text",
		Size:     "md",
		Required: true,
		ReadOnly: true,
	}
	
	style, err := converter.ExtractStyleFromSchema(inputSchema)
	if err != nil {
		t.Fatalf("Failed to extract style from input schema: %v", err)
	}
	
	// Check state classes
	if !containsString(style.StateClasses, "required") {
		t.Error("Style should contain required state class")
	}
	
	if !containsString(style.StateClasses, "readonly") {
		t.Error("Style should contain readonly state class")
	}
	
	// Check mapped type
	if style.Type != "input" {
		t.Errorf("Expected type 'input', got '%s'", style.Type)
	}
}

// TestSchemaToCSSSCustomStyles tests custom style extraction
func TestSchemaToCSSSCustomStyles(t *testing.T) {
	converter := NewSchemaToCSS()
	
	// Test button with custom styles
	buttonSchema := MockButtonSchema{
		Type:    "button",
		Variant: "custom",
		Style: map[string]any{
			"background-color": "#ff6b6b",
			"border-radius":    "50px",
			"font-weight":      "bold",
		},
	}
	
	style, err := converter.ExtractStyleFromSchema(buttonSchema)
	if err != nil {
		t.Fatalf("Failed to extract style from button schema: %v", err)
	}
	
	// Check custom styles
	if style.CustomStyles["background-color"] != "#ff6b6b" {
		t.Error("Custom background color should be extracted")
	}
	
	if style.CustomStyles["border-radius"] != "50px" {
		t.Error("Custom border radius should be extracted")
	}
	
	// Generate CSS and check that custom styles are included
	css, err := converter.GenerateCSSFromStyle(style)
	if err != nil {
		t.Fatalf("Failed to generate CSS from style: %v", err)
	}
	
	if !strings.Contains(css, "background-color: #ff6b6b") {
		t.Errorf("Generated CSS should contain custom background color. CSS output:\n%s", css)
	}
	
	// Also check that the custom class was generated
	if !strings.Contains(css, ".ui-btn-custom-md") {
		t.Errorf("Should generate custom button class .ui-btn-custom-md")
	}
}

// TestSchemaToCSSSClassName tests CSS class name extraction
func TestSchemaToCSSSClassName(t *testing.T) {
	converter := NewSchemaToCSS()
	
	// Test input with custom class names
	inputSchema := MockInputSchema{
		Type:      "input-text",
		ClassName: "custom-input another-class",
	}
	
	style, err := converter.ExtractStyleFromSchema(inputSchema)
	if err != nil {
		t.Fatalf("Failed to extract style from input schema: %v", err)
	}
	
	// Check base classes
	if !containsString(style.BaseClasses, "custom-input") {
		t.Error("Style should contain custom-input class")
	}
	
	if !containsString(style.BaseClasses, "another-class") {
		t.Error("Style should contain another-class class")
	}
}

// TestSchemaToCSSSDefaults tests default value handling
func TestSchemaToCSSSDefaults(t *testing.T) {
	converter := NewSchemaToCSS()
	
	// Test schema with minimal properties
	buttonSchema := MockButtonSchema{
		Type: "button",
		// Size and Variant not specified - should get defaults
	}
	
	style, err := converter.ExtractStyleFromSchema(buttonSchema)
	if err != nil {
		t.Fatalf("Failed to extract style from button schema: %v", err)
	}
	
	// Check defaults
	if style.Size != "md" {
		t.Errorf("Expected default size 'md', got '%s'", style.Size)
	}
	
	if style.Variant != "default" {
		t.Errorf("Expected default variant 'default', got '%s'", style.Variant)
	}
}

// TestSchemaToCSSSComponentLibrary tests generation for multiple components
func TestSchemaToCSSSComponentLibrary(t *testing.T) {
	converter := NewSchemaToCSS()
	
	// Create a library of component schemas
	schemas := map[string]interface{}{
		"primary-button": MockButtonSchema{
			Type:    "button",
			Variant: "primary",
			Size:    "lg",
		},
		"email-input": MockInputSchema{
			Type:     "input-email",
			Size:     "md",
			Required: true,
		},
		"contact-form": MockFormSchema{
			Type: "form",
			Mode: "horizontal",
		},
	}
	
	css, err := converter.GenerateComponentLibraryCSS(schemas)
	if err != nil {
		t.Fatalf("Failed to generate CSS from component library: %v", err)
	}
	
	// Check that CSS for all components was generated
	if !strings.Contains(css, ".ui-btn-primary-lg") {
		t.Error("Library CSS should contain button styles")
	}
	
	if !strings.Contains(css, ".ui-input-md") {
		t.Error("Library CSS should contain input styles")
	}
	
	// Check that utilities were generated (should be included by default)
	if !strings.Contains(css, ".ui-m-md") {
		t.Error("Library CSS should contain utility classes")
	}
	
	// Check that theme variables were generated
	if !strings.Contains(css, "--color-primary-500") {
		t.Error("Library CSS should contain theme variables")
	}
}

// TestTypedComponentCSSGenerator tests type-safe CSS generation
func TestTypedComponentCSSGenerator(t *testing.T) {
	generator := NewTypedComponentCSSGenerator()
	
	// Test typed button generation
	buttonSchema := MockButtonSchema{
		Type:    "action", // Different type, but should be forced to button
		Variant: "success",
		Size:    "sm",
	}
	
	css, err := generator.GenerateButtonCSS(buttonSchema)
	if err != nil {
		t.Fatalf("Failed to generate button CSS: %v", err)
	}
	
	// Should contain button-specific classes regardless of input type
	if !strings.Contains(css, ".ui-btn-success-sm") {
		t.Error("Typed generator should force button type")
	}
}

// TestTypedFormCSSGenerator tests form-specific CSS generation
func TestTypedFormCSSGenerator(t *testing.T) {
	generator := NewTypedComponentCSSGenerator()
	
	formSchema := MockFormSchema{
		Type: "form",
		Mode: "horizontal",
	}
	
	css, err := generator.GenerateFormCSS(formSchema)
	if err != nil {
		t.Fatalf("Failed to generate form CSS: %v", err)
	}
	
	// Check for form-specific styles
	if !strings.Contains(css, ".ui-form") {
		t.Error("Form CSS should contain form class")
	}
	
	// Check for form control spacing
	if !strings.Contains(css, ".ui-form .form-control") {
		t.Error("Form CSS should contain form control styles")
	}
}

// TestTypedTableCSSGenerator tests table-specific CSS generation
func TestTypedTableCSSGenerator(t *testing.T) {
	generator := NewTypedComponentCSSGenerator()
	
	// Mock table schema
	tableSchema := struct {
		Type    string `json:"type"`
		Variant string `json:"variant"`
	}{
		Type:    "table",
		Variant: "striped",
	}
	
	css, err := generator.GenerateTableCSS(tableSchema)
	if err != nil {
		t.Fatalf("Failed to generate table CSS: %v", err)
	}
	
	// Check for table-specific styles
	if !strings.Contains(css, ".ui-table-striped") {
		t.Error("Table CSS should contain table class")
	}
	
	// Check for table header styles
	if !strings.Contains(css, ".ui-table-striped th") {
		t.Error("Table CSS should contain header styles")
	}
	
	// Check for table cell styles
	if !strings.Contains(css, ".ui-table-striped td") {
		t.Error("Table CSS should contain cell styles")
	}
	
	// Check for expected table properties
	if !strings.Contains(css, "border-collapse: collapse") {
		t.Error("Table should have collapsed borders")
	}
}

// TestSchemaIntegrationConfig tests custom configuration
func TestSchemaIntegrationConfig(t *testing.T) {
	// Custom configuration
	config := &SchemaIntegrationConfig{
		GenerateUtilities:     false, // Disable utilities
		IncludeThemeVariables: false, // Disable theme variables
		GenerateResponsive:    false, // Disable responsive
		ComponentTypeMappings: map[string]string{
			"custom-button": "button",
		},
		ClassPatterns: map[string]string{
			"button": "custom-btn-{variant}",
		},
	}
	
	converter := NewSchemaToCSSSWithConfig(config)
	
	buttonSchema := struct {
		Type    string `json:"type"`
		Variant string `json:"variant"`
	}{
		Type:    "custom-button",
		Variant: "danger",
	}
	
	css, err := converter.GenerateCSSFromSchema(buttonSchema)
	if err != nil {
		t.Fatalf("Failed to generate CSS with custom config: %v", err)
	}
	
	// Check custom class pattern (with ui- prefix)
	if !strings.Contains(css, ".ui-custom-btn-danger") {
		t.Errorf("Should use custom class pattern. CSS output:\n%s", css)
	}
	
	// Check that utilities were not generated
	if strings.Contains(css, ".ui-m-md") {
		t.Error("Utilities should not be generated when disabled")
	}
	
	// Check that theme variables were not generated
	if strings.Contains(css, "--color-primary-500") {
		t.Error("Theme variables should not be generated when disabled")
	}
}

// TestSchemaToCSSSErrorHandling tests error handling for invalid schemas
func TestSchemaToCSSSErrorHandling(t *testing.T) {
	converter := NewSchemaToCSS()
	
	// Test with non-struct type
	_, err := converter.GenerateCSSFromSchema("not a struct")
	if err == nil {
		t.Error("Should return error for non-struct schema")
	}
	
	// Test with nil
	_, err = converter.GenerateCSSFromSchema(nil)
	if err == nil {
		t.Error("Should return error for nil schema")
	}
}

// TestBuildClassName tests CSS class name building
func TestBuildClassName(t *testing.T) {
	converter := NewSchemaToCSS()
	
	testCases := []struct {
		style    ComponentStyle
		expected string
	}{
		{
			style:    ComponentStyle{Type: "button", Variant: "primary", Size: "lg"},
			expected: ".btn-primary-lg",
		},
		{
			style:    ComponentStyle{Type: "input", Variant: "default", Size: "md"},
			expected: ".input-md",
		},
		{
			style:    ComponentStyle{Type: "card", Variant: "elevated", Size: "md"},
			expected: ".card-elevated",
		},
	}
	
	for _, tc := range testCases {
		className := converter.buildClassName(&tc.style)
		if className != tc.expected {
			t.Errorf("Expected class name '%s', got '%s'", tc.expected, className)
		}
	}
}

// Helper function to check if slice contains string
func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// BenchmarkSchemaToCSS benchmarks schema-to-CSS conversion performance
func BenchmarkSchemaToCSS(b *testing.B) {
	converter := NewSchemaToCSS()
	
	buttonSchema := MockButtonSchema{
		Type:    "button",
		Variant: "primary",
		Size:    "md",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = converter.GenerateCSSFromSchema(buttonSchema)
		converter.Reset() // Reset for next iteration
	}
}

// BenchmarkComponentLibraryCSS benchmarks library CSS generation
func BenchmarkComponentLibraryCSS(b *testing.B) {
	converter := NewSchemaToCSS()
	
	schemas := map[string]interface{}{
		"button1": MockButtonSchema{Type: "button", Variant: "primary", Size: "sm"},
		"button2": MockButtonSchema{Type: "button", Variant: "secondary", Size: "md"},
		"button3": MockButtonSchema{Type: "button", Variant: "success", Size: "lg"},
		"input1":  MockInputSchema{Type: "input-text", Size: "md", Required: true},
		"input2":  MockInputSchema{Type: "input-email", Size: "lg", ReadOnly: true},
		"form1":   MockFormSchema{Type: "form", Mode: "horizontal"},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = converter.GenerateComponentLibraryCSS(schemas)
		converter.Reset()
	}
}