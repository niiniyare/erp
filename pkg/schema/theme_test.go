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
		Name:        "Test Theme",
		Version:     "1.0.0", 
		Description: "A test theme",
		Author:      "Test Author",
	}
	
	suite.Require().Equal("Test Theme", theme.Name)
	suite.Require().Equal("1.0.0", theme.Version)
	suite.Require().Equal("A test theme", theme.Description)
	suite.Require().Equal("Test Author", theme.Author)
}

func (suite *ThemeTestSuite) TestFormTheme() {
	formTheme := &FormTheme{
		Background: "#ffffff",
		Padding:    "16px",
		Border:     "1px solid #e5e5e5",
		BorderRadius: "8px",
	}
	
	suite.Require().Equal("#ffffff", formTheme.Background)
	suite.Require().Equal("16px", formTheme.Padding)
	suite.Require().Equal("1px solid #e5e5e5", formTheme.Border)
	suite.Require().Equal("8px", formTheme.BorderRadius)
}

func (suite *ThemeTestSuite) TestFieldTheme() {
	fieldTheme := &FieldTheme{
		Background: "#f9f9f9",
		Border:     "1px solid #d1d5db",
		Padding:    "12px",
		BorderRadius: "6px",
	}
	
	suite.Require().Equal("#f9f9f9", fieldTheme.Background)
	suite.Require().Equal("1px solid #d1d5db", fieldTheme.Border)
	suite.Require().Equal("12px", fieldTheme.Padding)
	suite.Require().Equal("6px", fieldTheme.BorderRadius)
}

func (suite *ThemeTestSuite) TestButtonTheme() {
	buttonTheme := &ButtonTheme{
		PrimaryBackground: "#3b82f6",
		PrimaryColor:      "#ffffff", 
		Padding:           "10px 20px",
		BorderRadius:      "6px",
		CustomCSS:         "border: none;",
	}
	
	suite.Require().Equal("#3b82f6", buttonTheme.PrimaryBackground)
	suite.Require().Equal("#ffffff", buttonTheme.PrimaryColor)
	suite.Require().Equal("10px 20px", buttonTheme.Padding)
	suite.Require().Equal("6px", buttonTheme.BorderRadius)
	suite.Require().Equal("border: none;", buttonTheme.CustomCSS)
}

func (suite *ThemeTestSuite) TestLayoutTheme() {
	layoutTheme := &LayoutTheme{
		Colors: map[string]string{
			"background": "#ffffff",
			"border":     "#e5e7eb",
		},
		Spacing: map[string]string{
			"gap":    "16px",
			"margin": "8px",
		},
		BorderRadius: "6px",
		CustomCSS:    "display: grid;",
	}
	
	suite.Require().Equal("#ffffff", layoutTheme.Colors["background"])
	suite.Require().Equal("#e5e7eb", layoutTheme.Colors["border"])
	suite.Require().Equal("16px", layoutTheme.Spacing["gap"])
	suite.Require().Equal("8px", layoutTheme.Spacing["margin"])
	suite.Require().Equal("6px", layoutTheme.BorderRadius)
	suite.Require().Equal("display: grid;", layoutTheme.CustomCSS)
}

func (suite *ThemeTestSuite) TestDarkModeConfig() {
	darkModeConfig := &DarkModeConfig{
		Enabled:  true,
		Default:  true,
		Toggle:   true,
		Strategy: "auto",
	}
	
	suite.Require().True(darkModeConfig.Enabled)
	suite.Require().True(darkModeConfig.Default)
	suite.Require().True(darkModeConfig.Toggle)
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
		Name:        "Complete Theme",
		Version:     "2.0.0",
		Description: "A complete theme with all components",
		
		Form: &FormTheme{
			Background: "#ffffff",
			Padding:    "24px",
		},
		
		Field: &FieldTheme{
			Background: "#f8f9fa",
			Border:     "1px solid #dee2e6",
		},
		
		Button: &ButtonTheme{
			PrimaryBackground: "#007bff",
			PrimaryColor:      "#ffffff",
		},
		
		Layout: &LayoutTheme{
			Spacing: map[string]string{"gap": "16px"},
		},
		
		Colors: map[string]string{
			"primary":   "#007bff",
			"secondary": "#6c757d",
		},
		
		Fonts: map[string]string{
			"body":    "Inter, sans-serif",
			"heading": "Roboto, sans-serif",
		},
		
		DarkMode: &DarkModeConfig{
			Enabled: true,
			Default: true,
		},
		
		Accessibility: &AccessibilityConfig{
			HighContrast:     true,
			FocusIndicator:   true,
			ScreenReaderOnly: true,
		},
	}
	
	suite.Require().Equal("Complete Theme", theme.Name)
	suite.Require().NotNil(theme.Form)
	suite.Require().NotNil(theme.Field)
	suite.Require().NotNil(theme.Button)
	suite.Require().NotNil(theme.Layout)
	suite.Require().NotNil(theme.DarkMode)
	suite.Require().NotNil(theme.Accessibility)
	suite.Require().Equal("#007bff", theme.Colors["primary"])
	suite.Require().Equal("Inter, sans-serif", theme.Fonts["body"])
}