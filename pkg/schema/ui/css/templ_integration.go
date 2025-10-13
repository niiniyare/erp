package css

import (
	"encoding/json"
	"fmt"
	"strings"
)

// TemplStyleGenerator generates styles optimized for Templ component integration
// This bridges the gap between pkg/schema/ui/css and web/components/**/*
type TemplStyleGenerator struct {
	schemaDir string
	factory   *Factory
	loader    *SchemaLoader
}

// NewTemplStyleGenerator creates a generator for Templ integration
func NewTemplStyleGenerator(schemaDir string) *TemplStyleGenerator {
	return &TemplStyleGenerator{
		schemaDir: schemaDir,
		factory:   NewFactory(schemaDir),
		loader:    NewSchemaLoader(schemaDir),
	}
}

// ComponentStyleConfig represents styling configuration for Templ components
type ComponentStyleConfig struct {
	ComponentType string            `json:"component_type"` // e.g., "atoms.button", "molecules.field"
	Variant       string            `json:"variant"`        // e.g., "primary", "secondary"
	Size          string            `json:"size"`           // e.g., "sm", "md", "lg"
	State         string            `json:"state"`          // e.g., "default", "hover", "disabled"
	Theme         string            `json:"theme"`          // e.g., "light", "dark"
	CustomProps   map[string]any    `json:"custom_props"`   // Additional properties
	Responsive    map[string]string `json:"responsive"`     // Responsive overrides
}

// GenerateForTemplComponent generates CSS specifically for a Templ component
func (tsg *TemplStyleGenerator) GenerateForTemplComponent(config ComponentStyleConfig) (*ExpandedStyles, error) {
	styles := NewExpandedStyles(tsg.schemaDir)

	// Apply base component styles
	if err := tsg.applyBaseComponentStyles(styles, config); err != nil {
		return nil, fmt.Errorf("failed to apply base styles: %w", err)
	}

	// Apply variant-specific styles
	if err := tsg.applyVariantStyles(styles, config); err != nil {
		return nil, fmt.Errorf("failed to apply variant styles: %w", err)
	}

	// Apply size-specific styles
	if err := tsg.applySizeStyles(styles, config); err != nil {
		return nil, fmt.Errorf("failed to apply size styles: %w", err)
	}

	// Apply state-specific styles
	if err := tsg.applyStateStyles(styles, config); err != nil {
		return nil, fmt.Errorf("failed to apply state styles: %w", err)
	}

	// Apply theme-specific styles
	if err := tsg.applyThemeStyles(styles, config); err != nil {
		return nil, fmt.Errorf("failed to apply theme styles: %w", err)
	}

	// Apply custom properties
	tsg.applyCustomProperties(styles, config)

	// Apply responsive styles
	tsg.applyResponsiveStyles(styles, config)

	return styles, nil
}

// GenerateTemplStyleAttribute creates a Templ-compatible style attribute
func (tsg *TemplStyleGenerator) GenerateTemplStyleAttribute(config ComponentStyleConfig) (string, error) {
	styles, err := tsg.GenerateForTemplComponent(config)
	if err != nil {
		return "", err
	}

	return styles.ToTemplStyleAttribute(), nil
}

// GenerateTemplClass creates a CSS class for use in Templ templates
func (tsg *TemplStyleGenerator) GenerateTemplClass(className string, config ComponentStyleConfig) (string, error) {
	styles, err := tsg.GenerateForTemplComponent(config)
	if err != nil {
		return "", err
	}

	return styles.ToCSSClass(className), nil
}

// Component-specific style generators

// ButtonStyles generates comprehensive button styles for Templ integration
func (tsg *TemplStyleGenerator) ButtonStyles(variant, size, state string) (*ExpandedStyles, error) {
	styles := NewExpandedStyles(tsg.schemaDir)

	// Base button styles
	styles.WithDisplay("inline-flex").
		WithAlignItems("center").
		WithJustifyContent("center").
		WithFontWeight("500").
		WithBorderRadius("0.375rem").
		WithTransition("all 0.2s ease-in-out").
		WithCustomProperty("user-select", "none").
		WithCustomProperty("outline", "none")

	// Variant-specific styles
	switch variant {
	case "primary":
		styles.WithBackgroundColor("#3b82f6").
			WithColor("#ffffff").
			WithCustomProperty("border", "1px solid #3b82f6").
			WithCustomProperty("hover-bg", "#2563eb").
			WithCustomProperty("active-bg", "#1d4ed8")

	case "secondary":
		styles.WithBackgroundColor("#6b7280").
			WithColor("#ffffff").
			WithCustomProperty("border", "1px solid #6b7280").
			WithCustomProperty("hover-bg", "#4b5563").
			WithCustomProperty("active-bg", "#374151")

	case "outline":
		styles.WithBackgroundColor("transparent").
			WithColor("#3b82f6").
			WithCustomProperty("border", "1px solid #3b82f6").
			WithCustomProperty("hover-bg", "#f0f9ff").
			WithCustomProperty("active-bg", "#dbeafe")

	case "danger":
		styles.WithBackgroundColor("#ef4444").
			WithColor("#ffffff").
			WithCustomProperty("border", "1px solid #ef4444").
			WithCustomProperty("hover-bg", "#dc2626").
			WithCustomProperty("active-bg", "#b91c1c")

	case "success":
		styles.WithBackgroundColor("#10b981").
			WithColor("#ffffff").
			WithCustomProperty("border", "1px solid #10b981").
			WithCustomProperty("hover-bg", "#059669").
			WithCustomProperty("active-bg", "#047857")
	}

	// Size-specific styles
	switch size {
	case "xs":
		styles.WithFontSize("0.75rem").
			WithPadding("0.25rem 0.5rem").
			WithCustomProperty("min-height", "1.5rem")

	case "sm":
		styles.WithFontSize("0.875rem").
			WithPadding("0.375rem 0.75rem").
			WithCustomProperty("min-height", "2rem")

	case "md", "":
		styles.WithFontSize("0.875rem").
			WithPadding("0.5rem 1rem").
			WithCustomProperty("min-height", "2.5rem")

	case "lg":
		styles.WithFontSize("1rem").
			WithPadding("0.75rem 1.5rem").
			WithCustomProperty("min-height", "3rem")

	case "xl":
		styles.WithFontSize("1.125rem").
			WithPadding("1rem 2rem").
			WithCustomProperty("min-height", "3.5rem")
	}

	// State-specific styles
	switch state {
	case "disabled":
		styles.WithOpacity("0.5").
			WithCustomProperty("cursor", "not-allowed").
			WithCustomProperty("pointer-events", "none")

	case "loading":
		styles.WithCustomProperty("cursor", "progress").
			WithCustomProperty("position", "relative")
	}

	return styles, nil
}

// InputStyles generates comprehensive input styles for Templ integration
func (tsg *TemplStyleGenerator) InputStyles(variant, size, state string) (*ExpandedStyles, error) {
	styles := NewExpandedStyles(tsg.schemaDir)

	// Base input styles
	styles.WithDisplay("block").
		WithWidth("100%").
		WithFontSize("0.875rem").
		WithLineHeight("1.5").
		WithColor("#111827").
		WithBorderRadius("0.375rem").
		WithTransition("all 0.2s ease-in-out").
		WithCustomProperty("background-image", "none").
		WithCustomProperty("outline", "none")

	// Variant-specific styles
	switch variant {
	case "default", "":
		styles.WithCustomProperty("border", "1px solid #d1d5db").
			WithBackgroundColor("#ffffff").
			WithCustomProperty("focus-border-color", "#3b82f6").
			WithCustomProperty("focus-ring", "0 0 0 3px rgba(59, 130, 246, 0.1)")

	case "error":
		styles.WithCustomProperty("border", "1px solid #ef4444").
			WithBackgroundColor("#ffffff").
			WithCustomProperty("focus-border-color", "#ef4444").
			WithCustomProperty("focus-ring", "0 0 0 3px rgba(239, 68, 68, 0.1)")

	case "success":
		styles.WithCustomProperty("border", "1px solid #10b981").
			WithBackgroundColor("#ffffff").
			WithCustomProperty("focus-border-color", "#10b981").
			WithCustomProperty("focus-ring", "0 0 0 3px rgba(16, 185, 129, 0.1)")
	}

	// Size-specific styles
	switch size {
	case "sm":
		styles.WithFontSize("0.75rem").
			WithPadding("0.375rem 0.75rem").
			WithCustomProperty("min-height", "2rem")

	case "md", "":
		styles.WithFontSize("0.875rem").
			WithPadding("0.5rem 0.75rem").
			WithCustomProperty("min-height", "2.5rem")

	case "lg":
		styles.WithFontSize("1rem").
			WithPadding("0.75rem 1rem").
			WithCustomProperty("min-height", "3rem")
	}

	// State-specific styles
	switch state {
	case "disabled":
		styles.WithOpacity("0.5").
			WithBackgroundColor("#f9fafb").
			WithCustomProperty("cursor", "not-allowed")

	case "readonly":
		styles.WithBackgroundColor("#f9fafb").
			WithCustomProperty("cursor", "default")
	}

	return styles, nil
}

// CardStyles generates comprehensive card styles for Templ integration
func (tsg *TemplStyleGenerator) CardStyles(elevation, padding string) (*ExpandedStyles, error) {
	styles := NewExpandedStyles(tsg.schemaDir)

	// Base card styles
	styles.WithDisplay("block").
		WithBackgroundColor("#ffffff").
		WithBorderRadius("0.5rem").
		WithCustomProperty("border", "1px solid #e5e7eb").
		WithCustomProperty("overflow", "hidden")

	// Elevation-specific shadows
	switch elevation {
	case "none":
		// No shadow

	case "xs":
		styles.WithBoxShadow("0 1px 2px 0 rgba(0, 0, 0, 0.05)")

	case "sm", "":
		styles.WithBoxShadow("0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06)")

	case "md":
		styles.WithBoxShadow("0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)")

	case "lg":
		styles.WithBoxShadow("0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05)")

	case "xl":
		styles.WithBoxShadow("0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04)")

	case "2xl":
		styles.WithBoxShadow("0 25px 50px -12px rgba(0, 0, 0, 0.25)")
	}

	// Padding options
	switch padding {
	case "none":
		// No padding

	case "sm":
		styles.WithPadding("1rem")

	case "md", "":
		styles.WithPadding("1.5rem")

	case "lg":
		styles.WithPadding("2rem")

	case "xl":
		styles.WithPadding("3rem")
	}

	return styles, nil
}

// ModalStyles generates comprehensive modal styles for Templ integration
func (tsg *TemplStyleGenerator) ModalStyles(size, position string) (*ExpandedStyles, error) {
	styles := NewExpandedStyles(tsg.schemaDir)

	// Base modal styles
	styles.WithPosition("fixed").
		WithZIndex("50").
		WithCustomProperty("inset", "0").
		WithDisplay("flex").
		WithAlignItems("center").
		WithJustifyContent("center").
		WithCustomProperty("background-color", "rgba(0, 0, 0, 0.5)").
		WithCustomProperty("backdrop-filter", "blur(4px)")

	// Apply modal content styles to main styles
	styles.WithCustomProperty("modal-bg", "#ffffff").
		WithCustomProperty("modal-border-radius", "0.5rem").
		WithCustomProperty("modal-shadow", "0 25px 50px -12px rgba(0, 0, 0, 0.25)").
		WithCustomProperty("modal-max-height", "90vh").
		WithCustomProperty("modal-overflow", "auto")

	// Size-specific styles (apply to main modal styles)
	switch size {
	case "xs":
		styles.WithCustomProperty("modal-max-width", "20rem")

	case "sm":
		styles.WithCustomProperty("modal-max-width", "24rem")

	case "md", "":
		styles.WithCustomProperty("modal-max-width", "32rem")

	case "lg":
		styles.WithCustomProperty("modal-max-width", "48rem")

	case "xl":
		styles.WithCustomProperty("modal-max-width", "64rem")

	case "full":
		styles.WithCustomProperty("modal-max-width", "95vw").
			WithCustomProperty("modal-max-height", "95vh")
	}

	return styles, nil
}

// Helper methods for style application

func (tsg *TemplStyleGenerator) applyBaseComponentStyles(styles *ExpandedStyles, config ComponentStyleConfig) error {
	componentType := strings.ToLower(config.ComponentType)

	// Apply component-specific base styles by calling the dedicated methods
	switch {
	case strings.Contains(componentType, "button"):
		buttonStyles, err := tsg.ButtonStyles(config.Variant, config.Size, config.State)
		if err != nil {
			return err
		}
		// Copy all properties from button styles
		tsg.mergeStyles(styles, buttonStyles)

	case strings.Contains(componentType, "input"):
		inputStyles, err := tsg.InputStyles(config.Variant, config.Size, config.State)
		if err != nil {
			return err
		}
		tsg.mergeStyles(styles, inputStyles)

	case strings.Contains(componentType, "card"):
		// Default card styles
		styles.WithDisplay("block").
			WithBackgroundColor("#ffffff").
			WithBorderRadius("0.5rem")

	case strings.Contains(componentType, "field"):
		// Default field styles
		styles.WithDisplay("block").
			WithWidth("100%").
			WithMargin("0.5rem 0")

	case strings.Contains(componentType, "modal"):
		modalStyles, err := tsg.ModalStyles(config.Size, "center")
		if err != nil {
			return err
		}
		tsg.mergeStyles(styles, modalStyles)

		// Add more component types as needed
	}

	return nil
}

// mergeStyles copies properties from source to destination using nested structure
func (tsg *TemplStyleGenerator) mergeStyles(dest, src *ExpandedStyles) {
	// Layout properties
	if src.Layout.Display != "" {
		dest.Layout.Display = src.Layout.Display
	}
	if src.Layout.Position != "" {
		dest.Layout.Position = src.Layout.Position
	}
	if src.Layout.Top != "" {
		dest.Layout.Top = src.Layout.Top
	}
	if src.Layout.Right != "" {
		dest.Layout.Right = src.Layout.Right
	}
	if src.Layout.Bottom != "" {
		dest.Layout.Bottom = src.Layout.Bottom
	}
	if src.Layout.Left != "" {
		dest.Layout.Left = src.Layout.Left
	}
	if src.Layout.ZIndex != "" {
		dest.Layout.ZIndex = src.Layout.ZIndex
	}

	// Box model properties
	if src.BoxModel.Width != "" {
		dest.BoxModel.Width = src.BoxModel.Width
	}
	if src.BoxModel.Height != "" {
		dest.BoxModel.Height = src.BoxModel.Height
	}
	if src.BoxModel.MinWidth != "" {
		dest.BoxModel.MinWidth = src.BoxModel.MinWidth
	}
	if src.BoxModel.MinHeight != "" {
		dest.BoxModel.MinHeight = src.BoxModel.MinHeight
	}
	if src.BoxModel.MaxWidth != "" {
		dest.BoxModel.MaxWidth = src.BoxModel.MaxWidth
	}
	if src.BoxModel.MaxHeight != "" {
		dest.BoxModel.MaxHeight = src.BoxModel.MaxHeight
	}
	if src.BoxModel.BoxSizing != "" {
		dest.BoxModel.BoxSizing = src.BoxModel.BoxSizing
	}

	// Margin properties
	if src.Margin.Margin != "" {
		dest.Margin.Margin = src.Margin.Margin
	}
	if src.Margin.Top != "" {
		dest.Margin.Top = src.Margin.Top
	}
	if src.Margin.Right != "" {
		dest.Margin.Right = src.Margin.Right
	}
	if src.Margin.Bottom != "" {
		dest.Margin.Bottom = src.Margin.Bottom
	}
	if src.Margin.Left != "" {
		dest.Margin.Left = src.Margin.Left
	}

	// Padding properties
	if src.Padding.Padding != "" {
		dest.Padding.Padding = src.Padding.Padding
	}
	if src.Padding.Top != "" {
		dest.Padding.Top = src.Padding.Top
	}
	if src.Padding.Right != "" {
		dest.Padding.Right = src.Padding.Right
	}
	if src.Padding.Bottom != "" {
		dest.Padding.Bottom = src.Padding.Bottom
	}
	if src.Padding.Left != "" {
		dest.Padding.Left = src.Padding.Left
	}

	// Border properties
	if src.Border.Border != "" {
		dest.Border.Border = src.Border.Border
	}
	if src.Border.Radius != "" {
		dest.Border.Radius = src.Border.Radius
	}
	if src.Border.Width != "" {
		dest.Border.Width = src.Border.Width
	}
	if src.Border.Style != "" {
		dest.Border.Style = src.Border.Style
	}
	if src.Border.Color != "" {
		dest.Border.Color = src.Border.Color
	}

	// Background properties
	if src.Background.Background != "" {
		dest.Background.Background = src.Background.Background
	}
	if src.Background.Color != "" {
		dest.Background.Color = src.Background.Color
	}
	if src.Background.Image != "" {
		dest.Background.Image = src.Background.Image
	}

	// Typography properties
	if src.Typography.Color != "" {
		dest.Typography.Color = src.Typography.Color
	}
	if src.Typography.FontSize != "" {
		dest.Typography.FontSize = src.Typography.FontSize
	}
	if src.Typography.FontWeight != "" {
		dest.Typography.FontWeight = src.Typography.FontWeight
	}
	if src.Typography.FontFamily != "" {
		dest.Typography.FontFamily = src.Typography.FontFamily
	}
	if src.Typography.LineHeight != "" {
		dest.Typography.LineHeight = src.Typography.LineHeight
	}
	if src.Typography.TextAlign != "" {
		dest.Typography.TextAlign = src.Typography.TextAlign
	}

	// Flexbox properties
	if src.Flexbox.Direction != "" {
		dest.Flexbox.Direction = src.Flexbox.Direction
	}
	if src.Flexbox.JustifyContent != "" {
		dest.Flexbox.JustifyContent = src.Flexbox.JustifyContent
	}
	if src.Flexbox.AlignItems != "" {
		dest.Flexbox.AlignItems = src.Flexbox.AlignItems
	}
	if src.Flexbox.Wrap != "" {
		dest.Flexbox.Wrap = src.Flexbox.Wrap
	}
	if src.Flexbox.Flex != "" {
		dest.Flexbox.Flex = src.Flexbox.Flex
	}

	// Grid properties
	if src.Grid.TemplateColumns != "" {
		dest.Grid.TemplateColumns = src.Grid.TemplateColumns
	}
	if src.Grid.TemplateRows != "" {
		dest.Grid.TemplateRows = src.Grid.TemplateRows
	}
	if src.Grid.Gap != "" {
		dest.Grid.Gap = src.Grid.Gap
	}

	// Visual effects
	if src.VisualEffects.Opacity != "" {
		dest.VisualEffects.Opacity = src.VisualEffects.Opacity
	}
	if src.VisualEffects.BoxShadow != "" {
		dest.VisualEffects.BoxShadow = src.VisualEffects.BoxShadow
	}
	if src.VisualEffects.Transform != "" {
		dest.VisualEffects.Transform = src.VisualEffects.Transform
	}

	// Transition properties
	if src.Transition.Transition != "" {
		dest.Transition.Transition = src.Transition.Transition
	}

	// Overflow properties
	if src.Overflow.Overflow != "" {
		dest.Overflow.Overflow = src.Overflow.Overflow
	}
	if src.Overflow.X != "" {
		dest.Overflow.X = src.Overflow.X
	}
	if src.Overflow.Y != "" {
		dest.Overflow.Y = src.Overflow.Y
	}

	// Copy custom properties
	for k, v := range src.Custom {
		dest.Custom[k] = v
	}
}

func (tsg *TemplStyleGenerator) applyVariantStyles(styles *ExpandedStyles, config ComponentStyleConfig) error {
	if config.Variant == "" {
		return nil
	}

	// Variant styles are typically component-specific
	// This is handled in component-specific methods like ButtonStyles

	return nil
}

func (tsg *TemplStyleGenerator) applySizeStyles(styles *ExpandedStyles, config ComponentStyleConfig) error {
	if config.Size == "" {
		return nil
	}

	// Size styles are typically component-specific
	// This is handled in component-specific methods

	return nil
}

func (tsg *TemplStyleGenerator) applyStateStyles(styles *ExpandedStyles, config ComponentStyleConfig) error {
	switch config.State {
	case "disabled":
		styles.WithOpacity("0.5").
			WithCustomProperty("cursor", "not-allowed")

	case "loading":
		styles.WithCustomProperty("cursor", "progress")

	case "focused":
		styles.WithCustomProperty("outline", "2px solid #3b82f6").
			WithCustomProperty("outline-offset", "2px")

	case "hover":
		// Hover styles are typically handled via CSS :hover pseudo-class
		// Custom properties can be set for hover state values

	}

	return nil
}

func (tsg *TemplStyleGenerator) applyThemeStyles(styles *ExpandedStyles, config ComponentStyleConfig) error {
	switch config.Theme {
	case "dark":
		// Apply dark theme styles
		if styles.Background.Color == "#ffffff" || styles.Background.Color == "" {
			styles.WithBackgroundColor("#1f2937")
		}
		if styles.Typography.Color == "#111827" || styles.Typography.Color == "" {
			styles.WithColor("#f9fafb")
		}

	case "light", "":
		// Light theme is default

		// Add more theme variants as needed
	}

	return nil
}

func (tsg *TemplStyleGenerator) applyCustomProperties(styles *ExpandedStyles, config ComponentStyleConfig) {
	for key, value := range config.CustomProps {
		if strValue, ok := value.(string); ok {
			styles.WithCustomProperty(key, strValue)
		}
	}
}

func (tsg *TemplStyleGenerator) applyResponsiveStyles(styles *ExpandedStyles, config ComponentStyleConfig) {
	for breakpoint, value := range config.Responsive {
		switch breakpoint {
		case "sm":
			styles.WithCustomProperty("sm-override", value)
		case "md":
			styles.WithCustomProperty("md-override", value)
		case "lg":
			styles.WithCustomProperty("lg-override", value)
		case "xl":
			styles.WithCustomProperty("xl-override", value)
		}
	}
}

// Utility methods for Templ integration

// GenerateComponentJSON creates JSON configuration for Templ components
func (tsg *TemplStyleGenerator) GenerateComponentJSON(config ComponentStyleConfig) (string, error) {
	data, err := json.Marshal(config)
	return string(data), err
}

// ParseComponentJSON parses JSON configuration for component styling
func (tsg *TemplStyleGenerator) ParseComponentJSON(jsonStr string) (ComponentStyleConfig, error) {
	var config ComponentStyleConfig
	err := json.Unmarshal([]byte(jsonStr), &config)
	return config, err
}

// GenerateTemplHelperFunction creates a Go function for use in Templ components
func (tsg *TemplStyleGenerator) GenerateTemplHelperFunction(componentType string) string {
	functionName := strings.ReplaceAll(strings.Title(componentType), ".", "")

	return fmt.Sprintf(`
// %sStyles generates optimized CSS for %s component
func %sStyles(variant, size, state string, customProps map[string]any) string {
	generator := css.NewTemplStyleGenerator("../../../docs/ui/Schema")
	config := css.ComponentStyleConfig{
		ComponentType: "%s",
		Variant:       variant,
		Size:          size,
		State:         state,
		CustomProps:   customProps,
	}
	
	styleAttr, err := generator.GenerateTemplStyleAttribute(config)
	if err != nil {
		return ""
	}
	
	return styleAttr
}
`, functionName, componentType, functionName, componentType)
}

// Integration with web/engine/schema_registry.go
func (tsg *TemplStyleGenerator) IntegrateWithSchemaRegistry() map[string]any {
	return map[string]any{
		"generator": tsg,
		"methods": map[string]any{
			"button": tsg.ButtonStyles,
			"input":  tsg.InputStyles,
			"card":   tsg.CardStyles,
			"modal":  tsg.ModalStyles,
		},
	}
}
