package editors

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/niiniyare/erp/web/builder/core"
)

// StyleEditor handles CSS and design token editing
type StyleEditor struct {
	component *core.ComponentInstance
}

// StyleConfiguration represents styling settings
type StyleConfiguration struct {
	Classes      []string                   `json:"classes"`
	InlineStyles map[string]string          `json:"inlineStyles"`
	Tokens       map[string]any             `json:"tokens"`
	Responsive   map[string]ResponsiveStyle `json:"responsive,omitempty"`
	States       map[string]StateStyle      `json:"states,omitempty"`
	Animation    *AnimationConfig           `json:"animation,omitempty"`
	Theme        string                     `json:"theme,omitempty"`
}

// ResponsiveStyle represents responsive breakpoint styles
type ResponsiveStyle struct {
	Breakpoint string            `json:"breakpoint"` // "sm", "md", "lg", "xl"
	Classes    []string          `json:"classes"`
	Styles     map[string]string `json:"styles"`
}

// StateStyle represents component state styles
type StateStyle struct {
	State   string            `json:"state"` // "hover", "focus", "active", "disabled"
	Classes []string          `json:"classes"`
	Styles  map[string]string `json:"styles"`
}

// AnimationConfig represents animation settings
type AnimationConfig struct {
	Type       string         `json:"type"`     // "transition", "keyframe", "transform"
	Duration   string         `json:"duration"` // "0.3s", "300ms"
	Timing     string         `json:"timing"`   // "ease", "linear", "ease-in-out"
	Delay      string         `json:"delay,omitempty"`
	Iterations string         `json:"iterations,omitempty"` // "infinite", "3"
	Direction  string         `json:"direction,omitempty"`  // "normal", "reverse", "alternate"
	Config     map[string]any `json:"config"`
}

// NewStyleEditor creates a new style editor instance
func NewStyleEditor(component *core.ComponentInstance) *StyleEditor {
	return &StyleEditor{
		component: component,
	}
}

// GetStyleConfiguration returns the current style configuration
func (se *StyleEditor) GetStyleConfiguration() *StyleConfiguration {
	if se.component.Style == nil {
		return &StyleConfiguration{
			Classes:      []string{},
			InlineStyles: make(map[string]string),
			Tokens:       make(map[string]any),
			Responsive:   make(map[string]ResponsiveStyle),
			States:       make(map[string]StateStyle),
		}
	}

	config := &StyleConfiguration{}
	config.Classes = se.extractClasses()
	config.InlineStyles = se.extractInlineStyles()
	config.Tokens = se.extractTokens()
	config.Responsive = se.extractResponsiveStyles()
	config.States = se.extractStateStyles()
	config.Animation = se.extractAnimationConfig()
	config.Theme = se.extractTheme()

	return config
}

// UpdateStyleConfiguration applies new style configuration
func (se *StyleEditor) UpdateStyleConfiguration(config *StyleConfiguration) error {
	if se.component.Style == nil {
		se.component.Style = &core.ComponentStyle{}
	}

	// Apply classes
	se.applyClasses(config.Classes)

	// Apply inline styles
	se.applyInlineStyles(config.InlineStyles)

	// Apply design tokens
	se.applyTokens(config.Tokens)

	// Apply responsive styles
	se.applyResponsiveStyles(config.Responsive)

	// Apply state styles
	se.applyStateStyles(config.States)

	// Apply animation
	if config.Animation != nil {
		se.applyAnimationConfig(config.Animation)
	}

	// Apply theme
	if config.Theme != "" {
		se.applyTheme(config.Theme)
	}

	return nil
}

// GetStylePresets returns common style configurations
func (se *StyleEditor) GetStylePresets() []StylePreset {
	return []StylePreset{
		{
			ID:          "card-elevated",
			Name:        "Elevated Card",
			Description: "Card with shadow and rounded corners",
			Category:    "Layout",
			Classes:     []string{"bg-white", "rounded-lg", "shadow-lg", "p-6"},
			Tokens: map[string]any{
				"background":   "white",
				"borderRadius": "8px",
				"shadow":       "0 10px 25px rgba(0, 0, 0, 0.1)",
				"padding":      "24px",
			},
		},
		{
			ID:          "button-primary",
			Name:        "Primary Button",
			Description: "Primary action button style",
			Category:    "Button",
			Classes:     []string{"bg-blue-600", "text-white", "px-4", "py-2", "rounded-md", "hover:bg-blue-700", "focus:ring-2", "focus:ring-blue-500"},
			States: map[string]StateStyle{
				"hover": {
					State:   "hover",
					Classes: []string{"bg-blue-700"},
				},
				"focus": {
					State:   "focus",
					Classes: []string{"ring-2", "ring-blue-500"},
				},
			},
		},
		{
			ID:          "input-field",
			Name:        "Input Field",
			Description: "Standard form input styling",
			Category:    "Form",
			Classes:     []string{"w-full", "px-3", "py-2", "border", "border-gray-300", "rounded-md", "focus:outline-none", "focus:ring-2", "focus:ring-blue-500"},
			States: map[string]StateStyle{
				"focus": {
					State:   "focus",
					Classes: []string{"ring-2", "ring-blue-500", "border-blue-500"},
				},
			},
		},
		{
			ID:          "text-gradient",
			Name:        "Gradient Text",
			Description: "Text with gradient color effect",
			Category:    "Typography",
			Classes:     []string{"bg-gradient-to-r", "from-purple-600", "to-blue-600", "bg-clip-text", "text-transparent"},
			InlineStyles: map[string]string{
				"background":           "linear-gradient(to right, #9333ea, #2563eb)",
				"webkitBackgroundClip": "text",
				"webkitTextFillColor":  "transparent",
			},
		},
		{
			ID:          "fade-in-animation",
			Name:        "Fade In Animation",
			Description: "Smooth fade in animation",
			Category:    "Animation",
			Animation: &AnimationConfig{
				Type:     "transition",
				Duration: "0.5s",
				Timing:   "ease-in-out",
				Config: map[string]any{
					"property": "opacity",
					"from":     "0",
					"to":       "1",
				},
			},
		},
	}
}

// StylePreset represents a predefined style configuration
type StylePreset struct {
	ID           string                     `json:"id"`
	Name         string                     `json:"name"`
	Description  string                     `json:"description"`
	Category     string                     `json:"category"`
	Classes      []string                   `json:"classes,omitempty"`
	InlineStyles map[string]string          `json:"inlineStyles,omitempty"`
	Tokens       map[string]any             `json:"tokens,omitempty"`
	Responsive   map[string]ResponsiveStyle `json:"responsive,omitempty"`
	States       map[string]StateStyle      `json:"states,omitempty"`
	Animation    *AnimationConfig           `json:"animation,omitempty"`
	Icon         string                     `json:"icon,omitempty"`
}

// GetDesignTokens returns available design tokens
func (se *StyleEditor) GetDesignTokens() map[string]TokenCategory {
	return map[string]TokenCategory{
		"colors": {
			Name:        "Colors",
			Description: "Color palette tokens",
			Tokens: map[string]DesignToken{
				"primary": {
					Name:        "Primary",
					Value:       "#3b82f6",
					Type:        "color",
					Description: "Primary brand color",
				},
				"secondary": {
					Name:        "Secondary",
					Value:       "#64748b",
					Type:        "color",
					Description: "Secondary color",
				},
				"success": {
					Name:        "Success",
					Value:       "#10b981",
					Type:        "color",
					Description: "Success state color",
				},
				"warning": {
					Name:        "Warning",
					Value:       "#f59e0b",
					Type:        "color",
					Description: "Warning state color",
				},
				"error": {
					Name:        "Error",
					Value:       "#ef4444",
					Type:        "color",
					Description: "Error state color",
				},
			},
		},
		"spacing": {
			Name:        "Spacing",
			Description: "Spacing scale tokens",
			Tokens: map[string]DesignToken{
				"xs": {
					Name:        "Extra Small",
					Value:       "4px",
					Type:        "spacing",
					Description: "Extra small spacing",
				},
				"sm": {
					Name:        "Small",
					Value:       "8px",
					Type:        "spacing",
					Description: "Small spacing",
				},
				"md": {
					Name:        "Medium",
					Value:       "16px",
					Type:        "spacing",
					Description: "Medium spacing",
				},
				"lg": {
					Name:        "Large",
					Value:       "24px",
					Type:        "spacing",
					Description: "Large spacing",
				},
				"xl": {
					Name:        "Extra Large",
					Value:       "32px",
					Type:        "spacing",
					Description: "Extra large spacing",
				},
			},
		},
		"typography": {
			Name:        "Typography",
			Description: "Typography tokens",
			Tokens: map[string]DesignToken{
				"fontFamily": {
					Name:        "Font Family",
					Value:       "-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif",
					Type:        "fontFamily",
					Description: "Primary font family",
				},
				"fontSize": {
					Name:        "Font Size",
					Value:       "16px",
					Type:        "fontSize",
					Description: "Base font size",
				},
				"lineHeight": {
					Name:        "Line Height",
					Value:       "1.5",
					Type:        "lineHeight",
					Description: "Base line height",
				},
			},
		},
		"borders": {
			Name:        "Borders",
			Description: "Border tokens",
			Tokens: map[string]DesignToken{
				"radius": {
					Name:        "Border Radius",
					Value:       "6px",
					Type:        "borderRadius",
					Description: "Standard border radius",
				},
				"width": {
					Name:        "Border Width",
					Value:       "1px",
					Type:        "borderWidth",
					Description: "Standard border width",
				},
			},
		},
		"shadows": {
			Name:        "Shadows",
			Description: "Shadow tokens",
			Tokens: map[string]DesignToken{
				"sm": {
					Name:        "Small Shadow",
					Value:       "0 1px 2px rgba(0, 0, 0, 0.05)",
					Type:        "boxShadow",
					Description: "Small drop shadow",
				},
				"md": {
					Name:        "Medium Shadow",
					Value:       "0 4px 6px rgba(0, 0, 0, 0.1)",
					Type:        "boxShadow",
					Description: "Medium drop shadow",
				},
				"lg": {
					Name:        "Large Shadow",
					Value:       "0 10px 15px rgba(0, 0, 0, 0.1)",
					Type:        "boxShadow",
					Description: "Large drop shadow",
				},
			},
		},
	}
}

// TokenCategory represents a category of design tokens
type TokenCategory struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Tokens      map[string]DesignToken `json:"tokens"`
}

// DesignToken represents a design token
type DesignToken struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

// GetCSSProperties returns categorized CSS properties
func (se *StyleEditor) GetCSSProperties() map[string]PropertyCategory {
	return map[string]PropertyCategory{
		"layout": {
			Name: "Layout",
			Properties: []CSSProperty{
				{Name: "display", Type: "select", Options: []string{"block", "inline", "flex", "grid", "none"}},
				{Name: "position", Type: "select", Options: []string{"static", "relative", "absolute", "fixed", "sticky"}},
				{Name: "top", Type: "length"},
				{Name: "right", Type: "length"},
				{Name: "bottom", Type: "length"},
				{Name: "left", Type: "length"},
				{Name: "z-index", Type: "number"},
				{Name: "width", Type: "length"},
				{Name: "height", Type: "length"},
				{Name: "max-width", Type: "length"},
				{Name: "max-height", Type: "length"},
				{Name: "min-width", Type: "length"},
				{Name: "min-height", Type: "length"},
			},
		},
		"flexbox": {
			Name: "Flexbox",
			Properties: []CSSProperty{
				{Name: "flex-direction", Type: "select", Options: []string{"row", "column", "row-reverse", "column-reverse"}},
				{Name: "flex-wrap", Type: "select", Options: []string{"nowrap", "wrap", "wrap-reverse"}},
				{Name: "justify-content", Type: "select", Options: []string{"flex-start", "flex-end", "center", "space-between", "space-around", "space-evenly"}},
				{Name: "align-items", Type: "select", Options: []string{"flex-start", "flex-end", "center", "stretch", "baseline"}},
				{Name: "align-content", Type: "select", Options: []string{"flex-start", "flex-end", "center", "stretch", "space-between", "space-around"}},
				{Name: "gap", Type: "length"},
				{Name: "flex-grow", Type: "number"},
				{Name: "flex-shrink", Type: "number"},
				{Name: "flex-basis", Type: "length"},
			},
		},
		"grid": {
			Name: "Grid",
			Properties: []CSSProperty{
				{Name: "grid-template-columns", Type: "text"},
				{Name: "grid-template-rows", Type: "text"},
				{Name: "grid-template-areas", Type: "text"},
				{Name: "grid-column-gap", Type: "length"},
				{Name: "grid-row-gap", Type: "length"},
				{Name: "grid-column", Type: "text"},
				{Name: "grid-row", Type: "text"},
				{Name: "grid-area", Type: "text"},
			},
		},
		"spacing": {
			Name: "Spacing",
			Properties: []CSSProperty{
				{Name: "margin", Type: "length"},
				{Name: "margin-top", Type: "length"},
				{Name: "margin-right", Type: "length"},
				{Name: "margin-bottom", Type: "length"},
				{Name: "margin-left", Type: "length"},
				{Name: "padding", Type: "length"},
				{Name: "padding-top", Type: "length"},
				{Name: "padding-right", Type: "length"},
				{Name: "padding-bottom", Type: "length"},
				{Name: "padding-left", Type: "length"},
			},
		},
		"typography": {
			Name: "Typography",
			Properties: []CSSProperty{
				{Name: "font-family", Type: "text"},
				{Name: "font-size", Type: "length"},
				{Name: "font-weight", Type: "select", Options: []string{"100", "200", "300", "400", "500", "600", "700", "800", "900", "normal", "bold"}},
				{Name: "font-style", Type: "select", Options: []string{"normal", "italic", "oblique"}},
				{Name: "line-height", Type: "length"},
				{Name: "text-align", Type: "select", Options: []string{"left", "center", "right", "justify"}},
				{Name: "text-decoration", Type: "select", Options: []string{"none", "underline", "overline", "line-through"}},
				{Name: "text-transform", Type: "select", Options: []string{"none", "uppercase", "lowercase", "capitalize"}},
				{Name: "letter-spacing", Type: "length"},
				{Name: "word-spacing", Type: "length"},
			},
		},
		"appearance": {
			Name: "Appearance",
			Properties: []CSSProperty{
				{Name: "color", Type: "color"},
				{Name: "background-color", Type: "color"},
				{Name: "background-image", Type: "text"},
				{Name: "background-size", Type: "select", Options: []string{"auto", "cover", "contain"}},
				{Name: "background-position", Type: "text"},
				{Name: "background-repeat", Type: "select", Options: []string{"repeat", "no-repeat", "repeat-x", "repeat-y"}},
				{Name: "opacity", Type: "number", Min: 0, Max: 1, Step: 0.1},
			},
		},
		"borders": {
			Name: "Borders",
			Properties: []CSSProperty{
				{Name: "border", Type: "text"},
				{Name: "border-width", Type: "length"},
				{Name: "border-style", Type: "select", Options: []string{"solid", "dashed", "dotted", "double", "groove", "ridge", "inset", "outset", "none"}},
				{Name: "border-color", Type: "color"},
				{Name: "border-radius", Type: "length"},
				{Name: "border-top-left-radius", Type: "length"},
				{Name: "border-top-right-radius", Type: "length"},
				{Name: "border-bottom-left-radius", Type: "length"},
				{Name: "border-bottom-right-radius", Type: "length"},
			},
		},
		"effects": {
			Name: "Effects",
			Properties: []CSSProperty{
				{Name: "box-shadow", Type: "text"},
				{Name: "text-shadow", Type: "text"},
				{Name: "filter", Type: "text"},
				{Name: "backdrop-filter", Type: "text"},
				{Name: "transform", Type: "text"},
				{Name: "transition", Type: "text"},
			},
		},
	}
}

// PropertyCategory represents a category of CSS properties
type PropertyCategory struct {
	Name       string        `json:"name"`
	Properties []CSSProperty `json:"properties"`
}

// CSSProperty represents a CSS property definition
type CSSProperty struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"` // "text", "number", "color", "length", "select"
	Options     []string `json:"options,omitempty"`
	Min         float64  `json:"min,omitempty"`
	Max         float64  `json:"max,omitempty"`
	Step        float64  `json:"step,omitempty"`
	Unit        string   `json:"unit,omitempty"`
	Description string   `json:"description,omitempty"`
}

// GenerateCSS generates CSS string from configuration
func (se *StyleEditor) GenerateCSS(config *StyleConfiguration) string {
	var css strings.Builder

	// Add inline styles
	if len(config.InlineStyles) > 0 {
		for property, value := range config.InlineStyles {
			css.WriteString(fmt.Sprintf("%s: %s;\n", property, value))
		}
	}

	// Add responsive styles
	for breakpoint, style := range config.Responsive {
		css.WriteString(fmt.Sprintf("@media (min-width: %s) {\n", getBreakpointValue(breakpoint)))
		for property, value := range style.Styles {
			css.WriteString(fmt.Sprintf("  %s: %s;\n", property, value))
		}
		css.WriteString("}\n")
	}

	// Add state styles
	for state, style := range config.States {
		css.WriteString(fmt.Sprintf(":%s {\n", state))
		for property, value := range style.Styles {
			css.WriteString(fmt.Sprintf("  %s: %s;\n", property, value))
		}
		css.WriteString("}\n")
	}

	// Add animation
	if config.Animation != nil {
		animationCSS := se.generateAnimationCSS(config.Animation)
		css.WriteString(animationCSS)
	}

	return css.String()
}

// Helper methods for extracting configuration

func (se *StyleEditor) extractClasses() []string {
	if se.component.Style.Classes != nil {
		return se.component.Style.Classes
	}
	return []string{}
}

func (se *StyleEditor) extractInlineStyles() map[string]string {
	if se.component.Style.InlineStyles != nil {
		return se.component.Style.InlineStyles
	}
	return make(map[string]string)
}

func (se *StyleEditor) extractTokens() map[string]any {
	if se.component.Style.Tokens != nil {
		return se.component.Style.Tokens
	}
	return make(map[string]any)
}

func (se *StyleEditor) extractResponsiveStyles() map[string]ResponsiveStyle {
	styles := make(map[string]ResponsiveStyle)

	if se.component.Style.Responsive != nil {
		for breakpoint, style := range se.component.Style.Responsive {
			styles[breakpoint] = ResponsiveStyle{
				Breakpoint: breakpoint,
				Classes:    style.Classes,
				Styles:     style.Styles,
			}
		}
	}

	return styles
}

func (se *StyleEditor) extractStateStyles() map[string]StateStyle {
	styles := make(map[string]StateStyle)

	if se.component.Style.States != nil {
		for state, style := range se.component.Style.States {
			styles[state] = StateStyle{
				State:   state,
				Classes: style.Classes,
				Styles:  style.Styles,
			}
		}
	}

	return styles
}

func (se *StyleEditor) extractAnimationConfig() *AnimationConfig {
	if se.component.Style.Animation == nil {
		return nil
	}

	return &AnimationConfig{
		Type:       se.component.Style.Animation.Type,
		Duration:   se.component.Style.Animation.Duration,
		Timing:     se.component.Style.Animation.Timing,
		Delay:      se.component.Style.Animation.Delay,
		Iterations: se.component.Style.Animation.Iterations,
		Direction:  se.component.Style.Animation.Direction,
		Config:     se.component.Style.Animation.Config,
	}
}

func (se *StyleEditor) extractTheme() string {
	if se.component.Style.Theme != "" {
		return se.component.Style.Theme
	}
	return "default"
}

// Helper methods for applying configuration

func (se *StyleEditor) applyClasses(classes []string) {
	se.component.Style.Classes = classes
}

func (se *StyleEditor) applyInlineStyles(styles map[string]string) {
	se.component.Style.InlineStyles = styles
}

func (se *StyleEditor) applyTokens(tokens map[string]any) {
	se.component.Style.Tokens = tokens
}

func (se *StyleEditor) applyResponsiveStyles(styles map[string]ResponsiveStyle) {
	if se.component.Style.Responsive == nil {
		se.component.Style.Responsive = make(map[string]core.ResponsiveStyle)
	}

	for breakpoint, style := range styles {
		se.component.Style.Responsive[breakpoint] = core.ResponsiveStyle{
			Classes: style.Classes,
			Styles:  style.Styles,
		}
	}
}

func (se *StyleEditor) applyStateStyles(styles map[string]StateStyle) {
	if se.component.Style.States == nil {
		se.component.Style.States = make(map[string]core.StateStyle)
	}

	for state, style := range styles {
		se.component.Style.States[state] = core.StateStyle{
			Classes: style.Classes,
			Styles:  style.Styles,
		}
	}
}

func (se *StyleEditor) applyAnimationConfig(config *AnimationConfig) {
	if se.component.Style.Animation == nil {
		se.component.Style.Animation = &core.AnimationConfig{}
	}

	se.component.Style.Animation.Type = config.Type
	se.component.Style.Animation.Duration = config.Duration
	se.component.Style.Animation.Timing = config.Timing
	se.component.Style.Animation.Delay = config.Delay
	se.component.Style.Animation.Iterations = config.Iterations
	se.component.Style.Animation.Direction = config.Direction
	se.component.Style.Animation.Config = config.Config
}

func (se *StyleEditor) applyTheme(theme string) {
	se.component.Style.Theme = theme
}

// Helper functions

func getBreakpointValue(breakpoint string) string {
	breakpoints := map[string]string{
		"sm":  "640px",
		"md":  "768px",
		"lg":  "1024px",
		"xl":  "1280px",
		"2xl": "1536px",
	}

	if value, exists := breakpoints[breakpoint]; exists {
		return value
	}
	return "768px"
}

func (se *StyleEditor) generateAnimationCSS(config *AnimationConfig) string {
	var css strings.Builder

	switch config.Type {
	case "transition":
		properties := []string{}
		if property, ok := config.Config["property"].(string); ok {
			properties = append(properties, property)
		}
		if len(properties) == 0 {
			properties = []string{"all"}
		}

		transition := fmt.Sprintf("transition: %s %s %s",
			strings.Join(properties, ", "),
			config.Duration,
			config.Timing)

		if config.Delay != "" {
			transition += " " + config.Delay
		}

		css.WriteString(transition + ";\n")

	case "keyframe":
		animationName := "custom-animation"
		if name, ok := config.Config["name"].(string); ok {
			animationName = name
		}

		// Generate keyframe rule
		css.WriteString(fmt.Sprintf("@keyframes %s {\n", animationName))
		if keyframes, ok := config.Config["keyframes"].(map[string]any); ok {
			for percentage, styles := range keyframes {
				css.WriteString(fmt.Sprintf("  %s {\n", percentage))
				if styleMap, ok := styles.(map[string]any); ok {
					for property, value := range styleMap {
						css.WriteString(fmt.Sprintf("    %s: %s;\n", property, value))
					}
				}
				css.WriteString("  }\n")
			}
		}
		css.WriteString("}\n")

		// Apply animation
		animation := fmt.Sprintf("animation: %s %s %s", animationName, config.Duration, config.Timing)
		if config.Delay != "" {
			animation += " " + config.Delay
		}
		if config.Iterations != "" {
			animation += " " + config.Iterations
		}
		if config.Direction != "" {
			animation += " " + config.Direction
		}
		css.WriteString(animation + ";\n")

	case "transform":
		if transforms, ok := config.Config["transforms"].([]any); ok {
			transformValues := []string{}
			for _, transform := range transforms {
				if transformStr, ok := transform.(string); ok {
					transformValues = append(transformValues, transformStr)
				}
			}
			if len(transformValues) > 0 {
				css.WriteString(fmt.Sprintf("transform: %s;\n", strings.Join(transformValues, " ")))
			}
		}
	}

	return css.String()
}

// ValidateStyleConfiguration validates the style configuration
func (se *StyleEditor) ValidateStyleConfiguration(config *StyleConfiguration) []ValidationError {
	var errors []ValidationError

	// Validate CSS properties
	for property, value := range config.InlineStyles {
		if value == "" {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("inlineStyles.%s", property),
				Message: "CSS property value cannot be empty",
			})
		}
	}

	// Validate animation configuration
	if config.Animation != nil {
		if config.Animation.Duration == "" {
			errors = append(errors, ValidationError{
				Field:   "animation.duration",
				Message: "Animation duration is required",
			})
		}

		if config.Animation.Type == "keyframe" {
			if _, ok := config.Animation.Config["keyframes"]; !ok {
				errors = append(errors, ValidationError{
					Field:   "animation.config.keyframes",
					Message: "Keyframes are required for keyframe animations",
				})
			}
		}
	}

	// Validate responsive breakpoints
	validBreakpoints := map[string]bool{
		"sm": true, "md": true, "lg": true, "xl": true, "2xl": true,
	}

	for breakpoint := range config.Responsive {
		if !validBreakpoints[breakpoint] {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("responsive.%s", breakpoint),
				Message: "Invalid responsive breakpoint",
			})
		}
	}

	return errors
}

// ConvertTokensToCSS converts design tokens to CSS custom properties
func (se *StyleEditor) ConvertTokensToCSS(tokens map[string]any) string {
	var css strings.Builder

	css.WriteString(":root {\n")
	for name, value := range tokens {
		// Convert token name to CSS custom property
		cssVar := "--" + strings.ReplaceAll(name, ".", "-")
		css.WriteString(fmt.Sprintf("  %s: %v;\n", cssVar, value))
	}
	css.WriteString("}\n")

	return css.String()
}

// ParseColorValue parses and validates color values
func (se *StyleEditor) ParseColorValue(value string) (bool, string) {
	// Remove whitespace
	value = strings.TrimSpace(value)

	// Check for hex colors
	if strings.HasPrefix(value, "#") {
		if len(value) == 4 || len(value) == 7 {
			// Validate hex digits
			for _, char := range value[1:] {
				if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
					return false, "Invalid hex color format"
				}
			}
			return true, ""
		}
		return false, "Hex color must be 3 or 6 digits"
	}

	// Check for rgb/rgba
	if strings.HasPrefix(value, "rgb(") || strings.HasPrefix(value, "rgba(") {
		return true, "" // Basic validation, could be more thorough
	}

	// Check for hsl/hsla
	if strings.HasPrefix(value, "hsl(") || strings.HasPrefix(value, "hsla(") {
		return true, "" // Basic validation, could be more thorough
	}

	// Check for named colors (basic list)
	namedColors := map[string]bool{
		"black": true, "white": true, "red": true, "green": true, "blue": true,
		"yellow": true, "cyan": true, "magenta": true, "gray": true, "grey": true,
		"transparent": true, "currentColor": true,
	}

	if namedColors[strings.ToLower(value)] {
		return true, ""
	}

	return false, "Invalid color format"
}

// ParseLengthValue parses and validates length values
func (se *StyleEditor) ParseLengthValue(value string) (bool, string) {
	value = strings.TrimSpace(value)

	// Check for auto, inherit, initial, unset
	keywords := []string{"auto", "inherit", "initial", "unset", "none"}
	for _, keyword := range keywords {
		if strings.ToLower(value) == keyword {
			return true, ""
		}
	}

	// Check for calc() expressions
	if strings.HasPrefix(value, "calc(") && strings.HasSuffix(value, ")") {
		return true, "" // Basic validation for calc expressions
	}

	// Check for length units
	units := []string{"px", "em", "rem", "vh", "vw", "vmin", "vmax", "%", "pt", "pc", "in", "cm", "mm", "ex", "ch"}
	for _, unit := range units {
		if strings.HasSuffix(value, unit) {
			numStr := strings.TrimSuffix(value, unit)
			if _, err := strconv.ParseFloat(numStr, 64); err == nil {
				return true, ""
			}
		}
	}

	// Check for unitless zero
	if value == "0" {
		return true, ""
	}

	return false, "Invalid length value"
}
