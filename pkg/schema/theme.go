package schema

import (
	"encoding/json"
	"fmt"
	"sync"
)

// Theme represents the comprehensive theme configuration for a schema
// Integrates with the design token system and provides overrides for all components
type Theme struct {
	// Identity
	Name        string `json:"name" validate:"required" example:"corporate-blue"`
	Version     string `json:"version,omitempty" validate:"semver" example:"1.0.0"`
	Description string `json:"description,omitempty"`
	Author      string `json:"author,omitempty"`

	// Design tokens - base values
	Tokens *DesignTokens `json:"tokens,omitempty"` // Override default tokens

	// Component-specific theming
	Form    *FormTheme    `json:"form,omitempty"`    // Form container theme
	Field   *FieldTheme   `json:"field,omitempty"`   // Default field theme
	Button  *ButtonTheme  `json:"button,omitempty"`  // Button/action theme
	Layout  *LayoutTheme  `json:"layout,omitempty"`  // Layout theme
	Section *SectionTheme `json:"section,omitempty"` // Section theme
	Tab     *TabTheme     `json:"tab,omitempty"`     // Tab theme
	Step    *StepTheme    `json:"step,omitempty"`    // Step theme

	// Global overrides
	Colors      map[string]string `json:"colors,omitempty"`      // Color overrides
	Fonts       map[string]string `json:"fonts,omitempty"`       // Font overrides
	Breakpoints map[string]string `json:"breakpoints,omitempty"` // Breakpoint definitions
	CustomCSS   string            `json:"customCSS,omitempty"`   // Custom CSS
	CustomJS    string            `json:"customJS,omitempty"`    // Custom JavaScript

	// Dark mode support
	DarkMode *DarkModeConfig `json:"darkMode,omitempty"` // Dark mode configuration

	// Accessibility
	Accessibility *AccessibilityConfig `json:"accessibility,omitempty"` // Accessibility settings

	// Metadata
	Meta *ThemeMeta `json:"meta,omitempty"` // Theme metadata
}

// FormTheme defines theme for the form container
type FormTheme struct {
	Background   string            `json:"background,omitempty"`
	Padding      string            `json:"padding,omitempty"`
	MaxWidth     string            `json:"maxWidth,omitempty"`
	BorderRadius string            `json:"borderRadius,omitempty"`
	Shadow       string            `json:"shadow,omitempty"`
	Border       string            `json:"border,omitempty"`
	Colors       map[string]string `json:"colors,omitempty"`
	CustomCSS    string            `json:"customCSS,omitempty"`
}

// FieldTheme defines default theme for all fields
type FieldTheme struct {
	// Input styling
	Background      string `json:"background,omitempty"`
	BackgroundFocus string `json:"backgroundFocus,omitempty"`
	Border          string `json:"border,omitempty"`
	BorderFocus     string `json:"borderFocus,omitempty"`
	BorderError     string `json:"borderError,omitempty"`
	BorderRadius    string `json:"borderRadius,omitempty"`
	Padding         string `json:"padding,omitempty"`
	FontSize        string `json:"fontSize,omitempty"`
	FontWeight      string `json:"fontWeight,omitempty"`

	// Label styling
	LabelFontSize   string `json:"labelFontSize,omitempty"`
	LabelFontWeight string `json:"labelFontWeight,omitempty"`
	LabelColor      string `json:"labelColor,omitempty"`
	LabelMargin     string `json:"labelMargin,omitempty"`

	// Helper text styling
	HelperFontSize string `json:"helperFontSize,omitempty"`
	HelperColor    string `json:"helperColor,omitempty"`

	// Error styling
	ErrorColor    string `json:"errorColor,omitempty"`
	ErrorFontSize string `json:"errorFontSize,omitempty"`

	// Success styling
	SuccessColor  string `json:"successColor,omitempty"`
	SuccessBorder string `json:"successBorder,omitempty"`

	// Disabled styling
	DisabledBackground string `json:"disabledBackground,omitempty"`
	DisabledColor      string `json:"disabledColor,omitempty"`
	DisabledCursor     string `json:"disabledCursor,omitempty"`

	// Custom overrides
	Colors    map[string]string `json:"colors,omitempty"`
	CustomCSS string            `json:"customCSS,omitempty"`
}

// ButtonTheme defines theme for buttons and actions
type ButtonTheme struct {
	// Primary variant
	PrimaryBackground      string `json:"primaryBackground,omitempty"`
	PrimaryBackgroundHover string `json:"primaryBackgroundHover,omitempty"`
	PrimaryColor           string `json:"primaryColor,omitempty"`
	PrimaryBorder          string `json:"primaryBorder,omitempty"`

	// Secondary variant
	SecondaryBackground      string `json:"secondaryBackground,omitempty"`
	SecondaryBackgroundHover string `json:"secondaryBackgroundHover,omitempty"`
	SecondaryColor           string `json:"secondaryColor,omitempty"`
	SecondaryBorder          string `json:"secondaryBorder,omitempty"`

	// Outline variant
	OutlineBackground      string `json:"outlineBackground,omitempty"`
	OutlineBackgroundHover string `json:"outlineBackgroundHover,omitempty"`
	OutlineColor           string `json:"outlineColor,omitempty"`
	OutlineBorder          string `json:"outlineBorder,omitempty"`

	// Destructive variant
	DestructiveBackground      string `json:"destructiveBackground,omitempty"`
	DestructiveBackgroundHover string `json:"destructiveBackgroundHover,omitempty"`
	DestructiveColor           string `json:"destructiveColor,omitempty"`
	DestructiveBorder          string `json:"destructiveBorder,omitempty"`

	// Ghost variant
	GhostBackground      string `json:"ghostBackground,omitempty"`
	GhostBackgroundHover string `json:"ghostBackgroundHover,omitempty"`
	GhostColor           string `json:"ghostColor,omitempty"`

	// Common properties
	BorderRadius string `json:"borderRadius,omitempty"`
	FontWeight   string `json:"fontWeight,omitempty"`
	Padding      string `json:"padding,omitempty"`
	Shadow       string `json:"shadow,omitempty"`
	ShadowHover  string `json:"shadowHover,omitempty"`
	Transition   string `json:"transition,omitempty"`

	// Sizes
	SizeSmall  *ButtonSizeTheme `json:"sizeSmall,omitempty"`
	SizeMedium *ButtonSizeTheme `json:"sizeMedium,omitempty"`
	SizeLarge  *ButtonSizeTheme `json:"sizeLarge,omitempty"`

	// Custom overrides
	Colors    map[string]string `json:"colors,omitempty"`
	CustomCSS string            `json:"customCSS,omitempty"`
}

// ButtonSizeTheme defines size-specific button styling
type ButtonSizeTheme struct {
	Padding  string `json:"padding,omitempty"`
	FontSize string `json:"fontSize,omitempty"`
	Height   string `json:"height,omitempty"`
	MinWidth string `json:"minWidth,omitempty"`
	IconSize string `json:"iconSize,omitempty"`
}

// DarkModeConfig defines dark mode theme configuration
type DarkModeConfig struct {
	Enabled   bool              `json:"enabled"`                                              // Enable dark mode
	Default   bool              `json:"default,omitempty"`                                    // Use dark mode by default
	Toggle    bool              `json:"toggle,omitempty"`                                     // Allow user to toggle
	Strategy  string            `json:"strategy,omitempty" validate:"oneof=class media auto"` // Detection strategy
	Colors    map[string]string `json:"colors,omitempty"`                                     // Dark mode color overrides
	CustomCSS string            `json:"customCSS,omitempty"`                                  // Dark mode specific CSS
}

// AccessibilityConfig defines accessibility settings
type AccessibilityConfig struct {
	// ARIA
	AutoARIA        bool   `json:"autoAria,omitempty"`        // Auto-generate ARIA attributes
	AriaLive        string `json:"ariaLive,omitempty"`        // ARIA live region
	AriaDescribedBy bool   `json:"ariaDescribedBy,omitempty"` // Auto-link descriptions

	// Keyboard navigation
	KeyboardNav        bool `json:"keyboardNav,omitempty"`        // Enable keyboard navigation
	FocusIndicator     bool `json:"focusIndicator,omitempty"`     // Show focus indicators
	SkipLinks          bool `json:"skipLinks,omitempty"`          // Add skip links
	TabIndexManagement bool `json:"tabIndexManagement,omitempty"` // Manage tab indices

	// Screen reader
	ScreenReaderOnly  bool `json:"screenReaderOnly,omitempty"`  // Screen reader optimizations
	LiveAnnouncements bool `json:"liveAnnouncements,omitempty"` // Announce changes

	// Contrast
	HighContrast     bool    `json:"highContrast,omitempty"`     // High contrast mode
	MinContrastRatio float64 `json:"minContrastRatio,omitempty"` // Minimum contrast ratio

	// Focus management
	FocusOutlineColor string `json:"focusOutlineColor,omitempty"` // Focus outline color
	FocusOutlineWidth string `json:"focusOutlineWidth,omitempty"` // Focus outline width
	FocusOutlineStyle string `json:"focusOutlineStyle,omitempty"` // Focus outline style

	// Motion
	ReducedMotion bool `json:"reducedMotion,omitempty"` // Respect prefers-reduced-motion
}

// ThemeMeta contains theme metadata
type ThemeMeta struct {
	Tags       []string       `json:"tags,omitempty"`
	License    string         `json:"license,omitempty"`
	Repository string         `json:"repository,omitempty" validate:"url"`
	Homepage   string         `json:"homepage,omitempty" validate:"url"`
	Preview    string         `json:"preview,omitempty" validate:"url"` // Preview image URL
	CustomData map[string]any `json:"customData,omitempty"`
}

// ThemeRegistry manages theme storage and retrieval with thread safety
type ThemeRegistry struct {
	themes map[string]*Theme
	mu     sync.RWMutex
}

// NewThemeRegistry creates a new theme registry
func NewThemeRegistry() *ThemeRegistry {
	return &ThemeRegistry{
		themes: make(map[string]*Theme),
	}
}

// Register registers a new theme
func (tr *ThemeRegistry) Register(theme *Theme) error {
	if theme.Name == "" {
		return NewValidationError("theme_name_required", "theme name is required")
	}

	tr.mu.Lock()
	defer tr.mu.Unlock()

	tr.themes[theme.Name] = theme
	return nil
}

// Get retrieves a theme by name
func (tr *ThemeRegistry) Get(name string) (*Theme, error) {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	theme, exists := tr.themes[name]
	if !exists {
		return nil, NewValidationError("theme_not_found", fmt.Sprintf("theme not found: %s", name))
	}

	return theme, nil
}

// List returns all registered themes
func (tr *ThemeRegistry) List() []*Theme {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	themes := make([]*Theme, 0, len(tr.themes))
	for _, theme := range tr.themes {
		themes = append(themes, theme)
	}

	return themes
}

// Unregister removes a theme
func (tr *ThemeRegistry) Unregister(name string) error {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	if _, exists := tr.themes[name]; !exists {
		return NewValidationError("theme_not_found", fmt.Sprintf("theme not found: %s", name))
	}

	delete(tr.themes, name)
	return nil
}

// Update updates an existing theme
func (tr *ThemeRegistry) Update(theme *Theme) error {
	if theme.Name == "" {
		return NewValidationError("theme_name_required", "theme name is required")
	}

	tr.mu.Lock()
	defer tr.mu.Unlock()

	if _, exists := tr.themes[theme.Name]; !exists {
		return NewValidationError("theme_not_found", fmt.Sprintf("theme not found: %s", theme.Name))
	}

	tr.themes[theme.Name] = theme
	return nil
}

// Exists checks if a theme exists
func (tr *ThemeRegistry) Exists(name string) bool {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	_, exists := tr.themes[name]
	return exists
}

// Clone creates a copy of a theme
func (tr *ThemeRegistry) Clone(name, newName string) error {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	original, exists := tr.themes[name]
	if !exists {
		return NewValidationError("theme_not_found", fmt.Sprintf("theme not found: %s", name))
	}

	// Deep copy using JSON
	data, _ := json.Marshal(original)
	var cloned Theme
	json.Unmarshal(data, &cloned)
	cloned.Name = newName

	tr.themes[newName] = &cloned
	return nil
}

// Global theme registry
var (
	globalThemeRegistry *ThemeRegistry
	themeRegistryOnce   sync.Once
)

// GetGlobalThemeRegistry returns the global theme registry
func GetGlobalThemeRegistry() *ThemeRegistry {
	themeRegistryOnce.Do(func() {
		globalThemeRegistry = NewThemeRegistry()
		// Register default themes
		registerDefaultThemes(globalThemeRegistry)
	})
	return globalThemeRegistry
}

// ApplyTheme applies a theme to a schema
func (s *Schema) ApplyTheme(theme *Theme) {
	if theme == nil {
		return
	}

	// Apply theme to layout
	if s.Layout != nil && theme.Layout != nil {
		s.Layout.Theme = theme.Layout
	}

	// Apply theme to actions
	for i := range s.Actions {
		if theme.Button != nil {
			if s.Actions[i].Theme == nil {
				s.Actions[i].Theme = &ActionTheme{}
			}
			// Map ButtonTheme to ActionTheme
			s.Actions[i].Theme.Colors = theme.Button.Colors
			s.Actions[i].Theme.BorderRadius = theme.Button.BorderRadius
			s.Actions[i].Theme.CustomCSS = theme.Button.CustomCSS
		}
	}

	// Apply global colors if present
	if len(theme.Colors) > 0 {
		// Colors will be resolved during rendering
	}

	// Apply custom CSS
	if theme.CustomCSS != "" {
		// Custom CSS will be injected during rendering
	}
}

// GetTheme returns the theme from a theme name
func (s *Schema) GetTheme(themeName string) (*Theme, error) {
	registry := GetGlobalThemeRegistry()
	return registry.Get(themeName)
}

// Theme builder for fluent interface
type ThemeBuilder struct {
	theme *Theme
}

// NewTheme starts building a theme
func NewTheme(name string) *ThemeBuilder {
	return &ThemeBuilder{
		theme: &Theme{
			Name:   name,
			Colors: make(map[string]string),
			Fonts:  make(map[string]string),
		},
	}
}

// WithDescription sets the description
func (tb *ThemeBuilder) WithDescription(description string) *ThemeBuilder {
	tb.theme.Description = description
	return tb
}

// WithAuthor sets the author
func (tb *ThemeBuilder) WithAuthor(author string) *ThemeBuilder {
	tb.theme.Author = author
	return tb
}

// WithTokens sets the design tokens
func (tb *ThemeBuilder) WithTokens(tokens *DesignTokens) *ThemeBuilder {
	tb.theme.Tokens = tokens
	return tb
}

// WithFormTheme sets the form theme
func (tb *ThemeBuilder) WithFormTheme(form *FormTheme) *ThemeBuilder {
	tb.theme.Form = form
	return tb
}

// WithFieldTheme sets the field theme
func (tb *ThemeBuilder) WithFieldTheme(field *FieldTheme) *ThemeBuilder {
	tb.theme.Field = field
	return tb
}

// WithButtonTheme sets the button theme
func (tb *ThemeBuilder) WithButtonTheme(button *ButtonTheme) *ThemeBuilder {
	tb.theme.Button = button
	return tb
}

// WithColor adds a color override
func (tb *ThemeBuilder) WithColor(key, value string) *ThemeBuilder {
	tb.theme.Colors[key] = value
	return tb
}

// WithFont adds a font override
func (tb *ThemeBuilder) WithFont(key, value string) *ThemeBuilder {
	tb.theme.Fonts[key] = value
	return tb
}

// WithDarkMode enables dark mode
func (tb *ThemeBuilder) WithDarkMode(config *DarkModeConfig) *ThemeBuilder {
	tb.theme.DarkMode = config
	return tb
}

// WithAccessibility sets accessibility config
func (tb *ThemeBuilder) WithAccessibility(config *AccessibilityConfig) *ThemeBuilder {
	tb.theme.Accessibility = config
	return tb
}

// WithCustomCSS adds custom CSS
func (tb *ThemeBuilder) WithCustomCSS(css string) *ThemeBuilder {
	tb.theme.CustomCSS = css
	return tb
}

// Build returns the constructed theme
func (tb *ThemeBuilder) Build() *Theme {
	return tb.theme
}

// Register registers the theme to the global registry
func (tb *ThemeBuilder) Register() error {
	registry := GetGlobalThemeRegistry()
	return registry.Register(tb.theme)
}

// Predefined themes

// registerDefaultThemes registers default themes
func registerDefaultThemes(registry *ThemeRegistry) {
	// Light theme (default)
	lightTheme := NewTheme("light").
		WithDescription("Clean light theme").
		WithAuthor("Schema System").
		WithTokens(GetDefaultTokens()).
		Build()
	registry.Register(lightTheme)

	// Dark theme
	darkColors := make(map[string]string)
	darkColors["background"] = "hsl(0, 0%, 9%)"
	darkColors["text"] = "hsl(0, 0%, 98%)"
	darkColors["border"] = "hsl(0, 0%, 32%)"

	darkTheme := NewTheme("dark").
		WithDescription("Dark theme with high contrast").
		WithAuthor("Schema System").
		WithColor("background", "hsl(0, 0%, 9%)").
		WithColor("text", "hsl(0, 0%, 98%)").
		WithColor("border", "hsl(0, 0%, 32%)").
		WithDarkMode(&DarkModeConfig{
			Enabled: true,
			Default: true,
		}).
		Build()
	registry.Register(darkTheme)

	// Corporate theme
	corporateTheme := NewTheme("corporate").
		WithDescription("Professional corporate theme").
		WithAuthor("Schema System").
		WithColor("primary", "hsl(210, 100%, 45%)").
		WithColor("secondary", "hsl(210, 20%, 50%)").
		Build()
	registry.Register(corporateTheme)
}

// GetDefaultTheme returns the default light theme
func GetDefaultTheme() *Theme {
	registry := GetGlobalThemeRegistry()
	theme, _ := registry.Get("light")
	return theme
}

// Helper functions

// MergeThemes merges two themes (second overrides first)
func MergeThemes(base, override *Theme) *Theme {
	if override == nil {
		return base
	}
	if base == nil {
		return override
	}

	// Deep copy base
	data, _ := json.Marshal(base)
	var merged Theme
	json.Unmarshal(data, &merged)

	// Merge override properties
	if override.Tokens != nil {
		merged.Tokens = MergeTokens(merged.Tokens, override.Tokens)
	}

	if override.Form != nil {
		merged.Form = override.Form
	}
	if override.Field != nil {
		merged.Field = override.Field
	}
	if override.Button != nil {
		merged.Button = override.Button
	}
	if override.Layout != nil {
		merged.Layout = override.Layout
	}

	// Merge color maps
	for k, v := range override.Colors {
		merged.Colors[k] = v
	}
	for k, v := range override.Fonts {
		merged.Fonts[k] = v
	}

	if override.CustomCSS != "" {
		merged.CustomCSS += "\n" + override.CustomCSS
	}

	return &merged
}

// ValidateTheme validates a theme configuration
func ValidateTheme(theme *Theme) error {
	if theme.Name == "" {
		return NewValidationError("theme_name_required", "theme name is required")
	}

	// Validate version if present
	if theme.Version != "" {
		// TODO: Validate semver format
	}

	return nil
}

// ExportTheme exports a theme to JSON
func ExportTheme(theme *Theme) (string, error) {
	data, err := json.MarshalIndent(theme, "", "  ")
	if err != nil {
		return "", WrapError(err, "theme_export_failed", "failed to export theme")
	}
	return string(data), nil
}

// ImportTheme imports a theme from JSON
func ImportTheme(jsonData string) (*Theme, error) {
	var theme Theme
	if err := json.Unmarshal([]byte(jsonData), &theme); err != nil {
		return nil, WrapError(err, "theme_import_failed", "failed to import theme")
	}

	if err := ValidateTheme(&theme); err != nil {
		return nil, err
	}

	return &theme, nil
}
