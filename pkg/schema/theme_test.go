package schema

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// ThemeTestSuite tests the theme functionality
type ThemeTestSuite struct {
	suite.Suite
}

func TestThemeTestSuite(t *testing.T) {
	suite.Run(t, new(ThemeTestSuite))
}

// ==================== Basic Theme Tests ====================

func (suite *ThemeTestSuite) TestThemeStructure() {
	theme := &Theme{
		ID:          "test-theme-id",
		Name:        "Test Theme",
		Version:     "1.0.0", 
		Description: "A test theme",
		Author:      "Test Author",
		Tokens:      GetDefaultTokens(),
	}
	
	suite.Require().Equal("test-theme-id", theme.ID)
	suite.Require().Equal("Test Theme", theme.Name)
	suite.Require().Equal("1.0.0", theme.Version)
	suite.Require().Equal("A test theme", theme.Description)
	suite.Require().Equal("Test Author", theme.Author)
	suite.Require().NotNil(theme.Tokens)
}

func (suite *ThemeTestSuite) TestTokens() {
	tokens := GetDefaultTokens()
	
	suite.Require().NotNil(tokens)
	suite.Require().NotNil(tokens.Primitives)
	suite.Require().NotNil(tokens.Semantic)
	suite.Require().NotNil(tokens.Components)
}

func (suite *ThemeTestSuite) TestTokenReference() {
	ref := TokenReference("{colors.primary.base}")
	
	suite.Require().True(ref.IsReference())
	suite.Require().Equal("colors.primary.base", ref.Path())
	suite.Require().Equal("{colors.primary.base}", ref.String())
	
	// Test non-reference
	nonRef := TokenReference("#123456")
	suite.Require().False(nonRef.IsReference())
}

func (suite *ThemeTestSuite) TestThemeRegistry() {
	registry := NewThemeRegistry()
	suite.Require().NotNil(registry)
	
	// Test registry methods exist (they return placeholder values for now)
	exists := registry.Exists("test-theme")
	suite.Require().False(exists) // Should be false for non-existent theme
}

func (suite *ThemeTestSuite) TestThemeManager() {
	registry := NewThemeRegistry()
	tokenManager := NewTokenRegistry()
	manager := NewThemeManager(registry, tokenManager)
	suite.Require().NotNil(manager)
}

func (suite *ThemeTestSuite) TestDarkModeConfig() {
	darkModeConfig := &DarkModeConfig{
		Enabled:  true,
		Default:  true,
		Strategy: "auto",
	}
	
	suite.Require().True(darkModeConfig.Enabled)
	suite.Require().True(darkModeConfig.Default)
	suite.Require().Equal("auto", darkModeConfig.Strategy)
}

func (suite *ThemeTestSuite) TestAccessibilityConfig() {
	accessibilityConfig := &AccessibilityConfig{
		HighContrast:      true,
		ReducedMotion:     false,
		FocusIndicator:    true,
		ScreenReaderOnly:  true,
		KeyboardNav:       true,
	}
	
	suite.Require().True(accessibilityConfig.HighContrast)
	suite.Require().False(accessibilityConfig.ReducedMotion)
	suite.Require().True(accessibilityConfig.FocusIndicator)
	suite.Require().True(accessibilityConfig.ScreenReaderOnly)
	suite.Require().True(accessibilityConfig.KeyboardNav)
}

func (suite *ThemeTestSuite) TestThemeWithAllComponents() {
	theme := &Theme{
		ID:          "complete-theme-id",
		Name:        "Complete Theme",
		Version:     "2.0.0",
		Description: "A complete theme with all components",
		Tokens:      GetDefaultTokens(),
		
		DarkMode: &DarkModeConfig{
			Enabled: true,
			Default: true,
		},
		
		Accessibility: &AccessibilityConfig{
			HighContrast:     true,
			FocusIndicator:   true,
			ScreenReaderOnly: true,
		},
		
		CustomCSS: "/* Custom theme styles */",
	}
	
	suite.Require().Equal("complete-theme-id", theme.ID)
	suite.Require().Equal("Complete Theme", theme.Name)
	suite.Require().NotNil(theme.Tokens)
	suite.Require().NotNil(theme.DarkMode)
	suite.Require().NotNil(theme.Accessibility)
	suite.Require().Contains(theme.CustomCSS, "Custom theme styles")
}