package schema

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// TokenReference represents a reference to another token using the syntax "token.path".
// References are resolved at runtime, allowing dynamic theme customization.
// Example: "colors.primary.base" references the base primary color.
type TokenReference string

// IsReference checks if the value is a token reference (starts with { and ends with }).
func (t TokenReference) IsReference() bool {
	s := string(t)
	return len(s) >= 3 && strings.HasPrefix(s, "") && strings.HasSuffix(s, "")
}

// Path extracts the token path from a reference.
// For "colors.primary.base", it returns "colors.primary.base".
func (t TokenReference) Path() string {
	if !t.IsReference() {
		return string(t)
	}
	s := string(t)
	return strings.Trim(s, "")
}

// String returns the string representation of the token reference.
func (t TokenReference) String() string {
	return string(t)
}

// Validate checks if the token reference is well-formed.
func (t TokenReference) Validate() error {
	if !t.IsReference() {
		return nil // Non-references are always valid
	}

	path := t.Path()
	if path == "" {
		return NewValidationError("empty_token_path", "empty token path in reference")
	}

	// Token paths must contain at least one dot
	if !strings.Contains(path, ".") {
		return NewValidationError("invalid_token_path", fmt.Sprintf("invalid token path format: %s (must contain at least one dot)", path))
	}

	// Check for valid characters (alphanumeric, dots, underscores, hyphens)
	for _, char := range path {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '.' || char == '_' || char == '-') {
			return NewValidationError("invalid_token_char", fmt.Sprintf("invalid character '%c' in token path: %s", char, path))
		}
	}

	return nil
}

// DesignTokens is the root container for all design tokens.
// It follows a three-tier architecture for maximum flexibility and maintainability.
//
// Architecture:
//   - Primitives: Raw values (hsl(220, 50%, 50%))
//   - Semantic: Functional meaning ({colors.primary.base})
//   - Components: Component-specific styles ({semantic.interactive.default})
//
// Breaking Changes: The structure may evolve to add new token categories.
// Always use token references to ensure forward compatibility.
type DesignTokens struct {
	// Primitives contains raw design values - the foundation of the system
	Primitives *PrimitiveTokens `json:"primitives"`

	// Semantic contains functional token assignments
	Semantic *SemanticTokens `json:"semantic"`

	// Components contains component-specific token assignments
	Components *ComponentTokens `json:"components"`
}

// PrimitiveTokens contains the raw design values that form the foundation
// of the design system. These values should rarely change and represent
// the atomic design decisions.
type PrimitiveTokens struct {
	Colors     *ColorPrimitives      `json:"colors"`
	Spacing    *SpacingScale         `json:"spacing"`
	Typography *TypographyPrimitives `json:"typography"`
	Borders    *BorderPrimitives     `json:"borders"`
	Shadows    *ShadowScale          `json:"shadows"`
	Animations *AnimationPrimitives  `json:"animations"`
	Sizes      *SizeScale            `json:"sizes"`
	ZIndex     *ZIndexScale          `json:"zIndex"`
}

// ColorPrimitives defines the raw color palette.
// Uses HSL format for better manipulation and theming.
type ColorPrimitives struct {
	// Neutral colors - foundation for text, borders, and backgrounds
	Gray *GrayScale `json:"gray"`

	// Brand colors - full scales for primary, secondary, accent
	Blue   *ColorScale `json:"blue"`   // Primary brand color
	Purple *ColorScale `json:"purple"` // Secondary brand color
	Cyan   *ColorScale `json:"cyan"`   // Accent color

	// Semantic colors - full scales for feedback states
	Green  *ColorScale `json:"green"`  // Success states
	Red    *ColorScale `json:"red"`    // Error states
	Yellow *ColorScale `json:"yellow"` // Warning states
	Sky    *ColorScale `json:"sky"`    // Info states

	// Pure colors
	White string `json:"white"`
	Black string `json:"black"`
}

// ColorScale represents a complete color scale from 50 (lightest) to 900 (darkest).
// Follows industry standards for predictable color relationships.
type ColorScale struct {
	Scale50  string `json:"50"`
	Scale100 string `json:"100"`
	Scale200 string `json:"200"`
	Scale300 string `json:"300"`
	Scale400 string `json:"400"`
	Scale500 string `json:"500"` // Base color
	Scale600 string `json:"600"`
	Scale700 string `json:"700"`
	Scale800 string `json:"800"`
	Scale900 string `json:"900"`
}

// GrayScale represents the neutral color scale used throughout the system.
type GrayScale struct {
	Scale50  string `json:"50"`
	Scale100 string `json:"100"`
	Scale200 string `json:"200"`
	Scale300 string `json:"300"`
	Scale400 string `json:"400"`
	Scale500 string `json:"500"`
	Scale600 string `json:"600"`
	Scale700 string `json:"700"`
	Scale800 string `json:"800"`
	Scale900 string `json:"900"`
}

// SpacingScale defines the spacing scale used for margins, padding, and gaps.
type SpacingScale struct {
	None string `json:"0"`   // 0
	XS   string `json:"xs"`  // 0.25rem
	SM   string `json:"sm"`  // 0.5rem
	MD   string `json:"md"`  // 1rem - Base spacing unit
	LG   string `json:"lg"`  // 1.5rem
	XL   string `json:"xl"`  // 2rem
	XXL  string `json:"2xl"` // 3rem
	XXXL string `json:"3xl"` // 4rem
	Huge string `json:"4xl"` // 6rem
}

// TypographyPrimitives defines the raw typography values.
type TypographyPrimitives struct {
	FontSizes     *FontSizeScale      `json:"fontSizes"`
	FontWeights   *FontWeightScale    `json:"fontWeights"`
	LineHeights   *LineHeightScale    `json:"lineHeights"`
	FontFamilies  *FontFamilyScale    `json:"fontFamilies"`
	LetterSpacing *LetterSpacingScale `json:"letterSpacing"`
}

// FontSizeScale defines the typography size scale.
type FontSizeScale struct {
	XS   string `json:"xs"`   // 0.75rem
	SM   string `json:"sm"`   // 0.875rem
	Base string `json:"base"` // 1rem
	LG   string `json:"lg"`   // 1.125rem
	XL   string `json:"xl"`   // 1.25rem
	XXL  string `json:"2xl"`  // 1.5rem
	XXXL string `json:"3xl"`  // 1.875rem
	Huge string `json:"4xl"`  // 2.25rem
}

// FontWeightScale defines font weight values.
type FontWeightScale struct {
	Thin      string `json:"thin"`      // 100
	Light     string `json:"light"`     // 300
	Normal    string `json:"normal"`    // 400
	Medium    string `json:"medium"`    // 500
	Semibold  string `json:"semibold"`  // 600
	Bold      string `json:"bold"`      // 700
	Extrabold string `json:"extrabold"` // 800
	Black     string `json:"black"`     // 900
}

// LineHeightScale defines line height values.
type LineHeightScale struct {
	Tight   string `json:"tight"`   // 1.25
	Normal  string `json:"normal"`  // 1.5
	Relaxed string `json:"relaxed"` // 1.75
	Loose   string `json:"loose"`   // 2
}

// FontFamilyScale defines font family stacks.
type FontFamilyScale struct {
	Sans  string `json:"sans"`  // Sans-serif font stack
	Serif string `json:"serif"` // Serif font stack
	Mono  string `json:"mono"`  // Monospace font stack
}

// LetterSpacingScale defines letter spacing values.
type LetterSpacingScale struct {
	Tight  string `json:"tight"`  // -0.05em
	Normal string `json:"normal"` // 0
	Wide   string `json:"wide"`   // 0.05em
}

// BorderPrimitives defines border-related primitive values.
type BorderPrimitives struct {
	Width  *BorderWidthScale  `json:"width"`
	Radius *BorderRadiusScale `json:"radius"`
	Style  *BorderStyleScale  `json:"style"`
}

// BorderWidthScale defines border width values.
type BorderWidthScale struct {
	None   string `json:"none"`   // 0
	Thin   string `json:"thin"`   // 1px
	Medium string `json:"medium"` // 2px
	Thick  string `json:"thick"`  // 4px
}

// BorderRadiusScale defines border radius values.
type BorderRadiusScale struct {
	None string `json:"none"` // 0
	SM   string `json:"sm"`   // 0.125rem
	MD   string `json:"md"`   // 0.25rem
	LG   string `json:"lg"`   // 0.5rem
	XL   string `json:"xl"`   // 1rem
	Full string `json:"full"` // 9999px
}

// BorderStyleScale defines border style values.
type BorderStyleScale struct {
	Solid  string `json:"solid"`  // solid
	Dashed string `json:"dashed"` // dashed
	Dotted string `json:"dotted"` // dotted
	None   string `json:"none"`   // none
}

// ShadowScale defines elevation shadow values.
type ShadowScale struct {
	None  string `json:"none"`  // none
	SM    string `json:"sm"`    // 0 1px 2px rgba(0, 0, 0, 0.05)
	MD    string `json:"md"`    // 0 4px 6px rgba(0, 0, 0, 0.1)
	LG    string `json:"lg"`    // 0 10px 15px rgba(0, 0, 0, 0.1)
	XL    string `json:"xl"`    // 0 20px 25px rgba(0, 0, 0, 0.1)
	XXL   string `json:"2xl"`   // 0 25px 50px rgba(0, 0, 0, 0.15)
	Inner string `json:"inner"` // inset 0 2px 4px rgba(0, 0, 0, 0.06)
}

// AnimationPrimitives defines animation-related primitive values.
type AnimationPrimitives struct {
	Duration *AnimationDurationScale `json:"duration"`
	Easing   *AnimationEasingScale   `json:"easing"`
}

// AnimationDurationScale defines animation duration values.
type AnimationDurationScale struct {
	Fast   string `json:"fast"`   // 150ms
	Normal string `json:"normal"` // 300ms
	Slow   string `json:"slow"`   // 500ms
}

// AnimationEasingScale defines animation easing values.
type AnimationEasingScale struct {
	Linear    string `json:"linear"`    // linear
	EaseIn    string `json:"easeIn"`    // cubic-bezier(0.4, 0, 1, 1)
	EaseOut   string `json:"easeOut"`   // cubic-bezier(0, 0, 0.2, 1)
	EaseInOut string `json:"easeInOut"` // cubic-bezier(0.4, 0, 0.2, 1)
}

// SizeScale defines dimensional values for components.
type SizeScale struct {
	XS  string `json:"xs"`  // 1rem
	SM  string `json:"sm"`  // 1.5rem
	MD  string `json:"md"`  // 2rem
	LG  string `json:"lg"`  // 2.5rem
	XL  string `json:"xl"`  // 3rem
	XXL string `json:"2xl"` // 4rem
}

// ZIndexScale defines layering values.
type ZIndexScale struct {
	Dropdown string `json:"dropdown"` // 1000
	Sticky   string `json:"sticky"`   // 1100
	Fixed    string `json:"fixed"`    // 1200
	Modal    string `json:"modal"`    // 1300
	Popover  string `json:"popover"`  // 1400
	Tooltip  string `json:"tooltip"`  // 1500
}

// SemanticTokens contains functional token assignments that map to primitive values.
// These tokens provide semantic meaning to design decisions.
type SemanticTokens struct {
	Colors      *SemanticColors      `json:"colors"`
	Typography  *SemanticTypography  `json:"typography"`
	Spacing     *SemanticSpacing     `json:"spacing"`
	Interactive *SemanticInteractive `json:"interactive"`
}

// SemanticColors defines semantic color assignments.
type SemanticColors struct {
	// Background colors
	Background *BackgroundColors `json:"background"`

	// Text colors
	Text *TextColors `json:"text"`

	// Border colors
	Border *BorderColors `json:"border"`

	// Interactive colors
	Interactive *InteractiveColors `json:"interactive"`

	// Feedback colors
	Feedback *FeedbackColors `json:"feedback"`
}

// BackgroundColors defines semantic background color assignments.
type BackgroundColors struct {
	Default  TokenReference `json:"default"`  // Primary background
	Subtle   TokenReference `json:"subtle"`   // Subtle background
	Emphasis TokenReference `json:"emphasis"` // Emphasized background
	Overlay  TokenReference `json:"overlay"`  // Overlay background
}

// TextColors defines semantic text color assignments.
type TextColors struct {
	Default   TokenReference `json:"default"`   // Primary text
	Subtle    TokenReference `json:"subtle"`    // Secondary text
	Disabled  TokenReference `json:"disabled"`  // Disabled text
	Inverted  TokenReference `json:"inverted"`  // Inverted text
	Link      TokenReference `json:"link"`      // Link text
	LinkHover TokenReference `json:"linkHover"` // Link hover text
}

// BorderColors defines semantic border color assignments.
type BorderColors struct {
	Default TokenReference `json:"default"` // Default border
	Focus   TokenReference `json:"focus"`   // Focus border
	Strong  TokenReference `json:"strong"`  // Strong border
	Subtle  TokenReference `json:"subtle"`  // Subtle border
}

// InteractiveColors defines semantic interactive color assignments.
type InteractiveColors struct {
	Primary   *InteractiveColorSet `json:"primary"`   // Primary interactive
	Secondary *InteractiveColorSet `json:"secondary"` // Secondary interactive
	Accent    *InteractiveColorSet `json:"accent"`    // Accent interactive
}

// InteractiveColorSet defines a complete set of interactive colors.
type InteractiveColorSet struct {
	Default TokenReference `json:"default"` // Default state
	Hover   TokenReference `json:"hover"`   // Hover state
	Active  TokenReference `json:"active"`  // Active state
	Focus   TokenReference `json:"focus"`   // Focus state
}

// FeedbackColors defines semantic feedback color assignments.
type FeedbackColors struct {
	Success *FeedbackColorSet `json:"success"` // Success feedback
	Error   *FeedbackColorSet `json:"error"`   // Error feedback
	Warning *FeedbackColorSet `json:"warning"` // Warning feedback
	Info    *FeedbackColorSet `json:"info"`    // Info feedback
}

// FeedbackColorSet defines a complete set of feedback colors.
type FeedbackColorSet struct {
	Default TokenReference `json:"default"` // Default feedback color
	Subtle  TokenReference `json:"subtle"`  // Subtle feedback color
	Strong  TokenReference `json:"strong"`  // Strong feedback color
}

// SemanticTypography defines semantic typography assignments.
type SemanticTypography struct {
	Headings *HeadingTokens `json:"headings"`
	Body     *BodyTokens    `json:"body"`
	Labels   *LabelTokens   `json:"labels"`
	Captions *CaptionTokens `json:"captions"`
	Code     *CodeTokens    `json:"code"`
}

// HeadingTokens defines semantic heading typography.
type HeadingTokens struct {
	H1 *TypographyToken `json:"h1"`
	H2 *TypographyToken `json:"h2"`
	H3 *TypographyToken `json:"h3"`
	H4 *TypographyToken `json:"h4"`
	H5 *TypographyToken `json:"h5"`
	H6 *TypographyToken `json:"h6"`
}

// BodyTokens defines semantic body typography.
type BodyTokens struct {
	Large   *TypographyToken `json:"large"`
	Default *TypographyToken `json:"default"`
	Small   *TypographyToken `json:"small"`
}

// LabelTokens defines semantic label typography.
type LabelTokens struct {
	Large   *TypographyToken `json:"large"`
	Default *TypographyToken `json:"default"`
	Small   *TypographyToken `json:"small"`
}

// CaptionTokens defines semantic caption typography.
type CaptionTokens struct {
	Default *TypographyToken `json:"default"`
	Small   *TypographyToken `json:"small"`
}

// CodeTokens defines semantic code typography.
type CodeTokens struct {
	Inline *TypographyToken `json:"inline"`
	Block  *TypographyToken `json:"block"`
}

// TypographyToken represents a complete typography definition.
type TypographyToken struct {
	FontFamily    TokenReference `json:"fontFamily"`
	FontSize      TokenReference `json:"fontSize"`
	FontWeight    TokenReference `json:"fontWeight"`
	LineHeight    TokenReference `json:"lineHeight"`
	LetterSpacing TokenReference `json:"letterSpacing"`
}

// SemanticSpacing defines semantic spacing assignments.
type SemanticSpacing struct {
	Component *ComponentSpacing `json:"component"` // Component spacing
	Layout    *LayoutSpacing    `json:"layout"`    // Layout spacing
}

// ComponentSpacing defines semantic component spacing.
type ComponentSpacing struct {
	Tight   TokenReference `json:"tight"`   // Tight component spacing
	Default TokenReference `json:"default"` // Default component spacing
	Loose   TokenReference `json:"loose"`   // Loose component spacing
}

// LayoutSpacing defines semantic layout spacing.
type LayoutSpacing struct {
	Section TokenReference `json:"section"` // Section spacing
	Page    TokenReference `json:"page"`    // Page spacing
}

// SemanticInteractive defines semantic interactive assignments.
type SemanticInteractive struct {
	BorderRadius *InteractiveBorderRadius `json:"borderRadius"` // Interactive border radius
	Shadow       *InteractiveShadow       `json:"shadow"`       // Interactive shadows
}

// InteractiveBorderRadius defines semantic interactive border radius.
type InteractiveBorderRadius struct {
	Small   TokenReference `json:"small"`   // Small interactive radius
	Default TokenReference `json:"default"` // Default interactive radius
	Large   TokenReference `json:"large"`   // Large interactive radius
}

// InteractiveShadow defines semantic interactive shadows.
type InteractiveShadow struct {
	Default TokenReference `json:"default"` // Default interactive shadow
	Hover   TokenReference `json:"hover"`   // Hover interactive shadow
	Focus   TokenReference `json:"focus"`   // Focus interactive shadow
}

// ComponentTokens contains component-specific token assignments.
// These tokens are used directly by UI components.
type ComponentTokens struct {
	Button     *ButtonTokens     `json:"button"`
	Input      *InputTokens      `json:"input"`
	Card       *CardTokens       `json:"card"`
	Modal      *ModalTokens      `json:"modal"`
	Form       *FormTokens       `json:"form"`
	Table      *TableTokens      `json:"table"`
	Navigation *NavigationTokens `json:"navigation"`
}

// ButtonTokens defines component tokens for buttons.
type ButtonTokens struct {
	Primary     *ButtonVariantTokens `json:"primary"`
	Secondary   *ButtonVariantTokens `json:"secondary"`
	Outline     *ButtonVariantTokens `json:"outline"`
	Ghost       *ButtonVariantTokens `json:"ghost"`
	Destructive *ButtonVariantTokens `json:"destructive"`
}

// ButtonVariantTokens defines tokens for a button variant.
type ButtonVariantTokens struct {
	Background       TokenReference `json:"background"`
	BackgroundHover  TokenReference `json:"backgroundHover"`
	BackgroundActive TokenReference `json:"backgroundActive"`
	Color            TokenReference `json:"color"`
	ColorHover       TokenReference `json:"colorHover"`
	Border           TokenReference `json:"border"`
	BorderHover      TokenReference `json:"borderHover"`
	BorderRadius     TokenReference `json:"borderRadius"`
	Padding          TokenReference `json:"padding"`
	FontWeight       TokenReference `json:"fontWeight"`
	Shadow           TokenReference `json:"shadow"`
	ShadowHover      TokenReference `json:"shadowHover"`
}

// InputTokens defines component tokens for inputs.
type InputTokens struct {
	Background      TokenReference `json:"background"`
	BackgroundFocus TokenReference `json:"backgroundFocus"`
	Border          TokenReference `json:"border"`
	BorderFocus     TokenReference `json:"borderFocus"`
	BorderError     TokenReference `json:"borderError"`
	BorderRadius    TokenReference `json:"borderRadius"`
	Padding         TokenReference `json:"padding"`
	Color           TokenReference `json:"color"`
	Placeholder     TokenReference `json:"placeholder"`
}

// CardTokens defines component tokens for cards.
type CardTokens struct {
	Background   TokenReference `json:"background"`
	Border       TokenReference `json:"border"`
	BorderRadius TokenReference `json:"borderRadius"`
	Shadow       TokenReference `json:"shadow"`
	Padding      TokenReference `json:"padding"`
}

// ModalTokens defines component tokens for modals.
type ModalTokens struct {
	Background   TokenReference `json:"background"`
	Overlay      TokenReference `json:"overlay"`
	Border       TokenReference `json:"border"`
	BorderRadius TokenReference `json:"borderRadius"`
	Shadow       TokenReference `json:"shadow"`
	Padding      TokenReference `json:"padding"`
}

// FormTokens defines component tokens for forms.
type FormTokens struct {
	Background   TokenReference `json:"background"`
	Padding      TokenReference `json:"padding"`
	BorderRadius TokenReference `json:"borderRadius"`
	Shadow       TokenReference `json:"shadow"`
	Spacing      TokenReference `json:"spacing"`
}

// TableTokens defines component tokens for tables.
type TableTokens struct {
	Background       TokenReference `json:"background"`
	BackgroundHover  TokenReference `json:"backgroundHover"`
	Border           TokenReference `json:"border"`
	HeaderBackground TokenReference `json:"headerBackground"`
	HeaderColor      TokenReference `json:"headerColor"`
	Padding          TokenReference `json:"padding"`
}

// NavigationTokens defines component tokens for navigation.
type NavigationTokens struct {
	Background       TokenReference `json:"background"`
	BackgroundHover  TokenReference `json:"backgroundHover"`
	BackgroundActive TokenReference `json:"backgroundActive"`
	Color            TokenReference `json:"color"`
	ColorHover       TokenReference `json:"colorHover"`
	ColorActive      TokenReference `json:"colorActive"`
	Border           TokenReference `json:"border"`
	Padding          TokenReference `json:"padding"`
}

// TokenResolver provides methods for resolving token references to actual values.
// Implementations should handle circular reference detection and caching.
type TokenResolver interface {
	// Resolve resolves a token reference to its actual value
	Resolve(ctx context.Context, reference TokenReference, tokens *DesignTokens) (string, error)

	// ResolveAll resolves all token references in a token set
	ResolveAll(ctx context.Context, tokens *DesignTokens) (*DesignTokens, error)

	// ValidateReferences checks for circular references and invalid paths
	ValidateReferences(tokens *DesignTokens) error
}

// DefaultTokenResolver is the default implementation of TokenResolver with circular reference detection.
type DefaultTokenResolver struct {
	maxDepth int // Maximum resolution depth to prevent infinite loops
}

// NewDefaultTokenResolver creates a new default token resolver.
func NewDefaultTokenResolver() *DefaultTokenResolver {
	return &DefaultTokenResolver{
		maxDepth: 20, // Reasonable depth limit
	}
}

// Resolve resolves a token reference to its actual value with circular reference protection.
func (r *DefaultTokenResolver) Resolve(ctx context.Context, reference TokenReference, tokens *DesignTokens) (string, error) {
	if !reference.IsReference() {
		return reference.String(), nil
	}

	// Track resolution path to detect circular references
	visited := make(map[string]bool)
	return r.resolveWithTracking(ctx, reference, tokens, visited, 0)
}

// resolveWithTracking performs recursive resolution with circular reference detection.
func (r *DefaultTokenResolver) resolveWithTracking(ctx context.Context, reference TokenReference, tokens *DesignTokens, visited map[string]bool, depth int) (string, error) {
	if depth > r.maxDepth {
		return "", NewValidationError("max_depth_exceeded", fmt.Sprintf("maximum resolution depth exceeded for token: %s", reference.String()))
	}

	if !reference.IsReference() {
		return reference.String(), nil
	}

	path := reference.Path()

	// Check for circular reference
	if visited[path] {
		return "", NewValidationError("circular_reference", fmt.Sprintf("circular reference detected in token path: %s", path))
	}

	visited[path] = true
	defer delete(visited, path) // Clean up for backtracking

	// Get the value from the token structure
	value, err := r.getTokenValue(path, tokens)
	if err != nil {
		return "", err
	}

	// If the resolved value is itself a reference, resolve it recursively
	valueRef := TokenReference(value)
	if valueRef.IsReference() {
		return r.resolveWithTracking(ctx, valueRef, tokens, visited, depth+1)
	}

	return value, nil
}

// getTokenValue retrieves a token value by path from the token structure.
func (r *DefaultTokenResolver) getTokenValue(path string, tokens *DesignTokens) (string, error) {
	if tokens == nil {
		return "", NewValidationError("tokens_nil", "tokens is nil")
	}

	parts := strings.Split(path, ".")
	if len(parts) < 2 {
		return "", NewValidationError("invalid_token_path", fmt.Sprintf("invalid token path: %s", path))
	}

	// Navigate through the token structure using reflection
	current := reflect.ValueOf(tokens)
	if current.Kind() == reflect.Ptr {
		current = current.Elem()
	}

	for i, part := range parts {
		if !current.IsValid() {
			return "", NewValidationError("invalid_token_path", fmt.Sprintf("invalid token path at part %d: %s", i, path))
		}

		// Handle different types of navigation
		switch current.Kind() {
		case reflect.Struct:
			// Find field by name (case-insensitive search)
			field := r.findStructField(current, part)
			if !field.IsValid() {
				return "", NewValidationError("token_field_not_found", fmt.Sprintf("token field not found: %s in path %s", part, path))
			}
			current = field

		case reflect.Ptr:
			if current.IsNil() {
				return "", NewValidationError("nil_pointer_in_path", fmt.Sprintf("nil pointer encountered in token path: %s", path))
			}
			current = current.Elem()
			// Retry with the dereferenced value
			field := r.findStructField(current, part)
			if !field.IsValid() {
				return "", NewValidationError("token_field_not_found", fmt.Sprintf("token field not found: %s in path %s", part, path))
			}
			current = field

		default:
			return "", NewValidationError("unsupported_token_type", fmt.Sprintf("cannot navigate token path %s at part %s: unsupported type %s", path, part, current.Kind()))
		}

		// If we reach a pointer, dereference it
		if current.Kind() == reflect.Ptr {
			if current.IsNil() {
				return "", NewValidationError("nil_pointer_in_path", fmt.Sprintf("nil pointer encountered in token path: %s", path))
			}
			current = current.Elem()
		}
	}

	// Extract the final value
	switch current.Kind() {
	case reflect.String:
		return current.String(), nil
	case reflect.Interface:
		if str, ok := current.Interface().(string); ok {
			return str, nil
		}
		if tokenRef, ok := current.Interface().(TokenReference); ok {
			return tokenRef.String(), nil
		}
		return "", NewValidationError("unsupported_value_type", fmt.Sprintf("unsupported value type in token path: %s", path))
	default:
		// Try to convert to string if possible
		if current.CanInterface() {
			if str, ok := current.Interface().(string); ok {
				return str, nil
			}
			if tokenRef, ok := current.Interface().(TokenReference); ok {
				return tokenRef.String(), nil
			}
		}
		return "", NewValidationError("cannot_convert_to_string", fmt.Sprintf("cannot convert token value to string for path: %s", path))
	}
}

// findStructField finds a struct field by name (case-insensitive).
func (r *DefaultTokenResolver) findStructField(structValue reflect.Value, fieldName string) reflect.Value {
	structType := structValue.Type()

	// First try exact match
	if field := structValue.FieldByName(fieldName); field.IsValid() {
		return field
	}

	// Then try case-insensitive match
	for i := 0; i < structValue.NumField(); i++ {
		field := structType.Field(i)
		if strings.EqualFold(field.Name, fieldName) {
			return structValue.Field(i)
		}

		// Also check JSON tag
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			tagParts := strings.Split(jsonTag, ",")
			if len(tagParts) > 0 && strings.EqualFold(tagParts[0], fieldName) {
				return structValue.Field(i)
			}
		}
	}

	return reflect.Value{} // Invalid value
}

// ResolveAll resolves all token references in a token set.
func (r *DefaultTokenResolver) ResolveAll(ctx context.Context, tokens *DesignTokens) (*DesignTokens, error) {
	// This is a complex operation that would need to traverse the entire token structure
	// For now, return the original tokens (this can be implemented later if needed)
	return tokens, nil
}

// ValidateReferences checks for circular references and invalid paths.
func (r *DefaultTokenResolver) ValidateReferences(tokens *DesignTokens) error {
	if tokens == nil {
		return NewValidationError("tokens_nil", "tokens is nil")
	}

	// Collect all token references from the structure
	references := r.collectTokenReferences(tokens)

	// Validate each reference
	for _, ref := range references {
		if err := ref.Validate(); err != nil {
			return WrapError(err, "invalid_token_reference", fmt.Sprintf("invalid token reference %s", ref.String()))
		}

		// Check if the reference can be resolved (this will catch circular references)
		_, err := r.Resolve(context.Background(), ref, tokens)
		if err != nil {
			return WrapError(err, "token_resolution_failed", fmt.Sprintf("failed to resolve token reference %s", ref.String()))
		}
	}

	return nil
}

// collectTokenReferences recursively collects all token references from the token structure.
func (r *DefaultTokenResolver) collectTokenReferences(tokens *DesignTokens) []TokenReference {
	var references []TokenReference
	r.collectReferencesFromValue(reflect.ValueOf(tokens), &references)
	return references
}

// collectReferencesFromValue recursively collects token references from a reflect.Value.
func (r *DefaultTokenResolver) collectReferencesFromValue(v reflect.Value, references *[]TokenReference) {
	if !v.IsValid() {
		return
	}

	switch v.Kind() {
	case reflect.Ptr:
		if !v.IsNil() {
			r.collectReferencesFromValue(v.Elem(), references)
		}

	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			r.collectReferencesFromValue(v.Field(i), references)
		}

	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			r.collectReferencesFromValue(v.Index(i), references)
		}

	case reflect.Map:
		for _, key := range v.MapKeys() {
			r.collectReferencesFromValue(v.MapIndex(key), references)
		}

	case reflect.Interface:
		if !v.IsNil() {
			r.collectReferencesFromValue(v.Elem(), references)
		}

	case reflect.String:
		// Check if this string is a TokenReference
		if v.CanInterface() {
			if tokenRef, ok := v.Interface().(TokenReference); ok {
				if tokenRef.IsReference() {
					*references = append(*references, tokenRef)
				}
			}
		}
	}
}

// CompiledTokenMap provides O(1) token lookups for performance
type CompiledTokenMap struct {
	tokenMap map[string]string
	mu       sync.RWMutex
}

// NewCompiledTokenMap creates a new compiled token map
func NewCompiledTokenMap() *CompiledTokenMap {
	return &CompiledTokenMap{
		tokenMap: make(map[string]string),
	}
}

// Get retrieves a token value by path
func (ctm *CompiledTokenMap) Get(path string) (string, bool) {
	ctm.mu.RLock()
	defer ctm.mu.RUnlock()
	value, exists := ctm.tokenMap[path]
	return value, exists
}

// Set stores a token value at path
func (ctm *CompiledTokenMap) Set(path, value string) {
	ctm.mu.Lock()
	defer ctm.mu.Unlock()
	ctm.tokenMap[path] = value
}

// Clear removes all compiled tokens
func (ctm *CompiledTokenMap) Clear() {
	ctm.mu.Lock()
	defer ctm.mu.Unlock()
	ctm.tokenMap = make(map[string]string)
}

// Size returns the number of compiled tokens
func (ctm *CompiledTokenMap) Size() int {
	ctm.mu.RLock()
	defer ctm.mu.RUnlock()
	return len(ctm.tokenMap)
}

// TokenRegistry manages design tokens with thread safety and caching.
type TokenRegistry struct {
	tokens         *DesignTokens
	resolver       TokenResolver
	compiled       *CompiledTokenMap // Precompiled token map for O(1) lookups
	cache          map[string]string // Resolution cache
	needsRecompile bool              // Flag to trigger recompilation
	mu             sync.RWMutex
}

// NewTokenRegistry creates a new token registry with default values.
func NewTokenRegistry() *TokenRegistry {
	registry := &TokenRegistry{
		tokens:         GetDefaultTokens(),
		resolver:       NewDefaultTokenResolver(),
		compiled:       NewCompiledTokenMap(),
		cache:          make(map[string]string),
		needsRecompile: true,
	}

	// Trigger initial compilation
	registry.compileTokens()

	return registry
}

// NewTokenRegistryWithResolver creates a new token registry with a custom resolver.
func NewTokenRegistryWithResolver(resolver TokenResolver) *TokenRegistry {
	if resolver == nil {
		resolver = NewDefaultTokenResolver()
	}

	registry := &TokenRegistry{
		tokens:         GetDefaultTokens(),
		resolver:       resolver,
		compiled:       NewCompiledTokenMap(),
		cache:          make(map[string]string),
		needsRecompile: true,
	}

	// Trigger initial compilation
	registry.compileTokens()

	return registry
}

// SetTokens updates the entire token set (thread-safe).
//
// Example:
//
//	registry := schema.NewTokenRegistry()
//
//	// Create custom token structure
//	tokens := &schema.DesignTokens{
//	    Primitives: &schema.PrimitiveTokens{
//	        Colors: &schema.ColorPrimitives{
//	            Blue: &schema.ColorScale{
//	                Scale500: "#1e40af", // Custom blue
//	            },
//	            Gray: &schema.ColorScale{
//	                Scale100: "#f5f5f5",
//	                Scale900: "#111827",
//	            },
//	        },
//	        Spacing: &schema.SpacingScale{
//	            Base: "1rem",
//	            SM:   "0.5rem",
//	            LG:   "2rem",
//	        },
//	    },
//	    Semantic: &schema.SemanticTokens{
//	        Colors: &schema.SemanticColors{
//	            Background: &schema.BackgroundColors{
//	                Default: schema.TokenReference("primitives.colors.gray.100"),
//	            },
//	            Text: &schema.TextColors{
//	                Default: schema.TokenReference("primitives.colors.gray.900"),
//	            },
//	            Interactive: &schema.InteractiveColors{
//	                Primary: &schema.InteractiveColorSet{
//	                    Default: schema.TokenReference("primitives.colors.blue.500"),
//	                },
//	            },
//	        },
//	    },
//	}
//
//	// Set the tokens (validates references and compiles for performance)
//	err := registry.SetTokens(tokens)
//	if err != nil {
//	    return fmt.Errorf("failed to set tokens: %w", err)
//	}
//
//	// Now resolve tokens
//	ctx := context.Background()
//	bgColor, err := registry.ResolveToken(ctx, schema.TokenReference("semantic.colors.background.default"))
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Background color: %s\n", bgColor) // Output: "#f5f5f5"
func (tr *TokenRegistry) SetTokens(tokens *DesignTokens) error {
	if tokens == nil {
		return NewValidationError("tokens_nil", "tokens cannot be nil")
	}

	// Validate the tokens before setting them
	if err := tr.resolver.ValidateReferences(tokens); err != nil {
		return WrapError(err, "token_validation_failed", "token validation failed")
	}

	tr.mu.Lock()
	defer tr.mu.Unlock()

	tr.tokens = tokens
	tr.needsRecompile = true

	// Clear caches since tokens have changed
	tr.cache = make(map[string]string)
	tr.compiled.Clear()

	// Trigger recompilation
	tr.compileTokensUnsafe()

	return nil
}

// GetTokens returns a copy of current tokens (thread-safe).
func (tr *TokenRegistry) GetTokens() *DesignTokens {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	if tr.tokens == nil {
		return nil
	}

	// Return a deep copy to prevent external modification
	// For now, return the original (deep copy can be implemented later if needed)
	return tr.tokens
}

// ResolveToken resolves a single token reference to its actual value.
//
// Example:
//
//	registry := schema.GetDefaultRegistry()
//	ctx := context.Background()
//
//	// Resolve a primitive token
//	blueColor, err := registry.ResolveToken(ctx, schema.TokenReference("primitives.colors.blue.500"))
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Blue 500: %s\n", blueColor) // Output: "#3b82f6"
//
//	// Resolve a semantic token (references primitive)
//	primaryColor, err := registry.ResolveToken(ctx, schema.TokenReference("semantic.colors.interactive.primary"))
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Primary color: %s\n", primaryColor) // Output: "#3b82f6"
//
//	// Resolve a component token (references semantic)
//	buttonBg, err := registry.ResolveToken(ctx, schema.TokenReference("components.button.primary.background"))
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Button background: %s\n", buttonBg) // Output: "#3b82f6"
//
//	// Handle non-reference values (pass-through)
//	literal, err := registry.ResolveToken(ctx, schema.TokenReference("#ff0000"))
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Literal color: %s\n", literal) // Output: "#ff0000"
//
//	// Error handling for invalid tokens
//	_, err = registry.ResolveToken(ctx, schema.TokenReference("invalid.token.path"))
//	if err != nil {
//	    if schema.IsValidationError(err) {
//	        fmt.Printf("Token validation error: %s\n", err.Error())
//	    }
//	}
func (tr *TokenRegistry) ResolveToken(ctx context.Context, reference TokenReference) (string, error) {
	if !reference.IsReference() {
		return reference.String(), nil
	}

	path := reference.Path()

	// Check compiled map first (O(1) lookup)
	if compiled, exists := tr.compiled.Get(path); exists {
		return compiled, nil
	}

	// Check resolution cache
	tr.mu.RLock()
	if cached, exists := tr.cache[path]; exists {
		tr.mu.RUnlock()
		return cached, nil
	}
	tokens := tr.tokens
	needsRecompile := tr.needsRecompile
	tr.mu.RUnlock()

	// Recompile if needed
	if needsRecompile {
		tr.compileTokens()
		// Try compiled map again after recompilation
		if compiled, exists := tr.compiled.Get(path); exists {
			return compiled, nil
		}
	}

	if tokens == nil {
		return "", NewValidationError("tokens_unavailable", "no tokens available")
	}

	// Fallback to resolver (should be rare after compilation)
	resolved, err := tr.resolver.Resolve(ctx, reference, tokens)
	if err != nil {
		return "", err
	}

	// Cache the resolved value
	tr.mu.Lock()
	tr.cache[path] = resolved
	tr.mu.Unlock()

	// Add to compiled map for future O(1) access
	tr.compiled.Set(path, resolved)

	return resolved, nil
}

// ResolveAllTokens resolves all token references in the registry.
func (tr *TokenRegistry) ResolveAllTokens(ctx context.Context) (*DesignTokens, error) {
	tr.mu.RLock()
	tokens := tr.tokens
	tr.mu.RUnlock()

	if tokens == nil {
		return nil, NewValidationError("tokens_unavailable", "no tokens available")
	}

	return tr.resolver.ResolveAll(ctx, tokens)
}

// InvalidateCache clears the resolution cache.
func (tr *TokenRegistry) InvalidateCache() {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.cache = make(map[string]string)
}

// ValidateTokens validates the current token set for circular references and invalid paths.
func (tr *TokenRegistry) ValidateTokens() error {
	tr.mu.RLock()
	tokens := tr.tokens
	tr.mu.RUnlock()

	if tokens == nil {
		return NewValidationError("tokens_unavailable", "no tokens to validate")
	}

	return tr.resolver.ValidateReferences(tokens)
}

// compileTokens precompiles all token paths for O(1) lookups (thread-safe)
func (tr *TokenRegistry) compileTokens() {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.compileTokensUnsafe()
}

// compileTokensUnsafe precompiles all token paths (must hold lock)
func (tr *TokenRegistry) compileTokensUnsafe() {
	if tr.tokens == nil {
		return
	}

	// Clear existing compiled tokens
	tr.compiled.Clear()

	// Compile all token paths using reflection
	tr.compileTokenStruct(reflect.ValueOf(tr.tokens).Elem(), "")

	tr.needsRecompile = false
}

// compileTokenStruct recursively compiles token paths from a struct
func (tr *TokenRegistry) compileTokenStruct(structValue reflect.Value, basePath string) {
	structType := structValue.Type()

	for i := 0; i < structValue.NumField(); i++ {
		field := structType.Field(i)
		fieldValue := structValue.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Build field path
		fieldPath := field.Name
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			tagParts := strings.Split(jsonTag, ",")
			if len(tagParts) > 0 && tagParts[0] != "" && tagParts[0] != "-" {
				fieldPath = tagParts[0]
			}
		}

		fullPath := fieldPath
		if basePath != "" {
			fullPath = basePath + "." + fieldPath
		}

		// Handle different field types
		switch fieldValue.Kind() {
		case reflect.Ptr:
			if !fieldValue.IsNil() {
				tr.compileTokenStruct(fieldValue.Elem(), fullPath)
			}

		case reflect.Struct:
			tr.compileTokenStruct(fieldValue, fullPath)

		case reflect.String:
			// Store string values directly
			if fieldValue.String() != "" {
				tr.compiled.Set(fullPath, fieldValue.String())
			}

		case reflect.Interface:
			if fieldValue.CanInterface() {
				if str, ok := fieldValue.Interface().(string); ok && str != "" {
					tr.compiled.Set(fullPath, str)
				} else if tokenRef, ok := fieldValue.Interface().(TokenReference); ok {
					tr.compiled.Set(fullPath, tokenRef.String())
				}
			}
		}
	}
}

// GetCompiledSize returns the number of precompiled tokens
func (tr *TokenRegistry) GetCompiledSize() int {
	return tr.compiled.Size()
}

// GetDefaultTokens returns the default design token values.
// Following the design system specification.
func GetDefaultTokens() *DesignTokens {
	return &DesignTokens{
		Primitives: &PrimitiveTokens{
			Colors: &ColorPrimitives{
				Gray: &GrayScale{
					Scale50:  "hsl(0, 0%, 98%)",
					Scale100: "hsl(0, 0%, 96%)",
					Scale200: "hsl(0, 0%, 90%)",
					Scale300: "hsl(0, 0%, 83%)",
					Scale400: "hsl(0, 0%, 64%)",
					Scale500: "hsl(0, 0%, 50%)",
					Scale600: "hsl(0, 0%, 32%)",
					Scale700: "hsl(0, 0%, 21%)",
					Scale800: "hsl(0, 0%, 13%)",
					Scale900: "hsl(0, 0%, 9%)",
				},
				Blue: &ColorScale{
					Scale50:  "hsl(214, 100%, 97%)",
					Scale100: "hsl(214, 95%, 93%)",
					Scale200: "hsl(213, 97%, 87%)",
					Scale300: "hsl(212, 96%, 78%)",
					Scale400: "hsl(213, 94%, 68%)",
					Scale500: "hsl(217, 91%, 60%)", // Base blue
					Scale600: "hsl(221, 83%, 53%)",
					Scale700: "hsl(224, 76%, 48%)",
					Scale800: "hsl(226, 71%, 40%)",
					Scale900: "hsl(224, 64%, 33%)",
				},
				Purple: &ColorScale{
					Scale50:  "hsl(270, 100%, 98%)",
					Scale100: "hsl(269, 100%, 95%)",
					Scale200: "hsl(269, 100%, 92%)",
					Scale300: "hsl(269, 97%, 85%)",
					Scale400: "hsl(270, 95%, 75%)",
					Scale500: "hsl(270, 91%, 65%)", // Base purple
					Scale600: "hsl(271, 81%, 56%)",
					Scale700: "hsl(272, 72%, 47%)",
					Scale800: "hsl(272, 67%, 39%)",
					Scale900: "hsl(273, 66%, 32%)",
				},
				Cyan: &ColorScale{
					Scale50:  "hsl(183, 100%, 96%)",
					Scale100: "hsl(185, 96%, 90%)",
					Scale200: "hsl(186, 94%, 82%)",
					Scale300: "hsl(187, 92%, 69%)",
					Scale400: "hsl(188, 86%, 53%)",
					Scale500: "hsl(189, 94%, 43%)", // Base cyan
					Scale600: "hsl(192, 91%, 36%)",
					Scale700: "hsl(194, 74%, 31%)",
					Scale800: "hsl(195, 63%, 25%)",
					Scale900: "hsl(196, 64%, 24%)",
				},
				Green: &ColorScale{
					Scale50:  "hsl(138, 76%, 97%)",
					Scale100: "hsl(141, 84%, 93%)",
					Scale200: "hsl(141, 79%, 85%)",
					Scale300: "hsl(142, 77%, 73%)",
					Scale400: "hsl(142, 69%, 58%)",
					Scale500: "hsl(142, 71%, 45%)", // Base green
					Scale600: "hsl(142, 76%, 36%)",
					Scale700: "hsl(142, 72%, 29%)",
					Scale800: "hsl(143, 64%, 24%)",
					Scale900: "hsl(144, 61%, 20%)",
				},
				Red: &ColorScale{
					Scale50:  "hsl(0, 86%, 97%)",
					Scale100: "hsl(0, 93%, 94%)",
					Scale200: "hsl(0, 96%, 89%)",
					Scale300: "hsl(0, 94%, 82%)",
					Scale400: "hsl(0, 91%, 71%)",
					Scale500: "hsl(0, 84%, 60%)", // Base red
					Scale600: "hsl(0, 72%, 51%)",
					Scale700: "hsl(0, 74%, 42%)",
					Scale800: "hsl(0, 70%, 35%)",
					Scale900: "hsl(0, 63%, 31%)",
				},
				Yellow: &ColorScale{
					Scale50:  "hsl(55, 92%, 95%)",
					Scale100: "hsl(55, 97%, 88%)",
					Scale200: "hsl(53, 98%, 77%)",
					Scale300: "hsl(50, 98%, 64%)",
					Scale400: "hsl(48, 96%, 53%)",
					Scale500: "hsl(45, 93%, 47%)", // Base yellow
					Scale600: "hsl(41, 96%, 40%)",
					Scale700: "hsl(35, 91%, 33%)",
					Scale800: "hsl(32, 81%, 29%)",
					Scale900: "hsl(28, 73%, 26%)",
				},
				Sky: &ColorScale{
					Scale50:  "hsl(204, 100%, 97%)",
					Scale100: "hsl(204, 94%, 94%)",
					Scale200: "hsl(201, 94%, 86%)",
					Scale300: "hsl(199, 95%, 74%)",
					Scale400: "hsl(198, 93%, 60%)",
					Scale500: "hsl(199, 89%, 48%)", // Base sky
					Scale600: "hsl(200, 98%, 39%)",
					Scale700: "hsl(201, 96%, 32%)",
					Scale800: "hsl(201, 90%, 27%)",
					Scale900: "hsl(202, 80%, 24%)",
				},
				White: "hsl(0, 0%, 100%)",
				Black: "hsl(0, 0%, 0%)",
			},
			Spacing: &SpacingScale{
				None: "0",
				XS:   "0.25rem",
				SM:   "0.5rem",
				MD:   "1rem",
				LG:   "1.5rem",
				XL:   "2rem",
				XXL:  "3rem",
				XXXL: "4rem",
				Huge: "6rem",
			},
			Typography: &TypographyPrimitives{
				FontSizes: &FontSizeScale{
					XS:   "0.75rem",
					SM:   "0.875rem",
					Base: "1rem",
					LG:   "1.125rem",
					XL:   "1.25rem",
					XXL:  "1.5rem",
					XXXL: "1.875rem",
					Huge: "2.25rem",
				},
				FontWeights: &FontWeightScale{
					Thin:      "100",
					Light:     "300",
					Normal:    "400",
					Medium:    "500",
					Semibold:  "600",
					Bold:      "700",
					Extrabold: "800",
					Black:     "900",
				},
				LineHeights: &LineHeightScale{
					Tight:   "1.25",
					Normal:  "1.5",
					Relaxed: "1.75",
					Loose:   "2",
				},
				FontFamilies: &FontFamilyScale{
					Sans:  "ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, 'Noto Sans', sans-serif",
					Serif: "ui-serif, Georgia, Cambria, 'Times New Roman', Times, serif",
					Mono:  "ui-monospace, SFMono-Regular, 'SF Mono', Consolas, 'Liberation Mono', Menlo, monospace",
				},
				LetterSpacing: &LetterSpacingScale{
					Tight:  "-0.05em",
					Normal: "0",
					Wide:   "0.05em",
				},
			},
			Borders: &BorderPrimitives{
				Width: &BorderWidthScale{
					None:   "0",
					Thin:   "1px",
					Medium: "2px",
					Thick:  "4px",
				},
				Radius: &BorderRadiusScale{
					None: "0",
					SM:   "0.125rem",
					MD:   "0.25rem",
					LG:   "0.5rem",
					XL:   "1rem",
					Full: "9999px",
				},
				Style: &BorderStyleScale{
					Solid:  "solid",
					Dashed: "dashed",
					Dotted: "dotted",
					None:   "none",
				},
			},
			Shadows: &ShadowScale{
				None:  "none",
				SM:    "0 1px 2px 0 rgba(0, 0, 0, 0.05)",
				MD:    "0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)",
				LG:    "0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05)",
				XL:    "0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04)",
				XXL:   "0 25px 50px -12px rgba(0, 0, 0, 0.25)",
				Inner: "inset 0 2px 4px 0 rgba(0, 0, 0, 0.06)",
			},
			Animations: &AnimationPrimitives{
				Duration: &AnimationDurationScale{
					Fast:   "150ms",
					Normal: "300ms",
					Slow:   "500ms",
				},
				Easing: &AnimationEasingScale{
					Linear:    "linear",
					EaseIn:    "cubic-bezier(0.4, 0, 1, 1)",
					EaseOut:   "cubic-bezier(0, 0, 0.2, 1)",
					EaseInOut: "cubic-bezier(0.4, 0, 0.2, 1)",
				},
			},
			Sizes: &SizeScale{
				XS:  "1rem",
				SM:  "1.5rem",
				MD:  "2rem",
				LG:  "2.5rem",
				XL:  "3rem",
				XXL: "4rem",
			},
			ZIndex: &ZIndexScale{
				Dropdown: "1000",
				Sticky:   "1100",
				Fixed:    "1200",
				Modal:    "1300",
				Popover:  "1400",
				Tooltip:  "1500",
			},
		},
		Semantic: &SemanticTokens{
			Colors: &SemanticColors{
				Background: &BackgroundColors{
					Default:  TokenReference("primitives.colors.white"),
					Subtle:   TokenReference("primitives.colors.gray.50"),
					Emphasis: TokenReference("primitives.colors.gray.100"),
					Overlay:  TokenReference("rgba(0, 0, 0, 0.5)"),
				},
				Text: &TextColors{
					Default:   TokenReference("primitives.colors.gray.900"),
					Subtle:    TokenReference("primitives.colors.gray.600"),
					Disabled:  TokenReference("primitives.colors.gray.400"),
					Inverted:  TokenReference("primitives.colors.white"),
					Link:      TokenReference("primitives.colors.blue.500"),
					LinkHover: TokenReference("primitives.colors.blue.600"),
				},
				Border: &BorderColors{
					Default: TokenReference("primitives.colors.gray.300"),
					Focus:   TokenReference("primitives.colors.blue.500"),
					Strong:  TokenReference("primitives.colors.gray.400"),
					Subtle:  TokenReference("primitives.colors.gray.200"),
				},
				Interactive: &InteractiveColors{
					Primary: &InteractiveColorSet{
						Default: TokenReference("primitives.colors.blue.500"),
						Hover:   TokenReference("primitives.colors.blue.600"),
						Active:  TokenReference("primitives.colors.blue.700"),
						Focus:   TokenReference("primitives.colors.blue.500"),
					},
					Secondary: &InteractiveColorSet{
						Default: TokenReference("primitives.colors.gray.500"),
						Hover:   TokenReference("primitives.colors.gray.600"),
						Active:  TokenReference("primitives.colors.gray.700"),
						Focus:   TokenReference("primitives.colors.gray.500"),
					},
					Accent: &InteractiveColorSet{
						Default: TokenReference("primitives.colors.purple.500"),
						Hover:   TokenReference("primitives.colors.purple.600"),
						Active:  TokenReference("primitives.colors.purple.700"),
						Focus:   TokenReference("primitives.colors.purple.500"),
					},
				},
				Feedback: &FeedbackColors{
					Success: &FeedbackColorSet{
						Default: TokenReference("primitives.colors.green.500"),
						Subtle:  TokenReference("primitives.colors.green.50"),
						Strong:  TokenReference("primitives.colors.green.700"),
					},
					Error: &FeedbackColorSet{
						Default: TokenReference("primitives.colors.red.500"),
						Subtle:  TokenReference("primitives.colors.red.50"),
						Strong:  TokenReference("primitives.colors.red.700"),
					},
					Warning: &FeedbackColorSet{
						Default: TokenReference("primitives.colors.yellow.500"),
						Subtle:  TokenReference("primitives.colors.yellow.50"),
						Strong:  TokenReference("primitives.colors.yellow.700"),
					},
					Info: &FeedbackColorSet{
						Default: TokenReference("primitives.colors.sky.500"),
						Subtle:  TokenReference("primitives.colors.sky.50"),
						Strong:  TokenReference("primitives.colors.sky.700"),
					},
				},
			},
			Typography: &SemanticTypography{
				Headings: &HeadingTokens{
					H1: &TypographyToken{
						FontFamily:    TokenReference("primitives.typography.fontFamilies.sans"),
						FontSize:      TokenReference("primitives.typography.fontSizes.4xl"),
						FontWeight:    TokenReference("primitives.typography.fontWeights.bold"),
						LineHeight:    TokenReference("primitives.typography.lineHeights.tight"),
						LetterSpacing: TokenReference("primitives.typography.letterSpacing.tight"),
					},
					H2: &TypographyToken{
						FontFamily:    TokenReference("primitives.typography.fontFamilies.sans"),
						FontSize:      TokenReference("primitives.typography.fontSizes.3xl"),
						FontWeight:    TokenReference("primitives.typography.fontWeights.bold"),
						LineHeight:    TokenReference("primitives.typography.lineHeights.tight"),
						LetterSpacing: TokenReference("primitives.typography.letterSpacing.tight"),
					},
					H3: &TypographyToken{
						FontFamily:    TokenReference("primitives.typography.fontFamilies.sans"),
						FontSize:      TokenReference("primitives.typography.fontSizes.2xl"),
						FontWeight:    TokenReference("primitives.typography.fontWeights.semibold"),
						LineHeight:    TokenReference("primitives.typography.lineHeights.tight"),
						LetterSpacing: TokenReference("primitives.typography.letterSpacing.normal"),
					},
					H4: &TypographyToken{
						FontFamily:    TokenReference("primitives.typography.fontFamilies.sans"),
						FontSize:      TokenReference("primitives.typography.fontSizes.xl"),
						FontWeight:    TokenReference("primitives.typography.fontWeights.semibold"),
						LineHeight:    TokenReference("primitives.typography.lineHeights.normal"),
						LetterSpacing: TokenReference("primitives.typography.letterSpacing.normal"),
					},
					H5: &TypographyToken{
						FontFamily:    TokenReference("primitives.typography.fontFamilies.sans"),
						FontSize:      TokenReference("primitives.typography.fontSizes.lg"),
						FontWeight:    TokenReference("primitives.typography.fontWeights.semibold"),
						LineHeight:    TokenReference("primitives.typography.lineHeights.normal"),
						LetterSpacing: TokenReference("primitives.typography.letterSpacing.normal"),
					},
					H6: &TypographyToken{
						FontFamily:    TokenReference("primitives.typography.fontFamilies.sans"),
						FontSize:      TokenReference("primitives.typography.fontSizes.base"),
						FontWeight:    TokenReference("primitives.typography.fontWeights.semibold"),
						LineHeight:    TokenReference("primitives.typography.lineHeights.normal"),
						LetterSpacing: TokenReference("primitives.typography.letterSpacing.normal"),
					},
				},
				Body: &BodyTokens{
					Large: &TypographyToken{
						FontFamily:    TokenReference("primitives.typography.fontFamilies.sans"),
						FontSize:      TokenReference("primitives.typography.fontSizes.lg"),
						FontWeight:    TokenReference("primitives.typography.fontWeights.normal"),
						LineHeight:    TokenReference("primitives.typography.lineHeights.relaxed"),
						LetterSpacing: TokenReference("primitives.typography.letterSpacing.normal"),
					},
					Default: &TypographyToken{
						FontFamily:    TokenReference("primitives.typography.fontFamilies.sans"),
						FontSize:      TokenReference("primitives.typography.fontSizes.base"),
						FontWeight:    TokenReference("primitives.typography.fontWeights.normal"),
						LineHeight:    TokenReference("primitives.typography.lineHeights.normal"),
						LetterSpacing: TokenReference("primitives.typography.letterSpacing.normal"),
					},
					Small: &TypographyToken{
						FontFamily:    TokenReference("primitives.typography.fontFamilies.sans"),
						FontSize:      TokenReference("primitives.typography.fontSizes.sm"),
						FontWeight:    TokenReference("primitives.typography.fontWeights.normal"),
						LineHeight:    TokenReference("primitives.typography.lineHeights.normal"),
						LetterSpacing: TokenReference("primitives.typography.letterSpacing.normal"),
					},
				},
				Labels: &LabelTokens{
					Large: &TypographyToken{
						FontFamily:    TokenReference("primitives.typography.fontFamilies.sans"),
						FontSize:      TokenReference("primitives.typography.fontSizes.base"),
						FontWeight:    TokenReference("primitives.typography.fontWeights.medium"),
						LineHeight:    TokenReference("primitives.typography.lineHeights.normal"),
						LetterSpacing: TokenReference("primitives.typography.letterSpacing.normal"),
					},
					Default: &TypographyToken{
						FontFamily:    TokenReference("primitives.typography.fontFamilies.sans"),
						FontSize:      TokenReference("primitives.typography.fontSizes.sm"),
						FontWeight:    TokenReference("primitives.typography.fontWeights.medium"),
						LineHeight:    TokenReference("primitives.typography.lineHeights.normal"),
						LetterSpacing: TokenReference("primitives.typography.letterSpacing.normal"),
					},
					Small: &TypographyToken{
						FontFamily:    TokenReference("primitives.typography.fontFamilies.sans"),
						FontSize:      TokenReference("primitives.typography.fontSizes.xs"),
						FontWeight:    TokenReference("primitives.typography.fontWeights.medium"),
						LineHeight:    TokenReference("primitives.typography.lineHeights.normal"),
						LetterSpacing: TokenReference("primitives.typography.letterSpacing.normal"),
					},
				},
				Captions: &CaptionTokens{
					Default: &TypographyToken{
						FontFamily:    TokenReference("primitives.typography.fontFamilies.sans"),
						FontSize:      TokenReference("primitives.typography.fontSizes.sm"),
						FontWeight:    TokenReference("primitives.typography.fontWeights.normal"),
						LineHeight:    TokenReference("primitives.typography.lineHeights.normal"),
						LetterSpacing: TokenReference("primitives.typography.letterSpacing.normal"),
					},
					Small: &TypographyToken{
						FontFamily:    TokenReference("primitives.typography.fontFamilies.sans"),
						FontSize:      TokenReference("primitives.typography.fontSizes.xs"),
						FontWeight:    TokenReference("primitives.typography.fontWeights.normal"),
						LineHeight:    TokenReference("primitives.typography.lineHeights.normal"),
						LetterSpacing: TokenReference("primitives.typography.letterSpacing.normal"),
					},
				},
				Code: &CodeTokens{
					Inline: &TypographyToken{
						FontFamily:    TokenReference("primitives.typography.fontFamilies.mono"),
						FontSize:      TokenReference("primitives.typography.fontSizes.sm"),
						FontWeight:    TokenReference("primitives.typography.fontWeights.normal"),
						LineHeight:    TokenReference("primitives.typography.lineHeights.normal"),
						LetterSpacing: TokenReference("primitives.typography.letterSpacing.normal"),
					},
					Block: &TypographyToken{
						FontFamily:    TokenReference("primitives.typography.fontFamilies.mono"),
						FontSize:      TokenReference("primitives.typography.fontSizes.sm"),
						FontWeight:    TokenReference("primitives.typography.fontWeights.normal"),
						LineHeight:    TokenReference("primitives.typography.lineHeights.relaxed"),
						LetterSpacing: TokenReference("primitives.typography.letterSpacing.normal"),
					},
				},
			},
			Spacing: &SemanticSpacing{
				Component: &ComponentSpacing{
					Tight:   TokenReference("primitives.spacing.xs"),
					Default: TokenReference("primitives.spacing.md"),
					Loose:   TokenReference("primitives.spacing.lg"),
				},
				Layout: &LayoutSpacing{
					Section: TokenReference("primitives.spacing.2xl"),
					Page:    TokenReference("primitives.spacing.4xl"),
				},
			},
			Interactive: &SemanticInteractive{
				BorderRadius: &InteractiveBorderRadius{
					Small:   TokenReference("primitives.borders.radius.sm"),
					Default: TokenReference("primitives.borders.radius.md"),
					Large:   TokenReference("primitives.borders.radius.lg"),
				},
				Shadow: &InteractiveShadow{
					Default: TokenReference("primitives.shadows.sm"),
					Hover:   TokenReference("primitives.shadows.md"),
					Focus:   TokenReference("primitives.shadows.lg"),
				},
			},
		},
		Components: &ComponentTokens{
			Button: &ButtonTokens{
				Primary: &ButtonVariantTokens{
					Background:       TokenReference("semantic.colors.interactive.primary.default"),
					BackgroundHover:  TokenReference("semantic.colors.interactive.primary.hover"),
					BackgroundActive: TokenReference("semantic.colors.interactive.primary.active"),
					Color:            TokenReference("semantic.colors.text.inverted"),
					ColorHover:       TokenReference("semantic.colors.text.inverted"),
					Border:           TokenReference("semantic.colors.interactive.primary.default"),
					BorderHover:      TokenReference("semantic.colors.interactive.primary.hover"),
					BorderRadius:     TokenReference("semantic.interactive.borderRadius.default"),
					Padding:          TokenReference("semantic.spacing.component.default"),
					FontWeight:       TokenReference("primitives.typography.fontWeights.medium"),
					Shadow:           TokenReference("semantic.interactive.shadow.default"),
					ShadowHover:      TokenReference("semantic.interactive.shadow.hover"),
				},
				Secondary: &ButtonVariantTokens{
					Background:       TokenReference("semantic.colors.background.default"),
					BackgroundHover:  TokenReference("semantic.colors.background.subtle"),
					BackgroundActive: TokenReference("semantic.colors.background.emphasis"),
					Color:            TokenReference("semantic.colors.text.default"),
					ColorHover:       TokenReference("semantic.colors.text.default"),
					Border:           TokenReference("semantic.colors.border.default"),
					BorderHover:      TokenReference("semantic.colors.border.strong"),
					BorderRadius:     TokenReference("semantic.interactive.borderRadius.default"),
					Padding:          TokenReference("semantic.spacing.component.default"),
					FontWeight:       TokenReference("primitives.typography.fontWeights.medium"),
					Shadow:           TokenReference("semantic.interactive.shadow.default"),
					ShadowHover:      TokenReference("semantic.interactive.shadow.hover"),
				},
				Outline: &ButtonVariantTokens{
					Background:       TokenReference("transparent"),
					BackgroundHover:  TokenReference("semantic.colors.interactive.primary.default"),
					BackgroundActive: TokenReference("semantic.colors.interactive.primary.active"),
					Color:            TokenReference("semantic.colors.interactive.primary.default"),
					ColorHover:       TokenReference("semantic.colors.text.inverted"),
					Border:           TokenReference("semantic.colors.interactive.primary.default"),
					BorderHover:      TokenReference("semantic.colors.interactive.primary.hover"),
					BorderRadius:     TokenReference("semantic.interactive.borderRadius.default"),
					Padding:          TokenReference("semantic.spacing.component.default"),
					FontWeight:       TokenReference("primitives.typography.fontWeights.medium"),
					Shadow:           TokenReference("none"),
					ShadowHover:      TokenReference("semantic.interactive.shadow.hover"),
				},
				Ghost: &ButtonVariantTokens{
					Background:       TokenReference("transparent"),
					BackgroundHover:  TokenReference("semantic.colors.background.subtle"),
					BackgroundActive: TokenReference("semantic.colors.background.emphasis"),
					Color:            TokenReference("semantic.colors.text.default"),
					ColorHover:       TokenReference("semantic.colors.text.default"),
					Border:           TokenReference("transparent"),
					BorderHover:      TokenReference("transparent"),
					BorderRadius:     TokenReference("semantic.interactive.borderRadius.default"),
					Padding:          TokenReference("semantic.spacing.component.default"),
					FontWeight:       TokenReference("primitives.typography.fontWeights.medium"),
					Shadow:           TokenReference("none"),
					ShadowHover:      TokenReference("none"),
				},
				Destructive: &ButtonVariantTokens{
					Background:       TokenReference("semantic.colors.feedback.error.default"),
					BackgroundHover:  TokenReference("semantic.colors.feedback.error.strong"),
					BackgroundActive: TokenReference("semantic.colors.feedback.error.strong"),
					Color:            TokenReference("semantic.colors.text.inverted"),
					ColorHover:       TokenReference("semantic.colors.text.inverted"),
					Border:           TokenReference("semantic.colors.feedback.error.default"),
					BorderHover:      TokenReference("semantic.colors.feedback.error.strong"),
					BorderRadius:     TokenReference("semantic.interactive.borderRadius.default"),
					Padding:          TokenReference("semantic.spacing.component.default"),
					FontWeight:       TokenReference("primitives.typography.fontWeights.medium"),
					Shadow:           TokenReference("semantic.interactive.shadow.default"),
					ShadowHover:      TokenReference("semantic.interactive.shadow.hover"),
				},
			},
			Input: &InputTokens{
				Background:      TokenReference("semantic.colors.background.default"),
				BackgroundFocus: TokenReference("semantic.colors.background.default"),
				Border:          TokenReference("semantic.colors.border.default"),
				BorderFocus:     TokenReference("semantic.colors.border.focus"),
				BorderError:     TokenReference("semantic.colors.feedback.error.default"),
				BorderRadius:    TokenReference("semantic.interactive.borderRadius.default"),
				Padding:         TokenReference("semantic.spacing.component.default"),
				Color:           TokenReference("semantic.colors.text.default"),
				Placeholder:     TokenReference("semantic.colors.text.subtle"),
			},
			Card: &CardTokens{
				Background:   TokenReference("semantic.colors.background.default"),
				Border:       TokenReference("semantic.colors.border.subtle"),
				BorderRadius: TokenReference("semantic.interactive.borderRadius.large"),
				Shadow:       TokenReference("semantic.interactive.shadow.default"),
				Padding:      TokenReference("semantic.spacing.component.loose"),
			},
			Modal: &ModalTokens{
				Background:   TokenReference("semantic.colors.background.default"),
				Overlay:      TokenReference("semantic.colors.background.overlay"),
				Border:       TokenReference("semantic.colors.border.subtle"),
				BorderRadius: TokenReference("semantic.interactive.borderRadius.large"),
				Shadow:       TokenReference("semantic.interactive.shadow.focus"),
				Padding:      TokenReference("semantic.spacing.component.loose"),
			},
			Form: &FormTokens{
				Background:   TokenReference("semantic.colors.background.default"),
				Padding:      TokenReference("semantic.spacing.component.loose"),
				BorderRadius: TokenReference("semantic.interactive.borderRadius.default"),
				Shadow:       TokenReference("none"),
				Spacing:      TokenReference("semantic.spacing.component.default"),
			},
			Table: &TableTokens{
				Background:       TokenReference("semantic.colors.background.default"),
				BackgroundHover:  TokenReference("semantic.colors.background.subtle"),
				Border:           TokenReference("semantic.colors.border.subtle"),
				HeaderBackground: TokenReference("semantic.colors.background.emphasis"),
				HeaderColor:      TokenReference("semantic.colors.text.default"),
				Padding:          TokenReference("semantic.spacing.component.default"),
			},
			Navigation: &NavigationTokens{
				Background:       TokenReference("semantic.colors.background.default"),
				BackgroundHover:  TokenReference("semantic.colors.background.subtle"),
				BackgroundActive: TokenReference("semantic.colors.background.emphasis"),
				Color:            TokenReference("semantic.colors.text.default"),
				ColorHover:       TokenReference("semantic.colors.text.default"),
				ColorActive:      TokenReference("semantic.colors.interactive.primary.default"),
				Border:           TokenReference("semantic.colors.border.subtle"),
				Padding:          TokenReference("semantic.spacing.component.default"),
			},
		},
	}
}

// MergeTokens merges two token sets (second overrides first).
func MergeTokens(base, override *DesignTokens) *DesignTokens {
	if base == nil {
		return override
	}
	if override == nil {
		return base
	}

	// For now, return override (deep merging can be implemented later if needed)
	return override
}

// ValidateTokenPath checks if a token path is valid.
func ValidateTokenPath(path string) error {
	ref := TokenReference("" + path + "")
	return ref.Validate()
}

// TokenPath creates a token path from components.
func TokenPath(parts ...string) string {
	return strings.Join(parts, ".")
}

// Global token registry instance
var (
	defaultTokenRegistry *TokenRegistry
	registryOnce         sync.Once
)

// GetDefaultRegistry returns the global token registry.
func GetDefaultRegistry() *TokenRegistry {
	registryOnce.Do(func() {
		defaultTokenRegistry = NewTokenRegistry()
	})
	return defaultTokenRegistry
}

// Convenience functions for common token access

// GetSpacing returns a spacing token value.
func GetSpacing(key string) string {
	registry := GetDefaultRegistry()
	tokenRef := TokenReference(fmt.Sprintf("primitives.spacing.%s", key))

	resolved, err := registry.ResolveToken(context.Background(), tokenRef)
	if err != nil {
		// Fallback to semantic spacing
		semanticRef := TokenReference(fmt.Sprintf("semantic.spacing.%s", key))
		if semanticResolved, semanticErr := registry.ResolveToken(context.Background(), semanticRef); semanticErr == nil {
			return semanticResolved
		}
		return "1rem" // Safe fallback
	}

	return resolved
}

// GetColor returns a color token value.
func GetColor(path string) string {
	registry := GetDefaultRegistry()

	// Support both full paths and shorthand
	var tokenRef TokenReference
	if strings.Contains(path, ".") {
		tokenRef = TokenReference(fmt.Sprintf("%s", path))
	} else {
		// Try common color paths
		tokenRef = TokenReference(fmt.Sprintf("colors.%s", path))
	}

	resolved, err := registry.ResolveToken(context.Background(), tokenRef)
	if err != nil {
		// Try semantic colors as fallback
		semanticRef := TokenReference(fmt.Sprintf("semantic.colors.%s", path))
		if semanticResolved, semanticErr := registry.ResolveToken(context.Background(), semanticRef); semanticErr == nil {
			return semanticResolved
		}
		return "#000000" // Safe fallback
	}

	return resolved
}

// GetFontSize returns a font size token value.
func GetFontSize(key string) string {
	registry := GetDefaultRegistry()
	tokenRef := TokenReference(fmt.Sprintf("primitives.typography.fontSizes.%s", key))

	resolved, err := registry.ResolveToken(context.Background(), tokenRef)
	if err != nil {
		// Fallback to semantic typography
		semanticRef := TokenReference(fmt.Sprintf("semantic.typography.%s", key))
		if semanticResolved, semanticErr := registry.ResolveToken(context.Background(), semanticRef); semanticErr == nil {
			return semanticResolved
		}
		return "1rem" // Safe fallback
	}

	return resolved
}

// GetShadow returns a shadow token value.
func GetShadow(key string) string {
	registry := GetDefaultRegistry()
	tokenRef := TokenReference(fmt.Sprintf("primitives.shadows.%s", key))

	resolved, err := registry.ResolveToken(context.Background(), tokenRef)
	if err != nil {
		// Fallback to semantic shadows
		semanticRef := TokenReference(fmt.Sprintf("semantic.interactive.shadow.%s", key))
		if semanticResolved, semanticErr := registry.ResolveToken(context.Background(), semanticRef); semanticErr == nil {
			return semanticResolved
		}
		return "none" // Safe fallback
	}

	return resolved
}

// GetBorderRadius returns a border radius token value.
func GetBorderRadius(key string) string {
	registry := GetDefaultRegistry()
	tokenRef := TokenReference(fmt.Sprintf("primitives.borders.radius.%s", key))

	resolved, err := registry.ResolveToken(context.Background(), tokenRef)
	if err != nil {
		// Fallback to semantic border radius
		semanticRef := TokenReference(fmt.Sprintf("semantic.interactive.borderRadius.%s", key))
		if semanticResolved, semanticErr := registry.ResolveToken(context.Background(), semanticRef); semanticErr == nil {
			return semanticResolved
		}
		return "4px" // Safe fallback
	}

	return resolved
}
