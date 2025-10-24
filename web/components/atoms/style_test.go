package atoms

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// StyleTestSuite tests the style utility functions in style.go
type StyleTestSuite struct {
	suite.Suite
}

// SetupTest runs before each test method
func (suite *StyleTestSuite) SetupTest() {
	ResetIDCounter()
}

// TestStyleTestSuite runs the style test suite
func TestStyleTestSuite(t *testing.T) {
	suite.Run(t, new(StyleTestSuite))
}

// ============================================================================
// ROUNDED CLASS TESTS
// ============================================================================

func (suite *StyleTestSuite) TestGetRoundedClass() {
	tests := []struct {
		name     string
		rounded  bool
		size     Size
		expected string
	}{
		{"not rounded", false, SizeMD, RoundedSM},
		{"rounded XS", true, SizeXS, RoundedSM},
		{"rounded SM", true, SizeSM, RoundedSM},
		{"rounded MD", true, SizeMD, RoundedMD},
		{"rounded LG", true, SizeLG, RoundedLG},
		{"rounded XL", true, SizeXL, RoundedLG},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := GetRoundedClass(tt.rounded, tt.size)
			suite.Equal(tt.expected, result)
		})
	}
}

// ============================================================================
// THEME TESTS
// ============================================================================

func (suite *StyleTestSuite) TestDefaultTheme() {
	theme := DefaultTheme()
	
	suite.Equal("default", theme.Name)
	suite.NotEmpty(theme.Colors)
	suite.NotEmpty(theme.Spacing)
	suite.NotEmpty(theme.Typography)
	
	// Test essential colors exist
	suite.Contains(theme.Colors, "primary")
	suite.Contains(theme.Colors, "secondary")
	suite.Contains(theme.Colors, "success")
	suite.Contains(theme.Colors, "error")
	suite.Contains(theme.Colors, "warning")
	suite.Contains(theme.Colors, "info")
}

func (suite *StyleTestSuite) TestThemeGetColor() {
	theme := DefaultTheme()
	
	// Test existing color
	primaryColor := theme.GetColor("primary")
	suite.Equal("#3b82f6", primaryColor)
	
	// Test fallback to neutral
	nonExistentColor := theme.GetColor("non-existent")
	suite.Equal(theme.Colors["neutral"], nonExistentColor)
	
	// Test ultimate fallback
	emptyTheme := &Theme{Colors: map[string]string{}}
	fallbackColor := emptyTheme.GetColor("any")
	suite.Equal(DefaultFallbackColor, fallbackColor)
}

func (suite *StyleTestSuite) TestThemeGetSpacing() {
	theme := DefaultTheme()
	
	// Test existing spacing
	mdSpacing := theme.GetSpacing("md")
	suite.Equal("1rem", mdSpacing)
	
	// Test fallback to md
	nonExistentSpacing := theme.GetSpacing("non-existent")
	suite.Equal("1rem", nonExistentSpacing)
	
	// Test ultimate fallback
	emptyTheme := &Theme{Spacing: map[string]string{}}
	fallbackSpacing := emptyTheme.GetSpacing("any")
	suite.Equal("1rem", fallbackSpacing)
}

func (suite *StyleTestSuite) TestThemeGetTypography() {
	theme := DefaultTheme()
	
	// Test existing typography
	mdTypography := theme.GetTypography("md")
	suite.Equal("1rem", mdTypography.FontSize)
	suite.Equal("1.5rem", mdTypography.LineHeight)
	
	// Test fallback to md
	nonExistentTypography := theme.GetTypography("non-existent")
	suite.Equal("1rem", nonExistentTypography.FontSize)
	
	// Test ultimate fallback
	emptyTheme := &Theme{Typography: map[string]TypographyConfig{}}
	fallbackTypography := emptyTheme.GetTypography("any")
	suite.Equal("1rem", fallbackTypography.FontSize)
	suite.Equal("1.5rem", fallbackTypography.LineHeight)
	suite.Equal("400", fallbackTypography.FontWeight)
}

func (suite *StyleTestSuite) TestThemeGetShadow() {
	theme := DefaultTheme()
	
	// Test existing shadow
	mdShadow := theme.GetShadow("md")
	suite.NotEmpty(mdShadow)
	
	// Test fallback
	nonExistentShadow := theme.GetShadow("non-existent")
	suite.Equal("none", nonExistentShadow)
}

func (suite *StyleTestSuite) TestThemeGetBorderRadius() {
	theme := DefaultTheme()
	
	// Test existing border radius
	mdRadius := theme.GetBorderRadius("md")
	suite.Equal("0.375rem", mdRadius)
	
	// Test fallback to md
	nonExistentRadius := theme.GetBorderRadius("non-existent")
	suite.Equal("0.375rem", nonExistentRadius)
	
	// Test ultimate fallback
	emptyTheme := &Theme{BorderRadius: map[string]string{}}
	fallbackRadius := emptyTheme.GetBorderRadius("any")
	suite.Equal("0.375rem", fallbackRadius)
}

func (suite *StyleTestSuite) TestThemeGetAnimation() {
	theme := DefaultTheme()
	
	// Test existing animation
	normalAnimation := theme.GetAnimation("normal")
	suite.Equal("300ms", normalAnimation.Duration)
	
	// Test fallback to normal
	nonExistentAnimation := theme.GetAnimation("non-existent")
	suite.Equal("300ms", nonExistentAnimation.Duration)
	
	// Test ultimate fallback
	emptyTheme := &Theme{Animations: map[string]AnimationConfig{}}
	fallbackAnimation := emptyTheme.GetAnimation("any")
	suite.Equal("300ms", fallbackAnimation.Duration)
}

func (suite *StyleTestSuite) TestThemeGetZIndex() {
	theme := DefaultTheme()
	
	// Test existing z-index
	modalZIndex := theme.GetZIndex("modal")
	suite.Equal(1050, modalZIndex)
	
	// Test fallback
	nonExistentZIndex := theme.GetZIndex("non-existent")
	suite.Equal(1000, nonExistentZIndex)
}

// ============================================================================
// STATE COLOR TESTS
// ============================================================================

func (suite *StyleTestSuite) TestGetStateColors() {
	tests := []struct {
		state ValidationState
	}{
		{StateDefault},
		{StateSuccess},
		{StateError},
		{StateWarning},
		{StateInfo},
	}

	for _, tt := range tests {
		suite.Run(string(tt.state), func() {
			colors := GetStateColors(tt.state)
			suite.NotEmpty(colors.Border)
			suite.NotEmpty(colors.Background)
			suite.NotEmpty(colors.Text)
			suite.NotEmpty(colors.Ring)
			suite.NotEmpty(colors.Icon)
		})
	}

	// Test invalid state falls back to default
	invalidState := ValidationState("invalid")
	colors := GetStateColors(invalidState)
	defaultColors := GetStateColors(StateDefault)
	suite.Equal(defaultColors, colors)
}

func (suite *StyleTestSuite) TestGetFeedbackColors() {
	tests := []struct {
		state    ValidationState
		expected []string
	}{
		{StateDefault, []string{"text-gray-600", "dark:text-gray-400"}},
		{StateSuccess, []string{"text-green-600", "dark:text-green-500"}},
		{StateError, []string{"text-red-600", "dark:text-red-500"}},
		{StateWarning, []string{"text-yellow-600", "dark:text-yellow-500"}},
		{StateInfo, []string{"text-blue-600", "dark:text-blue-500"}},
	}

	for _, tt := range tests {
		suite.Run(string(tt.state), func() {
			colors := GetFeedbackColors(tt.state)
			suite.Equal(tt.expected, colors)
		})
	}
}

func (suite *StyleTestSuite) TestGetLabelColorClasses() {
	tests := []struct {
		state ValidationState
	}{
		{StateDefault},
		{StateError},
		{StateSuccess},
		{StateWarning},
		{StateInfo},
	}

	for _, tt := range tests {
		suite.Run(string(tt.state), func() {
			colors := GetLabelColorClasses(tt.state)
			suite.NotEmpty(colors)
			suite.True(len(colors) >= 1)
		})
	}
}

func (suite *StyleTestSuite) TestGetFeedbackClasses() {
	result := GetFeedbackClasses(StateError)
	suite.Contains(result, "mt-2")
	suite.Contains(result, "text-sm")
	suite.Contains(result, "text-red-600")
}

func (suite *StyleTestSuite) TestGetValidationStateClasses() { //nolint:revive // Test function complexity is acceptable
	tests := []struct {
		state ValidationState
	}{
		{StateDefault},
		{StateError},
		{StateSuccess},
		{StateWarning},
		{StateInfo},
	}

	for _, tt := range tests {
		suite.Run(string(tt.state), func() {
			classes := GetValidationStateClasses(tt.state)
			suite.NotEmpty(classes)
			
			// All should have border and focus classes
			hasbordClass := false
			hasFocusClass := false
			for _, class := range classes {
				if Contains([]string{"border-red-500", "border-green-500", "border-yellow-500", "border-blue-500", "border-gray-300"}, class) {
					hasbordClass = true
				}
				if Contains([]string{"focus:ring-red-500", "focus:ring-green-500", "focus:ring-yellow-500", "focus:ring-blue-500"}, class) {
					hasFocusClass = true
				}
			}
			suite.True(hasbordClass || hasFocusClass, "Should have border or focus classes")
		})
	}
}

// ============================================================================
// SIZE CLASS TESTS
// ============================================================================

func (suite *StyleTestSuite) TestGetInputSizeClasses() {
	tests := []struct {
		size     Size
		expected []string
	}{
		{SizeXS, []string{"text-xs", "py-1", "px-2", "h-7"}},
		{SizeSM, []string{"text-sm", "py-1.5", "px-3", "h-8"}},
		{SizeMD, []string{"text-sm", "py-2.5", "px-4", "h-10"}},
		{SizeLG, []string{"text-base", "py-3", "px-4", "h-12"}},
		{SizeXL, []string{"text-lg", "py-4", "px-5", "h-14"}},
	}

	for _, tt := range tests {
		suite.Run(string(tt.size), func() {
			result := GetInputSizeClasses(tt.size)
			suite.Equal(tt.expected, result)
		})
	}
}

func (suite *StyleTestSuite) TestGetButtonSizeClasses() {
	tests := []struct {
		size     Size
		expected []string
	}{
		{SizeXS, []string{"text-xs", "py-1", "px-2", "h-7"}},
		{SizeSM, []string{"text-sm", "py-2", "px-3", "h-8"}},
		{SizeMD, []string{"text-sm", "py-2.5", "px-5", "h-10"}},
		{SizeLG, []string{"text-base", "py-3", "px-5", "h-12"}},
		{SizeXL, []string{"text-lg", "py-3.5", "px-6", "h-14"}},
	}

	for _, tt := range tests {
		suite.Run(string(tt.size), func() {
			result := GetButtonSizeClasses(tt.size)
			suite.Equal(tt.expected, result)
		})
	}
}

func (suite *StyleTestSuite) TestGetCheckboxSizeClasses() {
	tests := []struct {
		size     Size
		expected []string
	}{
		{SizeXS, []string{"w-3", "h-3"}},
		{SizeSM, []string{"w-3.5", "h-3.5"}},
		{SizeMD, []string{"w-4", "h-4"}},
		{SizeLG, []string{"w-5", "h-5"}},
		{SizeXL, []string{"w-6", "h-6"}},
	}

	for _, tt := range tests {
		suite.Run(string(tt.size), func() {
			result := GetCheckboxSizeClasses(tt.size)
			suite.Equal(tt.expected, result)
		})
	}
}

func (suite *StyleTestSuite) TestGetIconSizeClasses() {
	tests := []struct {
		size     Size
		expected []string
	}{
		{SizeXS, []string{"w-3", "h-3"}},
		{SizeSM, []string{"w-4", "h-4"}},
		{SizeMD, []string{"w-5", "h-5"}},
		{SizeLG, []string{"w-6", "h-6"}},
		{SizeXL, []string{"w-8", "h-8"}},
	}

	for _, tt := range tests {
		suite.Run(string(tt.size), func() {
			result := GetIconSizeClasses(tt.size)
			suite.Equal(tt.expected, result)
		})
	}
}

func (suite *StyleTestSuite) TestGetImageSizeClasses() {
	tests := []struct {
		size     Size
		expected []string
	}{
		{SizeXS, []string{"w-8", "h-8"}},
		{SizeSM, []string{"w-12", "h-12"}},
		{SizeMD, []string{"w-16", "h-16"}},
		{SizeLG, []string{"w-24", "h-24"}},
		{SizeXL, []string{"w-32", "h-32"}},
	}

	for _, tt := range tests {
		suite.Run(string(tt.size), func() {
			result := GetImageSizeClasses(tt.size)
			suite.Equal(tt.expected, result)
		})
	}
}

// ============================================================================
// VARIANT STYLE TESTS
// ============================================================================

func (suite *StyleTestSuite) TestGetButtonVariantClasses() {
	tests := []struct {
		variant     Variant
		colorScheme ColorScheme
		expectedColors []string
	}{
		{VariantSolid, ColorPrimary, []string{"bg-blue-500", "text-white", "hover:bg-blue-600"}},
		{VariantDefault, ColorSecondary, []string{"bg-gray-500", "text-white", "hover:bg-gray-600"}},
		{VariantOutlined, ColorSuccess, []string{"bg-transparent", "border-green-500", "text-green-600"}},
		{VariantGhost, ColorDanger, []string{"bg-transparent", "text-red-600", "hover:bg-red-50"}},
		{VariantFilled, ColorWarning, []string{"bg-amber-100", "text-amber-700", "hover:bg-amber-200"}},
	}

	for _, tt := range tests {
		suite.Run(string(tt.variant)+"_"+string(tt.colorScheme), func() {
			classes := GetButtonVariantClasses(tt.variant, tt.colorScheme)
			suite.NotEmpty(classes)
			
			// Check that expected classes are present
			for _, expectedClass := range tt.expectedColors {
				suite.Contains(classes, expectedClass, "Expected class %s to be present in %v", expectedClass, classes)
			}
		})
	}

	// Test unknown color scheme fallback
	classes := GetButtonVariantClasses(VariantSolid, ColorScheme("unknown"))
	suite.Contains(classes, "bg-blue-500") // Should fallback to blue
	suite.Contains(classes, "text-white")

	// Test invalid variant fallback
	classes = GetButtonVariantClasses(Variant("invalid"), ColorPrimary)
	suite.Contains(classes, "bg-blue-500")
	suite.Contains(classes, "text-white")
}

func (suite *StyleTestSuite) TestGetInputVariantClasses() {
	tests := []struct {
		variant Variant
		state   ValidationState
	}{
		{VariantOutlined, StateDefault},
		{VariantDefault, StateError},
		{VariantFilled, StateSuccess},
		{VariantUnderlined, StateWarning},
	}

	for _, tt := range tests {
		suite.Run(string(tt.variant)+"_"+string(tt.state), func() {
			classes := GetInputVariantClasses(tt.variant, tt.state)
			suite.NotEmpty(classes)
			
			// All should have w-full and focus:outline-none
			suite.Contains(classes, "w-full")
			suite.Contains(classes, "focus:outline-none")
		})
	}
}

// ============================================================================
// BASE CLASS TESTS
// ============================================================================

func (suite *StyleTestSuite) TestGetBaseInputClasses() {
	classes := GetBaseInputClasses()
	
	expectedClasses := []string{
		"block", "w-full", "rounded-md", "focus:outline-none",
		"focus:ring-2", "focus:ring-offset-0", TransitionColors,
		"disabled:cursor-not-allowed", "disabled:opacity-50",
	}
	
	for _, expected := range expectedClasses {
		suite.Contains(classes, expected)
	}
}

func (suite *StyleTestSuite) TestGetBaseButtonClasses() {
	classes := GetBaseButtonClasses()
	
	expectedClasses := []string{
		"inline-flex", "items-center", "justify-center", "font-medium",
		"rounded-md", "focus:outline-none", "focus:ring-2", "focus:ring-offset-2",
		TransitionColors, "disabled:cursor-not-allowed", "disabled:opacity-50",
		"cursor-pointer",
	}
	
	for _, expected := range expectedClasses {
		suite.Contains(classes, expected)
	}
}

func (suite *StyleTestSuite) TestGetBaseCheckboxClasses() {
	classes := GetBaseCheckboxClasses()
	
	expectedClasses := []string{
		"rounded", "border-gray-300", "text-blue-600", "focus:ring-blue-500",
		"focus:ring-2", "focus:ring-offset-0", TransitionColors,
		"disabled:cursor-not-allowed", "disabled:opacity-50", "cursor-pointer",
	}
	
	for _, expected := range expectedClasses {
		suite.Contains(classes, expected)
	}
}

// ============================================================================
// UTILITY STYLE TESTS
// ============================================================================

func (suite *StyleTestSuite) TestGetDisabledClasses() {
	classes := GetDisabledClasses()
	expectedClasses := []string{"opacity-50", "cursor-not-allowed", "pointer-events-none"}
	
	suite.Equal(expectedClasses, classes)
}

func (suite *StyleTestSuite) TestGetLoadingClasses() {
	classes := GetLoadingClasses()
	expectedClasses := []string{"opacity-75", "cursor-wait", "pointer-events-none"}
	
	suite.Equal(expectedClasses, classes)
}

func (suite *StyleTestSuite) TestGetFocusClasses() {
	tests := []struct {
		colorScheme ColorScheme
		expectedColor string
	}{
		{ColorPrimary, "blue"},
		{ColorSecondary, "gray"},
		{ColorSuccess, "green"},
		{ColorDanger, "red"},
		{ColorWarning, "amber"},
		{ColorInfo, "blue"},
		{ColorDefault, "blue"},
	}

	for _, tt := range tests {
		suite.Run(string(tt.colorScheme), func() {
			classes := GetFocusClasses(tt.colorScheme)
			
			suite.Contains(classes, "focus:outline-none")
			suite.Contains(classes, "focus:ring-2")
			suite.Contains(classes, "focus:ring-offset-2")
			
			// Check that the color is used in focus ring
			expectedFocusRing := "focus:ring-" + tt.expectedColor + "-500"
			suite.Contains(classes, expectedFocusRing)
		})
	}
}