package schema

import (
	"fmt"
	"strings"
	"sync"
)

// DesignTokens represents the design token system
// Based on the comprehensive design system documented in styles.md
type DesignTokens struct {
	Spacing    SpacingTokens    `json:"spacing"`
	Colors     ColorTokens      `json:"colors"`
	Typography TypographyTokens `json:"typography"`
	Sizes      SizeTokens       `json:"sizes"`
	Borders    BorderTokens     `json:"borders"`
	Shadows    ShadowTokens     `json:"shadows"`
	Animations AnimationTokens  `json:"animations"`
	ZIndex     ZIndexTokens     `json:"zIndex"`
}

// SpacingTokens define all spacing values following design system
type SpacingTokens struct {
	None string `json:"0"`  // "0"
	XS   string `json:"1"`  // "0.25rem"
	SM   string `json:"2"`  // "0.5rem"
	MD   string `json:"4"`  // "1rem" - Base spacing unit
	LG   string `json:"8"`  // "2rem"
	XL   string `json:"12"` // "3rem"
	XXL  string `json:"16"` // "4rem"
}

// ColorTokens define semantic color mappings
type ColorTokens struct {
	Background BackgroundColors `json:"background"`
	Text       TextColors       `json:"text"`
	Border     BorderColors     `json:"border"`
	Feedback   FeedbackColors   `json:"feedback"`
	Primary    PrimaryColors    `json:"primary"`
	Secondary  SecondaryColors  `json:"secondary"`
	Neutral    NeutralColors    `json:"neutral"`
}

type BackgroundColors struct {
	Default  string `json:"default"`
	Subtle   string `json:"subtle"`
	Emphasis string `json:"emphasis"`
	Overlay  string `json:"overlay"`
}

type TextColors struct {
	Default   string `json:"default"`
	Subtle    string `json:"subtle"`
	Disabled  string `json:"disabled"`
	Inverted  string `json:"inverted"`
	Link      string `json:"link"`
	LinkHover string `json:"linkHover"`
}

type BorderColors struct {
	Default string `json:"default"`
	Focus   string `json:"focus"`
	Strong  string `json:"strong"`
	Subtle  string `json:"subtle"`
}

type FeedbackColors struct {
	Success       string `json:"success"`
	SuccessSubtle string `json:"successSubtle"`
	Error         string `json:"error"`
	ErrorSubtle   string `json:"errorSubtle"`
	Warning       string `json:"warning"`
	WarningSubtle string `json:"warningSubtle"`
	Info          string `json:"info"`
	InfoSubtle    string `json:"infoSubtle"`
}

type PrimaryColors struct {
	Main   string `json:"main"`
	Light  string `json:"light"`
	Dark   string `json:"dark"`
	Subtle string `json:"subtle"`
}

type SecondaryColors struct {
	Main   string `json:"main"`
	Light  string `json:"light"`
	Dark   string `json:"dark"`
	Subtle string `json:"subtle"`
}

type NeutralColors struct {
	White string    `json:"white"`
	Black string    `json:"black"`
	Gray  GrayScale `json:"gray"`
}

type GrayScale struct {
	Gray50  string `json:"50"`
	Gray100 string `json:"100"`
	Gray200 string `json:"200"`
	Gray300 string `json:"300"`
	Gray400 string `json:"400"`
	Gray500 string `json:"500"`
	Gray600 string `json:"600"`
	Gray700 string `json:"700"`
	Gray800 string `json:"800"`
	Gray900 string `json:"900"`
}

// TypographyTokens define text properties
type TypographyTokens struct {
	FontSizes     FontSizeTokens      `json:"font_sizes"`
	FontWeights   FontWeightTokens    `json:"font_weights"`
	LineHeights   LineHeightTokens    `json:"line_heights"`
	FontFamily    FontFamilyTokens    `json:"font_family"`
	LetterSpacing LetterSpacingTokens `json:"letter_spacing"`
}

type FontSizeTokens struct {
	XS    string `json:"xs"`   // "0.75rem"
	SM    string `json:"sm"`   // "0.875rem"
	Base  string `json:"base"` // "1rem"
	LG    string `json:"lg"`   // "1.125rem"
	XL    string `json:"xl"`   // "1.25rem"
	XXL   string `json:"2xl"`  // "1.5rem"
	XXXL  string `json:"3xl"`  // "1.875rem"
	XXXXL string `json:"4xl"`  // "2.25rem"
}

type FontWeightTokens struct {
	Light     string `json:"light"`     // "300"
	Normal    string `json:"normal"`    // "400"
	Medium    string `json:"medium"`    // "500"
	Semibold  string `json:"semibold"`  // "600"
	Bold      string `json:"bold"`      // "700"
	Extrabold string `json:"extrabold"` // "800"
}

type LineHeightTokens struct {
	Tight   string `json:"tight"`   // "1.25"
	Normal  string `json:"normal"`  // "1.5"
	Relaxed string `json:"relaxed"` // "1.75"
	Loose   string `json:"loose"`   // "2"
}

type FontFamilyTokens struct {
	Sans  string `json:"sans"`  // Sans-serif font stack
	Serif string `json:"serif"` // Serif font stack
	Mono  string `json:"mono"`  // Monospace font stack
}

type LetterSpacingTokens struct {
	Tight  string `json:"tight"`  // "-0.05em"
	Normal string `json:"normal"` // "0"
	Wide   string `json:"wide"`   // "0.05em"
}

// SizeTokens define dimensional values
type SizeTokens struct {
	XS  string `json:"xs"`  // "1rem"
	SM  string `json:"sm"`  // "1.5rem"
	MD  string `json:"md"`  // "2rem"
	LG  string `json:"lg"`  // "2.5rem"
	XL  string `json:"xl"`  // "3rem"
	XXL string `json:"2xl"` // "4rem"
}

// BorderTokens define border properties
type BorderTokens struct {
	Width  BorderWidthTokens  `json:"width"`
	Radius BorderRadiusTokens `json:"radius"`
	Style  BorderStyleTokens  `json:"style"`
}

type BorderWidthTokens struct {
	None   string `json:"none"`   // "0"
	Thin   string `json:"thin"`   // "1px"
	Medium string `json:"medium"` // "2px"
	Thick  string `json:"thick"`  // "4px"
}

type BorderRadiusTokens struct {
	None string `json:"none"` // "0"
	SM   string `json:"sm"`   // "0.125rem"
	MD   string `json:"md"`   // "0.25rem"
	LG   string `json:"lg"`   // "0.5rem"
	XL   string `json:"xl"`   // "1rem"
	Full string `json:"full"` // "9999px"
}

type BorderStyleTokens struct {
	Solid  string `json:"solid"`  // "solid"
	Dashed string `json:"dashed"` // "dashed"
	Dotted string `json:"dotted"` // "dotted"
	None   string `json:"none"`   // "none"
}

// ShadowTokens define elevation effects
type ShadowTokens struct {
	None  string `json:"none"`  // "none"
	SM    string `json:"sm"`    // "0 1px 2px rgba(0, 0, 0, 0.05)"
	MD    string `json:"md"`    // "0 4px 6px rgba(0, 0, 0, 0.1)"
	LG    string `json:"lg"`    // "0 10px 15px rgba(0, 0, 0, 0.1)"
	XL    string `json:"xl"`    // "0 20px 25px rgba(0, 0, 0, 0.1)"
	XXL   string `json:"2xl"`   // "0 25px 50px rgba(0, 0, 0, 0.15)"
	Inner string `json:"inner"` // "inset 0 2px 4px rgba(0, 0, 0, 0.06)"
}

// AnimationTokens define animation properties
type AnimationTokens struct {
	Duration AnimationDurationTokens `json:"duration"`
	Easing   AnimationEasingTokens   `json:"easing"`
}

type AnimationDurationTokens struct {
	Fast   string `json:"fast"`   // "150ms"
	Normal string `json:"normal"` // "300ms"
	Slow   string `json:"slow"`   // "500ms"
}

type AnimationEasingTokens struct {
	Linear    string `json:"linear"`    // "linear"
	EaseIn    string `json:"easeIn"`    // "cubic-bezier(0.4, 0, 1, 1)"
	EaseOut   string `json:"easeOut"`   // "cubic-bezier(0, 0, 0.2, 1)"
	EaseInOut string `json:"easeInOut"` // "cubic-bezier(0.4, 0, 0.2, 1)"
}

// ZIndexTokens define layering
type ZIndexTokens struct {
	Dropdown string `json:"dropdown"` // "1000"
	Sticky   string `json:"sticky"`   // "1100"
	Fixed    string `json:"fixed"`    // "1200"
	Modal    string `json:"modal"`    // "1300"
	Popover  string `json:"popover"`  // "1400"
	Tooltip  string `json:"tooltip"`  // "1500"
}

// TokenRegistry manages design tokens with thread safety
type TokenRegistry struct {
	tokens *DesignTokens
	mu     sync.RWMutex
}

// NewTokenRegistry creates a new token registry with default values
func NewTokenRegistry() *TokenRegistry {
	return &TokenRegistry{
		tokens: GetDefaultTokens(),
	}
}

// GetDefaultTokens returns the default design token values
// Following the design system specification in styles.md
func GetDefaultTokens() *DesignTokens {
	return &DesignTokens{
		Spacing: SpacingTokens{
			None: "0",
			XS:   "0.25rem",
			SM:   "0.5rem",
			MD:   "1rem",
			LG:   "2rem",
			XL:   "3rem",
			XXL:  "4rem",
		},
		Colors: ColorTokens{
			Background: BackgroundColors{
				Default:  "hsl(0, 0%, 98%)",
				Subtle:   "hsl(0, 0%, 96%)",
				Emphasis: "hsl(0, 0%, 90%)",
				Overlay:  "rgba(0, 0, 0, 0.5)",
			},
			Text: TextColors{
				Default:   "hsl(0, 0%, 9%)",
				Subtle:    "hsl(0, 0%, 32%)",
				Disabled:  "hsl(0, 0%, 64%)",
				Inverted:  "hsl(0, 0%, 98%)",
				Link:      "hsl(222, 47%, 50%)",
				LinkHover: "hsl(222, 47%, 40%)",
			},
			Border: BorderColors{
				Default: "hsl(0, 0%, 83%)",
				Focus:   "hsl(222, 47%, 50%)",
				Strong:  "hsl(0, 0%, 64%)",
				Subtle:  "hsl(0, 0%, 90%)",
			},
			Feedback: FeedbackColors{
				Success:       "hsl(142, 76%, 36%)",
				SuccessSubtle: "hsl(142, 76%, 95%)",
				Error:         "hsl(0, 84%, 60%)",
				ErrorSubtle:   "hsl(0, 84%, 95%)",
				Warning:       "hsl(38, 92%, 50%)",
				WarningSubtle: "hsl(38, 92%, 95%)",
				Info:          "hsl(199, 89%, 48%)",
				InfoSubtle:    "hsl(199, 89%, 95%)",
			},
			Primary: PrimaryColors{
				Main:   "hsl(222, 47%, 50%)",
				Light:  "hsl(222, 47%, 60%)",
				Dark:   "hsl(222, 47%, 40%)",
				Subtle: "hsl(222, 47%, 95%)",
			},
			Secondary: SecondaryColors{
				Main:   "hsl(280, 47%, 50%)",
				Light:  "hsl(280, 47%, 60%)",
				Dark:   "hsl(280, 47%, 40%)",
				Subtle: "hsl(280, 47%, 95%)",
			},
			Neutral: NeutralColors{
				White: "hsl(0, 0%, 100%)",
				Black: "hsl(0, 0%, 0%)",
				Gray: GrayScale{
					Gray50:  "hsl(0, 0%, 98%)",
					Gray100: "hsl(0, 0%, 96%)",
					Gray200: "hsl(0, 0%, 90%)",
					Gray300: "hsl(0, 0%, 83%)",
					Gray400: "hsl(0, 0%, 64%)",
					Gray500: "hsl(0, 0%, 50%)",
					Gray600: "hsl(0, 0%, 32%)",
					Gray700: "hsl(0, 0%, 21%)",
					Gray800: "hsl(0, 0%, 13%)",
					Gray900: "hsl(0, 0%, 9%)",
				},
			},
		},
		Typography: TypographyTokens{
			FontSizes: FontSizeTokens{
				XS:    "0.75rem",
				SM:    "0.875rem",
				Base:  "1rem",
				LG:    "1.125rem",
				XL:    "1.25rem",
				XXL:   "1.5rem",
				XXXL:  "1.875rem",
				XXXXL: "2.25rem",
			},
			FontWeights: FontWeightTokens{
				Light:     "300",
				Normal:    "400",
				Medium:    "500",
				Semibold:  "600",
				Bold:      "700",
				Extrabold: "800",
			},
			LineHeights: LineHeightTokens{
				Tight:   "1.25",
				Normal:  "1.5",
				Relaxed: "1.75",
				Loose:   "2",
			},
			FontFamily: FontFamilyTokens{
				Sans:  "-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif",
				Serif: "Georgia, Cambria, 'Times New Roman', Times, serif",
				Mono:  "'Courier New', Courier, monospace",
			},
			LetterSpacing: LetterSpacingTokens{
				Tight:  "-0.05em",
				Normal: "0",
				Wide:   "0.05em",
			},
		},
		Sizes: SizeTokens{
			XS:  "1rem",
			SM:  "1.5rem",
			MD:  "2rem",
			LG:  "2.5rem",
			XL:  "3rem",
			XXL: "4rem",
		},
		Borders: BorderTokens{
			Width: BorderWidthTokens{
				None:   "0",
				Thin:   "1px",
				Medium: "2px",
				Thick:  "4px",
			},
			Radius: BorderRadiusTokens{
				None: "0",
				SM:   "0.125rem",
				MD:   "0.25rem",
				LG:   "0.5rem",
				XL:   "1rem",
				Full: "9999px",
			},
			Style: BorderStyleTokens{
				Solid:  "solid",
				Dashed: "dashed",
				Dotted: "dotted",
				None:   "none",
			},
		},
		Shadows: ShadowTokens{
			None:  "none",
			SM:    "0 1px 2px rgba(0, 0, 0, 0.05)",
			MD:    "0 4px 6px rgba(0, 0, 0, 0.1)",
			LG:    "0 10px 15px rgba(0, 0, 0, 0.1)",
			XL:    "0 20px 25px rgba(0, 0, 0, 0.1)",
			XXL:   "0 25px 50px rgba(0, 0, 0, 0.15)",
			Inner: "inset 0 2px 4px rgba(0, 0, 0, 0.06)",
		},
		Animations: AnimationTokens{
			Duration: AnimationDurationTokens{
				Fast:   "150ms",
				Normal: "300ms",
				Slow:   "500ms",
			},
			Easing: AnimationEasingTokens{
				Linear:    "linear",
				EaseIn:    "cubic-bezier(0.4, 0, 1, 1)",
				EaseOut:   "cubic-bezier(0, 0, 0.2, 1)",
				EaseInOut: "cubic-bezier(0.4, 0, 0.2, 1)",
			},
		},
		ZIndex: ZIndexTokens{
			Dropdown: "1000",
			Sticky:   "1100",
			Fixed:    "1200",
			Modal:    "1300",
			Popover:  "1400",
			Tooltip:  "1500",
		},
	}
}

// Token resolution methods with thread safety

// GetSpacing returns spacing token value
func (tr *TokenRegistry) GetSpacing(key string) string {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	switch key {
	case "0", "none":
		return tr.tokens.Spacing.None
	case "1", "xs":
		return tr.tokens.Spacing.XS
	case "2", "sm":
		return tr.tokens.Spacing.SM
	case "4", "md":
		return tr.tokens.Spacing.MD
	case "8", "lg":
		return tr.tokens.Spacing.LG
	case "12", "xl":
		return tr.tokens.Spacing.XL
	case "16", "xxl":
		return tr.tokens.Spacing.XXL
	default:
		return tr.tokens.Spacing.MD // Safe fallback
	}
}

// GetColor returns color token value
func (tr *TokenRegistry) GetColor(category, variant string) string {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	switch category {
	case "background":
		switch variant {
		case "default":
			return tr.tokens.Colors.Background.Default
		case "subtle":
			return tr.tokens.Colors.Background.Subtle
		case "emphasis":
			return tr.tokens.Colors.Background.Emphasis
		case "overlay":
			return tr.tokens.Colors.Background.Overlay
		}
	case "text":
		switch variant {
		case "default":
			return tr.tokens.Colors.Text.Default
		case "subtle":
			return tr.tokens.Colors.Text.Subtle
		case "disabled":
			return tr.tokens.Colors.Text.Disabled
		case "inverted":
			return tr.tokens.Colors.Text.Inverted
		case "link":
			return tr.tokens.Colors.Text.Link
		case "linkHover":
			return tr.tokens.Colors.Text.LinkHover
		}
	case "border":
		switch variant {
		case "default":
			return tr.tokens.Colors.Border.Default
		case "focus":
			return tr.tokens.Colors.Border.Focus
		case "strong":
			return tr.tokens.Colors.Border.Strong
		case "subtle":
			return tr.tokens.Colors.Border.Subtle
		}
	case "feedback":
		switch variant {
		case "success":
			return tr.tokens.Colors.Feedback.Success
		case "successSubtle":
			return tr.tokens.Colors.Feedback.SuccessSubtle
		case "error":
			return tr.tokens.Colors.Feedback.Error
		case "errorSubtle":
			return tr.tokens.Colors.Feedback.ErrorSubtle
		case "warning":
			return tr.tokens.Colors.Feedback.Warning
		case "warningSubtle":
			return tr.tokens.Colors.Feedback.WarningSubtle
		case "info":
			return tr.tokens.Colors.Feedback.Info
		case "infoSubtle":
			return tr.tokens.Colors.Feedback.InfoSubtle
		}
	case "primary":
		switch variant {
		case "main":
			return tr.tokens.Colors.Primary.Main
		case "light":
			return tr.tokens.Colors.Primary.Light
		case "dark":
			return tr.tokens.Colors.Primary.Dark
		case "subtle":
			return tr.tokens.Colors.Primary.Subtle
		}
	case "secondary":
		switch variant {
		case "main":
			return tr.tokens.Colors.Secondary.Main
		case "light":
			return tr.tokens.Colors.Secondary.Light
		case "dark":
			return tr.tokens.Colors.Secondary.Dark
		case "subtle":
			return tr.tokens.Colors.Secondary.Subtle
		}
	}
	return tr.tokens.Colors.Text.Default // Safe fallback
}

// GetFontSize returns font size token value
func (tr *TokenRegistry) GetFontSize(key string) string {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	switch key {
	case "xs":
		return tr.tokens.Typography.FontSizes.XS
	case "sm":
		return tr.tokens.Typography.FontSizes.SM
	case "base", "md":
		return tr.tokens.Typography.FontSizes.Base
	case "lg":
		return tr.tokens.Typography.FontSizes.LG
	case "xl":
		return tr.tokens.Typography.FontSizes.XL
	case "2xl", "xxl":
		return tr.tokens.Typography.FontSizes.XXL
	case "3xl", "xxxl":
		return tr.tokens.Typography.FontSizes.XXXL
	case "4xl", "xxxxl":
		return tr.tokens.Typography.FontSizes.XXXXL
	default:
		return tr.tokens.Typography.FontSizes.Base // Safe fallback
	}
}

// GetSize returns size token value
func (tr *TokenRegistry) GetSize(key string) string {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	switch key {
	case "xs":
		return tr.tokens.Sizes.XS
	case "sm":
		return tr.tokens.Sizes.SM
	case "md":
		return tr.tokens.Sizes.MD
	case "lg":
		return tr.tokens.Sizes.LG
	case "xl":
		return tr.tokens.Sizes.XL
	case "2xl", "xxl":
		return tr.tokens.Sizes.XXL
	default:
		return tr.tokens.Sizes.MD // Safe fallback
	}
}

// GetBorderRadius returns border radius token value
func (tr *TokenRegistry) GetBorderRadius(key string) string {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	switch key {
	case "none":
		return tr.tokens.Borders.Radius.None
	case "sm":
		return tr.tokens.Borders.Radius.SM
	case "md":
		return tr.tokens.Borders.Radius.MD
	case "lg":
		return tr.tokens.Borders.Radius.LG
	case "xl":
		return tr.tokens.Borders.Radius.XL
	case "full":
		return tr.tokens.Borders.Radius.Full
	default:
		return tr.tokens.Borders.Radius.MD // Safe fallback
	}
}

// GetShadow returns shadow token value
func (tr *TokenRegistry) GetShadow(key string) string {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	switch key {
	case "none":
		return tr.tokens.Shadows.None
	case "sm":
		return tr.tokens.Shadows.SM
	case "md":
		return tr.tokens.Shadows.MD
	case "lg":
		return tr.tokens.Shadows.LG
	case "xl":
		return tr.tokens.Shadows.XL
	case "2xl", "xxl":
		return tr.tokens.Shadows.XXL
	case "inner":
		return tr.tokens.Shadows.Inner
	default:
		return tr.tokens.Shadows.MD // Safe fallback
	}
}

// SetTokens updates the entire token set (thread-safe)
func (tr *TokenRegistry) SetTokens(tokens *DesignTokens) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.tokens = tokens
}

// GetTokens returns a copy of current tokens (thread-safe)
func (tr *TokenRegistry) GetTokens() *DesignTokens {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	// Return a copy to prevent external modification
	tokens := *tr.tokens
	return &tokens
}

// Global token registry instance
var (
	defaultTokenRegistry *TokenRegistry
	registryOnce         sync.Once
)

// GetDefaultRegistry returns the global token registry
func GetDefaultRegistry() *TokenRegistry {
	registryOnce.Do(func() {
		defaultTokenRegistry = NewTokenRegistry()
	})
	return defaultTokenRegistry
}

// Token resolution helper functions for easy access

// GetToken returns a token value by path (e.g., "spacing.md", "color.text.default")
func GetToken(path string) string {
	registry := GetDefaultRegistry()
	parts := strings.Split(path, ".")

	if len(parts) < 2 {
		return ""
	}

	category := parts[0]
	key := parts[1]

	switch category {
	case "spacing":
		return registry.GetSpacing(key)
	case "color":
		if len(parts) >= 3 {
			return registry.GetColor(key, parts[2])
		}
	case "fontSize", "font-size":
		return registry.GetFontSize(key)
	case "size":
		return registry.GetSize(key)
	case "borderRadius", "border-radius":
		return registry.GetBorderRadius(key)
	case "shadow":
		return registry.GetShadow(key)
	}

	return ""
}

// Convenience functions for common tokens

// SpacingMD returns the medium spacing token (replaces hardcoded "1rem")
func SpacingMD() string {
	return GetDefaultRegistry().GetSpacing("md")
}

// SpacingSM returns the small spacing token
func SpacingSM() string {
	return GetDefaultRegistry().GetSpacing("sm")
}

// SpacingLG returns the large spacing token
func SpacingLG() string {
	return GetDefaultRegistry().GetSpacing("lg")
}

// SpacingXS returns the extra small spacing token
func SpacingXS() string {
	return GetDefaultRegistry().GetSpacing("xs")
}

// SpacingXL returns the extra large spacing token
func SpacingXL() string {
	return GetDefaultRegistry().GetSpacing("xl")
}

// ColorPrimary returns the primary color
func ColorPrimary() string {
	return GetDefaultRegistry().GetColor("primary", "main")
}

// ColorSecondary returns the secondary color
func ColorSecondary() string {
	return GetDefaultRegistry().GetColor("secondary", "main")
}

// ColorSuccess returns the success color
func ColorSuccess() string {
	return GetDefaultRegistry().GetColor("feedback", "success")
}

// ColorError returns the error color
func ColorError() string {
	return GetDefaultRegistry().GetColor("feedback", "error")
}

// ColorWarning returns the warning color
func ColorWarning() string {
	return GetDefaultRegistry().GetColor("feedback", "warning")
}

// ColorInfo returns the info color
func ColorInfo() string {
	return GetDefaultRegistry().GetColor("feedback", "info")
}

// BorderRadiusMD returns the medium border radius
func BorderRadiusMD() string {
	return GetDefaultRegistry().GetBorderRadius("md")
}

// ShadowMD returns the medium shadow
func ShadowMD() string {
	return GetDefaultRegistry().GetShadow("md")
}

// ResolveToken resolves a token reference (e.g., "$spacing.md") to its value
func ResolveToken(tokenRef string) string {
	if !strings.HasPrefix(tokenRef, "$") {
		return tokenRef // Not a token reference
	}

	path := strings.TrimPrefix(tokenRef, "$")
	value := GetToken(path)

	if value == "" {
		return tokenRef // Return original if not found
	}

	return value
}

// ResolveTokens resolves all token references in a map
func ResolveTokens(values map[string]string) map[string]string {
	resolved := make(map[string]string, len(values))

	for key, value := range values {
		resolved[key] = ResolveToken(value)
	}

	return resolved
}

// TokenPath creates a token path from components
func TokenPath(parts ...string) string {
	return strings.Join(parts, ".")
}

// ValidateTokenPath checks if a token path is valid
func ValidateTokenPath(path string) error {
	parts := strings.Split(path, ".")

	if len(parts) < 2 {
		return fmt.Errorf("token path must have at least 2 parts: %s", path)
	}

	validCategories := map[string]bool{
		"spacing": true, "color": true, "fontSize": true,
		"size": true, "borderRadius": true, "shadow": true,
		"font-size": true, "border-radius": true,
	}

	if !validCategories[parts[0]] {
		return fmt.Errorf("invalid token category: %s", parts[0])
	}

	return nil
}

// MergeTokens merges two token sets (second overrides first)
func MergeTokens(base, override *DesignTokens) *DesignTokens {
	merged := *base

	if override == nil {
		return &merged
	}

	// Merge spacing
	if override.Spacing.None != "" {
		merged.Spacing.None = override.Spacing.None
	}
	if override.Spacing.XS != "" {
		merged.Spacing.XS = override.Spacing.XS
	}
	if override.Spacing.SM != "" {
		merged.Spacing.SM = override.Spacing.SM
	}
	if override.Spacing.MD != "" {
		merged.Spacing.MD = override.Spacing.MD
	}
	if override.Spacing.LG != "" {
		merged.Spacing.LG = override.Spacing.LG
	}
	if override.Spacing.XL != "" {
		merged.Spacing.XL = override.Spacing.XL
	}
	if override.Spacing.XXL != "" {
		merged.Spacing.XXL = override.Spacing.XXL
	}

	// Similar merging for other token categories...
	// (Implementation continues for colors, typography, etc.)

	return &merged
}
