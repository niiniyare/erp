package css

import (
	"fmt"
	"strconv"
	"strings"
)

// DataTypeValidator provides validation for CSS data types based on JSON schemas
type DataTypeValidator struct {
	schemaLoader *SchemaLoader
	dataTypes    map[string]*DataTypeDefinition
}

// DataTypeDefinition represents a CSS data type schema
type DataTypeDefinition struct {
	ID     string               `json:"id"`
	Schema string               `json:"schema"`
	AnyOf  []DataTypeConstraint `json:"anyOf,omitempty"`
	Type   string               `json:"type,omitempty"`
	Const  any                  `json:"const,omitempty"`
	Enum   []any                `json:"enum,omitempty"`
	Ref    string               `json:"ref,omitempty"`
}

// DataTypeConstraint represents a constraint in a data type definition
type DataTypeConstraint struct {
	Type  string `json:"type,omitempty"`
	Const any    `json:"const,omitempty"`
	Ref   string `json:"ref,omitempty"`
}

// CSS Data Type Enums and Types

// Color data types
type CSSColor string

const (
	ColorCurrentColor CSSColor = "currentcolor"
	ColorTransparent  CSSColor = "transparent"
	ColorInitial      CSSColor = "initial"
	ColorInherit      CSSColor = "inherit"
	ColorUnset        CSSColor = "unset"
)

// Position data types
type BgPosition string

const (
	BgPositionTop    BgPosition = "top"
	BgPositionRight  BgPosition = "right"
	BgPositionBottom BgPosition = "bottom"
	BgPositionLeft   BgPosition = "left"
	BgPositionCenter BgPosition = "center"
)

// Size data types
type BgSize string

const (
	BgSizeAuto    BgSize = "auto"
	BgSizeCover   BgSize = "cover"
	BgSizeContain BgSize = "contain"
)

// Line width data types
type LineWidth string

const (
	LineWidthThin   LineWidth = "thin"
	LineWidthMedium LineWidth = "medium"
	LineWidthThick  LineWidth = "thick"
)

// Track breadth data types (for CSS Grid)
type TrackBreadth string

const (
	TrackBreadthAuto       TrackBreadth = "auto"
	TrackBreadthMaxContent TrackBreadth = "max-content"
	TrackBreadthMinContent TrackBreadth = "min-content"
)

// Animation data types
type SingleAnimation string

const (
	AnimationNone     SingleAnimation = "none"
	AnimationInfinite SingleAnimation = "infinite"
	AnimationRunning  SingleAnimation = "running"
	AnimationPaused   SingleAnimation = "paused"
)

// Transition data types
type SingleTransition string

const (
	TransitionNone SingleTransition = "none"
	TransitionAll  SingleTransition = "all"
)

// Blending and compositing
type BlendMode string

const (
	BlendModeNormal     BlendMode = "normal"
	BlendModeMultiply   BlendMode = "multiply"
	BlendModeScreen     BlendMode = "screen"
	BlendModeOverlay    BlendMode = "overlay"
	BlendModeDarken     BlendMode = "darken"
	BlendModeLighten    BlendMode = "lighten"
	BlendModeColorDodge BlendMode = "color-dodge"
	BlendModeColorBurn  BlendMode = "color-burn"
	BlendModeHardLight  BlendMode = "hard-light"
	BlendModeSoftLight  BlendMode = "soft-light"
	BlendModeDifference BlendMode = "difference"
	BlendModeExclusion  BlendMode = "exclusion"
	BlendModeHue        BlendMode = "hue"
	BlendModeSaturation BlendMode = "saturation"
	BlendModeColor      BlendMode = "color"
	BlendModeLuminosity BlendMode = "luminosity"
)

// Display data types
type DisplayInside string

const (
	DisplayInsideFlow     DisplayInside = "flow"
	DisplayInsideFlowRoot DisplayInside = "flow-root"
	DisplayInsideTable    DisplayInside = "table"
	DisplayInsideFlex     DisplayInside = "flex"
	DisplayInsideGrid     DisplayInside = "grid"
	DisplayInsideRuby     DisplayInside = "ruby"
)

type DisplayOutside string

const (
	DisplayOutsideBlock  DisplayOutside = "block"
	DisplayOutsideInline DisplayOutside = "inline"
	DisplayOutsideRunIn  DisplayOutside = "run-in"
)

// Font data types
type FontWeightAbsolute string

const (
	FontWeightAbsolute100 FontWeightAbsolute = "100"
	FontWeightAbsolute200 FontWeightAbsolute = "200"
	FontWeightAbsolute300 FontWeightAbsolute = "300"
	FontWeightAbsolute400 FontWeightAbsolute = "400"
	FontWeightAbsolute500 FontWeightAbsolute = "500"
	FontWeightAbsolute600 FontWeightAbsolute = "600"
	FontWeightAbsolute700 FontWeightAbsolute = "700"
	FontWeightAbsolute800 FontWeightAbsolute = "800"
	FontWeightAbsolute900 FontWeightAbsolute = "900"
)

type GenericFamily string

const (
	GenericFamilySerif     GenericFamily = "serif"
	GenericFamilySansSerif GenericFamily = "sans-serif"
	GenericFamilyMonospace GenericFamily = "monospace"
	GenericFamilyCursive   GenericFamily = "cursive"
	GenericFamilyFantasy   GenericFamily = "fantasy"
)

// Timing functions
type EasingFunction string

const (
	EasingFunctionLinear    EasingFunction = "linear"
	EasingFunctionEase      EasingFunction = "ease"
	EasingFunctionEaseIn    EasingFunction = "ease-in"
	EasingFunctionEaseOut   EasingFunction = "ease-out"
	EasingFunctionEaseInOut EasingFunction = "ease-in-out"
	EasingFunctionStepStart EasingFunction = "step-start"
	EasingFunctionStepEnd   EasingFunction = "step-end"
)

// NewDataTypeValidator creates a new data type validator
func NewDataTypeValidator(schemaDir string) *DataTypeValidator {
	return &DataTypeValidator{
		schemaLoader: NewSchemaLoader(schemaDir),
		dataTypes:    make(map[string]*DataTypeDefinition),
	}
}

// LoadDataTypes loads all CSS data type schemas
func (dtv *DataTypeValidator) LoadDataTypes() error {
	dataTypeFiles := []string{
		"DataType.Color", "DataType.BgPosition", "DataType.BgSize",
		"DataType.LineWidth", "DataType.TrackBreadth", "DataType.SingleAnimation",
		"DataType.SingleTransition", "DataType.BlendMode", "DataType.DisplayInside",
		"DataType.DisplayOutside", "DataType.FontWeightAbsolute", "DataType.GenericFamily",
		"DataType.EasingFunction", "DataType.Position", "DataType.Dasharray",
		"DataType.FinalBgLayer", "DataType.MaskLayer", "DataType.Paint",
		"DataType.AbsoluteSize", "DataType.AnimateableFeature", "DataType.Attachment",
		"DataType.Box", "DataType.CompatAuto", "DataType.CompositeStyle",
		"DataType.CompositingOperator", "DataType.ContentDistribution",
		"DataType.ContentList", "DataType.ContentPosition", "DataType.CubicBezierTimingFunction",
		"DataType.DeprecatedSystemColor", "DataType.DisplayInternal", "DataType.DisplayLegacy",
		"DataType.EastAsianVariantValues", "DataType.FontStretchAbsolute", "DataType.GeometryBox",
		"DataType.GridLine", "DataType.LineStyle", "DataType.MaskingMode",
		"DataType.NamedColor", "DataType.Quote", "DataType.RepeatStyle",
		"DataType.SelfPosition", "DataType.SingleAnimationComposition",
		"DataType.SingleAnimationDirection", "DataType.SingleAnimationFillMode",
		"DataType.SingleAnimationTimeline", "DataType.StepTimingFunction",
		"DataType.TimelineRangeName", "DataType.VisualBox",
	}

	for _, dataType := range dataTypeFiles {
		if err := dtv.loadDataType(dataType); err != nil {
			// Log error but continue loading other types
			continue
		}
	}

	return nil
}

// loadDataType loads a single data type schema
func (dtv *DataTypeValidator) loadDataType(dataTypeName string) error {
	// Use the existing schema loader method
	schema, err := dtv.schemaLoader.LoadPropertySchema(dataTypeName)
	if err != nil {
		return fmt.Errorf("failed to load data type schema %s: %w", dataTypeName, err)
	}

	// Convert PropertySchema to DataTypeDefinition
	definition := &DataTypeDefinition{
		ID:     schema.ID,
		Schema: schema.Schema,
		Type:   schema.Type,
		Const:  schema.Const,
		Enum:   schema.Enum,
	}

	// Convert AnyOf variants
	for _, variant := range schema.AnyOf {
		definition.AnyOf = append(definition.AnyOf, DataTypeConstraint{
			Type:  variant.Type,
			Const: variant.Const,
			Ref:   variant.Ref,
		})
	}

	dtv.dataTypes[dataTypeName] = definition
	return nil
}

// ValidateValue validates a CSS value against a data type
func (dtv *DataTypeValidator) ValidateValue(dataTypeName string, value string) error {
	definition, exists := dtv.dataTypes[dataTypeName]
	if !exists {
		// Try to load the data type if not loaded
		if err := dtv.loadDataType(dataTypeName); err != nil {
			return fmt.Errorf("data type %s not found: %w", dataTypeName, err)
		}
		definition = dtv.dataTypes[dataTypeName]
	}

	return dtv.validateAgainstDefinition(value, definition)
}

// validateAgainstDefinition validates a value against a data type definition
func (dtv *DataTypeValidator) validateAgainstDefinition(value string, definition *DataTypeDefinition) error {
	// Handle direct type constraint
	if definition.Type != "" {
		return dtv.validateType(value, definition.Type, definition.Const, definition.Enum)
	}

	// Handle anyOf constraints
	if len(definition.AnyOf) > 0 {
		for _, constraint := range definition.AnyOf {
			if err := dtv.validateConstraint(value, constraint); err == nil {
				return nil // At least one constraint matches
			}
		}
		return fmt.Errorf("value %s does not match any allowed constraints", value)
	}

	return fmt.Errorf("no validation rules found for value %s", value)
}

// validateConstraint validates a value against a single constraint
func (dtv *DataTypeValidator) validateConstraint(value string, constraint DataTypeConstraint) error {
	// Handle reference to another data type
	if constraint.Ref != "" {
		refName := strings.TrimPrefix(constraint.Ref, "#/definitions/")
		return dtv.ValidateValue(refName, value)
	}

	return dtv.validateType(value, constraint.Type, constraint.Const, nil)
}

// validateType validates a value against a specific type constraint
func (dtv *DataTypeValidator) validateType(value string, expectedType string, constValue any, enumValues []any) error {
	// Check constant value
	if constValue != nil {
		if value != fmt.Sprintf("%v", constValue) {
			return fmt.Errorf("expected constant value %v, got %s", constValue, value)
		}
		return nil
	}

	// Check enum values
	if len(enumValues) > 0 {
		for _, enumValue := range enumValues {
			if value == fmt.Sprintf("%v", enumValue) {
				return nil
			}
		}
		return fmt.Errorf("value %s is not in allowed enum values", value)
	}

	// Validate based on type
	switch expectedType {
	case "string":
		// Any string is valid
		return nil
	case "number":
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			return fmt.Errorf("value %s is not a valid number", value)
		}
		return nil
	case "integer":
		if _, err := strconv.Atoi(value); err != nil {
			return fmt.Errorf("value %s is not a valid integer", value)
		}
		return nil
	default:
		return fmt.Errorf("unknown type constraint: %s", expectedType)
	}
}

// GetAllowedValues returns all allowed values for a data type (for documentation/autocomplete)
func (dtv *DataTypeValidator) GetAllowedValues(dataTypeName string) ([]string, error) {
	definition, exists := dtv.dataTypes[dataTypeName]
	if !exists {
		if err := dtv.loadDataType(dataTypeName); err != nil {
			return nil, fmt.Errorf("data type %s not found: %w", dataTypeName, err)
		}
		definition = dtv.dataTypes[dataTypeName]
	}

	var values []string

	// Collect constant values from anyOf constraints
	for _, constraint := range definition.AnyOf {
		if constraint.Const != nil {
			values = append(values, fmt.Sprintf("%v", constraint.Const))
		}
	}

	// Add direct constant if present
	if definition.Const != nil {
		values = append(values, fmt.Sprintf("%v", definition.Const))
	}

	// Add enum values if present
	for _, enumValue := range definition.Enum {
		values = append(values, fmt.Sprintf("%v", enumValue))
	}

	return values, nil
}

// Enhanced CSS Value Types that integrate with DataTypes

// CSSValue represents a CSS value that can be validated against data types
type CSSValue struct {
	Raw       string
	DataType  string
	Validated bool
	validator *DataTypeValidator
}

// NewCSSValue creates a new CSS value with data type validation
func NewCSSValue(value string, dataType string, validator *DataTypeValidator) *CSSValue {
	return &CSSValue{
		Raw:       value,
		DataType:  dataType,
		validator: validator,
	}
}

// Validate validates the CSS value against its data type
func (cv *CSSValue) Validate() error {
	if cv.validator == nil {
		return fmt.Errorf("no validator available")
	}

	if err := cv.validator.ValidateValue(cv.DataType, cv.Raw); err != nil {
		return err
	}

	cv.Validated = true
	return nil
}

// String returns the string representation of the CSS value
func (cv *CSSValue) String() string {
	return cv.Raw
}

// Enhanced ExpandedStyles integration with DataTypes

// WithValidatedProperty sets a CSS property with data type validation
func (s *ExpandedStyles) WithValidatedProperty(property, value, dataType string) (*ExpandedStyles, error) {
	if s.loader == nil {
		return s, fmt.Errorf("no schema loader available for validation")
	}

	// Create data type validator
	validator := NewDataTypeValidator(s.loader.schemaDir)

	// Validate the value
	if err := validator.ValidateValue(dataType, value); err != nil {
		return s, fmt.Errorf("validation failed for property %s with value %s: %w", property, value, err)
	}

	// Set the property using existing method
	s.setPropertyByName(property, value)
	return s, nil
}

// WithValidatedColor sets a color property with data type validation
func (s *ExpandedStyles) WithValidatedColor(value string) (*ExpandedStyles, error) {
	return s.WithValidatedProperty("color", value, "DataType.Color")
}

// WithValidatedBackgroundColor sets a background-color property with data type validation
func (s *ExpandedStyles) WithValidatedBackgroundColor(value string) (*ExpandedStyles, error) {
	return s.WithValidatedProperty("background-color", value, "DataType.Color")
}

// WithValidatedBorderWidth sets a border-width property with data type validation
func (s *ExpandedStyles) WithValidatedBorderWidth(value string) (*ExpandedStyles, error) {
	return s.WithValidatedProperty("border-width", value, "DataType.LineWidth")
}

// WithValidatedBackgroundPosition sets a background-position property with data type validation
func (s *ExpandedStyles) WithValidatedBackgroundPosition(value string) (*ExpandedStyles, error) {
	return s.WithValidatedProperty("background-position", value, "DataType.BgPosition")
}

// WithValidatedBackgroundSize sets a background-size property with data type validation
func (s *ExpandedStyles) WithValidatedBackgroundSize(value string) (*ExpandedStyles, error) {
	return s.WithValidatedProperty("background-size", value, "DataType.BgSize")
}

// WithValidatedFontWeight sets a font-weight property with data type validation
func (s *ExpandedStyles) WithValidatedFontWeight(value string) (*ExpandedStyles, error) {
	return s.WithValidatedProperty("font-weight", value, "DataType.FontWeightAbsolute")
}

// WithValidatedAnimation sets an animation property with data type validation
func (s *ExpandedStyles) WithValidatedAnimation(value string) (*ExpandedStyles, error) {
	return s.WithValidatedProperty("animation", value, "DataType.SingleAnimation")
}

// WithValidatedTransition sets a transition property with data type validation
func (s *ExpandedStyles) WithValidatedTransition(value string) (*ExpandedStyles, error) {
	return s.WithValidatedProperty("transition", value, "DataType.SingleTransition")
}

// DataTypeInfo provides information about a CSS data type for documentation
type DataTypeInfo struct {
	Name          string
	Description   string
	AllowedValues []string
	ExampleValues []string
	Type          string
}

// GetDataTypeInfo returns information about a CSS data type
func (dtv *DataTypeValidator) GetDataTypeInfo(dataTypeName string) (*DataTypeInfo, error) {
	allowedValues, err := dtv.GetAllowedValues(dataTypeName)
	if err != nil {
		return nil, err
	}

	info := &DataTypeInfo{
		Name:          dataTypeName,
		AllowedValues: allowedValues,
	}

	// Add descriptions and examples based on data type
	switch dataTypeName {
	case "DataType.Color":
		info.Description = "CSS color values including named colors, hex, rgb, hsl, etc."
		info.ExampleValues = []string{"#ff0000", "red", "rgb(255,0,0)", "currentcolor"}
		info.Type = "string"
	case "DataType.BgPosition":
		info.Description = "Background position values"
		info.ExampleValues = []string{"top", "center", "50%", "10px 20px"}
		info.Type = "string|number"
	case "DataType.LineWidth":
		info.Description = "Line width values for borders and outlines"
		info.ExampleValues = []string{"thin", "medium", "thick", "1px", "0.5em"}
		info.Type = "string|number"
	case "DataType.FontWeightAbsolute":
		info.Description = "Absolute font weight values"
		info.ExampleValues = []string{"100", "200", "400", "700", "900"}
		info.Type = "string"
	default:
		info.Description = fmt.Sprintf("CSS data type: %s", dataTypeName)
		info.Type = "mixed"
	}

	return info, nil
}

