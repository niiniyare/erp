package builders

import "awo.so/internal/web/ast"

// TenantStatusSelect returns a Select node for filtering/setting tenant status.
func TenantStatusSelect(name, label string) ast.Node {
	return ast.SelectNode{
		Name:  name,
		Label: label,
		Options: []ast.SelectOption{
			{Label: "Pending", Value: "PENDING"},
			{Label: "Active", Value: "ACTIVE"},
			{Label: "Suspended", Value: "SUSPENDED"},
			{Label: "Archived", Value: "ARCHIVED"},
		},
		Clearable: true,
	}
}

// PlanPickerNode returns a Select node for choosing a subscription plan.
func PlanPickerNode(name, label string) ast.Node {
	return ast.SelectNode{
		Name:   name,
		Label:  label,
		Source: &ast.APISpec{Method: "get", URL: "/api/v1/platform/plans/options"},
	}
}
