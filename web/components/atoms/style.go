package atoms

import "fmt"

// ============================================================================
// SPACING CONSTANTS
// ============================================================================

const (
	// Wrapper/Container spacing
	WrapperSpacingDefault = "mb-4"
	WrapperSpacingTight   = "mb-2"
	WrapperSpacingLoose   = "mb-6"

	// Label spacing
	LabelSpacingBottom = "mb-2"
	LabelSpacingRight  = "ms-2"
	LabelSpacingLeft   = "me-2"

	// Feedback message spacing
	FeedbackSpacingTop = "mt-2"

	// Group spacing
	GroupSpacing       = "mb-5"
	GroupItemSpacing   = "space-y-2"
	GroupLegendSpacing = "mb-3"

	// Icon spacing
	IconSpacingLeft  = "me-2"
	IconSpacingRight = "ms-2"
)

// ============================================================================
// BORDER RADIUS CONSTANTS
// ============================================================================

const (
	RoundedNone = "rounded-none"
	RoundedSM   = "rounded-sm"
	RoundedMD   = "rounded-md"
	RoundedLG   = "rounded-lg"
	RoundedXL   = "rounded-xl"
	RoundedFull = "rounded-full"
)

// GetRoundedClass returns appropriate rounded class based on preferences.
func GetRoundedClass(rounded bool, size Size) string {
	if !rounded {
		return RoundedSM
	}

	// Scale rounding with size
	switch size {
	case SizeXS, SizeSM:
		return RoundedSM
	case SizeLG, SizeXL:
		return RoundedLG
	default:
		return RoundedMD
	}
}

// ============================================================================
// TRANSITION CONSTANTS
// ============================================================================

const (
	TransitionColors    = "transition-colors duration-200"
	TransitionAll       = "transition-all duration-200"
	TransitionOpacity   = "transition-opacity duration-200"
	TransitionTransform = "transition-transform duration-200"
	TransitionFast      = "transition-all duration-150"
	TransitionSlow      = "transition-all duration-300"
)

// ============================================================================
// THEME SYSTEM
// ============================================================================

// Theme contains design system values for consistent styling.
type Theme struct {
	Name         string
	Colors       map[string]string
	Spacing      map[string]string
	Typography   map[string]TypographyConfig
	Shadows      map[string]string
	BorderRadius map[string]string
	ZIndex       map[string]int
	Animations   map[string]AnimationConfig
}

// TypographyConfig represents typography settings.
type TypographyConfig struct {
	FontSize      string
	LineHeight    string
	FontWeight    string
	LetterSpacing string
}

// AnimationConfig represents animation settings.
type AnimationConfig struct {
	Duration string
	Easing   string
	Delay    string
}

// DefaultFallbackColor is the ultimate fallback when theme colors are missing.
const DefaultFallbackColor = "#6b7280"

// DefaultTheme provides a professional ERP design system theme.
func DefaultTheme() *Theme {
	return &Theme{
		Name: "default",
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

			// Info colors
			"info":       "#3b82f6", // Blue 500
			"info-light": "#dbeafe", // Blue 100
			"info-dark":  "#1e40af", // Blue 700

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

// GetColor safely retrieves a color from the theme.
func (t *Theme) GetColor(name string) string {
	if color, exists := t.Colors[name]; exists {
		return color
	}
	// Try neutral fallback
	if neutral, exists := t.Colors["neutral"]; exists {
		return neutral
	}
	// Ultimate fallback
	return DefaultFallbackColor
}

// GetSpacing safely retrieves spacing from the theme.
func (t *Theme) GetSpacing(name string) string {
	if spacing, exists := t.Spacing[name]; exists {
		return spacing
	}
	// Fallback to medium spacing
	if md, exists := t.Spacing["md"]; exists {
		return md
	}
	return "1rem"
}

// GetTypography safely retrieves typography config from the theme.
func (t *Theme) GetTypography(name string) TypographyConfig {
	if typography, exists := t.Typography[name]; exists {
		return typography
	}
	// Fallback to medium typography
	if md, exists := t.Typography["md"]; exists {
		return md
	}
	return TypographyConfig{
		FontSize:   "1rem",
		LineHeight: "1.5rem",
		FontWeight: "400",
	}
}

// GetShadow safely retrieves shadow from the theme.
func (t *Theme) GetShadow(name string) string {
	if shadow, exists := t.Shadows[name]; exists {
		return shadow
	}
	return "none"
}

// GetBorderRadius safely retrieves border radius from the theme.
func (t *Theme) GetBorderRadius(name string) string {
	if radius, exists := t.BorderRadius[name]; exists {
		return radius
	}
	if md, exists := t.BorderRadius["md"]; exists {
		return md
	}
	return "0.375rem"
}

// GetAnimation safely retrieves animation config from the theme.
func (t *Theme) GetAnimation(name string) AnimationConfig {
	if animation, exists := t.Animations[name]; exists {
		return animation
	}
	if normal, exists := t.Animations["normal"]; exists {
		return normal
	}
	return AnimationConfig{
		Duration: "300ms",
		Easing:   "cubic-bezier(0.4, 0, 0.2, 1)",
	}
}

// GetZIndex safely retrieves z-index from the theme.
func (t *Theme) GetZIndex(name string) int {
	if zIndex, exists := t.ZIndex[name]; exists {
		return zIndex
	}
	return 1000
}

// ============================================================================
// STATE COLOR CONFIGURATION
// ============================================================================

// StateColorConfig holds all color classes for a specific validation state.
type StateColorConfig struct {
	Border     []string // Border colors
	Background []string // Background colors
	Text       []string // Text colors
	Ring       []string // Focus ring colors
	Icon       []string // Icon colors (for feedback indicators)
}

// GetStateColors returns the color configuration for a validation state.
func GetStateColors(state ValidationState) StateColorConfig {
	configs := map[ValidationState]StateColorConfig{
		StateDefault: {
			Border:     []string{"border-gray-300", "dark:border-gray-600"},
			Background: []string{"bg-white", "dark:bg-gray-800"},
			Text:       []string{"text-gray-900", "dark:text-white"},
			Ring:       []string{"focus:ring-blue-500", "focus:border-blue-500"},
			Icon:       []string{"text-gray-500", "dark:text-gray-400"},
		},
		StateSuccess: {
			Border:     []string{"border-green-500", "dark:border-green-500"},
			Background: []string{"bg-green-50", "dark:bg-gray-800"},
			Text:       []string{"text-green-900", "dark:text-green-400"},
			Ring:       []string{"focus:ring-green-500", "focus:border-green-500"},
			Icon:       []string{"text-green-500", "dark:text-green-400"},
		},
		StateError: {
			Border:     []string{"border-red-500", "dark:border-red-500"},
			Background: []string{"bg-red-50", "dark:bg-gray-800"},
			Text:       []string{"text-red-900", "dark:text-red-400"},
			Ring:       []string{"focus:ring-red-500", "focus:border-red-500"},
			Icon:       []string{"text-red-500", "dark:text-red-400"},
		},
		StateWarning: {
			Border:     []string{"border-yellow-500", "dark:border-yellow-500"},
			Background: []string{"bg-yellow-50", "dark:bg-gray-800"},
			Text:       []string{"text-yellow-900", "dark:text-yellow-400"},
			Ring:       []string{"focus:ring-yellow-500", "focus:border-yellow-500"},
			Icon:       []string{"text-yellow-500", "dark:text-yellow-400"},
		},
		StateInfo: {
			Border:     []string{"border-blue-500", "dark:border-blue-500"},
			Background: []string{"bg-blue-50", "dark:bg-gray-800"},
			Text:       []string{"text-blue-900", "dark:text-blue-400"},
			Ring:       []string{"focus:ring-blue-500", "focus:border-blue-500"},
			Icon:       []string{"text-blue-500", "dark:text-blue-400"},
		},
	}

	if config, ok := configs[state]; ok {
		return config
	}
	return configs[StateDefault]
}

// GetFeedbackColors returns text color classes for feedback messages.
func GetFeedbackColors(state ValidationState) []string {
	colorMap := map[ValidationState][]string{
		StateDefault: {"text-gray-600", "dark:text-gray-400"},
		StateSuccess: {"text-green-600", "dark:text-green-500"},
		StateError:   {"text-red-600", "dark:text-red-500"},
		StateWarning: {"text-yellow-600", "dark:text-yellow-500"},
		StateInfo:    {"text-blue-600", "dark:text-blue-500"},
	}

	if colors, ok := colorMap[state]; ok {
		return colors
	}
	return colorMap[StateDefault]
}

// GetLabelColorClasses returns label color classes based on validation state.
func GetLabelColorClasses(state ValidationState) []string {
	colorMap := map[ValidationState][]string{
		StateDefault: {"text-gray-900", "dark:text-gray-300"},
		StateError:   {"text-red-900", "dark:text-red-400"},
		StateSuccess: {"text-green-900", "dark:text-green-400"},
		StateWarning: {"text-yellow-900", "dark:text-yellow-400"},
		StateInfo:    {"text-blue-900", "dark:text-blue-400"},
	}

	if colors, ok := colorMap[state]; ok {
		return colors
	}
	return colorMap[StateDefault]
}

// GetFeedbackClasses returns classes for validation feedback messages.
func GetFeedbackClasses(state ValidationState) string {
	classes := []string{
		FeedbackSpacingTop,
		"text-sm",
	}
	classes = append(classes, GetFeedbackColors(state)...)
	return JoinClasses(classes...)
}

// GetValidationStateClasses returns validation state classes for form inputs.
func GetValidationStateClasses(state ValidationState) []string {
	switch state {
	case StateError:
		return []string{
			"border-red-500", "dark:border-red-500",
			"focus:ring-red-500", "focus:border-red-500",
		}
	case StateSuccess:
		return []string{
			"border-green-500", "dark:border-green-500",
			"focus:ring-green-500", "focus:border-green-500",
		}
	case StateWarning:
		return []string{
			"border-yellow-500", "dark:border-yellow-500",
			"focus:ring-yellow-500", "focus:border-yellow-500",
		}
	case StateInfo:
		return []string{
			"border-blue-500", "dark:border-blue-500",
			"focus:ring-blue-500", "focus:border-blue-500",
		}
	default:
		return []string{
			"border-gray-300", "dark:border-gray-600",
			"focus:ring-blue-500", "focus:border-blue-500",
		}
	}
}

// ============================================================================
// SIZE-SPECIFIC CLASS MAPPINGS
// ============================================================================

// GetInputSizeClasses returns size classes for text inputs.
func GetInputSizeClasses(size Size) []string {
	switch size {
	case SizeXS:
		return []string{"text-xs", "py-1", "px-2", "h-7"}
	case SizeSM:
		return []string{"text-sm", "py-1.5", "px-3", "h-8"}
	case SizeLG:
		return []string{"text-base", "py-3", "px-4", "h-12"}
	case SizeXL:
		return []string{"text-lg", "py-4", "px-5", "h-14"}
	default: // SizeMD
		return []string{"text-sm", "py-2.5", "px-4", "h-10"}
	}
}

// GetButtonSizeClasses returns size classes for buttons.
func GetButtonSizeClasses(size Size) []string {
	switch size {
	case SizeXS:
		return []string{"text-xs", "py-1", "px-2", "h-7"}
	case SizeSM:
		return []string{"text-sm", "py-2", "px-3", "h-8"}
	case SizeLG:
		return []string{"text-base", "py-3", "px-5", "h-12"}
	case SizeXL:
		return []string{"text-lg", "py-3.5", "px-6", "h-14"}
	default: // SizeMD
		return []string{"text-sm", "py-2.5", "px-5", "h-10"}
	}
}

// GetCheckboxSizeClasses returns dimension classes for checkboxes/radios.
func GetCheckboxSizeClasses(size Size) []string {
	switch size {
	case SizeXS:
		return []string{"w-3", "h-3"}
	case SizeSM:
		return []string{"w-3.5", "h-3.5"}
	case SizeLG:
		return []string{"w-5", "h-5"}
	case SizeXL:
		return []string{"w-6", "h-6"}
	default: // SizeMD
		return []string{"w-4", "h-4"}
	}
}

// GetIconSizeClasses returns dimension classes for icons.
func GetIconSizeClasses(size Size) []string {
	switch size {
	case SizeXS:
		return []string{"w-3", "h-3"}
	case SizeSM:
		return []string{"w-4", "h-4"}
	case SizeLG:
		return []string{"w-6", "h-6"}
	case SizeXL:
		return []string{"w-8", "h-8"}
	default: // SizeMD
		return []string{"w-5", "h-5"}
	}
}

// GetImageSizeClasses returns dimension classes for images.
func GetImageSizeClasses(size Size) []string {
	switch size {
	case SizeXS:
		return []string{"w-8", "h-8"}
	case SizeSM:
		return []string{"w-12", "h-12"}
	case SizeLG:
		return []string{"w-24", "h-24"}
	case SizeXL:
		return []string{"w-32", "h-32"}
	default: // SizeMD
		return []string{"w-16", "h-16"}
	}
}

// ============================================================================
// VARIANT STYLE MAPPINGS
// ============================================================================

// GetButtonVariantClasses returns Tailwind classes for button variants.
func GetButtonVariantClasses(variant Variant, colorScheme ColorScheme) []string {
	// Map color schemes to Tailwind colors
	colorMap := map[ColorScheme]string{
		ColorPrimary:   "blue",
		ColorSecondary: "gray",
		ColorSuccess:   "green",
		ColorDanger:    "red",
		ColorWarning:   "amber",
		ColorInfo:      "blue",
		ColorDefault:   "blue",
	}

	color := colorMap[colorScheme]
	if color == "" {
		color = "blue"
	}

	switch variant {
	case VariantSolid, VariantDefault:
		return []string{
			fmt.Sprintf("bg-%s-500", color),
			"text-white",
			fmt.Sprintf("hover:bg-%s-600", color),
			fmt.Sprintf("active:bg-%s-700", color),
			fmt.Sprintf("focus:ring-%s-500", color),
		}
	case VariantOutlined:
		return []string{
			"bg-transparent",
			fmt.Sprintf("border-%s-500", color),
			fmt.Sprintf("text-%s-600", color),
			fmt.Sprintf("hover:bg-%s-50", color),
			"border",
		}
	case VariantGhost:
		return []string{
			"bg-transparent",
			fmt.Sprintf("text-%s-600", color),
			fmt.Sprintf("hover:bg-%s-50", color),
		}
	case VariantFilled:
		return []string{
			fmt.Sprintf("bg-%s-100", color),
			fmt.Sprintf("text-%s-700", color),
			fmt.Sprintf("hover:bg-%s-200", color),
		}
	default:
		return GetButtonVariantClasses(VariantSolid, colorScheme)
	}
}

// GetInputVariantClasses returns Tailwind classes for input variants.
func GetInputVariantClasses(variant Variant, state ValidationState) []string {
	stateColors := GetStateColors(state)

	baseClasses := []string{
		"w-full",
		"focus:outline-none",
		TransitionColors,
	}

	switch variant {
	case VariantOutlined, VariantDefault:
		classes := append(baseClasses, "border", "bg-white", "dark:bg-gray-800")
		classes = append(classes, stateColors.Border...)
		classes = append(classes, stateColors.Ring...)
		return classes

	case VariantFilled:
		classes := append(baseClasses, "border-0")
		classes = append(classes, stateColors.Background...)
		return classes

	case VariantUnderlined:
		return append(baseClasses,
			"border-0",
			"border-b-2",
			"bg-transparent",
			"rounded-none",
			"px-0",
		)

	default:
		return GetInputVariantClasses(VariantOutlined, state)
	}
}

// ============================================================================
// COMPONENT BASE CLASSES
// ============================================================================

// GetBaseInputClasses returns common input classes.
func GetBaseInputClasses() []string {
	return []string{
		"block",
		"w-full",
		"rounded-md",
		"focus:outline-none",
		"focus:ring-2",
		"focus:ring-offset-0",
		TransitionColors,
		"disabled:cursor-not-allowed",
		"disabled:opacity-50",
	}
}

// GetBaseButtonClasses returns common button classes.
func GetBaseButtonClasses() []string {
	return []string{
		"inline-flex",
		"items-center",
		"justify-center",
		"font-medium",
		"rounded-md",
		"focus:outline-none",
		"focus:ring-2",
		"focus:ring-offset-2",
		TransitionColors,
		"disabled:cursor-not-allowed",
		"disabled:opacity-50",
		"cursor-pointer",
	}
}

// GetBaseCheckboxClasses returns common checkbox/radio classes.
func GetBaseCheckboxClasses() []string {
	return []string{
		"rounded",
		"border-gray-300",
		"text-blue-600",
		"focus:ring-blue-500",
		"focus:ring-2",
		"focus:ring-offset-0",
		TransitionColors,
		"disabled:cursor-not-allowed",
		"disabled:opacity-50",
		"cursor-pointer",
	}
}

// ============================================================================
// UTILITY STYLE FUNCTIONS
// ============================================================================

// GetDisabledClasses returns classes for disabled state.
func GetDisabledClasses() []string {
	return []string{
		"opacity-50",
		"cursor-not-allowed",
		"pointer-events-none",
	}
}

// GetLoadingClasses returns classes for loading state.
func GetLoadingClasses() []string {
	return []string{
		"opacity-75",
		"cursor-wait",
		"pointer-events-none",
	}
}

// GetFocusClasses returns common focus ring classes.
func GetFocusClasses(color ColorScheme) []string {
	colorMap := map[ColorScheme]string{
		ColorPrimary:   "blue",
		ColorSecondary: "gray",
		ColorSuccess:   "green",
		ColorDanger:    "red",
		ColorWarning:   "amber",
		ColorInfo:      "blue",
		ColorDefault:   "blue",
	}

	c := colorMap[color]
	if c == "" {
		c = "blue"
	}

	return []string{
		"focus:outline-none",
		fmt.Sprintf("focus:ring-2"),
		fmt.Sprintf("focus:ring-%s-500", c),
		"focus:ring-offset-2",
	}
}
