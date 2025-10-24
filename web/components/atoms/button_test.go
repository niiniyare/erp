package atoms

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// ButtonTestSuite provides comprehensive tests for Button component
type ButtonTestSuite struct {
	suite.Suite
}

// SetupTest runs before each test method
func (suite *ButtonTestSuite) SetupTest() {
	ResetIDCounter()
}

// TestButtonTestSuite runs the button test suite
func TestButtonTestSuite(t *testing.T) {
	suite.Run(t, new(ButtonTestSuite))
}

// ============================================================================
// BUTTON PROPS TESTS
// ============================================================================

func (suite *ButtonTestSuite) TestButtonProps_DefaultValues() {
	props := ButtonProps{}
	
	// Test default values - check zero values
	suite.Equal(Variant(""), props.Variant) // Zero value for Variant
	suite.Equal(Size(""), props.Size)       // Zero value for Size
	suite.False(props.InteractionProps.Disabled) // Disabled is in InteractionProps
	suite.False(props.Loading)
}

func (suite *ButtonTestSuite) TestButtonProps_WithIcon() {
	props := ButtonProps{
		Text: "Save",
		Icon: IconProps{
			Name: "save",
			Size: IconSizeSM,
		},
		Variant: VariantPrimary,
	}
	
	suite.Equal("Save", props.Text)
	suite.Equal("save", props.Icon.Name)
	suite.Equal(IconSizeSM, props.Icon.Size)
	suite.Equal(VariantPrimary, props.Variant)
}

func (suite *ButtonTestSuite) TestButtonProps_LoadingState() {
	props := ButtonProps{
		Text:    "Submit",
		Loading: true,
		Variant: VariantPrimary,
	}
	
	suite.True(props.Loading)
	suite.Equal("Submit", props.Text)
}

func (suite *ButtonTestSuite) TestButtonProps_DisabledState() {
	props := ButtonProps{
		Text: "Disabled Button",
		InteractionProps: InteractionProps{
			Disabled: true,
		},
	}
	
	suite.True(props.InteractionProps.Disabled)
}

func (suite *ButtonTestSuite) TestButtonProps_HTMXAttributes() {
	props := ButtonProps{
		Text:     "HTMX Button",
		HxPost:   "/api/submit",
		HxTarget: "#result",
		HxSwap:   "innerHTML",
	}
	
	suite.Equal("/api/submit", props.HxPost)
	suite.Equal("#result", props.HxTarget)
	suite.Equal("innerHTML", props.HxSwap)
}

// ============================================================================
// BUTTON VARIANT TESTS
// ============================================================================

func (suite *ButtonTestSuite) TestButtonVariant_Types() {
	variants := []Variant{
		VariantDefault,
		VariantPrimary,
		VariantSecondary,
		VariantOutlined,
		VariantGhost,
		VariantLight,
		VariantDark,
	}
	
	for _, variant := range variants {
		props := ButtonProps{Variant: variant}
		suite.Equal(variant, props.Variant)
	}
}

func (suite *ButtonTestSuite) TestButtonVariant_StringConversion() {
	tests := []struct {
		variant  Variant
		expected string
	}{
		{VariantDefault, "default"},
		{VariantPrimary, "primary"},
		{VariantSecondary, "secondary"},
		{VariantOutlined, "outlined"},
		{VariantGhost, "ghost"},
	}
	
	for _, tt := range tests {
		suite.Equal(tt.expected, tt.variant.String())
	}
}

// ============================================================================
// BUTTON SIZE TESTS
// ============================================================================

func (suite *ButtonTestSuite) TestButtonSize_Types() {
	sizes := []Size{
		SizeXS, // Use the general Size constants
		SizeSM,
		SizeMD,
		SizeLG,
		SizeXL,
	}
	
	for _, size := range sizes {
		props := ButtonProps{Size: size}
		suite.Equal(size, props.Size)
	}
}

func (suite *ButtonTestSuite) TestButtonSize_StringConversion() {
	tests := []struct {
		size     Size
		expected string
	}{
		{SizeXS, "xs"}, // Use the general Size constants
		{SizeSM, "sm"},
		{SizeMD, "md"},
		{SizeLG, "lg"},
		{SizeXL, "xl"},
	}
	
	for _, tt := range tests {
		suite.Equal(tt.expected, tt.size.String())
	}
}

// ============================================================================
// BUTTON COMPOSITION TESTS
// ============================================================================

func (suite *ButtonTestSuite) TestButtonProps_BasePropsComposition() {
	props := ButtonProps{
		BaseProps: BaseProps{
			ID:    "test-button",
			Class: "custom-class",
		},
		Text: "Test",
	}
	
	suite.Equal("test-button", props.BaseProps.ID)
	suite.Equal("custom-class", props.BaseProps.Class)
}

func (suite *ButtonTestSuite) TestButtonProps_InteractionPropsComposition() {
	props := ButtonProps{
		InteractionProps: InteractionProps{
			Disabled: true,
			ReadOnly: false,
		},
		Text: "Test",
	}
	
	suite.True(props.InteractionProps.Disabled)
	suite.False(props.InteractionProps.ReadOnly)
	suite.False(props.InteractionProps.IsInteractive())
}

func (suite *ButtonTestSuite) TestButtonProps_AccessibilityPropsComposition() {
	ariaPressed := true
	ariaExpanded := false
	props := ButtonProps{
		AccessibilityProps: AccessibilityProps{
			AriaLabel:    "Close dialog",
			AriaPressed:  &ariaPressed,  // AriaPressed is *bool
			AriaExpanded: &ariaExpanded, // AriaExpanded is *bool
		},
		Icon: IconProps{Name: "x"},
	}
	
	suite.Equal("Close dialog", props.AccessibilityProps.AriaLabel)
	suite.True(*props.AccessibilityProps.AriaPressed)
	suite.False(*props.AccessibilityProps.AriaExpanded)
}

func (suite *ButtonTestSuite) TestButtonProps_AlpineEventHandlers() {
	props := ButtonProps{
		AlpinEventHandlers: AlpinEventHandlers{
			OnClick: "handleClick()",
			OnFocus: "handleFocus()",
			OnBlur:  "handleBlur()",
		},
		Text: "Interactive",
	}
	
	suite.Equal("handleClick()", props.AlpinEventHandlers.OnClick)
	suite.Equal("handleFocus()", props.AlpinEventHandlers.OnFocus)
	suite.Equal("handleBlur()", props.AlpinEventHandlers.OnBlur)
}

// ============================================================================
// BUTTON EDGE CASES
// ============================================================================

func (suite *ButtonTestSuite) TestButtonProps_EmptyText() {
	props := ButtonProps{
		Icon: IconProps{Name: "save"},
	}
	
	suite.Empty(props.Text)
	suite.Equal("save", props.Icon.Name)
}

func (suite *ButtonTestSuite) TestButtonProps_IconOnlyButton() {
	props := ButtonProps{
		Icon: IconProps{
			Name: "menu",
			Size: IconSizeMD,
		},
		AccessibilityProps: AccessibilityProps{
			AriaLabel: "Open menu",
		},
	}
	
	suite.Empty(props.Text)
	suite.Equal("menu", props.Icon.Name)
	suite.Equal("Open menu", props.AccessibilityProps.AriaLabel)
}

func (suite *ButtonTestSuite) TestButtonProps_LoadingWithIcon() {
	props := ButtonProps{
		Text:    "Submitting...",
		Loading: true,
		Icon:    IconProps{Name: "save"},
	}
	
	suite.True(props.Loading)
	suite.Equal("save", props.Icon.Name)
	suite.Equal("Submitting...", props.Text)
}

// ============================================================================
// BUTTON TYPE TESTS
// ============================================================================

func (suite *ButtonTestSuite) TestButtonProps_ButtonType() {
	tests := []struct {
		name         string
		buttonType   string
		expectedType string
	}{
		{"default type", "", ""},
		{"submit type", "submit", "submit"},
		{"reset type", "reset", "reset"},
		{"button type", "button", "button"},
	}
	
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			props := ButtonProps{
				Type: tt.buttonType,
				Text: "Test",
			}
			suite.Equal(tt.expectedType, props.Type)
		})
	}
}

// ============================================================================
// BACKWARD COMPATIBILITY TESTS
// ============================================================================

func (suite *ButtonTestSuite) TestButtonProps_DeprecatedConstants() {
	// Test that deprecated constants still work
	props := ButtonProps{
		Variant: ButtonPrimary, // This is an alias for VariantPrimary
		Size:    SizeLG,        // Use the standard size constant
	}
	
	suite.Equal(VariantPrimary, props.Variant)
	suite.Equal(SizeLG, props.Size)
}

func (suite *ButtonTestSuite) TestIconButton_Alias() {
	// IconButton should be an alias for ButtonProps
	iconButton := IconButton{
		Icon: IconProps{Name: "home"},
	}
	
	// Should be convertible to ButtonProps
	buttonProps := ButtonProps(iconButton)
	suite.Equal("home", buttonProps.Icon.Name)
}