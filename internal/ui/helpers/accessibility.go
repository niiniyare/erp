package helpers

import (
	"context"
	"fmt"
	"slices"
	"strings"
)

// AccessibilityHelper provides template helper functions for accessibility features
type AccessibilityHelper struct {
	// Configuration options
	config *AccessibilityConfig
}

// AccessibilityConfig holds accessibility configuration
type AccessibilityConfig struct {
	EnableScreenReader  bool
	EnableKeyboardNav   bool
	EnableHighContrast  bool
	EnableReducedMotion bool
	DefaultLanguage     string
	SupportedLanguages  []string
}

// NewAccessibilityHelper creates a new accessibility helper instance
func NewAccessibilityHelper(config *AccessibilityConfig) *AccessibilityHelper {
	if config == nil {
		config = DefaultAccessibilityConfig()
	}
	return &AccessibilityHelper{
		config: config,
	}
}

// DefaultAccessibilityConfig returns default accessibility configuration
func DefaultAccessibilityConfig() *AccessibilityConfig {
	return &AccessibilityConfig{
		EnableScreenReader:  true,
		EnableKeyboardNav:   true,
		EnableHighContrast:  true,
		EnableReducedMotion: true,
		DefaultLanguage:     "en",
		SupportedLanguages:  []string{"en", "es", "fr", "de"},
	}
}

// ARIA Helper Functions

// GenerateAriaLabel creates an accessible label for an element
func (h *AccessibilityHelper) GenerateAriaLabel(baseLabel string, context map[string]string) string {
	if baseLabel == "" {
		return ""
	}

	// Add contextual information to the label
	var parts []string
	parts = append(parts, baseLabel)

	if status, ok := context["status"]; ok && status != "" {
		parts = append(parts, fmt.Sprintf("Status: %s", status))
	}

	if count, ok := context["count"]; ok && count != "" {
		parts = append(parts, fmt.Sprintf("Count: %s", count))
	}

	if level, ok := context["level"]; ok && level != "" {
		parts = append(parts, fmt.Sprintf("Level: %s", level))
	}

	return strings.Join(parts, ", ")
}

// GenerateAriaDescribedBy creates describedby relationships
func (h *AccessibilityHelper) GenerateAriaDescribedBy(elementID string, descriptions []string) string {
	var ids []string
	for i, desc := range descriptions {
		if desc != "" {
			ids = append(ids, fmt.Sprintf("%s-desc-%d", elementID, i))
		}
	}
	return strings.Join(ids, " ")
}

// GenerateAriaLabelledBy creates labelledby relationships
func (h *AccessibilityHelper) GenerateAriaLabelledBy(elementID string, labelIDs []string) string {
	return strings.Join(labelIDs, " ")
}

// Keyboard Navigation Helpers

// GenerateTabIndex returns appropriate tabindex value
func (h *AccessibilityHelper) GenerateTabIndex(interactive bool, forceFocus bool) string {
	if !h.config.EnableKeyboardNav {
		return ""
	}

	if forceFocus {
		return "0"
	}

	if interactive {
		return "0"
	}

	return "-1"
}

// GenerateKeyboardShortcut creates keyboard shortcut attributes
func (h *AccessibilityHelper) GenerateKeyboardShortcut(key string, modifiers []string) map[string]string {
	if !h.config.EnableKeyboardNav {
		return nil
	}

	attrs := make(map[string]string)

	if key != "" {
		attrs["data-key"] = key
		if len(modifiers) > 0 {
			attrs["data-modifiers"] = strings.Join(modifiers, "+")
		}

		// Generate accesskey for simple shortcuts
		if len(modifiers) == 1 && modifiers[0] == "alt" {
			attrs["accesskey"] = key
		}

		// Generate title with shortcut info
		shortcut := key
		if len(modifiers) > 0 {
			shortcut = strings.Join(modifiers, "+") + "+" + key
		}
		attrs["title"] = fmt.Sprintf("Keyboard shortcut: %s", shortcut)
	}

	return attrs
}

// Screen Reader Helpers

// GenerateScreenReaderText creates text for screen readers only
func (h *AccessibilityHelper) GenerateScreenReaderText(text string) string {
	if !h.config.EnableScreenReader || text == "" {
		return ""
	}
	return text
}

// GenerateAriaLive creates appropriate aria-live values
func (h *AccessibilityHelper) GenerateAriaLive(urgency string) string {
	if !h.config.EnableScreenReader {
		return ""
	}

	switch urgency {
	case "high", "urgent", "error":
		return "assertive"
	case "medium", "info", "success":
		return "polite"
	case "low", "status":
		return "polite"
	default:
		return "polite"
	}
}

// Form Accessibility Helpers

// GenerateFormFieldAttributes creates comprehensive form field attributes
func (h *AccessibilityHelper) GenerateFormFieldAttributes(fieldConfig *FormFieldConfig) map[string]string {
	attrs := make(map[string]string)

	if fieldConfig == nil {
		return attrs
	}

	// Required field attributes
	if fieldConfig.Required {
		attrs["required"] = "true"
		attrs["aria-required"] = "true"
	}

	// Invalid field attributes
	if fieldConfig.HasError {
		attrs["aria-invalid"] = "true"
		if fieldConfig.ErrorID != "" {
			attrs["aria-describedby"] = fieldConfig.ErrorID
		}
	}

	// Autocomplete attributes
	if fieldConfig.AutoComplete != "" {
		attrs["autocomplete"] = fieldConfig.AutoComplete
	}

	// Input mode for mobile optimization
	if fieldConfig.InputMode != "" {
		attrs["inputmode"] = fieldConfig.InputMode
	}

	// Pattern for validation
	if fieldConfig.Pattern != "" {
		attrs["pattern"] = fieldConfig.Pattern
	}

	return attrs
}

// FormFieldConfig holds form field accessibility configuration
type FormFieldConfig struct {
	Required     bool
	HasError     bool
	ErrorID      string
	AutoComplete string
	InputMode    string // numeric, email, tel, url, etc.
	Pattern      string
}

// Table Accessibility Helpers

// GenerateTableAttributes creates accessible table attributes
func (h *AccessibilityHelper) GenerateTableAttributes(config *TableConfig) map[string]string {
	attrs := make(map[string]string)

	if config == nil {
		return attrs
	}

	// Table role and description
	attrs["role"] = "table"
	if config.Caption != "" {
		attrs["aria-label"] = config.Caption
	}

	// Sortable table
	if config.Sortable {
		attrs["aria-sort"] = "none"
	}

	// Row and column counts for screen readers
	if config.RowCount > 0 {
		attrs["aria-rowcount"] = fmt.Sprintf("%d", config.RowCount)
	}
	if config.ColCount > 0 {
		attrs["aria-colcount"] = fmt.Sprintf("%d", config.ColCount)
	}

	return attrs
}

// TableConfig holds table accessibility configuration
type TableConfig struct {
	Caption  string
	Sortable bool
	RowCount int
	ColCount int
}

// Button Accessibility Helpers

// GenerateButtonAttributes creates accessible button attributes
func (h *AccessibilityHelper) GenerateButtonAttributes(config *ButtonConfig) map[string]string {
	attrs := make(map[string]string)

	if config == nil {
		return attrs
	}

	// Button state
	if config.Pressed != nil {
		attrs["aria-pressed"] = fmt.Sprintf("%t", *config.Pressed)
	}

	if config.Expanded != nil {
		attrs["aria-expanded"] = fmt.Sprintf("%t", *config.Expanded)
	}

	// Button controls
	if config.Controls != "" {
		attrs["aria-controls"] = config.Controls
	}

	// Button popup
	if config.HasPopup {
		attrs["aria-haspopup"] = "true"
	}

	// Disabled state
	if config.Disabled {
		attrs["disabled"] = "true"
		attrs["aria-disabled"] = "true"
	}

	return attrs
}

// ButtonConfig holds button accessibility configuration
type ButtonConfig struct {
	Pressed  *bool
	Expanded *bool
	Controls string
	HasPopup bool
	Disabled bool
}

// Navigation Helpers

// GenerateNavAttributes creates accessible navigation attributes
func (h *AccessibilityHelper) GenerateNavAttributes(config *NavConfig) map[string]string {
	attrs := make(map[string]string)

	if config == nil {
		return attrs
	}

	// Navigation role and label
	attrs["role"] = "navigation"
	if config.Label != "" {
		attrs["aria-label"] = config.Label
	}

	// Current page indication
	if config.CurrentPage != "" {
		attrs["aria-current"] = "page"
	}

	return attrs
}

// NavConfig holds navigation accessibility configuration
type NavConfig struct {
	Label       string
	CurrentPage string
}

// Progressive Helpers

// GenerateProgressiveEnhancementAttrs creates attributes for progressive enhancement
func (h *AccessibilityHelper) GenerateProgressiveEnhancementAttrs(config *ProgressiveConfig) map[string]string {
	attrs := make(map[string]string)

	if config == nil {
		return attrs
	}

	// No-JS fallback
	if config.NoJSFallback {
		attrs["data-nojs-fallback"] = "true"
	}

	// Enhanced with JS
	if config.EnhancedBehavior != "" {
		attrs["data-enhanced"] = config.EnhancedBehavior
	}

	// Loading states
	if config.LoadingState {
		attrs["data-loading"] = "false"
		attrs["aria-busy"] = "false"
	}

	return attrs
}

// ProgressiveConfig holds progressive enhancement configuration
type ProgressiveConfig struct {
	NoJSFallback     bool
	EnhancedBehavior string
	LoadingState     bool
}

// Color and Contrast Helpers

// GetContrastClass returns appropriate contrast class
func (h *AccessibilityHelper) GetContrastClass(highContrast bool) string {
	if !h.config.EnableHighContrast {
		return ""
	}

	if highContrast {
		return "high-contrast"
	}

	return ""
}

// Motion Helpers

// GetMotionClass returns appropriate motion class
func (h *AccessibilityHelper) GetMotionClass(respectMotionPreference bool) string {
	if !h.config.EnableReducedMotion {
		return ""
	}

	if respectMotionPreference {
		return "respect-motion-preference"
	}

	return ""
}

// Language and Internationalization Helpers

// GetLanguageAttribute returns the best language attribute value
func (h *AccessibilityHelper) GetLanguageAttribute(ctx context.Context) string {
	// FIXME: Extract language from context or user preferences

	// Try to extract language from context (if middleware sets it)
	if lang, ok := ctx.Value("lang").(string); ok && h.IsLanguageSupported(lang) {
		return lang
	}
	// Fallback to default
	return h.config.DefaultLanguage
}

// IsLanguageSupported checks if a language is supported
func (h *AccessibilityHelper) IsLanguageSupported(lang string) bool {
	return slices.Contains(h.config.SupportedLanguages, lang)
}

// Utility Functions

// SanitizeID creates a valid HTML ID from a string
func (h *AccessibilityHelper) SanitizeID(input string) string {
	// Replace spaces and special characters with hyphens
	id := strings.ReplaceAll(input, " ", "-")
	id = strings.ToLower(id)

	// Remove invalid characters
	var result strings.Builder
	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// GenerateUniqueID creates a unique ID for an element
func (h *AccessibilityHelper) GenerateUniqueID(prefix string) string {
	// Simple implementation - in production, this might use a more sophisticated approach
	timestamp := fmt.Sprintf("%d", len(prefix)) // Simplified for demo
	return fmt.Sprintf("%s-%s", h.SanitizeID(prefix), timestamp)
}
