// Package css - Component-specific CSS generation for UI schemas
// This system provides intelligent CSS generation for each component type
package css

import (
	"fmt"
)

// ComponentCSSBuilder provides CSS generation specifically for UI components
type ComponentCSSBuilder struct {
	generator *CSSGenerator
	theme     *ComponentTheme
}

// ComponentTheme contains design system values for consistent styling
type ComponentTheme struct {
	// Color palette
	Colors map[string]string
	// Spacing scale (margins, padding)
	Spacing map[string]string
	// Typography scale
	Typography map[string]TypographyConfig
	// Shadow levels
	Shadows map[string]string
	// Border radius values
	BorderRadius map[string]string
	// Z-index scale
	ZIndex map[string]int
	// Breakpoints for responsive design
	Breakpoints map[string]string
	// Animation durations and easings
	Animations map[string]AnimationConfig
}

// TypographyConfig represents typography settings
type TypographyConfig struct {
	FontSize      string
	LineHeight    string
	FontWeight    string
	LetterSpacing string
}

// AnimationConfig represents animation settings
type AnimationConfig struct {
	Duration string
	Easing   string
	Delay    string
}

// ComponentCSSOptions contains options for component CSS generation
type ComponentCSSOptions struct {
	// Component type (button, input, etc.)
	ComponentType string
	// Size variant (xs, sm, md, lg, xl)
	Size string
	// Color variant (primary, secondary, success, etc.)
	Variant string
	// Whether component is disabled
	Disabled bool
	// Custom CSS classes to apply
	CustomClasses []string
	// Responsive breakpoints to generate
	Responsive map[string]map[string]string
	// States to generate (:hover, :focus, :active)
	States map[string]map[string]string
	// Custom class name to use instead of auto-generated one
	CustomClassName string
}

// NewComponentCSSBuilder creates a CSS builder with ERP theme defaults
func NewComponentCSSBuilder() *ComponentCSSBuilder {
	return &ComponentCSSBuilder{
		generator: NewCSSGenerator(),
		theme:     DefaultERPTheme(),
	}
}

// NewComponentCSSBuilderWithTheme creates a CSS builder with custom theme
func NewComponentCSSBuilderWithTheme(theme *ComponentTheme) *ComponentCSSBuilder {
	return &ComponentCSSBuilder{
		generator: NewCSSGenerator(),
		theme:     theme,
	}
}

// DefaultERPTheme provides a professional ERP design system theme
func DefaultERPTheme() *ComponentTheme {
	return &ComponentTheme{
		Colors: map[string]string{
			// Primary colors
			"primary-50":  "#eff6ff",
			"primary-100": "#dbeafe",
			"primary-500": "#3b82f6", // Main primary
			"primary-600": "#2563eb",
			"primary-700": "#1d4ed8",

			// Secondary colors
			"secondary-50":  "#f8fafc",
			"secondary-100": "#f1f5f9",
			"secondary-500": "#64748b", // Main secondary
			"secondary-600": "#475569",
			"secondary-700": "#334155",

			// Success colors
			"success-50":  "#f0fdf4",
			"success-100": "#dcfce7",
			"success-500": "#22c55e", // Main success
			"success-600": "#16a34a",
			"success-700": "#15803d",

			// Warning colors
			"warning-50":  "#fffbeb",
			"warning-100": "#fef3c7",
			"warning-500": "#f59e0b", // Main warning
			"warning-600": "#d97706",
			"warning-700": "#b45309",

			// Error colors
			"error-50":  "#fef2f2",
			"error-100": "#fee2e2",
			"error-500": "#ef4444", // Main error
			"error-600": "#dc2626",
			"error-700": "#b91c1c",

			// Neutral colors
			"neutral-50":  "#fafafa",
			"neutral-100": "#f5f5f5",
			"neutral-200": "#e5e5e5",
			"neutral-300": "#d4d4d4",
			"neutral-400": "#a3a3a3",
			"neutral-500": "#737373",
			"neutral-600": "#525252",
			"neutral-700": "#404040",
			"neutral-800": "#262626",
			"neutral-900": "#171717",
		},
		Spacing: map[string]string{
			"xs":  "0.25rem", // 4px
			"sm":  "0.5rem",  // 8px
			"md":  "1rem",    // 16px
			"lg":  "1.5rem",  // 24px
			"xl":  "2rem",    // 32px
			"2xl": "2.5rem",  // 40px
			"3xl": "3rem",    // 48px
		},
		Typography: map[string]TypographyConfig{
			"xs": {FontSize: "0.75rem", LineHeight: "1rem", FontWeight: "400", LetterSpacing: "0"},
			"sm": {FontSize: "0.875rem", LineHeight: "1.25rem", FontWeight: "400", LetterSpacing: "0"},
			"md": {FontSize: "1rem", LineHeight: "1.5rem", FontWeight: "400", LetterSpacing: "0"},
			"lg": {FontSize: "1.125rem", LineHeight: "1.75rem", FontWeight: "400", LetterSpacing: "0"},
			"xl": {FontSize: "1.25rem", LineHeight: "1.75rem", FontWeight: "500", LetterSpacing: "0"},
		},
		Shadows: map[string]string{
			"none": "none",
			"sm":   "0 1px 2px 0 rgb(0 0 0 / 0.05)",
			"md":   "0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1)",
			"lg":   "0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1)",
			"xl":   "0 20px 25px -5px rgb(0 0 0 / 0.1), 0 8px 10px -6px rgb(0 0 0 / 0.1)",
		},
		BorderRadius: map[string]string{
			"none": "0",
			"sm":   "0.125rem", // 2px
			"md":   "0.375rem", // 6px
			"lg":   "0.5rem",   // 8px
			"xl":   "0.75rem",  // 12px
			"full": "9999px",
		},
		ZIndex: map[string]int{
			"dropdown": 1000,
			"modal":    1050,
			"tooltip":  1100,
			"toast":    1200,
		},
		Breakpoints: map[string]string{
			"sm": "640px",
			"md": "768px",
			"lg": "1024px",
			"xl": "1280px",
		},
		Animations: map[string]AnimationConfig{
			"fast":   {Duration: "150ms", Easing: "cubic-bezier(0.4, 0, 0.2, 1)", Delay: "0ms"},
			"normal": {Duration: "300ms", Easing: "cubic-bezier(0.4, 0, 0.2, 1)", Delay: "0ms"},
			"slow":   {Duration: "500ms", Easing: "cubic-bezier(0.4, 0, 0.2, 1)", Delay: "0ms"},
		},
	}
}

// GenerateButtonCSS generates CSS for button components
func (b *ComponentCSSBuilder) GenerateButtonCSS(options ComponentCSSOptions) *ComponentCSSBuilder {
	size := options.Size
	if size == "" {
		size = "md"
	}
	variant := options.Variant
	if variant == "" {
		variant = "primary"
	}

	className := fmt.Sprintf(".btn-%s-%s", variant, size)

	// Base button styles
	baseStyles := map[string]string{
		"display":         "inline-flex",
		"align-items":     "center",
		"justify-content": "center",
		"font-weight":     "500",
		"transition":      "all " + b.theme.Animations["fast"].Duration + " " + b.theme.Animations["fast"].Easing,
		"cursor":          "pointer",
		"user-select":     "none",
		"border":          "1px solid transparent",
		"text-decoration": "none",
		"outline":         "none",
	}

	// Size-specific styles
	sizeStyles := map[string]map[string]string{
		"xs": {
			"padding":       b.theme.Spacing["xs"] + " " + b.theme.Spacing["sm"],
			"font-size":     b.theme.Typography["xs"].FontSize,
			"line-height":   b.theme.Typography["xs"].LineHeight,
			"border-radius": b.theme.BorderRadius["sm"],
		},
		"sm": {
			"padding":       b.theme.Spacing["sm"] + " " + b.theme.Spacing["md"],
			"font-size":     b.theme.Typography["sm"].FontSize,
			"line-height":   b.theme.Typography["sm"].LineHeight,
			"border-radius": b.theme.BorderRadius["md"],
		},
		"md": {
			"padding":       b.theme.Spacing["md"] + " " + b.theme.Spacing["lg"],
			"font-size":     b.theme.Typography["md"].FontSize,
			"line-height":   b.theme.Typography["md"].LineHeight,
			"border-radius": b.theme.BorderRadius["md"],
		},
		"lg": {
			"padding":       b.theme.Spacing["lg"] + " " + b.theme.Spacing["xl"],
			"font-size":     b.theme.Typography["lg"].FontSize,
			"line-height":   b.theme.Typography["lg"].LineHeight,
			"border-radius": b.theme.BorderRadius["lg"],
		},
	}

	// Variant-specific styles
	variantStyles := map[string]map[string]string{
		"primary": {
			"background-color": b.theme.Colors["primary-500"],
			"color":            "#ffffff",
			"border-color":     b.theme.Colors["primary-500"],
			"box-shadow":       b.theme.Shadows["sm"],
		},
		"secondary": {
			"background-color": b.theme.Colors["secondary-100"],
			"color":            b.theme.Colors["secondary-700"],
			"border-color":     b.theme.Colors["secondary-200"],
		},
		"success": {
			"background-color": b.theme.Colors["success-500"],
			"color":            "#ffffff",
			"border-color":     b.theme.Colors["success-500"],
		},
		"warning": {
			"background-color": b.theme.Colors["warning-500"],
			"color":            "#ffffff",
			"border-color":     b.theme.Colors["warning-500"],
		},
		"error": {
			"background-color": b.theme.Colors["error-500"],
			"color":            "#ffffff",
			"border-color":     b.theme.Colors["error-500"],
		},
		"outline": {
			"background-color": "transparent",
			"color":            b.theme.Colors["primary-600"],
			"border-color":     b.theme.Colors["primary-500"],
		},
		"ghost": {
			"background-color": "transparent",
			"color":            b.theme.Colors["primary-600"],
			"border-color":     "transparent",
		},
	}

	// Merge styles
	finalStyles := baseStyles
	for k, v := range sizeStyles[size] {
		finalStyles[k] = v
	}
	for k, v := range variantStyles[variant] {
		finalStyles[k] = v
	}

	b.generator.AddRule(className, finalStyles)

	// Add hover state
	hoverStyles := map[string]string{}
	switch variant {
	case "primary":
		hoverStyles["background-color"] = b.theme.Colors["primary-600"]
		hoverStyles["border-color"] = b.theme.Colors["primary-600"]
	case "secondary":
		hoverStyles["background-color"] = b.theme.Colors["secondary-200"]
	case "outline":
		hoverStyles["background-color"] = b.theme.Colors["primary-50"]
	case "ghost":
		hoverStyles["background-color"] = b.theme.Colors["primary-50"]
	}

	if len(hoverStyles) > 0 {
		b.generator.AddRuleWithPseudo(className, ":hover", hoverStyles)
	}

	// Add focus state
	focusStyles := map[string]string{
		"outline":        "2px solid " + b.theme.Colors["primary-500"],
		"outline-offset": "2px",
	}
	b.generator.AddRuleWithPseudo(className, ":focus", focusStyles)

	// Add disabled state
	if options.Disabled {
		disabledStyles := map[string]string{
			"opacity":        "0.5",
			"cursor":         "not-allowed",
			"pointer-events": "none",
		}
		b.generator.AddRuleWithPseudo(className, ":disabled", disabledStyles)
	}

	return b
}

// GenerateInputCSS generates CSS for input components
func (b *ComponentCSSBuilder) GenerateInputCSS(options ComponentCSSOptions) *ComponentCSSBuilder {
	size := options.Size
	if size == "" {
		size = "md"
	}

	className := fmt.Sprintf(".input-%s", size)

	// Base input styles
	baseStyles := map[string]string{
		"display":          "block",
		"width":            "100%",
		"border":           "1px solid " + b.theme.Colors["neutral-300"],
		"background-color": "#ffffff",
		"color":            b.theme.Colors["neutral-900"],
		"transition":       "all " + b.theme.Animations["fast"].Duration + " " + b.theme.Animations["fast"].Easing,
		"outline":          "none",
	}

	// Size-specific styles
	sizeStyles := map[string]map[string]string{
		"sm": {
			"padding":       b.theme.Spacing["sm"] + " " + b.theme.Spacing["md"],
			"font-size":     b.theme.Typography["sm"].FontSize,
			"line-height":   b.theme.Typography["sm"].LineHeight,
			"border-radius": b.theme.BorderRadius["md"],
		},
		"md": {
			"padding":       b.theme.Spacing["md"] + " " + b.theme.Spacing["md"],
			"font-size":     b.theme.Typography["md"].FontSize,
			"line-height":   b.theme.Typography["md"].LineHeight,
			"border-radius": b.theme.BorderRadius["md"],
		},
		"lg": {
			"padding":       b.theme.Spacing["lg"] + " " + b.theme.Spacing["lg"],
			"font-size":     b.theme.Typography["lg"].FontSize,
			"line-height":   b.theme.Typography["lg"].LineHeight,
			"border-radius": b.theme.BorderRadius["lg"],
		},
	}

	// Merge styles
	finalStyles := baseStyles
	for k, v := range sizeStyles[size] {
		finalStyles[k] = v
	}

	b.generator.AddRule(className, finalStyles)

	// Focus state
	focusStyles := map[string]string{
		"border-color": b.theme.Colors["primary-500"],
		"box-shadow":   "0 0 0 3px " + b.theme.Colors["primary-100"],
	}
	b.generator.AddRuleWithPseudo(className, ":focus", focusStyles)

	// Error state
	errorClassName := className + "-error"
	errorStyles := map[string]string{
		"border-color": b.theme.Colors["error-500"],
	}
	b.generator.AddRule(errorClassName, errorStyles)

	// Error focus state
	errorFocusStyles := map[string]string{
		"border-color": b.theme.Colors["error-600"],
		"box-shadow":   "0 0 0 3px " + b.theme.Colors["error-100"],
	}
	b.generator.AddRuleWithPseudo(errorClassName, ":focus", errorFocusStyles)

	return b
}

// GenerateCardCSS generates CSS for card components
func (b *ComponentCSSBuilder) GenerateCardCSS(options ComponentCSSOptions) *ComponentCSSBuilder {
	variant := options.Variant
	if variant == "" {
		variant = "default"
	}

	className := ".card"
	if variant != "default" {
		className = fmt.Sprintf(".card-%s", variant)
	}

	// Base card styles
	baseStyles := map[string]string{
		"display":          "block",
		"background-color": "#ffffff",
		"border-radius":    b.theme.BorderRadius["lg"],
		"border":           "1px solid " + b.theme.Colors["neutral-200"],
		"overflow":         "hidden",
		"transition":       "all " + b.theme.Animations["fast"].Duration + " " + b.theme.Animations["fast"].Easing,
	}

	// Variant-specific styles
	switch variant {
	case "elevated":
		baseStyles["box-shadow"] = b.theme.Shadows["md"]
		baseStyles["border"] = "none"
	case "outlined":
		baseStyles["border"] = "2px solid " + b.theme.Colors["neutral-300"]
	case "filled":
		baseStyles["background-color"] = b.theme.Colors["neutral-50"]
	}

	b.generator.AddRule(className, baseStyles)

	// Card header
	headerStyles := map[string]string{
		"padding":       b.theme.Spacing["lg"],
		"border-bottom": "1px solid " + b.theme.Colors["neutral-200"],
	}
	b.generator.AddRule(className+" .card-header", headerStyles)

	// Card body
	bodyStyles := map[string]string{
		"padding": b.theme.Spacing["lg"],
	}
	b.generator.AddRule(className+" .card-body", bodyStyles)

	// Card footer
	footerStyles := map[string]string{
		"padding":          b.theme.Spacing["lg"],
		"border-top":       "1px solid " + b.theme.Colors["neutral-200"],
		"background-color": b.theme.Colors["neutral-50"],
	}
	b.generator.AddRule(className+" .card-footer", footerStyles)

	return b
}

// GenerateResponsiveCSS generates responsive CSS rules
func (b *ComponentCSSBuilder) GenerateResponsiveCSS(selector string, responsiveStyles map[string]map[string]string) *ComponentCSSBuilder {
	for breakpoint, styles := range responsiveStyles {
		if mediaQuery, exists := b.theme.Breakpoints[breakpoint]; exists {
			b.generator.AddRuleWithMedia(selector, styles, fmt.Sprintf("(min-width: %s)", mediaQuery))
		}
	}
	return b
}

// GenerateUtilityCSS generates common utility classes
func (b *ComponentCSSBuilder) GenerateUtilityCSS() *ComponentCSSBuilder {
	// Spacing utilities
	for name, value := range b.theme.Spacing {
		// Margin utilities
		b.generator.AddRule(fmt.Sprintf(".m-%s", name), map[string]string{"margin": value})
		b.generator.AddRule(fmt.Sprintf(".mt-%s", name), map[string]string{"margin-top": value})
		b.generator.AddRule(fmt.Sprintf(".mr-%s", name), map[string]string{"margin-right": value})
		b.generator.AddRule(fmt.Sprintf(".mb-%s", name), map[string]string{"margin-bottom": value})
		b.generator.AddRule(fmt.Sprintf(".ml-%s", name), map[string]string{"margin-left": value})
		b.generator.AddRule(fmt.Sprintf(".mx-%s", name), map[string]string{"margin-left": value, "margin-right": value})
		b.generator.AddRule(fmt.Sprintf(".my-%s", name), map[string]string{"margin-top": value, "margin-bottom": value})

		// Padding utilities
		b.generator.AddRule(fmt.Sprintf(".p-%s", name), map[string]string{"padding": value})
		b.generator.AddRule(fmt.Sprintf(".pt-%s", name), map[string]string{"padding-top": value})
		b.generator.AddRule(fmt.Sprintf(".pr-%s", name), map[string]string{"padding-right": value})
		b.generator.AddRule(fmt.Sprintf(".pb-%s", name), map[string]string{"padding-bottom": value})
		b.generator.AddRule(fmt.Sprintf(".pl-%s", name), map[string]string{"padding-left": value})
		b.generator.AddRule(fmt.Sprintf(".px-%s", name), map[string]string{"padding-left": value, "padding-right": value})
		b.generator.AddRule(fmt.Sprintf(".py-%s", name), map[string]string{"padding-top": value, "padding-bottom": value})
	}

	// Text utilities
	for name, config := range b.theme.Typography {
		b.generator.AddRule(fmt.Sprintf(".text-%s", name), map[string]string{
			"font-size":   config.FontSize,
			"line-height": config.LineHeight,
			"font-weight": config.FontWeight,
		})
	}

	// Color utilities
	for name, value := range b.theme.Colors {
		b.generator.AddRule(fmt.Sprintf(".text-%s", name), map[string]string{"color": value})
		b.generator.AddRule(fmt.Sprintf(".bg-%s", name), map[string]string{"background-color": value})
		b.generator.AddRule(fmt.Sprintf(".border-%s", name), map[string]string{"border-color": value})
	}

	// Border radius utilities
	for name, value := range b.theme.BorderRadius {
		b.generator.AddRule(fmt.Sprintf(".rounded-%s", name), map[string]string{"border-radius": value})
	}

	// Shadow utilities
	for name, value := range b.theme.Shadows {
		b.generator.AddRule(fmt.Sprintf(".shadow-%s", name), map[string]string{"box-shadow": value})
	}

	return b
}

// AddThemeVariables adds CSS custom properties for the theme
func (b *ComponentCSSBuilder) AddThemeVariables() *ComponentCSSBuilder {
	// Add color variables
	for name, value := range b.theme.Colors {
		b.generator.AddVariable("color-"+name, value)
	}

	// Add spacing variables
	for name, value := range b.theme.Spacing {
		b.generator.AddVariable("spacing-"+name, value)
	}

	// Add typography variables
	for name, config := range b.theme.Typography {
		b.generator.AddVariable("font-size-"+name, config.FontSize)
		b.generator.AddVariable("line-height-"+name, config.LineHeight)
		b.generator.AddVariable("font-weight-"+name, config.FontWeight)
	}

	// Add other variables
	for name, value := range b.theme.Shadows {
		b.generator.AddVariable("shadow-"+name, value)
	}

	for name, value := range b.theme.BorderRadius {
		b.generator.AddVariable("radius-"+name, value)
	}

	return b
}

// Generate returns the final CSS string
func (b *ComponentCSSBuilder) Generate() string {
	return b.generator.Generate()
}

// GetGenerator returns the underlying CSS generator for advanced operations
func (b *ComponentCSSBuilder) GetGenerator() *CSSGenerator {
	return b.generator
}
