// Package css - Schema integration for automatic CSS generation from component schemas
// This system bridges UI component schemas with CSS runtime generation
package css

import (
	"fmt"
	"reflect"
	"strings"
)

// SchemaToCSS provides automatic CSS generation from component schemas
type SchemaToCSS struct {
	builder *ComponentCSSBuilder
	config  *SchemaIntegrationConfig
}

// SchemaIntegrationConfig contains configuration for schema-to-CSS conversion
type SchemaIntegrationConfig struct {
	// Auto-generate utility classes
	GenerateUtilities bool
	// Include theme variables in output
	IncludeThemeVariables bool
	// Generate responsive variants
	GenerateResponsive bool
	// Generate state variants (hover, focus, etc.)
	GenerateStates bool
	// Component type mapping
	ComponentTypeMappings map[string]string
	// Custom CSS class patterns
	ClassPatterns map[string]string
	// Theme customization
	CustomTheme *ComponentTheme
}

// ComponentStyle represents CSS styling extracted from a component schema
type ComponentStyle struct {
	// Component identifier
	ComponentID string
	// Component type (button, input, etc.)
	Type string
	// Base CSS classes to apply
	BaseClasses []string
	// Size variant
	Size string
	// Color/visual variant
	Variant string
	// State classes (disabled, loading, etc.)
	StateClasses []string
	// Custom style properties from schema
	CustomStyles map[string]string
	// Responsive breakpoint styles
	ResponsiveStyles map[string]map[string]string
}

// DefaultSchemaIntegrationConfig returns sensible defaults
func DefaultSchemaIntegrationConfig() *SchemaIntegrationConfig {
	return &SchemaIntegrationConfig{
		GenerateUtilities:     true,
		IncludeThemeVariables: true,
		GenerateResponsive:    true,
		GenerateStates:       true,
		ComponentTypeMappings: map[string]string{
			"input-text":   "input",
			"input-email":  "input",
			"input-number": "input",
			"input-file":   "file",
			"input-array":  "array",
			"select":       "select",
			"textarea":     "textarea",
			"checkbox":     "checkbox",
			"radio":        "radio",
			"button":       "button",
			"submit":       "button",
			"reset":        "button",
			"form":         "form",
			"table":        "table",
			"crud":         "table",
			"chart":        "chart",
			"dialog":       "modal",
			"drawer":       "drawer",
			"wizard":       "wizard",
			"tabs":         "tabs",
			"nav":          "nav",
			"card":         "card",
			"panel":        "card",
			"alert":        "alert",
			"combo":        "combo",
			"editor":       "editor",
		},
		ClassPatterns: map[string]string{
			"button":  "btn-{variant}-{size}",
			"input":   "input-{size}",
			"card":    "card-{variant}",
			"modal":   "modal-{size}",
			"alert":   "alert-{variant}",
		},
	}
}

// NewSchemaToCSS creates a new schema-to-CSS converter
func NewSchemaToCSS() *SchemaToCSS {
	return &SchemaToCSS{
		builder: NewComponentCSSBuilder(),
		config:  DefaultSchemaIntegrationConfig(),
	}
}

// NewSchemaToCSSSWithConfig creates a converter with custom configuration
func NewSchemaToCSSSWithConfig(config *SchemaIntegrationConfig) *SchemaToCSS {
	builder := NewComponentCSSBuilder()
	
	if config.CustomTheme != nil {
		builder = NewComponentCSSBuilderWithTheme(config.CustomTheme)
	}
	
	return &SchemaToCSS{
		builder: builder,
		config:  config,
	}
}

// GenerateCSSFromSchema generates CSS from any component schema using reflection
func (s *SchemaToCSS) GenerateCSSFromSchema(schema interface{}) (string, error) {
	style, err := s.ExtractStyleFromSchema(schema)
	if err != nil {
		return "", fmt.Errorf("failed to extract style from schema: %w", err)
	}
	
	return s.GenerateCSSFromStyle(style)
}

// ExtractStyleFromSchema extracts styling information from a component schema
func (s *SchemaToCSS) ExtractStyleFromSchema(schema interface{}) (*ComponentStyle, error) {
	v := reflect.ValueOf(schema)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("schema must be a struct, got %T", schema)
	}
	
	t := v.Type()
	style := &ComponentStyle{
		CustomStyles:     make(map[string]string),
		ResponsiveStyles: make(map[string]map[string]string),
		BaseClasses:      make([]string, 0),
		StateClasses:     make([]string, 0),
	}
	
	// Extract common fields using reflection
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)
		
		if !value.IsValid() || !value.CanInterface() {
			continue
		}
		
		switch strings.ToLower(field.Name) {
		case "type":
			if str, ok := value.Interface().(string); ok && str != "" {
				style.Type = s.mapComponentType(str)
				style.ComponentID = str
			}
		case "size":
			if str, ok := value.Interface().(string); ok && str != "" {
				style.Size = str
			}
		case "variant", "level", "color":
			if str, ok := value.Interface().(string); ok && str != "" {
				style.Variant = str
			}
		case "disabled":
			if disabled, ok := value.Interface().(bool); ok && disabled {
				style.StateClasses = append(style.StateClasses, "disabled")
			}
		case "readonly":
			if readonly, ok := value.Interface().(bool); ok && readonly {
				style.StateClasses = append(style.StateClasses, "readonly")
			}
		case "required":
			if required, ok := value.Interface().(bool); ok && required {
				style.StateClasses = append(style.StateClasses, "required")
			}
		case "loading":
			if loading, ok := value.Interface().(bool); ok && loading {
				style.StateClasses = append(style.StateClasses, "loading")
			}
		case "classname":
			if str, ok := value.Interface().(string); ok && str != "" {
				classes := strings.Fields(str)
				style.BaseClasses = append(style.BaseClasses, classes...)
			}
		case "style":
			if styleMap, ok := value.Interface().(map[string]any); ok {
				for prop, val := range styleMap {
					if strVal, ok := val.(string); ok {
						style.CustomStyles[prop] = strVal
					}
				}
			}
		}
	}
	
	// Set defaults if not specified
	if style.Type == "" {
		style.Type = "generic"
	}
	if style.Size == "" {
		style.Size = "md"
	}
	if style.Variant == "" {
		style.Variant = "default"
	}
	
	return style, nil
}

// mapComponentType maps schema type to CSS component type
func (s *SchemaToCSS) mapComponentType(schemaType string) string {
	if mapped, exists := s.config.ComponentTypeMappings[schemaType]; exists {
		return mapped
	}
	return schemaType
}

// GenerateCSSFromStyle generates CSS from ComponentStyle
func (s *SchemaToCSS) GenerateCSSFromStyle(style *ComponentStyle) (string, error) {
	// Build the custom class name that should be used
	className := s.buildClassName(style)
	
	// Generate component-specific CSS with custom class name
	options := ComponentCSSOptions{
		ComponentType: style.Type,
		Size:         style.Size,
		Variant:      style.Variant,
		Disabled:     s.hasStateClass(style, "disabled"),
		CustomClasses: style.BaseClasses,
		Responsive:   style.ResponsiveStyles,
		CustomClassName: className, // Pass the custom class name
	}
	
	// Generate CSS based on component type, but use schema-based class names
	switch style.Type {
	case "button":
		s.generateSchemaButtonCSS(style, options)
	case "input", "email", "text", "number":
		s.generateSchemaInputCSS(style, options)
	case "card", "panel":
		s.generateSchemaCardCSS(style, options)
	default:
		// Generate generic component CSS
		s.generateGenericComponentCSS(style, options)
	}
	
	// Generate responsive CSS if enabled
	if s.config.GenerateResponsive && len(style.ResponsiveStyles) > 0 {
		s.builder.GenerateResponsiveCSS(className, style.ResponsiveStyles)
	}
	
	// Custom styles are now merged into component-specific methods
	// for better integration and to avoid duplicate rules
	
	// Generate utility classes if enabled
	if s.config.GenerateUtilities {
		s.builder.GenerateUtilityCSS()
	}
	
	// Add theme variables if enabled
	if s.config.IncludeThemeVariables {
		s.builder.AddThemeVariables()
	}
	
	return s.builder.Generate(), nil
}

// generateGenericComponentCSS generates CSS for unrecognized component types
func (s *SchemaToCSS) generateGenericComponentCSS(style *ComponentStyle, options ComponentCSSOptions) {
	className := s.buildClassName(style)
	
	// Base generic styles
	baseStyles := map[string]string{
		"box-sizing": "border-box",
		"position":   "relative",
	}
	
	// Add size-based styles
	switch style.Size {
	case "xs":
		baseStyles["font-size"] = "0.75rem"
		baseStyles["padding"] = "0.25rem"
	case "sm":
		baseStyles["font-size"] = "0.875rem"
		baseStyles["padding"] = "0.5rem"
	case "md":
		baseStyles["font-size"] = "1rem"
		baseStyles["padding"] = "1rem"
	case "lg":
		baseStyles["font-size"] = "1.125rem"
		baseStyles["padding"] = "1.5rem"
	case "xl":
		baseStyles["font-size"] = "1.25rem"
		baseStyles["padding"] = "2rem"
	}
	
	s.builder.GetGenerator().AddRule(className, baseStyles)
	
	// Add state styles
	if s.hasStateClass(style, "disabled") {
		disabledStyles := map[string]string{
			"opacity":        "0.5",
			"pointer-events": "none",
		}
		s.builder.GetGenerator().AddRuleWithPseudo(className, ":disabled", disabledStyles)
	}
}

// buildClassName builds the CSS class name for a component
func (s *SchemaToCSS) buildClassName(style *ComponentStyle) string {
	if pattern, exists := s.config.ClassPatterns[style.Type]; exists {
		className := pattern
		className = strings.ReplaceAll(className, "{variant}", style.Variant)
		className = strings.ReplaceAll(className, "{size}", style.Size)
		className = strings.ReplaceAll(className, "{type}", style.Type)
		if !strings.HasPrefix(className, ".") {
			className = "." + className
		}
		return className
	}
	
	// Default pattern
	if style.Variant != "default" && style.Size != "md" {
		return fmt.Sprintf(".%s-%s-%s", style.Type, style.Variant, style.Size)
	} else if style.Variant != "default" {
		return fmt.Sprintf(".%s-%s", style.Type, style.Variant)
	} else if style.Size != "md" {
		return fmt.Sprintf(".%s-%s", style.Type, style.Size)
	}
	
	return fmt.Sprintf(".%s", style.Type)
}

// generateSchemaButtonCSS generates button CSS using schema-based class name
func (s *SchemaToCSS) generateSchemaButtonCSS(style *ComponentStyle, options ComponentCSSOptions) {
	className := strings.TrimPrefix(options.CustomClassName, ".")
	
	// Base button styles using theme
	baseStyles := map[string]string{
		"display":         "inline-flex",
		"align-items":     "center",
		"justify-content": "center",
		"border":          "1px solid transparent",
		"font-weight":     "500",
		"text-decoration": "none",
		"cursor":          "pointer",
		"transition":      "all 150ms cubic-bezier(0.4, 0, 0.2, 1)",
		"outline":         "none",
		"user-select":     "none",
	}
	
	// Merge custom styles into base styles
	for prop, value := range style.CustomStyles {
		baseStyles[prop] = value
	}
	
	// Size-specific styles
	switch options.Size {
	case "xs":
		baseStyles["padding"] = s.builder.theme.Spacing["xs"] + " " + s.builder.theme.Spacing["sm"]
		baseStyles["font-size"] = s.builder.theme.Typography["xs"].FontSize
		baseStyles["border-radius"] = s.builder.theme.BorderRadius["sm"]
	case "sm":
		baseStyles["padding"] = s.builder.theme.Spacing["sm"] + " " + s.builder.theme.Spacing["md"]
		baseStyles["font-size"] = s.builder.theme.Typography["sm"].FontSize
		baseStyles["border-radius"] = s.builder.theme.BorderRadius["md"]
	case "lg":
		baseStyles["padding"] = s.builder.theme.Spacing["lg"] + " " + s.builder.theme.Spacing["xl"]
		baseStyles["font-size"] = s.builder.theme.Typography["lg"].FontSize
		baseStyles["border-radius"] = s.builder.theme.BorderRadius["lg"]
	case "xl":
		baseStyles["padding"] = s.builder.theme.Spacing["xl"] + " " + s.builder.theme.Spacing["2xl"]
		baseStyles["font-size"] = s.builder.theme.Typography["xl"].FontSize
		baseStyles["border-radius"] = s.builder.theme.BorderRadius["xl"]
	default: // md
		baseStyles["padding"] = s.builder.theme.Spacing["md"] + " " + s.builder.theme.Spacing["lg"]
		baseStyles["font-size"] = s.builder.theme.Typography["md"].FontSize
		baseStyles["border-radius"] = s.builder.theme.BorderRadius["md"]
		baseStyles["line-height"] = s.builder.theme.Typography["md"].LineHeight
	}
	
	s.builder.GetGenerator().AddRule("."+className, baseStyles)
	
	// Focus state
	focusStyles := map[string]string{
		"outline":        "2px solid " + s.builder.theme.Colors["primary-500"],
		"outline-offset": "2px",
	}
	s.builder.GetGenerator().AddRuleWithPseudo("."+className, ":focus", focusStyles)
}

// generateSchemaInputCSS generates input CSS using schema-based class name
func (s *SchemaToCSS) generateSchemaInputCSS(style *ComponentStyle, options ComponentCSSOptions) {
	className := strings.TrimPrefix(options.CustomClassName, ".")
	
	// Base input styles
	baseStyles := map[string]string{
		"display":          "block",
		"width":            "100%",
		"border":           "1px solid " + s.builder.theme.Colors["neutral-300"],
		"border-radius":    s.builder.theme.BorderRadius["md"],
		"background-color": "#ffffff",
		"transition":       "all 150ms cubic-bezier(0.4, 0, 0.2, 1)",
		"outline":          "none",
	}
	
	// Size-specific styles
	switch options.Size {
	case "sm":
		baseStyles["padding"] = s.builder.theme.Spacing["sm"]
		baseStyles["font-size"] = s.builder.theme.Typography["sm"].FontSize
	case "lg":
		baseStyles["padding"] = s.builder.theme.Spacing["lg"]
		baseStyles["font-size"] = s.builder.theme.Typography["lg"].FontSize
	default: // md
		baseStyles["padding"] = s.builder.theme.Spacing["md"]
		baseStyles["font-size"] = s.builder.theme.Typography["md"].FontSize
	}
	
	s.builder.GetGenerator().AddRule("."+className, baseStyles)
	
	// Focus state
	focusStyles := map[string]string{
		"border-color": s.builder.theme.Colors["primary-500"],
		"box-shadow":   "0 0 0 1px " + s.builder.theme.Colors["primary-500"],
	}
	s.builder.GetGenerator().AddRuleWithPseudo("."+className, ":focus", focusStyles)
}

// generateSchemaCardCSS generates card CSS using schema-based class name
func (s *SchemaToCSS) generateSchemaCardCSS(style *ComponentStyle, options ComponentCSSOptions) {
	className := strings.TrimPrefix(options.CustomClassName, ".")
	
	// Base card styles
	baseStyles := map[string]string{
		"display":          "block",
		"background-color": "#ffffff",
		"border-radius":    s.builder.theme.BorderRadius["lg"],
		"border":           "1px solid " + s.builder.theme.Colors["neutral-200"],
		"overflow":         "hidden",
		"transition":       "all 150ms cubic-bezier(0.4, 0, 0.2, 1)",
	}
	
	// Variant-specific styles
	switch options.Variant {
	case "elevated":
		baseStyles["box-shadow"] = s.builder.theme.Shadows["md"]
	case "outlined":
		baseStyles["border-width"] = "2px"
	}
	
	s.builder.GetGenerator().AddRule("."+className, baseStyles)
}

// hasStateClass checks if a style has a specific state class
func (s *SchemaToCSS) hasStateClass(style *ComponentStyle, state string) bool {
	for _, class := range style.StateClasses {
		if class == state {
			return true
		}
	}
	return false
}

// GenerateComponentLibraryCSS generates CSS for an entire set of component schemas
func (s *SchemaToCSS) GenerateComponentLibraryCSS(schemas map[string]interface{}) (string, error) {
	// Reset builder for fresh generation
	s.builder = NewComponentCSSBuilder()
	if s.config.CustomTheme != nil {
		s.builder = NewComponentCSSBuilderWithTheme(s.config.CustomTheme)
	}
	
	// Process each schema
	var errors []string
	for name, schema := range schemas {
		_, err := s.GenerateCSSFromSchema(schema)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", name, err))
		}
	}
	
	// Generate utility classes once
	if s.config.GenerateUtilities {
		s.builder.GenerateUtilityCSS()
	}
	
	// Add theme variables once
	if s.config.IncludeThemeVariables {
		s.builder.AddThemeVariables()
	}
	
	if len(errors) > 0 {
		return s.builder.Generate(), fmt.Errorf("errors processing schemas: %s", strings.Join(errors, "; "))
	}
	
	return s.builder.Generate(), nil
}

// GenerateTypedComponentCSS provides type-safe CSS generation for specific schema types
type TypedComponentCSSGenerator struct {
	schemaToCSS *SchemaToCSS
}

// NewTypedComponentCSSGenerator creates a type-safe CSS generator
func NewTypedComponentCSSGenerator() *TypedComponentCSSGenerator {
	return &TypedComponentCSSGenerator{
		schemaToCSS: NewSchemaToCSS(),
	}
}

// GenerateButtonCSS generates CSS specifically for button schemas
func (g *TypedComponentCSSGenerator) GenerateButtonCSS(schema interface{}) (string, error) {
	// Extract button-specific properties using reflection
	style, err := g.schemaToCSS.ExtractStyleFromSchema(schema)
	if err != nil {
		return "", err
	}
	
	// Override type to ensure proper button handling
	style.Type = "button"
	
	return g.schemaToCSS.GenerateCSSFromStyle(style)
}

// GenerateFormCSS generates CSS specifically for form schemas
func (g *TypedComponentCSSGenerator) GenerateFormCSS(schema interface{}) (string, error) {
	style, err := g.schemaToCSS.ExtractStyleFromSchema(schema)
	if err != nil {
		return "", err
	}
	
	style.Type = "form"
	
	// Form-specific CSS generation
	className := g.schemaToCSS.buildClassName(style)
	
	formStyles := map[string]string{
		"display":        "block",
		"max-width":      "100%",
		"margin-bottom":  "1rem",
	}
	
	g.schemaToCSS.builder.GetGenerator().AddRule(className, formStyles)
	
	// Generate form control spacing
	controlSpacing := map[string]string{
		"margin-bottom": "1rem",
	}
	g.schemaToCSS.builder.GetGenerator().AddRule(className+" .form-control", controlSpacing)
	
	return g.schemaToCSS.builder.Generate(), nil
}

// GenerateTableCSS generates CSS specifically for table schemas
func (g *TypedComponentCSSGenerator) GenerateTableCSS(schema interface{}) (string, error) {
	style, err := g.schemaToCSS.ExtractStyleFromSchema(schema)
	if err != nil {
		return "", err
	}
	
	style.Type = "table"
	
	// Table-specific CSS generation
	className := g.schemaToCSS.buildClassName(style)
	
	tableStyles := map[string]string{
		"width":           "100%",
		"border-collapse": "collapse",
		"border-spacing":  "0",
		"background-color": "#ffffff",
	}
	
	g.schemaToCSS.builder.GetGenerator().AddRule(className, tableStyles)
	
	// Table header styles
	headerStyles := map[string]string{
		"background-color": "#f8f9fa",
		"font-weight":      "500",
		"padding":          "0.75rem",
		"border-bottom":    "1px solid #dee2e6",
	}
	g.schemaToCSS.builder.GetGenerator().AddRule(className+" th", headerStyles)
	
	// Table cell styles
	cellStyles := map[string]string{
		"padding":      "0.75rem",
		"border-bottom": "1px solid #dee2e6",
	}
	g.schemaToCSS.builder.GetGenerator().AddRule(className+" td", cellStyles)
	
	return g.schemaToCSS.builder.Generate(), nil
}

// GetBuilder returns the underlying CSS builder for advanced operations
func (s *SchemaToCSS) GetBuilder() *ComponentCSSBuilder {
	return s.builder
}

// Reset clears the current CSS generation state
func (s *SchemaToCSS) Reset() *SchemaToCSS {
	s.builder = NewComponentCSSBuilder()
	if s.config.CustomTheme != nil {
		s.builder = NewComponentCSSBuilderWithTheme(s.config.CustomTheme)
	}
	return s
}