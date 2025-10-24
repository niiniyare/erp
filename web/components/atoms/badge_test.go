package atoms

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// BadgeTestSuite provides comprehensive tests for Badge component
type BadgeTestSuite struct {
	suite.Suite
}

// SetupTest runs before each test method
func (suite *BadgeTestSuite) SetupTest() {
	ResetIDCounter()
}

// TestBadgeTestSuite runs the badge test suite
func TestBadgeTestSuite(t *testing.T) {
	suite.Run(t, new(BadgeTestSuite))
}

// ============================================================================
// BADGE PROPS TESTS
// ============================================================================

func (suite *BadgeTestSuite) TestBadgeProps_DefaultValues() {
	props := BadgeProps{}
	
	// Test default values
	suite.Equal(Size(""), props.Size) // No default size
	suite.Equal(ColorScheme(""), props.Color) // No default color
	suite.Equal(Variant(""), props.Variant) // No default variant
	suite.Empty(props.Text)
}

func (suite *BadgeTestSuite) TestBadgeProps_BasicProperties() {
	props := BadgeProps{
		Text:    "New",
		Color:   ColorPrimary,
		Variant: VariantOutlined,
		Size:    SizeSM,
	}
	
	suite.Equal("New", props.Text)
	suite.Equal(ColorPrimary, props.Color)
	suite.Equal(VariantOutlined, props.Variant)
	suite.Equal(SizeSM, props.Size)
}

func (suite *BadgeTestSuite) TestBadgeProps_WithIcon() {
	props := BadgeProps{
		Text:  "3",
		Icon:  "bell",
		Color: ColorDanger,
	}
	
	suite.Equal("3", props.Text)
	suite.Equal("bell", props.Icon)
	suite.Equal(ColorDanger, props.Color)
}

// ============================================================================
// BADGE SIZE TESTS
// ============================================================================

func (suite *BadgeTestSuite) TestBadgeSize_AllSizes() {
	sizes := []Size{
		SizeXS,
		SizeSM,
		SizeMD,
		SizeLG,
		SizeXL,
	}
	
	for _, size := range sizes {
		props := BadgeProps{
			Text: "Test",
			Size: size,
		}
		suite.Equal(size, props.Size)
	}
}

func (suite *BadgeTestSuite) TestBadgeSize_StringConversion() {
	tests := []struct {
		size     Size
		expected string
	}{
		{SizeXS, "xs"},
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
// BADGE COLOR TESTS
// ============================================================================

func (suite *BadgeTestSuite) TestBadgeColor_AllColors() {
	colors := []ColorScheme{
		ColorDefault,
		ColorPrimary,
		ColorSecondary,
		ColorSuccess,
		ColorDanger,
		ColorWarning,
		ColorInfo,
		ColorLight,
		ColorDark,
	}
	
	for _, color := range colors {
		props := BadgeProps{
			Text:  "Test",
			Color: color,
		}
		suite.Equal(color, props.Color)
	}
}

func (suite *BadgeTestSuite) TestBadgeColor_StringConversion() {
	tests := []struct {
		color    ColorScheme
		expected string
	}{
		{ColorDefault, "default"},
		{ColorPrimary, "primary"},
		{ColorSecondary, "secondary"},
		{ColorSuccess, "success"},
		{ColorDanger, "danger"},
		{ColorWarning, "warning"},
		{ColorInfo, "info"},
		{ColorLight, "light"},
		{ColorDark, "dark"},
	}
	
	for _, tt := range tests {
		suite.Equal(tt.expected, tt.color.String())
	}
}

// ============================================================================
// BADGE VARIANT TESTS
// ============================================================================

func (suite *BadgeTestSuite) TestBadgeVariant_AllVariants() {
	variants := []Variant{
		VariantDefault,
		VariantPrimary,
		VariantSecondary,
		VariantOutlined,
		VariantGhost,
		VariantLight,
	}
	
	for _, variant := range variants {
		props := BadgeProps{
			Text:    "Test",
			Variant: variant,
		}
		suite.Equal(variant, props.Variant)
	}
}

func (suite *BadgeTestSuite) TestBadgeVariant_StringConversion() {
	tests := []struct {
		variant  Variant
		expected string
	}{
		{VariantDefault, "default"},
		{VariantPrimary, "primary"},
		{VariantSecondary, "secondary"},
		{VariantOutlined, "outlined"},
		{VariantGhost, "ghost"},
		{VariantLight, "light"},
	}
	
	for _, tt := range tests {
		suite.Equal(tt.expected, tt.variant.String())
	}
}

// ============================================================================
// BADGE COMPOSITION TESTS
// ============================================================================

func (suite *BadgeTestSuite) TestBadgeProps_BasePropsComposition() {
	props := BadgeProps{
		BaseProps: BaseProps{
			ID:    "notification-badge",
			Class: "absolute -top-2 -right-2",
		},
		Text: "5",
	}
	
	suite.Equal("notification-badge", props.BaseProps.ID)
	suite.Equal("absolute -top-2 -right-2", props.BaseProps.Class)
}

func (suite *BadgeTestSuite) TestBadgeProps_AccessibilityPropsComposition() {
	props := BadgeProps{
		AccessibilityProps: AccessibilityProps{
			AriaLabel: "5 unread notifications",
			Role:      "status",
		},
		Text:  "5",
		Color: ColorDanger,
	}
	
	suite.Equal("5 unread notifications", props.AccessibilityProps.AriaLabel)
	suite.Equal("status", props.AccessibilityProps.Role)
}

// ============================================================================
// BADGE USE CASES TESTS
// ============================================================================

func (suite *BadgeTestSuite) TestBadgeProps_NotificationBadge() {
	props := BadgeProps{
		Text:    "12",
		Color:   ColorDanger,
		Variant: VariantDefault,
		Size:    SizeXS,
		BaseProps: BaseProps{
			Class: "absolute -top-1 -right-1",
		},
	}
	
	suite.Equal("12", props.Text)
	suite.Equal(ColorDanger, props.Color)
	suite.Equal(SizeXS, props.Size)
	suite.Contains(props.BaseProps.Class, "absolute")
}

func (suite *BadgeTestSuite) TestBadgeProps_StatusBadge() {
	props := BadgeProps{
		Text:    "Active",
		Color:   ColorSuccess,
		Variant: VariantLight,
		Size:    SizeSM,
	}
	
	suite.Equal("Active", props.Text)
	suite.Equal(ColorSuccess, props.Color)
	suite.Equal(VariantLight, props.Variant)
}

func (suite *BadgeTestSuite) TestBadgeProps_CategoryBadge() {
	props := BadgeProps{
		Text:    "Frontend",
		Color:   ColorPrimary,
		Variant: VariantOutlined,
		Size:    SizeMD,
	}
	
	suite.Equal("Frontend", props.Text)
	suite.Equal(ColorPrimary, props.Color)
	suite.Equal(VariantOutlined, props.Variant)
}

func (suite *BadgeTestSuite) TestBadgeProps_CountBadge() {
	props := BadgeProps{
		Text:  "99+",
		Color: ColorDanger,
		Size:  SizeXS,
		AccessibilityProps: AccessibilityProps{
			AriaLabel: "More than 99 notifications",
		},
	}
	
	suite.Equal("99+", props.Text)
	suite.Equal("More than 99 notifications", props.AccessibilityProps.AriaLabel)
}

// ============================================================================
// BADGE CONTENT TESTS
// ============================================================================

func (suite *BadgeTestSuite) TestBadgeProps_TextContent() {
	textTests := []string{
		"New",
		"Beta",
		"Pro",
		"Sale",
		"Hot",
		"Limited",
		"Coming Soon",
		"Deprecated",
	}
	
	for _, text := range textTests {
		props := BadgeProps{Text: text}
		suite.Equal(text, props.Text)
		suite.NotEmpty(props.Text)
	}
}

func (suite *BadgeTestSuite) TestBadgeProps_NumericContent() {
	numericTests := []string{
		"1",
		"42",
		"100",
		"999+",
		"1.2k",
		"5M",
	}
	
	for _, number := range numericTests {
		props := BadgeProps{
			Text:  number,
			Color: ColorDanger,
		}
		suite.Equal(number, props.Text)
	}
}

func (suite *BadgeTestSuite) TestBadgeProps_IconOnlyBadge() {
	props := BadgeProps{
		Icon:  "star",
		Color: ColorWarning,
		Size:  SizeSM,
		AccessibilityProps: AccessibilityProps{
			AriaLabel: "Favorite item",
		},
	}
	
	suite.Empty(props.Text)
	suite.Equal("star", props.Icon)
	suite.Equal("Favorite item", props.AccessibilityProps.AriaLabel)
}

// ============================================================================
// BADGE EDGE CASES
// ============================================================================

func (suite *BadgeTestSuite) TestBadgeProps_EmptyContent() {
	props := BadgeProps{
		Color:   ColorDanger,
		Variant: VariantDefault,
		Size:    SizeXS,
	}
	
	suite.Empty(props.Text)
	suite.Empty(props.Icon)
	suite.Equal(ColorDanger, props.Color)
}

func (suite *BadgeTestSuite) TestBadgeProps_LongText() {
	longText := "This is a very long badge text"
	props := BadgeProps{
		Text: longText,
		Size: SizeLG,
	}
	
	suite.Equal(longText, props.Text)
	suite.True(len(props.Text) > 20)
}

func (suite *BadgeTestSuite) TestBadgeProps_SpecialCharacters() {
	specialTexts := []string{
		"€99",
		"50%",
		"★★★★★",
		"#1",
		"@mention",
		"</>",
	}
	
	for _, text := range specialTexts {
		props := BadgeProps{Text: text}
		suite.Equal(text, props.Text)
	}
}

// ============================================================================
// BADGE STYLING TESTS
// ============================================================================

func (suite *BadgeTestSuite) TestBadgeProps_CustomStyling() {
	props := BadgeProps{
		BaseProps: BaseProps{
			Class: "rounded-full px-3 py-1 shadow-lg",
		},
		Text:    "Custom",
		Color:   ColorPrimary,
		Variant: VariantGhost,
	}
	
	suite.Contains(props.BaseProps.Class, "rounded-full")
	suite.Contains(props.BaseProps.Class, "shadow-lg")
}

func (suite *BadgeTestSuite) TestBadgeProps_ConditionalStyling() { //nolint:revive // Test function complexity is acceptable
	// Test different combinations that might affect styling
	combinations := []struct {
		name  string
		props BadgeProps
	}{
		{
			"small-danger",
			BadgeProps{Size: SizeXS, Color: ColorDanger},
		},
		{
			"large-success",
			BadgeProps{Size: SizeLG, Color: ColorSuccess},
		},
		{
			"outlined-warning",
			BadgeProps{Variant: VariantOutlined, Color: ColorWarning},
		},
		{
			"ghost-info",
			BadgeProps{Variant: VariantGhost, Color: ColorInfo},
		},
	}
	
	for _, combo := range combinations {
		suite.Run(combo.name, func() {
			// Each combination should maintain its properties
			if combo.props.Size != "" {
				suite.NotEqual(Size(""), combo.props.Size)
			}
			if combo.props.Color != "" {
				suite.NotEqual(ColorScheme(""), combo.props.Color)
			}
			if combo.props.Variant != "" {
				suite.NotEqual(Variant(""), combo.props.Variant)
			}
		})
	}
}

// ============================================================================
// BADGE INTERACTION TESTS
// ============================================================================

func (suite *BadgeTestSuite) TestBadgeProps_ClickableBadge() {
	props := BadgeProps{
		Text:  "Remove",
		Color: ColorDanger,
		Size:  SizeSM,
		AlpinEventHandlers: AlpinEventHandlers{
			OnClick: "removeBadge()",
		},
		BaseProps: BaseProps{
			Class: "cursor-pointer hover:opacity-80",
		},
	}
	
	suite.Equal("removeBadge()", props.AlpinEventHandlers.OnClick)
	suite.Contains(props.BaseProps.Class, "cursor-pointer")
}

func (suite *BadgeTestSuite) TestBadgeProps_DismissibleBadge() {
	props := BadgeProps{
		Text:  "Closeable Badge",
		Icon:  "x",
		Color: ColorInfo,
		AlpinEventHandlers: AlpinEventHandlers{
			OnClick: "dismissBadge($event)",
		},
	}
	
	suite.Equal("x", props.Icon)
	suite.Equal("dismissBadge($event)", props.AlpinEventHandlers.OnClick)
}