// Test file for CSS runtime system
package css

import (
	"fmt"
	"strings"
	"testing"
)

// TestCSSGenerator tests basic CSS generation functionality
func TestCSSGenerator(t *testing.T) {
	generator := NewCSSGenerator()

	// Test adding a simple rule
	generator.AddRule(".button", map[string]string{
		"background-color": "#007bff",
		"color":            "#ffffff",
		"padding":          "8px 16px",
		"border-radius":    "4px",
	})

	css := generator.Generate()

	// Check that CSS contains expected content
	if !strings.Contains(css, ".ui-button") {
		t.Error("Generated CSS should contain prefixed selector")
	}

	if !strings.Contains(css, "background-color: #007bff") {
		t.Error("Generated CSS should contain background-color declaration")
	}

	if !strings.Contains(css, "color: #ffffff") {
		t.Error("Generated CSS should contain color declaration")
	}
}

// TestCSSGeneratorWithMedia tests media query generation
func TestCSSGeneratorWithMedia(t *testing.T) {
	generator := NewCSSGenerator()

	// Add rule with media query
	generator.AddRuleWithMedia(".container", map[string]string{
		"max-width": "1200px",
	}, "(min-width: 1024px)")

	css := generator.Generate()

	if !strings.Contains(css, "@media (min-width: 1024px)") {
		t.Error("Generated CSS should contain media query")
	}

	if !strings.Contains(css, "max-width: 1200px") {
		t.Error("Generated CSS should contain max-width declaration")
	}
}

// TestCSSGeneratorWithPseudo tests pseudo-selector generation
func TestCSSGeneratorWithPseudo(t *testing.T) {
	generator := NewCSSGenerator()

	// Add rule with pseudo-selector
	generator.AddRuleWithPseudo(".button", ":hover", map[string]string{
		"background-color": "#0056b3",
	})

	css := generator.Generate()

	if !strings.Contains(css, ".ui-button:hover") {
		t.Error("Generated CSS should contain pseudo-selector")
	}

	if !strings.Contains(css, "background-color: #0056b3") {
		t.Error("Generated CSS should contain hover background-color")
	}
}

// TestCSSGeneratorVariables tests CSS custom properties
func TestCSSGeneratorVariables(t *testing.T) {
	generator := NewCSSGenerator()

	// Add CSS variables
	generator.AddVariable("primary-color", "#007bff")
	generator.AddVariable("secondary-color", "#6c757d")

	css := generator.Generate()

	if !strings.Contains(css, ":root {") {
		t.Error("Generated CSS should contain :root selector for variables")
	}

	if !strings.Contains(css, "--primary-color: #007bff") {
		t.Error("Generated CSS should contain primary color variable")
	}

	if !strings.Contains(css, "--secondary-color: #6c757d") {
		t.Error("Generated CSS should contain secondary color variable")
	}
}

// TestCSSGeneratorKeyframes tests keyframe animation generation
func TestCSSGeneratorKeyframes(t *testing.T) {
	generator := NewCSSGenerator()

	// Add keyframes
	keyframes := []CSSKeyframe{
		{Position: "0%", Declarations: map[string]string{"opacity": "0", "transform": "translateY(-10px)"}},
		{Position: "100%", Declarations: map[string]string{"opacity": "1", "transform": "translateY(0)"}},
	}

	generator.AddKeyframes("fadeInUp", keyframes)

	css := generator.Generate()

	if !strings.Contains(css, "@keyframes fadeInUp") {
		t.Error("Generated CSS should contain keyframes declaration")
	}

	if !strings.Contains(css, "0% {") {
		t.Error("Generated CSS should contain 0% keyframe")
	}

	if !strings.Contains(css, "opacity: 0") {
		t.Error("Generated CSS should contain opacity declaration in keyframe")
	}
}

// TestCSSGeneratorMinified tests minified output
func TestCSSGeneratorMinified(t *testing.T) {
	config := &CSSConfig{
		Minify:      true,
		ClassPrefix: "",
	}
	generator := NewCSSGeneratorWithConfig(config)

	generator.AddRule(".button", map[string]string{
		"background-color": "#007bff",
		"color":            "#ffffff",
	})

	css := generator.Generate()

	// Minified CSS should not contain extra whitespace
	if strings.Contains(css, "\n  ") {
		t.Error("Minified CSS should not contain indentation")
	}

	// Should still contain the declarations
	if !strings.Contains(css, "background-color:#007bff") {
		t.Error("Minified CSS should contain declarations without spaces")
	}
}

// TestCSSGeneratorRemoveDuplicates tests duplicate rule removal
func TestCSSGeneratorRemoveDuplicates(t *testing.T) {
	config := &CSSConfig{
		RemoveDuplicates: true,
		ClassPrefix:      "",
	}
	generator := NewCSSGeneratorWithConfig(config)

	// Add same rule twice
	generator.AddRule(".button", map[string]string{"color": "red"})
	generator.AddRule(".button", map[string]string{"color": "blue"})

	css := generator.Generate()

	// Should only contain one .button rule
	buttonCount := strings.Count(css, ".button {")
	if buttonCount != 1 {
		t.Errorf("Expected 1 .button rule, got %d", buttonCount)
	}
}

// TestComponentCSSBuilder tests component-specific CSS generation
func TestComponentCSSBuilder(t *testing.T) {
	builder := NewComponentCSSBuilder()

	// Generate button CSS
	builder.GenerateButtonCSS(ComponentCSSOptions{
		ComponentType: "button",
		Size:          "md",
		Variant:       "primary",
	})

	css := builder.Generate()

	// Check that button CSS is generated
	if !strings.Contains(css, ".ui-btn-primary-md") {
		t.Error("Generated CSS should contain button class")
	}

	// Check for expected button properties
	if !strings.Contains(css, "display: inline-flex") {
		t.Error("Button should have inline-flex display")
	}

	if !strings.Contains(css, "cursor: pointer") {
		t.Error("Button should have pointer cursor")
	}
}

// TestComponentCSSBuilderInput tests input CSS generation
func TestComponentCSSBuilderInput(t *testing.T) {
	builder := NewComponentCSSBuilder()

	// Generate input CSS
	builder.GenerateInputCSS(ComponentCSSOptions{
		ComponentType: "input",
		Size:          "md",
	})

	css := builder.Generate()

	// Check that input CSS is generated
	if !strings.Contains(css, ".ui-input-md") {
		t.Error("Generated CSS should contain input class")
	}

	// Check for expected input properties
	if !strings.Contains(css, "display: block") {
		t.Error("Input should have block display")
	}

	if !strings.Contains(css, "width: 100%") {
		t.Error("Input should have full width")
	}

	// Check for focus state
	if !strings.Contains(css, ".ui-input-md:focus") {
		t.Error("Input should have focus state")
	}
}

// TestComponentCSSBuilderCard tests card CSS generation
func TestComponentCSSBuilderCard(t *testing.T) {
	builder := NewComponentCSSBuilder()

	// Generate card CSS
	builder.GenerateCardCSS(ComponentCSSOptions{
		ComponentType: "card",
		Variant:       "elevated",
	})

	css := builder.Generate()

	// Check that card CSS is generated
	if !strings.Contains(css, ".ui-card-elevated") {
		t.Error("Generated CSS should contain card class")
	}

	// Check for card-specific properties
	if !strings.Contains(css, "box-shadow:") {
		t.Error("Elevated card should have box-shadow")
	}

	// Check for card sub-components
	if !strings.Contains(css, ".ui-card-elevated .card-header") {
		t.Error("Card should have header styles")
	}

	if !strings.Contains(css, ".ui-card-elevated .card-body") {
		t.Error("Card should have body styles")
	}
}

// TestComponentCSSBuilderUtilities tests utility class generation
func TestComponentCSSBuilderUtilities(t *testing.T) {
	builder := NewComponentCSSBuilder()

	// Generate utility CSS
	builder.GenerateUtilityCSS()

	css := builder.Generate()

	// Check for spacing utilities
	if !strings.Contains(css, ".ui-m-md") {
		t.Error("Should generate margin utilities")
	}

	if !strings.Contains(css, ".ui-p-lg") {
		t.Error("Should generate padding utilities")
	}

	// Check for text utilities
	if !strings.Contains(css, ".ui-text-md") {
		t.Error("Should generate text size utilities")
	}

	// Check for color utilities
	if !strings.Contains(css, ".ui-text-primary-500") {
		t.Error("Should generate text color utilities")
	}

	if !strings.Contains(css, ".ui-bg-primary-500") {
		t.Error("Should generate background color utilities")
	}
}

// TestComponentCSSBuilderThemeVariables tests theme variable generation
func TestComponentCSSBuilderThemeVariables(t *testing.T) {
	builder := NewComponentCSSBuilder()

	// Add theme variables
	builder.AddThemeVariables()

	css := builder.Generate()

	// Check for color variables
	if !strings.Contains(css, "--color-primary-500") {
		t.Error("Should generate color variables")
	}

	// Check for spacing variables
	if !strings.Contains(css, "--spacing-md") {
		t.Error("Should generate spacing variables")
	}

	// Check for typography variables
	if !strings.Contains(css, "--font-size-md") {
		t.Error("Should generate font size variables")
	}
}

// TestComponentCSSBuilderResponsive tests responsive CSS generation
func TestComponentCSSBuilderResponsive(t *testing.T) {
	builder := NewComponentCSSBuilder()

	// Generate responsive CSS
	responsiveStyles := map[string]map[string]string{
		"md": {
			"padding": "2rem",
		},
		"lg": {
			"padding": "3rem",
		},
	}

	builder.GenerateResponsiveCSS(".container", responsiveStyles)

	css := builder.Generate()

	// Check for media queries
	if !strings.Contains(css, "@media (min-width: 768px)") {
		t.Error("Should generate medium breakpoint media query")
	}

	if !strings.Contains(css, "@media (min-width: 1024px)") {
		t.Error("Should generate large breakpoint media query")
	}

	// Check for responsive styles
	if strings.Count(css, "padding: 2rem") < 1 {
		t.Error("Should contain medium responsive padding")
	}

	if strings.Count(css, "padding: 3rem") < 1 {
		t.Error("Should contain large responsive padding")
	}
}

// TestERPTheme tests the default ERP theme configuration
func TestERPTheme(t *testing.T) {
	theme := DefaultERPTheme()

	// Test color palette
	if theme.Colors["primary-500"] == "" {
		t.Error("Theme should have primary color")
	}

	if theme.Colors["success-500"] == "" {
		t.Error("Theme should have success color")
	}

	if theme.Colors["error-500"] == "" {
		t.Error("Theme should have error color")
	}

	// Test spacing scale
	if len(theme.Spacing) < 5 {
		t.Error("Theme should have comprehensive spacing scale")
	}

	// Test typography scale
	if len(theme.Typography) < 4 {
		t.Error("Theme should have comprehensive typography scale")
	}

	// Test shadows
	if theme.Shadows["md"] == "" {
		t.Error("Theme should have medium shadow")
	}

	// Test border radius
	if theme.BorderRadius["md"] == "" {
		t.Error("Theme should have medium border radius")
	}

	// Test breakpoints
	if len(theme.Breakpoints) < 3 {
		t.Error("Theme should have responsive breakpoints")
	}

	// Test animations
	if theme.Animations["fast"].Duration == "" {
		t.Error("Theme should have animation configurations")
	}
}

// BenchmarkCSSGeneration benchmarks CSS generation performance
func BenchmarkCSSGeneration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		generator := NewCSSGenerator()

		// Add multiple rules
		for j := 0; j < 100; j++ {
			generator.AddRule(fmt.Sprintf(".class-%d", j), map[string]string{
				"color":      "#000000",
				"background": "#ffffff",
				"padding":    "1rem",
				"margin":     "0.5rem",
			})
		}

		_ = generator.Generate()
	}
}

// BenchmarkComponentCSSGeneration benchmarks component CSS generation
func BenchmarkComponentCSSGeneration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		builder := NewComponentCSSBuilder()

		// Generate various component styles
		builder.GenerateButtonCSS(ComponentCSSOptions{Size: "md", Variant: "primary"})
		builder.GenerateInputCSS(ComponentCSSOptions{Size: "md"})
		builder.GenerateCardCSS(ComponentCSSOptions{Variant: "elevated"})
		builder.GenerateUtilityCSS()
		builder.AddThemeVariables()

		_ = builder.Generate()
	}
}
