package blocks

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/ui"
)

// ApprovalWorkflowBlock renders the approval status and action section.
// Gated on the session's approval_workflow feature flag.
// The block renders nothing-equivalent (empty section) when feature is disabled.
func ApprovalWorkflowBlock(sess ui.UISessionContext) ast.Node {
	if !sess.Flag("approval_workflow") {
		return ast.SectionNode{
			Title:     "Approval",
			Collapsed: true,
			Body:      []ast.Node{ast.InputTextNode{Name: "_approval_placeholder", Label: "Approval workflow not enabled", DisabledOn: "true"}},
		}
	}
	return ast.SectionNode{
		Title: "Approval",
		Body: []ast.Node{
			ast.SelectNode{
				Name:  "approval_status",
				Label: "Status",
				Options: []ast.SelectOption{
					{Label: "Pending", Value: "pending"},
					{Label: "Approved", Value: "approved"},
					{Label: "Rejected", Value: "rejected"},
				},
				DisabledOn: "!${can_approve}",
			},
			ast.InputTextNode{Name: "approval_note", Label: "Note", DisabledOn: "!${can_approve}"},
		},
	}
}
