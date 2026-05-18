package screens

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/dsl/blocks"
	"awo.so/internal/web/ui"
)

// ApprovalQueueScreen renders the approval inbox — pending items requiring action.
func ApprovalQueueScreen(sess ui.UISessionContext) ast.Node {
	filter := blocks.FilterBarBlock(sess, blocks.FilterBarConfig{
		ShowStatus: true,
		StatusOptions: []ast.SelectOption{
			{Label: "Pending", Value: "pending"},
			{Label: "Approved", Value: "approved"},
			{Label: "Rejected", Value: "rejected"},
		},
		ShowTypeFilter: true,
		TypeOptions: []ast.SelectOption{
			{Label: "Invoice", Value: "invoice"},
			{Label: "Purchase Order", Value: "purchase_order"},
			{Label: "Expense", Value: "expense_claim"},
			{Label: "Journal", Value: "journal_entry"},
		},
		TypeFieldName: "document_type",
		TypeLabel:     "Document Type",
		ShowSearch:    true,
	})
	table := blocks.DataTableBlock(sess, blocks.DataTableConfig{
		APIURL:  "/api/v1/approvals/queue",
		Columns: approvalQueueColumns(),
		Filter:  filter,
		RowActions: []ast.ActionNode{
			{Label: "Approve", ActionType: "ajax", Level: "success", API: &ast.APISpec{Method: "post", URL: "/api/v1/approvals/${id}/approve"}, ConfirmText: "Approve this document?"},
			{Label: "Reject", ActionType: "ajax", Level: "danger", API: &ast.APISpec{Method: "post", URL: "/api/v1/approvals/${id}/reject"}, ConfirmText: "Reject this document?"},
		},
	})
	return ast.PageNode{Title: "Approval Queue", Body: []ast.Node{table}}
}

func approvalQueueColumns() []blocks.ColumnDef {
	return []blocks.ColumnDef{
		{Name: "document_ref", Label: "Reference", Sortable: true},
		{Name: "document_type", Label: "Type"},
		{Name: "submitted_by", Label: "Submitted By"},
		{Name: "submitted_at", Label: "Date", Type: "date", Sortable: true},
		{Name: "amount", Label: "Amount", Type: "currency"},
		{Name: "status", Label: "Status", Type: "status"},
	}
}
