package helpers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"slices"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/niiniyare/erp/internal/shared/logger"
)

// Context key type for type-safe context values
type contextKey string

const (
	// LanguageContextKey is the key for language in context
	LanguageContextKey contextKey = "lang"
)

// AccessibilityHelper provides template helper functions for accessibility features
type AccessibilityHelper struct {
	// Configuration options
	config    *AccessibilityConfig
	mu        sync.RWMutex  // Protect config if needed
	idCounter uint64        // Atomic counter for unique IDs
	logger    logger.Logger // Logger instance
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
		logger: nil, // No logger by default for backward compatibility
	}
}

// NewAccessibilityHelperWithLogger creates a helper with custom logger
func NewAccessibilityHelperWithLogger(config *AccessibilityConfig, log logger.Logger) *AccessibilityHelper {
	helper := NewAccessibilityHelper(config)
	helper.logger = log
	return helper
}

// DefaultAccessibilityConfig returns default accessibility configuration
func DefaultAccessibilityConfig() *AccessibilityConfig {
	return &AccessibilityConfig{
		EnableScreenReader:  true,
		EnableKeyboardNav:   true,
		EnableHighContrast:  true,
		EnableReducedMotion: true,
		DefaultLanguage:     "en",
		SupportedLanguages:  []string{"en", "es", "fr", "de", "ar", "so", "sw"},
	}
}

// logDebug logs a debug message if logger is available
func (h *AccessibilityHelper) logDebug(msg string, fields ...logger.Fields) {
	if h.logger != nil {
		h.logger.Debug(msg, fields...)
	}
}

// logWarn logs a warning message if logger is available
func (h *AccessibilityHelper) logWarn(msg string, fields ...logger.Fields) {
	if h.logger != nil {
		h.logger.Warn(msg, fields...)
	}
}

// logError logs an error message if logger is available
func (h *AccessibilityHelper) logError(msg string, fields ...logger.Fields) {
	if h.logger != nil {
		h.logger.Error(msg, fields...)
	}
}

// GetConfig returns a copy of the current configuration (thread-safe)
func (h *AccessibilityHelper) GetConfig() AccessibilityConfig {
	h.mu.RLock()
	defer h.mu.RUnlock()
	// Return a copy to prevent external modification
	configCopy := *h.config
	configCopy.SupportedLanguages = make([]string, len(h.config.SupportedLanguages))
	copy(configCopy.SupportedLanguages, h.config.SupportedLanguages)
	return configCopy
}

// UpdateConfig safely updates the configuration (thread-safe)
func (h *AccessibilityHelper) UpdateConfig(config *AccessibilityConfig) {
	if config == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.config = config

	h.logDebug("Accessibility config updated", logger.Fields{
		"enable_screen_reader":  config.EnableScreenReader,
		"enable_keyboard_nav":   config.EnableKeyboardNav,
		"enable_high_contrast":  config.EnableHighContrast,
		"enable_reduced_motion": config.EnableReducedMotion,
		"default_language":      config.DefaultLanguage,
	})
}

// SetLogger sets the logger instance (useful for dependency injection)
func (h *AccessibilityHelper) SetLogger(log logger.Logger) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.logger = log
}

// ARIA Helper Functions

// GenerateAriaLabel creates an accessible label for an element
func (h *AccessibilityHelper) GenerateAriaLabel(baseLabel string, context map[string]string) string {
	if baseLabel == "" {
		h.logWarn("GenerateAriaLabel called with empty baseLabel")
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

	label := strings.Join(parts, ", ")
	h.logDebug("Generated ARIA label", logger.Fields{
		"base_label":  baseLabel,
		"final_label": label,
	})

	return label
}

// GenerateAriaLabelWithValidation is an enhanced version with validation
func (h *AccessibilityHelper) GenerateAriaLabelWithValidation(baseLabel string, context map[string]string) (string, error) {
	if baseLabel == "" {
		err := fmt.Errorf("baseLabel cannot be empty")
		h.logError("Validation failed for ARIA label", logger.Fields{
			"error": err.Error(),
		})
		return "", err
	}

	if len(baseLabel) > 255 {
		err := fmt.Errorf("baseLabel exceeds maximum length of 255 characters")
		h.logError("Validation failed for ARIA label", logger.Fields{
			"error":  err.Error(),
			"length": len(baseLabel),
		})
		return "", err
	}

	return h.GenerateAriaLabel(baseLabel, context), nil
}

// GenerateAriaDescribedBy creates describedby relationships
func (h *AccessibilityHelper) GenerateAriaDescribedBy(elementID string, descriptions []string) string {
	if elementID == "" {
		h.logWarn("GenerateAriaDescribedBy called with empty elementID")
		return ""
	}

	var ids []string
	for i, desc := range descriptions {
		if desc != "" {
			ids = append(ids, fmt.Sprintf("%s-desc-%d", elementID, i))
		}
	}

	result := strings.Join(ids, " ")
	h.logDebug("Generated aria-describedby", logger.Fields{
		"element_id":        elementID,
		"description_count": len(descriptions),
		"result":            result,
	})

	return result
}

// GenerateAriaLabelledBy creates labelledby relationships
func (h *AccessibilityHelper) GenerateAriaLabelledBy(elementID string, labelIDs []string) string {
	if elementID == "" {
		h.logWarn("GenerateAriaLabelledBy called with empty elementID")
	}

	result := strings.Join(labelIDs, " ")
	h.logDebug("Generated aria-labelledby", logger.Fields{
		"element_id":  elementID,
		"label_count": len(labelIDs),
	})

	return result
}

// Keyboard Navigation Helpers

// GenerateTabIndex returns appropriate tabindex value
func (h *AccessibilityHelper) GenerateTabIndex(interactive bool, forceFocus bool) string {
	h.mu.RLock()
	defer h.mu.RUnlock()

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
	h.mu.RLock()
	defer h.mu.RUnlock()

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

		h.logDebug("Generated keyboard shortcut", logger.Fields{
			"key":       key,
			"modifiers": modifiers,
			"shortcut":  shortcut,
		})
	}

	return attrs
}

// Screen Reader Helpers

// GenerateScreenReaderText creates text for screen readers only
func (h *AccessibilityHelper) GenerateScreenReaderText(text string) string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if !h.config.EnableScreenReader || text == "" {
		return ""
	}
	return text
}

// GenerateAriaLive creates appropriate aria-live values
func (h *AccessibilityHelper) GenerateAriaLive(urgency string) string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if !h.config.EnableScreenReader {
		return ""
	}

	var result string
	switch strings.ToLower(urgency) {
	case "high", "urgent", "error":
		result = "assertive"
	case "medium", "info", "success", "low", "status":
		result = "polite"
	case "off":
		result = "off"
	default:
		result = "polite"
	}

	h.logDebug("Generated aria-live", logger.Fields{
		"urgency": urgency,
		"result":  result,
	})

	return result
}

// Form Accessibility Helpers

// GenerateFormFieldAttributes creates comprehensive form field attributes
func (h *AccessibilityHelper) GenerateFormFieldAttributes(fieldConfig *FormFieldConfig) map[string]string {
	attrs := make(map[string]string)

	if fieldConfig == nil {
		h.logWarn("GenerateFormFieldAttributes called with nil config")
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

	h.logDebug("Generated form field attributes", logger.Fields{
		"required":     fieldConfig.Required,
		"has_error":    fieldConfig.HasError,
		"autocomplete": fieldConfig.AutoComplete,
		"input_mode":   fieldConfig.InputMode,
	})

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

// FormFieldConfigBuilder provides a fluent API for building form field configs
type FormFieldConfigBuilder struct {
	config FormFieldConfig
}

// NewFormFieldConfig creates a new form field config builder
func NewFormFieldConfig() *FormFieldConfigBuilder {
	return &FormFieldConfigBuilder{}
}

func (b *FormFieldConfigBuilder) Required() *FormFieldConfigBuilder {
	b.config.Required = true
	return b
}

func (b *FormFieldConfigBuilder) WithError(errorID string) *FormFieldConfigBuilder {
	b.config.HasError = true
	b.config.ErrorID = errorID
	return b
}

func (b *FormFieldConfigBuilder) AutoComplete(value string) *FormFieldConfigBuilder {
	b.config.AutoComplete = value
	return b
}

func (b *FormFieldConfigBuilder) InputMode(mode string) *FormFieldConfigBuilder {
	b.config.InputMode = mode
	return b
}

func (b *FormFieldConfigBuilder) Pattern(pattern string) *FormFieldConfigBuilder {
	b.config.Pattern = pattern
	return b
}

func (b *FormFieldConfigBuilder) Build() *FormFieldConfig {
	config := b.config
	return &config
}

// Table Accessibility Helpers

// GenerateTableAttributes creates accessible table attributes
func (h *AccessibilityHelper) GenerateTableAttributes(config *TableConfig) map[string]string {
	attrs := make(map[string]string)

	if config == nil {
		h.logWarn("GenerateTableAttributes called with nil config")
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

	h.logDebug("Generated table attributes", logger.Fields{
		"caption":   config.Caption,
		"sortable":  config.Sortable,
		"row_count": config.RowCount,
		"col_count": config.ColCount,
	})

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
		h.logWarn("GenerateButtonAttributes called with nil config")
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

	h.logDebug("Generated button attributes", logger.Fields{
		"has_pressed":  config.Pressed != nil,
		"has_expanded": config.Expanded != nil,
		"controls":     config.Controls,
		"has_popup":    config.HasPopup,
		"disabled":     config.Disabled,
	})

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

// ButtonConfigBuilder provides a fluent API for building button configs
type ButtonConfigBuilder struct {
	config ButtonConfig
}

// NewButtonConfig creates a new button config builder
func NewButtonConfig() *ButtonConfigBuilder {
	return &ButtonConfigBuilder{}
}

func (b *ButtonConfigBuilder) Pressed(pressed bool) *ButtonConfigBuilder {
	b.config.Pressed = &pressed
	return b
}

func (b *ButtonConfigBuilder) Expanded(expanded bool) *ButtonConfigBuilder {
	b.config.Expanded = &expanded
	return b
}

func (b *ButtonConfigBuilder) Controls(elementID string) *ButtonConfigBuilder {
	b.config.Controls = elementID
	return b
}

func (b *ButtonConfigBuilder) HasPopup() *ButtonConfigBuilder {
	b.config.HasPopup = true
	return b
}

func (b *ButtonConfigBuilder) Disabled() *ButtonConfigBuilder {
	b.config.Disabled = true
	return b
}

func (b *ButtonConfigBuilder) Build() *ButtonConfig {
	config := b.config
	return &config
}

// Navigation Helpers

// GenerateNavAttributes creates accessible navigation attributes
func (h *AccessibilityHelper) GenerateNavAttributes(config *NavConfig) map[string]string {
	attrs := make(map[string]string)

	if config == nil {
		h.logWarn("GenerateNavAttributes called with nil config")
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

	h.logDebug("Generated navigation attributes", logger.Fields{
		"label":        config.Label,
		"current_page": config.CurrentPage,
	})

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
		h.logWarn("GenerateProgressiveEnhancementAttrs called with nil config")
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
	h.mu.RLock()
	defer h.mu.RUnlock()

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
	h.mu.RLock()
	defer h.mu.RUnlock()

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
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Try to extract language from context using type-safe key
	if lang, ok := ctx.Value(LanguageContextKey).(string); ok && h.isLanguageSupportedUnsafe(lang) {
		return lang
	}

	// Fallback to default
	return h.config.DefaultLanguage
}

// GetLanguageAttributeWithFallback returns language with custom fallback
func (h *AccessibilityHelper) GetLanguageAttributeWithFallback(ctx context.Context, fallback string) string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if lang, ok := ctx.Value(LanguageContextKey).(string); ok && h.isLanguageSupportedUnsafe(lang) {
		return lang
	}

	if fallback != "" && h.isLanguageSupportedUnsafe(fallback) {
		return fallback
	}

	return h.config.DefaultLanguage
}

// IsLanguageSupported checks if a language is supported
func (h *AccessibilityHelper) IsLanguageSupported(lang string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.isLanguageSupportedUnsafe(lang)
}

// isLanguageSupportedUnsafe checks language support without locking (internal use)
func (h *AccessibilityHelper) isLanguageSupportedUnsafe(lang string) bool {
	return slices.Contains(h.config.SupportedLanguages, lang)
}

// Utility Functions

// SanitizeID creates a valid HTML ID from a string
func (h *AccessibilityHelper) SanitizeID(input string) string {
	if input == "" {
		h.logWarn("SanitizeID called with empty input")
		return ""
	}

	// Replace spaces and special characters with hyphens
	id := strings.ReplaceAll(input, " ", "-")
	id = strings.ToLower(id)

	// Remove invalid characters
	var result strings.Builder
	result.Grow(len(id)) // Pre-allocate capacity

	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			result.WriteRune(r)
		}
	}

	sanitized := result.String()

	// Ensure ID starts with a letter (HTML requirement)
	if len(sanitized) > 0 && (sanitized[0] >= '0' && sanitized[0] <= '9') {
		sanitized = "id-" + sanitized
	}

	h.logDebug("Sanitized ID", logger.Fields{
		"input":  input,
		"output": sanitized,
	})

	return sanitized
}

// GenerateUniqueID creates a unique ID for an element using atomic counter
func (h *AccessibilityHelper) GenerateUniqueID(prefix string) string {
	id := atomic.AddUint64(&h.idCounter, 1)
	sanitizedPrefix := h.SanitizeID(prefix)
	if sanitizedPrefix == "" {
		sanitizedPrefix = "element"
	}
	uniqueID := fmt.Sprintf("%s-%d", sanitizedPrefix, id)

	h.logDebug("Generated unique ID", logger.Fields{
		"prefix": prefix,
		"id":     uniqueID,
	})

	return uniqueID
}

// GenerateSecureUniqueID creates a cryptographically unique ID
func (h *AccessibilityHelper) GenerateSecureUniqueID(prefix string) string {
	sanitizedPrefix := h.SanitizeID(prefix)
	if sanitizedPrefix == "" {
		sanitizedPrefix = "element"
	}

	// Generate 8 random bytes (16 hex characters)
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// Fallback to atomic counter if crypto/rand fails
		h.logError("Failed to generate secure random ID, falling back to atomic counter", logger.Fields{
			"error":  err.Error(),
			"prefix": prefix,
		})
		return h.GenerateUniqueID(prefix)
	}

	secureID := fmt.Sprintf("%s-%s", sanitizedPrefix, hex.EncodeToString(b))
	h.logDebug("Generated secure unique ID", logger.Fields{
		"prefix": prefix,
		"id":     secureID,
	})

	return secureID
}

// Advanced Helpers (New additions for enhanced functionality)

// GenerateLandmarkRole returns appropriate landmark role
func (h *AccessibilityHelper) GenerateLandmarkRole(landmark string) string {
	validLandmarks := map[string]bool{
		"banner":        true,
		"complementary": true,
		"contentinfo":   true,
		"form":          true,
		"main":          true,
		"navigation":    true,
		"region":        true,
		"search":        true,
	}

	if validLandmarks[landmark] {
		return landmark
	}

	h.logWarn("Invalid landmark role requested", logger.Fields{
		"landmark": landmark,
	})
	return ""
}

// GenerateAriaHidden returns aria-hidden attribute
func (h *AccessibilityHelper) GenerateAriaHidden(hidden bool) string {
	if hidden {
		return "true"
	}
	return "false"
}

// GenerateRole returns a valid ARIA role with validation
func (h *AccessibilityHelper) GenerateRole(role string) string {
	// Common valid ARIA roles
	validRoles := map[string]bool{
		"alert": true, "alertdialog": true, "application": true, "article": true,
		"banner": true, "button": true, "cell": true, "checkbox": true,
		"columnheader": true, "combobox": true, "complementary": true,
		"contentinfo": true, "definition": true, "dialog": true, "directory": true,
		"document": true, "feed": true, "figure": true, "form": true, "grid": true,
		"gridcell": true, "group": true, "heading": true, "img": true, "link": true,
		"list": true, "listbox": true, "listitem": true, "log": true, "main": true,
		"marquee": true, "math": true, "menu": true, "menubar": true, "menuitem": true,
		"menuitemcheckbox": true, "menuitemradio": true, "navigation": true,
		"none": true, "note": true, "option": true, "presentation": true,
		"progressbar": true, "radio": true, "radiogroup": true, "region": true,
		"row": true, "rowgroup": true, "rowheader": true, "scrollbar": true,
		"search": true, "searchbox": true, "separator": true, "slider": true,
		"spinbutton": true, "status": true, "switch": true, "tab": true,
		"table": true, "tablist": true, "tabpanel": true, "term": true,
		"textbox": true, "timer": true, "toolbar": true, "tooltip": true,
		"tree": true, "treegrid": true, "treeitem": true,
	}

	if validRoles[role] {
		return role
	}

	h.logWarn("Invalid ARIA role requested", logger.Fields{
		"role": role,
	})
	return ""
}

// MergeAttributes safely merges multiple attribute maps
func (h *AccessibilityHelper) MergeAttributes(attrMaps ...map[string]string) map[string]string {
	result := make(map[string]string)
	for _, attrs := range attrMaps {
		for k, v := range attrs {
			result[k] = v
		}
	}
	return result
}

// ValidateAccessibilityAttributes checks if attributes are valid
func (h *AccessibilityHelper) ValidateAccessibilityAttributes(attrs map[string]string) []string {
	var warnings []string

	// Check for common mistakes
	if ariaPressedValue, exists := attrs["aria-pressed"]; exists {
		if ariaPressedValue != "true" && ariaPressedValue != "false" && ariaPressedValue != "mixed" {
			warning := fmt.Sprintf("Invalid aria-pressed value: %s", ariaPressedValue)
			warnings = append(warnings, warning)
			h.logWarn("Invalid accessibility attribute", logger.Fields{
				"attribute": "aria-pressed",
				"value":     ariaPressedValue,
			})
		}
	}

	if ariaInvalidValue, exists := attrs["aria-invalid"]; exists {
		validValues := map[string]bool{"true": true, "false": true, "grammar": true, "spelling": true}
		if !validValues[ariaInvalidValue] {
			warning := fmt.Sprintf("Invalid aria-invalid value: %s", ariaInvalidValue)
			warnings = append(warnings, warning)
			h.logWarn("Invalid accessibility attribute", logger.Fields{
				"attribute": "aria-invalid",
				"value":     ariaInvalidValue,
			})
		}
	}

	if len(warnings) > 0 {
		h.logWarn("Accessibility validation completed with warnings", logger.Fields{
			"warning_count": len(warnings),
		})
	}

	return warnings
}

// Helper function to create context with language
func SetLanguageContext(ctx context.Context, lang string) context.Context {
	return context.WithValue(ctx, LanguageContextKey, lang)
}
