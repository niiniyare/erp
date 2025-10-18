// Package css provides theme configuration for atomic components
package css

// Theme contains design system values for consistent styling
type Theme struct {
	// Color palette
	Colors map[string]string
	// Spacing scale (margins, padding)
	Spacing map[string]string
	// Typography scale
	Typography map[string]TypographyConfig
	// Shadow levels
	Shadows map[string]string
	// Border radius values
	BorderRadius map[string]string
	// Z-index scale
	ZIndex map[string]int
	// Animation durations and easings
	Animations map[string]AnimationConfig
}

// TypographyConfig represents typography settings
type TypographyConfig struct {
	FontSize      string
	LineHeight    string
	FontWeight    string
	LetterSpacing string
}

// AnimationConfig represents animation settings
type AnimationConfig struct {
	Duration string
	Easing   string
	Delay    string
}

// DefaultTheme provides a professional ERP design system theme
func DefaultTheme() *Theme {
	return &Theme{
		Colors: map[string]string{
			// Primary colors
			"primary":       "#3b82f6", // Blue 500
			"primary-light": "#dbeafe", // Blue 100
			"primary-dark":  "#1e40af", // Blue 700

			// Secondary colors
			"secondary":       "#64748b", // Slate 500
			"secondary-light": "#f1f5f9", // Slate 100
			"secondary-dark":  "#334155", // Slate 700

			// Success colors
			"success":       "#22c55e", // Green 500
			"success-light": "#dcfce7", // Green 100
			"success-dark":  "#15803d", // Green 700

			// Warning colors
			"warning":       "#f59e0b", // Amber 500
			"warning-light": "#fef3c7", // Amber 100
			"warning-dark":  "#b45309", // Amber 700

			// Error colors
			"error":       "#ef4444", // Red 500
			"error-light": "#fee2e2", // Red 100
			"error-dark":  "#b91c1c", // Red 700

			// Neutral colors
			"neutral":         "#6b7280", // Gray 500
			"neutral-light":   "#e5e7eb", // Gray 200
			"neutral-lighter": "#f9fafb", // Gray 50
			"neutral-dark":    "#374151", // Gray 700
			"neutral-darker":  "#1f2937", // Gray 800

			// Base colors
			"white": "#ffffff",
			"black": "#000000",
		},
		Spacing: map[string]string{
			"xs":  "0.25rem", // 4px
			"sm":  "0.5rem",  // 8px
			"md":  "1rem",    // 16px
			"lg":  "1.5rem",  // 24px
			"xl":  "2rem",    // 32px
			"2xl": "2.5rem",  // 40px
			"3xl": "3rem",    // 48px
		},
		Typography: map[string]TypographyConfig{
			"xs": {
				FontSize:      "0.75rem",
				LineHeight:    "1rem",
				FontWeight:    "400",
				LetterSpacing: "0",
			},
			"sm": {
				FontSize:      "0.875rem",
				LineHeight:    "1.25rem",
				FontWeight:    "400",
				LetterSpacing: "0",
			},
			"md": {
				FontSize:      "1rem",
				LineHeight:    "1.5rem",
				FontWeight:    "400",
				LetterSpacing: "0",
			},
			"lg": {
				FontSize:      "1.125rem",
				LineHeight:    "1.75rem",
				FontWeight:    "400",
				LetterSpacing: "0",
			},
			"xl": {
				FontSize:      "1.25rem",
				LineHeight:    "1.75rem",
				FontWeight:    "500",
				LetterSpacing: "0",
			},
			"2xl": {
				FontSize:      "1.5rem",
				LineHeight:    "2rem",
				FontWeight:    "600",
				LetterSpacing: "0",
			},
		},
		Shadows: map[string]string{
			"none": "none",
			"sm":   "0 1px 2px 0 rgb(0 0 0 / 0.05)",
			"md":   "0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1)",
			"lg":   "0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1)",
			"xl":   "0 20px 25px -5px rgb(0 0 0 / 0.1), 0 8px 10px -6px rgb(0 0 0 / 0.1)",
		},
		BorderRadius: map[string]string{
			"none": "0",
			"sm":   "0.125rem", // 2px
			"md":   "0.375rem", // 6px
			"lg":   "0.5rem",   // 8px
			"xl":   "0.75rem",  // 12px
			"full": "9999px",
		},
		ZIndex: map[string]int{
			"dropdown": 1000,
			"modal":    1050,
			"tooltip":  1100,
			"toast":    1200,
		},
		Animations: map[string]AnimationConfig{
			"fast": {
				Duration: "150ms",
				Easing:   "cubic-bezier(0.4, 0, 0.2, 1)",
				Delay:    "0ms",
			},
			"normal": {
				Duration: "300ms",
				Easing:   "cubic-bezier(0.4, 0, 0.2, 1)",
				Delay:    "0ms",
			},
			"slow": {
				Duration: "500ms",
				Easing:   "cubic-bezier(0.4, 0, 0.2, 1)",
				Delay:    "0ms",
			},
		},
	}
}

// GetColor safely retrieves a color from the theme
func (t *Theme) GetColor(name string) string {
	if color, exists := t.Colors[name]; exists {
		return color
	}
	// Fallback to neutral if color not found
	return t.Colors["neutral"]
}

// GetSpacing safely retrieves spacing from the theme
func (t *Theme) GetSpacing(name string) string {
	if spacing, exists := t.Spacing[name]; exists {
		return spacing
	}
	// Fallback to medium spacing
	return t.Spacing["md"]
}

// GetTypography safely retrieves typography config from the theme
func (t *Theme) GetTypography(name string) TypographyConfig {
	if typography, exists := t.Typography[name]; exists {
		return typography
	}
	// Fallback to medium typography
	return t.Typography["md"]
}

// GetShadow safely retrieves shadow from the theme
func (t *Theme) GetShadow(name string) string {
	if shadow, exists := t.Shadows[name]; exists {
		return shadow
	}
	// Fallback to no shadow
	return t.Shadows["none"]
}

// GetBorderRadius safely retrieves border radius from the theme
func (t *Theme) GetBorderRadius(name string) string {
	if radius, exists := t.BorderRadius[name]; exists {
		return radius
	}
	// Fallback to medium radius
	return t.BorderRadius["md"]
}

// GetAnimation safely retrieves animation config from the theme
func (t *Theme) GetAnimation(name string) AnimationConfig {
	if animation, exists := t.Animations[name]; exists {
		return animation
	}
	// Fallback to normal animation
	return t.Animations["normal"]
}
