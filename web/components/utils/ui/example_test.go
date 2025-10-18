package ui_test

import (
	"context"
	"testing"

	"github.com/niiniyare/erp/pkg/schema/ui"
)

func TestUISchemaCreation(t *testing.T) {
	// Create a registry
	registry := ui.NewRegistry()

	// Test creating an input component
	inputComponent := ui.CreateInput("email-input", "email").
		WithLabel("Email Address").
		WithName("email").
		Required().
		Build()

	if inputComponent.Type != ui.ComponentInput {
		t.Errorf("Expected input component type, got %v", inputComponent.Type)
	}

	if !inputComponent.Required {
		t.Error("Expected component to be required")
	}

	// Test creating a select component
	options := []ui.Option{
		{Value: "option1", Label: "Option 1"},
		{Value: "option2", Label: "Option 2"},
	}

	selectComponent := ui.CreateSelect("status-select", options).
		WithLabel("Status").
		WithName("status").
		Build()

	if selectComponent.Type != ui.ComponentSelect {
		t.Errorf("Expected select component type, got %v", selectComponent.Type)
	}

	// Test registry validation
	ctx := context.Background()
	if err := registry.Validate(ctx, inputComponent); err != nil {
		t.Errorf("Input validation failed: %v", err)
	}

	if err := registry.Validate(ctx, selectComponent); err != nil {
		t.Errorf("Select validation failed: %v", err)
	}
}

func TestFormCreation(t *testing.T) {
	registry := ui.NewRegistry()
	ctx := context.Background()

	// Create form with child components
	form := ui.NewComponent(ui.ComponentForm, "login-form").
		WithLabel("Login Form").
		WithChildren(
			ui.CreateInput("username", "text").
				WithLabel("Username").
				WithName("username").
				Required().
				Build(),
			ui.CreateInput("password", "password").
				WithLabel("Password").
				WithName("password").
				Required().
				Build(),
			ui.CreateButton("submit-btn", "Login").
				WithVariant(ui.VariantPrimary).
				Build(),
		).
		Build()

	// Test form has correct children
	if len(form.Children) != 3 {
		t.Errorf("Expected 3 children, got %d", len(form.Children))
	}

	// Test validation
	errors := ui.ValidateComponent(ctx, registry, form)
	if len(errors) > 0 {
		t.Errorf("Form validation failed: %v", errors)
	}
}

func TestTableCreation(t *testing.T) {
	columns := []ui.TableColumn{
		{
			Key:        "id",
			Title:      "ID",
			DataType:   ui.DataTypeNumber,
			Sortable:   true,
			Width:      "80px",
		},
		{
			Key:        "name",
			Title:      "Name",
			DataType:   ui.DataTypeText,
			Sortable:   true,
			Filterable: true,
		},
		{
			Key:      "email",
			Title:    "Email",
			DataType: ui.DataTypeText,
		},
		{
			Key:      "status",
			Title:    "Status",
			DataType: ui.DataTypeBadge,
			Align:    ui.AlignCenter,
		},
	}

	table := ui.CreateTable("users-table", columns).
		WithLabel("Users Table").
		Build()

	if table.Type != ui.ComponentTable {
		t.Errorf("Expected table component type, got %v", table.Type)
	}

	// Test that we can find specific components
	found := ui.FindComponentByID(table, "users-table")
	if found == nil {
		t.Error("Could not find table component by ID")
	}
}

func TestComponentTree(t *testing.T) {
	// Create a complex component tree
	dashboard := ui.NewComponent(ui.ComponentContainer, "dashboard").
		WithChildren(
			ui.CreateCard("stats-card").
				WithLabel("Statistics").
				WithChildren(
					ui.CreateChart("revenue-chart", ui.ChartLine).
						WithLabel("Revenue Chart").
						Build(),
				).
				Build(),
			ui.CreateTable("data-table", []ui.TableColumn{
				{Key: "name", Title: "Name", DataType: ui.DataTypeText},
			}).Build(),
		).
		Build()

	// Test walking the component tree
	var componentCount int
	ui.WalkComponents(dashboard, func(c ui.Component) error {
		componentCount++
		return nil
	})

	if componentCount != 4 { // container + card + chart + table
		t.Errorf("Expected 4 components in tree, got %d", componentCount)
	}

	// Test finding components by type
	charts := ui.FindComponentsByType(dashboard, ui.ComponentChart)
	if len(charts) != 1 {
		t.Errorf("Expected 1 chart component, got %d", len(charts))
	}
}

func ExampleComponentBuilder() {
	// Create a complete login form
	loginForm := ui.NewComponent(ui.ComponentForm, "login-form").
		WithLabel("User Login").
		WithConfig(ui.FormConfig{
			Method: "POST",
			Action: "/auth/login",
			Layout: ui.FormLayoutVertical,
		}).
		WithChildren(
			ui.CreateInput("username", "text").
				WithLabel("Username").
				WithName("username").
				WithDescription("Enter your username").
				Required().
				Build(),

			ui.CreateInput("password", "password").
				WithLabel("Password").
				WithName("password").
				Required().
				Build(),

			ui.CreateButton("submit", "Login").
				WithVariant(ui.VariantPrimary).
				WithSize(ui.SizeLG).
				WithConfig(ui.ButtonConfig{
					ButtonType: ui.ButtonSubmit,
				}).
				Build(),
		).
		Build()

	// Use the form component
	_ = loginForm
}