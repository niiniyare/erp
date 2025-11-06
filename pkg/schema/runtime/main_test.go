package runtime

import "github.com/niiniyare/erp/pkg/schema"

// Helper function to create a test schema
func createTestSchema() *schema.Schema {
	return &schema.Schema{
		ID:    "test_schema",
		Title: "Test Schema",
		Fields: []schema.Field{
			{
				Name:     "name",
				Type:     schema.FieldText,
				Label:    "Full Name",
				Required: true,
				Runtime: &schema.FieldRuntime{
					Visible:  true,
					Editable: true,
					Reason:   "",
				},
			},
			{
				Name:     "email",
				Type:     schema.FieldEmail,
				Label:    "Email Address",
				Required: true,
				Runtime: &schema.FieldRuntime{
					Visible:  true,
					Editable: true,
					Reason:   "",
				},
			},
			{
				Name:     "age",
				Type:     schema.FieldNumber,
				Label:    "Age",
				Required: false,
				Runtime: &schema.FieldRuntime{
					Visible:  true,
					Editable: true,
					Reason:   "",
				},
			},
		},
	}
}
