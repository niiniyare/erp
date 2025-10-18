package css

import (
	"testing"
)

func TestStylesCreation(t *testing.T) {
	// Test basic styles creation
	styles := NewStyles("../../../docs/ui/Schema")

	styles.WithDisplay("flex").
		WithJustifyContent("center").
		WithAlignItems("center").
		WithPadding("1rem").
		WithBackgroundColor("#ffffff").
		WithBorderRadius("8px")

	css := styles.ToCSS()

	// Check that all expected properties are present (order-independent)
	expectedProperties := []string{
		"display: flex",
		"justify-content: center",
		"align-items: center",
		"padding: 1rem",
		"background-color: #ffffff",
		"border-radius: 8px",
	}

	for _, prop := range expectedProperties {
		if !contains(css, prop) {
			t.Errorf("Expected CSS to contain '%s', got: %s", prop, css)
		}
	}
}

func TestStylesValidation(t *testing.T) {
	validator := NewValidator("../../../docs/ui/Schema")

	// Test valid properties
	tests := []struct {
		property string
		value    string
		valid    bool
	}{
		{"color", "#ffffff", true},
		{"color", "red", true},
		{"color", "rgb(255, 0, 0)", true},
		{"color", "invalid-color", false},
		{"display", "flex", true},
		{"display", "invalid-display", false},
		{"font-size", "16px", true},
		{"font-size", "large", true},
		{"font-weight", "bold", true},
		{"font-weight", "500", true},
		{"font-weight", "invalid", false},
	}

	for _, test := range tests {
		err := validator.ValidateProperty(test.property, test.value)
		if test.valid && err != nil {
			t.Errorf("Expected %s: %s to be valid, got error: %v", test.property, test.value, err)
		}
		if !test.valid && err == nil {
			t.Errorf("Expected %s: %s to be invalid, but it was valid", test.property, test.value)
		}
	}
}

func TestFactoryButtonStyles(t *testing.T) {
	factory := NewFactory("../../../docs/ui/Schema")

	// Test primary button
	primaryStyles := factory.ButtonStyles("primary")
	css := primaryStyles.ToCSS()

	if primaryStyles.Display != "inline-flex" {
		t.Errorf("Expected display: inline-flex, got: %s", primaryStyles.Display)
	}

	if primaryStyles.BackgroundColor != "#3b82f6" {
		t.Errorf("Expected background-color: #3b82f6, got: %s", primaryStyles.BackgroundColor)
	}

	// Test that CSS output contains expected properties
	expectedProps := []string{"display: inline-flex", "background-color: #3b82f6", "color: #ffffff"}
	for _, prop := range expectedProps {
		if !contains(css, prop) {
			t.Errorf("Expected CSS to contain '%s', got: %s", prop, css)
		}
	}
}

func TestFactoryInputStyles(t *testing.T) {
	factory := NewFactory("../../../docs/ui/Schema")

	// Test error state input
	errorStyles := factory.InputStyles("error")

	if errorStyles.Display != "block" {
		t.Errorf("Expected display: block, got: %s", errorStyles.Display)
	}

	if errorStyles.Width != "100%" {
		t.Errorf("Expected width: 100%%, got: %s", errorStyles.Width)
	}

	// Check custom properties
	if errorStyles.Custom["border"] != "1px solid #ef4444" {
		t.Errorf("Expected border custom property, got: %s", errorStyles.Custom["border"])
	}
}

func TestCardStyles(t *testing.T) {
	factory := NewFactory("../../../docs/ui/Schema")

	// Test card with medium elevation
	cardStyles := factory.CardStyles("md")

	if cardStyles.BackgroundColor != "#ffffff" {
		t.Errorf("Expected background-color: #ffffff, got: %s", cardStyles.BackgroundColor)
	}

	if cardStyles.BorderRadius != "0.5rem" {
		t.Errorf("Expected border-radius: 0.5rem, got: %s", cardStyles.BorderRadius)
	}

	expectedShadow := "0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)"
	if cardStyles.BoxShadow != expectedShadow {
		t.Errorf("Expected box-shadow: %s, got: %s", expectedShadow, cardStyles.BoxShadow)
	}
}

func TestResponsiveStyles(t *testing.T) {
	factory := NewFactory("../../../docs/ui/Schema")

	// Test responsive styles builder
	responsiveStyles := factory.ResponsiveStyles().
		Base(factory.FlexStyles("column", "center", "center")).
		MD(func(s *Styles) {
			s.WithFlexDirection("row")
			s.WithJustifyContent("space-between")
		}).
		Build()

	// Check base styles
	if responsiveStyles.Display != "flex" {
		t.Errorf("Expected display: flex, got: %s", responsiveStyles.Display)
	}

	// Check responsive custom properties
	if responsiveStyles.Custom["--md-flex-direction"] != "row" {
		t.Errorf("Expected --md-flex-direction: row, got: %s", responsiveStyles.Custom["--md-flex-direction"])
	}
}

func TestThemeStyles(t *testing.T) {
	factory := NewFactory("../../../docs/ui/Schema")

	// Test light theme
	lightTheme := factory.ThemeStyles("light").
		Primary("#000000", "#ffffff").
		Background("#ffffff", "#1a1a1a").
		Border("#e5e7eb", "#374151").
		Build()

	if lightTheme.Color != "#000000" {
		t.Errorf("Expected light theme color: #000000, got: %s", lightTheme.Color)
	}

	if lightTheme.BackgroundColor != "#ffffff" {
		t.Errorf("Expected light theme background: #ffffff, got: %s", lightTheme.BackgroundColor)
	}

	// Test dark theme
	darkTheme := factory.ThemeStyles("dark").
		Primary("#000000", "#ffffff").
		Background("#ffffff", "#1a1a1a").
		Border("#e5e7eb", "#374151").
		Build()

	if darkTheme.Color != "#ffffff" {
		t.Errorf("Expected dark theme color: #ffffff, got: %s", darkTheme.Color)
	}

	if darkTheme.BackgroundColor != "#1a1a1a" {
		t.Errorf("Expected dark theme background: #1a1a1a, got: %s", darkTheme.BackgroundColor)
	}
}

func TestCSSClassGeneration(t *testing.T) {
	styles := NewStyles("../../../docs/ui/Schema")

	styles.WithDisplay("flex").
		WithJustifyContent("center").
		WithAlignItems("center").
		WithPadding("1rem")

	cssClass := styles.ToCSSClass("test-component")

	// Check that the class contains the expected structure
	expectedStart := ".test-component {"
	expectedEnd := "}"
	expectedProperties := []string{
		"display: flex",
		"justify-content: center",
		"align-items: center",
		"padding: 1rem",
	}

	if !contains(cssClass, expectedStart) {
		t.Errorf("CSS class should start with '%s', got: %s", expectedStart, cssClass)
	}

	if !contains(cssClass, expectedEnd) {
		t.Errorf("CSS class should end with '%s', got: %s", expectedEnd, cssClass)
	}

	for _, prop := range expectedProperties {
		if !contains(cssClass, prop) {
			t.Errorf("CSS class should contain '%s', got: %s", prop, cssClass)
		}
	}
}

func TestPresetStyles(t *testing.T) {
	// Test FlexCenter preset
	flexCenter := FlexCenter()
	if flexCenter.Display != "flex" || flexCenter.JustifyContent != "center" || flexCenter.AlignItems != "center" {
		t.Error("FlexCenter preset not working correctly")
	}

	// Test Card preset
	card := Card()
	if card.Background != "#ffffff" || card.BorderRadius != "8px" {
		t.Error("Card preset not working correctly")
	}

	// Test Button preset
	button := Button()
	if button.Display != "inline-flex" || button.Padding != "0.5rem 1rem" {
		t.Error("Button preset not working correctly")
	}

	// Test Input preset
	input := Input()
	if input.Display != "block" || input.Width != "100%" {
		t.Error("Input preset not working correctly")
	}
}

// Example: Complete component with styling
func TestCompleteStyledComponent(t *testing.T) {
	factory := NewFactory("../../../docs/ui/Schema")

	// Create a styled button component
	buttonStyles := factory.ButtonStyles("primary").
		WithFontSize("1.1rem").
		WithPadding("0.75rem 1.5rem").
		WithCustomProperty("--hover-transform", "translateY(-1px)")

	t.Logf("\nButton CSS: \n")
	t.Log(buttonStyles.ToCSS())

	// Create a styled card
	cardStyles := factory.CardStyles("lg").
		WithPadding("2rem").
		WithMargin("1rem")

	t.Log("\nCard CSS:\n")
	t.Log(cardStyles.ToCSS())

	// Create responsive layout
	layoutStyles := factory.ResponsiveStyles().
		Base(factory.FlexStyles("column", "start", "stretch")).
		MD(func(s *Styles) {
			s.WithFlexDirection("row")
			s.WithJustifyContent("space-between")
		}).
		LG(func(s *Styles) {
			s.WithPadding("2rem")
		}).
		Build()

	t.Logf("\nResponsive Layout CSS:\n")
	t.Log(layoutStyles.ToCSS())

	// t.Log(layoutStyles.ToCSS())

	// Output:
	// Button CSS:
	// display: inline-flex; align-items: center; justify-content: center; padding: 0.75rem 1.5rem; border-radius: 0.375rem; font-weight: 500; transition: all 0.2s ease-in-out; background-color: #3b82f6; color: #ffffff; font-size: 1.1rem; --hover-bg: #2563eb; --hover-transform: translateY(-1px)
	//
	// Card CSS:
	// background-color: #ffffff; border-radius: 0.5rem; box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05); padding: 2rem; margin: 1rem; border: 1px solid #e5e7eb
	//
	// Responsive Layout CSS:
	// display: flex; flex-direction: column; justify-content: start; align-items: stretch; --md-flex-direction: row; --md-justify-content: space-between; --lg-padding: 2rem
}

// Helper function for testing
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[len(s)-len(substr):] == substr ||
		len(s) >= len(substr) && s[:len(substr)] == substr ||
		len(s) > len(substr) && s[len(s)/2-len(substr)/2:len(s)/2+len(substr)/2+len(substr)%2] == substr ||
		stringContains(s, substr)
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
