// Package css provides self-contained CSS generation for atomic components
// This is a minimal, dependency-free implementation focused on atoms
package css

import (
	"fmt"
	"sort"
	"strings"
)

// ============================================================================
// CORE CSS TYPES
// ============================================================================

// Styles represents CSS properties for atomic components
type Styles struct {
	// Layout Properties
	Display   string `json:"display,omitempty"`
	Position  string `json:"position,omitempty"`
	Width     string `json:"width,omitempty"`
	Height    string `json:"height,omitempty"`
	MaxWidth  string `json:"max_width,omitempty"`
	MaxHeight string `json:"max_height,omitempty"`
	MinWidth  string `json:"min_width,omitempty"`
	MinHeight string `json:"min_height,omitempty"`

	// Spacing Properties
	Margin        string `json:"margin,omitempty"`
	MarginTop     string `json:"margin_top,omitempty"`
	MarginBottom  string `json:"margin_bottom,omitempty"`
	MarginLeft    string `json:"margin_left,omitempty"`
	MarginRight   string `json:"margin_right,omitempty"`
	Padding       string `json:"padding,omitempty"`
	PaddingTop    string `json:"padding_top,omitempty"`
	PaddingBottom string `json:"padding_bottom,omitempty"`
	PaddingLeft   string `json:"padding_left,omitempty"`
	PaddingRight  string `json:"padding_right,omitempty"`

	// Typography Properties
	FontSize   string `json:"font_size,omitempty"`
	FontWeight string `json:"font_weight,omitempty"`
	FontFamily string `json:"font_family,omitempty"`
	LineHeight string `json:"line_height,omitempty"`
	Color      string `json:"color,omitempty"`
	TextAlign  string `json:"text_align,omitempty"`

	// Background Properties
	Background      string `json:"background,omitempty"`
	BackgroundColor string `json:"background_color,omitempty"`

	// Border Properties
	Border       string `json:"border,omitempty"`
	BorderRadius string `json:"border_radius,omitempty"`
	BorderColor  string `json:"border_color,omitempty"`
	BorderWidth  string `json:"border_width,omitempty"`
	BorderStyle  string `json:"border_style,omitempty"`

	// Flexbox Properties
	Flex           string `json:"flex,omitempty"`
	FlexDirection  string `json:"flex_direction,omitempty"`
	FlexWrap       string `json:"flex_wrap,omitempty"`
	JustifyContent string `json:"justify_content,omitempty"`
	AlignItems     string `json:"align_items,omitempty"`

	// Visual Effects
	BoxShadow string `json:"box_shadow,omitempty"`
	Opacity   string `json:"opacity,omitempty"`
	Transform string `json:"transform,omitempty"`

	// Animation Properties
	Transition         string `json:"transition,omitempty"`
	TransitionDuration string `json:"transition_duration,omitempty"`

	// Overflow Properties
	Overflow string `json:"overflow,omitempty"`

	// Z-Index and Visibility
	ZIndex     string `json:"z_index,omitempty"`
	Visibility string `json:"visibility,omitempty"`

	// Custom properties for framework-specific values
	Custom map[string]string `json:"custom,omitempty"`
}

// NewStyles creates a new Styles instance
func NewStyles() *Styles {
	return &Styles{
		Custom: make(map[string]string),
	}
}

// ============================================================================
// CSS GENERATION METHODS
// ============================================================================

// ToCSS generates CSS string from all properties
func (s *Styles) ToCSS() string {
	var cssRules []string
	properties := s.getAllProperties()

	for property, value := range properties {
		if value != "" {
			cssProperty := s.toCSSPropertyName(property)
			cssRules = append(cssRules, fmt.Sprintf("%s: %s", cssProperty, value))
		}
	}

	// Add custom properties
	for property, value := range s.Custom {
		if value != "" {
			cssRules = append(cssRules, fmt.Sprintf("%s: %s", property, value))
		}
	}

	if len(cssRules) == 0 {
		return ""
	}

	// Sort for consistent output
	sort.Strings(cssRules)
	return strings.Join(cssRules, "; ")
}

// ToCSSClass generates a CSS class with all properties
func (s *Styles) ToCSSClass(className string) string {
	cssRules := s.ToCSS()
	if cssRules == "" {
		return ""
	}

	return fmt.Sprintf(".%s { %s; }", className, cssRules)
}

// ============================================================================
// FLUENT API METHODS
// ============================================================================

// WithDisplay sets the display property
func (s *Styles) WithDisplay(display string) *Styles {
	s.Display = display
	return s
}

// WithPosition sets the position property
func (s *Styles) WithPosition(position string) *Styles {
	s.Position = position
	return s
}

// WithWidth sets the width property
func (s *Styles) WithWidth(width string) *Styles {
	s.Width = width
	return s
}

// WithHeight sets the height property
func (s *Styles) WithHeight(height string) *Styles {
	s.Height = height
	return s
}

// WithMargin sets the margin property
func (s *Styles) WithMargin(margin string) *Styles {
	s.Margin = margin
	return s
}

// WithPadding sets the padding property
func (s *Styles) WithPadding(padding string) *Styles {
	s.Padding = padding
	return s
}

// WithFontSize sets the font-size property
func (s *Styles) WithFontSize(fontSize string) *Styles {
	s.FontSize = fontSize
	return s
}

// WithColor sets the color property
func (s *Styles) WithColor(color string) *Styles {
	s.Color = color
	return s
}

// WithBackgroundColor sets the background-color property
func (s *Styles) WithBackgroundColor(backgroundColor string) *Styles {
	s.BackgroundColor = backgroundColor
	return s
}

// WithBorderRadius sets the border-radius property
func (s *Styles) WithBorderRadius(borderRadius string) *Styles {
	s.BorderRadius = borderRadius
	return s
}

// WithBorder sets the border property
func (s *Styles) WithBorder(border string) *Styles {
	s.Border = border
	return s
}

// WithFlex sets flexbox properties
func (s *Styles) WithFlex(flex string) *Styles {
	s.Flex = flex
	return s
}

// WithFlexDirection sets the flex-direction property
func (s *Styles) WithFlexDirection(direction string) *Styles {
	s.FlexDirection = direction
	return s
}

// WithJustifyContent sets the justify-content property
func (s *Styles) WithJustifyContent(justify string) *Styles {
	s.JustifyContent = justify
	return s
}

// WithAlignItems sets the align-items property
func (s *Styles) WithAlignItems(align string) *Styles {
	s.AlignItems = align
	return s
}

// WithFontWeight sets the font-weight property
func (s *Styles) WithFontWeight(weight string) *Styles {
	s.FontWeight = weight
	return s
}

// WithTransition sets the transition property
func (s *Styles) WithTransition(transition string) *Styles {
	s.Transition = transition
	return s
}

// WithBoxShadow sets the box-shadow property
func (s *Styles) WithBoxShadow(shadow string) *Styles {
	s.BoxShadow = shadow
	return s
}

// WithOpacity sets the opacity property
func (s *Styles) WithOpacity(opacity string) *Styles {
	s.Opacity = opacity
	return s
}

// WithCustomProperty sets a custom CSS property
func (s *Styles) WithCustomProperty(property, value string) *Styles {
	if s.Custom == nil {
		s.Custom = make(map[string]string)
	}
	s.Custom[property] = value
	return s
}

// ============================================================================
// HELPER METHODS
// ============================================================================

// getAllProperties returns all non-empty CSS properties as a map
func (s *Styles) getAllProperties() map[string]string {
	return map[string]string{
		"display":             s.Display,
		"position":            s.Position,
		"width":               s.Width,
		"height":              s.Height,
		"max-width":           s.MaxWidth,
		"max-height":          s.MaxHeight,
		"min-width":           s.MinWidth,
		"min-height":          s.MinHeight,
		"margin":              s.Margin,
		"margin-top":          s.MarginTop,
		"margin-bottom":       s.MarginBottom,
		"margin-left":         s.MarginLeft,
		"margin-right":        s.MarginRight,
		"padding":             s.Padding,
		"padding-top":         s.PaddingTop,
		"padding-bottom":      s.PaddingBottom,
		"padding-left":        s.PaddingLeft,
		"padding-right":       s.PaddingRight,
		"font-size":           s.FontSize,
		"font-weight":         s.FontWeight,
		"font-family":         s.FontFamily,
		"line-height":         s.LineHeight,
		"color":               s.Color,
		"text-align":          s.TextAlign,
		"background":          s.Background,
		"background-color":    s.BackgroundColor,
		"border":              s.Border,
		"border-radius":       s.BorderRadius,
		"border-color":        s.BorderColor,
		"border-width":        s.BorderWidth,
		"border-style":        s.BorderStyle,
		"flex":                s.Flex,
		"flex-direction":      s.FlexDirection,
		"flex-wrap":           s.FlexWrap,
		"justify-content":     s.JustifyContent,
		"align-items":         s.AlignItems,
		"box-shadow":          s.BoxShadow,
		"opacity":             s.Opacity,
		"transform":           s.Transform,
		"transition":          s.Transition,
		"transition-duration": s.TransitionDuration,
		"overflow":            s.Overflow,
		"z-index":             s.ZIndex,
		"visibility":          s.Visibility,
	}
}

// toCSSPropertyName converts Go property names to CSS property names
func (s *Styles) toCSSPropertyName(goPropertyName string) string {
	// Properties are already in CSS format from getAllProperties()
	return goPropertyName
}
