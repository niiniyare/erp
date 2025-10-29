package schema

// Example usage patterns for the schema system

// Example: Simple user registration form
func ExampleUserRegistrationForm() *Schema {
	return NewBuilder("user-registration", TypeForm, "Create Account").
		WithDescription("Register a new user account").
		WithCategory("auth").
		WithModule("users").
		WithConfig(NewSimpleConfig("/api/v1/users", "POST")).
		WithCSRF().
		WithHTMX("/api/v1/users", "#main-content").
		AddTextField("username", "Username", true).
		AddEmailField("email", "Email Address", true).
		AddPasswordField("password", "Password", true).
		AddPasswordField("confirm_password", "Confirm Password", true).
		AddCheckboxField("terms", "I agree to the terms and conditions", false).
		AddSubmitButton("Create Account").
		MustBuild()
}

// Example: Product form with sections
func ExampleProductForm() *Schema {
	schema := NewBuilder("product-form", TypeForm, "Add Product").
		WithDescription("Create a new product").
		WithModule("inventory").
		WithConfig(NewSimpleConfig("/api/v1/products", "POST")).
		WithCSRF().
		WithLayout(&Layout{
			Type: LayoutSections,
			Sections: []Section{
				{
					ID:      "basic",
					Title:   "Basic Information",
					Fields:  []string{"name", "sku", "category"},
					Columns: 2,
				},
				{
					ID:      "pricing",
					Title:   "Pricing",
					Fields:  []string{"cost", "price", "tax_rate"},
					Columns: 3,
				},
				{
					ID:      "inventory",
					Title:   "Inventory",
					Fields:  []string{"quantity", "min_stock", "location"},
					Columns: 3,
				},
			},
		}).
		AddTextField("name", "Product Name", true).
		AddTextField("sku", "SKU", true).
		AddSelectField("category", "Category", true, []Option{
			{Value: "electronics", Label: "Electronics"},
			{Value: "clothing", Label: "Clothing"},
			{Value: "food", Label: "Food & Beverage"},
		}).
		AddNumberField("cost", "Cost Price", true, float64Ptr(0), nil).
		AddNumberField("price", "Selling Price", true, float64Ptr(0), nil).
		AddNumberField("tax_rate", "Tax Rate (%)", false, float64Ptr(0), float64Ptr(100)).
		AddNumberField("quantity", "Current Stock", true, float64Ptr(0), nil).
		AddNumberField("min_stock", "Minimum Stock Level", false, float64Ptr(0), nil).
		AddTextField("location", "Storage Location", false).
		AddSubmitButton("Save Product").
		AddResetButton("Clear Form").
		MustBuild()

	return schema
}

// Example: Multi-tenant invoice form
func ExampleInvoiceForm() *Schema {
	return NewBuilder("invoice-form", TypeForm, "Create Invoice").
		WithDescription("Create a new invoice for a customer").
		WithModule("accounting").
		WithTenant("tenant_id", "strict").
		WithConfig(NewSimpleConfig("/api/v1/invoices", "POST")).
		WithCSRF().
		WithRateLimit(50, 3600).
		WithLayout(NewGridLayout(2)).
		AddFieldWithConfig(Field{
			Name:     "customer_id",
			Type:     FieldSelect,
			Label:    "Customer",
			Required: true,
			DataSource: &DataSource{
				Type:     "api",
				URL:      "/api/v1/customers",
				Method:   "GET",
				CacheTTL: 300,
			},
		}).
		AddDateField("invoice_date", "Invoice Date", true).
		AddDateField("due_date", "Due Date", true).
		AddTextField("invoice_number", "Invoice Number", true).
		AddFieldWithConfig(Field{
			Name:        "line_items",
			Type:        FieldJSON,
			Label:       "Line Items",
			Description: "Invoice line items in JSON format",
			Required:    true,
		}).
		AddNumberField("subtotal", "Subtotal", true, float64Ptr(0), nil).
		AddNumberField("tax", "Tax Amount", false, float64Ptr(0), nil).
		AddNumberField("total", "Total Amount", true, float64Ptr(0), nil).
		AddSubmitButton("Create Invoice").
		MustBuild()
}

// Example: Multi-step employee onboarding
func ExampleEmployeeOnboardingWizard() *Schema {
	return NewBuilder("employee-onboarding", TypeWorkflow, "Employee Onboarding").
		WithDescription("Multi-step employee onboarding process").
		WithModule("hr").
		WithLayout(&Layout{
			Type: LayoutSteps,
			Steps: []Step{
				{
					ID:          "personal",
					Title:       "Personal Information",
					Description: "Basic personal details",
					Fields:      []string{"first_name", "last_name", "email", "phone", "dob"},
					Order:       0,
					Validation:  true,
				},
				{
					ID:          "employment",
					Title:       "Employment Details",
					Description: "Job and department information",
					Fields:      []string{"job_title", "department", "start_date", "manager"},
					Order:       1,
					Validation:  true,
				},
				{
					ID:          "documents",
					Title:       "Documents",
					Description: "Upload required documents",
					Fields:      []string{"id_document", "contract", "bank_details"},
					Order:       2,
					Validation:  true,
				},
				{
					ID:          "review",
					Title:       "Review & Submit",
					Description: "Review all information before submission",
					Fields:      []string{"confirmation"},
					Order:       3,
					Validation:  true,
				},
			},
		}).
		// Step 1 fields
		AddTextField("first_name", "First Name", true).
		AddTextField("last_name", "Last Name", true).
		AddEmailField("email", "Email", true).
		AddTextField("phone", "Phone Number", true).
		AddDateField("dob", "Date of Birth", true).
		// Step 2 fields
		AddTextField("job_title", "Job Title", true).
		AddSelectField("department", "Department", true, []Option{
			{Value: "engineering", Label: "Engineering"},
			{Value: "sales", Label: "Sales"},
			{Value: "hr", Label: "Human Resources"},
		}).
		AddDateField("start_date", "Start Date", true).
		AddFieldWithConfig(Field{
			Name:     "manager",
			Type:     FieldSelect,
			Label:    "Reporting Manager",
			Required: true,
			DataSource: &DataSource{
				Type:   "api",
				URL:    "/api/v1/employees/managers",
				Method: "GET",
			},
		}).
		// Step 3 fields
		AddFieldWithConfig(Field{
			Name:     "id_document",
			Type:     FieldFile,
			Label:    "ID Document",
			Required: true,
			Config: map[string]any{
				"accept":  ".pdf,.jpg,.png",
				"maxSize": 5242880, // 5MB
			},
		}).
		AddFieldWithConfig(Field{
			Name:     "contract",
			Type:     FieldFile,
			Label:    "Signed Contract",
			Required: true,
		}).
		AddFieldWithConfig(Field{
			Name:     "bank_details",
			Type:     FieldFile,
			Label:    "Bank Details",
			Required: true,
		}).
		// Step 4 field
		AddCheckboxField("confirmation", "I confirm all information is correct", false).
		AddSubmitButton("Complete Onboarding").
		MustBuild()
}

// Example: Approval workflow for expense report
func ExampleExpenseReportWithWorkflow() *Schema {
	schema := NewBuilder("expense-report", TypeForm, "Expense Report").
		WithDescription("Submit an expense report for approval").
		WithModule("finance").
		WithTenant("tenant_id", "strict").
		WithConfig(NewSimpleConfig("/api/v1/expense-reports", "POST")).
		AddDateField("report_date", "Report Date", true).
		AddTextField("description", "Description", true).
		AddNumberField("amount", "Total Amount", true, float64Ptr(0), nil).
		AddFieldWithConfig(Field{
			Name:     "receipts",
			Type:     FieldFile,
			Label:    "Receipts",
			Required: true,
			Config: map[string]any{
				"multiple": true,
				"accept":   ".pdf,.jpg,.png",
			},
		}).
		AddSubmitButton("Submit for Approval").
		MustBuild()

	// Add workflow configuration
	schema.Workflow = &Workflow{
		Enabled: true,
		Stage:   "draft",
		Status:  "pending",
		Actions: []WorkflowAction{
			{
				ID:       "submit",
				Label:    "Submit",
				Type:     "submit",
				ToStage:  "pending_approval",
				ToStatus: "submitted",
				Icon:     "send",
				Variant:  "primary",
			},
			{
				ID:          "approve",
				Label:       "Approve",
				Type:        "approve",
				ToStage:     "approved",
				ToStatus:    "approved",
				Permissions: []string{"expense.approve"},
				Icon:        "check",
				Variant:     "primary",
			},
			{
				ID:          "reject",
				Label:       "Reject",
				Type:        "reject",
				ToStage:     "rejected",
				ToStatus:    "rejected",
				Permissions: []string{"expense.approve"},
				RequireNote: true,
				Icon:        "x",
				Variant:     "destructive",
			},
		},
		Approvals: &ApprovalConfig{
			Required:     true,
			MinApprovals: 1,
			Approvers: []ApproverConfig{
				{Type: "role", Value: "manager", Order: 1},
			},
		},
		Notifications: []Notification{
			{
				Event:    "approval_required",
				Template: "expense_approval_required",
				To:       []string{"manager"},
				Method:   "email",
			},
			{
				Event:    "approved",
				Template: "expense_approved",
				To:       []string{"submitter"},
				Method:   "email",
			},
		},
		History: true,
	}

	return schema
}

// Example: Using field builder
func ExampleFieldBuilder() Field {
	minLen := 3
	maxLen := 50

	return NewField("username", FieldText).
		WithLabel("Username").
		WithDescription("Choose a unique username").
		WithPlaceholder("Enter username").
		WithHelp("Username must be 3-50 characters").
		Required().
		WithValidation(&FieldValidation{
			MinLength: &minLen,
			MaxLength: &maxLen,
			Pattern:   "^[a-zA-Z0-9_]+$",
			Messages: Messages{
				Required:  "Username is required",
				MinLength: "Username must be at least 3 characters",
				MaxLength: "Username cannot exceed 50 characters",
				Pattern:   "Username can only contain letters, numbers, and underscores",
			},
		}).
		Build()
}

// Helper to create float64 pointer (for validation)
func float64Ptr(f float64) *float64 {
	return &f
}

// Helper to create int pointer
func intPtr(i int) *int {
	return &i
}
