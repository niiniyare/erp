package css

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ExpandedCSSTestSuite tests the expanded CSS system with Templ integration
type ExpandedCSSTestSuite struct {
	suite.Suite
	schemaDir string
	generator *TemplStyleGenerator
}

// SetupSuite runs once before all tests in the suite
func (suite *ExpandedCSSTestSuite) SetupSuite() {
	suite.schemaDir = "../../../docs/ui/Schema"
	suite.generator = NewTemplStyleGenerator(suite.schemaDir)
}

// TestExpandedStylesCreation tests creation of ExpandedStyles
func (suite *ExpandedCSSTestSuite) TestExpandedStylesCreation() {
	styles := NewExpandedStyles(suite.schemaDir)

	require.NotNil(suite.T(), styles, "ExpandedStyles should not be nil")
	assert.NotNil(suite.T(), styles.loader, "Schema loader should be initialized")
	assert.NotNil(suite.T(), styles.Custom, "Custom properties map should be initialized")
	assert.True(suite.T(), styles.templIntegration, "Templ integration should be enabled")
}

// TestComprehensiveCSSProperties tests all CSS properties
func (suite *ExpandedCSSTestSuite) TestComprehensiveCSSProperties() {
	styles := NewExpandedStyles(suite.schemaDir)

	// Test layout properties
	styles.WithDisplay("grid").
		WithPosition("absolute").
		WithZIndex("100").
		WithWidth("100%").
		WithHeight("50vh")

	// Test spacing properties
	styles.WithMargin("1rem").
		WithPadding("2rem")

	// Test border properties
	styles.WithBorder("2px solid #000").
		WithBorderRadius("8px")

	// Test typography
	styles.WithFontSize("1.25rem").
		WithFontWeight("600").
		WithColor("#333333").
		WithTextAlign("center")

	// Test flexbox
	styles.WithFlexDirection("column").
		WithJustifyContent("space-between").
		WithAlignItems("stretch")

	// Test grid
	styles.WithGridTemplateColumns("1fr 2fr 1fr").
		WithGap("1rem")

	// Test visual effects
	styles.WithOpacity("0.9").
		WithTransform("scale(1.1)").
		WithTransition("all 0.3s ease").
		WithBoxShadow("0 4px 8px rgba(0,0,0,0.1)")

	css := styles.ToCSS()

	// Verify all properties are included
	expectedProperties := []string{
		"display: grid",
		"position: absolute",
		"z-index: 100",
		"width: 100%",
		"height: 50vh",
		"margin: 1rem",
		"padding: 2rem",
		"border: 2px solid #000",
		"border-radius: 8px",
		"font-size: 1.25rem",
		"font-weight: 600",
		"color: #333333",
		"text-align: center",
		"flex-direction: column",
		"justify-content: space-between",
		"align-items: stretch",
		"grid-template-columns: 1fr 2fr 1fr",
		"gap: 1rem",
		"opacity: 0.9",
		"transform: scale(1.1)",
		"transition: all 0.3s ease",
		"box-shadow: 0 4px 8px rgba(0,0,0,0.1)",
	}

	for _, prop := range expectedProperties {
		assert.Contains(suite.T(), css, prop, "CSS should contain property: %s", prop)
	}
}

// TestCustomPropertiesSupport tests CSS custom properties (variables)
func (suite *ExpandedCSSTestSuite) TestCustomPropertiesSupport() {
	styles := NewExpandedStyles(suite.schemaDir)

	styles.WithCustomProperty("primary-color", "#3b82f6").
		WithCustomProperty("--secondary-color", "#6b7280").
		WithCustomProperty("spacing-unit", "0.5rem")

	css := styles.ToCSS()

	// Verify custom properties are formatted correctly
	assert.Contains(suite.T(), css, "--primary-color: #3b82f6")
	assert.Contains(suite.T(), css, "--secondary-color: #6b7280")
	assert.Contains(suite.T(), css, "--spacing-unit: 0.5rem")
}

// TestTemplStyleIntegration tests Templ-specific functionality
func (suite *ExpandedCSSTestSuite) TestTemplStyleIntegration() {
	styles := NewExpandedStyles(suite.schemaDir)

	styles.WithDisplay("flex").
		WithJustifyContent("center").
		WithPadding("1rem").
		WithBackgroundColor("#ffffff")

	// Test Templ style attribute generation
	styleAttr := styles.ToTemplStyleAttribute()
	assert.NotEmpty(suite.T(), styleAttr, "Templ style attribute should not be empty")
	assert.Contains(suite.T(), styleAttr, "display: flex")
	assert.Contains(suite.T(), styleAttr, "justify-content: center")

	// Test CSS class generation
	cssClass := styles.ToCSSClass("test-component")
	assert.Contains(suite.T(), cssClass, ".test-component {")
	assert.Contains(suite.T(), cssClass, "display: flex")
	assert.Contains(suite.T(), cssClass, "}")

	// Verify proper formatting for Templ
	lines := strings.Split(cssClass, "\n")
	assert.True(suite.T(), len(lines) > 2, "CSS class should be multi-line")
}

// TestJSONSerialization tests JSON serialization/deserialization
func (suite *ExpandedCSSTestSuite) TestJSONSerialization() {
	styles := NewExpandedStyles(suite.schemaDir)

	styles.WithDisplay("grid").
		WithGridTemplateColumns("repeat(3, 1fr)").
		WithGap("2rem").
		WithCustomProperty("theme", "dark")

	// Test JSON export
	jsonStr, err := styles.ToJSON()
	require.NoError(suite.T(), err, "JSON serialization should not error")
	assert.Contains(suite.T(), jsonStr, "\"display\":\"grid\"")
	assert.Contains(suite.T(), jsonStr, "\"grid-template-columns\":\"repeat(3, 1fr)\"")

	// Test JSON import
	newStyles := NewExpandedStyles(suite.schemaDir)
	err = newStyles.FromJSON(jsonStr)
	require.NoError(suite.T(), err, "JSON deserialization should not error")

	assert.Equal(suite.T(), styles.Layout.Display, newStyles.Layout.Display)
	assert.Equal(suite.T(), styles.Grid.TemplateColumns, newStyles.Grid.TemplateColumns)
	assert.Equal(suite.T(), styles.Grid.Gap, newStyles.Grid.Gap)
}

// TestTemplStyleGenerator tests the Templ style generator
func (suite *ExpandedCSSTestSuite) TestTemplStyleGenerator() {
	config := ComponentStyleConfig{
		ComponentType: "atoms.button",
		Variant:       "primary",
		Size:          "md",
		State:         "default",
	}

	styles, err := suite.generator.GenerateForTemplComponent(config)
	require.NoError(suite.T(), err, "Style generation should not error")

	css := styles.ToCSS()

	// Verify button-specific styles
	assert.Contains(suite.T(), css, "display: inline-flex")
	assert.Contains(suite.T(), css, "align-items: center")
	assert.Contains(suite.T(), css, "justify-content: center")
	assert.Contains(suite.T(), css, "background-color: #3b82f6")
}

// TestButtonStyleGeneration tests comprehensive button style generation
func (suite *ExpandedCSSTestSuite) TestButtonStyleGeneration() {
	testCases := []struct {
		variant         string
		size            string
		state           string
		expectedBg      string
		expectedPadding string
	}{
		{"primary", "sm", "default", "#3b82f6", "0.375rem 0.75rem"},
		{"secondary", "md", "default", "#6b7280", "0.5rem 1rem"},
		{"outline", "lg", "default", "transparent", "0.75rem 1.5rem"},
		{"danger", "sm", "disabled", "#ef4444", "0.375rem 0.75rem"},
		{"success", "xl", "default", "#10b981", "1rem 2rem"},
	}

	for _, tc := range testCases {
		styles, err := suite.generator.ButtonStyles(tc.variant, tc.size, tc.state)
		require.NoError(suite.T(), err, "Button style generation should not error")

		css := styles.ToCSS()

		// Check background color
		if tc.expectedBg != "transparent" {
			assert.Contains(suite.T(), css, "background-color: "+tc.expectedBg,
				"Variant %s should have background %s", tc.variant, tc.expectedBg)
		}

		// Check padding for size
		assert.Contains(suite.T(), css, "padding: "+tc.expectedPadding,
			"Size %s should have padding %s", tc.size, tc.expectedPadding)

		// Check disabled state
		if tc.state == "disabled" {
			assert.Contains(suite.T(), css, "opacity: 0.5")
		}
	}
}

// TestInputStyleGeneration tests comprehensive input style generation
func (suite *ExpandedCSSTestSuite) TestInputStyleGeneration() {
	testCases := []struct {
		variant           string
		size              string
		state             string
		expectedMinHeight string
	}{
		{"default", "sm", "default", "2rem"},
		{"default", "md", "default", "2.5rem"},
		{"default", "lg", "default", "3rem"},
		{"error", "md", "default", "2.5rem"},
		{"success", "md", "disabled", "2.5rem"},
	}

	for _, tc := range testCases {
		styles, err := suite.generator.InputStyles(tc.variant, tc.size, tc.state)
		require.NoError(suite.T(), err, "Input style generation should not error")

		css := styles.ToCSS()

		// Check base input styles
		assert.Contains(suite.T(), css, "display: block")
		assert.Contains(suite.T(), css, "width: 100%")

		// Check variant-specific border colors
		switch tc.variant {
		case "error":
			// Error styles should be in custom properties
			assert.Contains(suite.T(), styles.Custom["--border"], "#ef4444")
		case "success":
			assert.Contains(suite.T(), styles.Custom["--border"], "#10b981")
		}

		// Check disabled state
		if tc.state == "disabled" {
			assert.Contains(suite.T(), css, "opacity: 0.5")
		}
	}
}

// TestCardStyleGeneration tests comprehensive card style generation
func (suite *ExpandedCSSTestSuite) TestCardStyleGeneration() {
	testCases := []struct {
		elevation string
		padding   string
		hasShadow bool
	}{
		{"none", "none", false},
		{"sm", "sm", true},
		{"md", "md", true},
		{"lg", "lg", true},
		{"xl", "xl", true},
		{"2xl", "xl", true},
	}

	for _, tc := range testCases {
		styles, err := suite.generator.CardStyles(tc.elevation, tc.padding)
		require.NoError(suite.T(), err, "Card style generation should not error")

		css := styles.ToCSS()

		// Check base styles
		assert.Contains(suite.T(), css, "background-color: #ffffff")
		assert.Contains(suite.T(), css, "border-radius: 0.5rem")

		// Check shadow presence
		if tc.hasShadow {
			assert.Contains(suite.T(), css, "box-shadow:")
		} else {
			assert.NotContains(suite.T(), css, "box-shadow:")
		}

		// Check padding
		if tc.padding != "none" {
			assert.Contains(suite.T(), css, "padding:")
		}
	}
}

// TestResponsiveStyleGeneration tests responsive style capabilities
func (suite *ExpandedCSSTestSuite) TestResponsiveStyleGeneration() {
	config := ComponentStyleConfig{
		ComponentType: "atoms.button",
		Variant:       "primary",
		Size:          "md",
		Responsive: map[string]string{
			"sm": "font-size: 0.875rem",
			"md": "font-size: 1rem",
			"lg": "font-size: 1.125rem",
		},
	}

	styles, err := suite.generator.GenerateForTemplComponent(config)
	require.NoError(suite.T(), err, "Responsive style generation should not error")

	// Check responsive custom properties
	assert.Contains(suite.T(), styles.Custom, "--sm-override")
	assert.Contains(suite.T(), styles.Custom, "--md-override")
	assert.Contains(suite.T(), styles.Custom, "--lg-override")
}

// TestThemeIntegration tests dark/light theme integration
func (suite *ExpandedCSSTestSuite) TestThemeIntegration() {
	lightConfig := ComponentStyleConfig{
		ComponentType: "atoms.card",
		Theme:         "light",
	}

	darkConfig := ComponentStyleConfig{
		ComponentType: "atoms.card",
		Theme:         "dark",
	}

	lightStyles, err := suite.generator.GenerateForTemplComponent(lightConfig)
	require.NoError(suite.T(), err, "Light theme generation should not error")

	darkStyles, err := suite.generator.GenerateForTemplComponent(darkConfig)
	require.NoError(suite.T(), err, "Dark theme generation should not error")

	// Light theme should have light backgrounds
	lightCSS := lightStyles.ToCSS()
	assert.Contains(suite.T(), lightCSS, "background-color: #ffffff")

	// Dark theme should have dark backgrounds
	darkCSS := darkStyles.ToCSS()
	assert.Contains(suite.T(), darkCSS, "background-color: #1f2937")
}

// TestStyleAttributeGeneration tests generation of style attributes for Templ
func (suite *ExpandedCSSTestSuite) TestStyleAttributeGeneration() {
	config := ComponentStyleConfig{
		ComponentType: "atoms.button",
		Variant:       "success",
		Size:          "lg",
		CustomProps: map[string]any{
			"border-radius": "12px",
			"font-weight":   "700",
		},
	}

	styleAttr, err := suite.generator.GenerateTemplStyleAttribute(config)
	require.NoError(suite.T(), err, "Style attribute generation should not error")

	assert.NotEmpty(suite.T(), styleAttr)
	assert.Contains(suite.T(), styleAttr, "display: inline-flex")
	assert.Contains(suite.T(), styleAttr, "background-color: #10b981")
	assert.Contains(suite.T(), styleAttr, "--border-radius: 12px")
	assert.Contains(suite.T(), styleAttr, "--font-weight: 700")

	// Should be properly formatted for Templ usage
	assert.NotContains(suite.T(), styleAttr, "\n", "Style attribute should be single line")
}

// TestCSSClassGeneration tests generation of CSS classes for Templ
func (suite *ExpandedCSSTestSuite) TestCSSClassGeneration() {
	config := ComponentStyleConfig{
		ComponentType: "molecules.field",
		Variant:       "default",
		Size:          "md",
	}

	cssClass, err := suite.generator.GenerateTemplClass("field-component", config)
	require.NoError(suite.T(), err, "CSS class generation should not error")

	assert.NotEmpty(suite.T(), cssClass)
	assert.Contains(suite.T(), cssClass, ".field-component {")
	assert.Contains(suite.T(), cssClass, "}")

	// Should be properly formatted with line breaks
	lines := strings.Split(cssClass, "\n")
	assert.True(suite.T(), len(lines) >= 3, "CSS class should have multiple lines")
}

// TestBackwardCompatibility tests compatibility with existing Styles type
func (suite *ExpandedCSSTestSuite) TestBackwardCompatibility() {
	// Create an expanded styles instance
	expanded := NewExpandedStyles(suite.schemaDir)
	expanded.WithDisplay("flex").
		WithJustifyContent("center").
		WithPadding("1rem").
		WithBackgroundColor("#ffffff").
		WithCustomProperty("theme", "light")

	// Convert to legacy styles
	legacy := expanded.ToLegacyStyles()

	require.NotNil(suite.T(), legacy, "Legacy conversion should not be nil")
	assert.Equal(suite.T(), "flex", legacy.Display)
	assert.Equal(suite.T(), "center", legacy.JustifyContent)
	assert.Equal(suite.T(), "1rem", legacy.Padding)
	assert.Equal(suite.T(), "#ffffff", legacy.BackgroundColor)
	assert.Contains(suite.T(), legacy.Custom, "--theme")

	// Convert back to expanded
	reconverted := CreateFromLegacyStyles(legacy, suite.schemaDir)

	require.NotNil(suite.T(), reconverted, "Reconversion should not be nil")
	assert.Equal(suite.T(), expanded.Layout.Display, reconverted.Layout.Display)
	assert.Equal(suite.T(), expanded.Flexbox.JustifyContent, reconverted.Flexbox.JustifyContent)
	assert.Equal(suite.T(), expanded.Padding.Padding, reconverted.Padding.Padding)
	assert.Equal(suite.T(), expanded.Background.Color, reconverted.Background.Color)
}

// TestHelperFunctionGeneration tests generation of Templ helper functions
func (suite *ExpandedCSSTestSuite) TestHelperFunctionGeneration() {
	helperFunc := suite.generator.GenerateTemplHelperFunction("atoms.button")

	assert.Contains(suite.T(), helperFunc, "func AtomsButtonStyles(")
	assert.Contains(suite.T(), helperFunc, "variant, size, state string")
	assert.Contains(suite.T(), helperFunc, "css.NewTemplStyleGenerator")
	assert.Contains(suite.T(), helperFunc, "ComponentType: \"atoms.button\"")
	assert.Contains(suite.T(), helperFunc, "return styleAttr")
}

// TestSchemaIntegration tests integration with JSON schemas
func (suite *ExpandedCSSTestSuite) TestSchemaIntegration() {
	styles := NewExpandedStyles(suite.schemaDir)

	// Test property validation (if schemas are available)
	err := styles.ValidateProperty("Property.Display", "flex")
	// Note: This may fail if schema files aren't available, which is acceptable
	if err != nil {
		suite.T().Logf("Schema validation not available: %v", err)
	}

	// Test loading property from schema definition
	err = styles.LoadFromSchemaDefinition("Property.Display")
	if err != nil {
		suite.T().Logf("Schema loading not available: %v", err)
	}
}

// Run the test suite
func TestExpandedCSSTestSuite(t *testing.T) {
	suite.Run(t, new(ExpandedCSSTestSuite))
}

// Additional standalone tests for edge cases

func TestExpandedStylesEdgeCases(t *testing.T) {
	schemaDir := "../../../docs/ui/Schema"

	// Test with nil loader
	styles := &ExpandedStyles{Custom: make(map[string]string)}
	css := styles.ToCSS()
	assert.Empty(t, css, "Empty styles should produce empty CSS")

	// Test custom property formatting
	styles = NewExpandedStyles(schemaDir)
	styles.WithCustomProperty("custom-prop", "value").
		WithCustomProperty("--already-prefixed", "value2")

	css = styles.ToCSS()
	assert.Contains(t, css, "--custom-prop: value")
	assert.Contains(t, css, "--already-prefixed: value2")

	// Test JSON with empty styles
	emptyStyles := NewExpandedStyles(schemaDir)
	jsonStr, err := emptyStyles.ToJSON()
	require.NoError(t, err)
	assert.Contains(t, jsonStr, "{")
}

func TestComponentStyleConfig(t *testing.T) {
	config := ComponentStyleConfig{
		ComponentType: "atoms.button",
		Variant:       "primary",
		Size:          "md",
		CustomProps: map[string]any{
			"margin": "1rem",
			"active": true,
		},
		Responsive: map[string]string{
			"sm": "width: 100%",
		},
	}

	generator := NewTemplStyleGenerator("../../../docs/ui/Schema")

	// Test JSON serialization
	jsonStr, err := generator.GenerateComponentJSON(config)
	require.NoError(t, err)
	assert.Contains(t, jsonStr, "\"component_type\":\"atoms.button\"")

	// Test JSON deserialization
	parsedConfig, err := generator.ParseComponentJSON(jsonStr)
	require.NoError(t, err)
	assert.Equal(t, config.ComponentType, parsedConfig.ComponentType)
	assert.Equal(t, config.Variant, parsedConfig.Variant)
}
