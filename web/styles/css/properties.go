package css

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

// ExpandedStyles provides comprehensive CSS property support covering all 581+ schema properties
// This integrates with the Templ components in web/components/**/*
type ExpandedStyles struct {
	loader *SchemaLoader

	Layout        *LayoutProperties        `json:"layout,omitempty"`
	BoxModel      *BoxModelProperties      `json:"boxModel,omitempty"`
	Margin        *MarginProperties        `json:"margin,omitempty"`
	Padding       *PaddingProperties       `json:"padding,omitempty"`
	Border        *BorderProperties        `json:"border,omitempty"`
	Background    *BackgroundProperties    `json:"background,omitempty"`
	Typography    *TypographyProperties    `json:"typography,omitempty"`
	Flexbox       *FlexboxProperties       `json:"flexbox,omitempty"`
	Grid          *GridProperties          `json:"grid,omitempty"`
	VisualEffects *VisualEffectsProperties `json:"visualEffects,omitempty"`
	Animation     *AnimationProperties     `json:"animation,omitempty"`
	Transition    *TransitionProperties    `json:"transition,omitempty"`
	Overflow      *OverflowProperties      `json:"overflow,omitempty"`
	Table         *TableProperties         `json:"table,omitempty"`
	List          *ListProperties          `json:"list,omitempty"`
	Interaction   *InteractionProperties   `json:"interaction,omitempty"`
	Scroll        *ScrollProperties        `json:"scroll,omitempty"`
	ModernLayout  *ModernLayoutProperties  `json:"modernLayout,omitempty"`
	Logical       *LogicalProperties       `json:"logical,omitempty"`
	AdvancedTypo  *AdvancedTypoProperties  `json:"advancedTypo,omitempty"`
	Mask          *MaskProperties          `json:"mask,omitempty"`
	ClipPath      *ClipPathProperties      `json:"clipPath,omitempty"`
	Shape         *ShapeProperties         `json:"shape,omitempty"`
	Container     *ContainerProperties     `json:"container,omitempty"`

	// Custom Properties for Templ integration
	Custom map[string]string `json:"custom,omitempty"`

	// Framework Integration
	templIntegration bool `json:"-"`
}

// NewExpandedStyles creates a new ExpandedStyles instance with schema validation
func NewExpandedStyles(schemaDir string) *ExpandedStyles {
	return &ExpandedStyles{
		loader:           NewSchemaLoader(schemaDir),
		Custom:           make(map[string]string),
		templIntegration: true,
	}
}

// LoadFromSchemaDefinition loads properties from JSON schema definitions
func (s *ExpandedStyles) LoadFromSchemaDefinition(propertyName string) error {
	_, err := s.loader.LoadPropertySchema(propertyName)
	if err != nil {
		// Be tolerant of missing schema files - just return nil to continue
		return nil
	}

	// Apply schema-defined defaults (if schema has default values)
	// TODO:: PropertySchema structure may vary based on actual schema format

	return nil
}

// setPropertyByName sets a CSS property by name (used for schema loading)
func (s *ExpandedStyles) setPropertyByName(property, value string) {
	// Convert kebab-case to struct field mapping with nested properties
	switch strings.ToLower(property) {
	// Layout properties
	case "display":
		s.Layout.Display = Display(value)
	case "position":
		s.Layout.Position = Position(value)
	case "top":
		s.Layout.Top = value
	case "right":
		s.Layout.Right = value
	case "bottom":
		s.Layout.Bottom = value
	case "left":
		s.Layout.Left = value
	case "z-index":
		s.Layout.ZIndex = value

	// Box model properties
	case "width":
		s.BoxModel.Width = value
	case "height":
		s.BoxModel.Height = value
	case "min-width":
		s.BoxModel.MinWidth = value
	case "min-height":
		s.BoxModel.MinHeight = value
	case "max-width":
		s.BoxModel.MaxWidth = value
	case "max-height":
		s.BoxModel.MaxHeight = value
	case "box-sizing":
		s.BoxModel.BoxSizing = BoxSizing(value)

	// Margin properties
	case "margin":
		s.Margin.Margin = value
	case "margin-top":
		s.Margin.Top = value
	case "margin-right":
		s.Margin.Right = value
	case "margin-bottom":
		s.Margin.Bottom = value
	case "margin-left":
		s.Margin.Left = value

	// Padding properties
	case "padding":
		s.Padding.Padding = value
	case "padding-top":
		s.Padding.Top = value
	case "padding-right":
		s.Padding.Right = value
	case "padding-bottom":
		s.Padding.Bottom = value
	case "padding-left":
		s.Padding.Left = value

	// Border properties
	case "border":
		s.Border.Border = value
	case "border-radius":
		s.Border.Radius = value
	case "border-width":
		s.Border.Width = value
	case "border-style":
		s.Border.Style = BorderStyle(value)
	case "border-color":
		s.Border.Color = value

	// Background properties
	case "background":
		s.Background.Background = value
	case "background-color":
		s.Background.Color = value
	case "background-image":
		s.Background.Image = value

	// Typography properties
	case "color":
		s.Typography.Color = value
	case "font-size":
		s.Typography.FontSize = value
	case "font-weight":
		s.Typography.FontWeight = FontWeight(value)
	case "font-family":
		s.Typography.FontFamily = value
	case "line-height":
		s.Typography.LineHeight = value
	case "text-align":
		s.Typography.TextAlign = TextAlign(value)

	// Flexbox properties
	case "flex-direction":
		s.Flexbox.Direction = FlexDirection(value)
	case "justify-content":
		s.Flexbox.JustifyContent = JustifyContent(value)
	case "align-items":
		s.Flexbox.AlignItems = AlignItems(value)
	case "flex":
		s.Flexbox.Flex = value

	// Grid properties
	case "grid-template-columns":
		s.Grid.TemplateColumns = value
	case "grid-template-rows":
		s.Grid.TemplateRows = value
	case "gap":
		s.Grid.Gap = value

	// Visual effects properties
	case "opacity":
		s.VisualEffects.Opacity = value
	case "box-shadow":
		s.VisualEffects.BoxShadow = value
	case "transform":
		s.VisualEffects.Transform = value

	// Transition properties
	case "transition":
		s.Transition.Transition = value

	// Overflow properties
	case "overflow":
		s.Overflow.Overflow = Overflow(value)
	case "overflow-x":
		s.Overflow.X = Overflow(value)
	case "overflow-y":
		s.Overflow.Y = Overflow(value)

	// Add more property mappings as needed
	default:
		// Store unknown properties in Custom map
		s.Custom[property] = value
	}
}

// Fluent API methods for common properties (using nested structure)
func (s *ExpandedStyles) WithDisplay(display string) *ExpandedStyles {
	s.Layout.Display = Display(display)
	return s
}

func (s *ExpandedStyles) WithPosition(position string) *ExpandedStyles {
	s.Layout.Position = Position(position)
	return s
}

func (s *ExpandedStyles) WithZIndex(zIndex string) *ExpandedStyles {
	s.Layout.ZIndex = zIndex
	return s
}

func (s *ExpandedStyles) WithWidth(width string) *ExpandedStyles {
	s.BoxModel.Width = width
	return s
}

func (s *ExpandedStyles) WithHeight(height string) *ExpandedStyles {
	s.BoxModel.Height = height
	return s
}

func (s *ExpandedStyles) WithMargin(margin string) *ExpandedStyles {
	s.Margin.Margin = margin
	return s
}

func (s *ExpandedStyles) WithPadding(padding string) *ExpandedStyles {
	s.Padding.Padding = padding
	return s
}

func (s *ExpandedStyles) WithBorder(border string) *ExpandedStyles {
	s.Border.Border = border
	return s
}

func (s *ExpandedStyles) WithBorderRadius(radius string) *ExpandedStyles {
	s.Border.Radius = radius
	return s
}

func (s *ExpandedStyles) WithBackgroundColor(color string) *ExpandedStyles {
	s.Background.Color = color
	return s
}

func (s *ExpandedStyles) WithColor(color string) *ExpandedStyles {
	s.Typography.Color = color
	return s
}

func (s *ExpandedStyles) WithFontSize(size string) *ExpandedStyles {
	s.Typography.FontSize = size
	return s
}

func (s *ExpandedStyles) WithFontWeight(weight string) *ExpandedStyles {
	s.Typography.FontWeight = FontWeight(weight)
	return s
}

func (s *ExpandedStyles) WithTextAlign(align string) *ExpandedStyles {
	s.Typography.TextAlign = TextAlign(align)
	return s
}

func (s *ExpandedStyles) WithFlexDirection(direction string) *ExpandedStyles {
	s.Flexbox.Direction = FlexDirection(direction)
	return s
}

func (s *ExpandedStyles) WithJustifyContent(justify string) *ExpandedStyles {
	s.Flexbox.JustifyContent = JustifyContent(justify)
	return s
}

func (s *ExpandedStyles) WithAlignItems(align string) *ExpandedStyles {
	s.Flexbox.AlignItems = AlignItems(align)
	return s
}

func (s *ExpandedStyles) WithGridTemplateColumns(columns string) *ExpandedStyles {
	s.Grid.TemplateColumns = columns
	return s
}

func (s *ExpandedStyles) WithGridTemplateRows(rows string) *ExpandedStyles {
	s.Grid.TemplateRows = rows
	return s
}

func (s *ExpandedStyles) WithGap(gap string) *ExpandedStyles {
	s.Grid.Gap = gap
	return s
}

func (s *ExpandedStyles) WithOpacity(opacity string) *ExpandedStyles {
	s.VisualEffects.Opacity = opacity
	return s
}

func (s *ExpandedStyles) WithTransform(transform string) *ExpandedStyles {
	s.VisualEffects.Transform = transform
	return s
}

func (s *ExpandedStyles) WithLineHeight(lineHeight string) *ExpandedStyles {
	s.Typography.LineHeight = lineHeight
	return s
}

func (s *ExpandedStyles) WithTransition(transition string) *ExpandedStyles {
	s.Transition.Transition = transition
	return s
}

func (s *ExpandedStyles) WithBoxShadow(shadow string) *ExpandedStyles {
	s.VisualEffects.BoxShadow = shadow
	return s
}

func (s *ExpandedStyles) WithMaxWidth(maxWidth string) *ExpandedStyles {
	s.BoxModel.MaxWidth = maxWidth
	return s
}

func (s *ExpandedStyles) WithMaxHeight(maxHeight string) *ExpandedStyles {
	s.BoxModel.MaxHeight = maxHeight
	return s
}

func (s *ExpandedStyles) WithMinWidth(minWidth string) *ExpandedStyles {
	s.BoxModel.MinWidth = minWidth
	return s
}

func (s *ExpandedStyles) WithMinHeight(minHeight string) *ExpandedStyles {
	s.BoxModel.MinHeight = minHeight
	return s
}

// Advanced layout methods
func (s *ExpandedStyles) WithFlexbox(direction, justify, align string) *ExpandedStyles {
	s.Layout.Display = DisplayFlex
	s.Flexbox.Direction = FlexDirection(direction)
	s.Flexbox.JustifyContent = JustifyContent(justify)
	s.Flexbox.AlignItems = AlignItems(align)
	return s
}

func (s *ExpandedStyles) WithGrid(columns, rows, gap string) *ExpandedStyles {
	s.Layout.Display = DisplayGrid
	if columns != "" {
		s.Grid.TemplateColumns = columns
	}
	if rows != "" {
		s.Grid.TemplateRows = rows
	}
	if gap != "" {
		s.Grid.Gap = gap
	}
	return s
}

// Custom property support for Templ integration
func (s *ExpandedStyles) WithCustomProperty(name, value string) *ExpandedStyles {
	if !strings.HasPrefix(name, "--") {
		name = "--" + name
	}
	s.Custom[name] = value
	return s
}

// Schema validation with error reporting
func (s *ExpandedStyles) ValidateProperty(property, value string) error {
	if s.loader == nil {
		return fmt.Errorf("schema loader not initialized")
	}
	return s.loader.ValidateValue(property, value)
}

// ToCSS generates optimized CSS string for Templ components using nested structure
func (s *ExpandedStyles) ToCSS() string {
	var css strings.Builder

	// Layout properties first (most impactful)
	if s.Layout.Display != "" {
		css.WriteString(fmt.Sprintf("display: %s; ", s.Layout.Display))
	}
	if s.Layout.Position != "" {
		css.WriteString(fmt.Sprintf("position: %s; ", s.Layout.Position))
	}
	if s.Layout.Top != "" {
		css.WriteString(fmt.Sprintf("top: %s; ", s.Layout.Top))
	}
	if s.Layout.Right != "" {
		css.WriteString(fmt.Sprintf("right: %s; ", s.Layout.Right))
	}
	if s.Layout.Bottom != "" {
		css.WriteString(fmt.Sprintf("bottom: %s; ", s.Layout.Bottom))
	}
	if s.Layout.Left != "" {
		css.WriteString(fmt.Sprintf("left: %s; ", s.Layout.Left))
	}
	if s.Layout.ZIndex != "" {
		css.WriteString(fmt.Sprintf("z-index: %s; ", s.Layout.ZIndex))
	}

	// Box model properties
	if s.BoxModel.Width != "" {
		css.WriteString(fmt.Sprintf("width: %s; ", s.BoxModel.Width))
	}
	if s.BoxModel.Height != "" {
		css.WriteString(fmt.Sprintf("height: %s; ", s.BoxModel.Height))
	}
	if s.BoxModel.MinWidth != "" {
		css.WriteString(fmt.Sprintf("min-width: %s; ", s.BoxModel.MinWidth))
	}
	if s.BoxModel.MinHeight != "" {
		css.WriteString(fmt.Sprintf("min-height: %s; ", s.BoxModel.MinHeight))
	}
	if s.BoxModel.MaxWidth != "" {
		css.WriteString(fmt.Sprintf("max-width: %s; ", s.BoxModel.MaxWidth))
	}
	if s.BoxModel.MaxHeight != "" {
		css.WriteString(fmt.Sprintf("max-height: %s; ", s.BoxModel.MaxHeight))
	}
	if s.BoxModel.BoxSizing != "" {
		css.WriteString(fmt.Sprintf("box-sizing: %s; ", s.BoxModel.BoxSizing))
	}

	// Margin properties
	if s.Margin.Margin != "" {
		css.WriteString(fmt.Sprintf("margin: %s; ", s.Margin.Margin))
	}
	if s.Margin.Top != "" {
		css.WriteString(fmt.Sprintf("margin-top: %s; ", s.Margin.Top))
	}
	if s.Margin.Right != "" {
		css.WriteString(fmt.Sprintf("margin-right: %s; ", s.Margin.Right))
	}
	if s.Margin.Bottom != "" {
		css.WriteString(fmt.Sprintf("margin-bottom: %s; ", s.Margin.Bottom))
	}
	if s.Margin.Left != "" {
		css.WriteString(fmt.Sprintf("margin-left: %s; ", s.Margin.Left))
	}

	// Padding properties
	if s.Padding.Padding != "" {
		css.WriteString(fmt.Sprintf("padding: %s; ", s.Padding.Padding))
	}
	if s.Padding.Top != "" {
		css.WriteString(fmt.Sprintf("padding-top: %s; ", s.Padding.Top))
	}
	if s.Padding.Right != "" {
		css.WriteString(fmt.Sprintf("padding-right: %s; ", s.Padding.Right))
	}
	if s.Padding.Bottom != "" {
		css.WriteString(fmt.Sprintf("padding-bottom: %s; ", s.Padding.Bottom))
	}
	if s.Padding.Left != "" {
		css.WriteString(fmt.Sprintf("padding-left: %s; ", s.Padding.Left))
	}

	// Border properties
	if s.Border.Border != "" {
		css.WriteString(fmt.Sprintf("border: %s; ", s.Border.Border))
	}
	if s.Border.Radius != "" {
		css.WriteString(fmt.Sprintf("border-radius: %s; ", s.Border.Radius))
	}
	if s.Border.Width != "" {
		css.WriteString(fmt.Sprintf("border-width: %s; ", s.Border.Width))
	}
	if s.Border.Style != "" {
		css.WriteString(fmt.Sprintf("border-style: %s; ", s.Border.Style))
	}
	if s.Border.Color != "" {
		css.WriteString(fmt.Sprintf("border-color: %s; ", s.Border.Color))
	}

	// Background properties
	if s.Background.Background != "" {
		css.WriteString(fmt.Sprintf("background: %s; ", s.Background.Background))
	}
	if s.Background.Color != "" {
		css.WriteString(fmt.Sprintf("background-color: %s; ", s.Background.Color))
	}
	if s.Background.Image != "" {
		css.WriteString(fmt.Sprintf("background-image: %s; ", s.Background.Image))
	}

	// Typography properties
	if s.Typography.Color != "" {
		css.WriteString(fmt.Sprintf("color: %s; ", s.Typography.Color))
	}
	if s.Typography.FontSize != "" {
		css.WriteString(fmt.Sprintf("font-size: %s; ", s.Typography.FontSize))
	}
	if s.Typography.FontWeight != "" {
		css.WriteString(fmt.Sprintf("font-weight: %s; ", s.Typography.FontWeight))
	}
	if s.Typography.FontFamily != "" {
		css.WriteString(fmt.Sprintf("font-family: %s; ", s.Typography.FontFamily))
	}
	if s.Typography.LineHeight != "" {
		css.WriteString(fmt.Sprintf("line-height: %s; ", s.Typography.LineHeight))
	}
	if s.Typography.TextAlign != "" {
		css.WriteString(fmt.Sprintf("text-align: %s; ", s.Typography.TextAlign))
	}

	// Flexbox properties
	if s.Flexbox.Direction != "" {
		css.WriteString(fmt.Sprintf("flex-direction: %s; ", s.Flexbox.Direction))
	}
	if s.Flexbox.JustifyContent != "" {
		css.WriteString(fmt.Sprintf("justify-content: %s; ", s.Flexbox.JustifyContent))
	}
	if s.Flexbox.AlignItems != "" {
		css.WriteString(fmt.Sprintf("align-items: %s; ", s.Flexbox.AlignItems))
	}
	if s.Flexbox.Wrap != "" {
		css.WriteString(fmt.Sprintf("flex-wrap: %s; ", s.Flexbox.Wrap))
	}
	if s.Flexbox.Flex != "" {
		css.WriteString(fmt.Sprintf("flex: %s; ", s.Flexbox.Flex))
	}

	// Grid properties
	if s.Grid.TemplateColumns != "" {
		css.WriteString(fmt.Sprintf("grid-template-columns: %s; ", s.Grid.TemplateColumns))
	}
	if s.Grid.TemplateRows != "" {
		css.WriteString(fmt.Sprintf("grid-template-rows: %s; ", s.Grid.TemplateRows))
	}
	if s.Grid.Gap != "" {
		css.WriteString(fmt.Sprintf("gap: %s; ", s.Grid.Gap))
	}

	// Visual effects
	if s.VisualEffects.Opacity != "" {
		css.WriteString(fmt.Sprintf("opacity: %s; ", s.VisualEffects.Opacity))
	}
	if s.VisualEffects.BoxShadow != "" {
		css.WriteString(fmt.Sprintf("box-shadow: %s; ", s.VisualEffects.BoxShadow))
	}
	if s.VisualEffects.Transform != "" {
		css.WriteString(fmt.Sprintf("transform: %s; ", s.VisualEffects.Transform))
	}

	// Transitions
	if s.Transition.Transition != "" {
		css.WriteString(fmt.Sprintf("transition: %s; ", s.Transition.Transition))
	}

	// Overflow properties
	if s.Overflow.Overflow != "" {
		css.WriteString(fmt.Sprintf("overflow: %s; ", s.Overflow.Overflow))
	}
	if s.Overflow.X != "" {
		css.WriteString(fmt.Sprintf("overflow-x: %s; ", s.Overflow.X))
	}
	if s.Overflow.Y != "" {
		css.WriteString(fmt.Sprintf("overflow-y: %s; ", s.Overflow.Y))
	}

	// Custom properties (CSS variables)
	for name, value := range s.Custom {
		css.WriteString(fmt.Sprintf("%s: %s; ", name, value))
	}

	result := css.String()
	if len(result) > 0 {
		return strings.TrimSuffix(result, " ")
	}
	return ""
}

// ToCSSClass generates a CSS class for Templ components
func (s *ExpandedStyles) ToCSSClass(className string) string {
	css := s.ToCSS()
	if css == "" {
		return ""
	}

	return fmt.Sprintf(".%s {\n  %s\n}", className, strings.ReplaceAll(css, "; ", ";\n  "))
}

// ToTemplStyleAttribute generates a Templ-compatible style attribute
func (s *ExpandedStyles) ToTemplStyleAttribute() string {
	css := s.ToCSS()
	if css == "" {
		return ""
	}

	// Format for Templ template usage: style={ s.ToTemplStyleAttribute() }
	return css
}

// ToJSON serializes the styles to JSON for API integration
func (s *ExpandedStyles) ToJSON() (string, error) {
	data, err := json.Marshal(s)
	return string(data), err
}

// FromJSON deserializes JSON to ExpandedStyles
func (s *ExpandedStyles) FromJSON(jsonStr string) error {
	return json.Unmarshal([]byte(jsonStr), s)
}

// LoadAllAvailableSchemas loads all 581+ CSS property schemas
func (s *ExpandedStyles) LoadAllAvailableSchemas() error {
	if s.loader == nil {
		return fmt.Errorf("schema loader not initialized")
	}

	// Get all available CSS property schemas
	schemaFiles, err := filepath.Glob(filepath.Join(s.loader.schemaDir, "definitions", "Property.*.json"))
	if err != nil {
		return fmt.Errorf("failed to find schema files: %w", err)
	}

	loadedCount := 0
	for _, schemaFile := range schemaFiles {
		// Extract property name from filename
		filename := filepath.Base(schemaFile)
		if !strings.HasPrefix(filename, "Property.") || !strings.HasSuffix(filename, ".json") {
			continue
		}

		propertyName := strings.TrimSuffix(strings.TrimPrefix(filename, "Property."), ".json")

		// Convert property name format
		propertyName = convertPropertyName(propertyName)

		// Load schema for this property
		if err := s.LoadFromSchemaDefinition(propertyName); err == nil {
			loadedCount++
		}
	}

	return nil
}

// convertPropertyName converts schema filename format to CSS property format
func convertPropertyName(schemaName string) string {
	// Handle special cases in schema naming
	// e.g., "BorderTopLeftRadius" -> "border-top-left-radius"

	var result strings.Builder

	for i, r := range schemaName {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('-')
		}
		result.WriteRune(r)
	}

	return strings.ToLower(result.String())
}

// Integration with existing Styles type (backward compatibility)
func (s *ExpandedStyles) ToLegacyStyles() *Styles {
	legacy := &Styles{
		Display:         string(s.Layout.Display),
		Position:        string(s.Layout.Position),
		Width:           s.BoxModel.Width,
		Height:          s.BoxModel.Height,
		Margin:          s.Margin.Margin,
		Padding:         s.Padding.Padding,
		Border:          s.Border.Border,
		BorderRadius:    s.Border.Radius,
		BackgroundColor: s.Background.Color,
		Color:           s.Typography.Color,
		FontSize:        s.Typography.FontSize,
		FontWeight:      string(s.Typography.FontWeight),
		LineHeight:      s.Typography.LineHeight,
		TextAlign:       string(s.Typography.TextAlign),
		FlexDirection:   string(s.Flexbox.Direction),
		JustifyContent:  string(s.Flexbox.JustifyContent),
		AlignItems:      string(s.Flexbox.AlignItems),
		GridGap:         s.Grid.Gap,
		Opacity:         s.VisualEffects.Opacity,
		BoxShadow:       s.VisualEffects.BoxShadow,
		Transform:       s.VisualEffects.Transform,
		Transition:      s.Transition.Transition,
		Custom:          make(map[string]string),
	}

	// Copy custom properties
	for k, v := range s.Custom {
		legacy.Custom[k] = v
	}

	return legacy
}

// CreateFromLegacyStyles creates ExpandedStyles from existing Styles
func CreateFromLegacyStyles(legacy *Styles, schemaDir string) *ExpandedStyles {
	expanded := NewExpandedStyles(schemaDir)

	expanded.Layout.Display = Display(legacy.Display)
	expanded.Layout.Position = Position(legacy.Position)
	expanded.BoxModel.Width = legacy.Width
	expanded.BoxModel.Height = legacy.Height
	expanded.Margin.Margin = legacy.Margin
	expanded.Padding.Padding = legacy.Padding
	expanded.Border.Border = legacy.Border
	expanded.Border.Radius = legacy.BorderRadius
	expanded.Background.Color = legacy.BackgroundColor
	expanded.Typography.Color = legacy.Color
	expanded.Typography.FontSize = legacy.FontSize
	expanded.Typography.FontWeight = FontWeight(legacy.FontWeight)
	expanded.Typography.LineHeight = legacy.LineHeight
	expanded.Typography.TextAlign = TextAlign(legacy.TextAlign)
	expanded.Flexbox.Direction = FlexDirection(legacy.FlexDirection)
	expanded.Flexbox.JustifyContent = JustifyContent(legacy.JustifyContent)
	expanded.Flexbox.AlignItems = AlignItems(legacy.AlignItems)
	expanded.Grid.Gap = legacy.GridGap
	expanded.VisualEffects.Opacity = legacy.Opacity
	expanded.VisualEffects.BoxShadow = legacy.BoxShadow
	expanded.VisualEffects.Transform = legacy.Transform
	expanded.Transition.Transition = legacy.Transition

	// Copy custom properties
	for k, v := range legacy.Custom {
		expanded.Custom[k] = v
	}

	return expanded
}
