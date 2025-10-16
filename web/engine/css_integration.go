package engine

import (
	"fmt"
	"strings"

	"github.com/niiniyare/erp/pkg/schema/ui/css"
)

// CSSIntegrationEngine integrates the sophisticated CSS runtime from pkg/schema/ui/css
// with the schema-driven component system
type CSSIntegrationEngine struct {
	cssFactory   *css.Factory
	cssRuntime   *css.Runtime
	validator    *css.Validator
	integration  *css.SchemaIntegration
}

// NewCSSIntegrationEngine creates a new CSS integration engine
func NewCSSIntegrationEngine(schemaDir string) (*CSSIntegrationEngine, error) {
	cssFactory := css.NewFactory(schemaDir)
	cssRuntime := css.NewRuntime()
	validator := css.NewValidator()
	integration := css.NewSchemaIntegration(schemaDir)

	return &CSSIntegrationEngine{
		cssFactory:  cssFactory,
		cssRuntime:  cssRuntime,
		validator:   validator,
		integration: integration,
	}, nil
}

// ApplyStylesToComponent applies advanced CSS styling to a component
func (cie *CSSIntegrationEngine) ApplyStylesToComponent(component *TemplComponent, styleProps map[string]interface{}) error {
	// Convert style props to css.Styles
	styles, err := cie.convertPropsToStyles(styleProps)
	if err != nil {
		return fmt.Errorf("failed to convert props to styles: %w", err)
	}

	// Generate CSS classes using the factory
	cssClasses, err := cie.cssFactory.GenerateCSS(*styles)
	if err != nil {
		return fmt.Errorf("failed to generate CSS: %w", err)
	}

	// Apply CSS to component
	if component.CSS != "" {
		component.CSS = cie.mergeCSS(component.CSS, cssClasses)
	} else {
		component.CSS = cssClasses
	}

	return nil
}

// convertPropsToStyles converts style properties to css.Styles format
func (cie *CSSIntegrationEngine) convertPropsToStyles(styleProps map[string]interface{}) (*css.Styles, error) {
	styles := &css.Styles{}

	// Color properties
	if color, ok := styleProps["color"].(string); ok {
		styles.Color = color
	}
	if backgroundColor, ok := styleProps["backgroundColor"].(string); ok {
		styles.BackgroundColor = backgroundColor
	}
	if borderColor, ok := styleProps["borderColor"].(string); ok {
		styles.BorderColor = borderColor
	}

	// Typography
	if fontSize, ok := styleProps["fontSize"].(string); ok {
		styles.FontSize = fontSize
	}
	if fontWeight, ok := styleProps["fontWeight"].(string); ok {
		styles.FontWeight = fontWeight
	}
	if fontFamily, ok := styleProps["fontFamily"].(string); ok {
		styles.FontFamily = fontFamily
	}
	if lineHeight, ok := styleProps["lineHeight"].(string); ok {
		styles.LineHeight = lineHeight
	}
	if textAlign, ok := styleProps["textAlign"].(string); ok {
		styles.TextAlign = textAlign
	}

	// Spacing
	if margin, ok := styleProps["margin"].(string); ok {
		styles.Margin = margin
	}
	if padding, ok := styleProps["padding"].(string); ok {
		styles.Padding = padding
	}
	if marginTop, ok := styleProps["marginTop"].(string); ok {
		styles.MarginTop = marginTop
	}
	if marginBottom, ok := styleProps["marginBottom"].(string); ok {
		styles.MarginBottom = marginBottom
	}
	if paddingX, ok := styleProps["paddingX"].(string); ok {
		styles.PaddingLeft = paddingX
		styles.PaddingRight = paddingX
	}
	if paddingY, ok := styleProps["paddingY"].(string); ok {
		styles.PaddingTop = paddingY
		styles.PaddingBottom = paddingY
	}

	// Layout
	if display, ok := styleProps["display"].(string); ok {
		styles.Display = display
	}
	if position, ok := styleProps["position"].(string); ok {
		styles.Position = position
	}
	if width, ok := styleProps["width"].(string); ok {
		styles.Width = width
	}
	if height, ok := styleProps["height"].(string); ok {
		styles.Height = height
	}
	if minWidth, ok := styleProps["minWidth"].(string); ok {
		styles.MinWidth = minWidth
	}
	if maxWidth, ok := styleProps["maxWidth"].(string); ok {
		styles.MaxWidth = maxWidth
	}

	// Flexbox
	if flexDirection, ok := styleProps["flexDirection"].(string); ok {
		styles.FlexDirection = flexDirection
	}
	if justifyContent, ok := styleProps["justifyContent"].(string); ok {
		styles.JustifyContent = justifyContent
	}
	if alignItems, ok := styleProps["alignItems"].(string); ok {
		styles.AlignItems = alignItems
	}
	if flexWrap, ok := styleProps["flexWrap"].(string); ok {
		styles.FlexWrap = flexWrap
	}
	if gap, ok := styleProps["gap"].(string); ok {
		styles.Gap = gap
	}

	// Grid
	if gridTemplateColumns, ok := styleProps["gridTemplateColumns"].(string); ok {
		styles.GridTemplateColumns = gridTemplateColumns
	}
	if gridTemplateRows, ok := styleProps["gridTemplateRows"].(string); ok {
		styles.GridTemplateRows = gridTemplateRows
	}
	if gridGap, ok := styleProps["gridGap"].(string); ok {
		styles.GridGap = gridGap
	}

	// Border
	if border, ok := styleProps["border"].(string); ok {
		styles.Border = border
	}
	if borderWidth, ok := styleProps["borderWidth"].(string); ok {
		styles.BorderWidth = borderWidth
	}
	if borderStyle, ok := styleProps["borderStyle"].(string); ok {
		styles.BorderStyle = borderStyle
	}
	if borderRadius, ok := styleProps["borderRadius"].(string); ok {
		styles.BorderRadius = borderRadius
	}

	// Effects
	if boxShadow, ok := styleProps["boxShadow"].(string); ok {
		styles.BoxShadow = boxShadow
	}
	if opacity, ok := styleProps["opacity"].(string); ok {
		styles.Opacity = opacity
	}
	if transform, ok := styleProps["transform"].(string); ok {
		styles.Transform = transform
	}
	if transition, ok := styleProps["transition"].(string); ok {
		styles.Transition = transition
	}

	// Animation
	if animation, ok := styleProps["animation"].(string); ok {
		styles.Animation = animation
	}

	// Responsive breakpoints
	if sm, ok := styleProps["sm"].(map[string]interface{}); ok {
		smStyles, err := cie.convertPropsToStyles(sm)
		if err == nil {
			styles.SM = smStyles
		}
	}
	if md, ok := styleProps["md"].(map[string]interface{}); ok {
		mdStyles, err := cie.convertPropsToStyles(md)
		if err == nil {
			styles.MD = mdStyles
		}
	}
	if lg, ok := styleProps["lg"].(map[string]interface{}); ok {
		lgStyles, err := cie.convertPropsToStyles(lg)
		if err == nil {
			styles.LG = lgStyles
		}
	}
	if xl, ok := styleProps["xl"].(map[string]interface{}); ok {
		xlStyles, err := cie.convertPropsToStyles(xl)
		if err == nil {
			styles.XL = xlStyles
		}
	}

	return styles, nil
}

// mergeCSS merges two CSS class strings
func (cie *CSSIntegrationEngine) mergeCSS(existing, new string) string {
	if existing == "" {
		return new
	}
	if new == "" {
		return existing
	}

	// Split into individual classes and deduplicate
	existingClasses := strings.Fields(existing)
	newClasses := strings.Fields(new)

	classSet := make(map[string]bool)
	var result []string

	// Add existing classes
	for _, class := range existingClasses {
		if !classSet[class] {
			classSet[class] = true
			result = append(result, class)
		}
	}

	// Add new classes
	for _, class := range newClasses {
		if !classSet[class] {
			classSet[class] = true
			result = append(result, class)
		}
	}

	return strings.Join(result, " ")
}

// ValidateStyles validates CSS styles against the schema system
func (cie *CSSIntegrationEngine) ValidateStyles(styles *css.Styles) error {
	return cie.validator.ValidateStyles(styles)
}

// GenerateResponsiveCSS generates responsive CSS for different breakpoints
func (cie *CSSIntegrationEngine) GenerateResponsiveCSS(styles map[string]*css.Styles) (string, error) {
	var cssBuilder strings.Builder

	// Base styles
	if baseStyles, exists := styles["base"]; exists {
		baseCss, err := cie.cssFactory.GenerateCSS(*baseStyles)
		if err != nil {
			return "", fmt.Errorf("failed to generate base CSS: %w", err)
		}
		cssBuilder.WriteString(baseCss)
	}

	// Responsive styles
	breakpoints := map[string]string{
		"sm": "@media (min-width: 640px)",
		"md": "@media (min-width: 768px)",
		"lg": "@media (min-width: 1024px)",
		"xl": "@media (min-width: 1280px)",
	}

	for breakpoint, mediaQuery := range breakpoints {
		if breakpointStyles, exists := styles[breakpoint]; exists {
			breakpointCss, err := cie.cssFactory.GenerateCSS(*breakpointStyles)
			if err != nil {
				return "", fmt.Errorf("failed to generate %s CSS: %w", breakpoint, err)
			}

			if breakpointCss != "" {
				cssBuilder.WriteString(fmt.Sprintf("\n%s {\n  %s\n}", mediaQuery, breakpointCss))
			}
		}
	}

	return cssBuilder.String(), nil
}

// ApplyThemeStyles applies theme-specific styling
func (cie *CSSIntegrationEngine) ApplyThemeStyles(component *TemplComponent, theme string) error {
	themeStyles := cie.getThemeStyles(theme)
	
	// Apply theme classes
	if component.CSS != "" {
		component.CSS = cie.mergeCSS(component.CSS, themeStyles)
	} else {
		component.CSS = themeStyles
	}

	return nil
}

// getThemeStyles returns CSS classes for a specific theme
func (cie *CSSIntegrationEngine) getThemeStyles(theme string) string {
	switch theme {
	case "dark":
		return "dark-theme bg-gray-900 text-white"
	case "light":
		return "light-theme bg-white text-gray-900"
	case "system":
		return "system-theme"
	default:
		return "default-theme"
	}
}

// ExtractStylesFromSchema extracts style information from JSON schema
func (cie *CSSIntegrationEngine) ExtractStylesFromSchema(schema *JsonSchema) (map[string]interface{}, error) {
	styles := make(map[string]interface{})

	// Look for style-related properties in the schema
	if styleProps, exists := schema.Properties["style"]; exists {
		if styleProps.Type == "object" {
			styles["inline"] = true
		}
	}

	if classProps, exists := schema.Properties["className"]; exists {
		if classProps.Type == "string" {
			styles["classes"] = true
		}
	}

	// Check for variant and size properties
	if variantProps, exists := schema.Properties["variant"]; exists {
		if len(variantProps.Enum) > 0 {
			styles["variants"] = variantProps.Enum
		}
	}

	if sizeProps, exists := schema.Properties["size"]; exists {
		if len(sizeProps.Enum) > 0 {
			styles["sizes"] = sizeProps.Enum
		}
	}

	return styles, nil
}

// GenerateUtilityClasses generates utility CSS classes for a component
func (cie *CSSIntegrationEngine) GenerateUtilityClasses(componentType string, variants []string) string {
	var classes []string

	// Base component class
	classes = append(classes, fmt.Sprintf("component-%s", componentType))

	// Variant classes
	for _, variant := range variants {
		classes = append(classes, fmt.Sprintf("component-%s-%s", componentType, variant))
	}

	// Common utility classes
	classes = append(classes, "transition-colors", "duration-200", "ease-in-out")

	return strings.Join(classes, " ")
}