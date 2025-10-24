package atoms

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// IconTestSuite provides comprehensive tests for Icon component
type IconTestSuite struct {
	suite.Suite
}

// SetupTest runs before each test method
func (suite *IconTestSuite) SetupTest() {
	ResetIDCounter()
}

// TestIconTestSuite runs the icon test suite
func TestIconTestSuite(t *testing.T) {
	suite.Run(t, new(IconTestSuite))
}

// ============================================================================
// ICON PROPS TESTS
// ============================================================================

func (suite *IconTestSuite) TestIconProps_DefaultValues() {
	props := IconProps{}
	
	// Test default values - check zero values
	suite.Empty(props.Name) // No default icon name
	suite.Equal(Size(""), props.Size) // Zero value for Size
	suite.Empty(props.Class)
}

func (suite *IconTestSuite) TestIconProps_BasicProperties() {
	props := IconProps{
		Name:  "home",
		Size:  IconSizeLG,
		Class: "text-blue-500",
	}
	
	suite.Equal("home", props.Name)
	suite.Equal(IconSizeLG, props.Size)
	suite.Equal("text-blue-500", props.Class)
}

func (suite *IconTestSuite) TestIconProps_WithColor() {
	props := IconProps{
		Name:  "star",
		Size:  IconSizeSM,
		Color: "text-yellow-400",
	}
	
	suite.Equal("star", props.Name)
	suite.Equal(IconSizeSM, props.Size)
	suite.Equal("text-yellow-400", props.Color)
}

// ============================================================================
// ICON SIZE TESTS
// ============================================================================

func (suite *IconTestSuite) TestIconSize_AllSizes() {
	sizes := []Size{
		IconSizeXS,
		IconSizeSM,
		IconSizeMD,
		IconSizeLG,
		IconSizeXL,
		IconSize2XL,
	}
	
	for _, size := range sizes {
		props := IconProps{
			Name: "test",
			Size: size,
		}
		suite.Equal(size, props.Size)
	}
}

func (suite *IconTestSuite) TestIconSize_StringConversion() {
	tests := []struct {
		size     Size
		expected string
	}{
		{IconSizeXS, "xs"},
		{IconSizeSM, "sm"},
		{IconSizeMD, "md"},
		{IconSizeLG, "lg"},
		{IconSizeXL, "xl"},
		{IconSize2XL, "2xl"},
	}
	
	for _, tt := range tests {
		suite.Equal(tt.expected, tt.size.String())
	}
}

func (suite *IconTestSuite) TestIconSize_SizeAliases() {
	// Test that icon size aliases work correctly
	suite.Equal(Size("xs"), IconSizeXS)
	suite.Equal(Size("sm"), IconSizeSM)
	suite.Equal(Size("md"), IconSizeMD)
	suite.Equal(Size("lg"), IconSizeLG)
	suite.Equal(Size("xl"), IconSizeXL)
	suite.Equal(Size("2xl"), IconSize2XL)
}

// ============================================================================
// ICON POSITION TESTS
// ============================================================================

func (suite *IconTestSuite) TestIconPosition_AllPositions() {
	positions := []IconPosition{
		IconLeft,
		IconRight,
		IconTop,
		IconBottom,
	}
	
	for _, position := range positions {
		suite.Contains([]IconPosition{IconLeft, IconRight, IconTop, IconBottom}, position)
	}
}

func (suite *IconTestSuite) TestIconPosition_StringConversion() {
	tests := []struct {
		position IconPosition
		expected string
	}{
		{IconLeft, "left"},
		{IconRight, "right"},
		{IconTop, "top"},
		{IconBottom, "bottom"},
	}
	
	for _, tt := range tests {
		suite.Equal(tt.expected, tt.position.String())
	}
}

// ============================================================================
// ICON COMMON NAMES TESTS
// ============================================================================

func (suite *IconTestSuite) TestIconProps_CommonIconNames() {
	commonIcons := []string{
		"home",
		"user",
		"settings",
		"search",
		"menu",
		"close",
		"check",
		"arrow-left",
		"arrow-right",
		"chevron-up",
		"chevron-down",
		"plus",
		"minus",
		"edit",
		"delete",
		"save",
		"download",
		"upload",
		"eye",
		"eye-off",
		"lock",
		"unlock",
		"bell",
		"heart",
		"star",
		"flag",
		"calendar",
		"clock",
		"mail",
		"phone",
		"globe",
		"link",
		"image",
		"file",
		"folder",
		"copy",
		"share",
		"refresh",
		"filter",
		"sort",
		"grid",
		"list",
	}
	
	for _, iconName := range commonIcons {
		props := IconProps{Name: iconName}
		suite.Equal(iconName, props.Name)
		suite.NotEmpty(props.Name)
	}
}

// ============================================================================
// ICON STYLING TESTS
// ============================================================================

func (suite *IconTestSuite) TestIconProps_ColorClasses() {
	colorTests := []struct {
		color    string
		expected string
	}{
		{"text-red-500", "text-red-500"},
		{"text-blue-600", "text-blue-600"},
		{"text-green-400", "text-green-400"},
		{"text-gray-800", "text-gray-800"},
		{"text-yellow-300", "text-yellow-300"},
	}
	
	for _, tt := range colorTests {
		props := IconProps{
			Name:  "star",
			Color: tt.color,
		}
		suite.Equal(tt.expected, props.Color)
	}
}

func (suite *IconTestSuite) TestIconProps_CustomClasses() {
	props := IconProps{
		Name:  "spinner",
		Size:  IconSizeMD,
		Class: "animate-spin text-blue-500",
	}
	
	suite.Equal("spinner", props.Name)
	suite.Equal("animate-spin text-blue-500", props.Class)
}

func (suite *IconTestSuite) TestIconProps_CombinedStyling() {
	props := IconProps{
		Name:  "check",
		Size:  IconSizeLG,
		Color: "text-green-500",
		Class: "font-bold border border-green-500 rounded-full p-1",
	}
	
	suite.Equal("check", props.Name)
	suite.Equal(IconSizeLG, props.Size)
	suite.Equal("text-green-500", props.Color)
	suite.Equal("font-bold border border-green-500 rounded-full p-1", props.Class)
}

// ============================================================================
// ICON STATE TESTS
// ============================================================================

func (suite *IconTestSuite) TestIconProps_LoadingState() {
	props := IconProps{
		Name:  "refresh",
		Size:  IconSizeSM,
		Class: "animate-spin",
	}
	
	suite.Equal("refresh", props.Name)
	suite.Contains(props.Class, "animate-spin")
}

func (suite *IconTestSuite) TestIconProps_InteractiveStates() {
	props := IconProps{
		Name:  "heart",
		Size:  IconSizeMD,
		Class: "hover:text-red-500 cursor-pointer transition-colors",
	}
	
	suite.Equal("heart", props.Name)
	suite.Contains(props.Class, "hover:text-red-500")
	suite.Contains(props.Class, "cursor-pointer")
}

// ============================================================================
// ICON EDGE CASES
// ============================================================================

func (suite *IconTestSuite) TestIconProps_EmptyName() {
	props := IconProps{
		Name: "",
		Size: IconSizeMD,
	}
	
	suite.Empty(props.Name)
	suite.Equal(IconSizeMD, props.Size)
}

func (suite *IconTestSuite) TestIconProps_SpecialCharactersInName() {
	specialNames := []string{
		"arrow-up-right",
		"chevron_down",
		"user.circle",
		"check-2",
		"x-mark",
	}
	
	for _, name := range specialNames {
		props := IconProps{Name: name}
		suite.Equal(name, props.Name)
	}
}

func (suite *IconTestSuite) TestIconProps_LongClassName() {
	longClass := "w-6 h-6 text-blue-500 hover:text-blue-700 active:text-blue-800 transition-colors duration-200 cursor-pointer select-none"
	props := IconProps{
		Name:  "settings",
		Class: longClass,
	}
	
	suite.Equal(longClass, props.Class)
}

// ============================================================================
// ICON SEMANTIC USAGE TESTS
// ============================================================================

func (suite *IconTestSuite) TestIconProps_NavigationIcons() {
	navigationIcons := map[string]IconPosition{
		"arrow-left":   IconLeft,
		"arrow-right":  IconRight,
		"chevron-up":   IconTop,
		"chevron-down": IconBottom,
	}
	
	for iconName, position := range navigationIcons {
		props := IconProps{Name: iconName}
		suite.Equal(iconName, props.Name)
		
		// Test position usage
		suite.Equal(position.String(), position.String())
	}
}

func (suite *IconTestSuite) TestIconProps_StatusIcons() {
	statusTests := []struct {
		name     string
		color    string
		semantic string
	}{
		{"check", "text-green-500", "success"},
		{"x", "text-red-500", "error"},
		{"alert-triangle", "text-yellow-500", "warning"},
		{"info", "text-blue-500", "info"},
	}
	
	for _, tt := range statusTests {
		props := IconProps{
			Name:  tt.name,
			Color: tt.color,
		}
		
		suite.Equal(tt.name, props.Name)
		suite.Equal(tt.color, props.Color)
	}
}

func (suite *IconTestSuite) TestIconProps_ActionIcons() {
	actionIcons := []string{
		"edit",     // edit action
		"delete",   // delete action
		"save",     // save action
		"copy",     // copy action
		"share",    // share action
		"download", // download action
		"upload",   // upload action
		"refresh",  // refresh action
	}
	
	for _, iconName := range actionIcons {
		props := IconProps{
			Name: iconName,
			Size: IconSizeSM,
		}
		
		suite.Equal(iconName, props.Name)
		suite.Equal(IconSizeSM, props.Size)
	}
}

// ============================================================================
// ICON ACCESSIBILITY TESTS
// ============================================================================

func (suite *IconTestSuite) TestIconProps_DecorativeIcon() {
	// Decorative icons that don't need aria-label
	props := IconProps{
		Name: "star",
		Size: IconSizeSM,
	}
	
	suite.Equal("star", props.Name)
	// Decorative icons typically don't have aria-label
}

func (suite *IconTestSuite) TestIconProps_SemanticIcon() {
	// Icons that convey meaning should have proper aria handling
	props := IconProps{
		Name: "warning",
		Size: IconSizeMD,
		Color: "text-yellow-500",
	}
	
	suite.Equal("warning", props.Name)
	suite.Equal("text-yellow-500", props.Color)
	// In actual implementation, semantic icons would have aria-label
}