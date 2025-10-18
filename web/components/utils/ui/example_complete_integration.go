package ui

import (
	"context"
	"fmt"

	"github.com/niiniyare/erp/pkg/schema/ui/css"
)

// DemoStyledComponents demonstrates the complete CSS integration system
func DemoStyledComponents() {
	schemaDir := "../../../docs/ui/Schema"

	fmt.Println("=== ERP UI Schema with CSS Integration Demo ===")

	// 1. Create a styled login form
	loginForm := createStyledLoginForm(schemaDir)
	fmt.Println("1. Styled Login Form:")
	printComponentInfo(loginForm)
	fmt.Println()

	// 2. Create a styled dashboard card
	dashboardCard := createStyledDashboardCard(schemaDir)
	fmt.Println("2. Styled Dashboard Card:")
	printComponentInfo(dashboardCard)
	fmt.Println()

	// 3. Create a styled data table
	dataTable := createStyledDataTable(schemaDir)
	fmt.Println("3. Styled Data Table:")
	printComponentInfo(dataTable)
	fmt.Println()

	// 4. Create responsive layout
	responsiveLayout := createResponsiveLayout(schemaDir)
	fmt.Println("4. Responsive Layout:")
	printComponentInfo(responsiveLayout)
	fmt.Println()

	// 5. Demonstrate theme-aware components
	darkModeCard := createThemeAwareCard("dark", schemaDir)
	fmt.Println("5. Dark Mode Card:")
	printComponentInfo(darkModeCard)
	fmt.Println()

	// 6. Validate all components
	fmt.Println("6. Component Validation:")
	validateComponents([]Component{loginForm, dashboardCard, dataTable, responsiveLayout, darkModeCard}, schemaDir)

	// 7. Generate stylesheet
	fmt.Println("\n7. Generated Stylesheet:")
	stylesheet := GenerateStylesheet([]Component{loginForm, dashboardCard, dataTable})
	fmt.Println(stylesheet[:min(500, len(stylesheet))] + "...")
}

// createStyledLoginForm creates a complete login form with CSS styling
func createStyledLoginForm(schemaDir string) Component {
	// Create form container with styling
	return NewComponent(ComponentForm, "login-form").
		WithLabel("User Login").
		WithStyledVariant("default", schemaDir).
		WithCustomCSS("max-width", "400px").
		WithCustomCSS("margin", "0 auto").
		WithCustomCSS("padding", "2rem").
		WithChildren(
			// Email input with error state styling
			CreateStyledInput("email", "email", "default", schemaDir).
				WithLabel("Email Address").
				WithName("email").
				WithPlaceholder("Enter your email").
				Required().
				WithValidator(&Validator{
					Required: true,
					Pattern:  `^[^\s@]+@[^\s@]+\.[^\s@]+$`,
					Message:  "Please enter a valid email address",
				}).
				Build(),

			// Password input with custom styling
			CreateStyledInput("password", "password", "default", schemaDir).
				WithLabel("Password").
				WithName("password").
				WithPlaceholder("Enter your password").
				Required().
				WithValidator(&Validator{
					Required:  true,
					MinLength: func() *int { i := 8; return &i }(),
					Message:   "Password must be at least 8 characters",
				}).
				Build(),

			// Submit button with primary styling
			CreateStyledButton("login-btn", "Sign In", "primary", schemaDir).
				WithCustomCSS("width", "100%").
				WithCustomCSS("margin-top", "1rem").
				Build(),

			// Link for forgot password
			CreateStyledButton("forgot-link", "Forgot Password?", "outline", schemaDir).
				WithCustomCSS("width", "100%").
				WithCustomCSS("margin-top", "0.5rem").
				Build(),
		).
		Build()
}

// createStyledDashboardCard creates a dashboard statistics card
func createStyledDashboardCard(schemaDir string) Component {
	factory := css.NewFactory(schemaDir)

	// Custom card styling with hover effects
	cardStyles := factory.CardStyles("md").
		WithPadding("1.5rem").
		WithCustomProperty("transition", "transform 0.2s ease-in-out").
		WithCustomProperty("--hover-transform", "translateY(-2px)")

	return NewComponent(ComponentCard, "stats-card").
		WithLabel("Revenue Stats").
		WithStyles(cardStyles).
		WithChildren(
			// Card header
			NewComponent(ComponentContainer, "card-header").
				WithStyles(factory.FlexStyles("row", "space-between", "center")).
				WithChildren(
					NewComponent(ComponentContainer, "title").
						WithLabel("Total Revenue").
						WithCustomCSS("font-size", "1.25rem").
						WithCustomCSS("font-weight", "600").
						Build(),

					NewComponent(ComponentBadge, "trend-badge").
						WithLabel("+12%").
						WithStyledVariant("success", schemaDir).
						Build(),
				).
				Build(),

			// Card content
			NewComponent(ComponentContainer, "card-content").
				WithCustomCSS("margin-top", "1rem").
				WithChildren(
					NewComponent(ComponentContainer, "revenue-amount").
						WithLabel("$127,450").
						WithCustomCSS("font-size", "2.5rem").
						WithCustomCSS("font-weight", "700").
						WithCustomCSS("color", "#1f2937").
						Build(),

					NewComponent(ComponentContainer, "revenue-description").
						WithLabel("Compared to last month").
						WithCustomCSS("color", "#6b7280").
						WithCustomCSS("margin-top", "0.5rem").
						Build(),
				).
				Build(),
		).
		Build()
}

// createStyledDataTable creates a data table with advanced styling
func createStyledDataTable(schemaDir string) Component {
	columns := []TableColumn{
		{
			Key:      "id",
			Title:    "ID",
			DataType: DataTypeNumber,
			Width:    "80px",
			Sortable: true,
		},
		{
			Key:        "name",
			Title:      "Customer Name",
			DataType:   DataTypeText,
			Sortable:   true,
			Filterable: true,
			Searchable: true,
		},
		{
			Key:      "email",
			Title:    "Email",
			DataType: DataTypeText,
		},
		{
			Key:        "amount",
			Title:      "Order Amount",
			DataType:   DataTypeCurrency,
			Sortable:   true,
			Filterable: true,
		},
		{
			Key:      "status",
			Title:    "Status",
			DataType: DataTypeBadge,
			Align:    AlignCenter,
		},
		{
			Key:      "actions",
			Title:    "Actions",
			DataType: DataTypeActions,
			Width:    "120px",
		},
	}

	return CreateStyledTable("orders-table", columns, "hover", schemaDir).
		WithLabel("Customer Orders").
		WithCustomCSS("box-shadow", "0 1px 3px rgba(0, 0, 0, 0.1)").
		WithCustomCSS("border-radius", "8px").
		WithCustomCSS("overflow", "hidden").
		WithConfig(TableConfig{
			Columns:      columns,
			DataSource:   "/api/orders",
			EmptyMessage: "No orders found",
			Pagination: &Pagination{
				PageSize:        25,
				ShowSizer:       true,
				ShowQuickJumper: true,
			},
			Search: &Search{
				Placeholder: "Search orders...",
				Columns:     []string{"name", "email"},
				Debounce:    300,
			},
			Actions: []Action{
				{
					Key:     "export",
					Label:   "Export",
					Icon:    "download",
					Variant: VariantSecondary,
					OnClick: "exportOrders()",
				},
			},
			RowActions: []Action{
				{
					Key:     "view",
					Label:   "View",
					Icon:    "eye",
					OnClick: "viewOrder(row.id)",
				},
				{
					Key:     "edit",
					Label:   "Edit",
					Icon:    "edit",
					OnClick: "editOrder(row.id)",
				},
				{
					Key:     "delete",
					Label:   "Delete",
					Icon:    "trash",
					Variant: VariantDanger,
					Confirm: &ConfirmDialog{
						Title:       "Delete Order",
						Description: "Are you sure you want to delete this order?",
						OkText:      "Delete",
						CancelText:  "Cancel",
					},
					OnClick: "deleteOrder(row.id)",
				},
			},
		}).
		Build()
}

// createResponsiveLayout creates a responsive grid layout
func createResponsiveLayout(schemaDir string) Component {
	factory := css.NewFactory(schemaDir)

	// Responsive grid that adapts to screen size
	layoutStyles := factory.ResponsiveStyles().
		Base(factory.GridStyles("1fr", "1rem")).
		SM(func(s *css.Styles) {
			s.WithGridTemplateColumns("repeat(2, 1fr)")
		}).
		MD(func(s *css.Styles) {
			s.WithGridTemplateColumns("repeat(3, 1fr)")
		}).
		LG(func(s *css.Styles) {
			s.WithGridTemplateColumns("repeat(4, 1fr)")
			s.WithGridGap("1.5rem")
		}).
		Build()

	return NewComponent(ComponentContainer, "responsive-grid").
		WithLabel("Responsive Dashboard Grid").
		WithStyles(layoutStyles).
		WithChildren(
			createStyledDashboardCard(schemaDir),
			createStyledDashboardCard(schemaDir),
			createStyledDashboardCard(schemaDir),
			createStyledDashboardCard(schemaDir),
		).
		Build()
}

// createThemeAwareCard creates a card that adapts to light/dark themes
func createThemeAwareCard(theme, schemaDir string) Component {
	factory := css.NewFactory(schemaDir)

	themeStyles := factory.ThemeStyles(theme).
		Primary("#1f2937", "#f9fafb").
		Background("#ffffff", "#1f2937").
		Border("#e5e7eb", "#374151").
		Build()

	return NewComponent(ComponentCard, "theme-card").
		WithLabel("Theme-Aware Card").
		WithStyles(themeStyles).
		WithCustomCSS("padding", "1.5rem").
		WithChildren(
			NewComponent(ComponentContainer, "theme-content").
				WithLabel(fmt.Sprintf("This card adapts to %s theme", theme)).
				Build(),
		).
		Build()
}

// validateComponents validates all components with styling
func validateComponents(components []Component, schemaDir string) {
	registry := NewMockRegistry()
	ctx := context.Background()

	for _, component := range components {
		errors := ValidateComponentWithStyles(ctx, registry, component, schemaDir)
		if len(errors) == 0 {
			fmt.Printf("✓ Component '%s' (%s) is valid\n", component.ID, component.Type)
		} else {
			fmt.Printf("✗ Component '%s' has %d errors:\n", component.ID, len(errors))
			for _, err := range errors {
				fmt.Printf("  - %s\n", err.Message)
			}
		}
	}
}

// Helper functions
func printComponentInfo(component Component) {
	fmt.Printf("Component: %s (%s)\n", component.ID, component.Type)
	fmt.Printf("Label: %s\n", component.Label)

	if component.Styles != nil {
		css := component.Styles.ToCSS()
		if css != "" {
			fmt.Printf("CSS: %s\n", css[:min(100, len(css))]+"...")
		}
	}

	fmt.Printf("Children: %d\n", len(component.Children))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// NewMockRegistry creates a mock registry for examples
func NewMockRegistry() ComponentRegistry {
	// This is a mock implementation for examples
	return &mockRegistry{}
}

type mockRegistry struct{}

func (m *mockRegistry) Register(componentType ComponentType, factory ComponentFactory) {}
func (m *mockRegistry) Create(ctx context.Context, componentType ComponentType, config map[string]any) (Component, error) {
	return Component{}, nil
}
func (m *mockRegistry) GetTypes() []ComponentType                               { return []ComponentType{} }
func (m *mockRegistry) Validate(ctx context.Context, component Component) error { return nil }

