package schema

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
)

// SchemaI18n holds schema-level translations (matches docs structure)
type SchemaI18n struct {
	Title       map[string]string `json:"title,omitempty"`       // Localized schema titles
	Description map[string]string `json:"description,omitempty"` // Localized schema descriptions
	Messages    map[string]string `json:"messages,omitempty"`    // Custom schema messages
	Direction   map[string]string `json:"direction,omitempty"`   // Text direction: "ltr" or "rtl"
}

// I18nManager handles runtime translations and validation messages
type I18nManager struct {
	locale         string
	fallbackLocale string
	messages       map[string]map[string]string // [locale][key]message
	loader         *I18nLoader
	mu             sync.RWMutex
}

// I18nLoader loads translations from files or database
type I18nLoader struct {
	translationsDir string
	cache           map[string]map[string]string
	mu              sync.RWMutex
}

// PluralForms handles pluralization for different languages
type PluralForms struct {
	Zero  string `json:"zero,omitempty"`
	One   string `json:"one"`
	Two   string `json:"two,omitempty"`
	Few   string `json:"few,omitempty"`
	Many  string `json:"many,omitempty"`
	Other string `json:"other"`
}

// NewI18nManager creates a new internationalization manager
func NewI18nManager(locale, fallbackLocale string) *I18nManager {
	return &I18nManager{
		locale:         locale,
		fallbackLocale: fallbackLocale,
		messages:       make(map[string]map[string]string),
		loader:         NewI18nLoader(""),
	}
}

// NewI18nLoader creates a new translation loader
func NewI18nLoader(translationsDir string) *I18nLoader {
	return &I18nLoader{
		translationsDir: translationsDir,
		cache:           make(map[string]map[string]string),
	}
}

// LoadLocale loads translations for a specific locale
func (l *I18nLoader) LoadLocale(locale string) (map[string]string, error) {
	l.mu.RLock()
	if cached, exists := l.cache[locale]; exists {
		l.mu.RUnlock()
		return cached, nil
	}
	l.mu.RUnlock()

	filePath := filepath.Join(l.translationsDir, locale+".json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load translations for locale %s: %w", locale, err)
	}

	var translations map[string]any
	if err := json.Unmarshal(data, &translations); err != nil {
		return nil, fmt.Errorf("failed to parse translations for locale %s: %w", locale, err)
	}

	// Flatten nested JSON structure for easier key access
	flattened := flattenJSON(translations)

	l.mu.Lock()
	l.cache[locale] = flattened
	l.mu.Unlock()

	return flattened, nil
}

// flattenJSON flattens nested JSON for easier key access
func flattenJSON(data map[string]any) map[string]string {
	result := make(map[string]string)
	flatten(data, "", result)
	return result
}

// flatten recursively flattens nested map structures
func flatten(data map[string]any, prefix string, result map[string]string) {
	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch v := value.(type) {
		case string:
			result[fullKey] = v
		case map[string]any:
			flatten(v, fullKey, result)
		}
	}
}

// LoadTranslations loads translations for a locale into the manager
func (i *I18nManager) LoadTranslations(locale string, loader *I18nLoader) error {
	messages, err := loader.LoadLocale(locale)
	if err != nil {
		return err
	}

	i.mu.Lock()
	defer i.mu.Unlock()

	i.messages[locale] = messages
	return nil
}

// GetValidationMessage returns a localized validation message
func (i *I18nManager) GetValidationMessage(key string, params map[string]any) string {
	i.mu.RLock()
	defer i.mu.RUnlock()

	// Try current locale first
	if localeMessages, exists := i.messages[i.locale]; exists {
		if message, exists := localeMessages["validation."+key]; exists {
			return i.interpolateMessage(message, params)
		}
	}

	// Fall back to default locale
	if localeMessages, exists := i.messages[i.fallbackLocale]; exists {
		if message, exists := localeMessages["validation."+key]; exists {
			return i.interpolateMessage(message, params)
		}
	}

	// Return key if no translation found
	return key
}

// GetMessage returns a localized message for any key
func (i *I18nManager) GetMessage(key string, params map[string]any) string {
	i.mu.RLock()
	defer i.mu.RUnlock()

	// Try current locale first
	if localeMessages, exists := i.messages[i.locale]; exists {
		if message, exists := localeMessages[key]; exists {
			return i.interpolateMessage(message, params)
		}
	}

	// Fall back to default locale
	if localeMessages, exists := i.messages[i.fallbackLocale]; exists {
		if message, exists := localeMessages[key]; exists {
			return i.interpolateMessage(message, params)
		}
	}

	// Return key if no translation found
	return key
}

// interpolateMessage replaces template variables in messages
func (i *I18nManager) interpolateMessage(message string, params map[string]any) string {
	if len(params) == 0 {
		return message
	}

	// Use Go's template engine for interpolation
	tmpl, err := template.New("message").Parse(message)
	if err != nil {
		return message // Return original if template parsing fails
	}

	var result strings.Builder
	if err := tmpl.Execute(&result, params); err != nil {
		return message // Return original if execution fails
	}

	return result.String()
}

// SetLocale changes the current locale
func (i *I18nManager) SetLocale(locale string) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.locale = locale
}

// GetLocale returns the current locale
func (i *I18nManager) GetLocale() string {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.locale
}

// GetPluralMessage handles pluralization for different languages
func (i *I18nManager) GetPluralMessage(key string, count int, locale string) string {
	// Simple English pluralization rule
	if locale == "en" || locale == "" {
		if count == 1 {
			return i.GetMessage(key+".one", map[string]any{"count": count})
		}
		return i.GetMessage(key+".other", map[string]any{"count": count})
	}

	// For other languages, you would implement specific plural rules
	// Arabic: complex rules (0, 1, 2, 3-10, 11-99, 100+)
	// Russian: complex rules based on last digit
	// etc.

	// Fallback to "other" form
	return i.GetMessage(key+".other", map[string]any{"count": count})
}

// ApplySchemaI18nLocalization applies SchemaI18n translations to a schema for a specific locale
func (s *Schema) ApplySchemaI18nLocalization(schemaI18n *SchemaI18n, locale string) (*Schema, error) {
	if schemaI18n == nil || locale == "" {
		return s, nil
	}

	localizedSchema, err := s.Clone()
	if err != nil {
		return nil, err
	}

	// Apply schema-level translations
	if title, exists := schemaI18n.Title[locale]; exists {
		localizedSchema.Title = title
	}
	if description, exists := schemaI18n.Description[locale]; exists {
		localizedSchema.Description = description
	}

	// Apply field-level translations
	for i := range localizedSchema.Fields {
		field := &localizedSchema.Fields[i] // Get a pointer to the actual field
		if field.I18n != nil {
			if label, exists := field.I18n.Label[locale]; exists {
				field.Label = label
			}
			if placeholder, exists := field.I18n.Placeholder[locale]; exists {
				field.Placeholder = placeholder
			}
			if help, exists := field.I18n.Help[locale]; exists {
				field.Help = help
			}
			if tooltip, exists := field.I18n.Tooltip[locale]; exists {
				field.Tooltip = tooltip
			}
		}
	}

	return localizedSchema, nil
}

// GetLocalizedLabel returns localized label for a field
func (f *Field) GetLocalizedLabel(locale string) string {
	if f.I18n != nil && f.I18n.Label != nil {
		if label, exists := f.I18n.Label[locale]; exists {
			return label
		}
	}
	return f.Label // Fallback to default
}

// GetLocalizedPlaceholder returns localized placeholder for a field
func (f *Field) GetLocalizedPlaceholder(locale string) string {
	if f.I18n != nil && f.I18n.Placeholder != nil {
		if placeholder, exists := f.I18n.Placeholder[locale]; exists {
			return placeholder
		}
	}
	return f.Placeholder // Fallback to default
}

// GetLocalizedHelp returns localized help text for a field
func (f *Field) GetLocalizedHelp(locale string) string {
	if f.I18n != nil && f.I18n.Help != nil {
		if help, exists := f.I18n.Help[locale]; exists {
			return help
		}
	}
	return f.Help // Fallback to default
}

// GetLocalizedTooltip returns localized tooltip for a field
func (f *Field) GetLocalizedTooltip(locale string) string {
	if f.I18n != nil && f.I18n.Tooltip != nil {
		if tooltip, exists := f.I18n.Tooltip[locale]; exists {
			return tooltip
		}
	}
	return f.Tooltip // Fallback to default
}

// IsRTLForSchemaI18n checks if the given locale requires right-to-left text direction using SchemaI18n
func IsRTLForSchemaI18n(schemaI18n *SchemaI18n, locale string) bool {
	if schemaI18n != nil && schemaI18n.Direction != nil {
		if direction, exists := schemaI18n.Direction[locale]; exists {
			return direction == "rtl"
		}
	}

	// Common RTL languages
	rtlLocales := map[string]bool{
		"ar": true, // Arabic
		"he": true, // Hebrew
		"fa": true, // Persian
		"ur": true, // Urdu
		"yi": true, // Yiddish
	}

	return rtlLocales[locale]
}

// IsRTL checks if the given locale requires right-to-left text direction using existing I18n
func (s *Schema) IsRTL(locale string) bool {
	if s.I18n != nil && s.I18n.Direction != "" {
		return s.I18n.Direction == "rtl"
	}

	// Common RTL languages
	rtlLocales := map[string]bool{
		"ar": true, // Arabic
		"he": true, // Hebrew
		"fa": true, // Persian
		"ur": true, // Urdu
		"yi": true, // Yiddish
	}

	return rtlLocales[locale]
}

// CreateDefaultValidationMessages creates default English validation messages
func CreateDefaultValidationMessages() map[string]string {
	return map[string]string{
		"validation.required":         "{{.field}} is required",
		"validation.invalid_type":     "{{.field}} must be a valid {{.type}}",
		"validation.min_length":       "{{.field}} must be at least {{.min}} characters",
		"validation.max_length":       "{{.field}} cannot exceed {{.max}} characters",
		"validation.invalid_email":    "Please enter a valid email address",
		"validation.invalid_url":      "Please enter a valid URL",
		"validation.invalid_uuid":     "Please enter a valid UUID",
		"validation.invalid_date":     "Please enter a valid date",
		"validation.invalid_time":     "Please enter a valid time",
		"validation.invalid_datetime": "Please enter a valid date and time",
		"validation.min_value":        "{{.field}} must be at least {{.min}}",
		"validation.max_value":        "{{.field}} cannot exceed {{.max}}",
		"validation.invalid_range":    "{{.field}} must be between {{.min}} and {{.max}}",
		"validation.invalid_step":     "{{.field}} must be a multiple of {{.step}}",
		"validation.invalid_integer":  "{{.field}} must be a whole number",
		"validation.must_be_positive": "{{.field}} must be positive",
		"validation.must_be_negative": "{{.field}} must be negative",
		"validation.invalid_multiple": "{{.field}} must be a multiple of {{.value}}",
		"validation.min_items":        "{{.field}} must have at least {{.min}} items",
		"validation.max_items":        "{{.field}} cannot have more than {{.max}} items",
		"validation.unique_items":     "All items in {{.field}} must be unique",
		"validation.file_size":        "File size cannot exceed {{.size}}",
		"validation.file_type":        "File type not allowed",
		"validation.invalid_phone":    "Please enter a valid phone number",
		"validation.field_not_found":  "Field {{.field}} not found",

		// Common action messages
		"actions.save":     "Save",
		"actions.cancel":   "Cancel",
		"actions.submit":   "Submit",
		"actions.delete":   "Delete",
		"actions.edit":     "Edit",
		"actions.add":      "Add",
		"actions.remove":   "Remove",
		"actions.clear":    "Clear",
		"actions.back":     "Back",
		"actions.next":     "Next",
		"actions.previous": "Previous",
		"actions.finish":   "Finish",

		// Status messages
		"messages.save_success":      "Changes saved successfully",
		"messages.save_error":        "Failed to save changes",
		"messages.validation_failed": "Please correct the errors below",
		"messages.loading":           "Loading...",
		"messages.saving":            "Saving...",
		"messages.saved":             "Saved",
		"messages.error":             "Error",
		"messages.success":           "Success",
		"messages.warning":           "Warning",
		"messages.info":              "Info",
	}
}

// CreateSpanishValidationMessages creates Spanish validation messages
func CreateSpanishValidationMessages() map[string]string {
	return map[string]string{
		"validation.required":         "{{.field}} es requerido",
		"validation.invalid_type":     "{{.field}} debe ser un {{.type}} válido",
		"validation.min_length":       "{{.field}} debe tener al menos {{.min}} caracteres",
		"validation.max_length":       "{{.field}} no puede exceder {{.max}} caracteres",
		"validation.invalid_email":    "Por favor ingresa un email válido",
		"validation.invalid_url":      "Por favor ingresa una URL válida",
		"validation.invalid_uuid":     "Por favor ingresa un UUID válido",
		"validation.invalid_date":     "Por favor ingresa una fecha válida",
		"validation.invalid_time":     "Por favor ingresa una hora válida",
		"validation.invalid_datetime": "Por favor ingresa una fecha y hora válida",
		"validation.min_value":        "{{.field}} debe ser al menos {{.min}}",
		"validation.max_value":        "{{.field}} no puede exceder {{.max}}",
		"validation.invalid_range":    "{{.field}} debe estar entre {{.min}} y {{.max}}",
		"validation.invalid_step":     "{{.field}} debe ser un múltiplo de {{.step}}",
		"validation.invalid_integer":  "{{.field}} debe ser un número entero",
		"validation.must_be_positive": "{{.field}} debe ser positivo",
		"validation.must_be_negative": "{{.field}} debe ser negativo",
		"validation.invalid_multiple": "{{.field}} debe ser un múltiplo de {{.value}}",
		"validation.min_items":        "{{.field}} debe tener al menos {{.min}} elementos",
		"validation.max_items":        "{{.field}} no puede tener más de {{.max}} elementos",
		"validation.unique_items":     "Todos los elementos en {{.field}} deben ser únicos",
		"validation.file_size":        "El tamaño del archivo no puede exceder {{.size}}",
		"validation.file_type":        "Tipo de archivo no permitido",
		"validation.invalid_phone":    "Por favor ingresa un número de teléfono válido",
		"validation.field_not_found":  "Campo {{.field}} no encontrado",

		// Common action messages
		"actions.save":     "Guardar",
		"actions.cancel":   "Cancelar",
		"actions.submit":   "Enviar",
		"actions.delete":   "Eliminar",
		"actions.edit":     "Editar",
		"actions.add":      "Agregar",
		"actions.remove":   "Remover",
		"actions.clear":    "Limpiar",
		"actions.back":     "Atrás",
		"actions.next":     "Siguiente",
		"actions.previous": "Anterior",
		"actions.finish":   "Finalizar",

		// Status messages
		"messages.save_success":      "Cambios guardados exitosamente",
		"messages.save_error":        "Error al guardar cambios",
		"messages.validation_failed": "Por favor corrige los errores abajo",
		"messages.loading":           "Cargando...",
		"messages.saving":            "Guardando...",
		"messages.saved":             "Guardado",
		"messages.error":             "Error",
		"messages.success":           "Éxito",
		"messages.warning":           "Advertencia",
		"messages.info":              "Información",
	}
}

// LoadDefaultTranslations loads default English and Spanish translations
func (i *I18nManager) LoadDefaultTranslations() error {
	// Load English translations
	i.mu.Lock()
	i.messages["en"] = CreateDefaultValidationMessages()
	i.messages["es"] = CreateSpanishValidationMessages()
	i.mu.Unlock()

	return nil
}

