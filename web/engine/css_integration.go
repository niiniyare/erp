package engine

import (
	"fmt"
	"strings"
	"sync"

	"github.com/niiniyare/erp/pkg/schema/ui/css"
)

// StyleProps represents strongly-typed style properties
type StyleProps struct {
	// Colors
	Color           string
	BackgroundColor string
	BorderColor     string

	// Typography
	FontSize   string
	FontWeight string
	FontFamily string
	LineHeight string
	TextAlign  string

	// Spacing
	Margin        string
	Padding       string
	MarginTop     string
	MarginBottom  string
	MarginLeft    string
	MarginRight   string
	PaddingTop    string
	PaddingBottom string
	PaddingLeft   string
	PaddingRight  string

	// Layout
	Display   string
	Position  string
	Width     string
	Height    string
	MinWidth  string
	MaxWidth  string
	MinHeight string
	MaxHeight string

	// Flexbox
	FlexDirection  string
	JustifyContent string
	AlignItems     string
	FlexWrap       string
	Gap            string
	Flex           string

	// Grid
	GridTemplateColumns string
	GridTemplateRows    string
	GridGap             string
	GridColumn          string
	GridRow             string

	// Border
	Border       string
	BorderWidth  string
	BorderStyle  string
	BorderRadius string

	// Effects
	BoxShadow  string
	Opacity    string
	Transform  string
	Transition string
	Animation  string

	// Responsive breakpoints
	SM *StyleProps
	MD *StyleProps
	LG *StyleProps
	XL *StyleProps
}

// Theme represents a visual theme configuration
type Theme struct {
	Name       string
	Classes    string
	Properties map[string]string
}

// ThemeRegistry manages available themes
type ThemeRegistry struct {
	mu     sync.RWMutex
	themes map[string]*Theme
}

// NewThemeRegistry creates a new theme registry with default themes
func NewThemeRegistry() *ThemeRegistry {
	tr := &ThemeRegistry{
		themes: make(map[string]*Theme),
	}

	// Register default themes
	tr.Register(&Theme{
		Name:    "dark",
		Classes: "dark-theme bg-gray-900 text-white",
		Properties: map[string]string{
			"backgroundColor": "#1a202c",
			"color":           "#ffffff",
		},
	})

	tr.Register(&Theme{
		Name:    "light",
		Classes: "light-theme bg-white text-gray-900",
		Properties: map[string]string{
			"backgroundColor": "#ffffff",
			"color":           "#1a202c",
		},
	})

	tr.Register(&Theme{
		Name:    "system",
		Classes: "system-theme",
	})

	return tr
}

// Register adds a new theme to the registry
func (tr *ThemeRegistry) Register(theme *Theme) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.themes[theme.Name] = theme
}

// Get retrieves a theme by name
func (tr *ThemeRegistry) Get(name string) (*Theme, bool) {
	tr.mu.RLock()
	defer tr.mu.RUnlock()
	theme, exists := tr.themes[name]
	return theme, exists
}

// CSSIntegrationEngine integrates the sophisticated CSS runtime with the schema-driven component system
type CSSIntegrationEngine struct {
	cssFactory    *css.Factory
	validator     *css.Validator
	integration   *css.SchemaToCSS
	themeRegistry *ThemeRegistry
	mu            sync.RWMutex
}

// CSSEngineConfig holds configuration for the CSS engine
type CSSEngineConfig struct {
	SchemaDir        string
	EnableValidation bool
	ThemeRegistry    *ThemeRegistry
}

// NewCSSIntegrationEngine creates a new CSS integration engine with configuration
func NewCSSIntegrationEngine(config CSSEngineConfig) (*CSSIntegrationEngine, error) {
	if config.SchemaDir == "" {
		return nil, fmt.Errorf("schema directory is required")
	}

	cssFactory := css.NewFactory(config.SchemaDir)
	validator := css.NewValidator(config.SchemaDir)
	// Note: Using SchemaToCSS instead of SchemaIntegration
	integration := &css.SchemaToCSS{}

	themeRegistry := config.ThemeRegistry
	if themeRegistry == nil {
		themeRegistry = NewThemeRegistry()
	}

	return &CSSIntegrationEngine{
		cssFactory:    cssFactory,
		validator:     validator,
		integration:   integration,
		themeRegistry: themeRegistry,
	}, nil
}

// ApplyStylesToComponent applies advanced CSS styling to a component with validation
func (cie *CSSIntegrationEngine) ApplyStylesToComponent(component *TemplComponent, styleProps *StyleProps, validate bool) error {
	if component == nil {
		return fmt.Errorf("component cannot be nil")
	}
	if styleProps == nil {
		return fmt.Errorf("style props cannot be nil")
	}

	// Convert style props to css.Styles
	styles := cie.convertPropsToStyles(styleProps)

	// Validate if requested
	if validate {
		if err := cie.validator.ValidateStyles(styles); err != nil {
			return fmt.Errorf("style validation failed: %w", err)
		}
	}

	// Generate CSS classes using the factory
	cssClasses, err := cie.cssFactory.GenerateCSS(*styles)
	if err != nil {
		return fmt.Errorf("failed to generate CSS: %w", err)
	}

	// Apply CSS to component
	cie.mu.Lock()
	defer cie.mu.Unlock()

	component.CSS = cie.mergeCSS(component.CSS, cssClasses)

	return nil
}

// convertPropsToStyles converts StyleProps to css.Styles format
func (cie *CSSIntegrationEngine) convertPropsToStyles(styleProps *StyleProps) *css.Styles {
	if styleProps == nil {
		return css.NewStyles("")
	}

	styles := css.NewStyles(cie.cssFactory.GetSchemaDir())

	// Set layout and box model properties
	if styleProps.Display != "" {
		styles.Display = styleProps.Display
	}
	if styleProps.Position != "" {
		styles.Position = styleProps.Position
	}
	if styleProps.Width != "" {
		styles.Width = styleProps.Width
	}
	if styleProps.Height != "" {
		styles.Height = styleProps.Height
	}
	if styleProps.MinWidth != "" {
		styles.MinWidth = styleProps.MinWidth
	}
	if styleProps.MaxWidth != "" {
		styles.MaxWidth = styleProps.MaxWidth
	}
	if styleProps.MinHeight != "" {
		styles.MinHeight = styleProps.MinHeight
	}
	if styleProps.MaxHeight != "" {
		styles.MaxHeight = styleProps.MaxHeight
	}

	// Set typography properties
	if styleProps.FontSize != "" {
		styles.FontSize = styleProps.FontSize
	}
	if styleProps.FontWeight != "" {
		styles.FontWeight = styleProps.FontWeight
	}
	if styleProps.FontFamily != "" {
		styles.FontFamily = styleProps.FontFamily
	}
	if styleProps.LineHeight != "" {
		styles.LineHeight = styleProps.LineHeight
	}
	if styleProps.TextAlign != "" {
		styles.TextAlign = styleProps.TextAlign
	}
	if styleProps.Color != "" {
		styles.Color = styleProps.Color
	}

	// Set background properties
	if styleProps.BackgroundColor != "" {
		styles.BackgroundColor = styleProps.BackgroundColor
	}

	// Set margin properties
	if styleProps.Margin != "" {
		styles.Margin = styleProps.Margin
	}
	if styleProps.MarginTop != "" {
		styles.MarginTop = styleProps.MarginTop
	}
	if styleProps.MarginBottom != "" {
		styles.MarginBottom = styleProps.MarginBottom
	}
	if styleProps.MarginLeft != "" {
		styles.MarginLeft = styleProps.MarginLeft
	}
	if styleProps.MarginRight != "" {
		styles.MarginRight = styleProps.MarginRight
	}

	// Set padding properties
	if styleProps.Padding != "" {
		styles.Padding = styleProps.Padding
	}
	if styleProps.PaddingTop != "" {
		styles.PaddingTop = styleProps.PaddingTop
	}
	if styleProps.PaddingBottom != "" {
		styles.PaddingBottom = styleProps.PaddingBottom
	}
	if styleProps.PaddingLeft != "" {
		styles.PaddingLeft = styleProps.PaddingLeft
	}
	if styleProps.PaddingRight != "" {
		styles.PaddingRight = styleProps.PaddingRight
	}

	// Set border properties
	if styleProps.Border != "" {
		styles.Border = styleProps.Border
	}
	if styleProps.BorderWidth != "" {
		styles.BorderWidth = styleProps.BorderWidth
	}
	if styleProps.BorderStyle != "" {
		styles.BorderStyle = styleProps.BorderStyle
	}
	if styleProps.BorderRadius != "" {
		styles.BorderRadius = styleProps.BorderRadius
	}
	if styleProps.BorderColor != "" {
		styles.BorderColor = styleProps.BorderColor
	}

	// Set flexbox properties
	if styleProps.FlexDirection != "" {
		styles.FlexDirection = styleProps.FlexDirection
	}
	if styleProps.JustifyContent != "" {
		styles.JustifyContent = styleProps.JustifyContent
	}
	if styleProps.AlignItems != "" {
		styles.AlignItems = styleProps.AlignItems
	}
	if styleProps.FlexWrap != "" {
		styles.FlexWrap = styleProps.FlexWrap
	}
	if styleProps.Flex != "" {
		styles.Flex = styleProps.Flex
	}

	// Set grid properties
	if styleProps.GridTemplateColumns != "" {
		styles.GridTemplateColumns = styleProps.GridTemplateColumns
	}
	if styleProps.GridTemplateRows != "" {
		styles.GridTemplateRows = styleProps.GridTemplateRows
	}
	if styleProps.GridGap != "" {
		styles.GridGap = styleProps.GridGap
	}

	// Set visual effects
	if styleProps.BoxShadow != "" {
		styles.BoxShadow = styleProps.BoxShadow
	}
	if styleProps.Opacity != "" {
		styles.Opacity = styleProps.Opacity
	}
	if styleProps.Transform != "" {
		styles.Transform = styleProps.Transform
	}

	// Set animation properties
	if styleProps.Animation != "" {
		styles.Animation = styleProps.Animation
	}

	// Set transition properties
	if styleProps.Transition != "" {
		styles.Transition = styleProps.Transition
	}

	return styles
}

// ParseStylePropsFromMap converts map[string]interface{} to strongly-typed StyleProps
func ParseStylePropsFromMap(styleMap map[string]interface{}) (*StyleProps, error) {
	if styleMap == nil {
		return &StyleProps{}, nil
	}

	props := &StyleProps{}

	// Helper function to safely extract string values
	getString := func(key string) string {
		if val, ok := styleMap[key].(string); ok {
			return val
		}
		return ""
	}

	// Colors
	props.Color = getString("color")
	props.BackgroundColor = getString("backgroundColor")
	props.BorderColor = getString("borderColor")

	// Typography
	props.FontSize = getString("fontSize")
	props.FontWeight = getString("fontWeight")
	props.FontFamily = getString("fontFamily")
	props.LineHeight = getString("lineHeight")
	props.TextAlign = getString("textAlign")

	// Spacing
	props.Margin = getString("margin")
	props.Padding = getString("padding")
	props.MarginTop = getString("marginTop")
	props.MarginBottom = getString("marginBottom")
	props.MarginLeft = getString("marginLeft")
	props.MarginRight = getString("marginRight")
	props.PaddingTop = getString("paddingTop")
	props.PaddingBottom = getString("paddingBottom")
	props.PaddingLeft = getString("paddingLeft")
	props.PaddingRight = getString("paddingRight")

	// Handle shorthand properties
	if paddingX := getString("paddingX"); paddingX != "" {
		props.PaddingLeft = paddingX
		props.PaddingRight = paddingX
	}
	if paddingY := getString("paddingY"); paddingY != "" {
		props.PaddingTop = paddingY
		props.PaddingBottom = paddingY
	}

	// Layout
	props.Display = getString("display")
	props.Position = getString("position")
	props.Width = getString("width")
	props.Height = getString("height")
	props.MinWidth = getString("minWidth")
	props.MaxWidth = getString("maxWidth")

	// Flexbox
	props.FlexDirection = getString("flexDirection")
	props.JustifyContent = getString("justifyContent")
	props.AlignItems = getString("alignItems")
	props.FlexWrap = getString("flexWrap")
	props.Gap = getString("gap")

	// Grid
	props.GridTemplateColumns = getString("gridTemplateColumns")
	props.GridTemplateRows = getString("gridTemplateRows")
	props.GridGap = getString("gridGap")

	// Border
	props.Border = getString("border")
	props.BorderWidth = getString("borderWidth")
	props.BorderStyle = getString("borderStyle")
	props.BorderRadius = getString("borderRadius")

	// Effects
	props.BoxShadow = getString("boxShadow")
	props.Opacity = getString("opacity")
	props.Transform = getString("transform")
	props.Transition = getString("transition")
	props.Animation = getString("animation")

	// Responsive breakpoints
	breakpoints := []struct {
		key   string
		field **StyleProps
	}{
		{"sm", &props.SM},
		{"md", &props.MD},
		{"lg", &props.LG},
		{"xl", &props.XL},
	}

	for _, bp := range breakpoints {
		if bpMap, ok := styleMap[bp.key].(map[string]interface{}); ok {
			bpProps, err := ParseStylePropsFromMap(bpMap)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %s breakpoint: %w", bp.key, err)
			}
			*bp.field = bpProps
		}
	}

	return props, nil
}

// mergeCSS merges two CSS class strings and deduplicates
func (cie *CSSIntegrationEngine) mergeCSS(existing, new string) string {
	if existing == "" {
		return new
	}
	if new == "" {
		return existing
	}

	// Use map for O(1) lookups
	classSet := make(map[string]bool)
	result := make([]string, 0, strings.Count(existing, " ")+strings.Count(new, " ")+2)

	// Add existing classes
	for _, class := range strings.Fields(existing) {
		if class != "" && !classSet[class] {
			classSet[class] = true
			result = append(result, class)
		}
	}

	// Add new classes
	for _, class := range strings.Fields(new) {
		if class != "" && !classSet[class] {
			classSet[class] = true
			result = append(result, class)
		}
	}

	return strings.Join(result, " ")
}

// ValidateStyles validates CSS styles
func (cie *CSSIntegrationEngine) ValidateStyles(styles *css.Styles) error {
	if styles == nil {
		return fmt.Errorf("styles cannot be nil")
	}

	// The validator returns []ValidationError, convert to a single error
	errors := cie.validator.ValidateStyles(styles)
	if len(errors) > 0 {
		// Combine all errors into one
		var messages []string
		for _, err := range errors {
			messages = append(messages, err.Error())
		}
		return fmt.Errorf("validation failed: %s", strings.Join(messages, "; "))
	}
	return nil
}

// GenerateResponsiveCSS generates complete responsive CSS with proper media queries
func (cie *CSSIntegrationEngine) GenerateResponsiveCSS(componentID string, styles *StyleProps) (string, error) {
	if componentID == "" {
		return "", fmt.Errorf("component ID is required")
	}
	if styles == nil {
		return "", fmt.Errorf("styles cannot be nil")
	}

	var cssBuilder strings.Builder

	// Base styles
	baseStyles := cie.convertPropsToStyles(styles)
	baseCss, err := cie.cssFactory.GenerateCSS(*baseStyles)
	if err != nil {
		return "", fmt.Errorf("failed to generate base CSS: %w", err)
	}

	if baseCss != "" {
		cssBuilder.WriteString(fmt.Sprintf(".%s { %s }\n", componentID, baseCss))
	}

	// Responsive breakpoints
	breakpoints := []struct {
		name       string
		mediaQuery string
		styles     *StyleProps
	}{
		{"sm", "@media (min-width: 640px)", styles.SM},
		{"md", "@media (min-width: 768px)", styles.MD},
		{"lg", "@media (min-width: 1024px)", styles.LG},
		{"xl", "@media (min-width: 1280px)", styles.XL},
	}

	for _, bp := range breakpoints {
		if bp.styles != nil {
			bpCssStyles := cie.convertPropsToStyles(bp.styles)
			bpCss, err := cie.cssFactory.GenerateCSS(*bpCssStyles)
			if err != nil {
				return "", fmt.Errorf("failed to generate %s CSS: %w", bp.name, err)
			}

			if bpCss != "" {
				cssBuilder.WriteString(fmt.Sprintf("%s {\n  .%s { %s }\n}\n", bp.mediaQuery, componentID, bpCss))
			}
		}
	}

	return cssBuilder.String(), nil
}

// ApplyTheme applies a registered theme to a component
func (cie *CSSIntegrationEngine) ApplyTheme(component *TemplComponent, themeName string) error {
	if component == nil {
		return fmt.Errorf("component cannot be nil")
	}

	theme, exists := cie.themeRegistry.Get(themeName)
	if !exists {
		return fmt.Errorf("theme '%s' not found", themeName)
	}

	cie.mu.Lock()
	defer cie.mu.Unlock()

	component.CSS = cie.mergeCSS(component.CSS, theme.Classes)

	return nil
}

// RegisterTheme registers a new theme
func (cie *CSSIntegrationEngine) RegisterTheme(theme *Theme) error {
	if theme == nil {
		return fmt.Errorf("theme cannot be nil")
	}
	if theme.Name == "" {
		return fmt.Errorf("theme name is required")
	}

	cie.themeRegistry.Register(theme)
	return nil
}

// ExtractStylesFromSchema extracts style information from JSON schema
func (cie *CSSIntegrationEngine) ExtractStylesFromSchema(schema *JsonSchema) (map[string]interface{}, error) {
	if schema == nil {
		return nil, fmt.Errorf("schema cannot be nil")
	}

	styles := make(map[string]interface{})

	// Check for style-related properties
	if styleProps, exists := schema.Properties["style"]; exists {
		if styleProps.Type == "object" {
			styles["supportsInlineStyles"] = true
		}
	}

	if classProps, exists := schema.Properties["className"]; exists {
		if classProps.Type == "string" {
			styles["supportsClasses"] = true
		}
	}

	// Extract variants
	if variantProps, exists := schema.Properties["variant"]; exists {
		if len(variantProps.Enum) > 0 {
			styles["variants"] = variantProps.Enum
		}
	}

	// Extract sizes
	if sizeProps, exists := schema.Properties["size"]; exists {
		if len(sizeProps.Enum) > 0 {
			styles["sizes"] = sizeProps.Enum
		}
	}

	return styles, nil
}

// GenerateUtilityClasses generates utility CSS classes for a component
func (cie *CSSIntegrationEngine) GenerateUtilityClasses(componentType string, options UtilityClassOptions) string {
	if componentType == "" {
		return ""
	}

	classes := make([]string, 0, len(options.Variants)+3)

	// Base component class
	classes = append(classes, fmt.Sprintf("component-%s", componentType))

	// Variant classes
	for _, variant := range options.Variants {
		if variant != "" {
			classes = append(classes, fmt.Sprintf("component-%s-%s", componentType, variant))
		}
	}

	// Include common utilities if requested
	if options.IncludeTransitions {
		classes = append(classes, "transition-colors", "duration-200", "ease-in-out")
	}

	if options.CustomClasses != "" {
		classes = append(classes, options.CustomClasses)
	}

	return strings.Join(classes, " ")
}

// UtilityClassOptions configures utility class generation
type UtilityClassOptions struct {
	Variants           []string
	IncludeTransitions bool
	CustomClasses      string
}
