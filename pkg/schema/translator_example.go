package schema

import (
	"context"
	"fmt"

	"github.com/niiniyare/erp/pkg/schema/enrich"
	"github.com/niiniyare/erp/pkg/schema/runtime"
)

// Example: Complete translation workflow demonstrating the new embedded system

func ExampleTranslationWorkflow() {
	// 1. UI developer creates schema with English field IDs (no manual translations needed)
	schema := NewBuilder("user-form", TypeForm, "User Registration").
		AddTextField("firstName", "First Name", true).     // Will auto-translate based on "firstName" key
		AddTextField("lastName", "Last Name", true).       // Will auto-translate based on "lastName" key
		AddEmailField("email", "Email Address", true).     // Will auto-translate based on "email" key
		AddPasswordField("password", "Password", true).    // Will auto-translate based on "password" key
		AddSubmitButton("Create User").
		Build(context.Background())

	// 2. User with Spanish preference
	user := &ExampleUser{
		ID:              "user123",
		TenantID:        "tenant456", 
		PreferredLocale: "es",
		Permissions:     []string{"user:create"},
	}

	// 3. Browser sends Accept-Language header or user preference
	userBrowserLocales := []string{"es-ES", "es", "en"}

	// 4. Enricher automatically localizes schema based on user preference
	enricher := enrich.NewEnricher()
	localizedSchema, err := enricher.EnrichWithLocale(context.Background(), schema, user, "")
	if err != nil {
		panic(err)
	}

	// At this point:
	// - localizedSchema.Fields[0].Label = "Nombre" (from embedded es.json)
	// - localizedSchema.Fields[1].Label = "Apellido"
	// - localizedSchema.Fields[2].Label = "Correo Electrónico"  
	// - localizedSchema.Fields[3].Label = "Contraseña"

	// 5. Create runtime with locale detection
	rt := runtime.NewRuntimeBuilder(localizedSchema).
		WithLocale("es").
		Build(context.Background())

	// Or auto-detect from browser preferences
	detectedLocale := rt.DetectUserLocale(userBrowserLocales) // Returns "es"
	fmt.Printf("Detected locale: %s\n", detectedLocale)

	// 6. Runtime automatically provides localized validation messages
	rt.HandleFieldChange(context.Background(), "email", "invalid-email")
	errors := rt.GetErrors()["email"]
	// errors[0] = "Por favor ingresa un email válido" (localized validation)

	// 7. UI renderer gets fully localized schema
	for _, field := range localizedSchema.Fields {
		fmt.Printf("Field %s: %s\n", field.Name, field.Label)
	}

	// Output:
	// Field firstName: Nombre
	// Field lastName: Apellido  
	// Field email: Correo Electrónico
	// Field password: Contraseña
}

// ExampleUser implements the User interface with locale preference
type ExampleUser struct {
	ID              string
	TenantID        string
	PreferredLocale string
	Permissions     []string
	Roles           []string
}

func (u *ExampleUser) GetID() string                    { return u.ID }
func (u *ExampleUser) GetTenantID() string              { return u.TenantID }
func (u *ExampleUser) GetPermissions() []string         { return u.Permissions }
func (u *ExampleUser) GetRoles() []string               { return u.Roles }
func (u *ExampleUser) GetPreferredLocale() string       { return u.PreferredLocale }
func (u *ExampleUser) HasPermission(perm string) bool   { 
	for _, p := range u.Permissions {
		if p == perm { return true }
	}
	return false
}
func (u *ExampleUser) HasRole(role string) bool { 
	for _, r := range u.Roles {
		if r == role { return true }
	}
	return false
}

// Example: UI renderer integration with new translation system
func ExampleUIRendererIntegration() {
	// UI framework (React, Templ, etc.) receives localized schema
	schema := getLocalizedSchema("fr") // French
	
	// All field labels are already translated:
	// schema.Fields[0].Label = "Prénom" (firstName in French)
	// schema.Fields[1].Label = "Nom de famille" (lastName in French)
	
	// Render form with localized content
	html := renderForm(schema)
	fmt.Printf("Rendered form: %s\n", html)
}

func getLocalizedSchema(locale string) *Schema {
	// This would typically be called by your API handler
	schema := createBaseSchema()
	user := &ExampleUser{PreferredLocale: locale}
	
	enricher := enrich.NewEnricher()
	localizedSchema, _ := enricher.EnrichWithLocale(context.Background(), schema, user, locale)
	
	return localizedSchema
}

func createBaseSchema() *Schema {
	return NewBuilder("user-form", TypeForm, "User Form").
		AddTextField("firstName", "First Name", true).
		AddTextField("lastName", "Last Name", true).
		AddEmailField("email", "Email", true).
		Build(context.Background())
}

func renderForm(schema *Schema) string {
	// Simulated form rendering - in real implementation this would be
	// your UI framework (React, Templ, Go templates, etc.)
	html := "<form>"
	for _, field := range schema.Fields {
		html += fmt.Sprintf(`<div><label>%s</label><input name="%s" type="%s"></div>`, 
			field.Label, field.Name, field.Type)
	}
	html += "</form>"
	return html
}

// Example: Adding new locale support
func ExampleAddingNewLocale() {
	// 1. Check available locales
	available := Translator.GetAvailableLocales()
	fmt.Printf("Available locales: %v\n", available) // [en, es, fr, ar]

	// 2. Check if specific locale is supported
	hasGerman := T_HasLocale("de")
	fmt.Printf("German supported: %t\n", hasGerman) // false

	// 3. Detect best locale from user preferences
	userPrefs := []string{"de-DE", "de", "fr", "en"}
	bestLocale := T_DetectLocale(userPrefs)
	fmt.Printf("Best available locale: %s\n", bestLocale) // "fr" (since "de" not available)

	// 4. Check RTL support
	isRTL := T_IsRTL("ar")
	direction := T_Direction("ar")
	fmt.Printf("Arabic RTL: %t, direction: %s\n", isRTL, direction) // true, "rtl"
}

// Example: Custom field translations
func ExampleCustomFieldTranslations() {
	// For custom fields not in embedded translations, you can still use explicit I18n
	field := Field{
		Name:  "customBusinessField",
		Type:  FieldText,
		Label: "Custom Business Field", // English fallback
		I18n: &FieldI18n{
			Label: map[string]string{
				"en": "Custom Business Field",
				"es": "Campo de Negocio Personalizado",
				"fr": "Champ Commercial Personnalisé",
			},
		},
	}

	// GetLocalizedLabel will check I18n first, then fallback to embedded translations
	spanishLabel := field.GetLocalizedLabel("es")
	fmt.Printf("Spanish label: %s\n", spanishLabel) // "Campo de Negocio Personalizado"

	// If field ID exists in embedded translations, that takes precedence over humanization
	standardLabel := field.GetLocalizedLabel("fr")
	fmt.Printf("French label: %s\n", standardLabel) // Uses I18n map since no "customBusinessField" in embedded
}