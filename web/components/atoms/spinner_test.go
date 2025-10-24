package atoms

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// SpinnerTestSuite provides comprehensive tests for Spinner component
type SpinnerTestSuite struct {
	suite.Suite
}

// SetupTest runs before each test method
func (suite *SpinnerTestSuite) SetupTest() {
	ResetIDCounter()
}

// TestSpinnerTestSuite runs the spinner test suite
func TestSpinnerTestSuite(t *testing.T) {
	suite.Run(t, new(SpinnerTestSuite))
}

// ============================================================================
// SPINNER PROPS TESTS
// ============================================================================

func (suite *SpinnerTestSuite) TestSpinnerProps_DefaultValues() {
	props := SpinnerProps{}
	
	// Test default values - check zero values
	suite.Equal(Size(""), props.Size)      // Zero value for Size
	suite.Equal(ColorScheme(""), props.Color) // Zero value for ColorScheme
	suite.Empty(props.Label)  // No default label
}

func (suite *SpinnerTestSuite) TestSpinnerProps_BasicProperties() {
	props := SpinnerProps{
		Size:  SizeLG, // Use general Size constants
		Color: ColorPrimary, // Use ColorScheme constants
		Label: "Loading...",
	}
	
	suite.Equal(SizeLG, props.Size)
	suite.Equal(ColorPrimary, props.Color)
	suite.Equal("Loading...", props.Label)
}

func (suite *SpinnerTestSuite) TestSpinnerProps_WithCustomClass() {
	props := SpinnerProps{
		BaseProps: BaseProps{
			Class: "animate-pulse opacity-75",
		},
		Size: SizeSM, // Use general Size constants
	}
	
	suite.Equal("animate-pulse opacity-75", props.BaseProps.Class)
	suite.Equal(SizeSM, props.Size)
}

// ============================================================================
// SPINNER SIZE TESTS
// ============================================================================

func (suite *SpinnerTestSuite) TestSpinnerSize_AllSizes() {
	sizes := []Size{
		SizeXS, // Use general Size constants
		SizeSM,
		SizeMD,
		SizeLG,
		SizeXL,
	}
	
	for _, size := range sizes {
		props := SpinnerProps{Size: size}
		suite.Equal(size, props.Size)
	}
}

func (suite *SpinnerTestSuite) TestSpinnerSize_StringConversion() {
	tests := []struct {
		size     Size
		expected string
	}{
		{SizeXS, "xs"}, // Use general Size constants
		{SizeSM, "sm"},
		{SizeMD, "md"},
		{SizeLG, "lg"},
		{SizeXL, "xl"},
	}
	
	for _, tt := range tests {
		suite.Equal(tt.expected, tt.size.String())
	}
}

func (suite *SpinnerTestSuite) TestSpinnerSize_SizeAliases() {
	// Test that general size constants work correctly
	suite.Equal(Size("xs"), SizeXS)
	suite.Equal(Size("sm"), SizeSM)
	suite.Equal(Size("md"), SizeMD)
	suite.Equal(Size("lg"), SizeLG)
	suite.Equal(Size("xl"), SizeXL)
}

// ============================================================================
// SPINNER COLOR TESTS
// ============================================================================

func (suite *SpinnerTestSuite) TestSpinnerProps_ColorVariations() {
	colors := []ColorScheme{
		ColorPrimary,
		ColorSecondary,
		ColorSuccess,
		ColorDanger,
		ColorWarning,
		ColorInfo,
		ColorGray,
		ColorLight,
	}
	
	for _, color := range colors {
		props := SpinnerProps{
			Color: color,
			Size:  SizeMD, // Use general Size constants
		}
		suite.Equal(color, props.Color)
	}
}

func (suite *SpinnerTestSuite) TestSpinnerProps_CustomColorClasses() {
	props := SpinnerProps{
		Color: ColorInfo, // Use ColorScheme instead of custom string
		Size:  SizeLG, // Use general Size constants
	}
	
	suite.Equal(ColorInfo, props.Color)
	suite.Equal(SizeLG, props.Size)
}

// ============================================================================
// SPINNER TEXT TESTS
// ============================================================================

func (suite *SpinnerTestSuite) TestSpinnerProps_LoadingTexts() {
	loadingTexts := []string{
		"Loading...",
		"Please wait...",
		"Processing...",
		"Saving...",
		"Uploading...",
		"Downloading...",
		"Connecting...",
		"Authenticating...",
	}
	
	for _, text := range loadingTexts {
		props := SpinnerProps{
			Label: text,
			Size: SizeMD, // Use general Size constants
		}
		suite.Equal(text, props.Label)
		suite.NotEmpty(props.Label)
	}
}

func (suite *SpinnerTestSuite) TestSpinnerProps_NoText() {
	props := SpinnerProps{
		Size:  SizeSM, // Use general Size constants
		Color: ColorPrimary, // Use ColorScheme constants
	}
	
	suite.Empty(props.Label)
	suite.Equal(SizeSM, props.Size)
}

func (suite *SpinnerTestSuite) TestSpinnerProps_LongText() {
	longText := "This is a very long loading message that might wrap to multiple lines when displayed"
	props := SpinnerProps{
		Label: longText,
		Size: SizeLG, // Use general Size constants
	}
	
	suite.Equal(longText, props.Label)
	suite.True(len(props.Label) > 50)
}

// ============================================================================
// SPINNER COMPOSITION TESTS
// ============================================================================

func (suite *SpinnerTestSuite) TestSpinnerProps_BasePropsComposition() {
	props := SpinnerProps{
		BaseProps: BaseProps{
			ID:    "loading-spinner",
			Class: "absolute inset-0 flex items-center justify-center",
		},
		Label: "Loading content...",
	}
	
	suite.Equal("loading-spinner", props.BaseProps.ID)
	suite.Contains(props.BaseProps.Class, "absolute")
	suite.Contains(props.BaseProps.Class, "flex")
}

func (suite *SpinnerTestSuite) TestSpinnerProps_AccessibilityPropsComposition() {
	props := SpinnerProps{
		AccessibilityProps: AccessibilityProps{
			AriaLabel: "Loading content",
			Role:      "status",
		},
		Label: "Please wait...",
	}
	
	suite.Equal("Loading content", props.AccessibilityProps.AriaLabel)
	suite.Equal("status", props.AccessibilityProps.Role)
}

// ============================================================================
// SPINNER USE CASES TESTS
// ============================================================================

func (suite *SpinnerTestSuite) TestSpinnerProps_ButtonSpinner() {
	// Spinner used inside a button
	props := SpinnerProps{
		Size:  SizeXS, // Use general Size constants
		Color: ColorLight, // Use ColorScheme constants
	}
	
	suite.Equal(SizeXS, props.Size)
	suite.Equal(ColorLight, props.Color)
	suite.Empty(props.Label) // Button spinners usually don't have text
}

func (suite *SpinnerTestSuite) TestSpinnerProps_PageSpinner() {
	// Full page loading spinner
	props := SpinnerProps{
		Size:  SizeXL, // Use general Size constants
		Color: ColorPrimary, // Use ColorScheme constants
		Label: "Loading application...",
		BaseProps: BaseProps{
			Class: "fixed inset-0 bg-white bg-opacity-75 flex items-center justify-center z-50",
		},
	}
	
	suite.Equal(SizeXL, props.Size)
	suite.Equal("Loading application...", props.Label)
	suite.Contains(props.BaseProps.Class, "fixed")
	suite.Contains(props.BaseProps.Class, "z-50")
}

func (suite *SpinnerTestSuite) TestSpinnerProps_InlineSpinner() {
	// Inline spinner for content sections
	props := SpinnerProps{
		Size:  SizeSM, // Use general Size constants
		Color: ColorGray, // Use ColorScheme constants
		Label: "Refreshing data...",
		BaseProps: BaseProps{
			Class: "inline-flex items-center space-x-2",
		},
	}
	
	suite.Equal(SizeSM, props.Size)
	suite.Equal("Refreshing data...", props.Label)
	suite.Contains(props.BaseProps.Class, "inline-flex")
}

func (suite *SpinnerTestSuite) TestSpinnerProps_FormSpinner() {
	// Spinner for form submission
	props := SpinnerProps{
		Size:  SizeMD, // Use general Size constants
		Color: ColorSuccess, // Use ColorScheme constants
		Label: "Submitting form...",
		AccessibilityProps: AccessibilityProps{
			AriaLabel: "Form submission in progress",
			Role:      "status",
		},
	}
	
	suite.Equal("Submitting form...", props.Label)
	suite.Equal("Form submission in progress", props.AccessibilityProps.AriaLabel)
	suite.Equal("status", props.AccessibilityProps.Role)
}

// ============================================================================
// SPINNER STATE TESTS
// ============================================================================

func (suite *SpinnerTestSuite) TestSpinnerProps_ConditionalDisplay() {
	// Test props that might be used with conditional display
	props := SpinnerProps{
		Size:  SizeMD, // Use general Size constants
		Color: ColorPrimary, // Use ColorScheme constants
		Label: "Loading...",
		BaseProps: BaseProps{
			Class: "transition-opacity duration-300",
		},
	}
	
	suite.Contains(props.BaseProps.Class, "transition-opacity")
	suite.Equal("Loading...", props.Label)
}

func (suite *SpinnerTestSuite) TestSpinnerProps_DifferentStates() {
	// Test different loading states
	states := []struct {
		name  string
		props SpinnerProps
	}{
		{
			"initial-load",
			SpinnerProps{Size: SizeLG, Label: "Loading..."}, // Use general Size constants
		},
		{
			"saving",
			SpinnerProps{Size: SizeMD, Label: "Saving...", Color: ColorSuccess}, // Use ColorScheme constants
		},
		{
			"deleting",
			SpinnerProps{Size: SizeSM, Label: "Deleting...", Color: ColorDanger}, // Use ColorScheme constants
		},
		{
			"processing",
			SpinnerProps{Size: SizeXL, Label: "Processing...", Color: ColorInfo}, // Use ColorScheme constants
		},
	}
	
	for _, state := range states {
		suite.Run(state.name, func() {
			suite.NotEmpty(state.props.Label)
			suite.NotEqual(Size(""), state.props.Size)
		})
	}
}

// ============================================================================
// SPINNER EDGE CASES
// ============================================================================

func (suite *SpinnerTestSuite) TestSpinnerProps_MinimalSpinner() {
	// Minimal spinner with just size
	props := SpinnerProps{
		Size: SizeSM, // Use general Size constants
	}
	
	suite.Equal(SizeSM, props.Size)
	suite.Empty(props.Color)
	suite.Empty(props.Label)
}

func (suite *SpinnerTestSuite) TestSpinnerProps_SpinnerWithoutAnimation() {
	// Spinner that might not have default animation (custom styling)
	props := SpinnerProps{
		Size: SizeMD, // Use general Size constants
		BaseProps: BaseProps{
			Class: "animate-bounce", // Different animation
		},
	}
	
	suite.Contains(props.BaseProps.Class, "animate-bounce")
	suite.Equal(SizeMD, props.Size)
}

func (suite *SpinnerTestSuite) TestSpinnerProps_CustomSpinnerElement() {
	// Custom spinner with specific styling
	props := SpinnerProps{
		Size:  SizeLG, // Use general Size constants
		Color: ColorInfo, // Use ColorScheme constants
		BaseProps: BaseProps{
			Class: "animate-pulse rounded-full border-4 border-t-transparent border-indigo-600",
		},
	}
	
	suite.Equal(ColorInfo, props.Color)
	suite.Contains(props.BaseProps.Class, "border-4")
	suite.Contains(props.BaseProps.Class, "border-t-transparent")
}

// ============================================================================
// SPINNER ACCESSIBILITY TESTS
// ============================================================================

func (suite *SpinnerTestSuite) TestSpinnerProps_AccessibilityAttributes() {
	props := SpinnerProps{
		Size: SizeMD, // Use general Size constants
		Label: "Loading data...",
		AccessibilityProps: AccessibilityProps{
			AriaLabel: "Loading in progress",
			Role:      "status",
			AriaLive:  "polite",
			// Note: AriaAtomic is not a field in AccessibilityProps
		},
	}
	
	suite.Equal("Loading in progress", props.AccessibilityProps.AriaLabel)
	suite.Equal("status", props.AccessibilityProps.Role)
	suite.Equal("polite", props.AccessibilityProps.AriaLive)
}

func (suite *SpinnerTestSuite) TestSpinnerProps_ScreenReaderText() {
	props := SpinnerProps{
		Size: SizeSM, // Use general Size constants
		Label: "Loading", // Visible text
		AccessibilityProps: AccessibilityProps{
			AriaLabel: "Content is loading, please wait", // More descriptive for screen readers
		},
	}
	
	suite.Equal("Loading", props.Label)
	suite.Equal("Content is loading, please wait", props.AccessibilityProps.AriaLabel)
}