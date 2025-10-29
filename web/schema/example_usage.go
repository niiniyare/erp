package schema

import (
	"context"
	"fmt"
	"log"
	"strings"
)

// ExampleUsage demonstrates how to use the templ renderer
func ExampleUsage() {
	// Create a token resolver with default tokens
	tokens := NewDefaultTokenResolver()

	// Create a renderer registry
	registry := NewRendererRegistry(tokens)

	// Create a templ renderer
	templRenderer := NewTemplRenderer(registry, tokens)

	// Create a simple form schema
	form := NewFormSchema("user-registration", "User Registration").
		SetDescription("Create a new user account").
		AddTextField("name", "Full Name", "Enter your full name", true).
		AddEmailField("email", "Email Address", true).
		AddPasswordField("password", "Password", true).
		AddSubmitAction("Create Account", "primary")

	// Sample form data
	data := map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
	}

	// Sample validation errors
	errors := map[string][]string{
		"password": {"Password must be at least 8 characters"},
	}

	// Render the form using templ
	ctx := context.Background()
	html, err := templRenderer.RenderFormWithErrors(ctx, form.Schema, data, errors)
	if err != nil {
		log.Printf("Error rendering form: %v", err)
		return
	}

	fmt.Println("Rendered HTML:")
	fmt.Println(html)

	// You can also get the templ component for composition
	component := templRenderer.RenderFormAsComponent(form.Schema, data, errors)
	
	// In a real Fiber handler, you might do:
	// return templ.Handler(component).ServeHTTP(c.Response(), c.Request())
	
	// Or render to string for custom handling
	var builder strings.Builder
	err = component.Render(ctx, &builder)
	if err != nil {
		log.Printf("Error rendering component: %v", err)
		return
	}

	fmt.Println("Component HTML:")
	fmt.Println(builder.String())
}

// ExampleFiberIntegration shows how to integrate with Fiber
func ExampleFiberIntegration() {
	// This would typically be in a Fiber handler
	
	// Setup
	tokens := NewDefaultTokenResolver()
	registry := NewRendererRegistry(tokens)
	templRenderer := NewTemplRenderer(registry, tokens)

	// Create form schema
	contactForm := NewFormSchema("contact-form", "Contact Us").
		SetDescription("Send us a message and we'll get back to you").
		AddTextField("name", "Full Name", "Enter your full name", true).
		AddEmailField("email", "Email Address", true).
		AddTextareaField("message", "Message", true).
		AddSubmitAction("Send Message", "primary").
		AddResetAction("Clear Form")

	// Use the variables to avoid unused variable errors
	_ = templRenderer
	_ = contactForm

	// In a Fiber handler, you would:
	/*
	app.Get("/contact", func(c *fiber.Ctx) error {
		component := templRenderer.RenderFormAsComponent(contactForm.Schema, nil, nil)
		return templ.Handler(component).ServeHTTP(c.Response(), c.Request())
	})

	app.Post("/contact", func(c *fiber.Ctx) error {
		// Parse form data
		data := map[string]any{
			"name":    c.FormValue("name"),
			"email":   c.FormValue("email"),
			"message": c.FormValue("message"),
		}

		// Validate data (using the validation system)
		validator := NewValidator()
		result := validator.ValidateData(c.Context(), contactForm.Schema, data)
		
		if !result.Valid {
			// Render form with errors
			component := templRenderer.RenderFormAsComponent(contactForm.Schema, data, result.FieldErrors)
			return templ.Handler(component).ServeHTTP(c.Response(), c.Request())
		}

		// Process successful submission
		// ... save to database, send email, etc.

		// Redirect or show success message
		return c.Redirect("/contact/success")
	})
	*/

	fmt.Println("Fiber integration example setup complete")
}

// ExampleComplexForm demonstrates a more complex form with sections and validation
func ExampleComplexForm() {
	tokens := NewDefaultTokenResolver()
	registry := NewRendererRegistry(tokens)
	templRenderer := NewTemplRenderer(registry, tokens)

	// Create a complex form with sections
	userForm := NewFormSchema("user-profile", "User Profile").
		SetDescription("Manage your account information")

	// Add layout with sections
	userForm.Schema.Layout = &Layout{
		Type:    LayoutSections,
		Columns: 1,
		Gap:     "2rem",
		Sections: []Section{
			{
				ID:          "personal-info",
				Title:       "Personal Information",
				Description: "Basic information about yourself",
				Fields:      []string{"first_name", "last_name", "email", "phone"},
				Collapsible: false,
			},
			{
				ID:          "preferences",
				Title:       "Preferences",
				Description: "Customize your experience",
				Fields:      []string{"language", "timezone", "notifications"},
				Collapsible: true,
				Collapsed:   false,
			},
		},
	}

	// Add fields
	userForm.Schema.Fields = []Field{
		{
			Name:        "first_name",
			Type:        FieldText,
			Label:       "First Name",
			Description: "Enter your first name",
			Required:    true,
			Validation: &FieldValidation{
				Required:  true,
				MinLength: IntPtr(2),
				MaxLength: IntPtr(50),
			},
		},
		{
			Name:        "last_name",
			Type:        FieldText,
			Label:       "Last Name",
			Description: "Enter your last name",
			Required:    true,
			Validation: &FieldValidation{
				Required:  true,
				MinLength: IntPtr(2),
				MaxLength: IntPtr(50),
			},
		},
		{
			Name:        "email",
			Type:        FieldEmail,
			Label:       "Email Address",
			Description: "Your primary email address",
			Required:    true,
			Validation: &FieldValidation{
				Required: true,
				Email:    true,
			},
		},
		{
			Name:        "phone",
			Type:        FieldPhone,
			Label:       "Phone Number",
			Description: "Your contact phone number",
			Required:    false,
		},
		{
			Name:        "language",
			Type:        FieldSelect,
			Label:       "Language",
			Description: "Preferred interface language",
			Required:    true,
			Options: []Option{
				{Value: "en", Label: "English", Selected: true},
				{Value: "es", Label: "Spanish"},
				{Value: "fr", Label: "French"},
			},
		},
		{
			Name:        "timezone",
			Type:        FieldSelect,
			Label:       "Timezone",
			Description: "Your local timezone",
			Required:    true,
			DataSource: &DataSource{
				Type: DataSourceAPI,
				URL:  "/api/timezones",
			},
		},
		{
			Name:        "notifications",
			Type:        FieldCheckbox,
			Label:       "Email Notifications",
			Description: "Receive email notifications for updates",
			Required:    false,
		},
	}

	// Add actions
	userForm.Schema.Actions = []Action{
		{
			ID:      "save",
			Type:    ActionSubmit,
			Text:    "Save Profile",
			Variant: "primary",
		},
		{
			ID:      "cancel",
			Type:    ActionButton,
			Text:    "Cancel",
			Variant: "outline",
			Config: &ActionConfig{
				URL: "/profile",
			},
		},
	}

	// Sample data and errors
	data := map[string]any{
		"first_name":    "John",
		"last_name":     "Doe",
		"email":         "john@example.com",
		"language":      "en",
		"notifications": true,
	}

	errors := map[string][]string{
		"phone": {"Invalid phone number format"},
	}

	// Render the complex form
	ctx := context.Background()
	html, err := templRenderer.RenderFormWithErrors(ctx, userForm.Schema, data, errors)
	if err != nil {
		log.Printf("Error rendering complex form: %v", err)
		return
	}

	fmt.Println("Complex Form HTML rendered successfully")
	fmt.Printf("HTML length: %d characters\n", len(html))
}