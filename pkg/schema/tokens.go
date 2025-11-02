package schema

// DesignTokens represents the design token system
// Based on the comprehensive design system documented in styles.md
type DesignTokens struct {
	Spacing    SpacingTokens    `json:"spacing"`
	Colors     ColorTokens      `json:"colors"`
	Typography TypographyTokens `json:"typography"`
	Sizes      SizeTokens       `json:"sizes"`
	Borders    BorderTokens     `json:"borders"`
	Shadows    ShadowTokens     `json:"shadows"`
}

// SpacingTokens define all spacing values following design system
type SpacingTokens struct {
	None    string `json:"0"`     // "0"
	XS      string `json:"1"`     // "0.25rem"
	SM      string `json:"2"`     // "0.5rem"
	MD      string `json:"4"`     // "1rem" - Base spacing unit
	LG      string `json:"8"`     // "2rem"
	XL      string `json:"12"`    // "3rem"
	XXL     string `json:"16"`    // "4rem"
}

// ColorTokens define semantic color mappings
type ColorTokens struct {
	Background BackgroundColors `json:"background"`
	Text       TextColors       `json:"text"`
	Border     BorderColors     `json:"border"`
	Feedback   FeedbackColors   `json:"feedback"`
}

type BackgroundColors struct {
	Default string `json:"default"`
	Subtle  string `json:"subtle"`
	Emphasis string `json:"emphasis"`
}

type TextColors struct {
	Default  string `json:"default"`
	Subtle   string `json:"subtle"`
	Disabled string `json:"disabled"`
}

type BorderColors struct {
	Default string `json:"default"`
	Focus   string `json:"focus"`
	Strong  string `json:"strong"`
}

type FeedbackColors struct {
	Success string `json:"success"`
	Error   string `json:"error"`
	Warning string `json:"warning"`
	Info    string `json:"info"`
}

// TypographyTokens define text properties
type TypographyTokens struct {
	FontSizes   FontSizeTokens   `json:"font_sizes"`
	FontWeights FontWeightTokens `json:"font_weights"`
	LineHeights LineHeightTokens `json:"line_heights"`
}

type FontSizeTokens struct {
	XS   string `json:"xs"`   // "0.75rem"
	SM   string `json:"sm"`   // "0.875rem"
	Base string `json:"base"` // "1rem"
	LG   string `json:"lg"`   // "1.125rem"
	XL   string `json:"xl"`   // "1.25rem"
	XXL  string `json:"2xl"`  // "1.5rem"
}

type FontWeightTokens struct {
	Light   string `json:"light"`   // "300"
	Normal  string `json:"normal"`  // "400"
	Medium  string `json:"medium"`  // "500"
	Semibold string `json:"semibold"` // "600"
	Bold    string `json:"bold"`    // "700"
}

type LineHeightTokens struct {
	Tight  string `json:"tight"`  // "1.25"
	Normal string `json:"normal"` // "1.5"
	Relaxed string `json:"relaxed"` // "1.75"
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
}

type BorderWidthTokens struct {
	None   string `json:"none"` // "0"
	Thin   string `json:"thin"` // "1px"
	Medium string `json:"medium"` // "2px"
	Thick  string `json:"thick"` // "4px"
}

type BorderRadiusTokens struct {
	None string `json:"none"` // "0"
	SM   string `json:"sm"`   // "0.125rem"
	MD   string `json:"md"`   // "0.25rem"
	LG   string `json:"lg"`   // "0.5rem"
	XL   string `json:"xl"`   // "1rem"
	Full string `json:"full"` // "9999px"
}

// ShadowTokens define elevation effects
type ShadowTokens struct {
	SM string `json:"sm"` // "0 1px 2px rgba(0, 0, 0, 0.05)"
	MD string `json:"md"` // "0 4px 6px rgba(0, 0, 0, 0.1)"
	LG string `json:"lg"` // "0 10px 15px rgba(0, 0, 0, 0.1)"
	XL string `json:"xl"` // "0 20px 25px rgba(0, 0, 0, 0.1)"
}

// TokenRegistry manages design tokens
type TokenRegistry struct {
	tokens *DesignTokens
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
			MD:   "1rem",     // This replaces hardcoded "1rem"
			LG:   "2rem",
			XL:   "3rem",
			XXL:  "4rem",
		},
		Colors: ColorTokens{
			Background: BackgroundColors{
				Default:  "hsl(0, 0%, 98%)",
				Subtle:   "hsl(0, 0%, 96%)",
				Emphasis: "hsl(0, 0%, 90%)",
			},
			Text: TextColors{
				Default:  "hsl(0, 0%, 9%)",
				Subtle:   "hsl(0, 0%, 32%)",
				Disabled: "hsl(0, 0%, 64%)",
			},
			Border: BorderColors{
				Default: "hsl(0, 0%, 83%)",
				Focus:   "hsl(222, 47%, 50%)",
				Strong:  "hsl(0, 0%, 64%)",
			},
			Feedback: FeedbackColors{
				Success: "hsl(142, 76%, 36%)",
				Error:   "hsl(0, 84%, 60%)",
				Warning: "hsl(38, 92%, 50%)",
				Info:    "hsl(199, 89%, 48%)",
			},
		},
		Typography: TypographyTokens{
			FontSizes: FontSizeTokens{
				XS:   "0.75rem",
				SM:   "0.875rem",
				Base: "1rem",
				LG:   "1.125rem",
				XL:   "1.25rem",
				XXL:  "1.5rem",
			},
			FontWeights: FontWeightTokens{
				Light:    "300",
				Normal:   "400",
				Medium:   "500",
				Semibold: "600",
				Bold:     "700",
			},
			LineHeights: LineHeightTokens{
				Tight:   "1.25",
				Normal:  "1.5",
				Relaxed: "1.75",
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
		},
		Shadows: ShadowTokens{
			SM: "0 1px 2px rgba(0, 0, 0, 0.05)",
			MD: "0 4px 6px rgba(0, 0, 0, 0.1)",
			LG: "0 10px 15px rgba(0, 0, 0, 0.1)",
			XL: "0 20px 25px rgba(0, 0, 0, 0.1)",
		},
	}
}

// Token resolution methods

// GetSpacing returns spacing token value
func (tr *TokenRegistry) GetSpacing(key string) string {
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
	switch category {
	case "background":
		switch variant {
		case "default":
			return tr.tokens.Colors.Background.Default
		case "subtle":
			return tr.tokens.Colors.Background.Subtle
		case "emphasis":
			return tr.tokens.Colors.Background.Emphasis
		}
	case "text":
		switch variant {
		case "default":
			return tr.tokens.Colors.Text.Default
		case "subtle":
			return tr.tokens.Colors.Text.Subtle
		case "disabled":
			return tr.tokens.Colors.Text.Disabled
		}
	case "border":
		switch variant {
		case "default":
			return tr.tokens.Colors.Border.Default
		case "focus":
			return tr.tokens.Colors.Border.Focus
		case "strong":
			return tr.tokens.Colors.Border.Strong
		}
	case "feedback":
		switch variant {
		case "success":
			return tr.tokens.Colors.Feedback.Success
		case "error":
			return tr.tokens.Colors.Feedback.Error
		case "warning":
			return tr.tokens.Colors.Feedback.Warning
		case "info":
			return tr.tokens.Colors.Feedback.Info
		}
	}
	return tr.tokens.Colors.Text.Default // Safe fallback
}

// GetFontSize returns font size token value
func (tr *TokenRegistry) GetFontSize(key string) string {
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
	default:
		return tr.tokens.Typography.FontSizes.Base // Safe fallback
	}
}

// GetSize returns size token value
func (tr *TokenRegistry) GetSize(key string) string {
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

// Global token registry instance
var defaultTokenRegistry = NewTokenRegistry()

// Token resolution helper functions for easy access

// GetToken returns a token value by path (e.g., "spacing.md", "color.text.default")
func GetToken(path string) string {
	// Implementation would parse the path and return the appropriate token
	// For now, we'll implement the most common cases
	switch path {
	case "spacing.md", "spacing.4":
		return defaultTokenRegistry.GetSpacing("md")
	case "spacing.sm", "spacing.2":
		return defaultTokenRegistry.GetSpacing("sm")
	case "spacing.lg", "spacing.8":
		return defaultTokenRegistry.GetSpacing("lg")
	default:
		return ""
	}
}

// Convenience functions for common tokens

// SpacingMD returns the medium spacing token (replaces hardcoded "1rem")
func SpacingMD() string {
	return defaultTokenRegistry.GetSpacing("md")
}

// SpacingSM returns the small spacing token
func SpacingSM() string {
	return defaultTokenRegistry.GetSpacing("sm")
}

// SpacingLG returns the large spacing token
func SpacingLG() string {
	return defaultTokenRegistry.GetSpacing("lg")
}

// Theme represents the theme configuration for a schema
type Theme struct {
	Name        string            `json:"name" validate:"required"`
	Tokens      *DesignTokens     `json:"tokens,omitempty"`
	Colors      map[string]string `json:"colors,omitempty"`
	Fonts       map[string]string `json:"fonts,omitempty"`
	Breakpoints map[string]string `json:"breakpoints,omitempty"`
	CustomCSS   string            `json:"customCSS,omitempty"`
}

// Tokens represents the token configuration for a schema (alias to DesignTokens)
type Tokens = DesignTokens