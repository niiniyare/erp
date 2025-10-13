package css

import (
	"fmt"
	"strings"
)

// CSSProperty represents a type-safe CSS property value
type CSSProperty interface {
	ToCSS() string
	Validate() error
	PropertyName() string
}

// ============================================================================
// COMMON CSS VALUE TYPES
// ============================================================================

// ColorValue represents CSS color values
type ColorValue struct {
	Value string `json:"value"`
}

func (c ColorValue) ToCSS() string { return c.Value }
func (c ColorValue) Validate() error {
	// Basic validation - in practice would use the schema loader
	if c.Value == "" {
		return fmt.Errorf("color value cannot be empty")
	}
	return nil
}
func (c ColorValue) PropertyName() string { return "color" }

// SizeValue represents CSS size values (with units)
type SizeValue struct {
	Value string `json:"value"`
}

func (s SizeValue) ToCSS() string { return s.Value }
func (s SizeValue) Validate() error {
	if s.Value == "" {
		return fmt.Errorf("size value cannot be empty")
	}
	return nil
}
func (s SizeValue) PropertyName() string { return "size" }

// PositionValue represents CSS position values
type PositionValue struct {
	Value string `json:"value"`
}

func (p PositionValue) ToCSS() string { return p.Value }
func (p PositionValue) Validate() error {
	validPositions := []string{"top", "bottom", "left", "right", "center"}
	for _, valid := range validPositions {
		if p.Value == valid {
			return nil
		}
	}
	// Also allow numeric values and percentages
	if strings.Contains(p.Value, "px") || strings.Contains(p.Value, "%") {
		return nil
	}
	return fmt.Errorf("invalid position value: %s", p.Value)
}
func (p PositionValue) PropertyName() string { return "position" }

// ============================================================================
// STYLES CONTAINER
// ============================================================================

// Styles represents a collection of CSS properties with validation
type Styles struct {
	loader *SchemaLoader

	// Layout Properties
	Display       string `json:"display,omitempty"`
	Position      string `json:"position,omitempty"`
	Top           string `json:"top,omitempty"`
	Bottom        string `json:"bottom,omitempty"`
	Left          string `json:"left,omitempty"`
	Right         string `json:"right,omitempty"`
	Width         string `json:"width,omitempty"`
	Height        string `json:"height,omitempty"`
	MaxWidth      string `json:"max_width,omitempty"`
	MaxHeight     string `json:"max_height,omitempty"`
	MinWidth      string `json:"min_width,omitempty"`
	MinHeight     string `json:"min_height,omitempty"`
	
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
	FontSize      string `json:"font_size,omitempty"`
	FontWeight    string `json:"font_weight,omitempty"`
	FontFamily    string `json:"font_family,omitempty"`
	LineHeight    string `json:"line_height,omitempty"`
	Color         string `json:"color,omitempty"`
	TextAlign     string `json:"text_align,omitempty"`
	
	// Background Properties
	Background      string `json:"background,omitempty"`
	BackgroundColor string `json:"background_color,omitempty"`
	BackgroundImage string `json:"background_image,omitempty"`
	BackgroundSize  string `json:"background_size,omitempty"`
	
	// Border Properties
	Border            string `json:"border,omitempty"`
	BorderRadius      string `json:"border_radius,omitempty"`
	BorderColor       string `json:"border_color,omitempty"`
	BorderWidth       string `json:"border_width,omitempty"`
	BorderStyle       string `json:"border_style,omitempty"`
	
	// Flexbox Properties
	Flex          string `json:"flex,omitempty"`
	FlexDirection string `json:"flex_direction,omitempty"`
	FlexWrap      string `json:"flex_wrap,omitempty"`
	FlexGrow      string `json:"flex_grow,omitempty"`
	FlexShrink    string `json:"flex_shrink,omitempty"`
	FlexBasis     string `json:"flex_basis,omitempty"`
	JustifyContent string `json:"justify_content,omitempty"`
	AlignItems    string `json:"align_items,omitempty"`
	AlignContent  string `json:"align_content,omitempty"`
	
	// Grid Properties
	GridTemplateColumns string `json:"grid_template_columns,omitempty"`
	GridTemplateRows    string `json:"grid_template_rows,omitempty"`
	GridGap             string `json:"grid_gap,omitempty"`
	GridColumnGap       string `json:"grid_column_gap,omitempty"`
	GridRowGap          string `json:"grid_row_gap,omitempty"`
	
	// Visual Effects
	BoxShadow string `json:"box_shadow,omitempty"`
	Opacity   string `json:"opacity,omitempty"`
	Transform string `json:"transform,omitempty"`
	Filter    string `json:"filter,omitempty"`
	
	// Animation Properties
	Transition           string `json:"transition,omitempty"`
	TransitionDuration   string `json:"transition_duration,omitempty"`
	TransitionProperty   string `json:"transition_property,omitempty"`
	TransitionTimingFunction string `json:"transition_timing_function,omitempty"`
	Animation            string `json:"animation,omitempty"`
	AnimationDuration    string `json:"animation_duration,omitempty"`
	AnimationName        string `json:"animation_name,omitempty"`
	
	// Overflow Properties
	Overflow  string `json:"overflow,omitempty"`
	OverflowX string `json:"overflow_x,omitempty"`
	OverflowY string `json:"overflow_y,omitempty"`
	
	// Z-Index and Stacking
	ZIndex string `json:"z_index,omitempty"`
	
	// Visibility
	Visibility string `json:"visibility,omitempty"`
	
	// Custom properties for framework-specific values
	Custom map[string]string `json:"custom,omitempty"`
}

// NewStyles creates a new Styles instance with schema validation
func NewStyles(schemaDir string) *Styles {
	return &Styles{
		loader: NewSchemaLoader(schemaDir),
		Custom: make(map[string]string),
	}
}

// ============================================================================
// VALIDATION METHODS
// ============================================================================

// Validate validates all CSS properties against their schemas
func (s *Styles) Validate() error {
	properties := s.getAllProperties()
	
	for propertyName, value := range properties {
		if value == "" {
			continue // Skip empty values
		}
		
		if err := s.loader.ValidateValue(propertyName, value); err != nil {
			return fmt.Errorf("invalid value for property '%s': %w", propertyName, err)
		}
	}
	
	return nil
}

// ValidateProperty validates a specific CSS property
func (s *Styles) ValidateProperty(propertyName, value string) error {
	return s.loader.ValidateValue(propertyName, value)
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

// WithFontFamily sets the font-family property
func (s *Styles) WithFontFamily(family string) *Styles {
	s.FontFamily = family
	return s
}

// WithLineHeight sets the line-height property
func (s *Styles) WithLineHeight(height string) *Styles {
	s.LineHeight = height
	return s
}

// WithTextAlign sets the text-align property
func (s *Styles) WithTextAlign(align string) *Styles {
	s.TextAlign = align
	return s
}

// WithBackground sets the background property
func (s *Styles) WithBackground(background string) *Styles {
	s.Background = background
	return s
}

// WithBackgroundImage sets the background-image property
func (s *Styles) WithBackgroundImage(image string) *Styles {
	s.BackgroundImage = image
	return s
}

// WithBackgroundSize sets the background-size property
func (s *Styles) WithBackgroundSize(size string) *Styles {
	s.BackgroundSize = size
	return s
}

// WithBorder sets the border property
func (s *Styles) WithBorder(border string) *Styles {
	s.Border = border
	return s
}

// WithBorderColor sets the border-color property
func (s *Styles) WithBorderColor(color string) *Styles {
	s.BorderColor = color
	return s
}

// WithBorderWidth sets the border-width property
func (s *Styles) WithBorderWidth(width string) *Styles {
	s.BorderWidth = width
	return s
}

// WithBorderStyle sets the border-style property
func (s *Styles) WithBorderStyle(style string) *Styles {
	s.BorderStyle = style
	return s
}

// WithFlexWrap sets the flex-wrap property
func (s *Styles) WithFlexWrap(wrap string) *Styles {
	s.FlexWrap = wrap
	return s
}

// WithFlexGrow sets the flex-grow property
func (s *Styles) WithFlexGrow(grow string) *Styles {
	s.FlexGrow = grow
	return s
}

// WithFlexShrink sets the flex-shrink property
func (s *Styles) WithFlexShrink(shrink string) *Styles {
	s.FlexShrink = shrink
	return s
}

// WithFlexBasis sets the flex-basis property
func (s *Styles) WithFlexBasis(basis string) *Styles {
	s.FlexBasis = basis
	return s
}

// WithAlignContent sets the align-content property
func (s *Styles) WithAlignContent(align string) *Styles {
	s.AlignContent = align
	return s
}

// WithGridTemplateColumns sets the grid-template-columns property
func (s *Styles) WithGridTemplateColumns(columns string) *Styles {
	s.GridTemplateColumns = columns
	return s
}

// WithGridTemplateRows sets the grid-template-rows property
func (s *Styles) WithGridTemplateRows(rows string) *Styles {
	s.GridTemplateRows = rows
	return s
}

// WithGridGap sets the grid-gap property
func (s *Styles) WithGridGap(gap string) *Styles {
	s.GridGap = gap
	return s
}

// WithGridColumnGap sets the grid-column-gap property
func (s *Styles) WithGridColumnGap(gap string) *Styles {
	s.GridColumnGap = gap
	return s
}

// WithGridRowGap sets the grid-row-gap property
func (s *Styles) WithGridRowGap(gap string) *Styles {
	s.GridRowGap = gap
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

// WithTransform sets the transform property
func (s *Styles) WithTransform(transform string) *Styles {
	s.Transform = transform
	return s
}

// WithFilter sets the filter property
func (s *Styles) WithFilter(filter string) *Styles {
	s.Filter = filter
	return s
}

// WithTransition sets the transition property
func (s *Styles) WithTransition(transition string) *Styles {
	s.Transition = transition
	return s
}

// WithTransitionDuration sets the transition-duration property
func (s *Styles) WithTransitionDuration(duration string) *Styles {
	s.TransitionDuration = duration
	return s
}

// WithTransitionProperty sets the transition-property property
func (s *Styles) WithTransitionProperty(property string) *Styles {
	s.TransitionProperty = property
	return s
}

// WithTransitionTimingFunction sets the transition-timing-function property
func (s *Styles) WithTransitionTimingFunction(timing string) *Styles {
	s.TransitionTimingFunction = timing
	return s
}

// WithAnimation sets the animation property
func (s *Styles) WithAnimation(animation string) *Styles {
	s.Animation = animation
	return s
}

// WithAnimationDuration sets the animation-duration property
func (s *Styles) WithAnimationDuration(duration string) *Styles {
	s.AnimationDuration = duration
	return s
}

// WithAnimationName sets the animation-name property
func (s *Styles) WithAnimationName(name string) *Styles {
	s.AnimationName = name
	return s
}

// WithOverflow sets the overflow property
func (s *Styles) WithOverflow(overflow string) *Styles {
	s.Overflow = overflow
	return s
}

// WithOverflowX sets the overflow-x property
func (s *Styles) WithOverflowX(overflow string) *Styles {
	s.OverflowX = overflow
	return s
}

// WithOverflowY sets the overflow-y property
func (s *Styles) WithOverflowY(overflow string) *Styles {
	s.OverflowY = overflow
	return s
}

// WithZIndex sets the z-index property
func (s *Styles) WithZIndex(index string) *Styles {
	s.ZIndex = index
	return s
}

// WithVisibility sets the visibility property
func (s *Styles) WithVisibility(visibility string) *Styles {
	s.Visibility = visibility
	return s
}

// WithMaxWidth sets the max-width property
func (s *Styles) WithMaxWidth(width string) *Styles {
	s.MaxWidth = width
	return s
}

// WithMaxHeight sets the max-height property
func (s *Styles) WithMaxHeight(height string) *Styles {
	s.MaxHeight = height
	return s
}

// WithMinWidth sets the min-width property
func (s *Styles) WithMinWidth(width string) *Styles {
	s.MinWidth = width
	return s
}

// WithMinHeight sets the min-height property
func (s *Styles) WithMinHeight(height string) *Styles {
	s.MinHeight = height
	return s
}

// WithTop sets the top property
func (s *Styles) WithTop(top string) *Styles {
	s.Top = top
	return s
}

// WithBottom sets the bottom property
func (s *Styles) WithBottom(bottom string) *Styles {
	s.Bottom = bottom
	return s
}

// WithLeft sets the left property
func (s *Styles) WithLeft(left string) *Styles {
	s.Left = left
	return s
}

// WithRight sets the right property
func (s *Styles) WithRight(right string) *Styles {
	s.Right = right
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

// GetAllProperties returns all non-empty CSS properties as a map (public accessor)
func (s *Styles) GetAllProperties() map[string]string {
	return s.getAllProperties()
}

// getAllProperties returns all non-empty CSS properties as a map
func (s *Styles) getAllProperties() map[string]string {
	return map[string]string{
		"display":         s.Display,
		"position":        s.Position,
		"top":            s.Top,
		"bottom":         s.Bottom,
		"left":           s.Left,
		"right":          s.Right,
		"width":          s.Width,
		"height":         s.Height,
		"max-width":      s.MaxWidth,
		"max-height":     s.MaxHeight,
		"min-width":      s.MinWidth,
		"min-height":     s.MinHeight,
		"margin":         s.Margin,
		"margin-top":     s.MarginTop,
		"margin-bottom":  s.MarginBottom,
		"margin-left":    s.MarginLeft,
		"margin-right":   s.MarginRight,
		"padding":        s.Padding,
		"padding-top":    s.PaddingTop,
		"padding-bottom": s.PaddingBottom,
		"padding-left":   s.PaddingLeft,
		"padding-right":  s.PaddingRight,
		"font-size":      s.FontSize,
		"font-weight":    s.FontWeight,
		"font-family":    s.FontFamily,
		"line-height":    s.LineHeight,
		"color":          s.Color,
		"text-align":     s.TextAlign,
		"background":          s.Background,
		"background-color":    s.BackgroundColor,
		"background-image":    s.BackgroundImage,
		"background-size":     s.BackgroundSize,
		"border":              s.Border,
		"border-radius":       s.BorderRadius,
		"border-color":        s.BorderColor,
		"border-width":        s.BorderWidth,
		"border-style":        s.BorderStyle,
		"flex":                s.Flex,
		"flex-direction":      s.FlexDirection,
		"flex-wrap":           s.FlexWrap,
		"flex-grow":           s.FlexGrow,
		"flex-shrink":         s.FlexShrink,
		"flex-basis":          s.FlexBasis,
		"justify-content":     s.JustifyContent,
		"align-items":         s.AlignItems,
		"align-content":       s.AlignContent,
		"grid-template-columns": s.GridTemplateColumns,
		"grid-template-rows":    s.GridTemplateRows,
		"grid-gap":              s.GridGap,
		"grid-column-gap":       s.GridColumnGap,
		"grid-row-gap":          s.GridRowGap,
		"box-shadow":         s.BoxShadow,
		"opacity":            s.Opacity,
		"transform":          s.Transform,
		"filter":             s.Filter,
		"transition":         s.Transition,
		"transition-duration": s.TransitionDuration,
		"transition-property": s.TransitionProperty,
		"transition-timing-function": s.TransitionTimingFunction,
		"animation":          s.Animation,
		"animation-duration": s.AnimationDuration,
		"animation-name":     s.AnimationName,
		"overflow":           s.Overflow,
		"overflow-x":         s.OverflowX,
		"overflow-y":         s.OverflowY,
		"z-index":            s.ZIndex,
		"visibility":         s.Visibility,
	}
}

// toCSSPropertyName converts Go property names to CSS property names
func (s *Styles) toCSSPropertyName(goPropertyName string) string {
	// Most properties are already in CSS format from getAllProperties()
	return goPropertyName
}

// ============================================================================
// PRESET STYLES FOR COMMON PATTERNS
// ============================================================================

// FlexCenter creates styles for centered flex container
func FlexCenter() *Styles {
	return &Styles{
		Display:        "flex",
		JustifyContent: "center",
		AlignItems:     "center",
	}
}

// Card creates styles for a typical card component
func Card() *Styles {
	return &Styles{
		Background:    "#ffffff",
		BorderRadius:  "8px",
		BoxShadow:     "0 2px 4px rgba(0, 0, 0, 0.1)",
		Padding:       "1rem",
	}
}

// Button creates styles for a button component
func Button() *Styles {
	return &Styles{
		Display:       "inline-flex",
		AlignItems:    "center",
		JustifyContent: "center",
		Padding:       "0.5rem 1rem",
		BorderRadius:  "4px",
		Border:        "none",
		FontWeight:    "500",
		Transition:    "all 0.2s ease-in-out",
	}
}

// Input creates styles for form input components
func Input() *Styles {
	return &Styles{
		Display:      "block",
		Width:        "100%",
		Padding:      "0.5rem",
		BorderRadius: "4px",
		Border:       "1px solid #d1d5db",
		FontSize:     "1rem",
		Transition:   "border-color 0.15s ease-in-out, box-shadow 0.15s ease-in-out",
	}
}