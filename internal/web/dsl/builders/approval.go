package builders

import "awo.so/internal/web/ast"

// ApprovalStatusSelect returns a Select node for filtering by approval status.
func ApprovalStatusSelect(name, label string) ast.Node {
	return ast.SelectNode{
		Name:  name,
		Label: label,
		Options: []ast.SelectOption{
			{Label: "Pending", Value: "pending"},
			{Label: "Approved", Value: "approved"},
			{Label: "Rejected", Value: "rejected"},
			{Label: "Recalled", Value: "recalled"},
		},
		Clearable: true,
	}
}

// ApproverPickerNode returns a Select node for choosing an approver.
func ApproverPickerNode(name, label string) ast.Node {
	return ast.SelectNode{
		Name:       name,
		Label:      label,
		Source:     &ast.APISpec{Method: "get", URL: "/api/v1/iam/users/options"},
		Searchable: true,
		Clearable:  true,
	}
}
