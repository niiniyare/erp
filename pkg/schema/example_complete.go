package schema

import (
	"context"
	"fmt"
	"time"

	"github.com/niiniyare/erp/pkg/condition"
)

// ComprehensiveExample demonstrates the complete schema system with all features
func CreateComprehensiveExampleSchema() (*Schema, error) {
	// Create a builder with enterprise features
	builder := NewBuilder("comprehensive_example", TypeForm, "Comprehensive Example Form").
		WithDescription("A complete example showcasing all schema features").
		WithVersion("1.0.0").
		WithCategory("examples").
		WithTags("comprehensive", "demo", "full-featured")

	// Apply enterprise security
	builder.WithCSRF().
		WithRateLimit(100, 3600).
		WithTenant("tenant_id", "strict")

	// Configure HTMX and Alpine.js
	builder.WithHTMX("/api/examples", "#form-result").
		WithAlpine(`{
			formData: {},
			calculating: false,
			submitForm() {
				this.calculating = true;
				// Form submission logic
			}
		}`)

	// Apply foundational mixins
	builder.WithMixin("audit_fields").
		WithMixin("status_fields")

	// Add comprehensive field types with validation
	builder.AddEmailField("email", "Email Address", true)

	// Add number field with business rule integration
	builder.AddNumberField("salary", "Annual Salary", true, float64Ptr(0), float64Ptr(1000000))

	// Add select field with dynamic options
	departmentField := NewField("department", FieldSelect).
		WithLabel("Department").
		Required().
		WithOptions([]Option{
			CreateOption("engineering", "Engineering"),
			CreateOption("marketing", "Marketing"),
			CreateOption("sales", "Sales"),
			CreateOption("hr", "Human Resources"),
		}).
		Build()
	builder.AddFieldWithConfig(departmentField)

	// Add conditional field that shows based on department
	salaryBandField := NewField("salary_band", FieldSelect).
		WithLabel("Salary Band").
		WithOptions([]Option{
			CreateOption("junior", "Junior (0-2 years)"),
			CreateOption("mid", "Mid-level (3-5 years)"),
			CreateOption("senior", "Senior (6+ years)"),
		}).
		Build()
	builder.AddFieldWithConfig(salaryBandField)

	// Add repeatable field for skills
	skillsTemplate := []Field{
		{
			Name:     "skill_name",
			Type:     FieldText,
			Label:    "Skill",
			Required: true,
		},
		{
			Name:     "proficiency",
			Type:     FieldSelect,
			Label:    "Proficiency Level",
			Required: true,
			Options: []Option{
				CreateOption("beginner", "Beginner"),
				CreateOption("intermediate", "Intermediate"),
				CreateOption("advanced", "Advanced"),
				CreateOption("expert", "Expert"),
			},
		},
		{
			Name:     "years_experience",
			Type:     FieldNumber,
			Label:    "Years of Experience",
			Required: true,
			Validation: &FieldValidation{
				Min: float64Ptr(0),
				Max: float64Ptr(50),
			},
		},
	}

	skillsField, _ := NewRepeatableField("skills", "Skills & Experience").
		WithTemplate(skillsTemplate).
		WithMinItems(1).
		WithMaxItems(20).
		WithItemLabel("Skill %d").
		WithSortable(true).
		WithTexts("Add Skill", "Remove Skill").
		Build()

	builder.WithRepeatable(skillsField)

	// Add business rules for dynamic behavior
	// Rule 1: Show salary band field only for engineering department
	salaryBandVisibilityRule, _ := NewBusinessRule(
		"salary_band_visibility",
		"Show salary band for engineering",
		RuleTypeFieldVisibility,
	).WithCondition(&condition.ConditionGroup{
		ID:          "eng_check",
		Conjunction: condition.ConjunctionAnd,
		If:          "data.department == 'engineering'",
	}).WithAction(ActionShowField, "salary_band", nil).Build()

	builder.WithBusinessRule(salaryBandVisibilityRule)

	// Rule 2: Set salary default based on department and band
	salaryCalculationRule, _ := NewBusinessRule(
		"salary_calculation",
		"Calculate salary based on department and band",
		RuleTypeDataCalculation,
	).WithCondition(&condition.ConditionGroup{
		ID:          "salary_calc",
		Conjunction: condition.ConjunctionAnd,
		If:          "data.department == 'engineering' && data.salary_band != ''",
	}).WithActionAndParams(
		ActionCalculate,
		"salary",
		nil,
		map[string]any{
			"formula": "getSalaryForBand(department, salary_band)",
		},
	).Build()

	builder.WithBusinessRule(salaryCalculationRule)

	// Add custom validators
	builder.WithCustomValidator("professional_email", func(ctx context.Context, value any, params map[string]any) error {
		email, ok := value.(string)
		if !ok {
			return fmt.Errorf("value must be string")
		}
		
		// Check if email has professional domain
		if email != "" && !contains(email, []string{"@company.com", "@enterprise.org", "@business.net"}) {
			return fmt.Errorf("please use a professional email address")
		}
		return nil
	})

	// Add async validator for email uniqueness
	builder.WithAsyncValidator(&AsyncValidator{
		Name:     "email_uniqueness",
		Debounce: 500 * time.Millisecond,
		Cache:    true,
		CacheTTL: 5 * time.Minute,
		Validate: func(ctx context.Context, value any, params map[string]any) error {
			// Simulate database check
			time.Sleep(100 * time.Millisecond)
			// In real implementation, would check database
			return nil
		},
	})

	// Configure layout with tabs
	layout := NewTabLayout([]Tab{
		{
			ID:     "personal",
			Label:  "Personal Information",
			Icon:   "user",
			Fields: []string{"email", "department", "salary_band", "salary"},
		},
		{
			ID:     "skills",
			Label:  "Skills & Experience",
			Icon:   "star",
			Fields: []string{"skills"},
		},
		{
			ID:     "system",
			Label:  "System Fields",
			Icon:   "cog",
			Fields: []string{"created_at", "updated_at", "created_by", "updated_by", "status", "is_active"},
		},
	})
	builder.WithLayout(layout)

	// Add actions with different variants
	builder.AddSubmitButton("Save Employee").
		AddButton("preview", "Preview", "outline").
		AddButton("validate", "Validate", "secondary").
		AddActionWithConfig(Action{
			ID:      "export",
			Type:    ActionButton,
			Text:    "Export Data",
			Variant: "ghost",
			Icon:    "download",
			Config: &ActionConfig{
				Handler: "exportFormData",
			},
		})

	// Configure internationalization
	builder.WithI18n("en-US", "es-ES", "fr-FR")

	// Build the final schema
	return builder.Build()
}

// CreateInvoiceFormWithLineItems demonstrates repeatable fields for invoice line items
func CreateInvoiceFormWithLineItems() (*Schema, error) {
	builder := NewInvoiceFormBuilder()

	// Add workflow for approval process
	approvalWorkflow := &Workflow{
		Enabled: true,
		Actions: []WorkflowAction{
			{
				ID:      "submit_for_approval",
				Label:   "Submit for Approval",
				Type:    "submit",
				ToStage: "pending_approval",
				Icon:    "arrow-right",
				Variant: "primary",
			},
			{
				ID:      "approve",
				Label:   "Approve",
				Type:    "approve",
				ToStage: "approved",
				Icon:    "check",
				Variant: "success",
				Permissions: []string{"invoice:approve"},
			},
			{
				ID:      "reject",
				Label:   "Reject",
				Type:    "reject",
				ToStage: "rejected",
				Icon:    "x",
				Variant: "destructive",
				RequireNote: true,
				Permissions: []string{"invoice:approve"},
			},
		},
		Approvals: &ApprovalConfig{
			Required:     true,
			MinApprovals: 1,
			Approvers: []ApproverConfig{
				{
					Type:  "role",
					Value: "finance_manager",
					Order: 1,
				},
			},
		},
	}

	schema, err := builder.Build()
	if err != nil {
		return nil, err
	}

	schema.Workflow = approvalWorkflow

	return schema, nil
}

// DemonstrateDynamicSchemaWithBusinessRules shows how schemas can be modified at runtime
func DemonstrateDynamicSchemaWithBusinessRules() error {
	// Create base schema
	schema, err := CreateComprehensiveExampleSchema()
	if err != nil {
		return err
	}

	// Get business rule engine from builder
	builder := NewBuilder("temp", TypeForm, "Temp")
	ruleEngine := builder.GetBusinessRuleEngine()

	// Sample form data
	formData := map[string]any{
		"department":   "engineering",
		"salary_band":  "senior",
		"email":        "john.doe@company.com",
	}

	// Apply business rules to modify schema dynamically
	ctx := context.Background()
	modifiedSchema, err := ruleEngine.ApplyRules(ctx, schema, formData)
	if err != nil {
		return err
	}

	// The modified schema now has rules applied
	fmt.Printf("Original schema fields: %d\n", len(schema.Fields))
	fmt.Printf("Modified schema fields: %d\n", len(modifiedSchema.Fields))

	// Check if salary_band field is now visible (it should be for engineering)
	salaryBandField, exists := modifiedSchema.GetField("salary_band")
	if exists {
		fmt.Printf("Salary band field visible: %t\n", !salaryBandField.Hidden)
	}

	return nil
}

// Helper function for custom validation
func contains(str string, substrings []string) bool {
	for _, substring := range substrings {
		if len(str) >= len(substring) {
			for i := 0; i <= len(str)-len(substring); i++ {
				if str[i:i+len(substring)] == substring {
					return true
				}
			}
		}
	}
	return false
}