package schema

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

type I18nTestSuite struct {
	suite.Suite
	ctx     context.Context
	tempDir string
}

func (suite *I18nTestSuite) SetupTest() {
	suite.ctx = context.Background()

	// Create temporary directory for translation files
	tempDir, err := os.MkdirTemp("", "i18n-test-")
	suite.Require().NoError(err)
	suite.tempDir = tempDir
}

func (suite *I18nTestSuite) TearDownTest() {
	if suite.tempDir != "" {
		os.RemoveAll(suite.tempDir)
	}
}

func TestI18nTestSuite(t *testing.T) {
	suite.Run(t, new(I18nTestSuite))
}

// Test I18nManager creation and basic functionality
func (suite *I18nTestSuite) TestI18nManagerCreation() {
	manager := NewI18nManager("en", "en")
	suite.Require().NotNil(manager)
	suite.Require().Equal("en", manager.GetLocale())

	// Test locale change
	manager.SetLocale("es")
	suite.Require().Equal("es", manager.GetLocale())
}

// Test loading default translations
func (suite *I18nTestSuite) TestLoadDefaultTranslations() {
	manager := NewI18nManager("en", "en")
	err := manager.LoadDefaultTranslations()
	suite.Require().NoError(err)

	// Test English validation message
	message := manager.GetValidationMessage("required", map[string]any{
		"field": "Email",
	})
	suite.Require().Equal("Email is required", message)

	// Change to Spanish and test
	manager.SetLocale("es")
	message = manager.GetValidationMessage("required", map[string]any{
		"field": "Email",
	})
	suite.Require().Equal("Email es requerido", message)
}

// Test field localization methods
func (suite *I18nTestSuite) TestFieldLocalization() {
	field := &Field{
		Name:        "first_name",
		Type:        FieldText,
		Label:       "First Name",
		Placeholder: "Enter first name",
		Help:        "Your legal first name",
		Tooltip:     "This is required",
		I18n: &FieldI18n{
			Label: map[string]string{
				"en": "First Name",
				"es": "Nombre",
				"fr": "Prénom",
			},
			Placeholder: map[string]string{
				"en": "Enter first name",
				"es": "Ingresa tu nombre",
				"fr": "Entrez votre prénom",
			},
			Help: map[string]string{
				"en": "Your legal first name",
				"es": "Tu nombre legal",
				"fr": "Votre prénom légal",
			},
			Tooltip: map[string]string{
				"en": "This is required",
				"es": "Esto es requerido",
				"fr": "Ceci est requis",
			},
		},
	}

	// Test English (default)
	suite.Require().Equal("First Name", field.GetLocalizedLabel("en"))
	suite.Require().Equal("Enter first name", field.GetLocalizedPlaceholder("en"))
	suite.Require().Equal("Your legal first name", field.GetLocalizedHelp("en"))
	suite.Require().Equal("This is required", field.GetLocalizedTooltip("en"))

	// Test Spanish
	suite.Require().Equal("Nombre", field.GetLocalizedLabel("es"))
	suite.Require().Equal("Ingresa tu nombre", field.GetLocalizedPlaceholder("es"))
	suite.Require().Equal("Tu nombre legal", field.GetLocalizedHelp("es"))
	suite.Require().Equal("Esto es requerido", field.GetLocalizedTooltip("es"))

	// Test French
	suite.Require().Equal("Prénom", field.GetLocalizedLabel("fr"))
	suite.Require().Equal("Entrez votre prénom", field.GetLocalizedPlaceholder("fr"))
	suite.Require().Equal("Votre prénom légal", field.GetLocalizedHelp("fr"))
	suite.Require().Equal("Ceci est requis", field.GetLocalizedTooltip("fr"))

	// Test fallback for unsupported locale
	suite.Require().Equal("First Name", field.GetLocalizedLabel("de"))
}

// Test schema localization
func (suite *I18nTestSuite) TestSchemaLocalization() {
	schema := &Schema{
		ID:          "user-form",
		Title:       "User Registration",
		Description: "Create your account",
		Fields: []Field{
			{
				Name:        "email",
				Type:        FieldEmail,
				Label:       "Email",
				Placeholder: "Enter email",
				I18n: &FieldI18n{
					Label: map[string]string{
						"en": "Email",
						"es": "Correo Electrónico",
					},
					Placeholder: map[string]string{
						"en": "Enter email",
						"es": "Ingresa tu correo",
					},
				},
			},
		},
	}

	// Create separate SchemaI18n for testing
	schemaI18n := &SchemaI18n{
		Title: map[string]string{
			"en": "User Registration",
			"es": "Registro de Usuario",
			"fr": "Inscription Utilisateur",
		},
		Description: map[string]string{
			"en": "Create your account",
			"es": "Crea tu cuenta",
			"fr": "Créez votre compte",
		},
		Direction: map[string]string{
			"ar": "rtl",
			"he": "rtl",
		},
	}

	// Test English localization
	localizedSchema, err := schema.ApplySchemaI18nLocalization(schemaI18n, "en")
	suite.Require().NoError(err)
	suite.Require().Equal("User Registration", localizedSchema.Title)
	suite.Require().Equal("Create your account", localizedSchema.Description)
	suite.Require().Equal("Email", localizedSchema.Fields[0].Label)
	suite.Require().Equal("Enter email", localizedSchema.Fields[0].Placeholder)

	// Test Spanish localization
	localizedSchema, err = schema.ApplySchemaI18nLocalization(schemaI18n, "es")
	suite.Require().NoError(err)
	suite.Require().Equal("Registro de Usuario", localizedSchema.Title)
	suite.Require().Equal("Crea tu cuenta", localizedSchema.Description)
	suite.Require().Equal("Correo Electrónico", localizedSchema.Fields[0].Label)
	suite.Require().Equal("Ingresa tu correo", localizedSchema.Fields[0].Placeholder)

	// Test French localization (partial)
	localizedSchema, err = schema.ApplySchemaI18nLocalization(schemaI18n, "fr")
	suite.Require().NoError(err)
	suite.Require().Equal("Inscription Utilisateur", localizedSchema.Title)
	suite.Require().Equal("Créez votre compte", localizedSchema.Description)
	// Field should fall back to default since French is not defined
	suite.Require().Equal("Email", localizedSchema.Fields[0].Label)
	suite.Require().Equal("Enter email", localizedSchema.Fields[0].Placeholder)
}

// Test RTL language detection
func (suite *I18nTestSuite) TestRTLDetection() {
	schemaI18n := &SchemaI18n{
		Direction: map[string]string{
			"ar": "rtl",
			"he": "rtl",
			"en": "ltr",
		},
	}

	// Test explicit RTL
	suite.Require().True(IsRTLForSchemaI18n(schemaI18n, "ar"))
	suite.Require().True(IsRTLForSchemaI18n(schemaI18n, "he"))
	suite.Require().False(IsRTLForSchemaI18n(schemaI18n, "en"))

	// Test built-in RTL detection
	suite.Require().True(IsRTLForSchemaI18n(nil, "ar"))
	suite.Require().True(IsRTLForSchemaI18n(nil, "he"))
	suite.Require().True(IsRTLForSchemaI18n(nil, "fa"))
	suite.Require().True(IsRTLForSchemaI18n(nil, "ur"))
	suite.Require().False(IsRTLForSchemaI18n(nil, "en"))
	suite.Require().False(IsRTLForSchemaI18n(nil, "es"))
	suite.Require().False(IsRTLForSchemaI18n(nil, "fr"))

	// Test existing Schema.IsRTL method with I18n struct
	schema := &Schema{
		ID: "test-schema",
		I18n: &I18n{
			Direction: "rtl",
		},
	}
	suite.Require().True(schema.IsRTL("ar"))

	schemaLTR := &Schema{
		ID: "test-schema",
		I18n: &I18n{
			Direction: "ltr",
		},
	}
	suite.Require().False(schemaLTR.IsRTL("en"))
}

// Test I18nLoader file loading
func (suite *I18nTestSuite) TestI18nLoader() {
	// Create test translation files
	enTranslations := map[string]any{
		"validation": map[string]any{
			"required":   "{{.field}} is required",
			"min_length": "{{.field}} must be at least {{.min}} characters",
		},
		"actions": map[string]any{
			"save":   "Save",
			"cancel": "Cancel",
		},
	}

	esTranslations := map[string]any{
		"validation": map[string]any{
			"required":   "{{.field}} es requerido",
			"min_length": "{{.field}} debe tener al menos {{.min}} caracteres",
		},
		"actions": map[string]any{
			"save":   "Guardar",
			"cancel": "Cancelar",
		},
	}

	// Write translation files
	enData, err := json.Marshal(enTranslations)
	suite.Require().NoError(err)
	err = os.WriteFile(filepath.Join(suite.tempDir, "en.json"), enData, 0o644)
	suite.Require().NoError(err)

	esData, err := json.Marshal(esTranslations)
	suite.Require().NoError(err)
	err = os.WriteFile(filepath.Join(suite.tempDir, "es.json"), esData, 0o644)
	suite.Require().NoError(err)

	// Test loader
	loader := NewI18nLoader(suite.tempDir)

	// Load English
	enMessages, err := loader.LoadLocale("en")
	suite.Require().NoError(err)
	suite.Require().Equal("{{.field}} is required", enMessages["validation.required"])
	suite.Require().Equal("{{.field}} must be at least {{.min}} characters", enMessages["validation.min_length"])
	suite.Require().Equal("Save", enMessages["actions.save"])
	suite.Require().Equal("Cancel", enMessages["actions.cancel"])

	// Load Spanish
	esMessages, err := loader.LoadLocale("es")
	suite.Require().NoError(err)
	suite.Require().Equal("{{.field}} es requerido", esMessages["validation.required"])
	suite.Require().Equal("{{.field}} debe tener al menos {{.min}} caracteres", esMessages["validation.min_length"])
	suite.Require().Equal("Guardar", esMessages["actions.save"])
	suite.Require().Equal("Cancelar", esMessages["actions.cancel"])

	// Test caching - should not read file again
	cachedMessages, err := loader.LoadLocale("en")
	suite.Require().NoError(err)
	suite.Require().Equal(enMessages, cachedMessages)
}

// Test message interpolation
func (suite *I18nTestSuite) TestMessageInterpolation() {
	manager := NewI18nManager("en", "en")
	err := manager.LoadDefaultTranslations()
	suite.Require().NoError(err)

	// Test simple interpolation
	message := manager.GetValidationMessage("min_length", map[string]any{
		"field": "Password",
		"min":   8,
	})
	suite.Require().Equal("Password must be at least 8 characters", message)

	// Test range interpolation
	message = manager.GetValidationMessage("invalid_range", map[string]any{
		"field": "Age",
		"min":   18,
		"max":   65,
	})
	suite.Require().Equal("Age must be between 18 and 65", message)

	// Test message without parameters
	message = manager.GetValidationMessage("invalid_email", nil)
	suite.Require().Equal("Please enter a valid email address", message)
}

// Test fallback behavior
func (suite *I18nTestSuite) TestFallbackBehavior() {
	manager := NewI18nManager("de", "en") // German with English fallback
	err := manager.LoadDefaultTranslations()
	suite.Require().NoError(err)

	// Should fall back to English since German is not available
	message := manager.GetValidationMessage("required", map[string]any{
		"field": "Name",
	})
	suite.Require().Equal("Name is required", message)

	// Test unknown message key
	message = manager.GetValidationMessage("unknown_key", nil)
	suite.Require().Equal("unknown_key", message)
}

// Test action message retrieval
func (suite *I18nTestSuite) TestActionMessages() {
	manager := NewI18nManager("en", "en")
	err := manager.LoadDefaultTranslations()
	suite.Require().NoError(err)

	// Test English action messages
	suite.Require().Equal("Save", manager.GetMessage("actions.save", nil))
	suite.Require().Equal("Cancel", manager.GetMessage("actions.cancel", nil))
	suite.Require().Equal("Submit", manager.GetMessage("actions.submit", nil))
	suite.Require().Equal("Delete", manager.GetMessage("actions.delete", nil))

	// Test Spanish action messages
	manager.SetLocale("es")
	suite.Require().Equal("Guardar", manager.GetMessage("actions.save", nil))
	suite.Require().Equal("Cancelar", manager.GetMessage("actions.cancel", nil))
	suite.Require().Equal("Enviar", manager.GetMessage("actions.submit", nil))
	suite.Require().Equal("Eliminar", manager.GetMessage("actions.delete", nil))
}

// Test status messages
func (suite *I18nTestSuite) TestStatusMessages() {
	manager := NewI18nManager("en", "en")
	err := manager.LoadDefaultTranslations()
	suite.Require().NoError(err)

	// Test English status messages
	suite.Require().Equal("Changes saved successfully", manager.GetMessage("messages.save_success", nil))
	suite.Require().Equal("Failed to save changes", manager.GetMessage("messages.save_error", nil))
	suite.Require().Equal("Loading...", manager.GetMessage("messages.loading", nil))

	// Test Spanish status messages
	manager.SetLocale("es")
	suite.Require().Equal("Cambios guardados exitosamente", manager.GetMessage("messages.save_success", nil))
	suite.Require().Equal("Error al guardar cambios", manager.GetMessage("messages.save_error", nil))
	suite.Require().Equal("Cargando...", manager.GetMessage("messages.loading", nil))
}

// Test pluralization
func (suite *I18nTestSuite) TestPluralization() {
	manager := NewI18nManager("en", "en")

	// Add custom plural messages for testing
	manager.mu.Lock()
	if manager.messages["en"] == nil {
		manager.messages["en"] = make(map[string]string)
	}
	manager.messages["en"]["items.one"] = "{{.count}} item"
	manager.messages["en"]["items.other"] = "{{.count}} items"
	manager.mu.Unlock()

	// Test singular
	message := manager.GetPluralMessage("items", 1, "en")
	suite.Require().Equal("1 item", message)

	// Test plural
	message = manager.GetPluralMessage("items", 5, "en")
	suite.Require().Equal("5 items", message)

	// Test zero (should use "other")
	message = manager.GetPluralMessage("items", 0, "en")
	suite.Require().Equal("0 items", message)
}

// Test JSON flattening
func (suite *I18nTestSuite) TestJSONFlattening() {
	nested := map[string]any{
		"validation": map[string]any{
			"required": "Required field",
			"length": map[string]any{
				"min": "Too short",
				"max": "Too long",
			},
		},
		"actions": map[string]any{
			"save":   "Save",
			"cancel": "Cancel",
		},
		"simple": "Simple value",
	}

	flattened := flattenJSON(nested)

	suite.Require().Equal("Required field", flattened["validation.required"])
	suite.Require().Equal("Too short", flattened["validation.length.min"])
	suite.Require().Equal("Too long", flattened["validation.length.max"])
	suite.Require().Equal("Save", flattened["actions.save"])
	suite.Require().Equal("Cancel", flattened["actions.cancel"])
	suite.Require().Equal("Simple value", flattened["simple"])
}

// Test edge cases
func (suite *I18nTestSuite) TestEdgeCases() {
	manager := NewI18nManager("", "en") // Empty locale
	err := manager.LoadDefaultTranslations()
	suite.Require().NoError(err)

	// Should fall back to default locale
	message := manager.GetValidationMessage("required", map[string]any{
		"field": "Name",
	})
	suite.Require().Equal("Name is required", message)

	// Test nil parameters
	message = manager.GetValidationMessage("invalid_email", nil)
	suite.Require().Equal("Please enter a valid email address", message)

	// Test empty parameters
	message = manager.GetValidationMessage("required", map[string]any{})
	suite.Require().Contains(message, "is required")

	// Test field without I18n
	field := &Field{
		Name:  "test",
		Label: "Test Field",
	}
	suite.Require().Equal("Test Field", field.GetLocalizedLabel("es"))
	suite.Require().Equal("", field.GetLocalizedPlaceholder("es"))

	// Test schema without I18n
	schema := &Schema{
		ID:    "test",
		Title: "Test Schema",
	}
	localizedSchema, err := schema.ApplySchemaI18nLocalization(nil, "es")
	suite.Require().NoError(err)
	suite.Require().Equal("Test Schema", localizedSchema.Title)
}

// Test concurrent access
func (suite *I18nTestSuite) TestConcurrentAccess() {
	manager := NewI18nManager("en", "en")
	err := manager.LoadDefaultTranslations()
	suite.Require().NoError(err)

	// Test concurrent reads and writes
	done := make(chan bool, 10)

	// Concurrent readers
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				message := manager.GetValidationMessage("required", map[string]any{
					"field": "Test",
				})
				// Allow either English or Spanish message due to concurrent locale switching
				suite.Require().True(message == "Test is required" || message == "Test es requerido",
					"Expected 'Test is required' or 'Test es requerido', got: %s", message)
			}
			done <- true
		}()
	}

	// Concurrent locale switchers
	for i := 0; i < 5; i++ {
		go func(id int) {
			for j := 0; j < 50; j++ {
				if id%2 == 0 {
					manager.SetLocale("en")
				} else {
					manager.SetLocale("es")
				}
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify manager is still functional
	locale := manager.GetLocale()
	suite.Require().True(locale == "en" || locale == "es")
}

// Test complete localization workflow
func (suite *I18nTestSuite) TestCompleteWorkflow() {
	// Create a complete schema with I18n
	schema := &Schema{
		ID:          "user-registration",
		Title:       "User Registration",
		Description: "Please fill out the form",
		Fields: []Field{
			{
				Name:        "first_name",
				Type:        FieldText,
				Label:       "First Name",
				Placeholder: "Enter your first name",
				Help:        "Your legal first name",
				Required:    true,
				I18n: &FieldI18n{
					Label: map[string]string{
						"en": "First Name",
						"es": "Nombre",
						"fr": "Prénom",
						"ar": "الاسم الأول",
					},
					Placeholder: map[string]string{
						"en": "Enter your first name",
						"es": "Ingresa tu nombre",
						"fr": "Entrez votre prénom",
						"ar": "أدخل اسمك الأول",
					},
					Help: map[string]string{
						"en": "Your legal first name",
						"es": "Tu nombre legal",
						"fr": "Votre prénom légal",
						"ar": "اسمك القانوني الأول",
					},
				},
			},
			{
				Name:     "email",
				Type:     FieldEmail,
				Label:    "Email",
				Required: true,
				I18n: &FieldI18n{
					Label: map[string]string{
						"en": "Email Address",
						"es": "Dirección de Correo",
						"fr": "Adresse Email",
						"ar": "عنوان البريد الإلكتروني",
					},
				},
			},
		},
	}

	// Test localization for different languages
	testCases := []struct {
		locale              string
		expectedTitle       string
		expectedDescription string
		expectedFirstName   string
		expectedEmail       string
		isRTL               bool
	}{
		{
			locale:              "en",
			expectedTitle:       "User Registration",
			expectedDescription: "Please fill out the form",
			expectedFirstName:   "First Name",
			expectedEmail:       "Email Address",
			isRTL:               false,
		},
		{
			locale:              "es",
			expectedTitle:       "Registro de Usuario",
			expectedDescription: "Por favor llena el formulario",
			expectedFirstName:   "Nombre",
			expectedEmail:       "Dirección de Correo",
			isRTL:               false,
		},
		{
			locale:              "fr",
			expectedTitle:       "Inscription Utilisateur",
			expectedDescription: "Veuillez remplir le formulaire",
			expectedFirstName:   "Prénom",
			expectedEmail:       "Adresse Email",
			isRTL:               false,
		},
		{
			locale:              "ar",
			expectedTitle:       "تسجيل المستخدم",
			expectedDescription: "يرجى ملء النموذج",
			expectedFirstName:   "الاسم الأول",
			expectedEmail:       "عنوان البريد الإلكتروني",
			isRTL:               true,
		},
	}

	// Create separate SchemaI18n for the complete workflow test
	schemaI18n := &SchemaI18n{
		Title: map[string]string{
			"en": "User Registration",
			"es": "Registro de Usuario",
			"fr": "Inscription Utilisateur",
			"ar": "تسجيل المستخدم",
		},
		Description: map[string]string{
			"en": "Please fill out the form",
			"es": "Por favor llena el formulario",
			"fr": "Veuillez remplir le formulaire",
			"ar": "يرجى ملء النموذج",
		},
		Direction: map[string]string{
			"ar": "rtl",
		},
	}

	for _, tc := range testCases {
		suite.Run("locale_"+tc.locale, func() {
			localizedSchema, err := schema.ApplySchemaI18nLocalization(schemaI18n, tc.locale)
			suite.Require().NoError(err)

			// Test schema-level localization
			suite.Require().Equal(tc.expectedTitle, localizedSchema.Title)
			suite.Require().Equal(tc.expectedDescription, localizedSchema.Description)

			// Test field-level localization
			firstNameField := localizedSchema.Fields[0]
			emailField := localizedSchema.Fields[1]

			suite.Require().Equal(tc.expectedFirstName, firstNameField.Label)
			suite.Require().Equal(tc.expectedEmail, emailField.Label)

			// Test RTL detection using SchemaI18n
			suite.Require().Equal(tc.isRTL, IsRTLForSchemaI18n(schemaI18n, tc.locale))

			// Test localized getters
			suite.Require().Equal(tc.expectedFirstName, firstNameField.GetLocalizedLabel(tc.locale))
			suite.Require().Equal(tc.expectedEmail, emailField.GetLocalizedLabel(tc.locale))
		})
	}
}

func TestI18nManager_interpolateMessage(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		locale         string
		fallbackLocale string
		// Named input parameters for target function.
		message string
		params  map[string]any
		want    string
	}{
		{
			name:           "simple interpolation",
			locale:         "en-US",
			fallbackLocale: "en",
			message:        "Hello {{name}}",
			params:         map[string]any{"name": "John"},
			want:           "Hello John",
		},
		{
			name:           "multiple params",
			locale:         "en-US", 
			fallbackLocale: "en",
			message:        "{{field}} must be between {{min}} and {{max}}",
			params:         map[string]any{"field": "Age", "min": 18, "max": 65},
			want:           "Age must be between 18 and 65",
		},
		{
			name:           "no interpolation needed",
			locale:         "en-US",
			fallbackLocale: "en", 
			message:        "This is a simple message",
			params:         map[string]any{},
			want:           "This is a simple message",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := NewI18nManager(tt.locale, tt.fallbackLocale)
			got := i.interpolateMessage(tt.message, tt.params)
			if got != tt.want {
				t.Errorf("interpolateMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

