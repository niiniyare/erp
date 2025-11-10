package schema

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

//go:embed translations/*.json
var translationFiles embed.FS

// TranslationCatalog holds all locale translations with embedded files
type TranslationCatalog struct {
	locales       map[string]*LocaleTranslations
	defaultLocale string
	mu            sync.RWMutex
}

// LocaleTranslations contains all translation categories for a specific locale
type LocaleTranslations struct {
	Fields     map[string]string `json:"fields"`
	Validation map[string]string `json:"validation"`
	Actions    map[string]string `json:"actions"`
	Status     map[string]string `json:"status"`
}

// Global translation catalog instance
var Translator *TranslationCatalog

func init() {
	Translator = NewTranslationCatalog("en")
	if err := Translator.LoadEmbeddedTranslations(); err != nil {
		// Fallback to English-only if loading fails
		Translator.locales["en"] = &LocaleTranslations{
			Fields:     make(map[string]string),
			Validation: make(map[string]string),
			Actions:    make(map[string]string),
			Status:     make(map[string]string),
		}
	}
}

// NewTranslationCatalog creates a new translation catalog with the specified default locale
func NewTranslationCatalog(defaultLocale string) *TranslationCatalog {
	return &TranslationCatalog{
		locales:       make(map[string]*LocaleTranslations),
		defaultLocale: defaultLocale,
	}
}

// LoadEmbeddedTranslations loads all translation files from embedded filesystem
func (tc *TranslationCatalog) LoadEmbeddedTranslations() error {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	files, err := translationFiles.ReadDir("translations")
	if err != nil {
		return fmt.Errorf("failed to read translation directory: %w", err)
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		locale := strings.TrimSuffix(file.Name(), ".json")
		
		data, err := translationFiles.ReadFile("translations/" + file.Name())
		if err != nil {
			continue // Skip failed files
		}

		var localeData LocaleTranslations
		if err := json.Unmarshal(data, &localeData); err != nil {
			continue // Skip malformed files
		}

		tc.locales[locale] = &localeData
	}

	return nil
}

// HasLocale checks if a specific locale is available
func (tc *TranslationCatalog) HasLocale(locale string) bool {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	_, exists := tc.locales[locale]
	return exists
}

// GetAvailableLocales returns a list of all available locales
func (tc *TranslationCatalog) GetAvailableLocales() []string {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	
	locales := make([]string, 0, len(tc.locales))
	for locale := range tc.locales {
		locales = append(locales, locale)
	}
	return locales
}

// DetectBestLocale finds the best available locale from a list of preferences
func (tc *TranslationCatalog) DetectBestLocale(preferences []string) string {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	// Check exact matches first
	for _, pref := range preferences {
		if _, exists := tc.locales[pref]; exists {
			return pref
		}
	}

	// Check language-only matches (e.g., "en" for "en-US")
	for _, pref := range preferences {
		lang := strings.Split(pref, "-")[0]
		if _, exists := tc.locales[lang]; exists {
			return lang
		}
	}

	// Fallback to default locale
	return tc.defaultLocale
}

// FieldLabel returns the translated field label for the given locale and field ID
func (tc *TranslationCatalog) FieldLabel(locale, fieldID string) string {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	// Try requested locale first
	if localeData, exists := tc.locales[locale]; exists {
		if translation, exists := localeData.Fields[fieldID]; exists {
			return translation
		}
	}

	// Fallback to default locale
	if locale != tc.defaultLocale {
		if defaultData, exists := tc.locales[tc.defaultLocale]; exists {
			if translation, exists := defaultData.Fields[fieldID]; exists {
				return translation
			}
		}
	}

	// Return humanized field ID as last resort
	return humanizeFieldID(fieldID)
}

// ValidationMessage returns the translated validation message with parameter substitution
func (tc *TranslationCatalog) ValidationMessage(locale, messageKey string, params map[string]any) string {
	template := tc.getValidationTemplate(locale, messageKey)
	return tc.interpolateMessage(template, params)
}

// ActionText returns the translated action text
func (tc *TranslationCatalog) ActionText(locale, actionKey string) string {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	// Try requested locale first
	if localeData, exists := tc.locales[locale]; exists {
		if translation, exists := localeData.Actions[actionKey]; exists {
			return translation
		}
	}

	// Fallback to default locale
	if locale != tc.defaultLocale {
		if defaultData, exists := tc.locales[tc.defaultLocale]; exists {
			if translation, exists := defaultData.Actions[actionKey]; exists {
				return translation
			}
		}
	}

	// Return humanized action key as last resort
	return humanizeFieldID(actionKey)
}

// StatusText returns the translated status text
func (tc *TranslationCatalog) StatusText(locale, statusKey string) string {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	// Try requested locale first
	if localeData, exists := tc.locales[locale]; exists {
		if translation, exists := localeData.Status[statusKey]; exists {
			return translation
		}
	}

	// Fallback to default locale
	if locale != tc.defaultLocale {
		if defaultData, exists := tc.locales[tc.defaultLocale]; exists {
			if translation, exists := defaultData.Status[statusKey]; exists {
				return translation
			}
		}
	}

	// Return humanized status key as last resort
	return humanizeFieldID(statusKey)
}

// getValidationTemplate retrieves validation message template with fallback
func (tc *TranslationCatalog) getValidationTemplate(locale, messageKey string) string {
	// Try requested locale first
	if localeData, exists := tc.locales[locale]; exists {
		if template, exists := localeData.Validation[messageKey]; exists {
			return template
		}
	}

	// Fallback to default locale
	if locale != tc.defaultLocale {
		if defaultData, exists := tc.locales[tc.defaultLocale]; exists {
			if template, exists := defaultData.Validation[messageKey]; exists {
				return template
			}
		}
	}

	// Return key as fallback
	return messageKey
}

// interpolateMessage performs parameter substitution in message templates
func (tc *TranslationCatalog) interpolateMessage(template string, params map[string]any) string {
	if params == nil {
		return template
	}

	result := template
	for key, value := range params {
		placeholder := fmt.Sprintf("{%s}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}

	return result
}

// humanizeFieldID converts camelCase field IDs to readable labels
func humanizeFieldID(fieldID string) string {
	if fieldID == "" {
		return ""
	}

	// Convert camelCase to space-separated words
	var result strings.Builder
	for i, r := range fieldID {
		if i > 0 && (r >= 'A' && r <= 'Z') {
			result.WriteString(" ")
		}
		if i == 0 {
			result.WriteString(strings.ToUpper(string(r)))
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// IsRTL checks if the given locale uses right-to-left text direction
func (tc *TranslationCatalog) IsRTL(locale string) bool {
	rtlLocales := map[string]bool{
		"ar": true, // Arabic
		"he": true, // Hebrew
		"fa": true, // Persian/Farsi
		"ur": true, // Urdu
	}
	
	// Check both full locale and language part
	if rtlLocales[locale] {
		return true
	}
	
	lang := strings.Split(locale, "-")[0]
	return rtlLocales[lang]
}

// GetTextDirection returns the text direction for the locale ("ltr" or "rtl")
func (tc *TranslationCatalog) GetTextDirection(locale string) string {
	if tc.IsRTL(locale) {
		return "rtl"
	}
	return "ltr"
}

// Convenience functions for direct access to global translator

// T_Field returns translated field label using global translator
func T_Field(locale, fieldID string) string {
	return Translator.FieldLabel(locale, fieldID)
}

// T_Validation returns translated validation message using global translator
func T_Validation(locale, messageKey string, params map[string]any) string {
	return Translator.ValidationMessage(locale, messageKey, params)
}

// T_Action returns translated action text using global translator
func T_Action(locale, actionKey string) string {
	return Translator.ActionText(locale, actionKey)
}

// T_Status returns translated status text using global translator
func T_Status(locale, statusKey string) string {
	return Translator.StatusText(locale, statusKey)
}

// T_HasLocale checks if locale is available using global translator
func T_HasLocale(locale string) bool {
	return Translator.HasLocale(locale)
}

// T_DetectLocale finds best locale from preferences using global translator
func T_DetectLocale(preferences []string) string {
	return Translator.DetectBestLocale(preferences)
}

// T_IsRTL checks if locale is right-to-left using global translator
func T_IsRTL(locale string) bool {
	return Translator.IsRTL(locale)
}

// T_Direction returns text direction using global translator
func T_Direction(locale string) string {
	return Translator.GetTextDirection(locale)
}